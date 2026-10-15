---
status: resolved
trigger: "Phase36 Plan36-07 preserved Phase24 ownership-transfer witnesses refused by PathOracle"
created: 2026-10-15
updated: 2026-10-15
---

## Current Focus

bug_class: bohrbug
hypothesis: pathScalarCalleeValueOnly conflates local scalar borrows copied before return with transferred authority.
test: Fresh named regression/control roots, exact Plan36-07 Task2 command, frozen no-call baseline and controls, semantic single-site overlay mutation and revert/reapply.
expecting: Safe shared/exclusive borrow-copy callees and full retained witnesses pass; raw loan/resource/provenance/backedge corruptions refuse; no-op/restored overlays pass and reverted/authority-mutant overlays fail at specific semantic assertions.
next_action: Resolved after independent review and manager verification; resume Plan36-07 through its executor, preserving the separately owned consumer hunk for its Task2 commit.
reasoning_checkpoint:
  hypothesis: The scalar callee classifier rejects safe local borrow-copy results because it equates any local LoanID with authority crossing the call.
  confirming_evidence:
    - Both new source-derived shared/exclusive helpers pass independent local pointer replay but fail ValidateLocalOwnerPaths at scalar_call_context (fresh RED log).
    - Unchanged Phase35 full witness consumer and Phase34 mutation baseline fail on the frozen Phase24 transfer source facets; the only relevant callee rejection is its borrow operation/LoanID.
  falsification_test: If admitting only the existing source-blind pathU64CopySuccessorHelper shape leaves the frozen consumer failing, the proposed cause is insufficient; if unsafe raw neighbors pass, the boundary is unsound.
  fix_rationale: Reuse the already independently checked three-operation U64 borrow/copy/owned-return shape, require clean operation authority fields and pointer replay, while preserving all ordinary scalar signature/freshness checks and owner-path replay.
  blind_spots: No general borrowed scalar CFG or nested-call helper support is claimed; these remain refused. Cross-host execution is outside this bounded correction.
  candidate_causes:
    - code: local callee loan presence is confused with exported authority.
    - data: frozen legacy source or expected outcomes changed (eliminated by immutable bytes and source-derived reproduction).
    - environment: host/cache acquisition failure (eliminated by source-blind deterministic failure before engine execution).
  and_gate: no; the rejecting callee predicate alone explains the failure, with no environment or data defect required.

## Symptoms

expected: Existing examples/phase24/transfer.schway passes the complete session path used by eight immutable Phase34/35 surface witnesses, preserving independent loan/authority proof.
actual: The eight witnesses fail before copied-data mutation controls with invalid/core.pathoracle_refused.
errors: pathoracle.scalar_call_context: caller "s1:phase24.transfer:fn:main" operation "s1:phase24.transfer:fn:main:op:2" has an invalid scalar call contract or carries non-value authority
timeline: Exposed in Plan36-07 inherited witness verification after scalar call rules in Phase36; old witness bytes remain frozen.
reproduction: Read .planning/phases/36-scalar-helper-calls/deferred-items.md for the exact Task2 command and eight witness IDs. examples/phase24/transfer.schway calls shared_copy(value) and exclusive_copy(shared), whose bodies use borrow and borrow mut. TestPhase34SurfaceManifestWitnesses is an exact retained full-session consumer.

## Evidence

- Plan36-07 Task1 passed; Task2 new census/document/limit roots pass, inherited full-session witnesses and controls fail on the same eight cases.
- Plan36-07 owned work is uncommitted and must be preserved; the executor is idle awaiting this scoped production correction.
- Two earlier resolved compatibility repairs remain authoritative: phase36-legacy-call-routes and phase36-byteview-helper.
- 2026-10-15: Read active state, project/GSD workflow and references. No project skills exist; agent-skills query adds no separate skill. KB Phase25 local-loan/OpCall provenance pattern is a candidate only. SBFL skipped: no per-test coverage. Common-pattern category is contract/scope overreach.
- 2026-10-15: pathScalarCalleeValueOnly rejects any LoanID and admits neither borrow operation. transfer.schway shared_copy/exclusive_copy are straight-line borrow -> copy -> return helpers. Core/origin scalar call contract gates do not forbid local borrows. Existing PathOracle pointer replay already rejects direct loan escape and shared/exclusive overlap independent of serialized endpoints.
- 2026-10-15: Ran bash /tmp/schway34-root-pin.sh successfully before first edit; preserved executor dirty files and prior corrections.
- 2026-10-15: Fresh RED: GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/pathoracle -run '^TestPhase36ScalarCallLocalBorrowCopies$' -count=1 -v -timeout=600s. Both borrow modes fail scalar_call_context after pointer replay succeeds. Log /tmp/schway36-borrow-red-oracle.log.
- 2026-10-15: Fresh RED: GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase35SurfaceManifestWitnesses|TestPhase34SurfaceManifestControls)$' -count=1 -v -timeout=600s. Both fail on retained owner/borrow witness outcomes. Log /tmp/schway36-borrow-red-full-session.log. TestPhase34SurfaceManifestWitnesses alone passes before correction because it only compares u64.add, despite observing every source; this is not full witness acceptance evidence.
- 2026-10-15: New oracle regression/controls, Phase36 scalar authority controls, and eight retained Phase25 loan/ownership roots pass fresh in /tmp/schway36-borrow-green-oracle.log. Full retained Phase35 witnesses, Phase34 controls, and Phase36 call controls now pass in /tmp/schway36-borrow-green-surface.log. Exact Task2 command exposes an independent executor-owned TestPhase35SurfaceManifestControls missing-refusal mutation (line 197); manager notified and no protected file edited.
- 2026-10-15: Reviewer identified CalleeID as an additional OpCall-only field not included in the generic call-field cleaner. Fresh narrow raw control RED in /tmp/schway36-borrow-red-callee.log: all three helper ops accepted stray CalleeID. Add explicit empty CalleeID requirement within the new helper branch.
- 2026-10-15: Manager authorized ONLY the TestPhase35SurfaceManifestControls copied-depth-witness hunk: Phase36 adds a second active depth refusal witness, so deleting only phase35-depth-65 no longer removes facet coverage. Remove active copied-data coverage for that exact refusal facet and assert restored positive; keep shared file uncommitted for executor ownership.
- 2026-10-15: Frozen no-call baseline and Controls, Phase24 transfer discharge/mutations, Phase25 independent source/raw overlap/escape controls, legacy fallible/file-owner peers pass in /tmp/schway36-borrow-green-compat.log. Existing raw scalar peers/refusal-removal plus Phase29 release/owner-backedge and all three peers' Phase31 file-byte seeded faults pass in /tmp/schway36-borrow-green-raw-peers.log.
- 2026-10-15: Final oracle correction (including non-call CalleeID guard) passes new positives, 74 raw authority neighbors, existing scalar authority/refusal-removal and retained loan/owner roots: /tmp/schway36-borrow-final-oracle.log. Exact Task2 command passes all 15 selected roots across session/surfacemanifest after the narrowly authorized copied-data depth-control repair: /tmp/schway36-borrow-final-surface.log. Executor-only consumer delta is /tmp/schway36-borrow-phase35-consumer.patch against /tmp/schway36-borrow-phase35-before.go.
- 2026-10-15: No-op and restored Go overlays pass the source-derived two-mode regression; reverted production overlay reproduces both scalar_call_context failures; one-site replacement of the new authority-field predicate with false is killed by TestPhase36ScalarCallLocalBorrowControls/borrow/borrow/foreign, which observes unsafe raw acceptance (<nil>). No compile/refusal/skip/zero-selector kill. Logs: /tmp/schway36-borrow-noop.log, /tmp/schway36-borrow-revert.log, /tmp/schway36-borrow-authority-mutant.log, /tmp/schway36-borrow-restored.log. Reproducible recipe: python3 /tmp/schway36-borrow-mutation.py; overlays at /tmp/schway36-borrow-overlays/. Shared source never overwritten.
- 2026-10-15: Frozen Phase34/35 witness/census SHA256 bytes match all four hashes in deferred-items.md; git diff --check clean on owned production/tests and narrow shared consumer hunk. No corpus/golden/dependency/schema/backend or public API change.
- 2026-10-15: Manager independently reran all four exact commands below after final source changes. All 46 selected test-root occurrences passed with no skips or zero-selector results: 11 oracle roots, 15 exact Task2 surface roots, 9 frozen/ownership compatibility roots, and 11 independent raw-peer roots. Logs: /tmp/schway36-borrow-manager-oracle.log, /tmp/schway36-borrow-manager-surface.log, /tmp/schway36-borrow-manager-compat.log, /tmp/schway36-borrow-manager-raw-peers.log.
- 2026-10-15: Manager independently reran python3 /tmp/schway36-borrow-mutation.py. No-op/restored positives passed, reverted correction reproduced both specific scalar_call_context failures, and the authority-predicate mutant failed at TestPhase36ScalarCallLocalBorrowControls/borrow/borrow/foreign because malformed raw authority was accepted (<nil>); no compile/refusal/skip/zero-selector counted as a kill.
- 2026-10-15: Independent typed code reviewer approved both the closed production gate and the narrowly authorized copied-data consumer hunk; reviewer made static-review claims only. Manager fresh behavioral commands supply execution evidence.
- 2026-10-15: Root pin and exact-file whitespace/staging checks passed before canonical commit 1c32b347, containing only pathoracle.go and pathoracle_phase36_borrow_test.go. Shared session_phase35_surface_test.go stays uncommitted with executor-owned Plan36-07 changes. No phase/roadmap/state or unrelated record was advanced.
- 2026-10-15: Final compatibility and raw-peer commands rerun after the CalleeID guard are GREEN in /tmp/schway36-borrow-final-compat.log and /tmp/schway36-borrow-final-raw-peers.log. Host/toolchain: Darwin arm64, go1.24.0 darwin/arm64. Local automated evidence only; no native matrix, cross-host, hosted, or UAT claim. Production/tests stable for manager independent rerun.
- 2026-10-15: Independent reviewer approved the final closed production boundary including CalleeID and the separately authorized narrow consumer-control hunk. Static review only; manager owns independent fresh rerun, exact file commit, archive and Plan36-07 deferred disposition.

## Eliminated

- data/source drift: frozen retained witnesses remain present and the new independent source helper reproduces the same refusal.
- host/cache owner behavior: both minimal helpers fail without any owner operation or execution.

## Resolution

root_cause:
  The scalar callee classifier rejected every local borrow/LoanID, conflating safe copied owned U64 results with non-value authority transfer before the existing independent loan and owner checks could run.
fix:
  Admit only the existing source-blind three-operation U64 borrow/copy/owned-return shape, require all authority fields clean (LoanID allowed only on its borrow), and replay its straight-line loans independently. Preserve ordinary scalar and owner gates unchanged.
oracle_type: specified (owned scalar copy result, no borrowed/resource/error/provenance authority crossing calls)
verification:
  target_test: { result: pass, log: /tmp/schway36-borrow-final-oracle.log }
  mutation_check: { result: pass, mutant_killed: true, method: Go source overlay at the repaired authority-field predicate, log: /tmp/schway36-borrow-authority-mutant.log }
  no_op_deletion: { result: pass, deletion_justified_by_rca: false, detail: additive semantic eligibility with unchanged ordinary call/owner gates; no-op and restored positives pass }
  adjacent_tests: { result: pass, suites_run: [retained full surface witnesses and mutation controls, frozen no-call C baseline and Controls, raw scalar peers, shared/exclusive overlap and escape, owner discharge and backedge controls], logs: [/tmp/schway36-borrow-final-surface.log, /tmp/schway36-borrow-final-compat.log, /tmp/schway36-borrow-final-raw-peers.log] }
  revert_and_reconfirm: { result: pass, bug_returned_on_revert: true, fixed_on_reapply: true, logs: [/tmp/schway36-borrow-revert.log, /tmp/schway36-borrow-restored.log] }
  guardrail_verdict: accepted
  scope: local automated checks only; independent static peer review and manager fresh rerun complete; no human UAT, full-suite, broad native matrix, or cross-host claim
  manager_logs: [/tmp/schway36-borrow-manager-oracle.log, /tmp/schway36-borrow-manager-surface.log, /tmp/schway36-borrow-manager-compat.log, /tmp/schway36-borrow-manager-raw-peers.log]
  code_commit: 1c32b347
cycles: { investigation: 1, fix: 2 }
files_changed:
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/pathoracle/pathoracle_phase36_borrow_test.go
consumer_disposition:
  file: internal/compiler/session/session_phase35_surface_test.go
  ownership: executor; keep uncommitted with executor's existing edits
  change: copied-data control removes every active witness for the exact depth-refusal facet, still requires missing_refusal and restored positive
  patch: /tmp/schway36-borrow-phase35-consumer.patch

## Specialist Review

specialist_hint: go
skill: none (the Go hint has no specialist skill mapping)
independent_peer: Typed gsd-code-reviewer approved the closed helper shape, clean fields including CalleeID, independent loan replay, and refusal of CFG/nested-call/borrowed-return forms. The peer also approved the exact-facet copied-data control correction. Static review only; execution evidence is recorded separately above.

## Blameless Postmortem

why_not_caught: The new scalar callee classifier's focused admission tests did not include legacy local shared/exclusive borrow-copy callees; the retained complete-program surface consumers exposed that compatibility gap.
guard: TestPhase36ScalarCallLocalBorrowCopies and TestPhase36ScalarCallLocalBorrowControls preserve both access modes and 74 malformed authority neighbors at the repaired gate, with revert and semantic mutation controls; unchanged full-session witnesses remain live in the exact Task2 selector.
prevention: Keep local loan presence separate from authority crossing a call; require explicit independent replay and the closed helper shape before admitting any exception to scalar callee eligibility.

## Exact focused verification commands

Each command runs from /Users/jon/projects/schway. No full suite or broad native matrix was run.

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/pathoracle -run '^(TestPhase36ScalarCallLocalBorrowCopies|TestPhase36ScalarCallLocalBorrowControls|TestPhase36ScalarCallAuthorityBoundaries|TestPhase25PointerPathLiveOverlap|TestPhase25PointerPathBorrowedResultEscape|TestPhase25PointerPathDefersCallResultProvenance|TestPhase25PointerPathAllowsDeclaredBorrowReturn|TestPhase25PointerPathChecksOverlapAcrossCalls|TestPhase25PointerPathChecksEscapeInCallFunction|TestPhase25UtilityOwnerTransfer|TestPhase25PreservesExactPhase24DirectOwnerRoute)$' -count=1 -v -timeout=600s

GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session ./internal/compiler/surfacemanifest -run '^(TestPhase36SurfaceManifestCallCensus|TestPhase36SurfaceManifestCallControls|TestPhase36SurfaceLimitConsumers|TestPhase36SurfaceDocument|TestPhase34SurfaceManifestWitnesses|TestPhase34SurfaceManifestCensus|TestPhase34SurfaceManifestControls|TestPhase34SurfaceLimitConsumers|TestPhase35SurfaceManifestWitnesses|TestPhase35SurfaceManifestCensus|TestPhase35SurfaceManifestControls|TestPhase35SurfaceLimitConsumers|TestPhase34SurfaceDocument|TestPhase34SurfaceLimits|TestPhase35SurfaceDocument)$' -count=1 -v -timeout=600s

GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase36NoCallGeneratedCByteBaseline|TestPhase36NoCallGeneratedCByteBaselineControls|TestPhase24TransferPeerValidatesAcquisitionDerivedDischarge|TestPhase24TransferPeerRejectsMutatedOwnership|TestPhase25IndependentPeer|TestPhase25IndependentPeerMutations|TestPhase25CheckCommandPeerGate|TestPhase36LegacyFallibleCallRoutes|TestPhase36LegacyFileOwnerPeerRoutes)$' -count=1 -v -timeout=600s

GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/pathoracle -run '^(TestPhase36RawCallShapeAndFreshness|TestPhase36RawCallRefusalRemovalControls|TestPhase36OriginCallFreshnessAndTransfer|TestPhase36OriginCallRefusalRemovalControls|TestPhase29ReleaseBeforeViewUseRefused|TestPhase29OwnerLoopBlockRefused|TestPhase31FileBytesSeededFaults)$' -count=1 -v -timeout=600s

python3 /tmp/schway36-borrow-mutation.py
```
