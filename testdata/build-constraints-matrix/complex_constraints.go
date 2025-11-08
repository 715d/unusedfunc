//go:build (linux || darwin) && !windows && (amd64 || arm64) && !mips

package main

import "fmt"

// Complex build constraint combinations

// ComplexPlatformInit initializes on specific platform combinations - USED
func ComplexPlatformInit() {
	fmt.Println("Complex platform initialization")
	setupComplexFeatures()
}

// setupComplexFeatures sets up platform-specific features - USED
func setupComplexFeatures() {
	fmt.Println("Setting up complex features")
}

// ComplexOptimization performs complex optimizations - USED in benchmarks
func ComplexOptimization(data []byte) []byte {
	// Platform-specific optimization
	return append([]byte("optimized: "), data...)
}

// UnusedComplex is not used on these platforms - UNUSED
func UnusedComplex() {
	fmt.Println("Unused complex constraint function")
}
