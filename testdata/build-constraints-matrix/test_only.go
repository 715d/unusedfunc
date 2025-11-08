//go:build test

package main

import "fmt"

// Test-only helper functions

// TestHelper assists in testing - USED in tests when test tag
func TestHelper(name string) {
	fmt.Printf("Test helper: %s\n", name)
}

// UnusedTestHelper is not used even in tests - UNUSED
func UnusedTestHelper() {
	fmt.Println("Unused test helper")
}
