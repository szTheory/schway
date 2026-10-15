package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

type phase37CRCFixture struct {
	root, fullSource, lowSource, manifest string
}

var phase37CRCOriginHashes = map[string]string{
	"examples/crc32/crc32.schway":             "8a8c7570409a6bf96258a35b5964dc899a2347871b564fb59e081beb011388c9",
	"examples/crc32/crc_low.schway":           "797f9cd2e733c3590a593518fe671a1fda553eb6ce22ea50d70f1a26d606dccc",
	"examples/crc32/lib/crc32.schway":         "304d47059ac5f5266776bb250794a3c64ae76ce9fe7b55a2a490facf1e35fb4f",
	"examples/crc32/adapter.c":                "b758836a1d6dcff2fc6b4f903c0046ba3269ec83a8f3329c03553a90a5fbb1ae",
	"examples/crc32/adapter.h":                "2c29a00ed8939668179fddf1c41cdd5db23e5a9d8574b5ec3e841dae729b612d",
	"examples/crc32/file_bytes.bindings.json": "c2f960ff2a92a5d9a60356711812ea60595a2f81cf6c0ff2c4b259ce6380103c",
}

type phase37CRCAlgorithmMutant struct {
	name, vector, clean string
	mutate              func(*testing.T, string)
}

// TestPhase37CRC32AlgorithmMutants sends four private source mutations through
// the actual public check, interpreter, and native routes. A control counts as
// killed only after check/admission and a successful numeric run on both engines.
func TestPhase37CRC32AlgorithmMutants(t *testing.T) {
	phase37CRCRequireTrackedOrigins(t)
	t.Cleanup(func() { phase37CRCRequireTrackedOrigins(t) })
	cli := testsupport.BuildCLI(t)
	mutants := []phase37CRCAlgorithmMutant{
		{
			name: "wrong-reflected-polynomial", vector: "123456789", clean: "3421780262",
			mutate: func(t *testing.T, library string) {
				phase37CRCReplaceExact(t, library, "next = next ^ 0xEDB88320", "next = next ^ 0xEDB88321", 1)
			},
		},
		{
			name: "non-reflected-recurrence", vector: "\x80", clean: "1069182125",
			mutate: func(t *testing.T, library string) {
				phase37CRCReplaceExact(t, library, "crc = crc ^ byte_at(bytes, i)", "crc = crc ^ (byte_at(bytes, i) << 24)", 1)
				phase37CRCReplaceExact(t, library, "mixed >> 1", "mixed << 1", 2)
				phase37CRCReplaceExact(t, library, "if mixed & 1 == 1", "if mixed & 0x80000000 == 0x80000000", 1)
				phase37CRCReplaceExact(t, library, "next = next ^ 0xEDB88320", "next = next ^ 0x04C11DB7", 1)
			},
		},
		{
			name: "missing-final-xor", vector: "", clean: "0",
			mutate: func(t *testing.T, library string) {
				phase37CRCReplaceExact(t, library, "  crc = crc ^ 0xFFFFFFFF\n", "", 1)
			},
		},
		{
			name: "shortened-bit-round-bound", vector: "\x00", clean: "3523407757",
			mutate: func(t *testing.T, library string) {
				phase37CRCReplaceExact(t, library, "while round < 8", "while round < 7", 1)
			},
		},
	}

	for _, mutant := range mutants {
		mutant := mutant
		t.Run(mutant.name, func(t *testing.T) {
			clean := phase37CRCClone(t, t.TempDir())
			changed := phase37CRCClone(t, t.TempDir())
			phase37CRCWriteInput(t, clean.root, mutant.vector)
			cleanInput := filepath.Join(clean.root, "phase37-input.bin")
			changedInput := filepath.Join(changed.root, "phase37-input.bin")
			if err := os.WriteFile(changedInput, []byte(mutant.vector), 0o600); err != nil {
				t.Fatal(err)
			}
			mutant.mutate(t, filepath.Join(changed.root, "examples", "crc32", "lib", "crc32.schway"))

			phase37CRCCheckPublic(t, cli, clean, clean.fullSource)
			phase37CRCCheckPublic(t, cli, changed, changed.fullSource)
			cleanByEngine := make(map[string]phase37CLITuple, 2)
			mutantByEngine := make(map[string]phase37CLITuple, 2)
			for _, engine := range []string{"interpreter", "native"} {
				cleanTuple := phase37CRCRunPublic(t, cli, clean, clean.fullSource, cleanInput, engine)
				phase37CRCRequireExactNumber(t, cleanTuple, mutant.clean, mutant.name+" clean "+engine)
				mutantTuple := phase37CRCRunPublic(t, cli, changed, changed.fullSource, changedInput, engine)
				got := phase37CRCRequireNumber(t, mutantTuple, mutant.name+" mutant "+engine)
				want, _ := strconv.ParseUint(mutant.clean, 10, 64)
				if got == want {
					t.Fatalf("%s %s did not reach an independent numeric disagreement: got %d, clean oracle %s", mutant.name, engine, got, mutant.clean)
				}
				t.Logf("reached %s on %s: clean=%s mutated=%d", mutant.name, engine, mutant.clean, got)
				cleanByEngine[engine] = cleanTuple
				mutantByEngine[engine] = mutantTuple
			}
			if cleanByEngine["interpreter"] != cleanByEngine["native"] {
				t.Fatalf("%s clean engine parity mismatch: interpreter=%+v native=%+v", mutant.name, cleanByEngine["interpreter"], cleanByEngine["native"])
			}
			if mutantByEngine["interpreter"] != mutantByEngine["native"] {
				t.Fatalf("%s mutant engine parity mismatch: interpreter=%+v native=%+v", mutant.name, mutantByEngine["interpreter"], mutantByEngine["native"])
			}
		})
	}
}

// TestPhase37CRC32LibraryDualFlip is the fifth reached control: one admitted
// library-only result-bit edit changes both public consumers on both engines.
func TestPhase37CRC32LibraryDualFlip(t *testing.T) {
	phase37CRCRequireTrackedOrigins(t)
	t.Cleanup(func() { phase37CRCRequireTrackedOrigins(t) })
	cli := testsupport.BuildCLI(t)
	clean := phase37CRCClone(t, t.TempDir())
	changed := phase37CRCClone(t, t.TempDir())
	cleanInput := filepath.Join(clean.root, "phase37-input.bin")
	changedInput := filepath.Join(changed.root, "phase37-input.bin")
	for _, path := range []string{cleanInput, changedInput} {
		if err := os.WriteFile(path, []byte("123456789"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	phase37CRCReplaceExact(t, filepath.Join(changed.root, "examples", "crc32", "lib", "crc32.schway"), "crc = crc ^ 0xFFFFFFFF", "crc = crc ^ 0xFFFFFFFE", 1)

	for _, entry := range []struct {
		name, clean, flipped string
	}{
		{name: "crc32", clean: "3421780262", flipped: "3421780263"},
		{name: "crc_low", clean: "14630", flipped: "14631"},
	} {
		entry := entry
		t.Run(entry.name, func(t *testing.T) {
			cleanSource := filepath.Join(clean.root, "examples", "crc32", entry.name+".schway")
			changedSource := filepath.Join(changed.root, "examples", "crc32", entry.name+".schway")
			phase37CRCCheckPublic(t, cli, clean, cleanSource)
			phase37CRCCheckPublic(t, cli, changed, changedSource)
			cleanByEngine := make(map[string]phase37CLITuple, 2)
			flipByEngine := make(map[string]phase37CLITuple, 2)
			for _, engine := range []string{"interpreter", "native"} {
				cleanTuple := phase37CRCRunPublic(t, cli, clean, cleanSource, cleanInput, engine)
				phase37CRCRequireExactNumber(t, cleanTuple, entry.clean, entry.name+" clean "+engine)
				flipTuple := phase37CRCRunPublic(t, cli, changed, changedSource, changedInput, engine)
				phase37CRCRequireExactNumber(t, flipTuple, entry.flipped, entry.name+" flipped "+engine)
				if entry.clean == entry.flipped {
					t.Fatalf("%s clean and flipped pins unexpectedly match", entry.name)
				}
				cleanByEngine[engine] = cleanTuple
				flipByEngine[engine] = flipTuple
			}
			if cleanByEngine["interpreter"] != cleanByEngine["native"] || flipByEngine["interpreter"] != flipByEngine["native"] {
				t.Fatalf("%s public engine parity mismatch: clean=%+v/%+v flipped=%+v/%+v", entry.name, cleanByEngine["interpreter"], cleanByEngine["native"], flipByEngine["interpreter"], flipByEngine["native"])
			}
		})
	}
}

func phase37CRCClone(t *testing.T, root string) phase37CRCFixture {
	t.Helper()
	files := []string{
		"examples/crc32/crc32.schway",
		"examples/crc32/crc_low.schway",
		"examples/crc32/lib/crc32.schway",
		"examples/crc32/adapter.c",
		"examples/crc32/adapter.h",
		"examples/crc32/file_bytes.bindings.json",
	}
	for _, relative := range files {
		source := testsupport.ProjectPath(filepath.FromSlash(relative))
		destination := filepath.Join(root, filepath.FromSlash(relative))
		data, err := os.ReadFile(source)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(destination, data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	crcRoot := filepath.Join(root, "examples", "crc32")
	return phase37CRCFixture{
		root:       root,
		fullSource: filepath.Join(crcRoot, "crc32.schway"),
		lowSource:  filepath.Join(crcRoot, "crc_low.schway"),
		manifest:   filepath.Join(crcRoot, "file_bytes.bindings.json"),
	}
}

func phase37CRCWriteInput(t *testing.T, root, value string) string {
	t.Helper()
	path := filepath.Join(root, "phase37-input.bin")
	if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func phase37CRCReplaceExact(t *testing.T, path, before, after string, wantCount int) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(string(data), before); got != wantCount {
		t.Fatalf("mutation anchor %q in %s occurs %d times, want exactly %d", before, path, got, wantCount)
	}
	updated := strings.Replace(string(data), before, after, wantCount)
	if err := os.WriteFile(path, []byte(updated), 0o600); err != nil {
		t.Fatal(err)
	}
}

func phase37CRCCheckPublic(t *testing.T, cli string, fixture phase37CRCFixture, source string) {
	t.Helper()
	tmp := t.TempDir()
	got, err := phase37RunCLI(context.Background(), cli, []string{"check", source}, fixture.root, tmp)
	if err != nil {
		t.Fatalf("public check failed as a tool operation for %s: %v", source, err)
	}
	if got.Exit != 0 || got.Stderr != "" {
		t.Fatalf("public check did not admit %s: (%d,%q,%q)", source, got.Exit, got.Stdout, got.Stderr)
	}
}

func phase37CRCRunPublic(t *testing.T, cli string, fixture phase37CRCFixture, source, input, engine string) phase37CLITuple {
	t.Helper()
	manifest := ""
	if engine == "native" {
		manifest = fixture.manifest
	}
	tmp := t.TempDir()
	got, err := phase37RunCLI(context.Background(), cli, phase37RunArgs(engine, source, manifest, input), fixture.root, tmp)
	if err != nil {
		t.Fatalf("public %s run failed as a tool operation for %s: %v", engine, source, err)
	}
	phase37RequireTempClean(t, tmp)
	return got
}

func phase37CRCRequireExactNumber(t *testing.T, got phase37CLITuple, want, label string) {
	t.Helper()
	value := phase37CRCRequireNumber(t, got, label)
	wantValue, err := strconv.ParseUint(want, 10, 64)
	if err != nil {
		t.Fatalf("invalid literal oracle %q: %v", want, err)
	}
	if value != wantValue {
		t.Fatalf("%s output = %d, want independent literal %s", label, value, want)
	}
}

func phase37CRCRequireNumber(t *testing.T, got phase37CLITuple, label string) uint64 {
	t.Helper()
	if got.Exit != 0 || got.Stderr != "" {
		t.Fatalf("%s did not return a successful numeric result: (%d,%q,%q)", label, got.Exit, got.Stdout, got.Stderr)
	}
	line, ok := strings.CutSuffix(got.Stdout, "\n")
	if !ok || line == "" || strings.Contains(line, "\n") || strings.TrimSpace(line) != line {
		t.Fatalf("%s output is not one canonical decimal line: %q", label, got.Stdout)
	}
	value, err := strconv.ParseUint(line, 10, 64)
	if err != nil || strconv.FormatUint(value, 10) != line {
		t.Fatalf("%s output is not canonical U64 decimal: %q (parse err=%v)", label, line, err)
	}
	return value
}

func phase37CRCRequireTrackedOrigins(t *testing.T) {
	t.Helper()
	for path, want := range phase37CRCOriginHashes {
		data, err := os.ReadFile(testsupport.ProjectPath(filepath.FromSlash(path)))
		if err != nil {
			t.Fatalf("read tracked CRC origin %s: %v", path, err)
		}
		digest := sha256.Sum256(data)
		if got := hex.EncodeToString(digest[:]); got != want {
			t.Fatalf("tracked CRC origin %s sha256=%s, want pinned %s", path, got, want)
		}
	}
}
