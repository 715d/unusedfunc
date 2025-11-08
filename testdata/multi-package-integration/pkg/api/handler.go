// Package api contains public API handlers
package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/test/integration/internal/service"
)

// Handler provides HTTP handlers for the API
type Handler struct {
	userService *service.UserService
}

// NewHandler creates a new API handler
func NewHandler(userService *service.UserService) *Handler {
	return &Handler{
		userService: userService,
	}
}

// GetUser handles GET /users/{id} - EXPORTED, unused, should NOT be reported (public package)
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Path[len("/users/"):]

	user, exists := h.userService.GetUser(userID)
	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// CreateUser handles POST /users - EXPORTED, unused, should NOT be reported (public package)
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// This would use CreateUser if it was actually called
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "created"})
}

// UpdateUser handles PUT /users/{id} - EXPORTED, unused, should NOT be reported (public package)
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Path[len("/users/"):]
	_ = userID // Use the variable to avoid compilation error

	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// This would use UpdateUser if it was actually called
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// DeleteUser handles DELETE /users/{id} - EXPORTED, unused, should NOT be reported (public package)
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Path[len("/users/"):]

	// This would use DeleteUser if it was actually called
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted", "id": userID})
}

// ListUsers handles GET /users - EXPORTED, unused, should NOT be reported (public package)
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// This would use ListUsers if it was actually called
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]map[string]string{})
}

// HealthCheck handles GET /health - EXPORTED, unused, should NOT be reported (public package)
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// parseUserID extracts user ID from URL - unexported, unused, should be reported
func (h *Handler) parseUserID(path string) string {
	if len(path) > len("/users/") {
		return path[len("/users/"):]
	}
	return ""
}

// validateRequest validates incoming request - unexported, unused, should be reported
func (h *Handler) validateRequest(r *http.Request) error {
	if r.Method == "" {
		return fmt.Errorf("empty method")
	}
	return nil
}

// logRequest logs incoming request - unexported, unused, should be reported
func (h *Handler) logRequest(r *http.Request) {
	fmt.Printf("Request: %s %s\n", r.Method, r.URL.Path)
}

// handleError handles error responses - unexported, unused, should be reported
func (h *Handler) handleError(w http.ResponseWriter, err error, code int) {
	http.Error(w, err.Error(), code)
}

// Middleware represents HTTP middleware
type Middleware func(http.HandlerFunc) http.HandlerFunc

// AuthMiddleware provides authentication middleware - EXPORTED, unused, should NOT be reported (public package)
func (h *Handler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Simple auth check
		if r.Header.Get("Authorization") == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// LoggingMiddleware provides logging middleware - EXPORTED, unused, should NOT be reported (public package)
func (h *Handler) LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Request: %s %s\n", r.Method, r.URL.Path)
		next(w, r)
	}
}

// corsMiddleware provides CORS middleware - unexported, unused, should be reported
func (h *Handler) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next(w, r)
	}
}

// rateLimitMiddleware provides rate limiting - unexported, unused, should be reported
func (h *Handler) rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Simple rate limit check
		next(w, r)
	}
}
