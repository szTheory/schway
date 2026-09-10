package check

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// ---------------------------------------------------------------------
// Phase 09 Plan 04 (D-09-50): the synthetic-shape zero-divergence
// differential, plus the cycle-peer witness-agreement differential that
// settles TRU-04's "recursion" shape.
//
// D-09-50, verified at planning time: package check exports exactly one
// admission entry point, check.Program(ast.Program), so no OUTSIDE package
// can compute a check-side verdict for a synthetically constructed
// core.Program -- and the synthetic corpus is deliberately built against
// core's own types and never parsed (D-08-34), so there is no ast.Program
// to feed check.Program in the first place. This differential therefore
// lives INSIDE package check, the only package that can reach both the
// unexported post-assembly liveness pass (checkInterproceduralLoanLiveness)
// and corevalidate.Validate (already an ordinary test dependency of this
// package, check_test.go:1-25). The `.lang` corpus gate
// (session_peer_gate_test.go's peerDivergenceExpected) stays exactly where
// it is. This is a split of VEHICLE, never of TRUTH: both harnesses share
// ONE definition of divergence (check admits with zero diagnostics, the
// peer independently refuses), applied to two structurally different
// program sources.
//
// Generator Reachability Register (this differential's own row, in
// 09-VALIDATION.md's own vocabulary):
//
//	Generator: testsupport.GenerateCallGraphCorpus (5 shapes: chain,
//	  diamond, dense, parser-shaped, forward), relocated by Plan 09-02
//	  (D-09-46).
//	Reaches: acyclic call graphs of increasing size and sharing, relay
//	  depth >= 2 (D-09-23).
//	Provably does NOT reach: any cyclic shape (refused before any liveness
//	  derivation ever runs -- see the cycle-peer differential below
//	  instead, and TestCyclePeerAgreesWithCheckOnIndirectCycleWitness's own
//	  doc comment for why a *liveness* differential over recursion cannot
//	  exist in this language); loops or back edges (none exist in the
//	  language, LANGUAGE-MATURITY.md); arity > 1 (every function declared
//	  by this generator takes exactly one parameter); and -- the gap this
//	  paragraph exists to make loud, not silent -- a SINGLE
//	  core.OpBorrowShared/core.OpBorrowExclusive operation anywhere in the
//	  corpus (verified by inspection: every builder template in
//	  testsupport/callgraphcorpus.go only ever emits OpCall, OpCopy, and
//	  OpReturn). With no loan ever created on either side, this
//	  differential's own zero-divergence sweep below is proven VACUOUSLY
//	  true over these five shapes: neither peer's loan-liveness law has
//	  anything to disagree about. Task 3's mutation-kill test exists
//	  precisely because of this gap -- it hand-builds ONE fixture
//	  containing a genuine borrow, so the differential's own detection
//	  machinery is proven to still catch a real divergence, rather than
//	  resting on an untested vacuous truth.
// ---------------------------------------------------------------------

// syntheticShapeDivergenceExpected is this plan's synthetic-corpus sibling
// of session_peer_gate_test.go's peerDivergenceExpected: the same
// hand-maintained, both-directions-exact register of every
// "<shape>/<n>" key where check admits (zero diagnostics) and
// corevalidate.Validate independently refuses. It shares ONE definition of
// divergence with peerDivergenceExpected -- copy its own doc comment's
// discipline, never a differently-worded one -- and this file exists as a
// separate VEHICLE, never a second TRUTH, solely because package check
// exports no entry point that can accept a bare core.Program (D-09-50). It
// is declared EMPTY here and must stay empty: TRU-04 criterion 1's claim is
// that these five call-graph shapes show ZERO divergence between the two
// peers, not some declared, tolerated set of them.
var syntheticShapeDivergenceExpected = map[string]string{}

// syntheticShapeSizeLadder is this differential's own small sweep ladder,
// mirroring costcorpus_test.go's own ladder discipline in KIND (more than
// one size per shape, to exercise both depth and sharing) but not in SCALE:
// this differential asserts an exact-set divergence gate, not a fitted
// growth exponent, so it needs enough sizes to see a shape's structure
// vary, never hundreds of functions per point.
var syntheticShapeSizeLadder = []int{4, 16, 64}

// completeSyntheticProgramForCorevalidate fills in the structural scaffolding
// testsupport's own generator never populates (EntryPointID/ReturnPointID,
// per-function Types/Places), because that generator was built solely to
// feed check's own buildInterproceduralSummaries work-counter
// (costcorpus_test.go), which never consults those fields. corevalidate's
// full Validate, unlike check's own direct interprocedural-liveness call,
// runs a complete structural pass FIRST and refuses on a missing
// EntryPointID/ReturnPointID/Place/Type before it ever reaches the
// loan-liveness replay this differential exists to compare -- so every
// generated shape needs this decoration before corevalidate.Validate can
// even reach the fact under test. This never touches the CALL-GRAPH TOPOLOGY
// the relocated generator produced (D-09-23's "same generator" claim is
// about the shape of the call graph, never about incidental structural
// bookkeeping check itself does not need); it only adds the bookkeeping
// corevalidate's structural pass requires to read the same operations.
func completeSyntheticProgramForCorevalidate(program core.Program) core.Program {
	for i := range program.Functions {
		function := &program.Functions[i]
		function.EntryPointID = function.ID + ":point:entry"
		function.ReturnPointID = function.ID + ":point:return"
		// testsupport's own generator never sets ReturnType (it is
		// irrelevant to check's own work-counter); corevalidate's
		// structural pass requires it to match the synthesized type
		// fact's own Constructor ("Byte", set below) at both the final
		// OpReturn (core.final_claim_mismatch) and every OpCall's own
		// return-type peer check (core.call_return_type_mismatch).
		function.ReturnType = "Byte"
		if function.Linear == nil {
			continue
		}
		typeID := function.ID + ":type:0"
		function.Linear.Types = []core.TypeFact{{
			ID: typeID, Shape: core.TypeRef{Constructor: "Byte", Arguments: []core.TypeRef{}},
			Abilities:         []core.Ability{core.AbilityCopy, core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape},
			NegativeWitnesses: []core.AbilityWitness{},
		}}

		// corevalidate's own structural pass (core.place_order) requires
		// every place ID to read exactly "<functionID>:place:<index>", in
		// strict first-reference order -- testsupport's own generator makes
		// no such promise (its own result places are named "...:place:rN"),
		// so every place this function's operations reference is remapped
		// here, in first-reference order, to that exact scheme. This
		// renames PLACE IDENTITY bookkeeping only; it never touches the
		// CALL-GRAPH TOPOLOGY the relocated generator produced (which
		// function calls which, by CalleeID, is completely untouched).
		originalParameterID := function.Parameter.ID
		remapped := make(map[string]string)
		var orderedOldIDs []string
		remap := func(oldID string) string {
			if oldID == "" {
				return ""
			}
			if newID, ok := remapped[oldID]; ok {
				return newID
			}
			newID := fmt.Sprintf("%s:place:%d", function.ID, len(orderedOldIDs))
			remapped[oldID] = newID
			orderedOldIDs = append(orderedOldIDs, oldID)
			return newID
		}

		function.Parameter.ID = remap(function.Parameter.ID)
		for j := range function.Linear.Operations {
			operation := &function.Linear.Operations[j]
			operation.SourceID = remap(operation.SourceID)
			operation.TargetID = remap(operation.TargetID)
			operation.TypeID = typeID
			// core.operation_order requires both ID and PointID to read
			// exactly "<functionID>:op:<index>" / "<functionID>:point:linear:<index>",
			// in strict declaration order -- testsupport's own operation
			// IDs (e.g. "...op:call0") do not follow that scheme, so both
			// are rewritten unconditionally here, by the operation's own
			// position in program order (never reordering the operations
			// themselves).
			operation.ID = fmt.Sprintf("%s:op:%d", function.ID, j)
			operation.PointID = fmt.Sprintf("%s:point:linear:%d", function.ID, j)
		}

		places := make([]core.Place, 0, len(orderedOldIDs))
		for _, oldID := range orderedOldIDs {
			name := "result"
			if oldID == originalParameterID {
				name = function.Parameter.Name
			}
			places = append(places, core.Place{ID: remapped[oldID], Name: name, TypeID: typeID})
		}
		function.Linear.Places = places
	}
	return program
}

// checkSyntheticProgramVerdict computes check's own post-assembly verdict
// for a synthetic core.Program built directly against core's own types --
// exactly the sequence check.Program itself runs post-parse (build the
// signature table, build interprocedural summaries, run the interprocedural
// liveness pass), but callable directly on a core.Program that never went
// through the parser (D-09-50: check.Program itself cannot be called here,
// since it only accepts ast.Program).
func checkSyntheticProgramVerdict(program core.Program) []diagnostic.Diagnostic {
	table, err := buildCallSignatureTable(program)
	if err != nil {
		// Every shape this differential feeds is proven acyclic by
		// construction (testsupport's own generator, and this file's own
		// hand-built mutation fixture); a table-build failure here would
		// mean callgraph.Order itself refused the program, which is
		// unreachable in practice for either input this file constructs.
		return nil
	}
	summaries, _ := buildInterproceduralSummaries(program, table)
	return checkInterproceduralLoanLiveness(program, summaries, map[string]diagnostic.Span{})
}

// findSyntheticShapeDivergences sweeps every shape
// testsupport.CallGraphCorpusShapes() names at syntheticShapeSizeLadder's
// sizes, returning every "<shape>/<n>" key where check admits and
// corevalidate.Validate independently refuses -- peerDivergenceExpected's
// own divergence definition, applied to the synthetic corpus instead of the
// `.lang` corpus. Factored out of the test function itself so Task 3's
// mutation-kill can assert on the returned map directly, rather than on the
// test binary's own pass/fail.
func findSyntheticShapeDivergences(t *testing.T) map[string]string {
	t.Helper()
	found := make(map[string]string)
	for _, shape := range testsupport.CallGraphCorpusShapes() {
		for _, n := range syntheticShapeSizeLadder {
			program, err := testsupport.GenerateCallGraphCorpus(shape, n)
			if err != nil {
				t.Fatalf("testsupport.GenerateCallGraphCorpus(%q, %d): %v", shape, n, err)
			}
			// testsupport's own generator never sets Schema/Module/ModuleID
			// (D-09-46: it builds bare core.Program values for check's own
			// work-counter sweep, which never consults those fields).
			// corevalidate's structural pass DOES require Schema ==
			// lang.core/1 before it will even reach the loan-liveness
			// replay this differential cares about -- set it here so an
			// unrelated core.schema refusal is never mistaken for the
			// interprocedural loan-liveness divergence this test tracks.
			program.Schema = core.Schema1
			program.Module = "syntheticshape"
			program.ModuleID = "s1:syntheticshape:module:syntheticshape"
			program = completeSyntheticProgramForCorevalidate(program)
			if diags := checkSyntheticProgramVerdict(program); len(diags) > 0 {
				// check itself refuses -- not this differential's tracked
				// direction, regardless of what the peer would say.
				continue
			}
			validated := corevalidate.Validate(program)
			if validated.Valid {
				continue
			}
			code := "core.unknown_peer_refusal"
			if len(validated.Problems) > 0 {
				code = validated.Problems[0].Code
			}
			found[fmt.Sprintf("%s/%d", shape, n)] = code
		}
	}
	return found
}

// assertSyntheticShapeDivergencesMatchExpected applies
// peerDivergenceExpected's own two-directional exactness discipline
// (session_peer_gate_test.go:TestNoUndeclaredCheckPeerDivergenceAcrossCorpus):
// every declared entry that no longer diverges fails, and every found
// divergence that is not declared also fails. Copied in wording style from
// that test's own failure messages so a reader sees immediately this is the
// SAME gate, applied to a different program source.
func assertSyntheticShapeDivergencesMatchExpected(t *testing.T, found, expected map[string]string) {
	t.Helper()
	for key, wantCode := range expected {
		gotCode, ok := found[key]
		if !ok {
			t.Errorf("declared synthetic divergence %s no longer diverges (check admits, peer no longer refuses) -- stale entry in syntheticShapeDivergenceExpected", key)
			continue
		}
		if gotCode != wantCode {
			t.Errorf("%s: peer code = %q, want declared %q", key, gotCode, wantCode)
		}
	}
	for key, gotCode := range found {
		if _, declared := expected[key]; !declared {
			t.Errorf("UNDECLARED synthetic divergence: %s (check admits, corevalidate refuses with %s) is not in syntheticShapeDivergenceExpected", key, gotCode)
		}
	}
}

// TestSyntheticShapeDifferentialHasNoUndeclaredDivergence is this plan's
// headline proof of TRU-04 criterion 1: over every shape
// testsupport.CallGraphCorpusShapes() names, at more than one size each,
// check and corevalidate.Validate show ZERO divergence on the
// interprocedural loan-liveness fact. See this file's own header comment
// for why this is provably vacuous over these five shapes alone (no borrow
// operation ever occurs in them) -- Task 3's mutation-kill test below
// proves the detection machinery itself is not vacuous.
func TestSyntheticShapeDifferentialHasNoUndeclaredDivergence(t *testing.T) {
	found := findSyntheticShapeDivergences(t)
	assertSyntheticShapeDivergencesMatchExpected(t, found, syntheticShapeDivergenceExpected)
}

// TestSyntheticShapeDifferentialUsesTheRelocatedGenerator is D-09-23's own
// "same generator, disjoint consumers" mechanical enforcement: this
// differential sweeps every shape testsupport.CallGraphCorpusShapes()
// itself returns (proven by actually generating each one below, rather
// than trusting a hardcoded echo of the shape names), and this file
// declares no local corpus builder of its own -- a static scan of this
// file's own source text for every corpusXxx template identifier.
func TestSyntheticShapeDifferentialUsesTheRelocatedGenerator(t *testing.T) {
	shapes := testsupport.CallGraphCorpusShapes()
	if len(shapes) == 0 {
		t.Fatal("testsupport.CallGraphCorpusShapes() returned no shapes")
	}
	for _, shape := range shapes {
		if _, err := testsupport.GenerateCallGraphCorpus(shape, 4); err != nil {
			t.Fatalf("testsupport.GenerateCallGraphCorpus(%q, 4) failed for a shape CallGraphCorpusShapes() itself returned: %v", shape, err)
		}
	}

	const selfPath = "check_peer_shape_differential_test.go"
	contents, err := os.ReadFile(selfPath)
	if err != nil {
		t.Fatalf("reading %s: %v", selfPath, err)
	}
	// Built from two concatenated literals per builder name, exactly like
	// costcorpus_relocation_test.go's own TestCallGraphCorpusGeneratorHasOneHome,
	// so this file's own doc-comment prose discussing these names (above)
	// is never itself a false-positive match for the scan.
	forbidden := []string{
		"corpus" + "Relay", "corpus" + "Chain", "corpus" + "Layered",
		"corpus" + "ParserShaped", "corpus" + "Forward",
		"corpus" + "LeafUse", "corpus" + "LeafPass",
	}
	for _, builder := range forbidden {
		if strings.Contains(string(contents), builder) {
			t.Fatalf("%s references %q -- a local corpus builder duplicate, forbidden by D-09-23's 'same generator, disjoint consumers'", selfPath, builder)
		}
	}
}

// ---------------------------------------------------------------------
// Task 2: the cycle-peer differential -- TRU-04's "recursion" shape,
// settled with witness agreement.
//
// `callgraph`'s three-color DFS refuses core.call_graph_cycle BEFORE any
// liveness derivation ever runs, so there is no admitted recursive program
// in this compiler and a LIVENESS differential over recursion is a category
// error (D-09-22). TRU-04's "recursion" shape is instead satisfied here as
// a CYCLE-PEER differential: both check's own callgraph refusal and
// corevalidate's independent refusal must agree on refusal-or-not AND on
// the witness, for self, mutual, and indirect (3+ hop) cycles.
// ---------------------------------------------------------------------

// cyclePeerProgramFunctionID names one node in this file's own hand-built
// cycle-shape core.Program, mirroring
// corevalidate_cycle_peer_test.go's own syntheticFunctionID convention.
// Kept as a SEPARATE helper (never importing that one) because package
// check cannot reach an unexported symbol across the package boundary
// (D-09-50's vehicle split) -- not because the naming scheme itself needs
// to differ.
func cyclePeerProgramFunctionID(name string) string { return "s1:cyclepeeragree:fn:" + name }

// cyclePeerProgramFunction builds one fully structurally-valid core.Function
// named `name`, whose body calls each of `callees` (by node name) in
// sequence, then returns -- the same construction shape
// corevalidate_cycle_peer_test.go's own syntheticFunction uses, ported into
// package check because that helper is unexported and unreachable from
// here.
func cyclePeerProgramFunction(name string, callees []string) core.Function {
	fnID := cyclePeerProgramFunctionID(name)
	typeID := fnID + ":type:0"
	parameterID := fnID + ":place:0"
	typeFact := core.TypeFact{
		ID: typeID, Shape: core.TypeRef{Constructor: "Byte", Arguments: []core.TypeRef{}},
		Abilities:         []core.Ability{core.AbilityCopy, core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape},
		NegativeWitnesses: []core.AbilityWitness{},
	}
	places := []core.Place{{ID: parameterID, Name: "value", TypeID: typeID}}

	var operations []core.LinearOperation
	current := parameterID
	for index, callee := range callees {
		target := fmt.Sprintf("%s:place:%d", fnID, index+1)
		places = append(places, core.Place{ID: target, Name: fmt.Sprintf("result%d", index), TypeID: typeID})
		operations = append(operations, core.LinearOperation{
			ID: fmt.Sprintf("%s:op:%d", fnID, index), PointID: fmt.Sprintf("%s:point:linear:%d", fnID, index),
			Kind: core.OpCall, SourceID: current, TargetID: target, TypeID: typeID,
			CalleeID: cyclePeerProgramFunctionID(callee),
		})
		current = target
	}
	returnIndex := len(callees)
	operations = append(operations, core.LinearOperation{
		ID: fmt.Sprintf("%s:op:%d", fnID, returnIndex), PointID: fmt.Sprintf("%s:point:linear:%d", fnID, returnIndex),
		Kind: core.OpReturn, SourceID: current, TypeID: typeID,
	})

	return core.Function{
		ID: fnID, Name: name,
		EntryPointID: fnID + ":point:entry", ReturnPointID: fnID + ":point:return",
		Parameter: core.Parameter{ID: parameterID, Name: "value", Type: "Byte"}, ReturnType: "Byte",
		Linear: &core.LinearBody{
			ID: fnID + ":linear", Types: []core.TypeFact{typeFact}, Places: places, Operations: operations,
		},
	}
}

// cyclePeerProgram builds a fully structurally-valid core.Program directly
// from an adjacency description, keyed by short node names -- the same
// shape corevalidate_cycle_peer_test.go's own syntheticProgram builds,
// reimplemented here (never imported: that helper is unexported) so
// package check can construct the SAME three cycle shapes as core.Program
// values and run both check's and corevalidate's own independent refusals
// against them.
func cyclePeerProgram(edges map[string][]string) core.Program {
	declared := make(map[string]bool)
	var names []string
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
		callees := append([]string(nil), edges[name]...)
		sort.Strings(callees)
		functions = append(functions, cyclePeerProgramFunction(name, callees))
	}
	return core.Program{
		Schema: core.Schema1, Module: "cyclepeeragree", ModuleID: "s1:cyclepeeragree:module:cyclepeeragree",
		Functions: functions,
	}
}

// cycleWitnessMembers extracts check's own reported cycle_member cause
// Details from a core.call_graph_cycle diagnostic (checkCallGraphAcyclic's
// own shape), as a set -- never compared as an ordered list, since the two
// peers' independent discovery orders must never manufacture a false
// disagreement.
func cycleWitnessMembers(diag *diagnostic.Diagnostic) map[string]bool {
	members := make(map[string]bool, len(diag.Causes))
	for _, cause := range diag.Causes {
		if cause.Kind == "cycle_member" {
			members[cause.Detail] = true
		}
	}
	return members
}

// assertCyclePeerWitnessAgreement is the shared body for both witness-
// agreement tests below: run check's own callgraph refusal and
// corevalidate's independent refusal over the SAME program, assert both
// refuse, and assert the peer's own reported witness is a genuine MEMBER of
// check's reported cycle_member set. This is a set-membership comparison,
// never an exact-set-equality one: corevalidate's Problem carries only the
// ONE node whose back edge closed the cycle (corevalidate.go's own DFS,
// `return v.check(false, core.CallGraphCycle, child)`), while check's
// causes enumerate up to callgraph.MaxCycleCauses full members. Demanding
// exact multiset equality between an intentionally single-witness Detail
// and a multi-member cause list would be comparing two things that were
// never meant to carry the same shape of evidence; demanding SET
// membership is the honest comparison this asymmetry actually supports,
// and it is still a "compare as a set" discipline, not a comparison of the
// two peers' own differently-shaped rendered cause lists.
func assertCyclePeerWitnessAgreement(t *testing.T, name string, program core.Program) {
	t.Helper()

	checkDiag := checkCallGraphAcyclic(program, map[string]diagnostic.Span{})
	if checkDiag == nil {
		t.Fatalf("expected check's own callgraph refusal for the %s cycle, got none", name)
	}
	if checkDiag.Code != core.CallGraphCycle {
		t.Fatalf("expected %s, got %s", core.CallGraphCycle, checkDiag.Code)
	}
	checkMembers := cycleWitnessMembers(checkDiag)
	if len(checkMembers) == 0 {
		t.Fatalf("expected check's own diagnostic to name at least one cycle_member cause for the %s cycle, got %+v", name, checkDiag.Causes)
	}

	peerResult := corevalidate.Validate(program)
	if peerResult.Valid {
		t.Fatalf("expected corevalidate to independently refuse the %s cycle, got valid", name)
	}
	if len(peerResult.Problems) == 0 || peerResult.Problems[0].Code != core.CallGraphCycle {
		t.Fatalf("expected %s, got %+v", core.CallGraphCycle, peerResult.Problems)
	}
	peerWitness := peerResult.Problems[0].Detail
	if !checkMembers[peerWitness] {
		t.Fatalf("witness disagreement for the %s cycle: corevalidate named %q, which is not among check's own reported cycle members %+v", name, peerWitness, checkMembers)
	}
}

// TestCyclePeerAgreesWithCheckOnSelfAndMutualCycleWitness is TRU-04's
// "recursion" shape, settled as a cycle-peer differential (D-09-22): both
// check's callgraph refusal and corevalidate's independent refusal agree on
// refusal-or-not AND on the witness for a self-cycle and a two-node mutual
// cycle. A LIVENESS differential over recursion cannot exist in this
// language, because the compiler refuses every cycle before any liveness
// derivation ever runs -- there is no admitted recursive program for such a
// differential to observe. No artifact in this phase claims otherwise; this
// agreement, over the witness the two independent refusals discover, is the
// honest and complete form TRU-04's "recursion" shape takes here.
func TestCyclePeerAgreesWithCheckOnSelfAndMutualCycleWitness(t *testing.T) {
	cases := []struct {
		name  string
		edges map[string][]string
	}{
		{"self", map[string][]string{"loop": {"loop"}}},
		{"mutual", map[string][]string{"a": {"b"}, "b": {"a"}}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			assertCyclePeerWitnessAgreement(t, testCase.name, cyclePeerProgram(testCase.edges))
		})
	}
}

// TestCyclePeerAgreesWithCheckOnIndirectCycleWitness extends the witness
// agreement to indirect (3+ hop) cycles (a 3-hop and a 4-hop ring). Same
// disposition as TestCyclePeerAgreesWithCheckOnSelfAndMutualCycleWitness
// above: this is TRU-04's "recursion" shape, and a liveness differential
// over an admitted recursive program cannot exist in this language, because
// `callgraph`'s three-color DFS (and this peer's own independent one)
// refuse every cycle -- self, mutual, or indirect -- before any liveness
// derivation runs. Mirrors
// corevalidate_cycle_peer_test.go's own
// TestCyclePeerRefusesIndirectCycle/TestCyclePeerRefusesFourHopIndirectCycle,
// which prove the peer-only half of this same claim.
func TestCyclePeerAgreesWithCheckOnIndirectCycleWitness(t *testing.T) {
	cases := []struct {
		name  string
		edges map[string][]string
	}{
		{"3-hop", map[string][]string{"a": {"b"}, "b": {"c"}, "c": {"a"}}},
		{"4-hop", map[string][]string{"a": {"b"}, "b": {"c"}, "c": {"d"}, "d": {"a"}}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			assertCyclePeerWitnessAgreement(t, testCase.name, cyclePeerProgram(testCase.edges))
		})
	}
}

// TestCyclePeerAndCheckBothAcceptAcyclicDiamond is the accepting-path
// sibling required by this task's own acceptance criteria: an acyclic
// shared-leaf diamond must be refused by NEITHER layer, so "agreement" here
// is proven to mean agreement on ADMIT too, not merely agreement on
// refusal. corevalidate_cycle_peer_test.go's own
// TestCyclePeerAcceptsAcyclicDiamond already covers the peer side; this is
// the check-side half.
func TestCyclePeerAndCheckBothAcceptAcyclicDiamond(t *testing.T) {
	program := cyclePeerProgram(map[string][]string{
		"top":    {"left", "right"},
		"left":   {"shared"},
		"right":  {"shared"},
		"shared": {"leaf"},
		"leaf":   nil,
	})
	if diag := checkCallGraphAcyclic(program, map[string]diagnostic.Span{}); diag != nil {
		t.Fatalf("expected check to admit an acyclic diamond, got %+v", diag)
	}
	if result := corevalidate.Validate(program); !result.Valid {
		t.Fatalf("expected corevalidate to independently admit an acyclic diamond, got %+v", result.Problems)
	}
}

// ---------------------------------------------------------------------
// Task 3: mutation-kill for the synthetic-shape differential.
// ---------------------------------------------------------------------

// mutationCandidateProgram hand-builds a minimal two-function synthetic
// core.Program mirroring testdata/phase08/twin_a_accept.lang's own proven
// shape (escort borrows its own parameter, calls escortee with the
// borrowed place, then moves its own parameter after the call returns):
// escortee's body genuinely just moves-and-returns its own parameter (no
// borrow at all), so BOTH check's signature-table-derived
// returnsBorrowOfParam bit and corevalidate's own body-derived peer
// loan-carry fact agree it does not return a borrow of its parameter -- the
// identical baseline agreement Plan 09-01 already proved for the real
// fixture (see corevalidate_peer_liveness_test.go's own
// TestPeerLoanCarryDerivesForwardFromOperations "owned" case).
//
// This fixture is kept entirely separate from
// testsupport.CallGraphCorpusShapes() (never one of its five shapes,
// deliberately never reusing any corpusXxx builder): as this file's own
// header comment records, none of those five shapes ever contain a single
// OpBorrowShared/OpBorrowExclusive operation, so no loan ever exists there
// for a loan-carry seam to corrupt. This fixture exists solely to give
// Task 3's mutation-kill a genuine loan to break.
func mutationCandidateProgram() core.Program {
	calleeID := "s1:mutation:fn:escortee"
	callerID := "s1:mutation:fn:escort"

	bufferType := func(fnID string) core.TypeFact {
		return core.TypeFact{
			ID: fnID + ":type:0", Shape: core.TypeRef{Constructor: "Buffer", Arguments: []core.TypeRef{}},
			Abilities:         []core.Ability{core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape},
			NegativeWitnesses: []core.AbilityWitness{{Ability: core.AbilityCopy, Path: []string{"Buffer"}}},
		}
	}

	calleeType := bufferType(calleeID)
	calleeParam := calleeID + ":place:0"
	calleeTaken := calleeID + ":place:1"
	callee := core.Function{
		ID: calleeID, Name: "escortee",
		EntryPointID: calleeID + ":point:entry", ReturnPointID: calleeID + ":point:return",
		Parameter: core.Parameter{ID: calleeParam, Name: "buffer", Type: "Buffer"}, ReturnType: "Buffer",
		Linear: &core.LinearBody{
			ID:     calleeID + ":linear",
			Types:  []core.TypeFact{calleeType},
			Places: []core.Place{{ID: calleeParam, Name: "buffer", TypeID: calleeType.ID}, {ID: calleeTaken, Name: "taken", TypeID: calleeType.ID}},
			Operations: []core.LinearOperation{
				{ID: calleeID + ":op:0", PointID: calleeID + ":point:linear:0", Kind: core.OpMove, SourceID: calleeParam, TargetID: calleeTaken, TypeID: calleeType.ID},
				{ID: calleeID + ":op:1", PointID: calleeID + ":point:linear:1", Kind: core.OpReturn, SourceID: calleeTaken, TypeID: calleeType.ID},
			},
		},
	}

	callerType := bufferType(callerID)
	callerParam := callerID + ":place:0"
	borrowed := callerID + ":place:1"
	aliased := callerID + ":place:2"
	delivered := callerID + ":place:3"
	caller := core.Function{
		ID: callerID, Name: "escort",
		EntryPointID: callerID + ":point:entry", ReturnPointID: callerID + ":point:return",
		Parameter: core.Parameter{ID: callerParam, Name: "buffer", Type: "Buffer"}, ReturnType: "Buffer",
		Linear: &core.LinearBody{
			ID:    callerID + ":linear",
			Types: []core.TypeFact{callerType},
			Places: []core.Place{
				{ID: callerParam, Name: "buffer", TypeID: callerType.ID},
				{ID: borrowed, Name: "borrowed", TypeID: callerType.ID},
				{ID: aliased, Name: "aliased", TypeID: callerType.ID},
				{ID: delivered, Name: "delivered", TypeID: callerType.ID},
			},
			Operations: []core.LinearOperation{
				{ID: callerID + ":op:0", PointID: callerID + ":point:linear:0", Kind: core.OpBorrowShared, SourceID: callerParam, TargetID: borrowed, LoanID: callerID + ":loan:0", TypeID: callerType.ID},
				{ID: callerID + ":op:1", PointID: callerID + ":point:linear:1", Kind: core.OpCall, SourceID: borrowed, TargetID: aliased, TypeID: callerType.ID, CalleeID: calleeID},
				{ID: callerID + ":op:2", PointID: callerID + ":point:linear:2", Kind: core.OpMove, SourceID: callerParam, TargetID: delivered, TypeID: callerType.ID},
				{ID: callerID + ":op:3", PointID: callerID + ":point:linear:3", Kind: core.OpReturn, SourceID: aliased, TypeID: callerType.ID},
			},
		},
	}

	return core.Program{
		Schema: core.Schema1, Module: "mutation", ModuleID: "s1:mutation:module:mutation",
		Functions: []core.Function{callee, caller},
	}
}

// TestSyntheticShapeDifferentialMutationReintroducesDivergence is this
// plan's own mutation-kill (RETROSPECTIVE.md Key Lesson 1: "A control you
// have never seen fail is a claim"): it proves the differential can still
// detect a real divergence, not merely that it has never found one over a
// corpus that (per this file's own header comment) structurally cannot
// contain a loan at all.
//
// Mutation surface used: corevalidate's own cross-package fault-injection
// seam, corevalidate.SetForcePeerLoanCarryTrueForTest
// (corevalidate_peer_liveness.go, D-09-25/D-09-26), which forces every
// OpCall's loan-carry consult to report ReturnsBorrowOfParam == true
// regardless of the callee's actual body -- reverting corevalidate's own
// buildLoanChainIndex OpCall branch to its pre-Plan-09-01 (D-09-03)
// unconditional-propagation behavior. check's own admission path never
// consults this seam (it is declared entirely inside package corevalidate),
// so engaging it changes ONLY corevalidate's verdict.
//
// Two alternatives were considered and rejected:
//   - check.loanLivenessBoundSeam: trips check's OWN bound-exceeded
//     refusal, flipping check from ADMIT to REFUSE -- the WRONG direction
//     for this differential, which only tracks check-admits/peer-refuses
//     (mirroring peerDivergenceExpected's own asymmetry). Engaging it would
//     make check refuse, which this differential's own found-map does not
//     even inspect (findSyntheticShapeDivergences skips a case entirely
//     once check itself refuses).
//   - check.disableInterproceduralLoanLivenessForTest: only gates
//     check.Program's own call site to checkInterproceduralLoanLiveness.
//     This differential never calls check.Program at all (D-09-50 -- it
//     drives checkSyntheticProgramVerdict, which calls
//     checkInterproceduralLoanLiveness directly), so engaging this seam
//     would have had zero observable effect here.
func TestSyntheticShapeDifferentialMutationReintroducesDivergence(t *testing.T) {
	program := mutationCandidateProgram()

	// Baseline: both peers must independently agree the program is clean,
	// mirroring twin_a_accept.lang's own already-proven post-Plan-09-01
	// agreement (see mutationCandidateProgram's own doc comment).
	if diags := checkSyntheticProgramVerdict(program); len(diags) != 0 {
		t.Fatalf("expected check to admit the baseline mutation-candidate program, got %+v", diags)
	}
	baseline := corevalidate.Validate(program)
	if !baseline.Valid {
		t.Fatalf("expected corevalidate to independently admit the baseline mutation-candidate program before any mutation, got %+v", baseline.Problems)
	}

	const wantKey = "mutation/escort-escortee"

	// Mutation engaged, restored by an immediately deferred call inside
	// this closure -- the restore runs before this function does anything
	// further, so the "disengaged" check below genuinely observes the
	// seam OFF.
	mutatedFound := func() map[string]string {
		restore := corevalidate.SetForcePeerLoanCarryTrueForTest(true)
		defer restore()

		found := make(map[string]string)
		if diags := checkSyntheticProgramVerdict(program); len(diags) == 0 {
			if validated := corevalidate.Validate(program); !validated.Valid {
				code := "core.unknown_peer_refusal"
				if len(validated.Problems) > 0 {
					code = validated.Problems[0].Code
				}
				found[wantKey] = code
			}
		}
		return found
	}()

	gotCode, ok := mutatedFound[wantKey]
	if !ok {
		t.Fatalf("expected the mutation to reintroduce a divergence keyed %q, found none: %+v", wantKey, mutatedFound)
	}
	if gotCode != "core.move_while_borrowed" {
		t.Fatalf("expected the reintroduced divergence to carry core.move_while_borrowed, got %q", gotCode)
	}

	// Disengaged: the differential must report clean again on the SAME
	// program, proving the divergence was caused by the mutation, not by
	// the fixture itself.
	cleanAgain := corevalidate.Validate(program)
	if !cleanAgain.Valid {
		t.Fatalf("expected corevalidate to admit the mutation-candidate program again once the seam is restored, got %+v", cleanAgain.Problems)
	}
}
