---
phase: 32-clean-checkout-guide-and-final-receipts
plan: 03
subsystem: examples/frontier-retirement
tags: [dx-19, frontier, frozen-goldens, retired-allow-list]
requires: []
provides:
  - examples/checksum.schway retired (deleted) with tombstone sha256 guard
  - phase28RetiredEntries / phase30RetiredPaths dated allow-lists with armed controls
  - PROJECT.md 2026-10-12 Phase 32 DX-19 amendment and live-doc pointers
affects: [32-04, 32-11]
key-files:
  created:
    - internal/compiler/session/session_phase32_retire_test.go
  modified:
    - internal/compiler/session/session_phase28_manifest_test.go
    - internal/compiler/session/session_phase30_prechange_test.go
    - .planning/PROJECT.md
    - .planning/LANGUAGE-MATURITY.md
    - .planning/PRODUCT-ROADMAP.md
    - examples/checksum/README.md
    - .planning/EVIDENCE-RECONCILIATION.md
    - .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md
  deleted:
    - examples/checksum.schway
    - internal/compiler/session/session_phase20_test.go
key-decisions:
  - "Frozen goldens stay byte-identical; retired paths may be absent and are a failure if observed again"
  - "Archived 20-VALIDATION.md rows are reconciled in the PHASE-14 register (D-14-145/146), never rewritten"
status: complete
commits: 3
plan_head_before: 7ae2288659166a39a0da28cf8be3f8db9d70d5c5
plan_head_after: 22f405bca91369e96febc20883b5a976180baf68
actuals:
  tokens: 14000
  tasks: 3
  commits: 3
coverage:
  - id: D-05
    description: "examples/checksum.schway is deleted, the replacement exists, no renamed copy of the tombstone digest survives, and the armed control detects one"
    requirement: "DX-19"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase32_retire_test.go#TestPhase32ChecksumFrontierRetired"
        status: pass
    human_judgment: false
  - id: D-06
    description: "Both frozen goldens stay byte-identical; the retired path may be absent and is a failure if observed again, proven by armed controls"
    requirement: "DX-19"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase28_manifest_test.go#TestPhase28EntryExitClassManifestIsNotInert"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase30_prechange_test.go#TestPhase30PreChangeGoldenIsNotInert"
        status: pass
    human_judgment: false
  - id: D-07
    description: "Dated PROJECT.md amendment plus live pointers; the live-doc scan passes and its armed and dated-exempt controls fire"
    requirement: "DX-19"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase32_retire_test.go#TestPhase32ChecksumFrontierRetired"
        status: pass
    human_judgment: false
  - id: D-07a
    description: "The LANGUAGE-MATURITY.md corpus census reads 162 programs and 5,352 lines"
    requirement: "DX-19"
    verification:
      - kind: unit
        ref: "internal/compiler/session/self_describing_docs_test.go#TestLanguageMaturityCountsAreCurrent"
        status: pass
    human_judgment: false
duration: 35min
completed: 2026-10-12
---

# Phase 32 Plan 03: Frontier retirement Summary

`examples/checksum.schway` is deleted (tombstone sha256 `ae95e550...`), its pin is replaced by `TestPhase32ChecksumFrontierRetired`, both frozen goldens stay byte-frozen via dated retired-path allow-lists with armed controls, and one PROJECT.md amendment is the target of every live pointer.

## Accomplishments

- Task 1 (75d4d21b): deletion, `session_phase20_test.go` removed, retirement guard with `old-path-absent`, `replacement-exists`, `no-renamed-copy`, `armed-renamed-copy`; `phase28Compare` extracted; `phase28RetiredEntries`; `phase30RetiredPaths`; census set to 162 / 5,352.
- Task 2 (342eeb9b): new `TestPhase28EntryExitClassManifestIsNotInert` (`mutated-class`, `retired-observed-again`, `unretired-missing`) and a `retired-observed-again` subtest in the Phase 30 control.
- Task 3 (22f405bc): PROJECT.md amendment, LANGUAGE-MATURITY / PRODUCT-ROADMAP / README pointers, `phase32LiveDocViolations` with `live-doc-hygiene`, `armed-live-line`, `dated-line-exempt`.
- Golden JSON files, `milestones/` archived rows, `research/`, REQUIREMENTS.md, ROADMAP.md content and LANGUAGE-MATURITY.md lines 175-185 are unchanged (verified by diff against the plan base).

## Deviations from Plan

**1. [Rule 1 - Bug] Phase 30 non-inert control mutated the retired entry**
- **Found during:** Task 1
- **Issue:** `TestPhase30PreChangeGoldenIsNotInert` flipped a hash on `Entries[0]`, which is the now-retired path, so the flip was (correctly) not reported.
- **Fix:** the control picks the first non-retired entry.

**2. [Rule 3 - Blocking] Meta-tests flagged archived `20-VALIDATION.md:37-38` citations**
- **Found during:** Task 3 meta set (Reconciliation, Groundedness, DebtRegisters)
- **Issue:** the two archived rows cite the deleted `TestPhase20ChecksumFrontier`, which classifies R2 and must be reconciled.
- **Fix:** added `superseded` reconciliation entries D-14-145/146 (superseding phase P32, commit 75d4d21b, covering `TestPhase32ChecksumFrontierRetired`) to `PHASE-14-DEBT.md` (Items table rows, `items: 102`, detail sections) and the generated `EVIDENCE-RECONCILIATION.md` view (71 entries). The archived rows are untouched.
- **Commit:** 22f405bc

**3. [Rule 1 - Bug] Retiring the test broke the frozen hosted corpus receipt match (follow-up, orchestrator-directed)**
- **Found during:** post-plan full session run (`TestValidationRowGradesAreEarnedOverArchivedCorpus`)
- **Issue:** archived 20-VALIDATION rows cite `TestPhase20ChecksumFrontier`; once deleted it no longer resolved, so `requestedCorpusPairs` changed the session pair pattern and its digest no longer matched `testdata/phase16/validation-corpus-run-record.*` (revision 72a81740).
- **Fix (D-06, receipt stays byte-identical):** `corpusRetiredTests` (dated, reasoned) consulted only by `receiptTimeCorpusPairs`, which re-inserts retired names into the test index and re-runs the one derivation, so the manifest digest must equal the reconstructed set. `requestedCorpusPairs` and the producer seam still export current pairs. Rows citing a retired name are graded over the receipt-time index and require a `superseded` row in EVIDENCE-RECONCILIATION.md. Armed controls in `TestCorpusRetiredTestsAllowListIsNotInert`: stale (still resolves), unneeded (receipt already matches), unexplained drift (entry removed), unexplainable extra name; each observed failing when its guard was disabled.
- **Files modified:** `internal/compiler/session/evidence_grade_test.go`; `git diff 75d4d21b~1 -- testdata/phase16` is empty.
- **Commit:** 8c565e8b

## Deferred Issues

**`TestValidationRowGradesAreEarnedOverArchivedCorpus` fails: stale hosted corpus receipt (needs a decision / hosted run).**
The session pair pattern in `testdata/phase16/validation-corpus-run-record.jsonl` lists `TestPhase20ChecksumFrontier`; with that test gone, the requested pair pattern changes, so the pair digest in `validation-corpus-run-record.manifest.json` (`844eecec...`, current `543e045a...`) no longer matches. This test is not in the plan's meta set, and the research deletion experiment did not surface it. It passes at the plan base and is the only failure in `go test ./internal/compiler/session`; every other package passes. The receipt is a hosted-job artifact (see b308651e, ci.yml evidence step), and this plan may not dispatch CI, so I did not fabricate or locally regenerate it. It needs a refreshed hosted receipt (or an explicit choice to produce one locally) before the phase gate; suggest folding it into 32-07/32-11.

## Verification

- Retire, Phase 28/30 golden and control, LANGUAGE-MATURITY census and self-describing-docs tests pass.
- Plan session meta set (Inventory, Groundedness, Reconciliation, ValidationLifecycle, goldens, counts, exit-code docs, suppression, emitter inventory) passes.
- `go test` of all packages other than session passes.

## Known Stubs

None.

## Self-Check: PASSED

Created `session_phase32_retire_test.go` exists; commits 75d4d21b, 342eeb9b, 22f405bc are on the branch; the deleted paths are absent.
