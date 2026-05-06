package generator

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// ExtractJoltSources parses a CMake file (or fragment) and returns paths
// listed in the JOLT_PHYSICS_SRC_FILES variable, stripped of the
// ${JOLT_PHYSICS_ROOT}/ prefix.
func ExtractJoltSources(cmakePath string) ([]string, error) {
	f, err := os.Open(cmakePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var (
		inBlock bool
		out     []string
		rx      = regexp.MustCompile(`\$\{JOLT_PHYSICS_ROOT\}/(\S+\.cpp)`)
	)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		switch {
		case strings.Contains(line, "set(JOLT_PHYSICS_SRC_FILES"):
			inBlock = true
		case inBlock && strings.Contains(line, ")"):
			inBlock = false
		case inBlock:
			if m := rx.FindStringSubmatch(line); m != nil {
				out = append(out, m[1])
			}
		}
	}
	return out, sc.Err()
}
