package semantic

import "github.com/EladB1/The/internal/datatypes"

var builtinScope Scope = Scope{
	id:          "@built-in",
	parent:      nil,
	interfaces:  InterfaceSymbolTable{},
	structs:     StructSymbolTable{},
	namedBlocks: NamedBlockSymbolTable{},
	functions: FunctionSymbolTable{
		"print": FunctionSymbol{
			name:       "print",
			parameters: []datatypes.DataType{datatypes.Any},
			returnType: datatypes.None,
		},
		"println": FunctionSymbol{
			name:       "println",
			parameters: []datatypes.DataType{datatypes.Any},
			returnType: datatypes.None,
		},
		"printerr": FunctionSymbol{
			name:       "printerr",
			parameters: []datatypes.DataType{datatypes.Any},
			returnType: datatypes.None,
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
		"FLOAT_MIN": VariableSymbol{
			name: "FLOAT_MIN",
			Type: datatypes.Float,
		},
		"FLOAT_MAX": VariableSymbol{
			name: "FLOAT_MAX",
			Type: datatypes.Float,
		},
		"DOUBLE_MIN": VariableSymbol{
			name: "DOUBLE_MIN",
			Type: datatypes.Double,
		},
		"DOUBLE_MAX": VariableSymbol{
			name: "DOUBLE_MAX",
			Type: datatypes.Double,
		},
		"UINT32_MAX": VariableSymbol{
			name: "UINT32_MAX",
			Type: datatypes.Uint32,
		},
		"UINT64_MAX": VariableSymbol{
			name: "UINT64_MAX",
			Type: datatypes.Uint64,
		},
	},
}
