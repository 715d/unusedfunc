// Package assembly provides functionality for scanning and analyzing assembly files.
// to detect function implementations and calls.
package assembly

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Info contains information about functions found in assembly files.
type Info struct {
	// ImplementedFunctions maps function names to whether they are implemented in assembly.
	ImplementedFunctions map[string]struct{}
	// CalledFunctions maps function names to whether they are called from assembly.
	CalledFunctions map[string]struct{}
}

// Compile patterns once at package initialization.
var (
	// TEXT directive pattern: TEXT ·functionName(SB)
	// The · (middle dot) is used in Go assembly to denote package-level symbols.
	textPattern = regexp.MustCompile(`TEXT\s+·([a-zA-Z_][a-zA-Z0-9_]*)\(SB\)`)

	// CALL directive pattern: CALL ·functionName(SB)
	callPattern = regexp.MustCompile(`CALL\s+·([a-zA-Z_][a-zA-Z0-9_]*)\(SB\)`)
)

// ScanPackage scans all assembly files in a package and returns assembly information.
// This function uses pkg.OtherFiles which are already filtered by the build configuration.
// used during package loading, ensuring correct cross-architecture behavior.
func ScanPackage(pkg *packages.Package) (*Info, error) {
	info := &Info{
		ImplementedFunctions: make(map[string]struct{}),
		CalledFunctions:      make(map[string]struct{}),
	}

	// Handle nil package.
	if pkg == nil {
		return info, nil
	}

	// Use pkg.OtherFiles which contains non-Go files including assembly files.
	// that match the build constraints used when loading the package
	for _, file := range pkg.OtherFiles {
		if !strings.HasSuffix(file, ".s") {
			continue
		}
		if err := scanFile(file, info); err != nil {
			return info, fmt.Errorf("scan assembly file: %s: %w", file, err)
		}
	}

	return info, nil
}

// scanFile scans a single assembly file for function implementations and calls
func scanFile(filename string, info *Info) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return scanReader(file, info)
}

// scanReader scans an io.Reader for assembly directives
func scanReader(r io.Reader, info *Info) error {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments and empty lines.
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}

		// Check for TEXT directive (function implementation)
		if matches := textPattern.FindStringSubmatch(line); matches != nil {
			funcName := matches[1]
			info.ImplementedFunctions[funcName] = struct{}{}
		}

		// Check for CALL directive (function call)
		if matches := callPattern.FindStringSubmatch(line); matches != nil {
			funcName := matches[1]
			info.CalledFunctions[funcName] = struct{}{}
		}
	}

	return scanner.Err()
}
