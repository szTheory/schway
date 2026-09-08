package cgen

import "github.com/codename-lang/lang/internal/compiler/core"

// Test-only accessors. These make the emitters' reserved sets and the two
// identifier-namespace constructors visible to the external cgen_test package
// without exporting them from the production API.

var (
	MatchFixedNames  = matchFixedNames
	LinearFixedNames = linearFixedNames
)

func CName(name string) string  { return cName(name) }
func CLocal(name string) string { return cLocal(name) }

func AllocateFrom(reserved []string, preferred, category string, ordinal int) string {
	return newCNames(reserved...).allocate(preferred, category, ordinal)
}

// EmitLinearForTest exposes emitLinear directly to the external cgen_test
// package (plan 07-04 Task 3, Test 4): it operates on a single
// core.Function value and is never gated by Emit/EmitNative's
// len(program.Functions) != 1 guard, so a test can drive it on a
// checked+corevalidated function that genuinely carries a core.OpCall
// (A-02: no legal program reaching that guard ever has exactly one
// function, so this is the only way to exercise the arm at all).
func EmitLinearForTest(function core.Function) (string, error) {
	return emitLinear(function)
}

// SetOpCallGroupedArmForTest installs Task 3 Test 4's OpCall-grouped-arm
// mutation seam and returns a restore func. Callers MUST defer it
// immediately.
func SetOpCallGroupedArmForTest(mutate bool) (restore func()) {
	previous := opCallGroupedArmForTest
	opCallGroupedArmForTest = mutate
	return func() { opCallGroupedArmForTest = previous }
}
