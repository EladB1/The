package semantic

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/diagnostic"
	"github.com/EladB1/The/internal/lexer"
	"github.com/EladB1/The/internal/parser"
)

var messages diagnostic.PhaseDiagnostics = diagnostic.PhaseDiagnostics{}

func initScope() *Scope {
	builtinScope.addChild("@global")
	return builtinScope.children[0]
}

/* moving scope pointer that starts at global scope */
var currentScope *Scope = initScope()

/* global scope pointer that can be quickly referenced rather than going through full tree */
var globalScope *Scope = initScope()

func Analyze(ast parser.AST) (parser.AST, diagnostic.PhaseDiagnostics) {
	collectTypeNames(ast)
	analyzeInterfaces()
	fmt.Println(globalScope)
	return ast, messages
}

// Pass one
func collectTypeNames(ast parser.AST) {
	for _, node := range ast.Children {
		if node.Label != "interface" && node.Token.Value != "struct" {
			continue
		}
		name := node.Children[0].Token
		result := globalScope.lookup(name.GetValueString())
		if result != nil && (result.getSymbolType() == "interface" || result.getSymbolType() == "struct") {
			messages = messages.Complain(diagnostic.NameError, fmt.Sprintf("Name '%s' already in use", name.GetValueString()), name.Line, name.Column)
			continue
		}
		childScope := globalScope.addChild(name.GetValueString())
		if node.Label == "interface" {
			globalScope.interfaces[name.GetValueString()] = InterfaceSymbol{
				name:         name.GetValueString(),
				innerScope:   &childScope,
				interfaceDef: &node,
			}
		} else if node.Token.Value == "struct" {
			globalScope.structs[name.GetValueString()] = StructSymbol{
				name:       name.GetValueString(),
				innerScope: &childScope,
				structDef:  &node,
			}
		}
	}
}

// Pass two
func analyzeInterfaces() {
	for _, intf := range globalScope.interfaces {
		for _, node := range intf.interfaceDef.Children[1].Children {
			//fmt.Println(node)
			name := node.Children[0].Token.GetValueString()
			paramtypes, paramNames := getParams(node)
			returnType := getReturnType(node)
			signature := formFunctionSignature(name, paramtypes, returnType)
			childScope := intf.innerScope.addChild(name)
			if !reflect.DeepEqual(paramtypes, []datatypes.DataType{datatypes.None}) {
				for i := range len(paramtypes) {
					childScope.variables[paramNames[i]] = VariableSymbol{
						name: paramNames[i],
						Type: paramtypes[i],
					}
				}
			}
			intf.innerScope.functions[signature] = FunctionSymbol{
				name:                     name,
				parameters:               paramtypes,
				returnType:               returnType,
				hasDefaultImplementation: functionHasBody(node),
				funcDef:                  &node,
				innerScope:               &childScope,
			}
			fmt.Println(signature)
			// intf.innerScope.functions[name] = FunctionSymbol{
			// 	name: name,
			// }
		}
	}
}

// Pass three
func analyzeStructs(ast parser.AST) {

}

// Pass four
func collectFunctionSignatures(ast parser.AST) {

}

// Pass five
func analyzeFunctionsAndGlobalVariables(ast parser.AST) {

}

// helpers

func formFunctionSignature(name string, params []datatypes.DataType, returns datatypes.DataType) string {
	// TODO: add function parent(s)
	returnStr := ""
	if returns != datatypes.None {
		returnStr = fmt.Sprintf(" -> %s", returns.String())
	}
	paramStr := strings.Builder{}
	for i, param := range params {
		paramStr.WriteString(param.String())
		if i != len(params)-1 {
			paramStr.WriteRune(',')
		}
	}
	return fmt.Sprintf("fn %s(%s)%s", name, paramStr.String(), returnStr)
}

func getReturnType(fnNode parser.AST) datatypes.DataType {
	if len(fnNode.Children) == 1 {
		return datatypes.None
	}
	var typeIndex int
	if functionHasBody(fnNode) && len(fnNode.Children) > 2 {
		//fmt.Println(fnNode)
		typeIndex = len(fnNode.Children) - 2
	} else {
		typeIndex = len(fnNode.Children) - 1
	}
	typeNode := fnNode.Children[typeIndex]

	if typeNode.Label != "params" && (typeNode.Token.Kind == lexer.KW_TYPE || typeNode.Token.Kind == lexer.ID) {
		return nodeToType(typeNode)
	}
	return datatypes.None
}

func functionHasBody(fnNode parser.AST) bool {
	return fnNode.Children[len(fnNode.Children)-1].Label == "fn-body"
}

func getParams(fnNode parser.AST) ([]datatypes.DataType, []string) {
	types := []datatypes.DataType{}
	names := []string{}
	var paramsIndex int
	if functionHasBody(fnNode) {
		if getReturnType(fnNode) == datatypes.None {
			paramsIndex = len(fnNode.Children) - 2
		} else {
			paramsIndex = len(fnNode.Children) - 3
		}
	} else {
		paramsIndex = len(fnNode.Children) - 1
	}
	if fnNode.Children[paramsIndex].Label != "params" {
		types = append(types, datatypes.None)
	} else {
		for _, param := range fnNode.Children[paramsIndex].Children {
			types = append(types, nodeToType(param.Children[0]))
			names = append(names, param.Children[1].Token.Value)
		}
	}
	return types, names
}

func nodeToType(node parser.AST) datatypes.DataType {
	if node.Token.Kind == lexer.ID {
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
