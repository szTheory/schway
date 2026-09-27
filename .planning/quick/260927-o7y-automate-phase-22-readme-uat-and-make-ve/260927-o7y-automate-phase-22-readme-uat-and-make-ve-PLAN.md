---
phase: quick
plan: 260927-o7y
type: execute
wave: 1
depends_on: []
files_modified:
  - cmd/lang/main_test.go
  - examples/phase22/README.md
  - .planning/PROJECT.md
  - .planning/PRODUCT-ROADMAP.md
  - .planning/LANGUAGE-MATURITY.md
  - .planning/STATE.md
  - .planning/phases/22-native-application-build-and-single-execution/22-03-PLAN.md
  - .planning/phases/22-native-application-build-and-single-execution/22-03-SUMMARY.md
  - .planning/phases/22-native-application-build-and-single-execution/22-UAT.md
  - .planning/phases/22-native-application-build-and-single-execution/22-VERIFICATION.md
autonomous: true
must_haves:
  truths:
    - "A focused cmd/lang Go test fails when Phase 22 README command forms, bounds, failure states, C trust boundary, closure status, or modeled-world limits drift; its positive result is recurring in the existing go test ./... CI lanes."
    - "The test's claim is limited to the objective documentation contract and is backed by named CLI/native integration tests; no static text check is presented as proof of subjective readability."
    - "PROJECT.md makes early acceptance checks, proportionate evidence layers, negative and failure controls, justified CI recurrence, and narrow human handoffs durable GSD defaults."
    - "Phase 22's plan, summary, UAT, verifier, state, and living capability records agree about the new automated evidence, preserve the original human-review judgment as dated history, and retain all platform and semantic limits."
  artifacts:
    - path: cmd/lang/main_test.go
      provides: Focused Phase 22 README contract and regression controls
    - path: examples/phase22/README.md
      provides: Executable public command examples and bounded claims checked by the contract test
    - path: .planning/PROJECT.md
      provides: Durable shift-left verification operating preference
    - path: .planning/phases/22-native-application-build-and-single-execution/22-UAT.md
      provides: Dated resolution of the lone objective documentation UAT gate
  key_links:
    - from: .github/workflows/ci.yml
      to: cmd/lang/main_test.go
      via: Existing macOS/Linux go test ./... checks job
    - from: cmd/lang/main_test.go
      to: examples/phase22/README.md
      via: Contract categories grounded in implementation constants and named integration evidence
    - from: .planning/phases/22-native-application-build-and-single-execution/22-03-SUMMARY.md
      to: .planning/phases/22-native-application-build-and-single-execution/22-UAT.md
      via: Dated automated-evidence amendment and zero remaining objective UAT cases
---

<objective>
Replace Phase 22's mechanically checkable README UAT gate with recurring documentation-contract evidence and record the project's shift-left verification preference.

Purpose: Keep public command and safety-boundary claims aligned with the working CLI while closing the stale human handoff honestly.
Output: One focused Go test, a runnable evidence example, durable GSD policy, and consistent dated Phase 22 closeout records.
</objective>

<execution_context>
@~/.codex/gsd-core/workflows/execute-plan.md
@~/.codex/gsd-core/templates/summary.md
</execution_context>

<context>
@AGENTS.md
@.planning/PROJECT.md
@.planning/STATE.md
@.planning/PRODUCT-ROADMAP.md
@.planning/LANGUAGE-MATURITY.md
@.planning/phases/22-native-application-build-and-single-execution/22-03-PLAN.md
@.planning/phases/22-native-application-build-and-single-execution/22-03-SUMMARY.md
@.planning/phases/22-native-application-build-and-single-execution/22-UAT.md
@.planning/phases/22-native-application-build-and-single-execution/22-VERIFICATION.md
@examples/phase22/README.md
@.github/workflows/ci.yml
@cmd/lang/main_test.go
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Pin the Phase 22 public README contract in the existing Go suite</name>
  <files>cmd/lang/main_test.go, examples/phase22/README.md</files>
  <behavior>
    - A valid README names runnable build, app run, same-run evidence, and explicit replay command forms, including the manifest build and the argument separator.
    - Removing or contradicting each required contract category fails with a category-specific diagnostic.
    - A passing text contract claims documentation accuracy and scope only; existing named integration tests establish CLI/runtime behavior.
  </behavior>
  <action>Add `TestPhase22READMEContract` in package `cmd/lang`, loading `examples/phase22/README.md` through `testsupport.ProjectPath`. Check the documented command forms by normalizing shell-line continuations and tokens, including build, run, manifest build, an explicit `--report` plus `--evidence=events` run, and `app verify --cases ... --report ...`; add the missing runnable evidence example to the README. Check substantive contract categories with focused clauses and in-memory removal/contradiction controls: canonical U64 and 4,096-byte input, one process, 30-second timeout, transparent uncapped ordinary streams and distinct exit/signal/timeout/launch outcomes; 64 KiB evidence/case/report bounds and 16 MiB conformance documents; disabled/complete/incomplete/capacity-exhausted capture, write/capture failure, and `verified:false` for one-run evidence; closed local-C manifest inputs and 64 KiB/4 MiB/16 MiB limits, trusted declared C/compiler/SDK, fixed flags, and no Lang foreign/pointer admission; incomplete/non-cacheable host closure on macOS/Linux; source cases with independent answers and empty scripts versus verifier-model cases with false actual-host-IO/physical-cleanup claims. Derive numeric checks from exported `native`, `execution`, and `session` constants where available; link unexported manifest bounds and behavior claims to `TestPhase22BindingsRejectInvalidInputs`, `TestPhase22RunApplicationPreservesStreamsAndProcessOutcomes`, `TestPhase22EvidenceDisabledCompleteAndStreamIsolation`, and `TestPhase22AppVerifyIndependentIdentityCases` instead of implying the README parser proves runtime behavior. Keep D-22-02/D-22-04/D-22-05/D-22-06 separation and refusals unchanged. Existing `.github/workflows/ci.yml` runs `go test ./...` on both hosts, so this test needs no separate CI job.</action>
  <verify><automated>GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/lang -run '^TestPhase22READMEContract$|^TestPhase22(AppVerify|IdentityApplicationBuildAndRunCLI|AppRun)' -count=1</automated></verify>
  <done>The focused test passes for the documented contract, its in-memory negative controls fail by category, and the existing CLI/integration witnesses still pass. CI's existing go test ./... command reaches the new test.</done>
</task>

<task type="auto">
  <name>Task 2: Persist the verification default and refresh living capability status</name>
  <files>.planning/PROJECT.md, .planning/PRODUCT-ROADMAP.md, .planning/LANGUAGE-MATURITY.md</files>
  <action>Expand the existing `## Verification Operating Preference` in PROJECT.md in place: define acceptance checks during planning; choose reachable unit, seam, integration, end-to-end, smoke, negative-control, and failure-path checks as the claim warrants; run fast stable recurring checks in existing CI when regression value exceeds runtime and maintenance cost; avoid handing mechanically verifiable UAT to the user; retain human handoff only for irreducibly subjective judgments, real external systems/devices, or user-owned access/authority. Preserve mandatory gates and authorization. After Task 1 evidence exists, update PRODUCT-ROADMAP.md and LANGUAGE-MATURITY.md only where Phase 22's documentation-gate status changed, with a dated amendment that distinguishes source inspection, newly executed test evidence, and historical receipts. Preserve the current three recommendations and their program/blocker/slice/checker/evidence/owner/reprioritization details; retain the macOS-only observation, unrun Linux claim, incomplete/non-cacheable closure, and Phase 23/24/25 capability order.</action>
  <verify><automated>rg -q '^## Verification Operating Preference$' .planning/PROJECT.md &amp;&amp; rg -qi 'unit.*seam.*integration.*end-to-end.*smoke' .planning/PROJECT.md &amp;&amp; rg -qi 'negative.control.*failure.path' .planning/PROJECT.md &amp;&amp; rg -qi 'user-owned access|user-owned authority' .planning/PROJECT.md &amp;&amp; rg -q 'TestPhase22READMEContract' .planning/PRODUCT-ROADMAP.md .planning/LANGUAGE-MATURITY.md</automated></verify>
  <done>The project charter specifies the accepted verification policy once; living documents cite the new objective contract evidence without changing platform, runtime, or next-capability claims.</done>
</task>

<task type="auto">
  <name>Task 3: Amend Phase 22 closeout and regenerate its verdict</name>
  <files>.planning/phases/22-native-application-build-and-single-execution/22-03-PLAN.md, .planning/phases/22-native-application-build-and-single-execution/22-03-SUMMARY.md, .planning/phases/22-native-application-build-and-single-execution/22-UAT.md, .planning/phases/22-native-application-build-and-single-execution/22-VERIFICATION.md, .planning/STATE.md</files>
  <action>Add an explicit 2026-09-27 amendment to 22-03-PLAN and 22-03-SUMMARY recording the original reader-review judgment, the later user-approved shift to objective contract testing, `TestPhase22READMEContract` and the named behavior tests, and the remaining subjective limit. Update the summary's machine-readable D3 verification/human-judgment fields to describe the amended objective criterion while preserving its original judgment in the dated prose. Resolve 22-UAT's lone pending test as an objective documentation-contract check with the focused command and observed result; preserve its original wording and pending state as history in a dated amendment, and never record subjective readability as proven by phrase checks. After the source, plan, summary, and UAT changes are in place, invoke the existing typed `gsd-verifier` agent for Phase 22 through the `verify_phase_goal` workflow in `gsd-core/workflows/execute-phase.md`; have it regenerate 22-VERIFICATION with a new covered-input fingerprint, rather than editing or accepting the cached verdict. Require the regenerated report to cite the README contract test and its limited claim, retain all 17 implementation truths, distinguish the observed host from any unrun Linux lane, and preserve incomplete/non-cacheable SDK, linker, and runtime closure. Only after the agent returns, run the shared verification-status query and `phase uat-passed 22 --require-verification`; a stale, incomplete, or human-needed result blocks closeout. Refresh STATE.md from the actual GSD progress result after closeout rather than carrying forward the old pending-UAT route.</action>
  <verify><automated>GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/lang -run '^TestPhase22READMEContract$' -count=1 &amp;&amp; test "$(node ~/.codex/gsd-core/bin/gsd-tools.cjs query verification.status .planning/phases/22-native-application-build-and-single-execution --pick status)" = passed &amp;&amp; test "$(node ~/.codex/gsd-core/bin/gsd-tools.cjs phase uat-passed 22 --require-verification --pick passed)" = true &amp;&amp; rg -q '^score: 17/17 must-haves verified$' .planning/phases/22-native-application-build-and-single-execution/22-VERIFICATION.md &amp;&amp; rg -q '^covered_digest: "v1:sha256:' .planning/phases/22-native-application-build-and-single-execution/22-VERIFICATION.md &amp;&amp; rg -q 'TestPhase22READMEContract' .planning/phases/22-native-application-build-and-single-execution/22-03-SUMMARY.md .planning/phases/22-native-application-build-and-single-execution/22-UAT.md .planning/phases/22-native-application-build-and-single-execution/22-VERIFICATION.md &amp;&amp; rg -qi 'documentation.contract|README contract' .planning/phases/22-native-application-build-and-single-execution/22-VERIFICATION.md &amp;&amp; rg -qi 'subjective|reader clarity|readability' .planning/phases/22-native-application-build-and-single-execution/22-VERIFICATION.md &amp;&amp; rg -qi 'macOS' .planning/phases/22-native-application-build-and-single-execution/22-VERIFICATION.md &amp;&amp; rg -qi 'Linux[^.\n]*(unavailable|not (observed|run|claimed)|unrun|separate required host lane)' .planning/phases/22-native-application-build-and-single-execution/22-VERIFICATION.md &amp;&amp; rg -qi 'incomplete[^.\n]*non-cacheable|non-cacheable[^.\n]*incomplete' .planning/phases/22-native-application-build-and-single-execution/22-VERIFICATION.md &amp;&amp; rg -q '^pending: 0$' .planning/phases/22-native-application-build-and-single-execution/22-UAT.md</automated></verify>
  <done>The full Go suite passes; amended plan/summary/UAT and the newly generated fingerprinted verifier agree on objective coverage and historical intent; the shared status is passed and the required-verification UAT predicate passes with 17/17 truths, README evidence, and retained host/closure caveats; the lone pending UAT is closed without a claim of tested subjective readability, and STATE names the real next route.</done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Public README to developer | Incorrect commands or overbroad C/evidence claims can mislead a reader into unsafe or unsupported use. |
| Static documentation test to phase verdict | Text checks can establish required contract facts but cannot certify subjective readability or unseen host behavior. |

## STRIDE Threat Register (ASVS L1; high blocks)

| Threat ID | Category | Component | Severity | Disposition | Mitigation Plan |
|-----------|----------|-----------|----------|-------------|-----------------|
| T-QO7Y-01 | Tampering | README command and boundary claims | medium | mitigate | Task 1 checks the command/limit/trust categories with reached negative controls and anchors runtime claims to named integration tests. |
| T-QO7Y-02 | Repudiation | Phase 22 UAT and summary | medium | mitigate | Task 3 retains a dated record of the original reader judgment, cites newly run checks, and regenerates the fingerprinted verdict. |
| T-QO7Y-03 | Spoofing | Host/evidence completeness claims | high | mitigate | Tasks 1–3 retain `verified:false` for one-run capture, false model-world IO/cleanup claims, and incomplete non-cacheable host closure. |
</threat_model>

<source_audit>
| Source | Item | Task | Status |
|--------|------|------|--------|
| GOAL | Replace the lone mechanical README UAT with recurring objective evidence | 1, 3 | COVERED |
| REQ | Public commands, bounds, failure states, trusted local C, closure, and model limits | 1 | COVERED |
| REQ | Durable early verification, proportionate CI, and narrow human handoff policy | 2 | COVERED |
| REQ | Historical amendment and consistent Phase 22 closeout/progress | 2, 3 | COVERED |
| RESEARCH | Existing standard-library CLI tests and macOS/Linux go test ./... CI coverage | 1, 3 | COVERED |
| CONTEXT | D-22-02/D-22-04/D-22-05/D-22-06 boundaries retained; no new runtime capability | 1–3 | COVERED |
| PRE-HOOK | No external API, architectural cardinality change, or schema file in scope; ASVS L1 applies | 1–3 | COVERED |
</source_audit>

<verification>
Run the focused documentation and CLI test, then the full Go suite with the workspace GOCACHE. Confirm the existing CI `checks` job runs `go test ./...` on macOS and Linux. Treat a CI run as evidence only if actually observed; local execution establishes this host alone.
</verification>

<success_criteria>
Phase 22's public contract is checked on every existing CI test run, its former UAT gate has objective passing evidence and no pending case, and the project policy and closeout records state the evidence's exact limits.
</success_criteria>

<output>
Create `.planning/quick/260927-o7y-automate-phase-22-readme-uat-and-make-ve/260927-o7y-SUMMARY.md` when done.
</output>
