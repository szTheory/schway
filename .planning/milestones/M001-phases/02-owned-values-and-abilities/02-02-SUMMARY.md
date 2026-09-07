---
phase: 02-owned-values-and-abilities
plan: "02"
subsystem: compiler
tags: [go, abilities, generics, aggregates, bounded-parsing, tdd]

requires:
  - phase: 02-owned-values-and-abilities
    plan: "01"
    provides: Sealed primitive abilities, structured TypeRefs, linear typed-core facts, and exact test selection
provides:
  - Canonical Box and Pair source shapes materialized as ordered typed-core ability facts
  - Private structural conjunction proven over every 32-bit leaf mask and 1,024 Pair combinations
  - Request-local combiner call proof with stable negative witness paths and sealed authority surfaces
  - Lossless canonical generic syntax bounded at depth 64 and 4,096 type nodes
affects: [02-03, core-validation, owned-native-execution, evidence]

actuals:
  tokens: 6588
  tasks: 3
  commits: 6

tech-stack:
  added: []
  patterns: [private structural conjunction, request-local derivation seam, independent exhaustive oracle, declaration-bounded recovery]

key-files:
  created:
    - internal/compiler/ability/ability_test.go
    - testdata/phase2/ability_shapes.lang
  modified:
    - internal/compiler/ability/ability.go
    - internal/compiler/session/session_test.go
    - internal/compiler/syntax/parser.go
    - internal/compiler/syntax/syntax_test.go

key-decisions:
  - "Box and Pair derive through one package-private field-wise conjunction installed on a request-local deriver; no mutable global hook or arbitrary production mask exists."
  - "Negative ability witnesses select the first failing structural child in source order and prepend sealed constructor field names."
  - "Parser-boundary punctuation normalization clears lexer truncation only when no genuinely unknown token remains."

patterns-established:
  - "Non-tautological structural proof: an injected request-local spy proves the sealed TypeRef path invokes the same private combiner exhausted by an independent test oracle."
  - "Bounded generic recovery: malformed type applications stop at parameter, body, or declaration boundaries without consuming the next declaration."

requirements-progressed: [OWN-02]

coverage:
  - id: D1
    description: "Canonical Box<Byte>, Box<Buffer>, and Pair<Byte, Buffer> source shapes produce ordered abilities and smallest stable negative witnesses in typed core."
    requirement: OWN-02
    verification:
      - kind: integration
        ref: "internal/compiler/session/session_test.go#TestSourceBoxPairAbilityFacts,TestSourceCannotGrantAbilityRoots"
        status: pass
    human_judgment: false
  - id: D2
    description: "All 32 Box masks and 1,024 Pair mask combinations agree with an independent conjunction oracle while production derivation demonstrably uses the private combiner."
    requirement: OWN-02
    verification:
      - kind: unit
        ref: "internal/compiler/ability/ability_test.go#TestBoxAllAbilityMasks,TestPairAllAbilityMasks,TestTypeRefDerivationUsesStructuralCombiner,TestNegativeWitnessPath,TestArbitraryMasksRemainTestPrivate"
        status: pass
    human_judgment: false
  - id: D3
    description: "Phase 2 generic syntax is lossless and canonical, recovers within declaration boundaries, and terminates at explicit depth and node caps."
    requirement: OWN-02
    verification:
      - kind: unit
        ref: "internal/compiler/syntax/syntax_test.go#TestOwnershipRoundTrip,TestOwnershipRecovery,TestTypeApplicationLimits,FuzzParseFormat"
        status: pass
      - kind: other
        ref: "env GOCACHE=/tmp/ai-lang-phase2-cache go test ./..."
        status: pass
    human_judgment: false

duration: 9min
completed: 2026-09-03
status: complete
---

# Phase 02 Plan 02: Structural Ability Proof Summary

**Canonical Box and Pair source now reaches ordered typed-core ability facts through one sealed structural conjunction, exhaustively pressure-tested without widening source or wire authority.**

## Performance

- **Duration:** 9 min
- **Started:** 2026-09-03T20:32:11Z
- **Completed:** 2026-09-03T20:40:40Z
- **Tasks:** 3
- **Files modified:** 6

## Accomplishments

- Added one canonical source fixture covering `Box<Byte>`, `Box<Buffer>`, and `Pair<Byte, Buffer>` through the real parser, checker, session, and serialized typed-core path.
- Replaced per-ability structural recursion with a request-local private combiner, then independently exhausted all 32 leaf masks and 1,024 Pair combinations and proved the production TypeRef path invokes it.
- Preserved sealed authority: arbitrary masks remain test-private, the combiner is unexported, and source/AST/core/checker surfaces cannot accept positive ability roots.
- Made type-application parsing deterministic at depth 64 and 4,096 nodes while preserving bytes, comments, canonical spacing, fixed points, and declaration-local recovery.

## Task Commits

Each TDD task was committed with a RED contract followed by its GREEN implementation:

1. **Task 02-02-01 RED: source ability contracts** - `69a4de0` (test)
2. **Task 02-02-01 GREEN: canonical Box/Pair source facts** - `04b1afe` (feat)
3. **Task 02-02-02 RED: exhaustive structural model** - `54a7984` (test)
4. **Task 02-02-02 GREEN: sealed structural combiner** - `71a3cb3` (feat)
5. **Task 02-02-03 RED: bounded generic syntax contracts** - `4f25c06` (test)
6. **Task 02-02-03 GREEN: canonical bounded recovery** - `652bbe8` (fix)

## Files Created/Modified

- `testdata/phase2/ability_shapes.lang` - Canonical generic and aggregate source fixture with a preserved structural comment.
- `internal/compiler/session/session_test.go` - End-to-end typed-core fact and source-authority rejection contracts.
- `internal/compiler/ability/ability.go` - Private ability set, pure conjunction, request-local derivation seam, limits, and stable witnesses.
- `internal/compiler/ability/ability_test.go` - Independent exhaustive oracle, spy proof, witness checks, and authority-surface guards.
- `internal/compiler/syntax/parser.go` - Normalized punctuation truncation fix and enclosing-boundary recovery.
- `internal/compiler/syntax/syntax_test.go` - Phase 2 fixed/fuzz seeds, closed-body span erasure, recovery, and exact cap tests.

## Decisions Made

- Structural constructor abilities are field-wise conjunctions over private value sets; only ordered granted facts and negative witnesses cross the read-only exported boundary.
- The production entry point creates a fresh deriver and installs `combineStructural` on it, allowing same-package proof without any mutable or global test hook.
- Witnesses choose the first failing child in structural source order, making paths both minimal and deterministic.
- Type punctuation remains normalized at the parser boundary so Phase 1 lexer behavior and scope remain unchanged.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Cleared false lexer truncation after generic punctuation normalization**

- **Found during:** Task 02-02-03 (`TestTypeApplicationLimits` RED run)
- **Issue:** More than 20 valid `<`, `>`, or `,` tokens set the Phase 1 lexer truncation bit before the parser normalized those provisional tokens, so a valid depth-64 type received `syntax.too_many_errors` after all primary diagnostics were removed.
- **Fix:** Clear the pre-normalization truncation bit only when normalization leaves no genuinely unknown tokens; invalid UTF-8 and unexpected bytes still fail closed.
- **Files modified:** `internal/compiler/syntax/parser.go`
- **Verification:** Exact syntax selector and full Go suite passed.
- **Committed in:** `652bbe8`

---

**Total deviations:** 1 auto-fixed bug
**Impact on plan:** The fix was required for the specified valid boundary and preserved the parser-owned provisional token seam.

## Issues Encountered

- The first asymmetric-law assertion required every sample to differ from every incorrect law, but conjunction necessarily equals the first child for some subset inputs. The test was corrected while RED to distinguish first-child, union, and both implication directions across the asymmetric sample set; the exhaustive independent expected value remains direct per-bit boolean conjunction.
- Plan 02-01 had already established sealed Box/Pair recursion, so Task 1's GREEN implementation was the missing canonical source fixture; Task 2 then replaced that recursion with the required private combiner and proved the production call path.

## Verification Evidence

- Session exact selector passed both source-to-core and source-authority tests.
- Ability exact selector passed all five named exhaustive, spy, witness, and surface tests.
- Syntax exact selector passed all four named round-trip, recovery, limit, and fuzz-seed tests.
- `env GOCACHE=/tmp/ai-lang-phase2-cache go test ./...` exited 0 across every package after the final code change.
- `git diff --check` reported no whitespace errors.

## Known Stubs

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plan 02-03 can add causal ownership rejection and independent short-sequence pressure on top of sealed, ordered TypeFacts.
- `OWN-02` is progressed by this plan but remains pending until final Phase 2 verification, together with `OWN-01`.
- No blockers remain for sequential Phase 2 execution.

## Self-Check: PASSED

All six key implementation/test files, the canonical fixture, this summary, and all six TDD commits were found. The summary contains `requirements-progressed: [OWN-02]` and no `requirements-completed` claim.

---
*Phase: 02-owned-values-and-abilities*
*Completed: 2026-09-03*
