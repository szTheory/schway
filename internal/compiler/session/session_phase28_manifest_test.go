package session_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/callgraph"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/protocol"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// Phase 28 behavior-class manifest (D-05, D-20). The golden at
// testdata/phase28/entry-exit-class-manifest.json was generated at the
// PLAN_BASE recorded in its revision field, before any Phase 28 production
// change. Per .schway file under examples/ and testdata/ it records what
// check, build, run --engine=interpreter and run --engine=native return
// (exit class and first diagnostic code), the entry shape, the resolved
// entry and the dead functions. Native and build paths use a missing-clang
// runner, so the sweep never spawns a compiler.

const phase28ManifestSchema = "phase28.entry-exit-class-manifest/1"

// phase28MinCorpus is the corpus count observed when the golden was
// generated; it keeps the walk from going inert.
const phase28MinCorpus = 155

// phase28ClassChanges lists the only paths whose observed manifest entry may
// differ from the golden, each with the new expected entry. Later Phase 28
// plans add one entry per deliberate class change, each with a dated reason
// comment. An entry equal to the golden fails as stale.
var phase28ClassChanges = map[string]phase28ManifestEntry{}

// phase28NewlyAddedEntries lists the only paths allowed to be absent from the
// golden (sources added after generation), with their expected entry.
var phase28NewlyAddedEntries = map[string]phase28ManifestEntry{}

type phase28Manifest struct {
	Schema   string                 `json:"schema"`
	Revision string                 `json:"revision"`
	Entries  []phase28ManifestEntry `json:"entries"`
}

type phase28Class struct {
	Exit int    `json:"exit"`
	Code string `json:"code,omitempty"`
}

type phase28ManifestEntry struct {
	Path           string       `json:"path"`
	Check          phase28Class `json:"check"`
	Build          phase28Class `json:"build"`
	InterpreterRun phase28Class `json:"interpreterRun"`
	NativeRun      phase28Class `json:"nativeRun"`
	EntryShape     string       `json:"entryShape"`
	Entry          string       `json:"entry,omitempty"`
	EntryError     string       `json:"entryError,omitempty"`
	DeadFunctions  []string     `json:"deadFunctions,omitempty"`
}

func phase28ManifestPath() string {
	return testsupport.ProjectPath("testdata", "phase28", "entry-exit-class-manifest.json")
}

func phase28Classify(result protocol.Result) phase28Class {
	class := phase28Class{Exit: protocol.ExitCode(result.Status)}
	if len(result.Diagnostics) > 0 {
		class.Code = result.Diagnostics[0].Code
	}
	return class
}

func phase28Edges(program core.Program) (map[string][]string, map[string]int) {
	edges := map[string][]string{}
	inDegree := map[string]int{}
	declared := map[string]bool{}
	for _, function := range program.Functions {
		declared[function.ID] = true
		inDegree[function.ID] += 0
	}
	for _, function := range program.Functions {
		if function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind != core.OpCall || !declared[operation.CalleeID] {
				continue
			}
			edges[function.ID] = append(edges[function.ID], operation.CalleeID)
			inDegree[operation.CalleeID]++
		}
	}
	return edges, inDegree
}

func phase28Shape(checked session.CheckResult) string {
	if len(checked.Diagnostics) != 0 {
		return "unchecked"
	}
	functions := checked.Program.Functions
	if len(functions) == 0 {
		return "no-function"
	}
	_, inDegree := phase28Edges(checked.Program)
	roots := 0
	for _, function := range functions {
		if inDegree[function.ID] == 0 {
			roots++
		}
	}
	switch {
	case roots == 0:
		return "no-root"
	case roots == 1:
		return "single-root"
	}
	for _, function := range functions {
		if function.Name == "main" {
			return "multi-root-main"
		}
	}
	return "multi-root-no-main"
}

func phase28Dead(program core.Program, entryID string) []string {
	edges, _ := phase28Edges(program)
	seen := map[string]bool{entryID: true}
	queue := []string{entryID}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, next := range edges[current] {
			if !seen[next] {
				seen[next] = true
				queue = append(queue, next)
			}
		}
	}
	var dead []string
	for _, function := range program.Functions {
		if !seen[function.ID] {
			dead = append(dead, function.ID)
		}
	}
	sort.Strings(dead)
	return dead
}

func phase28BuildClass(ctx context.Context, t *testing.T, path string, runner native.Runner) phase28Class {
	t.Helper()
	_, diagnostics, err := session.BuildApplicationFile(ctx, path, filepath.Join(t.TempDir(), "app"), runner)
	if len(diagnostics) > 0 {
		return phase28Class{Exit: 2, Code: diagnostics[0].Code}
	}
	if err != nil {
		code := "native.build_failed"
		var toolError *native.ToolError
		if errors.As(err, &toolError) && toolError.Code != "" {
			code = toolError.Code
		}
		return phase28Class{Exit: 3, Code: code}
	}
	return phase28Class{}
}

func phase28ObserveCorpus(t *testing.T) []phase28ManifestEntry {
	t.Helper()
	projectRoot := testsupport.ProjectPath(".")
	runner := native.Runner{ClangPath: filepath.Join(t.TempDir(), "missing-clang")}
	ctx := context.Background()
	var entries []phase28ManifestEntry
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
			relative, err := filepath.Rel(projectRoot, path)
			if err != nil {
				return err
			}
			entry := phase28ManifestEntry{Path: filepath.ToSlash(relative)}
			checkResult, err := session.CheckCommandFile(path)
			if err != nil {
				t.Fatalf("%s: check: %v", entry.Path, err)
			}
			entry.Check = phase28Classify(checkResult)
			interpResult, err := session.RunInterpreterCommandFile(path)
			if err != nil {
				t.Fatalf("%s: interpreter run: %v", entry.Path, err)
			}
			entry.InterpreterRun = phase28Classify(interpResult)
			nativeResult, err := session.RunNativeCommandFile(ctx, path, runner)
			if err != nil {
				t.Fatalf("%s: native run: %v", entry.Path, err)
			}
			entry.NativeRun = phase28Classify(nativeResult)
			entry.Build = phase28BuildClass(ctx, t, path, runner)

			checked := session.Check(source)
			entry.EntryShape = phase28Shape(checked)
			if len(checked.Diagnostics) == 0 {
				resolved, entryErr := callgraph.EntryFunction(checked.Program)
				if entryErr == nil {
					entry.Entry = resolved.Name
					entry.DeadFunctions = phase28Dead(checked.Program, resolved.ID)
				} else if refusal, ok := callgraph.EntryAmbiguousError(entryErr); ok {
					entry.EntryError = refusal.Code()
				} else if refusal, ok := callgraph.UnresolvedCalleeError(entryErr); ok {
					entry.EntryError = refusal.Code()
				} else {
					entry.EntryError = entryErr.Error()
				}
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

func TestPhase28EntryExitClassManifest(t *testing.T) {
	observed := phase28ObserveCorpus(t)
	goldenPath := phase28ManifestPath()
	if os.Getenv("SCHWAY_PHASE28_WRITE_GOLDEN") == "1" {
		revision := strings.TrimSpace(os.Getenv("SCHWAY_PHASE28_GOLDEN_REVISION"))
		if revision == "" {
			t.Fatal("SCHWAY_PHASE28_GOLDEN_REVISION must name the commit the golden is generated at")
		}
		encoded, err := json.MarshalIndent(phase28Manifest{Schema: phase28ManifestSchema, Revision: revision, Entries: observed}, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenPath, append(encoded, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %d entries at %s", len(observed), revision)
		return
	}

	raw, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatal(err)
	}
	var golden phase28Manifest
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatal(err)
	}
	if golden.Schema != phase28ManifestSchema {
		t.Fatalf("golden schema %q, want %q", golden.Schema, phase28ManifestSchema)
	}
	if len(observed) < phase28MinCorpus {
		t.Fatalf("observed %d corpus files, want at least %d", len(observed), phase28MinCorpus)
	}
	if !sort.SliceIsSorted(golden.Entries, func(i, j int) bool { return golden.Entries[i].Path < golden.Entries[j].Path }) {
		t.Fatal("golden entries are not sorted by path")
	}
	goldenByPath := map[string]phase28ManifestEntry{}
	for _, entry := range golden.Entries {
		goldenByPath[entry.Path] = entry
	}
	observedByPath := map[string]phase28ManifestEntry{}
	for _, entry := range observed {
		observedByPath[entry.Path] = entry
	}
	for _, got := range observed {
		want, inGolden := goldenByPath[got.Path]
		added, isAdded := phase28NewlyAddedEntries[got.Path]
		changed, isChanged := phase28ClassChanges[got.Path]
		switch {
		case !inGolden && !isAdded:
			t.Errorf("%s: absent from the golden and not in phase28NewlyAddedEntries: %+v", got.Path, got)
		case !inGolden:
			if !reflect.DeepEqual(got, added) {
				t.Errorf("%s: newly added entry differs from allow-list: got %+v want %+v", got.Path, got, added)
			}
		case isChanged:
			if reflect.DeepEqual(changed, want) {
				t.Errorf("%s: stale phase28ClassChanges entry equals the golden", got.Path)
			}
			if !reflect.DeepEqual(got, changed) {
				t.Errorf("%s: observed differs from phase28ClassChanges: got %+v want %+v", got.Path, got, changed)
			}
		default:
			if !reflect.DeepEqual(got, want) {
				t.Errorf("%s: class changed without allow-list: got %+v want %+v", got.Path, got, want)
			}
		}
	}
	for path := range goldenByPath {
		if _, ok := observedByPath[path]; !ok {
			t.Errorf("%s: in the golden at %s but no longer observed", path, golden.Revision)
		}
	}
}
