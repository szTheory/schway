package originvalidate_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/originvalidate"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// TestOriginValidateImportsNeitherCheckNorAst is 03-06-01's structural
// falsifier for the plan's binding prohibition: originvalidate decides from
// the typed-core artifact alone and must never import the checker or the
// AST package. Reading the actual import lists (not trusting a doc comment)
// is the same technique the codebase already applies to enforce boundaries
// mechanically rather than by convention.
// TestOriginValidatorImportsStayIndependent is Task 04-06-02's extension of
// TestOriginValidateImportsNeitherCheckNorAst (T-04-37): the new
// checkForeignOriginOmitted path (D-04-28) must derive its refusal from the
// core artifact alone, exactly like every other check in this package --
// this re-runs the identical file-scan so the new function is covered by
// construction rather than by a second, drifting assertion.
func TestOriginValidatorImportsStayIndependent(t *testing.T) {
	dir := testsupport.ProjectPath("internal", "compiler", "originvalidate")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	fileSet := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fileSet, filepath.Join(dir, entry.Name()), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imported := range file.Imports {
			path := strings.Trim(imported.Path.Value, `"`)
			if strings.HasSuffix(path, "/compiler/check") || strings.HasSuffix(path, "/compiler/ast") {
				t.Fatalf("%s imports %s, which originvalidate (including the foreign-origin path) must never depend on", entry.Name(), path)
			}
		}
	}
}

func TestOriginValidateImportsNeitherCheckNorAst(t *testing.T) {
	dir := testsupport.ProjectPath("internal", "compiler", "originvalidate")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	fileSet := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fileSet, filepath.Join(dir, entry.Name()), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imported := range file.Imports {
			path := strings.Trim(imported.Path.Value, `"`)
			if strings.HasSuffix(path, "/compiler/check") || strings.HasSuffix(path, "/compiler/ast") {
				t.Fatalf("%s imports %s, which originvalidate must never depend on", entry.Name(), path)
			}
		}
	}
}

func honestProgram(t testing.TB, fixture string) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase3", fixture))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
	}
	if checked.Program.Functions[0].PublicOrigin == nil {
		t.Fatalf("expected PublicOrigin on the honestly-checked function")
	}
	return checked.Program
}

// TestOriginUnderstatedRejected is 03-06-02's falsifier for the "omitted"
// defect: a declared origin set that omits a path the body actually derives
// from is rejected as core.origin_understated, detected purely by
// source-blind recomputation from the typed core (never by trusting the
// declaration). The dishonest declaration is injected by mutating an
// honestly-checked program's own fact — check.go's honest producer can never
// construct this shape itself, exactly as OV-02-01's mutation-kill precedent
// establishes for a different fact.
func TestOriginUnderstatedRejected(t *testing.T) {
	program := honestProgram(t, "public_view_understated.lang")
	program.Functions[0].PublicOrigin = &core.PublicOrigin{Paths: []string{}, Access: "shared"}
	problems := originvalidate.ValidatePublished(program)
	if len(problems) != 1 || problems[0].Code != "core.origin_understated" {
		t.Fatalf("expected exactly core.origin_understated, got %+v", problems)
	}
}

// TestOriginAccessMismatchRejected is 03-06-02's falsifier for the
// "impossible" defect: a declared access mode the body cannot produce is
// rejected as core.origin_access_mismatch by the same recomputation.
func TestOriginAccessMismatchRejected(t *testing.T) {
	program := honestProgram(t, "public_view_impossible.lang")
	program.Functions[0].PublicOrigin = &core.PublicOrigin{Paths: []string{"buffer"}, Access: "exclusive"}
	problems := originvalidate.ValidatePublished(program)
	if len(problems) != 1 || problems[0].Code != "core.origin_access_mismatch" {
		t.Fatalf("expected exactly core.origin_access_mismatch, got %+v", problems)
	}
}

// TestMixedAccessChainDerivesShared is 03-08-01's falsifier for CR-01: a
// reborrow chain whose closest-to-return hop is a shared reborrow of an
// exclusively-borrowed place must derive access "shared" — the hop nearest
// the returned place decides, not an earlier hop further up the chain. This
// is the honest recomputation (no mutation): check.go's own producer
// constructs the mismatching *declaration* on this fixture, but
// RecomputeOrigin's own body-derived answer must still be correct in
// isolation.
func TestMixedAccessChainDerivesShared(t *testing.T) {
	program := honestProgram(t, "public_view_mixed_access.lang")
	paths, access, ok := originvalidate.RecomputeOrigin(program.Functions[0])
	if !ok {
		t.Fatalf("expected RecomputeOrigin to succeed")
	}
	if access != "shared" {
		t.Fatalf("expected access shared (closest-to-return hop), got %q", access)
	}
	if len(paths) != 1 || paths[0] != "buffer" {
		t.Fatalf("expected origin path [buffer], got %+v", paths)
	}
}

// TestMixedAccessChainRejectedAsAccessMismatch is 03-08-01's falsifier for
// CR-01's downstream effect: ValidatePublished must reject the fixture's
// declared borrow mut(buffer) (exclusive) against a body that only ever
// derives shared, naming both modes in the detail. Unlike
// TestOriginAccessMismatchRejected, no mutation is needed here — check.go's
// honest producer already constructs this exact mismatching declaration
// from the source's own `borrow mut(buffer)` return-type annotation.
//
// The reverse hop ordering (an exclusive reborrow of a shared loan) is
// exercised here too: it must be derived correctly (shared decides nothing
// once overridden by a closer exclusive hop) or rejected with a named code —
// never silently accepted as matching a declaration it does not support.
func TestMixedAccessChainRejectedAsAccessMismatch(t *testing.T) {
	program := honestProgram(t, "public_view_mixed_access.lang")
	problems := originvalidate.ValidatePublished(program)
	if len(problems) != 1 || problems[0].Code != "core.origin_access_mismatch" {
		t.Fatalf("expected exactly core.origin_access_mismatch, got %+v", problems)
	}
	if !strings.Contains(problems[0].Detail, `"exclusive"`) || !strings.Contains(problems[0].Detail, `"shared"`) {
		t.Fatalf("expected detail to name both declared and body-derived modes, got %q", problems[0].Detail)
	}

	// Reverse ordering: an exclusive reborrow of a shared loan. The closest
	// hop to the return is exclusive, so the recomputed answer must be
	// exclusive — matching a correctly-declared borrow mut(buffer) — proving
	// the guard is symmetric rather than one-sided.
	reverseSource := `module owned.public_view_mixed_access_reverse

export {
  fn view
}

fn view(buffer: Buffer) -> borrow mut(buffer) Buffer {
  let shared = borrow buffer
  let exclusive = borrow mut shared
  exclusive
}
`
	checkedReverse := session.Check([]byte(reverseSource))
	if len(checkedReverse.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics on reverse-ordering fixture: %+v", checkedReverse.Diagnostics)
	}
	if paths, access, ok := originvalidate.RecomputeOrigin(checkedReverse.Program.Functions[0]); !ok || access != "exclusive" || len(paths) != 1 || paths[0] != "buffer" {
		t.Fatalf("expected reverse-ordering chain to derive exclusive, got paths=%+v access=%q ok=%v", paths, access, ok)
	}
	reverseProblems := originvalidate.ValidatePublished(checkedReverse.Program)
	if len(reverseProblems) != 0 {
		t.Fatalf("expected the correctly-declared reverse-ordering fixture to have no problems, got %+v", reverseProblems)
	}
}

// TestOmittedOriginRejected is 03-09-01's falsifier for D-03-02/GAP 2: a
// function with NO declared origin, whose body's only return is a
// borrow-derived place, must be refused publication with
// core.origin_omitted — the category ValidatePublished's old
// PublicOrigin == nil short-circuit made definitionally unreachable.
func TestOmittedOriginRejected(t *testing.T) {
	program := honestOmittedProgram(t, "public_view_omitted.lang")
	problems := originvalidate.ValidatePublished(program)
	if len(problems) != 1 || problems[0].Code != "core.origin_omitted" {
		t.Fatalf("expected exactly core.origin_omitted, got %+v", problems)
	}
	if !strings.Contains(problems[0].Detail, "buffer") || !strings.Contains(problems[0].Detail, `"exclusive"`) {
		t.Fatalf("expected detail to name the derived paths and access, got %q", problems[0].Detail)
	}
}

// TestOwnedReturnWithNoOriginStillPublishes is 03-09-01's falsifier for the
// non-over-firing half of the same behavior: a function with no declared
// origin whose return is NOT borrow-derived (RecomputeOrigin reports
// not-ok) must still publish with no problems, whether the function is
// straight-line (shared_shared_accept.lang, an owned take/return) or
// match-bodied (borrowed_view.lang, a branch function whose every arm
// returns an owned take).
func TestOwnedReturnWithNoOriginStillPublishes(t *testing.T) {
	for _, fixture := range []string{"shared_shared_accept.lang", "borrowed_view.lang"} {
		program := honestOmittedProgram(t, fixture)
		if problems := originvalidate.ValidatePublished(program); len(problems) != 0 {
			t.Fatalf("%s: expected no problems for an owned return with no declared origin, got %+v", fixture, problems)
		}
	}
}

// TestExclusiveBorrowCleanShapeChecksButCannotPublish is 03-09-01's
// falsifier for the fixture_disposition: session.Check on the exact
// exclusive_borrow_clean / relay source (embedded verbatim from
// check_exclusive_test.go, so the disposition is machine-checked rather
// than asserted in prose) still returns zero diagnostics, while
// ValidatePublished on that same checked program now returns
// core.origin_omitted — the gate lives on the publication path only.
func TestExclusiveBorrowCleanShapeChecksButCannotPublish(t *testing.T) {
	const cleanSource = `module owned.exclusive_borrow_clean

export {
  fn relay
}

fn relay(buffer: Buffer) -> Buffer {
  let view = borrow mut buffer
  let reviewed = borrow view
  view
}
`
	checked := session.Check([]byte(cleanSource))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("expected session.Check to still accept the exclusive_borrow_clean shape unchanged, got %+v", checked.Diagnostics)
	}
	problems := originvalidate.ValidatePublished(checked.Program)
	if len(problems) != 1 || problems[0].Code != "core.origin_omitted" {
		t.Fatalf("expected exactly core.origin_omitted on the publication path, got %+v", problems)
	}
}

// honestOmittedProgram is like honestProgram but does NOT assert a non-nil
// PublicOrigin — it is used for fixtures that deliberately declare none
// (public_view_omitted.lang) or whose shape never carries one
// (match-bodied borrowed_view.lang).
func honestOmittedProgram(t testing.TB, fixture string) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase3", fixture))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
	}
	return checked.Program
}

// TestPerReturnOriginsCoverEveryArm is Task 03-10-01's falsifier for the
// core defect: RecomputeOriginPerReturn must return one element per
// core.OpReturn, not just the first. A straight-line function collapses to
// exactly one element; the multi-arm fixture collapses to exactly two, one
// per arm, and only the second (the borrow-returning arm) is borrow-derived.
func TestPerReturnOriginsCoverEveryArm(t *testing.T) {
	straightLine := honestProgram(t, "public_view.lang")
	straightOrigins := originvalidate.RecomputeOriginPerReturn(straightLine.Functions[0])
	if len(straightOrigins) != 1 {
		t.Fatalf("expected exactly 1 per-return origin for a straight-line function, got %+v", straightOrigins)
	}
	if !straightOrigins[0].Derived || straightOrigins[0].Access != "shared" {
		t.Fatalf("expected the straight-line origin to be shared-derived, got %+v", straightOrigins[0])
	}

	program := honestOmittedProgram(t, "public_view_multi_arm_omitted.lang")
	origins := originvalidate.RecomputeOriginPerReturn(program.Functions[0])
	if len(origins) != 2 {
		t.Fatalf("expected exactly 2 per-return origins (one per arm), got %+v", origins)
	}
	derivedCount := 0
	for _, origin := range origins {
		if origin.Derived {
			derivedCount++
			if origin.Access != "shared" {
				t.Fatalf("expected the borrow-derived arm's access to be shared, got %+v", origin)
			}
		}
	}
	if derivedCount != 1 {
		t.Fatalf("expected exactly 1 borrow-derived return among the 2 arms, got %d in %+v", derivedCount, origins)
	}
}

// TestMultiArmOmittedOriginRejected is Task 03-10-01's falsifier for
// SC3/SC4: a match-bodied function whose first arm returns owned and second
// arm returns a live borrow, with no declared origin (match functions cannot
// declare one), must be refused publication with core.origin_omitted — not
// silently accepted because RecomputeOrigin only looked at the first arm.
func TestMultiArmOmittedOriginRejected(t *testing.T) {
	program := honestOmittedProgram(t, "public_view_multi_arm_omitted.lang")
	paths, access, ok := originvalidate.RecomputeOrigin(program.Functions[0])
	if !ok {
		t.Fatalf("expected RecomputeOrigin to succeed on the multi-arm leak")
	}
	if access != "shared" {
		t.Fatalf("expected combined access shared, got %q", access)
	}
	if len(paths) != 1 || paths[0] != "flag" {
		t.Fatalf("expected origin path [flag], got %+v", paths)
	}
	problems := originvalidate.ValidatePublished(program)
	if len(problems) != 1 || problems[0].Code != "core.origin_omitted" {
		t.Fatalf("expected exactly core.origin_omitted, got %+v", problems)
	}
}

// TestOwnedArmsStillPublish is Task 03-10-01's non-over-firing falsifier:
// the widened, every-OpReturn walk must not newly refuse a match-bodied
// function whose every arm returns an owned value.
func TestOwnedArmsStillPublish(t *testing.T) {
	for _, fixture := range []string{"branch_view.lang", "borrowed_view.lang"} {
		program := honestOmittedProgram(t, fixture)
		if problems := originvalidate.ValidatePublished(program); len(problems) != 0 {
			t.Fatalf("%s: expected no problems for every-arm-owned, got %+v", fixture, problems)
		}
	}
}

// TestMultiArmAccessConflictDerivesNeitherArm is Task 03-10-02's falsifier
// for the combination law's case 3: two arms deriving different access
// modes must combine to the AccessConflicting sentinel, with paths unioned
// (not duplicated) — neither arm's own answer wins.
func TestMultiArmAccessConflictDerivesNeitherArm(t *testing.T) {
	program := honestOmittedProgram(t, "public_view_multi_arm_access_conflict.lang")
	origins := originvalidate.RecomputeOriginPerReturn(program.Functions[0])
	if len(origins) != 2 {
		t.Fatalf("expected exactly 2 per-return origins, got %+v", origins)
	}
	seenShared, seenExclusive := false, false
	for _, origin := range origins {
		if !origin.Derived {
			t.Fatalf("expected both arms borrow-derived, got %+v", origin)
		}
		switch origin.Access {
		case "shared":
			seenShared = true
		case "exclusive":
			seenExclusive = true
		}
	}
	if !seenShared || !seenExclusive {
		t.Fatalf("expected one shared and one exclusive arm, got %+v", origins)
	}
	paths, access, ok := originvalidate.RecomputeOrigin(program.Functions[0])
	if !ok || access != originvalidate.AccessConflicting {
		t.Fatalf("expected AccessConflicting, got access=%q ok=%v", access, ok)
	}
	if len(paths) != 1 || paths[0] != "flag" {
		t.Fatalf("expected exactly one unioned path [flag], got %+v", paths)
	}
}

// TestUnionPathsAreNotDuplicated pins the same non-duplication requirement
// directly, independent of the access-mode assertions above.
func TestUnionPathsAreNotDuplicated(t *testing.T) {
	program := honestOmittedProgram(t, "public_view_multi_arm_access_conflict.lang")
	paths, _, ok := originvalidate.RecomputeOrigin(program.Functions[0])
	if !ok {
		t.Fatalf("expected RecomputeOrigin to succeed")
	}
	seen := make(map[string]int)
	for _, path := range paths {
		seen[path]++
	}
	for path, count := range seen {
		if count != 1 {
			t.Fatalf("path %q duplicated %d times in %+v", path, count, paths)
		}
	}
}

// TestMultiArmAccessConflictRejectedWhenDeclaredShared is Task 03-10-02's
// falsifier: an undeclared conflicting-arms function is refused with
// core.origin_omitted (the same code single-arm omission uses); a
// conflicting-arms function DECLARING "shared" is refused with
// core.origin_access_mismatch, whose Detail names the sentinel.
func TestMultiArmAccessConflictRejectedWhenDeclaredShared(t *testing.T) {
	undeclared := honestOmittedProgram(t, "public_view_multi_arm_access_conflict.lang")
	problems := originvalidate.ValidatePublished(undeclared)
	if len(problems) != 1 || problems[0].Code != "core.origin_omitted" {
		t.Fatalf("expected exactly core.origin_omitted for the undeclared conflicting-arms function, got %+v", problems)
	}

	declaredShared := honestOmittedProgram(t, "public_view_multi_arm_access_conflict.lang")
	declaredShared.Functions[0].PublicOrigin = &core.PublicOrigin{Paths: []string{"flag"}, Access: "shared"}
	problems = originvalidate.ValidatePublished(declaredShared)
	if len(problems) != 1 || problems[0].Code != "core.origin_access_mismatch" {
		t.Fatalf("expected exactly core.origin_access_mismatch, got %+v", problems)
	}
	if !strings.Contains(problems[0].Detail, originvalidate.AccessConflicting) {
		t.Fatalf("expected detail to name the conflicting sentinel, got %q", problems[0].Detail)
	}
}

// TestMultiArmAccessConflictRejectedWhenDeclaredExclusive mirrors the shared
// case with the other declarable access mode.
func TestMultiArmAccessConflictRejectedWhenDeclaredExclusive(t *testing.T) {
	declaredExclusive := honestOmittedProgram(t, "public_view_multi_arm_access_conflict.lang")
	declaredExclusive.Functions[0].PublicOrigin = &core.PublicOrigin{Paths: []string{"flag"}, Access: "exclusive"}
	problems := originvalidate.ValidatePublished(declaredExclusive)
	if len(problems) != 1 || problems[0].Code != "core.origin_access_mismatch" {
		t.Fatalf("expected exactly core.origin_access_mismatch, got %+v", problems)
	}
	if !strings.Contains(problems[0].Detail, originvalidate.AccessConflicting) {
		t.Fatalf("expected detail to name the conflicting sentinel, got %q", problems[0].Detail)
	}
}

// TestDeclaredConflictingAccessIsRefused is Task 03-10-02's falsifier for
// the declared-access domain check: a declared PublicOrigin.Access equal to
// the sentinel string itself is refused with core.origin_access_mismatch
// BEFORE any comparison with the recomputed answer — the sentinel is never
// a declarable mode, so a mutated summary cannot declare it and match a
// conflicting recomputation.
func TestDeclaredConflictingAccessIsRefused(t *testing.T) {
	// public_view.lang's recomputed access is "shared", not the sentinel, so
	// declaring the sentinel there is caught by the ordinary
	// declared-vs-recomputed mismatch comparison regardless of whether a
	// dedicated domain check exists. To actually falsify the domain check,
	// declare the sentinel on the ONE fixture whose own recomputed answer IS
	// the sentinel (public_view_multi_arm_access_conflict.lang) — without a
	// domain check running BEFORE the comparison, declared == recomputed and
	// this would incorrectly report no problems at all.
	conflicting := honestOmittedProgram(t, "public_view_multi_arm_access_conflict.lang")
	conflicting.Functions[0].PublicOrigin = &core.PublicOrigin{Paths: []string{"flag"}, Access: originvalidate.AccessConflicting}
	problems := originvalidate.ValidatePublished(conflicting)
	if len(problems) != 1 || problems[0].Code != "core.origin_access_mismatch" {
		t.Fatalf("expected exactly core.origin_access_mismatch for a declared sentinel access, got %+v", problems)
	}

	program := honestProgram(t, "public_view.lang")
	program.Functions[0].PublicOrigin = &core.PublicOrigin{Paths: []string{"buffer"}, Access: originvalidate.AccessConflicting}
	problems = originvalidate.ValidatePublished(program)
	if len(problems) != 1 || problems[0].Code != "core.origin_access_mismatch" {
		t.Fatalf("expected exactly core.origin_access_mismatch for a declared sentinel access, got %+v", problems)
	}
}

// TestPublishedOriginConsistentWithEveryReturn is Task 03-10-02's
// generalized every-return invariant (the tripwire against a future
// regression back to a single-return assumption): for every function in
// every testdata/phase3 fixture, RecomputeOrigin's combined triple must be
// exactly the conservative combination of RecomputeOriginPerReturn's
// elements.
func TestPublishedOriginConsistentWithEveryReturn(t *testing.T) {
	dir := testsupport.ProjectPath("testdata", "phase3")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	checkedAny := false
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".lang") {
			continue
		}
		source, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		checked := session.Check(source)
		if len(checked.Diagnostics) != 0 {
			continue
		}
		for _, function := range checked.Program.Functions {
			checkedAny = true
			perReturn := originvalidate.RecomputeOriginPerReturn(function)
			var wantPaths []string
			pathSeen := make(map[string]bool)
			wantAccess := ""
			derivedCount := 0
			conflict := false
			for _, origin := range perReturn {
				if !origin.Derived {
					continue
				}
				derivedCount++
				if wantAccess == "" {
					wantAccess = origin.Access
				} else if origin.Access != wantAccess {
					conflict = true
				}
				for _, path := range origin.Paths {
					if !pathSeen[path] {
						pathSeen[path] = true
						wantPaths = append(wantPaths, path)
					}
				}
			}
			gotPaths, gotAccess, gotOK := originvalidate.RecomputeOrigin(function)
			if derivedCount == 0 {
				if gotOK {
					t.Fatalf("%s/%s: expected not-ok for an owned function, got paths=%+v access=%q", entry.Name(), function.ID, gotPaths, gotAccess)
				}
				continue
			}
			if !gotOK {
				t.Fatalf("%s/%s: expected ok=true for a borrow-derived function, got not-ok", entry.Name(), function.ID)
			}
			expectedAccess := wantAccess
			if conflict {
				expectedAccess = originvalidate.AccessConflicting
			}
			if gotAccess != expectedAccess {
				t.Fatalf("%s/%s: expected combined access %q, got %q", entry.Name(), function.ID, expectedAccess, gotAccess)
			}
			if len(gotPaths) != len(wantPaths) {
				t.Fatalf("%s/%s: expected paths %+v, got %+v", entry.Name(), function.ID, wantPaths, gotPaths)
			}
			for index, path := range wantPaths {
				if gotPaths[index] != path {
					t.Fatalf("%s/%s: expected paths %+v, got %+v", entry.Name(), function.ID, wantPaths, gotPaths)
				}
			}
		}
	}
	if !checkedAny {
		t.Fatal("expected at least one function from testdata/phase3 to be checked")
	}
}

// TestStaleSummaryRejectedBeforeOtherChecks is 03-06-02's falsifier for
// T-03-03: a summary whose recorded digest does not match the core artifact
// it is checked against is rejected before any origin or access question is
// even asked — CheckSummary never unmarshals coreBytes into a struct that
// could carry a body, so the ordering is structural, not merely sequenced.
func TestStaleSummaryRejectedBeforeOtherChecks(t *testing.T) {
	program := honestProgram(t, "public_view.lang")
	summary, err := originvalidate.BuildInterface(program)
	if err != nil {
		t.Fatal(err)
	}
	// 07-01 Task 3 fills ClosureDigest for real; here it only needs to pass
	// DecodeInterface's shape check so CheckSummary's own staleness check
	// (the thing this test asserts) is what actually rejects the document.
	summary.Functions[0].ClosureDigest = validClosureDigestPlaceholder
	summaryBytes, err := json.Marshal(summary)
	if err != nil {
		t.Fatal(err)
	}
	staleCore := []byte(`{"not":"the real core artifact"}`)
	if _, err := originvalidate.CheckSummary(summaryBytes, staleCore); err == nil {
		t.Fatal("expected a stale-summary rejection")
	} else if code := errorCode(err); code != "origin.stale_summary" {
		t.Fatalf("expected origin.stale_summary, got %q (%v)", code, err)
	}

	// A summary whose declared origin would itself be dishonest must still be
	// rejected for staleness first, proving the digest check runs before any
	// origin-shaped decision — CheckSummary has no other check to reorder
	// against, which is itself the point: there IS no origin/access check
	// left to run once the digest fails.
	dishonest := summary
	dishonest.Functions[0].Return.Mode = "exclusive"
	dishonest.Functions[0].Return.Paths = []string{"nonexistent"}
	dishonestBytes, err := json.Marshal(dishonest)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := originvalidate.CheckSummary(dishonestBytes, staleCore); errorCode(err) != "origin.stale_summary" {
		t.Fatalf("expected origin.stale_summary ahead of any origin content, got %v", err)
	}
}

// TestOriginEscapeIsNamed is 03-06-02's falsifier naming the accepted
// TestCheckSummaryRoutesThroughDecodeInterface is 07-01 Task 2's Test 5
// (D-07-36): a /1 document missing Mode is refused when passed to
// CheckSummary, asserted through CheckSummary itself, not only through
// core.DecodeInterface directly -- this is the exact gap codex found
// ("CheckSummary unmarshals straight into core.Interface and checks only
// JSON validity and CoreDigest").
func TestCheckSummaryRoutesThroughDecodeInterface(t *testing.T) {
	missingModeDoc := []byte(`{
		"schema": "lang.interface/1",
		"module_id": "m",
		"core_digest": "` + validClosureDigestPlaceholder + `",
		"functions": [{
			"id": "f", "name": "f",
			"parameters": [{"id":"p","name":"p","type":"Byte","mode":"","drops":false}],
			"return": {"type":"Byte","mode":"owned","paths":[],"fresh":false},
			"abilities": [], "callable": false,
			"foreign": {"allocator":"","unwind":"","nonlocal_exit":""},
			"closure_digest": "` + validClosureDigestPlaceholder + `"
		}]
	}`)
	if _, err := originvalidate.CheckSummary(missingModeDoc, []byte(`{}`)); err == nil {
		t.Fatal("expected CheckSummary to refuse a /1 document with a missing Mode")
	} else if code := errorCode(err); code != "core.interface_missing_field" {
		t.Fatalf("expected core.interface_missing_field, got %q (%v)", code, err)
	}
}

// TestCheckSummaryRefusesV0Document is D-07-36/T-07-02's CheckSummary-level
// falsifier: a lang.interface/0 document is decodable but never admissible
// for a call, so CheckSummary must refuse it rather than silently answering
// origin questions from a frozen legacy shape it was never validated
// against.
func TestCheckSummaryRefusesV0Document(t *testing.T) {
	if _, err := originvalidate.CheckSummary([]byte(pinnedV0DocumentForCheckSummary), []byte(`{}`)); err == nil {
		t.Fatal("expected CheckSummary to refuse a lang.interface/0 document")
	} else if code := errorCode(err); code != "origin.summary_not_admissible" {
		t.Fatalf("expected origin.summary_not_admissible, got %q (%v)", code, err)
	}
}

const pinnedV0DocumentForCheckSummary = `{"schema":"lang.interface/0","module_id":"m1","core_digest":"` + validClosureDigestPlaceholder + `","functions":[{"id":"f1","name":"identity","parameter":{"id":"p1","name":"buffer","type":"Buffer"},"return_type":"Buffer","abilities":[]}]}`

// residual: a coordinated frontend-and-summary lie is declared as a named
// expected escape, never solved and never silently absent.
func TestOriginEscapeIsNamed(t *testing.T) {
	if originvalidate.KnownEscape == "" {
		t.Fatal("KnownEscape must be a non-empty named constant")
	}
	escapes := originvalidate.ExpectedEscapes()
	found := false
	for _, escape := range escapes {
		if escape == originvalidate.KnownEscape {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected KnownEscape among ExpectedEscapes, got %+v", escapes)
	}
}

// TestInterfaceSummaryOmitsBodies is 03-06-01's falsifier: the exported
// interface summary contains no linear or match body — not merely an omitted
// field, but a shape (core.FunctionSignature) that structurally has no such
// field to omit.
func TestInterfaceSummaryOmitsBodies(t *testing.T) {
	program := honestProgram(t, "public_view.lang")
	summary, err := originvalidate.BuildInterface(program)
	if err != nil {
		t.Fatal(err)
	}
	if len(summary.Functions) != 1 {
		t.Fatalf("expected one function signature, got %d", len(summary.Functions))
	}
	if summary.Functions[0].Return.Mode != "shared" {
		t.Fatalf("expected the origin fact to survive stripping: %+v", summary.Functions[0])
	}
	encoded, err := json.Marshal(summary)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{`"linear"`, `"match"`, `"operations"`, `"blocks"`} {
		if bytes.Contains(encoded, []byte(forbidden)) {
			t.Fatalf("interface summary leaked a body field %s: %s", forbidden, encoded)
		}
	}
}

// validClosureDigestPlaceholder is a syntactically valid sha256:+64-hex
// shape used only to satisfy DecodeInterface's shape check in tests that
// predate 07-01 Task 3 (which computes ClosureDigest for real) — it is not
// a real content digest of anything.
const validClosureDigestPlaceholder = "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func errorCode(err error) string {
	var typed *originvalidate.Error
	if errors.As(err, &typed) {
		return typed.Code
	}
	return ""
}

// phase4Program checks a testdata/phase4 fixture (rather than phase3's)
// through session.Check and returns the resulting core.Program, failing the
// test on any diagnostic.
func phase4Program(t testing.TB, fixture string) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", fixture))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
	}
	return checked.Program
}

// TestForeignBorrowDerivedReturnRecognised is D-04-28's positive falsifier:
// a foreign call declared to borrow its argument produces a return
// RecomputeOriginPerReturn recognises as Derived (access "shared"), and a
// matching declared PublicOrigin publishes cleanly through ValidatePublished
// -- proving the widening does not merely refuse the omitted case, it also
// correctly validates the declared one. check.go's checkFallibleLinear does
// not (this phase) wire a source-level `borrow(path)` return-type
// annotation onto a foreign-tracer function, so the declared side of this
// fixture is attached directly onto the checked core.Function -- exercising
// exactly the same originvalidate entry points a real declaration would.
func TestForeignBorrowDerivedReturnRecognised(t *testing.T) {
	program := phase4Program(t, "foreign_acquire_one.lang")
	if program.Functions[0].ForeignContract == nil {
		t.Fatal("expected foreign_acquire_one.lang to carry a ForeignContract")
	}
	program.Functions[0].ForeignContract.Alias = "borrow"

	origins := originvalidate.RecomputeOriginPerReturn(program.Functions[0])
	found := false
	for _, origin := range origins {
		if origin.Derived {
			found = true
			if origin.Access != "shared" || len(origin.Paths) != 1 || origin.Paths[0] != "request" {
				t.Fatalf("expected a shared derivation from %q, got %+v", "request", origin)
			}
		}
	}
	if !found {
		t.Fatalf("expected at least one borrow-derived return, got %+v", origins)
	}

	program.Functions[0].PublicOrigin = &core.PublicOrigin{Paths: []string{"request"}, Access: "shared"}
	if problems := originvalidate.ValidatePublished(program); len(problems) != 0 {
		t.Fatalf("expected a correctly declared foreign-borrow origin to publish cleanly, got %+v", problems)
	}
}

// TestForeignOriginOmittedRejected is D-04-28's negative falsifier: the same
// shape with NO declared PublicOrigin is refused with
// core.foreign_origin_omitted, naming the offending function and the
// argument the origin derives from.
func TestForeignOriginOmittedRejected(t *testing.T) {
	program := phase4Program(t, "foreign_origin_omitted.lang")
	problems := originvalidate.ValidatePublished(program)
	if len(problems) != 1 || problems[0].Code != "core.foreign_origin_omitted" {
		t.Fatalf("expected exactly one core.foreign_origin_omitted problem, got %+v", problems)
	}
	if !strings.Contains(problems[0].Detail, "request") {
		t.Fatalf("expected the detail to name the argument the origin derives from: %+v", problems[0])
	}
}

// TestOriginWalksEveryTerminator is D-04-29's falsifier for originvalidate:
// before this widening, RecomputeOriginPerReturn's backward-walk collection
// loop recognised only core.OpReturn, so a function with a fail-only or
// defect-only path contributed nothing to the per-terminator origin
// picture. defect_terminal.lang's "Halt" arm exits ONLY through core.OpDefect
// (no OpReturn in that arm at all) and foreign_acquire_one.lang's err block
// exits ONLY through core.OpFail -- both are reachable inputs the old
// single-terminator condition would have missed entirely.
func TestOriginWalksEveryTerminator(t *testing.T) {
	defectProgram := phase4Program(t, "defect_terminal.lang")
	defectOrigins := originvalidate.RecomputeOriginPerReturn(defectProgram.Functions[0])
	if len(defectOrigins) != 2 {
		t.Fatalf("expected one origin entry per terminator (return + defect), got %d: %+v", len(defectOrigins), defectOrigins)
	}

	foreignProgram := phase4Program(t, "foreign_acquire_one.lang")
	foreignOrigins := originvalidate.RecomputeOriginPerReturn(foreignProgram.Functions[0])
	if len(foreignOrigins) != 2 {
		t.Fatalf("expected one origin entry per terminator (return + fail), got %d: %+v", len(foreignOrigins), foreignOrigins)
	}
}

// TestTerminatorSetReadFromRegistry asserts originvalidate.RecognizesTerminator
// agrees with core.TerminatorKinds() exactly -- every registered terminator
// is recognised, and a non-terminator kind (core.OpCopy) is not -- proving
// the package reads the registry rather than restating a private copy of it.
func TestTerminatorSetReadFromRegistry(t *testing.T) {
	for _, terminator := range core.TerminatorKinds() {
		if !originvalidate.RecognizesTerminator(terminator) {
			t.Fatalf("expected originvalidate to recognise registered terminator %q", terminator)
		}
	}
	if originvalidate.RecognizesTerminator(core.OpCopy) {
		t.Fatalf("expected originvalidate to NOT recognise core.OpCopy as a terminator")
	}
}

// TestTerminatorWalkMutationKilled is D-09's automated mutation-kill
// falsifier for D-04-29: narrowing the recognised terminator set (deleting
// core.OpFail, the exact mutation the throwaway-detached-worktree
// demonstration performs on the source) must make originvalidate lose the
// fail-only fixture's origin fact -- proving the widening actually bites,
// not merely that a differential stays green.
func TestTerminatorWalkMutationKilled(t *testing.T) {
	program := phase4Program(t, "foreign_acquire_one.lang")
	full := originvalidate.RecomputeOriginPerReturn(program.Functions[0])

	original := originvalidate.TerminatorKindsOverride
	originvalidate.TerminatorKindsOverride = func() []core.OperationKind {
		return []core.OperationKind{core.OpReturn, core.OpDefect} // OpFail deleted
	}
	defer func() { originvalidate.TerminatorKindsOverride = original }()
	mutated := originvalidate.RecomputeOriginPerReturn(program.Functions[0])

	if len(mutated) >= len(full) {
		t.Fatalf("mutation (deleting OpFail) had no observable effect: full=%d mutated=%d", len(full), len(mutated))
	}
}

// TestInterfaceV1FieldInvariantsAcrossCorpus is 07-01 Task 1's acceptance
// criterion: over the entire testdata/phase1..phase4 corpus, BuildInterface
// must produce, for every function, a Mode value in {owned, shared,
// exclusive} for every parameter and every return, and Return.Paths empty
// EXACTLY when Return.Mode == "owned" (D-07-09). Only fixtures that check
// cleanly (zero diagnostics) are exercised -- a rejected/malformed fixture
// never reaches BuildInterface in the real pipeline either.
func TestInterfaceV1FieldInvariantsAcrossCorpus(t *testing.T) {
	checkedAny := false
	for _, phase := range []string{"phase1", "phase2", "phase3", "phase4"} {
		dir := testsupport.ProjectPath("testdata", phase)
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read %s: %v", dir, err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".lang") {
				continue
			}
			source, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				t.Fatalf("read %s/%s: %v", phase, entry.Name(), err)
			}
			checked := session.Check(source)
			if len(checked.Diagnostics) > 0 {
				continue // rejected fixture: never reaches BuildInterface for real
			}
			summary, err := originvalidate.BuildInterface(checked.Program)
			if err != nil {
				t.Fatalf("%s/%s: BuildInterface: %v", phase, entry.Name(), err)
			}
			for _, function := range summary.Functions {
				checkedAny = true
				for _, parameter := range function.Parameters {
					if parameter.Mode != "owned" && parameter.Mode != "shared" && parameter.Mode != "exclusive" {
						t.Fatalf("%s/%s: function %s parameter %s: Mode %q outside {owned,shared,exclusive}", phase, entry.Name(), function.ID, parameter.Name, parameter.Mode)
					}
				}
				if function.Return.Mode != "owned" && function.Return.Mode != "shared" && function.Return.Mode != "exclusive" {
					t.Fatalf("%s/%s: function %s Return.Mode %q outside {owned,shared,exclusive}", phase, entry.Name(), function.ID, function.Return.Mode)
				}
				pathsEmpty := len(function.Return.Paths) == 0
				if (function.Return.Mode == "owned") != pathsEmpty {
					t.Fatalf("%s/%s: function %s Return.Paths=%v must be empty exactly when Mode==owned (Mode=%q)", phase, entry.Name(), function.ID, function.Return.Paths, function.Return.Mode)
				}
			}
		}
	}
	if !checkedAny {
		t.Fatal("expected at least one function across testdata/phase1..phase4 to be checked")
	}
}
