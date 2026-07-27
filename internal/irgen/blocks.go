package irgen

import (
	"fmt"

	dt "github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/lexer"
	"github.com/EladB1/The/internal/parser"
)

func translateBlock(nodes []*parser.AST) []TAC {
	instructions := []TAC{}
	for _, node := range nodes {
		if node.Label == "Variable" {
			instructions = append(instructions, variableDeclaration(node)...)
		} else if node.Label == "control-flow" {
			instructions = append(instructions, translateControlFlow(node)...)
		} else if node.Token.Kind == lexer.OPERATOR_ASSIGN {
			instructions = append(instructions, translateAssignment(node)...)
		} else if node.Label == "while" {
			instructions = append(instructions, translateWhile(node)...)
		} else if node.Label == "for" {
			instructions = append(instructions, translateFor(node)...)
		} else if node.Label == "if-block" {
			instructions = append(instructions, translateIfBlock(node)...)
		} else {
			expression, _ := translateExpression(*node)
			instructions = append(instructions, expression...)
		}
	}
	return instructions
}

func translateControlFlow(node *parser.AST) []TAC {
	instructions := []TAC{}
	if node.Children[0].Token.Value == "return" {
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

func translateWhile(node *parser.AST) []TAC {
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
	loopBody.Code = append(loopBody.Code, translateBlock(node.Children[1].Children)...)
	loopBody.Code = append(loopBody.Code, Instruction{
		Operation: JMP,
		Operand1: Operand{
			Label: loop.Label,
		},
	}) // go back to the start of the loop
	// loop through body
	currScope = scope
	loop.Code = append(loop.Code, loopBody)
	outerBlock.Code = append(outerBlock.Code, loop)
	fmt.Println("HERE", outerBlock)
	loopIndex++
	instructions = append(instructions, outerBlock)
	return instructions
}

func translateFor(node *parser.AST) []TAC {
	instructions := []TAC{}

	loopIndex++
	return instructions
}

func translateIfBlock(node *parser.AST) []TAC {
	instructions := []TAC{}

	return instructions
}
