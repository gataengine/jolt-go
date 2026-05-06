package generator_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gataengine/jolt-go/generator"
)

var _ = Describe("Annotations", func() {
	It("declares default sidecar values", func() {
		a := generator.Annotations
		Expect(a.RefCounted).To(ContainElement("Shape"))
		Expect(a.Inheritance["BoxShape"]).To(Equal("ConvexShape"))
		// BodyCreationSettings is opaque in joltc (Create/Destroy + accessor
		// methods, no public field layout).  It must NOT appear in SettingsInit
		// which is the layout-mirror (Approach 1) map.
		Expect(a.SettingsInit).NotTo(HaveKey("BodyCreationSettings"))
	})
	It("flags listener procs structs", func() {
		l := findListener(generator.Annotations.Listeners, "JPH_ContactListener")
		Expect(l.Methods).To(HaveLen(4))
	})
})

func findListener(ls []generator.ListenerSpec, c string) generator.ListenerSpec {
	for _, l := range ls {
		if l.CType == c {
			return l
		}
	}
	Fail("listener not found: " + c)
	return generator.ListenerSpec{}
}
