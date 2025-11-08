//go:build !debug

package main

import "fmt"

// DebugInit stub for non-debug builds
func DebugInit() {
	fmt.Println("Debug stub - should not be called in non-debug builds")
}
