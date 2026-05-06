package generator_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gataengine/jolt-go/generator"
)

var _ = Describe("Parser", func() {
	Describe("Funcs", func() {
		var h *generator.Header
		BeforeEach(func() {
			var err error
			h, err = generator.ParseHeader("testdata/funcs.h")
			Expect(err).NotTo(HaveOccurred())
		})

		It("extracts free functions", func() {
			names := funcNames(h.Funcs)
			Expect(names).To(ContainElements("JPH_RegisterTypes", "JPH_UnregisterTypes"))
		})

		It("extracts methods with self pointer", func() {
			f := findFunc(h.Funcs, "JPH_Body_GetPosition")
			Expect(f.Return).To(Equal("JPH_Vec3"))
			Expect(f.Params).To(HaveLen(1))
			Expect(f.Params[0].CType).To(Equal("const JPH_Body *"))
		})

		It("extracts methods with multiple params", func() {
			f := findFunc(h.Funcs, "JPH_Body_SetLinearVelocity")
			Expect(f.Params).To(HaveLen(2))
			Expect(f.Params[1].CType).To(Equal("const JPH_Vec3 *"))
		})

		It("renders pointer return type as 'JPH_Body *'", func() {
			f := findFunc(h.Funcs, "JPH_BodyInterface_CreateBody")
			Expect(f.Return).To(Equal("JPH_Body *"))
		})
	})
})

func funcNames(fs []generator.Func) []string {
	names := make([]string, len(fs))
	for i, f := range fs {
		names[i] = f.Name
	}
	return names
}

func findFunc(fs []generator.Func, name string) generator.Func {
	for _, f := range fs {
		if f.Name == name {
			return f
		}
	}
	Fail("func not found: " + name)
	return generator.Func{}
}
