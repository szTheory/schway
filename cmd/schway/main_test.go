package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// TestPhase6CorpusDispatchRequiresMarker asserts isPhase6Corpus recognizes
// testdata/phase6 by its own characteristic marker fixture
// (heldout_match_defect.schway, per testdata/phase6/README) and returns
// false for testdata/phase5 and for an empty temp directory -- a directory
// with no Phase 6 marker must never be silently treated as a Phase 6
// corpus (FND-04's empty-input edge, plan 06-15's acceptance criteria).
func TestPhase6CorpusDispatchRequiresMarker(t *testing.T) {
	if !isPhase6Corpus("../../testdata/phase6") {
		t.Fatal("isPhase6Corpus(testdata/phase6) = false, want true")
	}
	if isPhase6Corpus("../../testdata/phase5") {
		t.Fatal("isPhase6Corpus(testdata/phase5) = true, want false")
	}
	empty := t.TempDir()
	if isPhase6Corpus(empty) {
		t.Fatal("isPhase6Corpus(empty directory) = true, want false")
	}
	if isPhase5Corpus(empty) {
		t.Fatal("isPhase5Corpus(empty directory) = true, want false")
	}
	if _, err := os.Stat("../../testdata/phase6/heldout_match_defect.schway"); err != nil {
		t.Fatalf("testdata/phase6's own marker fixture is missing: %v", err)
	}
}

// TestPhase7CorpusDispatchRequiresMarker asserts isPhase7Corpus recognizes
// testdata/phase07 by its own characteristic marker fixture
// (call_basic.schway, the Phase 07 tracer fixture) and returns false for
// testdata/phase6 and for an empty temp directory -- a directory with no
// Phase 07 marker must never be silently treated as a Phase 07 corpus.
func TestPhase7CorpusDispatchRequiresMarker(t *testing.T) {
	if !isPhase7Corpus("../../testdata/phase07") {
		t.Fatal("isPhase7Corpus(testdata/phase07) = false, want true")
	}
	if isPhase7Corpus("../../testdata/phase6") {
		t.Fatal("isPhase7Corpus(testdata/phase6) = true, want false")
	}
	empty := t.TempDir()
	if isPhase7Corpus(empty) {
		t.Fatal("isPhase7Corpus(empty directory) = true, want false")
	}
	if _, err := os.Stat("../../testdata/phase07/call_basic.schway"); err != nil {
		t.Fatalf("testdata/phase07's own marker fixture is missing: %v", err)
	}
}

func TestPhase22IdentityApplicationBuildAndRunCLI(t *testing.T) {
	type expectedCase struct {
		Input  string `json:"input"`
		Stdout string `json:"stdout"`
	}
	var expected struct {
		Schema string         `json:"schema"`
		Cases  []expectedCase `json:"cases"`
	}
	data, err := os.ReadFile(testsupport.ProjectPath("examples", "phase22", "identity.expected.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &expected); err != nil {
		t.Fatal(err)
	}
	if expected.Schema != "schway.phase22-identity-expected/1" || len(expected.Cases) != 2 {
		t.Fatalf("independent expected-output fixture has wrong identity or cases: %+v", expected)
	}
	source := testsupport.ProjectPath("examples", "phase22", "identity.schway")
	artifact := filepath.Join(t.TempDir(), "identity")
	code, stdout, stderr := captureLangRun(t, []string{"build", source, "--output", artifact})
	if code != 0 || stdout != "built "+artifact+"\n" || stderr != "" {
		t.Fatalf("build code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	for _, test := range expected.Cases {
		t.Run(test.Input, func(t *testing.T) {
			code, stdout, stderr := captureLangRun(t, []string{"app", "run", artifact, "--", test.Input})
			if code != 0 || stdout != test.Stdout || stderr != "" {
				t.Fatalf("app run input=%q code=%d stdout=%q want=%q stderr=%q", test.Input, code, stdout, test.Stdout, stderr)
			}
		})
	}

	for _, input := range []string{"", "+7", "-7", "letters", "18446744073709551616", strings.Repeat("1", 21)} {
		t.Run("reject/"+input, func(t *testing.T) {
			code, stdout, _ := captureLangRun(t, []string{"app", "run", artifact, "--", input})
			if code != 65 || stdout != "" {
				t.Fatalf("malformed input=%q code=%d stdout=%q, want exit 65 before output", input, code, stdout)
			}
		})
	}
	code, stdout, _ = captureLangRun(t, []string{"app", "run", artifact, "--", strings.Repeat("1", native.MaxApplicationArgumentBytes+1)})
	if code != 65 || stdout != "" {
		t.Fatalf("over-transport input code=%d stdout=%q, want exit 65 before launch", code, stdout)
	}
	code, _, _ = captureLangRun(t, []string{"app", "run", artifact, "--", "7", "extra"})
	if code != 64 {
		t.Fatalf("extra app-run token exit=%d, want usage 64", code)
	}
}

func TestPhase23PublicFileByte(t *testing.T) {
	source := testsupport.ProjectPath("examples", "phase23", "file_byte.schway")
	manifest := testsupport.ProjectPath("examples", "phase23", "file_byte.bindings.json")
	artifact := filepath.Join(t.TempDir(), "file-byte")
	code, stdout, stderr := captureLangRun(t, []string{"build", source, "--manifest", manifest, "--output", artifact})
	if code != 0 || stdout != "built "+artifact+"\n" || stderr != "" {
		t.Fatalf("build code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	var expected struct {
		Schema string `json:"schema"`
		Cases  []struct {
			ID           string `json:"id"`
			FileBytesHex string `json:"file_bytes_hex"`
			ExitCode     int    `json:"exit_code"`
			Stdout       string `json:"stdout"`
			Stderr       string `json:"stderr"`
		} `json:"cases"`
	}
	expectedPath := testsupport.ProjectPath("examples", "phase23", "file_byte.expected.json")
	data, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &expected); err != nil {
		t.Fatal(err)
	}
	if expected.Schema != "schway.phase23-file-byte-expected/1" || len(expected.Cases) != 5 {
		t.Fatalf("independent file-byte answer fixture has wrong identity or cases: %+v", expected)
	}
	for _, test := range expected.Cases {
		t.Run(test.ID, func(t *testing.T) {
			fileBytes, err := hex.DecodeString(test.FileBytesHex)
			if err != nil {
				t.Fatalf("invalid independent file bytes %q: %v", test.FileBytesHex, err)
			}
			input := filepath.Join(t.TempDir(), "input.bin")
			if err := os.WriteFile(input, fileBytes, 0o600); err != nil {
				t.Fatal(err)
			}
			code, stdout, stderr := captureLangRun(t, []string{"app", "run", artifact, "--", input})
			if code != test.ExitCode || stdout != test.Stdout || stderr != test.Stderr || len(stderr) > 128 {
				t.Fatalf("public file-byte case %s: code=%d stdout=%q stderr=%q; want code=%d stdout=%q stderr=%q (diagnostic limit 128 bytes)", test.ID, code, stdout, stderr, test.ExitCode, test.Stdout, test.Stderr)
			}
		})
	}
	for _, test := range []struct {
		name       string
		input      string
		wantCode   int
		wantStderr string
	}{
		{name: "empty token", input: "", wantCode: 65},
		{name: "one byte reaches adapter", input: "/", wantCode: 65, wantStderr: "schway_file_byte_acquire: NotRegular\n"},
		{name: "4096 bytes reaches adapter", input: strings.Repeat("p", 4096), wantCode: 65, wantStderr: "schway_file_byte_acquire: OpenFailed\n"},
		{name: "4097 bytes rejected by runner", input: strings.Repeat("p", 4097), wantCode: 65, wantStderr: fmt.Sprintf("schway app run: native.input_too_long: argument is %d bytes; limit is %d\n", 4097, native.MaxApplicationArgumentBytes)},
		{name: "embedded NUL rejected by runner", input: "safe\x00tail", wantCode: 65, wantStderr: "schway app run: native.input_contains_nul: argument contains an embedded NUL byte\n"},
	} {
		t.Run("path-boundary/"+test.name, func(t *testing.T) {
			code, stdout, stderr := captureLangRun(t, []string{"app", "run", artifact, "--", test.input})
			if code != test.wantCode || stdout != "" || stderr != test.wantStderr || len(stderr) > 128 {
				t.Fatalf("path=%q code=%d stdout=%q stderr=%q; want code=%d empty stdout bounded stderr=%q", test.input, code, stdout, stderr, test.wantCode, test.wantStderr)
			}
		})
	}
	t.Run("acquire-error/non-regular", func(t *testing.T) {
		input := t.TempDir()
		code, stdout, stderr := captureLangRun(t, []string{"app", "run", artifact, "--", input})
		if code != 65 || stdout != "" || stderr != "schway_file_byte_acquire: NotRegular\n" {
			t.Fatalf("directory code=%d stdout=%q stderr=%q", code, stdout, stderr)
		}
	})
	t.Run("acquire-error/open-failed", func(t *testing.T) {
		input := filepath.Join(t.TempDir(), "missing.bin")
		code, stdout, stderr := captureLangRun(t, []string{"app", "run", artifact, "--", input})
		if code != 65 || stdout != "" || stderr != "schway_file_byte_acquire: OpenFailed\n" {
			t.Fatalf("missing file code=%d stdout=%q stderr=%q", code, stdout, stderr)
		}
	})
	t.Run("acquire-error/unreadable", func(t *testing.T) {
		input := filepath.Join(t.TempDir(), "unreadable.bin")
		if err := os.WriteFile(input, []byte{0x41}, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(input, 0); err != nil {
			t.Fatal(err)
		}
		probe, err := os.Open(input)
		if err == nil {
			probe.Close()
			t.Log("this host permits opening mode-000 files; open failure is injected at the native adapter seam in TestPhase23AcquireFailuresInitializeAndFreePartialAllocations")
			return
		}
		code, stdout, stderr := captureLangRun(t, []string{"app", "run", artifact, "--", input})
		if code != 65 || stdout != "" || stderr != "schway_file_byte_acquire: OpenFailed\n" {
			t.Fatalf("unreadable file code=%d stdout=%q stderr=%q", code, stdout, stderr)
		}
	})

	identity := testsupport.ProjectPath("examples", "phase22", "identity.schway")
	identityArtifact := filepath.Join(t.TempDir(), "identity")
	code, stdout, stderr = captureLangRun(t, []string{"build", identity, "--output", identityArtifact})
	if code != 0 || stdout != "built "+identityArtifact+"\n" || stderr != "" {
		t.Fatalf("Phase 22 build code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	code, stdout, stderr = captureLangRun(t, []string{"app", "run", identityArtifact, "--", "7"})
	if code != 0 || stdout != "7\n" || stderr != "" {
		t.Fatalf("Phase 22 run code=%d stdout=%q stderr=%q, want 7", code, stdout, stderr)
	}
}

func TestPhase23ReadmeContract(t *testing.T) {
	readmePath := testsupport.ProjectPath("examples", "phase23", "README.md")
	readme, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(readme)
	contractText := strings.Join(strings.Fields(text), " ")
	commands := normalizePhase23ReadmeCommands(text)
	for name, command := range map[string]string{
		"clean-checkout CLI build":   "go build -o ./schway ./cmd/schway",
		"native application build":   "./schway build examples/phase23/file_byte.schway --manifest examples/phase23/file_byte.bindings.json --output ./file-byte",
		"0x41 success":               "./schway app run ./file-byte -- ./byte-41.bin",
		"0x42 success":               "./schway app run ./file-byte -- ./byte-42.bin",
		"empty acquisition error":    "./schway app run ./file-byte -- ./empty.bin",
		"two-byte acquisition error": "./schway app run ./file-byte -- ./two-bytes.bin",
		"0x43 use error":             "./schway app run ./file-byte -- ./byte-43.bin",
	} {
		if !strings.Contains(commands, command) {
			t.Errorf("command form: missing %s example", name)
		}
	}
	categories := map[string][]string{
		"bounded raw-byte input": {"caller-created raw-byte files", "exactly one raw file byte", "probes for EOF", "4,096-byte"},
		"independent answers":    {"file_byte.expected.json", "0x41 → 65", "0x42 → 66", "not values copied from compiler events"},
		"failure provenance":     {"schway_file_byte_acquire: EmptyFile", "schway_file_byte_acquire: FileTooLong", "schway_file_byte_use: UnsupportedByte", "performs the matching release, and only then reports", "bounded to 128 bytes"},
		"evidence scope":         {"independent native observer", "exact binary and host exercised", "Interpreter receipts are", "they do not perform host IO or prove physical cleanup"},
		"Phase 22 compatibility": {"Phase 22 U64 application example", "still prints `7`"},
	}
	for category, clauses := range categories {
		if diagnostic := phase23ReadmeCategoryDiagnostic(contractText, category, clauses); diagnostic != "" {
			t.Error(diagnostic)
		}
		mutated := strings.ReplaceAll(contractText, clauses[len(clauses)-1], "")
		if diagnostic := phase23ReadmeCategoryDiagnostic(mutated, category, clauses); !strings.Contains(diagnostic, category) {
			t.Errorf("negative control for %q did not fail by category: %q", category, diagnostic)
		}
	}
	if native.MaxApplicationArgumentBytes != 4096 {
		t.Fatalf("Phase 23 path-token limit changed to %d; update the README contract", native.MaxApplicationArgumentBytes)
	}
}

func normalizePhase23ReadmeCommands(text string) string {
	text = strings.ReplaceAll(text, "\\\n", " ")
	return strings.Join(strings.Fields(text), " ")
}

func phase23ReadmeCategoryDiagnostic(text, category string, clauses []string) string {
	for _, clause := range clauses {
		if !strings.Contains(text, clause) {
			return "contract category " + category + ": missing " + clause
		}
	}
	return ""
}

func TestPhase22READMEContract(t *testing.T) {
	readme, err := os.ReadFile(testsupport.ProjectPath("examples", "phase22", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(readme)
	contractText := strings.Join(strings.Fields(text), " ")
	commands := normalizePhase22READMECommands(text)
	for name, command := range map[string]string{
		"build":             "schway build examples/phase22/identity.schway --output ./identity",
		"run":               "schway app run ./identity -- 7",
		"manifest build":    "schway build examples/phase22/identity.schway --manifest examples/phase22/identity.bindings.json --output ./identity",
		"same-run evidence": "schway app run ./identity --report ./identity-evidence.json --evidence=events -- 7",
		"explicit replay":   "schway app verify examples/phase22/identity.schway --cases examples/phase22/identity.cases.json --report ./identity-verification.json",
	} {
		if !strings.Contains(commands, command) {
			t.Errorf("command form: missing runnable %s example", name)
		}
	}
	categories := map[string][]string{
		"input and process":               {"Canonical decimal `U64`", "4,096-byte", "one retained application process", "30-second", "without an added output cap", "exit, signal, timeout, and launch"},
		"evidence and conformance bounds": {"64 KiB", "16 MiB", "capacity_exhausted", "write errors fail", "`verified: false`"},
		"local C boundary":                {"64 KiB", "4 MiB", "16 MiB", "fixed C17 flags", "not a sandbox for hostile C", "authority does not admit Lang foreign calls or pointers"},
		"host closure":                    {"closure remains `incomplete` on macOS and Linux", "not cacheable"},
		"replay scope":                    {"independently authored expected `U64`", "explicit empty `foreign_outcomes`", "does not launch the retained application", "`verifier_model_only`", "`actual_host_io` and `physical_cleanup` to `false`", "does not establish host IO or physical resource cleanup"},
	}
	for category, clauses := range categories {
		if diagnostic := phase22READMECategoryDiagnostic(contractText, category, clauses); diagnostic != "" {
			t.Error(diagnostic)
		}
		// Each in-memory omission must be reached and diagnosed under its own category.
		mutated := strings.ReplaceAll(contractText, clauses[len(clauses)-1], "")
		if diagnostic := phase22READMECategoryDiagnostic(mutated, category, clauses); !strings.Contains(diagnostic, category) {
			t.Errorf("negative control for %q did not fail by category: %q", category, diagnostic)
		}
	}
	if native.MaxApplicationArgumentBytes != 4096 || execution.MaxApplicationEvidenceBytes != 64*1024 || execution.MaxDocumentBytes != 16*1024*1024 || session.MaxReplayCasesBytes != 64*1024 {
		t.Fatal("Phase 22 exported limits changed; update the README contract expectations")
	}
	for _, documentedLimit := range []string{
		fmt.Sprintf("%s-byte", phase22CommaInt(native.MaxApplicationArgumentBytes)),
		fmt.Sprintf("%d KiB", execution.MaxApplicationEvidenceBytes/1024),
		fmt.Sprintf("%d MiB", execution.MaxDocumentBytes/(1024*1024)),
		fmt.Sprintf("%d KiB", session.MaxReplayCasesBytes/1024),
	} {
		if !strings.Contains(text, documentedLimit) {
			t.Errorf("README does not document exported limit %q", documentedLimit)
		}
	}
	// This is a documentation-contract check. Runtime behavior is established by
	// TestPhase22BindingsRejectInvalidInputs, TestPhase22RunApplicationPreservesStreamsAndProcessOutcomes,
	// TestPhase22EvidenceDisabledCompleteAndStreamIsolation, and TestPhase22AppVerifyIndependentIdentityCases.
}

func phase22CommaInt(value int) string {
	formatted := fmt.Sprintf("%d", value)
	if len(formatted) > 3 {
		return formatted[:len(formatted)-3] + "," + formatted[len(formatted)-3:]
	}
	return formatted
}

func phase22READMECategoryDiagnostic(text, category string, clauses []string) string {
	for _, clause := range clauses {
		if !strings.Contains(text, clause) {
			return "contract category " + category + ": missing " + clause
		}
	}
	return ""
}

func normalizePhase22READMECommands(text string) string {
	text = strings.ReplaceAll(text, "\\\n", " ")
	return strings.Join(strings.Fields(text), " ")
}

func TestPhase22AppRunKeepsOpaqueTokenStreamsAndChildStatus(t *testing.T) {
	artifact := writeCLIPhase22Script(t, "printf 'stdout:%s {not protocol}\\n' \"$1\"\nprintf 'stderr:%s\\n' \"$1\" >&2\nexit 19\n")
	code, stdout, stderr := captureLangRun(t, []string{"app", "run", artifact, "--", "--json"})
	if code != 19 || stdout != "stdout:--json {not protocol}\n" || stderr != "stderr:--json\n" {
		t.Fatalf("opaque app child code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
}

func TestPhase22ConformanceRunStillEmitsExecutionDocument(t *testing.T) {
	fixture := testsupport.ProjectPath("testdata", "phase19", "literal_tracer.schway")
	code, stdout, stderr := captureLangRun(t, []string{"--json", "run", "--engine=native", fixture})
	if code != 0 || stderr != "" || !strings.Contains(stdout, `"schema":"schway.execution/2"`) {
		t.Fatalf("conformance run code=%d stdout=%q stderr=%q, want its existing execution document", code, stdout, stderr)
	}
}

func TestPhase22AppVerifyIndependentIdentityCases(t *testing.T) {
	source := testsupport.ProjectPath("examples", "phase22", "identity.schway")
	cases := testsupport.ProjectPath("examples", "phase22", "identity.cases.json")
	reportPath := filepath.Join(t.TempDir(), "identity-verification.json")
	code, stdout, stderr := captureLangRun(t, []string{"app", "verify", source, "--cases", cases, "--report", reportPath})
	report := readCLIReplayReport(t, reportPath)
	if code != exitSuccess || stdout != "verified 2 replay cases\n" || stderr != "" {
		t.Fatalf("identity verification code=%d stdout=%q stderr=%q report=%+v", code, stdout, stderr, report)
	}
	if report.Schema != session.ApplicationVerifySchema || !report.Verified || report.Status != session.ReplayStatusPass {
		t.Fatalf("identity verification report has wrong verdict: %+v", report)
	}
	if report.DependencyClosure != "incomplete" || report.Cacheable || report.SourceDigest == "" || report.BuildID == "" || report.InputSetID == "" || report.EmittedCDigest == "" {
		t.Fatalf("verification report overstates provenance or omits identity: %+v", report)
	}
	if report.ActualHostIO || report.PhysicalCleanup || len(report.Cases) != 2 {
		t.Fatalf("verification report has wrong scope or case count: %+v", report)
	}
	for index, want := range []string{"7", "42"} {
		caseReport := report.Cases[index]
		if caseReport.Status != session.ReplayStatusPass || caseReport.InputID == "" || len(caseReport.Engines) != 3 {
			t.Fatalf("case %q did not carry three-tier evidence: %+v", caseReport.ID, caseReport)
		}
		for tier, execution := range caseReport.Engines {
			if execution.Outcome.Kind != "returned" || execution.Outcome.Value != want {
				t.Fatalf("case %q tier %s returned %+v, want %s", caseReport.ID, tier, execution.Outcome, want)
			}
		}
	}
}

func TestPhase22AppVerifyWrongExpectedAnswerControl(t *testing.T) {
	cases := session.ReplayCases{Schema: session.ReplayCasesSchema, Cases: []session.ReplayCase{{
		ID: "wrong-answer", Kind: session.ReplayCaseSource, Input: "7",
		Expected: session.ReplayExpected{Kind: "returned", Value: "8"}, ForeignOutcomes: []session.ReplayForeignOutcome{},
	}}}
	code, report := runCLIReplayCases(t, cases)
	if code != exitOperational || report.Verified || report.Cases[0].Status != session.ReplayStatusFail {
		t.Fatalf("wrong expected-answer control passed: code=%d report=%+v", code, report)
	}
	if len(report.Cases[0].Engines) != 3 || !strings.Contains(report.Cases[0].Diagnostic, "expected returned:8") {
		t.Fatalf("wrong-answer report omitted completed comparison evidence: %+v", report.Cases[0])
	}
}

func TestPhase22AppVerifyModelOnlyOutcomes(t *testing.T) {
	validOutcome := session.ReplayForeignOutcome{Operation: "fixture.value", Type: "U64", Value: "42"}
	for _, tc := range []struct {
		name     string
		outcomes []session.ReplayForeignOutcome
		want     string
		expected string
		pass     bool
	}{
		{name: "positive", outcomes: []session.ReplayForeignOutcome{validOutcome}, want: session.ReplayStatusPass, pass: true},
		{name: "missing", outcomes: []session.ReplayForeignOutcome{}, want: "modeled outcome is missing"},
		{name: "duplicate", outcomes: []session.ReplayForeignOutcome{validOutcome, validOutcome}, want: "modeled outcome is duplicated"},
		{name: "unconsumed", outcomes: []session.ReplayForeignOutcome{validOutcome, {Operation: "fixture.unused", Type: "U64", Value: "7"}}, want: "not consumed"},
		{name: "mismatched_expected", outcomes: []session.ReplayForeignOutcome{validOutcome}, want: "expected returned:43", expected: "43"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			expected := tc.expected
			if expected == "" {
				expected = "42"
			}
			cases := session.ReplayCases{Schema: session.ReplayCasesSchema, Cases: []session.ReplayCase{{
				ID: tc.name, Kind: session.ReplayCaseVerifierModel, Operation: "fixture.value",
				Expected: session.ReplayExpected{Kind: "returned", Value: expected}, ForeignOutcomes: tc.outcomes,
			}}}
			code, report := runCLIReplayCases(t, cases)
			caseReport := report.Cases[0]
			if tc.pass {
				if code != exitSuccess || !report.Verified || caseReport.Status != session.ReplayStatusPass {
					t.Fatalf("positive model outcome failed: code=%d report=%+v", code, report)
				}
				if caseReport.ModeledOutcome == nil || caseReport.ModeledOutcome.Value != "42" || caseReport.ActualHostIO || caseReport.PhysicalCleanup {
					t.Fatalf("model report overstates the scripted world: %+v", caseReport)
				}
				return
			}
			if code != exitOperational || report.Verified || caseReport.Status == session.ReplayStatusPass || !strings.Contains(caseReport.Diagnostic, tc.want) {
				t.Fatalf("negative model control %s passed or was not reached: code=%d report=%+v", tc.name, code, report)
			}
		})
	}
}

func TestPhase22AppVerifyRejectsOrdinaryForeignScriptsAndLocalC(t *testing.T) {
	caseWithScript := session.ReplayCases{Schema: session.ReplayCasesSchema, Cases: []session.ReplayCase{{
		ID: "ordinary-script", Kind: session.ReplayCaseSource, Input: "7",
		Expected:        session.ReplayExpected{Kind: "returned", Value: "7"},
		ForeignOutcomes: []session.ReplayForeignOutcome{{Operation: "fixture.value", Type: "U64", Value: "7"}},
	}}}
	code, report := runCLIReplayCases(t, caseWithScript)
	if code != exitOperational || report.Verified || report.Cases[0].Status != session.ReplayStatusUnsupported || !strings.Contains(report.Cases[0].Diagnostic, "does not consume") {
		t.Fatalf("ordinary source script was not explicitly refused: code=%d report=%+v", code, report)
	}

	casesPath := writeCLIReplayCases(t, session.ReplayCases{Schema: session.ReplayCasesSchema, Cases: []session.ReplayCase{{
		ID: "identity-7", Kind: session.ReplayCaseSource, Input: "7",
		Expected: session.ReplayExpected{Kind: "returned", Value: "7"}, ForeignOutcomes: []session.ReplayForeignOutcome{},
	}}})
	reportPath := filepath.Join(t.TempDir(), "local-c-report.json")
	source := testsupport.ProjectPath("examples", "phase22", "identity.schway")
	code, _, stderr := captureLangRun(t, []string{"app", "verify", source, "--cases", casesPath, "--report", reportPath, "--manifest", testsupport.ProjectPath("examples", "phase22", "identity.bindings.json")})
	if code != exitUsage || !strings.Contains(stderr, "local C manifests are unsupported") {
		t.Fatalf("local C manifest was not rejected on replay route: code=%d stderr=%q", code, stderr)
	}
	if _, err := os.Stat(reportPath); !os.IsNotExist(err) {
		t.Fatalf("manifest rejection unexpectedly created a verification report: err=%v", err)
	}
}

func TestPhase22OrdinaryAppRunDoesNotEnterReplayRoute(t *testing.T) {
	artifact := writeCLIPhase22Script(t, "printf 'ordinary:%s\\n' \"$1\"\n")
	code, stdout, stderr := captureLangRun(t, []string{"app", "run", artifact, "--", "7"})
	if code != 0 || stdout != "ordinary:7\n" || stderr != "" {
		t.Fatalf("ordinary app run changed route or output: code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
}

func runCLIReplayCases(t *testing.T, cases session.ReplayCases) (int, session.ReplayReport) {
	t.Helper()
	casesPath := writeCLIReplayCases(t, cases)
	reportPath := filepath.Join(t.TempDir(), "report.json")
	source := testsupport.ProjectPath("examples", "phase22", "identity.schway")
	code, _, _ := captureLangRun(t, []string{"app", "verify", source, "--cases", casesPath, "--report", reportPath})
	return code, readCLIReplayReport(t, reportPath)
}

func writeCLIReplayCases(t *testing.T, cases session.ReplayCases) string {
	t.Helper()
	encoded, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "cases.json")
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func readCLIReplayReport(t *testing.T, path string) session.ReplayReport {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read verification report: %v", err)
	}
	var report session.ReplayReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("decode verification report: %v", err)
	}
	return report
}

func captureLangRun(t *testing.T, args []string) (int, string, string) {
	t.Helper()
	stdoutRead, stdoutWrite, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stderrRead, stderrWrite, err := os.Pipe()
	if err != nil {
		stdoutRead.Close()
		stdoutWrite.Close()
		t.Fatal(err)
	}
	previousStdout, previousStderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = stdoutWrite, stderrWrite
	code := run(args)
	stdoutWrite.Close()
	stderrWrite.Close()
	os.Stdout, os.Stderr = previousStdout, previousStderr
	stdoutBytes, stdoutErr := io.ReadAll(stdoutRead)
	stderrBytes, stderrErr := io.ReadAll(stderrRead)
	stdoutRead.Close()
	stderrRead.Close()
	if stdoutErr != nil || stderrErr != nil {
		t.Fatalf("capture stdout=%v stderr=%v", stdoutErr, stderrErr)
	}
	return code, string(stdoutBytes), string(stderrBytes)
}

func writeCLIPhase22Script(t *testing.T, body string) string {
	t.Helper()
	artifact := filepath.Join(t.TempDir(), "app")
	if err := os.WriteFile(artifact, []byte("#!/bin/sh\nset -eu\n"+body), 0o700); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(artifact)
	if err != nil {
		t.Fatal(err)
	}
	executableDigest := sha256.Sum256(contents)
	receipt := native.BuildReceipt{
		Schema:              native.ApplicationBuildSchema,
		ExecutableDigest:    hex.EncodeToString(executableDigest[:]),
		SourceDigest:        "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		EmittedCDigest:      "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		Compiler:            "/test/clang",
		CompilerVersion:     "test clang",
		Target:              runtime.GOOS + "/" + runtime.GOARCH,
		Flags:               []string{"test"},
		RuntimeDependencies: []string{"platform-c-runtime"},
		DependencyClosure:   "incomplete",
		Cacheable:           false,
	}
	encoded, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(artifact+".schway-build.json", encoded, 0o600); err != nil {
		t.Fatal(err)
	}
	return artifact
}
