package session_test

import (
	"context"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/measure"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/protocol"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// TestVerifyPhase6ControlsAndWork runs Phase 6's own control-and-work gate
// end-to-end and asserts every required control fires with nonzero
// RecomputedWork, mirroring TestVerifyPhase5ControlsAndWork's own shape.
func TestVerifyPhase6ControlsAndWork(t *testing.T) {
	result, err := session.VerifyPhase6ControlsAndWork(context.Background())
	if err != nil {
		t.Fatalf("VerifyPhase6ControlsAndWork returned an error: %v", err)
	}
	if result.Status != protocol.StatusPass {
		foundTerminalLane := false
		for _, lane := range result.Lanes {
			if lane.ID == "lane:defect-cleanup-injection" && lane.Status == protocol.StatusOperational {
				foundTerminalLane = true
			}
		}
		if !foundTerminalLane {
			t.Fatalf("VerifyPhase6ControlsAndWork failed outside the terminal cleanup lane: status=%s lanes=%+v diagnostics=%+v", result.Status, result.Lanes, result.Diagnostics)
		}
		path := testsupport.ProjectPath("testdata", "phase4", "acquire_three_success.lang")
		_, _, refusal := session.RunNativeFile(context.Background(), path, native.DefaultRunner())
		requirePhase16M004Refusal(t, refusal, "foreign")
		return
	}
	for _, required := range session.Phase6RequiredControls() {
		found := false
		for _, lane := range result.Lanes {
			for _, control := range lane.Controls {
				if control == required {
					found = true
				}
			}
		}
		if !found {
			t.Fatalf("VerifyPhase6ControlsAndWork never fired required control %s", required)
		}
	}
	for _, lane := range result.Lanes {
		if lane.RecomputedWork == 0 {
			t.Fatalf("lane %s reported zero RecomputedWork", lane.ID)
		}
	}
	for _, escape := range session.Phase6ExpectedEscapes() {
		found := false
		for _, declared := range result.ExpectedEscapes {
			if declared == escape {
				found = true
			}
		}
		if !found {
			t.Fatalf("VerifyPhase6ControlsAndWork's ExpectedEscapes is missing declared escape %s", escape)
		}
	}
}

// phase6VerifierScriptText reads scripts/verify-phase6.sh's own text,
// mirroring phase5VerifierScriptText's precedent exactly.
func phase6VerifierScriptText(t testing.TB) string {
	t.Helper()
	data, err := os.ReadFile(testsupport.ProjectPath("scripts", "verify-phase6.sh"))
	if err != nil {
		t.Fatalf("read scripts/verify-phase6.sh: %v", err)
	}
	return string(data)
}

// phase6OwnControlIdentifiers extracts the exact set of `control:` tokens
// from the script's OWN required-control block (the `for control in ...`
// loop following the "Phase 6's own required-control set" comment),
// mirroring phase5OwnControlIdentifiers's anchor-plus-block-extraction
// technique exactly.
func phase6OwnControlIdentifiers(t testing.TB, text string) []string {
	t.Helper()
	anchor := "Phase 6's own required-control set"
	anchorIndex := strings.Index(text, anchor)
	if anchorIndex == -1 {
		t.Fatalf("scripts/verify-phase6.sh is missing its own required-control-set comment anchor")
	}
	rest := text[anchorIndex:]
	doneIndex := strings.Index(rest, "\ndone")
	if doneIndex == -1 {
		t.Fatalf("scripts/verify-phase6.sh's own required-control block has no closing done")
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

// TestPhase6RequiredControlsMatchScript asserts set equality, in both
// directions, between session.Phase6RequiredControls() and the
// identifiers named in the script's own required-control block.
func TestPhase6RequiredControlsMatchScript(t *testing.T) {
	text := phase6VerifierScriptText(t)
	fromScript := phase6OwnControlIdentifiers(t, text)
	scriptSet := make(map[string]bool, len(fromScript))
	for _, control := range fromScript {
		scriptSet[control] = true
	}
	sessionSet := make(map[string]bool, len(session.Phase6RequiredControls()))
	for _, control := range session.Phase6RequiredControls() {
		sessionSet[control] = true
	}
	for control := range sessionSet {
		if !scriptSet[control] {
			t.Fatalf("session.Phase6RequiredControls() names %s, which the script's own required-control block does not", control)
		}
	}
	for control := range scriptSet {
		if !sessionSet[control] {
			t.Fatalf("the script's required-control block names %s, which session.Phase6RequiredControls() does not", control)
		}
	}
	if len(scriptSet) != len(sessionSet) {
		t.Fatalf("script control set (%d) and session control set (%d) differ in size: script=%v session=%v", len(scriptSet), len(sessionSet), fromScript, session.Phase6RequiredControls())
	}
}

// phase6ScriptShellVarValue extracts the value of a `name=value` shell
// assignment from text, mirroring phase5ScriptShellVarValue exactly.
func phase6ScriptShellVarValue(t testing.TB, text, name string) string {
	t.Helper()
	anchor := name + "="
	index := strings.Index(text, anchor)
	if index == -1 {
		t.Fatalf("scripts/verify-phase6.sh is missing shell variable %s", name)
	}
	rest := text[index+len(anchor):]
	end := strings.IndexAny(rest, "\n")
	if end == -1 {
		end = len(rest)
	}
	return strings.TrimSpace(rest[:end])
}

// TestPhase6BoundsMatchScript asserts the script's own duplicated
// sample-count, CoV-threshold, and explain/query bound values equal their
// Go constants -- the same "the script and Go cannot drift apart"
// discipline TestPhase5CorpusBoundMatchesScript established.
func TestPhase6BoundsMatchScript(t *testing.T) {
	text := phase6VerifierScriptText(t)
	intCases := []struct {
		name string
		want int
	}{
		{"phase6_warm_sample_count", measure.WarmSampleCount},
		{"phase6_explain_default_depth", protocol.ExplainDefaultDepth},
		{"phase6_explain_max_nodes", protocol.ExplainMaxNodes},
		{"phase6_query_max_facts_per_page", protocol.QueryMaxFactsPerPage},
	}
	for _, testCase := range intCases {
		got := phase6ScriptShellVarValue(t, text, testCase.name)
		want := strconv.Itoa(testCase.want)
		if got != want {
			t.Fatalf("script's %s=%s does not match Go constant %s", testCase.name, got, want)
		}
	}
	gotCoV := phase6ScriptShellVarValue(t, text, "phase6_cov_demotion_threshold")
	wantCoV := strconv.FormatFloat(measure.CoVDemotionThreshold, 'g', -1, 64)
	if gotCoV != wantCoV {
		t.Fatalf("script's phase6_cov_demotion_threshold=%s does not match measure.CoVDemotionThreshold %s", gotCoV, wantCoV)
	}
}

// TestPhase6VerifierScriptContract asserts the script names every one of
// Phase 6's required controls, invokes every phase's own corpus, still
// pins the sanitizer options, and leaves every prior phase's own gate
// script untouched -- mirroring TestPhase5VerifierScriptContract's shape.
func TestPhase6VerifierScriptContract(t *testing.T) {
	text := phase6VerifierScriptText(t)
	for _, control := range session.Phase6RequiredControls() {
		if !strings.Contains(text, control) {
			t.Fatalf("scripts/verify-phase6.sh's text is missing required control %s", control)
		}
	}
	for phase := 1; phase <= 6; phase++ {
		want := "json verify testdata/phase" + strconv.Itoa(phase)
		if !strings.Contains(text, want) {
			t.Fatalf("scripts/verify-phase6.sh is missing invocation %q", want)
		}
	}
	if !strings.Contains(text, "ASAN_OPTIONS=") {
		t.Fatal("scripts/verify-phase6.sh does not pin ASAN_OPTIONS")
	}
	if !strings.Contains(text, "UBSAN_OPTIONS=") {
		t.Fatal("scripts/verify-phase6.sh does not pin UBSAN_OPTIONS")
	}
	phase5Path := testsupport.ProjectPath("scripts", "verify-phase5.sh")
	if _, err := os.Stat(phase5Path); err != nil {
		t.Fatalf("scripts/verify-phase5.sh must still exist unchanged: %v", err)
	}
}

// TestCIWorkflowRunsCurrentAggregateGate pins the historical Phase 6 baseline
// and the current focused evidence aggregate without recursively executing the
// workflow: the baseline itself runs `go test ./...`.
func TestCIWorkflowRunsCurrentAggregateGate(t *testing.T) {
	path := testsupport.ProjectPath(".github", "workflows", "ci.yml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read .github/workflows/ci.yml: %v", err)
	}
	text := string(raw)

	for _, required := range []string{
		"checks:",
		"evidence-aggregate:",
		"current evidence aggregate",
		"sh scripts/verify-phase6.sh",
		"ubuntu-latest",
		"macos-latest",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf(".github/workflows/ci.yml is missing current aggregate requirement %q", required)
		}
	}

	for _, required := range []struct {
		pkg      string
		testName string
	}{
		{"./internal/compiler/native", "TestDecodeExecutionSchema2AdmissionSeam"},
		{"./internal/compiler/session", "TestSchema2ComparisonRequiresPeerVerdict"},
		{"./internal/compiler/session", "TestPhase5CompareProgramEnginesPreservesLegacySchemas"},
		{"./internal/compiler/session", "TestPhase11InterproceduralDifferential/DiamondSharedLeaf"},
		{"./internal/compiler/session", "TestPhase15CollisionGuardIsNotInert"},
	} {
		if !workflowRunsTestInPackage(text, required.pkg, required.testName) {
			t.Fatalf(".github/workflows/ci.yml must select %s from %s", required.testName, required.pkg)
		}
	}
}

func TestCIWorkflowSelectionPinsPackageOwnership(t *testing.T) {
	tests := []struct {
		name     string
		workflow string
		pkg      string
		testName string
		want     bool
	}{
		{
			name:     "native decoder selected from owning package",
			workflow: "run: go test ./internal/compiler/native -run 'TestDecodeExecutionSchema2AdmissionSeam' -count=1 -v",
			pkg:      "./internal/compiler/native",
			testName: "TestDecodeExecutionSchema2AdmissionSeam",
			want:     true,
		},
		{
			name:     "native decoder rejected from session package",
			workflow: "run: go test ./internal/compiler/session -run 'TestDecodeExecutionSchema2AdmissionSeam' -count=1 -v",
			pkg:      "./internal/compiler/native",
			testName: "TestDecodeExecutionSchema2AdmissionSeam",
			want:     false,
		},
		{
			name:     "session seam selected from owning package",
			workflow: "run: go test ./internal/compiler/session -run 'TestSchema2ComparisonRequiresPeerVerdict' -count=1 -v",
			pkg:      "./internal/compiler/session",
			testName: "TestSchema2ComparisonRequiresPeerVerdict",
			want:     true,
		},
		{
			name:     "missing selection rejected",
			workflow: "run: go test ./internal/compiler/native -run 'TestOther' -count=1 -v",
			pkg:      "./internal/compiler/native",
			testName: "TestDecodeExecutionSchema2AdmissionSeam",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := workflowRunsTestInPackage(tt.workflow, tt.pkg, tt.testName); got != tt.want {
				t.Fatalf("workflowRunsTestInPackage(%q, %q) = %t, want %t", tt.pkg, tt.testName, got, tt.want)
			}
		})
	}
}

// workflowRunsTestInPackage reports whether one focused go test invocation
// couples a test selector with the package that owns the test.
func workflowRunsTestInPackage(workflow, pkg, testName string) bool {
	for _, line := range strings.Split(workflow, "\n") {
		start := strings.Index(line, "go test ")
		if start < 0 {
			continue
		}
		args := strings.Fields(line[start:])
		if len(args) < 4 || args[0] != "go" || args[1] != "test" || args[2] != pkg {
			continue
		}
		for i, arg := range args[3:] {
			selector := ""
			switch {
			case arg == "-run" && i+4 < len(args):
				selector = args[i+4]
			case strings.HasPrefix(arg, "-run="):
				selector = strings.TrimPrefix(arg, "-run=")
			default:
				continue
			}
			selector = strings.Trim(selector, "'\"")
			if strings.Contains(selector, testName) {
				return true
			}
		}
	}
	return false
}

// TestPhase6ScriptInvokesNoPriorGate asserts the script never references
// any prior phase's own gate script by path -- a phase gate is a peer,
// never a descendant (D-06-13's carried-forward standing rule).
func TestPhase6ScriptInvokesNoPriorGate(t *testing.T) {
	text := phase6VerifierScriptText(t)
	for _, forbidden := range []string{"verify-phase1.sh", "verify-phase2.sh", "verify-phase3.sh", "verify-phase4.sh", "verify-phase5.sh"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("scripts/verify-phase6.sh must never invoke a previous-phase gate script, but its text contains %q", forbidden)
		}
	}
}

// TestPhase6SamplingLoopMatchesGoStatistics feeds a fixed 20-element
// sample set through measure.Samples.Summary() and asserts its p50/p95
// match the exact sorted-index rule scripts/verify-phase2.sh's own
// sed-based extraction implements (`sed -n '10p'`/`sed -n '19p'` over a
// 1-indexed sorted stream) -- the same rule scripts/verify-phase6.sh's
// own `stats` seam delegates to, so the shell side never reimplements
// percentile indexing itself.
func TestPhase6SamplingLoopMatchesGoStatistics(t *testing.T) {
	samples := measure.Samples{
		100, 105, 98, 110, 102, 99, 101, 103, 97, 104,
		106, 100, 108, 95, 109, 101, 103, 100, 107, 102,
	}
	summary, err := samples.Summary()
	if err != nil {
		t.Fatalf("Summary() returned an error: %v", err)
	}

	sorted := append(measure.Samples{}, samples...)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j] < sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	// sed -n '10p'/'19p' over a 1-indexed sorted stream is 0-indexed
	// position 9 (p50) and 18 (p95) for a 20-element set.
	wantP50 := sorted[9]
	wantP95 := sorted[18]
	if summary.P50 != wantP50 {
		t.Fatalf("Summary().P50 = %d, want %d (script's sed -n '10p' equivalent)", summary.P50, wantP50)
	}
	if summary.P95 != wantP95 {
		t.Fatalf("Summary().P95 = %d, want %d (script's sed -n '19p' equivalent)", summary.P95, wantP95)
	}
	if summary.CoV < 0 {
		t.Fatalf("Summary().CoV = %v, want a non-negative coefficient of variation", summary.CoV)
	}
}
