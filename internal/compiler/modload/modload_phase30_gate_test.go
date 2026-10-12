package modload

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

const gateModuleSource = "module examples.real\n\nexport {\n  fn main\n}\n\nfn main(n: U64) -> U64 {\n  n\n}\n"

func gateWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func gateOpen(t *testing.T, dir string) *resolver {
	t.Helper()
	r, err := newResolver(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { r.close() })
	return r
}

// gateRead reads segments through the production resolver and reports the first
// diagnostic code, "" on success.
func gateRead(r *resolver, segments []string) ([]byte, string) {
	source, diags := r.readSegments(segments, "", diagnostic.Span{})
	if len(diags) != 0 {
		return nil, diags[0].Code
	}
	return source.Bytes, ""
}

// TestPhase30GateRootEscapesAndSymlinks is flag (a) of the D-14 gate.
func TestPhase30GateRootEscapesAndSymlinks(t *testing.T) {
	base := t.TempDir()
	rootDir := filepath.Join(base, "root")
	gateWrite(t, filepath.Join(rootDir, "real.schway"), gateModuleSource)
	gateWrite(t, filepath.Join(rootDir, "sub", "n.schway"), gateModuleSource)
	gateWrite(t, filepath.Join(base, "outside", "secret.schway"), gateModuleSource)
	links := map[string]string{
		"esc.schway":   "../outside/secret.schway",
		"outdir":       "../outside",
		"alias.schway": "real.schway",
		"subl":         "sub",
	}
	for name, target := range links {
		if err := os.Symlink(target, filepath.Join(rootDir, name)); err != nil {
			t.Skipf("env:symlink unavailable on this host: %v", err)
		}
	}
	r := gateOpen(t, rootDir)

	data, code := gateRead(r, []string{"real"})
	if code != "" {
		t.Fatalf("tracer row refused: %s", code)
	}
	if parsed := syntax.Parse(data); len(parsed.Diagnostics) != 0 {
		t.Fatalf("tracer bytes did not parse cleanly: %v", parsed.Diagnostics)
	}

	rows := []struct {
		name     string
		segments []string
		want     string
	}{
		{"escaping file link", []string{"esc"}, CodeSymlink},
		{"escaping dir link", []string{"outdir", "secret"}, CodeSymlink},
		{"in-root file link", []string{"alias"}, CodeSymlink},
		{"in-root dir link", []string{"subl", "n"}, CodeSymlink},
		{"dotdot segment", []string{"..", "outside", "secret"}, CodeEscape},
	}
	for _, row := range rows {
		if got, code := gateRead(r, row.segments); code != row.want || got != nil {
			t.Errorf("%s: code %q bytes %d, want %q and no bytes", row.name, code, len(got), row.want)
		}
	}

	t.Run("control skip per-component Lstat", func(t *testing.T) {
		skipComponentLstatForTest = true
		defer func() { skipComponentLstatForTest = false }()
		if data, code := gateRead(r, []string{"subl", "n"}); code != "" || len(data) == 0 {
			t.Fatalf("control must observe the in-root dir link read through, got code %q", code)
		}
	})
}

// TestPhase30GateCaseExact is flag (b) of the D-14 gate.
func TestPhase30GateCaseExact(t *testing.T) {
	rootDir := t.TempDir()
	gateWrite(t, filepath.Join(rootDir, "Lib.schway"), gateModuleSource)
	gateWrite(t, filepath.Join(rootDir, "Sub", "n.schway"), gateModuleSource)
	r := gateOpen(t, rootDir)

	if _, code := gateRead(r, []string{"lib"}); code != CodeCaseMismatch {
		t.Fatalf("lib with only Lib.schway: code %q, want %q", code, CodeCaseMismatch)
	}
	if _, code := gateRead(r, []string{"sub", "n"}); code != CodeCaseMismatch {
		t.Fatalf("sub.n with only Sub/: code %q, want %q", code, CodeCaseMismatch)
	}
	if _, code := gateRead(r, []string{"Lib"}); code != "" {
		t.Fatalf("exact Lib: code %q", code)
	}

	t.Run("control case-fold accept", func(t *testing.T) {
		caseFoldAcceptForTest = true
		defer func() { caseFoldAcceptForTest = false }()
		if _, code := gateRead(r, []string{"lib"}); code == CodeCaseMismatch {
			t.Fatalf("control must differ from %q", CodeCaseMismatch)
		}
	})
}
