package parser

import (
	"fmt"
	"strings"

	"github.com/EladB1/The/internal/diagnostic"
	"github.com/EladB1/The/internal/lexer"
)

var (
	errLevel diagnostic.Severity = diagnostic.SyntaxError
	state    *parserState        = &parserState{}
)

/*
return current token without moving
*/
func peek() lexer.Token {
	return state.tokens[state.ptr]
}

func peekAhead(n int) lexer.Token {
	if state.ptr+n >= state.length {
		return state.tokens[state.length-1]
	}
	return state.tokens[state.ptr+n]
}

/*
return current token and move
*/
func consume() lexer.Token {
	token := peek()
	if !checkKind(lexer.EOF) {
		state.ptr++
	}
	return token
}

/*
Match the current token's kind (or report) error and advance
*/
func expectKind(kind lexer.TokenType) lexer.Token {
	if checkKind(kind) {
		return consume()
	}
	if checkKind(lexer.EOF) {
		err := fmt.Sprintf("Expected %s but got EOF", kind)
		state.addError(err)
		panic(err)
	}
	token := peek()
	/*if errorCtx == "fn" {
		errorRecoveryFunctionDefintion()
	}*/
	state.addError(fmt.Sprintf("Expected '%s' but got '%s'", kind, token.Kind))
	return lexer.Token{
		Kind:    kind,
		Value:   "",
		Missing: true,
		Line:    token.Line,
		Column:  token.Column,
	}
}

/*
Match the current token's value (or report error) and advance
*/
func expectValue(value string) lexer.Token {
	if checkValue(value) {
		return consume()
	}
	if checkKind(lexer.EOF) {
		err := fmt.Sprintf("Expected '%s' but got EOF", value)
		state.addError(err)
		panic(err)
	}
	token := peek()
	/*if errorCtx == "fn" {
		errorRecoveryFunctionDefintion()
	}*/
	state.addError(fmt.Sprintf("Expected '%s' but got '%s'", value, token.GetValueString()))
	return lexer.Token{
		Kind:    lexer.Virtual,
		Value:   value,
		Missing: true,
		Line:    token.Line,
		Column:  token.Column,
	}
}

/*
Check if the current token is a valid type (or report error) and advance
*/
func expectType() lexer.Token {
	if checkKind(lexer.ID) || checkKind(lexer.KW_TYPE) {
		return consume()
	}
	if checkKind(lexer.EOF) {
		err := "Expected type but got EOF"
		state.addError(err)
		panic(err)
	}
	token := peek()
	state.addError(fmt.Sprintf("Expected type but got '%s'", token.GetValueString()))
	return lexer.Token{
		Kind:    lexer.Virtual,
		Value:   "none",
		Missing: true,
		Line:    token.Line,
		Column:  token.Column,
	}
}

/*
Match the current token's kind without reporting error or advancing
*/
func checkKind(kind lexer.TokenType) bool {
	return peek().Kind == kind
}

/*
Match the current token's value without reporting error or advancing
*/
func checkValue(value string) bool {
	return peek().Value == value
}

func checkValueAhead(value string, n int) bool {
	if state.ptr+n >= state.length {
		return false
	}
	return state.tokens[state.ptr+n].Value == value
}

func checkKindAhead(kind lexer.TokenType, n int) bool {
	if state.ptr+n >= state.length {
		return false
	}
	return state.tokens[state.ptr+n].Kind == kind
}

func checkNonVariableDeclaration() bool {
	return checkValue("fn") || checkValue("struct") || checkValue("interface")
}

func checkVariableDeclaration() bool {
	return !checkKind(lexer.EOF) && (((checkKind(lexer.KW_TYPE) || checkKind(lexer.ID)) && checkKindAhead(lexer.ID, 1)) ||
		(checkKind(lexer.KW_MODIFIER) && (!checkValueAhead("fn", 1) && !checkValueAhead("{", 1))) ||
		(checkKind(lexer.KW_MODIFIER) && checkKindAhead(lexer.KW_MODIFIER, 1) && !checkValueAhead("fn", 2)))
}

func checkExpressionStart() bool {
	return isLiteral() || checkKind(lexer.OPERATOR_UNARY) || checkKind(lexer.ID) || checkValue("(")
}

/*
 *	program = { declaration } ;
 */
func Parse(lexerTokens []lexer.Token) (AST, diagnostic.PhaseDiagnostics) {
	state = initState(lexerTokens)
	root := AST{
		label: "program",
	}
	for !checkKind(lexer.EOF) {
		func() {
			defer func() {
				if rec := recover(); rec != nil {
					return
				}
			}()
			if checkNonVariableDeclaration() || checkVariableDeclaration() {
				root.AddChildren(parseDeclaration())
			} else {
				state.addError(fmt.Sprintf("Expected declaration but found '%s'", peek().GetValueString()))
				errorRecoveryTopLevel()
				//consume()
			}
		}()
	}
	defer func() {

	}()
	return root, state.messages
}

/*
 *	declaration = function | struct | interface | ( variable ";" ) ;
 */
func parseDeclaration() AST {
	if checkValue("fn") {
		return parseFunction()
	} else if checkValue("struct") {
		return parseStruct()
	} else if checkValue("interface") {
		return parseInterface()
	} else if checkVariableDeclaration() {
		variable := parseVariable()
		expectValue(";")
		return variable
	} else {
		return AST{}
	}
}

/*
 * function = "fn" identifier "(" [ parameters ] ")" [ "->" type ] ( ";" | block ) ;
 */
func parseFunction() AST {
	//fmt.Println("In function with:", peek())
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
		ast.AddChildToken(expectType())
	}
	if checkValue("{") {
		ast.AddChildren(parseBlock("fn-body"))
	} else {
		expectValue(";")
	}
	return ast
}

/*
 * parameters = parameter { "," parameter } ;
 */
func parseParameters() AST {
	//fmt.Println("In parameters with:", peek())
	ast := AST{label: "params"}
	ast.AddChildren(parseParameter())
	for checkValue(",") {
		consume()
		ast.AddChildren(parseParameter())
	}
	return ast
}

/*
 * parameter = type identifier ;
 */
func parseParameter() AST {
	//fmt.Println("In parameter with:", peek())
	ast := AST{label: "param"}
	ast.AddChildToken(expectType())
	ast.AddChildToken(expectKind(lexer.ID))
	return ast
}

/*
 * block = "{" { statement | branch } "}" ;
 */
func parseBlock(label string) AST {
	//fmt.Println("In block with:", peek())
	ast := AST{label: label}
	expectValue("{")
	for !checkValue("}") && !checkKind(lexer.EOF) {
		if checkKind(lexer.KW_BRANCH) {
			ast.AddChildren(parseBranch())
		} else if checkNonVariableDeclaration() {
			state.addError(fmt.Sprintf("Declaration %s not valid in block", peek().GetValueString()))
			synchronize()
		} else {
			ast.AddChildren(parseStatement())
		}
	}
	expectValue("}")
	return ast
}

/*
 * statement = ( variable | assignment | expression | control_flow ) ";" ;
 */
func parseStatement() AST {
	//fmt.Println("In statement with:", peek())
	var ast AST
	isAssignment := false
	if (checkKind(lexer.ID) || checkKind(lexer.LIT_STRING)) && checkValueAhead(".", 1) { // let semantic analyzer complain about `"hello".length = 5`
		for i := state.ptr + 2; i < state.length-2; i += 2 {
			curr := state.tokens[i]
			next := state.tokens[i+1]
			if curr.Kind == lexer.ID && next.Kind == lexer.OPERATOR_ASSIGN {
				isAssignment = true
				break
			} else if curr.Kind == lexer.ID && next.Value != "." {
				break
			}
		}
	}
	if checkVariableDeclaration() {
		ast = parseVariable()
	} else if isAssignment || (checkKind(lexer.ID) && checkKindAhead(lexer.OPERATOR_ASSIGN, 1)) {
		ast = parseAssignment()
	} else if checkKind(lexer.KW_FLOW) {
		ast = parseControlFlow()
	} else if checkExpressionStart() {
		ast = parseExpression()
	} else {
		state.addError(fmt.Sprintf("Expected statement but got %s", peek().GetValueString()))
		ast = AST{token: lexer.Token{
			Kind:    lexer.Virtual,
			Value:   "statement",
			Missing: true,
			Line:    peek().Line,
			Column:  peek().Column,
		}}
		return ast
	}
	expectValue(";")
	return ast
}

/*
 * branch = if_block | while | for ;
 */
func parseBranch() AST {
	if checkValue("while") {
		return parseWhile()
	} else if checkValue("for") {
		return parseFor()
	} else {
		return parseIfBlock()
	}
}

/*
 * struct = "struct" identifier [ "impl" interface_list ] struct_body ;
 */
func parseStruct() AST {
	//fmt.Println("In struct with:", peek())
	ast := AST{token: expectValue("struct")}
	ast.AddChildToken(expectKind(lexer.ID))
	if checkValue("impl") {
		consume()
		ast.AddChildren(parseInterfaceList())
	}
	ast.AddChildren(parseStructBody())
	return ast
}

/*
 * interface_list = identifier { "," identifier };
 */
func parseInterfaceList() AST {
	//fmt.Println("In interface list with:", peek())
	ast := AST{label: "interface_list"}
	ast.AddChildToken(expectKind(lexer.ID))
	for checkValue(",") {
		consume()
		ast.AddChildToken(expectKind(lexer.ID))
	}
	return ast
}

/*
 * struct_body =  "{" { ( variable ";" ) | function | named_block } "}" ;
 */
func parseStructBody() AST {
	//fmt.Println("In struct body with:", peek())
	ast := AST{label: "struct-body"}
	expectValue("{")
	for !checkValue("}") && !checkKind(lexer.EOF) {
		if checkVariableDeclaration() {
			ast.AddChildren(parseVariable())
			expectValue(";")
		} else if (checkValue("private") || checkKind(lexer.ID)) && checkValueAhead("{", 1) {
			ast.AddChildren(parseNamedBlock())
		} else if checkValue("fn") {
			ast.AddChildren(parseFunction())
		} else {
			state.addError(fmt.Sprintf("Only variables, functions, and named blocks supported in struct definition, found %s", peek().GetValueString()))
			synchronize()
			//consume()
		}
	}
	expectValue("}")
	return ast
}

/*
 * named_block = identifier "{" { function | ( variable ";" ) } "}" ;
 */
func parseNamedBlock() AST {
	//fmt.Println("In named block with:", peek())
	ast := AST{label: "named-block"}
	if checkValue("private") {
		ast.AddChildToken(consume())
	} else {
		ast.AddChildToken(expectKind(lexer.ID))
	}
	body := AST{label: "named-block-body"}
	expectValue("{")
	for !checkValue("}") && !checkKind(lexer.EOF) {
		if checkVariableDeclaration() {
			body.AddChildren(parseVariable())
			expectValue(";")
		} else if checkValue("fn") {
			body.AddChildren(parseFunction())
		} else {
			state.addError(fmt.Sprintf("Only functions and variable definitions supported in named blocks, found %s", peek().GetValueString()))
			synchronize()
		}
	}
	expectValue("}")
	ast.AddChildren(body)
	return ast
}

/*
 * interface = "interface" identifier "{" { function } "}" ;
 */
func parseInterface() AST {
	//fmt.Println("In interface: ", peek())
	ast := AST{label: "interface"}
	consume() // interface keyword
	ast.AddChildToken(expectKind(lexer.ID))
	expectValue("{")
	for !checkValue("}") {
		if checkValue("fn") {
			ast.AddChildren(parseFunction())
		} else {
			state.addError(fmt.Sprintf("Only function definitions supported within interface body. Found %s", consume().GetValueString()))
			synchronize()
		}
	}
	expectValue("}")
	return ast
}

/*
 * variable = [ modifiers ] type identifier [ "=" expression ] ;
 */
func parseVariable() AST {
	//fmt.Println("In variable with:", peek())
	ast := AST{label: "Variable"}
	if checkKind(lexer.KW_MODIFIER) {
		ast.AddChildren(parseModifiers())
	}
	varType := expectType()
	if checkKind(lexer.KW_MODIFIER) {
		consume()
		varType = expectType()
	}
	ast.AddChildren(AST{token: varType}, AST{token: expectKind(lexer.ID)})
	if checkValue("=") {
		consume()
		ast.AddChildren(parseExpression())
	}
	return ast
}

/*
 * if_block = if { "else" if } [ "else" conditional_body ] ;
 */
func parseIfBlock() AST {
	//fmt.Println("In ifBlock with:", peek())
	ast := AST{label: "if-block"}
	ast.AddChildren(parseIf(false))
	for checkValue("else") && checkValueAhead("if", 1) {
		ast.AddChildren(parseIf(true))
	}
	if checkValue("else") {
		node := AST{label: "else"}
		consume()
		node.AddChildren(parseConditionalBody())
		ast.AddChildren(node)
	}
	return ast
}

/*
 * if = "if" "(" expression ")" conditional_body ;
 */
func parseIf(elseIf bool) AST {
	//fmt.Println("In if with:", peek())
	var ast AST
	if elseIf {
		ast = AST{label: "else if"}
		expectValue("else")
	} else {
		ast = AST{label: "if"}
	}
	expectValue("if")
	expectValue("(")
	ast.AddChildren(parseExpression())
	expectValue(")")
	ast.AddChildren(parseConditionalBody())
	return ast
}

/*
 * conditional_body = block | statement ;
 */
func parseConditionalBody() AST {
	//fmt.Println("In conditional body with:", peek())
	ast := AST{label: "conditional-body"}
	if checkValue("{") {
		return parseBlock("conditional_body")
	} else {
		ast.AddChildren(parseStatement())
	}
	return ast
}

/*
 * while = "while" "(" expression ")" block;
 */
func parseWhile() AST {
	// fmt.Println("In while with: ", peek())
	ast := AST{label: "while"}
	consume()
	expectValue("(")
	//fmt.Println(checkExpressionStart(), peek())
	if !checkExpressionStart() {
		state.addError(fmt.Sprintf("Expected expression but got '%s'", peek().GetValueString()))
		synchronize()
	} else {
		ast.AddChildren(parseExpression())
		expectValue(")")
	}
	ast.AddChildren(parseBlock("loop-body"))
	return ast
}

/*
 * for = "for" "(" for_conditions ")" block ;
 */
func parseFor() AST {
	//fmt.Println("In for with:", peek())
	ast := AST{label: "for"}
	consume()
	expectValue("(")
	ast.AddChildren(parseForConditions())
	expectValue(")")
	if !checkValue("{") {
		// TODO
	}
	ast.AddChildren(parseBlock("loop-body"))
	return ast
}

/*
 * for_conditions = ( ( variable | assignment ) ";" expression ";" expression ) | ( variable [ "," variable ] "in" range ) ;
 */
func parseForConditions() AST {
	//fmt.Println("In for condition with:", peek())
	ast := AST{label: "loop-condition"}
	if checkVariableDeclaration() {
		ast.AddChildren(parseVariable())
		if checkValue(";") {
			expectValue(";")
			ast.AddChildren(parseExpression())
			expectValue(";")
			ast.AddChildren(parseExpression())
		} else {
			if checkValue(",") {
				consume()
				ast.AddChildren(parseVariable())
			}
			ast.AddChildToken(expectValue("in"))
			ast.AddChildren(parseRange())
		}
	} else if checkKind(lexer.ID) {
		ast.AddChildren(parseAssignment())
		ast.AddChildToken(expectValue(";"))
		ast.AddChildren(parseExpression())
		ast.AddChildToken(expectValue(";"))
		ast.AddChildren(parseExpression())
	} else {
		loopForms := strings.Join([]string{
			"for ([ modifier ] type identifier = expression; condition; increment) {}",
			"for (identifier = expression; condition; increment) {}",
			"for (type identifier[ , type identifier ] in range-expression) {}",
		}, "\n\t")
		state.addError("Invalid for loop syntax")
		state.messages = state.messages.ProvideInfo(fmt.Sprintf("Valid loop forms:\n\t%s", loopForms))
		synchronizeInParens()
	}
	return ast
}

/*
 * range = expression [ range_operator expression [ ".." expression ] ] ;
 */
func parseRange() AST {
	//fmt.Println("In range with:", peek())
	ast := AST{label: "range"}
	expr := parseExpression()
	if !checkKind(lexer.OPERATOR_RANGE) {
		return expr
	}
	ast.AddChildren(expr, AST{token: consume()}, parseExpression())
	if checkValue("..") {
		ast.AddChildren(AST{token: consume()}, parseExpression())
	}
	return ast
}

/*
 * assignment = member assign_operator expression ;
 */
func parseAssignment() AST {
	//fmt.Println("In assignment: ", peek())
	member := parseMember()
	ast := AST{token: expectKind(lexer.OPERATOR_ASSIGN)}
	ast.AddChildren(member, parseExpression())
	return ast
}

/*
 * expression = logical_or ;
 */
func parseExpression() AST {
	//fmt.Println("In expression with:", peek())
	if !checkExpressionStart() {
		state.addError(fmt.Sprintf("Expected expression but got '%s'", peek().GetValueString()))
		synchronize()
	}
	return parseLogicalOr()
}

/*
 * logical_or = logical_and { "||" logical_and } ;
 */
func parseLogicalOr() AST {
	//fmt.Println("In logical or with:", peek())
	var operand AST
	ast := parseLogicalAnd()
	for checkValue("||") {
		operand = ast
		ast = AST{token: consume()}
		ast.AddChildren(operand, parseLogicalAnd())
	}
	if checkValue("||") {
		state.addError(fmt.Sprintf("Expected operand but got %s", peek().GetValueString()))
	}
	return ast
}

/*
 * logical_and = logical_not { "&&" logical_not } ;
 */
func parseLogicalAnd() AST {
	//fmt.Println("In logical and with:", peek())
	var operand AST
	ast := parseLogicalNot()
	for checkValue("&&") {
		operand = ast
		ast = AST{token: consume()}
		ast.AddChildren(operand, parseLogicalNot())
	}
	if checkValue("&&") {
		state.addError(fmt.Sprintf("Expected operand but got %s", peek().GetValueString()))
	}
	return ast
}

/*
 * logical_not = [ "!" ] comparison ;
 */
func parseLogicalNot() AST {
	//fmt.Println("In logical not with:", peek())
	ast := AST{}
	hasNot := false
	if checkValue("!") {
		ast = AST{token: consume()}
		hasNot = true
	}
	compare := parseComparison()
	if hasNot {
		ast.AddChildren(compare)
		return ast
	}
	return compare
}

/*
 * comparison = bitshift [ compare_operator bitshift ] ;
 */
func parseComparison() AST {
	//fmt.Println("In comparison with:", peek())
	bw := parseBitshift()
	if !checkKind(lexer.OPERATOR_COMPARE) {
		return bw
	}
	ast := AST{token: consume()} // operator is the root of the tree
	ast.AddChildren(bw)
	ast.AddChildren(parseBitshift())
	return ast
}

/*
 *	bitshift = bitwise { ( "<<" | ">>" ) bitwise } ;
 */
func parseBitshift() AST {
	// fmt.Println("In bitshift with:", peek())
	var operand AST
	ast := parseBitwise()
	for checkKind(lexer.OPERATOR_BS) {
		operand = ast
		ast = AST{token: consume()}
		ast.AddChildren(operand, parseBitwise())
	}
	if checkKind(lexer.OPERATOR_BS) {
		state.addError(fmt.Sprintf("Expected operand but got %s", consume().GetValueString()))
	}
	return ast
}

/*
 * bitwise =  add { bitwise_operator add };
 */
func parseBitwise() AST {
	//fmt.Println("In bitwise with:", peek())
	var operand AST
	ast := parseAdd()
	for checkKind(lexer.OPERATOR_BW) {
		operand = ast
		ast = AST{token: consume()}
		ast.AddChildren(operand, parseAdd())
	}
	if checkKind(lexer.OPERATOR_BW) {
		state.addError(fmt.Sprintf("Expected operand but got %s", consume().GetValueString()))
	}

	return ast
}

/*
 * add = mult { ( "+" | "-" ) mult } ;
 */
func parseAdd() AST {
	//fmt.Println("In add with:", peek())
	var operand AST
	ast := parseMult()
	for checkKind(lexer.OPERATOR_ADD) {
		operand = ast
		ast = AST{token: consume()}
		ast.AddChildren(operand)
		ast.AddChildren(parseMult())
	}
	if checkKind(lexer.OPERATOR_ADD) {
		state.addError(fmt.Sprintf("Expected operand but got %s", consume().GetValueString()))
	}
	return ast
}

/*
 * mult = expo { multiplication_operator expo } ;
 */
func parseMult() AST {
	//fmt.Println("In mult with:", peek())
	var operand AST
	ast := parseExpo()
	for checkKind(lexer.OPERATOR_MULT) {
		operand = ast
		ast = AST{token: consume()}
		ast.AddChildren(operand)
		ast.AddChildren(parseExpo())
	}
	if checkKind(lexer.OPERATOR_MULT) {
		state.addError(fmt.Sprintf("Expected operand but got %s", consume().GetValueString()))
	}
	return ast
}

/*
 * expo = unary { "**" expo } ;
 */
func parseExpo() AST {
	//fmt.Println("In expo with:", peek())
	var operand AST
	ast := parseUnary()
	for checkValue("**") {
		operand = ast
		ast = AST{token: consume()}
		ast.AddChildren(operand)
		ast.AddChildren(parseExpo())
	}
	if checkValue("**") {
		state.addError(fmt.Sprintf("Expected operand but got %s", consume().GetValueString()))
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
	//fmt.Println("In left unary with:", peek())
	ast := AST{label: "Unary"}
	ast.AddChildToken(consume())
	ast.AddChildren(parseTypecast())
	return ast
}

/*
 * right_unary = typecast [ right_unary_operators ] ;
 */
func parseRightUnary() AST {
	//fmt.Println("In right unary with:", peek())
	ast := AST{label: "Unary"}
	typecast := parseTypecast()
	if !checkKind(lexer.OPERATOR_UNARY) {
		return typecast
	}
	ast.AddChildren(typecast)
	ast.AddChildToken(consume())
	return ast
}

/*
 * typecast = index [ "as" type ] ;
 */
func parseTypecast() AST {
	//fmt.Println("In typecast with:", peek())
	ast := AST{label: "typecast"}
	index := parseIndex()
	if !checkValue("as") {
		return index
	}
	ast.AddChildren(index)
	consume()
	ast.AddChildToken(expectType())
	return ast
}

/*
 * index = term { "[" index_value "]" } ;
 */
func parseIndex() AST {
	//fmt.Println("In index with:", peek())
	var operand AST
	ast := AST{label: "index"}
	operand = parseTerm()
	if !checkValue("[") {
		return operand
	}
	ast.AddChildren(operand)
	for checkValue("[") {
		consume()
		operand = ast
		ast = AST{label: "index"}
		ast.AddChildren(operand, parseIndexValue())
		expectValue("]")
	}
	return ast
}

/*
 * term = literal | member | call | "(" expression ")" ;
 */
func parseTerm() AST {
	//fmt.Println("In term with:", peek())
	if isLiteral() {
		return parseLiteral()
	}
	if checkKind(lexer.ID) || checkKind(lexer.LIT_STRING) {
		if checkValueAhead(".", 1) {
			for i := state.ptr + 2; i < state.length-2; i += 2 {
				curr := state.tokens[i]
				next := state.tokens[i+1]
				if curr.Kind == lexer.ID && next.Value == "(" {
					return parseCall()
				} else if curr.Kind == lexer.ID && next.Value != "." {
					break
				}
			}
			return parseMember()
		}
		if checkValueAhead("(", 1) {
			return parseCall()
		}
		return parseMember() // identifier
	}
	if checkValue("(") {
		consume()
		expr := parseExpression()
		expectValue(")")
		return expr
	}
	state.addError(fmt.Sprintf("Expected expression but got %s", peek().GetValueString()))
	return AST{token: lexer.Token{
		Kind:    lexer.Virtual,
		Value:   "term",
		Missing: true,
		Line:    peek().Line,
		Column:  peek().Column,
	}}
}

func isLiteral() bool {
	return (checkKind(lexer.LIT_CHAR) ||
		(checkKind(lexer.LIT_STRING) && !checkValueAhead(".", 1)) ||
		checkKind(lexer.KW_BOOLVALUE) ||
		checkKind(lexer.LIT_FLOAT) ||
		checkKind(lexer.LIT_HEX) ||
		checkKind(lexer.LIT_INT) ||
		(checkKind(lexer.OPERATOR_ADD) &&
			(checkKindAhead(lexer.LIT_FLOAT, 1) ||
				checkKindAhead(lexer.LIT_HEX, 1) ||
				checkKindAhead(lexer.LIT_INT, 1))) ||
		(checkKind(lexer.ID) && checkValueAhead("{", 1)))
}

/*
 * index_value =  slice | expression | array_end ;
 */
func parseIndexValue() AST {
	//fmt.Println("In index value with:", peek())
	if isSlice() {
		return parseSlice()
	}
	if checkValue("^") && checkValueAhead("]", 2) {
		return parseArrayEnd()
	}
	return parseExpression()
}

func isSlice() bool {
	for i := state.ptr; i < state.length-1 && state.tokens[i].Value != "]"; i++ {
		if state.tokens[i].Kind == lexer.OPERATOR_RANGE {
			return true
		}
	}
	return false
}

/*
 * slice = [ expression | array_end ] range_operator [ expression | array_end ] ;
 */
func parseSlice() AST {
	//fmt.Println("In slice with:", peek())
	ast := AST{label: "slice"}
	if checkKind(lexer.OPERATOR_RANGE) {
		rangeOp := AST{token: consume()}
		if peek().Value == "]" {
			ast.AddChildren(rangeOp)
			return ast
		} else {
			if checkValue("^") {
				ast.AddChildren(parseArrayEnd())
			} else {
				ast.AddChildren(parseExpression())
			}
			return ast
		}
	} else if checkValue("^") {
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
	//fmt.Println("In array_end with:", peek())
	ast := AST{label: "ARR-END"}
	ast.AddChildToken(consume())
	ast.AddChildren(parseExpression())
	return ast
}

/*
 * literal = bool_literal | char_literal | string_literal | number_literal | struct_literal;
 */
func parseLiteral() AST {
	if checkKind(lexer.LIT_CHAR) || checkKind(lexer.LIT_STRING) || checkKind(lexer.KW_BOOLVALUE) {
		return AST{token: consume()}
	} else if checkKind(lexer.LIT_INT) || checkKind(lexer.LIT_HEX) || checkKind(lexer.LIT_FLOAT) || checkValue("+") || checkValue("-") {
		return parseNumLiteral()
	} else {
		return parseStructLiteral()
	}
}

/*
 * number_literal = [ "+" | "-" ] ( hex | float | int ) ;
 */
func parseNumLiteral() AST {
	//fmt.Println("In number_literal with:", peek())
	sign := "+"
	var token lexer.Token
	if checkValue("+") || checkValue("-") {
		sign = consume().Value
		token = consume()
		token.IsSigned = true
	} else {
		token = consume()
	}
	switch token.Kind {
	case lexer.LIT_INT, lexer.LIT_HEX:
		if sign == "-" {
			token.IntVal = -token.IntVal
		}
	case lexer.LIT_FLOAT:
		if sign == "-" {
			token.FloatVal = -token.FloatVal
		}
	}
	return AST{token: token}
}

/*
 * struct_literal = identifier "{" [ properties ] "}";
 */
func parseStructLiteral() AST {
	//fmt.Println("In struct_literal with:", peek())
	ast := AST{label: "struct_literal"}
	ast.AddChildToken(expectKind(lexer.ID))
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
func parseProperties() AST {
	//fmt.Println("In properties with:", peek())
	ast := AST{label: "properties"}
	ast.AddChildren(parseProperty())
	for !checkValue("}") && !checkKind(lexer.EOF) {
		if !checkKindAhead(lexer.ID, 1) {
			break
		}
		expectValue(",")
		ast.AddChildren(parseProperty())
	}
	if checkValue(",") {
		consume()
	}
	return ast
}

/*
 * property = identifier ":" expression ;
 */
func parseProperty() AST {
	//fmt.Println("In property with:", peek())
	ast := AST{label: "property"}
	ast.AddChildToken(expectKind(lexer.ID))
	expectValue(":")
	ast.AddChildren(parseExpression())
	return ast
}

/*
 * modifiers = "private" [ "mut" ] | "mut" [ "private" ] ;
 */
func parseModifiers() AST {
	//fmt.Println("In modifiers with:", peek())
	ast := AST{
		label: "modifiers",
	}
	modifier := consume()
	ast.AddChildToken(modifier)
	if checkKind(lexer.KW_MODIFIER) {
		if checkValue(modifier.Value) {
			message := fmt.Sprintf("Invalid variable modifiers: %s %s", modifier.Value, peek().Value)
			state.addError(message)
			consume()
		} else {
			ast.AddChildToken(consume())
		}
		/*if checkKind(lexer.KW_MODIFIER) {
			state.addError("Too many variable modifiers")
			consume()
		}*/
	}
	return ast
}

/*
 * member = ( identifier | string_literal ) { "." identifier } ;
 */
func parseMember() AST {
	//fmt.Println("In member with:", peek())
	ast := AST{token: consume()}
	for checkValue(".") {
		lhs := ast
		ast = AST{label: "dot"}
		consume() // skip over the dot
		ast.AddChildren(lhs, AST{token: expectKind(lexer.ID)})
	}
	return ast
}

/*
 * call = member "(" [  expression { "," expression } ]")" ;
 */
func parseCall() AST {
	//fmt.Println("In call with:", peek())
	ast := AST{label: "call"}
	ast.AddChildren(parseMember())
	expectValue("(")
	if checkValue(")") {
		consume()
		return ast
	}
	params := AST{label: "params"}
	params.AddChildren(parseExpression())
	for checkValue(",") {
		consume()
		params.AddChildren(parseExpression())
	}
	ast.AddChildren(params)
	expectValue(")")
	return ast
}

/*
 * control_flow = "return" [ expression ] | "continue" | "break" ;
 */
func parseControlFlow() AST {
	//fmt.Println("In control_flow with:", peek())
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
