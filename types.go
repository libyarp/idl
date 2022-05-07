package idl

import "fmt"

//go:generate stringer -type=PrimitiveType,Element -output=types_string.go

// PrimitiveType represents a single type recognized by YARP
type PrimitiveType int

const (
	Invalid PrimitiveType = iota
	Uint8
	Uint16
	Uint32
	Uint64
	Int8
	Int16
	Int32
	Int64
	Float32
	Float64
	Struct
	OneOf
	Bool
	String
)

// Element represents a single Token element kind in a source file
type Element int

const (
	InvalidElement Element = iota
	Identifier             // [a-z][0-9a-z_]
	OpenCurly              // {
	CloseCurly             // }
	OpenParen              // (
	CloseParen             // )
	OpenAngled             // <
	CloseAngled            // >
	Comma                  // ,
	Dot                    // .
	LineBreak              // \n
	Equal                  // =
	Number                 // 0-9+
	Arrow                  // ->
	Semi                   // ;
	Comment                // Anything from # onwards
	Annotation             // Anything from @ until next space
	StringElement          // Anything between "
	EOF
)

// Token represents a single token present in a source file
type Token struct {
	Type   Element
	Value  string
	Line   int
	Column int
}

func (t Token) is(o Element) bool { return t.Type == o }
func (t Token) String() string {
	return fmt.Sprintf("Token{Type=%d (%s), Value=%#v, Line=%d, Column=%d}", t.Type, t.Type.String(), t.Value, t.Line, t.Column)
}

// TypeType represents the concrete type of a Field.
type TypeType int

const (
	TypeInvalid TypeType = iota
	TypePrimitive
	TypeArray
	TypeMap
	TypeUnresolved
)

type Type interface {
	Type() TypeType
}

type Primitive struct {
	Kind PrimitiveType
}

func (Primitive) Type() TypeType { return TypePrimitive }

type Array struct {
	Of Type
}

func (Array) Type() TypeType { return TypeArray }

type Map struct {
	Key   PrimitiveType
	Value Type
}

func (Map) Type() TypeType { return TypeMap }

type Unresolved struct {
	Name string
}

func (Unresolved) Type() TypeType { return TypeUnresolved }
