package main

import (
	"fmt"
	"regexp"
)

// This tests a package-level closure pattern
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

// DB type to test methods in this context
type DB struct {
	value string
}

// Association type
type Association struct {
	db *DB
}

// Test if methods are still detected when this pattern exists
func (db *DB) Create(value interface{}) *DB {
	db.assignInterfacesToValue(value)
	return db
}

// Unexported method that should be detected as used
func (db *DB) assignInterfacesToValue(values ...interface{}) {
	fmt.Printf("assignInterfacesToValue called with %v\n", values)
}

// Another unexported method
func (a *Association) buildCondition() *DB {
	fmt.Println("buildCondition called")
	return a.db
}

// Exported method using buildCondition
func (a *Association) Find(out interface{}) error {
	a.buildCondition()
	return nil
}

// Package-level unexported function
func joins(db *DB, joinType string, query string) *DB {
	fmt.Printf("joins called: %s %s\n", joinType, query)
	return db
}

// Exported function calling joins
func Joins(query string) *DB {
	return joins(&DB{}, "INNER JOIN", query)
}

// Another unexported function
func parseZeroValueTag(tag string) string {
	return "parsed:" + tag
}

// Exported function using parseZeroValueTag
func ParseTag(tag string) string {
	return parseZeroValueTag(tag)
}

// Unexported method
func (a *Association) saveAssociation(clear bool, values ...interface{}) {
	fmt.Printf("saveAssociation: clear=%v, values=%v\n", clear, values)
}

// Exported method using saveAssociation
func (a *Association) Append(values ...interface{}) {
	a.saveAssociation(false, values...)
}

func main() {
	// Use the closure
	table, column := matchName("users.id")
	fmt.Printf("Parsed: table=%s, column=%s\n", table, column)

	// Test DB methods
	db := &DB{}
	db.Create("test")

	// Test Association methods
	assoc := &Association{db: db}
	assoc.Find(nil)
	assoc.Append("a", "b", "c")

	// Test package functions
	Joins("users ON users.id = posts.user_id")
	result := ParseTag("zero")
	fmt.Println("Parse result:", result)
}
