// Package generator parses joltc.h and emits the cgo + idiomatic Go layers.
package generator

// Param is a function parameter or field descriptor.
type Param struct {
	Name  string
	CType string // raw C type as it appears in the header
}

// Func is a C function declaration.
type Func struct {
	Name   string
	Return string
	Params []Param
}

// Struct is a C struct (or opaque tag).
type Struct struct {
	Name   string
	Opaque bool
	Fields []Param
}

// Enum is a C enum.
type Enum struct {
	Name    string
	Members []EnumMember
}

type EnumMember struct {
	Name  string
	Value int64
}

// Header is the parsed contents of joltc.h.
type Header struct {
	Funcs    []Func
	Structs  []Struct
	Enums    []Enum
	Typedefs map[string]string // alias -> underlying
}
