package hscan

import (
	"fmt"
	"testing"
)

// TimeType is defined in TEST file but implements Scanner interface
type TimeType struct{}

// ScanRedis is an exported method implementing Scanner
func (t *TimeType) ScanRedis(s string) error {
	fmt.Println("TimeType.ScanRedis called with:", s)
	return nil
}

func TestScan(t *testing.T) {
	tt := &TimeType{}
	if err := Scan(tt); err != nil {
		t.Errorf("Scan failed: %v", err)
	}
}
