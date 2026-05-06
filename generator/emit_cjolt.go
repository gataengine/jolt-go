package generator

import (
	"fmt"
	"go/format"
	"sort"
	"strings"
)

// goReservedWords is the set of Go reserved words that must be suffixed with _
// when used as parameter names.
var goReservedWords = map[string]bool{
	"break": true, "case": true, "chan": true, "const": true,
	"continue": true, "default": true, "defer": true, "else": true,
	"fallthrough": true, "for": true, "func": true, "go": true,
	"goto": true, "if": true, "import": true, "interface": true,
	"map": true, "package": true, "range": true, "return": true,
	"select": true, "struct": true, "switch": true, "type": true,
	"var": true,
}

// safeParamName returns name suffixed with _ if it is a Go reserved word.
func safeParamName(name string) string {
	if goReservedWords[name] {
		return name + "_"
	}
	return name
}

// cScalarToGoNative maps C basic scalar type names to their Go-native equivalents.
// Types in this map are passed as Go-native in cjolt wrappers and cast to C at the call site.
var cScalarToGoNative = map[string]string{
	"float":        "float32",
	"double":       "float64",
	"bool":         "bool",
	"_Bool":        "bool",
	"int":          "int32",
	"unsigned int": "uint32",
	"uint8_t":      "uint8",
	"int8_t":       "int8",
	"uint16_t":     "uint16",
	"int16_t":      "int16",
	"uint32_t":     "uint32",
	"int32_t":      "int32",
	"uint64_t":     "uint64",
	"int64_t":      "int64",
}

// cScalarToCCastName maps C type names to the cgo cast identifier.
// For most types this is the same as the key, but multi-word types like
// "unsigned int" need their cgo alias ("uint").
var cScalarToCCastName = map[string]string{
	"unsigned int": "uint",
	"_Bool":        "bool",
	// All others: same as the C type name.
}

// cjoltGoType converts a C type string to its Go/cgo representation for the
// cjolt layer.
//
//   - "void *"             →  "unsafe.Pointer"
//   - "const T *" | "T *"  →  "*C.T"
//   - C basic scalar       →  Go-native type (float32, int32, etc.)
//   - scalar/enum (no *)   →  "C.T"
func cjoltGoType(cType string) string {
	s := cType
	// Strip leading "const " if present.
	s = strings.TrimPrefix(s, "const ")
	s = strings.TrimSpace(s)
	// void * → unsafe.Pointer (the renderType sentinel for all untyped pointers).
	if s == "void *" {
		return "unsafe.Pointer"
	}
	// Pointer-to-pointer (T **): use unsafe.Pointer for the Go API, with the C call
	// site emitting an explicit cast. We detect this by double-star suffix.
	if strings.HasSuffix(s, " **") {
		return "unsafe.Pointer"
	}
	// Pointer type: ends with '*' after trimming.
	if base, ok := strings.CutSuffix(s, "*"); ok {
		base = strings.TrimSpace(base)
		// Strip any nested const from the base.
		base = strings.TrimPrefix(base, "const ")
		base = strings.TrimSpace(base)
		// Primitive C types and char* → unsafe.Pointer so callers (including the jolt
		// layer) can pass Go pointers / unsafe.Pointer without C type imports.
		switch base {
		case "char", "float", "double",
			"uint8_t", "uint16_t", "uint32_t", "uint64_t",
			"int8_t", "int16_t", "int32_t", "int64_t", "bool", "_Bool":
			return "unsafe.Pointer"
		}
		return "*C." + base
	}
	// Check scalar map first.
	if goNative, ok := cScalarToGoNative[s]; ok {
		return goNative
	}
	// Scalar / enum type.
	return "C." + s
}

// cjoltReturnType converts a C return type string to its Go/cgo representation
// for the cjolt layer.
//
//   - "void *" or "const T *" or "T *"  → "*C.T" (const stripped)
//   - "_Bool" or "bool"                  → "bool" (Go native, avoids C._Bool in API)
//   - otherwise                          → "C.T"
func cjoltReturnType(ret string) string {
	s := strings.TrimSpace(ret)
	// Strip outer const qualifier.
	s = strings.TrimPrefix(s, "const ")
	s = strings.TrimSpace(s)
	// void * → unsafe.Pointer.
	if s == "void *" {
		return "unsafe.Pointer"
	}
	if base, ok := strings.CutSuffix(s, "*"); ok {
		base = strings.TrimSpace(base)
		// Strip any nested const (e.g. "const const char *").
		base = strings.TrimPrefix(base, "const ")
		base = strings.TrimSpace(base)
		return "*C." + base
	}
	// Map scalar C types to Go-native types to avoid C._Ctype_xxx in the API.
	if goNative, ok := cScalarToGoNative[s]; ok {
		return goNative
	}
	return "C." + s
}

// contactListenerCBridge is the forward-declaration emitted into the cgo
// comment block so that Go init() can call register_contact_listener_procs
// via C.register_contact_listener_procs(). The actual implementation lives in
// internal/cjolt/contact_listener_procs.c — a plain C file (not a cgo
// preamble) that includes _cgo_export.h (made available by cgo on the include
// path at build time). That file #includes _cgo_export.h to get the extern
// declarations of the //export'd Go trampolines, then assembles the
// JPH_ContactListener_Procs struct and calls JPH_ContactListener_SetProcs.
//
// We cannot do any of that inside the cgo comment preamble because
// _cgo_export.h does not exist yet when the preamble is compiled. A separate
// .c file in the same package is the standard cgo idiom for this pattern.
const contactListenerCBridge = `
// Forward declaration of the C helper defined in contact_listener_procs.c.
void register_contact_listener_procs(void);
`

// hasContactListener reports whether the annotation set includes a
// JPH_ContactListener listener spec (which requires the C bridge).
func hasContactListener() bool {
	for _, spec := range Annotations.Listeners {
		if spec.CType == "JPH_ContactListener" {
			return true
		}
	}
	return false
}

// EmitCjolt returns the source of internal/cjolt/cjolt_generated.go.
func EmitCjolt(h *Header) string {
	var b strings.Builder
	b.WriteString("// Code generated by jolt-go generator. DO NOT EDIT.\n\npackage cjolt\n\n")

	// The cgo comment block: always includes joltc.h; when a ContactListener
	// is present it also embeds the extern declarations + static C bridge so
	// that Go init() can call C.register_contact_listener_procs().
	b.WriteString("/*\n#include \"joltc.h\"\n")
	if hasContactListener() {
		b.WriteString(contactListenerCBridge)
	}
	b.WriteString("*/\nimport \"C\"\n\n")

	// Emit imports. Always include "unsafe" because void* params use unsafe.Pointer.
	// Include "runtime/cgo" when listeners are present.
	if len(Annotations.Listeners) > 0 {
		b.WriteString("import (\n\t\"runtime/cgo\"\n\t\"unsafe\"\n)\n\n")
	} else {
		b.WriteString("import \"unsafe\"\n\n")
	}

	for _, s := range h.Structs {
		fmt.Fprintf(&b, "type C_%s = C.%s\n", s.Name, s.Name)
	}
	for _, e := range h.Enums {
		fmt.Fprintf(&b, "type C_%s = C.%s\n", e.Name, e.Name)
	}
	// Emit type aliases for scalar typedefs (e.g. JPH_BodyID = uint32_t → C_JPH_BodyID = C.JPH_BodyID)
	// and non-scalar typedefs (e.g. JPH_RVec3 = JPH_Vec3 → C_JPH_RVec3 = C.JPH_RVec3).
	// This lets the jolt package reference cjolt.C_JPH_BodyID, cjolt.C_JPH_RVec3, etc.
	{
		var aliases []string
		for alias := range h.Typedefs {
			aliases = append(aliases, alias)
		}
		// Sort for deterministic output.
		sort.Strings(aliases)
		for _, alias := range aliases {
			fmt.Fprintf(&b, "type C_%s = C.%s\n", alias, alias)
		}
	}
	// Emit type aliases for function-pointer typedefs (e.g. JPH_TireMaxImpulseCallback).
	for _, ft := range h.FuncTypedefs {
		fmt.Fprintf(&b, "type C_%s = C.%s\n", ft, ft)
	}
	b.WriteString("\n")

	// Emit listener handle types and trampolines.
	for _, spec := range Annotations.Listeners {
		b.WriteString(emitListenerHandleType(spec))
		b.WriteString(emitListenerTrampolines(spec))
		b.WriteString(emitListenerProcsInit(spec))
	}

	for _, f := range h.Funcs {
		b.WriteString(emitCjoltFunc(f))
	}
	out, err := format.Source([]byte(b.String()))
	if err != nil {
		return b.String() // unformatted; tests will catch it
	}
	return string(out)
}

// emitListenerHandleType emits the ContactListenerHandle struct with C.*-typed func fields.
func emitListenerHandleType(spec ListenerSpec) string {
	if spec.CType != "JPH_ContactListener" {
		return "" // only ContactListener is supported in MVP
	}
	return `// ContactListenerHandle is the value type stored in cgo.Handle for joltc
// ContactListener instances. The Go-side bridge in package jolt sets the
// non-nil func fields.
type ContactListenerHandle struct {
	OnContactValidate  func(b1, b2 *C.JPH_Body, baseOffset *C.JPH_RVec3, result *C.JPH_CollideShapeResult) C.JPH_ValidateResult
	OnContactAdded     func(b1, b2 *C.JPH_Body, manifold *C.JPH_ContactManifold, settings *C.JPH_ContactSettings)
	OnContactPersisted func(b1, b2 *C.JPH_Body, manifold *C.JPH_ContactManifold, settings *C.JPH_ContactSettings)
	OnContactRemoved   func(subShapePair *C.JPH_SubShapeIDPair)
}

`
}

// emitListenerTrampolines emits one //export trampoline per listener method.
func emitListenerTrampolines(spec ListenerSpec) string {
	if spec.CType != "JPH_ContactListener" {
		return ""
	}
	return `//export jolt_go_contactListener_OnContactValidate
func jolt_go_contactListener_OnContactValidate(userData unsafe.Pointer, b1, b2 *C.JPH_Body, baseOffset *C.JPH_RVec3, result *C.JPH_CollideShapeResult) C.JPH_ValidateResult {
	h := cgo.Handle(uintptr(userData))
	cl := h.Value().(*ContactListenerHandle)
	if cl.OnContactValidate != nil {
		return cl.OnContactValidate(b1, b2, baseOffset, result)
	}
	return 0
}

//export jolt_go_contactListener_OnContactAdded
func jolt_go_contactListener_OnContactAdded(userData unsafe.Pointer, b1, b2 *C.JPH_Body, manifold *C.JPH_ContactManifold, settings *C.JPH_ContactSettings) {
	h := cgo.Handle(uintptr(userData))
	cl := h.Value().(*ContactListenerHandle)
	if cl.OnContactAdded != nil {
		cl.OnContactAdded(b1, b2, manifold, settings)
	}
}

//export jolt_go_contactListener_OnContactPersisted
func jolt_go_contactListener_OnContactPersisted(userData unsafe.Pointer, b1, b2 *C.JPH_Body, manifold *C.JPH_ContactManifold, settings *C.JPH_ContactSettings) {
	h := cgo.Handle(uintptr(userData))
	cl := h.Value().(*ContactListenerHandle)
	if cl.OnContactPersisted != nil {
		cl.OnContactPersisted(b1, b2, manifold, settings)
	}
}

//export jolt_go_contactListener_OnContactRemoved
func jolt_go_contactListener_OnContactRemoved(userData unsafe.Pointer, subShapePair *C.JPH_SubShapeIDPair) {
	h := cgo.Handle(uintptr(userData))
	cl := h.Value().(*ContactListenerHandle)
	if cl.OnContactRemoved != nil {
		cl.OnContactRemoved(subShapePair)
	}
}

`
}

// emitListenerProcsInit emits the init() function that registers the trampolines via SetProcs.
// The actual registration is done by the static C helper register_contact_listener_procs
// that is embedded in the cgo comment block by EmitCjolt (see contactListenerCBridge).
// Calling it from Go init() is safe because cgo allows calling static C functions.
func emitListenerProcsInit(spec ListenerSpec) string {
	if spec.CType != "JPH_ContactListener" {
		return ""
	}
	return `func init() {
	C.register_contact_listener_procs()
}

`
}

// cjoltArgExpr returns the expression to use when passing a Go-side param of the
// given C type to the C function. For Go-native scalar types, wraps in a C cast;
// for pointer and enum types passes through unchanged.
func cjoltArgExpr(cType, argName string) string {
	s := strings.TrimPrefix(cType, "const ")
	s = strings.TrimSpace(s)
	// Pointer-to-pointer (T **): Go side is unsafe.Pointer; C side needs (**C.T) cast.
	if innerBase, ok := strings.CutSuffix(s, " **"); ok {
		innerBase = strings.TrimSpace(innerBase)
		return "(**C." + innerBase + ")(unsafe.Pointer(" + argName + "))"
	}
	// Pointers pass through, with special handling for primitive pointer types
	// whose Go type is unsafe.Pointer (needs explicit C cast at the call site).
	if base, ok := strings.CutSuffix(s, "*"); ok {
		base = strings.TrimSpace(base)
		base = strings.TrimPrefix(base, "const ")
		base = strings.TrimSpace(base)
		// Primitive pointer types: Go side is unsafe.Pointer; C side needs (*C.T) cast.
		switch base {
		case "char", "float", "double",
			"uint8_t", "uint16_t", "uint32_t", "uint64_t",
			"int8_t", "int16_t", "int32_t", "int64_t", "bool", "_Bool":
			return "(*C." + base + ")(unsafe.Pointer(" + argName + "))"
		}
		// void * and typed JPH_* pointers: unsafe.Pointer passes through directly
		// for void *, and *T passes through for typed pointers.
		return argName
	}
	// If it maps to a Go-native type, emit C.<castname>(argName).
	if _, ok := cScalarToGoNative[s]; ok {
		castName := s
		if alias, ok2 := cScalarToCCastName[s]; ok2 {
			castName = alias
		}
		return "C." + castName + "(" + argName + ")"
	}
	// Enum / JPH_* type: pass through.
	return argName
}

func emitCjoltFunc(f Func) string {
	// Build parameter list and argument list.
	paramParts := make([]string, 0, len(f.Params))
	argParts := make([]string, 0, len(f.Params))
	for _, p := range f.Params {
		safe := safeParamName(p.Name)
		goType := cjoltGoType(p.CType)
		paramParts = append(paramParts, safe+" "+goType)
		argParts = append(argParts, cjoltArgExpr(p.CType, safe))
	}
	paramList := strings.Join(paramParts, ", ")
	argList := strings.Join(argParts, ", ")

	if f.Return == "void" {
		return fmt.Sprintf("func %s(%s) { C.%s(%s) }\n\n",
			f.Name, paramList, f.Name, argList)
	}
	retType := cjoltReturnType(f.Return)
	// If the return type is a Go-native scalar (e.g. bool, uint32), the cgo call
	// returns a C._Ctype_xxx which needs explicit conversion.
	retStripped := strings.TrimPrefix(f.Return, "const ")
	retStripped = strings.TrimSpace(retStripped)
	if _, isScalar := cScalarToGoNative[retStripped]; isScalar {
		if retStripped == "_Bool" || retStripped == "bool" {
			// C._Bool is a CGo bool type; explicit conversion to Go bool is needed.
			return fmt.Sprintf("func %s(%s) %s {\nreturn bool(C.%s(%s))\n}\n\n",
				f.Name, paramList, retType, f.Name, argList)
		}
		return fmt.Sprintf("func %s(%s) %s {\nreturn %s(C.%s(%s))\n}\n\n",
			f.Name, paramList, retType, retType, f.Name, argList)
	}
	return fmt.Sprintf("func %s(%s) %s {\nreturn C.%s(%s)\n}\n\n",
		f.Name, paramList, retType, f.Name, argList)
}
