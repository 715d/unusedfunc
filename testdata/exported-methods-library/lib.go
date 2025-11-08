// Package lib simulates a library package with exported methods that call unexported methods.
// This tests that exported methods are properly considered as entry points.
package lib

type Client struct {
	url string
}

// PublicMethod is an exported method that should be an entry point
func (c *Client) PublicMethod() string {
	return c.privateHelper()
}

// privateHelper is called by PublicMethod and should NOT be reported as unused
func (c *Client) privateHelper() string {
	return c.url + "/api"
}

// unusedPrivateMethod is never called and SHOULD be reported as unused
func (c *Client) unusedPrivateMethod() string {
	return "never called"
}

// PublicFunction is an exported function that should be an entry point
func PublicFunction() string {
	return helperFunction()
}

// helperFunction is called by PublicFunction and should NOT be reported as unused
func helperFunction() string {
	return "helper"
}

// unusedFunction is never called and SHOULD be reported as unused
func unusedFunction() string {
	return "never called"
}
