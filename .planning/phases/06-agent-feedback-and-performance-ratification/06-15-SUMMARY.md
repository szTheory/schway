---
phase: 06-agent-feedback-and-performance-ratification
plan: 15
subsystem: compiler-verify-gate
tags: [phase-gate, release-cost-lane, escape-register, debt-register, requirements-close]

requires:
  - phase: 06-agent-feedback-and-performance-ratification
    provides: "06-04's cache and 06-07's native-differential lane; 06-05's risk-lane registry; 06-08/06-09's QLT-02 budget manifest; 06-11 through 06-14's five defect injectors and repair driver"
provides:
  - "scripts/verify-phase6.sh -- the Phase 6 release-cost lane, a peer of verify-phase5.sh"
  - "session.Phase6RequiredControls() / session.VerifyPhase6ControlsAndWork() -- the Phase 6 control-and-work gate"
  - "session.Phase6ExpectedEscapes() -- the five declared, gate-visible Phase 6 escapes"
  - "cmd/lang/main.go isPhase6Corpus dispatch and the `lang stats` seam"
  - "06-DEBT.md -- the phase-close debt register"
affects: []

actuals:
  tokens: 46000
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Phase-gate peer discipline, extended: scripts/verify-phase6.sh copies verify-phase5.sh's skeleton step for step (mktemp/trap, pinned GOCACHE, self-test sentinel, test/race/vet/build, per-corpus JSON capture, control grep loops, expected-escape greps) without forking, extending, or invoking it -- a sixth peer, never a descendant."
    - "Statistics computed in Go, driven thin from shell: the shipped binary gained a `stats` command (measure.Samples.Summary()) that the shell's 20-sample D-06-19 loop pipes raw elapsed_ns samples through, rather than reimplementing sorted-index percentile selection in sed."
    - "The gate assembles already-shipped lanes (risk-lane audit, QLT-02 audit, cache-backed native differential) plus five new, small, single-purpose defect-injection lanes that each prove one injector class fires against its held-out fixture and produces a genuinely rejected mutation -- never a fabricated pass."

key-files:
  created:
    - internal/compiler/session/session_phase6.go
    - internal/compiler/session/session_phase6_test.go
    - internal/compiler/session/session_phase6_escapes.go
    - internal/compiler/session/session_phase6_escapes_test.go
    - cmd/lang/main_test.go
    - scripts/verify-phase6.sh
    - .planning/phases/06-agent-feedback-and-performance-ratification/06-DEBT.md
  modified:
    - cmd/lang/main.go
    - internal/compiler/session/session_phase6_pin_test.go
    - .planning/REQUIREMENTS.md

decisions:
  - "isPhase6Corpus is checked BEFORE isPhase5Corpus in runVerify's dispatch chain, so a Phase 6 corpus (marked by heldout_match_defect.lang) is never accidentally swallowed by an earlier check; a directory with neither marker falls through unchanged to the default VerifyCorpusFile path."
  - "The `lang stats` command is a small, genuinely new CLI seam (reads newline-separated int64 samples from stdin, prints measure.Samples.Summary() as one line of JSON) added specifically so the shell gate's own 20-sample loop never reimplements percentile/CoV arithmetic in sed -- the statistical logic stays the same Go code internal/compiler/measure's own test suite already covers."
  - "The Phase 6 control-and-work gate is driven against testdata/phase1's toggle.lang for the native-differential control (the only corpus carrying that shared fixture) and against testdata/phase6's/testdata/phase4's fixtures for the five defect-injection controls, rather than requiring one dedicated corpus to carry every fixture the gate needs."
  - "All five requirements this phase set out to deliver (FND-04, DX-02, DX-03, DX-04, QLT-02) are marked complete by this plan via requirements.ready-ids/mark-complete, since it is the phase's final plan and the readiness check reported all five ready with none blocked."

coverage:
  - id: D1
    description: "scripts/verify-phase6.sh is a runnable Phase 6 release-cost lane that runs green end-to-end, and is executed on every push and pull request on both host priorities"
    requirement: "DX-04"
    verification:
      - kind: e2e
        ref: ".github/workflows/ci.yml#phase-gate -- runs `sh scripts/verify-phase6.sh` on ubuntu-latest and macos-latest for every push and pull request"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_test.go#TestCIWorkflowRunsPhase6Gate"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_test.go#TestPhase6VerifierScriptContract"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_test.go#TestPhase6ScriptInvokesNoPriorGate"
        status: pass
    human_judgment: false
  - id: D2
    description: "The gate script and its Go sources cannot drift apart in either direction: required controls, pinned bounds, and declared escapes are each asserted against the script's own text"
    requirement: "DX-04"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_test.go#TestPhase6RequiredControlsMatchScript"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_test.go#TestPhase6BoundsMatchScript"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_escapes_test.go#TestPhase6EscapeGrepsMatchScript"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_escapes_test.go#TestPhase6EscapesAreNeverPresentedAsControls"
        status: pass
    human_judgment: false
  - id: D3
    description: "The Phase 6 control-and-work gate, its corpus dispatch, and the `lang stats` sampling seam are each reachable and correct through the shipped binary, not only in process"
    requirement: "QLT-02"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_test.go#TestVerifyPhase6ControlsAndWork"
        status: pass
      - kind: unit
        ref: "cmd/lang/main_test.go#TestPhase6CorpusDispatchRequiresMarker"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_test.go#TestPhase6SamplingLoopMatchesGoStatistics"
        status: pass
      - kind: e2e
        ref: "internal/compiler/testsupport/cli_phase6_test.go#TestStatsCLIComputesGoStatistics"
        status: pass
    human_judgment: false

metrics:
  duration: ~40 min
  completed: 2026-09-07
status: complete
---

# Phase 6 Plan 15: Phase 6 Release Gate, Escape Register, and Debt Close Summary

Closed Phase 6 with a runnable, peer release-cost lane (`scripts/verify-phase6.sh`), a
control-and-work gate reachable through the shipped binary, a declared escape register
that can never masquerade as a solved control, and a dated debt register recording every
honest gap this phase leaves open.

## What Was Built

**Task 1 — the Phase 6 control set, corpus dispatch, and a runnable gate.**
`internal/compiler/session/session_phase6.go` adds `Phase6RequiredControls()` (twelve
identifiers: the risk-lane registry audit's two, the QLT-02 budget audit's four, the
cache-backed native differential's one, and DX-04's five defect-injection controls) and
`VerifyPhase6ControlsAndWork(ctx)`, mirroring `VerifyPhase5ControlsAndWork`'s shape with
its own local `addLane` closure. The gate assembles already-shipped machinery — the
risk-lane registry audit (06-05), the QLT-02 budget-manifest audit (06-09), and the
cache-backed native differential lane (06-04/06-07, driven here over `testdata/phase1`'s
shared `toggle.lang` fixture) — alongside five small, new defect-injection lanes, one per
DX-04 injector class, each proving that class's injector fires against its held-out
fixture (`testdata/phase6/heldout_*.lang` for match/move/borrow, `testdata/phase4`'s
`acquire_three_success.lang` compiled to real generated C for cleanup, and
`testdata/phase6/stale_evidence_subject.lang` through a real `evidence.Build`/
`evidence.Validate` round-trip for stale-evidence) and produces a genuinely rejected
mutation, never a silent unmutated pass.

`cmd/lang/main.go` gained `isPhase6Corpus` (marker fixture `heldout_match_defect.lang`,
checked before `isPhase5Corpus` so a Phase 6 corpus is never swallowed by an earlier
check) dispatching `lang verify testdata/phase6` to the new gate, and a `lang stats`
command reading newline-separated samples from stdin and printing
`measure.Samples.Summary()` as JSON — Task 2's statistics seam, added in the same file
edit since both changes touch `main.go`'s flat dispatcher.

**Task 2 — pinning the script against Go, including the 20-sample loop.**
`scripts/verify-phase6.sh` is a byte-frozen-file-respecting peer of `verify-phase5.sh`:
the same `mktemp`/`trap`/pinned-`GOCACHE` skeleton, the same self-test sentinel (extended
with two new assert-go-tests invocations — `./internal/compiler/session` and
`./cmd/lang`, since `TestPhase6CorpusDispatchRequiresMarker` lives beside `isPhase6Corpus`
in the `main` package rather than in `session`), `go test`/`-race`/`vet`/build, per-corpus
JSON captures for phase1 through phase6 (sanitizer options pinned byte-identically on the
phase5 capture, unchanged), the existing prior-phase control grep loops verbatim, Phase
6's own twelve-control loop, five expected-escape greps, and the D-06-19 20-sample warm
loop over `verify testdata/phase6` — piping raw `elapsed_ns` samples through `lang stats`
rather than `sed`/`sort` percentile indexing. `session_phase6_test.go` proves the gate
fires every required control with nonzero work, and that the script and Go can never
silently drift apart in either direction (control set, sample-count/CoV/explain/query
bound values, and the sampling loop's own percentile/CoV computation).

**Task 3 — the escape register and the debt register.**
`internal/compiler/session/session_phase6_escapes.go` declares
`Phase6ExpectedEscapes()`: D-06-13's four cache-soundness holes
(`escape:cache-undeclared-environment`, `escape:cache-clang-version-string-stable`,
`escape:cache-nondeterministic-codegen`, `escape:cache-directory-hand-edited`) plus
D-06-29's residual overfitting debt (`escape:repair-heldout-corpus-residual-overfitting`).
`session_phase6_escapes_test.go` proves these are exactly the declared set, that their
intersection with `Phase6RequiredControls()` is empty (an escape can never be presented as
a solved control), and that the script's own escape greps are set-equal with the Go list
in both directions. `06-DEBT.md` records D-06-13, D-06-20 (deferred peak-RSS), D-06-29,
the four unresolved `unclassified`-category edge probes (DX-02, DX-03, DX-04, QLT-02)
carried across all fourteen prior plans without resolution, D-06-30 (the recorded
agent-legibility exercise, explicitly not run as part of this plan's close), the
D-03-02/D-05-41 carry-forward, and the closure record for the Phase 5 baseline-machine
blocker (closed by 06-09).

## Verification

`sh scripts/verify-phase6.sh` runs green end-to-end on the real tree: `go test ./...`,
`go test -race ./...`, and `go vet ./...` are clean; all six corpora produce their
required controls; every Phase 6 control fires with nonzero `RecomputedWork`; all five
declared escapes appear in the gate's `expected_escapes` output; the 20-sample loop
reports `p50`/`p95`/`cov` via `lang stats`. `TestVerifyPhase6ControlsAndWork` and the
full `TestPhase6*` suite pass. `scripts/verify-phase5.sh` is byte-unchanged by this plan
(no hunks touch it).

Requirements readiness was checked via `requirements.ready-ids` against this plan's
declared `[FND-04, DX-02, DX-03, DX-04, QLT-02]`: all five reported ready, none blocked,
and all five were marked complete via `requirements.mark-complete`.

## Deviations from Plan

**1. [Sequencing] `session_phase6_escapes.go` (plan's Task 3 file) was built alongside
`session_phase6.go` (Task 1) rather than strictly after it.** `VerifyPhase6ControlsAndWork`'s
own `result.ExpectedEscapes` wiring is a hard compile dependency on
`Phase6ExpectedEscapes()`, so the two files were authored together and committed in the
same (first) commit. `session_phase6_escapes_test.go` still lands in the third commit,
matching the plan's own Task 3 file list for the test half.

**2. [Rule 1 — pre-existing pin, mechanically updated] `TestLaneSchemaLiteralSiteCountIsPinned`
(`session_phase6_pin_test.go`) required updating from 15 to 16.** This plan's `addLane`
closure in `session_phase6.go` adds one new, legitimate `protocol.LaneSchema1` composite-
literal site — never a `/0` literal moved, the same species of addition 06-07's own
deviation note already established for `session_phase6_verify.go`'s three sites. The
per-file map and total were updated to include `session_phase6.go: 1` (16 total),
following that same precedent's reasoning exactly.

**3. [Deliberate simplification, documented] The cleanup and stale-evidence defect-injection
lanes prove their injector fires and produces a real mutation/mismatch, without re-running
a full native compile+sanitize cycle or a full `lang evidence --validate` CLI round-trip
inside the gate itself.** The cleanup lane compiles `testdata/phase4/acquire_three_success.lang`
to real generated C and asserts `CleanupInjector.Inject` removes a genuine
`lang:release-site` marker (distinct mutated bytes), rather than also driving the mutated C
through a sanitized native run — that full behavioral proof already lives in the `go test`
suite (`session_phase6_injectors_test.go`, part of `go test ./...`, which
`scripts/verify-phase6.sh` also runs). The stale-evidence lane builds a real
`evidence.Build`/`evidence.Validate` round-trip (not a CLI subprocess) using the same
production functions `lang evidence`/`lang evidence --validate` themselves call. Both
choices keep the gate's own runtime bounded and avoid re-deriving proof burden the phase's
existing test suite already carries, while still exercising genuinely shipped code paths
rather than a fabricated pass.

No auth gates encountered. No architectural changes required (Rule 4 not triggered).

## Known Stubs

None — every lane this plan adds drives real shipped code paths (`cache.Consult`,
`evidence.Build`/`Validate`, the five DX-04 injectors, the risk-lane/QLT-02 audits)
against real fixtures, never a hardcoded or mocked verdict.

## Self-Check: PASSED

- `internal/compiler/session/session_phase6.go` — FOUND
- `internal/compiler/session/session_phase6_escapes.go` — FOUND
- `internal/compiler/session/session_phase6_test.go` — FOUND
- `internal/compiler/session/session_phase6_escapes_test.go` — FOUND
- `cmd/lang/main_test.go` — FOUND
- `scripts/verify-phase6.sh` — FOUND (executable)
- `.planning/phases/06-agent-feedback-and-performance-ratification/06-DEBT.md` — FOUND
- Commit `70c448e` — FOUND in `git log --oneline`
- Commit `e7fecab` — FOUND in `git log --oneline`
- Commit `3f023f0` — FOUND in `git log --oneline`
