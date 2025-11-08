package main

import (
	"fmt"

	pkg "github.com/715d/unusedfunc/testdata/interface-type-alias"
)

// ProductionType implements the Scanner interface via type alias
type ProductionType struct{}

func (p *ProductionType) ScanRedis(s string) error {
	fmt.Println("ProductionType.ScanRedis called with:", s)
	return nil
}

func main() {
	// Use through the type alias
	p := &ProductionType{}
	_ = pkg.Scan(p)
}
