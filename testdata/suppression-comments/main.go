// Package edgecases provides test cases for edge cases and suppression
package main

// Calculator provides calculation methods
type Calculator struct {
	precision int
}

// Add performs addition (USED)
func (c *Calculator) Add(a, b float64) float64 {
	return a + b
}

// Subtract performs subtraction (SUPPRESSED)
//
//nolint:unusedfunc This method is intentionally unused for testing
func (c *Calculator) Subtract(a, b float64) float64 {
	return a - b
}

// Multiply performs multiplication (SUPPRESSED)
// lint:ignore unusedfunc Used by external tools
func (c *Calculator) Multiply(a, b float64) float64 {
	return a * b
}

// Divide performs division (UNUSED - should be reported)
func (c *Calculator) Divide(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
}

// MethodValue demonstrates method values
type Handler struct {
	name string
}

// Handle handles requests (USED as method value)
func (h *Handler) Handle(request string) string {
	return h.name + ": " + request
}

// Process processes data (UNUSED - should be reported)
func (h *Handler) Process(data []byte) []byte {
	return data
}

// Example usage
func Example() {
	calc := &Calculator{precision: 2}
	result := calc.Add(1.0, 2.0)

	// Method value usage
	handler := &Handler{name: "test"}
	handleFunc := handler.Handle // This creates a method value
	response := handleFunc("request")

	_ = result
	_ = response
}

func main() {
	Example()
}
