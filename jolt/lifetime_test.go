package jolt_test

import (
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gataengine/jolt-go/jolt"
)

var _ = Describe("PhysicsSystem lifecycle", func() {

	BeforeEach(func() {
		Expect(jolt.Init()).To(BeTrue(), "JPH_Init must succeed")
	})

	AfterEach(func() {
		jolt.Shutdown()
	})

	// newSystem is a helper that wires up Mask-style layer filters and
	// constructs a PhysicsSystem with default capacities.
	newSystem := func() *jolt.PhysicsSystem {
		bpli := jolt.NewBroadPhaseLayerInterfaceMask(2)
		Expect(bpli).NotTo(BeNil())

		olpf := jolt.NewObjectLayerPairFilterMask()
		Expect(olpf).NotTo(BeNil())

		ovbplf := jolt.NewObjectVsBroadPhaseLayerFilterMask(bpli)
		Expect(ovbplf).NotTo(BeNil())

		sys := jolt.NewPhysicsSystemWithMaskFilters(
			10240, 65536, 10240,
			bpli, olpf, ovbplf,
		)
		Expect(sys).NotTo(BeNil())
		return sys
	}

	It("creates and closes without panic", func() {
		sys := newSystem()
		Expect(sys.Close()).To(Succeed())
	})

	It("Close is idempotent", func() {
		sys := newSystem()
		Expect(sys.Close()).To(Succeed())
		Expect(sys.Close()).To(Succeed()) // second call must not panic
	})

	It("finalizer rescues a forgotten Close", func() {
		func() {
			sys := newSystem()
			// Deliberately do NOT call sys.Close().
			// The finalizer registered by NewPhysicsSystem should clean up.
			_ = sys
		}()
		runtime.GC()
		runtime.GC()
		// If the finalizer panics, the Go runtime propagates it and the test
		// harness will fail.  Reaching this line means it ran cleanly.
	})
})
