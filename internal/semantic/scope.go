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
