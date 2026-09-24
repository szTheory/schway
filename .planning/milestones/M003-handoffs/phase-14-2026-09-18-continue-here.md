---
context: default
phase: 14-evidence-instrument-and-honest-scoping
task: null
total_tasks: null
status: paused
last_updated: 2026-09-18T16:02:56.970Z
---

# BLOCKING CONSTRAINTS — Read Before Anything Else

> These are not suggestions. Each constraint below was discovered through failure.
> Acknowledge each one explicitly before proceeding.

- [ ] CONSTRAINT: verification artifacts are code, not prose — a phase is not green
      until the suite is re-run *after* the verifier writes `NN-VERIFICATION.md`.
      This session marked Phase 14 complete and reported it green while `main` was
      actually red. The post-merge build/test gate ran before `gsd-verifier` rewrote
      `14-VERIFICATION.md`; that rewrite (shipped inside the phase-completion commit
      `511399f`) introduced four elided `go test ... -run 'X'` commands and one
      unresolvable package path, which this repo's own EVD-01 groundedness lint
      correctly rejected. Three guards failed on `main` for ~40 minutes.
      **Structural mitigation:** in `execute-phase`, treat `verify_phase_goal` as a
      step that mutates enforced-tier planning docs, and re-run
      `go test ./internal/compiler/session/ -run 'Groundedness|Reconcil' -count=1`
      after it and before `update_roadmap`. Do not rely on the earlier post-merge gate.

**Do not proceed until all boxes are checked.**

## Critical Anti-Patterns

| Pattern | Description | Severity | Prevention Mechanism |
|---------|-------------|----------|---------------------|
| Gate-before-mutation | A quality gate was run, then a later workflow step mutated the very artifacts the gate covers, and the stale green result was reported as final. Manifested as Phase 14 being declared complete while three groundedness guards were failing on `main`. | blocking | Re-run the planning-doc guard set after any step that writes to `.planning/**` enforced-tier docs — specifically after `verify_phase_goal` and before `update_roadmap`. |
| Full-parallel suite flake | `go test ./...` at default parallelism intermittently fails `TestCacheKeyCoversEveryDeclaredInput` (cache) and `TestPayloadTracerThreeEngineAgreement` (cgen) on this machine. Under contention those packages take ~24s vs ~9s; both pass in isolation. | advisory | Run the full suite with `-p 2` for gate purposes. Do not "fix" these tests in response to a parallel-load failure without first reproducing them serially. |
| Trusting subagent self-reports | Three executors reported "full suite green"; the tree was green when they said so, but the composition of their work plus a later step was not. One executor also self-reported an accidental `git stash push -u`. | advisory | Independently re-run the gate yourself after the last mutating step, and check `git stash list` / `git status --short` at phase close. |

<current_state>
Between phases. Phase 14 (evidence-instrument-and-honest-scoping) is complete,
verified 11/11 requirements, and genuinely green as of `349ede6`. Phase 15
(Event Identity, `lang.execution/2`) has not started — no `15-CONTEXT.md`, no plans.
Working tree is clean apart from untracked `.planning/milestone.lock`.
Branch `main`, tip `e1effab`.
</current_state>

<completed_work>

This session executed the three remaining Phase 14 gap-closure plans and closed the phase:

- 14-11 (`b78b3dd`): completion-witnessed run record — the corpus-wide evidence run can
  no longer report a result it did not finish producing. Budget measured and asserted:
  90.3s against a new 480s ceiling (previously a 300.11s near-miss against 300s).
  Anchored producer/consumer name contract, 3 new seeded-fault proofs. Closes WR-01, WR-02.
- 14-13 (`d88fa43`): the `//go:build` suppression surface is now actually checked rather
  than merely collected — declared allowlist, live check in `suppressionProblems`, scan
  reordered so host-excluded files still surface their own constraint, fourth seeded
  fault. Closes WR-03 / EVD-04's inert branch.
- 14-12 (`4e234b0`): EVD-02 exemption removed; the corpus-wide `>=EXERCISED` bar confirmed
  reflexively against Phase 14's own `14-VALIDATION.md`, with a permanent guard against
  the exemption returning and owner-required row narrowing. Closes EVD-02, PRC-01.
- Code review (`44bcad4`): 41 files, standard depth. 0 critical, 2 warning, 2 info.
- Verification (`511399f`): 11/11 requirements, status `passed`; verifier independently
  re-ran the closure evidence rather than trusting the summaries.
- Regression fix (`349ede6`): repaired the groundedness breakage that `511399f` introduced
  (see BLOCKING CONSTRAINTS above).
- Audit fixes (`e1effab`): refreshed three stale STATE.md fields; marked Phase 14's one
  deferred item `status: resolved` with both closing commits cited.
</completed_work>

<remaining_work>

- Phase 15: Event Identity (`lang.execution/2`) — not started. Goal: give event identity an
  owner. Shared-leaf diamond call graphs collide (D-11-51), `OpCall` emits no event at all,
  and D-12-21 cannot close until both are fixed.
- Phases 16-20 of M003 remain after that.
</remaining_work>

<decisions_made>

- Stopped rather than auto-advancing into Phase 15, even though `workflow.auto_advance=true`
  would have chained discuss -> plan -> execute. The user was asked and chose to stop.
- Treated the first post-merge full-suite failure as a parallel-load flake rather than a
  regression — confirmed by isolation runs and a green `-p 2` serialized run, not assumed.
- Fixed the groundedness regression at its cause. Only the one genuine prose fragment kept
  a pinned-frontier entry (line anchor corrected 86 -> 90); the elided commands and the
  trailing-slash package path were corrected rather than silenced.
- Kept D-14-122's identifier while removing its now-dangling reconciliation block, because
  `PHASE-14-DEBT.md:1582` cites the range "D-14-48 through D-14-122". Flagged for owner review.
</decisions_made>

<blockers>
- None blocking. Two advisory code-review warnings and one open window remain open by choice
  (see Infrastructure State).
</blockers>

## Required Reading (in order)

1. The BLOCKING CONSTRAINTS section above — explains how `main` was reported green while red.
2. `AGENTS.md` — primary project instruction file (per `CLAUDE.md` routing).
3. `.planning/phases/14-evidence-instrument-and-honest-scoping/14-VERIFICATION.md` — what
   Phase 14 actually proved, and the shape verification artifacts must take to survive the lint.
4. `.planning/phases/14-evidence-instrument-and-honest-scoping/14-REVIEW.md` — the two open
   advisory warnings.
5. `.planning/phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md` — the debt
   register, including the D-14-122 row awaiting an owner decision.

## Infrastructure State

- Branch `main` at `e1effab`; working tree clean except untracked `.planning/milestone.lock`.
- One pre-existing `git stash` entry predating this session (`WIP on main: 9c395ed`) —
  deliberately untouched, not created by this work.
- Open windows ledger: 1 open, 0 waived — stale `~60s`/`~120s` estimated-runtime prose in
  `08-VALIDATION.md` (lines 27, 39, 94) and in 10/11/12/13-VALIDATION.md. Narrative rather
  than a verification command span, so no Phase 14 plan owned it; recommended for QLT-10.
  `/gsd-ship` blocks while it remains open.
- `.planning/EVIDENCE-RECONCILIATION.md` is GENERATED (`entries: 66`) and byte-compared by
  `TestEvidenceReconciliationViewIsCurrent`. There is no regeneration command — it is derived
  from the `*-DEBT.md` registers' ```reconciliation``` fenced blocks. Never hand-edit it.
- Full suite takes ~5-6 min; `internal/compiler/session` alone is ~250-320s.

<context>
The through-line of this session: Phase 14's entire premise is that an instrument must not
report green for work that is merely wired — and the phase's own completion violated exactly
that, then its own lint caught it. That is a good outcome for the instrument and a bad one
for the workflow around it. The gap-before-mutation constraint above is the durable lesson;
everything else is bookkeeping.

Phase 15 is a genuinely different kind of work — event identity and `OpCall` emission, not
evidence machinery. Do not carry Phase 14's debt-register reflexes into it uncritically.
</context>

<next_action>
Start with: `/gsd-discuss-phase 15` — no `15-CONTEXT.md` exists yet. Before running any
workflow that ends in a phase-completion step, read the BLOCKING CONSTRAINTS section above
and apply the gate-after-mutation mitigation.
</next_action>
