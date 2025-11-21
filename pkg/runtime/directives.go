// Package runtime provides functionality for detecting and handling.
// Go runtime-specific directives and functions.
package runtime

import (
	"go/ast"
	"strings"
)

// DirectiveType represents different types of Go compiler directives.
type DirectiveType int

const (
	DirectiveNone DirectiveType = iota
	DirectiveNosplit
	DirectiveNoinline
	DirectiveNorace
	DirectiveNocheckptr
	DirectiveLinkname
	DirectiveCGoExport // CGo export directive
)

// DirectiveInfo contains information about a runtime directive found on a function.
type DirectiveInfo struct {
	Type      DirectiveType
	Directive string
	Valid     bool
}

// runtimeDirectives maps directive strings to their types
var runtimeDirectives = map[string]DirectiveType{
	"go:nosplit":    DirectiveNosplit,
	"go:noinline":   DirectiveNoinline,
	"go:norace":     DirectiveNorace,
	"go:nocheckptr": DirectiveNocheckptr,
	"go:linkname":   DirectiveLinkname,
}

// runtimeHookFunctions contains function names that are known runtime hooks
// These functions may be called by the Go runtime even if not explicitly called in code.
var runtimeHookFunctions = map[string]bool{
	"mallocHook":      true,
	"freeHook":        true,
	"gcCallback":      true,
	"runGCCallbacks":  true,
	"panicHook":       true,
	"recoverHook":     true,
	"scheduleHook":    true,
	"preemptHook":     true,
	"sighandler":      true,
	"cpuProfilerHook": true,
	"memprofHook":     true,
	"uintptrEscapes":  true,
	"allocNotInHeap":  true,
}

// HasRuntimeDirective checks if a function declaration has any runtime directive.
func HasRuntimeDirective(fn *ast.FuncDecl) *DirectiveInfo {
	if fn.Doc == nil || len(fn.Doc.List) == 0 {
		return &DirectiveInfo{Type: DirectiveNone, Valid: false}
	}

	// Check each comment in the function's doc comment group.
	for _, comment := range fn.Doc.List {
		if directive := parseDirective(comment.Text); directive.Type != DirectiveNone {
			return directive
		}
	}

	return &DirectiveInfo{Type: DirectiveNone, Valid: false}
}

// parseDirective parses a comment to check if it contains a valid runtime directive
func parseDirective(comment string) *DirectiveInfo {
	// Remove leading "//" and any spaces.
	text := strings.TrimPrefix(comment, "//")

	// Check for CGo export directive first.
	// CGo export has format "//export FuncName" (note: no colon after //)
	if strings.HasPrefix(text, "export ") {
		return &DirectiveInfo{
			Type:      DirectiveCGoExport,
			Directive: "export",
			Valid:     true,
		}
	}

	// If there's a space immediately after "//" then it's not a valid directive.
	// Valid directives must be exactly "//go:" with no space.
	if strings.HasPrefix(comment, "// ") {
		return &DirectiveInfo{Type: DirectiveNone, Valid: false}
	}

	// Check for each known runtime directive.
	for directive, directiveType := range runtimeDirectives {
		if after, ok := strings.CutPrefix(text, directive); ok {
			// Ensure it's exactly the directive or followed by space/args.
			if after == "" || strings.HasPrefix(after, " ") {
				return &DirectiveInfo{
					Type:      directiveType,
					Directive: directive,
					Valid:     true,
				}
			}
		}
	}

	return &DirectiveInfo{Type: DirectiveNone, Valid: false}
}

// IsRuntimeHookFunction checks if a function name is a known runtime hook.
func IsRuntimeHookFunction(name string) bool {
	return runtimeHookFunctions[name]
}
