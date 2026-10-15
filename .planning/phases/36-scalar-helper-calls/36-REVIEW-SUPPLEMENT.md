---
phase: 36-scalar-helper-calls
reviewed: 2026-10-15
depth: standard
status: clean
files_reviewed: 7
findings:
  critical: 0
  warning: 0
  info: 0
  total: 0
---

# Phase 36 proof-bridge review supplement

The orchestrator reviewed the Task 2 supplement in `969fb22b`: the phase-local
Node test adapter, three JSON subjects, Plan 08, its summary, and
`36-PROHIBITION-EVIDENCE.md`. This supplements the independent clean 52-file
compiler review; it does not relabel that review as covering these later files.

The adapter uses only Node built-ins and the existing Go tests. Subject paths
and JSON shapes form a closed allowlist; fixtures supply no command, package,
source path, or replacement text. Subprocesses use argument arrays without a
shell, explicit cache, timeout and output bounds. Temporary overlay files live
in a private temporary directory and are removed in `finally`; tracked source
bytes are checked before/after each run. Compiler/runtime code and dependencies
remain unchanged. These source checks also cover the narrow L1 trust boundary
introduced by the adapter, without claiming a new production security audit.

The negative classifier requires one actual named Go root, its exact expected
assertions and only the declared failing children. Infrastructure errors,
skips, missing roots, malformed subjects, anchor drift and unrelated failures
skip both Node tests, so the canonical producer cannot accept either a
non-vacuous RED or GREEN. The clean fixture exercises the same subject-env
path, distinguishing content-caused failure from merely setting an env var.
The default run restores original source behavior; overlays do not mutate the
checkout.

After reviewing the source, the orchestrator reran both exact canonical
`check prohibition-enforcement` requests recorded in the evidence artifact.
Both exited 0 and returned `status: green`, `flagged: false`, `failFirst: true`,
`passed: true`, `failFirstProof: violation-fixture`. Node syntax and
`git diff --check` also passed. No concrete defect was found.

The causal mutant proves the numeric-limit assertion has teeth; it does not
prove source intent or arithmetic. Those remain separate checker and engine
claims supported by the current 45-root capture. The two producer verdicts
apply only to these named checks on the recorded local toolchain.
