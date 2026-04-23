// Package fournftest provides test helpers for 4NF validation.
package fournftest

import (
	"testing"

	"github.com/blobmasterbrian/fournf"
)

// ValidateGraph is a test helper that loads the schema graph and fails the
// test if any 4NF violations are found. Use this in CI for a safety net
// independent of code generation.
//
//	func TestFourNF(t *testing.T) {
//	    fournftest.ValidateGraph(t, "./schema", "mymodule/ent")
//	}
func ValidateGraph(tb testing.TB, schemaDir, pkg string) {
	tb.Helper()
	violations, err := fournf.ValidateGraph(schemaDir, pkg)
	if err != nil {
		tb.Fatalf("loading schema graph: %v", err)
	}
	for _, v := range violations {
		tb.Errorf("4NF violation: %s", v)
	}
}
