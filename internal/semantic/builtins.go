package semantic

import "github.com/EladB1/The/internal/datatypes"

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
				Parameters:               []datatypes.DataType{datatypes.Any},
				HasDefaultImplementation: true,
			}},
			ReturnType: datatypes.None,
		},
		"println": FunctionSymbol{
			Name: "println",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []datatypes.DataType{datatypes.Any},
				HasDefaultImplementation: true,
			}},
			ReturnType: datatypes.None,
		},
		"printerr": FunctionSymbol{
			Name: "printerr",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []datatypes.DataType{datatypes.Any},
				HasDefaultImplementation: true,
			}},
			ReturnType: datatypes.None,
		},
		"typeOf": FunctionSymbol{
			Name: "typeOf",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []datatypes.DataType{datatypes.Any},
				HasDefaultImplementation: true,
			}},
			ReturnType: datatypes.String,
		},
		"exit": FunctionSymbol{
			Name: "exit",
			Overloads: []FnOverloadSymbol{
				{
					Parameters:               []datatypes.DataType{datatypes.Int32},
					HasDefaultImplementation: true,
				},
				{
					Parameters:               []datatypes.DataType{datatypes.Int32, datatypes.String},
					HasDefaultImplementation: true,
				},
			},
			ReturnType: datatypes.None,
		},
		"sleep": FunctionSymbol{
			Name: "sleep",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []datatypes.DataType{datatypes.Double},
				HasDefaultImplementation: true,
			}},
			ReturnType: datatypes.None,
		},
		"getEnv": FunctionSymbol{
			Name: "getEnv",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []datatypes.DataType{datatypes.String},
				HasDefaultImplementation: true,
			}},
			ReturnType: datatypes.String,
		},
		"setEnv": FunctionSymbol{
			Name: "setEnv",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []datatypes.DataType{datatypes.String, datatypes.String},
				HasDefaultImplementation: true,
			}},
			ReturnType: datatypes.None,
		},
		"indexOf": FunctionSymbol{
			Name: "indexOf",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []datatypes.DataType{datatypes.String, datatypes.Char},
				HasDefaultImplementation: true,
			}},
			ReturnType: datatypes.Int32,
		},
		"contains": FunctionSymbol{
			Name: "contains",
			Overloads: []FnOverloadSymbol{
				{
					Parameters:               []datatypes.DataType{datatypes.String, datatypes.Char},
					HasDefaultImplementation: true,
				},
				{
					Parameters:               []datatypes.DataType{datatypes.String, datatypes.String},
					HasDefaultImplementation: true,
				},
			},
			ReturnType: datatypes.Bool,
		},
		"startsWith": FunctionSymbol{
			Name: "startsWith",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []datatypes.DataType{datatypes.String, datatypes.String},
				HasDefaultImplementation: true,
			}},
			ReturnType: datatypes.Bool,
		},
		"endsWith": FunctionSymbol{
			Name: "endsWith",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []datatypes.DataType{datatypes.String, datatypes.String},
				HasDefaultImplementation: true,
			}},
			ReturnType: datatypes.Bool,
		},
		"replace": FunctionSymbol{
			Name: "replace",
			Overloads: []FnOverloadSymbol{
				{
					Parameters:               []datatypes.DataType{datatypes.String, datatypes.String, datatypes.String},
					HasDefaultImplementation: true,
				},
				{
					Parameters:               []datatypes.DataType{datatypes.String, datatypes.Char, datatypes.Char},
					HasDefaultImplementation: true,
				},
			},
			ReturnType: datatypes.String,
		},
		"replaceAll": FunctionSymbol{
			Name: "replace",
			Overloads: []FnOverloadSymbol{
				{
					Parameters:               []datatypes.DataType{datatypes.String, datatypes.String, datatypes.String},
					HasDefaultImplementation: true,
				},
				{
					Parameters:               []datatypes.DataType{datatypes.String, datatypes.Char, datatypes.Char},
					HasDefaultImplementation: true,
				},
			},
			ReturnType: datatypes.String,
		},
		"reverse": FunctionSymbol{
			Name: "reverse",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []datatypes.DataType{datatypes.String},
				HasDefaultImplementation: true,
			}},
			ReturnType: datatypes.String,
		},
		"toUpper": FunctionSymbol{
			Name: "toUpper",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []datatypes.DataType{datatypes.String},
				HasDefaultImplementation: true,
			}},
			ReturnType: datatypes.String,
		},
		"toLower": FunctionSymbol{
			Name: "toLower",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []datatypes.DataType{datatypes.String},
				HasDefaultImplementation: true,
			}},
			ReturnType: datatypes.String,
		},
		"trim": FunctionSymbol{
			Name: "trim",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []datatypes.DataType{datatypes.String},
				HasDefaultImplementation: true,
			}},
			ReturnType: datatypes.String,
		},
		"trimStart": FunctionSymbol{
			Name: "trimStart",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []datatypes.DataType{datatypes.String},
				HasDefaultImplementation: true,
			}},
			ReturnType: datatypes.String,
		},
		"trimEnd": FunctionSymbol{
			Name: "trimEnd",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []datatypes.DataType{datatypes.String},
				HasDefaultImplementation: true,
			}},
			ReturnType: datatypes.String,
		},
		"assert": FunctionSymbol{
			Name: "assert",
			Overloads: []FnOverloadSymbol{
				{
					Parameters:               []datatypes.DataType{datatypes.Bool},
					HasDefaultImplementation: true,
				},
				{
					Parameters:               []datatypes.DataType{datatypes.Bool, datatypes.String},
					HasDefaultImplementation: true,
				},
			},
			ReturnType: datatypes.None,
		},
		"prompt": FunctionSymbol{
			Name: "prompt",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []datatypes.DataType{datatypes.String},
				HasDefaultImplementation: true,
			}},
			ReturnType: datatypes.String,
		},
		"secretPrompt": FunctionSymbol{
			Name: "secretPrompt",
			Overloads: []FnOverloadSymbol{{
				Parameters:               []datatypes.DataType{datatypes.String},
				HasDefaultImplementation: true,
			}},
			ReturnType: datatypes.String,
		},
	},
	Variables: VariableSymbolTable{
		"INT_MIN": VariableSymbol{
			Name: "INT_MIN",
			Type: datatypes.Int32,
		},
		"INT_MAX": VariableSymbol{
			Name: "INT_MAX",
			Type: datatypes.Int32,
		},
		"INT64_MIN": VariableSymbol{
			Name: "INT64_MIN",
			Type: datatypes.Int64,
		},
		"INT64_MAX": VariableSymbol{
			Name: "INT64_MAX",
			Type: datatypes.Int64,
		},
		"UINT32_MAX": VariableSymbol{
			Name: "UINT32_MAX",
			Type: datatypes.Uint32,
		},
		"UINT64_MAX": VariableSymbol{
			Name: "UINT64_MAX",
			Type: datatypes.Uint64,
		},
		"FLOAT_MIN": VariableSymbol{
			Name: "FLOAT_MIN",
			Type: datatypes.Float,
		},
		"FLOAT_MIN_POSITIVE": VariableSymbol{
			Name: "FLOAT_MIN_POSITIVE",
			Type: datatypes.Float,
		},
		"FLOAT_MAX": VariableSymbol{
			Name: "FLOAT_MAX",
			Type: datatypes.Float,
		},
		"FLOAT_EPSILON": VariableSymbol{
			Name: "FLOAT_EPSILON",
			Type: datatypes.Float,
		},
		"FLOAT_NaN": VariableSymbol{
			Name: "FLOAT_NaN",
			Type: datatypes.Float,
		},
		"FLOAT_INF": VariableSymbol{
			Name: "FLOAT_INF",
			Type: datatypes.Float,
		},
		"FLOAT_NEG_INF": VariableSymbol{
			Name: "FLOAT_NEG_INF",
			Type: datatypes.Float,
		},
		"DOUBLE_MIN": VariableSymbol{
			Name: "DOUBLE_MIN",
			Type: datatypes.Double,
		},
		"DOUBLE_MIN_POSITIVE": VariableSymbol{
			Name: "DOUBLE_MIN_POSITIVE",
			Type: datatypes.Double,
		},
		"DOUBLE_MAX": VariableSymbol{
			Name: "DOUBLE_MAX",
			Type: datatypes.Double,
		},
		"DOUBLE_EPSILON": VariableSymbol{
			Name: "DOUBLE_EPSILON",
			Type: datatypes.Double,
		},
		"DOUBLE_NaN": VariableSymbol{
			Name: "DOUBLE_NaN",
			Type: datatypes.Double,
		},
		"DOUBLE_INF": VariableSymbol{
			Name: "DOUBLE_INF",
			Type: datatypes.Double,
		},
		"DOUBLE_NEG_INF": VariableSymbol{
			Name: "DOUBLE_NEG_INF",
			Type: datatypes.Double,
		},
		"PI": VariableSymbol{
			Name: "PI",
			Type: datatypes.Double,
		},
		"E": VariableSymbol{
			Name: "E",
			Type: datatypes.Double,
		},
	},
}
