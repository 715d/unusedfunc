package utils

import (
	"fmt"
	"testing"
)

func TestUsedInTest(t *testing.T) {
	// Uses UsedInTest function
	result := UsedInTest("hello")
	if result != "test: hello" {
		t.Errorf("unexpected result: %s", result)
	}
}

func ExampleSanitizeInput() {
	// Example functions are special - they're documentation
	// Uses SanitizeInput function
	clean := SanitizeInput("user input")
	fmt.Println(clean)
	// Output: user input
}

func BenchmarkUsedInBenchmark(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// Uses UsedInBenchmark function
		_ = UsedInBenchmark("benchmark string")
	}
}
