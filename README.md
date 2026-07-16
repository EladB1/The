# The

![Language Icon](https://preview.redd.it/i-tried-recreating-the-the-from-sponge-bob-v0-ua5uaaq0gilg1.jpg?width=582&format=pjpg&auto=webp&s=3a830e4421eae16ce0b4ce282c437b200c15df3b)

A statically typed language which compiles to WebAssembly

The goal of **The** language is to start with something simple to build a compiler for and expand from there.

#### Initial version features

- [ ] Static strong typing
- [ ] Immutability by default; `mut` used to declare mutability
- [ ] User defined types and interfaces
- [ ] Generates WAT file and automatically calls `wat2wasm` on it

Future state features can be found [here](TODO.md)

 > Subject to change as development progresses

Table of contents
---

1. [Installation and CLI Usage](#installation-and-cli-usage)
2. [Language specifications](#language-specifications)
3. [Compiler Developer Info](#compiler-developer-info)

## Installation and CLI Usage

> TODO: installation

For getting all command line options run: `the -h` or `the --help`

A `.the` file is required as input. For example:

`the examples/src/functions.the` is valid

`the` and `the file.txt` are not valid

## Language Specifications

[Formal grammar](grammar.md)

### Entry point

Right now, the language doesn't support splitting code up into multiple files, so every file is treated as its own application. Applications must have an entry point so each `.the` file must have a main function defined that returns an `int` (the status code after your program completes).

```

fn main() -> int {
    //
    if (condition)
        return 1; // indicates error
    return 0; // indicates success
}
```

When a program is compiled all type, function, and variable definitions will be placed in the symbol table. When the program runs, it will go through `main` line by line and only use definitions function/type definitions included in main.

### Comments

Inline comments can be made using `//`. 

Multiline comments start with `/*` and end with `*/`.

### Operators

Operator types in order of precedence:

1. Indexing: `[]`
2. Typecasting: `as`
3. Unary: `++`, `--`, `^` (string end shortcut found [here](#strings))
4. Exponents: `**`
5. Multiplication / Division / Modulo: `*`, `/`, `%`
6. Addition/Subtraction: `+`, `-`
7. Bitwise (binary): `&`, `|`, `^` (XOR), `>>`, `<<`
8. Comparison: `==`, `!=`, `<`, `<=`, `>`, `>=`
9. Logical Not: `!`
10. Logical And: `&&`
11. Logical Or: `||`
12. Assignment: `=`, `+=`, `-=`, `*=`, `/=`

### Primitive Types

| Type | Default Value |
| --- | --- |
| int | 0 |
| int64 | 0 |
| uint32 | 0 |
| uint64 | 0 |
| float | 0 |
| double | 0 |
| bool | false |
| char | '' |
| String | "" |

### Strings

The `String` type is a shorthand for an array of `char` values. 

You can get the length of a string using `.length`. Strings can be indexed to get the character at that location. 

```
String greeting = "hello";
char c = greeting[1]; // 'e'
int len = greeting.length;
bool couldBePalindrome = greeting[0] == greeting[greeting.length - 1]; // comparing the first and last characters
```
There is a shorthand for `string.length - number` using the unary operator `^` followed by a value greater than 0. Think of `str[^n]` as `str[str.length - n]` so array bounds still apply.

```
String str = "Helloo!";
int n = -5;
char c = str[^0]; // error
char c = str[^n]; // error
char c = str[^1]; // valid; value would be '!'
char c = str[^(n * -1)]; // valid; value would be 'e'
```

Strings can also be sliced. A slice will take a portion of the string, based on the provided start and end, and return a new string which is made up of the values at the index that fit those ranges. Both bounds of the slice must fit within array bounds and can be omitted. The left value must be less than or equal to the right value. 

The syntax is `string[start .. end]` where it will begin at `start` and stop before it gets to `end` (not inclusive). To make it inclusive the syntax is `string[start ..= end]`; this will go to start and stop after it gets to `end`. Examples:

```
String data = "abcde";
data[4 .. 2]; // error; left value > right value
data[2 .. 4]; // valid; returns "cd"
data[2 ..= 4]; // valid; return "cde"
data[2 .. ]; // valid; returns "cde"
data[2 ..= ]; // error; this is trying to include nth character in a string of length n (bounds error)
data[.. 4]; // valid; returns "abcd"
data[..= 4]; // valid; returns "abcde"
data[..]; // valid; returns full string
data[..=]; // error; this is trying to include nth character in a string of length n (bounds error)

// Using string end shorthand
data[^3 .. ^1]; // valid; returns "cd"
data[^3 ..= ^1]; // valid; returns "cde"
data[^3 ..]; // valid; returns "cde"
data[.. ^1]; // valid; returns "abcd"
data[..= ^1]; // valid; returns "abcde"
```

### Type Compatibility

| Operation | Result | Valid | Context |
| --- | --- | --- | --- |
| `int` **operator** `int64` | `int64` | ✅ | |
| `int` **operator** `float` | `float` | ✅ | |
| `int` **operator** `double` | `double` | ✅ | |
| `int64` **operator** `float` | **Error** | ❌ | 64 bit signed int too large for 32 bit float |
| `int64` **operator** `double` | `double` | ✅ | |
| `float` **operator** `double` | `double` | ✅ | |
| `uint32` **operator** `uint64` | `uint64` | ✅ | |
| `uint32`/`uint64` **operator** any other type | **Error** | ❌ | signed / unsigned conversions are unclear and/or lead to bugs |
| `String + char` | `String` | ✅ | Only concatenation supported between them |
| `String` **operator** any other type | **Error** | ❌ | |
| `char + char` | `String` | ✅ | Concatenate characters to form a `String` |
| `char + String` | `String` | ✅ | Only concatenation supported between them |
| `char` **operator** any other type | **Error** | ❌ | |
| `bool` **operator** any other type | **Error** | ❌ | |

### Type casting

Primitive types can be cast to each other (depending on the original and target type) using the `as` keyword. Any lossy conversions will result in a warning; types that cannot support casting from the original type will result in an error. 

All primitive types can be casted to a `String`. If you want to print something like `"My age is 12"` and the value of 12 was stored in an `int` variable `age`, you would need to do `println("My age is " + age as String);`.

Example:

```
int i = 0;
double phi = 1.618;
uint32 age = 10;

double j = i as double; // valid
int lossy = phi as int; // warning: lossy conversion

if (age as int == i) {} // valid; helps bridge incompatability between `int` and `uint32` types
double something = phi + age as double; // valid; helps bridge incompatability between `double` and `uint32` type
```

| Original | Target | Valid |
| --- | --- | --- |
| `int` | `int64` | ✅ |
| `int` | `float` | ✅ |
| `int` | `double` | ✅ |
| `int64` | `int` | ⚠️ |
| `int64` | `float` | ⚠️ |
| `int64` | `double` | ✅ |
| `float` | `double` | ✅ |
| `double` | `float` | ⚠️ |
| `uint32` | `int` | ✅ |
| `uint32` | `int64` | ✅ |
| `uint32` | `float` | ✅ |
| `uint32` | `double` | ✅ |
| `uint64` | `int` | ✅ |
| `uint64` | `uint32` | ⚠️ |
| `uint64` | `float` | ⚠️ |
| `uint64` | `double` | ✅ |
| `bool` | numeric | ❌ |
| `bool` | `char` | ❌ |
| `char` | numeric | ❌ |
| `char` | `bool` | ❌ |
| any | `String` | ✅ |
| `String` | any | ❌ |

Typecasting will help with operations on incompatible types, but won't fix all incompatabilities.

### Variables

All variables are immutable by default; to define a mutable variable, preface the definiton with `mut`.
An immutable variable must have an assigned value when defined.

```
bool isOpen = false; // Cannot be changed
mut bool isCorrect = verify(value); // Can be changed

double pi; // This will cause an error since pi is immutable but has no value
mut double e; // Since it's mutable, a value can be set later so this is valid
```

Any variables defined within a function (including parameters) are local and any properties within a `struct` are local to that.
Any variables outside of a function or struct definition are global and global variables can be mutable but will produce a warning.
Function parameters cannot be mutable.

### Functions

No parameters and no return:

```
fn printHello() {
    println("Hello, world!");
}

//...

printHello();
```

With parameters and return type:

```
fn divisibleByTwo(int i) -> bool {
    return i % 2 == 0;
}

// ...

bool isEven = divisibleByTwo(value);
```

The `divisibleByTwo` function takes an `int` parameter and returns a `bool` value.

A function named can be reused as long as the return type is the same and the parameters are different which would make their function signatures different but compatible. Example:

```
fn test() -> TestResult {};
fn test(TestParam param) -> TestResult {}; // valid since the have the same return type and different parameters
fn test() -> TestResult {}; // invalid; same signature defined twice
fn test() -> int {}; // invalid; different return type
```

### Loops

Loops come in different shapes depending on the use case

There are two types of loops: `while` and `for`

While loops are like most other languages:

```
while (condition) {
    // code goes here
}
```

Since variables are immutable by default, there are restrictions on using C-style for loops. The code below will not work since it violates the immutability rules enforced by the compiler.

```
for (int i = 0; i < limit; i++) {
    // code goes here
} // this will fail since i is immutable but the loop is trying to increment it
```

For iterating a set of int values, this syntax can be used instead:

```
for (int i in 0 .. limit) {
    // code goes here
}
```

This is equivalent to `for (int i = 0; i < limit; i++)` but the compiler will handle the mutability of `i` for you so no error happens. Within the body of the loop `i` cannot be changed.

What if you wanted to change `i < limit` to `i <= limit`?

```
for (int i in 0 ..= limit) {
    // code goes here
}
```

So far we've only handled the case of `i++` but what if we want to use a different increment value than 1?

```
for (int i in limit ..= 0 .. -2) {
    // code goes here
}
```

The loop above will start at limit and stop when it's greater than or equal to 0, it will iterate by -2 each iteration.

So far, we've only handled addition... what if you want multiplication as the iteration part of the loop?

For that, you would have to use a C-style loop with a mutable iterator (can also be done for any of the above cases).

```
    for (mut int i = 1; i < limit; i *= 2) { // i is mutable so this will work
        // code goes here
    }
```

Looping over strings works in a similar way to the for loops covered already:

```
for (char c in "Hello, world!") {
    // code goes here
}
```

"The" will automatically set c to each character in the string and stops when it hits the end.

If you want to include the index of the character, you can change the loop variables to support that.

```
for (int i, char c in "Hello, world!") {
    // code goes here
}
```

The compiler will automatically handle setting both the `i` with the index of the character in the string and `c` with the character itself.

> C-style mutable loops can be used as well for iterating over strings but indexing a string has not been planned yet.

Flow control keywords `break` (to stop the loop) and `continue` (to skip to the next iteration) are only valid in loops.

### User defined types

AKA structs

A user can define a type which includes its own data and methods

```
struct Employee {
    uint32 id;
    String name;
    float salary;
}
Employee emp = Employee {
    id: 12,
    name: "CEO",
    salary: 0.01,
}; // can include or omit trailing comma
println(emp.id)
```

In the definition of a `struct`, all properties must have a type and end with a semi-colon.

By default, everything is public, but you can use `private` or create a `private` block to prevent outside access to properties.

```
struct Account {
    String name;
    private {
        uint32 number;
        String SSN;
    } // contain multiple values rather than writing private multiple times
    private float balance; // private used outside of a block
}
```

A `struct` can also have functions embedded within its definition. Struct functions can be in the same `private` blocks as properties or be explicitly marked with `private`. All struct properties are in scope and can be referenced with their names or the developer can explicitly use `this.propertyName`.

If a struct has a property or method with the same name as something global declared, the default behavior is to check the struct's inner scope first so `property` and `method()` (also `this.property` and `this.method()`) would only come from within the struct. To be able to call the global versions, you can use `global.propety` and `global.method()`

> **NOTE**: `this` and `global` are reserved variables in every struct so it cannot be used as a variable/function/property name

```
    struct File {
        String name;
        String path;
        Permissions permissions;
        uint64 size;

        fn read() -> String {
            String fullPath = path + name; // both in scope
            if (!checkPermissions()) {
                return "";
            }
            // ...
            return fileContents;
        }

        private {
            String owner;
            String group;
            fn checkPermissions() -> bool { // checkPermissions cannot be called directly outside of the struct definition
                if (this.permissions.asInt() != 755) { // explicit use of this
                    return false;
                }
                return true;
            }
        }
    }
```

The `struct` functions allow you to use the syntactic convention of `instanceName.function()` rather than having to pass the `struct` type as a parameter

```
File file = File {
    //...
};
String doc = file.read();
```

A mutable instance of a `struct` means that any of its public properties are mutable as well; private properties can be updated via methods but not directly.
A private variable can be mutable independent of any instance and the declaration order does not matter.
For example:

```
struct Time {
    uint32 hour;
    uint32 minute;
    uint32 second;
    private uint64 nanoseconds;
    // both lines below are valid and equivalent
    private mut String timezone;
    mut private String region;
}

mut Time now = Time {
    hour: 12,
    minute: 0,
    second: 0,
    nanoseconds: 0
};
now.hour = 23;
now.minute = 59;
now.second = 59;
now.setNanoSeconds(1111111); // only works the instance is mutable or the private property is marked as mutable

```

All user defined types can be typecasted to a `String` using `instance as String` where `instance` is a variable with a `struct` type; the default implementation of this typecasting prints only the public properties and none of the private ones. The default `String` typecasting implementation can be changed, but that will be covered further in this document.

Any struct values not instantiated or set on the instance will be treat as their default values (or an empty struct if they're not a primitive type). An empty Struct instance can be declared with `MyType instance = MyType {}`. Mutability rules still apply to empty structs.

### Named blocks

You may have noticed this in the [User defined types](#user-defined-types) section:

```
private {}
```

This is called a named block which gives the compiler extra information about aspects of the new type. 
There are some built-in named blocks which can be used to add functionality to your types.

`private`
---

The `private` named block tells the compiler about the visibility of the properties and functions contained within. Any code outside of the struct that tries to reference something in a `private` block outside of the struct definition (on the instance) will cause the compiler to throw an error. The `private` named block is the **only** one that can contain properties; the rest can only contain functions.

If you mark a variable within a `private` block as `private`, the compiler will warn you about it since it is unnecessary to mark it multiple times. Example:

```
private {
    int x;
    private int y; // generates compiler warning
}
```

`cast`
---

The `cast` named block tells the compiler about how to handle type casting between the `struct` and any other type. Casting can be done using `as TargetType`. For example:

```
struct TypeA { bool aProp; }
struct TypeB {
    int bProp;

    cast {
        fn toString() -> String { // override the default String casting for this type
            return "I am TypeB";
        }

        fn toTypeA() -> TypeA {
            return TypeA {
                aProp: bProp == 0,
            };
        }
        /*
        fn toTypeA2() -> TypeA { // this would cause an error since the compiler won't know which function to use for type casting
            return TypeA {};
        }
        */
    }
}

TypeB typeb = TypeB { bProp: 0 };
TypeA casted = typeb as TypeA; // valid now that a type casting function has been written
```

The name of the casting function does not really matter (but should be appropriately descriptive) as long as the return type is the one being casted to. If multiple casting functions are defined for the same type, that would be an error. Casting functions cannot take any parameters and must have a return type.

It should be noted that the type casting of one user defined type to another, creates a new instance of the target type; casting will not do an in-place transformation and will not clean up the memory associated with the original object.

`compare`
---

By default all instances of user defined types are comparable with each other using `==` and `!=`. The two instances will run an equality check on each subfield (including `private` ones).

In order to either overwrite equality/non-equality or add support for other comparisons, the `struct` can contain a `compare` block which has special functions that can be written to overload the operators. The function signatures are:

1. `fn equals(MatchingStructType any_name_here) -> bool`
2. `fn lessThan(MatchingStructType any_name_here) -> bool`
3. `fn greaterThan(MatchingStructType any_name_here) -> bool`

> If the parameter type doesn't match the containing `struct`, it will result in an error

All three functions are optional, but anything in the `compare` that is not one of these three functions (or is a duplicate) is an error.

```
struct Response {
    int status;
    private String headers;

    compare {
        fn equals(Response other) -> bool { // overwrite the default equality implementation
            return this.status == other.status;
        }

        fn lessThan(Response other) -> bool { // check if this instance is less than the other instance
            return this.status < other.status;
        }

        fn greaterThan(Response other) -> bool {
            return this.status > other.status;
        }
    }
}

Response valid = Response { status: 200 };
Response notFound = Response { status: 404 };

if (valid != notFound) {
    return valid <= notFound;
}
```

It's important to note that even if only `equals`, `lessThan`, and `greaterThan` are supported, the compiler can extend equality to also support non-equality (`!=`), extend `lessThan` to also support less than or equal to (`<=`), and extend `greaterThan` to also support greater than or equal to (`>=`).

> **NOTE**: If there are multiple of the same named blocks, they will be combined together

### Interfaces

An interface is just a contract of functions that can be used. 

Interface functions usually don't have bodies, but if they do, that will be treated as the default implementation. If a function has a body, any consumers of the interface don't have to implement that function, but if they do, their implementation would override the default.

Any types that implement all functions (other than default implementations) within the interface, can be considered to be the same type. 
If a function accepts the `Vehicle` interface as a parameter and the `Car` type implements the interface, then the function can take a `Car`. 
All interface functions are public and must remain public during implementation.

```
interface Vehicle {
    fn move(int position, int distance) -> int;
}
```

To implement the interface, you must include a named block for each interface that named block must contain all of the interface's functions and nothing else. For example:

```
    interface Device {
        fn powerOn();
        fn powerOff();
    }

    interface Chargable {
        fn charge();
    }

    struct Speaker impl Device, Chargable {
        Device {
            fn powerOn() {
                // implement
            }
            fn powerOff() {
                // implement
            }
        }
        Chargable {
            fn charge() {
                // implement
            }
        }
    }
```

A struct can implement multiple interfaces (comma separated) using the `impl` keyword in the definition.
Each interface must have a block with all of its functions implemented to be properly defined.

An interface cannot be directly instantiated to a variable, but it can be used in the left hand side of a variable declaration only if the right hand side is a `struct` that implements the interface.

```
Device speaker; // invalid
mut Device headphones; // invalid
Device phone = Device{}; // invalid

Device speaker2 = Speaker{}; // valid
mut Device speaker3 = Speaker{}; // valid

speaker3 = NonDeviceStruct{}; // invalid since `NonDeviceStruct` does not implement `Device` so this is violating the type checking
```

### Disambiguation

What if a struct uses multiple interfaces that all have the same function signature within them?

```
interface A {
    fn do(int a);
}

interface B {
    fn do(int b);
}

struct C impl A, B {
    A {
        fn do(int a) {
            // implementation
        }
    }
    B {
        fn do(int b) {
            // implementation
        }
    }
}

C instance = C {};
instance.do(0); // Which do() is used when this is called?
```

The example above would result in an error since the compiler won't know which `do()` to pick, but there's a way to fix it.

The caller must pick a specific implementation of the function. It doesn't matter where in the code this is done.

```
instance.A.do(0);
instance.B.do(0);
```

> **Note**: `instance.Interface.function()` can be be used even if there are no conflicts

### Standard Library

Standard library functions/variables are included without importing anything (imports/exports not current supported)

#### Constants:

| Variable | Type | Description | Approx. Value |
| --- | --- | --- | --- |
| `INT_MIN` | `int` | Smallest 32-bit signed integer | -2<sup>31</sup> | 
| `INT_MAX` | `int` | Largest 32-bit signed integer | 2<sup>31</sup> - 1 |
| `INT64_MIN` | `int64` | Smallest 64-bit signed integer | -2<sup>63</sup> |
| `INT64_MAX` | `int64` | Largest 64-bit signed integer | 2<sup>63</sup> - 1 | 
| `UINT32_MAX` | `uint32` | Smallest 32-bit unsigned integer | 2<sup>32</sup> - 1 |
| `UINT64_MAX` | `uint64` | Largest 64-bit unsigned integer | 2<sup>64</sup> - 1 |
| `FLOAT_MIN` | `float` | Smallest 32-bit float | -3.4 * 10<sup>38</sup> |
| `FLOAT_MIN_POSITIVE` | `float` | Smallest (normal) 32-bit float before reaching 0 | 1.18 * 10<sup>-38</sup>
| `FLOAT_MAX` | `float` | Largest 32-bit float | 3.4 * 10<sup>38</sup> |
| `FLOAT_EPSILON` | `float` | Value for precise float comparisons; represents max rounding error | 1.19 * 10<sup>-7</sup> |
| `FLOAT_NaN` | `float` | NaN for 32 bit float | 0x7FC00000 |
| `FLOAT_INF` | `float` | Positive infinity for 32 bit float | 0x7F800000 |
| `FLOAT_NEG_INF` | `float` | Negative inifinity for 32 bit float | 0xFF800000 |
| `DOUBLE_MIN` | `double` | Smallest 64-bit float | -1.79 * 10<sup>308</sup> |
| `DOUBLE_MIN_POSITIVE` | `double` | Smallest (normal) 64-bit float before reaching 0 | 2.23 * 10<sup>-308</sup> |
| `DOUBLE_MAX` | `double` | Largest 64-bit float | 1.79 * 10<sup>308</sup> |
| `DOUBLE_EPSILON` | `double` | Value for precise double comparisons; represents max rounding error | 2.22 * 10<sup>-16</sup> |
| `DOUBLE_NaN` | `double` | NaN for 64 bit float | 0x7FF8000000000000 |
| `DOUBLE_INF` | `double` | Positive infinity for 64 bit float | 0x7FF0000000000000 |
| `DOUBLE_NEG_INF` | `double` | Negative inifinity for 64 bit float | 0xFFF0000000000000 |
| `PI` | `double` | Value of pi | 3.141592653589793 |
| `E` | `double` | Value of e (Euler's number) | 2.718281828459045 |

> **NOTE**: `FLOAT_MIN_POSITIVE` and `DOUBLE_MIN_POSITIVE` are both normalized value minimums (hardware accelaration supported). Anything smaller than those will be rounded down to 0. Subnormal floats may be supported later.


#### Functions:

| Function | Description
| --- | --- |
| `print(Any value)` | Print value to stdout |
| `println(Any value)` | Print value to stdout with new line ending |
| `printerr(Any value)` | Print value to stderr |
| `typeOf(Any value) -> String` | Get the type of a value as a string | 
| `exit(int status)` | Terminate execution with status code |
| `exit(int status, String error)` | Terminate execution with status code and error message (will print to stderr) |
| `sleep(double seconds)` | Block the thread for specified amount of seconds |
| `getEnv(String key) -> String` | Get the value of environment variable |
| `setEnv(String key, String value)` | Set the value of environment variable |
| `indexOf(String string, char chr) -> int` | Get the index of `chr` |
| `contains(String string, String substring) -> bool` | Check if a string contains a substring |
| `contains(String string, char chr) -> bool` | Check if a string contains a character |
| `startsWith(String string, String prefix) -> bool` | Check if a string starts with `prefix` |
| `endsWith(String string, String suffix) -> bool` | Check if a string ends with `suffix` |
| `replace(String string, String old, String new) -> String` | Replace the first occurence of `old` with `new` |
| `replace(String string, char old, char new) -> String` | Replace the first occurence of `old` with `new` |
| `replaceAll(String string, String old, String new) -> String` |  Replace all occurences of `old` with `new` |
| `replaceAll(String string, char old, char new) -> String` |  Replace all occurences of `old` with `new` |
| `reverse(String string) -> String` | Get a string in reverse order |
| `toUpper(String string) -> String` | Change all characters to uppercase |
| `toLower(String string) -> String` | Change all characters to lowercase |
| `trim(String string) -> String` | Remove whitespace from the start and end of string |
| `trimStart(String string) -> String` | Remove whitespace from the start of string |
| `trimEnd(String string) -> String` | Remove whitespace from the end of string |
| `assert(bool condition)` | Check if a condition is true and fail otherwise |
| `assert(bool condition, String message)` | Check if a condition is true and fail otherwise with a message |
| `prompt(String promptText) -> String` | Print `promptText` and read from stdin |
| `secretPrompt(String promptText) -> String` | Print `promptText` and read from stdin but hide characters being typed |


### Memory Management

If the compiler knows the amount of memory of something ahead of time, it will be allocated on the stack and cleaned up when the scope exits

For dynamic data (strings) and structs containing dynamic data, they will be allocated on the heap.

Memory will be managed using Automatic Reference Counting (ARC). After the MVP of the compiler/language, Cycle Detection will be added as well to deal with some of the issues with ARC.

## Compiler Developer Info

Phases:

    1. Lexical Analysis
    2. Parsing
    3. Semantic Analysis <- WIP
    4. IR Generation
    5. Code Generation (WAT)
    6. Execution via wasmtime

> Optimization phase will come much later

### Tooling

1. Go
2. Make
3. go-snaps (for unit/integration testing)
3. WABT
4. wasmtime

### Testing

Testing strategy:

1. Compiler phase unit tests: Use snapshot testing to confirm valid, warnings, and errors for each phase
2. Integration tests: Use snapshot testing to confirm the inter-phase behavior of the compiler
3. Fuzzing: Make sure each phase doesn't have unexpected crashes, infinite loops, or unrecoverable panics
4. Execution test: Make sure the generated code is valid and produces expected results

As you move down the test types, the number of tests decreases but the complexity of each test increases.

Snapshot testing using [go-snaps](https://github.com/gkampitakis/go-snaps). When changing snapshots, you **must** review them to verify their correctness.

The [Makefile](Makefile) has options for updating snapshots, getting coverage information, and running the different types of tests.

High test coverage is the goal (not enforced yet). Aim for at least 85%-90% coverage of unit tests.


### Parser Error Recovery

The parser uses several strategies to recover from errors. It will insert virtual tokens into the AST if a simple expected token is missing. For syntax errors that could span several tokens, the parser will look for a synchronization point (i.e. `;`, `}`, `)`, `fn`, `struct`, `interface`) to reduce the amount of noise generated; anything between the current token and synchronization token will be discarded. 

Since the compiler will stop if the parser produces any errors, it's less critical to have an accurate AST generated for errors than for valid code, but being as close to accurate as possible will help with debugging certain issues in the compiler.

> **NOTE:** Parser error handling is not entirely complete yet. Known limitations include cascading errors, misinterpretation of developer intent, and improper reading of nested structures. This will be an improvement after the completion of the compiler MVP.

### Semantic Analysis Design

Semantic analysis will run in multiple passes over the AST.

The passes will be:

1. Custom type names: Get interface and struct names; error on any duplicates
2. Analyze interface function signatures: Collect function signatures from each interface
3. Analyze struct function signatures: Collect function signatures from each struct
4. Collect function signatures
5. Analyze global variables
6. Analyze interface function bodies
7. Analyze struct function bodies
8. Analyze interface implementation: Make any structs that claim to implement an interface actually do
9. Analyze function bodies

Scopes will be a tree of scope where each node contains the following:
    
    - interface symbol table
    - struct symbol table
    - function symbol table
    - variable symbol table

The top two levels of the scope tree are the built-in scope and the global scope; interfaces and structs are only valid on those two levels. The built-in scope will contain anything that's part of the language standard library. The global scope will contain any top level declarations (interfaces, structs, functions, global variables). 

Any nested blocks of code will create a child scope. For example this code

```
fn main() -> int {
    mut int x;
    while (true) {
        if (x >= 5) {
            mut int y = -x;
        }
    }
    return 0;
}
```

Will produce this scope tree:

```
    ------------
    | built-in |
    |   ...    |
    ------------
         |
    ------------
    |  global  |
    | fn: main |
    ------------
         |
    ------------
    |   main   |
    |  var: x  |
    ------------
         |
    ------------
    |  while#0 |
    |   ...    |
    ------------
         |
    ------------
    |   if#0   |
    |  var: y  |
    ------------
```

> The `while#0` and `if#0` are internal scope names used by the compiler to different various scopes which could have the same name