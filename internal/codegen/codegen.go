package codegen

import (
	"fmt"
	"strings"

	ds "github.com/EladB1/The/internal/datastructures"
	"github.com/EladB1/The/internal/irgen"
)

func Generate(filename string, ir irgen.Program, literals ds.LiteralPool) CompileTarget {
	target := CompileTarget{
		WatFilepath:     strings.Replace(filename, ".the", ".wat", 1),
		WasmFilepath:    strings.Replace(filename, ".the", ".wasm", 1),
		DataSection:     buildData(literals),
		GlobalVariables: getVariables(ir.Code, true),
		Functions:       getFunctions(ir),
	}
	return target
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
			offset += 4
		} else {
			offset += len(literal)
		}
	}
	return data
}

func getVariables(ir []irgen.TAC, inGlobalScope bool) []Variable {
	vars := []Variable{}
	for _, tac := range ir {
		if tac.GetTACType() != "Instruction" {
			continue
		}
		var vis WatVisibility = Local
		empty := irgen.Variable{}
		if instruction, ok := tac.(irgen.Instruction); ok {
			if instruction.Destination != empty {
				dest := instruction.Destination
				if inGlobalScope {
					vis = Global
				}
				vars = append(vars, Variable{
					Name:       dest.Name,
					DataType:   lowerIRTypeToWatType(dest.DataType),
					Visibility: vis,
				})
			}
			if instruction.Operation == irgen.Store {
				if instruction.Operand1.Var.Visibility == irgen.Global {
					vis = Global
				}

				variable := instruction.Operand1.Var
				vars = append(vars, Variable{
					Name:       variable.Name,
					DataType:   lowerIRTypeToWatType(variable.DataType),
					Value:      instruction.Operand2,
					Visibility: vis,
				})
			}
		}
	}
	return vars
}

func getFunctions(ir irgen.Program) []Function {
	fns := []Function{}
	for _, tac := range ir.Code {
		if tac.GetTACType() != "Function" {
			continue
		}
		if fn, ok := tac.(irgen.Function); ok {
			fns = append(fns, generateFunction(fn))
		}
	}
	return fns
}

func generateFunction(fn irgen.Function) Function {
	function := Function{
		Name:       fn.Name,
		ReturnType: lowerIRTypeToWatType(fn.ReturnType),
	}
	params := []Parameter{}
	for _, param := range fn.Parameters {
		params = append(params, Parameter{
			Name:     param.Name,
			DataType: lowerIRTypeToWatType(param.Type),
		})
	}
	function.Parameters = params
	function.LocalVariables = getVariables(fn.Code, false)
	function.Code = generateFunctionBody(fn.Code)
	return function
}

func generateFunctionBody(code []irgen.TAC) []Statement {
	statements := []Statement{}
	for _, tac := range code {
		if instruction, ok := tac.(irgen.Instruction); ok {
			if instruction.Operation == irgen.Return {
				statements = append(statements, translateOperand(instruction.Operand1))
				statements = append(statements, ControlInstruction{
					Operator: Return,
				})
			}
		}
	}
	return statements
}

func translateOperand(op irgen.Operand) Statement {
	if op.Constant != nil {
		return NumericInstruction{
			DataType: lowerIRTypeToWatType(op.Type),
			Operator: "const",
			Value:    op.Constant,
		}
	}
	return nil
}
