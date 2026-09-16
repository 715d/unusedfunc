// Package genericpublicapireachability verifies that normal-mode public generic APIs
// retain the helpers reached from their bodies.
package genericpublicapireachability

type NewBox struct{}

func (NewBox) Public[T any](value T) T {
	newDirectHelper()
	NewBox{}.privateHelper()
	NewBox{}.explicitHelper[int]()
	asmHelper()
	return value
}

func newDirectHelper() {
	newDirectLeaf()
}

func newDirectLeaf() {}

func (NewBox) privateHelper() {
	newMethodLeaf()
}

func newMethodLeaf() {}

func (NewBox) explicitHelper[T any]() {
	newExplicitLeaf()
}

func newExplicitLeaf() {}

func (NewBox) ValueReferences[T any]() func() {
	assigned := newAssignedHelper
	assigned()
	acceptCallback(newPassedHelper)
	instantiated := newGenericValueHelper[int]
	instantiated()
	return newReturnedHelper
}

func newAssignedHelper() {}

func newPassedHelper() {}

func newGenericValueHelper[T any]() {}

func newReturnedHelper() {}

func acceptCallback(callback func()) {
	callback()
}

func (NewBox) MethodReferences[T any]() func() {
	method := NewBox{}.valueMethod
	method()
	return NewBox{}.returnedMethod
}

func (NewBox) valueMethod() {}

func (NewBox) returnedMethod() {}

type OldBox[T any] struct{}

func (OldBox[T]) Public(value T) T {
	oldDirectHelper()
	OldBox[T]{}.privateHelper()
	return value
}

func oldDirectHelper() {
	oldDirectLeaf()
}

func oldDirectLeaf() {}

func (OldBox[T]) privateHelper() {
	oldMethodLeaf()
}

func oldMethodLeaf() {}

func PublicFunction[T any](value T) T {
	functionDirectHelper()
	functionExplicitHelper[int]()
	dispatch(worker{})
	func() runner { return literalWorker{} }().run()
	cb := callback(dispatch)
	cb(callbackWorker{})
	return value
}

func functionDirectHelper() {
	functionDirectLeaf()
}

func functionDirectLeaf() {}

func functionExplicitHelper[T any]() {
	functionExplicitLeaf()
}

func functionExplicitLeaf() {}

type callback func(runner)

type literalWorker struct{}

func (literalWorker) run() { literalLeaf() }
func literalLeaf()         {}

type callbackWorker struct{}

func (callbackWorker) run() { callbackLeaf() }
func callbackLeaf()         {}

type runner interface {
	run()
}

type worker struct{}

func dispatch(value runner) {
	value.run()
}

func (worker) run() {
	dispatchLeaf()
}

func dispatchLeaf() {}

func PublicVariadic[T any]() {
	dispatchAll(variadicWorker{})
}

func dispatchAll(values ...runner) {
	for _, value := range values {
		value.run()
	}
}

type variadicWorker struct{}

func (variadicWorker) run() {
	variadicLeaf()
}

func variadicLeaf() {}

func PublicAssignment[T any]() {
	var value runner
	value = assignedWorker{}
	dispatch(value)
}

type assignedWorker struct{}

func (assignedWorker) run() {
	assignedLeaf()
}

func assignedLeaf() {}

func PublicReturn[T any]() runner {
	return returnedWorker{}
}

type returnedWorker struct{}

func (returnedWorker) run() {
	returnedLeaf()
}

func returnedLeaf() {}

type unrelatedWorker struct{}

func (unrelatedWorker) run() {
	unrelatedLeaf()
}

func unrelatedLeaf() {}

func asmHelper()

func unusedControl() {}

func unusedGenericControl[T any]() {}

func (NewBox) unusedNewMethodControl() {}

func (OldBox[T]) unusedOldMethodControl() {}
