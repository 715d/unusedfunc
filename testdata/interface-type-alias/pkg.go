package pkg

import "github.com/715d/unusedfunc/testdata/interface-type-alias/internal/hscan"

// Scanner is a type alias that re-exports the internal interface
type Scanner = hscan.Scanner

// Scan re-exports the internal function
var Scan = hscan.Scan
