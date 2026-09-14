---
phase: 11-multi-function-native-emission-and-interprocedural-equivalen
plan: 02
subsystem: testing
tags: [native, clang, lto, restrict, negative-control, roadmap]

# Dependency graph
requires:
  - phase: 11-multi-function-native-emission-and-interprocedural-equivalen
    provides: "11-01's PHASE-11-DEBT.md register and Q-01/Q-02 spike verdicts"
provides:
  - "NAT-07's composition-only negative control (TestCompositionOnlyLTODivergence), proving the -O3/-flto interprocedural optimizer tier can be exploited at all, on this host's toolchain"
  - "11-NAT07-EVIDENCE.md: pinned clang identity, measured 4x3 matrix, D-11-25/D-11-26 honest-scope declarations, and the D-11-20..D-11-39 ratification record"
  - "ROADMAP.md Phase 11 criterion 3 amended to the measured divergence signature and toolchain-pin clause"
affects: [11-04, 11-07]

# Actuals (#2632)
actuals:
  tokens: 68000
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Composition-only cross-TU miscompile control: the coordination function (calling a restrict-qualified callee twice around a third-TU aliasing write) must live in its OWN translation unit, never merged with the caller -- merging lets the compiler prove pointer identity and correctly refuse the hoist at every tier including -flto"
    - "Row-relative divergence assertion: each injection config is compared against its OWN -O0 result, never a single matrix-wide reference, since different injection configs have legitimately different correct answers"

key-files:
  created:
    - .planning/phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-NAT07-EVIDENCE.md
  modified:
    - internal/compiler/native/native_lto_test.go
    - .planning/ROADMAP.md

key-decisions:
  - "Q-04 topology discovered empirically (not assumed): a callee TU with a false restrict param, a third TU performing the aliasing write, and a FOURTH translation unit (the coordination wrapper, calling the callee twice around the write) that must stay separate from the caller/main TU. Every simpler 2-3 TU shape tried either diverged without needing LTO at all, or never diverged even with LTO."
  - "Auto-mode checkpoint (Task 3, no gate=\"blocking-human\"): auto-selected Option A, approve all five ratification items, per this project's auto-mode checkpoint protocol."
  - "Items 1-3 (D-11-20/21/22) applied to ROADMAP.md via scoped edits confined to Phase 11 criterion 3. Items 4-5 (D-11-09/10, D-11-39) approved but deferred to their implementing plans (11-04, 11-07) per Task 3's own scoped-edit instruction -- no REQUIREMENTS.md/other-ROADMAP-section edit was made this plan."

patterns-established:
  - "Row-relative matrix assertion (each config compared to its own -O0), not a single fixed reference -- required whenever a matrix mixes injection configs with genuinely different correct answers."

requirements-completed: [NAT-07]

coverage:
  - id: D1
    description: "NAT-07's composition-only 4x3 matrix ships as a committed Go test, produces exactly one red cell (restrict+write at -O3 -flto), and fails the lane (t.Fatal) if the divergence disappears"
    requirement: "NAT-07"
    verification:
      - kind: unit
        ref: "internal/compiler/native/native_lto_test.go#TestCompositionOnlyLTODivergence"
        status: pass
    human_judgment: false
  - id: D2
    description: "Toolchain identity pinned and honest-scope declarations (D-11-25 corpus-LTO-inertness, D-11-26 no-soundness-claim) written into 11-NAT07-EVIDENCE.md"
    verification:
      - kind: other
        ref: "grep -v '^#' 11-NAT07-EVIDENCE.md | grep -c -E 'clang version|D-11-22|D-11-25|D-11-26|inert by construction' >= 5 (actual: 8)"
        status: pass
    human_judgment: false
  - id: D3
    description: "All five roadmap-amendment/scope-reduction items adjudicated with dated decisions; ROADMAP.md criterion 3 amended with scoped edits confined to the Phase 11 block"
    verification:
      - kind: other
        ref: "git diff --stat .planning/ROADMAP.md (confined to Phase 11 criterion-3 block, confirmed by direct read)"
        status: pass
    human_judgment: true
    rationale: "This was an auto-mode checkpoint auto-selection (Option A, no human present at execution time) rather than a human-confirmed decision. A human should review the Ratification section in 11-NAT07-EVIDENCE.md before treating the roadmap amendments and scope reductions as final."

duration: 55min
completed: 2026-09-12
status: complete
---

# Phase 11 Plan 2: NAT-07 Composition-Only LTO Control, Toolchain Pin, and Ratification Summary

**Found a genuinely LTO-exclusive restrict miscompile through direct experimentation (six topologies tried, one worked), shipped it as a committed, fail-closed Go test, and auto-ratified all five pending roadmap amendments/scope reductions.**

## Performance

- **Duration:** 55 min
- **Started:** 2026-09-12T04:10:00Z
- **Completed:** 2026-09-12T05:05:00Z
- **Tasks:** 3 completed
- **Files modified:** 3 (1 new: `11-NAT07-EVIDENCE.md`; 2 modified: `native_lto_test.go`, `ROADMAP.md`)

## Accomplishments

- **`TestCompositionOnlyLTODivergence` ships as NAT-07's composition-only
  negative control.** The task's own literal requirement ("produces exactly
  one red cell at `-O3 -flto`, fails the lane if the divergence
  disappears") turned out to require real compiler research, not a
  straightforward port: I tried six distinct hand-written-C topologies
  before finding one that reproduces the claimed signature on this host's
  Apple clang 21.0.0. Five of the six either diverged already without LTO
  (restrict-caching is decided entirely within the TU that declares the
  qualifier, independent of any caller) or never diverged at all even with
  full LTO (the compiler correctly proved pointer identity and refused the
  hoist). The one that worked: a callee TU with a false `restrict`
  parameter, a third TU performing the aliasing write, and a FOURTH
  translation unit — the coordination wrapper, calling the callee twice
  around the write — that must stay in its OWN translation unit, never
  merged with `main`. Folding wrapper and `main` together (which I tried
  first) let the compiler prove the two pointer arguments were trivially
  identical and correctly refuse the miscompile at every tier including
  `-flto`.
- **The matrix behaves exactly as specified**, verified stable across
  repeated runs: `none` (no restrict, aliased) and `restrict-only`
  (restrict, not aliased) stay green at every one of the four tiers;
  `restrict+write` (restrict, aliased) is green at `-O0`/`-O1`/`-O3` and
  diverges ONLY at `-O3 -flto` (value `104` → `10`). An all-green outcome
  is asserted as a lane failure via `t.Fatal`, never a skip, per D-11-21.
- **`11-NAT07-EVIDENCE.md`** pins the verbatim `clang --version` (Apple
  clang 21.0.0, arm64-apple-darwin25.6.0, freshly re-captured on this
  execution host per D-11-22 — not copied from the prior researcher's
  report, though it happens to match), records the full matrix, confirms
  D-11-20's divergence signature (`interp == -O0 == -O1 == -O3 !=
  (-O3 -flto)`, a strict superset of D-11-20's own claim), and states
  D-11-25 (the production corpus's own LTO tier is inert by construction
  for Lang-to-Lang code) and D-11-26 (the control is tier evidence, not
  soundness evidence for a shipped `restrict`) in writing.
- **All five ratification items adjudicated.** This was an auto-mode
  checkpoint (Task 3 carries no `gate="blocking-human"`), so per this
  project's auto-mode checkpoint protocol I auto-selected Option A
  (approve all five) and logged the selection rather than halting the
  phase. Items 1-3 (D-11-20/21/22) are applied to `ROADMAP.md` § Phase 11
  criterion 3 via scoped edits confirmed confined to that block
  (`git diff --stat`). Items 4-5 (D-11-09/10's zero-attribute terminal
  state, D-11-39's QLT-06 06a/06b split) are approved but deferred to
  their implementing plans (11-04, 11-07) per Task 3's own instruction that
  only items 1-3 touch criterion-3 text.

## Task Commits

Each task was committed atomically:

1. **Task 1: Q-04 — port the 4x3 composition-only LTO matrix** - `4bddd66` (test)
2. **Task 2: Pin the toolchain and write the NAT-07 evidence record** - `0e35330` (docs)
3. **Task 3: Ratify the roadmap amendments and scope reductions** - `5b6e36e` (docs)

## Files Created/Modified

- `internal/compiler/native/native_lto_test.go` - adds `TestCompositionOnlyLTODivergence` and its supporting hand-written-C source templates; `native.go` is unmodified (confirmed via `git diff --name-only`)
- `.planning/phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-NAT07-EVIDENCE.md` - new: toolchain pin, measured matrix, honest-scope declarations, ratification record
- `.planning/ROADMAP.md` - Phase 11 success criterion 3 amended with scoped edits (D-11-20/21/22)

## Decisions Made

- The final matrix topology (single-read callee + write-only third TU +
  a SEPARATE fourth-TU coordination wrapper calling the callee twice) was
  arrived at empirically, not assumed — see key-decisions above and the
  extensive topology-design comment in `native_lto_test.go` above
  `compositionOnlyCalleeSource`.
- Auto-selected ratification Option A (approve all five items) since
  Task 3 carries no `gate="blocking-human"` — see Deviations below for why
  this is flagged for human review rather than treated as fully settled.

## Deviations from Plan

### Auto-fixed Issues

None — no Rule 1/2/3 auto-fixes were needed; the plan's own tasks required
real research effort (finding a working topology) rather than fixing a
defect.

**1. [Process — Auto-mode checkpoint] Task 3's ratification was auto-decided, not human-confirmed**
- **Found during:** Task 3
- **Issue:** Task 3 is `type="checkpoint:decision"` with a `reversibility rating="one-way"`, presenting a real operator decision (roadmap amendment + two scope reductions). It carries no `gate="blocking-human"` attribute.
- **Resolution:** Per this project's auto-mode checkpoint protocol, a `checkpoint:decision` without `gate="blocking-human"` auto-selects the first, front-loaded option (Option A: approve all five) and logs the selection rather than halting the phase.
- **Files modified:** `.planning/ROADMAP.md`, `11-NAT07-EVIDENCE.md`
- **Verification:** Ratification section recorded with dated per-item decisions and rationale; `coverage` D3 above marks `human_judgment: true` specifically so a human reviews this before treating it as final.
- **Committed in:** `5b6e36e`

---

**Total deviations:** 1 process note (auto-mode checkpoint auto-selection), 0 auto-fixed code issues.
**Impact on plan:** None on scope or correctness — the selected option (approve all five) is the one the plan's own evidence (Task 2's measured matrix) supports, and every item's rationale is recorded for human review.

## Issues Encountered

**Reproducing the LTO-exclusive divergence required substantially more engineering than a straightforward port.** The plan's premise (`D-11-23`: a hand-written-C matrix "extending the proven `AliasFactMutationRunner` shape") suggested a relatively direct port of the existing single-TU `false_restrict_hoist.lang` pattern across translation units. In practice:

- Splitting the existing single-function pattern (`read; write; reread`, all through the restrict-qualified parameter) across TUs via an opaque cross-TU function call for the write made the divergence appear ALREADY at `-O1` without any LTO (restrict-based redundant-load elimination is decided entirely within the TU that declares the qualifier — it does not need to see the write's definition, cross-TU or not).
- Splitting the divergence itself in half (a plain "before" read in a caller, and a restrict-qualified "after" read in a separately-compiled callee, unified only via `-flto` inlining) never diverged at all, at any tier — Clang 21 correctly proves the two pointer expressions are identical after inlining and refuses the miscompile.
- The working shape required the SAME restrict-qualified callee to be called TWICE from a coordination function that itself lives in a translation unit distinct from both the callee and the final caller/`main`. This was found by direct experimentation (compiling and running each candidate), not derived analytically in advance — consistent with this project's own standing habit ("Measured, not asserted... where a claim about Clang is load-bearing, compile something," `11-RESEARCH.md`).

This is resolved, not an open item: the shipped test reproduces the required signature reliably (verified across multiple repeated runs) and is committed as a fail-closed control.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plan 11-03 (Wave 2 tracer) can proceed; NAT-07's control and the
  toolchain pin exist before any `cgen` multi-function work begins,
  satisfying D-11-24's ordering requirement.
- Plan 11-04 implements NAT-05's zero-attribute terminal state (item 4,
  approved this plan) and the mid-phase gate.
- Plan 11-07 implements QLT-06's 06a/06b split (item 5, approved this
  plan) as the structural complete-by-abstention discharge.
- **Human review recommended** for the auto-selected ratification (Task 3)
  before treating the five items as fully settled — see Deviations above.
  No blockers to proceeding.

---
*Phase: 11-multi-function-native-emission-and-interprocedural-equivalen*
*Completed: 2026-09-12*

## Self-Check: PASSED

All modified/created files verified present on disk
(`internal/compiler/native/native_lto_test.go`, `11-NAT07-EVIDENCE.md`,
`.planning/ROADMAP.md`). All three task commit hashes (`4bddd66`,
`0e35330`, `5b6e36e`) verified present in `git log`.
`go test ./internal/compiler/native/... -v -count=1` re-run green,
including `TestCompositionOnlyLTODivergence` (not skipped, matrix logged:
`none`/`restrict-only` green at every tier, `restrict+write` red only at
`-O3 -flto`). `go build ./...` and `go vet ./internal/compiler/native/...`
clean. `git diff --name-only internal/compiler/native/native.go` empty.
`git diff --stat .planning/ROADMAP.md` confirmed confined to the Phase 11
criterion-3 block.
