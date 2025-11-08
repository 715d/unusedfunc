//go:build !prod
// +build !prod

package main

import (
	"fmt"
	"reflect"
	"unsafe"
)

// Calculator with various suppression scenarios
type Calculator struct {
	value float64
}

func (c *Calculator) Add(n float64) float64 {
	c.value += n
	return c.value
}

func (c *Calculator) Subtract(n float64) float64 {
	c.value -= n
	return c.value
}

//nolint:unusedfunc
func (c *Calculator) Multiply(n float64) float64 {
	c.value *= n
	return c.value
}

//lint:ignore unusedfunc this is required for interface compliance
func (c *Calculator) Divide(n float64) float64 {
	if n != 0 {
		c.value /= n
	}
	return c.value
}

//nolint:unusedfunc // legacy support
func (c *Calculator) Power(n float64) float64 {
	result := 1.0
	for i := 0; i < int(n); i++ {
		result *= c.value
	}
	c.value = result
	return c.value
}

//lint:ignore unusedfunc needed for backward compatibility with v1 API
func (c *Calculator) Reset() {
	c.value = 0
}

// Methods with various suppression formats
// nolint
func (c *Calculator) Square() float64 {
	c.value *= c.value
	return c.value
}

//nolint:deadcode,unusedfunc
func (c *Calculator) Cube() float64 {
	c.value = c.value * c.value * c.value
	return c.value
}

// Method without suppression - should be reported as unused
func (c *Calculator) SquareRoot() float64 {
	if c.value >= 0 {
		// Simple square root approximation
		c.value = c.value / 2
	}
	return c.value
}

// Handler with reflection scenarios
type Handler struct {
	methods map[string]reflect.Value
}

func (h *Handler) Handle(methodName string, args []interface{}) interface{} {
	if method, exists := h.methods[methodName]; exists {
		values := make([]reflect.Value, len(args))
		for i, arg := range args {
			values[i] = reflect.ValueOf(arg)
		}
		results := method.Call(values)
		if len(results) > 0 {
			return results[0].Interface()
		}
	}
	return nil
}

func (h *Handler) Register(name string, fn interface{}) {
	h.methods[name] = reflect.ValueOf(fn)
}

//nolint:unusedfunc
func (h *Handler) Process(data interface{}) interface{} {
	// This method might be called via reflection
	return fmt.Sprintf("Processed: %v", data)
}

//lint:ignore unusedfunc called via reflection
func (h *Handler) Transform(input string) string {
	return fmt.Sprintf("Transformed: %s", input)
}

// Unregistered method that should be reported
func (h *Handler) Validate(input interface{}) bool {
	return input != nil
}

// Method values and function assignments
type Processor struct {
	handlers []func(interface{}) interface{}
}

func (p *Processor) AddHandler(handler func(interface{}) interface{}) {
	p.handlers = append(p.handlers, handler)
}

func (p *Processor) ProcessAll(data interface{}) []interface{} {
	results := make([]interface{}, len(p.handlers))
	for i, handler := range p.handlers {
		results[i] = handler(data)
	}
	return results
}

func (p *Processor) DefaultHandler(data interface{}) interface{} {
	return fmt.Sprintf("Default: %v", data)
}

//nolint:unusedfunc
func (p *Processor) ErrorHandler(data interface{}) interface{} {
	return fmt.Sprintf("Error: %v", data)
}

// Unused method - should be reported
func (p *Processor) DebugHandler(data interface{}) interface{} {
	return fmt.Sprintf("Debug: %v", data)
}

// Embedded structs and method promotion
type BaseLogger struct {
	prefix string
}

func (bl *BaseLogger) Log(message string) {
	fmt.Printf("%s: %s\n", bl.prefix, message)
}

func (bl *BaseLogger) SetPrefix(prefix string) {
	bl.prefix = prefix
}

//nolint:unusedfunc
func (bl *BaseLogger) Debug(message string) {
	fmt.Printf("DEBUG %s: %s\n", bl.prefix, message)
}

// Unused base method
func (bl *BaseLogger) Error(message string) {
	fmt.Printf("ERROR %s: %s\n", bl.prefix, message)
}

type FileLogger struct {
	BaseLogger
	filename string
}

func (fl *FileLogger) LogToFile(message string) {
	fmt.Printf("Writing to %s: %s: %s\n", fl.filename, fl.prefix, message)
}

func (fl *FileLogger) SetFilename(filename string) {
	fl.filename = filename
}

//lint:ignore unusedfunc required for cleanup
func (fl *FileLogger) Close() {
	fmt.Printf("Closing file: %s\n", fl.filename)
}

// Unused FileLogger method
func (fl *FileLogger) Flush() {
	fmt.Printf("Flushing file: %s\n", fl.filename)
}

// Unsafe operations and linkname scenarios
type UnsafeStruct struct {
	data uintptr
}

func (us *UnsafeStruct) GetPointer() unsafe.Pointer {
	return unsafe.Pointer(us.data)
}

//nolint:unusedfunc
func (us *UnsafeStruct) RuntimeMethod() {
	// This method might be linked to runtime (linkname removed for test compatibility)
}

// Method that uses unsafe operations but is unused
func (us *UnsafeStruct) UnsafeOperation() uintptr {
	return uintptr(unsafe.Pointer(us))
}

// Build tag conditional compilation

type DebugStruct struct {
	debug bool
}

func (ds *DebugStruct) EnableDebug() {
	ds.debug = true
}

//nolint:unusedfunc
func (ds *DebugStruct) DisableDebug() {
	ds.debug = false
}

// Should be unused in non-debug builds
func (ds *DebugStruct) PrintDebugInfo() {
	if ds.debug {
		fmt.Println("Debug mode enabled")
	}
}

// Complex method value scenarios
type Executor struct {
	operations map[string]func(*Executor, interface{}) interface{}
}

func (e *Executor) Execute(operation string, data interface{}) interface{} {
	if op, exists := e.operations[operation]; exists {
		return op(e, data)
	}
	return nil
}

func (e *Executor) RegisterOperation(name string, op func(*Executor, interface{}) interface{}) {
	if e.operations == nil {
		e.operations = make(map[string]func(*Executor, interface{}) interface{})
	}
	e.operations[name] = op
}

func (e *Executor) DefaultOperation(data interface{}) interface{} {
	return fmt.Sprintf("Default operation: %v", data)
}

//nolint:unusedfunc
func (e *Executor) SpecialOperation(data interface{}) interface{} {
	return fmt.Sprintf("Special operation: %v", data)
}

// These methods might be used as method values
func (e *Executor) TransformOperation(data interface{}) interface{} {
	return fmt.Sprintf("Transformed: %v", data)
}

func (e *Executor) FilterOperation(data interface{}) interface{} {
	return fmt.Sprintf("Filtered: %v", data)
}

// Unused executor method
func (e *Executor) ValidateOperation(data interface{}) interface{} {
	return fmt.Sprintf("Validated: %v", data)
}

// Interface with method promotion complexity
type Writer interface {
	Write(data []byte) (int, error)
}

type Logger interface {
	Log(message string)
}

type WriterLogger interface {
	Writer
	Logger
}

type ComplexWriter struct {
	destination string
}

func (cw *ComplexWriter) Write(data []byte) (int, error) {
	fmt.Printf("Writing to %s: %s\n", cw.destination, string(data))
	return len(data), nil
}

func (cw *ComplexWriter) Log(message string) {
	fmt.Printf("Log to %s: %s\n", cw.destination, message)
}

func (cw *ComplexWriter) SetDestination(dest string) {
	cw.destination = dest
}

func (cw *ComplexWriter) Flush() error {
	fmt.Printf("Flushing %s\n", cw.destination)
	return nil
}

func (cw *ComplexWriter) Close() error {
	fmt.Printf("Closing %s\n", cw.destination)
	return nil
}

// Example usage demonstrating complex scenarios
func ExampleAdvanced() {
	// Calculator usage with suppressions
	calc := &Calculator{value: 10}
	fmt.Printf("Add result: %f\n", calc.Add(5))
	fmt.Printf("Subtract result: %f\n", calc.Subtract(3))

	// Handler with reflection
	handler := &Handler{methods: make(map[string]reflect.Value)}
	handler.Register("process", handler.Process)
	result := handler.Handle("process", []interface{}{"test data"})
	fmt.Printf("Handler result: %v\n", result)

	// Processor with method values
	processor := &Processor{}

	// Add method values as handlers
	processor.AddHandler(func(data interface{}) interface{} {
		return processor.DefaultHandler(data)
	})

	// This creates a method value
	transformFunc := processor.DefaultHandler
	processor.AddHandler(transformFunc)

	results := processor.ProcessAll("test")
	fmt.Printf("Processor results: %v\n", results)

	// Embedded struct usage
	fileLogger := &FileLogger{
		BaseLogger: BaseLogger{prefix: "APP"},
		filename:   "app.log",
	}
	fileLogger.Log("Application started")
	fileLogger.LogToFile("File operation")
	fileLogger.SetPrefix("SYS")
	fileLogger.SetFilename("system.log")

	// Executor with dynamic operations
	executor := &Executor{}

	// Register method values as operations
	executor.RegisterOperation("transform", func(e *Executor, data interface{}) interface{} {
		return e.TransformOperation(data)
	})
	executor.RegisterOperation("filter", func(e *Executor, data interface{}) interface{} {
		return e.FilterOperation(data)
	})

	transformResult := executor.Execute("transform", "test data")
	fmt.Printf("Transform result: %v\n", transformResult)

	// Complex writer usage
	writer := &ComplexWriter{destination: "console"}
	writer.Write([]byte("Hello, World!"))
	writer.Log("Application message")
	writer.SetDestination("file")

	// Use as interfaces
	var w Writer = writer
	var l Logger = writer
	var wl WriterLogger = writer

	w.Write([]byte("Interface write"))
	l.Log("Interface log")
	wl.Write([]byte("Combined interface"))
	wl.Log("Combined log")
}

func main() {
	ExampleAdvanced()
}
