package session_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

type phase33CommandResult struct {
	exit           int
	stdout, stderr string
}

func phase33RunAt(t *testing.T, cwd, script string, args ...string) phase33CommandResult {
	t.Helper()
	full := append([]string{"-c", `cd "$0" && script=$1 && shift && exec sh "$script" "$@"`, cwd, script}, args...)
	result, err := testsupport.RunCLIErr(context.Background(), "sh", nil, full...)
	if err != nil && result.Exit == 0 {
		t.Fatalf("spawn %s: %v", script, err)
	}
	return phase33CommandResult{result.Exit, string(result.Stdout), string(result.Stderr)}
}

func TestFreezeGateDifferential(t *testing.T) {
	read := func(path string) []byte {
		data, err := os.ReadFile(testsupport.ProjectPath(filepath.FromSlash(path)))
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	dir := t.TempDir()
	write := func(name string, data []byte) {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("cmd/a.go", []byte("package a\n"))
	write("scripts/codedigest.sh", read("scripts/codedigest.sh"))
	write("scripts/check-phase32-freeze.sh", read("scripts/check-phase32-freeze.sh"))
	write("scripts/check-freeze.sh", read("scripts/check-freeze.sh"))
	write("testdata/receipts/M006.manifest", read("testdata/receipts/M006.manifest"))
	phase32Git(t, dir, "init", "-q")
	phase32Git(t, dir, "add", "-A")
	phase32Git(t, dir, "commit", "-qm", "base")
	base := phase32GateHead(t, dir)
	write("cmd/a.go", []byte("package a // freeze\n"))
	phase32Git(t, dir, "commit", "-qam", "freeze")
	freeze := phase32GateHead(t, dir)
	write(".planning/x.md", []byte("plan\n"))
	phase32Git(t, dir, "add", "-A")
	phase32Git(t, dir, "commit", "-qm", "planning")
	planning := phase32GateHead(t, dir)
	write("cmd/README.md", []byte("docs\n"))
	phase32Git(t, dir, "add", "-A")
	phase32Git(t, dir, "commit", "-qm", "markdown")
	markdown := phase32GateHead(t, dir)
	write("cmd/a.go", []byte("package a // changed\n"))
	phase32Git(t, dir, "commit", "-qam", "code")
	changed := phase32GateHead(t, dir)
	phase32Git(t, dir, "checkout", "-q", "--detach", base)
	write("other.txt", []byte("sibling\n"))
	phase32Git(t, dir, "add", "-A")
	phase32Git(t, dir, "commit", "-qm", "sibling")
	sibling := phase32GateHead(t, dir)
	old := filepath.Join(dir, "scripts/check-phase32-freeze.sh")
	new := filepath.Join(dir, "scripts/check-freeze.sh")
	manifest := "testdata/receipts/M006.manifest"
	rows := []struct {
		name, reason     string
		oldArgs, newArgs []string
	}{
		{"usage", "reason=usage", nil, []string{manifest}},
		{"bad_sha", "reason=bad_sha", []string{freeze, "BAD"}, []string{manifest, freeze, "BAD"}},
		{"bad_same_as", "reason=bad_sha", []string{freeze, planning, "BAD"}, []string{manifest, freeze, planning, "BAD"}},
		{"sha_mismatch", "reason=sha_mismatch", []string{freeze, planning, freeze}, []string{manifest, freeze, planning, freeze}},
		{"unknown_commit", "reason=unknown_commit", []string{freeze, strings.Repeat("0", 40)}, []string{manifest, freeze, strings.Repeat("0", 40)}},
		{"not_descendant", "reason=not_descendant", []string{freeze, sibling}, []string{manifest, freeze, sibling}},
		{"digest_change", "reason=digest_path_changed_after_freeze", []string{freeze, changed}, []string{manifest, freeze, changed}},
		{"md_only_pass", "status=pass", []string{freeze, markdown, markdown}, []string{manifest, freeze, markdown, markdown}},
		{"planning_only_pass", "status=pass", []string{freeze, planning, planning}, []string{manifest, freeze, planning, planning}},
		{"same_commit_pass", "status=pass", []string{freeze, freeze}, []string{manifest, freeze, freeze}},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			oldResult := phase33RunAt(t, dir, old, row.oldArgs...)
			newResult := phase33RunAt(t, dir, new, row.newArgs...)
			if oldResult != newResult {
				t.Fatalf("old/new differ\nold=%#v\nnew=%#v", oldResult, newResult)
			}
			actual := oldResult.stderr + oldResult.stdout
			if !strings.Contains(actual, row.reason) {
				t.Fatalf("expected %q in %q", row.reason, actual)
			}
		})
	}
}
