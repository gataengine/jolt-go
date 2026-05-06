package generator_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gataengine/jolt-go/generator"
)

var _ = Describe("EmitCgoDirectives", func() {
	It("emits CXXFLAGS and per-arch SIMD flags", func() {
		out := generator.EmitCgoDirectives()
		Expect(out).To(ContainSubstring(`#cgo CXXFLAGS: -std=c++17`))
		Expect(out).To(ContainSubstring(`#cgo CXXFLAGS: -I${SRCDIR}/../../third_party/joltc/include`))
		Expect(out).To(ContainSubstring(`#cgo CXXFLAGS: -I${SRCDIR}/../../third_party/JoltPhysics`))
		Expect(out).To(ContainSubstring(`#cgo CXXFLAGS: -DJPH_OBJECT_LAYER_BITS=32`))
		Expect(out).To(ContainSubstring(`#cgo CFLAGS: -I${SRCDIR}/../../third_party/joltc/include`))
		Expect(out).To(ContainSubstring(`#cgo amd64 CXXFLAGS: -DJPH_USE_SSE4_1 -DJPH_USE_SSE4_2 -DJPH_USE_AVX`))
		Expect(out).To(ContainSubstring(`#cgo arm64 CXXFLAGS: -DJPH_USE_NEON`))
		Expect(out).To(ContainSubstring(`#cgo joltdouble CXXFLAGS: -DJPH_DOUBLE_PRECISION`))
		Expect(out).To(ContainSubstring(`#cgo LDFLAGS: -lstdc++ -lm`))
		Expect(out).To(ContainSubstring(`package cjolt`))
	})
})
