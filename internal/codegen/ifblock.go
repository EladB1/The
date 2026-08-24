package codegen

import (
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
	return block
}
