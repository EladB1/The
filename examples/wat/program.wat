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

    (global $malloc_start i32 (i32.const 1024))
    (global $malloc_next i32 (i32.const 1024))

    (func $__malloc (param $size i32) (result (;memory address;) i32)
        (local $curr_alloc_addr i32)
        (local $next_alloc_addr i32)

        (global.get $malloc_next)
        (local.set $curr_alloc_addr)

        (i32.add (local.get $curr_alloc_addr) (local.get $size))
        (local.set $next_alloc_addr)

        (local.get $next_alloc_addr)
        (i32.store (global.get $malloc_next))

        (local.get $curr_alloc_addr)
    )
    
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

    (func $__read_stdin (result i32)
        (i32.store (global.get $iovec_stdin) (global.get $stdin_buffer))
        (i32.store (global.get $iovec_stdin_len) (i32.const 256)) ;; allow 256 bytes
        (call $__fd_read
            (i32.const 0) ;; stdin
            (global.get $iovec_stdin)
            (i32.const 1)
            (global.get $stdin_byte_counter)
        )
        drop
        (i32.store (global.get $iovec_stdin_len) (i32.load (global.get $stdin_byte_counter)))
        (global.get $stdin_buffer)
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

    (global $__bool_true i32 (i32.const 40))
    (data (i32.const 40) "true")
    (global $__bool_false i32 (i32.const 45))
    (data (i32.const 45) "false")
    (data (i32.const 51) "0123456789")
    (global $__default_assertion_error i32 (i32.const 80))
    (data (i32.const 80) "Assertion error")
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
        (local.set $addr (call $__malloc (local.get $numlen)))
        (if (i32.eq (local.get $isnegative) (i32.const 1))
            (then ;; result[0] = '-'
                (i32.store8 (local.get $addr) (i32.const 45))
            )
        )
        (local.set $writeidx 
            (i32.sub 
                (i32.add (local.get $addr) (local.get $numlen))
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
            (i32.eq (local.get $writeidx) (i32.add (local.get $addr) (local.get $index)))
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
        (local.set $addr (call $__malloc (local.get $numlen)))
        (if (i32.eq (local.get $isnegative) (i32.const 1))
            (then ;; result[0] = '-'
                (i32.store8 (local.get $addr) (i32.const 45))
            )
        )
        (local.set $writeidx 
            (i32.sub 
                (i32.add (local.get $addr) (local.get $numlen))
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
            (i32.eq (local.get $writeidx) (i32.add (local.get $addr) (local.get $index)))
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
        (call $__malloc (i32.const 2))
        (local.set $addr)
        (i32.store8 (i32.add (local.get $addr) (i32.const 0)) (local.get $value))
        (i32.store8 (i32.add (local.get $addr) (i32.const 1)) (i32.const 0)) ;; null terminator
        (local.get $addr)
    )

    (func $__str_indexOf (export "__str_indexOf") (param $char i32) (param $ptr i32) (result i32)
       (local $i i32)
       (local $curr i32)
       (loop $str_loop (block $exit__str_loop
        (i32.load8_u (i32.add (local.get $ptr) (local.get $i)))
        (local.tee $curr)
        (if (i32.eqz)
            (then
                (br $exit__str_loop)
            )
        )
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

    (data $__str_const0 (i32.const 100) "")
    (func $main (export "main") (result i32)
        (call $__i32_pow (i32.const 2) (i32.const 10))
        (call $__str_fromInt32)
        (call $println)
        (call $__i64_pow (i64.const 10) (i64.const 3))
        (call $__str_fromInt64)
        (call $println)
        (i32.const 0)
        (return)
    )

)