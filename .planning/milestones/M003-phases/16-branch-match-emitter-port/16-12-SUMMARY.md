---
phase: 16-branch-match-emitter-port
plan: "12"
subsystem: compiler-session-native-evidence
tags: [native, session, schema-2, m004, frozen-evidence]
requires: [16-11]
provides: [refusal-first-phase4-5-6-session-controls]
affects: [16-13]
tech-stack:
  added: []
  patterns: [refusal-first-control-routing, digest-bound-file-and-generated-artifacts, schema2-native-projection]
key-files:
  created: [testdata/phase16/file-frozen-evidence.json, testdata/phase16/generated-frozen-evidence.json]
  modified: [internal/compiler/session/session_phase5.go, internal/compiler/session/session_phase5_alias.go, internal/compiler/session/session_phase5_corpus_test.go, internal/compiler/session/session_phase6.go, internal/compiler/cgen/legacy_emitter_evidence_test.go, testdata/phase16/public-emitter-consumers.json]
decisions:
  - "M004 controls invoke public EmitNative first, then accept only a family-matching refusal and immutable digest-bound C evidence."
  - "Generated cut controls bind canonical enumerated Program bytes, generator identity, and artifact digest rather than using a hand-authored exception path."
  - "All native corpus comparisons project O0/O3 observations through schema-2 before comparison."
metrics:
  duration: "~2h"
  completed: 2026-09-20
  tasks: 2
  commits: 10
  plan_head_before: 8a8233e
actuals:
  tokens: 1333824
  tasks: 2
  commits: 10
status: complete
---

# Phase 16 Plan 12: Classified Session Native Evidence Summary

Phase 4/5/6 session controls now preserve their distinct native evidence while cut M004 inputs remain refusal-first, witness-bound, and digest checked.

## Completed Tasks

1. Routed Phase 5 corpus, alias, sanitizer, mismatch, LTO, and generated-closure controls through `Phase16ControlNativeC`. Admitted programs continue with live native emission and schema-2 comparison; foreign and by-pointer programs require the current public refusal before historical C is used.
2. Bound Phase 4 acquire-three and Phase 6 foreign/cleanup controls to frozen artifacts where cut, leaving supported paths dynamic. Added generated and file-backed manifests covering the enumerated closure and corpus controls, plus authoritative registry inventory enforcement.
3. Added fail-closed validation and mutation tests for file source and artifact digests, generated schema and generator identity, canonical Program bytes digest, artifact digest, and duplicate records.

## Verification

- `go test ./internal/compiler/cgen -run '^Test(LegacyEmitterEvidence|FileFrozenEvidenceRejectsFaults|GeneratedFrozenEvidenceRejectsFaults)$' -count=1` — passed.
- `go test ./internal/compiler/session -run '^TestPhase16PublicEmitterConsumerInventory$' -count=1` — passed.
- `go test ./internal/compiler/session -run 'TestPhase5|Test.*(Alias|Mismatch|Sanitize|Corpus|LTO)' -count=1` — passed.
- `go test ./internal/compiler/session -run 'Test.*(Phase4|Phase6|Injector|Foreign|Retained|Acquire|Cleanup)' -count=1` — passed.
- `git diff --check` — passed before the evidence-validator commit.

## Decisions Made

- The file-backed and generated frozen manifests are both authoritative validation inputs, rather than documentation-only inventories.
- Registry rows classify public API calls, while refusal rows name a machine-resolvable `probe:` witness; aliases no longer call the emitter directly.
- The validation-grade issue observed in an earlier broad run was not reproduced by either exact Plan 12 verification command after the migration.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing critical functionality] Added complete cut-fixture evidence coverage.**
- **Found during:** Task 2
- **Issue:** The initial Plan 16-11 ledger did not bind all enumerated and file-backed Phase 5 cut programs, so session controls could not preserve evidence after public M004 refusal.
- **Fix:** Added deterministic generated-source/program identity records and file-backed fixture records with immutable artifacts, current refusal checks, and digest checks.
- **Files modified:** `internal/compiler/session/session_phase5.go`, `testdata/phase16/generated-frozen-evidence.json`, `testdata/phase16/file-frozen-evidence.json`, and historical artifacts.
- **Commits:** `e2962b5`, `9ffb0f7`, `88b3a23`.

**2. [Rule 1 - Bug] Compared corpus native results under the same schema-2 projection as interpreter evidence.**
- **Found during:** Task 1
- **Issue:** The three-engine corpus assertion compared unprojected native output against schema-2 interpreter output.
- **Fix:** Projected both O0 and O3 native observations before `Phase5CompareEngines`.
- **Files modified:** `internal/compiler/session/session_phase5.go`.
- **Commit:** `1bf23b9`.

**3. [Rule 2 - Missing critical functionality] Made evidence manifests and the consumer inventory fail closed.**
- **Found during:** Task 2
- **Issue:** File-backed and generated bindings lacked one shared validator/mutation suite, and the registry still named removed direct alias calls.
- **Fix:** Validated every file/generated record, rejected altered schema/generator/source-program/artifact values and duplicates, and synchronized the AST-scanned public-emitter registry.
- **Files modified:** `internal/compiler/cgen/legacy_emitter_evidence_test.go`, `testdata/phase16/public-emitter-consumers.json`.
- **Commit:** `9e092a1`.

## Coordination Note

`STATE.md` and `.planning/config.json` were already modified by another active worker. They were intentionally preserved and not staged by this plan; the phase orchestrator owns their consolidated state update.

## Self-Check: PASSED

- All nine task commits and this execution-record commit from `8a8233e` exist.
- Both frozen manifests, their immutable artifacts, the session router, and the authoritative validator exist.
