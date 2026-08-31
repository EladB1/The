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

        (call $__align (i32.const 4))

        (local.get $curr_alloc_addr)
    )

    (func $__align (param $size i32)
        (i32.eqz (i32.rem_u (global.get $malloc_next) (local.get $size)))
        if
            (return)
        else
            (i32.mul
                (local.get $size)
                (i32.add
                    (i32.const 1)
                    (i32.div_u (global.get $malloc_next) (local.get $size))
                )
            )
            (global.set $malloc_next)
        end
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
        (local.get $addr)
    )
