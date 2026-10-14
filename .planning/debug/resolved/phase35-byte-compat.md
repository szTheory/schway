---
status: resolved
trigger: "Phase35 unused-Byte validation rejects historical Byte/resource programs"
created: 2026-10-14
updated: 2026-10-14
---

## Symptoms

expected: The frozen Phase35 pre-change corpus retains all old source/core/identity and three-emitter outcomes, except the already-authored exact malformed-numeric delta.
actual: The binding comparator fails with 158 mismatch clauses across 34 old fixtures.
error: 20 fixtures receive core.scalar_byte_parameter_used and 14 receive core.callee_not_callable before C emission.
timeline: The corpus was frozen at 1fd69337; current C tasks b64d2159/ed8475b9 pass their focused selectors, but upstream Phase35 validator guards changed historical admission.
repro: GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase35PreChangeEmitterCorpus$' -count=1 -v -timeout=600s

## Current Focus

hypothesis: The new unused-Byte guard is applied to every legacy Byte function rather than the narrow scalar bridge.
test: Inspect corevalidate's unconditional Byte guard and originvalidate's phase35UnusedByteProblem publication precheck; reproduce frozen corpus.
expecting: Narrowing the bridge discriminator preserves legacy contracts while existing forged scalar-Byte controls still reject.
bug_class: Bohrbug (deterministic admission regression)
known_pattern_candidate: phase25-pathoracle-regression (legacy interpretation overreach); tested against current code rather than assumed
sbfl: skipped; no per-test coverage available
next_action: manager commits the four exact validator paths, scoped receipt and resolved debug artifact; continue authorized Phase35 work under its own plan
reasoning_checkpoint:
  hypothesis: The scalar-only unused-Byte guard changes legacy admission because Byte parameter type alone selects it in both peers.
  confirming_evidence:
    - Fresh corpus repro fails before C emission; new agent-authored LegacyByteCompatibility tests fail in both peers on copy, foreign/resource, zero-call, and callable fixtures.
    - Source producer always emits reserved Bool at :type:2 and Unit at :type:3 in scalar Byte IR; legacy signature/literal producer does not emit that table.
  falsification_test: If legacy compatibility or the frozen corpus still fails after independently narrowing both guards, the guards do not fully explain this regression.
  fix_rationale: Applicability follows structured-scalar IR evidence, while unchanged eligibility checks reject used Byte, wrong return and non-scalar locals; constant-only forged bodies retain type-table evidence.
  blind_spots: Coordinated removal of all scalar instructions and reserved scalar type identity is indistinguishable from legacy IR to a source-blind peer; no wider source authority is claimed.
  candidate_causes:
    - code: overbroad Byte guard and callable derivation
    - data: frozen corpus or malformed numeric delta drift
    - environment: host emitter variation
  and_gate: No; the independent peer overreach explains upstream admission and emission mismatches without emitter or fixture changes. Corpus controls will check the data alternative.

## Evidence

- 2026-10-14: Final checks after last code/test edit passed with no skips: core six named checks 0.042s; origin six 0.026s; session corpus and two controls 2.590s; four Cgen checks 0.652s. Logs /tmp/phase35-byte-compat-final-{corevalidate,originvalidate,session,cgen}.log. git diff --check passed.
- 2026-10-14: Manager states prior user auto-follow authorization covers deterministic auto-verification and directs resolved artifact without another human checkpoint; exact-path commits remain manager-owned.

- 2026-10-14: Revert-and-reconfirm restored only the old applicability behavior, leaving tests and eligibility laws unchanged. Agent-authored legacy regression failed in core (0.006s) and origin (0.007s); logs /tmp/phase35-byte-compat-revert-corevalidate.log and /tmp/phase35-byte-compat-revert-originvalidate.log. Narrow conditions reapplied; final fresh checks pending.

- 2026-10-14: First post-fix runs all passed: six named core peer tests (0.043s), six named origin peer tests (0.025s), corpus plus two controls (2.563s), four Cgen boundary tests (0.679s). Constant-only fake bridges, unreachable uses, wrong return and local types retain independent refusals.

- 2026-10-14: Agent-authored LegacyByteCompatibility tests fail in both peers before the fix (core 0.007s, origin 0.007s). Independent failures specifically observe old Byte copy, fallible acquisition, nonlocal foreign and zero-call identity; ordinary call fixture loses callable admission.
- 2026-10-14: Manager reviewed direction and authorized narrow fix without approval checkpoint. No matching Go specialist skill; continue inline. No commit ownership delegated to this agent.

- 2026-10-14: Fresh TestPhase35PreChangeEmitterCorpus execution failed in 1.303s, reporting legacy Byte fixtures rejected before emitter comparison. Includes identity, ordinary calls and resource/foreign fixtures; no emitter edit can fix source admission.
- 2026-10-14: git log identifies 26e4b509 (35-04) as the guard introduction. corevalidate linearShape unconditionally calls scalarByteFormalUnused for Byte; origin PublishProblemsFor unconditionally calls phase35UnusedByteProblem for every linear Byte function.
- 2026-10-14: Source inspection of check/scalar.go proves scalar Byte producer always emits ordered Byte/U64/Bool/Unit facts, including constant-only bodies; legacy check.go uses signature facts and appends literal U64 at :type:u64. Applicability must retain scalar IR evidence independently of return type, local type and Byte uses.

- Parent source inspection: corevalidate.go line1213 checks every Byte parameter; originvalidate.go phase35UnusedByteProblem starts with every linear Byte parameter, before legacy publication.
- Plan35-06 is incomplete: both tasks committed, no SUMMARY or state update; no task replay is authorized.
- User explicitly delegated coherent recommendations and auto-follow, then requested retry after a capacity message. Routine regression fixes are authorized; no invented approval is needed.

## Constraints

Ownership: corevalidate.go/corevalidate_phase35_test.go and originvalidate.go/originvalidate_phase35_test.go only, plus this debug session and a scoped Phase35 compatibility-fix receipt.
Keep frozen fixtures and source untouched; no broad baseline exemptions. Validate raw malformed new bridge IR, including used formal, wrong return/type/non-scalar locals/unreachable use. Independent peers keep independent laws.
Use focused named Go selectors only, always GOCACHE=/tmp/ai-lang-verification-gocache. No full local suite, whole-package suite, broad native matrix, dependency/backend, push/hosted actions or Phase36 work.
Root pin: bash /tmp/schway34-root-pin.sh before edits/every commit. SDK node /Users/jon/.codex/gsd-core/bin/gsd-tools.cjs query commit exact paths with require_escalated when necessary. Preserve user config edits. Do not mutate main STATE/ROADMAP or mark requirements complete here.

## Resolution

root_cause: Phase35 applied its scalar-only unused-Byte restriction to all legacy linear Byte functions in both independent validators, changing admission and callable derivation before emission.
fix: Scope the Byte precheck using independently derived scalar operation or reserved scalar type-table evidence; retain every existing bridge eligibility law. Add independent legacy and constant-only forged bridge controls.
oracle_type: specified (frozen historical contracts and bridge admission laws)
verification:
  target_test: {result: pass, tests: both agent-authored TestPhase35LegacyByteCompatibility and original TestPhase35PreChangeEmitterCorpus}
  mutation_check: {result: skipped, reason_if_skipped: Stryker not configured for this Go repository; manual old-applicability mutant independently killed by both legacy tests}
  no_op_deletion: {result: pass, deletion_justified_by_rca: true, detail: only applicability changes; every scalar bridge eligibility law remains intact and adversarial controls execute}
  adjacent_tests: {result: pass, suites_run: [core six named Phase35 tests, origin six named Phase35 tests, session corpus plus two controls, four Cgen Phase35 boundary tests], skips: 0}
  revert_and_reconfirm: {result: pass, bug_returned_on_revert: true, fixed_on_reapply: true, detail: only original Byte applicability conditions restored and both legacy tests failed; final fresh focused selections pass after reapplication}
  guardrail_verdict: accepted
prevention:
  why_not_caught: 35-04 focused scalar controls did not cover historical Byte admission; the frozen corpus comparator caught the regression.
  recurrence_guard: independent TestPhase35LegacyByteCompatibility in both peers, constant-only fake bridge subtests under TestPhase35UnusedByteIRAdmission, and frozen corpus with both named controls
  knowledge_base: manager-owned archive followup; this agent's authorized ownership excludes knowledge-base.md
files_changed: [internal/compiler/corevalidate/corevalidate.go, internal/compiler/corevalidate/corevalidate_phase35_test.go, internal/compiler/originvalidate/originvalidate.go, internal/compiler/originvalidate/originvalidate_phase35_test.go]
