package parser

import "github.com/EladB1/The/internal/lexer"

func errorRecoveryTopLevel() {
	for !checkKind(lexer.EOF) && !checkKind(lexer.KW_MODIFIER) && !checkKind(lexer.KW_TYPE) && !checkValue("fn") && !checkValue("interface") {
		consume()
	}
}

func errorRecoveryFunctionDefintion() {
	//line := peek().Line
	consume()
	for !checkKind(lexer.EOF) && !checkValue("fn") && !checkValue("struct") && !checkValue("interface") && !checkValue("}") /*|| peek().Line == line*/ {
		consume()
	}
	state.in_error = false
}
