package main

import (
	"bytes"

	"example.com/test/external"
)

// BufWriter implements external.Writer interface.
type BufWriter struct {
	*bytes.Buffer
}

// Write implements external.Writer - called via interface.
func (b *BufWriter) Write(p []byte) (n int, err error) {
	return b.Buffer.Write(p)
}

// Available implements external.Writer but is never called directly.
// This should NOT be reported as unused because of the type assertion below.
func (b *BufWriter) Available() int {
	return 1000
}

// Flush implements external.Writer but is never called directly.
// This should NOT be reported as unused because of the type assertion below.
func (b *BufWriter) Flush() error {
	return nil
}

// processWithInterface is called by external library with BufWriter as Writer interface.
// The type assertion from interface to concrete type PROVES BufWriter is used as Writer.
func processWithInterface(w external.Writer) {
	// Type assertion from interface to concrete type.
	// Pattern: ctx := w.(*ConcreteType) where w is an interface
	// This proves *BufWriter implements external.Writer and requires all interface methods.
	buf := w.(*BufWriter)
	_ = buf
}

func main() {
	// Simulate external library calling our function.
	// In reality, the conversion w := external.Writer(&BufWriter{}) happens in external code.
	processWithInterface(nil)
}
