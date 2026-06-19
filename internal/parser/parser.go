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

/*
return current token without moving
*/
func peek() lexer.Token {
	return tokens[ptr]
}

/*
return current token and move
*/
func consume() lexer.Token {
	token := curr
	if !checkKind(lexer.EOF) {
		ptr++
		curr = tokens[ptr]
	}
	return token
}

/*
Match the current token's kind (or report) error and advance
*/
func expectKind(kind lexer.TokenType) lexer.Token {
	if !checkKind(kind) {
		complainAboutToken("Unexpected", curr) // TODO
	}
	return consume()
}

/*
Match the current token's value (or report error) and advance
*/
func expectValue(value string) lexer.Token {
	if !checkValue(value) {
		complainAboutToken("Bad", curr) // TODO
	}
	return consume()
}

/*
Match the current token's kind without reporting error or advancing
*/
func checkKind(kind lexer.TokenType) bool {
	return curr.Kind == kind
}

/*
Match the current token's value without reporting error or advancing
*/
func checkValue(value string) bool {
	return curr.Value == value
}

func getToken() lexer.Token {
	return tokens[ptr]
}

func checkValueAhead(value string, n int) bool {
	if ptr+n >= length {
		return false
	}
	return tokens[ptr+n].Value == value
}

func checkKindAhead(kind lexer.TokenType, n int) bool {
	if ptr+n >= length {
		return false
	}
	return tokens[ptr+n].Kind == kind
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
	//for curr = tokens[ptr]; ptr < length; ptr++ {
	for !checkKind(lexer.EOF) {
		//fmt.Println(curr, isVariableDeclaration())
		if checkValue("fn") || checkValue("struct") || checkValue("interface") || isVariableDeclaration() { // TODO: add variable
			root.AddChildren(parseDeclaration())
		} else {
			//fmt.Printf("Token: %v\n", token)
			//ptr++
			//complainAboutToken("Only declarations supported at top level", curr)
		}
		//consume()
	}
	return root, messages
}

func isVariableDeclaration() bool {
	return !checkKind(lexer.EOF) && ((checkKind(lexer.KW_TYPE)) ||
		(checkKind(lexer.KW_MODIFIER) && !checkValueAhead("fn", 1)) ||
		(checkKind(lexer.KW_MODIFIER) && checkKindAhead(lexer.KW_MODIFIER, 1) && !checkValueAhead("fn", 2)) ||
		(checkKind(lexer.ID) && checkKindAhead(lexer.ID, 1)))
}

/*
 *	declaration = function | struct | interface | variable ;
 */
func parseDeclaration() AST {
	if checkValue("fn") {
		return parseFunction()
	} else if checkValue("struct") {
		return parseStruct()
	} else if checkValue("interface") {
		return parseInterface()
	} else {
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

	if ptr >= length-1 || next.Kind != lexer.ID {
		complainAboutToken("Expected identifier after `fn` but did not find it", start)
		return ast
	}
	movePtr()
	ast.AddChildToken(curr) // add name
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
		if ptr >= length-1 {
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
		if ptr >= length-1 {
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
	ast := AST{label: "params"}
	for curr = tokens[ptr]; ptr < length-1; ptr++ {
		ast.AddChildren(parseParameter())
		ptr++
		if ptr >= length-1 {
			break
		}
		if curr.Value == "," {
			continue
		}
	}
	return ast
}

/*
 * parameter = type identifier ;
 */
func parseParameter() AST {
	ast := AST{label: "param"}
	if curr.Kind != lexer.KW_TYPE && curr.Kind != lexer.ID {
		complainAboutToken(fmt.Sprintf("Expected type but got %s", curr.Value), curr)
		return ast
	}
	ast.AddChildToken(curr)
	ptr++
	if ptr >= length-1 {
		complainAboutToken("Expected identifier but got EOF", curr)
		return ast
	}
	curr = tokens[ptr]
	if curr.Kind != lexer.ID {
		complainAboutToken(fmt.Sprintf("Expected identifier but got %s", curr.Value), curr)
		return ast
	}
	ast.AddChildToken(curr)
	return ast
}

/*
 * block = "{" { statement } "}" ;
 */
func parseBlock() AST {
	ast := AST{}
	return ast
}

/*
 * statement = ( ( variable | assignment | expression | control_flow ) ";" ) | branch ;
 */
func parseStatement() AST {
	token := getToken()
	if token.Kind == lexer.KW_FLOW {
		return parseControlFlow()
	}
	ast := AST{}
	return ast
}

/*
 * branch = if_block | while | for ;
 */
func parseBranch() AST {
	ast := AST{}
	return ast
}

/*
 * expression = logical_or | "(" logical_or ")" ;
 */
func parseExpression() AST {
	ast := AST{}
	in_parens := false
	if curr.Value == "(" {
		in_parens = true
		ptr++
		if ptr >= length-1 {
			complainAboutToken("Found opening parenthesis but no expression and no closing parenthesis", curr)
		}
		curr = tokens[ptr]
	}
	ast.AddChildren(parseLogicalOr())
	if in_parens {
		ptr++
		if ptr >= length-1 {
			complainAboutToken("Found opening parenthesis but no closing parenthesis", curr)
		}
		curr = tokens[ptr]
		if curr.Value != ")" {
			complainAboutToken(fmt.Sprintf("Expected closing parenthesis but found %s", curr.Value), curr)
		}
	}
	return ast
}

/*
 * struct = "struct" identifier [ "impl" interface_list ] struct_body ;
 */
func parseStruct() AST {
	ast := AST{}

	return ast
}

/*
 * interface_list = identifier { "," identifier };
 */
func parseInterfaceList() AST {
	ast := AST{}

	return ast
}

/*
 * struct_body =  "{" { variable | function | named_block } "}" ;
 */
func parseStructBody() AST {
	ast := AST{}

	return ast
}

/*
 * named_block = identifier "{" { function | variable } "}" ;
 */
func parseNamedBlock() AST {
	ast := AST{}

	return ast
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
	//var errMessage string
	ast := AST{label: "Variable"}
	fmt.Println("CURR", curr)
	if checkKind(lexer.KW_MODIFIER) {
		fmt.Println("HERE")
		ast.AddChildren(parseModifiers())
	}
	if (checkKind(lexer.KW_TYPE) || checkKind(lexer.ID)) && checkKindAhead(lexer.ID, 1) {
		varType := consume()
		id := consume()
		ast.AddChildren(AST{token: varType}, AST{token: id})
		//fmt.Println("Curr: ", curr, "varType:", varType, "id: ", id)
		if checkValue(";") {
			consume()
			return ast
		}
		if checkValue("=") {
			consume()
			ast.AddChildren(parseExpression())
			expectValue(";")
		} else {
			complainAboutToken("Expected ';' or '='", curr)
			consume()
		}
	}
	return ast
}

/*
 * if_block = if { "else" if } [ "else" conditional_body ] ;
 */
func parseIfBlock() AST {
	ast := AST{}

	return ast
}

/*
 * if = "if" "(" expression ")" conditional_body ;
 */
func parseIf() AST {
	ast := AST{}

	return ast
}

/*
 * conditional_body = block | statement ;
 */
func parseConditionalBody() AST {
	ast := AST{}

	return ast
}

/*
 * while = "while" "(" expression ")" block;
 */
func parseWhile() AST {
	ast := AST{}

	return ast
}

/*
 * for = "for" "(" for_conditions ")" block ;
 */
func parseFor() AST {
	ast := AST{}

	return ast
}

/*
 * for_conditions = ( ( variable | assignment ) ";" expression ";" expression ) | ( variable [ "," variable ] "in" range ) ;
 */
func parseForConditions() AST {
	ast := AST{}

	return ast
}

/*
 * range = expression [ range_operator expression [ ".." expression ] ] ;
 */
func parseRange() AST {
	ast := AST{}

	return ast
}

/*
 * assignment = member assign_operator expression ;
 */
func parseAssignment() AST {
	ast := AST{}

	return ast
}

/*
 * logical_or = logical_and { "||" logical_and } ;
 */
func parseLogicalOr() AST {
	ast := parseLogicalAnd()
	ptr++
	if ptr >= length-1 {
		return ast
	}
	for curr = tokens[ptr]; curr.Value == "||" && ptr < length-1; ptr++ {
		ast.AddChildToken(curr)
		ptr++
		if ptr >= length-1 {
			return ast
		}
		curr = tokens[ptr]
		ast.AddChildren(parseLogicalAnd())
	}
	if curr.Value == "||" {
		complainAboutToken(fmt.Sprintf("Expected operand but got %s", curr.Value), curr)
	}
	return ast
}

/*
 * logical_and = logical_not { "&&" logical_not } ;
 */
func parseLogicalAnd() AST {
	ast := parseLogicalNot()
	ptr++
	if ptr >= length-1 {
		return ast
	}
	for curr = tokens[ptr]; curr.Value == "&&" && ptr < length-1; ptr++ {
		ast.AddChildToken(curr)
		ptr++
		if ptr >= length-1 {
			return ast
		}
		curr = tokens[ptr]
		ast.AddChildren(parseLogicalNot())
	}
	if curr.Value == "&&" {
		complainAboutToken(fmt.Sprintf("Expected operand but got %s", curr.Value), curr)
	}
	return ast
}

/*
 * logical_not = [ "!" ] comparison ;
 */
func parseLogicalNot() AST {
	ast := AST{}
	if curr.Value == "!" {
		ast.AddChildToken(curr)
		ptr++
		if ptr >= length-1 {
			complainAboutToken(fmt.Sprintf("Expected operand but got %s", curr.Value), curr)
			return ast
		}
		curr = tokens[ptr]
	}
	ast.AddChildren(parseComparison())
	return ast
}

/*
 * comparison = bitwise [ compare_operator bitwise ] ;
 */
func parseComparison() AST {
	ast := parseBitwise()
	ptr++
	if ptr >= length-1 {
		return ast
	}
	curr = tokens[ptr]
	if curr.Kind != lexer.OPERATOR_COMPARE {
		return ast
	}
	ast.AddChildToken(curr)
	ptr++
	if ptr >= length-1 {
		complainAboutToken(fmt.Sprintf("Expected operand but got %s", curr.Value), curr)
		return ast
	}
	curr = tokens[ptr]
	ast.AddChildren(parseBitwise())
	return ast
}

/*
 * bitwise =  add { bitwise_operator add };
 */
func parseBitwise() AST {
	ast := parseAdd()
	ptr++
	if ptr >= length-1 {
		return ast
	}
	for curr = tokens[ptr]; curr.Kind == lexer.OPERATOR_BW && ptr < length-1; ptr++ {
		ast.AddChildToken(curr)
		ptr++
		curr = tokens[ptr]
		ast.AddChildren(parseAdd())
	}
	if curr.Kind == lexer.OPERATOR_BW {
		complainAboutToken(fmt.Sprintf("Expected operand but got %s", curr.Value), curr)
	}

	return ast
}

/*
 * add = mult { ( "+" | "-" ) mult } ;
 */
func parseAdd() AST {
	ast := parseMult()
	ptr++
	if ptr >= length-1 {
		return ast
	}
	for curr = tokens[ptr]; curr.Kind == lexer.OPERATOR_ADD && ptr < length-1; ptr++ {
		ast.AddChildToken(curr)
		ptr++
		curr = tokens[ptr]
		ast.AddChildren(parseMult())
	}
	if curr.Kind == lexer.OPERATOR_ADD {
		complainAboutToken(fmt.Sprintf("Expected operand but got %s", curr.Value), curr)
	}
	return ast
}

/*
 * mult = expo { multiplication_operator expo } ;
 */
func parseMult() AST {
	ast := parseExpo()
	ptr++
	if ptr >= length-1 {
		return ast
	}
	for curr = tokens[ptr]; curr.Kind == lexer.OPERATOR_MULT && ptr < length-1; ptr++ {
		ast.AddChildToken(curr)
		ptr++
		curr = tokens[ptr]
		ast.AddChildren(parseExpo())
	}
	if curr.Kind == lexer.OPERATOR_MULT {
		complainAboutToken(fmt.Sprintf("Expected operand but got %s", curr.Value), curr)
	}
	return ast
}

/*
 * expo = unary { "**" expo } ;
 */
func parseExpo() AST {
	ast := parseUnary()
	ptr++
	if ptr >= length-1 {
		return ast
	}
	fmt.Println("END: ", length, ", PTR: ", ptr)
	for curr = tokens[ptr]; curr.Value == "**" && ptr < length-1; ptr++ {
		ast.AddChildToken(curr)
		ptr++
		curr = tokens[ptr]
		ast.AddChildren(parseExpo())
	}
	if curr.Value == "**" {
		complainAboutToken(fmt.Sprintf("Expected operand but got %s", curr.Value), curr)
	}
	return ast
}

/*
 * unary = left_unary | right_unary ;
 */
func parseUnary() AST {
	if curr.Value == "-" || curr.Kind == lexer.OPERATOR_UNARY {
		return parseLeftUnary()
	}
	return parseRightUnary()
}

/*
 * left_unary = [ "-" | right_unary_operators ] typecast ;
 */
func parseLeftUnary() AST {
	ast := AST{label: "Unary"}
	ast.AddChildToken(curr)
	ptr++
	if ptr >= length-1 {
		complainAboutToken(fmt.Sprintf("Expected operand but got %s", curr.Value), curr)
		return ast
	}
	curr = tokens[ptr]
	ast.AddChildren(parseTypecast())
	return ast
}

/*
 * right_unary = typecast [ right_unary_operators ] ;
 */
func parseRightUnary() AST {
	ast := AST{label: "Unary"}
	ast.AddChildren(parseTypecast())
	ptr++
	if ptr >= length-1 {
		return ast
	}
	curr = tokens[ptr]
	if curr.Kind != lexer.OPERATOR_UNARY {
		return ast
	}
	ast.AddChildToken(curr)
	return ast
}

/*
 * typecast = index [ "as" type ] ;
 */
func parseTypecast() AST {
	ast := AST{label: "typecast"}
	index := parseIndex()
	ptr++
	if ptr >= length-1 {
		return index
	}
	curr = tokens[ptr]
	if curr.Value != "as" {
		return index
	}
	ast.AddChildren(index)
	ptr++
	if ptr >= length-1 {
		complainAboutToken("Expected but did not find type after 'as'", curr)
		return ast
	}
	curr = tokens[ptr]
	if curr.Kind == lexer.KW_TYPE || curr.Kind == lexer.ID {
		ast.AddChildToken(curr)
	} else {
		complainAboutToken(fmt.Sprintf("Expected type but found %s", curr.Value), curr)
	}
	return ast
}

/*
 * index = term { "[" index_value "]" } ;
 */
func parseIndex() AST {
	ast := AST{label: "index"}
	term := parseTerm()
	ptr++
	if ptr >= length-1 {
		return term
	}
	for curr = tokens[ptr]; curr.Value == "[" && ptr < length-1; ptr++ {
		// TODO
	}
	return ast
}

/*
 * term = literal | member | call | expression ;
 */
func parseTerm() AST {
	ast := AST{}

	return ast
}

/*
 * index_value =  slice | expression | array_end ;
 */
func parseIndexValue() AST {
	ast := AST{}

	return ast
}

/*
 * slice = [ expression | array_end ] range_operator [ expression | array_end ] ;
 */
func parseSlice() AST {
	ast := AST{}

	return ast
}

/*
 * array_end = "^" expression ;
 */
func parseArrayEnd() AST {
	ast := AST{label: "ARR-END"}
	ast.AddChildToken(curr)
	ptr++
	if ptr >= length-1 {
		complainAboutToken("Expected expression", curr)
	}
	ast.AddChildren(parseExpression())
	return ast
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
	ast := AST{label: "struct_literal"}
	ast.AddChildToken(curr)
	ptr++
	if ptr >= length-1 {
		complainAboutToken("Expected '{' but got EOF", curr)
		return ast
	}
	curr = tokens[ptr]
	if curr.Value != "{" {
		complainAboutToken(fmt.Sprintf("Expected '{' but got %s", curr.Value), curr)
		return ast
	}
	ptr++
	if ptr >= length-1 {
		complainAboutToken("Expected '}' but got EOF", curr)
		return ast
	}
	curr = tokens[ptr]
	if curr.Value == "}" {
		return ast
	}
	ast.AddChildren(parseProperties())
	ptr++
	if ptr >= length-1 {
		complainAboutToken("Expected '}' but got EOF", curr)
		return ast
	}
	curr = tokens[ptr]
	if curr.Value != "}" {
		complainAboutToken(fmt.Sprintf("Expected '{' but got %s", curr.Value), curr)
	}
	return ast
}

/*
 * properties =  property { ","  property } [ "," ] ;
 */
func parseProperties() AST {
	ast := AST{label: "properties"}
	for curr = tokens[ptr]; ptr < length-1; ptr++ {
		ast.AddChildren(parseProperty())
		ptr++
		if ptr >= length-1 {
			break
		}
		curr = tokens[ptr]
		if curr.Value == "," {
			continue
		}
	}
	if curr.Value == "," {
		ptr++
	}
	return ast
}

/*
 * property = identifier ":" expression ;
 */
func parseProperty() AST {
	ast := AST{label: "property"}
	if curr.Kind != lexer.ID {
		complainAboutToken(fmt.Sprintf("Expected identifier but got %s", curr.Value), curr)
		return ast
	}
	ast.AddChildToken(curr)
	ptr++
	if ptr >= length-1 {
		complainAboutToken("Expected ':' but got EOF", curr)
		return ast
	}
	curr = tokens[ptr]
	if curr.Value != ":" {
		complainAboutToken(fmt.Sprintf("Expected ':' but got %s", curr.Value), curr)
		return ast
	}
	ptr++
	if ptr >= length-1 {
		complainAboutToken("Expected expression but got EOF", curr)
		return ast
	}
	ast.AddChildren(parseExpression())
	return ast
}

/*
 * modifiers = "private" [ "mut" ] | "mut" [ "private" ] ;
 */
func parseModifiers() AST {
	ast := AST{
		label: "modifiers",
	}
	modifier := consume()
	fmt.Println(modifier)
	ast.AddChildToken(modifier)
	if checkKind(lexer.KW_MODIFIER) {
		if checkValue(modifier.Value) {
			message := fmt.Sprintf("Invalid variable modifiers: %s %s", modifier.Value, curr.Value)
			complainAboutToken(message, curr)
		} else {
			ast.AddChildToken(consume())
		}
	}
	return ast
}

/*
 * member = ( identifier | string_literal ) { "." identifier } ;
 */
func parseMember() AST {
	lhs := AST{token: curr}
	if ptr >= length-1 {
		return lhs
	}
	ptr++
	curr = tokens[ptr]
	if curr.Value == "." {
		ast := AST{label: "dot"}
		ptr++
		if ptr >= length-1 {
			complainAboutToken(fmt.Sprintf("Expected identifier but got %s", curr.Value), curr)
			return ast
		}
		curr = tokens[ptr]
		if curr.Kind != lexer.ID {
			complainAboutToken(fmt.Sprintf("Expected identifier but got %s", curr.Value), curr)
			return ast
		}
		ast.AddChildren(lhs, AST{token: curr})
	}
	return lhs
}

/*
 * call = member "(" [  expression { "," expression } ]")" ;
 */
func parseCall() AST {
	ast := AST{label: "call"}
	ast.AddChildren(parseMember())
	ptr++
	if ptr >= length-1 {
		// TODO
	}
	curr = tokens[ptr]
	if curr.Value != "(" {
		complainAboutToken(fmt.Sprintf("Expected '(' but got %s", curr.Value), curr)
		return ast
	}
	ptr++
	if ptr >= length-1 {
		// TODO
	}
	curr = tokens[ptr]
	if curr.Value == ")" {
		return ast
	}
	params := AST{label: "params"}
	ptr++
	if ptr >= length-1 {
		// TODO
	}
	curr = tokens[ptr]
	for curr = tokens[ptr]; ptr < length-1; ptr++ {
		params.AddChildren(parseExpression())
		ptr++
		if ptr >= length-1 {
			complainAboutToken("Expected ')' but could not find it", curr)
			return ast
		}
		if curr.Value == ")" {
			break
		}
		if curr.Value == "," {
			ptr++
			if ptr >= length-1 {
				complainAboutToken("Expected ',' but reached EOF", curr)
				return ast
			}
			curr = tokens[ptr]
			params.AddChildren(parseExpression())
		} else {
			complainAboutToken(fmt.Sprintf("Expected ',' but got %s", curr.Value), curr)
			return ast
		}
	}
	ast.AddChildren(params)
	return ast
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
