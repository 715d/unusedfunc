package api

import (
	"crosspackage/pkg/service"
	"testing"
)

// Test file that uses functions from handler.go

func TestHandleRequest(t *testing.T) {
	svc := service.NewService()
	h := NewHandler(svc)

	// Uses HandleRequest method
	result, err := h.HandleRequest("test data")
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty result")
	}
}

func TestValidateRequest(t *testing.T) {
	svc := service.NewService()
	h := NewHandler(svc)

	// Uses ValidateRequest method
	err := h.ValidateRequest("valid data")
	if err != nil {
		t.Errorf("ValidateRequest failed: %v", err)
	}

	// Test invalid data
	err = h.ValidateRequest("")
	if err == nil {
		t.Error("Expected validation error for empty data")
	}
}

// Note: GetStats() is not tested, so it should be reported as unused
// Note: handleInternal() is not tested (and is unexported)
