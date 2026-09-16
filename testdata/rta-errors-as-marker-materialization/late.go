package main

import "errors"

type lateOuter struct{ inner lateInner }

func (lateOuter) Error() string   { return "outer" }
func (w lateOuter) Unwrap() error { return w.inner }

type lateInner struct{}

func (lateInner) Error() string { return "inner" }
func (lateInner) Unwrap() error { return lateLeaf() }
func lateLeaf() error           { return AsE{} }

func makeLateError() error  { return makeLateError2() }
func makeLateError2() error { return makeLateError3() }
func makeLateError3() error { return makeLateError4() }
func makeLateError4() error { return makeLateError5() }
func makeLateError5() error { return lateOuter{} }

type lateTypeOuter struct{ inner lateTypeInner }

func (lateTypeOuter) Error() string   { return "outer" }
func (w lateTypeOuter) Unwrap() error { return w.inner }

type lateTypeInner struct{}

func (lateTypeInner) Error() string   { return "inner" }
func (lateTypeInner) Unwrap() []error { return []error{lateTypeLeaf()} }
func lateTypeLeaf() error             { return AsTypeE{} }

func makeLateTypeError() error  { return makeLateTypeError2() }
func makeLateTypeError2() error { return makeLateTypeError3() }
func makeLateTypeError3() error { return makeLateTypeError4() }
func makeLateTypeError4() error { return makeLateTypeError5() }
func makeLateTypeError5() error { return lateTypeOuter{} }

type customAs struct{}

func (customAs) Error() string { return "custom" }
func (customAs) As(target any) bool {
	switch p := target.(type) {
	case *AsMarker:
		*p = AsE{}
	case *AsTypeMarker:
		*p = AsTypeE{}
	default:
		return false
	}
	return true
}

func checkLateErrors() {
	var marker AsMarker
	if !errors.As(makeLateError(), &marker) {
		panic("late As did not match")
	}
	if _, ok := errors.AsType[AsTypeMarker](makeLateTypeError()); !ok {
		panic("late AsType did not match")
	}
	if !errors.As(customAs{}, &marker) {
		panic("custom As did not match")
	}
	if _, ok := errors.AsType[AsTypeMarker](customAs{}); !ok {
		panic("custom AsType did not match")
	}
}
