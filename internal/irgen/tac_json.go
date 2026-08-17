package irgen

import (
	"encoding/json"
	"fmt"

	ds "github.com/EladB1/The/internal/datastructures"
	"github.com/EladB1/The/internal/datatypes"
)

type (
	SerializedProgram struct {
		Code []SerializedTAC
	}

	SerializedFunction struct {
		Name        string
		Parameters  []Parameter
		ReturnType  datatypes.IRType
		Code        []SerializedTAC
		SrcPosition ds.SourceLocation
	}

	SerializedBlock struct {
		Label string
		Code  []SerializedTAC
	}
	SerializedLoop struct {
		Label string
		Code  []SerializedTAC
	}
	SerializedIfBlock struct {
		IfCondition Variable
		IfCode      *[]SerializedTAC
		ElseCode    *[]SerializedTAC
		SrcPosition ds.SourceLocation
		// As of now else if will be an if embedded within an else block
	}

	SerializedTAC struct {
		Kind string
		Code json.RawMessage
	}
)

func (prog Program) MarshalJSON() ([]byte, error) {
	code := []SerializedTAC{}
	for _, tac := range prog.Code {
		instruction, err := serializeTAC(tac)
		if err != nil {
			return nil, err
		}
		code = append(code, instruction)
	}
	return json.Marshal(SerializedProgram{
		Code: code,
	})
}

func (fn Function) MarshalJSON() ([]byte, error) {
	code := []SerializedTAC{}
	for _, tac := range fn.Code {
		instruction, err := serializeTAC(tac)
		if err != nil {
			return nil, err
		}
		code = append(code, instruction)
	}
	return json.Marshal(SerializedFunction{
		Name:        fn.Name,
		Parameters:  fn.Parameters,
		ReturnType:  fn.ReturnType,
		Code:        code,
		SrcPosition: fn.SrcPosition,
	})
}

func (bl Block) MarshalJSON() ([]byte, error) {
	code := []SerializedTAC{}
	for _, tac := range bl.Code {
		instruction, err := serializeTAC(tac)
		if err != nil {
			return nil, err
		}
		code = append(code, instruction)
	}
	return json.Marshal(SerializedBlock{
		Label: bl.Label,
		Code:  code,
	})
}

func (bl IfBlock) MarshalJSON() ([]byte, error) {
	ifcode := []SerializedTAC{}
	elsecode := []SerializedTAC{}
	for _, tac := range *bl.IfCode {
		instruction, err := serializeTAC(tac)
		if err != nil {
			return nil, err
		}
		ifcode = append(ifcode, instruction)
	}
	for _, tac := range *bl.ElseCode {
		instruction, err := serializeTAC(tac)
		if err != nil {
			return nil, err
		}
		elsecode = append(elsecode, instruction)
	}
	return json.Marshal(SerializedIfBlock{
		IfCondition: bl.IfCondition,
		IfCode:      &ifcode,
		ElseCode:    &elsecode,
		SrcPosition: bl.SrcPosition,
	})
}

func (loop Loop) MarshalJSON() ([]byte, error) {
	code := []SerializedTAC{}
	for _, tac := range loop.Code {
		instruction, err := serializeTAC(tac)
		if err != nil {
			return nil, err
		}
		code = append(code, instruction)
	}
	return json.Marshal(SerializedLoop{
		Label: loop.Label,
		Code:  code,
	})
}

func serializeTAC(tac TAC) (SerializedTAC, error) {
	code, err := json.Marshal(tac)
	if err != nil {
		return SerializedTAC{}, err
	}
	return SerializedTAC{
		Kind: tac.GetTACType(),
		Code: code,
	}, nil
}

func (prog *Program) UnmarshalJSON(data []byte) error {
	var serialized SerializedProgram
	if err := json.Unmarshal(data, &serialized); err != nil {
		return err
	}
	var err error
	*prog, err = serialized.transform()
	return err
}

func transformCodeBlock(serialized []SerializedTAC) ([]TAC, error) {
	tac := []TAC{}
	for _, code := range serialized {
		switch code.Kind {
		case "Instruction":
			var instruction Instruction
			if err := json.Unmarshal(code.Code, &instruction); err != nil {
				return tac, err
			}
			tac = append(tac, instruction)
		case "Function":
			var instruction SerializedFunction
			if err := json.Unmarshal(code.Code, &instruction); err != nil {
				return tac, err
			}
			fn, err := instruction.transform()
			if err != nil {
				return tac, err
			}
			tac = append(tac, fn)
		case "IfBlock":
			var instruction SerializedIfBlock
			if err := json.Unmarshal(code.Code, &instruction); err != nil {
				return tac, err
			}
			block, err := instruction.transform()
			if err != nil {
				return tac, err
			}
			tac = append(tac, block)
		case "Block":
			var instruction SerializedBlock
			if err := json.Unmarshal(code.Code, &instruction); err != nil {
				return tac, err
			}
			block, err := instruction.transform()
			if err != nil {
				return tac, err
			}
			tac = append(tac, block)
		case "Loop":
			var instruction SerializedLoop
			if err := json.Unmarshal(code.Code, &instruction); err != nil {
				return tac, err
			}
			loop, err := instruction.transform()
			if err != nil {
				return tac, err
			}
			tac = append(tac, loop)
		default:
			return tac, fmt.Errorf("Unknown TAC type: '%s'", code.Kind)
		}
	}
	return tac, nil
}

func (serialized *SerializedProgram) transform() (Program, error) {
	prog := Program{}
	var err error
	prog.Code, err = transformCodeBlock(serialized.Code)
	return prog, err
}

func (serialized *SerializedFunction) transform() (Function, error) {
	fn := Function{
		Name:        serialized.Name,
		Parameters:  serialized.Parameters,
		ReturnType:  serialized.ReturnType,
		SrcPosition: serialized.SrcPosition,
	}
	code, err := transformCodeBlock(serialized.Code)
	fn.Code = code
	return fn, err
}

func (serialized *SerializedBlock) transform() (Block, error) {
	var err error
	block := Block{
		Label: serialized.Label,
	}
	block.Code, err = transformCodeBlock(serialized.Code)
	return block, err
}

func (serialized *SerializedIfBlock) transform() (IfBlock, error) {
	var err error
	block := IfBlock{
		IfCondition: serialized.IfCondition,
		SrcPosition: serialized.SrcPosition,
	}
	ifcode, err := transformCodeBlock(*serialized.IfCode)
	if err != nil {
		return block, err
	}
	block.IfCode = &ifcode
	elsecode, err := transformCodeBlock(*serialized.ElseCode)
	block.ElseCode = &elsecode
	return block, err
}

func (serialized *SerializedLoop) transform() (Loop, error) {
	var err error
	loop := Loop{
		Label: serialized.Label,
	}
	loop.Code, err = transformCodeBlock(serialized.Code)
	return loop, err
}
