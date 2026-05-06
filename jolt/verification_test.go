package jolt_test

// verification_test.go covers Tasks 36-39 of the MVP plan:
//   Task 36 – TempAllocator + JobSystem (create / Close / finalizer)
//   Task 37 – BodyCreationSettings construction + default field values
//   Task 38 – BoxShape construction + finalizer
//   Task 39 – All 11 "easy" MVP shape constructors (DescribeTable)
//
// The 6 deferred shapes (ConvexHull, Triangle, Mesh, HeightField,
// StaticCompound, MutableCompound) require substantial input data and are
// deferred to integration tests.

import (
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gataengine/jolt-go/jolt"
)

// ---------------------------------------------------------------------------
// Task 36: TempAllocator + JobSystem lifecycle
// ---------------------------------------------------------------------------

var _ = Describe("TempAllocator lifecycle", func() {
	BeforeEach(func() {
		Expect(jolt.Init()).To(BeTrue(), "JPH_Init must succeed")
	})
	AfterEach(func() {
		jolt.Shutdown()
	})

	It("creates and closes without panic", func() {
		ta := jolt.NewTempAllocator(8 * 1024 * 1024) // 8 MiB
		Expect(ta).NotTo(BeNil())
		Expect(ta.Close()).To(Succeed())
	})

	It("Close is idempotent", func() {
		ta := jolt.NewTempAllocator(4 * 1024 * 1024)
		Expect(ta).NotTo(BeNil())
		Expect(ta.Close()).To(Succeed())
		Expect(ta.Close()).To(Succeed())
	})

	It("finalizer rescues a forgotten Close", func() {
		func() {
			ta := jolt.NewTempAllocator(2 * 1024 * 1024)
			Expect(ta).NotTo(BeNil())
			_ = ta // no Close
		}()
		runtime.GC()
		runtime.GC()
	})
})

var _ = Describe("JobSystem lifecycle", func() {
	BeforeEach(func() {
		Expect(jolt.Init()).To(BeTrue(), "JPH_Init must succeed")
	})
	AfterEach(func() {
		jolt.Shutdown()
	})

	It("creates a thread-pool job system and closes without panic", func() {
		// maxJobs=2048, maxBarriers=8, numThreads=-1 (auto)
		js := jolt.NewJobSystemThreadPool(2048, 8, -1)
		Expect(js).NotTo(BeNil())
		Expect(js.Close()).To(Succeed())
	})

	It("Close is idempotent", func() {
		js := jolt.NewJobSystemThreadPool(512, 4, 1)
		Expect(js).NotTo(BeNil())
		Expect(js.Close()).To(Succeed())
		Expect(js.Close()).To(Succeed())
	})

	It("finalizer rescues a forgotten Close", func() {
		func() {
			js := jolt.NewJobSystemThreadPool(512, 4, 1)
			Expect(js).NotTo(BeNil())
			_ = js // no Close
		}()
		runtime.GC()
		runtime.GC()
	})
})

// ---------------------------------------------------------------------------
// Task 37: BodyCreationSettings construction + Jolt default field values
// ---------------------------------------------------------------------------

var _ = Describe("BodyCreationSettings", func() {
	BeforeEach(func() {
		Expect(jolt.Init()).To(BeTrue())
	})
	AfterEach(func() {
		jolt.Shutdown()
	})

	It("constructs without panic and can be closed", func() {
		bcs := jolt.NewBodyCreationSettings()
		Expect(bcs).NotTo(BeNil())
		Expect(bcs.Close()).To(Succeed())
	})

	It("has Jolt default Friction = 0.2", func() {
		bcs := jolt.NewBodyCreationSettings()
		defer bcs.Close() //nolint:errcheck
		Expect(bcs.Friction()).To(BeNumerically("~", 0.2, 1e-5))
	})

	It("has Jolt default LinearDamping = 0.05", func() {
		bcs := jolt.NewBodyCreationSettings()
		defer bcs.Close() //nolint:errcheck
		Expect(bcs.LinearDamping()).To(BeNumerically("~", 0.05, 1e-5))
	})

	It("has Jolt default AngularDamping = 0.05", func() {
		bcs := jolt.NewBodyCreationSettings()
		defer bcs.Close() //nolint:errcheck
		Expect(bcs.AngularDamping()).To(BeNumerically("~", 0.05, 1e-5))
	})
})

// ---------------------------------------------------------------------------
// Tasks 38 + 39: BoxShape + all 11 MVP shape constructors
// ---------------------------------------------------------------------------

var _ = Describe("Shape constructors", func() {
	BeforeEach(func() {
		Expect(jolt.Init()).To(BeTrue())
	})
	AfterEach(func() {
		jolt.Shutdown()
	})

	// Task 38: BoxShape — explicit test with finalizer exercise
	Describe("BoxShape", func() {
		It("constructs with valid half-extents and convex radius", func() {
			box := jolt.NewBoxShape(jolt.Vec3{X: 1, Y: 1, Z: 1}, 0.05)
			Expect(box).NotTo(BeNil())
		})

		It("finalizer runs without panic when Close is skipped", func() {
			func() {
				box := jolt.NewBoxShape(jolt.Vec3{X: 0.5, Y: 0.5, Z: 0.5}, 0.05)
				Expect(box).NotTo(BeNil())
				_ = box
			}()
			runtime.GC()
			runtime.GC()
		})
	})

	// Task 39: DescribeTable for all 11 "easy" MVP shapes
	//
	// Deferred (need substantial input data or compound sub-shapes):
	//   ConvexHull  — needs vertex array
	//   Triangle    — needs 3 explicit vertices (trivial but not a typical use case)
	//   Mesh        — needs triangle mesh data
	//   HeightField — needs height-field sample array
	//   StaticCompound  — needs pre-built sub-shape list
	//   MutableCompound — needs pre-built sub-shape list
	//
	// 8 simple convex shapes + 3 decorator shapes = 11 tested here.

	// Helper: construct a BoxShape and return its embedded *Shape for use by
	// decorator constructors. BoxShape embeds ConvexShape which embeds Shape,
	// so the Shape is addressable as &b.ConvexShape.Shape.
	box := func() (*jolt.BoxShape, *jolt.Shape) {
		b := jolt.NewBoxShape(jolt.Vec3{X: 0.5, Y: 0.5, Z: 0.5}, 0.05)
		return b, &b.ConvexShape.Shape
	}

	DescribeTable("simple convex shapes are non-nil after construction",
		func(construct func() interface{}) {
			shape := construct()
			Expect(shape).NotTo(BeNil())
		},
		Entry("Box",
			func() interface{} {
				return jolt.NewBoxShape(jolt.Vec3{X: 1, Y: 1, Z: 1}, 0.05)
			},
		),
		Entry("Sphere",
			func() interface{} { return jolt.NewSphereShape(1.0) },
		),
		Entry("Capsule",
			func() interface{} { return jolt.NewCapsuleShape(0.5, 0.3) },
		),
		Entry("TaperedCapsule",
			func() interface{} {
				settings := jolt.NewTaperedCapsuleShapeSettings(0.5, 0.3, 0.2)
				Expect(settings).NotTo(BeNil())
				return settings.CreateShape()
			},
		),
		Entry("Cylinder",
			func() interface{} { return jolt.NewCylinderShape(0.5, 0.3) },
		),
		Entry("TaperedCylinder",
			func() interface{} {
				settings := jolt.NewTaperedCylinderShapeSettings(0.5, 0.3, 0.2, 0.05)
				Expect(settings).NotTo(BeNil())
				return settings.CreateShape()
			},
		),
		Entry("Plane",
			func() interface{} {
				// GoPlane is a Go-native layout-compatible mirror of JPH_Plane.
				// NewPlaneShapeFromGoPlane is a handwritten helper that handles
				// the unsafe conversion internally.
				plane := jolt.GoPlane{Normal: jolt.Vec3{X: 0, Y: 1, Z: 0}, Distance: 0}
				return jolt.NewPlaneShapeFromGoPlane(plane, nil, 1000.0)
			},
		),
		Entry("Empty",
			func() interface{} {
				settings := jolt.NewEmptyShapeSettings(jolt.Vec3{X: 0, Y: 0, Z: 0})
				Expect(settings).NotTo(BeNil())
				return settings.CreateShape()
			},
		),
	)

	DescribeTable("decorator shapes wrap a BoxShape and are non-nil",
		func(construct func(*jolt.Shape) interface{}) {
			_, inner := box()
			shape := construct(inner)
			Expect(shape).NotTo(BeNil())
		},
		Entry("RotatedTranslated",
			func(inner *jolt.Shape) interface{} {
				return jolt.NewRotatedTranslatedShape(
					jolt.Vec3{X: 1, Y: 0, Z: 0},
					jolt.Quat{X: 0, Y: 0, Z: 0, W: 1}, // identity rotation
					inner,
				)
			},
		),
		Entry("Scaled",
			func(inner *jolt.Shape) interface{} {
				return jolt.NewScaledShape(inner, jolt.Vec3{X: 2, Y: 2, Z: 2})
			},
		),
		Entry("OffsetCenterOfMass",
			func(inner *jolt.Shape) interface{} {
				return jolt.NewOffsetCenterOfMassShape(jolt.Vec3{X: 0, Y: 0.1, Z: 0}, inner)
			},
		),
	)
})
