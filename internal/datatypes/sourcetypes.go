package datatypes

import (
	"fmt"
	"slices"
	"strings"
)

type SourceType interface {
	String() string
	IsPrimitive() bool
	IsScopeRef() bool
	IsRef() bool
	GetScopes() []string
	GetSizeInBytes() int
}

type PrimitiveType string
type DynamicType string
type ScopeRef struct {
	Scopes []string
}
type Ref struct {
	Scope string
}

const (
	Int32  PrimitiveType = "int"
	Int64  PrimitiveType = "int64"
	Uint32 PrimitiveType = "uint32"
	Uint64 PrimitiveType = "uint64"
	Float  PrimitiveType = "float"
	Double PrimitiveType = "double"
	Bool   PrimitiveType = "bool"
	Char   PrimitiveType = "char"
	String PrimitiveType = "String"
	None   PrimitiveType = "None"
	Any    PrimitiveType = "any"
)

func (type_ PrimitiveType) String() string {
	return string(type_)
}

func (type_ PrimitiveType) IsPrimitive() bool {
	return true
}

func (type_ PrimitiveType) IsScopeRef() bool {
	return false
}

func (type_ PrimitiveType) IsRef() bool {
	return false
}

func (type_ PrimitiveType) GetScopes() []string {
	return nil
}

func (type_ PrimitiveType) GetSizeInBytes() int {
	switch type_ {
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

func (type_ DynamicType) String() string {
	return string(type_)
}

func (type_ DynamicType) IsPrimitive() bool {
	return false
}

func (type_ DynamicType) IsScopeRef() bool {
	return false
}

func (type_ DynamicType) IsRef() bool {
	return false
}

func (type_ DynamicType) GetScopes() []string {
	return nil
}

func (type_ DynamicType) GetSizeInBytes() int {
	return 4 // treat it as a 32 bit pointer
}

func (type_ ScopeRef) String() string {
	return fmt.Sprintf("ScopeRef(%s)", strings.Join(type_.Scopes, ","))
}

func (type_ ScopeRef) IsPrimitive() bool {
	return false
}

func (type_ ScopeRef) IsScopeRef() bool {
	return true
}

func (type_ ScopeRef) IsRef() bool {
	return false
}

func (type_ ScopeRef) GetScopes() []string {
	return type_.Scopes
}

func (type_ ScopeRef) GetSizeInBytes() int {
	return 0
}

func (type_ Ref) String() string {
	return fmt.Sprintf("Ref(%s)", type_.Scope)
}

func (type_ Ref) IsPrimitive() bool {
	return false
}

func (type_ Ref) IsScopeRef() bool {
	return false
}

func (type_ Ref) IsRef() bool {
	return true
}

func (type_ Ref) GetScopes() []string {
	return []string{type_.Scope}
}

func (type_ Ref) GetSizeInBytes() int {
	return 0
}

func Join(types []SourceType) string {
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

var (
	UnsignedTypes  []SourceType = []SourceType{Uint32, Uint64}
	SignedIntTypes []SourceType = []SourceType{Int32, Int64}
	IntTypes       []SourceType = slices.Concat(UnsignedTypes, SignedIntTypes)
	FloatTypes     []SourceType = []SourceType{Float, Double}
	SignedTypes    []SourceType = slices.Concat(SignedIntTypes, FloatTypes)
	NumericTypes   []SourceType = slices.Concat(UnsignedTypes, SignedTypes)
)
