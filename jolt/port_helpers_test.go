package jolt_test

// port_helpers_test.go — shared fixtures for the joltc port suite
// (port_core_test.go, port_shapes_test.go, port_physics_system_test.go).
//
// These mirror the C++ test fixtures in third_party/joltc/tests/
// (test_*.cpp) but use the Mask-scheme broad-phase / object-layer
// filters already wired up in our handwritten code, for consistency
// with integration_test.go.

import (
	. "github.com/onsi/gomega"

	"github.com/gataengine/jolt-go/jolt"
)

// portSystem bundles a PhysicsSystem with the layer-filter handles that
// must outlive it. Callers MUST call portSystem.Close() (typically via
// defer) to release the system before Shutdown.
type portSystem struct {
	System *jolt.PhysicsSystem
	bpli   *jolt.BroadPhaseLayerInterface
	olpf   *jolt.ObjectLayerPairFilter
	ovbplf *jolt.ObjectVsBroadPhaseLayerFilter
}

// newTwoLayerSystem creates a 2-layer (NON_MOVING / MOVING) PhysicsSystem
// with capacities matching the joltc test fixtures (1024 bodies, 1024 body
// pairs, 1024 contact constraints).
func newTwoLayerSystem() *portSystem {
	bpli := jolt.NewBroadPhaseLayerInterfaceMask(2)
	jolt.SetupDefaultBroadPhaseLayers(bpli)

	olpf := jolt.NewObjectLayerPairFilterMask()
	ovbplf := jolt.NewObjectVsBroadPhaseLayerFilterMask(bpli)

	sys := jolt.NewPhysicsSystemWithMaskFilters(1024, 1024, 1024, bpli, olpf, ovbplf)
	Expect(sys).NotTo(BeNil())
	return &portSystem{System: sys, bpli: bpli, olpf: olpf, ovbplf: ovbplf}
}

// Close releases the PhysicsSystem. Idempotent.
func (p *portSystem) Close() {
	if p.System != nil {
		Expect(p.System.Close()).To(Succeed())
		p.System = nil
	}
	// bpli/olpf/ovbplf are GC-managed; no explicit Close.
}

// createSphereBody mirrors the CreateSphereBody helper in test_collision.cpp:
// build a sphere shape, body settings, create the body, add it to the world,
// return its ID. The shape and BodyCreationSettings are GC-managed.
func createSphereBody(
	bi *jolt.BodyInterface,
	x, y, z float32,
	radius float32,
	motionType jolt.MotionType,
	layer jolt.ObjectLayer,
) jolt.BodyID {
	shape := jolt.NewSphereShape(radius)
	settings := jolt.NewBodyCreationSettingsWithShape(
		&shape.Shape,
		jolt.RVec3{X: jolt.Real(x), Y: jolt.Real(y), Z: jolt.Real(z)},
		motionType,
		layer,
	)
	return bi.CreateAndAddBody(settings, jolt.ActivationActivate)
}
