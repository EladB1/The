package codegen

import (
	ds "github.com/EladB1/The/internal/datastructures"
	dt "github.com/EladB1/The/internal/datatypes"
)

const (
	dataStart = 100
)

type (
	Data struct {
		Name     string
		MemoryId int
		Offset   int
		Value    ds.Literal
	}

	Function struct {
		Name       string
		ReturnType dt.IRType
		Parameters []Parameter
	}

	Parameter struct {
		Name     string
		DataType dt.IRType
	}

	GlobalVariables struct {
		Name     string
		DataType dt.IRType
		Value    any
	}
)

func lowerIRTypeToWatType(irType dt.IRType) dt.IRType {
	if irType == dt.I32 || irType == dt.I64 || irType == dt.F32 || irType == dt.F64 {
		return irType
	}
	if irType == dt.Ptr || irType == dt.Str_const || irType == dt.U32 {
		return dt.I32
	}
	if irType == dt.U64 {
		return dt.I64
	}
	return dt.NoneIR
}
