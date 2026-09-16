package main

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type T struct{}

var calls int

func (T) Used() {
	calls++
}

func (T) unusedControl() {}

type fmtControl struct{}

func (fmtControl) UnusedExported() {
	fmtControl{}.unusedPrivateHelper()
}

func (fmtControl) unusedPrivateHelper() {}

type jsonControl struct{}

func (jsonControl) UnusedExported() {
	jsonControl{}.unusedPrivateHelper()
}

func (jsonControl) unusedPrivateHelper() {}

func unusedFunctionControl() {}

func main() {
	fmt.Println(T{})
	fmt.Println(fmtControl{})
	_, _ = json.Marshal(jsonControl{})
	_ = reflect.TypeOf(0)

	for _, method := range reflect.ValueOf(T{}).Methods() {
		method.Call(nil)
	}
	for method := range reflect.TypeOf(T{}).Methods() {
		method.Func.Call([]reflect.Value{reflect.ValueOf(T{})})
	}
	value := reflect.ValueOf(T{})
	for i := 0; i < value.NumMethod(); i++ {
		value.Method(i).Call(nil)
	}
	if calls != 3 {
		panic("unexpected reflected calls")
	}
}
