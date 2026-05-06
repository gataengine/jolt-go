package generator_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gataengine/jolt-go/generator"
)

var _ = Describe("EmitJolt — owned types", func() {
	h := &generator.Header{
		Funcs: []generator.Func{
			{Name: "JPH_PhysicsSystem_Create", Return: "JPH_PhysicsSystem*"},
			{Name: "JPH_PhysicsSystem_Destroy", Return: "void",
				Params: []generator.Param{{Name: "self", CType: "JPH_PhysicsSystem *"}}},
		},
	}
	It("generates type, constructor, Close, finalizer", func() {
		src := generator.EmitJolt(h)
		Expect(src).To(ContainSubstring("type PhysicsSystem struct {"))
		Expect(src).To(ContainSubstring("func NewPhysicsSystem() *PhysicsSystem {"))
		Expect(src).To(ContainSubstring("runtime.SetFinalizer(p, (*PhysicsSystem).finalize)"))
		Expect(src).To(ContainSubstring("func (p *PhysicsSystem) Close() error {"))
		Expect(src).To(ContainSubstring("cjolt.JPH_PhysicsSystem_Destroy(p.h)"))
		Expect(src).To(ContainSubstring("runtime.SetFinalizer(p, nil)"))
	})
})
