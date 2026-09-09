package corevalidate

import (
	"fmt"
	"go/parser"
	"go/token"
	"sort"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// syntheticProgram builds a fully structurally-valid core.Program directly
// from an adjacency description, keyed by short node NAMES rather than full
// IDs -- one core.Function per name (mirroring
// pathoracle_test.go:188-203's hand-built-artifact convention:
// "{functionID}:place:N", "{functionID}:point:linear:N"), and one
// core.OpCall operation per outgoing edge, chained sequentially so each
// function stays structurally valid (unique, correctly-ordered operation
// and place IDs, non-empty CalleeID, a final core.OpReturn). It never
// imports syntax or check (see
// TestCyclePeerTestFileImportsStayIndependent below) and must produce
// programs that pass every structural validation corevalidate's own
// pipeline runs BEFORE the cycle peer -- so a rejection from Validate is
// unambiguously the peer's own cycle refusal, never a malformed fixture.
//
// Every node named as either a caller (a map key) or a callee (a value in
// some other node's edge list) is declared as its own function, even if it
// carries no outgoing edges of its own. Both the node names and each
// node's own outgoing-edge list are iterated in SORTED order (never raw
// map range order) so a map-backed builder cannot pass a determinism
// assertion by luck.
func syntheticProgram(edges map[string][]string) core.Program {
	declared := make(map[string]bool)
	names := make([]string, 0, len(edges))
	for name, callees := range edges {
		if !declared[name] {
			declared[name] = true
			names = append(names, name)
		}
		for _, callee := range callees {
			if !declared[callee] {
				declared[callee] = true
				names = append(names, callee)
			}
		}
	}
	sort.Strings(names)

	functions := make([]core.Function, 0, len(names))
	for _, name := range names {
		functions = append(functions, syntheticFunction(name, sortedCopy(edges[name])))
	}

	return core.Program{
		Schema: core.Schema1, Module: "cyclepeer", ModuleID: "s1:cyclepeer:module:cyclepeer",
		Functions: functions,
	}
}

// syntheticFunctionID is syntheticProgram's own node-name-to-function-ID
// convention, exported to this file's own siblings (the unresolved-callee
// and self-edge cases below) so they can name a callee consistently
// without re-deriving the scheme.
func syntheticFunctionID(name string) string { return "s1:cyclepeer:fn:" + name }

func sortedCopy(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}

// syntheticFunction builds one fully structurally-valid core.Function
// named `name`, whose body calls each of `callees` (by node name) in
// sequence, then returns. A function with zero callees is a bare
// borrow-free straight-line return.
func syntheticFunction(name string, callees []string) core.Function {
	fnID := syntheticFunctionID(name)
	typeID := fnID + ":type:0"
	parameterID := fnID + ":place:0"
	typeFact := core.TypeFact{
		ID: typeID, Shape: core.TypeRef{Constructor: "Byte", Arguments: []core.TypeRef{}},
		Abilities:         []core.Ability{core.AbilityCopy, core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape},
		NegativeWitnesses: []core.AbilityWitness{},
	}
	places := []core.Place{{ID: parameterID, Name: "value", TypeID: typeID}}

	var operations []core.LinearOperation
	currentSource := parameterID
	for i, callee := range callees {
		targetID := fmt.Sprintf("%s:place:%d", fnID, i+1)
		places = append(places, core.Place{ID: targetID, Name: fmt.Sprintf("result%d", i), TypeID: typeID})
		operations = append(operations, core.LinearOperation{
			ID: fmt.Sprintf("%s:op:%d", fnID, i), PointID: fmt.Sprintf("%s:point:linear:%d", fnID, i),
			Kind: core.OpCall, SourceID: currentSource, TargetID: targetID, TypeID: typeID,
			CalleeID: syntheticFunctionID(callee),
		})
		currentSource = targetID
	}
	returnIndex := len(callees)
	operations = append(operations, core.LinearOperation{
		ID: fmt.Sprintf("%s:op:%d", fnID, returnIndex), PointID: fmt.Sprintf("%s:point:linear:%d", fnID, returnIndex),
		Kind: core.OpReturn, SourceID: currentSource, TypeID: typeID,
	})

	return core.Function{
		ID: fnID, Name: name,
		EntryPointID: fnID + ":point:entry", ReturnPointID: fnID + ":point:return",
		Parameter: core.Parameter{ID: parameterID, Name: "value", Type: "Byte"}, ReturnType: "Byte",
		Linear: &core.LinearBody{
			ID:         fnID + ":linear",
			Types:      []core.TypeFact{typeFact},
			Places:     places,
			Operations: operations,
		},
	}
}

// hasProblem reports whether problems contains one with the exact code.
func hasProblem(problems []Problem, code string) bool {
	for _, problem := range problems {
		if problem.Code == code {
			return true
		}
	}
	return false
}

// TestCyclePeerTestFileImportsStayIndependent asserts THIS file's own Go
// import list -- parsed from source, never assumed -- contains neither
// "compiler/syntax" nor "compiler/check": the peer's tests must never be
// written through the parser (this project's recorded failure shape), and
// building the peer's inputs through the producer's own front end is what
// makes a peer inert.
func TestCyclePeerTestFileImportsStayIndependent(t *testing.T) {
	path := testsupport.ProjectPath("internal", "compiler", "corevalidate", "corevalidate_cycle_peer_test.go")
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatal(err)
	}
	for _, imported := range file.Imports {
		importPath := strings.Trim(imported.Path.Value, `"`)
		if strings.HasSuffix(importPath, "/compiler/syntax") || strings.HasSuffix(importPath, "/compiler/check") {
			t.Fatalf("corevalidate_cycle_peer_test.go imports %s, making the peer's own tests inert", importPath)
		}
	}
}

// TestCyclePeerRefusesSyntheticMutualCycle is D-07-19's headline proof: a
// hand-built synthetic two-node mutual cycle that check never produced is
// refused by corevalidate's OWN independent traversal, with the shared
// inert code core.CallGraphCycle.
func TestCyclePeerRefusesSyntheticMutualCycle(t *testing.T) {
	program := syntheticProgram(map[string][]string{
		"a": {"b"},
		"b": {"a"},
	})
	result := Validate(program)
	if result.Valid {
		t.Fatalf("expected a synthetic mutual cycle to be refused, got valid: %+v", result)
	}
	if !hasProblem(result.Problems, core.CallGraphCycle) {
		t.Fatalf("expected %s, got %+v", core.CallGraphCycle, result.Problems)
	}
}

// TestCyclePeerRefusesSelfEdge proves a length-1 synthetic cycle (a
// function calling itself) is refused too -- the peer's gray-marking of a
// root before visiting its own children catches a self-edge without any
// dedicated seam.
func TestCyclePeerRefusesSelfEdge(t *testing.T) {
	program := syntheticProgram(map[string][]string{"loop": {"loop"}})
	result := Validate(program)
	if result.Valid {
		t.Fatalf("expected a synthetic self-edge cycle to be refused, got valid: %+v", result)
	}
	if !hasProblem(result.Problems, core.CallGraphCycle) {
		t.Fatalf("expected %s, got %+v", core.CallGraphCycle, result.Problems)
	}
}

// TestCyclePeerRefusesCycleWithNoEntryPoint proves roots are ALL declared
// functions, never merely an entry point check happens to reach: the
// cyclic pair here is never named as a callee by any OTHER function in the
// program, so nothing would ever "reach" it under an entry-point-rooted
// traversal -- yet it is still refused (fail-closed).
func TestCyclePeerRefusesCycleWithNoEntryPoint(t *testing.T) {
	program := syntheticProgram(map[string][]string{
		"isolated_a": {"isolated_b"},
		"isolated_b": {"isolated_a"},
		"unrelated":  nil,
	})
	result := Validate(program)
	if result.Valid {
		t.Fatalf("expected an unreachable synthetic cycle to still be refused, got valid: %+v", result)
	}
	if !hasProblem(result.Problems, core.CallGraphCycle) {
		t.Fatalf("expected %s, got %+v", core.CallGraphCycle, result.Problems)
	}
}

// TestCyclePeerAcceptsAcyclicDiamond is the accepting-path sibling: a
// shared-leaf diamond (two callers of the same callee) is NOT a cycle, and
// must validate cleanly. This is also the base fixture Test 4's mutation
// matrix (below) wrongly refuses once the gray-versus-visited seam is
// engaged.
func TestCyclePeerAcceptsAcyclicDiamond(t *testing.T) {
	program := diamondSyntheticProgram()
	result := Validate(program)
	if !result.Valid {
		t.Fatalf("expected an acyclic diamond to validate cleanly, got %+v", result)
	}
}

// diamondSyntheticProgram builds top -> {left, right} -> shared -> leaf: a
// diamond whose shared leaf is legitimately visited twice via two
// independent callers, never a cycle. This is the ONLY fixture shape that
// can kill the gray-versus-visited mutation (mirrors callgraph's own
// deep_diamond_acyclic.lang rationale): a straight chain never revisits a
// node, so the black-vs-gray distinction is decorative on a chain.
func diamondSyntheticProgram() core.Program {
	return syntheticProgram(map[string][]string{
		"top":    {"left", "right"},
		"left":   {"shared"},
		"right":  {"shared"},
		"shared": {"leaf"},
		"leaf":   nil,
	})
}

// TestCyclePeerRejectsMalformedProgramBeforeReachingPeer proves the peer
// runs strictly AFTER structural validation: an acyclic but structurally
// malformed synthetic program (an operation referencing an unknown place)
// is rejected by the PRE-EXISTING structural validation, by exact code,
// never by the cycle peer.
func TestCyclePeerRejectsMalformedProgramBeforeReachingPeer(t *testing.T) {
	program := syntheticProgram(map[string][]string{"solo": nil})
	program.Functions[0].Linear.Operations[0].SourceID = "does-not-exist"
	result := Validate(program)
	if result.Valid {
		t.Fatalf("expected the malformed synthetic program to be rejected, got valid")
	}
	if hasProblem(result.Problems, core.CallGraphCycle) {
		t.Fatalf("expected a structural rejection, not the cycle peer, got %+v", result.Problems)
	}
	if !hasProblem(result.Problems, "core.unknown_place") {
		t.Fatalf("expected core.unknown_place, got %+v", result.Problems)
	}
}

// TestCyclePeerRefusesUnresolvedCallee is Task 2's forged-artifact case: an
// OpCall whose CalleeID names no declared function must refuse through
// 07-03's unresolved-callee identity (core.CallCalleeUnresolved), never be
// silently dropped -- a dropped edge is precisely how a cycle escapes
// detection. This is a forged artifact no parser can produce, exercised
// directly against a hand-built synthetic core.Program, exactly as
// callgraph's own TestOrderRefusesUnresolvedCallee (07-06) exercises
// callgraph.Order.
func TestCyclePeerRefusesUnresolvedCallee(t *testing.T) {
	program := syntheticProgram(map[string][]string{"lonely": nil})
	function := &program.Functions[0]
	targetID := function.ID + ":place:1"
	typeID := function.ID + ":type:0"
	function.Linear.Places = append(function.Linear.Places, core.Place{ID: targetID, Name: "result0", TypeID: typeID})
	function.Linear.Operations = []core.LinearOperation{
		{
			ID: function.ID + ":op:0", PointID: function.ID + ":point:linear:0",
			Kind: core.OpCall, SourceID: function.Parameter.ID, TargetID: targetID, TypeID: typeID,
			CalleeID: "s1:cyclepeer:fn:does-not-exist",
		},
		{
			ID: function.ID + ":op:1", PointID: function.ID + ":point:linear:1",
			Kind: core.OpReturn, SourceID: targetID, TypeID: typeID,
		},
	}
	result := Validate(program)
	if result.Valid {
		t.Fatalf("expected an unresolved callee to be refused, got valid")
	}
	if !hasProblem(result.Problems, core.CallCalleeUnresolved) {
		t.Fatalf("expected %s, got %+v", core.CallCalleeUnresolved, result.Problems)
	}
}

// TestCyclePeerMutationMatrix is Task 3 Test 4: the peer's own
// gray-versus-visited fault-injection seam (peerGrayVsVisitedMutationForTest),
// proven load-bearing against diamondSyntheticProgram -- a straight chain
// could never kill this mutation, since reverse postorder never revisits a
// node on a chain. Follows the four-beat mutation-kill body: assert clean,
// save/override/defer restore, assert an observable effect, assert the
// SPECIFIC code by exact string.
func TestCyclePeerMutationMatrix(t *testing.T) {
	t.Run("gray_vs_visited", func(t *testing.T) {
		program := diamondSyntheticProgram()

		clean := Validate(program)
		if !clean.Valid {
			t.Fatalf("expected the diamond to validate cleanly before the mutation, got %+v", clean)
		}

		previous := peerGrayVsVisitedMutationForTest
		peerGrayVsVisitedMutationForTest = true
		defer func() { peerGrayVsVisitedMutationForTest = previous }()

		mutated := Validate(program)
		if mutated.Valid {
			t.Fatalf("mutation (gray-versus-visited) had no observable effect: expected the diamond to be wrongly refused")
		}
		if !hasProblem(mutated.Problems, core.CallGraphCycle) {
			t.Fatalf("expected the wrongful refusal to carry %s, got %+v", core.CallGraphCycle, mutated.Problems)
		}
	})
}

// TestCyclePeerDeterministicUnderRepeat runs the same synthetic mutual
// cycle through Validate twice and requires an identical refusal verdict
// both times -- sorted adjacency and sorted roots are what make this
// deterministic rather than incidental; combined with `go test -count=2
// -shuffle=on`, this is Test 6's determinism requirement.
func TestCyclePeerDeterministicUnderRepeat(t *testing.T) {
	program := syntheticProgram(map[string][]string{
		"a": {"b"},
		"b": {"a"},
	})
	first := Validate(program)
	second := Validate(program)
	if first.Valid || second.Valid {
		t.Fatalf("expected both runs to refuse, got first.Valid=%v second.Valid=%v", first.Valid, second.Valid)
	}
	if len(first.Problems) != 1 || len(second.Problems) != 1 {
		t.Fatalf("expected exactly one problem per run, got first=%+v second=%+v", first.Problems, second.Problems)
	}
	if first.Problems[0].Code != second.Problems[0].Code {
		t.Fatalf("expected identical refusal code across repeated runs, got %q vs %q", first.Problems[0].Code, second.Problems[0].Code)
	}
}
