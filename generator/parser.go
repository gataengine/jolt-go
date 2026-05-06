package generator

import (
	"fmt"
	"runtime"

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

	h := &Header{
		// populated by typedef extraction in Task 13
		Typedefs: make(map[string]string),
	}

	for l := ast.TranslationUnit; l != nil; l = l.TranslationUnit {
		ed := l.ExternalDeclaration
		if ed.Case != cc.ExternalDeclarationDecl {
			continue
		}
		decl := ed.Declaration
		if decl.Case != cc.DeclarationDecl {
			continue
		}

		for idl := decl.InitDeclaratorList; idl != nil; idl = idl.InitDeclaratorList {
			id := idl.InitDeclarator
			if id == nil || id.Declarator == nil {
				continue
			}
			d := id.Declarator
			t := d.Type()
			if t == nil || t.Kind() != cc.Function {
				continue
			}

			fn := extractFunc(d, t.(*cc.FunctionType))
			h.Funcs = append(h.Funcs, fn)
		}
	}

	return h, nil
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
