package irgen

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
		Label    string // use for JMP/JMPIF
	}
	Instruction struct {
		Destination Variable
		Operation   string
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
