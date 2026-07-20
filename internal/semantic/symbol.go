package semantic

import (
	"fmt"
	"strings"

	"github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/parser"
)

type (
	TypeSymbol interface {
		GetSymbolType() string
		GetInnerScope() *Scope
		GetNamedBlockIfExists(string) *NamedBlockSymbol
		getConflicts(string) []string
	}
	FunctionSymbol struct {
		Name       string
		ReturnType datatypes.SourceType
		Overloads  []FnOverloadSymbol
	}
	FnOverloadSymbol struct {
		Parameters               []datatypes.SourceType
		IsPrivate                bool
		HasDefaultImplementation bool
		Body                     *parser.AST
		InnerScope               *Scope
	}
	FnCreateSymbol struct {
		name                     string
		returnType               datatypes.SourceType
		parameters               []datatypes.SourceType
		isPrivate                bool
		hasDefaultImplementation bool
		Body                     *parser.AST
		innerScope               *Scope
	}
	VariableCtx    string
	VariableSymbol struct {
		Name        string
		Type        datatypes.SourceType
		isPrivate   bool
		isMutable   bool
		Def         *parser.AST
		Initialized bool
		Ctx         VariableCtx
	}
	InterfaceSymbol struct {
		name       string
		Def        *parser.AST
		innerScope *Scope
	}
	StructSymbol struct {
		Name        string
		Implements  []string
		SizeInBytes int
		Def         *parser.AST
		InnerScope  *Scope
		implFnNames map[string][]string
	}
	NamedBlockSymbol struct {
		Name           string
		isSpecialBlock bool
		Def            *parser.AST
		InnerScope     *Scope
	}

	FunctionSymbolTable   map[string]FunctionSymbol
	VariableSymbolTable   map[string]VariableSymbol
	InterfaceSymbolTable  map[string]InterfaceSymbol
	StructSymbolTable     map[string]StructSymbol
	NamedBlockSymbolTable map[string]NamedBlockSymbol
)

/* TypeSymbol interface functions */
func (intf InterfaceSymbol) GetSymbolType() string {
	return "interface"
}
func (str StructSymbol) GetSymbolType() string {
	return "struct"
}

func (intf InterfaceSymbol) GetInnerScope() *Scope {
	return intf.innerScope
}
func (str StructSymbol) GetInnerScope() *Scope {
	return str.InnerScope
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

func (str StructSymbol) UpdateImplFnNames(fn string, intf string) {
	if names, ok := str.implFnNames[fn]; ok {
		str.implFnNames[fn] = append(names, intf)
	} else {
		str.implFnNames[fn] = []string{intf}
	}
}

func (intf InterfaceSymbol) GetNamedBlockIfExists(name string) *NamedBlockSymbol {
	return nil
}

func (str StructSymbol) GetNamedBlockIfExists(name string) *NamedBlockSymbol {
	return str.InnerScope.LookupNamedBlock(name)
}

func (intf InterfaceSymbol) String() string {
	return fmt.Sprintf("{name: %s}", intf.name)
}

func (str StructSymbol) String() string {
	impl := strings.Builder{}
	if len(str.Implements) != 0 {
		impl.WriteString(", implements: [")
		for i, intf := range str.Implements {
			impl.WriteString(intf)
			if i != len(str.Implements)-1 {
				impl.WriteString(", ")
			}
		}
		impl.WriteRune(']')
	}
	return fmt.Sprintf("{name: %s, size: %d%s}", str.Name, str.SizeInBytes, impl.String())
}

func (fn FunctionSymbol) String() string {
	overloads := strings.Builder{}
	for _, symbol := range fn.Overloads {
		priv := ""
		if symbol.IsPrivate {
			priv = ", isPrivate: true"
		}
		overloads.WriteString(fmt.Sprintf("{parameters: (%s)%s, implemented: %v}", datatypes.Join(symbol.Parameters), priv, symbol.HasDefaultImplementation))
	}
	return fmt.Sprintf("{name: %s, returns: %s, overloads: [%s]}", fn.Name, fn.ReturnType, overloads.String())
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
	return fmt.Sprintf("{name: %s, Type: %s%s%s, Initialized: %v}", variable.Name, variable.Type, priv, mut, variable.Initialized)
}

func (nb NamedBlockSymbol) String() string {
	return fmt.Sprintf("{name: %s}", nb.Name)
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
		Parameters:               symbol.parameters,
		HasDefaultImplementation: symbol.hasDefaultImplementation,
		IsPrivate:                symbol.isPrivate,
		Body:                     symbol.Body,
		InnerScope:               symbol.innerScope,
	}
}

func (fn FunctionSymbol) getMatchingOverload(params []datatypes.SourceType) *FnOverloadSymbol {
	count := len(params)
	for _, overload := range fn.Overloads {
		matches := false
		if count == len(overload.Parameters) {
			if count == 0 {
				return &overload
			}
			for i := range count {
				param := overload.Parameters[i]
				if params[i] == param || param == datatypes.Any || ImplementsInterface(param, params[i]) {
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

const (
	Local  VariableCtx = "local"
	Global VariableCtx = "global"
	Param  VariableCtx = "param"
)
