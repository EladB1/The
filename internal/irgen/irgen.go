package irgen

import (
	"github.com/EladB1/The/internal/datatypes"
	dt "github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/diagnostic"
	"github.com/EladB1/The/internal/lexer"
	"github.com/EladB1/The/internal/parser"
	"github.com/EladB1/The/internal/semantic"
)

var messages diagnostic.PhaseDiagnostics
var currScope *semantic.Scope
var tempVarIndex uint32

func Generate(ast parser.AST, scopeTree *semantic.Scope) (Program, diagnostic.PhaseDiagnostics) {
	tempVarIndex = 0
	prog := Program{}
	messages = diagnostic.PhaseDiagnostics{}
	currScope = scopeTree.Children[0] // get the global scope using the built-in scope
	for _, node := range ast.Children {
		if node.Label == "Variable" {
			prog.appendCode(variableDeclaration(node, scopeTree))
		}
	}
	return prog, messages
}

func variableDeclaration(ast *parser.AST, scopeTree *semantic.Scope) []TAC {
	/*
		no value => add "zero" value
		value => get code for value
	*/
	details := ast.Children
	var name string = details[1].Token.Value
	var vis VariableScope
	var valueNode *parser.AST = nil
	var value Operand
	instructions := []TAC{}
	if currScope.Id == "@global" {
		vis = Global
	} else { // TODO handle function parameters
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
	if !sourceType.IsPrimitive() {
		// TODO
	}
	return Operand{
		Type:     dt.TranslateSourceType(sourceType),
		Constant: 0,
	}
}

func translateLiteral(node parser.AST) Operand {
	irType := datatypes.NoneIR
	var value any
	switch node.Token.Kind {
	case lexer.KW_BOOLVALUE:
		irType = datatypes.I32
		if node.Token.Value == "true" {
			value = 1
		} else {
			value = 0
		}
	case lexer.LIT_CHAR:
		irType = datatypes.I32
		value = node.Token.CharVal
	case lexer.LIT_INT, lexer.LIT_HEX:
		switch node.Type {
		case dt.Int32:
			irType = datatypes.I32
			value = int32(node.Token.IntVal)
		case dt.Uint32:
			irType = datatypes.U32
			value = uint32(node.Token.IntVal)
		case dt.Int64:
			irType = datatypes.I64
			value = node.Token.IntVal
		case dt.Uint64:
			irType = datatypes.U64
			value = uint64(node.Token.IntVal)
		case dt.Float:
			irType = datatypes.F32
			value = float32(node.Token.FloatVal)
		case dt.Double:
			irType = datatypes.F64
			value = node.Token.FloatVal
		}
	case lexer.LIT_FLOAT:
		if node.Type == dt.Float {
			irType = datatypes.F32
			value = float32(node.Token.FloatVal)
		} else {
			irType = datatypes.F64
			value = node.Token.FloatVal
		}
	case lexer.LIT_STRING:
		irType = datatypes.Str_const
		value = node.Token.StrIndex
	}
	return Operand{
		Type:     irType,
		Constant: value,
	}
}
