package codegen

import (
	"slices"
	"strings"

	"github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/irgen"
)

func handleInstruction(instruction irgen.Instruction) []Statement {
	statements := []Statement{}
	switch instruction.Operation {
	case irgen.Return:
		statements = generateReturn(instruction)
	case irgen.Store:
		statements = generateStoreOperation(instruction)
	case irgen.PrepareParam:
		statements = append(statements, translateOperand(instruction.Operand1))
	case irgen.Call:
		statements = generateCall(instruction)
	default:
		operation := instruction.Operation
		if isTypeCast(operation) {
			statements = append(statements, generateTypecast(instruction)...)
		} else if op := getNameIfTypedOperation(operation); op != "" {
			switch op {
			case "add", "sub", "mul", "eq", "ne", "xor", "and", "or", "shl":
				statements = generateBinarySignAgnosticTypedOperation(instruction, op)
			case "div", "rem", "lt", "le", "gt", "ge", "shr":
				statements = append(statements, generateBinaryTypedOperation(instruction, op)...)
			}
		}
	}
	return statements
}

func generateReturn(instruction irgen.Instruction) []Statement {
	statements := []Statement{}
	statements = append(statements, translateOperand(instruction.Operand1))
	statements = append(statements, ControlInstruction{
		Operator: Return,
	})
	return statements
}

func generateStoreOperation(instruction irgen.Instruction) []Statement {
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
			Value: translateOperand(instruction.Operand2),
		})
	} else {
		statements = append(statements, VariableInstruction{
			Visibility: Local,
			Operator:   Set,
			Name:       instruction.Operand1.Var.Name,
			Value:      translateOperand(instruction.Operand2),
		})
	}
	return statements
}

func generateCall(instruction irgen.Instruction) []Statement {
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
	if hasDestination(instruction) {
		statements = append(statements, VariableInstruction{
			Visibility: Local,
			Operator:   Set,
			Name:       instruction.Destination.Name,
		})
	}
	return statements
}

func generateBinarySignAgnosticTypedOperation(instruction irgen.Instruction, operator string) []Statement {
	statements := []Statement{
		translateOperand(instruction.Operand1),
		translateOperand(instruction.Operand2),
		NumericInstruction{
			DataType: lowerIRTypeToWatType(instruction.Operand1.Type),
			Operator: operator,
		},
	}
	if hasDestination(instruction) {
		statements = append(statements, VariableInstruction{
			Visibility: Local,
			Operator:   Set,
			Name:       instruction.Destination.Name,
		})
	}
	return statements
}

func generateBinaryTypedOperation(instruction irgen.Instruction, operator string) []Statement {
	suffix := "_s"
	statements := []Statement{
		translateOperand(instruction.Operand1),
		translateOperand(instruction.Operand2),
	}
	if instruction.Operand1.Type == datatypes.U32 || instruction.Operand1.Type == datatypes.U64 {
		suffix = "_u"
	}
	statements = append(statements, NumericInstruction{
		DataType: lowerIRTypeToWatType(instruction.Operand1.Type),
		Operator: operator + suffix,
	})
	if hasDestination(instruction) {
		statements = append(statements, VariableInstruction{
			Visibility: Local,
			Operator:   Set,
			Name:       instruction.Destination.Name,
		})
	}
	return statements
}

func generateTypecast(instruction irgen.Instruction) []Statement {
	statements := []Statement{
		translateOperand(instruction.Operand1),
		NumericInstruction{
			DataType: lowerIRTypeToWatType(instruction.Destination.DataType),
			Operator: string(instruction.Operation),
		},
		VariableInstruction{
			Visibility: Local,
			Operator:   Set,
			Name:       instruction.Destination.Name,
		},
	}
	return statements
}

func translateOperand(op irgen.Operand) Statement {
	emptyVar := irgen.Variable{}
	if op.Constant != nil {
		if op.Type == datatypes.Str_const {
			if index, ok := op.Constant.(int); ok {
				return NumericInstruction{
					DataType: datatypes.I32,
					Operator: "const",
					Value:    data[index].Offset,
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
	if op.Var != emptyVar {
		vis := Local
		if op.Var.Visibility == irgen.Global {
			vis = Global
		}
		return VariableInstruction{
			Visibility: vis,
			Operator:   Get,
			Name:       op.Var.Name,
		}
	}
	return nil
}

func isTypeCast(operation irgen.Operation) bool {
	typecasts := []irgen.Operation{
		// int -> otherType
		irgen.I32ToI64,
		irgen.I32ToF32,
		irgen.I32ToF64,
		// int64 -> otherType
		irgen.I64ToI32,
		irgen.I64ToF32,
		irgen.I64ToF64,
		// uint32 -> otherType
		irgen.U32ToI64,
		irgen.U32ToF32,
		irgen.U32ToF64,
		// uint64 -> otherType
		irgen.U64ToI32,
		irgen.U64ToF32,
		irgen.U64ToF64,
		// float -> otherType
		irgen.F32ToI32,
		irgen.F32ToU32,
		irgen.F32ToI64,
		irgen.F32ToU64,
		irgen.F32ToF64,
		// double -> otherType
		irgen.F64ToI32,
		irgen.F64ToU32,
		irgen.F64ToI64,
		irgen.F64ToU64,
		irgen.F64ToF32,
	}
	return slices.Contains(typecasts, operation)
}

func getNameIfTypedOperation(operation irgen.Operation) string {
	if !strings.Contains(string(operation), ".") {
		return ""
	}
	parts := strings.Split(string(operation), ".")
	return parts[1]
}

func hasDestination(instruction irgen.Instruction) bool {
	emptyVar := irgen.Variable{}
	return instruction.Destination != emptyVar
}
