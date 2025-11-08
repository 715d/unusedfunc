// Package main tests detection of marker methods required for interface compliance
// when using type assertions. These methods are never directly called but are
// required for type assertions to work.
package main

import (
	"errors"
	"fmt"
)

// CustomError is our interface with a marker method
type CustomError interface {
	error
	// IsCustom is a marker method that's never directly called
	// but required for type assertions to work
	IsCustom()
}

// AppError implements CustomError
type AppError struct {
	message string
}

func (e AppError) Error() string {
	return e.message
}

// IsCustom is the marker method - never directly called but required
// for the type assertion err.(CustomError) to succeed
func (e AppError) IsCustom() {}

// ValidationError also implements CustomError
type ValidationError struct {
	field string
}

func (v ValidationError) Error() string {
	return "validation failed: " + v.field
}

// IsCustom marker method for ValidationError
func (v ValidationError) IsCustom() {}

// UnusedError doesn't implement CustomError
type UnusedError struct {
	code int
}

func (u UnusedError) Error() string {
	return fmt.Sprintf("error code: %d", u.code)
}

// This method should be reported as unused since UnusedError
// is never used and doesn't implement any interface
func (u UnusedError) unusedMethod() string {
	return "never called"
}

// ProcessError checks if an error is a CustomError using type assertion
func ProcessError(err error) string {
	// This type assertion requires IsCustom() to exist
	var customErr CustomError
	if errors.As(err, &customErr) {
		// We never call customErr.IsCustom(), but it must exist
		return "custom error: " + customErr.Error()
	}
	return "standard error: " + err.Error()
}

func main() {
	// Create errors that will be type-asserted
	appErr := AppError{message: "app failed"}
	valErr := ValidationError{field: "email"}

	// These will trigger the type assertion in ProcessError
	fmt.Println(ProcessError(appErr))
	fmt.Println(ProcessError(valErr))

	// Standard error (not CustomError)
	fmt.Println(ProcessError(fmt.Errorf("generic error")))
}
