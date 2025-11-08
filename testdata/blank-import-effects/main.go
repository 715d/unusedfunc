// Package blank_imports tests functions used via blank import side effects
package main

import (
	"fmt"

	"blank_imports/driver"

	// Blank imports for side effects
	_ "blank_imports/drivers/mysql"
	_ "blank_imports/drivers/postgres"
	_ "blank_imports/registry"
)

// UnusedMainFunc should be reported as unused
func UnusedMainFunc() {
	fmt.Println("This is not used")
}

func main() {
	// List registered drivers
	fmt.Println("Available drivers:")
	// Note: drivers are registered via init() in imported packages

	// Try to get a driver
	if drv, ok := driver.GetDriver("mysql"); ok {
		err := drv.Connect("user:pass@tcp(localhost:3306)/db")
		if err != nil {
			fmt.Printf("Connection failed: %v\n", err)
		}
		drv.Close()
	}
}
