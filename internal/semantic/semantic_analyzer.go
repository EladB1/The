package semantic

import (
	"fmt"

	"github.com/EladB1/The/internal/diagnostic"
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
				name:       name.GetValueString(),
				innerScope: &childScope,
				bodyStart:  &ast.Children[1],
			}
		} else if node.Token.Value == "struct" {
			globalScope.structs[name.GetValueString()] = StructSymbol{
				name:       name.GetValueString(),
				innerScope: &childScope,
				bodyStart:  &ast.Children[len(ast.Children)-1],
			}
		}
	}
}

// Pass two
func analyzeInterfaces(ast parser.AST) {
	for _, node := range ast.Children {
		if node.Label == "interface" {

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
