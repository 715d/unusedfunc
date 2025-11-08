package main

import (
	"cmp"
	"fmt"
)

type Container[T any] struct {
	items []T
}

func (c *Container[T]) Add(item T) {
	c.items = append(c.items, item)
}

func (c *Container[T]) Get(index int) T {
	var zero T
	if index >= 0 && index < len(c.items) {
		return c.items[index]
	}
	return zero
}

func (c *Container[T]) Size() int {
	return len(c.items)
}

func (c *Container[T]) Clear() {
	c.items = nil
}

func (c *Container[T]) Remove(index int) T {
	var zero T
	if index >= 0 && index < len(c.items) {
		item := c.items[index]
		c.items = append(c.items[:index], c.items[index+1:]...)
		return item
	}
	return zero
}

func (c *Container[T]) Contains(item T) bool {
	for _, existing := range c.items {
		// This would need comparable constraint in real code
		_ = existing
	}
	return false
}

type Processor[T any] interface {
	Process(item T) T
	Reset()
	GetCount() int
}

type StringProcessor struct {
	count int
}

func (sp *StringProcessor) Process(item string) string {
	sp.count++
	return fmt.Sprintf("processed: %s", item)
}

func (sp *StringProcessor) Reset() {
	sp.count = 0
}

func (sp *StringProcessor) GetCount() int {
	sp.count++
	return sp.count
}

func (sp *StringProcessor) Validate(item string) bool {
	return len(item) > 0
}

func GenericFunction[T any](item T) T {
	fmt.Printf("Processing item: %v\n", item)
	return item
}

func Sum[T interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}](slice []T) T {
	var result T
	for _, item := range slice {
		result += item
	}
	return result
}

func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Zip[T, U any](slice1 []T, slice2 []U) []Pair[T, U] {
	result := make([]Pair[T, U], 0, len(slice1))
	for i := 0; i < len(slice1) && i < len(slice2); i++ {
		result = append(result, Pair[T, U]{First: slice1[i], Second: slice2[i]})
	}
	return result
}

type Pair[T, U any] struct {
	First  T
	Second U
}

func (p *Pair[T, U]) GetFirst() T {
	return p.First
}

func (p *Pair[T, U]) GetSecond() U {
	return p.Second
}

func (p *Pair[T, U]) SetFirst(value T) {
	p.First = value
}

func (p *Pair[T, U]) SetSecond(value U) {
	p.Second = value
}

func (p *Pair[T, U]) Swap() {
	fmt.Println("Swapping values")
}

func (p *Pair[T, U]) String() string {
	return fmt.Sprintf("Pair(%v, %v)", p.First, p.Second)
}

type Map[K comparable, V any] struct {
	data map[K]V
}

func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{
		data: make(map[K]V),
	}
}

func (m *Map[K, V]) Set(key K, value V) {
	m.data[key] = value
}

func (m *Map[K, V]) Get(key K) (V, bool) {
	value, exists := m.data[key]
	return value, exists
}

func (m *Map[K, V]) Delete(key K) {
	delete(m.data, key)
}

func (m *Map[K, V]) Size() int {
	return len(m.data)
}

func (m *Map[K, V]) Clear() {
	m.data = make(map[K]V)
}

func (m *Map[K, V]) Keys() []K {
	keys := make([]K, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

func (m *Map[K, V]) Values() []V {
	values := make([]V, 0, len(m.data))
	for _, v := range m.data {
		values = append(values, v)
	}
	return values
}

type Calculator struct {
	operations int
}

func CalculateGeneric[T interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}](c *Calculator, a, b T, op string) T {
	c.operations++
	switch op {
	case "+":
		return a + b
	case "-":
		return a - b
	default:
		return a
	}
}

func (c *Calculator) GetOperations() int {
	return c.operations
}

func (c *Calculator) Reset() {
	c.operations = 0
}

type Comparable[T comparable] interface {
	Compare(other T) int
	Equals(other T) bool
}

type IntComparable struct {
	value int
}

func (ic *IntComparable) Compare(other int) int {
	if ic.value < other {
		return -1
	} else if ic.value > other {
		return 1
	}
	return 0
}

func (ic *IntComparable) Equals(other int) bool {
	return ic.value == other
}

func (ic *IntComparable) GetValue() int {
	return ic.value
}

func ProcessComparable[T comparable](comp Comparable[T], value T) bool {
	return comp.Equals(value)
}

type StringContainer = Container[string]
type IntContainer = Container[int]
type StringIntPair = Pair[string, int]

func ExampleAdvanced() {
	// Container usage
	stringContainer := &Container[string]{}
	stringContainer.Add("hello")
	stringContainer.Add("world")
	fmt.Printf("String container size: %d\n", stringContainer.Size())
	fmt.Printf("First item: %s\n", stringContainer.Get(0))

	intContainer := &Container[int]{}
	intContainer.Add(1)
	intContainer.Add(2)
	intContainer.Add(3)
	fmt.Printf("Int container size: %d\n", intContainer.Size())

	// Processor usage
	processor := &StringProcessor{}
	result := processor.Process("test")
	fmt.Printf("Processed: %s\n", result)
	fmt.Printf("Count: %d\n", processor.GetCount())

	// Generic function usage
	str := GenericFunction("hello")
	num := GenericFunction(42)
	fmt.Printf("Generic results: %s, %d\n", str, num)

	// Sum function usage
	intSum := Sum([]int{1, 2, 3, 4, 5})
	floatSum := Sum([]float64{1.1, 2.2, 3.3})
	fmt.Printf("Sums: %d, %f\n", intSum, floatSum)

	// Pair usage
	pair := &Pair[string, int]{First: "answer", Second: 42}
	fmt.Printf("Pair: %s = %d\n", pair.GetFirst(), pair.GetSecond())

	// Zip usage
	names := []string{"a", "b", "c"}
	numbers := []int{1, 2, 3}
	zipped := Zip(names, numbers)
	for _, pair := range zipped {
		fmt.Printf("Zipped: %s -> %d\n", pair.GetFirst(), pair.GetSecond())
	}

	// Map usage
	stringIntMap := NewMap[string, int]()
	stringIntMap.Set("one", 1)
	stringIntMap.Set("two", 2)

	if value, exists := stringIntMap.Get("one"); exists {
		fmt.Printf("Map value: %d\n", value)
	}

	// Calculator usage
	calc := &Calculator{}
	intResult := CalculateGeneric(calc, 10, 5, "+")
	floatResult := CalculateGeneric(calc, 10.5, 2.3, "-")
	fmt.Printf("Calculator results: %d, %f\n", intResult, floatResult)

	// Comparable usage
	intComp := &IntComparable{value: 42}
	isEqual := ProcessComparable(intComp, 42)
	fmt.Printf("Comparable result: %t\n", isEqual)

	// Type alias usage
	aliasContainer := &StringContainer{}
	aliasContainer.Add("alias test")
	fmt.Printf("Alias container: %s\n", aliasContainer.Get(0))
}

func main() {
	ExampleAdvanced()
}
