(module
    (import "wasi_snapshot_preview1" "proc_exit" (func $__exit (param i32)))
    (import "wasi_snapshot_preview1" "fd_write"
        (func $__print (param $fd i32) (param $iovec i32) (param $len i32) (param $written i32) (result i32))
    )
    (memory (export "memory") 1)
    (global $iovec_base i32 (i32.const 0))
    (global $iovec_ptr i32 (i32.const 0))
    (global $iovec_len i32 (i32.const 4))
    (global $return_space i32 (i32.const 8))
    ;;
    ;; compiler library
    ;;
    (func $__fd_write (param $fd i32) (param $ptr i32)
        (local $len i32)
        (call $__str_length (local.get $ptr))
        (local.set $len)
        ;; Update reusable iovec
        (i32.store (global.get $iovec_ptr) (local.get $ptr))
        (i32.store (global.get $iovec_len) (local.get $len))
        (call $__print
            (local.get $fd)
            (global.get $iovec_base)
            (i32.const 1)
            (global.get $return_space)

        )
        drop
    )

    (func $__str_length (param $ptr i32) (result i32)
        (local $len i32)
        (local $curr i32)

        (local.set $len (i32.const 0))

        (loop $length_loop
            ;; Use pointer arithmetic to load a byte from memory
            (i32.load8_u (i32.add (local.get $ptr) (local.get $len)))
            (local.set $curr)
            ;; check for null terminator
            (if (i32.eqz (local.get $curr))
                (then
                    (return (local.get $len))
                )
            )
            (local.set $len (i32.add (local.get $len) (i32.const 1)))
            (br $length_loop)
        )
        (local.get $len)
    )
    ;; (func $__i32_pow (param $base i32) (param $exponent i32) (result i32)
    ;;     ;; TODO
    ;; )

    ;; (func $__i64_pow (param $base i64) (param $exponent i64) (result i64)
    ;;     ;; TODO
    ;; )

    ;; (func $__u32_pow (param $base i32) (param $exponent i32) (result i32)
    ;;     ;; TODO
    ;; )

    ;; (func $__u64_pow (param $base i64) (param $exponent i64) (result i64)
    ;;     ;; TODO
    ;; )

    ;; (func $__f32_pow (param $base f32) (param $exponent f32) (result f32)
    ;;     ;; TODO
    ;; )

    ;; (func $__f64_pow (param $base f64) (param $exponent f64) (result f64)
    ;;     ;; TODO
    ;; )

    ;;
    ;; built-ins
    ;;
    (global $INT_MIN i32 (i32.const -2147483648))
    (global $INT_MAX i32 (i32.const 2147483647))

    (global $INT64_MIN i64 (i64.const -9223372036854775808))
    (global $INT64_MAX i64 (i64.const 9223372036854775807))

    (global $UINT32_MAX i32 (i32.const 4294967295))
    (global $UINT64_MAX i64 (i64.const 18446744073709551615))

    (global $FLOAT_MIN f32 (f32.const -0x1.ff933c78cdfadp+127))
    (global $FLOAT_MIN_POSITIVE f32 (f32.const 0x1.00fb32c6204c4p-126))
    (global $FLOAT_MAX f32 (f32.const 0x1.ff933c78cdfadp+127))
    (global $FLOAT_EPSILON f32 (f32.const 0x1.ff19e23a836fcp-24))
    (global $FLOAT_NaN f32 (f32.const 0x7FC00000))
    (global $FLOAT_INF f32 (f32.const 0x7F800000))
    (global $FLOAT_NEG_INF f32 (f32.const 0xFF800000))

    (global $DOUBLE_MIN f64 (f64.const -0x1.fdcf158adbb99p+1023))
    (global $DOUBLE_MIN_POSITIVE f64 (f64.const 0x1.0091177587f82p-1022))
    (global $DOUBLE_MAX f64 (f64.const 0x1.fdcf158adbb99p+1023))
    (global $DOUBLE_EPSILON f64 (f64.const 0x1.ffe5ab7e8ad5ep-53))
    (global $DOUBLE_NaN f64 (f64.const 0x7FF8000000000000))
    (global $DOUBLE_INF f64 (f64.const 0x7FF0000000000000))
    (global $DOUBLE_NEG_INF f64 (f64.const 0xFFF0000000000000))

    (global $PI f64 (f64.const 3.141592653589793))
    (global $E f64 (f64.const 2.718281828459045))

    (func $print (export "print") (param $ptr i32)
        (call $__fd_write (i32.const 1) (local.get $ptr))
    )
    ;; TODO: println

    (func $printerr (export "printerr") (param $ptr i32)
        (call $__fd_write (i32.const 2) (local.get $ptr))
    )

    (func $exit_int (export "exit_int") (param $exitCode i32)
        (call $__exit (local.get $exitCode))

    )

    (func $exit_int_String (export "exit_int_String") (param $exitCode i32) (param $message i32)
        (call $printerr (local.get $message))
        (call $__exit (local.get $exitCode))
    )

    ;;
    ;; entry point 
    ;;

    (func (export "_start")
        (call $main)
        (call $exit_int)
    )

    ;;
    ;; Generated code
    ;;

    (data (i32.const 100) "hello, world!\n")
    
    ;; will eventually get rid of this but using it for POC
    (func $main (result i32)
        ;;(call $__str_length (i32.const 50))
        
        (call $printerr
            (i32.const 100)
        )
        (i32.const 0)
    )
)