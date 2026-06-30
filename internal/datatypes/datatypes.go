package datatypes

type DataType interface {
	String() string
}

type PrimitiveType string
type DynamicType string

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
	None   PrimitiveType = ""
	Any    PrimitiveType = "any"
)

func (type_ PrimitiveType) String() string {
	return string(type_)
}

func (type_ DynamicType) String() string {
	return string(type_)
}
