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
