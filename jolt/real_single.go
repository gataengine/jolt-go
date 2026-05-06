//go:build !joltdouble

package jolt

// Real is the scalar type used by world-space coordinates.
//
// Defaults to float32. Build with -tags joltdouble to switch to float64,
// which also enables JPH_DOUBLE_PRECISION in the cgo layer.
type Real = float32
