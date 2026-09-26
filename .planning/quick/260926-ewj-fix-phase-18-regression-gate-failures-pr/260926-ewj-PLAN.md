---
id: 260926-ewj
phase: quick
plan: 260926-ewj
type: execute
mode: quick
status: planned
wave: 1
depends_on: []
files_modified:
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_program.go
  - .planning/LANGUAGE-MATURITY.md
autonomous: true
must_haves:
  truths:
    - "Historical Phase 1 and owned-transfer generated C remains byte-identical to its committed golden, with unchanged Phase 1 evidence identity and C digest."
    - "The schema-2 computed-match terminal writer still escapes and streams a long alternative tag and its runtime payload within the bounded output contract."
    - "The generated JSON string-content identifier is reserved in both emitter namespaces."
    - "The current maturity snapshot states 142 .lang programs and 4,661 lines, and its independent self-check passes."
    - "The reported focused regressions and the full Go suite pass."
  artifacts:
    - path: internal/compiler/cgen/cgen.go
      provides: Conditional schema-2 content helper with historical writer spelling and complete fixed-name reservations
    - path: internal/compiler/cgen/cgen_program.go
      provides: Terminal-writer need passed to the schema-2 support emitter
    - path: .planning/LANGUAGE-MATURITY.md
      provides: Current corpus snapshot matching the independent source-tree scan
  key_links:
    - from: internal/compiler/cgen/cgen_program.go
      to: internal/compiler/cgen/cgen.go
      via: The checked entry return branch selects the content helper only when its payload terminal writer consumes it
    - from: internal/compiler/cgen/cgen.go
      to: internal/compiler/cgen/cgen_names_test.go
      via: Both generated-identifier reservation sets cover the content helper
    - from: .planning/LANGUAGE-MATURITY.md
      to: internal/compiler/session/self_describing_docs_test.go
      via: TestLanguageMaturityCountsAreCurrent independently derives corpus totals
---

<objective>
Repair the Phase 18 full-suite regressions introduced by terminal JSON streaming and one added source fixture.

Purpose: Retain pinned historical generated C and evidence identities while keeping the accepted long-payload behavior and an accurate corpus inventory.
Output: Scoped emitter and maturity-document edits, followed by focused and full-suite verification.
</objective>

<execution_context>
@~/.codex/gsd-core/workflows/execute-plan.md
@~/.codex/gsd-core/templates/summary.md
</execution_context>

<context>
@AGENTS.md
@.planning/STATE.md
@.planning/phases/18-branch-on-a-computed-value/18-10-PLAN.md
@internal/compiler/cgen/cgen.go
@internal/compiler/cgen/cgen_program.go
@internal/compiler/cgen/cgen_names_test.go
@internal/compiler/cgen/cgen_program_test.go
@internal/compiler/session/self_describing_docs_test.go
@.planning/LANGUAGE-MATURITY.md

The completed `go test ./...` reported `TestGeneratedIdentifierNamespacesStayConfined/{match,linear_Buffer,linear_Byte}` at `cgen_names_test.go:218` with unreserved identifier `[lang_write_json_string_content]`; `TestLegacyEventWritersFrozen` at `cgen_program_test.go:1102` with `legacy /0 event writer bytes changed`; `TestEmitProgramSingleFunctionBytesUnchanged` at `cgen_program_test.go:1273` with `single-function output moved`; and `TestOwnedTransferInterpreterNative` with `owned C golden changed`. `TestPhase1EvidenceGoldenUnchanged` and `TestCanonicalEvidence` observed actual `id=evidence:ded89497418f89e44c54a667`, `c_digest=sha256:abeae8190c35368bdf6e0ceb5d4713b40a3ba5887436dd4fcea02367f8b7c1fa`, versus pinned `id=evidence:82c1c6d6e2ed63634e6d7208`, `c_digest=sha256:1fd8aff8ee28de7ec39e559a7ca9ce50e480ecfffede617c36b2282c60cc122a`. `TestLanguageMaturityCountsAreCurrent` and two `TestSelfDescribingDocsGuardIsNotInert` subtests reported `.planning/LANGUAGE-MATURITY.md:87` stated 141/4,641 versus derived 142/4,661.

The plan:pre detectors returned structured `detected:false` for external API coverage and assumption delta. The Go/C/document scope contains no ORM or database schema path, so schema push does not apply.
</context>

<tasks>

<task type="auto">
  <name>Task 1: Keep historical C bytes while reserving the terminal helper</name>
  <files>internal/compiler/cgen/cgen.go, internal/compiler/cgen/cgen_program.go</files>
  <action>In `emitEventSupportSchema2`, emit `lang_write_json_string_content` and its quote/content/quote wrapper only when the checked entry's branch return has a payload and `emitProgramTerminalValueWriter` will consume that helper. For other entries, preserve the prior inline `lang_write_json_string` emission byte for byte, including its existing escaping and bound checks; the schema-1 legacy `emitEventSupport` remains unchanged. Derive the predicate from the same `entryReturnIsBranch && entryReturnBranch.hasPayload` condition used by the terminal writer in `emitProgram`, and pass it to the support emitter before C serialization so helper emission and consumption cannot diverge. Add `lang_write_json_string_content` to both `matchFixedNames` and `linearFixedNames`, since the namespace test checks both and over-reservation is inert. Keep the long-tag schema-2 streaming path, output bound, and runtime payload read intact. Preserve committed C goldens, evidence pins, source fixtures, UAT, and validation records.</action>
  <verify>
    <automated>GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen ./internal/compiler/session ./internal/compiler/evidence -run '^(TestGeneratedIdentifierNamespacesStayConfined|TestLegacyEventWritersFrozen|TestEmitProgramSingleFunctionBytesUnchanged|TestOwnedTransferInterpreterNative|TestPhase1EvidenceGoldenUnchanged|TestCanonicalEvidence|TestPhase18LongPayloadPlaceReturn|TestPhase18LongTagWrongSlotMutation)$' -count=1</automated>
  </verify>
  <done>The previously failing namespace, C golden, and evidence tests pass with their original pinned bytes and digests, while the long-tag terminal and wrong-slot controls still pass.</done>
</task>

<task type="auto">
  <name>Task 2: Refresh the current maturity corpus snapshot</name>
  <files>.planning/LANGUAGE-MATURITY.md</files>
  <action>Update only the current 2026-09-24 corpus re-assessment from 141 programs and 4,641 lines to the independently derived 142 programs and 4,661 lines. The rounded average remains about 33 lines and the stated 193-line maximum remains intact. Keep historical dated snapshots, guard inventory, and language-capability claims unchanged. The added Phase 18 source fixture is an input to the count, not a file to edit. After both tasks, run the focused documentation self-check and the exact full Go suite with the writable cache path; record actual outcomes in the quick summary.</action>
  <verify>
    <automated>GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestLanguageMaturityCountsAreCurrent|TestSelfDescribingDocsGuardIsNotInert)$' -count=1</automated>
    <automated>GOCACHE=/tmp/ai-lang-verification-gocache go test ./...</automated>
    <automated>git diff --check -- internal/compiler/cgen/cgen.go internal/compiler/cgen/cgen_program.go .planning/LANGUAGE-MATURITY.md</automated>
  </verify>
  <done>The current snapshot reads 142 programs and 4,661 lines, the maturity self-check and non-inert controls pass, the full Go suite passes, and the scoped diff has no whitespace errors.</done>
</task>

</tasks>

<threat_model>
ASVS level 1 applies to this quick repair.

## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Checked source to generated C | A source-derived alternative name and runtime payload become a native JSON terminal value. |
| Generated C to evidence identity | Any historical C byte change alters pinned digests and evidence IDs. |
| Repository corpus to maturity document | An inaccurate count makes the documented evidence diverge from the source tree. |

## STRIDE Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation Plan |
|-----------|----------|-----------|----------|-------------|-----------------|
| T-QEWJ-01 | Tampering | Schema-2 terminal JSON | high | mitigate | Emit the content helper exactly when the payload terminal writer consumes it; retain escaping and bounded writes and run long-tag and wrong-slot controls. |
| T-QEWJ-02 | Repudiation | Historical generated C and evidence identity | medium | mitigate | Restore frozen C spelling for entries without a payload terminal writer and require golden and evidence pin tests. |
| T-QEWJ-03 | Spoofing | Generated C identifier namespace | medium | mitigate | Reserve the helper in both fixed-name sets and require the namespace confinement test. |
| T-QEWJ-04 | Tampering | Maturity corpus snapshot | low | mitigate | Use the independent corpus scan and its non-inert controls to verify the stated 142/4,661 totals. |
</threat_model>

<source_audit>
| Source | Item | Task | Status |
|--------|------|------|--------|
| GOAL | Repair the reported Phase 18 full-suite regressions | 1, 2 | COVERED |
| REQ | Preserve historical C bytes and evidence pins | 1 | COVERED |
| REQ | Keep schema-2 long-payload behavior and reserve the helper | 1 | COVERED |
| REQ | Update the maturity corpus to 142 programs and 4,661 lines | 2 | COVERED |
| REQ | Rerun focused failures and the full Go suite | 1, 2 | COVERED |
| RESEARCH | No research artifact required for this local regression repair | — | EXCLUDED |
| CONTEXT | Parent quick-task scope and completed test evidence | 1, 2 | COVERED |
</source_audit>

<verification>
Run Task 1's named regression and long-tag controls, Task 2's independent maturity controls, then `GOCACHE=/tmp/ai-lang-verification-gocache go test ./...`. Inspect the changed-file list and diff for only the three declared files. The completed UAT, validation artifacts, committed goldens, and original Phase 18 fixture are outside this repair.
</verification>

<success_criteria>
All reported failures disappear, historical C/evidence pins retain their original values, schema-2 long payloads remain correct, the maturity count matches the tree, and the full Go suite passes.
</success_criteria>

<output>
Create `.planning/quick/260926-ewj-fix-phase-18-regression-gate-failures-pr/260926-ewj-SUMMARY.md` when done.
</output>
