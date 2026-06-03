# The

![Language Icon](https://preview.redd.it/i-tried-recreating-the-the-from-sponge-bob-v0-ua5uaaq0gilg1.jpg?width=582&format=pjpg&auto=webp&s=3a830e4421eae16ce0b4ce282c437b200c15df3b)

A statically typed language which compiles to WebAssembly

The goal of **The** language is to start with something simple to build a compiler for and expand from there. 

#### Initial version features

- [ ] Static strong typing
- [ ] Immutability by default; `mut` used to declare mutability
- [ ] User defined types and interfaces
- Generates WAT file and automatically calls `wat2wasm` on it

#### Future state features

 - [ ] Structured concurrency
 - [ ] Exhaustive pattern matching
 - [ ] Function overloading
 - [ ] Null safety using the `Maybe<Type>` which could be empty or contain a value
 - [ ] Error handling using `Try<Type, Error>` as well as built-in and definable error types
 - [ ] Container types (i.e. `Array<Type>`)
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


#### Basic types

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
| enum |

#### Variables

All variables are immutable by default; to define a mutable variable, preference the definiton with `mut`

```
bool isOpen = false; // Cannot be changed
mut bool isCorrect = verify(value); // Can be changed
```

An immutable variable must have an assigned value when defined

Any variables defined within a function (including parameters) are local and any properties within a `struct` are local to that instance

Any variables outside of a function are global and global variables cannot be set as mutable

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

The `divisibleByTwo` function takes an int parameter and returns a bool value


#### Loops

Loops come in different shapes depending on the use case

There are two types of loops: `while` and `for`

While loops are like most other languages

```
while (condition) {
    // code goes here
}
```

Since variables are immutable by default, there are restrictions on using C-style for loops

```
for (int i = 0; i < limit; i++) {
    // code goes here
} // this will fail since i is immutable but the loop is trying to increment it
```

For iterating a set of int values, this syntax can be used instead

```
for (int i in 0 .. limit) {
    // code goes here
}
```

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

the loop above will start at limit and stop when it's greater than or equal to 0, it will iterate by -2 each iteration

So far, we've only handled addition... what if you want multiplication as the iteration part of the loop?

For that, you would have to use a C-style loop with a mutable iterator (can also be done for any of the above cases)

```
    for (mut int i = 1; i < limit; i *= 2) { // i is mutable to this should work
    // code goes here
}
```

Looping over strings works similar to the for loops covered already

```
for (char c in "Hello, world!") {
    // code goes here
}
```

If you want to include the index of the character, you can change the loop variables to support that

```
for (int i, char c in "Hello, world!") {
    // code goes here
}
```

The compiler will automatically handle both "shapes" of loops so either case is simple

> C-style mutable loops can be used as well but indexing a string has not been planned yet

Flow control keywords `break` to stop the loop and `continue` to skip to the next iteration are only valid in loops

#### User defined types

A user can define a type which includes its own data and methods

```
struct Employee {
    uint32 id;
    String name;
    float salary;
}
Employee emp = Employ {
    id: 12,
    name: "CEO",
    salary: 0.01,
}; // can include or omit trailing comma
println(emp.id)
```

In the definition of a `struct`, all properties must have a type and end with a semi-colon

By default, everything is public, but you can use `private` or create a `private` block to prevent outside access to properties

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

The `struct` functions allow you to use the syntactic convention of `instanceName.function()` rather than having to pass a type as a parameter

```
File file = File {
    //...
};
String doc = file.read();
```

A mutable instance of a `struct` means that any of its public properties are mutable as well; private properties can be updated via methods but not directly. For example:

```
struct Time {
    uint32 hour;
    uint32 minute;
    uint32 second;
    private uint64 nanoseconds;
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

An interface is just a contract of functions that can be used

```
interface Vehicle {
    fn move(int position, int distance) -> int;
}
```

Any types that implement all functions within the interface, can be considered to be the same type

If a function accepts the `Vehicle` interface as a parameter and the `Car` type implements the interface, then the function can take a `Car`

All interface functions are public and must remain public during implementation

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

A struct can implent multiple interfaces (comma separated) using the `impl` keyword in the definition

Each interface must have a block with all of its functions implemented to be proprely defined