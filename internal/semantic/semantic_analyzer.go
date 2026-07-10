package semantic

import (
	"fmt"
	"slices"

	"github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/diagnostic"
	"github.com/EladB1/The/internal/parser"
)

var specialBlocks []string = []string{"private", "cast", "compare"}

var messages diagnostic.PhaseDiagnostics

func initScope() *Scope {
	globalScope := rootScope.addChild("@global", Default)
	return globalScope
}

/* moving scope pointer that starts at global scope */
var currentScope *Scope

/* global scope pointer that can be quickly referenced rather than going through full tree */
var globalScope *Scope

func setup() {
	messages = diagnostic.PhaseDiagnostics{}
	currentScope = initScope()
	globalScope = initScope()
}

func Analyze(ast parser.AST) (parser.AST, diagnostic.PhaseDiagnostics) {
	setup()
	collectTypeNames(ast)
	analyzeInterfaceFnSignatures()
	analyzeStructFnSignatures()
	collectFunctionSignatures(ast)
	analyzeInterfaceImplementation()
	analyzeGlobals(ast)
	analyzeInterfaceFnBodies()
	// TODO: Uncomment code below when ready
	// missingEntry := false
	// if fn := globalScope.lookupFunction("main"); fn != nil {
	// 	if _, found := fn.overloads[""]; !found || fn.returnType != datatypes.Int32 {
	// 		missingEntry = true
	// 	}
	// } else {
	// 	missingEntry = true
	// }
	// if missingEntry {
	// 	messages.ComplainPositionless(diagnostic.Error, "Missing entrypoint function 'fn main()->int'")
	// }
	fmt.Println(globalScope)
	messages.Sort()
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
			messages.Complain(diagnostic.NameError, nameNode.Location, "Name '%s' already in use", name)
			continue
		}
		if node.Label == "interface" {
			forbidden_names := []string{"cast", "compare"}
			if slices.Contains(forbidden_names, name) {
				messages.Complain(diagnostic.NameError, node.Location, "Cannot name interface '%s'", name)
				continue
			}
			childScope := globalScope.addChild(name, Interface)
			globalScope.interfaces[name] = InterfaceSymbol{
				name:       name,
				innerScope: childScope,
				Def:        &node,
			}
		} else if node.Token.Value == "struct" {
			childScope := globalScope.addChild(name, Struct)
			childScope.variables["this"] = VariableSymbol{
				name:        "this",
				Type:        datatypes.Ref,
				isPrivate:   true,
				isMutable:   false,
				Initialized: true,
				Def:         nil,
			}
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
			err := currentScope.functions.add(symbol)
			if err != nil {
				messages.Complain(diagnostic.IllegalStatementError, node.Location, "%v", err)
			}
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
					messages.Complain(diagnostic.NameError, node.Location, "Could not find interface name: '%s'", node.Token.Value)
				} else {
					impl = append(impl, node.Token.Value)
					str.implements = impl
					globalScope.structs[str.name] = str
				}
			}
		}
		for _, node := range body.Children {
			switch node.Label {
			case "fn":
				symbol := processFunctionSignature(node)
				if err := currentScope.functions.add(symbol); err != nil {
					messages.Complain(diagnostic.IllegalStatementError, node.Location, "%s", err.Error())
				}
			case "named-block":
				symbol := analyzeNamedBlock(node, str.name, impl)
				if symbol != nil {
					currentScope.namedBlocks[symbol.name] = *symbol
				}
			default:
				symbol := analyzeVariable(node)
				str.sizeInBytes += symbol.Type.GetSizeInBytes()
				globalScope.structs[str.name] = str
				if symbol != nil {
					currentScope.variables[symbol.name] = *symbol
				}
			}
		}
		str.implFnNames = map[string][]string{}
		for _, nb := range currentScope.namedBlocks {
			for _, fn := range nb.innerScope.functions {
				str.implFnNames[fn.name] = append(str.implFnNames[fn.name], nb.name)
			}
		}
		globalScope.structs[str.name] = str
	}
	currentScope = globalScope // reset the current scope
}

// Pass four
func collectFunctionSignatures(ast parser.AST) {
	for _, node := range ast.Children {
		if node.Label == "fn" {
			symbol := processFunctionSignature(node)
			if err := globalScope.functions.add(symbol); err != nil {
				messages.Complain(diagnostic.IllegalStatementError, node.Location, "%s", err.Error())
			}
		}
	}
}

// Pass five
func analyzeInterfaceImplementation() {
	for _, str := range globalScope.structs {
		if len(str.implements) == 0 { // no interface_list node
			continue
		}
		impl := []string{}
		for _, intfName := range str.implements {
			intf := globalScope.lookupInterface(intfName)
			if intf == nil {
				messages.Complain(diagnostic.NameError, str.Def.Location, "Could not find interface %s", intfName)
				continue
			}
			if slices.Contains(impl, intfName) {
				messages.Complain(diagnostic.ImplementationError, str.Def.Location, "struct cannot implement interface multiple times")
				continue
			}
			impl = append(impl, intfName)
			namedBlock := str.innerScope.lookupNamedBlock(intfName)
			if namedBlock == nil {
				messages.Complain(diagnostic.ImplementationError, str.Def.Location, "struct %s is missing named block for interface %s", str.name, intfName)
			} else {
				str.innerScope.variables[intfName] = VariableSymbol{
					name: intfName,
					Type: datatypes.ScopeRef{
						Scopes: []string{str.name, intfName},
					},
					isPrivate:   false,
					isMutable:   false,
					Def:         namedBlock.Def,
					Initialized: true,
				}
				for _, fn := range intf.innerScope.functions {
					missing := false
					returnStr := ""
					if fn.returnType != datatypes.None {
						returnStr = fmt.Sprintf("->%s", fn.returnType)
					}
					nb_fn := namedBlock.innerScope.lookupFunctionByName(fn.name)
					if nb_fn == nil {
						missing = true
						namedBlock.innerScope.functions[fn.name] = fn
						nb_fn = namedBlock.innerScope.lookupFunctionByName(fn.name)
					} else if nb_fn.returnType != fn.returnType {
						messages.Complain(diagnostic.ImplementationError, namedBlock.Def.Location, "Implementation function %s returns %s but interface %s returns %s", fn.name, nb_fn.returnType, intfName, fn.returnType)
						continue
					}
					for i, overload := range fn.overloads {
						params := datatypes.Join(overload.parameters)
						if missing {
							if overload.hasDefaultImplementation { // copy it over from the interface
								nb_fn.overloads[i].parameters = overload.parameters
								nb_fn.overloads[i].innerScope = namedBlock.innerScope.addChild(fmt.Sprintf("%s@%s", fn.name, namedBlock.innerScope.id), Function)
								nb_fn.overloads[i].innerScope.variables = overload.innerScope.variables
								namedBlock.innerScope.functions[fn.name] = *nb_fn
							} else {
								messages.Complain(diagnostic.ImplementationError, namedBlock.Def.Location, "Interface %s implementation missing 'fn %s(%s)%s'", intfName, fn.name, params, returnStr)
							}
						} else {
							match := nb_fn.getMatchingOverload(overload.parameters)
							if overload.hasDefaultImplementation {
								if match == nil {
									nb_fn.overloads[i].parameters = overload.parameters
									namedBlock.innerScope.functions[nb_fn.name] = *nb_fn
								}
							} else {
								if match == nil {
									messages.Complain(diagnostic.ImplementationError, namedBlock.Def.Location, "Interface %s implementation missing 'fn %s(%s)%s'", intfName, fn.name, params, returnStr)
								}
							}
						}
					}
				}
				for _, fn := range namedBlock.innerScope.functions {
					returnStr := ""
					if fn.returnType != datatypes.None {
						returnStr = fmt.Sprintf("->%s", fn.returnType)
					}
					intf_fn := intf.innerScope.lookupFunctionByName(fn.name)
					if intf_fn == nil {
						messages.Complain(diagnostic.ImplementationError, namedBlock.Def.Location, "Named block %s contains function %s which its interface does not", intfName, fn.name)
						continue
					}
					for _, overload := range fn.overloads {
						if match := intf_fn.getMatchingOverload(overload.parameters); match == nil {
							messages.Complain(diagnostic.ImplementationError, namedBlock.Def.Location, "Named block %s contains function %s(%s)%s which its interface does not", intfName, fn.name, datatypes.Join(overload.parameters), returnStr)
						}
					}
				}
			}
		}
	}
}

// Pass six
func analyzeGlobals(ast parser.AST) {
	for _, node := range ast.Children {
		if node.Label == "Variable" {
			symbol := analyzeVariable(node)
			if symbol == nil {
				continue
			}
			if symbol.isMutable {
				messages.Warn(node.Location, "Mutable global variable declared")
			}
			if symbol.isPrivate {
				messages.Complain(diagnostic.AccessError, node.Location, "Cannot use private modifier outside of a struct")
				continue
			}
			globalScope.variables[symbol.name] = *symbol
		}
	}
}

// Pass seven
func analyzeInterfaceFnBodies() {
	for _, intf := range globalScope.interfaces {
		for _, fn := range intf.innerScope.functions {
			analyzeFunctionBody(fn)
		}
	}
}

// Pass eight
func analyzeStructMethodBodies() {}

// Pass nine
func analyzeFunctionsBodies(ast parser.AST) {

}

// helpers
