---
phase: 27-exact-fizzbuzz-and-cross-host-evidence
plan: 05
subsystem: compiler
tags: [literal-output, parser, formatter, typed-core, validation]

# Dependency graph
requires:
  - phase: 27-04
    provides: Unit application entry and typed output operations, with the checker/core CFG path ready for literal writes.
provides:
  - Canonical `write_text("literal")` statement parsing and formatting with escape-safe token spans.
  - A literal-only typed core operation carrying raw spelling, exact decoded bytes, and byte length without a runtime String place.
  - Independent peer checks for literal provenance, shape, and byte-length consistency.
affects: [APP-08, 27-06, phase27-evidence]

# Actuals (#2632)
actuals:
  tokens: 8299
  tasks: 2
  commits: 1

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Quoted source spelling stays lossless in syntax/AST; the checker decodes once into a byte payload and both peers independently re-derive it.
    - Literal writes are typed Unit events with no runtime String value, source operand, or result place.

key-files:
  created: []
  modified:
    - internal/compiler/ast/ast.go
    - internal/compiler/syntax/lexer.go
    - internal/compiler/syntax/parser.go
    - internal/compiler/syntax/format.go
    - internal/compiler/core/core.go
    - internal/compiler/core/core_test.go
    - internal/compiler/check/scalar.go
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/originvalidate/originvalidate.go
    - internal/compiler/session/session_phase27_test.go

key-decisions:
  - "Represent text writes as a dedicated no-result Unit operation with the exact quoted token, decoded bytes, and explicit byte length; do not create a runtime String type."
  - "Let the checker decode the quoted token with strconv.Unquote once; core and origin peers independently compare their own decoding against the stored bytes and length."
  - "Preserve malformed escape spelling through syntax recovery so the checker can reject the complete literal token with a source span."

patterns-established:
  - "A compile-time-only source value may have an AST literal node without gaining a runtime type or place."
  - "Operation registries pin the new typed kind before later plans add interpreter and native execution support."

requirements-completed: [APP-08]

coverage:
  - id: D1
    description: "Quoted text writes accept valid Go-style escapes including NUL, lower to exact bytes and length, and reject dynamic, compound, or malformed operands with source spans; canonical formatting preserves source spelling, comments, recovery, and final LF."
    requirement: APP-08
    verification:
      - kind: unit
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/syntax ./internal/compiler/session -run '^TestPhase27(LiteralWriteAdmission|LiteralEscapeFormat)$' -count=1"
        status: pass
    human_judgment: false
  - id: D2
    description: "Both independent peers admit checked literal facts and refuse mismatched payload/length, name-derived provenance, wrong operation kind, and forged embedded-zero length."
    requirement: APP-08
    verification:
      - kind: unit
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/session -run '^TestPhase27LiteralWriteAdmission$|^TestPhase26(AnalysisExhaustion|BackEdgeAuthority)$' -count=1"
        status: pass
    human_judgment: false
  - id: D3
    description: "The core operation registry includes the new non-terminating literal-write kind."
    verification:
      - kind: unit
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/core -run '^Test(AllOperationKindsRegistered|Phase27TextWriteOperationInventory)$' -count=1"
        status: pass
    human_judgment: false

# Metrics
duration: 38min
completed: 2026-10-03
status: complete
---

# Phase 27 Plan 05: Literal-Only Text Write Admission Summary

**`write_text` now preserves quoted source spelling and lowers only validated literals to exact typed bytes, with independent peer rejection of forged payload facts.**

## Performance

- **Duration:** 38 min
- **Started:** 2026-10-03T15:46:00-04:00
- **Completed:** 2026-10-03T16:24:00-04:00
- **Tasks:** 2
- **Files modified:** 10

## Accomplishments

- Added escape-aware quoted-token scanning, `write_text` statement parsing, and stable canonical formatting without changing literal spelling.
- Lowered only a valid quoted literal to `OpWriteText` with its raw token, decoded byte payload, and exact length; no runtime String type or value place was introduced.
- Added independent core/origin derivation and negative controls for bad lengths, forged bytes/provenance, the wrong operation kind, and an embedded-zero length mutation.
- Preserved Phase 26 scalar-analysis/exhaustion and back-edge refusal controls.

## Task Commits

The two tightly coupled admission tasks were committed together:

1. **Tasks 1–2: Admit literal text writes and reject forged payloads** - `4cfea96` (feat)

**Plan metadata:** pending.

## Files Created/Modified

- `internal/compiler/ast/ast.go` - Documents the text-literal expression and write statement kinds.
- `internal/compiler/syntax/lexer.go` - Finds escaped closing quotes while preserving the raw token for source-spanned checking.
- `internal/compiler/syntax/parser.go` - Parses `write_text` statements and quoted literal operands.
- `internal/compiler/syntax/format.go` - Keeps the literal write call in canonical statement position.
- `internal/compiler/core/core.go` - Adds the no-result `OpWriteText` kind and source/bytes/length facts.
- `internal/compiler/core/core_test.go` - Registers the Phase 27 operation kind and confirms it is not a terminator.
- `internal/compiler/check/scalar.go` - Decodes valid quoted literals once and emits their exact byte payload and length.
- `internal/compiler/corevalidate/corevalidate.go` - Independently validates payload provenance, shape, and length.
- `internal/compiler/originvalidate/originvalidate.go` - Re-derives literal provenance and byte length independently.
- `internal/compiler/session/session_phase27_test.go` - Covers exact byte lowering, escapes, formatting, source spans, and forged-core controls.

## Decisions Made

- Reused the existing lossless `TokenString` and scalar statement pipeline, restricting text values to the `write_text` call site.
- Stored both the exact quoted token and decoded bytes so the independent peers can verify that the payload corresponds to a literal.
- Kept the operation payload byte-oriented, including embedded zero, rather than adding a general string value or UTF-8 runtime type.

## Deviations from Plan

The checker implementation lives in `internal/compiler/check/scalar.go` rather than the plan's listed `check.go`; that file owns `checkScalarFunction` and the scalar operation builder. The AST kind comment and core operation-registry test were also updated to document and register the new narrow operation. These additions support the planned path and do not widen the language surface.

**Total deviations:** 1 necessary file-list extension covering the actual checker implementation and operation registry.

## Issues Encountered

- A branchless Unit witness is refused by the existing CFG validator with `core.release_order_indeterminate` because its terminal block has no incoming edge. The literal witness uses the same guarded CFG shape as the existing Phase 27 Unit-write controls; the validator contract was not broadened in this plan.
- The operation registry test had a stale expected count, so it was updated and its focused selector now passes.

## User Setup Required

None - no external service configuration or added dependency is required.

## Next Phase Readiness

- Plan 27-06 can now add interpreter and native emission for the independently admitted exact byte payload.
- Both Plan 27-05 acceptance selectors and the focused operation-registry check passed, including Phase 26 regression controls. No full project suite was run.
- APP-08's broader end-to-end output and cross-host evidence remains pending; this plan establishes literal admission only.

---
*Phase: 27-exact-fizzbuzz-and-cross-host-evidence*
*Completed: 2026-10-03*
