package session_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
var phase28ClassChanges = map[string]phase28ManifestEntry{
	// 2026-10-10 (Phase 28, D-10/D-18): native run now admits the FULL program
	// before pruning or any tool, so the dead function `touch` (omitted origin)
	// is refused with core.origin_omitted like check/build/interpreter
	// (native.tool_failure -> exit 2; superseding the 28-05 tool_missing).
	"testdata/phase11/multi_function_gate_corpus.schway": {
		Path:           "testdata/phase11/multi_function_gate_corpus.schway",
		Check:          phase28Class{Exit: 2, Code: "core.origin_omitted"},
		Build:          phase28Class{Exit: 2, Code: "core.origin_omitted"},
		InterpreterRun: phase28Class{Exit: 2, Code: "core.origin_omitted"},
		NativeRun:      phase28Class{Exit: 2, Code: "core.origin_omitted"},
		EntryShape:     "multi-root-main",
		Entry:          "main",
		DeadFunctions:  []string{"s1:phase11.multi_function_gate_corpus:fn:touch"},
	},
	// 2026-10-10 (Phase 28, D-10/D-12): native run admits before any tool, so
	// the source refusal is exit 2 with the check code, not an operational
	// native.tool_* failure.
	"testdata/phase07/clean_but_unpublishable.schway": {
		Path:           "testdata/phase07/clean_but_unpublishable.schway",
		Check:          phase28Class{Exit: 2, Code: "core.origin_omitted"},
		Build:          phase28Class{Exit: 2, Code: "core.origin_omitted"},
		InterpreterRun: phase28Class{Exit: 2, Code: "core.origin_omitted"},
		NativeRun:      phase28Class{Exit: 2, Code: "core.origin_omitted"},
		EntryShape:     "single-root",
		Entry:          "relay",
	},
	// 2026-10-10 (Phase 28, D-09/D-10): build, interpreter run and native run
	// report the corevalidate refusal as a diagnostic (exit 2
	// core.duplicate_function_id) like check; was native.build_failed /
	// tool.run_failed / native.tool_failure (exit 3).
	"testdata/phase07/duplicate_function_name.schway": {
		Path:           "testdata/phase07/duplicate_function_name.schway",
		Check:          phase28Class{Exit: 2, Code: "core.duplicate_function_id"},
		Build:          phase28Class{Exit: 2, Code: "core.duplicate_function_id"},
		InterpreterRun: phase28Class{Exit: 2, Code: "core.duplicate_function_id"},
		NativeRun:      phase28Class{Exit: 2, Code: "core.duplicate_function_id"},
		EntryShape:     "single-root",
		Entry:          "main",
	},
	// 2026-10-10 (Phase 28, D-10/D-12): native run admits before any tool, so
	// the source refusal is exit 2 with the check code, not an operational
	// native.tool_* failure.
	"testdata/phase10/compose_per_path_borrow_callee_accept.schway": {
		Path:           "testdata/phase10/compose_per_path_borrow_callee_accept.schway",
		Check:          phase28Class{Exit: 2, Code: "core.origin_omitted"},
		Build:          phase28Class{Exit: 2, Code: "core.origin_omitted"},
		InterpreterRun: phase28Class{Exit: 2, Code: "core.origin_omitted"},
		NativeRun:      phase28Class{Exit: 2, Code: "core.origin_omitted"},
		EntryShape:     "single-root",
		Entry:          "callee",
	},
	// 2026-10-10 (Phase 28, D-10/D-12): native run admits before any tool, so
	// the source refusal is exit 2 with the check code, not an operational
	// native.tool_* failure.
	"testdata/phase25/exclusive_escape_reject.schway": {
		Path:           "testdata/phase25/exclusive_escape_reject.schway",
		Check:          phase28Class{Exit: 2, Code: "core.origin_omitted"},
		Build:          phase28Class{Exit: 2, Code: "core.origin_omitted"},
		InterpreterRun: phase28Class{Exit: 2, Code: "core.origin_omitted"},
		NativeRun:      phase28Class{Exit: 2, Code: "core.origin_omitted"},
		EntryShape:     "single-root",
		Entry:          "relay",
	},
	// 2026-10-10 (Phase 28, D-10/D-12): native run admits before any tool, so
	// the source refusal is exit 2 with the check code, not an operational
	// native.tool_* failure.
	// 2026-10-10 (Phase 28, D-09): the interpreter command file now admits (and
	// refuses core.origin_omitted, exit 2) instead of failing later with
	// tool.run_failed (exit 3); native run likewise per D-10.
	"testdata/phase25/shared_copy_accept.schway": {
		Path:           "testdata/phase25/shared_copy_accept.schway",
		Check:          phase28Class{Exit: 2, Code: "core.origin_omitted"},
		Build:          phase28Class{Exit: 2, Code: "core.origin_omitted"},
		InterpreterRun: phase28Class{Exit: 2, Code: "core.origin_omitted"},
		NativeRun:      phase28Class{Exit: 2, Code: "core.origin_omitted"},
		EntryShape:     "single-root",
		Entry:          "shared_copy",
	},
	// 2026-10-10 (Phase 28, D-10/D-12): native run admits before any tool, so
	// the source refusal is exit 2 with the check code, not an operational
	// native.tool_* failure.
	"testdata/phase3/public_view_mixed_access.schway": {
		Path:           "testdata/phase3/public_view_mixed_access.schway",
		Check:          phase28Class{Exit: 2, Code: "core.origin_access_mismatch"},
		Build:          phase28Class{Exit: 2, Code: "core.origin_access_mismatch"},
		InterpreterRun: phase28Class{Exit: 2, Code: "core.origin_access_mismatch"},
		NativeRun:      phase28Class{Exit: 2, Code: "core.origin_access_mismatch"},
		EntryShape:     "single-root",
		Entry:          "view",
	},
	// 2026-10-10 (Phase 28, D-10/D-12): native run admits before any tool, so
	// the source refusal is exit 2 with the check code, not an operational
	// native.tool_* failure.
	"testdata/phase3/public_view_multi_arm_access_conflict.schway": {
		Path:           "testdata/phase3/public_view_multi_arm_access_conflict.schway",
		Check:          phase28Class{Exit: 2, Code: "core.origin_omitted"},
		Build:          phase28Class{Exit: 2, Code: "core.origin_omitted"},
		InterpreterRun: phase28Class{Exit: 2, Code: "core.origin_omitted"},
		NativeRun:      phase28Class{Exit: 2, Code: "core.origin_omitted"},
		EntryShape:     "single-root",
		Entry:          "choose",
	},
	// 2026-10-10 (Phase 28, D-10/D-12): native run admits before any tool, so
	// the source refusal is exit 2 with the check code, not an operational
	// native.tool_* failure.
	"testdata/phase3/public_view_multi_arm_omitted.schway": {
		Path:           "testdata/phase3/public_view_multi_arm_omitted.schway",
		Check:          phase28Class{Exit: 2, Code: "core.origin_omitted"},
		Build:          phase28Class{Exit: 2, Code: "core.origin_omitted"},
		InterpreterRun: phase28Class{Exit: 2, Code: "core.origin_omitted"},
		NativeRun:      phase28Class{Exit: 2, Code: "core.origin_omitted"},
		EntryShape:     "single-root",
		Entry:          "choose",
	},
	// 2026-10-10 (Phase 28, D-10/D-12): native run admits before any tool, so
	// the source refusal is exit 2 with the check code, not an operational
	// native.tool_* failure.
	"testdata/phase3/public_view_omitted.schway": {
		Path:           "testdata/phase3/public_view_omitted.schway",
		Check:          phase28Class{Exit: 2, Code: "core.origin_omitted"},
		Build:          phase28Class{Exit: 2, Code: "core.origin_omitted"},
		InterpreterRun: phase28Class{Exit: 2, Code: "core.origin_omitted"},
		NativeRun:      phase28Class{Exit: 2, Code: "core.origin_omitted"},
		EntryShape:     "single-root",
		Entry:          "relay",
	},
	// 2026-10-10 (Phase 28, D-10/D-12): native run admits before any tool, so
	// the source refusal is exit 2 with the check code, not an operational
	// native.tool_* failure.
	"testdata/phase4/foreign_origin_omitted.schway": {
		Path:           "testdata/phase4/foreign_origin_omitted.schway",
		Check:          phase28Class{Exit: 2, Code: "core.foreign_origin_omitted"},
		Build:          phase28Class{Exit: 2, Code: "core.foreign_origin_omitted"},
		InterpreterRun: phase28Class{Exit: 2, Code: "core.foreign_origin_omitted"},
		NativeRun:      phase28Class{Exit: 2, Code: "core.foreign_origin_omitted"},
		EntryShape:     "single-root",
		Entry:          "main",
	},
	// 2026-10-10 (Phase 28, D-10/D-12): native run admits before any tool, so
	// the source refusal is exit 2 with the check code, not an operational
	// native.tool_* failure.
	"testdata/phase5/false_restrict_hoist.schway": {
		Path:           "testdata/phase5/false_restrict_hoist.schway",
		Check:          phase28Class{Exit: 2, Code: "core.origin_omitted"},
		Build:          phase28Class{Exit: 2, Code: "core.origin_omitted"},
		InterpreterRun: phase28Class{Exit: 2, Code: "core.origin_omitted"},
		NativeRun:      phase28Class{Exit: 2, Code: "core.origin_omitted"},
		EntryShape:     "single-root",
		Entry:          "touch",
	},
	// 2026-10-10 (Phase 28, D-10/D-12): native run admits before any tool, so
	// the source refusal is exit 2 with the check code, not an operational
	// native.tool_* failure.
	"testdata/phase5/restrict_borrow.schway": {
		Path:           "testdata/phase5/restrict_borrow.schway",
		Check:          phase28Class{Exit: 2, Code: "core.origin_omitted"},
		Build:          phase28Class{Exit: 2, Code: "core.origin_omitted"},
		InterpreterRun: phase28Class{Exit: 2, Code: "core.origin_omitted"},
		NativeRun:      phase28Class{Exit: 2, Code: "core.origin_omitted"},
		EntryShape:     "single-root",
		Entry:          "touch",
	},
	// 2026-10-10 (Phase 28, D-01/D-20): main wins the tie, the unused function
	// is pruned, interpreter run now succeeds and native run reaches the runner.
	"testdata/phase11/multi_function_zero_call.schway": {
		Path:           "testdata/phase11/multi_function_zero_call.schway",
		Check:          phase28Class{},
		Build:          phase28Class{Exit: 3, Code: "native.build_failed"},
		InterpreterRun: phase28Class{},
		NativeRun:      phase28Class{Exit: 3, Code: "native.tool_missing"},
		EntryShape:     "multi-root-main",
		Entry:          "main",
		DeadFunctions:  []string{"s1:phase11.multi_function_zero_call:fn:unused"},
	},
	// 2026-10-10 (Phase 28, D-01/D-02/D-20): no fn main and two equal-closure
	// roots; every surface now refuses with core.entry_ambiguous (exit 2).
	"testdata/phase12/payload_borrow_interaction.schway": {
		Path:           "testdata/phase12/payload_borrow_interaction.schway",
		Check:          phase28Class{Exit: 2, Code: "core.entry_ambiguous"},
		Build:          phase28Class{Exit: 2, Code: "core.entry_ambiguous"},
		InterpreterRun: phase28Class{Exit: 2, Code: "core.entry_ambiguous"},
		NativeRun:      phase28Class{Exit: 2, Code: "core.entry_ambiguous"},
		EntryShape:     "multi-root-no-main",
		EntryError:     "core.entry_ambiguous",
	},
}

// phase28NewlyAddedEntries lists the only paths allowed to be absent from the
// golden (sources added after generation), with their expected entry.
var phase28NewlyAddedEntries = map[string]phase28ManifestEntry{
	// 2026-10-10 (Phase 28, D-04, ROADMAP SC3): new runnable example with an
	// uncalled function; Build and native run stop at the missing clang.
	"examples/phase28/unused_function.schway": {
		Path:           "examples/phase28/unused_function.schway",
		Check:          phase28Class{},
		Build:          phase28Class{Exit: 3, Code: "native.tool_missing"},
		InterpreterRun: phase28Class{Exit: 3, Code: "tool.run_failed"},
		NativeRun:      phase28Class{Exit: 3, Code: "native.tool_failure"},
		EntryShape:     "multi-root-main",
		Entry:          "main",
		DeadFunctions:  []string{"s1:examples.phase28.unused_function:fn:unused"},
	},
	// 2026-10-11 (Phase 29, plan 29-00): the adler32 byte-view library; its
	// adler32 function is dead from main and pruned per engine.
	"testdata/phase29/adler32_library.schway": {
		Path:           "testdata/phase29/adler32_library.schway",
		Check:          phase28Class{},
		Build:          phase28Class{Exit: 3, Code: "native.tool_missing"},
		InterpreterRun: phase28Class{Exit: 3, Code: "tool.run_failed"},
		NativeRun:      phase28Class{Exit: 3, Code: "native.tool_failure"},
		EntryShape:     "multi-root-main",
		Entry:          "main",
		DeadFunctions:  []string{"s1:testdata.phase29.adler32_library:fn:adler32"},
	},
	// 2026-10-11 (Phase 29, plan 29-06; reconfirmed plan 29-07): the
	// fixture-owner program checks clean and has one root (main calls
	// adler32). These classes are observed WITHOUT a binding manifest, the way
	// this walker invokes build and run: Build stops at native.build_failed
	// because the fixture owner needs its manifest (the manifest-bearing
	// `schway build --manifest` plus `app run` route is the runnable gain,
	// pinned by TestPhase29FixtureAppCLI), and `schway run` is unchanged for
	// this program (D-06/D-16).
	"examples/phase29/adler32_fixture.schway": {
		Path:           "examples/phase29/adler32_fixture.schway",
		Check:          phase28Class{},
		Build:          phase28Class{Exit: 3, Code: "native.build_failed"},
		InterpreterRun: phase28Class{Exit: 3, Code: "tool.run_failed"},
		NativeRun:      phase28Class{Exit: 3, Code: "native.tool_failure"},
		EntryShape:     "single-root",
		Entry:          "main",
	},
	// 2026-10-11 (Phase 29, plan 29-10): the out-of-bounds control; same
	// manifest-less observations as the fixture program it is cut from.
	"examples/phase29/byte_index_oob.schway": {
		Path:           "examples/phase29/byte_index_oob.schway",
		Check:          phase28Class{},
		Build:          phase28Class{Exit: 3, Code: "native.build_failed"},
		InterpreterRun: phase28Class{Exit: 3, Code: "tool.run_failed"},
		NativeRun:      phase28Class{Exit: 3, Code: "native.tool_failure"},
		EntryShape:     "single-root",
		Entry:          "main",
	},
	// 2026-10-12 (Phase 31, plan 31-00, D-01): moved into examples/checksum/.
	// 2026-10-12 (Phase 30, plan 30-09, D-27): the Adler-32 library module.
	// The walker reads each file alone, so the library is a one-root program
	// whose only function is adler32; build and run need an entry with main
	// and fail the same operational way the Phase 29 fixture does without a
	// manifest.
	"examples/checksum/adler32.schway": {
		Path:           "examples/checksum/adler32.schway",
		Check:          phase28Class{},
		Build:          phase28Class{Exit: 3, Code: "native.build_failed"},
		InterpreterRun: phase28Class{Exit: 3, Code: "tool.run_failed"},
		NativeRun:      phase28Class{Exit: 3, Code: "native.tool_failure"},
		EntryShape:     "single-root",
		Entry:          "adler32",
	},
	// 2026-10-12 (Phase 30, plan 30-09, D-27): the importing entry. The
	// byte-taking walker refuses imports (module.imports_not_loaded) so the
	// shape is unchecked; the module-aware route is pinned by
	// TestPhase30AdlerModuleAppCLI and TestPhase30AdlerModuleInterpreter.
	"examples/checksum/adler32_cli.schway": {
		Path:           "examples/checksum/adler32_cli.schway",
		Check:          phase28Class{},
		Build:          phase28Class{Exit: 3, Code: "native.build_failed"},
		InterpreterRun: phase28Class{Exit: 3, Code: "tool.run_failed"},
		NativeRun:      phase28Class{Exit: 3, Code: "native.tool_failure"},
		EntryShape:     "unchecked",
	},
	// 2026-10-12 (Phase 31, plan 31-01): the bounded file checksum entry. The
	// byte-taking walker refuses imports, so every plain engine class matches
	// adler32_cli; the file route is pinned by TestPhase31ChecksumPublicRoute.
	"examples/checksum/checksum.schway": {
		Path:           "examples/checksum/checksum.schway",
		Check:          phase28Class{},
		Build:          phase28Class{Exit: 3, Code: "native.build_failed"},
		InterpreterRun: phase28Class{Exit: 3, Code: "tool.run_failed"},
		NativeRun:      phase28Class{Exit: 3, Code: "native.tool_failure"},
		EntryShape:     "unchecked",
	},
	// 2026-10-12 (Phase 31, plan 31-07): the second consumer of the adler32
	// library; plain engine classes match checksum, the file route is pinned
	// by TestPhase31AdlerLowPublicRoute.
	"examples/checksum/adler_low.schway": {
		Path:           "examples/checksum/adler_low.schway",
		Check:          phase28Class{},
		Build:          phase28Class{Exit: 3, Code: "native.build_failed"},
		InterpreterRun: phase28Class{Exit: 3, Code: "tool.run_failed"},
		NativeRun:      phase28Class{Exit: 3, Code: "native.tool_failure"},
		EntryShape:     "unchecked",
	},
}

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
	for _, problem := range phase28Compare(golden, observed) {
		t.Error(problem)
	}
}

// phase28RetiredEntries lists frozen-golden paths that may be absent because
// the source was deliberately retired; observing one again is a failure.
var phase28RetiredEntries = map[string]bool{
	// 2026-10-12 (Phase 32, DX-19, D-06): retired path; may be absent; a failure if observed again. The golden stays frozen.
	"examples/checksum.schway": true,
}

// phase28Compare reports every difference between the golden and the observed
// entries. An empty result means the recorded classes hold.
func phase28Compare(golden phase28Manifest, observed []phase28ManifestEntry) []string {
	var problems []string
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
		case phase28RetiredEntries[got.Path]:
			problems = append(problems, fmt.Sprintf("%s: retired entry observed again", got.Path))
		case !inGolden && !isAdded:
			problems = append(problems, fmt.Sprintf("%s: absent from the golden and not in phase28NewlyAddedEntries: %+v", got.Path, got))
		case !inGolden:
			if !reflect.DeepEqual(got, added) {
				problems = append(problems, fmt.Sprintf("%s: newly added entry differs from allow-list: got %+v want %+v", got.Path, got, added))
			}
		case isChanged:
			if reflect.DeepEqual(changed, want) {
				problems = append(problems, fmt.Sprintf("%s: stale phase28ClassChanges entry equals the golden", got.Path))
			}
			if !reflect.DeepEqual(got, changed) {
				problems = append(problems, fmt.Sprintf("%s: observed differs from phase28ClassChanges: got %+v want %+v", got.Path, got, changed))
			}
		default:
			if !reflect.DeepEqual(got, want) {
				problems = append(problems, fmt.Sprintf("%s: class changed without allow-list: got %+v want %+v", got.Path, got, want))
			}
		}
	}
	for path := range goldenByPath {
		if phase28RetiredEntries[path] {
			continue
		}
		if _, ok := observedByPath[path]; !ok {
			problems = append(problems, fmt.Sprintf("%s: in the golden at %s but no longer observed", path, golden.Revision))
		}
	}
	sort.Strings(problems)
	return problems
}
