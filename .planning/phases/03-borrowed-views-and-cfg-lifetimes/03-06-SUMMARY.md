---
phase: 03-borrowed-views-and-cfg-lifetimes
plan: "06"
subsystem: compiler-frontend-backend
tags: [ownership, borrowing, public-origins, separate-compilation, originvalidate, corevalidate, cli]

requires:
  - phase: 03-borrowed-views-and-cfg-lifetimes
    provides: "03-01's match-arm-body groundwork and core-graph facts (Block/Edge/LoanEndpoint) — this plan needed only the parser/check/core scaffolding 03-01 built, not the exclusive-loan/CFG-liveness work in 03-02..03-05, per the plan's re-pin rationale"
provides:
  - "A function's return type may carry a `borrow(path)`/`borrow mut(path)` annotation naming a declared origin path and access mode, lowered by check.go into a `core.PublicOrigin` fact"
  - "internal/compiler/originvalidate: a new package that independently recomputes a function's actual origin/access from the typed core alone (imports neither check nor ast), rejects understated origin sets and impossible access modes by recomputation, and checks a stale digest before any other question"
  - "core.Interface / core.FunctionSignature: a body-stripped, digest-bound separate-compilation summary with no Linear/Match field"
  - "Three new CLI subcommands (`interface export`, `interface core`, `interface check`) demonstrating separate compilation as real, separate process invocations — the consuming invocation never unmarshals the bound core artifact into anything that could carry a body"
  - "originvalidate.KnownEscape: the coordinated frontend-and-summary lie, named and surfaced next to corevalidate.KnownEscape, never claimed as solved"
  - "D-02-02 closed: evidence.ErrorCode(err) now reaches the CLI's evidence command instead of a hardcoded fallback string"
affects: [03-07-debug-lineage-and-phase-close]

actuals:
  tokens: 16788
  tasks: 3
  commits: 2

tech-stack:
  added: []
  patterns:
    - "Syntactic discriminant resolved before sameType: a return type's parsed TypeRef either carries a borrow-origin annotation (relaxed branch) or takes the unchanged identity path — resolved once, at parse time, never inferred from body shape"
    - "Independent recomputation by backward trace: originvalidate.RecomputeOrigin walks a TargetID->SourceID chain from OpReturn backward to the parameter place, a mechanism deliberately different from check.go's forward declaration-driven construction of the same fact (D-12)"
    - "Digest-first rejection: CheckSummary hashes raw, unparsed coreBytes and compares before ever touching origin content — body-blindness is structural (the digest check is the only operation performed on those bytes), not merely ordered first by discipline"

key-files:
  created:
    - internal/compiler/originvalidate/originvalidate.go
    - internal/compiler/originvalidate/originvalidate_test.go
    - internal/compiler/check/check_origin_test.go
    - testdata/phase3/public_view.lang
    - testdata/phase3/public_view_understated.lang
    - testdata/phase3/public_view_impossible.lang
  modified:
    - internal/compiler/ast/ast.go
    - internal/compiler/syntax/parser.go
    - internal/compiler/syntax/format.go
    - internal/compiler/syntax/syntax_test.go
    - internal/compiler/core/core.go
    - internal/compiler/check/check.go
    - internal/compiler/protocol/protocol.go
    - internal/compiler/session/session.go
    - internal/compiler/testsupport/cli_test.go
    - internal/compiler/evidence/evidence_test.go
    - cmd/lang/main.go

key-decisions:
  - "Origin paths are scoped to the function's own single parameter name — this reduced language has one parameter per function and no field-path-bearing executable shape (Box/Pair are rejected at check.go's execution-admission gate before a body is ever analyzed), so a richer field-path/per-alternative grammar would have no honest source input to exercise this phase. The underlying shape (Paths []string, plural) is kept additive for a future extension."
  - "A borrowed-view function's ability set is not separately re-derived with an explicit escape:false witness. corevalidate's existing type-fact loop recomputes abilities for every fact in linear.Types and requires an exact match against the declared set for the SAME shape; a synthetic 'view' TypeFact with a suppressed escape ability would desync from that independent recomputation. PublicOrigin therefore carries only Paths and Access — declarative metadata riding alongside the unmodified ability derivation, not a new ability fact."
  - "Dropped a first-draft check.go gate that rejected any straight-line function returning a borrow-derived value with no declared origin. 03-02's own `exclusive_borrow_clean` fixture (`check_exclusive_test.go`) legitimately returns a borrow-derived place with no annotation — this language has no separate View type, so a borrowed place types identically to an owned one, and this is pre-existing, already-shipped behavior outside this plan's scope to change. 'The origin annotation is mandatory' is instead enforced exactly where the plan's own text makes it testable: a `borrow` token in return-type position is grammatically required to carry `(path)` — omitting it is a parse-time, causal, span-bearing rejection (`syntax.expected_lparen`), not an inference."
  - "Separate compilation is demonstrated as three CLI invocations, not two: `interface export` (producer) and `interface check` (consumer) are OWN-04's required two-invocation demonstration; `interface core` is an additional standalone way to obtain the exact core-artifact bytes a summary's digest is bound to, standing in for what a real build pipeline already retains from `lang check`. Every byte `interface core` writes is verified identical to what `originvalidate.BuildInterface` digests, so the two-invocation claim (export, then check) is not weakened by the third command's existence."
  - "Match-bodied (S1) functions — bare-arm or arm-body/branch alike — are rejected uniformly for a declared borrow origin at one check point (before either match-handling branch runs), with a single named diagnostic (`ownership.match_borrowed_return_unsupported`) rather than two separately-coded gates, since both cases need the identical scope-limit rejection."

requirements-completed: [OWN-04]

coverage:
  - id: D1
    description: "A borrow(path)/borrow mut(path) return-type annotation parses losslessly, formats to a fixed point, and preserves token identity across reparse"
    requirement: OWN-04
    verification:
      - {kind: unit, ref: "internal/compiler/syntax#TestPublicViewRoundTrips", status: pass}
    human_judgment: false
  - id: D2
    description: "Checking a declared borrowed-view return produces a PublicOrigin fact naming the declared path and access mode; a path other than the function's own parameter is a causal rejection"
    requirement: OWN-04
    verification:
      - {kind: unit, ref: "internal/compiler/check#TestPublicOriginFactLowered", status: pass}
    human_judgment: false
  - id: D3
    description: "A borrow annotation with no origin path is a parse-level, causal, span-bearing rejection, not an inference"
    requirement: OWN-04
    verification:
      - {kind: unit, ref: "internal/compiler/check#TestBorrowedReturnWithoutOriginRejected", status: pass}
    human_judgment: false
  - id: D4
    description: "A match-bodied function (bare-arm or arm-body/branch) cannot declare a borrowed return origin this phase; rejected with a named, span-bearing cause, not the generic type.return_mismatch"
    requirement: OWN-04
    verification:
      - {kind: unit, ref: "internal/compiler/check#TestMatchBodyBorrowedReturnRejectedWithCause", status: pass}
    human_judgment: false
  - id: D5
    description: "originvalidate imports neither check nor ast, and independently recomputes origin/access from the typed core, rejecting understated origin sets and impossible access modes"
    requirement: OWN-04
    verification:
      - {kind: unit, ref: "internal/compiler/originvalidate#TestOriginValidateImportsNeitherCheckNorAst", status: pass}
      - {kind: unit, ref: "internal/compiler/originvalidate#TestOriginUnderstatedRejected", status: pass}
      - {kind: unit, ref: "internal/compiler/originvalidate#TestOriginAccessMismatchRejected", status: pass}
    human_judgment: false
  - id: D6
    description: "A stale summary (digest mismatch against the core artifact) is rejected before any origin or access question is even asked"
    requirement: OWN-04
    verification:
      - {kind: unit, ref: "internal/compiler/originvalidate#TestStaleSummaryRejectedBeforeOtherChecks", status: pass}
    human_judgment: false
  - id: D7
    description: "The coordinated frontend-and-summary lie is named and surfaced as an expected escape, never silently absent"
    requirement: OWN-04
    verification:
      - {kind: unit, ref: "internal/compiler/originvalidate#TestOriginEscapeIsNamed", status: pass}
    human_judgment: false
  - id: D8
    description: "The exported interface summary is body-stripped (no Linear/Match field, structurally and by content)"
    requirement: OWN-04
    verification:
      - {kind: unit, ref: "internal/compiler/originvalidate#TestInterfaceSummaryOmitsBodies", status: pass}
    human_judgment: false
  - id: D9
    description: "Separate compilation is two real, separate CLI process invocations (export, then check); the consuming invocation never touches a body field even when handed bytes that could not be decoded as a core artifact at all"
    requirement: OWN-04
    verification:
      - {kind: e2e, ref: "internal/compiler/testsupport#TestInterfaceCheckIsBodyBlindCLI", status: pass}
      - {kind: manual_procedural, ref: "shipped ./cmd/lang binary: interface export / interface core / interface check on a hand-written, non-corpus public-view program (module owned.handwritten_public_view), plus format --check / check / run --engine=interpreter / run --engine=native on the same program", status: pass}
    human_judgment: true
    rationale: "The shipped-binary drive (D-11) was performed manually this session against a hand-written program outside any corpus; it is verified session evidence, not captured as an automated regression test beyond the CLI test's own body-blindness proof."
  - id: D10
    description: "Every field this phase (and 03-01/03-02) introduced is absent, key by key as parsed JSON, from the serialized core of a Phase 1 program and a Phase 2 program"
    requirement: OWN-04
    verification:
      - {kind: unit, ref: "internal/compiler/evidence#TestPhase3FieldsAreOmittedWhenAbsent", status: pass}
      - {kind: unit, ref: "internal/compiler/evidence#TestPhase1EvidenceGoldenUnchanged", status: pass}
      - {kind: unit, ref: "internal/compiler/evidence#TestCanonicalEvidence", status: pass}
    human_judgment: false
  - id: D11
    description: "D-02-02 closed: evidence.ErrorCode(err) reaches the CLI's evidence command instead of a hardcoded fallback code"
    requirement: OWN-04
    verification:
      - {kind: unit, ref: "internal/compiler/evidence#TestEvidenceErrorCodeReachesCLI", status: pass}
    human_judgment: false

duration: 70min
completed: 2026-09-04
status: complete
---

# Phase 03 Plan 06: Public Borrowed-View Origins and Body-Blind Separate Compilation Summary

**A function's return type can now declare a `borrow(path)`/`borrow mut(path)` origin and access mode, verified against the typed core by a new independent `originvalidate` package that never trusts the declaration, and separate compilation is demonstrated as three real, separate `./cmd/lang` invocations whose consuming half never touches a body field.**

## Performance

- **Duration:** ~70 min
- **Tasks:** 3 completed
- **Files modified:** 17 (6 created, 11 modified)
- **Commits:** 2 (interleaved per-task work landed as 2 commits; see Task Commits)

## Accomplishments

- Added `ast.BorrowOrigin` and the `borrow(path)`/`borrow mut(path)`
  return-type annotation grammar, resolved as a syntactic discriminant
  before `sameType` ever runs (`syntax/parser.go`'s new `borrowOrigin()`),
  with the formatter taught the two new spacing rules a borrow-origin
  annotation's parens need (`format.go`).
- Added `core.PublicOrigin` (an additive, `omitempty` sibling of `Match`/
  `Linear` on `core.Function`, not a field on `TypeFact`) and
  `core.Interface`/`core.FunctionSignature` (the body-stripped
  separate-compilation summary type — no `Linear`/`Match` field exists on
  `FunctionSignature` at all, not merely omitted).
- `check.go` lowers a declared origin into `PublicOrigin{Paths, Access}`,
  rejects a path other than the function's own parameter with a causal
  `origin.unknown_path` diagnostic, and rejects any match-bodied function
  (bare-arm or branch) that declares an origin with a single named,
  span-bearing `ownership.match_borrowed_return_unsupported` diagnostic
  before either match-handling branch runs — never the generic
  `type.return_mismatch`.
- New package `internal/compiler/originvalidate` (verified by test to
  import neither `check` nor `ast`): `RecomputeOrigin` independently
  derives a function's actual origin path and access mode by tracing the
  `OpReturn` operation's source backward through the flat `Operations`
  list — a materially different mechanism from `check.go`'s forward,
  declaration-driven construction (D-12). `ValidatePublished` compares the
  recomputation against the declaration and rejects `core.origin_understated`
  or `core.origin_access_mismatch`, first-problem-only, matching
  `corevalidate`'s accumulation convention.
- `BuildInterface` strips every function body and binds the summary to a
  SHA-256 digest of the full core artifact (the same digest scheme
  `evidence.go`'s `CoreDigest` already uses). `CheckSummary` is the
  body-blind consumer: it hashes the raw, unparsed core-artifact bytes it
  is handed and compares against the summary's recorded digest — the
  digest comparison is the *only* operation ever performed on those bytes,
  so body-blindness is structural, not a matter of discipline. A digest
  mismatch is rejected (`origin.stale_summary`) before any origin or
  access content is even read.
- `originvalidate.KnownEscape` (`escape:coordinated-frontend-summary-lie`)
  names the accepted residual next to `corevalidate.KnownEscape`, surfaced
  via `ExpectedEscapes()`.
- Three new CLI subcommands: `interface export SRC OUT` (checks SRC,
  runs `ValidatePublished`, and only on success writes the body-stripped
  summary), `interface core SRC OUT` (writes the exact core-artifact bytes
  a summary's digest is bound to — a standalone way to obtain what a real
  pipeline already retains from `lang check`), and `interface check
  SUMMARY CORE` (the genuinely separate consuming invocation). `protocol.go`
  gained `InterfaceSummary`/`InterfaceFunctionAnswer` for the JSON/human
  projections, following the existing `EvidenceSummary` pattern.
- Closed D-02-02: `cmd/lang/main.go`'s `evidence` command now routes
  `evidence.ErrorCode(err)` instead of a hardcoded `evidence.operation_failed`
  string, so a distinct `ValidationError` code (e.g.
  `evidence.canonical_unstable`) reaches the CLI end to end.
- Added `TestPhase3FieldsAreOmittedWhenAbsent`: a key-by-key (parsed JSON,
  not struct-tag-assumed) falsifier that every field this phase and
  03-01/03-02 introduced — `public_origin`, `blocks`, `edges`,
  `loan_endpoints`, `block_id` — is completely absent from the serialized
  core of a Phase 1 and a Phase 2 program.
- Proved the slice through the shipped `./cmd/lang` binary (D-11):
  `format --check`, `check`, `run --engine=interpreter`,
  `run --engine=native`, and all three `interface` subcommands, on a
  hand-written, non-corpus public-view program
  (`module owned.handwritten_public_view`) — interpreter/native agreement
  held, and the full export→core→check round trip succeeded.

## Task Commits

1. **Task 03-06-01 + 03-06-02 (interleaved): end-to-end public borrowed view, independent origin verification** — `2b2649e` (feat)
2. **Task 03-06-03: golden byte identity + D-02-02** — `36478c5` (test)

**Plan metadata:** (this commit)

_Note: Tasks 1 and 2 landed in a single commit rather than split by task
boundary. `internal/compiler/originvalidate/originvalidate.go` is the one
package both tasks build incrementally — `RecomputeOrigin`,
`ValidatePublished`, `BuildInterface`, and `CheckSummary` were all needed for
Task 1's own end-to-end CLI demonstration to be meaningful (an `interface
export` that never validated anything would not demonstrate the producer-
verification contract Task 1's behavior list requires), so writing Task 1's
production code without Task 2's recomputation/rejection logic already
present would have required a throwaway intermediate implementation with no
functional benefit — the same interleaving precedent 03-02's summary
documents for its own single-file continuous edit. Task 2's own tests
(`TestOriginUnderstatedRejected`, `TestOriginAccessMismatchRejected`,
`TestStaleSummaryRejectedBeforeOtherChecks`, `TestOriginEscapeIsNamed`) and
fixtures (`public_view_understated.lang`, `public_view_impossible.lang`) are
committed in the same commit as the originvalidate.go they exercise. Task
3's D-02-02 fix and its own falsifying test are committed together in the
second commit, deliberately kept separate from Task 1/2's commit even though
both touch `cmd/lang/main.go`, so the fix and its test-level proof travel
together as one unit._

## Files Created/Modified

- `internal/compiler/ast/ast.go` — `ast.BorrowOrigin`, `FuncDecl.ReturnOrigin`
- `internal/compiler/syntax/parser.go` — `borrowOrigin()` grammar
- `internal/compiler/syntax/format.go` — two spacing rules for the origin annotation's parens
- `internal/compiler/core/core.go` — `core.PublicOrigin`, `core.Interface`, `core.FunctionSignature`, `core.InterfaceSchema`
- `internal/compiler/check/check.go` — origin lowering, `origin.unknown_path`, `ownership.match_borrowed_return_unsupported`
- `internal/compiler/originvalidate/originvalidate.go` — the new independent origin-verification package
- `internal/compiler/protocol/protocol.go` — `InterfaceSummary`/`InterfaceFunctionAnswer`
- `internal/compiler/session/session.go` — `InterfaceExportCommandFile`, `InterfaceCoreCommandFile`, `InterfaceCheckCommandFile`
- `cmd/lang/main.go` — `interface export|core|check` subcommands, D-02-02 fix
- `testdata/phase3/public_view.lang`, `public_view_understated.lang`, `public_view_impossible.lang` — fixtures

## Decisions Made

See `key-decisions` in frontmatter.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] An initial check.go gate broke 03-02's shipped `exclusive_borrow_clean` fixture**
- **Found during:** Task 03-06-01, first full-suite run after adding an implicit-borrowed-return rejection
- **Issue:** A first-draft gate rejected any straight-line function whose return derived from a borrow operation but declared no `borrow(path)` origin. `check_exclusive_test.go`'s `exclusive_borrow_clean` fixture (`let view = borrow mut buffer; let reviewed = borrow view; view`) — a pre-existing, already-shipped 03-02 test — legitimately returns a borrow-derived place with no such annotation, since this language has no separate View type and a borrowed place types identically to an owned one.
- **Fix:** Removed the gate entirely. "The origin annotation is mandatory" is instead enforced exactly where the plan's own text makes it testable: a `borrow` token in return-type position is grammatically required to carry `(path)` (`syntax.expected_lparen` otherwise) — a parse-time, causal, span-bearing rejection that never reaches check.go as a legal program, satisfying the plan's literal requirement without retrofitting new semantics onto 03-02's already-shipped straight-line return behavior.
- **Files modified:** `internal/compiler/check/check.go` (net: removed code, not added)
- **Verification:** Full `go test ./...` green, including `TestExclusiveBorrowLowersToCore`, `TestExclusiveBorrowAuthorizedIndependently`, `TestCorevalidateIndependentlyRejectsBorrowConflict`, `TestExclusiveBorrowInterpreterNative`
- **Committed in:** `2b2649e` (the gate was added and removed within the same working session, before any commit — no separate revert commit needed)

---

**Total deviations:** 1 auto-fixed (1 bug, caught and fixed before any commit). No scope creep — the fix narrowed check.go's new surface rather than adding to it.

## Issues Encountered

- The formatter's token-stream projector needed two small, narrowly-scoped
  rules to canonicalize the new `borrow(path)`/`borrow mut(path)` spelling:
  a space after the annotation's closing paren before the return-type
  identifier (unambiguous, since the parameter-list `)` is always followed
  by `->`, never an identifier), and no space between `borrow`/`mut` and
  the annotation's opening paren (unambiguous, since `(` only follows
  `borrow`/`mut` in this one grammar position). Both were caught immediately
  by `TestPublicViewRoundTrips`'s canonical-fixed-point assertion, not
  discovered downstream.
- The first `interface core` implementation appended a trailing newline to
  the written core-artifact file, which does not match what
  `originvalidate.BuildInterface` digests (`json.Marshal` with no trailing
  byte) — `interface check` correctly reported `origin.stale_summary` for a
  byte-identical-in-content-but-not-in-bytes core artifact, which is exactly
  the digest-first fail-closed behavior working as designed. Fixed by
  writing the core bytes with no appended newline, verified against a real
  CLI round trip (`interface export` → `interface core` → `interface
  check`, exit 0).

## User Setup Required

None — no external service configuration required.

## Mutation-Kill / Non-Regression Evidence (recorded verbatim per D-09)

- `git diff <phase-start>..HEAD -- testdata/phase1/` is empty.
- `git diff <phase-start>..HEAD -- testdata/phase2/` is empty (no owned_transfer golden or Phase 2 evidence golden moved).
- `TestPhase3FieldsAreOmittedWhenAbsent` independently confirms, key by key
  as parsed JSON (not assumed from struct tags), that `public_origin`,
  `blocks`, `edges`, `loan_endpoints`, and `block_id` are absent from a
  freshly-built Phase 1 and Phase 2 evidence core.
- `TestOriginValidateImportsNeitherCheckNorAst` reads `originvalidate`'s
  actual Go import list (via `go/parser`, not a doc comment) and fails if
  it ever imports `check` or `ast`.
- `TestOriginUnderstatedRejected`/`TestOriginAccessMismatchRejected` inject
  dishonesty by mutating an honestly-checked `core.Program`'s own
  `PublicOrigin` fact — check.go's honest producer can never construct
  either shape itself — and assert `originvalidate.ValidatePublished`
  independently catches both by recomputation, not by trusting the
  declaration.
- `TestStaleSummaryRejectedBeforeOtherChecks` asserts a digest mismatch is
  rejected even when the summary's own origin content is separately
  dishonest, proving the digest check is not merely first in sequence but
  the only check CheckSummary performs once it fails.
- `TestInterfaceCheckIsBodyBlindCLI` hands `interface check` a "core
  artifact" that is deliberately not valid `core.Program` JSON at all (not
  parseable into any struct carrying a `Linear`/`Match` field) and confirms
  the command still answers correctly once the digest matches — proving the
  consuming path never attempts anything beyond hashing those bytes.
- `env GOCACHE=/tmp/ai-lang-phase3-cache go test ./...`,
  `go test -race ./...`, and `go vet ./...` all pass with zero findings.
- `sh scripts/verify-phase2.sh` exits 0 and reports all Phase 1/2 controls
  unchanged through a freshly built binary including this plan's new
  package.
- Shipped-binary drive (D-11): `format --check`, `check`,
  `run --engine=interpreter`, `run --engine=native`, `interface export`,
  `interface core`, and `interface check` all pass on a hand-written,
  non-corpus program (`module owned.handwritten_public_view`, a shared
  borrowed-view return) — full interpreter/native agreement, and the
  export→core→check round trip succeeds end to end.

## Known Stubs

None. The origin path grammar supports only the function's own single
parameter name this phase (documented as a key decision, not a stub): the
underlying `Paths []string` shape is additive-ready for a future field-path
or per-alternative extension, but no executable shape in this language today
has fields to name, so there is no honest fixture that could exercise a
richer grammar.

## Next Phase Readiness

- OWN-04's public-origin declaration, independent recomputation, and
  body-blind separate-compilation demonstration are complete and
  independently verified.
- `originvalidate` is a clean, source-blind, body-blind package other
  future work can extend (e.g. a richer origin grammar) without touching
  `check` or `ast`.
- D-02-02 is closed; the remaining Phase 2 debt items not folded into this
  or a prior plan remain tracked in `02-DEBT.md` for whichever plan already
  touches their file.
- No blockers for 03-07.

---
*Phase: 03-borrowed-views-and-cfg-lifetimes*
*Plan: 06*
*Completed: 2026-09-04*

## Self-Check: PASSED
