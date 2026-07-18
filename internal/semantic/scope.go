package semantic

import (
	"fmt"
	"slices"
	"strings"

	"github.com/EladB1/The/internal/datatypes"
)

type ScopeType int

const (
	Default ScopeType = iota
	Interface
	Struct
	NamedBlock
	Function
	Loop
	Branch
)

type Scope struct {
	Id          string
	Kind        ScopeType
	Parent      *Scope
	Children    []*Scope
	Functions   FunctionSymbolTable
	Variables   VariableSymbolTable
	Interfaces  InterfaceSymbolTable
	Structs     StructSymbolTable
	NamedBlocks NamedBlockSymbolTable
}

func (scope *Scope) addChild(id string, kind ScopeType) *Scope {
	newScope := Scope{
		Id:          id,
		Kind:        kind,
		Parent:      scope,
		Functions:   FunctionSymbolTable{},
		Variables:   VariableSymbolTable{},
		Interfaces:  InterfaceSymbolTable{},
		Structs:     StructSymbolTable{},
		NamedBlocks: NamedBlockSymbolTable{},
	}
	scope.Children = append(scope.Children, &newScope)
	return &newScope
}

func (scope *Scope) String() string {
	return scope.to_string(0)
}

func (scope *Scope) to_string(indentLevel int) string {
	prefix := strings.Repeat("\t", indentLevel)
	builder := strings.Builder{}
	builder.WriteString(prefix)
	builder.WriteString("Scope: { id: ")
	builder.WriteString(scope.Id)
	if scope.Kind != Default {
		builder.WriteString(fmt.Sprintf(", type: %v", scope.Kind))
	}
	if scope.Parent != nil {
		builder.WriteString(", parent: ")
		builder.WriteString(scope.Parent.Id)
	}
	if len(scope.Interfaces) != 0 {
		builder.WriteString(fmt.Sprintf(", interfaces: %v", scope.Interfaces))
	}
	if len(scope.Structs) != 0 {
		builder.WriteString(fmt.Sprintf(", structs: %v", scope.Structs))
	}
	if len(scope.NamedBlocks) != 0 {
		builder.WriteString(fmt.Sprintf(", namedBlocks: %v", scope.NamedBlocks))
	}
	if len(scope.Functions) != 0 {
		builder.WriteString(fmt.Sprintf(", functions: %v", scope.Functions))
	}
	if len(scope.Variables) != 0 {
		builder.WriteString(fmt.Sprintf(", variables: %v", scope.Variables))
	}
	count := len(scope.Children)
	if count > 0 {
		builder.WriteString(", children: [\n")
		for i, child := range scope.Children {
			builder.WriteString(child.to_string(indentLevel + 1))
			if i != count-1 {
				builder.WriteString(",\n")
			}
		}
		builder.WriteString(fmt.Sprintf("\n%s]", prefix))
	}
	builder.WriteString(" }")
	return builder.String()
}

func (scope *Scope) HasParentScope(other *Scope) bool {
	if scope.Id == other.Id {
		return true
	}
	curr := scope
	for curr != rootScope {
		if curr.Id == other.Id {
			return true
		}
		curr = curr.Parent
	}
	return false
}

func (scope *Scope) HasScopeTypeAncestor(sType ScopeType) bool {
	if scope.Kind == sType {
		return true
	}
	curr := scope
	for curr != rootScope {
		if curr.Kind == sType {
			return true
		}
		curr = curr.Parent
	}
	return false
}

func (scope *Scope) LookupType(name string) TypeSymbol {
	curr := scope
	for curr != nil {
		if intf, ok := curr.Interfaces[name]; ok {
			return intf
		}
		if str, ok := curr.Structs[name]; ok {
			return str
		}
		curr = curr.Parent
	}
	return nil
}

func (scope *Scope) LookupInterface(name string) *InterfaceSymbol {
	curr := scope
	for curr != nil {
		if intf, ok := curr.Interfaces[name]; ok {
			return &intf
		}
		curr = curr.Parent
	}
	return nil
}

func (scope *Scope) lookupStruct(name string) *StructSymbol {
	curr := scope
	for curr != nil {
		if str, ok := curr.Structs[name]; ok {
			return &str
		}
		curr = curr.Parent
	}
	return nil
}

func (scope *Scope) LookupNamedBlock(name string) *NamedBlockSymbol {
	curr := scope
	for curr != nil {
		if nb, ok := curr.NamedBlocks[name]; ok {
			return &nb
		}
		curr = curr.Parent
	}
	return nil
}

func (nb NamedBlockSymbol) HasReturnType(returnType datatypes.DataType) bool {
	for _, fnSymbol := range nb.InnerScope.Functions {
		if fnSymbol.ReturnType == returnType {
			return true
		}
	}
	return false
}

func (scope *Scope) LookupVariable(name string) *VariableSymbol {
	curr := scope
	for curr != nil {
		if variable, ok := curr.Variables[name]; ok {
			return &variable
		}
		curr = curr.Parent
	}
	return nil
}

func (scope *Scope) LookupFunctionsByReturnType(returnType datatypes.DataType) []*FunctionSymbol {
	matching := []*FunctionSymbol{}
	for _, fn := range scope.Functions {
		if fn.ReturnType == returnType {
			matching = append(matching, &fn)
		}
	}
	return matching
}

func (scope *Scope) LookupFunctionByName(name string) *FunctionSymbol {
	curr := scope
	for curr != nil {
		if fn, ok := curr.Functions[name]; ok {
			return &fn
		}
		curr = curr.Parent
	}
	return nil
}

func (table FunctionSymbolTable) add(symbol FnCreateSymbol) error {
	fn, ok := table[symbol.name]
	if ok {
		if fn.ReturnType != symbol.returnType {
			if fn.ReturnType == datatypes.None {
				return fmt.Errorf("Function name '%s' already defined without a return type; cannot overload with return type %s", symbol.name, symbol.returnType)
			}
			return fmt.Errorf("Function name '%s' can only be overloaded with return type %s. Found: %s", symbol.name, fn.ReturnType, symbol.returnType)
		}
		if fn.getMatchingOverload(symbol.parameters) != nil {
			return fmt.Errorf("Function with signature '%s' cannot be redefined", symbol.getSignature())
		} else {
			fn.Overloads = append(fn.Overloads, symbol.toOverload())
			table[fn.Name] = fn
		}
	} else {
		table[symbol.name] = FunctionSymbol{
			Name:       symbol.name,
			ReturnType: symbol.returnType,
			Overloads:  []FnOverloadSymbol{symbol.toOverload()},
		}
	}
	return nil
}

func ImplementsInterface(possibleIntf, type_ datatypes.DataType) bool {
	if possibleIntf.IsPrimitive() || type_.IsPrimitive() {
		return false
	}
	if intf := globalScope.LookupInterface(possibleIntf.String()); intf != nil {
		if str := globalScope.lookupStruct(type_.String()); str != nil {
			return slices.Contains(str.Implements, intf.name)
		}
	}
	return false
}

func FindAncestorScopeById(id string) *Scope {
	if id == globalScope.Id {
		return globalScope
	}
	if id == currentScope.Id {
		return currentScope
	}
	for scope := currentScope; scope.Id != rootScope.Id; scope = scope.Parent {
		if id == scope.Id {
			return scope
		}
	}
	return nil
}
