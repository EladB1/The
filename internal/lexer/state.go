package lexer

import (
	"fmt"
	"strings"

	"github.com/EladB1/The/internal/diagnostic"
)

type (
	lexerState struct {
		sequence             strings.Builder
		startPosition        int
		tokens               []Token
		messages             diagnostic.PhaseDiagnostics
		lineNum              int
		lineIndex            int
		in_multiline_comment bool
		in_word              bool
	}
)

func (stateMchn *lexerState) addError(message string) {
	lineIndex := stateMchn.lineIndex
	if stateMchn.startPosition != stateMchn.lineIndex {
		lineIndex = stateMchn.startPosition
	}
	stateMchn.messages = stateMchn.messages.Complain(errLevel, message, stateMchn.lineNum, lineIndex)
}

func (stateMchn *lexerState) push(char byte) {
	stateMchn.sequence.WriteByte(char)
}

func (stateMchn *lexerState) clearSequence() {
	stateMchn.sequence.Reset()
}

func (stateMchn *lexerState) buildAndAppendToken(tokenType TokenType) {
	column := stateMchn.lineIndex
	if stateMchn.sequence.Len() > 1 && stateMchn.startPosition != stateMchn.lineIndex {
		column = stateMchn.startPosition
	}
	stateMchn.tokens = append(stateMchn.tokens, Token{
		tokenType: tokenType,
		value:     stateMchn.sequence.String(),
		line:      stateMchn.lineNum + 1,
		column:    column + 1,
	})
	stateMchn.clearSequence()
}

func (stateMchn *lexerState) buildAndAppendTokenFromByte(tokenType TokenType, char byte) {
	stateMchn.tokens = append(stateMchn.tokens, Token{
		tokenType: tokenType,
		value:     string(char),
		line:      stateMchn.lineNum + 1,
		column:    stateMchn.lineIndex + 1,
	})
	stateMchn.clearSequence()
}

func (stateMchn *lexerState) debug() {
	fmt.Printf("State: {Sequence: %s, position: %d, flags: {in_multiline_comment: %v}, sequence start: %d}\n",
		stateMchn.sequence.String(),
		stateMchn.startPosition,
		stateMchn.in_multiline_comment,
		stateMchn.startPosition,
	)
}
