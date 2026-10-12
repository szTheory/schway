package modload

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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
	source, _, diags := r.read(module, resolveImporter, resolveAt, nil)
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

// The per-module cap boundary, spelled as the decimal literals D-13 names.
const (
	maxModuleBytes  = 1048576
	overModuleBytes = 1048577
)

// padModule returns a module-shaped file of exactly size bytes.
func padModule(size int) []byte {
	const head = "module m\n\n// "
	out := make([]byte, size)
	copy(out, head)
	for index := len(head); index < size-1; index++ {
		out[index] = 'x'
	}
	out[size-1] = '\n'
	return out
}

// TestPhase30FakeFSToctouAndStatSize covers D-10: identity must survive Open,
// and the size never comes from Stat.
func TestPhase30FakeFSToctouAndStatSize(t *testing.T) {
	t.Run("identity swap between Lstat and Open", func(t *testing.T) {
		fsys := newFakeFileSystem()
		entry := fsys.addFile("lib.schway", gateModuleSource, 7)
		entry.identityAfterOpen = 8
		r := newFakeResolver(fsys)
		if source, code := resolveCode(t, r, "lib"); code != CodeUnstableFile || source.Bytes != nil {
			t.Fatalf("swap: code %q bytes %d, want %q and no bytes", code, len(source.Bytes), CodeUnstableFile)
		}
		skipSameFileForTest = true
		defer func() { skipSameFileForTest = false }()
		if _, code := resolveCode(t, r, "lib"); code != "" {
			t.Fatalf("control must observe the swapped file accepted, got %q", code)
		}
	})

	t.Run("Stat lies about size", func(t *testing.T) {
		fsys := newFakeFileSystem()
		entry := fsys.addBytes("lib.schway", padModule(overModuleBytes), 7)
		entry.statSize = 10
		r := newFakeResolver(fsys)
		diags := func() []diagnostic.Diagnostic {
			_, _, d := r.read("lib", resolveImporter, resolveAt, nil)
			return d
		}()
		if len(diags) != 1 || diags[0].Code != "syntax.input_limit" || diags[0].Unit != "lib" {
			t.Fatalf("oversized behind a small Stat: %+v, want syntax.input_limit in unit lib", diags)
		}
		sizeFromStatForTest = true
		defer func() { sizeFromStatForTest = false }()
		if _, code := resolveCode(t, r, "lib"); code != "" {
			t.Fatalf("control must observe the oversized module accepted, got %q", code)
		}
	})
}

// TestPhase30PerModuleByteCap covers D-13 on a real filesystem.
func TestPhase30PerModuleByteCap(t *testing.T) {
	rootDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(rootDir, "exact.schway"), padModule(maxModuleBytes), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rootDir, "over.schway"), padModule(overModuleBytes), 0o644); err != nil {
		t.Fatal(err)
	}
	r := gateOpen(t, rootDir)

	source, code := resolveCode(t, r, "exact")
	if code != "" || len(source.Bytes) != maxModuleBytes || syntax.MaxSourceBytes != maxModuleBytes {
		t.Fatalf("exact cap: code %q len %d", code, len(source.Bytes))
	}
	_, _, diags := r.read("over", resolveImporter, resolveAt, nil)
	if len(diags) != 1 || diags[0].Code != "syntax.input_limit" || diags[0].Unit != "over" {
		t.Fatalf("cap+1: %+v, want syntax.input_limit in unit over", diags)
	}
	if strings.ContainsAny(diags[0].Message, "0123456789") {
		t.Errorf("message carries a size number: %q", diags[0].Message)
	}

	t.Run("control off-by-one", func(t *testing.T) {
		byteCapOffByOneForTest = true
		defer func() { byteCapOffByOneForTest = false }()
		if _, code := resolveCode(t, r, "exact"); code != "syntax.input_limit" {
			t.Fatalf("control must observe the exact-cap module refused, got %q", code)
		}
	})
}

// TestPhase30ResolverOpensOnce covers D-10: one open per read, none for any
// pre-open refusal.
func TestPhase30ResolverOpensOnce(t *testing.T) {
	fsys := newFakeFileSystem()
	fsys.addFile("lib.schway", gateModuleSource, 2)
	fsys.addFile("Cased.schway", gateModuleSource, 3)
	fsys.entries["alias.schway"] = &fakeEntry{mode: fs.ModeSymlink | 0o777, id: 4}
	fsys.entries["pipe.schway"] = &fakeEntry{mode: fs.ModeNamedPipe | 0o644, id: 5}
	r := newFakeResolver(fsys)

	if _, code := resolveCode(t, r, "lib"); code != "" || fsys.opens != 1 {
		t.Fatalf("success: code %q opens %d, want opens 1", code, fsys.opens)
	}
	for _, row := range []struct{ module, want string }{
		{"alias", CodeSymlink},
		{"pipe", CodeNotRegular},
		{"cased", CodeCaseMismatch},
		{"gone", CodeNotFound},
		{"Bad", CodeInvalidSegment},
	} {
		before := fsys.opens
		if _, code := resolveCode(t, r, row.module); code != row.want || fsys.opens != before {
			t.Errorf("%s: code %q opens +%d, want %q and no open", row.module, code, fsys.opens-before, row.want)
		}
	}
}

// TestPhase30DirectoryListingBounded covers D-24.
func TestPhase30DirectoryListingBounded(t *testing.T) {
	build := func(total int) *resolver {
		fsys := newFakeFileSystem()
		fsys.addFile("lib.schway", gateModuleSource, 2)
		fsys.entries["."].extraNames = total - 1
		return newFakeResolver(fsys)
	}
	if source, code := resolveCode(t, build(MaxDirectoryEntries), "lib"); code != "" || len(source.Bytes) == 0 {
		t.Fatalf("%d entries: code %q, want success", MaxDirectoryEntries, code)
	}
	if _, code := resolveCode(t, build(MaxDirectoryEntries+1), "lib"); code != CodeDirectoryLimit {
		t.Fatalf("%d entries: code %q, want %q", MaxDirectoryEntries+1, code, CodeDirectoryLimit)
	}

	t.Run("real listing is bounded", func(t *testing.T) {
		dir := t.TempDir()
		for _, name := range []string{"a", "b", "c", "d"} {
			gateWrite(t, filepath.Join(dir, name), "")
		}
		fsys := gateOpen(t, dir).fsys
		if names, truncated, err := fsys.ListDir(".", 4); err != nil || truncated || len(names) != 4 {
			t.Fatalf("max 4: %v truncated %v err %v", names, truncated, err)
		}
		if names, truncated, err := fsys.ListDir(".", 3); err != nil || !truncated || len(names) != 3 {
			t.Fatalf("max 3: %v truncated %v err %v", names, truncated, err)
		}
	})
}

// TestPhase30SegmentGrammar covers D-07 and D-18: the grammar is enforced
// before any filesystem access.
func TestPhase30SegmentGrammar(t *testing.T) {
	accepted := []string{"a", "adler32", "a_b", "x1", "a.b.c", strings.Repeat("a", 64), "a." + strings.Repeat("b", 64)}
	refused := []string{"A", "Lib", "a__b", "a_", "_a", "1a", "café", strings.Repeat("a", 65), "a..b", "", ".a", "a.", "a-b", "a b", "a/b", "a\x00b"}
	for _, module := range accepted {
		if !validateModulePath(module) {
			t.Errorf("%q refused, want accepted", module)
		}
	}
	for _, module := range refused {
		if validateModulePath(module) {
			t.Errorf("%q accepted, want refused", module)
		}
	}

	fsys := newFakeFileSystem()
	fsys.addFile("lib.schway", gateModuleSource, 2)
	r := newFakeResolver(fsys)
	for _, module := range refused {
		if source, code := resolveCode(t, r, module); code != CodeInvalidSegment || source.Bytes != nil {
			t.Errorf("%q: code %q, want %q", module, code, CodeInvalidSegment)
		}
	}
	if fsys.opens != 0 {
		t.Errorf("refused modules caused %d opens, want 0", fsys.opens)
	}
	if _, code := resolveCode(t, r, "lib"); code != "" {
		t.Errorf("control: lib refused with %q", code)
	}
}

// TestPhase30ErrorsNeverLeakAbsolutePaths covers D-10's information boundary:
// diagnostics carry module names and root-relative paths only.
func TestPhase30ErrorsNeverLeakAbsolutePaths(t *testing.T) {
	base := t.TempDir()
	rootDir := filepath.Join(base, "root")
	gateWrite(t, filepath.Join(rootDir, "real.schway"), gateModuleSource)
	gateWrite(t, filepath.Join(rootDir, "Cased.schway"), gateModuleSource)
	gateWrite(t, filepath.Join(base, "outside", "secret.schway"), gateModuleSource)
	if err := os.WriteFile(filepath.Join(rootDir, "over.schway"), padModule(overModuleBytes), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join("..", "outside", "secret.schway"), filepath.Join(rootDir, "esc.schway")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	modules := []string{"esc", "cased", "gone", "over", "Bad"}
	if pipe := makeFIFO(t, rootDir); pipe {
		modules = append(modules, "pipe")
	}
	r := gateOpen(t, rootDir)

	forms := []string{base}
	if resolved, err := filepath.EvalSymlinks(base); err == nil {
		forms = append(forms, resolved)
	}
	check := func(label string, diags []diagnostic.Diagnostic, module string) {
		t.Helper()
		if len(diags) != 1 {
			t.Fatalf("%s: %d diagnostics, want 1", label, len(diags))
		}
		encoded, err := json.Marshal(diags[0])
		if err != nil {
			t.Fatal(err)
		}
		for _, form := range forms {
			if strings.Contains(string(encoded), form) {
				t.Errorf("%s: diagnostic leaks %q: %s", label, form, encoded)
			}
		}
		if diags[0].Code != "syntax.input_limit" && !strings.Contains(diags[0].Message, module) {
			t.Errorf("%s: message %q does not name the module", label, diags[0].Message)
		}
	}
	for _, module := range modules {
		_, _, diags := r.read(module, resolveImporter, resolveAt, nil)
		check(module, diags, module)
	}

	// The escape and read_failed paths render an operating-system error; a fake
	// filesystem whose errors carry the temp path proves production strips it.
	leaky := &leakyFileSystem{dir: base}
	lr := &resolver{fsys: leaky, sameFile: fakeSameFile}
	_, _, diags := lr.read("lib", resolveImporter, resolveAt, nil)
	check("leaky production", diags, "lib")
	if diags[0].Code != CodeEscape {
		t.Fatalf("leaky Lstat: code %q, want %q", diags[0].Code, CodeEscape)
	}

	t.Run("control format error text", func(t *testing.T) {
		formatErrorTextForTest = true
		defer func() { formatErrorTextForTest = false }()
		_, _, diags := lr.read("lib", resolveImporter, resolveAt, nil)
		encoded, _ := json.Marshal(diags)
		if !strings.Contains(string(encoded), base) {
			t.Fatalf("control must observe the temp path leak, got %s", encoded)
		}
	})
}

// leakyFileSystem fails Lstat with an error whose text names an absolute path.
type leakyFileSystem struct{ dir string }

func (l *leakyFileSystem) Lstat(name string) (fs.FileInfo, error) {
	return nil, &fs.PathError{Op: "lstat", Path: filepath.Join(l.dir, name), Err: errors.New("denied")}
}
func (l *leakyFileSystem) Open(string) (moduleFile, error) { return nil, errors.New("unused") }
func (l *leakyFileSystem) ListDir(string, int) ([]string, bool, error) {
	return nil, false, errors.New("unused")
}
