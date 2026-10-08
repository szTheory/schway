---
id: SEED-261008-trz
status: dormant
planted: 2026-10-08
planted_during: M005/Phase 27 quick task
trigger_when: A named program failure cannot be diagnosed from current capture, observers, debug map, and source replay evidence.
scope: One optional local read-only diagnostic bound to a named build and consumer failure.
---

# SEED-261008-trz: Consumer-triggered failure observability

## Why This Matters

Schway already captures application evidence with explicit disabled, incomplete, and capacity-exhausted states; replay distinguishes source conformance from verifier-model evidence; physical observer controls reject plausible but false cleanup traces. Broader observability is justified only by a concrete failure these mechanisms cannot explain.

## When to Surface

**Trigger:** A named failure in a selected program remains unexplained after current capture, physical observers, semantic debug map, and replay evidence are reviewed. Do not surface for generic “agents need telemetry” arguments.

## Smallest Consumer and Obligations

Compare current evidence against one optional, local, read-only diagnostic for a named checksum/file utility failure, bound to exact build/source identity and semantic IDs. Preserve original process outcome and report missing, truncated, or unavailable evidence distinctly. Before expanding the trust boundary, address `os.ReadFile(capturePath)` preceding the capacity check in `internal/compiler/native/native_app.go`; source inspection shows the classification limit does not bound allocation, but this is not a reproduced exploit. No mandatory VM, tracing heap, resident service, upload, evaluation, or mutation authority.

## Independent Answers and Negative Controls

Use an independently specified expected diagnosis. Control wrong-build attribution, stale source, missing/oversized/truncated capture, disabled capture, and redacted fields. Verify that diagnostic failure preserves process outcome and compare enabled/disabled overhead with cold/warm p50/p95. Define minimal fields, local retention, consent, and sensitivity of paths and structural IDs; hashes do not anonymize. Reject the extension if current evidence suffices, if it cannot keep authority read-only, or if privacy/cost outweighs the named diagnostic gain.

## Cost, Owner, and Next Action

**Cost:** One opt-in diagnostic and bounded evidence path for one consumer, with allocation, retention, privacy, and performance review.

**Owner/next action:** The selected consumer phase owner, with native/evidence and security review, should at a relevant phase transition, name the failure and demonstrate the current evidence gap before promoting through normal roadmap/requirements ownership.

## Supersession and Breadcrumbs

Supersede if current evidence explains the failure. Reject generic telemetry or any design requiring ambient runtime authority. Existing design context: `wiki/debug-evidence-symbolication-and-proof.md`; current implementation: `internal/compiler/native/native_app.go`, `internal/compiler/native/native_app_test.go`, `internal/compiler/session/session_app_verify.go`, `internal/compiler/native/phase23_observer_test.go`, and `internal/compiler/native/phase24_observer_test.go`. Source-inspected test anchors include `TestPhase22EvidenceDisabledCompleteAndStreamIsolation`, `TestPhase22EvidenceMissingPartialAndCapacityControls`, `TestPhase22EvidenceReportWriteFailurePreservesAppOutcome`, `TestPhase23ObserverMutationControlsPreservePlausibleEvents`, and `TestPhase24ObserverReachedPhysicalDestructorControls`. See the [Dashbit assessment](../sources/DASHBIT-AI-ERA.md).
