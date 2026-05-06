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
