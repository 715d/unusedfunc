//go:build !test

package main

import (
	"fmt"
	"runtime"
)

// Platform initialization based on build constraints
func initPlatform() {
	CommonFunc()
	fmt.Printf("Initializing for %s/%s\n", runtime.GOOS, runtime.GOARCH)

	// Platform-specific initialization
	switch runtime.GOOS {
	case "linux", "darwin", "freebsd":
		UnixInit()
		if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
			LinuxAMD64Optimize()
		}
	case "windows":
		WindowsInit()
	}

	// Complex constraint platforms
	if (runtime.GOOS == "linux" || runtime.GOOS == "darwin") &&
		(runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64") {
		ComplexPlatformInit()
	}
}

// Initialize based on debug/release
func initMode() {
	// One of these will be available based on build tags
	if isDebugBuild() {
		DebugInit()
	} else {
		ReleaseInit()
	}
}

// isDebugBuild checks if this is a debug build
func isDebugBuild() bool {
	// This would be determined by build tags
	return false
}

// Generic helper used across platforms - USED
func genericHelper() string {
	return CommonHelper("generic")
}

// UnusedMain is not used in main - UNUSED
func UnusedMain() {
	fmt.Println("Unused in main")
}

func main() {
	initPlatform()
	initMode()

	result := genericHelper()
	fmt.Println(result)

	// Use release check
	if ReleaseCheck() {
		fmt.Println("Release checks passed")
	}
}
