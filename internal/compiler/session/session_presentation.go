package session

import (
	"github.com/szTheory/schway/internal/compiler/modload"
	"github.com/szTheory/schway/internal/compiler/protocol"
)

// presentationFromLoaded joins loader-confirmed file labels with the trees
// parsed from the same loaded byte snapshots. The entry label comes from CLI
// authority; imported labels remain relative to the entry directory.
func presentationFromLoaded(set modload.Set, entryLabel string) *protocol.PresentationContext {
	if len(set.Sources) == 0 {
		return nil
	}
	sources := make(map[string]protocol.CapturedSource, len(set.Sources))
	for _, source := range set.Sources {
		tree, ok := set.Trees[source.Module]
		if !ok {
			continue
		}
		label := source.RelPath
		if source.Module == "" {
			label = entryLabel
		}
		sources[source.Module] = protocol.CapturedSource{Label: label, Bytes: tree.Source}
	}
	return &protocol.PresentationContext{Sources: sources}
}

func commandWithCheckedPresentation(result protocol.Result, checked CheckResult) protocol.Result {
	if checked.Presentation == nil {
		return result
	}
	return protocol.WithPresentation(result, *checked.Presentation)
}
