package semantic

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	dt "github.com/EladB1/The/internal/datatypes"
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
	Functions   *SymbolTable[FunctionSymbol]
	Variables   *SymbolTable[VariableSymbol]
	Interfaces  *SymbolTable[InterfaceSymbol]
	Structs     *SymbolTable[StructSymbol]
	NamedBlocks *SymbolTable[NamedBlockSymbol]
}

type SerializedScope struct {
	Id          string
	Kind        ScopeType
	ParentId    string
	Children    []*Scope
	Functions   *SymbolTable[FunctionSymbol]
	Variables   *SymbolTable[VariableSymbol]
	Interfaces  *SymbolTable[InterfaceSymbol]
	Structs     *SymbolTable[StructSymbol]
	NamedBlocks *SymbolTable[NamedBlockSymbol]
}

func (scope *Scope) MarshalJSON() ([]byte, error) {
	parentId := ""
	if scope.Parent != nil {
		parentId = scope.Parent.Id
	}
	return json.Marshal(SerializedScope{
		Id:          scope.Id,
		Kind:        scope.Kind,
		ParentId:    parentId,
		Children:    scope.Children,
		Functions:   scope.Functions,
		Variables:   scope.Variables,
		Interfaces:  scope.Interfaces,
		Structs:     scope.Structs,
		NamedBlocks: scope.NamedBlocks,
	})
}

func (scope *Scope) UnmarshalJSON(data []byte) error {
	var serialized SerializedScope
	if err := json.Unmarshal(data, &serialized); err != nil {
		return err
	}
	*scope = serialized.transform()
	return nil
}

func (serialized *SerializedScope) transform() Scope {
	scope := Scope{
		Id:          serialized.Id,
		Kind:        serialized.Kind,
		Functions:   serialized.Functions,
		Variables:   serialized.Variables,
		Interfaces:  serialized.Interfaces,
		Structs:     serialized.Structs,
		NamedBlocks: serialized.NamedBlocks,
		Children:    serialized.Children,
	}
	return scope
}

func (scope *Scope) addChild(id string, kind ScopeType) *Scope {
	newScope := Scope{
		Id:          id,
		Kind:        kind,
		Parent:      scope,
		Functions:   NewTable[FunctionSymbol](),
		Variables:   NewTable[VariableSymbol](),
		Interfaces:  NewTable[InterfaceSymbol](),
		Structs:     NewTable[StructSymbol](),
		NamedBlocks: NewTable[NamedBlockSymbol](),
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
	if !scope.Interfaces.isEmpty() {
		builder.WriteString(fmt.Sprintf(", interfaces: %v", scope.Interfaces))
	}
	if !scope.Structs.isEmpty() {
		builder.WriteString(fmt.Sprintf(", structs: %v", scope.Structs))
	}
	if !scope.NamedBlocks.isEmpty() {
		builder.WriteString(fmt.Sprintf(", namedBlocks: %v", scope.NamedBlocks))
	}
	if !scope.Functions.isEmpty() {
		builder.WriteString(fmt.Sprintf(", functions: %v", scope.Functions))
	}
	if !scope.Variables.isEmpty() {
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

func (scope *Scope) GetChildScopeById(id string) *Scope {
	for _, child := range scope.Children {
		if child.Id == id {
			return child
		}
	}
	return nil
}

func (scope *Scope) LookupType(name string) TypeSymbol {
	curr := scope
	for curr != nil {
		if intf, ok := curr.Interfaces.Lookup(name); ok {
			return intf
		}
		if str, ok := curr.Structs.Lookup(name); ok {
			return str
		}
		curr = curr.Parent
	}
	return nil
}

func (scope *Scope) LookupInterface(name string) *InterfaceSymbol {
	curr := scope
	for curr != nil {
		if intf, ok := curr.Interfaces.Lookup(name); ok {
			return &intf
		}
		curr = curr.Parent
	}
	return nil
}

func (scope *Scope) LookupStruct(name string) *StructSymbol {
	curr := scope
	for curr != nil {
		if str, ok := curr.Structs.Lookup(name); ok {
			return &str
		}
		curr = curr.Parent
	}
	return nil
}

func (scope *Scope) LookupNamedBlock(name string) *NamedBlockSymbol {
	curr := scope
	for curr != nil {
		if nb, ok := curr.NamedBlocks.Lookup(name); ok {
			return &nb
		}
		curr = curr.Parent
	}
	return nil
}

func (nb NamedBlockSymbol) HasReturnType(returnType dt.SourceType) bool {
	for fnSymbol := range nb.InnerScope.Functions.All() {
		if fnSymbol.ReturnType.Equals(returnType) {
			return true
		}
	}
	return false
}

func (scope *Scope) LookupVariable(name string) *VariableSymbol {
	curr := scope
	for curr != nil {
		if variable, ok := curr.Variables.Lookup(name); ok {
			return &variable
		}
		curr = curr.Parent
	}
	return nil
}

func (scope *Scope) LookupFunctionsByReturnType(returnType dt.SourceType) []*FunctionSymbol {
	matching := []*FunctionSymbol{}
	for fn := range scope.Functions.All() {
		if fn.ReturnType.Equals(returnType) {
			matching = append(matching, &fn)
		}
	}
	return matching
}

func (scope *Scope) LookupFunctionByName(name string) *FunctionSymbol {
	curr := scope
	for curr != nil {
		if fn, ok := curr.Functions.Lookup(name); ok {
			return &fn
		}
		curr = curr.Parent
	}
	return nil
}

func (scope *Scope) LookupFunctionByNameAndIRName(name, irName string) *FnOverloadSymbol {
	curr := scope
	for curr != nil {
		if fn, ok := curr.Functions.Lookup(name); ok {
			for _, overload := range fn.Overloads {
				if overload.IRName == irName {
					return &overload
				}
			}
		}
		curr = curr.Parent
	}
	return nil
}

func add(symbol FnCreateSymbol, table *FunctionTable) (*FnOverloadSymbol, error) {
	var overload *FnOverloadSymbol = nil
	fn, ok := table.Lookup(symbol.name)
	if ok {
		if !fn.ReturnType.Equals(symbol.returnType) {
			if fn.ReturnType.Equals(dt.NoneType) {
				return overload, fmt.Errorf("Function name '%s' already defined without a return type; cannot overload with return type %s", symbol.name, symbol.returnType)
			}
			return overload, fmt.Errorf("Function name '%s' can only be overloaded with return type %s. Found: %s", symbol.name, fn.ReturnType, symbol.returnType)
		}
		if fn.GetMatchingOverload(symbol.parameters) != nil {
			return overload, fmt.Errorf("Function with signature '%s' cannot be redefined", symbol.getSignature())
		} else {
			overload = symbol.toOverload(true)
			fn.Overloads = append(fn.Overloads, *overload)
			table.update(fn, fn.Name)
		}
	} else {
		overload = symbol.toOverload(false)
		table.Add(FunctionSymbol{
			Name:       symbol.name,
			ReturnType: symbol.returnType,
			Overloads:  []FnOverloadSymbol{*overload},
		}, symbol.name)
	}
	return overload, nil
}

func ImplementsInterface(possibleIntf, type_ dt.SourceType) bool {
	if !possibleIntf.IsDynamic || !type_.IsDynamic {
		return false
	}
	if intf := globalScope.LookupInterface(possibleIntf.String()); intf != nil {
		if str := globalScope.LookupStruct(type_.String()); str != nil {
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
