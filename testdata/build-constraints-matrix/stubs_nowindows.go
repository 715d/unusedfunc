//go:build !windows

package main

import "fmt"

// WindowsInit stub for non-Windows platforms
func WindowsInit() {
	fmt.Println("Windows stub - should not be called on non-Windows")
}
