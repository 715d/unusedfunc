// Package driver defines the driver interface and registry
package driver

import "fmt"

// Driver interface that drivers implement
type Driver interface {
	Connect(dsn string) error
	Close() error
}

// Global driver registry
var drivers = make(map[string]Driver)

// RegisterDriver registers a database driver - USED by imported packages
func RegisterDriver(name string, driver Driver) {
	drivers[name] = driver
	fmt.Printf("Registered driver: %s\n", name)
}

// GetDriver retrieves a registered driver - USED in main
func GetDriver(name string) (Driver, bool) {
	d, ok := drivers[name]
	return d, ok
}

// UnusedRegistryFunc should be reported as unused
func UnusedRegistryFunc() {
	fmt.Println("This is not used")
}
