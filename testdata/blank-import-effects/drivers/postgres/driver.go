// Package postgres provides PostgreSQL driver implementation
package postgres

import (
	"blank_imports/driver"
	"fmt"
)

// PostgresDriver implements the Driver interface
type PostgresDriver struct {
	connected bool
	config    *Config
}

// Config holds driver configuration
type Config struct {
	MaxConnections int
	SSL            bool
}

// Connect connects to PostgreSQL
func (d *PostgresDriver) Connect(dsn string) error {
	d.config = getDefaultConfig()
	d.connected = true
	return nil
}

// Close closes the connection
func (d *PostgresDriver) Close() error {
	d.connected = false
	return nil
}

// init registers the PostgreSQL driver
func init() {
	driver.RegisterDriver("postgres", &PostgresDriver{})
	driver.RegisterDriver("postgresql", &PostgresDriver{}) // Alias
}

// Functions used via init side effects

// getDefaultConfig returns default configuration - USED
func getDefaultConfig() *Config {
	return &Config{
		MaxConnections: 10,
		SSL:            true,
	}
}

// Package-level initialization
var (
	// Initialized at package level
	defaultTimeout = calculateTimeout()
)

// calculateTimeout is used in package-level var init - USED
func calculateTimeout() int {
	return 30
}

// unusedPostgresHelper is not used - UNUSED
func unusedPostgresHelper() {
	fmt.Println("Not used")
}
