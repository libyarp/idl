// Code generated by "stringer -type=PrimitiveType,Element -output=types_string.go"; DO NOT EDIT.

package idl

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Invalid-0]
	_ = x[Uint8-1]
	_ = x[Uint16-2]
	_ = x[Uint32-3]
	_ = x[Uint64-4]
	_ = x[Int8-5]
	_ = x[Int16-6]
	_ = x[Int32-7]
	_ = x[Int64-8]
	_ = x[Float32-9]
	_ = x[Float64-10]
	_ = x[Struct-11]
	_ = x[OneOf-12]
	_ = x[Bool-13]
	_ = x[String-14]
}

const _PrimitiveType_name = "InvalidUint8Uint16Uint32Uint64Int8Int16Int32Int64Float32Float64StructOneOfBoolString"

var _PrimitiveType_index = [...]uint8{0, 7, 12, 18, 24, 30, 34, 39, 44, 49, 56, 63, 69, 74, 78, 84}

func (i PrimitiveType) String() string {
	if i < 0 || i >= PrimitiveType(len(_PrimitiveType_index)-1) {
		return "PrimitiveType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _PrimitiveType_name[_PrimitiveType_index[i]:_PrimitiveType_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[InvalidElement-0]
	_ = x[Identifier-1]
	_ = x[OpenCurly-2]
	_ = x[CloseCurly-3]
	_ = x[OpenParen-4]
	_ = x[CloseParen-5]
	_ = x[OpenAngled-6]
	_ = x[CloseAngled-7]
	_ = x[Comma-8]
	_ = x[Dot-9]
	_ = x[LineBreak-10]
	_ = x[Equal-11]
	_ = x[Number-12]
	_ = x[Arrow-13]
	_ = x[Semi-14]
	_ = x[Comment-15]
	_ = x[Annotation-16]
	_ = x[StringElement-17]
	_ = x[EOF-18]
}

const _Element_name = "InvalidElementIdentifierOpenCurlyCloseCurlyOpenParenCloseParenOpenAngledCloseAngledCommaDotLineBreakEqualNumberArrowSemiCommentAnnotationStringElementEOF"

var _Element_index = [...]uint8{0, 14, 24, 33, 43, 52, 62, 72, 83, 88, 91, 100, 105, 111, 116, 120, 127, 137, 150, 153}

func (i Element) String() string {
	if i < 0 || i >= Element(len(_Element_index)-1) {
		return "Element(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Element_name[_Element_index[i]:_Element_index[i+1]]
}