# Spike Wrap-Up Summary

**Date:** 2026-09-17
**Spikes processed:** 6 (001–006, all of them; none were previously wrapped)
**Idea:** `ownership-kernel`
**Feature areas:** Ownership kernel semantics · Loan liveness dataflow · Public
interface origins and abilities · Evidence and the trust boundary · Native
lowering and the FFI contract · Differential verification harness
**Skill output:** `./.claude/skills/spike-findings-ai-lang/`

## Processed Spikes

| # | Name | Type | Verdict | Feature Area |
|---|------|------|---------|--------------|
| 001 | ownership-kernel-workbench | standard | VALIDATED | Ownership kernel semantics |
| 002 | cfg-edge-last-use | standard | VALIDATED | Loan liveness dataflow |
| 003 | public-origins-generic-abilities | comparison | VALIDATED | Public interface origins and abilities |
| 004 | independent-certificate-checker | comparison | PARTIAL | Evidence and the trust boundary |
| 005 | native-ffi-provenance-cleanup | standard | PARTIAL | Native lowering and the FFI contract |
| 006 | interprocedural-liveness-cost-scaling | comparison | VALIDATED | Loan liveness dataflow |

All six spikes contribute to the cross-cutting
`differential-verification-harness` reference.

## Key Findings

**Semantics.** A finite ownership lattice reaches branch joins and zero-or-more
loop fixed points without solver machinery. Loan identities must be unique for
the whole program — reuse, even after a loan ends, makes evidence ambiguous.
Release evidence records the owning lexical scope, not a nesting depth. A join
giving one static loan identity different owners or modes is rejected
conservatively.

**A loan's end is not always attached to a source expression.** At divergent
control flow and at loop exit it is an *edge* fact. The typed IR and machine
diagnostics therefore need stable edge identities even though ordinary source
contains no loan-ending syntax. Exhaustive path expansion belongs in the bounded
oracle; the production-shaped analyzer is a monotone finite-set worklist that
converged in 1,501 transfer evaluations on a 1,501-block graph.

**The interprocedural memo table is the mechanism, not an optimization.**
Memoized summary derivation is linear in program operation count on all six call-
graph shapes at ~5 work units per operation across a 100x size range. The
deliberately charitable unmemoized arm costs 76x more at 512 functions and the
multiplier doubles with every doubling of program size. Reverse postorder over
the already-proven-acyclic call graph needs no call-graph-level fixpoint at all —
the acyclicity guarantee is what is load-bearing, not the ordering choice.
Two preconditions came with it: summary derivation must consume operations in
**program order** (reversed, the same derivation is 192x worse at k=512 and
quadratic in body length), and a **cross-run** cache is a separate claim — one
leaf edit invalidates 92% of a parser-shaped cache worst case, 100% on a chain.

**Public summaries.** Value-parameter origin paths plus tagged alternatives carry
enough for a body-blind consumer to check borrowed and generic results, with
shared/exclusive access as a separate fact and a fresh quantified origin
introduced only at genuine callback lending boundaries. `copy`, `drop`, `share`,
`send`, `escape` stay independent through wrappers and type erasure — erasing a
data representation is not permission to erase lifetime or authority facts.
Production warning from a harness defect: exported summaries need immutable,
content-owned representations, not merely different top-level structs.

**Evidence.** Bind the certificate to a canonical typed-core digest that
canonicalizes unordered declarations but preserves semantic event order. Compact
recomputation (242 bytes in the fixture) is both smaller and faster than full
replay (~1.0 MB at scale) and loses no soundness, because the checker must
recompute the transition relation either way. The validator belongs at trust
crossings — package import, cache acceptance, CI/release — not after every edit.
Its claim is narrow on purpose: a coordinated false typed-core statement with a
matching summary, digest, and certificate escapes by design, and the mutation
matrix says so explicitly.

**Native.** A frontend that emits `noalias`/`restrict` without proof creates
wrong code, not a conservative performance loss — the false contract is latent at
`-O0` and returns a different value at `-O3`. Linker success is not ABI evidence:
reordering fields in the consumer translation unit only still links and corrupts
the result. `longjmp` skips cleanup entirely with no sanitizer report. Sanitizer,
unoptimized, optimized, and semantic-oracle lanes each catch a class the others
miss.

**Methodology (the most transferable result).** Three of the most valuable
findings across the series were defects in the *harness*: a shallow-copied
interface boundary that let fault injection mutate the oracle (003), a reused
loan identifier that made evidence ambiguous (001), and an interprocedural fact
placed in the backward transfer instead of a program-order pre-pass (006). Each
surfaced because implementations were written independently, identities were
stable, ordered traces were compared rather than verdicts, and every harness
carried an armed fault-injection path plus at least one explicitly expected
escape.

## Stopping point

Spike 005 records the stopping point for disposable Go/C ownership workbenches.
Spikes 001–005 define enough versioned fixtures and failure modes to begin the
real lossless frontend and portable typed core; the native harness should become
a consumer of that core rather than grow a parallel pseudo-language. Spike 006 is
a pre-phase gate (S-006) on the *planning* of Phase 08 and answers it: design the
memoized summary cache in from the start.
