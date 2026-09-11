package check

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/syntax"
)

// ---------------------------------------------------------------------
// Phase 09 Plan 07, Task 1: diagnostic-ordering-stability gate.
//
// D-09-13's risk: `core.Program.Functions` is iterated in declaration
// order today (checkInterproceduralLoanLiveness, check.go:844), and
// `Code + Span + Causes` fold into a published SHA-256 diagnostic
// identity (diagnostic.go:96-112, `Error`/`ErrorWithRepairs`). Moving
// `ownership.move_while_borrowed` / `ownership.borrow_conflict` from a
// per-function early exit (analyzeArmBody:2074, analyzeStraightLine:3287)
// to a single post-assembly pass over the WHOLE program changes WHEN a
// diagnostic is decided, which can change WHICH diagnostic a multi-function
// program reports FIRST -- an ordering change, not a content change, but
// one that still mints a different `diagnostic:<hash>` identity.
//
// A corpus replay cannot catch this: replaying the corpus against a golden
// file and regenerating the golden file the moment it differs absorbs the
// exact change this gate exists to catch. So this file states the RULE
// (declaration order decides which diagnostic is reported first) as an
// explicit assertion, THEN separately commits a literal pre-restructure
// baseline table that plan 09-09 must reproduce, or justify each
// difference for, in writing.
// ---------------------------------------------------------------------

// TestDiagnosticSelectionFollowsFunctionDeclarationOrder states the rule
// directly: given a two-function program where each function carries a
// KNOWN, DIFFERENT error (`name.unknown` in one, `ownership.move_while_borrowed`
// in the other), check.Program's FIRST reported diagnostic code is the one
// belonging to whichever function is declared FIRST in source -- and
// swapping the two functions' declaration order swaps which code is
// reported first. This is deliberately about the RULE (declaration order
// decides), not about one fixture's current output: a golden-file replay of
// either program alone would not distinguish "the rule holds" from "this
// particular ordering happens to still work."
//
// The second half of this test asserts the structural precondition the
// rule rests on: `core.Program.Functions` must be a SLICE (deterministic,
// declaration-ordered iteration), not a Go map (randomized iteration order
// per run) -- Pitfall 4's second, silent ordering hazard. A map here would
// make "declaration order decides" false regardless of what any single run
// happens to observe.
// D-09-08/D-09-09 (plan 09-09): the pre-restructure fixture pair here paired
// a LOWERING-TIME code (name.unknown) against a code
// (ownership.move_while_borrowed) that ALSO used to decide at lowering time.
// After D-09-08's authorized deletion, ownership.move_while_borrowed is
// DEFERRED to the post-assembly pass, which runs only once EVERY function's
// own intraprocedural admission has produced zero diagnostics program-wide
// (D-08-27's interim rule, unchanged by this plan) -- so pairing it against
// a genuinely lowering-time code no longer demonstrates "declaration order
// decides": name.unknown (a per-function lowering failure, present
// regardless of which function declares it first) would ALWAYS mask the
// deferred code's post-assembly pass from ever running, for EITHER
// declaration order, which is a real, understood consequence of the
// precedence rule (D-09-13) rather than a bug. The rule under test is
// re-demonstrated here with TWO functions that BOTH admit cleanly at
// lowering and BOTH carry a deferred code (`first`: ownership.borrow_conflict,
// `second`: ownership.move_while_borrowed) -- checkInterproceduralLoanLiveness
// iterates program.Functions in DECLARATION order and returns one diagnostic
// per offending function without stopping at the first, so result.Diagnostics[0]
// is exactly whichever function is declared first, and swapping declaration
// order swaps which is reported first, precisely as the original rule
// claims.
func TestDiagnosticSelectionFollowsFunctionDeclarationOrder(t *testing.T) {
	const firstThenSecond = `module test.ordering

export {
  fn first
  fn second
}

fn first(buffer: Buffer) -> Buffer {
  let a = borrow buffer
  let b = borrow mut buffer
  let observed = take a
  b
}

fn second(buffer: Buffer) -> Buffer {
  let view = borrow buffer
  let taken = take buffer
  let observed = take view
  taken
}
`
	const secondThenFirst = `module test.ordering

export {
  fn second
  fn first
}

fn second(buffer: Buffer) -> Buffer {
  let view = borrow buffer
  let taken = take buffer
  let observed = take view
  taken
}

fn first(buffer: Buffer) -> Buffer {
  let a = borrow buffer
  let b = borrow mut buffer
  let observed = take a
  b
}
`
	cases := []struct {
		name      string
		source    string
		wantFirst string
	}{
		{"first-declared-first", firstThenSecond, "ownership.borrow_conflict"},
		{"second-declared-first", secondThenFirst, "ownership.move_while_borrowed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parsed := syntax.Parse([]byte(tc.source))
			if len(parsed.Diagnostics) != 0 {
				t.Fatalf("synthetic program failed to parse: %+v", parsed.Diagnostics)
			}
			result := Program(parsed.Program)
			if len(result.Diagnostics) == 0 {
				t.Fatalf("expected at least one diagnostic, got none")
			}
			if got := result.Diagnostics[0].Code; got != tc.wantFirst {
				t.Fatalf("first reported diagnostic = %q, want %q (result.Diagnostics=%+v)", got, tc.wantFirst, result.Diagnostics)
			}
		})
	}

	// Structural precondition: the iterated collection must be a slice, not
	// a map. A reflective check here is what makes the ordering assertion
	// above meaningful across runs, not merely lucky on this one.
	var program core.Program
	field, ok := reflect.TypeOf(program).FieldByName("Functions")
	if !ok {
		t.Fatalf("core.Program has no Functions field")
	}
	if field.Type.Kind() != reflect.Slice {
		t.Fatalf("core.Program.Functions is a %s, not a slice -- declaration-order iteration is only deterministic over a slice", field.Type.Kind())
	}
}

// interproceduralOrderingBaselineDirs is the set of testdata directories
// TestInterproceduralDiagnosticOrderingStability walks. Matches the phase
// directories check_test.go's own corpus-wide tests already enumerate
// (TestCheckerVerdictsUnchanged's phase1..4 set, extended through the
// phase07/phase08 directories those later phases' own corpus-walking tests
// use).
var interproceduralOrderingBaselineDirs = []string{
	"phase1", "phase2", "phase3", "phase4", "phase5", "phase6", "phase07", "phase08",
}

// interproceduralOrderingBaseline is THE PRE-RESTRUCTURE BASELINE captured
// under plan 09-07, before `computeLoanLastUses` (check.go:3743) and its two
// summary-blind call sites are deleted. For every `.lang` fixture under
// testdata/ that check.Program currently REFUSES, this table records the
// exact tuple (fixture path relative to testdata/, first diagnostic Code,
// first diagnostic ID). Generated from the tree at authoring time via a
// throwaway walk over the same corpus TestInterproceduralDiagnosticOrderingStability
// re-walks below; committed here as literal data, not regenerated.
//
// Plan 09-09 (the deletion plan) MUST reproduce this table exactly, or
// justify each difference in its own SUMMARY.md: name the fixture, its old
// and new tuple, classify the difference (a genuinely different defect now
// reported first -- the D-09-13 risk realized -- or the same defect
// carrying a different span/cause list -- an identity change), and state in
// one sentence why the new behavior is correct. A silent regeneration of
// this table is never an acceptable resolution; that is the exact failure
// mode a golden-file replay would have permitted and this literal table is
// built to prevent.
var interproceduralOrderingBaseline = map[string][2]string{
	"phase07/call_argument_used_twice.lang":    {"ownership.use_after_move", "diagnostic:8e3116d77bcb23bd6b84396b"},
	"phase07/call_type_mismatch.lang":          {"check.call_argument_type_mismatch", "diagnostic:eec74c3869e957e7eccaafb8"},
	"phase07/call_uncallable_callee.lang":      {"core.callee_not_callable", "diagnostic:5735171812e7ad4920c3bb72"},
	"phase07/cycle_indirect.lang":              {"core.call_graph_cycle", "diagnostic:1e6c260432c9010ac6a196a9"},
	"phase07/cycle_mutual.lang":                {"core.call_graph_cycle", "diagnostic:39bb0a1a48d08fc9674307a4"},
	"phase07/cycle_self.lang":                  {"core.call_graph_cycle", "diagnostic:6e836fd6f98202bf58f1dfdb"},
	"phase07/cycle_through_match_arm.lang":     {"core.call_graph_cycle", "diagnostic:0fe333002fca961ebee8f00a"},
	"phase07/cycle_unreachable.lang":           {"core.call_graph_cycle", "diagnostic:03edcf9d691da106106c2fc0"},
	"phase07/foreign_symbol_shadowing.lang":    {"core.call_graph_cycle", "diagnostic:3fa66ee8773176867ffb9b4e"},
	"phase07/relay_escort_witness.lang":        {"check.interprocedural_loan_liveness", "diagnostic:58c1b5b2072cda60f77d721e"},
	"phase08/negative_control_fails.lang":      {"check.interprocedural_loan_liveness", "diagnostic:893138210397ed2cd76b7655"},
	"phase08/negative_control_infallible.lang": {"check.interprocedural_loan_liveness", "diagnostic:c0950895fc782a3440902dfd"},
	"phase08/relay_depth2_refuse.lang":         {"check.interprocedural_loan_liveness", "diagnostic:300248748c05f20eb1fe3948"},
	"phase08/twin_a_refuse.lang":               {"check.interprocedural_loan_liveness", "diagnostic:e01ad6316e27899deeb610f7"},
	// phase08/twin_b_accept.lang and phase08/twin_b_refuse.lang, plan 09-09:
	// BOTH now report check.interprocedural_loan_liveness instead of
	// ownership.move_while_borrowed -- an IDENTITY change compounded with a
	// found-but-not-fixed pre-existing defect, not merely an ordering change.
	// D-09-08's authorized deletion correctly stopped the summary-blind
	// lowering-time gate from masking these two fixtures before the real
	// interprocedural pass ever ran (D-08-41's whole point). But Pattern B's
	// promised SPLIT (twin_b_refuse refused, twin_b_accept ADMITTED) does NOT
	// materialize: deriveFunctionUsesParam's `default: usesParam = true`
	// branch (check.go, pre-existing since Phase 08, untouched by this plan)
	// incorrectly treats a plain OpMove/OpCopy forward of the parameter as a
	// "use" even when the chain leads directly to the function's own
	// terminating return -- so twin_b_accept's `probe` (`taken = take
	// buffer; taken`, intended usesParam=false per the fixture's own header)
	// is ALSO derived as usesParam=true, collapsing the interprocedural
	// distinguishing signal Pattern B depends on. Both twins now report the
	// SAME code via the SAME backward-direction (D-08-08) classification.
	// This is a genuine, diagnosed, pre-existing defect this deletion merely
	// stopped masking (never introduced) -- fixing deriveFunctionUsesParam is
	// out of this plan's scope (Task 2 authorizes extending
	// checkInterproceduralLoanLiveness only) and is recorded as new debt
	// (see this plan's SUMMARY and PHASE-09-DEBT.md) rather than adjusted
	// here.
	"phase08/twin_b_accept.lang":   {"check.interprocedural_loan_liveness", "diagnostic:5fe0501da5249e177b0ada39"},
	"phase08/twin_b_refuse.lang":   {"check.interprocedural_loan_liveness", "diagnostic:d15b65a04fd8515f92e999e2"},
	"phase1/malformed.lang":        {"syntax.unexpected_byte", "diagnostic:ccb9bcd29f3e8d96fa0368b6"},
	"phase1/non_exhaustive.lang":   {"match.non_exhaustive", "diagnostic:f9582fb8c4ad9f90fbe75fa8"},
	"phase2/ability_shapes.lang":   {"check.unexecutable_shape", "diagnostic:7da7df4418c3945f03ab2d82"},
	"phase2/implicit_noncopy.lang": {"ownership.transfer_requires_take", "diagnostic:040662ef397be67eebb23003"},
	// phase2/move_while_borrowed.lang, plan 09-09: SPAN-only identity change.
	// `let observed = view` (an implicit copy of the borrowed view, Buffer's
	// only source-level way to keep the loan referenced past the move) was
	// rewritten to `let observed = take view`: once lowering stops deciding
	// ownership.move_while_borrowed inline and proceeds through every
	// binding unconditionally (D-09-09), the OLD implicit copy would newly
	// fail with the timing-INDEPENDENT ownership.transfer_requires_take
	// (Buffer withholds Copy) BEFORE the post-assembly pass ever reaches the
	// deferred verdict this fixture exists to demonstrate -- the precedence
	// rule (D-09-13) correctly masking the intended defect with a real but
	// unintended one. `take view` preserves the loan's later reference
	// without requiring Copy, so the SAME code fires; only the source bytes
	// (and therefore the span-derived ID) moved.
	"phase2/move_while_borrowed.lang":  {"ownership.move_while_borrowed", "diagnostic:3c53a97b806dac0a128d2902"},
	"phase2/reborrow_while_moved.lang": {"ownership.move_while_borrowed", "diagnostic:eeddc92047e60cb51c3fd65a"},
	"phase2/use_after_move.lang":       {"ownership.use_after_move", "diagnostic:26c8fdff5f83afd58fb43905"},
	// phase3/branch_one_arm_shared_reject.lang, plan 09-09: IDENTITY change,
	// source untouched. Code stays ownership.move_while_borrowed, but the ID
	// moved because the diagnostic is now built by
	// moveWhileBorrowedDiagnosticPostAssembly (check.go) at the post-assembly
	// decision point instead of inline during analyzeArmBody's lowering-time
	// walk -- the two constructions are believed cause-for-cause equivalent,
	// but a byte-level difference in a reconstructed span or cause Detail
	// string is plausible and is exactly what a hash-based identity is
	// designed to catch rather than let pass silently.
	"phase3/branch_one_arm_shared_reject.lang": {"ownership.move_while_borrowed", "diagnostic:56f0f78afd023dba87d78a24"},
	// phase3/exclusive_exclusive_reject.lang and phase3/shared_exclusive_reject.lang
	// and phase3/exclusive_move_reject.lang, plan 09-09: SPAN-only identity
	// change, same masking-avoidance rewrite as phase2/move_while_borrowed.lang
	// above (`let observed = first`/`let observed = view` -> `let observed =
	// take first`/`let observed = take view`) -- Buffer withholds Copy, and
	// these three fixtures are ALSO consumed directly by
	// internal/compiler/session's TestBorrowConflictMatrix/
	// TestBorrowConflictCauseChain/TestExclusiveMoveRejected (out of this
	// plan's file scope), which independently confirm the intended CODE
	// still fires after the rewrite.
	"phase3/exclusive_exclusive_reject.lang":      {"ownership.borrow_conflict", "diagnostic:6a3d8582000b05e32a575ebf"},
	"phase3/exclusive_move_reject.lang":           {"ownership.move_while_borrowed", "diagnostic:8038bd1955c8b4cbd98b445d"},
	"phase3/shared_exclusive_reject.lang":         {"ownership.borrow_conflict", "diagnostic:2ac14b431c0d5a98db75e10d"},
	"phase4/fallible_call_unconsumed.lang":        {"syntax.fallible_call_not_consumed", "diagnostic:91c8b8c7a5d359d9a14fe6a8"},
	"phase4/foreign_call_target_not_foreign.lang": {"core.call_target_not_foreign", "diagnostic:8ee0be5e21030f21f990f9f6"},
	"phase4/foreign_policy_value_injection.lang":  {"check.foreign_policy_value_unsafe", "diagnostic:204a40c7d8537f0809622fbb"},
	"phase4/foreign_unwind_undeclared.lang":       {"foreign.unwind_policy_undeclared", "diagnostic:e9e51b10ac76db2d660d791b"},
	"phase5/explain_use_after_move.lang":          {"ownership.use_after_move", "diagnostic:d8b679b47be2fa6555514f78"},
}

// TestInterproceduralDiagnosticOrderingStability re-walks the SAME corpus
// interproceduralOrderingBaseline was generated from and asserts every
// currently-refused fixture's (code, id) tuple matches the committed
// baseline exactly. A fixture that check.Program no longer refuses, or
// refuses with a different code/id than the baseline names, fails this test
// -- which is the point: it converts "the baseline still holds" from an
// assumption into something this test run itself checks, both today (must
// pass, proving the table was transcribed correctly) and after plan 09-09's
// restructure (must pass, or the difference must be justified and the table
// entry updated in that plan's own commit).
func TestInterproceduralDiagnosticOrderingStability(t *testing.T) {
	var paths []string
	for _, dir := range interproceduralOrderingBaselineDirs {
		matches, err := filepath.Glob(filepath.Join("../../../testdata", dir, "*.lang"))
		if err != nil {
			t.Fatalf("glob testdata/%s: %v", dir, err)
		}
		paths = append(paths, matches...)
	}
	sort.Strings(paths)
	if len(paths) == 0 {
		t.Fatal("expected at least one fixture under testdata/")
	}

	seen := map[string]bool{}
	for _, path := range paths {
		rel, err := filepath.Rel("../../../testdata", path)
		if err != nil {
			t.Fatalf("rel %s: %v", path, err)
		}
		rel = filepath.ToSlash(rel)

		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		parsed := syntax.Parse(source)
		var code, id string
		if len(parsed.Diagnostics) > 0 {
			code, id = parsed.Diagnostics[0].Code, parsed.Diagnostics[0].ID
		} else {
			result := Program(parsed.Program)
			if len(result.Diagnostics) > 0 {
				code, id = result.Diagnostics[0].Code, result.Diagnostics[0].ID
			}
		}

		want, tracked := interproceduralOrderingBaseline[rel]
		if code == "" {
			if tracked {
				t.Fatalf("%s: baseline expected refusal %+v, but the fixture now checks clean", rel, want)
			}
			continue
		}
		if !tracked {
			t.Fatalf("%s: newly refused with code %q, id %q -- not present in the pre-restructure baseline; if this is expected, this table must be updated with a written justification, never silently", rel, code, id)
		}
		seen[rel] = true
		if code != want[0] || id != want[1] {
			t.Fatalf("%s: baseline mismatch: want (code=%q, id=%q), got (code=%q, id=%q) -- a difference here must be justified in the owning plan's SUMMARY, not silently regenerated", rel, want[0], want[1], code, id)
		}
	}
	for rel := range interproceduralOrderingBaseline {
		if !seen[rel] {
			t.Fatalf("baseline entry %s was not found among the walked fixtures -- corpus shrank or the entry is stale", rel)
		}
	}
}
