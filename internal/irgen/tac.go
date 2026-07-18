package irgen

import (
	"fmt"
	"strings"
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
		Type     Datatype
		Var      Variable
		Constant any
		Unsigned bool
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
		Type Datatype
	}
	Function struct {
		Name       string
		Parameters []Parameter
		ReturnType Datatype
		Code       []TAC
	}

	Variable struct {
		Name       string
		DataType   Datatype
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
	i32       Datatype = "i32"
	i64       Datatype = "i64"
	f32       Datatype = "f32"
	f64       Datatype = "f64"
	str_const Datatype = "str_const"
	ptr       Datatype = "ptr"
	none      Datatype = "none"
)

const (
	Local  VariableScope = "local"
	Global VariableScope = "global"
	Param  VariableScope = "param"
)

const (
	Store        Operation = "STORE"
	Get          Operation = "GET"
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
	// TODO f32 operations
	// TODO f64 operations
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
	for _, line := range prog.Code {
		switch line.getTACType() {
		case "Instruction":
			inst, ok := line.(Instruction)
			if !ok {
				break
			}
			if (inst.Destination != Variable{}) {
				output.WriteString(fmt.Sprintf("%s: %s =", inst.Destination.Name, inst.Destination.DataType))
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
	return output.String()
}

func (op Operand) String() string {
	output := strings.Builder{}
	if op.Label != "" {
		output.WriteString(op.Label)
	}
	if (op.Var != Variable{}) {
		output.WriteString(fmt.Sprintf(" %s.%s", op.Var.Visibility, op.Var.Name))
	} else {
		output.WriteString(fmt.Sprintf(" %s(%v)", op.Type, op.Constant))
	}
	return output.String()
}
