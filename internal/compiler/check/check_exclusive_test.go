package check_test

import (
	"os"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/syntax"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

const exclusiveBorrowSource = `module owned.exclusive_borrow_lowering

export {
  fn relay
}

fn relay(buffer: Buffer) -> Buffer {
  let view = borrow mut buffer
  let delivered = take buffer
  let observed = take view
  delivered
}
`

// TestExclusiveBorrowLowersToCore is 03-02-01's checker falsifier: checking
// `let view = borrow mut buffer` emits exactly one core.OpBorrowExclusive
// operation carrying a non-empty loan ID, gated on the same AbilityShare
// requirement shared borrows are gated on (this fixture's Buffer parameter
// grants share, so the gate is satisfied, not exercised as a rejection —
// see check.go's borrow_mut case for the D-10 note that the rejection path
// is source-unreachable, exactly like the shared case beside it). The
// existing move_while_borrowed law still fires generically for the
// exclusive loan (it is not re-derived per access mode).
//
// Plan 09-09 (D-09-08's authorized deletion): `observed = view` (an implicit
// copy) is now `observed = take view` instead. Buffer withholds Copy, so
// once lowering stops deciding ownership.move_while_borrowed inline and
// proceeds through every binding unconditionally (D-09-09), the OLD implicit
// copy of `view` would newly fail with the timing-INDEPENDENT
// ownership.transfer_requires_take BEFORE the post-assembly pass ever
// decides the timing-dependent code this test exists to exercise -- the
// precedence rule (D-09-13) working exactly as designed, just not what this
// specific test is about. `take view` preserves the loan's later reference
// (still extending its last use past the move) without requiring Copy.
func TestExclusiveBorrowLowersToCore(t *testing.T) {
	checked := session.Check([]byte(exclusiveBorrowSource))
	if len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != "ownership.move_while_borrowed" {
		t.Fatalf("expected exactly the pre-existing move-while-borrowed rejection, got %+v", checked.Diagnostics)
	}

	// A variant that never moves the owner while the exclusive loan is live
	// proves the exclusive operation itself lowers and validates cleanly.
	// 07-02 D-07-44: this exact shape is also the extracted
	// testdata/phase07/clean_but_unpublishable.lang negative control (module
	// name changed only), read here from that single source of truth rather
	// than embedded a second time.
	clean, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase07", "clean_but_unpublishable.lang"))
	if err != nil {
		t.Fatal(err)
	}
	checkedClean := session.Check(clean)
	if len(checkedClean.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checkedClean.Diagnostics)
	}
	if len(checkedClean.Program.Functions) != 1 {
		t.Fatalf("want one function, got %d", len(checkedClean.Program.Functions))
	}
	function := checkedClean.Program.Functions[0]
	var exclusive *core.LinearOperation
	for index, operation := range function.Linear.Operations {
		if operation.Kind == core.OpBorrowExclusive {
			exclusive = &function.Linear.Operations[index]
		}
	}
	if exclusive == nil {
		t.Fatalf("expected one borrow_exclusive operation among %+v", function.Linear.Operations)
	}
	if exclusive.LoanID == "" {
		t.Fatalf("exclusive borrow operation carries no loan id: %+v", exclusive)
	}
	if result := corevalidate.Validate(checkedClean.Program); !result.Valid {
		t.Fatalf("independent validator rejected a valid exclusive-borrow program: %+v", result.Problems)
	}
}

// mustCheckSource parses and checks source directly via check.Program
// (rather than session.Check) so the test can reach check.Result.AliasFacts,
// an unexported-derivation-backed field session.CheckResult does not thread
// through.
func mustCheckSource(t *testing.T, source string) check.Result {
	t.Helper()
	parsed := syntax.Parse([]byte(source))
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("fixture failed to parse: %+v", parsed.Diagnostics)
	}
	result := check.Program(parsed.Program)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("fixture failed to check: %+v", result.Diagnostics)
	}
	return result
}

func readFixtureSource(t *testing.T, path string) string {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath(strings.Split(path, "/")...))
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}

// phase1Through4AliasFactFixtures is the same accepting-fixture domain
// cgen_test.go's phase1Through4AcceptingFixtures pins (duplicated here
// rather than imported, since that list is unexported in an external _test
// package with no exported symbol for it — matching that file's own
// duplication rationale for the identical list borrowed from
// core_test.go's pinnedFixtures).
var phase1Through4AliasFactFixtures = []string{
	"testdata/phase1/comments.lang",
	"testdata/phase1/toggle.lang",
	"testdata/phase2/implicit_copy.lang",
	"testdata/phase2/owned_transfer.lang",
	"testdata/phase3/borrowed_view.lang",
	"testdata/phase3/branch_one_arm_shared_accept.lang",
	"testdata/phase3/branch_view.lang",
	"testdata/phase3/public_view.lang",
	"testdata/phase3/public_view_impossible.lang",
	"testdata/phase3/public_view_mixed_access.lang",
	"testdata/phase3/public_view_multi_arm_access_conflict.lang",
	"testdata/phase3/public_view_multi_arm_omitted.lang",
	"testdata/phase3/public_view_omitted.lang",
	"testdata/phase3/public_view_understated.lang",
	"testdata/phase3/sequential_shared_then_exclusive_accept.lang",
	"testdata/phase3/shared_shared_accept.lang",
	"testdata/phase4/acquire_three_fail_second.lang",
	"testdata/phase4/acquire_three_fail_third.lang",
	"testdata/phase4/acquire_three_success.lang",
	"testdata/phase4/defect_terminal.lang",
	"testdata/phase4/discard_because.lang",
	"testdata/phase4/foreign_acquire_one.lang",
	"testdata/phase4/nonlocal_exit_probe.lang",
}

// partialCoverageExclusiveSource is D-05-01's own load-bearing negative
// case: the exclusive loan `first` is created and referenced once more (by
// `second`), but the function's own terminator returns `buffer` directly --
// bypassing the loan chain entirely, since a borrow never moves ownership.
// The loan's own last use is therefore NOT the terminator: a partially-
// covering loan, which must never produce an alias fact.
const partialCoverageExclusiveSource = `module owned.alias_fact_partial_coverage

export {
  fn relay
}

fn relay(buffer: Buffer) -> Buffer {
  let first = borrow mut buffer
  let second = borrow first
  buffer
}
`

// sharedOnlySource never creates an exclusive loan at all -- the first
// operation is a SHARED borrow, so deriveAliasFacts must reject it outright
// regardless of whether the chain otherwise covers the whole call.
const sharedOnlySource = `module owned.alias_fact_shared_only

export {
  fn relay
}

fn relay(buffer: Buffer) -> Buffer {
  let first = borrow buffer
  let second = borrow first
  second
}
`

// TestAliasFactRequiresWholeCallExclusivity is D-05-01's own falsifier: the
// alias fact is derived ONLY when the parameter's exclusive loan covers
// every operation from its own creation to the function's terminator, never
// merely "at some point".
func TestAliasFactRequiresWholeCallExclusivity(t *testing.T) {
	t.Run("restrict_borrow fixture yields exactly one exclusive_borrow fact", func(t *testing.T) {
		source := readFixtureSource(t, "testdata/phase5/restrict_borrow.lang")
		result := mustCheckSource(t, source)
		if len(result.AliasFacts) != 1 {
			t.Fatalf("want exactly one alias fact, got %d: %+v", len(result.AliasFacts), result.AliasFacts)
		}
		if result.AliasFacts[0].Kind != check.AliasFactExclusiveBorrow {
			t.Fatalf("want Kind %q, got %q", check.AliasFactExclusiveBorrow, result.AliasFacts[0].Kind)
		}
		if result.AliasFacts[0].LoanID == "" {
			t.Fatal("alias fact carries no loan id")
		}
	})

	t.Run("a loan ending before the terminator yields zero facts", func(t *testing.T) {
		result := mustCheckSource(t, partialCoverageExclusiveSource)
		if len(result.AliasFacts) != 0 {
			t.Fatalf("want zero alias facts for a partially-covering loan, got %+v", result.AliasFacts)
		}
	})

	t.Run("a shared loan yields zero facts", func(t *testing.T) {
		result := mustCheckSource(t, sharedOnlySource)
		if len(result.AliasFacts) != 0 {
			t.Fatalf("want zero alias facts for a shared-only loan, got %+v", result.AliasFacts)
		}
	})

	t.Run("every Phase 1-4 fixture yields zero facts", func(t *testing.T) {
		for _, path := range phase1Through4AliasFactFixtures {
			t.Run(path, func(t *testing.T) {
				source := readFixtureSource(t, path)
				result := mustCheckSource(t, source)
				if len(result.AliasFacts) != 0 {
					t.Fatalf("want zero alias facts, got %+v", result.AliasFacts)
				}
			})
		}
	})
}

// TestAliasFactAgreesWithByPointerSelection proves deriveAliasFacts (exposed
// via check.Result.AliasFacts) and cgen's selectsByPointerLowering
// (exported as SelectsByPointerLowering) are the SAME condition on every
// corpus fixture: the attribute can never be emitted on a function the
// checker's own admission-gating fact does not cover, and the checker's
// fact can never exist for a function cgen would not also select for
// by-pointer lowering.
func TestAliasFactAgreesWithByPointerSelection(t *testing.T) {
	fixtures := append([]string{"testdata/phase5/restrict_borrow.lang"}, phase1Through4AliasFactFixtures...)
	for _, path := range fixtures {
		t.Run(path, func(t *testing.T) {
			source := readFixtureSource(t, path)
			result := mustCheckSource(t, source)
			factedFunctions := make(map[string]bool, len(result.AliasFacts))
			for _, fact := range result.AliasFacts {
				factedFunctions[fact.Function] = true
			}
			for _, function := range result.Program.Functions {
				selected := cgen.SelectsByPointerLowering(function, function.Linear)
				faceted := factedFunctions[function.ID]
				if selected != faceted {
					t.Fatalf("function %q: selectsByPointerLowering=%v, has alias fact=%v", function.Name, selected, faceted)
				}
			}
		})
	}
}
