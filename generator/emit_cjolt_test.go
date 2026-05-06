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

var _ = Describe("EmitCjolt — math types", func() {
	Context("math types", func() {
		h := &generator.Header{
			Funcs: []generator.Func{
				{
					Name:   "JPH_Body_GetPosition",
					Return: "JPH_RVec3",
					Params: []generator.Param{{Name: "self", CType: "const JPH_Body *"}},
				},
				{
					Name:   "JPH_Body_SetLinearVelocity",
					Return: "void",
					Params: []generator.Param{
						{Name: "self", CType: "JPH_Body *"},
						{Name: "v", CType: "const JPH_Vec3 *"},
					},
				},
			},
		}
		It("returns RVec3 by value, taking the C type and passing through", func() {
			src := generator.EmitCjolt(h)
			Expect(src).To(ContainSubstring(
				"func JPH_Body_GetPosition(self *C.JPH_Body) C.JPH_RVec3 {",
			))
		})
		It("accepts JPH_Vec3 pointer through unsafe.Pointer cast at higher layer", func() {
			src := generator.EmitCjolt(h)
			// cjolt is "raw"; it accepts the C struct pointer directly. The unsafe
			// cast happens in the public API. Verify the wrapper signature here.
			Expect(src).To(ContainSubstring(
				"func JPH_Body_SetLinearVelocity(self *C.JPH_Body, v *C.JPH_Vec3) {",
			))
		})
	})
})

var _ = Describe("EmitCjolt — scalar Go-native types", func() {
	Context("float and double params use Go-native types", func() {
		h := &generator.Header{
			Funcs: []generator.Func{
				{
					Name:   "JPH_BoxShape_Create",
					Return: "JPH_Shape*",
					Params: []generator.Param{
						{Name: "halfExtent", CType: "const JPH_Vec3 *"},
						{Name: "convexRadius", CType: "float"},
					},
				},
			},
		}
		It("emits float32 for float param with C.float cast in body", func() {
			src := generator.EmitCjolt(h)
			Expect(src).To(ContainSubstring(
				"func JPH_BoxShape_Create(halfExtent *C.JPH_Vec3, convexRadius float32) *C.JPH_Shape {",
			))
			Expect(src).To(ContainSubstring(
				"return C.JPH_BoxShape_Create(halfExtent, C.float(convexRadius))",
			))
		})
	})

	Context("integer scalar params use Go-native types", func() {
		h := &generator.Header{
			Funcs: []generator.Func{
				{
					Name:   "JPH_Foo_SetCount",
					Return: "void",
					Params: []generator.Param{
						{Name: "n", CType: "int"},
					},
				},
				{
					Name:   "JPH_Foo_SetUCount",
					Return: "void",
					Params: []generator.Param{
						{Name: "n", CType: "unsigned int"},
					},
				},
			},
		}
		It("emits int32 for int param with C.int cast", func() {
			src := generator.EmitCjolt(h)
			Expect(src).To(ContainSubstring(
				"func JPH_Foo_SetCount(n int32) { C.JPH_Foo_SetCount(C.int(n)) }",
			))
		})
		It("emits uint32 for unsigned int param with C.uint cast", func() {
			src := generator.EmitCjolt(h)
			Expect(src).To(ContainSubstring(
				"func JPH_Foo_SetUCount(n uint32) { C.JPH_Foo_SetUCount(C.uint(n)) }",
			))
		})
	})

	Context("JPH_* enum types still use C types (no Go-native mapping)", func() {
		h := &generator.Header{
			Funcs: []generator.Func{
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
		It("keeps C.JPH_MotionType for enum params", func() {
			src := generator.EmitCjolt(h)
			Expect(src).To(ContainSubstring(
				"func JPH_Body_SetMotionType(self *C.JPH_Body, type_ C.JPH_MotionType) {",
			))
		})
	})
})

var _ = Describe("EmitCjolt — opaque type re-exports", func() {
	Context("opaque type re-exports", func() {
		h := &generator.Header{
			Structs: []generator.Struct{
				{Name: "JPH_PhysicsSystem", Opaque: true},
				{Name: "JPH_BodyInterface", Opaque: true},
			},
		}
		It("emits Go aliases for each opaque type", func() {
			src := generator.EmitCjolt(h)
			Expect(src).To(ContainSubstring("type C_JPH_PhysicsSystem = C.JPH_PhysicsSystem"))
			Expect(src).To(ContainSubstring("type C_JPH_BodyInterface = C.JPH_BodyInterface"))
		})
	})
})

var _ = Describe("EmitCjolt — listener trampolines", func() {
	Context("listener trampolines", func() {
		h := &generator.Header{}
		It("emits ContactListenerHandle struct with C.*-typed func fields", func() {
			src := generator.EmitCjolt(h)
			Expect(src).To(ContainSubstring("type ContactListenerHandle struct {"))
			Expect(src).To(ContainSubstring(
				"OnContactAdded     func(b1, b2 *C.JPH_Body, manifold *C.JPH_ContactManifold, settings *C.JPH_ContactSettings)",
			))
		})
		It("emits exported trampolines and init() stub (SetProcs wired in C, Task 41)", func() {
			src := generator.EmitCjolt(h)
			Expect(src).To(ContainSubstring("//export jolt_go_contactListener_OnContactAdded"))
			Expect(src).To(ContainSubstring(
				"func jolt_go_contactListener_OnContactAdded(userData unsafe.Pointer, b1, b2 *C.JPH_Body, manifold *C.JPH_ContactManifold, settings *C.JPH_ContactSettings) {",
			))
			Expect(src).To(ContainSubstring(`h := cgo.Handle(uintptr(userData))`))
			Expect(src).To(ContainSubstring(`cl := h.Value().(*ContactListenerHandle)`))
			Expect(src).To(ContainSubstring("func init() {"))
			// SetProcs registration is handled by hand-written C file (Task 41);
			// the generated init() is intentionally an empty stub.
			Expect(src).To(ContainSubstring("TODO(Task 41)"))
		})
	})
})
