package generator_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gataengine/jolt-go/generator"
)

var _ = Describe("EmitCjolt", func() {
	Context("free functions", func() {
		h := &generator.Header{
			Funcs: []generator.Func{
				{Name: "JPH_RegisterTypes", Return: "void"},
				{Name: "JPH_UnregisterTypes", Return: "void"},
			},
		}
		It("emits Go wrappers calling C functions", func() {
			src := generator.EmitCjolt(h)
			Expect(src).To(ContainSubstring("package cjolt"))
			Expect(src).To(ContainSubstring("func JPH_RegisterTypes() { C.JPH_RegisterTypes() }"))
			Expect(src).To(ContainSubstring("func JPH_UnregisterTypes() { C.JPH_UnregisterTypes() }"))
		})
	})
})
