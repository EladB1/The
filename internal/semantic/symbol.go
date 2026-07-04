package semantic

import (
	"fmt"
	"strings"

	"github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/parser"
)

type (
	LookupMode int
	TypeSymbol interface {
		getSymbolType() string
		getInnerScope() *Scope
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

/* TypeSymbol interface functions */
func (intf InterfaceSymbol) getSymbolType() string {
	return "interface"
}
func (str StructSymbol) getSymbolType() string {
	return "struct"
}

func (intf InterfaceSymbol) getInnerScope() *Scope {
	return intf.innerScope
}
func (str StructSymbol) getInnerScope() *Scope {
	return str.innerScope
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

func (scope *Scope) lookupFunction(name string) *FunctionSymbol {
	curr := scope
	for curr != nil {
		if fn, ok := curr.functions[name]; ok {
			return &fn
		}
		curr = curr.parent
	}
	return nil
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

func (nb NamedBlockSymbol) HasReturnType(returnType datatypes.DataType) bool {
	for _, fnSymbol := range nb.innerScope.functions {
		if fnSymbol.returnType == returnType {
			return true
		}
	}
	return false
}

func (symbol FnCreateSymbol) getSignature() string {
	returns := ""
	if symbol.returnType != datatypes.None {
		returns = fmt.Sprintf("->%s", symbol.returnType)
	}
	return fmt.Sprintf("fn %s(%s)%s", symbol.name, datatypes.Join(symbol.parameters), returns)
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
