//go:build linux || darwin || freebsd

package main

import "testing"

func TestUnixOnly(t *testing.T) {
	// Uses UnixOnly function
	result := UnixOnly()
	if result != "Unix only function" {
		t.Errorf("unexpected result: %s", result)
	}
}
