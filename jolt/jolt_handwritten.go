package jolt

import (
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
