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
	// TODO str_const operations
	// TODO ptr operations
	// TODO runtime library functions/constants
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
