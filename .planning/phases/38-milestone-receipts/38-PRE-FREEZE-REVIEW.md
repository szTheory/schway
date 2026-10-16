---
phase: 38-milestone-receipts
reviewed: 2026-10-16T23:42:44Z
depth: standard
files_reviewed: 5
files_reviewed_list:
  - .github/workflows/ci.yml
  - scripts/verify-phase32.sh
  - scripts/check-milestone-binding.sh
  - testdata/receipts/M007.manifest
  - internal/compiler/session/session_phase38_receipts_test.go
source_sha: 02cfd716b0d53b390b558f1d6bf9ed96098816a5
codedigest: d4f72f9a81ca85874b99c9505412b8710937ad2d23934afe8655e4b072c168cd
source_hashes:
  .github/workflows/ci.yml: a8a913b4adfa36970bfd953e44ef541460acb302a9bfb78c75ddb05dc5b12c3b
  scripts/verify-phase32.sh: bffa9aeff47c7b9ba8393d73944f4203fabd926631e089f052f9165940e1eece
  scripts/check-milestone-binding.sh: 98d580f1f6aa83dcb53e5df398e8a8d89d6872bbc91e4f931cfc1f72d54bff53
  testdata/receipts/M007.manifest: 7fcfeeab05e89d9253006e2a333ac9b7e670fb2412d57a55a252b0041d515985
  internal/compiler/session/session_phase38_receipts_test.go: f7e71f123d93a407747e9acc0b4971249734ade3838c97acd523753c324a3d8b
findings:
  critical: 0
  warning: 1
  info: 0
  total: 1
status: issues_found
---

# Phase 38: Pre-freeze Code Review

**Reviewed:** 2026-10-16T23:42:44Z  
**Depth:** standard  
**Files Reviewed:** 5  
**Status:** issues_found

## Summary

I reviewed the five scoped files at source commit `02cfd716b0d53b390b558f1d6bf9ed96098816a5`. The supplied per-file SHA-256 values and codedigest match the working tree. Source inspection and the focused controls confirm the earlier producer lifecycle, event inventory, M007 row count, preparation-record, JSON type, and run-identity findings are fixed. The manifest retains the explicitly required `TestPhase35BitwiseOptimizationMatrix`; the earlier 16-root/341-event concern is closed because those were interim progress figures, not the plan's scope.

The remaining warning is a missing negative control for package-identity rejection. Both parsers compare each JSON event's `Package` with the resolved expected import path, and the new controls prove real package lifecycle records pass. The suite does not mutate that field and assert that a wrong-package event is refused.

**Newly run:** the bounded focused command `GOCACHE=/tmp/schway-phase38-review-gocache go test ./internal/compiler/session -run '^(TestPhase38ProducerInventoryControls|TestPhase38BindingControls|TestPhase38FreezeWorktreeControls|TestPhase38ManifestConsumerBindings|TestPhase38WorkflowFailureControls)$' -count=1 -v -timeout=120s` passed. It exercised the real bounded producer capture, complete Go package lifecycle through the binding parser, inventory count controls, preparation-record refusals, freeze/workflow controls, and existing M006 row preservation. I also recomputed all five file hashes, the codedigest, and the manifest totals (106 rows, 17 roots, 480 events, 119 reached bindings). The executor's focused union and standalone CRC probe are prior receipts, not checks I reran. No full M007 producer, full/race suite, broad matrix, pressure test, or hosted workflow was run.

## Narrative Findings (AI reviewer)

### Warnings

#### WR-01: [WARNING] Wrong-package event rejection has no negative control

**File:** `internal/compiler/session/session_phase38_receipts_test.go:228-256,288-305`  
**Severity:** WARNING

**Issue:** The producer and binding parser both reject events whose `Package` differs from the package import path resolved from the manifest (producer at `scripts/verify-phase32.sh:355-357`; binding parser at `scripts/check-milestone-binding.sh:308-310`). The focused suite proves valid package lifecycle events pass, and its malformed-event cases cover skips, missing names, duplicate events, and incomplete streams. It does not mutate `Package` to a different valid import path and assert refusal at either production parser. This leaves the package-identity guard without a regression control at a boundary used to validate supplied receipt logs.

**Fix:** Add a binding-fixture mutation that changes the package field while keeping all other events and log metadata valid, and assert `control_event_inventory`. Add a bounded producer control that supplies a complete lifecycle stream with the wrong package through the real producer parser and asserts the producer fails without a host-result success marker.

---

_Reviewed: 2026-10-16T23:42:44Z_  
_Reviewer: the agent (gsd-code-reviewer)_  
_Depth: standard_

