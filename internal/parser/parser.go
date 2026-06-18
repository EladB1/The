package parser

import (
	"fmt"

	"github.com/EladB1/The/internal/diagnostic"
	"github.com/EladB1/The/internal/lexer"
)

var (
	messages diagnostic.PhaseDiagnostics = diagnostic.PhaseDiagnostics{}
	errLevel diagnostic.Severity         = diagnostic.SyntaxError
	tokens   []lexer.Token               = []lexer.Token{}
	length   int                         = 0
	ptr      int                         = 0
	curr     lexer.Token                 = lexer.Token{}
	next     lexer.Token                 = lexer.Token{}
)

func getToken() lexer.Token {
	return tokens[ptr]
}

func peekToken() lexer.Token {
	if ptr == len(tokens)-1 {
		return lexer.Token{}
	}
	return tokens[ptr+1]
}

func movePtr() {
	ptr++
	if ptr == len(tokens) {
		return
	} else if ptr == len(tokens)-1 {
		curr = tokens[ptr]
		next = lexer.Token{}
	} else {
		curr = tokens[ptr]
		next = tokens[ptr+1]
	}
}

func movePtrN(n int) {
	ptr += n
	if ptr >= len(tokens) {
		return
	} else if ptr == len(tokens)-n {
		curr = tokens[ptr]
		next = lexer.Token{}
	} else {
		curr = tokens[ptr]
		next = tokens[ptr+1]
	}
}

func complainAboutToken(message string, token lexer.Token) {
	messages = messages.Complain(errLevel, message, token.Line, token.Column)
}

/*
 *	program = { declaration } ;
 */
func Parse(lexerTokens []lexer.Token) (AST, diagnostic.PhaseDiagnostics) {
	tokens = lexerTokens
	root := AST{
		label: "program",
	}
	length = len(tokens)
	curr = tokens[0]
	if length > 1 {
		next = tokens[1]
	}
	// fmt.Println("First: ", tokens[0], ptr)
	for ptr < length {
		// fmt.Println(curr, next, isVariableDeclaration())
		if curr.Value == "fn" || curr.Value == "struct" || curr.Value == "interface" || isVariableDeclaration() { // TODO: add variable
			root.AddChildren(parseDeclaration())
		} else {
			//fmt.Printf("Token: %v\n", token)
			//ptr++
			//complainAboutToken("Only declarations supported at top level", curr)
		}
		movePtr()
	}
	return root, messages
}

func isVariableDeclaration() bool {
	if ptr == length-1 {
		return false
	}
	if curr.Kind == lexer.KW_TYPE {
		return true
	}
	if curr.Kind == lexer.KW_MODIFIER {
		if next.Kind == lexer.KW_MODIFIER {
			if ptr == length-2 {
				return false
			} else if tokens[ptr+2].Value == "fn" {
				return false
			} else {
				return true
			}
		} else {
			if next.Value != "fn" {
				return true
			}
		}
	}
	if curr.Kind == lexer.ID && next.Kind == lexer.ID {
		return true
	}
	return false
}

/*
 *	declaration = function | struct | interface | variable ;
 */
func parseDeclaration() AST {
	switch getToken().Value {
	case "fn":
		return parseFunction()
	case "struct":
		return parseStruct()
	case "interface":
		return parseInterface()
	default:
		return parseVariable()
	}
}

/*
 * function = "fn" identifier "(" [ parameters ] ")" [ "->" type ] ( ";" | block ) ;
 */
func parseFunction() AST {
	ast := AST{
		label: "fn",
	}
	start := curr

	if ptr == length-1 || next.Kind != lexer.ID {
		complainAboutToken("Expected identifier after `fn` but did not find it", start)
		return ast
	}
	movePtr()
	ast.AddChildren(AST{token: curr}) // add name
	if next.Value != "(" {
		complainAboutToken(fmt.Sprintf("Expected '(' but got %s", next.Value), next)
	}
	movePtr()
	ast.AddChildren(parseParameters())
	movePtr()
	if curr.Value != ")" {
		complainAboutToken(fmt.Sprintf("Expected ')' but got %s", curr.Value), curr)
	}
	if next.Value == "->" {
		movePtr()
		if ptr == length-1 {
			complainAboutToken("Expected function return type but reached EOF", curr)
			return ast
		}
		if next.Kind == lexer.ID || next.Kind == lexer.KW_TYPE {
			ast.AddChildren(AST{token: next})
		} else {
			complainAboutToken(fmt.Sprintf("Expected function return type but got %s", next), next)
		}
		movePtr()
	}
	switch next.Value {
	case ";":
		return ast
	case "{":
		movePtr()
		ast.AddChildren(parseBlock())
		movePtr()
		if curr.Value != "}" {
			complainAboutToken(fmt.Sprintf("Expected '}' but got %s", curr.Value), curr)
		}
	default:
		if ptr == length-1 {
			complainAboutToken("Expected ';' or '{' but reached EOF", curr)
		} else {
			complainAboutToken(fmt.Sprintf("Expected ';' or '{' but got %s", next.Value), next)
		}
	}
	return ast
}

/*
 * parameters = parameter { "," parameter } ;
 */
func parseParameters() AST {
	return AST{}
}

/*
 * parameter = type identifier ;
 */
func parseParameter() AST {
	return AST{}
}

/*
 * block = "{" { statement } "}" ;
 */
func parseBlock() AST {
	return AST{}
}

/*
 * statement = ( ( variable | assignment | expression | control_flow ) ";" ) | branch ;
 */
func parseStatement() AST {
	token := getToken()
	if token.Kind == lexer.KW_FLOW {
		return parseControlFlow()
	}
	return AST{}
}

/*
 * branch = if_block | while | for ;
 */
func parseBranch() AST {
	return AST{}
}

/*
 * expression = logical_or | "(" logical_or ")" ;
 */
func parseExpression() AST {
	return AST{}
}

/*
 * struct = "struct" identifier [ "impl" interface_list ] struct_body ;
 */
func parseStruct() AST {
	return AST{}
}

/*
 * interface_list = identifier { "," identifier };
 */
func parseInterfaceList() AST {
	return AST{}
}

/*
 * struct_body =  "{" { variable | function | named_block } "}" ;
 */
func parseStructBody() AST {
	return AST{}
}

/*
 * named_block = identifier "{" { function | variable } "}" ;
 */
func parseNamedBlock() AST {
	return AST{}
}

/*
 * interface = "interface" identifier "{" { function } "}" ;
 */
func parseInterface() AST {
	ast := AST{
		label: "interface",
	}
	return ast
}

/*
 * variable = [ modifiers ] type identifier [ "=" expression ] ;
 */
func parseVariable() AST {
	var errMessage string
	ast := AST{label: "Variable"}
	// fmt.Println("CURR", curr)
	if curr.Kind == lexer.KW_MODIFIER {
		ast.AddChildren(parseModifiers())
		movePtr()
	}
	if (curr.Kind == lexer.KW_TYPE || curr.Kind == lexer.ID) && next.Kind == lexer.ID {
		ast.AddChildren(AST{token: curr}, AST{token: next})
		movePtr()
		switch next.Value {
		case ";":
			movePtr()
			return ast
		case "=":
			ast.AddChildren(parseExpression())
			if curr.Value != ";" {
				errMessage = "Missing semicolon"
			} else {
				movePtr()
				return ast
			}
		default:
			errMessage = "Variable declaration missing value and ';'"
		}

	} else {
		errMessage = fmt.Sprintf("Improper variable declaration: %s %s", curr, next)
	}
	if errMessage != "" {
		complainAboutToken(errMessage, curr)
	}
	return ast
}

/*
 * if_block = if { "else" if } [ "else" conditional_body ] ;
 */
func parseIfBlock() AST {
	return AST{}
}

/*
 * if = "if" "(" expression ")" conditional_body ;
 */
func parseIf() AST {
	return AST{}
}

/*
 * conditional_body = block | statement ;
 */
func parseConditionalBody() AST {
	return AST{}
}

/*
 * while = "while" "(" expression ")" block;
 */
func parseWhile() AST {
	return AST{}
}

/*
 * for = "for" "(" for_conditions ")" block ;
 */
func parseFor() AST {
	return AST{}
}

/*
 * for_conditions = ( ( variable | assignment ) ";" expression ";" expression ) | ( variable [ "," variable ] "in" range ) ;
 */
func parseForConditions() AST {
	return AST{}
}

/*
 * range = expression [ range_operator expression [ ".." expression ] ] ;
 */
func parseRange() AST {
	return AST{}
}

/*
 * assignment = member assign_operator expression ;
 */
func parseAssignment() AST {
	return AST{}
}

/*
 * logical_or = logical_and { "||" logical_and } ;
 */
func parseLogicalOr() AST {
	return AST{}
}

/*
 * logical_and = logical_not { "&&" logical_not } ;
 */
func parseLogicalAnd() AST {
	return AST{}
}

/*
 * logical_not = [ "!" ] comparison ;
 */
func parseLogicalNot() AST {
	return AST{}
}

/*
 * comparison = bitwise [ compare_operator bitwise ] ;
 */
func parseComparison() AST {
	return AST{}
}

/*
 * bitwise =  add { bitwise_operator add };
 */
func parseBitwise() AST {
	return AST{}
}

/*
 * add = mult { ( "+" | "-" ) mult } ;
 */
func parseAdd() AST {
	return AST{}
}

/*
 * mult = expo { multiplication_operator expo } ;
 */
func parseMult() AST {
	return AST{}
}

/*
 * expo = unary { "**" expo } ;
 */
func parseExpo() AST {
	return AST{}
}

/*
 * unary = left_unary | right_unary ;
 */
func parseUnary() AST {
	return AST{}
}

/*
 * left_unary = [ "^" | "-" | right_unary_operators ] typecast ;
 */
func parseLeftUnary() AST {
	return AST{}
}

/*
 * right_unary = typecast [ right_unary_operators ] ;
 */
func parseRightUnary() AST {
	return AST{}
}

/*
 * typecast = index [ "as" type ] ;
 */
func parseTypecast() AST {
	return AST{}
}

/*
 * index = term { "[" index_value "]" } ;
 */
func parseIndex() AST {
	return AST{}
}

/*
 * term = literal | member | call | expression ;
 */
func parseTerm() AST {
	return AST{}
}

/*
 * index_value =  slice | expression ;
 */
func parseIndexValue() AST {
	return AST{}
}

/*
 * slice = [ expression | array_end ] range_operator [ expression | array_end ] ;
 */
func parseSlice() AST {
	return AST{}
}

/*
 * array_end = "^" ( ( "1" ... "9" ) { "0" ... "9" } ) ;
 */
func parseArrayEnd() AST {
	return AST{}
}

/*
 * literal = bool_literal | char_literal | string_literal | number_literal | struct_literal;
 */
func parseLiteral() AST {
	ast := AST{}
	token := getToken()
	if token.Kind == lexer.LIT_CHAR || token.Kind == lexer.LIT_STRING {
		ast.token = token
	} else if token.Kind == lexer.LIT_INT || token.Kind == lexer.LIT_HEX || token.Kind == lexer.LIT_FLOAT || token.Value == "+" || token.Value == "-" {
		return parseNumLiteral()
	} else {
		return parseStructLiteral()
	}
	return ast
}

/*
 * number_literal = [ "+" | "-" ] ( hex | float | int ) ;
 */
func parseNumLiteral() AST {
	sign := ""
	if curr.Value == "+" || curr.Value == "-" {
		sign = curr.Value
		next.IsSigned = true
	}
	switch next.Kind {
	case lexer.LIT_INT, lexer.LIT_HEX:
		if sign == "+" {
			next.IntVal = +next.IntVal
		} else {
			next.IntVal = -next.IntVal
		}
	case lexer.LIT_FLOAT:
		if sign == "+" {
			next.FloatVal = +next.FloatVal
		} else {
			next.FloatVal = -next.FloatVal
		}
	}
	return AST{token: next}
}

/*
 * struct_literal = identifier "{" [ properties ] "}";
 */
func parseStructLiteral() AST {
	return AST{}
}

/*
 * properties =  property { ","  property } [ "," ] ;
 */
func parseProperties() AST {
	return AST{}
}

/*
 * property = identifier ":" expression ;
 */
func parseProperty() AST {
	return AST{}
}

/*
 * modifiers = "private" [ "mut" ] | "mut" [ "private" ] ;
 */
func parseModifiers() AST {
	ast := AST{
		label: "modifiers",
	}
	fmt.Println("HEEEERRREEE: ", curr, next)
	ast.AddChildren(AST{token: curr})
	if next.Kind == lexer.KW_MODIFIER {
		if curr.Value == next.Value {
			message := fmt.Sprintf("Invalid variable modifiers: %s %s", curr.Value, next.Value)
			complainAboutToken(message, curr)
		} else {
			ast.AddChildren(AST{token: next})
		}
		movePtr()
	}
	return ast
}

/*
 * member = identifier { "." identifier } ;
 */
func parseMember() AST {
	return AST{}
}

/*
 * call = member "(" [  expression { "," expression } ]")" ;
 */
func parseCall() AST {
	return AST{}
}

/*
 * control_flow = "return" [ expression ] | "continue" | "break" ;
 */
func parseControlFlow() AST {
	ast := AST{
		label: "control-flow",
	}
	token := getToken()
	if token.Value == "continue" || token.Value == "break" {
		ast.AddChildren(AST{token: token})
	} else {
		ast.AddChildren(AST{token: token})
		ptr++
		token := peekToken()
		if token.Value == ";" { // return;
			return ast
		} else { // return i + 1;
			ast.AddChildren(parseExpression())
		}
	}
	return ast
}
