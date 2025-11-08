// Package main demonstrates the marker method pattern for interface sealing and type discrimination.
// This pattern is commonly used in error handling where empty methods serve as interface markers.
package main

import (
	"errors"
	"fmt"
)

// AppError is a sealed interface that only types in this package can implement.
// The marker() method is never called but is essential for interface satisfaction.
type AppError interface {
	error
	marker() // Marker method - never called, but required for interface implementation
}

// NotFoundError represents a resource not found error.
type NotFoundError struct {
	resource string
}

// marker implements AppError interface (USED for interface satisfaction - should NOT be reported).
func (e NotFoundError) marker() {}

// Error implements error interface.
func (e NotFoundError) Error() string {
	return fmt.Sprintf("resource not found: %s", e.resource)
}

// ValidationError represents a validation failure.
type ValidationError struct {
	field   string
	message string
}

// marker implements AppError interface (USED for interface satisfaction - should NOT be reported).
func (e ValidationError) marker() {}

// Error implements error interface.
func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s: %s", e.field, e.message)
}

// getField returns the field name (UNUSED - should be reported).
func (e ValidationError) getField() string {
	return e.field
}

// PermissionError represents an authorization failure.
type PermissionError struct {
	action string
	user   string
}

// marker implements AppError interface (USED for interface satisfaction - should NOT be reported).
func (e PermissionError) marker() {}

// Error implements error interface.
func (e PermissionError) Error() string {
	return fmt.Sprintf("user %s not authorized for action: %s", e.user, e.action)
}

// getUser returns the user (UNUSED - should be reported).
func (e PermissionError) getUser() string {
	return e.user
}

// HandleError demonstrates using errors.As() to check for AppError interface.
// This is the key usage pattern - errors.As uses reflection to check if the error
// implements AppError, which requires the marker() method to exist.
func HandleError(err error) string {
	// Check if error is an AppError using type discrimination.
	var appErr AppError
	if errors.As(err, &appErr) {
		return fmt.Sprintf("application error: %s", appErr.Error())
	}
	return fmt.Sprintf("unknown error: %s", err.Error())
}

// ProcessRequest simulates business logic that returns different error types.
func ProcessRequest(requestType string) error {
	switch requestType {
	case "notfound":
		return NotFoundError{resource: "user"}
	case "validation":
		return ValidationError{field: "email", message: "invalid format"}
	case "permission":
		return PermissionError{action: "delete", user: "guest"}
	default:
		return fmt.Errorf("generic error")
	}
}

// unusedHelper is truly unused and should be reported.
func unusedHelper() string {
	return "never called"
}

func main() {
	// Test different error types.
	requests := []string{"notfound", "validation", "permission", "other"}

	for _, req := range requests {
		err := ProcessRequest(req)
		result := HandleError(err)
		fmt.Println(result)

		// Additional type assertions to demonstrate interface usage.
		var appErr AppError
		if errors.As(err, &appErr) {
			fmt.Printf("Confirmed: %s is an AppError\n", req)
		}

		// Specific type checks.
		var notFoundErr NotFoundError
		if errors.As(err, &notFoundErr) {
			fmt.Printf("Specific: NotFoundError for resource: %s\n", notFoundErr.resource)
		}

		var validationErr ValidationError
		if errors.As(err, &validationErr) {
			fmt.Printf("Specific: ValidationError for field: %s\n", validationErr.field)
		}

		var permErr PermissionError
		if errors.As(err, &permErr) {
			fmt.Printf("Specific: PermissionError for action: %s\n", permErr.action)
		}
	}
}
