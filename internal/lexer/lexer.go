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
		"'",
		"\"",
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
	in_num               bool    = false
	in_hex               bool    = false
)

func (token Token) String() string {
	return fmt.Sprintf("{Value: \"%s\", Type: %s, Line: %d, Column: %d}", token.value, token.tokenType, token.line, token.column)
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
	for i, line := range sourceCode {
		sequence.Reset()
		startPosition = 0
		for col := 0; col < len(line)-1; col++ {
			curr = line[col]
			next = line[col+1]
			if (curr == '/' && next == '/') && !in_string && !in_multiline_comment {
				fmt.Printf("Found comment at line: %d, col: %d\n", i+1, col+1)
				break
			}
			if curr == '/' && next == '*' {
				in_multiline_comment = true
				col++ // skip over next char
				continue
			}
			if in_multiline_comment && curr != '*' && next != '/' {
				continue
			}
			if in_multiline_comment && curr == '*' && next == '/' {
				in_multiline_comment = false
				col++
				fmt.Printf("End multiline comment at line: %d, col: %d\n", i+1, col+1)
				continue
			}

			if unicode.IsSpace(rune(curr)) { // TODO
				continue
			}
			if _, ok := separators[string(curr)]; ok {
				tokens = buildAndAppendTokenFromByte(SEPARATOR, curr, i, col)
				continue
			}
			if sequence.String() == "->" {
				tokens = buildAndAppendToken(SEPARATOR, sequence.String(), i, col-1)
				sequence.Reset()
				//continue
			}
			sequence.WriteByte(curr)
			if in_hex && (!unicode.IsDigit(rune(next)) && (next < 'a' || next > 'f') && (next < 'A' || next > 'F')) { // TODO: Hex validation
				in_hex = false
				tokens = buildAndAppendToken(LIT_HEX, sequence.String(), i, startPosition)
				startPosition = col + 1
				sequence.Reset()
				continue

			}
			if unicode.IsDigit(rune(curr)) && curr != '0' && next != 'x' {
				// int logic
				// float logic
				// validate any invalid characters
			}
			if curr == '0' && next == 'x' {
				startPosition = col
				fmt.Println(sequence.String())
				sequence.WriteByte(next)
				in_hex = true
				col++
				continue
			}
			if !in_word && isIdentifierFirstChar(curr) {
				sequence.Reset()
				sequence.WriteByte(curr)
				in_word = true
			}
			if in_word && !isIdentifierChar(next) {
				in_word = false
				// check if keyword
				var tokenType TokenType = ID
				_, ok := keywords[sequence.String()]
				if ok {
					tokenType = KEYWORD
				}
				tokens = append(tokens, Token{
					tokenType: tokenType,
					value:     sequence.String(),
					line:      i + 1,
					column:    col + 1,
				})
				sequence.Reset()
			}
		}
	}
	// EOF actions
	if in_multiline_comment {
		diagnostic.ReportFatalStringPositionless(diagnostic.SyntaxError, "Reached EOF while scanning for */", 1)
	}
	PrintTokens(tokens)
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
