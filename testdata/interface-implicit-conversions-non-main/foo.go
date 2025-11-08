package foo

import "fmt"

// Define interfaces
type Writer interface {
	Write([]byte) (int, error)
}

type Reader interface {
	Read([]byte) (int, error)
}

type ReadWriter interface {
	Reader
	Writer
}

// Concrete type that implements Writer
type MyWriter struct{}

func (m *MyWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

// Concrete type that implements Reader
type MyReader struct{}

// Read is never used, and MyReader is never instantiated, but it isn't marked as unused because this package is importable.
func (m *MyReader) Read(p []byte) (int, error) {
	return len(p), nil
}

// Concrete type that implements ReadWriter
type MyReadWriter struct{}

func (m *MyReadWriter) Read(p []byte) (int, error) {
	return len(p), nil
}

func (m *MyReadWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

// Test 1: Implicit conversion in function call
func acceptWriter(w Writer) {
	w.Write([]byte("test"))
}

func testCallConversion() {
	mw := &MyWriter{}
	acceptWriter(mw) // Implicit conversion from *MyWriter to Writer
}

// Test 2: Implicit conversion in variable assignment (Store)
func testStoreConversion() {
	var w Writer
	mw := &MyWriter{}
	w = mw // Implicit conversion via Store instruction
	w.Write([]byte("test"))
}

// Test 3: Implicit conversion in channel send
func testSendConversion() {
	ch := make(chan Writer, 1)
	mw := &MyWriter{}
	ch <- mw // Implicit conversion via Send instruction
	<-ch
}

// Test 4: Phi nodes with interface values
func testPhiConversion(flag bool) Writer {
	var w Writer
	if flag {
		mw := &MyWriter{}
		w = mw
	} else {
		// Different branch, Phi node will merge
		mw2 := &MyWriter{}
		w = mw2
	}
	return w // Phi node merges the two paths
}

// Test 5: Slice/array of interfaces
func testSliceConversion() {
	writers := []Writer{
		&MyWriter{}, // Implicit conversion in composite literal
	}
	_ = writers
}

// Test 6: Map with interface values
func testMapConversion() {
	m := map[string]Writer{
		"writer": &MyWriter{}, // Implicit conversion in map literal
	}
	_ = m
}

// Test 7: Interface field in struct
type Container struct {
	W Writer
}

func testStructFieldConversion() {
	c := Container{
		W: &MyWriter{}, // Implicit conversion in struct literal
	}
	_ = c
}

// Test 8: Method with interface receiver (less common)
func testMethodCall() {
	mrw := &MyReadWriter{}
	var r ReadWriter = mrw
	r.Read([]byte("test"))
}

func main() {
	fmt.Println("Testing implicit interface conversions")
	testCallConversion()
	testStoreConversion()
	testSendConversion()
	testPhiConversion(true)
	testSliceConversion()
	testMapConversion()
	testStructFieldConversion()
	testMethodCall()
}
