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
		name       string
		returnType datatypes.DataType
		overloads  map[string]FnOverloadSymbol
	}
	FnOverloadSymbol struct {
		parameters               []datatypes.DataType
		isPrivate                bool
		hasDefaultImplementation bool
		Body                     *parser.AST
		innerScope               *Scope
	}
	FnCreateSymbol struct {
		name                     string
		returnType               datatypes.DataType
		parameters               []datatypes.DataType
		isPrivate                bool
		hasDefaultImplementation bool
		Body                     *parser.AST
		innerScope               *Scope
	}
	VariableSymbol struct {
		name        string
		Type        datatypes.DataType
		isPrivate   bool
		isMutable   bool
		Def         *parser.AST
		Initialized bool
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

func (fn FnCreateSymbol) getSymbolType() string {
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
	overloads := strings.Builder{}
	for key, symbol := range fn.overloads {
		priv := ""
		if symbol.isPrivate {
			priv = ", isPrivate: true"
		}
		overloads.WriteString(fmt.Sprintf("{parameters: (%s)%s, implemented: %v}", key, priv, symbol.hasDefaultImplementation))
	}
	return fmt.Sprintf("{name: %s, returns: %s, overloads: [%s]}", fn.name, fn.returnType, overloads.String())
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
	return fmt.Sprintf("{name: %s, Type: %s%s%s, Initialized: %v}", variable.name, variable.Type, priv, mut, variable.Initialized)
}

func (nb NamedBlockSymbol) String() string {
	return fmt.Sprintf("{name: %s}", nb.name)
}

func (symbol FnCreateSymbol) stringifyParams() string {
	paramStr := strings.Builder{}
	for i, param := range symbol.parameters {
		paramStr.WriteString(param.String())
		if i < len(symbol.parameters)-1 {
			paramStr.WriteRune(',')
		}
	}
	return paramStr.String()
}

func (symbol FnCreateSymbol) getSignature() string {
	returns := ""
	if symbol.returnType != datatypes.None {
		returns = fmt.Sprintf("->%s", symbol.returnType)
	}
	return fmt.Sprintf("fn %s(%s)%s", symbol.name, symbol.stringifyParams(), returns)
}

func (symbol FnCreateSymbol) toOverload() FnOverloadSymbol {
	return FnOverloadSymbol{
		parameters:               symbol.parameters,
		hasDefaultImplementation: symbol.hasDefaultImplementation,
		isPrivate:                symbol.isPrivate,
		Body:                     symbol.Body,
		innerScope:               symbol.innerScope,
	}
}

func (table FunctionSymbolTable) add(symbol FnCreateSymbol) error {
	fn, ok := table[symbol.name]
	if ok {
		if fn.returnType != symbol.returnType {
			return fmt.Errorf("Function name '%s' can only be overloaded with return type %s. Found: %s", symbol.name, fn.returnType, symbol.returnType)
		}
		params := symbol.stringifyParams()
		if _, ok := fn.overloads[params]; ok {
			return fmt.Errorf("Function with signature '%s' cannot be redefined", symbol.getSignature())
		} else {
			fn.overloads[params] = symbol.toOverload()
		}
	} else {
		params := symbol.stringifyParams()
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
