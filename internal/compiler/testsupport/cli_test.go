package testsupport_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/protocol"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

func TestCLIOutputContract(t *testing.T) {
	binary := testsupport.BuildCLI(t)
	fixture := testsupport.ProjectPath("testdata", "phase1", "toggle.lang")
	for _, arguments := range [][]string{
		{"--json", "check", fixture},
		{"check", "--json", fixture},
		{"check", fixture, "--json"},
	} {
		result := testsupport.RunCLI(t, binary, nil, arguments...)
		if result.Exit != 0 || len(result.Stderr) != 0 {
			t.Fatalf("JSON success stream contract failed for %v: exit=%d stdout=%q stderr=%q", arguments, result.Exit, result.Stdout, result.Stderr)
		}
		var decoded protocol.Result
		if err := json.Unmarshal(result.Stdout, &decoded); err != nil {
			t.Fatalf("invalid JSON for %v: %v\n%s", arguments, err, result.Stdout)
		}
		if decoded.Metrics.OutputBytes != len(result.Stdout) || decoded.Metrics.RecomputedWork == 0 {
			t.Fatalf("missing command observations for %v: metrics=%+v actual_bytes=%d", arguments, decoded.Metrics, len(result.Stdout))
		}
		canonical, err := protocol.JSON(decoded)
		if err != nil || !bytes.Equal(canonical, result.Stdout) {
			t.Fatalf("noncanonical JSON for %v:\ngot=%q\nwant=%q err=%v", arguments, result.Stdout, canonical, err)
		}
	}
}

func TestCLIExitTaxonomy(t *testing.T) {
	binary := testsupport.BuildCLI(t)
	valid := testsupport.ProjectPath("testdata", "phase1", "toggle.lang")
	invalid := testsupport.ProjectPath("testdata", "phase1", "non_exhaustive.lang")

	cases := []struct {
		name string
		env  []string
		args []string
		exit int
	}{
		{name: "success", args: []string{"check", valid}, exit: 0},
		{name: "invalid source", args: []string{"check", invalid}, exit: 2},
		{name: "operational read", args: []string{"check", filepath.Join(t.TempDir(), "absent.lang")}, exit: 3},
		{name: "operational tool", env: []string{"PATH=" + t.TempDir()}, args: []string{"run", "--engine=native", valid}, exit: 3},
		{name: "usage", args: []string{"unknown"}, exit: 64},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			result := testsupport.RunCLI(t, binary, test.env, test.args...)
			if result.Exit != test.exit {
				t.Fatalf("exit=%d want=%d stdout=%q stderr=%q", result.Exit, test.exit, result.Stdout, result.Stderr)
			}
		})
	}
	if got := protocol.ExitCode(protocol.StatusMismatch); got != 4 {
		t.Fatalf("semantic mismatch exit=%d want=4", got)
	}
}

func TestCLISourceByteLimitBoundary(t *testing.T) {
	binary := testsupport.BuildCLI(t)
	prefix := []byte("module limits.source\nexport { fn keep }\nfn keep(code: Byte) -> Byte { code }\n//")
	for _, test := range []struct {
		name string
		size int
		exit int
		code string
	}{
		{name: "exact", size: 1 << 20, exit: 0},
		{name: "one over", size: 1<<20 + 1, exit: 2, code: "syntax.input_limit"},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "input.lang")
			source := append(append([]byte(nil), prefix...), bytes.Repeat([]byte{'x'}, test.size-len(prefix))...)
			if err := os.WriteFile(path, source, 0o600); err != nil {
				t.Fatal(err)
			}
			result := testsupport.RunCLI(t, binary, nil, "--json", "check", path)
			if result.Exit != test.exit {
				t.Fatalf("exit=%d want=%d stdout=%s", result.Exit, test.exit, result.Stdout)
			}
			if test.code != "" && !bytes.Contains(result.Stdout, []byte(test.code)) {
				t.Fatalf("missing %s: %s", test.code, result.Stdout)
			}
		})
	}
}

func TestCLICheckWorkScalesWithSource(t *testing.T) {
	binary := testsupport.BuildCLI(t)
	work := make([]int, 0, 3)
	for _, count := range []int{1, 10, 100} {
		var source strings.Builder
		source.WriteString("module work.scale\nexport { fn keep }\nfn keep(code: Byte) -> Byte {\n")
		for index := 0; index < count; index++ {
			fmt.Fprintf(&source, "  let value%d = code\n", index)
		}
		source.WriteString("  code\n}\n")
		path := filepath.Join(t.TempDir(), fmt.Sprintf("scale-%d.lang", count))
		if err := os.WriteFile(path, []byte(source.String()), 0o600); err != nil {
			t.Fatal(err)
		}
		result := testsupport.RunCLI(t, binary, nil, "--json", "check", path)
		var decoded protocol.Result
		if result.Exit != 0 || json.Unmarshal(result.Stdout, &decoded) != nil {
			t.Fatalf("count=%d result=%+v", count, result)
		}
		work = append(work, decoded.Metrics.RecomputedWork)
	}
	if !(work[0] < work[1] && work[1] < work[2]) {
		t.Fatalf("check work does not scale: %v", work)
	}
}

func TestCLICheckRejectsWorkBeyondLimit(t *testing.T) {
	binary := testsupport.BuildCLI(t)
	var source strings.Builder
	source.WriteString("module work.limit\nexport { fn keep }\nfn keep(code: Byte) -> Byte {\n")
	for index := 0; index < 12_000; index++ {
		fmt.Fprintf(&source, "  let value%d = code\n", index)
	}
	source.WriteString("  code\n}\n")
	path := filepath.Join(t.TempDir(), "over-work.lang")
	if err := os.WriteFile(path, []byte(source.String()), 0o600); err != nil {
		t.Fatal(err)
	}
	result := testsupport.RunCLI(t, binary, nil, "--json", "check", path)
	if result.Exit != 2 || !bytes.Contains(result.Stdout, []byte("check.work_limit")) {
		t.Fatalf("work limit result=%+v", result)
	}
}

func TestHumanJSONIdentityParity(t *testing.T) {
	binary := testsupport.BuildCLI(t)
	for _, test := range []struct {
		name string
		args []string
	}{
		{name: "check success", args: []string{"check", testsupport.ProjectPath("testdata", "phase1", "toggle.lang")}},
		{name: "check invalid", args: []string{"check", testsupport.ProjectPath("testdata", "phase1", "non_exhaustive.lang")}},
		{name: "interpreter", args: []string{"run", "--engine=interpreter", testsupport.ProjectPath("testdata", "phase1", "toggle.lang")}},
	} {
		t.Run(test.name, func(t *testing.T) {
			human := testsupport.RunCLI(t, binary, nil, test.args...)
			jsonArguments := append([]string{"--json"}, test.args...)
			machine := testsupport.RunCLI(t, binary, nil, jsonArguments...)
			if human.Exit != machine.Exit || len(machine.Stderr) != 0 {
				t.Fatalf("projection stream/exit mismatch: human=%+v machine=%+v", human, machine)
			}
			var decoded protocol.Result
			if err := json.Unmarshal(machine.Stdout, &decoded); err != nil {
				t.Fatal(err)
			}
			humanBytes := human.Stdout
			if human.Exit != 0 {
				humanBytes = human.Stderr
				if len(human.Stdout) != 0 {
					t.Fatalf("human failure wrote stdout: %q", human.Stdout)
				}
			}
			if !strings.Contains(string(humanBytes), decoded.ID) {
				t.Fatalf("human projection lacks result ID %q: %q", decoded.ID, humanBytes)
			}
			for _, problem := range decoded.Diagnostics {
				if !strings.Contains(string(humanBytes), problem.ID) {
					t.Fatalf("human projection lacks diagnostic ID %q: %q", problem.ID, humanBytes)
				}
			}
			for _, execution := range decoded.Executions {
				for _, event := range execution.Events {
					if !strings.Contains(string(humanBytes), event.ID) {
						t.Fatalf("human projection lacks event ID %q: %q", event.ID, humanBytes)
					}
				}
			}
		})
	}

	jsonFailure := testsupport.RunCLI(t, binary, os.Environ(), "--json", "check", testsupport.ProjectPath("testdata", "phase1", "non_exhaustive.lang"))
	if jsonFailure.Exit != 2 || len(jsonFailure.Stderr) != 0 || len(jsonFailure.Stdout) == 0 {
		t.Fatalf("JSON failure stream contract: %+v", jsonFailure)
	}
}

func TestHumanJSONMixedDiagnosticVersionParity(t *testing.T) {
	binary := testsupport.BuildCLI(t)
	for _, fixture := range []string{
		testsupport.ProjectPath("testdata", "phase1", "non_exhaustive.lang"),
		testsupport.ProjectPath("testdata", "phase2", "use_after_move.lang"),
	} {
		human := testsupport.RunCLI(t, binary, nil, "check", fixture)
		machine := testsupport.RunCLI(t, binary, nil, "--json", "check", fixture)
		if human.Exit != 2 || machine.Exit != 2 || len(human.Stdout) != 0 || len(machine.Stderr) != 0 {
			t.Fatalf("projection stream/exit mismatch for %s: human=%+v machine=%+v", fixture, human, machine)
		}
		var decoded protocol.Result
		if err := json.Unmarshal(machine.Stdout, &decoded); err != nil {
			t.Fatal(err)
		}
		if decoded.Schema != "lang.command/0" || len(decoded.Diagnostics) != 1 {
			t.Fatalf("command envelope or diagnostic missing: %+v", decoded)
		}
		if !strings.Contains(string(human.Stderr), decoded.ID) || !strings.Contains(string(human.Stderr), decoded.Diagnostics[0].ID) {
			t.Fatalf("human projection lost machine identities: human=%q machine=%+v", human.Stderr, decoded)
		}
	}

	// 03-02-03 (D-02-09/D-07): testdata/phase2/ability_shapes.lang now
	// carries one check.unexecutable_shape diagnostic per function (three
	// total, not one), so it is verified separately from the single-
	// diagnostic loop above — but the same human/JSON identity-parity
	// contract must hold for every one of them.
	fixture := testsupport.ProjectPath("testdata", "phase2", "ability_shapes.lang")
	human := testsupport.RunCLI(t, binary, nil, "check", fixture)
	machine := testsupport.RunCLI(t, binary, nil, "--json", "check", fixture)
	if human.Exit != 2 || machine.Exit != 2 || len(human.Stdout) != 0 || len(machine.Stderr) != 0 {
		t.Fatalf("projection stream/exit mismatch for %s: human=%+v machine=%+v", fixture, human, machine)
	}
	var decoded protocol.Result
	if err := json.Unmarshal(machine.Stdout, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Schema != "lang.command/0" || len(decoded.Diagnostics) != 3 {
		t.Fatalf("command envelope or diagnostics missing: %+v", decoded)
	}
	if !strings.Contains(string(human.Stderr), decoded.ID) {
		t.Fatalf("human projection lost machine result identity: human=%q machine=%+v", human.Stderr, decoded)
	}
	for _, problem := range decoded.Diagnostics {
		if problem.Code != "check.unexecutable_shape" {
			t.Fatalf("unexpected diagnostic code %q: %+v", problem.Code, problem)
		}
		if !strings.Contains(string(human.Stderr), problem.ID) {
			t.Fatalf("human projection lost machine diagnostic identity %q: human=%q", problem.ID, human.Stderr)
		}
	}
}

func TestVerifyCLI(t *testing.T) {
	binary := testsupport.BuildCLI(t)
	corpus := testsupport.ProjectPath("testdata", "phase1")
	machine := testsupport.RunCLI(t, binary, nil, "--json", "verify", corpus)
	if machine.Exit != 0 || len(machine.Stderr) != 0 {
		t.Fatalf("verify CLI failed: %+v", machine)
	}
	var result protocol.Result
	if err := json.Unmarshal(machine.Stdout, &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != protocol.StatusPass || len(result.Lanes) != 5 {
		t.Fatalf("verify result omitted lanes: %+v", result)
	}
	for _, required := range []string{"control:match.non_exhaustive", "control:evidence.source_mismatch", "control:interpreter-o0-o3"} {
		if !bytes.Contains(machine.Stdout, []byte(required)) {
			t.Fatalf("verify JSON omitted %s: %s", required, machine.Stdout)
		}
	}
}

func TestVerifyPhase2CLI(t *testing.T) {
	binary := testsupport.BuildCLI(t)
	machine := testsupport.RunCLI(t, binary, nil, "--json", "verify", testsupport.ProjectPath("testdata", "phase2"))
	if machine.Exit != 0 || len(machine.Stderr) != 0 {
		t.Fatalf("Phase 2 verify CLI failed: %+v", machine)
	}
	var result protocol.Result
	if err := json.Unmarshal(machine.Stdout, &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != protocol.StatusPass || !reflect.DeepEqual(result.ExpectedEscapes, []string{"escape:coordinated-source-core-lie"}) {
		t.Fatalf("Phase 2 verify result omitted expected escape: %+v", result)
	}
	for _, required := range []string{"control:ownership.use_after_move", "control:ownership.move_while_borrowed", "control:ownership.move_while_reborrowed", "control:ownership.transfer_requires_take", "control:ability.forged_copy", "control:core.duplicate_operation_id", "control:interpreter-o0-o3-owned", "control:evidence.core_mismatch", "control:backend.runtime_causality"} {
		if !bytes.Contains(machine.Stdout, []byte(required)) {
			t.Fatalf("verify JSON omitted %s", required)
		}
	}
}

func TestVerifyCorpusSourceByteLimitBoundary(t *testing.T) {
	binary := testsupport.BuildCLI(t)
	for _, test := range []struct {
		name  string
		phase string
		file  string
		size  int
		exit  int
	}{
		{name: "base exact", phase: "phase1", file: "toggle.lang", size: 1 << 20, exit: 0},
		{name: "base over", phase: "phase1", file: "toggle.lang", size: 1<<20 + 1, exit: 2},
		{name: "owned exact", phase: "phase2", file: "owned_transfer.lang", size: 1 << 20, exit: 0},
		{name: "owned over", phase: "phase2", file: "owned_transfer.lang", size: 1<<20 + 1, exit: 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			corpus := t.TempDir()
			sourceCorpus := testsupport.ProjectPath("testdata", test.phase)
			entries, err := os.ReadDir(sourceCorpus)
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".lang") {
					continue
				}
				data, readErr := os.ReadFile(filepath.Join(sourceCorpus, entry.Name()))
				if readErr != nil {
					t.Fatal(readErr)
				}
				if entry.Name() == test.file {
					data = append(data, '/', '/')
					data = append(data, bytes.Repeat([]byte{'x'}, test.size-len(data)-1)...)
					data = append(data, '\n')
				}
				if writeErr := os.WriteFile(filepath.Join(corpus, entry.Name()), data, 0o600); writeErr != nil {
					t.Fatal(writeErr)
				}
			}
			result := testsupport.RunCLI(t, binary, nil, "--json", "verify", corpus)
			if result.Exit != test.exit {
				t.Fatalf("exit=%d want=%d stdout=%s", result.Exit, test.exit, result.Stdout)
			}
			if test.exit != 0 && !bytes.Contains(result.Stdout, []byte("verify.fixture_input_limit")) {
				t.Fatalf("missing bounded fixture diagnostic: %s", result.Stdout)
			}
		})
	}
}

// TestInterfaceCheckIsBodyBlindCLI is 03-06-01's CLI falsifier for OWN-04
// success criterion 4: separate compilation is demonstrated as two real
// invocations of the built binary, and the consuming (`interface check`)
// invocation never touches a body field. It proves the second claim
// structurally rather than by inspection: the "core artifact" file handed to
// `interface check` is deliberately NOT valid core.Program JSON at all (it
// cannot be unmarshaled into a struct carrying a Linear/Match field), yet the
// command still answers correctly once its digest matches — which is only
// possible if the consuming path never attempts to decode it as anything
// other than a byte string to hash.
func TestInterfaceCheckIsBodyBlindCLI(t *testing.T) {
	binary := testsupport.BuildCLI(t)
	fixture := testsupport.ProjectPath("testdata", "phase3", "public_view.lang")
	dir := t.TempDir()
	summaryPath := filepath.Join(dir, "summary.json")
	corePath := filepath.Join(dir, "core.json")

	exportResult := testsupport.RunCLI(t, binary, nil, "--json", "interface", "export", fixture, summaryPath)
	if exportResult.Exit != 0 {
		t.Fatalf("interface export failed: %+v", exportResult)
	}
	var exported protocol.Result
	if err := json.Unmarshal(exportResult.Stdout, &exported); err != nil {
		t.Fatal(err)
	}
	if exported.Interface == nil || len(exported.Interface.Functions) != 1 || exported.Interface.Functions[0].Access != "shared" {
		t.Fatalf("interface export omitted the origin answer: %+v", exported.Interface)
	}
	summaryBytes, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{`"linear"`, `"match"`} {
		if bytes.Contains(summaryBytes, []byte(forbidden)) {
			t.Fatalf("exported summary leaked a body field %s: %s", forbidden, summaryBytes)
		}
	}

	// The "core artifact" this invocation binds against is deliberately not
	// decodable as core.Program at all — a real core artifact never looks
	// like this. The digest is computed over these exact bytes by
	// `interface core`, so we compute the digest the same way (SHA-256 over
	// the raw file bytes) and write a summary bound to it, proving the check
	// path only ever hashes the bytes rather than parsing them.
	notACoreProgram := []byte(`this is deliberately not JSON and has no "linear" or "match" body field, only bytes to hash`)
	if err := os.WriteFile(corePath, notACoreProgram, 0o600); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(notACoreProgram)
	digest := "sha256:" + hex.EncodeToString(sum[:])
	var summary map[string]any
	if err := json.Unmarshal(summaryBytes, &summary); err != nil {
		t.Fatal(err)
	}
	summary["core_digest"] = digest
	rebound, err := json.Marshal(summary)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(summaryPath, rebound, 0o600); err != nil {
		t.Fatal(err)
	}

	checkResult := testsupport.RunCLI(t, binary, nil, "--json", "interface", "check", summaryPath, corePath)
	if checkResult.Exit != 0 {
		t.Fatalf("interface check rejected a summary bound to non-core bytes it should never have parsed: %+v", checkResult)
	}
	var checked protocol.Result
	if err := json.Unmarshal(checkResult.Stdout, &checked); err != nil {
		t.Fatal(err)
	}
	if checked.Interface == nil || len(checked.Interface.Functions) != 1 || checked.Interface.Functions[0].Paths[0] != "buffer" || checked.Interface.Functions[0].Access != "shared" {
		t.Fatalf("interface check did not decide from the summary alone: %+v", checked.Interface)
	}
	found := false
	for _, escape := range checked.ExpectedEscapes {
		if escape == "escape:coordinated-frontend-summary-lie" {
			found = true
		}
	}
	if !found {
		t.Fatalf("interface check omitted the named origin escape: %+v", checked.ExpectedEscapes)
	}

	// A digest mismatch is still rejected before anything else, using the
	// genuine two-invocation flow end to end.
	corePathReal := filepath.Join(dir, "core-real.json")
	coreResult := testsupport.RunCLI(t, binary, nil, "interface", "core", fixture, corePathReal)
	if coreResult.Exit != 0 {
		t.Fatalf("interface core failed: %+v", coreResult)
	}
	freshExport := testsupport.RunCLI(t, binary, nil, "interface", "export", fixture, summaryPath)
	if freshExport.Exit != 0 {
		t.Fatalf("interface export failed: %+v", freshExport)
	}
	staleResult := testsupport.RunCLI(t, binary, nil, "--json", "interface", "check", summaryPath, corePath)
	if staleResult.Exit != 2 || !bytes.Contains(staleResult.Stdout, []byte("origin.stale_summary")) {
		t.Fatalf("expected origin.stale_summary for a mismatched core artifact: %+v", staleResult)
	}
	freshCheck := testsupport.RunCLI(t, binary, nil, "--json", "interface", "check", summaryPath, corePathReal)
	if freshCheck.Exit != 0 {
		t.Fatalf("interface check against the real core artifact failed: %+v", freshCheck)
	}
}

// TestBranchFixturesThroughCLI is 03-03-02's D-11 shipped-binary proof: the
// edge-specific accept/reject fixture pair is driven through the real
// `./cmd/lang` binary's JSON projection (not only the in-process harness),
// asserting the exact exit codes and diagnostic codes the CLI reports.
func TestBranchFixturesThroughCLI(t *testing.T) {
	binary := testsupport.BuildCLI(t)
	accept := testsupport.ProjectPath("testdata", "phase3", "branch_one_arm_shared_accept.lang")
	reject := testsupport.ProjectPath("testdata", "phase3", "branch_one_arm_shared_reject.lang")

	acceptResult := testsupport.RunCLI(t, binary, nil, "--json", "check", accept)
	if acceptResult.Exit != 0 {
		t.Fatalf("accept fixture: exit=%d stdout=%s stderr=%s", acceptResult.Exit, acceptResult.Stdout, acceptResult.Stderr)
	}
	var acceptDecoded protocol.Result
	if err := json.Unmarshal(acceptResult.Stdout, &acceptDecoded); err != nil {
		t.Fatalf("accept fixture: invalid JSON: %v\n%s", err, acceptResult.Stdout)
	}
	if len(acceptDecoded.Diagnostics) != 0 {
		t.Fatalf("accept fixture: unexpected diagnostics via CLI: %+v", acceptDecoded.Diagnostics)
	}

	rejectResult := testsupport.RunCLI(t, binary, nil, "--json", "check", reject)
	if rejectResult.Exit != 2 {
		t.Fatalf("reject fixture: exit=%d want=2 stdout=%s stderr=%s", rejectResult.Exit, rejectResult.Stdout, rejectResult.Stderr)
	}
	var rejectDecoded protocol.Result
	if err := json.Unmarshal(rejectResult.Stdout, &rejectDecoded); err != nil {
		t.Fatalf("reject fixture: invalid JSON: %v\n%s", err, rejectResult.Stdout)
	}
	if len(rejectDecoded.Diagnostics) != 1 || rejectDecoded.Diagnostics[0].Code != "ownership.move_while_borrowed" {
		t.Fatalf("reject fixture: diagnostics via CLI diverged: %+v", rejectDecoded.Diagnostics)
	}
}

func TestPhase2VerifierScriptContract(t *testing.T) {
	script, err := os.ReadFile(testsupport.ProjectPath("scripts", "verify-phase2.sh"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(script)
	if strings.Contains(text, "verify-phase1.sh") || strings.Count(text, "\ngo test ./...\n") != 1 || strings.Count(text, "\ngo test -race ./...\n") != 1 || strings.Count(text, "\ngo vet ./...\n") != 1 {
		t.Fatalf("Phase 2 gate duplicates or nests shared verification:\n%s", text)
	}
	for _, required := range []string{"assert-go-tests.sh --self-test", "TestOwnedBackendMutationIsMismatch", "control:backend.runtime_causality", "verify testdata/phase1", "verify testdata/phase2", "warm_samples=20", "peak_rss=unavailable", "p50_ns=", "p95_ns=", "min_ns=", "max_ns=", "output_bytes=", "work="} {
		if !strings.Contains(text, required) {
			t.Fatalf("Phase 2 gate omitted %q", required)
		}
	}
}
