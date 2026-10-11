package callgraph

import (
	"fmt"

	"github.com/szTheory/schway/internal/compiler/core"
)

// PruneToReachable returns program restricted to the functions reachable from
// entryID (Phase 28, D-04/D-05). It runs only after the checker and every
// admission step have seen the full program (D-18), so dead code is validated
// before it is dropped. A program whose functions are all reachable is
// returned unchanged (deep-equal, so emitted bytes cannot drift); otherwise
// survivors keep their original declaration order. Traversal is an iterative
// queue, never recursion, like the rest of this package.
func PruneToReachable(program core.Program, entryID string) (core.Program, error) {
	adjacency, _, err := buildAdjacency(program)
	if err != nil {
		return core.Program{}, err
	}
	if _, ok := adjacency[entryID]; !ok {
		return core.Program{}, fmt.Errorf("callgraph.prune_entry_unknown: entry %q is not a declared function", entryID)
	}

	reachable := map[string]bool{entryID: true}
	queue := []string{entryID}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, callee := range adjacency[current] {
			if reachable[callee] {
				continue
			}
			reachable[callee] = true
			queue = append(queue, callee)
		}
	}
	if len(reachable) == len(adjacency) {
		return program, nil
	}

	pruned := program
	pruned.Functions = make([]core.Function, 0, len(reachable))
	for _, function := range program.Functions {
		if reachable[function.ID] {
			pruned.Functions = append(pruned.Functions, function)
		}
	}
	return pruned, nil
}
