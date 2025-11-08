// Package build_constraints tests various build constraint combinations
package main

import "fmt"

// Common functions available on all platforms

// CommonFunc is used on all platforms - USED
func CommonFunc() {
	fmt.Println("Common function")
}

// CommonHelper assists common operations - USED
func CommonHelper(s string) string {
	return "common: " + s
}

// UnusedCommon is not used on any platform - UNUSED
func UnusedCommon() {
	fmt.Println("This common function is unused")
}

// Suppressed common function
//
//nolint:unusedfunc
func SuppressedCommon() {
	fmt.Println("Suppressed common function")
}
