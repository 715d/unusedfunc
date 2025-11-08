package service

import (
	"testing"
)

func TestGetProcessor(t *testing.T) {
	s := NewService()

	// Uses GetProcessor method
	p := s.GetProcessor()
	if p == nil {
		t.Error("Expected non-nil processor")
	}
}

func BenchmarkProcessForBenchmark(b *testing.B) {
	s := NewService()

	for i := 0; i < b.N; i++ {
		// Uses processForBenchmark
		_ = s.processForBenchmark("benchmark data")
	}
}
