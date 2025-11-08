// Package linkname tests //go:linkname directive handling
package main

import (
	_ "unsafe" // Required for go:linkname
)

// These functions are linked to runtime or other packages
// They should NOT be reported as unused even though they appear unused locally

//go:linkname linkedToRuntime runtime.fastrand
func linkedToRuntime() uint32

//go:linkname linkedToInternal internal/bytealg.IndexString
func linkedToInternal(s string, substr string) int

// This function is exposed to another package via linkname
//
//go:linkname exposedViaLinkname
func exposedViaLinkname() string {
	return "exposed via linkname"
}

// This function is linked from another file in the same package
//
//go:linkname internalLinked
func internalLinked() int {
	return 42
}

// Regular unused function - SHOULD be reported
func unusedRegular() string {
	return "I am unused"
}

// Function used normally
func usedNormally() string {
	return "I am used"
}

// Suppressed via nolint
//
//nolint:unusedfunc
func suppressedFunction() string {
	return "I am suppressed"
}

// This is actually unused but has a misleading comment
// go:linkname notReallyLinked
func notReallyLinked() string {
	return "The comment above is not a valid directive (has space)"
}

// Complex linkname scenario - linked to unexported symbol
//
//go:linkname complexLinked some/package.unexportedFunc
func complexLinked(x, y int) int {
	return x + y
}

// Function that calls linked functions
func usesSome() {
	// Call the normally used function
	result := usedNormally()
	println(result)
}

func main() {
	usesSome()
}
