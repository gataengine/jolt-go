package generator_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gataengine/jolt-go/generator"
)

var _ = Describe("EmitCjolt — method-shape", func() {
	Context("method-shaped functions", func() {
		h := &generator.Header{
			Funcs: []generator.Func{
				{
					Name:   "JPH_Body_GetID",
					Return: "JPH_BodyID",
					Params: []generator.Param{{Name: "self", CType: "const JPH_Body *"}},
				},
				{
					Name:   "JPH_Body_SetMotionType",
					Return: "void",
					Params: []generator.Param{
						{Name: "self", CType: "JPH_Body *"},
						{Name: "type", CType: "JPH_MotionType"},
					},
				},
			},
		}
		It("emits self pointer wrapper", func() {
			src := generator.EmitCjolt(h)
			Expect(src).To(ContainSubstring(
				"func JPH_Body_GetID(self *C.JPH_Body) C.JPH_BodyID {",
			))
			Expect(src).To(ContainSubstring(
				"return C.JPH_Body_GetID(self)",
			))
		})
		It("emits enum-arg method", func() {
			src := generator.EmitCjolt(h)
			Expect(src).To(ContainSubstring(
				"func JPH_Body_SetMotionType(self *C.JPH_Body, type_ C.JPH_MotionType) {",
			))
		})
	})
})

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
