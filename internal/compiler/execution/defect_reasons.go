package execution

// Defect reasons shared by execution producers, validators, and native lowering.
const (
	DefectReasonU64AddOverflow       = "U64 addition overflow"
	DefectReasonU64RemainderByZero   = "U64 remainder by zero"
	DefectReasonByteIndexOutOfBounds = "byte index out of bounds"
)
