package irgen

import (
	"fmt"

	ds "github.com/EladB1/The/internal/datastructures"
	dt "github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/diagnostic"
	"github.com/EladB1/The/internal/lexer"
	"github.com/EladB1/The/internal/parser"
	"github.com/EladB1/The/internal/semantic"
)

var messages diagnostic.PhaseDiagnostics
var currScope *semantic.Scope
var scopes *semantic.Scope
var tempVarIndex uint32
var loopIndex uint32

func Generate(ast parser.AST, scopeTree *semantic.Scope) (Program, diagnostic.PhaseDiagnostics) {
	scopes = scopeTree
	tempVarIndex = 0
	loopIndex = 0
	prog := Program{}
	messages = diagnostic.PhaseDiagnostics{}
	currScope = scopeTree.Children[0] // get the global scope using the built-in scope
	for _, node := range ast.Children {
		if node.Token.Value == "struct" {
			struct_name := node.Children[0].Token.Value
			str := currScope.LookupStruct(struct_name)

			if str == nil {
				continue
			}
			scope := currScope
			currScope = str.InnerScope
			prog.appendCode(structFunctionDefinitions(node, str))
			currScope = scope
			// TODO: struct fn definitions
		}
	}
	for _, node := range ast.Children {
		switch node.Label {
		case "Variable":
			prog.appendCode(variableDeclaration(node))
		case "fn":
			prog.appendCode(functionDefinition(node, false))
		}
	}
	return prog, messages
}

func structFunctionDefinitions(ast *parser.AST, str *semantic.StructSymbol) []TAC {
	instructions := []TAC{}
	// TODO: get inherited functions implementations
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
	fmt.Println(found_nb_fns)
	scope := currScope
	currScope = str.InnerScope
	for _, nb := range currScope.NamedBlocks {
		currScope = nb.InnerScope
		for _, fn := range currScope.Functions {
			for _, overload := range fn.Overloads {
				if _, ok := found_nb_fns[nb.Name][overload.IRName]; !ok {
					instructions = append(instructions, addMissingOverloadDefinition(overload, fn.ReturnType)...)
					//def := functionDefinition(overload.Body, true)

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

func functionDefinition(ast *parser.AST, inStruct bool) []TAC {
	fn := Function{}
	fn.Name = ast.IRName
	returnType := ast.Type
	fn.ReturnType = dt.TranslateSourceType(returnType)
	name := ast.Children[0].Token.Value
	overload := currScope.LookupFunctionByNameAndIRName(name, fn.Name)
	if overload == nil {
		return []TAC{fn}
	}
	scope := currScope
	fn.Parameters = []Parameter{}
	if len(overload.Parameters) != 0 {
		params := ast.Children[1]
		for i := range len(overload.Parameters) {
			fn.Parameters = append(fn.Parameters, Parameter{
				Name: params.Children[i].Children[1].Token.Value,
				Type: dt.TranslateSourceType(overload.Parameters[i]),
			})
		}
	}
	if inStruct {
		fn.Parameters = append(fn.Parameters, Parameter{
			Name: "__this",
			Type: dt.Ptr,
		})
	}
	if overload.HasDefaultImplementation {
		currScope = overload.InnerScope
		fn.Code = append(fn.Code, translateBlock(overload.Body.Children, "", "")...)
	}
	currScope = scope
	return []TAC{fn}
}

func variableDeclaration(ast *parser.AST) []TAC {
	details := ast.Children
	var name string = details[1].Token.Value
	var vis VariableScope
	var valueNode *parser.AST = nil
	var value Operand
	instructions := []TAC{}
	if currScope.Id == "@global" {
		vis = Global
	} else {
		vis = Local
	}
	if details[0].Label == "modifiers" {
		name = ast.Children[2].Token.Value
		if len(details) == 4 {
			valueNode = details[3]
		}
	} else {
		if len(details) == 3 {
			valueNode = details[2]
		}
	}
	irType := dt.TranslateSourceType(ast.Type)
	if valueNode == nil {
		value = getZeroValue(ast.Type)
	} else {
		instructions, value = translateExpression(*valueNode)

	}
	return append(instructions, Instruction{
		Operation: Store,
		Operand1: Operand{
			Type: irType,
			Var: Variable{
				Name:       name,
				DataType:   irType,
				Visibility: vis,
			},
		},
		Operand2: value,
	})
}

func getZeroValue(sourceType dt.SourceType) Operand {
	if sourceType.IsDynamic {
		// TODO
	}
	return Operand{
		Type:     dt.TranslateSourceType(sourceType),
		Constant: 0,
	}
}

func translateLiteral(node parser.AST) Operand {
	irType := dt.NoneIR
	var value any
	switch node.Token.Kind {
	case lexer.KW_BOOLVALUE:
		irType = dt.I32
		if node.Token.Value == "true" {
			value = 1
		} else {
			value = 0
		}
	case lexer.LIT_CHAR:
		irType = dt.I32
		value = node.Token.CharVal
	case lexer.LIT_INT, lexer.LIT_HEX:
		if node.Type.Equals(dt.Int32Type) {
			irType = dt.I32
			value = int32(node.Token.IntVal)
		} else if node.Type.Equals(dt.Uint32Type) {
			irType = dt.U32
			value = uint32(node.Token.IntVal)
		} else if node.Type.Equals(dt.Int64Type) {
			irType = dt.I64
			value = node.Token.IntVal
		} else if node.Type.Equals(dt.Uint64Type) {
			irType = dt.U64
			value = uint64(node.Token.IntVal)
		} else if node.Type.Equals(dt.FloatType) {
			irType = dt.F32
			value = float32(node.Token.FloatVal)
		} else if node.Type.Equals(dt.DoubleType) {
			irType = dt.F64
			value = node.Token.FloatVal
		}
	case lexer.LIT_FLOAT:
		if node.Type.Equals(dt.FloatType) {
			irType = dt.F32
			value = float32(node.Token.FloatVal)
		} else {
			irType = dt.F64
			value = node.Token.FloatVal
		}
	case lexer.LIT_STRING:
		irType = dt.Str_const
		value = node.Token.StrIndex
	}
	return Operand{
		Type:     irType,
		Constant: value,
	}
}

func translateStructLiteral(node parser.AST) ([]TAC, Operand) {
	instructions := []TAC{}

	struct_name := node.Children[0].Token.Value
	symbol := currScope.LookupStruct(struct_name)
	if symbol == nil {
		return instructions, Operand{}
	}
	instance := formTempVar(dt.Ptr)
	instructions = append(instructions, Instruction{
		Destination: instance,
		Operation:   Malloc,
		Operand1: Operand{
			Constant: symbol.SizeInBytes,
		},
	})
	foundProps := ds.HashSet{}
	var offset semantic.OffsetValue
	for _, prop := range node.Children[1].Children {
		propname := prop.Children[0].Token.Value
		foundProps.Append(propname)
		propVar := symbol.InnerScope.LookupVariable(propname)
		if propVar != nil {
			offset = propVar.Offset
		}
		propvalue_in, propvalue := translateExpression(*prop.Children[1])
		instructions = append(instructions, propvalue_in...)
		instructions = append(instructions, Instruction{
			Operation: Set,
			Operand1: Operand{
				Var:    instance,
				Offset: offset,
			},
			Operand2: propvalue,
		})
	}

	var value_op Operand
	var value_in []TAC
	// fill in default values for missing properties
	for _, variable := range symbol.InnerScope.Variables {
		if _, ok := foundProps[variable.Name]; ok || variable.Type.RootEquals(dt.ScopeRef) || variable.Type.RootEquals(dt.Ref) {
			continue
		}
		if variable.Initialized && variable.Def != nil {
			value_in, value_op = translateExpression(*variable.Def.Children[len(variable.Def.Children)-1])
			instructions = append(instructions, value_in...)
		} else {
			value_op = getZeroValue(variable.Type)
		}
		instructions = append(instructions, value_in...)
		instructions = append(instructions, Instruction{
			Operation: Set,
			Operand1: Operand{
				Var:    instance,
				Offset: variable.Offset,
			},
			Operand2: value_op,
		})
	}
	operand := Operand{
		Var: instance,
	}
	return instructions, operand
}

func formTempVar(irType dt.IRType) Variable {
	tempVar := Variable{
		Name:     fmt.Sprintf("__t%d", tempVarIndex),
		DataType: irType,
	}
	tempVarIndex++
	return tempVar

}

func storeVariable(variable semantic.VariableSymbol, value Operand) Instruction {
	if variable.Ctx == semantic.StructProp {
		return Instruction{
			Operation: Set,
			Operand1: Operand{
				Var:    Variable{},
				Offset: variable.Offset,
			},
			Operand2: value,
		}
	}
	return Instruction{
		Operation: Store,
		Operand1: Operand{
			Var: Variable{
				Name:       variable.Name,
				DataType:   dt.TranslateSourceType(variable.Type),
				Visibility: VariableScope(variable.Ctx),
			},
		},
		Operand2: value,
	}
}
