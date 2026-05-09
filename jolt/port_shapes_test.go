package jolt_test

// port_shapes_test.go ports third_party/joltc/tests/test_shapes.cpp.

import (
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gataengine/jolt-go/jolt"
)

var _ = Describe("ShapesTest (joltc test_shapes.cpp)", func() {
	BeforeEach(func() {
		Expect(jolt.Init()).To(BeTrue())
	})
	AfterEach(func() {
		jolt.Shutdown()
	})

	Describe("SphereShape", func() {
		It("SphereShape_Create", func() {
			shape := jolt.NewSphereShape(1.0)
			Expect(shape).NotTo(BeNil())
			Expect(shape.Radius()).To(BeNumerically("~", 1.0, 1e-5))
		})

		It("SphereShape_CreateWithSettings", func() {
			settings := jolt.NewSphereShapeSettings(2.0)
			Expect(settings).NotTo(BeNil())
			Expect(settings.Radius()).To(BeNumerically("~", 2.0, 1e-5))

			shape := settings.CreateShape()
			Expect(shape).NotTo(BeNil())
			Expect(shape.Radius()).To(BeNumerically("~", 2.0, 1e-5))
		})

		It("SphereShapeSettings_SetRadius", func() {
			settings := jolt.NewSphereShapeSettings(1.0)
			Expect(settings).NotTo(BeNil())

			settings.SetRadius(3.0)
			Expect(settings.Radius()).To(BeNumerically("~", 3.0, 1e-5))
		})

		It("SphereShape_GetType", func() {
			shape := jolt.NewSphereShape(1.0)
			Expect(shape.Type()).To(Equal(jolt.ShapeTypeConvex))
			Expect(shape.SubType()).To(Equal(jolt.ShapeSubTypeSphere))
		})

		It("SphereShape_GetVolume", func() {
			shape := jolt.NewSphereShape(1.0)
			expected := float32(4.0/3.0) * float32(math.Pi)
			Expect(shape.Volume()).To(BeNumerically("~", expected, 1e-3))
		})
	})

	Describe("BoxShape", func() {
		It("BoxShape_Create", func() {
			shape := jolt.NewBoxShape(jolt.Vec3{X: 1, Y: 2, Z: 3}, jolt.DefaultConvexRadius)
			Expect(shape).NotTo(BeNil())

			h := shape.GetHalfExtent()
			Expect(h.X).To(BeNumerically("~", 1.0, 1e-5))
			Expect(h.Y).To(BeNumerically("~", 2.0, 1e-5))
			Expect(h.Z).To(BeNumerically("~", 3.0, 1e-5))
		})

		It("BoxShape_CreateWithSettings", func() {
			settings := jolt.NewBoxShapeSettings(jolt.Vec3{X: 2, Y: 2, Z: 2}, jolt.DefaultConvexRadius)
			Expect(settings).NotTo(BeNil())

			shape := settings.CreateShape()
			Expect(shape).NotTo(BeNil())

			h := shape.GetHalfExtent()
			Expect(h.X).To(BeNumerically("~", 2.0, 1e-5))
			Expect(h.Y).To(BeNumerically("~", 2.0, 1e-5))
			Expect(h.Z).To(BeNumerically("~", 2.0, 1e-5))
		})

		It("BoxShape_GetConvexRadius", func() {
			shape := jolt.NewBoxShape(jolt.Vec3{X: 1, Y: 1, Z: 1}, 0.1)
			Expect(shape.ConvexRadius()).To(BeNumerically("~", 0.1, 1e-5))
		})

		It("BoxShape_GetType", func() {
			shape := jolt.NewBoxShape(jolt.Vec3{X: 1, Y: 1, Z: 1}, jolt.DefaultConvexRadius)
			Expect(shape.Type()).To(Equal(jolt.ShapeTypeConvex))
			Expect(shape.SubType()).To(Equal(jolt.ShapeSubTypeBox))
		})
	})

	Describe("CapsuleShape", func() {
		It("CapsuleShape_Create", func() {
			shape := jolt.NewCapsuleShape(1.0, 0.5)
			Expect(shape).NotTo(BeNil())
			Expect(shape.HalfHeightOfCylinder()).To(BeNumerically("~", 1.0, 1e-5))
			Expect(shape.Radius()).To(BeNumerically("~", 0.5, 1e-5))
		})

		It("CapsuleShape_CreateWithSettings", func() {
			settings := jolt.NewCapsuleShapeSettings(2.0, 1.0)
			Expect(settings).NotTo(BeNil())

			shape := settings.CreateShape()
			Expect(shape).NotTo(BeNil())
			Expect(shape.HalfHeightOfCylinder()).To(BeNumerically("~", 2.0, 1e-5))
			Expect(shape.Radius()).To(BeNumerically("~", 1.0, 1e-5))
		})

		It("CapsuleShape_GetType", func() {
			shape := jolt.NewCapsuleShape(1.0, 0.5)
			Expect(shape.SubType()).To(Equal(jolt.ShapeSubTypeCapsule))
		})
	})

	Describe("CylinderShape", func() {
		It("CylinderShape_Create", func() {
			shape := jolt.NewCylinderShape(1.0, 0.5)
			Expect(shape).NotTo(BeNil())
			Expect(shape.HalfHeight()).To(BeNumerically("~", 1.0, 1e-5))
			Expect(shape.Radius()).To(BeNumerically("~", 0.5, 1e-5))
		})

		It("CylinderShape_CreateWithSettings", func() {
			settings := jolt.NewCylinderShapeSettings(2.0, 1.0, jolt.DefaultConvexRadius)
			Expect(settings).NotTo(BeNil())

			shape := settings.CreateShape()
			Expect(shape).NotTo(BeNil())
			Expect(shape.HalfHeight()).To(BeNumerically("~", 2.0, 1e-5))
			Expect(shape.Radius()).To(BeNumerically("~", 1.0, 1e-5))
		})

		It("CylinderShape_GetType", func() {
			shape := jolt.NewCylinderShape(1.0, 0.5)
			Expect(shape.SubType()).To(Equal(jolt.ShapeSubTypeCylinder))
		})
	})

	Describe("TriangleShape", func() {
		v1 := jolt.Vec3{X: 0, Y: 0, Z: 0}
		v2 := jolt.Vec3{X: 1, Y: 0, Z: 0}
		v3 := jolt.Vec3{X: 0, Y: 1, Z: 0}

		It("TriangleShape_Create", func() {
			shape := jolt.NewTriangleShape(v1, v2, v3, 0.0)
			Expect(shape).NotTo(BeNil())

			Expect(shape.GetVertex1()).To(Equal(v1))
			Expect(shape.GetVertex2()).To(Equal(v2))
			Expect(shape.GetVertex3()).To(Equal(v3))
		})

		It("TriangleShape_GetType", func() {
			shape := jolt.NewTriangleShape(v1, v2, v3, 0.0)
			Expect(shape.SubType()).To(Equal(jolt.ShapeSubTypeTriangle))
		})
	})

	Describe("TaperedCapsuleShape", func() {
		It("TaperedCapsuleShape_Create", func() {
			settings := jolt.NewTaperedCapsuleShapeSettings(1.0, 0.5, 0.3)
			Expect(settings).NotTo(BeNil())

			shape := settings.CreateShape()
			Expect(shape).NotTo(BeNil())
			Expect(shape.TopRadius()).To(BeNumerically("~", 0.5, 1e-5))
			Expect(shape.BottomRadius()).To(BeNumerically("~", 0.3, 1e-5))
			Expect(shape.HalfHeight()).To(BeNumerically("~", 1.0, 1e-5))
		})

		It("TaperedCapsuleShape_GetType", func() {
			settings := jolt.NewTaperedCapsuleShapeSettings(1.0, 0.5, 0.3)
			shape := settings.CreateShape()
			Expect(shape.SubType()).To(Equal(jolt.ShapeSubTypeTaperedCapsule))
		})
	})

	Describe("TaperedCylinderShape", func() {
		It("TaperedCylinderShape_Create", func() {
			settings := jolt.NewTaperedCylinderShapeSettings(1.0, 0.5, 1.0, jolt.DefaultConvexRadius)
			Expect(settings).NotTo(BeNil())

			shape := settings.CreateShape()
			Expect(shape).NotTo(BeNil())
			Expect(shape.TopRadius()).To(BeNumerically("~", 0.5, 1e-5))
			Expect(shape.BottomRadius()).To(BeNumerically("~", 1.0, 1e-5))
			Expect(shape.HalfHeight()).To(BeNumerically("~", 1.0, 1e-5))
		})
	})

	Describe("ConvexHullShape", func() {
		// Tetrahedron — same vertices as the joltc test.
		points := []jolt.Vec3{
			{X: 0, Y: 0, Z: 0},
			{X: 1, Y: 0, Z: 0},
			{X: 0.5, Y: 1, Z: 0},
			{X: 0.5, Y: 0.5, Z: 1},
		}

		It("ConvexHullShape_Create", func() {
			settings := jolt.NewConvexHullShapeSettings(points, jolt.DefaultConvexRadius)
			Expect(settings).NotTo(BeNil())

			shape := settings.CreateShape()
			Expect(shape).NotTo(BeNil())
			Expect(shape.NumPoints()).To(Equal(uint32(4)))
			Expect(shape.NumFaces()).To(Equal(uint32(4))) // tetrahedron
		})

		It("ConvexHullShape_GetPoint", func() {
			settings := jolt.NewConvexHullShapeSettings(points, jolt.DefaultConvexRadius)
			shape := settings.CreateShape()
			Expect(shape).NotTo(BeNil())

			// Hull may reorder points; just assert non-NaN.
			p := shape.GetPoint(0)
			Expect(math.IsNaN(float64(p.X))).To(BeFalse())
			Expect(math.IsNaN(float64(p.Y))).To(BeFalse())
			Expect(math.IsNaN(float64(p.Z))).To(BeFalse())
		})
	})

	Describe("PlaneShape", func() {
		It("PlaneShape_Create", func() {
			plane := jolt.GoPlane{Normal: jolt.Vec3{X: 0, Y: 1, Z: 0}, Distance: 0}
			shape := jolt.NewPlaneShapeFromGoPlane(plane, nil, 1000.0)
			Expect(shape).NotTo(BeNil())

			result := shape.GetPlane()
			Expect(result.Normal.X).To(BeNumerically("~", 0.0, 1e-5))
			Expect(result.Normal.Y).To(BeNumerically("~", 1.0, 1e-5))
			Expect(result.Normal.Z).To(BeNumerically("~", 0.0, 1e-5))
			Expect(result.Distance).To(BeNumerically("~", 0.0, 1e-5))

			Expect(shape.HalfExtent()).To(BeNumerically("~", 1000.0, 1e-3))
		})
	})

	Describe("Shape (base)", func() {
		It("Shape_UserData", func() {
			shape := jolt.NewSphereShape(1.0)
			shape.SetUserData(12345)
			Expect(shape.UserData()).To(Equal(uint64(12345)))
		})

		It("Shape_GetLocalBounds", func() {
			shape := jolt.NewSphereShape(1.0)
			b := shape.GetLocalBounds()
			Expect(b.Min.X).To(BeNumerically("~", -1.0, 1e-5))
			Expect(b.Min.Y).To(BeNumerically("~", -1.0, 1e-5))
			Expect(b.Min.Z).To(BeNumerically("~", -1.0, 1e-5))
			Expect(b.Max.X).To(BeNumerically("~", 1.0, 1e-5))
			Expect(b.Max.Y).To(BeNumerically("~", 1.0, 1e-5))
			Expect(b.Max.Z).To(BeNumerically("~", 1.0, 1e-5))
		})

		It("Shape_GetCenterOfMass", func() {
			shape := jolt.NewSphereShape(1.0)
			com := shape.GetCenterOfMass()
			Expect(com.X).To(BeNumerically("~", 0.0, 1e-5))
			Expect(com.Y).To(BeNumerically("~", 0.0, 1e-5))
			Expect(com.Z).To(BeNumerically("~", 0.0, 1e-5))
		})

		It("Shape_GetInnerRadius", func() {
			shape := jolt.NewSphereShape(1.0)
			Expect(shape.InnerRadius()).To(BeNumerically("~", 1.0, 1e-5))
		})

		It("Shape_MustBeStatic", func() {
			sphere := jolt.NewSphereShape(1.0)
			Expect(sphere.MustBeStatic()).To(BeFalse())

			plane := jolt.GoPlane{Normal: jolt.Vec3{X: 0, Y: 1, Z: 0}, Distance: 0}
			planeShape := jolt.NewPlaneShapeFromGoPlane(plane, nil, 1000.0)
			Expect(planeShape.MustBeStatic()).To(BeTrue())
		})

		It("Shape_GetMassProperties", Pending, func() {
			// MassProperties accessors not yet emitted by the generator.
			// joltc test asserts massProps.mass > 0.
		})
	})
})
