package cgen

import (
	"github.com/codename-lang/lang/internal/compiler/callgraph"
	"github.com/codename-lang/lang/internal/compiler/core"
)

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

// SetCallBoundaryAttributeSetForTest installs
// TestEmittedAttributeSetCommentIsDerivedNotLiteral's own derivation
// seam (D-04-12/D-11-09) and returns a restore func. Callers MUST defer
// it immediately.
func SetCallBoundaryAttributeSetForTest(set []string) (restore func()) {
	previous := callBoundaryAttributeSetForTest
	callBoundaryAttributeSetForTest = set
	return func() { callBoundaryAttributeSetForTest = previous }
}

// SetInvocationPreflightBypassForTest temporarily removes the bounded
// invocation preflight check. Its companion observation proves the bypass
// reaches generated-C serialization, so the guard test cannot pass inertly.
func SetInvocationPreflightBypassForTest(bypass bool) (restore func()) {
	previous := invocationPreflightBypassForTest
	invocationPreflightBypassForTest = bypass
	invocationSerializationReachedForTest = false
	return func() {
		invocationPreflightBypassForTest = previous
	}
}

// InvocationSerializationReachedForTest reports whether emitProgram began C
// serialization since the invocation-preflight test seam was installed.
func InvocationSerializationReachedForTest() bool {
	return invocationSerializationReachedForTest
}

// InvocationPathNodeCountForTest runs the same validated preflight used by
// emitProgram and exposes its bounded occurrence count to cgen's black-box
// tests without making a production configuration surface.
func InvocationPathNodeCountForTest(program core.Program) (int, error) {
	if _, err := callgraph.Order(program); err != nil {
		return 0, err
	}
	entry, err := callgraph.EntryFunction(program)
	if err != nil {
		return 0, err
	}
	nodes, err := preflightInvocationPathTable(program, entry)
	return len(nodes), err
}

// ExecutionOutputSizeForTest exposes the production preflight estimator so
// boundary tests can derive N-1/N limits without copying its arithmetic.
func ExecutionOutputSizeForTest(program core.Program) (int, error) {
	entry, err := callgraph.EntryFunction(program)
	if err != nil {
		return 0, err
	}
	nodes, err := preflightInvocationPathTable(program, entry)
	if err != nil {
		return 0, err
	}
	paths, err := buildInvocationPathTable(entry.ID, nodes)
	if err != nil {
		return 0, err
	}
	byID := make(map[string]core.Function, len(program.Functions))
	for _, function := range program.Functions {
		byID[function.ID] = function
	}
	return schema2ExecutionDocumentSize(entry, nodes, paths, byID)
}

// SetExecutionOutputLimitForTest installs a boundary-only limit and resets the
// serialization observation, proving refusal occurs before C output begins.
func SetExecutionOutputLimitForTest(limit int) (restore func()) {
	previous := executionOutputLimit
	executionOutputLimit = limit
	invocationSerializationReachedForTest = false
	return func() { executionOutputLimit = previous }
}

// EmitProgramForTest exposes the whole-program assembler without the public
// EmitNative corevalidation wrapper so ordering controls can feed malformed
// graph/entry artifacts directly to its first two validation gates.
func EmitProgramForTest(program core.Program) (string, error) {
	return emitProgram(program, false)
}
