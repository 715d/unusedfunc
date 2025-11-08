// Package cgo tests CGo exported functions
package main

// #include <stdlib.h>
// #include <stdio.h>
//
// extern void GoCallback(int);
//
// static void callGoCallback(int value) {
//     GoCallback(value);
// }
//
// typedef struct {
//     int x;
//     int y;
// } Point;
import "C"
import (
	"fmt"
	"unsafe"
)

// These functions are exported to C and should NOT be reported as unused

//export GoAdd
func GoAdd(a, b C.int) C.int {
	return a + b
}

//export GoSubtract
func GoSubtract(a, b C.int) C.int {
	return a - b
}

//export GoCallback
func GoCallback(value C.int) {
	fmt.Printf("Callback from C with value: %d\n", int(value))
}

//export GoProcessPoint
func GoProcessPoint(p *C.Point) C.int {
	return p.x + p.y
}

//export GoAllocateMemory
func GoAllocateMemory(size C.size_t) unsafe.Pointer {
	return C.malloc(size)
}

//export GoFreeMemory
func GoFreeMemory(ptr unsafe.Pointer) {
	C.free(ptr)
}

// Regular Go functions

// UsedByCGo is called from exported functions
func UsedByCGo(x int) int {
	return x * 2
}

//export GoUseHelper
func GoUseHelper(x C.int) C.int {
	// This exported function uses a regular Go function
	result := UsedByCGo(int(x))
	return C.int(result)
}

// UnusedHelper should be reported as unused
func UnusedHelper(x int) int {
	return x * 3
}

// UsedInGo is used in Go code
func UsedInGo() {
	fmt.Println("This is used in Go code")
}

// Suppressed function
//
//nolint:unusedfunc
func SuppressedCGo() {
	fmt.Println("This is suppressed")
}

// Complex CGo scenarios

//export GoStringOperation
func GoStringOperation(cs *C.char) *C.char {
	goStr := C.GoString(cs)
	result := processString(goStr)
	return C.CString(result)
}

// Used by GoStringOperation
func processString(s string) string {
	return "Processed: " + s
}

//export GoArrayOperation
func GoArrayOperation(arr *C.int, size C.int) C.int {
	// Convert C array to Go slice
	slice := (*[1 << 30]C.int)(unsafe.Pointer(arr))[:size:size]

	sum := C.int(0)
	for i := 0; i < int(size); i++ {
		sum += slice[i]
	}
	return sum
}

// This looks like export but isn't (missing comment)
// export GoNotReallyExported
func GoNotReallyExported() {
	fmt.Println("This should be reported as unused")
}

// Test calling C from Go
func CallCFunctions() {
	// Test the callback
	C.callGoCallback(42)

	// Create a point
	point := C.Point{x: 10, y: 20}
	result := GoProcessPoint(&point)
	fmt.Printf("Point sum: %d\n", result)
}

func main() {
	UsedInGo()
	CallCFunctions()
}
