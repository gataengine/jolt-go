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
})
