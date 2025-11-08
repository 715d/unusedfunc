package main

import (
	"fmt"
)

type QueryInterface interface {
	QueryClauses(field string) []string
	UpdateClauses(field string) []string
}

type MyType struct {
	Field string
}

// Interface methods that call unexported helper functions
func (m *MyType) QueryClauses(field string) []string {
	parsed := parseZeroValueTag(field)
	condition := buildCondition()
	return []string{parsed, condition}
}

func (m *MyType) UpdateClauses(field string) []string {
	assigned := assignInterfacesToValue(field)
	joined := joins("update_query")
	return []string{assigned, joined}
}

// These unexported functions should be detected as USED because they're called
// from interface methods that are invoked through interface dispatch

func parseZeroValueTag(field string) string {
	return "parsed_" + field
}

func buildCondition() string {
	return "condition_built"
}

func assignInterfacesToValue(field string) string {
	return "assigned_" + field
}

func joins(query string) string {
	return "joined_" + query
}

func saveAssociation(t bool) string {
	return fmt.Sprintf("saved_association_%t", t)
}

// Function that calls interface methods through interface dispatch
func processWithInterface(qi QueryInterface, field string) {
	// Interface method dispatch
	queryClauses := qi.QueryClauses(field)
	updateClauses := qi.UpdateClauses(field)

	fmt.Println("Query clauses:", queryClauses)
	fmt.Println("Update clauses:", updateClauses)
}

func main() {
	// Create instance and call through interface
	myType := &MyType{Field: "test"}

	// This calls the interface methods through interface dispatch
	processWithInterface(myType, "deleted_at")

	fmt.Println("Done")
}
