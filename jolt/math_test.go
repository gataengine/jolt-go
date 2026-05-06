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
