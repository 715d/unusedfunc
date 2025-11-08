// Package assembly tests assembly file integration
package main

// These functions are implemented in assembly files
// They should NOT be reported as unused

// Add performs addition - implemented in math_amd64.s
func Add(x, y int) int

// Subtract performs subtraction - implemented in math_amd64.s
func Subtract(x, y int) int

// Multiply is implemented in Go but has assembly optimization
func Multiply(x, y int) int {
	// Fallback Go implementation
	// Assembly version in math_amd64.s will be used on amd64
	return x * y
}

// These are helper functions used by assembly

// helperFunc is called from assembly code
func helperFunc(x int) int {
	return x * 2
}

// asmCallback is passed to assembly as a function pointer
func asmCallback(data []byte) {
	println("Callback from assembly with", len(data), "bytes")
}

// Regular Go functions

// UsedFunction demonstrates usage
func UsedFunction() {
	a := Add(10, 20)
	b := Subtract(30, 15)
	c := Multiply(5, 6)
	println(a, b, c)
}

// UnusedFunction should be reported as unused
func UnusedFunction() int {
	return 42
}

// Suppressed function
//
//nolint:unusedfunc
func SuppressedAsm() {
	println("Used in special assembly bootstrap")
}

// Function with no body but not assembly - should be reported
func EmptyNotAsm() int

// Complex assembly integration

// VectorAdd performs SIMD vector addition - implemented in math_amd64.s
func VectorAdd(dst, a, b []float64)

// MatrixMultiply uses assembly for performance
func MatrixMultiply(c, a, b [][]float64) {
	// Stub - actual implementation in assembly
}

func main() {
	UsedFunction()

	// Test vector operations
	v1 := []float64{1.0, 2.0, 3.0, 4.0}
	v2 := []float64{5.0, 6.0, 7.0, 8.0}
	result := make([]float64, 4)
	VectorAdd(result, v1, v2)
}
