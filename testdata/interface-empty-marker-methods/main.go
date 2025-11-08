package main

import "fmt"

func main() {
	validators := []Validator{
		&EmailValidator{},
		&AgeValidator{},
		&PhoneValidator{},
	}

	for _, v := range validators {
		Process(v)
	}

	// Use oneof pattern.
	msg := &Message{
		Content: &Message_TextField{TextField: "hello"},
	}
	PrintMessage(msg)
}

// Validator is an interface with a marker method.
// The isValidator method exists only for type constraint - it's never called directly.
type Validator interface {
	isValidator()
}

type EmailValidator struct{}

// isValidator is an empty marker method that makes EmailValidator satisfy the Validator interface.
// This should NOT be reported as unused - it exists for interface satisfaction.
func (*EmailValidator) isValidator() {}

type AgeValidator struct{}

// isValidator is an empty marker method that makes AgeValidator satisfy the Validator interface.
// This should NOT be reported as unused - it exists for interface satisfaction.
func (*AgeValidator) isValidator() {}

type PhoneValidator struct{}

// isValidator is an empty marker method that makes PhoneValidator satisfy the Validator interface.
// This should NOT be reported as unused - it exists for interface satisfaction.
func (*PhoneValidator) isValidator() {}

// Process uses the Validator interface but never calls the marker method directly.
func Process(v Validator) {
	fmt.Printf("Processing validator: %T\n", v)
}

// Message demonstrates the protobuf oneof pattern.
type Message struct {
	// Content is one of TextField or ImageField.
	Content isMessage_Content
}

// isMessage_Content is the oneof interface.
// Note: Underscores are intentional to match protobuf-generated code patterns.
//
//nolint:revive // Intentionally using underscores to match protobuf naming
type isMessage_Content interface {
	isMessage_Content()
}

// Message_TextField is a oneof field wrapper.
// Note: Underscores are intentional to match protobuf-generated code patterns.
//
//nolint:revive // Intentionally using underscores to match protobuf naming
type Message_TextField struct {
	TextField string
}

// isMessage_Content is an empty marker method for the oneof pattern.
// This should NOT be reported as unused - it exists for type constraint.
//
//nolint:revive // Intentionally using underscores to match protobuf naming
func (*Message_TextField) isMessage_Content() {}

// Message_ImageField is a oneof field wrapper.
// Note: Underscores are intentional to match protobuf-generated code patterns.
//
//nolint:revive // Intentionally using underscores to match protobuf naming
type Message_ImageField struct {
	ImageField string
}

// isMessage_Content is an empty marker method for the oneof pattern.
// This should NOT be reported as unused - it exists for type constraint.
//
//nolint:revive // Intentionally using underscores to match protobuf naming
func (*Message_ImageField) isMessage_Content() {}

// PrintMessage uses the oneof interface but never calls the marker method.
func PrintMessage(m *Message) {
	switch content := m.Content.(type) {
	case *Message_TextField:
		fmt.Printf("Text: %s\n", content.TextField)
	case *Message_ImageField:
		fmt.Printf("Image: %s\n", content.ImageField)
	}
}

// UnusedValidator is never used anywhere.
type UnusedValidator struct{}

// isValidator is on an unused type, so this SHOULD be reported as unused.
func (*UnusedValidator) isValidator() {}
