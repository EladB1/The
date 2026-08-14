package codegen

import (
	"fmt"
	"strings"

	ds "github.com/EladB1/The/internal/datastructures"
	"github.com/EladB1/The/internal/irgen"
)

func Generate(filename string, ir irgen.Program, literals ds.LiteralPool) {
	wat_name := strings.Replace(filename, ".the", ".wat", 1)
	fmt.Println(wat_name, literals)
	fmt.Println(buildData(literals))
	fmt.Println(getGlobalVariables(ir))
	// Create data from literalPool
	// Create functions and variables from ir code
}

func buildData(literals ds.LiteralPool) []Data {
	data := []Data{}
	offset := dataStart
	for i, literal := range literals {
		data = append(data, Data{
			Name:     fmt.Sprintf("$__str_const%d", i),
			MemoryId: 0,
			Offset:   offset,
			Value:    literal,
		})
		if len(literal) == 0 {
			offset++
		} else {
			offset += len(literal)
		}
	}
	return data
}

func getGlobalVariables(ir irgen.Program) []GlobalVariables {
	vars := []GlobalVariables{}
	for _, tac := range ir.Code {
		if tac.GetTACType() != "Instruction" {
			continue
		}
		if instruction, ok := tac.(irgen.Instruction); ok {
			if instruction.Operation == irgen.Store && instruction.Operand1.Var.Visibility == irgen.Global {
				glob := instruction.Operand1.Var
				vars = append(vars, GlobalVariables{
					Name:     glob.Name,
					DataType: lowerIRTypeToWatType(glob.DataType),
					Value:    instruction.Operand2,
				})
			}
		}
	}
	return vars
}
