// Package interfaces provides test cases for interface method calls that require precision analysis
package main

// Writer interface defines a write method
type Writer interface {
	Write(data []byte) error
}

// FileWriter implements Writer
type FileWriter struct {
	filename string
}

// Write implements the Writer interface (USED via interface)
func (fw *FileWriter) Write(data []byte) error {
	// Implementation details...
	return nil
}

// Close closes the file writer (UNUSED - should be reported)
func (fw *FileWriter) Close() error {
	return nil
}

// BufferWriter implements Writer
type BufferWriter struct {
	buffer []byte
}

// Write implements the Writer interface (USED via interface)
func (bw *BufferWriter) Write(data []byte) error {
	bw.buffer = append(bw.buffer, data...)
	return nil
}

// Flush flushes the buffer (UNUSED - should be reported)
func (bw *BufferWriter) Flush() error {
	bw.buffer = bw.buffer[:0]
	return nil
}

// ProcessData demonstrates interface method calls
func ProcessData(w Writer, data []byte) error {
	// This call requires precision analysis to determine which Write methods are used
	return w.Write(data)
}

// Example usage
func Example() {
	fw := &FileWriter{filename: "test.txt"}
	bw := &BufferWriter{}

	data := []byte("test data")

	// These calls go through interface dispatch
	ProcessData(fw, data)
	ProcessData(bw, data)
}

func main() {
	Example()
}
