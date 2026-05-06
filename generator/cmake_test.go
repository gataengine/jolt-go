package generator_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gataengine/jolt-go/generator"
)

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
