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
