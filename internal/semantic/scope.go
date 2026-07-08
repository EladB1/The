package semantic

import (
	"fmt"
	"strings"

	"github.com/EladB1/The/internal/datatypes"
)

type Scope struct {
	id          string
	parent      *Scope
	children    []*Scope
	functions   FunctionSymbolTable
	variables   VariableSymbolTable
	interfaces  InterfaceSymbolTable
	structs     StructSymbolTable
	namedBlocks NamedBlockSymbolTable
}

func (scope *Scope) addChild(id string) *Scope {
	newScope := Scope{
		id:          id,
		parent:      scope,
		functions:   FunctionSymbolTable{},
		variables:   VariableSymbolTable{},
		interfaces:  InterfaceSymbolTable{},
		structs:     StructSymbolTable{},
		namedBlocks: NamedBlockSymbolTable{},
	}
	scope.children = append(scope.children, &newScope)
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
	builder.WriteString(scope.id)
	if scope.parent != nil {
		builder.WriteString(", parent: ")
		builder.WriteString(scope.parent.id)
	}
	if len(scope.interfaces) != 0 {
		builder.WriteString(fmt.Sprintf(", interfaces: %v", scope.interfaces))
	}
	if len(scope.structs) != 0 {
		builder.WriteString(fmt.Sprintf(", structs: %v", scope.structs))
	}
	if len(scope.namedBlocks) != 0 {
		builder.WriteString(fmt.Sprintf(", namedBlocks: %v", scope.namedBlocks))
	}
	if len(scope.functions) != 0 {
		builder.WriteString(fmt.Sprintf(", functions: %v", scope.functions))
	}
	if len(scope.variables) != 0 {
		builder.WriteString(fmt.Sprintf(", variables: %v", scope.variables))
	}
	count := len(scope.children)
	if count > 0 {
		builder.WriteString(", children: [\n")
		for i, child := range scope.children {
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
	if scope.id == other.id {
		return true
	}
	curr := scope
	for curr != &rootScope {
		if curr.id == other.id {
			return true
		}
		curr = curr.parent
	}
	return false
}

func (scope *Scope) lookupType(name string) TypeSymbol {
	curr := scope
	for curr != nil {
		if intf, ok := curr.interfaces[name]; ok {
			return intf
		}
		if str, ok := curr.structs[name]; ok {
			return str
		}
		curr = curr.parent
	}
	return nil
}

func (scope *Scope) lookupInterface(name string) *InterfaceSymbol {
	curr := scope
	for curr != nil {
		if intf, ok := curr.interfaces[name]; ok {
			return &intf
		}
		curr = curr.parent
	}
	return nil
}

func (scope *Scope) lookupStruct(name string) *StructSymbol {
	curr := scope
	for curr != nil {
		if str, ok := curr.structs[name]; ok {
			return &str
		}
		curr = curr.parent
	}
	return nil
}

func (scope *Scope) lookupNamedBlock(name string) *NamedBlockSymbol {
	curr := scope
	for curr != nil {
		if nb, ok := curr.namedBlocks[name]; ok {
			return &nb
		}
		curr = curr.parent
	}
	return nil
}

func (nb NamedBlockSymbol) HasReturnType(returnType datatypes.DataType) bool {
	for _, fnSymbol := range nb.innerScope.functions {
		if fnSymbol.returnType == returnType {
			return true
		}
	}
	return false
}

func (scope *Scope) lookupVariable(name string) *VariableSymbol {
	curr := scope
	for curr != nil {
		if variable, ok := curr.variables[name]; ok {
			return &variable
		}
		curr = curr.parent
	}
	return nil
}

func (scope *Scope) lookupFunctionsByReturnType(returnType datatypes.DataType) []*FunctionSymbol {
	matching := []*FunctionSymbol{}
	for _, fn := range scope.functions {
		if fn.returnType == returnType {
			matching = append(matching, &fn)
		}
	}
	return matching
}

func (scope *Scope) lookupFunctionByName(name string) *FunctionSymbol {
	curr := scope
	for curr != nil {
		if fn, ok := curr.functions[name]; ok {
			return &fn
		}
		curr = curr.parent
	}
	return nil
}

func (table FunctionSymbolTable) add(symbol FnCreateSymbol) error {
	fn, ok := table[symbol.name]
	if ok {
		if fn.returnType != symbol.returnType {
			if fn.returnType == datatypes.None {
				return fmt.Errorf("Function name '%s' already defined without a return type; cannot overload with return type %s", symbol.name, symbol.returnType)
			}
			return fmt.Errorf("Function name '%s' can only be overloaded with return type %s. Found: %s", symbol.name, fn.returnType, symbol.returnType)
		}
		params := datatypes.Join(symbol.parameters)
		if _, ok := fn.overloads[params]; ok {
			return fmt.Errorf("Function with signature '%s' cannot be redefined", symbol.getSignature())
		} else {
			fn.overloads[params] = symbol.toOverload()
		}
	} else {
		params := datatypes.Join(symbol.parameters)
		table[symbol.name] = FunctionSymbol{
			name:       symbol.name,
			returnType: symbol.returnType,
			overloads: map[string]FnOverloadSymbol{
				params: symbol.toOverload(),
			},
		}
	}
	return nil
}
