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
    t1: i32 = GET local.i
    t2: i32 = i32.add t1 i32(1)
    return t2 
}
fn main -> i32 {
    STORE local.i: i32 i32(1)
    STORE local.pi: f64 f64(3.14159)
    t3: f32 = GET global.e
    PARAM: f32 e
    PARAM: f64 local.pi
    t4: f64 = CALL __pow 2 // __pow is part of the runtime library
    t5: f64 = f64.sub t4 i32(1)
    STORE local.value: f64 t5
    PARAM: i32 local.i
    t6: i32 = CALL test 1 // number of arguments
    t7: f64 = GET local.value
    t8: f64 = f64.sub t7 t6
    STORE local.value: f64 t8
    STORE local.db: ptr STR_CONST(0)
    STORE local.isOpen: i32 i32(1) // under the hood, treat bools as i32
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
fn main -> i32 {
    STORE local.limit: i32 i32(100)
    block loop_exit@0: {
        // Loop initialization(s)
        STORE local.i: i32 i32(0)
        loop for@0: {
            // Loop condition
            t1: i32 = i32.le local.i local.limit
            t2: i32 = i32.eq t1 i32(0) // condition is false
            JMPIF loop_exit@0 t2
            block loop_body@0: {
                t3: i32 = i32.mod local.i i32(7)
                t4: i32 = i32.eq t3 i32(0)
                if t4 {
                    STORE local.j: i32 local.i
                    block loop_exit@1: {
                        loop for@1: {
                            // Loop condition
                            t5: i32 = add local.i i32(7)
                            t6: i32 = i32.lt j t5
                            t7: i32 = i32.eq t6 i32(0)
                            JMPIF loop_exit@1 t7
                            block loop_body@1: {
                                t8: i32 = add local.j local.i
                                t9: i32 = i32.mod t8 i32(12)
                                t10: i32 = i32.eq t9 i32(0)
                                if t10 {
                                    JMP loop_exit@1
                                }
                                PARAM: i32 t8
                                t11: ptr = CALL __str_cast_i32 1
                                PARAM: ptr t11
                                CALL __println 1
                                t12: i32 = add local.j i32(1)
                                STORE local.j: i32 t12
                                }
                                // repeat loop
                                JMP for@1
                        }
                    }
                }
                else {
                    t13 i32 = i32.mod local.i i32(2)
                    t14: i32 = i32.eq t13 i32(0)
                    if t14 {
                        JMP @loop_body0
                    }             
                    else {
                        PARAM: i32 local.i
                        t15: ptr = CALL __str_cast_i32 1
                        PARAM: ptr t15
                        CALL __println 1
                    }
                }
            }
            t16: i32 = i32.add local.i i32(1) 
            STORE local.i: i32 t16
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

