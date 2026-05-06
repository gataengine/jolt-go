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

var _ = Describe("EmitJolt — ref-counted types", func() {
	// Fixture: Shape (root ref-counted), ConvexShape (child of Shape),
	// BoxShape (child of ConvexShape) with a Create function taking
	// a Vec3 pointer and float.
	h := &generator.Header{
		Structs: []generator.Struct{
			{Name: "JPH_Shape", Opaque: true},
			{Name: "JPH_ConvexShape", Opaque: true},
			{Name: "JPH_BoxShape", Opaque: true},
		},
		Funcs: []generator.Func{
			{
				Name:   "JPH_BoxShape_Create",
				Return: "JPH_Shape*",
				Params: []generator.Param{
					{Name: "halfExtent", CType: "const JPH_Vec3 *"},
					{Name: "convexRadius", CType: "float"},
				},
			},
			{
				Name:   "JPH_Shape_Release",
				Return: "void",
				Params: []generator.Param{{Name: "self", CType: "JPH_Shape *"}},
			},
		},
	}

	It("emits struct type hierarchy for ref-counted types", func() {
		src := generator.EmitJolt(h)
		// Root type holds the handle (gofmt uses compact single-field form)
		Expect(src).To(ContainSubstring("type Shape struct{"))
		Expect(src).To(ContainSubstring("h *cjolt.C_JPH_Shape"))
		// Child with embedded parent
		Expect(src).To(ContainSubstring("type ConvexShape struct{"))
		Expect(src).To(ContainSubstring("Shape"))
		// Deeper child
		Expect(src).To(ContainSubstring("type BoxShape struct{"))
		Expect(src).To(ContainSubstring("ConvexShape"))
	})

	It("emits constructor with finalizer for BoxShape_Create", func() {
		src := generator.EmitJolt(h)
		Expect(src).To(ContainSubstring("func NewBoxShape(halfExtent Vec3, convexRadius float32) *BoxShape {"))
		Expect(src).To(ContainSubstring("h := cjolt.JPH_BoxShape_Create("))
		Expect(src).To(ContainSubstring("(*cjolt.C_JPH_Vec3)(unsafe.Pointer(&halfExtent))"))
		Expect(src).To(ContainSubstring("runtime.SetFinalizer(s, (*BoxShape).finalize)"))
		// Embedded chain literal
		Expect(src).To(ContainSubstring("ConvexShape: ConvexShape{Shape: Shape{h: h}}"))
	})

	It("does NOT emit public AddRef or Release methods", func() {
		src := generator.EmitJolt(h)
		Expect(src).NotTo(ContainSubstring("func (s *BoxShape) AddRef()"))
		Expect(src).NotTo(ContainSubstring("func (s *BoxShape) Release()"))
		Expect(src).NotTo(ContainSubstring("func (s *Shape) AddRef()"))
		Expect(src).NotTo(ContainSubstring("func (s *Shape) Release()"))
	})
})

var _ = Describe("EmitJolt — non-Shape ref-counted root", func() {
	h := &generator.Header{
		Funcs: []generator.Func{
			{Name: "JPH_PhysicsMaterial_Create", Return: "JPH_PhysicsMaterial*"},
			{Name: "JPH_PhysicsMaterial_AddRef", Return: "void",
				Params: []generator.Param{{Name: "self", CType: "JPH_PhysicsMaterial *"}}},
			{Name: "JPH_PhysicsMaterial_Release", Return: "void",
				Params: []generator.Param{{Name: "self", CType: "JPH_PhysicsMaterial *"}}},
		},
	}
	It("emits constructor with correct root-type Release call", func() {
		src := generator.EmitJolt(h)
		Expect(src).To(ContainSubstring("func NewPhysicsMaterial() *PhysicsMaterial {"))
		// Critical: finalize must use PhysicsMaterial_Release, not Shape_Release
		Expect(src).To(ContainSubstring("cjolt.JPH_PhysicsMaterial_Release(s.h)"))
		Expect(src).NotTo(ContainSubstring("cjolt.JPH_Shape_Release(s.h)"))
	})
})

var _ = Describe("EmitJolt — methods on receiver types", func() {
	Context("methods on receiver types", func() {
		h := &generator.Header{
			Funcs: []generator.Func{
				{Name: "JPH_Body_GetPosition", Return: "JPH_RVec3",
					Params: []generator.Param{{Name: "self", CType: "const JPH_Body *"}}},
				{Name: "JPH_Body_SetLinearVelocity", Return: "void",
					Params: []generator.Param{
						{Name: "self", CType: "JPH_Body *"},
						{Name: "v", CType: "const JPH_Vec3 *"},
					}},
			},
			Structs: []generator.Struct{{Name: "JPH_Body", Opaque: true}},
		}
		It("emits getter returning value", func() {
			src := generator.EmitJolt(h)
			Expect(src).To(ContainSubstring("func (b *Body) Position() RVec3 {"))
			Expect(src).To(ContainSubstring("return *(*RVec3)(unsafe.Pointer(&out))"))
		})
		It("emits setter taking value, KeepAlive on receiver", func() {
			src := generator.EmitJolt(h)
			Expect(src).To(ContainSubstring("func (b *Body) SetLinearVelocity(v Vec3) {"))
			Expect(src).To(ContainSubstring("cjolt.JPH_Body_SetLinearVelocity(b.h, (*cjolt.C_JPH_Vec3)(unsafe.Pointer(&v)))"))
			Expect(src).To(ContainSubstring("runtime.KeepAlive(b)"))
		})
	})
})
