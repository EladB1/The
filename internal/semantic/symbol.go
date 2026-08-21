package semantic

import (
	"fmt"
	"strings"

	dt "github.com/EladB1/The/internal/datatypes"
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
		ReturnType dt.SourceType
		Overloads  []FnOverloadSymbol
	}
	FnOverloadSymbol struct {
		Parameters               []dt.SourceType
		ParameterNames           []string
		IsPrivate                bool
		HasDefaultImplementation bool
		Body                     *parser.AST
		InnerScope               *Scope
		IRName                   string
	}
	FnCreateSymbol struct {
		name                     string
		returnType               dt.SourceType
		parameters               []dt.SourceType
		parameterNames           []string
		isPrivate                bool
		hasDefaultImplementation bool
		Body                     *parser.AST
		innerScope               *Scope
	}
	VariableCtx    string
	VariableSymbol struct {
		Name        string
		Type        dt.SourceType
		isPrivate   bool
		isMutable   bool
		Def         *parser.AST
		Initialized bool
		Ctx         VariableCtx
		Offset      OffsetValue
	}
	InterfaceSymbol struct {
		name       string
		Def        *parser.AST
		innerScope *Scope
	}
	StructSymbol struct {
		Name              string
		Implements        []string
		SizeInBytes       int
		Def               *parser.AST
		InnerScope        *Scope
		implFnNames       map[string][]string
		OrderedProperties []string
	}
	NamedBlockSymbol struct {
		Name           string
		isSpecialBlock bool
		Def            *parser.AST
		InnerScope     *Scope
	}

	SymbolTable[T any] struct {
		OrderedNames []string
		Symbols      map[string]T
	}

	FunctionTable = SymbolTable[FunctionSymbol]

	PrimitiveTypeMembers struct {
		Properties SymbolTable[VariableSymbol]
		Methods    FunctionTable
	}

	PrimitiveTypeTables map[dt.DataType]PrimitiveTypeMembers

	OffsetValue struct {
		IsSet bool
		Value uint32
	}
)

func NewTable[T any]() *SymbolTable[T] {
	return &SymbolTable[T]{
		OrderedNames: []string{},
		Symbols:      map[string]T{},
	}
}

func (table *SymbolTable[T]) Add(symbol T, name string) {
	table.OrderedNames = append(table.OrderedNames, name)
	table.Symbols[name] = symbol
}

func (table *SymbolTable[T]) update(symbol T, name string) {
	table.Symbols[name] = symbol
}

func (table *SymbolTable[T]) isEmpty() bool {
	return len(table.Symbols) == 0
}

func (table *SymbolTable[T]) Lookup(name string) *T {
	if symbol, ok := table.Symbols[name]; ok {
		return &symbol
	}
	return nil
}

func (table *SymbolTable[T]) GetByIndex(index int) *T {
	if symbol, ok := table.Symbols[table.OrderedNames[index]]; ok {
		return &symbol
	} else {
		return nil
	}
}

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
		overloads.WriteString(fmt.Sprintf("{parameters: (%s)%s, implemented: %v}", dt.JoinTypes(symbol.Parameters), priv, symbol.HasDefaultImplementation))
	}
	return fmt.Sprintf("{name: %s, returns: %s, overloads: [%s]}", fn.Name, fn.ReturnType, overloads.String())
}

func (variable VariableSymbol) String() string {
	priv := ""
	mut := ""
	off := ""
	if variable.isPrivate {
		priv = ", isPrivate: true"
	}
	if variable.isMutable {
		mut = ", isMutable: true"
	}
	if variable.Offset.IsSet {
		off = fmt.Sprintf(", Offset: %d", variable.Offset.Value)
	}
	return fmt.Sprintf("{name: %s, Type: %s%s%s%s, Initialized: %v}", variable.Name, variable.Type, priv, mut, off, variable.Initialized)
}

func (nb NamedBlockSymbol) String() string {
	return fmt.Sprintf("{name: %s}", nb.Name)
}

func (symbol FnCreateSymbol) getSignature() string {
	returns := ""
	if !symbol.returnType.Equals(dt.NoneType) {
		returns = fmt.Sprintf("->%s", symbol.returnType)
	}
	return fmt.Sprintf("fn %s(%s)%s", symbol.name, dt.JoinTypes(symbol.parameters), returns)
}

func (symbol FnCreateSymbol) toOverload(hasMatch bool) *FnOverloadSymbol {
	return &FnOverloadSymbol{
		Parameters:               symbol.parameters,
		ParameterNames:           symbol.parameterNames,
		HasDefaultImplementation: symbol.hasDefaultImplementation,
		IsPrivate:                symbol.isPrivate,
		Body:                     symbol.Body,
		InnerScope:               symbol.innerScope,
		IRName:                   symbol.getIRName(hasMatch),
	}
}

// getX()@TestName
// main()
// read scope id until ( -> "raw" function name
// if one @, treat as struct method
// if two @s, treat as struct, named block method

func getIRNamePrefix(scopeId string) string {
	prefix := strings.Split(scopeId, "(")[0]
	if strings.ContainsRune(scopeId, '@') {
		parts := strings.Split(scopeId, "@")
		for i := 1; i < len(parts); i++ {
			prefix = parts[i] + "_" + prefix // prepend the prefix
		}
		prefix = "__" + prefix
	}
	return prefix
}

func (symbol FnCreateSymbol) getIRName(hasMatch bool) string {
	if symbol.innerScope == nil {
		return ""
	}
	name := getIRNamePrefix(symbol.innerScope.Id)
	if !hasMatch {
		return name
	}
	params := strings.Builder{}
	for _, param := range symbol.parameters {
		params.WriteString("--")
		params.WriteString(param.String())
	}
	if len(symbol.parameters) == 0 {
		return name
	}
	return fmt.Sprintf("%s%s", name, params.String())
}

func (fn FunctionSymbol) GetMatchingOverload(params []dt.SourceType) *FnOverloadSymbol {
	count := len(params)
	for _, overload := range fn.Overloads {
		matches := false
		if count == len(overload.Parameters) {
			if count == 0 {
				return &overload
			}
			for i := range count {
				param := overload.Parameters[i]
				if params[i].Equals(param) || param.Equals(dt.AnyType) || ImplementsInterface(param, params[i]) {
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
	Local         VariableCtx = "local"
	Global        VariableCtx = "global"
	Param         VariableCtx = "param"
	PrimitiveProp VariableCtx = "primitive_property"
	StructProp    VariableCtx = "prop"
)

var (
	PrimitiveMembers PrimitiveTypeTables = PrimitiveTypeTables{
		dt.String: PrimitiveTypeMembers{
			Properties: SymbolTable[VariableSymbol]{
				OrderedNames: []string{
					"length",
				},
				Symbols: map[string]VariableSymbol{
					"length": {
						Name: "length",
						Type: dt.Int32Type,
						Ctx:  PrimitiveProp,
					},
				},
			},
			Methods: FunctionTable{
				OrderedNames: []string{
					"indexOf",
					"contains",
					"startsWith",
					"endsWith",
					"replace",
					"replaceAll",
					"reverse",
					"toUpper",
					"toLower",
					"trim",
					"trimStart",
					"trimEnd",
				},
				Symbols: map[string]FunctionSymbol{
					"indexOf": {
						Name: "indexOf",
						Overloads: []FnOverloadSymbol{{
							Parameters:               []dt.SourceType{dt.CharType},
							HasDefaultImplementation: true,
							IRName:                   "__str_indexOf",
						}},
						ReturnType: dt.Int32Type,
					},
					"contains": {
						Name: "contains",
						Overloads: []FnOverloadSymbol{
							{
								Parameters:               []dt.SourceType{dt.CharType},
								HasDefaultImplementation: true,
								IRName:                   "__str_contains_char",
							},
							{
								Parameters:               []dt.SourceType{dt.StringType},
								HasDefaultImplementation: true,
								IRName:                   "__str_contains_String",
							},
						},
						ReturnType: dt.BoolType,
					},
					"startsWith": {
						Name: "startsWith",
						Overloads: []FnOverloadSymbol{{
							Parameters:               []dt.SourceType{dt.StringType},
							HasDefaultImplementation: true,
							IRName:                   "__str_startsWith",
						}},
						ReturnType: dt.BoolType,
					},
					"endsWith": {
						Name: "endsWith",
						Overloads: []FnOverloadSymbol{{
							Parameters:               []dt.SourceType{dt.StringType},
							HasDefaultImplementation: true,
							IRName:                   "__str_endsWith",
						}},
						ReturnType: dt.BoolType,
					},
					"replace": {
						Name: "replace",
						Overloads: []FnOverloadSymbol{
							{
								Parameters:               []dt.SourceType{dt.StringType, dt.StringType},
								HasDefaultImplementation: true,
								IRName:                   "__str_replace_String_String",
							},
							{
								Parameters:               []dt.SourceType{dt.CharType, dt.CharType},
								HasDefaultImplementation: true,
								IRName:                   "__str_replace_char_char",
							},
						},
						ReturnType: dt.StringType,
					},
					"replaceAll": {
						Name: "replaceAll",
						Overloads: []FnOverloadSymbol{
							{
								Parameters:               []dt.SourceType{dt.StringType, dt.StringType},
								HasDefaultImplementation: true,
								IRName:                   "__str_replaceAll_String_String",
							},
							{
								Parameters:               []dt.SourceType{dt.CharType, dt.CharType},
								HasDefaultImplementation: true,
								IRName:                   "__str_replaceAll_char_char",
							},
						},
						ReturnType: dt.StringType,
					},
					"reverse": {
						Name: "reverse",
						Overloads: []FnOverloadSymbol{{
							Parameters:               []dt.SourceType{},
							HasDefaultImplementation: true,
							IRName:                   "__str_reverse",
						}},
						ReturnType: dt.StringType,
					},
					"toUpper": {
						Name: "toUpper",
						Overloads: []FnOverloadSymbol{{
							Parameters:               []dt.SourceType{},
							HasDefaultImplementation: true,
							IRName:                   "__str_toUpper",
						}},
						ReturnType: dt.StringType,
					},
					"toLower": {
						Name: "toLower",
						Overloads: []FnOverloadSymbol{{
							Parameters:               []dt.SourceType{},
							HasDefaultImplementation: true,
							IRName:                   "__str_toLower",
						}},
						ReturnType: dt.StringType,
					},
					"trim": {
						Name: "trim",
						Overloads: []FnOverloadSymbol{{
							Parameters:               []dt.SourceType{},
							HasDefaultImplementation: true,
							IRName:                   "__str_trim",
						}},
						ReturnType: dt.StringType,
					},
					"trimStart": {
						Name: "trimStart",
						Overloads: []FnOverloadSymbol{{
							Parameters:               []dt.SourceType{},
							HasDefaultImplementation: true,
							IRName:                   "__str_trimStart",
						}},
						ReturnType: dt.StringType,
					},
					"trimEnd": {
						Name: "trimEnd",
						Overloads: []FnOverloadSymbol{{
							Parameters:               []dt.SourceType{},
							HasDefaultImplementation: true,
							IRName:                   "__str_trimEnd",
						}},
						ReturnType: dt.StringType,
					},
				},
			},
		},
	}
)
