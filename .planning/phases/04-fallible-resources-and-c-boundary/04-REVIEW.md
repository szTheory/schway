---
phase: 04-fallible-resources-and-c-boundary
reviewed: 2026-09-05T00:00:00Z
depth: deep
files_reviewed: 24
files_reviewed_list:
  - internal/compiler/ast/ast.go
  - internal/compiler/cgen/cgen.go
  - internal/compiler/check/check.go
  - internal/compiler/core/core.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/evidence/evidence.go
  - internal/compiler/execution/execution.go
  - internal/compiler/interp/interp.go
  - internal/compiler/native/foreign_nonlocal.go
  - internal/compiler/native/foreign_resource.go
  - internal/compiler/native/native.go
  - internal/compiler/native/symbols.go
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/session/session.go
  - internal/compiler/syntax/format.go
  - internal/compiler/syntax/lexer.go
  - internal/compiler/syntax/parser.go
  - internal/compiler/syntax/token.go
  - native/lang_foreign_nonlocal.c
  - native/lang_foreign_resource.c
  - native/lang_foreign_resource_private.h
  - scripts/verify-phase4.sh
  - testdata/phase4/foreign_layout_mismatch.golden.c
findings:
  critical: 2
  warning: 3
  info: 2
  total: 7
status: issues_found
---

# Phase 4: Code Review Report

**Reviewed:** 2026-09-05
**Depth:** deep
**Files Reviewed:** 24 (production Go across check/corevalidate/interp/cgen/native/originvalidate/pathoracle/session/evidence/core, the two frozen foreign C translation units, and the phase gate script)
**Status:** issues_found

## Summary

Phase 4 is unusually well-documented and self-critical (29 locked context decisions, per-plan revert-and-fail demonstrations, an honest debt register). Most of the areas flagged as highest-risk in the review brief — the six `OperationKind` dispatch sites, the `setjmp`/`longjmp` landing pad's avoidance of non-`static` state, exactly-once release on the return/typed-failure paths, and the D-04-29 terminator-walk widening — are implemented carefully and match their own documentation.

Two real gaps survived the phase's own layered verification, both in the "independent defense-in-depth" and "the control must not pass vacuously" categories the review brief calls out explicitly:

1. `corevalidate`'s independent release-order rederivation (`checkReleaseOrder`) silently **skips** its own check for any terminal block whose incoming-edge count is not exactly 1, rather than refusing such a shape. Nothing else in `corevalidate` enforces that a return/return-shaped terminal block has exactly one incoming edge, so this is a genuine soundness hole in the validator that is supposed to be independent of `check`'s own bookkeeping (D-04-07/D-04-12).
2. `control:foreign.no_unproven_attributes` (D-04-13's "cgen emits zero optimizer-visible attributes... scanning all emitted C") never scans the output of `cgen.EmitForeignHeader` or `cgen.EmitForeignConformance` — two of the three inspectable layers D-04-12 introduces — in either the production gate lane or any test. A banned attribute injected into the generated header (the artifact a human reviewer is most likely to actually read) would go completely undetected.

Three further warnings and two info items round out the review; none are debt already recorded in `04-DEBT.md`.

## Critical Issues

### CR-01: `corevalidate.checkReleaseOrder` silently skips validation for any terminal block that doesn't have exactly one incoming edge

**File:** `internal/compiler/corevalidate/corevalidate.go:1240-1245` (the guard is at line 1242)
**Issue:**

```go
for _, block := range linear.Blocks {
    ...
    incoming := edgesByTo[block.ID]
    if len(incoming) != 1 {
        continue   // <-- release-order rederivation is simply skipped
    }
    expected := rederive(incoming[0])
    ...
}
```

`checkReleaseOrder` is documented (and, per D-04-07/D-04-12, designed) to be an *independent* rederivation of the reverse-order release sequence, re-walking the block/edge graph without trusting anything `check.go` wrote. But when a terminal block (one ending in `OpFail` or `OpReturn`) has zero or more-than-one incoming edges, the function does not refuse the shape — it just `continue`s past that block, performing **no release-order check at all** for it.

Nothing else in `corevalidate` closes this gap for a return-terminated block: `blocksAndEdges` (corevalidate.go:411-449) enforces "every incoming edge into a fail-terminated block has Pattern `"err"`, and at least one exists" **only for `OpFail`-terminated blocks**; an `OpReturn`-terminated (or `OpFail`-terminated, for that matter) block with, say, two incoming "ok" edges from two different call chains is never rejected as malformed, and `checkReleaseOrder`'s own `len(incoming) != 1` guard then silently declines to check its release order.

Concretely: a core artifact whose terminal success block is a merge point (reachable from two different acquisition chains with different completed-acquisition sets) would have its release order — the exact guarantee RES-01 exists to provide — **entirely unchecked** by the one control that is supposed to independently prove it, while `control:kind.exhaustive_dispatch`/`replayBlocks`' generic per-operation checks (which only confirm a release names *some* real, already-produced place) would still pass. This is precisely the "a control that could pass vacuously" failure mode the project's own comments elsewhere are careful to guard against (see the `core.fail_reached_without_err_edge` comment at corevalidate.go:418-424, which explicitly defends against "a hand-corrupted core.Program" doing the analogous thing for `OpFail`).

`check.go`'s own emission never happens to produce a merge-point terminal block today, so no currently-shipped fixture exercises this path — but that is exactly why an *independent* validator exists: to catch a shape the producer doesn't happen to generate, not only to differential-test against it.

**Fix:** Treat an unexpected incoming-edge count on a terminal block as a hard refusal, not a skip, e.g.:

```go
incoming := edgesByTo[block.ID]
if !v.check(len(incoming) == 1, "core.release_order_indeterminate", block.ID) {
    return false
}
expected := rederive(incoming[0])
```

and add a corresponding structural check (a peer of the existing fail-edge check in `blocksAndEdges`) that every `OpReturn`/`OpFail`-terminated block has exactly one incoming edge, independent of `checkReleaseOrder` itself.

## Warnings

### WR-01: `control:foreign.no_unproven_attributes` never scans two of the three generated inspectable layers

**File:** `internal/compiler/session/session.go:2178-2212` (the scan call is at line 2204); compare `internal/compiler/cgen/cgen.go:1270` (`EmitForeignHeader`) and `internal/compiler/cgen/cgen.go:1327` (`EmitForeignConformance`)
**Issue:** D-04-13 requires the zero-attribute control to be "scanning all emitted C and the manifest's `emitted_attributes`". In practice, both the production gate lane and every test that exercises `cgen.ScanForBannedAttributes` only ever pass it the output of `cgen.Emit()` (the compiled program) and `cgen.EmitForeignManifest()` (the JSON sidecar):

```go
tracerCSource, _ := cgen.Emit(positiveChecked.Program)
tracerManifest, _ := cgen.EmitForeignManifest(positiveChecked.Program)
releaseCSource, _ := cgen.Emit(releaseChecked.Program)
releaseManifest, _ := cgen.EmitForeignManifest(releaseChecked.Program)
if len(cgen.ScanForBannedAttributes(tracerCSource, tracerManifest, releaseCSource, releaseManifest)) != 0 { ... }
```

`cgen.EmitForeignHeader` (the generated `_LANG_`-namespaced header with the obligation comment block — one of D-04-12's three named inspectable layers, and the one a human is most likely to actually read) and `cgen.EmitForeignConformance` (which literally embeds `EmitForeignHeader`'s text via `out.WriteString(header)` at cgen.go:1347) are never fed to `ScanForBannedAttributes` anywhere in the codebase — not in `session.go`'s production lane, not in `cgen_test.go`, not in `session_test.go`. `grep -rn "ScanForBannedAttributes"` across the repo turns up only the three call sites in `session_test.go` (lines 1337, 1364, 1371), none of which touch `EmitForeignHeader`/`EmitForeignConformance` output.

The 04-03-SUMMARY's own revert-and-fail demo ("injected a banned token into `EmitForeignHeader`'s generated extern declaration line") appears to have worked only because `EmitForeignHeader`'s extern-declaration `Fprintf` (cgen.go:1308) is byte-for-byte the same format string as `emitLinearForeign`'s own extern declaration (cgen.go:373) — a blind text mutation matching that pattern would touch both call sites at once, incidentally making the *compiled-program* scan (which the test actually checks) go red. That coincidence masks the fact that a mutation isolated to `EmitForeignHeader` alone (or to `EmitForeignConformance`'s own added `_Static_assert` text) would not be caught by any control or test in the suite.

**Fix:** Add `cgen.EmitForeignHeader(...)` and `cgen.EmitForeignConformance(...)` output to both the `lane:foreign-no-unproven-attributes` scan in `session.go` and `TestNoUnprovenAttributesEmitted`, and add a dedicated mutation-kill test that injects a banned token into `EmitForeignHeader` *only* (with the `emitLinearForeign` extern line left untouched) to prove the two artifacts are independently covered.

### WR-02: `interp.RunLinearBlockDirect` is exported production API that exists solely to serve a test

**File:** `internal/compiler/interp/interp.go:377-476`
**Issue:** The function's own doc comment states plainly: "This is not used by any production CLI path; it exists solely so the interpreter's real OpRelease/OpFail event-emission logic can be compared, genuinely executed, against the real native build's genuinely executed same block." It is nonetheless an exported (capitalized), fully public function in `interp.go` proper, not in a `_test.go` file or an internal test-support package, so it ships in the compiler's real API surface and is discoverable/callable by any importer of `interp`, with no compiler-enforced guarantee that it stays test-only.
**Fix:** Move `RunLinearBlockDirect` (and `foreignFailureLiteralDirect`, which exists only to back it) into a `testonly`-suffixed file guarded by a build tag, or into an internal test-support package under `interp`, so the production API surface doesn't carry a function whose own comment disclaims production use.

### WR-03: `evidence.Validate` compares `ForeignDigest` in constant time while every sibling digest comparison uses plain equality

**File:** `internal/compiler/evidence/evidence.go:375` vs. lines 383-384 (and the `checks` slice at 342-356)
**Issue:** `ForeignDigest` is checked with `subtle.ConstantTimeCompare`, while `SourceDigest`, `CoreDigest`, `Schema`, `IDAlgorithm`, and every other identity field in the same function are checked with ordinary `!=`/`==`. There is no adversarial-timing threat model anywhere else in this evidence-manifest code (it is a build-provenance artifact compared locally, not a secret-bearing authentication check), so the inconsistency reads as an accidental copy from an unrelated pattern rather than a deliberate security boundary — and `subtle.ConstantTimeCompare` silently returns `0` (i.e., "not equal") whenever the two byte slices differ in length, which is fine here but is exactly the kind of subtlety worth being consistent about if it's meant to matter.
**Fix:** Either use plain string comparison for `ForeignDigest` to match every other field in this function, or document why this one field specifically needs constant-time comparison when its siblings don't.

## Info

### IN-01: `resourceLedger.emitLeakEvents`/`markLive` share their marker-string constants with `session.go` by hand-duplication, not by import

**File:** `internal/compiler/cgen/cgen.go:476-477`, `633`, `internal/compiler/session/session.go:178-181`
**Issue:** `padInstallMarker`, `padEndMarker`, `ledgerPopulateMarker`, and `releaseMarker` are each declared twice — once in `cgen.go` (the producer) and once in `session.go` (the mutation-runner consumer) — kept in sync only by comment ("duplicated, verbatim, from cgen.go's own constant"). This is a documented, deliberate pattern already used elsewhere in the phase (`nonlocalExitDefectReason` between `cgen.go` and `interp.go` is the same shape), so it is consistent with established project convention rather than a new defect — noted here only because a future rename of any one of these four literals in `cgen.go` without the matching edit in `session.go` would silently break a mutation-kill control rather than fail to compile.

### IN-02: `native/lang_foreign_nonlocal.c`'s first-call acquisition frees its block before ever reporting success

**File:** `native/lang_foreign_nonlocal.c:57-67`
**Issue:** The first call path does `block[0] = argument; free(block); result.ok = 1; result.value = argument;` — the allocated block is freed unconditionally before the function returns "success", so nothing about the real allocation actually stays live past this call; the "acquisition becomes live" fact tracked by Lang's own ledger has no corresponding real allocation on the C side once the call returns. This mirrors the same documented narrowing already called out for `lang_foreign_resource.c` (the by-value one-byte result, not an opaque handle) and is consistent with 04-01/04-05's stated scope, so it is not a new gap — flagged only so a future plan extending this fixture toward a genuine opaque handle doesn't inherit the free-before-return shape by accident.

---

_Reviewed: 2026-09-05_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: deep_
