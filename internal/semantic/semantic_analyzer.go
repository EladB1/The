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
		if node.Label == "interface" {
			name := node.Children[0].Token
			if globalScope.lookup(name.GetValueString()) != nil {
				messages = messages.Complain(diagnostic.NameError, fmt.Sprintf("Name '%s' already in use", name.GetValueString()), name.Line, name.Column)
			} else {
				globalScope.interfaces[name.GetValueString()] = InterfaceSymbol{
					name: name.GetValueString(),
				}
			}
		} else if node.Token.Value == "struct" {
			name := node.Children[0].Token
			if globalScope.lookup(name.GetValueString()) != nil {
				messages = messages.Complain(diagnostic.NameError, fmt.Sprintf("Name '%s' already in use", name.GetValueString()), name.Line, name.Column)
			} else {
				globalScope.structs[name.GetValueString()] = StructSymbol{
					name: name.GetValueString(),
				}
			}
		}
	}
}

// Pass two
func analyzeInterfaces(ast parser.AST) {

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
