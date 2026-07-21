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
	operand := Operand{
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
	// TODO: handle structs
	left := node.Children[0]
	right := node.Children[1]
	l_in, l_op := translateExpression(*left)
	instructions = append(instructions, l_in...)
	r_in, r_op := translateExpression(*right)
	instructions = append(instructions, r_in...)
	fmt.Println(l_op.Type, r_op.Type)
	var irType datatypes.IRType
	var typecast Operation
	if l_op.Type != r_op.Type {
		irType = getHigherType(l_op.Type, r_op.Type)
		if l_op.Type != irType {
			typecast = getTypeCastOperation(l_op.Type, irType)
			cast := formTempVar(irType)
			instructions = append(instructions, Instruction{
				Destination: cast,
				Operation:   typecast,
				Operand1:    l_op,
			})
			l_op = Operand{
				Type: irType,
				Var:  cast,
			}
		} else {
			typecast = getTypeCastOperation(r_op.Type, irType)
			cast := formTempVar(irType)
			instructions = append(instructions, Instruction{
				Destination: cast,
				Operation:   typecast,
				Operand1:    r_op,
			})
			r_op = Operand{
				Type: irType,
				Var:  cast,
			}
		}
	} else {
		irType = l_op.Type
	}
	var operation Operation
	switch node.Token.Value {
	case "==":
		operation = typedOperation(irType, "eq")
	case "!=":
		operation = typedOperation(irType, "ne")
	case "<":
		operation = typedOperation(irType, "lt")
	case "<=":
		operation = typedOperation(irType, "le")
	case ">":
		operation = typedOperation(irType, "gt")
	case ">=":
		operation = typedOperation(irType, "ge")
	}
	tempVar := formTempVar(datatypes.I32)
	instructions = append(instructions, Instruction{
		Destination: tempVar,
		Operation:   operation,
		Operand1:    l_op,
		Operand2:    r_op,
	})
	operand = Operand{
		Type: datatypes.I32,
		Var:  tempVar,
	}
	return instructions, operand
}

func translateBitOperation(node parser.AST) ([]TAC, Operand) {
	instructions := []TAC{}
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
	operand := Operand{
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
	operand := Operand{
		Type: operationType,
		Var:  tempVar,
	}
	return instructions, operand
}

func translateExponent(node parser.AST) ([]TAC, Operand) {
	instructions := []TAC{}
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
	operand := Operand{
		Type: irType,
		Var:  tempVar,
	}
	return instructions, operand
}

func translateUnary(node parser.AST) ([]TAC, Operand) {
	instructions := []TAC{}
	operand := Operand{}
	left := node.Children[0]
	right := node.Children[1]
	if leftTok := left.Token; leftTok.Kind == lexer.OPERATOR_UNARY || leftTok.Value == "-" { // left unary
		r_in, r_op := translateExpression(*right)
		instructions = append(instructions, r_in...)
		switch leftTok.Value {
		case "!":
			tempVar := formTempVar(r_op.Type)
			instructions = append(instructions, Instruction{
				Destination: tempVar,
				Operation:   typedOperation(datatypes.I32, "xor"),
				Operand1:    r_op,
				Operand2: Operand{
					Type:     datatypes.I32,
					Constant: 1,
				},
			})
			operand = Operand{
				Type: datatypes.I32,
				Var:  tempVar,
			}
		case "-":
			zero := getZeroValue(right.Type)
			tempVar := formTempVar(r_op.Type)
			instructions = append(instructions, Instruction{
				Destination: tempVar,
				Operation:   typedOperation(r_op.Type, "sub"),
				Operand1:    zero,
				Operand2:    r_op,
			})
			operand = Operand{
				Type: r_op.Type,
				Var:  tempVar,
			}
		case "~":
			tempVar := formTempVar(r_op.Type)
			instructions = append(instructions, Instruction{
				Destination: tempVar,
				Operation:   typedOperation(r_op.Type, "xor"),
				Operand1:    r_op,
				Operand2: Operand{
					Type:     r_op.Type,
					Constant: -1,
				},
			})
			operand = Operand{
				Type: r_op.Type,
				Var:  tempVar,
			}
		default: // ++, --
			variable := currScope.LookupVariable(right.Token.Value)
			if variable == nil {
				return instructions, operand
			}
			var operation Operation
			switch leftTok.Value {
			case "++":
				operation = typedOperation(r_op.Type, "add")
			case "--":
				operation = typedOperation(r_op.Type, "sub")
			}

			tempVar := formTempVar(r_op.Type)
			operand = Operand{
				Var:  tempVar,
				Type: tempVar.DataType,
			}
			increment := []TAC{
				Instruction{
					Destination: tempVar,
					Operation:   operation,
					Operand1:    r_op,
					Operand2: Operand{
						Type:     r_op.Type,
						Constant: 1,
					},
				},
				Instruction{
					Operation: Store,
					Operand1: Operand{
						Var: Variable{
							Name:       variable.Name,
							DataType:   datatypes.TranslateSourceType(variable.Type),
							Visibility: VariableScope(variable.Ctx),
						},
					},
					Operand2: Operand{
						Var: tempVar,
					},
				},
			}
			instructions = append(instructions, increment...)
		}
	} else { // right unary
		variable := currScope.LookupVariable(left.Token.Value)
		if variable == nil {
			return instructions, operand
		}
		var operation Operation
		l_in, l_op := translateExpression(*left)
		instructions = append(instructions, l_in...)
		switch right.Token.Value {
		case "++":
			operation = typedOperation(l_op.Type, "add")
		case "--":
			operation = typedOperation(l_op.Type, "sub")
		}

		tempVar := formTempVar(l_op.Type)
		operand = l_op
		increment := []TAC{
			Instruction{
				Destination: tempVar,
				Operation:   operation,
				Operand1:    l_op,
				Operand2: Operand{
					Type:     l_op.Type,
					Constant: 1,
				},
			},
			Instruction{
				Operation: Store,
				Operand1: Operand{
					Var: Variable{
						Name:       variable.Name,
						DataType:   datatypes.TranslateSourceType(variable.Type),
						Visibility: VariableScope(variable.Ctx),
					},
				},
				Operand2: Operand{
					Var: tempVar,
				},
			},
		}
		instructions = append(instructions, increment...)
	}
	return instructions, operand
}

func translateTypecast(node parser.AST) ([]TAC, Operand) {
	instructions := []TAC{}
	// TODO: handle toString()
	// TODO: handle struct as source
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
	key := fmt.Sprintf("%s->%s", src, target)
	operations := map[string]Operation{
		"i32->i64": I32ToI64,
		"i32->f32": I32ToF32,
		"i32->f64": I32ToF64,
		"u32->i64": U32ToI64,
		"u32->f32": U32ToF32,
		"u32->f64": U32ToF64,
		"i64->i32": I64ToI32,
		"i64->f32": I64ToF32,
		"i64->f64": I64ToF64,
		"u64->i32": U64ToI32,
		"u64->f32": U64ToF32,
		"u64->f64": U64ToF64,
		"f32->i32": F32ToI32,
		"f32->u32": F32ToU32,
		"f32->i64": F32ToI64,
		"f32->u64": F32ToU64,
		"f32->f64": F32ToF64,
		"f64->i32": F64ToI32,
		"f64->u32": F64ToU32,
		"f64->i64": F64ToI64,
		"f64->u64": F64ToU64,
		"f64->f32": F64ToF32,
	}
	operation, ok := operations[key]
	if !ok {
		return Operation("")
	}
	return operation
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
		Type: tempVar.DataType,
	}
	return instructions, operand
}

func getHigherType(type1, type2 datatypes.IRType) datatypes.IRType {
	if type1 == datatypes.F64 || type2 == datatypes.F64 {
		return datatypes.F64
	}
	if type1 == datatypes.I64 || type2 == datatypes.I64 {
		return datatypes.I64
	}
	if type1 == datatypes.U64 || type2 == datatypes.U64 {
		return datatypes.U64
	}
	if type1 == datatypes.F32 || type2 == datatypes.F32 {
		return datatypes.F32
	}
	if type1 == datatypes.I32 || type2 == datatypes.I32 {
		return datatypes.I32
	}
	return datatypes.U32
}
