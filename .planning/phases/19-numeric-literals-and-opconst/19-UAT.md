---
status: complete
phase: 19-numeric-literals-and-opconst
source:
  - 19-01-SUMMARY.md
  - 19-02-SUMMARY.md
  - 19-03-SUMMARY.md
  - 19-04-SUMMARY.md
  - 19-05-SUMMARY.md
  - 19-06-SUMMARY.md
  - 19-07-SUMMARY.md
started: 2026-09-25T16:00:00-04:00
updated: 2026-09-25T16:01:29-04:00
---

## Current Test

[testing complete]

## Tests

### 1. Direct numeric U64-return source witness
expected: The numeric source witness retains its current diagnostic frontier.
result: pass
source: automated
coverage_id: D1 (19-01)

### 2. Overflow and malformed numeric source witnesses
expected: Overflow and malformed spellings produce separate refused source witnesses.
result: pass
source: automated
coverage_id: D2 (19-01)

### 3. Scalar projection compatibility
expected: U64 values project as canonical decimal while legacy scalar and tagged values keep their established projections.
result: pass
source: automated
coverage_id: D1 (19-02)

### 4. Scalar replay mutation control and literal refusal frontier
expected: The scalar replay detects a seeded projection fault and the literal tracer remains refused at the Wave 1 frontier.
result: pass
source: automated
coverage_id: D2 (19-02)

### 5. Numeric syntax and canonical formatting
expected: Decimal, hexadecimal, and binary integer spellings survive lexing, parsing, and formatting; malformed candidates are refused as whole tokens.
result: pass
source: automated
verification: go test ./internal/compiler/syntax ./internal/compiler/session -run 'TestPhase19' -count=1

### 6. U64 ability and OpConst semantic identity
expected: U64 ability and OpConst preserve canonical semantic identity and legacy JSON.
result: pass
source: automated
coverage_id: D1 (19-04)

### 7. Literal admission and refusal boundaries
expected: Valid direct literals lower to typed OpConst facts and overflow or type mismatches are refused.
result: pass
source: automated
coverage_id: D2 (19-04)

### 8. Independent core validation of constants
expected: Core validation admits typed canonical constants and rejects forged facts.
result: pass
source: automated
coverage_id: D1 (19-05)

### 9. Constant origin and path roots
expected: Path and origin peers treat constants as source-free roots without inherited loans or parameter origins.
result: pass
source: automated
coverage_id: D2 (19-05)

### 10. Literal placement in match expressions
expected: U64 literals are available in match entry prefixes and arm bodies.
result: pass
source: automated
coverage_id: D3 (19-05)

### 11. Interpreter execution of U64 constants
expected: The interpreter executes canonical U64 constants and preserves existing scalar projection.
result: pass
source: automated
coverage_id: D1 (19-06)

### 12. Exact-width native lowering
expected: Native C emits and runs exact-width U64 constants and rejects a missing exact-width capability.
result: pass
source: automated
coverage_id: D2 (19-06)

### 13. OpConst registration and exhaustive dispatch
expected: OpConst is registered and a real literal fixture reaches all six consumers in both exhaustive controls.
result: pass
source: automated
coverage_id: D1 (19-07)

### 14. Four-tier exact literal results
expected: Decimal, zero, maximum U64, hexadecimal, and binary literals return exact decimal values on all four tiers and compare across five axes.
result: pass
source: automated
coverage_id: D2 (19-07)

### 15. Public run commands and frozen baselines
expected: Public interpreter and native run commands return the literal value; scalar and frozen C baselines remain unchanged.
result: pass
source: automated
coverage_id: D3 (19-07)

## Automated Evidence

- `GOCACHE=/private/tmp/ai-lang-gocache go test ./internal/compiler/syntax ./internal/compiler/check ./internal/compiler/core ./internal/compiler/corevalidate ./internal/compiler/interp ./internal/compiler/pathoracle ./internal/compiler/originvalidate ./internal/compiler/cgen ./internal/compiler/session -run 'TestPhase19|TestAllOperationKinds(Registered|HandledAtEverySite)|TestPayloadCorpusCharacterizationReplay|TestPreviousPhaseGoldenCUnchanged|TestPhase16GoldenChangeLedger' -count=1` — passed; all nine packages.
- `GOCACHE=/private/tmp/ai-lang-gocache go test ./internal/compiler/check ./internal/compiler/cgen -count=1` — passed.
- `GOCACHE=/private/tmp/ai-lang-gocache go test ./internal/compiler/core ./internal/compiler/session -run 'Test(AllOperationKinds|Phase7DispatchControlsMutationKilled|Phase19Dispatch)' -count=1` — passed.

## Summary

total: 15
passed: 15
issues: 0
pending: 0
skipped: 0

## Gaps

[none]
