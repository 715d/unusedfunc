package main

import "fmt"

type Node struct {
	value    string
	children []*Node
}

type Processor struct {
	data map[string]interface{}
}

// Exported method that starts the processing
func (p *Processor) Process(values ...interface{}) {
	p.assignValuesToData(values...)
}

// Unexported recursive method - should be detected as used
func (p *Processor) assignValuesToData(values ...interface{}) {
	for _, value := range values {
		switch v := value.(type) {
		case []interface{}:
			// Recursive call
			p.assignValuesToData(v...)
		case map[string]interface{}:
			for _, val := range v {
				// Recursive call with nested values
				p.assignValuesToData(val)
			}
		default:
			fmt.Printf("Processing value: %v\n", v)
		}
	}
}

// Another type with recursive pattern
func (n *Node) Walk() {
	n.walkInternal(0)
}

// Unexported recursive method - should be detected as used
func (n *Node) walkInternal(depth int) {
	fmt.Printf("%*s%s\n", depth*2, "", n.value)
	for _, child := range n.children {
		// Recursive call on different instance
		child.walkInternal(depth + 1)
	}
}

// Test with mutual recursion
type Parser struct {
	tokens []string
}

func (p *Parser) Parse() {
	p.parseExpression(0)
}

// Unexported methods with mutual recursion - both should be detected as used
func (p *Parser) parseExpression(pos int) int {
	fmt.Println("parseExpression at", pos)
	if pos < len(p.tokens) && p.tokens[pos] == "(" {
		return p.parseTerm(pos + 1)
	}
	return pos
}

func (p *Parser) parseTerm(pos int) int {
	fmt.Println("parseTerm at", pos)
	if pos < len(p.tokens) && p.tokens[pos] == "[" {
		return p.parseExpression(pos + 1)
	}
	return pos
}

func main() {
	// Test recursive processing
	p := &Processor{data: make(map[string]interface{})}
	p.Process(
		"simple",
		[]interface{}{"nested", "array"},
		map[string]interface{}{
			"key": []interface{}{"deep", "nesting"},
		},
	)

	// Test tree walking
	root := &Node{
		value: "root",
		children: []*Node{
			{value: "child1"},
			{value: "child2", children: []*Node{{value: "grandchild"}}},
		},
	}
	root.Walk()

	// Test mutual recursion
	parser := &Parser{tokens: []string{"(", "[", "]", ")"}}
	parser.Parse()
}
