package semantic

import (
	"fmt"
	"strings"

	"github.com/EladB1/The/internal/datatypes"
)

type (
	Symbol interface {
		getSymbolType() string
	}
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
	}
	InterfaceSymbol struct {
		name string
	}
	StructSymbol struct {
		name        string
		implements  []string
		sizeInBytes int
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

func (fn FunctionSymbol) getSymbolType() string {
	return "function"
}

func (intf InterfaceSymbol) getSymbolType() string {
	return "interface"
}

func (str StructSymbol) getSymbolType() string {
	return "struct"
}

func (variable VariableSymbol) getSymbolType() string {
	return "variable"
}

func (nb NamedBlockSymbol) getSymbolType() string {
	return "named-block"
}

func (scope *Scope) lookup(name string) Symbol {
	curr := scope
	for curr != nil {
		if intf, ok := scope.interfaces[name]; ok {
			return intf
		} else if str, ok := scope.structs[name]; ok {
			return str
		} else if fn, ok := scope.functions[name]; ok {
			return fn
		} else if variable, ok := scope.variables[name]; ok {
			return variable
		} else if nb, ok := scope.namedBlocks[name]; ok {
			return nb
		}
		curr = curr.parent
	}
	return nil
}

func (fn FunctionSymbol) getSignature() string {
	builder := strings.Builder{}
	builder.WriteString("fn")
	builder.WriteString(fmt.Sprintf("%s(", fn.name))
	for i, param := range fn.parameters {
		builder.WriteString(string(param))
		if i < len(fn.parameters)-1 {
			builder.WriteByte(',')
		}
	}
	builder.WriteByte(')')
	if fn.returnType != datatypes.None {
		builder.WriteString(fmt.Sprintf("->%s", fn.returnType))
	}
	return builder.String()
}
