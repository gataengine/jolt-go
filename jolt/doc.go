// Package jolt is an idiomatic Go binding for JoltPhysics.
//
// Thread safety: physics callbacks (e.g. ContactListener) are invoked from
// Jolt's worker threads. Callers are responsible for synchronizing access to
// state shared with their listener implementations. The binding does not
// serialize callbacks.
package jolt

//go:generate go run ../generator/cmd/gen
