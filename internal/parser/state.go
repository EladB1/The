package parser

import (
	"github.com/EladB1/The/internal/diagnostic"
	"github.com/EladB1/The/internal/lexer"
)

type (
	parserState struct {
		tokens   []lexer.Token
		length   int
		ptr      int
		messages diagnostic.PhaseDiagnostics
		in_error bool
	}
)

func initState(tokens []lexer.Token) *parserState {
	return &parserState{
		tokens:   tokens,
		length:   len(tokens),
		ptr:      0,
		messages: diagnostic.PhaseDiagnostics{},
		in_error: false,
	}
}

func (stateMchn *parserState) addError(message string) {
	token := stateMchn.tokens[stateMchn.ptr]
	stateMchn.messages = stateMchn.messages.Complain(errLevel, message, token.Line, token.Column)
}
