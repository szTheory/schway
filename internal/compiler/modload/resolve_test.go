package modload

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

var resolveAt = diagnostic.Span{Start: 7, End: 15}

const resolveImporter = "main"

// resolveCode reads module through the production resolver and returns the
// first diagnostic code, "" on success.
func resolveCode(t *testing.T, r *resolver, module string) (Source, string) {
	t.Helper()
	source, diags := r.read(module, resolveImporter, resolveAt)
	if len(diags) == 0 {
		return source, ""
	}
	if len(diags) != 1 {
		t.Fatalf("%s: %d diagnostics, want exactly one", module, len(diags))
	}
	d := diags[0]
	if d.Primary != resolveAt {
		t.Errorf("%s: primary span %v, want %v", module, d.Primary, resolveAt)
	}
	return source, d.Code
}

// TestPhase30SymlinkAndEscapeRefused covers D-08: every symlink, escaping or
// not, file or directory, is refused by name.
func TestPhase30SymlinkAndEscapeRefused(t *testing.T) {
	base := t.TempDir()
	rootDir := filepath.Join(base, "root")
	gateWrite(t, filepath.Join(rootDir, "real.schway"), gateModuleSource)
	gateWrite(t, filepath.Join(rootDir, "sub", "n.schway"), gateModuleSource)
	secret := filepath.Join(base, "outside", "secret.schway")
	gateWrite(t, secret, gateModuleSource)
	if err := os.MkdirAll(filepath.Join(rootDir, "deep"), 0o755); err != nil {
		t.Fatal(err)
	}
	links := map[string]string{
		"esc.schway":                 "../outside/secret.schway",
		"outdir":                     "../outside",
		"alias.schway":               "real.schway",
		"subl":                       "sub",
		"abs.schway":                 secret,
		filepath.Join("deep", "lnk"): filepath.Join("..", "sub"),
	}
	for name, target := range links {
		if err := os.Symlink(target, filepath.Join(rootDir, name)); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
	}
	r := gateOpen(t, rootDir)

	source, code := resolveCode(t, r, "real")
	if code != "" {
		t.Fatalf("tracer row refused: %s", code)
	}
	if source.Module != "real" || source.RelPath != "real.schway" {
		t.Fatalf("source identity %q %q", source.Module, source.RelPath)
	}
	if parsed := syntax.Parse(source.Bytes); len(parsed.Diagnostics) != 0 {
		t.Fatalf("tracer bytes did not parse cleanly: %v", parsed.Diagnostics)
	}

	rows := []struct {
		name, module, want string
	}{
		{"escaping file link", "esc", CodeSymlink},
		{"escaping dir link", "outdir.secret", CodeSymlink},
		{"in-root file link", "alias", CodeSymlink},
		{"in-root dir link", "subl.n", CodeSymlink},
		{"nested dir link at depth 2", "deep.lnk.n", CodeSymlink},
		{"absolute-target file link", "abs", CodeSymlink},
	}
	for _, row := range rows {
		got, code := resolveCode(t, r, row.module)
		if code != row.want || got.Bytes != nil {
			t.Errorf("%s: code %q bytes %d, want %q and no bytes", row.name, code, len(got.Bytes), row.want)
		}
	}

	// The segment grammar stops a dotdot before the filesystem; the root layer
	// is the second defence for a segment that bypasses the grammar.
	if _, diags := r.readSegments([]string{"..", "outside", "secret"}, resolveImporter, resolveAt); len(diags) != 1 || diags[0].Code != CodeEscape {
		t.Errorf("dotdot segment: %v, want %s", diags, CodeEscape)
	}

	t.Run("control skip per-component Lstat", func(t *testing.T) {
		skipComponentLstatForTest = true
		defer func() { skipComponentLstatForTest = false }()
		if source, code := resolveCode(t, r, "subl.n"); code != "" || len(source.Bytes) == 0 {
			t.Fatalf("control must observe the in-root dir link read through, got code %q", code)
		}
	})
	t.Run("control bypass root", func(t *testing.T) {
		bypassRootForTest = true
		defer func() { bypassRootForTest = false }()
		if source, code := resolveCode(t, r, "esc"); code != "" || len(source.Bytes) == 0 {
			t.Fatalf("control must observe the escaping link followed, got code %q", code)
		}
	})
}

// TestPhase30CaseExactMatch covers D-09: exact case through a root listing.
func TestPhase30CaseExactMatch(t *testing.T) {
	rootDir := t.TempDir()
	gateWrite(t, filepath.Join(rootDir, "Lib.schway"), gateModuleSource)
	r := gateOpen(t, rootDir)
	if _, code := resolveCode(t, r, "lib"); code != CodeCaseMismatch {
		t.Fatalf("lib with only Lib.schway: code %q, want %q", code, CodeCaseMismatch)
	}
	if _, code := resolveCode(t, r, "missing"); code != CodeNotFound {
		t.Fatalf("missing: code %q, want %q", code, CodeNotFound)
	}
	t.Run("control case-fold accept", func(t *testing.T) {
		caseFoldAcceptForTest = true
		defer func() { caseFoldAcceptForTest = false }()
		if _, code := resolveCode(t, r, "lib"); code == CodeCaseMismatch {
			t.Fatalf("control must differ from %q", CodeCaseMismatch)
		}
	})

	t.Run("case-sensitive filesystem loads the lowercase file", func(t *testing.T) {
		dir := t.TempDir()
		gateWrite(t, filepath.Join(dir, "Lib.schway"), "upper")
		lower := filepath.Join(dir, "lib.schway")
		if _, err := os.Stat(lower); err == nil {
			t.Skip("case-insensitive filesystem")
		}
		if err := os.WriteFile(lower, []byte("lower"), 0o644); err != nil {
			t.Fatal(err)
		}
		if entries, _ := os.ReadDir(dir); len(entries) != 2 {
			t.Skip("case-insensitive filesystem")
		}
		source, code := resolveCode(t, gateOpen(t, dir), "lib")
		if code != "" || string(source.Bytes) != "lower" {
			t.Fatalf("lib: code %q bytes %q, want the lowercase file", code, source.Bytes)
		}
	})
}
