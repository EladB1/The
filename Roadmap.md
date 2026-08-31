Roadmap
---

> Plans for the future of the compiler and language. \
> Nothing is guaranteed to be included and there are no dates in mind

# Future State

- [Compiler](#compiler)
    - [Optimization Phase](#optimization-phase)
    - [LSP/Code editor](#lspcode-editor)
    - [Variable cache](#semanticir-variable-cache)
    - [Package/Dependency Management](#packagedependency-management)
- [Language](#language)
    - [Structured concurrency](#structured-concurrency)
    - [Types](#types)
        - [Enums](#enums)
        - [Fixed size arrays](#fixed-size-arrays)
        - [Variable size arrays](#variable-size-arrays)
        - [Set](#set)
        - [Map](#map)
        - [User defined container types](#user-defined-container-types)
        - [File types](#file-types)
        - [Maybe (Null Safety)](#maybe-null-safety)
        - [Try (Error handling)](#try-error-handling)
    - [Standard Library functions](#standard-library-functions)
    - [Strings](#strings)
        - [Format strings](#format-strings)
        - [Multi-line strings](#multi-line-strings)
    - [Imports/Exports](#importsexports)

## Compiler

### Optimization Phase

- Constant folding
- Dead code elimination
- etc.


### LSP/Code editor

- Syntax highlighting
- Error detection
- etc.

### Semantic/IR Variable cache

For code like this:

```
if (x == 0)
    return x;
else if (x % 2 == 0)
    return x / 2;
else
    return x * x;
```

Each reference to x, triggers:

1. A call to the symbol table in the semantic analyzer
2. IR instructions to load the variable into an IR temporary

The IR for this would look like:

```
_t0: i32 = GET local.x
_t1: i32 = i32.eq _t0 i32(0)
if _t1 {
    _t2: i32 = GET local.x // repeat
    return _t2
}
else {
    _t3: i32 = GET local.x // repeat
    _t4: i32 = i32.mod _t3 i32(2)
    _t5: i32 = i32.eq _t4 i32(0)
    if _t5 {
        _t6: i32 = GET local.x // repeat
        _t7: i32 = i32.div _t6 i32(2)
        return _t7
    }
    else {
        _t8: i32 = GET local.x // repeat
        _t9: i32 = i32.mul _t8 _t8
        return _t9
    }
}

```

Want to cut out this repetition but it could make things complicated while the MVP version is still being developed

After Caching:

```
_t0: i32 = GET local.x
_t1: i32 = i32.eq _t0 i32(0)
if _t1 {
    return _t0
}
else {
    _t3: i32 = i32.mod _t0 i32(2)
    _t4: i32 = i32.eq _t3 i32(0)
    if _t4 {
        _t5: i32 = i32.div _t0 i32(2)
        return _t5
    }
    else {
        _t6: i32 = i32.mul _t0 _t0
        return _t6
    }
}

```

### Package/Dependency Management

Build a way to publish and consume packages similar to the way `package.json` and `go.mod` work.

## Language

> **NOTE**: Any code examples in this file are just ideas. The syntax and semantics are subject to change as development progresses and ideas are fleshed out a bit more.

### Structured concurrency


Any function marked with `async` can be called with `await` to run multiple threads. If any `async`/`await` calls fail in a function with multiple, the errors will be reported to the caller which can either handle them gracefully or return early; an early return would stop all other threads spawned by the function from running.

### Exhaustive pattern matching
---

Use `match` block to do pattern. Developers will be required to handle all cases, but can use `else` to avoid writing them all out.

Idea for how code will look:

```
int i = generateNum();

match i {
    case 0: {
        // some logic
    }
    case 1, 2: {
        // some logic
    }
    else: {
        // all other cases
    }
}
```

### Types

#### Enums
May support simple enums that are names and numbers or may decide to make these more complex by making it a container type

#### Fixed size arrays

```
mut int[5] arr = {1, 2, 3};
arr[3] = 10; // result: {1, 2, 3, 10}
arr[4] = 20; // result: {1, 2, 3, 10, 20}
arr[0]++; // result: {2, 2, 3, 10, 20}
println(arr.length); // result: 5
println(arr.capacity); // result: 5
arr[5]; // bounds error
```

#### Variable size arrays

```
mut Vector<int> arr = {1, 2, 3};
arr.append(10); // result: {1, 2, 3, 10}
arr.append(20); // result: {1, 2, 3, 10, 20}
arr[0]++; // result: {2, 2, 3, 10, 20}
println(arr.length); // result: 5
println(arr.capacity); // result: default size (TBD)
```

#### Set

A hashset: a vector with hash key for each unique entry

```
mut Set<int> set = {1, 2, 3};
set.append(4); // result: {1, 2, 3, 4}
set.append(1); // result: {1, 2, 3, 4}
```

#### Map


A hashmap: a key/value store

```
mut Map<String, int> values = {"name1": 10, "name2": 23};
values["name1"] += 5; // result: {"name1": 15, "name2": 23};
values["name3"] = 32; // result: {"name1": 15, "name2": 23, "name3": 32};
for (string key, int value in values) {
    println(key + ": " + value as String);
}
```

#### User defined container types

Provide some mechanism for a user to define a struct which can contain some generic value

```
struct Container<T> {
    Vector<T> values;
    container {
        fn index(int i) -> T {
            return values[i];
        }
        fn append(T value) {
            values.append(value);
        }
    }
}

Container<String> c = Container<String>{};
c.append("The");
println(c[0]);
```

#### File types

Provided some native file system types for I/O.

```
File file = Open("somefile.txt");
println(file.name); // result: "somefile.txt"
println(file.permissions); // result: 0644

FilePermissions perms = FilePermissions{
    User: {'r', 'w', 'x'},
    Group: '{'r'},
    Public: {}
};

Directory dir = CreateDir("/tmp/results", perms);
```


#### Maybe (Null Safety)

Establish a `Maybe<Type>` which states that it could either contain a value of type `Type` or could be empty. Both cases would have to be handled either by `match`, if statements, or a special syntax like `?` for `Type` and `?:` for empty. The "zero value" of a `Maybe` is empty.

Idea for how the code would look:

```
struct Node {
    int value;
    Maybe<Node> next;
}
fn addNode(Maybe<Node> head, Node node) {
    if (head.empty)
        return node;
    Node curr = head.resolved;
    while (!curr.next.empty) {
        curr = curr.next.resolved;
    }
    curr.next.Resolve(node); // set the Maybe to node
}

fn deleteEverythingAfterNode(Node head) {
    head.next.Empty(); // set the Maybe to empty
}
```

> **Under the hood:**\
> `Maybe<T>` is a pointer to a value of type `T`\
> An empty value means the pointer is null\
> Resolved is used to safely dereference the pointer

#### Try (Error Handling)

Create an error handling system where there are a set of built-in errors as well as user defined errors. Anything that could throw an error must have type `Try<Type, ErrorType>`. Both causes would have to be handled either by `match`, if statements, or a special syntax like `?` for `Type` and `?:` for error. Essentially, `Try<Type, ErrorType>` is 

There are 

Idea for how the code would look:

```

fn getFile(String path) -> Try<File, IOError> {
    if (!exists(path))
        raise FileNotFoundError("Could not find path " + path); // indicate the failure case using `raise` keyword
    return open(path); // indicate the success with the return of type File
}

fn main() -> int {
    String lines = getFile("/etc/hosts").fold(
        resolved: file => file.read(),
        failed: error => {
            exit(1, err);
            // can return a String from here to handle the error gracefully
        }
    );
    return 0;
}

```

> **Under the hood:**\
> `Try<T, E>` is a tagged union with an internal value flag\
> If resolved, the value flag is set to 0; otherwise, set it to 1\
> `raise` sets the internal flag to 1, indicating that the error type should be used\
> `fold` resolves down to conditional blocks that run the `resolved` and `failed` code depending on the tagged union flag

### Standard Library Functions

| Signature(s) | Description |
| --- | --- |
| `typeOf(Any value) -> String` | Get the type of a value as a string |
| `sizeOf(Any value) -> uint32` | Get the size of a value in bytes | 
| `secretPrompt(String promptText) -> String` | Print `promptText` and read from stdin but hide characters being typed |
| `getEnv(String key) -> String` | Get the value of environment variable |
| `setEnv(String key, String value)` | Set the value of environment variable |
| `readEnv() -> Map<String, String>` | Get all environment variables |


### Strings

#### Format strings

Strings that can insert dynamic values to cut down on typecasting and string concatenation

```
int i = 0;
String s = `i is {i}`;
println(`\{ {i+5} \});
```

Internally, this would be treated as:

```
String s = "i is " + (i as String);
println("{" + ((i+5) as string) + "}");
```

EBNF rule:

```ebnf
FormatStringLiteral = "`" .* { "{" expression "}" } .* "`" ;
```

#### Multi-line strings

Strings that span multiple lines. Optionally, can be made compatible with format strings as well. Would start with `/"` and end with `"\`. 

### Imports/Exports

The ability to share code between multiple files. Thinking of something like this:

```
    package somename; 

    import something, otherthing from package;
    import thirdthing from "../data/types.the";
    import otherpackage;

    export int constant = 1001 + something;
    export fn func() {}
    export interface Template {}
    export struct Structure {}
    export {
        String key = "" + otherpackage.exportedValue;
        fn otherFunc() {}
        interface OtherInterface {}
        struct OtherStruct {}
    }
```

