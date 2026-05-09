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
