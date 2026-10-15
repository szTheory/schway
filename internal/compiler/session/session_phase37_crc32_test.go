package session_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase37CRC32NativeTracer(t *testing.T) {
	exampleRoot := testsupport.ProjectPath("examples", "crc32")
	entryPath := filepath.Join(exampleRoot, "crc32.schway")
	libraryPath := filepath.Join(exampleRoot, "lib", "crc32.schway")
	entrySource, err := os.ReadFile(entryPath)
	if err != nil {
		t.Fatal(err)
	}
	librarySource, err := os.ReadFile(libraryPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(entrySource), "import lib.crc32 {") != 1 {
		t.Fatal("CRC entry must import lib.crc32 exactly once")
	}
	if !strings.Contains(string(librarySource), "export {\n  fn crc32\n}") || !strings.Contains(string(librarySource), "fn crc_byte(") {
		t.Fatal("CRC library must export crc32 and keep its crc_byte helper private")
	}

	checked, err := session.CheckFile(entryPath)
	if err != nil || len(checked.Diagnostics) != 0 {
		t.Fatalf("check CRC entry: err=%v diagnostics=%+v", err, checked.Diagnostics)
	}
	crc32Functions := 0
	for _, function := range checked.Program.Functions {
		if function.Name == "crc32" {
			crc32Functions++
			if function.Module != "lib.crc32" {
				t.Fatalf("crc32 module = %q, want lib.crc32", function.Module)
			}
		}
	}
	if crc32Functions != 1 {
		t.Fatalf("merged program has %d crc32 definitions, want one imported implementation", crc32Functions)
	}

	artifact := filepath.Join(t.TempDir(), "crc32")
	manifest := testsupport.ProjectPath("examples", "checksum", "file_bytes.bindings.json")
	_, diagnostics, err := session.BuildApplicationFile(context.Background(), entryPath, artifact, native.DefaultRunner(), manifest)
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("build CRC app: err=%v diagnostics=%+v", err, diagnostics)
	}
	input := filepath.Join(t.TempDir(), "input.bin")
	if err := os.WriteFile(input, []byte("123456789"), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	outcome, err := native.RunApplication(context.Background(), artifact, input, &stdout, &stderr)
	if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 0 {
		t.Fatalf("native app outcome=%+v err=%v stdout=%q stderr=%q", outcome, err, stdout.String(), stderr.String())
	}
	if stdout.String() != "3421780262\n" || stderr.String() != "" {
		t.Fatalf("native CRC result stdout=%q stderr=%q; want %q and empty stderr", stdout.String(), stderr.String(), "3421780262\n")
	}
}

func TestPhase37CRC32ByteHelperCalls(t *testing.T) {
	entryPath := testsupport.ProjectPath("examples", "crc32", "crc32.schway")
	checked, err := session.CheckFile(entryPath)
	if err != nil || len(checked.Diagnostics) != 0 {
		t.Fatalf("check CRC entry for helper activation: err=%v diagnostics=%+v", err, checked.Diagnostics)
	}
	helperID := ""
	for _, function := range checked.Program.Functions {
		if function.Name != "crc_byte" || function.Module != "lib.crc32" {
			continue
		}
		if helperID != "" {
			t.Fatal("checked CRC closure contains more than one private crc_byte helper")
		}
		helperID = function.ID
	}
	if helperID == "" {
		t.Fatal("checked CRC closure has no private crc_byte helper")
	}

	input := []byte{0x12, 0x34, 0x56}
	result, err := interp.RunByteViewForTest(checked.Program, "crc32", input)
	if err != nil {
		t.Fatalf("run actual CRC byte program: %v", err)
	}
	if result.Outcome.Kind != "returned" {
		t.Fatalf("small CRC execution outcome=%+v; want a returned value", result.Outcome)
	}
	activations := 0
	for _, event := range result.Events {
		if event.Kind == "function.called" && event.CalleeFunctionID == helperID {
			activations++
		}
	}
	if activations != len(input) {
		t.Fatalf("crc_byte activations=%d for %d input bytes; want one private helper activation per byte", activations, len(input))
	}
}

func TestPhase37CRC32Consumers(t *testing.T) {
	exampleRoot := testsupport.ProjectPath("examples", "crc32")
	for _, name := range []string{"adapter.c", "adapter.h", "file_bytes.bindings.json"} {
		checksumPath := testsupport.ProjectPath("examples", "checksum", name)
		crcPath := filepath.Join(exampleRoot, name)
		want, err := os.ReadFile(checksumPath)
		if err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(crcPath)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("CRC binding copy %s differs from the canonical checksum bytes", name)
		}
	}

	var sources []string
	err := filepath.WalkDir(exampleRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".schway") {
			rel, err := filepath.Rel(testsupport.ProjectPath("."), path)
			if err != nil {
				return err
			}
			sources = append(sources, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(sources)
	wantSources := []string{
		"examples/crc32/crc32.schway",
		"examples/crc32/crc_low.schway",
		"examples/crc32/lib/crc32.schway",
	}
	if strings.Join(sources, "\n") != strings.Join(wantSources, "\n") {
		t.Fatalf("CRC local source census = %v, want exactly %v", sources, wantSources)
	}

	librarySource, err := os.ReadFile(filepath.Join(exampleRoot, "lib", "crc32.schway"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(librarySource), "while ") != 2 || strings.Count(string(librarySource), "crc_byte(crc)") != 1 {
		t.Fatal("byte and eight-round loops must remain in the one shared CRC library")
	}
	for _, entryName := range []string{"crc32.schway", "crc_low.schway"} {
		entrySource, err := os.ReadFile(filepath.Join(exampleRoot, entryName))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(entrySource), "import lib.crc32 {") != 1 || strings.Contains(string(entrySource), "while ") || strings.Contains(string(entrySource), "crc_byte(") {
			t.Fatalf("%s must share lib.crc32 without a copied loop or helper", entryName)
		}
	}
	// This old helper is a Phase 36 witness, not a second Phase 37 library.
	crcStep, err := os.ReadFile(testsupport.ProjectPath("examples", "helper_calls", "crc_step.schway"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(crcStep), "fn crc_step(") || !strings.Contains(string(crcStep), "0xEDB88320") {
		t.Fatal("historical crc_step witness disappeared from the scoped implementation census")
	}

	input := filepath.Join(t.TempDir(), "input.bin")
	if err := os.WriteFile(input, []byte("123456789"), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(exampleRoot, "file_bytes.bindings.json")
	for _, tc := range []struct {
		entry string
		want  string
	}{
		{entry: "crc32.schway", want: "3421780262\n"},
		{entry: "crc_low.schway", want: "14630\n"},
	} {
		t.Run(tc.entry, func(t *testing.T) {
			entryPath := filepath.Join(exampleRoot, tc.entry)
			checked, err := session.CheckFile(entryPath)
			if err != nil || len(checked.Diagnostics) != 0 {
				t.Fatalf("check %s: err=%v diagnostics=%+v", tc.entry, err, checked.Diagnostics)
			}
			crc32Functions := 0
			for _, function := range checked.Program.Functions {
				if function.Name == "crc32" {
					crc32Functions++
					if function.Module != "lib.crc32" {
						t.Fatalf("crc32 module = %q, want lib.crc32", function.Module)
					}
				}
			}
			if crc32Functions != 1 {
				t.Fatalf("checked closure has %d crc32 definitions, want one shared imported definition", crc32Functions)
			}
			artifact := filepath.Join(t.TempDir(), strings.TrimSuffix(tc.entry, ".schway"))
			_, diagnostics, err := session.BuildApplicationFile(context.Background(), entryPath, artifact, native.DefaultRunner(), manifest)
			if err != nil || len(diagnostics) != 0 {
				t.Fatalf("build %s with local manifest: err=%v diagnostics=%+v", tc.entry, err, diagnostics)
			}
			var stdout, stderr bytes.Buffer
			outcome, err := native.RunApplication(context.Background(), artifact, input, &stdout, &stderr)
			if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 0 {
				t.Fatalf("native %s outcome=%+v err=%v stdout=%q stderr=%q", tc.entry, outcome, err, stdout.String(), stderr.String())
			}
			if stdout.String() != tc.want || stderr.String() != "" {
				t.Fatalf("native %s streams stdout=%q stderr=%q; want stdout=%q and empty stderr", tc.entry, stdout.String(), stderr.String(), tc.want)
			}
		})
	}
}
