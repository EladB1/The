package semantic

import (
	"fmt"

	ds "github.com/EladB1/The/internal/datastructures"
	dt "github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/diagnostic"
	"github.com/EladB1/The/internal/lexer"
	"github.com/EladB1/The/internal/parser"
)

func evalLiteral(ast *parser.AST, expectedType dt.SourceType) dt.SourceType {
	if ast.Label == "struct_literal" {
		return evalStructLiteral(ast)
	}
	switch ast.Token.Kind {
	case lexer.LIT_CHAR:
		return dt.CharType
	case lexer.LIT_STRING:
		return dt.StringType
	case lexer.LIT_FLOAT:
		if expectedType.Equals(dt.FloatType) {
			return dt.FloatType
		}
		return dt.DoubleType
	case lexer.LIT_INT, lexer.LIT_HEX:
		if expectedType.IsNumeric() {
			return expectedType
		}
		return dt.Int32Type
	case lexer.KW_BOOLVALUE:
		return dt.BoolType
	}
	return dt.NoneType
}

func evalStructLiteral(ast *parser.AST) dt.SourceType {
	name := ast.Children[0].Token.Value
	symbol := globalScope.lookupStruct(name)
	if symbol == nil {
		if intf := globalScope.LookupInterface(name); intf != nil {
			messages.Complain(diagnostic.IllegalStatementError, ast.Location, "Cannot create interface literal value")
		} else {
			messages.Complain(diagnostic.NameError, ast.Location, "Struct %s not defined", name)
		}
		return dt.NoneType
	}
	if len(ast.Children) == 1 {
		return dt.NewDynamicType(name)
	}
	properties := ast.Children[1].Children
	visited := ds.HashSet{}
	for _, prop := range properties {
		propId := prop.Children[0].Token.Value
		innerScope := symbol.InnerScope
		property := innerScope.LookupVariable(propId)
		if property == nil {
			messages.Complain(diagnostic.NameError, prop.Location, "Could not find property '%s' in struct %s", propId, symbol.Name)
			continue
		}
		if _, ok := visited[propId]; ok {
			messages.Complain(diagnostic.IllegalStatementError, prop.Location, "Cannot define struct property '%s' multiple times in struct literal", propId)
			continue
		}
		value := prop.Children[1]
		if valueType, hasErr := evalType(value, property.Type); !property.Type.Equals(valueType) && !ImplementsInterface(property.Type, valueType) && !hasErr {
			messages.Complain(diagnostic.TypeError, value.Location, "Property type %s expected but found %s", property.Type, valueType)
		}
		visited.Append(propId)
	}
	return dt.NewDynamicType(name)
}

func evalType(ast *parser.AST, expectedType dt.SourceType) (dt.SourceType, bool) {
	hasError := false
	nodeType := dt.NoneType
	if ast.IsLiteral() {
		nodeType = evalLiteral(ast, expectedType)
	} else if ast.Token.Kind == lexer.ID {
		symbol := currentScope.LookupVariable(ast.Token.Value)
		if symbol != nil {
			nodeType = symbol.Type
		} else {
			messages.Complain(diagnostic.NameError, ast.Location, "Variable '%s' not defined in this scope", ast.Token.Value)
			hasError = true
		}
	} else if ast.Label == "Unary" {
		nodeType, hasError = evalUnary(ast.Children[0], ast.Children[1], expectedType)
	} else if ast.Label == "typecast" {
		nodeType, hasError = evalTypecast(ast.Children[0], ast.Children[1], expectedType)
	} else if ast.Label == "index" {
		nodeType, hasError = evalIndex(ast.Children[0], ast.Children[1], expectedType)
	} else if ast.Label == "slice" {
		nodeType, hasError = evalSlice(ast, expectedType)
	} else if ast.Label == "ARR-END" {
		nodeType, hasError = evalArrayEnd(ast.Children[0], expectedType)
	} else if ast.Label == "range" {
		nodeType, hasError = evalRange(ast, expectedType)
	} else if ast.Label == "call" {
		nodeType, hasError = handleFunctionCall(ast.Children)
	} else if ast.Token.Kind == lexer.OPERATOR_ADD {
		nodeType, hasError = evalAdd(ast.Children[0], ast.Children[1], ast.Token, expectedType)
	} else if ast.Token.Kind == lexer.OPERATOR_MULT {
		nodeType, hasError = evalMult(ast.Children[0], ast.Children[1], ast.Token, expectedType)
	} else if ast.Token.Kind == lexer.OPERATOR_BS || ast.Token.Kind == lexer.OPERATOR_BW {
		nodeType, hasError = evalBitOperation(ast.Children[0], ast.Children[1], ast.Token, expectedType)
	} else if ast.Token.Kind == lexer.OPERATOR_COMPARE {
		nodeType, hasError = evalCompare(ast.Children[0], ast.Children[1], ast.Token, expectedType)
	} else if ast.Token.Value == "&&" || ast.Token.Value == "||" {
		nodeType, hasError = evalLogicalOperation(ast.Children[0], ast.Children[1], ast.Token, expectedType)
	} else if ast.Token.Value == "**" {
		nodeType, hasError = evalExponent(ast.Children[0], ast.Children[1], expectedType)
	} else if ast.Label == "dot" {
		nodeType, hasError = handleDot(ast.Children[0], ast.Children[1], false, false, false)
	}
	if !hasError {
		ast.Type = nodeType
	}
	return nodeType, hasError
}

func comparableCheck(lhs dt.SourceType, rhs dt.SourceType) bool {
	return lhs.Equals(rhs) || (lhs.IsUnsignedType() && rhs.IsUnsignedType()) || (lhs.IsSignedType() && rhs.IsSignedType())
}

func handleBinaryNumberExpression(left *parser.AST, right *parser.AST, operator string, expectedType dt.SourceType) (dt.SourceType, error) {
	inferLeft := left.IsLiteralExpression() && !right.IsLiteralExpression()
	inferRight := !left.IsLiteralExpression() && right.IsLiteralExpression()
	var lhs, rhs dt.SourceType
	var lHasErr, rHasErr bool
	if inferLeft {
		rhs, rHasErr = evalType(right, expectedType)
		lhs, lHasErr = evalType(left, rhs)
	} else if inferRight {
		lhs, lHasErr = evalType(left, expectedType)
		rhs, rHasErr = evalType(right, lhs)
	} else {
		lhs, lHasErr = evalType(left, expectedType)
		rhs, rHasErr = evalType(right, expectedType)
	}
	if lHasErr || rHasErr {
		return dt.NoneType, nil
	}
	return decideNumberType(lhs, rhs, operator)
}

func decideNumberType(lhs dt.SourceType, rhs dt.SourceType, operator string) (dt.SourceType, error) {
	if lhs.Equals(rhs) {
		return lhs, nil
	}
	if (lhs.Equals(dt.Uint32Type) || lhs.Equals(dt.Uint64Type)) && (rhs.Equals(dt.Uint32Type) || rhs.Equals(dt.Uint64Type)) {
		return dt.Uint64Type, nil
	} else if (lhs.Equals(dt.Int32Type) || lhs.Equals(dt.Int64Type)) && (rhs.Equals(dt.Int32Type) || rhs.Equals(dt.Int64Type)) {
		return dt.Int64Type, nil
	} else if (lhs.Equals(dt.Int32Type) || lhs.Equals(dt.FloatType)) && (rhs.Equals(dt.Int32Type) || rhs.Equals(dt.FloatType)) {
		return dt.FloatType, nil
	} else if (lhs.Equals(dt.Int32Type) || lhs.Equals(dt.DoubleType)) && (rhs.Equals(dt.Int32Type) || rhs.Equals(dt.DoubleType)) {
		return dt.DoubleType, nil
	} else if (lhs.Equals(dt.Int64Type) || lhs.Equals(dt.DoubleType)) && (rhs.Equals(dt.Int64Type) || rhs.Equals(dt.DoubleType)) {
		return dt.DoubleType, nil
	} else if (lhs.Equals(dt.FloatType) || lhs.Equals(dt.DoubleType)) && (rhs.Equals(dt.FloatType) || rhs.Equals(dt.DoubleType)) {
		return dt.DoubleType, nil
	} else {
		return dt.NoneType, fmt.Errorf("Cannot use operator '%s' between %s and %s", operator, lhs, rhs)
	}
}

func nodeToType(node *parser.AST) dt.SourceType {
	if node.Token.Kind == lexer.ID {
		symbol := globalScope.LookupType(node.Token.Value)
		if symbol == nil || (symbol.GetSymbolType() != "interface" && symbol.GetSymbolType() != "struct") {
			messages.Complain(diagnostic.TypeError, node.Location, "Invalid type '%s' provided", node.Token.Value)
			return dt.NoneType
		}
		return dt.NewDynamicType(node.Token.Value)
	}
	switch node.Token.Value {
	case "int":
		return dt.Int32Type
	case "int64":
		return dt.Int64Type
	case "uint32":
		return dt.Uint32Type
	case "uint64":
		return dt.Uint64Type
	case "float":
		return dt.FloatType
	case "double":
		return dt.DoubleType
	case "char":
		return dt.CharType
	case "bool":
		return dt.BoolType
	case "String":
		return dt.StringType
	default:
		return dt.NoneType
	}
}

func isCompatibleType(expectedType dt.SourceType, actualType dt.SourceType) bool {
	return (ImplementsInterface(expectedType, actualType) ||
		expectedType.Equals(dt.Int64Type) && actualType.Equals(dt.Int32Type) ||
		expectedType.Equals(dt.Uint64Type) && actualType.Equals(dt.Uint32Type) ||
		expectedType.Equals(dt.DoubleType) && actualType.IsSignedType() ||
		expectedType.Equals(dt.FloatType) && actualType.Equals(dt.Int32Type))
}
