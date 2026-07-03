package semantic

import (
	"fmt"

	"github.com/EladB1/The/internal/diagnostic"
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
	collectFunctionSignatures(ast)
	analyzeGlobals(ast)
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
		result := globalScope.lookupType(name)
		if result != nil {
			messages = messages.Complain(diagnostic.NameError, nameNode.Location, "Name '%s' already in use", name)
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
			currentScope.functions.add(symbol)
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
				if globalScope.lookupInterface(node.Token.Value) == nil {
					messages = messages.Complain(diagnostic.NameError, node.Location, "Could not find interface name: '%s'", node.Token.Value)
				} else {
					impl = append(impl, node.Token.Value)
				}
			}
		}
		for _, node := range body.Children {
			switch node.Label {
			case "fn":
				symbol := processFunctionSignature(node)
				if err := currentScope.functions.add(symbol); err != nil {
					messages = messages.Complain(diagnostic.IllegalStatementError, node.Location, "%s", err.Error())
				}
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
func collectFunctionSignatures(ast parser.AST) {
	for _, node := range ast.Children {
		if node.Label == "fn" {
			symbol := processFunctionSignature(node)
			if err := globalScope.functions.add(symbol); err != nil {
				messages = messages.Complain(diagnostic.IllegalStatementError, node.Location, "%s", err.Error())
			}
		}
	}
}

// Pass five
func analyzeGlobals(ast parser.AST) {
	for _, node := range ast.Children {
		if node.Label == "Variable" {
			symbol := analyzeVariable(node)
			if symbol.isMutable {
				messages = messages.Warn(node.Location, "Mutable global variable declared")
			}
			if symbol.isPrivate {
				messages = messages.Complain(diagnostic.AccessError, node.Location, "Cannot use private modifier outside of a struct")
				continue
			}
			if symbol != nil {
				globalScope.variables[symbol.name] = *symbol
			}
		}
	}
}

// Pass six
func analyzeInterfaceFnBodies() {
	for _, intf := range globalScope.interfaces {
		for _, fn := range intf.innerScope.functions {
			// if !fn.hasDefaultImplementation {
			// 	continue
			// }
			analyzeFunctionBody(fn)
		}
	}
}

// Pass seven
func analyzeStructMethodBodies() {}

// Pass eight
func analyzeInterfaceImplementation() {}

// Pass nine
func analyzeFunctionsBodies(ast parser.AST) {

}

// helpers
