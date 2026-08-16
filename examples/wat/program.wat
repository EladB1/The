(module
    (import "wasi_snapshot_preview1" "proc_exit" (func $__exit (param i32)))
    (import "wasi_snapshot_preview1" "fd_write"
        (func $__print (param $fd i32) (param $iovec i32) (param $len i32) (param $written i32) (result i32))
    )
    (memory (export "memory") 1)
    ;;
    ;; compiler library
    ;;
    (global $iovec_base i32 (i32.const 0))
    (global $iovec_ptr i32 (i32.const 0))
    (global $iovec_len i32 (i32.const 4))
    (global $iovec_newline_ptr i32 (i32.const 8))
    (global $iovec_newline_len i32 (i32.const 12))
    (global $return_space i32 (i32.const 16))
    (global $NEWLINE_CHAR_ADDR i32 (i32.const 1000))
    (data (i32.const 5000) "0123456789")
    (global $itoa_out_buf i32 (i32.const 5010))
    (data (i32.const 1000) "\n")
    
    (func $__fd_write (param $fd i32) (param $ptr i32) (param $newline i32)
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
        (if (i32.eq (local.get $newline) (i32.const 1))
            (then
                (i32.store (global.get $iovec_newline_ptr) (global.get $NEWLINE_CHAR_ADDR))
                (i32.store (global.get $iovec_newline_len) (i32.const 1))
                (call $__print
                    (local.get $fd)
                    (global.get $iovec_newline_ptr)
                    (i32.const 1)
                    (global.get $return_space)
                )
                drop
            )
        )
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

    ;; (func $__itoa (param $value i32) (result i32)
    ;;     (local $is_negative i32)
    ;;     (local $digit i32)
    ;;     (local $index i32)
    ;;     (local.set $index (i32.const 0))
    ;;     (if (i32.eqz (local.get $value))
    ;;         (then
    ;;             (i32.store (global.get $itoa_out_buf) (i32.load8_u (i32.const 5000)))
    ;;             (return (global.get $itoa_out_buf))
    ;;         )
    ;;     )
    ;;     (if (i32.lt_s (local.get $value) (i32.const 0))
    ;;         (then
    ;;             (local.set $is_negative (i32.const 1))
    ;;             (local.set $value (i32.mul (i32.const -1) (local.get $value)))
    ;;             (i32.store8 (global.get $itoa_out_buf) (i32.const 45))
    ;;             (local.set $index (i32.const 1))
    ;;         )
    ;;         ;; (else (
    ;;         ;;     (local.set $is_negative (i32.const 0))
    ;;         ;; ))
    ;;     )
    ;;     (loop $digit_loop
    ;;         (local.set $digit (i32.rem_u (local.get $value) (i32.const 10)))
    ;;         (i32.sub )
    ;;     )
    ;;     (i32.const 10)
    ;;     return 

    ;; )
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
    ;; constants
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

    (data (i32.const 4000) "Assertion error")

    ;;
    ;; functions
    ;;
    (func $print (export "print") (param $ptr i32)
        (call $__fd_write (i32.const 1) (local.get $ptr) (i32.const 0))
    )
    
    (func $println (export "println") (param $ptr i32)
        (call $__fd_write (i32.const 1) (local.get $ptr) (i32.const 1))
    )

    (func $printerr (export "printerr") (param $ptr i32)
        (call $__fd_write (i32.const 2) (local.get $ptr) (i32.const 1))
    )

    (func $exit_int (export "exit_int") (param $exitCode i32)
        (call $__exit (local.get $exitCode))
    )

    (func $exit_int_String (export "exit_int_String") (param $exitCode i32) (param $message i32)
        (call $printerr (local.get $message))
        (call $__exit (local.get $exitCode))
    )

    (func $assert_bool (export "assert_bool") (param $cond i32)
        (if (i32.eqz (local.get $cond))
            (then
                (call $exit_int_String (i32.const 1) (i32.const 4000))
            )
        )
    )
    (func $assert_bool_String (export "assert_bool_String") (param $cond i32) (param $message i32)
        (if (i32.eqz (local.get $cond))
            (then
                (call $exit_int_String (i32.const 1) (local.get $message))
            )
        )
    )

    ;;
    ;; Entry point
    ;;
    (func (export "_start")
        (call $main)
        (call $exit_int)
    )

    ;;
    ;; Compiler Generated Code
    ;;
    (data $__str_const0 (i32.const 100) "")
    (data $__str_const1 (i32.const 104) "Error")
    (func $main (result i32)
        (i32.eqz (i32.const 0))
        (i32.const 104)
        (call $assert_bool_String)
        (i32.const 0)
        (return)
    )
)