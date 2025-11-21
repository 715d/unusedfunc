package analysis

import (
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

// TestFuncInfo_NewFuncInfo tests the creation of new FuncInfo instances.
func TestFuncInfo_NewFuncInfo(t *testing.T) {
	tests := []struct {
		name             string
		setupPackage     func() (*packages.Package, types.Object)
		expectedExported bool
		expectedInternal bool
	}{
		{
			name: "exported func in regular package",
			setupPackage: func() (*packages.Package, types.Object) {
				pkg := &packages.Package{
					ID:      "example.com/test",
					PkgPath: "example.com/test",
				}
				// Create a func object for exported func.
				obj := types.NewFunc(token.NoPos, pkg.Types, "ExportedFunc", nil)
				return pkg, obj
			},
			expectedExported: true,
			expectedInternal: false,
		},
		{
			name: "unexported func in regular package",
			setupPackage: func() (*packages.Package, types.Object) {
				pkg := &packages.Package{
					ID:      "example.com/test",
					PkgPath: "example.com/test",
				}
				obj := types.NewFunc(token.NoPos, pkg.Types, "unexportedFunc", nil)
				return pkg, obj
			},
			expectedExported: false,
			expectedInternal: false,
		},
		{
			name: "exported func in internal package",
			setupPackage: func() (*packages.Package, types.Object) {
				pkg := &packages.Package{
					ID:      "example.com/test/internal/handler",
					PkgPath: "example.com/test/internal/handler",
				}
				obj := types.NewFunc(token.NoPos, pkg.Types, "ExportedFunc", nil)
				return pkg, obj
			},
			expectedExported: true,
			expectedInternal: true,
		},
		{
			name: "unexported func in internal package",
			setupPackage: func() (*packages.Package, types.Object) {
				pkg := &packages.Package{
					ID:      "example.com/test/internal/handler",
					PkgPath: "example.com/test/internal/handler",
				}
				obj := types.NewFunc(token.NoPos, pkg.Types, "unexportedFunc", nil)
				return pkg, obj
			},
			expectedExported: false,
			expectedInternal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg, obj := tt.setupPackage()
			nameCache := NewNameCache()
			f := NewFuncInfo(obj, pkg, nameCache, false)

			require.NotNil(t, f, "NewFuncInfo returned nil")

			require.Equal(t, obj, f.Object)

			require.Equal(t, tt.expectedExported, f.IsExported)

			require.Equal(t, tt.expectedInternal, f.IsInInternal)

			require.Equal(t, pkg, f.Package)

			// IsUsed should be false by default.
			require.False(t, f.IsUsed, "Expected IsUsed to be false by default")
		})
	}
}

// TestFuncInfo_IsInInternalPackage tests internal package detection.
func TestFuncInfo_IsInInternalPackage(t *testing.T) {
	tests := []struct {
		name     string
		pkgPath  string
		expected bool
	}{
		{"regular package", "example.com/test", false},
		{"internal package", "example.com/test/internal", true},
		{"nested internal package", "example.com/test/internal/handler", true},
		{"package with internal in name but not path", "example.com/testinternal", false},
		{"deep internal package", "example.com/test/internal/deep/nested", true},
		{"standard library", "fmt", false},
		{"standard library internal", "internal/abi", true},
		{"vendor internal", "vendor/example.com/lib/internal", true},
		{"go modules internal", "go.mod.internal", false},
		{"internal at root", "internal", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg := &packages.Package{
				ID:      tt.pkgPath,
				PkgPath: tt.pkgPath,
			}
			obj := types.NewFunc(token.NoPos, pkg.Types, "TestFunc", nil)
			f := NewFuncInfo(obj, pkg, NewNameCache(), false)

			result := f.IsInInternalPackage()
			require.Equal(t, tt.expected, result, "Expected IsInInternalPackage() to return %v for package %s", tt.expected, tt.pkgPath)

			// Also test the field is set correctly.
			require.Equal(t, tt.expected, f.IsInInternal, "Expected IsInInternal field to be %v for package %s", tt.expected, tt.pkgPath)
		})
	}
}

// TestFuncInfo_ShouldReport tests the reporting logic.
func TestFuncInfo_ShouldReport(t *testing.T) {
	tests := []struct {
		name        string
		isExported  bool
		isInternal  bool
		isUsed      bool
		expected    bool
		description string
	}{
		{
			name:        "unused unexported function in regular package",
			isExported:  false,
			isInternal:  false,
			isUsed:      false,
			expected:    true,
			description: "should report unused unexported funcs",
		},
		{
			name:        "used unexported function in regular package",
			isExported:  false,
			isInternal:  false,
			isUsed:      true,
			expected:    false,
			description: "should not report used unexported funcs",
		},
		{
			name:        "unused exported function in regular package",
			isExported:  true,
			isInternal:  false,
			isUsed:      false,
			expected:    false,
			description: "should not report unused exported funcs in regular packages",
		},
		{
			name:        "used exported function in regular package",
			isExported:  true,
			isInternal:  false,
			isUsed:      true,
			expected:    false,
			description: "should not report used exported funcs in regular packages",
		},
		{
			name:        "unused unexported function in internal package",
			isExported:  false,
			isInternal:  true,
			isUsed:      false,
			expected:    true,
			description: "should report unused unexported funcs in internal packages",
		},
		{
			name:        "used unexported function in internal package",
			isExported:  false,
			isInternal:  true,
			isUsed:      true,
			expected:    false,
			description: "should not report used unexported funcs in internal packages",
		},
		{
			name:        "unused exported function in internal package",
			isExported:  true,
			isInternal:  true,
			isUsed:      false,
			expected:    true,
			description: "should report unused exported funcs in internal packages",
		},
		{
			name:        "used exported function in internal package",
			isExported:  true,
			isInternal:  true,
			isUsed:      true,
			expected:    false,
			description: "should not report used exported funcs in internal packages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgPath := "example.com/test"
			if tt.isInternal {
				pkgPath = "example.com/test/internal"
			}

			pkg := &packages.Package{
				ID:      pkgPath,
				PkgPath: pkgPath,
			}

			funcName := "testFunc"
			if tt.isExported {
				funcName = "TestFunc"
			}

			obj := types.NewFunc(token.NoPos, pkg.Types, funcName, nil)
			f := NewFuncInfo(obj, pkg, NewNameCache(), false)
			f.IsUsed = tt.isUsed

			result := f.ShouldReport()
			require.Equal(t, tt.expected, result, "Expected ShouldReport() to return %v. %s", tt.expected, tt.description)
		})
	}
}
