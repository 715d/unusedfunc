// Package simple provides test cases for direct method calls without interfaces or generics
package main

// User represents a simple user type
type User struct {
	Name  string
	Email string
}

// GetName returns the user's name (USED)
func (u *User) GetName() string {
	return u.Name
}

// SetName sets the user's name (USED)
func (u *User) SetName(name string) {
	u.Name = name
}

// GetEmail returns the user's email (UNUSED - should be reported)
func (u *User) GetEmail() string {
	return u.Email
}

// setInternal is an unexported method (UNUSED - should be reported)
func (u *User) setInternal(value string) {
	// This method is not used anywhere
}

// Example usage
func Example() {
	user := &User{Name: "John", Email: "john@example.com"}

	// Use GetName
	name := user.GetName()

	// Use SetName
	user.SetName(name + " Doe")
}

func main() {
	Example()
}
