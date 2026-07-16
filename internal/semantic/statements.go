package semantic

import (
	"slices"
	"sort"
	"strings"

	"github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/diagnostic"
	"github.com/EladB1/The/internal/lexer"
	"github.com/EladB1/The/internal/parser"
)

func analyzeForCondition(condition []parser.AST) {
	parts := len(condition)
	if condition[0].Label == "Variable" {
		symbol := analyzeVariable(condition[0])
		if symbol != nil {
			currentScope.variables[symbol.name] = *symbol
		}
		if condition[parts-2].Token.Value == "in" {
			if parts == 4 { // int i, char c in string
				symbol2 := analyzeVariable(condition[1])
				if symbol2 != nil {
					currentScope.variables[symbol2.name] = *symbol2
				}
				if symbol != nil && symbol2 != nil && !((slices.Contains(datatypes.IntTypes, symbol.Type) && symbol2.Type == datatypes.Char) || (symbol.Type == datatypes.Char && slices.Contains(datatypes.IntTypes, symbol2.Type))) {
					messages.Complain(diagnostic.TypeError, condition[2].Location, "Cannot use %s and %s as loop variables", symbol.Type.String(), symbol2.Type.String())
				}

			} else {
				if condition[parts-1].Label == "range" {
					if symbol != nil {
						expr, hasErr := evalType(&condition[parts-1], symbol.Type)
						if !hasErr && expr != symbol.Type {
							messages.Complain(diagnostic.TypeError, condition[parts-1].Location, "Variable of type %s not compatible with range expression of type %s", symbol.Type, expr)
						}
					}
				} else { // char c in string
					if symbol != nil && symbol.Type != datatypes.Char {
						messages.Complain(diagnostic.TypeError, condition[0].Location, "Cannot use %s as loop variables", symbol.Type.String())
					}
					rhs, hasErr := evalType(&condition[parts-1], datatypes.String)
					if !hasErr && rhs != datatypes.String {
						messages.Complain(diagnostic.TypeError, condition[parts-1].Location, "Cannot loop over %s", rhs.String())
					}
				}
			}
		} else {
			if symbol != nil {
				cond, hasErr := evalType(&condition[1], datatypes.Bool)
				if !hasErr && cond != datatypes.Bool {
					messages.Complain(diagnostic.TypeError, condition[1].Location, "Expected bool as loop condition but got %s", cond.String())
				}
				var expr datatypes.DataType
				if condition[2].Token.Kind == lexer.OPERATOR_ASSIGN {
					expr, hasErr = analyzeAssignment(&condition[2])

				} else {
					expr, hasErr = evalType(&condition[2], symbol.Type)
				}
				if !hasErr && symbol.Type != expr {
					messages.Complain(diagnostic.TypeError, condition[2].Location, "Expected %s as loop expression but got %s", symbol.Type.String(), expr.String())
				}
			}
		}
	} else {
		iter, _ := analyzeAssignment(&condition[0])
		if iter != datatypes.None {
			cond, hasErr := evalType(&condition[1], datatypes.Bool)
			if !hasErr && cond != datatypes.Bool {
				messages.Complain(diagnostic.TypeError, condition[1].Location, "Expected bool as loop condition but got %s", cond.String())
			}
			var expr datatypes.DataType
			if condition[2].Token.Kind == lexer.OPERATOR_ASSIGN {
				expr, hasErr = analyzeAssignment(&condition[2])

			} else {
				expr, hasErr = evalType(&condition[2], iter)
			}
			if !hasErr && iter != expr {
				messages.Complain(diagnostic.TypeError, condition[2].Location, "Expected %s as loop expression but got %s", iter.String(), expr.String())
			}
		}

	}
}

func analyzeAssignment(stmt *parser.AST) (datatypes.DataType, bool) {
	hasError := false
	left := stmt.Children[0]
	var lhs datatypes.DataType
	var lHasErr bool
	if left.Label == "dot" {
		lhs, lHasErr = handleDot(left.Children[0], left.Children[1], false, true, false)
		if lHasErr {
			return datatypes.None, true
		}
	} else {
		symbol := currentScope.lookupVariable(left.Token.Value)
		if symbol == nil {
			messages.Complain(diagnostic.NameError, left.Location, "Variable '%s' is not defined in this scope", left.Token.Value)
			return datatypes.None, true
		}
		if !symbol.isMutable {
			messages.Complain(diagnostic.AccessError, left.Location, "Cannot change value of immutable variable '%s'", left.Token.Value)
			return datatypes.None, true
		}
		lhs = symbol.Type
	}
	rhs, hasErr := evalType(&stmt.Children[1], lhs)
	if hasErr {
		return datatypes.None, true
	}
	stmt.Children[0].Type = lhs
	operator := stmt.Token.Value
	switch operator {
	case "+=":
		if lhs == datatypes.String {
			if rhs != datatypes.Char && rhs != datatypes.String {
				messages.Complain(diagnostic.TypeError, stmt.Children[1].Location, "Cannot concatenate %s to String", rhs.String())
				hasError = true
			}
		} else if slices.Contains(datatypes.NumericTypes, lhs) {
			result, err := decideNumberType(lhs, rhs, operator)
			if err != nil {
				messages.Complain(diagnostic.TypeError, stmt.Location, "%s", err.Error())
				hasError = true
			} else if result != lhs {
				messages.Complain(diagnostic.TypeError, stmt.Children[1].Location, "Cannot assign %s to %s", rhs, lhs)
				hasError = true
			}
		} else {
			messages.Complain(diagnostic.TypeError, stmt.Children[1].Location, "Invalid operation between %s to %s", rhs.String(), lhs)
			hasError = true
		}
	case "-=", "*=", "/=":
		if slices.Contains(datatypes.NumericTypes, lhs) {
			result, err := decideNumberType(lhs, rhs, operator)
			if err != nil {
				messages.Complain(diagnostic.TypeError, stmt.Location, "%s", err.Error())
				hasError = true
			} else if result != lhs {
				messages.Complain(diagnostic.TypeError, stmt.Children[1].Location, "Cannot assign %s to %s", rhs, lhs)
				hasError = true
			}
		} else {
			messages.Complain(diagnostic.TypeError, stmt.Children[1].Location, "Invalid operation between %s to %s", rhs.String(), lhs)
			hasError = true
		}
	default:
		if lhs != rhs {
			if !ImplementsInterface(lhs, rhs) {
				messages.Complain(diagnostic.TypeError, stmt.Children[1].Location, "Cannot assign %s to %s", rhs, lhs)
				hasError = true
			}
		} else if intf := globalScope.lookupInterface(rhs.String()); intf != nil {
			messages.Complain(diagnostic.IllegalStatementError, stmt.Children[1].Location, "Cannot use interface type as value")
		}

	}
	stmt.Children[1].Type = rhs
	return lhs, hasError
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
			nodeType, err := handleBinaryNumberExpression(left, right, operator.Value, expectedType)
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
		nodeType, err := handleBinaryNumberExpression(left, right, operator.Value, expectedType)
		if err != nil {
			if slices.Contains(datatypes.UnsignedTypes, lhs) && slices.Contains(datatypes.SignedIntTypes, rhs) && left.Token.Kind == lexer.ID && right.IsLiteral() {
				nodeType = lhs
				return nodeType, hasError
			} else {
				messages.Complain(diagnostic.TypeError, operator.Location, "%s", err.Error())
				hasError = true
			}
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
		if slices.Contains(datatypes.NumericTypes, lhs) && slices.Contains(datatypes.NumericTypes, rhs) {
			_, err := handleBinaryNumberExpression(left, right, operator.Value, expectedType)
			if err == nil {
				return datatypes.Bool, hasError
			} else {
				messages.Complain(diagnostic.TypeError, operator.Location, "%s", err.Error())
				hasError = true
			}
		}
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
					if symbol := compareBlock.innerScope.lookupFunctionByName("lessThan"); symbol == nil || symbol.returnType != datatypes.Bool || !symbol.overloads[0].hasDefaultImplementation {
						messages.Complain(diagnostic.TypeError, operator.Location, "Unsupported comparison. To support operators '<' and '<=', add function 'fn lessThan(%s)->bool' to compare block in %s definition", lhs, lhs)
						hasError = true
					}
				case ">", ">=":
					if symbol := compareBlock.innerScope.lookupFunctionByName("greaterThan"); symbol == nil || symbol.returnType != datatypes.Bool || !symbol.overloads[0].hasDefaultImplementation {
						messages.Complain(diagnostic.TypeError, operator.Location, "Unsupported comparison. To support operators '>' and '>=', add function 'fn greaterThan(%s)->bool' to compare block in %s definition", lhs, lhs)
						hasError = true
					}
				}
			}
		}
	}
	return datatypes.Bool, hasError
}

func evalUnary(left *parser.AST, right *parser.AST, expectedType datatypes.DataType) (datatypes.DataType, bool) {
	hasError := false
	hasErr := false
	var nodeType datatypes.DataType = datatypes.None
	if leftTok := left.Token; leftTok.Kind == lexer.OPERATOR_UNARY || leftTok.Value == "-" { // left unary
		rhs, hasErr := evalType(right, expectedType)
		hasError = hasError || hasErr
		switch leftTok.Value {
		case "!":
			if rhs != datatypes.Bool {
				messages.Complain(diagnostic.TypeError, right.Location, "Must use bool value with unary '!'")
				hasError = true
			} else {
				nodeType = datatypes.Bool
			}
		case "-":
			if !slices.Contains(datatypes.NumericTypes, rhs) {
				messages.Complain(diagnostic.TypeError, right.Location, "Cannot use unary '-' with type %s", rhs)
				hasError = true
			} else {
				nodeType = rhs
			}
		case "~":
			if !slices.Contains(datatypes.IntTypes, rhs) {
				messages.Complain(diagnostic.TypeError, right.Location, "Cannot negate value of type %s", rhs)
				hasError = true
			} else {
				nodeType = rhs
			}
		default: // ++, --
			nodeType, hasErr = checkIncrementOperator(right.Token, left.Token)
			hasError = hasError || hasErr
		}
	} else { // right unary
		nodeType, hasErr = checkIncrementOperator(left.Token, right.Token)
		hasError = hasError || hasErr
	}
	return nodeType, hasError
}

func checkIncrementOperator(operand lexer.Token, operator lexer.Token) (datatypes.DataType, bool) {
	hasError := false
	var nodeType datatypes.DataType = datatypes.None
	symbol := currentScope.lookupVariable(operand.Value)
	if symbol != nil {
		if !slices.Contains(datatypes.NumericTypes, symbol.Type) {
			messages.Complain(diagnostic.TypeError, operand.Location, "Cannot use '%s' with type %s", operator.Value, symbol.Type)
			hasError = true
		} else if !symbol.isMutable {
			messages.Complain(diagnostic.AccessError, operand.Location, "Cannot change value of immutable variable '%s'", symbol.name)
			hasError = true
		} else {
			nodeType = symbol.Type
		}
	} else {
		messages.Complain(diagnostic.NameError, operand.Location, "Could not find variable with name %s", operand.Value)
		hasError = true
	}
	return nodeType, hasError
}

func evalTypecast(original *parser.AST, targetType *parser.AST, expectedType datatypes.DataType) (datatypes.DataType, bool) {
	lhs, hasErr := evalType(original, expectedType)
	hasError := hasErr
	target := nodeToType(*targetType)
	typeCastError := "Typecasting from %s to %s not allowed"
	if lhs != target && target != datatypes.String {
		if lhs == datatypes.Uint64 || lhs == datatypes.Int64 || lhs == datatypes.Double {
			if target == datatypes.Uint32 || target == datatypes.Float || target == datatypes.Int32 {
				messages.Warn(targetType.Location, "Lossy conversion from %s to %s", lhs, target)
			} else if !slices.Contains(datatypes.NumericTypes, target) {
				messages.Complain(diagnostic.CastError, targetType.Location, typeCastError, lhs, target)
				hasError = true
			}
		} else if lhs == datatypes.String {
			messages.Complain(diagnostic.CastError, targetType.Location, typeCastError, lhs, target)
			hasError = true
		} else if !lhs.IsPrimitive() {
			str := globalScope.lookupStruct(lhs.String())
			if str == nil {
				messages.Complain(diagnostic.CastError, original.Location, "Cannot typecast %s to %s. Could not find definition of '%s'", lhs, target, lhs)
				hasError = true
			} else {
				castBlock := str.innerScope.lookupNamedBlock("cast")
				if castBlock == nil {
					messages.Complain(diagnostic.CastError, targetType.Location, "Cannot typecast %s to %s. To support typecasting add a cast block with a function returning the target type", lhs, target)
					hasError = true
				} else if !castBlock.HasReturnType(target) {
					// TODO: check if the matching function has a body
					messages.Complain(diagnostic.CastError, targetType.Location, "Cannot typecast %s to %s. To support this typecasting add a function returning target type to cast block", lhs, target)
					hasError = true
				}
			}
		} else if !slices.Contains(datatypes.NumericTypes, lhs) && !slices.Contains(datatypes.NumericTypes, target) {
			messages.Complain(diagnostic.CastError, targetType.Location, typeCastError, lhs, target)
			hasError = true
		}

	}
	return target, hasError
}

func evalIndex(left *parser.AST, right *parser.AST, expectedType datatypes.DataType) (datatypes.DataType, bool) {
	lhs, lHasErr := evalType(left, expectedType)
	rhs, rHasErr := evalType(right, expectedType)
	var nodeType datatypes.DataType
	hasError := lHasErr || rHasErr
	if lhs != datatypes.String {
		messages.Complain(diagnostic.TypeError, left.Location, "Cannot index type %s", lhs.String())
		hasError = true
	}
	if right.Label == "slice" {
		// check slice type?
		nodeType = datatypes.String
	} else {
		if lhs != datatypes.String && !slices.Contains(datatypes.IntTypes, rhs) {
			messages.Complain(diagnostic.TypeError, right.Location, "Cannot index String with %s type", rhs)
			hasError = true
		}
		nodeType = datatypes.Char
	}
	return nodeType, hasError
}

func evalSlice(ast *parser.AST, expectedType datatypes.DataType) (datatypes.DataType, bool) {
	length := len(ast.Children)
	var nodeType datatypes.DataType
	hasError := false
	var err error = nil
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
		operand, hasErr := evalType(&expr, expectedType)
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
		lhs, lHasErr := evalType(&ast.Children[0], expectedType)
		rhs, rHasErr := evalType(&ast.Children[2], expectedType)
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
	return nodeType, hasError
}

func evalArrayEnd(exprNode *parser.AST, expectedType datatypes.DataType) (datatypes.DataType, bool) {
	expr, hasError := evalType(exprNode, expectedType)
	if !hasError && !slices.Contains(datatypes.IntTypes, expr) {
		messages.Complain(diagnostic.TypeError, exprNode.Location, "Cannot use %s as array end value", expr)
		hasError = true
	}
	return expr, hasError
}

func evalRange(ast *parser.AST, expectedType datatypes.DataType) (datatypes.DataType, bool) {
	var err error = nil
	var nodeType datatypes.DataType = datatypes.None
	parts := len(ast.Children)
	first, firstHasErr := evalType(&ast.Children[0], expectedType)
	second, secondHasErr := evalType(&ast.Children[2], expectedType)
	hasError := firstHasErr || secondHasErr
	if !firstHasErr && !slices.Contains(datatypes.IntTypes, first) {
		messages.Complain(diagnostic.TypeError, ast.Children[0].Location, "Cannot use %s as range start", first)
		hasError = true
	}
	if !secondHasErr && !slices.Contains(datatypes.IntTypes, second) {
		messages.Complain(diagnostic.TypeError, ast.Children[1].Location, "Cannot use %s as range end", second)
		hasError = true
	}
	nodeType, err = decideNumberType(first, second, ast.Children[1].Token.Value)
	if err != nil {
		messages.Complain(diagnostic.TypeError, ast.Children[2].Location, "%s", err.Error())
		hasError = true
	}
	if parts == 5 { // first .. second .. third
		if ast.Children[3].Token.Value == "..=" {
			messages.Complain(diagnostic.IllegalStatementError, ast.Children[3].Location, "Cannot use inclusive operator for range step value")
			hasError = true
		}
		third, thirdHasErr := evalType(&ast.Children[4], nodeType)
		hasError = hasError || thirdHasErr
		if !thirdHasErr && !slices.Contains(datatypes.IntTypes, third) {
			messages.Complain(diagnostic.TypeError, ast.Children[4].Location, "Cannot use %s as range step value", third)
			hasError = true
		}
		nodeType, err = decideNumberType(nodeType, third, "..")
		if err != nil {
			messages.Complain(diagnostic.TypeError, ast.Children[4].Location, "%s", err.Error())
			hasError = true
		}
	}
	return nodeType, hasError
}

func evalExponent(baseNode *parser.AST, expoNode *parser.AST, expectedType datatypes.DataType) (datatypes.DataType, bool) {
	base, baseHasErr := evalType(baseNode, expectedType)
	expo, expoHasErr := evalType(expoNode, expectedType)
	if baseHasErr || expoHasErr {
		return datatypes.None, true
	}
	var nodeType datatypes.DataType = datatypes.None
	hasError := false
	if !slices.Contains(datatypes.NumericTypes, base) || !slices.Contains(datatypes.NumericTypes, expo) {
		messages.Complain(diagnostic.TypeError, expoNode.Location, "Cannot use exponent with types %s and %s", base, expo)
		hasError = true
	} else {
		// determine type
		if slices.Contains(datatypes.FloatTypes, expo) {
			nodeType = expo
		} else {
			nodeType = base
		}
	}
	return nodeType, hasError
}

func evalLogicalOperation(left *parser.AST, right *parser.AST, operator lexer.Token, expectedType datatypes.DataType) (datatypes.DataType, bool) {
	lhs, lHasErr := evalType(left, expectedType)
	rhs, rHasErr := evalType(right, expectedType)
	if lHasErr || rHasErr {
		return datatypes.None, true
	}
	if lhs != datatypes.Bool || rhs != datatypes.Bool {
		messages.Complain(diagnostic.TypeError, operator.Location, "Cannot use operator '%s' between %s and %s", operator.Value, lhs, rhs)
		return datatypes.None, true
	} else {
		return datatypes.Bool, false
	}
}

func handleFunctionCall(details []parser.AST) (datatypes.DataType, bool) {
	hasError := false
	scope := currentScope
	var name lexer.Token
	if details[0].Label == "dot" {
		name = details[0].Children[1].Token
		lhs, hasErr := handleDot(details[0].Children[0], details[0].Children[1], true, false, false)
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
			}
		} else if lhs.IsRef() {
			scope = FindAncestorScopeById(lhs.GetScopes()[0])
		} else {
			symbol := globalScope.lookupType(lhs.String())
			if symbol == nil {
				messages.Complain(diagnostic.NameError, details[0].Children[1].Location, "Could not find type %s", lhs)
				return datatypes.None, true
			}

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
			if !fn.hasDefaultImplementation && !currentScope.HasScopeTypeAncestor(Interface) && symbol.returnType != datatypes.None {
				messages.Complain(diagnostic.CallError, details[0].Location, "Function without body cannot be called")
				hasError = true
			}
			return symbol.returnType, hasError
		}
	} else {
		// TODO: find closest error
		messages.Complain(diagnostic.CallError, details[0].Location, "Could not find function '%s(%s)->%s'", name.Value, paramList, symbol.returnType)
		hasError = true
	}
	return datatypes.None, hasError
}

func handleDot(left parser.AST, right parser.AST, isFnCall bool, isAssignment bool, recursed bool) (datatypes.DataType, bool) {
	hasError := false
	var lhs datatypes.DataType = datatypes.None
	if left.Label != "dot" {
		lhs, hasError = evalType(&left, datatypes.None)
		if isFnCall && !recursed {
			return lhs, hasError
		}
	} else {
		lhs, hasError = handleDot(left.Children[0], left.Children[1], isFnCall, isAssignment, true)
	}
	rname := right.Token.Value
	if lhs == datatypes.String && rname == "length" {
		if isAssignment {
			messages.Complain(diagnostic.AccessError, left.Location, "Cannot assign value to string length")
			hasError = true
		}
		return datatypes.Int32, hasError
	} else if isFnCall && (lhs.IsScopeRef() || lhs.IsRef()) {
		return lhs, hasError
	} else if !lhs.IsPrimitive() {
		variable := currentScope.lookupVariable(left.Token.Value)
		if variable != nil {
			if isAssignment && !variable.isMutable {
				messages.Complain(diagnostic.AccessError, left.Location, "Cannot change value of immutable property or variable")
				hasError = true
			}
		}
		if lhs.IsRef() {
			scope := FindAncestorScopeById(lhs.GetScopes()[0])
			if prop := scope.lookupVariable(rname); prop != nil {
				if isAssignment && !prop.isMutable && !hasError {
					messages.Complain(diagnostic.AccessError, left.Location, "Cannot change value of immutable variable '%s'", prop.name)
					hasError = true
				} else {
					return prop.Type, hasError
				}
			} else {
				messages.Complain(diagnostic.NameError, right.Location, "Could not find variable %s", rname)
				return datatypes.None, hasError
			}
		}
		symbol := globalScope.lookupType(lhs.String())
		if symbol == nil {
			messages.Complain(diagnostic.NameError, left.Location, "Could not find type %s", lhs)
			hasError = true
		} else {
			scope := symbol.getInnerScope()
			if prop := scope.lookupVariable(rname); prop != nil {
				if prop.isPrivate && !currentScope.HasParentScope(symbol.getInnerScope()) {
					messages.Complain(diagnostic.AccessError, right.Location, "Cannot access private property from outside struct definition")
					hasError = true
				} else if isAssignment && !prop.isMutable && !hasError {
					messages.Complain(diagnostic.AccessError, left.Location, "Cannot change value of immutable property or variable")
					hasError = true
				} else {
					return prop.Type, hasError
				}
			} else if !hasError {
				messages.Complain(diagnostic.NameError, right.Location, "Could not find property %s in type %s", rname, lhs)
				hasError = true
			}
		}
	} else if !hasError {
		messages.Complain(diagnostic.TypeError, right.Location, "Cannot access property %s of type %s", rname, lhs)
		hasError = true
	}
	return lhs, hasError
}
