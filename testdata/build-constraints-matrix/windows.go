//go:build windows

package main

import "fmt"

// Windows-specific functions

// WindowsInit initializes Windows systems - USED on Windows
func WindowsInit() {
	fmt.Println("Windows initialization")
	windowsSetup()
}

// windowsSetup performs Windows setup - USED
func windowsSetup() {
	configureWindowsRegistry()
	setupWindowsPaths()
}

// configureWindowsRegistry configures registry - USED
func configureWindowsRegistry() {
	fmt.Println("Configuring Windows registry")
}

// setupWindowsPaths sets up Windows paths - USED
func setupWindowsPaths() {
	fmt.Println("Setting up Windows paths")
}

// WindowsOnly is only available on Windows - USED in windows_test.go
func WindowsOnly() string {
	return "Windows only function"
}

// UnusedWindows is not used on Windows - UNUSED
func UnusedWindows() {
	fmt.Println("Unused Windows function")
}
