package client

import (
	"context"

	"example.com/project/internal/pool"
)

// BaseClient provides base functionality
type BaseClient struct {
	connPool pool.Pooler
}

// NewBaseClient creates a new base client
func NewBaseClient(p pool.Pooler) *BaseClient {
	return &BaseClient{connPool: p}
}

// GetConnection gets a connection from the pool
func (c *BaseClient) GetConnection(ctx context.Context) (*pool.Conn, error) {
	return c.connPool.Get(ctx)
}

// PutConnection returns a connection to the pool
func (c *BaseClient) PutConnection(ctx context.Context, conn *pool.Conn) {
	c.connPool.Put(ctx, conn)
}

// Close closes the client
func (c *BaseClient) Close() error {
	return c.connPool.Close()
}
