package irgen

import (
	"fmt"
	"strings"

	ds "github.com/EladB1/The/internal/datastructures"
	"github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/semantic"
)

const indentDelim string = "  "

type (
	Datatype        string
	VariableScope   string
	Operation       string
	RuntimeFunction string
	/* interface for Instruction, Function, Loop, IfBlock, and Block */
	TAC interface {
		GetTACType() string
	}
	Program struct {
		Code []TAC
	}
	Operand struct {
		Type        datatypes.IRType
		Var         Variable
		Offset      semantic.OffsetValue
		Constant    any
		Label       string // use for JMP/JMPIF
		SrcPosition ds.SourceLocation
	}
	Instruction struct {
		Destination Variable
		Operation   Operation
		Operand1    Operand
		Operand2    Operand
		SrcPosition ds.SourceLocation
	}
	Parameter struct {
		Name        string
		Type        datatypes.IRType
		SrcPosition ds.SourceLocation
	}
	Function struct {
		Name        string
		Parameters  []Parameter
		ReturnType  datatypes.IRType
		Code        []TAC
		SrcPosition ds.SourceLocation
	}

	Variable struct {
		Name        string
		DataType    datatypes.IRType
		Visibility  VariableScope
		SrcPosition ds.SourceLocation
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
		SrcPosition ds.SourceLocation
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
	Set          Operation = "SET"
	Load         Operation = "LOAD"
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

func (ins Instruction) GetTACType() string {
	return "Instruction"
}

func (block IfBlock) GetTACType() string {
	return "IfBlock"
}

func (block Block) GetTACType() string {
	return "Block"
}

func (loop Loop) GetTACType() string {
	return "Loop"
}

func (fn Function) GetTACType() string {
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
	output.WriteString(strings.Repeat(indentDelim, indentLevel))
	output.WriteString("if ")
	output.WriteString(block.IfCondition.Name)
	output.WriteString(" {\n")
	output.WriteString(stringifyCode(*block.IfCode, indentLevel+1))
	output.WriteString(strings.Repeat(indentDelim, indentLevel))
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
	output.WriteString(strings.Repeat(indentDelim, indentLevel))
	output.WriteString(block.Label)
	output.WriteString(": {\n")
	output.WriteString(stringifyCode(block.Code, indentLevel+1))
	output.WriteString(strings.Repeat(indentDelim, indentLevel))
	output.WriteString("}")
	return output.String()
}

func (loop Loop) String(indentLevel int) string {
	output := strings.Builder{}
	output.WriteString(strings.Repeat(indentDelim, indentLevel))
	output.WriteString(loop.Label)
	output.WriteString(": {\n")
	output.WriteString(stringifyCode(loop.Code, indentLevel+1))
	output.WriteString(strings.Repeat(indentDelim, indentLevel))
	output.WriteString("}")
	return output.String()
}

func (inst Instruction) String(indentLevel int) string {
	output := strings.Builder{}
	output.WriteString(strings.Repeat(indentDelim, indentLevel))
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
		off := ""
		vis := ""
		if op.Var.Visibility != "" {
			vis = fmt.Sprintf("%s.", op.Var.Visibility)
		}
		if op.Offset.IsSet {
			off = fmt.Sprintf("+%d", op.Offset.Value)
		}
		output.WriteString(fmt.Sprintf(" %s%s%s", vis, op.Var.Name, off))
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
		switch line.GetTACType() {
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
