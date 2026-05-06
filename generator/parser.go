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

			fn, err := extractFunc(d, t.(*cc.FunctionType))
			if err != nil {
				continue
			}
			h.Funcs = append(h.Funcs, fn)
		}
	}

	return h, nil
}

// extractFunc builds a Func from a declarator with function type.
func extractFunc(d *cc.Declarator, ft *cc.FunctionType) (Func, error) {
	f := Func{
		Name:   d.Name(),
		Return: renderReturnType(ft.Result()),
	}

	for _, p := range ft.Parameters() {
		pt := p.Type()
		// A single `void` parameter means no parameters in C.
		if pt.Kind() == cc.Void {
			break
		}
		f.Params = append(f.Params, Param{
			Name:  p.Name(),
			CType: renderParamType(pt),
		})
	}

	return f, nil
}

// renderReturnType converts a cc Type to its C type string for return values.
// For void, returns "void". For struct/typedef types, returns the typedef name
// or struct name. For pointers, uses renderParamType.
func renderReturnType(t cc.Type) string {
	if t == nil {
		return "void"
	}
	if t.Kind() == cc.Void {
		return "void"
	}
	// Use typedef name if available.
	if name := t.Typedef().Name(); name != "" {
		return name
	}
	return t.String()
}

// renderParamType converts a cc Type to a C type string in the canonical form
// expected by the test: "const JPH_Body *" (space before star, const before base).
func renderParamType(t cc.Type) string {
	switch t.Kind() {
	case cc.Ptr:
		pt := t.(*cc.PointerType)
		elem := pt.Elem()
		base := elemTypeName(elem)
		if elem.Attributes().IsConst() {
			return "const " + base + " *"
		}
		return base + " *"
	default:
		if name := t.Typedef().Name(); name != "" {
			return name
		}
		return t.String()
	}
}

// elemTypeName returns the name for a pointer's element type (without const).
func elemTypeName(t cc.Type) string {
	// Use typedef name if available.
	if name := t.Typedef().Name(); name != "" {
		return name
	}
	return t.String()
}
