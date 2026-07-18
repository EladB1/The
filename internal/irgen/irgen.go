package irgen

import (
	"github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/diagnostic"
	"github.com/EladB1/The/internal/lexer"
	"github.com/EladB1/The/internal/parser"
	"github.com/EladB1/The/internal/semantic"
)

var messages diagnostic.PhaseDiagnostics
var currScope *semantic.Scope

func Generate(ast parser.AST, scopeTree *semantic.Scope) (Program, diagnostic.PhaseDiagnostics) {
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
	irType := translateSourceType(ast.Type)
	if valueNode == nil {
		value = getZeroValue(ast.Type)
	} else {
		value = handleLiteral(*valueNode)
	}
	return []TAC{Instruction{
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
	}}
}

func getZeroValue(sourceType datatypes.DataType) Operand {
	if !sourceType.IsPrimitive() {
		// TODO
	}
	return Operand{
		Type:     translateSourceType(sourceType),
		Constant: 0,
	}
}

func handleLiteral(node parser.AST) Operand {
	irType := none
	unsigned := false
	var value any
	switch node.Token.Kind {
	case lexer.KW_BOOLVALUE:
		irType = i32
		if node.Token.Value == "true" {
			value = 1
		} else {
			value = 0
		}
	case lexer.LIT_CHAR:
		irType = i32
		value = node.Token.CharVal
	case lexer.LIT_INT, lexer.LIT_HEX:
		unsigned = node.Type == datatypes.Uint32 || node.Type == datatypes.Uint64
		switch node.Type {
		case datatypes.Int32, datatypes.Uint32:
			irType = i32
		case datatypes.Int64, datatypes.Uint64:
			irType = i64
			value = node.Token.IntVal
		case datatypes.Float:
			irType = f32
		case datatypes.Double:
			irType = f64
		}
	case lexer.LIT_FLOAT:
		if node.Type == datatypes.Float {
			irType = f32
		} else {
			irType = f64
		}
		value = node.Token.FloatVal
	case lexer.LIT_STRING:
		irType = str_const
		value = node.Token.StrIndex
	}
	return Operand{
		Type:     irType,
		Constant: value,
		Unsigned: unsigned,
	}
}

func translateSourceType(sourceType datatypes.DataType) Datatype {
	switch sourceType {
	case datatypes.String:
		return str_const
	case datatypes.Char, datatypes.Bool, datatypes.Int32, datatypes.Uint32:
		return i32
	case datatypes.Int64, datatypes.Uint64:
		return i64
	case datatypes.Float:
		return f32
	case datatypes.Double:
		return f64
	default:
		return none
	}
}
