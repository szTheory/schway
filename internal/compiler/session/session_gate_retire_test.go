package session_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func phase33GateCallers(files map[string][]byte) []string {
	needle := "check-" + "phase32"
	var callers []string
	for path, content := range files {
		if strings.Contains(string(content), needle) {
			callers = append(callers, path)
		}
	}
	return callers
}

func TestNoCheckPhase32Caller(t *testing.T) {
	root := testsupport.ProjectPath()
	result, err := testsupport.RunCLIErr(context.Background(), "git", nil, "-C", root, "ls-files", "-z")
	if err != nil || result.Exit != 0 {
		t.Fatalf("git ls-files: exit=%d err=%v stderr=%s", result.Exit, err, result.Stderr)
	}
	files := map[string][]byte{}
	for _, name := range strings.Split(string(result.Stdout), "\x00") {
		if name == "" || strings.HasPrefix(name, ".planning/") {
			continue
		}
		path := filepath.Join(root, filepath.FromSlash(name))
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read tracked file %s: %v", name, err)
		}
		files[name] = content
	}
	if callers := phase33GateCallers(files); len(callers) > 0 {
		t.Fatalf("tracked files call retired Phase 32 gate scripts: %v", callers)
	}
	freeze := "scripts/check-" + "phase32" + "-freeze.sh"
	receipts := "scripts/check-" + "phase32" + "-receipts.sh"
	for _, name := range []string{freeze, receipts} {
		if _, err := os.Stat(filepath.Join(root, name)); err == nil {
			t.Errorf("retired gate script still exists: %s", name)
		}
	}
	t.Run("armed-control", func(t *testing.T) {
		control := map[string][]byte{"fixture/caller.sh": []byte("sh " + "check-" + "phase32" + "-freeze.sh")}
		if got := phase33GateCallers(control); len(got) != 1 || got[0] != "fixture/caller.sh" {
			t.Fatalf("seeded caller scan = %v, want fixture/caller.sh", got)
		}
	})
}
