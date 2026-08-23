package semantic

import (
	ds "github.com/EladB1/The/internal/datastructures"
	dt "github.com/EladB1/The/internal/datatypes"
)

type (
	PrimitiveTypeMembers struct {
		Properties ds.OrderedMap[VariableSymbol]
		Methods    ds.OrderedMap[FunctionSymbol]
	}

	PrimitiveTypeTables map[dt.DataType]PrimitiveTypeMembers
)

var (
	PrimitiveMembers PrimitiveTypeTables = PrimitiveTypeTables{
		dt.String: PrimitiveTypeMembers{
			Properties: ds.OrderedMap[VariableSymbol]{
				OrderedKeys: []string{"length"},
				Values: map[string]VariableSymbol{
					"length": {
						Name: "length",
						Type: dt.Int32Type,
						Ctx:  PrimitiveProp,
					},
				},
			},
			Methods: ds.OrderedMap[FunctionSymbol]{
				OrderedKeys: []string{
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
				Values: map[string]FunctionSymbol{
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
