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

type phase38FreezeFixture struct {
	Schema        string                      `json:"schema"`
	FreezeSHA     string                      `json:"freeze_sha"`
	CodeDigest    string                      `json:"codedigest"`
	Manifest      phase38ManifestIdentity    `json:"manifest"`
	Repository    string                      `json:"repository"`
	Destination   string                      `json:"destination_ref"`
	WorkflowPath  string                      `json:"workflow_path"`
	Inputs        map[string]bool             `json:"inputs"`
	Preparation   map[string]phase38SourceLink `json:"preparation"`
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
	Schema       string                 `json:"schema"`
	Repository   string                 `json:"repository"`
	WorkflowPath string                 `json:"workflow_path"`
	Event        string                 `json:"event"`
	RunID        int64                  `json:"run_id"`
	Attempt      int64                  `json:"attempt"`
	URL          string                 `json:"url"`
	Ref          string                 `json:"ref"`
	HeadSHA      string                 `json:"head_sha"`
	Status       string                 `json:"status"`
	Conclusion   string                 `json:"conclusion"`
	Jobs         []phase38JobFixture    `json:"jobs"`
	Artifacts    []phase38ArtifactFixture `json:"artifacts"`
	Logs         []phase38LogFixture    `json:"logs"`
}

type phase38JobFixture struct {
	ID         int64             `json:"id"`
	Name       string            `json:"name"`
	Conclusion string            `json:"conclusion"`
	Steps      map[string]string `json:"steps,omitempty"`
}

type phase38ArtifactFixture struct {
	Name         string `json:"name"`
	ID           int64  `json:"id"`
	ArchiveDigest string `json:"archive_digest"`
	ExpiresAt    string `json:"expires_at"`
	RunID        int64  `json:"run_id"`
	Attempt      int64  `json:"attempt"`
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
	phase38Write(t, root, "testdata/receipts/M007.manifest", []byte("label=phase38\nprograms=crc32\ncontrols=smoke\n# case crc32 check 0 3421780262 -\n"))
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
		Manifest: phase38ManifestIdentity{Path: "testdata/receipts/M007.manifest", SHA256: hex.EncodeToString(manifestHash[:])},
		Repository: "szTheory/schway", Destination: "refs/heads/phase26-validation", WorkflowPath: ".github/workflows/ci.yml",
		Inputs: map[string]bool{"record_validation_corpus": false, "phase26_focused": false, "spike008_strict": false},
		Preparation: map[string]phase38SourceLink{
			"review": {Path: ".planning/phases/38-milestone-receipts/38-PRE-FREEZE-REVIEW.md", SHA256: strings.Repeat("a", 64), SourceSHA: head, CodeDigest: digest},
			"security": {Path: ".planning/phases/38-milestone-receipts/38-PRE-FREEZE-SECURITY.md", SHA256: strings.Repeat("b", 64), SourceSHA: head, CodeDigest: digest},
			"validation": {Path: ".planning/phases/38-milestone-receipts/38-VALIDATION.md", SHA256: strings.Repeat("c", 64), SourceSHA: head, CodeDigest: digest},
			"audit": {Path: ".planning/M007-MILESTONE-AUDIT.md", SHA256: strings.Repeat("d", 64), SourceSHA: head, CodeDigest: digest},
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
		runPath: ".planning/phases/38-milestone-receipts/38-HOST-RUN.json",
		linuxPath: ".planning/phases/38-milestone-receipts/receipt-ubuntu-latest.log",
		macosPath: ".planning/phases/38-milestone-receipts/receipt-macos-latest.log",
		freeze: freeze, run: run, linuxLog: linux, macosLog: macos,
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
		"phase38 case program=crc32 input=check expected_exit=0 actual_exit=0 stdout=exact stderr=exact status=pass\n" +
		"phase38 control name=smoke package=./cmd/schway tests=TestSmoke status=pass\n" +
		fmt.Sprintf("phase38 host result host=%s codedigest=%s tree=clean cases=1 controls=1 programs=crc32 status=pass\n", host, digest))
}

func phase38RunBinding(t *testing.T, fixture phase38BindingFixture, run phase38RunFixture, linux, macos []byte) phase33CommandResult {
	t.Helper()
	phase38WriteJSON(t, fixture.root, fixture.runPath, run)
	phase38Write(t, fixture.root, fixture.linuxPath, linux)
	phase38Write(t, fixture.root, fixture.macosPath, macos)
	linuxHash, macosHash := sha256.Sum256(linux), sha256.Sum256(macos)
	run.Logs[0].Bytes, run.Logs[0].SHA256 = int64(len(linux)), hex.EncodeToString(linuxHash[:])
	run.Logs[1].Bytes, run.Logs[1].SHA256 = int64(len(macos)), hex.EncodeToString(macosHash[:])
	phase38WriteJSON(t, fixture.root, fixture.runPath, run)
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
