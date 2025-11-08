package main

import (
	"net"
	"syscall"
	"time"
)

// MockRawConn implements syscall.RawConn interface
type MockRawConn struct{}

// Control implements syscall.RawConn interface but is never called directly
func (m *MockRawConn) Control(f func(fd uintptr)) error {
	return nil
}

// Read implements syscall.RawConn interface and is actually used
func (m *MockRawConn) Read(f func(fd uintptr) (done bool)) error {
	return nil
}

// Write implements syscall.RawConn interface but is never called directly
func (m *MockRawConn) Write(f func(fd uintptr) (done bool)) error {
	return nil
}

// MockConn implements net.Conn and syscall.Conn interfaces
type MockConn struct {
	rawConn *MockRawConn
}

func (c *MockConn) Read(b []byte) (n int, err error)   { return 0, nil }
func (c *MockConn) Write(b []byte) (n int, err error)  { return 0, nil }
func (c *MockConn) Close() error                       { return nil }
func (c *MockConn) LocalAddr() net.Addr                { return nil }
func (c *MockConn) RemoteAddr() net.Addr               { return nil }
func (c *MockConn) SetDeadline(t time.Time) error      { return nil }
func (c *MockConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *MockConn) SetWriteDeadline(t time.Time) error { return nil }

// SyscallConn returns the underlying raw connection
func (c *MockConn) SyscallConn() (syscall.RawConn, error) {
	return c.rawConn, nil
}

// ConnectionChecker checks if a connection is healthy
type ConnectionChecker struct{}

// CheckConnection tests connection health using only the Read method
func (cc *ConnectionChecker) CheckConnection(conn net.Conn) error {
	// Type assert to get syscall.Conn interface
	syscallConn, ok := conn.(syscall.Conn)
	if !ok {
		return nil
	}

	rawConn, err := syscallConn.SyscallConn()
	if err != nil {
		return err
	}

	// Only uses Read method - Control and Write are never called
	return rawConn.Read(func(fd uintptr) bool {
		return true
	})
}

func main() {
	checker := &ConnectionChecker{}
	mockConn := &MockConn{
		rawConn: &MockRawConn{},
	}

	// This will only call Read method through the interface
	checker.CheckConnection(mockConn)
}
