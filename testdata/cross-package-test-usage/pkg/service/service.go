package service

import (
	"crosspackage/internal/core"
	"fmt"
)

// Service provides business logic
type Service struct {
	processor *core.Processor
}

// NewService creates a new service
func NewService() *Service {
	return &Service{
		processor: core.NewProcessor(),
	}
}

// Process processes data - USED by api package
func (s *Service) Process(data string) (string, error) {
	// Used by api.Handler.HandleRequest
	if data == "" {
		return "", fmt.Errorf("empty data")
	}

	result := s.processor.Transform(data)
	return postProcess(result), nil
}

// postProcess applies post-processing - USED
func postProcess(data string) string {
	// Used by Process method
	return "processed: " + data
}

// GetProcessor returns the processor - ONLY USED IN TESTS
func (s *Service) GetProcessor() *core.Processor {
	// This is only used in service_test.go
	return s.processor
}

// UpdateConfig updates service configuration - UNUSED
func (s *Service) UpdateConfig(config map[string]string) {
	// This method is not used anywhere
	// Should be reported as unused
}

// internalMethod is unexported and unused - UNUSED
func (s *Service) internalMethod() {
	// Should be reported as unused
}

// Benchmark-only function
func (s *Service) processForBenchmark(data string) string {
	// Only used in benchmarks
	return s.processor.Transform(data)
}
