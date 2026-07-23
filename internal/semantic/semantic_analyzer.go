package semantic

import (
	"fmt"
	"slices"

	dt "github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/diagnostic"
	"github.com/EladB1/The/internal/parser"
)

var specialBlocks []string = []string{"private", "cast", "compare"}

var messages diagnostic.PhaseDiagnostics

func initScope() *Scope {
	rootScope.Children = nil
	globalScope := rootScope.addChild("@global", Default)
	return globalScope
}

/* moving scope pointer that starts at global scope */
var currentScope *Scope

/* global scope pointer that can be quickly referenced rather than going through full tree */
var globalScope *Scope

func setup() {
	messages = diagnostic.PhaseDiagnostics{}
	globalScope = initScope()
	currentScope = globalScope
}

func Analyze(ast *parser.AST) (*Scope, diagnostic.PhaseDiagnostics) {
	setup()
	collectTypeNames(ast)
	analyzeInterfaceFnSignatures()
	analyzeStructFnSignatures()
	collectFunctionSignatures(ast)
	analyzeInterfaceImplementation()
	analyzeGlobals(ast)
	analyzeInterfaceFnBodies()
	analyzeStructMethodBodies()
	analyzeFunctionsBodies()
	missingEntry := true
	if fn := globalScope.LookupFunctionByName("main"); fn != nil {
		for _, overload := range fn.Overloads {
			if len(overload.Parameters) == 0 && fn.ReturnType.Equals(dt.Int32Type) {
				missingEntry = false
				break
			}
		}
	}
	if missingEntry {
		messages.ComplainPositionless(diagnostic.Error, "Missing entrypoint function 'fn main()->int'")
	}
	messages.Sort()
	return rootScope, messages
}

// Pass one
func collectTypeNames(ast *parser.AST) {
	for _, node := range ast.Children {
		if node.Label != "interface" && node.Token.Value != "struct" {
			continue
		}
		nameNode := node.Children[0]
		name := nameNode.Token.Value
		result := globalScope.LookupType(name)
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
			globalScope.Interfaces[name] = InterfaceSymbol{
				name:       name,
				innerScope: childScope,
				Def:        node,
			}
		} else if node.Token.Value == "struct" {
			childScope := globalScope.addChild(name, Struct)
			childScope.Variables["this"] = VariableSymbol{
				Name:        "this",
				Type:        dt.NewContainerType(dt.Ref, dt.NewReferenceSubType(childScope.Id)),
				isPrivate:   true,
				isMutable:   true,
				Initialized: true,
				Def:         nil,
			}
			childScope.Variables["global"] = VariableSymbol{
				Name:        "global",
				Type:        dt.GlobalRefType,
				isPrivate:   true,
				isMutable:   true,
				Initialized: true,
				Def:         nil,
			}
			globalScope.Structs[name] = StructSymbol{
				Name:       name,
				InnerScope: childScope,
				Def:        node,
			}
		}
	}
}

// Pass two
func analyzeInterfaceFnSignatures() {
	for _, intf := range globalScope.Interfaces {
		currentScope = intf.innerScope
		for _, node := range intf.Def.Children[1].Children {
			symbol := processFunctionSignature(node)
			err := currentScope.Functions.add(symbol)
			if err != nil {
				messages.Complain(diagnostic.IllegalStatementError, node.Location, "%v", err)
			}
		}
	}
	currentScope = globalScope // reset the current scope
}

// Pass three
func analyzeStructFnSignatures() {
	for _, str := range globalScope.Structs {
		currentScope = str.InnerScope
		// collect impl values
		def := str.Def.Children
		body := def[1]
		impl := []string{}
		if def[1].Label == "interface_list" {
			body = def[2]
			for _, node := range def[1].Children {
				if globalScope.LookupInterface(node.Token.Value) == nil {
					messages.Complain(diagnostic.NameError, node.Location, "Could not find interface name: '%s'", node.Token.Value)
				} else {
					impl = append(impl, node.Token.Value)
					str.Implements = impl
					globalScope.Structs[str.Name] = str
				}
			}
		}
		for _, node := range body.Children {
			switch node.Label {
			case "fn":
				symbol := processFunctionSignature(node)
				if err := currentScope.Functions.add(symbol); err != nil {
					messages.Complain(diagnostic.IllegalStatementError, node.Location, "%s", err.Error())
				}
			case "named-block":
				symbol := analyzeNamedBlock(node, str.Name, impl)
				if symbol != nil {
					currentScope.NamedBlocks[symbol.Name] = *symbol
				}
			default:
				symbol := analyzeVariable(node)
				str.SizeInBytes += symbol.Type.GetSizeInBytes()
				globalScope.Structs[str.Name] = str
				if symbol != nil {
					currentScope.Variables[symbol.Name] = *symbol
				}
			}
		}
		str.implFnNames = map[string][]string{}
		for _, nb := range currentScope.NamedBlocks {
			for _, fn := range nb.InnerScope.Functions {
				str.implFnNames[fn.Name] = append(str.implFnNames[fn.Name], nb.Name)
			}
		}
		globalScope.Structs[str.Name] = str
	}
	currentScope = globalScope // reset the current scope
}

// Pass four
func collectFunctionSignatures(ast *parser.AST) {
	for _, node := range ast.Children {
		if node.Label == "fn" {
			symbol := processFunctionSignature(node)
			if err := globalScope.Functions.add(symbol); err != nil {
				messages.Complain(diagnostic.IllegalStatementError, node.Location, "%s", err.Error())
			}
		}
	}
}

// Pass five
func analyzeInterfaceImplementation() {
	for _, str := range globalScope.Structs {
		if len(str.Implements) == 0 { // no interface_list node
			continue
		}
		impl := []string{}
		for _, intfName := range str.Implements {
			intf := globalScope.LookupInterface(intfName)
			if intf == nil {
				messages.Complain(diagnostic.NameError, str.Def.Location, "Could not find interface %s", intfName)
				continue
			}
			if slices.Contains(impl, intfName) {
				messages.Complain(diagnostic.ImplementationError, str.Def.Location, "struct cannot implement interface multiple times")
				continue
			}
			impl = append(impl, intfName)
			namedBlock := str.InnerScope.LookupNamedBlock(intfName)
			if namedBlock == nil {
				messages.Complain(diagnostic.ImplementationError, str.Def.Location, "struct %s is missing named block for interface %s", str.Name, intfName)
			} else {
				str.InnerScope.Variables[intfName] = VariableSymbol{
					Name:        intfName,
					Type:        dt.NewContainerType(dt.ScopeRef, dt.NewReferenceSubType(str.Name), dt.NewReferenceSubType(intfName)),
					isPrivate:   false,
					isMutable:   false,
					Def:         namedBlock.Def,
					Initialized: true,
				}
				for _, fn := range intf.innerScope.Functions {
					missing := false
					returnStr := ""
					if !fn.ReturnType.Equals(dt.NoneType) {
						returnStr = fmt.Sprintf("->%s", fn.ReturnType)
					}
					nb_fn := namedBlock.InnerScope.LookupFunctionByName(fn.Name)
					if nb_fn == nil {
						missing = true
						namedBlock.InnerScope.Functions[fn.Name] = fn
						nb_fn = namedBlock.InnerScope.LookupFunctionByName(fn.Name)
					} else if !nb_fn.ReturnType.Equals(fn.ReturnType) {
						messages.Complain(diagnostic.ImplementationError, namedBlock.Def.Location, "Implementation function %s returns %s but interface %s returns %s", fn.Name, nb_fn.ReturnType, intfName, fn.ReturnType)
						continue
					}
					for i, overload := range fn.Overloads {
						params := dt.JoinTypes(overload.Parameters)
						if missing {
							str.UpdateImplFnNames(fn.Name, intfName)
							if overload.HasDefaultImplementation { // copy it over from the interface
								nb_fn.Overloads[i].Parameters = overload.Parameters
								nb_fn.Overloads[i].InnerScope = namedBlock.InnerScope.addChild(fmt.Sprintf("%s@%s", fn.Name, namedBlock.InnerScope.Id), Function)
								nb_fn.Overloads[i].InnerScope.Variables = overload.InnerScope.Variables
								namedBlock.InnerScope.Functions[fn.Name] = *nb_fn
							} else {
								messages.Complain(diagnostic.ImplementationError, namedBlock.Def.Location, "Interface %s implementation missing 'fn %s(%s)%s'", intfName, fn.Name, params, returnStr)
							}
						} else {
							match := nb_fn.GetMatchingOverload(overload.Parameters)
							if overload.HasDefaultImplementation {
								if match == nil {
									nb_fn.Overloads[i].Parameters = overload.Parameters
									namedBlock.InnerScope.Functions[nb_fn.Name] = *nb_fn
								}
							} else {
								if match == nil {
									messages.Complain(diagnostic.ImplementationError, namedBlock.Def.Location, "Interface %s implementation missing 'fn %s(%s)%s'", intfName, fn.Name, params, returnStr)
								}
							}
						}
					}
				}
				for _, fn := range namedBlock.InnerScope.Functions {
					returnStr := ""
					if !fn.ReturnType.Equals(dt.NoneType) {
						returnStr = fmt.Sprintf("->%s", fn.ReturnType)
					}
					intf_fn := intf.innerScope.LookupFunctionByName(fn.Name)
					if intf_fn == nil {
						messages.Complain(diagnostic.ImplementationError, namedBlock.Def.Location, "Named block %s contains function %s which its interface does not", intfName, fn.Name)
						continue
					}
					for _, overload := range fn.Overloads {
						if match := intf_fn.GetMatchingOverload(overload.Parameters); match == nil {
							messages.Complain(diagnostic.ImplementationError, namedBlock.Def.Location, "Named block %s contains function %s(%s)%s which its interface does not", intfName, fn.Name, dt.JoinTypes(overload.Parameters), returnStr)
						}
					}
				}
			}
		}
	}
}

// Pass six
func analyzeGlobals(ast *parser.AST) {
	for i, node := range ast.Children {
		if node.Label == "Variable" {
			symbol := analyzeVariable(node)
			ast.Children[i] = node
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
			symbol.Ctx = Global
			globalScope.Variables[symbol.Name] = *symbol
		}
	}
}

// Pass seven
func analyzeInterfaceFnBodies() {
	for _, intf := range globalScope.Interfaces {
		for _, fn := range intf.innerScope.Functions {
			analyzeFunctionBody(fn)
		}
	}
}

// Pass eight
func analyzeStructMethodBodies() {
	for _, str := range globalScope.Structs {
		for _, fn := range str.InnerScope.Functions {
			analyzeFunctionBody(fn)
		}
		for _, nb := range str.InnerScope.NamedBlocks {
			for _, fn := range nb.InnerScope.Functions {
				analyzeFunctionBody(fn)
			}
		}
	}
}

// Pass nine
func analyzeFunctionsBodies() {
	for _, fn := range globalScope.Functions {
		analyzeFunctionBody(fn)
	}
}
