package reduce

import (
	"fmt"
	"sort"

	"github.com/szTheory/schway/internal/compiler/core"
)

// seedEntryInvalidError is WR-01's fail-closed refusal for a multi-function
// Seed whose EntryFunctionID names nothing in the seed's own program.
// Mirrors callgraph.entryAmbiguousError's shape deliberately -- an Error()
// string, a Code() method, and an accessor exposing the refusal's own
// witness data -- so a caller can dispatch on either refusal the same way.
// declared is stored pre-sorted (lexicographically by function ID) so
// Error()'s formatted message is byte-stable regardless of the order the
// caller's own seed declared its functions in, matching
// entryAmbiguousError's own byte-stability guarantee.
type seedEntryInvalidError struct {
	supplied string
	declared []string
}

func (e *seedEntryInvalidError) Error() string {
	return fmt.Sprintf(
		"reduce.seed_entry_invalid: multi-function seed requires Seed.EntryFunctionID to name one of its %d declared function(s), got %q, declared %v",
		len(e.declared), e.supplied, e.declared,
	)
}

// Code reports core.SeedEntryInvalid, WR-01's stable identity for this
// refusal.
func (e *seedEntryInvalidError) Code() string { return core.SeedEntryInvalid }

// Supplied reports the EntryFunctionID the caller actually set (the empty
// string when the caller left Seed.EntryFunctionID at its zero value --
// the exact hazard WR-01 names).
func (e *seedEntryInvalidError) Supplied() string { return e.supplied }

// Declared reports the seed program's own declared function IDs, sorted
// lexicographically.
func (e *seedEntryInvalidError) Declared() []string {
	return append([]string(nil), e.declared...)
}

// SeedEntryInvalidError type-asserts err as reduce's own seed-entry
// refusal, mirroring the errors.As-style accessor convention
// callgraph.EntryAmbiguousError and session.EngineMismatch already use.
func SeedEntryInvalidError(err error) (*seedEntryInvalidError, bool) {
	e, ok := err.(*seedEntryInvalidError)
	return e, ok
}

// Validate is Seed's fail-closed admission check (WR-01), applied by Reduce
// before any move runs.
//
// A MULTI-function seed must carry an EntryFunctionID matching one of
// Seed.Program.Functions[i].ID. This is the one fact dropOrphanFunction
// (reduce.go) consults to decide which function it must never delete, and
// a program's genuine entry point is by construction in-degree-zero --
// nothing else in the program calls it. So an empty or stale
// EntryFunctionID does not degrade gracefully: no function matches the
// exemption, the real entry looks exactly like an orphan, and reduction
// silently deletes the very function the seed is about rather than
// refusing. Refusing here is this codebase's own stated discipline for
// precisely this shape of ambiguity -- callgraph.EntryFunction refuses
// rather than falling back to Functions[0] for the same reason, and
// cache.InputsFor's doc comment states the general rule: ambiguity always
// resolves to run it, never to skip it, never guess.
//
// A SINGLE-function seed is exempt, and deliberately so: dropOrphanFunction
// returns early on len(p.Functions) <= 1, so the entry fact is unreachable
// for such a seed and no deletion hazard exists. This keeps every
// pre-Phase-11 single-function caller (and every committed single-function
// golden) working unchanged -- the guard adds a refusal exactly where the
// hazard is, and nowhere else.
//
// A zero-function seed is likewise exempt for the same structural reason;
// there is no function for the move to delete.
func (s Seed) Validate() error {
	if len(s.Program.Functions) <= 1 {
		return nil
	}
	declared := make([]string, 0, len(s.Program.Functions))
	found := false
	for _, fn := range s.Program.Functions {
		declared = append(declared, fn.ID)
		if fn.ID == s.EntryFunctionID {
			found = true
		}
	}
	if found {
		return nil
	}
	sort.Strings(declared)
	return &seedEntryInvalidError{supplied: s.EntryFunctionID, declared: declared}
}
