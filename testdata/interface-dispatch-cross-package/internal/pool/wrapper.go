package pool

import "context"

// WrapperPool wraps another pool and adds functionality
type WrapperPool struct {
	wrapped Pooler
}

// Ensure WrapperPool implements Pooler
var _ Pooler = (*WrapperPool)(nil)

// NewWrapperPool creates a new wrapper pool
func NewWrapperPool(p Pooler) *WrapperPool {
	return &WrapperPool{wrapped: p}
}

// Get delegates to wrapped pool
func (w *WrapperPool) Get(ctx context.Context) (*Conn, error) {
	return w.wrapped.Get(ctx)
}

// Put delegates to wrapped pool
func (w *WrapperPool) Put(ctx context.Context, conn *Conn) {
	w.wrapped.Put(ctx, conn)
}

// Remove delegates to wrapped pool
func (w *WrapperPool) Remove(ctx context.Context, conn *Conn, err error) {
	w.wrapped.Remove(ctx, conn, err)
}

// Close delegates to wrapped pool
func (w *WrapperPool) Close() error {
	return w.wrapped.Close()
}

// Len delegates to wrapped pool
func (w *WrapperPool) Len() int {
	return w.wrapped.Len()
}

// IdleLen delegates to wrapped pool
func (w *WrapperPool) IdleLen() int {
	return w.wrapped.IdleLen()
}
