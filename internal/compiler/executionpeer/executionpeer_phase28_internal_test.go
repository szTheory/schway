package executionpeer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

// legacyResolveEntry is a verbatim copy of the pre-Phase-32.1 resolveEntry body
// (no `main` tier, no entry-module filter). It exists only so the parity
// control below can prove it discriminates: on an unpruned program with a dead
// larger-closure root, this rule disagrees with callgraph.EntryFunction.
func legacyResolveEntry(functions map[string]core.Function, operations map[string]operation) (string, error) {
	inDegree := make(map[string]int, len(functions))
	adjacency := make(map[string][]string, len(functions))
	for id := range functions {
		inDegree[id] = 0
	}
	for _, op := range operations {
		if op.value.Kind == core.OpCall {
			adjacency[op.functionID] = append(adjacency[op.functionID], op.value.CalleeID)
			inDegree[op.value.CalleeID]++
		}
	}
	roots := make([]string, 0)
	for id, degree := range inDegree {
		if degree == 0 {
			roots = append(roots, id)
		}
	}
	sort.Strings(roots)
	if len(roots) == 1 {
		return roots[0], nil
	}
	best, count := -1, 0
	bestID := ""
	for _, root := range roots {
		visited := map[string]bool{root: true}
		queue := []string{root}
		size := 0
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			for _, callee := range adjacency[current] {
				if !visited[callee] {
					visited[callee] = true
					size++
					queue = append(queue, callee)
				}
			}
		}
		if size > best {
			best, count, bestID = size, 1, root
		} else if size == best {
			count++
		}
	}
	if count != 1 {
		return "", refusal("entry_ambiguous", -1, "", "", "", fmt.Sprintf("roots %v", roots))
	}
	return bestID, nil
}

func phase321Function(id, name, module string, callees ...string) core.Function {
	function := core.Function{
		ID: id, Name: name, Module: module, EntryPointID: id + ":point:entry", ReturnPointID: id + ":point:return",
		Parameter: core.Parameter{ID: id + ":place:0", Name: "value", Type: "Byte"},
		Linear:    &core.LinearBody{ID: id + ":linear"},
	}
	for index, callee := range callees {
		function.Linear.Operations = append(function.Linear.Operations, core.LinearOperation{
			ID: id + ":op:" + string(rune('0'+index)), Kind: core.OpCall, CalleeID: callee,
		})
	}
	return function
}

// TestPhase321UnprunedMainPrecedenceParity proves D-15/D-16/D-26: the peer
// re-derives the Phase 28 D-01 entry rule independently and agrees with
// callgraph.EntryFunction on unpruned input.
func TestPhase321UnprunedMainPrecedenceParity(t *testing.T) {
	t.Run("unpruned-main-precedence-parity", func(t *testing.T) {
		program := core.Program{Functions: []core.Function{
			phase321Function("p:fn:main", "main", "", "p:fn:shared"),
			phase321Function("p:fn:dead", "dead", "", "p:fn:shared", "p:fn:a"),
			phase321Function("p:fn:shared", "shared", ""),
			phase321Function("p:fn:a", "a", "", "p:fn:b"),
			phase321Function("p:fn:b", "b", ""),
		}}
		functions, operations := peerEntryInputs(program)
		for id, function := range functions {
			if function.Name == "" {
				t.Fatalf("peer input function %s has no Name", id)
			}
		}
		oracle, err := callgraph.EntryFunction(program)
		if err != nil || oracle.ID != "p:fn:main" {
			t.Fatalf("callgraph entry = %q err=%v, want main", oracle.ID, err)
		}
		got, err := resolveEntry(functions, operations)
		if err != nil || got != "p:fn:main" {
			t.Fatalf("peer entry = %q err=%v, want main", got, err)
		}
		legacy, err := legacyResolveEntry(functions, operations)
		if err != nil || legacy != "p:fn:dead" {
			t.Fatalf("control: legacy rule = %q err=%v, want the dead larger-closure root", legacy, err)
		}
	})
}
