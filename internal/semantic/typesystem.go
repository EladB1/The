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

func evalType(ast *parser.AST, expectedType datatypes.DataType) datatypes.DataType {
	numericTypes := []datatypes.DataType{datatypes.Int32, datatypes.Int64, datatypes.Uint32, datatypes.Uint64, datatypes.Float, datatypes.Double}
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
	case lexer.KW_BOOLVALUE:
		return datatypes.Bool
	default:
		if ast.Label == "struct_literal" {
			name := ast.Children[0].Token.Value
			symbol := globalScope.lookupStruct(name)
			if symbol == nil {
				messages = messages.Complain(diagnostic.NameError, fmt.Sprintf("Struct %s not defined", name), ast.Location)
				break
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
		} else if ast.Label == "Unary" {

		} else if ast.Label == "typecast" {
			// TODO
		} else if ast.Label == "index" {

		} else if ast.Label == "slice" {

		} else if ast.Label == "ARR-END" {

		} else if ast.Label == "call" {
			return handleFunctionCall()
		}
	}

	return datatypes.None
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
