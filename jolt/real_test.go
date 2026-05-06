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
