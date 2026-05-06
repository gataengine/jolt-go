package main

import (
	"fmt"
	"log"
	"os"

	gen "github.com/gataengine/jolt-go/generator"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: gen <command> [args...]\n  commands: emit-cgo, emit-unity")
	}
	switch os.Args[1] {
	case "emit-cgo":
		fmt.Print(gen.EmitCgoDirectives())
	case "emit-unity":
		if len(os.Args) != 4 {
			log.Fatal("usage: gen emit-unity <jolt-cmake-path> <joltc-source-list-comma-separated>")
		}
		joltSources, err := gen.ExtractJoltSources(os.Args[2])
		if err != nil {
			log.Fatalf("extract: %v", err)
		}
		// Parse joltc sources from arg 3 (comma-separated)
		joltcSources := []string{}
		for _, p := range splitCSV(os.Args[3]) {
			joltcSources = append(joltcSources, p)
		}
		fmt.Print(gen.EmitUnity(joltSources, joltcSources))
	default:
		log.Fatalf("unknown command: %s", os.Args[1])
	}
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	out := []string{}
	cur := ""
	for _, c := range s {
		if c == ',' {
			out = append(out, cur)
			cur = ""
			continue
		}
		cur += string(c)
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}
