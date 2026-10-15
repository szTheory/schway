---
phase: "35"
slug: bitwise-operators-literal-shifts-and-hex-literals
status: validated
nyquist_compliant: true
wave_0_complete: true
created: "2026-10-14"
---

# Phase 35 — Validation Strategy

> Executed validation record. The per-task map preserves the exact owner selectors. The 22 task commands passed during Phase 35 execution; this audit refreshed the value-peer transfer lanes, actual implementation mutants, Phase 34 compatibility controls, and closeout evidence grounding. Historical native matrix receipts remain explicitly labeled below.

## Test Infrastructure

| Property | Value |
|---|---|
| Framework | Go 1.24 standard-library testing; existing Clang native runner |
| Config | go.mod; no added test dependency |
| Existing quick baseline | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase34PreChangeCProgramGolden|TestPhase34PreChangeCGoldenIsNotInert|TestPhase34BoolNameBaseline|TestPhase34IdentityBaseline)$' -count=1 -v -timeout=600s` |
| Phase gate | Union of final PLAN automated named selectors; no local whole-package/project suite or broad native matrix |
| Runtime | Focused command runtimes are reported in task summaries and receipts; all Go test commands are bounded by 600 seconds |

## Sampling Rate

- Before compiler edits: commit new immutable enumerated C corpus and numeric-route policy baselines; require the reached byte-drift control.
- Per task: run the focused owner selector after the last edit, inspect exit status and require each selected test name in verbose/JSON run output.
- Per wave: corresponding final-plan selectors, including inherited relevant controls.
- Before phase completion: all final scoped requirement/decision/security controls freshly passed; no matcher with zero test executions may count as evidence.
- Existing recurring CI owns broad lanes; Phase 38 owns hosted two-host codedigest binding. Local host evidence cannot close missing host receipts.

## Per-Task Verification Map

Eight plans, 22 tasks, seven execution waves. Wave 5 has disjoint 35-05/35-06; all other dependencies are sequential. Baseline commits 35-01 precede compiler edits. 35-05 task 2 owns the example, exhaustive fixture registration, and independently authored `value-vectors.json`; 35-07 consumes these committed artifacts after 35-05/06, reading the literal fixture for independent engine comparators and optimization/sanitizer evidence. The task rows below are execution receipts mapped to their exact owner commands and summaries, not a claim of a phase-wide review.

| Task | Wave | Requirements | Task type | Behavior / owner obligation | Exact command and fails_when | Status |
|---|---|---|---|---|---|---|
| 35-01-01 | 1 | BIT-01, BIT-02, BIT-03, BIT-04 | tracer | Commit one source-to-C baseline path, then its complete frozen inventory | [35-01-01](#verify-35-01-01) | passed; see 35-01-SUMMARY.md |
| 35-01-02 | 1 | BIT-01, BIT-02, BIT-03, BIT-04 | auto | Freeze route-local numeric policy and characterize recursive admission | [35-01-02](#verify-35-01-02) | passed; see 35-01-SUMMARY.md |
| 35-02-01 | 2 | BIT-01, BIT-02, BIT-03 | tracer | Parse and format the locked grouped operator witness end to end | [35-02-01](#verify-35-02-01) | passed; see 35-02-SUMMARY.md |
| 35-02-02 | 2 | BIT-01, BIT-02, BIT-03 | auto | Bound every scalar recursive entry and preserve recovery siblings | [35-02-02](#verify-35-02-02) | passed; see 35-02-SUMMARY.md |
| 35-02-03 | 2 | BIT-01, BIT-02, BIT-03 | auto | Complete hex admission with exact lexical ownership | [35-02-03](#verify-35-02-03) | passed; see 35-02-SUMMARY.md |
| 35-03-01 | 3 | BIT-01, BIT-02, BIT-03 | tracer | Check one grouped literal shift from source through canonical core | [35-03-01](#verify-35-03-01) | passed; see 35-03-SUMMARY.md |
| 35-03-02 | 3 | BIT-01, BIT-02, BIT-03 | auto | Admit only a provably unused Byte entry without widening scalar input | [35-03-02](#verify-35-03-02) | passed; see 35-03-SUMMARY.md |
| 35-03-03 | 3 | BIT-01, BIT-02, BIT-03 | auto | Reach independent checker result and literal-refusal mutants | [35-03-03](#verify-35-03-03) | passed; see 35-03-SUMMARY.md |
| 35-04-01 | 4 | BIT-01, BIT-02 | tracer | Validate one raw literal shift and unused-Byte bridge through both peers | [35-04-01](#verify-35-04-01) | passed; see 35-04-SUMMARY.md |
| 35-04-02 | 4 | BIT-01, BIT-02 | auto | Complete core field/range laws and all scalar transfer routes | [35-04-02](#verify-35-04-02) | passed; see 35-04-SUMMARY.md |
| 35-04-03 | 4 | BIT-01, BIT-02 | auto | Complete origin laws, exact transfers and publication reachability | [35-04-03](#verify-35-04-03) | passed; see 35-04-SUMMARY.md |
| 35-05-01 | 5 | BIT-01, BIT-02 | tracer | Execute one checked literal shift to an event-free interpreter outcome | [35-05-01](#verify-35-05-01) | passed; see 35-05-SUMMARY.md |
| 35-05-02 | 5 | BIT-01, BIT-02 | auto | Expand literal vectors and execute the unused-Byte bridge | [35-05-02](#verify-35-05-02) | passed; see 35-05-SUMMARY.md |
| 35-05-03 | 5 | BIT-01, BIT-02 | auto | Validate per-op causal admission and forged-event rejection | [35-05-03](#verify-35-05-03) | passed; see 35-05-SUMMARY.md |
| 35-06-01 | 5 | BIT-01, BIT-02 | tracer | Compile one checked literal shift to its pinned native answer | [35-06-01](#verify-35-06-01) | passed; see 35-06-SUMMARY.md |
| 35-06-02 | 5 | BIT-01, BIT-02 | auto | Census all scalar/event seams and reach unsafe generated-shift controls | [35-06-02](#verify-35-06-02) | passed; see 35-06-SUMMARY.md |
| 35-07-01 | 6 | BIT-01, BIT-02, BIT-03 | tracer | Run the exact unused-Byte witness through both public conformance engines | [35-07-01](#verify-35-07-01) | passed; see 35-07-SUMMARY.md |
| 35-07-02 | 6 | BIT-01, BIT-02, BIT-03 | auto | Pin focused optimization/sanitizer answers and installed source facts | [35-07-02](#verify-35-07-02) | passed; see 35-07-SUMMARY.md |
| 35-07-03 | 6 | BIT-01, BIT-02, BIT-03 | auto | Keep complete independent manifest and limit census controls current | [35-07-03](#verify-35-07-03) | passed; see 35-07-SUMMARY.md |
| 35-08-01 | 7 | BIT-01, BIT-02, BIT-03, BIT-04 | tracer | Run Adler with its one-line shift through existing consumers | [35-08-01](#verify-35-08-01) | passed; see 35-EVIDENCE.md |
| 35-08-02 | 7 | BIT-01, BIT-02, BIT-03, BIT-04 | auto | Document exact operator semantics and update observed transition facts | [35-08-02](#verify-35-08-02) | passed; see 35-EVIDENCE.md |
| 35-08-03 | 7 | BIT-01, BIT-02, BIT-03, BIT-04 | auto | Ground scoped compatibility and coverage receipts for independent verification | [35-08-03](#verify-35-08-03) | passed; see 35-EVIDENCE.md |

All automated invocations use the isolated cache and 600-second Go test timeout. The historical Phase 35-07 O0/O3/O3-LTO and UBSan lanes had no skips; this audit did not rerun that matrix. Hosted Phase 38 binding remains separate and pending.

Execution update (2026-10-14): all 22 task rows passed their owner selectors, with task evidence in the corresponding summaries and 35-08 receipts in `35-EVIDENCE.md`. Fresh 2026-10-14 audit reruns include the checker/corevalidate/originvalidate transfer lanes, 15 actual implementation-source mutations plus 3 no-op overlay controls, Phase 34 historical golden/admission controls, and the final evidence/summary gates. See [35-MUTATION-EVIDENCE.md](35-MUTATION-EVIDENCE.md) for exact mutation outputs and source bindings. This validation record does not mark phase-wide security review, code review, or hosted verification complete.

### Verify 35-01-01

Owned files: internal/compiler/session/session_phase35_prechange_test.go, testdata/phase35/prechange-corpus.json.

Each selector below was checked in its owner package and executed; verbose output must show every selected top-level name as RUN and PASS in the correct package. Commands assume repo root.

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase35PreChangeEmitterCorpus|TestPhase35PreChangeCorpusControls)$' -count=1 -v -timeout=600s
```

fails_when: Either named test is absent, a selected emitter is unexercised, no refusal or emitted rows exist, changed bytes/IDs/status or omitted rows escape attribution, or baseline capture can overwrite literals during normal verification.

Acceptance: Revision-pinned source inventory equals the selected tracked corpus, all three entry points have explicit outcome rows, reached controls fail their comparator with the named cause, and clean rerun passes.

### Verify 35-01-02

Owned files: internal/compiler/check/check_phase35_baseline_test.go, internal/compiler/syntax/parser_phase35_baseline_test.go, testdata/phase35/numeric-policy-baseline.json.

Each selector below was checked in its owner package and executed; verbose output must show every selected top-level name as RUN and PASS in the correct package. Commands assume repo root.

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/check ./internal/compiler/syntax -run '^(TestPhase35NumericPolicyBaseline|TestPhase35NestedIndexBaseline)$' -count=1 -v -timeout=600s
```

fails_when: The two owner-specific tests do not both execute, scalar/legacy policy is unified, historical depth observations are rewritten from current output, or baseline commits follow any production edit.

Acceptance: Both route policies and observed old nesting statuses are revision-pinned; legal Bool identity witnesses remain unchanged; the future depth ledger can distinguish history from the new bound.

### Verify 35-02-01

Owned files: internal/compiler/syntax/parser.go, internal/compiler/syntax/format.go, internal/compiler/surface/refusals.go, internal/compiler/syntax/parser_phase35_test.go.

Each selector below was checked in its owner package and executed; verbose output must show every selected top-level name as RUN and PASS in the correct package. Commands assume repo root.

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/syntax -run '^(TestPhase35ScalarGrammar|TestPhase35FormatterRoundTrip|TestFormatLetBinaryExpressions)$' -count=1 -v -timeout=600s
```

fails_when: Any named test is missing, formatted source changes the admitted tree, spaced angles become a shift, logical OR is split, equal-tier operations right associate, or generic/declaration punctuation drifts.

Acceptance: AST for the locked expression and discriminating alternate grouping differ as required; numeric spellings and CST bytes survive; all old and new formatting examples are idempotent.

### Verify 35-02-02

Owned files: internal/compiler/syntax/parser.go, internal/compiler/surface/refusals.go, internal/compiler/syntax/parser_phase35_test.go, testdata/phase35/numeric-delta.json.

Each selector below was checked in its owner package and executed; verbose output must show every selected top-level name as RUN and PASS in the correct package. Commands assume repo root.

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/syntax -run '^(TestPhase35ExpressionDepth|TestPhase35FormatterRoundTrip|TestPhase34RecoveryCorpus|TestPhase34RecoveryControls)$' -count=1 -v -timeout=600s
```

fails_when: Depth 64 is refused, 65 is admitted, recursive index or logical-not bypasses the bound, recovery loops/panics, independent siblings disappear, or the resource delta rewrites baseline facts.

Acceptance: Named limit applies to all recursive scalar paths with exact boundary/recovery results, distinct type-depth identity, and a dated explicit resource delta.

### Verify 35-02-03

Owned files: internal/compiler/syntax/lexer.go, internal/compiler/syntax/parser.go, internal/compiler/check/check.go, internal/compiler/syntax/numeric_phase35_test.go, internal/compiler/check/check_phase35_numeric_test.go.

Each selector below was checked in its owner package and executed; verbose output must show every selected top-level name as RUN and PASS in the correct package. Commands assume repo root.

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/syntax ./internal/compiler/check -run '^(TestPhase35NumericOwnership|TestPhase35NumericRoutePolicy|TestPhase19NumericToken|TestPhase19NumericMalformed|TestPhase19NumericLiteralRoundTrip|TestPhase19LiteralAdmission|TestPhase19LiteralRange|TestPhase19NumericRefusalFrontiers)$' -count=1 -v -timeout=600s
```

fails_when: A selected test fails to run, malformed input is truncated/duplicated or loses another cause, uppercase legacy hex fails, overflow code is silently unified, spelling changes, or 010/0B10 policy changes.

Acceptance: All decimal-admitted contexts gain matching valid hex values; exact malformed ownership, range refusals and formatting remain pinned; both numeric route debts remain explicit.

### Verify 35-03-01

Owned files: internal/compiler/core/core.go, internal/compiler/core/core_phase35_test.go, internal/compiler/check/scalar.go, internal/compiler/check/check_phase35_test.go.

Each selector below was checked in its owner package and executed; verbose output must show every selected top-level name as RUN and PASS in the correct package. Commands assume repo root.

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/core ./internal/compiler/check -run '^(TestPhase35CoreShape|TestPhase35ShiftAdmission|TestAllOperationKindsRegistered)$' -count=1 -v -timeout=600s
```

fails_when: Named tests are absent, grouped literal creates an amount place, missing/noncanonical immediate is produced, ConstU64 appears on another kind, new schema fields/IDs drift, or nonliteral/64 source is accepted.

Acceptance: One complete source-to-core shift tracer passes before operator expansion, all kinds are registered, and stable-field contracts are explicit.

### Verify 35-03-02

Owned files: internal/compiler/check/check.go, internal/compiler/check/scalar.go, internal/compiler/check/check_phase35_test.go.

Each selector below was checked in its owner package and executed; verbose output must show every selected top-level name as RUN and PASS in the correct package. Commands assume repo root.

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/check -run '^(TestPhase35UnusedByteSourceAdmission|TestPhase35ShiftAdmission|TestPhase34BoolSpellingsRemainIdentifiers|TestPhase34CallArityBoundaries)$' -count=1 -v -timeout=600s
```

fails_when: Any Byte reference is accepted, a Byte value is copied/converted as U64, existing Byte program routing changes, signature/arity expands, or the witness fails to lower.

Acceptance: Unused Byte with U64/Bool locals lowers safely; all Byte-use forms fail the dedicated law; legacy source behavior is unchanged.

### Verify 35-03-03

Owned files: internal/compiler/check/scalar.go, internal/compiler/check/check_phase35_test.go.

Each selector below was checked in its owner package and executed; verbose output must show every selected top-level name as RUN and PASS in the correct package. Commands assume repo root.

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/check -run '^(TestPhase35ScalarTransferMutants|TestPhase35ShiftAdmission|TestPhase35UnusedByteSourceAdmission|TestPhase35NumericRoutePolicy)$' -count=1 -v -timeout=600s
```

fails_when: Any per-op fault is unreached or leaves the expected analysis verdict unchanged, any source restriction-removal control escapes, malformed/overflow diagnostics change ownership, or restoration does not pass.

Acceptance: All five exact transfer controls and separate literal/range/Byte source-law controls are reached and restored with named reasons.

### Verify 35-04-01

Owned files: internal/compiler/corevalidate/corevalidate.go, internal/compiler/corevalidate/corevalidate_phase35_test.go, internal/compiler/originvalidate/originvalidate.go, internal/compiler/originvalidate/originvalidate_phase35_test.go.

Each selector below was checked in its owner package and executed; verbose output must show every selected top-level name as RUN and PASS in the correct package. Commands assume repo root.

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/corevalidate ./internal/compiler/originvalidate -run '^(TestPhase35ShiftImmediateAdmission|TestPhase35UnusedByteIRAdmission)$' -count=1 -v -timeout=600s
```

fails_when: Either owner-specific test is missing, either peer trusts producer admission, Byte use passes, immediate shape is accepted with RightID or wrong type, or valid immediate cannot reach both authorities.

Acceptance: One raw-core tracer passes independently through both peers; per-peer malformed and Byte controls produce their exact law failures.

### Verify 35-04-02

Owned files: internal/compiler/corevalidate/corevalidate.go, internal/compiler/corevalidate/corevalidate_phase35_test.go.

Each selector below was checked in its owner package and executed; verbose output must show every selected top-level name as RUN and PASS in the correct package. Commands assume repo root.

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/corevalidate -run '^(TestPhase35ShiftImmediateAdmission|TestPhase35ScalarTransferMutants|TestPhase35AnalysisTriggers|TestPhase35UnusedByteIRAdmission|TestPhase19OpConstIndependentAdmission|TestPhase19ForgedOpConstFactsRefused)$' -count=1 -v -timeout=600s
```

fails_when: Any required law/trigger/transfer route is unexercised, a refusal-removal or wrong-result mutant escapes, raw count64 is admitted, or a control is not restored.

Acceptance: Corevalidation has independent complete immediate/type/range controls and reached per-op transfers across distinct analysis routes.

### Verify 35-04-03

Owned files: internal/compiler/originvalidate/originvalidate.go, internal/compiler/originvalidate/originvalidate_phase35_test.go.

Each selector below was checked in its owner package and executed; verbose output must show every selected top-level name as RUN and PASS in the correct package. Commands assume repo root.

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/originvalidate -run '^(TestPhase35ShiftImmediateAdmission|TestPhase35ScalarTransferMutants|TestPhase35AnalysisTriggers|TestPhase35UnusedByteIRAdmission)$' -count=1 -v -timeout=600s
```

fails_when: A no-remainder new op bypasses admission/publication, any field/range law removal or value fault escapes its intended sink, or origin delegates value/type derivation to another peer.

Acceptance: Origin independently derives and publishes all five new operation obligations; every targeted admission/transfer/trigger fault is reached and restored.

### Verify 35-05-01

Owned files: internal/compiler/interp/interp.go, internal/compiler/interp/interp_phase35_test.go, internal/compiler/pathoracle/pathoracle.go, internal/compiler/pathoracle/pathoracle_phase35_test.go.

Each selector below was checked in its owner package and executed; verbose output must show every selected top-level name as RUN and PASS in the correct package. Commands assume repo root.

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/interp ./internal/compiler/pathoracle -run '^(TestPhase35BitwiseShiftVectorsEventFree|TestPhase35PureOperationAncestry)$' -count=1 -v -timeout=600s
```

fails_when: Either owner test is missing, outcome differs, shift emits a pure-op event/defect, pathoracle rejects a registered op or invents a loan, or malformed runtime count/type reaches unchecked evaluation.

Acceptance: One exact runtime tracer and ancestry check pass before full vector expansion, with real operation coverage.

### Verify 35-05-02

Owned files: internal/compiler/interp/interp.go, internal/compiler/interp/interp_phase35_test.go, examples/bitwise.schway, internal/compiler/core/core_test.go, testdata/phase35/value-vectors.json.

Each selector below was checked in its owner package and executed; verbose output must show every selected top-level name as RUN and PASS in the correct package. Commands assume repo root.

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/interp -run '^(TestPhase35BitwiseShiftVectorsEventFree|TestPhase35InterpreterWrongAnswerControls|TestPhase35UnusedByteRuntime)$' -count=1 -v -timeout=600s
```

fails_when: Any per-op wrong answer passes, high-bit truncation/logical behavior is wrong, Byte is converted/read as U64, input rules change, no-event assertions are vacuous, the real example is missing or lacks a reached kind, the literal vector fixture is missing or differs from independently authored expected answers, or the exhaustive fixture path does not resolve.

Acceptance: All literal vectors and per-op wrong-answer controls execute; independently authored value-vectors.json is interpreter-checked and committed for 35-07; the exact tracked Byte-unused example yields264 with unchanged causal envelope, contains all five reached kinds and has an additive exhaustiveDispatchFixtures registration ready for 35-07's all-site check.

### Verify 35-05-03

Owned files: internal/compiler/executionpeer/executionpeer.go, internal/compiler/executionpeer/executionpeer_phase35_test.go.

Each selector below was checked in its owner package and executed; verbose output must show every selected top-level name as RUN and PASS in the correct package. Commands assume repo root.

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/executionpeer -run '^TestPhase35PureOperationCausalControls$' -count=1 -v -timeout=600s
```

fails_when: The test is absent, a per-op forged step/defect is accepted or rejected for an unrelated reason, a classification fault is unreachable, or numeric replay/events are introduced.

Acceptance: Five clean no-event admissions and ten per-op fabricated-event controls reach the causal law, restore and pass.

### Verify 35-06-01

Owned files: internal/compiler/cgen/cgen.go, internal/compiler/cgen/cgen_program.go, internal/compiler/cgen/cgen_phase35_test.go.

Each selector below was checked in its owner package and executed; verbose output must show every selected top-level name as RUN and PASS in the correct package. Commands assume repo root.

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen -run '^(TestPhase35NativeShiftTracer|TestPhase35EmitterAdmissionGates)$' -count=1 -v -timeout=600s
```

fails_when: Either named test is absent, no native executable runs, literal answer differs, malformed rawcore reaches any emitter, count is dynamic/64+, or unsigned operand/final exact-width cast is absent.

Acceptance: Native O0 tracer returns the independently pinned answer; three route gates have explicit admitted/refused outcomes and malformed-immediate rejection.

### Verify 35-06-02

Owned files: internal/compiler/cgen/cgen_program.go, internal/compiler/cgen/cgen_phase35_test.go, testdata/phase35/shift-sites.json.

Each selector below was checked in its owner package and executed; verbose output must show every selected top-level name as RUN and PASS in the correct package. Commands assume repo root.

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen -run '^(TestPhase35UnsignedShiftEmissionAndEventProjection|TestPhase35ShiftSiteControls|TestPhase35EmitterAdmissionGates|TestPhase35NativeShiftTracer)$' -count=1 -v -timeout=600s
```

fails_when: Any consumer is absent/unreached, a pure-op site generates an event/reason, a wrong capacity/projection row escapes, scan passes with zero sites, or dynamic/64+/signed/removal C mutant is accepted.

Acceptance: Finite emitter/projection/capacity inventory is reached and independent site count/unsafe-C controls fail exactly then restore.

### Verify 35-07-01

Owned files: cmd/schway/main_phase35_test.go, internal/compiler/session/session_phase35_values_test.go.

Each selector below was checked in its owner package and executed; verbose output must show every selected top-level name as RUN and PASS in the correct package. Commands assume repo root.

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway ./internal/compiler/session ./internal/compiler/core -run '^(TestPhase35PublicConformanceWitness|TestPhase35FocusedValueAnswers|TestPhase321RunInputRequiredCLI|TestAllOperationKindsHandledAtEverySite)$' -count=1 -v -timeout=600s
```

fails_when: Any expected top-level name is unexecuted, either public outcome differs from264, alternate yields264, app output differs from264 newline, U64 default input appears, or an exhaustive new-kind site is skipped.

Acceptance: Actual CLI executions on both engines pin264 in their conformance projections; application observations and alternate8 are separately asserted; existing input-required behavior and closed-kind dispatch pass.

### Verify 35-07-02

Owned files: internal/compiler/session/session_phase35_values_test.go, internal/compiler/surfacemanifest/manifest.go, internal/compiler/surfacemanifest/manifest_test.go.

Read-only input: testdata/phase35/value-vectors.json, created and committed by 35-05 task 2 before this dependent plan runs. Expected answers are independently authored literals, never captured from either engine or producer output.

Each selector below was checked in its owner package and executed; verbose output must show every selected top-level name as RUN and PASS in the correct package. Commands assume repo root.

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session ./internal/compiler/surfacemanifest -run '^(TestPhase35FocusedValueAnswers|TestPhase35BitwiseOptimizationMatrix|TestPhase35ShiftUBSan|TestPhase35EngineWrongAnswerControls|TestPhase35SurfaceDocument|TestPhase34SurfaceDocument|TestPhase34SurfaceLimits)$' -count=1 -v -timeout=600s
```

fails_when: Any required vector/op/mode is unexecuted or wrong, UBSan does not actually execute shifts, wrong answers pass literal comparison, a required lane skip is called complete, or manifest alters schema/input/call authority.

Acceptance: Literal vectors and per-op controls pass every required local mode, sanitizer executes, and installed contract exposes exact verified current facts; any environment gap is explicit and blocks that lane's acceptance.

### Verify 35-07-03

Owned files: internal/compiler/session/session_phase34_surface_test.go, internal/compiler/session/session_phase34_limit_consumers_test.go, internal/compiler/session/session_phase35_surface_test.go, testdata/phase35/surface-witnesses.json, testdata/phase35/surface-census.json.

Each selector below was checked in its owner package and executed; verbose output must show every selected top-level name as RUN and PASS in the correct package. Commands assume repo root.

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase35SurfaceManifestWitnesses|TestPhase35SurfaceManifestCensus|TestPhase35SurfaceManifestControls|TestPhase35SurfaceLimitConsumers|TestPhase34SurfaceManifestWitnesses|TestPhase34SurfaceManifestCensus|TestPhase34SurfaceManifestControls|TestPhase34SurfaceLimitConsumers)$' -count=1 -v -timeout=600s
```

fails_when: Any name is absent, a producer claim lacks an independent full source, old fixture bytes change, any drift direction/limit consumer is omitted, weakened restriction passes, or comparator is widened to ignore new/old mismatches.

Acceptance: Full manifest and finite consumer census agree independently; all old and new meta-test controls are reached and restored with exact reasons.

### Verify 35-08-01

Owned files: examples/checksum/adler32.schway, internal/compiler/session/session_phase35_adler_test.go.

Each selector below was checked in its owner package and executed; verbose output must show every selected top-level name as RUN and PASS in the correct package. Commands assume repo root.

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session ./internal/compiler/native -run '^(TestPhase35ExistingAdlerAnswers|TestPhase35AdlerWrongShiftControl|TestPhase29InterpreterFixtureVectors|TestPhase29NativeFixtureVectors|TestPhase31CaseMatrix)$' -count=1 -v -timeout=600s
```

fails_when: Any selected owner test is absent, an original expected answer is modified, checksum/adler_low output or cleanup changes, source still contains the sixteen doublings, or count15 control is unreached/uncaught.

Acceptance: Original Phase29 fixture and native Phase31 case matrix answers pass unchanged; source uses b<<16 and wrong-shift control reaches the answer comparator.

### Verify 35-08-02

Owned files: docs/agent-surface.md, .planning/PRODUCT-ROADMAP.md, .planning/LANGUAGE-MATURITY.md, internal/compiler/session/session_phase35_closeout_test.go.

Each selector below was checked in its owner package and executed; verbose output must show every selected top-level name as RUN and PASS in the correct package. Commands assume repo root.

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase35OperatorDocumentation|TestPhase35PublicDocumentValues)$' -count=1 -v -timeout=600s
```

fails_when: Either new test is absent, guide claims wrong264/8/count/low64/Byte/causal semantics, staged host evidence is portrayed as fresh, historical text is overwritten, or dormant seeds/later-phase ownership are promoted silently.

Acceptance: Documented source-visible semantics are independently pinned and copied misleading claims fail; dated current facts and three scoped next actions preserve history and ownership.

### Verify 35-08-03

Owned files: internal/compiler/session/session_phase35_closeout_test.go, .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-COMPATIBILITY-DELTA.md, .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-EVIDENCE.md, .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-VALIDATION.md.

Each selector below was checked in its owner package and executed; verbose output must show every selected top-level name as RUN and PASS in the correct package. Commands assume repo root.

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase35EvidenceGrounding|TestSummaryCoverageGroundedness|TestPhase35PreChangeEmitterCorpus|TestPhase35PreChangeCorpusControls)$' -count=1 -v -timeout=600s
```

fails_when: Any name is absent, an unknown/stale/kind-invalid/missing-lane/causal-numeric control passes, baseline history changes, summary refs do not resolve or a skipped required lane becomes a pass.

Acceptance: All task coverage maps resolve actual tracked test names and closed kinds, the baseline remains immutable, and execution receipts stay scoped to their declared lanes.


The research report's example places `TestPhase31CaseMatrix` under session; actual source is `internal/compiler/native/native_phase31_matrix_test.go`. The plan uses the grounded native package. Proposed research names were checked against actual owner-package tests before they were recorded as executed receipts.

## 2026-10-14 audit reruns

These fresh named runs exercised the three value-deriving owners after the
independent expected-value sinks were corrected. Their earlier copied
operation/input variants remain differential controls; they are not counted as
implementation-source mutants. The session overlay test mutates the actual
operator expressions in temporary source copies and checks all 15 failures at
the peer's independent static zero-remainder sink.

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/check -run '^(TestPhase35ScalarTransferMutants|TestPhase35ShiftAdmission|TestPhase35UnusedByteSourceAdmission|TestPhase35NumericRoutePolicy)$' -count=1 -v -timeout=600s
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/corevalidate -run '^(TestPhase35ShiftImmediateAdmission|TestPhase35ScalarTransferMutants|TestPhase35AnalysisTriggers|TestPhase35UnusedByteIRAdmission|TestPhase19OpConstIndependentAdmission|TestPhase19ForgedOpConstFactsRefused)$' -count=1 -v -timeout=600s
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/originvalidate -run '^(TestPhase35ShiftImmediateAdmission|TestPhase35ScalarTransferMutants|TestPhase35AnalysisTriggers|TestPhase35UnusedByteIRAdmission)$' -count=1 -v -timeout=600s
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase35PeerTransferMutationControls$' -count=1 -v -timeout=600s
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase34PreChangeCProgramGolden|TestPhase34PreChangeCGoldenIsNotInert)$' -count=1 -v -timeout=600s
```

Every command passed. The session mutation run reported three no-op overlay
passes and fifteen named `mutant KILLED` results. The Phase 34 adapter retained
all frozen admission/outcome checks and exact-row hash comparisons; details are
separately recorded in [35-MUTATION-EVIDENCE.md](35-MUTATION-EVIDENCE.md).

## BIT requirement coverage cross-check

| Requirement | Independently observed behavior and named owners | Audit result |
|---|---|---|
| BIT-01 | Grammar and grouping (`TestPhase35ScalarGrammar`); check/corevalidate/originvalidate transfer sinks and 15 actual source-mutant child failures (`TestPhase35ScalarTransferMutants`, `TestPhase35PeerTransferMutationControls`); independent interpreter/native expected answers (`TestPhase35FocusedValueAnswers`, historical `TestPhase35BitwiseOptimizationMatrix`); separate pure causal/event controls (`TestPhase35BitwiseShiftVectorsEventFree`, `TestPhase35PureOperationCausalControls`). | PASS; causal controls are not numeric evidence. |
| BIT-02 | Source literal/range law (`TestPhase35ShiftAdmission`); independent raw-core immediate/type/range laws in corevalidate and originvalidate (`TestPhase35ShiftImmediateAdmission`); value/mode/UBSan shift evidence (`TestPhase35FocusedValueAnswers`, historical `TestPhase35BitwiseOptimizationMatrix` and `TestPhase35ShiftUBSan`); emitted C shift-site controls (`TestPhase35UnsignedShiftEmissionAndEventProjection`, `TestPhase35ShiftSiteControls`). | PASS; O0/O3/O3-LTO/UBSan receipt is historical Phase 35-07 evidence. |
| BIT-03 | Malformed/overflow ownership and route policy (`TestPhase35NumericOwnership`, `TestPhase35NumericRoutePolicy`); frozen decimal/binary/Bool and nesting policy (`TestPhase35NumericPolicyBaseline`, `TestPhase35NestedIndexBaseline`); independent literal answers and formatter/source spelling checks (`TestPhase35FocusedValueAnswers`, `TestPhase35ScalarGrammar`). | PASS. |
| BIT-04 | Existing checksum and `adler_low` interpreter/native expected answers plus reached wrong-shift control (`TestPhase35ExistingAdlerAnswers`, `TestPhase35AdlerWrongShiftControl`); Phase 29/31 consumer outcomes (`TestPhase29InterpreterFixtureVectors`, `TestPhase29NativeFixtureVectors`, `TestPhase31CaseMatrix`); exact source/core/C hashes (`TestPhase35EvidenceGrounding`). | PASS; consumer answer receipt is Phase 35-08 evidence. |

## Wave 0 Requirements

- [x] Immutable baseline corpus with emitting/non-emitting cases, all selected emitters and reached changed-byte control, committed before compiler changes.
- [x] Numeric compatibility (`010`, `0B10`, Bool names), grouped grammar, exact lexical ownership, expression depth/recovery and token/generic formatter controls.
- [x] Independent source/core/origin immediate-only shape/range and unused-Byte bridge controls; scalar transfer fault sinks.
- [x] Interpreter/native literal vectors, per-op causal event-free/forged events, evidence capacity/projection inventories and generated shift static scan mutants.
- [x] Focused O0/O3/LTO and UBSan feature checks, manifest admission/refusal/limit census and bidirectional controls.
- [x] Existing Adler/checksum/adler_low literal answers unchanged and reached wrong-shift control.
- [x] Exact task map, threat IDs, real test paths/names and fails_when bindings filled by final plans.

## Manual-Only Verifications

None intended: the scoped contracts have deterministic proxies. Human readability and whole-loop cold/warm latency remain unmeasured and are not acceptance claims.

## Validation Sign-Off

- [x] Every task has an executable named verify or explicit Wave 0 dependency.
- [x] No three consecutive implementation tasks lack verification.
- [x] Every expected test name was executed, all fault controls reached/restored.
- [x] No watch mode, whole-project local test, push/dispatch or premature hosted binding.
- [x] Validation lifecycle changes to validated/compliant only from fresh execution evidence.

**Validation sign-off (2026-10-14):** the scoped task map, requirement evidence, and Wave 0 checks are complete from executed evidence. Phase 38 hosted binding and phase-wide security/code review remain separately owned.

## Validation Audit 2026-10-14

| Metric | Count |
|---|---|
| Gaps found | 1 |
| Resolved | 1 |
| Escalated | 0 |

## Supplemental admission-refusal verification — 2026-10-14

Owning verification: D-09/D-10/D-11, 35-04 tasks 2/3 and the source literal/range
authority in `TestPhase35ShiftAdmission`.

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase35AdmissionRefusalMutationControls$' -count=1 -v -timeout=600s
```

Fresh local PASS (exit 0, 9.566s): exactly 26 actual-source refusal-removal
controls (12 corevalidate, 12 originvalidate, 2 check), 55 cardinality-checked
source replacement sites and 46 intended named child assertion failures.
Three no-op and three restored-source controls passed, with 34 named raw rows
per peer and nine source rows. All three production source hashes stayed
identical. Compilation/vet/missing-selector/unrelated failures are not kills.
The old source-marker parser defect was independently replayed through a
test-file overlay and the current control passed after that failing replay.

The complete inventory, production hashes, attribution rules, guard/fixture
repairs and limits are in
[35-ADMISSION-MUTATION-EVIDENCE.md](35-ADMISSION-MUTATION-EVIDENCE.md).
This receipt supplements the separate 15 arithmetic transfer mutants; that
harness was unaffected and was not rerun. No production change, broad/native/
hosted run, phase completion, or review/verification ownership change is
claimed.

## Successor refresh after Phase 37 — 2026-10-15

All 22 final-plan task selectors were resolved against actual package test
roots and executed uncached after the evidence adapter repair in
35-REFRESH-FIX.md. Fresh result: 66 distinct names / 72 package roots across
13 packages, zero missing/fail/skip, 25.468s wall time. This includes the
focused 33-vector O0/O3/O3-LTO and 16-shift UBSan tests on this host; these are
new local executions, separate from the preserved historical lane receipts.
Exact commands, current input hashes and raw-log hashes: 35-REFRESH-RUN.json.

The sole newly observed gap was the missing description in synthetic coverage
metadata; repaired and verified without weakening the shared validator. All
four BIT requirements retain automated coverage. Additional actual-source
mutation controls are checked in the subsequent regression/verifier gate;
this paragraph does not claim their current result. Completed implementation,
UAT, historical evidence, numeric policy and frozen fixtures are preserved.
Phase 38 remains the final audited hosted two-host receipt owner.
