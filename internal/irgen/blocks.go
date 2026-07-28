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
				Operation: Return,
			})
		} else {
			value_in, operand := translateExpression(*node.Children[1])
			instructions = append(instructions, value_in...)
			instructions = append(instructions, Instruction{
				Operation: Return,
				Operand1:  operand,
			})
		}
	case "continue":
		instructions = append(instructions, Instruction{
			Operation: JMP,
			Operand1: Operand{
				Label: startLabel,
			},
		})
	case "break":
		instructions = append(instructions, Instruction{
			Operation: JMP,
			Operand1: Operand{
				Label: exitLabel,
			},
		})
	}
	return instructions
}

func translateAssignment(node *parser.AST) []TAC {
	instructions := []TAC{}
	name := node.Children[0].Token.Value
	value := node.Children[1]
	value_in, value_op := translateExpression(*value)
	if node.Token.Value != "=" {
		var_in, operand := loadVariable(*node.Children[0])
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
		})
		value_op = Operand{
			Type: operand.Type,
			Var:  result,
		}
	} else {
		instructions = append(instructions, value_in...)
	}
	variable := currScope.LookupVariable(name)
	if variable != nil {
		instructions = append(instructions, Instruction{
			Operation: Store,
			Operand1: Operand{
				Var: Variable{
					Name:       variable.Name,
					DataType:   dt.TranslateSourceType(variable.Type),
					Visibility: VariableScope(variable.Ctx),
				},
			},
			Operand2: value_op,
		})
	}
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
	compare := formTempVar(dt.I32)
	check := []TAC{
		Instruction{
			Destination: compare,
			Operation:   typedOperation(dt.I32, "eq"),
			Operand1:    cond,
			Operand2: Operand{
				Type:     dt.I32,
				Constant: 0,
			},
		},
		Instruction{
			Operation: JMPIF,
			Operand1: Operand{
				Label: outerBlock.Label,
			},
			Operand2: Operand{
				Type: dt.I32,
				Var:  compare,
			},
		}, // exit the loop if the condition is not true
	}
	loop.Code = append(loop.Code, check...)
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
		iter_in, _ = translateExpression(*loopConditions.Children[2])
	case semantic.AssignmentLoop:
		init = translateAssignment(loopConditions.Children[0])
		limit_in, limit = translateExpression(*loopConditions.Children[1])
		iter_in, _ = translateExpression(*loopConditions.Children[2])
	case semantic.RangeLoop:
		//
	case semantic.Foreach:
		// set the index to 0
		iterator_name := "__foreach_iter"
		variable := semantic.VariableSymbol{
			Name:        iterator_name,
			Type:        dt.Int32Type,
			Initialized: true,
			Ctx:         semantic.Local,
		}
		currScope.Variables[iterator_name] = variable
		init = []TAC{
			Instruction{
				Operation: Store,
				Operand1: Operand{
					Var: Variable{
						Name:       iterator_name,
						DataType:   dt.I32,
						Visibility: VariableScope(variable.Ctx),
					},
				},
				Operand2: getZeroValue(dt.Int32Type),
			},
		}
		container_in, container := translateExpression(*loopConditions.Children[2])
		outerBlock.Code = append(outerBlock.Code, container_in...)
		length_in, length := getArrayLength(container)
		outerBlock.Code = append(outerBlock.Code, length_in...)
		curr := formTempVar(dt.I32)
		limit := formTempVar(dt.I32)
		limit_in = []TAC{
			Instruction{
				Destination: curr,
				Operation:   Get,
				Operand1: Operand{
					Var: Variable{
						Name:       variable.Name,
						DataType:   dt.I32,
						Visibility: VariableScope(variable.Ctx),
					},
				},
			},
			Instruction{
				Destination: limit,
				Operation:   typedOperation(dt.I32, "lt"),
				Operand1: Operand{
					Type: dt.I32,
					Var:  curr,
				},
				Operand2: length,
			},
		}
		index := formTempVar(dt.TranslateSourceType(loopConditions.Children[0].Type))
		eachVar := currScope.LookupVariable(loopConditions.Children[0].Children[1].Token.Value)
		index_in := []TAC{
			Instruction{
				Operation: PrepareParam,
				Operand1:  container,
			},
			Instruction{
				Operation: PrepareParam,
				Operand1: Operand{
					Type: dt.I32,
					Var:  curr,
				},
			},
			Instruction{
				Destination: index,
				Operation:   Call,
				Operand1: Operand{
					Constant: "__str_index",
				}, // TODO: make this work for any container type
				Operand2: Operand{
					Constant: 2,
				},
			},
			Instruction{
				Operation: Store,
				Operand1: Operand{
					Var: Variable{
						Name:       eachVar.Name,
						DataType:   dt.TranslateSourceType(eachVar.Type),
						Visibility: VariableScope(eachVar.Ctx),
					},
				},
				Operand2: Operand{
					Var: index,
				},
			},
		}

		loopBody.Code = append(loopBody.Code, index_in...)
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
			},
			Instruction{
				Operation: Store,
				Operand1: Operand{
					Var: Variable{
						Name:       iterator_name,
						DataType:   dt.I32,
						Visibility: VariableScope(variable.Ctx),
					},
				},
				Operand2: Operand{
					Type: dt.I32,
					Var:  next,
				},
			},
		}
		// TODO: refactor
	case semantic.IndexedForeach:
		//
	default:
		//
	}
	outerBlock.Code = append(outerBlock.Code, init...)
	loop.Code = append(loop.Code, limit_in...)
	compare := formTempVar(dt.I32)
	loop.Code = append(loop.Code, Instruction{
		Destination: compare,
		Operation:   typedOperation(dt.I32, "eq"),
		Operand1:    limit,
		Operand2: Operand{
			Type:     dt.I32,
			Constant: 0,
		},
	}) // check that the condition is false
	loop.Code = append(loop.Code, Instruction{
		Operation: JMPIF,
		Operand1: Operand{
			Label: outerBlock.Label,
		},
		Operand2: Operand{
			Type: dt.I32,
			Var:  compare,
		},
	})
	loopBody.Code = append(loopBody.Code, translateBlock(node.Children[1].Children, outerBlock.Label, loop.Label)...)
	loopBody.Code = append(loopBody.Code, iter_in...)
	loopBody.Code = append(loopBody.Code, Instruction{
		Operation: JMP,
		Operand1: Operand{
			Label: loop.Label,
		},
	})
	// calculate limit, start, and increment
	// come up with loop structure (like while loop)
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
		currScope = scope
		return []TAC{}
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
			currScope = scope
			return []TAC{}
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
