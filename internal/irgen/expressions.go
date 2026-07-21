package irgen

import (
	"fmt"

	"github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/lexer"
	"github.com/EladB1/The/internal/parser"
)

func translateExpression(node parser.AST) ([]TAC, Operand) {
	instructions := []TAC{}
	operand := Operand{}
	token := node.Token
	if token.Value == "||" || token.Value == "&&" {
		instructions, operand = translateLogicalAndOr(node)
	} else if token.Kind == lexer.OPERATOR_COMPARE {
		instructions, operand = translateComparison(node)
	} else if token.Kind == lexer.OPERATOR_BS || token.Kind == lexer.OPERATOR_BW {
		instructions, operand = translateBitOperation(node)
	} else if token.Kind == lexer.OPERATOR_ADD {
		instructions, operand = translateAddition(node)
	} else if token.Kind == lexer.OPERATOR_MULT {
		instructions, operand = translateMultiplication(node)
	} else if token.Value == "**" {
		instructions, operand = translateExponent(node)
	} else if node.Label == "Unary" {
		instructions, operand = translateUnary(node)
	} else if node.Label == "typecast" {
		instructions, operand = translateTypecast(node)
	} else if node.IsLiteral() {
		operand = translateLiteral(node)
	} else if node.Token.Kind == lexer.ID {
		instructions, operand = loadVariable(node)
	}
	return instructions, operand
}

func translateLogicalAndOr(node parser.AST) ([]TAC, Operand) {
	instructions := []TAC{}
	operand := Operand{}
	irType := datatypes.I32
	left := node.Children[0]
	right := node.Children[1]
	var operation Operation
	if node.Token.Value == "&&" {
		operation = typedOperation(irType, "and")
	} else {
		operation = typedOperation(irType, "or")
	}
	l_in, l_op := translateExpression(*left)
	instructions = append(instructions, l_in...)
	r_in, r_op := translateExpression(*right)
	instructions = append(instructions, r_in...)
	tempVar := formTempVar(irType)
	instructions = append(instructions, Instruction{
		Destination: tempVar,
		Operation:   operation,
		Operand1:    l_op,
		Operand2:    r_op,
	})
	operand = Operand{
		Var: Variable{
			Name:     tempVar.Name,
			DataType: irType,
		},
	}
	return instructions, operand
}

func translateComparison(node parser.AST) ([]TAC, Operand) {
	instructions := []TAC{}
	operand := Operand{}

	return instructions, operand
}

func translateBitOperation(node parser.AST) ([]TAC, Operand) {
	instructions := []TAC{}
	operand := Operand{}
	rootType := node.Type
	irType := datatypes.TranslateSourceType(rootType)
	left := node.Children[0]
	right := node.Children[1]
	var operation Operation
	// TODO: typecasting and signed/unsigned
	switch node.Token.Value {
	case "^":
		operation = typedOperation(irType, "xor")
	case "&":
		operation = typedOperation(irType, "and")
	case "|":
		operation = typedOperation(irType, "or")
	case "<<":
		operation = typedOperation(irType, "lshift")
	case ">>":
		operation = typedOperation(irType, "rshift")
	}
	l_in, l_op := translateExpression(*left)
	instructions = append(instructions, l_in...)
	r_in, r_op := translateExpression(*right)
	instructions = append(instructions, r_in...)
	tempVar := formTempVar(irType)
	instructions = append(instructions, Instruction{
		Destination: tempVar,
		Operation:   operation,
		Operand1:    l_op,
		Operand2:    r_op,
	})
	operand = Operand{
		Type: irType,
		Var:  tempVar,
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

func translateMultiplication(node parser.AST) ([]TAC, Operand) {
	instructions := []TAC{}
	operand := Operand{}
	rootType := node.Type
	left := node.Children[0]
	right := node.Children[1]
	var operation Operation
	l_in, l_op := translateExpression(*left)
	instructions = append(instructions, l_in...)
	r_in, r_op := translateExpression(*right)
	instructions = append(instructions, r_in...)
	operationType := datatypes.TranslateSourceType(rootType)
	switch node.Token.Value {
	case "*":
		operation = typedOperation(operationType, "mul")
	case "/":
		// Handle: unsigned vs signed
		operation = typedOperation(operationType, "div")
	case "%":
		// Handle: unsigned vs signed
		operation = typedOperation(operationType, "mod")
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
	return instructions, operand
}

func translateExponent(node parser.AST) ([]TAC, Operand) {
	instructions := []TAC{}
	operand := Operand{}
	rootType := node.Type
	irType := datatypes.TranslateSourceType(rootType)
	left := node.Children[0]
	right := node.Children[1]
	if left.Type != rootType {
		// TODO
	}
	if right.Type != rootType {
		// TODO
	}
	l_in, l_op := translateExpression(*left)
	instructions = append(instructions, l_in...)
	r_in, r_op := translateExpression(*right)
	instructions = append(instructions, r_in...)
	instructions = append(instructions, Instruction{
		Operation: PrepareParam,
		Operand1:  l_op,
	})
	instructions = append(instructions, Instruction{
		Operation: PrepareParam,
		Operand1:  r_op,
	})
	tempVar := formTempVar(irType)
	instructions = append(instructions, Instruction{
		Destination: tempVar,
		Operation:   Call,
		Operand1: Operand{
			Constant: fmt.Sprintf("__%s_pow", irType),
		},
		Operand2: Operand{
			Constant: 2,
		},
	})
	operand = Operand{
		Type: irType,
		Var:  tempVar,
	}
	return instructions, operand
}

func translateUnary(node parser.AST) ([]TAC, Operand) {
	instructions := []TAC{}
	operand := Operand{}

	return instructions, operand
}

func translateTypecast(node parser.AST) ([]TAC, Operand) {
	instructions := []TAC{}

	targetType := datatypes.TranslateSourceType(node.Type)
	sourceType := datatypes.TranslateSourceType(node.Children[0].Type)
	l_in, l_op := translateExpression(*node.Children[0])
	instructions = append(instructions, l_in...)
	operation := getTypeCastOperation(sourceType, targetType)
	tempVar := formTempVar(targetType)
	instructions = append(instructions, Instruction{
		Destination: tempVar,
		Operation:   operation,
		Operand1:    l_op,
	})
	operand := Operand{
		Type: targetType,
		Var:  tempVar,
	}
	return instructions, operand
}

func getTypeCastOperation(src datatypes.IRType, target datatypes.IRType) Operation {
	switch src {
	case datatypes.I32:
		switch target {
		case datatypes.I64:
			return I32ToI64
		case datatypes.F32:
			return I32ToF32
		case datatypes.F64:
			return I32ToF64
		}
	case datatypes.U32:
		switch target {
		case datatypes.I64:
			return U32ToI64
		case datatypes.F32:
			return U32ToF32
		case datatypes.F64:
			return U32ToF64
		}
	case datatypes.I64:
		switch target {
		case datatypes.I32:
			return I64ToI32
		case datatypes.F32:
			return I64ToF32
		case datatypes.F64:
			return I64ToF64
		}
	case datatypes.U64:
		switch target {
		case datatypes.I32:
			return U64ToI32
		case datatypes.F32:
			return U64ToF32
		case datatypes.F64:
			return U64ToF64
		}
	case datatypes.F32:
		switch target {
		case datatypes.I32:
			return F32ToI32
		case datatypes.U32:
			return F32ToU32
		case datatypes.I64:
			return F32ToI64
		case datatypes.U64:
			return F32ToU64
		case datatypes.F64:
			return F32ToF64
		}
	case datatypes.F64:
		switch target {
		case datatypes.I32:
			return F64ToI32
		case datatypes.U32:
			return F64ToU32
		case datatypes.I64:
			return F64ToI64
		case datatypes.U64:
			return F64ToU64
		case datatypes.F32:
			return F64ToF32
		}
	}
	return Operation("")
}

func loadVariable(node parser.AST) ([]TAC, Operand) {
	instructions := []TAC{}
	operand := Operand{}
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
	return instructions, operand
}
