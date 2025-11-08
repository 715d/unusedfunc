package unusedfunc

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

func TestDeduplicatePackages(t *testing.T) {
	tests := []struct {
		name     string
		input    []*packages.Package
		expected int // expected number of packages after deduplication
		desc     string
	}{
		{
			name: "regular_and_test_variant",
			input: []*packages.Package{
				{PkgPath: "example.com/pkg", ID: "example.com/pkg"},
				{PkgPath: "example.com/pkg", ID: "example.com/pkg [example.com/pkg.test]"},
			},
			expected: 1,
			desc:     "should keep only test variant when both exist",
		},
		{
			name: "test_binary_filtered",
			input: []*packages.Package{
				{PkgPath: "example.com/pkg", ID: "example.com/pkg"},
				{PkgPath: "example.com/pkg", ID: "example.com/pkg.test"},
			},
			expected: 1,
			desc:     "should filter out test binary, keep regular package",
		},
		{
			name: "external_test_package",
			input: []*packages.Package{
				{PkgPath: "example.com/pkg", ID: "example.com/pkg"},
				{PkgPath: "example.com/pkg_test", ID: "example.com/pkg_test [example.com/pkg.test]"},
			},
			expected: 2,
			desc:     "should keep both regular and external test package",
		},
		{
			name: "only_regular_package",
			input: []*packages.Package{
				{PkgPath: "example.com/pkg", ID: "example.com/pkg"},
			},
			expected: 1,
			desc:     "should keep regular package when no test variant exists",
		},
		{
			name: "only_test_variant",
			input: []*packages.Package{
				{PkgPath: "example.com/pkg", ID: "example.com/pkg [example.com/pkg.test]"},
			},
			expected: 1,
			desc:     "should keep test variant when no regular package exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deduplicatePackages(tt.input)
			require.Len(t, result, tt.expected, tt.desc)

			// Verify no duplicates by ImportPath.
			seen := make(map[string]bool)
			for _, pkg := range result {
				require.False(t, seen[pkg.PkgPath], "duplicate ImportPath found: %s", pkg.PkgPath)
				seen[pkg.PkgPath] = true
			}
		})
	}
}

func TestIsSuperset(t *testing.T) {
	tests := []struct {
		name     string
		pkg      *packages.Package
		existing *packages.Package
		expected bool
		desc     string
	}{
		{
			name:     "test_is_superset_of_regular",
			pkg:      &packages.Package{ID: "pkg [pkg.test]"},
			existing: &packages.Package{ID: "pkg"},
			expected: true,
			desc:     "test variant is superset of regular package",
		},
		{
			name:     "regular_not_superset_of_test",
			pkg:      &packages.Package{ID: "pkg"},
			existing: &packages.Package{ID: "pkg [pkg.test]"},
			expected: false,
			desc:     "regular package is not superset of test variant",
		},
		{
			name:     "both_regular_not_superset",
			pkg:      &packages.Package{ID: "pkg"},
			existing: &packages.Package{ID: "pkg"},
			expected: false,
			desc:     "neither is superset when both regular",
		},
		{
			name:     "both_test_not_superset",
			pkg:      &packages.Package{ID: "pkg [pkg.test]"},
			existing: &packages.Package{ID: "pkg [pkg.test]"},
			expected: false,
			desc:     "neither is superset when both test variants",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSuperset(tt.pkg, tt.existing)
			require.Equal(t, tt.expected, result, tt.desc)
		})
	}
}
