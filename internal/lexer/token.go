package lexer

import (
	"fmt"
	"strings"

	ds "github.com/EladB1/The/internal/datastructures"
)

type (
	TokenType string
	Token     struct {
		tokenType TokenType
		value     string
		line      int
		column    int
	}
	a int
)

const (
	ID        TokenType = "identifier"
	SEPARATOR TokenType = "separator"
	// literals
	LIT_INT    TokenType = "int literal"
	LIT_HEX    TokenType = "hex literal"
	LIT_FLOAT  TokenType = "float literal"
	LIT_STRING TokenType = "string literal"
	LIT_CHAR   TokenType = "char literal"
	// keywords
	KW_TYPE      TokenType = "type keyword"
	KW_STRUCTURE TokenType = "structure keyword"
	KW_FLOW      TokenType = "flow keyword"
	KW_OPERATOR  TokenType = "operator keyword"
	KW_MODIFIER  TokenType = "modifier keyword"
	KW_BOOLVALUE TokenType = "boolean value keyword"
	// operators
	OPERATOR         TokenType = "operator"
	OPERATOR_ADD     TokenType = "add operator"
	OPERATOR_MULT    TokenType = "multiply operator"
	OPERATOR_BW      TokenType = "bitwise operator"
	OPERATOR_COMPARE TokenType = "compare operator"
	OPERATOR_ASSIGN  TokenType = "assign operator"
	OPERATOR_RANGE   TokenType = "range operator"
	OPERATOR_UNARY   TokenType = "unary operator"
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

var (
	add_operators ds.HashSet = ds.BuildHashSet(
		"+",
		"-",
	)
	mult_operators ds.HashSet = ds.BuildHashSet(
		"*",
		"/",
		"%",
	)
	bitwise_operators ds.HashSet = ds.BuildHashSet(
		"|",
		"&",
		"^",
		"<<",
		">>",
	)
	compare_operators ds.HashSet = ds.BuildHashSet(
		">",
		">=",
		"<",
		"<=",
		"!=",
		"==",
	)
	assign_operators ds.HashSet = ds.BuildHashSet(
		"=",
		"+=",
		"-=",
		"*=",
		"/=",
	)
	range_operators ds.HashSet = ds.BuildHashSet(
		"..",
		"..=",
	)
	unary_operators ds.HashSet = ds.BuildHashSet(
		"!",
		"++",
		"--",
	)
	// any other operators that can't fit into the other categories
	operators ds.HashSet = ds.BuildHashSet(
		"**",
		"||",
		"&&",
		".",
	)
	type_keywords ds.HashSet = ds.BuildHashSet(
		"int",
		"int64",
		"uint32",
		"uint64",
		"float",
		"double",
		"String",
		"char",
		"bool",
	)
	structure_keywords ds.HashSet = ds.BuildHashSet(
		"fn",
		"struct",
		"interface",
		"for",
		"while",
		"if",
		"else",
	)
	flow_keywords ds.HashSet = ds.BuildHashSet(
		"return",
		"continue",
		"break",
	)
	operator_keywords ds.HashSet = ds.BuildHashSet(
		"in",
		"as",
	)
	modifier_keywords ds.HashSet = ds.BuildHashSet(
		"mut",
		"private",
		"impl",
	)
	bool_keywords ds.HashSet = ds.BuildHashSet(
		"true",
		"false",
	)
	separators ds.HashSet = ds.BuildHashSet(
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
)

func getTokenTypeForWord(sequence strings.Builder) TokenType {
	word := sequence.String()
	if _, ok := structure_keywords[word]; ok {
		return KW_STRUCTURE
	} else if _, ok := type_keywords[word]; ok {
		return KW_TYPE
	} else if _, ok := bool_keywords[word]; ok {
		return KW_BOOLVALUE
	} else if _, ok := flow_keywords[word]; ok {
		return KW_FLOW
	} else if _, ok := operator_keywords[word]; ok {
		return KW_OPERATOR
	} else if _, ok := modifier_keywords[word]; ok {
		return KW_MODIFIER
	} else {
		return ID
	}
}

func getTokenTypeForOperator(sequence strings.Builder) TokenType {
	operator := sequence.String()
	if _, ok := assign_operators[operator]; ok {
		return OPERATOR_ASSIGN
	} else if _, ok := add_operators[operator]; ok {
		return OPERATOR_ADD
	} else if _, ok := mult_operators[operator]; ok {
		return OPERATOR_MULT
	} else if _, ok := unary_operators[operator]; ok {
		return OPERATOR_UNARY
	} else if _, ok := bitwise_operators[operator]; ok {
		return OPERATOR_BW
	} else if _, ok := unary_operators[operator]; ok {
		return OPERATOR_UNARY
	} else if _, ok := range_operators[operator]; ok {
		return OPERATOR_RANGE
	} else {
		return OPERATOR
	}
}
