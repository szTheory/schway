package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// TestPhase6CorpusDispatchRequiresMarker asserts isPhase6Corpus recognizes
// testdata/phase6 by its own characteristic marker fixture
// (heldout_match_defect.lang, per testdata/phase6/README) and returns
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
	if _, err := os.Stat("../../testdata/phase6/heldout_match_defect.lang"); err != nil {
		t.Fatalf("testdata/phase6's own marker fixture is missing: %v", err)
	}
}

// TestPhase7CorpusDispatchRequiresMarker asserts isPhase7Corpus recognizes
// testdata/phase07 by its own characteristic marker fixture
// (call_basic.lang, the Phase 07 tracer fixture) and returns false for
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
	if _, err := os.Stat("../../testdata/phase07/call_basic.lang"); err != nil {
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
	if expected.Schema != "lang.phase22-identity-expected/1" || len(expected.Cases) != 2 {
		t.Fatalf("independent expected-output fixture has wrong identity or cases: %+v", expected)
	}
	source := testsupport.ProjectPath("examples", "phase22", "identity.lang")
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

func TestPhase22AppRunKeepsOpaqueTokenStreamsAndChildStatus(t *testing.T) {
	artifact := writeCLIPhase22Script(t, "printf 'stdout:%s {not protocol}\\n' \"$1\"\nprintf 'stderr:%s\\n' \"$1\" >&2\nexit 19\n")
	code, stdout, stderr := captureLangRun(t, []string{"app", "run", artifact, "--", "--json"})
	if code != 19 || stdout != "stdout:--json {not protocol}\n" || stderr != "stderr:--json\n" {
		t.Fatalf("opaque app child code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
}

func TestPhase22ConformanceRunStillEmitsExecutionDocument(t *testing.T) {
	fixture := testsupport.ProjectPath("testdata", "phase19", "literal_tracer.lang")
	code, stdout, stderr := captureLangRun(t, []string{"--json", "run", "--engine=native", fixture})
	if code != 0 || stderr != "" || !strings.Contains(stdout, `"schema":"lang.execution/2"`) {
		t.Fatalf("conformance run code=%d stdout=%q stderr=%q, want its existing execution document", code, stdout, stderr)
	}
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
	if err := os.WriteFile(artifact+".lang-build.json", encoded, 0o600); err != nil {
		t.Fatal(err)
	}
	return artifact
}
