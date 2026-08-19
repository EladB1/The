Basic Program
---

<table>
<tr>
<th><center>Source Code</center></th>
<th><center>IR</center></th>
</tr>
<tr>
<td>

```
float e = 2.718;

fn test(int i) -> int {
    return i + 1;
}

fn main() -> int {
    int i = 1;
    double pi = 3.14159;
    mut double value = e ** pi - 1;
    value -= test(i);
    String db = "mariadb";
    bool isOpen = true;
    return 0;
}

```

</td>
<td>

```
STORE global.e: f32 f32(2.718)

fn test(param.i: i32) -> i32 {
    __t0: i32 = i32.add local.i i32(1)
    return __t0 
}
fn main() -> i32 {
    STORE local.i i32(1)
    STORE local.pi f64(3.14159)
    PARAM global.e
    PARAM local.pi
    __t1: f64 = CALL __pow 2 // __pow is part of the runtime library
    __t2: f64 = f64.sub __t1 i32(1)
    STORE local.value __t2
    PARAM local.i
    __t3: i32 = CALL test 1 // number of arguments
    __t4: f64 = GET local.value
    __t5: f64 = f64.sub __t4 __t3
    STORE local.value: f64 __t5
    STORE local.db STR_CONST(0)
    STORE local.isOpen i32(1) // under the hood, treat bools as i32
    return i32(0)
}
```

</td>
</tr>
</table>

Control Flow
---

<table>
<tr>
<th><center>Source Code</center></th>
<th><center>IR</center></th>
</tr>
<tr>
<td>

```
fn main() -> int {
    int limit = 100;
    for (int i in 0 ..= limit) {
        if (i % 7 == 0) {
            mut int j = i;
            while (j < i + 7) {
                if ((j+i) % 12 == 0)
                    break;
                println(j+i);
                j++;
            }
        }
        else if (i % 2 == 0) {
            continue;
        }
        else
            println(i);
    }

    return 0;
}

```

</td>
<td>

```
fn main() -> i32 {
    STORE local.limit i32(100)
    block loop_exit@0: {
        // Loop initialization(s)
        STORE local.i i32(0)
        loop for@0: {
            // Loop condition
            __t0: i32 = i32.le local.i local.limit
            __t1: i32 = i32.eq __t0 i32(0) // condition is false
            JMPIF loop_exit@0 __t1
            block loop_body@0: {
                __t2: i32 = i32.mod local.i i32(7)
                __t3: i32 = i32.eq __t2 i32(0)
                if __t3 {
                    STORE local.j: i32 local.i
                    block loop_exit@1: {
                        loop for@1: {
                            // Loop condition
                            __t4: i32 = add local.i i32(7)
                            __t5: i32 = i32.lt local.j __t4
                            __t6: i32 = i32.eq __t5 i32(0)
                            JMPIF loop_exit@1 __t6
                            block loop_body@1: {
                                __t7: i32 = add local.j local.i
                                __t8: i32 = i32.mod __t7 i32(12)
                                __t9: i32 = i32.eq __t8 i32(0)
                                if __t9 {
                                    JMP loop_exit@1
                                }
                                 __t10: i32 = add local.j local.i
                                PARAM __t10
                                __t11: str_const = CALL __str_cast_i32 1
                                PARAM __t11
                                CALL println 1
                                __t12: i32 = add local.j i32(1)
                                STORE local.j __t12
                                }
                                // repeat loop
                                JMP for@1
                        }
                    }
                }
                else {
                    __t13 i32 = i32.mod local.i i32(2)
                    __t14: i32 = i32.eq t13 i32(0)
                    if __t14 {
                        JMP @loop_body0
                    }             
                    else {
                        PARAM local.i
                        __t15: str_const = CALL __str_cast_i32 1
                        PARAM __t15
                        CALL println 1
                    }
                }
            }
            __t16: i32 = i32.add local.i i32(1) 
            STORE local.i __t16
            JMP for@0
        }
    }
    return i32(0)
}
```

</td>
</tr>
</table>

>  JMP/JMPIF: \
>    block: execute from the end of the last "instruction" \
>    loop:  execute from the first "instruction"
    

Strings
---



Structs
---

