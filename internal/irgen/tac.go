package irgen

import (
	"fmt"
	"strings"

	"github.com/EladB1/The/internal/datatypes"
)

const indentDelim string = "  "

type (
	Datatype        string
	VariableScope   string
	Operation       string
	RuntimeFunction string
	/* interface for Instruction, Function, Loop, IfBlock, and Block */
	TAC interface {
		getTACType() string
	}
	Program struct {
		Code []TAC
	}
	Operand struct {
		Type     datatypes.IRType
		Var      Variable
		Constant any
		Label    string // use for JMP/JMPIF
	}
	Instruction struct {
		Destination Variable
		Operation   Operation
		Operand1    Operand
		Operand2    Operand
	}
	Parameter struct {
		Name string
		Type datatypes.IRType
	}
	Function struct {
		Name       string
		Parameters []Parameter
		ReturnType datatypes.IRType
		Code       []TAC
	}

	Variable struct {
		Name       string
		DataType   datatypes.IRType
		Visibility VariableScope
	}
	/* Used within loops to break/continue */
	Block struct {
		Label string
		Code  []TAC
	}
	Loop struct {
		Label string
		Code  []TAC
	}
	IfBlock struct {
		IfCondition Variable
		IfCode      *[]TAC
		ElseCode    *[]TAC
		// As of now else if will be an if embedded within an else block
	}
)

const (
	Local  VariableScope = "local"
	Global VariableScope = "global"
	Param  VariableScope = "param"
)

const (
	Store        Operation = "STORE"
	Get          Operation = "GET"
	Set          Operation = "Set"
	Return       Operation = "return"
	PrepareParam Operation = "PARAM"
	Call         Operation = "CALL"
	JMP          Operation = "JMP"
	JMPIF        Operation = "JMPIF"
	JMPIFNOT     Operation = "JMPIFNOT"
	Malloc       Operation = "Malloc"
	// int -> otherType
	I32ToI64 Operation = "i64.extend_i32_s"
	I32ToF32 Operation = "f32.convert_i32_s"
	I32ToF64 Operation = "f64.convert_i32_s"
	// int64 -> otherType
	I64ToI32 Operation = "i32.wrap_i64"
	I64ToF32 Operation = "i64.trunc_f32_s"
	I64ToF64 Operation = "i64.trunc_f64_s"
	// uint32 -> otherType
	U32ToI64 Operation = "i64.extend_i32_u"
	U32ToF32 Operation = "f32.convert_i32_u"
	U32ToF64 Operation = "f64.convert_i32_u"
	// uint64 -> otherType
	U64ToI32 Operation = "i32.wrap_i64"
	U64ToF32 Operation = "i64.trunc_f32_u"
	U64ToF64 Operation = "i64.trunc_f64_u"
	// float -> otherType
	F32ToI32 Operation = "i32.trunc_f32_s"
	F32ToU32 Operation = "i32.trunc_f32_u"
	F32ToI64 Operation = "i64.trunc_f32_s"
	F32ToU64 Operation = "i64.trunc_f32_u"
	F32ToF64 Operation = "f64.promote_f32"
	// double -> otherType
	F64ToI32 Operation = "i32.trunc_f64_s"
	F64ToU32 Operation = "i32.trunc_f64_u"
	F64ToI64 Operation = "i64.trunc_f64_s"
	F64ToU64 Operation = "i64.trunc_f64_u"
	F64ToF32 Operation = "f32.demote_f64"
	// TODO ptr operations
)

const (
	Stringlen            RuntimeFunction = "__str_length"
	StringIndex          RuntimeFunction = "__str_index"
	StringFromInt32      RuntimeFunction = "__str_fromInt32"
	StringFromInt64      RuntimeFunction = "__str_fromInt64"
	StringFromFloat32    RuntimeFunction = "__str_fromFloat32"
	StringFromFloat64    RuntimeFunction = "__str_fromFloat64"
	StringFromChar       RuntimeFunction = "__str_fromChar"
	StringFromBool       RuntimeFunction = "__str_fromBool"
	StringSlice          RuntimeFunction = "__str_slice"
	StringSliceInclusive RuntimeFunction = "__str_slice_inclusive"
	StringConcat         RuntimeFunction = "__str_concat"
	StringConcatChar     RuntimeFunction = "__str_concat_char"
	CharConcat           RuntimeFunction = "__char_concat"
	CharConcatString     RuntimeFunction = "__char_concat_str"
)

func typedOperation(irType datatypes.IRType, operation string) Operation {
	return Operation(fmt.Sprintf("%s.%s", string(irType), operation))
}

// TAC interface consumers

func (ins Instruction) getTACType() string {
	return "Instruction"
}

func (block IfBlock) getTACType() string {
	return "IfBlock"
}

func (block Block) getTACType() string {
	return "Block"
}

func (loop Loop) getTACType() string {
	return "Loop"
}

func (fn Function) getTACType() string {
	return "Function"
}

func (prog *Program) appendCode(code []TAC) {
	prog.Code = append(prog.Code, code...)
}

func (prog *Program) String() string {
	output := strings.Builder{}
	output.WriteString("Program: [")
	if len(prog.Code) > 0 {
		output.WriteRune('\n')
		output.WriteString(stringifyCode(prog.Code, 1))
	}
	output.WriteString("]\n")
	return output.String()
}

func (fn Function) String() string {
	output := strings.Builder{}
	output.WriteString(fmt.Sprintf("%sfn %s(", indentDelim, fn.Name))
	for i, param := range fn.Parameters {
		output.WriteString(fmt.Sprintf("%s: %s", param.Name, param.Type))
		if i != len(fn.Parameters)-1 {
			output.WriteString(", ")
		}
	}
	output.WriteRune(')')
	if fn.ReturnType != datatypes.NoneIR {
		output.WriteString(fmt.Sprintf("->%s", fn.ReturnType))
	}
	output.WriteString(" {")
	if len(fn.Code) != 0 {
		output.WriteString("\n")
	}
	output.WriteString(stringifyCode(fn.Code, 1))
	output.WriteString(indentDelim)
	output.WriteString("}\n")
	return output.String()
}

func (block IfBlock) String(indentLevel int) string {
	output := strings.Builder{}
	output.WriteString(strings.Repeat(indentDelim, indentLevel+1))
	output.WriteString("if ")
	output.WriteString(block.IfCondition.Name)
	output.WriteString(" {\n")
	output.WriteString(stringifyCode(*block.IfCode, indentLevel+1))
	output.WriteString(strings.Repeat(indentDelim, indentLevel+1))
	output.WriteString("}")
	if block.ElseCode != nil && len(*block.ElseCode) > 0 {
		output.WriteString(" else {\n")
		output.WriteString(stringifyCode(*block.ElseCode, indentLevel+1))
		output.WriteString(strings.Repeat(indentDelim, indentLevel+1))
		output.WriteString("}")
	}
	return output.String()
}

func (block Block) String(indentLevel int) string {
	output := strings.Builder{}
	output.WriteString(strings.Repeat(indentDelim, indentLevel+1))
	output.WriteString(block.Label)
	output.WriteString(": {\n")
	output.WriteString(stringifyCode(block.Code, indentLevel))
	output.WriteString(strings.Repeat(indentDelim, indentLevel+1))
	output.WriteString("}")
	return output.String()
}

func (loop Loop) String(indentLevel int) string {
	output := strings.Builder{}
	output.WriteString(strings.Repeat(indentDelim, indentLevel+1))
	output.WriteString(loop.Label)
	output.WriteString(": {\n")
	output.WriteString(stringifyCode(loop.Code, indentLevel))
	output.WriteString(strings.Repeat(indentDelim, indentLevel+1))
	output.WriteString("}")
	return output.String()
}

func (inst Instruction) String(indentLevel int) string {
	output := strings.Builder{}
	output.WriteString(strings.Repeat(indentDelim, indentLevel+1))
	if (inst.Destination != Variable{}) {
		output.WriteString(fmt.Sprintf("%s: %s = ", inst.Destination.Name, inst.Destination.DataType))
	}
	output.WriteString(string(inst.Operation))
	if (inst.Operand1 != Operand{}) {
		output.WriteString(inst.Operand1.String())
	}
	if (inst.Operand2 != Operand{}) {
		output.WriteString(inst.Operand2.String())
	}
	return output.String()
}

func (op Operand) String() string {
	output := strings.Builder{}
	if op.Label != "" {
		output.WriteRune(' ')
		output.WriteString(op.Label)
	} else if (op.Var != Variable{}) {
		vis := ""
		if op.Var.Visibility != "" {
			vis = fmt.Sprintf("%s.", op.Var.Visibility)
		}
		output.WriteString(fmt.Sprintf(" %s%s", vis, op.Var.Name))
	} else {
		if op.Type == "" {
			output.WriteString(fmt.Sprintf(" %v", op.Constant))
		} else {
			output.WriteString(fmt.Sprintf(" %s(%v)", op.Type, op.Constant))
		}
	}
	return output.String()
}

func stringifyCode(code []TAC, indentLevel int) string {
	output := strings.Builder{}
	for _, line := range code {
		switch line.getTACType() {
		case "Instruction":
			inst, ok := line.(Instruction)
			if !ok {
				break
			}
			output.WriteString(inst.String(indentLevel + 1))
		case "Function":
			fn, ok := line.(Function)
			if !ok {
				break
			}
			output.WriteString(fn.String())
		case "IfBlock":
			ifblock, ok := line.(IfBlock)
			if !ok {
				break
			}
			output.WriteString(ifblock.String(indentLevel + 1))
		case "Block":
			block, ok := line.(Block)
			if !ok {
				break
			}
			output.WriteString(block.String(indentLevel + 1))
		case "Loop":
			loop, ok := line.(Loop)
			if !ok {
				break
			}
			output.WriteString(loop.String(indentLevel + 1))
		}
		output.WriteRune('\n')
	}
	return output.String()
}
