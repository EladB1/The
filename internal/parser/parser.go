package parser

import (
	"github.com/EladB1/The/internal/diagnostic"
	"github.com/EladB1/The/internal/lexer"
)

/*
 *	program = { declaration } ;
 */
func Parse(tokens []lexer.Token) (AST, diagnostic.PhaseDiagnostics) {
	root := AST{}
	report := diagnostic.PhaseDiagnostics{}
	// TODO
	return root, report
}

/*
 *	declaration = function | struct | interface | variable ;
 */
func parseDeclaration() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * function = "fn" identifier "(" [ parameters ] ")" [ "->" type ] ( ";" | block ) ;
 */
func parseFunction() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * parameters = parameter { "," parameter } ;
 */
func parseParameters() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * parameter = type identifier ;
 */
func parseParameter() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * block = "{" { statement } "}" ;
 */
func parseBlock() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * statement = ( ( variable | assignment | expression | control_flow ) ";" ) | branch ;
 */
func parseStatement() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * branch = if_block | while | for ;
 */
func parseBranch() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * expression = logical_or | "(" logical_or ")" ;
 */
func parseExpression() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * struct = "struct" identifier [ "impl" interface_list ] struct_body ;
 */
func parseStruct() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * interface_list = identifier { "," identifier };
 */
func parseInterfaceList() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * struct_body =  "{" { variable | function | named_block } "}" ;
 */
func parseStructBody() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * named_block = identifier "{" { function | variable } "}" ;
 */
func parseNamedBlock() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * interface = "interface" identifier "{" { function } "}" ;
 */
func parseInterface() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * variable = [ modifiers ] type identifier [ assignment ] ;
 */
func parseVariable() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * if_block = if { "else" if } [ "else" conditional_body ] ;
 */
func parseIfBlock() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * if = "if" "(" expression ")" conditional_body ;
 */
func parseIf() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * conditional_body = block | statement ;
 */
func parseConditionalBody() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * while = "while" "(" expression ")" block;
 */
func parseWhile() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * for = "for" "(" for_conditions ")" block ;
 */
func parseFor() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * for_conditions = ( ( variable | assignment ) ";" expression ";" expression ) | ( variable [ "," variable ] "in" range ) ;
 */
func parseForConditions() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * range = expression [ range_operator expression [ ".." expression ] ] ;
 */
func parseRange() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * assignment = member assign_operator expression ;
 */
func parseAssignment() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * logical_or = logical_and { "||" logical_and } ;
 */
func parseLogicalOr() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * logical_and = logical_not { "&&" logical_not } ;
 */
func parseLogicalAnd() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * logical_not = [ "!" ] comparison ;
 */
func parseLogicalNot() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * comparison = bitwise [ compare_operator bitwise ] ;
 */
func parseComparison() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * bitwise =  add { bitwise_operator add };
 */
func parseBitwise() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * add = mult { ( "+" | "-" ) mult } ;
 */
func parseAdd() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * mult = expo { multiplication_operator expo } ;
 */
func parseMult() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * expo = unary { "**" expo } ;
 */
func parseExpo() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * unary = left_unary | right_unary ;
 */
func parseUnary() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * left_unary = [ "^" | "-" | right_unary_operators ] typecast ;
 */
func parseLeftUnary() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * right_unary = typecast [ right_unary_operators ] ;
 */
func parseRightUnary() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * typecast = index [ "as" type ] ;
 */
func parseTypecast() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * index = term { "[" index_value "]" } ;
 */
func parseIndex() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * term = literal | member | call | expression ;
 */
func parseTerm() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * index_value =  slice | expression ;
 */
func parseIndexValue() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * slice = [ expression | array_end ] range_operator [ expression | array_end ] ;
 */
func parseSlice() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * array_end = "^" ( ( "1" ... "9" ) { "0" ... "9" } ) ;
 */
func parseArrayEnd() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * literal = bool_literal | char_literal | string_literal | number_literal | struct_literal;
 */
func parseLiteral() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * struct_literal = identifier "{" [ properties ] "}";
 */
func parseStructLiteral() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * properties =  property { ","  property } [ "," ] ;
 */
func parseProperties() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * property = identifier ":" expression ;
 */
func parseProperty() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * modifiers = "private" [ "mut" ] | "mut" [ "private" ] ;
 */
func parseModifiers() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * member = identifier { "." identifier } ;
 */
func parseMember() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * call = member "(" [  expression { "," expression } ]")" ;
 */
func parseCall() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}

/*
 * control_flow = "return" [ expression ] | "continue" | "break" ;
 */
func parseControlFlow() (AST, diagnostic.PhaseDiagnostics) {
	return AST{}, diagnostic.PhaseDiagnostics{}
}
