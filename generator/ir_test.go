package generator_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gataengine/jolt-go/generator"
)

var _ = Describe("IR", func() {
	It("Func has Name and Params", func() {
		f := generator.Func{Name: "JPH_Body_GetID", Params: []generator.Param{{Name: "self", CType: "JPH_Body*"}}}
		Expect(f.Name).To(Equal("JPH_Body_GetID"))
		Expect(f.Params).To(HaveLen(1))
	})
})

var _ = Describe("MethodPattern", func() {
	It("splits JPH_Body_GetPosition", func() {
		recv, method, ok := generator.SplitMethod("JPH_Body_GetPosition")
		Expect(ok).To(BeTrue())
		Expect(recv).To(Equal("Body"))
		Expect(method).To(Equal("GetPosition"))
	})
	It("returns ok=false for free functions", func() {
		_, _, ok := generator.SplitMethod("JPH_RegisterTypes")
		Expect(ok).To(BeFalse())
	})
	It("respects multi-segment receivers (JPH_BodyInterface_CreateBody)", func() {
		// JPH_BodyInterface_CreateBody → recv="BodyInterface", method="CreateBody"
		recv, method, _ := generator.SplitMethod("JPH_BodyInterface_CreateBody")
		Expect(recv).To(Equal("BodyInterface"))
		Expect(method).To(Equal("CreateBody"))
	})
})
