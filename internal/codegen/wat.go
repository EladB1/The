package codegen

import (
	"fmt"
	"strings"

	ds "github.com/EladB1/The/internal/datastructures"
	dt "github.com/EladB1/The/internal/datatypes"
)

const (
	dataStart          = 100
	indentDelim string = "    "
)

type (
	WatVisibility string

	ControlOperator  string
	VariableOperator string
	MemoryOperator   string
	NumericOperator  string

	CompileTarget struct {
		WatFilepath     string
		WasmFilepath    string
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
		Value      Statement
	}

	Statement interface {
		GetStatementType() string
		String(int) string
	}

	ControlInstruction struct {
		Operator ControlOperator
	}

	VariableInstruction struct {
		Visibility WatVisibility
		Operator   VariableOperator
		Name       string
		Value      Statement
	}
	MemoryInstruction struct {
		DataType dt.IRType
		Operator MemoryOperator
		Address  Statement
		Value    Statement
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

func (inst VariableInstruction) GetStatementType() string {
	return "VariableInstruction"
}

func (inst MemoryInstruction) GetStatementType() string {
	return "MemoryInstruction"
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
	for _, glob := range target.GlobalVariables {
		output.WriteString(glob.String(1))
	}
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
	for _, local := range fn.LocalVariables {
		output.WriteString(local.String(2))
	}
	output.WriteString(stringifyBody(fn.Code, 1))
	output.WriteString(indentDelim)
	output.WriteString(")\n")
	return output.String()
}

func stringifyBody(body []Statement, indentLevel int) string {
	output := strings.Builder{}
	for _, stmt := range body {
		output.WriteString(stmt.String(indentLevel + 1))
		output.WriteString("\n")
	}
	return output.String()
}

func (inst MemoryInstruction) String(indentLevel int) string {
	output := strings.Builder{}
	output.WriteString("(")
	output.WriteString(fmt.Sprintf("%s.%s ", string(inst.DataType), inst.Operator))
	output.WriteString(inst.Address.String(0))
	output.WriteString(" ")
	output.WriteString(inst.Value.String(0))
	output.WriteString(")")
	return output.String()
}

func (inst VariableInstruction) String(indentLevel int) string {
	output := strings.Builder{}
	output.WriteString(strings.Repeat(indentDelim, indentLevel))
	output.WriteRune('(')
	output.WriteString(fmt.Sprintf("%s.%s $%s ", inst.Visibility, inst.Operator, inst.Name))
	output.WriteString(inst.Value.String(0))
	output.WriteString(")")
	return output.String()
}

func (inst NumericInstruction) String(indentLevel int) string {
	output := strings.Builder{}
	output.WriteString(strings.Repeat(indentDelim, indentLevel))
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
	output.WriteString(strings.Repeat(indentDelim, indentLevel))
	output.WriteRune('(')
	output.WriteString(string(inst.Operator)) // TODO: expand
	output.WriteString(")")
	return output.String()
}

func (variable Variable) String(indentLevel int) string {
	name := fmt.Sprintf("$%s", variable.Name)
	output := strings.Builder{}
	output.WriteString(strings.Repeat(indentDelim, indentLevel))

	output.WriteString("(")
	output.WriteString(string(variable.Visibility))
	output.WriteString(" ")
	output.WriteString(name)
	output.WriteString(" ")
	output.WriteString(string(variable.DataType))
	if variable.Visibility == Global {
		output.WriteString(" ")
		if val, ok := variable.Value.(NumericInstruction); ok {
			output.WriteString(val.String(0))
		}
		output.WriteString(")\n")
	} else {
		output.WriteString(")\n")
		// TODO: value
	}
	return output.String()
}
