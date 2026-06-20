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
		complainAboutToken("Unexpected") // TODO
	}
	return consume()
}

/*
Match the current token's value (or report error) and advance
*/
func expectValue(value string) lexer.Token {
	if !checkValue(value) {
		complainAboutToken("Bad") // TODO
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

func complainAboutToken(message string) {
	token := peek()
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
	// fmt.Println("First: ", tokens[0], ptr)
	//for curr = tokens[ptr]; ptr < length; ptr++ {
	for !checkKind(lexer.EOF) {
		//fmt.Println(curr, isVariableDeclaration())
		if checkValue("fn") || checkValue("struct") || checkValue("interface") || isVariableDeclaration() { // TODO: add variable
			root.AddChildren(parseDeclaration())
		} else {
			//fmt.Printf("Token: %v\n", token)
			//ptr++
			//complainAboutToken("Only declarations supported at top level")
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
	expectValue("fn")
	ast.AddChildToken(expectKind(lexer.ID))
	expectValue("(")
	if !checkValue(")") {
		ast.AddChildren(parseParameters())
	}
	expectValue(")")
	if checkValue("->") {
		consume()
		if checkKind(lexer.KW_TYPE) || checkKind(lexer.ID) {
			ast.AddChildToken(consume())
		} else {
			complainAboutToken(fmt.Sprintf("Expected function return type but got %s", peek().Value))
			return ast
		}
	}
	if checkValue(";") {
		consume()
		return ast
	}
	if checkValue("{") {
		ast.AddChildren(parseBlock())
	} else {
		// TODO
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
	if !checkKind(lexer.KW_TYPE) && !checkKind(lexer.ID) {
		complainAboutToken(fmt.Sprintf("Expected type but got %s", peek().Value))
		return ast
	}
	ast.AddChildToken(consume())
	ast.AddChildToken(expectKind(lexer.ID))
	return ast
}

/*
 * block = "{" { statement } "}" ;
 */
func parseBlock() AST {
	ast := AST{}
	expectValue("{")
	// TODO
	expectValue("}")
	return ast
}

/*
 * statement = ( ( variable | assignment | expression | control_flow ) ";" ) | branch ;
 */
func parseStatement() AST {
	if checkKind(lexer.KW_FLOW) {
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
	if checkValue("(") {
		in_parens = true
		consume()
	}
	ast.AddChildren(parseLogicalOr())
	if in_parens {
		expectValue(")")
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
			complainAboutToken("Expected ';' or '='")
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
	for checkValue("||") {
		ast.AddChildToken(consume())
		ast.AddChildren(parseLogicalAnd())
	}
	if checkValue("||") {
		complainAboutToken(fmt.Sprintf("Expected operand but got %s", peek().Value))
	}
	return ast
}

/*
 * logical_and = logical_not { "&&" logical_not } ;
 */
func parseLogicalAnd() AST {
	ast := parseLogicalNot()
	for checkValue("&&") {
		ast.AddChildToken(consume())
		ast.AddChildren(parseLogicalNot())
	}
	if curr.Value == "&&" {
		complainAboutToken(fmt.Sprintf("Expected operand but got %s", peek().Value))
	}
	return ast
}

/*
 * logical_not = [ "!" ] comparison ;
 */
func parseLogicalNot() AST {
	ast := AST{}
	if checkValue("!") {
		ast = AST{token: consume()}
	}
	ast.AddChildren(parseComparison())
	return ast
}

/*
 * comparison = bitwise [ compare_operator bitwise ] ;
 */
func parseComparison() AST {
	bw := parseBitwise()
	if !checkKind(lexer.OPERATOR_COMPARE) {
		return bw
	}
	ast := AST{token: consume()} // operator is the root of the tree
	ast.AddChildren(bw)
	ast.AddChildren(parseBitwise())
	return ast
}

/*
 * bitwise =  add { bitwise_operator add };
 */
func parseBitwise() AST {
	var operand AST
	ast := parseAdd()
	for checkKind(lexer.OPERATOR_BW) {
		operand = ast
		ast = AST{token: consume()}
		ast.AddChildren(operand)
		ast.AddChildren(parseAdd())
	}
	if curr.Kind == lexer.OPERATOR_BW {
		complainAboutToken(fmt.Sprintf("Expected operand but got %s", consume().Value))
	}

	return ast
}

/*
 * add = mult { ( "+" | "-" ) mult } ;
 */
func parseAdd() AST {
	var operand AST
	ast := parseMult()
	for checkKind(lexer.OPERATOR_ADD) {
		operand = ast
		ast = AST{token: consume()}
		ast.AddChildren(operand)
		ast.AddChildren(parseMult())
	}
	if curr.Kind == lexer.OPERATOR_ADD {
		complainAboutToken(fmt.Sprintf("Expected operand but got %s", consume().Value))
	}
	return ast
}

/*
 * mult = expo { multiplication_operator expo } ;
 */
func parseMult() AST {
	var operand AST
	ast := parseExpo()
	for checkKind(lexer.OPERATOR_MULT) {
		operand = ast
		ast = AST{token: consume()}
		ast.AddChildren(operand)
		ast.AddChildren(parseExpo())
	}
	if curr.Kind == lexer.OPERATOR_MULT {
		complainAboutToken(fmt.Sprintf("Expected operand but got %s", consume().Value))
	}
	return ast
}

/*
 * expo = unary { "**" expo } ;
 */
func parseExpo() AST {
	var operand AST
	ast := parseUnary()
	for checkValue("**") {
		operand = ast
		ast = AST{token: consume()}
		ast.AddChildren(operand)
		ast.AddChildren(parseExpo())
	}
	if curr.Value == "**" {
		complainAboutToken(fmt.Sprintf("Expected operand but got %s", curr.Value))
	}
	return ast
}

/*
 * unary = left_unary | right_unary ;
 */
func parseUnary() AST {
	if checkValue("-") || checkKind(lexer.OPERATOR_UNARY) {
		return parseLeftUnary()
	}
	return parseRightUnary()
}

/*
 * left_unary = [ "-" | right_unary_operators ] typecast ;
 */
func parseLeftUnary() AST {
	ast := AST{label: "Unary"}
	ast.AddChildToken(consume())
	ast.AddChildren(parseTypecast())
	return ast
}

/*
 * right_unary = typecast [ right_unary_operators ] ;
 */
func parseRightUnary() AST {
	ast := AST{label: "Unary"}
	ast.AddChildren(parseTypecast())
	if !checkKind(lexer.OPERATOR_UNARY) {
		return ast
	}
	ast.AddChildToken(consume())
	return ast
}

/*
 * typecast = index [ "as" type ] ;
 */
func parseTypecast() AST {
	ast := AST{label: "typecast"}
	index := parseIndex()
	if checkValue("as") {
		return index
	}
	ast.AddChildren(index)
	consume()
	if checkKind(lexer.KW_TYPE) || checkKind(lexer.ID) {
		ast.AddChildToken(consume())
	} else {
		complainAboutToken(fmt.Sprintf("Expected type but found %s", peek().Value))
	}
	return ast
}

/*
 * index = term { "[" index_value "]" } ;
 */
func parseIndex() AST {
	var operand AST
	ast := AST{label: "index"}
	ast.AddChildren(parseTerm())
	for checkValue("[") {
		operand = ast
		ast = AST{label: "index"}
		ast.AddChildren(operand, parseIndexValue())
		expectValue("]")
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
	if checkValue("^") {
		return parseArrayEnd()
	}
	// TODO
	return ast
}

/*
 * slice = [ expression | array_end ] range_operator [ expression | array_end ] ;
 */
func parseSlice() AST {
	ast := AST{label: "slice"}
	if checkKind(lexer.OPERATOR_RANGE) {
		ast.AddChildToken(consume())
		return ast
	}
	if checkValue("^") {
		ast.AddChildren(parseArrayEnd())
	} else {
		ast.AddChildren(parseExpression())
	}
	ast.AddChildToken(expectKind(lexer.OPERATOR_RANGE))
	if checkValue("]") {
		return ast
	}
	if checkValue("^") {
		ast.AddChildren(parseArrayEnd())
	} else {
		ast.AddChildren(parseExpression())
	}
	return ast
}

/*
 * array_end = "^" expression ;
 */
func parseArrayEnd() AST {
	ast := AST{label: "ARR-END"}
	ast.AddChildToken(consume())
	ast.AddChildren(parseExpression())
	return ast
}

/*
 * literal = bool_literal | char_literal | string_literal | number_literal | struct_literal;
 */
func parseLiteral() AST {
	ast := AST{}
	if checkKind(lexer.LIT_CHAR) || checkKind(lexer.LIT_STRING) {
		ast.AddChildToken(consume())
	} else if checkKind(lexer.LIT_INT) || checkKind(lexer.LIT_HEX) || checkKind(lexer.LIT_FLOAT) || checkValue("+") || checkValue("-") {
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
	if checkValue("+") || checkValue("-") {
		sign = curr.Value
		curr.IsSigned = true
		consume()
	}
	switch curr.Kind { //TODO
	case lexer.LIT_INT, lexer.LIT_HEX:
		if sign == "+" {
			curr.IntVal = +curr.IntVal
		} else {
			curr.IntVal = -curr.IntVal
		}
	case lexer.LIT_FLOAT:
		if sign == "+" {
			curr.FloatVal = +curr.FloatVal
		} else {
			curr.FloatVal = -curr.FloatVal
		}
	}
	num := consume()
	return AST{token: num}
}

/*
 * struct_literal = identifier "{" [ properties ] "}";
 */
func parseStructLiteral() AST {
	ast := AST{label: "struct_literal"}
	ast.AddChildToken(consume())
	expectValue("{")
	if !checkValue("}") {
		ast.AddChildren(parseProperties())
	}
	expectValue("}")
	return ast
}

/*
 * properties =  property { ","  property } [ "," ] ;
 */
func parseProperties() AST { //TODO
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
	id := expectKind(lexer.ID)
	ast.AddChildToken(id)
	expectValue(":")
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
	ast.AddChildToken(modifier)
	if checkKind(lexer.KW_MODIFIER) {
		if checkValue(modifier.Value) {
			message := fmt.Sprintf("Invalid variable modifiers: %s %s", modifier.Value, peek().Value)
			complainAboutToken(message)
			consume()
		} else {
			ast.AddChildToken(consume())
		}
		if checkKind(lexer.KW_MODIFIER) {
			complainAboutToken("Too many variable modifiers")
		}
	}
	return ast
}

/*
 * member = ( identifier | string_literal ) { "." identifier } ;
 */
func parseMember() AST {
	lhs := AST{token: consume()}
	if checkValue(".") {
		ast := AST{label: "dot"}
		consume()
		token := expectKind(lexer.ID)
		ast.AddChildren(lhs, AST{token: token})
	}
	return lhs
}

/*
 * call = member "(" [  expression { "," expression } ]")" ;
 */
func parseCall() AST {
	ast := AST{label: "call"}
	ast.AddChildren(parseMember())
	if curr.Value != "(" {
		complainAboutToken(fmt.Sprintf("Expected '(' but got %s", peek().Value))
		return ast
	}
	expectValue("(")
	if checkValue(")") {
		return ast
	}
	params := AST{label: "params"}
	// TODO
	for curr = tokens[ptr]; ptr < length-1; ptr++ {
		params.AddChildren(parseExpression())
		ptr++
		if ptr >= length-1 {
			complainAboutToken("Expected ')' but could not find it")
			return ast
		}
		if curr.Value == ")" {
			break
		}
		if curr.Value == "," {
			ptr++
			if ptr >= length-1 {
				complainAboutToken("Expected ',' but reached EOF")
				return ast
			}
			curr = tokens[ptr]
			params.AddChildren(parseExpression())
		} else {
			complainAboutToken(fmt.Sprintf("Expected ',' but got %s", peek().Value))
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
	if checkValue("continue") || checkValue("break") {
		ast.AddChildToken(consume())
	} else {
		ast.AddChildToken(consume())
		if checkValue(";") { // return;
			return ast
		} else { // return i + 1;
			ast.AddChildren(parseExpression())
		}
	}
	return ast
}
