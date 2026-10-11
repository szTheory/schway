package executionpeer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/callgraph"
	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

// peerEntryInputs builds the functions and operations maps for resolveEntry
// the same way the peer's own index construction does.
func peerEntryInputs(program core.Program) (map[string]core.Function, map[string]operation) {
	functions := make(map[string]core.Function, len(program.Functions))
	operations := map[string]operation{}
	for _, function := range program.Functions {
		functions[function.ID] = function
		if function.Linear == nil {
			continue
		}
		for _, op := range function.Linear.Operations {
			operations[op.ID] = operation{functionID: function.ID, value: op}
		}
	}
	return functions, operations
}

// TestPhase28PeerEntryRefusalUnreachableAfterPrune proves D-02/D-19: the peer
// keeps its own entry refusal, but a program pruned from its callgraph entry
// has exactly one root, so that refusal cannot surface from a production path.
func TestPhase28PeerEntryRefusalUnreachableAfterPrune(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	checkedPrograms := map[string]core.Program{}
	for _, dir := range []string{"examples", "testdata"} {
		err := filepath.Walk(filepath.Join(root, dir), func(path string, info os.FileInfo, walkErr error) error {
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
			parsed := syntax.Parse(source)
			if len(parsed.Diagnostics) != 0 {
				return nil
			}
			result := check.Program(parsed.Program)
			if len(result.Diagnostics) != 0 {
				return nil
			}
			checkedPrograms[filepath.ToSlash(path)] = result.Program
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	verified := 0
	for path, program := range checkedPrograms {
		entry, err := callgraph.EntryFunction(program)
		if err != nil {
			continue
		}
		pruned, err := callgraph.PruneToReachable(program, entry.ID)
		if err != nil {
			t.Errorf("%s: prune: %v", path, err)
			continue
		}
		functions, operations := peerEntryInputs(pruned)
		got, err := resolveEntry(functions, operations)
		if err != nil || got != entry.ID {
			t.Errorf("%s: peer resolved %q err=%v after prune, want %q", path, got, err, entry.ID)
			continue
		}
		verified++
	}
	if verified < 80 {
		t.Fatalf("only %d corpus programs verified; the walk went inert", verified)
	}

	t.Run("negative-control-unpruned-zero-call", func(t *testing.T) {
		var program core.Program
		found := false
		for path, candidate := range checkedPrograms {
			if strings.HasSuffix(path, "testdata/phase11/multi_function_zero_call.schway") {
				program, found = candidate, true
			}
		}
		if !found {
			t.Fatal("multi_function_zero_call.schway is not in the checked corpus")
		}
		functions, operations := peerEntryInputs(program)
		if _, err := resolveEntry(functions, operations); err == nil {
			t.Fatal("the peer must refuse the unpruned two-root program; the property would be vacuous")
		}
	})
}
