package publicapi

// PublicType is an exported type used in the test.
type PublicType struct {
	Value string
}

// UsedMethod is called and should NOT be reported.
func (p *PublicType) UsedMethod() string {
	return p.Value
}

// UnusedMethod is exported but never used - should be reported in strict mode.
func (p *PublicType) UnusedMethod() string {
	return "never called"
}

// UsedPublicFunction is called and should NOT be reported.
func UsedPublicFunction() {
	p := &PublicType{Value: "test"}
	_ = p.UsedMethod()
}

// UnusedPublicFunction is exported but never used - should be reported in strict mode.
func UnusedPublicFunction() {
	println("never called")
}

// unusedPrivateFunction should always be reported (both normal and strict mode).
func unusedPrivateFunction() {
	println("private and unused")
}
