---
phase: 19-numeric-literals-and-opconst
verified: 2026-09-25T21:20:22Z
status: passed
score: 4/4 must-haves verified
covered_files:
  - .planning/LANGUAGE-MATURITY.md
  - .planning/REQUIREMENTS.md
  - .planning/phases/19-numeric-literals-and-opconst/19-01-PLAN.md
  - .planning/phases/19-numeric-literals-and-opconst/19-01-SUMMARY.md
  - .planning/phases/19-numeric-literals-and-opconst/19-02-PLAN.md
  - .planning/phases/19-numeric-literals-and-opconst/19-02-SUMMARY.md
  - .planning/phases/19-numeric-literals-and-opconst/19-03-PLAN.md
  - .planning/phases/19-numeric-literals-and-opconst/19-03-SUMMARY.md
  - .planning/phases/19-numeric-literals-and-opconst/19-04-PLAN.md
  - .planning/phases/19-numeric-literals-and-opconst/19-04-SUMMARY.md
  - .planning/phases/19-numeric-literals-and-opconst/19-05-PLAN.md
  - .planning/phases/19-numeric-literals-and-opconst/19-05-SUMMARY.md
  - .planning/phases/19-numeric-literals-and-opconst/19-06-PLAN.md
  - .planning/phases/19-numeric-literals-and-opconst/19-06-SUMMARY.md
  - .planning/phases/19-numeric-literals-and-opconst/19-07-PLAN.md
  - .planning/phases/19-numeric-literals-and-opconst/19-07-SUMMARY.md
  - .planning/phases/19-numeric-literals-and-opconst/19-REVIEW.md
  - .planning/phases/19-numeric-literals-and-opconst/19-SCALAR-REVIEW.md
  - .planning/phases/19-numeric-literals-and-opconst/19-VALIDATION.md
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_program_test.go
  - internal/compiler/check/check.go
  - internal/compiler/check/check_phase19_test.go
  - internal/compiler/core/core.go
  - internal/compiler/core/core_test.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/corevalidate/corevalidate_phase19_test.go
  - internal/compiler/interp/interp.go
  - internal/compiler/interp/interp_phase19_integration_test.go
  - internal/compiler/interp/interp_phase19_test.go
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/originvalidate/originvalidate_test.go
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/pathoracle/pathoracle_test.go
  - internal/compiler/session/session_phase19_test.go
  - internal/compiler/session/session_phase7.go
  - internal/compiler/session/session_phase7_mutation_test.go
  - internal/compiler/session/verification_groundedness_test.go
  - internal/compiler/syntax/parser.go
  - internal/compiler/syntax/parser_phase19_test.go
  - internal/compiler/syntax/syntax_test.go
  - testdata/phase16/public-emitter-consumers.json
  - testdata/phase19/literal_tracer.lang
covered_digest: "v1:sha256:552d2deb13489e5453cb180105bb0630f75a4ee13ade5b2d807344c5eca9eb55"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: passed
  previous_score: 4/4
  gaps_closed: []
  gaps_remaining: []
  regressions: []
---

# Phase 19: Numeric Literals and `OpConst` Verification Report

**Phase Goal:** Lang can name a value it was not given — the first value the language creates rather than moves.
**Verified:** 2026-09-25T21:20:22Z
**Status:** passed
**Re-verification:** Yes — refreshed the stale fingerprint; no prior gaps were open.

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|---|---|---|
| 1 | A numeric literal parses and formats losslessly, checks as the single fixed-width unsigned type, interprets, lowers to native code, and produces the named value on both engines. | VERIFIED | `TestPhase19NumericLiteralRoundTrip`, syntax/check tests, and `TestPhase19LiteralRun` passed. `testdata/phase19/literal_tracer.lang` is checked through public interpreter and native command paths; both return `42`. U64 boundary/radix cases pass in the integration tests. |
| 2 | `OpConst` reaches all six consumers, and both exhaustive-dispatch controls are non-vacuous and mutation-sensitive. | VERIFIED | `TestAllOperationKindsRegistered`, `TestAllOperationKindsHandledAtEverySite`, `TestPhase19Dispatch`, and the session mutation-control tests passed. The fixture assertion checks an actual `OpConst` carrying canonical `42`; path and origin peers have explicit constant-root assertions. |
| 3 | The literal-bearing program agrees across interpreter, `-O0`, `-O3`, and `-O3 -flto`. | VERIFIED | `TestPhase19FourTierLiteral` passed for decimal 42, zero, maximum U64, hexadecimal 42, and binary 42. It checks each expected decimal result and uses the all-pairs five-axis comparator. `TestPhase19WrongResultControl` rejects a shared wrong value. |
| 4 | Prior scalar execution bytes remain stable, or any movement is justified in writing. | VERIFIED | `TestPayloadCorpusCharacterizationReplay`, `TestPreviousPhaseGoldenCUnchanged`, `TestPhase16GoldenChangeLedger`, and `TestLegacyEmitterEvidence` passed. `19-SCALAR-REVIEW.md` records unchanged corpus and frozen C digests; no moved golden needs justification. |

**Score:** 4/4 truths verified (0 present, behavior-unverified)

### Required Artifacts

| Artifact | Expected | Status | Details |
|---|---|---|---|
| `internal/compiler/syntax/parser.go` and `parser_phase19_test.go` | Lossless numeric literal syntax | VERIFIED | Parser carries original numeric spelling; round-trip test exercises formatting and reparsing. |
| `internal/compiler/check/check.go` and `check_phase19_test.go` | Checked U64 literal admission and refusal boundaries | VERIFIED | Admission, range, type, malformed, and overflow cases are covered. |
| `internal/compiler/core/core.go` and `corevalidate/corevalidate.go` | Canonical `OpConst` representation and independent validation | VERIFIED | Operation registry includes `OpConst`; validator checks canonical decimal payload and U64 type facts, with forged-fact negative tests. |
| `internal/compiler/interp/interp.go` and `cgen/cgen.go` | Interpreter execution and exact-width native lowering | VERIFIED | Tests cover interpreter constants, U64 native output, entry input, and target guard. |
| `internal/compiler/session/session_phase19_test.go` | Public run, exhaustive dispatch, and four-tier comparison | VERIFIED | Tests exercise the real fixture and exact expected outputs. |
| `19-SCALAR-REVIEW.md` | Reviewed scalar and generated-C baselines | VERIFIED | Documents unchanged baseline digests and test outcomes. |

### Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| Numeric source token | `OpConst` | parser → checker → lowering | WIRED | Round-trip and checker tests validate the literal; core representation carries canonical U64 payload. |
| `OpConst` | Six dispatch consumers | registered operation kind and consumer dispatch | WIRED | Exhaustive registration test and both controls pass; integration tests prove the operation is encountered. |
| `literal_tracer.lang` | Interpreter/native output | public `session.Run*CommandFile` paths | WIRED | Both return pass and the asserted value `42`. |
| Literal-bearing program | Four execution tiers | session differential helper and five-axis comparator | WIRED | Each tier produces the expected value and all pairs compare. |

### Data-Flow Trace (Level 4)

| Artifact | Data variable | Source | Produces real data | Status |
|---|---|---|---|---|
| Interpreter execution | `Outcome.Value` | Parsed/check-lowered program and `OpConst` execution | Yes; asserted as canonical decimal for five literal forms | FLOWING |
| Native execution | `Outcome.Value` | Generated C constant and native runner | Yes; public native run and four-tier test assert result | FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|---|---|---|---|
| Phase 19 implementation and evidence tests | `GOCACHE=/private/tmp/ai-lang-gocache go test ./internal/compiler/syntax ./internal/compiler/check ./internal/compiler/core ./internal/compiler/corevalidate ./internal/compiler/interp ./internal/compiler/pathoracle ./internal/compiler/originvalidate ./internal/compiler/cgen ./internal/compiler/session -run 'TestPhase19|TestAllOperationKinds(Registered|HandledAtEverySite)|TestPayloadCorpusCharacterizationReplay|TestPreviousPhaseGoldenCUnchanged|TestPhase16GoldenChangeLedger' -count=1` | Re-run during fingerprint refresh; all nine packages passed (session 5.727s). Covers literal parsing/checking/interpreter/native execution, four-tier comparison, and scalar/golden evidence. | PASS |
| Both exhaustive-dispatch mutation controls | `GOCACHE=/private/tmp/ai-lang-gocache go test ./internal/compiler/core ./internal/compiler/session -run 'TestPhase7DispatchControlsMutationKilled|TestPhase7DispatchControlsMutationKilledPhase07Lane' -count=1` | Both packages passed; the in-process and CLI-observable controls reject seeded dispatch omissions. | PASS |

### Probe Execution

No probe scripts are declared by Phase 19 plans or summaries; probe execution was not applicable.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|---|---|---|---|---|
| VAL-01 | 19-01 through 19-07 | Numeric literal can be written, checked, interpreted, and lowered. | SATISFIED | Syntax/check/core/interpreter/native tests and public run smoke checks passed. |
| VAL-02 | 19-04 through 19-07 | Every operation kind is handled at all six dispatch sites, proven by exhaustive-dispatch control. | SATISFIED | Registration, dispatch controls, mutation tests, and constant-root peer assertions passed. |
| VAL-03 | 19-02, 19-06, 19-07 | Literal-bearing program agrees across interpreter and three native optimization tiers. | SATISFIED | Four-tier exact-result tests and five-axis all-pairs comparison passed. |

No additional Phase 19 requirement mappings were orphaned.

### Decision Coverage

All five trackable `19-CONTEXT.md` decisions are honored by shipped artifacts (`check.decision-coverage-verify`, non-blocking gate).

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|---|---:|---|---|---|
| — | — | None in Phase 19 implementation files. Matches for words such as “placeholder” are test identifiers/comments for unrelated schema tests or literal fields, not unfinished implementations. | — | No blocker or warning. |

### Human Verification Required

None. Phase validation states all behaviors have automated verification, and the scalar/frozen-C review found no golden movement requiring intent review.

### Gaps Summary

No goal gaps remain. The four roadmap truths and VAL-01 through VAL-03 are supported by implementation evidence and passing focused behavioral tests. The full repository suite is reported as passing after the regression fixes.

---

_Verified: 2026-09-25T21:20:22Z_
_Verifier: the agent (gsd-verifier)_

### Advisory (New Scope, Unevidenced)

Not applicable: this was an initial verification, not a re-verification.
