package lexer

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/EladB1/The/internal/diagnostic"
)

type (
	TokenType string
	Token     struct {
		tokenType TokenType
		value     string
		line      int
		column    int
	}
	hashSet    map[string]struct{}
	lexerState struct {
		sequence             strings.Builder
		startPosition        int
		tokens               []Token
		in_string            bool
		in_char              bool
		in_multiline_comment bool
		in_word              bool
		in_int               bool
		in_float             bool
		in_hex               bool
	}
)

const (
	ID        TokenType = "identifier"
	KEYWORD   TokenType = "keyword"
	OPERATOR  TokenType = "operator"
	SEPARATOR TokenType = "separator"
	// literals
	LIT_INT    TokenType = "int literal"
	LIT_HEX    TokenType = "hex literal"
	LIT_FLOAT  TokenType = "float literal"
	LIT_STRING TokenType = "string literal"
	LIT_CHAR   TokenType = "char literal"
)

func buildHashSet(items ...string) hashSet {
	set := make(map[string]struct{})
	for _, item := range items {
		set[item] = struct{}{}
	}
	return set
}

var (
	operators hashSet = buildHashSet(
		"+",
		"-",
		"*",
		"/",
		"%",
		"**",
		"++",
		"--",
		"!",
		"||",
		"&&",
		"|",
		"&",
		"^",
		"<<",
		">>",
		">",
		">=",
		"<",
		"<=",
		"!=",
		"==",
		"=",
		"+=",
		"-=",
		"*=",
		"/=",
		"..",
		"..=",
		".",
	)
	keywords hashSet = buildHashSet(
		"fn",
		"mut",
		"private",
		"in",
		"as",
		"int",
		"int64",
		"uint32",
		"uint64",
		"float",
		"double",
		"String",
		"char",
		"bool",
		"struct",
		"interface",
		"impl",
		"if",
		"else",
		"for",
		"while",
		"true",
		"false",
		"return",
		"continue",
		"break",
	)
	separators hashSet = buildHashSet(
		"(",
		")",
		"{",
		"}",
		"[",
		"]",
		";",
		":",
		",",
		"->",
	)
	curr byte
	next byte
)

func (token Token) String() string {
	return fmt.Sprintf("{Value: %s, Type: %s, Line: %d, Column: %d}", token.value, token.tokenType, token.line, token.column)
}

func (token Token) HasValue(value string) bool {
	return token.value == value
}

func PrintTokens(tokens []Token) {
	for _, token := range tokens {
		fmt.Println(token)
	}
}

func (stateMchn *lexerState) reset() {
	stateMchn.sequence.Reset()
	stateMchn.in_char = false
	stateMchn.in_string = false
	stateMchn.in_int = false
	stateMchn.in_hex = false
	stateMchn.in_float = false
	stateMchn.in_word = false
	// manually need to reset in_multiline_comment
	// do not reset tokens
}

func (stateMchn *lexerState) push(char byte) {
	stateMchn.sequence.WriteByte(char)
}

func (stateMchn *lexerState) clearSequence() {
	stateMchn.sequence.Reset()
}

func (stateMchn *lexerState) buildAndAppendToken(tokenType TokenType, line int, startCol int) {
	stateMchn.tokens = append(stateMchn.tokens, Token{
		tokenType: tokenType,
		value:     stateMchn.sequence.String(),
		line:      line + 1,
		column:    startCol + 1,
	})
}

func (stateMchn *lexerState) buildAndAppendTokenFromByte(tokenType TokenType, char byte, line int, startCol int) {
	stateMchn.tokens = append(stateMchn.tokens, Token{
		tokenType: tokenType,
		value:     string(char),
		line:      line + 1,
		column:    startCol + 1,
	})
}

func (stateMchn *lexerState) debug() {
	fmt.Printf("State: {Sequence: %s, position: %d, flags: {in_char: %v, in_string: %v, in_word: %v, in_hex: %v, in_int: %v, in_float: %v, in_multiline_comment: %v}}\n",
		stateMchn.sequence.String(),
		stateMchn.startPosition,
		stateMchn.in_char,
		stateMchn.in_string,
		stateMchn.in_word,
		stateMchn.in_hex,
		stateMchn.in_int,
		stateMchn.in_float,
		stateMchn.in_multiline_comment,
	)
}

func Lex(sourceCode []string) ([]Token, diagnostic.PhaseDiagnostics) {
	var report diagnostic.PhaseDiagnostics = []diagnostic.Diagnostic{}
	state := &lexerState{
		tokens:               []Token{},
		sequence:             strings.Builder{},
		startPosition:        0,
		in_string:            false,
		in_char:              false,
		in_multiline_comment: false,
		in_word:              false,
		in_int:               false,
		in_float:             false,
		in_hex:               false,
	}
	for i, line := range sourceCode {
		state.reset()
	lineLoop:
		for col := 0; col < len(line); col++ {
			//state.debug()
			curr = line[col]
			if col == len(line)-1 {
				next = 0
			} else {
				next = line[col+1]
			}
			if state.in_string {
				if curr != '\\' && next == '"' {
					state.in_string = false
					state.push(curr)
					state.push(next)
					state.buildAndAppendToken(LIT_STRING, i, state.startPosition)
					col++
					state.clearSequence()
					continue
				} else {
					state.push(curr)
					continue
				}
			}
			if state.in_char {
				if curr != '\\' && next == '\'' {
					state.in_char = false
					state.push(curr)
					state.push(next)
					state.buildAndAppendToken(LIT_CHAR, i, state.startPosition)
					col++
					state.clearSequence()
					continue
				} else {
					state.push(curr)
					continue
				}
			}
			if state.in_multiline_comment {
				if curr == '*' && next == '/' {
					state.in_multiline_comment = false
					state.reset()
					col++
					continue
				} else {
					continue
				}
			}
			if unicode.IsSpace(rune(curr)) { // TODO
				continue
			}

			state.push(curr)
			switch curr {
			case '+':
				if next == '+' || next == '=' {
					state.push(next)
					state.buildAndAppendToken(OPERATOR, i, col)
					col++
					state.clearSequence()
					continue
				} else {
					state.buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '-':
				if next == '>' {
					state.push(next)
					state.buildAndAppendToken(SEPARATOR, i, col)
					col++
					state.clearSequence()
					continue
				} else if next == '-' || next == '=' {
					state.push(next)
					state.buildAndAppendToken(OPERATOR, i, col)
					col++
					state.clearSequence()
					continue
				} else {
					state.buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '*':
				if next == '*' || next == '=' {
					state.push(next)
					state.buildAndAppendToken(OPERATOR, i, col)
					col++
					state.clearSequence()
					continue
				} else {
					state.buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '/':
				if next == '/' {
					state.clearSequence()
					break lineLoop // skip to the next line
				} else if next == '*' {
					state.clearSequence()
					state.in_multiline_comment = true
					col++ // skip over next char
					continue
				} else if next == '=' {
					state.push(next)
					state.buildAndAppendToken(OPERATOR, i, col)
					col++
					state.clearSequence()
					continue
				} else {
					state.buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '.':
				if !state.in_int && !state.in_hex && !state.in_float && unicode.IsDigit(rune(next)) {
					state.in_float = true
					state.startPosition = col
					state.push(next)
					col++
					continue
				}
				if next == '.' {
					if state.in_hex {
						state.in_hex = false
						if err := validateHexLiteral(state.sequence); err != nil {
							report = append(report, diagnostic.Complain(diagnostic.SyntaxError, err.Error(), i+1, state.startPosition+1))
						} else {
							state.buildAndAppendToken(LIT_HEX, i, state.startPosition)
						}
						state.clearSequence()
						state.push(curr)
					}
					state.push(next)
					// check if character after next is =
					state.startPosition = col
					if col < len(line)-1 {
						next = line[col+2]
						if next == '=' {
							state.push(next)
							col++
						}
					}
					col++
					state.buildAndAppendToken(OPERATOR, i, state.startPosition)
					state.clearSequence()
					continue
				} else if state.in_hex {
					if err := validateHexLiteral(state.sequence); err != nil {
						report = append(report, diagnostic.Complain(diagnostic.SyntaxError, err.Error(), i+1, state.startPosition+1))
					}
					continue
				} else if state.in_int {
					state.in_int = false
					state.in_float = next != '.' // handling the case of 1..5 (range operator)
				}
				if state.in_float {
					if !unicode.IsDigit(rune(next)) {
						if err := validateFloatLiteral(state.sequence); err != nil {
							report = append(report, diagnostic.Complain(diagnostic.SyntaxError, err.Error(), i, state.startPosition))
						}
						state.buildAndAppendToken(LIT_FLOAT, i, state.startPosition)
					}
				} else {
					state.buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '!':
				if next == '=' {
					state.push(next)
					state.buildAndAppendToken(OPERATOR, i, col)
					col++
					state.clearSequence()
					continue
				} else {
					state.buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '<':
				if next == '=' || next == '<' {
					state.push(next)
					state.buildAndAppendToken(OPERATOR, i, col)
					col++
					state.clearSequence()
					continue
				} else {
					state.buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '>':
				if next == '=' || next == '>' {
					state.push(next)
					state.buildAndAppendToken(OPERATOR, i, col)
					col++
					state.clearSequence()
					continue
				} else {
					state.buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '=':
				if next == '=' {
					state.push(next)
					state.buildAndAppendToken(OPERATOR, i, col)
					col++
					state.clearSequence()
					continue
				} else {
					state.buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '|':
				if next == '|' {
					state.push(next)
					state.buildAndAppendToken(OPERATOR, i, col)
					col++
					state.clearSequence()
					continue
				} else {
					state.buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '&':
				if next == '&' {
					state.push(next)
					state.buildAndAppendToken(OPERATOR, i, col)
					col++
					state.clearSequence()
					continue
				} else {
					state.buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '"':
				state.startPosition = col
				state.in_string = true
				continue
			case '\'':
				state.startPosition = col
				state.in_char = true
				continue
			default:
				if !state.in_hex && !state.in_word && isIdentifierFirstChar(curr) {
					state.clearSequence()
					state.push(curr)
					state.in_word = true
					state.startPosition = col
				}
				if state.in_word && !isIdentifierChar(next) {
					state.in_word = false
					// check if keyword
					var tokenType TokenType = ID
					_, ok := keywords[state.sequence.String()]
					if ok {
						tokenType = KEYWORD
					}
					state.buildAndAppendToken(tokenType, i, state.startPosition)
					state.clearSequence()
					continue
				}
				if state.in_hex {
					if !isHexChar(next) {
						state.in_hex = false
						if err := validateHexLiteral(state.sequence); err != nil {
							report = append(report, diagnostic.Complain(diagnostic.SyntaxError, err.Error(), i+1, state.startPosition+1))
						}

						state.buildAndAppendToken(LIT_HEX, i, state.startPosition)
						state.clearSequence()
						continue
					}
				}
				if !state.in_word && !state.in_hex {
					if unicode.IsDigit(rune(curr)) {
						if !state.in_int && !state.in_hex && !state.in_float { // default state
							if curr == '0' || next == 'x' {
								state.startPosition = col
								state.clearSequence()
								state.push(curr)
								state.push(next)
								state.in_hex = true
								col++
							} else {
								state.in_int = true
								state.startPosition = col
							}
						} else if state.in_int {
							if !unicode.IsDigit(rune(next)) && next != '.' {
								state.in_int = false
								state.buildAndAppendToken(LIT_INT, i, state.startPosition)
								state.clearSequence()
							}
						} else if state.in_float {
							if !unicode.IsDigit(rune(next)) {
								state.in_float = false
								if err := validateFloatLiteral(state.sequence); err != nil {
									report = append(report, diagnostic.Complain(diagnostic.SyntaxError, err.Error(), i, state.startPosition))
								}
								state.buildAndAppendToken(LIT_FLOAT, i, state.startPosition)
								state.clearSequence()
							}
						}
					} else if _, ok := separators[string(curr)]; ok {
						state.buildAndAppendTokenFromByte(SEPARATOR, curr, i, col)
						state.clearSequence()
					} else if _, ok := operators[string(curr)]; ok {
						state.buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
						state.clearSequence()
					} else {
						message := fmt.Sprintf("Unrecognized character: '%c'", curr)
						report = append(report, diagnostic.Complain(diagnostic.SyntaxError, message, i+1, col+1))
					}
				}
			}
		}
		// EOL actions
		state.clearSequence()
		if state.in_string {
			report = append(report, diagnostic.Complain(diagnostic.SyntaxError, "Unterminated string literal", i+1, state.startPosition+1))
		}
		if state.in_char {
			report = append(report, diagnostic.Complain(diagnostic.SyntaxError, "Unterminated character literal", i+1, state.startPosition+1))
		}
	}
	// EOF actions
	if state.in_multiline_comment {
		diagnostic.ReportFatalStringPositionless(diagnostic.SyntaxError, "Reached EOF while scanning for */", 1)
	}
	return state.tokens, report
}

func isIdentifierFirstChar(chr byte) bool {
	return chr == '_' || unicode.IsLetter(rune(chr))
}

func isIdentifierChar(chr byte) bool {
	return isIdentifierFirstChar(chr) || unicode.IsDigit(rune(chr))
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
