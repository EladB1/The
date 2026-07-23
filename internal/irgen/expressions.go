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
	} else if node.IsLiteral() {
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
		// TODO: handle structs
	} else if l_op.Type == dt.Str_const {
		tempVar := formTempVar(dt.I32)
		call := []TAC{
			Instruction{
				Operation: PrepareParam,
				Operand1:  l_op,
			},
			Instruction{
				Operation: PrepareParam,
				Operand1:  r_op,
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
	operation := typedOperation(irType, comp)
	tempVar := formTempVar(dt.I32)
	instructions = append(instructions, Instruction{
		Destination: tempVar,
		Operation:   operation,
		Operand1:    l_op,
		Operand2:    r_op,
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
		var fn string
		if left.Type.Equals(dt.CharType) && right.Type.Equals(dt.CharType) {
			fn = "__char_concat"
		} else if left.Type.Equals(dt.CharType) && right.Type.Equals(dt.StringType) {
			fn = "__char_concat_str"
		} else if left.Type.Equals(dt.StringType) && right.Type.Equals(dt.CharType) {
			fn = "__str_concat_char"
		} else { // string + string
			fn = "__str_concat"
		}
		tempVar := formTempVar(dt.Str_const)
		call := []TAC{
			Instruction{
				Operation: PrepareParam,
				Operand1:  l_op,
			},
			Instruction{
				Operation: PrepareParam,
				Operand2:  r_op,
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
		})
		r_op = Operand{
			Type: irType,
			Var:  cast,
		}
	}

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
				Operation:   typedOperation(dt.I32, "xor"),
				Operand1:    r_op,
				Operand2: Operand{
					Type:     dt.I32,
					Constant: 1,
				},
			})
			operand = Operand{
				Type: dt.I32,
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
							DataType:   dt.TranslateSourceType(variable.Type),
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
						DataType:   dt.TranslateSourceType(variable.Type),
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
	var tempVar Variable
	sourceType := dt.TranslateSourceType(node.Children[0].Type)
	l_in, l_op := translateExpression(*node.Children[0])
	instructions = append(instructions, l_in...)
	// TODO: handle struct as source
	targetType := dt.TranslateSourceType(node.Type)
	if targetType == dt.Str_const {
		fn := getToStringFn(node.Children[0].Type)
		tempVar = formTempVar(dt.Str_const)
		call := []TAC{
			Instruction{
				Operation: PrepareParam,
				Operand1:  l_op,
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
			},
		}
		instructions = append(instructions, call...)
	} else {
		operation := getTypeCastOperation(sourceType, targetType)
		tempVar = formTempVar(targetType)
		instructions = append(instructions, Instruction{
			Destination: tempVar,
			Operation:   operation,
			Operand1:    l_op,
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
			Operation: PrepareParam,
			Operand1:  container_op,
		},
		Instruction{
			Operation: PrepareParam,
			Operand1:  r_op,
		},
		Instruction{
			Destination: tempVar,
			Operation:   Call,
			Operand1: Operand{
				Constant: "__str_index",
			},
			Operand2: Operand{
				Constant: 2,
			},
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
	start_op := getZeroValue(dt.Int32Type)
	var end_op Operand
	rangeIndex := 0
	length := len(node.Children)
	fn := "__str_slice"
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
			})
			end_op = Operand{
				Type: dt.I32,
				Var:  cast,
			}
		}
	}
	slice = formTempVar(dt.I32)
	if node.Children[rangeIndex].Token.Value == "..=" {
		fn = "__str_slice_inclusive"
	}
	call := []TAC{
		Instruction{
			Operation: PrepareParam,
			Operand1:  arr,
		},
		Instruction{
			Operation: PrepareParam,
			Operand1:  start_op,
		},
		Instruction{
			Operation: PrepareParam,
			Operand1:  end_op,
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
	if left.Label == "dot" {

	} else if mem, ok := semantic.PrimitiveMembers[left.Type.Root]; ok {
		if pr, ok := mem.Properties[prop.Token.Value]; ok {
			fn := builtinPropToFunction(left.Type, pr)
			l_in, l_op := translateExpression(*left)
			instructions = append(instructions, l_in...)
			irType := dt.TranslateSourceType(pr.Type)
			result := formTempVar(irType)
			call := []TAC{
				Instruction{
					Operation: PrepareParam,
					Operand1:  l_op,
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
				},
			}
			instructions = append(instructions, call...)
			operand = Operand{
				Type: irType,
				Var:  result,
			}
		}
	} else if left.Type.Equals(dt.GlobalRefType) {

	} else if left.Type.RootEquals(dt.Ref) {

	} else {
		// struct
	}
	return instructions, operand
}

func translateCall(node parser.AST) ([]TAC, Operand) {
	instructions := []TAC{}
	operand := Operand{}
	name := node.Children[0].Token.Value
	irParamTypes := []dt.IRType{}
	srcParamTypes := []dt.SourceType{}
	loadParams := []TAC{}
	for _, param := range node.Children[1].Children {
		param_in, param_op := translateExpression(*param)
		instructions = append(instructions, param_in...)
		srcParamTypes = append(srcParamTypes, param.Type)
		irParamTypes = append(irParamTypes, dt.TranslateSourceType(param.Type))
		loadParams = append(loadParams, Instruction{
			Operation: PrepareParam,
			Operand1:  param_op,
		})
	}
	symbol := currScope.LookupFunctionByName(name)
	irName := ""
	if symbol != nil {
		if len(symbol.Overloads) == 1 {
			irName = name
		} else {
			overload := symbol.GetMatchingOverload(srcParamTypes)
			if overload != nil {
				irName = overload.IRName
			}
		}
	}
	instructions = append(instructions, loadParams...)
	result := formTempVar(dt.TranslateSourceType(symbol.ReturnType))
	instructions = append(instructions, Instruction{
		Destination: result,
		Operation:   Call,
		Operand1: Operand{
			Constant: irName,
		},
		Operand2: Operand{
			Constant: len(srcParamTypes),
		},
	})
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
	})
	operand := Operand{
		Var: arrayEnd,
	}
	return instructions, operand
}

func getArrayLength(arr Operand) ([]TAC, Operand) {
	len := formTempVar(dt.I32)
	instructions := []TAC{
		Instruction{
			Operation: PrepareParam,
			Operand1:  arr,
		},
		Instruction{
			Destination: len,
			Operation:   Call,
			Operand1: Operand{
				Constant: "__str_length",
			},
			Operand2: Operand{
				Constant: 1,
			},
		},
	}
	operand := Operand{
		Type: dt.I32,
		Var:  len,
	}
	return instructions, operand
}

func getToStringFn(src dt.SourceType) string {
	// TODO: handle struct
	if src.Equals(dt.Int32Type) || src.Equals(dt.Uint32Type) {
		return "__str_fromInt32"
	} else if src.Equals(dt.Int64Type) || src.Equals(dt.Uint64Type) {
		return "__str_fromInt64"
	} else if src.Equals(dt.BoolType) {
		return "__str_fromBool"
	} else if src.Equals(dt.CharType) {
		return "__str_fromChar"
	} else if src.Equals(dt.FloatType) {
		return "__str_fromFloat32"
	} else if src.Equals(dt.DoubleType) {
		return "__str_fromFloat64"
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
