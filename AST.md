AST Examples
---

# Arithmetic Operations

Grammar rule(s):

```ebnf
add = mult { ( "+" | "-" ) mult } ;
mult = expo { multiplication_operator expo } ;
expo = unary { "**" expo } ; 
```

<table>
<tr>
<th><center>Source Code</center></th>
<th><center>AST</center></th>
</tr>
<tr>
<td>

```
2 + 4 + 5
```

</td>
<td>

```
        +
      /   \
     +     5
    / \    
   2   4
```

</td>
</tr>
<tr>
<td>

```
2 + 3 * 5
```

</td>
<td>

```
        +
      /   \
     2     *
          / \
         3   5
```

</td>
</tr>
<tr>
<td>

```
2 ** 3 ** 4
```

</td>
<td>

```
    **
  /    \
 2      **
      /    \
     3      4
```

</td>
</tr>
</table>

# Variable Declaration

Grammar rule(s):

```ebnf

variable = [ modifiers ] type identifier [ "=" expression ] ;
modifiers = "private" [ "mut" ] | "mut" [ "private" ] ;

```

<table>
<tr>
<th><center>Source Code</center></th>
<th><center>AST</center></th>
</tr>
<tr>
<td>

```
int x = 2 * i - 1;
```

</td>
<td>

```
    variable
   /    |   \
 int    x    -
            / \
           *   1
         2  i
```

</td>
</tr>
<tr>
<td>

```
private mut int y;
```

</td>
<td>

```
          variable
        /     |   \
modifiers    int   y
  /     \
private mut
```

</td>
</tr>
</table>


# Function Declaration
Grammar rule(s):

```ebnf
function = "fn" identifier "(" [ parameters ] ")" [ "->" type ] ( ";" | block ) ;
parameters =  parameter { "," parameter } ;
parameter = type identifier ;
block = "{" { statement } "}" ;
statement = ( ( variable | assignment | expression | control_flow ) ";" ) | branch ;
branch = if_block | while | for ;
expression = logical_or | "(" logical_or ")" ;
```

<table>
<tr>
<th><center>Source Code</center></th>
<th><center>AST</center></th>
</tr>
<tr>
<td>

```
fn test() {}
```

</td>
<td>

```
    fn
    |
   test
```

</td>
</tr>
<tr>
<td>

```
fn test() -> int {
    return 0;
}
```

</td>
<td>

```

    fn
  /  |   \
test int  body
           |
        control-flow
         /          \
     return          0
```

</td>
</tr>
<tr>
<td>

```
fn test(int value) -> bool; // equivalent to an empty body
```

</td>
<td>

```
           fn
        /  |    \
    test params bool
           |
           param
           /   \
          int  value
```

</td>
</tr>
<tr>
<td>

```
    fn test(int x, int y) -> bool {
        return x % y == 0;
    }
```

</td>
<td>

```
        fn
    /    |    \     \
  test params bool  body
      /      \        |
 param     param    control-flow
 /   \     /   \    /           \
int  x    int  y   return       ==
                               /  \
                              %    0
                            /  \
                          x     y
```

</td>
</tr>
</table>

# Struct Declaration
Grammar rule(s):

```ebnf
struct = "struct" identifier [ "impl" interface_list ] struct_body ;
interface_list = identifier { "," identifier };
struct_body =  "{" { ( variable ";" ) | function | named_block } "}" ;
named_block = identifier "{" { function | ( variable ";" ) } "}" ;
```

<table>
<tr>
<th><center>Source Code</center></th>
<th><center>AST</center></th>
</tr>
<tr>
<td>

```
struct A{}
```

</td>
<td>

```
struct_def
      |
      A
```

</td>
</tr>
<tr>
<td>

```
struct Coordinate {
    mut int x;
    int y;
} 
```

</td>
<td>

```
         struct_def
        /         \
Coordinate        body
                /      \
        variable        variable
       /   |    \       /       \
modifiers  int   x     int       y
       |
      mut 
```

</td>
</tr>
<tr>
<td>

```
struct B impl Read,Write {
    private {
        String owner;
    }
    Read {
        fn read() {}
    }
    Write {
        fn write() {}
    }
}
```

</td>
<td>

```
           struct_def
     /         |           \
    B  interfaces         body
    /    |           /    |  \
Read     Write      NB    NB        NB
                    / |      | \     |    \
            private  body Read body Write body
                      |         |          |
                variable        fn         fn
                /      \         |          |
            String      owner    read      write
```

</td>
</tr>
</table>

# Struct Instance Declaration
Grammar rule(s):

```ebnf
struct_literal = identifier "{" [ properties ] "}";
properties =  property { ","  property } [ "," ] ;
property = identifier ":" expression ;
```

<table>
<tr>
<th><center>Source Code</center></th>
<th><center>AST</center></th>
</tr>
<tr>
<td>

```
Employee emp = Employee {
    title: "CEO",
    rate: 100,
    payType: "Annual", // trailing comma optional
};
```

</td>
<td>

```
        variable
       /   |    \
Employee  emp   struct_literal
                /           \
        Employee         properties
                        /     |      \
                property   property   property
                   /    \     |  \    |       \
                title  "CEO" rate 100 payType "Annual"
```

</td>
</tr>
<tr>
<td>

```
AST{}
```

</td>
<td>

```
    struct_literal
          |
         AST
```

</td>
</tr>
</table>

# Unary operator
Grammar rule(s):

```ebnf
unary = left_unary | right_unary ;
left_unary = [ "-" | right_unary_operators ] typecast ;
right_unary = typecast [ right_unary_operators ] ;
right_unary_operators = "++" | "--" ;
typecast = postfix [ "as" type ] ;
postfix = primary { postfix_op } ;
primary = literal | identifier | "(" expression ")" ;
postfix_op = "." identifier | "(" [  expression { "," expression } ] ")" | "[" index_value "]" ;
```

<table>
<tr>
<th><center>Source Code</center></th>
<th><center>AST</center></th>
</tr>
<tr>
<td>

```
i++
```

</td>
<td>

```
        unary
       /     \
      i       ++
```

</td>
</tr>
<tr>
<td>

```
-"1" as int64
```

</td>
<td>

```
        unary
       /     \
      -    typecast
          /        \
        "1"         int64
```

</td>
</tr>
</table>

# String/Array index
Grammar rule(s):

```ebnf
postfix = primary { postfix_op } ;
primary = literal | identifier | "(" expression ")" ;
postfix_op = "." identifier | "(" [  expression { "," expression } ] ")" | "[" index_value "]" ;
index_value =  slice | expression | array_end ;
slice = [ expression | array_end ] range_operator [ expression | array_end ] ;
range_operator = ".." [ "=" ] ;
array_end = "^" expression ;
```

<table>
<tr>
<th><center>Source Code</center></th>
<th><center>AST</center></th>
</tr>
<tr>
<td>

```
"Hello"[1]
```

</td>
<td>

```
    index
   /     \
"Hello"   1
```

</td>
</tr>
<tr>
<td>

```
str[^n+1]
```

</td>
<td>

```
    index
   /     \
str      arr-end
            |
            +
          /   \
         n     1
```

</td>
</tr>
<tr>
<td>

```
getName(person)[1 ..= ^i]
```

</td>
<td>

```
              index
            /       \
        call         slice  
       /    \       /  |  \
getName     params 1  ..=  arr-end
              |                |
            person             i

```

</td>
</tr>
<tr>
<td>

```
a[1][0]
```

</td>
<td>

```
     index
    /     \
  index    0
 /     \
a       1
```

</td>
</tr>
</table>

# Member Access / Function Call

Grammar rule(s):

```ebnf
assignment = postfix assign_operator expression ;
assign_operator =  "=" | "+=" | "-=" | "*=" | "/=" ;
postfix = primary { postfix_op } ;
primary = literal | identifier | "(" expression ")" ;
postfix_op = "." identifier | "(" [  expression { "," expression } ] ")" | "[" index_value "]" ;
```

<table>
<tr>
<th><center>Source Code</center></th>
<th><center>AST</center></th>
</tr>
<tr>
<td>

```
shape.length = 5;
```

</td>
<td>

```
      assign
     /      \
    dot      5
  /     \
shape   length
  
```

</td>
</tr>
<tr>
<td>

```
func(x.a, y.getB())
```

</td>
<td>

```
        call
       /    \
     func    params
            /      \
          dot      call
          / \        |
         x   a      dot
                    / \
                   y  getB
```

</td>
</tr>
<tr>
<td>

```
doNothing();
```

</td>
<td>

```
     call
      |
   doNothing
```

</td>
</tr>
<tr>
<td>

```
"hello".length
```

</td>
<td>

```
        dot
       /   \
"hello"     length

```

</td>
</tr>
<tr>
<td>

```
a.b.c.d
```

</td>
<td>

```
        dot
       /   \
      dot   d
     /   \
  dot     c
 /   \
a     b
```

</td>
</tr>
<tr>
<td>

```
a.b.c.Action(1+5, -5)
```

</td>
<td>

```
        call
       /    \
      dot   params   
     /   \     |   \
  dot  Action  +   unary
     /   \    / \   |   \
  dot     c  1   5  -    5
 /   \
a     b

```

</td>
</tr>
</table>

# if statements

Grammar rule(s):

```ebnf
if_block = if { "else" if } [ "else" conditional_body ] ;
if = "if" "(" expression ")" conditional_body ;
conditional_body = block | statement ;
```

<table>
<tr>
<th><center>Source Code</center></th>
<th><center>AST</center></th>
</tr>
<tr>
<td>

```
if (true)
    i++
```

</td>
<td>

```
    if-block
       |
       if
      /  \
    true  unary
         /     \
        i       ++
```

</td>
</tr>
<tr>
<td>

```
if (isOpen)
    x = 0;
else if (x < 10) {
    x += 10;
    println(x);
} else {
    continue;
}
```

</td>
<td>

```
            if-block
      /          |          \
     if         else if   else
   /    \         |   \       \
isOpen cond-body  < cond-body  cond-body
          |      / \    |   \           \
          =     x   10  +=    call      control-flow
         / \           / \   /     \           |
        x  10         x  10 println params    continue
                                      |
                                      x
```

</td>
</tr>
</table>

# Loops

Grammar rule(s):

```ebnf
while = "while" "(" expression ")" block;
for = "for" "(" for_conditions ")" block ;
for_conditions = ( ( variable | assignment ) ";" expression ";" expression ) | ( variable [ "," variable ] "in" range ) ;
range = expression [ range_operator expression [ ".." expression ] ] ; 
block = "{" { statement | branch } "}" ;
```

<table>
<tr>
<th><center>Source Code</center></th>
<th><center>AST</center></th>
</tr>
<tr>
<td>

```
    while(x < 10) {
        x++;
    }
```

</td>
<td>

```
    while
   /     \
  <      unary
 / \     /   \
x  10   x     ++
```

</td>
</tr>
<tr>
<td>

```
for (int i in 0..lim) {
    print(i);
}
```

</td>
<td>

```
                    for
                /         \
        condition         loop-body
      /    |     \            |
variable   in     range       call
 /      \         / |  \     /    \
int      i       0  .. lim  print  params
                                    |
                                    i
```

</td>
</tr>
<tr>
<td>

```
for (int i, char c in str[1..6]) {}
```

</td>
<td>

```
                  for
             /           \
        condition       loop-body
      /   |   |   \
    var  var  in  index
   / |  /  \      /    \
 int i char c   str    slice
                      /  |  \
                     1   ..  6 
```

</td>
</tr>
<tr>
<td>

```
for (i = 0; i < lim; i++) {
    sum += i;
}
```

</td>
<td>

```
              for
            /       \
      condition     loop-body
    /     |    \          |
  =       <    unary      +=
 / \     / \    | \      /  \
i   0   i  lim  i  ++   sum  i

```

</td>
</tr>
</table>

---

# Template

Grammar rule(s):

```ebnf

```

<table>
<tr>
<th><center>Source Code</center></th>
<th><center>AST</center></th>
</tr>
<tr>
<td>

```

```

</td>
<td>

```

```

</td>
</tr>
</table>