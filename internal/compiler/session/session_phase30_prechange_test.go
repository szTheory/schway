package session_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// Phase 30 pre-change byte-identity golden (D-15, plan 30-00). The golden at
// testdata/phase30/prechange-corpus.json was generated at the PLAN_BASE commit
// recorded in its revision field, before any module change. For every
// committed .schway under examples/ and testdata/, it pins SHA-256 of the
// CheckCommandFile result JSON, the core JSON, cgen.EmitApplication output (or
// error), Phase16ControlNativeC output (or error), the DebugMapCommandFile
// result JSON (or error), and the syntax.Format output.
//
// The evidence, explain and query routes stay single-file (D-22) on unchanged
// code paths. Their byte identity follows from unchanged core and C and is
// already pinned by TestPhase272AcceptedCorpusByteIdentity and
// TestPhase273ConformanceCorpusByteIdentity, so no evidence, explain or query
// field is recorded here (D-15 scope note, 2026-10-12).
//
// The golden is never regenerated after a production change in this phase.
//
// Only cgen.EmitApplication and session.Phase16ControlNativeC are called as
// emitters: the Phase 16 public-emitter inventory tracks every other public
// emitter call in test files too.

const phase30GoldenSchema = "phase30.prechange-corpus/1"

// phase30PreChangeMinCorpus is the entry count observed when the golden was
// generated (walked files plus the one generated row); it keeps the walk from
// going inert.
const phase30PreChangeMinCorpus = 160

// phase30GeneratedOverWorkKey names the generated no-import row. It is not a
// path: it is built in a TempDir at write and compare time and never committed
// (D-27).
const phase30GeneratedOverWorkKey = "generated:phase30/no-import-over-work"

// phase30OverWorkFunctions and phase30OverWorkStatements size the generated
// source so session.Check reports check.work_limit while the source stays
// within syntax.MaxSourceBytes, syntax.MaxTokens and the per-file function cap.
// Counts only; no timing is pinned (D-28).
const (
	phase30OverWorkFunctions  = 400
	phase30OverWorkStatements = 10
)

// phase30NewCommittedPaths lists the only walked sources allowed to be absent
// from the golden: sources committed after the golden was generated.
var phase30NewCommittedPaths = map[string]bool{
	// 2026-10-12 (Phase 30, D-27): the adler32 module library and its CLI
	// entry are new committed examples added after the golden was generated;
	// the golden stays frozen (D-15).
	"examples/checksum/adler32.schway":     true,
	"examples/checksum/adler32_cli.schway": true,
	// 2026-10-12 (Phase 31, plan 31-01): the bounded file checksum entry.
	"examples/checksum/checksum.schway": true,
	// 2026-10-12 (Phase 31, plan 31-07): the second consumer of the library.
	"examples/checksum/adler_low.schway": true,
}

type phase30PreChangeGolden struct {
	Schema   string                  `json:"schema"`
	Revision string                  `json:"revision"`
	Entries  []phase30PreChangeEntry `json:"entries"`
}

type phase30PreChangeEntry struct {
	Path            string `json:"path"`
	CheckSHA256     string `json:"check_sha256"`
	CoreSHA256      string `json:"core_sha256"`
	EmitApplication string `json:"emit_application"`
	ControlNativeC  string `json:"control_native_c"`
	DebugMap        string `json:"debug_map"`
	FormatSHA256    string `json:"format_sha256"`
}

func phase30GoldenPath() string {
	return testsupport.ProjectPath("testdata", "phase30", "prechange-corpus.json")
}

func phase30Hash(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func phase30JSONHash(t *testing.T, value any) string {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return phase30Hash(encoded)
}

// phase30OverWorkNoImportSource builds the deterministic no-import source for
// the generated over-work row. Plan 30-08 reuses it to prove the debug-map
// route keeps its guard-free result.
func phase30OverWorkNoImportSource() []byte {
	var b strings.Builder
	b.WriteString("module examples.phase30.over_work\n\nexport {\n  fn main\n}\n\n")
	for i := 0; i < phase30OverWorkFunctions; i++ {
		fmt.Fprintf(&b, "fn f%d(n: U64) -> U64 {\n", i)
		for j := 0; j < phase30OverWorkStatements; j++ {
			fmt.Fprintf(&b, "  let v%d = n + %d\n", j, j)
		}
		b.WriteString("  n\n}\n\n")
	}
	b.WriteString("fn main(n: U64) -> U64 {\n  n\n}\n")
	return []byte(b.String())
}

// phase30DebugMapField records the DebugMapCommandFile result JSON hash, the
// verbatim error, or the recovered panic text. Some committed sources already
// panic inside debugmap.Build at PLAN_BASE; that behavior is part of the bytes
// plan 30-08 must not move, so it is pinned rather than hidden.
func phase30DebugMapField(t *testing.T, path string) (field string) {
	t.Helper()
	defer func() {
		if recovered := recover(); recovered != nil {
			field = fmt.Sprintf("panic: %v", recovered)
		}
	}()
	result, err := session.DebugMapCommandFile(path, "")
	if err != nil {
		return "error: " + err.Error()
	}
	return phase30JSONHash(t, result)
}

// phase30ObserveFile records the six pinned fields for one source file on disk.
func phase30ObserveFile(t *testing.T, key, path string) phase30PreChangeEntry {
	t.Helper()
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	entry := phase30PreChangeEntry{Path: key}

	checkResult, checkErr := session.CheckCommandFile(path)
	if checkErr != nil {
		entry.CheckSHA256 = "error: " + checkErr.Error()
	} else {
		entry.CheckSHA256 = phase30JSONHash(t, checkResult)
	}

	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		const skipped = "skipped: check diagnostics"
		entry.CoreSHA256 = skipped
		entry.EmitApplication = skipped
		entry.ControlNativeC = skipped
	} else {
		entry.CoreSHA256 = phase30JSONHash(t, checked.Program)
		if emitted, emitErr := cgen.EmitApplication(checked.Program); emitErr != nil {
			entry.EmitApplication = "error: " + emitErr.Error()
		} else {
			entry.EmitApplication = phase30Hash([]byte(emitted))
		}
		if emitted, emitErr := session.Phase16ControlNativeC(checked.Program, "phase30-prechange"); emitErr != nil {
			entry.ControlNativeC = "error: " + emitErr.Error()
		} else {
			entry.ControlNativeC = phase30Hash([]byte(emitted))
		}
	}

	entry.DebugMap = phase30DebugMapField(t, path)

	entry.FormatSHA256 = phase30Hash(syntax.Format(syntax.Parse(source).Tree))
	return entry
}

func phase30Observe(t *testing.T) []phase30PreChangeEntry {
	t.Helper()
	projectRoot := testsupport.ProjectPath(".")
	var entries []phase30PreChangeEntry
	for _, root := range []string{testsupport.ProjectPath("examples"), testsupport.ProjectPath("testdata")} {
		err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if info.IsDir() || !strings.HasSuffix(path, ".schway") {
				return nil
			}
			if info.Size() > int64(syntax.MaxSourceBytes) {
				return nil
			}
			relative, err := filepath.Rel(projectRoot, path)
			if err != nil {
				return err
			}
			entries = append(entries, phase30ObserveFile(t, filepath.ToSlash(relative), path))
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", root, err)
		}
	}

	generated := filepath.Join(t.TempDir(), "over_work.schway")
	source := phase30OverWorkNoImportSource()
	if len(source) > syntax.MaxSourceBytes {
		t.Fatalf("generated over-work source is %d bytes, over the %d byte limit", len(source), syntax.MaxSourceBytes)
	}
	if err := os.WriteFile(generated, source, 0o644); err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	encodedDiagnostics, err := json.Marshal(checked.Diagnostics)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encodedDiagnostics), "check.work_limit") {
		t.Fatalf("generated over-work source must report check.work_limit, got %s", encodedDiagnostics)
	}
	entries = append(entries, phase30ObserveFile(t, phase30GeneratedOverWorkKey, generated))

	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return entries
}

// phase30Compare reports every difference between the golden and the observed
// entries. An empty result means byte identity holds.
func phase30Compare(golden phase30PreChangeGolden, observed []phase30PreChangeEntry) []string {
	var problems []string
	byPath := map[string]phase30PreChangeEntry{}
	for _, entry := range observed {
		byPath[entry.Path] = entry
	}
	goldenPaths := map[string]bool{}
	for _, want := range golden.Entries {
		goldenPaths[want.Path] = true
		got, ok := byPath[want.Path]
		if !ok {
			if want.Path == phase30GeneratedOverWorkKey {
				problems = append(problems, want.Path+": generated row missing")
			} else {
				problems = append(problems, want.Path+": recorded at golden revision "+golden.Revision+" but no longer present")
			}
			continue
		}
		fields := []struct{ name, got, want string }{
			{"check_sha256", got.CheckSHA256, want.CheckSHA256},
			{"core_sha256", got.CoreSHA256, want.CoreSHA256},
			{"emit_application", got.EmitApplication, want.EmitApplication},
			{"control_native_c", got.ControlNativeC, want.ControlNativeC},
			{"debug_map", got.DebugMap, want.DebugMap},
			{"format_sha256", got.FormatSHA256, want.FormatSHA256},
		}
		for _, field := range fields {
			if field.got != field.want {
				problems = append(problems, fmt.Sprintf("%s: %s changed: got %q want %q", want.Path, field.name, field.got, field.want))
			}
		}
	}
	for _, entry := range observed {
		if !goldenPaths[entry.Path] && !phase30NewCommittedPaths[entry.Path] {
			problems = append(problems, entry.Path+": walked source is absent from the golden and not in the allow-list")
		}
	}
	return problems
}

func phase30LoadGolden(t *testing.T) phase30PreChangeGolden {
	t.Helper()
	raw, err := os.ReadFile(phase30GoldenPath())
	if err != nil {
		t.Fatal(err)
	}
	var golden phase30PreChangeGolden
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatal(err)
	}
	if golden.Schema != phase30GoldenSchema {
		t.Fatalf("golden schema %q, want %q", golden.Schema, phase30GoldenSchema)
	}
	return golden
}

func TestPhase30PreChangeGolden(t *testing.T) {
	observed := phase30Observe(t)
	if os.Getenv("SCHWAY_PHASE30_WRITE_GOLDEN") == "1" {
		// The revision is supplied by the caller (git rev-parse HEAD) rather
		// than spawned here, so this test never starts a subprocess.
		revision := strings.TrimSpace(os.Getenv("SCHWAY_PHASE30_GOLDEN_REVISION"))
		if revision == "" {
			t.Fatal("SCHWAY_PHASE30_GOLDEN_REVISION must name the commit the golden is generated at")
		}
		golden := phase30PreChangeGolden{Schema: phase30GoldenSchema, Revision: revision, Entries: observed}
		encoded, err := json.MarshalIndent(golden, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(phase30GoldenPath()), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(phase30GoldenPath(), append(encoded, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %d entries at %s", len(observed), golden.Revision)
		return
	}

	golden := phase30LoadGolden(t)
	if len(golden.Entries) < phase30PreChangeMinCorpus {
		t.Fatalf("golden has %d entries, want at least %d", len(golden.Entries), phase30PreChangeMinCorpus)
	}
	for _, problem := range phase30Compare(golden, observed) {
		t.Error(problem)
	}
}

func TestPhase30PreChangeGoldenIsNotInert(t *testing.T) {
	golden := phase30LoadGolden(t)
	if len(golden.Entries) == 0 {
		t.Fatal("golden has no entries")
	}
	observed := phase30Observe(t)
	if problems := phase30Compare(golden, observed); len(problems) != 0 {
		t.Fatalf("baseline comparison must be clean before the control: %v", problems)
	}

	mutated := phase30PreChangeGolden{Schema: golden.Schema, Revision: golden.Revision, Entries: append([]phase30PreChangeEntry(nil), golden.Entries...)}
	target := &mutated.Entries[0]
	flipped := []byte(target.CheckSHA256)
	if flipped[0] == '0' {
		flipped[0] = '1'
	} else {
		flipped[0] = '0'
	}
	target.CheckSHA256 = string(flipped)
	if problems := phase30Compare(mutated, observed); len(problems) == 0 {
		t.Fatal("flipping one recorded hash was not reported: the golden comparison is inert")
	}
}
