package main

import (
	"fmt"
	"regexp"
)

// Multiple init functions and complex package-level variables
// This tests complex initialization patterns

var globalState = make(map[string]interface{})

func init() {
	globalState["initialized"] = true
}

// First IIFE pattern - simple
var simpleMatch = func() func(s string) bool {
	pattern := regexp.MustCompile(`^\d+$`)
	return func(s string) bool {
		return pattern.MatchString(s)
	}
}()

// Another init function
func init() {
	globalState["matcher"] = simpleMatch
}

// Complex IIFE with closure pattern
var matchName = func() func(tableColumn string) (table, column string) {
	nameMatcher := regexp.MustCompile(`^(?:\W?(\w+?)\W?\.)?(?:(\*)|\W?(\w+?)\W?)$`)
	return func(tableColumn string) (table, column string) {
		if matches := nameMatcher.FindStringSubmatch(tableColumn); len(matches) == 4 {
			table = matches[1]
			star := matches[2]
			columnName := matches[3]
			if star != "" {
				return table, star
			}
			return table, columnName
		}
		return "", tableColumn
	}
}()

// More init functions
func init() {
	// Use the function to ensure it's not optimized away
	table, col := matchName("users.id")
	globalState["table"] = table
	globalState["column"] = col
}

// Another complex IIFE
var validator = func() func(string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

	return func(input string) error {
		if emailRegex.MatchString(input) || phoneRegex.MatchString(input) {
			return nil
		}
		return fmt.Errorf("invalid input")
	}
}()

func init() {
	globalState["validator"] = validator
}

// Test the patterns with unexported functions that should be detected as used
type DB struct {
	name string
}

func (db *DB) assignInterfacesToValue(values ...interface{}) {
	fmt.Printf("assignInterfacesToValue: %v\n", values)
}

func (db *DB) Create(value interface{}) *DB {
	db.assignInterfacesToValue(value)
	return db
}

func joins(db *DB, joinType string, query string) *DB {
	fmt.Printf("joins: %s %s\n", joinType, query)
	return db
}

func Joins(query string) *DB {
	return joins(&DB{}, "INNER JOIN", query)
}

func main() {
	fmt.Printf("Global state: %v\n", globalState)

	// Test the matchers
	fmt.Printf("simpleMatch('123'): %v\n", simpleMatch("123"))
	table, col := matchName("users.id")
	fmt.Printf("matchName result: table=%s, column=%s\n", table, col)

	// Test the unexported functions
	db := &DB{name: "test"}
	db.Create("value")
	Joins("users ON users.id = posts.user_id")
}
