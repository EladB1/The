package semantic

import "github.com/EladB1/The/internal/datatypes"

var rootScope *Scope = &Scope{
	id:          "@built-in",
	kind:        Default,
	parent:      nil,
	interfaces:  InterfaceSymbolTable{},
	structs:     StructSymbolTable{},
	namedBlocks: NamedBlockSymbolTable{},
	functions: FunctionSymbolTable{
		"print": FunctionSymbol{
			name: "print",
			overloads: []FnOverloadSymbol{{
				parameters: []datatypes.DataType{datatypes.Any},
			}},
			returnType: datatypes.None,
		},
		"println": FunctionSymbol{
			name: "println",
			overloads: []FnOverloadSymbol{{
				parameters: []datatypes.DataType{datatypes.Any},
			}},
			returnType: datatypes.None,
		},
		"printerr": FunctionSymbol{
			name: "printerr",
			overloads: []FnOverloadSymbol{{
				parameters: []datatypes.DataType{datatypes.Any},
			}},
			returnType: datatypes.None,
		},
		"typeOf": FunctionSymbol{
			name: "typeOf",
			overloads: []FnOverloadSymbol{{
				parameters: []datatypes.DataType{datatypes.Any},
			}},
			returnType: datatypes.String,
		},
		"exit": FunctionSymbol{
			name: "exit",
			overloads: []FnOverloadSymbol{
				{
					parameters: []datatypes.DataType{datatypes.Int32},
				},
				{
					parameters: []datatypes.DataType{datatypes.Int32, datatypes.String},
				},
			},
			returnType: datatypes.None,
		},
		"sleep": FunctionSymbol{
			name: "sleep",
			overloads: []FnOverloadSymbol{{
				parameters: []datatypes.DataType{datatypes.Double},
			}},
			returnType: datatypes.None,
		},
		"getEnv": FunctionSymbol{
			name: "getEnv",
			overloads: []FnOverloadSymbol{{
				parameters: []datatypes.DataType{datatypes.String},
			}},
			returnType: datatypes.String,
		},
		"setEnv": FunctionSymbol{
			name: "setEnv",
			overloads: []FnOverloadSymbol{{
				parameters: []datatypes.DataType{datatypes.String, datatypes.String},
			}},
			returnType: datatypes.None,
		},
		"indexOf": FunctionSymbol{
			name: "indexOf",
			overloads: []FnOverloadSymbol{{
				parameters: []datatypes.DataType{datatypes.String, datatypes.Char},
			}},
			returnType: datatypes.Int32,
		},
		"contains": FunctionSymbol{
			name: "contains",
			overloads: []FnOverloadSymbol{
				{
					parameters: []datatypes.DataType{datatypes.String, datatypes.Char},
				},
				{
					parameters: []datatypes.DataType{datatypes.String, datatypes.String},
				},
			},
			returnType: datatypes.Bool,
		},
		"startsWith": FunctionSymbol{
			name: "startsWith",
			overloads: []FnOverloadSymbol{{
				parameters: []datatypes.DataType{datatypes.String, datatypes.String},
			}},
			returnType: datatypes.Bool,
		},
		"endsWith": FunctionSymbol{
			name: "endsWith",
			overloads: []FnOverloadSymbol{{
				parameters: []datatypes.DataType{datatypes.String, datatypes.String},
			}},
			returnType: datatypes.Bool,
		},
		"replace": FunctionSymbol{
			name: "replace",
			overloads: []FnOverloadSymbol{
				{
					parameters: []datatypes.DataType{datatypes.String, datatypes.String, datatypes.String},
				},
				{
					parameters: []datatypes.DataType{datatypes.String, datatypes.Char, datatypes.Char},
				},
			},
			returnType: datatypes.String,
		},
		"replaceAll": FunctionSymbol{
			name: "replace",
			overloads: []FnOverloadSymbol{
				{
					parameters: []datatypes.DataType{datatypes.String, datatypes.String, datatypes.String},
				},
				{
					parameters: []datatypes.DataType{datatypes.String, datatypes.Char, datatypes.Char},
				},
			},
			returnType: datatypes.String,
		},
		"reverse": FunctionSymbol{
			name: "reverse",
			overloads: []FnOverloadSymbol{{
				parameters: []datatypes.DataType{datatypes.String},
			}},
			returnType: datatypes.String,
		},
		"toUpper": FunctionSymbol{
			name: "toUpper",
			overloads: []FnOverloadSymbol{{
				parameters: []datatypes.DataType{datatypes.String},
			}},
			returnType: datatypes.String,
		},
		"toLower": FunctionSymbol{
			name: "toLower",
			overloads: []FnOverloadSymbol{{
				parameters: []datatypes.DataType{datatypes.String},
			}},
			returnType: datatypes.String,
		},
		"trim": FunctionSymbol{
			name: "trim",
			overloads: []FnOverloadSymbol{{
				parameters: []datatypes.DataType{datatypes.String},
			}},
			returnType: datatypes.String,
		},
		"trimStart": FunctionSymbol{
			name: "trimStart",
			overloads: []FnOverloadSymbol{{
				parameters: []datatypes.DataType{datatypes.String},
			}},
			returnType: datatypes.String,
		},
		"trimEnd": FunctionSymbol{
			name: "trimEnd",
			overloads: []FnOverloadSymbol{{
				parameters: []datatypes.DataType{datatypes.String},
			}},
			returnType: datatypes.String,
		},
		"assert": FunctionSymbol{
			name: "assert",
			overloads: []FnOverloadSymbol{
				{
					parameters: []datatypes.DataType{datatypes.Bool},
				},
				{
					parameters: []datatypes.DataType{datatypes.Bool, datatypes.String},
				},
			},
			returnType: datatypes.None,
		},
		"prompt": FunctionSymbol{
			name: "prompt",
			overloads: []FnOverloadSymbol{{
				parameters: []datatypes.DataType{datatypes.String},
			}},
			returnType: datatypes.String,
		},
		"secretPrompt": FunctionSymbol{
			name: "secretPrompt",
			overloads: []FnOverloadSymbol{{
				parameters: []datatypes.DataType{datatypes.String},
			}},
			returnType: datatypes.String,
		},
	},
	variables: VariableSymbolTable{
		"INT_MIN": VariableSymbol{
			name: "INT_MIN",
			Type: datatypes.Int32,
		},
		"INT_MAX": VariableSymbol{
			name: "INT_MAX",
			Type: datatypes.Int32,
		},
		"INT64_MIN": VariableSymbol{
			name: "INT64_MIN",
			Type: datatypes.Int64,
		},
		"INT64_MAX": VariableSymbol{
			name: "INT64_MAX",
			Type: datatypes.Int64,
		},
		"UINT32_MAX": VariableSymbol{
			name: "UINT32_MAX",
			Type: datatypes.Uint32,
		},
		"UINT64_MAX": VariableSymbol{
			name: "UINT64_MAX",
			Type: datatypes.Uint64,
		},
		"FLOAT_MIN": VariableSymbol{
			name: "FLOAT_MIN",
			Type: datatypes.Float,
		},
		"FLOAT_MIN_POSITIVE": VariableSymbol{
			name: "FLOAT_MIN_POSITIVE",
			Type: datatypes.Float,
		},
		"FLOAT_MAX": VariableSymbol{
			name: "FLOAT_MAX",
			Type: datatypes.Float,
		},
		"FLOAT_EPSILON": VariableSymbol{
			name: "FLOAT_EPSILON",
			Type: datatypes.Float,
		},
		"FLOAT_NaN": VariableSymbol{
			name: "FLOAT_NaN",
			Type: datatypes.Float,
		},
		"FLOAT_INF": VariableSymbol{
			name: "FLOAT_INF",
			Type: datatypes.Float,
		},
		"FLOAT_NEG_INF": VariableSymbol{
			name: "FLOAT_NEG_INF",
			Type: datatypes.Float,
		},
		"DOUBLE_MIN": VariableSymbol{
			name: "DOUBLE_MIN",
			Type: datatypes.Double,
		},
		"DOUBLE_MIN_POSITIVE": VariableSymbol{
			name: "DOUBLE_MIN_POSITIVE",
			Type: datatypes.Double,
		},
		"DOUBLE_MAX": VariableSymbol{
			name: "DOUBLE_MAX",
			Type: datatypes.Double,
		},
		"DOUBLE_EPSILON": VariableSymbol{
			name: "DOUBLE_EPSILON",
			Type: datatypes.Double,
		},
		"DOUBLE_NaN": VariableSymbol{
			name: "DOUBLE_NaN",
			Type: datatypes.Double,
		},
		"DOUBLE_INF": VariableSymbol{
			name: "DOUBLE_INF",
			Type: datatypes.Double,
		},
		"DOUBLE_NEG_INF": VariableSymbol{
			name: "DOUBLE_NEG_INF",
			Type: datatypes.Double,
		},
		"PI": VariableSymbol{
			name: "PI",
			Type: datatypes.Double,
		},
		"E": VariableSymbol{
			name: "E",
			Type: datatypes.Double,
		},
	},
}
