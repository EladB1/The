package datatypes

import (
	"fmt"
	"strings"
)

type DataType string

const (
	Int32     DataType = "int"
	Int64     DataType = "int64"
	Uint32    DataType = "uint32"
	Uint64    DataType = "uint64"
	Float     DataType = "float"
	Double    DataType = "double"
	Bool      DataType = "bool"
	Char      DataType = "char"
	String    DataType = "String"
	None      DataType = "None"
	Any       DataType = "any"
	GlobalRef DataType = "Ref"
	Ref       DataType = "Ref"
	ScopeRef  DataType = "ScopeRef"
)

// DynamicType would be DataType(struct_or_interface_name)

type SourceType struct {
	Root      DataType
	SubTypes  []SourceType
	IsDynamic bool
}

var (
	EmptySourceType SourceType = SourceType{}
	Int32Type       SourceType = newPrimitive(Int32)
	Int64Type       SourceType = newPrimitive(Int64)
	Uint32Type      SourceType = newPrimitive(Uint32)
	Uint64Type      SourceType = newPrimitive(Uint64)
	FloatType       SourceType = newPrimitive(Float)
	DoubleType      SourceType = newPrimitive(Double)
	BoolType        SourceType = newPrimitive(Bool)
	CharType        SourceType = newPrimitive(Char)
	StringType      SourceType = newPrimitive(String)
	NoneType        SourceType = newPrimitive(None)
	AnyType         SourceType = newPrimitive(Any)
	GlobalRefType   SourceType = NewContainerType(GlobalRef, NewReferenceSubType("@global"))
)

func newPrimitive(dt DataType) SourceType {
	return SourceType{Root: dt, IsDynamic: false}
}

func NewDynamicType(dt string) SourceType {
	return SourceType{Root: DataType(dt), IsDynamic: true}
}

func NewDynamicContainerType(dt DataType, subTypes ...SourceType) SourceType {
	return SourceType{Root: dt, SubTypes: subTypes, IsDynamic: true}
}

func NewContainerType(dt DataType, subTypes ...SourceType) SourceType {
	return SourceType{Root: dt, SubTypes: subTypes, IsDynamic: false}
}

func NewReferenceSubType(dt string) SourceType {
	return SourceType{Root: DataType(dt), IsDynamic: false}
}

func (st SourceType) String() string {
	if st.Root == Ref {
		return fmt.Sprintf("Ref(%s)", JoinTypes(st.SubTypes))
	} else if st.Root == ScopeRef {
		return fmt.Sprintf("ScopeRef(%s)", JoinTypes(st.SubTypes))
	} else {
		return string(st.Root)
	}
}

func (st SourceType) RootEquals(dt DataType) bool {
	return st.Root == dt
}

func (st SourceType) Equals(other SourceType) bool {
	return st.Root == other.Root && st.IsDynamic == other.IsDynamic && st.HasMatchingSubTypes(other)
}

func (st SourceType) HasMatchingSubTypes(other SourceType) bool {
	if len(st.SubTypes) != len(other.SubTypes) {
		return false
	}
	for i := range len(st.SubTypes) {
		if !st.SubTypes[i].Equals(other.SubTypes[i]) {
			return false
		}
	}
	return true
}

func (st SourceType) GetSizeInBytes() int {
	if st.IsDynamic {
		return 4
	}
	switch st.Root {
	case Int32, Uint32, Float, Char, Bool, String:
		// Strings are 32 bit pointers
		// chars are 32 bit unicode characters
		// bool will be stored about 32 bit int in IR/Wasm
		return 4
	case Int64, Uint64, Double:
		return 8
	default:
		return 0
	}
}

func (st SourceType) IsNumeric() bool {
	return st.IsSignedType() || st.IsUnsignedType()
}

func (st SourceType) IsSignedType() bool {
	return st.IsFloatType() || st.IsSignedIntType()
}

func (st SourceType) IsFloatType() bool {
	return st.Root == Float || st.Root == Double
}

func (st SourceType) IsIntType() bool {
	return st.IsUnsignedType() || st.IsSignedIntType()
}

func (st SourceType) IsUnsignedType() bool {
	return st.Root == Uint32 || st.Root == Uint64
}

func (st SourceType) IsSignedIntType() bool {
	return st.Root == Int32 || st.Root == Int64
}

func JoinTypes(types []SourceType) string {
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
