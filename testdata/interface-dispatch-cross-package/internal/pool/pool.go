// Package pool provides connection pooling implementations
package pool

import "context"

// Pooler is the interface that all pool implementations must satisfy
type Pooler interface {
	Get(context.Context) (*Conn, error)
	Put(context.Context, *Conn)
	Remove(context.Context, *Conn, error)
	Close() error
	Len() int
	IdleLen() int
}

// Conn represents a connection
type Conn struct {
	ID int
}
