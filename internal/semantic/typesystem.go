package semantic

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	ds "github.com/EladB1/The/internal/datastructures"
	"github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/diagnostic"
	"github.com/EladB1/The/internal/lexer"
	"github.com/EladB1/The/internal/parser"
)

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
		if slices.Contains(datatypes.NumericTypes, expectedType) {
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
		innerScope := symbol.innerScope
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
		if valueType, hasErr := evalType(&value, property.Type); property.Type != valueType && !hasErr {
			messages.Complain(diagnostic.TypeError, value.Location, "Property type %s expected but found %s", property.Type, valueType)
		}
		visited.Append(propId)
	}
	return datatypes.DynamicType(name)
}

func evalType(ast *parser.AST, expectedType datatypes.DataType) (datatypes.DataType, bool) {
	hasError := false
	hasErr := false
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
			hasError = true
		}
	} else if ast.Label == "Unary" {
		if leftTok := ast.Children[0].Token; leftTok.Kind == lexer.OPERATOR_UNARY || leftTok.Value == "-" {
			rhs, hasErr := evalType(&ast.Children[1], expectedType)
			if hasErr {
				hasError = hasErr
			}
			switch leftTok.Value {
			case "!":
				if rhs != datatypes.Bool {
					messages.Complain(diagnostic.TypeError, ast.Location, "Must use bool value with unary '!'")
					hasError = true
				} else {
					nodeType = datatypes.Bool
				}
			case "-":
				if !slices.Contains(datatypes.NumericTypes, rhs) {
					messages.Complain(diagnostic.TypeError, ast.Location, "Cannot use unary '-' with type %s", rhs)
					hasError = true
				} else {
					nodeType = rhs
				}
			case "~":
				if !slices.Contains(datatypes.IntTypes, rhs) {
					messages.Complain(diagnostic.TypeError, ast.Location, "Cannot negate value of type %s", rhs)
					hasError = true
				} else {
					nodeType = rhs
				}
			default: // ++, --
				operand := ast.Children[1].Token
				symbol := currentScope.lookupVariable(operand.Value)
				if symbol != nil {
					if !slices.Contains(datatypes.NumericTypes, symbol.Type) {
						messages.Complain(diagnostic.TypeError, operand.Location, "Cannot use '%s' with type %s", ast.Children[0].Token.Value, symbol.Type)
						hasError = true
					} else {
						nodeType = symbol.Type
					}
				} else {
					messages.Complain(diagnostic.NameError, operand.Location, "Could not find variable with name %s", operand.Value)
					hasError = true
				}
			}
		} else {
			operand := ast.Children[0].Token
			symbol := currentScope.lookupVariable(operand.Value)
			if symbol != nil {
				if !slices.Contains(datatypes.NumericTypes, symbol.Type) {
					messages.Complain(diagnostic.TypeError, operand.Location, "Cannot use '%s' with type %s", ast.Children[1].Token.Value, symbol.Type)
					hasError = true
				} else {
					nodeType = symbol.Type
				}
			} else {
				messages.Complain(diagnostic.NameError, operand.Location, "Could not find variable with name %s", operand.Value)
				hasError = true
			}
		}

	} else if ast.Label == "typecast" {
		lhs, hasErr := evalType(&ast.Children[0], datatypes.None)
		hasError = hasError || hasErr
		target := nodeToType(ast.Children[1])
		if lhs != target && target != datatypes.String {
			if lhs == datatypes.Uint64 {
				if target == datatypes.Uint32 || target == datatypes.Float || target == datatypes.Int32 {
					messages.Warn(ast.Location, "Lossy conversion from %s to %s", lhs, target)
				} else if !slices.Contains(datatypes.NumericTypes, target) {
					messages.Complain(diagnostic.CastError, ast.Location, "Typecasting from %s to %s not allowed", lhs, target)
					hasError = true
				}
			} else if lhs == datatypes.Int64 {
				if target == datatypes.Uint32 || target == datatypes.Float || target == datatypes.Int32 {
					messages.Warn(ast.Location, "Lossy conversion from %s to %s", lhs, target)
				} else if !slices.Contains(datatypes.NumericTypes, target) {
					messages.Complain(diagnostic.CastError, ast.Location, "Typecasting from %s to %s not allowed", lhs, target)
					hasError = true
				}
			} else if lhs == datatypes.String {
				messages.Complain(diagnostic.CastError, ast.Location, "Typecasting from %s to %s not allowed", lhs, target)
			} else if lhs == datatypes.Double {
				if target == datatypes.Uint32 || target == datatypes.Float || target == datatypes.Int32 {
					messages.Warn(ast.Location, "Lossy conversion from %s to %s", lhs, target)
				} else if !slices.Contains(datatypes.NumericTypes, target) {
					messages.Complain(diagnostic.CastError, ast.Location, "Typecasting from %s to %s not allowed", lhs, target)
					hasError = true
				}
			} else if !lhs.IsPrimitive() {
				// look for cast function
				str := globalScope.lookupStruct(lhs.String())
				if str == nil {
					messages.Complain(diagnostic.CastError, ast.Location, "Cannot typecast %s to %s. Could not find definition of '%s'", lhs, target, lhs)
					hasError = true
				} else {
					castBlock := str.innerScope.lookupNamedBlock("cast")
					if castBlock == nil {
						messages.Complain(diagnostic.CastError, ast.Location, "Cannot typecast %s to %s. To support typecasting add a cast block with a function returning the target type", lhs, target)
						hasError = true
					} else if !castBlock.HasReturnType(target) {
						messages.Complain(diagnostic.CastError, ast.Location, "Cannot typecast %s to %s. To support this typecasting add a function returning target type to cast block", lhs, target)
						hasError = true
					}
				}
			} else if slices.Contains(datatypes.NumericTypes, lhs) && slices.Contains(datatypes.NumericTypes, target) {
				return target, hasError
			} else {
				messages.Complain(diagnostic.CastError, ast.Location, "Typecasting from %s to %s not allowed", lhs, target)
				hasError = true
			}
		}
		return target, hasError

	} else if ast.Label == "index" {
		lhs, lHasErr := evalType(&ast.Children[0], datatypes.None)
		rhs, rHasErr := evalType(&ast.Children[1], datatypes.None)
		if lHasErr || rHasErr {
			hasError = true
		}
		if lhs != datatypes.String {
			messages.Complain(diagnostic.TypeError, ast.Children[0].Location, "Cannot index type %s", lhs.String())
			hasError = true
		}
		if ast.Children[1].Label == "slice" {
			nodeType = datatypes.String
		} else {
			if lhs != datatypes.String && !slices.Contains(datatypes.IntTypes, rhs) {
				messages.Complain(diagnostic.TypeError, ast.Location, "Cannot index String with %s type", rhs)
				hasError = true
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
			operand, hasErr := evalType(&expr, datatypes.None)
			if hasErr {
				hasError = hasErr
			}
			if !slices.Contains(datatypes.IntTypes, operand) {
				messages.Complain(diagnostic.TypeError, expr.Location, "Invalid type %s used in range expression", operand)
				hasError = true
			} else {
				nodeType = operand
			}
		case 3:
			lhs, lHasErr := evalType(&ast.Children[0], datatypes.None)
			rhs, rHasErr := evalType(&ast.Children[2], datatypes.None)
			if lHasErr || rHasErr {
				hasError = true
			} else {
				if !slices.Contains(datatypes.IntTypes, lhs) || !slices.Contains(datatypes.IntTypes, rhs) {
					messages.Complain(diagnostic.TypeError, ast.Location, "Both sides of slice expression must be an int type; got %s and %s", lhs, rhs)
					hasError = true
				} else {
					nodeType, err = decideNumberType(lhs, rhs, ast.Children[1].Token.Value)
					if err != nil {
						messages.Complain(diagnostic.TypeError, ast.Location, "%s", err.Error())
						hasError = true
					}
				}
			}
		}

	} else if ast.Label == "ARR-END" {
		expr, hasErr := evalType(&ast.Children[0], datatypes.None)
		hasError = hasError || hasErr
		if !hasErr && !slices.Contains(datatypes.IntTypes, expr) {
			messages.Complain(diagnostic.TypeError, ast.Location, "Cannot use %s as array end value", expr)
			hasError = true
		} else {
			nodeType = expr
		}
	} else if ast.Label == "call" {
		hasErr := false
		nodeType, hasErr = handleFunctionCall(ast.Children)
		hasError = hasError || hasErr
	} else if ast.Token.Kind == lexer.OPERATOR_ADD {
		nodeType, hasErr = evalAdd(&ast.Children[0], &ast.Children[1], ast.Token, expectedType)
		hasError = hasError || hasErr
	} else if ast.Token.Kind == lexer.OPERATOR_MULT {
		nodeType, hasErr = evalMult(&ast.Children[0], &ast.Children[1], ast.Token, expectedType)
		hasError = hasError || hasErr
	} else if ast.Token.Kind == lexer.OPERATOR_BS || ast.Token.Kind == lexer.OPERATOR_BW {
		nodeType, hasErr = evalBitOperation(&ast.Children[0], &ast.Children[1], ast.Token, expectedType)
		hasError = hasError || hasErr
	} else if ast.Token.Kind == lexer.OPERATOR_COMPARE {
		nodeType, hasErr = evalCompare(&ast.Children[0], &ast.Children[1], ast.Token, expectedType)
		hasError = hasError || hasErr
	} else if ast.Token.Value == "&&" || ast.Token.Value == "||" {
		lhs, lHasErr := evalType(&ast.Children[0], expectedType)
		rhs, rHasErr := evalType(&ast.Children[1], expectedType)
		if lHasErr || rHasErr {
			return datatypes.None, true
		}
		if lhs != datatypes.Bool || rhs != datatypes.Bool {
			messages.Complain(diagnostic.TypeError, ast.Location, "Cannot use operator '%s' between %s and %s", ast.Token.Value, lhs, rhs)
			hasError = true
		} else {
			nodeType = datatypes.Bool
		}
	} else if ast.Token.Value == "**" {
		base, baseHasErr := evalType(&ast.Children[0], expectedType)
		expo, expoHasErr := evalType(&ast.Children[1], expectedType)
		if baseHasErr || expoHasErr {
			return datatypes.None, true
		}
		if !slices.Contains(datatypes.NumericTypes, base) || !slices.Contains(datatypes.NumericTypes, expo) {
			messages.Complain(diagnostic.TypeError, ast.Location, "Cannot use exponent with types %s and %s", base, expo)
			hasError = true
		} else {
			// determine type
			if slices.Contains(datatypes.FloatTypes, expo) {
				nodeType = expo
			} else {
				nodeType = base
			}
		}
	} else if ast.Label == "dot" {
		hasErr := false
		nodeType, hasErr = handleDot(ast.Children[0], ast.Children[1], false)
		if hasErr {
			hasError = hasErr
		}
	}
	ast.Type = nodeType
	return nodeType, hasError
}

func evalAdd(left *parser.AST, right *parser.AST, operator lexer.Token, expectedType datatypes.DataType) (datatypes.DataType, bool) {
	hasError := false
	lhs, lHasErr := evalType(left, expectedType)
	rhs, rHasErr := evalType(right, expectedType)
	if lHasErr || rHasErr {
		return datatypes.None, true
	}
	if (lhs == datatypes.String || lhs == datatypes.Char) && (rhs == datatypes.String || rhs == datatypes.Char) {
		return datatypes.String, hasError
	} else if slices.Contains(datatypes.NumericTypes, lhs) && slices.Contains(datatypes.NumericTypes, rhs) {
		if lhs == rhs {
			return lhs, hasError
		} else {
			nodeType, err := decideNumberType(lhs, rhs, operator.Value)
			if err != nil {
				messages.Complain(diagnostic.TypeError, operator.Location, "%s", err.Error())
				hasError = true
			} else {
				return nodeType, hasError
			}
		}
	} else {
		messages.Complain(diagnostic.TypeError, operator.Location, "Cannot use operator '%s' between %s and %s", operator.Value, lhs, rhs)
		hasError = true
	}
	return datatypes.None, hasError
}

func evalMult(left *parser.AST, right *parser.AST, operator lexer.Token, expectedType datatypes.DataType) (datatypes.DataType, bool) {
	hasError := false
	lhs, lHasErr := evalType(left, expectedType)
	rhs, rHasErr := evalType(right, expectedType)
	if lHasErr || rHasErr {
		return datatypes.None, true
	}
	if !slices.Contains(datatypes.NumericTypes, lhs) || !slices.Contains(datatypes.NumericTypes, rhs) {
		messages.Complain(diagnostic.TypeError, operator.Location, "Cannot use operator '%s' between %s and %s", operator.Value, lhs, rhs)
		hasError = true
	} else {
		nodeType, err := decideNumberType(lhs, rhs, operator.Value)
		if err != nil {
			messages.Complain(diagnostic.TypeError, operator.Location, "%s", err.Error())
			hasError = true
		} else {
			return nodeType, hasError
		}
	}
	return datatypes.None, hasError
}

func evalBitOperation(left *parser.AST, right *parser.AST, operator lexer.Token, expectedType datatypes.DataType) (datatypes.DataType, bool) {
	hasError := false
	lhs, lHasErr := evalType(left, expectedType)
	rhs, rHasErr := evalType(right, expectedType)
	if lHasErr || rHasErr {
		return datatypes.None, true
	}
	if (lhs == datatypes.Uint32 || lhs == datatypes.Uint64) && (rhs == datatypes.Uint32 || rhs == datatypes.Uint64) ||
		(lhs == datatypes.Int32 || lhs == datatypes.Int64) && (rhs == datatypes.Int32 || rhs == datatypes.Int64) {
		nodeType, _ := decideNumberType(lhs, rhs, operator.Value)
		return nodeType, hasError
	}
	messages.Complain(diagnostic.TypeError, operator.Location, "Cannot use operator '%s' between %s and %s", operator.Value, lhs, rhs)
	return datatypes.None, true
}

func evalCompare(left *parser.AST, right *parser.AST, operator lexer.Token, expectedType datatypes.DataType) (datatypes.DataType, bool) {
	hasError := false
	lhs, lHasErr := evalType(left, expectedType)
	rhs, rHasErr := evalType(right, expectedType)
	if lHasErr || rHasErr {
		return datatypes.None, true
	}
	if !comparableCheck(lhs, rhs) {
		messages.Complain(diagnostic.TypeError, operator.Location, "Invalid comparison between %s and %s", lhs, rhs)
		return datatypes.None, true
	}
	if lhs == rhs && !lhs.IsPrimitive() && operator.Value != "==" && operator.Value != "!=" {
		str := globalScope.lookupStruct(lhs.String())
		if str == nil {
			messages.Complain(diagnostic.NameError, operator.Location, "Cannot find struct definition for %s", lhs.String())
		} else {
			compareBlock := str.innerScope.lookupNamedBlock("compare")
			if compareBlock == nil {
				messages.Complain(diagnostic.TypeError, operator.Location, "Cannot compare %s using operator '%s'. To support this comparison add a compare block with the appropriate functions", lhs, operator.Value)
			} else {
				switch operator.Value {
				case "<", "<=":
					if symbol := compareBlock.innerScope.lookupFunctionByName("lessThan"); symbol == nil || symbol.returnType != datatypes.Bool {
						messages.Complain(diagnostic.TypeError, operator.Location, "Unsupported comparison. To support operators '<' and '<=', add function 'fn lessThan(%s)->bool' to compare block in %s definition", lhs, lhs)
						hasError = true
					}
				case ">", ">=":
					if symbol := compareBlock.innerScope.lookupFunctionByName("greaterThan"); symbol == nil || symbol.returnType != datatypes.Bool {
						messages.Complain(diagnostic.TypeError, operator.Location, "Unsupported comparison. To support operators '>' and '>=', add function 'fn greaterThan(%s)->bool' to compare block in %s definition", lhs, lhs)
						hasError = true
					}
				}
			}
		}
	}
	return datatypes.Bool, hasError
}

func comparableCheck(lhs datatypes.DataType, rhs datatypes.DataType) bool {
	lhsUnsigned := slices.Contains(datatypes.UnsignedTypes, lhs)
	rhsUnsigned := slices.Contains(datatypes.UnsignedTypes, rhs)
	lhsSigned := slices.Contains(datatypes.SignedIntTypes, lhs) || slices.Contains(datatypes.FloatTypes, lhs)
	rhsSigned := slices.Contains(datatypes.SignedIntTypes, rhs) || slices.Contains(datatypes.FloatTypes, rhs)

	return lhs == rhs || (lhsUnsigned && rhsUnsigned) || (lhsSigned && rhsSigned)
}

func decideNumberType(lhs datatypes.DataType, rhs datatypes.DataType, operator string) (datatypes.DataType, error) {
	if lhs == rhs {
		return lhs, nil
	}
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

func handleFunctionCall(details []parser.AST) (datatypes.DataType, bool) {
	// TODO: handle calls to functions without bodies
	hasError := false
	scope := currentScope
	var name lexer.Token
	if details[0].Label == "dot" {
		lhs, hasErr := handleDot(details[0].Children[0], details[0].Children[1], true)
		if hasErr {
			hasError = hasErr
		}
		if lhs.IsScopeRef() {
			scopes := lhs.GetScopes()
			if len(scopes) < 2 {
				messages.Complain(diagnostic.ReferenceError, details[0].Location, "Could not find reference value")
				hasError = true
			} else {
				str := globalScope.lookupStruct(scopes[0])
				if str == nil {
					messages.Complain(diagnostic.NameError, details[0].Location, "Could not find struct %s", scopes[0])
					hasError = true
				}
				nb := str.innerScope.lookupNamedBlock(scopes[1])
				if nb == nil {
					messages.Complain(diagnostic.NameError, details[0].Location, "Could not find named block %s in struct %s", scopes[1], scopes[0])
					hasError = true
				}
				scope = nb.innerScope
				name = details[0].Children[1].Token
			}
		} else {
			symbol := globalScope.lookupType(lhs.String())
			if symbol == nil {
				messages.Complain(diagnostic.NameError, details[0].Children[1].Location, "Could not find type %s", lhs)
				return datatypes.None, true
			}
			name = details[0].Children[1].Token
			scope = symbol.getInnerScope()
			if symbol.getSymbolType() == "struct" {
				if conflicts := symbol.getConflicts(name.Value); len(conflicts) > 1 {
					sort.Strings(conflicts)
					messages.Complain(diagnostic.AmbiguityError, name.Location, "Interfaces %s both contain function named %s. Change the function call to pick which one to use", strings.Join(conflicts, ","), name.Value)
					return datatypes.None, true
				} else if len(conflicts) == 1 {
					if nb := scope.lookupNamedBlock(conflicts[0]); nb != nil {
						scope = nb.innerScope
					} else {
						messages.Complain(diagnostic.NameError, details[0].Location, "Could not find function %s", name.Value)
						hasError = true
					}
				}
			}
		}
	} else {
		name = details[0].Token
	}
	symbol := scope.lookupFunctionByName(name.Value)
	if symbol == nil {
		messages.Complain(diagnostic.NameError, name.Location, "Could not find function %s in scope", name.Value)
		return datatypes.None, true
	}
	// check parameters
	var params []datatypes.DataType = []datatypes.DataType{}
	if len(details) == 2 {
		for _, param := range details[1].Children {
			parameter, hasErr := evalType(&param, datatypes.None)
			params = append(params, parameter)
			if hasErr {
				hasError = hasErr
			}
		}
	}
	paramList := datatypes.Join(params)
	if fn := symbol.getMatchingOverload(params); fn != nil {
		if fn.isPrivate {
			messages.Complain(diagnostic.AccessError, details[0].Location, "Cannot access private function '%s' from outside struct definition", name.Value)
			hasError = true
		} else {
			return symbol.returnType, hasError
		}
	} else {
		// TODO: find closest error
		messages.Complain(diagnostic.CallError, details[1].Location, "Could not find function '%s(%s)->%s'", name.Value, paramList, symbol.returnType)
		hasError = true
	}
	return datatypes.None, hasError
}

func handleDot(left parser.AST, right parser.AST, isFnCall bool) (datatypes.DataType, bool) {
	hasError := true
	hasErr := false
	var lhs datatypes.DataType = datatypes.None
	if left.Token.Value != "dot" {
		lhs, hasErr = evalType(&left, datatypes.None)
		if hasErr {
			hasError = hasErr
		}
	} else {
		lhs, hasErr = handleDot(left.Children[0], left.Children[1], isFnCall)
	}
	if isFnCall {
		return lhs, hasError
	}
	rname := right.Token.Value
	if lhs == datatypes.String && rname == "length" {
		return datatypes.Int32, hasError
	} else if !lhs.IsPrimitive() {
		symbol := globalScope.lookupType(lhs.String())
		if symbol == nil {
			messages.Complain(diagnostic.NameError, left.Location, "Could not find type %s", lhs)
			hasError = true
		} else {
			scope := symbol.getInnerScope()
			if prop := scope.lookupVariable(rname); prop != nil {
				if prop.isPrivate && !currentScope.HasParentScope(symbol.getInnerScope()) {
					messages.Complain(diagnostic.AccessError, right.Location, "Cannot access private property from outside struct definition")
				} else {
					return prop.Type, hasError
				}
			} else {
				messages.Complain(diagnostic.NameError, right.Location, "Could not find property %s in type %s", rname, lhs)
				hasError = true
			}
		}
	} else {
		messages.Complain(diagnostic.TypeError, right.Location, "Cannot access property %s of type %s", rname, lhs)
		hasError = true
	}
	return datatypes.None, hasError
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
