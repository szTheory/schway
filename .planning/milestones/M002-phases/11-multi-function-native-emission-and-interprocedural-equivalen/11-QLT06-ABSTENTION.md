# QLT-06 Split: Abstention Record

This records why QLT-06 ships as a **split row** — QLT-06a Complete, QLT-06b
Complete-by-abstention — rather than a single row that would read as "we
built the closure-keyed interprocedural cache." No such cache was built,
deliberately, and this document is where that fact is written down so no
later reader mistakes abstention for construction (D-11-39, the OWN-05a/05b
precedent).

In one sentence per half: QLT-06a is mutation-killed by two independent
knowers, including `TestCalleeChangeInvalidatesCallerClosureDigest`; QLT-06b
rests on the fact that a whole-program hash strictly dominates any
call-graph-closure key in a single-unit language, so nothing interprocedural
needed to be cached, and the S-006 figures cited earlier as motivation are
only an upper bound on a model, not a measurement of this codebase's actual
digest-chaining behavior.

## QLT-06a — Complete (Phase 07)

The callee-changes-invalidates-caller regression exists in this codebase
and is mutation-killed by **two independent knowers**:

- `TestCalleeChangeInvalidatesCallerClosureDigest` in
  `internal/compiler/originvalidate/originvalidate_closure_chain_test.go`
  (around line 312) — the producer-side chain: changing a callee's
  signature (via a declared `ForeignContract`) changes the caller's
  `ClosureDigest`, chained in `callgraph.Order`'s proven-acyclic reverse
  postorder.
- `TestPeerClosureDigestEmptyCalleesMutationKilled` in
  `internal/compiler/corevalidate/corevalidate_closure_chain_mutation_test.go`
  (around line 95) — `corevalidate`'s independently re-derived peer chain
  (`chainPeerClosureDigests`), mutation-killed by forcing empty callee pairs
  under `SetClosureDigestEmptyCalleesForTest(true)` and observing the
  callee-changes-invalidates-caller property disappear, then confirming it
  is restored once the seam is disabled.

Both tests derive the SAME property — a callee's signature change must move
its caller's chained digest — from two separately-written implementations
(`originvalidate`'s producer chain and `corevalidate`'s independent peer
chain), never a shared helper re-invoked twice.

## QLT-06b — Complete-by-abstention (Phase 11), structurally gated

No interprocedural fact is marked cacheable, because **none is cached**.
This is discharged by Phase 11 plan 11-07 Task 2's structural tests in
`internal/compiler/cache/probe_test.go`, named individually:

- `TestDeclaredInputNamesStableAndNoInterproceduralImport` — asserts the
  declared-input list's first seven names stay byte-identical and in their
  original order (exact-string, order-sensitive), runs a direct `go/parser`
  `ImportsOnly` scan over `internal/compiler/cache`'s own non-test source
  files, AND a bounded/timed transitive `go list -deps` scan, both
  asserting `cache` never imports `core`, `corevalidate`, or
  `originvalidate`.
- `TestCacheDirectImportGuardCanFail` / `TestCacheTransitiveImportGuardCanFail`
  — each scan's own negative control, proving the guard can actually go
  red against a synthetic forbidden-import fixture, not merely that it has
  never yet found anything.
- `TestNoClosureDigestInCache` — the mechanical form of D-11-38's "cache
  nothing new": a source scan over `internal/compiler/cache`'s
  PRODUCTION (non-test) files asserts none mentions the interprocedural
  per-function digest field, and no declared input name looks closure- or
  call-graph-derived.
- `TestCacheKeyIsIdempotent` — computing the cache key twice for identical
  inputs yields a byte-identical `Key.ID`, and a second `Consult` on
  unchanged inputs reports reuse rather than recomputing — the property
  that makes "nothing new is cached" a stable, repeatable fact rather than
  an accident of one run.

These are two independent knowers (the direct scan and the transitive
scan) plus their own falsifiability proofs, not one assertion that nothing
was built.

## Why abstention is the correct answer, not a shortcut (D-11-38)

`ArtifactSpec.FixtureSource` already hashes the **entire `.lang` source
file** for every fixture this cache serves. Lang has one module, one file,
and no separate compilation units. A whole-program hash therefore
**strictly dominates** any call-graph-closure key: every fact a closure key
could ever distinguish, the whole-program hash already distinguishes, and
strictly more besides (it also reacts to a change nowhere near the call
graph — say, a comment, or an unrelated local variable — that a closure key
by construction could never see, but which also could never legally alter
runtime behavior in a way the differential comparator wouldn't catch
independently of the cache).

A closure key here could only ever admit MORE cache hits, each on strictly
LESS evidence than the whole-program hash already provides — a
soundness-**loosening** change that buys nothing and costs real complexity
(a new digest chain, a new invalidation path, a new opportunity for the
D-11-41-shaped hole this same plan just closed for `cgen`). QLT-06's named
failure mode — "per-unit hashes" causing stale-but-self-consistent
callers — is not present in this repo and **cannot be**, because there are
no compilation units to hash per-unit in the first place.

## Why the row is split (D-11-39)

A single row flipped Complete would read as "we built the closure-keyed
cache," which is false: no `ClosureDigest`-derived key was ever wired into
`cache.Input`, and none should be, per the dominance argument above. This
follows the OWN-05a/OWN-05b precedent (D-09-37) exactly: OWN-05 was split
rather than left as a single row flipped Complete on two of three
derivers, because a single-row overclaim is exactly the
requirement-vs-code mismatch the debt registers exist to catch. QLT-06
splits for the identical reason, on the identical structure: one half
(06a) is a real mutation-killed construction; the other half (06b) is a
correct, evidenced decision NOT to construct something, made structural
rather than left as an unenforced assertion.

## Correction to carried evidence (D-11-40)

Spike S-006's eviction figures — 100% chain / 92% worst / 43% mean — are an
upper bound on a model, not a measurement: the spike modelled every
callee edit as unconditionally moving the caller's digest. In fact,
`ClosureDigestDomainSeparator`'s preimage
(`internal/compiler/originvalidate/originvalidate.go`, around lines
525-590) chains over **signature SUMMARIES only, never bodies** — a
body-only edit that preserves the callee's declared signature does not
move the caller's `ClosureDigest` at all. This is *early cutoff* falling
out of the trust boundary the digest is built over, not a mitigation
bolted on afterward. Do not re-quote S-006's 100%/92%/43% figures as
measured eviction rates going forward; they describe a strictly worse
hypothetical than what this codebase's chaining actually does.

## The D-11-41 outcome

Plan 11-01's Q-02 pre-planning spike settled **BRANCH A: the hole
reproduces.** `TestQ02StaleCgenServesReusedArtifact`
(`internal/compiler/session/session_phase6_cache_hole_test.go`) proved the
D-11-41 stale-`cgen` cache-reuse hole end-to-end: an unchanged `.lang`
fixture's cache key stayed identical, and the cache served a stale
artifact, even though the C source a rewritten `cgen` would emit for that
exact fixture had changed.

Plan 11-07 (this plan) shipped the real fix: `cache.DeclaredInputNames()`
gains `cgen_source` as an eighth, appended entry (the original seven stay
byte-identical and ordered), computed by `CgenSourceDigest` hashing
`internal/compiler/cgen/*.go` directly via `crypto/sha256`/`os` — never by
importing `core`, `corevalidate`, or `originvalidate` — and threaded into
`phase6ArtifactSpec` so the computed `Key.ID` genuinely moves when `cgen`'s
own source changes. `probe.go`'s D-06-13 hole-comment gained a fifth,
CLOSED item; the original four remain untouched.

**What this means for cache soundness going forward:** editing `cgen` and
re-running the native-differential lane against an unchanged `.lang`
fixture can no longer silently serve a binary compiled by the old `cgen`.
The hole D-11-41 named is closed by a real, key-moving mechanism, not
merely a note that it was investigated.

## When to revisit (deferred)

A closure-keyed native cache becomes worth revisiting only if Lang gains
**separate compilation units** — at which point a whole-program hash would
stop strictly dominating a call-graph-closure key, since the whole program
would no longer be one hashable unit. No such units exist on any roadmap
at the time of this writing (see `.planning/LANGUAGE-MATURITY.md`).
