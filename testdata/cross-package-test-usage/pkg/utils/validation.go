package utils

import "fmt"

// Validate validates input data - USED by api package
func Validate(data string) error {
	if data == "" {
		return fmt.Errorf("data cannot be empty")
	}
	if len(data) > 1000 {
		return fmt.Errorf("data too large")
	}
	return nil
}

// SanitizeInput sanitizes user input - ONLY USED IN EXAMPLE TEST
func SanitizeInput(input string) string {
	// This is only used in validation_test.go Example function
	return sanitizeHelper(input)
}

// sanitizeHelper helps with sanitization - USED
func sanitizeHelper(s string) string {
	// Used by SanitizeInput
	// Simple sanitization
	return s
}

// UnusedValidator validates but is unused - UNUSED
func UnusedValidator(data interface{}) bool {
	// Not used anywhere
	return data != nil
}

// UsedInTest is only used in tests - USED IN TEST
func UsedInTest(s string) string {
	// Only used in validation_test.go
	return "test: " + s
}

// UsedInBenchmark is only used in benchmarks
func UsedInBenchmark(s string) int {
	// Only used in benchmark
	return len(s)
}
