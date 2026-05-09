# Test-Coverage Port Implementation Plan (Phases A–C)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Port the MVP-bound subset of `third_party/joltc/tests/` into Ginkgo specs, surfacing and fixing the systemic out-pointer accessor bug as carve-outs. End state: ~63 new specs across Core, Shapes, and PhysicsSystem/BodyInterface, all passing.

**Architecture:** Tests live in `jolt/` alongside existing test files. One Go file per joltc source file plus a shared helpers file. Carve-outs for broken generated accessors land in `jolt/jolt_handwritten.go`, named `GetX()` to coexist with the broken generated `X()`. Missing Go-level `New*Settings` constructors land in `jolt_handwritten.go`. Layer setup uses the Mask scheme already wired up in our handwritten code (consistent with `integration_test.go`).

**Spec:** `docs/superpowers/specs/2026-05-07-test-coverage-port-design.md`.

**Tech Stack:** Go 1.25, Ginkgo v2 + Gomega, cgo via the existing `internal/cjolt` package.

---

## Conventions

**TDD cycle for each batch:**
1. Add any new carve-outs / `New*Settings` constructors needed by this batch (in `jolt_handwritten.go`).
2. Write the new test specs.
3. Run `go test ./jolt/...` (or focused via `-ginkgo.focus`) — expect PASS.
4. Commit.

If a carve-out is *not* needed (the generated accessor returns by value and works), skip step 1.

**Ginkgo style:**
- Each test file: `var _ = Describe("FooName (joltc test_foo.cpp)", func() { ... })`
- Per-fixture group: nested `Describe` with `BeforeEach` / `AfterEach`
- Per-test: `It("TestName", func() { ... })` mirroring the joltc `TEST_F` name verbatim
- Float comparisons: `Expect(x).To(BeNumerically("~", target, 1e-5))` for `EXPECT_FLOAT_EQ`; larger eps where simulation steady-state requires it
- Pointer-equality: `Expect(p).NotTo(BeNil())` for `ASSERT_NE(p, nullptr)`

**Carve-out pattern:** Each new `Get*` carve-out follows the existing `BodyInterface.GetPosition` template:

```go
// GetX returns ... (carve-out for the broken generated X(r T) accessor).
func (recv *Type) GetX() T {
    var out T
    cjolt.JPH_TypeName_GetX(recv.h, (*cjolt.C_TypeName_Of_T)(unsafe.Pointer(&out)))
    runtime.KeepAlive(recv)
    return out
}
```

**Commits:** Conventional Commits — `test:`, `feat:`, `fix:`. Commit at the end of every task.

**Body cleanup in tests:** tests that create bodies must call `bi.RemoveAndDestroyBody(id)` before returning, mirroring the joltc cleanup. Use `defer` immediately after `CreateAndAddBody`. (Bodies are *not* GC-managed — they live in the PhysicsSystem.)

---

## File map

| Path | Action | Purpose |
|---|---|---|
| `jolt/jolt_handwritten.go` | Modify (append) | New `Get*` carve-outs, missing `New*Settings` ctors, enum constants |
| `jolt/port_helpers_test.go` | Create | Shared test fixtures (system setup, body creation) |
| `jolt/port_core_test.go` | Create | Phase A — ports `test_core.cpp` |
| `jolt/port_shapes_test.go` | Create | Phase B — ports `test_shapes.cpp` |
| `jolt/port_physics_system_test.go` | Create | Phase C — ports `test_physics_system.cpp` |

---

## Phase 0 — Setup

### Task 1: Enum constants

Add Go constants for the joltc enums the ports reference. The generator emits the *types* (`ShapeType`, `ShapeSubType`, `BodyType`, `PhysicsUpdateError`) but no Go-level constants for their values, so the port has no way to express `JPH_ShapeType_Convex` or `JPH_BodyType_Rigid`.

**Files:**
- Modify: `jolt/jolt_handwritten.go` (append)

- [ ] **Step 1: Add the constants block**

Append to `jolt/jolt_handwritten.go` (after the existing `Activation` block, before `ObjectLayerMaskMake`):

```go
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
	ShapeSubTypeSphere              ShapeSubType = 0
	ShapeSubTypeBox                 ShapeSubType = 1
	ShapeSubTypeTriangle            ShapeSubType = 2
	ShapeSubTypeCapsule             ShapeSubType = 3
	ShapeSubTypeTaperedCapsule      ShapeSubType = 4
	ShapeSubTypeCylinder            ShapeSubType = 5
	ShapeSubTypeConvexHull          ShapeSubType = 6
	ShapeSubTypeStaticCompound      ShapeSubType = 7
	ShapeSubTypeMutableCompound     ShapeSubType = 8
	ShapeSubTypeRotatedTranslated   ShapeSubType = 9
	ShapeSubTypeScaled              ShapeSubType = 10
	ShapeSubTypeOffsetCenterOfMass  ShapeSubType = 11
	ShapeSubTypeMesh                ShapeSubType = 12
	ShapeSubTypeHeightField         ShapeSubType = 13
	ShapeSubTypeSoftBody            ShapeSubType = 14
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
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./...`
Expected: no output, exit code 0.

- [ ] **Step 3: Commit**

```bash
git add jolt/jolt_handwritten.go
git commit -m "feat(jolt): add ShapeType/ShapeSubType/BodyType/PhysicsUpdateError constants"
```

---

### Task 2: Shared test helpers

Create the shared helpers used by every Phase B/C test file: a two-layer system constructor and a sphere-body constructor that mirror the C++ test fixtures. Phase A doesn't use them directly (its tests don't need a PhysicsSystem), but defining helpers up front means later tasks just reference them.

**Files:**
- Create: `jolt/port_helpers_test.go`

- [ ] **Step 1: Write the helper file**

```go
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
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./jolt/...`
Expected: no output, exit code 0. (No tests actually exercise the helpers yet, but the file must compile alongside the rest of the test package.)

- [ ] **Step 3: Verify existing tests still pass**

Run: `go test ./jolt/...`
Expected: PASS. The helpers should not affect existing tests.

- [ ] **Step 4: Commit**

```bash
git add jolt/port_helpers_test.go
git commit -m "test(jolt): add shared port-suite helpers (newTwoLayerSystem, createSphereBody)"
```

---

## Phase A — Core

### Task 3: Port test_core.cpp (4 specs)

**Files:**
- Create: `jolt/port_core_test.go`
- Modify: `jolt/jolt_handwritten.go` (add `NewJobSystemThreadPoolDefault`)

- [ ] **Step 1: Add the default-config JobSystem constructor**

Append to `jolt/jolt_handwritten.go` (right after `NewJobSystemThreadPool`):

```go
// NewJobSystemThreadPoolDefault creates a Jolt thread-pool job system
// using the Jolt-recommended defaults (nullptr config in C: maxJobs =
// JPH_MAX_PHYSICS_JOBS, maxBarriers = JPH_MAX_PHYSICS_BARRIERS,
// numThreads = -1 / auto).
func NewJobSystemThreadPoolDefault() *JobSystem {
	js := &JobSystem{h: cjolt.JPH_JobSystemThreadPool_Create(nil)}
	runtime.SetFinalizer(js, (*JobSystem).finalize)
	return js
}
```

- [ ] **Step 2: Write the port_core_test.go file**

```go
package jolt_test

// port_core_test.go ports third_party/joltc/tests/test_core.cpp.
//
// Deferred (require Go callback carve-outs):
//   JobSystemTest.SetTraceHandler
//   JobSystemTest.SetAssertFailureHandler

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gataengine/jolt-go/jolt"
)

var _ = Describe("CoreTest (joltc test_core.cpp)", func() {
	It("Init_Shutdown", func() {
		Expect(jolt.Init()).To(BeTrue())
		jolt.Shutdown()
	})

	It("Init_Multiple_Times", func() {
		Expect(jolt.Init()).To(BeTrue())
		jolt.Shutdown()

		Expect(jolt.Init()).To(BeTrue())
		jolt.Shutdown()
	})
})

var _ = Describe("JobSystemTest (joltc test_core.cpp)", func() {
	BeforeEach(func() {
		Expect(jolt.Init()).To(BeTrue())
	})
	AfterEach(func() {
		jolt.Shutdown()
	})

	It("ThreadPool_Create_Default", func() {
		js := jolt.NewJobSystemThreadPoolDefault()
		Expect(js).NotTo(BeNil())
		Expect(js.Close()).To(Succeed())
	})

	It("ThreadPool_Create_WithConfig", func() {
		// joltc passes JPH_MAX_PHYSICS_JOBS / JPH_MAX_PHYSICS_BARRIERS
		// (2048 / 8 in joltc.h); use the same here.
		js := jolt.NewJobSystemThreadPool(2048, 8, 2)
		Expect(js).NotTo(BeNil())
		Expect(js.Close()).To(Succeed())
	})
})
```

- [ ] **Step 3: Run the tests**

Run: `go test ./jolt/... -v -ginkgo.focus="CoreTest|JobSystemTest \(joltc"`
Expected: 4 specs PASS. Existing suite continues to pass (`go test ./jolt/...` afterwards).

- [ ] **Step 4: Run the full suite**

Run: `go test ./jolt/...`
Expected: PASS, no regressions.

- [ ] **Step 5: Commit**

```bash
git add jolt/port_core_test.go jolt/jolt_handwritten.go
git commit -m "test(jolt): port joltc test_core.cpp specs (Phase A)"
```

---

## Phase B — Shapes

Phase B is split into 8 tasks, one per logical group of specs in `test_shapes.cpp`. Each task: add carve-outs / Settings constructors as needed, port the specs, run, commit.

### Task 4: SphereShape (5 specs)

Specs: `SphereShape_Create`, `SphereShape_CreateWithSettings`, `SphereShapeSettings_SetRadius`, `SphereShape_GetType`, `SphereShape_GetVolume`.

**Files:**
- Create: `jolt/port_shapes_test.go` (new)
- Modify: `jolt/jolt_handwritten.go` (add `NewSphereShapeSettings`)

- [ ] **Step 1: Add the missing Settings constructor**

Append to `jolt/jolt_handwritten.go` (after `NewEmptyShapeSettings`):

```go
// NewSphereShapeSettings creates a SphereShapeSettings with the given radius.
func NewSphereShapeSettings(radius float32) *SphereShapeSettings {
	h := cjolt.JPH_SphereShapeSettings_Create(radius)
	return &SphereShapeSettings{h: h}
}
```

- [ ] **Step 2: Create port_shapes_test.go with the SphereShape Describe**

```go
package jolt_test

// port_shapes_test.go ports third_party/joltc/tests/test_shapes.cpp.

import (
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gataengine/jolt-go/jolt"
)

var _ = Describe("ShapesTest (joltc test_shapes.cpp)", func() {
	BeforeEach(func() {
		Expect(jolt.Init()).To(BeTrue())
	})
	AfterEach(func() {
		jolt.Shutdown()
	})

	Describe("SphereShape", func() {
		It("SphereShape_Create", func() {
			shape := jolt.NewSphereShape(1.0)
			Expect(shape).NotTo(BeNil())
			Expect(shape.Radius()).To(BeNumerically("~", 1.0, 1e-5))
		})

		It("SphereShape_CreateWithSettings", func() {
			settings := jolt.NewSphereShapeSettings(2.0)
			Expect(settings).NotTo(BeNil())
			Expect(settings.Radius()).To(BeNumerically("~", 2.0, 1e-5))

			shape := settings.CreateShape()
			Expect(shape).NotTo(BeNil())
			Expect(shape.Radius()).To(BeNumerically("~", 2.0, 1e-5))
		})

		It("SphereShapeSettings_SetRadius", func() {
			settings := jolt.NewSphereShapeSettings(1.0)
			Expect(settings).NotTo(BeNil())

			settings.SetRadius(3.0)
			Expect(settings.Radius()).To(BeNumerically("~", 3.0, 1e-5))
		})

		It("SphereShape_GetType", func() {
			shape := jolt.NewSphereShape(1.0)
			Expect(shape.Type()).To(Equal(jolt.ShapeTypeConvex))
			Expect(shape.SubType()).To(Equal(jolt.ShapeSubTypeSphere))
		})

		It("SphereShape_GetVolume", func() {
			shape := jolt.NewSphereShape(1.0)
			expected := float32(4.0/3.0) * float32(math.Pi)
			Expect(shape.Volume()).To(BeNumerically("~", expected, 1e-3))
		})
	})
})
```

- [ ] **Step 3: Run focused tests**

Run: `go test ./jolt/... -v -ginkgo.focus="SphereShape"`
Expected: 5 specs PASS.

- [ ] **Step 4: Commit**

```bash
git add jolt/port_shapes_test.go jolt/jolt_handwritten.go
git commit -m "test(jolt): port SphereShape specs from test_shapes.cpp"
```

---

### Task 5: BoxShape (4 specs)

Specs: `BoxShape_Create`, `BoxShape_CreateWithSettings`, `BoxShape_GetConvexRadius`, `BoxShape_GetType`.

The `Create` and `CreateWithSettings` specs both call `JPH_BoxShape_GetHalfExtent` — out-pointer accessor → carve-out needed.

**Files:**
- Modify: `jolt/jolt_handwritten.go` (add `NewBoxShapeSettings`, `BoxShape.GetHalfExtent`)
- Modify: `jolt/port_shapes_test.go` (append BoxShape Describe inside the existing ShapesTest)

- [ ] **Step 1: Add Settings constructor + GetHalfExtent carve-out**

Append to `jolt/jolt_handwritten.go`:

```go
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
```

- [ ] **Step 2: Append the BoxShape Describe to port_shapes_test.go**

Inside the existing `Describe("ShapesTest …", func() { ... })` block, after the `Describe("SphereShape", …)` close, append:

```go
	Describe("BoxShape", func() {
		It("BoxShape_Create", func() {
			shape := jolt.NewBoxShape(jolt.Vec3{X: 1, Y: 2, Z: 3}, jolt.DefaultConvexRadius)
			Expect(shape).NotTo(BeNil())

			h := shape.GetHalfExtent()
			Expect(h.X).To(BeNumerically("~", 1.0, 1e-5))
			Expect(h.Y).To(BeNumerically("~", 2.0, 1e-5))
			Expect(h.Z).To(BeNumerically("~", 3.0, 1e-5))
		})

		It("BoxShape_CreateWithSettings", func() {
			settings := jolt.NewBoxShapeSettings(jolt.Vec3{X: 2, Y: 2, Z: 2}, jolt.DefaultConvexRadius)
			Expect(settings).NotTo(BeNil())

			shape := settings.CreateShape()
			Expect(shape).NotTo(BeNil())

			h := shape.GetHalfExtent()
			Expect(h.X).To(BeNumerically("~", 2.0, 1e-5))
			Expect(h.Y).To(BeNumerically("~", 2.0, 1e-5))
			Expect(h.Z).To(BeNumerically("~", 2.0, 1e-5))
		})

		It("BoxShape_GetConvexRadius", func() {
			shape := jolt.NewBoxShape(jolt.Vec3{X: 1, Y: 1, Z: 1}, 0.1)
			Expect(shape.ConvexRadius()).To(BeNumerically("~", 0.1, 1e-5))
		})

		It("BoxShape_GetType", func() {
			shape := jolt.NewBoxShape(jolt.Vec3{X: 1, Y: 1, Z: 1}, jolt.DefaultConvexRadius)
			Expect(shape.Type()).To(Equal(jolt.ShapeTypeConvex))
			Expect(shape.SubType()).To(Equal(jolt.ShapeSubTypeBox))
		})
	})
```

- [ ] **Step 3: Run focused tests**

Run: `go test ./jolt/... -v -ginkgo.focus="BoxShape"`
Expected: 4 specs PASS.

- [ ] **Step 4: Commit**

```bash
git add jolt/port_shapes_test.go jolt/jolt_handwritten.go
git commit -m "test(jolt): port BoxShape specs + GetHalfExtent carve-out"
```

---

### Task 6: CapsuleShape, CylinderShape, TriangleShape, TaperedCapsule, TaperedCylinder (8 specs)

Specs: `CapsuleShape_Create`, `CapsuleShape_CreateWithSettings`, `CapsuleShape_GetType`, `CylinderShape_Create`, `CylinderShape_CreateWithSettings`, `CylinderShape_GetType`, `TriangleShape_Create`, `TriangleShape_GetType`, `TaperedCapsuleShape_Create`, `TaperedCapsuleShape_GetType`, `TaperedCylinderShape_Create`.

Carve-outs needed: `TriangleShape.GetVertex1/2/3` (all out-pointer).

**Files:**
- Modify: `jolt/jolt_handwritten.go` (add `NewCapsuleShapeSettings`, `NewCylinderShapeSettings`, three `TriangleShape.GetVertexN` carve-outs)
- Modify: `jolt/port_shapes_test.go` (append five Describe blocks)

- [ ] **Step 1: Add Settings constructors + Triangle vertex carve-outs**

Append to `jolt/jolt_handwritten.go`:

```go
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
```

- [ ] **Step 2: Append Describe blocks to port_shapes_test.go**

Append inside `Describe("ShapesTest …")`:

```go
	Describe("CapsuleShape", func() {
		It("CapsuleShape_Create", func() {
			shape := jolt.NewCapsuleShape(1.0, 0.5)
			Expect(shape).NotTo(BeNil())
			Expect(shape.HalfHeightOfCylinder()).To(BeNumerically("~", 1.0, 1e-5))
			Expect(shape.Radius()).To(BeNumerically("~", 0.5, 1e-5))
		})

		It("CapsuleShape_CreateWithSettings", func() {
			settings := jolt.NewCapsuleShapeSettings(2.0, 1.0)
			Expect(settings).NotTo(BeNil())

			shape := settings.CreateShape()
			Expect(shape).NotTo(BeNil())
			Expect(shape.HalfHeightOfCylinder()).To(BeNumerically("~", 2.0, 1e-5))
			Expect(shape.Radius()).To(BeNumerically("~", 1.0, 1e-5))
		})

		It("CapsuleShape_GetType", func() {
			shape := jolt.NewCapsuleShape(1.0, 0.5)
			Expect(shape.SubType()).To(Equal(jolt.ShapeSubTypeCapsule))
		})
	})

	Describe("CylinderShape", func() {
		It("CylinderShape_Create", func() {
			shape := jolt.NewCylinderShape(1.0, 0.5)
			Expect(shape).NotTo(BeNil())
			Expect(shape.HalfHeight()).To(BeNumerically("~", 1.0, 1e-5))
			Expect(shape.Radius()).To(BeNumerically("~", 0.5, 1e-5))
		})

		It("CylinderShape_CreateWithSettings", func() {
			settings := jolt.NewCylinderShapeSettings(2.0, 1.0, jolt.DefaultConvexRadius)
			Expect(settings).NotTo(BeNil())

			shape := settings.CreateShape()
			Expect(shape).NotTo(BeNil())
			Expect(shape.HalfHeight()).To(BeNumerically("~", 2.0, 1e-5))
			Expect(shape.Radius()).To(BeNumerically("~", 1.0, 1e-5))
		})

		It("CylinderShape_GetType", func() {
			shape := jolt.NewCylinderShape(1.0, 0.5)
			Expect(shape.SubType()).To(Equal(jolt.ShapeSubTypeCylinder))
		})
	})

	Describe("TriangleShape", func() {
		v1 := jolt.Vec3{X: 0, Y: 0, Z: 0}
		v2 := jolt.Vec3{X: 1, Y: 0, Z: 0}
		v3 := jolt.Vec3{X: 0, Y: 1, Z: 0}

		It("TriangleShape_Create", func() {
			shape := jolt.NewTriangleShape(v1, v2, v3, 0.0)
			Expect(shape).NotTo(BeNil())

			Expect(shape.GetVertex1()).To(Equal(v1))
			Expect(shape.GetVertex2()).To(Equal(v2))
			Expect(shape.GetVertex3()).To(Equal(v3))
		})

		It("TriangleShape_GetType", func() {
			shape := jolt.NewTriangleShape(v1, v2, v3, 0.0)
			Expect(shape.SubType()).To(Equal(jolt.ShapeSubTypeTriangle))
		})
	})

	Describe("TaperedCapsuleShape", func() {
		It("TaperedCapsuleShape_Create", func() {
			settings := jolt.NewTaperedCapsuleShapeSettings(1.0, 0.5, 0.3)
			Expect(settings).NotTo(BeNil())

			shape := settings.CreateShape()
			Expect(shape).NotTo(BeNil())
			Expect(shape.TopRadius()).To(BeNumerically("~", 0.5, 1e-5))
			Expect(shape.BottomRadius()).To(BeNumerically("~", 0.3, 1e-5))
			Expect(shape.HalfHeight()).To(BeNumerically("~", 1.0, 1e-5))
		})

		It("TaperedCapsuleShape_GetType", func() {
			settings := jolt.NewTaperedCapsuleShapeSettings(1.0, 0.5, 0.3)
			shape := settings.CreateShape()
			Expect(shape.SubType()).To(Equal(jolt.ShapeSubTypeTaperedCapsule))
		})
	})

	Describe("TaperedCylinderShape", func() {
		It("TaperedCylinderShape_Create", func() {
			settings := jolt.NewTaperedCylinderShapeSettings(1.0, 0.5, 1.0, jolt.DefaultConvexRadius)
			Expect(settings).NotTo(BeNil())

			shape := settings.CreateShape()
			Expect(shape).NotTo(BeNil())
			Expect(shape.TopRadius()).To(BeNumerically("~", 0.5, 1e-5))
			Expect(shape.BottomRadius()).To(BeNumerically("~", 1.0, 1e-5))
			Expect(shape.HalfHeight()).To(BeNumerically("~", 1.0, 1e-5))
		})
	})
```

- [ ] **Step 3: Run focused tests**

Run: `go test ./jolt/... -v -ginkgo.focus="CapsuleShape|CylinderShape|TriangleShape|TaperedCapsuleShape|TaperedCylinderShape"`
Expected: 11 specs PASS.

- [ ] **Step 4: Commit**

```bash
git add jolt/port_shapes_test.go jolt/jolt_handwritten.go
git commit -m "test(jolt): port Capsule/Cylinder/Triangle/Tapered shape specs + carve-outs"
```

---

### Task 7: ConvexHullShape and PlaneShape (3 specs)

Specs: `ConvexHullShape_Create`, `ConvexHullShape_GetPoint`, `PlaneShape_Create`.

Carve-outs needed: `ConvexHullShape.GetPoint(i uint32) Vec3`, `PlaneShape.GetPlane() GoPlane`.

**Files:**
- Modify: `jolt/jolt_handwritten.go` (add `NewConvexHullShapeSettings`, two carve-outs)
- Modify: `jolt/port_shapes_test.go` (append two Describe blocks)

- [ ] **Step 1: Add Settings constructor + carve-outs**

Append to `jolt/jolt_handwritten.go`:

```go
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
```

- [ ] **Step 2: Append Describe blocks to port_shapes_test.go**

Append inside `Describe("ShapesTest …")`:

```go
	Describe("ConvexHullShape", func() {
		// Tetrahedron — same vertices as the joltc test.
		points := []jolt.Vec3{
			{X: 0, Y: 0, Z: 0},
			{X: 1, Y: 0, Z: 0},
			{X: 0.5, Y: 1, Z: 0},
			{X: 0.5, Y: 0.5, Z: 1},
		}

		It("ConvexHullShape_Create", func() {
			settings := jolt.NewConvexHullShapeSettings(points, jolt.DefaultConvexRadius)
			Expect(settings).NotTo(BeNil())

			shape := settings.CreateShape()
			Expect(shape).NotTo(BeNil())
			Expect(shape.NumPoints()).To(Equal(uint32(4)))
			Expect(shape.NumFaces()).To(Equal(uint32(4))) // tetrahedron
		})

		It("ConvexHullShape_GetPoint", func() {
			settings := jolt.NewConvexHullShapeSettings(points, jolt.DefaultConvexRadius)
			shape := settings.CreateShape()
			Expect(shape).NotTo(BeNil())

			// Hull may reorder points; just assert non-NaN.
			p := shape.GetPoint(0)
			Expect(math.IsNaN(float64(p.X))).To(BeFalse())
			Expect(math.IsNaN(float64(p.Y))).To(BeFalse())
			Expect(math.IsNaN(float64(p.Z))).To(BeFalse())
		})
	})

	Describe("PlaneShape", func() {
		It("PlaneShape_Create", func() {
			plane := jolt.GoPlane{Normal: jolt.Vec3{X: 0, Y: 1, Z: 0}, Distance: 0}
			shape := jolt.NewPlaneShapeFromGoPlane(plane, nil, 1000.0)
			Expect(shape).NotTo(BeNil())

			result := shape.GetPlane()
			Expect(result.Normal.X).To(BeNumerically("~", 0.0, 1e-5))
			Expect(result.Normal.Y).To(BeNumerically("~", 1.0, 1e-5))
			Expect(result.Normal.Z).To(BeNumerically("~", 0.0, 1e-5))
			Expect(result.Distance).To(BeNumerically("~", 0.0, 1e-5))

			Expect(shape.HalfExtent()).To(BeNumerically("~", 1000.0, 1e-3))
		})
	})
```

- [ ] **Step 3: Run focused tests**

Run: `go test ./jolt/... -v -ginkgo.focus="ConvexHullShape|PlaneShape"`
Expected: 3 specs PASS.

- [ ] **Step 4: Commit**

```bash
git add jolt/port_shapes_test.go jolt/jolt_handwritten.go
git commit -m "test(jolt): port ConvexHull/Plane shape specs + GetPoint/GetPlane carve-outs"
```

---

### Task 8: Shape common API (6 specs)

Specs: `Shape_UserData`, `Shape_GetLocalBounds`, `Shape_GetCenterOfMass`, `Shape_GetInnerRadius`, `Shape_GetMassProperties`, `Shape_MustBeStatic`.

Carve-outs needed: `Shape.GetLocalBounds() AABox`, `Shape.GetCenterOfMass() Vec3`. The other accessors return by value.

**Files:**
- Modify: `jolt/jolt_handwritten.go` (add two carve-outs)
- Modify: `jolt/port_shapes_test.go` (append Describe block)

- [ ] **Step 1: Add carve-outs**

Append to `jolt/jolt_handwritten.go`:

```go
// GetLocalBounds returns the local-space AABB of the shape.
func (s *Shape) GetLocalBounds() AABox {
	var out AABox
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
```

- [ ] **Step 2: Append the Describe block**

Append inside `Describe("ShapesTest …")`:

```go
	Describe("Shape (base)", func() {
		It("Shape_UserData", func() {
			shape := jolt.NewSphereShape(1.0)
			shape.SetUserData(12345)
			Expect(shape.UserData()).To(Equal(uint64(12345)))
		})

		It("Shape_GetLocalBounds", func() {
			shape := jolt.NewSphereShape(1.0)
			b := shape.GetLocalBounds()
			Expect(float32(b.Min.x)).To(BeNumerically("~", -1.0, 1e-5))
			Expect(float32(b.Min.y)).To(BeNumerically("~", -1.0, 1e-5))
			Expect(float32(b.Min.z)).To(BeNumerically("~", -1.0, 1e-5))
			Expect(float32(b.Max.x)).To(BeNumerically("~", 1.0, 1e-5))
			Expect(float32(b.Max.y)).To(BeNumerically("~", 1.0, 1e-5))
			Expect(float32(b.Max.z)).To(BeNumerically("~", 1.0, 1e-5))
		})

		It("Shape_GetCenterOfMass", func() {
			shape := jolt.NewSphereShape(1.0)
			com := shape.GetCenterOfMass()
			Expect(com.X).To(BeNumerically("~", 0.0, 1e-5))
			Expect(com.Y).To(BeNumerically("~", 0.0, 1e-5))
			Expect(com.Z).To(BeNumerically("~", 0.0, 1e-5))
		})

		It("Shape_GetInnerRadius", func() {
			shape := jolt.NewSphereShape(1.0)
			Expect(shape.InnerRadius()).To(BeNumerically("~", 1.0, 1e-5))
		})

		It("Shape_MustBeStatic", func() {
			sphere := jolt.NewSphereShape(1.0)
			Expect(sphere.MustBeStatic()).To(BeFalse())

			plane := jolt.GoPlane{Normal: jolt.Vec3{X: 0, Y: 1, Z: 0}, Distance: 0}
			planeShape := jolt.NewPlaneShapeFromGoPlane(plane, nil, 1000.0)
			Expect(planeShape.MustBeStatic()).To(BeTrue())
		})
	})
```

> **Note on `AABox`:** `jolt.AABox` is a CGo type alias (`= cjolt.C_JPH_AABox`). The fields `Min` and `Max` are `JPH_Vec3` whose components in CGo are unexported (`x`, `y`, `z`). The cast to `float32` plus lowercase access works because we're inside the same package boundary as the cjolt alias. Verify by running the test — if the lowercase access fails to compile, fall back to extracting via `unsafe.Pointer` to a `Vec3` view. (Mentioned here so the implementer doesn't blindly assume the type is Go-friendly.)

> If lowercase access fails: introduce a Go-native `GoAABox` mirror struct in `jolt_handwritten.go` (same pattern as `GoPlane`), have `GetLocalBounds()` return `GoAABox`, and update the test accordingly.

> **Skipped (deferred):** `Shape_GetMassProperties` requires accessor methods on the `MassProperties` type (`Mass`, `InertiaTensor`, etc.) that the generator does not currently emit on the bare struct. Add a `Pending("Shape_GetMassProperties")` placeholder spec referencing this gap.

Add this skip spec inside the `Shape (base)` Describe:

```go
		It("Shape_GetMassProperties", Pending, func() {
			// MassProperties accessors not yet emitted by the generator.
			// joltc test asserts massProps.mass > 0.
		})
```

- [ ] **Step 3: Run focused tests**

Run: `go test ./jolt/... -v -ginkgo.focus="Shape \(base\)"`
Expected: 5 specs PASS, 1 pending. If any test fails to compile due to the AABox field access, fall back to the `GoAABox` mirror.

- [ ] **Step 4: Commit**

```bash
git add jolt/port_shapes_test.go jolt/jolt_handwritten.go
git commit -m "test(jolt): port Shape base-class specs + GetLocalBounds/GetCenterOfMass carve-outs"
```

---

### Task 9: ConvexShape density + ShapeSettings UserData (3 specs)

Specs: `ConvexShape_Density`, `ConvexShape_SetDensity`, `ShapeSettings_UserData`.

These all use accessors that already return-by-value or take a primitive — no carve-outs needed. The first two test `ConvexShapeSettings.GetDensity/SetDensity` (on settings, not shape); the generator's emitted methods are `(*ConvexShape).Density()` (on the *built* shape). Verify the joltc test variant — first test uses `JPH_ConvexShapeSettings_GetDensity` on a `JPH_SphereShapeSettings*`, which means the SettingsType must expose `Density()`. Let me check the actual generated bindings before writing.

**Files:**
- Modify: `jolt/port_shapes_test.go` (append Describe)

- [ ] **Step 1: Verify generated SettingsType density methods exist**

Run: `grep -nE "ConvexShapeSettings\) (Density|SetDensity)|^func.*ShapeSettings.*Density" jolt/jolt_generated.go`
Expected output should include `func (cs *ConvexShapeSettings) Density() float32` (or similar). If absent, **stop** and use `ConvexShape` (the built shape) for the test instead — both joltc tests assert the same density round-trip behaviour, just on different objects.

- [ ] **Step 2: Append Describe**

Append inside `Describe("ShapesTest …")`:

```go
	Describe("ConvexShape Density", func() {
		It("ConvexShape_Density", func() {
			shape := jolt.NewSphereShape(1.0)
			defaultDensity := shape.Density()
			Expect(defaultDensity).To(BeNumerically(">", 0.0))

			shape.SetDensity(2000.0)
			Expect(shape.Density()).To(BeNumerically("~", 2000.0, 1e-3))
		})

		It("ConvexShape_SetDensity", func() {
			shape := jolt.NewSphereShape(1.0)
			defaultDensity := shape.Density()
			Expect(defaultDensity).To(BeNumerically(">", 0.0))

			shape.SetDensity(5000.0)
			Expect(shape.Density()).To(BeNumerically("~", 5000.0, 1e-3))
		})
	})

	Describe("ShapeSettings UserData", func() {
		It("ShapeSettings_UserData", func() {
			settings := jolt.NewSphereShapeSettings(1.0)
			Expect(settings).NotTo(BeNil())

			// The joltc test calls JPH_ShapeSettings_SetUserData on the upcast
			// pointer. SphereShapeSettings embeds ConvexShapeSettings which
			// embeds ShapeSettings — methods are promoted.
			settings.SetUserData(99999)
			Expect(settings.UserData()).To(Equal(uint64(99999)))
		})
	})
```

- [ ] **Step 3: Run focused tests**

Run: `go test ./jolt/... -v -ginkgo.focus="ConvexShape Density|ShapeSettings UserData"`
Expected: 3 specs PASS.

- [ ] **Step 4: Commit**

```bash
git add jolt/port_shapes_test.go
git commit -m "test(jolt): port ConvexShape density + ShapeSettings UserData specs"
```

---

### Task 10: Decorated and compound shapes (5 specs)

Specs: `RotatedTranslatedShape_Create`, `ScaledShape_Create`, `OffsetCenterOfMassShape_Create`, `StaticCompoundShape_Create`, `MutableCompoundShape_Create`, `EmptyShape_Create`.

Carve-outs needed:
- `RotatedTranslatedShape.GetPosition() Vec3`
- `RotatedTranslatedShape.GetRotation() Quat`
- `ScaledShape.GetScale() Vec3`
- `OffsetCenterOfMassShape.GetOffset() Vec3`
- `DecoratedShape.GetInnerShape() *Shape`

`StaticCompoundShape_Create` and `MutableCompoundShape_Create` need `NewStaticCompoundShapeSettings` / `NewMutableCompoundShapeSettings` constructors.

**Files:**
- Modify: `jolt/jolt_handwritten.go`
- Modify: `jolt/port_shapes_test.go` (append Describe)

- [ ] **Step 1: Add constructors and carve-outs**

Append to `jolt/jolt_handwritten.go`:

```go
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
```

> **Note on `DecoratedShape.GetInnerShape`:** the joltc test `RotatedTranslatedShape_Create` calls `JPH_DecoratedShape_GetInnerShape` and asserts pointer equality with the input shape. We can write that test only if the generator already emits `(*DecoratedShape).InnerShape()` (returning `*Shape`). Verify with `grep -E 'DecoratedShape\) (InnerShape|GetInnerShape)' jolt/jolt_generated.go`. If absent, mark that sub-assertion as a Pending placeholder (`Pending("DecoratedShape.GetInnerShape not yet bound")`).

- [ ] **Step 2: Append the Describe block**

Append inside `Describe("ShapesTest …")`:

```go
	Describe("Decorated shapes", func() {
		It("RotatedTranslatedShape_Create", func() {
			inner := jolt.NewSphereShape(1.0)
			pos := jolt.Vec3{X: 1, Y: 2, Z: 3}
			rot := jolt.Quat{X: 0, Y: 0, Z: 0, W: 1}

			shape := jolt.NewRotatedTranslatedShape(pos, rot, &inner.Shape)
			Expect(shape).NotTo(BeNil())

			Expect(shape.GetPosition()).To(Equal(pos))
			Expect(shape.GetRotation().W).To(BeNumerically("~", 1.0, 1e-5))
		})

		It("ScaledShape_Create", func() {
			inner := jolt.NewSphereShape(1.0)
			scale := jolt.Vec3{X: 2, Y: 2, Z: 2}

			shape := jolt.NewScaledShape(&inner.Shape, scale)
			Expect(shape).NotTo(BeNil())
			Expect(shape.GetScale()).To(Equal(scale))
		})

		It("OffsetCenterOfMassShape_Create", func() {
			inner := jolt.NewSphereShape(1.0)
			offset := jolt.Vec3{X: 0, Y: 1, Z: 0}

			shape := jolt.NewOffsetCenterOfMassShape(offset, &inner.Shape)
			Expect(shape).NotTo(BeNil())
			Expect(shape.GetOffset()).To(Equal(offset))
		})
	})

	Describe("Compound shapes", func() {
		It("StaticCompoundShape_Create", func() {
			settings := jolt.NewStaticCompoundShapeSettings()
			Expect(settings).NotTo(BeNil())

			s1 := jolt.NewSphereShape(1.0)
			s2 := jolt.NewSphereShape(0.5)
			rot := jolt.Quat{X: 0, Y: 0, Z: 0, W: 1}

			settings.AddShape2(jolt.Vec3{X: -1, Y: 0, Z: 0}, rot, &s1.Shape, 0)
			settings.AddShape2(jolt.Vec3{X: 1, Y: 0, Z: 0}, rot, &s2.Shape, 0)

			shape := jolt.NewStaticCompoundShape(settings)
			Expect(shape).NotTo(BeNil())
			Expect(shape.NumSubShapes()).To(Equal(uint32(2)))
		})

		It("MutableCompoundShape_Create", func() {
			settings := jolt.NewMutableCompoundShapeSettings()
			Expect(settings).NotTo(BeNil())

			shape := jolt.NewMutableCompoundShape(settings)
			Expect(shape).NotTo(BeNil())

			s := jolt.NewSphereShape(1.0)
			rot := jolt.Quat{X: 0, Y: 0, Z: 0, W: 1}

			idx := shape.AddShape(jolt.Vec3{X: 0, Y: 0, Z: 0}, rot, &s.Shape, 0, ^uint32(0))
			Expect(idx).To(Equal(uint32(0)))
			Expect(shape.NumSubShapes()).To(Equal(uint32(1)))

			shape.RemoveShape(0)
			Expect(shape.NumSubShapes()).To(Equal(uint32(0)))
		})
	})

	Describe("EmptyShape", func() {
		It("EmptyShape_Create", func() {
			settings := jolt.NewEmptyShapeSettings(jolt.Vec3{})
			Expect(settings).NotTo(BeNil())

			shape := settings.CreateShape()
			Expect(shape).NotTo(BeNil())
			Expect(shape.Volume()).To(BeNumerically("~", 0.0, 1e-5))
		})
	})
```

- [ ] **Step 3: Run focused tests**

Run: `go test ./jolt/... -v -ginkgo.focus="Decorated shapes|Compound shapes|EmptyShape"`
Expected: 6 specs PASS.

- [ ] **Step 4: Commit**

```bash
git add jolt/port_shapes_test.go jolt/jolt_handwritten.go
git commit -m "test(jolt): port decorated/compound/empty shape specs + carve-outs"
```

---

### Task 11: Shape ray cast and point collision (4 specs)

Specs: `Shape_CastRay`, `Shape_CastRay_Miss`, `Shape_CollidePoint_Inside`, `Shape_CollidePoint_Outside`.

The generated `(*Shape).CastRay(o, d Vec3, h unsafe.Pointer) bool` accepts the result via `unsafe.Pointer` — workable but ugly. Add a typed helper carve-out for clarity.

**Files:**
- Modify: `jolt/jolt_handwritten.go` (one helper)
- Modify: `jolt/port_shapes_test.go` (append Describe)

- [ ] **Step 1: Add the typed CastRay helper**

Append to `jolt/jolt_handwritten.go`:

```go
// CastRayResult is a typed wrapper for Shape.CastRay that returns the hit
// info via a Go-native struct rather than an unsafe.Pointer.
//
// Returns (hit, true) on a hit and (zero, false) on a miss.
func (s *Shape) CastRayResult(origin, direction Vec3) (RayCastResult, bool) {
	var out RayCastResult
	hit := cjolt.JPH_Shape_CastRay(s.h, (*cjolt.C_JPH_Vec3)(unsafe.Pointer(&origin)), (*cjolt.C_JPH_Vec3)(unsafe.Pointer(&direction)), (*cjolt.C_JPH_RayCastResult)(unsafe.Pointer(&out)))
	runtime.KeepAlive(s)
	if !hit {
		return RayCastResult{}, false
	}
	return out, true
}
```

> **Note:** `RayCastResult` is `= cjolt.C_JPH_RayCastResult`, a CGo struct alias. Its field `fraction` (lowercase) may not be accessible across the package boundary the same way as `AABox` fields. Test by running — if it fails to compile, introduce a `GoRayCastResult` Go-native mirror struct (same pattern as `GoPlane`) and update accordingly. The cjolt struct layout: `JPH_BodyID bodyID` (uint32), `JPH_SubShapeID subShapeID` (uint32), `float fraction` — total 12 bytes. The Go-native mirror would be:
>
> ```go
> type GoRayCastResult struct {
>     BodyID     BodyID
>     SubShapeID SubShapeID
>     Fraction   float32
> }
> ```

- [ ] **Step 2: Append the Describe**

Append inside `Describe("ShapesTest …")`:

```go
	Describe("Shape CastRay", func() {
		It("Shape_CastRay", func() {
			shape := jolt.NewSphereShape(1.0)
			origin := jolt.Vec3{X: -5, Y: 0, Z: 0}
			direction := jolt.Vec3{X: 10, Y: 0, Z: 0} // unnormalized; Jolt expects ray length encoded

			hit, ok := shape.CastRayResult(origin, direction)
			Expect(ok).To(BeTrue())
			// Hit at x = -1, fraction = (5 - 1) / 10 = 0.4
			Expect(float32(hit.fraction)).To(BeNumerically("~", 0.4, 0.01))
		})

		It("Shape_CastRay_Miss", func() {
			shape := jolt.NewSphereShape(1.0)
			origin := jolt.Vec3{X: 0, Y: 5, Z: 0}
			direction := jolt.Vec3{X: 1, Y: 0, Z: 0}

			_, ok := shape.CastRayResult(origin, direction)
			Expect(ok).To(BeFalse())
		})
	})

	Describe("Shape CollidePoint", func() {
		It("Shape_CollidePoint_Inside", func() {
			shape := jolt.NewSphereShape(1.0)
			Expect(shape.CollidePoint(jolt.Vec3{X: 0, Y: 0, Z: 0}, nil)).To(BeTrue())
		})

		It("Shape_CollidePoint_Outside", func() {
			shape := jolt.NewSphereShape(1.0)
			Expect(shape.CollidePoint(jolt.Vec3{X: 5, Y: 0, Z: 0}, nil)).To(BeFalse())
		})
	})
```

- [ ] **Step 3: Run focused tests**

Run: `go test ./jolt/... -v -ginkgo.focus="Shape CastRay|Shape CollidePoint"`
Expected: 4 specs PASS. If `hit.fraction` lowercase access fails, swap to `GoRayCastResult` per the note above.

- [ ] **Step 4: Commit**

```bash
git add jolt/port_shapes_test.go jolt/jolt_handwritten.go
git commit -m "test(jolt): port Shape CastRay/CollidePoint specs + CastRayResult helper"
```

---

### Task 12: Shape scale validation (2 specs)

Specs: `Shape_IsValidScale`, `Shape_MakeScaleValid`.

`MakeScaleValid` is an out-pointer — carve-out needed.

**Files:**
- Modify: `jolt/jolt_handwritten.go`
- Modify: `jolt/port_shapes_test.go` (append Describe)

- [ ] **Step 1: Add the carve-out**

Append to `jolt/jolt_handwritten.go`:

```go
// MakeScaleValidResult clamps the given scale to a valid value for this shape
// and returns the result.
func (s *Shape) MakeScaleValidResult(scale Vec3) Vec3 {
	var out Vec3
	cjolt.JPH_Shape_MakeScaleValid(s.h, (*cjolt.C_JPH_Vec3)(unsafe.Pointer(&scale)), (*cjolt.C_JPH_Vec3)(unsafe.Pointer(&out)))
	runtime.KeepAlive(s)
	return out
}
```

- [ ] **Step 2: Append the Describe**

Append inside `Describe("ShapesTest …")`:

```go
	Describe("Shape scale validation", func() {
		It("Shape_IsValidScale", func() {
			shape := jolt.NewSphereShape(1.0)
			Expect(shape.IsValidScale(jolt.Vec3{X: 2, Y: 2, Z: 2})).To(BeTrue())
			Expect(shape.IsValidScale(jolt.Vec3{X: 2, Y: 1, Z: 1})).To(BeFalse())
		})

		It("Shape_MakeScaleValid", func() {
			shape := jolt.NewSphereShape(1.0)
			fixed := shape.MakeScaleValidResult(jolt.Vec3{X: 2, Y: 1, Z: 1})
			Expect(shape.IsValidScale(fixed)).To(BeTrue())
		})
	})
```

- [ ] **Step 3: Run focused tests**

Run: `go test ./jolt/... -v -ginkgo.focus="Shape scale validation"`
Expected: 2 specs PASS.

- [ ] **Step 4: Run the full Phase B suite**

Run: `go test ./jolt/... -v -ginkgo.focus="ShapesTest"`
Expected: All Phase B specs (~37 + 1 pending) PASS.

- [ ] **Step 5: Commit**

```bash
git add jolt/port_shapes_test.go jolt/jolt_handwritten.go
git commit -m "test(jolt): port Shape scale-validation specs + MakeScaleValidResult carve-out"
```

---

## Phase C — PhysicsSystem & BodyInterface

### Task 13: System creation and queries (7 specs)

Specs: `Create_ValidSettings`, `GetBodyInterface`, `GetBodyInterfaceNoLock`, `SetGravity` (round-trip), `GetNumBodies_Initial`, `GetNumActiveBodies_Initial`, `GetMaxBodies`.

Carve-out needed: `PhysicsSystem.GetGravity() Vec3` (out-pointer).

**Files:**
- Modify: `jolt/jolt_handwritten.go`
- Create: `jolt/port_physics_system_test.go`

- [ ] **Step 1: Add GetGravity carve-out**

Append to `jolt/jolt_handwritten.go`:

```go
// GetGravity returns the world gravity vector currently configured on the
// PhysicsSystem.
func (ps *PhysicsSystem) GetGravity() Vec3 {
	var out Vec3
	cjolt.JPH_PhysicsSystem_GetGravity(ps.h, (*cjolt.C_JPH_Vec3)(unsafe.Pointer(&out)))
	runtime.KeepAlive(ps)
	return out
}
```

- [ ] **Step 2: Create port_physics_system_test.go**

```go
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
		js  *jolt.JobSystem
		ps  *portSystem
		bi  *jolt.BodyInterface
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
})
```

- [ ] **Step 3: Run focused tests**

Run: `go test ./jolt/... -v -ginkgo.focus="PhysicsSystemTest \(joltc.*"`
Expected: 7 specs PASS.

- [ ] **Step 4: Commit**

```bash
git add jolt/port_physics_system_test.go jolt/jolt_handwritten.go
git commit -m "test(jolt): port PhysicsSystem creation/queries specs + GetGravity carve-out"
```

---

### Task 14: Body lifecycle (2 specs)

Specs: `CreateBody_Sphere`, `CreateBody_Box`.

These exercise `BodyInterface.CreateBody` (returns `*Body`), `Body.ID()`, `BodyInterface.AddBody` (with both Activate variants), `PhysicsSystem.NumBodies` / `NumActiveBodies`, `BodyInterface.RemoveBody` / `DestroyBody`.

No new carve-outs needed.

**Files:**
- Modify: `jolt/port_physics_system_test.go` (append Describe inside the existing top-level Describe block)

- [ ] **Step 1: Append the lifecycle specs**

Append inside `Describe("PhysicsSystemTest …")`, after the existing `It` blocks:

```go
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
```

- [ ] **Step 2: Run focused tests**

Run: `go test ./jolt/... -v -ginkgo.focus="CreateBody_"`
Expected: 2 specs PASS.

- [ ] **Step 3: Commit**

```bash
git add jolt/port_physics_system_test.go
git commit -m "test(jolt): port Body lifecycle specs (CreateBody_Sphere/Box)"
```

---

### Task 15: Body kinematics (2 specs + GetPosition reuse)

Specs: `Body_GetPosition`, `Body_SetPosition`.

Both already use `BodyInterface.GetPosition` (carve-out exists).

**Files:**
- Modify: `jolt/port_physics_system_test.go` (append)

- [ ] **Step 1: Append the kinematics specs**

```go
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
```

- [ ] **Step 2: Run focused tests**

Run: `go test ./jolt/... -v -ginkgo.focus="Body_GetPosition|Body_SetPosition"`
Expected: 2 specs PASS.

- [ ] **Step 3: Commit**

```bash
git add jolt/port_physics_system_test.go
git commit -m "test(jolt): port Body GetPosition/SetPosition specs"
```

---

### Task 16: Body velocity & impulse (2 specs)

Specs: `Body_SetLinearVelocity`, `Body_AddImpulse`.

Carve-out needed: `BodyInterface.GetLinearVelocity(b BodyID) Vec3`.

**Files:**
- Modify: `jolt/jolt_handwritten.go`
- Modify: `jolt/port_physics_system_test.go` (append)

- [ ] **Step 1: Add GetLinearVelocity carve-out**

Append to `jolt/jolt_handwritten.go`:

```go
// GetLinearVelocity returns the linear velocity of the body with the given ID.
func (bi *BodyInterface) GetLinearVelocity(b BodyID) Vec3 {
	var out Vec3
	cjolt.JPH_BodyInterface_GetLinearVelocity(bi.h, cjolt.C_JPH_BodyID(b), (*cjolt.C_JPH_Vec3)(unsafe.Pointer(&out)))
	runtime.KeepAlive(bi)
	return out
}
```

- [ ] **Step 2: Append the specs**

```go
	It("Body_SetLinearVelocity", func() {
		shape := jolt.NewSphereShape(1.0)
		settings := jolt.NewBodyCreationSettingsWithShape(
			&shape.Shape,
			jolt.RVec3{X: 0, Y: 10, Z: 0},
			jolt.MotionTypeDynamic, jolt.ObjectLayerMoving,
		)
		body := bi.CreateBody(settings)
		bodyID := body.ID()
		bi.AddBody(bodyID, jolt.ActivationActivate)

		bi.SetLinearVelocity(bodyID, jolt.Vec3{X: 1, Y: 2, Z: 3})
		v := bi.GetLinearVelocity(bodyID)
		Expect(v.X).To(BeNumerically("~", 1.0, 1e-5))
		Expect(v.Y).To(BeNumerically("~", 2.0, 1e-5))
		Expect(v.Z).To(BeNumerically("~", 3.0, 1e-5))

		bi.RemoveAndDestroyBody(bodyID)
	})

	It("Body_AddImpulse", func() {
		shape := jolt.NewSphereShape(1.0)
		settings := jolt.NewBodyCreationSettingsWithShape(
			&shape.Shape,
			jolt.RVec3{X: 0, Y: 10, Z: 0},
			jolt.MotionTypeDynamic, jolt.ObjectLayerMoving,
		)
		body := bi.CreateBody(settings)
		bodyID := body.ID()
		bi.AddBody(bodyID, jolt.ActivationActivate)

		Expect(bi.GetLinearVelocity(bodyID).X).To(BeNumerically("~", 0.0, 1e-5))

		bi.AddImpulse(bodyID, jolt.Vec3{X: 10, Y: 0, Z: 0})
		Expect(bi.GetLinearVelocity(bodyID).X).To(BeNumerically(">", 0.0))

		bi.RemoveAndDestroyBody(bodyID)
	})
```

- [ ] **Step 3: Run focused tests**

Run: `go test ./jolt/... -v -ginkgo.focus="Body_SetLinearVelocity|Body_AddImpulse"`
Expected: 2 specs PASS.

- [ ] **Step 4: Commit**

```bash
git add jolt/port_physics_system_test.go jolt/jolt_handwritten.go
git commit -m "test(jolt): port Body velocity/impulse specs + GetLinearVelocity carve-out"
```

---

### Task 17: Body motion type, activation, restitution, friction (4 specs)

Specs: `Body_GetMotionType`, `Body_ActivateDeactivate`, `Body_Restitution`, `Body_Friction`.

All four use accessors that already return-by-value. No carve-outs needed.

**Files:**
- Modify: `jolt/port_physics_system_test.go` (append)

- [ ] **Step 1: Append the specs**

```go
	It("Body_GetMotionType", func() {
		shape := jolt.NewSphereShape(1.0)
		settings := jolt.NewBodyCreationSettingsWithShape(
			&shape.Shape,
			jolt.RVec3{X: 0, Y: 10, Z: 0},
			jolt.MotionTypeDynamic, jolt.ObjectLayerMoving,
		)
		body := bi.CreateBody(settings)
		bodyID := body.ID()
		bi.AddBody(bodyID, jolt.ActivationActivate)

		Expect(bi.MotionType(bodyID)).To(Equal(jolt.MotionTypeDynamic))

		bi.RemoveAndDestroyBody(bodyID)
	})

	It("Body_ActivateDeactivate", func() {
		shape := jolt.NewSphereShape(1.0)
		settings := jolt.NewBodyCreationSettingsWithShape(
			&shape.Shape,
			jolt.RVec3{X: 0, Y: 10, Z: 0},
			jolt.MotionTypeDynamic, jolt.ObjectLayerMoving,
		)
		body := bi.CreateBody(settings)
		bodyID := body.ID()

		bi.AddBody(bodyID, jolt.ActivationDontActivate)
		Expect(bi.IsActive(bodyID)).To(BeFalse())

		bi.ActivateBody(bodyID)
		Expect(bi.IsActive(bodyID)).To(BeTrue())

		bi.DeactivateBody(bodyID)
		Expect(bi.IsActive(bodyID)).To(BeFalse())

		bi.RemoveAndDestroyBody(bodyID)
	})

	It("Body_Restitution", func() {
		shape := jolt.NewSphereShape(1.0)
		settings := jolt.NewBodyCreationSettingsWithShape(
			&shape.Shape,
			jolt.RVec3{X: 0, Y: 10, Z: 0},
			jolt.MotionTypeDynamic, jolt.ObjectLayerMoving,
		)
		body := bi.CreateBody(settings)
		bodyID := body.ID()
		bi.AddBody(bodyID, jolt.ActivationActivate)

		bi.SetRestitution(bodyID, 0.8)
		Expect(bi.Restitution(bodyID)).To(BeNumerically("~", 0.8, 1e-5))

		bi.RemoveAndDestroyBody(bodyID)
	})

	It("Body_Friction", func() {
		shape := jolt.NewSphereShape(1.0)
		settings := jolt.NewBodyCreationSettingsWithShape(
			&shape.Shape,
			jolt.RVec3{X: 0, Y: 10, Z: 0},
			jolt.MotionTypeDynamic, jolt.ObjectLayerMoving,
		)
		body := bi.CreateBody(settings)
		bodyID := body.ID()
		bi.AddBody(bodyID, jolt.ActivationActivate)

		bi.SetFriction(bodyID, 0.5)
		Expect(bi.Friction(bodyID)).To(BeNumerically("~", 0.5, 1e-5))

		bi.RemoveAndDestroyBody(bodyID)
	})
```

- [ ] **Step 2: Run focused tests**

Run: `go test ./jolt/... -v -ginkgo.focus="Body_GetMotionType|Body_ActivateDeactivate|Body_Restitution|Body_Friction"`
Expected: 4 specs PASS.

- [ ] **Step 3: Commit**

```bash
git add jolt/port_physics_system_test.go
git commit -m "test(jolt): port Body motion-type/activate/restitution/friction specs"
```

---

### Task 18: Simulation specs (2 specs)

Specs: `Update_FallingBody`, `MultipleBodies`.

`Update_FallingBody` calls `JPH_PhysicsSystem_Update` (single-step variant, no TempAllocator); the generated method is `(*PhysicsSystem).Update(d float32, c int32, j *JobSystem) PhysicsUpdateError`.

**Files:**
- Modify: `jolt/port_physics_system_test.go` (append)

- [ ] **Step 1: Append the simulation specs**

```go
	// ------------------------------------------------------------------
	// Simulation
	// ------------------------------------------------------------------

	It("Update_FallingBody", func() {
		shape := jolt.NewSphereShape(1.0)
		settings := jolt.NewBodyCreationSettingsWithShape(
			&shape.Shape,
			jolt.RVec3{X: 0, Y: 100, Z: 0},
			jolt.MotionTypeDynamic, jolt.ObjectLayerMoving,
		)
		body := bi.CreateBody(settings)
		bodyID := body.ID()
		bi.AddBody(bodyID, jolt.ActivationActivate)

		ps.System.SetGravity(jolt.Vec3{X: 0, Y: -10, Z: 0})

		initialPos := bi.GetPosition(bodyID)

		ps.System.OptimizeBroadPhase()
		err := ps.System.Update(1.0/60.0, 1, js)
		Expect(err).To(Equal(jolt.PhysicsUpdateErrorNone))

		newPos := bi.GetPosition(bodyID)
		Expect(float32(newPos.Y)).To(BeNumerically("<", float32(initialPos.Y)))

		bi.RemoveAndDestroyBody(bodyID)
	})

	It("MultipleBodies", func() {
		const numBodies = 10
		bodyIDs := make([]jolt.BodyID, numBodies)

		for i := range numBodies {
			shape := jolt.NewSphereShape(1.0)
			settings := jolt.NewBodyCreationSettingsWithShape(
				&shape.Shape,
				jolt.RVec3{X: jolt.Real(float32(i) * 3.0), Y: 10, Z: 0},
				jolt.MotionTypeDynamic, jolt.ObjectLayerMoving,
			)
			body := bi.CreateBody(settings)
			bodyIDs[i] = body.ID()
			bi.AddBody(bodyIDs[i], jolt.ActivationActivate)
		}

		Expect(ps.System.NumBodies()).To(Equal(uint32(numBodies)))
		Expect(ps.System.NumActiveBodies(jolt.BodyTypeRigid)).To(Equal(uint32(numBodies)))

		for _, id := range bodyIDs {
			bi.RemoveAndDestroyBody(id)
		}
	})
```

- [ ] **Step 2: Run focused tests**

Run: `go test ./jolt/... -v -ginkgo.focus="Update_FallingBody|MultipleBodies"`
Expected: 2 specs PASS.

- [ ] **Step 3: Run the full Phase C suite**

Run: `go test ./jolt/... -v -ginkgo.focus="PhysicsSystemTest \(joltc.*"`
Expected: 19 specs PASS.

- [ ] **Step 4: Commit**

```bash
git add jolt/port_physics_system_test.go
git commit -m "test(jolt): port simulation specs (Update_FallingBody, MultipleBodies)"
```

---

## Phase D — Verification

### Task 19: Full suite green run + lint

**Files:**
- (none — verification only)

- [ ] **Step 1: Run the full test suite**

Run: `go test ./...`
Expected: PASS, no flakes.

- [ ] **Step 2: Run `go vet`**

Run: `go vet ./...`
Expected: no output, exit code 0.

- [ ] **Step 3: Verify formatting**

Run: `gofmt -l jolt/`
Expected: no output (no files need formatting).

- [ ] **Step 4: Smoke-check spec count**

Run: `go test ./jolt/... -v 2>&1 | grep -E "^(Ran|PASS|FAIL)"`
Expected: total spec count ≥ 35 (existing) + 4 (Phase A) + 36 (Phase B portable) + 19 (Phase C) ≈ 94 specs (exact figure depends on Pending placeholders).

- [ ] **Step 5: No commit needed**

If all four checks pass, the port is complete.

---

## Self-review checklist (for the implementer)

After all tasks, verify:

- [ ] Every joltc TEST_F listed in the spec table has either a corresponding `It(...)` block, a `Pending(...)` placeholder, or is documented as deferred in this plan.
- [ ] Every new `Get*` carve-out is exercised by at least one `It` block.
- [ ] Every newly added `New*ShapeSettings` constructor is exercised by at least one `*_CreateWithSettings` spec (the constructor exists *because* a spec needed it).
- [ ] No new exported function is left without a doc comment.
- [ ] `git log --oneline` shows one commit per task — no batched changes.
- [ ] The TODO(generator) comment lives near a single carve-out, not duplicated.
