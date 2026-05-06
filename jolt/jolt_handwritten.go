package jolt

import (
	"runtime"
	"unsafe"

	"github.com/gataengine/jolt-go/internal/cjolt"
)

// jolt_handwritten.go is the home for carve-outs that the generator can't
// or shouldn't produce.

// ---------------------------------------------------------------------------
// Global lifecycle
// ---------------------------------------------------------------------------

// Init initialises the Jolt Physics library. Must be called once before any
// other jolt function. Returns false if initialisation fails.
func Init() bool { return cjolt.JPH_Init() }

// Shutdown tears down the Jolt Physics library. Call once when you are done.
func Shutdown() { cjolt.JPH_Shutdown() }

// ---------------------------------------------------------------------------
// Layer-filter Mask-style helpers
// ---------------------------------------------------------------------------

// NewBroadPhaseLayerInterfaceMask creates a broad-phase layer interface that
// uses a bit-mask scheme. numBroadPhaseLayers is typically 2 (static /
// dynamic).
func NewBroadPhaseLayerInterfaceMask(numBroadPhaseLayers uint32) *BroadPhaseLayerInterface {
	h := cjolt.JPH_BroadPhaseLayerInterfaceMask_Create(numBroadPhaseLayers)
	return &BroadPhaseLayerInterface{h: h}
}

// NewObjectLayerPairFilterMask creates an object-layer pair filter that uses
// a bit-mask scheme.
func NewObjectLayerPairFilterMask() *ObjectLayerPairFilter {
	h := cjolt.JPH_ObjectLayerPairFilterMask_Create()
	return &ObjectLayerPairFilter{h: h}
}

// NewObjectVsBroadPhaseLayerFilterMask creates an object-vs-broad-phase-layer
// filter that uses a bit-mask scheme. bpli must be the same
// BroadPhaseLayerInterface used when constructing the PhysicsSystem.
func NewObjectVsBroadPhaseLayerFilterMask(bpli *BroadPhaseLayerInterface) *ObjectVsBroadPhaseLayerFilter {
	h := cjolt.JPH_ObjectVsBroadPhaseLayerFilterMask_Create(bpli.h)
	return &ObjectVsBroadPhaseLayerFilter{h: h}
}

// ---------------------------------------------------------------------------
// PhysicsSystemSettings — Go-side layout-compatible struct
//
// The generator emits "type PhysicsSystemSettings = cjolt.C_JPH_PhysicsSystemSettings"
// (a type alias for the C struct), but the CGo-exported C struct fields are
// unexported and cannot be set from Go code. This named Go struct mirrors the
// C layout exactly so it can be passed via unsafe.Pointer to NewPhysicsSystem.
//
// C layout (64-bit):
//   uint32  maxBodies               offset  0
//   uint32  numBodyMutexes          offset  4
//   uint32  maxBodyPairs            offset  8
//   uint32  maxContactConstraints   offset 12
//   uint32  _padding                offset 16
//   [4-byte implicit pad on 64-bit] offset 20  <- pointer alignment
//   pointer broadPhaseLayerInterface   offset 24
//   pointer objectLayerPairFilter      offset 32
//   pointer objectVsBroadPhaseLayerFilter offset 40
//   total 48 bytes
// ---------------------------------------------------------------------------

// goPhysicsSystemSettings is a layout-compatible mirror of
// JPH_PhysicsSystemSettings used internally so we can populate the C struct
// without cgo field access restrictions.
//
// We embed the scalar fields and store the pointer fields as unsafe.Pointer
// so that Go's GC does not move the objects out from under us during the
// Create call.
type goPhysicsSystemSettings struct {
	maxBodies             uint32
	numBodyMutexes        uint32
	maxBodyPairs          uint32
	maxContactConstraints uint32
	_padding              uint32
	_alignpad             uint32 // 4 bytes to reach offset 24 on 64-bit
	broadPhaseLayerInterface       unsafe.Pointer
	objectLayerPairFilter          unsafe.Pointer
	objectVsBroadPhaseLayerFilter  unsafe.Pointer
}

// ---------------------------------------------------------------------------
// JobSystem — thread-pool variant
//
// The generator emits the JobSystem type but no constructor because the C
// function takes a settings struct by pointer and returns the type via a
// shared handle.  We add the missing constructor and Close here.
// ---------------------------------------------------------------------------

// goJobSystemThreadPoolConfig is a layout-compatible mirror of the CGo
// JobSystemThreadPoolConfig struct whose fields are unexported from Go.
// C layout: uint32 maxJobs, uint32 maxBarriers, int32 numThreads
type goJobSystemThreadPoolConfig struct {
	maxJobs    uint32
	maxBarriers uint32
	numThreads  int32
}

// NewJobSystemThreadPool creates a Jolt thread-pool job system.
//   - maxJobs: maximum number of jobs in the queue (e.g. 2048)
//   - maxBarriers: maximum number of barrier objects (e.g. 8)
//   - numThreads: number of worker threads; -1 lets Jolt choose (= CPU count - 1)
func NewJobSystemThreadPool(maxJobs, maxBarriers uint32, numThreads int32) *JobSystem {
	cfg := goJobSystemThreadPoolConfig{
		maxJobs:    maxJobs,
		maxBarriers: maxBarriers,
		numThreads:  numThreads,
	}
	js := &JobSystem{h: cjolt.JPH_JobSystemThreadPool_Create((*cjolt.C_JobSystemThreadPoolConfig)(unsafe.Pointer(&cfg)))}
	runtime.SetFinalizer(js, (*JobSystem).finalize)
	return js
}

// Close destroys the underlying job system and releases its resources.
func (js *JobSystem) Close() error {
	if js.h != nil {
		cjolt.JPH_JobSystem_Destroy(js.h)
		js.h = nil
		runtime.SetFinalizer(js, nil)
	}
	return nil
}

func (js *JobSystem) finalize() { _ = js.Close() }

// ---------------------------------------------------------------------------
// Shape-settings constructors missing from the generator
// ---------------------------------------------------------------------------

// NewTaperedCapsuleShapeSettings creates a TaperedCapsuleShapeSettings.
// halfHeight is the half-height of the central cylinder portion;
// topRadius and bottomRadius are the sphere radii at each end.
func NewTaperedCapsuleShapeSettings(halfHeight, topRadius, bottomRadius float32) *TaperedCapsuleShapeSettings {
	h := cjolt.JPH_TaperedCapsuleShapeSettings_Create(halfHeight, topRadius, bottomRadius)
	return &TaperedCapsuleShapeSettings{h: h}
}

// NewTaperedCylinderShapeSettings creates a TaperedCylinderShapeSettings.
// halfHeight is half the cylinder height; topRadius and bottomRadius are the
// radii at the two flat ends; convexRadius is the edge-rounding radius.
func NewTaperedCylinderShapeSettings(halfHeight, topRadius, bottomRadius, convexRadius float32) *TaperedCylinderShapeSettings {
	h := cjolt.JPH_TaperedCylinderShapeSettings_Create(halfHeight, topRadius, bottomRadius, convexRadius, nil)
	return &TaperedCylinderShapeSettings{h: h}
}

// NewEmptyShapeSettings creates an EmptyShapeSettings with the given
// center-of-mass offset.
func NewEmptyShapeSettings(centerOfMass Vec3) *EmptyShapeSettings {
	h := cjolt.JPH_EmptyShapeSettings_Create((*cjolt.C_JPH_Vec3)(unsafe.Pointer(&centerOfMass)))
	return &EmptyShapeSettings{h: h}
}

// ---------------------------------------------------------------------------
// PlaneShape — Go-native layout-compatible Plane helper
//
// cjolt.C_JPH_Plane (= C.JPH_Plane) has unexported CGo fields; we provide
// a Go-native mirror so callers do not need unsafe outside this package.
// ---------------------------------------------------------------------------

// GoPlane is a layout-compatible mirror of JPH_Plane.
// C layout: JPH_Vec3 normal (12 bytes) + float distance (4 bytes) = 16 bytes.
type GoPlane struct {
	Normal   Vec3
	Distance float32
}

// NewPlaneShapeFromGoPlane creates a PlaneShape from a GoPlane and an optional
// half-extent.  Pass nil for material to use the default material.
func NewPlaneShapeFromGoPlane(p GoPlane, material *PhysicsMaterial, halfExtent float32) *PlaneShape {
	var mh *cjolt.C_JPH_PhysicsMaterial
	if material != nil {
		mh = material.h
	}
	raw := cjolt.JPH_PlaneShape_Create((*cjolt.C_JPH_Plane)(unsafe.Pointer(&p)), mh, halfExtent)
	h := (*cjolt.C_JPH_Shape)(unsafe.Pointer(raw))
	s := &PlaneShape{ConvexShape: ConvexShape{Shape: Shape{h: h}}}
	runtime.SetFinalizer(s, (*PlaneShape).finalize)
	return s
}

// NewPhysicsSystemMask is a convenience constructor that wires up the Mask
// style layer filters and creates a PhysicsSystem in one call. It calls
// Init() / Shutdown() internally only when the library has not yet been
// initialised; callers that manage their own lifecycle should call Init()
// first and use NewPhysicsSystem directly.
//
// Parameters use the joltc defaults:
//   maxBodies = 10240, maxBodyPairs = 65536, maxContactConstraints = 10240
func NewPhysicsSystemWithMaskFilters(
	maxBodies, maxBodyPairs, maxContactConstraints uint32,
	bpli *BroadPhaseLayerInterface,
	olpf *ObjectLayerPairFilter,
	ovbplf *ObjectVsBroadPhaseLayerFilter,
) *PhysicsSystem {
	s := goPhysicsSystemSettings{
		maxBodies:                    maxBodies,
		maxBodyPairs:                 maxBodyPairs,
		maxContactConstraints:        maxContactConstraints,
		broadPhaseLayerInterface:     unsafe.Pointer(bpli.h),
		objectLayerPairFilter:        unsafe.Pointer(olpf.h),
		objectVsBroadPhaseLayerFilter: unsafe.Pointer(ovbplf.h),
	}
	return NewPhysicsSystem(unsafe.Pointer(&s))
}
