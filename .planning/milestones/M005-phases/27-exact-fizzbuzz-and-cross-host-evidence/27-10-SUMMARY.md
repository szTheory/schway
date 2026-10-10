---
phase: 27-exact-fizzbuzz-and-cross-host-evidence
plan: 10
subsystem: testing
tags: [documentation, cross-host-evidence, hosted-ci, receipts]

requires:
  - phase: 27-exact-fizzbuzz-and-cross-host-evidence
    provides: Native host driver, two-host receipt gate, CI wiring
provides:
  - Single root-linked clean-checkout guide for FizzBuzz and sum_to_n
  - Hosted macOS and Linux Phase 27 receipts at one revision, gate passing
  - Repaired full-suite regressions from earlier Phase 27 plans
affects: [phase-27-verification]

status: complete
commits: 3
plan_head_before: 795c4eb
plan_head_after: 2563745a6477dcca9b572f5464378511d0518985
actuals:
  tokens: 30000
  tasks: 2
  commits: 3

key-files:
  created:
    - examples/phase27/README.md
    - internal/compiler/session/session_phase27_docs_test.go
    - .planning/milestones/M005-phases/27-exact-fizzbuzz-and-cross-host-evidence/27-HOST-EVIDENCE.md
    - .planning/milestones/M005-phases/27-exact-fizzbuzz-and-cross-host-evidence/host-macos.receipt
    - .planning/milestones/M005-phases/27-exact-fizzbuzz-and-cross-host-evidence/host-linux.receipt
  modified:
    - README.md
    - internal/compiler/syntax/format.go
    - internal/compiler/cgen/cgen.go
key-decisions:
  - "The hosted receipts are claimed only from run 38052601527 at fa5a169; the earlier failing run at c7bf17c is recorded as superseded."
  - "Receipt files are the verbatim phase27 log rows from the downloaded per-host artifacts."
requirements-completed: [DX-16, EVD-12, APP-08, APP-09, U64-02]
duration: 95min
completed: 2026-10-10
---

# Phase 27 Plan 10: Guide and Hosted Receipts Summary

One root-linked guide now reproduces both programs from a clean checkout, and a hosted macOS (arm64) plus Linux (x86_64) run at `fa5a169` produced 10/10 cases and 3/3 controls each, accepted by the two-host gate.

## Accomplishments

- `examples/phase27/README.md` is the single command owner: build and run commands, exact answers, 0 to 1000 bounds, 1001 pre-output refusal, 65536-byte output limit, remainder-by-zero behavior, focused controls, two-host gate, and historical receipt scope. Root `README.md` links it once; `TestPhase27GuideContract` pins the link, sections, and named paths.
- Guide commands were followed in fresh clones at `c7bf17c` and `fa5a169` (cmp against committed fixtures, 55/500500, 1001 gives exit 65).
- Hosted run 38052601527 (https://github.com/szTheory/schway/actions/runs/38052601527) succeeded: both `checks` and both `current evidence aggregate` jobs. `sh scripts/check-phase27-receipts.sh` with `PHASE27_EXPECTED_REVISION=fa5a169216b60cce94929c881001b1ff02fbf09a` printed `status=pass hosts=macOS,Linux`.
- Provenance, hashes, and job IDs are in `27-HOST-EVIDENCE.md`.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] First hosted run (38052011600 at c7bf17c) failed the full-suite lanes on both hosts**
- **Found during:** Task 2 (hosted verification)
- **Issue:** Earlier Phase 27 plans were only checked with focused selectors, so these regressions were never seen: legacy non-application `schway_write_bytes` bytes changed (frozen emitter goldens, manifest and evidence IDs moved); the canonical formatter split `let x = y % N` across lines and broke the fizzbuzz formatting test; guard inventories were stale (LinearOperation fields, operation-kind dispatch corpus, emitter call-site registry, LANGUAGE-MATURITY counts); the host gate test spawned unbounded processes; `27-VALIDATION.md` was draft and had a go-test cell the groundedness scanner rejects.
- **Fix:** Restored the legacy writer bytes, kept remainder bindings on one line (added `TestPhase27FormatterRemainder`), registered Phase 27 fields, ops, call sites and counts, used the bounded `testsupport.RunCLIErr` spawn path, and marked 27-VALIDATION complete with scanner-safe wording. The tracer formatting test now formats a dedicated commented source.
- **Files modified:** internal/compiler/cgen/cgen.go, internal/compiler/syntax/format.go, internal/compiler/core/core_test.go, internal/compiler/core/core_convention_absence_test.go, internal/compiler/session/session_phase27_test.go, internal/compiler/session/session_phase27_host_test.go, testdata/phase16/public-emitter-consumers.json, .planning/LANGUAGE-MATURITY.md, 27-VALIDATION.md
- **Commit:** fa5a169
- Locally verified only the previously failing selectors and the fast core, evidence, cgen, syntax, and cmd packages; the hosted rerun passed the full and race lanes.

Otherwise the plan executed as written, with Task 2's push and dispatch under the user's explicit authorization (fast-forward pushes only, twice).

## Open Items

Commits after `fa5a169` (the evidence commit and this summary) touch only `.planning/` files, so receipts remain valid for the claimed revision. Local commits remain unpushed after the push at `fa5a169`.

## Known Stubs

None.

## Threat Flags

None.

## Self-Check: PASSED

Files present: examples/phase27/README.md, session_phase27_docs_test.go, 27-HOST-EVIDENCE.md, both receipt files. Commits c7bf17c, fa5a169, 2563745 exist on phase26-validation.
