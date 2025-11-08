// Package main provides the server entry point
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/test/integration/internal/service"
	"github.com/test/integration/pkg/api"
)

// Server represents the HTTP server
type Server struct {
	handler *api.Handler
	port    int
}

// NewServer creates a new server instance
func NewServer(port int) *Server {
	userService := service.NewUserService()
	handler := api.NewHandler(userService)

	return &Server{
		handler: handler,
		port:    port,
	}
}

// Start starts the HTTP server - EXPORTED, unused, should NOT be reported (main package)
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Register routes - this uses the GetUser method from handler
	mux.HandleFunc("/users/", s.handler.GetUser)
	mux.HandleFunc("/health", s.handler.HealthCheck)

	addr := fmt.Sprintf(":%d", s.port)
	fmt.Printf("Server starting on %s\n", addr)

	return http.ListenAndServe(addr, mux)
}

// Stop stops the HTTP server - EXPORTED, unused, should NOT be reported (main package)
func (s *Server) Stop() error {
	fmt.Println("Server stopped")
	return nil
}

// GetPort returns the server port - EXPORTED, unused, should NOT be reported (main package)
func (s *Server) GetPort() int {
	return s.port
}

// SetPort sets the server port - EXPORTED, unused, should NOT be reported (main package)
func (s *Server) SetPort(port int) {
	s.port = port
}

// setupRoutes configures HTTP routes - unexported, unused, should be reported
func (s *Server) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/users/", s.handler.GetUser)
	mux.HandleFunc("/health", s.handler.HealthCheck)
	return mux
}

// configureMiddleware sets up middleware chain - unexported, unused, should be reported
func (s *Server) configureMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	// Chain middleware
	wrapped := s.handler.LoggingMiddleware(handler)
	wrapped = s.handler.AuthMiddleware(wrapped)
	return wrapped
}

// validateConfig validates server configuration - unexported, unused, should be reported
func (s *Server) validateConfig() error {
	if s.port <= 0 || s.port > 65535 {
		return fmt.Errorf("invalid port: %d", s.port)
	}
	return nil
}

// Config represents server configuration
type Config struct {
	Port    int    `json:"port"`
	Host    string `json:"host"`
	Debug   bool   `json:"debug"`
	LogFile string `json:"log_file"`
}

// LoadConfig loads configuration from file - EXPORTED, unused, should NOT be reported (main package)
func LoadConfig(filename string) (*Config, error) {
	// In a real implementation, this would read from file
	return &Config{
		Port:    8080,
		Host:    "localhost",
		Debug:   false,
		LogFile: "server.log",
	}, nil
}

// SaveConfig saves configuration to file - EXPORTED, unused, should NOT be reported (main package)
func SaveConfig(config *Config, filename string) error {
	// In a real implementation, this would write to file
	fmt.Printf("Saving config to %s\n", filename)
	return nil
}

// ValidateConfig validates configuration - EXPORTED, unused, should NOT be reported (main package)
func ValidateConfig(config *Config) error {
	if config.Port <= 0 {
		return fmt.Errorf("invalid port: %d", config.Port)
	}
	if config.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}
	return nil
}

// parseFlags parses command line flags - unexported, unused, should be reported
func parseFlags() *Config {
	return &Config{
		Port:  8080,
		Host:  "localhost",
		Debug: false,
	}
}

// setupLogging configures application logging - unexported, unused, should be reported
func setupLogging(config *Config) error {
	if config.LogFile != "" {
		fmt.Printf("Logging to file: %s\n", config.LogFile)
	}
	return nil
}

// shutdownHandler handles graceful shutdown - unexported, unused, should be reported
func shutdownHandler(server *Server) {
	// Handle graceful shutdown
	server.Stop()
}

// main is the entry point
func main() {
	config, err := LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := ValidateConfig(config); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	server := NewServer(config.Port)

	fmt.Printf("Starting server on port %d\n", server.GetPort())

	if err := server.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
