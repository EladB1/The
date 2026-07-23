package semantic

import (
	"sort"
	"strings"

	dt "github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/diagnostic"
	"github.com/EladB1/The/internal/lexer"
	"github.com/EladB1/The/internal/parser"
)

func analyzeForCondition(condition []*parser.AST) {
	parts := len(condition)
	if condition[0].Label == "Variable" {
		symbol := analyzeVariable(condition[0])
		if symbol != nil {
			currentScope.Variables[symbol.Name] = *symbol
		}
		if condition[parts-2].Token.Value == "in" {
			if parts == 4 { // int i, char c in string
				symbol2 := analyzeVariable(condition[1])
				if symbol2 != nil {
					currentScope.Variables[symbol2.Name] = *symbol2
				}
				if symbol != nil && symbol2 != nil && !(symbol.Type.IsIntType() && symbol2.Type.Equals(dt.CharType) || (symbol.Type.Equals(dt.CharType) && symbol2.Type.IsIntType())) {
					messages.Complain(diagnostic.TypeError, condition[2].Location, "Cannot use %s and %s as loop variables", symbol.Type.String(), symbol2.Type.String())
				}

			} else {
				if condition[parts-1].Label == "range" {
					if symbol != nil {
						expr, hasErr := evalType(condition[parts-1], symbol.Type)
						if !hasErr && !expr.Equals(symbol.Type) {
							messages.Complain(diagnostic.TypeError, condition[parts-1].Location, "Variable of type %s not compatible with range expression of type %s", symbol.Type, expr)
						}
					}
				} else { // char c in string
					if symbol != nil && !symbol.Type.Equals(dt.CharType) {
						messages.Complain(diagnostic.TypeError, condition[0].Location, "Cannot use %s as loop variables", symbol.Type.String())
					}
					rhs, hasErr := evalType(condition[parts-1], dt.StringType)
					if !hasErr && !rhs.Equals(dt.StringType) {
						messages.Complain(diagnostic.TypeError, condition[parts-1].Location, "Cannot loop over %s", rhs.String())
					}
				}
			}
		} else {
			if symbol != nil {
				cond, hasErr := evalType(condition[1], dt.BoolType)
				if !hasErr && !cond.Equals(dt.BoolType) {
					messages.Complain(diagnostic.TypeError, condition[1].Location, "Expected bool as loop condition but got %s", cond.String())
				}
				var expr dt.SourceType
				if condition[2].Token.Kind == lexer.OPERATOR_ASSIGN {
					expr, hasErr = analyzeAssignment(condition[2])

				} else {
					expr, hasErr = evalType(condition[2], symbol.Type)
				}
				if !hasErr && !symbol.Type.Equals(expr) {
					messages.Complain(diagnostic.TypeError, condition[2].Location, "Expected %s as loop expression but got %s", symbol.Type.String(), expr.String())
				}
			}
		}
	} else {
		iter, _ := analyzeAssignment(condition[0])
		if !iter.Equals(dt.NoneType) {
			cond, hasErr := evalType(condition[1], dt.BoolType)
			if !hasErr && !cond.Equals(dt.BoolType) {
				messages.Complain(diagnostic.TypeError, condition[1].Location, "Expected bool as loop condition but got %s", cond.String())
			}
			var expr dt.SourceType
			if condition[2].Token.Kind == lexer.OPERATOR_ASSIGN {
				expr, hasErr = analyzeAssignment(condition[2])

			} else {
				expr, hasErr = evalType(condition[2], iter)
			}
			if !hasErr && !iter.Equals(expr) {
				messages.Complain(diagnostic.TypeError, condition[2].Location, "Expected %s as loop expression but got %s", iter.String(), expr.String())
			}
		}

	}
}

func analyzeAssignment(stmt *parser.AST) (dt.SourceType, bool) {
	hasError := false
	left := stmt.Children[0]
	var lhs dt.SourceType
	var lHasErr bool
	if left.Label == "dot" {
		lhs, lHasErr = handleDot(left.Children[0], left.Children[1], false, true, false)
		if lHasErr {
			return dt.NoneType, true
		}
	} else {
		symbol := currentScope.LookupVariable(left.Token.Value)
		if symbol == nil {
			messages.Complain(diagnostic.NameError, left.Location, "Variable '%s' is not defined in this scope", left.Token.Value)
			return dt.NoneType, true
		}
		if !symbol.isMutable {
			messages.Complain(diagnostic.AccessError, left.Location, "Cannot change value of immutable variable '%s'", left.Token.Value)
			return dt.NoneType, true
		}
		lhs = symbol.Type
	}
	rhs, hasErr := evalType(stmt.Children[1], lhs)
	if hasErr {
		return dt.NoneType, true
	}
	stmt.Children[0].Type = lhs
	operator := stmt.Token.Value
	switch operator {
	case "+=":
		if lhs.Equals(dt.StringType) {
			if !rhs.Equals(dt.CharType) && !rhs.Equals(dt.StringType) {
				messages.Complain(diagnostic.TypeError, stmt.Children[1].Location, "Cannot concatenate %s to String", rhs.String())
				hasError = true
			}
		} else if lhs.IsNumeric() {
			result, err := decideNumberType(lhs, rhs, operator)
			if err != nil {
				messages.Complain(diagnostic.TypeError, stmt.Location, "%s", err.Error())
				hasError = true
			} else if !result.Equals(lhs) {
				messages.Complain(diagnostic.TypeError, stmt.Children[1].Location, "Cannot assign %s to %s", rhs, lhs)
				hasError = true
			}
		} else {
			messages.Complain(diagnostic.TypeError, stmt.Children[1].Location, "Invalid operation between %s to %s", rhs.String(), lhs)
			hasError = true
		}
	case "-=", "*=", "/=":
		if lhs.Equals(dt.CharType) {
			if !hasError && operator != "-=" {
				messages.Complain(diagnostic.TypeError, stmt.Location, "Cannot use operator %s on char", operator)
				hasError = true
			}
			if !hasError && !rhs.Equals(dt.CharType) {
				messages.Complain(diagnostic.TypeError, stmt.Location, "Cannot substract %s from char", rhs)
				hasError = true
			}
		} else if lhs.IsNumeric() {
			result, err := decideNumberType(lhs, rhs, operator)
			if err != nil {
				messages.Complain(diagnostic.TypeError, stmt.Location, "%s", err.Error())
				hasError = true
			} else if !result.Equals(lhs) {
				messages.Complain(diagnostic.TypeError, stmt.Children[1].Location, "Cannot assign %s to %s", rhs, lhs)
				hasError = true
			}
		} else {
			messages.Complain(diagnostic.TypeError, stmt.Children[1].Location, "Invalid operation between %s to %s", rhs.String(), lhs)
			hasError = true
		}
	default:
		if !lhs.Equals(rhs) {
			if !ImplementsInterface(lhs, rhs) {
				messages.Complain(diagnostic.TypeError, stmt.Children[1].Location, "Cannot assign %s to %s", rhs, lhs)
				hasError = true
			}
		} else if intf := globalScope.LookupInterface(rhs.String()); intf != nil {
			messages.Complain(diagnostic.IllegalStatementError, stmt.Children[1].Location, "Cannot use interface type as value")
		}

	}
	stmt.Children[1].Type = rhs
	stmt.Type = rhs
	return lhs, hasError
}

func evalAdd(left *parser.AST, right *parser.AST, operator lexer.Token, expectedType dt.SourceType) (dt.SourceType, bool) {
	hasError := false
	lhs, lHasErr := evalType(left, expectedType)
	rhs, rHasErr := evalType(right, expectedType)
	if lHasErr || rHasErr {
		return dt.NoneType, true
	}
	if operator.Value == "+" && (lhs.Equals(dt.StringType) || lhs.Equals(dt.CharType)) && (rhs.Equals(dt.StringType) || rhs.Equals(dt.CharType)) {
		return dt.StringType, hasError
	} else if operator.Value == "-" && lhs.Equals(dt.CharType) && rhs.Equals(dt.CharType) {
		return dt.CharType, hasError
	} else if lhs.IsNumeric() && rhs.IsNumeric() {
		if lhs.Equals(rhs) {
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
	return dt.NoneType, hasError
}

func evalMult(left *parser.AST, right *parser.AST, operator lexer.Token, expectedType dt.SourceType) (dt.SourceType, bool) {
	hasError := false
	lhs, lHasErr := evalType(left, expectedType)
	rhs, rHasErr := evalType(right, expectedType)
	if lHasErr || rHasErr {
		return dt.NoneType, true
	}
	if !lhs.IsNumeric() || !rhs.IsNumeric() {
		messages.Complain(diagnostic.TypeError, operator.Location, "Cannot use operator '%s' between %s and %s", operator.Value, lhs, rhs)
		hasError = true
	} else {
		nodeType, err := handleBinaryNumberExpression(left, right, operator.Value, expectedType)
		if err != nil {
			if lhs.IsUnsignedType() && rhs.IsSignedIntType() && left.Token.Kind == lexer.ID && right.IsLiteral() {
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
	return dt.NoneType, hasError
}

func evalBitOperation(left *parser.AST, right *parser.AST, operator lexer.Token, expectedType dt.SourceType) (dt.SourceType, bool) {
	hasError := false
	lhs, lHasErr := evalType(left, expectedType)
	rhs, rHasErr := evalType(right, expectedType)
	if lHasErr || rHasErr {
		return dt.NoneType, true
	}
	if (lhs.Equals(dt.Uint32Type) || lhs.Equals(dt.Uint64Type)) && (rhs.Equals(dt.Uint32Type) || rhs.Equals(dt.Uint64Type)) ||
		(lhs.Equals(dt.Int32Type) || lhs.Equals(dt.Int64Type)) && (rhs.Equals(dt.Int32Type) || rhs.Equals(dt.Int64Type)) {
		nodeType, _ := decideNumberType(lhs, rhs, operator.Value)
		return nodeType, hasError
	}
	messages.Complain(diagnostic.TypeError, operator.Location, "Cannot use operator '%s' between %s and %s", operator.Value, lhs, rhs)
	return dt.NoneType, true
}

func evalCompare(left *parser.AST, right *parser.AST, operator lexer.Token, expectedType dt.SourceType) (dt.SourceType, bool) {
	hasError := false
	lhs, lHasErr := evalType(left, expectedType)
	rhs, rHasErr := evalType(right, expectedType)
	if lHasErr || rHasErr {
		return dt.NoneType, true
	}
	if !comparableCheck(lhs, rhs) {
		if lhs.IsNumeric() && rhs.IsNumeric() {
			_, err := handleBinaryNumberExpression(left, right, operator.Value, expectedType)
			if err == nil {
				return dt.BoolType, hasError
			} else {
				messages.Complain(diagnostic.TypeError, operator.Location, "%s", err.Error())
				hasError = true
			}
		}
		messages.Complain(diagnostic.TypeError, operator.Location, "Invalid comparison between %s and %s", lhs, rhs)
		return dt.NoneType, true
	}
	if lhs.Equals(rhs) && lhs.IsDynamic && operator.Value != "==" && operator.Value != "!=" {
		str := globalScope.lookupStruct(lhs.String())
		if str == nil {
			messages.Complain(diagnostic.NameError, operator.Location, "Cannot find struct definition for %s", lhs.String())
		} else {
			compareBlock := str.InnerScope.LookupNamedBlock("compare")
			if compareBlock == nil {
				messages.Complain(diagnostic.TypeError, operator.Location, "Cannot compare %s using operator '%s'. To support this comparison add a compare block with the appropriate functions", lhs, operator.Value)
			} else {
				switch operator.Value {
				case "<", "<=":
					if symbol := compareBlock.InnerScope.LookupFunctionByName("lessThan"); symbol == nil || !symbol.ReturnType.Equals(dt.BoolType) || !symbol.Overloads[0].HasDefaultImplementation {
						messages.Complain(diagnostic.TypeError, operator.Location, "Unsupported comparison. To support operators '<' and '<=', add function 'fn lessThan(%s)->bool' to compare block in %s definition", lhs, lhs)
						hasError = true
					}
				case ">", ">=":
					if symbol := compareBlock.InnerScope.LookupFunctionByName("greaterThan"); symbol == nil || !symbol.ReturnType.Equals(dt.BoolType) || !symbol.Overloads[0].HasDefaultImplementation {
						messages.Complain(diagnostic.TypeError, operator.Location, "Unsupported comparison. To support operators '>' and '>=', add function 'fn greaterThan(%s)->bool' to compare block in %s definition", lhs, lhs)
						hasError = true
					}
				}
			}
		}
	}
	return dt.BoolType, hasError
}

func evalUnary(left *parser.AST, right *parser.AST, expectedType dt.SourceType) (dt.SourceType, bool) {
	hasError := false
	hasErr := false
	var nodeType dt.SourceType = dt.NoneType
	if leftTok := left.Token; leftTok.Kind == lexer.OPERATOR_UNARY || leftTok.Value == "-" { // left unary
		rhs, hasErr := evalType(right, expectedType)
		hasError = hasError || hasErr
		switch leftTok.Value {
		case "!":
			if !rhs.Equals(dt.BoolType) {
				messages.Complain(diagnostic.TypeError, right.Location, "Must use bool value with unary '!'")
				hasError = true
			} else {
				nodeType = dt.BoolType
			}
		case "-":
			if !rhs.IsNumeric() {
				messages.Complain(diagnostic.TypeError, right.Location, "Cannot use unary '-' with type %s", rhs)
				hasError = true
			} else {
				nodeType = rhs
			}
		case "~":
			if !rhs.IsIntType() {
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

func checkIncrementOperator(operand lexer.Token, operator lexer.Token) (dt.SourceType, bool) {
	hasError := false
	var nodeType dt.SourceType = dt.NoneType
	symbol := currentScope.LookupVariable(operand.Value)
	if symbol != nil {
		if !symbol.Type.IsNumeric() {
			messages.Complain(diagnostic.TypeError, operand.Location, "Cannot use '%s' with type %s", operator.Value, symbol.Type)
			hasError = true
		} else if !symbol.isMutable {
			messages.Complain(diagnostic.AccessError, operand.Location, "Cannot change value of immutable variable '%s'", symbol.Name)
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

func evalTypecast(original *parser.AST, targetType *parser.AST, expectedType dt.SourceType) (dt.SourceType, bool) {
	lhs, hasErr := evalType(original, expectedType)
	hasError := hasErr
	target := nodeToType(targetType)
	typeCastError := "Typecasting from %s to %s not allowed"
	if !lhs.Equals(target) && !target.Equals(dt.StringType) {
		if lhs.Equals(dt.Uint64Type) || lhs.Equals(dt.Int64Type) || lhs.Equals(dt.DoubleType) {
			if target.Equals(dt.Uint32Type) || target.Equals(dt.FloatType) || target.Equals(dt.Int32Type) {
				messages.Warn(targetType.Location, "Lossy conversion from %s to %s", lhs, target)
			} else if !target.IsNumeric() {
				messages.Complain(diagnostic.CastError, targetType.Location, typeCastError, lhs, target)
				hasError = true
			}
		} else if lhs.Equals(dt.StringType) {
			messages.Complain(diagnostic.CastError, targetType.Location, typeCastError, lhs, target)
			hasError = true
		} else if lhs.IsDynamic {
			str := globalScope.lookupStruct(lhs.String())
			if str == nil {
				messages.Complain(diagnostic.CastError, original.Location, "Cannot typecast %s to %s. Could not find definition of '%s'", lhs, target, lhs)
				hasError = true
			} else {
				castBlock := str.InnerScope.LookupNamedBlock("cast")
				if castBlock == nil {
					messages.Complain(diagnostic.CastError, targetType.Location, "Cannot typecast %s to %s. To support typecasting add a cast block with a function returning the target type", lhs, target)
					hasError = true
				} else if !castBlock.HasReturnType(target) {
					messages.Complain(diagnostic.CastError, targetType.Location, "Cannot typecast %s to %s. To support this typecasting add a function returning target type to cast block", lhs, target)
					hasError = true
				}
			}
		} else if !lhs.IsNumeric() && !target.IsNumeric() {
			messages.Complain(diagnostic.CastError, targetType.Location, typeCastError, lhs, target)
			hasError = true
		}

	}
	return target, hasError
}

func evalIndex(left *parser.AST, right *parser.AST, expectedType dt.SourceType) (dt.SourceType, bool) {
	lhs, lHasErr := evalType(left, expectedType)
	rhs, rHasErr := evalType(right, expectedType)
	var nodeType dt.SourceType
	hasError := lHasErr || rHasErr
	if !lhs.Equals(dt.StringType) {
		messages.Complain(diagnostic.TypeError, left.Location, "Cannot index type %s", lhs.String())
		hasError = true
	}
	if right.Label == "slice" {
		// check slice type?
		nodeType = dt.StringType
	} else {
		if !lhs.Equals(dt.StringType) && !rhs.IsIntType() {
			messages.Complain(diagnostic.TypeError, right.Location, "Cannot index String with %s type", rhs)
			hasError = true
		}
		nodeType = dt.CharType
	}
	return nodeType, hasError
}

func evalSlice(ast *parser.AST, expectedType dt.SourceType) (dt.SourceType, bool) {
	length := len(ast.Children)
	var nodeType dt.SourceType
	hasError := false
	var err error = nil
	switch length {
	case 1:
		nodeType = dt.Int32Type
	case 2:
		var expr *parser.AST
		if ast.Children[0].Token.Kind == lexer.OPERATOR_RANGE {
			expr = ast.Children[1]
		} else {
			expr = ast.Children[0]
		}
		operand, hasErr := evalType(expr, expectedType)
		if hasErr {
			hasError = hasErr
		}
		if !operand.IsIntType() {
			messages.Complain(diagnostic.TypeError, expr.Location, "Invalid type %s used in range expression", operand)
			hasError = true
		} else {
			nodeType = operand
		}
	case 3:
		lhs, lHasErr := evalType(ast.Children[0], expectedType)
		rhs, rHasErr := evalType(ast.Children[2], expectedType)
		if lHasErr || rHasErr {
			hasError = true
		} else {
			if !lhs.IsIntType() || !rhs.IsIntType() {
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

func evalArrayEnd(exprNode *parser.AST, expectedType dt.SourceType) (dt.SourceType, bool) {
	expr, hasError := evalType(exprNode, expectedType)
	if !hasError && !expr.IsIntType() {
		messages.Complain(diagnostic.TypeError, exprNode.Location, "Cannot use %s as array end value", expr)
		hasError = true
	}
	return expr, hasError
}

func evalRange(ast *parser.AST, expectedType dt.SourceType) (dt.SourceType, bool) {
	var err error = nil
	var nodeType dt.SourceType = dt.NoneType
	parts := len(ast.Children)
	first, firstHasErr := evalType(ast.Children[0], expectedType)
	second, secondHasErr := evalType(ast.Children[2], expectedType)
	hasError := firstHasErr || secondHasErr
	if !firstHasErr && !first.IsIntType() {
		messages.Complain(diagnostic.TypeError, ast.Children[0].Location, "Cannot use %s as range start", first)
		hasError = true
	}
	if !secondHasErr && !second.IsIntType() {
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
		third, thirdHasErr := evalType(ast.Children[4], nodeType)
		hasError = hasError || thirdHasErr
		if !thirdHasErr && !third.IsIntType() {
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

func evalExponent(baseNode *parser.AST, expoNode *parser.AST, expectedType dt.SourceType) (dt.SourceType, bool) {
	base, baseHasErr := evalType(baseNode, expectedType)
	expo, expoHasErr := evalType(expoNode, expectedType)
	if baseHasErr || expoHasErr {
		return dt.NoneType, true
	}
	var nodeType dt.SourceType = dt.NoneType
	hasError := false
	if !base.IsNumeric() || !expo.IsNumeric() {
		messages.Complain(diagnostic.TypeError, expoNode.Location, "Cannot use exponent with types %s and %s", base, expo)
		hasError = true
	} else {
		// determine type
		if expo.IsFloatType() {
			nodeType = expo
		} else {
			nodeType = base
		}
	}
	return nodeType, hasError
}

func evalLogicalOperation(left *parser.AST, right *parser.AST, operator lexer.Token, expectedType dt.SourceType) (dt.SourceType, bool) {
	lhs, lHasErr := evalType(left, expectedType)
	rhs, rHasErr := evalType(right, expectedType)
	if lHasErr || rHasErr {
		return dt.NoneType, true
	}
	if !lhs.Equals(dt.BoolType) || !rhs.Equals(dt.BoolType) {
		messages.Complain(diagnostic.TypeError, operator.Location, "Cannot use operator '%s' between %s and %s", operator.Value, lhs, rhs)
		return dt.NoneType, true
	} else {
		return dt.BoolType, false
	}
}

func handleFunctionCall(details []*parser.AST) (dt.SourceType, bool) {
	hasError := false
	scope := currentScope
	var name lexer.Token
	if details[0].Label == "dot" {
		name = details[0].Children[1].Token
		lhs, hasErr := handleDot(details[0].Children[0], details[0].Children[1], true, false, false)
		if hasErr {
			hasError = hasErr
		}
		if lhs.RootEquals(dt.ScopeRef) {
			scopes := lhs.SubTypes
			if len(scopes) < 2 {
				messages.Complain(diagnostic.ReferenceError, details[0].Location, "Could not find reference value")
				hasError = true
			} else {
				str := globalScope.lookupStruct(scopes[0].String())
				if str == nil {
					messages.Complain(diagnostic.NameError, details[0].Location, "Could not find struct %s", scopes[0])
					hasError = true
				}
				nb := str.InnerScope.LookupNamedBlock(scopes[1].String())
				if nb == nil {
					messages.Complain(diagnostic.NameError, details[0].Location, "Could not find named block %s in struct %s", scopes[1], scopes[0])
					hasError = true
				}
				scope = nb.InnerScope
			}
		} else if lhs.RootEquals(dt.Ref) || lhs.Equals(dt.GlobalRefType) {
			scope = FindAncestorScopeById(lhs.SubTypes[0].String())
		} else if mem, ok := PrimitiveMembers[lhs.Root]; ok {
			returnType := dt.NoneType
			if method, ok := mem.Methods[name.Value]; ok {
				var params []dt.SourceType = []dt.SourceType{}
				if len(details) == 2 {
					for _, param := range details[1].Children {
						parameter, hasErr := evalType(param, dt.NoneType)
						params = append(params, parameter)
						if hasErr {
							hasError = hasErr
						}
					}
				}
				paramList := dt.JoinTypes(params)
				if fn := method.getMatchingOverload(params); fn != nil {
					returnType = method.ReturnType
				} else {
					// TODO: find closest error
					messages.Complain(diagnostic.CallError, details[0].Location, "Could not find function '%s(%s)->%s'", name.Value, paramList, method.ReturnType)
					hasError = true
				}
			} else {
				messages.Complain(diagnostic.NameError, details[0].Location, "Could not find function %s", name.Value)
				hasError = true
			}
			return returnType, hasError
		} else {
			symbol := globalScope.LookupType(lhs.String())
			if symbol == nil {
				messages.Complain(diagnostic.NameError, details[0].Children[1].Location, "Could not find type %s", lhs)
				return dt.NoneType, true
			}

			scope = symbol.GetInnerScope()
			if symbol.GetSymbolType() == "struct" {
				if conflicts := symbol.getConflicts(name.Value); len(conflicts) > 1 {
					sort.Strings(conflicts)
					messages.Complain(diagnostic.AmbiguityError, name.Location, "Interfaces %s both contain function named %s. Change the function call to pick which one to use", strings.Join(conflicts, ","), name.Value)
					return dt.NoneType, true
				} else if len(conflicts) == 1 {
					if nb := scope.LookupNamedBlock(conflicts[0]); nb != nil {
						scope = nb.InnerScope
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
	symbol := scope.LookupFunctionByName(name.Value)
	if symbol == nil {
		messages.Complain(diagnostic.NameError, name.Location, "Could not find function %s in scope", name.Value)
		return dt.NoneType, true
	}
	// check parameters
	var params []dt.SourceType = []dt.SourceType{}
	if len(details) == 2 {
		for _, param := range details[1].Children {
			parameter, hasErr := evalType(param, dt.NoneType)
			params = append(params, parameter)
			if hasErr {
				hasError = hasErr
			}
		}
	}
	paramList := dt.JoinTypes(params)
	if fn := symbol.getMatchingOverload(params); fn != nil {
		if fn.IsPrivate {
			messages.Complain(diagnostic.AccessError, details[0].Location, "Cannot access private function '%s' from outside struct definition", name.Value)
			hasError = true
		} else {
			if !fn.HasDefaultImplementation && !currentScope.HasScopeTypeAncestor(Interface) && !symbol.ReturnType.Equals(dt.NoneType) {
				messages.Complain(diagnostic.CallError, details[0].Location, "Function without body cannot be called")
				hasError = true
			}
			return symbol.ReturnType, hasError
		}
	} else {
		// TODO: find closest error
		messages.Complain(diagnostic.CallError, details[0].Location, "Could not find function '%s(%s)->%s'", name.Value, paramList, symbol.ReturnType)
		hasError = true
	}
	return dt.NoneType, hasError
}

func handleDot(left *parser.AST, right *parser.AST, isFnCall bool, isAssignment bool, recursed bool) (dt.SourceType, bool) {
	hasError := false
	var lhs dt.SourceType = dt.NoneType
	if left.Label != "dot" {
		lhs, hasError = evalType(left, dt.NoneType)
		if isFnCall && !recursed {
			return lhs, hasError
		}
	} else {
		lhs, hasError = handleDot(left.Children[0], left.Children[1], isFnCall, isAssignment, true)
	}
	rname := right.Token.Value
	if mem, ok := PrimitiveMembers[lhs.Root]; ok {
		propType := dt.NoneType
		if isAssignment {
			messages.Complain(diagnostic.AccessError, left.Location, "Cannot assign value to %s.%s", lhs.Root, rname)
			hasError = true
		} else {
			if prop, ok := mem.Properties[rname]; ok {
				propType = prop.Type
			} else {
				messages.Complain(diagnostic.NameError, left.Location, "Could not find %s.%s", lhs.Root, rname)
				hasError = true
			}
		}
		return propType, hasError
	} else if isFnCall && (lhs.RootEquals(dt.ScopeRef) || lhs.RootEquals(dt.Ref) || lhs.Equals(dt.GlobalRefType)) {
		return lhs, hasError
	} else if lhs.IsDynamic {
		variable := currentScope.LookupVariable(left.Token.Value)
		if variable != nil {
			if isAssignment && !variable.isMutable {
				messages.Complain(diagnostic.AccessError, left.Location, "Cannot change value of immutable property or variable")
				hasError = true
			}
		}
		symbol := globalScope.LookupType(lhs.String())
		if symbol == nil {
			messages.Complain(diagnostic.NameError, left.Location, "Could not find type %s", lhs)
			hasError = true
		} else {
			scope := symbol.GetInnerScope()
			if prop := scope.LookupVariable(rname); prop != nil {
				if prop.isPrivate && !currentScope.HasParentScope(symbol.GetInnerScope()) {
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
	} else if lhs.RootEquals(dt.Ref) || lhs.Equals(dt.GlobalRefType) {
		scope := FindAncestorScopeById(lhs.SubTypes[0].String())
		if prop := scope.LookupVariable(rname); prop != nil {
			if isAssignment && !prop.isMutable && !hasError {
				messages.Complain(diagnostic.AccessError, left.Location, "Cannot change value of immutable variable '%s'", prop.Name)
				hasError = true
			} else {
				return prop.Type, hasError
			}
		} else {
			messages.Complain(diagnostic.NameError, right.Location, "Could not find variable %s", rname)
			return dt.NoneType, hasError
		}
	} else if !hasError {
		messages.Complain(diagnostic.TypeError, right.Location, "Cannot access property %s of type %s", rname, lhs)
		hasError = true
	}
	return lhs, hasError
}
