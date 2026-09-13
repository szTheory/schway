---
phase: "13"
slug: "agent-loop-for-interprocedural-defects"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-09-13"
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
  `go test ./internal/compiler/check/... -run TestOrderingStability -v -count=1` to confirm
  that exactly the six expected-churn pinned-ID rows changed (D-13-09a) and nothing else did
- **Max feedback latency:** ~120 seconds

---

## Per-Task Verification Map

Task IDs are assigned by the planner; this map states the requirement→command binding each
task must satisfy. The planner MUST carry every row into a task `<automated>` verify block.

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| TBD | TBD | 0 | DX-06 | — | Blame resolver classifies every interprocedural code; `blame_undetermined` fires for any unclassified declared fact rather than falling through to caller-blame (D-13-07 fail-open closure) | unit | `go test ./internal/compiler/check/... -run TestBlame -v -count=1` | ❌ W0 | ⬜ pending |
| TBD | TBD | 0 | DX-06 | — | B1 exhaustiveness guard is compile-time; removing a declared-contract field arm fails to build | mutation-kill | `go test ./internal/compiler/check/... -run TestBlameExhaustiveness -v -count=1` | ❌ W0 | ⬜ pending |
| TBD | TBD | 1 | DX-06 (criterion 3) | — | Twin-pair repair-then-re-check: naive detection-site fix leaves the other caller broken, so the program does NOT re-check clean; only the compiler-named fix site yields clean | integration | `go test ./cmd/lang-repair/... -run TestTwinPairBlame -v -count=1` | ❌ W0 | ⬜ pending |
| TBD | TBD | 1 | DX-06 (criterion 3) | — | Mutation-kill D-13-30(d): applying the repair at the detection-site function instead of the compiler-named one must NOT re-check clean | mutation-kill | `go test ./cmd/lang-repair/... -run TestTwinPairBlameGuard -v -count=1` | ❌ W0 | ⬜ pending |
| TBD | TBD | 1 | DX-07 | — | Each of the three Set A classes reaches `repaired` via the JSON protocol on held-out fixtures, single-pass | integration | `go test ./cmd/lang-repair/... -run TestRepairDriverFixesEveryDefectClassSinglePass -v -count=1` | ✅ extend | ⬜ pending |
| TBD | TBD | 1 | DX-07 | — | `use_matching_argument` uniqueness gate (D-13-10): zero or ≥2 matches emit NO repair; driver returns `unrepairable`, never a false `repaired` | unit + integration | `go test ./internal/compiler/check/... -run TestUseMatchingArgumentUniquenessGate -v -count=1` | ❌ W0 | ⬜ pending |
| TBD | TBD | 1 | DX-07 | — | `unrepairable` is never laundered into a pass, extended to the three new classes | integration | `go test ./cmd/lang-repair/... -run TestUnrepairableDefectFailsTheGate -v -count=1` | ✅ extend | ⬜ pending |
| TBD | TBD | 1 | DX-07 | — | `wrap_call_in_try` negative case: a non-fallible enclosing shape trades the diagnostic and must report `reverify_failed`, never `repaired` | integration | `go test ./cmd/lang-repair/... -run TestWrapCallInTryReverifyFailed -v -count=1` | ❌ W0 | ⬜ pending |
| TBD | TBD | 1 | DX-07 (non-contamination) | — | `TestRepairDriverSourceNeverReferencesHeldoutFixtures` extended beyond `cmd/lang-repair` to every non-test file in `internal/compiler/check` | static/AST inspection | `go test ./... -run TestRepairDriverSourceNeverReferencesHeldoutFixtures -v -count=1` | ✅ extend scope | ⬜ pending |
| TBD | TBD | 1 | DX-07 (D-13-27) | — | `testdata/phase13/HELDOUT.sha256` matches current fixture bytes; a one-byte flip in a temp copy fails the control | unit + mutation-kill | `go test ./internal/compiler/session/... -run TestHeldoutCorpusSealed -v -count=1` | ❌ W0 | ⬜ pending |
| TBD | TBD | 1 | DX-07 (D-13-26) | — | Topology-triple disjointness between held-out and derivation; held-out hop distance ≥ 2 | unit | `go test ./internal/compiler/session/... -run TestCorpusTopologyDisjoint -v -count=1` | ❌ W0 | ⬜ pending |
| TBD | TBD | 1 | DX-07 (D-13-30b) | — | **The control that proves the split is not ceremonial:** feeding the topology control an alpha-renamed copy of a derivation fixture must fail | mutation-kill | `go test ./internal/compiler/session/... -run TestCorpusTopologyGuardIsNotInert -v -count=1` | ❌ W0 | ⬜ pending |
| TBD | TBD | 1 | DX-07 (D-13-30a) | — | Each new interprocedural injector refuses fail-closed when its marker disappears; guard-disabled twin proves non-inertness | mutation-kill | `go test ./internal/compiler/session/... -run 'TestEveryInjectorRefusesWhenMarkerDisappears\|TestInjectorMarkerCountGuardIsNotInert\|TestInjectorTargetChoiceIsSpecified' -v -count=1` | ✅ extend | ⬜ pending |
| TBD | TBD | 2 | DX-05 | — | D-13-23 three-function-plus-decoy fixture: every node's `function_id` equals the innermost AST function containing its span, per an independently written test-side resolver | unit | `go test ./internal/compiler/session/... -run TestExplainFunctionAttribution -v -count=1` | ❌ W0 | ⬜ pending |
| TBD | TBD | 2 | DX-05 | — | **The anti-hardcoding mutation:** reordering function declarations shifts every span but changes no `function_id` | unit | `go test ./internal/compiler/session/... -run TestExplainFunctionIdentitySurvivesReorder -v -count=1` | ❌ W0 | ⬜ pending |
| TBD | TBD | 2 | DX-05 | — | A cause with `Span == nil` emits no function field and `availability: not_captured` rather than guessing | unit | `go test ./internal/compiler/session/... -run TestExplainNilSpanOmitsFunction -v -count=1` | ❌ W0 | ⬜ pending |
| TBD | TBD | 2 | DX-05 (D-13-19) | — | A cause whose span equals a function declaration span parents a later in-body cause with `caused_by`, never `narrows`; cross-function parenting is `caused_by` | unit | `go test ./internal/compiler/session/... -run TestNarrowsFunctionScopeGuard -v -count=1` | ❌ W0 | ⬜ pending |
| TBD | TBD | 2 | DX-05 (D-13-19) | — | Mutation-kill: disabling the function-scope guard makes the above test fail | mutation-kill | `go test ./internal/compiler/session/... -run TestNarrowsFunctionScopeGuardIsNotInert -v -count=1` | ❌ W0 | ⬜ pending |
| TBD | TBD | 2 | DX-05 (criterion 1) | — | Existing `truncated:explain.node_budget` and `truncated:explain.depth` codes remain stable and unchanged across function boundaries; no new truncation code introduced (D-13-21) | unit | `go test ./internal/compiler/session/... -run TestExplainTruncationCodesStable -v -count=1` | ❌ W0 | ⬜ pending |
| TBD | TBD | 2 | DX-05 | — | Node ordering contract `(depth, span.start, span.end, id)` still holds with function fields present | regression | `go test ./internal/compiler/session/... -run TestExplainNodeOrderIsStableOnTies -v -count=1` | ✅ exists | ⬜ pending |
| TBD | TBD | 2 | D-13-20 (bump trigger) | — | **Decision gate, not just a test:** run existing explain fixtures before landing the guard; if any edge kind flips on unchanged single-function input, `lang.explain/0` must bump to `/1` | regression | `go test ./internal/compiler/session/... -run TestExplain -v -count=1` | ✅ exists | ⬜ pending |
| TBD | TBD | 3 | D-13-09a (expected churn) | — | Exactly the six pinned-ID rows change; nothing else does | regression | `go test ./internal/compiler/check/... -run TestOrderingStability -v -count=1` | ✅ re-pin | ⬜ pending |
| TBD | TBD | 3 | D-13-33 | — | `testdata/phase6` distinctness strengthened beyond byte-inequality with a structural predicate valid for intraprocedural fixtures | unit | `go test ./internal/compiler/session/... -run TestPhase6DefectCorpusIsHeldOut -v -count=1` | ✅ strengthen | ⬜ pending |
| TBD | TBD | all | D-13-32 | — | `cmd/lang-repair` SOURCE is unchanged; only its tests grow | static | `git diff --quiet -- cmd/lang-repair/repair.go cmd/lang-repair/main.go` | ✅ exists | ⬜ pending |
| TBD | TBD | all | boundary | — | The structural import-boundary lint still holds; no new test imports `internal/` | static | `go test ./cmd/lang-repair/... -run TestImportBoundary -v -count=1` | ✅ exists | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] **Read `cmd/lang-repair/antitheater_test.go` in full** — RESEARCH.md flags its
      new-class acceptance contract as its one LOW-confidence area (794 lines, not read
      during research). Every new defect class must satisfy it. This is a Wave 0
      *investigation* item, and its findings may add rows to the map above.
- [ ] Blame resolver unit test file under `internal/compiler/check/` — covers DX-06
- [ ] D-13-23's three-function-plus-decoy `.lang` fixture + the independently written
      test-side function resolver — covers DX-05
- [ ] D-13-28's twin-pair `.lang` fixtures, **both halves** — covers DX-06 criterion 3,
      the phase's riskiest-assumption gate. Either half alone is a coin flip.
- [ ] New interprocedural injectors (in `session_phase6_injectors.go` or a phase-13
      sibling), each with its marker constant and guard-disabled twin — covers DX-07
      criterion 2
- [ ] `testdata/phase13/` corpus + `HELDOUT.sha256` manifest and its sealing test —
      covers D-13-27
- [ ] Framework install: **none** — Go stdlib `testing` only

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| D-13-20 schema-bump decision | DX-05 | Whether `lang.explain/0` bumps to `/1` depends on observing whether the narrows guard flips any edge kind on existing fixtures. The observation is automated; the **decision** that follows is a human judgement about published-output stability. | Run the explain regression suite before and after landing the guard, diff the emitted edge kinds, and decide. RESEARCH.md assesses the flip risk as likely non-existent (cycle_member causes carry narrow OpCall-token spans, not whole-function spans), but this must be confirmed by an actual run, not reasoning. |
| D-13-33 escalation call | QLT-08 | If retro-strengthening `testdata/phase6` turns currently-green tests red, that is a real hole in shipped M001 evidence. CONTEXT.md directs escalation rather than weakening the predicate — a human decides. | Apply the structural predicate, observe, and if red, escalate to the user rather than relaxing it. |
| D-13-08 accepted limitation | DX-06 | Mutual-consistency-with-incompatible-intent is unfalsifiable by repair-then-re-check by construction. No automated test can cover it. | Confirm by inspection that such cases emit both sites as `RequiresConfirmation`, so `DriverEligible` refuses to auto-apply either. |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 120s
- [ ] Every new control has a QLT-08 mutation-kill seam (D-13-30 a–d all present)
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
