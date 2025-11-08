package main

import "fmt"

// DB represents a database connection with error tracking
type DB struct {
	Error error
	value string
}

// Association represents a relationship with error tracking
type Association struct {
	DB    *DB
	Error error
	name  string
}

// Find is an exported method that chains method calls
func (association *Association) Find(out interface{}, conds ...interface{}) error {
	if association.Error == nil {
		// This exact pattern matches the code: method().method().field
		association.Error = association.buildCondition().Find(out, conds...).Error
	}
	return association.Error
}

// buildCondition is an unexported method that returns *DB
func (association *Association) buildCondition() *DB {
	fmt.Printf("buildCondition called for %s\n", association.name)
	// Return a DB instance
	return &DB{value: "conditioned"}
}

// Find method on DB to complete the chain
func (db *DB) Find(out interface{}, conds ...interface{}) *DB {
	fmt.Printf("DB.Find called with value: %s\n", db.value)
	return db
}

// Another pattern: Count method
func (association *Association) Count() (int64, error) {
	var count int64
	// Same chaining pattern
	association.Error = association.buildCondition().Model(&count).Error
	return count, association.Error
}

// Model method on DB
func (db *DB) Model(value interface{}) *DB {
	fmt.Printf("DB.Model called\n")
	return db
}

// Test the saveAssociation pattern
func (association *Association) Append(values ...interface{}) error {
	association.saveAssociation(false, values...)
	return association.Error
}

// saveAssociation is unexported and called with variadic args
func (association *Association) saveAssociation(clear bool, values ...interface{}) {
	fmt.Printf("saveAssociation called with clear=%v, values=%v\n", clear, values)
}

// Test the assignInterfacesToValue pattern with recursion
type Statement struct {
	values []interface{}
}

type DBWithStatement struct {
	Statement *Statement
}

func (db *DBWithStatement) Create(value interface{}) {
	db.assignInterfacesToValue(value)
}

// Recursive unexported method
func (db *DBWithStatement) assignInterfacesToValue(values ...interface{}) {
	for _, value := range values {
		switch v := value.(type) {
		case []interface{}:
			// Recursive call
			db.assignInterfacesToValue(v...)
		default:
			fmt.Printf("Processing value: %v\n", v)
		}
	}
}

// Test the joins pattern
func Joins(query string, args ...interface{}) *DB {
	return joins(&DB{}, "INNER", query, args...)
}

func InnerJoins(query string, args ...interface{}) *DB {
	return joins(&DB{}, "INNER", query, args...)
}

// Unexported function called by multiple exported functions
func joins(db *DB, joinType string, query string, args ...interface{}) *DB {
	fmt.Printf("joins called: type=%s, query=%s\n", joinType, query)
	return db
}

// Test parseZeroValueTag pattern
type Field struct {
	Tag string
}

type SoftDelete struct{}

func (sd SoftDelete) Setup(f *Field) {
	result := parseZeroValueTag(f)
	fmt.Printf("Setup with result: %s\n", result)
}

// Unexported function
func parseZeroValueTag(f *Field) string {
	return "parsed:" + f.Tag
}

func main() {
	// Test Association methods
	assoc := &Association{
		DB:   &DB{},
		name: "test-assoc",
	}
	assoc.Find(nil, "condition1")
	assoc.Count()
	assoc.Append("value1", "value2")

	// Test DB methods
	dbStmt := &DBWithStatement{
		Statement: &Statement{},
	}
	dbStmt.Create([]interface{}{"a", []interface{}{"b", "c"}})

	// Test joins
	Joins("users u ON u.id = p.user_id")
	InnerJoins("posts p ON p.id = c.post_id")

	// Test soft delete
	sd := SoftDelete{}
	sd.Setup(&Field{Tag: "test"})
}
