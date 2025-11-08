//go:build linux || darwin || freebsd

package main

import "fmt"

// Unix-specific functions

// UnixInit initializes Unix systems - USED on Unix
func UnixInit() {
	fmt.Println("Unix initialization")
	unixSetup()
}

// unixSetup performs Unix setup - USED
func unixSetup() {
	configureUnixSignals()
	setupUnixPaths()
}

// configureUnixSignals configures signal handling - USED
func configureUnixSignals() {
	fmt.Println("Configuring Unix signals")
}

// setupUnixPaths sets up Unix paths - USED
func setupUnixPaths() {
	fmt.Println("Setting up Unix paths")
}

// UnixOnly is only available on Unix - USED in unix_test.go
func UnixOnly() string {
	return "Unix only function"
}

// UnusedUnix is not used on Unix - UNUSED
func UnusedUnix() {
	fmt.Println("Unused Unix function")
}
