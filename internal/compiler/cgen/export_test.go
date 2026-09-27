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
	return schema2ExecutionDocumentSize(entry, nodes, paths, byID, deriveProgramLiveResources(program), program.DataTypes)
}

// SetProgramLiveResourcesForTest replaces the surviving emitter's derived
// resource collection for a mutation control. It is test-only and never
// creates resource ownership semantics on a production path.
func SetProgramLiveResourcesForTest(resources []string) (restore func()) {
	previous := programLiveResourcesForTest
	programLiveResourcesForTest = append([]string(nil), resources...)
	return func() { programLiveResourcesForTest = previous }
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
