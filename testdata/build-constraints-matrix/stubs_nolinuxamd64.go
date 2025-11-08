//go:build !linux || !amd64

package main

import "fmt"

// LinuxAMD64Optimize stub for non-Linux AMD64 platforms
func LinuxAMD64Optimize() {
	fmt.Println("Linux AMD64 stub - should not be called on non-Linux AMD64")
}
