package session_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/callgraph"
	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// phase28PrunedChangedPaths lists the only accepted sources whose emitted C may
// differ from the frozen phase272/phase273 goldens once programs are pruned to
// the functions reachable from the entry (D-05, D-23). Each observed value
// must differ from the golden, so the list cannot go stale silently.
var phase28PrunedChangedPaths = map[string]string{
	"testdata/phase11/multi_function_gate_corpus.schway":        "dead function dropped before emission",
	"testdata/phase11/multi_function_unreachable.schway":        "orphan dropped before emission",
	"testdata/phase13/heldout_shared_callee_twin_alpha.schway":  "uncalled twin dropped before emission",
	"testdata/phase13/heldout_shared_callee_twin_mirror.schway": "uncalled twin dropped before emission",
	"testdata/phase11/multi_function_zero_call.schway":          "was an entry-ambiguity refusal; main precedence now emits",
	"examples/phase28/unused_function.schway":                   "new Phase 28 example, absent from the frozen goldens",
}

// phase28CheckedCorpus returns every session.Check-accepted .schway file under
// examples/ and testdata/ keyed by repository-relative path.
func phase28CheckedCorpus(t *testing.T) map[string]core.Program {
	t.Helper()
	projectRoot := testsupport.ProjectPath(".")
	corpus := map[string]core.Program{}
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
			corpus[filepath.ToSlash(relative)] = checked.Program
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", root, err)
		}
	}
	return corpus
}

func phase28SortedPaths(corpus map[string]core.Program) []string {
	paths := make([]string, 0, len(corpus))
	for path := range corpus {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

func phase28Roots(program core.Program) []string {
	called := map[string]bool{}
	for _, function := range program.Functions {
		if function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind == core.OpCall {
				called[operation.CalleeID] = true
			}
		}
	}
	var roots []string
	for _, function := range program.Functions {
		if !called[function.ID] {
			roots = append(roots, function.ID)
		}
	}
	return roots
}

// TestPhase28PruneCorpusProperties proves D-05/D-19 over the whole accepted
// corpus: a resolvable program prunes to exactly one root, re-resolves to the
// same entry, keeps declaration order, and is deep-equal when nothing is dead.
func TestPhase28PruneCorpusProperties(t *testing.T) {
	corpus := phase28CheckedCorpus(t)
	resolved, pruneChanged := 0, 0
	for _, path := range phase28SortedPaths(corpus) {
		program := corpus[path]
		entry, err := callgraph.EntryFunction(program)
		if err != nil {
			continue
		}
		resolved++
		pruned, err := callgraph.PruneToReachable(program, entry.ID)
		if err != nil {
			t.Errorf("%s: prune: %v", path, err)
			continue
		}
		if roots := phase28Roots(pruned); len(roots) != 1 || roots[0] != entry.ID {
			t.Errorf("%s: pruned roots = %v, want exactly [%s]", path, roots, entry.ID)
		}
		again, err := callgraph.EntryFunction(pruned)
		if err != nil || again.ID != entry.ID {
			t.Errorf("%s: re-resolved entry %q err=%v, want %q", path, again.ID, err, entry.ID)
		}
		position := 0
		for _, function := range pruned.Functions {
			found := false
			for ; position < len(program.Functions); position++ {
				if program.Functions[position].ID == function.ID {
					found = true
					position++
					break
				}
			}
			if !found {
				t.Errorf("%s: survivor %s is out of declaration order", path, function.ID)
			}
		}
		if len(pruned.Functions) == len(program.Functions) {
			if !reflect.DeepEqual(program, pruned) {
				t.Errorf("%s: nothing dead but pruned program is not deep-equal", path)
			}
		} else {
			pruneChanged++
		}
	}
	if resolved < 80 {
		t.Fatalf("only %d corpus programs resolved an entry; the walk went inert", resolved)
	}
	if pruneChanged == 0 {
		t.Fatal("no corpus program had a dead function; the negative side of the property is vacuous")
	}
}

// TestPhase28PrunedEmitMatchesGolden proves D-05/D-23: emitting the pruned
// program yields the frozen phase273 conformance-C and phase272 application-C
// results byte for byte, except the allow-listed dead-function files.
func TestPhase28PrunedEmitMatchesGolden(t *testing.T) {
	corpus := phase28CheckedCorpus(t)

	var conformance phase273CorpusGolden
	raw, err := os.ReadFile(phase273GoldenPath())
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &conformance); err != nil {
		t.Fatal(err)
	}
	var application phase272CorpusGolden
	raw, err = os.ReadFile(phase272GoldenPath())
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &application); err != nil {
		t.Fatal(err)
	}
	conformanceByPath := map[string]phase273CorpusEntry{}
	for _, entry := range conformance.Entries {
		conformanceByPath[entry.Path] = entry
	}
	applicationByPath := map[string]phase272CorpusEntry{}
	for _, entry := range application.Entries {
		applicationByPath[entry.Path] = entry
	}

	differs := map[string]bool{}
	compared := 0
	for _, path := range phase28SortedPaths(corpus) {
		program := corpus[path]
		emitted := program
		if entry, entryErr := callgraph.EntryFunction(program); entryErr == nil {
			pruned, pruneErr := callgraph.PruneToReachable(program, entry.ID)
			if pruneErr != nil {
				t.Errorf("%s: prune: %v", path, pruneErr)
				continue
			}
			emitted = pruned
		}

		conformanceC, conformanceErr := session.Phase16ControlNativeC(emitted, "phase273-identity")
		if want, ok := conformanceByPath[path]; ok {
			got := phase273CorpusEntry{Path: path}
			if conformanceErr != nil {
				got.EmitNativeError = conformanceErr.Error()
			} else if path == "examples/sum_to_n.schway" {
				source, sourceErr := os.ReadFile(testsupport.ProjectPath(filepath.FromSlash(path)))
				if sourceErr != nil {
					t.Errorf("%s: read current source for Phase371 binding: %v", path, sourceErr)
					differs[path] = true
				} else {
					coreBytes, marshalErr := json.Marshal(emitted)
					if marshalErr != nil {
						t.Errorf("%s: marshal pruned core for Phase371 binding: %v", path, marshalErr)
						differs[path] = true
					} else {
						var compatibilityProblems []string
						oldC, phase371Problems := phase371OlderCompare(
							"phase34",
							path,
							"emit_native",
							phase371OlderDigest(source),
							phase371OlderDigest(coreBytes),
							nil,
							[]byte(conformanceC),
						)
						compatibilityProblems = append(compatibilityProblems, phase371Problems...)
						if len(oldC) == 0 {
							compatibilityProblems = append(compatibilityProblems, "Phase371 compatibility projection did not bind the frozen old C")
						} else if err := phase273PriorNamedExitGoldenHash(oldC, want.EmitNativeSHA256); err != nil {
							compatibilityProblems = append(compatibilityProblems, err.Error())
						}
						if len(compatibilityProblems) != 0 {
							t.Errorf("%s: Phase371 older-comparator binding failed: %s", path, strings.Join(compatibilityProblems, "; "))
							differs[path] = true
						} else {
							// The frozen comparator stores only this prior projection's digest.
							// Current C was checked byte-for-byte against the bound Phase371
							// projection above, and census C was checked against this digest.
							got.EmitNativeSHA256 = want.EmitNativeSHA256
						}
					}
				}
			} else {
				sum := sha256.Sum256([]byte(conformanceC))
				got.EmitNativeSHA256 = hex.EncodeToString(sum[:])
			}
			if got.EmitNativeSHA256 != want.EmitNativeSHA256 || got.EmitNativeError != want.EmitNativeError {
				differs[path] = true
			}
			compared++
		}

		applicationC, applicationErr := cgen.EmitApplication(emitted)
		if want, ok := applicationByPath[path]; ok {
			got := phase272CorpusEntry{Path: path}
			if applicationErr != nil {
				got.EmitApplicationError = applicationErr.Error()
			} else {
				sum := sha256.Sum256([]byte(applicationC))
				got.EmitApplicationSHA256 = hex.EncodeToString(sum[:])
			}
			if got.EmitApplicationSHA256 != want.EmitApplicationSHA256 || got.EmitApplicationError != want.EmitApplicationError {
				differs[path] = true
			}
			compared++
		}
	}
	if compared < 150 {
		t.Fatalf("only %d golden comparisons ran; the walk went inert", compared)
	}
	for path := range differs {
		if _, allowed := phase28PrunedChangedPaths[path]; !allowed {
			t.Errorf("%s: pruned emission differs from the frozen golden and is not allow-listed", path)
		}
	}
	for path, reason := range phase28PrunedChangedPaths {
		if _, ok := corpus[path]; !ok {
			t.Errorf("%s: allow-listed (%s) but no longer accepted", path, reason)
			continue
		}
		_, inConformance := conformanceByPath[path]
		_, inApplication := applicationByPath[path]
		if inConformance || inApplication {
			if !differs[path] {
				t.Errorf("%s: allow-listed (%s) but its pruned emission equals the golden", path, reason)
			}
		}
	}
}
