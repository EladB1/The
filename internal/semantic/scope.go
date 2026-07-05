package semantic

import (
	"fmt"
	"strings"
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
	fmt.Printf(scope.id, other.id)
	curr := scope
	for curr != &rootScope {
		if curr.id == other.id {
			return true
		}
		curr = curr.parent
	}
	return false
}
