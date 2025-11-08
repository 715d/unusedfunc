package main

import (
	"fmt"
	"io"
)

// Shape interface with multiple methods
type Shape interface {
	Area() float64
	Perimeter() float64
	String() string
}

// DrawableShape extends Shape
type DrawableShape interface {
	Shape
	Draw() error
	SetColor(color string)
}

// Rectangle implements Shape and DrawableShape
type Rectangle struct {
	Width  float64
	Height float64
	Color  string
}

func (r *Rectangle) Area() float64 {
	return r.Width * r.Height
}

func (r *Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

func (r *Rectangle) String() string {
	return fmt.Sprintf("Rectangle(%fx%f)", r.Width, r.Height)
}

func (r *Rectangle) Draw() error {
	fmt.Printf("Drawing %s rectangle in %s\n", r.String(), r.Color)
	return nil
}

func (r *Rectangle) SetColor(color string) {
	r.Color = color
}

// Unused method on Rectangle
func (r *Rectangle) Scale(factor float64) {
	r.Width *= factor
	r.Height *= factor
}

// Circle implements Shape
type Circle struct {
	Radius float64
	Color  string
}

func (c *Circle) Area() float64 {
	return 3.14159 * c.Radius * c.Radius
}

func (c *Circle) Perimeter() float64 {
	return 2 * 3.14159 * c.Radius
}

func (c *Circle) String() string {
	return fmt.Sprintf("Circle(r=%f)", c.Radius)
}

// Circle doesn't implement DrawableShape fully
func (c *Circle) SetColor(color string) {
	c.Color = color
}

// Unused Circle method
func (c *Circle) GetDiameter() float64 {
	return 2 * c.Radius
}

// Processor interface for complex scenarios
type Processor interface {
	Process(data interface{}) error
	Reset()
	GetStats() map[string]interface{}
}

// DataProcessor implements Processor
type DataProcessor struct {
	processed int
	errors    int
}

func (dp *DataProcessor) Process(data interface{}) error {
	dp.processed++
	fmt.Printf("Processing: %v\n", data)
	return nil
}

func (dp *DataProcessor) Reset() {
	dp.processed = 0
	dp.errors = 0
}

func (dp *DataProcessor) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"processed": dp.processed,
		"errors":    dp.errors,
	}
}

// Unused DataProcessor methods
func (dp *DataProcessor) IncrementErrors() {
	dp.errors++
}

func (dp *DataProcessor) GetProcessedCount() int {
	return dp.processed
}

// FileProcessor implements Processor and io.Writer
type FileProcessor struct {
	DataProcessor
	filename string
}

func (fp *FileProcessor) Write(p []byte) (n int, err error) {
	fmt.Printf("Writing to %s: %s\n", fp.filename, string(p))
	return len(p), nil
}

func (fp *FileProcessor) SetFilename(filename string) {
	fp.filename = filename
}

// Unused FileProcessor method
func (fp *FileProcessor) GetFilename() string {
	return fp.filename
}

// Generic interface usage
type Container interface {
	Add(item interface{})
	Get(index int) interface{}
	Size() int
}

type SliceContainer struct {
	items []interface{}
}

func (sc *SliceContainer) Add(item interface{}) {
	sc.items = append(sc.items, item)
}

func (sc *SliceContainer) Get(index int) interface{} {
	if index >= 0 && index < len(sc.items) {
		return sc.items[index]
	}
	return nil
}

func (sc *SliceContainer) Size() int {
	return len(sc.items)
}

// Unused SliceContainer methods
func (sc *SliceContainer) Clear() {
	sc.items = nil
}

func (sc *SliceContainer) Remove(index int) interface{} {
	if index >= 0 && index < len(sc.items) {
		item := sc.items[index]
		sc.items = append(sc.items[:index], sc.items[index+1:]...)
		return item
	}
	return nil
}

// Functions that use interfaces
func CalculateShapeArea(shape Shape) float64 {
	return shape.Area()
}

func DrawShape(drawable DrawableShape) error {
	drawable.SetColor("blue")
	return drawable.Draw()
}

func ProcessData(processor Processor, data []interface{}) error {
	processor.Reset()
	for _, item := range data {
		if err := processor.Process(item); err != nil {
			return err
		}
	}
	fmt.Printf("Stats: %v\n", processor.GetStats())
	return nil
}

func WriteToProcessor(writer io.Writer, data string) error {
	_, err := writer.Write([]byte(data))
	return err
}

func UseContainer(container Container) {
	container.Add("item1")
	container.Add("item2")
	container.Add("item3")

	fmt.Printf("Container size: %d\n", container.Size())
	fmt.Printf("First item: %v\n", container.Get(0))
}

// Type assertion scenarios
func HandleShape(shape interface{}) {
	switch s := shape.(type) {
	case *Rectangle:
		fmt.Printf("Rectangle area: %f\n", s.Area())
	case *Circle:
		fmt.Printf("Circle area: %f\n", s.Area())
	case Shape:
		fmt.Printf("Generic shape area: %f\n", s.Area())
	default:
		fmt.Println("Unknown shape")
	}
}

// Interface composition
type ReadWriteProcessor interface {
	Processor
	io.Writer
	io.Reader
}

// Multiple interface implementation
func ProcessAndWrite(processor ReadWriteProcessor, data []interface{}) error {
	err := ProcessData(processor, data)
	if err != nil {
		return err
	}

	return WriteToProcessor(processor, "Processing complete")
}

// Example usage showing interface method calls
func ExampleAdvanced() {
	// Shape interface usage
	rectangle := &Rectangle{Width: 10, Height: 5, Color: "red"}
	circle := &Circle{Radius: 3, Color: "green"}

	fmt.Printf("Rectangle area: %f\n", CalculateShapeArea(rectangle))
	fmt.Printf("Circle area: %f\n", CalculateShapeArea(circle))

	// DrawableShape interface usage
	err := DrawShape(rectangle)
	if err != nil {
		fmt.Printf("Error drawing: %v\n", err)
	}

	// Processor interface usage
	processor := &DataProcessor{}
	data := []interface{}{"item1", "item2", "item3"}
	err = ProcessData(processor, data)
	if err != nil {
		fmt.Printf("Error processing: %v\n", err)
	}

	// io.Writer interface usage through FileProcessor
	fileProcessor := &FileProcessor{}
	fileProcessor.SetFilename("output.txt")
	err = WriteToProcessor(fileProcessor, "Hello, World!")
	if err != nil {
		fmt.Printf("Error writing: %v\n", err)
	}

	// Container interface usage
	container := &SliceContainer{}
	UseContainer(container)

	// Type assertion usage
	HandleShape(rectangle)
	HandleShape(circle)
	HandleShape("not a shape")
}

func main() {
	ExampleAdvanced()
}
