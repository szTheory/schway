package check

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/ast"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/syntax"
)

// ---------------------------------------------------------------------
// Phase 09 Plan 07, Tasks 2 and 3: settling D-09-49 Q1 by enumeration, and
// the AST-shadow-path reachability corpus.
//
// Both tasks share this file because both are about the SAME summary-blind
// mechanism -- computeLoanLastUses (check.go:3743, fed by
// analyzeArmBody:2074 and analyzeStraightLine:3287) -- from two angles:
// Task 2 asks whether removing its inline `move_while_borrowed` short
// circuit changes what `use_after_move` observes for any EXISTING fixture;
// Task 3 asks which of the five `ownership.*` codes the shadow machinery
// ITSELF (activeLoans/expiringLoans, fed by computeLoanLastUses' loanUses)
// actually decides, versus which fire from immediate per-binding facts
// regardless of that machinery's answer.
// ---------------------------------------------------------------------

// ===== Task 2: does deferring move_while_borrowed change use_after_move? =====

// moveWhileBorrowedFixtures is the ENUMERATED set of every `.lang` fixture
// under testdata/ where `ownership.move_while_borrowed` currently fires,
// named by the function and parameter the diagnostic's own cause chain
// blames. TestUseAfterMoveUnchangedByDeferredMoveWhileBorrowed both
// verifies this set is exhaustive (via testMoveWhileBorrowedCorpusIsExhaustive,
// which independently re-derives the same set from the corpus dynamically)
// and applies the positive per-fixture property to each entry.
//
// Plan 09-09 (D-09-08's authorized deletion): the phase08/twin_b_accept.lang
// and phase08/twin_b_refuse.lang entries plan 09-07 recorded here are
// REMOVED. Both fixtures now report check.interprocedural_loan_liveness
// instead of ownership.move_while_borrowed (see
// check_ordering_stability_test.go's twin_b_accept/refuse baseline comment
// and check_test.go's TestInterproceduralLivenessTwinPatternBRealFixtures
// for the full, diagnosed explanation -- a pre-existing deriveFunctionUsesParam
// defect this deletion stopped masking, recorded as new debt, not fixed
// here) -- they are no longer members of the set THIS enumeration is about.
var moveWhileBorrowedFixtures = []struct {
	path      string
	function  string
	parameter string
}{
	{"phase2/move_while_borrowed.lang", "relay", "buffer"},
	{"phase2/reborrow_while_moved.lang", "relay", "code"},
	{"phase3/branch_one_arm_shared_reject.lang", "choose", "flag"},
	{"phase3/exclusive_move_reject.lang", "relay", "buffer"},
}

// linearBodiesOf returns every straight-line body a function declares: its
// own Body.Linear for a straight-line function, or each match arm's own
// Body for a branch-shaped one. Both shapes are checked by the same
// analyzeStraightLine/analyzeArmBody machinery this file's tests reason
// about.
func linearBodiesOf(function ast.FuncDecl) []*ast.LinearBody {
	if function.Body.Linear != nil {
		return []*ast.LinearBody{function.Body.Linear}
	}
	var bodies []*ast.LinearBody
	for index := range function.Body.Arms {
		if function.Body.Arms[index].Body != nil {
			bodies = append(bodies, function.Body.Arms[index].Body)
		}
	}
	return bodies
}

// moveWithNoSubsequentAccess implements Task 2(a)'s positive, checkable
// property over one LinearBody: find the binding that TAKES parameterName
// (the offending move `ownership.move_while_borrowed` blames), then check
// whether ANY binding strictly after it -- including the body's own Result
// position -- still references parameterName (as a plain source, or as a
// call argument). A true return for `violated` is the concrete D-09-49 Q1
// risk: lowering proceeding past a deferred move far enough to ALSO trip
// `ownership.use_after_move` on the same place.
func moveWithNoSubsequentAccess(body *ast.LinearBody, parameterName string) (moveFound bool, moveIndex int, violated bool, violatingIndex int) {
	for index, binding := range body.Bindings {
		if binding.RHS.Kind != "take" || binding.RHS.Source != parameterName {
			continue
		}
		moveFound, moveIndex = true, index
		for later := index + 1; later < len(body.Bindings); later++ {
			candidate := body.Bindings[later]
			if candidate.RHS.Kind == "call" {
				for _, argument := range candidate.RHS.Arguments {
					if argument == parameterName {
						return moveFound, moveIndex, true, later
					}
				}
				continue
			}
			if candidate.RHS.Source == parameterName {
				return moveFound, moveIndex, true, later
			}
		}
		if body.Result == parameterName {
			return moveFound, moveIndex, true, len(body.Bindings)
		}
		return moveFound, moveIndex, false, -1
	}
	return false, -1, false, -1
}

// testMoveWhileBorrowedCorpusIsExhaustive independently re-derives, by
// walking the WHOLE testdata/ corpus through the real check.Program
// pipeline, the set of fixtures where `ownership.move_while_borrowed`
// fires anywhere in the reported diagnostics -- and asserts it is EXACTLY
// moveWhileBorrowedFixtures. This is what makes "enumerate, do not sample"
// true of this table: a new fixture added later that also triggers this
// code, or a corpus fixture that stops triggering it, fails this check
// rather than being silently absorbed.
func testMoveWhileBorrowedCorpusIsExhaustive(t *testing.T) map[string]bool {
	t.Helper()
	dirs := []string{"phase1", "phase2", "phase3", "phase4", "phase5", "phase6", "phase07", "phase08"}
	var paths []string
	for _, dir := range dirs {
		matches, err := filepath.Glob(filepath.Join("../../../testdata", dir, "*.lang"))
		if err != nil {
			t.Fatalf("glob testdata/%s: %v", dir, err)
		}
		paths = append(paths, matches...)
	}
	sort.Strings(paths)

	found := map[string]bool{}
	for _, path := range paths {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		parsed := syntax.Parse(source)
		if len(parsed.Diagnostics) > 0 {
			continue
		}
		result := Program(parsed.Program)
		for _, diag := range result.Diagnostics {
			if diag.Code == "ownership.move_while_borrowed" {
				rel, _ := filepath.Rel("../../../testdata", path)
				found[filepath.ToSlash(rel)] = true
			}
		}
	}

	want := map[string]bool{}
	for _, entry := range moveWhileBorrowedFixtures {
		want[entry.path] = true
	}
	if len(found) != len(want) {
		t.Fatalf("dynamic move_while_borrowed scan found %d fixtures %v, table declares %d %v -- enumeration drifted from the corpus", len(found), found, len(want), want)
	}
	for path := range want {
		if !found[path] {
			t.Fatalf("table declares %s but the dynamic scan no longer finds ownership.move_while_borrowed there", path)
		}
	}
	return found
}

// TestUseAfterMoveUnchangedByDeferredMoveWhileBorrowed is D-09-49 Q1's
// settling test. Q1 asks: does deferring `move_while_borrowed`'s decision
// to post-assembly change `use_after_move`'s behavior for any EXISTING
// fixture? The mechanism: today's inline refusal short-circuits BEFORE any
// later access to the same moved place is reached; a deferred refusal lets
// lowering proceed, so if any such later access exists, lowering would ALSO
// trip use_after_move -- two errors where one existed, and a different
// first-reported code.
//
// ANSWER (recorded here, restated in this plan's SUMMARY): no existing
// fixture exhibits the interaction. Enumerated over all 6 fixtures where
// ownership.move_while_borrowed currently fires, in every case the moved
// place's only later reference is through the BORROWED VIEW binding (a
// different name, already holding the alias), never the moved place's own
// name again -- so no fixture's own place-access shape can trip
// use_after_move once the move is no longer refused inline.
func TestUseAfterMoveUnchangedByDeferredMoveWhileBorrowed(t *testing.T) {
	found := testMoveWhileBorrowedCorpusIsExhaustive(t)
	if len(found) != len(moveWhileBorrowedFixtures) {
		t.Fatalf("exhaustiveness check disagreed with the table length; see the corpus-scan sub-test failure above")
	}

	for _, entry := range moveWhileBorrowedFixtures {
		t.Run(entry.path, func(t *testing.T) {
			source, err := os.ReadFile(filepath.Join("../../../testdata", entry.path))
			if err != nil {
				t.Fatalf("read %s: %v", entry.path, err)
			}
			program := mustParseProgram(t, source)
			var function *ast.FuncDecl
			for index := range program.Funcs {
				if program.Funcs[index].Name == entry.function {
					function = &program.Funcs[index]
					break
				}
			}
			if function == nil {
				t.Fatalf("%s: function %q not found", entry.path, entry.function)
			}
			var moveFound bool
			var violated bool
			var moveIndex, violatingIndex int
			for _, body := range linearBodiesOf(*function) {
				found, index, isViolated, atIndex := moveWithNoSubsequentAccess(body, entry.parameter)
				if found {
					moveFound, moveIndex = true, index
					if isViolated {
						violated, violatingIndex = true, atIndex
					}
				}
			}
			if !moveFound {
				t.Fatalf("%s: expected a `take %s` binding in fn %s, found none -- the fixture no longer matches this table's shape", entry.path, entry.parameter, entry.function)
			}
			if violated {
				t.Fatalf(
					"%s: fn %s moves %q at binding %d, and binding/result %d still references %q afterward -- this is the concrete D-09-49 Q1 risk realized; this fixture must be added to an exception set with a stated expected post-restructure behavior and escalated to plan 09-08's gate, not silently passed",
					entry.path, entry.function, entry.parameter, moveIndex, violatingIndex, entry.parameter,
				)
			}
		})
	}
}

// ===== Task 2(b): fence the three timing-independent codes =====

// checkSourceLines lazily reads and caches check.go's own source lines for
// the source-scan half of the fence below. Read once per test binary run.
var checkSourceLinesCache []string

func checkSourceLines(t *testing.T) []string {
	t.Helper()
	if checkSourceLinesCache != nil {
		return checkSourceLinesCache
	}
	content, err := os.ReadFile("check.go")
	if err != nil {
		t.Fatalf("read check.go: %v", err)
	}
	checkSourceLinesCache = strings.Split(string(content), "\n")
	return checkSourceLinesCache
}

// findGuardAbove searches backward from anchorIndex (inclusive) for the
// nearest line matching pattern, within maxLookback lines. This is how this
// file locates a raise site's OWN guarding condition without assuming a
// fixed line number: check.go is explicitly prohibited from changing in
// this plan, but the technique itself must survive plan 09-09's restructure
// re-running this same fence.
func findGuardAbove(lines []string, anchorIndex int, pattern *regexp.Regexp, maxLookback int) (guardIndex int, ok bool) {
	for index := anchorIndex; index >= 0 && anchorIndex-index <= maxLookback; index-- {
		if pattern.MatchString(lines[index]) {
			return index, true
		}
	}
	return -1, false
}

var (
	initializedGuardPattern  = regexp.MustCompile(`if\s+!\S+\.initialized\s*\{`)
	hasShareGuardPattern     = regexp.MustCompile(`hasTypeAbility\(typeFact,\s*core\.AbilityShare\)`)
	hasCopyGuardPattern      = regexp.MustCompile(`hasTypeAbility\(typeFact,\s*core\.AbilityCopy\)`)
	closureDefinitionPattern = regexp.MustCompile(`useAfterMove\s*:=\s*func\(`)
	useAfterMoveCallPattern  = regexp.MustCompile(`useAfterMove\(`)
)

// timingIndependentGuardWindows locates every raise site for the three
// timing-independent codes and returns, for each, the text window from its
// guarding condition down to the raise site itself. Two of the three codes
// (borrow_requires_share, transfer_requires_take) guard inline, immediately
// above the diagnostic construction; use_after_move's guard sits at the
// CALL site of a shared `useAfterMove` closure (analyzeArmBody,
// analyzeStraightLine) or inline in resolveCallBinding -- this function
// finds the right anchor for each shape rather than assuming one.
func timingIndependentGuardWindows(t *testing.T, code string) []string {
	t.Helper()
	lines := checkSourceLines(t)
	var windows []string

	switch code {
	case "ownership.use_after_move":
		for index, line := range lines {
			if useAfterMoveCallPattern.MatchString(line) {
				guardIndex, ok := findGuardAbove(lines, index, initializedGuardPattern, 5)
				if !ok {
					t.Fatalf("check.go:%d: useAfterMove(...) call has no `if !X.initialized {` guard within 5 lines above", index+1)
				}
				windows = append(windows, strings.Join(lines[guardIndex:index+1], "\n"))
			}
		}
		for index, line := range lines {
			if !strings.Contains(line, `"ownership.use_after_move"`) {
				continue
			}
			if closureDefinitionPattern.MatchString(strings.Join(lines[max(0, index-25):index], "\n")) {
				// The closure's OWN definition -- its guard lives at each
				// call site, already covered by the loop above.
				continue
			}
			guardIndex, ok := findGuardAbove(lines, index, initializedGuardPattern, 15)
			if !ok {
				t.Fatalf("check.go:%d: inline ownership.use_after_move raise has no `if !X.initialized {` guard within 15 lines above", index+1)
			}
			windows = append(windows, strings.Join(lines[guardIndex:index+1], "\n"))
		}
	case "ownership.borrow_requires_share":
		for index, line := range lines {
			if !strings.Contains(line, `"ownership.borrow_requires_share"`) {
				continue
			}
			guardIndex, ok := findGuardAbove(lines, index, hasShareGuardPattern, 15)
			if !ok {
				t.Fatalf("check.go:%d: ownership.borrow_requires_share raise has no hasTypeAbility(..., AbilityShare) guard within 15 lines above", index+1)
			}
			windows = append(windows, strings.Join(lines[guardIndex:index+1], "\n"))
		}
	case "ownership.transfer_requires_take":
		for index, line := range lines {
			if !strings.Contains(line, `"ownership.transfer_requires_take"`) {
				continue
			}
			guardIndex, ok := findGuardAbove(lines, index, hasCopyGuardPattern, 15)
			if !ok {
				t.Fatalf("check.go:%d: ownership.transfer_requires_take raise has no hasTypeAbility(..., AbilityCopy) guard within 15 lines above", index+1)
			}
			windows = append(windows, strings.Join(lines[guardIndex:index+1], "\n"))
		}
	default:
		t.Fatalf("unhandled code %q", code)
	}
	if len(windows) == 0 {
		t.Fatalf("found zero raise sites for %q -- the code may have been retired, which is out of OWN-09's scope", code)
	}
	return windows
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// TestTimingIndependentOwnershipCodesFireFromPerBindingFacts is D-09-47's
// own fence, made checkable rather than merely stated: `ownership.use_after_move`,
// `ownership.borrow_requires_share`, and `ownership.transfer_requires_take`
// must keep firing from immediate per-binding facts (`!source.initialized`,
// `hasTypeAbility`), never from the timing machinery
// (`activeLoans`/`loanUses`) plan 09-09 restructures. Two halves: POSITIVE
// (each code still fires on a fixture that triggers it, through the real
// pipeline or, where the language has no source-reachable type lacking
// Share, through the same synthetic-type-fact technique
// nonShareableTypeFact already establishes) and STRUCTURAL (every raise
// site's own guarding condition mentions `initialized` or `hasTypeAbility`
// and mentions NEITHER `activeLoans` NOR `loanUses`).
func TestTimingIndependentOwnershipCodesFireFromPerBindingFacts(t *testing.T) {
	t.Run("use_after_move fires on a real fixture", func(t *testing.T) {
		source, err := os.ReadFile("../../../testdata/phase2/use_after_move.lang")
		if err != nil {
			t.Fatalf("read fixture: %v", err)
		}
		result := Program(mustParseProgram(t, source))
		if len(result.Diagnostics) == 0 || result.Diagnostics[0].Code != "ownership.use_after_move" {
			t.Fatalf("expected ownership.use_after_move, got %+v", result.Diagnostics)
		}
	})
	t.Run("transfer_requires_take fires on a real fixture", func(t *testing.T) {
		source, err := os.ReadFile("../../../testdata/phase2/implicit_noncopy.lang")
		if err != nil {
			t.Fatalf("read fixture: %v", err)
		}
		result := Program(mustParseProgram(t, source))
		if len(result.Diagnostics) == 0 || result.Diagnostics[0].Code != "ownership.transfer_requires_take" {
			t.Fatalf("expected ownership.transfer_requires_take, got %+v", result.Diagnostics)
		}
	})
	t.Run("borrow_requires_share fires on the established synthetic-type fixture", func(t *testing.T) {
		// No source-reachable Phase 1-8 type constructor withholds Share
		// (nonShareableTypeFact's own doc comment, check_test.go) -- this is
		// the SAME synthetic-type technique TestOwnershipSequenceExhaustive
		// already uses to exercise this gate, not a new one invented here.
		denied := ast.LinearBody{
			Bindings: []ast.Binding{binding("view", "borrow", "owner", 10)},
			Result:   "view",
			Span:     diagnostic.Span{Start: 10, End: 20},
		}
		got := analyzeStraightLine("test:fence", "owner", diagnostic.Span{Start: 1, End: 6}, nonShareableTypeFact(), &denied, nil, nil)
		if got.DiagnosticCode != "ownership.borrow_requires_share" {
			t.Fatalf("expected ownership.borrow_requires_share, got %+v", got)
		}
	})

	for _, code := range []string{"ownership.use_after_move", "ownership.borrow_requires_share", "ownership.transfer_requires_take"} {
		t.Run("guard structure: "+code, func(t *testing.T) {
			windows := timingIndependentGuardWindows(t, code)
			for _, window := range windows {
				if strings.Contains(window, "activeLoans") || strings.Contains(window, "loanUses") {
					t.Fatalf("%s: guarding window mentions activeLoans/loanUses -- this code must fire from per-binding facts only:\n%s", code, window)
				}
				hasInitialized := strings.Contains(window, "initialized")
				hasAbility := strings.Contains(window, "hasTypeAbility")
				if !hasInitialized && !hasAbility {
					t.Fatalf("%s: guarding window mentions neither `initialized` nor `hasTypeAbility`:\n%s", code, window)
				}
			}
		})
	}
}

// ===== Task 3: the AST-shadow-path reachability corpus =====
//
// Generator Reachability Register (this file's own declared scope, per this
// plan's own requirement that an undeclared unreached space is the failure
// mode the standing process rules exist to prevent): the corpus below
// extends generatedOwnershipBody/branchSequenceProgram (check_test.go), the
// SAME 48-symbol alphabet TestOwnershipSequenceExhaustive and
// TestBranchSequenceExhaustive already enumerate exhaustively. It reaches
// every combination of {implicit copy, take, borrow, borrow_mut} sourced
// from {every visible place, an out-of-scope name} at bounded sequence
// length, over a two-arm branch (mirroring checkBranch's own topology). It
// does NOT reach, and cannot be extended by this technique alone to reach:
// loop-carried liveness, loop-exit edges, or per-iteration loan identity
// (this language has no loops, so no corpus can exercise that); or arity > 1
// (every callable this language can declare takes exactly one parameter,
// D-07-07). Any claim below is scoped to the language this generator can
// express, not to a hypothetical superset of it.
//
// Operational definition, RE-DERIVED for the surviving pass (plan 09-09,
// D-09-08's authorized deletion): the AST-shadow path (computeLoanLastUses
// and its two summary-blind lowering-time call sites) is GONE. A code now
// counts as reachable through the SURVIVING post-assembly timing law
// (checkInterproceduralLoanLiveness's extension to the purely-intraprocedural
// case) if and only if its own raise site's guarding logic consults
// `lastUseIndexByLoan` -- the surviving pass's OWN last-use derivation
// (materializeLoanEndpoints' output, mapped to operation index), playing
// exactly the role the deleted activeLoans/loanUses state used to play.
// `ownership.use_after_move`, `ownership.borrow_requires_share`, and
// `ownership.transfer_requires_take` are unaffected: they still fire inline
// during lowering from immediate per-binding facts that never consult any
// loan-timing state (Task 2(b)'s fence, TestTimingIndependentOwnershipCodesFireFromPerBindingFacts,
// still applies unchanged). The PRE-DELETION sets plan 09-07 recorded are
// retained below as historical record, immediately followed by the
// re-derived, currently-checked register.
//
// HISTORICAL (plan 09-07, pre-deletion): shadowPathReachableCodes =
// {ownership.move_while_borrowed: guarded by `if loans :=
// activeLoans[source.place.ID]; len(loans) > 0`, analyzeArmBody/
// analyzeStraightLine's `take` case; ownership.borrow_conflict: guarded by
// `conflictingLoan(activeLoans[source.place.ID], ...)`, the `borrow`/
// `borrow_mut` cases} -- both driven by computeLoanLastUses' loanUses
// output, both now DELETED call sites.

// shadowPathReachableCodes is the set of `ownership.*` codes whose admission
// decision is genuinely gated by the SURVIVING pass's own loan-timing
// derivation (lastUseIndexByLoan, checkInterproceduralLoanLiveness). Both
// entries are the two codes D-09-47 predicted and plan 09-07 already proved
// reachable through the (now-deleted) shadow path; TestShadowPathSubsumptionCorpus
// verifies the SURVIVING pass's actual reach still agrees.
var shadowPathReachableCodes = map[string]string{
	"ownership.move_while_borrowed": "guarded by `lastUseIndexByLoan[loanID]` inline in moveWhileBorrowedDiagnosticPostAssembly's own call site (check.go, checkInterproceduralLoanLiveness) -- lastUseIndexByLoan is materializeLoanEndpoints' own last-use derivation over the assembled function, the surviving pass's replacement for the deleted activeLoans/loanUses state",
	"ownership.borrow_conflict":     "guarded by `lastUseIndexByLoan` in the new-loan-vs-existing-loan conflict scan (check.go, checkInterproceduralLoanLiveness) immediately preceding borrowConflictDiagnosticPostAssembly's own call site -- the SAME lastUseIndexByLoan state move_while_borrowed's guard reads",
}

// shadowPathUnreachableCodes is the set of `ownership.*` codes the surviving
// pass does NOT decide: each still fires inline during lowering from an
// immediate per-binding fact that never consults lastUseIndexByLoan (or the
// deleted activeLoans/loanUses before it), verified by the identical
// source-scan technique TestTimingIndependentOwnershipCodesFireFromPerBindingFacts
// applies to the same three raise sites.
var shadowPathUnreachableCodes = map[string]string{
	"ownership.use_after_move":         "guard is `!source.initialized` only -- a per-binding fact set the instant a prior `take` runs; never consults lastUseIndexByLoan or the deleted activeLoans/loanUses",
	"ownership.borrow_requires_share":  "guard is `hasTypeAbility(typeFact, core.AbilityShare)` only -- a declared-type fact; never consults lastUseIndexByLoan or the deleted activeLoans/loanUses",
	"ownership.transfer_requires_take": "guard is `hasTypeAbility(typeFact, core.AbilityCopy)` only -- a declared-type fact; never consults lastUseIndexByLoan or the deleted activeLoans/loanUses",
}

var (
	lastUseIndexByLoanGuardPattern       = regexp.MustCompile(`lastUseIndexByLoan`)
	moveWhileBorrowedPostAssemblyPattern = regexp.MustCompile(`moveWhileBorrowedDiagnosticPostAssembly\(`)
	borrowConflictPostAssemblyPattern    = regexp.MustCompile(`borrowConflictDiagnosticPostAssembly\(`)
)

// reachableGuardWindows is timingIndependentGuardWindows' mirror image for
// the two SURVIVING-pass-DEPENDENT codes: it locates each raise site's own
// guarding condition and returns the text window from guard to raise site,
// so the caller can assert the window DOES consult lastUseIndexByLoan (the
// opposite assertion Task 2(b)'s fence makes for the three independent
// codes).
func reachableGuardWindows(t *testing.T, code string) []string {
	t.Helper()
	lines := checkSourceLines(t)
	var windows []string

	switch code {
	case "ownership.move_while_borrowed":
		for index, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "func ") {
				continue
			}
			if !moveWhileBorrowedPostAssemblyPattern.MatchString(line) {
				continue
			}
			guardIndex, ok := findGuardAbove(lines, index, lastUseIndexByLoanGuardPattern, 20)
			if !ok {
				t.Fatalf("check.go:%d: moveWhileBorrowedDiagnosticPostAssembly(...) call has no lastUseIndexByLoan guard within 20 lines above", index+1)
			}
			windows = append(windows, strings.Join(lines[guardIndex:index+1], "\n"))
		}
	case "ownership.borrow_conflict":
		for index, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "func ") {
				// Skip borrowConflictDiagnosticPostAssembly's own definition
				// line: its guard lives at each CALL site, not at its own
				// signature.
				continue
			}
			if !borrowConflictPostAssemblyPattern.MatchString(line) {
				continue
			}
			guardIndex, ok := findGuardAbove(lines, index, lastUseIndexByLoanGuardPattern, 20)
			if !ok {
				t.Fatalf("check.go:%d: borrowConflictDiagnosticPostAssembly(...) call has no lastUseIndexByLoan guard within 20 lines above", index+1)
			}
			windows = append(windows, strings.Join(lines[guardIndex:index+1], "\n"))
		}
	default:
		t.Fatalf("unhandled code %q", code)
	}
	if len(windows) == 0 {
		t.Fatalf("found zero raise sites for %q", code)
	}
	return windows
}

// shadowPathRefusalBaseline is Task 3(c)'s LITERAL PRE-DELETION REFUSAL
// BASELINE: for every synthetic two-arm program the generator below
// produces (armA built by generatedOwnershipBody at the given length and
// encoding under byteTypeFact, armB empty) that the shadow path currently
// REFUSES with one of the two timing-dependent codes, this table records
// (identity, code). Generated from the tree at authoring time; committed as
// literal data, not regenerated. Plan 09-09 must re-run this EXACT corpus
// after the deletion and reproduce every one of these refusals -- now
// decided by the surviving single post-assembly pass -- or report a
// regression as a blocker, never silently accommodate one.
var shadowPathRefusalBaseline = map[string]string{
	"len=2:encoded=198":  "ownership.move_while_borrowed",
	"len=2:encoded=200":  "ownership.move_while_borrowed",
	"len=2:encoded=201":  "ownership.move_while_borrowed",
	"len=2:encoded=203":  "ownership.move_while_borrowed",
	"len=2:encoded=222":  "ownership.move_while_borrowed",
	"len=2:encoded=224":  "ownership.move_while_borrowed",
	"len=2:encoded=225":  "ownership.move_while_borrowed",
	"len=2:encoded=227":  "ownership.move_while_borrowed",
	"len=2:encoded=1926": "ownership.move_while_borrowed",
	"len=2:encoded=1928": "ownership.move_while_borrowed",
	"len=2:encoded=1929": "ownership.move_while_borrowed",
	"len=2:encoded=1931": "ownership.move_while_borrowed",
	"len=2:encoded=1950": "ownership.move_while_borrowed",
	"len=2:encoded=1952": "ownership.move_while_borrowed",
	"len=2:encoded=1953": "ownership.move_while_borrowed",
	"len=2:encoded=1955": "ownership.move_while_borrowed",
	"len=2:encoded=345":  "ownership.borrow_conflict",
	"len=2:encoded=347":  "ownership.borrow_conflict",
	"len=2:encoded=369":  "ownership.borrow_conflict",
	"len=2:encoded=371":  "ownership.borrow_conflict",
	"len=2:encoded=486":  "ownership.borrow_conflict",
	"len=2:encoded=488":  "ownership.borrow_conflict",
	"len=2:encoded=489":  "ownership.borrow_conflict",
	"len=2:encoded=491":  "ownership.borrow_conflict",
	"len=2:encoded=510":  "ownership.borrow_conflict",
	"len=2:encoded=512":  "ownership.borrow_conflict",
	"len=2:encoded=513":  "ownership.borrow_conflict",
	"len=2:encoded=515":  "ownership.borrow_conflict",
	"len=2:encoded=2073": "ownership.borrow_conflict",
	"len=2:encoded=2075": "ownership.borrow_conflict",
	"len=2:encoded=2097": "ownership.borrow_conflict",
	"len=2:encoded=2099": "ownership.borrow_conflict",
	"len=2:encoded=2214": "ownership.borrow_conflict",
	"len=2:encoded=2216": "ownership.borrow_conflict",
	"len=2:encoded=2217": "ownership.borrow_conflict",
	"len=2:encoded=2219": "ownership.borrow_conflict",
	"len=2:encoded=2238": "ownership.borrow_conflict",
	"len=2:encoded=2240": "ownership.borrow_conflict",
	"len=2:encoded=2241": "ownership.borrow_conflict",
	"len=2:encoded=2243": "ownership.borrow_conflict",
}

// shadowPathAlphabet/shadowPathGeneratorLength match generatedOwnershipBody's
// own 48-symbol alphabet and the depth swept to build shadowPathRefusalBaseline.
const (
	shadowPathAlphabet       = 48
	shadowPathGeneratorDepth = 2
)

// TestShadowPathSubsumptionCorpus is Task 3's full gate: it declares the
// two reachability sets, checks their union and disjointness, verifies by
// source scan that the classification matches the ACTUAL guard structure
// (not merely D-09-47's prediction), demonstrates by a seeded-fault
// perturbation that the two included codes' verdicts genuinely depend on
// the shadow timing index, and reproduces the literal pre-deletion refusal
// baseline.
func TestShadowPathSubsumptionCorpus(t *testing.T) {
	allFiveCodes := []string{
		"ownership.use_after_move",
		"ownership.move_while_borrowed",
		"ownership.borrow_conflict",
		"ownership.borrow_requires_share",
		"ownership.transfer_requires_take",
	}

	t.Run("declared sets partition exactly the five ownership.* codes", func(t *testing.T) {
		union := map[string]bool{}
		for code := range shadowPathReachableCodes {
			if _, doubled := shadowPathUnreachableCodes[code]; doubled {
				t.Fatalf("%s is declared in BOTH sets", code)
			}
			union[code] = true
		}
		for code := range shadowPathUnreachableCodes {
			union[code] = true
		}
		if len(union) != len(allFiveCodes) {
			t.Fatalf("declared sets have %d total entries, want exactly %d (the five ownership.* codes): %v", len(union), len(allFiveCodes), union)
		}
		for _, code := range allFiveCodes {
			if !union[code] {
				t.Fatalf("%s is not accounted for in either declared set", code)
			}
		}
	})

	t.Run("structural: reachable codes' guards actually consult lastUseIndexByLoan", func(t *testing.T) {
		for code := range shadowPathReachableCodes {
			for _, window := range reachableGuardWindows(t, code) {
				if !strings.Contains(window, "lastUseIndexByLoan") {
					t.Fatalf("%s: guarding window does not consult lastUseIndexByLoan, contradicting its shadowPathReachableCodes classification:\n%s", code, window)
				}
			}
		}
	})

	t.Run("perturbation: neutralizing the shadow timing index changes a timing-dependent verdict", func(t *testing.T) {
		// testOnlyForceUniformLoanJoin (check.go) is an EXISTING production
		// fault-injection seam (03-03-02), not one this plan adds: forcing
		// every loan's computed last use to the arm's own join point is
		// exactly "as if the shadow path's own timing answer had been
		// deleted and every loan ended uniformly at the join instead" (see
		// its own doc comment). If move_while_borrowed's verdict is
		// genuinely decided by that timing answer, flipping the seam must
		// change SOME swept program's verdict; if it never does, the
		// "reachable" classification above would be asserted, not proven.
		defer func() { testOnlyForceUniformLoanJoin = false }()
		emptyArm := generatedOwnershipBody(0, 0, shadowPathAlphabet)
		flipped := false
		for length := 0; length <= shadowPathGeneratorDepth && !flipped; length++ {
			cases := 1
			for i := 0; i < length; i++ {
				cases *= shadowPathAlphabet
			}
			for encoded := 0; encoded < cases; encoded++ {
				bodyA := generatedOwnershipBody(encoded, length, shadowPathAlphabet)
				program := branchSequenceProgram(bodyA, emptyArm)

				testOnlyForceUniformLoanJoin = false
				normal := Program(program)
				normalHasMWB := diagnosticsContainCode(normal.Diagnostics, "ownership.move_while_borrowed")

				testOnlyForceUniformLoanJoin = true
				perturbed := Program(program)
				perturbedHasMWB := diagnosticsContainCode(perturbed.Diagnostics, "ownership.move_while_borrowed")
				testOnlyForceUniformLoanJoin = false

				if normalHasMWB != perturbedHasMWB {
					flipped = true
					break
				}
			}
		}
		if !flipped {
			t.Fatalf("neutralizing the shadow timing index (testOnlyForceUniformLoanJoin) never changed ownership.move_while_borrowed's verdict anywhere in the swept corpus -- its classification as shadow-path-reachable is unproven")
		}
	})

	t.Run("pre-deletion refusal baseline reproduces", func(t *testing.T) {
		// Plan 09-09 (D-09-08's authorized deletion): re-run at the MOVED
		// contract -- production evaluated through
		// straightLineSupportAtDecisionPoint (lowering, then the extended
		// post-assembly pass), never analyzeStraightLine directly, which no
		// longer decides either timing-dependent code (D-09-09).
		byteFact := byteTypeFact()
		seen := map[string]bool{}
		for length := 0; length <= shadowPathGeneratorDepth; length++ {
			cases := 1
			for i := 0; i < length; i++ {
				cases *= shadowPathAlphabet
			}
			for encoded := 0; encoded < cases; encoded++ {
				body := generatedOwnershipBody(encoded, length, shadowPathAlphabet)
				got := straightLineSupportAtDecisionPoint("test:fn", "owner", diagnostic.Span{Start: 1, End: 6}, byteFact, &body, nil, nil)
				identity := fmt.Sprintf("len=%d:encoded=%d", length, encoded)
				want, tracked := shadowPathRefusalBaseline[identity]
				if !tracked {
					if got.DiagnosticCode == "ownership.move_while_borrowed" || got.DiagnosticCode == "ownership.borrow_conflict" {
						t.Fatalf("%s: newly refused with %q, not present in the pre-deletion baseline -- the baseline must be updated with a written justification, never silently", identity, got.DiagnosticCode)
					}
					continue
				}
				seen[identity] = true
				if got.DiagnosticCode != want {
					t.Fatalf("%s: baseline mismatch: want %q, got %q -- a coverage-losing difference here is a blocker, report it rather than accommodating it", identity, want, got.DiagnosticCode)
				}
			}
		}
		for identity := range shadowPathRefusalBaseline {
			if !seen[identity] {
				t.Fatalf("baseline entry %s was not produced by the re-walked generator -- the swept depth/alphabet shrank or the entry is stale", identity)
			}
		}
	})
}

func diagnosticsContainCode(diagnostics []diagnostic.Diagnostic, code string) bool {
	for _, diag := range diagnostics {
		if diag.Code == code {
			return true
		}
	}
	return false
}
