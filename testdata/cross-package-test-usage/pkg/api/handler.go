package api

import (
	"crosspackage/pkg/service"
	"crosspackage/pkg/utils"
	"fmt"
)

// Handler handles API requests
type Handler struct {
	svc *service.Service
}

// NewHandler creates a new handler
func NewHandler(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

// HandleRequest processes a request - USED in tests
func (h *Handler) HandleRequest(data string) (string, error) {
	// This is only used in handler_test.go
	result, err := h.svc.Process(data)
	if err != nil {
		return "", fmt.Errorf("processing failed: %w", err)
	}
	return result, nil
}

// HandleInternal processes internal requests - NOT EXPORTED, NOT USED
func (h *Handler) handleInternal(data string) string {
	// This unexported method is not used anywhere
	return "internal: " + data
}

// ValidateRequest validates input - USED in tests
func (h *Handler) ValidateRequest(data string) error {
	// Only used in handler_test.go
	return utils.Validate(data)
}

// GetStats returns handler statistics - EXPORTED BUT UNUSED
func (h *Handler) GetStats() map[string]int {
	// This exported method is not used anywhere
	return map[string]int{
		"requests": 0,
		"errors":   0,
	}
}

// Suppressed method
//
//nolint:unusedfunc
func (h *Handler) suppressedMethod() {
	// This is suppressed
}

// Helper functions

// formatError formats an error - UNUSED
func formatError(err error) string {
	// Not actually used anywhere
	return fmt.Sprintf("ERROR: %v", err)
}

// unusedHelper is not used - UNUSED
func unusedHelper(s string) string {
	return "helper: " + s
}
