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
