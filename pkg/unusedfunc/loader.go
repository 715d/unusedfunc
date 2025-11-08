// Package unusedfunc provides unused function/method analysis.
package unusedfunc

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"
)

// defaultLoadMode specifies the standard packages.Mode flags used throughout
// the project for loading Go packages with all necessary information for analysis.
// Note: NeedTypesInfo is required for SSA construction but causes significant
// memory allocation (2GB+ for recordTypeAndValue). This is unavoidable for
// accurate SSA-based analysis.
const defaultLoadMode = packages.NeedDeps |
	packages.NeedName |
	packages.NeedFiles |
	packages.NeedCompiledGoFiles |
	packages.NeedImports |
	packages.NeedTypes |
	packages.NeedSyntax |
	packages.NeedTypesInfo

// LoaderOptions configures package loading behavior.
type LoaderOptions struct {
	// Packages are the package patterns to load.
	Packages []string

	// BuildTags are build tags to apply during loading.
	BuildTags []string

	// Dir is the directory to load packages from.
	// If empty, uses the current working directory.
	Dir string

	// Env is the environment to use for loading.
	// If nil, uses a copy of os.Environ() with CGO_ENABLED=0.
	Env []string
}

// LoadPackages loads Go packages with consistent configuration for unusedfunc analysis.
func LoadPackages(ctx context.Context, opts LoaderOptions) ([]*packages.Package, error) {
	// Default to current directory patterns.
	patterns := opts.Packages
	if len(patterns) == 0 {
		patterns = []string{"./..."}
	}

	cfg := &packages.Config{
		Context: ctx,
		Mode:    defaultLoadMode,
		Tests:   true, // Always load test files to detect usage from tests
		Env:     opts.Env,
	}

	if opts.Dir != "" {
		cfg.Dir = opts.Dir
	}

	if len(opts.BuildTags) > 0 {
		cfg.BuildFlags = append(cfg.BuildFlags, "-tags", strings.Join(opts.BuildTags, ","))
	}

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, fmt.Errorf("loading packages: %w", err)
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages found matching patterns: %v", patterns)
	}

	// Check for errors in loaded packages.
	var errorMessages []string
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			for _, err := range pkg.Errors {
				errorMsg := fmt.Sprintf("package %s: %v", pkg.PkgPath, err)
				errorMessages = append(errorMessages, errorMsg)
			}
		}
	}

	if len(errorMessages) > 0 {
		return nil, fmt.Errorf("package errors:\n%s", strings.Join(errorMessages, "\n"))
	}

	return deduplicatePackages(pkgs), nil
}

// deduplicatePackages removes duplicate packages, preferring test variants over regular packages.
// Test variants (IDs containing "[...]") are supersets that include all production code plus.
// test-only exports, so they should be preferred to avoid analyzing the same functions twice.
func deduplicatePackages(pkgs []*packages.Package) []*packages.Package {
	best := make(map[string]*packages.Package)
	for _, pkg := range pkgs {
		if strings.HasSuffix(pkg.ID, ".test") && !strings.Contains(pkg.ID, "[") {
			continue
		}

		existing, exists := best[pkg.PkgPath]
		if !exists {
			best[pkg.PkgPath] = pkg
			continue
		}

		// Replace with package if it's a superset of the existing one.
		if isSuperset(pkg, existing) {
			best[pkg.PkgPath] = pkg
		}
	}
	return slices.Collect(maps.Values(best))
}

// isSuperset returns true if pkg is a superset of existing.
// Test variants (containing "[...]" in ID) are supersets of regular packages.
// because they include all production code plus test-only exports.
func isSuperset(pkg, existing *packages.Package) bool {
	pkgIsTest := strings.Contains(pkg.ID, "[")
	existingIsTest := strings.Contains(existing.ID, "[")

	// Test variant is a superset of regular package.
	if pkgIsTest && !existingIsTest {
		return true
	}

	// Regular package is not a superset of test variant.
	if !pkgIsTest && existingIsTest {
		return false
	}

	// Both are same type (both test or both regular), neither is a superset.
	return false
}
