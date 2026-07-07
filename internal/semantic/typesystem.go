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
		messages.Complain(diagnostic.NameError, ast.Location, "Struct %s not defined", name)
		return datatypes.None
	}
	if len(ast.Children) == 1 {
		return datatypes.DynamicType(name)
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
			messages.Complain(diagnostic.NameError, prop.Location, "Could not find property '%s' in struct %s", propId, symbol.name)
			continue
		}
		if _, ok := visited[propId]; ok {
			messages.Complain(diagnostic.IllegalStatementError, prop.Location, "Cannot define struct property '%s' multiple times in struct literal", propId)
			continue
		}
		value := prop.Children[1]
		if valueType := evalType(&value, property.Type); property.Type != valueType {
			messages.Complain(diagnostic.TypeError, value.Location, "Property type %s expected but found %s", property.Type, valueType)
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
			messages.Complain(diagnostic.NameError, ast.Location, "Variable '%s' not defined in this scope", ast.Token.Value)
		}
	} else if ast.Label == "Unary" {
		if leftTok := ast.Children[0].Token; leftTok.Kind == lexer.OPERATOR_UNARY || leftTok.Value == "-" {
			rhs := evalType(&ast.Children[1], datatypes.None)
			switch leftTok.Value {
			case "!":
				if rhs != datatypes.Bool {
					messages.Complain(diagnostic.TypeError, ast.Location, "Must use bool value with unary '!'")
				} else {
					nodeType = rhs
				}
			case "-":
				if !slices.Contains(numericTypes, rhs) {
					messages.Complain(diagnostic.TypeError, ast.Location, "Cannot use unary '-' with type %s", rhs)
				} else {
					nodeType = rhs
				}
			case "~":
				if !slices.Contains(unsignedTypes, rhs) && !slices.Contains(signedIntTypes, rhs) {
					messages.Complain(diagnostic.TypeError, ast.Location, "Cannot negate value of type %s", rhs)
				} else {
					nodeType = rhs
				}
			default: // ++, --
				operand := ast.Children[1].Token
				symbol := currentScope.lookupVariable(operand.Value)
				if symbol != nil {
					if !slices.Contains(numericTypes, symbol.Type) {
						messages.Complain(diagnostic.TypeError, operand.Location, "Cannot use '%s' with type %s", ast.Children[0].Token.Value, symbol.Type)
					} else {
						nodeType = symbol.Type
					}
				} else {
					messages.Complain(diagnostic.NameError, operand.Location, "Could not find variable with name %s", operand.Value)
				}
			}
		} else {
			operand := ast.Children[0].Token
			symbol := currentScope.lookupVariable(operand.Value)
			if symbol != nil {
				if !slices.Contains(numericTypes, symbol.Type) {
					messages.Complain(diagnostic.TypeError, operand.Location, "Cannot use '%s' with type %s", ast.Children[1].Token.Value, symbol.Type)
				} else {
					nodeType = symbol.Type
				}
			} else {
				messages.Complain(diagnostic.NameError, operand.Location, "Could not find variable with name %s", operand.Value)
			}
		}

	} else if ast.Label == "typecast" {
		lhs := evalType(&ast.Children[0], datatypes.None)
		target := nodeToType(ast.Children[1])
		if lhs != target && target != datatypes.String {
			if lhs == datatypes.Uint64 {
				if target == datatypes.Uint32 || target == datatypes.Float || target == datatypes.Int32 {
					messages.Warn(ast.Location, "Lossy conversion from %s to %s", lhs, target)
				} else if !slices.Contains(numericTypes, target) {
					messages.Complain(diagnostic.CastError, ast.Location, "Typecasting from %s to %s not allowed", lhs, target)
				}
			} else if lhs == datatypes.Int64 {
				if target == datatypes.Uint32 || target == datatypes.Float || target == datatypes.Int32 {
					messages.Warn(ast.Location, "Lossy conversion from %s to %s", lhs, target)
				} else if !slices.Contains(numericTypes, target) {
					messages.Complain(diagnostic.CastError, ast.Location, "Typecasting from %s to %s not allowed", lhs, target)
				}
			} else if lhs == datatypes.String {
				messages.Complain(diagnostic.CastError, ast.Location, "Typecasting from %s to %s not allowed", lhs, target)
			} else if lhs == datatypes.Double {
				if target == datatypes.Uint32 || target == datatypes.Float || target == datatypes.Int32 {
					messages.Warn(ast.Location, "Lossy conversion from %s to %s", lhs, target)
				} else if !slices.Contains(numericTypes, target) {
					messages.Complain(diagnostic.CastError, ast.Location, "Typecasting from %s to %s not allowed", lhs, target)
				}
			} else if !lhs.IsPrimitive() {
				// look for cast function
				str := globalScope.lookupStruct(lhs.String())
				if str == nil {
					messages.Complain(diagnostic.CastError, ast.Location, "Cannot typecast %s to %s. Could not find definition of '%s'", lhs, target, lhs)
				} else {
					castBlock := str.innerScope.lookupNamedBlock("cast")
					if castBlock == nil {
						messages.Complain(diagnostic.CastError, ast.Location, "Cannot typecast %s to %s. To support typecasting add a cast block with a function returning the target type", lhs, target)
					} else if !castBlock.HasReturnType(target) {
						messages.Complain(diagnostic.CastError, ast.Location, "Cannot typecast %s to %s. To support this typecasting add a function returning target type to cast block", lhs, target)
					}
				}
			} else if slices.Contains(numericTypes, lhs) && slices.Contains(numericTypes, target) {
				return target
			} else {
				messages.Complain(diagnostic.CastError, ast.Location, "Typecasting from %s to %s not allowed", lhs, target)
			}
		}
		return target

	} else if ast.Label == "index" {
		// only supports one index until arrays added
		if len(ast.Children) > 2 {
			messages.Warn(ast.Location, "Multiple indexes not yet supported")
		}
		lhs := evalType(&ast.Children[0], datatypes.None)
		rhs := evalType(&ast.Children[1], datatypes.None)
		if ast.Children[1].Label == "slice" {
			nodeType = datatypes.String
		} else {
			if lhs != datatypes.String && (!slices.Contains(signedIntTypes, rhs) && !slices.Contains(unsignedTypes, rhs)) {
				messages.Complain(diagnostic.TypeError, ast.Location, "Cannot index String with %s type", rhs)
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
				messages.Complain(diagnostic.TypeError, expr.Location, "Invalid type %s used in range expression", operand)
			} else {
				nodeType = operand
			}
		case 3:
			lhs := evalType(&ast.Children[0], datatypes.None)
			rhs := evalType(&ast.Children[2], datatypes.None)
			if (!slices.Contains(signedIntTypes, lhs) && !slices.Contains(unsignedTypes, lhs)) || (!slices.Contains(signedIntTypes, rhs) && !slices.Contains(unsignedTypes, rhs)) {
				messages.Complain(diagnostic.TypeError, ast.Location, "Both sides of slice expression must be an int type; got %s and %s", lhs, rhs)
			} else {
				nodeType, err = decideNumberType(lhs, rhs, ast.Children[1].Token.Value)
				if err != nil {
					messages.Complain(diagnostic.TypeError, ast.Location, "%s", err.Error())
				}
			}
		}

	} else if ast.Label == "ARR-END" {
		expr := evalType(&ast.Children[0], datatypes.None)
		if !slices.Contains(unsignedTypes, expr) && !slices.Contains(signedIntTypes, expr) {
			messages.Complain(diagnostic.TypeError, ast.Location, "Cannot use %s as array end value", expr)
		} else {
			nodeType = expr
		}
	} else if ast.Label == "call" {
		nodeType = handleFunctionCall(ast.Children)
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
					messages.Complain(diagnostic.TypeError, ast.Location, "%s", err.Error())
				}
			}
		} else {
			messages.Complain(diagnostic.TypeError, ast.Location, "Cannot use operator '%s' between %s and %s", ast.Token.Value, lhs, rhs)
		}
	} else if ast.Token.Kind == lexer.OPERATOR_MULT {
		lhs := evalType(&ast.Children[0], datatypes.None)
		rhs := evalType(&ast.Children[1], datatypes.None)
		if !slices.Contains(numericTypes, lhs) || !slices.Contains(numericTypes, rhs) {
			messages.Complain(diagnostic.TypeError, ast.Location, "Cannot use operator '%s' between %s and %s", ast.Token.Value, lhs, rhs)
		} else {
			nodeType, err = decideNumberType(lhs, rhs, ast.Token.Value)
			if err != nil {
				messages.Complain(diagnostic.TypeError, ast.Location, "%s", err.Error())
			}
		}
	} else if ast.Token.Kind == lexer.OPERATOR_BS || ast.Token.Kind == lexer.OPERATOR_BW {
		lhs := evalType(&ast.Children[0], datatypes.None)
		rhs := evalType(&ast.Children[1], datatypes.None)
		if (lhs == datatypes.Uint32 || lhs == datatypes.Uint64) && (rhs == datatypes.Uint32 || rhs == datatypes.Uint64) ||
			(lhs == datatypes.Int32 || lhs == datatypes.Int64) && (rhs == datatypes.Int32 || rhs == datatypes.Int64) {
			nodeType, _ = decideNumberType(lhs, rhs, ast.Token.Value)
		} else {
			messages.Complain(diagnostic.TypeError, ast.Location, "Cannot use operator '%s' between %s and %s", ast.Token.Value, lhs, rhs)
		}
	} else if ast.Token.Kind == lexer.OPERATOR_COMPARE {
		lhs := evalType(&ast.Children[0], datatypes.None)
		rhs := evalType(&ast.Children[1], datatypes.None)
		if !comparableCheck(lhs, rhs) {
			messages.Complain(diagnostic.TypeError, ast.Location, "Invalid comparison between %s and %s", lhs, rhs)
		} else {
			if lhs == rhs && !lhs.IsPrimitive() && ast.Token.Value != "==" && ast.Token.Value != "!=" {
				str := globalScope.lookupStruct(lhs.String())
				if str == nil {
					messages.Complain(diagnostic.NameError, ast.Location, "Cannot find struct definition for %s", lhs.String())
				} else {
					operator := ast.Token.Value
					compareBlock := str.innerScope.lookupNamedBlock("compare")
					if compareBlock == nil {
						messages.Complain(diagnostic.TypeError, ast.Location, "Cannot compare %s using operator '%s'. To support this comparison add a compare block with the appropriate functions", lhs, operator)
					} else {
						switch operator {
						case "<", "<=":
							if symbol := compareBlock.innerScope.lookupFunction("lessThan"); symbol == nil || symbol.returnType != datatypes.Bool {
								messages.Complain(diagnostic.TypeError, ast.Location, "Unsupported comparison. To support operators '<' and '<=', add function 'fn lessThan(%s)->bool' to compare block in %s definition", lhs, lhs)
							}
						case ">", ">=":
							if symbol := compareBlock.innerScope.lookupFunction("greaterThan"); symbol == nil || symbol.returnType != datatypes.Bool {
								messages.Complain(diagnostic.TypeError, ast.Location, "Unsupported comparison. To support operators '>' and '>=', add function 'fn greaterThan(%s)->bool' to compare block in %s definition", lhs, lhs)
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
			messages.Complain(diagnostic.TypeError, ast.Location, "Cannot use operator '%s' between %s and %s", ast.Token.Value, lhs, rhs)
		} else {
			nodeType = datatypes.Bool
		}
	} else if ast.Token.Value == "**" {
		base := evalType(&ast.Children[0], datatypes.None)
		expo := evalType(&ast.Children[1], datatypes.None)
		if !slices.Contains(numericTypes, base) || !slices.Contains(numericTypes, expo) {
			messages.Complain(diagnostic.TypeError, ast.Location, "Cannot use exponent with types %s and %s", base, expo)
		} else {
			// determine type
			if slices.Contains(floatTypes, expo) {
				nodeType = expo
			} else {
				nodeType = base
			}
		}
	} else if ast.Label == "dot" {
		nodeType = handleDot(ast.Children[0], ast.Children[1], false)
	}
	ast.Type = nodeType
	return nodeType
}

func comparableCheck(lhs datatypes.DataType, rhs datatypes.DataType) bool {
	lhsUnsigned := slices.Contains(unsignedTypes, lhs)
	rhsUnsigned := slices.Contains(unsignedTypes, rhs)
	lhsSigned := slices.Contains(signedIntTypes, lhs) || slices.Contains(floatTypes, lhs)
	rhsSigned := slices.Contains(signedIntTypes, rhs) || slices.Contains(floatTypes, rhs)

	return lhs == rhs || (lhsUnsigned && rhsUnsigned) || (lhsSigned && rhsSigned)
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

func handleFunctionCall(details []parser.AST) datatypes.DataType {
	// TODO: handle calls to functions without bodies
	scope := currentScope
	var name lexer.Token
	if details[0].Label == "dot" {
		lhs := handleDot(details[0].Children[0], details[0].Children[1], true)
		symbol := globalScope.lookupType(lhs.String())
		if symbol == nil {
			messages.Complain(diagnostic.NameError, details[0].Children[1].Location, "Could not find type %s", lhs)
			return datatypes.None
		}
		name = details[0].Children[1].Token
		scope = symbol.getInnerScope()
		// TODO: handle named block functions
		if symbol.getSymbolType() == "struct" {

		}
	} else {
		name = details[0].Token
	}
	symbol := scope.lookupFunction(name.Value)
	if symbol == nil {
		messages.Complain(diagnostic.NameError, name.Location, "Could not find function %s in scope", name.Value)
		return datatypes.None
	}
	// check parameters
	var params []datatypes.DataType = []datatypes.DataType{}
	if len(details) == 2 {
		for _, param := range details[1].Children {
			params = append(params, evalType(&param, datatypes.None))
		}
	}
	paramList := datatypes.Join(params)
	if fn, ok := symbol.overloads[paramList]; ok {
		if fn.isPrivate && !currentScope.HasParentScope(scope) {
			messages.Complain(diagnostic.AccessError, details[1].Location, "Cannot access private function '%s' from outside struct definition", name.Value)
		} else {
			return symbol.returnType
		}
	} else {
		// TODO: find closest error
		messages.Complain(diagnostic.CallError, details[1].Location, "Could not find function '%s(%s)->%s'", name.Value, paramList, symbol.returnType)
	}
	return datatypes.None
}

func handleDot(left parser.AST, right parser.AST, isFnCall bool) datatypes.DataType {
	var lhs datatypes.DataType = datatypes.None
	if left.Token.Value != "dot" {
		lhs = evalType(&left, datatypes.None)
	} else {
		lhs = handleDot(left.Children[0], left.Children[1], isFnCall)
	}
	if isFnCall {
		return lhs
	}
	rname := right.Token.Value
	if lhs == datatypes.String && rname == "length" {
		return datatypes.Int32
	} else if !lhs.IsPrimitive() {
		symbol := globalScope.lookupType(lhs.String())
		if symbol == nil {
			messages.Complain(diagnostic.NameError, left.Location, "Could not find type %s", lhs)
		} else {
			var scope *Scope
			if privateBlock := symbol.getNamedBlockIfExists("private"); privateBlock != nil {
				scope = privateBlock.innerScope
			} else {
				scope = symbol.getInnerScope()
			}
			if prop := scope.lookupVariable(rname); prop != nil {
				if prop.isPrivate && !currentScope.HasParentScope(symbol.getInnerScope()) {
					messages.Complain(diagnostic.AccessError, right.Location, "Cannot access private property from outside struct definition")
				} else {
					return prop.Type
				}
			} else {
				messages.Complain(diagnostic.NameError, right.Location, "Could not find property %s in type %s", rname, lhs)
			}
		}
	} else {
		messages.Complain(diagnostic.TypeError, right.Location, "Cannot access property %s of type %s", rname, lhs)
	}
	return datatypes.None
}

func nodeToType(node parser.AST) datatypes.DataType {
	if node.Token.Kind == lexer.ID {
		symbol := globalScope.lookupType(node.Token.Value)
		if symbol == nil || (symbol.getSymbolType() != "interface" && symbol.getSymbolType() != "struct") {
			messages.Complain(diagnostic.TypeError, node.Location, "Invalid type '%s' provided", node.Token.Value)
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
