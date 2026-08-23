package semantic

import dt "github.com/EladB1/The/internal/datatypes"

var rootScope *Scope = &Scope{
	Id:          "@built-in",
	Kind:        Default,
	Parent:      nil,
	Interfaces:  NewTable[InterfaceSymbol](),
	Structs:     NewTable[StructSymbol](),
	NamedBlocks: NewTable[NamedBlockSymbol](),
	Functions: &SymbolTable[FunctionSymbol]{
		OrderedNames: []string{
			"print",
			"println",
			"printerr",
			"exit",
			"assert",
			"prompt",
			"typeOf",
			"sleep",
			"getEnv",
			"setEnv",
			"secretPrompt",
		},
		Symbols: map[string]FunctionSymbol{
			"print": {
				Name: "print",
				Overloads: []FnOverloadSymbol{{
					Parameters:               []dt.SourceType{dt.AnyType},
					HasDefaultImplementation: true,
				}},
				ReturnType: dt.NoneType,
			},
			"println": {
				Name: "println",
				Overloads: []FnOverloadSymbol{{
					Parameters:               []dt.SourceType{dt.AnyType},
					HasDefaultImplementation: true,
				}},
				ReturnType: dt.NoneType,
			},
			"printerr": {
				Name: "printerr",
				Overloads: []FnOverloadSymbol{{
					Parameters:               []dt.SourceType{dt.AnyType},
					HasDefaultImplementation: true,
				}},
				ReturnType: dt.NoneType,
			},
			"exit": {
				Name: "exit",
				Overloads: []FnOverloadSymbol{
					{
						Parameters:               []dt.SourceType{dt.Int32Type},
						HasDefaultImplementation: true,
						IRName:                   "exit_int",
					},
					{
						Parameters:               []dt.SourceType{dt.Int32Type, dt.StringType},
						HasDefaultImplementation: true,
						IRName:                   "exit_int_String",
					},
				},
				ReturnType: dt.NoneType,
			},
			"assert": {
				Name: "assert",
				Overloads: []FnOverloadSymbol{
					{
						Parameters:               []dt.SourceType{dt.BoolType},
						HasDefaultImplementation: true,
						IRName:                   "assert_bool",
					},
					{
						Parameters:               []dt.SourceType{dt.BoolType, dt.StringType},
						HasDefaultImplementation: true,
						IRName:                   "assert_bool_String",
					},
				},
				ReturnType: dt.NoneType,
			},
			"prompt": {
				Name: "prompt",
				Overloads: []FnOverloadSymbol{{
					Parameters:               []dt.SourceType{dt.StringType},
					HasDefaultImplementation: true,
				}},
				ReturnType: dt.StringType,
			},
			"typeOf": {
				Name: "typeOf",
				Overloads: []FnOverloadSymbol{{
					Parameters:               []dt.SourceType{dt.AnyType},
					HasDefaultImplementation: true,
				}},
				ReturnType: dt.StringType,
			},

			"sleep": {
				Name: "sleep",
				Overloads: []FnOverloadSymbol{{
					Parameters:               []dt.SourceType{dt.DoubleType},
					HasDefaultImplementation: true,
				}},
				ReturnType: dt.NoneType,
			},
			"getEnv": {
				Name: "getEnv",
				Overloads: []FnOverloadSymbol{{
					Parameters:               []dt.SourceType{dt.StringType},
					HasDefaultImplementation: true,
				}},
				ReturnType: dt.StringType,
			},
			"setEnv": {
				Name: "setEnv",
				Overloads: []FnOverloadSymbol{{
					Parameters:               []dt.SourceType{dt.StringType, dt.StringType},
					HasDefaultImplementation: true,
				}},
				ReturnType: dt.NoneType,
			},

			"secretPrompt": {
				Name: "secretPrompt",
				Overloads: []FnOverloadSymbol{{
					Parameters:               []dt.SourceType{dt.StringType},
					HasDefaultImplementation: true,
				}},
				ReturnType: dt.StringType,
			},
		},
	},
	Variables: &SymbolTable[VariableSymbol]{
		OrderedNames: []string{
			"INT_MIN",
			"INT_MAX",
			"INT64_MIN",
			"INT64_MAX",
			"UINT32_MAX",
			"UINT64_MAX",
			"FLOAT_MIN",
			"FLOAT_MIN_POSITIVE",
			"FLOAT_MAX",
			"FLOAT_EPSILON",
			"FLOAT_NaN",
			"FLOAT_INF",
			"FLOAT_NEG_INF",
			"DOUBLE_MIN",
			"DOUBLE_MIN_POSITIVE",
			"DOUBLE_MAX",
			"DOUBLE_EPSILON",
			"DOUBLE_NaN",
			"DOUBLE_INF",
			"DOUBLE_NEG_INF",
			"PI",
			"E",
		},
		Symbols: map[string]VariableSymbol{
			"INT_MIN": {
				Name: "INT_MIN",
				Type: dt.Int32Type,
				Ctx:  Global,
			},
			"INT_MAX": {
				Name: "INT_MAX",
				Type: dt.Int32Type,
				Ctx:  Global,
			},
			"INT64_MIN": {
				Name: "INT64_MIN",
				Type: dt.Int64Type,
				Ctx:  Global,
			},
			"INT64_MAX": {
				Name: "INT64_MAX",
				Type: dt.Int64Type,
				Ctx:  Global,
			},
			"UINT32_MAX": {
				Name: "UINT32_MAX",
				Type: dt.Uint32Type,
				Ctx:  Global,
			},
			"UINT64_MAX": {
				Name: "UINT64_MAX",
				Type: dt.Uint64Type,
				Ctx:  Global,
			},
			"FLOAT_MIN": {
				Name: "FLOAT_MIN",
				Type: dt.FloatType,
				Ctx:  Global,
			},
			"FLOAT_MIN_POSITIVE": {
				Name: "FLOAT_MIN_POSITIVE",
				Type: dt.FloatType,
				Ctx:  Global,
			},
			"FLOAT_MAX": {
				Name: "FLOAT_MAX",
				Type: dt.FloatType,
				Ctx:  Global,
			},
			"FLOAT_EPSILON": {
				Name: "FLOAT_EPSILON",
				Type: dt.FloatType,
				Ctx:  Global,
			},
			"FLOAT_NaN": {
				Name: "FLOAT_NaN",
				Type: dt.FloatType,
				Ctx:  Global,
			},
			"FLOAT_INF": {
				Name: "FLOAT_INF",
				Type: dt.FloatType,
				Ctx:  Global,
			},
			"FLOAT_NEG_INF": {
				Name: "FLOAT_NEG_INF",
				Type: dt.FloatType,
				Ctx:  Global,
			},
			"DOUBLE_MIN": {
				Name: "DOUBLE_MIN",
				Type: dt.DoubleType,
				Ctx:  Global,
			},
			"DOUBLE_MIN_POSITIVE": {
				Name: "DOUBLE_MIN_POSITIVE",
				Type: dt.DoubleType,
				Ctx:  Global,
			},
			"DOUBLE_MAX": {
				Name: "DOUBLE_MAX",
				Type: dt.DoubleType,
				Ctx:  Global,
			},
			"DOUBLE_EPSILON": {
				Name: "DOUBLE_EPSILON",
				Type: dt.DoubleType,
				Ctx:  Global,
			},
			"DOUBLE_NaN": {
				Name: "DOUBLE_NaN",
				Type: dt.DoubleType,
				Ctx:  Global,
			},
			"DOUBLE_INF": {
				Name: "DOUBLE_INF",
				Type: dt.DoubleType,
				Ctx:  Global,
			},
			"DOUBLE_NEG_INF": {
				Name: "DOUBLE_NEG_INF",
				Type: dt.DoubleType,
				Ctx:  Global,
			},
			"PI": {
				Name: "PI",
				Type: dt.DoubleType,
				Ctx:  Global,
			},
			"E": {
				Name: "E",
				Type: dt.DoubleType,
				Ctx:  Global,
			},
		},
	},
}
