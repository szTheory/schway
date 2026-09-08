---
phase: 07-calls-signatures-and-call-graph-refusal
plan: 01
subsystem: compiler-core
tags: [interface-schema, decoder, digest, ownership-summary, go]

requires:
  - phase: 03-borrowed-views-and-cfg-lifetimes
    provides: core.PublicOrigin, originvalidate.RecomputeOrigin/ValidatePublished, the /0 Interface artifact this plan versions
  - phase: 04-fallible-resources-and-c-boundary
    provides: core.ForeignContract (the ForeignReach/Fails vocabulary precedent)
provides:
  - "lang.interface/1 schema: core.FunctionSignature/ParameterContract/ReturnContract/ForeignReach"
  - "core.DecodeInterface: schema-peek dispatch, strict /1 validation, pinned /0 legacy decode"
  - "core.InterfaceV0/FunctionSignatureV0: frozen, structurally non-admissible /0 shape"
  - "canonical, non-self-referential ClosureDigest preimage (zero-callee base case)"
affects: [07-02-callable-and-caller-admission, 07-03, 07-08-closure-digest-chaining]

actuals:
  tokens: 15520
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "schema-peek dispatch: unmarshal only {schema} first, branch to the pinned legacy decoder or strict current-version validation"
    - "unexported same-package fault-injection seam (decoderRequireNonEmpty, closureDigestSortOverride) for mutation-kill tests, never an exported package-level var on a production path (D-07-42)"
    - "raw-JSON key-presence check via map[string]json.RawMessage where a struct field's Go zero value is itself a legal value (ForeignReach)"

key-files:
  created:
    - internal/compiler/core/core_internal_test.go
    - internal/compiler/originvalidate/originvalidate_internal_test.go
  modified:
    - internal/compiler/core/core.go
    - internal/compiler/core/core_test.go
    - internal/compiler/originvalidate/originvalidate.go
    - internal/compiler/originvalidate/originvalidate_test.go
    - internal/compiler/protocol/protocol.go
    - internal/compiler/session/session.go

key-decisions:
  - "D-07-08/D-07-09/D-07-10 implemented verbatim: lang.interface/1 minted alongside a frozen, pinned /0 decoder; Parameters is a slice even at arity 1."
  - "D-07-36: DecodeInterface is the only decode route; originvalidate.CheckSummary now routes through it, so a /0 document and a structurally invalid /1 document are refused before any origin/access question is asked."
  - "D-07-37/D-07-38: ClosureDigest's canonical preimage zeroes itself, prefixes a domain separator, and sorts callee pairs by ID; this plan computes only the zero-callee base case and defers the chaining arm to 07-08, per D-07-38's cycle-refusal-first ordering."
  - "R-01 authority table implemented field-by-field with doc comments naming each field's source; Callable stays at its fail-closed zero value pending 07-02's predicate."
  - "R-02: protocol.InterfaceSummary/InterfaceFunctionAnswer are documented as a deliberately lossy projection; core.Interface remains the artifact of record."
  - "Deviation: fixed a real regression this plan's own strict decoder introduced against a pre-existing session.go verify lane and an existing originvalidate test, both of which passed a BuildInterface-produced summary (with Task 1's deliberately-empty ClosureDigest) through CheckSummary before Task 3 gave ClosureDigest a real value. Resolved by implementing Task 3's zero-callee ClosureDigest before finalizing the plan, matching the plan's own admission that this three-task sequence is internally staged."

patterns-established:
  - "Frozen-/0-fixture discipline: a pinned literal JSON string with a doc comment naming when it was computed and what a failure means, decode -> struct-compare -> re-encode -> byte-compare, never regenerated to make a later test pass."
  - "Every /1 field's doc comment names its R-01 authority and states that absence is a decoder-refused error, not a tag-level nicety."

requirements-completed: [SEM-05, QLT-08]

coverage:
  - id: D1
    description: "lang.interface/1 schema minted: FunctionSignature/ParameterContract/ReturnContract/ForeignReach with every required field non-omitempty and doc-commented to its R-01 authority; PublicOrigin subsumed into ReturnContract."
    requirement: SEM-05
    verification:
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_test.go#TestInterfaceV1FieldInvariantsAcrossCorpus"
        status: pass
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestPreviousPhaseCoreBytesUnchanged"
        status: pass
    human_judgment: false
  - id: D2
    description: "core.DecodeInterface: schema-peek dispatch to pinned /0 (never admissible) or strict /1 validation refusing missing/empty required fields, out-of-closed-set Mode, malformed digests, and duplicate function IDs; originvalidate.CheckSummary rerouted through it."
    requirement: SEM-05
    verification:
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestDecodeInterfaceV1RequiredFieldsRefused"
        status: pass
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestDecodeInterfaceV1ValueDomainRefused"
        status: pass
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestDecodeInterfaceV0NeverAdmissible"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_test.go#TestCheckSummaryRoutesThroughDecodeInterface"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_test.go#TestCheckSummaryRefusesV0Document"
        status: pass
    human_judgment: false
  - id: D3
    description: "QLT-08/D-07-41 mutation-kill controls: the decoder's required-field predicate and the ClosureDigest callee-pair sort are each proven load-bearing via an unexported same-package seeded-mutation seam (D-07-42)."
    requirement: QLT-08
    verification:
      - kind: unit
        ref: "internal/compiler/core/core_internal_test.go#TestDecodeInterfaceRequiredModeMutationKilled"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_internal_test.go#TestClosureDigestSortMutationKilled"
        status: pass
    human_judgment: false
  - id: D4
    description: "ClosureDigest's canonical, non-self-referential, order-independent preimage (zero-callee base case); chaining arm explicitly deferred to 07-08 per D-07-38."
    requirement: SEM-05
    verification:
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_internal_test.go#TestClosureDigestNonSelfReferential"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_internal_test.go#TestClosureDigestDomainSeparatorLoadBearing"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_internal_test.go#TestClosureDigestZeroCalleeBaseCase"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_internal_test.go#TestClosureDigestCalleeOrderIndependent"
        status: pass
    human_judgment: false

duration: ~55min
completed: 2026-09-08
status: complete
---

# Phase 07 Plan 01: lang.interface/1 signature summary, strict decoder, and canonical ClosureDigest Summary

**Minted `lang.interface/1` end to end — real `DecodeInterface` with schema-peek dispatch, a frozen structurally-non-admissible `/0` legacy decoder, and a non-self-referential `ClosureDigest` preimage — closing the three HIGH-severity gaps cross-AI review found in the prior draft (unsupported strict-decoding claim, unwired `/0` dispatch, self-referential digest).**

## Performance

- **Duration:** ~55 min
- **Tasks:** 3 (plus 1 checkpoint:decision, auto-approved per auto-mode)
- **Files modified:** 6 modified, 2 created

## Accomplishments

- `core.InterfaceSchema1` (`lang.interface/1`) minted alongside the frozen `lang.interface/0`; `core.FunctionSignature` widened to `Parameters []ParameterContract`, `Return ReturnContract`, `Callable`, `Fails`, `Foreign ForeignReach`, `ClosureDigest`, dropping `Parameter`/`ReturnType`/`PublicOrigin` (subsumed into `ReturnContract`). Every field's doc comment names its R-01 authority.
- `core.DecodeInterface` implements D-07-36 for real: schema-peek dispatch, a pinned `InterfaceV0`/`FunctionSignatureV0` legacy path that is structurally incapable of being admitted for a call, and strict `/1` validation refusing missing/empty required fields, an out-of-closed-set `Mode`, a malformed `sha256:`+64-hex digest, and duplicate function IDs. `originvalidate.CheckSummary` now routes through it.
- `ClosureDigest` has a canonical, non-self-referential preimage (`ClosureDigestDomainSeparator` + the signature with `ClosureDigest` zeroed + callee `(ID, ClosureDigest)` pairs sorted by ID). Only the zero-callee base case is computed here; the chaining arm over real callees is explicitly deferred to `07-08` per D-07-38.
- Two QLT-08 mutation-kill controls added, each behind an unexported same-package seam per D-07-42 (never an exported production-path var): the decoder's required-field predicate, and the digest's callee-pair sort.
- `TestPreviousPhaseCoreBytesUnchanged` stays green — no `core.Program` byte moved.

## Task Commits

1. **Task 1: End-to-end lang.interface/1 slice — producer to projection** - `6542532` (feat)
2. **Task 2: DecodeInterface — strict /1 validation, pinned /0 dispatch, mutation kill** - `8345e8f` (feat)
3. **Task 3: Canonical ClosureDigest preimage — zero-callee base case** - `b288866` (feat)

**Checkpoint:** the plan's `checkpoint:decision` (ratifying D-07-08/D-07-10/D-07-36/D-07-37 plus R-01/R-02 as one-way commitments) auto-approved under active auto-mode (`workflow.auto_advance: true`, `gate="blocking"` default) — no `gate="blocking-human"` on this task, so per checkpoint protocol Rule 5 the "Proceed as decided" default option was selected and logged before Task 1 began.

## Files Created/Modified

- `internal/compiler/core/core.go` — `/1` schema types, `InterfaceV0`/`FunctionSignatureV0`, `DecodeInterface` and its private validation/seam helpers
- `internal/compiler/core/core_test.go` — pinned-`/0`-bytes, baseline-accepted, required-field, and value-domain decoder tests; extends the existing corpus test
- `internal/compiler/core/core_internal_test.go` — new; white-box `decoderRequireNonEmpty` mutation-kill test (D-07-42)
- `internal/compiler/originvalidate/originvalidate.go` — `BuildInterface` populates every `/1` field from its R-01 authority; `parameterEscapesOwned` (Drops derivation); `CheckSummary` rerouted through `DecodeInterface`; `ClosureDigestDomainSeparator`/`closureDigestPreimageBytes`/`computeClosureDigest`
- `internal/compiler/originvalidate/originvalidate_test.go` — updated tests that referenced the removed `PublicOrigin` field on `FunctionSignature`; new `CheckSummary`-routing tests; corpus test extended to cover `ClosureDigest` shape/determinism
- `internal/compiler/originvalidate/originvalidate_internal_test.go` — new; white-box `ClosureDigest` behavior tests (non-self-reference, domain separator, zero-callee shape, order-independence) and the sort mutation-kill test (D-07-42)
- `internal/compiler/protocol/protocol.go` — R-02 doc comments naming the `InterfaceSummary`/`InterfaceFunctionAnswer` projection as deliberately lossy
- `internal/compiler/session/session.go` — `interfaceProjection` reads `Function.Return` instead of the removed `PublicOrigin` field

## Decisions Made

See `key-decisions` in frontmatter. The one decision made in-flight beyond the plan's own R-01/R-02: `ParameterContract.Drops` is derived via a dedicated `parameterEscapesOwned` walk (follows only `OpMove`/`OpCopy` chains, stops at any borrow/foreign-call op) rather than the return-origin walk `RecomputeOriginPerReturn` already uses, since Drops asks a materially different question (did ownership move to the caller) than origin derivation (did a borrow chain reach the parameter).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Task 2's stricter `CheckSummary` broke a pre-existing verify lane and test before Task 3 landed**
- **Found during:** Task 2, confirmed via full `go test ./internal/compiler/session/...`
- **Issue:** Task 1 deliberately leaves `ClosureDigest` empty ("Task 3 fills it"). Task 2 makes `DecodeInterface` refuse an empty `ClosureDigest`. Between Task 2's and Task 3's commits, any `BuildInterface`-produced document passed through the now-strict `CheckSummary` failed at `core.interface_missing_field` instead of reaching the digest-staleness/origin checks those pre-existing consumers were asserting — breaking `session.go`'s `lane:borrowed-origin-controls` verify lane and `originvalidate_test.go`'s `TestStaleSummaryRejectedBeforeOtherChecks`.
- **Fix:** Set a syntactically valid placeholder `ClosureDigest` on the `TestStaleSummaryRejectedBeforeOtherChecks` fixture (transitional, until Task 3); implemented Task 3 immediately after Task 2 in the same plan so `BuildInterface` fills a real `ClosureDigest` for every function, which independently fixed `session.go`'s verify lane with no production-code change needed there.
- **Files modified:** `internal/compiler/originvalidate/originvalidate_test.go` (test fixture only); `internal/compiler/originvalidate/originvalidate.go` (Task 3's `BuildInterface` wiring, which was the real fix)
- **Verification:** Full `go test ./...` green except the pre-existing, unrelated `PHASE-07-DEBT.md` frontmatter failure (confirmed present before this plan via `git stash`).
- **Committed in:** `8345e8f` (test fixture), `b288866` (real fix)

---

**Total deviations:** 1 auto-fixed (1 bug, self-caused by this plan's own staged task sequence).
**Impact on plan:** No scope creep — the fix is exactly Task 3's own deliverable landing where the plan already scheduled it; only the test fixture needed a placeholder edit to stay meaningful mid-sequence.

## Issues Encountered

None beyond the deviation above.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `lang.interface/1` is a real, decodable, strictly-validated, canonically-digested artifact with one producer (`BuildInterface`) and one decode route (`DecodeInterface`) — ready for `07-02` to define `Callable`'s predicate and for `07-05` onward to consume it in caller admission.
- `ClosureDigest`'s chaining arm (real callee digests) is explicitly NOT built here (D-07-38) — `07-08` must supply it after `07-06`/`07-07` prove acyclicity. The preimage function's signature already accepts a callee-pairs parameter, so `07-08` should not need to change the preimage definition itself.
- No blockers. `go test ./...` is green except the pre-existing (unrelated to this phase's own scope) `PHASE-07-DEBT.md` frontmatter-format failure, which predates this plan.

---
*Phase: 07-calls-signatures-and-call-graph-refusal*
*Completed: 2026-09-08*

## Self-Check: PASSED

All key-files verified present on disk; all three task commits (`6542532`, `8345e8f`, `b288866`) verified present in `git log`. Full `go test ./...` and `go vet ./...` re-run clean except the pre-existing, out-of-scope `PHASE-07-DEBT.md` frontmatter failure (confirmed to predate this plan via `git stash`).
