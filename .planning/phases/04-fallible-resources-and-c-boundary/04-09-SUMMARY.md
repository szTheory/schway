---
phase: 04-fallible-resources-and-c-boundary
plan: "09"
subsystem: compiler
tags: [cgen, session, gap-closure, mutation-kill, foreign-boundary, attribute-scan]

requires:
  - phase: 04-fallible-resources-and-c-boundary
    provides: "commit cb701b5's eight-argument ScanForBannedAttributes wiring in session.go's lane:foreign-no-unproven-attributes (WR-01 fix pass), which this plan falsifies rather than re-implements"
provides:
  - "A falsifier for the third D-04-12 inspectable layer: TestAttributeInjectionIntoConformanceOnlyMakesControlFail proves a banned token injected into EmitForeignConformance's own added text (the region after the embedded header) is detected, independent of the header-only test"
  - "A pin on the production lane's scanned-artifact count: TestAttributeScanLaneCoversEveryInspectableLayer asserts lane:foreign-no-unproven-attributes reports RecomputedWork exactly 8, so dropping a scan argument from session.go turns a named test red"
  - "Two new Mutation-Kill Register rows in 04-VALIDATION.md, both demonstrated with verbatim revert-and-fail output"
affects: [04-VERIFICATION.md gap 2, future cgen/session work touching the foreign-attribute scan lane]

actuals:
  tokens: 2250
  tasks: 2
  commits: 2

tech-stack:
  added: []
  patterns:
    - "A layer whose text is a superset of another layer's (EmitForeignConformance embeds EmitForeignHeader verbatim) requires isolating the region AFTER the shared prefix before injecting, or the falsifier proves only what the inner layer's own test already proves"
    - "A control that names full coverage over N artifacts is only as strong as an assertion pinning N; asserting status=='pass' alone lets the argument list shrink silently"

key-files:
  created: []
  modified:
    - internal/compiler/session/session_test.go
    - .planning/phases/04-fallible-resources-and-c-boundary/04-VALIDATION.md

key-decisions:
  - "Both new tests consume cgen.BannedOptimizerAttributes (joined into a single injected token string) rather than a hard-coded literal, per the acceptance criteria and matching the existing header-only test's stated reason: widening the banned set must not silently bypass the mutation-kill."
  - "TestAttributeScanLaneCoversEveryInspectableLayer asserts both an exact work value of 8 AND RecomputedWork>0 (redundant given Lane.RecomputedWork is the single field addLane's work parameter populates, but both assertions are kept per the plan's literal behavior spec so the failure message names the observed value explicitly)."
  - "No production code changed in either task -- session.go and cgen.go were already correct as of commit cb701b5; this plan supplies the two missing falsifiers (conformance-only injection, artifact-count pin) that fix shipped without."

patterns-established:
  - "A gap-closure plan's mutation-kill demonstration for a lane whose test doesn't exist in git history yet copies the new test function into the throwaway detached worktree (via awk-extracted snippet) before reverting the production hunk, so the revert's effect on the NEW test can be observed even though the test itself is uncommitted at demo time."

requirements-completed: [FFI-01]

coverage:
  - id: D1
    description: "The conformance translation unit's own added text (the region after the embedded header) is independently scanned for banned optimizer attributes -- a banned token there is detected while the untouched header and compiled-program C stay clean"
    requirement: "FFI-01"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestAttributeInjectionIntoConformanceOnlyMakesControlFail"
        status: pass
    human_judgment: false
  - id: D2
    description: "The production lane's scanned-artifact count is pinned at 8 (two fixture programs x four artifacts each), so dropping any scan argument from session.go turns a named test red instead of silently shrinking coverage"
    requirement: "FFI-01"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestAttributeScanLaneCoversEveryInspectableLayer"
        status: pass
    human_judgment: false
  - id: D3
    description: "Both falsifiers are mutation-killed: reverting the conformance arguments (in a throwaway worktree) leaves both TestVerifyPhase4ControlsAndWork and TestNoUnprovenAttributesEmitted green today (proving the gap), and separately reverting conformance-only or header-only arguments against the new count-pin test both turn it red"
    verification:
      - kind: other
        ref: "manual git-worktree revert-and-run demonstrations (see Mutation-Kill Demonstrations below)"
        status: pass
    human_judgment: false
  - id: D4
    description: "The whole Phase 4 gate stays green after both new falsifiers land, with control:foreign.no_unproven_attributes reporting recomputed_work=8"
    verification:
      - kind: other
        ref: "sh scripts/verify-phase4.sh (exit 0, all lanes pass, lane:foreign-no-unproven-attributes recomputed_work=8)"
        status: pass
      - kind: other
        ref: "go test ./... and go vet ./... (both clean)"
        status: pass
    human_judgment: false

duration: 25min
completed: 2026-09-05
status: complete
---

# Phase 4 Plan 09: Conformance-Layer Falsifier and Attribute-Scan Artifact-Count Pin Summary

**Two new session_test.go tests close 04-VERIFICATION.md gap 2: one proves the generated conformance unit's own added text is independently scanned for banned optimizer attributes, the other pins the production lane's scanned-artifact count at 8 so a future argument-list regression turns a named test red instead of passing silently.**

## Performance

- **Duration:** 25 min
- **Started:** 2026-09-05T18:05:00Z
- **Completed:** 2026-09-05T18:30:00Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments

- **Task 1:** Added `TestAttributeInjectionIntoConformanceOnlyMakesControlFail` to `session_test.go`. `EmitForeignConformance` embeds `EmitForeignHeader`'s full output verbatim before writing its own added text (the private-header `#include` line and per-field `_Static_assert` triples), so the existing header-only falsifier proves nothing about the conformance unit's own region. The new test isolates that post-header region via `strings.Index(conformance, header)`, injects a token built from `cgen.BannedOptimizerAttributes` into it alone, and asserts: the real emitted artifacts are clean first (fail loudly otherwise, or the test would be vacuous); the header is found inside the conformance output (fail loudly otherwise, since the embedding assumption would no longer hold); the untouched header and compiled-program C stay clean under the injection; and the reassembled injected conformance unit is flagged. Mutation-kill demonstration: reverting `tracerConformance`/`releaseConformance` from `session.go`'s `ScanForBannedAttributes` call in a throwaway detached worktree left `TestVerifyPhase4ControlsAndWork` and `TestNoUnprovenAttributesEmitted` both green -- confirming the production regression this task's falsifier is designed to catch does not, today, turn any *other* test red, which is precisely why Task 2 exists.
- **Task 2:** Added `TestAttributeScanLaneCoversEveryInspectableLayer` to `session_test.go`, near `TestVerifyPhase4ControlsAndWork`. It runs the Phase 4 corpus verify, locates `lane:foreign-no-unproven-attributes` (failing loudly if the lane is absent entirely, not merely if its assertions fail), and asserts status `pass` and `RecomputedWork == 8` (two fixture programs times four artifacts: compiled program, sidecar manifest, generated header, generated conformance unit). Mutation-kill demonstrated twice in separate throwaway detached worktrees: dropping the conformance arguments (work lowered to 6) turned the new test red; dropping the header arguments instead (work also lowered to 6) also turned it red -- proving the pin catches either layer's omission, not just one. Appended two rows to `04-VALIDATION.md`'s Mutation-Kill Register, both owning plan `04-09`, after 04-08's rows with no deletions.

## Task Commits

Each task was committed atomically:

1. **Task 1: The conformance-only falsifier** - `d99dde5` (test)
2. **Task 2: Pin the production lane's scanned-artifact count** - `0340b54` (test)

_Note: both tasks are test-only additions against an already-shipped, already-correct production fix (commit `cb701b5`) -- neither `session.go` nor `cgen.go` was modified by this plan._

## Files Created/Modified

- `internal/compiler/session/session_test.go` - Adds `TestAttributeInjectionIntoConformanceOnlyMakesControlFail` (Task 1) and `TestAttributeScanLaneCoversEveryInspectableLayer` (Task 2).
- `.planning/phases/04-fallible-resources-and-c-boundary/04-VALIDATION.md` - Two new Mutation-Kill Register rows, owning plan `04-09`, both `✅ green`.

## Decisions Made

See key-decisions in frontmatter. In short: both tests consume `cgen.BannedOptimizerAttributes` rather than hard-coded literals (acceptance-criteria requirement); no production code was touched since the WR-01 fix pass (`cb701b5`) already wired all eight arguments correctly -- this plan supplies the falsifiers that fix shipped without.

## Mutation-Kill Demonstrations

### Task 1 revert: commit `cb701b5`'s conformance arguments in `session.go`'s `ScanForBannedAttributes` call

Reverted in a throwaway detached worktree (`git worktree add --detach`, then a targeted Python edit dropping `tracerConformance`/`releaseConformance` from the argument list, with `_ = tracerConformance`/`_ = releaseConformance` added to keep the build clean, plus the equivalent edit to `TestNoUnprovenAttributesEmitted`'s own scan call). The new falsifier test (`TestAttributeInjectionIntoConformanceOnlyMakesControlFail`, copied into the worktree via an `awk`-extracted snippet since it was not yet committed to git history at demo time) was then run alongside the existing lane tests:

```
=== RUN   TestNoUnprovenAttributesEmitted
--- PASS: TestNoUnprovenAttributesEmitted (0.00s)
=== RUN   TestVerifyPhase4ControlsAndWork
--- PASS: TestVerifyPhase4ControlsAndWork (1.40s)
=== RUN   TestAttributeInjectionIntoConformanceOnlyMakesControlFail
--- PASS: TestAttributeInjectionIntoConformanceOnlyMakesControlFail (0.00s)
PASS
ok  	github.com/codename-lang/lang/internal/compiler/session	1.616s
```

Both the production lane's own test (`TestVerifyPhase4ControlsAndWork`) and its manifest-level regression test (`TestNoUnprovenAttributesEmitted`) stay green with the conformance arguments dropped -- confirming that today, without Task 2's pin, this exact regression turns no test red even though the required control still reports "pass". The new falsifier test itself still demonstrates the scan's capability (it constructs its own injected string and calls `cgen.ScanForBannedAttributes` directly, independent of the production lane's argument list). The worktree was deleted immediately after; no revert reached `main`.

### Task 2 revert 1: dropping the conformance arguments and lowering declared work to 6

Reverted in a throwaway detached worktree, with the new `TestAttributeScanLaneCoversEveryInspectableLayer` copied in via the same `awk`-extraction method, then run:

```
=== RUN   TestAttributeScanLaneCoversEveryInspectableLayer
    session_test.go:2704: lane:foreign-no-unproven-attributes work = 6, want 8 (two fixture programs x four artifacts each); a value below 8 means a scanned artifact was dropped from session.go's ScanForBannedAttributes argument list: &{Schema:lang.verify-lane/0 ID:lane:foreign-no-unproven-attributes Status:pass Controls:[control:foreign.no_unproven_attributes] RecomputedWork:6 ElapsedNS:469375 PeakRSSStatus:unavailable PeakRSSBytes:0 OutputBytes:21560}
--- FAIL: TestAttributeScanLaneCoversEveryInspectableLayer (1.51s)
FAIL
FAIL	github.com/codename-lang/lang/internal/compiler/session	1.777s
FAIL
```

The worktree was deleted immediately after; no revert reached `main`.

### Task 2 revert 2: dropping the header arguments instead (keeping conformance) and lowering declared work to 6

Reverted in a second throwaway detached worktree, same test-copy method:

```
=== RUN   TestAttributeScanLaneCoversEveryInspectableLayer
    session_test.go:2703: lane:foreign-no-unproven-attributes work = 6, want 8 (two fixture programs x four artifacts each); a value below 8 means a scanned artifact was dropped from session.go's ScanForBannedAttributes argument list: &{Schema:lang.verify-lane/0 ID:lane:foreign-no-unproven-attributes Status:pass Controls:[control:foreign.no_unproven_attributes] RecomputedWork:6 ElapsedNS:517833 PeakRSSStatus:unavailable PeakRSSBytes:0 OutputBytes:23590}
--- FAIL: TestAttributeScanLaneCoversEveryInspectableLayer (1.34s)
FAIL
FAIL	github.com/codename-lang/lang/internal/compiler/session	1.541s
FAIL
```

Both reverts turn the pin red regardless of which layer (header or conformance) is dropped, proving the count assertion catches either omission -- not just the conformance-specific one Task 1 targets. The worktree was deleted immediately after; no revert reached `main`.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None. Both new tests passed on first implementation; all four mutation-kill reverts (one for Task 1, two for Task 2, plus the demonstration that Task 1's regression turns no *other* test red today) produced their expected output on first attempt.

## Verification Results

- `env GOCACHE=/tmp/ai-lang-phase4-cache go test ./...` - exits 0, all packages pass
- `env GOCACHE=/tmp/ai-lang-phase4-cache go vet ./...` - exits 0, clean
- `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/session TestAttributeInjectionIntoConformanceOnlyMakesControlFail TestAttributeInjectionIntoHeaderOnlyMakesControlFail TestAttributeInjectionMakesControlFail TestNoUnprovenAttributesEmitted TestNoreturnExemptionIsNamed` - exits 0
- `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/session TestAttributeScanLaneCoversEveryInspectableLayer TestVerifyPhase4ControlsAndWork TestVerifyPhase4CLI TestPhase4RequiredControlsMatchScript TestPhase4VerifierScriptContract TestExpectedEscapesAreVisibleNotSolved` - exits 0
- `sh scripts/verify-phase4.sh` - exits 0; `lane:foreign-no-unproven-attributes` reports `"status":"pass"`, `"recomputed_work":8`; all required control IDs present including `control:foreign.no_unproven_attributes`
- `git diff --stat internal/compiler/session/session.go internal/compiler/cgen/cgen.go scripts/verify-phase4.sh` - empty (no production or gate-script file changed)
- `git diff --stat .planning/phases/04-fallible-resources-and-c-boundary/04-VALIDATION.md` - 2 insertions only, no deletions of existing rows

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- 04-VERIFICATION.md gap 2 is closed: item 1 ("add cgen.EmitForeignHeader/EmitForeignConformance output to the scan") was already implemented by commit `cb701b5`; item 2 ("add a dedicated mutation-kill test injecting a banned token into the header/conformance layer independently, since a byte-identical format string could give incidental coverage") is now closed for the conformance layer by Task 1, and the production lane's argument-list regression this whole gap was about is now pinned red by Task 2.
- Re-running `/gsd-verify-work` on Phase 04 should flip the 04-03 must_have "one authoritative core.ForeignContract... and all three inspectable layers are derived from it" from `failed` to verified, on the evidence of `TestAttributeInjectionIntoConformanceOnlyMakesControlFail` (constructs the previously-untested conformance-only shape) and `TestAttributeScanLaneCoversEveryInspectableLayer` (pins the count so the exact regression 04-VERIFICATION.md found cannot recur silently) -- not on the strength of a control identifier that names coverage it does not demonstrate.
- Both of 04-VERIFICATION.md's confirmed gaps (gap 1 via 04-08, gap 2 via this plan) are now closed. No further gap-closure plans are known to be required for Phase 04 pending re-verification.

## Self-Check: PASSED

- `internal/compiler/session/session_test.go` exists and contains both new tests - verified.
- `.planning/phases/04-fallible-resources-and-c-boundary/04-VALIDATION.md` exists and contains the two new register rows - verified.
- Commit `d99dde5` found in `git log --oneline --all`.
- Commit `0340b54` found in `git log --oneline --all`.
- All plan-level `<verification>` commands re-run and passing at summary time.

---
*Phase: 04-fallible-resources-and-c-boundary*
*Completed: 2026-09-05*
