package main

import (
	"fmt"
	"reflect"
)

// Statement represents a SQL statement builder
type Statement struct {
	Schema *Schema
	Model  interface{}
}

// Schema represents database schema
type Schema struct {
	Fields []*Field
}

// Field represents a database field
type Field struct {
	Name string
}

// DB represents a database connection
type DB struct {
	Statement *Statement
	Error     error
}

// Association represents a relationship
type Association struct {
	DB           *DB
	Relationship *Relationship
	Error        error
}

// Relationship represents a database relationship
type Relationship struct {
	Field       *Field
	FieldSchema *Schema
}

// Exported method that returns *DB for chaining
func (db *DB) Model(value interface{}) *DB {
	db.Statement.Model = value
	return db
}

// Find method for chaining
func (db *DB) Find(out interface{}, conds ...interface{}) *DB {
	fmt.Println("Find called")
	return db
}

// Exported method that calls unexported method
func (db *DB) Create(value interface{}) *DB {
	// Call unexported method with variadic interface{}
	db.assignInterfacesToValue(value)
	return db
}

// Unexported method with complex interface{} handling and recursion
func (db *DB) assignInterfacesToValue(values ...interface{}) {
	for _, value := range values {
		switch v := value.(type) {
		case []interface{}:
			// Recursive call
			db.assignInterfacesToValue(v...)
		case map[string]interface{}:
			exprs := db.Statement.buildConditions(v)
			if len(exprs) > 0 {
				// Recursive call with built conditions
				db.assignInterfacesToValue(exprs)
			}
		default:
			if db.Statement.Schema != nil {
				fmt.Printf("Assigning value: %v\n", v)
			}
		}
	}
}

// Unexported method on Statement
func (s *Statement) buildConditions(value interface{}) []interface{} {
	// Complex logic that might not be easily traced
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Map {
		return []interface{}{value}
	}
	return nil
}

// Exported method on Association
func (association *Association) Find(out interface{}, conds ...interface{}) error {
	// Calls unexported method that returns *DB
	association.Error = association.buildCondition().Find(out, conds...).Error
	return association.Error
}

// Another exported method
func (association *Association) Count() (int64, error) {
	var count int64
	// Another call to the same unexported method
	association.Error = association.buildCondition().Model(&count).Error
	return count, association.Error
}

// Unexported method that returns *DB
func (association *Association) buildCondition() *DB {
	tx := association.DB.Model(struct{}{})
	// Complex logic
	if association.Relationship != nil && association.Relationship.FieldSchema != nil {
		fmt.Println("Building condition for relationship")
	}
	return tx
}

// Package-level unexported function used by exported functions
func joins(db *DB, joinType string, query string, args ...interface{}) *DB {
	fmt.Printf("Join type: %s, query: %s\n", joinType, query)
	return db
}

// Exported function that calls unexported function
func Joins(query string, args ...interface{}) *DB {
	return joins(&DB{}, "INNER JOIN", query, args...)
}

// Another exported function calling the same unexported function
func InnerJoins(query string, args ...interface{}) *DB {
	return joins(&DB{}, "INNER JOIN", query, args...)
}

func main() {
	// Test DB methods
	db := &DB{
		Statement: &Statement{
			Schema: &Schema{},
		},
	}
	db.Create(map[string]interface{}{
		"name": "test",
		"data": []interface{}{"a", "b", "c"},
	})

	// Test Association methods
	assoc := &Association{
		DB: db,
		Relationship: &Relationship{
			FieldSchema: &Schema{},
		},
	}
	assoc.Find(nil)
	assoc.Count()

	// Test package functions
	Joins("users ON users.id = posts.user_id")
	InnerJoins("categories ON categories.id = posts.category_id")
}
