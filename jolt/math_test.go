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
