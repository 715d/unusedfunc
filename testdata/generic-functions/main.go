// Package generics provides test cases for generic types and methods
package main

// Container is a generic container type
type Container[T any] struct {
	items []T
}

// Add adds an item to the container (USED)
func (c *Container[T]) Add(item T) {
	c.items = append(c.items, item)
}

// Get retrieves an item from the container (USED)
func (c *Container[T]) Get(index int) T {
	return c.items[index]
}

// Size returns the number of items (UNUSED - should be reported)
func (c *Container[T]) Size() int {
	return len(c.items)
}

// Clear clears all items (UNUSED - should be reported)
func (c *Container[T]) Clear() {
	c.items = c.items[:0]
}

// Processor is a generic processor interface
type Processor[T any] interface {
	Process(item T) T
}

// StringProcessor implements Processor[string]
type StringProcessor struct{}

// Process processes a string (USED via generic interface)
func (sp *StringProcessor) Process(item string) string {
	return "processed: " + item
}

// Validate validates a string (UNUSED - should be reported)
func (sp *StringProcessor) Validate(item string) bool {
	return len(item) > 0
}

// GenericFunction is a generic function
func GenericFunction[T any](processor Processor[T], item T) T {
	return processor.Process(item)
}

// Example usage
func Example() {
	// Use Container[int]
	intContainer := &Container[int]{}
	intContainer.Add(42)
	value := intContainer.Get(0)

	// Use Container[string]
	stringContainer := &Container[string]{}
	stringContainer.Add("hello")

	// Use generic processor
	processor := &StringProcessor{}
	result := GenericFunction(processor, "test")

	_ = value
	_ = result
}

func main() {
	Example()
}
