package session_test

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// TestPhase5RequiredControlsMatchExportedConstants guards against literal
// spelling drift between Phase5RequiredControls()'s own flat string list
// (grep target for scripts/verify-phase5.sh's parity block) and the
// exported control-ID constants used to fire those same controls
// elsewhere in this package (ControlAliasFalseNoAlias,
// ControlDiagnosticRejectProgramIDEquivalence, the four sanitize
// controls).
func TestPhase5RequiredControlsMatchExportedConstants(t *testing.T) {
	named := map[string]bool{
		session.ControlAliasFalseNoAlias:                    true,
		session.ControlDiagnosticRejectProgramIDEquivalence: true,
		session.ControlSanitizeRetainedPointer:              true,
		session.ControlSanitizeUBSanNoRecover:               true,
		session.ControlSanitizeAllocatorMismatch:            true,
		session.ControlSanitizeUseAfterFree:                 true,
	}
	for literal := range named {
		found := false
		for _, required := range session.Phase5RequiredControls() {
			if required == literal {
				found = true
			}
		}
		if !found {
			t.Fatalf("exported control constant %q is not present, byte-identical, in Phase5RequiredControls()", literal)
		}
	}
}

func TestVerifyPhase5ControlsAndWork(t *testing.T) {
	result, err := session.VerifyPhase5ControlsAndWork(context.Background())
	if err != nil {
		t.Fatalf("VerifyPhase5ControlsAndWork returned an unexpected error: %v", err)
	}
	if result.Status != "pass" {
		t.Fatalf("expected a pass result, got status=%q diagnostics=%+v lanes=%+v", result.Status, result.Diagnostics, result.Lanes)
	}
	for _, required := range session.Phase5RequiredControls() {
		found := false
		for _, lane := range result.Lanes {
			for _, control := range lane.Controls {
				if control == required {
					found = true
				}
			}
		}
		if !found {
			t.Fatalf("required control %q never fired in any lane: %+v", required, result.Lanes)
		}
	}
}

func TestPhase5ControlsAllHaveNonzeroWork(t *testing.T) {
	result, err := session.VerifyPhase5ControlsAndWork(context.Background())
	if err != nil {
		t.Fatalf("VerifyPhase5ControlsAndWork returned an unexpected error: %v", err)
	}
	if result.Status != "pass" {
		t.Fatalf("expected a pass result, got status=%q lanes=%+v", result.Status, result.Lanes)
	}
	for _, lane := range result.Lanes {
		if lane.RecomputedWork == 0 {
			t.Fatalf("lane %s fired with zero RecomputedWork -- a control that fires with zero counted work is exactly as unproven as one that never fired", lane.ID)
		}
	}
}

// TestPhase5ToolMissingIsNamedObligationNotPass is T-05-33's own
// falsifier: with a broken ClangPath, every sanitizer control must still
// appear as a named, gate-visible operational obligation -- never omitted,
// and never rendered as a pass.
func TestPhase5ToolMissingIsNamedObligationNotPass(t *testing.T) {
	original := session.Phase5ClangPathOverrideForTest
	session.Phase5ClangPathOverrideForTest = testsupport.ProjectPath("testdata", "does-not-exist", "clang-broken")
	defer func() { session.Phase5ClangPathOverrideForTest = original }()

	result, err := session.VerifyPhase5ControlsAndWork(context.Background())
	if err != nil {
		t.Fatalf("VerifyPhase5ControlsAndWork returned an unexpected hard error for a broken toolchain: %v", err)
	}
	if result.Status == "pass" {
		t.Fatal("expected a non-pass result with a broken ClangPath, got a pass")
	}
	for _, required := range []string{
		session.ControlSanitizeRetainedPointer,
		session.ControlSanitizeUseAfterFree,
		session.ControlSanitizeAllocatorMismatch,
		session.ControlSanitizeUBSanNoRecover,
	} {
		found := false
		var laneStatus string
		for _, lane := range result.Lanes {
			for _, control := range lane.Controls {
				if control == required {
					found = true
					laneStatus = lane.Status
				}
			}
		}
		if !found {
			t.Fatalf("sanitizer control %q was not named as an obligation with a broken toolchain: %+v", required, result.Lanes)
		}
		if laneStatus == "pass" {
			t.Fatalf("sanitizer control %q rendered as a pass with a broken toolchain", required)
		}
	}
}

// phase5ExtractFunctionBody returns the exact text of the named top-level
// function, from its own "func name(" line up to (but excluding) the next
// top-level "\nfunc " line -- the same anchor-string-plus-block-extraction
// technique session_test.go's phase4OwnControlIdentifiers uses, applied to
// a Go function body instead of a shell `for control in` block.
func phase5ExtractFunctionBody(t *testing.T, source, name string) string {
	t.Helper()
	anchor := "func " + name + "("
	start := strings.Index(source, anchor)
	if start == -1 {
		t.Fatalf("function %s not found", name)
	}
	rest := source[start+len(anchor):]
	end := strings.Index(rest, "\nfunc ")
	if end == -1 {
		return rest
	}
	return rest[:end]
}

// TestSanitizeLaneNotInEditOrCheck asserts D-05-17's cost-placement rule
// at the CLI dispatch layer: the `check`/`format` commands (session.Check/
// session.CheckCommandFile, wired from cmd/lang/main.go's runCheck/
// runFormat -- the project's own edit-loop pipeline) never reference the
// sanitizer lane or the Phase 5 gate that invokes it. `runVerify` itself
// (the `verify`/`release` cost lane) is exempt -- it is the plan's own
// required integration point, per D-05-17's "mandatory in verify and
// release" placement -- so this test scopes to the check/format function
// bodies specifically, not the whole file.
func TestSanitizeLaneNotInEditOrCheck(t *testing.T) {
	mainSource, err := os.ReadFile(testsupport.ProjectPath("cmd", "lang", "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, editFunction := range []string{"runCheck", "runFormat"} {
		body := phase5ExtractFunctionBody(t, string(mainSource), editFunction)
		for _, forbidden := range []string{"VerifyPhase5SanitizeLane", "VerifyPhase5ControlsAndWork"} {
			if strings.Contains(body, forbidden) {
				t.Fatalf("cmd/lang/main.go's %s must not reference %s -- the sanitizer lane is verify/release-only", editFunction, forbidden)
			}
		}
	}

	checkSource, err := os.ReadFile(testsupport.ProjectPath("internal", "compiler", "session", "session.go"))
	if err != nil {
		t.Fatal(err)
	}
	// session.go declares Check/CheckCommandFile (the `check` pipeline) and
	// must not itself reference the sanitizer lane or the Phase 5 gate --
	// those live only in the new session_phase5*.go files this plan adds,
	// which `check` never imports into its own call graph.
	for _, forbidden := range []string{"VerifyPhase5SanitizeLane", "VerifyPhase5ControlsAndWork"} {
		if strings.Contains(string(checkSource), forbidden) {
			t.Fatalf("session.go must not reference %s", forbidden)
		}
	}
}

// TestNAT03MutationsCiteExistingPrograms is T-05-36's own falsifier,
// enumerated over the WHOLE table rather than sampled: every row with
// Subjected: true must cite a CorpusProgram path that exists on disk,
// including the two rows plan 05-07 left marked PENDING-05-08 (now closed
// by plan 05-08's fixtures landing).
func TestNAT03MutationsCiteExistingPrograms(t *testing.T) {
	rows := session.NAT03Mutations()
	subjected := 0
	for _, row := range rows {
		if !row.Subjected {
			continue
		}
		subjected++
		path := testsupport.ProjectPath(row.CorpusProgram)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("row %s cites a CorpusProgram that does not exist on disk: %s (%v)", row.ControlID, row.CorpusProgram, err)
		}
	}
	if subjected != 6 {
		t.Fatalf("expected exactly 6 subjected rows, got %d", subjected)
	}
}

// TestNAT03SanitizerRowMovesItsClaimedAxis closes D-05-22 for the row
// plan 05-07 could not assert (control:native.sanitize.allocator_mismatch,
// PENDING-05-08 until this plan): Phase5AssertMutationMovesAnAxis succeeds
// against plan 05-08's testdata/phase5/allocator_mismatch.lang, asserting
// the row's own claimed axis specifically.
func TestNAT03SanitizerRowMovesItsClaimedAxis(t *testing.T) {
	var target *session.NAT03Mutation
	for _, row := range session.NAT03Mutations() {
		row := row
		if row.ControlID == session.ControlSanitizeAllocatorMismatch {
			target = &row
		}
	}
	if target == nil {
		t.Fatal("no NAT-03 row declares control:native.sanitize.allocator_mismatch")
	}
	if !target.Subjected {
		t.Fatalf("expected the allocator-mismatch row to be Subjected: true, got %+v", target)
	}
	if err := session.Phase5AssertMutationMovesAnAxis(context.Background(), *target); err != nil {
		t.Fatalf("allocator-mismatch row did not move its claimed axis: %v", err)
	}
}

// TestNoNAT03RowRemainsPending refuses a gate whose citation is unclosed:
// no NAT-03 row's OWN declaration in session_phase5_alias.go may still
// carry plan 05-07's pending marker comment now that plan 05-08's fixtures
// exist. Checks for the exact marker-comment line, not any prose mention
// of the string PENDING-05-08 elsewhere in the file (e.g. a stale error
// message), matching the plan's own "delete exactly the two marker
// comments, nothing else" scope.
func TestNoNAT03RowRemainsPending(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("internal", "compiler", "session", "session_phase5_alias.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(source), "\n") {
		if strings.TrimSpace(line) == "// PENDING-05-08" {
			t.Fatalf("session_phase5_alias.go still carries a PENDING-05-08 marker comment: %q", line)
		}
	}
}

// phase5VerifierScriptText reads scripts/verify-phase5.sh's own text --
// shared by every test below that inspects the script, peering
// phase4VerifierScriptText's own precedent (session_test.go).
func phase5VerifierScriptText(t testing.TB) string {
	t.Helper()
	data, err := os.ReadFile(testsupport.ProjectPath("scripts", "verify-phase5.sh"))
	if err != nil {
		t.Fatalf("read scripts/verify-phase5.sh: %v", err)
	}
	return string(data)
}

// phase5OwnControlIdentifiers extracts the exact set of `control:` tokens
// from the script's OWN required-control block (the `for control in ...`
// loop following the "Phase 5's own required-control set" comment), as
// distinct from the earlier non-regression blocks re-asserting Phase 2-4
// controls against their own JSON -- conflating the blocks would let an
// earlier phase's control identifier masquerade as a Phase 5 one. Peers
// phase4OwnControlIdentifiers's own anchor-string-plus-block-extraction
// technique (session_test.go) exactly.
func phase5OwnControlIdentifiers(t testing.TB, text string) []string {
	t.Helper()
	anchor := "Phase 5's own required-control set"
	anchorIndex := strings.Index(text, anchor)
	if anchorIndex == -1 {
		t.Fatalf("scripts/verify-phase5.sh is missing its own required-control-set comment anchor")
	}
	rest := text[anchorIndex:]
	doneIndex := strings.Index(rest, "\ndone")
	if doneIndex == -1 {
		t.Fatalf("scripts/verify-phase5.sh's own required-control block has no closing done")
	}
	block := rest[:doneIndex]
	var identifiers []string
	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(line), "\\"))
		if strings.HasPrefix(trimmed, "control:") {
			identifiers = append(identifiers, trimmed)
		}
	}
	return identifiers
}

// TestPhase5RequiredControlsMatchScript asserts set equality between
// session.Phase5RequiredControls() and the identifiers named in the
// script's own required-control block, so a control added to one and
// forgotten in the other fails here rather than silently drifting apart
// (T-05-32).
func TestPhase5RequiredControlsMatchScript(t *testing.T) {
	text := phase5VerifierScriptText(t)
	fromScript := phase5OwnControlIdentifiers(t, text)
	scriptSet := make(map[string]bool, len(fromScript))
	for _, control := range fromScript {
		scriptSet[control] = true
	}
	sessionSet := make(map[string]bool, len(session.Phase5RequiredControls()))
	for _, control := range session.Phase5RequiredControls() {
		sessionSet[control] = true
	}
	for control := range sessionSet {
		if !scriptSet[control] {
			t.Fatalf("session.Phase5RequiredControls() names %s, which the script's own required-control block does not", control)
		}
	}
	for control := range scriptSet {
		if !sessionSet[control] {
			t.Fatalf("the script's required-control block names %s, which session.Phase5RequiredControls() does not", control)
		}
	}
	if len(scriptSet) != len(sessionSet) {
		t.Fatalf("script control set (%d) and session control set (%d) differ in size: script=%v session=%v", len(scriptSet), len(sessionSet), fromScript, session.Phase5RequiredControls())
	}
}

// phase5ScriptShellVarValue extracts the value of a `name=value` shell
// assignment from text -- the exact form the D-05-18 corpus-bound anchor
// block declares.
func phase5ScriptShellVarValue(t testing.TB, text, name string) string {
	t.Helper()
	anchor := name + "="
	index := strings.Index(text, anchor)
	if index == -1 {
		t.Fatalf("scripts/verify-phase5.sh is missing shell variable %s", name)
	}
	rest := text[index+len(anchor):]
	end := strings.IndexAny(rest, "\n")
	if end == -1 {
		end = len(rest)
	}
	return strings.TrimSpace(rest[:end])
}

// TestPhase5CorpusBoundMatchesScript asserts the three D-05-18 bound
// values duplicated into the script's own anchor block are byte-identical
// to session_phase5_corpus.go's exported constants.
func TestPhase5CorpusBoundMatchesScript(t *testing.T) {
	text := phase5VerifierScriptText(t)
	anchor := "Phase 5's own corpus bound"
	if !strings.Contains(text, anchor) {
		t.Fatalf("scripts/verify-phase5.sh is missing its own corpus-bound comment anchor")
	}
	cases := []struct {
		name string
		want int
	}{
		{"phase5_corpus_bound_version", session.Phase5CorpusBoundVersion},
		{"phase5_enumeration_max_depth", session.Phase5EnumerationMaxDepth},
		{"phase5_enumeration_max_statements", session.Phase5EnumerationMaxStatements},
	}
	for _, testCase := range cases {
		got := phase5ScriptShellVarValue(t, text, testCase.name)
		want := strconv.Itoa(testCase.want)
		if got != want {
			t.Fatalf("script's %s=%s does not match session.%s constant %s", testCase.name, got, testCase.name, want)
		}
	}
}

// TestPhase5VerifierScriptContract is this plan's own contract test over
// the gate script's own text: it asserts the script never invokes a
// previous-phase gate script, names every one of the ten required Phase 5
// control identifiers, names all five `--json verify testdata/phaseN`
// invocations, and pins both sanitizer option strings.
func TestPhase5VerifierScriptContract(t *testing.T) {
	text := phase5VerifierScriptText(t)
	for _, forbidden := range []string{"verify-phase1.sh", "verify-phase2.sh", "verify-phase3.sh", "verify-phase4.sh"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("scripts/verify-phase5.sh must never invoke a previous-phase gate script, but its text contains %q", forbidden)
		}
	}
	for _, control := range session.Phase5RequiredControls() {
		if !strings.Contains(text, control) {
			t.Fatalf("scripts/verify-phase5.sh's text is missing required control %s", control)
		}
	}
	for phase := 1; phase <= 5; phase++ {
		want := fmt.Sprintf("json verify testdata/phase%d", phase)
		if !strings.Contains(text, want) {
			t.Fatalf("scripts/verify-phase5.sh is missing invocation %q", want)
		}
	}
	if !strings.Contains(text, "ASAN_OPTIONS=") {
		t.Fatal("scripts/verify-phase5.sh does not pin ASAN_OPTIONS")
	}
	if !strings.Contains(text, "UBSAN_OPTIONS=") {
		t.Fatal("scripts/verify-phase5.sh does not pin UBSAN_OPTIONS")
	}
	phase4Path := testsupport.ProjectPath("scripts", "verify-phase4.sh")
	if _, err := os.Stat(phase4Path); err != nil {
		t.Fatalf("scripts/verify-phase4.sh must still exist unchanged: %v", err)
	}
}

// TestPhase5SanitizerOptionsMatchScript asserts the ASAN_OPTIONS string
// pinned in scripts/verify-phase5.sh is byte-identical to
// native.ASanOptions -- what catches a typo'd options string (D-05-13).
func TestPhase5SanitizerOptionsMatchScript(t *testing.T) {
	text := phase5VerifierScriptText(t)
	if !strings.Contains(text, "ASAN_OPTIONS='"+native.ASanOptions+"'") {
		t.Fatalf("scripts/verify-phase5.sh's pinned ASAN_OPTIONS is not byte-identical to native.ASanOptions (%q)", native.ASanOptions)
	}
	if !strings.Contains(text, "UBSAN_OPTIONS='"+native.UBSanOptions+"'") {
		t.Fatalf("scripts/verify-phase5.sh's pinned UBSAN_OPTIONS is not byte-identical to native.UBSanOptions (%q)", native.UBSanOptions)
	}
}

// TestPhase5RequiredControlsIsFifteen pins the control count at exactly 15
// (this plan's close-out task): the ten controls plans 05-04 through 05-08
// introduced, plus the three reducer vacuity controls (plan 05-12) and the
// two QLT-01 registry controls (plan 05-11) this plan wires in. This is a
// COUNT assertion, distinct from TestPhase5RequiredControlsMatchScript's
// SET-EQUALITY assertion -- the two catch different mistakes: a control
// silently dropped from both copies in lockstep would still pass set
// equality (both sides shrink together) but fails this count pin.
func TestPhase5RequiredControlsIsFifteen(t *testing.T) {
	got := len(session.Phase5RequiredControls())
	if got != 15 {
		t.Fatalf("expected exactly 15 required Phase 5 controls, got %d: %v", got, session.Phase5RequiredControls())
	}
}

// TestPhase5EveryDeclaredControlActuallyFires asserts set-equality between
// Phase5RequiredControls() and the set of control IDs that actually
// reported a fired-or-operational status in a real
// VerifyPhase5ControlsAndWork run. A declared-but-never-firing control is
// "not run rendered as pass", the exact shape this project's gate
// discipline exists to prevent (T-05-53).
func TestPhase5EveryDeclaredControlActuallyFires(t *testing.T) {
	result, err := session.VerifyPhase5ControlsAndWork(context.Background())
	if err != nil {
		t.Fatalf("VerifyPhase5ControlsAndWork returned an unexpected error: %v", err)
	}
	if result.Status != "pass" {
		t.Fatalf("expected a pass result, got status=%q diagnostics=%+v lanes=%+v", result.Status, result.Diagnostics, result.Lanes)
	}
	fired := make(map[string]bool)
	for _, lane := range result.Lanes {
		for _, control := range lane.Controls {
			fired[control] = true
		}
	}
	declared := make(map[string]bool, len(session.Phase5RequiredControls()))
	for _, control := range session.Phase5RequiredControls() {
		declared[control] = true
	}
	for control := range declared {
		if !fired[control] {
			t.Fatalf("declared control %q never actually fired in a real VerifyPhase5ControlsAndWork run: %+v", control, result.Lanes)
		}
	}
	for control := range fired {
		if !declared[control] {
			t.Fatalf("control %q fired in a real run but is not declared in Phase5RequiredControls()", control)
		}
	}
}
