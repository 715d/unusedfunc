package main

import "fmt"

// DB represents a database connection
type DB struct {
	value string
}

// Association represents a database association
type Association struct {
	db *DB
}

// Exported method that calls unexported method
func (db *DB) Create(value interface{}) *DB {
	// This tests the pattern where exported methods call unexported ones
	db.assignInterfacesToValue(value)
	return db
}

// Unexported method that should be detected as used
// This tests interface assignment patterns
func (db *DB) assignInterfacesToValue(values ...interface{}) {
	fmt.Println("assignInterfacesToValue called with:", values)
	db.value = fmt.Sprintf("%v", values)
}

// Exported method on Association
func (a *Association) Append(values ...interface{}) error {
	// This tests the pattern
	a.saveAssociation(false, values...)
	return nil
}

// Unexported method that should be detected as used
// Helper function for saving associations
func (a *Association) saveAssociation(clear bool, values ...interface{}) {
	fmt.Println("saveAssociation called with clear:", clear, "values:", values)
}

// Another unexported method
// Helper function for building conditions
func (a *Association) buildCondition() *DB {
	fmt.Println("buildCondition called")
	return a.db
}

// Exported method that uses buildCondition
func (a *Association) Find(out interface{}) error {
	db := a.buildCondition()
	fmt.Println("Find called with db:", db)
	return nil
}

// Exported function that calls unexported function
// This tests the Joins/InnerJoins pattern
func Joins(query string) *DB {
	return joins(&DB{}, "JOIN", query)
}

// Unexported function that should be detected as used
// Helper function for joins
func joins(db *DB, joinType string, query string) *DB {
	fmt.Println("joins called with type:", joinType, "query:", query)
	return db
}

// Unexported function called within same package
// Helper function for parsing tags
func parseZeroValueTag(field string) string {
	return "parsed:" + field
}

// Exported function that uses parseZeroValueTag
func ParseField(field string) string {
	return parseZeroValueTag(field)
}

func main() {
	db := &DB{}
	db.Create("test value")

	assoc := &Association{db: db}
	assoc.Append("value1", "value2")
	assoc.Find(nil)

	Joins("users ON users.id = posts.user_id")

	result := ParseField("test_field")
	fmt.Println("ParseField result:", result)
}
