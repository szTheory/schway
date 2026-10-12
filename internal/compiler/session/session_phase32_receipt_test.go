package session_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// phase32DigestPaths is the receipt contract (D-08a): the paths whose HEAD
// listing is hashed. scripts/codedigest.sh names the same set.
var phase32DigestPaths = []string{"cmd", "internal", "examples", "native", "scripts", "testdata", "go.mod", ".github/workflows/ci.yml"}

// phase32CodeDigest is the Go twin of scripts/codedigest.sh.
func phase32CodeDigest(t testing.TB, root string) string {
	t.Helper()
	arguments := append([]string{"-C", root, "-c", "core.quotePath=false", "ls-tree", "-r", "HEAD", "--"}, phase32DigestPaths...)
	result, err := testsupport.RunCLIErr(context.Background(), "git", nil, arguments...)
	if err != nil || result.Exit != 0 {
		t.Fatalf("git ls-tree in %s: exit=%d err=%v stderr=%s", root, result.Exit, err, result.Stderr)
	}
	var kept []string
	for _, line := range strings.Split(strings.TrimRight(string(result.Stdout), "\n"), "\n") {
		if line == "" || strings.HasSuffix(line, ".md") {
			continue
		}
		kept = append(kept, line)
	}
	sum := sha256.Sum256([]byte(strings.Join(kept, "\n") + "\n"))
	return hex.EncodeToString(sum[:])
}

func phase32ShDigest(t testing.TB, root string) string {
	t.Helper()
	result, err := testsupport.RunCLIErr(context.Background(), "sh", nil, "-c", `cd "$0" && sh scripts/codedigest.sh`, root)
	if err != nil || result.Exit != 0 {
		t.Fatalf("codedigest.sh in %s: exit=%d err=%v stderr=%s", root, result.Exit, err, result.Stderr)
	}
	return strings.TrimSpace(string(result.Stdout))
}

func phase32Git(t testing.TB, dir string, arguments ...string) {
	t.Helper()
	full := append([]string{"-C", dir, "-c", "user.name=phase32", "-c", "user.email=phase32@example.invalid"}, arguments...)
	result, err := testsupport.RunCLIErr(context.Background(), "git", nil, full...)
	if err != nil || result.Exit != 0 {
		t.Fatalf("git %v: exit=%d err=%v stderr=%s", arguments, result.Exit, err, result.Stderr)
	}
}

func TestPhase32CodeDigest(t *testing.T) {
	root := testsupport.ProjectPath()
	t.Run("sh-equals-go", func(t *testing.T) {
		sh, goDigest := phase32ShDigest(t, root), phase32CodeDigest(t, root)
		if sh != goDigest {
			t.Fatalf("sh digest %s != go digest %s", sh, goDigest)
		}
	})

	script, err := os.ReadFile(testsupport.ProjectPath("scripts", "codedigest.sh"))
	if err != nil {
		t.Fatal(err)
	}
	tempRepo := func(t *testing.T) string {
		dir := t.TempDir()
		for name, content := range map[string][]byte{
			"cmd/a.go":              []byte("package a\n"),
			"cmd/README.md":         []byte("readme\n"),
			".planning/x.md":        []byte("plan\n"),
			"scripts/codedigest.sh": script,
		} {
			path := filepath.Join(dir, filepath.FromSlash(name))
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, content, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		phase32Git(t, dir, "init", "-q")
		phase32Git(t, dir, "add", "-A")
		phase32Git(t, dir, "commit", "-qm", "base")
		return dir
	}
	for _, row := range []struct {
		name, file string
		moves      bool
	}{
		{"armed-code-edit", "cmd/a.go", true},
		{"md-edit-ignored", "cmd/README.md", false},
		{"planning-edit-ignored", ".planning/x.md", false},
	} {
		t.Run(row.name, func(t *testing.T) {
			dir := tempRepo(t)
			before := phase32CodeDigest(t, dir)
			if sh := phase32ShDigest(t, dir); sh != before {
				t.Fatalf("before edit: sh %s != go %s", sh, before)
			}
			path := filepath.Join(dir, filepath.FromSlash(row.file))
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, append(content, '!'), 0o644); err != nil {
				t.Fatal(err)
			}
			phase32Git(t, dir, "add", "-A")
			phase32Git(t, dir, "commit", "-qm", "edit")
			after := phase32CodeDigest(t, dir)
			if sh := phase32ShDigest(t, dir); sh != after {
				t.Fatalf("after edit: sh %s != go %s", sh, after)
			}
			if (before != after) != row.moves {
				t.Fatalf("digest moved=%v, want %v", before != after, row.moves)
			}
		})
	}
}
