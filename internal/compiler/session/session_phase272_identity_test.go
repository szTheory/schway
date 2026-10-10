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
	"examples/phase272/straight_line_copy.schway": true,
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
	for _, entry := range observed {
		if !goldenPaths[entry.Path] && !phase272NewlyAcceptedPaths[entry.Path] {
			t.Errorf("%s: newly accepted source is absent from the golden and not in the allow-list", entry.Path)
		}
	}
}
