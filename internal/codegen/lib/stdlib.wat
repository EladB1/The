    (global $__bool_true i32 (i32.const 4000))
    (data (i32.const 4000) "\04\00\00\00true")
    (global $__bool_false i32 (i32.const 4500))
    (data (i32.const 4500) "\05\00\00\00false")
    (data (i32.const 51) "0123456789")
    (global $bounds_error i32 (i32.const 65))
    (global $__default_assertion_error i32 (i32.const 80))
    (data (i32.const 80) "\FF\00\00\00Assertion error")
    (global $error_prefix i32 (i32.const 9000))
    (data (i32.const 9000) "\19\00\00\00\1b[1;31mRuntimeError:\1b[0m ")
    (global $slice_error1 i32 (i32.const 5000))
    (data (i32.const 5000) "\0C\00\00\00slice start ")
    (global $slice_error2 i32 (i32.const 6000))
    (data (i32.const 6000) "\22\00\00\00 cannot be greater than slice end ")
    (global $bounds_error1 i32 (i32.const 7000))
    (data (i32.const 7000) "\06\00\00\00index ")
    (global $bounds_error2 i32 (i32.const 8000))
    (data (i32.const 8000) "\0E\00\00\00 out of range ")

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
                (call $exit_int_String (i32.const 1) (global.get $__default_assertion_error))
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
        (local $addr i32)
        (local $start i32)
        (local.set $isnegative (i32.lt_s (local.get $num) (i32.const 0)))
        (i32.eq (local.get $isnegative) (i32.const 1))
        (if
            (then
                (local.set $index (i32.const 1))
                ;; num *= -1
                (local.set $num (i32.mul (local.get $num) (i32.const -1)))
            )
        )
        (local.set $numlen (local.get $index))
        ;; count number of characters and save in $numlen
        (i32.lt_s (local.get $num) (i32.const 10))
        if
            (local.set $numlen (i32.add (i32.const 1) (local.get $index)))
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
        (local.set $addr (call $__malloc (i32.add (local.get $numlen) (i32.const 4))))
        (local.set $start (i32.add (local.get $addr) (i32.const 4)))
        (i32.store offset=0 (local.get $addr) (local.get $numlen))
        (if (i32.eq (local.get $isnegative) (i32.const 1))
            (then ;; result[0] = '-'
                (i32.store8 (local.get $start) (i32.const 45))
            )
        )
        (local.set $writeidx 
            (i32.sub 
                (i32.add (local.get $start) (local.get $numlen))
                (i32.const 1)
            ))
        
        (loop $writeloop (block $breakwriteloop
            ;; digit = num % 10
            (local.set $digit (i32.rem_u (local.get $num) (i32.const 10)))
            (local.set $dchar (i32.load8_u offset=51 (local.get $digit)))

            ;; mem[writeidx] = dchar
            (i32.store8 (local.get $writeidx) (local.get $dchar))

            ;; num /= 10
            (local.set $num (i32.div_u (local.get $num) (i32.const 10)))

            ;; if wrote first index, exit loop
            (i32.eq (local.get $writeidx) (i32.add (local.get $start) (local.get $index)))
            br_if $breakwriteloop

            (local.set $writeidx (i32.sub (local.get $writeidx) (i32.const 1)))
            br $writeloop
        ))
        (local.get $addr)
    )

    (func $__str_fromInt64 (export "__str_fromInt64") (param $num i64) (result i32)
        (local $numtemp i64)
        (local $numlen i32)
        (local $writeidx i32)
        (local $digit i32)
        (local $dchar i32)
        (local $isnegative i32)
        (local $index i32)
        (local $addr i32)
        (local $start i32)
        (local.set $isnegative (i64.lt_s (local.get $num) (i64.const 0)))

        (i32.eq (local.get $isnegative) (i32.const 1))
        (if
            (then
                (local.set $index (i32.const 1))
                ;; num *= -1
                (local.set $num (i64.mul (local.get $num) (i64.const -1)))
            )
        )
        (local.set $numlen (local.get $index))
        ;; count number of characters and save in $numlen
        (i64.lt_s (local.get $num) (i64.const 10))
        if
            (local.set $numlen (i32.add (i32.const 1) (local.get $index)))
        else
            (local.set $numlen (local.get $index))
            (local.set $numtemp (local.get $num))
            (loop $countloop (block $breakcountloop
                (i64.eqz (local.get $numtemp))
                br_if $breakcountloop

                (local.set $numtemp (i64.div_u (local.get $numtemp) (i64.const 10)))
                (local.set $numlen (i32.add (local.get $numlen) (i32.const 1)))
                br $countloop
            ))
        end
        (local.set $addr (call $__malloc (i32.add (local.get $numlen) (i32.const 4))))
        (local.set $start (i32.add (local.get $addr) (i32.const 4)))
        (i32.store offset=0 (local.get $addr) (local.get $numlen))
        (if (i32.eq (local.get $isnegative) (i32.const 1))
            (then ;; result[0] = '-'
                (i32.store8 (local.get $start) (i32.const 45))
            )
        )
        (local.set $writeidx 
            (i32.sub 
                (i32.add (local.get $start) (local.get $numlen))
                (i32.const 1)
            )
        )
        
        (loop $writeloop (block $breakwriteloop
            ;; digit = num % 10
            (i64.rem_u (local.get $num) (i64.const 10))
            (i32.wrap_i64)
            (local.set $digit)
            (local.set $dchar (i32.load8_u offset=51 (local.get $digit)))

            ;; mem[writeidx] = dchar
            (i32.store8 (local.get $writeidx) (local.get $dchar))

            ;; num /= 10
            (local.set $num (i64.div_u (local.get $num) (i64.const 10)))

            ;; if wrote first index, exit loop
            (i32.eq (local.get $writeidx) (i32.add (local.get $start) (local.get $index)))
            br_if $breakwriteloop

            (local.set $writeidx (i32.sub (local.get $writeidx) (i32.const 1)))
            br $writeloop
        ))
        (local.get $addr)
    )

    (func $__str_fromBool (export "__str_fromBool") (param $value i32) (result i32)
        (if (result i32) 
            (i32.eqz (local.get $value))
            (then (global.get $__bool_false))
            (else (global.get $__bool_true))
        )
    )

    (func $__str_fromChar (export "__str_fromChar") (param $value i32) (result i32)
        (local $addr i32)
        (call $__malloc (i32.const 5))
        (local.set $addr)
        (i32.store8 offset=0 (local.get $addr) (i32.const 1))
        (i32.store8 offset=4 (local.get $addr) (local.get $value))
        ;;(i32.store8 (i32.add (local.get $addr) (i32.const 1)) (i32.const 0)) ;; null terminator
        (local.get $addr)
    )

    (func $__str_indexOf (export "__str_indexOf") (param $char i32) (param $ptr i32) (result i32)
       (local $i i32)
       (local $curr i32)
       (local $start i32)
       (local $len i32)
       (local.set $start (i32.add (local.get $ptr) (i32.const 4)))
       (local.set $len (call $__str_length (local.get $ptr)))
       (loop $str_loop (block $exit__str_loop
        (if (i32.eq (local.get $len) (local.get $i))
            (then
                (br $exit__str_loop)
            )
        )
        (i32.load8_u (i32.add (local.get $start) (local.get $i)))
        (local.set $curr)
        (if (i32.eq (local.get $curr) (local.get $char))
            (then
                (return (local.get $i))
            )
        )
        (local.set $i (i32.add (local.get $i) (i32.const 1)))
        (br $str_loop)
       ))
       (i32.const -1)
       (return)
    )

    (func $__str_contains_char (export "__str_contains_char") (param $char i32) (param $ptr i32) (result i32)
        (local $index i32)
        (call $__str_indexOf (local.get $char) (local.get $ptr))
        (i32.ne (i32.const -1))
        (return)
    )

    (func $__newBoundsError (param $index i32) (param $len i32)
        (local $error i32)
        (local $indexStr i32)
        (local $lenStr i32)
        (call $__str_fromInt32 (local.get $index))
        (local.set $indexStr)
        (call $__str_fromInt32 (local.get $len))
        (local.set $lenStr)
        (call $__malloc (i32.const 16))
        (local.set $error)
        (i32.store offset=0 (local.get $error) (global.get $bounds_error1))
        (i32.store offset=4 (local.get $error) (local.get $indexStr))
        (i32.store offset=8 (local.get $error) (global.get $bounds_error2))
        (i32.store offset=12 (local.get $error) (local.get $lenStr))
        (call $__panic (local.get $error) (i32.const 4))
    )

    (func $__str_index (export "__str_index") (param $str i32) (param $index i32) (result i32)
        (local $len i32)
        (call $__str_length (local.get $str))
        (local.set $len)
        (i32.or (i32.lt_s (local.get $index) (i32.const 0)) (i32.ge_s (local.get $index) (local.get $len)))
        if
            (call $__newBoundsError (local.get $index) (local.get $len))
        end
        (i32.load8_u offset=4 (i32.add (local.get $str) (local.get $index)))
    )

    (func $__str_slice (export "__str_slice") (param $str i32) (param $start i32) (param $end i32) (result i32)
        (local $len i32)
        (local $slice i32)
        (local $sliceLen i32)
        (local $strStart i32)
        (local $sliceStart i32)
        (local $error i32)
        (local $startStr i32)
        (local $endStr i32)
        (local.set $len (call $__str_length (local.get $str)))
        (if (i32.lt_s (local.get $start) (i32.const 0))
            (then
                (call $__newBoundsError (local.get $start) (local.get $len))
            )
            (else
                (i32.ge_s (local.get $end) (local.get $len))
                if
                    (call $__newBoundsError (local.get $end) (local.get $len))
                end
            )
        )
        (i32.gt_s (local.get $start) (local.get $end))
        if 
            (call $__str_fromInt32 (local.get $start))
            (local.set $startStr)
            (call $__str_fromInt32 (local.get $end))
            (local.set $endStr)
            (call $__malloc (i32.const 16))
            (local.set $error)
            (i32.store offset=0 (local.get $error) (global.get $slice_error1))
            (i32.store offset=4 (local.get $error) (local.get $startStr))
            (i32.store offset=8 (local.get $error) (global.get $slice_error2))
            (i32.store offset=12 (local.get $error) (local.get $endStr))
            (call $__panic (local.get $error) (i32.const 4))
        end
        (i32.sub (local.get $end) (local.get $start))
        (local.set $sliceLen)
        (call $__malloc (i32.add (local.get $sliceLen) (i32.const 4)))
        (local.set $slice)
        (i32.store (local.get $slice) (local.get $sliceLen))
        (local.set $sliceStart (i32.add (local.get $slice) (i32.const 4)))
        (local.set $strStart (i32.add (local.get $str) (i32.const 4)))
        (memory.copy
            (local.get $sliceStart)
            (i32.add (local.get $strStart) (local.get $start))
            (local.get $sliceLen)
        )
        (local.get $slice)
    )

    (func $__str_slice_inclusive (export "__str_slice_inclusive") (param $str i32) (param $start i32) (param $end i32) (result i32)
        (local $len i32)
        (local $slice i32)
        (local $sliceLen i32)
        (local $strStart i32)
        (local $sliceStart i32)
        (local $error i32)
        (local $startStr i32)
        (local $endStr i32)
        (local.set $len (call $__str_length (local.get $str)))
        (if (i32.lt_s (local.get $start) (i32.const 0))
            (then
                (call $__newBoundsError (local.get $start) (local.get $len))
            )
            (else
                (i32.ge_s (local.get $end) (local.get $len))
                if
                    (call $__newBoundsError (local.get $end) (local.get $len))
                end
            )
        )
        (i32.gt_s (local.get $start) (local.get $end))
        if 
            (call $__str_fromInt32 (local.get $start))
            (local.set $startStr)
            (call $__str_fromInt32 (local.get $end))
            (local.set $endStr)
            (call $__malloc (i32.const 16))
            (local.set $error)
            (i32.store offset=0 (local.get $error) (global.get $slice_error1))
            (i32.store offset=4 (local.get $error) (local.get $startStr))
            (i32.store offset=8 (local.get $error) (global.get $slice_error2))
            (i32.store offset=12 (local.get $error) (local.get $endStr))
            (call $__panic (local.get $error) (i32.const 4))
        end
        (i32.sub (i32.add (local.get $end) (i32.const 1)) (local.get $start))
        (local.set $sliceLen)
        (call $__malloc (i32.add (local.get $sliceLen) (i32.const 4)))
        (local.set $slice)
        (i32.store (local.get $slice) (local.get $sliceLen))
        (local.set $sliceStart (i32.add (local.get $slice) (i32.const 4)))
        (local.set $strStart (i32.add (local.get $str) (i32.const 4)))
        (memory.copy
            (local.get $sliceStart)
            (i32.add (local.get $strStart) (local.get $start))
            (local.get $sliceLen)
        )
        (local.get $slice)
    )

    (func $__char_concat (export "__char_concat") (param $char1 i32) (param $char2 i32) (result i32)
        (local $string i32)
        (call $__malloc (i32.const 6))
        (local.set $string)
        (i32.store (local.get $string) (i32.const 2))
        (i32.store8 (i32.add (local.get $string) (i32.const 4)) (local.get $char1))
        (i32.store8 (i32.add (local.get $string) (i32.const 5)) (local.get $char2))
        (local.get $string)
    )

    (func $__char_concat_str (export "__char_concat_str") (param $char i32) (param $str i32) (result i32)
        (local $string i32)
        (local $start i32)
        (local $len i32)
        (call $__str_length (local.get $str))
        (local.set $len)
        (call $__malloc (i32.add (local.get $len) (i32.const 5)))
        (local.set $string)
        (i32.store (local.get $string) (i32.add (local.get $len) (i32.const 1)))
        (local.set $start (i32.add (local.get $string) (i32.const 4)))
        (i32.store8 (local.get $start) (local.get $char))
        (memory.copy 
            (i32.add (local.get $start) (i32.const 1))
            (i32.add (local.get $str) (i32.const 4))
            (local.get $len)
        )        
        (local.get $string)
    )

    (func $__str_concat (export "__str_concat") (param $str1 i32) (param $str2 i32) (result i32)
        (local $string i32)
        (local $len1 i32)
        (local $len2 i32)
        (local $total_len i32)
        (local $start i32)
        (call $__str_length (local.get $str1))
        (local.set $len1)
        (call $__str_length (local.get $str2))
        (local.set $len2)
        (local.set $total_len (i32.add (local.get $len1) (local.get $len2)))
        (call $__malloc (i32.add (local.get $total_len) (i32.const 4)))
        (local.set $string)
        (i32.store (local.get $string) (local.get $total_len))
        (local.set $start (i32.add (local.get $string) (i32.const 4)))
        (memory.copy 
            (local.get $start)
            (i32.add (local.get $str1) (i32.const 4))
            (local.get $len1)
        )
        (memory.copy 
            (i32.add (local.get $start) (local.get $len1))
            (i32.add (local.get $str2) (i32.const 4))
            (local.get $len2)
        )    
        (local.get $string)
    )

    (func $__str_concat_char (export "__str_concat_char") (param $str i32) (param $char i32) (result i32)
        (local $string i32)
        (local $len i32)
        (local $start i32)
        (call $__str_length (local.get $str))
        (local.set $len)
        (call $__malloc (i32.add (local.get $len) (i32.const 5))) ;; length + 1 (char) + 4 (length prefix for strings)
        (local.set $string)
        (i32.store (local.get $string) (i32.add (local.get $len) (i32.const 1)))
        (local.set $start (i32.add (local.get $string) (i32.const 4)))
        (memory.copy 
            (local.get $start)
            (i32.add (local.get $str) (i32.const 4))
            (local.get $len)
        )
        (i32.store8 (i32.add (local.get $start) (local.get $len)) (local.get $char))
        (local.get $string)
    )