package datatypes

type IRType string

const (
	I32       IRType = "i32"
	I64       IRType = "i64"
	U32       IRType = "u32"
	U64       IRType = "u64"
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
	case Char, Bool, Int32:
		return I32
	case Uint32:
		return U32
	case Uint64:
		return U64
	case Int64:
		return I64
	case Float:
		return F32
	case Double:
		return F64
	}
	return irType
}
