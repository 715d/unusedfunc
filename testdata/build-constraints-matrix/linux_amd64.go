//go:build linux && amd64

package main

import "fmt"

// Linux AMD64 specific optimizations

// LinuxAMD64Optimize performs platform-specific optimization - USED
func LinuxAMD64Optimize() {
	fmt.Println("Linux AMD64 optimizations")
	useAVX2()
}

// useAVX2 uses AVX2 instructions - USED
func useAVX2() {
	fmt.Println("Using AVX2 instructions")
}

// LinuxAMD64Special provides special functionality - USED in tests
func LinuxAMD64Special() string {
	return "Linux AMD64 special"
}

// UnusedLinuxAMD64 is not used - UNUSED
func UnusedLinuxAMD64() {
	fmt.Println("Unused Linux AMD64 function")
}
