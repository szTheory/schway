---
status: resolved
trigger: Phase35 finite admission mutation harness defects
created: 2026-10-14
updated: 2026-10-14
---

## Current Focus

hypothesis: Source edit selectors, redundant law guards, and marker attribution prevent 13 of the 26 bounded admission controls from yielding valid named-assertion kills.
test: Focus failed controls first, then exact finite harness command.
expecting: All 26 compilable, vet-clean law mutants fail their intended named assertion; no-op and restored baselines pass.
next_action: Commit this resolved automated debug record and the two scoped phase evidence files; parent reports the exact commits and limits. No human-only observation remains in this authorized task.
bug_class: Bohrbug
reasoning_checkpoint:
  hypothesis: Harness source selectors and incomplete coherent law removal cause invalid/masked mutants; overlapping marker substrings double-count source failures.
  confirming_evidence:
    - Focused pre-fix run reproduced 13 failures; /tmp/phase35-admission-before.log records actual compiler/vet/cardinality/survival and diagnostic mismatches.
    - Source inspection shows missing immediates still fail ParseUint error checks, replay selectors also match other operations, and target mutations abandon downstream copy inputs.
    - Origin foreign-field witness reaches local_owner_lifecycle after scalar shape refusal removal; source marker SOURCE_ADMISSION_ACCEPTED contains ADMISSION_ACCEPTED.
  falsification_test: A compilable targeted overlay still refuses outside the intended named assertion after coherent guard removal and fixture consumer repair.
  fix_rationale: Constrain edits to shift law regions, retain range during canonical removal, remove redundant guards for that law only, preserve fixture consumers, and parse exact per-test markers.
  blind_spots: Focused runs have not yet established all 26 controls and three no-op/restored peer baselines together.
  candidate_causes:
    - code: broad substring selectors, unused bindings and overlapping marker parsing in harness
    - data: raw fixture target/type edits leave incompatible downstream consumers; foreign witness activates a separate owner contract
    - environment: Go compiler/vet correctly rejects malformed mutant source; not a production semantic regression
  and_gate: yes; valid source mutation, removal of redundant same-law guards and discriminating fixture consumers must all hold before attribution is meaningful.

## Symptoms

DATA_START
expected: Exact finite 26 controls (12 core, 12 origin, 2 source check), D09/D10/D11 and 35-04 plan tasks 2/3, actual source overlays disable each independent law and require its intended named child assertion to fail.
actual: Previous bounded audit credited only 13/26 kills. No-op and restored three package baseline passed. Remaining issues are invalid/mismapped source edits, compile/vet failures, redundant refusal guards and overlapping marker parsing.
errors: Core canonical diagnostic mismatch; source/result type edit cardinality; result place unused target; initialization/identity masked by place_uninitialized; store edge false-or-false vet. Origin canonical survived; result place zero edits; identity/freshness unused freshTarget; unrelated fields masked by local_owner_lifecycle. Source range reaches decimal_64 SOURCE_ADMISSION_ACCEPTED but substring ADMISSION_ACCEPTED counts it twice.
timeline: Authorized audit work remains uncommitted after validator freshness repair 3800080e.
reproduction: GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase35AdmissionRefusalMutationControls$' -count=1 -v -timeout=600s
historical_credit: core range, RightID, freshness, unrelated fields, ConstU64 exclusivity; origin range, RightID, source type, result op type, initialized source, store edge fields, ConstU64 exclusivity; source number leaf.
production_hashes: core 6c8892a0399ff817e86349fe80bd635429e969d63bd738c85ae3585c87c4466f; origin 583468465ff408e25b92854174f61546de65fd4a5444b24b33349c806e0e0fa6; check a2cacabc99cd3908cacdd11f50a3c29487576a8d3489a77d86f322c8cd6c0167.
DATA_END

## Scope and Authorization

- Own internal/compiler/session/session_phase35_admission_mutation_test.go and assertion/table edits in internal/compiler/check/check_phase35_test.go, internal/compiler/corevalidate/corevalidate_phase35_test.go, internal/compiler/originvalidate/originvalidate_phase35_test.go. Preserve and finish this uncommitted authorized audit work.
- Never edit, stage or revert preexisting dirty .planning/config.json. No production writes/hooks/runtime costs, dependencies, backend, install, broad tests/vet/build/native matrix, hosted or push actions.
- Keep exact finite 26 controls. Compile/missing/unrelated-failure credit prohibited. No -vet=off. Coherent redundant-guard edits per law allowed; global validator disable or weakened positive oracles prohibited.
- Existing 15 arithmetic RHS mutants are a separate proven harness; no rerun unless affected.
- Root pin: bash /tmp/schway34-root-pin.sh before first edit and every commit. Use exact GSD query commit files, with escalation justification if needed. No git add -A, reset, stash, worktree or branch.
- On full green, write phase-local 35-ADMISSION-MUTATION-EVIDENCE.md and update 35-VALIDATION.md with actual controls, hashes, commands and limits. Test and docs commits separate preferred. Do not alter STATE, ROADMAP, REQUIREMENTS, SECURITY, REVIEW or VERIFICATION, or mark phase complete.
- User authorized autonomous routine follow-up; only genuinely unobservable human validation needs checkpoint.

## Evidence

- timestamp: 2026-10-14
  checked: Exact GSD query commit of four owned test files after root-pin guard
  found: Commit 397bfb6d contains only session_phase35_admission_mutation_test.go plus check/corevalidate/originvalidate phase35 assertion files. The config file remains excluded and dirty.
  implication: Tested harness/assertion repair is atomically committed. Supplemental receipt and this session are committed separately.

- timestamp: 2026-10-14
  checked: Exact full command GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase35AdmissionRefusalMutationControls$' -count=1 -v -timeout=600s
  found: PASS exit 0 in 9.566s on go1.24.0 darwin/arm64. 26 MUTANT_KILLED results (12 core, 12 origin, 2 source), 55 exact replacement sites, 46 intended assertion failures, 3 NO_OP_PASS and 3 RESTORED_PASS. Baseline rows are 34/core, 34/origin and 9/check. All production hashes equal immutable Symptoms hashes.
  implication: Automated scoped objective is verified; finite control counts are guarded in source. Detailed inventory is recorded in 35-ADMISSION-MUTATION-EVIDENCE.md.

- timestamp: 2026-10-14
  checked: Old harness test-file overlay with source-range selector, then the same selector on current harness
  found: Saved old harness fails with exact 2-versus-1 source marker count bug (exit 1, 0.177s); current harness passes (exit 0, 0.187s). Live files were never reverted. git diff --check passes.
  implication: The attribution fix is causal; production hashes remain identical. User authorization allows routine locally observable verification without a human checkpoint.

- timestamp: 2026-10-14
  checked: Cycle 2 focused core result-place control, /tmp/phase35-admission-cycle2.log
  found: PASS with three exact source-edit sites and intended wrong_result_place_type ADMISSION_ACCEPTED assertion; no compiler/vet failure credited.
  implication: All originally failing mechanisms have focused passing evidence; full finite matrix is now appropriate.

- timestamp: 2026-10-14
  checked: Focused cycle 1, /tmp/phase35-admission-cycle1.log
  found: 18 of 19 selected controls yield their intended assertion kills; only core result-place source mutation fails vet because two adjacent guards were replaced by true && true. Production hashes remain identical to Symptoms.
  implication: Replace the two adjacent guard clauses as one bounded edit; no semantic production change is needed.

- timestamp: 2026-10-14
  checked: Focused pre-fix 19-control run (exact selector in /tmp/phase35-admission-before.log)
  found: 13 failed controls reproduced; 6 selected historical controls passed. No production writes. Knowledge-base keyword scan yielded no matching harness diagnosis; SBFL skipped because per-test coverage is unavailable. No project skill index present.
  implication: Deterministic harness/fixture repair is warranted; no production fix is justified.

- timestamp: 2026-10-14
  checked: Parent audit handoff and working-tree status
  found: Four owned test files dirty/untracked; .planning/config.json preexisting dirty; production files clean.
  implication: Repair harness and assertion tables only; preserve production hashes.

## Eliminated

## Resolution

root_cause: Source edit selectors matched unrelated operation branches; canonical/missing and initialization laws retained redundant refusals; result/identity fixtures abandoned downstream consumers; mutant edits introduced dead bindings or redundant boolean clauses rejected by compiler/vet; source acceptance markers overlapped raw markers during substring counting; origin foreign witness activated an unrelated owner lifecycle refusal.
fix: Constrain source regions and exact edit counts; coherently remove redundant same-law guards while preserving other laws; preserve used bindings and avoid duplicate boolean literals; rewire fixture consumers; isolate origin unrelated-field law with payload witness; attribute exact marker tokens to the intended child; enforce 12/12/2 inventory. No production edits.
oracle_type: specified (existing shift IR/source admission and exact diagnostic contracts)
verification:
  target_test: { result: pass, controls: 26, intended_assertions: 46, source_replacement_sites: 55, no_op_peers: 3, restored_peers: 3 }
  mutation_check: { result: skipped, reason_if_skipped: No Stryker/Go mutation runner is configured for the changed harness parser; no dependency installation authorized. Executed 26 bounded owner-source mutants are recorded separately. }
  no_op_deletion: { result: pass, deletion_justified_by_rca: true, detail: Test-only overlay guard removals intentionally seed coherent law faults; production behavior is unchanged and positive admission/diagnostic oracles are retained. }
  adjacent_tests: { result: pass, suites_run: [TestPhase35ShiftImmediateAdmission, TestPhase35ShiftTargetFreshness, TestPhase35ShiftAdmission], detail: Each complete named peer baseline ran before mutations and after restoration; includes accepted 0/63 and rejected malformed/range/type/field neighbors. }
  revert_and_reconfirm: { result: pass, bug_returned_on_revert: true, fixed_on_reapply: true, detail: Old harness replayed via test-file overlay on exact source-range control; current harness rerun passed without live file revert. }
  guardrail_verdict: accepted
files_changed:
  - internal/compiler/session/session_phase35_admission_mutation_test.go
  - internal/compiler/check/check_phase35_test.go
  - internal/compiler/corevalidate/corevalidate_phase35_test.go
  - internal/compiler/originvalidate/originvalidate_phase35_test.go
  - .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-ADMISSION-MUTATION-EVIDENCE.md
  - .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-VALIDATION.md
commit: 397bfb6d
limits: Finite law-level controls, not every spelling/route/operation Cartesian product; no production change, broad suite, separate build/vet, native/sanitizer, hosted or cross-platform run. Existing 15 arithmetic transfer controls unaffected and not rerun. No phase completion claimed. Knowledge base not written because outside delegated owned paths; scoped prevention is exact source-edit cardinalities, finite inventory, retained baseline neighbors and exact marker attribution.
