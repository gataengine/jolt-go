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
	)

	BeforeEach(func() {
		Expect(jolt.Init()).To(BeTrue())
		js = jolt.NewJobSystemThreadPoolDefault()
		ps = newTwoLayerSystem()
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
})
