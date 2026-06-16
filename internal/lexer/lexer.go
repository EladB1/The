package lexer

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/EladB1/The/internal/diagnostic"
)

const (
	errLevel diagnostic.Severity = diagnostic.SyntaxError
)

var (
	curr byte
	next byte
)

func Lex(sourceCode []string, debug bool) ([]Token, diagnostic.PhaseDiagnostics) {
	state := &lexerState{
		tokens:               []Token{},
		sequence:             strings.Builder{},
		startPosition:        0,
		in_multiline_comment: false,
		messages:             diagnostic.PhaseDiagnostics{},
		lineNum:              0,
		lineIndex:            0,
	}
	for ; state.lineNum < len(sourceCode); state.lineNum++ {
		state.clearSequence()
		state.lineIndex = 0
		state.lexLine(sourceCode[state.lineNum])
	}
	// EOF actions
	if state.in_multiline_comment {
		state.messages = state.messages.ComplainPositionless(errLevel, "Reached EOF while scanning for */")
	}
	return state.tokens, state.messages
}

func (state *lexerState) lexLine(line string) {
	length := len(line)
	for ; state.lineIndex < length; state.lineIndex++ {
		curr = line[state.lineIndex]
		next = state.getNextChar(line)
		if state.in_multiline_comment {
			if curr == '*' && next == '/' {
				state.in_multiline_comment = false
				state.clearSequence()
				state.lineIndex++
			}
			continue
		}
		if unicode.IsSpace(rune(curr)) {
			continue
		}
		state.push(curr)
		//fmt.Printf("line: %d, curr: %c, next: %c\n", state.lineNum, curr, next)
		state.startPosition = state.lineIndex
		switch curr {
		case '+':
			if next == '+' || next == '=' {
				state.push(next)
				tokenType := getTokenTypeForOperator(state.sequence)
				state.buildAndAppendToken(tokenType)
				state.lineIndex++
			} else {
				state.buildAndAppendTokenFromByte(OPERATOR_ADD, curr)
			}
		case '-':
			switch next {
			case '>':
				state.push(next)
				state.buildAndAppendToken(SEPARATOR)
				state.lineIndex++
			case '-', '=':
				state.push(next)
				tokenType := getTokenTypeForOperator(state.sequence)
				state.buildAndAppendToken(tokenType)
				state.lineIndex++
			default:
				state.buildAndAppendTokenFromByte(OPERATOR_ADD, curr)
			}
		case '*':
			if next == '*' || next == '=' {
				state.push(next)
				tokenType := getTokenTypeForOperator(state.sequence)
				state.buildAndAppendToken(tokenType)
				state.lineIndex++
			} else if next == '/' && !state.in_multiline_comment {
				state.addError("Found multiline comment close but no open")
				state.lineIndex++
				state.clearSequence()
			} else {
				state.buildAndAppendTokenFromByte(OPERATOR_MULT, curr)
			}
		case '/':
			switch next {
			case '/':
				state.lineIndex++
				state.clearSequence()
				return // skip to the next line
			case '*':
				state.lineIndex++
				state.clearSequence()
				state.in_multiline_comment = true
			case '=':
				state.push(next)
				state.buildAndAppendToken(OPERATOR_ASSIGN)
				state.lineIndex++
			default:
				state.buildAndAppendTokenFromByte(OPERATOR_MULT, curr)
			}
		case '%':
			state.buildAndAppendTokenFromByte(OPERATOR_MULT, curr)
		case '.':
			if next == '.' { // .. or ..=
				state.push(next)
				if state.lineIndex < length-2 { // check if character after next is =
					next = line[state.lineIndex+2]
					if next == '=' {
						state.push(next)
						state.lineIndex++
					}
				}
				state.lineIndex++
				state.buildAndAppendToken(OPERATOR_RANGE)
			} else if unicode.IsDigit(rune(next)) && !(curr == '0' && next == 'x') { // Example: .234
				state.lineIndex++
				for ; state.lineIndex < length; state.lineIndex++ {
					curr = line[state.lineIndex]
					next = state.getNextChar(line)
					state.push(curr)
					if !unicode.IsDigit(rune(next)) {
						if err := validateFloatLiteral(state.sequence); err != nil {
							state.addError(err.Error())
						} else {
							state.buildAndAppendToken(LIT_FLOAT)
						}
						break
					}
				}
			} else {
				state.buildAndAppendTokenFromByte(OPERATOR, curr)
			}
		case '!':
			if next == '=' {
				state.push(next)
				state.buildAndAppendToken(OPERATOR_COMPARE)
				state.lineIndex++
			} else {
				state.buildAndAppendTokenFromByte(OPERATOR_UNARY, curr)
			}
		case '<', '>':
			if next == '=' || next == curr { // <=, <<, >=, or >>
				state.push(next)
				tokenType := getTokenTypeForOperator(state.sequence)
				state.buildAndAppendToken(tokenType)
				state.lineIndex++
			} else {
				state.buildAndAppendTokenFromByte(OPERATOR_COMPARE, curr)
			}
		case '=':
			if next == '=' {
				state.push(next)
				state.buildAndAppendToken(OPERATOR_COMPARE)
				state.lineIndex++
			} else {
				state.buildAndAppendTokenFromByte(OPERATOR_ASSIGN, curr)
			}
		case '|', '&':
			if next == curr { // || or &&
				state.push(next)
				state.buildAndAppendToken(OPERATOR)
				state.lineIndex++
			} else {
				state.buildAndAppendTokenFromByte(OPERATOR_BW, curr)
			}
		case '^':
			state.buildAndAppendTokenFromByte(OPERATOR_BW, curr)
		case '"', '\'':
			state.tokenizeQuotes(line)
		default:
			if isWordStartChar(curr) {
				state.tokenizeWord(line)
			} else if unicode.IsDigit(rune(curr)) {
				state.tokenizeNumber(line)
			} else if _, ok := separators[string(curr)]; ok {
				state.buildAndAppendTokenFromByte(SEPARATOR, curr)
			} else {
				state.addError(fmt.Sprintf("Unrecognized character: '%c'", curr))
				state.clearSequence()
			}
		}
	}
}

func (state *lexerState) tokenizeQuotes(line string) {
	length := len(line)
	delim := curr
	var literal string
	var tokenType TokenType
	escaped := false
	if curr == '"' {
		literal = "string"
		tokenType = LIT_STRING
	} else {
		literal = "character"
		tokenType = LIT_CHAR
	}
	for state.lineIndex = state.startPosition; state.lineIndex < length-1; state.lineIndex++ {
		curr = line[state.lineIndex]
		next = line[state.lineIndex+1]
		if state.lineIndex != state.startPosition {
			state.push(curr)
		}
		if curr == '\\' && next == '\\' {
			escaped = !escaped
			continue
		}
		if next == delim {
			if curr != '\\' || (escaped && curr == '\\') {
				state.push(next)
				state.buildAndAppendToken(tokenType)
				state.lineIndex++
				return
			}
		}
	}
	state.addError(fmt.Sprintf("Unterminated %s literal", literal))
	state.clearSequence()
}

func (state *lexerState) tokenizeWord(line string) {
	length := len(line)
	for state.lineIndex = state.startPosition; state.lineIndex < length; state.lineIndex++ {
		curr = line[state.lineIndex]
		next = state.getNextChar(line)

		if state.lineIndex != state.startPosition {
			state.push(curr)
		}

		if !isWordChar(next) {
			tokenType := getTokenTypeForWord(state.sequence)
			state.buildAndAppendToken(tokenType)
			return
		}
	}
}

func (state *lexerState) tokenizeNumber(line string) {
	state.lineIndex = state.startPosition
	length := len(line)
	if curr == '0' && next == 'x' { // hex numbers
		state.push(next)
		state.lineIndex++
		state.tokenizeHex(line)

	} else { // int or float numbers
		in_float := false
		for ; state.lineIndex < length; state.lineIndex++ {
			curr = line[state.lineIndex]
			next = state.getNextChar(line)
			if state.lineIndex != state.startPosition {
				state.push(curr)
			}
			if next == '.' {
				if state.lineIndex == length-2 {
					state.push(next)
					err := fmt.Sprintf("Invalid float point literal: %s", state.sequence.String())
					state.addError(err)
				}
				if state.lineIndex < length-2 && line[state.lineIndex+2] == '.' { // check for .. or ..= (range operators)
					if in_float {
						if err := validateFloatLiteral(state.sequence); err != nil {
							state.addError(err.Error())
						} else {
							state.buildAndAppendToken(LIT_FLOAT)
						}
					} else {
						state.buildAndAppendToken(LIT_INT)
					}
					return
				} else {
					in_float = true
					state.push(next)
					state.lineIndex++
				}
			} else if !unicode.IsDigit(rune(next)) {
				var tokenType TokenType = LIT_INT
				if in_float {
					tokenType = LIT_FLOAT
					if err := validateFloatLiteral(state.sequence); err != nil {
						state.addError(err.Error())
						return
					}
				}
				state.buildAndAppendToken(tokenType)
				return
			}
		}
	}
}

func (state *lexerState) tokenizeHex(line string) {
	length := len(line)
	for ; state.lineIndex < length; state.lineIndex++ {
		curr = line[state.lineIndex]
		next = state.getNextChar(line)

		if state.lineIndex != state.startPosition+1 { // already processed 0 and x
			state.push(curr)
		}

		if !isHexChar(next) {
			if err := validateHexLiteral(state.sequence); err != nil {
				state.addError(err.Error())
				state.clearSequence()
			} else {
				state.buildAndAppendToken(LIT_HEX)
			}
			return
		}
	}
}

func (state *lexerState) getNextChar(line string) byte {
	if state.lineIndex == len(line)-1 {
		return 0
	}
	return line[state.lineIndex+1]
}

func isWordStartChar(chr byte) bool {
	return chr == '_' || unicode.IsLetter(rune(chr))
}

func isWordChar(chr byte) bool {
	return isWordStartChar(chr) || unicode.IsDigit(rune(chr))
}

func isHexChar(char byte) bool {
	return unicode.IsDigit(rune(char)) || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')
}

func validateHexLiteral(hexVal strings.Builder) error {
	literal := hexVal.String()
	if literal == "0x" {
		return fmt.Errorf("Incomplete hex literal: %s", literal)
	}
	if strings.ContainsAny(literal, ".") {
		return fmt.Errorf("Floating point hexadecimal literals not supported")
	}
	return nil
}

func validateFloatLiteral(floatVal strings.Builder) error {
	literal := floatVal.String()
	if literal[len(literal)-1] == '.' || strings.Count(literal, ".") > 1 {
		return fmt.Errorf("Invalid float point literal: %s", literal)
	}
	return nil
}
