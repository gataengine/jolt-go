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
