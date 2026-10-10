---
phase: 27-exact-fizzbuzz-and-cross-host-evidence
plan: 09
subsystem: testing
tags: [cross-host-evidence, receipts, ci-matrix, native-apps]

requires:
  - phase: 27-exact-fizzbuzz-and-cross-host-evidence
    provides: Exact FizzBuzz and sum applications, output ceiling, and independent controls
provides:
  - POSIX native host driver emitting a provenance-bearing per-host receipt
  - Fail-closed two-host (macOS + Linux) receipt gate
  - CI wiring of the driver and per-host artifact upload in both evidence-aggregate lanes
affects: [phase-27-10-hosted-receipts]

status: complete
commits: 2
plan_head_before: 25dc503f2c66f9bcdca386dadf9844b06f6ac796
plan_head_after: 27737e8
actuals:
  tokens: 14000
  tasks: 2
  commits: 2

tech-stack:
  added: []
  patterns: [identity-case-control-result receipt rows, gate recomputes expected hashes from checkout, mutation-controlled gate test]
key-files:
  created:
    - scripts/verify-phase27.sh
    - scripts/check-phase27-receipts.sh
    - internal/compiler/session/session_phase27_host_test.go
  modified:
    - .github/workflows/ci.yml
key-decisions:
  - "The gate recomputes expected source/fixture hashes from its own checkout and takes the revision from HEAD (or PHASE27_EXPECTED_REVISION), so a log can never vouch for its own inputs."
  - "The gate requires tree=clean; a local dirty-tree run is a development receipt only and cannot satisfy acceptance."
  - "The driver step runs first in evidence-aggregate with an always() artifact upload, so its log survives failures of later steps."
duration: 25min
completed: 2026-10-10
---

# Phase 27 Plan 09: Native Host Driver and Two-Host Receipt Gate Summary

POSIX driver builds the committed FizzBuzz and sum_to_n apps on the actual host and runs the 0/1/15/1000/1001 cases with exact stdout, stderr and exit checks. A separate gate accepts only a clean, same-revision, same-hash macOS+Linux pair, and both CI host lanes now run the driver and upload their logs.

## Accomplishments

- `scripts/verify-phase27.sh` builds `cmd/schway`, builds both examples, compares stdout byte-for-byte to the committed 15/1000 fixtures and literal values, checks stderr and exit status separately (1001 gives empty stdout, exit 65, `schway app: input exceeds 1000`), then runs the native dynamic-zero and output-limit controls. Rows: one identity row (host, arch, Go, Clang target, revision, tree, source/fixture git hashes, input set), 10 case rows, 3 control rows, and a final result row that is only `pass` after everything finishes. Failures print a named `status=fail` result row.
- `scripts/check-phase27-receipts.sh` rejects missing, empty, duplicate-host, unavailable, failed, dirty-tree, wrong-revision, altered-hash, wrong-input-set, missing/altered case or control, and malformed rows.
- `TestPhase27HostReceiptGate` has a known-good pair (literal rows, only file hashes measured) and 21 mutation controls. `TestPhase27CIHostMatrix` pins both matrix OSes, the driver step, `shell: bash` (pipefail), `if: always()` upload, host-specific artifact names, and the retained baseline steps and full/race lanes.
- Local development receipt (not dual-host evidence): Darwin arm64, Go 1.24.0, Apple Clang 21.0.0, 10/10 cases and 3/3 controls passed, tree=modified (pre-existing `.planning/config.json`), so the gate would reject it.

## Task Commits

1. Task 1 (driver, gate, gate test): d43f5f3
2. Task 2 (CI wiring, CI contract test): 27737e8

## Deviations from Plan

None - plan executed as written. The plan's `<verify>` for Task 1 includes `sh scripts/verify-phase27.sh`, which passed locally.

## Open Items

EVD-12 remains open: no hosted macOS or Linux run exists yet for Phase 27. Plan 27-10 must acquire same-revision hosted logs and run the gate. Phase 25/26 receipts and this local run do not satisfy it.

## Known Stubs

None.

## Threat Flags

None. The CI change reuses existing pinned action majors (checkout@v4, setup-go@v5, upload-artifact@v4) and adds no package installs.

## Self-Check: PASSED

Files present: scripts/verify-phase27.sh, scripts/check-phase27-receipts.sh, internal/compiler/session/session_phase27_host_test.go. Commits d43f5f3 and 27737e8 exist on phase26-validation.
