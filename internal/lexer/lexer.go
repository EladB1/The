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
	hashSet map[string]struct{}
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
	curr                 byte
	next                 byte
	sequence             strings.Builder
	startPosition        int     = 0
	tokens               []Token = []Token{}
	in_string            bool    = false
	in_char              bool    = false
	in_multiline_comment bool    = false
	in_word              bool    = false
	in_int               bool    = false
	in_float             bool    = false
	in_hex               bool    = false
	in_operator          bool    = false
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

func Lex(sourceCode []string) ([]Token, diagnostic.PhaseDiagnostics) {
	var report diagnostic.PhaseDiagnostics = []diagnostic.Diagnostic{}
	var next byte
	for i, line := range sourceCode {
		// reset state
		sequence.Reset()
		startPosition = 0
		in_string = false
		in_char = false
	lineLoop:
		for col := 0; col < len(line); col++ {
			curr = line[col]
			if col == len(line)-1 {
				next = 0
			} else {
				next = line[col+1]
			}
			if in_string {
				if curr != '\\' && next == '"' {
					in_string = false
					sequence.WriteByte(curr)
					sequence.WriteByte(next)
					tokens = buildAndAppendToken(LIT_STRING, sequence.String(), i, startPosition)
					col++
					sequence.Reset()
					continue
				} else {
					sequence.WriteByte(curr)
					continue
				}
			}
			if in_char {
				if curr != '\\' && next == '\'' {
					in_char = false
					sequence.WriteByte(curr)
					sequence.WriteByte(next)
					tokens = buildAndAppendToken(LIT_CHAR, sequence.String(), i, startPosition)
					col++
					sequence.Reset()
					continue
				} else {
					sequence.WriteByte(curr)
					continue
				}
			}
			if in_multiline_comment {
				if curr == '*' && next == '/' {
					in_multiline_comment = false
					col++
					continue
				} else {
					continue
				}
			}
			if unicode.IsSpace(rune(curr)) { // TODO
				continue
			}

			if in_hex && !isHexChar(next) { // TODO: Hex validation
				if isHexChar(curr) {
					sequence.WriteByte(curr)
				}
				in_hex = false
				if err := validateHexLiteral(sequence); err != nil {
					report = append(report, diagnostic.Complain(diagnostic.SyntaxError, err.Error(), i+1, startPosition+1))
				}
				tokens = buildAndAppendToken(LIT_HEX, sequence.String(), i, startPosition)
				startPosition = 0
				sequence.Reset()
				continue
			}
			sequence.WriteByte(curr)
			switch curr {
			case '+':
				if next == '+' || next == '=' {
					sequence.WriteByte(next)
					tokens = buildAndAppendToken(OPERATOR, sequence.String(), i, col)
					col++
					sequence.Reset()
					continue
				} else {
					tokens = buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '-':
				if next == '>' {
					sequence.WriteByte(next)
					tokens = buildAndAppendToken(SEPARATOR, sequence.String(), i, col)
					col++
					sequence.Reset()
					continue
				} else if next == '-' || next == '=' {
					sequence.WriteByte(next)
					tokens = buildAndAppendToken(OPERATOR, sequence.String(), i, col)
					col++
					sequence.Reset()
					continue
				} else {
					tokens = buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '*':
				if next == '*' || next == '=' {
					sequence.WriteByte(next)
					tokens = buildAndAppendToken(OPERATOR, sequence.String(), i, col)
					col++
					sequence.Reset()
					continue
				} else {
					tokens = buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '/':
				if next == '/' {
					sequence.Reset()
					break lineLoop // skip to the next line
				} else if next == '*' {
					sequence.Reset()
					in_multiline_comment = true
					col++ // skip over next char
					continue
				} else if next == '=' {
					sequence.WriteByte(next)
					tokens = buildAndAppendToken(OPERATOR, sequence.String(), i, col)
					col++
					sequence.Reset()
					continue
				} else {
					tokens = buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '.':
				if !in_int && !in_hex && !in_float && unicode.IsDigit(rune(next)) {
					in_float = true
					startPosition = col
					sequence.WriteByte(next)
					col++
					continue
				}
				if next == '.' {
					sequence.WriteByte(next)
					// check if character after next is =
					startPosition = col
					col++
					if col < len(line)-1 {
						next = line[col+1]
						if next == '=' {
							sequence.WriteByte(next)
						}
					}
					tokens = buildAndAppendToken(OPERATOR, sequence.String(), i, startPosition)
				} else if in_int {
					in_int = false
					in_float = next != '.' // handling the case of 1..5 (range operator)
				}
				if in_float {
					if !unicode.IsDigit(rune(next)) {
						tokens = buildAndAppendToken(LIT_FLOAT, sequence.String(), i, startPosition)
					}
				} else {
					tokens = buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '!':
				if next == '=' {
					sequence.WriteByte(next)
					tokens = buildAndAppendToken(OPERATOR, sequence.String(), i, col)
					col++
					sequence.Reset()
					continue
				} else {
					tokens = buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '<':
				if next == '=' || next == '<' {
					sequence.WriteByte(next)
					tokens = buildAndAppendToken(OPERATOR, sequence.String(), i, col)
					col++
					sequence.Reset()
					continue
				} else {
					tokens = buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '>':
				if next == '=' || next == '>' {
					sequence.WriteByte(next)
					tokens = buildAndAppendToken(OPERATOR, sequence.String(), i, col)
					col++
					sequence.Reset()
					continue
				} else {
					tokens = buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '=':
				if next == '=' {
					sequence.WriteByte(next)
					tokens = buildAndAppendToken(OPERATOR, sequence.String(), i, col)
					col++
					sequence.Reset()
					continue
				} else {
					tokens = buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '|':
				if next == '|' {
					sequence.WriteByte(next)
					tokens = buildAndAppendToken(OPERATOR, sequence.String(), i, col)
					col++
					sequence.Reset()
					continue
				} else {
					tokens = buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '&':
				if next == '&' {
					sequence.WriteByte(next)
					tokens = buildAndAppendToken(OPERATOR, sequence.String(), i, col)
					col++
					sequence.Reset()
					continue
				} else {
					tokens = buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
				}
			case '"':
				startPosition = col
				in_string = true
				continue
			case '\'':
				startPosition = col
				in_char = true
				continue
			default:
				if !in_int && !in_float && !in_hex && unicode.IsDigit(rune(curr)) && curr != '0' && next != 'x' {

				} else if !in_int && !in_float && curr == '0' && next == 'x' {

				}

				if in_int && unicode.IsDigit(rune(curr)) {
					if next != '.' && !unicode.IsDigit(rune(next)) {
						in_int = false
						tokens = buildAndAppendToken(LIT_INT, sequence.String(), i, startPosition)
					}
				}

				if !in_word && isIdentifierFirstChar(curr) {
					sequence.Reset()
					sequence.WriteByte(curr)
					in_word = true
					startPosition = col
				}
				if in_word && !isIdentifierChar(next) {
					in_word = false
					// check if keyword
					var tokenType TokenType = ID
					_, ok := keywords[sequence.String()]
					if ok {
						tokenType = KEYWORD
					}
					tokens = buildAndAppendToken(tokenType, sequence.String(), i, startPosition)
					sequence.Reset()
					continue
				}
				if !in_word {
					if unicode.IsDigit(rune(curr)) {
						if !in_int && !in_hex && !in_float { // default state
							if curr == '0' || next == 'x' {
								startPosition = col
								sequence.Reset()
								sequence.WriteByte(curr)
								sequence.WriteByte(next)
								in_hex = true
								col++
							} else {
								in_int = true
								startPosition = col
							}
						} else if in_int {
							if !unicode.IsDigit(rune(next)) && next != '.' {
								in_int = false
								tokens = buildAndAppendToken(LIT_INT, sequence.String(), i, startPosition)
								sequence.Reset()
							}
						} else if in_hex {
							if !isHexChar(next) {
								in_int = false
								tokens = buildAndAppendToken(LIT_INT, sequence.String(), i, startPosition)
								sequence.Reset()
							}
						} else if in_float {
							if !unicode.IsDigit(rune(next)) {
								in_float = false
								tokens = buildAndAppendToken(LIT_FLOAT, sequence.String(), i, startPosition)
								sequence.Reset()
							}
						}
						continue
					}
					if _, ok := separators[string(curr)]; ok {
						tokens = buildAndAppendTokenFromByte(SEPARATOR, curr, i, col)
						sequence.Reset()
					} else if _, ok := operators[string(curr)]; ok {
						tokens = buildAndAppendTokenFromByte(OPERATOR, curr, i, col)
						sequence.Reset()
					} else {
						message := fmt.Sprintf("Unrecognized character: '%c'", curr)
						report = append(report, diagnostic.Complain(diagnostic.SyntaxError, message, i+1, col+1))
					}
				}
			}
		}
		// EOL actions
		sequence.Reset()
		if in_string {
			report = append(report, diagnostic.Complain(diagnostic.SyntaxError, "Unterminated string literal", i+1, startPosition+1))
		}
		if in_char {
			report = append(report, diagnostic.Complain(diagnostic.SyntaxError, "Unterminated character literal", i+1, startPosition+1))
		}
	}
	// EOF actions
	if in_multiline_comment {
		diagnostic.ReportFatalStringPositionless(diagnostic.SyntaxError, "Reached EOF while scanning for */", 1)
	}
	return tokens, report
}

func buildAndAppendToken(tokenType TokenType, value string, line int, startCol int) []Token {
	return append(tokens, Token{
		tokenType: tokenType,
		value:     value,
		line:      line + 1,
		column:    startCol + 1,
	})
}

func buildAndAppendTokenFromByte(tokenType TokenType, value byte, line int, col int) []Token {
	return buildAndAppendToken(tokenType, string(value), line, col)
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
	if hexVal.String() == "0x" {
		return fmt.Errorf("Incomplete hex literal")
	}
	return nil
}
