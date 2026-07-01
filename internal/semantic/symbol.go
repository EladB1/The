package semantic

import (
	"fmt"
	"strings"

	"github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/parser"
)

type (
	LookupMode int
	Symbol     interface {
		getSymbolType() string
	}
	FunctionSymbol struct {
		name                     string
		parameters               []datatypes.DataType
		returnType               datatypes.DataType
		isPrivate                bool
		hasDefaultImplementation bool
		Def                      *parser.AST
		innerScope               *Scope
	}
	VariableSymbol struct {
		name      string
		Type      datatypes.DataType
		isPrivate bool
		isMutable bool
		Def       *parser.AST
	}
	InterfaceSymbol struct {
		name       string
		Def        *parser.AST
		innerScope *Scope
	}
	StructSymbol struct {
		name        string
		implements  []string
		sizeInBytes int
		Def         *parser.AST
		innerScope  *Scope
	}
	NamedBlockSymbol struct {
		name           string
		isSpecialBlock bool
		Def            *parser.AST
		innerScope     *Scope
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

const (
	ANY LookupMode = iota
	TYPE
	Struct
	Interface
	Function
	Variable
	NB
)

func (scope *Scope) lookup(name string, mode LookupMode) Symbol {
	curr := scope
	for curr != nil {
		if mode == ANY || mode == TYPE || mode == Interface {
			if intf, ok := scope.interfaces[name]; ok {
				return intf
			}
		}
		if mode == ANY || mode == TYPE || mode == Struct {
			if str, ok := scope.structs[name]; ok {
				return str
			}
		}
		if mode == ANY || mode == Function {
			if fn, ok := scope.functions[name]; ok {
				return fn
			}
		}
		if mode == ANY || mode == Variable {
			if variable, ok := scope.variables[name]; ok {
				return variable
			}
		}
		if mode == ANY || mode == NB {
			if nb, ok := scope.namedBlocks[name]; ok {
				return nb
			}
		}
		curr = curr.parent
	}
	return nil
}

func (fn FunctionSymbol) getSignature() string {
	builder := strings.Builder{}
	builder.WriteString("fn")
	builder.WriteString(fmt.Sprintf(" %s(", fn.name))
	for i, param := range fn.parameters {
		builder.WriteString(string(param.String()))
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

func (intf InterfaceSymbol) String() string {
	return fmt.Sprintf("{name: %s}", intf.name)
}

func (str StructSymbol) String() string {
	impl := strings.Builder{}
	if len(str.implements) != 0 {
		impl.WriteString(", implements: [")
		for i, intf := range str.implements {
			impl.WriteString(intf)
			if i != len(str.implements)-1 {
				impl.WriteString(", ")
			}
		}
		impl.WriteRune(']')
	}
	return fmt.Sprintf("{name: %s, size: %d%s}", str.name, str.sizeInBytes, impl.String())
}

func (fn FunctionSymbol) String() string {
	sig := fn.getSignature()
	priv := ""
	if fn.isPrivate {
		priv = ", isPrivate: true"
	}
	return fmt.Sprintf("{Signature: %s%s, hasImplementation: %v}", sig, priv, fn.hasDefaultImplementation)
}

func (variable VariableSymbol) String() string {
	priv := ""
	mut := ""
	if variable.isPrivate {
		priv = ", isPrivate: true"
	}
	if variable.isMutable {
		mut = ", isMutable: true"
	}
	return fmt.Sprintf("{name: %s, Type: %s%s%s}", variable.name, variable.Type, priv, mut)
}

func (nb NamedBlockSymbol) String() string {
	return fmt.Sprintf("{name: %s}", nb.name)
}
