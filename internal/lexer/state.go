package lexer

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	ds "github.com/EladB1/The/internal/datastructures"
	"github.com/EladB1/The/internal/diagnostic"
)

type (
	lexerState struct {
		sequence             strings.Builder
		startPosition        int
		tokens               []Token
		pool                 ds.LiteralPool
		messages             diagnostic.PhaseDiagnostics
		lineNum              int
		lineIndex            int
		in_multiline_comment bool
		in_word              bool
	}
)

func initState() *lexerState {
	return &lexerState{

		sequence:             strings.Builder{},
		startPosition:        0,
		tokens:               []Token{},
		pool:                 ds.LiteralPool{},
		messages:             diagnostic.PhaseDiagnostics{},
		lineNum:              0,
		lineIndex:            0,
		in_multiline_comment: false,
		in_word:              false,
	}
}

func (state *lexerState) addError(message string, args ...any) {
	lineIndex := state.lineIndex
	if state.startPosition != state.lineIndex {
		lineIndex = state.startPosition
	}
	pos := ds.SourceLocation{
		Line:   state.lineNum,
		Column: lineIndex,
	}
	state.messages.Complain(errLevel, pos, message, args...)
}

func (state *lexerState) push(char byte) {
	state.sequence.WriteByte(char)
}

func (state *lexerState) clearSequence() {
	state.sequence.Reset()
}

func (state *lexerState) buildAndAppendToken(tokenType TokenType) {
	column := state.lineIndex
	if state.sequence.Len() > 1 && state.startPosition != state.lineIndex {
		column = state.startPosition
	}
	var token Token
	switch tokenType {
	case LIT_CHAR:
		str, err := strconv.Unquote(state.sequence.String())
		if err != nil {
			state.addError("Invalid character literal %s", state.sequence.String())
			return
		}
		char, size := utf8.DecodeRuneInString(str)
		if char == utf8.RuneError && size > 1 {
			state.addError("Invalid character literal %s", state.sequence.String())
			return
		} else if size == 0 {
			char = 0
		}
		token = Token{
			Kind:    tokenType,
			CharVal: char,
			Location: ds.SourceLocation{
				Line:   state.lineNum,
				Column: column,
			},
		}
	case LIT_STRING:
		str, err := strconv.Unquote(state.sequence.String())
		if err != nil {
			state.addError("Invalid string literal %s", state.sequence.String())
			return
		}
		index := 0
		state.pool, index = state.pool.Add(str)
		token = Token{
			Kind:     tokenType,
			StrIndex: index,
			Location: ds.SourceLocation{
				Line:   state.lineNum,
				Column: column,
			},
		}
	case LIT_FLOAT:
		val, err := strconv.ParseFloat(state.sequence.String(), 64)
		if err != nil {
			state.addError("Invalid floating point literal %s", state.sequence.String())
			return
		}
		token = Token{
			Kind:     tokenType,
			FloatVal: val,
			Location: ds.SourceLocation{
				Line:   state.lineNum,
				Column: column,
			},
		}
	case LIT_INT:
		val, err := strconv.ParseInt(state.sequence.String(), 10, 64)
		if err != nil {
			state.addError("Invalid integer literal %s", state.sequence.String())
			return
		}
		token = Token{
			Kind:   tokenType,
			IntVal: int64(val),
			Location: ds.SourceLocation{
				Line:   state.lineNum,
				Column: column,
			},
		}
	case LIT_HEX:
		val, err := strconv.ParseInt(state.sequence.String(), 0, 64)
		if err != nil {
			state.addError("Invalid hexadecimal literal %s", state.sequence.String())
			return
		}
		token = Token{
			Kind:   tokenType,
			IntVal: int64(val),
			Location: ds.SourceLocation{
				Line:   state.lineNum,
				Column: column,
			},
		}
	default:
		token = Token{
			Kind:  tokenType,
			Value: state.sequence.String(),
			Location: ds.SourceLocation{
				Line:   state.lineNum,
				Column: column,
			},
		}
	}
	state.tokens = append(state.tokens, token)
	state.clearSequence()
}

func (state *lexerState) buildAndAppendTokenFromByte(tokenType TokenType, char byte) {
	state.tokens = append(state.tokens, Token{
		Kind:  tokenType,
		Value: string(char),
		Location: ds.SourceLocation{
			Line:   state.lineNum,
			Column: state.lineIndex,
		},
	})
	state.clearSequence()
}

func (state *lexerState) debug() {
	fmt.Printf("State: {Sequence: %s, position: %d, flags: {in_multiline_comment: %v}, sequence start: %d}\n",
		state.sequence.String(),
		state.startPosition,
		state.in_multiline_comment,
		state.startPosition,
	)
}
