---
phase: 11-multi-function-native-emission-and-interprocedural-equivalen
plan: 07
subsystem: build-cache
tags: [cache, cgen, sha256, go-parser, go-list-deps, closure-digest, qlt-06]

# Dependency graph
requires:
  - phase: 11 (plan 11-01)
    provides: "Q-02 verdict (BRANCH A, hole reproduces) deciding this plan ships a real eighth cache input"
provides:
  - "cache.DeclaredInputNames() eighth entry (cgen_source), closing D-11-41"
  - "QLT-06b's structural discharge tests (dual import scans + abstention proof)"
  - "11-QLT06-ABSTENTION.md, the QLT-06a/b split record"
affects: []

# Actuals (#2632)
actuals:
  tokens: 32000
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Additive-sibling discipline for a declared cache input list: append, never insert/rename/reorder, and pin the exact byte-identical prefix in a test"
    - "Dual independent import scans (direct go/parser ImportsOnly + bounded/timed transitive go list -deps) each with its own negative control, copied verbatim from originvalidate_test.go's own precedent"
    - "Forbidden-identifier source scan built via string concatenation so the test's own name/doc-comment cannot trip its own assertion"

key-files:
  created:
    - .planning/phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-QLT06-ABSTENTION.md
  modified:
    - internal/compiler/cache/probe.go
    - internal/compiler/cache/probe_test.go
    - internal/compiler/cache/cache_test.go
    - internal/compiler/session/session_phase6_verify.go
    - internal/compiler/session/session_phase6_escapes.go
    - internal/compiler/session/session_phase6_cache_hole_test.go

key-decisions:
  - "Q-02's BRANCH A verdict (from plan 11-01) was confirmed still current: DeclaredInputNames() widened from seven to eight names, eighth named cgen_source, computed by hashing internal/compiler/cgen/*.go directly via crypto/sha256+os."
  - "QLT-06 ships as a split row (QLT-06a Complete / QLT-06b Complete-by-abstention) per the OWN-05a/05b precedent, recorded in 11-QLT06-ABSTENTION.md; REQUIREMENTS.md's own table is left as a single QLT-06 row since this plan's files_modified does not include it -- the abstention doc is the authoritative split record."
  - "TestNoClosureDigestInCache's source scan is restricted to non-test files: the mandated test name itself contains the forbidden identifier as a substring, so a whole-directory scan (including the test's own file) would produce a permanent, unfixable false positive. The forbidden identifier itself is built via string concatenation in the test body to avoid a second self-trip."

patterns-established:
  - "A declared cache input list widens by strict append with a byte-identical, order-pinned prefix assertion -- never insertion, rename, or reorder."

requirements-completed: [QLT-06]

coverage:
  - id: D1
    description: "cache.DeclaredInputNames() gains cgen_source as an eighth, appended declared input, wired into phase6ArtifactSpec/ComputeKey so an edited cgen genuinely moves the artifact key"
    requirement: "QLT-06"
    verification:
      - kind: unit
        ref: "internal/compiler/cache/probe_test.go#TestCacheKeyCoversEveryDeclaredInput/cgen_source"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_cache_hole_test.go#TestQ02DeclaredInputNamesStillSevenNoCgen"
        status: pass
    human_judgment: false
  - id: D2
    description: "QLT-06b's structural discharge: dual independent import scans (direct + transitive) each with a negative control, plus a mechanical no-ClosureDigest-in-cache proof and a key-idempotency proof"
    requirement: "QLT-06"
    verification:
      - kind: unit
        ref: "internal/compiler/cache/probe_test.go#TestDeclaredInputNamesStableAndNoInterproceduralImport"
        status: pass
      - kind: unit
        ref: "internal/compiler/cache/probe_test.go#TestCacheDirectImportGuardCanFail"
        status: pass
      - kind: unit
        ref: "internal/compiler/cache/probe_test.go#TestCacheTransitiveImportGuardCanFail"
        status: pass
      - kind: unit
        ref: "internal/compiler/cache/probe_test.go#TestNoClosureDigestInCache"
        status: pass
      - kind: unit
        ref: "internal/compiler/cache/probe_test.go#TestCacheKeyIsIdempotent"
        status: pass
    human_judgment: false
  - id: D3
    description: "11-QLT06-ABSTENTION.md records the QLT-06a/b split, the whole-program-hash dominance argument, the D-11-40 correction, and Q-02's outcome"
    verification:
      - kind: other
        ref: "grep -v '^#' 11-QLT06-ABSTENTION.md | grep -c -E 'QLT-06a|QLT-06b|strictly dominates|upper bound on a model|TestCalleeChangeInvalidatesCallerClosureDigest' -- returns 8 (>= 5 required)"
        status: pass
    human_judgment: false

duration: 45min
completed: 2026-09-12
status: complete
---

# Phase 11 Plan 7: Eighth Cache Input and QLT-06 Split Summary

**`cache.DeclaredInputNames()` now returns eight names (BRANCH A confirmed): the original seven, byte-identical and ordered, plus `cgen_source` appended eighth, closing the live D-11-41 stale-cgen cache-reuse hole with a real key-moving mechanism and a two-scan structural proof that nothing interprocedural is ever cached.**

## Performance

- **Duration:** ~45 min
- **Started:** 2026-09-12T08:35:00Z (approx.)
- **Completed:** 2026-09-12T09:24:10Z
- **Tasks:** 3 completed
- **Files modified:** 6 (5 modified, 1 created)

## Accomplishments

- **The eighth declared cache input shipped for real, not just declared.**
  `CgenSourceDigest` hashes every `*.go` file directly under a caller-supplied
  `cgenDir` via `crypto/sha256`/`os` (no import of `core`/`corevalidate`/
  `originvalidate`), fails closed on an empty file set, and is threaded through
  `phase6ArtifactSpec`'s new `CgenSourceDir` field into `ComputeKey` so an edited
  `cgen` genuinely moves the artifact `Key.ID`. `probe.go`'s D-06-13 hole-comment
  and `session_phase6_escapes.go`'s mirrored disclosure both gained a fifth,
  CLOSED item; the original four are byte-for-byte untouched
  (`git diff` on both files shows only additions before/after the existing
  numbered lists).
- **QLT-06b discharged by two independent, individually falsifiable import
  scans plus a mechanical abstention proof**, copying
  `originvalidate_test.go`'s three-part guard shape verbatim: a direct
  `go/parser` `ImportsOnly` scan and a bounded/timed transitive `go list
  -deps` scan, each with its own negative control
  (`TestCacheDirectImportGuardCanFail` / `TestCacheTransitiveImportGuardCanFail`),
  plus `TestNoClosureDigestInCache` (D-11-38's abstention made mechanical) and
  `TestCacheKeyIsIdempotent` (idempotency truth from the plan's must-haves).
- **QLT-06 recorded as a split row** in `11-QLT06-ABSTENTION.md`: QLT-06a
  Complete (Phase 07, two independent mutation-killed knowers), QLT-06b
  Complete-by-abstention (Phase 11, this plan's structural tests), the
  whole-program-hash dominance argument, the D-11-40 correction to S-006's
  carried eviction figures, and Q-02's BRANCH A outcome with what it means for
  cache soundness going forward.

## Task Commits

Each task was committed atomically:

1. **Task 1: Declare cgen's source as the eighth cache input** - `5e8c054` (feat)
2. **Task 2: QLT-06b's structural discharge — two independent import scans plus their negative controls** - `1123d61` (test)
3. **Task 3: Write the QLT-06a/b split and the abstention record** - `108a4b1` (docs)

## Files Created/Modified

- `internal/compiler/cache/probe.go` - `CgenSourceDigest`, `ArtifactSpec.CgenSourceDir`, `DeclaredInputNames()` widened to eight, fifth hole-disclosure item appended
- `internal/compiler/cache/probe_test.go` - `baseArtifactSpec`/`TestCacheKeyCoversEveryDeclaredInput` updated for the eighth input; five new QLT-06b structural-discharge tests appended
- `internal/compiler/cache/cache_test.go` - `newArtifactSpecFixture` updated to supply a `CgenSourceDir` fixture
- `internal/compiler/session/session_phase6_verify.go` - `phase6CgenSourceDir()` helper; `phase6ArtifactSpec` threads `CgenSourceDir`
- `internal/compiler/session/session_phase6_escapes.go` - mirrored fifth-hole-closed disclosure comment
- `internal/compiler/session/session_phase6_cache_hole_test.go` - `TestQ02DeclaredInputNamesStillSevenNoCgen` repinned from seven to eight names
- `.planning/phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-QLT06-ABSTENTION.md` (new) - the QLT-06a/b split record

## Decisions Made

- Q-02's BRANCH A verdict from plan 11-01 was taken as authoritative and re-verified against 11-01-SUMMARY.md before writing code; no new spike was re-run.
- `cache_test.go`'s `newArtifactSpecFixture` and `probe_test.go`'s `baseArtifactSpec` (both pre-existing shared test fixtures, out of this plan's declared `files_modified` but load-bearing for every existing cache test) were updated to supply a `CgenSourceDir`, since `InputsFor` now fails closed without one. This is the minimum viable widening to keep the pre-existing test suite green under the new eighth required input — see Deviations.
- REQUIREMENTS.md's QLT-06 row is left as a single row rather than formally split into QLT-06a/QLT-06b text, since this plan's `files_modified` does not list it; `11-QLT06-ABSTENTION.md` is the authoritative split record per the OWN-05a/05b precedent's own spirit.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Existing cache test fixtures needed a `CgenSourceDir` to keep passing**
- **Found during:** Task 1
- **Issue:** `InputsFor` now fails closed (`cache.input_undeclared`) on an empty `CgenSourceDir`, which every pre-existing `ArtifactSpec` fixture in `cache_test.go` and `probe_test.go` omitted. Running the existing suite after adding the eighth input broke `TestCacheEqualDeclaredInputsShareOneEntry`, `TestCacheFailsClosedToNotCacheable`, `TestCacheEmptyStoreReportsCold`, and `TestCacheKeyCoversEveryDeclaredInput` itself.
- **Fix:** Added a small synthetic `cgen`-shaped fixture directory to `baseArtifactSpec` (`probe_test.go`) and `newArtifactSpecFixture` (`cache_test.go`), and widened `TestCacheKeyCoversEveryDeclaredInput`'s self-invalidating perturbation table with a `cgen_source` row.
- **Files modified:** `internal/compiler/cache/probe_test.go`, `internal/compiler/cache/cache_test.go`
- **Verification:** `go test ./internal/compiler/cache/... -v -count=1` green, all pre-existing tests plus the five new ones passing.
- **Committed in:** `5e8c054` (probe_test.go fixture piece) / `1123d61` (the new Task 2 tests, in the same file)

**2. [Rule 1 - Bug] Task 2's mandated test name would self-trip a whole-directory literal scan**
- **Found during:** Task 2
- **Issue:** The plan names the abstention-proof test `TestNoClosureDigestInCache`, and the plan's top-level `<verification>` bullet is `grep -r 'ClosureDigest' internal/compiler/cache/` returns nothing. The test's own function name and doc comment necessarily contain the forbidden identifier as a substring, so a literal whole-directory scan (test files included) can never return zero once this test exists — the mandated name and the mandated verification bullet are structurally incompatible.
- **Fix:** Scoped the test's own source scan to non-test `.go` files (the actual binding guarantee: production code never references the identifier), and built the forbidden identifier via string concatenation (`"Closure" + "Digest"`) in the test body so the check logic itself does not add a second, avoidable occurrence. The unavoidable occurrence is the mandated test's own name and doc comment (2 lines, both explaining what the test does).
- **Files modified:** `internal/compiler/cache/probe_test.go`
- **Verification:** `TestNoClosureDigestInCache` passes; `grep -r 'ClosureDigest' internal/compiler/cache/` returns exactly the two lines naming the test itself (its declaration and doc comment), not a reference to the actual field or a cache mechanism using it.
- **Committed in:** `1123d61`

---

**Total deviations:** 2 auto-fixed (1 Rule 3 blocking-fix, 1 Rule 1 bug/self-consistency fix). **Impact on plan:** Both fixes were necessary to keep the existing suite green and to make the mandated test name coexist with its own assertion; neither widens scope beyond what Task 1/Task 2 already required.

## Issues Encountered

None beyond the two deviations above, both resolved inline.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- The D-11-41 stale-cgen cache-soundness hole is closed with a real,
  key-moving mechanism and a passing test (`TestCacheKeyCoversEveryDeclaredInput/cgen_source`).
- QLT-06 is fully discharged (both halves) and recorded in
  `11-QLT06-ABSTENTION.md`; `requirements.mark-complete` will check off the
  single QLT-06 row in REQUIREMENTS.md (the split text lives in the
  abstention doc, per this plan's `files_modified` scope).
- `go build ./...` and `go test ./...` are green across the whole repository
  after these changes.
- No blockers for remaining Phase 11 plans (11-08 already landed per STATE.md;
  this was the last scoped item for the cache-soundness track).

## Self-Check: PASSED

- `internal/compiler/cache/probe.go` — FOUND
- `internal/compiler/cache/probe_test.go` — FOUND
- `internal/compiler/cache/cache_test.go` — FOUND
- `internal/compiler/session/session_phase6_verify.go` — FOUND
- `internal/compiler/session/session_phase6_escapes.go` — FOUND
- `internal/compiler/session/session_phase6_cache_hole_test.go` — FOUND
- `.planning/phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-QLT06-ABSTENTION.md` — FOUND
- Commit `5e8c054` — FOUND in `git log --oneline --all`
- Commit `1123d61` — FOUND in `git log --oneline --all`
- Commit `108a4b1` — FOUND in `git log --oneline --all`
- `go build ./...` — PASSED
- `go test ./...` — PASSED (all packages ok, two `[no test files]` notices unrelated to this plan)
- `go test ./internal/compiler/cache/... -v -count=1` — PASSED, 21 tests including the 5 new ones
- Plan-level `<verification>` re-run: `go test ./internal/compiler/cache/... -v -count=1` green; `go test ./...` green; `grep -r 'ClosureDigest' internal/compiler/cache/` returns only the mandated test's own name/comment (documented deviation, not a production reference); `11-QLT06-ABSTENTION.md` grep check returns 8 (>= 5 required).

---
*Phase: 11-multi-function-native-emission-and-interprocedural-equivalen*
*Completed: 2026-09-12*
