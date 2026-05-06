package jolt_test

import (
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gataengine/jolt-go/jolt"
)

var _ = Describe("Integration: ContactListener fires", func() {
	It("OnContactAdded is invoked when two bodies collide", func() {
		// ----------------------------------------------------------------
		// Lifecycle
		// ----------------------------------------------------------------
		Expect(jolt.Init()).To(BeTrue())
		defer jolt.Shutdown()

		// ----------------------------------------------------------------
		// Layer filters — same two-layer setup as drop-a-box test.
		// ----------------------------------------------------------------
		bpli := jolt.NewBroadPhaseLayerInterfaceMask(2)
		jolt.SetupDefaultBroadPhaseLayers(bpli)

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
		js := jolt.NewJobSystemThreadPool(2048, 8, -1)
		defer js.Close()

		// ----------------------------------------------------------------
		// ContactListener — count OnContactAdded calls.
		// The callback may be invoked from Jolt worker threads, so guard
		// the counter with a mutex.
		// ----------------------------------------------------------------
		var (
			mu       sync.Mutex
			contacts int
		)
		listener := jolt.NewContactListener(jolt.ContactListener{
			OnContactAdded: func(b1, b2 *jolt.Body, _ *jolt.ContactManifold, _ *jolt.ContactSettings) {
				mu.Lock()
				defer mu.Unlock()
				contacts++
			},
		})
		// ContactListener has no Close() — the GC finalizer handles cleanup.
		ps.SetContactListener(listener)

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
		// Dynamic box: 1 m cube (half-extents 0.5) starting at y = 10.
		// It will fall and hit the floor.
		// ----------------------------------------------------------------
		boxHalfExt := jolt.Vec3{X: 0.5, Y: 0.5, Z: 0.5}
		boxShape := jolt.NewBoxShape(boxHalfExt, 0.05)
		boxSettings := jolt.NewBodyCreationSettingsWithShape(
			&boxShape.Shape,
			jolt.RVec3{X: 0, Y: 10, Z: 0},
			jolt.MotionTypeDynamic,
			jolt.ObjectLayerMoving,
		)
		bi.CreateAndAddBody(boxSettings, jolt.ActivationActivate)

		// ----------------------------------------------------------------
		// Step 120 frames @ 60 Hz  (= 2 seconds of simulation)
		// ----------------------------------------------------------------
		const dt float32 = 1.0 / 60.0
		for range 120 {
			ps.Update2(dt, 1, ta, js)
		}

		// ----------------------------------------------------------------
		// Assert: OnContactAdded must have been called at least once.
		// ----------------------------------------------------------------
		mu.Lock()
		finalContacts := contacts
		mu.Unlock()

		Expect(finalContacts).To(BeNumerically(">", 0),
			"OnContactAdded should be invoked at least once when the box hits the floor")
	})
})
