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
	if srcType.Equals(StringType) {
		return Str_const
	} else if srcType.Equals(CharType) || srcType.Equals(BoolType) || srcType.Equals(Int32Type) {
		return I32
	} else if srcType.Equals(Uint32Type) {
		return U32
	} else if srcType.Equals(Uint64Type) {
		return U64
	} else if srcType.Equals(Int64Type) {
		return I64
	} else if srcType.Equals(FloatType) {
		return F32
	} else if srcType.Equals(DoubleType) {
		return F64
	}
	return irType
}
