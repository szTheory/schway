package cache

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeExecutableFixture(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "fixture-clang.sh")
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeMutationRunnerFixture(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "runner.go")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// writeCgenSourceFixture materializes a tiny synthetic "cgen source
// directory" containing one .go file with the given content, standing in
// for internal/compiler/cgen/*.go without this test depending on the real
// package's current contents.
func writeCgenSourceFixture(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "cgen.go"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func baseArtifactSpec(t *testing.T) ArtifactSpec {
	t.Helper()
	dir := t.TempDir()
	return ArtifactSpec{
		Kind:                     KindCompiledBinary,
		FixtureSource:            []byte("fn main() {}\n"),
		BuildFlags:               "-O0 target=x86_64-apple-darwin",
		ClangPath:                writeExecutableFixture(t, dir, "#!/bin/sh\necho fixture clang version 1.0\n"),
		RuntimeIdentity:          "none",
		ForeignTranslationUnit:   []byte("// frozen TU v1\n"),
		MutationRunnerSourcePath: writeMutationRunnerFixture(t, dir, "package fake\n// mutation runner fixture v1\n"),
		GoToolchain:              "go1.24.0",
		CgenSourceDir:            writeCgenSourceFixture(t, "package cgen\n// cgen source fixture v1\n"),
	}
}

func TestCacheKeyCoversEveryDeclaredInput(t *testing.T) {
	names := DeclaredInputNames()
	want := []string{
		"fixture_source", "build_flags", "clang_identity", "runtime_identity",
		"foreign_translation_unit", "mutation_runner_source", "go_toolchain",
		"cgen_source",
	}
	if len(names) != len(want) {
		t.Fatalf("DeclaredInputNames() = %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("DeclaredInputNames()[%d] = %q, want %q (order is D-06-07's determinism guarantee)", i, names[i], want[i])
		}
	}

	ctx := context.Background()
	base := baseArtifactSpec(t)
	baseInputs, err := InputsFor(ctx, base)
	if err != nil {
		t.Fatal(err)
	}
	baseKey, err := ComputeKey(baseInputs)
	if err != nil {
		t.Fatal(err)
	}

	covered := map[string]bool{}
	cases := []struct {
		name   string
		mutate func(t *testing.T, spec *ArtifactSpec)
	}{
		{"fixture_source", func(t *testing.T, spec *ArtifactSpec) {
			spec.FixtureSource = []byte("fn main() { changed() }\n")
		}},
		{"build_flags", func(t *testing.T, spec *ArtifactSpec) {
			spec.BuildFlags = "-O3 target=x86_64-apple-darwin"
		}},
		{"clang_identity", func(t *testing.T, spec *ArtifactSpec) {
			spec.ClangPath = writeExecutableFixture(t, t.TempDir(), "#!/bin/sh\necho fixture clang version 2.0\n")
		}},
		{"runtime_identity", func(t *testing.T, spec *ArtifactSpec) {
			spec.RuntimeIdentity = "asan+ubsan+libc++"
		}},
		{"foreign_translation_unit", func(t *testing.T, spec *ArtifactSpec) {
			spec.ForeignTranslationUnit = []byte("// frozen TU v2\n")
		}},
		{"mutation_runner_source", func(t *testing.T, spec *ArtifactSpec) {
			spec.MutationRunnerSourcePath = writeMutationRunnerFixture(t, t.TempDir(), "package fake\n// mutation runner fixture v2\n")
		}},
		{"go_toolchain", func(t *testing.T, spec *ArtifactSpec) {
			spec.GoToolchain = "go1.99.0"
		}},
		{"cgen_source", func(t *testing.T, spec *ArtifactSpec) {
			spec.CgenSourceDir = writeCgenSourceFixture(t, "package cgen\n// cgen source fixture CHANGED\n")
		}},
	}
	for _, testCase := range cases {
		covered[testCase.name] = true
		t.Run(testCase.name, func(t *testing.T) {
			mutated := base
			testCase.mutate(t, &mutated)
			inputs, err := InputsFor(ctx, mutated)
			if err != nil {
				t.Fatal(err)
			}
			key, err := ComputeKey(inputs)
			if err != nil {
				t.Fatal(err)
			}
			if key.ID == baseKey.ID {
				t.Fatalf("perturbing declared input %q did not change the key", testCase.name)
			}
		})
	}
	// Self-invalidating: a declared input name with no perturbation row above
	// must fail this test, so DeclaredInputNames() can never silently grow
	// past this table's coverage.
	for _, name := range names {
		if !covered[name] {
			t.Fatalf("declared input %q has no perturbation row in this table", name)
		}
	}
}

func TestMutationRunnerSourceHashIsADeclaredInput(t *testing.T) {
	ctx := context.Background()
	base := baseArtifactSpec(t)
	baseInputs, err := InputsFor(ctx, base)
	if err != nil {
		t.Fatal(err)
	}
	baseKey, err := ComputeKey(baseInputs)
	if err != nil {
		t.Fatal(err)
	}

	changed := base
	changed.MutationRunnerSourcePath = writeMutationRunnerFixture(t, t.TempDir(), "package fake\n// mutation runner fixture CHANGED\n")
	changedInputs, err := InputsFor(ctx, changed)
	if err != nil {
		t.Fatal(err)
	}
	changedKey, err := ComputeKey(changedInputs)
	if err != nil {
		t.Fatal(err)
	}
	if baseKey.ID == changedKey.ID {
		t.Fatal("a mutant binary built by a changed mutation runner must miss the cache")
	}
}

func selfExecFactory(mode string) commandFactory {
	return func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=TestCacheProbeHelper", "--", mode)
		cmd.Env = append(os.Environ(), "GO_WANT_CACHE_PROBE_HELPER=1")
		return cmd
	}
}

func TestClangDigestIsNotJustTheVersionString(t *testing.T) {
	dir := t.TempDir()
	binaryA := filepath.Join(dir, "clang-a")
	binaryB := filepath.Join(dir, "clang-b")
	if err := os.WriteFile(binaryA, []byte("binary-bytes-a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binaryB, []byte("binary-bytes-b"), 0o644); err != nil {
		t.Fatal(err)
	}

	factory := selfExecFactory("same-version")
	digestA, err := probeClangIdentityWithCommand(context.Background(), binaryA, factory)
	if err != nil {
		t.Fatal(err)
	}
	digestB, err := probeClangIdentityWithCommand(context.Background(), binaryB, factory)
	if err != nil {
		t.Fatal(err)
	}
	if digestA == digestB {
		t.Fatal("two probes reporting the same version but different binary bytes must produce different digests")
	}
}

func TestClangDigestProbeIsBounded(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "fake-clang")
	if err := os.WriteFile(binary, []byte("fake-clang-bytes"), 0o644); err != nil {
		t.Fatal(err)
	}

	for _, testCase := range []struct {
		name     string
		mode     string
		wantCode string
		timeout  time.Duration
	}{
		{name: "stdout truncated", mode: "stdout", wantCode: "cache.probe_stdout_truncated"},
		{name: "stderr truncated", mode: "stderr", wantCode: "cache.probe_stderr_truncated"},
		{name: "timeout", mode: "block", wantCode: "cache.probe_timeout", timeout: 50 * time.Millisecond},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			if testCase.timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, testCase.timeout)
				defer cancel()
			}
			_, err := probeClangIdentityWithCommand(ctx, binary, selfExecFactory(testCase.mode))
			var probeErr *Error
			if !errors.As(err, &probeErr) || probeErr.Code != testCase.wantCode {
				t.Fatalf("error=%v want code=%s", err, testCase.wantCode)
			}
		})
	}
}

// TestCacheProbeHelper is a self-exec test helper (evidence's
// TestEvidenceToolProbeHelper precedent): it only does anything when
// launched as a subprocess by selfExecFactory above via os.Args[0].
func TestCacheProbeHelper(t *testing.T) {
	if os.Getenv("GO_WANT_CACHE_PROBE_HELPER") != "1" {
		return
	}
	mode := os.Args[len(os.Args)-1]
	switch mode {
	case "stdout":
		_, _ = os.Stdout.WriteString(strings.Repeat("x", MaxProbeBytes+1))
	case "stderr":
		_, _ = os.Stderr.WriteString(strings.Repeat("x", MaxProbeBytes+1))
	case "block":
		time.Sleep(10 * time.Second)
	default:
		_, _ = os.Stdout.WriteString("fixture clang version 1.0\n")
	}
	os.Exit(0)
}

