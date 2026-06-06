# Future State

## Compiler

- Optimization stage
- Direct WASM byte code generation (skipping WAT)
- LSP
- Debugger

## Language

> **NOTE**: Any code examples in this file are just ideas. The syntax and semantics are subject to change as development progresses and ideas are fleshed out a bit more.

Structured concurrency
---

Any function marked with `async` can be called with `await` to run multiple threads. If any `async`/`await` calls fail in a function with multiple, the errors will be reported to the caller which can either handle them gracefully or return early; an early return would stop all other threads spawned by the function from running.

Exhaustive pattern matching
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

Container Types
---

Support for types which can contain other subtypes such as `Array<subtype>` or `Map<subtype1, subtype2>`

Error Handling
---

Create an error handling system where there are a set of built-in errors as well as user defined errors. Anything that could throw an error must have type `Try<Type, ErrorType>`. Both causes would have to be handled either by `match`, if statements, or a special syntax like `?` for `Type` and `?:` for empty.

Idea for how the code would look:

```

Try<Response, HTTPError> response = Try {
    resolve: TryResolve {
        condition: httpCall.status >= 200 && httpCall.status < 400,
        value: httpCall.payload
    },
    fail: TryFail {
        error: HTTPError {
            status: httpCall.status
        }
    }
};

if (response.hasFailure()) {
    printErr(response.fail);
}
else {
    println(response.resolve);
}

```

Null Safety
---

Establish a `Maybe<Type>` which states that it could either contain a value of type `Type` or could be empty (essentially null). Both causes would have to be handled either by `match`, if statements, or a special syntax like `?` for `Type` and `?:` for empty.

Idea for how the code would look:

```
Maybe<Node> next = Maybe {
    empty: MaybeEmpty {
        condition: curr.isEnd
    }
    resolve: MaybeValue {
        value: curr.getNext();
    }
};

match next {
    case Resolved: {
        println(next.value);
    }
    case Failed: {
        printErr("Attempt to access next from last node");
        exit(1);
    }
};


```

Enums
---

May support simple enums that are names and numbers or may decide to make these more complex by making it a container type

Format strings
---

Strings that can insert dynamic values to cut down on typecasting and string concatenation

Multi-line strings
---

Strings that span multiple lines. Optionally, can be made compatible with format strings as well. Would start with `/"` and end with `"\`. 

Imports/Exports
---

The ability to share code between multiple files. Thinking of something like this:

```
    importUrl("https://website.com/packageA") alias A;
    importRelative("helpers/packageB") alias B;
    importAbsolute("/usr/shared/utils/packageC") alias C;

    C.useConstants(A.constant, B.constant);

    export struct MyStruct {};

    export fn MyFunc() {};
```

Package/Dependency Management
---

Build a way to publish and consume packages similar to the way `package.json` and `go.mod` work.