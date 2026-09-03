---
phase: 02-owned-values-and-abilities
plan: "05"
subsystem: compiler
tags: [go, c17, native-execution, bounded-io, strict-json, semantic-differential, tdd]

requires:
  - phase: 02-owned-values-and-abilities
    plan: "04"
    provides: Independently validated owned core and fail-closed interpreter/native engine entry
provides:
  - Readable inline-value owned C17 driven by validated explicit operations
  - Four separately bounded compile/run stdout/stderr streams with stable operational codes
  - Strict stdout-only single-document execution decoding
  - Exact interpreter/O0/O3 outcome, event, identity, order, and live-resource comparison
affects: [02-06, owned-evidence, phase2-verification, native-boundary]

actuals:
  tokens: 8361
  tasks: 2
  commits: 4

tech-stack:
  added: []
  patterns: [max-plus-one stream writer, stdout-only strict decode, engine-neutral semantic equality, legacy evidence byte preservation]

key-files:
  created:
    - internal/compiler/execution/execution_test.go
    - internal/compiler/native/native_test.go
    - testdata/phase2/owned_transfer.golden.c
  modified:
    - internal/compiler/execution/execution.go
    - internal/compiler/cgen/cgen.go
    - internal/compiler/native/native.go
    - internal/compiler/session/session.go
    - internal/compiler/session/session_test.go

key-decisions:
  - "Each compile stdout, compile stderr, run stdout, and run stderr stream owns an independent 64 KiB max-plus-one writer; no process stream is merged."
  - "Native stdout carries exactly one strict execution document, while any successful-run stderr is an operational failure and never enters semantic decoding."
  - "Canonical execution equality covers every outcome, ordered event, semantic ID/place/type field, and live resource while excluding optimization and process observations."
  - "Legacy match Emit bytes remain frozen for Phase 1 evidence; native execution selects a dedicated strict-JSON emission path."

patterns-established:
  - "Bounded child boundary: overflow is classified per stage and stream before process output can reach decoding or diagnostics."
  - "Semantic differential: decoded native facts are compared byte-canonically against the interpreter rather than translated through backend-specific output strings."

requirements-progressed: [OWN-01]

coverage:
  - id: D1
    description: "Validated owned operations emit readable C17 and produce identical interpreter, O0, and O3 executions through four independently bounded streams."
    requirement: OWN-01
    verification:
      - kind: integration
        ref: "internal/compiler/native/native_test.go#TestNativeStreamsIndependentlyBounded,TestNativeNeverUsesCombinedOutput"
        status: pass
      - kind: e2e
        ref: "internal/compiler/session/session_test.go#TestOwnedTransferInterpreterNative,TestNativeToolFailureIsOperational"
        status: pass
    human_judgment: false
  - id: D2
    description: "Malformed native stdout fails operationally while every decoded outcome/event/identity/order/resource mutation fails as semantic mismatch."
    requirement: OWN-01
    verification:
      - kind: unit
        ref: "internal/compiler/native/native_test.go#TestExecutionDecoderRejectsMalformedOutput"
        status: pass
      - kind: integration
        ref: "internal/compiler/session/session_test.go#TestOwnedEventReorderIsMismatch,TestOwnedExecutionFieldMutationMatrix"
        status: pass
    human_judgment: false

duration: 12min
completed: 2026-09-03
status: complete
---

# Phase 02 Plan 05: Bounded Owned Native Differential Summary

**Validated owned core now lowers to reviewable inline-value C17 whose strict bounded O0/O3 execution documents match the interpreter on every ordered semantic fact.**

## Performance

- **Duration:** 12 min
- **Started:** 2026-09-03T21:18:19Z
- **Completed:** 2026-09-03T21:29:43Z
- **Tasks:** 2
- **Files modified:** 8

## Accomplishments

- Replaced merged process capture with independent 64 KiB max-plus-one writers for compile stdout, compile stderr, run stdout, and run stderr, each with a stage/stream-specific truncation code.
- Added readable C17 for owned Buffer transfer and Byte copy using stack/inline values and explicit checked operations without heap, cleanup, alias, provenance, layout, or ABI claims.
- Made native execution a strict stdout-only JSON consumer that rejects unknown, missing, trailing, duplicate, malformed, truncated, or oversized facts as operational failures.
- Replaced output-string comparison with exact canonical outcome, ordered-event, semantic-ID/place/type, and live-resource equivalence, preserving semantic mismatch exit 4 separately from operational exit 3.

## Task Commits

Each TDD task was committed with a RED contract followed by its GREEN implementation:

1. **Task 02-05-01 RED: bounded owned native boundary contracts** - `a1143d4` (test)
2. **Task 02-05-01 GREEN: validated owned C17 and four bounded streams** - `cc64e8b` (feat)
3. **Task 02-05-02 RED: malformed fact and semantic mutation contracts** - `8f3c589` (test)
4. **Task 02-05-02 GREEN: strict decode and complete semantic differential** - `203b311` (feat)

## Files Created/Modified

- `internal/compiler/execution/execution.go` - Canonical equality for complete inert execution facts.
- `internal/compiler/execution/execution_test.go` - Guard that physical process observations stay outside semantic bytes.
- `internal/compiler/cgen/cgen.go` - Validated match/linear native emission and readable inline Buffer ownership operations.
- `internal/compiler/native/native.go` - Four independent bounded writers, shell-free execution, stderr policy, and strict execution decoding.
- `internal/compiler/native/native_test.go` - Per-stream flood helpers plus unknown/missing/trailing/duplicate/malformed/truncated/oversized controls.
- `internal/compiler/session/session.go` - Native runner interface, complete engine comparison, and mismatch classification.
- `internal/compiler/session/session_test.go` - Owned O0/O3 tracer, field/order mutation matrix, and exit-taxonomy checks.
- `testdata/phase2/owned_transfer.golden.c` - Reviewable committed C17 for the canonical Buffer transfer.

## Decisions Made

- Successful native runs reject any stderr with `native.run_stderr`; stderr remains bounded operational detail and is never parsed as semantic output.
- Stream overflow takes precedence over decoding and receives stable compile/run plus stdout/stderr codes, so flooding cannot masquerade as malformed semantics.
- The runner reports process timing and byte counts only outside `execution.Execution`; optimization labels, addresses, timing, and process details cannot alter semantic bytes.
- Session accepts a narrow runner interface so deterministic test doubles can inject decoded semantic mutations without weakening the production process boundary.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Preserved frozen Phase 1 C evidence bytes while adding JSON native output**

- **Found during:** Task 02-05-01 full-suite verification
- **Issue:** Making the legacy match emitter print execution JSON changed the Phase 1 C digest and failed its canonical evidence golden.
- **Fix:** Kept `cgen.Emit` byte-identical for legacy evidence and added `cgen.EmitNative` to select strict JSON output only for actual native execution; owned linear emission uses the new path in both cases.
- **Files modified:** `internal/compiler/cgen/cgen.go`, `internal/compiler/session/session.go`
- **Verification:** `env GOCACHE=/tmp/ai-lang-phase2-cache go test ./...` and `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/verify-phase1.sh`
- **Committed in:** `cc64e8b`

---

**Total deviations:** 1 auto-fixed bug
**Impact on plan:** The fix preserved Phase 1 reproducibility while retaining the required strict native protocol; no feature scope expanded.

## Issues Encountered

- The second decoder micro-cycle showed that Go's JSON decoder accepts `{}` as syntactically valid. Required execution-document and event fields are now checked explicitly before decoded facts leave the native boundary.

## Verification Evidence

- Both exact fail-on-zero selectors passed every Plan 02-05 named test.
- `env GOCACHE=/tmp/ai-lang-phase2-cache go test ./...` passed across all packages.
- `env GOCACHE=/tmp/ai-lang-phase2-cache go test -race ./...` and `env GOCACHE=/tmp/ai-lang-phase2-cache go vet ./...` passed.
- `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/verify-phase1.sh` exited 0 with all five lanes passing and 18 recomputed work units.
- Static scanning found no `CombinedOutput` in `internal/compiler/native/native.go`, and `git diff --check` reported no whitespace errors.

## Known Stubs

None.

## Threat Flags

None. The new child-process and backend-differential surfaces are the planned T-02-06 and T-02-07 mitigations; no additional trust boundary was introduced.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plan 02-06 can bind validated owned core, generated C, and exact ordered executions into evidence and the Phase 2 verification gate.
- `OWN-01` is progressed but remains pending until final Phase 2 verification; neither `OWN-01` nor `OWN-02` was marked complete here.
- No blockers remain for sequential Phase 2 execution.

## Self-Check: PASSED

All eight changed implementation/test/golden files, this summary, and all four RED/GREEN commits were found. The summary contains `requirements-progressed: [OWN-01]`; `REQUIREMENTS.md` and `.planning/milestone.lock` were not modified.

---
*Phase: 02-owned-values-and-abilities*
*Completed: 2026-09-03*
