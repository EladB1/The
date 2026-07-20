package irgen

import (
	"fmt"
	"strings"

	"github.com/EladB1/The/internal/datatypes"
)

type (
	Datatype      string
	VariableScope string
	Operation     string
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
		IfCode      []TAC
		ElseCode    []TAC
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
	Malloc       Operation = "Malloc"
	Addi32       Operation = "i32.add"
	Subi32       Operation = "i32.sub"
	Muli32       Operation = "i32.mul"
	Divi32       Operation = "i32.div"
	Modi32       Operation = "i32.mod"
	EQi32        Operation = "i32.eq"
	NEi32        Operation = "i32.ne"
	LTi32        Operation = "i32.lt"
	LEi32        Operation = "i32.le"
	GTi32        Operation = "i32.gt"
	GEi32        Operation = "i32.ge"
	LShifti32    Operation = "i32.lshift"
	RShifti32    Operation = "i32.rshift"
	XORi32       Operation = "i32.xor"
	ORi32        Operation = "i32.or"
	ANDi32       Operation = "i32.and"
	// TODO Handle unsigned operations (check WAT supported operations)
	// TODO i64 operations
	Addi64    Operation = "i64.add"
	Subi64    Operation = "i64.sub"
	Muli64    Operation = "i64.mul"
	Divi64    Operation = "i64.div"
	Modi64    Operation = "i64.mod"
	EQi64     Operation = "i64.eq"
	NEi64     Operation = "i64.ne"
	LTi64     Operation = "i64.lt"
	LEi64     Operation = "i64.le"
	GTi64     Operation = "i64.gt"
	GEi64     Operation = "i64.ge"
	LShifti64 Operation = "i64.lshift"
	RShifti64 Operation = "i64.rshift"
	XORi64    Operation = "i64.xor"
	ORi64     Operation = "i64.or"
	ANDi64    Operation = "i32.and"
	// TODO f32 operations
	Addf32    Operation = "f32.add"
	Subf32    Operation = "f32.sub"
	Mulf32    Operation = "f32.mul"
	Divf32    Operation = "f32.div"
	Modf32    Operation = "f32.mod"
	EQf32     Operation = "f32.eq"
	NEf32     Operation = "f32.ne"
	LTf32     Operation = "f32.lt"
	LEf32     Operation = "f32.le"
	GTf32     Operation = "f32.gt"
	GEf32     Operation = "f32.ge"
	LShiftf32 Operation = "f32.lshift"
	RShiftf32 Operation = "f32.rshift"
	XORf32    Operation = "f32.xor"
	ORf32     Operation = "f32.or"
	ANDf32    Operation = "i32.and"
	// TODO f64 operations
	Addf64    Operation = "f64.add"
	Subf64    Operation = "f64.sub"
	Mulf64    Operation = "f64.mul"
	Divf64    Operation = "f64.div"
	Modf64    Operation = "f64.mod"
	EQf64     Operation = "f64.eq"
	NEf64     Operation = "f64.ne"
	LTf64     Operation = "f64.lt"
	LEf64     Operation = "f64.le"
	GTf64     Operation = "f64.gt"
	GEf64     Operation = "f64.ge"
	LShiftf64 Operation = "f64.lshift"
	RShiftf64 Operation = "f64.rshift"
	XORf64    Operation = "f64.xor"
	ORf64     Operation = "f64.or"
	ANDf64    Operation = "f64.and"
	// TODO str_const operations
	// TODO ptr operations
	// TODO runtime library functions/constants
)

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
	output.WriteString("Program: [\n")
	for _, line := range prog.Code {
		output.WriteRune('\t')
		switch line.getTACType() {
		case "Instruction":
			inst, ok := line.(Instruction)
			if !ok {
				break
			}
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

		case "IfBlock":
			//
		case "Block":
			//
		case "Loop":
			//
		}
		output.WriteRune('\n')
	}
	output.WriteString("]\n")
	return output.String()
}

func (op Operand) String() string {
	output := strings.Builder{}
	if op.Label != "" {
		output.WriteString(op.Label)
	}
	if (op.Var != Variable{}) {
		vis := ""
		if op.Var.Visibility != "" {
			vis = fmt.Sprintf("%s.", op.Var.Visibility)
		}
		output.WriteString(fmt.Sprintf(" %s%s", vis, op.Var.Name))
	} else {
		output.WriteString(fmt.Sprintf(" %s(%v)", op.Type, op.Constant))
	}
	return output.String()
}
