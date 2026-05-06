# jolt-go MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a code-generated Go binding over `amerkoleci/joltc` exposing an idiomatic, OOP-shaped API for the JoltPhysics MVP surface (lifetime management, BodyInterface, all shape types, ContactListener), proven by two end-to-end integration tests.

**Architecture:** Three layers — (1) handwritten Go math types and a hand-written codegen program in `generator/`, (2) a generated low-level cgo layer in `internal/cjolt/` that bundles `joltc` + `JoltPhysics` source via a unity-build stub, (3) a generated idiomatic public API in `jolt/`. Annotations sidecar (Go struct, not YAML) supplies the things `joltc.h` cannot express: inheritance, ref-counted types, listener vtables, settings init functions.

**Tech Stack:** Go 1.25, `modernc.org/cc/v4` (C parser), `onsi/ginkgo/v2` + `onsi/gomega` (BDD test framework), cgo with C++17, `amerkoleci/joltc` and `jrouwe/JoltPhysics` as git submodules.

**Spec:** `/Users/wizard/workspace/gamedev/JoltPhysics/Docs/superpowers/specs/2026-05-06-jolt-go-binding-design.md`.

**Deferred:** CI is intentionally out of scope. `.syso` prebuilts, custom layer filters, queries, constraints, body-lock interface, character/vehicle/ragdoll/softbody, debug renderer, and the `StateRecorder` carve-out are all post-MVP.

---

## Conventions

**TDD with codegen:** Generator code is tested by feeding a fixture C header into a parser/emitter and asserting on the emitted Go (or unity-stub) output. Public API is tested by integration tests that exercise the generated bindings end-to-end.

**Test framework:** Ginkgo v2 + Gomega. One bootstrap file per package (`*_suite_test.go`) calling `RunSpecs(t, "Suite name")`. Individual test files use `Describe` / `Context` / `It`. Match assertions: `Expect(x).To(Equal(y))`, `Expect(f).To(BeNumerically("~", target, eps))`, `Expect(err).NotTo(HaveOccurred())`.

**Commits:** Conventional Commits (`feat:`, `fix:`, `chore:`, `test:`, `docs:`). Commit at the end of every task.

**Modern Go (1.25):** Use `any`, `slices`, `maps`, `cmp.Or`, `for i := range n`, `wg.Go(fn)`, `t.Context()`, `omitzero`, `b.Loop()`, `strings.SplitSeq`. No legacy patterns.

**Generator output:** Always run `gofmt` on emitted Go and pin formatting in tests so output is byte-stable. The generator should write only when content changes; commit the regenerated files.

---

## File map (created across the plan)

| Path | Created in | Purpose |
|---|---|---|
| `go.mod`, `go.sum` | Phase 0 | Module + dependencies |
| `.gitignore`, `.gitmodules`, `LICENSE`, `README.md` | Phase 0 | Repo metadata |
| `third_party/joltc/`, `third_party/JoltPhysics/` | Phase 0 | Submodules |
| `jolt/doc.go`, `jolt/math.go`, `jolt/math_test.go`, `jolt/jolt_suite_test.go` | Phase 1 | Handwritten math + ginkgo bootstrap |
| `generator/main.go`, `generator/parser.go`, `generator/ir.go`, `generator/annotations.go` | Phases 2, 5 | Generator skeleton + parser + sidecar |
| `generator/emit_unity.go`, `generator/emit_cjolt.go`, `generator/emit_jolt.go` | Phases 3, 4, 5–7 | Three emitters |
| `generator/*_test.go`, `generator/testdata/*.h` | Phases 2–7 | Fixture-driven generator tests |
| `internal/cjolt/cgo.go`, `internal/cjolt/cjolt_generated.go`, `internal/cjolt/cjolt_unity.cpp` | Phases 3, 4, 8 | Generated low-level layer |
| `jolt/jolt_generated.go`, `jolt/jolt_handwritten.go` | Phase 8 | Generated idiomatic API + empty carve-out file |
| `jolt/integration_test.go` | Phase 11 | Drop-a-box + contact-listener integration |

---

## Phase 0 — Bootstrap

### Task 1: Module path, dependencies, and ginkgo bootstrap

**Files:**
- Modify: `go.mod`
- Create: `go.sum`
- Create: `jolt/jolt_suite_test.go`

- [ ] **Step 1: Update `go.mod`**

```go
module github.com/gataengine/jolt-go

go 1.25.0
```

- [ ] **Step 2: Add dependencies**

Run:
```bash
go get modernc.org/cc/v4@latest
go get github.com/onsi/ginkgo/v2@latest
go get github.com/onsi/gomega@latest
go mod tidy
```

Expected: `go.mod` lists three `require` lines, `go.sum` is populated.

- [ ] **Step 3: Write the failing ginkgo bootstrap**

Create `jolt/jolt_suite_test.go`:

```go
package jolt_test

import (
    "testing"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

func TestJolt(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "Jolt Suite")
}

var _ = Describe("Sanity", func() {
    It("runs", func() {
        Expect(1 + 1).To(Equal(2))
    })
})
```

- [ ] **Step 4: Run the suite — confirm it passes**

Run: `go test ./jolt/...`
Expected: `Ran 1 of 1 Specs ... PASS`. (This task verifies plumbing works; the assertion is trivial on purpose.)

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum jolt/jolt_suite_test.go
git commit -m "chore: bootstrap module, dependencies, ginkgo suite"
```

---

### Task 2: Repo metadata

**Files:**
- Create: `.gitignore`
- Create: `.gitmodules`
- Create: `LICENSE` (MIT, matching joltc)
- Create: `README.md` (skeleton)

- [ ] **Step 1: Add `.gitignore`**

```
# Build artifacts
*.o
*.so
*.dylib
*.dll
*.test
*.out

# Editor
.idea/
.vscode/
*.swp

# Go build cache
/dist/
/bin/

# OS
.DS_Store
```

- [ ] **Step 2: Add empty `.gitmodules`**

(Submodules are added in Task 3; the file is created here so Task 3 can edit it.)

```
# Submodules added in Task 3
```

- [ ] **Step 3: Add `LICENSE` (MIT)**

Standard MIT license, copyright `gataengine`.

- [ ] **Step 4: Add `README.md` skeleton**

```markdown
# jolt-go

Go binding for [JoltPhysics](https://github.com/jrouwe/JoltPhysics) via the [`amerkoleci/joltc`](https://github.com/amerkoleci/joltc) C binding.

**Status:** MVP under construction.

Module: `github.com/gataengine/jolt-go`

## License

MIT. See [LICENSE](./LICENSE).
```

- [ ] **Step 5: Commit**

```bash
git add .gitignore .gitmodules LICENSE README.md
git commit -m "chore: add repo metadata (license, gitignore, readme skeleton)"
```

---

### Task 3: Pin submodules to latest synchronized commits

**Files:**
- Modify: `.gitmodules`
- Create: `third_party/joltc/` (submodule)
- Create: `third_party/JoltPhysics/` (submodule)
- Modify: `README.md` (record the pinned commits)

This task is research + setup — not TDD. It produces commit pins the rest of the plan relies on.

- [ ] **Step 1: Look up the latest joltc release that pins a Jolt version**

```bash
# Inspect joltc README + recent commits to find the Jolt tag it tracks (e.g. v5.5.0).
git ls-remote --tags https://github.com/amerkoleci/joltc.git | tail -5
git ls-remote --tags https://github.com/jrouwe/JoltPhysics.git | tail -5
```

Pick the most recent joltc commit whose README/CMake declares a specific Jolt tag, then pick that Jolt tag commit.

- [ ] **Step 2: Add submodules pinned to those commits**

```bash
git submodule add https://github.com/amerkoleci/joltc.git third_party/joltc
git submodule add https://github.com/jrouwe/JoltPhysics.git third_party/JoltPhysics
( cd third_party/joltc && git checkout <JOLTC_COMMIT> )
( cd third_party/JoltPhysics && git checkout <JOLT_COMMIT> )
```

- [ ] **Step 3: Verify the joltc header exists**

Run: `test -f third_party/joltc/include/joltc.h && echo OK`
Expected: `OK`.

- [ ] **Step 4: Record the pins in README**

Append to `README.md`:

```markdown
## Pinned upstream

- `amerkoleci/joltc` @ `<JOLTC_COMMIT>` (tracks Jolt `v<X.Y.Z>`)
- `jrouwe/JoltPhysics` @ `<JOLT_COMMIT>` (tag `v<X.Y.Z>`)

To regenerate bindings after bumping submodules: `go generate ./jolt`.
```

- [ ] **Step 5: Commit**

```bash
git add .gitmodules third_party README.md
git commit -m "chore: pin joltc + JoltPhysics submodules to synchronized v<X.Y.Z>"
```

---

## Phase 1 — Math types (handwritten, pure Go)

These are pure Go and exist before any cgo work. Layouts are designed to match `JPH_Vec3` / `JPH_Quat` / `JPH_Mat4` so later cgo wrappers can `unsafe.Pointer`-cast for free.

### Task 4: Package doc + Real type alias under build tag

**Files:**
- Create: `jolt/doc.go`
- Create: `jolt/real_single.go`
- Create: `jolt/real_double.go`
- Create: `jolt/real_test.go`

- [ ] **Step 1: Write the failing test**

`jolt/real_test.go`:

```go
package jolt_test

import (
    "unsafe"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"

    "github.com/gataengine/jolt-go/jolt"
)

var _ = Describe("Real", func() {
    It("is 4 bytes by default", func() {
        var r jolt.Real
        Expect(unsafe.Sizeof(r)).To(Equal(uintptr(4)))
    })
})
```

- [ ] **Step 2: Run, expect failure**

Run: `go test ./jolt/... -run TestJolt`
Expected: build error — `jolt.Real` undefined.

- [ ] **Step 3: Create the source files**

`jolt/doc.go`:

```go
// Package jolt is an idiomatic Go binding for JoltPhysics.
//
// Thread safety: physics callbacks (e.g. ContactListener) are invoked from
// Jolt's worker threads. Callers are responsible for synchronizing access to
// state shared with their listener implementations. The binding does not
// serialize callbacks.
package jolt
```

`jolt/real_single.go`:

```go
//go:build !joltdouble

package jolt

// Real is the scalar type used by world-space coordinates.
//
// Defaults to float32. Build with -tags joltdouble to switch to float64,
// which also enables JPH_DOUBLE_PRECISION in the cgo layer.
type Real = float32
```

`jolt/real_double.go`:

```go
//go:build joltdouble

package jolt

type Real = float64
```

- [ ] **Step 4: Run, expect pass; also try double build**

Run: `go test ./jolt/...`
Expected: PASS.

Run: `go build -tags joltdouble ./jolt/...`
Expected: build succeeds.

- [ ] **Step 5: Commit**

```bash
git add jolt/doc.go jolt/real_single.go jolt/real_double.go jolt/real_test.go
git commit -m "feat(math): add Real type alias gated by joltdouble build tag"
```

---

### Task 5: Vec3 type, zero value, layout

**Files:**
- Create: `jolt/math.go`
- Modify: `jolt/math_test.go` (new file)

- [ ] **Step 1: Write the failing test**

`jolt/math_test.go`:

```go
package jolt_test

import (
    "unsafe"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"

    "github.com/gataengine/jolt-go/jolt"
)

var _ = Describe("Vec3", func() {
    It("has float32 layout matching JPH_Vec3", func() {
        var v jolt.Vec3
        Expect(unsafe.Sizeof(v)).To(Equal(uintptr(12)))
        Expect(unsafe.Offsetof(v.Y)).To(Equal(uintptr(4)))
        Expect(unsafe.Offsetof(v.Z)).To(Equal(uintptr(8)))
    })

    It("has zero value (0,0,0)", func() {
        var v jolt.Vec3
        Expect(v).To(Equal(jolt.Vec3{}))
    })
})
```

- [ ] **Step 2: Run, expect failure (build error: jolt.Vec3 undefined)**

Run: `go test ./jolt/...`
Expected: build error.

- [ ] **Step 3: Implement minimally**

`jolt/math.go`:

```go
package jolt

// Vec3 is a 3-component single-precision vector. Layout matches JPH_Vec3.
type Vec3 struct{ X, Y, Z float32 }
```

- [ ] **Step 4: Run, expect pass**

Run: `go test ./jolt/...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add jolt/math.go jolt/math_test.go
git commit -m "feat(math): add Vec3 type with JPH_Vec3-compatible layout"
```

---

### Task 6: Vec3 arithmetic — Add, Sub, Scale, Neg

**Files:**
- Modify: `jolt/math.go`
- Modify: `jolt/math_test.go`

- [ ] **Step 1: Write the failing test**

Append to `jolt/math_test.go`:

```go
var _ = Describe("Vec3 arithmetic", func() {
    a := jolt.Vec3{X: 1, Y: 2, Z: 3}
    b := jolt.Vec3{X: 4, Y: 5, Z: 6}

    It("Add", func() {
        Expect(a.Add(b)).To(Equal(jolt.Vec3{X: 5, Y: 7, Z: 9}))
    })
    It("Sub", func() {
        Expect(b.Sub(a)).To(Equal(jolt.Vec3{X: 3, Y: 3, Z: 3}))
    })
    It("Scale", func() {
        Expect(a.Scale(2)).To(Equal(jolt.Vec3{X: 2, Y: 4, Z: 6}))
    })
    It("Neg", func() {
        Expect(a.Neg()).To(Equal(jolt.Vec3{X: -1, Y: -2, Z: -3}))
    })
})
```

- [ ] **Step 2: Run, expect failure**

Run: `go test ./jolt/...`
Expected: methods undefined.

- [ ] **Step 3: Implement**

Append to `jolt/math.go`:

```go
func (v Vec3) Add(o Vec3) Vec3   { return Vec3{v.X + o.X, v.Y + o.Y, v.Z + o.Z} }
func (v Vec3) Sub(o Vec3) Vec3   { return Vec3{v.X - o.X, v.Y - o.Y, v.Z - o.Z} }
func (v Vec3) Scale(s float32) Vec3 { return Vec3{v.X * s, v.Y * s, v.Z * s} }
func (v Vec3) Neg() Vec3         { return Vec3{-v.X, -v.Y, -v.Z} }
```

- [ ] **Step 4: Run, expect pass**

Run: `go test ./jolt/...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add jolt/math.go jolt/math_test.go
git commit -m "feat(math): add Vec3 arithmetic (Add/Sub/Scale/Neg)"
```

---

### Task 7: Vec3 Dot / Cross / Length / Normalized

**Files:**
- Modify: `jolt/math.go`
- Modify: `jolt/math_test.go`

- [ ] **Step 1: Write the failing test**

Append:

```go
var _ = Describe("Vec3 geometry", func() {
    It("Dot", func() {
        Expect(jolt.Vec3{1, 2, 3}.Dot(jolt.Vec3{4, -5, 6})).
            To(BeNumerically("~", 12, 1e-6))
    })
    It("Cross", func() {
        c := jolt.Vec3{1, 0, 0}.Cross(jolt.Vec3{0, 1, 0})
        Expect(c).To(Equal(jolt.Vec3{0, 0, 1}))
    })
    It("Length and Normalized", func() {
        v := jolt.Vec3{0, 3, 4}
        Expect(v.Length()).To(BeNumerically("~", 5, 1e-6))
        n := v.Normalized()
        Expect(n.Length()).To(BeNumerically("~", 1, 1e-6))
    })
})
```

- [ ] **Step 2: Run, expect failure**

Run: `go test ./jolt/...`
Expected: undefined methods.

- [ ] **Step 3: Implement**

Append to `jolt/math.go`:

```go
import "math"

func (v Vec3) Dot(o Vec3) float32 { return v.X*o.X + v.Y*o.Y + v.Z*o.Z }

func (v Vec3) Cross(o Vec3) Vec3 {
    return Vec3{
        X: v.Y*o.Z - v.Z*o.Y,
        Y: v.Z*o.X - v.X*o.Z,
        Z: v.X*o.Y - v.Y*o.X,
    }
}

func (v Vec3) LengthSq() float32 { return v.Dot(v) }
func (v Vec3) Length() float32   { return float32(math.Sqrt(float64(v.LengthSq()))) }

func (v Vec3) Normalized() Vec3 {
    l := v.Length()
    if l == 0 {
        return Vec3{}
    }
    return v.Scale(1 / l)
}
```

(Add the `import "math"` at file top — replace bare `package jolt` line if needed.)

- [ ] **Step 4: Run, expect pass**

Run: `go test ./jolt/...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add jolt/math.go jolt/math_test.go
git commit -m "feat(math): add Vec3 geometry (Dot/Cross/Length/Normalized)"
```

---

### Task 8: Quat — type, identity, Mul, Inverse, Conjugate, RotateVec3

**Files:**
- Modify: `jolt/math.go`
- Modify: `jolt/math_test.go`

- [ ] **Step 1: Write the failing test**

```go
var _ = Describe("Quat", func() {
    It("Identity multiplied is identity", func() {
        q := jolt.QuatIdentity()
        Expect(q).To(Equal(jolt.Quat{W: 1}))
        Expect(q.Mul(q)).To(Equal(q))
    })
    It("Inverse of unit quat equals Conjugate", func() {
        q := jolt.Quat{X: 0, Y: 0, Z: 0.7071068, W: 0.7071068}
        inv := q.Inverse()
        Expect(inv.X).To(BeNumerically("~", q.Conjugate().X, 1e-5))
        Expect(inv.Y).To(BeNumerically("~", q.Conjugate().Y, 1e-5))
        Expect(inv.Z).To(BeNumerically("~", q.Conjugate().Z, 1e-5))
        Expect(inv.W).To(BeNumerically("~", q.Conjugate().W, 1e-5))
    })
    It("RotateVec3 of identity is identity", func() {
        v := jolt.Vec3{1, 2, 3}
        Expect(jolt.QuatIdentity().RotateVec3(v)).To(Equal(v))
    })
})
```

- [ ] **Step 2: Run, expect failure**

- [ ] **Step 3: Implement**

```go
// Quat is a single-precision unit quaternion. Layout matches JPH_Quat.
type Quat struct{ X, Y, Z, W float32 }

func QuatIdentity() Quat { return Quat{W: 1} }

func (q Quat) Mul(o Quat) Quat {
    return Quat{
        X: q.W*o.X + q.X*o.W + q.Y*o.Z - q.Z*o.Y,
        Y: q.W*o.Y - q.X*o.Z + q.Y*o.W + q.Z*o.X,
        Z: q.W*o.Z + q.X*o.Y - q.Y*o.X + q.Z*o.W,
        W: q.W*o.W - q.X*o.X - q.Y*o.Y - q.Z*o.Z,
    }
}

func (q Quat) Conjugate() Quat { return Quat{-q.X, -q.Y, -q.Z, q.W} }

func (q Quat) Inverse() Quat {
    n := q.X*q.X + q.Y*q.Y + q.Z*q.Z + q.W*q.W
    if n == 0 {
        return Quat{}
    }
    inv := 1 / n
    c := q.Conjugate()
    return Quat{c.X * inv, c.Y * inv, c.Z * inv, c.W * inv}
}

func (q Quat) RotateVec3(v Vec3) Vec3 {
    // v' = q * (0,v) * q^-1, simplified for unit quats
    u := Vec3{q.X, q.Y, q.Z}
    s := q.W
    return u.Scale(2 * u.Dot(v)).Add(
        v.Scale(s*s - u.Dot(u)).Add(
            u.Cross(v).Scale(2 * s),
        ),
    )
}
```

- [ ] **Step 4: Run, expect pass**

- [ ] **Step 5: Commit**

```bash
git add jolt/math.go jolt/math_test.go
git commit -m "feat(math): add Quat with Mul/Inverse/Conjugate/RotateVec3"
```

---

### Task 9: Mat44 — type, identity, Mul, Transform

**Files:**
- Modify: `jolt/math.go`
- Modify: `jolt/math_test.go`

- [ ] **Step 1: Write the failing test**

```go
var _ = Describe("Mat44", func() {
    It("Identity layout is column-major and 64 bytes", func() {
        var m jolt.Mat44
        Expect(unsafe.Sizeof(m)).To(Equal(uintptr(64)))
        id := jolt.Mat44Identity()
        Expect(id.M[0]).To(Equal(float32(1)))
        Expect(id.M[5]).To(Equal(float32(1)))
        Expect(id.M[10]).To(Equal(float32(1)))
        Expect(id.M[15]).To(Equal(float32(1)))
    })
    It("Mul(Identity, X) = X", func() {
        x := jolt.Mat44Identity()
        x.M[12] = 7 // tx
        Expect(jolt.Mat44Identity().Mul(x)).To(Equal(x))
    })
    It("Transform translates", func() {
        m := jolt.Mat44Identity()
        m.M[12], m.M[13], m.M[14] = 1, 2, 3
        Expect(m.Transform(jolt.Vec3{0, 0, 0})).To(Equal(jolt.Vec3{1, 2, 3}))
    })
})
```

- [ ] **Step 2: Run, expect failure**

- [ ] **Step 3: Implement**

```go
// Mat44 is a column-major 4x4 single-precision matrix.
// Layout matches JPH_Mat4 (16 floats, columns first).
type Mat44 struct{ M [16]float32 }

func Mat44Identity() Mat44 {
    return Mat44{M: [16]float32{
        1, 0, 0, 0,
        0, 1, 0, 0,
        0, 0, 1, 0,
        0, 0, 0, 1,
    }}
}

func (a Mat44) Mul(b Mat44) Mat44 {
    var r Mat44
    for col := range 4 {
        for row := range 4 {
            var s float32
            for k := range 4 {
                s += a.M[k*4+row] * b.M[col*4+k]
            }
            r.M[col*4+row] = s
        }
    }
    return r
}

// Transform applies the matrix to v as a point (w=1).
func (m Mat44) Transform(v Vec3) Vec3 {
    return Vec3{
        X: m.M[0]*v.X + m.M[4]*v.Y + m.M[8]*v.Z + m.M[12],
        Y: m.M[1]*v.X + m.M[5]*v.Y + m.M[9]*v.Z + m.M[13],
        Z: m.M[2]*v.X + m.M[6]*v.Y + m.M[10]*v.Z + m.M[14],
    }
}
```

- [ ] **Step 4: Run, expect pass**

- [ ] **Step 5: Commit**

```bash
git add jolt/math.go jolt/math_test.go
git commit -m "feat(math): add Mat44 column-major with Mul/Transform"
```

---

### Task 10: RVec3 + RMat44 (build-tag-aware)

**Files:**
- Modify: `jolt/math.go`
- Modify: `jolt/math_test.go`

- [ ] **Step 1: Write the failing test**

```go
var _ = Describe("RVec3 / RMat44", func() {
    It("RVec3 size matches Real * 3", func() {
        var r jolt.RVec3
        var s jolt.Real
        Expect(unsafe.Sizeof(r)).To(Equal(unsafe.Sizeof(s) * 3))
    })
    It("RMat44 size is 16 * sizeof(Real)", func() {
        var r jolt.RMat44
        var s jolt.Real
        Expect(unsafe.Sizeof(r)).To(Equal(unsafe.Sizeof(s) * 16))
    })
})
```

- [ ] **Step 2: Run, expect failure**

- [ ] **Step 3: Implement**

```go
// RVec3 is a Real-precision 3-component vector. Layout matches JPH_RVec3
// (which is JPH_Vec3 in single precision, JPH_DVec3 in double precision).
type RVec3 struct{ X, Y, Z Real }

// RMat44 is a Real-precision column-major 4x4 matrix. Layout matches JPH_RMat4.
type RMat44 struct{ M [16]Real }
```

- [ ] **Step 4: Run, expect pass (single + double)**

```bash
go test ./jolt/...
go test -tags joltdouble ./jolt/...
```

Expected: both PASS.

- [ ] **Step 5: Commit**

```bash
git add jolt/math.go jolt/math_test.go
git commit -m "feat(math): add RVec3 and RMat44 with build-tag-aware Real"
```

---

## Phase 2 — Generator parser

The parser turns `joltc.h` into an in-memory IR. Tested with synthetic fixture headers in `generator/testdata/` so we don't depend on submodules.

### Task 11: Generator skeleton (`main.go`) + IR types

**Files:**
- Create: `generator/main.go`
- Create: `generator/ir.go`
- Create: `generator/ir_test.go`
- Create: `generator/generator_suite_test.go`

- [ ] **Step 1: Write the failing test**

`generator/generator_suite_test.go`:

```go
package generator_test

import (
    "testing"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

func TestGenerator(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "Generator Suite")
}
```

`generator/ir_test.go`:

```go
package generator_test

import (
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"

    "github.com/gataengine/jolt-go/generator"
)

var _ = Describe("IR", func() {
    It("Func has Name and Params", func() {
        f := generator.Func{Name: "JPH_Body_GetID", Params: []generator.Param{{Name: "self", CType: "JPH_Body*"}}}
        Expect(f.Name).To(Equal("JPH_Body_GetID"))
        Expect(f.Params).To(HaveLen(1))
    })
})
```

- [ ] **Step 2: Run, expect failure**

Run: `go test ./generator/...`
Expected: build error — package `generator` doesn't exist yet.

- [ ] **Step 3: Implement**

`generator/ir.go`:

```go
// Package generator parses joltc.h and emits the cgo + idiomatic Go layers.
package generator

// Param is a function parameter or field descriptor.
type Param struct {
    Name  string
    CType string // raw C type as it appears in the header
}

// Func is a C function declaration.
type Func struct {
    Name   string
    Return string
    Params []Param
}

// Struct is a C struct (or opaque tag).
type Struct struct {
    Name   string
    Opaque bool
    Fields []Param
}

// Enum is a C enum.
type Enum struct {
    Name    string
    Members []EnumMember
}

type EnumMember struct {
    Name  string
    Value int64
}

// Header is the parsed contents of joltc.h.
type Header struct {
    Funcs    []Func
    Structs  []Struct
    Enums    []Enum
    Typedefs map[string]string // alias -> underlying
}
```

`generator/main.go`:

```go
package main

import (
    "log"
    "os"
)

func main() {
    if len(os.Args) < 2 {
        log.Fatal("usage: generator <path/to/joltc.h>")
    }
    log.Printf("generator scaffold; input=%s (parsing wired in next tasks)", os.Args[1])
}
```

(Note: `main.go` is `package main` while the IR is `package generator`. The exec lives at `generator/cmd/gen/main.go` if we want a clean split, but for the MVP keeping main.go in the same dir is fine — adjust import path inside main.go to `import gen "github.com/gataengine/jolt-go/generator"` and rename `generator/main.go` → `generator/cmd/gen/main.go` with `package main`.)

**Resolve directory structure now:** create `generator/cmd/gen/main.go` with `package main` and put library code in `generator/` with `package generator`.

`generator/cmd/gen/main.go`:

```go
package main

import (
    "log"
    "os"
)

func main() {
    if len(os.Args) < 2 {
        log.Fatal("usage: gen <path/to/joltc.h>")
    }
    log.Printf("generator scaffold; input=%s (parsing wired in next tasks)", os.Args[1])
}
```

`generator/ir.go` and tests stay in `package generator`.

- [ ] **Step 4: Run, expect pass**

Run:
```bash
go test ./generator/...
go build ./generator/cmd/gen
```
Expected: tests PASS, binary builds.

- [ ] **Step 5: Commit**

```bash
git add generator/
git commit -m "feat(generator): add skeleton, IR types, ginkgo suite"
```

---

### Task 12: Parser — preprocess header, extract function decls

**Files:**
- Create: `generator/parser.go`
- Create: `generator/parser_test.go`
- Create: `generator/testdata/funcs.h`

- [ ] **Step 1: Write the failing test**

`generator/testdata/funcs.h`:

```c
#ifndef FUNCS_H
#define FUNCS_H

#include <stdbool.h>
#include <stdint.h>

typedef struct JPH_Body JPH_Body;
typedef struct JPH_Vec3 { float x, y, z; } JPH_Vec3;

void JPH_RegisterTypes(void);
void JPH_UnregisterTypes(void);
JPH_Vec3 JPH_Body_GetPosition(const JPH_Body* self);
void JPH_Body_SetLinearVelocity(JPH_Body* self, const JPH_Vec3* v);

#endif
```

`generator/parser_test.go`:

```go
package generator_test

import (
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"

    "github.com/gataengine/jolt-go/generator"
)

var _ = Describe("Parser", func() {
    Describe("Funcs", func() {
        var h *generator.Header
        BeforeEach(func() {
            var err error
            h, err = generator.ParseHeader("testdata/funcs.h")
            Expect(err).NotTo(HaveOccurred())
        })

        It("extracts free functions", func() {
            names := funcNames(h.Funcs)
            Expect(names).To(ContainElements("JPH_RegisterTypes", "JPH_UnregisterTypes"))
        })

        It("extracts methods with self pointer", func() {
            f := findFunc(h.Funcs, "JPH_Body_GetPosition")
            Expect(f.Return).To(Equal("JPH_Vec3"))
            Expect(f.Params).To(HaveLen(1))
            Expect(f.Params[0].CType).To(Equal("const JPH_Body *"))
        })

        It("extracts methods with multiple params", func() {
            f := findFunc(h.Funcs, "JPH_Body_SetLinearVelocity")
            Expect(f.Params).To(HaveLen(2))
            Expect(f.Params[1].CType).To(Equal("const JPH_Vec3 *"))
        })
    })
})

func funcNames(fs []generator.Func) []string {
    names := make([]string, len(fs))
    for i, f := range fs {
        names[i] = f.Name
    }
    return names
}

func findFunc(fs []generator.Func, name string) generator.Func {
    for _, f := range fs {
        if f.Name == name {
            return f
        }
    }
    Fail("func not found: " + name)
    return generator.Func{}
}
```

- [ ] **Step 2: Run, expect failure**

Run: `go test ./generator/...`
Expected: `ParseHeader` undefined.

- [ ] **Step 3: Implement parser**

`generator/parser.go`:

```go
package generator

import (
    "fmt"
    "os"

    "modernc.org/cc/v4"
)

// ParseHeader pre-processes the file with cc/v4 default config and extracts
// the IR.
func ParseHeader(path string) (*Header, error) {
    cfg, err := cc.NewConfig(cc.HostABI())
    if err != nil {
        return nil, fmt.Errorf("cc config: %w", err)
    }
    src, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    ast, err := cc.Parse(cfg, []cc.Source{{Name: path, Value: string(src)}})
    if err != nil {
        return nil, fmt.Errorf("parse: %w", err)
    }
    h := &Header{Typedefs: map[string]string{}}
    for n := ast.TranslationUnit; n != nil; n = n.TranslationUnit {
        if d := n.ExternalDeclaration.Declaration; d != nil {
            walkDeclaration(h, d)
        }
    }
    return h, nil
}

// walkDeclaration extracts func/struct/enum/typedef from a top-level decl.
func walkDeclaration(h *Header, d *cc.Declaration) {
    if d.InitDeclaratorList == nil {
        return
    }
    for idl := d.InitDeclaratorList; idl != nil; idl = idl.InitDeclaratorList {
        decl := idl.InitDeclarator.Declarator
        if decl == nil {
            continue
        }
        t := decl.Type()
        switch t.Kind() {
        case cc.Function:
            ft := t.(*cc.FunctionType)
            f := Func{
                Name:   decl.Name(),
                Return: ft.Result().String(),
            }
            for i := 0; i < ft.Parameters().Len(); i++ {
                p := ft.Parameters().At(i)
                f.Params = append(f.Params, Param{
                    Name:  p.Name(),
                    CType: p.Type().String(),
                })
            }
            h.Funcs = append(h.Funcs, f)
        }
        // Struct/enum/typedef cases are added in Task 13.
    }
}
```

Note: the exact cc/v4 API names (`InitDeclaratorList`, `Type().Kind()`, `FunctionType`) may differ — `go doc modernc.org/cc/v4` from the repo to confirm names before writing. The structural shape (walk decls → switch on Kind → extract Name/Return/Params) is what matters; refine to whatever names the installed version uses.

- [ ] **Step 4: Run, expect pass**

Run: `go test ./generator/...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add generator/parser.go generator/parser_test.go generator/testdata/funcs.h
git commit -m "feat(generator): parse C function declarations from joltc.h-style header"
```

---

### Task 13: Parser — struct, enum, typedef extraction

**Files:**
- Modify: `generator/parser.go`
- Modify: `generator/parser_test.go`
- Create: `generator/testdata/types.h`

- [ ] **Step 1: Write the failing test**

`generator/testdata/types.h`:

```c
#ifndef TYPES_H
#define TYPES_H

#include <stdint.h>

typedef struct JPH_Shape JPH_Shape;          // opaque
typedef struct JPH_BodyID { uint32_t id; } JPH_BodyID;

typedef enum JPH_MotionType {
    JPH_MotionType_Static = 0,
    JPH_MotionType_Kinematic = 1,
    JPH_MotionType_Dynamic = 2,
} JPH_MotionType;

typedef uint32_t JPH_ObjectLayer;

#endif
```

Append tests:

```go
Describe("Structs/enums/typedefs", func() {
    var h *generator.Header
    BeforeEach(func() {
        var err error
        h, err = generator.ParseHeader("testdata/types.h")
        Expect(err).NotTo(HaveOccurred())
    })

    It("recognizes opaque struct", func() {
        s := findStruct(h.Structs, "JPH_Shape")
        Expect(s.Opaque).To(BeTrue())
    })
    It("extracts struct fields", func() {
        s := findStruct(h.Structs, "JPH_BodyID")
        Expect(s.Fields).To(HaveLen(1))
        Expect(s.Fields[0].Name).To(Equal("id"))
    })
    It("extracts enum members", func() {
        e := findEnum(h.Enums, "JPH_MotionType")
        Expect(e.Members).To(HaveLen(3))
        Expect(e.Members[2]).To(Equal(generator.EnumMember{Name: "JPH_MotionType_Dynamic", Value: 2}))
    })
    It("extracts simple typedefs", func() {
        Expect(h.Typedefs["JPH_ObjectLayer"]).To(Equal("uint32_t"))
    })
})

func findStruct(ss []generator.Struct, name string) generator.Struct { /* … */ }
func findEnum(es []generator.Enum, name string) generator.Enum         { /* … */ }
```

- [ ] **Step 2: Run, expect failure**

- [ ] **Step 3: Extend `walkDeclaration` and add helpers**

Walk struct definitions, enum specifiers, and typedef declarations; populate `h.Structs`, `h.Enums`, `h.Typedefs`. An opaque struct is one with no field list at first sight. (See cc/v4 docs for `StructOrUnionSpecifier`.)

- [ ] **Step 4: Run, expect pass**

- [ ] **Step 5: Commit**

```bash
git add generator/parser.go generator/parser_test.go generator/testdata/types.h
git commit -m "feat(generator): parse structs, enums, typedefs"
```

---

### Task 14: IR — method-pattern recognition (`JPH_<Type>_<Method>`)

**Files:**
- Modify: `generator/ir.go`
- Modify: `generator/ir_test.go`

- [ ] **Step 1: Write the failing test**

```go
Describe("MethodPattern", func() {
    It("splits JPH_Body_GetPosition", func() {
        recv, method, ok := generator.SplitMethod("JPH_Body_GetPosition")
        Expect(ok).To(BeTrue())
        Expect(recv).To(Equal("Body"))
        Expect(method).To(Equal("GetPosition"))
    })
    It("returns ok=false for free functions", func() {
        _, _, ok := generator.SplitMethod("JPH_RegisterTypes")
        Expect(ok).To(BeFalse())
    })
    It("respects multi-segment receivers (JPH_BodyInterface_CreateBody)", func() {
        // JPH_BodyInterface_CreateBody → recv="BodyInterface", method="CreateBody"
        recv, method, _ := generator.SplitMethod("JPH_BodyInterface_CreateBody")
        Expect(recv).To(Equal("BodyInterface"))
        Expect(method).To(Equal("CreateBody"))
    })
})
```

- [ ] **Step 2: Run, expect failure**

- [ ] **Step 3: Implement**

Add to `ir.go`:

```go
// SplitMethod splits a JPH_<Receiver>_<Method> name. Returns ok=false if the
// name doesn't match (free function, e.g. JPH_RegisterTypes).
//
// Disambiguation rule for multi-segment receivers like BodyInterface: the
// receiver is determined by matching against KnownTypes (set elsewhere by
// the caller). For Phase 2 we use a simple heuristic — recv is the part
// up to the last underscore — and refine in Phase 3 once we have a type set.
func SplitMethod(name string) (recv, method string, ok bool) {
    const prefix = "JPH_"
    if len(name) < len(prefix) || name[:len(prefix)] != prefix {
        return "", "", false
    }
    rest := name[len(prefix):]
    // Find last underscore; everything before is the receiver, everything
    // after is the method.
    for i := len(rest) - 1; i >= 0; i-- {
        if rest[i] == '_' {
            return rest[:i], rest[i+1:], true
        }
    }
    return "", "", false
}
```

(Refinement using `KnownTypes` happens later if multi-segment ambiguity bites; for the MVP, last-underscore is sufficient because joltc consistently uses `<Type>_<MethodName>` with method names always being CamelCase single tokens.)

- [ ] **Step 4: Run, expect pass**

- [ ] **Step 5: Commit**

```bash
git add generator/ir.go generator/ir_test.go
git commit -m "feat(generator): SplitMethod recognizes JPH_<Type>_<Method> pattern"
```

---

## Phase 3 — Unity-build stub + cgo directives

### Task 15: Extract Jolt source-file list from `Jolt.cmake`

**Files:**
- Create: `generator/cmake.go`
- Create: `generator/cmake_test.go`
- Create: `generator/testdata/Jolt.cmake.fragment`

- [ ] **Step 1: Write the failing test**

`generator/testdata/Jolt.cmake.fragment`:

```cmake
set(JOLT_PHYSICS_SRC_FILES
    ${JOLT_PHYSICS_ROOT}/Core/Color.cpp
    ${JOLT_PHYSICS_ROOT}/Core/Factory.cpp
    ${JOLT_PHYSICS_ROOT}/Core/IssueReporting.cpp
    ${JOLT_PHYSICS_ROOT}/Geometry/ConvexHullBuilder.cpp
    ${JOLT_PHYSICS_ROOT}/Geometry/EPAPenetrationDepth.cpp
)
```

`generator/cmake_test.go`:

```go
var _ = Describe("ExtractJoltSources", func() {
    It("extracts .cpp paths from JOLT_PHYSICS_SRC_FILES", func() {
        paths, err := generator.ExtractJoltSources("testdata/Jolt.cmake.fragment")
        Expect(err).NotTo(HaveOccurred())
        Expect(paths).To(ContainElements(
            "Core/Color.cpp",
            "Core/Factory.cpp",
            "Geometry/ConvexHullBuilder.cpp",
        ))
        Expect(paths).To(HaveLen(5))
    })
})
```

- [ ] **Step 2: Run, expect failure**

- [ ] **Step 3: Implement**

`generator/cmake.go`:

```go
package generator

import (
    "bufio"
    "os"
    "regexp"
    "strings"
)

// ExtractJoltSources parses a CMake file (or fragment) and returns paths
// listed in the JOLT_PHYSICS_SRC_FILES variable, stripped of the
// ${JOLT_PHYSICS_ROOT}/ prefix.
func ExtractJoltSources(cmakePath string) ([]string, error) {
    f, err := os.Open(cmakePath)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    var (
        inBlock bool
        out     []string
        rx      = regexp.MustCompile(`\$\{JOLT_PHYSICS_ROOT\}/(\S+\.cpp)`)
    )
    sc := bufio.NewScanner(f)
    for sc.Scan() {
        line := sc.Text()
        switch {
        case strings.Contains(line, "set(JOLT_PHYSICS_SRC_FILES"):
            inBlock = true
        case inBlock && strings.Contains(line, ")"):
            inBlock = false
        case inBlock:
            if m := rx.FindStringSubmatch(line); m != nil {
                out = append(out, m[1])
            }
        }
    }
    return out, sc.Err()
}
```

- [ ] **Step 4: Run, expect pass**

- [ ] **Step 5: Commit**

```bash
git add generator/cmake.go generator/cmake_test.go generator/testdata/Jolt.cmake.fragment
git commit -m "feat(generator): extract Jolt source files from Jolt.cmake"
```

---

### Task 16: Emit `cjolt_unity.cpp`

**Files:**
- Create: `generator/emit_unity.go`
- Create: `generator/emit_unity_test.go`

- [ ] **Step 1: Write the failing test**

```go
var _ = Describe("EmitUnity", func() {
    It("emits a .cpp with #include lines for each source", func() {
        sources := []string{"Core/Color.cpp", "Geometry/ConvexHullBuilder.cpp"}
        joltcSources := []string{"src/joltc.cpp"}
        out := generator.EmitUnity(sources, joltcSources)
        Expect(out).To(ContainSubstring(`#include "../../third_party/JoltPhysics/Jolt/Core/Color.cpp"`))
        Expect(out).To(ContainSubstring(`#include "../../third_party/joltc/src/joltc.cpp"`))
        Expect(out).To(HavePrefix("// Code generated by jolt-go generator. DO NOT EDIT.\n"))
    })
})
```

- [ ] **Step 2: Run, expect failure**

- [ ] **Step 3: Implement**

```go
package generator

import "strings"

// EmitUnity returns the unity-build .cpp content that compiles all of Jolt
// + joltc as one translation unit per platform.
func EmitUnity(joltSources, joltcSources []string) string {
    var b strings.Builder
    b.WriteString("// Code generated by jolt-go generator. DO NOT EDIT.\n")
    b.WriteString("// Build aggregator: includes all Jolt and joltc translation units.\n\n")
    for _, p := range joltSources {
        b.WriteString(`#include "../../third_party/JoltPhysics/Jolt/` + p + `"` + "\n")
    }
    for _, p := range joltcSources {
        b.WriteString(`#include "../../third_party/joltc/` + p + `"` + "\n")
    }
    return b.String()
}
```

- [ ] **Step 4: Run, expect pass**

- [ ] **Step 5: Commit**

```bash
git add generator/emit_unity.go generator/emit_unity_test.go
git commit -m "feat(generator): emit unity-build C++ stub for Jolt + joltc"
```

---

### Task 17: Emit `cgo.go` with cgo directives

**Files:**
- Create: `generator/emit_cgo_directives.go`
- Create: `generator/emit_cgo_directives_test.go`

- [ ] **Step 1: Write the failing test**

```go
var _ = Describe("EmitCgoDirectives", func() {
    It("emits CXXFLAGS and per-arch SIMD flags", func() {
        out := generator.EmitCgoDirectives()
        Expect(out).To(ContainSubstring(`#cgo CXXFLAGS: -std=c++17`))
        Expect(out).To(ContainSubstring(`#cgo CXXFLAGS: -I${SRCDIR}/../../third_party/joltc/include`))
        Expect(out).To(ContainSubstring(`#cgo CXXFLAGS: -I${SRCDIR}/../../third_party/JoltPhysics`))
        Expect(out).To(ContainSubstring(`#cgo amd64 CXXFLAGS: -DJPH_USE_SSE4_1 -DJPH_USE_SSE4_2 -DJPH_USE_AVX`))
        Expect(out).To(ContainSubstring(`#cgo arm64 CXXFLAGS: -DJPH_USE_NEON`))
        Expect(out).To(ContainSubstring(`#cgo joltdouble CXXFLAGS: -DJPH_DOUBLE_PRECISION`))
        Expect(out).To(ContainSubstring(`#cgo LDFLAGS: -lstdc++ -lm`))
        Expect(out).To(ContainSubstring(`package cjolt`))
    })
})
```

- [ ] **Step 2: Run, expect failure**

- [ ] **Step 3: Implement**

```go
package generator

// EmitCgoDirectives returns the contents of internal/cjolt/cgo.go — the file
// that declares cgo build flags. The file imports "C" so cgo treats the
// directives in its leading comment as CXXFLAGS / LDFLAGS for the package.
func EmitCgoDirectives() string {
    return `// Code generated by jolt-go generator. DO NOT EDIT.

package cjolt

/*
#cgo CXXFLAGS: -std=c++17 -O2
#cgo CXXFLAGS: -I${SRCDIR}/../../third_party/joltc/include
#cgo CXXFLAGS: -I${SRCDIR}/../../third_party/JoltPhysics
#cgo amd64 CXXFLAGS: -DJPH_USE_SSE4_1 -DJPH_USE_SSE4_2 -DJPH_USE_AVX
#cgo arm64 CXXFLAGS: -DJPH_USE_NEON
#cgo joltdouble CXXFLAGS: -DJPH_DOUBLE_PRECISION
#cgo LDFLAGS: -lstdc++ -lm
*/
import "C"
`
}
```

- [ ] **Step 4: Run, expect pass**

- [ ] **Step 5: Commit**

```bash
git add generator/emit_cgo_directives.go generator/emit_cgo_directives_test.go
git commit -m "feat(generator): emit internal/cjolt/cgo.go with cgo directives"
```

---

### Task 18: Smoke compile — write skeleton `internal/cjolt/` and verify cgo build

**Files:**
- Create: `internal/cjolt/cgo.go` (handwritten this once; later regenerated)
- Create: `internal/cjolt/cjolt_unity.cpp` (handwritten this once; later regenerated)
- Create: `internal/cjolt/cjolt_generated.go` (empty stub)
- Create: `internal/cjolt/cjolt_test.go`

This task confirms that joltc + Jolt actually compile via cgo on this machine before we generate anything. If this fails, no later task matters.

- [ ] **Step 1: Write the failing test**

`internal/cjolt/cjolt_test.go`:

```go
package cjolt_test

import (
    "testing"

    "github.com/gataengine/jolt-go/internal/cjolt"
)

func TestCgoLinks(t *testing.T) {
    // Calling RegisterTypes should not panic. UnregisterTypes pairs with it.
    cjolt.RegisterTypes()
    cjolt.UnregisterTypes()
}
```

- [ ] **Step 2: Run, expect failure**

Run: `go test ./internal/cjolt/...`
Expected: build error — package `cjolt` undefined.

- [ ] **Step 3: Hand-write minimal cjolt to prove compilation works**

Run the generator manually to produce `cgo.go` and `cjolt_unity.cpp`:

```bash
go run ./generator/cmd/gen --emit-cgo > internal/cjolt/cgo.go
go run ./generator/cmd/gen --emit-unity third_party/JoltPhysics/Jolt/Jolt.cmake third_party/joltc/CMakeLists.txt > internal/cjolt/cjolt_unity.cpp
```

(Adjust generator main to support `--emit-cgo` and `--emit-unity` flags as part of this task — minimal flag plumbing is fine.)

`internal/cjolt/cjolt_generated.go`:

```go
// Code generated by jolt-go generator. DO NOT EDIT.

package cjolt

/*
#include "joltc.h"
*/
import "C"

func RegisterTypes()   { C.JPH_RegisterTypes() }
func UnregisterTypes() { C.JPH_UnregisterTypes() }
```

- [ ] **Step 4: Run, expect pass**

Run: `go test ./internal/cjolt/...`
Expected: PASS. First build will take 30–60s as Jolt compiles.

If linker errors appear (missing symbols, layer-bits mismatches), fix the cgo flags or the unity-stub include list. This is the canonical "the source build works" gate.

- [ ] **Step 5: Commit**

```bash
git add generator/ internal/cjolt/
git commit -m "feat(cjolt): smoke-compile joltc + Jolt unity build via cgo"
```

---

## Phase 4 — Generator: cjolt raw wrappers

### Task 19: Emit raw cgo wrappers for free functions

**Files:**
- Create: `generator/emit_cjolt.go`
- Create: `generator/emit_cjolt_test.go`

- [ ] **Step 1: Write the failing test**

```go
var _ = Describe("EmitCjolt", func() {
    Context("free functions", func() {
        h := &generator.Header{
            Funcs: []generator.Func{
                {Name: "JPH_RegisterTypes", Return: "void"},
                {Name: "JPH_UnregisterTypes", Return: "void"},
            },
        }
        It("emits Go wrappers calling C functions", func() {
            src := generator.EmitCjolt(h)
            Expect(src).To(ContainSubstring("package cjolt"))
            Expect(src).To(ContainSubstring("func JPH_RegisterTypes() { C.JPH_RegisterTypes() }"))
            Expect(src).To(ContainSubstring("func JPH_UnregisterTypes() { C.JPH_UnregisterTypes() }"))
        })
    })
})
```

- [ ] **Step 2: Run, expect failure**

- [ ] **Step 3: Implement**

`generator/emit_cjolt.go`:

```go
package generator

import (
    "fmt"
    "go/format"
    "strings"
)

// EmitCjolt returns the source of internal/cjolt/cjolt_generated.go.
func EmitCjolt(h *Header) string {
    var b strings.Builder
    b.WriteString("// Code generated by jolt-go generator. DO NOT EDIT.\n\npackage cjolt\n\n")
    b.WriteString("/*\n#include \"joltc.h\"\n*/\nimport \"C\"\n\n")
    for _, f := range h.Funcs {
        b.WriteString(emitCjoltFunc(f))
    }
    out, err := format.Source([]byte(b.String()))
    if err != nil {
        return b.String() // unformatted; tests will catch it
    }
    return string(out)
}

func emitCjoltFunc(f Func) string {
    if f.Return == "void" && len(f.Params) == 0 {
        return fmt.Sprintf("func %s() { C.%s() }\n", f.Name, f.Name)
    }
    // Other shapes implemented in later tasks.
    return ""
}
```

- [ ] **Step 4: Run, expect pass**

- [ ] **Step 5: Commit**

```bash
git add generator/emit_cjolt.go generator/emit_cjolt_test.go
git commit -m "feat(generator): emit raw cgo wrappers for void/no-arg free functions"
```

---

### Task 20: Emit raw cgo wrappers for methods (self pointer + scalar args)

**Files:**
- Modify: `generator/emit_cjolt.go`
- Modify: `generator/emit_cjolt_test.go`

- [ ] **Step 1: Write the failing test**

```go
Context("method-shaped functions", func() {
    h := &generator.Header{
        Funcs: []generator.Func{
            {
                Name:   "JPH_Body_GetID",
                Return: "JPH_BodyID",
                Params: []generator.Param{{Name: "self", CType: "const JPH_Body *"}},
            },
            {
                Name:   "JPH_Body_SetMotionType",
                Return: "void",
                Params: []generator.Param{
                    {Name: "self", CType: "JPH_Body *"},
                    {Name: "type", CType: "JPH_MotionType"},
                },
            },
        },
    }
    It("emits self pointer wrapper", func() {
        src := generator.EmitCjolt(h)
        Expect(src).To(ContainSubstring(
            "func JPH_Body_GetID(self *C.JPH_Body) C.JPH_BodyID {",
        ))
        Expect(src).To(ContainSubstring(
            "return C.JPH_Body_GetID(self)",
        ))
    })
    It("emits enum-arg method", func() {
        src := generator.EmitCjolt(h)
        Expect(src).To(ContainSubstring(
            "func JPH_Body_SetMotionType(self *C.JPH_Body, type_ C.JPH_MotionType) {",
        ))
    })
})
```

- [ ] **Step 2: Run, expect failure**

- [ ] **Step 3: Extend `emitCjoltFunc`**

Handle:
- Pointer params (`<T> *` or `const <T> *`) → `*C.<T>`
- Scalar/enum params → `C.<CType>`
- Reserved-word param names (`type`, `func`, etc.) → suffix with `_`
- Return values: pass through `C.<Return>` unless `void`.

Generate:
```go
func <Name>(<params>) [<retType>] {
    [return ]C.<Name>(<argList>)
}
```

- [ ] **Step 4: Run, expect pass**

- [ ] **Step 5: Commit**

```bash
git add generator/emit_cjolt.go generator/emit_cjolt_test.go
git commit -m "feat(generator): emit raw cgo wrappers for method-shaped funcs"
```

---

### Task 21: Math-type translation at the cgo boundary (`Vec3`/`Quat`/`Mat44`/`RVec3`)

**Files:**
- Modify: `generator/emit_cjolt.go`
- Modify: `generator/emit_cjolt_test.go`

- [ ] **Step 1: Write the failing test**

```go
Context("math types", func() {
    h := &generator.Header{
        Funcs: []generator.Func{
            {
                Name:   "JPH_Body_GetPosition",
                Return: "JPH_RVec3",
                Params: []generator.Param{{Name: "self", CType: "const JPH_Body *"}},
            },
            {
                Name:   "JPH_Body_SetLinearVelocity",
                Return: "void",
                Params: []generator.Param{
                    {Name: "self", CType: "JPH_Body *"},
                    {Name: "v", CType: "const JPH_Vec3 *"},
                },
            },
        },
    }
    It("returns RVec3 by value, taking the C type and passing through", func() {
        src := generator.EmitCjolt(h)
        Expect(src).To(ContainSubstring(
            "func JPH_Body_GetPosition(self *C.JPH_Body) C.JPH_RVec3 {",
        ))
    })
    It("accepts JPH_Vec3 pointer through unsafe.Pointer cast at higher layer", func() {
        src := generator.EmitCjolt(h)
        // cjolt is "raw"; it accepts the C struct pointer directly. The unsafe
        // cast happens in the public API. Verify the wrapper signature here.
        Expect(src).To(ContainSubstring(
            "func JPH_Body_SetLinearVelocity(self *C.JPH_Body, v *C.JPH_Vec3) {",
        ))
    })
})
```

- [ ] **Step 2: Run, expect failure**

(Should already work for value-return and pointer-arg if Task 20 was thorough. If not, add handling now.)

- [ ] **Step 3: Implement / extend**

The cjolt layer is intentionally raw — math types stay as their C struct pointers. The public `jolt` layer does the `unsafe.Pointer(&goVec)` cast (Phase 5+).

- [ ] **Step 4: Run, expect pass**

- [ ] **Step 5: Commit**

```bash
git add generator/emit_cjolt.go generator/emit_cjolt_test.go
git commit -m "feat(generator): emit cjolt wrappers for math-type signatures"
```

---

## Phase 5 — Generator: idiomatic Go OOP

### Task 22: Annotations sidecar

**Files:**
- Create: `generator/annotations.go`
- Create: `generator/annotations_test.go`

- [ ] **Step 1: Write the failing test**

```go
var _ = Describe("Annotations", func() {
    It("declares default sidecar values", func() {
        a := generator.Annotations
        Expect(a.RefCounted).To(ContainElement("Shape"))
        Expect(a.Inheritance["BoxShape"]).To(Equal("ConvexShape"))
        Expect(a.SettingsInit["BodyCreationSettings"]).To(Equal("JPH_BodyCreationSettings_Init"))
    })
    It("flags listener procs structs", func() {
        l := findListener(generator.Annotations.Listeners, "JPH_ContactListener")
        Expect(l.Methods).To(HaveLen(4))
    })
})

func findListener(ls []generator.ListenerSpec, c string) generator.ListenerSpec { /* ... */ }
```

- [ ] **Step 2: Run, expect failure**

- [ ] **Step 3: Implement**

`generator/annotations.go`:

```go
package generator

// Strategy selects how a settings struct is emitted.
type Strategy int

const (
    StrategyDirect Strategy = iota // Approach 1 — layout-match
    StrategyTranslate              // Approach 2 — field-by-field
)

// ListenerSpec describes a JPH_*Listener vtable that the generator emits as
// a struct-of-funcs Go API.
type ListenerSpec struct {
    CType     string // e.g., "JPH_ContactListener"
    ProcsType string // e.g., "JPH_ContactListener_Procs"
    SetProcs  string // e.g., "JPH_ContactListener_SetProcs"
    Methods   []ListenerMethod
}

type ListenerMethod struct {
    Name    string
    Returns string // empty for void
}

// Sidecar covers the things joltc.h alone can't express.
type Sidecar struct {
    Inheritance      map[string]string   // child Go type -> direct parent Go type
    RefCounted       []string            // Go types using AddRef/Release
    Listeners        []ListenerSpec
    SettingsInit     map[string]string   // settings type -> JPH_*_Init function
    SettingsStrategy map[string]Strategy // override per type
}

// Annotations is the MVP sidecar populated for Phase 8 codegen runs.
// Add entries as MVP scope expands.
var Annotations = Sidecar{
    Inheritance: map[string]string{
        "BoxShape":              "ConvexShape",
        "SphereShape":           "ConvexShape",
        "CapsuleShape":          "ConvexShape",
        "TaperedCapsuleShape":   "ConvexShape",
        "CylinderShape":         "ConvexShape",
        "TaperedCylinderShape":  "ConvexShape",
        "ConvexHullShape":       "ConvexShape",
        "TriangleShape":         "ConvexShape",
        "PlaneShape":            "ConvexShape",
        "EmptyShape":            "Shape",
        "MeshShape":             "Shape",
        "HeightFieldShape":      "Shape",
        "StaticCompoundShape":   "CompoundShape",
        "MutableCompoundShape":  "CompoundShape",
        "CompoundShape":         "Shape",
        "RotatedTranslatedShape": "DecoratedShape",
        "ScaledShape":           "DecoratedShape",
        "OffsetCenterOfMassShape": "DecoratedShape",
        "DecoratedShape":        "Shape",
        "ConvexShape":           "Shape",
    },
    RefCounted: []string{
        "Shape", "ConvexShape", "CompoundShape", "DecoratedShape",
        "BoxShape", "SphereShape", "CapsuleShape", "TaperedCapsuleShape",
        "CylinderShape", "TaperedCylinderShape", "ConvexHullShape",
        "TriangleShape", "PlaneShape", "EmptyShape", "MeshShape",
        "HeightFieldShape", "StaticCompoundShape", "MutableCompoundShape",
        "RotatedTranslatedShape", "ScaledShape", "OffsetCenterOfMassShape",
        "PhysicsMaterial",
    },
    Listeners: []ListenerSpec{{
        CType:     "JPH_ContactListener",
        ProcsType: "JPH_ContactListener_Procs",
        SetProcs:  "JPH_ContactListener_SetProcs",
        Methods: []ListenerMethod{
            {Name: "OnContactValidate", Returns: "JPH_ValidateResult"},
            {Name: "OnContactAdded"},
            {Name: "OnContactPersisted"},
            {Name: "OnContactRemoved"},
        },
    }},
    SettingsInit: map[string]string{
        "BodyCreationSettings": "JPH_BodyCreationSettings_Init",
    },
    SettingsStrategy: map[string]Strategy{},
}
```

- [ ] **Step 4: Run, expect pass**

- [ ] **Step 5: Commit**

```bash
git add generator/annotations.go generator/annotations_test.go
git commit -m "feat(generator): add annotations sidecar for inheritance/refcounting/listeners/settings"
```

---

### Task 23: Emit owned types — `Close()` + finalizer pattern

**Files:**
- Create: `generator/emit_jolt.go`
- Create: `generator/emit_jolt_test.go`

- [ ] **Step 1: Write the failing test**

```go
var _ = Describe("EmitJolt — owned types", func() {
    h := &generator.Header{
        Funcs: []generator.Func{
            {Name: "JPH_PhysicsSystem_Create", Return: "JPH_PhysicsSystem*"},
            {Name: "JPH_PhysicsSystem_Destroy", Return: "void",
                Params: []generator.Param{{Name: "self", CType: "JPH_PhysicsSystem *"}}},
        },
    }
    It("generates type, constructor, Close, finalizer", func() {
        src := generator.EmitJolt(h)
        Expect(src).To(ContainSubstring("type PhysicsSystem struct {"))
        Expect(src).To(ContainSubstring("func NewPhysicsSystem() *PhysicsSystem {"))
        Expect(src).To(ContainSubstring("runtime.SetFinalizer(p, (*PhysicsSystem).finalize)"))
        Expect(src).To(ContainSubstring("func (p *PhysicsSystem) Close() error {"))
        Expect(src).To(ContainSubstring("C.JPH_PhysicsSystem_Destroy(p.h)"))
        Expect(src).To(ContainSubstring("runtime.SetFinalizer(p, nil)"))
    })
})
```

- [ ] **Step 2: Run, expect failure**

- [ ] **Step 3: Implement `EmitJolt` for owned types**

`generator/emit_jolt.go`:

```go
package generator

import (
    "fmt"
    "go/format"
    "strings"
)

// EmitJolt returns the source for jolt/jolt_generated.go.
func EmitJolt(h *Header) string {
    var b strings.Builder
    b.WriteString("// Code generated by jolt-go generator. DO NOT EDIT.\n\npackage jolt\n\n")
    b.WriteString(`import (
    "runtime"
    "unsafe"

    "github.com/gataengine/jolt-go/internal/cjolt"
)
`)
    b.WriteString("// silence unused-import warnings until first use\nvar _ = unsafe.Sizeof(0)\nvar _ = runtime.GC\n\n")

    // Pair Create/Destroy by type name and emit one OwnedType block per pair
    // that is *not* in the RefCounted set.
    pairs := pairCreateDestroy(h)
    for _, p := range pairs {
        if isRefCounted(p.GoType) {
            continue
        }
        b.WriteString(emitOwned(p))
    }

    out, err := format.Source([]byte(b.String()))
    if err != nil {
        return b.String()
    }
    return string(out)
}

// pairCreateDestroy returns matched (Create, Destroy) function pairs.
type ownedPair struct {
    GoType  string
    Create  Func
    Destroy Func
}

func pairCreateDestroy(h *Header) []ownedPair {
    creates := map[string]Func{}
    destroys := map[string]Func{}
    for _, f := range h.Funcs {
        recv, method, ok := SplitMethod(f.Name)
        if !ok {
            continue
        }
        switch method {
        case "Create":
            creates[recv] = f
        case "Destroy":
            destroys[recv] = f
        }
    }
    var out []ownedPair
    for t, c := range creates {
        if d, ok := destroys[t]; ok {
            out = append(out, ownedPair{GoType: t, Create: c, Destroy: d})
        }
    }
    return out
}

func isRefCounted(t string) bool {
    for _, r := range Annotations.RefCounted {
        if r == t {
            return true
        }
    }
    return false
}

func emitOwned(p ownedPair) string {
    return fmt.Sprintf(`type %[1]s struct{ h *cjolt.C_JPH_%[1]s }

func New%[1]s() *%[1]s {
    p := &%[1]s{h: cjolt.JPH_%[1]s_Create()}
    runtime.SetFinalizer(p, (*%[1]s).finalize)
    return p
}

func (p *%[1]s) Close() error {
    if p.h != nil {
        cjolt.JPH_%[1]s_Destroy(p.h)
        p.h = nil
        runtime.SetFinalizer(p, nil)
    }
    return nil
}

func (p *%[1]s) finalize() { _ = p.Close() }

`, p.GoType)
}
```

(Note: `cjolt.C_JPH_PhysicsSystem` is the Go alias the cjolt layer must export. Add to Phase 4 emitter in a follow-up if not there yet — re-export of the C type pointers as Go-named types is required for the public API to reference them without re-importing C.)

- [ ] **Step 4: Run, expect pass**

- [ ] **Step 5: Commit**

```bash
git add generator/emit_jolt.go generator/emit_jolt_test.go
git commit -m "feat(generator): emit owned-type Go API (Close + finalizer)"
```

---

### Task 24: Emit `cjolt` re-exports of opaque C types

**Files:**
- Modify: `generator/emit_cjolt.go`
- Modify: `generator/emit_cjolt_test.go`

The public `jolt` layer cannot say `C.JPH_PhysicsSystem` (cgo names don't cross packages). The cjolt layer must re-export each opaque struct as a Go-named alias (`type C_JPH_PhysicsSystem = C.JPH_PhysicsSystem`).

- [ ] **Step 1: Write the failing test**

```go
Context("opaque type re-exports", func() {
    h := &generator.Header{
        Structs: []generator.Struct{
            {Name: "JPH_PhysicsSystem", Opaque: true},
            {Name: "JPH_BodyInterface", Opaque: true},
        },
    }
    It("emits Go aliases for each opaque type", func() {
        src := generator.EmitCjolt(h)
        Expect(src).To(ContainSubstring("type C_JPH_PhysicsSystem = C.JPH_PhysicsSystem"))
        Expect(src).To(ContainSubstring("type C_JPH_BodyInterface = C.JPH_BodyInterface"))
    })
})
```

- [ ] **Step 2: Run, expect failure**

- [ ] **Step 3: Extend `EmitCjolt`**

After the `import "C"` block, emit a `type C_<CName> = C.<CName>` line for every opaque struct.

- [ ] **Step 4: Run, expect pass**

- [ ] **Step 5: Commit**

```bash
git add generator/emit_cjolt.go generator/emit_cjolt_test.go
git commit -m "feat(generator): re-export opaque C types as Go-visible aliases"
```

---

### Task 25: Emit RefTarget types — hidden ref counting + finalizer

**Files:**
- Modify: `generator/emit_jolt.go`
- Modify: `generator/emit_jolt_test.go`

- [ ] **Step 1: Write the failing test**

```go
Context("ref-counted types", func() {
    h := &generator.Header{
        Funcs: []generator.Func{
            {Name: "JPH_BoxShape_Create", Return: "JPH_Shape*",
                Params: []generator.Param{
                    {Name: "halfExtent", CType: "const JPH_Vec3 *"},
                    {Name: "convexRadius", CType: "float"},
                }},
            {Name: "JPH_Shape_AddRef", Return: "void",
                Params: []generator.Param{{Name: "self", CType: "JPH_Shape *"}}},
            {Name: "JPH_Shape_Release", Return: "void",
                Params: []generator.Param{{Name: "self", CType: "JPH_Shape *"}}},
        },
    }
    It("emits BoxShape with embedded ConvexShape", func() {
        src := generator.EmitJolt(h)
        Expect(src).To(ContainSubstring("type Shape struct{ h *cjolt.C_JPH_Shape }"))
        Expect(src).To(ContainSubstring("type ConvexShape struct{ Shape }"))
        Expect(src).To(ContainSubstring("type BoxShape struct{ ConvexShape }"))
    })
    It("emits NewBoxShape constructor with finalizer registering Release", func() {
        src := generator.EmitJolt(h)
        Expect(src).To(ContainSubstring("func NewBoxShape(halfExtent Vec3, convexRadius float32) *BoxShape {"))
        Expect(src).To(ContainSubstring("runtime.SetFinalizer(s, (*BoxShape).finalize)"))
        Expect(src).To(ContainSubstring("cjolt.JPH_Shape_Release(s.h)"))
    })
    It("does not export AddRef/Release on Shape", func() {
        src := generator.EmitJolt(h)
        Expect(src).NotTo(ContainSubstring("func (s *Shape) AddRef("))
        Expect(src).NotTo(ContainSubstring("func (s *Shape) Release("))
    })
})
```

- [ ] **Step 2: Run, expect failure**

- [ ] **Step 3: Implement**

Walk `Annotations.Inheritance` to emit each child as `type <Child> struct { <Parent> }`. Walk `Annotations.RefCounted` to emit constructors with finalizer-registered `Release`.

For each `JPH_<Type>_Create` returning a ref-counted type, emit:

```go
func New<ChildName>(<args>) *<ChildName> {
    h := cjolt.JPH_<ChildName>_Create(<argList>)
    s := &<ChildName>{ /* embedded chain initializer to root Shape{h: h} */ }
    runtime.SetFinalizer(s, (*<ChildName>).finalize)
    return s
}

func (s *<ChildName>) finalize() {
    if s.h != nil {
        cjolt.JPH_Shape_Release(s.h)
        s.h = nil
    }
}
```

`<argList>` translates Go-side math types via `(*cjolt.C_JPH_Vec3)(unsafe.Pointer(&halfExtent))`.

- [ ] **Step 4: Run, expect pass**

- [ ] **Step 5: Commit**

```bash
git add generator/emit_jolt.go generator/emit_jolt_test.go
git commit -m "feat(generator): emit ref-counted shape types with hidden refcounting"
```

---

### Task 26: Emit method bindings on receiver types

**Files:**
- Modify: `generator/emit_jolt.go`
- Modify: `generator/emit_jolt_test.go`

- [ ] **Step 1: Write the failing test**

```go
Context("methods on receiver types", func() {
    h := &generator.Header{
        Funcs: []generator.Func{
            {Name: "JPH_Body_GetPosition", Return: "JPH_RVec3",
                Params: []generator.Param{{Name: "self", CType: "const JPH_Body *"}}},
            {Name: "JPH_Body_SetLinearVelocity", Return: "void",
                Params: []generator.Param{
                    {Name: "self", CType: "JPH_Body *"},
                    {Name: "v", CType: "const JPH_Vec3 *"},
                }},
        },
        Structs: []generator.Struct{{Name: "JPH_Body", Opaque: true}},
    }
    It("emits getter returning value", func() {
        src := generator.EmitJolt(h)
        Expect(src).To(ContainSubstring("func (b *Body) Position() RVec3 {"))
        Expect(src).To(ContainSubstring("return *(*RVec3)(unsafe.Pointer(&out))"))
    })
    It("emits setter taking value, KeepAlive on receiver", func() {
        src := generator.EmitJolt(h)
        Expect(src).To(ContainSubstring("func (b *Body) SetLinearVelocity(v Vec3) {"))
        Expect(src).To(ContainSubstring("cjolt.JPH_Body_SetLinearVelocity(b.h, (*cjolt.C_JPH_Vec3)(unsafe.Pointer(&v)))"))
        Expect(src).To(ContainSubstring("runtime.KeepAlive(b)"))
    })
})
```

- [ ] **Step 2: Run, expect failure**

- [ ] **Step 3: Extend `EmitJolt`**

For every `JPH_<Recv>_<Method>` function whose first param is `<Recv>*`, emit a Go method on `*<Recv>`. Translate parameter types:

| C type | Go param | Wrapper conversion |
|---|---|---|
| `const JPH_Vec3 *` | `Vec3` | `(*cjolt.C_JPH_Vec3)(unsafe.Pointer(&v))` |
| `const JPH_Quat *` | `Quat` | `(*cjolt.C_JPH_Quat)(unsafe.Pointer(&q))` |
| `const JPH_RVec3 *` | `RVec3` | `(*cjolt.C_JPH_RVec3)(unsafe.Pointer(&v))` |
| `JPH_<Owned> *` | `*<Owned>` | `x.h` |
| `JPH_<Refcounted> *` | `*<Refcounted>` | `x.h` |
| Scalar / enum | mapped Go type | direct |
| Returns math by value | Returns Go math | `*(*Vec3)(unsafe.Pointer(&out))` |
| Returns opaque ptr | Returns wrapped Go type | wrap with finalizer |

Emit `runtime.KeepAlive(b)` after the cgo call when any pointer arg or receiver is involved.

- [ ] **Step 4: Run, expect pass**

- [ ] **Step 5: Commit**

```bash
git add generator/emit_jolt.go generator/emit_jolt_test.go
git commit -m "feat(generator): emit OOP methods with math-type unsafe.Pointer casts"
```

---

## Phase 6 — Generator: settings (Approach 1)

### Task 27: Settings — emit Go struct with layout matching C struct

**Files:**
- Modify: `generator/emit_jolt.go`
- Modify: `generator/emit_jolt_test.go`
- Modify: `generator/parser.go` (add struct field extraction with C scalar types)

- [ ] **Step 1: Write the failing test**

Add a settings fixture:

`generator/testdata/settings.h`:

```c
#ifndef SETTINGS_H
#define SETTINGS_H
#include <stdint.h>
#include <stdbool.h>

typedef struct JPH_Vec3 { float x, y, z; } JPH_Vec3;
typedef struct JPH_Quat { float x, y, z, w; } JPH_Quat;
typedef struct JPH_RVec3 { float x, y, z; } JPH_RVec3;
typedef struct JPH_Shape JPH_Shape;
typedef uint16_t JPH_ObjectLayer;
typedef uint8_t JPH_MotionType;

typedef struct JPH_BodyCreationSettings {
    JPH_RVec3       Position;
    JPH_Quat        Rotation;
    JPH_Vec3        LinearVelocity;
    JPH_Vec3        AngularVelocity;
    uint64_t        UserData;
    JPH_ObjectLayer ObjectLayer;
    JPH_MotionType  MotionType;
    bool            AllowSleeping;
    float           Friction;
    float           Restitution;
    float           LinearDamping;
    float           AngularDamping;
    JPH_Shape*      Shape;
} JPH_BodyCreationSettings;

void JPH_BodyCreationSettings_Init(JPH_BodyCreationSettings*);

#endif
```

Test:

```go
Context("settings struct (Approach 1)", func() {
    h, err := generator.ParseHeader("testdata/settings.h")
    BeforeEach(func() { Expect(err).NotTo(HaveOccurred()) })
    It("emits a layout-match Go struct", func() {
        src := generator.EmitJolt(h)
        Expect(src).To(ContainSubstring("type BodyCreationSettings struct {"))
        Expect(src).To(ContainSubstring("Position        RVec3"))
        Expect(src).To(ContainSubstring("Rotation        Quat"))
        Expect(src).To(ContainSubstring("UserData        uint64"))
        Expect(src).To(ContainSubstring("ObjectLayer     ObjectLayer"))
        Expect(src).To(ContainSubstring("MotionType      MotionType"))
        Expect(src).To(ContainSubstring("AllowSleeping   bool"))
        Expect(src).To(ContainSubstring("shape           uintptr"))
    })
    It("emits SetShape using uintptr", func() {
        src := generator.EmitJolt(h)
        Expect(src).To(ContainSubstring("func (s *BodyCreationSettings) SetShape(shape *Shape) {"))
        Expect(src).To(ContainSubstring("s.shape = uintptr(unsafe.Pointer(shape.h))"))
        Expect(src).To(ContainSubstring("runtime.KeepAlive(shape)"))
    })
})
```

- [ ] **Step 2: Run, expect failure**

- [ ] **Step 3: Implement**

Extend the parser to extract struct fields with their C types (scalar, math typedefs, opaque pointer). In the emitter, walk `Annotations.SettingsInit` (every key is a settings type by definition). For each:
- Emit Go struct mirroring layout — math typedefs map to Go math, scalar typedefs map to their Go integer alias, `Shape*` maps to unexported `shape uintptr`.
- Emit `Set<Field>` setters for each pointer field.
- Emit `New<Type>` constructor calling `JPH_<Type>_Init` on a zeroed C struct, copying defaults out, returning a Go struct.

- [ ] **Step 4: Run, expect pass**

- [ ] **Step 5: Commit**

```bash
git add generator/parser.go generator/emit_jolt.go generator/emit_jolt_test.go generator/testdata/settings.h
git commit -m "feat(generator): emit Approach-1 settings (layout-match + uintptr shapes)"
```

---

### Task 28: Settings — `_Static_assert` size guard in unity stub

**Files:**
- Modify: `generator/emit_unity.go`
- Modify: `generator/emit_unity_test.go`

- [ ] **Step 1: Write the failing test**

```go
It("emits _Static_assert(sizeof) for each Approach-1 settings type", func() {
    out := generator.EmitUnity(nil, nil, generator.UnityOpts{
        SettingsSizeAsserts: map[string]int{
            "JPH_BodyCreationSettings": 152,
        },
    })
    Expect(out).To(ContainSubstring(
        `_Static_assert(sizeof(JPH_BodyCreationSettings) == 152, "JPH_BodyCreationSettings layout drift");`,
    ))
})
```

- [ ] **Step 2: Run, expect failure**

- [ ] **Step 3: Implement**

Extend `EmitUnity` to take a `UnityOpts{ SettingsSizeAsserts map[string]int }` param. Compute sizes by emitting a tiny C++ helper that `printf`s `sizeof()` during a generator-time invocation? — too heavy. Simpler: have the generator emit each assert with the **current `sizeof`** discovered by calling Go's `unsafe.Sizeof(C.JPH_BodyCreationSettings{})` from the cjolt layer at codegen time. For simplicity in MVP, hard-code the expected sizes from the first successful build into the sidecar (`Annotations.SettingsSizes`). When sizes drift, the `_Static_assert` fails at C++ compile time.

Add `Annotations.SettingsSizes map[string]int` and pass it through.

- [ ] **Step 4: Run, expect pass**

- [ ] **Step 5: Commit**

```bash
git add generator/
git commit -m "feat(generator): emit _Static_assert size guards for Approach-1 settings"
```

---

### Task 29: Settings — emit call-site translation (`unsafe.Pointer` cast)

**Files:**
- Modify: `generator/emit_jolt.go`
- Modify: `generator/emit_jolt_test.go`

- [ ] **Step 1: Write the failing test**

```go
Context("BodyInterface.CreateBody passes settings via unsafe.Pointer cast", func() {
    h := &generator.Header{
        Funcs: []generator.Func{
            {Name: "JPH_BodyInterface_CreateBody", Return: "JPH_Body*",
                Params: []generator.Param{
                    {Name: "self", CType: "JPH_BodyInterface *"},
                    {Name: "settings", CType: "const JPH_BodyCreationSettings *"},
                    {Name: "activation", CType: "JPH_Activation"},
                }},
        },
    }
    It("emits CreateBody method with cast", func() {
        src := generator.EmitJolt(h)
        Expect(src).To(ContainSubstring(
            "func (bi *BodyInterface) CreateBody(s *BodyCreationSettings, a Activation) *Body {",
        ))
        Expect(src).To(ContainSubstring(
            "cjolt.JPH_BodyInterface_CreateBody(bi.h, (*cjolt.C_JPH_BodyCreationSettings)(unsafe.Pointer(s)), cjolt.C_JPH_Activation(a))",
        ))
        Expect(src).To(ContainSubstring("runtime.KeepAlive(s)"))
    })
})
```

- [ ] **Step 2: Run, expect failure**

- [ ] **Step 3: Implement**

In the parameter-translation table for the method emitter, add:
- `const JPH_<Settings> *` → Go param `*<Settings>`, wrapper `(*cjolt.C_JPH_<Settings>)(unsafe.Pointer(s))` + `runtime.KeepAlive(s)` after the call.

- [ ] **Step 4: Run, expect pass**

- [ ] **Step 5: Commit**

```bash
git add generator/emit_jolt.go generator/emit_jolt_test.go
git commit -m "feat(generator): emit settings call-site as unsafe.Pointer cast"
```

---

## Phase 7 — Generator: listener vtables

### Task 30: Emit ContactListener struct-of-funcs Go API

**Files:**
- Modify: `generator/emit_jolt.go`
- Modify: `generator/emit_jolt_test.go`

- [ ] **Step 1: Write the failing test**

```go
Context("ContactListener struct-of-funcs", func() {
    h := &generator.Header{} // listener emitted from Annotations.Listeners
    It("emits public Go struct with func fields", func() {
        src := generator.EmitJolt(h)
        Expect(src).To(ContainSubstring("type ContactListener struct {"))
        Expect(src).To(ContainSubstring(
            "OnContactValidate  func(b1, b2 *Body, baseOffset RVec3, result *CollideShapeResult) ValidateResult",
        ))
        Expect(src).To(ContainSubstring(
            "OnContactAdded     func(b1, b2 *Body, manifold *ContactManifold, settings *ContactSettings)",
        ))
        Expect(src).To(ContainSubstring(
            "OnContactPersisted func(b1, b2 *Body, manifold *ContactManifold, settings *ContactSettings)",
        ))
        Expect(src).To(ContainSubstring(
            "OnContactRemoved   func(subShapePair *SubShapeIDPair)",
        ))
        Expect(src).To(ContainSubstring("h      *cjolt.C_JPH_ContactListener"))
        Expect(src).To(ContainSubstring("handle cgo.Handle"))
    })
    It("emits NewContactListener registering cgo.Handle as userData", func() {
        src := generator.EmitJolt(h)
        Expect(src).To(ContainSubstring("func NewContactListener(l ContactListener) *ContactListener {"))
        Expect(src).To(ContainSubstring("cl.handle = cgo.NewHandle(cl)"))
        Expect(src).To(ContainSubstring("cl.h = cjolt.JPH_ContactListener_Create(unsafe.Pointer(uintptr(cl.handle)))"))
        Expect(src).To(ContainSubstring("runtime.SetFinalizer(cl, (*ContactListener).finalize)"))
    })
})
```

- [ ] **Step 2: Run, expect failure**

- [ ] **Step 3: Implement**

Walk `Annotations.Listeners`. For each `ListenerSpec`:
- Emit Go struct with one func field per `Method` (signatures defined by the spec; for MVP these are hardcoded into `ListenerMethod` until parser-driven inference exists — fine for one listener).
- Emit `New<Listener>(l <Listener>) *<Listener>` factory that creates a `cgo.Handle`, calls the C `_Create` with `userData = unsafe.Pointer(uintptr(handle))`, and registers a finalizer that destroys the C object and Deletes the handle.

Add `"runtime/cgo"` to the import list.

- [ ] **Step 4: Run, expect pass**

- [ ] **Step 5: Commit**

```bash
git add generator/emit_jolt.go generator/emit_jolt_test.go
git commit -m "feat(generator): emit ContactListener struct-of-funcs API + cgo.Handle plumbing"
```

---

### Task 31: Emit listener trampolines + global `_SetProcs` registration

**Files:**
- Modify: `generator/emit_cjolt.go`
- Modify: `generator/emit_cjolt_test.go`

The trampolines live in cjolt because cgo.Handle access happens on the Go side, but the C function pointers exposed to joltc must be Go functions with `//export` directives. We emit `*_trampoline.go` content into `cjolt_generated.go`.

- [ ] **Step 1: Write the failing test**

```go
Context("listener trampolines", func() {
    h := &generator.Header{}
    It("emits exported trampolines and init() that registers Procs", func() {
        src := generator.EmitCjolt(h)
        Expect(src).To(ContainSubstring("//export jolt_go_contactListener_OnContactAdded"))
        Expect(src).To(ContainSubstring(
            "func jolt_go_contactListener_OnContactAdded(userData unsafe.Pointer, b1, b2 *C.JPH_Body, manifold *C.JPH_ContactManifold, settings *C.JPH_ContactSettings) {",
        ))
        Expect(src).To(ContainSubstring(
            `h := cgo.Handle(uintptr(userData))`,
        ))
        Expect(src).To(ContainSubstring(`cl := h.Value().(*ContactListenerHandle)`))
        Expect(src).To(ContainSubstring("func init() {"))
        Expect(src).To(ContainSubstring("C.JPH_ContactListener_SetProcs(C.JPH_ContactListener_Procs{"))
    })
})
```

- [ ] **Step 2: Run, expect failure**

- [ ] **Step 3: Implement**

For each listener, emit:

```go
// ContactListenerHandle is the value type stored in cgo.Handle for joltc
// ContactListener instances. The Go-side handlers in package jolt set the
// non-nil func fields on this struct before installing the listener.
type ContactListenerHandle struct {
    OnContactValidate  func(...) C.JPH_ValidateResult
    OnContactAdded     func(...)
    OnContactPersisted func(...)
    OnContactRemoved   func(...)
}

//export jolt_go_contactListener_OnContactAdded
func jolt_go_contactListener_OnContactAdded(userData unsafe.Pointer, b1, b2 *C.JPH_Body, manifold *C.JPH_ContactManifold, settings *C.JPH_ContactSettings) {
    h := cgo.Handle(uintptr(userData))
    cl := h.Value().(*ContactListenerHandle)
    if cl.OnContactAdded != nil {
        cl.OnContactAdded(b1, b2, manifold, settings)
    }
}
// ... similar for the other three methods

func init() {
    C.JPH_ContactListener_SetProcs(C.JPH_ContactListener_Procs{
        OnContactValidate:  (C.JPH_ContactListener_OnContactValidate_Func)(C.jolt_go_contactListener_OnContactValidate),
        OnContactAdded:     (C.JPH_ContactListener_OnContactAdded_Func)(C.jolt_go_contactListener_OnContactAdded),
        OnContactPersisted: (C.JPH_ContactListener_OnContactPersisted_Func)(C.jolt_go_contactListener_OnContactPersisted),
        OnContactRemoved:   (C.JPH_ContactListener_OnContactRemoved_Func)(C.jolt_go_contactListener_OnContactRemoved),
    })
}
```

The `jolt` package's `NewContactListener` wraps the user-supplied `ContactListener` into a `cjolt.ContactListenerHandle`, builds adapters that translate C types to Go types and invoke the user's `func` fields, then stores the handle.

- [ ] **Step 4: Run, expect pass**

- [ ] **Step 5: Commit**

```bash
git add generator/emit_cjolt.go generator/emit_cjolt_test.go
git commit -m "feat(generator): emit listener trampolines + init() Procs registration"
```

---

## Phase 8 — Run generator on real `joltc.h`

### Task 32: Wire generator main + `go generate` directive

**Files:**
- Modify: `generator/cmd/gen/main.go`
- Modify: `jolt/doc.go` (add `go:generate` directive)

- [ ] **Step 1: Write the failing test**

(Behavioral test — the binary should regenerate all four files when invoked with no flags from the repo root.)

```go
var _ = Describe("Generator binary", func() {
    It("regenerates all four output files when run", func() {
        // 1. delete a generated file
        // 2. run `go run ./generator/cmd/gen`
        // 3. assert the file is recreated
        Skip("integration; runs in Phase 8 Task 35 after submodules are wired")
    })
})
```

The skip is intentional — this test gates Task 35.

- [ ] **Step 2: N/A (skipped test passes vacuously)**

- [ ] **Step 3: Implement `main.go`**

```go
package main

import (
    "log"
    "os"
    "path/filepath"

    gen "github.com/gataengine/jolt-go/generator"
)

func main() {
    repoRoot, err := os.Getwd()
    if err != nil {
        log.Fatal(err)
    }
    headerPath := filepath.Join(repoRoot, "third_party", "joltc", "include", "joltc.h")
    cmakePath := filepath.Join(repoRoot, "third_party", "JoltPhysics", "Jolt", "Jolt.cmake")

    h, err := gen.ParseHeader(headerPath)
    if err != nil {
        log.Fatalf("parse: %v", err)
    }
    joltSources, err := gen.ExtractJoltSources(cmakePath)
    if err != nil {
        log.Fatalf("cmake: %v", err)
    }

    write := func(path, content string) {
        if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
            log.Fatalf("write %s: %v", path, err)
        }
    }
    write(filepath.Join(repoRoot, "internal", "cjolt", "cgo.go"), gen.EmitCgoDirectives())
    write(filepath.Join(repoRoot, "internal", "cjolt", "cjolt_unity.cpp"),
        gen.EmitUnity(joltSources, gen.JoltcSources, gen.UnityOpts{
            SettingsSizeAsserts: gen.Annotations.SettingsSizes,
        }))
    write(filepath.Join(repoRoot, "internal", "cjolt", "cjolt_generated.go"), gen.EmitCjolt(h))
    write(filepath.Join(repoRoot, "jolt", "jolt_generated.go"), gen.EmitJolt(h))
    log.Print("generated")
}
```

Add to `jolt/doc.go`:

```go
//go:generate go run ../generator/cmd/gen
```

- [ ] **Step 4: Run, expect compile**

```bash
go build ./generator/cmd/gen
```

- [ ] **Step 5: Commit**

```bash
git add generator/cmd/gen/main.go jolt/doc.go
git commit -m "feat(generator): wire main binary + go:generate directive"
```

---

### Task 33: Run generator against real `joltc.h`, fix issues until output compiles

**Files:**
- Regenerated: `internal/cjolt/cgo.go`, `internal/cjolt/cjolt_generated.go`, `internal/cjolt/cjolt_unity.cpp`, `jolt/jolt_generated.go`
- Possibly modified: `generator/parser.go`, `generator/emit_cjolt.go`, `generator/emit_jolt.go`, `generator/annotations.go`

This is the integration moment: the real header pushes against parser and emitter assumptions. Expect to iterate.

- [ ] **Step 1: Run the generator**

```bash
go generate ./jolt
```

Expected on first run: probable failures around C constructs the parser doesn't yet handle (function-pointer typedefs, unions, anonymous structs). Each failure is fixed in the parser/emitter and re-run.

- [ ] **Step 2: Verify cgo build succeeds**

```bash
go build ./...
```

Expected: success (compile time 30–60s on first build). If the C++ side fails on a `_Static_assert`, update `Annotations.SettingsSizes` with the printed actual size, regenerate, retry.

- [ ] **Step 3: Verify generator unit tests still pass**

```bash
go test ./generator/...
```

Expected: PASS.

- [ ] **Step 4: Verify the cjolt smoke test still passes**

```bash
go test ./internal/cjolt/...
```

Expected: PASS — `RegisterTypes` / `UnregisterTypes` still callable.

- [ ] **Step 5: Commit the generated outputs (and any generator fixes)**

```bash
git add internal/cjolt/ jolt/jolt_generated.go generator/
git commit -m "feat: first end-to-end generation against real joltc.h"
```

---

### Task 34: Add `jolt_handwritten.go` carve-out file (empty in MVP)

**Files:**
- Create: `jolt/jolt_handwritten.go`

- [ ] **Step 1: Write the failing test**

(No test — this file exists as a slot for post-MVP carve-outs like `StateRecorder`. Verifying its presence is via `go vet`.)

- [ ] **Step 2: Run vet, confirm no errors**

- [ ] **Step 3: Create the file**

```go
package jolt

// jolt_handwritten.go is the home for carve-outs that the generator can't
// or shouldn't produce. Empty in MVP; populated post-MVP (e.g., StateRecorder
// in Slice 11).
```

- [ ] **Step 4: Run `go vet ./...`**

Expected: clean.

- [ ] **Step 5: Commit**

```bash
git add jolt/jolt_handwritten.go
git commit -m "chore(jolt): add empty handwritten carve-out file"
```

---

## Phase 9 — Public API verification (lifetime types)

These tasks verify the *generated* code actually works. If a test reveals a missing/wrong emission, fix the generator (or annotations), regenerate, retest.

### Task 35: PhysicsSystem create / Close / finalize

**Files:**
- Create: `jolt/lifetime_test.go`

- [ ] **Step 1: Write the failing test**

```go
package jolt_test

import (
    "runtime"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"

    "github.com/gataengine/jolt-go/jolt"
)

var _ = Describe("PhysicsSystem", func() {
    It("creates and closes without leaks", func() {
        jolt.RegisterTypes()
        defer jolt.UnregisterTypes()

        ps := jolt.NewPhysicsSystem()
        Expect(ps).NotTo(BeNil())
        Expect(ps.Close()).NotTo(HaveOccurred())
        // double-close is a no-op
        Expect(ps.Close()).NotTo(HaveOccurred())
    })

    It("finalizes if user forgets Close", func() {
        jolt.RegisterTypes()
        defer jolt.UnregisterTypes()

        func() {
            _ = jolt.NewPhysicsSystem()
            // intentionally drop the reference
        }()
        runtime.GC()
        runtime.GC()
        // No assertion; if the finalizer panics we'd see it
    })
})
```

- [ ] **Step 2: Run, expect failure or pass**

If the generator emitted PhysicsSystem correctly, this passes immediately. If it failed, fix annotations / generator, regenerate, retry.

- [ ] **Step 3: Fix as needed and regenerate**

`go generate ./jolt && go test ./jolt/...`

- [ ] **Step 4: Run, expect pass**

- [ ] **Step 5: Commit**

```bash
git add jolt/lifetime_test.go [generator/, jolt/jolt_generated.go if changed]
git commit -m "test(jolt): verify PhysicsSystem create/Close/finalize"
```

---

### Task 36: TempAllocator + JobSystem create / Close

**Files:**
- Modify: `jolt/lifetime_test.go`

- [ ] **Step 1: Write the failing test**

```go
var _ = Describe("TempAllocator", func() {
    It("creates with a size and closes", func() {
        ta := jolt.NewTempAllocator(10 * 1024 * 1024) // 10 MB
        Expect(ta).NotTo(BeNil())
        Expect(ta.Close()).NotTo(HaveOccurred())
    })
})

var _ = Describe("JobSystem", func() {
    It("creates a thread pool job system and closes", func() {
        js := jolt.NewJobSystemThreadPool(jolt.JobSystemDefaultConfig())
        Expect(js).NotTo(BeNil())
        Expect(js.Close()).NotTo(HaveOccurred())
    })
})
```

(Concrete constructor names may need to follow joltc's actual names — adjust to whatever `JPH_TempAllocator_Create` and `JPH_JobSystemThreadPool_Create` accept after parser inspection. Update annotations if needed for `JobSystemDefaultConfig` etc.)

- [ ] **Step 2: Run, expect failure or pass**

- [ ] **Step 3: Fix generator/annotations as needed**

- [ ] **Step 4: Run, expect pass**

- [ ] **Step 5: Commit**

```bash
git add jolt/ generator/
git commit -m "test(jolt): verify TempAllocator + JobSystem lifetime"
```

---

## Phase 10 — Public API verification (BodyInterface, BodyCreationSettings, Body, Shape)

### Task 37: BodyCreationSettings constructor + defaults

**Files:**
- Create: `jolt/body_settings_test.go`

- [ ] **Step 1: Write the failing test**

```go
var _ = Describe("BodyCreationSettings", func() {
    It("populates defaults via _Init", func() {
        s := jolt.NewBodyCreationSettings()
        // Defaults from Jolt: Friction=0.2, Restitution=0, LinearDamping=0.05, AngularDamping=0.05
        Expect(s.Friction).To(BeNumerically("~", 0.2, 1e-6))
        Expect(s.LinearDamping).To(BeNumerically("~", 0.05, 1e-6))
        Expect(s.AngularDamping).To(BeNumerically("~", 0.05, 1e-6))
    })
    It("SetShape stores a uintptr to the shape handle", func() {
        s := jolt.NewBodyCreationSettings()
        shape := jolt.NewBoxShape(jolt.Vec3{1, 1, 1}, 0.05)
        s.SetShape(shape.AsShape())
        // Round-trip via the settings struct's package-private accessor (test-only export
        // through internal helper) — or just verify CreateBody works in Task 39.
    })
})
```

- [ ] **Step 2: Run, expect failure**

- [ ] **Step 3: Fix generator if defaults wrong; ensure `*BoxShape` exposes `AsShape() *Shape` upcast (emit on every embedded type pointing at the root)**

- [ ] **Step 4: Run, expect pass**

- [ ] **Step 5: Commit**

```bash
git add jolt/body_settings_test.go [generator/]
git commit -m "test(jolt): verify BodyCreationSettings _Init defaults + SetShape"
```

---

### Task 38: BoxShape construction + finalizer-driven `Release`

**Files:**
- Create: `jolt/shape_test.go`

- [ ] **Step 1: Write the failing test**

```go
var _ = Describe("BoxShape", func() {
    It("creates and is collected without panic", func() {
        s := jolt.NewBoxShape(jolt.Vec3{1, 1, 1}, 0.05)
        Expect(s).NotTo(BeNil())
        // Drop reference, force GC.
        s = nil
        runtime.GC()
        runtime.GC()
    })
    It("can be used to construct a body", func() {
        jolt.RegisterTypes()
        defer jolt.UnregisterTypes()
        ps := jolt.NewPhysicsSystem()
        defer ps.Close()
        bs := jolt.NewBodyCreationSettings()
        bs.Position = jolt.RVec3{0, 5, 0}
        bs.MotionType = jolt.MotionTypeDynamic
        bs.SetShape(jolt.NewBoxShape(jolt.Vec3{0.5, 0.5, 0.5}, 0.0).AsShape())
        body := ps.BodyInterface().CreateBody(bs, jolt.ActivationActivate)
        Expect(body).NotTo(BeNil())
    })
})
```

- [ ] **Step 2: Run, expect failure**

- [ ] **Step 3: Fix as needed (generator should emit `MotionTypeDynamic`, `ActivationActivate` enums, plus `BodyInterface()` accessor)**

- [ ] **Step 4: Run, expect pass**

- [ ] **Step 5: Commit**

```bash
git add jolt/shape_test.go [generator/]
git commit -m "test(jolt): verify BoxShape construction + Body creation"
```

---

### Task 39: Verify all 17 MVP shape types construct successfully

**Files:**
- Modify: `jolt/shape_test.go`

- [ ] **Step 1: Write the failing test**

Group into one `Describe` with one `It` per shape — quick smoke that each shape's constructor is correct and it can `AsShape()`.

```go
var _ = Describe("All MVP shapes construct", func() {
    DescribeTable("constructor smoke",
        func(name string, build func() interface{ AsShape() *jolt.Shape }) {
            s := build()
            Expect(s).NotTo(BeNil())
            Expect(s.AsShape()).NotTo(BeNil())
        },
        Entry("Box",                func() interface{ AsShape() *jolt.Shape } { return jolt.NewBoxShape(jolt.Vec3{1, 1, 1}, 0) }),
        Entry("Sphere",             func() interface{ AsShape() *jolt.Shape } { return jolt.NewSphereShape(0.5) }),
        Entry("Capsule",            func() interface{ AsShape() *jolt.Shape } { return jolt.NewCapsuleShape(0.5, 0.25) }),
        Entry("TaperedCapsule",     func() interface{ AsShape() *jolt.Shape } { return jolt.NewTaperedCapsuleShape(0.5, 0.25, 0.15) }),
        Entry("Cylinder",           func() interface{ AsShape() *jolt.Shape } { return jolt.NewCylinderShape(0.5, 0.25, 0) }),
        Entry("TaperedCylinder",    func() interface{ AsShape() *jolt.Shape } { return jolt.NewTaperedCylinderShape(0.5, 0.25, 0.15, 0) }),
        // ConvexHull, Triangle, Plane, Empty, Mesh, HeightField, StaticCompound, MutableCompound,
        // RotatedTranslated, Scaled, OffsetCenterOfMass — fill in following the same pattern.
    )
})
```

- [ ] **Step 2: Run, expect some failures for any constructor the generator handled wrong**

- [ ] **Step 3: Fix generator/annotations as needed; iterate**

For shapes whose constructors take complex inputs (e.g., `MeshShape` needs `MeshShapeSettings` containing arrays of vertices/triangles), they may need Approach 2 settings. Mark in `Annotations.SettingsStrategy` and add the per-call translation emitter — this likely needs an explicit Phase 6.5 task if not foreseen. For MVP, restrict each shape's smoke test to the simplest constructor available.

- [ ] **Step 4: Run, expect pass**

- [ ] **Step 5: Commit**

```bash
git add jolt/shape_test.go [generator/]
git commit -m "test(jolt): verify all MVP shape constructors"
```

---

## Phase 11 — Integration tests

### Task 40: Integration test — drop-a-box scenario

**Files:**
- Create: `jolt/integration_test.go`

- [ ] **Step 1: Write the failing test**

```go
package jolt_test

import (
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"

    "github.com/gataengine/jolt-go/jolt"
)

var _ = Describe("Integration: drop-a-box", func() {
    It("box falls and rests on the floor", func() {
        jolt.RegisterTypes()
        defer jolt.UnregisterTypes()

        ta := jolt.NewTempAllocator(10 * 1024 * 1024)
        defer ta.Close()
        js := jolt.NewJobSystemThreadPool(jolt.JobSystemDefaultConfig())
        defer js.Close()

        ps := jolt.NewPhysicsSystem()
        defer ps.Close()
        bi := ps.BodyInterface()

        // Static floor: large flat box at y=0
        floorShape := jolt.NewBoxShape(jolt.Vec3{50, 1, 50}, 0)
        floorSettings := jolt.NewBodyCreationSettings()
        floorSettings.Position = jolt.RVec3{0, -1, 0}
        floorSettings.MotionType = jolt.MotionTypeStatic
        floorSettings.ObjectLayer = jolt.ObjectLayerNonMoving
        floorSettings.SetShape(floorShape.AsShape())
        floorBody := bi.CreateAndAddBody(floorSettings, jolt.ActivationDontActivate)
        Expect(floorBody).NotTo(Equal(jolt.BodyID{}))

        // Dynamic box: 1m cube at y=10
        boxShape := jolt.NewBoxShape(jolt.Vec3{0.5, 0.5, 0.5}, 0)
        boxSettings := jolt.NewBodyCreationSettings()
        boxSettings.Position = jolt.RVec3{0, 10, 0}
        boxSettings.MotionType = jolt.MotionTypeDynamic
        boxSettings.ObjectLayer = jolt.ObjectLayerMoving
        boxSettings.SetShape(boxShape.AsShape())
        boxID := bi.CreateAndAddBody(boxSettings, jolt.ActivationActivate)

        // Step 120 frames at 60Hz
        const dt = 1.0 / 60.0
        for range 120 {
            ps.Update(dt, 1, ta, js)
        }

        // Box should now be near y ≈ 0.5 (resting on the +1m-thick floor's top
        // surface = y=0 plus half its 1m height).
        finalPos := bi.GetPosition(boxID)
        Expect(float64(finalPos.Y)).To(BeNumerically("~", 0.5, 0.05))
        Expect(float64(finalPos.X)).To(BeNumerically("~", 0, 0.05))
        Expect(float64(finalPos.Z)).To(BeNumerically("~", 0, 0.05))
    })
})
```

- [ ] **Step 2: Run, expect failure**

This will exercise *everything*. Likely failures expose missing emissions: `BodyInterface.CreateAndAddBody`, `PhysicsSystem.Update`, `ObjectLayerNonMoving` constant, the default broad-phase wiring inside `JPH_PhysicsSystem_Create`. Each requires adding to the generator/annotations.

- [ ] **Step 3: Fix generator/annotations until the test passes**

Iterate. Document any joltc init quirks discovered (`memset(0)` issue from spec §15) — make sure settings constructors zero before `_Init`.

- [ ] **Step 4: Run, expect pass**

```bash
go test ./jolt/...
```

Expected: PASS, including the new spec.

- [ ] **Step 5: Commit**

```bash
git add jolt/integration_test.go [generator/, jolt/jolt_generated.go]
git commit -m "test(jolt): drop-a-box integration test passing end-to-end"
```

---

### Task 41: Integration test — ContactListener fires `OnContactAdded`

**Files:**
- Modify: `jolt/integration_test.go`

- [ ] **Step 1: Write the failing test**

```go
var _ = Describe("Integration: ContactListener fires", func() {
    It("OnContactAdded is invoked when two bodies collide", func() {
        jolt.RegisterTypes()
        defer jolt.UnregisterTypes()

        ta := jolt.NewTempAllocator(10 * 1024 * 1024)
        defer ta.Close()
        js := jolt.NewJobSystemThreadPool(jolt.JobSystemDefaultConfig())
        defer js.Close()

        ps := jolt.NewPhysicsSystem()
        defer ps.Close()
        bi := ps.BodyInterface()

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
        defer listener.Close()
        ps.SetContactListener(listener)

        // Static floor + dynamic box dropped on it (same as drop-a-box test)
        floor := jolt.NewBoxShape(jolt.Vec3{50, 1, 50}, 0)
        floorSettings := jolt.NewBodyCreationSettings()
        floorSettings.Position = jolt.RVec3{0, -1, 0}
        floorSettings.MotionType = jolt.MotionTypeStatic
        floorSettings.ObjectLayer = jolt.ObjectLayerNonMoving
        floorSettings.SetShape(floor.AsShape())
        bi.CreateAndAddBody(floorSettings, jolt.ActivationDontActivate)

        box := jolt.NewBoxShape(jolt.Vec3{0.5, 0.5, 0.5}, 0)
        boxSettings := jolt.NewBodyCreationSettings()
        boxSettings.Position = jolt.RVec3{0, 5, 0}
        boxSettings.MotionType = jolt.MotionTypeDynamic
        boxSettings.ObjectLayer = jolt.ObjectLayerMoving
        boxSettings.SetShape(box.AsShape())
        bi.CreateAndAddBody(boxSettings, jolt.ActivationActivate)

        const dt = 1.0 / 60.0
        for range 120 {
            ps.Update(dt, 1, ta, js)
        }

        mu.Lock()
        defer mu.Unlock()
        Expect(contacts).To(BeNumerically(">", 0))
    })
})
```

- [ ] **Step 2: Run, expect failure**

This exercises the listener vtable, cgo.Handle round-trip, and Procs init().

- [ ] **Step 3: Fix until it passes**

- [ ] **Step 4: Run, expect pass**

- [ ] **Step 5: Commit**

```bash
git add jolt/integration_test.go
git commit -m "test(jolt): contact-listener integration test passing"
```

---

## Phase 12 — Documentation

### Task 42: Flesh out README

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Write the failing check**

(Documentation task — no automated test. Step 1 is reading the spec and listing what the README must cover.)

Required sections:
- Install / `go get`
- Build prerequisites (C++17 toolchain, ~30–60s first build)
- Quickstart (drop-a-box example, ~25 lines)
- Build tags (`joltdouble`, `jolt_cross_platform_deterministic`)
- Regenerating bindings (`go generate ./jolt`)
- Module layout pointer
- License + acknowledgements (joltc, Jolt)
- Pinned upstream commits (already added in Task 3 — extend as needed)

- [ ] **Step 2: N/A (documentation)**

- [ ] **Step 3: Write the README content**

Use the integration test from Task 40 as the quickstart's basis (trimmed). Mention thread-safety contract for listeners with a one-line summary linking to `package jolt` doc comment.

- [ ] **Step 4: Verify links and code samples compile**

For any code blocks in the README, paste-test them in a scratch file:

```bash
go build ./...
```

Expected: success.

- [ ] **Step 5: Commit**

```bash
git add README.md
git commit -m "docs: flesh out README with quickstart, build prerequisites, regen instructions"
```

---

### Task 43: Final smoke test on darwin/arm64 (or current machine)

**Files:**
- None (verification only)

- [ ] **Step 1: Run the full suite**

```bash
go test ./...
```

Expected: PASS across `generator/`, `internal/cjolt/`, `jolt/`.

- [ ] **Step 2: Run with double precision**

```bash
go test -tags joltdouble ./...
```

Expected: PASS. Math types should still pass; integration tests should still pass after `JPH_DOUBLE_PRECISION` cgo flag flips.

- [ ] **Step 3: Run vet and a quick lint**

```bash
go vet ./...
gofmt -l .
```

Expected: vet clean, gofmt produces no diff.

- [ ] **Step 4: If anything fails, fix and commit; otherwise tag the state**

```bash
git tag mvp-v0.0.1
```

(Optional — confirm with user before pushing tag.)

- [ ] **Step 5: Commit (if any fixes)**

```bash
git status
# if dirty:
git add . && git commit -m "fix: misc fixes from final MVP smoke"
```

---

## Self-review checklist

After implementation completes, verify against the spec (`§12 MVP scope`):

- [ ] Repo skeleton, submodules, go.mod, dependencies → Phase 0
- [ ] Generator (parser, annotations, three emitters) → Phases 2–7
- [ ] Math types and operations (handwritten) → Phase 1
- [ ] Build pipeline (unity stub, cgo directives, build tags) → Phase 3, Task 18
- [ ] `RegisterTypes` / `UnregisterTypes` → Task 18 + Task 35
- [ ] Default broad-phase / object-layer filter setup → Task 40 (drop-a-box), iterating in Phase 11
- [ ] Lifetime patterns: `PhysicsSystem`, `TempAllocator`, `JobSystem` → Tasks 35–36
- [ ] `BodyInterface`, `Body`, `BodyID`, `BodyCreationSettings` (Approach 1) → Phase 6, Tasks 37–38
- [ ] All 17 MVP shape types → Task 39
- [ ] `ContactListener` only (struct-of-funcs + cgo.Handle trampolines) → Tasks 30–31, 41
- [ ] Two MVP integration tests on `linux/amd64` and `darwin/arm64` (CI deferred) → Tasks 40–41

Spec items NOT in MVP (per §13) are explicitly omitted: queries, constraints, body lock, character/vehicle/ragdoll/softbody, debug renderer, StateRecorder, `.syso` prebuilts, custom `OBJECT_LAYER_BITS`. Each is a future slice with its own brainstorm → spec → plan cycle.

## Risks and mitigations (carried forward from spec §15)

- **Layout drift on Approach 1 settings:** Task 28 emits `_Static_assert(sizeof(...))` per type; build breaks instead of memory corruption.
- **joltc init-bug pattern (#66):** `New<Type>Settings()` constructors zero the C struct before calling `_Init` — structural avoidance.
- **Cgo callback overhead:** ContactListener's `func` fields can be left nil; the trampoline early-exits without re-entering Go.
- **First build time (~30–60s):** Documented in README; `.syso` prebuilts deferred to slice 12.
- **joltc upstream tempo:** Pinning to a specific commit; hand-written carve-outs (Phase post-MVP) remain available.
