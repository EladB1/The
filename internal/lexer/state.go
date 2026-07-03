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
		messages             diagnostic.PhaseDiagnostics
		lineNum              int
		lineIndex            int
		in_multiline_comment bool
		in_word              bool
	}
)

func initState() *lexerState {
	return &lexerState{
		tokens:               []Token{},
		sequence:             strings.Builder{},
		startPosition:        0,
		in_multiline_comment: false,
		messages:             diagnostic.PhaseDiagnostics{},
		lineNum:              0,
		lineIndex:            0,
	}
}

func (stateMchn *lexerState) addError(message string, args ...any) {
	lineIndex := stateMchn.lineIndex
	if stateMchn.startPosition != stateMchn.lineIndex {
		lineIndex = stateMchn.startPosition
	}
	pos := ds.SourceLocation{
		Line:   stateMchn.lineNum,
		Column: lineIndex,
	}
	stateMchn.messages = stateMchn.messages.Complain(errLevel, pos, message, args...)
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
	var token Token
	switch tokenType {
	case LIT_CHAR:
		str, err := strconv.Unquote(stateMchn.sequence.String())
		if err != nil {
			stateMchn.addError("Invalid character literal %s", stateMchn.sequence.String())
			return
		}
		char, size := utf8.DecodeRuneInString(str)
		if char == utf8.RuneError && size > 1 {
			stateMchn.addError("Invalid character literal %s", stateMchn.sequence.String())
			return
		} else if size == 0 {
			char = 0
		}
		token = Token{
			Kind:    tokenType,
			CharVal: char,
			Location: ds.SourceLocation{
				Line:   stateMchn.lineNum,
				Column: column,
			},
		}
	case LIT_STRING:
		str, err := strconv.Unquote(stateMchn.sequence.String())
		if err != nil {
			stateMchn.addError("Invalid string literal %s", stateMchn.sequence.String())
			return
		}
		index := 0
		ds.LiteralStorage, index = ds.LiteralStorage.Add(str)
		token = Token{
			Kind:     tokenType,
			StrIndex: index,
			Location: ds.SourceLocation{
				Line:   stateMchn.lineNum,
				Column: column,
			},
		}
	case LIT_FLOAT:
		val, err := strconv.ParseFloat(stateMchn.sequence.String(), 64)
		if err != nil {
			stateMchn.addError("Invalid floating point literal %s", stateMchn.sequence.String())
			return
		}
		token = Token{
			Kind:     tokenType,
			FloatVal: val,
			Location: ds.SourceLocation{
				Line:   stateMchn.lineNum,
				Column: column,
			},
		}
	case LIT_INT:
		val, err := strconv.ParseInt(stateMchn.sequence.String(), 10, 64)
		if err != nil {
			stateMchn.addError("Invalid integer literal %s", stateMchn.sequence.String())
			return
		}
		token = Token{
			Kind:   tokenType,
			IntVal: int64(val),
			Location: ds.SourceLocation{
				Line:   stateMchn.lineNum,
				Column: column,
			},
		}
	case LIT_HEX:
		val, err := strconv.ParseInt(stateMchn.sequence.String(), 0, 64)
		if err != nil {
			stateMchn.addError("Invalid hexadecimal literal %s", stateMchn.sequence.String())
			return
		}
		token = Token{
			Kind:   tokenType,
			IntVal: int64(val),
			Location: ds.SourceLocation{
				Line:   stateMchn.lineNum,
				Column: column,
			},
		}
	default:
		token = Token{
			Kind:  tokenType,
			Value: stateMchn.sequence.String(),
			Location: ds.SourceLocation{
				Line:   stateMchn.lineNum,
				Column: column,
			},
		}
	}
	stateMchn.tokens = append(stateMchn.tokens, token)
	stateMchn.clearSequence()
}

func (stateMchn *lexerState) buildAndAppendTokenFromByte(tokenType TokenType, char byte) {
	stateMchn.tokens = append(stateMchn.tokens, Token{
		Kind:  tokenType,
		Value: string(char),
		Location: ds.SourceLocation{
			Line:   stateMchn.lineNum,
			Column: stateMchn.lineIndex,
		},
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
