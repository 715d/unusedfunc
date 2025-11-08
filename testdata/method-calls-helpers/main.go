package main

import (
	"fmt"
)

// This tests the pattern where methods call unexported helper functions
type DB struct {
	data string
}

// Unexported functions that should be detected as USED because they're called from methods
func assignInterfacesToValue(values ...interface{}) {
	fmt.Println("assignInterfacesToValue called")
}

func buildCondition() string {
	return "condition built"
}

func joins(query string) string {
	return "joined: " + query
}

func parseZeroValueTag(field string) string {
	return "parsed: " + field
}

func saveAssociation(clear bool, values ...interface{}) {
	fmt.Println("saveAssociation called")
}

// Exported methods that call the unexported functions
func (db *DB) Find(out interface{}, conds ...interface{}) *DB {
	// This replicates the pattern from association.go
	condition := buildCondition()
	fmt.Println("Condition:", condition)
	assignInterfacesToValue(out, conds)
	return db
}

func (db *DB) Count(count *int64) *DB {
	// This replicates the pattern from association.go
	condition := buildCondition()
	fmt.Println("Condition:", condition)
	return db
}

func (db *DB) LeftJoin(query string, args ...interface{}) *DB {
	// This replicates the pattern from chainable_api.go
	result := joins(query)
	fmt.Println("Result:", result)
	assignInterfacesToValue(args...)
	return db
}

func (db *DB) SoftDelete(field string) *DB {
	// This replicates the pattern from soft_delete.go
	parsed := parseZeroValueTag(field)
	fmt.Println("Parsed:", parsed)
	return db
}

func (db *DB) SaveAssoc(clear bool, values ...interface{}) {
	// This replicates the pattern from association.go
	saveAssociation(clear, values...)
}

// This function is genuinely unused and should be reported
func actuallyUnused() {
	fmt.Println("This should be reported as unused")
}

func main() {
	db := &DB{data: "test"}
	var result interface{}
	var count int64

	// Use the methods that call the unexported functions
	db.Find(&result).Count(&count).LeftJoin("users").SoftDelete("deleted_at")
	db.SaveAssoc(false, "value1", "value2")

	fmt.Println("Done")
}
