---
status: resolved
trigger: "Fix Phase 15 schema-2 preflight ordering regression caught by TestLTOInertnessOnMultiFunctionEmission."
created: "2026-09-19"
updated: "2026-09-19T01:30:00-04:00"
---

# Debug Session: Phase 15 LTO Diagnostic Order

## Symptoms

- Expected: unsupported `Match` validation reports its established `unsupported linear C type \"Switch\"` diagnostic.
- Actual: Phase 15 `emitProgram` runs schema-2 preflight/size estimation first, changing the rejection path before unsupported-form validation.
- Reproduction: `go build ./... && go vet ./... && go test ./... -count=1`, failing `TestLTOInertnessOnMultiFunctionEmission`.

## Current Focus

- hypothesis: "schema2ExecutionDocumentSize calls linearInput(entry) before emitProgram's unsupported-body validation, so a Match entry's Switch parameter error masks the established multi-function branch-body refusal."
- known_pattern_candidate: "phase15-native-capacity — exact schema-2 size preflight added to cgen emitProgram"
- bug_class: "Bohrbug — deterministic diagnostic precedence regression"
- sbfl: "skipped — there is one deterministic failing test but no per-test coverage spectrum in the workflow"
- reasoning_checkpoint:
    hypothesis: "emitProgram runs occurrence/output preflight before validating its supported body shapes; schema2ExecutionDocumentSize therefore calls linearInput on a Match entry and returns the incidental Switch-type error before the established Match-body refusal."
    confirming_evidence:
      - "The focused session witness and new cgen regression both deterministically return unsupported linear C type \"Switch\" instead of the named branch-body refusal."
      - "Direct control-flow inspection shows schema2ExecutionDocumentSize before the function.Match check, and commit 00742ee introduced the estimator call at that exact location."
    falsification_test: "If moving only structural body-shape validation ahead of preflight does not green both diagnostic tests, or if occurrence/output boundary controls stop refusing before serialization, this hypothesis is false or the fix is invalid."
    fix_rationale: "Validate and collect the ordered supported function set immediately after graph/entry resolution, then run the unchanged occurrence-table, event-capacity, and exact output-size preflights. This restores diagnostic precedence without bypassing or weakening any safeguard for supported programs."
    blind_spots: "Only the tracked Match fixture exercises unsupported-shape precedence; foreign-block and foreign-contract branches are structurally adjacent but not the reported failure. Cross-host behavior is not relevant to this pure deterministic Go control flow."
    candidate_causes:
      - "code: schema-2 estimator is ordered before emitProgram's supported-shape validation"
      - "config: executionOutputLimit could change a refusal only after estimation, but cannot produce the observed Switch error"
      - "data: the valid checked fixture has a Match entry with Switch type, which triggers the ordering defect but is not malformed input"
      - "environment: no host-dependent operation occurs before the deterministic error"
    and_gate: "no — the valid Match fixture is the triggering equivalence class, while the single code-ordering defect fully accounts for the wrong diagnostic."
- test: "Completed."
- expecting: "Satisfied."
- next_action: "Commit the two-file code fix atomically, archive this resolved session, append the knowledge-base prevention entry, and commit planning docs separately."

## Evidence

- timestamp: "2026-09-19T00:00:00-04:00"
  checked: "Required Phase 15 verification artifact path"
  found: "The supplied path .planning/phases/15-native-occurrence-identity-and-graph-evidence/15-VERIFICATION.md does not exist in the current worktree."
  implication: "Resolve the actual Phase 15 artifact location before relying on its verification claims."
- timestamp: "2026-09-19T00:05:00-04:00"
  checked: "Phase artifact discovery and repository status"
  found: "The verification artifact is .planning/phases/15-event-identity-lang-execution-2/15-VERIFICATION.md. The debug session and Phase 15 review/verification artifacts are untracked; unrelated planning state/config files are modified."
  implication: "Preserve unrelated user-owned planning changes and limit any code commit to the targeted source/test files."
- timestamp: "2026-09-19T00:10:00-04:00"
  checked: "Phase 15 verification report and failing witness definition"
  found: "The report records a deterministic focused failure: EmitNative returns unsupported linear C type \"Switch\" instead of the established multi-function branch-body refusal. The witness explicitly requires the named branch-body restriction."
  implication: "This is a deterministic diagnostic-precedence Bohrbug; the public refusal contract must be restored without deleting the schema-2 safeguards."
- timestamp: "2026-09-19T00:15:00-04:00"
  checked: "Phase 0 knowledge-base recall"
  found: "The matching prior resolution phase15-native-capacity added schema2ExecutionDocumentSize and pre-serialization output refusal in cgen_program.go; its recurrence guards require those safeguards to remain live."
  implication: "Test the preflight insertion as the first candidate cause, but any ordering fix must retain path-table, event-capacity, and output-bound behavior."
- timestamp: "2026-09-19T00:25:00-04:00"
  checked: "emitProgram control flow, commit 00742ee, and focused witness reproduction"
  found: "emitProgram resolves graph/entry, unfolds invocation paths, then schema2ExecutionDocumentSize calls linearInput(entry), and only afterward checks function.Match. Commit 00742ee introduced the size call at that location. The focused witness deterministically fails with unsupported linear C type \"Switch\"."
  implication: "The code-ordering mechanism is directly observed. Structural validation must precede schema-2 preflight while graph/entry resolution retains first precedence."
- timestamp: "2026-09-19T00:35:00-04:00"
  checked: "Agent-authored cgen regression TestUnsupportedProgramShapePrecedesSchema2Preflight before production changes"
  found: "The test fails red with unsupported linear C type \"Switch\", matching the original session witness. Oracle is the specified established refusal contract."
  implication: "The bug is reproducible at the emitter boundary and the regression test directly judges diagnostic precedence."
- timestamp: "2026-09-19T00:50:00-04:00"
  checked: "Focused fix verification"
  found: "The new cgen regression and original session witness pass. Invocation ordering, 4096/4097 capacity, preflight mutation seam, exact output N-1/N, and deep-diamond O0/O3 execution tests all pass unchanged."
  implication: "The fix restores refusal precedence while retaining the native occurrence-capacity and output safeguards."
- timestamp: "2026-09-19T01:00:00-04:00"
  checked: "Scoped diff and revert-and-reconfirm guardrail"
  found: "The diff only reorders existing validation/preflight blocks and adds one focused test; it does not delete or short-circuit behavior. With the regression retained, reverting the production reorder restores the Switch-type failure; reapplying it restores pass."
  implication: "The targeted reorder is load-bearing and the test kills removal of the fix. Stryker is not configured/applicable for this Go repository, so this direct revert mutant supplies the available mutation evidence."
- timestamp: "2026-09-19T01:10:00-04:00"
  checked: "Exact build, vet, uncached full-suite gate"
  found: "Build and vet passed. All code packages including cgen/native passed, but session failed TestVerificationGroundednessFrontierIsPinned because untracked Phase 15 15-VERIFICATION.md line 107 contains an unparseable `go test` evidence command."
  implication: "The emitter fix has no observed code regression; automated acceptance remains blocked on a Phase 15 planning-artifact groundedness violation that must be classified before commit."
- timestamp: "2026-09-19T01:15:00-04:00"
  checked: "Phase 15 verification command and focused groundedness law"
  found: "Line 107 used prose containing backticked `go test`, which the evidence extractor correctly classified as an unparseable command. Replacing it with the exact grounded Phase 15 integration command makes TestVerificationGroundednessFrontierIsPinned pass."
  implication: "The first full-suite failure was an independent mechanical defect in the untracked verification artifact; it is repaired separately from the code fix."
- timestamp: "2026-09-19T01:30:00-04:00"
  checked: "Final automated acceptance gate"
  found: "go build ./..., go vet ./..., and go test ./... -count=1 all exited zero. The session package passed in 253.106s, including the original LTO witness and verification-groundedness law."
  implication: "Under the user's explicit no-human-UAT policy, automated acceptance confirms the fix end to end and the session may be resolved."
- timestamp: "2026-09-19T01:35:00-04:00"
  checked: "Semantic debug recall/capture availability"
  found: "Project configuration has mempalace.enabled=false and no MemPalace tool is available; the durable knowledge-base entry was written successfully."
  implication: "Semantic indexing is skipped by protocol; .planning/debug/knowledge-base.md remains the durable fallback."

## Eliminated

- hypothesis: "The configured execution output limit causes the wrong refusal."
  evidence: "The observed error is returned inside schema2ExecutionDocumentSize by linearInput before executionBytes is compared with executionOutputLimit."
  timestamp: "2026-09-19T00:35:00-04:00"
- hypothesis: "Host or toolchain environment changes diagnostic ordering."
  evidence: "Both focused Go tests fail identically before invoking Clang, native execution, or any environment-dependent operation."
  timestamp: "2026-09-19T00:35:00-04:00"

## Resolution

root_cause: "emitProgram performs schema-2 occurrence/output preflight before supported-body validation, allowing schema2ExecutionDocumentSize's linearInput call on a Match entry to mask the established multi-function branch-body refusal."
fix: "Move emitProgram's ordered supported-shape validation ahead of the unchanged invocation-path, event-capacity, and exact output-size preflights; add a direct cgen diagnostic-precedence regression."
oracle_type: "specified (established named multi-function branch-body refusal)"
verification:
  target_test: {result: pass}
  mutation_check: {result: skipped, reason_if_skipped: "Stryker is not configured/applicable for Go; scoped revert-and-reconfirm killed removal of the production reorder", mutant_killed: false}
  no_op_deletion: {result: pass, deletion_justified_by_rca: false}
  adjacent_tests: {result: pass, suites_run: ["focused diagnostic-precedence tests", "occurrence/output safeguard tests", "go build ./...", "go vet ./...", "go test ./... -count=1"]}
  revert_and_reconfirm: {result: pass, bug_returned_on_revert: true, fixed_on_reapply: true}
  guardrail_verdict: accepted
files_changed:
  - "internal/compiler/cgen/cgen_program.go"
  - "internal/compiler/cgen/cgen_program_test.go"
  - ".planning/phases/15-event-identity-lang-execution-2/15-VERIFICATION.md"
commit: "b55c1fc"

## Prevention

- code branch: "The output-size estimator assumed straight-line linear bodies but was invoked before emitProgram established that supported-shape invariant."
- data branch: "A valid checked multi-function program with a Match entry exercised the unsupported-shape refusal and exposed the incidental Switch-type error."
- why_not_caught: "The native-capacity regression suite proved bounds for supported straight-line programs but did not pair those preflights with an unsupported-body diagnostic-precedence case."
- recurrence_guard: "internal/compiler/cgen/cgen_program_test.go:TestUnsupportedProgramShapePrecedesSchema2Preflight, plus the existing 4096/4097 occurrence and N-1/N output-bound controls."
