package session_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func phase33FreezeRun(t *testing.T, cwd, script string, args ...string) (int, string, string) {
	t.Helper()
	full := append([]string{"-c", `cd "$0" && script=$1 && shift && exec sh "$script" "$@"`, cwd, script}, args...)
	result, err := testsupport.RunCLIErr(context.Background(), "sh", nil, full...)
	if err != nil && result.Exit == 0 {
		t.Fatalf("spawn: %v", err)
	}
	return result.Exit, string(result.Stdout), string(result.Stderr)
}

func TestFreezeGate(t *testing.T) {
	root := testsupport.ProjectPath()
	script, err := os.ReadFile(testsupport.ProjectPath("scripts", "check-freeze.sh"))
	if err != nil {
		t.Fatal(err)
	}
	digest, err := os.ReadFile(testsupport.ProjectPath("scripts", "codedigest.sh"))
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := os.ReadFile(testsupport.ProjectPath("testdata", "receipts", "M006.manifest"))
	if err != nil {
		t.Fatal(err)
	}
	write := func(dir, name string, content []byte) {
		t.Helper()
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, content, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	build := func(t *testing.T) (string, string, string, string) {
		dir := t.TempDir()
		write(dir, "cmd/a.go", []byte("package a\n"))
		write(dir, "scripts/check-freeze.sh", script)
		write(dir, "scripts/codedigest.sh", digest)
		write(dir, "testdata/receipts/M006.manifest", manifest)
		phase32Git(t, dir, "init", "-q")
		phase32Git(t, dir, "add", "-A")
		phase32Git(t, dir, "commit", "-qm", "base")
		base := phase32GateHead(t, dir)
		write(dir, "cmd/a.go", []byte("package a // frozen\n"))
		phase32Git(t, dir, "commit", "-qam", "freeze")
		freeze := phase32GateHead(t, dir)
		write(dir, ".planning/only.md", []byte("plan\n"))
		phase32Git(t, dir, "add", "-A")
		phase32Git(t, dir, "commit", "-qm", "planning")
		plan := phase32GateHead(t, dir)
		write(dir, "cmd/a.go", []byte("package a // changed\n"))
		phase32Git(t, dir, "commit", "-qam", "code")
		code := phase32GateHead(t, dir)
		return dir, base, freeze, plan + "\n" + code
	}
	gate := filepath.Join(root, "scripts", "check-freeze.sh")
	t.Run("zero-args-usage", func(t *testing.T) {
		exit, out, stderr := phase33FreezeRun(t, root, gate)
		if exit != 1 || out != "" || stderr != "freeze gate: status=fail reason=usage\n" {
			t.Fatalf("exit=%d stdout=%q stderr=%q", exit, out, stderr)
		}
	})
	t.Run("manifest-missing", func(t *testing.T) {
		exit, out, stderr := phase33FreezeRun(t, root, gate, "/nonexistent.manifest", strings.Repeat("0", 40), strings.Repeat("0", 40))
		if exit != 1 || out != "" || stderr != "freeze gate: status=fail reason=manifest_missing\n" {
			t.Fatalf("exit=%d stdout=%q stderr=%q", exit, out, stderr)
		}
	})
	t.Run("label-less-manifest", func(t *testing.T) {
		manifest := filepath.Join(t.TempDir(), "manifest")
		if err := os.WriteFile(manifest, []byte("programs=x\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		exit, out, stderr := phase33FreezeRun(t, root, gate, manifest, strings.Repeat("0", 40), strings.Repeat("0", 40))
		if exit != 1 || out != "" || stderr != "freeze gate: status=fail reason=manifest_missing\n" {
			t.Fatalf("exit=%d stdout=%q stderr=%q", exit, out, stderr)
		}
	})
	t.Run("no-repo-root", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, "scripts"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "scripts", "check-freeze.sh"), script, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "manifest"), manifest, 0o600); err != nil {
			t.Fatal(err)
		}
		exit, out, stderr := phase33FreezeRun(t, dir, "scripts/check-freeze.sh", "manifest", strings.Repeat("0", 40), strings.Repeat("0", 40))
		if exit != 1 || out != "" || stderr != "phase32 freeze gate: status=fail reason=no_repo_root\n" {
			t.Fatalf("exit=%d stdout=%q stderr=%q", exit, out, stderr)
		}
	})
	t.Run("paths-match-digest", func(t *testing.T) {
		result, err := testsupport.RunCLIErr(context.Background(), "sh", nil, testsupport.ProjectPath("scripts", "codedigest.sh"), "--paths")
		if err != nil || result.Exit != 0 || strings.Join(strings.Fields(string(result.Stdout)), " ") != strings.Join(phase32DigestPaths, " ") {
			t.Fatalf("paths=%q err=%v", result.Stdout, err)
		}
	})
	t.Run("hash-length-derived", func(t *testing.T) {
		got := phase32ShDigest(t, root)
		empty, err := testsupport.RunCLIErr(context.Background(), "sh", nil, "-c", `if command -v shasum >/dev/null 2>&1; then printf '' | shasum -a 256; else printf '' | sha256sum; fi`)
		if err != nil || empty.Exit != 0 {
			t.Fatalf("empty-input hash: exit=%d err=%v", empty.Exit, err)
		}
		emptyHash := strings.Fields(string(empty.Stdout))[0]
		if len(got) != len(emptyHash) || len(got) != 64 {
			t.Fatalf("digest length=%d, empty-input hasher length=%d, want 64", len(got), len(emptyHash))
		}
	})
	t.Run("m006-real-repo", func(t *testing.T) {
		shas := []string{"0aac0a7b4801292b3f54f7ef57ed5fb83374c08e", "9f8f878fc5328f953eac6ed04a716478ae986fab", "31d71e9428eeb3aa89cd376b2a6fd5047bb4e69d", "7ab2f48e28f2401f0545e9c414c496795651090b"}
		for _, sha := range shas {
			result, err := testsupport.RunCLIErr(context.Background(), "git", nil, "-C", root, "cat-file", "-e", sha+"^{commit}")
			if err != nil || result.Exit != 0 {
				t.Fatalf("missing pinned commit %s; run git fetch --tags: %v %s", sha, err, result.Stderr)
			}
		}
		cases := []struct{ from, to, reason string }{{shas[0], shas[1], "pass"}, {shas[0], shas[2], "digest_path_changed_after_freeze"}, {shas[0], shas[3], "digest_path_changed_after_freeze"}}
		for _, tc := range cases {
			exit, out, stderr := phase33FreezeRun(t, root, gate, filepath.Join(root, "testdata/receipts/M006.manifest"), tc.from, tc.to)
			if tc.reason == "pass" {
				if exit != 0 || !strings.Contains(out, "status=pass") {
					t.Fatalf("expected pass: %d %q %q", exit, out, stderr)
				}
			} else if exit == 0 || !strings.Contains(stderr, "reason="+tc.reason) {
				t.Fatalf("expected %s: %d %q", tc.reason, exit, stderr)
			}
		}
		headResult, err := testsupport.RunCLIErr(context.Background(), "git", nil, "-C", root, "rev-parse", "HEAD")
		if err != nil || headResult.Exit != 0 {
			t.Fatalf("project HEAD: %v", err)
		}
		head := strings.TrimSpace(string(headResult.Stdout))
		exit, _, stderr := phase33FreezeRun(t, root+"/cmd", "../scripts/check-freeze.sh", "../testdata/receipts/M006.manifest", shas[0], head)
		if exit == 0 || !strings.Contains(stderr, "reason=digest_path_changed_after_freeze") {
			t.Fatalf("M006 freeze to project HEAD from cmd: exit=%d stderr=%s", exit, stderr)
		}
	})
	f, base, freeze, combined := build(t)
	commits := strings.Split(combined, "\n")
	plan, code := commits[0], commits[1]
	t.Run("same-commit-accepted", func(t *testing.T) {
		exit, out, stderr := phase33FreezeRun(t, f, "scripts/check-freeze.sh", "testdata/receipts/M006.manifest", freeze, freeze, freeze)
		if exit != 0 || !strings.Contains(out, "status=pass") {
			t.Fatalf("%d %q %q", exit, out, stderr)
		}
	})
	t.Run("same-as-mismatch", func(t *testing.T) {
		exit, _, stderr := phase33FreezeRun(t, f, "scripts/check-freeze.sh", "testdata/receipts/M006.manifest", freeze, plan, freeze)
		if exit == 0 || !strings.Contains(stderr, "reason=sha_mismatch") {
			t.Fatalf("%d %q", exit, stderr)
		}
	})
	t.Run("planning-only-and-code-change", func(t *testing.T) {
		exit, out, _ := phase33FreezeRun(t, f, "scripts/check-freeze.sh", "testdata/receipts/M006.manifest", freeze, plan)
		if exit != 0 || !strings.Contains(out, "status=pass") {
			t.Fatalf("planning-only failed: %d %q", exit, out)
		}
		exit, _, stderr := phase33FreezeRun(t, f, "scripts/check-freeze.sh", "testdata/receipts/M006.manifest", freeze, code)
		if exit == 0 || !strings.Contains(stderr, "reason=digest_path_changed_after_freeze") {
			t.Fatalf("code change accepted: %d %q", exit, stderr)
		}
	})
	t.Run("md-only-change-after-freeze-accepted", func(t *testing.T) {
		dir, _, freeze, _ := build(t)
		phase32Git(t, dir, "checkout", "-q", freeze)
		path := filepath.Join(dir, "cmd", "README.md")
		if err := os.WriteFile(path, []byte("documentation only\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		phase32Git(t, dir, "add", "-A")
		phase32Git(t, dir, "commit", "-qm", "markdown only")
		candidate := phase32GateHead(t, dir)
		changed := phase33RunAt(t, dir, "scripts/codedigest.sh", "--changed", freeze, candidate)
		if changed.exit != 0 || changed.stdout != "" {
			t.Fatalf("markdown-only changed paths: %#v", changed)
		}
		gate := phase33RunAt(t, dir, "scripts/check-freeze.sh", "testdata/receipts/M006.manifest", freeze, candidate)
		if gate.exit != 0 || !strings.Contains(gate.stdout, "status=pass") {
			t.Fatalf("markdown-only freeze rejected: %#v", gate)
		}
	})
	t.Run("non-ascii-md-path-filtered", func(t *testing.T) {
		dir, _, freeze, _ := build(t)
		phase32Git(t, dir, "checkout", "-q", freeze)
		path := filepath.Join(dir, "cmd", "résumé.md")
		if err := os.WriteFile(path, []byte("documentation only\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		phase32Git(t, dir, "add", "-A")
		phase32Git(t, dir, "commit", "-qm", "non-ascii markdown")
		candidate := phase32GateHead(t, dir)
		changed := phase33RunAt(t, dir, "scripts/codedigest.sh", "--changed", freeze, candidate)
		if changed.exit != 0 || changed.stdout != "" {
			t.Fatalf("non-ASCII markdown path was not filtered: %#v", changed)
		}
	})
	t.Run("revert-inside-range-rejected", func(t *testing.T) {
		dir, _, freeze, combined := build(t)
		code := strings.Split(combined, "\n")[1]
		file := filepath.Join(dir, "cmd", "a.go")
		if err := os.WriteFile(file, []byte("package a // frozen\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		phase32Git(t, dir, "add", "-A")
		phase32Git(t, dir, "commit", "-qm", "revert in range")
		candidate := phase32GateHead(t, dir)
		netDiff, err := testsupport.RunCLIErr(context.Background(), "git", nil, "-C", dir, "diff", "--quiet", freeze, candidate, "--", "cmd/a.go")
		if err != nil || netDiff.Exit != 0 {
			t.Fatalf("fixture should have an empty net code diff after %s: exit=%d err=%v", code, netDiff.Exit, err)
		}
		diff := phase33RunAt(t, dir, "scripts/codedigest.sh", "--changed", freeze, candidate)
		if diff.exit != 0 || diff.stdout == "" {
			t.Fatalf("reverted range was not observed in history: %#v", diff)
		}
		gate := phase33RunAt(t, dir, "scripts/check-freeze.sh", "testdata/receipts/M006.manifest", freeze, candidate)
		if gate.exit == 0 || !strings.Contains(gate.stderr, "reason=digest_path_changed_after_freeze") {
			t.Fatalf("reverted code change was accepted: %#v", gate)
		}
	})
	t.Run("subdirectory", func(t *testing.T) {
		phase32Git(t, f, "checkout", "-q", code)
		exit, _, stderr := phase33FreezeRun(t, f+"/cmd", "../scripts/check-freeze.sh", "../testdata/receipts/M006.manifest", freeze, code)
		if exit == 0 || !strings.Contains(stderr, "reason=digest_path_changed_after_freeze") {
			t.Fatalf("subdirectory verdict: %d %q", exit, stderr)
		}
	})
	t.Run("base-witness", func(t *testing.T) {
		if base == "" {
			t.Fatal("empty fixture base")
		}
	})
}
