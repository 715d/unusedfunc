//go:build !amd64
// +build !amd64

package main

// Stub implementations for non-amd64 architectures

func Add(x, y int) int {
	return x + y
}

func Subtract(x, y int) int {
	return x - y
}

// Note: Multiply already has a Go implementation in math.go

func VectorAdd(dst, a, b []float64) {
	if len(dst) != len(a) || len(dst) != len(b) {
		panic("vector length mismatch")
	}
	for i := range dst {
		dst[i] = a[i] + b[i]
	}
}

// This demonstrates platform-specific assembly
// On amd64, the assembly version is used
// On other platforms, this Go version is used
func PlatformSpecific() string {
	return "Go implementation"
}
