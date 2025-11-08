package hscan

import "reflect"

// Scan uses reflection to check for Scanner interface
func Scan(dst interface{}) error {
	v := reflect.ValueOf(dst)

	if v.Kind() == reflect.Ptr && v.Type().NumMethod() > 0 && v.CanInterface() {
		switch scan := v.Interface().(type) {
		case Scanner:
			return scan.ScanRedis("data")
		default:
			// Handle other types
		}
	}
	return nil
}
