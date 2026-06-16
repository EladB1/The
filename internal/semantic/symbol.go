package semantic

type (
	FunctionSymbol struct {
		name                     string
		parameters               []string
		returnType               string
		isPrivate                bool
		hasDefaultImplementation bool
		// overrides?
	}
	VariableSymbol struct {
		name      string
		Type      string // TODO: change?
		isPrivate bool
		isMutable bool
	}
	InterfaceSymbol struct {
		name string
	}
	StructSymbol struct {
		name       string
		implements []string
	}
	NamedBlockSymbol struct {
		name           string
		isSpecialBlock bool
	}

	FunctionSymbolTable   map[string]FunctionSymbol
	VariableSymbolTable   map[string]VariableSymbol
	InterfaceSymbolTable  map[string]InterfaceSymbol
	StructSymbolTable     map[string]StructSymbol
	NamedBlockSymbolTable map[string]NamedBlockSymbol
)
