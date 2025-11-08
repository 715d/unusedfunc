package main

import "testing"

func TestHelper(t *testing.T) {
	got := helperFunc()
	want := "helper"
	if got != want {
		t.Errorf("helperFunc() = %v, want %v", got, want)
	}
}
