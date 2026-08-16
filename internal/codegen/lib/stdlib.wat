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

