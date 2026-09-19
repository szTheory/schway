---
status: resolved
trigger: "Fix Phase 15 full-suite regressions after completing plans 15-05 through 15-07 task 1."
created: "2026-09-19"
updated: "2026-09-19T15:46:00-04:00"
---

# Debug Session: Phase 15 Full-Suite Regressions

## Symptoms

- Expected: `go test ./... -count=1` passes so Plan 15-07 Task 2 may close debt and validation evidence.
- Actual: the full suite fails after Phase 15 native `/2` evidence, peer-gate, and diamond-gate changes.
- Errors observed: native process static-policy violation in `executionpeer_test.go`; cache/measure/cgen timeouts; archived evidence/hash and validation-record failures; native differential lane failures.
- Timeline: began during Phase 15 execution after 15-05 through 15-07 Task 1 commits.
- Reproduction: run `go test ./... -count=1` from the repository root.

## Current Focus

- hypothesis: "Confirmed fixed: all deterministic Phase 15 full-suite regressions are repaired and all closure artifacts are mechanically current."
- test: "Use the user's explicit automated-confirmation policy: the recorded build, vet, full-suite, integration/E2E/smoke, and seam gates at exit 0 constitute end-to-end confirmation."
- expecting: "Satisfied by the accepted full gate and focused guardrail evidence."
- next_action: "Archive the resolved session, append its prevention entry to the debug knowledge base, and commit documentation according to commit_docs=true."
- bug_class: "bohrbug (multiple deterministic regression clusters; environment-sensitive timeout reports did not reproduce)"
- known_pattern_candidate: "none — no project debug knowledge base exists"
- sbfl: "skipped — the suite has failures and passes but no per-test coverage pipeline; direct differential output localizes the mismatch to schema selection"
- reasoning_checkpoint:
    hypothesis: "Four independent Phase 15 changes caused the aggregate failure: function-count-only interpreter schema selection erased the legacy bare-match /0 versus frame /1 split; executionpeer's import test used forbidden unbounded subprocess helpers; the intentional multi-function /2 migration lacked a successor characterization ledger; and Phase 15 evidence documents violated the Phase 14 groundedness/grade shape."
    confirming_evidence:
      - "Focused differential output showed interpreter lang.execution/1 versus native lang.execution/0; pre-562a281 history and targeted tests prove bare match /0 and frame /1 were distinct contracts."
      - "The native static-policy test named exec.Command and Output in executionpeer_test.go directly."
      - "After restoring legacy selection, every old digest matched except exactly 18 multi-function fixtures required by Phase 15 to move to /2."
      - "Groundedness and grade tests directly named brace package syntax, the stale DiamondFrontierIsPinned identifier, and the absent Grade column."
    falsification_test: "Any focused original reproducer remaining red after its corresponding narrow repair, or the bug not returning when the four code files are reverted, would falsify this combined diagnosis."
    fix_rationale: "Each repair restores or records the violated contract at its source: preserve caller-selected legacy schema, bound the subprocess, retain the old digest ledger plus explicit /2 successors, and make evidence commands/grades mechanically resolvable."
    blind_spots: "Only macOS local execution was available; cross-host stability is inferred from deterministic Go tests and bounded process behavior, not observed on Linux CI."
    candidate_causes:
      - "code: schema selector and unbounded subprocess helper introduced deterministic regressions"
      - "data/docs: characterization and planning evidence lagged the authorized /2 transition"
      - "environment: earlier timeout reports were host-load artifacts and did not reproduce in fresh focused or full runs"
    and_gate: "yes for the aggregate suite — four independent deterministic clusters jointly made it red; each cluster has its own sufficient cause and narrow repair, while timeout reports are a separate environmental observation"

## Evidence

- timestamp: "2026-09-19T00:00:00-04:00"
  checked: "Phase 0 knowledge-base recall and project skill discovery"
  found: "No .planning/debug/knowledge-base.md exists and AGENTS.md reports no project-local skills; the worktree contains unrelated modified planning state plus the untracked active debug directory."
  implication: "There is no prior known-pattern shortcut; preserve planning-state changes and restrict edits to confirmed regression files and this debug session."
- timestamp: "2026-09-19T14:39:55-04:00"
  checked: "Fresh go test ./... -count=1 reproduction"
  found: "Native static-policy deterministically rejects exec.Command and Output in executionpeer_test.go. Session failures include legacy interpreter lang.execution/1 versus native lang.execution/0 mismatches, moved payload hashes, validation-grade/document findings, and Phase 15 verification-groundedness findings. Cache, cgen, measure, executionpeer, and other packages passed; the suite completed in about 4m10s rather than timing out."
  implication: "The current failure set has at least three independent deterministic clusters; previously reported package timeouts are not current regressions on this run. The schema mismatch is upstream of native-differential, CLI, and payload-hash cascades."
- timestamp: "2026-09-19T14:50:00-04:00"
  checked: "Focused legacy schema/hash/native tests and producer history"
  found: "The focused tests reproduce deterministically. Before 562a281, bare match returned Schema0 while frame-stack execution returned Schema1. The new function-count-only selector returns Schema1 for both. Native intentionally retains the legacy single-function writers, producing the observed /1-versus-/0 mismatch."
  implication: "Root cause is confirmed for the legacy differential cascade; preserve the historical producer-specific fallback and reserve Schema2 for multi-function programs."
- timestamp: "2026-09-19T14:55:00-04:00"
  checked: "Focused verification after legacy schema fix"
  found: "Interpreter /0-/1-/2 boundary tests pass; TestFeatureSpecificCoreExecutionSchemas, TestNativeToggleO0O3, both Phase 6 native-differential tests, and the 59-second Phase 5 corpus differential all pass."
  implication: "The legacy schema root cause and fix are verified for adjacent execution paths. Remaining failures are independent process-policy, intentional /2 baseline migration, and planning evidence closure."
- timestamp: "2026-09-19T15:00:00-04:00"
  checked: "Focused characterization replay after restoring legacy schema selection"
  found: "All single-function fixtures now match their original digests. Exactly 18 multi-function fixtures still differ, and Phase 15 explicitly requires those producer paths to emit Schema2 with new invocation and call-edge evidence."
  implication: "Do not overwrite or weaken the Phase 12 historical ledger; add a separate, schema-checked Phase 15 successor ledger for the authorized multi-function transition."
- timestamp: "2026-09-19T15:05:00-04:00"
  checked: "Phase 15 successor characterization ledger and mutation control"
  found: "The replay passes with 18 explicitly pinned Schema2 successors while retaining the original Phase 12 map; the D-12 scalar evidence-invisibility mutation control also passes."
  implication: "The intentional wire migration is now explicit and non-inert. Remaining deterministic failures are confined to Phase 15 planning-evidence closure."
- timestamp: "2026-09-19T15:10:00-04:00"
  checked: "Focused groundedness, reconciliation, and corpus-wide validation-grade laws"
  found: "Groundedness frontier passes with R1=0/R2=0/R3=0; reconciliation is 66 live findings to 66 entries; all validation documents including Phase 15 pass the 105.6-second evidence run and grade derivation."
  implication: "Planning evidence repair is verified. Proceed to the exact full gate before changing completion/debt status."
- timestamp: "2026-09-19T15:30:00-04:00"
  checked: "Exact Phase 15 Task 2 closure gate: go build ./... && go vet ./internal/compiler/session/... && go test ./... -count=1"
  found: "All three commands exited zero. The uncached full suite was green, including session at 251.847s and testsupport at 36.613s; the previously reported cache/cgen/measure/executionpeer timeouts did not recur."
  implication: "The deterministic regressions are repaired and the timeout reports are stale/environmental rather than current code failures. Debt and validation status may now be closed against observed evidence."
- timestamp: "2026-09-19T15:36:00-04:00"
  checked: "Focused debt-register, groundedness, reconciliation, validation-grade, and derived-view checks after closing D-11-51/D-12-21"
  found: "Debt shape, identifier uniqueness, R1/R2/R3 zero frontier, 66-entry reconciliation, and grade laws pass. TestUnreachableClaimsViewIsCurrent reports the expected derived delta only: entries increase 19 to 21 by adding D-11-51 and D-12-21 with the live collision-control witness."
  implication: "The authored debt closures are sound; synchronize the deterministic derived view rather than weakening or removing their witnesses."
- timestamp: "2026-09-19T15:45:00-04:00"
  checked: "Final Phase 15 planning closure and fix-acceptance guardrail"
  found: "Diamond four-tier/frontier/collision tests, payload characterization plus mutation control, debt/register/current-view checks, groundedness, 66/66 reconciliation, and the complete 01-15 validation grade audit all pass. Revert-and-reconfirm restored the original schema/process/payload failures without the four-file code fix and cleared them after reapplication. Code commit is 15ee061; planning closure commit is 3b817d7."
  implication: "All applicable automated signals accept the fix. The session may advance to human verification but must not be archived before confirmation."
- timestamp: "2026-09-19T15:46:00-04:00"
  checked: "Human-verification checkpoint response"
  found: "The user explicitly authorized automated confirmation for this checkpoint when the specified build, vet, full-suite, integration/E2E/smoke, and seam tests pass, and directed that the recorded full gate exit 0 be treated as confirmed fixed."
  implication: "The end-to-end verification requirement is satisfied by explicit user policy; the session is resolved and may be archived without an additional manual UAT run."

## Eliminated

## Resolution

root_cause:
  - "Code regression: Phase 15 interpreter schema selection erased the legacy bare-match /0 versus frame-based /1 distinction."
  - "Code policy regression: the new executionpeer import-boundary test used an unbounded subprocess constructor and output helper."
  - "Data/code migration gap: intentional multi-function /1-to-/2 documents changed canonical hashes without a versioned successor characterization ledger."
  - "Planning evidence gap: Phase 15 used ungrounded command shapes, a stale test name, and omitted the mandatory Grade column."
fix:
  - "Preserve caller-selected legacy schemas and override only multi-function programs to /2, with /0-/1-/2 boundary coverage."
  - "Bound executionpeer's go-list subprocess by time and independent stdout/stderr byte ceilings."
  - "Retain the Phase 12 baseline and add a schema-checked Phase 15 successor digest ledger for 18 multi-function fixtures."
  - "Ground Phase 15 commands/test names and add evidence-derived grades."
  - "Close D-11-51 and D-12-21 with their original histories preserved, mark Phase 15 validation complete, and synchronize the mechanically derived claims view."
verification:
  target_test:
    result: pass
  mutation_check:
    result: skipped
    reason_if_skipped: "Stryker is not configured or applicable for this Go repository; repository-native collision and payload characterization mutation controls both pass."
    mutant_killed: false
  no_op_deletion:
    result: pass
    deletion_justified_by_rca: false
  adjacent_tests:
    result: pass
    suites_run:
      - "focused interpreter/native/differential/process-policy tests"
      - "payload characterization replay and mutation control"
      - "debt, groundedness, reconciliation, derived-view, and archived validation-grade laws"
      - "go build ./... && go vet ./internal/compiler/session/... && go test ./... -count=1"
  revert_and_reconfirm:
    result: pass
    bug_returned_on_revert: true
    fixed_on_reapply: true
  guardrail_verdict: accepted
oracle_type: "specified (frozen wire schemas, process policy, Phase 15 /2 contract, and Phase 14 evidence contract)"
files_changed:
  - internal/compiler/interp/interp.go
  - internal/compiler/interp/interp_test.go
  - internal/compiler/executionpeer/executionpeer_test.go
  - internal/compiler/session/session_payload_replay_test.go
  - .planning/phases/15-event-identity-lang-execution-2/15-RESEARCH.md
  - .planning/phases/15-event-identity-lang-execution-2/15-VALIDATION.md
  - .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/PHASE-11-DEBT.md
  - .planning/milestones/M002-phases/12-result-payloads/PHASE-12-DEBT.md
  - .planning/UNREACHABLE-CLAIMS.md

## Prevention

### Blameless branching 5-Whys

- **Code branch:** The schema selector collapsed producer-specific legacy behavior into a function-count rule because the new `/2` rule was expressed as a replacement instead of a narrow multi-function override. That was possible because the existing tests covered `/2` behavior but did not place bare-match `/0`, frame-based `/1`, and multi-function `/2` beside one another as a single boundary contract.
- **Process-policy branch:** The import-boundary test used convenient subprocess helpers because its implementation was reviewed for dependency independence, while the repository-wide bounded-process policy was only exercised later by the full suite. The focused package test therefore could be green before the static policy gate ran.
- **Data/evidence branch:** The authorized `/1`-to-`/2` wire migration moved canonical digests and validation evidence because the historical ledger had no versioned successor representation and the Phase 15 document edits were not checked against the Phase 14 groundedness/grade contract until closure.
- **Environment branch:** Earlier timeout reports accompanied the aggregate failure, but fresh focused and full runs did not reproduce them. They were host-load observations rather than contributing code causes and are intentionally not treated as part of the root-cause set.
- **AND-gate:** The red full suite required multiple independent regression clusters. No single repair could make the aggregate gate green; each code/data branch needed its own targeted guard.

### Why this was not caught earlier

The Phase 15 focused implementation gates did not compose the legacy schema boundaries, repository-wide subprocess policy, versioned payload migration, and planning-evidence laws in one pre-closure run. The exact build + vet + uncached full-suite gate caught the aggregate before Phase 15 completion.

### Recurrence guards

- `internal/compiler/interp/interp_test.go` now covers the `/0` bare-match, `/1` single-frame, and `/2` multi-function boundary in one test surface.
- `internal/compiler/session/session_test.go:TestSourceNeverSpawnsUnboundedProcesses` enforces bounded subprocess construction repository-wide; `internal/compiler/executionpeer/executionpeer_test.go` now uses independent capped stdout/stderr writers and a timeout.
- `internal/compiler/session/session_payload_replay_test.go:TestPayloadCorpusCharacterizationReplay` retains the Phase 12 ledger and checks the explicit 18-entry Phase 15 `/2` successor ledger; `TestPayloadCorpusCharacterizationReplayMutationKilled` proves the characterization is load-bearing.
- The groundedness, reconciliation, validation-grade, debt-register, and `TestUnreachableClaimsViewIsCurrent` laws mechanically prevent stale Phase 15 evidence and derived views from closing green.
