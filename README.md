# The

![Language Icon](https://preview.redd.it/i-tried-recreating-the-the-from-sponge-bob-v0-ua5uaaq0gilg1.jpg?width=582&format=pjpg&auto=webp&s=3a830e4421eae16ce0b4ce282c437b200c15df3b)

A statically typed language which compiles to WebAssembly

The goal of **The** language is to start with something simple to build a compiler for and expand from there. 

#### Initial version features

- [ ] Static strong typing
- [ ] Immutability by default; `mut` used to declare mutability
- [ ] User defined types and interfaces
- [ ] Generates WAT file and automatically calls `wat2wasm` on it

#### Future state features

 - [ ] Structured concurrency
 - [ ] Exhaustive pattern matching
 - [ ] Function overloading
 - [ ] Null safety using the `Maybe<Type>` which could be empty or contain a value
 - [ ] Error handling using `Try<Type, Error>` as well as built-in and definable error types
 - [ ] Container types (i.e. `Array<Type>`)
 - [ ] `enum` support
 - [ ] operator overloading for user defined types (via `operator` block)
 - [ ] Format strings
 - [ ] Import/Export system
 - [ ] Package/Dependency management system
 - [ ] Directly generate WASM code
 - [ ] LSP

 > Subject to change as development progresses

## Compiler Structure

Phases:

    1. Lexical Analysis
    2. Parsing
    3. Semantic Analysis
    4. IR Generation
    5. Code Generation (WAT)
    6. Convert WAT to WASM
    7. Execution in browser and/or CLI

> Optimization phase will come much later

## Language Specifications


### Primitive Types

| Type |
| --- |
| int |
| int64 |
| uint32 |
| uint64 |
| float |
| double |
| bool |
| char |
| String |

#### Compatibility

| Operation | Result | Valid |
| --- | --- | --- |
| `int` **operator** `int64` | `int64` | ✅ |
| `int` **operator** `float` | `float` | ✅ |
| `int` **operator** `double` | `double` | ✅ |
| `int64` **operator** `float` | **Error** | ❌ |
| `int64` **operator** `double` | `double` | ✅ |
| `float` **operator** `double` | `double` | ✅ |
| `uint32` **operator** `uint64` | `uint64` | ✅ |
| `uint32`/`uint64` **operator** any other type | **Error** | ❌ |
| `String + char` | `String` | ✅ |
| `String` **operator** any other type | **Error** | ❌ |
| `char + char` | `String` | ✅ |
| `char` **operator** any other type | **Error** | ❌ |
| `bool` **operator** any other type | **Error** | ❌ |

#### Type casting

Primitive types can be cast to each other (depending on the original and target type) using the `as` keyword. Any lossy conversions will result in a warning; types that cannot support casting from the original type will result in an error. 

All primitive types can be casted to a `String`.

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
| `uint64` | `int64` | ✅ |
| `uint32` | `float` | ⚠️ |
| `uint32` | `double` | ✅ |
| `bool` | numeric | ❌ |
| `bool` | `char` | ❌ |
| `char` | numeric | ❌ |
| `char` | `bool` | ❌ |
| any | `String` | ✅ |
| `String` | any | ❌ |

Typecasting will help with operations on incompatible types, but won't fix all incompatabilities.

#### Variables

All variables are immutable by default; to define a mutable variable, preface the definiton with `mut`.
An immutable variable must have an assigned value when defined.

```
bool isOpen = false; // Cannot be changed
mut bool isCorrect = verify(value); // Can be changed

double pi; // This will cause an error since pi is immutable but has no value
mut double e; // Since it's mutable, a value can be set later so this is valid
```

Any variables defined within a function (including parameters) are local and any properties within a `struct` are local to that.
Any variables outside of a function or struct definition are global and global variables cannot be set as mutable.
Function parameters cannot be mutable.

#### Functions

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


#### Loops

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

#### User defined types

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
            return String fileContents;
        }

        private {
            String owner;
            String group;
            fn checkPermissions() -> bool { // checkPermissions cannot be called directly outside of the struct definition
                if (this.permissions != 755) { // explicit use of this
                    return false
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
now.minute = 59;
now.setNanoSeconds(1111111); // only works the instance is mutable or the private property is marked as mutable

```


#### Interfaces

An interface is just a contract of functions that can be used. 
Any types that implement all functions within the interface, can be considered to be the same type. 
If a function accepts the `Vehicle` interface as a parameter and the `Car` type implements the interface, then the function can take a `Car`. 
All interface functions are public and must remain public during implementation.

```
interface Vehicle {
    fn move(int position, int distance) -> int;
}
```

To implement the interface:

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

A struct can implent multiple interfaces (comma separated) using the `impl` keyword in the definition.
Each interface must have a block with all of its functions implemented to be proprely defined.

An interface cannot be directly instantiated to a variable, but it can be used in the left hand side of a variable declaration only if the right hand side is a `struct` that implements the interface.

```
Device speaker; // invalid
mut Device headphones; // invalid
Device phone = Device{}; // invalid

Device speaker2 = Speaker{}; // valid
mut Device speaker3 = Speaker{}; // valid

speaker3 = NonDeviceStruct{}; // invalid since `NonDeviceStruct` does not implement `Device` so this is violating the type checking
```