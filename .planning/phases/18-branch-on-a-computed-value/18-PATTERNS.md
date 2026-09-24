# Phase 18: Branch on a Computed Value — Pattern Map

**Mapped:** 2026-09-24  
**Files analyzed:** 13 likely implementation, test, fixture, and CI paths  
**Analogs found:** 12 / 13

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `internal/compiler/syntax/parser.go` | parser | source-to-AST transform | `internal/compiler/syntax/syntax_test.go` (`TestGeneratedBranchBodiesRoundTrip`) | role-match |
| `internal/compiler/ast/ast.go` | model | transform | `internal/compiler/ast/ast.go:101-123` (`Body`, `MatchExpr`, `MatchArm`) | exact |
| `internal/compiler/check/check.go` | semantic checker / CFG builder | transform | `internal/compiler/check/check.go:2211-2255` (`checkBranch`) | exact |
| `internal/compiler/check/check_branch_test.go` | test | batch / transform | `TestBranchEdgeLastUseAcceptAndReject` | exact |
| `internal/compiler/corevalidate/corevalidate.go` | independent validator | transform / admission | `internal/compiler/corevalidate/corevalidate.go:885-923` (`branch`) | exact |
| `internal/compiler/originvalidate/originvalidate.go` | independent validator | batch / transform | `internal/compiler/originvalidate/originvalidate_payload_test.go` | role-match |
| `internal/compiler/interp/interp.go` | interpreter | execution / transform | `internal/compiler/session/session_branch_test.go:19` (`TestBranchInterpreterNative`) | role-match |
| `internal/compiler/cgen/cgen_program.go` | native emitter | transform / file I/O (generated C) | `internal/compiler/cgen/cgen_program_test.go:139-232` | exact |
| `internal/compiler/session/session_phase18_test.go` | integration test | request-response / differential | `internal/compiler/session/session_phase17_test.go:180-207` | role-match |
| `internal/compiler/session/session_payload_control_test.go` | mutation test | differential | `TestPayloadSlotSwapMutationKilled` | exact |
| `testdata/phase18/*.lang` | fixture/configuration | source-to-core transform | `testdata/phase17/*.lang`, `testdata/phase12/*.lang` | exact |
| `.github/workflows/ci.yml` | CI config | batch | `.github/workflows/ci.yml:35-79` | exact |
| `internal/compiler/session/session_phase5_compare.go` | comparator utility | transform / differential | `Phase5CompareProgramEngines` | exact |

The list follows the required fixture-first sequence: the first CTL-01 source fixture must be checked in and its refusal pinned before production admission changes. Add tests beside their owning packages; extend existing fixtures/harnesses instead of adding a test-only compiler path.

## Pattern Assignments

### Parser, AST, and source fixtures

**Analogs:** `internal/compiler/ast/ast.go:101-123`; `internal/compiler/syntax/syntax_test.go` (`TestGeneratedBranchBodiesRoundTrip`); `testdata/phase17/return_type_tracer.lang`.

`ast.Body` is a closed variant (`MatchExpr` or `Linear`), while `MatchExpr` carries a named scrutinee and arms. Keep the existing terminal-match syntax and preserve source spans. If computed-prefix syntax needs representation, make the narrowest AST extension that preserves the closed-variant invariant. Parser tests should round-trip the new form and retain exact source spans; source fixtures should separately witness the initial refusal, Result-returning computed match, payload return, and loan-live-in-one-arm case.

**Validation pattern:** fixture files live in phase-numbered `testdata` directories and tests load them through project-path helpers or explicit fixture readers. The initial frontier test pins the existing diagnostic by ID and must later prove it moves through the production parser/checker path; do not guess the diagnostic string in a planning artifact.

### Checker and production CFG

**Analog:** `internal/compiler/check/check.go:2211-2255` (`checkBranch`), with current admission at `check.go:296-330`.

`checkBranch` owns type facts, stable place identities, entry/join/arm blocks, and branch edges. Current source admission enforces parameter-only matching at lines 307-309 and the branch builder seeds its data type and first place from the parameter. Extend this path to resolve the actual in-scope scrutinee place and type, while keeping stable global ordinals across linear-prefix and arm operations. Do not add a second CFG builder or branch law.

**Ownership analog:** `internal/compiler/check/check_branch_test.go`, `TestBranchEdgeLastUseAcceptAndReject`; bounded liveness is implemented by `loanLivenessFixpoint` / `loanLivenessBound` in `check.go:2714-2732,2784-2852`. Preserve the existing `point`/`edge` endpoint vocabulary, cycle refusal, and fail-closed bound `4 × blocks × (distinct loans + 1)`; tests should drive the branch through source, not only hand-constructed CFG data.

### Independent core and origin admission

**Core analog:** `internal/compiler/corevalidate/corevalidate.go:885-923` and `926-955`; branch scrutinee admission currently ties the match to the parameter. Keep this peer source-blind: derive scrutinee existence, type, and ownership from core facts and reject forged/malformed places independently of checker success. Existing `corevalidate_branch_test.go` and mutation-matrix tests show where structural refusal cases belong.

**Origin analog:** `internal/compiler/originvalidate/originvalidate.go:1-8` documents its intentional non-importing boundary; `originvalidate_payload_test.go:15-77` shows independent payload-origin derivation and an understated-origin negative control. Do not make origin validation call checker helpers or trust checker-derived scrutinee/type facts. Add the smallest peer tests needed to show valid computed places pass and malformed or forged facts fail.

### Interpreter, emitter, and differential acceptance

**Interpreter/native analogs:** `internal/compiler/session/session_branch_test.go:19` (`TestBranchInterpreterNative`); `internal/compiler/cgen/cgen_program_test.go:139-232` (branch tracer and payload lowering). Keep execution on the existing core interpreter and native path (`emitProgram`); preserve the shared branch representation and existing payload layout contract.

**Five-axis comparator analog:** `internal/compiler/session/session_phase5_compare.go:14-27,66-125`. `Phase5CompareEngines` checks the four execution axes; `Phase5CompareProgramEngines` first validates schema-2 documents with the independent execution peer. The diagnostic-ID axis is the separate refusal comparator. CTL-02 must use the full session/program acceptance route and all required engines (`interpreter`, `-O0`, `-O3`, and `-O3 -flto`) as the applicable harness specifies; do not describe the four-axis helper alone as five-axis proof.

**Call-contract analog:** `internal/compiler/session/session_phase17_test.go:1-97,180-207` demonstrates parsing, checking, independent core/origin peers, and deterministic interpreter execution around a Result-returning contract. Reuse its fixture-loading and peer-run conventions, but ensure Phase 18 adds source-level matching of the returned computed value and native/differential evidence.

### Payload wrong-slot mutation

**Analog:** `internal/compiler/session/session_payload_control_test.go:212-277`, `TestPayloadSlotSwapMutationKilled`.

The Phase 12 control explicitly demonstrates the current blind spot: the seam confirms injection, yet tag-only terminal values make a wrong-slot write invisible. CTL-03 must extend the source return/evidence shape until the seeded write causes a disagreement specifically on `session.AxisTerminalOutcome`. Assert mutation injection count and retain an unmutated companion that agrees. A green test that merely records the still-invisible mutation is not acceptance for CTL-03.

### Recurring CI

**Analog:** `.github/workflows/ci.yml:35-79` has macOS and Linux jobs, requires installed Clang, and already runs vet, build, all Go tests, and race tests; the evidence aggregate has explicit named focused commands. Add a focused Phase 18 command only if Wave 0 measures show useful latency and its recurring regression value justifies maintenance. Ensure it exercises real source fixtures and native evidence; do not create a redundant workflow or leave a skip as a verdict.

## Shared Patterns

- **Independent derivation:** checker, core validator, and origin validator maintain separate trust boundaries; do not share semantic facts across those peers.
- **Fixture-first evidence:** pin a refused `.lang` construct before opening it, then assert that the production path crosses that frontier.
- **Fail-closed resource limits:** preserve parser/CFG ceilings, cycle refusal, and bounded liveness; return an explicit refusal instead of partial analysis.
- **One branch law:** reuse terminal `match`, `core.Match`, existing blocks/edges, interpreter, and unified C emitter. `if`/`else`, arithmetic, loops, aggregate values, and wider call arity are outside this phase.
- **Automation preference:** objective acceptance belongs in focused tests and recurring macOS/Linux CI when measured recurring value justifies runtime and maintenance. Keep manual review only for qualities that tests cannot judge; do not make objective acceptance a human UAT step.

## No Analog Found

| File | Role | Data Flow | Reason |
|---|---|---|---|
| A named Phase 18 session integration test file (if created) | integration test | differential | No existing Phase 18 file exists yet; `session_phase17_test.go` is the closest session/peer/call precedent and `session_branch_test.go` is the closest branch execution precedent. |

## Metadata

**Analog search scope:** `internal/compiler/{syntax,ast,check,corevalidate,originvalidate,interp,cgen,session}`, `testdata/phase12`, `testdata/phase17`, `.github/workflows`.  
**Tracked-source check:** all named implementation/test/CI analogs were verified with `git ls-files`; none are runtime mirrors.  
**Pattern extraction date:** 2026-09-24
