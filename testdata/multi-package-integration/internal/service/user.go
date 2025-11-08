// Package service contains internal business logic
package service

import (
	"fmt"
	"time"
)

// UserService handles user operations
type UserService struct {
	users map[string]*User
}

// User represents a user entity
type User struct {
	ID        string
	Name      string
	Email     string
	CreatedAt time.Time
	Active    bool
}

// NewUserService creates a new user service
func NewUserService() *UserService {
	return &UserService{
		users: make(map[string]*User),
	}
}

// CreateUser creates a new user - EXPORTED in internal package, unused, should be reported
func (us *UserService) CreateUser(id, name, email string) *User {
	user := &User{
		ID:        id,
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
		Active:    true,
	}
	us.users[id] = user
	return user
}

// GetUser retrieves a user by ID - EXPORTED in internal package, used, should NOT be reported
func (us *UserService) GetUser(id string) (*User, bool) {
	user, exists := us.users[id]
	return user, exists
}

// UpdateUser updates user information - EXPORTED in internal package, unused, should be reported
func (us *UserService) UpdateUser(id string, name, email string) error {
	user, exists := us.users[id]
	if !exists {
		return fmt.Errorf("user not found: %s", id)
	}
	user.Name = name
	user.Email = email
	return nil
}

// DeleteUser removes a user - EXPORTED in internal package, unused, should be reported
func (us *UserService) DeleteUser(id string) error {
	if _, exists := us.users[id]; !exists {
		return fmt.Errorf("user not found: %s", id)
	}
	delete(us.users, id)
	return nil
}

// ListUsers returns all users - EXPORTED in internal package, unused, should be reported
func (us *UserService) ListUsers() []*User {
	users := make([]*User, 0, len(us.users))
	for _, user := range us.users {
		users = append(users, user)
	}
	return users
}

// validateEmail validates user email - unexported, unused, should be reported
func (us *UserService) validateEmail(email string) bool {
	return len(email) > 0 && len(email) < 100
}

// generateID generates a unique ID - unexported, used, should NOT be reported
func (us *UserService) generateID() string {
	return fmt.Sprintf("user_%d", time.Now().UnixNano())
}

// notifyUserCreated sends notification - unexported, unused, should be reported
func (us *UserService) notifyUserCreated(user *User) error {
	fmt.Printf("User created: %s\n", user.ID)
	return nil
}

// User methods

func (u *User) GetDisplayName() string {
	return fmt.Sprintf("%s <%s>", u.Name, u.Email)
}

func (u *User) IsActive() bool {
	return u.Active
}

func (u *User) Activate() {
	u.Active = true
}

func (u *User) Deactivate() {
	u.Active = false
}

// setCreatedAt sets creation time - unexported, unused, should be reported
func (u *User) setCreatedAt(t time.Time) {
	u.CreatedAt = t
}

// age calculates user age - unexported, unused, should be reported
func (u *User) age() int {
	return int(time.Since(u.CreatedAt).Hours() / 24 / 365)
}

func ProcessUser(service *UserService, userID string) {
	user, exists := service.GetUser(userID)
	if exists {
		id := service.generateID()
		fmt.Printf("Processing user %s with new ID %s\n", user.ID, id)
	}
}
