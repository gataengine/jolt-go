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
	Funcs        []Func
	Structs      []Struct
	Enums        []Enum
	Typedefs     map[string]string // alias -> underlying (scalar typedefs)
	FuncTypedefs []string          // names of function-pointer typedef types (e.g. JPH_TireMaxImpulseCallback)
}

// SplitMethod splits a JPH_<Receiver>_<Method> name. Returns ok=false if the
// name doesn't match (free function, e.g. JPH_RegisterTypes).
//
// Disambiguation rule for multi-segment receivers like BodyInterface: the
// receiver is determined by matching against KnownTypes (set elsewhere by
// the caller). For Phase 2 we use a simple heuristic — recv is the part
// up to the last underscore — and refine in Phase 3 once we have a type set.
func SplitMethod(name string) (recv, method string, ok bool) {
	const prefix = "JPH_"
	if len(name) < len(prefix) || name[:len(prefix)] != prefix {
		return "", "", false
	}
	rest := name[len(prefix):]
	// Find last underscore; everything before is the receiver, everything
	// after is the method.
	for i := len(rest) - 1; i >= 0; i-- {
		if rest[i] == '_' {
			return rest[:i], rest[i+1:], true
		}
	}
	return "", "", false
}
