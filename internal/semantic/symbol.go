package semantic

import "github.com/EladB1/The/internal/datatypes"

type (
	FunctionSymbol struct {
		name                     string
		parameters               []datatypes.DataType
		returnType               datatypes.DataType
		isPrivate                bool
		hasDefaultImplementation bool
		// overrides?
	}
	VariableSymbol struct {
		name      string
		Type      datatypes.DataType
		isPrivate bool
		isMutable bool
		size      int
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
