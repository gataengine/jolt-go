package jolt

import (
	"runtime"
	"unsafe"

	"github.com/gataengine/jolt-go/internal/cjolt"
)

// jolt_handwritten.go is the home for carve-outs that the generator can't
// or shouldn't produce.

// ---------------------------------------------------------------------------
// Enum constants — MotionType, Activation, ObjectLayer (Mask scheme)
//
// These are C enum values that cannot be expressed as Go typed constants
// directly (CGo enum types are not iota-friendly).  We declare them as typed
// constants using their numeric values from joltc.h.
// ---------------------------------------------------------------------------

// MotionType values — mirror JPH_MotionType in joltc.h.
const (
	MotionTypeStatic    MotionType = 0 // JPH_MotionType_Static
	MotionTypeKinematic MotionType = 1 // JPH_MotionType_Kinematic
	MotionTypeDynamic   MotionType = 2 // JPH_MotionType_Dynamic
)

// Activation values — mirror JPH_Activation in joltc.h.
const (
	ActivationActivate     Activation = 0 // JPH_Activation_Activate
	ActivationDontActivate Activation = 1 // JPH_Activation_DontActivate
)

// ShapeType values — mirror JPH_ShapeType in joltc.h.
const (
	ShapeTypeConvex      ShapeType = 0 // JPH_ShapeType_Convex
	ShapeTypeCompound    ShapeType = 1 // JPH_ShapeType_Compound
	ShapeTypeDecorated   ShapeType = 2 // JPH_ShapeType_Decorated
	ShapeTypeMesh        ShapeType = 3 // JPH_ShapeType_Mesh
	ShapeTypeHeightField ShapeType = 4 // JPH_ShapeType_HeightField
	ShapeTypeSoftBody    ShapeType = 5 // JPH_ShapeType_SoftBody
)

// ShapeSubType values — mirror JPH_ShapeSubType in joltc.h.
const (
	ShapeSubTypeSphere             ShapeSubType = 0
	ShapeSubTypeBox                ShapeSubType = 1
	ShapeSubTypeTriangle           ShapeSubType = 2
	ShapeSubTypeCapsule            ShapeSubType = 3
	ShapeSubTypeTaperedCapsule     ShapeSubType = 4
	ShapeSubTypeCylinder           ShapeSubType = 5
	ShapeSubTypeConvexHull         ShapeSubType = 6
	ShapeSubTypeStaticCompound     ShapeSubType = 7
	ShapeSubTypeMutableCompound    ShapeSubType = 8
	ShapeSubTypeRotatedTranslated  ShapeSubType = 9
	ShapeSubTypeScaled             ShapeSubType = 10
	ShapeSubTypeOffsetCenterOfMass ShapeSubType = 11
	ShapeSubTypeMesh               ShapeSubType = 12
	ShapeSubTypeHeightField        ShapeSubType = 13
	ShapeSubTypeSoftBody           ShapeSubType = 14
)

// BodyType values — mirror JPH_BodyType in joltc.h.
const (
	BodyTypeRigid BodyType = 0
	BodyTypeSoft  BodyType = 1
)

// PhysicsUpdateError values — mirror JPH_PhysicsUpdateError in joltc.h.
const (
	PhysicsUpdateErrorNone                   PhysicsUpdateError = 0
	PhysicsUpdateErrorManifoldCacheFull      PhysicsUpdateError = 1 << 0
	PhysicsUpdateErrorBodyPairCacheFull      PhysicsUpdateError = 1 << 1
	PhysicsUpdateErrorContactConstraintsFull PhysicsUpdateError = 1 << 2
)

// DefaultConvexRadius mirrors JPH_DEFAULT_CONVEX_RADIUS from joltc.h.
const DefaultConvexRadius float32 = 0.05

// ObjectLayerMaskMake encodes group and mask bits into an ObjectLayer value
// for use with the ObjectLayerPairFilterMask / BroadPhaseLayerInterfaceMask
// scheme.
//
// joltc is built with JPH_OBJECT_LAYER_BITS=32 (see third_party/joltc/CMakeLists.txt),
// so cNumBits = 32/2 = 16 and cMask = 0xFFFF.
// Encoding: (group & 0xFFFF) | (mask << 16).
//
// Two objects collide when:
//
//	(group(A) & mask(B)) != 0 && (group(B) & mask(A)) != 0
func ObjectLayerMaskMake(group, mask uint32) ObjectLayer {
	const numBits = 16 // JPH_OBJECT_LAYER_BITS/2 = 32/2 = 16
	return ObjectLayer((group & 0xFFFF) | (mask << numBits))
}

// ObjectLayer constants for a standard two-layer (NON_MOVING / MOVING) Mask
// scheme using the 32-bit ObjectLayer encoding (16-bit group/mask split).
//
//	NonMoving: group=0x0001, mask=0x0002  →  0x0002_0001
//	Moving:    group=0x0002, mask=0x0003  →  0x0003_0002
//
// Use SetupDefaultBroadPhaseLayers to configure the BroadPhaseLayerInterface.
var (
	ObjectLayerNonMoving = ObjectLayerMaskMake(0x0001, 0x0002) // static / floor
	ObjectLayerMoving    = ObjectLayerMaskMake(0x0002, 0x0003) // dynamic / actors
)

// SetupDefaultBroadPhaseLayers configures a 2-layer BroadPhaseLayerInterface
// created with NewBroadPhaseLayerInterfaceMask(2) for the standard
// NON_MOVING (BP layer 0) / MOVING (BP layer 1) scheme used by the
// ObjectLayerNonMoving / ObjectLayerMoving constants above.
//
// BP layer 0 accepts objects whose group bit 0 is set (NON_MOVING).
// BP layer 1 accepts objects whose group bit 1 is set (MOVING).
func SetupDefaultBroadPhaseLayers(bpli *BroadPhaseLayerInterface) {
	// BP layer 0 accepts objects whose group bit 0 is set (= NON_MOVING group)
	bpli.ConfigureLayer(0, 0x0001, 0)
	// BP layer 1 accepts objects whose group bit 1 is set (= MOVING group)
	bpli.ConfigureLayer(1, 0x0002, 0)
}

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
	maxBodies                     uint32
	numBodyMutexes                uint32
	maxBodyPairs                  uint32
	maxContactConstraints         uint32
	_padding                      uint32
	_alignpad                     uint32 // 4 bytes to reach offset 24 on 64-bit
	broadPhaseLayerInterface      unsafe.Pointer
	objectLayerPairFilter         unsafe.Pointer
	objectVsBroadPhaseLayerFilter unsafe.Pointer
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
	maxJobs     uint32
	maxBarriers uint32
	numThreads  int32
}

// NewJobSystemThreadPool creates a Jolt thread-pool job system.
//   - maxJobs: maximum number of jobs in the queue (e.g. 2048)
//   - maxBarriers: maximum number of barrier objects (e.g. 8)
//   - numThreads: number of worker threads; -1 lets Jolt choose (= CPU count - 1)
func NewJobSystemThreadPool(maxJobs, maxBarriers uint32, numThreads int32) *JobSystem {
	cfg := goJobSystemThreadPoolConfig{
		maxJobs:     maxJobs,
		maxBarriers: maxBarriers,
		numThreads:  numThreads,
	}
	js := &JobSystem{h: cjolt.JPH_JobSystemThreadPool_Create((*cjolt.C_JobSystemThreadPoolConfig)(unsafe.Pointer(&cfg)))}
	runtime.SetFinalizer(js, (*JobSystem).finalize)
	return js
}

// NewJobSystemThreadPoolDefault creates a Jolt thread-pool job system
// using the Jolt-recommended defaults (nullptr config in C: maxJobs =
// JPH_MAX_PHYSICS_JOBS, maxBarriers = JPH_MAX_PHYSICS_BARRIERS,
// numThreads = -1 / auto).
func NewJobSystemThreadPoolDefault() *JobSystem {
	js := &JobSystem{h: cjolt.JPH_JobSystemThreadPool_Create(nil)}
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
//
//	maxBodies = 10240, maxBodyPairs = 65536, maxContactConstraints = 10240
func NewPhysicsSystemWithMaskFilters(
	maxBodies, maxBodyPairs, maxContactConstraints uint32,
	bpli *BroadPhaseLayerInterface,
	olpf *ObjectLayerPairFilter,
	ovbplf *ObjectVsBroadPhaseLayerFilter,
) *PhysicsSystem {
	s := goPhysicsSystemSettings{
		maxBodies:                     maxBodies,
		maxBodyPairs:                  maxBodyPairs,
		maxContactConstraints:         maxContactConstraints,
		broadPhaseLayerInterface:      unsafe.Pointer(bpli.h),
		objectLayerPairFilter:         unsafe.Pointer(olpf.h),
		objectVsBroadPhaseLayerFilter: unsafe.Pointer(ovbplf.h),
	}
	return NewPhysicsSystem(unsafe.Pointer(&s))
}

// ---------------------------------------------------------------------------
// BodyCreationSettings — convenience constructor with shape
//
// The generator only emits NewBodyCreationSettings() (no shape).
// JPH_BodyCreationSettings_Create3 in joltc takes a Shape directly and sets
// position, rotation, motionType and objectLayer in one call.
// ---------------------------------------------------------------------------

// NewBodyCreationSettingsWithShape creates a BodyCreationSettings pre-wired
// with the given shape, position, motion type, and object layer.
// rotation is set to the identity quaternion automatically.
func NewBodyCreationSettingsWithShape(
	shape *Shape,
	position RVec3,
	motionType MotionType,
	objectLayer ObjectLayer,
) *BodyCreationSettings {
	identity := Quat{X: 0, Y: 0, Z: 0, W: 1}
	h := cjolt.JPH_BodyCreationSettings_Create3(
		shape.h,
		(*cjolt.C_JPH_RVec3)(unsafe.Pointer(&position)),
		(*cjolt.C_JPH_Quat)(unsafe.Pointer(&identity)),
		cjolt.C_JPH_MotionType(motionType),
		cjolt.C_JPH_ObjectLayer(objectLayer),
	)
	p := &BodyCreationSettings{h: h}
	runtime.SetFinalizer(p, (*BodyCreationSettings).finalize)
	return p
}

// ---------------------------------------------------------------------------
// BodyInterface — GetPosition returns RVec3 by value
//
// The generated Position(b BodyID, r RVec3) passes r by value so the C write
// is lost.  GetPosition returns the value directly.
// ---------------------------------------------------------------------------

// GetPosition returns the world-space position of the body with the given ID.
func (bi *BodyInterface) GetPosition(b BodyID) RVec3 {
	var out RVec3
	cjolt.JPH_BodyInterface_GetPosition(bi.h, cjolt.C_JPH_BodyID(b), (*cjolt.C_JPH_RVec3)(unsafe.Pointer(&out)))
	runtime.KeepAlive(bi)
	return out
}

// ---------------------------------------------------------------------------
// BroadPhaseLayerInterface — ConfigureLayer for Mask scheme
// ---------------------------------------------------------------------------

// ConfigureLayer configures a broad-phase layer for the Mask scheme.
// broadPhaseLayer is the layer index (0..numBroadPhaseLayers-1).
// groupsToInclude and groupsToExclude are bit-masks of object-layer groups.
func (bpli *BroadPhaseLayerInterface) ConfigureLayer(broadPhaseLayer uint8, groupsToInclude, groupsToExclude uint32) {
	cjolt.JPH_BroadPhaseLayerInterfaceMask_ConfigureLayer(bpli.h, cjolt.C_JPH_BroadPhaseLayer(broadPhaseLayer), groupsToInclude, groupsToExclude)
	runtime.KeepAlive(bpli)
}
