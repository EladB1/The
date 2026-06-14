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
		in_multiline_comment bool
		in_word              bool
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

	errLevel diagnostic.Severity = diagnostic.SyntaxError
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
	stateMchn.clearSequence()
}

func (stateMchn *lexerState) buildAndAppendTokenFromByte(tokenType TokenType, char byte, line int, startCol int) {
	stateMchn.tokens = append(stateMchn.tokens, Token{
		tokenType: tokenType,
		value:     string(char),
		line:      line + 1,
		column:    startCol + 1,
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

func Lex(sourceCode []string) ([]Token, diagnostic.PhaseDiagnostics) {
	var report diagnostic.PhaseDiagnostics = diagnostic.PhaseDiagnostics{}
	state := &lexerState{
		tokens:               []Token{},
		sequence:             strings.Builder{},
		startPosition:        0,
		in_multiline_comment: false,
	}
	for i, line := range sourceCode {
		state.reset()
		length := len(line)
	lineLoop:
		for col := 0; col < length; col++ {
			// if !state.in_multiline_comment {
			// 	state.debug()
			// }
			curr = line[col]
			if col == length-1 {
				next = 0
			} else {
				next = line[col+1]
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
			if unicode.IsSpace(rune(curr)) {
				continue
			}

			state.push(curr)
			switch curr {
			case '+':
				if next == '+' || next == '=' {
					state.push(next)
					state.buildAndAppendToken(OPERATOR, i, col)
					col++
					continue
				} else {
					state.buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '-':
				if next == '>' {
					state.push(next)
					state.buildAndAppendToken(SEPARATOR, i, col)
					col++
					continue
				} else if next == '-' || next == '=' {
					state.push(next)
					state.buildAndAppendToken(OPERATOR, i, col)
					col++
					continue
				} else {
					state.buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '*':
				if next == '*' || next == '=' {
					state.push(next)
					state.buildAndAppendToken(OPERATOR, i, col)
					col++
					continue
				} else {
					state.buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '/':
				if next == '/' {
					state.clearSequence()
					i++
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
					continue
				} else {
					state.buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '.':
				if next == '.' { // .. or ..=
					state.push(next)
					// check if character after next is =
					state.startPosition = col
					if col < length-1 {
						next = line[col+2]
						if next == '=' {
							state.push(next)
							col++
						}
					}
					col++
					state.buildAndAppendToken(OPERATOR, i, state.startPosition)
					continue
				} else if unicode.IsDigit(rune(next)) && !(curr == '0' && next == 'x') { // Example: .234
					state.startPosition = col
					col++
					for col < length {
						curr = line[col]
						if col == length-1 {
							next = 0
						} else {
							next = line[col+1]
						}
						state.push(curr)
						if !unicode.IsDigit(rune(next)) {
							if err := validateFloatLiteral(state.sequence); err != nil {
								report = report.Complain(errLevel, err.Error(), i, state.startPosition)
								break
							}
							state.buildAndAppendToken(LIT_FLOAT, i, state.startPosition)
							break
						}
						col++
					}
				} else {
					state.buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '!':
				if next == '=' {
					state.push(next)
					state.buildAndAppendToken(OPERATOR, i, col)
					col++
					continue
				} else {
					state.buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '<':
				if next == '=' || next == '<' {
					state.push(next)
					state.buildAndAppendToken(OPERATOR, i, col)
					col++
					continue
				} else {
					state.buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '>':
				if next == '=' || next == '>' {
					state.push(next)
					state.buildAndAppendToken(OPERATOR, i, col)
					col++
					continue
				} else {
					state.buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '=':
				if next == '=' {
					state.push(next)
					state.buildAndAppendToken(OPERATOR, i, col)
					col++
					continue
				} else {
					state.buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
				continue
			case '|':
				if next == '|' {
					state.push(next)
					state.buildAndAppendToken(OPERATOR, i, col)
					col++
					continue
				} else {
					state.buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '&':
				if next == '&' {
					state.push(next)
					state.buildAndAppendToken(OPERATOR, i, col)
					col++
					continue
				} else {
					state.buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '"':
				state.startPosition = col
				for col < length-1 {
					curr = line[col]
					next = line[col+1]
					if col != state.startPosition {
						state.push(curr)
					}
					if curr != '\\' && next == '"' {
						state.push(next)
						state.buildAndAppendToken(LIT_STRING, i, state.startPosition)
						col++
						continue lineLoop
					}
					col++
				}
				curr = line[col]
				if col == length-1 {
					next = 0
				} else {
					next = line[col+1]
				}
				if curr != '\\' && next == '"' {
					state.push(curr)
					state.push(next)
					state.buildAndAppendToken(LIT_STRING, i, state.startPosition)
					col++
					continue
				} else {
					report = report.Complain(errLevel, "Unterminated string literal", i, state.startPosition)
					state.clearSequence()
				}
			case '\'':
				state.startPosition = col
				for col < length-1 {
					curr = line[col]
					next = line[col+1]

					if col != state.startPosition {
						state.push(curr)
					}

					if curr != '\\' && next == '\'' {
						state.push(next)
						state.buildAndAppendToken(LIT_CHAR, i, state.startPosition)
						col++
						continue lineLoop
					}
					col++
				}
				curr = line[col]
				if col == length-1 {
					next = 0
				} else {
					next = line[col+1]
				}
				if curr != '\\' && next == '\'' {
					state.push(curr)
					state.push(next)
					state.buildAndAppendToken(LIT_CHAR, i, state.startPosition)
					col++
					continue
				} else {
					report = report.Complain(errLevel, "Unterminated character literal", i, state.startPosition)
					state.clearSequence()
				}
			default:
				if isWordStartChar(curr) { // identifiers and keywords
					state.startPosition = col
					var tokenType TokenType = ID
					for col < length {
						curr = line[col]
						if col == length-1 {
							next = 0
						} else {
							next = line[col+1]
						}

						if col != state.startPosition {
							state.push(curr)
						}

						if !isWordChar(next) {
							_, ok := keywords[state.sequence.String()]
							if ok {
								tokenType = KEYWORD
							}
							state.buildAndAppendToken(tokenType, i, state.startPosition)
							break
						}
						col++
					}
				} else if unicode.IsDigit(rune(curr)) { // number literals
					state.startPosition = col
					if curr == '0' && next == 'x' { // hex numbers
						state.push(next)
						col++
						for col < length {
							curr = line[col]
							if col == length-1 {
								next = 0
							} else {
								next = line[col+1]
							}

							if col != state.startPosition+1 {
								state.push(curr)
							}

							if col == length-1 || !isHexChar(next) {
								if err := validateHexLiteral(state.sequence); err != nil {
									report = report.Complain(errLevel, err.Error(), i, state.startPosition)
								}
								state.buildAndAppendToken(LIT_HEX, i, state.startPosition)
								break
							}

							col++
						}
					} else { // int or float numbers
						in_float := false
						for col < length {
							curr = line[col]
							if col == length-1 {
								next = 0
							} else {
								next = line[col+1]
							}
							if col != state.startPosition {
								state.push(curr)
							}
							if next == '.' {
								if col == length-2 {
									state.push(next)
									err := fmt.Sprintf("Invalid float point literal: %s", state.sequence.String())
									report = report.Complain(errLevel, err, i, state.startPosition)
								}
								if col < length-2 && line[col+2] == '.' {
									if in_float {
										if err := validateFloatLiteral(state.sequence); err != nil {
											report = report.Complain(errLevel, err.Error(), i, state.startPosition)
										}
										state.buildAndAppendToken(LIT_FLOAT, i, state.startPosition)

									} else {
										state.buildAndAppendToken(LIT_INT, i, state.startPosition)
									}
									break
								} else {
									in_float = true
									state.push(next)
									col++
								}
							}
							if !unicode.IsDigit(rune(next)) && next != '.' {
								var tokenType TokenType = LIT_INT
								if in_float {
									tokenType = LIT_FLOAT
									if err := validateFloatLiteral(state.sequence); err != nil {
										report = report.Complain(errLevel, err.Error(), i, state.startPosition)
									}
								}
								state.buildAndAppendToken(tokenType, i, state.startPosition)
								break
							}
							col++
						}
					}
				} else if _, ok := separators[string(curr)]; ok {
					state.buildAndAppendTokenFromByte(SEPARATOR, curr, i, col)
				} else if _, ok := operators[string(curr)]; ok {
					state.buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				} else {
					message := fmt.Sprintf("Unrecognized character: '%c'", curr)
					report = report.Complain(errLevel, message, i, col)
					state.clearSequence()
				}
			}
		}
	}
	// EOF actions
	if state.in_multiline_comment {
		report = report.ComplainPositionless(errLevel, "Reached EOF while scanning for */")
	}
	return state.tokens, report
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
