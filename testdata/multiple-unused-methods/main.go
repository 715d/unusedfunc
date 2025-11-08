package main

import "fmt"

// User represents a user with various methods
type User struct {
	ID       int
	Name     string
	Email    string
	Password string
	Active   bool
}

// Exported methods that are used
func (u *User) GetID() int {
	return u.ID
}

func (u *User) GetName() string {
	return u.Name
}

func (u *User) SetName(name string) {
	u.Name = name
}

func (u *User) GetEmail() string {
	return u.Email
}

func (u *User) SetEmail(email string) {
	u.Email = email
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

// Exported methods that are NOT used - should be reported in internal packages only
func (u *User) GetPassword() string {
	return u.Password
}

func (u *User) SetPassword(password string) {
	u.Password = password
}

func (u *User) Clone() *User {
	return &User{
		ID:       u.ID,
		Name:     u.Name,
		Email:    u.Email,
		Password: u.Password,
		Active:   u.Active,
	}
}

// Unexported methods that are used
func (u *User) validateEmail() bool {
	return u.Email != "" && len(u.Email) > 3
}

func (u *User) hashPassword() string {
	// Simple hash simulation
	return fmt.Sprintf("hash_%s", u.Password)
}

// Unexported methods that are NOT used - should be reported
func (u *User) setInternal(value string) {
	// This method is not used anywhere
}

func (u *User) calculateScore() int {
	// This method is not used anywhere
	return u.ID * 10
}

func (u *User) formatDisplay() string {
	// This method is not used anywhere
	return fmt.Sprintf("%s <%s>", u.Name, u.Email)
}

// Method values and expressions
func (u *User) GetDisplayName() string {
	return u.Name + " (" + u.Email + ")"
}

// Admin represents an admin user
type Admin struct {
	User
	Permissions []string
}

// Admin methods
func (a *Admin) AddPermission(perm string) {
	a.Permissions = append(a.Permissions, perm)
}

func (a *Admin) HasPermission(perm string) bool {
	for _, p := range a.Permissions {
		if p == perm {
			return true
		}
	}
	return false
}

// Unused admin method
func (a *Admin) removeAllPermissions() {
	a.Permissions = nil
}

// Package-level functions
func CreateUser(name, email string) *User {
	return &User{
		Name:  name,
		Email: email,
	}
}

func CreateAdmin(name, email string) *Admin {
	return &Admin{
		User: User{
			Name:  name,
			Email: email,
		},
		Permissions: []string{},
	}
}

// Unused package function
func helperFunction() string {
	return "unused helper"
}

// Example usage demonstrating which methods are called
func ExampleComplex() {
	user := CreateUser("John", "john@example.com")

	// Use various User methods
	fmt.Println("ID:", user.GetID())
	fmt.Println("Name:", user.GetName())
	fmt.Println("Email:", user.GetEmail())
	fmt.Println("Active:", user.IsActive())

	user.SetName("John Doe")
	user.SetEmail("john.doe@example.com")
	user.Activate()

	// Use unexported methods
	if user.validateEmail() {
		hash := user.hashPassword()
		fmt.Println("Password hash:", hash)
	}

	// Method value usage
	displayFunc := user.GetDisplayName
	fmt.Println("Display:", displayFunc())

	// Admin usage
	admin := CreateAdmin("Admin", "admin@example.com")
	admin.AddPermission("read")
	admin.AddPermission("write")

	if admin.HasPermission("read") {
		fmt.Println("Admin has read permission")
	}

	// Embedded method usage
	fmt.Println("Admin name:", admin.GetName())
}

func main() {
	ExampleComplex()
}
