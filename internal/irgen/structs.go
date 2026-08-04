package irgen

import (
	"fmt"

	ds "github.com/EladB1/The/internal/datastructures"
	dt "github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/parser"
	"github.com/EladB1/The/internal/semantic"
)

func structDefaults(str *semantic.StructSymbol) []TAC {
	instructions := []TAC{}
	addDefaultToString := false
	addDefaultEquals := false
	if castBlock := currScope.LookupNamedBlock("cast"); castBlock == nil {
		addDefaultToString = true
	} else {
		if fn := castBlock.InnerScope.LookupFunctionsByReturnType(dt.StringType); len(fn) == 0 {
			addDefaultToString = true
		}
	}
	if compareBlock := currScope.LookupNamedBlock("compareBlock"); compareBlock == nil {
		addDefaultEquals = true
	} else {
		if fn := compareBlock.InnerScope.LookupFunctionByName("equals"); fn == nil {
			addDefaultEquals = true
		}
	}
	if addDefaultToString {
		instructions = append(instructions, structDefaultToString(str))
	}
	if addDefaultEquals {
		instructions = append(instructions, structDefaultEquals(str))
	}
	return instructions
}

func structDefaultEquals(str *semantic.StructSymbol) Function {
	fn := Function{}
	fn.Name = fmt.Sprintf("__%s_compare__default__equals", str.Name)
	fn.ReturnType = dt.I32
	fn.Parameters = []Parameter{
		{
			Name: "__this",
			Type: dt.Ptr,
		},
		{
			Name: "__other",
			Type: dt.Ptr,
		},
	}
	loadThis := formTempVar(dt.Ptr)
	fn.Code = append(fn.Code, Instruction{
		Destination: loadThis,
		Operation:   Get,
		Operand1: Operand{
			Var: Variable{
				Name:       "__this",
				DataType:   dt.Ptr,
				Visibility: Param,
			},
		},
	})
	loadOther := formTempVar(dt.Ptr)
	fn.Code = append(fn.Code, Instruction{
		Destination: loadOther,
		Operation:   Get,
		Operand1: Operand{
			Var: Variable{
				Name:       "__other",
				DataType:   dt.Ptr,
				Visibility: Param,
			},
		},
	})
	comparisons := []Variable{}
	for _, prop := range str.InnerScope.Variables {
		if !prop.Offset.IsSet {
			continue
		}
		propType := dt.TranslateSourceType(prop.Type)
		loadProp := formTempVar(propType)
		loadOtherProp := formTempVar(propType)
		equals := formTempVar(dt.I32)
		compare := []TAC{
			Instruction{
				Destination: loadProp,
				Operation:   Load,
				Operand1: Operand{
					Var: loadThis,
				},
				Operand2: Operand{
					Constant: prop.Offset.Value,
				},
			},
			Instruction{
				Destination: loadOtherProp,
				Operation:   Load,
				Operand1: Operand{
					Var: loadOther,
				},
				Operand2: Operand{
					Constant: prop.Offset.Value,
				},
			},
		}
		var equality []TAC
		if prop.Type.IsDynamic {
			// TODO
		} else if prop.Type.Equals(dt.StringType) {
			equality = []TAC{
				Instruction{
					Operation: PrepareParam,
					Operand1: Operand{
						Var: loadProp,
					},
				},
				Instruction{
					Operation: PrepareParam,
					Operand1: Operand{
						Var: loadOtherProp,
					},
				},
				Instruction{
					Destination: equals,
					Operation:   Call,
					Operand1: Operand{
						Constant: "__str_eq",
					},
					Operand2: Operand{
						Constant: 2,
					},
				},
			}
		} else {
			equality = []TAC{
				Instruction{
					Destination: equals,
					Operation:   typedOperation(propType, "eq"),
					Operand1: Operand{
						Var: loadProp,
					},
					Operand2: Operand{
						Var: loadOtherProp,
					},
				},
			}
		}
		comparisons = append(comparisons, equals)
		compare = append(compare, equality...)
		// chain together with result of previous comparison
		fn.Code = append(fn.Code, compare...)
	}
	if len(comparisons) == 0 {
		// return true
		fn.Code = append(fn.Code, Instruction{
			Operation: Return,
			Operand1: Operand{
				Type:     dt.I32,
				Constant: 1,
			},
		})
	} else if len(comparisons) == 1 {
		fn.Code = append(fn.Code, Instruction{
			Operation: Return,
			Operand1: Operand{
				Var: comparisons[0],
			},
		})
	} else {
		and := formTempVar(dt.I32)
		fn.Code = append(fn.Code, Instruction{
			Destination: and,
			Operation:   typedOperation(dt.I32, "and"),
			Operand1: Operand{
				Var: comparisons[0],
			},
			Operand2: Operand{
				Var: comparisons[1],
			},
		})
		for i := 2; i < len(comparisons); i++ {
			next_and := formTempVar(dt.I32)
			fn.Code = append(fn.Code, Instruction{
				Destination: next_and,
				Operation:   typedOperation(dt.I32, "and"),
				Operand1: Operand{
					Var: and,
				},
				Operand2: Operand{
					Var: comparisons[i],
				},
			})
			and = next_and
		}
		fn.Code = append(fn.Code, Instruction{
			Operation: Return,
			Operand1: Operand{
				Var: and,
			},
		})
	}
	return fn
}

func structDefaultToString(str *semantic.StructSymbol) Function {
	fn := Function{}
	fn.Name = fmt.Sprintf("__%s_cast__default_toString", str.Name)
	fn.ReturnType = dt.Str_const
	fn.Parameters = []Parameter{
		{
			Name: "__this",
			Type: dt.Ptr,
		},
	}
	load := formTempVar(dt.Ptr)
	fn.Code = append(fn.Code, Instruction{
		Destination: load,
		Operation:   Get,
		Operand1: Operand{
			Var: Variable{
				Name:       "__this",
				DataType:   dt.Ptr,
				Visibility: Param,
			},
		},
	})
	propLoads := []Operand{}
	for _, prop := range str.InnerScope.Variables {
		if prop.Offset.IsSet {
			loadProp := formTempVar(dt.TranslateSourceType(prop.Type))
			fn.Code = append(fn.Code, Instruction{
				Destination: loadProp,
				Operation:   Load,
				Operand1: Operand{
					Var: load,
				},
				Operand2: Operand{
					Constant: prop.Offset.Value,
				},
			})
			if prop.Type.Equals(dt.StringType) {
				propLoads = append(propLoads, Operand{
					Var: loadProp,
				})
			} else if prop.Type.IsDynamic {
				toString := getStructToString(prop.Type.String())
				str_var := formTempVar(dt.Str_const)
				call := []TAC{
					Instruction{
						Operation: PrepareParam,
						Operand1: Operand{
							Var: loadProp,
						},
					},
					Instruction{
						Destination: str_var,
						Operation:   Call,
						Operand1: Operand{
							Constant: toString,
						},
						Operand2: Operand{
							Constant: 1,
						},
					},
				}
				fn.Code = append(fn.Code, call...)
				propLoads = append(propLoads, Operand{
					Var: str_var,
				})
			} else {
				str_var := formTempVar(dt.Str_const)
				call := []TAC{
					Instruction{
						Operation: PrepareParam,
						Operand1: Operand{
							Var: loadProp,
						},
					},
					Instruction{
						Destination: str_var,
						Operation:   Call,
						Operand1: Operand{
							Constant: getToStringFn(prop.Type),
						},
						Operand2: Operand{
							Constant: 1,
						},
					},
				}
				fn.Code = append(fn.Code, call...)
				propLoads = append(propLoads, Operand{
					Var: str_var,
				})
			}
		}
	}
	if len(propLoads) == 0 {
		strVal := formTempVar(dt.Str_const)
		returnStr := []TAC{
			Instruction{
				Operation: PrepareParam,
				Operand1: Operand{
					Type:     dt.I32,
					Constant: '{',
				},
			},
			Instruction{
				Operation: PrepareParam,
				Operand1: Operand{
					Type:     dt.I32,
					Constant: '}',
				},
			},
			Instruction{
				Destination: strVal,
				Operation:   Call,
				Operand1: Operand{
					Constant: CharConcat,
				},
				Operand2: Operand{
					Constant: 2,
				},
			},
			Instruction{
				Operation: Return,
				Operand1: Operand{
					Var: strVal,
				},
			},
		}
		fn.Code = append(fn.Code, returnStr...)
	} else {
		start := formTempVar(dt.Str_const)
		returnStr := []TAC{
			Instruction{
				Operation: PrepareParam,
				Operand1: Operand{
					Type:     dt.I32,
					Constant: '{',
				},
			},
			Instruction{
				Operation: PrepareParam,
				Operand1:  propLoads[0],
			},
			Instruction{
				Destination: start,
				Operation:   Call,
				Operand1: Operand{
					Constant: CharConcatString,
				},
				Operand2: Operand{
					Constant: 2,
				},
			},
		}
		for i := 1; i < len(propLoads); i++ {
			curr := formTempVar(dt.Str_const)
			currStr := []TAC{
				Instruction{
					Operation: PrepareParam,
					Operand1: Operand{
						Type:     dt.I32,
						Constant: ',',
					},
				},
				Instruction{
					Operation: PrepareParam,
					Operand1:  propLoads[i],
				},
				Instruction{
					Destination: curr,
					Operation:   Call,
					Operand1: Operand{
						Constant: CharConcatString,
					},
					Operand2: Operand{
						Constant: 2,
					},
				},
			}
			returnStr = append(returnStr, currStr...)
			combine := formTempVar(dt.Str_const)
			combineStrs := []TAC{
				Instruction{
					Operation: PrepareParam,
					Operand1: Operand{
						Var: start,
					},
				},
				Instruction{
					Operation: PrepareParam,
					Operand1: Operand{
						Var: curr,
					},
				},
				Instruction{
					Destination: combine,
					Operation:   Call,
					Operand1: Operand{
						Constant: StringConcat,
					},
					Operand2: Operand{
						Constant: 2,
					},
				},
			}
			start = combine
			returnStr = append(returnStr, combineStrs...)
		}
		returnVal := formTempVar(dt.Str_const)
		end := []TAC{
			Instruction{
				Operation: PrepareParam,
				Operand1: Operand{
					Var: start,
				},
			},
			Instruction{
				Operation: PrepareParam,
				Operand2: Operand{
					Type:     dt.I32,
					Constant: '}',
				},
			},
			Instruction{
				Destination: returnVal,
				Operation:   Call,
				Operand1: Operand{
					Constant: StringConcatChar,
				},
				Operand2: Operand{
					Constant: 2,
				},
			},
			Instruction{
				Operation: Return,
				Operand1: Operand{
					Var: returnVal,
				},
			},
		}
		returnStr = append(returnStr, end...)
		fn.Code = append(fn.Code, returnStr...)
	}
	return fn
}

func structFunctionDefinitions(ast *parser.AST, str *semantic.StructSymbol) []TAC {
	instructions := []TAC{}
	found_nb_fns := map[string]ds.HashSet{}
	for _, node := range ast.Children[len(ast.Children)-1].Children {
		if node.Label == "Variable" {
			continue
		} else if node.Label == "named-block" {
			scope := currScope
			var nbscope *semantic.Scope
			nbName := node.Children[0].Token.Value
			if nbName == "private" {
				nbscope = currScope // don't change scope
			} else {
				nbscope = currScope.GetChildScopeById(node.IRName)
			}
			currScope = nbscope
			for _, child := range node.Children[1].Children {
				if child.Label == "fn" {
					if hs, ok := found_nb_fns[nbName]; ok {
						hs.Append(child.IRName)
					} else {
						found_nb_fns[nbName] = ds.BuildHashSet(child.IRName)
					}
					fn := functionDefinition(child, true)
					instructions = append(instructions, fn...)
				}
				currScope = nbscope
			}
			currScope = scope
		} else if node.Label == "fn" {
			scope := currScope
			fn := functionDefinition(node, true)
			instructions = append(instructions, fn...)
			currScope = scope
		}
	}
	// get function definitions from interface not in struct def
	scope := currScope
	currScope = str.InnerScope
	for _, nb := range currScope.NamedBlocks {
		currScope = nb.InnerScope
		for _, fn := range currScope.Functions {
			for _, overload := range fn.Overloads {
				if _, ok := found_nb_fns[nb.Name][overload.IRName]; !ok {
					instructions = append(instructions, addMissingOverloadDefinition(overload, fn.ReturnType)...)
				}
			}
		}
	}
	currScope = scope
	return instructions
}

func addMissingOverloadDefinition(overload semantic.FnOverloadSymbol, returnType dt.SourceType) []TAC {
	fn := Function{}
	fn.Name = overload.IRName
	fn.ReturnType = dt.TranslateSourceType(returnType)
	for i := range overload.Parameters {
		fn.Parameters = append(fn.Parameters, Parameter{
			Name: overload.ParameterNames[i],
			Type: dt.TranslateSourceType(overload.Parameters[i]),
		})
	}
	fn.Code = translateBlock(overload.Body.Children, "", "")
	return []TAC{fn}
}

func getStructToString(struct_name string) string {
	irName := ""
	if str := currScope.LookupStruct(struct_name); str != nil {
		castBlock := str.InnerScope.LookupNamedBlock("cast")
		if castBlock == nil {
			irName = fmt.Sprintf("__%s_cast__default_toString", str.Name)
			return irName
		}
		fn := castBlock.InnerScope.LookupFunctionsByReturnType(dt.StringType)
		if len(fn) == 0 {
			irName = fmt.Sprintf("__%s_cast__default_toString", str.Name)
		} else {
			overload := fn[0].Overloads[0]
			irName = overload.IRName
		}
	}
	return irName
}
