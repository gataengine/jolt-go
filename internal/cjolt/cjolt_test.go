package cjolt_test

import (
	"testing"

	"github.com/gataengine/jolt-go/internal/cjolt"
)

func TestCgoLinks(t *testing.T) {
	cjolt.RegisterTypes()
	cjolt.UnregisterTypes()
}
