package session_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase38ReceiptTracer(t *testing.T) {
	cli := testsupport.BuildCLI(t)
	work := t.TempDir()
	input := filepath.Join(work, "check.bin")
	if err := os.WriteFile(input, []byte("123456789"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, engine := range []string{"interpreter", "native"} {
		t.Run(engine, func(t *testing.T) {
			args := []string{"run", "--engine=" + engine, testsupport.ProjectPath("examples", "crc32", "crc32.schway")}
			if engine == "native" {
				args = append(args, "--manifest", testsupport.ProjectPath("examples", "crc32", "file_bytes.bindings.json"))
			}
			args = append(args, "--", input)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, cli, args...)
			cmd.Dir = work
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			err := cmd.Run()
			if ctx.Err() != nil {
				t.Fatalf("CRC public route exceeded its deadline: %v", ctx.Err())
			}
			if err != nil {
				t.Fatalf("CRC public route failed: %v; stderr=%q", err, stderr.String())
			}
			if stdout.String() != "3421780262\n" || stderr.Len() != 0 {
				t.Fatalf("CRC tuple = (stdout=%q, stderr=%q), want (3421780262 newline, empty)", stdout.String(), stderr.String())
			}
		})
	}

	t.Run("bounded-public-probe", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "sh", testsupport.ProjectPath("scripts", "verify-phase32.sh"), "--milestone", "M007", "--probe", "crc32/check")
		cmd.Dir = testsupport.ProjectPath()
		var stdout, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		if err := cmd.Run(); ctx.Err() != nil {
			t.Fatalf("bounded receipt probe exceeded its deadline: %v", ctx.Err())
		} else if err != nil {
			t.Fatalf("bounded receipt probe failed: %v; stdout=%q stderr=%q", err, stdout.String(), stderr.String())
		}
		if !strings.Contains(stdout.String(), "phase38 probe result status=pass") || strings.Contains(stdout.String(), "phase38 host result") {
			t.Fatalf("probe output did not have its bounded marker: %q", stdout.String())
		}
	})

	t.Run("synthetic-two-host-binding-and-revisions", func(t *testing.T) {
		fixture := phase38NewBindingFixture(t)
		got := phase38RunBinding(t, fixture, fixture.run, fixture.linuxLog, fixture.macosLog)
		if got.exit != 0 || !strings.Contains(got.stdout, "phase38 binding: status=pass") {
			t.Fatalf("valid synthetic receipt was not bound: exit=%d stdout=%q stderr=%q", got.exit, got.stdout, got.stderr)
		}

		wrongRun := fixture.run
		wrongRun.HeadSHA = strings.Repeat("1", 40)
		got = phase38RunBinding(t, fixture, wrongRun, fixture.linuxLog, fixture.macosLog)
		phase38RequireRefusal(t, got, "wrong_revision")

		for _, host := range []string{"Linux", "macOS"} {
			logs := map[string][]byte{"Linux": fixture.linuxLog, "macOS": fixture.macosLog}
			mutated := bytes.Replace(logs[host], []byte("headsha="+fixture.freeze.FreezeSHA), []byte("headsha="+strings.Repeat("2", 40)), 1)
			if bytes.Equal(mutated, logs[host]) {
				t.Fatalf("could not mutate %s log headsha", host)
			}
			if host == "Linux" {
				got = phase38RunBinding(t, fixture, fixture.run, mutated, fixture.macosLog)
			} else {
				got = phase38RunBinding(t, fixture, fixture.run, fixture.linuxLog, mutated)
			}
			phase38RequireRefusal(t, got, "wrong_revision")
		}
	})

	t.Run("probe-is-not-a-host-receipt", func(t *testing.T) {
		fixture := phase38NewBindingFixture(t)
		probe := []byte("phase38 probe result status=pass\n")
		got := phase38RunBinding(t, fixture, fixture.run, probe, probe)
		phase38RequireRefusal(t, got, "probe_not_full_receipt")
	})
}

func TestPhase38ManifestConsumerBindings(t *testing.T) {
	manifest, err := os.ReadFile(testsupport.ProjectPath("testdata", "receipts", "M007.manifest"))
	if err != nil {
		t.Fatal(err)
	}
	old, err := os.ReadFile(testsupport.ProjectPath("testdata", "receipts", "M006.manifest"))
	if err != nil {
		t.Fatal(err)
	}
	actualOld := make([]string, 0, 39)
	for _, line := range strings.Split(string(manifest), "\n") {
		if strings.HasPrefix(line, "# case ") && !strings.HasPrefix(line, "# case crc32 ") && !strings.HasPrefix(line, "# case crc_low ") && !strings.HasPrefix(line, "# case helper_") {
			actualOld = append(actualOld, line)
		}
	}
	wantOld := make([]string, 0, 39)
	for _, line := range strings.Split(string(old), "\n") {
		if strings.HasPrefix(line, "# case ") {
			wantOld = append(wantOld, line)
		}
	}
	if !reflect.DeepEqual(actualOld, wantOld) || len(wantOld) != 39 {
		t.Fatalf("M007 inherited M006 case rows differ: actual=%d want=%d", len(actualOld), len(wantOld))
	}
	lines := map[string]int{}
	for _, line := range strings.Split(string(manifest), "\n") {
		if line != "" {
			lines[line]++
		}
	}
	for _, consumer := range []string{"crc32", "crc_low"} {
		cases, err := testsupport.CRC32CaseMatrix()
		if err != nil {
			t.Fatal(err)
		}
		for _, row := range cases {
			if row.Kind == "direct-argument" {
				continue
			}
			want := row.Expected[consumer]
			program := consumer
			caseLine := fmt.Sprintf("# case %s %s %d %s %s", program, row.ID, want.Exit, phase38ManifestField(want.Stdout), phase38ManifestField(want.Stderr))
			if lines[caseLine] != 1 {
				t.Errorf("manifest does not bind fixture tuple exactly once: %s", caseLine)
			}
			for _, engine := range []string{"native", "interpreter"} {
				name := fmt.Sprintf("TestPhase37PublicFileMatrix/%s/%s/%s", row.ID, consumer, engine)
				line := fmt.Sprintf("# reached %s %s ./cmd/schway %s", program, row.ID, name)
				if lines[line] != 1 {
					t.Errorf("manifest is missing reached case binding: %s", line)
				}
			}
		}
	}
	helper := phase36ReadHelperValues(t)
	for _, row := range helper.Cases {
		program := "helper_inline"
		if row.Source == "crc_step.schway" {
			program = "helper_crc_step"
		} else if row.Source == "importer.schway" {
			program = "helper_importer"
		}
		caseLine := fmt.Sprintf("# case %s %s %d %s %s", program, row.Case, row.ExpectedExit, phase38ManifestField(row.ExpectedStdout), phase38ManifestField(row.ExpectedStderr))
		if lines[caseLine] != 1 {
			t.Errorf("helper fixture tuple is not inventoried exactly once: %s", caseLine)
		}
		line := fmt.Sprintf("# reached %s %s ./internal/compiler/session TestPhase36HelperAnswersAndDefects/%s", program, row.Case, row.Case)
		if lines[line] != 1 {
			t.Errorf("helper fixture consumer binding is missing: %s", line)
		}
	}
	for _, record := range []string{
		"# structured public-helper ./cmd/schway TestPhase36PublicHelperConformance/interpreter TestPhase36PublicHelperConformance/native value=1994146192 calls=2 returns=3 resources=0",
		"# structured bitwise-locked ./cmd/schway TestPhase35PublicConformanceWitness/interpreter/locked-264 TestPhase35PublicConformanceWitness/native/locked-264 value=264",
		"# structured bitwise-alternate ./cmd/schway TestPhase35PublicConformanceWitness/interpreter/alternate-8 TestPhase35PublicConformanceWitness/native/alternate-8 value=8",
		"# root bitwise-optimization ./internal/compiler/session ^TestPhase35BitwiseOptimizationMatrix$",
		"# structured bitwise-optimization-locked ./internal/compiler/session TestPhase35BitwiseOptimizationMatrix/locked-association/O0 TestPhase35BitwiseOptimizationMatrix/locked-association/O3 TestPhase35BitwiseOptimizationMatrix/locked-association/O3-flto value=264",
		"# structured bitwise-optimization-alternate ./internal/compiler/session TestPhase35BitwiseOptimizationMatrix/alternate-final-mask-association/O0 TestPhase35BitwiseOptimizationMatrix/alternate-final-mask-association/O3 TestPhase35BitwiseOptimizationMatrix/alternate-final-mask-association/O3-flto value=8",
	} {
		if lines[record] != 1 {
			t.Errorf("manifest structured witness is missing: %s", record)
		}
	}
	if !strings.Contains(string(manifest), "controls=guide,route-audit,codedigest,receipt-rules,crc-public,helper-answers,helper-optimization,public-helper,bitwise-public,bitwise-optimization,") {
		t.Fatal("manifest controls omit the named Phase 35 optimization matrix")
	}
	for _, vector := range phase35ReadValueVectors(t).Cases {
		for _, suffix := range []string{"", "/O0", "/O3", "/O3-flto"} {
			record := "# event bitwise-optimization TestPhase35BitwiseOptimizationMatrix/" + vector.Name + suffix
			if lines[record] != 1 {
				t.Errorf("bitwise optimization event inventory is incomplete: %s", record)
			}
		}
	}
}

func TestPhase38ProducerInventoryControls(t *testing.T) {
	fixture := phase38NewBindingFixture(t)
	valid := phase38RunBinding(t, fixture, fixture.run, fixture.linuxLog, fixture.macosLog)
	if valid.exit != 0 || !strings.Contains(valid.stdout, "phase38 binding: status=pass") {
		t.Fatalf("complete synthetic event inventory was refused: %#v", valid)
	}
	for _, tc := range []struct {
		name, events, reason string
	}{
		{name: "skipped-required-root", events: string(phase38SyntheticEvents("skip")), reason: "control_test_skipped"},
		{name: "truncated-before-pass", events: string(phase38SyntheticEvents("run")), reason: "control_event_incomplete"},
		{name: "missing-required-name", events: string(phase38SyntheticEvents("run", "pass", "extra")), reason: "control_event_inventory"},
		{name: "duplicate-run-event", events: string(phase38SyntheticEvents("run", "run", "pass")), reason: "control_event_inventory"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mutated := bytes.Replace(fixture.linuxLog,
				[]byte(hex.EncodeToString(phase38SyntheticEvents("run", "output", "pass"))),
				[]byte(hex.EncodeToString([]byte(tc.events))), 1)
			got := phase38RunBinding(t, fixture, fixture.run, mutated, fixture.macosLog)
			phase38RequireRefusal(t, got, tc.reason)
		})
	}

	badSelector := phase33RunAtEnv(t, testsupport.ProjectPath(), "scripts/verify-phase32.sh", os.Environ(), "--milestone", "M007", "--probe", "crc32/not-a-case")
	if badSelector.exit == 0 || !strings.Contains(badSelector.stdout+badSelector.stderr, "invalid_selector") {
		t.Fatalf("producer accepted an empty or unrecognized probe selector: %#v", badSelector)
	}
	missingGo := phase33RunAtEnv(t, testsupport.ProjectPath(), "scripts/verify-phase32.sh", envWith("PATH", "/usr/bin:/bin"), "--milestone", "M007", "--probe", "crc32/check")
	if missingGo.exit == 0 || !strings.Contains(missingGo.stdout+missingGo.stderr, "unavailable_tool_go") {
		t.Fatalf("producer did not fail closed when Go is unavailable: %#v", missingGo)
	}
}

func TestPhase38BindingControls(t *testing.T) {
	fixture := phase38NewBindingFixture(t)
	valid := phase38RunBindingWithOrder(t, fixture, fixture.run, fixture.linuxLog, fixture.macosLog, true)
	if valid.exit != 0 || !strings.Contains(valid.stdout, "phase38 binding: status=pass") {
		t.Fatalf("valid reversed host/log order was refused: %#v", valid)
	}

	for _, tc := range []struct {
		name, reason string
		mutate       func(*phase38RunFixture)
	}{
		{"mixed-attempt", "mixed_attempt", func(run *phase38RunFixture) { run.Artifacts[0].Attempt++ }},
		{"duplicate-host", "required_logs", func(run *phase38RunFixture) { run.Logs[1].Host = run.Logs[0].Host }},
		{"missing-job", "required_jobs", func(run *phase38RunFixture) { run.Jobs = run.Jobs[:3] }},
		{"missing-artifact", "required_artifacts", func(run *phase38RunFixture) { run.Artifacts = run.Artifacts[:1] }},
		{"zero-attempt", "run_identity", func(run *phase38RunFixture) { run.Attempt = 0 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run := phase38CloneRun(t, fixture.run)
			tc.mutate(&run)
			got := phase38RunBinding(t, fixture, run, fixture.linuxLog, fixture.macosLog)
			phase38RequireRefusal(t, got, tc.reason)
		})
	}

	t.Run("fractional-run-id", func(t *testing.T) {
		phase38RunBinding(t, fixture, fixture.run, fixture.linuxLog, fixture.macosLog)
		raw, err := os.ReadFile(filepath.Join(fixture.root, filepath.FromSlash(fixture.runPath)))
		if err != nil {
			t.Fatal(err)
		}
		mutated := bytes.Replace(raw, []byte(`"run_id": 9001`), []byte(`"run_id": 9001.5`), 1)
		if bytes.Equal(mutated, raw) {
			mutated = bytes.Replace(raw, []byte(`"run_id":9001`), []byte(`"run_id":9001.5`), 1)
		}
		if bytes.Equal(mutated, raw) {
			t.Fatal("could not mutate run ID to a fractional JSON number")
		}
		phase38Write(t, fixture.root, fixture.runPath, mutated)
		phase38RequireRefusal(t, phase38RunBindingFromFiles(t, fixture), "run_identity")
	})

	t.Run("log-bytes-altered-after-hash", func(t *testing.T) {
		phase38RunBinding(t, fixture, fixture.run, fixture.linuxLog, fixture.macosLog)
		path := filepath.Join(fixture.root, filepath.FromSlash(fixture.linuxPath))
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		data[0] = 'x'
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
		phase38RequireRefusal(t, phase38RunBindingFromFiles(t, fixture), "log_bytes_mismatch")
	})
}

func TestPhase38FreezeWorktreeControls(t *testing.T) {
	t.Run("clean-source-and-planning-descendant", func(t *testing.T) {
		fixture := phase38NewBindingFixture(t)
		valid := phase38RunFreeze(t, fixture, fixture.freeze)
		if valid.exit != 0 || !strings.Contains(valid.stdout, "phase38 freeze-preflight: status=pass") {
			t.Fatalf("clean freeze preflight was refused: %#v", valid)
		}
		phase38Write(t, fixture.root, ".planning/only.md", []byte("planning-only\n"))
		phase32Git(t, fixture.root, "add", ".planning/only.md")
		phase32Git(t, fixture.root, "commit", "-qm", "planning-only descendant")
		valid = phase38RunFreeze(t, fixture, fixture.freeze)
		if valid.exit != 0 || !strings.Contains(valid.stdout, "phase38 freeze-preflight: status=pass") {
			t.Fatalf("planning-only descendant invalidated the digest freeze: %#v", valid)
		}
	})

	for _, tc := range []struct {
		name, reason string
		stage        bool
		path         string
	}{
		{"unstaged-digest-input", "digest_worktree_dirty", false, "cmd/probe.go"},
		{"staged-digest-input", "digest_worktree_dirty", true, "cmd/probe.go"},
		{"untracked-digest-input", "digest_worktree_dirty", false, "cmd/new.go"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := phase38NewBindingFixture(t)
			path := tc.path
			if tc.name == "unstaged-digest-input" || tc.name == "staged-digest-input" {
				phase38Write(t, fixture.root, path, []byte("package main\nfunc Changed() {}\n"))
			} else {
				phase38Write(t, fixture.root, path, []byte("package main\n"))
			}
			if tc.stage {
				phase32Git(t, fixture.root, "add", path)
			}
			phase38RequireFreezeRefusal(t, phase38RunFreeze(t, fixture, fixture.freeze), tc.reason)
		})
	}

	t.Run("changed-and-reverted-history", func(t *testing.T) {
		fixture := phase38NewBindingFixture(t)
		phase38Write(t, fixture.root, "cmd/probe.go", []byte("package main\nfunc Probe() int { return 1 }\n"))
		phase32Git(t, fixture.root, "add", "cmd/probe.go")
		phase32Git(t, fixture.root, "commit", "-qm", "change digest input")
		phase38Write(t, fixture.root, "cmd/probe.go", []byte("package main\nfunc Probe() int { return 3421780262 }\n"))
		phase32Git(t, fixture.root, "add", "cmd/probe.go")
		phase32Git(t, fixture.root, "commit", "-qm", "restore digest input")
		phase38RequireFreezeRefusal(t, phase38RunFreeze(t, fixture, fixture.freeze), "freeze_history_changed")
	})

	t.Run("config-hash-and-staging", func(t *testing.T) {
		fixture := phase38NewBindingFixture(t)
		config := mustRead(t, testsupport.ProjectPath(".planning", "config.json"))
		phase38Write(t, fixture.root, ".planning/config.json", config)
		valid := phase38RunFreeze(t, fixture, fixture.freeze)
		if valid.exit != 0 || !strings.Contains(valid.stdout, "phase38 freeze-preflight: status=pass") {
			t.Fatalf("unchanged unstaged config invalidated freeze: %#v", valid)
		}
		phase32Git(t, fixture.root, "add", ".planning/config.json")
		phase38RequireFreezeRefusal(t, phase38RunFreeze(t, fixture, fixture.freeze), "config_staged")
	})
}

func TestPhase38WorkflowFailureControls(t *testing.T) {
	workflow := string(mustRead(t, testsupport.ProjectPath(".github", "workflows", "ci.yml")))
	name := "      - name: scripts/verify-phase32.sh M007\n"
	start := strings.Index(workflow, name)
	if start < 0 {
		t.Fatal("CI workflow is missing the M007 producer step")
	}
	section := workflow[start:]
	if next := strings.Index(section[len(name):], "      - name:"); next >= 0 {
		section = section[:len(name)+next]
	}
	for _, required := range []string{"shell: bash", "set -o pipefail", "sh scripts/verify-phase32.sh --milestone M007 2>&1 | tee"} {
		if !strings.Contains(section, required) {
			t.Fatalf("M007 producer step is missing %q", required)
		}
	}
	uploadName := "      - name: Upload M007 host receipt\n"
	uploadStart := strings.Index(workflow[start+len(name):], uploadName)
	if uploadStart >= 0 {
		uploadStart += start + len(name)
	}
	if uploadStart < 0 {
		t.Fatal("CI workflow is missing the M007 receipt upload")
	}
	upload := workflow[uploadStart:]
	if next := strings.Index(upload[len(uploadName):], "      - name:"); next >= 0 {
		upload = upload[:len(uploadName)+next]
	}
	for _, required := range []string{"if: ${{ always() }}", "${{ github.sha }}-${{ github.run_id }}-${{ github.run_attempt }}", "if-no-files-found: error"} {
		if !strings.Contains(upload, required) {
			t.Fatalf("M007 upload is missing %q", required)
		}
	}

	runLine := "        run: |\n"
	run := strings.Index(section, runLine)
	if run < 0 {
		t.Fatal("M007 step has no run block")
	}
	body := section[run+len(runLine):]
	var commands []string
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "      - name:") {
			break
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		if !strings.HasPrefix(line, "          ") {
			t.Fatalf("unexpected indentation in extracted M007 run block: %q", line)
		}
		commands = append(commands, strings.TrimPrefix(line, "          "))
	}
	script := strings.Join(commands, "\n") + "\n"
	script = strings.ReplaceAll(script, "${{ matrix.os }}", "Linux")
	script = strings.ReplaceAll(script, "${{ github.run_id }}", "9001")
	script = strings.ReplaceAll(script, "${{ github.run_attempt }}", "1")
	work := t.TempDir()
	fakeBin := filepath.Join(work, "bin")
	if err := os.MkdirAll(fakeBin, 0o755); err != nil {
		t.Fatal(err)
	}
	fakeProducer := filepath.Join(fakeBin, "sh")
	if err := os.WriteFile(fakeProducer, []byte("#!/bin/sh\nprintf 'synthetic producer failure\\n'\nexit 29\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	temp := filepath.Join(work, "runner-temp")
	if err := os.MkdirAll(temp, 0o755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("/bin/bash", "-c", script)
	cmd.Dir = work
	cmd.Env = phase38EnvWith(os.Environ(), "PATH", fakeBin+":"+os.Getenv("PATH"))
	cmd.Env = phase38EnvWith(cmd.Env, "RUNNER_TEMP", temp)
	output, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(output), "synthetic producer failure") {
		t.Fatalf("CI pipe masked producer failure: err=%v output=%q script=%q", err, output, script)
	}
	receipt := filepath.Join(temp, "phase38", "receipt-Linux-9001-1.log")
	if data, readErr := os.ReadFile(receipt); readErr != nil || !strings.Contains(string(data), "synthetic producer failure") {
		t.Fatalf("always-upload receipt path was not populated: data=%q err=%v", data, readErr)
	}
}

func phase38CloneRun(t *testing.T, run phase38RunFixture) phase38RunFixture {
	t.Helper()
	data, err := json.Marshal(run)
	if err != nil {
		t.Fatal(err)
	}
	var cloned phase38RunFixture
	if err := json.Unmarshal(data, &cloned); err != nil {
		t.Fatal(err)
	}
	return cloned
}

func phase38RunFreeze(t *testing.T, fixture phase38BindingFixture, freeze phase38FreezeFixture) phase33CommandResult {
	t.Helper()
	phase38WriteJSON(t, fixture.root, fixture.freezePath, freeze)
	return phase33RunAtEnv(t, fixture.root, "scripts/check-milestone-binding.sh", envWith("EXPECTED_CODEDIGEST", fixture.digest), "freeze", fixture.freezePath)
}

func phase38RequireFreezeRefusal(t *testing.T, got phase33CommandResult, reason string) {
	t.Helper()
	if got.exit == 0 || !strings.Contains(got.stdout+got.stderr, "phase38 freeze-preflight: status=fail reason="+reason) {
		t.Fatalf("wanted freeze refusal %q, got exit=%d stdout=%q stderr=%q", reason, got.exit, got.stdout, got.stderr)
	}
}

func phase38EnvWith(env []string, key, value string) []string {
	out := make([]string, 0, len(env)+1)
	prefix := key + "="
	for _, item := range env {
		if !strings.HasPrefix(item, prefix) {
			out = append(out, item)
		}
	}
	return append(out, prefix+value)
}

func phase38ManifestField(value string) string {
	value = strings.TrimSuffix(value, "\n")
	if value == "" {
		return "-"
	}
	return value
}

type phase38FreezeFixture struct {
	Schema       string                       `json:"schema"`
	FreezeSHA    string                       `json:"freeze_sha"`
	CodeDigest   string                       `json:"codedigest"`
	Manifest     phase38ManifestIdentity      `json:"manifest"`
	Repository   string                       `json:"repository"`
	Destination  string                       `json:"destination_ref"`
	WorkflowPath string                       `json:"workflow_path"`
	Inputs       map[string]bool              `json:"inputs"`
	Preparation  map[string]phase38SourceLink `json:"preparation"`
}

type phase38ManifestIdentity struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type phase38SourceLink struct {
	Path       string `json:"path"`
	SHA256     string `json:"sha256"`
	SourceSHA  string `json:"source_sha"`
	CodeDigest string `json:"codedigest"`
}

type phase38RunFixture struct {
	Schema       string                   `json:"schema"`
	Repository   string                   `json:"repository"`
	WorkflowPath string                   `json:"workflow_path"`
	Event        string                   `json:"event"`
	RunID        int64                    `json:"run_id"`
	Attempt      int64                    `json:"attempt"`
	URL          string                   `json:"url"`
	Ref          string                   `json:"ref"`
	HeadSHA      string                   `json:"head_sha"`
	Status       string                   `json:"status"`
	Conclusion   string                   `json:"conclusion"`
	Jobs         []phase38JobFixture      `json:"jobs"`
	Artifacts    []phase38ArtifactFixture `json:"artifacts"`
	Logs         []phase38LogFixture      `json:"logs"`
}

type phase38JobFixture struct {
	ID         int64             `json:"id"`
	Name       string            `json:"name"`
	Conclusion string            `json:"conclusion"`
	Steps      map[string]string `json:"steps,omitempty"`
}

type phase38ArtifactFixture struct {
	Name          string `json:"name"`
	ID            int64  `json:"id"`
	ArchiveDigest string `json:"archive_digest"`
	ExpiresAt     string `json:"expires_at"`
	RunID         int64  `json:"run_id"`
	Attempt       int64  `json:"attempt"`
}

type phase38LogFixture struct {
	ArtifactID int64  `json:"artifact_id"`
	Host       string `json:"host"`
	Path       string `json:"path"`
	Bytes      int64  `json:"bytes"`
	SHA256     string `json:"sha256"`
}

type phase38BindingFixture struct {
	root       string
	digest     string
	freezePath string
	runPath    string
	linuxPath  string
	macosPath  string
	freeze     phase38FreezeFixture
	run        phase38RunFixture
	linuxLog   []byte
	macosLog   []byte
}

func phase38NewBindingFixture(t *testing.T) phase38BindingFixture {
	t.Helper()
	root := t.TempDir()
	for _, dir := range []string{"cmd", "scripts", "testdata/receipts", ".github/workflows", ".planning/phases/38-milestone-receipts"} {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(dir)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	phase38Write(t, root, "cmd/probe.go", []byte("package main\nfunc Probe() int { return 3421780262 }\n"))
	phase38Write(t, root, ".github/workflows/ci.yml", []byte("name: CI\n"))
	phase38Write(t, root, "testdata/receipts/M006.manifest", []byte("label=phase32\nprograms=crc32\ncontrols=smoke\n# case crc32 check 0 3421780262 -\n"))
	phase38Write(t, root, "testdata/receipts/M007.manifest", []byte("label=phase38\nprograms=crc32\ncontrols=smoke\n# case crc32 check 0 3421780262 -\n# root smoke ./cmd/schway ^TestSmoke$\n# event smoke TestSmoke\n# reached crc32 check ./cmd/schway TestSmoke\n"))
	for _, name := range []string{"scripts/codedigest.sh", "scripts/check-freeze.sh", "scripts/check-receipts.sh"} {
		phase38Write(t, root, name, mustRead(t, testsupport.ProjectPath(filepath.FromSlash(name))))
	}
	if source, err := os.ReadFile(testsupport.ProjectPath("scripts", "check-milestone-binding.sh")); err == nil {
		phase38Write(t, root, "scripts/check-milestone-binding.sh", source)
	}
	phase32Git(t, root, "init", "-q")
	phase32Git(t, root, "add", "cmd", ".github", "scripts", "testdata")
	phase32Git(t, root, "commit", "-qm", "synthetic M007 preparation")
	head := phase32GateHead(t, root)
	digest := phase32ShDigest(t, root)
	manifestBytes := mustRead(t, filepath.Join(root, "testdata/receipts/M007.manifest"))
	manifestHash := sha256.Sum256(manifestBytes)
	freeze := phase38FreezeFixture{
		Schema: "schway.m007-freeze/1", FreezeSHA: head, CodeDigest: digest,
		Manifest:   phase38ManifestIdentity{Path: "testdata/receipts/M007.manifest", SHA256: hex.EncodeToString(manifestHash[:])},
		Repository: "szTheory/schway", Destination: "refs/heads/phase26-validation", WorkflowPath: ".github/workflows/ci.yml",
		Inputs: map[string]bool{"record_validation_corpus": false, "phase26_focused": false, "spike008_strict": false},
		Preparation: map[string]phase38SourceLink{
			"review":     {Path: ".planning/phases/38-milestone-receipts/38-PRE-FREEZE-REVIEW.md", SHA256: strings.Repeat("a", 64), SourceSHA: head, CodeDigest: digest},
			"security":   {Path: ".planning/phases/38-milestone-receipts/38-PRE-FREEZE-SECURITY.md", SHA256: strings.Repeat("b", 64), SourceSHA: head, CodeDigest: digest},
			"validation": {Path: ".planning/phases/38-milestone-receipts/38-VALIDATION.md", SHA256: strings.Repeat("c", 64), SourceSHA: head, CodeDigest: digest},
			"audit":      {Path: ".planning/M007-MILESTONE-AUDIT.md", SHA256: strings.Repeat("d", 64), SourceSHA: head, CodeDigest: digest},
		},
	}
	linux := phase38SyntheticLog("Linux", head, digest)
	macos := phase38SyntheticLog("macOS", head, digest)
	linuxHash, macosHash := sha256.Sum256(linux), sha256.Sum256(macos)
	run := phase38RunFixture{
		Schema: "schway.m007-hosted-run/1", Repository: freeze.Repository, WorkflowPath: freeze.WorkflowPath,
		Event: "workflow_dispatch", RunID: 9001, Attempt: 1,
		URL: "https://github.com/szTheory/schway/actions/runs/9001/attempts/1",
		Ref: freeze.Destination, HeadSHA: head, Status: "completed", Conclusion: "success",
		Jobs: []phase38JobFixture{
			{ID: 11, Name: "checks (ubuntu-latest)", Conclusion: "success", Steps: map[string]string{"full": "success", "race": "success"}},
			{ID: 12, Name: "checks (macos-latest)", Conclusion: "success", Steps: map[string]string{"full": "success", "race": "success"}},
			{ID: 13, Name: "current evidence aggregate (ubuntu-latest)", Conclusion: "success"},
			{ID: 14, Name: "current evidence aggregate (macos-latest)", Conclusion: "success"},
		},
		Artifacts: []phase38ArtifactFixture{
			{Name: "phase38-receipt-ubuntu-latest-" + head + "-9001-1", ID: 21, ArchiveDigest: "sha256:" + strings.Repeat("e", 64), ExpiresAt: time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339), RunID: 9001, Attempt: 1},
			{Name: "phase38-receipt-macos-latest-" + head + "-9001-1", ID: 22, ArchiveDigest: "sha256:" + strings.Repeat("f", 64), ExpiresAt: time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339), RunID: 9001, Attempt: 1},
		},
		Logs: []phase38LogFixture{
			{ArtifactID: 21, Host: "Linux", Path: ".planning/phases/38-milestone-receipts/receipt-ubuntu-latest.log", Bytes: int64(len(linux)), SHA256: hex.EncodeToString(linuxHash[:])},
			{ArtifactID: 22, Host: "macOS", Path: ".planning/phases/38-milestone-receipts/receipt-macos-latest.log", Bytes: int64(len(macos)), SHA256: hex.EncodeToString(macosHash[:])},
		},
	}
	fixture := phase38BindingFixture{
		root: root, digest: digest,
		freezePath: ".planning/phases/38-milestone-receipts/38-FREEZE.json",
		runPath:    ".planning/phases/38-milestone-receipts/38-HOST-RUN.json",
		linuxPath:  ".planning/phases/38-milestone-receipts/receipt-ubuntu-latest.log",
		macosPath:  ".planning/phases/38-milestone-receipts/receipt-macos-latest.log",
		freeze:     freeze, run: run, linuxLog: linux, macosLog: macos,
	}
	phase38WriteJSON(t, root, fixture.freezePath, freeze)
	phase38WriteJSON(t, root, fixture.runPath, run)
	phase38Write(t, root, fixture.linuxPath, linux)
	phase38Write(t, root, fixture.macosPath, macos)
	return fixture
}

func phase38SyntheticLog(host, headSHA, digest string) []byte {
	arch, target, goHost := "x86_64", "x86_64-pc-linux-gnu", "linux/amd64"
	if host == "macOS" {
		arch, target, goHost = "arm64", "arm64-apple-darwin", "darwin/arm64"
	}
	return []byte(fmt.Sprintf("phase38 host identity host=%s arch=%s go=%s clang_target=%s headsha=%s codedigest=%s tree=clean\n", host, arch, goHost, target, headSHA, digest) +
		"phase38 toolchain go=\"go version go1.24.0\" clang=\"clang version\"\n" +
		"phase38-control-events-hex smoke " + hex.EncodeToString(phase38SyntheticEvents("run", "output", "pass")) + "\n" +
		"phase38 case program=crc32 input=check expected_exit=0 actual_exit=0 stdout=exact stderr=exact status=pass\n" +
		"phase38 control name=smoke package=./cmd/schway tests=TestSmoke status=pass\n" +
		fmt.Sprintf("phase38 host result host=%s codedigest=%s tree=clean cases=1 controls=1 programs=crc32 status=pass\n", host, digest))
}

func phase38SyntheticEvents(actions ...string) []byte {
	var lines []string
	for _, action := range actions {
		if action == "extra" {
			lines = append(lines, `{"Action":"run","Test":"TestOther"}`, `{"Action":"pass","Test":"TestOther"}`)
			continue
		}
		if action == "output" {
			lines = append(lines, `{"Action":"output","Test":"TestSmoke","Output":"interpreter-with-clang-unavailable"}`)
			continue
		}
		lines = append(lines, `{"Action":"`+action+`","Test":"TestSmoke"}`)
	}
	return []byte(strings.Join(lines, "\n") + "\n")
}

func phase38RunBinding(t *testing.T, fixture phase38BindingFixture, run phase38RunFixture, linux, macos []byte) phase33CommandResult {
	return phase38RunBindingWithOrder(t, fixture, run, linux, macos, false)
}

func phase38RunBindingWithOrder(t *testing.T, fixture phase38BindingFixture, run phase38RunFixture, linux, macos []byte, reverse bool) phase33CommandResult {
	t.Helper()
	phase38WriteJSON(t, fixture.root, fixture.runPath, run)
	phase38Write(t, fixture.root, fixture.linuxPath, linux)
	phase38Write(t, fixture.root, fixture.macosPath, macos)
	logs := map[string][]byte{"Linux": linux, "macOS": macos}
	for i := range run.Logs {
		data := logs[run.Logs[i].Host]
		digest := sha256.Sum256(data)
		run.Logs[i].Bytes, run.Logs[i].SHA256 = int64(len(data)), hex.EncodeToString(digest[:])
	}
	phase38WriteJSON(t, fixture.root, fixture.runPath, run)
	firstPath, secondPath := fixture.linuxPath, fixture.macosPath
	if reverse {
		firstPath, secondPath = secondPath, firstPath
	}
	return phase33RunAtEnv(t, fixture.root, "scripts/check-milestone-binding.sh", envWith("EXPECTED_CODEDIGEST", fixture.digest),
		"run", fixture.freezePath, fixture.runPath, firstPath, secondPath)
}

func phase38RunBindingFromFiles(t *testing.T, fixture phase38BindingFixture) phase33CommandResult {
	t.Helper()
	return phase33RunAtEnv(t, fixture.root, "scripts/check-milestone-binding.sh", envWith("EXPECTED_CODEDIGEST", fixture.digest),
		"run", fixture.freezePath, fixture.runPath, fixture.linuxPath, fixture.macosPath)
}

func phase38RequireRefusal(t *testing.T, got phase33CommandResult, reason string) {
	t.Helper()
	if got.exit == 0 || !strings.Contains(got.stdout+got.stderr, "phase38 binding: status=fail reason="+reason) {
		t.Fatalf("wanted binding refusal %q, got exit=%d stdout=%q stderr=%q", reason, got.exit, got.stdout, got.stderr)
	}
}

func phase38Write(t *testing.T, root, name string, data []byte) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func phase38WriteJSON(t *testing.T, root, name string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	phase38Write(t, root, name, append(data, '\n'))
}
