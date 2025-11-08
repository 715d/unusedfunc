// Package registry provides global registration hooks
package registry

import (
	"fmt"
	"sync"
)

// Global hooks that can be registered
var (
	startupHooks  []func()
	shutdownHooks []func()
	hooksMutex    sync.Mutex
)

// RegisterStartupHook registers a startup hook - USED in init
func RegisterStartupHook(hook func()) {
	hooksMutex.Lock()
	defer hooksMutex.Unlock()
	startupHooks = append(startupHooks, hook)
}

// RegisterShutdownHook registers a shutdown hook - USED in init
func RegisterShutdownHook(hook func()) {
	hooksMutex.Lock()
	defer hooksMutex.Unlock()
	shutdownHooks = append(shutdownHooks, hook)
}

// RunStartupHooks executes all startup hooks - UNUSED (would be used by main)
func RunStartupHooks() {
	hooksMutex.Lock()
	hooks := make([]func(), len(startupHooks))
	copy(hooks, startupHooks)
	hooksMutex.Unlock()

	for _, hook := range hooks {
		hook()
	}
}

// RunShutdownHooks executes all shutdown hooks - UNUSED (would be used by main)
func RunShutdownHooks() {
	hooksMutex.Lock()
	hooks := make([]func(), len(shutdownHooks))
	copy(hooks, shutdownHooks)
	hooksMutex.Unlock()

	for _, hook := range hooks {
		hook()
	}
}

// Package init registers some default hooks
func init() {
	RegisterStartupHook(defaultStartup)
	RegisterShutdownHook(defaultShutdown)

	// Register via function variable
	RegisterStartupHook(func() {
		initializeLogging()
	})
}

// defaultStartup is registered in init - USED
func defaultStartup() {
	fmt.Println("Default startup hook")
}

// defaultShutdown is registered in init - USED
func defaultShutdown() {
	fmt.Println("Default shutdown hook")
}

// initializeLogging is called from anonymous func in init - USED
func initializeLogging() {
	fmt.Println("Logging initialized")
}

// unusedHook is not registered - UNUSED
func unusedHook() {
	fmt.Println("This hook is not registered")
}

// Complex registration pattern
type Handler interface {
	Handle() error
}

var handlers []Handler

// RegisterHandler registers a handler - USED
func RegisterHandler(h Handler) {
	handlers = append(handlers, h)
}

type startupHandler struct{}

func (s *startupHandler) Handle() error {
	return nil
}

func init() {
	// Register handler via interface
	RegisterHandler(&startupHandler{})
}

// unusedHandler is not registered
type unusedHandler struct{}

func (u *unusedHandler) Handle() error {
	// Not registered, should be reported as unused
	return nil
}
