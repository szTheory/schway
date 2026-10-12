package session_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// Phase 27.3 conformance-C byte-identity guard (D-27.3-12, D-27.3-13). The
// golden at testdata/phase273/conformance-corpus-hashes.json was generated at
// the pre-fix HEAD recorded in its revision field. Every source that
// session.Check accepts must keep the same conformance C (hash, or the
// verbatim refusal text) after the capacity-exit fix, except the one
// allow-listed loop program, whose only permitted difference is the single
// named-exit recorder line.
//
// Conformance C is obtained only through session.Phase16ControlNativeC: the
// Phase 16 public-emitter inventory tracks direct public emitter (Emit and
// EmitNative) calls in test files too (D-27.3-14).

// phase273MinAcceptedCorpus is the accepted-corpus count observed when the
// golden was generated; it keeps the walk from going inert.
const phase273MinAcceptedCorpus = 93

const phase273GoldenSchema = "phase273.conformance-corpus-hashes/1"

// phase273PriorCapacityRecorderLine is the recorder guard before the fix.
const phase273PriorCapacityRecorderLine = "  if (schway_event_count >= SCHWAY_EVENT_CAPACITY) return 0;\n"

// phase273NamedCapacityRecorderLine is the one-line substitution the fix is
// allowed to make in conformance C (D-27.3-01, D-27.3-12).
const phase273NamedCapacityRecorderLine = "  if (schway_event_count >= SCHWAY_EVENT_CAPACITY) { fputs(\"schway: event capacity exhausted\\n\", stderr); exit(69); }\n"

// phase273ChangedConformancePaths lists the only accepted sources whose
// conformance C may differ from the golden, and only by the named line.
var phase273ChangedConformancePaths = map[string]bool{
	"examples/sum_to_n.schway": true,
}

// phase273NewlyAcceptedPaths lists the only accepted sources allowed to be
// absent from the golden: sources added after the golden was generated.
var phase273NewlyAcceptedPaths = map[string]bool{
	// 2026-10-10 (Phase 28, D-20): new runnable example added after the
	// golden was generated; golden stays frozen (D-23).
	"examples/phase28/unused_function.schway": true,
	// 2026-10-11 (Phase 29, plan 29-00): adler32 byte-view library fixture;
	// golden stays frozen (D-23).
	"testdata/phase29/adler32_library.schway": true,
	// 2026-10-11 (Phase 29, plan 29-06): fixture-owner adler32 program;
	// golden stays frozen (D-23).
	"examples/phase29/adler32_fixture.schway": true,
	// 2026-10-11 (Phase 29, plan 29-10): out-of-bounds control; golden stays
	// frozen (D-23).
	"examples/phase29/byte_index_oob.schway": true,
	// 2026-10-12 (Phase 30, plan 30-09, D-27): the Adler-32 library module;
	// golden stays frozen (D-23). The importing entry is refused by the
	// byte-taking Check (module.imports_not_loaded) and never joins this walk.
	"examples/checksum/adler32.schway": true,
}

// phase273EntryPrecedencePaths lists the only golden entries whose
// EmitNative outcome may differ because main now wins entry resolution.
// 2026-10-10 (Phase 28, D-01/D-23): the golden is frozen and records the old
// callgraph.entry_ambiguous refusal; the observed entry must differ from it.
var phase273EntryPrecedencePaths = map[string]string{
	"testdata/phase11/multi_function_zero_call.schway": "D-01 main precedence; golden frozen per D-23",
}

type phase273CorpusGolden struct {
	Schema   string                `json:"schema"`
	Revision string                `json:"revision"`
	Entries  []phase273CorpusEntry `json:"entries"`
}

type phase273CorpusEntry struct {
	Path             string `json:"path"`
	EmitNativeSHA256 string `json:"emitNativeSHA256,omitempty"`
	EmitNativeError  string `json:"emitNativeError,omitempty"`
}

func phase273GoldenPath() string {
	return testsupport.ProjectPath("testdata", "phase273", "conformance-corpus-hashes.json")
}

func phase273ConformanceCorpus(t *testing.T) []phase273CorpusEntry {
	t.Helper()
	entries, _ := phase273ConformanceCorpusWithC(t)
	return entries
}

// phase273ConformanceCorpusWithC also returns the observed C text by path so
// the allow-listed entry can be reversed to the prior line.
func phase273ConformanceCorpusWithC(t *testing.T) ([]phase273CorpusEntry, map[string]string) {
	t.Helper()
	projectRoot := testsupport.ProjectPath(".")
	var entries []phase273CorpusEntry
	observedC := map[string]string{}
	for _, root := range []string{testsupport.ProjectPath("examples"), testsupport.ProjectPath("testdata")} {
		err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if info.IsDir() || !strings.HasSuffix(path, ".schway") {
				return nil
			}
			source, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if len(source) > syntax.MaxSourceBytes {
				return nil
			}
			checked := session.Check(source)
			if len(checked.Diagnostics) != 0 {
				return nil
			}
			relative, err := filepath.Rel(projectRoot, path)
			if err != nil {
				return err
			}
			entry := phase273CorpusEntry{Path: filepath.ToSlash(relative)}
			emitted, emitErr := session.Phase16ControlNativeC(checked.Program, "phase273-identity")
			if emitErr != nil {
				entry.EmitNativeError = emitErr.Error()
			} else {
				sum := sha256.Sum256([]byte(emitted))
				entry.EmitNativeSHA256 = hex.EncodeToString(sum[:])
				observedC[entry.Path] = emitted
			}
			entries = append(entries, entry)
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", root, err)
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return entries, observedC
}

func TestPhase273ConformanceCorpusByteIdentity(t *testing.T) {
	observed, observedC := phase273ConformanceCorpusWithC(t)
	goldenPath := phase273GoldenPath()
	if os.Getenv("SCHWAY_PHASE273_WRITE_GOLDEN") == "1" {
		// The revision is supplied by the caller (git rev-parse HEAD) rather
		// than spawned here, so this test never starts a subprocess.
		revision := strings.TrimSpace(os.Getenv("SCHWAY_PHASE273_GOLDEN_REVISION"))
		if revision == "" {
			t.Fatal("SCHWAY_PHASE273_GOLDEN_REVISION must name the commit the golden is generated at")
		}
		golden := phase273CorpusGolden{
			Schema:   phase273GoldenSchema,
			Revision: revision,
			Entries:  observed,
		}
		encoded, err := json.MarshalIndent(golden, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenPath, append(encoded, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %d accepted entries at %s", len(observed), golden.Revision)
		return
	}

	raw, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatal(err)
	}
	var golden phase273CorpusGolden
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatal(err)
	}
	if golden.Schema != phase273GoldenSchema {
		t.Fatalf("golden schema %q, want %q", golden.Schema, phase273GoldenSchema)
	}
	if len(golden.Entries) < phase273MinAcceptedCorpus {
		t.Fatalf("golden has %d entries, want at least %d", len(golden.Entries), phase273MinAcceptedCorpus)
	}
	byPath := map[string]phase273CorpusEntry{}
	for _, entry := range observed {
		byPath[entry.Path] = entry
	}
	goldenPaths := map[string]bool{}
	for _, want := range golden.Entries {
		goldenPaths[want.Path] = true
		got, ok := byPath[want.Path]
		if !ok {
			t.Errorf("%s: accepted at golden revision %s but no longer accepted by session.Check", want.Path, golden.Revision)
			continue
		}
		if reason, flipped := phase273EntryPrecedencePaths[want.Path]; flipped {
			if !strings.Contains(want.EmitNativeError, "callgraph.entry_ambiguous") {
				t.Errorf("%s: frozen golden must carry the old callgraph.entry_ambiguous error, got %q", want.Path, want.EmitNativeError)
			}
			if got.EmitNativeSHA256 == want.EmitNativeSHA256 && got.EmitNativeError == want.EmitNativeError {
				t.Errorf("%s: allow-listed (%s) but observed entry equals the golden", want.Path, reason)
			}
			t.Logf("%s: entry-precedence flip observed: hash=%q error=%q", want.Path, got.EmitNativeSHA256, got.EmitNativeError)
			continue
		}
		if phase273ChangedConformancePaths[want.Path] {
			if want.EmitNativeSHA256 == "" {
				t.Errorf("%s: allow-listed golden entry must carry a hash, got error %q", want.Path, want.EmitNativeError)
				continue
			}
			c := observedC[want.Path]
			if c == "" {
				t.Errorf("%s: conformance C no longer emitted: %q", want.Path, got.EmitNativeError)
				continue
			}
			if n := strings.Count(c, phase273NamedCapacityRecorderLine); n > 1 {
				t.Errorf("%s: named capacity line appears %d times, want at most 1", want.Path, n)
				continue
			}
			reversed := strings.Replace(c, phase273NamedCapacityRecorderLine, phase273PriorCapacityRecorderLine, 1)
			sum := sha256.Sum256([]byte(reversed))
			if gotHash := hex.EncodeToString(sum[:]); gotHash != want.EmitNativeSHA256 {
				t.Errorf("%s: conformance C differs from the golden beyond the one named-exit line: got %s want %s", want.Path, gotHash, want.EmitNativeSHA256)
			}
			continue
		}
		if got.EmitNativeSHA256 != want.EmitNativeSHA256 {
			t.Errorf("%s: EmitNative hash changed: got %q want %q", want.Path, got.EmitNativeSHA256, want.EmitNativeSHA256)
		}
		if got.EmitNativeError != want.EmitNativeError {
			t.Errorf("%s: EmitNative error changed: got %q want %q", want.Path, got.EmitNativeError, want.EmitNativeError)
		}
	}
	for _, entry := range observed {
		if !goldenPaths[entry.Path] && !phase273NewlyAcceptedPaths[entry.Path] {
			t.Errorf("%s: newly accepted source is absent from the golden and not in the allow-list", entry.Path)
		}
	}
}
