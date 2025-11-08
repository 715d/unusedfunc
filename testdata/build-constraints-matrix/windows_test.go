//go:build windows

package main

import "testing"

func TestWindowsOnly(t *testing.T) {
	// Uses WindowsOnly function
	result := WindowsOnly()
	if result != "Windows only function" {
		t.Errorf("unexpected result: %s", result)
	}
}
