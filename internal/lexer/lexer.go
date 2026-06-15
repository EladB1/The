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
		messages             diagnostic.PhaseDiagnostics
		lineNum              int
		lineIndex            int
		in_multiline_comment bool
		in_word              bool
	}
)

func buildHashSet(items ...string) hashSet {
	set := make(map[string]struct{})
	for _, item := range items {
		set[item] = struct{}{}
	}
	return set
}

func (token Token) String() string {
	return fmt.Sprintf("{Value: %s, Type: %s, Line: %d, Column: %d}", token.value, token.tokenType, token.line, token.column)
}

func (token Token) HasValue(value string) bool {
	return token.value == value
}

func (stateMchn *lexerState) reset() {
	stateMchn.sequence.Reset()
	stateMchn.startPosition = 0
	stateMchn.lineIndex = 0
	// manually need to reset in_multiline_comment
	// do not reset tokens
}

func (stateMchn *lexerState) addError(message string, lineIndex int) {
	stateMchn.messages = stateMchn.messages.Complain(errLevel, message, stateMchn.lineNum, lineIndex)
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

	errLevel diagnostic.Severity = diagnostic.SyntaxError
)

var (
	add_operators hashSet = buildHashSet(
		"+",
		"-",
	)
	mult_operators hashSet = buildHashSet(
		"*",
		"/",
		"%",
	)
	bitwise_operators hashSet = buildHashSet(
		"|",
		"&",
		"^",
		"<<",
		">>",
	)
	compare_operators hashSet = buildHashSet(
		">",
		">=",
		"<",
		"<=",
		"!=",
		"==",
	)
	assign_operators hashSet = buildHashSet(
		"=",
		"+=",
		"-=",
		"*=",
		"/=",
	)
	range_operators hashSet = buildHashSet(
		"..",
		"..=",
	)
	unary_operators hashSet = buildHashSet(
		"!",
		"++",
		"--",
	)
	// any other operators that can't fit into the other categories
	operators hashSet = buildHashSet(
		"**",
		"||",
		"&&",
		".",
	)
	type_keywords hashSet = buildHashSet(
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
	structure_keywords hashSet = buildHashSet(
		"fn",
		"struct",
		"interface",
		"for",
		"while",
		"if",
		"else",
	)
	flow_keywords hashSet = buildHashSet(
		"return",
		"continue",
		"break",
	)
	operator_keywords hashSet = buildHashSet(
		"in",
		"as",
	)
	modifier_keywords hashSet = buildHashSet(
		"mut",
		"private",
		"impl",
	)
	bool_keywords hashSet = buildHashSet(
		"true",
		"false",
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
	lines := len(sourceCode)
	for ; state.lineNum < lines; state.lineNum++ {
		state.reset()
		line := sourceCode[state.lineNum]
		state.lexLine(line)
	}
	// EOF actions
	if state.in_multiline_comment {
		state.messages = state.messages.ComplainPositionless(errLevel, "Reached EOF while scanning for */")
	}
	return state.tokens, state.messages
}

func (state *lexerState) lexLine(line string) {
	length := len(line)
	for i := 0; i < length; i++ {
		curr = line[i]
		if i == length-1 {
			next = 0
		} else {
			next = line[i+1]
		}
		if state.in_multiline_comment {
			if curr == '*' && next == '/' {
				state.in_multiline_comment = false
				state.reset()
				i++
			}
			continue
		}
		if unicode.IsSpace(rune(curr)) {
			continue
		}
		//fmt.Printf("curr: %c, next: %c\n", curr, next)
		state.push(curr)
		switch curr {
		case '+':
			if next == '+' || next == '=' {
				state.push(next)
				tokenType := getTokenTypeForOperator(state.sequence)
				state.buildAndAppendToken(tokenType, state.lineNum, i)
				i++
			} else {
				state.buildAndAppendTokenFromByte(OPERATOR_ADD, curr, state.lineNum, i)
			}
		case '-':
			switch next {
			case '>':
				state.push(next)
				state.buildAndAppendToken(SEPARATOR, state.lineNum, i)
				i++
			case '-', '=':
				state.push(next)
				tokenType := getTokenTypeForOperator(state.sequence)
				state.buildAndAppendToken(tokenType, state.lineNum, i)
				i++
			default:
				state.buildAndAppendTokenFromByte(OPERATOR_ADD, curr, state.lineNum, i)
			}
		case '*':
			if next == '*' || next == '=' {
				state.push(next)
				tokenType := getTokenTypeForOperator(state.sequence)
				state.buildAndAppendToken(tokenType, state.lineNum, i)
				i++
			} else {
				state.buildAndAppendTokenFromByte(OPERATOR_MULT, curr, state.lineNum, i)
			}
		case '/':
			switch next {
			case '/':
				i++
				state.clearSequence()
				return // skip to the next line
			case '*':
				i++
				state.clearSequence()
				state.in_multiline_comment = true
			case '=':
				state.push(next)
				state.buildAndAppendToken(OPERATOR_ASSIGN, state.lineNum, i)
				i++
			default:
				state.buildAndAppendTokenFromByte(OPERATOR_MULT, curr, state.lineNum, i)
			}
		case '%':
			state.buildAndAppendTokenFromByte(OPERATOR_MULT, curr, state.lineNum, i)
		case '.':
			if next == '.' { // .. or ..=
				state.push(next)
				state.startPosition = i
				if i < length-2 { // check if character after next is =
					next = line[i+2]
					if next == '=' {
						state.push(next)
						i++
					}
				}
				i++
				state.buildAndAppendToken(OPERATOR_RANGE, state.lineNum, state.startPosition)
			} else if unicode.IsDigit(rune(next)) && !(curr == '0' && next == 'x') { // Example: .234
				state.startPosition = i
				i++
				for i < length {
					curr = line[i]
					if i == length-1 {
						next = 0
					} else {
						next = line[i+1]
					}
					state.push(curr)
					if !unicode.IsDigit(rune(next)) {
						if err := validateFloatLiteral(state.sequence); err != nil {
							state.addError(err.Error(), state.startPosition)
						} else {
							state.buildAndAppendToken(LIT_FLOAT, state.lineNum, state.startPosition)
						}
						break
					}
					i++
				}
			} else {
				state.buildAndAppendTokenFromByte(OPERATOR, curr, state.lineNum, i)
			}
		case '!':
			if next == '=' {
				state.push(next)
				state.buildAndAppendToken(OPERATOR_COMPARE, state.lineNum, i)
				i++
			} else {
				state.buildAndAppendTokenFromByte(OPERATOR_UNARY, curr, state.lineNum, i)
			}
		case '<', '>':
			if next == '=' || next == curr { // <=, <<, >=, or >>
				state.push(next)
				tokenType := getTokenTypeForOperator(state.sequence)
				state.buildAndAppendToken(tokenType, state.lineNum, i)
				i++
			} else {
				state.buildAndAppendTokenFromByte(OPERATOR_COMPARE, curr, state.lineNum, i)
			}
		case '=':
			if next == '=' {
				state.push(next)
				state.buildAndAppendToken(OPERATOR_COMPARE, state.lineNum, i)
				i++
				continue
			} else {
				state.buildAndAppendTokenFromByte(OPERATOR_ASSIGN, curr, state.lineNum, i)
			}
		case '|', '&':
			if next == curr { // || or &&
				state.push(next)
				state.buildAndAppendToken(OPERATOR, state.lineNum, i)
				i++
				continue
			} else {
				state.buildAndAppendTokenFromByte(OPERATOR_BW, curr, state.lineNum, i)
			}
		case '^':
			state.buildAndAppendTokenFromByte(OPERATOR_BW, curr, state.lineNum, i)
		case '"', '\'':
			state.startPosition = i
			state.tokenizeQuotes(line)
			i = state.lineIndex
		default:
			if isWordStartChar(curr) {
				state.startPosition = i
				state.tokenizeWord(line)
				i = state.lineIndex
			} else if unicode.IsDigit(rune(curr)) {
				state.startPosition = i
				state.tokenizeNumber(line)
				i = state.lineIndex
			} else if _, ok := separators[string(curr)]; ok {
				state.buildAndAppendTokenFromByte(SEPARATOR, curr, state.lineNum, i)
			} else {
				state.addError(fmt.Sprintf("Unrecognized character: '%c'", curr), i)
				state.clearSequence()
			}
		}
	}
}

func (state *lexerState) tokenizeQuotes(line string) {
	state.lineIndex = state.startPosition
	length := len(line)
	delim := curr
	var literal string
	var tokenType TokenType
	end := false
	if curr == '"' {
		literal = "string"
		tokenType = LIT_STRING
	} else {
		literal = "character"
		tokenType = LIT_CHAR
	}
	for ; state.lineIndex < length-1; state.lineIndex++ {
		curr = line[state.lineIndex]
		next = line[state.lineIndex+1]
		if state.lineIndex != state.startPosition {
			state.push(curr)
		}
		if curr != '\\' && next == delim {
			state.push(next)
			state.buildAndAppendToken(tokenType, state.lineNum, state.startPosition)
			state.lineIndex++
			end = true
			break
		}
	}
	if !end {
		state.addError(fmt.Sprintf("Unterminated %s literal", literal), state.startPosition)
		state.clearSequence()
	}
}

func (state *lexerState) tokenizeWord(line string) {
	state.lineIndex = state.startPosition
	length := len(line)
	for ; state.lineIndex < length; state.lineIndex++ {
		curr = line[state.lineIndex]
		if state.lineIndex == length-1 {
			next = 0
		} else {
			next = line[state.lineIndex+1]
		}

		if state.lineIndex != state.startPosition {
			state.push(curr)
		}

		if !isWordChar(next) {
			tokenType := getTokenTypeForWord(state.sequence)
			state.buildAndAppendToken(tokenType, state.lineNum, state.startPosition)
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
		for state.lineIndex < length {
			curr = line[state.lineIndex]
			if state.lineIndex == length-1 {
				next = 0
			} else {
				next = line[state.lineIndex+1]
			}
			if state.lineIndex != state.startPosition {
				state.push(curr)
			}
			if next == '.' {
				if state.lineIndex == length-2 {
					state.push(next)
					err := fmt.Sprintf("Invalid float point literal: %s", state.sequence.String())
					state.addError(err, state.startPosition)
				}
				if state.lineIndex < length-2 && line[state.lineIndex+2] == '.' { // check for .. or ..= (range operators)
					if in_float {
						if err := validateFloatLiteral(state.sequence); err != nil {
							state.addError(err.Error(), state.startPosition)
						}
						state.buildAndAppendToken(LIT_FLOAT, state.lineNum, state.startPosition)

					} else {
						state.buildAndAppendToken(LIT_INT, state.lineNum, state.startPosition)
					}
					break
				} else {
					in_float = true
					state.push(next)
					state.lineIndex++
				}
			}
			if !unicode.IsDigit(rune(next)) && next != '.' {
				var tokenType TokenType = LIT_INT
				if in_float {
					tokenType = LIT_FLOAT
					if err := validateFloatLiteral(state.sequence); err != nil {
						state.addError(err.Error(), state.startPosition)
					}
				}
				state.buildAndAppendToken(tokenType, state.lineNum, state.startPosition)
				break
			}
			state.lineIndex++
		}
	}
}

func (state *lexerState) tokenizeHex(line string) {
	length := len(line)
	for state.lineIndex < length {
		curr = line[state.lineIndex]
		if state.lineIndex == length-1 {
			next = 0
		} else {
			next = line[state.lineIndex+1]
		}

		if state.lineIndex != state.startPosition+1 {
			state.push(curr)
		}

		if state.lineIndex == length-1 || !isHexChar(next) {
			if err := validateHexLiteral(state.sequence); err != nil {
				state.addError(err.Error(), state.startPosition)
			}
			state.buildAndAppendToken(LIT_HEX, state.lineNum, state.startPosition)
			break
		}

		state.lineIndex++
	}
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

func PrintTokens(tokens []Token) {
	for _, token := range tokens {
		fmt.Println(token)
	}
}
