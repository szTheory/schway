package cgen

import "github.com/codename-lang/lang/internal/compiler/core"

// Test-only accessors. These make the emitters' reserved sets and the two
// identifier-namespace constructors visible to the external cgen_test package
// without exporting them from the production API.

var (
	MatchFixedNames  = matchFixedNames
	LinearFixedNames = linearFixedNames
)

// SelectsByPointerLowering exposes D-05-02's structural selection predicate
// to cgen_test for TestPhase5ByPointerLoweringIsAdditive, following this
// file's existing seam convention (never exported from the production API).
func SelectsByPointerLowering(function core.Function, linear *core.LinearBody) bool {
	return selectsByPointerLowering(function, linear)
}

func CName(name string) string  { return cName(name) }
func CLocal(name string) string { return cLocal(name) }

func AllocateFrom(reserved []string, preferred, category string, ordinal int) string {
	return newCNames(reserved...).allocate(preferred, category, ordinal)
}
