---
phase: 02-owned-values-and-abilities
plan: "01"
subsystem: compiler
tags: [go, ownership, affine-values, typed-core, interpreter, tdd]

requires:
  - phase: 01-canonical-pure-spine
    provides: Lossless syntax, typed match core, deterministic interpreter, and stable /0 schemas
provides:
  - Explicit Buffer transfer and implicit Byte copy from source through typed core to interpreter events
  - Sealed five-ability derivation for Byte and Buffer without arbitrary production masks
  - Closed match-or-linear bodies with function-local semantic ordinal identities
  - Fail-closed exact Go test selection with a negative self-test
affects: [02-02, ownership-checking, core-validation, native-owned-execution, evidence]

actuals:
  tokens: 9580
  tasks: 2
  commits: 5

tech-stack:
  added: []
  patterns: [closed body union, sealed ability roots, inert execution facts, ordinal semantic IDs, exact test selection]

key-files:
  created:
    - scripts/assert-go-tests.sh
    - internal/compiler/ability/ability.go
    - internal/compiler/execution/execution.go
    - testdata/phase2/owned_transfer.lang
    - testdata/phase2/implicit_copy.lang
  modified:
    - internal/compiler/syntax/parser.go
    - internal/compiler/core/core.go
    - internal/compiler/check/check.go
    - internal/compiler/interp/interp.go
    - internal/compiler/session/session.go
    - internal/compiler/session/session_test.go

key-decisions:
  - "Linear type, place, operation, and point IDs are function-local source-order ordinals, never source offsets."
  - "Phase 1 match artifacts retain lang.core/0 and lang.execution/0 while owned linear artifacts use /1."
  - "Ownership token normalization lives at the declared parser boundary so the Phase 1 lexer remains unchanged."

patterns-established:
  - "Closed body union: every function has exactly one of match or linear, checked before interpretation."
  - "Engine-neutral execution facts: interpreter behavior emits records owned by the execution package."
  - "Sealed abilities: Byte and Buffer facts come only from compiler-owned structural rules."

requirements-completed: [OWN-01, OWN-02]

coverage:
  - id: D1
    description: "Explicit Buffer take transfers exactly once and emits deterministic owned interpreter events."
    requirement: OWN-01
    verification:
      - kind: integration
        ref: "internal/compiler/session/session_test.go#TestOwnedTransferInterpreter"
        status: pass
      - kind: other
        ref: "env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh --self-test ./internal/compiler/session TestTogglePipeline TestOwnedTransferInterpreter TestImplicitByteCopy"
        status: pass
    human_judgment: false
  - id: D2
    description: "Bare Byte binding copies implicitly while its source remains initialized."
    requirement: OWN-02
    verification:
      - kind: integration
        ref: "internal/compiler/session/session_test.go#TestImplicitByteCopy"
        status: pass
    human_judgment: false
  - id: D3
    description: "Closed bodies, semantic ordinal IDs, and feature-specific /0 and /1 schemas are preserved."
    requirement: OWN-01
    verification:
      - kind: integration
        ref: "internal/compiler/session/session_test.go#TestClosedBodyUnion,TestLinearIdentityStability,TestFeatureSpecificCoreExecutionSchemas"
        status: pass
      - kind: other
        ref: "env GOCACHE=/tmp/ai-lang-phase2-cache go test ./..."
        status: pass
    human_judgment: false

duration: 10min
completed: 2026-09-03
status: complete
---

# Phase 02 Plan 01: Owned Interpreter Tracer Summary

**Explicit Buffer transfer and implicit Byte copy now lower to stable linear core operations and deterministic `/1` interpreter events while Phase 1 remains byte-compatible on `/0`.**

## Performance

- **Duration:** 10 min
- **Started:** 2026-09-03T20:18:39Z
- **Completed:** 2026-09-03T20:28:36Z
- **Tasks:** 2
- **Files modified:** 14

## Accomplishments

- Added canonical owned-transfer and implicit-copy source programs with provisional formatter-owned `let`, `take`, `borrow`, and bounded type-application parsing.
- Added explicit linear type/place/operation facts, sealed Byte/Buffer ability derivation, and independent interpreter replay into inert engine-neutral execution records.
- Added an exact Go test selector whose self-test proves a nonexistent sentinel is rejected before known and planned positive targets run.
- Preserved untouched Phase 1 core/execution schemas and behavior while assigning stable function-local ordinal identities to Phase 2 facts.

## Task Commits

Each task was committed atomically with TDD RED and GREEN gates:

1. **Task 02-01-01 RED: owned-value tracer contracts** - `d011ecb` (test)
2. **Task 02-01-01 GREEN: owned interpreter slice** - `42717dc` (feat)
3. **Task 02-01-02 RED: closed-body and identity contracts** - `24fd934` (test)
4. **Task 02-01-02 GREEN: ordinal linear identities** - `37b8cce` (feat)
5. **GREEN refactor: keep token handling in declared parser scope** - `7ae24f0` (refactor)

## Files Created/Modified

- `scripts/assert-go-tests.sh` - Exact discovery/execution gate with a guaranteed-negative self-test.
- `internal/compiler/syntax/token.go` - Provisional ownership and type-application token kinds.
- `internal/compiler/syntax/parser.go` - Closed body parsing, bounded type applications, and parser-boundary token normalization.
- `internal/compiler/syntax/format.go` - Canonical straight-line bindings and generic type punctuation.
- `internal/compiler/ast/ast.go` - Closed match-or-linear AST body and structured type references.
- `internal/compiler/core/core.go` - Linear bodies, type/ability/place/operation facts, and `/1` schema.
- `internal/compiler/ability/ability.go` - Sealed Byte/Buffer and Box/Pair structural ability rules.
- `internal/compiler/check/check.go` - Explicit copy/move/borrow/return lowering with ordinal identities.
- `internal/compiler/execution/execution.go` - Engine-neutral versioned outcome and event records.
- `internal/compiler/interp/interp.go` - Private replay for match and linear core variants.
- `internal/compiler/session/session.go` - Feature-specific deterministic interpreter inputs.
- `internal/compiler/session/session_test.go` - Vertical tracer, closed union, identity, and schema contracts.
- `testdata/phase2/owned_transfer.lang` - Canonical explicit Buffer transfer tracer.
- `testdata/phase2/implicit_copy.lang` - Canonical implicit Byte copy tracer.

## Decisions Made

- Linear identities use source-order semantic ordinals within the function; comments, byte offsets, paths, maps, and addresses do not participate.
- The execution package owns inert records while the interpreter retains private transition behavior.
- Byte grants copy/drop/share/send/escape; Buffer grants drop/send/escape. Production code exposes no arbitrary mask constructor.
- Phase 1 match bodies and artifacts retain `/0`; only the new linear owned feature selects `/1`.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Removed incidental type spans from semantic AST equality**

- **Found during:** Task 02-01-01 full-suite verification
- **Issue:** The initial structured type node stored byte offsets, causing Phase 1 generated format round-trips to compare unequal after whitespace canonicalization.
- **Fix:** Kept diagnostic spans at the existing parameter/function boundaries and removed offsets from semantic type identity.
- **Files modified:** `internal/compiler/ast/ast.go`, `internal/compiler/syntax/parser.go`
- **Verification:** `env GOCACHE=/tmp/ai-lang-phase2-cache go test ./...`
- **Committed in:** `42717dc`

---

**Total deviations:** 1 auto-fixed bug
**Impact on plan:** The fix preserved the required Phase 1 behavior and strengthened the prohibition on incidental identity data without expanding feature scope.

## Issues Encountered

- Token recognition was initially placed in `syntax/lexer.go`, which was outside this plan's declared files. The final net diff restores that file exactly and normalizes provisional ownership tokens in the declared parser boundary (`7ae24f0`).

## Verification Evidence

- Tracer gate: self-test rejected `TestCodenameLangSelectionGuardMustNotExist`, then the known Phase 1 test and both ownership tests passed.
- Contract gate: `TestClosedBodyUnion`, `TestLinearIdentityStability`, and `TestFeatureSpecificCoreExecutionSchemas` passed through exact discovery/execution.
- Full suite: `env GOCACHE=/tmp/ai-lang-phase2-cache go test ./...` exited 0 across all packages.
- Both committed Phase 2 fixtures exactly match the formatter projection.

## Known Stubs

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plan 02-02 can extend the sealed structural fold through source-visible Box/Pair facts and exhaustive independent masks.
- The linear core and execution `/1` seams are ready for independent validation, negative ownership diagnostics, native event decoding, and evidence binding in later Phase 2 plans.
- No blockers remain for sequential Phase 2 execution.

## Self-Check: PASSED

All key created files and all five task/TDD commits were found on disk and in Git history.

---
*Phase: 02-owned-values-and-abilities*
*Completed: 2026-09-03*
