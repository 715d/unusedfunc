package main

func main() {
	// Main does nothing.
}

// helperCalledFromTest is used by TestUsingHelper in main_test.go.
// This should NOT be reported as unused because Test* functions are entry points.
func helperCalledFromTest() string {
	return "helper"
}

// anotherHelperFromBenchmark is called from BenchmarkUsingHelper.
// This should also NOT be reported as unused.
func anotherHelperFromBenchmark(n int) int {
	sum := 0
	for i := range n {
		sum += i
	}
	return sum
}

// exampleHelper is used by ExampleUsage.
// This should NOT be reported as unused.
func exampleHelper() string {
	return "example output"
}

// trulyUnusedHelper is never called anywhere.
// This SHOULD be reported as unused.
func trulyUnusedHelper() string {
	return "I am not used"
}
