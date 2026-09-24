---
id: 260924-kto
phase: quick
plan: 260924-kto
type: execute
mode: quick-full
status: planned
wave: 1
depends_on: []
files_modified:
  - internal/compiler/check/check.go
  - internal/compiler/check/check_phase18_test.go
  - internal/compiler/check/check_test.go
  - internal/compiler/cgen/cgen_payload_tracer_test.go
autonomous: true
must_haves:
  truths:
    - "All pinned Phase 1–5 core bytes and manifest IDs retain their recorded values, including the seven Phase 3–5 fixtures that regressed."
    - "A computed terminal match that returns a distinct data type still initializes and returns the selected arm value place."
    - "The S-010 production fixture retains its one-arm point and sibling-edge endpoints, and a valid both-arm control proves that endpoint classification changes."
    - "The payload three-engine tracer and the full Go suite pass with GOCACHE=/tmp/ai-lang-gocache."
  artifacts:
    - path: internal/compiler/check/check.go
      provides: Computed-terminal-only arm value-place emission with historical lowering preserved
    - path: internal/compiler/check/check_phase18_test.go
      provides: Valid S-010 controls and computed-match regression assertions
    - path: internal/compiler/check/check_test.go
      provides: Both-arm core CFG control if source ownership rules cannot express it
  key_links:
    - from: internal/compiler/check/check.go
      to: internal/compiler/core/core_test.go
      via: Historical serialized core and manifest pins
    - from: internal/compiler/check/check.go
      to: internal/compiler/check/check_phase18_test.go
      via: Selected-arm ValuePlaceID and loan endpoints
    - from: internal/compiler/cgen/cgen_payload_tracer_test.go
      to: internal/compiler/native/native.go
      via: Three-engine run with a bounded Clang runner timeout
---

<objective>
Restore the full Go suite after Phase 18 while preserving historical core bytes and current computed-match behavior.

Purpose: Keep source-to-core identity and the S-010 endpoint proof trustworthy across the whole repository.
Output: A scoped checker correction, a valid both-arm control, and evidence-driven handling of the payload tracer timeout.
</objective>

<execution_context>
@~/.codex/gsd-core/workflows/execute-plan.md
@~/.codex/gsd-core/templates/summary.md
</execution_context>

<context>
@AGENTS.md
@.planning/phases/18-branch-on-a-computed-value/18-CONTEXT.md
@internal/compiler/check/check.go
@internal/compiler/check/check_phase18_test.go
@internal/compiler/check/check_test.go
@internal/compiler/core/core_test.go
@internal/compiler/cgen/cgen_payload_tracer_test.go
@internal/compiler/native/native.go
@testdata/phase18/loan_across_branch.lang
@testdata/phase18/result_computed_match.lang
</context>

<tasks>

<task type="tracer" tdd="true">
  <name>Task 1: Preserve historical core bytes while returning computed arm values</name>
  <files>internal/compiler/check/check.go, internal/compiler/check/check_phase18_test.go</files>
  <behavior>
    - The seven failing Phase 3–5 fixtures retain their pinned core SHA-256 digests and evidence manifest IDs; do not edit the pin table or golden files.
    - The Phase 18 Result-returning computed terminal match still gives its bare selected arm a nonempty ValuePlaceID, a matching place of the return type, and an OpReturn reading that place.
    - A parameter-scrutinee historical match retains its original place and operation serialization even when return and scrutinee types differ.
  </behavior>
  <action>Trace the seven digest differences to `analyzePayloadArm` and its `returnTypeID != sourceTypeID` branch. Pass the computed-terminal provenance already known in `checkBranch` (`prefixes` contains the linear prefix) into bare-arm analysis or apply an equally explicit condition at the call site. Emit `ValuePlaceID` and the selected-arm result place only for the computed terminal match that needs this Phase 18 behavior. Preserve historical operation IDs, place IDs, ordering, JSON omission behavior, and the existing Result computed-match return path. Add a focused regression assertion in check_phase18_test.go for the distinct-return-type selected arm. Do not change historical hashes, manifest IDs, fixture content, or validators to mask changed core.</action>
  <verify>
    <automated>GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/check ./internal/compiler/core -run '^(TestPhase18ResultComputedMatchChecker|TestPreviousPhaseCoreBytesUnchanged|TestPreviousPhaseManifestIDsUnchanged)$' -count=1</automated>
  </verify>
  <done>Both historical pin tests and the computed Result checker test pass with the original pinned values.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Replace the invalid both-arm S-010 anti-control</name>
  <files>internal/compiler/check/check_phase18_test.go, internal/compiler/check/check_test.go</files>
  <behavior>
    - The accepted source fixture still has exactly one point endpoint at the On-arm use and one edge endpoint toward Off.
    - The both-arm witness contains an actual reference to the same pre-branch loan in each successor and has no inherited one-arm edge endpoint.
    - The neither-arm witness remains valid and has no inherited one-arm edge endpoint.
  </behavior>
  <action>The current `both arms` string replacement adds a second `take view`, which the source ownership checker rejects with `ownership.use_after_move`; replace it with a source shape the checker admits only if it still proves two successor uses of the same pre-branch loan. Otherwise remove that source subcase and add an explicit, valid core-level two-arm CFG control in same-package check_test.go using `loanLivenessFixpoint` and `materializeLoanEndpoints` (following `TestEdgeSpecificLiveOut`), asserting actual uses in both arms, no one-arm edge endpoint, and the expected point endpoints. Keep `TestPhase18LoanAcrossBranchFixture`'s accepted production fixture and neither-arm source control. Require every control to assert that its intended source or core topology was constructed; never treat a diagnostic, missing function, or empty endpoint set as success.</action>
  <verify>
    <automated>GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/check -run '^(TestPhase18LoanAcrossBranchFixture|TestEdgeSpecificLiveOut|TestPhase18BothArmLoanEndpointControl)$' -count=1</automated>
  </verify>
  <done>The one-arm fixture and both valid anti-controls pass, and the both-arm check discriminates two real uses of one loan from the one-arm edge result.</done>
</task>

<task type="auto">
  <name>Task 3: Reproduce tracer timing and close the full CI test gate</name>
  <files>internal/compiler/cgen/cgen_payload_tracer_test.go</files>
  <action>Run `TestPayloadTracerThreeEngineAgreement` repeatedly in isolation and then in the full suite with the same `GOCACHE` path. Capture whether the 5-second `native.DefaultRunner()` budget is exceeded only under parallel package load. If the failure recurs and elapsed/tool-error evidence identifies that budget, set a justified, bounded timeout on this test's runner in cgen_payload_tracer_test.go; keep native.DefaultRunner and production timeout semantics unchanged, and preserve all three-engine agreement and payload-tag assertions. If the tracer passes and no timing evidence supports a change, leave the test file untouched. Finish with the exact full-suite command below. Inspect `git diff --check` and changed-file scope; preserve all unrelated modifications and untracked files.</action>
  <verify>
    <automated>GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/cgen -run '^TestPayloadTracerThreeEngineAgreement$' -count=5</automated>
    <automated>GOCACHE=/tmp/ai-lang-gocache go test ./...</automated>
    <automated>git diff --check -- internal/compiler/check/check.go internal/compiler/check/check_phase18_test.go internal/compiler/check/check_test.go internal/compiler/cgen/cgen_payload_tracer_test.go</automated>
  </verify>
  <done>The tracer remains a three-engine semantic check, any timeout adjustment has reproduced timing evidence, and the full Go suite passes without golden or unrelated-file changes.</done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Source to checked core | New place creation can silently alter historical evidence identities. |
| CFG witness to endpoint assertion | An invalid control can appear to prove a liveness property without reaching endpoint derivation. |
| Clang runner to CI test | A loaded host can exceed a per-invocation deadline while semantic agreement remains correct. |

## STRIDE Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation Plan |
|-----------|----------|-----------|----------|-------------|-----------------|
| T-QKTO-01 | Tampering | Core serialization and manifest identity | high | mitigate | Preserve pinned hashes and require both historical pin tests after restricting ValuePlaceID creation. |
| T-QKTO-02 | Repudiation | S-010 anti-control | medium | mitigate | Require admitted source or valid explicit CFG with two uses of the same loan and exact endpoint assertions. |
| T-QKTO-03 | Denial of service | Native test runner | low | mitigate | Adjust only the test-local bounded timeout after reproducing a deadline failure under load. |
</threat_model>

<source_audit>

| Source | Item | Task | Status |
|--------|------|------|--------|
| GOAL | Restore full-suite regression gate after Phase 18 | 1–3 | COVERED |
| REQ | Preserve seven historical Phase 3–5 core hashes and manifest IDs | 1 | COVERED |
| REQ | Limit ValuePlaceID to computed terminal matches without losing Result return | 1 | COVERED |
| REQ | Repair both-arm S-010 anti-control without weakening endpoint proof | 2 | COVERED |
| REQ | Reproduce payload tracer timeout and adjust only with evidence | 3 | COVERED |
| REQ | Require exact full GOCACHE Go suite green and preserve unrelated files | 3 | COVERED |
| RESEARCH | No separate quick-task research artifact; current code and failures identify the seams | — | EXCLUDED |
| CONTEXT | D-18-01, D-18-03, D-18-04, D-18-05 computed match, S-010, Result evidence, recurring checks | 1–3 | COVERED |
</source_audit>

<verification>Run the focused checker/core pins, the S-010 controls, repeated isolated tracer, and `GOCACHE=/tmp/ai-lang-gocache go test ./...`. Inspect the scoped diff for accidental golden or unrelated-file edits.</verification>
<success_criteria>The original historical evidence pins, Phase 18 Result return, and S-010 controls pass; the tracer timing disposition is evidence based; the exact full-suite command is green.</success_criteria>
<output>Create `.planning/quick/260924-kto-restore-historical-core-compatibility-an/260924-kto-SUMMARY.md` when done.</output>
