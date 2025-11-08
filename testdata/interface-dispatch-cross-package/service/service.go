package service

import (
	"context"

	"example.com/project/internal/pool"
)

// BaseClient provides a client with connection pooling
type BaseClient struct {
	connPool pool.Pooler
}

// NewClient creates a new client with a default pool
func NewClient() *BaseClient {
	// This demonstrates pool creation in different contexts
	return &BaseClient{
		connPool: initPool(),
	}
}

// Conn returns a client with a sticky connection pool
func (c *BaseClient) Conn() *BaseClient {
	return &BaseClient{
		connPool: pool.NewStickyConnPool(c.connPool),
	}
}

// Transaction creates a new transaction with a sticky pool
func (c *BaseClient) Transaction() *Tx {
	return newTx(c.connPool)
}

// initPool initializes a connection pool
func initPool() pool.Pooler {
	// Create a single connection pool
	return pool.NewSingleConnPool(nil)
}

// Tx represents a transaction
type Tx struct {
	connPool pool.Pooler
}

// newTx creates a new transaction
func newTx(basePool pool.Pooler) *Tx {
	return &Tx{
		connPool: pool.NewStickyConnPool(basePool),
	}
}

// Execute runs operations in the transaction
func (tx *Tx) Execute(ctx context.Context) error {
	conn, err := tx.connPool.Get(ctx)
	if err != nil {
		return err
	}
	tx.connPool.Put(ctx, conn)
	return nil
}
