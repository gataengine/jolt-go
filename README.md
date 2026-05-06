# jolt-go

Go binding for [JoltPhysics](https://github.com/jrouwe/JoltPhysics) via the
[`amerkoleci/joltc`](https://github.com/amerkoleci/joltc) C binding.

Module: `github.com/gataengine/jolt-go`

---

## Install

```
go get github.com/gataengine/jolt-go/jolt
```

## Build prerequisites

- **C++17 toolchain** — `clang++` (macOS/Linux) or `g++` (Linux/Windows)
- **CMake** is _not_ required; the binding is built entirely by `cgo` using
  source files under `third_party/`
- **First build** compiles `~40 000 lines` of Jolt Physics C++ — expect
  **30–60 seconds** on a modern machine; subsequent builds are cached

> The binding requires the two Git submodules to be present.  After a fresh
> clone, run:
> ```
> git submodule update --init --recursive
> ```

---

## Quickstart — drop a box

```go
package main

import "github.com/gataengine/jolt-go/jolt"

func main() {
    jolt.Init()
    defer jolt.Shutdown()

    // Layer filters — two broad-phase layers: 0=NON_MOVING, 1=MOVING
    bpli := jolt.NewBroadPhaseLayerInterfaceMask(2)
    jolt.SetupDefaultBroadPhaseLayers(bpli)
    olpf := jolt.NewObjectLayerPairFilterMask()
    ovbplf := jolt.NewObjectVsBroadPhaseLayerFilterMask(bpli)

    // Physics world
    ps := jolt.NewPhysicsSystemWithMaskFilters(10240, 65536, 10240, bpli, olpf, ovbplf)
    defer ps.Close()

    // Allocators
    ta := jolt.NewTempAllocator(10 * 1024 * 1024) // 10 MiB scratch
    defer ta.Close()
    js := jolt.NewJobSystemThreadPool(2048, 8, -1) // -1 = auto thread count
    defer js.Close()

    bi := ps.BodyInterface()

    // Static floor: 100×2×100 box centred at (0, -1, 0)
    floorShape := jolt.NewBoxShape(jolt.Vec3{X: 50, Y: 1, Z: 50}, 0.05)
    bi.CreateAndAddBody(
        jolt.NewBodyCreationSettingsWithShape(
            &floorShape.Shape,
            jolt.RVec3{X: 0, Y: -1, Z: 0},
            jolt.MotionTypeStatic,
            jolt.ObjectLayerNonMoving,
        ),
        jolt.ActivationDontActivate,
    )

    // Dynamic 1 m box starting at y = 10
    boxShape := jolt.NewBoxShape(jolt.Vec3{X: 0.5, Y: 0.5, Z: 0.5}, 0.05)
    boxID := bi.CreateAndAddBody(
        jolt.NewBodyCreationSettingsWithShape(
            &boxShape.Shape,
            jolt.RVec3{X: 0, Y: 10, Z: 0},
            jolt.MotionTypeDynamic,
            jolt.ObjectLayerMoving,
        ),
        jolt.ActivationActivate,
    )

    // Simulate 2 seconds at 60 Hz
    const dt float32 = 1.0 / 60.0
    for range 120 {
        ps.Update2(dt, 1, ta, js) // Note: use Update2, not Update
    }

    pos := bi.GetPosition(boxID)
    _ = pos // pos.Y ≈ 0.5 — box resting on the floor
}
```

### Contact listener

Collision callbacks are delivered from **Jolt's worker threads**.  Protect any
shared state with a mutex:

```go
var mu sync.Mutex
var contacts int

listener := jolt.NewContactListener(jolt.ContactListener{
    OnContactAdded: func(b1, b2 *jolt.Body, _ *jolt.ContactManifold, _ *jolt.ContactSettings) {
        mu.Lock()
        contacts++
        mu.Unlock()
    },
})
ps.SetContactListener(listener)
// listener has no Close() — the GC finalizer handles cleanup.
```

---

## Build tags

| Tag | Effect |
|-----|--------|
| `joltdouble` | Switches `Real` from `float32` to `float64` and compiles Jolt with `JPH_DOUBLE_PRECISION` (double-precision world coordinates). Note: changes C struct layouts; treat as experimental. |
| `jolt_cross_platform_deterministic` | Passed through to Jolt at C++ compile time via `#define JPH_CROSS_PLATFORM_DETERMINISTIC`. Enforces deterministic simulation across platforms at a small performance cost. |

---

## Regenerating bindings

The Go + cjolt files under `jolt/` and `internal/cjolt/` are generated from
`third_party/joltc/include/joltc.h`:

```
go generate ./jolt
```

This runs `generator/cmd/gen` which parses the C header and emits
`jolt/jolt_generated.go` and `internal/cjolt/cjolt_generated.go`.  Run it
after bumping the `joltc` submodule.

---

## Module layout

```
.
├── jolt/           # Public Go API  (package jolt)
├── internal/cjolt/ # cgo shim       (package cjolt, not importable by users)
├── generator/      # Binding generator (standalone Go tool)
└── third_party/
    ├── joltc/      # amerkoleci/joltc  C header + sources (submodule)
    └── JoltPhysics/ # jrouwe/JoltPhysics C++ sources      (submodule)
```

---

## Pinned upstream

| Submodule | Commit | Notes |
|-----------|--------|-------|
| `amerkoleci/joltc` | `52d8c98df523f449eb3e01b1060a0fde052970d1` | tracks Jolt v5.5.0 |
| `jrouwe/JoltPhysics` | `23dadd0e603f1b321142d4c74df07fce85064989` | tag `v5.5.0` |

---

## License

MIT. See [LICENSE](./LICENSE).

## Acknowledgements

- **[JoltPhysics](https://github.com/jrouwe/JoltPhysics)** by Jorrit Rouwé —
  MIT licensed; the physics engine itself.
- **[joltc](https://github.com/amerkoleci/joltc)** by Amer Koleci —
  MIT licensed; the C binding layer that makes cgo integration possible.
