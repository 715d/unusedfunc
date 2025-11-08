// Package init_chains tests init() function chains and dependencies
package main

import (
	"fmt"
	"sync"
)

var (
	// Package-level variables initialized in init
	config     *Config
	logger     *Logger
	registered []Handler

	// Initialization tracking
	initOrder []string
	initMutex sync.Mutex
)

// Types used in initialization

type Config struct {
	Name  string
	Debug bool
}

type Logger struct {
	prefix string
}

type Handler func(string) error

// Functions called from init() - should NOT be reported as unused

func setupConfig() {
	recordInit("setupConfig")
	config = &Config{
		Name:  "InitChainTest",
		Debug: true,
	}
}

func setupLogger() {
	recordInit("setupLogger")
	logger = &Logger{
		prefix: config.Name,
	}
	logger.setup()
}

func (l *Logger) setup() {
	recordInit("Logger.setup")
	// Logger initialization
}

func registerDefaultHandlers() {
	recordInit("registerDefaultHandlers")
	registered = append(registered, defaultHandler)
	registered = append(registered, errorHandler)
}

func defaultHandler(msg string) error {
	// Used via registration in init
	fmt.Printf("Default: %s\n", msg)
	return nil
}

func errorHandler(msg string) error {
	// Used via registration in init
	return fmt.Errorf("error: %s", msg)
}

// Complex init chain with multiple init functions

func init() {
	recordInit("init#1")
	// First init - setup basic config
	setupConfig()
}

func init() {
	recordInit("init#2")
	// Second init - setup logger (depends on config)
	setupLogger()
}

func init() {
	recordInit("init#3")
	// Third init - register handlers
	registerDefaultHandlers()

	// Register more handlers
	registerHandler(customHandler)

	// Initialize subsystems
	initializeSubsystem()
}

func registerHandler(h Handler) {
	recordInit("registerHandler")
	registered = append(registered, h)
}

func customHandler(msg string) error {
	// Used via registration in init
	return nil
}

func initializeSubsystem() {
	recordInit("initializeSubsystem")
	// Complex initialization
	mgr := newSubsystemManager()
	mgr.initialize()
}

type subsystemManager struct {
	ready bool
}

func newSubsystemManager() *subsystemManager {
	recordInit("newSubsystemManager")
	return &subsystemManager{}
}

func (s *subsystemManager) initialize() {
	recordInit("subsystemManager.initialize")
	s.ready = true
	s.setupInternals()
}

func (s *subsystemManager) setupInternals() {
	recordInit("subsystemManager.setupInternals")
	// Internal setup called from initialize
}

// Package-level variable with function value
var (
	// Function assigned at package level
	processor = processData

	// Function assigned in init
	validator Handler
)

func init() {
	recordInit("init#4")
	// Assign function to package variable
	validator = validateData
}

func processData(data string) string {
	// Used via package-level variable assignment
	return "processed: " + data
}

func validateData(data string) error {
	// Used via init assignment
	if data == "" {
		return fmt.Errorf("empty data")
	}
	return nil
}

// Functions NOT called from init - should be reported as unused

func unusedInInit() {
	// Not called from any init
	fmt.Println("Not used in init")
}

func unusedHelper() string {
	// Not used anywhere
	return "unused"
}

// Functions used in main

func useHandlers() {
	for _, h := range registered {
		_ = h("test message")
	}
}

func recordInit(name string) {
	initMutex.Lock()
	defer initMutex.Unlock()
	initOrder = append(initOrder, name)
}

func printInitOrder() {
	fmt.Println("Init order:")
	for i, name := range initOrder {
		fmt.Printf("%d: %s\n", i+1, name)
	}
}

// Suppressed function
//
//nolint:unusedfunc
func suppressedInit() {
	// Suppressed via nolint
}

// Edge case: circular init dependencies (should be handled carefully)

type Registry struct {
	items map[string]interface{}
}

var globalRegistry *Registry

func init() {
	recordInit("init#registry")
	globalRegistry = newRegistry()
	globalRegistry.register("self", globalRegistry)
}

func newRegistry() *Registry {
	recordInit("newRegistry")
	return &Registry{
		items: make(map[string]interface{}),
	}
}

func (r *Registry) register(name string, item interface{}) {
	recordInit("Registry.register")
	r.items[name] = item
}

// Unused registry method
func (r *Registry) unregister(name string) {
	// Not used anywhere
	delete(r.items, name)
}

func main() {
	printInitOrder()
	useHandlers()

	// Use package-level function variables
	result := processor("test")
	fmt.Println(result)

	if err := validator(""); err != nil {
		fmt.Printf("Validation error: %v\n", err)
	}
}
