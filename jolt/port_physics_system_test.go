package jolt_test

// port_physics_system_test.go ports
// third_party/joltc/tests/test_physics_system.cpp.

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gataengine/jolt-go/jolt"
)

var _ = Describe("PhysicsSystemTest (joltc test_physics_system.cpp)", func() {
	var (
		js *jolt.JobSystem
		ps *portSystem
		bi *jolt.BodyInterface
	)

	BeforeEach(func() {
		Expect(jolt.Init()).To(BeTrue())
		js = jolt.NewJobSystemThreadPoolDefault()
		ps = newTwoLayerSystem()
		bi = ps.System.BodyInterface()
	})
	AfterEach(func() {
		ps.Close()
		Expect(js.Close()).To(Succeed())
		jolt.Shutdown()
	})

	// ------------------------------------------------------------------
	// System
	// ------------------------------------------------------------------

	It("Create_ValidSettings", func() {
		Expect(ps.System).NotTo(BeNil())
	})

	It("GetBodyInterface", func() {
		Expect(ps.System.BodyInterface()).NotTo(BeNil())
	})

	It("GetBodyInterfaceNoLock", func() {
		Expect(ps.System.BodyInterfaceNoLock()).NotTo(BeNil())
	})

	It("SetGravity", func() {
		ps.System.SetGravity(jolt.Vec3{X: 0, Y: -10, Z: 0})
		g := ps.System.GetGravity()
		Expect(g.X).To(BeNumerically("~", 0.0, 1e-5))
		Expect(g.Y).To(BeNumerically("~", -10.0, 1e-5))
		Expect(g.Z).To(BeNumerically("~", 0.0, 1e-5))
	})

	It("GetNumBodies_Initial", func() {
		Expect(ps.System.NumBodies()).To(Equal(uint32(0)))
	})

	It("GetNumActiveBodies_Initial", func() {
		Expect(ps.System.NumActiveBodies(jolt.BodyTypeRigid)).To(Equal(uint32(0)))
	})

	It("GetMaxBodies", func() {
		// newTwoLayerSystem allocates 1024 bodies (matching joltc).
		Expect(ps.System.MaxBodies()).To(Equal(uint32(1024)))
	})

	// ------------------------------------------------------------------
	// Body lifecycle
	// ------------------------------------------------------------------

	It("CreateBody_Sphere", func() {
		shape := jolt.NewSphereShape(1.0)
		settings := jolt.NewBodyCreationSettingsWithShape(
			&shape.Shape,
			jolt.RVec3{X: 0, Y: 10, Z: 0},
			jolt.MotionTypeDynamic, jolt.ObjectLayerMoving,
		)

		body := bi.CreateBody(settings)
		Expect(body).NotTo(BeNil())

		bodyID := body.ID()
		Expect(uint32(bodyID)).NotTo(Equal(uint32(0xFFFFFFFF))) // JPH_INVALID_BODY_ID
		Expect(ps.System.NumBodies()).To(Equal(uint32(1)))

		bi.AddBody(bodyID, jolt.ActivationActivate)
		Expect(ps.System.NumActiveBodies(jolt.BodyTypeRigid)).To(Equal(uint32(1)))

		bi.RemoveAndDestroyBody(bodyID)
	})

	It("CreateBody_Box", func() {
		shape := jolt.NewBoxShape(jolt.Vec3{X: 1, Y: 1, Z: 1}, jolt.DefaultConvexRadius)
		settings := jolt.NewBodyCreationSettingsWithShape(
			&shape.Shape,
			jolt.RVec3{X: 0, Y: 0, Z: 0},
			jolt.MotionTypeStatic, jolt.ObjectLayerNonMoving,
		)

		body := bi.CreateBody(settings)
		Expect(body).NotTo(BeNil())

		bodyID := body.ID()
		bi.AddBody(bodyID, jolt.ActivationDontActivate)
		Expect(ps.System.NumActiveBodies(jolt.BodyTypeRigid)).To(Equal(uint32(0)))

		bi.RemoveAndDestroyBody(bodyID)
	})

	// ------------------------------------------------------------------
	// Body kinematics
	// ------------------------------------------------------------------

	It("Body_GetPosition", func() {
		shape := jolt.NewSphereShape(1.0)
		settings := jolt.NewBodyCreationSettingsWithShape(
			&shape.Shape,
			jolt.RVec3{X: 1, Y: 2, Z: 3},
			jolt.MotionTypeDynamic, jolt.ObjectLayerMoving,
		)
		body := bi.CreateBody(settings)
		bodyID := body.ID()
		bi.AddBody(bodyID, jolt.ActivationActivate)

		p := bi.GetPosition(bodyID)
		Expect(float32(p.X)).To(BeNumerically("~", 1.0, 1e-5))
		Expect(float32(p.Y)).To(BeNumerically("~", 2.0, 1e-5))
		Expect(float32(p.Z)).To(BeNumerically("~", 3.0, 1e-5))

		bi.RemoveAndDestroyBody(bodyID)
	})

	It("Body_SetPosition", func() {
		shape := jolt.NewSphereShape(1.0)
		settings := jolt.NewBodyCreationSettingsWithShape(
			&shape.Shape,
			jolt.RVec3{X: 0, Y: 0, Z: 0},
			jolt.MotionTypeDynamic, jolt.ObjectLayerMoving,
		)
		body := bi.CreateBody(settings)
		bodyID := body.ID()
		bi.AddBody(bodyID, jolt.ActivationActivate)

		bi.SetPosition(bodyID, jolt.RVec3{X: 5, Y: 10, Z: 15}, jolt.ActivationActivate)
		p := bi.GetPosition(bodyID)
		Expect(float32(p.X)).To(BeNumerically("~", 5.0, 1e-5))
		Expect(float32(p.Y)).To(BeNumerically("~", 10.0, 1e-5))
		Expect(float32(p.Z)).To(BeNumerically("~", 15.0, 1e-5))

		bi.RemoveAndDestroyBody(bodyID)
	})
})
