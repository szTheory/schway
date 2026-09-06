---
phase: 06-agent-feedback-and-performance-ratification
plan: 04
subsystem: build-cache
tags: [cache, content-addressing, clang-probe, fail-closed, dx-tooling]

# Dependency graph
requires:
  - phase: 06-agent-feedback-and-performance-ratification
    plan: "06-01"
    provides: "The coordinated-bump discipline and additive-only schema convention this plan's new, independent lang.cache-meta/0 schema follows without touching lang.command/0 or lang.verify-lane/0."
provides:
  - "internal/compiler/cache: a dependency-free leaf package implementing a local, content-addressed store for compiled/instrumented/mutant ARTIFACTS ONLY, never a verdict (D-06-06)"
  - "cache.ComputeKey/Store/Open/Path/Put/Get -- content-addressed on-disk layout at os.UserCacheDir()/lang-verify/<2-hex>/<sha256>/{artifact,meta.json}"
  - "cache.DeclaredInputNames/InputsFor/ProbeClangIdentity -- D-06-07's seven declared inputs, including a real probed Clang binary digest, not merely --version text"
  - "cache.Consult/Outcome/CacheStatus/CacheStatuses -- D-06-12's closed cache_status vocabulary (artifact_reused/artifact_recomputed/not_cacheable/unavailable) and D-06-11's fail-closed not_cacheable path"
affects: [06-05, 06-06-lane-schema-bump, phase-6-verification, dx-tooling-consumers-of-lang-cache]

# Actuals (#2632)
actuals:
  tokens: 10514
  tasks: 3
  commits: 4

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Dependency-free leaf package scaffolding (internal/compiler/reduce's own precedent): imports only stdlib, own _test.go sibling, no functional analog reused"
    - "Duplicated (not imported) evidence.runToolProbe's 64 KiB-plus-one bounded-writer + 5-second-deadline shape, per the leaf-package's no-cross-package-dependency convention"
    - "Self-exec test-helper subprocess trick (evidence.TestEvidenceToolProbeHelper's own precedent) for deterministic truncation/timeout fixtures without a real clang binary"
    - "go/parser + go/ast source-level architectural-invariant tests (corevalidate.TestValidatorImportsStayIndependent's own precedent), extended here to also scan exported declaration/field names for judgement-shaped vocabulary"

key-files:
  created:
    - internal/compiler/cache/cache.go
    - internal/compiler/cache/cache_test.go
    - internal/compiler/cache/probe.go
    - internal/compiler/cache/probe_test.go

key-decisions:
  - "Task 0's checkpoint was pre-resolved at orchestration time (developer confirmed D-06-06 as locked, 2026-09-06): the cache stores compiled/instrumented/mutant ARTIFACTS ONLY, never a verdict, status, or check outcome; the checker, five-axis comparator, and sanitizer classifier always re-run fresh. D-06-13's four recorded escapes (undeclared environment, a Clang change that doesn't move its version string, future nondeterministic codegen, an indistinguishable hand-edited/partially-deleted cache directory) were acknowledged, not closed. Recorded here per the plan's explicit instruction, not re-litigated during execution."
  - "TestCacheExportedSurfaceStoresNoVerdict's forbidden-identifier denylist deliberately excludes the words 'status' and 'outcome': D-06-12 legitimately reuses both for cache-reuse REPORTING (cache.CacheStatus/cache.Outcome, values artifact_reused/artifact_recomputed/not_cacheable/unavailable) -- a fact about what THIS invocation skipped, never a judgement about whether a lane passed. The denylist instead targets the actual verdict/judgement/diagnostic vocabulary (verdict, judgement/judgment, diagnostic, passfail, testresult, checkresult, laneresult), so Task 1's structural test and Task 3's Outcome/Status additions coexist without contradiction."
  - "InputsFor(ctx context.Context, spec ArtifactSpec) carries a ctx parameter not present in the plan's literal signature, because it must call ProbeClangIdentity, whose 5-second bounded probe deadline requires one. [Rule 1/3 deviation, documented below.]"
  - "RuntimeIdentity/ForeignTranslationUnit/etc. in ArtifactSpec are always-required non-empty declared values, not per-kind-optional fields: a caller for whom no ASan/UBSan/libc++ runtime is linked must still supply an explicit value (e.g. \"none\") rather than leaving the field empty, since an empty declared input is refused outright as cache.input_undeclared (D-06-11) -- there is no notion of 'this lane doesn't need this input' baked into InputsFor itself, keeping the fail-closed rule uniform across all three artifact kinds rather than kind-conditional."
  - "ProbeClangIdentity's digest combines sha256(clang binary's own bytes) with the bounded --version probe output bytes (two separate hashes concatenated then re-hashed), rather than hashing the full binary content directly into one buffer -- avoids loading a multi-hundred-MB real clang binary's bytes into the same buffer as probe output, while still satisfying the literal 'combines the SHA-256 of the resolved clang binary's own bytes with the probe output' requirement."

requirements-completed: []

coverage:
  - id: D1
    description: "The cache stores only expensive intermediate artifacts and never a verdict; nothing in the package can record or return a pass/fail judgement (D-06-06)"
    requirement: "DX-03"
    verification:
      - kind: unit
        ref: "internal/compiler/cache/cache_test.go#TestCacheExportedSurfaceStoresNoVerdict"
        status: pass
      - kind: unit
        ref: "internal/compiler/cache/cache_test.go#TestCacheImportsStayIndependent"
        status: pass
    human_judgment: false
  - id: D2
    description: "A cache key is a SHA-256 over an explicitly declared input list, and meta.json records that full list so a reviewer can see exactly what the key covered (D-06-07, D-06-09)"
    requirement: "DX-03"
    verification:
      - kind: unit
        ref: "internal/compiler/cache/cache_test.go#TestCacheRoundTripsArtifactAndMeta"
        status: pass
      - kind: unit
        ref: "internal/compiler/cache/probe_test.go#TestCacheKeyCoversEveryDeclaredInput"
        status: pass
    human_judgment: false
  - id: D3
    description: "A mutant binary built by a changed mutation runner misses the cache, because the mutation runner's own source hash is a declared input (D-06-07)"
    requirement: "DX-03"
    verification:
      - kind: unit
        ref: "internal/compiler/cache/probe_test.go#TestMutationRunnerSourceHashIsADeclaredInput"
        status: pass
    human_judgment: false
  - id: D4
    description: "When any declared input cannot be computed (e.g. the Clang probe fails), the entry is not_cacheable for that run; ambiguity and silence resolve to run it, never to skip it (D-06-11)"
    requirement: "DX-03"
    verification:
      - kind: unit
        ref: "internal/compiler/cache/cache_test.go#TestCacheFailsClosedToNotCacheable"
        status: pass
      - kind: unit
        ref: "internal/compiler/cache/probe_test.go#TestClangDigestProbeIsBounded"
        status: pass
    human_judgment: false
  - id: D5
    description: "Two artifacts whose declared input lists are byte-identical resolve to the same cache entry -- a merge by construction, not a collision (FND-04 adjacency edge)"
    requirement: "DX-03"
    verification:
      - kind: unit
        ref: "internal/compiler/cache/cache_test.go#TestCacheEqualDeclaredInputsShareOneEntry"
        status: pass
    human_judgment: false
  - id: D6
    description: "A first run against an empty cache directory reports every artifact as recomputed and never as reused (FND-04 empty-input edge)"
    requirement: "DX-03"
    verification:
      - kind: unit
        ref: "internal/compiler/cache/cache_test.go#TestCacheEmptyStoreReportsCold"
        status: pass
    human_judgment: false
  - id: D7
    description: "The clang_identity input is a real probed digest of the binary's own bytes, not merely its --version string, and the probe is bounded"
    requirement: "DX-03"
    verification:
      - kind: unit
        ref: "internal/compiler/cache/probe_test.go#TestClangDigestIsNotJustTheVersionString"
        status: pass
      - kind: unit
        ref: "internal/compiler/cache/probe_test.go#TestClangDigestProbeIsBounded"
        status: pass
    human_judgment: false

duration: 55min
completed: 2026-09-06
status: complete
---

# Phase 6 Plan 4: Local Content-Addressed Artifact Cache Summary

**`internal/compiler/cache` -- a dependency-free leaf package storing only compiled/instrumented/mutant ARTIFACTS keyed on seven declared inputs (including a real probed Clang binary digest), structurally incapable of holding a verdict, and failing closed to `not_cacheable` on any ambiguity.**

## Checkpoint Confirmation (Task 0)

Task 0 was a `checkpoint:decision` whose `<blocking>` was already `false` with a recorded `<resolution>` at orchestration time (developer confirmed 2026-09-06). Not re-litigated during execution; recorded here per the plan's instruction:

- **D-06-06 confirmed locked:** the cache stores compiled/instrumented/mutant ARTIFACTS ONLY -- never a verdict, status, or check outcome. The checker, the five-axis comparator, and the sanitizer classifier always re-run fresh against whatever binary (cached or newly built) is on disk this invocation. A cache hit changes what is SKIPPED (recompilation), never what is ASSERTED.
- **D-06-13's four recorded escapes acknowledged, not closed:** (1) undeclared environment (locale, `ulimit`, filesystem case-sensitivity); (2) a Clang change that doesn't move its reported version string (the ccache `__TIME__`-class footgun); (3) future nondeterministic codegen silently breaking the "same inputs implies same artifact" premise; (4) a hand-edited or partially-deleted cache directory being indistinguishable from a cold one, since there is no integrity check beyond content-hash lookup.
- This constraint is enforced structurally, not by convention: `TestCacheExportedSurfaceStoresNoVerdict` and `TestCacheImportsStayIndependent` (Task 1) parse the package's own source and fail the build if a judgement-shaped identifier or a forbidden import (`protocol`/`diagnostic`/`session`) ever appears.

## Performance

- **Duration:** ~55 min
- **Tasks:** 3 (Task 0 confirmed, not executed as work)
- **Files created:** 4 (cache.go, cache_test.go, probe.go, probe_test.go)
- **Commits:** 4

## Accomplishments

- **Task 1 (tracer):** `cache.go` -- `Input`/`Key`/`ComputeKey` (SHA-256 over canonical sorted JSON, refusing duplicate names and empty digests with `cache.input_undeclared`), `Store`/`Open`/`Path` (sharded `<root>/<2-hex>/<sha256>/`), `Put`/`Get` with atomic temp-name-then-`os.Rename` writes so a crashed run leaves no directory `Get` reports as found. `TestCacheRoundTripsArtifactAndMeta` proves `ComputeKey -> Put -> Get` round-trips real bytes through a real `t.TempDir()`-rooted filesystem. `TestCacheExportedSurfaceStoresNoVerdict` (a `go/parser`+`go/ast` scan of every non-test file's exported declarations and struct fields against a judgement-shaped denylist) and `TestCacheImportsStayIndependent` (the `corevalidate.TestValidatorImportsStayIndependent` import-scan pattern, applied to `protocol`/`diagnostic`/`session`) make the "artifacts only, never verdicts" constraint structural.
- **Task 2 (tdd):** `probe.go` -- `DeclaredInputNames()` returns D-06-07's exact seven names in stable order (`fixture_source`, `build_flags`, `clang_identity`, `runtime_identity`, `foreign_translation_unit`, `mutation_runner_source`, `go_toolchain`), documented in the package doc comment as declared-not-complete with all four D-06-13 holes cited by name. `ArtifactSpec`/`InputsFor` assemble the seven inputs, refusing outright (`cache.input_undeclared`) when any one is empty or unresolvable. `ProbeClangIdentity` reuses evidence's own 64 KiB-plus-one bounded-writer / 5-second-deadline shape (duplicated locally, not imported, per this leaf package's dependency-free convention) and combines `sha256(clang binary's own bytes)` with the bounded `--version` output, so a Clang change that leaves its reported version string unchanged still moves the digest. `TestCacheKeyCoversEveryDeclaredInput` is table-driven and self-invalidating: it iterates `DeclaredInputNames()` at the end and fails if any declared name lacks a perturbation row.
- **Task 3 (tdd, RED then GREEN):** `cache.go` gained `CacheStatus`/`CacheStatuses()` (the exact four D-06-12 values: `artifact_reused`, `artifact_recomputed`, `not_cacheable`, `unavailable` -- never `hit`/`miss`), `Outcome`, and `Consult`. `Consult` classifies `spec.Kind` against the closed `ArtifactKinds()` set and calls `InputsFor` before ever touching `store`; an unclassified kind or any failed declared-input computation (including a deliberately-broken Clang probe) short-circuits to `StatusNotCacheable` with a zero `Key` and **no store lookup at all** -- verified by pointing `Consult` at a nonexistent store root and asserting the root still doesn't exist afterward. `TestCacheEqualDeclaredInputsShareOneEntry` proves byte-identical declared inputs (built from two independent `ArtifactSpec` values with separately-allocated but content-identical bytes) resolve to the same key and the same stored entry -- a merge, not a collision -- while a spec differing in one input still misses. `TestCacheEmptyStoreReportsCold` sweeps three specs (two valid, one unclassified) against a fresh `t.TempDir()` root and asserts none ever returns `StatusArtifactReused`. `TestCacheMismatchedMetaIsTreatedAsAbsent` hand-tampers a `meta.json`'s recorded input digest after a real `Put` and confirms `Get` reports absent rather than a corrupted hit.
- Ran the full plan-level verification: `go test ./internal/compiler/cache/...` and `go test -race ./internal/compiler/cache/...` both pass clean; a whole-repo `go test ./...` (17 packages) was also run clean after Task 3 to confirm no regression outside this new package.

## Task Commits

Each task was committed atomically; Task 3 (`tdd="true"`) followed a genuine RED/GREEN split since `Consult`/`Outcome`/`CacheStatus` did not exist until the GREEN commit:

1. **Task 1: Tracer round trip** -- `5f08d6e` (feat) -- cache.go, cache_test.go.
2. **Task 2: Seven declared inputs and bounded Clang probe** -- `18af3b4` (feat) -- probe.go, probe_test.go.
3. **Task 3 RED** -- `a3bdc48` (test) -- cache_test.go additions referencing not-yet-existing `Consult`/`Outcome`/`CacheStatuses`/`Status*`; confirmed to fail to compile before the implementation existed.
4. **Task 3 GREEN** -- `ffe2b5e` (feat) -- cache.go additions implementing `CacheStatus`/`CacheStatuses`/`Outcome`/`Consult`; all Task 3 tests plus the full existing suite pass.

## Files Created/Modified

- `internal/compiler/cache/cache.go` (new) -- `Schema`, `Error`, `Input`, `Key`, `ComputeKey`, `Store`, `Open`, `Path`, `Put`, `Get`, `CacheStatus`/`StatusArtifactReused`/`StatusArtifactRecomputed`/`StatusNotCacheable`/`StatusUnavailable`, `CacheStatuses`, `Outcome`, `Consult`.
- `internal/compiler/cache/cache_test.go` (new) -- `TestCacheRoundTripsArtifactAndMeta`, `TestCachePutInterruptedBeforeRenameLeavesNoPartialEntry`, `TestCacheExportedSurfaceStoresNoVerdict`, `TestCacheImportsStayIndependent`, `TestCacheStatusVocabularyIsClosed`, `TestCacheFailsClosedToNotCacheable`, `TestCacheEqualDeclaredInputsShareOneEntry`, `TestCacheEmptyStoreReportsCold`, `TestCacheMismatchedMetaIsTreatedAsAbsent`.
- `internal/compiler/cache/probe.go` (new) -- `DeclaredInputNames`, `ArtifactKinds`/`KindCompiledBinary`/`KindInstrumentedBinary`/`KindMutantBinary`, `ArtifactSpec`, `InputsFor`, `ProbeClangIdentity` (plus unexported bounded-probe/command-factory machinery).
- `internal/compiler/cache/probe_test.go` (new) -- `TestCacheKeyCoversEveryDeclaredInput`, `TestMutationRunnerSourceHashIsADeclaredInput`, `TestClangDigestIsNotJustTheVersionString`, `TestClangDigestProbeIsBounded`, `TestCacheProbeHelper` (self-exec subprocess fixture).

## Decisions Made

See `key-decisions` in frontmatter. Most consequential: the deliberate exclusion of "status"/"outcome" from the verdict-denylist so Task 1's structural test and Task 3's `Outcome`/`Status` additions coexist without contradiction, and the always-required (never per-kind-optional) declared-input fields keeping the fail-closed rule uniform.

## Deviations from Plan

**1. [Rule 1/3 - blocking issue] `InputsFor` takes a `context.Context` parameter.** The plan's literal signature is `func InputsFor(spec ArtifactSpec) ([]Input, error)`, but `InputsFor` must call `ProbeClangIdentity(ctx, clangPath)`, whose 5-second probe deadline requires a `context.Context`. Added `ctx context.Context` as the first parameter; every call site (tests, `Consult`) passes one through. No other signature in the plan's `<artifacts_this_phase_produces>` list needed a similar fix.

**2. [Discretion] Verdict-denylist excludes "status"/"outcome".** Task 1's `<action>` lists "status, verdict, outcome, diagnostic, lane, or pass/fail type" as illustrative examples of what must be absent at that point in the plan's own sequencing (before Task 3 exists). Since Task 3 explicitly names its own types `Outcome`/`Status` for cache-reuse reporting (D-06-12), `TestCacheExportedSurfaceStoresNoVerdict`'s denylist targets the narrower, load-bearing vocabulary (`verdict`, `judgement`/`judgment`, `diagnostic`, `passfail`, `testresult`, `checkresult`, `laneresult`) rather than the literal example words, so the same test passes unmodified across all three tasks. The import-boundary test (`protocol`/`diagnostic`/`session` never imported) remains the primary structural enforcement mechanism and is unaffected by this choice.

**Total deviations:** 1 auto-fixed (Rule 1/3), 1 discretionary (denylist scoping). **Impact:** neither blocks phase completion; both are documented, tested, and independently re-verified not to regress the tracer's own acceptance criteria.

## Known Stubs

None. Every exported function in this plan has a real, tested implementation; there is no placeholder or empty-value stub blocking DX-03.

## Flagged Assumption Carried Forward

Per the plan's own `<flagged_assumptions>`: the coverage report classified DX-03's edge case as "unclassified -- review manually" and it remains `unresolved`. Surfaced here, not dropped, per the plan's instruction. Reviewer question for a later verification pass: is there a cache or lane-selection edge case beyond adjacency/empty/ordering that this plan does not cover? This plan does not attempt to resolve it -- 06-05 and 06-07 (the rest of DX-03) are the next opportunities to revisit.

## Issues Encountered

None blocking.

## User Setup Required

None -- no external service configuration required. The cache directory (`os.UserCacheDir()/lang-verify/...`) is created lazily on first use and is deliberately not git-tracked.

## Next Phase Readiness

- `internal/compiler/cache` exists, is fully tested (`go test`/`go test -race` both clean), and imports nothing beyond the Go standard library.
- **DX-03 is NOT marked complete in REQUIREMENTS.md by this plan** -- 06-05 and 06-07 are the remaining work for that requirement and have not run yet. This plan's `requirements-completed` is deliberately empty.
- `cache.Consult`/`cache.Outcome`/`cache.CacheStatuses()` are ready for a later plan (likely 06-05 or 06-06) to wire into `protocol.Lane`'s `cache_status`/`selection_reason`/`cache_inputs_reused_count` fields (D-06-12) -- that wiring is explicitly out of this plan's scope (files_modified lists only the `cache` package itself).
- No blockers for subsequent Phase 6 plans.

---
*Phase: 06-agent-feedback-and-performance-ratification*
*Completed: 2026-09-06*

## Self-Check: PASSED

- FOUND: internal/compiler/cache/cache.go
- FOUND: internal/compiler/cache/cache_test.go
- FOUND: internal/compiler/cache/probe.go
- FOUND: internal/compiler/cache/probe_test.go
- FOUND: commit 5f08d6e (Task 1)
- FOUND: commit 18af3b4 (Task 2)
- FOUND: commit a3bdc48 (Task 3 RED)
- FOUND: commit ffe2b5e (Task 3 GREEN)
