# Roadmap: Schway

## Milestones

- ✅ **M001 — Source-to-Native Semantic Spine** — Phases 1–6 (shipped 2026-09-07) — [archive](milestones/M001-ROADMAP.md)
- ✅ **M002 — Interprocedural Semantic Spine** — Phases 07–13 (shipped 2026-09-14) — [archive](milestones/M002-ROADMAP.md)
- ✅ **M003 — Computation and Honest Instruments** — Phases 14–20 (shipped 2026-09-26) — [archive](milestones/M003-ROADMAP.md)
- ✅ **M004 — Native Emission Ownership and Resource Discharge** — Phases 21–25 (completed 2026-10-02) — [archive](milestones/M004-ROADMAP.md)
- 🚧 **M005 — Practical Computation** — Phases 26–27 (planning)

## Overview

M005 makes ordinary Schway source useful for bounded scalar computation. Phase
26 delivers caller-run `sum_to_n` with checked U64 addition, conditions, and
scalar loops. Phase 27 composes that path with remainder and bounded text to
print exact FizzBuzz, and closes the shared output, documentation, and independent
evidence contract. The existing Go 1.24 standard-library host, interpreter,
single C17/Clang native route, and application runner remain the implementation
path; no new dependency or backend is required.

The current requirements are in [REQUIREMENTS.md](REQUIREMENTS.md). [Research
synthesis](research/SUMMARY.md) records source-inspected boundaries and
historical receipts; no M005 arithmetic or loop check was executed during
research. The living [product roadmap](PRODUCT-ROADMAP.md) and
[language-maturity record](LANGUAGE-MATURITY.md) retain the ranked successors
and current refusal frontier.

## Phases

Phase IDs continue from M004's Phase 25. Each feature phase has a runnable
source/input/output gain. Phase plans will choose the narrow source spelling and
implementation details while preserving the acceptance behavior below.

- [x] **Phase 26: Checked Scalar Sum** - A caller runs bounded `sum_to_n` from ordinary Schway source, with checked U64 addition and safe scalar loops. (completed 2026-10-03)
- [x] **Phase 27: Exact FizzBuzz and Cross-Host Evidence** - A caller runs exact FizzBuzz, and both M005 programs have bounded output and independent interpreter/native evidence. (completed 2026-10-10)

## Phase Details

### Phase 26: Checked Scalar Sum

**Goal**: A developer can run a source-authored scalar loop that computes and prints the exact bounded `sum_to_n` result, with defined overflow and explicit refusal of unsafe loop-carried state.

**Depends on**: Completed M004 Phase 25 application route

**Requirements**: U64-01, FLOW-01, FLOW-02, APP-07

**Success Criteria** (what must be TRUE):

1. A caller can build and run `sum_to_n` through the public application route; input `0`, `10`, and `1,000` produces exactly `0\n`, `55\n`, and `500500\n` on stdout, respectively.
2. Ordinary Schway source can compare U64 values, branch on Bool conditions with `if/else`, and repeat a predicate-controlled scalar loop whose changing U64/Bool state is admitted consistently by checking, independent validation, interpretation, and native execution.
3. In-range U64 addition is exact; overflow produces the same defined checked failure in the interpreter and C17 native path, without a wrapped success result.
4. A loop carrying ownership, a resource, a loan, or loan-derived provenance across its back edge is refused with a stable source-attributed diagnostic. A loop analysis that exceeds its deterministic bound fails closed rather than claiming admission.
5. The `sum_to_n` result is checked against separately authored expected answers in each engine; a reached wrong-result or skipped-iteration control fails. The accepted input bound is `0 ≤ n ≤ 1,000`; a larger input fails without application output.

**Plans**: 5/5 plans complete in 4 waves
**Wave 1**

- [x] 26-01-PLAN.md

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 26-02-PLAN.md
- [x] 26-03-PLAN.md

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 26-04-PLAN.md

**Wave 4** *(blocked on Wave 3 completion)*

- [x] 26-05-PLAN.md

**Runnable witness**: Check in the ordinary `sum_to_n.schway` program and its
public build/run commands before broadening loop support. Pin the exact
stdout/stderr/status contract, including overflow and rejected input, before
implementation. Existing resource/loan loop refusals remain live negative
controls. This phase may implement the comparison/Bool subset of U64-02 needed
by the sum; U64-02 as a complete requirement closes in Phase 27 when remainder
and zero-divisor behavior are admitted.

**Admission and evidence boundary**: Use a finite, monotone scalar CFG
fixed-point analysis with independent peer derivation. Keep dynamic loop-event
occurrences distinguishable wherever an admitted loop can produce events, and
fail closed if the evidence budget is exhausted. Phase 27 owns the full EVD-12
cross-program and dual-host evidence claim. Source inspection and historical
M004 receipts do not prove this phase's new semantics.

### Phase 27: Exact FizzBuzz and Cross-Host Evidence

**Goal**: A developer can run bounded FizzBuzz from ordinary Schway source with exact text and defined remainder behavior, while both M005 programs have reproducible interpreter/native evidence and clean-checkout commands.

**Depends on**: Phase 26

**Requirements**: U64-02, APP-08, APP-09, EVD-12, DX-16

**Success Criteria** (what must be TRUE):

1. A caller can run FizzBuzz for `0 ≤ n ≤ 1,000`. For each integer from 1
   through `n`, stdout contains exactly one line: `FizzBuzz` for multiples of
   15, `Fizz` for other multiples of 3, `Buzz` for other multiples of 5,
   and decimal digits otherwise. Every nonempty result ends with a newline;
   `n = 0` produces empty stdout.
2. U64 comparison, equality, Bool branching, and remainder have one defined
   meaning across checker, independent validators, interpreter, and native C17.
   A constant zero remainder divisor is rejected during checking; a dynamic
   zero divisor produces the same defined checked failure in both engines.
3. Both programs accept only `0 ≤ n ≤ 1,000`; larger inputs fail without
   application output. Their writes obey the existing 65,536-byte ceiling, and
   exceeding it produces a non-success outcome rather than truncated success.
4. Independently pinned answers adjudicate interpreter and native runs
   separately at ordinary and boundary inputs. Reached controls detect wrong
   arithmetic/remainder, skipped iterations, overflow, zero divisors, repeated
   loop-event identity collisions, and exhausted or incomplete evidence.
   Native evidence records actual macOS and Linux execution at the claimed
   source revision.
5. From a fresh checkout, a developer can follow documented commands to build
   and run both programs and find the exact expected output, accepted input
   range, checked failure behavior, and output limit.

**Plans**: 10/10 plans complete in 9 waves

The safety slices enforce three named obligations before composition:
source write-operation admission before byte emission, D-27-04's checked zero
guard before native C remainder, and literal-only text admission before literal
byte emission. Plan 27-02 delivers the runnable end-to-end FizzBuzz tracer.

**Wave 1**

- [x] 27-01-PLAN.md — source/core/independent-peer admission for Unit decimal/LF writes

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 27-02-PLAN.md — runnable interpreter/native Unit FizzBuzz tracer

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 27-03-PLAN.md — equality/remainder source, typed core and static-zero checking

**Wave 4** *(blocked on Wave 3 completion)*

- [x] 27-04-PLAN.md — independent arithmetic peers and guarded runtime remainder

**Wave 5** *(blocked on Wave 4 completion)*

- [x] 27-05-PLAN.md — literal-only text source and independent admission

**Wave 6** *(blocked on Wave 5 completion)*

- [x] 27-06-PLAN.md — literal byte emission and complete exact FizzBuzz

**Wave 7** *(blocked on Wave 6 completion)*

- [x] 27-07-PLAN.md — shared incremental output limit and public failure channels

**Wave 8** *(blocked on Wave 7 completion)*

- [x] 27-08-PLAN.md — independent answers, reached controls, event and capacity evidence
- [x] 27-09-PLAN.md — native host driver, two-host receipt gate, CI matrix

**Wave 9** *(blocked on Wave 8 completion)*

- [x] 27-10-PLAN.md — clean-checkout guide and actual macOS/Linux receipts

**Runnable witness**: Check in `fizzbuzz.schway` with independently authored
golden outputs for `n = 0`, `1`, `15`, and `1,000`; include a rejected
`1,001` case and a reached output-ceiling control. The implementation can
reuse Phase 26's loop and decimal writer, adding only fixed text and U64
remainder needed by this program.

**Admission and evidence boundary**: Keep application stdout/stderr/status
separate from evidence documents and from evidence capacity. Compare each
engine with an independently specified answer, then compare engines. A
successful hosted macOS/Linux receipt must be tied to the emitted source and
declared native inputs; an unavailable host is recorded as unavailable, not
as a pass. General strings, broad IO, division, resource/loan loop carries,
and new dependencies remain outside M005.

## Progress

**Execution order:** 26 → 27

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 26. Checked Scalar Sum | 5/5 | Complete    | 2026-10-03 |
| 27. Exact FizzBuzz and Cross-Host Evidence | 10/10 | Complete    | 2026-10-10 |

**Coverage:** 9/9 M005 requirements have exactly one phase owner. U64-02's
comparison/Bool subset enables the first runnable program; its complete
remainder and zero-divisor guarantee closes in Phase 27. APP-09 and EVD-12 cover
both programs and therefore close with Phase 27, while Phase 26 verifies its
own bounded sum witness on the way.

### Phase 27.1: Address tech debt: build-route origin validation and receipt-gate docs (INSERTED)

**Goal:** Every public application admission route (`schway build`, `schway app verify`, the interpreter application route) applies checker, then `corevalidate`, then the origin peer before entry analysis or emission, so an origin-unpublishable program is refused with its origin code before any build or execution (`schway build` leaves no artifact or sidecar; `schway app verify` writes only its declared fail report, never an executable or other evidence file); and the Phase 27 guide's receipt-gate commands are correct, labeled, and execution-tested.
**Requirements**: None newly mapped; strengthens assurance depth for U64-01, U64-02, FLOW-01, FLOW-02, APP-07, APP-08, APP-09, EVD-12, DX-16.
**Depends on:** Phase 27

**Success Criteria** (what must be TRUE):

1. `schway build testdata/phase25/shared_copy_accept.schway --output X` exits 2 with `status:invalid` and `core.origin_omitted` (same diagnostic as `schway check`), and neither `X` nor `X.schway-build.json` exists; red/green evidence recorded.
2. `BuildApplication`, `RunInterpreterApplication` and `VerifyApplicationCases` refuse checker, corevalidate and origin rejects through one shared helper, with every pre-existing error string unchanged, shown by a behavior-level parity test; the `app verify` file route on the origin reject writes only its fail report (verified false, no cases) and nothing else. Routes are not claimed uniform in CLI status (`build` exits 2 invalid, `app verify` exits 3).
3. All shipped examples that built before still build and match their pinned answers; one fresh hosted two-host `ci.yml` run at the final SHA is green and its fresh Phase 27 receipts pass the gate pinned to that SHA.
4. `examples/phase27/README.md` documents the fresh-artifact use and the pinned re-check of the committed 6224f8b receipts; `TestPhase27GuideContract` executes the pinned command and requires `status=pass`; the gate prints a stderr hint on revision mismatch without changing reason names or exit status.
5. Behavior-neutral audit records are closed: `ValidateFullCoverage` marked test-only, 27-09/27-10 SUMMARY `requirements-completed`, 26-05 amendment confirmed, dated addenda in 26/27 VERIFICATION citing the new run.

**Plans:** 3/3 plans complete in 2 waves

**Wave 1**

- [x] 27.1-01-PLAN.md — shared application admission helper; build/interp/verify routes apply the origin peer (red/green, parity, positive controls)
- [x] 27.1-02-PLAN.md — receipt-gate guide uses, stderr revision hint, executed guide contract, audit record fixes

**Wave 2** *(blocked on Wave 1 completion; contains a human push/dispatch checkpoint)*

- [x] 27.1-03-PLAN.md — transition records and deferred dispositions, then one authorized hosted two-host run with fresh gated receipts and 26/27 addenda
