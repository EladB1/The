package semantic

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

func (scope *Scope) addChild(id string) {
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
}
