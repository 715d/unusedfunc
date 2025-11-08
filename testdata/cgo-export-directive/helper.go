package main

// #include <string.h>
import "C"
import "unsafe"

// Additional CGo exported functions in a separate file

//export GoStrlen
func GoStrlen(s *C.char) C.size_t {
	return C.strlen(s)
}

//export GoMemcpy
func GoMemcpy(dst, src unsafe.Pointer, n C.size_t) unsafe.Pointer {
	return C.memcpy(dst, src, n)
}

// Helper functions that might be used by exported functions

// CStringToGoString converts C string to Go string - used by multiple exports
func CStringToGoString(cs *C.char) string {
	return C.GoString(cs)
}

// GoStringToCString converts Go string to C string
func GoStringToCString(s string) *C.char {
	return C.CString(s)
}

// UnusedCGoHelper should be reported as unused
func UnusedCGoHelper() {
	println("This CGo helper is not used")
}

// Complex export with error handling

//export GoComplexOperation
func GoComplexOperation(input *C.char, output **C.char) C.int {
	if input == nil {
		return -1
	}

	goStr := CStringToGoString(input)
	result := complexProcess(goStr)

	*output = GoStringToCString(result)
	return C.int(len(result))
}

// Used by GoComplexOperation
func complexProcess(input string) string {
	return "Complex: " + input
}

// Callback registration pattern

type CallbackFunc func(int)

var registeredCallbacks []CallbackFunc

//export GoRegisterCallback
func GoRegisterCallback() {
	// Register a Go function to be called from C
	registeredCallbacks = append(registeredCallbacks, internalCallback)
}

// Used via callback registration
func internalCallback(value int) {
	println("Internal callback called with:", value)
}

//export GoTriggerCallbacks
func GoTriggerCallbacks(value C.int) {
	for _, cb := range registeredCallbacks {
		cb(int(value))
	}
}
