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
