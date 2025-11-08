package pool

import (
	"context"
	"errors"
	"sync"
)

// StickyConnPool provides a sticky connection with state management
type StickyConnPool struct {
	basePool Pooler
	mu       sync.Mutex
	conn     *Conn
	closed   bool
	sticky   bool
}

// Ensure StickyConnPool implements Pooler
var _ Pooler = (*StickyConnPool)(nil)

// NewStickyConnPool creates a new sticky connection pool
func NewStickyConnPool(basePool Pooler) *StickyConnPool {
	return &StickyConnPool{
		basePool: basePool,
		conn:     &Conn{ID: 2},
	}
}

// Get returns the sticky connection
func (p *StickyConnPool) Get(_ context.Context) (*Conn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, errors.New("pool is closed")
	}

	p.sticky = true
	return p.conn, nil
}

// Put returns the connection to the pool
func (p *StickyConnPool) Put(_ context.Context, _ *Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.sticky = false
}

// Remove removes the connection due to an error
func (p *StickyConnPool) Remove(_ context.Context, _ *Conn, _ error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.closed = true
	p.sticky = false
}

// Close closes the pool
func (p *StickyConnPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.closed = true
	p.sticky = false
	return nil
}

// Len returns 1 if sticky, 0 otherwise
func (p *StickyConnPool) Len() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed || !p.sticky {
		return 0
	}
	return 1
}

// IdleLen returns 1 if not sticky and not closed
func (p *StickyConnPool) IdleLen() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed || p.sticky {
		return 0
	}
	return 1
}
