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

var numericTypes []datatypes.DataType = []datatypes.DataType{datatypes.Int32, datatypes.Int64, datatypes.Uint32, datatypes.Uint64, datatypes.Float, datatypes.Double}

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
		//rightTok := ast.Children[1].Token
		if leftTok := ast.Children[0].Token; leftTok.Kind == lexer.OPERATOR_UNARY || leftTok.Value == "-" {
			// left unary
		} else {
			// right unary
		}

	} else if ast.Label == "typecast" {
		// TODO
	} else if ast.Label == "index" {

	} else if ast.Label == "slice" {

	} else if ast.Label == "ARR-END" {

	} else if ast.Label == "call" {
		nodeType = handleFunctionCall()
	} else if ast.Token.Kind == lexer.OPERATOR_ADD {
		lhs := evalType(&ast.Children[0], datatypes.None)
		rhs := evalType(&ast.Children[1], datatypes.None)
		fmt.Println(lhs, rhs)
		// TODO: check compatibility
		if (lhs == datatypes.String || lhs == datatypes.Char) && (rhs == datatypes.String || rhs == datatypes.Char) {
			nodeType = datatypes.String
		} else if slices.Contains(numericTypes, lhs) && slices.Contains(numericTypes, rhs) {
			if lhs == rhs {
				nodeType = lhs
			} else {
				if (lhs == datatypes.Uint32 || lhs == datatypes.Uint64) && (rhs == datatypes.Uint32 || rhs == datatypes.Uint64) {
					nodeType = datatypes.Uint64
				} else if (lhs == datatypes.Int32 || lhs == datatypes.Int64) && (rhs == datatypes.Int32 || rhs == datatypes.Int64) {
					nodeType = datatypes.Int64
				} else if (lhs == datatypes.Int32 || lhs == datatypes.Float) && (rhs == datatypes.Int32 || rhs == datatypes.Float) {
					nodeType = datatypes.Float
				} else if (lhs == datatypes.Int32 || lhs == datatypes.Double) && (rhs == datatypes.Int32 || rhs == datatypes.Double) {
					nodeType = datatypes.Double
				} else if (lhs == datatypes.Int64 || lhs == datatypes.Double) && (rhs == datatypes.Int64 || rhs == datatypes.Double) {
					nodeType = datatypes.Double
				} else if (lhs == datatypes.Float || lhs == datatypes.Double) && (rhs == datatypes.Float || rhs == datatypes.Double) {
					nodeType = datatypes.Double
				} else {
					messages = messages.Complain(diagnostic.TypeError, fmt.Sprintf("Cannot use operator '%s' between %s and %s", ast.Token.Value, lhs, rhs), ast.Location)
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
		}
		// TODO: determine type of result
	} else if ast.Token.Kind == lexer.OPERATOR_BS {
		lhs := evalType(&ast.Children[0], datatypes.None)
		rhs := evalType(&ast.Children[1], datatypes.None)
		fmt.Println(lhs, rhs)
		// TODO: check compatibility
	} else if ast.Token.Kind == lexer.OPERATOR_BW {
		lhs := evalType(&ast.Children[0], datatypes.None)
		rhs := evalType(&ast.Children[1], datatypes.None)
		fmt.Println(lhs, rhs)
		// TODO: check compatibility
	} else if ast.Token.Kind == lexer.OPERATOR_COMPARE {
		lhs := evalType(&ast.Children[0], datatypes.None)
		rhs := evalType(&ast.Children[1], datatypes.None)
		fmt.Println(lhs, rhs)
		// TODO: check compatibility
		nodeType = datatypes.Bool
	} else if ast.Token.Kind == lexer.OPERATOR {
		lhs := evalType(&ast.Children[0], datatypes.None)
		rhs := evalType(&ast.Children[1], datatypes.None)
		fmt.Println(lhs, rhs)
		// TODO: check compatibility
	}
	ast.Type = nodeType
	return nodeType
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
