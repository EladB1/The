package codegen

import (
	"github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/irgen"
)

func (target CompileTarget) handleInstruction(instruction irgen.Instruction) []Statement {
	statements := []Statement{}
	switch instruction.Operation {
	case irgen.Return:
		statements = target.generateReturn(instruction)
	case irgen.Store:
		statements = target.generateStoreOperation(instruction)
	case irgen.PrepareParam:
		statements = append(statements, target.translateOperand(instruction.Operand1))
	case irgen.Call:
		statements = append(statements, target.generateCall(instruction)...)
	}
	return statements
}

func (target CompileTarget) generateReturn(instruction irgen.Instruction) []Statement {
	statements := []Statement{}
	statements = append(statements, target.translateOperand(instruction.Operand1))
	statements = append(statements, ControlInstruction{
		Operator: Return,
	})
	return statements
}

func (target CompileTarget) generateStoreOperation(instruction irgen.Instruction) []Statement {
	statements := []Statement{}
	if instruction.Operand1.Var.Visibility == irgen.Global {
		statements = append(statements, MemoryInstruction{
			DataType: lowerIRTypeToWatType(instruction.Operand1.Var.DataType),
			Operator: Store,
			Address: VariableInstruction{
				Visibility: Global,
				Operator:   Get,
				Name:       instruction.Operand1.Var.Name,
			},
			Value: target.translateOperand(instruction.Operand2),
		})
	} else {
		statements = append(statements, VariableInstruction{
			Visibility: Local,
			Operator:   Set,
			Name:       instruction.Operand1.Var.Name,
			Value:      target.translateOperand(instruction.Operand2),
		})
	}
	return statements
}

func (target CompileTarget) generateCall(instruction irgen.Instruction) []Statement {
	statements := []Statement{}
	if name, ok := instruction.Operand1.Constant.(string); !ok {
		// TODO
	} else {
		statements = append(statements, ControlInstruction{
			Operator:   Call,
			Identifier: name,
		},
		)
	}
	return statements
}

func (target CompileTarget) translateOperand(op irgen.Operand) Statement {
	if op.Constant != nil {
		if op.Type == datatypes.Str_const {
			if index, ok := op.Constant.(int); ok {
				return NumericInstruction{
					DataType: datatypes.I32,
					Operator: "const",
					Value:    target.DataSection[index].Offset,
				}
			}
		} else {
			return NumericInstruction{
				DataType: lowerIRTypeToWatType(op.Type),
				Operator: "const",
				Value:    op.Constant,
			}
		}
	}
	return nil
}
