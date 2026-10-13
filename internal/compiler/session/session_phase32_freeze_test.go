package session_test

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func phase32FreezeRun(t *testing.T, repo string, args ...string) (int, string) {
	t.Helper()
	full := append([]string{"-c", `cd "$0" && sh scripts/check-phase32-freeze.sh "$@"`, repo}, args...)
	result, err := testsupport.RunCLIErr(context.Background(), "sh", nil, full...)
	if err != nil && result.Exit == 0 {
		t.Fatalf("spawn: %v", err)
	}
	return result.Exit, string(result.Stdout) + string(result.Stderr)
}

func TestPhase32FreezeGate(t *testing.T) {
	root := testsupport.ProjectPath()
	freezeScript, err := os.ReadFile(testsupport.ProjectPath("scripts", "check-phase32-freeze.sh"))
	if err != nil {
		t.Fatal(err)
	}
	digestScript, err := os.ReadFile(testsupport.ProjectPath("scripts", "codedigest.sh"))
	if err != nil {
		t.Fatal(err)
	}

	expectPass := func(t *testing.T, repo string, args ...string) {
		t.Helper()
		exit, out := phase32FreezeRun(t, repo, args...)
		if exit != 0 || !strings.Contains(out, "phase32 freeze gate: status=pass") {
			t.Fatalf("expected pass, exit=%d out=%s", exit, out)
		}
	}
	expectReject := func(t *testing.T, repo, reason string, args ...string) {
		t.Helper()
		exit, out := phase32FreezeRun(t, repo, args...)
		if exit == 0 || strings.Contains(out, "status=pass") || !strings.Contains(out, "reason="+reason) {
			t.Fatalf("expected reason=%s, exit=%d out=%s", reason, exit, out)
		}
	}

	t.Run("real-repo", func(t *testing.T) {
		result, err := testsupport.RunCLIErr(context.Background(), "git", nil, "-C", root, "rev-parse", "HEAD")
		if err != nil || result.Exit != 0 {
			t.Fatalf("rev-parse: %v", err)
		}
		head := strings.TrimSpace(string(result.Stdout))
		expectPass(t, root, head, head)
	})

	type fixture struct {
		dir                      string
		base, freeze, plan, late string
	}
	build := func(t *testing.T) fixture {
		dir := t.TempDir()
		write := func(name, content string) {
			path := filepath.Join(dir, filepath.FromSlash(name))
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		head := func() string {
			result, err := testsupport.RunCLIErr(context.Background(), "git", nil, "-C", dir, "rev-parse", "HEAD")
			if err != nil || result.Exit != 0 {
				t.Fatalf("rev-parse: %v", err)
			}
			return strings.TrimSpace(string(result.Stdout))
		}
		write("cmd/a.go", "package a\n")
		write(".planning/x.md", "plan\n")
		write("scripts/codedigest.sh", string(digestScript))
		write("scripts/check-phase32-freeze.sh", string(freezeScript))
		phase32Git(t, dir, "init", "-q")
		phase32Git(t, dir, "add", "-A")
		phase32Git(t, dir, "commit", "-qm", "base")
		f := fixture{dir: dir, base: head()}
		write("cmd/a.go", "package a // freeze\n")
		phase32Git(t, dir, "commit", "-qam", "freeze")
		f.freeze = head()
		write(".planning/x.md", "plan changed\n")
		phase32Git(t, dir, "commit", "-qam", "plan")
		f.plan = head()
		write("cmd/a.go", "package a // late\n")
		phase32Git(t, dir, "commit", "-qam", "late")
		f.late = head()
		return f
	}
	f := build(t)
	zero := strings.Repeat("0", 40)

	t.Run("planning-only-head-move-accepted", func(t *testing.T) {
		expectPass(t, f.dir, f.freeze, f.plan, f.plan)
		g := build(t)
		phase32Git(t, g.dir, "reset", "-q", "--hard", g.freeze)
		atFreeze := phase32ShDigest(t, g.dir)
		phase32Git(t, g.dir, "reset", "-q", "--hard", g.plan)
		if atPlan := phase32ShDigest(t, g.dir); atPlan != atFreeze {
			t.Fatalf("digest moved across a planning-only commit: %s != %s", atFreeze, atPlan)
		}
		phase32Git(t, g.dir, "reset", "-q", "--hard", g.late)
		if atLate := phase32ShDigest(t, g.dir); atLate == atFreeze {
			t.Fatal("digest did not move across a code commit")
		}
	})
	t.Run("same-commit-accepted", func(t *testing.T) {
		expectPass(t, f.dir, f.freeze, f.freeze)
	})
	t.Run("headsha-mismatch", func(t *testing.T) {
		expectReject(t, f.dir, "sha_mismatch", f.freeze, f.plan, f.late)
		expectReject(t, f.dir, "sha_mismatch", f.freeze, f.plan, zero)
	})
	t.Run("origin-mismatch", func(t *testing.T) {
		expectReject(t, f.dir, "sha_mismatch", f.freeze, f.plan, f.plan, f.freeze)
	})
	t.Run("post-freeze-digest-change", func(t *testing.T) {
		expectReject(t, f.dir, "digest_path_changed_after_freeze", f.freeze, f.late)
	})
	t.Run("md-only-change-after-freeze-accepted", func(t *testing.T) {
		g := build(t)
		phase32Git(t, g.dir, "reset", "-q", "--hard", g.freeze)
		if err := os.WriteFile(filepath.Join(g.dir, "cmd", "README.md"), []byte("prose\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		phase32Git(t, g.dir, "add", "-A")
		phase32Git(t, g.dir, "commit", "-qm", "readme")
		result, err := testsupport.RunCLIErr(context.Background(), "git", nil, "-C", g.dir, "rev-parse", "HEAD")
		if err != nil || result.Exit != 0 {
			t.Fatalf("rev-parse: %v", err)
		}
		readme := strings.TrimSpace(string(result.Stdout))
		expectPass(t, g.dir, g.freeze, readme, readme)
		// Armed counterpart: a .go edit in the same directory still rejects.
		expectReject(t, g.dir, "digest_path_changed_after_freeze", g.freeze, g.late)
	})
	t.Run("not-descendant", func(t *testing.T) {
		expectReject(t, f.dir, "not_descendant", f.freeze, f.base)
	})
	t.Run("bad-sha", func(t *testing.T) {
		expectReject(t, f.dir, "bad_sha", f.freeze, "abc")
		expectReject(t, f.dir, "usage", f.freeze)
	})
	t.Run("unknown-commit", func(t *testing.T) {
		expectReject(t, f.dir, "unknown_commit", f.freeze, strings.Repeat("1", 40))
	})
	t.Run("paths-match-digest", func(t *testing.T) {
		match := regexp.MustCompile(`(?m)^phase32_freeze_paths="([^"]*)"$`).FindSubmatch(freezeScript)
		if match == nil {
			t.Fatal("phase32_freeze_paths line not found")
		}
		if got, want := string(match[1]), strings.Join(phase32DigestPaths, " "); got != want {
			t.Fatalf("freeze paths %q != digest paths %q", got, want)
		}
	})
}
