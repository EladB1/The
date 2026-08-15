package codegen

import (
	"fmt"
	"strings"

	ds "github.com/EladB1/The/internal/datastructures"
	dt "github.com/EladB1/The/internal/datatypes"
)

const (
	dataStart          = 100
	indentDelim string = "  "
)

type (
	WatVisibility string

	ControlOperator  string
	VariableOperator string
	MemoryOperator   string
	NumericOperator  string

	CompileTarget struct {
		Filepath        string
		DataSection     []Data
		GlobalVariables []Variable
		Functions       []Function
	}

	Data struct {
		Name     string
		MemoryId int
		Offset   int
		Value    ds.Literal
	}

	Function struct {
		Name           string
		ReturnType     dt.IRType
		Parameters     []Parameter
		LocalVariables []Variable
		Code           []Statement
	}

	Parameter struct {
		Name     string
		DataType dt.IRType
	}

	Variable struct {
		Name       string
		DataType   dt.IRType
		Visibility WatVisibility
		Value      any
	}

	Statement interface {
		GetStatementType() string
	}

	ControlInstruction struct {
		Operator ControlOperator
	}

	VariableInstruction struct {
		Visibility WatVisibility
		Operator   VariableOperator
	}
	MemoryInstruction struct {
		DataType    dt.IRType
		MemoryValue string
		Operator    MemoryOperator
	}
	NumericInstruction struct {
		DataType dt.IRType
		Operator string
		Value    any
	}
)

const (
	Local  WatVisibility = "local"
	Global WatVisibility = "global"
)

const (
	BR     ControlOperator = "br"
	BRIF   ControlOperator = "br_if"
	Throw  ControlOperator = "throw"
	Call   ControlOperator = "call"
	Drop   ControlOperator = "drop"
	Return ControlOperator = "return"
	// TODO: add if/else, loops, blocks
)

const (
	Get VariableOperator = "get"
	Set VariableOperator = "set"
	Tee VariableOperator = "tee"
)

const (
	Load     MemoryOperator = "load"
	Load8_s  MemoryOperator = "load8_s"
	Load8_u  MemoryOperator = "load8_u"
	Load16_s MemoryOperator = "load16_s"
	Load16_u MemoryOperator = "load16_u"
	Load32_s MemoryOperator = "load32_s"
	Load32_u MemoryOperator = "load32_u"
	Store    MemoryOperator = "store"
	Store8   MemoryOperator = "store8"
	Store16  MemoryOperator = "store16"
	Store32  MemoryOperator = "store32"
	Size     MemoryOperator = "size"
	Grow     MemoryOperator = "grow"
	Init     MemoryOperator = "init"
	MemDrop  MemoryOperator = "drop"
	Copy     MemoryOperator = "copy"
	Fill     MemoryOperator = "fill"
)

// const (
// 	Const NumericOperator = "const"
// 	Eqz
// 	Eq
// 	Ne
// 	Lt_s
// 	Lt_u
// 	Gt_s
// 	Gt_u
// 	Le_s
// 	Le_u
// 	Ge_s
// 	Ge_u
// 	Lt
// 	Le
// 	Gt
// 	Ge
// )

func lowerIRTypeToWatType(irType dt.IRType) dt.IRType {
	if irType == dt.I32 || irType == dt.I64 || irType == dt.F32 || irType == dt.F64 {
		return irType
	}
	if irType == dt.Ptr || irType == dt.Str_const || irType == dt.U32 {
		return dt.I32
	}
	if irType == dt.U64 {
		return dt.I64
	}
	return dt.NoneIR
}

// Statement interface
func (inst ControlInstruction) GetStatementType() string {
	return "ControlInstruction"
}

func (inst NumericInstruction) GetStatementType() string {
	return "NumericInstruction"
}

// String representations
func (target CompileTarget) String() string {
	output := strings.Builder{}
	for _, data := range target.DataSection {
		output.WriteString(indentDelim)
		output.WriteString("(data ")
		output.WriteString(data.Name)
		output.WriteString(" ")
		output.WriteString(NumericInstruction{
			DataType: dt.I32,
			Operator: "const",
			Value:    data.Offset,
		}.String(0))
		output.WriteString(" ")
		output.WriteString(fmt.Sprintf("\"%s\")\n", string(data.Value)))
	}
	// for _, glob := range target.GlobalVariables {
	// 	output.WriteString(glob.String())
	// }
	for _, fn := range target.Functions {
		output.WriteString(fn.String())
	}
	return output.String()
}

func (fn Function) String() string {
	output := strings.Builder{}
	output.WriteString(indentDelim)
	output.WriteString("(func $")
	output.WriteString(fn.Name)
	if len(fn.Parameters) > 0 {
		for _, param := range fn.Parameters {
			output.WriteString(fmt.Sprintf(" (param $%s %s)", param.Name, param.DataType))
		}
	}
	if fn.ReturnType != dt.NoneIR {
		output.WriteString(fmt.Sprintf(" (result %s)", fn.ReturnType))
	}
	output.WriteString("\n")
	output.WriteString(stringifyBody(fn.Code, 1))
	output.WriteString(indentDelim)
	output.WriteString(")\n")
	return output.String()
}

func stringifyBody(body []Statement, indentLevel int) string {
	output := strings.Builder{}
	for _, stmt := range body {
		switch stmt.GetStatementType() {
		case "NumericInstruction":
			numeric, ok := stmt.(NumericInstruction)
			if !ok {
				break
			}
			output.WriteString(numeric.String(indentLevel + 1))
		case "ControlInstruction":
			control, ok := stmt.(ControlInstruction)
			if !ok {
				break
			}
			output.WriteString(control.String(indentLevel + 1))
		}
		output.WriteString("\n")
	}
	return output.String()
}

func (inst NumericInstruction) String(indentLevel int) string {
	output := strings.Builder{}
	output.WriteString(strings.Repeat(indentDelim, indentLevel+1))
	output.WriteRune('(')
	output.WriteString(fmt.Sprintf("%s.%s", inst.DataType, inst.Operator))
	if inst.Value != nil {
		output.WriteString(fmt.Sprintf(" %v", inst.Value))
	}
	output.WriteString(")")
	return output.String()
}

func (inst ControlInstruction) String(indentLevel int) string {
	output := strings.Builder{}
	output.WriteString(strings.Repeat(indentDelim, indentLevel+1))
	output.WriteRune('(')
	output.WriteString(string(inst.Operator)) // TODO: expand
	output.WriteString(")")
	return output.String()
}
