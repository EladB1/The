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

func evalLiteral(ast *parser.AST, expectedType datatypes.DataType) datatypes.DataType {
	if ast.Label == "struct_literal" {
		return evalStructLiteral(ast)
	}
	switch ast.Token.Kind {
	case lexer.LIT_CHAR:
		return datatypes.Char
	case lexer.LIT_STRING:
		return datatypes.String
	case lexer.LIT_FLOAT:
		if expectedType == datatypes.Float {
			return datatypes.Float
		}
		return datatypes.Double
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
		if intf := globalScope.lookupInterface(name); intf != nil {
			messages.Complain(diagnostic.IllegalStatementError, ast.Location, "Cannot create interface literal value")
		} else {
			messages.Complain(diagnostic.NameError, ast.Location, "Struct %s not defined", name)
		}
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
		if valueType, hasErr := evalType(&value, property.Type); property.Type != valueType && !ImplementsInterface(property.Type, valueType) && !hasErr {
			messages.Complain(diagnostic.TypeError, value.Location, "Property type %s expected but found %s", property.Type, valueType)
		}
		visited.Append(propId)
	}
	return datatypes.DynamicType(name)
}

func evalType(ast *parser.AST, expectedType datatypes.DataType) (datatypes.DataType, bool) {
	hasError := false
	var nodeType datatypes.DataType = datatypes.None
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
		nodeType, hasError = evalUnary(&ast.Children[0], &ast.Children[1], expectedType)
	} else if ast.Label == "typecast" {
		nodeType, hasError = evalTypecast(&ast.Children[0], &ast.Children[1], expectedType)
	} else if ast.Label == "index" {
		nodeType, hasError = evalIndex(&ast.Children[0], &ast.Children[1], expectedType)
	} else if ast.Label == "slice" {
		nodeType, hasError = evalSlice(ast, expectedType)
	} else if ast.Label == "ARR-END" {
		nodeType, hasError = evalArrayEnd(&ast.Children[0], expectedType)
	} else if ast.Label == "range" {
		nodeType, hasError = evalRange(ast, expectedType)
	} else if ast.Label == "call" {
		nodeType, hasError = handleFunctionCall(ast.Children)
	} else if ast.Token.Kind == lexer.OPERATOR_ADD {
		nodeType, hasError = evalAdd(&ast.Children[0], &ast.Children[1], ast.Token, expectedType)
	} else if ast.Token.Kind == lexer.OPERATOR_MULT {
		nodeType, hasError = evalMult(&ast.Children[0], &ast.Children[1], ast.Token, expectedType)
	} else if ast.Token.Kind == lexer.OPERATOR_BS || ast.Token.Kind == lexer.OPERATOR_BW {
		nodeType, hasError = evalBitOperation(&ast.Children[0], &ast.Children[1], ast.Token, expectedType)
	} else if ast.Token.Kind == lexer.OPERATOR_COMPARE {
		nodeType, hasError = evalCompare(&ast.Children[0], &ast.Children[1], ast.Token, expectedType)
	} else if ast.Token.Value == "&&" || ast.Token.Value == "||" {
		nodeType, hasError = evalLogicalOperation(&ast.Children[0], &ast.Children[1], ast.Token, expectedType)
	} else if ast.Token.Value == "**" {
		nodeType, hasError = evalExponent(&ast.Children[0], &ast.Children[1], expectedType)
	} else if ast.Label == "dot" {
		nodeType, hasError = handleDot(ast.Children[0], ast.Children[1], false)
	}
	if !hasError {
		ast.Type = nodeType
	}
	return nodeType, hasError
}

func comparableCheck(lhs datatypes.DataType, rhs datatypes.DataType) bool {
	lhsUnsigned := slices.Contains(datatypes.UnsignedTypes, lhs)
	rhsUnsigned := slices.Contains(datatypes.UnsignedTypes, rhs)
	lhsSigned := slices.Contains(datatypes.SignedIntTypes, lhs) || slices.Contains(datatypes.FloatTypes, lhs)
	rhsSigned := slices.Contains(datatypes.SignedIntTypes, rhs) || slices.Contains(datatypes.FloatTypes, rhs)

	return lhs == rhs || (lhsUnsigned && rhsUnsigned) || (lhsSigned && rhsSigned)
}

func handleBinaryNumberExpression(left *parser.AST, right *parser.AST, operator string, expectedType datatypes.DataType) (datatypes.DataType, error) {
	inferLeft := left.IsLiteralExpression() && !right.IsLiteralExpression()
	inferRight := !left.IsLiteralExpression() && right.IsLiteralExpression()
	var lhs, rhs datatypes.DataType
	var lHasErr, rHasErr bool
	if inferLeft {
		rhs, rHasErr = evalType(right, expectedType)
		lhs, lHasErr = evalType(left, rhs)
	} else if inferRight {
		lhs, lHasErr = evalType(left, expectedType)
		rhs, rHasErr = evalType(right, lhs)
	}
	if lHasErr || rHasErr {
		return datatypes.None, nil
	}
	return decideNumberType(lhs, rhs, operator)
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
