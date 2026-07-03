package semantic

import (
	"fmt"
	"slices"

	ds "github.com/EladB1/The/internal/datastructures"
	"github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/diagnostic"
	"github.com/EladB1/The/internal/lexer"
	"github.com/EladB1/The/internal/parser"
)

var unsignedTypes []datatypes.DataType = []datatypes.DataType{datatypes.Uint32, datatypes.Uint64}
var signedIntTypes []datatypes.DataType = []datatypes.DataType{datatypes.Int32, datatypes.Int64}
var floatTypes []datatypes.DataType = []datatypes.DataType{datatypes.Float, datatypes.Double}
var numericTypes []datatypes.DataType = slices.Concat(unsignedTypes, signedIntTypes, floatTypes)

func evalLiteral(ast *parser.AST, expectedType datatypes.DataType) datatypes.DataType {
	if ast.Label == "struct_literal" {
		return evalStructLiteral(ast)
	}
	switch ast.Token.Kind {
	case lexer.LIT_CHAR:
		ast.Type = datatypes.Char
		return datatypes.Char
	case lexer.LIT_STRING:
		return datatypes.String
	case lexer.LIT_FLOAT:
		if expectedType == datatypes.Double {
			return datatypes.Double
		}
		return datatypes.Float
	case lexer.LIT_INT, lexer.LIT_HEX:
		if slices.Contains(numericTypes, expectedType) {
			return expectedType
		}
		return datatypes.Int32
	case lexer.KW_BOOLVALUE:
		return datatypes.Bool
	}
	return datatypes.None
}

func evalStructLiteral(ast *parser.AST) datatypes.DataType {
	name := ast.Children[0].Token.Value
	symbol := globalScope.lookupStruct(name)
	if symbol == nil {
		messages = messages.Complain(diagnostic.NameError, fmt.Sprintf("Struct %s not defined", name), ast.Location)
		return datatypes.None
	}
	properties := ast.Children[1].Children
	visited := ds.HashSet{}
	for _, prop := range properties {
		propId := prop.Children[0].Token.Value
		var innerScope *Scope
		if privateBlock := symbol.innerScope.lookupNamedBlock("private"); privateBlock != nil {
			innerScope = privateBlock.innerScope
		} else {
			innerScope = symbol.innerScope
		}
		property := innerScope.lookupVariable(propId)
		if property == nil {
			messages = messages.Complain(diagnostic.NameError, fmt.Sprintf("Could not find property '%s' in struct %s", propId, symbol.name), prop.Location)
			continue
		}
		if _, ok := visited[propId]; ok {
			messages = messages.Complain(diagnostic.IllegalStatementError, fmt.Sprintf("Cannot define struct property '%s' multiple times in struct literal", propId), prop.Location)
			continue
		}
		value := prop.Children[1]
		if valueType := evalType(&value, property.Type); property.Type != valueType {
			messages = messages.Complain(diagnostic.TypeError, fmt.Sprintf("Property type %s expected but found %s", property.Type, valueType), value.Location)
		}
		visited.Append(propId)
	}
	return datatypes.DynamicType(name)
}

func evalType(ast *parser.AST, expectedType datatypes.DataType) datatypes.DataType {
	var nodeType datatypes.DataType = datatypes.None
	var err error = nil
	if ast.IsLiteral() {
		nodeType = evalLiteral(ast, expectedType)
	} else if ast.Token.Kind == lexer.ID {
		symbol := currentScope.lookupVariable(ast.Token.Value)
		if symbol != nil {
			nodeType = symbol.Type
		} else {
			messages = messages.Complain(diagnostic.NameError, fmt.Sprintf("Variable '%s' not defined in this scope", ast.Token.Value), ast.Location)
		}
	} else if ast.Label == "Unary" {
		if leftTok := ast.Children[0].Token; leftTok.Kind == lexer.OPERATOR_UNARY || leftTok.Value == "-" {
			rhs := evalType(&ast.Children[1], datatypes.None)
			switch leftTok.Value {
			case "!":
				if rhs != datatypes.Bool {
					messages = messages.Complain(diagnostic.TypeError, "Must use bool value with unary '!'", ast.Location)
				} else {
					nodeType = rhs
				}
			case "-":
				if !slices.Contains(numericTypes, rhs) {
					messages = messages.Complain(diagnostic.TypeError, fmt.Sprintf("Cannot use unary '-' with type %s", rhs), ast.Location)
				} else {
					nodeType = rhs
				}
			case "~":
				if !slices.Contains(unsignedTypes, rhs) && !slices.Contains(signedIntTypes, rhs) {
					messages = messages.Complain(diagnostic.TypeError, fmt.Sprintf("Cannot negate value of type %s", rhs), ast.Location)
				} else {
					nodeType = rhs
				}
			default: // ++, --
				operand := ast.Children[1].Token
				symbol := currentScope.lookupVariable(operand.Value)
				if symbol != nil {
					if !slices.Contains(numericTypes, symbol.Type) {
						messages = messages.Complain(diagnostic.TypeError, fmt.Sprintf("Cannot use '%s' with type %s", ast.Children[0].Token.Value, symbol.Type), operand.Location)
					} else {
						nodeType = symbol.Type
					}
				} else {
					messages = messages.Complain(diagnostic.NameError, fmt.Sprintf("Could not find variable with name %s", operand.Value), operand.Location)
				}
			}
		} else {
			operand := ast.Children[0].Token
			symbol := currentScope.lookupVariable(operand.Value)
			if symbol != nil {
				if !slices.Contains(numericTypes, symbol.Type) {
					messages = messages.Complain(diagnostic.TypeError, fmt.Sprintf("Cannot use '%s' with type %s", ast.Children[1].Token.Value, symbol.Type), operand.Location)
				} else {
					nodeType = symbol.Type
				}
			} else {
				messages = messages.Complain(diagnostic.NameError, fmt.Sprintf("Could not find variable with name %s", operand.Value), operand.Location)
			}
		}

	} else if ast.Label == "typecast" {
		lhs := evalType(&ast.Children[0], datatypes.None)
		target := nodeToType(ast.Children[1])
		if lhs != target && target != datatypes.String {
			if lhs == datatypes.Uint64 {
				if target == datatypes.Uint32 || target == datatypes.Float || target == datatypes.Int32 {
					messages = messages.Warn(fmt.Sprintf("Lossy conversion from %s to %s", lhs, target), ast.Location)
				} else if !slices.Contains(numericTypes, target) {
					messages = messages.Complain(diagnostic.CastError, fmt.Sprintf("Typecasting from %s to %s not allowed", lhs, target), ast.Location)
				}
			} else if lhs == datatypes.Int64 {
				if target == datatypes.Uint32 || target == datatypes.Float || target == datatypes.Int32 {
					messages = messages.Warn(fmt.Sprintf("Lossy conversion from %s to %s", lhs, target), ast.Location)
				} else if !slices.Contains(numericTypes, target) {
					messages = messages.Complain(diagnostic.CastError, fmt.Sprintf("Typecasting from %s to %s not allowed", lhs, target), ast.Location)
				}
			} else if lhs == datatypes.String {
				messages = messages.Complain(diagnostic.CastError, fmt.Sprintf("Typecasting from %s to %s not allowed", lhs, target), ast.Location)
			} else if lhs == datatypes.Double {
				if target == datatypes.Uint32 || target == datatypes.Float || target == datatypes.Int32 {
					messages = messages.Warn(fmt.Sprintf("Lossy conversion from %s to %s", lhs, target), ast.Location)
				} else if !slices.Contains(numericTypes, target) {
					messages = messages.Complain(diagnostic.CastError, fmt.Sprintf("Typecasting from %s to %s not allowed", lhs, target), ast.Location)
				}
			} else if !lhs.IsPrimitive() {
				// look for cast function
				str := globalScope.lookupStruct(lhs.String())
				if str == nil {
					messages = messages.Complain(diagnostic.CastError, fmt.Sprintf("Cannot typecast %s to %s. Could not find definition of '%s'", lhs, target, lhs), ast.Location)
				} else {
					castBlock := str.innerScope.lookupNamedBlock("cast")
					if castBlock == nil {
						messages = messages.Complain(diagnostic.CastError, fmt.Sprintf("Cannot typecast %s to %s. To support typecasting add a cast block with a function returning the target type", lhs, target), ast.Location)
					} else if !castBlock.HasReturnType(target) {
						messages = messages.Complain(diagnostic.CastError, fmt.Sprintf("Cannot typecast %s to %s. To support this typecasting add a function returning target type to cast block", lhs, target), ast.Location)
					}
				}
			} else if slices.Contains(numericTypes, lhs) && slices.Contains(numericTypes, target) {
				return target
			} else {
				messages = messages.Complain(diagnostic.CastError, fmt.Sprintf("Typecasting from %s to %s not allowed", lhs, target), ast.Location)
			}
		}
		return target

	} else if ast.Label == "index" {
		// only supports one index until arrays added
		if len(ast.Children) > 2 {
			messages = messages.Warn("Multiple indexes not yet supported", ast.Location)
		}
		lhs := evalType(&ast.Children[0], datatypes.None)
		rhs := evalType(&ast.Children[1], datatypes.None)
		if ast.Children[1].Label == "slice" {
			nodeType = datatypes.String
		} else {
			if lhs != datatypes.String && (!slices.Contains(signedIntTypes, rhs) && !slices.Contains(unsignedTypes, rhs)) {
				messages = messages.Complain(diagnostic.TypeError, "Cannot index String with %s type", ast.Location)
			}
			nodeType = datatypes.Char
		}
	} else if ast.Label == "slice" {
		length := len(ast.Children)
		switch length {
		case 1:
			nodeType = datatypes.Int32
		case 2:
			var expr parser.AST
			if ast.Children[0].Token.Kind == lexer.OPERATOR_RANGE {
				expr = ast.Children[1]
			} else {
				expr = ast.Children[0]
			}
			operand := evalType(&expr, datatypes.None)
			if !slices.Contains(signedIntTypes, operand) && !slices.Contains(unsignedTypes, operand) {
				messages = messages.Complain(diagnostic.TypeError, fmt.Sprintf("Invalid type %s used in range expression", operand), expr.Location)
			} else {
				nodeType = operand
			}
		case 3:
			lhs := evalType(&ast.Children[0], datatypes.None)
			rhs := evalType(&ast.Children[2], datatypes.None)
			if (!slices.Contains(signedIntTypes, lhs) && !slices.Contains(unsignedTypes, lhs)) || (!slices.Contains(signedIntTypes, rhs) && !slices.Contains(unsignedTypes, rhs)) {
				messages = messages.Complain(diagnostic.TypeError, fmt.Sprintf("Both sides of slice expression must be an int type; got %s and %s", lhs, rhs), ast.Location)
			} else {
				nodeType, err = decideNumberType(lhs, rhs, ast.Children[1].Token.Value)
				if err != nil {
					messages = messages.Complain(diagnostic.TypeError, err.Error(), ast.Location)
				}
			}
		}

	} else if ast.Label == "ARR-END" {
		expr := evalType(&ast.Children[0], datatypes.None)
		if !slices.Contains(unsignedTypes, expr) && !slices.Contains(signedIntTypes, expr) {
			messages = messages.Complain(diagnostic.TypeError, fmt.Sprintf("Cannot use %s as array end value", expr), ast.Location)
		} else {
			nodeType = expr
		}
	} else if ast.Label == "call" {
		nodeType = handleFunctionCall()
	} else if ast.Token.Kind == lexer.OPERATOR_ADD {
		lhs := evalType(&ast.Children[0], datatypes.None)
		rhs := evalType(&ast.Children[1], datatypes.None)
		if (lhs == datatypes.String || lhs == datatypes.Char) && (rhs == datatypes.String || rhs == datatypes.Char) {
			nodeType = datatypes.String
		} else if slices.Contains(numericTypes, lhs) && slices.Contains(numericTypes, rhs) {
			if lhs == rhs {
				nodeType = lhs
			} else {
				nodeType, err = decideNumberType(lhs, rhs, ast.Token.Value)
				if err != nil {
					messages = messages.Complain(diagnostic.TypeError, err.Error(), ast.Location)
				}
			}
		} else {
			messages = messages.Complain(diagnostic.TypeError, fmt.Sprintf("Cannot use operator '%s' between %s and %s", ast.Token.Value, lhs, rhs), ast.Location)
		}
	} else if ast.Token.Kind == lexer.OPERATOR_MULT {
		lhs := evalType(&ast.Children[0], datatypes.None)
		rhs := evalType(&ast.Children[1], datatypes.None)
		fmt.Println(lhs, rhs)
		if !slices.Contains(numericTypes, lhs) || !slices.Contains(numericTypes, rhs) {
			messages = messages.Complain(diagnostic.TypeError, fmt.Sprintf("Cannot use operator '%s' between %s and %s", ast.Token.Value, lhs, rhs), ast.Location)
		} else {
			nodeType, err = decideNumberType(lhs, rhs, ast.Token.Value)
			if err != nil {
				messages = messages.Complain(diagnostic.TypeError, err.Error(), ast.Location)
			}
		}
	} else if ast.Token.Kind == lexer.OPERATOR_BS || ast.Token.Kind == lexer.OPERATOR_BW {
		lhs := evalType(&ast.Children[0], datatypes.None)
		rhs := evalType(&ast.Children[1], datatypes.None)
		if (lhs == datatypes.Uint32 || lhs == datatypes.Uint64) && (rhs == datatypes.Uint32 || rhs == datatypes.Uint64) ||
			(lhs == datatypes.Int32 || lhs == datatypes.Int64) && (rhs == datatypes.Int32 || rhs == datatypes.Int64) {
			nodeType, _ = decideNumberType(lhs, rhs, ast.Token.Value)
		} else {
			messages = messages.Complain(diagnostic.TypeError, fmt.Sprintf("Cannot use operator '%s' between %s and %s", ast.Token.Value, lhs, rhs), ast.Location)
		}
	} else if ast.Token.Kind == lexer.OPERATOR_COMPARE {
		lhs := evalType(&ast.Children[0], datatypes.None)
		rhs := evalType(&ast.Children[1], datatypes.None)
		if lhs != rhs && !comparableCheck(lhs, rhs) {
			messages = messages.Complain(diagnostic.TypeError, fmt.Sprintf("Invalid comparison between %s and %s", lhs, rhs), ast.Location)
		} else {
			if lhs == rhs && !lhs.IsPrimitive() && ast.Token.Value != "==" && ast.Token.Value != "!=" {
				str := globalScope.lookupStruct(lhs.String())
				if str == nil {
					messages = messages.Complain(diagnostic.NameError, fmt.Sprintf("Cannot find struct definition for %s", lhs.String()), ast.Location)
				} else {
					operator := ast.Token.Value
					compareBlock := str.innerScope.lookupNamedBlock("compare")
					if compareBlock == nil {
						messages = messages.Complain(diagnostic.TypeError, fmt.Sprintf("Cannot compare %s using operator '%s'. To support this comparison add a compare block with the appropriate functions", lhs, operator), ast.Location)
					} else {
						switch operator {
						case "<", "<=":
							if symbol := compareBlock.innerScope.lookupFunction("lessThan"); symbol == nil || symbol.returnType != datatypes.Bool {
								messages = messages.Complain(diagnostic.TypeError, fmt.Sprintf("Unsupported comparison. To support operators '<' and '<=', add function 'fn lessThan(%s)->bool' to compare block in %s definition", lhs, lhs), ast.Location)
							}
						case ">", ">=":
							if symbol := compareBlock.innerScope.lookupFunction("greaterThan"); symbol == nil || symbol.returnType != datatypes.Bool {
								messages = messages.Complain(diagnostic.TypeError, fmt.Sprintf("Unsupported comparison. To support operators '>' and '>=', add function 'fn greaterThan(%s)->bool' to compare block in %s definition", lhs, lhs), ast.Location)
							}
						}
					}
				}
			}
			nodeType = datatypes.Bool
		}
	} else if ast.Token.Value == "&&" || ast.Token.Value == "||" {
		lhs := evalType(&ast.Children[0], datatypes.None)
		rhs := evalType(&ast.Children[1], datatypes.None)
		if lhs != datatypes.Bool || rhs != datatypes.Bool {
			messages = messages.Complain(diagnostic.TypeError, fmt.Sprintf("Cannot use operator '%s' between %s and %s", ast.Token.Value, lhs, rhs), ast.Location)
		} else {
			nodeType = datatypes.Bool
		}
	}
	ast.Type = nodeType
	return nodeType
}

func comparableCheck(lhs datatypes.DataType, rhs datatypes.DataType) bool {
	lhsUnsigned := slices.Contains(unsignedTypes, lhs)
	rhsUnsigned := slices.Contains(unsignedTypes, rhs)
	return lhsUnsigned == rhsUnsigned
}

func decideNumberType(lhs datatypes.DataType, rhs datatypes.DataType, operator string) (datatypes.DataType, error) {
	if (lhs == datatypes.Uint32 || lhs == datatypes.Uint64) && (rhs == datatypes.Uint32 || rhs == datatypes.Uint64) {
		return datatypes.Uint64, nil
	} else if (lhs == datatypes.Int32 || lhs == datatypes.Int64) && (rhs == datatypes.Int32 || rhs == datatypes.Int64) {
		return datatypes.Int64, nil
	} else if (lhs == datatypes.Int32 || lhs == datatypes.Float) && (rhs == datatypes.Int32 || rhs == datatypes.Float) {
		return datatypes.Float, nil
	} else if (lhs == datatypes.Int32 || lhs == datatypes.Double) && (rhs == datatypes.Int32 || rhs == datatypes.Double) {
		return datatypes.Double, nil
	} else if (lhs == datatypes.Int64 || lhs == datatypes.Double) && (rhs == datatypes.Int64 || rhs == datatypes.Double) {
		return datatypes.Double, nil
	} else if (lhs == datatypes.Float || lhs == datatypes.Double) && (rhs == datatypes.Float || rhs == datatypes.Double) {
		return datatypes.Double, nil
	} else {
		return datatypes.None, fmt.Errorf("Cannot use operator '%s' between %s and %s", operator, lhs, rhs)
	}
}

func handleFunctionCall() datatypes.DataType {
	return datatypes.None
}

func nodeToType(node parser.AST) datatypes.DataType {
	if node.Token.Kind == lexer.ID {
		symbol := globalScope.lookupType(node.Token.Value)
		if symbol == nil || (symbol.getSymbolType() != "interface" && symbol.getSymbolType() != "struct") {
			messages = messages.Complain(diagnostic.TypeError, fmt.Sprintf("Invalid type '%s' provided", node.Token.Value), node.Location)
			return datatypes.None
		}
		return datatypes.DynamicType(node.Token.Value)
	}
	switch node.Token.Value {
	case "int":
		return datatypes.Int32
	case "int64":
		return datatypes.Int64
	case "uint32":
		return datatypes.Uint32
	case "uint64":
		return datatypes.Uint64
	case "float":
		return datatypes.Float
	case "double":
		return datatypes.Double
	case "char":
		return datatypes.Char
	case "bool":
		return datatypes.Bool
	case "String":
		return datatypes.String
	default:
		return datatypes.None
	}
}
