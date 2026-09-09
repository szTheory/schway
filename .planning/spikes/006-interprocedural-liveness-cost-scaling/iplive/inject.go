package iplive

// Fault seams. This is spike code, so the seams are exported methods rather
// than the unexported same-package hooks production would use -- the CLI has
// to be able to arm them, and every convention this workbench follows says a
// harness must be seen to fail before its green run means anything.

// InjectSummaryDrop arms fault seam 1: the derivation forgets that a forwarded
// parameter comes back out through the return. Callers then believe an
// argument loan dies at the call, which under-approximates liveness and drops
// real move-while-borrowed refusals.
func (m *MemoProvider) InjectSummaryDrop() { m.dropReturnsBorrow = true }

// InjectSummaryDrop arms the same seam on the unmemoized arm.
func (r *RecomputeProvider) InjectSummaryDrop() { r.dropReturnsBorrow = true }

// InjectStaleKeys arms fault seam 2: the cache is keyed by function ID alone,
// so editing a callee never invalidates its callers. This is exactly the
// stale-but-self-consistent hole that Phase 07's ClosureDigest chain exists to
// close, priced here against a cache that has to be correct to be usable.
func (m *MemoProvider) InjectStaleKeys() { m.staleKeyed = true }

// Rebind points a warm provider at an edited program without clearing its
// cache -- the "recompile after a one-function edit" situation.
func (m *MemoProvider) Rebind(p *Program) { m.byID = p.index() }

// MutateLeafToUse rewrites one function from "hands the parameter back" to
// "reads the parameter and returns owned" -- a one-function edit that changes
// every transitive caller's summary. It exists so the CLI can replay a callee
// edit against a warm cache.
func MutateLeafToUse(p *Program, leafID string) *Program {
	mutated := &Program{Roots: append([]string(nil), p.Roots...)}
	for _, fn := range p.Functions {
		if fn.ID == leafID {
			mutated.Functions = append(mutated.Functions, leafUse(leafID))
			continue
		}
		mutated.Functions = append(mutated.Functions, fn)
	}
	return mutated
}
