// Package linkname - demonstrates external linkname usage
package main

import _ "unsafe"

// This file simulates linking to functions defined in main.go

// Link to the internalLinked function from main.go
//
//go:linkname useInternalLinked main.internalLinked
func useInternalLinked() int

// Function that uses the linked function
func callsLinkedFunction() {
	value := useInternalLinked()
	println("Got value from linked function:", value)
}

// Another unused function to test
func anotherUnused() {
	println("This should be reported as unused")
}
