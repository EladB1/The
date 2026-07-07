package datatypes

import "strings"

type DataType interface {
	String() string
	IsPrimitive() bool
	GetSizeInBytes() int
}

type PrimitiveType string
type DynamicType string

const (
	Int32    PrimitiveType = "int"
	Int64    PrimitiveType = "int64"
	Uint32   PrimitiveType = "uint32"
	Uint64   PrimitiveType = "uint64"
	Float    PrimitiveType = "float"
	Double   PrimitiveType = "double"
	Bool     PrimitiveType = "bool"
	Char     PrimitiveType = "char"
	String   PrimitiveType = "String"
	Ref      PrimitiveType = "Ref"
	ScopeRef PrimitiveType = "ScopeRef"
	None     PrimitiveType = "None"
	Any      PrimitiveType = "any"
)

func (type_ PrimitiveType) String() string {
	return string(type_)
}

func (type_ PrimitiveType) IsPrimitive() bool {
	return true
}

func (type_ PrimitiveType) GetSizeInBytes() int {
	switch type_ {
	case Int32, Uint32, Float, Char, Bool:
		return 4
	case String:
		return 4 // treat as 32 bit pointer
	case Int64, Uint64, Double:
		return 8
	default:
		return 0
	}
}

func (type_ DynamicType) String() string {
	return string(type_)
}

func (type_ DynamicType) IsPrimitive() bool {
	return false
}

func (type_ DynamicType) GetSizeInBytes() int {
	return 4 // treat it as a 32 bit pointer
}

func Join(types []DataType) string {
	paramStr := strings.Builder{}
	end := len(types) - 1
	for i, param := range types {
		paramStr.WriteString(param.String())
		if i < end {
			paramStr.WriteRune(',')
		}
	}
	return paramStr.String()
}
