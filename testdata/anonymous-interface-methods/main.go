// Package main demonstrates a bug where methods required by unnamed interfaces
// are incorrectly reported as unused by the SSA analysis.
package main

import (
	"fmt"
	"net"
	"time"
)

// Conn represents a network connection with metadata.
// This mimics the structure found in go-redis/internal/pool.
type Conn struct {
	conn      net.Conn
	createdAt time.Time
}

// NetConn returns the underlying net.Conn.
// This method is required by unnamed interfaces but will be incorrectly
// reported as unused due to the bug in buildComplianceIndex.
func (c *Conn) NetConn() net.Conn {
	return c.conn
}

// SetCreatedAt sets the creation timestamp.
// This method is required by unnamed interfaces but will be incorrectly
// reported as unused due to the bug in buildComplianceIndex.
func (c *Conn) SetCreatedAt(t time.Time) {
	c.createdAt = t
}

// ProcessConnection accepts a connection via an unnamed interface.
// The unnamed interface requires the NetConn() method.
func ProcessConnection(conn interface{ NetConn() net.Conn }) {
	nc := conn.NetConn()
	fmt.Printf("Processing connection: %v\n", nc.RemoteAddr())
}

// InitializeConnection accepts a connection via an unnamed interface.
// The unnamed interface requires the SetCreatedAt() method.
func InitializeConnection(conn interface{ SetCreatedAt(time.Time) }) {
	conn.SetCreatedAt(time.Now())
	fmt.Println("Connection initialized")
}

// StoreConnection demonstrates a common pattern where unnamed interfaces
// are used in variable declarations.
var connectionStore = []interface {
	NetConn() net.Conn
	SetCreatedAt(time.Time)
}{}

func main() {
	// Create a connection
	conn := &Conn{}

	// Pass the connection to functions that require unnamed interfaces
	// This makes the methods reachable through interface dispatch
	ProcessConnection(conn)
	InitializeConnection(conn)

	// Also add to the store which requires both methods
	connectionStore = append(connectionStore, conn)
}
