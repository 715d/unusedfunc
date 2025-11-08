package main

import (
	"fmt"
	"regexp"
)

// This tests a pattern where IIFE returns a function that calls other functions
var matchName = func() func(input string) string {
	// Local variable in the IIFE
	nameMatcher := regexp.MustCompile(`^(\w+)$`)

	// Return a function that calls other package functions
	return func(input string) string {
		// These function calls should be detected by SSA reachability analysis
		// but they need to be detected as used
		result := processInput(input)
		result = transformResult(result)
		result = finalizeResult(result)

		if matches := nameMatcher.FindStringSubmatch(result); len(matches) > 0 {
			return matches[0]
		}
		return ""
	}
}()

// These functions should be detected as USED because they're called from the IIFE-returned function
// But if our hypothesis is correct, they will be reported as UNUSED

func processInput(input string) string {
	return "processed_" + input
}

func transformResult(input string) string {
	return "transformed_" + input
}

func finalizeResult(input string) string {
	return "finalized_" + input
}

// This function is genuinely unused and should be reported
func actuallyUnused() {
	fmt.Println("This should be reported as unused")
}

func main() {
	// Use the IIFE-returned function
	result := matchName("test")
	fmt.Println("Result:", result)
}
