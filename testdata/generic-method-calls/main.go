// Package genericmethodcalls tests that generic methods calling other generic methods
// are correctly tracked as used. This reproduces the bug found in GORM where
// helper methods on generic types were reported as unused when used only through
// interface returns without concrete instantiation in the same package.
package genericmethodcalls

// Processor is a generic interface that processes values
type Processor[T any] interface {
	Process() T
}

// processor is an unexported generic type implementing Processor
type processor[T any] struct {
	value T
}

// helper is an unexported helper method on generic type.
// BUG: Currently reported as unused, but it IS used by Process().
// This should NOT appear in the unused report.
func (p *processor[T]) helper() T {
	return p.value
}

// validate is another unexported helper.
// BUG: Currently reported as unused, but it IS used by Process().
// This should NOT appear in the unused report.
func (p *processor[T]) validate() bool {
	return true
}

// Process is exported and implements the Processor interface.
// It calls the unexported helper methods.
func (p *processor[T]) Process() T {
	if !p.validate() {
		var zero T
		return zero
	}
	return p.helper()
}

// NewProcessor is an exported factory function returning the interface.
// This is the entry point that makes processor reachable.
// Note: No concrete instantiation in this package - consumers will call NewProcessor[ConcreteType]()
func NewProcessor[T any](val T) Processor[T] {
	return &processor[T]{value: val}
}

// Container demonstrates a more complex case with multiple levels
type Container[T any] struct {
	data []T
}

// internalProcess is a helper used by exported methods.
// BUG: Currently reported as unused, but it IS used by Add().
// This should NOT appear in the unused report.
func (c *Container[T]) internalProcess(item T) T {
	return item
}

// internalValidate is another helper.
// BUG: Currently reported as unused, but it IS used by Add().
// This should NOT appear in the unused report.
func (c *Container[T]) internalValidate(item T) bool {
	return true
}

// Add is exported and uses the internal helpers
func (c *Container[T]) Add(item T) {
	if c.internalValidate(item) {
		processed := c.internalProcess(item)
		c.data = append(c.data, processed)
	}
}

// Get is exported
func (c *Container[T]) Get(index int) T {
	return c.data[index]
}

// unusedHelper is genuinely unused and SHOULD be reported
func (c *Container[T]) unusedHelper() T {
	var zero T
	return zero
}

// NewContainer is exported - consumers will instantiate with concrete types
func NewContainer[T any]() *Container[T] {
	return &Container[T]{data: make([]T, 0)}
}
