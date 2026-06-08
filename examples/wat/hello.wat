(module
    (import "wasi_snapshot_preview1" "proc_exit" (func $exit (param i32)))
    (import "wasi_snapshot_preview1" "fd_write"
        (func $print (param $fd i32) (param $iovec i32) (param $len i32) (param $written i32) (result i32))
    )
    (memory (export "memory") 1)
    (data (i32.const 0) "Hello, World!\n")
    (func (export "_start") ;; wasm entry point
        (local $iovs i32)
        
        (i32.store (i32.const 16) (i32.const 0)) ;; start of memory to write to (store at 16th byte)
        (i32.store (i32.const 20) (i32.const 14)) ;; number of bytes of memory to write (stored at 20th byte)

        (local.set $iovs (i32.const 16)) ;; set value 16 for starting position for reading memory
        (call $print
            (i32.const 1) ;; file descriptor: 1 for stdout, 2 for stderr
            (local.get $iovs) ;; starting position for memory read
            (i32.const 1) ;; number of times to read memory (incremented by 4 bytes from starting position)
            (i32.const 24) ;; memory index to store the number of bytes written to fd
        )
        drop ;; Will get an error since the return of $print stays on the stack
        (call $exit (i32.const 0)) ;; replace this 0 with the return value from main
    )
)

;; Execute this with wasmtime examples/wat/hello.wat