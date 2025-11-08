//go:build debug

package main

import "fmt"

// Debug-only functions

// DebugInit initializes debug mode - USED when debug tag
func DebugInit() {
	fmt.Println("Debug mode enabled")
	enableDebugLogging()
	setupDebugHandlers()
}

// enableDebugLogging enables verbose logging - USED
func enableDebugLogging() {
	fmt.Println("Debug logging enabled")
}

// setupDebugHandlers sets up debug handlers - USED
func setupDebugHandlers() {
	fmt.Println("Debug handlers configured")
}

// DebugDump dumps debug information - USED in debug mode
func DebugDump(data interface{}) {
	fmt.Printf("DEBUG: %+v\n", data)
}

// UnusedDebug is not used even in debug mode - UNUSED
func UnusedDebug() {
	fmt.Println("Unused debug function")
}
