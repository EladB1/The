(module
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
    (global $NEWLINE_CHAR_ADDR i32 (i32.const 1000))
    (data (i32.const 5000) "0123456789")
    (global $itoa_out_buf i32 (i32.const 5010))
    (data (i32.const 1000) "\n")
    (global $stdin_buffer i32 (i32.const 200))
    (data (i32.const 200) "\00")
    (global $iovec_stdin i32 (i32.const 300))
    (global $iovec_stdin_len i32 (i32.const 304))
    (data (i32.const 300) "\00\00\00\00\00\00\00\00")
    ;; byte counter
    (global $stdin_byte_counter i32 (i32.const 400))
    (data (i32.const 400) "\00\00\00\00")
    
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


