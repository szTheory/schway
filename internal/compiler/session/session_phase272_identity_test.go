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

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// Phase 27.2 byte-identity guard (D-27.2-05, D-27.2-14). The golden at
// testdata/phase272/accepted-corpus-hashes.json was generated at the pre-fix
// HEAD recorded in its revision field. Every source that session.Check accepts
// must keep the same core-document hash and the same cgen.EmitApplication
// result (hash, or the verbatim refusal text) after the producer fix.
//
// Only cgen.EmitApplication is called here: the Phase 16 public-emitter
// inventory tracks every other public emitter call in test files too.

// phase272MinAcceptedCorpus is the accepted-corpus count observed when the
// golden was generated; it keeps the walk from going inert.
const phase272MinAcceptedCorpus = 92

// phase272NewlyAcceptedPaths lists the only accepted sources allowed to be
// absent from the golden: sources added after the golden was generated.
var phase272NewlyAcceptedPaths = map[string]bool{
	// 2026-10-14 (Phase 35, D-07): new public bitwise witness, absent at
	// the frozen golden revision; historical sources/hashes remain unchanged.
	"examples/bitwise.schway":                     true,
	"examples/phase272/straight_line_copy.schway": true,
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

// phase272SuccessorAcceptedPaths composes the byte-checkable Phase36 D-20
// and Phase37 Plan01 additions with the frozen identity corpora. The existing
// inventory owner validates both complete finite ledgers, source hashes,
// reasons, and tracked-source closure before either path gains an allowance.
// TestPhase34PreChangeCGoldenIsNotInert exercises its omission, hash, path,
// and old-source-exemption controls; neither historical golden is rewritten.
func phase272SuccessorAcceptedPaths(t *testing.T) map[string]bool {
	t.Helper()
	phase34TrackedSchwayPaths(t)
	return map[string]bool{
		"examples/helper_calls/crc_step.schway": true,
		"examples/crc32/lib/crc32.schway":       true,
	}
}

// phase272EntryPrecedencePaths lists the only golden entries whose
// EmitApplication outcome may differ because main now wins entry resolution.
// 2026-10-10 (Phase 28, D-01/D-23): the golden is frozen and records the old
// callgraph.entry_ambiguous refusal; the observed entry must differ from it.
var phase272EntryPrecedencePaths = map[string]string{
	"testdata/phase11/multi_function_zero_call.schway": "D-01 main precedence; golden frozen per D-23",
}

type phase272CorpusGolden struct {
	Schema   string                `json:"schema"`
	Revision string                `json:"revision"`
	Entries  []phase272CorpusEntry `json:"entries"`
}

type phase272CorpusEntry struct {
	Path                  string `json:"path"`
	CoreSHA256            string `json:"coreSHA256"`
	EmitApplicationSHA256 string `json:"emitApplicationSHA256,omitempty"`
	EmitApplicationError  string `json:"emitApplicationError,omitempty"`
}

const phase272GoldenSchema = "phase272.accepted-corpus-hashes/1"

func phase272GoldenPath() string {
	return testsupport.ProjectPath("testdata", "phase272", "accepted-corpus-hashes.json")
}

func phase272AcceptedCorpus(t *testing.T) []phase272CorpusEntry {
	t.Helper()
	projectRoot := testsupport.ProjectPath(".")
	var entries []phase272CorpusEntry
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
			document, err := json.Marshal(checked.Program)
			if err != nil {
				return err
			}
			coreSum := sha256.Sum256(document)
			entry := phase272CorpusEntry{Path: filepath.ToSlash(relative), CoreSHA256: hex.EncodeToString(coreSum[:])}
			emitted, emitErr := cgen.EmitApplication(checked.Program)
			if emitErr != nil {
				entry.EmitApplicationError = emitErr.Error()
			} else {
				emitSum := sha256.Sum256([]byte(emitted))
				entry.EmitApplicationSHA256 = hex.EncodeToString(emitSum[:])
			}
			entries = append(entries, entry)
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", root, err)
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return entries
}

func TestPhase272AcceptedCorpusByteIdentity(t *testing.T) {
	observed := phase272AcceptedCorpus(t)
	goldenPath := phase272GoldenPath()
	if os.Getenv("SCHWAY_PHASE272_WRITE_GOLDEN") == "1" {
		// The revision is supplied by the caller (git rev-parse HEAD) rather
		// than spawned here, so this test never starts a subprocess.
		revision := strings.TrimSpace(os.Getenv("SCHWAY_PHASE272_GOLDEN_REVISION"))
		if revision == "" {
			t.Fatal("SCHWAY_PHASE272_GOLDEN_REVISION must name the commit the golden is generated at")
		}
		golden := phase272CorpusGolden{
			Schema:   phase272GoldenSchema,
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
	var golden phase272CorpusGolden
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatal(err)
	}
	if golden.Schema != phase272GoldenSchema {
		t.Fatalf("golden schema %q, want %q", golden.Schema, phase272GoldenSchema)
	}
	if len(golden.Entries) < phase272MinAcceptedCorpus {
		t.Fatalf("golden has %d entries, want at least %d", len(golden.Entries), phase272MinAcceptedCorpus)
	}
	byPath := map[string]phase272CorpusEntry{}
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
		if reason, flipped := phase272EntryPrecedencePaths[want.Path]; flipped {
			if got.CoreSHA256 != want.CoreSHA256 {
				t.Errorf("%s: core document hash changed (%s): got %s want %s", want.Path, reason, got.CoreSHA256, want.CoreSHA256)
			}
			if !strings.Contains(want.EmitApplicationError, "callgraph.entry_ambiguous") {
				t.Errorf("%s: frozen golden must carry the old callgraph.entry_ambiguous error, got %q", want.Path, want.EmitApplicationError)
			}
			if got.EmitApplicationSHA256 == want.EmitApplicationSHA256 && got.EmitApplicationError == want.EmitApplicationError {
				t.Errorf("%s: allow-listed (%s) but observed entry equals the golden", want.Path, reason)
			}
			t.Logf("%s: entry-precedence flip observed: hash=%q error=%q", want.Path, got.EmitApplicationSHA256, got.EmitApplicationError)
			continue
		}
		if got.CoreSHA256 != want.CoreSHA256 {
			t.Errorf("%s: core document hash changed: got %s want %s", want.Path, got.CoreSHA256, want.CoreSHA256)
		}
		if got.EmitApplicationSHA256 != want.EmitApplicationSHA256 {
			t.Errorf("%s: EmitApplication hash changed: got %q want %q", want.Path, got.EmitApplicationSHA256, want.EmitApplicationSHA256)
		}
		if got.EmitApplicationError != want.EmitApplicationError {
			t.Errorf("%s: EmitApplication error changed: got %q want %q", want.Path, got.EmitApplicationError, want.EmitApplicationError)
		}
	}
	successorPaths := phase272SuccessorAcceptedPaths(t)
	for _, entry := range observed {
		if !goldenPaths[entry.Path] && !phase272NewlyAcceptedPaths[entry.Path] && !successorPaths[entry.Path] {
			t.Errorf("%s: newly accepted source is absent from the golden and not in the allow-list", entry.Path)
		}
	}
}
