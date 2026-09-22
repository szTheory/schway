---
status: resolved
trigger: "Phase 17 review found unchecked terminal return types and C return layouts derived from parameter alternatives; go test ./... has affected legacy digest/golden failures."
created: "2026-09-22T00:00:00Z"
updated: "2026-09-22T21:00:00Z"
---

## Symptoms

- Expected: terminal values conform to each function's declared return type, native lowering derives return layouts from the declared return type, and the full Go suite passes after intentional evidence updates.
- Actual: the canonical Resource-to-Result tracer accepts a parameter-typed terminal value; cgen casts it, and return alternatives are built from parameter alternatives. `go test ./...` fails legacy C, core, corevalidate, and evidence digest/golden checks.
- Error messages: `TestLegacyEmitterEvidence` digest mismatch; `TestPreviousPhaseCoreBytesUnchanged`; `TestPreviousPhaseManifestIDsUnchanged`; corevalidate closure digest mismatch; evidence golden mismatch.
- Timeline: introduced or exposed during Phase 17 execution.
- Reproduction: run the Phase 17 review cases and `go test ./...`.

## Current Focus

- bug_class: bohrbug
- reasoning_checkpoint:
    hypothesis: "Phase 16 Plan 14 made M004 refusal terminal in production, but several older Phase 4–6 tests still call the production boundary to obtain historical C; this causes deterministic failures after Phase 17 correctly exposes those refused shapes."
    confirming_evidence:
      - "REQUIREMENTS NAT-09 and completed 16-14 PLAN/SUMMARY explicitly cut foreign/by-pointer emitters and restrict frozen C to provenance-bound _test.go helpers."
      - "The failing injector tests call Phase16ControlNativeC immediately after asserting cgen.EmitNative must refuse the same program."
      - "The external Phase 5 differential already has a canonical-program-bound phase16FileFrozenEvidenceC test helper, but called it even for publicly admitted programs."
    falsification_test: "If replacing only stale test consumers with live-first emission plus provenance-bound frozen evidence does not make their focused tests pass, the contract mismatch is not sufficient to explain those failures."
    fix_rationale: "Moving historical artifact access behind _test.go-only canonical provenance preserves the production M004 cut while retaining the old semantic evidence; live admitted programs continue through the sole public emitter."
    blind_spots: "Production Phase 4–6 gate tests may intentionally need their expected result changed from pass to terminal refusal, and concurrent untracked Phase 15 corpus artifacts still affect unrelated full-suite guards."
    candidate_causes:
      - "code: stale tests invoke Phase16ControlNativeC as though it were the pre-Plan-14 evidence fallback"
      - "data: synthesized core programs lack a frozen-evidence provenance record and therefore cannot legally obtain historical C"
      - "environment: concurrent untracked corpus and planning files alter corpus counts and document-derived gates"
    and_gate: "yes — the observed suite is red from both the stale test routing and independent concurrent workspace artifacts; only the former belongs to this contract fix."
- hypothesis: "Test-only live-first/frozen-second routing resolves the stale historical semantic tests without weakening Phase 16 or Phase 17."
- test: "Run focused Phase 5 corpus and Phase 6 injector tests, then rerun the full suite after concurrent workspace state settles."
- expecting: "Admitted corpus programs use cgen live; refused source fixtures use canonical frozen evidence; refused synthesized core programs are explicitly outside the native differential."
- next_action: "Archived after the requested Phase 17 and Phase 16 contract scope passed its focused and adjacent guards; separately owned Phase 11/15 planning/workspace failures remain recorded below."

## Eliminated

- hypothesis: legacy golden failures are the root cause rather than derived evidence drift
  evidence: direct source inspection shows the checker emits OpReturn with the returned source TypeID without comparing its nominal shape to ReturnType, and cgen calls branchTypeFor(ReturnType, Parameter.Type).
  timestamp: "2026-09-22T14:00:00Z"

## Evidence

- timestamp: "2026-09-22T14:00:00Z"
  checked: testdata/phase17/return_type_tracer.lang and checker lowering
  found: classify(Resource) -> Result terminates with the Resource parameter, and analyzeArmBody emits OpReturn using that source place's TypeID without validating it against the declared return fact.
  implication: the canonical tracer itself deterministically reproduces the missing terminal return contract.
- timestamp: "2026-09-22T14:00:00Z"
  checked: internal/compiler/cgen/cgen_program.go branch type construction
  found: branchTypeFor(function.ReturnType, function.Parameter.Type) populates a return C type from parameter alternatives, and a one-alternative fallback hides unknown return alternatives.
  implication: native layouts and values are derived from the wrong nominal declaration; the one-alternative tracer masks the defect.
- timestamp: "2026-09-22T14:00:00Z"
  checked: bug reproducibility and suite topology
  found: both defects are deterministic for fixed core/source inputs; focused package tests exist, but no per-test coverage/SBFL harness is configured for this fault.
  implication: classify as Bohrbug and use direct reproduction plus working-backwards fault localization; SBFL is skipped with a logged reason.
- timestamp: "2026-09-22T14:30:00Z"
  checked: focused Phase 17 package tests after the fix
  found: checker, corevalidate, originvalidate, cgen, session, and executionpeer Phase 17 tests pass, including four-tier interpreter/O0/O3/O3-LTO differential execution.
  implication: the return contract and native layout changes work end-to-end on the reported path.
- timestamp: "2026-09-22T14:35:00Z"
  checked: go test ./...
  found: full suite still reports pre-existing/expected legacy digest and golden drift plus exact-work/closure digest expectations; a schema cross-lock test also encodes the old prohibition on mixed core/1 programs containing a pure match function.
  implication: focused behavior is fixed, but the full-suite evidence baseline and affected structural expectations still require reconciliation before guardrail acceptance.
- timestamp: "2026-09-22T15:20:00Z"
  checked: historical core bytes, manifest IDs, evidence golden, and legacy emitter evidence
  found: all drift disappeared when same-type functions reused type:0 and foreign error facts resumed contiguous numbering; no historical digest or golden regeneration was justified.
  implication: the global baseline failures were a production encoding regression, not intentional evidence drift.
- timestamp: "2026-09-22T15:45:00Z"
  checked: schema cross-lock and summary closure digest failures
  found: mixed core/1 programs require pure-match helpers to be admitted, but all-pure-match programs relabeled core/1 remain non-canonical; nil versus empty abilities and an err-type/type:1 collision caused peer digest divergence.
  implication: program-level schema canonicality and constructor-checked type:0 fallback fix the structural failures without weakening validation.
- timestamp: "2026-09-22T16:05:00Z"
  checked: checker, core, corevalidate, cgen, evidence, interp, executionpeer, originvalidate, reduce, and Phase 17 differential/payload replay tests
  found: all passed after commit f278b7e, including legacy core/evidence pins and TestPhase17TwoTypeFourTierDifferential.
  implication: the revised fix preserves old contracts and covers the distinct-return path end to end.
- timestamp: "2026-09-22T16:20:00Z"
  checked: go test ./...
  found: all non-session packages passed; session failures are currently dominated by concurrent Phase 15 workspace state (untracked testdata/phase16/historical/phase11_gate_*.c and modified planning files), which changes corpus counts, frozen native inputs, registries, and groundedness scans. The Phase 17 differential failure observed during this run was independently fixed and its focused rerun passes.
  implication: Phase 17 code is green in its affected graph, but the repository-wide guardrail cannot be accepted until concurrent workspace mutations settle and go test ./... is rerun.
- timestamp: "2026-09-22T17:00:00Z"
  checked: fresh full-suite guardrail after f278b7e
  found: established Phase 4–6 session paths fail with Phase 17-only foreign-call and by-pointer native-emission refusals, and existing corpus programs now fail current-M004 refusal expectations; these failures are independent of untracked historical C artifacts.
  implication: the prior workspace-noise attribution is disproven; reopen investigation around f278b7e schema/admission changes.
- timestamp: "2026-09-22T17:25:00Z"
  checked: isolated worktree at f278b7e^ (7e6d456)
  found: TestCleanupInjectorReusesReleaseOmissionRunner and TestFalseRestrictFixtureIsCleanUnmutated already fail with the same foreign/by-pointer public-emitter refusals before f278b7e.
  implication: the representative Phase 4–6 failures are not caused by f278b7e schema/admission changes.
- timestamp: "2026-09-22T17:30:00Z"
  checked: current Phase 16 control contracts
  found: session_phase16_control_test requires Phase16ControlNativeC to preserve public M004 refusal and forbids production access to frozen evidence, while older Phase 4–6 tests require that same boundary and RunNative to execute the refused foreign/by-pointer programs; Phase 5 frozen-evidence tests expect same-type bare matches to refuse while the current Phase 16 admitted-control test requires toggle.lang to emit successfully.
  implication: the repository contains mutually exclusive historical and current contracts; no implementation-only admission predicate can satisfy both without an authoritative decision about the Phase 16 boundary.
- timestamp: "2026-09-22T18:00:00Z"
  checked: current REQUIREMENTS.md NAT-09 amendment, ROADMAP Phase 16 success criteria, and completed 16-14 PLAN/SUMMARY
  found: all current authoritative artifacts require emitLinearForeign and both by-pointer families to remain cut until M004, require Phase16ControlNativeC to propagate public refusal unchanged, and allow historical C only in provenance-bound test helpers.
  implication: Phase 16 Plan 14 supersedes older Phase 4–6 production-execution assertions; preserve Phase 17 return-contract code and update stale tests/evidence consistently.
- timestamp: "2026-09-22T18:15:00Z"
  checked: focused Phase 5 corpus and Phase 6 injector tests after commit 49757d7
  found: TestPhase5CorpusThreeEngineAgreement and all three affected injector tests pass; admitted programs use live emission, refused source fixtures use canonical frozen evidence, and refused synthesized core programs do not obtain invented evidence.
  implication: the Phase 16 test-only evidence routing resolves this subset without weakening the production refusal or changing Phase 17 return contracts.
- timestamp: "2026-09-22T18:20:00Z"
  checked: full internal/compiler/session package after the focused repair
  found: remaining failures include older Phase 4–6 production gates that still require cut foreign/by-pointer execution, plus independent concurrent Phase 15 corpus artifacts, maturity counts, registries, and planning groundedness drift.
  implication: commit 49757d7 is valid and focused, but the debug session cannot be archived until the broader Plan 14 migration and concurrent workspace reconciliation make go test ./... green.
- timestamp: "2026-09-22T19:05:00Z"
  checked: fresh internal/compiler/session run after continuation
  found: stale Phase 4–6 tests still expect production execution from RunNative, VerifyPhase4, Phase 5 alias/sanitizer/mismatch runners, and Phase 6 native differential; separate failures come from untracked Phase 15 corpus files, emitter-inventory line drift, planning groundedness, and Phase 11 gates.
  implication: migrate only Phase 4–6 assertions to terminal refusal; preserve the independently-owned concurrent files and report their failures separately.
- timestamp: "2026-09-22T19:25:00Z"
  checked: Phase 5 stale-control migration in commit 141ec5e
  found: 18 focused alias, sanitizer, mismatch, mutation-table, aggregate-gate, and legacy-comparator tests pass after asserting exact foreign/by-pointer M004 refusal text.
  implication: the shared assertion and Phase 5 migration preserve the production cut; remaining contract failures are concentrated in Phase 4 session_test.go and Phase 6 result expectations.
- timestamp: "2026-09-22T19:45:00Z"
  checked: fresh focused Phase 4/6 reproduction
  found: Phase 4 direct runners fail with the exact foreign M004 refusal, Phase 4 aggregate gates stop at the release-omission lane, and the Phase 6 cleanup lane is operational because its production verifier still reaches the terminal public boundary; the admitted Phase 1 toggle lane separately reports semantic_mismatch.
  implication: migrate only the stale foreign execution assertions to terminal refusal, while investigating the admitted-toggle mismatch as a distinct Phase 17 adjacency rather than masking it.
- timestamp: "2026-09-22T19:55:00Z"
  checked: Phase 6 native differential mismatch
  found: the lane compared the interpreter's retained execution/0 document directly against native execution/2; RunNative and the Phase 5 corpus already project both sides through ProjectExecutionSchema2 before comparison.
  implication: projecting both documents in the Phase 6 lane restores the admitted toggle differential without weakening the comparator or the M004 cut.
- timestamp: "2026-09-22T20:00:00Z"
  checked: focused Phase 4 and Phase 6 tests after commit 776b49b
  found: all migrated direct runners, aggregate gates, frozen-evidence attribute tests, Phase 4 corpus paths, Phase 6 cleanup refusal, and Phase 6 admitted-toggle differential pass.
  implication: the remaining Phase 4/6 migration is complete and the authoritative test-only provenance/public-refusal split is preserved.
- timestamp: "2026-09-22T20:05:00Z"
  checked: go test ./... after commit 776b49b
  found: all packages outside internal/compiler/session pass; remaining session failures are Phase 11 by-pointer gates, concurrent Phase 15 corpus/planning drift, Phase 16 emitter-registry line/cardinality drift, and legacy owned-transfer/origin assertions outside the assigned Phase 4/6 contract migration.
  implication: the assigned migration is verified, but the repository-wide guardrail remains pending on separately owned follow-up work.
- timestamp: "2026-09-22T20:30:00Z"
  checked: legacy owned-transfer/origin assertions and Phase 16 public-emitter inventory after commit 31a0060
  found: focused ownership, origin, inventory, mutation-control, and M004 provenance tests pass; the session suite no longer reports any failure owned by the Phase 17 return fix or Phase 4–6 contract migration.
  implication: the requested return-contract and authoritative Phase 16 migration scope is complete without weakening production M004 refusal.
- timestamp: "2026-09-22T21:00:00Z"
  checked: final go test ./... after all scoped fixes
  found: every package except internal/compiler/session passed; the remaining session failures are exactly the separately owned Phase 11 by-pointer gates, Phase 15 corpus maturity/digest drift, stale PHASE-13 debt witness names, and planning groundedness findings. No Phase 17, Phase 4–6 migration, emitter-inventory, owned-transfer, origin, legacy digest, or four-tier differential failure remains.
  implication: close this session with the requested scope resolved and retain the exact independent failures for their owning debug/planning sessions.

## Resolution

root_cause: "Phase 17 omitted terminal nominal return validation in checker/corevalidate and cgen populated return nominal layouts from the parameter declaration; one-alternative fixtures and a fallback cast masked both defects."
fix: "Checker and core peer enforce terminal nominal return conformity; cgen derives return layouts from ReturnType; same-type functions preserve legacy type:0/event encoding while distinct contracts use type:1/match:return; schema canonicality and peer ability lookup handle mixed programs without baseline drift; stale Phase 4–6 tests now honor terminal production M004 refusal and use provenance-bound test-only historical C."
verification: "target_test: pass (focused Phase 17 checker/core/origin/cgen/session graph and four-tier differential); mutation_check: skipped (no Go mutation runner configured); no_op_deletion: pass; adjacent_tests: pass (legacy digests, Phase 4–6 M004/frozen-evidence consumers, owned-transfer/origin, public-emitter inventory); full_suite: all non-session packages pass, with session red only on separately owned Phase 11/15 planning/workspace drift; revert_and_reconfirm: pass by before/after focused reproduction; guardrail_verdict: accepted for requested scope with independent failures explicitly recorded"
oracle_type: "specified — declared return contracts and byte-stability pins"
files_changed: [internal/compiler/check/check.go, internal/compiler/core/core.go, internal/compiler/core/core_test.go, internal/compiler/corevalidate/corevalidate.go, internal/compiler/corevalidate/corevalidate_test.go, internal/compiler/cgen/cgen_program.go, internal/compiler/executionpeer/executionpeer.go, internal/compiler/interp/interp.go, internal/compiler/originvalidate/originvalidate.go, internal/compiler/reduce/reduce_test.go, internal/compiler/session/session_test.go, internal/compiler/session/session_phase5_alias_test.go, internal/compiler/session/session_phase5_corpus_test.go, internal/compiler/session/session_phase6_injectors_test.go, internal/compiler/session/session_phase6_verify_test.go, internal/compiler/session/session_phase16_emitter_inventory_test.go]

## Prevention

- why_not_caught: "The Phase 17 focused tests exercised the new distinct-return path but did not pair it with same-type byte-stability pins and the repository-wide Phase 16 production-refusal boundary in one gate."
- recurrence_guard: "The parameter-typed terminal rejection, mixed-schema canonicality, four-tier distinct-return differential, legacy digest pins, exact terminal M004 assertions, provenance-bound frozen-evidence helpers, and public-emitter inventory now jointly guard this class."
