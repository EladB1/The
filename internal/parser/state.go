package parser

import (
	ds "github.com/EladB1/The/internal/datastructures"
	"github.com/EladB1/The/internal/diagnostic"
	"github.com/EladB1/The/internal/lexer"
)

type (
	parserState struct {
		tokens   []lexer.Token
		pool     ds.LiteralPool
		length   int
		ptr      int
		messages diagnostic.PhaseDiagnostics
		in_error bool
	}
)

func initState(tokens []lexer.Token, pool ds.LiteralPool) *parserState {
	return &parserState{
		tokens:   tokens,
		pool:     pool,
		length:   len(tokens),
		ptr:      0,
		messages: diagnostic.PhaseDiagnostics{},
		in_error: false,
	}
}

func (state *parserState) addError(message string, args ...any) {
	token := state.tokens[state.ptr]
	state.messages.Complain(errLevel, token.Location, message, args...)
}
