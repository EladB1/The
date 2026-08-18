package codegen

import (
	"fmt"
	"strings"

	ds "github.com/EladB1/The/internal/datastructures"
	"github.com/EladB1/The/internal/irgen"
)

func Generate(filename string, ir irgen.Program, literals ds.LiteralPool) CompileTarget {
	target := CompileTarget{
		WatFilepath:  strings.Replace(filename, ".the", ".wat", 1),
		WasmFilepath: strings.Replace(filename, ".the", ".wasm", 1),
		DataSection:  buildData(literals),
	}
	target.GlobalVariables = target.getVariables(ir.Code, true)
	target.Functions = target.getFunctions(ir)
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

func (target CompileTarget) getVariables(ir []irgen.TAC, inGlobalScope bool) map[string]Variable {
	vars := map[string]Variable{}
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
				if _, ok := vars[dest.Name]; !ok {
					vars[dest.Name] = Variable{
						Name:       dest.Name,
						DataType:   lowerIRTypeToWatType(dest.DataType),
						Visibility: vis,
					}
				}
			}
			if instruction.Operation == irgen.Store {
				if instruction.Operand1.Var.Visibility == irgen.Global {
					vis = Global
				}

				variable := instruction.Operand1.Var
				if _, ok := vars[variable.Name]; !ok {
					vars[variable.Name] = Variable{
						Name:       variable.Name,
						DataType:   lowerIRTypeToWatType(variable.DataType),
						Value:      target.translateOperand(instruction.Operand2),
						Visibility: vis,
					}
				}
			}
		}
	}
	return vars
}

func (target CompileTarget) getFunctions(ir irgen.Program) []Function {
	fns := []Function{}
	for _, tac := range ir.Code {
		if tac.GetTACType() != "Function" {
			continue
		}
		if fn, ok := tac.(irgen.Function); ok {
			fns = append(fns, target.generateFunction(fn))
		}
	}
	return fns
}

func (target CompileTarget) generateFunction(fn irgen.Function) Function {
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
	function.LocalVariables = target.getVariables(fn.Code, false)
	function.Code = target.generateFunctionBody(fn)
	return function
}

func (target CompileTarget) generateFunctionBody(fn irgen.Function) []Statement {
	statements := []Statement{}
	for _, tac := range fn.Code {
		if instruction, ok := tac.(irgen.Instruction); ok {
			statements = append(statements, target.handleInstruction(instruction)...)
		}
	}
	return statements
}
