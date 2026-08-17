    (data (i32.const 5000) "0123456789")
    (global $itoa_out_buf i32 (i32.const 5050))
    
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
    (func $prompt (export "prompt") (param $promptText i32) (result i32)
        (local $ptr i32)
        (call $print (local.get $promptText))
        (call $__read_stdin)
        (local.tee $ptr)
        (return)
    )

    (func $__str_fromInt32 (export "__str_fromInt32") (param $num i32) (result i32)
        (local $numtemp i32)
        (local $numlen i32)
        (local $writeidx i32)
        (local $digit i32)
        (local $dchar i32)
        (local $isnegative i32)
        (local $index i32)
        (local.set $isnegative (i32.lt_s (local.get $num) (i32.const 0)))

        (i32.eq (local.get $isnegative) (i32.const 1))
        (if
            (then
                (local.set $index (i32.const 1))
                ;; result[0] = '-'
                (i32.store8 (global.get $itoa_out_buf) (i32.const 45))
                ;; num *= -1
                (local.set $num (i32.mul (local.get $num) (i32.const -1)))
            )
        )

        ;; count number of characters and save in $numlen
        (i32.lt_s (local.get $num) (i32.const 10))
        if
            (local.set $numlen (i32.const 1))
        else
            (local.set $numlen (local.get $index))
            (local.set $numtemp (local.get $num))
            (loop $countloop (block $breakcountloop
                (i32.eqz (local.get $numtemp))
                br_if $breakcountloop

                (local.set $numtemp (i32.div_u (local.get $numtemp) (i32.const 10)))
                (local.set $numlen (i32.add (local.get $numlen) (i32.const 1)))
                br $countloop
            ))
        end
        (local.set $writeidx 
            (i32.sub 
                (i32.add (global.get $itoa_out_buf) (local.get $numlen))
                (i32.const 1)
            ))
        
        (loop $writeloop (block $breakwriteloop
            ;; digit = num % 10
            (local.set $digit (i32.rem_u (local.get $num) (i32.const 10)))
            (local.set $dchar (i32.load8_u offset=5000 (local.get $digit)))

            ;; mem[writeidx] = dchar
            (i32.store8 (local.get $writeidx) (local.get $dchar))

            ;; num /= 10
            (local.set $num (i32.div_u (local.get $num) (i32.const 10)))

            ;; if wrote first index, exit loop
            (i32.eq (local.get $writeidx) (i32.add (global.get $itoa_out_buf) (local.get $index)))
            br_if $breakwriteloop

            (local.set $writeidx (i32.sub (local.get $writeidx) (i32.const 1)))
            br $writeloop
        ))
        (global.get $itoa_out_buf)
    )

