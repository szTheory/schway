---
phase: 03-borrowed-views-and-cfg-lifetimes
reviewed: 2026-09-04T17:07:33Z
depth: standard
files_reviewed: 8
files_reviewed_list:
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/originvalidate/originvalidate_test.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_test.go
  - internal/compiler/testsupport/cli_test.go
  - scripts/verify-phase3.sh
  - testdata/phase3/public_view_multi_arm_omitted.lang
  - testdata/phase3/public_view_multi_arm_access_conflict.lang
findings:
  critical: 0
  warning: 0
  info: 1
  total: 1
status: clean
---

# Phase 03: Code Review Report (03-10, third gap-closure round)

**Reviewed:** 2026-09-04T17:07:33Z
**Depth:** standard
**Files Reviewed:** 8
**Status:** clean

## Summary

This is the 03-10 change set that closes the multi-arm origin defect: `RecomputeOrigin` used to walk backward from only the FIRST `core.OpReturn` in a function's flat `Linear.Operations`, so a match-bodied function whose non-first arm returned a borrow-derived place could export with no `public_origin` at all. The fix introduces `RecomputeOriginPerReturn` as the sole backward-walk site, walking every `core.OpReturn` using one shared `TargetID -> SourceOperation` map, and makes `RecomputeOrigin` a conservative combiner over the per-return results (owned-if-none-derived, agreed-access-if-all-agree, `AccessConflicting` sentinel if they disagree). A declared-access domain check in `ValidatePublished` refuses any declared `Access` outside `{"shared","exclusive"}` before ever comparing against a recomputed answer, closing the path by which the sentinel could round-trip through a declaration.

I traced the combination law against all the edge cases called out in the review brief (derived-empty, all-agree, disagree, one-arm-owned-one-derived, broken/cyclic chains, zero-`OpReturn` bodies) and found the logic correct in every case. I also traced the shared-`sourceOf`-map safety claim against `check.go`'s actual arm lowering (`checkBranch`/`analyzeArmBody`, outside the required-reading set but load-bearing for this claim): place/operation IDs are minted from a single monotonically increasing `nextIndex` counter that is never reset between arms, so `TargetID` collisions across arms are structurally impossible regardless of what identifier names source-level bindings reuse across arms — the doc comment's claim holds, not merely by convention. `walkReturnOrigin`'s `visited` map is allocated fresh per call (per return), so no per-arm walk can contaminate a sibling arm's cycle state either.

I confirmed the sentinel (`AccessConflicting`) has no path to a written interface summary: `checkBranch` (the only producer of match-bodied `core.Function` values) never sets `PublicOrigin` at all, so a match function can never carry a declared sentinel from honest source; the domain check in `ValidatePublished` additionally refuses any declared access outside the two legal values before the declared-vs-recomputed comparison runs, so even a hand-mutated summary declaring the sentinel is caught (proven by `TestDeclaredConflictingAccessIsRefused`, which specifically targets the one fixture whose own recomputed answer IS the sentinel — the case that would slip through if the domain check ran after the comparison instead of before). `BuildInterface` is only ever called after `ValidatePublished` succeeds (`InterfaceExportCommandFile` in session.go), so no code path can package a body carrying the sentinel into a shipped summary.

The two new fixtures are structurally distinct from each other (one owned-then-borrow arm pair triggering `core.origin_omitted`; one shared-then-exclusive arm pair triggering the `AccessConflicting` sentinel and, on injected declaration, `core.origin_access_mismatch`), and I confirmed `TestMultiArmOmittedOriginRejectedThroughCLI` and `TestMultiArmOmittedOriginRejected` would both fail against the pre-03-10 single-`OpReturn` code (the first arm in that fixture is owned, so the old code would report not-ok and publish cleanly).

`scripts/verify-phase3.sh`'s two new required controls are fail-closed: both are read via `readBoundedFile`, and a missing/renamed fixture returns `protocol.StatusOperational`/`verify.fixture_missing` from `verifyBorrowedCorpus`, which under the script's `set -eu` aborts the `lang --json verify testdata/phase3` invocation before the `grep -q` control checks even run — there is no silent-skip path. Both new control names are present in the required list in all three enforcement points (`session.go`'s `requiredControls`, `session_test.go`'s `TestVerifyPhase3ControlsAndWork`, and `verify-phase3.sh`'s `for control in` loop), and `TestPhase3VerifierScriptContract` pins the script's literal text so the two new `grep` lines cannot be silently dropped from the script without failing that test. The escape list (`escape:coordinated-frontend-summary-lie`, `escape:coordinated-source-core-lie`) is untouched by this diff — grepped every reference and confirmed neither constant nor `ExpectedEscapes()` was modified.

Backward compatibility: no `lang.core` schema version bump, no new core field (this round adds no field to `core.Function`/`core.PublicOrigin` — `AccessConflicting` is a package-level Go string constant, never serialized as a schema addition), and `go build`/`go vet` are clean across the changed packages. Existing single-return fixtures and tests are unaffected — `RecomputeOrigin`'s cases 1 and 2 are documented and tested (via `TestPublishedOriginConsistentWithEveryReturn`, which loops over every `testdata/phase3/*.lang` fixture, not just the two new ones) as byte-identical to the pre-03-10 single-return behavior.

I did not find a bug in this round. The last two rounds each shipped believing they were done and were not; I looked hard for the same failure mode here (a second, still-uncovered multi-valued iteration, e.g. a `sourceOf` map keyed in a way that could alias across arms, or a comparison ordering that could let the sentinel slip through) and did not find one.

## Info

### IN-01: Shared-map safety proof lives outside required_reading, undocumented as a cross-file dependency

**File:** `internal/compiler/originvalidate/originvalidate.go:88-93` (doc comment on `RecomputeOriginPerReturn`)
**Issue:** The comment's central safety claim — that reusing one flat `sourceOf` map across all arms cannot let one arm's walk cross into a sibling arm's operations — depends entirely on `check.go`'s arm lowering (`checkBranch`'s `nextIndex` counter, `analyzeArmBody`'s `startIndex`-offset IDs) minting globally unique place/operation IDs across arms. That invariant is real (I traced it) but lives in a different package this file explicitly must never import, and nothing in `originvalidate` or its tests would fail loudly if a future `check.go` change broke it in a way that produced a colliding `TargetID` across two arms (the map write would just silently overwrite the earlier arm's operation with the later one's, corrupting one arm's walk with a splice from the other). `TestPublishedOriginConsistentWithEveryReturn`'s cross-check only re-derives the same combination from the same `RecomputeOriginPerReturn` output, so it cannot catch a `check.go`-side ID-collision regression either — it would just be consistently wrong on both sides of the comparison.
**Fix:** Not a defect to fix in this round (the invariant currently holds and is out of this package's control by design), but worth a light defensive addition: `RecomputeOriginPerReturn` could assert (or a dedicated test in `originvalidate_test.go` could assert) that every non-return operation's `TargetID` is unique across the whole function before building `sourceOf`, and treat a violation as a hard error/problem rather than a silent overwrite. That converts a currently-implicit cross-package invariant into a locally-enforced one, consistent with this package's stated source-blind, trust-nothing posture.

---

_Reviewed: 2026-09-04T17:07:33Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
