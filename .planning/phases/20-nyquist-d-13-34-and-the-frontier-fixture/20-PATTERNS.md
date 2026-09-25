# Phase 20: Nyquist, D-13-34, and the Frontier Fixture — Pattern Map

**Mapped:** 2026-09-24  
**Files analyzed:** 9 likely implementation/documentation targets  
**Analogs found:** 9 / 9 (some decisions are deliberately human-owned)

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `internal/compiler/session/verification_groundedness_test.go` | test / evidence instrument | batch transform over planning docs and Go test index | same file's `pinnedFrontier`, `r2bLandingPhases`, and frontier tests | exact |
| `internal/compiler/session/session_phase20_test.go` (or existing focused session test) | test | request-response checker probe | `session_phase19_test.go` | role-match |
| `examples/checksum.lang` | source fixture | file I/O / loop transform | `testdata/phase19/literal_tracer.lang`; refused controls in `testdata/phase19/` | role-match |
| `internal/compiler/session/session_phase5_corpus_test.go` | test | batch transform / differential execution | `TestPhase5Corpus...` and `enumerated-closure` | exact |
| `internal/compiler/cache/cache.go` and `internal/compiler/cache/cache_test.go` (only if cache API must be extended) | service / test | file I/O, content-addressed artifact reuse | existing cache key/store tests | exact |
| `internal/compiler/session/session_test.go` | test / debt gate | batch transform over debt registers | `TestDebtRegistersAreWellFormed` and `debtRegisterProblems` | exact |
| `.planning/**/*-VALIDATION.md` and possibly lifecycle test in `session_test.go` | documentation / test | batch reconciliation | `.planning/phases/14-evidence-instrument-and-honest-scoping/14-VALIDATION.md` and Phase 14 lifecycle gates | role-match |
| `testdata/phase6/{heldout,derivation}_{move,borrow}_defect.lang`, `PHASE-13-DEBT.md`, and/or Phase 20 decision record | fixture / debt record | batch evidence transform | Phase 6 corpus, D-13-34 probe, and `PHASE-13-DEBT.md` | exact for mechanics; no analog can decide ratification |
| `.planning/phases/20-*/20-VALIDATION.md` / `20-VERIFICATION.md` | documentation / metadata | request-response evidence | Phase 14 and Phase 19 validation/verification maps | role-match |

The task scope does not imply every listed candidate must change. In particular, a D-13-34 fixture rewrite is conditional on the durable outcome selected at the required human checkpoint.

## Pattern Assignments

### Evidence frontier and validation-command reconciliation

**Analog:** `internal/compiler/session/verification_groundedness_test.go` (tracked).

The file is 2,221 lines; the excerpts below are non-overlapping targeted ranges. Imports use only Go standard library plus the project test-support package (lines 20–43). The frontier is a literal set of full finding records, not a count or ceiling (lines 1137–1155). `TestVerificationGroundednessFrontierIsPinned` measures current documents, excludes findings only through the explicit reconciliation mechanism, and asserts set equality in both directions (lines 1326–1367). `r2bLandingPhases` ties each exact per-branch record to a closed owner (lines 2115–2146); `TestVerificationGroundednessThreeClassesAreEmpty` keeps R1/R2/R3 distinct, requires every residual R2b owner, and asserts document/command corpus floors (lines 2155–2221).

**Planner direction:** capture the starting scan before editing docs. Reconcile against measured findings and update exact line/command records after document edits. Keep raw-discovery and post-reconciliation counts distinct. The Phase 20 ROADMAP's “zero draft” criterion covers all scoped validation files; do not infer completion from only phases 07/08/11 and 12/13.

### Refused checksum fixture and moved diagnostic

**Analog:** `internal/compiler/session/session_phase19_test.go` (tracked), `TestPhase19NumericRefusalFrontiers` (lines 137–170).

The test table associates each fixture with a witness literal, diagnostic code, and source offset, reads through `testsupport.ProjectPath`, confirms the witness remains present, runs the real `session.Check`, and pins the observed code and span. The nearby accepted-frontier test (`TestPhase19LiteralFrontier`, lines 126–135) also checks semantic reachability rather than mere file presence.

**Fixture analog:** `testdata/phase19/literal_overflow.lang` and `literal_malformed.lang` (tracked). Reuse the repository's compact `.lang` source style and give the checksum fixture a stable, testable intent witness. The new test should compare structured diagnostic identity/location with a durable M003-open pin, and assert inequality as the roadmap requires. Determine both values from actual checker/history evidence; do not guess codes or offsets. No current `examples/checksum.lang` exists.

### Closure evidence reuse

**Analog:** `internal/compiler/session/session_phase5_corpus_test.go` (tracked), especially the `enumerated-closure` subtest (lines 553–600); `internal/compiler/cache/cache.go` and `cache_test.go` (tracked).

The Phase 5 test enumerates the closure, rejects an empty corpus, then still runs each program through the three-engine comparison. Cache `ComputeKey` sorts named declared inputs, rejects duplicate names and empty digests, and hashes canonical JSON (`cache.go:52–86`). `TestCacheRoundTripsArtifactAndMeta` exercises cold absence, put/get, and changed-key absence (`cache_test.go:23–49`); `TestCacheEqualDeclaredInputsShareOneEntry` checks canonical input ordering and shared entries (`cache_test.go:346–394`). The cache package documents and structurally tests its key rule: cache artifacts only, never pass/fail verdicts (`cache.go:12–23`; `cache_test.go:120–179`).

**Planner direction:** key reuse on a complete named input set for the deterministic closure/evidence artifact, exercise same-input reuse and seeded-input invalidation, and retain fresh comparator/assertion execution on every invocation. Capture a comparable suite-wall-clock measurement against the recorded Phase 14 baseline. Avoid adding a second cache mechanism or reusing a cached green verdict.

### Unowned debt cap

**Analog:** `internal/compiler/session/session_test.go` (tracked), `TestDebtRegistersAreWellFormed` (lines 3039–3071), `checkDebtRegister` and `debtRegisterProblems` (around lines 3189–3235).

The existing gate discovers every debt register, checks each in a named subtest, and delegates to a deterministic pure function that reports all problems rather than stopping at the first. Its documented current obligations include frontmatter/table agreement, ownership vocabulary, and detail/table bijection. Extend the same authoritative parse/model to count the explicitly agreed PRC-02 population; put a seeded over-cap control beside the gate so the numerical assertion is not inert. The research notes the denominator is unresolved (8 current generated claims versus 10 M002 carry-forward versus 40 historical cells); the implementation must derive the count from a documented authoritative population rather than selecting whichever number passes.

### Validation metadata and evidence docs

**Analogs:** `.planning/phases/14-evidence-instrument-and-honest-scoping/14-VALIDATION.md` (tracked), `.planning/phases/19-numeric-literals-and-opconst/19-VALIDATION.md` (tracked), and the existing `TestVerificationGroundedness...` gates above.

Phase 14's validation map shows the project's row structure: requirement/task, validation command, existence, status, evidence grade, and notes. Phase 19's map demonstrates current package/test naming. For draft reconciliation, use Phase 14's executed lint to locate stale, elided, or zero-test commands; update commands from the actual `go test -list` surface and only promote lifecycle status once evidence is earned. The Phase 20 zero-draft criterion is broader than its named examples: refresh the complete `.planning/**/*-VALIDATION.md` status inventory before finalizing. Validation and verification metadata should state the measured commands/results, not claim that a document status alone proves execution.

### D-13-34 probe and ratification

**Analog:** `internal/compiler/session/session_phase6_injectors_test.go` (tracked), `TestPhase6HeldoutPairsAreAlphaRenamesOnly` (lines 200–220), the `phase6StructuralSummary` predicate (lines 221–258), and `TestPhase6DefectCorpusDistinctnessGuardIsNotInert` (line 324 onward). Historical decision record: `.planning/milestones/M002-phases/13-agent-loop-for-interprocedural-defects/13-VALIDATION.md` (tracked), with Phase 6 debt context in `.planning/milestones/M001-phases/06-agent-feedback-and-performance-ratification/06-DEBT.md` (tracked).

The probe compares identifier-independent summaries for held-out and derivation move/borrow pairs and fails with explicit instructions when the weakness closes; the summary counts structural facts rather than source bytes. The prior Phase 13 record says the actual checkpoint ratified keeping fixtures unchanged and carrying D-13-34 as debt. Phase 20's roadmap explicitly re-triggers that decision now that the surface changed.

**Planner direction:** keep the existing structural predicate and its anti-overfitting rationale. At the required blocking human checkpoint, select either genuinely distinct replacement fixtures or a reasoned re-ratification with an owning milestone. Do not silently alter fixture history, weaken the predicate, or treat prior Option B as automatic permission to repeat it. Make D-06-29's implication for the two source classes explicit as restored evidence or a written-off claim.

## Shared Patterns

### Project paths and source checks

**Source:** `internal/compiler/session/session_phase19_test.go:14–18,137–170`.

Use `testsupport.ProjectPath(...)` for repository-rooted fixture/doc reads; check read errors immediately; assert a semantic witness before trusting a fixture; run the real production checker and pin stable structured outcomes.

### Evidence records stay exact and non-vacuous

**Source:** `internal/compiler/session/verification_groundedness_test.go:1326–1367,2155–2221`.

Use set equality for exact frontiers, preserve classification distinctions, give every residual finding an explicit owner, and assert corpus floors in the same run as an empty-class claim.

### Cache reuse never replaces assertions

**Source:** `internal/compiler/cache/cache.go:12–23,52–86`; `internal/compiler/session/session_phase5_corpus_test.go:553–600`.

Declared input digests determine reusable content. A cache hit may skip expensive generation, but semantic/differential assertions still run fresh.

### Debt is explicit; honesty remains reviewable

**Source:** `internal/compiler/session/session_test.go:3039–3071`; Phase 6 debt register `06-DEBT.md` frontmatter and Items table (lines 1–45).

Keep machine-checkable shape/count rules separate from judgment about whether the deferral is honest. Record first-seen reason, severity, source, and landing owner in the established register form.

## No Analog Found

No close existing artifact exists for the specific source program `examples/checksum.lang` or its M003-open checksum diagnostic pin. Use the Phase 19 refusal-test mechanics but establish the baseline from authoritative history or an explicit durable source record before claiming movement.

## Metadata

**Analog search scope:** `internal/compiler/session/`, `internal/compiler/cache/`, `testdata/phase6/`, `testdata/phase19/`, Phase 14/19 planning artifacts, and M001/M002 debt and validation records.  
**Tracked-source gate:** every named source/document analog was checked with `git ls-files`; all paths cited above are tracked.  
**Pattern extraction date:** 2026-09-24.
