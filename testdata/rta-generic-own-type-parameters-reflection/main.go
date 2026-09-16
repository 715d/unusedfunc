package main

import "reflect"

type T struct{}

func (T) Used[X any](value X) X {
	return value
}

func (T) Unused[X any](value X) X {
	return value
}

func unusedControl() {}

func main() {
	var value T
	if value.Used(3) != 3 {
		panic("unexpected generic result")
	}
	_ = reflect.TypeOf(value)
}
