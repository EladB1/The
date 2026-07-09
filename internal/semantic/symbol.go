package semantic

import (
	"fmt"
	"slices"
	"strings"

	"github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/parser"
)

type (
	TypeSymbol interface {
		getSymbolType() string
		getInnerScope() *Scope
		getNamedBlockIfExists(string) *NamedBlockSymbol
		getConflicts(string) []string
	}
	FunctionSymbol struct {
		name       string
		returnType datatypes.DataType
		overloads  []FnOverloadSymbol
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
		implFnNames map[string][]string
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

func (intf InterfaceSymbol) getConflicts(fn string) []string {
	return nil
}

func (str StructSymbol) getConflicts(fn string) []string {
	if names, ok := str.implFnNames[fn]; ok {
		return names
	}
	return nil
}

func (intf InterfaceSymbol) getNamedBlockIfExists(name string) *NamedBlockSymbol {
	return nil
}

func (str StructSymbol) getNamedBlockIfExists(name string) *NamedBlockSymbol {
	return str.innerScope.lookupNamedBlock(name)
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
	for _, symbol := range fn.overloads {
		priv := ""
		if symbol.isPrivate {
			priv = ", isPrivate: true"
		}
		overloads.WriteString(fmt.Sprintf("{parameters: (%s)%s, implemented: %v}", datatypes.Join(symbol.parameters), priv, symbol.hasDefaultImplementation))
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

func (fn FunctionSymbol) getMatchingOverload(params []datatypes.DataType) *FnOverloadSymbol {
	count := len(params)
	for _, overload := range fn.overloads {
		matches := false
		if count == len(overload.parameters) {
			for i := range count {
				param := overload.parameters[i]
				isImplementor := false
				if !param.IsPrimitive() && !params[i].IsPrimitive() && params[i] != param {
					if intf := globalScope.lookupInterface(param.String()); intf != nil {
						if str := globalScope.lookupStruct(params[i].String()); str != nil {
							if slices.Contains(str.implements, param.String()) {
								isImplementor = true
							}
						}
					}
				}
				if params[i] == param || param == datatypes.Any || isImplementor {
					matches = true
				} else {
					matches = false
				}
			}
		}
		if matches {
			return &overload
		}
	}
	return nil

}
