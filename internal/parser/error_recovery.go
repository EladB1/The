package parser

import "github.com/EladB1/The/internal/lexer"

type SyncCtx int32

const (
	topLevelCtx SyncCtx = iota
	fnSignatureCtx
	blockCtx
	structBodyCtx
	ifSignatureCtx
	whileSignatureCtx
	forSignatureCtx
	interfaceCtx
	expressionCtx
	indexValueCtx
)

func sync(ctx SyncCtx) {
	switch ctx {
	case blockCtx, structBodyCtx, expressionCtx:
		synchronize()
	case whileSignatureCtx, forSignatureCtx:
		synchronizeInParens()
	case interfaceCtx:
		synchronizeInterface()
	default:
		errorRecoveryTopLevel()
	}
}

func errorRecoveryTopLevel() {
	for !checkKind(lexer.EOF) && !checkNonVariableDeclaration() {
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

func synchronizeInterface() {
	depth := 0
	for !checkKind(lexer.EOF) {
		if checkValue("{") {
			depth++
		} else if checkValue("}") {
			if depth == 0 {
				return
			}
			depth--
		} else if checkValue("fn") && depth == 0 {
			return
		}
		consume()
	}
}
