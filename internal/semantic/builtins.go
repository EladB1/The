package semantic

import dt "github.com/EladB1/The/internal/datatypes"

var rootScope *Scope = &Scope{
	Id:          "@built-in",
	Kind:        Default,
	Parent:      nil,
	Interfaces:  InterfaceSymbolTable{},
	Structs:     StructSymbolTable{},
	NamedBlocks: NamedBlockSymbolTable{},
	Functions: FunctionSymbolTable{
		"print": FunctionSymbol{
			Name: "print",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []dt.SourceType{dt.AnyType},
				HasDefaultImplementation: true,
			}},
			ReturnType: dt.NoneType,
		},
		"println": FunctionSymbol{
			Name: "println",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []dt.SourceType{dt.AnyType},
				HasDefaultImplementation: true,
			}},
			ReturnType: dt.NoneType,
		},
		"printerr": FunctionSymbol{
			Name: "printerr",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []dt.SourceType{dt.AnyType},
				HasDefaultImplementation: true,
			}},
			ReturnType: dt.NoneType,
		},
		"typeOf": FunctionSymbol{
			Name: "typeOf",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []dt.SourceType{dt.AnyType},
				HasDefaultImplementation: true,
			}},
			ReturnType: dt.StringType,
		},
		"exit": FunctionSymbol{
			Name: "exit",
			Overloads: []FnOverloadSymbol{
				{
					Parameters:               []dt.SourceType{dt.Int32Type},
					HasDefaultImplementation: true,
				},
				{
					Parameters:               []dt.SourceType{dt.Int32Type, dt.StringType},
					HasDefaultImplementation: true,
				},
			},
			ReturnType: dt.NoneType,
		},
		"sleep": FunctionSymbol{
			Name: "sleep",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []dt.SourceType{dt.DoubleType},
				HasDefaultImplementation: true,
			}},
			ReturnType: dt.NoneType,
		},
		"getEnv": FunctionSymbol{
			Name: "getEnv",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []dt.SourceType{dt.StringType},
				HasDefaultImplementation: true,
			}},
			ReturnType: dt.StringType,
		},
		"setEnv": FunctionSymbol{
			Name: "setEnv",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []dt.SourceType{dt.StringType, dt.StringType},
				HasDefaultImplementation: true,
			}},
			ReturnType: dt.NoneType,
		},
		"indexOf": FunctionSymbol{
			Name: "indexOf",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []dt.SourceType{dt.StringType, dt.CharType},
				HasDefaultImplementation: true,
			}},
			ReturnType: dt.Int32Type,
		},
		"contains": FunctionSymbol{
			Name: "contains",
			Overloads: []FnOverloadSymbol{
				{
					Parameters:               []dt.SourceType{dt.StringType, dt.CharType},
					HasDefaultImplementation: true,
				},
				{
					Parameters:               []dt.SourceType{dt.StringType, dt.StringType},
					HasDefaultImplementation: true,
				},
			},
			ReturnType: dt.BoolType,
		},
		"startsWith": FunctionSymbol{
			Name: "startsWith",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []dt.SourceType{dt.StringType, dt.StringType},
				HasDefaultImplementation: true,
			}},
			ReturnType: dt.BoolType,
		},
		"endsWith": FunctionSymbol{
			Name: "endsWith",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []dt.SourceType{dt.StringType, dt.StringType},
				HasDefaultImplementation: true,
			}},
			ReturnType: dt.BoolType,
		},
		"replace": FunctionSymbol{
			Name: "replace",
			Overloads: []FnOverloadSymbol{
				{
					Parameters:               []dt.SourceType{dt.StringType, dt.StringType, dt.StringType},
					HasDefaultImplementation: true,
				},
				{
					Parameters:               []dt.SourceType{dt.StringType, dt.CharType, dt.CharType},
					HasDefaultImplementation: true,
				},
			},
			ReturnType: dt.StringType,
		},
		"replaceAll": FunctionSymbol{
			Name: "replace",
			Overloads: []FnOverloadSymbol{
				{
					Parameters:               []dt.SourceType{dt.StringType, dt.StringType, dt.StringType},
					HasDefaultImplementation: true,
				},
				{
					Parameters:               []dt.SourceType{dt.StringType, dt.CharType, dt.CharType},
					HasDefaultImplementation: true,
				},
			},
			ReturnType: dt.StringType,
		},
		"reverse": FunctionSymbol{
			Name: "reverse",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []dt.SourceType{dt.StringType},
				HasDefaultImplementation: true,
			}},
			ReturnType: dt.StringType,
		},
		"toUpper": FunctionSymbol{
			Name: "toUpper",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []dt.SourceType{dt.StringType},
				HasDefaultImplementation: true,
			}},
			ReturnType: dt.StringType,
		},
		"toLower": FunctionSymbol{
			Name: "toLower",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []dt.SourceType{dt.StringType},
				HasDefaultImplementation: true,
			}},
			ReturnType: dt.StringType,
		},
		"trim": FunctionSymbol{
			Name: "trim",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []dt.SourceType{dt.StringType},
				HasDefaultImplementation: true,
			}},
			ReturnType: dt.StringType,
		},
		"trimStart": FunctionSymbol{
			Name: "trimStart",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []dt.SourceType{dt.StringType},
				HasDefaultImplementation: true,
			}},
			ReturnType: dt.StringType,
		},
		"trimEnd": FunctionSymbol{
			Name: "trimEnd",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []dt.SourceType{dt.StringType},
				HasDefaultImplementation: true,
			}},
			ReturnType: dt.StringType,
		},
		"assert": FunctionSymbol{
			Name: "assert",
			Overloads: []FnOverloadSymbol{
				{
					Parameters:               []dt.SourceType{dt.BoolType},
					HasDefaultImplementation: true,
				},
				{
					Parameters:               []dt.SourceType{dt.BoolType, dt.StringType},
					HasDefaultImplementation: true,
				},
			},
			ReturnType: dt.NoneType,
		},
		"prompt": FunctionSymbol{
			Name: "prompt",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []dt.SourceType{dt.StringType},
				HasDefaultImplementation: true,
			}},
			ReturnType: dt.StringType,
		},
		"secretPrompt": FunctionSymbol{
			Name: "secretPrompt",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []dt.SourceType{dt.StringType},
				HasDefaultImplementation: true,
			}},
			ReturnType: dt.StringType,
		},
	},
	Variables: VariableSymbolTable{
		"INT_MIN": VariableSymbol{
			Name: "INT_MIN",
			Type: dt.Int32Type,
			Ctx:  Global,
		},
		"INT_MAX": VariableSymbol{
			Name: "INT_MAX",
			Type: dt.Int32Type,
			Ctx:  Global,
		},
		"INT64_MIN": VariableSymbol{
			Name: "INT64_MIN",
			Type: dt.Int64Type,
			Ctx:  Global,
		},
		"INT64_MAX": VariableSymbol{
			Name: "INT64_MAX",
			Type: dt.Int64Type,
			Ctx:  Global,
		},
		"UINT32_MAX": VariableSymbol{
			Name: "UINT32_MAX",
			Type: dt.Uint32Type,
			Ctx:  Global,
		},
		"UINT64_MAX": VariableSymbol{
			Name: "UINT64_MAX",
			Type: dt.Uint64Type,
			Ctx:  Global,
		},
		"FLOAT_MIN": VariableSymbol{
			Name: "FLOAT_MIN",
			Type: dt.FloatType,
			Ctx:  Global,
		},
		"FLOAT_MIN_POSITIVE": VariableSymbol{
			Name: "FLOAT_MIN_POSITIVE",
			Type: dt.FloatType,
			Ctx:  Global,
		},
		"FLOAT_MAX": VariableSymbol{
			Name: "FLOAT_MAX",
			Type: dt.FloatType,
			Ctx:  Global,
		},
		"FLOAT_EPSILON": VariableSymbol{
			Name: "FLOAT_EPSILON",
			Type: dt.FloatType,
			Ctx:  Global,
		},
		"FLOAT_NaN": VariableSymbol{
			Name: "FLOAT_NaN",
			Type: dt.FloatType,
			Ctx:  Global,
		},
		"FLOAT_INF": VariableSymbol{
			Name: "FLOAT_INF",
			Type: dt.FloatType,
			Ctx:  Global,
		},
		"FLOAT_NEG_INF": VariableSymbol{
			Name: "FLOAT_NEG_INF",
			Type: dt.FloatType,
			Ctx:  Global,
		},
		"DOUBLE_MIN": VariableSymbol{
			Name: "DOUBLE_MIN",
			Type: dt.DoubleType,
			Ctx:  Global,
		},
		"DOUBLE_MIN_POSITIVE": VariableSymbol{
			Name: "DOUBLE_MIN_POSITIVE",
			Type: dt.DoubleType,
			Ctx:  Global,
		},
		"DOUBLE_MAX": VariableSymbol{
			Name: "DOUBLE_MAX",
			Type: dt.DoubleType,
			Ctx:  Global,
		},
		"DOUBLE_EPSILON": VariableSymbol{
			Name: "DOUBLE_EPSILON",
			Type: dt.DoubleType,
			Ctx:  Global,
		},
		"DOUBLE_NaN": VariableSymbol{
			Name: "DOUBLE_NaN",
			Type: dt.DoubleType,
			Ctx:  Global,
		},
		"DOUBLE_INF": VariableSymbol{
			Name: "DOUBLE_INF",
			Type: dt.DoubleType,
			Ctx:  Global,
		},
		"DOUBLE_NEG_INF": VariableSymbol{
			Name: "DOUBLE_NEG_INF",
			Type: dt.DoubleType,
			Ctx:  Global,
		},
		"PI": VariableSymbol{
			Name: "PI",
			Type: dt.DoubleType,
			Ctx:  Global,
		},
		"E": VariableSymbol{
			Name: "E",
			Type: dt.DoubleType,
			Ctx:  Global,
		},
	},
}
