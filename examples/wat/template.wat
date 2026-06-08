(module
    (import "wasi_snapshot_preview1" "proc_exit" (func $exit (param i32)))
    (memory (export "memory") 1)
    (func (export "_start") ;; wasm entry point
        ;; generated code
        (call $exit (i32.const 0)) ;; replace this 0 with the return value from main
    )
)