---
phase: 05-native-equivalence-and-adversarial-evidence
plan: 14
subsystem: testing
tags: [go, native-equivalence, adversarial-evidence, gate, roadmap, debt-register]

# Dependency graph
requires:
  - phase: 05-native-equivalence-and-adversarial-evidence (05-09)
    provides: "Phase5RequiredControls()/Phase5ExpectedEscapes()/VerifyPhase5ControlsAndWork and scripts/verify-phase5.sh's own required-control block, extended in this plan to their final 15/2 shape"
  - phase: 05-native-equivalence-and-adversarial-evidence (05-10, 05-11)
    provides: "the three reducer vacuity controls (control:reduce.*) and the two QLT-01 registry controls (control:qlt01.*), left deliberately unwired until this plan"
  - phase: 05-native-equivalence-and-adversarial-evidence (05-12, 05-13)
    provides: "lang.mismatch/0 and escape:coordinated-source-to-core-false-claim, both explicitly deferred to this plan for final wiring"
provides:
  - "Phase5RequiredControls() at its final 15 identifiers; Phase5ExpectedEscapes() at its final 2; scripts/verify-phase5.sh in lockstep with both"
  - "TestPhase5RequiredControlsIsFifteen and TestPhase5EveryDeclaredControlActuallyFires — two independent assertions catching different regression shapes"
  - "The developer-confirmed, one-way OpCall/interprocedural-equivalence deferral to M002, recorded in .planning/ROADMAP.md's new M002 Charter section"
  - ".planning/phases/05-native-equivalence-and-adversarial-evidence/05-DEBT.md's D-05-42 entry (D-05-33's verbatim wording), six named residual limitations, and the D-04-26/D-03-01/D-04-33 closure record"
  - "A fully green final Phase 5 gate: sh scripts/verify-phase5.sh, go build/vet/test ./..., go test -race ./..., and zero diff over every frozen prior-phase artifact"
affects: [06]

actuals:
  tokens: 6900
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Two independent regression assertions per control-set change (a count pin and a set-equality/fired-status check) so a dropped control and a divergent-but-same-count control fail differently."
    - "A milestone-scope deferral is recorded in two places at once (ROADMAP.md charter entry + phase debt register verbatim entry), never left as an implicit carry."

key-files:
  created: []
  modified:
    - internal/compiler/session/session_phase5.go
    - internal/compiler/session/session_phase5_test.go
    - scripts/verify-phase5.sh
    - .planning/ROADMAP.md
    - .planning/phases/05-native-equivalence-and-adversarial-evidence/05-DEBT.md

key-decisions:
  - "DECISION (developer-confirmed, one-way for this milestone): OpCall, interprocedural loan liveness in both admission layers, call-graph construction, cycle refusal, a bounded interpreter call stack, and the cross-function rebuild of Phase 3's exhaustive differentials are OUT of M001 and become M002's LEAD charter item (D-05-32/D-05-33), gated on callable subset-of publishable (D-04-03)."
  - "Accepted consequence 1: M001 ships WITHOUT Lang-to-Lang calls."
  - "Accepted consequence 2: interprocedural -O3 equivalence is outside M001's proof scope BY CONSTRUCTION, since M001 ships without calls -- not an accident of SC1's wording but a direct structural consequence of this deferral."
  - "Accepted consequence 3: D-03-02 remains OPEN past the milestone (unreachable-but-unfixed; lift condition unchanged per D-04-03). 03-DEBT.md's own historical closure record (03-09, the single-function publication-path fix) stands unedited and is not reverted; what re-opens in 05-DEBT.md is the interprocedural half of the same hazard class, unreachable today with no OpCall and newly load-bearing the moment M002 lands it."
  - "The deferral is recorded in TWO places (ROADMAP.md's new M002 Charter section and 05-DEBT.md's D-05-42), per D-05-33's explicit instruction that silence is the real footgun -- no phase in the current roadmap otherwise owns this work, since Phase 6 is agent feedback and performance ratification, not language surface."

requirements-completed: [INT-02, NAT-02, NAT-03, QLT-01]

coverage:
  - id: D1
    description: "All fifteen Phase 5 controls and both escapes are declared in Go and the shell gate in lockstep; every declared control is proven to actually fire (not merely declared)"
    requirement: QLT-01
    verification:
      - kind: unit
        ref: "internal/compiler/session#TestPhase5RequiredControlsMatchScript"
        status: pass
      - kind: unit
        ref: "internal/compiler/session#TestPhase5RequiredControlsIsFifteen"
        status: pass
      - kind: unit
        ref: "internal/compiler/session#TestPhase5EveryDeclaredControlActuallyFires"
        status: pass
      - kind: unit
        ref: "internal/compiler/session#TestBothPhase5EscapesAreVisible"
        status: pass
    human_judgment: false
  - id: D2
    description: "The one-way OpCall/interprocedural-equivalence scope decision is developer-confirmed and recorded verbatim (D-05-33's exact wording) in both .planning/ROADMAP.md's M002 Charter section and 05-DEBT.md's D-05-42 entry, including all three accepted consequences stated plainly"
    requirement: QLT-01
    verification:
      - kind: manual_procedural
        ref: "human decision recorded in <user_response> of this continuation's prompt: defer-to-m002, confirmed"
        status: pass
    human_judgment: true
    rationale: "This is a one-way milestone-scope decision (D-05-32) that the plan's own acceptance criteria require a human, not an automated check, to confirm before any roadmap/debt-register edit is made."
  - id: D3
    description: "The full Phase 5 gate is green with no prior-phase artifact changed: sh scripts/verify-phase5.sh, go build/vet/test ./..., go test -race ./..., and zero diff over every frozen prior-phase corpus/script/foreign-TU path"
    requirement: NAT-02
    verification:
      - kind: integration
        ref: "sh scripts/verify-phase5.sh (exit 0, all 5 verify-command JSON results status:pass)"
        status: pass
      - kind: integration
        ref: "go build ./... && go vet ./... && go test ./... && go test -race ./... (all exit 0)"
        status: pass
      - kind: integration
        ref: "git diff --stat over testdata/phase1-4, scripts/verify-phase1-4.sh, native/lang_foreign_resource.c, native/lang_foreign_nonlocal.c, internal/compiler/session/session.go (empty)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Every named residual limitation and every closed carried-debt item (D-04-26, D-03-01, D-04-33) is recorded explicitly in 05-DEBT.md, not absorbed into prose"
    requirement: QLT-01
    verification:
      - kind: other
        ref: "grep -c 'Interprocedural equivalence is explicitly outside M001' .planning/phases/05-native-equivalence-and-adversarial-evidence/05-DEBT.md (returns 1)"
        status: pass
      - kind: other
        ref: "grep -c 'OpCall' .planning/ROADMAP.md (returns >=1)"
        status: pass
    human_judgment: false

duration: ~35min (Task 2/3 continuation only; Task 1 by prior executor ~unrecorded)
completed: 2026-09-06
status: complete
---

# Phase 5 Plan 14: Final Gate, M002 Charter, and the OpCall Deferral Summary

**Phase 5's required-control set reaches its final 15/2 shape (Task 1, prior executor), the developer confirms `OpCall` and interprocedural `-O3` equivalence are OUT of M001 and become M002's lead charter item, and the full Phase 5 gate — shell script, build/vet/test, race, and every frozen prior-phase artifact — runs green to close the phase.**

## Performance

- **Duration:** ~35 min (this continuation's Task 2 confirmation + Task 3 recording/gate; Task 1 was executed and committed by a prior executor session)
- **Started:** 2026-09-06T18:00:00Z (approx, first file read)
- **Completed:** 2026-09-06T18:35:00Z
- **Tasks:** 3 total (1 prior + 2 this continuation)
- **Files modified:** 5 total across the plan (3 from Task 1, 2 from Task 2/3)

## Accomplishments

- **Task 1 (prior executor, commit `7ff8ca8`):** `Phase5RequiredControls()` extended from 10 to 15 identifiers (the three reducer vacuity controls and the two QLT-01 registry controls), `Phase5ExpectedEscapes()` extended to both Phase 5 escapes, `VerifyPhase5ControlsAndWork` wired to invoke `VerifyQLT01Registry`/`VerifyCoordinatedLieEscape`/the reducer lane, and `scripts/verify-phase5.sh`'s own control block updated in the same commit. `TestPhase5RequiredControlsIsFifteen` and `TestPhase5EveryDeclaredControlActuallyFires` added as two independently-failing regression assertions.
- **Task 2 (this continuation, checkpoint resumed):** The developer confirmed **defer-to-m002** — `OpCall`, interprocedural loan liveness in both admission layers, call-graph construction, cycle refusal, a bounded interpreter call stack, and the cross-function rebuild of Phase 3's exhaustive differentials stay OUT of M001 and become M002's LEAD charter item. The confirmation is recorded verbatim above in `key-decisions` and was obtained BEFORE any roadmap/debt-register edit, per the plan's own acceptance criteria.
- **Task 3 (this continuation):** `.planning/ROADMAP.md` gained a new "M002 Charter (deferred from Phase 5, D-05-32/D-05-33)" section naming `OpCall` and every deferred item explicitly, plus Phase 5's status flipped to complete (14/14 plans). `.planning/phases/05-native-equivalence-and-adversarial-evidence/05-DEBT.md` gained D-05-42 (D-05-33's verbatim sentence plus all three accepted consequences stated plainly), the six named residual limitations, and the closure record for D-04-26/D-03-01 (plan 05-02) and D-04-33 (plan 05-01). The full gate then ran green: `sh scripts/verify-phase5.sh` (exit 0, all five lane-groups pass), `go build ./...`, `go vet ./...`, `go test ./...`, `go test -race ./...` (all exit 0), and `git diff --stat` over every frozen Phase 1-4 corpus path, gate script, foreign TU, and `session.go` itself (empty on all).

## Task Commits

Each task was committed atomically:

1. **Task 1: Complete the required-control and escape sets in Go and the shell gate together** - `7ff8ca8` (feat) — completed by a prior executor session before this continuation began
2. **Task 2: Confirm the one-way scope decision — OpCall stays out of M001** - checkpoint decision, no code commit (decision recorded in this summary and in Task 3's commit message)
3. **Task 3: Record the M002 charter and debt-register entries, then run the final green gate** - `c09220c` (docs)

**Plan metadata:** committed as part of this SUMMARY's own commit (see below).

## Files Created/Modified

- `internal/compiler/session/session_phase5.go` — (Task 1) `Phase5RequiredControls()`/`Phase5ExpectedEscapes()`/`VerifyPhase5ControlsAndWork` at final 15/2 shape
- `internal/compiler/session/session_phase5_test.go` — (Task 1) `TestPhase5RequiredControlsIsFifteen`, `TestPhase5EveryDeclaredControlActuallyFires`
- `scripts/verify-phase5.sh` — (Task 1) required-control block extended to 15, in lockstep with Go
- `.planning/ROADMAP.md` — (Task 3) new M002 Charter section naming `OpCall`; Phase 5 marked complete (14/14)
- `.planning/phases/05-native-equivalence-and-adversarial-evidence/05-DEBT.md` — (Task 3) D-05-42, six named residuals, D-04-26/D-03-01/D-04-33 closure record

## Decisions Made

See `key-decisions` in frontmatter: the developer-confirmed, one-way `defer-to-m002` decision and its three accepted consequences, stated plainly rather than softened, and recorded in two places (ROADMAP.md + 05-DEBT.md) per D-05-33's explicit instruction against silent, unowned carries.

## Deviations from Plan

None — plan executed exactly as written across both this continuation's tasks and the prior executor's Task 1. The checkpoint at Task 2 was answered by the human exactly as the plan required (no roadmap/debt-register edit was made before the confirmation), and no `replan-with-opcall` amendment was requested.

## Issues Encountered

None.

## User Setup Required

None — no external service configuration required.

## Known Stubs

None. This plan is documentation/registration plus a wiring completion over already-shipped machinery (05-10 through 05-13's reducer, QLT-01 registry, and escape lanes); no new stub surfaces were introduced.

## Next Phase Readiness

- **Phase 5 is COMPLETE (14/14 plans).** All four Phase 5 success criteria hold: (1) interpreter/`-O0`/`-O3` agreement across the milestone corpus, (2) all seven native hostile mutations plus applicable ownership/certificate mutations detected, (3) ASan/UBSan evidence isolated from semantic-equivalence evidence, (4) any injected mismatch reports a minimized case via `lang.mismatch/0`, with the coordinated source-to-core false claim remaining a documented, demonstrable escape.
- **Phase 6 (Agent Feedback and Performance Ratification) is next.** Its charter does NOT include `OpCall` or interprocedural equivalence — those are explicitly M002's lead item, not Phase 6's, per this plan's recorded decision. Phase 6 planning should read `.planning/ROADMAP.md`'s new "M002 Charter" section to avoid re-absorbing this deferred work.
- **M002 planning (post-M001) should start from**: `.planning/ROADMAP.md`'s M002 Charter section, `05-DEBT.md`'s D-05-42, and `04-CONTEXT.md`'s D-04-03 (the callable ⊆ publishable lift condition) as the three canonical entry points.
- **D-05-41 remains open, unwidened**, as recorded in `05-DEBT.md` at the mid-phase gate — unrelated in cause and scope to this plan's D-05-42.

## Self-Check: PASSED

- `internal/compiler/session/session_phase5.go` — FOUND
- `internal/compiler/session/session_phase5_test.go` — FOUND
- `scripts/verify-phase5.sh` — FOUND
- `.planning/ROADMAP.md` — FOUND, contains "OpCall" (M002 Charter section)
- `.planning/phases/05-native-equivalence-and-adversarial-evidence/05-DEBT.md` — FOUND, contains "Interprocedural equivalence is explicitly outside M001's proof scope" verbatim
- Commits `7ff8ca8`, `c09220c` — both present in `git log --oneline --all`
- Full gate re-confirmed green in this session: `go build ./...` (clean), `sh scripts/verify-phase5.sh` (exit 0), `go vet ./...` (clean), `go test -race ./...` (all packages `ok`), `git diff --stat` over every frozen prior-phase path (empty)

---
*Phase: 05-native-equivalence-and-adversarial-evidence*
*Completed: 2026-09-06*
