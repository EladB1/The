package irgen

import (
	"fmt"

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
		switch node.Label {
		case "Variable":
			prog.appendCode(variableDeclaration(node))
		case "fn":
			prog.appendCode(functionDefinition(node))
		}
	}
	return prog, messages
}

func functionDefinition(ast *parser.AST) []TAC {
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
	if overload.HasDefaultImplementation {
		currScope = overload.InnerScope
		fn.Code = append(fn.Code, translateBlock(overload.Body.Children)...)
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

func formTempVar(irType dt.IRType) Variable {
	tempVar := Variable{
		Name:     fmt.Sprintf("__t%d", tempVarIndex),
		DataType: irType,
	}
	tempVarIndex++
	return tempVar

}
