---
phase: 16-branch-match-emitter-port
verified: 2026-09-21T23:21:58Z
status: passed
score: 6/6 must-haves verified
covered_files:
  - .planning/REQUIREMENTS.md
  - .planning/phases/16-branch-match-emitter-port/16-01-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-01-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-02-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-02-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-03-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-03-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-04-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-04-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-05-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-05-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-06-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-06-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-07-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-07-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-08-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-08-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-09-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-09-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-10-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-10-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-11-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-11-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-12-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-12-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-13-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-13-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-14-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-14-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-15-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-15-SUMMARY.md
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_n1_convergence_test.go
  - internal/compiler/cgen/cgen_names_test.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/cgen/cgen_program_test.go
  - internal/compiler/cgen/cgen_restrict_probe_test.go
  - internal/compiler/cgen/cgen_test.go
  - internal/compiler/cgen/export_test.go
  - internal/compiler/cgen/legacy_emitter_evidence_test.go
  - internal/compiler/core/core_test.go
  - internal/compiler/evidence/evidence_test.go
  - internal/compiler/executionpeer/executionpeer.go
  - internal/compiler/native/foreign_retained_test.go
  - internal/compiler/native/native_lto_test.go
  - internal/compiler/native/native_test.go
  - internal/compiler/native/symbols_test.go
  - internal/compiler/session/evidence_grade_test.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_payload_control_test.go
  - internal/compiler/session/session_phase11_differential_test.go
  - internal/compiler/session/session_phase16_control_test.go
  - internal/compiler/session/session_phase16_emitter_inventory_test.go
  - internal/compiler/session/session_phase16_frozen_evidence_external_test.go
  - internal/compiler/session/session_phase5.go
  - internal/compiler/session/session_phase5_alias.go
  - internal/compiler/session/session_phase5_corpus_test.go
  - internal/compiler/session/session_phase5_mismatch.go
  - internal/compiler/session/session_phase5_sanitize.go
  - internal/compiler/session/session_phase6.go
  - internal/compiler/session/session_phase6_injectors_test.go
  - internal/compiler/session/session_phase6_pin_test.go
  - internal/compiler/session/session_phase7.go
  - internal/compiler/session/session_phase7_test.go
  - internal/compiler/session/session_test.go
  - internal/compiler/session/witness_registry_test.go
  - testdata/phase16/file-frozen-evidence.json
  - testdata/phase16/generated-frozen-evidence.json
  - testdata/phase16/legacy-emitter-artifacts.json
  - testdata/phase16/legacy-emitter-evidence.json
  - testdata/phase16/public-emitter-consumers.json
  - testdata/phase16/restrict_readonly_probe.c
  - testdata/phase16/validation-corpus-run-record.jsonl
  - testdata/phase16/validation-corpus-run-record.manifest.json
covered_digest: "v1:sha256:fcc1bdcd702f2091a8a8dc8697a9a2cb02be0c88aa5946a6d690e61232bc29de"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: gaps_found
  previous_score: 4/6
  gaps_closed:
    - "The sole public production dispatcher selects schema-2 emitProgram and has no legacy-emitter fallback."
    - "Every direct public cgen emitter use under internal/compiler is classified by a deterministic registry."
  gaps_remaining: []
  regressions: []
decision_coverage:
  honored: 11
  total: 11
  not_honored: []
---

# Phase 16: Branch/Match Emitter Port Verification Report

**Phase Goal:** One emission law lowers every admissible program, instead of two laws split by a function-count guard.

**Verified:** 2026-09-21T23:21:58Z

**Status:** passed

**Re-verification:** Yes — after gap closure

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
| --- | --- | --- | --- |
| 1 | The three superseded public dispatch emitters are deleted in the atomic dispatch-flip commit. | ✓ VERIFIED | `rg` finds no definitions for `emitMatch`, `emitBranch`, or `emitLinear`. Commit `0607486` both replaces the two public dispatch bodies with `emitProgram` calls and deletes the three bodies. |
| 2 | N=1 public/direct convergence is byte-identical for the admitted fixtures, while cut families retain M004 refusal and each moved golden has provenance. | ✓ VERIFIED | `TestN1ConvergenceDifferential`, `TestLegacyEmitterEvidence`, and both frozen-evidence fault controls pass; `TestPreviousPhaseGoldenCUnchanged` passes for all four preserved golden files. |
| 3 | The three no-consumer families are formally cut to M004 with owner, reopening condition, and `-flto` consequence. | ✓ VERIFIED | The NAT-09 amendment names `emitLinearForeign`, `emitLinearBorrowedByPointer`, and `emitLinearBorrowedByPointerPlain`; `PHASE-16-DEBT.md` records D-16-11 through D-16-13 with all required fields. |
| 4 | Public emission has no function-count fork and derives `live_resources`. | ✓ VERIFIED | `cgen.Emit` and `EmitNative` directly delegate to `emitProgram`; no `len(program.Functions) != 1` condition remains. `deriveProgramLiveResources` feeds JSON serialization, exercised by both resource-tail tests. |
| 5 | The sole production dispatcher has no fallback that converts an M004 refusal into C. | ✓ VERIFIED | `Phase16ControlNativeC` is exactly `return cgen.EmitNative(program)`. The admitted/refusal equality controls and non-test-source frozen-route scan pass, including all Phase 16 M004 corpus refusal cases. |
| 6 | Every direct public emitter consumer is classified by the deterministic registry. | ✓ VERIFIED | `TestPhase16PublicEmitterConsumerInventory` passes against the current AST scan and registry. The registry has 77 sorted entries (63 dynamic, 14 refusal witnesses); duplicate, stale, missing, invalid-classification, and missing-witness mutations all fail the shared validator. |

**Score:** 6/6 truths verified (0 present, behavior-unverified)

### Required Artifacts

| Artifact | Expected | Status | Details |
| --- | --- | --- | --- |
| `internal/compiler/cgen/cgen.go` | Sole public dispatch to `emitProgram` | ✓ VERIFIED | Both public APIs validate then call `emitProgram` once. |
| `internal/compiler/cgen/cgen_program.go` | Program, branch, and match lowering with derived resource tail | ✓ VERIFIED | Substantive production implementation; tracer, match-payload, branch, ordering, and resource tests pass. |
| `internal/compiler/session/session_phase5.go` | Production refusal boundary | ✓ VERIFIED | The helper has no artifact lookup and propagates the public emitter result unchanged. |
| `internal/compiler/session/session_phase16_frozen_evidence_external_test.go` | Test-only, program-bound historical evidence loader | ✓ VERIFIED | `_test.go`-only loader checks checked fixture/program canonical bytes, current refusal, and artifact digests before returning historical bytes. |
| `internal/compiler/session/session_phase16_emitter_inventory_test.go` | AST-derived source/registry bijection gate | ✓ VERIFIED | Current-source inventory and five negative controls pass. |
| `testdata/phase16/public-emitter-consumers.json` | Sorted consumer registry | ✓ VERIFIED | Parsed by the passing bijection gate; no stale rows remain. |
| `PHASE-16-DEBT.md` plus NAT-09 amendment | Formal M004 cut | ✓ VERIFIED | All three named families have M004 ownership, prerequisite/reopening conditions, witnesses, and honest `-flto` consequences. |

### Key Link Verification

| From | To | Via | Status | Details |
| --- | --- | --- | --- | --- |
| `cgen.go` | `cgen_program.go` | `Emit` and `EmitNative` | ✓ WIRED | Each API delegates to `emitProgram` after validation. |
| `session_phase5.go` | `cgen.go` | `Phase16ControlNativeC` | ✓ WIRED | One direct delegation; terminal error propagates. |
| test-only frozen-evidence loader | `file-frozen-evidence.json` | canonical program/refusal/artifact checks | ✓ WIRED | Loader is external test-package code and no non-test session source references phase-16 frozen selectors. |
| inventory test | consumer registry | AST identities checked both directions | ✓ WIRED | The live-source bijection and all negative controls pass. |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| --- | --- | --- | --- | --- |
| `cgen_program.go` | `liveResources` | `deriveProgramLiveResources(program)` → JSON writer | Yes; seed control changes generated C serialization | ✓ FLOWING |
| `session_phase5.go` | returned C/error | live `cgen.EmitNative(program)` | Yes; no historical artifact branch exists in production | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| --- | --- | --- | --- |
| Session production boundary and registry bijection | focused `go test ./internal/compiler/session` Phase-16 controls | 0; admitted/refusal, non-bypass, inventory, and all mutations passed | ✓ PASS |
| Public/direct convergence and derived lowering | focused `go test ./internal/compiler/cgen` Phase-16 controls | 0; N=1, branch/match, resource, and frozen-evidence controls passed | ✓ PASS |
| Previous golden preservation | `go test ./internal/compiler/core -run '^TestPreviousPhaseGoldenCUnchanged$' -count=1 -v` | 0; four golden files passed | ✓ PASS |
| Workspace regression | `go test ./...` | 0 | ✓ PASS |

### Probe Execution

Step 7c: SKIPPED — no Phase 16 probe script is declared or present under `scripts/*/tests/probe-*.sh`.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| --- | --- | --- | --- | --- |
| NAT-08 | 16-01 through 16-15 | One emission law lowers every admissible program; superseded emitters are atomically deleted. | ✓ SATISFIED | Atomic cutover/deletion is in `0607486`; public APIs, production session boundary, convergence controls, and current consumer registry all use the one program-emitter authority. |
| NAT-09 | 16-05, 16-06, 16-11 through 16-15 | Three no-consumer emitter families are formally cut rather than deferred a third time. | ✓ SATISFIED | Amendment and three debt rows are complete; public and session controls refuse the cut corpus rather than fabricating live C. |

No orphaned Phase 16 requirements were found.

### Test Quality Audit

| Test File | Linked Req | Active | Skipped | Circular | Assertion Level | Verdict |
| --- | --- | --- | --- | --- | --- | --- |
| `cgen_n1_convergence_test.go` | NAT-08 | yes | no | no | Value | Compares public/direct output and named refusal outcomes. |
| `session_phase16_control_test.go` | NAT-08/NAT-09 | yes | no | no | Behavioral | Verifies admitted equality, terminal refusal, and absence of production artifact selectors. |
| `session_phase16_emitter_inventory_test.go` | NAT-08 | yes | no | no | Behavioral/bijection | Exercises the live AST/registry law plus five independent invalid mutations. |
| `session_phase16_frozen_evidence_external_test.go` | NAT-09 | yes | no | no | Value | Warning: `TestPhase16FileFrozenEvidenceRejectsProgramSubstitution` compares distinct canonical programs but does not invoke the private loader with the substitute. The loader contains the rejecting comparison; add that direct negative call to strengthen the test. |

No disabled requirement test or circular expected-value generator was found. The final row is a non-blocking test-quality warning, not evidence of a production boundary failure.

### Decision Coverage

All 11/11 trackable Phase 16 decisions are honored by shipped artifacts. This is non-blocking.

### Anti-Patterns Found

No blocking anti-pattern was found in the Phase 16 implementation. No unreferenced `TBD`, `FIXME`, or `XXX` debt marker was found in the inspected implementation or control files.

## Human Verification

N/A — infrastructure/foundation phase with no user-facing elements. All phase-goal acceptance criteria were verified programmatically.

## Gaps Summary

The two previous blockers are closed. The session boundary no longer returns historical C after a public M004 refusal, and the source-derived public-emitter registry is now an exact, passing bijection. No remaining gap blocks NAT-08 or NAT-09.

_Verified: 2026-09-21T23:21:58Z_

_Verifier: the agent (gsd-verifier)_
