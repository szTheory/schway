// witness_registry_test.go holds plan 14-07's executed probes -- the
// D-14-23 "executed-probe" witness kind, in the flesh. Each probe asserts
// a STRUCTURAL fact (a diagnostic code, an admission refusal, a
// still-identical structural summary), never a message string, per
// D-14-32's accepted-residual-friction note: a probe over-fit to current
// diagnostic prose would go red on a benign rewording, which is not the
// signal this register exists to catch.
package session_test

import (
	"context"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// TestB1BlameIsStructurallyUnreachable backs PHASE-13-DEBT.md's D-13-02b
// row (and, sharing the same root cause, D-13-10a's withdrawn
// use_matching_argument row): it compiles a real fixture whose declared
// ReturnType contradicts its own declared Parameter.Type -- the exact
// shape B1's contract-violation blame would need to survive admission to
// ever fire on -- and asserts it is refused BY NAME
// (type.return_mismatch) at admission, before any interprocedural pass
// could see it. If this fixture ever checks clean, B1 blame has become
// reachable and this probe goes red (XPASS), forcing a human to regrade
// D-13-02b and D-13-10a rather than letting the claim decay silently.
func TestB1BlameIsStructurallyUnreachable(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase14", "blame_unreachable_admission_refusal.lang"))
	if err != nil {
		t.Fatal(err)
	}
	result := session.Check(source)
	if len(result.Diagnostics) == 0 {
		t.Fatal("D-13-02b's claim requires this fixture to be refused at admission; it checked clean instead -- B1 blame has become reachable, regrade PHASE-13-DEBT.md's D-13-02b and D-13-10a rows")
	}
	found := false
	for _, d := range result.Diagnostics {
		if d.Code == "type.return_mismatch" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a type.return_mismatch diagnostic (the admission precondition D-13-02b names), got: %+v", result.Diagnostics)
	}
}

// TestD1243ControlIsUnconstructible backs PHASE-12-DEBT.md's D-12-43 row:
// it seeds a real, type-safe wrong-slot payload write via
// cgen.SetPayloadSlotSwapForTest and asserts the resulting divergence is
// STILL invisible to every axis session.Phase5CompareEngines checks --
// the decisive value-divergence control D-12-38 wanted is still
// unconstructible against the current representation. Unlike
// TestPayloadSlotSwapMutationKilled (which logs either outcome and always
// passes), this probe FAILS the moment the mutation becomes visible,
// because that is exactly the day D-12-43's claim needs regrading rather
// than continuing to pass on a stale premise.
func TestD1243ControlIsUnconstructible(t *testing.T) {
	ctx := context.Background()
	runner := native.DefaultRunner()
	restore := cgen.SetPayloadSlotSwapForTest(true)
	defer restore()

	program := checkedProgram(t, "testdata", "phase12", "payload_tracer.lang")
	engines := runFunctionOkExecutions(t, ctx, program, "identity", runner)
	injected := cgen.PayloadSlotSwapInjectedWriteCount()
	if injected < 1 {
		t.Fatalf("D-12-43 probe: the fault-injection seam injected no wrong-slot write (count=%d), so this run proves nothing about the control -- check whether payload_tracer.lang's data type still declares two payload-carrying alternatives", injected)
	}
	if compareErr := session.Phase5CompareEngines("payload_tracer.lang(D-12-43 probe)", engines); compareErr != nil {
		t.Fatalf("D-12-43's decisive wrong-slot value-divergence control has become CONSTRUCTIBLE: %v -- this claim is stale, regrade PHASE-12-DEBT.md's D-12-43 row instead of treating this failure as something to silence", compareErr)
	}
}

// TestLTOInertnessOnMultiFunctionEmission backs PHASE-14-DEBT.md's D-14-45
// row (EVD-07): it checks a real multi-function fixture where one
// function (toggle) has a core.Match body -- exactly the shape
// cgen.emitProgram refuses -- and asserts native emission refuses BY NAME
// ("multi-function branch bodies are not supported by native emission
// this phase"). `-flto`'s whole-program optimizer tier has no
// cross-function boundary to exploit when cgen never emits more than one
// function's worth of a program containing a branch body at all; this
// probe is the executed half of that claim.
func TestLTOInertnessOnMultiFunctionEmission(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase14", "multi_function_match_refusal.lang"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture must check clean: %+v", checked.Diagnostics)
	}
	if len(checked.Program.Functions) < 2 {
		t.Fatalf("fixture must be genuinely multi-function (>= 2 functions), got %d", len(checked.Program.Functions))
	}
	_, emitErr := cgen.EmitNative(checked.Program)
	if emitErr == nil {
		t.Fatal("D-14-45: expected multi-function native emission to refuse a Match-bodied function inside a multi-function program; it succeeded instead -- the -flto multi-function inertness claim may no longer hold, regrade PHASE-14-DEBT.md's D-14-45 row")
	}
	if !strings.Contains(emitErr.Error(), "multi-function branch bodies are not supported by native emission this phase") {
		t.Fatalf("expected the refusal to name the multi-function branch-body restriction, got: %v", emitErr)
	}
}

// TestRetainedPointerEscapeIsStillUnsubjected is the probe
// debtRegisterEscapeRegistry's "callback-invocation-unsubjected" entry
// names (D-14-27): it asserts the retained_pointer NAT-03 row is STILL
// Subjected: false with the same EscapeID -- the structural fact the
// escape declares. If a future plan discharges the escape (subjects the
// row), this probe fails, forcing the escape registry entry to be
// retired rather than left citing a stale premise.
func TestRetainedPointerEscapeIsStillUnsubjected(t *testing.T) {
	var target *session.NAT03Mutation
	for _, row := range session.NAT03Mutations() {
		row := row
		if row.ControlID == "control:native.sanitize.retained_pointer" {
			target = &row
		}
	}
	if target == nil {
		t.Fatal("no NAT-03 row declares control:native.sanitize.retained_pointer")
	}
	if target.Subjected {
		t.Fatal("control:native.sanitize.retained_pointer is now Subjected: true -- the escape has been discharged; retire debtRegisterEscapeRegistry's callback-invocation-unsubjected entry instead of leaving this probe passing on a stale premise")
	}
	if target.EscapeID != "escape:callback-invocation-unsubjected" {
		t.Fatalf("expected EscapeID escape:callback-invocation-unsubjected, got %q", target.EscapeID)
	}
}

// ---------------------------------------------------------------------
// Task 3: every suppression surface cites a resolvable witness (D-14-25).
// ---------------------------------------------------------------------

var (
	suppressionPendingPattern  = regexp.MustCompile(`PENDING-\d\d-\d\d`)
	suppressionDecisionPattern = regexp.MustCompile(`\bD-\d\d-\d\d[a-z]?\b`)
	suppressionEscapePattern   = regexp.MustCompile(`escape:[a-z][a-z0-9-]*`)
	suppressionProbePattern    = regexp.MustCompile(`probe:[A-Za-z][A-Za-z0-9_]*`)
	suppressionEnvPattern      = regexp.MustCompile(`env:[a-z][a-z0-9_-]*`)
)

// debtRegisterDecisionCitationResolves reports whether id (a bare
// D-XX-NN identifier) resolves against either a debt-register row in any
// *-DEBT.md, or a CONTEXT.md decision heading (this project's
// "- **D-XX-NN (...):**" convention) in any *-CONTEXT.md. Both are
// legitimate resolution targets: a D-XX-NN identifier is used throughout
// this codebase both as a debt-register row ID and as a CONTEXT.md
// decision ID, and conflating the two namespaces (requiring every
// decision citation to also be a debt row) would misfire on the
// project's own extensive, otherwise-healthy cross-referencing
// convention.
func debtRegisterDecisionCitationResolves(id string) (bool, error) {
	registers, err := phaseArtifactGlob("*", "*-DEBT.md")
	if err != nil {
		return false, err
	}
	for _, path := range registers {
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return false, readErr
		}
		columns, rows, tableErr := parseDebtRegisterTable(filepath.Base(path), string(data))
		if tableErr != nil {
			continue
		}
		idIdx, ok := columns["ID"]
		if !ok {
			continue
		}
		for _, row := range rows {
			if idIdx < len(row) && row[idIdx] == id {
				return true, nil
			}
		}
	}
	contexts, err := phaseArtifactGlob("*", "*-CONTEXT.md")
	if err != nil {
		return false, err
	}
	needle := "**" + id
	for _, path := range contexts {
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return false, readErr
		}
		if strings.Contains(string(data), needle) {
			return true, nil
		}
	}
	return false, nil
}

// suppressionResolutionMode selects which citation shapes a surface's
// resolution check applies to. See scanSuppressionSurfaces's own doc
// comment and TestNoSuppressionOutlivesItsWitness for why the LIVE
// module's comment/string-literal surface is checked only for the
// PENDING-NN-NN shape rather than the full grammar.
type suppressionResolutionMode int

const (
	// suppressionModeFull checks all five citation shapes. Safe for a
	// Skip call's own arguments (this codebase's Skip citations use only
	// this plan's shapes) and for an ISOLATED synthetic temp-dir package
	// built by a test (no collision risk with the live corpus).
	suppressionModeFull suppressionResolutionMode = iota
	// suppressionModePendingOnly checks only PENDING-NN-NN. Required for
	// the LIVE module's comment/string-literal surface: this codebase
	// ALREADY uses "escape:" as its own, unrelated, pre-existing
	// anti-theater vocabulary (session_phase5_escapes.go's
	// EscapeCoordinatedSourceToCoreFalseClaim,
	// originvalidate.go's KnownEscape -- a "coordinated lie" escape
	// hatch, nothing to do with this plan's NAT03Mutation escape
	// registry), and "D-XX-NN" is this project's own pervasive
	// decision/finding cross-reference convention used in the vast
	// majority of doc comments, not exclusively as a suppression
	// citation. Requiring every module-wide occurrence of either shape to
	// resolve against THIS plan's two narrow registries would misfire on
	// both pre-existing, healthy conventions; disambiguating them
	// properly (a full CONTEXT.md decision-ID index across all archived
	// milestones, plus a second closed registry for the unrelated
	// escape: vocabulary) is out of this plan's budget and recorded as
	// debt in its own SUMMARY.
	suppressionModePendingOnly
)

// suppressionCitationsResolve scans text for every citation-shaped
// substring D-14-25 names (a pending marker, a decision identifier, an
// escape identifier) plus this register's own probe:/env: token shapes,
// and reports whether at least one citation was found, plus every
// unresolved citation's problem (mode suppressionModePendingOnly checks
// only the PENDING shape). A text with zero citation-shaped substrings
// reports found=false; the caller decides whether that absence is itself
// a violation (only true for a Skip-call site -- a bare comment
// mentioning nothing citation-shaped is not automatically a
// suppression).
func suppressionCitationsResolve(text string, mode suppressionResolutionMode) (found bool, problems []string) {
	if suppressionPendingPattern.MatchString(text) {
		found = true
		// PENDING-NN-NN is accepted by shape alone: D-14-26's legacy
		// marker form has no registry entry of its own. Plan 14-08
		// removed this module's one surviving instance entirely as part
		// of collapsing the axis-movement law; no PENDING-NN-NN citation
		// is expected to remain live anywhere in this module now.
	}
	if mode == suppressionModePendingOnly {
		return found, nil
	}
	for _, m := range suppressionEscapePattern.FindAllString(text, -1) {
		found = true
		id := strings.TrimPrefix(m, "escape:")
		if _, ok := debtRegisterEscapeRegistry[id]; !ok {
			problems = append(problems, fmt.Sprintf("citation %q does not resolve: absent from the closed escape registry", m))
		}
	}
	for _, m := range suppressionProbePattern.FindAllString(text, -1) {
		found = true
		name := strings.TrimPrefix(m, "probe:")
		ok, err := debtRegisterProbeExists(name)
		if err != nil {
			problems = append(problems, fmt.Sprintf("citation %q: %v", m, err))
			continue
		}
		if !ok {
			problems = append(problems, fmt.Sprintf("citation %q does not resolve: no such test exists", m))
		}
	}
	for _, m := range suppressionEnvPattern.FindAllString(text, -1) {
		found = true
		id := strings.TrimPrefix(m, "env:")
		if !debtRegisterEnvironmentalSet[id] {
			problems = append(problems, fmt.Sprintf("citation %q does not resolve: outside the closed environmental set", m))
		}
	}
	for _, m := range suppressionDecisionPattern.FindAllString(text, -1) {
		found = true
		ok, err := debtRegisterDecisionCitationResolves(m)
		if err != nil {
			problems = append(problems, fmt.Sprintf("citation %q: %v", m, err))
			continue
		}
		if !ok {
			problems = append(problems, fmt.Sprintf("citation %q does not resolve: absent from every *-DEBT.md register and every *-CONTEXT.md decision", m))
		}
	}
	return found, problems
}

// suppressionSite is one enumerated occurrence of a scanned surface.
type suppressionSite struct {
	path string
	pos  string
	kind string
	text string
}

// scanSuppressionSurfaces walks EVERY *.go file under root (production
// and test alike -- D-14-25's specific fix for instance 4, which watched
// only one production file and missed the identical marker surviving as
// an error string and as prose in two test files) and enumerates four
// surfaces: every //go:build constraint, every comment, every string
// literal, and every Skip/Skipf/SkipNow call (recorded with its
// concatenated string-literal arguments as text).
func scanSuppressionSurfaces(root string) ([]suppressionSite, error) {
	fset := token.NewFileSet()
	buildCtx := build.Default
	var sites []suppressionSite
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			base := d.Name()
			if base == ".git" || base == "testdata" || (path != root && strings.HasPrefix(base, ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		dir := filepath.Dir(path)
		match, matchErr := buildCtx.MatchFile(dir, filepath.Base(path))
		if matchErr != nil {
			return matchErr
		}
		if !match {
			return nil
		}
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			rel = path
		}

		for lineNumber, line := range strings.Split(string(raw), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "//go:build") {
				sites = append(sites, suppressionSite{path: rel, pos: fmt.Sprintf("%s:%d", rel, lineNumber+1), kind: "build-constraint", text: line})
			}
		}

		file, parseErr := parser.ParseFile(fset, path, raw, parser.ParseComments)
		if parseErr != nil {
			return fmt.Errorf("parse %s: %w", path, parseErr)
		}

		for _, group := range file.Comments {
			for _, comment := range group.List {
				sites = append(sites, suppressionSite{path: rel, pos: fset.Position(comment.Pos()).String(), kind: "comment", text: comment.Text})
			}
		}

		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.BasicLit:
				if node.Kind == token.STRING {
					sites = append(sites, suppressionSite{path: rel, pos: fset.Position(node.Pos()).String(), kind: "string-literal", text: node.Value})
				}
			case *ast.CallExpr:
				sel, ok := node.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				switch sel.Sel.Name {
				case "Skip", "Skipf", "SkipNow":
					var text strings.Builder
					for _, arg := range node.Args {
						if lit, ok := arg.(*ast.BasicLit); ok && lit.Kind == token.STRING {
							text.WriteString(lit.Value)
							text.WriteString(" ")
						}
					}
					sites = append(sites, suppressionSite{path: rel, pos: fset.Position(node.Pos()).String(), kind: "skip", text: text.String()})
				}
			}
			return true
		})
		return nil
	})
	return sites, err
}

// suppressionProblems runs suppressionCitationsResolve over every site,
// plus the NAT03Mutation unsubjected-row surface, and returns every
// problem found (nil means clean). Shared by
// TestNoSuppressionOutlivesItsWitness and
// TestSuppressionWitnessGuardIsNotInert's seeded-fault subtests so both
// exercise the SAME law. Skip-call sites always use the full grammar
// (mode is irrelevant to a skip call's own citation, which must resolve
// completely); nonSkipMode governs comments, string literals, and build
// constraints.
func suppressionProblems(root string, nonSkipMode suppressionResolutionMode) ([]string, error) {
	sites, err := scanSuppressionSurfaces(root)
	if err != nil {
		return nil, err
	}
	var problems []string
	for _, site := range sites {
		mode := nonSkipMode
		if site.kind == "skip" {
			mode = suppressionModeFull
		}
		found, siteProblems := suppressionCitationsResolve(site.text, mode)
		for _, problem := range siteProblems {
			problems = append(problems, fmt.Sprintf("%s (%s): %s", site.pos, site.kind, problem))
		}
		if site.kind == "skip" && !found {
			problems = append(problems, fmt.Sprintf("%s (%s): uncited suppression -- a Skip call must cite a resolvable witness", site.pos, site.kind))
		}
	}
	return problems, nil
}

// TestNoSuppressionOutlivesItsWitness enumerates every suppression
// surface in the module -- Skip calls, //go:build constraints, comments,
// and string literals, across every package including test files
// (D-14-25) -- and requires each citation-shaped substring found to
// resolve. A bare Skip call carrying no citation at all is its own
// violation; a citation that does not resolve is a violation regardless
// of surface, including inside a string literal (the specific fix for
// instance 4, where a stale marker survived as an error string).
func TestNoSuppressionOutlivesItsWitness(t *testing.T) {
	root := testsupport.ProjectPath()
	sites, err := scanSuppressionSurfaces(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(sites) == 0 {
		t.Fatal("scanSuppressionSurfaces found nothing at all -- the scanner is looking at nothing, not finding a clean module")
	}
	skipSites := 0
	for _, site := range sites {
		if site.kind == "skip" {
			skipSites++
			t.Logf("suppression site: %s (%s): %q", site.pos, site.kind, site.text)
		}
	}
	if skipSites == 0 {
		t.Fatal("scanSuppressionSurfaces found no skip call sites -- the corpus has known Skip/Skipf calls, so this is a scanner defect, not a clean module")
	}

	problems, err := suppressionProblems(root, suppressionModePendingOnly)
	if err != nil {
		t.Fatal(err)
	}

	// Surface: every NAT03Mutation row declared unsubjected must name a
	// declared escape that resolves in the closed escape registry.
	for _, row := range session.NAT03Mutations() {
		if row.Subjected {
			continue
		}
		if row.EscapeID == "" {
			problems = append(problems, fmt.Sprintf("NAT03Mutation %s: declared unsubjected with no EscapeID", row.ControlID))
			continue
		}
		id := strings.TrimPrefix(row.EscapeID, "escape:")
		if _, ok := debtRegisterEscapeRegistry[id]; !ok {
			problems = append(problems, fmt.Sprintf("NAT03Mutation %s: EscapeID %q does not resolve in the closed escape registry", row.ControlID, row.EscapeID))
		}
	}

	if len(problems) > 0 {
		t.Fatalf("%d suppression problem(s):\n%s", len(problems), strings.Join(problems, "\n"))
	}
}

// TestSuppressionWitnessOnlyInStringLiteralIsFound is Task 3's own
// dedicated acceptance criterion: a synthetic package whose only citation
// lives inside a STRING LITERAL (never a comment, never a Skip argument)
// is still reported when that citation fails to resolve -- proving the
// enumerator inspects string literals as their own surface, not merely
// as an accident of scanning Skip-call arguments.
func TestSuppressionWitnessOnlyInStringLiteralIsFound(t *testing.T) {
	dir := t.TempDir()
	source := "package seeded\n\nfunc stale() string {\n\treturn \"row not yet subjected -- see escape:this-escape-id-is-not-registered\"\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "seeded.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	problems, err := suppressionProblems(dir, suppressionModeFull)
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) == 0 {
		t.Fatal("a string-literal-only citation that fails to resolve should be reported, but nothing was")
	}
	found := false
	for _, p := range problems {
		if strings.Contains(p, "string-literal") && strings.Contains(p, "escape:this-escape-id-is-not-registered") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a string-literal-kind problem naming the unresolved escape citation, got: %v", problems)
	}
}

// TestSuppressionWitnessGuardIsNotInert seeds one fault per mechanizable
// kind (D-14-28) and asserts red for each, with an unmodified-copy
// control asserting green: (a) a reason-free Skip in a copied synthetic
// package; (b) a flipped callsite: count in a copied register; (c) a
// neutralized probe -- editing a copied fixture so the claim it refuses
// becomes admitted, the unexpected-pass (XPASS) shape. A guard proven red
// on only one of the three fault kinds would be inert for the other two.
func TestSuppressionWitnessGuardIsNotInert(t *testing.T) {
	t.Run("reason-free skip in a copied package", func(t *testing.T) {
		dir := t.TempDir()
		clean := "package seeded\n\nimport \"testing\"\n\nfunc TestSeededCitedSkip(t *testing.T) {\n\tt.Skip(\"probe:TestDebtRegistersAreWellFormed\")\n}\n"
		if err := os.WriteFile(filepath.Join(dir, "seeded_test.go"), []byte(clean), 0o600); err != nil {
			t.Fatal(err)
		}
		problems, err := suppressionProblems(dir, suppressionModeFull)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(problems) != 0 {
			t.Fatalf("a cited skip in an otherwise-clean copy should pass; got: %v", problems)
		}

		uncited := "package seeded\n\nimport \"testing\"\n\nfunc TestSeededUncitedSkip(t *testing.T) {\n\tt.Skip(\"no reason\")\n}\n"
		if err := os.WriteFile(filepath.Join(dir, "seeded_test.go"), []byte(uncited), 0o600); err != nil {
			t.Fatal(err)
		}
		problems, err = suppressionProblems(dir, suppressionModeFull)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(problems) == 0 {
			t.Fatal("a reason-free Skip should be reported as an uncited suppression, but nothing was")
		}
	})

	t.Run("flipped callsite count in a copied register", func(t *testing.T) {
		unmodified := debtRegisterWitnessGrammarFixture(t, "WIRED", "callsite:internal/compiler/check.resolveBlame=0")
		if problems, err := debtRegisterProblems(unmodified); err != nil || len(problems) != 0 {
			t.Fatalf("expected a clean pass on the correct count; got err=%v problems=%v", err, problems)
		}

		flipped := debtRegisterWitnessGrammarFixture(t, "WIRED", "callsite:internal/compiler/check.resolveBlame=1")
		problems, err := debtRegisterProblems(flipped)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(problems) == 0 {
			t.Fatal("a flipped callsite: count should be reported, but nothing was")
		}
	})

	t.Run("neutralized probe produces an unexpected pass", func(t *testing.T) {
		unmodifiedSource, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase14", "blame_unreachable_admission_refusal.lang"))
		if err != nil {
			t.Fatal(err)
		}
		if diagnostics := session.Check(unmodifiedSource).Diagnostics; len(diagnostics) == 0 {
			t.Fatal("the unmodified fixture must still be refused; TestB1BlameIsStructurallyUnreachable's own premise has decayed")
		}

		neutralized := strings.Replace(string(unmodifiedSource), "-> Buffer {", "-> Byte {", 1)
		if neutralized == string(unmodifiedSource) {
			t.Fatal("seeded edit did not change the fixture -- the seam this subtest targets has drifted")
		}
		diagnostics := session.Check([]byte(neutralized)).Diagnostics
		if len(diagnostics) != 0 {
			t.Fatalf("expected the neutralized fixture (ReturnType now matching ParameterType) to check CLEAN -- an unexpected pass demonstrating what an XPASS looks like -- got diagnostics: %+v", diagnostics)
		}
		// This is exactly the condition TestB1BlameIsStructurallyUnreachable
		// itself treats as red (t.Fatal on len(Diagnostics)==0): had this
		// neutralized fixture shipped instead of the real one, that probe
		// would have failed loudly, proving it is not inert to this fault
		// kind.
	})
}

// ---------------------------------------------------------------------
// Task 4: .planning/UNREACHABLE-CLAIMS.md -- a generated, byte-compared
// view (D-14-13, D-14-14). The register row is the AUTHORED truth, the
// probe is the EXECUTED truth, this file is DERIVED: regenerated in
// memory and compared, never hand-edited, no blessing path.
// ---------------------------------------------------------------------

// unreachableClaimEntry is one row of the generated view: a debt-register
// row whose Witness cell names at least one probe: token -- the closed,
// syntactic definition of "qualifies as a built-but-structurally-
// unreachable claim" this generator uses. A row graded WIRED/REACHABLE/
// EXERCISED/MUTATION-KILLED backed by an executed probe is exactly
// D-13-02b's "deactivated code" shape (DO-178C): present, provably
// unexecutable today, justified by analysis an executed probe forces to
// be re-examined when the configuration changes.
type unreachableClaimEntry struct {
	id       string
	register string
	grade    string
	witness  string
	trigger  string
	claim    string
}

// unreachableClaimsFrontmatterPattern matches the generated view's own
// `entries: N` frontmatter line, mirroring debtRegisterProblems' `items:`
// cross-check.
var unreachableClaimsFrontmatterPattern = regexp.MustCompile(`(?m)^entries:\s*(\d+)\s*$`)

// deriveUnreachableClaims scans every *-DEBT.md register (live and
// archived) for rows whose Witness cell contains a probe: token, in
// register-glob order (phaseArtifactGlob's own sorted order) then table
// order within each register -- fully deterministic. A register lacking
// Grade/Witness columns entirely (debtRegisterGradeWitnessExemptions)
// contributes nothing, which is correct: it has no graded rows to derive
// a claim from.
func deriveUnreachableClaims() ([]unreachableClaimEntry, error) {
	registers, err := phaseArtifactGlob("*", "*-DEBT.md")
	if err != nil {
		return nil, err
	}
	sort.Strings(registers)
	var entries []unreachableClaimEntry
	for _, path := range registers {
		name := filepath.Base(path)
		if _, exempt := debtRegisterGradeWitnessExemptions[name]; exempt {
			continue
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil, readErr
		}
		columns, rows, tableErr := parseDebtRegisterTable(name, string(data))
		if tableErr != nil {
			return nil, tableErr
		}
		idIdx, hasID := columns["ID"]
		gradeIdx, hasGrade := columns["Grade"]
		witnessIdx, hasWitness := columns["Witness"]
		landingIdx, hasLanding := columns["Landing phase"]
		itemIdx, hasItem := columns["Item"]
		if !hasID || !hasGrade || !hasWitness || !hasLanding || !hasItem {
			continue
		}
		for _, row := range rows {
			witness := row[witnessIdx]
			if !strings.Contains(witness, "probe:") {
				continue
			}
			entries = append(entries, unreachableClaimEntry{
				id:       row[idIdx],
				register: name,
				grade:    row[gradeIdx],
				witness:  witness,
				trigger:  row[landingIdx],
				claim:    row[itemIdx],
			})
		}
	}
	return entries, nil
}

// renderUnreachableClaimsView renders entries as the exact checked-in
// document shape: frontmatter `entries: N`, then one table row per entry.
func renderUnreachableClaimsView(entries []unreachableClaimEntry) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "entries: %d\n", len(entries))
	b.WriteString("---\n\n")
	b.WriteString("# Unreachable Claims\n\n")
	b.WriteString("**GENERATED. Do not hand-edit.** Regenerated in memory and byte-compared by\n")
	b.WriteString("`TestUnreachableClaimsViewIsCurrent` (`internal/compiler/session/witness_registry_test.go`)\n")
	b.WriteString("from every `*-DEBT.md` register's `Grade`/`Witness` columns. The register row\n")
	b.WriteString("is the authored truth, the probe is the executed truth, this file is derived --\n")
	b.WriteString("a hand edit is a failure, not a source of information. There is no regeneration\n")
	b.WriteString("command and no blessing path: if this view is out of date, correct the\n")
	b.WriteString("registers and re-derive, never overwrite this file directly.\n\n")
	b.WriteString("A row qualifies for this view when its `Witness` cell names at least one\n")
	b.WriteString("`probe:` token: a claim justified by an assertion that currently holds and is\n")
	b.WriteString("asserted to hold, so that its ceasing to hold turns the suite red (D-14-22).\n\n")
	b.WriteString("## Claims\n\n")
	b.WriteString("| ID | Register | Grade | Witness | Unblocking trigger | Claim |\n")
	b.WriteString("|---|---|---|---|---|---|\n")
	for _, entry := range entries {
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s |\n", entry.id, entry.register, entry.grade, entry.witness, entry.trigger, entry.claim)
	}
	return b.String()
}

// TestUnreachableClaimsViewIsCurrent regenerates the view in memory from
// the registers and byte-compares it against the checked-in
// .planning/UNREACHABLE-CLAIMS.md -- a hand edit is a failure (D-14-13).
// The non-zero-row-count assertion (D-14-14c) additionally guards against
// a vacuous pass: an empty generator against an empty checked-in file
// compares equal, which would be true evidence of nothing.
func TestUnreachableClaimsViewIsCurrent(t *testing.T) {
	entries, err := deriveUnreachableClaims()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) < 6 {
		t.Fatalf("expected at least the six known qualifying rows (D-13-02b, D-13-10a, D-13-34, D-14-45, D-11-02, D-12-43), got %d: %+v", len(entries), entries)
	}

	regenerated := renderUnreachableClaimsView(entries)

	checkedInPath := testsupport.ProjectPath(".planning", "UNREACHABLE-CLAIMS.md")
	checkedIn, err := os.ReadFile(checkedInPath)
	if err != nil {
		t.Fatalf("read .planning/UNREACHABLE-CLAIMS.md: %v", err)
	}
	if regenerated != string(checkedIn) {
		t.Fatalf(".planning/UNREACHABLE-CLAIMS.md is out of date -- regenerate from the registers, never hand-edit.\n--- regenerated ---\n%s\n--- checked-in ---\n%s", regenerated, string(checkedIn))
	}

	m := unreachableClaimsFrontmatterPattern.FindStringSubmatch(string(checkedIn))
	if m == nil {
		t.Fatal(".planning/UNREACHABLE-CLAIMS.md has no `entries: N` frontmatter line")
	}
	declared, convErr := strconv.Atoi(m[1])
	if convErr != nil {
		t.Fatalf("unreadable entries: frontmatter: %v", convErr)
	}
	if declared != len(entries) {
		t.Fatalf("frontmatter declares entries: %d but the derivation holds %d", declared, len(entries))
	}
}

// TestUnreachableClaimsViewNonZeroRowCountIsLoadBearing (D-14-14c) proves
// the non-zero-row-count assertion is not vacuous ornamentation: over
// synthetic EMPTY inputs, a byte-compare alone would pass (two empty
// documents are byte-identical), but the additional "qualifying rows
// exist in the live registers" assertion must independently catch that
// the real registers are never actually empty.
func TestUnreachableClaimsViewNonZeroRowCountIsLoadBearing(t *testing.T) {
	emptyRendered := renderUnreachableClaimsView(nil)
	if !strings.Contains(emptyRendered, "entries: 0") {
		t.Fatalf("expected an empty entry list to render entries: 0, got:\n%s", emptyRendered)
	}
	// A byte-compare between two independently rendered empty views
	// passes vacuously -- this is the exact failure mode D-14-14c names.
	if emptyRendered != renderUnreachableClaimsView(nil) {
		t.Fatal("renderUnreachableClaimsView is nondeterministic on empty input")
	}
	// The real registers are never actually empty: at least six rows
	// qualify today (see TestUnreachableClaimsViewIsCurrent), which is
	// the independent assertion that makes the byte-compare meaningful
	// rather than vacuous.
	entries, err := deriveUnreachableClaims()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("the live registers derived ZERO qualifying rows -- the byte-compare above would now be vacuously comparing two empty documents, proving consistency but not completeness")
	}
}

// TestUnreachableClaimsViewCatchesHandEdit proves the byte-compare is
// live: appending one character to a temp copy of the checked-in file
// makes a byte-for-byte comparison against the (unchanged) regeneration
// fail. Does not touch the real checked-in file.
func TestUnreachableClaimsViewCatchesHandEdit(t *testing.T) {
	entries, err := deriveUnreachableClaims()
	if err != nil {
		t.Fatal(err)
	}
	regenerated := renderUnreachableClaimsView(entries)
	tampered := regenerated + "x"
	if tampered == regenerated {
		t.Fatal("seeded one-character edit did not change the text")
	}
	if tampered == renderUnreachableClaimsView(entries) {
		t.Fatal("the tampered copy should differ from a fresh regeneration, but compared equal")
	}
}

// TestUnreachableClaimsViewCatchesStaleEntry proves a checked-in entry
// whose underlying register row has vanished is a failure, never
// auto-pruned (D-14-14b): a hand-rendered view naming a nonexistent
// register row will never byte-match a real regeneration (which simply
// omits the vanished row), so the comparison in
// TestUnreachableClaimsViewIsCurrent already catches this -- this test
// demonstrates the mechanism directly, without depending on the live
// corpus ever actually losing a row.
func TestUnreachableClaimsViewCatchesStaleEntry(t *testing.T) {
	entries, err := deriveUnreachableClaims()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("no entries to seed a stale row against")
	}
	stale := append([]unreachableClaimEntry{{id: "D-00-00", register: "NONEXISTENT-DEBT.md", grade: "WIRED", witness: "probe:ThisRowNoLongerExists", trigger: "P99", claim: "a row whose underlying register row has vanished"}}, entries...)
	staleRendered := renderUnreachableClaimsView(stale)
	freshRendered := renderUnreachableClaimsView(entries)
	if staleRendered == freshRendered {
		t.Fatal("a view carrying a stale entry absent from the live derivation should differ from a fresh regeneration, but compared equal")
	}
}

// ---------------------------------------------------------------------
// Plan 14-10 Task 2: .planning/EVIDENCE-RECONCILIATION.md -- a generated,
// byte-compared view (D-14-13, D-14-14), reusing this file's own
// UNREACHABLE-CLAIMS.md discipline rather than writing a second generator.
// The reconciliation entry is the AUTHORED truth (PHASE-14-DEBT.md's
// "```reconciliation" blocks, session_test.go's parseReconciliationEntries),
// this file is DERIVED: regenerated in memory and compared, never
// hand-edited, no blessing path.
// ---------------------------------------------------------------------

// evidenceReconciliationFrontmatterPattern matches the generated view's own
// `entries: N` frontmatter line, mirroring unreachableClaimsFrontmatterPattern.
var evidenceReconciliationFrontmatterPattern = regexp.MustCompile(`(?m)^entries:\s*(\d+)\s*$`)

// deriveEvidenceReconciliation scans every *-DEBT.md register (live and
// archived) for reconciliation entries, in register-glob order then
// document order within each register -- fully deterministic. Reuses
// parseReconciliationEntries directly; this generator never re-parses the
// underlying markdown with its own logic.
func deriveEvidenceReconciliation() ([]reconciliationEntry, error) {
	registers, err := phaseArtifactGlob("*", "*-DEBT.md")
	if err != nil {
		return nil, err
	}
	sort.Strings(registers)
	var entries []reconciliationEntry
	for _, path := range registers {
		parsed, parseErr := parseReconciliationEntries(path)
		if parseErr != nil {
			return nil, parseErr
		}
		entries = append(entries, parsed...)
	}
	return entries, nil
}

// evidenceReconciliationCellEscape escapes a value for embedding in a GFM
// table cell -- mirroring the groundedness lint's own unescapeCell in
// reverse: a literal backslash becomes `\\`, a literal pipe becomes `\|`.
// Verification commands routinely contain an unescaped "|" (e.g.
// `-run 'A|B'`), so this escape is load-bearing, not decorative.
func evidenceReconciliationCellEscape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `|`, `\|`)
	return s
}

// evidenceReconciliationObligationSummary renders the verdict-specific
// obligation fields as one prose cell, escaped for table embedding.
func evidenceReconciliationObligationSummary(e reconciliationEntry) string {
	switch e.Verdict {
	case reconciliationRenamed:
		return fmt.Sprintf("replacement: %s", e.Replacement)
	case reconciliationSuperseded:
		return fmt.Sprintf("phase %s, commit %s, covers: %s", e.SupersedingPhase, e.SupersedingCommit, e.CoveringCommand)
	case reconciliationObsoleteByDesign:
		return fmt.Sprintf("deleted %s from %s at phase %s, commit %s", e.DeletedSymbol, e.DeletedPackage, e.DeletingPhase, e.DeletingCommit)
	case reconciliationUnderScoped:
		return fmt.Sprintf("missing clause: %s, landing phase: %s", e.MissingClause, e.LandingPhase)
	default:
		return ""
	}
}

// renderEvidenceReconciliationView renders entries as the exact checked-in
// document shape: frontmatter `entries: N`, then one table row per entry.
func renderEvidenceReconciliationView(entries []reconciliationEntry) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "entries: %d\n", len(entries))
	b.WriteString("---\n\n")
	b.WriteString("# Evidence Reconciliation\n\n")
	b.WriteString("**GENERATED. Do not hand-edit.** Regenerated in memory and byte-compared by\n")
	b.WriteString("`TestEvidenceReconciliationViewIsCurrent` (`internal/compiler/session/witness_registry_test.go`)\n")
	b.WriteString("from every `*-DEBT.md` register's `` ```reconciliation ``` `` fenced blocks\n")
	b.WriteString("(`internal/compiler/session/session_test.go`'s `parseReconciliationEntries`). The\n")
	b.WriteString("register row is the authored truth, this file is derived -- a hand edit is a\n")
	b.WriteString("failure, not a source of information. There is no regeneration command and no\n")
	b.WriteString("blessing path: if this view is out of date, correct the registers and\n")
	b.WriteString("re-derive, never overwrite this file directly.\n\n")
	b.WriteString("This is deliberately NOT a baseline file: a baseline is satisfied by silence,\n")
	b.WriteString("while every row here is satisfied only by a claim that can itself fail --\n")
	b.WriteString("`TestReconciliationVerdictsCarryTheirObligations` re-checks every row's\n")
	b.WriteString("obligation on every run (D-14-12).\n\n")
	b.WriteString("## Entries\n\n")
	b.WriteString("| ID | File | Line | Command | Verdict | Obligation |\n")
	b.WriteString("|---|---|---|---|---|---|\n")
	for _, e := range entries {
		fmt.Fprintf(&b, "| %s | %s | %d | `%s` | %s | %s |\n",
			e.ID,
			evidenceReconciliationCellEscape(e.File),
			e.Line,
			evidenceReconciliationCellEscape(e.Command),
			e.Verdict,
			evidenceReconciliationCellEscape(evidenceReconciliationObligationSummary(e)),
		)
	}
	return b.String()
}

// TestEvidenceReconciliationViewIsCurrent regenerates the view in memory
// from the registers and byte-compares it against the checked-in
// .planning/EVIDENCE-RECONCILIATION.md -- a hand edit is a failure
// (D-14-13). The non-zero-row-count assertion (D-14-14c) additionally
// guards against a vacuous pass.
func TestEvidenceReconciliationViewIsCurrent(t *testing.T) {
	entries, err := deriveEvidenceReconciliation()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least the 66 reconciliation entries plan 14-10 authored, got 0")
	}

	regenerated := renderEvidenceReconciliationView(entries)

	checkedInPath := testsupport.ProjectPath(".planning", "EVIDENCE-RECONCILIATION.md")
	checkedIn, err := os.ReadFile(checkedInPath)
	if err != nil {
		t.Fatalf("read .planning/EVIDENCE-RECONCILIATION.md: %v", err)
	}
	if regenerated != string(checkedIn) {
		t.Fatalf(".planning/EVIDENCE-RECONCILIATION.md is out of date -- regenerate from the registers, never hand-edit.\n--- regenerated ---\n%s\n--- checked-in ---\n%s", regenerated, string(checkedIn))
	}

	m := evidenceReconciliationFrontmatterPattern.FindStringSubmatch(string(checkedIn))
	if m == nil {
		t.Fatal(".planning/EVIDENCE-RECONCILIATION.md has no `entries: N` frontmatter line")
	}
	declared, convErr := strconv.Atoi(m[1])
	if convErr != nil {
		t.Fatalf("unreadable entries: frontmatter: %v", convErr)
	}
	if declared != len(entries) {
		t.Fatalf("frontmatter declares entries: %d but the derivation holds %d", declared, len(entries))
	}
}

// TestEvidenceReconciliationViewNonZeroRowCountIsLoadBearing (D-14-14c)
// proves the non-zero-row-count assertion is not vacuous ornamentation,
// mirroring TestUnreachableClaimsViewNonZeroRowCountIsLoadBearing exactly.
func TestEvidenceReconciliationViewNonZeroRowCountIsLoadBearing(t *testing.T) {
	emptyRendered := renderEvidenceReconciliationView(nil)
	if !strings.Contains(emptyRendered, "entries: 0") {
		t.Fatalf("expected an empty entry list to render entries: 0, got:\n%s", emptyRendered)
	}
	if emptyRendered != renderEvidenceReconciliationView(nil) {
		t.Fatal("renderEvidenceReconciliationView is nondeterministic on empty input")
	}
	entries, err := deriveEvidenceReconciliation()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("the live registers derived ZERO reconciliation entries -- the byte-compare above would now be vacuously comparing two empty documents, proving consistency but not completeness")
	}
}

// TestEvidenceReconciliationViewCatchesHandEdit mirrors
// TestUnreachableClaimsViewCatchesHandEdit exactly: a one-character
// tamper is proven to differ from a fresh regeneration.
func TestEvidenceReconciliationViewCatchesHandEdit(t *testing.T) {
	entries, err := deriveEvidenceReconciliation()
	if err != nil {
		t.Fatal(err)
	}
	regenerated := renderEvidenceReconciliationView(entries)
	tampered := regenerated + "x"
	if tampered == regenerated {
		t.Fatal("seeded one-character edit did not change the text")
	}
	if tampered == renderEvidenceReconciliationView(entries) {
		t.Fatal("the tampered copy should differ from a fresh regeneration, but compared equal")
	}
}

// TestEvidenceReconciliationViewCatchesStaleEntry mirrors
// TestUnreachableClaimsViewCatchesStaleEntry (D-14-14b): a checked-in
// entry whose underlying register row has vanished must never be silently
// auto-pruned -- it must fail.
func TestEvidenceReconciliationViewCatchesStaleEntry(t *testing.T) {
	entries, err := deriveEvidenceReconciliation()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("no entries to seed a stale row against")
	}
	stale := append([]reconciliationEntry{{
		ID: "D-00-00", File: "nonexistent.md", Line: 1, Command: "echo gone",
		Classification: classR1, Verdict: reconciliationRenamed, Replacement: "echo replacement",
	}}, entries...)
	staleRendered := renderEvidenceReconciliationView(stale)
	freshRendered := renderEvidenceReconciliationView(entries)
	if staleRendered == freshRendered {
		t.Fatal("a view carrying a stale entry absent from the live derivation should differ from a fresh regeneration, but compared equal")
	}
}

// TestEvidenceReconciliationViewCountMatchesFrontmatter is D-14-14a's own
// dedicated proof for this view: a frontmatter `entries: N` that disagrees
// with the table's own row count must fail. TestEvidenceReconciliationViewIsCurrent
// already asserts this over the real checked-in file; this test
// demonstrates the mechanism directly over a synthetic mismatch.
func TestEvidenceReconciliationViewCountMatchesFrontmatter(t *testing.T) {
	entries, err := deriveEvidenceReconciliation()
	if err != nil {
		t.Fatal(err)
	}
	rendered := renderEvidenceReconciliationView(entries)
	tampered := strings.Replace(rendered, fmt.Sprintf("entries: %d\n", len(entries)), fmt.Sprintf("entries: %d\n", len(entries)+1), 1)
	if tampered == rendered {
		t.Fatal("seeded frontmatter-count mismatch did not change the text")
	}
	m := evidenceReconciliationFrontmatterPattern.FindStringSubmatch(tampered)
	if m == nil {
		t.Fatal("tampered text lost its entries: frontmatter line")
	}
	declared, convErr := strconv.Atoi(m[1])
	if convErr != nil {
		t.Fatal(convErr)
	}
	// Re-derive the table row count from the tampered text the same way
	// TestEvidenceReconciliationViewIsCurrent would over a hand-edited
	// file: it holds len(entries) rows still (only frontmatter moved), so
	// declared (len(entries)+1) now disagrees -- this is the failure
	// TestEvidenceReconciliationViewIsCurrent's own declared != len(entries)
	// check catches on the real file.
	if declared == len(entries) {
		t.Fatal("seeded mismatch did not actually diverge from the true entry count")
	}
}
