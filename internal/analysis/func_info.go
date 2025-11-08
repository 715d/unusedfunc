// Package analysis provides function metadata and classification for unused function detection.
package analysis

import (
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

// FuncInfo represents information about a function in the codebase.
type FuncInfo struct {
	// Object is the types.Object representing this function.
	Object types.Object

	// Name is the display name for this function, including generic type instantiation.
	// For non-generic functions, this is the same as Object.Name()
	// For generic functions/methods, this includes type parameters (e.g., "Container[T].Clear")
	Name string

	// IsUsed indicates whether this function has been found to be used.
	IsUsed bool

	// IsExported indicates whether this function is exported (starts with uppercase)
	IsExported bool

	// IsInInternal indicates whether this function is defined in an internal package.
	IsInInternal bool

	// IsSuppressed indicates whether this function has suppression comments.
	IsSuppressed bool

	// HasLinkname indicates whether this function has a //go:linkname directive.
	HasLinkname bool

	// HasRuntimeDirective indicates whether this function has runtime directives.
	HasRuntimeDirective bool

	// HasAssemblyImplementation indicates whether this function is implemented in assembly.
	HasAssemblyImplementation bool

	// CalledFromAssembly indicates whether this function is called from assembly code.
	CalledFromAssembly bool

	// HasCGoExport indicates whether this function has a //export directive for CGo.
	HasCGoExport bool

	// DeclarationPos is the position where this function is declared.
	DeclarationPos token.Pos

	// Package is the package containing this function.
	Package *packages.Package
}

// NewFuncInfo creates a new FuncInfo for the given function object and package.
func NewFuncInfo(obj types.Object, pkg *packages.Package, nameCache *NameCache) *FuncInfo {
	if obj == nil {
		return &FuncInfo{
			Object:  nil,
			Name:    "",
			Package: pkg,
		}
	}

	fi := &FuncInfo{
		Object:         obj,
		Name:           nameCache.ComputeObjectName(obj),
		IsUsed:         false, // Always start with false
		IsExported:     obj.Exported(),
		Package:        pkg,
		DeclarationPos: obj.Pos(),
	}

	fi.IsInInternal = fi.IsInInternalPackage()
	return fi
}

// IsInInternalPackage checks if this function is defined in an internal package.
func (fi *FuncInfo) IsInInternalPackage() bool {
	if fi.Package == nil {
		return false
	}

	pkgPath := fi.Package.PkgPath

	// Check if the package path contains "internal" as a complete path segment.
	return strings.Contains(pkgPath, "/internal/") ||
		strings.HasSuffix(pkgPath, "/internal") ||
		strings.HasPrefix(pkgPath, "internal/") ||
		pkgPath == "internal"
}

// ShouldReport determines if this function should be reported as unused.
// Returns true if:
// - Method is unexported and unused, OR
// - Method is exported, unused, AND in an internal package
func (fi *FuncInfo) ShouldReport() bool {
	if fi.IsUsed {
		return false
	}

	// Don't report suppressed functions.
	if fi.IsSuppressed {
		return false
	}

	// Don't report functions with linkname directives.
	if fi.HasLinkname {
		return false
	}

	// Don't report functions with runtime directives.
	if fi.HasRuntimeDirective {
		return false
	}

	// Don't report functions with assembly implementations.
	if fi.HasAssemblyImplementation {
		return false
	}

	// Don't report functions called from assembly.
	if fi.CalledFromAssembly {
		return false
	}

	// Don't report CGo exported functions.
	if fi.HasCGoExport {
		return false
	}

	// Report unexported unused functions.
	if !fi.IsExported {
		return true
	}

	// Report exported unused functions if:
	// 1. They're in internal packages, OR
	// 2. They're in a main package (not externally accessible)
	if fi.IsInInternal {
		return true
	}

	// Check if this is in a main package.
	if fi.Package != nil && fi.Package.Name == "main" {
		return true
	}

	return false
}

// FuncSignature represents a function signature for interface matching.
type FuncSignature struct {
	Name       string
	Params     []string // Parameter type names
	Results    []string // Result type names
	IsVariadic bool
}
