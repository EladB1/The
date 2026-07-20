package datatypes

type IRType string

const (
	I32       IRType = "i32"
	I64       IRType = "i64"
	F32       IRType = "f32"
	F64       IRType = "f64"
	Str_const IRType = "str_const"
	Ptr       IRType = "ptr"
	NoneIR    IRType = "none"
)

func TranslateSourceType(srcType SourceType) IRType {
	irType := NoneIR
	switch srcType {
	case String:
		return Str_const
	case Char, Bool, Int32, Uint32:
		return I32
	case Int64, Uint64:
		return I64
	case Float:
		return F32
	case Double:
		return F64
	}
	return irType
}
