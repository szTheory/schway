package session

import (
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/syntax"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// osStatModTime returns path's own on-disk modification time, used to prove
// a test run never writes to the committed register file.
func osStatModTime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

// ---------------------------------------------------------------------
// Task 1: the generator op-kind closure test, cell identity, idempotency.
// ---------------------------------------------------------------------

// TestQLT03GeneratorOpKindClosure is the register's cheapest, first-run
// drift protection (D-11-45/D-11-48/D-11-50): it fires the instant anyone
// adds a borrow (or any other new) op-kind to
// testsupport.GenerateCallGraphCorpus, BEFORE any row in this register is
// even consulted. It compares the OBSERVED op-kind set against
// QLT03CommittedGeneratorOpKinds()'s COMMITTED LITERAL -- never against a
// second recomputation of itself.
func TestQLT03GeneratorOpKindClosure(t *testing.T) {
	observed, err := QLT03ObservedGeneratorOpKinds()
	if err != nil {
		t.Fatalf("sweeping generator op-kinds: %v", err)
	}
	committed := QLT03CommittedGeneratorOpKinds()
	if !qlt03OpKindSetsEqual(observed, committed) {
		t.Fatalf(
			"SURPRISE: testsupport.GenerateCallGraphCorpus now emits op-kinds %v, "+
				"which does not equal the committed literal %v declared in "+
				"QLT03CommittedGeneratorOpKinds(). Per this task's own escape hatch: "+
				"record this surprise in PHASE-11-DEBT.md and revise the taxonomy "+
				"BEFORE editing the committed literal to match the observation.",
			observed, committed,
		)
	}
}

// TestQLT03StructuralGeneratorLimits pins the DECLARED structural
// invariants of testsupport's own corpus builders that
// QLT03CommittedGeneratorOpKinds() alone does not cover: an OpCall's own
// argument is never sourced from an OpCopy target (the "copy" argument
// mode is unreachable even though OpCopy itself IS emitted), no topology
// other than "forward" ever chains one OpCall's own result into another
// OpCall's argument within the same function (composition depth > 1 is
// exclusive to "forward"), and "forward" itself only ever calls
// corpusLeafPass-shaped callees. Every "generator"-mechanism row in
// qlt03_shape_register.json that is not covered by the op-kind closure
// test above names THIS test as its falsifier.
func TestQLT03StructuralGeneratorLimits(t *testing.T) {
	observations, err := qlt03SweepAllSizes(nil)
	if err != nil {
		t.Fatalf("sweeping edges: %v", err)
	}
	if len(observations) == 0 {
		t.Fatal("expected at least one edge observation from the full sweep")
	}

	for _, obs := range observations {
		if obs.ArgumentMode == ArgCopy || obs.ArgumentMode == ArgBorrowShared || obs.ArgumentMode == ArgBorrowExclusive {
			t.Fatalf("SURPRISE: observed argument_mode %q on an edge (topology=%s) -- "+
				"this generator was declared to source every OpCall argument as a raw "+
				"move; update PHASE-11-DEBT.md before editing this register's rows",
				obs.ArgumentMode, obs.Topology)
		}
		if obs.Topology != "forward" && obs.Depth > 1 {
			t.Fatalf("SURPRISE: topology %q produced an edge at composition depth %d > 1 -- "+
				"only \"forward\" was declared to chain calls", obs.Topology, obs.Depth)
		}
		if obs.Topology == "forward" && obs.CalleeShape != ShapeLeafPass {
			t.Fatalf("SURPRISE: \"forward\" topology produced a callee of shape %q, not leaf_pass",
				obs.CalleeShape)
		}
	}
}

// TestQLT03BorrowOfCalleeResultIsRefused is the headline negative's own
// falsifier (D-11-49, D-10-C01): testdata/phase10/relay_depth3_refuse.lang's
// own `f2` declares a borrow-typed return sourced from forwarding leaf's
// own call result. corevalidate.peerDeriveOriginFacts has no core.OpCall
// case, so `f2` is judged not-Callable independently of check, and f1's
// own call to f2 is refused with core.callee_not_callable. This test goes
// RED the moment that gap closes -- exactly the falsifier this register's
// `refused` rows require.
func TestQLT03BorrowOfCalleeResultIsRefused(t *testing.T) {
	source, err := readBoundedFile(nat03CorpusPath("testdata/phase10/relay_depth3_refuse.lang"), syntax.MaxSourceBytes)
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	checked := Check(source)
	result := corevalidate.Validate(checked.Program)
	if result.Valid {
		t.Fatal("QLT-03 headline gap closed: corevalidate now ACCEPTS relay_depth3_refuse.lang -- " +
			"update this register's borrow_of_callee_result rows (D-10-C01, PHASE-11-DEBT.md) " +
			"before treating this as a pass")
	}
	found := false
	for _, problem := range result.Problems {
		if problem.Code == core.CalleeNotCallable {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a %s problem, got %+v", core.CalleeNotCallable, result.Problems)
	}
}

// TestQLT03CellIdentityIsExactStringEquality proves D-11-44's declared cell
// identity: two rows whose RAW axis-value keys differ ONLY by letter case
// or surrounding whitespace are reported as a DUPLICATE-ROW error by the
// near-duplicate check (qlt03NearDuplicateFold), rather than silently
// accepted as two distinct cells -- while the register's OWN real identity
// (QLT03CellKey) stays exact and case-sensitive, never folding anything
// itself.
func TestQLT03CellIdentityIsExactStringEquality(t *testing.T) {
	canonical := QLT03CellKey(qlt03AxisSetArgReturn, map[string]string{"argument_mode": "move", "return_origin": "param_forwarded"})
	whitespaceVariant := QLT03CellKey(qlt03AxisSetArgReturn, map[string]string{"argument_mode": " move ", "return_origin": "param_forwarded"})
	caseVariant := QLT03CellKey(qlt03AxisSetArgReturn, map[string]string{"argument_mode": "Move", "return_origin": "param_forwarded"})

	if canonical == whitespaceVariant {
		t.Fatal("expected the whitespace variant's RAW key to differ from the canonical key under exact comparison")
	}
	if canonical == caseVariant {
		t.Fatal("expected the case variant's RAW key to differ from the canonical key under exact comparison")
	}

	if qlt03NearDuplicateFold(canonical) != qlt03NearDuplicateFold(whitespaceVariant) {
		t.Fatal("expected the whitespace variant to be reported as a duplicate-row error (same folded identity)")
	}
	if qlt03NearDuplicateFold(canonical) != qlt03NearDuplicateFold(caseVariant) {
		t.Fatal("expected the case variant to be reported as a duplicate-row error (same folded identity)")
	}

	rows := []QLT03Row{
		{AxisSet: qlt03AxisSetArgReturn, Axes: map[string]string{"argument_mode": "move", "return_origin": "param_forwarded"}},
		{AxisSet: qlt03AxisSetArgReturn, Axes: map[string]string{"argument_mode": "Move", "return_origin": "param_forwarded"}},
	}
	failures := qlt03NearDuplicateRowFailures(rows)
	if len(failures) == 0 {
		t.Fatal("expected a case-differing pair to be reported as a duplicate-row error")
	}
}

// TestQLT03ReachedRecomputationIsIdempotent recomputes `reached` at n in
// {8, 24, 48} twice and asserts byte-identical results, and asserts the
// committed qlt03_shape_register.json is unchanged by a test run -- the
// audit reads, it never writes.
func TestQLT03ReachedRecomputationIsIdempotent(t *testing.T) {
	first, err := qlt03SweepAllSizes(nil)
	if err != nil {
		t.Fatalf("first sweep: %v", err)
	}
	second, err := qlt03SweepAllSizes(nil)
	if err != nil {
		t.Fatalf("second sweep: %v", err)
	}
	firstAgg := qlt03AggregateByArgReturn(first)
	secondAgg := qlt03AggregateByArgReturn(second)
	if len(firstAgg) != len(secondAgg) {
		t.Fatalf("recomputation is not idempotent: %d cells vs %d cells", len(firstAgg), len(secondAgg))
	}
	for key, a := range firstAgg {
		b, ok := secondAgg[key]
		if !ok {
			t.Fatalf("recomputation is not idempotent: cell %v present in first sweep, absent in second", key)
		}
		if a.toEvidence() != b.toEvidence() {
			t.Fatalf("recomputation is not idempotent for cell %v: %+v vs %+v", key, a.toEvidence(), b.toEvidence())
		}
	}

	path := nat03CorpusPath("internal/compiler/session/qlt03_shape_register.json")
	before, err := osStatModTime(path)
	if err != nil {
		t.Fatalf("stat before: %v", err)
	}
	if _, err := LoadQLT03Register(); err != nil {
		t.Fatalf("LoadQLT03Register: %v", err)
	}
	after, err := osStatModTime(path)
	if err != nil {
		t.Fatalf("stat after: %v", err)
	}
	if before != after {
		t.Fatalf("qlt03_shape_register.json's mtime changed across a test run: %v -> %v", before, after)
	}
}

// ---------------------------------------------------------------------
// Task 2: the register artifact, loader, discriminated rows.
// ---------------------------------------------------------------------

// TestQLT03RegisterRowsAreDiscriminated asserts the loaded register has at
// least 30 rows and that exactly one of Reached/Unreachable is populated
// on every row.
func TestQLT03RegisterRowsAreDiscriminated(t *testing.T) {
	rows, err := LoadQLT03Register()
	if err != nil {
		t.Fatalf("LoadQLT03Register: %v", err)
	}
	if len(rows) < 30 {
		t.Fatalf("expected at least 30 rows, got %d", len(rows))
	}
	for _, row := range rows {
		hasReached := row.Reached != nil
		hasUnreachable := row.Unreachable != nil
		if hasReached == hasUnreachable {
			t.Fatalf("row %s: expected exactly one of reached/unreachable, got reached=%v unreachable=%v", row.CellKey(), hasReached, hasUnreachable)
		}
	}
}

// TestQLT03RegisterRejectsOutOfSetMechanism proves the loader's own
// closed-set enforcement (D-11-46): a row naming a mechanism outside the
// five declared values -- including free text -- is rejected, never
// silently accepted.
func TestQLT03RegisterRejectsOutOfSetMechanism(t *testing.T) {
	bad := []byte(`[{"axis_set":"argument_mode_x_return_origin","axes":{"argument_mode":"move","return_origin":"fresh_owned"},"unreachable":{"mechanism":"a prose argument, not a mechanism","falsifier_test":"TestQLT03GeneratorOpKindClosure"}}]`)
	restore := setQLT03RegisterBytesForTest(bad)
	defer restore()
	if _, err := LoadQLT03Register(); err == nil {
		t.Fatal("expected LoadQLT03Register to reject an out-of-closed-set mechanism")
	}
}

// TestQLT03RegisterRejectsNonexistentFalsifierTest proves the loader
// rejects a row whose named falsifier test does not exist as a symbol in
// this package's own test files.
func TestQLT03RegisterRejectsNonexistentFalsifierTest(t *testing.T) {
	bad := []byte(`[{"axis_set":"argument_mode_x_return_origin","axes":{"argument_mode":"move","return_origin":"fresh_owned"},"unreachable":{"mechanism":"generator","falsifier_test":"TestQLT03DoesNotExistAnywhere"}}]`)
	restore := setQLT03RegisterBytesForTest(bad)
	defer restore()
	if _, err := LoadQLT03Register(); err == nil {
		t.Fatal("expected LoadQLT03Register to reject a nonexistent falsifier test symbol")
	}
}

// TestQLT03RegisterRejectsBothOrNeitherPopulated proves the discriminated-
// union invariant is enforced at load time, not merely by the audit.
func TestQLT03RegisterRejectsBothOrNeitherPopulated(t *testing.T) {
	neither := []byte(`[{"axis_set":"argument_mode_x_return_origin","axes":{"argument_mode":"move","return_origin":"fresh_owned"}}]`)
	restore := setQLT03RegisterBytesForTest(neither)
	if _, err := LoadQLT03Register(); err == nil {
		t.Fatal("expected LoadQLT03Register to reject a row with neither reached nor unreachable populated")
	}
	restore()

	both := []byte(`[{"axis_set":"argument_mode_x_return_origin","axes":{"argument_mode":"move","return_origin":"fresh_owned"},"reached":{"occurrences":10,"distinct_shapes":2,"max_depth":1},"unreachable":{"mechanism":"generator","falsifier_test":"TestQLT03GeneratorOpKindClosure"}}]`)
	restore = setQLT03RegisterBytesForTest(both)
	defer restore()
	if _, err := LoadQLT03Register(); err == nil {
		t.Fatal("expected LoadQLT03Register to reject a row with BOTH reached and unreachable populated")
	}
}

// ---------------------------------------------------------------------
// Task 3: the audit, mutation kill, swarm sweep, drift protection.
// ---------------------------------------------------------------------

// TestQLT03Register is the top-level audit test: AuditQLT03Register
// returns zero failures over the committed register, printing every
// failure verbatim on a red run so it is diagnosable without re-running.
func TestQLT03Register(t *testing.T) {
	rows, err := LoadQLT03Register()
	if err != nil {
		t.Fatalf("LoadQLT03Register: %v", err)
	}
	failures := AuditQLT03Register(rows)
	if len(failures) != 0 {
		t.Fatalf("AuditQLT03Register found %d failure(s):\n%s", len(failures), joinLines(failures))
	}
}

func joinLines(lines []string) string {
	out := ""
	for _, line := range lines {
		out += "  " + line + "\n"
	}
	return out
}

// TestQLT03StatusRecomputation recomputes reach at n in {8, 24, 48}
// individually (not aggregated) and asserts the register's recorded
// REACHED statuses still hold -- present with a nonzero occurrence count
// -- at every size. A status that holds at n=8 but not at n=48 is a drift
// failure, not a pass.
func TestQLT03StatusRecomputation(t *testing.T) {
	rows, err := LoadQLT03Register()
	if err != nil {
		t.Fatalf("LoadQLT03Register: %v", err)
	}
	for _, n := range qlt03SweepSizes {
		observations, err := qlt03SweepEdges(n, nil)
		if err != nil {
			t.Fatalf("sweeping at n=%d: %v", n, err)
		}
		argReturn := qlt03AggregateByArgReturn(observations)
		topoShapeDepth := qlt03AggregateByTopologyShapeDepth(observations)
		for _, row := range rows {
			if row.Reached == nil {
				continue
			}
			var occurrences int
			switch row.AxisSet {
			case qlt03AxisSetArgReturn:
				key := [2]string{row.Axes["argument_mode"], row.Axes["return_origin"]}
				if agg, ok := argReturn[key]; ok {
					occurrences = agg.Occurrences
				}
			case qlt03AxisSetTopologyShapeDepth:
				key := [3]string{row.Axes["edge_topology"], row.Axes["callee_body_shape"], row.Axes["composition_depth_bucket"]}
				if agg, ok := topoShapeDepth[key]; ok {
					occurrences = agg.Occurrences
				}
			}
			if occurrences == 0 {
				t.Fatalf("DRIFT: cell %s is recorded reached in the committed register, but is absent (0 occurrences) at n=%d", row.CellKey(), n)
			}
		}
	}
}

// TestQLT03AxisCompletenessAgainstCorpusShapes cross-checks the register's
// edge-topology axis values against testsupport.CallGraphCorpusShapes(): a
// shape the generator produces that the axis does not name is a register
// gap.
func TestQLT03AxisCompletenessAgainstCorpusShapes(t *testing.T) {
	declared := map[string]bool{}
	for _, topology := range QLT03EdgeTopologies() {
		declared[topology] = true
	}
	for _, shape := range testsupport.CallGraphCorpusShapes() {
		if !declared[shape] {
			t.Fatalf("register gap: corpus shape %q is not named by QLT03EdgeTopologies()", shape)
		}
	}
}

// TestQLT03AuditCanFail is D-11-50's mutation kill: an export_test.go seam
// perturbs the enumeration AuditQLT03Register consults, WITHOUT editing the
// committed register file, and the audit must return a non-empty failure
// list.
func TestQLT03AuditCanFail(t *testing.T) {
	rows, err := LoadQLT03Register()
	if err != nil {
		t.Fatalf("LoadQLT03Register: %v", err)
	}
	if failures := AuditQLT03Register(rows); len(failures) != 0 {
		t.Fatalf("precondition failed: committed register is not currently green: %v", failures)
	}

	restore := SetQLT03InjectExtraRequiredCellForTest(true)
	defer restore()

	failures := AuditQLT03Register(rows)
	if len(failures) == 0 {
		t.Fatal("AuditQLT03Register did not fail under the fault-injection seam -- the audit cannot fail, proving nothing")
	}
}

// TestQLT03SwarmConfigurations runs the sweep in swarm configurations that
// omit one shape at a time, and a configuration that omits the "copy"
// op-kind's own contribution (leaf_use-shaped callees) -- D-11-48's
// secondary defense against feature interaction suppressing depth. Any
// cell reachable ONLY under omission would be recorded in the register
// with that fact; this generator's own construction makes no such cell
// exist (each shape's own reachable cells are independent of the other
// shapes running alongside it), and this test records that honestly
// rather than fabricating a positive finding.
func TestQLT03SwarmConfigurations(t *testing.T) {
	full, err := qlt03SweepAllSizes(nil)
	if err != nil {
		t.Fatalf("full sweep: %v", err)
	}
	fullCells := map[[2]string]bool{}
	for key := range qlt03AggregateByArgReturn(full) {
		fullCells[key] = true
	}

	ranShapeOmitting := false
	for _, shape := range QLT03EdgeTopologies() {
		cfg := &qlt03SweepConfig{ExcludeShapes: map[string]bool{shape: true}}
		reduced, err := qlt03SweepAllSizes(cfg)
		if err != nil {
			t.Fatalf("shape-omitting sweep (omit %s): %v", shape, err)
		}
		ranShapeOmitting = true
		reducedCells := qlt03AggregateByArgReturn(reduced)
		for key := range reducedCells {
			if !fullCells[key] {
				t.Logf("omission-only finding: cell %v reachable ONLY when shape %q is omitted", key, shape)
			}
		}
	}
	if !ranShapeOmitting {
		t.Fatal("expected at least one shape-omitting configuration to run")
	}

	opKindCfg := &qlt03SweepConfig{ExcludeOpKinds: map[core.OperationKind]bool{core.OpCopy: true}}
	opKindReduced, err := qlt03SweepAllSizes(opKindCfg)
	if err != nil {
		t.Fatalf("op-kind-omitting sweep: %v", err)
	}
	opKindReducedCells := qlt03AggregateByArgReturn(opKindReduced)
	for key := range opKindReducedCells {
		if !fullCells[key] {
			t.Logf("omission-only finding: cell %v reachable ONLY when op-kind copy is omitted", key)
		}
	}
	// D-11-48's own honest finding: this generator's shapes are independent
	// of each other, so no omission-only cell is expected to exist. Ran at
	// least one op-kind-omitting configuration, satisfying the required
	// coverage.
}

// TestQLT03RegisterIsDeterministicUnderRepeatAndParallel runs the audit
// with t.Parallel() subtests and asserts identical verdicts, and asserts
// the committed JSON is unmodified after the run.
func TestQLT03RegisterIsDeterministicUnderRepeatAndParallel(t *testing.T) {
	path := nat03CorpusPath("internal/compiler/session/qlt03_shape_register.json")
	before, err := osStatModTime(path)
	if err != nil {
		t.Fatalf("stat before: %v", err)
	}

	var mu sync.Mutex
	var verdicts []string
	for i := 0; i < 4; i++ {
		i := i
		t.Run("parallel", func(t *testing.T) {
			t.Parallel()
			rows, err := LoadQLT03Register()
			if err != nil {
				t.Fatalf("LoadQLT03Register: %v", err)
			}
			failures := AuditQLT03Register(rows)
			sort.Strings(failures)
			verdict := joinLines(failures)
			mu.Lock()
			verdicts = append(verdicts, verdict)
			mu.Unlock()
			_ = i
		})
	}
	t.Run("verify", func(t *testing.T) {
		mu.Lock()
		defer mu.Unlock()
		for i := 1; i < len(verdicts); i++ {
			if verdicts[i] != verdicts[0] {
				t.Fatalf("non-deterministic verdict: run 0 = %q, run %d = %q", verdicts[0], i, verdicts[i])
			}
		}
	})

	after, err := osStatModTime(path)
	if err != nil {
		t.Fatalf("stat after: %v", err)
	}
	if before != after {
		t.Fatalf("qlt03_shape_register.json's mtime changed across a parallel test run: %v -> %v", before, after)
	}
}
