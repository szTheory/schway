---
phase: 36-scalar-helper-calls
plan: "07"
subsystem: compiler-surface
tags: [scalar-calls, surface-manifest, source-census, tdd, go]

requires:
  - phase: "36-06"
    provides: ["checked local/imported scalar helper calls and composed Phase34/35 witness consumers"]
provides:
  - "Source-free schway.surface/1 scalar-call contract with bounded argument depth and exact U64/Bool signature pairs."
  - "Independent complete-program witnesses and a finite Phase36 source/limit census composed with immutable Phase34/35 fixtures."
  - "Live admission, refusal, hint, restriction, census, and limit-consumer controls across the composed manifest consumers."
affects: ["36-08", "Phase 37", surface-manifest, CALL-01, CALL-02]

actuals:
  tokens: 29028.25
  tasks: 2
  commits: 8
plan_head_before: 4a92f3f61d279bd10613f756c18e7582179f6873
plan_head_after: 5a0df76d618c18a45179b37c2e2214c1526a17ef
commits: 8

tech-stack:
  added: []
  patterns:
    - "Exact claim and branch-key supersessions replace only stale Phase34/35 facts while their JSON fixtures remain immutable."
    - "Complete-program witnesses and finite source census rows bind each published call claim to named producer branches."
    - "Copied-data controls remove the full active coverage for one facet, require the named finding, and verify restored positives."

key-files:
  created:
    - testdata/phase36/surface-census.json
    - testdata/phase36/surface-witnesses.json
    - internal/compiler/session/session_phase36_surface_test.go
    - .planning/phases/36-scalar-helper-calls/36-07-task1-red-evidence.json
    - .planning/phases/36-scalar-helper-calls/36-07-task2-red-evidence.json
  modified:
    - internal/compiler/surfacemanifest/manifest.go
    - internal/compiler/surfacemanifest/manifest_test.go
    - internal/compiler/session/session_phase34_surface_test.go
    - internal/compiler/session/session_phase35_surface_test.go
    - .planning/phases/36-scalar-helper-calls/deferred-items.md

key-decisions:
  - "Publish direct arity-one scalar calls only for U64/Bool parameter and result pairs, with imported targets explicitly exported and call graphs acyclic."
  - "Supersede only exact obsolete no-call claims and branch links; preserve owner/path full-session proofs and unrelated refusal boundaries."
patterns-established:
  - "Manifest claims are independently pinned by full source programs rather than generated expectations."
  - "Parser depth, runtime call depth, static preflight sizes, and byte capture budgets remain distinct limit classes."

# Plan36 is still in progress; CALL-01/CALL-02 remain phase-owned until Plan36-08 and phase gates complete.
requirements-completed: []
coverage:
  - id: D1
    description: "The source-free schway.surface/1 document publishes the bounded arity-one U64/Bool scalar-call contract and preserves unsupported boundaries."
    requirement: CALL-01
    verification:
      - kind: unit
        ref: "internal/compiler/surfacemanifest/manifest_test.go#TestPhase36SurfaceDocument"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase36_surface_test.go#TestPhase36SurfaceManifestCallWitnesses"
        status: pass
    human_judgment: false
  - id: D2
    description: "Independent complete programs, exact source census rows, and composed controls keep call claims, refusals, hints, restrictions, and limit routes live across Phase34/35/36."
    requirement: CALL-01
    verification:
      - kind: integration
        ref: "36-07-PLAN.md Task 2 exact focused selector command; TestPhase36SurfaceManifestCallCensus and TestPhase36SurfaceManifestCallControls"
        status: pass
      - kind: integration
        ref: "36-07-PLAN.md Task 2 exact focused selector command; Phase34/35 witness, census, controls, document, and limit roots"
        status: pass
    human_judgment: false

duration: 35min
completed: 2026-10-15
status: complete
---

# Phase 36 Plan 07: Scalar Call Surface Summary

**The source-free surface now describes bounded scalar helper calls, backed by independent complete programs and exact Phase34/35/36 census and control evidence.**

## Performance

- **Duration:** 35 minutes, measured from the first recorded RED run through final focused verification.
- **Started:** 2026-10-15T04:35:16Z
- **Completed:** 2026-10-15T05:10:27Z
- **Tasks:** 2/2 complete.
- **Files modified:** 9 plan source, test, fixture, and RED-evidence files, plus the deferred record and summary metadata.
- **Actual token estimate:** 29,028.25 (`116,113` diff characters divided by four), measured over the plan-owned source, test, fixture, and RED-evidence files.
- **Plan commit measurement:** 8 commits from `4a92f3f61d279bd10613f756c18e7582179f6873` through `5a0df76d618c18a45179b37c2e2214c1526a17ef`, measured before the final summary/STATE/ROADMAP metadata commit.

## Accomplishments

- Published direct arity-one local scalar calls for `U64 -> U64`, `U64 -> Bool`, `Bool -> U64`, and `Bool -> Bool`, with supported expression contexts, source-order argument lowering, the shared 64-level expression bound, exported-import requirements, and acyclic call graphs.
- Added independent complete-program witnesses for all four pairs, scalar binding/assignment, arithmetic and comparison operands, branch and loop conditions, byte indices, nested calls, imported calls, exact arity/type/private/cycle refusals, Bool identifier tails, and 64/65 call-argument depth.
- Composed a finite source census with exact parser/checker branch keys and distinct source, interpreter, static-preflight, and runtime-capture limits. Phase34/35 fixtures remain byte-identical.
- Kept missing/false admission and refusal, wrong hint, weakened restriction, omitted/added branch, limit-consumer, and restored-positive controls live across the full composed surface consumers.

## Exact Supersessions and Frozen Fixtures

Phase36 replaces only these inherited witness IDs:

| Facet | Prior witness IDs | Replacement witness IDs |
| --- | --- | --- |
| `type:bool` | `bool-local-state`, `bool-comparison-results` | `phase36-bool-local-state`, `phase36-bool-comparison-results`, `phase36-bool-function-parameter-and-result` |
| `restriction:scalar_expression_calls` | `restriction-scalar-expressions` | `phase36-pair-u64-u64` |
| `restriction:scalar_expression_grouping` | `phase35-grouping` | `phase36-grouping-with-helper-call` |
| `call:declaration` | `call-declaration-arity-one` | `phase36-call-declaration-arity-one` |
| `call:ordinary` | `call-linear-binding-arity-one`, `call-match-arm-arity-one`, `call-ordinary-arity-two` | `phase36-call-linear-binding-arity-one`, `phase36-call-match-arm-arity-one`, `phase36-call-ordinary-arity-two` |
| `call:scalar_expression` | `call-scalar-context` | `phase36-call-scalar-expression` |
| `probe:scalar-helper-call` | `probe-scalar-helper-call` | `phase36-probe-scalar-helper-call` |

The finite census uses exact branch-key replacements for the unchanged Bool derivation (`internal/compiler/ability/ability.go|deriveAt|shape.Constructor|Bool`), grouping admission (`internal/compiler/syntax/parser.go|scalarExprAtom|leftToken.Kind|TokenLParen`), the ordinary-call arity diagnostic and guard in `resolveCallBinding`, and the shared depth refusal (`internal/compiler/syntax/parser.go|nestedScalarExpr|if-condition|p.expressionDepth >= MaxExpressionDepth`). Each replacement keeps its original facet and links the Phase36 witness that now states the current claim.

The earlier frozen fixture digests remain:

| Fixture | SHA-256 |
| --- | --- |
| `testdata/phase34/surface-witnesses.json` | `823ac61fac34274ba6ee4cb8719ab0618443600fc976ca5587e1ffee76869742` |
| `testdata/phase34/surface-census.json` | `3934469136ed0f741dbbbca79428a98bc8134c7bfc0e5862ca440e85a94c256a` |
| `testdata/phase35/surface-witnesses.json` | `0fcfe36a942a5670a9819bdc789b4aebcf2182e336bbef9b9ef9d8a11fd2f64a` |
| `testdata/phase35/surface-census.json` | `9af1ecafc10e01bfcdca8882b44c85019f637c51790207f01808013e2461c98f` |

Limit contexts retain their distinct meanings: parser source-expression depth 64 levels, interpreter call depth 128 frames, CGen and execution-peer static preflight ceilings of 4096 nodes/occurrences, execution-document capture at 16 MiB, and application evidence/output capture at 65,536 bytes each. Capture budgets are not semantic iteration limits.

## TDD Gate Compliance

Both `tdd="true"` tasks have valid RED evidence preceding GREEN commits. The RED classifiers returned `RED_EVIDENCE_OK`; the named targets failed on their intended contract assertions. Neither task required a separate REFACTOR commit.

| Task | RED | GREEN | Evidence |
| --- | --- | --- | --- |
| 1: source-free call contract and complete-program witnesses | `a0f91760` | `4ee5f303` | `36-07-task1-red-evidence.json`; `TestPhase36SurfaceDocument` and `TestPhase36SurfaceManifestCallWitnesses` pass |
| 2: source census and inherited controls | `0ca6dbf3` | `5a0df76d` | `36-07-task2-red-evidence.json`; the exact 15-selector Task2 command passes |

## Task Commits

1. **Task 1 RED** — `a0f91760` (`test(36-07): add scalar call surface witnesses`)
2. **Task 1 GREEN** — `4ee5f303` (`feat(36-07): publish scalar call contract`)
3. **Task 2 RED** — `0ca6dbf3` (`test(36-07): add scalar call census controls`)
4. **Blocker record** — `c510ef7c` (`docs(36-07): record inherited pathoracle blocker`)
5. **Preserve prior deferred history** — `49eca44b` (`docs(36-07): preserve prior deferred record`)
6. **Scoped PathOracle compatibility repair** — `1c32b347` (`fix(pathoracle): preserve checked local borrow copy helpers`)
7. **GSD debug resolution record** — `67814db6` (`docs(debug): resolve phase36-borrow-call-paths session`)
8. **Task 2 GREEN** — `5a0df76d` (`feat(36-07): close scalar call census and controls`)

The final GSD metadata commit captures this summary, updated STATE/ROADMAP position, and the dated deferred-record resolution.

## Verification

- Task 1 exact verification passed both named roots:

  ```sh
  GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/surfacemanifest ./internal/compiler/session -run '^(TestPhase36SurfaceDocument|TestPhase36SurfaceManifestCallWitnesses)$' -count=1 -v -timeout=600s
  ```

- Task 2 exact verification passed all 15 named roots, including the Phase34/35 full-session witnesses and controls, the Phase36 census and controls, and all limit consumers:

  ```sh
  GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session ./internal/compiler/surfacemanifest -run '^(TestPhase36SurfaceManifestCallCensus|TestPhase36SurfaceManifestCallControls|TestPhase36SurfaceLimitConsumers|TestPhase36SurfaceDocument|TestPhase34SurfaceManifestWitnesses|TestPhase34SurfaceManifestCensus|TestPhase34SurfaceManifestControls|TestPhase34SurfaceLimitConsumers|TestPhase35SurfaceManifestWitnesses|TestPhase35SurfaceManifestCensus|TestPhase35SurfaceManifestControls|TestPhase35SurfaceLimitConsumers|TestPhase34SurfaceDocument|TestPhase34SurfaceLimits|TestPhase35SurfaceDocument)$' -count=1 -v -timeout=600s
  ```

- `python3 -m json.tool` accepted both Phase36 fixture roots. A focused placeholder/TODO/FIXME scan found no stub markers in plan source, test, or fixture files.
- The original RED classifiers returned `RED_EVIDENCE_OK`; the exact current Task1 and Task2 commands passed after the compatibility repair.
- No full repository suite, broad native matrix, cross-host, hosted, or human UAT result is claimed by this plan.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking issue] Preserve retained Phase24 ownership-transfer witnesses through scalar calls**
- **Found during:** Task 2
- **Issue:** The exact composed consumer suite showed `core.pathoracle_refused` for eight immutable Phase34/35 full-session witnesses whose Phase24 program calls `shared_copy(value)` and `exclusive_copy(shared)`.
- **Fix:** The parent routed the production compatibility defect through the `phase36-borrow-call-paths` GSD debug workflow. The narrow PathOracle gate admits only the independently checked U64 borrow → copy → owned-return helper shape, checks clean authority fields, and replays the straight-line loan path; existing owner/path, provenance, CFG/backedge, and borrowed-result gates remain. See the [resolved debug record](../../debug/resolved/phase36-borrow-call-paths.md).
- **Files modified:** `internal/compiler/pathoracle/pathoracle.go`, `internal/compiler/pathoracle/pathoracle_phase36_borrow_test.go` (parent-owned debug commit `1c32b347`).
- **Verification:** Fresh exact Task2 selectors pass; the debug record also preserves the named raw-authority, revert, no-op, and restored-positive controls.

**2. [Rule 3 - Blocking test-control issue] Keep the composed depth-refusal mutation live**
- **Found during:** Task 2
- **Issue:** Removing only `phase35-depth-65` left the new Phase36 depth-65 witness active, so the composed test no longer reached `missing_refusal`.
- **Fix:** The copied-data control now removes every active witness for that exact refusal facet, guards against removing zero rows, asserts `missing_refusal`, and verifies the original witnesses restore a passing result. Frozen Phase35 JSON is unchanged.
- **Files modified:** `internal/compiler/session/session_phase35_surface_test.go`.
- **Verification:** `TestPhase35SurfaceManifestControls` passes as part of the exact Task2 selector command.

**Total deviations:** 2 blocking compatibility/control corrections. No new schema, dependency, backend, public route, or Phase37 scope was added.

## Issues Encountered

The PathOracle refusal and its eight legacy witness IDs remain recorded with their dated resolution in [deferred-items.md](./deferred-items.md). That record also retains the earlier resolved Plan36-05 history. The fresh exact task selectors passed after the narrow correction; the Plan36 phase gates remain owned by the parent.

## User Setup Required

None - no external service or user configuration is required.

## Next Phase Readiness

Plan36-07 is complete and Plan36-08 is next. Phase36 remains in progress; CALL-01/CALL-02 and phase-level success criteria remain open until their remaining plan and phase gates complete. No Phase37 work was started.

## Self-Check: PASSED

- The SUMMARY file exists at the canonical plan path.
- All eight pre-metadata plan-range commits are reachable from current HEAD.
- Coverage metadata classifies both deliverables as automated and passing.

---
*Phase: 36-scalar-helper-calls*
*Status: plan complete; Phase36 remains in progress*
