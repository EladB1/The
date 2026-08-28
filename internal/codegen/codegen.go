package codegen

import (
	"fmt"
	"strings"

	ds "github.com/EladB1/The/internal/datastructures"
	"github.com/EladB1/The/internal/irgen"
)

var (
	data []Data
)

func Generate(filename string, ir irgen.Program, literals ds.LiteralPool) CompileTarget {
	data = []Data{}
	target := CompileTarget{
		WatFilepath:  strings.Replace(filename, ".the", ".wat", 1),
		WasmFilepath: strings.Replace(filename, ".the", ".wasm", 1),
		DataSection:  buildData(literals),
	}
	data = target.DataSection
	target.GlobalVariables = getVariables(ir.Code, true)
	target.Functions = getFunctions(ir)
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
		if literal.Length == 0 {
			offset += 5
		} else {
			offset += literal.Length + 5
		}
	}
	return data
}

func getVariables(ir []irgen.TAC, inGlobalScope bool) *ds.OrderedMap[Variable] {
	vars := ds.NewOrderedMap[Variable]()
	for _, tac := range ir {
		var vis WatVisibility = Local
		empty := irgen.Variable{}
		if instruction, ok := tac.(irgen.Instruction); ok {
			if instruction.Destination != empty {
				dest := instruction.Destination
				if inGlobalScope {
					vis = Global
				}
				if _, ok := vars.Lookup(dest.Name); !ok {
					vars.Add(Variable{
						Name:       dest.Name,
						DataType:   lowerIRTypeToWatType(dest.DataType),
						Visibility: vis,
					}, dest.Name)
				}
			}
			if instruction.Operation == irgen.Store {
				if instruction.Operand1.Var.Visibility == irgen.Global {
					vis = Global
				}

				variable := instruction.Operand1.Var
				if _, ok := vars.Lookup(variable.Name); !ok {
					vars.Add(Variable{
						Name:       variable.Name,
						DataType:   lowerIRTypeToWatType(variable.DataType),
						Value:      translateOperand(instruction.Operand2),
						Visibility: vis,
					}, variable.Name)
				}
			}
		} else if ifblock, ok := tac.(irgen.IfBlock); ok {
			vars.AddAll(getVariables(*ifblock.IfCode, false))
			if ifblock.ElseCode != nil {
				vars.AddAll(getVariables(*ifblock.ElseCode, false))
			}
		} else if block, ok := tac.(irgen.Block); ok {
			vars.AddAll(getVariables(block.Code, false))
		} else if loop, ok := tac.(irgen.Loop); ok {
			vars.AddAll(getVariables(loop.Code, false))
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
		Export:     fn.Name == "main",
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
	function.Code = generateBody(fn.Code)
	return function
}

func generateBody(body []irgen.TAC) []Statement {
	statements := []Statement{}
	for _, tac := range body {
		if instruction, ok := tac.(irgen.Instruction); ok {
			statements = append(statements, handleInstruction(instruction)...)
		} else if ifblock, ok := tac.(irgen.IfBlock); ok {
			statements = append(statements, handleIfBlock(ifblock))
		} else if block, ok := tac.(irgen.Block); ok {
			statements = append(statements, handleBlock(block))
		} else if loop, ok := tac.(irgen.Loop); ok {
			statements = append(statements, handleLoop(loop))
		}
	}
	return statements
}
