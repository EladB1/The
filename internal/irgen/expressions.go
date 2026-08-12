package irgen

import (
	"fmt"

	dt "github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/lexer"
	"github.com/EladB1/The/internal/parser"
	"github.com/EladB1/The/internal/semantic"
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
	} else if node.Label == "index" {
		instructions, operand = translateIndex(node)
	} else if node.Label == "dot" {
		instructions, operand = translateDot(node)
	} else if node.Label == "call" {
		instructions, operand = translateCall(node)
	} else if node.Label == "struct_literal" {
		instructions, operand = translateStructLiteral(node)
	} else if node.IsLiteral() && node.Label != "struct_literal" {
		operand = translateLiteral(node)
	} else if node.Token.Kind == lexer.ID {
		instructions, operand = loadVariable(node)
	}
	return instructions, operand
}

func translateLogicalAndOr(node parser.AST) ([]TAC, Operand) {
	instructions := []TAC{}
	irType := dt.I32
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
		SrcPosition: node.Location,
	})
	operand := Operand{
		Var: Variable{
			Name:     tempVar.Name,
			DataType: irType,
		},
		SrcPosition: node.Location,
	}
	return instructions, operand
}

func translateComparison(node parser.AST) ([]TAC, Operand) {
	instructions := []TAC{}
	operand := Operand{}
	left := node.Children[0]
	right := node.Children[1]
	l_in, l_op := translateExpression(*left)
	instructions = append(instructions, l_in...)
	r_in, r_op := translateExpression(*right)
	instructions = append(instructions, r_in...)
	var irType dt.IRType
	var comp string
	switch node.Token.Value {
	case "==":
		comp = "eq"
	case "!=":
		comp = "ne"
	case "<":
		comp = "lt"
	case "<=":
		comp = "le"
	case ">":
		comp = "gt"
	case ">=":
		comp = "ge"
	}
	if left.Type.IsDynamic {
		scomp_in, scomp := translateStructComparison(l_op, r_op, left.Type.String(), comp)
		instructions = append(instructions, scomp_in...)
		return instructions, scomp
	} else if l_op.Type == dt.Str_const {
		tempVar := formTempVar(dt.I32)
		call := []TAC{
			Instruction{
				Operation:   PrepareParam,
				Operand1:    l_op,
				SrcPosition: left.Location,
			},
			Instruction{
				Operation:   PrepareParam,
				Operand1:    r_op,
				SrcPosition: right.Location,
			},
			Instruction{
				Destination: tempVar,
				Operation:   Call,
				Operand1: Operand{
					Constant: fmt.Sprintf("__str_%s", comp),
				},
				Operand2: Operand{
					Constant: 2,
				},
				SrcPosition: node.Location,
			},
		}
		instructions = append(instructions, call...)
		operand = Operand{
			Type: dt.I32,
			Var:  tempVar,
		}
		return instructions, operand
	} else if l_op.Type != r_op.Type {
		var typecast Operation
		irType = getHigherType(l_op.Type, r_op.Type)
		if l_op.Type != irType {
			typecast = getTypeCastOperation(l_op.Type, irType)
			cast := formTempVar(irType)
			instructions = append(instructions, Instruction{
				Destination: cast,
				Operation:   typecast,
				Operand1:    l_op,
				SrcPosition: left.Location,
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
				SrcPosition: right.Location,
			})
			r_op = Operand{
				Type: irType,
				Var:  cast,
			}
		}
	} else {
		irType = l_op.Type
	}
	operation := typedOperation(irType, comp)
	tempVar := formTempVar(dt.I32)
	instructions = append(instructions, Instruction{
		Destination: tempVar,
		Operation:   operation,
		Operand1:    l_op,
		Operand2:    r_op,
		SrcPosition: node.Location,
	})
	operand = Operand{
		Type: dt.I32,
		Var:  tempVar,
	}
	return instructions, operand
}

func translateBitOperation(node parser.AST) ([]TAC, Operand) {
	instructions := []TAC{}
	rootType := node.Type
	irType := dt.TranslateSourceType(rootType)
	left := node.Children[0]
	right := node.Children[1]
	var operation Operation
	l_in, l_op := translateExpression(*left)
	instructions = append(instructions, l_in...)
	r_in, r_op := translateExpression(*right)
	instructions = append(instructions, r_in...)
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
				SrcPosition: left.Location,
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
				SrcPosition: right.Location,
			})
			r_op = Operand{
				Type: irType,
				Var:  cast,
			}
		}
	} else {
		irType = l_op.Type
	}
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
	tempVar := formTempVar(irType)
	instructions = append(instructions, Instruction{
		Destination: tempVar,
		Operation:   operation,
		Operand1:    l_op,
		Operand2:    r_op,
		SrcPosition: node.Location,
	})
	operand := Operand{
		Type: irType,
		Var:  tempVar,
	}
	return instructions, operand
}

func translateAddition(node parser.AST) ([]TAC, Operand) {
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
	if rootType.Equals(dt.StringType) {
		var fn RuntimeFunction
		if left.Type.Equals(dt.CharType) && right.Type.Equals(dt.CharType) {
			fn = CharConcat
		} else if left.Type.Equals(dt.CharType) && right.Type.Equals(dt.StringType) {
			fn = CharConcatString
		} else if left.Type.Equals(dt.StringType) && right.Type.Equals(dt.CharType) {
			fn = StringConcatChar
		} else { // string + string
			fn = StringConcat
		}
		tempVar := formTempVar(dt.Str_const)
		call := []TAC{
			Instruction{
				Operation:   PrepareParam,
				Operand1:    l_op,
				SrcPosition: left.Location,
			},
			Instruction{
				Operation:   PrepareParam,
				Operand2:    r_op,
				SrcPosition: right.Location,
			},
			Instruction{
				Destination: tempVar,
				Operation:   Call,
				Operand1: Operand{
					Constant: fn,
				},
				Operand2: Operand{
					Constant: 2,
				},
				SrcPosition: node.Location,
			},
		}
		instructions = append(instructions, call...)
		operand = Operand{
			Type: dt.Str_const,
			Var:  tempVar,
		}
	} else {
		var irType dt.IRType
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
					SrcPosition: left.Location,
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
					SrcPosition: right.Location,
				})
				r_op = Operand{
					Type: irType,
					Var:  cast,
				}
			}
		} else {
			irType = l_op.Type
		}
		operationType := dt.TranslateSourceType(rootType)
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
			SrcPosition: node.Location,
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
	var irType dt.IRType
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
				SrcPosition: left.Location,
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
				SrcPosition: right.Location,
			})
			r_op = Operand{
				Type: irType,
				Var:  cast,
			}
		}
	} else {
		irType = l_op.Type
	}
	operationType := dt.TranslateSourceType(rootType)
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
		SrcPosition: node.Location,
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
	irType := dt.TranslateSourceType(rootType)
	left := node.Children[0]
	right := node.Children[1]
	l_in, l_op := translateExpression(*left)
	instructions = append(instructions, l_in...)
	r_in, r_op := translateExpression(*right)
	instructions = append(instructions, r_in...)
	if !left.Type.Equals(rootType) {
		cast := formTempVar(irType)
		instructions = append(instructions, Instruction{
			Destination: cast,
			Operation:   getTypeCastOperation(l_op.Type, irType),
			SrcPosition: left.Location,
		})
		l_op = Operand{
			Type: irType,
			Var:  cast,
		}
	}
	if !right.Type.Equals(rootType) {
		cast := formTempVar(irType)
		instructions = append(instructions, Instruction{
			Destination: cast,
			Operation:   getTypeCastOperation(r_op.Type, irType),
			SrcPosition: right.Location,
		})
		r_op = Operand{
			Type: irType,
			Var:  cast,
		}
	}

	instructions = append(instructions, Instruction{
		Operation:   PrepareParam,
		Operand1:    l_op,
		SrcPosition: left.Location,
	})
	instructions = append(instructions, Instruction{
		Operation:   PrepareParam,
		Operand1:    r_op,
		SrcPosition: right.Location,
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
		SrcPosition: node.Location,
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
	if left.Token.Value == "++" || left.Token.Value == "--" {
		return translateIncrement(left, right, true)
	}
	if leftTok := left.Token; leftTok.Kind == lexer.OPERATOR_UNARY || leftTok.Value == "-" { // left unary
		r_in, r_op := translateExpression(*right)
		instructions = append(instructions, r_in...)
		switch leftTok.Value {
		case "!":
			tempVar := formTempVar(r_op.Type)
			instructions = append(instructions, Instruction{
				Destination: tempVar,
				Operation:   typedOperation(dt.I32, "xor"),
				Operand1:    r_op,
				Operand2: Operand{
					Type:     dt.I32,
					Constant: 1,
				},
				SrcPosition: left.Location,
			})
			operand = Operand{
				Type: dt.I32,
				Var:  tempVar,
			}
		case "-":
			_, zero := getZeroValue(right.Type)
			tempVar := formTempVar(r_op.Type)
			instructions = append(instructions, Instruction{
				Destination: tempVar,
				Operation:   typedOperation(r_op.Type, "sub"),
				Operand1:    zero,
				Operand2:    r_op,
				SrcPosition: left.Location,
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
				SrcPosition: left.Location,
			})
			operand = Operand{
				Type: r_op.Type,
				Var:  tempVar,
			}
		}
	} else { // right unary
		return translateIncrement(left, right, false)
	}
	return instructions, operand
}

func translateIncrement(left, right *parser.AST, isLeft bool) ([]TAC, Operand) {
	instructions := []TAC{}
	var operand Operand
	var value *parser.AST
	var operator string
	if isLeft {
		value = right
		operator = left.Token.Value
	} else {
		value = left
		operator = right.Token.Value
	}

	value_in, value_op := translateExpression(*value)
	instructions = append(instructions, value_in...)
	var operation Operation
	switch operator {
	case "++":
		operation = typedOperation(value_op.Type, "add")
	case "--":
		operation = typedOperation(value_op.Type, "sub")
	}

	var tempVar Variable
	increment := []TAC{}
	if value.Label == "dot" {
		last_in := value_in[len(value_in)-1]
		if last, ok := last_in.(Instruction); ok {
			addr := last.Operand1
			offset := semantic.OffsetValue{}
			if offset_val, ok := last.Operand2.Constant.(uint32); ok {
				offset = semantic.OffsetValue{
					IsSet: true,
					Value: offset_val,
				}
			}
			tempVar = formTempVar(value_op.Type)

			increment = []TAC{
				Instruction{
					Destination: tempVar,
					Operation:   operation,
					Operand1:    value_op,
					Operand2: Operand{
						Type:     value_op.Type,
						Constant: 1,
					},
					SrcPosition: right.Location,
				},
				Instruction{
					Operation: Set,
					Operand1: Operand{
						Var:    addr.Var,
						Offset: offset,
					},
					Operand2: Operand{
						Var: tempVar,
					},
				},
			}
		}
	} else {
		variable := currScope.LookupVariable(value.Token.Value)
		tempVar = formTempVar(value_op.Type)
		increment = []TAC{
			Instruction{
				Destination: tempVar,
				Operation:   operation,
				Operand1:    value_op,
				Operand2: Operand{
					Type:     value_op.Type,
					Constant: 1,
				},
				SrcPosition: right.Location,
			},
			storeVariable(*variable, Operand{
				Var: tempVar,
			}),
		}

	}
	if !isLeft {
		operand = value_op
	} else {
		operand = Operand{
			Var:  tempVar,
			Type: tempVar.DataType,
		}
	}
	instructions = append(instructions, increment...)
	return instructions, operand
}

func translateTypecast(node parser.AST) ([]TAC, Operand) {
	instructions := []TAC{}
	var tempVar Variable
	var irName string
	irType := dt.TranslateSourceType(node.Children[0].Type)
	l_in, l_op := translateExpression(*node.Children[0])
	instructions = append(instructions, l_in...)
	targetType := dt.TranslateSourceType(node.Type)
	if sourceType := node.Children[0].Type; sourceType.IsDynamic {
		if node.Type.Equals(dt.StringType) {
			irName = getStructToString(sourceType.String())
		} else {
			if str := currScope.LookupStruct(sourceType.String()); str != nil {
				if castBlock := str.InnerScope.LookupNamedBlock("cast"); castBlock != nil {
					fn := castBlock.InnerScope.LookupFunctionsByReturnType(node.Type)
					if len(fn) != 0 {
						overload := fn[0].Overloads[0]
						irName = overload.IRName
					}
				}
			}
		}
		tempVar = formTempVar(targetType)
		call := []TAC{
			Instruction{
				Operation:   PrepareParam,
				Operand1:    l_op,
				SrcPosition: node.Children[0].Location,
			},
			Instruction{
				Destination: tempVar,
				Operation:   Call,
				Operand1: Operand{
					Constant: irName,
				},
				Operand2: Operand{
					Constant: 1,
				},
				SrcPosition: node.Location,
			},
		}
		instructions = append(instructions, call...)
	} else if targetType == dt.Str_const {
		fn := getToStringFn(node.Children[0].Type)
		tempVar = formTempVar(dt.Str_const)
		call := []TAC{
			Instruction{
				Operation:   PrepareParam,
				Operand1:    l_op,
				SrcPosition: node.Children[0].Location,
			},
			Instruction{
				Destination: tempVar,
				Operation:   Call,
				Operand1: Operand{
					Constant: fn,
				},
				Operand2: Operand{
					Constant: 1,
				},
				SrcPosition: node.Location,
			},
		}
		instructions = append(instructions, call...)
	} else {
		operation := getTypeCastOperation(irType, targetType)
		tempVar = formTempVar(targetType)
		instructions = append(instructions, Instruction{
			Destination: tempVar,
			Operation:   operation,
			Operand1:    l_op,
			SrcPosition: node.Location,
		})
	}
	operand := Operand{
		Type: targetType,
		Var:  tempVar,
	}
	return instructions, operand
}

func translateIndex(node parser.AST) ([]TAC, Operand) {
	instructions := []TAC{}
	left := node.Children[0]
	right := node.Children[1]
	l_in, container_op := translateExpression(*left)
	instructions = append(instructions, l_in...)
	var r_in []TAC
	var r_op Operand
	switch right.Label {
	case "slice":
		r_in, r_op = translateSlice(*right, container_op)
		return r_in, r_op
	case "ARR-END":
		r_in, r_op = translateArrayEnd(*right, container_op)
	default:
		r_in, r_op = translateExpression(*right)
	}
	instructions = append(instructions, r_in...)
	if r_op.Type != dt.I32 {
		cast := formTempVar(dt.I32)
		typecast := []TAC{
			Instruction{
				Destination: cast,
				Operation:   getTypeCastOperation(r_op.Type, dt.I32),
				Operand1:    r_op,
				SrcPosition: right.Location,
			},
		}
		instructions = append(instructions, typecast...)
		r_op = Operand{
			Type: dt.I32,
			Var:  cast,
		}
	}
	tempVar := formTempVar(dt.I32)
	call := []TAC{
		Instruction{
			Operation:   PrepareParam,
			Operand1:    container_op,
			SrcPosition: left.Location,
		},
		Instruction{
			Operation:   PrepareParam,
			Operand1:    r_op,
			SrcPosition: right.Location,
		},
		Instruction{
			Destination: tempVar,
			Operation:   Call,
			Operand1: Operand{
				Constant: StringIndex,
			},
			Operand2: Operand{
				Constant: 2,
			},
			SrcPosition: node.Location,
		},
	}
	instructions = append(instructions, call...)
	operand := Operand{
		Type: dt.I32,
		Var:  tempVar,
	}
	return instructions, operand
}

func translateArrayEnd(node parser.AST, arr Operand) ([]TAC, Operand) {
	instructions := []TAC{}
	in, op := translateExpression(*node.Children[0])
	instructions = append(instructions, in...)
	len_in, len := getArrayLength(arr)
	instructions = append(instructions, len_in...)
	if op.Type != dt.I32 {
		operation := getTypeCastOperation(op.Type, dt.I32)
		cast := formTempVar(dt.I32)
		instructions = append(instructions, Instruction{
			Destination: cast,
			Operation:   operation,
			Operand1:    op,
			SrcPosition: node.Children[0].Location,
		})
		op = Operand{
			Type: dt.I32,
			Var:  cast,
		}
	}
	tempVar := formTempVar(dt.I32)
	sub := Instruction{
		Destination: tempVar,
		Operation:   typedOperation(dt.I32, "sub"),
		Operand1:    len,
		Operand2:    op,
		SrcPosition: node.Children[0].Location,
	}
	instructions = append(instructions, sub)
	operand := Operand{
		Type: dt.I32,
		Var:  tempVar,
	}
	return instructions, operand
}

func translateSlice(node parser.AST, arr Operand) ([]TAC, Operand) {
	instructions := []TAC{}
	var operand Operand
	var slice Variable
	start_in := []TAC{}
	end_in := []TAC{}
	_, start_op := getZeroValue(dt.Int32Type)
	var end_op Operand
	rangeIndex := 0
	length := len(node.Children)
	fn := StringSlice
	switch length {
	case 1: // str[..]
		end_in, end_op = getArrayEnd(arr)
		instructions = append(instructions, end_in...)
	case 2:
		if node.Children[0].Token.Kind == lexer.OPERATOR_RANGE { // str[..1]
			rangeIndex = 0
			end_in, end_op = translateExpression(*node.Children[1])
			instructions = append(instructions, end_in...)
			if end_op.Type != dt.I32 {
				cast := formTempVar(dt.I32)
				instructions = append(instructions, Instruction{
					Destination: cast,
					Operation:   getTypeCastOperation(end_op.Type, dt.I32),
					Operand1:    end_op,
					SrcPosition: node.Children[1].Location,
				})
				end_op = Operand{
					Type: dt.I32,
					Var:  cast,
				}
			}
		} else { // str[1..]
			rangeIndex = 1
			start_in, start_op = translateExpression(*node.Children[0])
			instructions = append(instructions, start_in...)
			if start_op.Type != dt.I32 {
				cast := formTempVar(dt.I32)
				instructions = append(instructions, Instruction{
					Destination: cast,
					Operation:   getTypeCastOperation(start_op.Type, dt.I32),
					Operand1:    start_op,
					SrcPosition: node.Children[0].Location,
				})
				start_op = Operand{
					Type: dt.I32,
					Var:  cast,
				}
			}
			end_in, end_op = getArrayEnd(arr)
			instructions = append(instructions, end_in...)
		}
	case 3: // str[1..5]
		rangeIndex = 1
		start_in, start_op = translateExpression(*node.Children[0])
		instructions = append(instructions, start_in...)
		if start_op.Type != dt.I32 {
			cast := formTempVar(dt.I32)
			instructions = append(instructions, Instruction{
				Destination: cast,
				Operation:   getTypeCastOperation(start_op.Type, dt.I32),
				Operand1:    start_op,
				SrcPosition: node.Children[0].Location,
			})
			start_op = Operand{
				Type: dt.I32,
				Var:  cast,
			}
		}
		end_in, end_op = translateExpression(*node.Children[2])
		instructions = append(instructions, end_in...)
		if end_op.Type != dt.I32 {
			cast := formTempVar(dt.I32)
			instructions = append(instructions, Instruction{
				Destination: cast,
				Operation:   getTypeCastOperation(end_op.Type, dt.I32),
				Operand1:    end_op,
				SrcPosition: node.Children[2].Location,
			})
			end_op = Operand{
				Type: dt.I32,
				Var:  cast,
			}
		}
	}
	slice = formTempVar(dt.I32)
	if node.Children[rangeIndex].Token.Value == "..=" {
		fn = StringSliceInclusive
	}
	call := []TAC{
		Instruction{
			Operation:   PrepareParam,
			Operand1:    arr,
			SrcPosition: arr.SrcPosition,
		},
		Instruction{
			Operation:   PrepareParam,
			Operand1:    start_op,
			SrcPosition: start_op.SrcPosition,
		},
		Instruction{
			Operation:   PrepareParam,
			Operand1:    end_op,
			SrcPosition: end_op.SrcPosition,
		},
		Instruction{
			Destination: slice,
			Operation:   Call,
			Operand1: Operand{
				Constant: fn,
			},
			Operand2: Operand{
				Constant: 3,
			},
			SrcPosition: node.Location,
		},
	}
	instructions = append(instructions, call...)
	operand = Operand{
		Type: dt.I32,
		Var:  slice,
	}
	return instructions, operand
}

func translateDot(node parser.AST) ([]TAC, Operand) {
	instructions := []TAC{}
	operand := Operand{}
	left := node.Children[0]
	prop := node.Children[1]
	if mem, ok := semantic.PrimitiveMembers[left.Type.Root]; ok {
		if pr, ok := mem.Properties[prop.Token.Value]; ok {
			fn := builtinPropToFunction(left.Type, pr)
			l_in, l_op := translateExpression(*left)
			instructions = append(instructions, l_in...)
			irType := dt.TranslateSourceType(pr.Type)
			result := formTempVar(irType)
			call := []TAC{
				Instruction{
					Operation:   PrepareParam,
					Operand1:    l_op,
					SrcPosition: left.Location,
				},
				Instruction{
					Destination: result,
					Operation:   Call,
					Operand1: Operand{
						Constant: fn,
					},
					Operand2: Operand{
						Constant: 1,
					},
					SrcPosition: node.Location,
				},
			}
			instructions = append(instructions, call...)
			operand = Operand{
				Type: irType,
				Var:  result,
			}
		}
	} else if left.Type.Equals(dt.GlobalRefType) {
		scope := currScope
		currScope = globalScope
		prop_in, prop_op := translateExpression(*prop)
		instructions = append(instructions, prop_in...)
		operand = prop_op
		currScope = scope
	} else if left.Type.RootEquals(dt.Ref) {
		prop_in, prop_op := translateExpression(*prop)
		instructions = append(instructions, prop_in...)
		operand = prop_op
	} else if prop.Type.RootEquals(dt.ScopeRef) {
		ptr_in, ptr := translateExpression(*left)
		instructions = append(instructions, ptr_in...)
		operand = ptr
	} else {
		//loadStruct := formTempVar(dt.TranslateSourceType(left.Type))
		ptr_in, ptr := translateExpression(*left)
		instructions = append(instructions, ptr_in...)
		str := currScope.LookupStruct(string(left.Type.Root))
		if str == nil {
		}
		propSymbol := str.InnerScope.LookupVariable(prop.Token.Value)
		loadProp := formTempVar(dt.TranslateSourceType(propSymbol.Type))
		instructions = append(instructions, Instruction{
			Destination: loadProp,
			Operation:   Load,
			Operand1:    ptr,
			Operand2: Operand{
				Constant: propSymbol.Offset.Value,
			},
			SrcPosition: node.Location,
		})
		operand = Operand{
			Type: loadProp.DataType,
			Var:  loadProp,
		}
	}
	return instructions, operand
}

func translateCall(node parser.AST) ([]TAC, Operand) {
	instructions := []TAC{}
	operand := Operand{}
	var nameNode *parser.AST

	irParamTypes := []dt.IRType{}
	srcParamTypes := []dt.SourceType{}
	loadParams := []TAC{}
	var object *parser.AST = nil
	if node.Children[0].Label == "dot" {
		nameNode = node.Children[0].Children[1]
		object = node.Children[0].Children[0]
	} else {
		nameNode = node.Children[0]
	}
	name := nameNode.Token.Value
	irName := name
	if nameNode.IRName != "" {
		irName = nameNode.IRName
	}
	if len(node.Children) == 2 {
		params := node.Children[1].Children

		for _, param := range params {
			param_in, param_op := translateExpression(*param)
			instructions = append(instructions, param_in...)
			srcParamTypes = append(srcParamTypes, param.Type)
			irParamTypes = append(irParamTypes, dt.TranslateSourceType(param.Type))
			loadParams = append(loadParams, Instruction{
				Operation:   PrepareParam,
				Operand1:    param_op,
				SrcPosition: param.Location,
			})
		}
	}
	if node.Children[0].Label == "dot" {
		var obj_in []TAC
		var obj Operand
		if object.Type.IsDynamic {
			irParamTypes = append(irParamTypes, dt.Ptr)
			obj_in, obj = loadVariable(*object)
		} else {
			irParamTypes = append(irParamTypes, dt.TranslateSourceType(object.Type))
			obj_in, obj = translateExpression(*object)
		}
		instructions = append(instructions, obj_in...)
		loadParams = append(loadParams, Instruction{
			Operation:   PrepareParam,
			Operand1:    obj,
			SrcPosition: obj.SrcPosition,
		})
	}
	symbol := currScope.LookupFunctionByName(name)

	returnType := dt.NoneIR
	if symbol != nil {
		returnType = dt.TranslateSourceType(symbol.ReturnType)
	} else if object != nil && !object.Type.IsDynamic {
		if prim, ok := semantic.PrimitiveMembers[object.Type.Root]; ok {
			if method, ok := prim.Methods[name]; ok {
				returnType = dt.TranslateSourceType(method.ReturnType)
			}
		}
	}
	instructions = append(instructions, loadParams...)
	if returnType == dt.NoneIR {
		instructions = append(instructions, Instruction{
			Operation: Call,
			Operand1: Operand{
				Constant: irName,
			},
			Operand2: Operand{
				Constant: len(irParamTypes),
			},
			SrcPosition: node.Location,
		})
		return instructions, operand
	}
	result := formTempVar(returnType)
	instructions = append(instructions, Instruction{
		Destination: result,
		Operation:   Call,
		Operand1: Operand{
			Constant: irName,
		},
		Operand2: Operand{
			Constant: len(irParamTypes),
		},
		SrcPosition: node.Location,
	})
	operand = Operand{
		Type: result.DataType,
		Var:  result,
	}
	return instructions, operand
}

func getTypeCastOperation(src dt.IRType, target dt.IRType) Operation {
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
	if variable.Ctx == semantic.StructProp {
		get := formTempVar(dt.Ptr)
		instructions = append(instructions, Instruction{
			Destination: get,
			Operation:   Get,
			Operand1: Operand{
				Var: Variable{
					Name:       "__this",
					DataType:   dt.Ptr,
					Visibility: Param,
				},
			},
		})
		load := formTempVar(dt.TranslateSourceType(variable.Type))
		instructions = append(instructions, Instruction{
			Destination: load,
			Operation:   Load,
			Operand1: Operand{
				Type: load.DataType,
				Var:  get,
			},
			Operand2: Operand{
				Constant: variable.Offset.Value,
			},
			SrcPosition: node.Location,
		})
		return instructions, Operand{
			Var:         load,
			Type:        load.DataType,
			SrcPosition: node.Location,
		}
	}
	varType := dt.TranslateSourceType(variable.Type)
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
		SrcPosition: node.Location,
	})
	operand = Operand{
		Var: Variable{
			Name:     tempVar.Name,
			DataType: tempVar.DataType,
		},
		Type:        tempVar.DataType,
		SrcPosition: node.Location,
	}
	return instructions, operand
}

func getHigherType(type1, type2 dt.IRType) dt.IRType {
	if type1 == dt.F64 || type2 == dt.F64 {
		return dt.F64
	}
	if type1 == dt.I64 || type2 == dt.I64 {
		return dt.I64
	}
	if type1 == dt.U64 || type2 == dt.U64 {
		return dt.U64
	}
	if type1 == dt.F32 || type2 == dt.F32 {
		return dt.F32
	}
	if type1 == dt.I32 || type2 == dt.I32 {
		return dt.I32
	}
	return dt.U32
}

func getArrayEnd(arr Operand) ([]TAC, Operand) {
	instructions, len := getArrayLength(arr)
	arrayEnd := formTempVar(dt.I32)
	instructions = append(instructions, Instruction{
		Destination: arrayEnd,
		Operation:   typedOperation(dt.I32, "sub"),
		Operand1:    len,
		Operand2: Operand{
			Type:     dt.I32,
			Constant: 1,
		},
		SrcPosition: arr.SrcPosition,
	})
	operand := Operand{
		Var:         arrayEnd,
		SrcPosition: arr.SrcPosition,
	}
	return instructions, operand
}

func getArrayLength(arr Operand) ([]TAC, Operand) {
	len := formTempVar(dt.I32)
	instructions := []TAC{
		Instruction{
			Operation:   PrepareParam,
			Operand1:    arr,
			SrcPosition: arr.SrcPosition,
		},
		Instruction{
			Destination: len,
			Operation:   Call,
			Operand1: Operand{
				Constant: Stringlen,
			},
			Operand2: Operand{
				Constant: 1,
			},
			SrcPosition: arr.SrcPosition,
		},
	}
	operand := Operand{
		Type:        dt.I32,
		Var:         len,
		SrcPosition: arr.SrcPosition,
	}
	return instructions, operand
}

func getToStringFn(src dt.SourceType) RuntimeFunction {
	if src.Equals(dt.Int32Type) || src.Equals(dt.Uint32Type) {
		return StringFromInt32
	} else if src.Equals(dt.Int64Type) || src.Equals(dt.Uint64Type) {
		return StringFromInt64
	} else if src.Equals(dt.BoolType) {
		return StringFromBool
	} else if src.Equals(dt.CharType) {
		return StringFromChar
	} else if src.Equals(dt.FloatType) {
		return StringFromFloat32
	} else if src.Equals(dt.DoubleType) {
		return StringFromFloat64
	}
	return ""
}

func builtinPropToFunction(object dt.SourceType, vs semantic.VariableSymbol) string {
	prefix := ""
	if object.Equals(dt.StringType) {
		prefix = "str"
	}
	return fmt.Sprintf("__%s_%s", prefix, vs.Name)
}
