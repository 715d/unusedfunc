// Package external simulates an external library interface.
package external

// Writer is an interface from an external library (like io.Writer or goldmark's util.BufWriter).
type Writer interface {
	Write(p []byte) (n int, err error)
	Available() int
	Flush() error
}
