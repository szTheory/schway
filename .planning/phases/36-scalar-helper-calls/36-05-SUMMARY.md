---
phase: 36-scalar-helper-calls
plan: "05"
subsystem: compiler/native
tags: [c17, scalar-calls, invocation-paths, evidence-capacity, clang]
requires:
  - phase: "36-03"
    provides: "Checked scalar-call inputs and compiler facts consumed by native emission."
provides:
  - "Native scalar helper calls with sequenced typed arguments/results and stable dynamic invocation paths."
  - "Bounded optional call capture with explicit capacity reporting and preserved application outcomes."
  - "Active emitter inventory and negative controls for call values, path lifetime, and capacity guards."
affects: ["36-06", "36-07", "36-08"]
actuals:
  tokens: 19357
  tasks: 3
  commits: 4
commits: 4
plan_head_before: be3ba76342f2b35bc8aebbc69e1182ea6e4e9311
plan_head_after: 3794dfaabb19d177eb94d689103e0d26ac013c1c
tech-stack:
  added: []
  patterns:
    - "Generated C uses bounded append-only storage for retained invocation paths."
    - "Optional evidence exhaustion remains separate from the native process outcome."
key-files:
  created:
    - internal/compiler/cgen/cgen_phase36_test.go
    - internal/compiler/native/native_phase36_test.go
    - .planning/phases/36-scalar-helper-calls/36-05-task1-red-evidence.json
    - .planning/phases/36-scalar-helper-calls/36-05-task2-red-evidence.json
    - .planning/phases/36-scalar-helper-calls/36-05-task3-red-evidence.json
    - .planning/phases/36-scalar-helper-calls/deferred-items.md
  modified:
    - internal/compiler/cgen/cgen.go
    - internal/compiler/cgen/cgen_program.go
    - .planning/WINDOWS.md
key-decisions:
  - "Admit checked scalar call-only wrappers while keeping legacy, ownership, and foreign-call refusals in place."
  - "Use retained bounded path storage so nested call identities survive callee return."
  - "Keep optional capture exhaustion separate from actual output, exit status, and later defect reporting."
patterns-established:
  - "Native invocation ancestry is retained in a bounded arena rather than a reused stack buffer."
  - "Call evidence capacity is checked before storage or ordinal narrowing; app execution continues with incomplete capture."
requirements-completed: []
coverage:
  - id: D1
    description: "Checked repeated scalar helper calls emit correct values and stable nested invocation paths."
    verification:
      - kind: integration
        ref: "internal/compiler/cgen/cgen_phase36_test.go#TestPhase36NativeDynamicInvocationPaths"
        status: pass
    human_judgment: false
  - id: D2
    description: "Event, path, encoded-output, and ordinal limits fail closed while application outcomes and streams remain intact."
    verification:
      - kind: integration
        ref: "internal/compiler/cgen/cgen_phase36_test.go#TestPhase36CallEventCapacityBoundary"
        status: pass
      - kind: integration
        ref: "internal/compiler/cgen/cgen_phase36_test.go#TestPhase36NativeOrdinalAndPathBoundaries"
        status: pass
      - kind: integration
        ref: "internal/compiler/native/native_phase36_test.go#TestPhase36CaptureOutcomeSeparation"
        status: pass
    human_judgment: false
  - id: D3
    description: "New emitter sites and dropped-value/path/capacity negative controls are enumerated without changing frozen C expectations."
    verification:
      - kind: unit
        ref: "internal/compiler/cgen/cgen_phase36_test.go#TestPhase36EmitterCallInventoryControls"
        status: pass
      - kind: unit
        ref: "internal/compiler/cgen/cgen_phase35_test.go#TestPhase35EmitterAdmissionGates"
        status: pass
    human_judgment: false
duration: 28min
completed: 2026-10-14
status: complete
---

# Phase 36 Plan 05: Native Scalar Helper Calls Summary

**Checked scalar helper calls now run as sequenced C17 calls with stable dynamic ancestry, bounded optional capture, and preserved real application outcomes.**

## Performance

- **Duration:** 28 minutes
- **Started:** 2026-10-15T02:57:04Z
- **Completed:** 2026-10-15T03:25:17Z
- **Tasks:** 3
- **Task files and evidence modified:** 9, including the broken-windows entry
- **Host/toolchain:** Darwin arm64; Go 1.24.0; Apple clang 21.0.0

## Accomplishments

- Emitted typed, sequenced scalar helper calls across all four U64/Bool argument/result pairs. Repeated and nested calls retain complete invocation paths after callees return, including under O3 and O3-LTO.
- Added independent exact and plus-one controls for call-event, path-byte, encoded-report, and maximum ordinal bounds. Conformance capacity failures use the named exit-69 path; optional application capture reports exhaustion while retaining actual streams, exit status, and later U64 overflow defects.
- Kept capture-disabled execution free of call identity/counter work and verified more than 4096 dynamic calls. Extended the emitter inventory and native negative controls while preserving old no-call C expectations.

## Task Commits

1. **Task 1: Compile and run one finite repeated helper with stable canonical paths** — `187bb6e3` (RED), `caf98322` (implementation).
2. **Task 2: Reach exact event/path/ordinal exhaustion and preserve actual application outcomes** — `397a23ed`.
3. **Task 3: Adapt inherited emitter inventories and reach dropped call/value/guard controls** — `3794dfaa`.

The plan metadata commit follows this summary and the canonical state/roadmap updates.

## Files Created/Modified

- `internal/compiler/cgen/cgen.go` — selects eligible checked scalar call programs for native emission.
- `internal/compiler/cgen/cgen_program.go` — emits scalar call statements, bounded dynamic invocation paths, and capacity capture reporting.
- `internal/compiler/cgen/cgen_phase36_test.go` — native path, event/path/output/ordinal boundary, value/path mutation, and emitter inventory controls.
- `internal/compiler/native/native_phase36_test.go` — application outcome and bounded capture controls, including late defect and capture-disabled high-call-count cases.
- `.planning/phases/36-scalar-helper-calls/36-05-task1-red-evidence.json`, `.planning/phases/36-scalar-helper-calls/36-05-task2-red-evidence.json`, and `.planning/phases/36-scalar-helper-calls/36-05-task3-red-evidence.json` — accepted TDD RED receipts.
- `.planning/phases/36-scalar-helper-calls/deferred-items.md` and `.planning/WINDOWS.md` — recorded inherited core-validator refusal shifts found by the frozen corpus guard.

## Decisions Made

- Keep scalar call-only wrappers eligible through a narrow checked-shape admission path; preserve legacy, ownership, and foreign-call refusals.
- Store invocation ancestry in bounded append-only memory so nested paths remain valid after helper return.
- Treat capture capacity as optional evidence state for applications; never let it replace the real process outcome or later defect attribution.

## TDD Evidence

- Task 1 RED: `36-05-task1-red-evidence.json`; implementation and emitted C passed the native O3 and O3-LTO runs.
- Task 2 RED: `36-05-task2-red-evidence.json`; exact/plus-one and real application outcome controls passed.
- Task 3 RED: `36-05-task3-red-evidence.json`; the checked call-only wrapper moved from the named native-emission refusal to a passing emitted/run result.

## Verification

- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen -run '^TestPhase36NativeDynamicInvocationPaths$' -count=1 -v -timeout=600s` — PASS; O3 and O3-LTO.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen ./internal/compiler/native -run '^(TestPhase36CallEventCapacityBoundary|TestPhase36NativeOrdinalAndPathBoundaries|TestPhase36CaptureOutcomeSeparation|TestPhase33CaptureReadBounded)$' -count=1 -v -timeout=600s` — PASS; both package roots ran, exact/plus-one boundaries passed, and the native app retained success and later exit-65 defect outcomes.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen ./internal/compiler/core -run '^(TestPhase36EmitterCallInventoryControls|TestPhase35EmitterAdmissionGates|TestPhase35UnsignedShiftEmissionAndEventProjection|TestInvocationPathTableBoundary|TestInvocationPreflightGuardIsNotInert|TestProgramInvocationIndexThreading|TestNativeFunctionCalledPreorder|TestPhase273NamedCapacityExitGate|TestPhase273CapacityContractLiteralsAgree|TestAllOperationKindsHandledAtEverySite)$' -count=1 -v -timeout=600s` — PASS; both package roots ran.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase36NoCallGeneratedCByteBaseline$' -count=1 -v -timeout=600s` — FAILS only on three inherited core-validator refusal-route changes listed below; generated-C byte snapshots match.

## Deviations from Plan

The inherited core and cgen inventory files needed no edits: their scoped controls passed unchanged. The new call-site census and mutation controls live with the Phase36 emitter tests, preserving the existing inherited assertions.

## Deferred Issues

The final frozen no-call corpus run reports three route shifts before cgen: `adler_low.schway` and `checksum.schway` now refuse with `core.scalar_call_context_invalid`, and `examples/phase24/error.schway` now refuses with `core.scalar_backedge_authority`. Their historical receipts expected cgen/application-shell outcomes. The generated-C byte snapshots match; this plan does not edit core validation, weaken the corpus guard, or change interpreter/peer behavior. Details are in [`deferred-items.md`](./deferred-items.md), with open broken-windows entry 13 in `.planning/WINDOWS.md` for the owning phase to reconcile.

## Known Stubs

None found in the files created or modified by this plan.

## Threat Flags

None. The capacity marker uses the existing application evidence path and adds no endpoint, schema, or new trust boundary.

## Issues Encountered

No issue blocked the plan-specific verification lanes. The frozen corpus refusal shifts are out of this plan's owned core-validation scope and remain visible as deferred work.

## User Setup Required

None — installed Go and Clang toolchains were used; no dependency or service setup was added.

## Next Phase Readiness

Plan 36-05 is complete. Plans 36-06 through 36-08 remain, and CALL-01/CALL-02 remain phase-level requirements until their full obligations are complete. Preserve the frozen corpus guard and reconcile its three inherited refusal-route shifts in the core-validator owning phase.

---
*Plan: 36-05*
*Completed: 2026-10-14*

## Self-Check: PASSED

The summary, all plan-owned source/evidence/deferred files, and the windows ledger exist. All four task commits are ancestors of the current plan head.

## Dated residual resolution — 2026-10-14

The subsequent authorized `phase36-legacy-call-routes` debug investigation
restored the three historical application-emission routes without changing the
frozen source corpus or C expectations. The new scalar helper laws had selected
legacy owning-ByteView and fallible PathToken calls by their scalar result;
core/origin applicability and the independent PathToken contract gates were
corrected, with existing lifecycle, cleanup, error-edge and back-edge controls
retained. Held-out origin file-owner-tail tests also restored the established
owner refusal precedence.

Fresh exact roots `TestPhase36NoCallGeneratedCByteBaseline` and
`TestPhase36NoCallGeneratedCByteBaselineControls` both pass. New applicability,
file-owner and fallible-call tests pass alongside the Phase24/29/31 lifecycle
controls and Phase36 raw core/origin/path safety selectors. Seven fix-site
overlay mutants were killed by semantic assertions. Full bounded commands and
guardrail results are recorded in `.planning/debug/resolved/phase36-legacy-call-routes.md`;
the original failed verification above remains an accurate historical receipt.
The implementation residual is cleared for Plan36-06; debug review/archive and
the exact-file commit are owned by the session manager. This amendment adds no
hosted/native-matrix or phase-completion claim.

## Dated audit follow-up — 2026-10-15

This scoped continuation adds the Task 3 controls requested by the independent
Phase 36 security review while preserving the original plan history and task
count. The test-only follow-up is committed as `04ce6fa1`.

- **T-36-13:** `TestPhase36EmitterCallInventoryControls` now redirects the
  repeated nested helper's invocation-table pointer to a local stack array. A
  loop reuses that buffer, and a named retained-path assertion catches the
  missing first canonical child path after the helper returns. Clean, stack-copy
  no-op, and restored emitted C all run through the normal native result path
  and retain the first and final expected paths. The mutant compiles with the
  installed Clang and runs as native code with the correct returned value `10`;
  its raw execution document is inspected so the retained-path assertion, not
  earlier native path validation, identifies the lifetime failure.
- **T-36-16:** The same test now runs a region-scoped census against the emitted
  user function, keyed by checked call IDs. Separate live mutations remove and
  insert call, `function.called` event, and invocation-reservation sites. Each
  mutation yields a named missing/extra category and site finding, and the
  unmodified region passes again after every mutation. Support-library text is
  outside the census.
- No new RED evidence record was created: this audit follow-up changes test
  coverage only and has no product behavior to implement. The existing Task 3
  RED receipt remains the historical evidence for the original admission work.

## Audit follow-up verification

- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen -run '^TestPhase36EmitterCallInventoryControls$' -count=1 -v -timeout=600s` — PASS; clean/no-op/restored native positives, the stack-path semantic control, and all six region census mutations ran.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen ./internal/compiler/core -run '^(TestPhase36EmitterCallInventoryControls|TestPhase35EmitterAdmissionGates|TestPhase35UnsignedShiftEmissionAndEventProjection|TestInvocationPathTableBoundary|TestInvocationPreflightGuardIsNotInert|TestProgramInvocationIndexThreading|TestNativeFunctionCalledPreorder|TestPhase273NamedCapacityExitGate|TestPhase273CapacityContractLiteralsAgree|TestAllOperationKindsHandledAtEverySite)$' -count=1 -v -timeout=600s` — PASS; both package roots ran and all selected controls passed.

## Audit follow-up self-check: PASSED

The dated amendment is present, the scoped test commit `04ce6fa1` is an
ancestor of HEAD, and both recorded focused verification commands passed.

## Dated T-36-16 census correction — 2026-10-15

This amendment corrects the earlier T-36-16 follow-up: its call-site category
was still counted from `/* call: ... */` markers. The census now tracks markers
separately and enumerates actual `SCHWAY_` C function invocations inside the
bounded emitted user-function body. Expected call counts and symbols come from
the caller's checked `core.OpCall` contracts and checked callee functions; the
callee symbol is verified against both its emitted definition and the emitted
call expression. The invocation scan strips comments and string literals, so
neither can stand in for a call. It counts any `SCHWAY_` invocation, including
one with no expected call marker.

The live controls now remove the actual invocation while retaining its marker
and insert a second actual callee invocation without a marker. They require
named `missing call site "<callee-symbol>"` and `extra call site
"<callee-symbol>"` findings. Separate missing/extra marker, event, and
reservation controls remain, with clean/no-op/restored census checks. These
T-36-16 mutations are static census assertions; they make no C compile or
runtime semantic-kill claim. The existing T-36-13 returned-stack mutation and
native retained-path assertion remain in place. This test-only correction is
committed as `50b90a30`.

## T-36-16 correction verification

- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen -run '^TestPhase36EmitterCallInventoryControls$' -count=1 -v -timeout=600s` — PASS; all eight census mutations emitted their named findings, restored positives passed, and the existing native retained-path control ran.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen ./internal/compiler/core -run '^(TestPhase36EmitterCallInventoryControls|TestPhase35EmitterAdmissionGates|TestPhase35UnsignedShiftEmissionAndEventProjection|TestInvocationPathTableBoundary|TestInvocationPreflightGuardIsNotInert|TestProgramInvocationIndexThreading|TestNativeFunctionCalledPreorder|TestPhase273NamedCapacityExitGate|TestPhase273CapacityContractLiteralsAgree|TestAllOperationKindsHandledAtEverySite)$' -count=1 -v -timeout=600s` — PASS; both roots ran, including the native path control and all eight census mutations.
- `git diff --check` — PASS.

## T-36-16 correction self-check: PASSED

The amendment is present, scoped test commit `50b90a30` is an ancestor of
HEAD, and both focused verification commands passed.
