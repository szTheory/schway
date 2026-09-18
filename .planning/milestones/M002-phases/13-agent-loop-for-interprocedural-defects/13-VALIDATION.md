---
phase: "13"
slug: "agent-loop-for-interprocedural-defects"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: true
wave_0_complete: true
created: "2026-09-13"
evidence_vocabulary: v1
graded_rows: 25
---

# Phase 13 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go stdlib `testing` — no external test framework anywhere in the tree |
| **Config file** | none — standard `go test` |
| **Quick run command** | `go test ./internal/compiler/check/... ./internal/compiler/session/... -count=1` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~60–120 seconds full suite (per prior-phase SUMMARY timings) |

---

## Sampling Rate

- **After every task commit:** Run the package-scoped quick command for the package touched —
  `go test ./internal/compiler/check/... -count=1`,
  `go test ./internal/compiler/session/... -count=1`, or
  `go test ./cmd/lang-repair/... -count=1`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd-verify-work`:** Full suite must be green, **plus** an explicit re-run of
  `go test ./internal/compiler/check/... -run TestInterproceduralDiagnosticOrderingStability -v -count=1`
  to confirm that exactly the six expected-churn pinned-ID rows changed (D-13-09a) and nothing
  else did. **Corrected 13-07 (Task 3):** the original name here, `TestOrderingStability`, matches
  no test function in the tree — `go test -run` exits 0 when its pattern matches nothing, so this
  row could never have failed. Verified: `go test ./internal/compiler/check/... -run
  TestOrderingStability -v -count=1` → `testing: warning: no tests to run` / `ok ... [no tests to run]`.
  The real test is `TestInterproceduralDiagnosticOrderingStability`
  (`internal/compiler/check/check_ordering_stability_test.go:359`).
- **Max feedback latency:** ~120 seconds

---

## Per-Task Verification Map

Every row below was run for real against the tree at plan 13-07 Task 3 (not read from a
SUMMARY's claim). Where the original command's test name did not exist in the tree, it is
corrected here and the correction is noted.

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Grade | Non-inertness |
|---|---|---|---|---|---|---|---|---|---|---|
| Task 2 | 13-02 | 0 | DX-06 | T-13-33 | Blame resolver classifies every currently-shipped interprocedural code; `blame_undetermined` fires for any unclassified declared fact rather than falling through to caller-blame (D-13-07 fail-open closure) | unit | `go test ./internal/compiler/check/... -run TestBlameClassifiesEveryInterproceduralCode -v -count=1` | ✅ | EXERCISED | — |
| Task 2 | 13-02 | 0 | DX-06 | T-13-33 | B1 exhaustiveness guard (`blameFieldWitness`) is compile-time, not a runtime-`-run`-able test: a fixed-size array literal that fails `go build` if a declared-contract-field arm is deleted. **Corrected 13-07:** no test function of this shape exists to `-run`; the guard was demonstrated once by a scratch deletion + `go build ./...`, then reverted (13-02-SUMMARY.md). The durable, repeatable proxy is that the package still builds with all five arms present | mutation-kill (demonstrated once, guarded by ordinary compilation thereafter) | `go build ./...` | ✅ | WIRED | — |
| Task 1 | 13-06 | 1 | DX-06 (criterion 3) | T-13-31 | Twin-pair repair-then-re-check: naive detection-site fix leaves the other caller broken, so the program does NOT re-check clean; only the compiler-named fix site yields clean | integration | `go test ./cmd/lang-repair/... -run TestTwinPairBlame -v -count=1` | ✅ | EXERCISED | — |
| Task 1 | 13-06 | 1 | DX-06 (criterion 3) | T-13-33 | Mutation-kill D-13-30(d): applying the repair at the detection-site function instead of the compiler-named one must NOT re-check clean | mutation-kill | `go test ./cmd/lang-repair/... -run TestTwinPairBlameGuard -v -count=1` | ✅ | EXERCISED | — |
| Task 2 | 13-06 | 1 | DX-07 | — | Both genuinely repairable Set A classes reach `repaired` via the JSON protocol, single-pass, on the real corpus (`use_matching_argument` withdrawn, D-13-10a — see PHASE-13-DEBT.md) | integration | `go test ./cmd/lang-repair/... -run TestRepairDriverFixesEveryDefectClassSinglePass -v -count=1` | ✅ | EXERCISED | — |
| Task 2 | 13-05 | 1 | DX-07 | T-13-33 | `use_matching_argument` uniqueness gate (D-13-10, D-13-10a): zero, one, or ≥2 matches all emit NO repair — the class ships zero repairs on every partition, honestly `unrepairable` rather than a no-op `repaired` | unit | `go test ./internal/compiler/check/... -run TestUseMatchingArgumentUniquenessGate -v -count=1` | ✅ | EXERCISED | — |
| Task 3 | 13-06 | 1 | DX-07 | T-13-34 | `unrepairable` is never laundered into a pass, extended to the new classes as class-specific subtests | integration | `go test ./cmd/lang-repair/... -run TestUnrepairableDefectFailsTheGate -v -count=1` | ✅ | EXERCISED | — |
| Task 2 | 13-06 | 1 | DX-07 | — | `wrap_call_in_try` negative case: a non-fallible enclosing shape trades the diagnostic and must report `reverify_failed`, never `repaired` | integration | `go test ./cmd/lang-repair/... -run TestWrapCallInTryReverifyFailed -v -count=1` | ✅ | EXERCISED | — |
| Task 3 | 13-06 | 1 | DX-07 (non-contamination) | T-13-33 | `TestRepairDriverSourceNeverReferencesHeldoutFixtures` widened beyond `cmd/lang-repair` to scan every non-test file in `internal/compiler/check`, with its own non-inert proof (`TestRepairDriverSourceHeldoutScanIsNotInert`) | static/AST inspection | `go test ./cmd/lang-repair/... -run 'TestRepairDriverSourceNeverReferencesHeldoutFixtures\|TestRepairDriverSourceHeldoutScanIsNotInert' -v -count=1` | ✅ | EXERCISED | — |
| Task 3 | 13-04 | 1 | DX-07 (D-13-27) | T-13-33 | `testdata/phase13/HELDOUT.sha256` matches current fixture bytes; a one-byte flip in a temp copy fails the control (D-13-30(c)) | unit + mutation-kill | `go test ./internal/compiler/session/... -run 'TestHeldoutCorpusSealed\|TestHeldoutCorpusSealGuardIsNotInert' -v -count=1` | ✅ | EXERCISED | — |
| Task 2 | 13-04 | 1 | DX-07 (D-13-26) | — | Topology-triple disjointness between held-out and derivation; held-out hop distance ≥ 2, derivation ≤ 1 | unit | `go test ./internal/compiler/session/... -run TestCorpusTopologyDisjoint -v -count=1` | ✅ | EXERCISED | — |
| Task 2 | 13-04 | 1 | DX-07 (D-13-30b) | T-13-31 | **The control that proves the split is not ceremonial:** feeding the topology control an alpha-renamed copy of a derivation fixture must come out topologically EQUAL, proving the split would have caught it as held-out | mutation-kill | `go test ./internal/compiler/session/... -run TestCorpusTopologyGuardIsNotInert -v -count=1` | ✅ | EXERCISED | — |
| Task 1 | 13-04 | 1 | DX-07 (D-13-30a) | T-13-32 | Each new interprocedural injector refuses fail-closed when its marker disappears; guard-disabled twin proves non-inertness; target-choice is specified, not incidental | mutation-kill | `go test ./internal/compiler/session/... -run 'TestPhase13InjectorGuardsAreNotInert\|TestEveryInjectorRefusesWhenMarkerDisappears\|TestInjectorMarkerCountGuardIsNotInert\|TestInjectorTargetChoiceIsSpecified' -v -count=1` | ✅ | EXERCISED | — |
| Task 1 | 13-03 | 2 | DX-05 | — | D-13-23 three-function-plus-decoy fixture: every node's `function_id` equals the innermost AST function containing its span, per an independently written test-side resolver | unit | `go test ./internal/compiler/session/... -run TestExplainFunctionAttribution -v -count=1` | ✅ | EXERCISED | — |
| Task 1 | 13-03 | 2 | DX-05 | — | **The anti-hardcoding mutation:** reordering function declarations shifts every span but changes no `function_id` | unit | `go test ./internal/compiler/session/... -run TestExplainFunctionIdentitySurvivesReorder -v -count=1` | ✅ | EXERCISED | — |
| Task 1 | 13-03 | 2 | DX-05 | — | A cause with `Span == nil` emits no function field and `availability: not_captured` rather than guessing | unit | `go test ./internal/compiler/session/... -run TestExplainNilSpanOmitsFunction -v -count=1` | ✅ | EXERCISED | — |
| Task 3 | 13-03 | 2 | DX-05 (D-13-19) | T-13-31 | A cause whose span equals a function declaration span parents a later in-body cause with `caused_by`, never `narrows`; cross-function parenting is `caused_by` | unit | `go test ./internal/compiler/session/... -run TestNarrowsFunctionScopeGuard -v -count=1` | ✅ | EXERCISED | — |
| Task 3 | 13-03 | 2 | DX-05 (D-13-19) | T-13-31 | Mutation-kill: disabling the function-scope guard makes the above test fail | mutation-kill | `go test ./internal/compiler/session/... -run TestNarrowsFunctionScopeGuardIsNotInert -v -count=1` | ✅ | EXERCISED | — |
| Task 1 | 13-03 | 2 | DX-05 (criterion 1) | — | Existing `truncated:explain.node_budget` and `truncated:explain.depth` codes remain stable and unchanged across function boundaries; no new truncation code introduced (D-13-21) | unit | `go test ./internal/compiler/session/... -run TestExplainTruncationCodesStable -v -count=1` | ✅ | EXERCISED | — |
| Task 1 | 13-03 | 2 | DX-05 | — | Node ordering contract `(depth, span.start, span.end, id)` still holds with function fields present | regression | `go test ./internal/compiler/session/... -run TestExplainNodeOrderIsStableOnTies -v -count=1` | ✅ exists pre-phase | EXERCISED | — |
| Task 2 | 13-03 | 2 | D-13-20 (bump trigger) | — | **Decision gate, not just a test:** captured pre-guard edge-kind baseline, re-verified green after the D-13-19 guard landed — no edge kind flipped on unchanged single-function input, so `lang.explain/0` does NOT bump | regression | `go test ./internal/compiler/session/... -run TestExplainEdgeKindsAreRecordedForGuardComparison -v -count=1` | ✅ | EXERCISED | — |
| Task 3 | 13-05 | 3 | D-13-09a (expected churn) | — | Exactly the six pinned-ID rows change; nothing else does — the final two rows re-pinned this task, proving the phase's total churn is exactly six | regression | `go test ./internal/compiler/check/... -run TestInterproceduralDiagnosticOrderingStability -v -count=1` | ✅ re-pinned | EXERCISED | — |
| Task 1 | 13-07 | 3 | D-13-33 | T-13-31, T-13-32 | `testdata/phase6` distinctness strengthened beyond byte-inequality with an identifier-independent structural predicate valid for intraprocedural fixtures; move/borrow class pairs found structurally identical and carried as ratified debt (D-13-34), `match` class pair discriminates and passes | unit | `go test ./internal/compiler/session/... -run 'TestPhase6DefectCorpusIsHeldOut\|TestPhase6DefectCorpusDistinctnessGuardIsNotInert' -v -count=1` | ✅ strengthened | EXERCISED | — |
| Task 3 (13-01) | 13-01 | all | D-13-32 | — | `cmd/lang-repair` SOURCE is unchanged for the whole phase; only its tests grow | static | `git diff --quiet -- cmd/lang-repair/repair.go cmd/lang-repair/main.go` | ✅ exists | REACHABLE | — |
| pre-existing | n/a | all | boundary | T-13-SC | The structural import-boundary lint still holds; no new test imports `internal/`. **Corrected 13-07:** the original pattern `-run TestImportBoundary` matches only `TestImportBoundaryTestIsNotInert` by substring — the actual lint, `TestRepairDriverImportsStayOutsideInternal`, does not contain that substring and was never run by this row. Both are now named explicitly | static | `go test ./cmd/lang-repair/... -run 'TestRepairDriverImportsStayOutsideInternal\|TestImportBoundaryTestIsNotInert' -v -count=1` | ✅ exists | EXERCISED | — |

*Status legend: ✅ green · ❌ red · ⚠️ flaky (no row below is left unresolved at phase close)*

---

## QLT-08 Mutation-Kill Completeness

One row per new control shipped this phase, naming the control, its mutation-kill seam, the
test proving the seam is not inert, and the command that runs it. All four D-13-30 seams (a–d)
are named explicitly, plus D-13-33's extension of seam (b) to M001's own corpus and the one
seam that was withdrawn rather than shipped.

| Seam | Control | Mutation-Kill Test | Command | Status |
|------|---------|---------------------|---------|--------|
| (a) injector guards | Three new phase-13 interprocedural defect injectors (`InterproceduralLoanInjector`, `FallibleConsumeInjector`, `CallArgumentTypeInjector`) | `TestPhase13InjectorGuardsAreNotInert` (each guard-disabled twin, given marker-free source, returns it unchanged) plus the extension of `TestEveryInjectorRefusesWhenMarkerDisappears`/`TestInjectorMarkerCountGuardIsNotInert`/`TestInjectorTargetChoiceIsSpecified` to the three new markers | `go test ./internal/compiler/session/... -run 'TestPhase13InjectorGuardsAreNotInert\|TestEveryInjectorRefusesWhenMarkerDisappears\|TestInjectorMarkerCountGuardIsNotInert\|TestInjectorTargetChoiceIsSpecified' -v -count=1` | ✅ green |
| (b) topology-distinctness control (Phase 13 corpus) | `computeCorpusTopology`'s (function count, call-edge count, hop distance) triple over `testdata/phase13` | `TestCorpusTopologyGuardIsNotInert` — an alpha-renamed derivation copy comes out topologically EQUAL to the original, proving it would have been caught had it been submitted as held-out | `go test ./internal/compiler/session/... -run TestCorpusTopologyGuardIsNotInert -v -count=1` | ✅ green |
| (b) extended — topology/structural-distinctness control (M001 corpus, D-13-33) | `computePhase6StructuralSummary`'s (bindingCount, matchArmCount, borrowCount, takeCount, maxDepth) predicate over `testdata/phase6`, replacing byte-inequality per D-13-33's retro-strengthening instruction | `TestPhase6DefectCorpusDistinctnessGuardIsNotInert` — an alpha-renamed derivation copy comes out structurally EQUAL to the original, and is structurally distinct from the real held-out sibling, proving the predicate genuinely discriminates | `go test ./internal/compiler/session/... -run TestPhase6DefectCorpusDistinctnessGuardIsNotInert -v -count=1` | ✅ green |
| (c) sealed-digest control | `testdata/phase13/HELDOUT.sha256` manifest, asserted to match current bytes of all five `heldout_*.lang` fixtures | `TestHeldoutCorpusSealGuardIsNotInert` — a one-byte flip in a `t.TempDir()` temp copy fails the seal check | `go test ./internal/compiler/session/... -run TestHeldoutCorpusSealGuardIsNotInert -v -count=1` | ✅ green |
| (d) criterion-3 detection-site control | Twin-pair repair-then-re-check oracle (D-13-28): the shared-callee `sink` fixture pair, blamed at the compiler-named function | `TestTwinPairBlameGuard` — applying the repair at the detection-site function (`alpha`) instead of the compiler-named one (`sink`) must NOT re-check clean | `go test ./cmd/lang-repair/... -run TestTwinPairBlameGuard -v -count=1` | ✅ green |
| (withdrawn) `use_matching_argument` repair-correctness seam | Would have been a mutation-kill proving the emitted `Replacement` actually repairs a real mismatch | **WITHDRAWN, not silently dropped** — D-13-10a found the class's only possible `Replacement` is byte-identical to the existing token on every real trigger (every Lang function shares one type fact, so the uniqueness gate's one match is always the argument's own already-passed place). No repair is emitted on any partition, so there is nothing for a correctness seam to mutate-kill. The honest replacement assertion is `TestUseMatchingArgumentUniquenessGate`, which now proves zero repairs on every partition (zero, one, two-or-more matches) rather than proving a repair is correct | `go test ./internal/compiler/check/... -run TestUseMatchingArgumentUniquenessGate -v -count=1` | ✅ green (assertion changed in kind, per D-13-10a — see PHASE-13-DEBT.md) |

---

## Unresolved edge-coverage assumptions

Two `unclassified` probe rows were never resolved to "fully met" — DX-06 and DX-07 — restated
here explicitly so phase verification disposes of them rather than inheriting them silently.
Both are ratified terminal findings, recorded in `PHASE-13-DEBT.md`, not open work items.

- **DX-06 — "Cross-function blame attribution points at the correct fix location; a
  repair-then-re-check regression covers cases where the non-obvious function is the right
  one."** Plan 13-02's planner assumption was that the B1/B2/B3 contract-boundary rule would
  have at least one B1-shaped (contract-violation) production trigger to discriminate against
  B2 (caller misuse). Plan 13-04 established empirically, and the phase-close checkpoint
  ratified (D-13-02b, PHASE-13-DEBT.md), that **no B1-shaped interprocedural diagnostic is
  constructible at this language's current maturity** — `sameType(ReturnType, Parameter.Type)`
  is an admission precondition enforced independent of any call, so a callee that breaks its own
  contract is refused before the interprocedural pass can reach it. Criterion 3's twin pair
  (D-13-28) genuinely discriminates detection-site blame from caller blame, which is real and
  useful, but does not exercise the contract-boundary rule's distinctive B1 claim. DX-06 is
  **partially met**: the regression exists and is sound; the specific "non-obvious function
  broke its own contract" case is unconstructible and untested, by construction of the language,
  not by a testing gap.

- **DX-07 — "`lang-repair` reaches and fixes at least three new interprocedural defect classes
  through the JSON protocol alone, proven on a held-out fixture split."** Plan 13-05's planner
  assumption was that `use_matching_argument`'s uniqueness gate (D-13-10) would, on its
  exactly-one-match partition, name a genuinely different and correct argument. Plan 13-05
  found empirically, and the phase-close checkpoint ratified (D-13-10a, PHASE-13-DEBT.md), that
  every Lang function admits exactly one parameter and one type fact, so the "unique match" the
  gate could ever find is always the argument's own already-passed place — the `Replacement` is
  byte-identical to the original token on every real trigger. `use_matching_argument` now emits
  no repair on any partition (`unrepairable`, honestly, never a no-op `repaired`). DX-07 ships
  with **two** genuinely repairable classes, not three: `move_after_interprocedural_loan`
  (backward direction only, D-13-09b) and `wrap_call_in_try`. DX-07 is **partially met**.

Both findings were adjudicated at plan 13-07's `gate="blocking-human"` checkpoint (2026-09-13)
rather than absorbed silently or worked around inside this phase's budget. `REQUIREMENTS.md`'s
traceability table marks both `Partial` with a pointer to the owning `PHASE-13-DEBT.md` row.

---

## Wave 0 Requirements

- [x] **Read `cmd/lang-repair/antitheater_test.go` in full** — done at plan 13-01 (13-01-SUMMARY.md
      confirms this closed RESEARCH.md's single LOW-confidence area)
- [x] Blame resolver unit test file under `internal/compiler/check/` — `check_blame_test.go`
      (plan 13-02), covers DX-06
- [x] D-13-23's three-function-plus-decoy `.lang` fixture + the independently written test-side
      function resolver — `testdata/phase13/explain_three_function_chain[_reordered].lang`
      (plan 13-03), covers DX-05
- [x] D-13-28's twin-pair `.lang` fixtures, **both halves** —
      `heldout_shared_callee_twin_alpha.lang` / `heldout_shared_callee_twin_mirror.lang`
      (plan 13-04), covers DX-06 criterion 3, the phase's riskiest-assumption gate
- [x] New interprocedural injectors (`internal/compiler/session/session_phase13_injectors.go`),
      each with its marker constant and guard-disabled twin — covers DX-07 criterion 2
- [x] `testdata/phase13/` corpus + `HELDOUT.sha256` manifest and its sealing test (plan 13-04) —
      covers D-13-27
- [x] Framework install: **none** — Go stdlib `testing` only, confirmed unchanged for the whole
      phase

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions | Resolution |
|----------|-------------|------------|--------------------|------------|
| D-13-20 schema-bump decision | DX-05 | Whether `lang.explain/0` bumps to `/1` depends on observing whether the narrows guard flips any edge kind on existing fixtures. The observation is automated; the **decision** that follows is a human judgement about published-output stability. | Run the explain regression suite before and after landing the guard, diff the emitted edge kinds, and decide. | **Resolved, no bump.** Plan 13-03 Task 2 captured the pre-guard baseline (`TestExplainEdgeKindsAreRecordedForGuardComparison`); Task 3 landed the D-13-19 guard and re-ran it green — no edge kind flipped on unchanged single-function input, confirming RESEARCH.md's prediction rather than assuming it. `lang.explain/0` unchanged. |
| D-13-33 escalation call | QLT-08 | If retro-strengthening `testdata/phase6` turns currently-green tests red, that is a real hole in shipped M001 evidence. CONTEXT.md directs escalation rather than weakening the predicate — a human decides. | Apply the structural predicate, observe, and if red, escalate to the user rather than relaxing it. | **Resolved.** Plan 13-07 Task 1 applied the predicate: `move` and `borrow` class pairs came back structurally identical (real hole). Escalated at Task 2's `gate="blocking-human"` checkpoint; developer ratified Option B — keep the predicate unweakened, record the finding as permanent M001 evidence debt (D-13-34, PHASE-13-DEBT.md), do not edit shipped fixtures, do not scope the predicate down. |
| D-13-08 accepted limitation | DX-06 | Mutual-consistency-with-incompatible-intent is unfalsifiable by repair-then-re-check by construction. No automated test can cover it. | Confirm by inspection that such cases emit both sites as `RequiresConfirmation`, so `DriverEligible` refuses to auto-apply either. | **Confirmed by inspection** during plan 13-02 (`blame_undetermined` fail-open closure, `TestBlameUndeterminedPublishesBothSitesUnapplied`) — both sites published, neither `MachineApplicable`. |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify — every task across all
      seven plans (13-01 through 13-07) carries at least one `<automated>` verify in its own
      PLAN.md, confirmed above by running each one
- [x] Wave 0 covers all MISSING references — all six items closed, see § Wave 0 Requirements
- [x] No watch-mode flags — every command above is a one-shot `go test`/`go build`/`git diff`
- [x] Feedback latency < 120s — every individual command above completed in well under 1 second;
      the full suite (`go test ./...`) is the ~60–120s outer bound
- [x] Every new control has a QLT-08 mutation-kill seam (D-13-30 a–d all present) — see
      § QLT-08 Mutation-Kill Completeness; the one seam that could not exist (`use_matching_argument`
      repair-correctness) is recorded as withdrawn, not silently omitted
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** signed off at plan 13-07 Task 3, 2026-09-13. Every row in the Per-Task
Verification Map and the QLT-08 Mutation-Kill Completeness matrix was run for real against the
tree during this task, not inferred from a prior SUMMARY's claim. Two grounding errors were
found and corrected in the process: the D-13-09a row's `-run TestOrderingStability` pattern
matched no test (real name `TestInterproceduralDiagnosticOrderingStability`), and the
pre-existing import-boundary row's `-run TestImportBoundary` pattern matched only the not-inert
companion test, never the actual lint (`TestRepairDriverImportsStayOutsideInternal`) — both are
the same underlying defect class `go test -run`'s zero-match-exits-0 behavior enables, corrected
here rather than left for a future phase to rediscover.
