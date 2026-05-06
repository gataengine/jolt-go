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
		// Constructor uses raw := then h := (cast to root type)
		Expect(src).To(ContainSubstring("raw := cjolt.JPH_BoxShape_Create("))
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
	It("emits constructor with correct root-type Destroy call", func() {
		src := generator.EmitJolt(h)
		Expect(src).To(ContainSubstring("func NewPhysicsMaterial() *PhysicsMaterial {"))
		// Critical: finalize must use PhysicsMaterial_Destroy, not Shape_Destroy
		Expect(src).To(ContainSubstring("cjolt.JPH_PhysicsMaterial_Destroy(s.h)"))
		Expect(src).NotTo(ContainSubstring("cjolt.JPH_Shape_Destroy(s.h)"))
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

var _ = Describe("EmitJolt — settings call site", func() {
	Context("BodyInterface.CreateBody passes settings via unsafe.Pointer cast", func() {
		h := &generator.Header{
			Funcs: []generator.Func{
				{Name: "JPH_BodyInterface_CreateBody", Return: "JPH_Body*",
					Params: []generator.Param{
						{Name: "self", CType: "JPH_BodyInterface *"},
						{Name: "settings", CType: "const JPH_BodyCreationSettings *"},
						{Name: "activation", CType: "JPH_Activation"},
					}},
			},
			Structs: []generator.Struct{
				{Name: "JPH_BodyInterface", Opaque: true},
				{Name: "JPH_Body", Opaque: true},
			},
		}
		It("emits CreateBody method with cast", func() {
			src := generator.EmitJolt(h)
			Expect(src).To(ContainSubstring(
				"func (bi *BodyInterface) CreateBody(s *BodyCreationSettings, a Activation) *Body {",
			))
			Expect(src).To(ContainSubstring(
				"cjolt.JPH_BodyInterface_CreateBody(bi.h, (*cjolt.C_JPH_BodyCreationSettings)(unsafe.Pointer(s)), cjolt.C_JPH_Activation(a))",
			))
			Expect(src).To(ContainSubstring("runtime.KeepAlive(s)"))
		})
	})
})

var _ = Describe("EmitJolt — settings struct", func() {
	// JPH_BodyCreationSettings is opaque in joltc (Create/Destroy + accessor methods,
	// no public field layout).  The generator must emit it as an Owned handle type
	// (not a layout-mirror struct) so call sites use s.h, not unsafe.Pointer(s).
	Context("BodyCreationSettings is opaque — emitted as owned handle", func() {
		var h *generator.Header
		BeforeEach(func() {
			var err error
			h, err = generator.ParseHeader("testdata/settings.h")
			Expect(err).NotTo(HaveOccurred())
		})
		It("emits BodyCreationSettings as an opaque handle struct (not layout-mirror)", func() {
			src := generator.EmitJolt(h)
			// Must have the opaque-handle form with a single .h field.
			Expect(src).To(ContainSubstring("type BodyCreationSettings struct {"))
			Expect(src).To(ContainSubstring("h *cjolt.C_JPH_BodyCreationSettings"))
			// Must NOT expose any C layout fields.
			Expect(src).NotTo(ContainSubstring("Position"))
			Expect(src).NotTo(ContainSubstring("AllowSleeping"))
		})
		It("emits constructor, Close and accessor methods for the opaque handle", func() {
			src := generator.EmitJolt(h)
			Expect(src).To(ContainSubstring("func NewBodyCreationSettings() *BodyCreationSettings {"))
			Expect(src).To(ContainSubstring("func (p *BodyCreationSettings) Close() error {"))
			// Accessor methods should use bcs.h, not unsafe.Pointer(bcs).
			Expect(src).To(ContainSubstring("func (bcs *BodyCreationSettings) Friction("))
			Expect(src).To(ContainSubstring("bcs.h"))
		})
	})
})

var _ = Describe("EmitJolt — ContactListener", func() {
	Context("ContactListener struct-of-funcs", func() {
		h := &generator.Header{}
		It("emits public Go struct with func fields", func() {
			src := generator.EmitJolt(h)
			Expect(src).To(ContainSubstring("type ContactListener struct {"))
			Expect(src).To(ContainSubstring(
				"OnContactValidate  func(b1, b2 *Body, baseOffset RVec3, result *CollideShapeResult) ValidateResult",
			))
			Expect(src).To(ContainSubstring(
				"OnContactAdded     func(b1, b2 *Body, manifold *ContactManifold, settings *ContactSettings)",
			))
			Expect(src).To(ContainSubstring(
				"OnContactPersisted func(b1, b2 *Body, manifold *ContactManifold, settings *ContactSettings)",
			))
			Expect(src).To(ContainSubstring(
				"OnContactRemoved   func(subShapePair *SubShapeIDPair)",
			))
			Expect(src).To(ContainSubstring("h      *cjolt.C_JPH_ContactListener"))
			Expect(src).To(ContainSubstring("handle cgo.Handle"))
		})
		It("emits NewContactListener registering cgo.Handle as userData", func() {
			src := generator.EmitJolt(h)
			Expect(src).To(ContainSubstring("func NewContactListener(l ContactListener) *ContactListener {"))
			// Task 31: factory now constructs a bridge ContactListenerHandle and stores
			// THAT in the cgo.Handle (not the public ContactListener), because the
			// cjolt-side trampolines expect *ContactListenerHandle as h.Value().
			Expect(src).To(ContainSubstring("bridge := &cjolt.ContactListenerHandle{"))
			Expect(src).To(ContainSubstring("cl.handle = cgo.NewHandle(bridge)"))
			Expect(src).To(ContainSubstring("cl.h = cjolt.JPH_ContactListener_Create(unsafe.Pointer(uintptr(cl.handle)))"))
			Expect(src).To(ContainSubstring("runtime.SetFinalizer(cl, (*ContactListener).finalize)"))
		})
	})
})
