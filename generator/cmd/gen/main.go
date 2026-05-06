package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	gen "github.com/gataengine/jolt-go/generator"
)

var joltcUnitySources = []string{
	"src/joltc.cpp",
	"src/joltc_assert.cpp",
	// joltc.c is empty (4 lines, just #include) — skipped.
}

func main() {
	if len(os.Args) >= 2 {
		// Single-mode invocations: emit one file to stdout.
		switch os.Args[1] {
		case "--emit-cgo":
			fmt.Print(gen.EmitCgoDirectives())
			return
		case "--emit-unity":
			if len(os.Args) < 3 {
				log.Fatal("usage: gen --emit-unity <jolt.cmake-path>")
			}
			joltSources, err := gen.ExtractJoltSources(os.Args[2])
			if err != nil {
				log.Fatalf("extract jolt sources: %v", err)
			}
			fmt.Print(gen.EmitUnity(joltSources, joltcUnitySources, gen.UnityOpts{
				SettingsSizeAsserts: gen.Annotations.SettingsSizes,
			}))
			return
		default:
			log.Fatalf("unknown mode: %s", os.Args[1])
		}
	}

	// Default: full pipeline. Parse joltc.h, extract Jolt sources, write all 4 files.
	runFullPipeline()
}

func runFullPipeline() {
	repoRoot, err := findRepoRoot()
	if err != nil {
		log.Fatalf("find repo root: %v", err)
	}
	headerPath := filepath.Join(repoRoot, "third_party", "joltc", "include", "joltc.h")
	cmakePath := filepath.Join(repoRoot, "third_party", "JoltPhysics", "Jolt", "Jolt.cmake")

	h, err := gen.ParseHeader(headerPath)
	if err != nil {
		log.Fatalf("parse %s: %v", headerPath, err)
	}
	joltSources, err := gen.ExtractJoltSources(cmakePath)
	if err != nil {
		log.Fatalf("extract jolt sources from %s: %v", cmakePath, err)
	}

	write := func(path, content string) {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			log.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			log.Fatalf("write %s: %v", path, err)
		}
		log.Printf("wrote %s", path)
	}

	write(filepath.Join(repoRoot, "internal", "cjolt", "cgo.go"), gen.EmitCgoDirectives())
	write(filepath.Join(repoRoot, "internal", "cjolt", "cjolt_unity.cpp"),
		gen.EmitUnity(joltSources, joltcUnitySources, gen.UnityOpts{
			SettingsSizeAsserts: gen.Annotations.SettingsSizes,
		}))
	write(filepath.Join(repoRoot, "internal", "cjolt", "cjolt_generated.go"), gen.EmitCjolt(h))
	write(filepath.Join(repoRoot, "jolt", "jolt_generated.go"), gen.EmitJolt(h))
	log.Print("done")
}

// findRepoRoot walks up from cwd looking for go.mod with the jolt-go module path.
// When go generate runs from the jolt/ directory, cwd is jolt/. We need the repo root.
func findRepoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for d := cwd; ; d = filepath.Dir(d) {
		modPath := filepath.Join(d, "go.mod")
		data, err := os.ReadFile(modPath)
		if err == nil && strings.Contains(string(data), "module github.com/gataengine/jolt-go") {
			return d, nil
		}
		if d == filepath.Dir(d) {
			// Reached filesystem root.
			return "", fmt.Errorf("could not find repo root (no go.mod with jolt-go module above %s)", cwd)
		}
	}
}
