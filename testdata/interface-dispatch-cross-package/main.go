// Package main demonstrates interface dispatch across package boundaries
// This reproduces the pattern where SingleConnPool and StickyConnPool
// show non-deterministic behavior in the unusedfunc analyzer.
package main

import (
	"context"

	"example.com/project/service"
)

func main() {
	// Create a service client with connection pooling
	client := service.NewClient()

	// Use the client for a connection
	conn := client.Conn()

	// Create a transaction
	tx := client.Transaction()
	ctx := context.Background()

	// Execute the transaction
	_ = tx.Execute(ctx)

	// Another transaction from the connection
	tx2 := conn.Transaction()
	_ = tx2.Execute(ctx)
}
