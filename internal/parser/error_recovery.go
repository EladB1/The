package parser

import "github.com/EladB1/The/internal/lexer"

func errorRecoveryTopLevel() {
	for !checkKind(lexer.EOF) && !checkVariableDeclaration() && !checkNonVariableDeclaration() {
		consume()
	}
}

func synchronize() {
	depth := 0
	for !checkKind(lexer.EOF) {
		if checkValue("{") {
			depth++
		} else if checkValue("}") {
			if depth == 0 {
				consume()
				return
			}
			depth--
		} else if checkValue(";") && depth == 0 {
			consume()
			return
		}
		consume()
	}
}

func synchronizeInParens() {
	depth := 0
	for !checkKind(lexer.EOF) {
		if checkValue("(") {
			depth++
		} else if checkValue(")") {
			if depth == 0 {
				//consume()
				return
			}
			depth--
		} else if checkValue(";") && depth == 0 {
			consume()
			return
		}
		consume()
	}
}
