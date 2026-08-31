(module $the
    (import "wasi_snapshot_preview1" "proc_exit" (func $__exit (param i32)))
    (import "wasi_snapshot_preview1" "fd_write"
        (func $__print (param $fd i32) (param $iovec i32) (param $len i32) (param $written i32) (result i32))
    )
    (import "wasi_snapshot_preview1" "fd_read"
        (func $__fd_read (param $fd i32) (param $iovec i32) (param $len i32) (param $written i32) (result i32))
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
    (global $NEWLINE_CHAR_ADDR i32 (i32.const 20))
    (data (i32.const 20) "\n")
    (global $stdin_buffer i32 (i32.const 24))
    (data (i32.const 24) "\00")
    (global $iovec_stdin i32 (i32.const 28))
    (global $iovec_stdin_len i32 (i32.const 32))
    (data (i32.const 32) "\00\00\00\00\00\00\00\00")
    ;; byte counter
    (global $stdin_byte_counter i32 (i32.const 36))
    (data (i32.const 36) "\00\00\00\00")

    ;; (tag $bounds_error (param i32))
    ;; (export "bounds_error_tag" (tag $bounds_error))
    (global $error_prefix i32 (i32.const 9000))
    (data (i32.const 9000) "\00\01\00\00\1b[1;31mRuntimeError:\1b[0m ")
    (global $bounds_error1 i32 (i32.const 7000))
    (data (i32.const 7000) "\06\00\00\00index ")
    (global $bounds_error2 i32 (i32.const 8000))
    (data (i32.const 8000) "\FC\00\00\00 out of range ")
    (global $slice_error1 i32 (i32.const 5000))
    (data (i32.const 5000) "\0C\00\00\00slice start ")
    (global $slice_error2 i32 (i32.const 6000))
    (data (i32.const 6000) "\22\00\00\00 cannot be greater than slice end ")

    (global $malloc_start i32 (i32.const 1024))
    (global $malloc_next (mut i32) (i32.const 1024))

    (func $__malloc (param $size i32) (result (;memory address;) i32)
        (local $curr_alloc_addr i32)
        (local $next_alloc_addr i32)

        (global.get $malloc_next)
        (local.set $curr_alloc_addr)

        (i32.add (local.get $curr_alloc_addr) (local.get $size))
        (local.set $next_alloc_addr)

        (global.set $malloc_next (local.get $next_alloc_addr))

        (call $__align)

        (local.get $curr_alloc_addr)
    )

    (func $__align
        (i32.eqz (i32.rem_u (global.get $malloc_next) (i32.const 4)))
        if
            (return)
        else
            (i32.mul
                (i32.const 4)
                (i32.add
                    (i32.const 1)
                    (i32.div_u (global.get $malloc_next) (i32.const 4))
                )
            )
            (global.set $malloc_next)
        end
    )
    
    (func $__fd_write (param $fd i32) (param $ptr i32) (param $newline i32)
        ;; Update reusable iovec
        (i32.store (global.get $iovec_ptr) (i32.add (local.get $ptr) (i32.const 4)))
        (i32.store (global.get $iovec_len) (call $__str_length (local.get $ptr)))
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

    (func $__panic (param $list i32) (param $len i32)
        (local $msg i32)
        (local $index i32)
        (local $tmp i32)
        (local.set $index (i32.const 0))
        (global.get $error_prefix)
        (local.set $msg)
        (loop $concat_loop (block $exit_concat_loop
            (i32.eq (local.get $index) (local.get $len))
            br_if $exit_concat_loop
            (i32.load
                (i32.add 
                    (local.get $list)
                    (i32.mul (local.get $index) (i32.const 4))
                )
            )
            (local.set $tmp)
            (call $__str_concat (local.get $msg) (local.get $tmp))
            (local.set $msg)
            (local.set $index (i32.add (local.get $index) (i32.const 1)))
            
            br $concat_loop
        ))
        (call $exit_int_String (i32.const 1) (local.get $msg))
    )

    (func $__read_stdin (result i32)
        (local $result i32)
        (local $len i32)
        (i32.store (global.get $iovec_stdin) (global.get $stdin_buffer))
        (i32.store (global.get $iovec_stdin_len) (i32.const 256)) ;; allow 256 bytes
        (call $__fd_read
            (i32.const 0) ;; stdin
            (global.get $iovec_stdin)
            (i32.const 1)
            (global.get $stdin_byte_counter)
        )
        drop
        (call $__malloc (i32.add (global.get $stdin_byte_counter) (i32.const 4)))
        (local.set $result)
        (i32.sub (i32.load (global.get $stdin_byte_counter)) (i32.const 1)) ;; remove EOL character
        (local.set $len)
        (i32.store (local.get $result) (local.get $len))
        (memory.copy
            (i32.add (local.get $result) (i32.const 4))
            (global.get $stdin_buffer)
            (local.get $len)
        )
        (local.get $result)
    )

    (func $__str_length (param $ptr i32) (result i32)
        (i32.load offset=0 (local.get $ptr))
    )

    (func $__i32_pow (param $base i32) (param $expo i32) (result i32)
        (local $result i32)
        (local $i i32)
        (i32.eqz (local.get $expo))
        if 
            (i32.const 1)
            (return)
        else
            (i32.lt_s (local.get $expo) (i32.const 0))
            if 
                (i32.const 0)
                (return)
            end
        end
        (local.set $i (i32.const 0))
        (local.set $result (i32.const 1))
        (loop $pow_loop (block $exit_pow_loop
            (local.set $result (i32.mul (local.get $result) (local.get $base)))
            (local.set $i (i32.add (local.get $i) (i32.const 1)))
            (i32.eq (local.get $i) (local.get $expo))
            br_if $exit_pow_loop
            br $pow_loop
        )
        )
        (local.get $result)
    )

    (func $__i64_pow (param $base i64) (param $expo i64) (result i64)
        (local $result i64)
        (local $i i64)
        (i64.eqz (local.get $expo))
        if 
            (i64.const 1)
            (return)
        else
            (i64.lt_s (local.get $expo) (i64.const 0))
            if 
                (i64.const 0)
                (return)
            end
        end
        (local.set $i (i64.const 0))
        (local.set $result (i64.const 1))
        (loop $pow_loop (block $exit_pow_loop
            (local.set $result (i64.mul (local.get $result) (local.get $base)))
            (local.set $i (i64.add (local.get $i) (i64.const 1)))
            (i64.eq (local.get $i) (local.get $expo))
            br_if $exit_pow_loop
            br $pow_loop
        )
        )
        (local.get $result)
    )

    (func $__u32_pow (param $base i32) (param $expo i32) (result i32)
        (local $result i32)
        (local $i i32)
        (i32.eqz (local.get $expo))
        if 
            (i32.const 1)
            (return)
        end
        (local.set $i (i32.const 0))
        (local.set $result (i32.const 1))
        (loop $pow_loop (block $exit_pow_loop
            (local.set $result (i32.mul (local.get $result) (local.get $base)))
            (local.set $i (i32.add (local.get $i) (i32.const 1)))
            (i32.eq (local.get $i) (local.get $expo))
            br_if $exit_pow_loop
            br $pow_loop
        )
        )
        (local.get $result)
    )

    (func $__u64_pow (param $base i64) (param $expo i64) (result i64)
        (local $result i64)
        (local $i i64)
        (i64.eqz (local.get $expo))
        if 
            (i64.const 1)
            (return)
        end
        (local.set $i (i64.const 0))
        (local.set $result (i64.const 1))
        (loop $pow_loop (block $exit_pow_loop
            (local.set $result (i64.mul (local.get $result) (local.get $base)))
            (local.set $i (i64.add (local.get $i) (i64.const 1)))
            (i64.eq (local.get $i) (local.get $expo))
            br_if $exit_pow_loop
            br $pow_loop
        )
        )
        (local.get $result)
    )

    (global $__bool_true i32 (i32.const 4000))
    (data (i32.const 4000) "\04\00\00\00true")
    (global $__bool_false i32 (i32.const 4500))
    (data (i32.const 4500) "\05\00\00\00false")
    (data (i32.const 51) "0123456789")
    (global $__default_assertion_error i32 (i32.const 80))
    (data (i32.const 80) "\FF\00\00\00Assertion error")
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
        ;; (local.set $addr (call $__malloc (local.get $numlen)))
        (local.set $addr (call $__malloc (i32.add (local.get $numlen) (i32.const 4))))
        (local.set $start (i32.add (local.get $addr) (i32.const 4)))
        (i32.store offset=0 (local.get $addr) (local.get $numlen))
        (if (i32.eq (local.get $isnegative) (i32.const 1))
            (then ;; result[0] = '-'
                ;; (i32.store8 (local.get $addr) (i32.const 45))
                (i32.store8 (local.get $start) (i32.const 45))

            )
        )
        (local.set $writeidx 
            (i32.sub 
                ;; (i32.add (local.get $addr) (local.get $numlen))
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
        (i32.load8_u (i32.add (i32.add (local.get $str) (i32.const 4)) (local.get $index)))
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
                (i32.gt_s (local.get $end) (local.get $len))
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

    (func $__str_eq (export "__str_eq") (param $str1 i32) (param $str2 i32) (result i32)
        (local $len1 i32)
        (local $len2 i32)
        (local $start1 i32)
        (local $start2 i32)
        (local $i i32)
        (local.set $len1 (call $__str_length (local.get $str1)))
        (local.set $len2 (call $__str_length (local.get $str2)))
        
        (if (i32.ne (local.get $len1) (local.get $len2))
            (then
                (return (i32.const 0))
            )
        )
        (local.set $start1 (i32.add (local.get $str1) (i32.const 4)))
        (local.set $start2 (i32.add (local.get $str2) (i32.const 4)))
        (loop $eq_loop (block $exit_eq_loop
            (i32.eq (local.get $i) (local.get $len1))
            br_if $exit_eq_loop
            (i32.load8_u (i32.add (local.get $start1) (local.get $i)))
            (i32.load8_u (i32.add (local.get $start2) (local.get $i)))
            (if (i32.ne)
                (then (return (i32.const 0)))  
            )
            (local.set $i (i32.add (local.get $i) (i32.const 1)))
            br $eq_loop
        ))
        i32.const 1
    )

    (func $__str_ne (export "__str_ne") (param $str1 i32) (param $str2 i32) (result i32)
        (call $__str_eq (local.get $str1) (local.get $str2))
        (i32.xor (i32.const 1))
    )

    (func $__str_lt (export "__str_lt") (param $str1 i32) (param $str2 i32) (result i32)
        (local $len1 i32)
        (local $len2 i32)
        (local $start1 i32)
        (local $start2 i32)
        (local $char1 i32)
        (local $char2 i32)
        (local $len i32)
        (local $i i32)
        (local.set $start1 (i32.add (local.get $str1) (i32.const 4)))
        (local.set $start2 (i32.add (local.get $str2) (i32.const 4)))
        (local.set $len1 (call $__str_length (local.get $str1)))
        (local.set $len2 (call $__str_length (local.get $str2)))
        (local.set $len (local.get $len1))
        (if (i32.lt_u (local.get $len2) (local.get $len1))
            (then (local.set $len (local.get $len2)))
        )
        (loop $compare_loop (block $exit_compare_loop
            (i32.eq (local.get $i) (local.get $len))
            (br_if $exit_compare_loop)
            (local.set $char1 (i32.load8_u (i32.add (local.get $start1) (local.get $i))))
            (local.set $char2 (i32.load8_u (i32.add (local.get $start2) (local.get $i))))
            (if (i32.ne (local.get $char1) (local.get $char2))
                (then
                    (return (i32.lt_u (local.get $char1) (local.get $char2)))
                )
            )
            (local.set $i (i32.add (local.get $i) (i32.const 1)))
            (br $compare_loop)
        ))
        (i32.lt_u (local.get $len1) (local.get $len2))
    )

    (func $__str_le (export "__str_le") (param $str1 i32) (param $str2 i32) (result i32)
        (local $len1 i32)
        (local $len2 i32)
        (local $start1 i32)
        (local $start2 i32)
        (local $char1 i32)
        (local $char2 i32)
        (local $len i32)
        (local $i i32)
        (local.set $start1 (i32.add (local.get $str1) (i32.const 4)))
        (local.set $start2 (i32.add (local.get $str2) (i32.const 4)))
        (local.set $len1 (call $__str_length (local.get $str1)))
        (local.set $len2 (call $__str_length (local.get $str2)))
        (local.set $len (local.get $len1))
        (if (i32.lt_u (local.get $len2) (local.get $len1))
            (then (local.set $len (local.get $len2)))
        )
        (loop $compare_loop (block $exit_compare_loop
            (i32.eq (local.get $i) (local.get $len))
            (br_if $exit_compare_loop)
            (local.set $char1 (i32.load8_u (i32.add (local.get $start1) (local.get $i))))
            (local.set $char2 (i32.load8_u (i32.add (local.get $start2) (local.get $i))))
            (if (i32.ne (local.get $char1) (local.get $char2))
                (then
                    (return (i32.lt_u (local.get $char1) (local.get $char2)))
                )
            )
            (local.set $i (i32.add (local.get $i) (i32.const 1)))
            (br $compare_loop)
        ))
        (i32.le_u (local.get $len1) (local.get $len2))
    )

    (func $__str_gt (export "__str_gt") (param $str1 i32) (param $str2 i32) (result i32)
        (local $len1 i32)
        (local $len2 i32)
        (local $start1 i32)
        (local $start2 i32)
        (local $char1 i32)
        (local $char2 i32)
        (local $len i32)
        (local $i i32)
        (local.set $start1 (i32.add (local.get $str1) (i32.const 4)))
        (local.set $start2 (i32.add (local.get $str2) (i32.const 4)))
        (local.set $len1 (call $__str_length (local.get $str1)))
        (local.set $len2 (call $__str_length (local.get $str2)))
        (local.set $len (local.get $len1))
        (if (i32.lt_u (local.get $len2) (local.get $len1))
            (then (local.set $len (local.get $len2)))
        )
        (loop $compare_loop (block $exit_compare_loop
            (i32.eq (local.get $i) (local.get $len))
            (br_if $exit_compare_loop)
            (local.set $char1 (i32.load8_u (i32.add (local.get $start1) (local.get $i))))
            (local.set $char2 (i32.load8_u (i32.add (local.get $start2) (local.get $i))))
            (if (i32.ne (local.get $char1) (local.get $char2))
                (then
                    (return (i32.gt_u (local.get $char1) (local.get $char2)))
                )
            )
            (local.set $i (i32.add (local.get $i) (i32.const 1)))
            (br $compare_loop)
        ))
        (i32.gt_u (local.get $len1) (local.get $len2))
    )

    (func $__str_ge (export "__str_ge") (param $str1 i32) (param $str2 i32) (result i32)
        (local $len1 i32)
        (local $len2 i32)
        (local $start1 i32)
        (local $start2 i32)
        (local $char1 i32)
        (local $char2 i32)
        (local $len i32)
        (local $i i32)
        (local.set $start1 (i32.add (local.get $str1) (i32.const 4)))
        (local.set $start2 (i32.add (local.get $str2) (i32.const 4)))
        (local.set $len1 (call $__str_length (local.get $str1)))
        (local.set $len2 (call $__str_length (local.get $str2)))
        (local.set $len (local.get $len1))
        (if (i32.lt_u (local.get $len2) (local.get $len1))
            (then (local.set $len (local.get $len2)))
        )
        (loop $compare_loop (block $exit_compare_loop
            (i32.eq (local.get $i) (local.get $len))
            (br_if $exit_compare_loop)
            (local.set $char1 (i32.load8_u (i32.add (local.get $start1) (local.get $i))))
            (local.set $char2 (i32.load8_u (i32.add (local.get $start2) (local.get $i))))
            (if (i32.ne (local.get $char1) (local.get $char2))
                (then
                    (return (i32.gt_u (local.get $char1) (local.get $char2)))
                )
            )
            (local.set $i (i32.add (local.get $i) (i32.const 1)))
            (br $compare_loop)
        ))
        (i32.ge_u (local.get $len1) (local.get $len2))
    )

    (func $__str_reverse (export "__str_reverse") (param $str i32) (result i32)
        (local $reversed i32)
        (local $start i32)
        (local $rstart i32)
        (local $len i32)
        (local $i i32)
        (local.set $start (i32.add (local.get $str) (i32.const 4)))
        (local.set $len (call $__str_length (local.get $str)))
        (local.set $i (i32.sub (local.get $len) (i32.const 1)))
        (call $__malloc (i32.add (local.get $len) (i32.const 4)))
        (local.set $reversed)
        (local.set $rstart (i32.add (local.get $reversed) (i32.const 4)))
        (i32.store (local.get $reversed) (local.get $len))
        (loop $str_loop (block $exit_str_loop
            (i32.lt_s (local.get $i) (i32.const 0))
            (br_if $exit_str_loop)
            (i32.store8 
                (i32.add (local.get $rstart) (i32.sub (i32.sub (local.get $len) (i32.const 1)) (local.get $i)))
                (i32.load8_u (i32.add (local.get $start) (local.get $i)))
            )
            (local.set $i (i32.sub (local.get $i) (i32.const 1)))
            (br $str_loop)
        ))
        (local.get $reversed)
    )

    (func $__str_startsWith (export "__str_startsWith") (param $prefix i32) (param $str i32) (result i32)
        (local $prefixlen i32)
        (local $len i32)
        (local $prefixstart i32)
        (local $start i32)
        (local $i i32)
        (local.set $prefixlen (call $__str_length (local.get $prefix)))
        (local.set $len (call $__str_length (local.get $str)))
        (if (i32.lt_u (local.get $len) (local.get $prefixlen))
            (then
                (return (i32.const 0))
            )
        )
        (local.set $prefixstart (i32.add (local.get $prefix) (i32.const 4)))
        (local.set $start (i32.add (local.get $str) (i32.const 4)))
        (loop $compare_loop (block $exit_compare_loop
            (i32.eq (local.get $i) (local.get $prefixlen))
            (br_if $exit_compare_loop)
            (i32.load8_u (i32.add (local.get $prefixstart) (local.get $i)))
            (i32.load8_u (i32.add (local.get $start) (local.get $i)))
            (i32.ne)
            if
                (return (i32.const 0))
            end
            (local.set $i (i32.add (local.get $i) (i32.const 1)))
            (br $compare_loop)
        ))
        (i32.const 1)
    )

    (func $__str_endsWith (export "__str_endsWith") (param $suffix i32) (param $str i32) (result i32)
        (local $suffixlen i32)
        (local $len i32)
        (local $suffixstart i32)
        (local $start i32)
        (local $i i32)
        (local.set $suffixlen (call $__str_length (local.get $suffix)))
        (local.set $len (call $__str_length (local.get $str)))
        (if (i32.lt_u (local.get $len) (local.get $suffixlen))
            (then
                (return (i32.const 0))
            )
        )
        (local.set $suffixstart (i32.add (local.get $suffix) (i32.const 4)))
        (local.set $start (i32.add 
            (i32.add (local.get $str) (i32.const 4)) 
            (i32.sub (local.get $len) (local.get $suffixlen)))
        ) ;; str + 4 + (len - suffixlen)
        (loop $compare_loop (block $exit_compare_loop
            (i32.eq (local.get $i) (local.get $suffixlen))
            (br_if $exit_compare_loop)
            (i32.load8_u (i32.add (local.get $suffixstart) (local.get $i)))
            (i32.load8_u (i32.add (local.get $start) (local.get $i)))
            (i32.ne)
            if
                (return (i32.const 0))
            end
            (local.set $i (i32.add (local.get $i) (i32.const 1)))
            (br $compare_loop)
        ))
        (i32.const 1)
    )

    (func $__str_contains_String (export "__str_contains_String") (param $inner i32) (param $str i32) (result i32)
        (local $innerLen i32)
        (local $innerStart i32)
        (local $len i32)
        (local $start i32)
        (local $matchLen i32)
        (local $i i32)
        (local $j i32)
        (local.set $innerLen (call $__str_length (local.get $inner)))
        (local.set $len (call $__str_length (local.get $str)))
        (if (i32.lt_u (local.get $len) (local.get $innerLen))
            (then (return (i32.const 0)))
        )
        (local.set $innerStart (i32.add (local.get $inner) (i32.const 4)))
        (local.set $start (i32.add (local.get $str) (i32.const 4)))
        (loop $outerloop (block $exit_outerloop
            (i32.eq (local.get $i) (local.get $len))
            (br_if $exit_outerloop)
            (if (i32.eq (local.get $matchLen) (local.get $innerLen))
                (then (return (i32.const 1)))
            )
            (i32.load8_u (i32.add (local.get $start) (local.get $i)))
            (i32.load8_u (local.get $innerStart))
            (if (i32.eq)
                (then
                    (local.set $matchLen (i32.const 1))
                    (local.set $j (i32.add (local.get $i) (i32.const 1)))
                    (loop $innerLoop (block $exit_innerLoop
                        (i32.eq (local.get $j) (local.get $len))
                        (br_if $exit_innerLoop)
                        (if (i32.eq (local.get $matchLen) (local.get $innerLen))
                            (then (return (i32.const 1)))
                        )
                        (if (i32.ne 
                            (i32.load8_u (i32.add (local.get $start) (local.get $j)))
                            (i32.load8_u (i32.add (local.get $innerStart) (local.get $matchLen)))
                        )
                            (then
                                (local.set $matchLen (i32.const 0))
                                (br $exit_innerLoop)
                            )
                        )
                        (local.set $matchLen (i32.add (local.get $matchLen) (i32.const 1)))
                        (local.set $j (i32.add (local.get $j) (i32.const 1)))
                        (br $innerLoop)
                    ))
                )
            )
            (local.set $i (i32.add (local.get $i) (i32.const 1)))
            (br $outerloop)
        ))
        (i32.eq (local.get $matchLen) (local.get $innerLen))
    )

    (func $__str_toUpper (export "__str_toUpper") (param $str i32) (result i32)
        (local $len i32)
        (local $newStr i32)
        (local $newStart i32)
        (local $start i32)
        (local $char i32)
        (local $i i32)
        (local.set $len (call $__str_length (local.get $str)))
        (local.set $start (i32.add (local.get $str) (i32.const 4)))
        (call $__malloc (i32.add (local.get $len) (i32.const 4)))
        (local.set $newStr)
        (i32.store (local.get $newStr) (local.get $len))
        (local.set $newStart (i32.add (local.get $newStr) (i32.const 4)))
        (loop $upperLoop (block $exit_upperLoop
            (i32.eq (local.get $i) (local.get $len))
            (br_if $exit_upperLoop)
            (local.set $char (i32.load8_u (i32.add (local.get $start) (local.get $i))))
            (if (i32.and
                    (i32.ge_u (local.get $char) (i32.const 97))
                    (i32.le_u (local.get $char) (i32.const 122))
                )
                (then
                    (i32.store8 (i32.add (local.get $newStart) (local.get $i))
                        (i32.sub 
                            (local.get $char)
                            (i32.const 32))
                    )
                )
                (else
                    (i32.store8 (i32.add (local.get $newStart) (local.get $i)) (local.get $char))
                )
            )
            (local.set $i (i32.add (local.get $i) (i32.const 1)))
            (br $upperLoop)
        ))
        (local.get $newStr)
    )

    (func $__str_toLower (export "__str_toLower") (param $str i32) (result i32)
        (local $len i32)
        (local $newStr i32)
        (local $newStart i32)
        (local $start i32)
        (local $char i32)
        (local $i i32)
        (local.set $len (call $__str_length (local.get $str)))
        (local.set $start (i32.add (local.get $str) (i32.const 4)))
        (call $__malloc (i32.add (local.get $len) (i32.const 4)))
        (local.set $newStr)
        (i32.store (local.get $newStr) (local.get $len))
        (local.set $newStart (i32.add (local.get $newStr) (i32.const 4)))
        (loop $upperLoop (block $exit_upperLoop
            (i32.eq (local.get $i) (local.get $len))
            (br_if $exit_upperLoop)
            (local.set $char (i32.load8_u (i32.add (local.get $start) (local.get $i))))
            (if (i32.and
                    (i32.ge_u (local.get $char) (i32.const 65))
                    (i32.le_u (local.get $char) (i32.const 90))
                )
                (then
                    (i32.store8 (i32.add (local.get $newStart) (local.get $i))
                        (i32.add 
                            (local.get $char)
                            (i32.const 32))
                    )
                )
                (else
                    (i32.store8 (i32.add (local.get $newStart) (local.get $i)) (local.get $char))
                )
            )
            (local.set $i (i32.add (local.get $i) (i32.const 1)))
            (br $upperLoop)
        ))
        (local.get $newStr)
    )

    (func $__str_trim (export "__str_trim") (param $str i32) (result i32)
        (call $__str_trimStart (local.get $str))
        (call $__str_trimEnd)
    )

    (func $__str_trimStart (export "__str_trimStart") (param $str i32) (result i32)
        (local $len i32)
        (local $first_nonspace i32)
        (local $newStr i32)
        (local $newStart i32)
        (local $newLen i32)
        (local $start i32)
        (local $char i32)
        (local $i i32)
        (local.set $len (call $__str_length (local.get $str)))
        (local.set $start (i32.add (local.get $str) (i32.const 4)))
        (loop $count_loop (block $exit_count_loop
            (i32.eq (local.get $i) (local.get $len))
            (br_if $exit_count_loop)
            (local.set $char (i32.load8_u (i32.add (local.get $start) (local.get $i))))
            (i32.or
                (i32.eq (local.get $char) (i32.const 9))
                (i32.or
                    (i32.eq (local.get $char) (i32.const 10))
                    (i32.or
                        (i32.eq (local.get $char) (i32.const 11))
                        (i32.or
                            (i32.eq (local.get $char) (i32.const 12))
                            (i32.or
                                (i32.eq (local.get $char) (i32.const 13))
                                (i32.eq (local.get $char) (i32.const 32))
                            )
                        )
                    )
                )
            )
            (i32.eqz)
            (br_if $exit_count_loop)
            (local.set $first_nonspace (i32.add (local.get $first_nonspace) (i32.const 1)))
            (local.set $i (i32.add (local.get $i) (i32.const 1)))
            (br $count_loop)
        ))
        (local.set $newLen (i32.sub (local.get $len) (local.get $first_nonspace)))
        (call $__malloc (i32.add (local.get $newLen) (i32.const 4)))
        (local.set $newStr)
        (i32.store (local.get $newStr) (local.get $newLen))
        (local.set $newStart (i32.add (local.get $newStr) (i32.const 4)))
        (memory.copy
            (local.get $newStart)
            (i32.add (local.get $start) (local.get $first_nonspace))
            (local.get $newLen)
        )
        (local.get $newStr)
    )

    (func $__str_trimEnd (export "__str_trimEnd") (param $str i32) (result i32)
        (local $len i32)
        (local $last_nonspace i32)
        (local $newStr i32)
        (local $newStart i32)
        (local $newLen i32)
        (local $start i32)
        (local $char i32)
        (local $i i32)
        (local.set $len (call $__str_length (local.get $str)))
        (local.set $start (i32.add (local.get $str) (i32.const 4)))
        (local.set $i (i32.sub (local.get $len) (i32.const 1)))
        (local.set $last_nonspace (i32.sub (local.get $len) (i32.const 1)))
        (loop $count_loop (block $exit_count_loop
            (i32.lt_s (local.get $i) (i32.const 0))
            (br_if $exit_count_loop)
            (local.set $char (i32.load8_u (i32.add (local.get $start) (local.get $i))))
            (i32.or
                (i32.eq (local.get $char) (i32.const 9))
                (i32.or
                    (i32.eq (local.get $char) (i32.const 10))
                    (i32.or
                        (i32.eq (local.get $char) (i32.const 11))
                        (i32.or
                            (i32.eq (local.get $char) (i32.const 12))
                            (i32.or
                                (i32.eq (local.get $char) (i32.const 13))
                                (i32.eq (local.get $char) (i32.const 32))
                            )
                        )
                    )
                )
            )
            (i32.eqz)
            (br_if $exit_count_loop)
            (local.set $last_nonspace (i32.sub (local.get $last_nonspace) (i32.const 1)))
            (local.set $i (i32.sub (local.get $i) (i32.const 1)))
            (br $count_loop)
        ))
        (local.set $newLen (i32.sub (local.get $len) (i32.sub (i32.sub (local.get $len) (i32.const 1)) (local.get $last_nonspace))))
        (call $__malloc (i32.add (local.get $newLen) (i32.const 4)))
        (local.set $newStr)
        (i32.store (local.get $newStr) (local.get $newLen))
        (local.set $newStart (i32.add (local.get $newStr) (i32.const 4)))
        (memory.copy
            (local.get $newStart)
            (local.get $start)
            (local.get $newLen)
        )
        (local.get $newStr)
    )
    
    (func $__str_replace_char_char (export "__str_replace_char_char") (param $old i32) (param $new i32) (param $str i32) (result i32)
        (local $newStr i32)
        (local $len i32)
        (local $newLen i32)
        (local $replaced i32)
        (local $start i32)
        (local $newStart i32)
        (local $char i32)
        (local $i i32)
        (local.set $len (call $__str_length (local.get $str)))
        (local.set $start (i32.add (local.get $str) (i32.const 4)))
        (if (i32.eq (local.get $old) (local.get $new))
            (then
                (call $__str_slice (local.get $str) (i32.const 0) (local.get $len)) ;; make a copy
                (return)
            )
        )
        (if (i32.eqz (local.get $len))
            (then 
                (if (i32.eqz (local.get $new))
                    (then
                        (local.set $newStr (call $__malloc (i32.const 4)))
                        (i32.store (local.get $newStr) (i32.const 0))
                        (return (local.get $newStr))
                    )
                    (else
                        (if (i32.eqz (local.get $old))
                            (then
                                (local.set $newStr (call $__malloc (i32.const 5)))
                                (i32.store (local.get $newStr) (i32.const 1))
                                (i32.store8 (i32.add (local.get $newStr) (i32.const 4)) (local.get $new))
                                (return (local.get $newStr))
                            )
                        )
                        
                    )
                )
            )
        )
        (if (i32.eqz (local.get $old))
            (then
                (call $__char_concat_str (local.get $new) (local.get $str))
                (return)
            )
        )
        (call $__malloc (i32.add (local.get $len) (i32.const 4)))
        (local.set $newStr)
        (i32.store (local.get $newStr) (local.get $len))
        (local.set $newStart (i32.add (local.get $newStr) (i32.const 4)))
        (loop $replace_loop (block $exit_replace_loop
            (i32.eq (local.get $i) (local.get $len))
            (br_if $exit_replace_loop)
            (i32.load8_u (i32.add (local.get $start) (local.get $i)))
            (local.set $char)
            (if (i32.and (i32.eq (local.get $char) (local.get $old)) (i32.eqz (local.get $replaced)))
                (then
                    (local.set $replaced (i32.const 1))
                    (if (i32.ne(local.get $new) (i32.const 0))
                        (then
                            (i32.store8 (i32.add (local.get $newStart) (local.get $i)) (local.get $new))
                        )
                    )
                )
                (else
                    (i32.store8 (i32.add (local.get $newStart) (local.get $i)) (local.get $char))
                )
            )
            (local.set $i (i32.add (local.get $i) (i32.const 1)))
            (br $replace_loop)
        ))
        (local.get $newStr)
    )

    (func $__str_replaceAll_char_char (export "__str_replaceAll_char_char") (param $old i32) (param $new i32) (param $str i32) (result i32)
        (local $newStr i32)
        (local $len i32)
        (local $newLen i32)
        (local $start i32)
        (local $newStart i32)
        (local $char i32)
        (local $insert i32)
        (local $i i32)
        (local $j i32)
        (local.set $len (call $__str_length (local.get $str)))
        (local.set $newLen (local.get $len))
        (local.set $start (i32.add (local.get $str) (i32.const 4)))
        (if (i32.eq (local.get $old) (local.get $new))
            (then
                (call $__str_slice (local.get $str) (i32.const 0) (local.get $len)) ;; make a copy
                (return)
            )
        )
        (if (i32.eqz (local.get $len))
            (then 
                (if (i32.eqz (local.get $new))
                    (then
                        (local.set $newStr (call $__malloc (i32.const 4)))
                        (i32.store (local.get $newStr) (i32.const 0))
                        (return (local.get $newStr))
                    )
                    (else
                        (if (i32.eqz (local.get $old))
                            (then
                                (local.set $newStr (call $__malloc (i32.const 5)))
                                (i32.store (local.get $newStr) (i32.const 1))
                                (i32.store8 (i32.add (local.get $newStr) (i32.const 4)) (local.get $new))
                                (return (local.get $newStr))
                            )
                        )
                        
                    )
                )
            )
        )
        (if (i32.eqz (local.get $old))
            (then
                (local.set $insert (local.get $new))
                (local.set $newLen (i32.add (i32.mul (local.get $len) (i32.const 2)) (i32.const 1)))
            )
        )
        (call $__malloc (i32.add (local.get $newLen) (i32.const 4)))
        (local.set $newStr)
        (i32.store (local.get $newStr) (local.get $newLen))
        (local.set $newStart (i32.add (local.get $newStr) (i32.const 4)))
        (loop $replace_loop (block $exit_replace_loop
            (i32.eq (local.get $i) (local.get $len))
            (br_if $exit_replace_loop)
            (if (i32.ne (local.get $insert) (i32.const 0))
                (then
                    (i32.store8 (i32.add (local.get $newStart) (local.get $j)) (local.get $insert))
                    (local.set $j (i32.add (local.get $j) (i32.const 1)))
                )
            )
            (i32.load8_u (i32.add (local.get $start) (local.get $i)))
            (local.set $char)
            (if (i32.eq (local.get $char) (local.get $old))
                (then
                    (if (i32.ne(local.get $new) (i32.const 0))
                        (then
                            (i32.store8 (i32.add (local.get $newStart) (local.get $j)) (local.get $new))
                        )
                    )
                )
                (else
                    (i32.store8 (i32.add (local.get $newStart) (local.get $j)) (local.get $char))
                )
            )
            (local.set $i (i32.add (local.get $i) (i32.const 1)))
            (local.set $j (i32.add (local.get $j) (i32.const 1)))
            (br $replace_loop)
        ))
        (if (i32.ne (local.get $insert) (i32.const 0))
            (then
                (call $__str_concat_char (local.get $newStr) (local.get $insert))
                (return)
            )
        )
        (local.get $newStr)
    )

    ;;
    ;; Entry point
    ;;
    (func $_start (export "_start")
        (call $main)
        (call $exit_int)
    )

    ;;
    ;; Compiler Generated Code
    ;;

    (data $__str_const0 (i32.const 100) "\0c\00\00\00hello, world")
    (data $__str_const1 (i32.const 500) "\06\00\00\00Name: ")
    (data $__str_const2 (i32.const 300) "\06\00\00\00\n\tLL \n")
    (data $__str_const3 (i32.const 310) "\00\00\00\00")
    ;;(data $__str_const1 (i32.const 200) "\02\00\00\00!\n")
    (func $main (export "main") (result i32)
        (call $println (i32.const 100))
        ;;(call $__str_fromInt32 (i32.const -11203))
        ;;(call $__str_fromInt64 (i64.const 56))
        ;;(call $__str_fromChar (i32.const 65))
        ;;(call $__str_fromBool (i32.const 1))
        ;;(call $__str_indexOf (i32.const 101) (i32.const 100))
        ;;(call $__str_contains_char (i32.const 101) (i32.const 100))
        ;; (call $__str_fromChar 
        ;;     (call $__str_index (i32.const 100) (i32.const 16))
        ;; )
        ;;(call $__char_concat (i32.const 65) (i32.const 45))
        ;;(call $__char_concat_str (i32.const 45) (i32.const 100))
        ;;(call $__str_concat (i32.const 100) (i32.const 200))
        ;;(call $__str_concat_char (i32.const 100) (i32.const 72))
        ;;(call $prompt (i32.const 500))
        ;;(call $__str_slice (i32.const 100) (i32.const 10) (i32.const 4))
        ;;(call $__str_slice_inclusive (i32.const 100) (i32.const 1) (i32.const 11))
        ;; (call $__str_reverse (i32.const 100))
        ;;(call $__str_toUpper (i32.const 100))
        ;;(call $__str_toLower (i32.const 300))
        ;;(call $__str_trimStart (i32.const 300))
        ;;(call $__str_trimEnd (i32.const 300))
        ;;(call $__str_trim (i32.const 300))
        ;; (call $__str_length)
        ;; (call $__str_fromInt32)
        ;; (call $__str_replace_char_char (i32.const 108) (i32.const 95) (i32.const 100))
        (call $__str_replaceAll_char_char (i32.const 0) (i32.const 95) (i32.const 100))
        (call $println)
        (i32.const 0)
        ;;(call $__str_eq (i32.const 100) (i32.const 100))
        ;;(call $__str_gt (i32.const 100) (i32.const 300))
        ;;(call $__str_endsWith (i32.const 300) (i32.const 100))
        ;;(call $__str_contains_String (i32.const 300) (i32.const 100))
        (return)
    )

)