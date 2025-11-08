package main

import (
	"crosspackage/pkg/api"
	"crosspackage/pkg/service"
	"fmt"
)

func main() {
	// Create service and handler
	svc := service.NewService()
	handler := api.NewHandler(svc)

	// Note: We're NOT using most handler methods here
	// Only NewHandler is used in main

	fmt.Println("Application started")
	_ = handler // Prevent unused variable error
}

// unusedMain should be reported as unused
func unusedMain() {
	fmt.Println("This is not used")
}
