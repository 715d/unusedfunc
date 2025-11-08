//go:build !linux && !darwin && !freebsd

package main

import "fmt"

// UnixInit stub for non-Unix platforms
func UnixInit() {
	fmt.Println("Unix stub - should not be called on non-Unix")
}
