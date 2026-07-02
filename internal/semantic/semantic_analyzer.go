package semantic

import (
	"fmt"
	"slices"
	"strings"

	"github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/diagnostic"
	"github.com/EladB1/The/internal/lexer"
	"github.com/EladB1/The/internal/parser"
)

var specialBlocks []string = []string{"private", "cast", "compare"}

var messages diagnostic.PhaseDiagnostics = diagnostic.PhaseDiagnostics{}

func initScope() *Scope {
	globalScope := rootScope.addChild("@global")
	return globalScope
}

/* moving scope pointer that starts at global scope */
var currentScope *Scope = initScope()

/* global scope pointer that can be quickly referenced rather than going through full tree */
var globalScope *Scope = initScope()

func Analyze(ast parser.AST) (parser.AST, diagnostic.PhaseDiagnostics) {
	collectTypeNames(ast)
	analyzeInterfaceFnSignatures()
	analyzeStructFnSignatures()
	fmt.Println(globalScope)
	return ast, messages
}

// Pass one
func collectTypeNames(ast parser.AST) {
	for _, node := range ast.Children {
		if node.Label != "interface" && node.Token.Value != "struct" {
			continue
		}
		nameNode := node.Children[0]
		name := nameNode.Token.Value
		result := globalScope.lookup(name, TYPE)
		if result != nil {
			messages = messages.Complain(diagnostic.NameError, fmt.Sprintf("Name '%s' already in use", name), nameNode.Location)
			continue
		}
		childScope := globalScope.addChild(name)
		if node.Label == "interface" {
			globalScope.interfaces[name] = InterfaceSymbol{
				name:       name,
				innerScope: childScope,
				Def:        &node,
			}
		} else if node.Token.Value == "struct" {
			globalScope.structs[name] = StructSymbol{
				name:       name,
				innerScope: childScope,
				Def:        &node,
			}
		}
	}
}

// Pass two
func analyzeInterfaceFnSignatures() {
	for _, intf := range globalScope.interfaces {
		currentScope = intf.innerScope
		for _, node := range intf.Def.Children[1].Children {
			symbol := processFunctionSignature(node)
			currentScope.functions[symbol.getSignature()] = symbol
		}
	}
	currentScope = globalScope // reset the current scope
}

// Pass three
func analyzeStructFnSignatures() {
	for _, str := range globalScope.structs {
		currentScope = str.innerScope
		// collect impl values
		def := str.Def.Children
		body := def[1]
		impl := []string{}
		if def[1].Label == "interface_list" {
			body = def[2]
			for _, node := range def[1].Children {
				if globalScope.lookup(node.Token.Value, Interface) == nil {
					messages = messages.Complain(diagnostic.NameError, fmt.Sprintf("Could not find interface name: '%s'", node.Token.Value), node.Location)
				} else {
					impl = append(impl, node.Token.Value)
				}
			}
		}
		for _, node := range body.Children {
			switch node.Label {
			case "fn":
				symbol := processFunctionSignature(node)
				currentScope.functions[symbol.getSignature()] = symbol
			case "named-block":
				symbol := analyzeNamedBlock(node, str.name, impl)
				if symbol != nil {
					currentScope.namedBlocks[symbol.name] = *symbol
				}
			default:
				symbol := analyzeVariable(node)
				if symbol != nil {
					currentScope.variables[symbol.name] = *symbol
				}
			}
		}
	}
	currentScope = globalScope // reset the current scope
}

// Pass four
func analyzeInterfaceFnBodies() {}

// Pass five
func analyzeStructMethodBodies() {}

// Pass six
func analyzeInterfaceImplementation() {}

// Pass seven
func collectFunctionSignatures(ast parser.AST) {

}

// Pass eight
func analyzeFunctionsAndGlobalVariables(ast parser.AST) {

}

// helpers

func processFunctionSignature(fnNode parser.AST) FunctionSymbol {
	details := fnNode.Children
	length := len(details)
	name := details[0].Token.Value
	var paramNode *parser.AST = nil
	var returnTypeNode *parser.AST = nil
	var bodyNode *parser.AST = nil
	var newScope *Scope = nil
	switch length {
	case 2:
		if details[1].Label == "params" { // fn name(type pname);
			paramNode = &details[1]
		} else if details[1].Token.Kind == lexer.ID || details[1].Token.Kind == lexer.KW_TYPE { // fn name() -> type;
			returnTypeNode = &details[1]
		} else { // fn name() {}
			bodyNode = &details[1]
		}
	case 3:
		if details[1].Token.Kind == lexer.ID || details[1].Token.Kind == lexer.KW_TYPE { // fn name() -> type {}
			returnTypeNode = &details[1]
			bodyNode = &details[2]
		} else {
			paramNode = &details[1]
			if details[2].Token.Kind == lexer.ID || details[2].Token.Kind == lexer.KW_TYPE { // fn name(type pname) -> type;
				returnTypeNode = &details[2]
			} else { // fn name(type pname) {}
				bodyNode = &details[2]
			}

		}
	case 4: // fn name(type pname) -> type {}
		paramNode = &details[1]
		returnTypeNode = &details[2]
		bodyNode = &details[3]
	}
	var paramNames []string
	var paramTypes []datatypes.DataType
	var returnType datatypes.DataType = datatypes.None
	if returnTypeNode != nil {
		returnType = nodeToType(*returnTypeNode)
	}
	if paramNode != nil {
		for _, param := range paramNode.Children {
			paramTypes = append(paramTypes, nodeToType(param.Children[0]))
			paramNames = append(paramNames, param.Children[1].Token.Value)
		}
	}
	if bodyNode != nil {
		newScope = currentScope.addChild(fmt.Sprintf("fn@%s", name))
		for i := range len(paramNames) {
			newScope.variables[paramNames[i]] = VariableSymbol{
				name: paramNames[i],
				Type: paramTypes[i],
			}
		}
	}
	symbol := FunctionSymbol{
		name:                     name,
		parameters:               paramTypes,
		returnType:               returnType,
		hasDefaultImplementation: bodyNode == nil,
		Def:                      &fnNode,
		innerScope:               newScope,
	}
	return symbol
}

func analyzeFunctionBody(name string, returns datatypes.DataType, body *parser.AST) {
	// TODO
}

func analyzeNamedBlock(nbNode parser.AST, structName string, impl []string) *NamedBlockSymbol {
	details := nbNode.Children
	name := details[0].Token.Value
	if !slices.Contains(specialBlocks, name) && !slices.Contains(impl, name) {
		messages = messages.Complain(diagnostic.NameError, fmt.Sprintf("Block '%s' not supported", name), nbNode.Location)
		return nil
	}
	body := details[1].Children
	scope := currentScope
	newScope := currentScope.addChild(fmt.Sprintf("%s@%s", name, currentScope.id))
	currentScope = newScope
	for _, node := range body {
		switch node.Label {
		case "fn":
			symbol := processFunctionSignature(node)
			switch name {
			case "compare":
				supported := []string{
					fmt.Sprintf("fn equals(%s)->bool", structName),
					fmt.Sprintf("fn lessThan(%s)->bool", structName),
					fmt.Sprintf("fn greaterThan(%s)->bool", structName),
				}
				if !slices.Contains(supported, symbol.getSignature()) {
					messages = messages.Complain(diagnostic.NamedBlockError, fmt.Sprintf("Function signature '%s' not supported; only '%s' supported", symbol.getSignature(), strings.Join(supported, ",")), node.Location)
				}
			case "cast":
				if len(symbol.parameters) > 0 || symbol.returnType == datatypes.None || symbol.returnType == datatypes.DynamicType(structName) {
					messages = messages.Complain(diagnostic.NamedBlockError, "Functions in cast block must take no parameters and return a different type", node.Location)
				}
			case "private":
				symbol.isPrivate = true
			}
			currentScope.functions[symbol.getSignature()] = symbol
		case "Variable":
			if name != "private" {
				messages = messages.Complain(diagnostic.IllegalStatementError, "Variable declaration only allowed in struct or private block", node.Location)
			} else {
				symbol := analyzeVariable(node)
				if symbol.isPrivate {
					messages = messages.Complain(diagnostic.Warning, "Redundant use of private in private block", node.Location)
				}
				currentScope.variables[symbol.name] = *symbol

			}
		}
	}
	currentScope = scope
	return &NamedBlockSymbol{
		name:           name,
		isSpecialBlock: slices.Contains(specialBlocks, name),
		Def:            &nbNode,
		innerScope:     newScope,
	}
}

func analyzeVariable(varNode parser.AST) *VariableSymbol {
	details := varNode.Children
	typeNode := details[0]
	name := details[1].Token
	isPrivate := false
	isMutable := false
	if details[0].Label == "modifiers" {
		typeNode = details[1]
		name = details[2].Token
		for _, modifier := range details[0].Children {
			if modifier.Token.Value == "private" {
				isPrivate = true
			}
			if modifier.Token.Value == "mut" {
				isMutable = true
			}
		}
	}
	varType := nodeToType(typeNode)
	if currentScope.lookup(name.Value, Variable) != nil {
		messages = messages.Complain(diagnostic.NameError, fmt.Sprintf("Name: '%s' already defined", name.Value), name.Location)
		return nil
	}
	return &VariableSymbol{
		name:      name.Value,
		Type:      varType,
		isPrivate: isPrivate,
		isMutable: isMutable,
		Def:       &varNode,
	}
}

func nodeToType(node parser.AST) datatypes.DataType {
	if node.Token.Kind == lexer.ID {
		symbol := globalScope.lookup(node.Token.Value, TYPE)
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
