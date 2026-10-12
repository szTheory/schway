package session_test

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// phase31CopyAllowList is the frozen set of non-library files that may carry
// the Adler-32 loop (D-04). Growing it requires changing a sha256 constant in
// a reviewed commit.
var phase31CopyAllowList = map[string]string{
	"examples/phase29/adler32_fixture.schway": "96ae56eb593001ab77b3525396cb03dd013b35cc948516851d63758b3ff1a1bf",
	"testdata/phase29/adler32_library.schway": "ec1f3a40c576cc19d47563b5e5e76ca2213de1ad35a578845da9169c94237a36",
}

const phase31CopyLibrary = "examples/checksum/adler32.schway"

func phase31CopyDir(t *testing.T, src string) string {
	t.Helper()
	dst := t.TempDir()
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(src, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dst, entry.Name()), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dst
}

// phase31CopyScan returns the sorted root-relative .schway files under the
// given roots that contain both `fn adler32(` and `% 65521`.
func phase31CopyScan(t *testing.T, base string, roots ...string) []string {
	t.Helper()
	var found []string
	for _, root := range roots {
		err := filepath.WalkDir(filepath.Join(base, root), func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || filepath.Ext(path) != ".schway" {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			text := string(data)
			if strings.Contains(text, "fn adler32(") && strings.Contains(text, "% 65521") {
				rel, err := filepath.Rel(base, path)
				if err != nil {
					return err
				}
				found = append(found, filepath.ToSlash(rel))
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	sort.Strings(found)
	return found
}

func phase31CopySha(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// phase31CopyReplaceOnce replaces exactly one occurrence of old in the file.
func phase31CopyReplaceOnce(t *testing.T, path, old, replacement string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(data), old) != 1 {
		t.Fatalf("%s does not contain exactly one %q", path, old)
	}
	if err := os.WriteFile(path, []byte(strings.Replace(string(data), old, replacement, 1)), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestPhase31SingleCopyAndDualFlip(t *testing.T) {
	base := testsupport.ProjectPath()
	checksumDir := testsupport.ProjectPath("examples", "checksum")

	t.Run("static-scan", func(t *testing.T) {
		got := phase31CopyScan(t, base, "examples", "testdata")
		want := []string{phase31CopyLibrary, "examples/phase29/adler32_fixture.schway", "testdata/phase29/adler32_library.schway"}
		if strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Fatalf("files carrying the Adler-32 loop = %v, want %v", got, want)
		}
	})

	t.Run("allow-list-hash", func(t *testing.T) {
		for path, want := range phase31CopyAllowList {
			data, err := os.ReadFile(filepath.Join(base, path))
			if err != nil {
				t.Fatal(err)
			}
			if got := phase31CopySha(data); got != want {
				t.Fatalf("%s sha256 = %s, want %s", path, got, want)
			}
			flipped := append([]byte(nil), data...)
			flipped[0] ^= 1
			if phase31CopySha(flipped) == want {
				t.Fatalf("%s: a one-byte flip did not change the hash", path)
			}
		}
	})

	t.Run("per-app-single-adler32", func(t *testing.T) {
		for _, app := range []string{"checksum.schway", "adler_low.schway", "adler32_cli.schway"} {
			checked, err := session.CheckFile(filepath.Join(checksumDir, app))
			if err != nil || len(checked.Diagnostics) != 0 {
				t.Fatalf("%s: err=%v diagnostics=%+v", app, err, checked.Diagnostics)
			}
			count := 0
			for _, function := range checked.Program.Functions {
				if function.Name == "adler32" {
					count++
					if function.Module != "adler32" {
						t.Fatalf("%s: adler32 Module = %q, want adler32", app, function.Module)
					}
				}
			}
			if count != 1 {
				t.Fatalf("%s: merged program has %d adler32 functions, want 1", app, count)
			}
		}
	})

	t.Run("copied-library-is-detected", func(t *testing.T) {
		clone := phase31CopyDir(t, checksumDir)
		library, err := os.ReadFile(filepath.Join(clone, "adler32.schway"))
		if err != nil {
			t.Fatal(err)
		}
		text := string(library)
		body := text[strings.Index(text, "fn adler32("):]
		entryPath := filepath.Join(clone, "adler_low.schway")
		entry, err := os.ReadFile(entryPath)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(entryPath, append(entry, []byte("\n"+body)...), 0o600); err != nil {
			t.Fatal(err)
		}
		if got := phase31CopyScan(t, clone, "."); len(got) != 2 {
			t.Fatalf("clone with a copied library has loops in %v, want two files", got)
		}
		checked, err := session.CheckFile(entryPath)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, d := range checked.Diagnostics {
			if d.Code == "module.import_collision" {
				found = true
			}
		}
		if !found {
			t.Fatalf("diagnostics = %+v, want module.import_collision", checked.Diagnostics)
		}
	})

	t.Run("adler-low-interpreter", func(t *testing.T) {
		checked, err := session.CheckFile(filepath.Join(checksumDir, "adler_low.schway"))
		if err != nil || len(checked.Diagnostics) != 0 {
			t.Fatalf("err=%v diagnostics=%+v", err, checked.Diagnostics)
		}
		for _, row := range testsupport.ChecksumOracleRows() {
			if row.App != "adler_low" {
				continue
			}
			path := filepath.Join(t.TempDir(), "in.bin")
			if err := os.WriteFile(path, row.Input, 0o600); err != nil {
				t.Fatal(err)
			}
			run, err := interp.RunWithFileBytes(checked.Program, "main", path, testsupport.ChecksumReadFile)
			if err != nil {
				t.Fatal(err)
			}
			if row.Outcome == "FileTooLarge" {
				if run.Outcome.Kind != execution.OutcomeTypedFailure || run.Outcome.Value != "FileTooLarge" {
					t.Fatalf("%s outcome = %+v, want typed_failure FileTooLarge", row.Name, run.Outcome)
				}
				continue
			}
			if run.Outcome.Kind != execution.OutcomeReturned || run.Outcome.Value != strconv.FormatUint(row.Want, 10) {
				t.Fatalf("%s outcome = %+v, want returned %d", row.Name, run.Outcome, row.Want)
			}
		}
	})
	//GSD-TASK3
}
