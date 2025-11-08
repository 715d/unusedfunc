package hscan

import (
	"fmt"
	"reflect"
)

// Scanner interface - defined in production code
type Scanner interface {
	ScanRedis(s string) error
}

// ScanValue mimics the pattern used in go-redis/internal/hscan/structmap.go
// This function is called from outside the internal package
func ScanValue(v interface{}) error {
	rv := reflect.ValueOf(v)

	// Handle pointer types
	if rv.Kind() == reflect.Ptr && !rv.IsNil() {
		// This is the critical pattern from go-redis
		if rv.Type().NumMethod() > 0 && rv.CanInterface() {
			// Type assertion via reflection - this is where the bug manifests
			switch scan := rv.Interface().(type) {
			case Scanner:
				// This calls ScanRedis via interface dispatch
				// Use a valid RFC3339 timestamp for TimeRFC3339Nano
				return scan.ScanRedis("2006-01-02T15:04:05.999999999Z")
			default:
				// Not a Scanner
			}
		}
	}

	return fmt.Errorf("value does not implement Scanner")
}
