package irgen

import (
	"fmt"

	dt "github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/lexer"
	"github.com/EladB1/The/internal/parser"
	"github.com/EladB1/The/internal/semantic"
)

func translateBlock(nodes []*parser.AST, exitLabel string, startLabel string) []TAC {
	instructions := []TAC{}
	for _, node := range nodes {
		if node.Label == "Variable" {
			instructions = append(instructions, variableDeclaration(node)...)
		} else if node.Label == "control-flow" {
			instructions = append(instructions, translateControlFlow(node, exitLabel, startLabel)...)
		} else if node.Token.Kind == lexer.OPERATOR_ASSIGN {
			instructions = append(instructions, translateAssignment(node)...)
		} else if node.Label == "while" {
			instructions = append(instructions, translateWhileLoop(node)...)
		} else if node.Label == "for" {
			instructions = append(instructions, translateForLoop(node)...)
		} else if node.Label == "if-block" {
			instructions = append(instructions, translateIfBlock(node, exitLabel, startLabel)...)
		} else {
			expression, _ := translateExpression(*node)
			instructions = append(instructions, expression...)
		}
	}
	return instructions
}

func translateControlFlow(node *parser.AST, exitLabel string, startLabel string) []TAC {
	instructions := []TAC{}
	switch node.Children[0].Token.Value {
	case "return":
		if len(node.Children) == 1 {
			instructions = append(instructions, Instruction{
				Operation:   Return,
				SrcPosition: node.Location,
			})
		} else {
			value_in, operand := translateExpression(*node.Children[1])
			instructions = append(instructions, value_in...)
			instructions = append(instructions, Instruction{
				Operation:   Return,
				Operand1:    operand,
				SrcPosition: node.Location,
			})
		}
	case "continue":
		instructions = append(instructions, Instruction{
			Operation: JMP,
			Operand1: Operand{
				Label: startLabel,
			},
			SrcPosition: node.Location,
		})
	case "break":
		instructions = append(instructions, Instruction{
			Operation: JMP,
			Operand1: Operand{
				Label: exitLabel,
			},
			SrcPosition: node.Location,
		})
	}
	return instructions
}

func translateAssignment(node *parser.AST) []TAC {
	instructions := []TAC{}
	scope := currScope
	var nameNode *parser.AST
	if node.Children[0].Label == "dot" {
		if node.Children[0].Children[0].Type.Equals(dt.GlobalRefType) {
			nameNode = node.Children[0].Children[1]
			currScope = globalScope
		} else {
			return translateDotAssignment(node)
		}
	} else {
		nameNode = node.Children[0]
	}
	name := nameNode.Token.Value
	value := node.Children[1]
	var value_in []TAC
	var value_op Operand
	if node.Token.Value != "=" {
		var_in, operand := loadVariable(*nameNode)
		value_in, value_op = translateExpression(*value)
		instructions = append(instructions, var_in...)
		var operation Operation
		result := formTempVar(operand.Type)
		switch node.Token.Value {
		case "+=":
			operation = typedOperation(operand.Type, "add")
		case "-=":
			operation = typedOperation(operand.Type, "sub")
		case "*=":
			operation = typedOperation(operand.Type, "mul")
		case "/=":
			operation = typedOperation(operand.Type, "div")
		}
		instructions = append(instructions, value_in...)
		instructions = append(instructions, Instruction{
			Destination: result,
			Operation:   operation,
			Operand1:    operand,
			Operand2:    value_op,
			SrcPosition: node.Location,
		})
		value_op = Operand{
			Type:        operand.Type,
			Var:         result,
			SrcPosition: node.Location,
		}
	} else {
		value_in, value_op = translateExpression(*value)
		instructions = append(instructions, value_in...)
	}
	variable := currScope.LookupVariable(name)
	if variable != nil {
		if variable.Ctx == semantic.StructProp {
			this := Variable{
				Name:        "__this",
				DataType:    dt.Ptr,
				Visibility:  Param,
				SrcPosition: node.Location,
			}
			instructions = append(instructions, Instruction{
				Operand1: Operand{
					Var: this,
				},
				SrcPosition: node.Location,
			},
				Instruction{
					Operation: Set,
					Operand1: Operand{
						Var:         this,
						Offset:      variable.Offset,
						SrcPosition: node.Location,
					},
					Operand2:    value_op,
					SrcPosition: node.Location,
				})
		} else {
			instructions = append(instructions, storeVariable(*variable, value_op))
		}
	}
	currScope = scope
	return instructions
}

func translateDotAssignment(node *parser.AST) []TAC {
	instructions := []TAC{}
	left := node.Children[0].Children[0]
	prop := node.Children[0].Children[1]
	var propSymbol *semantic.VariableSymbol
	var ptr_in []TAC
	var ptr Operand
	if left.Type.RootEquals(dt.Ref) {
		ptr = Operand{
			Type: dt.Ptr,
			Var: Variable{
				DataType:    dt.Ptr,
				Name:        "__this",
				Visibility:  Param,
				SrcPosition: left.Location,
			},
		}
		str := currScope.LookupStruct(left.Type.SubTypes[0].String())
		if str == nil {
			panic(fmt.Sprintf("Unable to find struct %s in scopeId %s", left.Type.SubTypes[0], currScope.Id))
		}
		propSymbol = str.InnerScope.LookupVariable(prop.Token.Value)
	} else {
		ptr_in, ptr = translateExpression(*left)
		instructions = append(instructions, ptr_in...)
		str := currScope.LookupStruct(string(left.Type.Root))
		if str == nil {
			panic(fmt.Sprintf("Unable to find struct %s in scopeId %s", left.Type.Root, currScope.Id))
		}
		propSymbol = str.InnerScope.LookupVariable(prop.Token.Value)
	}
	var value_in []TAC
	var value_op Operand
	if node.Token.Value != "=" {
		var operation Operation
		load := formTempVar(dt.TranslateSourceType(propSymbol.Type))
		instructions = append(instructions, Instruction{
			Destination: load,
			Operation:   Load,
			Operand1:    ptr,
			Operand2: Operand{
				Constant: propSymbol.Offset.Value,
			},
			SrcPosition: node.Location,
		})
		operand := Operand{
			Type:        load.DataType,
			Var:         load,
			SrcPosition: node.Location,
		}
		switch node.Token.Value {
		case "+=":
			operation = typedOperation(operand.Type, "add")
		case "-=":
			operation = typedOperation(operand.Type, "sub")
		case "*=":
			operation = typedOperation(operand.Type, "mul")
		case "/=":
			operation = typedOperation(operand.Type, "div")
		}
		value_in, value_op = translateExpression(*node.Children[1])
		instructions = append(instructions, value_in...)
		result := formTempVar(operand.Type)
		instructions = append(instructions, Instruction{
			Destination: result,
			Operation:   operation,
			Operand1: Operand{
				Var:    ptr.Var,
				Offset: propSymbol.Offset,
			},
			Operand2:    value_op,
			SrcPosition: node.Location,
		})
		value_op = Operand{
			Type:        result.DataType,
			Var:         result,
			SrcPosition: node.Location,
		}
	} else {
		value_in, value_op = translateExpression(*node.Children[1])
		instructions = append(instructions, value_in...)
	}

	instructions = append(instructions, Instruction{
		Operation: Set,
		Operand1: Operand{
			Type:        value_op.Type,
			Var:         ptr.Var,
			Offset:      propSymbol.Offset,
			SrcPosition: prop.Location,
		},
		Operand2:    value_op,
		SrcPosition: node.Location,
	})

	return instructions
}

func translateWhileLoop(node *parser.AST) []TAC {
	instructions := []TAC{}
	cond_in, cond := translateExpression(*node.Children[0])
	scope := currScope
	outerBlock := Block{
		Label: fmt.Sprintf("loop_exit@%d", loopIndex),
	}
	loop := Loop{
		Label: fmt.Sprintf("loop@%d", loopIndex),
		Code:  cond_in,
	}
	loop.Code = append(loop.Code, Instruction{
		Operation: JMPIFNOT,
		Operand1: Operand{
			Label: outerBlock.Label,
		},
		Operand2: cond,
	})
	currScope = scope.GetChildScopeById(node.IRName) // enter the loop scope
	if currScope == nil {
		currScope = scope
		return instructions
	}
	loopBody := Block{
		Label: fmt.Sprintf("loop_body@%d", loopIndex),
	}
	loopBody.Code = append(loopBody.Code, translateBlock(node.Children[1].Children, outerBlock.Label, loop.Label)...)
	loopBody.Code = append(loopBody.Code, Instruction{
		Operation: JMP,
		Operand1: Operand{
			Label: loop.Label,
		},
	}) // go back to the start of the loop
	currScope = scope
	loop.Code = append(loop.Code, loopBody)
	outerBlock.Code = append(outerBlock.Code, loop)
	loopIndex++
	instructions = append(instructions, outerBlock)
	return instructions
}

func translateForLoop(node *parser.AST) []TAC {
	instructions := []TAC{}
	scope := currScope
	currScope = currScope.GetChildScopeById(node.IRName)
	outerBlock := Block{
		Label: fmt.Sprintf("loop_exit@%d", loopIndex),
	}
	loop := Loop{
		Label: fmt.Sprintf("loop@%d", loopIndex),
	}
	loopBody := Block{
		Label: fmt.Sprintf("loop_body@%d", loopIndex),
	}
	loopConditions := node.Children[0]
	loopType := semantic.ForLoopType(loopConditions.IRName)
	var init []TAC
	var limit_in []TAC
	var limit Operand
	var iter_in []TAC
	switch loopType {
	case semantic.DeclarationLoop:
		init = variableDeclaration(loopConditions.Children[0])
		limit_in, limit = translateExpression(*loopConditions.Children[1])
	case semantic.AssignmentLoop:
		init = translateAssignment(loopConditions.Children[0])
		limit_in, limit = translateExpression(*loopConditions.Children[1])
	case semantic.RangeLoop:
		iterator := loopConditions.Children[0].Children[1]
		variable := currScope.LookupVariable(iterator.Token.Value)
		rangeExpr := loopConditions.Children[2].Children
		value_in, value := translateExpression(*rangeExpr[0])
		outerBlock.Code = append(outerBlock.Code, value_in...)
		init = []TAC{
			storeVariable(*variable, value),
		}
		loadVar_in, loadVar := loadVariable(*iterator)
		loop.Code = append(loop.Code, loadVar_in...)
		var compareOp Operation
		if rangeExpr[1].Token.Value == ".." {
			compareOp = typedOperation(value.Type, "lt")
		} else {
			compareOp = typedOperation(value.Type, "le")
		}
		limit_in, limit = translateExpression(*rangeExpr[2])
		compare := formTempVar(dt.I32)
		limit_in = append(limit_in, Instruction{
			Destination: compare,
			Operation:   compareOp,
			Operand1:    loadVar,
			Operand2:    limit,
			SrcPosition: rangeExpr[2].Location,
		})
		limit = Operand{
			Var: compare,
		}
		var incr_val Operand
		if len(rangeExpr) == 3 {
			incr_val = Operand{
				Type:        loadVar.Type,
				Constant:    1,
				SrcPosition: rangeExpr[1].Location,
			}
		} else {
			iter_in, incr_val = translateExpression(*rangeExpr[4])
		}
		addop := formTempVar(loadVar.Type)
		add := []TAC{
			Instruction{
				Destination: addop,
				Operation:   typedOperation(loadVar.Type, "add"),
				Operand1:    loadVar,
				Operand2:    incr_val,
				SrcPosition: rangeExpr[1].Location,
			},
			storeVariable(*variable, Operand{Var: addop}),
		}
		iter_in = append(iter_in, add...)

	case semantic.Foreach, semantic.IndexedForeach:
		var variable *semantic.VariableSymbol
		var containerNode *parser.AST
		var eachNode *parser.AST = loopConditions.Children[0]

		if loopType == semantic.Foreach {
			iterator_name := "__foreach_iter"
			variable = &semantic.VariableSymbol{
				Name:        iterator_name,
				Type:        dt.Int32Type,
				Initialized: true,
				Ctx:         semantic.Local,
			}
			currScope.Variables[iterator_name] = *variable
			containerNode = loopConditions.Children[2]
		} else {
			containerNode = loopConditions.Children[3]
			var nameNode *parser.AST
			if loopConditions.Children[0].Type.Equals(dt.Int32Type) {
				nameNode = loopConditions.Children[0].Children[1]
				eachNode = loopConditions.Children[1]
			} else {
				nameNode = loopConditions.Children[1].Children[1]
			}
			variable = currScope.LookupVariable(nameNode.Token.Value)
		}
		_, zero_val := getZeroValue(dt.Int32Type, loopConditions.Location)
		init = []TAC{
			storeVariable(*variable, zero_val),
		}
		container_in, container := translateExpression(*containerNode)
		length_in, length := getArrayLength(container)
		curr := Variable{
			Name:        variable.Name,
			DataType:    dt.I32,
			Visibility:  VariableScope(variable.Ctx),
			SrcPosition: eachNode.Location,
		}
		eachVar := currScope.LookupVariable(eachNode.Children[1].Token.Value)
		index_in, index := callFunction(string(StringIndex), dt.TranslateSourceType(eachNode.Type), eachNode.Location, container, Operand{
			Type:        dt.I32,
			Var:         curr,
			SrcPosition: eachNode.Children[1].Location,
		})
		index_in = append(index_in, storeVariable(*eachVar, index))

		loop.Code = append(loop.Code, index_in...)
		outerBlock.Code = append(outerBlock.Code, container_in...)
		outerBlock.Code = append(outerBlock.Code, length_in...)
		limit := formTempVar(dt.I32)
		limit_in = []TAC{
			Instruction{
				Destination: limit,
				Operation:   typedOperation(dt.I32, "lt"),
				Operand1: Operand{
					Type: dt.I32,
					Var:  curr,
				},
				Operand2:    length,
				SrcPosition: eachNode.Location,
			},
		}

		next := formTempVar(dt.I32)
		iter_in = []TAC{
			Instruction{
				Destination: next,
				Operation:   typedOperation(dt.I32, "add"),
				Operand1: Operand{
					Type: dt.I32,
					Var:  curr,
				},
				Operand2: Operand{
					Type:     dt.I32,
					Constant: 1,
				},
				SrcPosition: eachNode.Location,
			},
			storeVariable(*variable, Operand{
				Type: dt.I32,
				Var:  next,
			}),
		}
	}
	outerBlock.Code = append(outerBlock.Code, init...)
	loop.Code = append(loop.Code, limit_in...)
	loop.Code = append(loop.Code, Instruction{
		Operation: JMPIFNOT,
		Operand1: Operand{
			Label: outerBlock.Label,
		},
		Operand2:    limit,
		SrcPosition: limit.SrcPosition,
	})
	loopBody.Code = append(loopBody.Code, translateBlock(node.Children[1].Children, outerBlock.Label, loop.Label)...)
	if len(iter_in) == 0 {
		iter_in, _ = translateExpression(*loopConditions.Children[2])
	}
	loopBody.Code = append(loopBody.Code, iter_in...)
	loopBody.Code = append(loopBody.Code, Instruction{
		Operation: JMP,
		Operand1: Operand{
			Label: loop.Label,
		},
	})
	loop.Code = append(loop.Code, loopBody)
	outerBlock.Code = append(outerBlock.Code, loop)
	loopIndex++
	instructions = append(instructions, outerBlock)
	currScope = scope
	return instructions
}

func translateIfBlock(node *parser.AST, exitLabel string, startLabel string) []TAC {
	instructions := []TAC{}
	ifBlock := IfBlock{}
	cond_in, cond := translateExpression(*node.Children[0].Children[0])
	instructions = append(instructions, cond_in...)
	blocksIndex := 0
	scope := currScope
	currScope = currScope.GetChildScopeById(node.Children[0].IRName)
	if currScope == nil {
		panic(fmt.Sprintf("Could not find scope with Id %s", node.Children[0].IRName))
	}
	blocks := [][]TAC{translateBlock(node.Children[0].Children[1].Children, exitLabel, startLabel)}
	conditionsIndex := 0
	conditions := []Variable{cond.Var}
	conditionSetup := [][]TAC{cond_in}
	for i := 1; i < len(node.Children); i++ {
		currScope = scope
		block := node.Children[i]
		currScope = currScope.GetChildScopeById(block.IRName)
		if currScope == nil {
			panic(fmt.Sprintf("Could not find scope with Id %s", block.IRName))
		}
		if block.Label == "else if" {
			cond_in, cond := translateExpression(*block.Children[0])
			conditions = append(conditions, cond.Var)
			conditionSetup = append(conditionSetup, cond_in)
			conditionsIndex++

			blocks = append(blocks, translateBlock(block.Children[1].Children, exitLabel, startLabel))
			blocksIndex++
		} else {
			blocks = append(blocks, translateBlock(block.Children[0].Children, exitLabel, startLabel))
			blocksIndex++
		}
	}
	ifBlock.IfCondition = conditions[0]
	ifBlock.IfCode = &blocks[0]
	ifBlock.ElseCode = &[]TAC{}
	code := ifBlock.ElseCode
	for i := 1; i <= conditionsIndex; i++ {
		*code = append(*code, conditionSetup[i]...)
		inner := IfBlock{
			IfCondition: conditions[i],
			IfCode:      &blocks[i],
			ElseCode:    &[]TAC{},
		}
		*code = append(*code, inner)
		code = inner.ElseCode
	}
	if conditionsIndex < blocksIndex {
		*code = append(*code, blocks[blocksIndex]...)
	}

	instructions = append(instructions, ifBlock)
	currScope = scope
	return instructions
}
