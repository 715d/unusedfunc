package pool

import (
	"context"
	"errors"
)

// DefaultPool is a basic pool implementation
type DefaultPool struct {
	connections []*Conn
	closed      bool
}

// Ensure DefaultPool implements Pooler
var _ Pooler = (*DefaultPool)(nil)

// NewDefaultPool creates a new default pool
func NewDefaultPool() *DefaultPool {
	return &DefaultPool{
		connections: make([]*Conn, 0, 10),
	}
}

// Get returns a new connection
func (p *DefaultPool) Get(_ context.Context) (*Conn, error) {
	if p.closed {
		return nil, errors.New("pool is closed")
	}
	conn := &Conn{ID: len(p.connections) + 1}
	p.connections = append(p.connections, conn)
	return conn, nil
}

// Put does nothing in default pool
func (p *DefaultPool) Put(_ context.Context, _ *Conn) {
	// No-op
}

// Remove does nothing in default pool
func (p *DefaultPool) Remove(_ context.Context, _ *Conn, _ error) {
	// No-op
}

// Close closes the pool
func (p *DefaultPool) Close() error {
	p.closed = true
	return nil
}

// Len returns the number of connections
func (p *DefaultPool) Len() int {
	return len(p.connections)
}

// IdleLen always returns 0
func (p *DefaultPool) IdleLen() int {
	return 0
}
