package main

import "errors"

type AsTypeMarker interface {
	error
	asTypeMarker()
}

type AsTypeE struct{}

func (AsTypeE) Error() string {
	return "AsType wrapped"
}

func (AsTypeE) asTypeMarker() {}

func (AsTypeE) unusedAsTypeControl() {}

type AsTypeW struct{}

func (AsTypeW) Error() string {
	return "AsType wrapper"
}

func (AsTypeW) Unwrap() error {
	return AsTypeE{}
}

type AsTypeAsHookControl struct{}

func (AsTypeAsHookControl) As(any) bool {
	AsTypeAsHookControl{}.unusedAsHookHelper()
	return false
}

func (AsTypeAsHookControl) unusedAsHookHelper() {}

type AsTypeUnwrapHookControl struct{}

func (AsTypeUnwrapHookControl) Unwrap() error {
	AsTypeUnwrapHookControl{}.unusedUnwrapHookHelper()
	return nil
}

func (AsTypeUnwrapHookControl) unusedUnwrapHookHelper() {}

type AsTypeUnwrapManyHookControl struct{}

func (AsTypeUnwrapManyHookControl) Unwrap() []error {
	AsTypeUnwrapManyHookControl{}.unusedUnwrapManyHookHelper()
	return nil
}

func (AsTypeUnwrapManyHookControl) unusedUnwrapManyHookHelper() {}

type AsMarker interface {
	error
	asMarker()
}

type AsE struct{}

func (AsE) Error() string {
	return "As wrapped"
}

func (AsE) asMarker() {}

func (AsE) unusedAsControl() {}

type AsW struct{}

func (AsW) Error() string {
	return "As wrapper"
}

func (AsW) Unwrap() error {
	return AsE{}
}

func main() {
	checkLateErrors()
	if _, ok := errors.AsType[AsTypeMarker](AsTypeW{}); !ok {
		panic("AsType did not match")
	}

	var marker AsMarker
	if !errors.As(AsW{}, &marker) {
		panic("As did not match")
	}
}
