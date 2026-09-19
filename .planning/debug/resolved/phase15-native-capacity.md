---
status: resolved
trigger: "Fix two Phase 15 code-review blockers in native /2 occurrence capacity and output bounds."
created: "2026-09-19"
updated: "2026-09-19T01:40:00-04:00"
---

# Debug Session: Phase 15 Native Capacity

## Symptoms

- Expected: native execution accepts and runs the preflight-admitted shared-leaf diamond and has an explicit, honest bound for `/2` output.
- Actual: `go run ./cmd/lang run --engine=native testdata/phase07/deep_diamond_acyclic.lang --json` returns `operational_failure` / `native.run_signaled`.
- Code review identifies: event capacity counts declared bodies rather than unfolded occurrences; 64 KiB generator/runner output limits silently contradict the advertised 4,096-node admission bound.
- Reproduction: run the command above; inspect `15-REVIEW.md` CR-01 and CR-02.

## Current Focus

- bug_class: "bohrbug"
- reasoning_checkpoint:
    hypothesis: "CR-02: after occurrence-weighted capacity removes the abort, the same admitted deep diamond exits 74 because its schema-2 JSON exceeds the generated 65,536-byte writer ceiling; the runner independently enforces the same generic stream ceiling."
    confirming_evidence:
      - "With only occurrence-weighted event capacity applied, both O0 and O3 change from SIGABRT to deterministic exit status 74, the generated writer's overflow return path."
      - "The interpreter CLI envelope for the identical schema-2 execution is 84,244 bytes, already beyond 65,536, and both cgen and native hard-code that smaller ceiling."
    falsification_test: "If the generated execution's canonical bytes fit within 65,536, or exit 74 originates outside lang_write_bytes, the hypothesis is false."
    fix_rationale: "A shared schema-2 document contract plus pre-serialization size calculation makes the byte bound explicit, accepts documents within it, refuses larger programs with a distinct cgen diagnostic, and gives the runner exactly the same stdout bound."
    blind_spots: "The estimator must match the C writer's field/suffix rules and the largest outcome spelling; differential and boundary tests must kill estimator drift."
    candidate_causes:
      - "code: schema-2 generation has no preflight document-size check"
      - "config: schema-2 incorrectly inherits the legacy 65,536-byte generic stream limit"
      - "data: complete invocation ancestry makes document size grow with unfolded occurrences"
    and_gate: "yes — oversized ancestry-bearing data and the unversioned legacy limit jointly cause exit 74; both require an explicit schema-2 size contract."
- test: "Completed."
- expecting: "Satisfied."
- next_action: "Archive session and append the durable knowledge-base recurrence guard."

## Evidence

- timestamp: "2026-09-19T00:10:00-04:00"
  checked: ".planning/phases/15-event-identity-lang-execution-2/15-REVIEW.md"
  found: "Two required blockers are documented: dynamic event capacity is derived from declared bodies, and generator/runner output is independently capped at 65,536 bytes despite a 4,096-occurrence preflight limit."
  implication: "Both capacity derivation and output-bound semantics require direct tests; compile-only coverage cannot verify the execution contract."
- timestamp: "2026-09-19T00:20:00-04:00"
  checked: "Knowledge base and common bug patterns"
  found: "No prior resolution matches native dynamic-capacity/output sizing. The closest general categories are boundary/capacity and config limits."
  implication: "Treat review claims as hypotheses and reproduce directly; no known-pattern shortcut applies."
- timestamp: "2026-09-19T00:20:00-04:00"
  checked: "SBFL prerequisites"
  found: "The reported failing behavior has no existing failing automated test and the Go suite does not provide per-test coverage spectra in this workflow."
  implication: "SBFL skipped: no failing/passing per-test coverage spectrum is available."
- timestamp: "2026-09-19T00:25:00-04:00"
  checked: "Exact CLI reproduction plus focused existing cgen tests"
  found: "The deep-diamond native command deterministically returned operational_failure/native.run_signaled, while TestInvocationPathTableDeepDiamondMeasures61 and TestInvocationPathTableBoundary passed."
  implication: "Path preflight and C serialization succeed; failure occurs during generated native execution, consistent with event-buffer exhaustion rather than graph/preflight failure."
- timestamp: "2026-09-19T00:40:00-04:00"
  checked: "Agent-authored TestDeepDiamondExecutesAcrossNativeOptimizationTiers before production changes"
  found: "Both O0 and O3 failed red with native.run_signaled/SIGABRT."
  implication: "The automated differential reproduces CR-01 with a derived semantic oracle."
- timestamp: "2026-09-19T00:45:00-04:00"
  checked: "Single-variable experiment: occurrence-weighted capacity only"
  found: "Both tiers stopped aborting and instead exited 74; the interpreter CLI JSON envelope measures 84,244 bytes against the 65,536-byte writer/runner cap."
  implication: "CR-01's mechanism is confirmed and corrected by occurrence weighting; CR-02 is an independent next failure, not an alternative explanation."
- timestamp: "2026-09-19T01:05:00-04:00"
  checked: "Focused CR-01/CR-02 regression and native boundary suites"
  found: "Deep diamond passes derived interpreter equality under O0/O3; exact schema-2 N-1 refuses before serialization, N succeeds; native execution-document truncation remains distinct from terminal-record absence."
  implication: "Both fixes address their root mechanisms and preserve the runner's independent compile/stderr bounds."
- timestamp: "2026-09-19T01:25:00-04:00"
  checked: "Exact CLI, focused/adjacent packages, vet, and uncached full repository suite"
  found: "CLI exits 0 with 84,244 bytes; cgen/native/execution/session packages pass; go vet ./... passes; go test ./... -count=1 passes including the 248-second session suite."
  implication: "Original behavior and adjacent/full-repository automated gates are green; revert-and-reconfirm remains before acceptance."
- timestamp: "2026-09-19T01:30:00-04:00"
  checked: "First revert-and-reconfirm attempt"
  found: "The scoped stash also removed the new regression test, so go test reported no tests to run and could not serve as a valid revert oracle; changes restored cleanly."
  implication: "This attempt is invalid rather than passing; use the immutable original CLI reproduction for the revert signal."
- timestamp: "2026-09-19T01:35:00-04:00"
  checked: "Revert-and-reconfirm with the original CLI reproduction"
  found: "Removing the seven-file patch restored nonzero native.run_signaled; restoring the patch returned exit 0/status pass."
  implication: "The committed change, not an environmental change, fixes the original defect."

## Prevention

- code branch: Unique-body sizing was reusable Phase 11 logic, but Phase 15 changed the semantic unit to occurrences without changing the capacity unit.
- config/data branch: A legacy 64 KiB limit remained implicit while complete invocation ancestry increased evidence volume beyond it.
- why_not_caught: The existing 61-node test stopped at EmitNative and the 4,096-node test asserted admission only; neither executed native code or checked document bytes.
- recurrence_guard: TestDeepDiamondExecutesAcrossNativeOptimizationTiers checks exact occurrence capacity, O0/O3 interpreter equality, and estimator-vs-runtime bytes; TestSchema2ExecutionOutputBoundIsPreflighted pins N-1/N refusal before serialization.

## Eliminated

## Resolution

root_cause:
  - "CR-01: schema-2 event capacity counts unique declared bodies instead of occurrence-weighted executed operations."
  - "CR-02: schema-2 output inherits a 65,536-byte legacy writer/runner ceiling without a schema-specific preflight size contract."
fix: "Derive event capacity from every preflight occurrence's operation count; introduce a shared 16 MiB schema-2 document contract, exact pre-serialization size estimator and cgen.execution_output_exceeded refusal, while retaining 64 KiB for non-document process streams."
oracle_type: "derived (interpreter/native differential)"
verification:
  target_test: {result: pass}
  mutation_check: {result: skipped, reason_if_skipped: "Stryker is not applicable/configured for Go; revert-and-reconfirm killed the whole patch instead", mutant_killed: false}
  no_op_deletion: {result: pass, deletion_justified_by_rca: false}
  adjacent_tests: {result: pass, suites_run: ["focused cgen/native", "cgen/native/execution/session", "go vet ./...", "go test ./... -count=1", "exact CLI smoke"]}
  revert_and_reconfirm: {result: pass, bug_returned_on_revert: true, fixed_on_reapply: true}
  guardrail_verdict: accepted
commit: "00742ee"
files_changed:
  - "internal/compiler/cgen/cgen_program.go"
  - "internal/compiler/cgen/cgen.go"
  - "internal/compiler/cgen/cgen_program_test.go"
  - "internal/compiler/cgen/export_test.go"
  - "internal/compiler/execution/execution.go"
  - "internal/compiler/native/native.go"
  - "internal/compiler/native/native_test.go"
