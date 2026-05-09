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

// NewSphereShapeSettings creates a SphereShapeSettings with the given radius.
func NewSphereShapeSettings(radius float32) *SphereShapeSettings {
	h := cjolt.JPH_SphereShapeSettings_Create(radius)
	return &SphereShapeSettings{h: h}
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

// NewBoxShapeSettings creates a BoxShapeSettings.
func NewBoxShapeSettings(halfExtent Vec3, convexRadius float32) *BoxShapeSettings {
	h := cjolt.JPH_BoxShapeSettings_Create((*cjolt.C_JPH_Vec3)(unsafe.Pointer(&halfExtent)), convexRadius)
	return &BoxShapeSettings{h: h}
}

// GetHalfExtent returns the half-extent of the BoxShape.
//
// TODO(generator): emit Get* variants for out-pointer accessors so this
// kind of carve-out becomes unnecessary across the API.
func (bs *BoxShape) GetHalfExtent() Vec3 {
	var out Vec3
	cjolt.JPH_BoxShape_GetHalfExtent((*cjolt.C_JPH_BoxShape)(unsafe.Pointer(bs.h)), (*cjolt.C_JPH_Vec3)(unsafe.Pointer(&out)))
	runtime.KeepAlive(bs)
	return out
}

// NewCapsuleShapeSettings creates a CapsuleShapeSettings.
func NewCapsuleShapeSettings(halfHeightOfCylinder, radius float32) *CapsuleShapeSettings {
	h := cjolt.JPH_CapsuleShapeSettings_Create(halfHeightOfCylinder, radius)
	return &CapsuleShapeSettings{h: h}
}

// NewCylinderShapeSettings creates a CylinderShapeSettings.
func NewCylinderShapeSettings(halfHeight, radius, convexRadius float32) *CylinderShapeSettings {
	h := cjolt.JPH_CylinderShapeSettings_Create(halfHeight, radius, convexRadius)
	return &CylinderShapeSettings{h: h}
}

// GetVertex1 returns vertex 1 of the TriangleShape.
func (ts *TriangleShape) GetVertex1() Vec3 {
	var out Vec3
	cjolt.JPH_TriangleShape_GetVertex1((*cjolt.C_JPH_TriangleShape)(unsafe.Pointer(ts.h)), (*cjolt.C_JPH_Vec3)(unsafe.Pointer(&out)))
	runtime.KeepAlive(ts)
	return out
}

// GetVertex2 returns vertex 2 of the TriangleShape.
func (ts *TriangleShape) GetVertex2() Vec3 {
	var out Vec3
	cjolt.JPH_TriangleShape_GetVertex2((*cjolt.C_JPH_TriangleShape)(unsafe.Pointer(ts.h)), (*cjolt.C_JPH_Vec3)(unsafe.Pointer(&out)))
	runtime.KeepAlive(ts)
	return out
}

// GetVertex3 returns vertex 3 of the TriangleShape.
func (ts *TriangleShape) GetVertex3() Vec3 {
	var out Vec3
	cjolt.JPH_TriangleShape_GetVertex3((*cjolt.C_JPH_TriangleShape)(unsafe.Pointer(ts.h)), (*cjolt.C_JPH_Vec3)(unsafe.Pointer(&out)))
	runtime.KeepAlive(ts)
	return out
}

// NewConvexHullShapeSettings creates a ConvexHullShapeSettings from a slice
// of vertices. The slice must have at least 4 points to form a valid hull.
func NewConvexHullShapeSettings(points []Vec3, maxConvexRadius float32) *ConvexHullShapeSettings {
	if len(points) == 0 {
		return nil
	}
	h := cjolt.JPH_ConvexHullShapeSettings_Create(
		(*cjolt.C_JPH_Vec3)(unsafe.Pointer(&points[0])),
		uint32(len(points)),
		maxConvexRadius,
	)
	return &ConvexHullShapeSettings{h: h}
}

// GetPoint returns the i'th hull vertex.
func (chs *ConvexHullShape) GetPoint(i uint32) Vec3 {
	var out Vec3
	cjolt.JPH_ConvexHullShape_GetPoint((*cjolt.C_JPH_ConvexHullShape)(unsafe.Pointer(chs.h)), i, (*cjolt.C_JPH_Vec3)(unsafe.Pointer(&out)))
	runtime.KeepAlive(chs)
	return out
}

// GetPlane returns the plane equation of the PlaneShape.
func (ps *PlaneShape) GetPlane() GoPlane {
	var out GoPlane
	cjolt.JPH_PlaneShape_GetPlane((*cjolt.C_JPH_PlaneShape)(unsafe.Pointer(ps.h)), (*cjolt.C_JPH_Plane)(unsafe.Pointer(&out)))
	runtime.KeepAlive(ps)
	return out
}

// ---------------------------------------------------------------------------
// Shape base-class carve-outs
//
// The generated Shape methods for GetLocalBounds and GetCenterOfMass use
// out-pointer parameters with CGo types whose fields are unexported from Go.
// These carve-outs return Go-native types instead.
// ---------------------------------------------------------------------------

// GoAABox is a layout-compatible mirror of JPH_AABox so test/user code
// can read Min/Max without unexported-field constraints.
// C layout: JPH_Vec3 min; JPH_Vec3 max — 24 bytes.
type GoAABox struct {
	Min Vec3
	Max Vec3
}

// ---------------------------------------------------------------------------
// ShapeSettings UserData promotion carve-outs
//
// The generator emits SetUserData/UserData on *ShapeSettings only.
// Concrete settings types hold their own h pointer, so Go embedding is not
// used and method promotion does not occur automatically.  The helpers below
// forward to the base-class C function via an unsafe upcast — valid because
// JPH_SphereShapeSettings* IS-A JPH_ShapeSettings* in C.
// ---------------------------------------------------------------------------

// SetUserData sets the opaque user-data value on the SphereShapeSettings.
func (sss *SphereShapeSettings) SetUserData(u uint64) {
	cjolt.JPH_ShapeSettings_SetUserData((*cjolt.C_JPH_ShapeSettings)(unsafe.Pointer(sss.h)), u)
	runtime.KeepAlive(sss)
}

// UserData returns the opaque user-data value from the SphereShapeSettings.
func (sss *SphereShapeSettings) UserData() uint64 {
	result := cjolt.JPH_ShapeSettings_GetUserData((*cjolt.C_JPH_ShapeSettings)(unsafe.Pointer(sss.h)))
	runtime.KeepAlive(sss)
	return result
}

// GetLocalBounds returns the local-space AABB of the shape.
func (s *Shape) GetLocalBounds() GoAABox {
	var out GoAABox
	cjolt.JPH_Shape_GetLocalBounds(s.h, (*cjolt.C_JPH_AABox)(unsafe.Pointer(&out)))
	runtime.KeepAlive(s)
	return out
}

// GetCenterOfMass returns the center-of-mass position in local space.
func (s *Shape) GetCenterOfMass() Vec3 {
	var out Vec3
	cjolt.JPH_Shape_GetCenterOfMass(s.h, (*cjolt.C_JPH_Vec3)(unsafe.Pointer(&out)))
	runtime.KeepAlive(s)
	return out
}

// ---------------------------------------------------------------------------
// Compound shape settings constructors and carve-outs
// ---------------------------------------------------------------------------

// NewStaticCompoundShapeSettings creates an empty StaticCompoundShapeSettings.
// Add sub-shapes via the embedded CompoundShapeSettings.AddShape2 method.
func NewStaticCompoundShapeSettings() *StaticCompoundShapeSettings {
	h := cjolt.JPH_StaticCompoundShapeSettings_Create()
	return &StaticCompoundShapeSettings{h: h}
}

// NewMutableCompoundShapeSettings creates an empty MutableCompoundShapeSettings.
func NewMutableCompoundShapeSettings() *MutableCompoundShapeSettings {
	h := cjolt.JPH_MutableCompoundShapeSettings_Create()
	return &MutableCompoundShapeSettings{h: h}
}

// AddShape2 adds a sub-shape to a StaticCompoundShapeSettings.
//
// The base CompoundShapeSettings.AddShape2 method is not promoted because
// StaticCompoundShapeSettings does not embed CompoundShapeSettings (the
// generator emits each Settings type as an independent struct). Carve out
// an upcast wrapper here. JPH_StaticCompoundShapeSettings* IS-A
// JPH_CompoundShapeSettings* in C.
func (s *StaticCompoundShapeSettings) AddShape2(position Vec3, rotation Quat, shape *Shape, userData uint32) {
	cjolt.JPH_CompoundShapeSettings_AddShape2(
		(*cjolt.C_JPH_CompoundShapeSettings)(unsafe.Pointer(s.h)),
		(*cjolt.C_JPH_Vec3)(unsafe.Pointer(&position)),
		(*cjolt.C_JPH_Quat)(unsafe.Pointer(&rotation)),
		shape.h,
		userData,
	)
	runtime.KeepAlive(s)
	runtime.KeepAlive(shape)
}

// ---------------------------------------------------------------------------
// Decorated shape carve-outs — GetPosition/GetRotation/GetScale/GetOffset
//
// The generator emits Position(Vec3), Rotation(Quat), Scale(Vec3), Offset(Vec3)
// as out-pointer methods that pass by value, so the C write is lost.  These
// carve-outs return the value directly.
// ---------------------------------------------------------------------------

// GetPosition returns the offset position of the RotatedTranslatedShape.
func (rts *RotatedTranslatedShape) GetPosition() Vec3 {
	var out Vec3
	cjolt.JPH_RotatedTranslatedShape_GetPosition((*cjolt.C_JPH_RotatedTranslatedShape)(unsafe.Pointer(rts.h)), (*cjolt.C_JPH_Vec3)(unsafe.Pointer(&out)))
	runtime.KeepAlive(rts)
	return out
}

// GetRotation returns the offset rotation of the RotatedTranslatedShape.
func (rts *RotatedTranslatedShape) GetRotation() Quat {
	var out Quat
	cjolt.JPH_RotatedTranslatedShape_GetRotation((*cjolt.C_JPH_RotatedTranslatedShape)(unsafe.Pointer(rts.h)), (*cjolt.C_JPH_Quat)(unsafe.Pointer(&out)))
	runtime.KeepAlive(rts)
	return out
}

// GetScale returns the scale of the ScaledShape.
func (ss *ScaledShape) GetScale() Vec3 {
	var out Vec3
	cjolt.JPH_ScaledShape_GetScale((*cjolt.C_JPH_ScaledShape)(unsafe.Pointer(ss.h)), (*cjolt.C_JPH_Vec3)(unsafe.Pointer(&out)))
	runtime.KeepAlive(ss)
	return out
}

// GetOffset returns the COM offset of the OffsetCenterOfMassShape.
func (ocoms *OffsetCenterOfMassShape) GetOffset() Vec3 {
	var out Vec3
	cjolt.JPH_OffsetCenterOfMassShape_GetOffset((*cjolt.C_JPH_OffsetCenterOfMassShape)(unsafe.Pointer(ocoms.h)), (*cjolt.C_JPH_Vec3)(unsafe.Pointer(&out)))
	runtime.KeepAlive(ocoms)
	return out
}

// ---------------------------------------------------------------------------
// Shape ray-cast and point-collision carve-outs
//
// The generated Shape.CastRay passes the result via unsafe.Pointer.
// GoRayCastResult is a layout-compatible Go-native mirror so callers can
// read fields without unexported CGo field constraints.
// C layout: JPH_BodyID bodyID (uint32) + float fraction + JPH_SubShapeID subShapeID2 (uint32) = 12 bytes.
//
// The generated Shape.CollidePoint does not guard against a nil shapeFilter,
// but the underlying C function accepts NULL. The carve-out below handles nil.
// ---------------------------------------------------------------------------

// CollidePoint returns true if point p is inside the shape.
// shapeFilter may be nil (no filter).
func (s *Shape) CollidePoint(p Vec3, shapeFilter *ShapeFilter) bool {
	var sfh *cjolt.C_JPH_ShapeFilter
	if shapeFilter != nil {
		sfh = shapeFilter.h
	}
	result := cjolt.JPH_Shape_CollidePoint(s.h, (*cjolt.C_JPH_Vec3)(unsafe.Pointer(&p)), sfh)
	runtime.KeepAlive(s)
	runtime.KeepAlive(shapeFilter)
	return bool(result)
}

// GoRayCastResult is a layout-compatible mirror of JPH_RayCastResult so
// callers can read fields without unexported-name CGo constraints.
// C layout: JPH_BodyID bodyID (uint32) + float fraction + JPH_SubShapeID subShapeID2 (uint32) = 12 bytes.
type GoRayCastResult struct {
	BodyID     BodyID
	Fraction   float32
	SubShapeID SubShapeID
}

// CastRayResult is a typed wrapper for Shape.CastRay that returns the hit
// info via a Go-native struct rather than an unsafe.Pointer.
//
// Returns (hit, true) on a hit and (zero, false) on a miss.
func (s *Shape) CastRayResult(origin, direction Vec3) (GoRayCastResult, bool) {
	var out GoRayCastResult
	hit := cjolt.JPH_Shape_CastRay(s.h, (*cjolt.C_JPH_Vec3)(unsafe.Pointer(&origin)), (*cjolt.C_JPH_Vec3)(unsafe.Pointer(&direction)), (*cjolt.C_JPH_RayCastResult)(unsafe.Pointer(&out)))
	runtime.KeepAlive(s)
	if !hit {
		return GoRayCastResult{}, false
	}
	return out, true
}

// MakeScaleValidResult clamps the given scale to a valid value for this shape
// and returns the result.
func (s *Shape) MakeScaleValidResult(scale Vec3) Vec3 {
	var out Vec3
	cjolt.JPH_Shape_MakeScaleValid(s.h, (*cjolt.C_JPH_Vec3)(unsafe.Pointer(&scale)), (*cjolt.C_JPH_Vec3)(unsafe.Pointer(&out)))
	runtime.KeepAlive(s)
	return out
}
