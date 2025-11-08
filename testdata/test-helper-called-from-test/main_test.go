package main

import (
	"fmt"
	"testing"
)

func TestUsingHelper(t *testing.T) {
	result := helperCalledFromTest()
	if result != "helper" {
		t.Errorf("got %q, want %q", result, "helper")
	}
}

func BenchmarkUsingHelper(b *testing.B) {
	for range b.N {
		_ = anotherHelperFromBenchmark(100)
	}
}

func ExampleUsage() {
	result := exampleHelper()
	fmt.Println(result)
	// Output: example output
}
