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
	var token Token
	switch tokenType {
	case LIT_CHAR:
		str, err := strconv.Unquote(stateMchn.sequence.String())
		if err != nil {
			stateMchn.addError(fmt.Sprintf("Invalid character literal %s", stateMchn.sequence.String()))
			return
		}
		char, size := utf8.DecodeRuneInString(str)
		if char == utf8.RuneError && size > 1 {
			stateMchn.addError(fmt.Sprintf("Invalid character literal %s", stateMchn.sequence.String()))
			return
		} else if size == 0 {
			char = 0
		}
		token = Token{
			tokenType: tokenType,
			CharVal:   char,
			line:      stateMchn.lineNum + 1,
			column:    column + 1,
		}
	case LIT_STRING:
		str, err := strconv.Unquote(stateMchn.sequence.String())
		if err != nil {
			fmt.Println(stateMchn.sequence.String())
			stateMchn.addError(fmt.Sprintf("Invalid string literal %s", stateMchn.sequence.String()))
			return
		}
		index := 0
		ds.LiteralStorage, index = ds.LiteralStorage.Add(str)
		token = Token{
			tokenType: tokenType,
			StrIndex:  index,
			line:      stateMchn.lineNum + 1,
			column:    column + 1,
		}
	case LIT_FLOAT:
		val, err := strconv.ParseFloat(stateMchn.sequence.String(), 64)
		if err != nil {
			stateMchn.addError(fmt.Sprintf("Invalid floating point literal %s", stateMchn.sequence.String()))
			return
		}
		token = Token{
			tokenType: tokenType,
			FloatVal:  val,
			line:      stateMchn.lineNum + 1,
			column:    column + 1,
		}
	case LIT_INT:
		val, err := strconv.ParseInt(stateMchn.sequence.String(), 10, 64)
		if err != nil {
			stateMchn.addError(fmt.Sprintf("Invalid integer literal %s", stateMchn.sequence.String()))
			return
		}
		token = Token{
			tokenType: tokenType,
			IntVal:    uint64(val),
			line:      stateMchn.lineNum + 1,
			column:    column + 1,
		}
	case LIT_HEX:
		val, err := strconv.ParseInt(stateMchn.sequence.String(), 0, 64)
		if err != nil {
			stateMchn.addError(fmt.Sprintf("Invalid hexadecimal literal %s", stateMchn.sequence.String()))
			return
		}
		token = Token{
			tokenType: tokenType,
			IntVal:    uint64(val),
			line:      stateMchn.lineNum + 1,
			column:    column + 1,
		}
	default:
		token = Token{
			tokenType: tokenType,
			value:     stateMchn.sequence.String(),
			line:      stateMchn.lineNum + 1,
			column:    column + 1,
		}
	}
	stateMchn.tokens = append(stateMchn.tokens, token)
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
