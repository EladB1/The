package irgen

import (
	"fmt"

	"github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/lexer"
	"github.com/EladB1/The/internal/parser"
	//dt "github.com/EladB1/The/internal/datatypes"
)

func translateExpression(node parser.AST) ([]TAC, Operand) {
	instructions := []TAC{}
	operand := Operand{}
	token := node.Token
	if token.Value == "||" {

	} else if token.Value == "&&" {

	} else if token.Kind == lexer.OPERATOR_COMPARE {

	} else if token.Kind == lexer.OPERATOR_BS {

	} else if token.Kind == lexer.OPERATOR_BW {

	} else if token.Kind == lexer.OPERATOR_ADD {
		instructions, operand = translateAddition(node)
	} else if token.Kind == lexer.OPERATOR_MULT {

	} else if token.Value == "**" {

	} else if node.Label == "Unary" {

	} else if node.Label == "typecast" {

	} else if node.IsLiteral() {
		operand = translateLiteral(node)
	} else if node.Token.Kind == lexer.ID {
		variable := currScope.LookupVariable(node.Token.Value)
		varType := datatypes.TranslateSourceType(variable.Type)
		tempVar := formTempVar(varType)
		instructions = append(instructions, Instruction{
			Destination: tempVar,
			Operation:   Get,
			Operand1: Operand{
				Var: Variable{
					Name:       variable.Name,
					DataType:   varType,
					Visibility: VariableScope(variable.Ctx),
				},
			},
		})
		operand = Operand{
			Var: Variable{
				Name:     tempVar.Name,
				DataType: tempVar.DataType,
			},
		}
		fmt.Println(variable)
	}
	return instructions, operand
}

func translateAddition(node parser.AST) ([]TAC, Operand) {
	/*
		1. Get type from root
		2. Check left and right types for typecasting requirement
		3. Call translateExpression(left)
		4. Call translateExpression(right)
		5. Decide which operation to use
		6. return
	*/
	instructions := []TAC{}
	operand := Operand{}
	rootType := node.Type
	left := node.Children[0]
	right := node.Children[1]
	var operation Operation
	if rootType == datatypes.String {
		// char + char
		// char + string
		// string + char
		// string + string
	} else {

		l_in, l_op := translateExpression(*left)
		instructions = append(instructions, l_in...)
		r_in, r_op := translateExpression(*right)
		instructions = append(instructions, r_in...)
		if rootType == datatypes.Int32 {
		}
		operationType := datatypes.TranslateSourceType(rootType)
		// i64
		// i32 (unsigned)
		// i64 (unsigned)
		// f32
		// f64
		if node.Token.Value == "+" {
			operation = typedOperation(operationType, "add")
		} else {
			operation = typedOperation(operationType, "sub")
		}
		tempVar := formTempVar(operationType)
		op := Instruction{
			Destination: tempVar,
			Operation:   operation,
			Operand1:    l_op,
			Operand2:    r_op,
		}
		instructions = append(instructions, op)
		operand = Operand{
			Type: operationType,
			Var:  tempVar,
		}
	}
	return instructions, operand
}
