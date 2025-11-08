package core

import "errors"

// Error definitions

var (
	// ErrInvalidConfig is returned when configuration is invalid - UNUSED
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrProcessingFailed is returned when processing fails - UNUSED
	ErrProcessingFailed = errors.New("processing failed")
)

// Note: These error variables are defined but not used anywhere
// However, exported variables might be used by external packages
// In internal packages, they should be reported if unused
