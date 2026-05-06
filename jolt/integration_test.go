package jolt_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gataengine/jolt-go/jolt"
)

var _ = Describe("Integration: drop-a-box", func() {
	It("box falls and rests on the floor", func() {
		// ----------------------------------------------------------------
		// Lifecycle
		// ----------------------------------------------------------------
		Expect(jolt.Init()).To(BeTrue())
		defer jolt.Shutdown()

		// ----------------------------------------------------------------
		// Layer filters — two BP layers: 0=NON_MOVING, 1=MOVING
		// ----------------------------------------------------------------
		bpli := jolt.NewBroadPhaseLayerInterfaceMask(2)
		jolt.SetupDefaultBroadPhaseLayers(bpli) // configures BP layers 0 and 1

		olpf := jolt.NewObjectLayerPairFilterMask()
		ovbplf := jolt.NewObjectVsBroadPhaseLayerFilterMask(bpli)

		// ----------------------------------------------------------------
		// PhysicsSystem
		// ----------------------------------------------------------------
		ps := jolt.NewPhysicsSystemWithMaskFilters(10240, 65536, 10240, bpli, olpf, ovbplf)
		defer ps.Close()

		// ----------------------------------------------------------------
		// TempAllocator (10 MiB) + JobSystem (thread pool)
		// ----------------------------------------------------------------
		ta := jolt.NewTempAllocator(10 * 1024 * 1024)
		defer ta.Close()
		// maxJobs=2048, maxBarriers=8, numThreads=-1 (auto)
		js := jolt.NewJobSystemThreadPool(2048, 8, -1)
		defer js.Close()

		bi := ps.BodyInterface()

		// ----------------------------------------------------------------
		// Static floor: 100×2×100 box centred at (0, -1, 0).
		// Top surface at y = 0.
		// ----------------------------------------------------------------
		floorHalfExt := jolt.Vec3{X: 50, Y: 1, Z: 50}
		floorShape := jolt.NewBoxShape(floorHalfExt, 0.05)
		floorSettings := jolt.NewBodyCreationSettingsWithShape(
			&floorShape.Shape,
			jolt.RVec3{X: 0, Y: -1, Z: 0},
			jolt.MotionTypeStatic,
			jolt.ObjectLayerNonMoving,
		)
		bi.CreateAndAddBody(floorSettings, jolt.ActivationDontActivate)

		// ----------------------------------------------------------------
		// Dynamic box: 1m cube (half-extents 0.5) starting at y = 10.
		// At rest it will sit with centre at y = 0.5.
		// ----------------------------------------------------------------
		boxHalfExt := jolt.Vec3{X: 0.5, Y: 0.5, Z: 0.5}
		boxShape := jolt.NewBoxShape(boxHalfExt, 0.05)
		boxSettings := jolt.NewBodyCreationSettingsWithShape(
			&boxShape.Shape,
			jolt.RVec3{X: 0, Y: 10, Z: 0},
			jolt.MotionTypeDynamic,
			jolt.ObjectLayerMoving,
		)
		boxID := bi.CreateAndAddBody(boxSettings, jolt.ActivationActivate)

		// ----------------------------------------------------------------
		// Step 120 frames @ 60 Hz  (= 2 seconds of simulation)
		// ----------------------------------------------------------------
		const dt float32 = 1.0 / 60.0
		for range 120 {
			ps.Update2(dt, 1, ta, js)
		}

		// ----------------------------------------------------------------
		// Assert: box centre should be at y ≈ 0.5 (resting on floor)
		// ----------------------------------------------------------------
		finalPos := bi.GetPosition(boxID)
		Expect(float64(finalPos.Y)).To(BeNumerically("~", 0.5, 0.05),
			"box Y position should be ~0.5 after falling onto floor")
	})
})
