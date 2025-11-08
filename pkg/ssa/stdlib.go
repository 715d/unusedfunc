// Package ssa provides SSA-based reachability analysis for Go programs.
package ssa

import (
	"log/slog"
	"sync"

	"golang.org/x/tools/go/packages"
)

var getStdLibSet = sync.OnceValue(func() Set[string] {
	pkgs, _ := packages.Load(&packages.Config{Mode: packages.NeedName}, "std")
	m := make(Set[string], len(pkgs)+1)
	for _, p := range pkgs {
		m[p.PkgPath] = struct{}{}
	}
	m["unsafe"] = struct{}{} // not in `go list std`
	slog.Debug("loaded std lib packages", "num", len(m))
	return m
})

// isTargetPackage tells the analysis whether to walk into p's functions.
func isTargetPackage(p *packages.Package) bool {
	if _, ok := getStdLibSet()[p.PkgPath]; ok {
		return false
	}
	if p.Module != nil {
		// Modules-on: analyse only our main module, skip all deps.
		return p.Module.Main
	}
	// GOPATH fallback: anything outside stdlib is assumed to be user code.
	return true
}
