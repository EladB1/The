AST Examples
---

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
index = term { "[" index_value "]" } ;
term = literal | member | call | expression ;
index_value =  slice | expression | array_end ;
slice = [ expression | array_end ] range_operator [ expression | array_end ] ;
range_operator = ".." [ "=" ] ;
array_end = "^" expression ;
member = ( identifier | string_literal ) { "." identifier } ;
call = member "(" [  expression { "," expression } ]")" ;
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
</table>

# Member Access / Function Call

Grammar rule(s):

```ebnf
member = ( identifier | string_literal ) { "." identifier } ;
call = member "(" [  expression { "," expression } ]")" ;
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