package codegen

import (
	dt "github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/irgen"
)

func handleIfBlock(ifblock irgen.IfBlock) IfBlock {
	block := IfBlock{
		IfCode:   generateBody(*ifblock.IfCode),
		ElseCode: generateBody(*ifblock.ElseCode),
	}
	variable := ifblock.IfCondition
	vis := Local
	if variable.Visibility == irgen.Global {
		vis = Global
	}
	block.IfCondition = VariableInstruction{
		Visibility: vis,
		Operator:   Get,
		Name:       variable.Name,
	}
	block.ReturnType = lowerIRTypeToWatType(getBlockReturnType(ifblock))
	return block
}

func getBranchReturnType(branch []irgen.TAC) dt.IRType {
	returnType := dt.NoneIR
	emptyOperand := irgen.Operand{}
	for i, instruction := range branch {
		if inst, ok := instruction.(irgen.Instruction); ok {
			if inst.Operation == irgen.Return {
				if inst.Operand1 != emptyOperand {
					returnType = inst.Operand1.Type
				}
			}
		} else if ifblock := instruction.(irgen.IfBlock); ok {
			if i == len(branch)-1 {
				returnType = getBlockReturnType(ifblock)
			}
		}
	}
	return returnType
}

func getBlockReturnType(ifblock irgen.IfBlock) dt.IRType {
	returnType := dt.NoneIR
	if ifblock.ElseCode == nil || len(*ifblock.ElseCode) == 0 {
		return returnType
	}
	returnType = getBranchReturnType(*ifblock.IfCode)
	if returnType == dt.NoneIR {
		return returnType
	}
	secondType := getBranchReturnType(*ifblock.ElseCode)
	if secondType != returnType {
		return dt.NoneIR
	}
	return returnType
}
