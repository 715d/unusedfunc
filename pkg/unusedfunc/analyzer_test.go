package unusedfunc

import (
	"go/types"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"

	"github.com/715d/unusedfunc/internal/analysis"
)

// createTestPackage creates a test package for analyzer tests
func createTestPackage(id, pkgPath string) *packages.Package {
	return &packages.Package{
		ID:      id,
		PkgPath: pkgPath,
		Types:   types.NewPackage(pkgPath, id),
	}
}

// TestAnalyzer_NewAnalyzer tests analyzer creation.
func TestAnalyzer_NewAnalyzer(t *testing.T) {
	analyzer := NewAnalyzer(AnalyzerOptions{})
	require.NotNil(t, analyzer, "NewAnalyzer returned nil")
	require.NotNil(t, analyzer.suppressions, "Expected suppressions to be initialized")
}

// TestAnalyzer_ValidatePackages tests package validation.
func TestAnalyzer_ValidatePackages(t *testing.T) {
	// Package validation is done internally in the Analyze method.
	// This test verifies that invalid packages are handled properly.

	tests := []struct {
		name          string
		setupPkgs     func() []*packages.Package
		expectError   bool
		errorContains string
	}{
		{
			name: "valid packages",
			setupPkgs: func() []*packages.Package {
				return []*packages.Package{
					createTestPackage("test1", "test1"),
					createTestPackage("test2", "test2"),
				}
			},
			expectError: false,
		},
		{
			name: "empty package slice",
			setupPkgs: func() []*packages.Package {
				return []*packages.Package{}
			},
			expectError:   true,
			errorContains: "no packages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewAnalyzer(AnalyzerOptions{})
			pkgs := tt.setupPkgs()

			// Analyze will validate packages internally.
			result, err := analyzer.Analyze(pkgs)

			if tt.expectError {
				require.Error(t, err, "Expected error, got nil")
			} else {
				require.NoError(t, err)
			}

			if !tt.expectError {
				require.NotNil(t, result, "Expected non-nil result")
			}
		})
	}
}

// TestAnalyzer_Analyze tests full analysis workflow.
func TestAnalyzer_Analyze(t *testing.T) {
	tests := []struct {
		name        string
		setupPkgs   func() []*packages.Package
		expectError bool
		checkResult func(*testing.T, map[types.Object]*analysis.FuncInfo)
	}{
		{
			name: "simple analysis",
			setupPkgs: func() []*packages.Package {
				pkg := createTestPackage("test", "test")
				// In a real test, this would have actual AST and type information.
				return []*packages.Package{pkg}
			},
			expectError: false,
			checkResult: func(t *testing.T, funcs map[types.Object]*analysis.FuncInfo) {
				require.NotNil(t, funcs, "Expected functions map to be initialized")
				// NOTE: Without actual code in the test package,
				// the SSA analyzer will analyze an empty package.
			},
		},
		{
			name: "invalid packages",
			setupPkgs: func() []*packages.Package {
				return []*packages.Package{nil}
			},
			expectError: true,
			checkResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewAnalyzer(AnalyzerOptions{})
			pkgs := tt.setupPkgs()

			result, err := analyzer.Analyze(pkgs)

			if tt.expectError {
				require.Error(t, err, "Expected error, got nil")
			} else {
				require.NoError(t, err)
			}

			if tt.checkResult != nil && !tt.expectError {
				tt.checkResult(t, result)
			}
		})
	}
}
