package pool

import (
	"context"
	"errors"
)

// SingleConnPool always returns the same connection
type SingleConnPool struct {
	basePool Pooler
	conn     *Conn
	closed   bool
}

// Ensure SingleConnPool implements Pooler
var _ Pooler = (*SingleConnPool)(nil)

// NewSingleConnPool creates a new single connection pool
func NewSingleConnPool(basePool Pooler) *SingleConnPool {
	return &SingleConnPool{
		basePool: basePool,
		conn:     &Conn{ID: 1},
	}
}

// Get returns the single connection
func (p *SingleConnPool) Get(_ context.Context) (*Conn, error) {
	if p.closed {
		return nil, errors.New("pool is closed")
	}
	return p.conn, nil
}

// Put does nothing for single connection pool
func (p *SingleConnPool) Put(_ context.Context, _ *Conn) {
	// No-op
}

// Remove marks the pool as having an error
func (p *SingleConnPool) Remove(_ context.Context, _ *Conn, _ error) {
	p.closed = true
}

// Close closes the pool
func (p *SingleConnPool) Close() error {
	p.closed = true
	return nil
}

// Len always returns 1 for single connection pool
func (p *SingleConnPool) Len() int {
	if p.closed {
		return 0
	}
	return 1
}

// IdleLen always returns 0 for single connection pool
func (p *SingleConnPool) IdleLen() int {
	return 0
}
