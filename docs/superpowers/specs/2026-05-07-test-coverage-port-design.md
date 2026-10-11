# Design — Port joltc Tests to Go (Phases A–C)

> **Status (audit 2026-10-11): Implemented.**

**Date:** 2026-05-07
**Status:** Draft for review
**Goal:** Lift the test coverage of `jolt/` from a thin smoke suite (~35 specs) to systematic per-API coverage by porting the relevant subset of `third_party/joltc/tests/` into Ginkgo tests, fixing generator out-parameter bugs inline as carve-outs.

---

## Why

After the first stage of bindings landed, the public Go surface in `jolt/` has only minimal coverage:

- `math_test.go` — Vec3/Quat/Mat44 layout and a handful of arithmetic ops.
- `lifetime_test.go`, `verification_test.go` — lifecycle and shape construction.
- `integration_test.go`, `contact_listener_integration_test.go` — two end-to-end specs.
- `real_test.go` — `Real` size guard.

Total: ~669 LOC across ~35 specs. The generator emits ~1 095 functions in `jolt/jolt_generated.go`; the bulk of that surface is exercised only transitively through the two integration specs.

There is one specific systemic risk:

> The generator emits out-pointer accessors as **pass-by-value**. For example, `BodyInterface.Position(b BodyID, r RVec3)` writes into `r`, which is a copy on the Go stack — the caller sees nothing. Only `GetPosition` has been carved out by hand. **Every other generated `Get*`-shaped method is silently broken** and there is currently no test that would catch this.

Porting tests against the joltc C API surface — which the Go bindings wrap 1:1 — systematically exposes these and pins down the contract for every shape and physics operation we expose.

## Source corpora and choice

Two upstream test corpora are available:

| Corpus | Lines | Framework | Tests what |
|---|---|---|---|
| `third_party/JoltPhysics/UnitTests/` | ~10 K+ | doctest | Internal C++ classes — most surface is unreachable through joltc |
| `third_party/joltc/tests/` ⭐ | ~3 300 | gtest | The C API our Go bindings wrap 1:1 |

We port from **joltc/tests/** because each test maps directly to the same call our generator produces. JoltPhysics/UnitTests/ would be a porting target only if we eventually expose Jolt's C++ classes — which is not in scope.

## Scope

Three phases, scoped to the MVP-bound surface defined in `docs/superpowers/plans/2026-05-06-jolt-go-mvp.md`.

### Phase A — Core (`port_core_test.go`)

Ports `third_party/joltc/tests/test_core.cpp`.

| joltc `TEST_F` | Action | Notes |
|---|---|---|
| `CoreTest.Init_Shutdown` | Port | Documented baseline |
| `CoreTest.Init_Multiple_Times` | Port | New |
| `JobSystemTest.ThreadPool_Create_Default` | Port | Adds `NewJobSystemThreadPoolDefault()` helper for nil-config path |
| `JobSystemTest.ThreadPool_Create_WithConfig` | Port | Already covered by verification_test.go; port for completeness |
| `JobSystemTest.SetTraceHandler` | Defer | Needs Go callback carve-out (non-trivial) |
| `JobSystemTest.SetAssertFailureHandler` | Defer | Same |

**Net:** 4 new specs, 0 binding carve-outs, possibly 1 small handwritten constructor.

### Phase B — Shapes (`port_shapes_test.go`)

Ports `third_party/joltc/tests/test_shapes.cpp` (44 tests; ~40 portable).

Organisation: nested Describes — one per shape type, plus a `Shape (base)` group and a `ConvexShape` group.

| Group | Specs |
|---|---|
| Per-shape constructors and type ID | Sphere, Box, Capsule, Cylinder, Triangle, TaperedCapsule, ConvexHull, Plane, TaperedCylinder, RotatedTranslated, Scaled, OffsetCenterOfMass, StaticCompound, MutableCompound, Empty |
| Per-shape properties | Sphere `GetVolume`, Box `GetConvexRadius`, ConvexHull `GetPoint` |
| Base `Shape` | `UserData`, `GetLocalBounds`, `GetCenterOfMass`, `GetInnerRadius`, `GetMassProperties`, `MustBeStatic`, `CastRay` (hit + miss), `CollidePoint` (inside + outside), `IsValidScale`, `MakeScaleValid` |
| `ConvexShape` | `Density`, `SetDensity` |
| `ShapeSettings` | `UserData` round-trip |

The 11 shape constructors already covered in `verification_test.go` are **not** removed — that file uses `DescribeTable` for breadth; this file adds depth (properties + behaviour). They complement.

**Net:** ~40 specs, ~6–8 carve-outs surfaced.

### Phase C — PhysicsSystem & BodyInterface (`port_physics_system_test.go`)

Ports `third_party/joltc/tests/test_physics_system.cpp` (19 tests; all portable).

| Group | Specs |
|---|---|
| System | `Create_ValidSettings`, `GetBodyInterface`, `GetBodyInterfaceNoLock`, `SetGravity` (round-trip), `GetNumBodies_Initial`, `GetNumActiveBodies_Initial`, `GetMaxBodies` |
| Body lifecycle | `CreateBody_Sphere`, `CreateBody_Box` |
| Kinematics / dynamics | `GetPosition`, `SetPosition`, `SetLinearVelocity`, `AddImpulse`, `GetMotionType`, `ActivateDeactivate`, `Restitution`, `Friction` |
| Simulation | `Update_FallingBody` (asserts y-velocity, distinct from drop-a-box), `MultipleBodies` (multi-sphere pile) |

**Net:** 19 specs, ~3 carve-outs surfaced.

## Architecture

### File layout

New files added to `jolt/`:

```
jolt/
  port_helpers_test.go          — shared fixtures (system setup, body creation)
  port_core_test.go             — Phase A
  port_shapes_test.go           — Phase B
  port_physics_system_test.go   — Phase C
```

Each file opens with a header comment naming its joltc source file, so traceability stays explicit.

### Test-structure mapping

| C++ pattern | Go translation |
|---|---|
| `class FooTest : public ::testing::Test` | `var _ = Describe("Foo", func() { ... })` |
| `SetUp()` / `TearDown()` | `BeforeEach` / `AfterEach` inside Describe |
| `TEST_F(FooTest, Bar)` | `It("Bar", func() { ... })` |
| `EXPECT_EQ(a, b)` | `Expect(a).To(Equal(b))` |
| `EXPECT_NEAR(a, b, eps)` | `Expect(a).To(BeNumerically("~", b, eps))` |
| `EXPECT_FLOAT_EQ(a, b)` | `Expect(a).To(BeNumerically("~", b, 1e-5))` |
| `ASSERT_NE(p, nullptr)` | `Expect(p).NotTo(BeNil())` |

### Shared helpers (`port_helpers_test.go`)

Two helpers, consolidated from the C++ test fixtures:

- **`newTwoLayerSystem()`** — creates a 2-layer (NON_MOVING / MOVING) `PhysicsSystem` using the **Mask scheme** already wired up in our handwritten code. The joltc tests use the Table scheme; we use Mask for consistency with our existing `integration_test.go` and to avoid binding new BroadPhaseLayerInterfaceTable plumbing for tests alone. End-to-end behaviour (two layers that collide) is identical.
- **`createSphereBody(bi *BodyInterface, x, y, z, radius float32, motionType MotionType, layer ObjectLayer) BodyID`** — mirrors the C++ helper.

### Carve-out strategy for out-parameter bugs

The generator emits out-pointer accessors as pass-by-value. The fix already in tree (`GetPosition`) is a hand-written carve-out. We extend the same pattern in `jolt_handwritten.go` for every accessor a ported test exercises.

Naming: the carve-out is `GetX()`; the broken generated `X()` continues to coexist (we don't rename or delete it — that's a future generator change). A `// TODO(generator): emit Get* variants for out-pointer accessors` note is added next to the first new carve-out so the systemic fix lands in the generator backlog.

Expected new carve-outs (final list determined by the port):

- `PhysicsSystem.GetGravity() Vec3`
- `BodyInterface.GetLinearVelocity(b BodyID) Vec3`
- `BodyInterface.GetCenterOfMassPosition(b BodyID) RVec3`
- `BodyInterface.GetRotation(b BodyID) Quat`
- `BoxShape.GetHalfExtent() Vec3`
- `Shape.GetCenterOfMass() Vec3`
- `Shape.GetLocalBounds() AABox`
- `Shape.GetMassProperties() MassProperties`

Each carve-out is ~5 lines: declare local, call `cjolt.JPH_*` with `&out`, `runtime.KeepAlive(receiver)`, return.

### Conventions

- Float comparisons: `BeNumerically("~", target, 1e-5)` for `EXPECT_FLOAT_EQ`; larger eps for simulation steady-state assertions.
- Body cleanup: `defer bi.RemoveAndDestroyBody(id)` immediately after creation, mirroring the C++ helper.
- Where a port reveals an unbound joltc function, add a `Pending()` spec naming the unbound function so it surfaces in test output rather than getting silently dropped.
- Tests must run cleanly under `go test ./jolt/...` with no flakes; simulation steady-state assertions use sufficient frame counts (≥120 @ 60 Hz where the existing tests already do).

## Deliverables

- `jolt/port_helpers_test.go`, `jolt/port_core_test.go`, `jolt/port_shapes_test.go`, `jolt/port_physics_system_test.go`.
- New `Get*` carve-outs in `jolt/jolt_handwritten.go`, each accompanied by a spec that exercises it.
- One generator TODO comment marking the systemic issue.

## Out of scope

- **Phase D — Collision queries** (`test_collision.cpp`, ~6 tests): RayCast hit/miss/closest, broad/narrow phase queries. Defer.
- **Phase E — Math parity** (`test_math.cpp`, ~51 tests): targeted gaps in `math.go` — Quat slerp/lerp, Mat44 rotation/translation matrices, axis getters. Defer.
- **Constraints** (`test_constraints.cpp`, 18 tests): bindings exist but constraints are post-MVP per the original plan.
- **Character / Vehicle / Skeleton / Ragdoll** (post-MVP per plan).
- **Trace / Assert handlers**: require Go callback carve-outs.
- **Generator-side fix for out-pointer emission**: separate spec.
- **`JoltPhysics/UnitTests/`**: tests internal C++ classes, not reachable through joltc.

## Approach alternatives considered

- **Port everything, including post-MVP areas.** ~3.3 K LOC of test → roughly 2–3 K LOC of Go test. **Rejected** — exceeds MVP scope; some post-MVP bindings (vehicle, ragdoll) aren't generated yet.
- **Hand-write a thin smoke suite from scratch.** Fast but no upstream traceability. **Rejected** — the user's stated concern is exactly that minimal coverage misses bugs.
- **Phases A–C only (chosen).** ~63 specs across the MVP-bound C-API surface, fix carve-outs inline.

## Success criteria

- `go test ./jolt/...` passes with no flakes after each phase lands.
- Every generated `Get*`-shaped accessor exercised by a port has a working carve-out and a spec that asserts a non-trivial round-trip.
- A reader can map any spec back to its joltc source line without ambiguity.
