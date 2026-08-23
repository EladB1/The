package irgen

import (
	"fmt"
	"strings"

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
	if compareBlock := currScope.LookupNamedBlock("compare"); compareBlock == nil {
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
			Name:        "__this",
			Type:        dt.Ptr,
			SrcPosition: str.Def.Location,
		},
		{
			Name:        "__other",
			Type:        dt.Ptr,
			SrcPosition: str.Def.Location,
		},
	}
	loadThis := Variable{
		Name:        "__this",
		DataType:    dt.Ptr,
		Visibility:  Param,
		SrcPosition: str.Def.Location,
	}
	loadOther := Variable{
		Name:        "__other",
		DataType:    dt.Ptr,
		Visibility:  Param,
		SrcPosition: str.Def.Location,
	}
	comparisons := []Variable{}
	for prop := range str.InnerScope.Variables.All() {
		if !prop.Offset.IsSet {
			continue
		}
		propType := dt.TranslateSourceType(prop.Type)
		loadProp := formTempVar(propType)
		loadOtherProp := formTempVar(propType)
		var equals Variable
		compare := []TAC{
			Instruction{
				Destination: loadProp,
				Operation:   Load,
				Operand1: Operand{
					Var:         loadThis,
					SrcPosition: str.Def.Location,
				},
				Operand2: Operand{
					Constant:    prop.Offset.Value,
					SrcPosition: str.Def.Location,
				},
				SrcPosition: str.Def.Location,
			},
			Instruction{
				Destination: loadOtherProp,
				Operation:   Load,
				Operand1: Operand{
					Var:         loadOther,
					SrcPosition: str.Def.Location,
				},
				Operand2: Operand{
					Constant:    prop.Offset.Value,
					SrcPosition: str.Def.Location,
				},
				SrcPosition: str.Def.Location,
			},
		}
		var equality []TAC
		if prop.Type.IsDynamic {
			equals = formTempVar(dt.I32)
			var eq Operand
			equality, eq = translateStructComparison(Operand{Var: loadProp}, Operand{Var: loadOtherProp}, prop.Type.String(), "eq")
			equals = eq.Var
		} else if prop.Type.Equals(dt.StringType) {
			var operand Operand
			equality, operand = callFunction("__str_eq", dt.I32, str.Def.Location, Operand{Var: loadProp}, Operand{Var: loadOtherProp})
			equals = operand.Var
		} else {
			equals = formTempVar(dt.I32)
			equality = []TAC{
				Instruction{
					Destination: equals,
					Operation:   typedOperation(propType, "eq"),
					Operand1: Operand{
						Var:         loadProp,
						SrcPosition: str.Def.Location,
					},
					Operand2: Operand{
						Var:         loadOtherProp,
						SrcPosition: str.Def.Location,
					},
					SrcPosition: str.Def.Location,
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
			SrcPosition: str.Def.Location,
		})
	} else if len(comparisons) == 1 {
		fn.Code = append(fn.Code, Instruction{
			Operation: Return,
			Operand1: Operand{
				Var: comparisons[0],
			},
			SrcPosition: str.Def.Location,
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
			SrcPosition: str.Def.Location,
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
				SrcPosition: str.Def.Location,
			})
			and = next_and
		}
		fn.Code = append(fn.Code, Instruction{
			Operation: Return,
			Operand1: Operand{
				Var: and,
			},
			SrcPosition: str.Def.Location,
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
	load := Variable{
		Name:       "__this",
		DataType:   dt.Ptr,
		Visibility: Param,
	}
	propLoads := []Operand{}
	for prop := range str.InnerScope.Variables.All() {
		if !prop.Offset.IsSet {
			continue
		}
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
			call, operand := callFunction(toString, dt.Str_const, str.Def.Location, Operand{Var: loadProp})
			fn.Code = append(fn.Code, call...)
			propLoads = append(propLoads, operand)
		} else {
			call, operand := callFunction(string(getToStringFn(prop.Type)), dt.Str_const, str.Def.Location, Operand{Var: loadProp})
			fn.Code = append(fn.Code, call...)
			propLoads = append(propLoads, operand)
		}
	}
	open := Operand{
		Type:        dt.I32,
		Constant:    '{',
		SrcPosition: str.Def.Location,
	}
	close := Operand{
		Type:        dt.I32,
		Constant:    '}',
		SrcPosition: str.Def.Location,
	}
	if len(propLoads) == 0 {
		returnStr, strVal := callFunction(string(CharConcat), dt.Str_const, str.Def.Location, open, close)
		fn.Code = append(fn.Code, returnStr...)
		fn.Code = append(fn.Code, Instruction{
			Operation:   Return,
			Operand1:    strVal,
			SrcPosition: str.Def.Location,
		})
	} else {
		returnStr, start := callFunction(string(CharConcatString), dt.Str_const, str.Def.Location, open, propLoads[0])
		comma := Operand{
			Type:        dt.I32,
			Constant:    ',',
			SrcPosition: str.Def.Location,
		}
		for i := 1; i < len(propLoads); i++ {
			currStr, curr := callFunction(string(CharConcatString), dt.Str_const, str.Def.Location, comma, propLoads[i])
			returnStr = append(returnStr, currStr...)
			combineStrs, combine := callFunction(string(StringConcat), dt.Str_const, str.Def.Location, start, curr)
			start = combine
			returnStr = append(returnStr, combineStrs...)
		}
		end, returnVal := callFunction(string(StringConcatChar), dt.Str_const, str.Def.Location, start, close)
		end = append(end, Instruction{
			Operation:   Return,
			Operand1:    returnVal,
			SrcPosition: str.Def.Location,
		})
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
	for nb := range currScope.NamedBlocks.All() {
		currScope = nb.InnerScope
		for fn := range currScope.Functions.All() {
			for _, overload := range fn.Overloads {
				if _, ok := found_nb_fns[nb.Name][overload.IRName]; !ok {
					instructions = append(instructions, addMissingOverloadDefinition(fn.Name, overload, fn.ReturnType)...)
				}
			}
		}
	}
	currScope = scope
	return instructions
}

func addMissingOverloadDefinition(fnName string, overload semantic.FnOverloadSymbol, returnType dt.SourceType) []TAC {
	fn := Function{}
	fn.Name = overload.IRName
	scope := currScope
	currScope = currScope.GetChildScopeById(fmt.Sprintf("%s@%s", fnName, currScope.Id))
	fn.ReturnType = dt.TranslateSourceType(returnType)
	for i := range overload.Parameters {
		fn.Parameters = append(fn.Parameters, Parameter{
			Name: overload.ParameterNames[i],
			Type: dt.TranslateSourceType(overload.Parameters[i]),
		})
	}
	fn.Code = translateBlock(overload.Body.Children, "", "")
	currScope = scope
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

func getStructEquals(compareBlock *semantic.NamedBlockSymbol, struct_name string) string {
	equalsName := fmt.Sprintf("__%s_compare__default__equals", struct_name)
	if compareBlock == nil {
		return equalsName
	}
	equalsFn := compareBlock.InnerScope.LookupFunctionByName("equals")
	if equalsFn != nil {
		equalsName = equalsFn.Overloads[0].IRName
	}
	return equalsName
}

func translateStructComparison(l_op, r_op Operand, struct_name string, comp string) ([]TAC, Operand) {
	instructions := []TAC{}
	str := currScope.LookupStruct(struct_name)
	equalsName := getStructEquals(str.InnerScope.LookupNamedBlock("compare"), struct_name)
	compared := false
	//compare := formTempVar(dt.I32)
	var compare Operand
	var call []TAC
	if compareBlock := str.GetInnerScope().LookupNamedBlock("compare"); compareBlock != nil {
		if strings.Contains(comp, "l") {
			if lessFn := compareBlock.InnerScope.LookupFunctionByName("lessThan"); lessFn != nil {
				lessName := lessFn.Overloads[0].IRName
				call, compare = callFunction(lessName, dt.I32, l_op.SrcPosition, l_op, r_op)
				instructions = append(instructions, call...)
				compared = true
			}
		}
		if strings.Contains(comp, "g") {
			if greaterFn := compareBlock.InnerScope.LookupFunctionByName("greaterThan"); greaterFn != nil {
				greaterName := greaterFn.Overloads[0].IRName
				call, compare = callFunction(greaterName, dt.I32, l_op.SrcPosition, l_op, r_op)
				instructions = append(instructions, call...)
				compared = true
			}
		}
	}
	if strings.Contains(comp, "e") {
		first := compare
		call, compare = callFunction(equalsName, dt.I32, l_op.SrcPosition, l_op, r_op)
		instructions = append(instructions, call...)
		if compared {
			or := formTempVar(dt.I32)
			instructions = append(instructions, Instruction{
				Destination: or,
				Operation:   typedOperation(dt.I32, "or"),
				Operand1:    first,
				Operand2:    compare,
			})
			compare.Var = or
		}
	}
	if comp == "ne" {
		not := formTempVar(dt.I32)
		instructions = append(instructions, Instruction{
			Destination: not,
			Operation:   typedOperation(dt.I32, "xor"),
			Operand1:    compare,
			Operand2: Operand{
				Type:     dt.I32,
				Constant: 1,
			},
		})
		compare.Var = not
	}
	return instructions, compare
}
