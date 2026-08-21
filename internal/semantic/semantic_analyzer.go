package semantic

import (
	"fmt"
	"log"
	"slices"
	"strings"

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
		if strings.HasPrefix(name, "__") {
			messages.Complain(diagnostic.NameError, nameNode.Location, "Name '%s' uses reserved prefix '__'", name)
		}
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
			childScope.Variables.Add(VariableSymbol{
				Name:        "global",
				Type:        dt.GlobalRefType,
				isPrivate:   true,
				isMutable:   true,
				Initialized: true,
				Def:         nil,
			}, "global")
			globalScope.Interfaces.Add(InterfaceSymbol{
				name:       name,
				innerScope: childScope,
				Def:        node,
			}, name)
		} else if node.Token.Value == "struct" {
			childScope := globalScope.addChild(name, Struct)
			childScope.Variables.Add(VariableSymbol{
				Name:        "this",
				Type:        dt.NewContainerType(dt.Ref, dt.NewReferenceSubType(childScope.Id)),
				isPrivate:   true,
				isMutable:   true,
				Initialized: true,
				Def:         nil,
			}, "this")
			childScope.Variables.Add(VariableSymbol{
				Name:        "global",
				Type:        dt.GlobalRefType,
				isPrivate:   true,
				isMutable:   true,
				Initialized: true,
				Def:         nil,
			}, "global")
			globalScope.Structs.Add(StructSymbol{
				Name:       name,
				InnerScope: childScope,
				Def:        node,
			}, name)
		}
	}
}

// Pass two
func analyzeInterfaceFnSignatures() {
	interfaces := globalScope.Interfaces
	for i := range globalScope.Interfaces.OrderedNames {
		intf := interfaces.GetByIndex(i)
		if intf == nil {
			continue
		}
		currentScope = intf.innerScope
		for _, node := range intf.Def.Children[1].Children {
			symbol := processFunctionSignature(node)
			overload, err := create(currentScope.Functions, symbol)
			if err != nil {
				messages.Complain(diagnostic.IllegalStatementError, node.Location, "%v", err)
			} else {
				node.IRName = overload.IRName
			}
		}
	}
	currentScope = globalScope // reset the current scope
}

// Pass three
func analyzeStructFnSignatures() {
	structs := globalScope.Structs
	for i := range structs.OrderedNames {
		str := structs.GetByIndex(i)
		if str == nil {
			continue
		}
		currentScope = str.InnerScope
		// collect impl values
		def := str.Def.Children
		body := def[1]
		impl := []string{}
		var offset uint32 = 0
		if def[1].Label == "interface_list" {
			body = def[2]
			for _, node := range def[1].Children {
				if globalScope.LookupInterface(node.Token.Value) == nil {
					messages.Complain(diagnostic.NameError, node.Location, "Could not find interface name: '%s'", node.Token.Value)
				} else {
					impl = append(impl, node.Token.Value)
					str.Implements = impl
					globalScope.Structs.update(*str, str.Name)
				}
			}
		}
		for _, node := range body.Children {
			switch node.Label {
			case "fn":
				symbol := processFunctionSignature(node)
				if overload, err := create(currentScope.Functions, symbol); err != nil {
					messages.Complain(diagnostic.IllegalStatementError, node.Location, "%s", err.Error())
				} else {
					node.IRName = overload.IRName
				}
			case "named-block":
				symbol, extraSize, extraProps := analyzeNamedBlock(node, str.Name, impl, &offset)
				str.SizeInBytes += extraSize
				str.OrderedProperties = append(str.OrderedProperties, extraProps...)
				if symbol != nil {
					currentScope.NamedBlocks.Add(*symbol, symbol.Name)
				}
			default:
				symbol := analyzeVariable(node)
				symbol.Offset.Value = offset
				symbol.Offset.IsSet = true
				symbol.Ctx = StructProp
				str.OrderedProperties = append(str.OrderedProperties, symbol.Name)
				size := symbol.Type.GetSizeInBytes()
				offset += uint32(size)
				str.SizeInBytes += size
				globalScope.Structs.update(*str, str.Name)
				if symbol != nil {
					currentScope.Variables.Add(*symbol, symbol.Name)
				}
			}
		}
		str.implFnNames = map[string][]string{}
		namedblocks := currentScope.NamedBlocks
		for j := range namedblocks.OrderedNames {
			nb := namedblocks.GetByIndex(j)
			if nb == nil {
				continue
			}
			functions := nb.InnerScope.Functions
			for k := range functions.OrderedNames {
				fn := functions.GetByIndex(k)
				if fn == nil {
					continue
				}
				str.implFnNames[fn.Name] = append(str.implFnNames[fn.Name], nb.Name)
			}
		}
		globalScope.Structs.update(*str, str.Name)
	}

	currentScope = globalScope // reset the current scope
}

// Pass four
func collectFunctionSignatures(ast *parser.AST) {
	for _, node := range ast.Children {
		if node.Label == "fn" {
			symbol := processFunctionSignature(node)
			if overload, err := create(globalScope.Functions, symbol); err != nil {
				messages.Complain(diagnostic.IllegalStatementError, node.Location, "%s", err.Error())
			} else {
				node.IRName = overload.IRName
			}
		}
	}
}

// Pass five
func analyzeInterfaceImplementation() {
	structs := globalScope.Structs
	for i := range structs.OrderedNames {
		str := structs.GetByIndex(i)
		if str == nil {
			continue
		}
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
				str.InnerScope.Variables.Add(VariableSymbol{
					Name:        intfName,
					Type:        dt.NewContainerType(dt.ScopeRef, dt.NewReferenceSubType(str.Name), dt.NewReferenceSubType(intfName)),
					isPrivate:   false,
					isMutable:   false,
					Def:         namedBlock.Def,
					Initialized: true,
				}, intfName)
				functions := intf.innerScope.Functions
				for j := range functions.OrderedNames {
					fn := functions.GetByIndex(j)
					if fn == nil {
						continue
					}
					missing := false
					returnStr := ""
					if !fn.ReturnType.Equals(dt.NoneType) {
						returnStr = fmt.Sprintf("->%s", fn.ReturnType)
					}
					nb_fn := namedBlock.InnerScope.LookupFunctionByName(fn.Name)
					if nb_fn == nil {
						missing = true
						namedBlock.InnerScope.Functions.Add(*fn, fn.Name)
						nb_fn = namedBlock.InnerScope.LookupFunctionByName(fn.Name)
					} else if !nb_fn.ReturnType.Equals(fn.ReturnType) {
						messages.Complain(diagnostic.ImplementationError, namedBlock.Def.Location, "Implementation function %s returns %s but interface %s returns %s", fn.Name, nb_fn.ReturnType, intfName, fn.ReturnType)
						continue
					}
					copy := *nb_fn
					for k, overload := range fn.Overloads {
						params := dt.JoinTypes(overload.Parameters)
						if missing {
							str.UpdateImplFnNames(fn.Name, intfName)
							if overload.HasDefaultImplementation { // copy it over from the interface
								log.Println(k, overload.IRName, namedBlock.Name, str.Name)
								nb_fn.Overloads[k].Parameters = overload.Parameters
								nb_fn.Overloads[k].ParameterNames = overload.ParameterNames
								nb_fn.Overloads[k].InnerScope = namedBlock.InnerScope.addChild(fmt.Sprintf("%s@%s", fn.Name, namedBlock.InnerScope.Id), Function)
								nb_fn.Overloads[k].InnerScope.Variables = overload.InnerScope.Variables
								nb_fn.Overloads[k].IRName = strings.Replace(overload.IRName, fmt.Sprintf("__%s", intfName), fmt.Sprintf("__%s_%s", str.Name, intfName), 1)
								namedBlock.InnerScope.Functions.update(*nb_fn, fn.Name)
							} else {
								messages.Complain(diagnostic.ImplementationError, namedBlock.Def.Location, "Interface %s implementation missing 'fn %s(%s)%s'", intfName, fn.Name, params, returnStr)
							}
						} else {
							match := nb_fn.GetMatchingOverload(overload.Parameters)
							if overload.HasDefaultImplementation {
								if match == nil {
									copy.Overloads[k].Parameters = overload.Parameters
									copy.Overloads[k].ParameterNames = overload.ParameterNames
									namedBlock.InnerScope.Functions.update(copy, nb_fn.Name)
								}
							} else {
								if match == nil {
									messages.Complain(diagnostic.ImplementationError, namedBlock.Def.Location, "Interface %s implementation missing 'fn %s(%s)%s'", intfName, fn.Name, params, returnStr)
								}
							}
						}
					}
				}
				functions = namedBlock.InnerScope.Functions
				for j := range functions.OrderedNames {
					fn := functions.GetByIndex(j)
					if fn == nil {
						continue
					}
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
			globalScope.Variables.Add(*symbol, symbol.Name)
		}
	}
}

// Pass seven
func analyzeInterfaceFnBodies() {
	interfaces := globalScope.Interfaces
	for i := range interfaces.OrderedNames {
		intf := interfaces.GetByIndex(i)
		if intf == nil {
			continue
		}
		functions := intf.innerScope.Functions
		for j := range functions.OrderedNames {
			if fn := functions.GetByIndex(j); fn != nil {
				analyzeFunctionBody(*fn)
			}
		}
	}
}

// Pass eight
func analyzeStructMethodBodies() {
	for i := range globalScope.Structs.OrderedNames {
		str := globalScope.Structs.GetByIndex(i)
		if str == nil {
			continue
		}
		functions := str.InnerScope.Functions
		for j := range functions.OrderedNames {
			if fn := functions.GetByIndex(j); fn != nil {
				analyzeFunctionBody(*fn)
			}
		}
		for j := range str.InnerScope.NamedBlocks.OrderedNames {
			nb := str.InnerScope.NamedBlocks.GetByIndex(j)
			if nb == nil {
				continue
			}
			functions := nb.InnerScope.Functions
			for k := range functions.OrderedNames {
				if fn := functions.GetByIndex(k); fn != nil {
					analyzeFunctionBody(*fn)
				}
			}
		}
	}
}

// Pass nine
func analyzeFunctionsBodies() {
	functions := globalScope.Functions
	for i := range functions.OrderedNames {
		if fn := functions.GetByIndex(i); fn != nil {
			analyzeFunctionBody(*fn)
		}
	}
}
