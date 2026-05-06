package generator

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"modernc.org/cc/v4"
)

// ParseHeader pre-processes the file with cc/v4 and extracts the IR.
func ParseHeader(path string) (*Header, error) {
	cfg, err := cc.NewConfig(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return nil, fmt.Errorf("cc.NewConfig: %w", err)
	}

	ast, err := cc.Translate(cfg, []cc.Source{
		{Name: "<predefined>", Value: cfg.Predefined},
		{Name: "<builtin>", Value: cc.Builtin},
		{Name: path},
	})
	if err != nil {
		return nil, fmt.Errorf("cc.Translate %s: %w", path, err)
	}

	// Resolve the absolute path of the user's header for origin filtering.
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	h := &Header{
		Typedefs: make(map[string]string),
	}

	// Track structs and enums we have already emitted (by name) to avoid
	// double-counting typedef struct / typedef enum definitions.
	seenStructs := make(map[string]bool)
	seenEnums := make(map[string]bool)

	for l := ast.TranslationUnit; l != nil; l = l.TranslationUnit {
		ed := l.ExternalDeclaration
		if ed.Case != cc.ExternalDeclarationDecl {
			continue
		}
		decl := ed.Declaration
		if decl.Case != cc.DeclarationDecl {
			continue
		}

		// Filter: only process declarations that originate from the user's file.
		// cc/v4 Position() returns the position of the first token of the node.
		// For declarations from system headers, the filename will differ from absPath.
		pos := decl.Position()
		declFile, _ := filepath.Abs(pos.Filename)
		if declFile != absPath {
			continue
		}

		// Detect whether this is a typedef declaration by walking DeclarationSpecifiers.
		isTypedef := false
		for ds := decl.DeclarationSpecifiers; ds != nil; ds = ds.DeclarationSpecifiers {
			if ds.Case == cc.DeclarationSpecifiersStorage &&
				ds.StorageClassSpecifier != nil &&
				ds.StorageClassSpecifier.Case == cc.StorageClassSpecifierTypedef {
				isTypedef = true
				break
			}
		}

		// Find the TypeSpecifier within DeclarationSpecifiers (handles the
		// "typedef" storage class which sits before it in the chain).
		ts := findTypeSpecifier(decl.DeclarationSpecifiers)

		// --- Struct handling ---
		if ts != nil && ts.Case == cc.TypeSpecifierStructOrUnion && ts.StructOrUnionSpecifier != nil {
			sus := ts.StructOrUnionSpecifier
			// Only handle struct (not union).
			if sus.StructOrUnion != nil && strings.ToLower(sus.StructOrUnion.Token.SrcStr()) == "struct" {
				structName := sus.Token.SrcStr()
				if structName != "" && !seenStructs[structName] {
					seenStructs[structName] = true
					s := Struct{Name: structName}
					if sus.Case == cc.StructOrUnionSpecifierTag {
						// No field list → opaque forward declaration.
						s.Opaque = true
					} else {
						// Has field list → extract fields.
						s.Fields = extractFields(sus.StructDeclarationList)
					}
					h.Structs = append(h.Structs, s)
				}
				// Don't add struct typedefs to h.Typedefs — they are captured above.
				continue
			}
		}

		// --- Enum handling ---
		if ts != nil && ts.Case == cc.TypeSpecifierEnum && ts.EnumSpecifier != nil {
			es := ts.EnumSpecifier
			// Token is the "enum" keyword; Token2 is the enum tag identifier.
			enumName := es.Token2.SrcStr()
			if enumName != "" && !seenEnums[enumName] {
				seenEnums[enumName] = true
				e := Enum{Name: enumName}
				if es.Case == cc.EnumSpecifierDef {
					e.Members = extractEnumMembers(es.EnumeratorList)
				}
				h.Enums = append(h.Enums, e)
			}
			// Don't add enum typedefs to h.Typedefs.
			continue
		}

		// --- Function and simple typedef handling ---
		// For simple typedefs, get the underlying type from the declaration
		// specifiers (e.g. for "typedef uint32_t JPH_ObjectLayer", this gives
		// "uint32_t" rather than the newly-defined alias name).
		dsType := decl.DeclarationSpecifiers.Type()

		for idl := decl.InitDeclaratorList; idl != nil; idl = idl.InitDeclaratorList {
			id := idl.InitDeclarator
			if id == nil || id.Declarator == nil {
				continue
			}
			d := id.Declarator
			t := d.Type()
			if t == nil {
				continue
			}

			switch t.Kind() {
			case cc.Function:
				fn := extractFunc(d, t.(*cc.FunctionType))
				h.Funcs = append(h.Funcs, fn)

			default:
				// Simple typedef alias (e.g. typedef uint32_t JPH_ObjectLayer).
				// Use the DeclarationSpecifiers type to get the underlying type name
				// so we get "uint32_t" rather than the alias being defined.
				if isTypedef {
					alias := d.Name()
					underlying := ""
					if dsType != nil {
						if name := dsType.Typedef().Name(); name != "" {
							underlying = name
						} else {
							underlying = dsType.String()
						}
					}
					if alias != "" && underlying != "" {
						h.Typedefs[alias] = underlying
					}
				}
			}
		}
	}

	return h, nil
}

// findTypeSpecifier walks a DeclarationSpecifiers chain and returns the first
// TypeSpecifier node found.
func findTypeSpecifier(ds *cc.DeclarationSpecifiers) *cc.TypeSpecifier {
	for ds != nil {
		if ds.Case == cc.DeclarationSpecifiersTypeSpec && ds.TypeSpecifier != nil {
			return ds.TypeSpecifier
		}
		ds = ds.DeclarationSpecifiers
	}
	return nil
}

// extractFields walks a StructDeclarationList and returns Params for each field.
func extractFields(sdl *cc.StructDeclarationList) []Param {
	var fields []Param
	for sdl != nil {
		sd := sdl.StructDeclaration
		if sd != nil && sd.Case == cc.StructDeclarationDecl {
			for sfl := sd.StructDeclaratorList; sfl != nil; sfl = sfl.StructDeclaratorList {
				sf := sfl.StructDeclarator
				if sf != nil && sf.Declarator != nil {
					d := sf.Declarator
					fields = append(fields, Param{
						Name:  d.Name(),
						CType: renderType(d.Type()),
					})
				}
			}
		}
		sdl = sdl.StructDeclarationList
	}
	return fields
}

// extractEnumMembers walks an EnumeratorList and returns EnumMember entries.
func extractEnumMembers(el *cc.EnumeratorList) []EnumMember {
	var members []EnumMember
	for el != nil {
		en := el.Enumerator
		if en != nil {
			name := en.Token.SrcStr()
			var val int64
			if v := en.Value(); v != nil {
				if iv, ok := v.(cc.Int64Value); ok {
					val = int64(iv)
				}
			}
			members = append(members, EnumMember{Name: name, Value: val})
		}
		el = el.EnumeratorList
	}
	return members
}

// extractFunc builds a Func from a declarator with function type.
func extractFunc(d *cc.Declarator, ft *cc.FunctionType) Func {
	f := Func{
		Name:   d.Name(),
		Return: renderType(ft.Result()),
	}

	for _, p := range ft.Parameters() {
		pt := p.Type()
		// A single `void` parameter means no parameters in C.
		if pt.Kind() == cc.Void {
			break
		}
		f.Params = append(f.Params, Param{
			Name:  p.Name(),
			CType: renderType(pt),
		})
	}

	return f
}

// renderType converts a cc Type to its canonical C type string.
// For void (or nil), returns "void". For pointer types, returns "BASE *" or
// "const BASE *". For other types, returns the typedef name when available,
// falling back to the type's own String() representation.
func renderType(t cc.Type) string {
	if t == nil || t.Kind() == cc.Void {
		return "void"
	}
	if t.Kind() == cc.Ptr {
		pt := t.(*cc.PointerType)
		elem := pt.Elem()
		base := elemTypeName(elem)
		if elem.Attributes().IsConst() {
			return "const " + base + " *"
		}
		return base + " *"
	}
	if name := t.Typedef().Name(); name != "" {
		return name
	}
	return t.String()
}

// elemTypeName returns the name for a pointer's element type (without const).
func elemTypeName(t cc.Type) string {
	// Use typedef name if available.
	if name := t.Typedef().Name(); name != "" {
		return name
	}
	return t.String()
}
