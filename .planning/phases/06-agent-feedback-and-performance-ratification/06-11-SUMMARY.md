---
phase: 06-agent-feedback-and-performance-ratification
plan: 11
subsystem: compiler-diagnostics
tags: [diagnostic, repair, rustfix, applicability, identity, D-06-24]

requires:
  - phase: 06-01
    provides: "lang.diagnostic/1 schema, ErrorWithRepairs identity discipline, and the identity-enumeration guard pattern (TestMetricsAndLaneFieldsExcludedFromIdentity) this plan mirrors for Repair"
provides:
  - "diagnostic.Repair grown into an applicable edit: optional Span, Replacement, and a closed Applicability enum, all additive on lang.diagnostic/1"
  - "diagnostic.DriverEligible(r) — the single narrow predicate 06-13's repair driver must call before applying anything"
  - "Executable, self-invalidating proof that span/replacement/applicability can never move a published diagnostic ID"
affects: ["06-12", "06-13", "06-14", "06-15"]

actuals:
  tokens: 9500
  tasks: 3
  commits: 4

tech-stack:
  added: []
  patterns:
    - "Non-identity-bearing field extension: new struct fields added to an existing identity-hashed type, explicitly excluded from the identity struct with a comment at the exclusion site, and pinned by a reflection-based field-enumeration guard so a future field cannot silently join or skip the identity boundary."
    - "Fail-closed eligibility predicate: DriverEligible requires BOTH a declared enum state and the material (span+replacement) to act on — a claim without evidence is not eligible."

key-files:
  created:
    - internal/compiler/diagnostic/diagnostic_test.go
  modified:
    - internal/compiler/diagnostic/diagnostic.go
    - internal/compiler/check/check.go

key-decisions:
  - "D-06-24 confirmed at orchestration time (Task 0, non-blocking checkpoint, resolution recorded in PLAN.md): Repair gains optional Span, Replacement, and a rustfix-style Applicability enum (MachineApplicable | RequiresConfirmation | Unspecified), ALL in the non-identity-bearing bucket; only Kind stays identity-bearing; sort key unchanged; lang.diagnostic/0 bytes stay frozen."
  - "Chose check.go's insert_take repair (ownership.transfer_requires_take, straight-line body path) as the one real call site wired end-to-end: Span = binding.RHS.Span, Replacement = \"take \" + binding.RHS.Source, Applicability = MachineApplicable — a genuinely mechanical fix (insert the take keyword before the moved identifier)."
  - "NormalizeApplicability is a separate function, not automatic on construction — a zero-value Repair must marshal with none of the three new keys present, which requires Applicability to stay empty rather than auto-populating to Unspecified at construction time."

patterns-established:
  - "Reflection field-enumeration guard on Repair (TestRepairFieldEnumerationIsExhaustive), mirroring 06-01's identity-enumeration guard on Result — the mechanism for keeping an identity/non-identity split honest as a type grows."

requirements-completed: []  # DX-04 spans 06-11 through 06-15; not marked complete here — see plan's success_criteria note

coverage:
  - id: D1
    description: "diagnostic.Repair carries optional Span, Replacement, and a closed Applicability enum, all non-identity-bearing"
    requirement: "DX-04"
    verification:
      - kind: unit
        ref: "internal/compiler/diagnostic/diagnostic_test.go#TestRepairCarriesSpanReplacementApplicability"
        status: pass
      - kind: unit
        ref: "internal/compiler/diagnostic/diagnostic_test.go#TestApplicabilityVocabularyIsClosed"
        status: pass
      - kind: unit
        ref: "internal/compiler/diagnostic/diagnostic_test.go#TestEmptyApplicabilityDefaultsToUnspecified"
        status: pass
      - kind: other
        ref: "lang --json check testdata/phase2/implicit_noncopy.lang (shipped binary, manual run)"
        status: pass
    human_judgment: false
  - id: D2
    description: "The identity split holds: span/replacement/applicability never move a diagnostic ID; kind does; the field set is exhaustively enumerated"
    requirement: "DX-04"
    verification:
      - kind: unit
        ref: "internal/compiler/diagnostic/diagnostic_test.go#TestRepairIdentityUsesKindOnly"
        status: pass
      - kind: unit
        ref: "internal/compiler/diagnostic/diagnostic_test.go#TestRepairSpanShiftDoesNotChangeDiagnosticID"
        status: pass
      - kind: unit
        ref: "internal/compiler/diagnostic/diagnostic_test.go#TestRepairFieldEnumerationIsExhaustive"
        status: pass
      - kind: other
        ref: "manual mutation-kill: folding Span into ErrorWithRepairs's identity struct flips TestRepairSpanShiftDoesNotChangeDiagnosticID; adding an unused Repair field flips TestRepairFieldEnumerationIsExhaustive"
        status: pass
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestPreviousPhaseCoreBytesUnchanged"
        status: pass
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestPreviousPhaseManifestIDsUnchanged"
        status: pass
    human_judgment: false
  - id: D3
    description: "Only a complete MachineApplicable repair is driver-eligible; lang.diagnostic/0 stays frozen and repair-free"
    requirement: "DX-04"
    verification:
      - kind: unit
        ref: "internal/compiler/diagnostic/diagnostic_test.go#TestOnlyMachineApplicableIsDriverEligible"
        status: pass
      - kind: unit
        ref: "internal/compiler/diagnostic/diagnostic_test.go#TestMachineApplicableWithoutMaterialIsNotEligible"
        status: pass
      - kind: unit
        ref: "internal/compiler/diagnostic/diagnostic_test.go#TestDiagnosticZeroBytesUnchanged"
        status: pass
    human_judgment: false

duration: 18min
completed: 2026-09-07
status: complete
---

# Phase 06 Plan 11: Repair Grows Into an Applicable Edit Summary

**`diagnostic.Repair` gains an optional Span, Replacement, and a closed rustfix-style Applicability enum, wired end-to-end through a real ownership repair, with an executable proof that the new fields can never move a published diagnostic ID.**

## Performance

- **Duration:** 18 min
- **Started:** 2026-09-07T00:00:16Z
- **Completed:** 2026-09-07T00:18:28Z
- **Tasks:** 3 (plus Task 0, a non-blocking checkpoint resolved at orchestration time)
- **Files modified:** 3 (2 modified, 1 created)

## Accomplishments

- Extended `diagnostic.Repair` with `Span *Span`, `Replacement string`, and `Applicability string`, all `omitempty` and all excluded from `ErrorWithRepairs`'s identity struct and `repairKinds` derivation.
- Declared the closed `Applicability` vocabulary as exported constants (`ApplicabilityMachineApplicable`, `ApplicabilityRequiresConfirmation`, `ApplicabilityUnspecified`) plus `Applicabilities()`, `ValidateRepair()`, and `NormalizeApplicability()` (empty never defaults to MachineApplicable).
- Wired the `insert_take` repair in `check.go`'s straight-line ownership analysis (`ownership.transfer_requires_take`) with real `Span`/`Replacement`/`Applicability: MachineApplicable` values, and proved through the shipped `lang` binary that `--json check` on `testdata/phase2/implicit_noncopy.lang` emits the three new `repairs[]` keys while the diagnostic ID stays byte-identical to its pre-extension value (`diagnostic:040662ef397be67eebb23003`).
- Added `TestRepairIdentityUsesKindOnly`, `TestRepairSpanShiftDoesNotChangeDiagnosticID`, and `TestRepairFieldEnumerationIsExhaustive` — mutation-killed manually (folding `Span` into the identity struct flips the span test; an unused field flips the enumeration test).
- Added `DriverEligible(r Repair) bool`, requiring `MachineApplicable` AND a non-nil `Span` AND non-empty `Replacement`; `TestOnlyMachineApplicableIsDriverEligible` and `TestMachineApplicableWithoutMaterialIsNotEligible` cover the full table including incomplete-`MachineApplicable` rows.
- Added `TestDiagnosticZeroBytesUnchanged`, pinning a literal `/0` diagnostic ID (`diagnostic:d34396fcc8ea6530a1f04dc3`) and asserting `Error()`'s `Repairs` field stays nil — the `/0` path is untouched.

## Task Commits

Each task was committed atomically, following true RED/GREEN structure where the underlying behavior was genuinely new (Task 3):

1. **Task 1: Extend Repair additively, end-to-end through a real diagnostic** — `252de56` (feat)
2. **Task 2: Prove the identity split — a coordinate shift cannot move an ID** — `973f36d` (test; pre-existing invariant re-asserted, no implementation change needed)
3. **Task 3: Only MachineApplicable is driver-eligible, and /0 stays frozen** — `013c461` (test, RED — `DriverEligible` undefined, build fails as expected) then `a822b16` (feat, GREEN — `DriverEligible` added, full suite green)

**Plan metadata:** committed alongside this SUMMARY.

_Note: Task 0 (`checkpoint:decision`) carried `<blocking>false</blocking>` with a recorded `<resolution>` confirming D-06-24 at orchestration time on 2026-09-06 — the executor did not halt, per the plan's explicit instruction._

## Files Created/Modified

- `internal/compiler/diagnostic/diagnostic.go` — `Repair` struct extended with `Span`/`Replacement`/`Applicability`; new constants, `Applicabilities()`, `ValidateRepair()`, `NormalizeApplicability()`, `DriverEligible()`; a comment at the `repairKinds` loop documenting the exclusion.
- `internal/compiler/diagnostic/diagnostic_test.go` — new file; 9 tests covering additive extension, JSON omission, the identity split (with mutation-kill verification performed manually), and driver eligibility / `/0` freeze.
- `internal/compiler/check/check.go` — one `ErrorWithRepairs` call site (the `insert_take` repair in `analyzeStraightLine`'s `default:` binding-kind arm) now populates `Span`, `Replacement`, and `Applicability: MachineApplicable` from facts the checker already holds (`binding.RHS.Span`, `binding.RHS.Source`).

## Decisions Made

- **D-06-24 confirmed as locked** (Task 0, resolved at orchestration time, non-blocking): `Repair` gains optional `Span`, `Replacement`, and a rustfix-style `Applicability` enum, all in the non-identity-bearing bucket; only `Kind` stays identity-bearing; the canonical sort key (`Kind`, then `Detail`) is unchanged; `lang.diagnostic/0` bytes remain frozen.
- Picked `insert_take` (not `use_transfer_target` or `move_after_last_borrow_use`) as the one wired repair — it maps to a genuinely mechanical source edit (`take <ident>`), unlike repairs whose "target" is an internal place ID rather than source-level text.
- `NormalizeApplicability` kept as an explicit function rather than auto-populating `Applicability` at construction, so a zero-value `Repair` marshals with none of the three new JSON keys present (required by acceptance criteria).

## Deviations from Plan

None — plan executed exactly as written, including the non-blocking Task 0 checkpoint resolution.

## Issues Encountered

None.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- `diagnostic.DriverEligible` is the single narrow predicate 06-13's repair driver will call; only a complete `MachineApplicable` repair passes.
- The identity split is now self-enforcing: any future field added to `Repair` without updating `TestRepairFieldEnumerationIsExhaustive` fails loudly, naming the field.
- `lang.diagnostic/0` remains provably frozen and repair-free (`TestDiagnosticZeroBytesUnchanged`).
- DX-04 is NOT marked complete — 06-12, 06-13, 06-14, and 06-15 are the remaining plans of that requirement.
- Full suite (`go test ./...`), `go test -race ./...`, and `go vet ./...` are all clean at HEAD.

---
*Phase: 06-agent-feedback-and-performance-ratification*
*Completed: 2026-09-07*
