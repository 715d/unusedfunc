//go:build debug

package main

import "fmt"

// ReleaseInit stub for debug builds
func ReleaseInit() {
	fmt.Println("Release stub - should not be called in debug builds")
}

// ReleaseCheck stub for debug builds
func ReleaseCheck() bool {
	fmt.Println("Release check stub - should not be called in debug builds")
	return false
}
