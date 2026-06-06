Formal language grammar for parser

```ebnf
(* structural *)
program     =   { declaration } ;
declaration = function | struct | interface | variable ;
function = "fn" identifier "(" [ parameters ] ")" [ "->" type ] ( ";" | block ) ;
parameters =  parameter { "," parameter } ;
parameter = type identifier ;
block = "{" { statement } "}" ;
statement = ( ( variable | assignment | expression | control_flow ) ";" ) | branch ;
branch = if_block | while | for ;
expression = logical_or ;
struct = "struct" identifier [ "impl" interface_list ] struct_body ;
interface_list = identifier { "," identifier };
struct_body =  "{" { variable | function | named_block } "}" ;
named_block = identifier "{" { function | variable } "}" ;
interface = "interface" identifier "{" { function } "}" ;
variable = [ modifiers ] type identifier [ assignment ] ;
if_block = if { "else" if } [ "else" conditional_body ] ;
if = "if" "(" expression ")" conditional_body ;
conditional_body = block | statement ;
while = "while" "(" expression ")" block;
for = "for" "(" for_conditions ")" block ;
for_conditions = ( ( variable | assignment ) ";" expression ";" expression ) | ( variable [ "," variable ] "in" range ) ;
range = expression [ range_operator expression [ ".." expression ] ] ;  

(* operators: reverse order of precendence *)
assignment = member assign_operator expression ;
assign_operator =  "=" | "+=" | "-=" | "*=" | "/=" ;
logical_or = logical_and { "||" logical_and } ;
logical_and = logical_not { "&&" logical_not } ;
logical_not = [ "!" ] comparison ;
comparison = bitwise [ compare_operator bitwise ] ;
compare_operator = "==" | "!=" | "<" | ">" | "<=" | ">=" ;
bitwise =  add { bitwise_operator add };
bitwise_operator = "^" | "&" | "|" ;
add = mult { ( "+" | "-" ) mult } ;
mult = expo { multiplication_operator expo } ;
expo = unary { "**" expo } ; 
multiplication_operator = "*" | "/" | "%" ;
unary = left_unary | right_unary ;
left_unary = [ "^" | "-" | right_unary_operators ] typecast ;
right_unary = typecast [ right_unary_operators ] ;
right_unary_operators = "++" | "--" ;
typecast = index [ "as" type ] ;
index = term { "[" index_value "]" } ;
term = literal | member | call | "(" expression ")" ;
index_value =  slice | expression ;
slice = [ expression | array_end ] range_operator [ expression | array_end ] ;
range_operator = ".." [ "=" ] ;
array_end = "^" ( ( "1" ... "9" ) { "0" ... "9" } ) ;

(* literals *)
literal = bool_literal | char_literal | string_literal | number_literal | struct_literal;
bool_literal = "true" | "false" ;
char_literal = "'" .+ "'" ;
string_literal = '"' .+ '"';
number_literal = [ "+" | "-" ] ( hex | float | int ) ;
hex = "0x" ( "0" ... "9" | "a" ... "f" | "A" ... "F" )+ ;
float = [ int ] "." int+ ;
int = ("0" ... "9" )+ ;
struct_literal = identifier "{" [ properties ] "}";
properties =  property { ","  property } [ "," ] ;
property = identifier ":" expression ;

(* variable info *)
modifiers = "private" | "mut" | "private mut" | "mut private" ;
type = "int" | "int64" | "uint32" | "uint64" | "float" | "double" | "String" | "char" | "bool" | identifier ;
member = identifier { "." identifier } ;
identifier =  ( "A" ... "Z" | "a" ... "z" | "_" ) { "A" ... "Z" | "a" ... "z" | "_" | "0" ... "9" } ;

(* control flow *)
call = member "(" [  expression { "," expression } ]")" ;
control_flow = "return" [ expression ] | "continue" | "break" ;
```