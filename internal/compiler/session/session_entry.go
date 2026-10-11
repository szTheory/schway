package session

import (
	"sort"
	"strings"

	"github.com/szTheory/schway/internal/compiler/callgraph"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
)

// entryOf is the only session call to callgraph.EntryFunction (D-04). Sites
// that need the entry of a program that has already been resolved and pruned
// (or that only project facts about it) use this; sites that feed an engine
// use resolveEntry.
func entryOf(program core.Program) (core.Function, error) {
	return callgraph.EntryFunction(program)
}

// resolveEntry is the single entry-resolution point for check, build, run and
// app verify (Phase 28, D-04). It runs after the checker and every existing
// admission step (D-18), so dead code is validated before it is dropped, and
// before any engine sees the program. It resolves the entry (the function
// named main when present, D-01), refuses ambiguity as a source diagnostic
// (core.entry_ambiguous, exit 2, D-02), and returns the program pruned to the
// functions reachable from that entry (D-05) so every engine, including the
// independent execution peer, sees exactly one in-degree-zero root (D-19).
//
// Phase 30 composition point (D-06, informational; nothing module-specific is
// built here): entry candidates come from the root module only, library
// exports are never roots, and unused exports are pruned but still checked.
func resolveEntry(program core.Program) (core.Function, core.Program, []diagnostic.Diagnostic, error) {
	entry, err := entryOf(program)
	if err != nil {
		if refusal, ok := entryRefusalDiagnostic(err); ok {
			return core.Function{}, program, []diagnostic.Diagnostic{refusal}, nil
		}
		return core.Function{}, program, nil, err
	}
	pruned, err := callgraph.PruneToReachable(program, entry.ID)
	if err != nil {
		return core.Function{}, program, nil, err
	}
	return entry, pruned, nil, nil
}

// entryRefusalDiagnostic maps an entry-resolution refusal to its source
// diagnostic. The `fn main` hint lives here, never in callgraph's Error().
func entryRefusalDiagnostic(err error) (diagnostic.Diagnostic, bool) {
	if ambiguous, ok := callgraph.EntryAmbiguousError(err); ok {
		roots := append([]string(nil), ambiguous.Roots()...)
		sort.Strings(roots)
		message := "program entry is ambiguous: in-degree-zero roots " + strings.Join(roots, ", ") + "; add `fn main` to disambiguate"
		return diagnostic.Error(ambiguous.Code(), diagnostic.Span{}, message), true
	}
	if unresolved, ok := callgraph.UnresolvedCalleeError(err); ok {
		return diagnostic.Error(core.CallCalleeUnresolved, diagnostic.Span{}, unresolved.Error()), true
	}
	return diagnostic.Diagnostic{}, false
}
