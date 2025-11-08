package main

import (
	"fmt"
	"os"

	"github.com/715d/unusedfunc/testdata/interface-internal-test-type/pkg"
)

// ProductionType is defined in production code, not test files
type ProductionType struct {
	Data string
}

// ScanRedis implements the Scanner interface for production type
func (p *ProductionType) ScanRedis(s string) error {
	if s == "" {
		return fmt.Errorf("empty data")
	}
	p.Data = s
	return nil
}

func main() {
	// Use the production type - this shows that production code
	// can use the Scanner interface, but test types are only in tests
	prod := &ProductionType{}
	if err := pkg.ProcessData(prod); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Also test with nil to show the non-Scanner path
	_ = pkg.ProcessData(nil)

	// Important: We NEVER instantiate or use TimeRFC3339Nano or AnotherTestType here
	// Those types only exist in test files
}
