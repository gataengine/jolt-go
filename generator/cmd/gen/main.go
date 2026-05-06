package main

import (
	"fmt"
	"log"
	"os"

	gen "github.com/gataengine/jolt-go/generator"
)

var joltcUnitySources = []string{
	"src/joltc.cpp",
	"src/joltc_assert.cpp",
	// joltc.c is empty (4 lines, just #include) — skipped.
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: gen --emit-cgo | --emit-unity <jolt.cmake-path>")
	}
	switch os.Args[1] {
	case "--emit-cgo":
		fmt.Print(gen.EmitCgoDirectives())
	case "--emit-unity":
		if len(os.Args) < 3 {
			log.Fatal("usage: gen --emit-unity <jolt.cmake-path>")
		}
		joltSources, err := gen.ExtractJoltSources(os.Args[2])
		if err != nil {
			log.Fatalf("extract jolt sources: %v", err)
		}
		fmt.Print(gen.EmitUnity(joltSources, joltcUnitySources))
	default:
		log.Fatalf("unknown mode: %s", os.Args[1])
	}
}
