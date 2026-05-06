package cjolt_test

import (
	"testing"

	"github.com/gataengine/jolt-go/internal/cjolt"
)

func TestCgoLinks(t *testing.T) {
	if !cjolt.JPH_Init() {
		t.Fatal("JPH_Init failed")
	}
	cjolt.JPH_Shutdown()
}
