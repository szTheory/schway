---
spike: 001
idea: ownership-kernel
name: ownership-kernel-workbench
type: standard
validates: "Given small linear, scoped, and branching ownership programs, when independently checked and dynamically monitored, then valid cases agree and injected semantic drift produces stable counterexamples"
verdict: VALIDATED
related: []
tags: [ownership, interpreter, checker, differential, diagnostics]
---

# Spike 001: Ownership kernel workbench

## What this validates

Given small linear, scoped, and branching ownership programs, when independently
implemented analyzers and dynamic/path oracles evaluate them, then:

- accepted linear programs have the same semantic event trace;
- rejected linear programs identify the same stable diagnostic code and
  source operation;
- last-use normalization ends local loans before later non-conflicting access;
- lexical resources release in reverse acquisition order on success, expected
  error, cancellation, and contained panic;
- a deliberately faulty checker produces a small machine-readable
  counterexample;
- branch joins preserve definitely-live, moved, released, and
  maybe-uninitialized states and agree with bounded exhaustive path expansion
  on validity;
- branch-crossing loans preserve definite, ended, and maybe-inactive states,
  with differing origins or modes rejected conservatively.

This validates the experiment mechanism, the first linear state machine, a
bounded branch-join slice, and a finite zero-or-more loop ownership lattice.
It does not validate arbitrary loop semantics, returned origins, generics,
async, FFI, native lowering, final source syntax, or the full ownership
proposal.

## Research

The workbench uses only the Go 1.24 standard library. The spike is about the
project's semantic hypothesis, so an external parser, solver, IR, or test
framework would add more assumptions than evidence at this stage.

| Approach | Strength | Weakness | Disposition |
|---|---|---|---|
| one shared transition function | smallest implementation and impossible internal drift | checker/oracle agreement would be largely tautological | rejected |
| separately coded checker and oracle | exposes accidental implementation disagreement while retaining one fixture format | both can still share a mistaken specification or normalizer | chosen for Gate 1–2 |
| exhaustive formal model checker first | stronger state-space reasoning | delays executable diagnostics and integration shape | later independent evidence |
| parser plus surface syntax first | tests user-facing text | freezes decoration before semantic behavior | rejected for this spike |

The final readiness audit records why independent native, hostile, and formal
evidence remain necessary even after this result.

## How to run

From this directory:

```sh
go test ./...
go test -race ./...
go test -cover ./...
go run ./cmd/ownership-lab -pretty=false -max-depth=4
```

To prove that the differential search can detect a plausible checker defect:

```sh
go run ./cmd/ownership-lab \
  -pretty=false \
  -max-depth=3 \
  -inject-checker-bug
```

The injected-fault command intentionally exits unsuccessfully after reporting
the mismatch.

## What to expect

The ordinary run reports:

- 28 fixture passes and zero fixture failures;
- 12 control-flow fixture passes and zero control-flow failures;
- 7,381 breadth-first generated programs through depth four;
- 64 small branch products agreeing with exhaustive path expansion;
- no checker/oracle disagreement.

The injected checker incorrectly permits moving an owner while a shared loan
remains live. The search reports a four-operation counterexample:

```text
declare p
borrow_shared p as l
move p to q
read_loan l
```

The oracle rejects the move with `ownership.move_while_borrowed`; the faulty
checker accepts the complete program.

## Observability

Every run can emit one JSON document containing:

- schema version;
- fixture comparisons;
- ordered semantic events;
- stable program and source-operation identities;
- diagnostic code, causes, and semantic repair candidates;
- operation counts;
- exhaustive-search size and first minimal mismatch;
- elapsed wall time as a measurement, not a semantic result.

No timestamps or unstable map iteration enter semantic event output.

## Investigation trail

### Iteration 1 — executable state machine

Implemented owned places, shared/exclusive loans, reads, mutation, transfer,
release, lexical cleanup, and last-use loan termination. The first test run
failed before compilation because the sandbox denied Go's default user cache;
the verification commands were rerun with task-local cache directories.

### Iteration 2 — counterexample expectation corrected

The first fault-injection test expected a three-operation program. Last-use
inference correctly ended an unused loan immediately, so a later loan read was
required to keep the borrow live. The expected minimal length became four.

### Iteration 3 — stable loan identity defect

The fault-injection CLI then found a different four-operation case that reused
one loan identifier. The normalizer tracked only its first occurrence, making
the evidence identity ambiguous. Loan identities are now unique for the whole
program; reuse is rejected even after a loan ends. The intended transfer/use
counterexample then became the first mismatch.

### Iteration 4 — evidence tightened

Comparison originally checked only validity and error codes. It now also
requires identical ordered semantic events. Invalid-operation counting was
corrected to avoid double-counting the rejecting instruction, and common
ownership failures now include structured semantic repair candidates.

### Iteration 5 — corpus expanded

Expanded from 14 to 22 fixtures, adding exclusive owner conflicts, loan-ID
reuse, occupied move targets, unknown loans, cancellation, contained panic,
and resource release immediately after a final loan use.

### Iteration 6 — branch joins and loop fixed points

Added a separate forward branch-state analyzer and an exhaustive acyclic path
oracle. Conditional moves now produce an explicit
`ownership.maybe_uninitialized` state when the source is used after a join;
identical transfers on both paths preserve a definitely initialized target;
dead conditional moves remain legal. Five durable flow fixtures, 64 generated
branch products, nested branches, and an unsound-join fault injection agree
with the path oracle. A finite ownership lattice then supported zero-or-more
loop fixed points: read-only loops converge, while a repeated move is rejected
and independently witnessed by bounded path unrolling.

### Iteration 7 — independent loan-end derivation

Replaced the checker/oracle's shared linear loan-end normalizer with two
separately implemented passes and compared their normalized programs across
all fixtures and 7,381 generated inputs. The comparison found that an
`end_loan` occurring before a borrow incorrectly suppressed the checker's
inferred endpoint. The source program already failed on that early operation,
so result-only differential comparison had hidden the support defect. The rule
now considers only an explicit end after the corresponding borrow.

### Iteration 8 — nested lexical scopes

Added stable scope identities, balanced begin/end checks, inner-value
invalidation, reverse-order release at scope exit, and full active-scope unwind
for error, cancellation, and contained-panic exits. Six scope fixtures include
a borrowed view that attempts to outlive its owner's scope. Release evidence
now records the owning lexical scope rather than relying on nesting depth.

### Iteration 9 — loans across branch joins

Extended the finite flow lattice with active, ended, and absent loan states.
Five fixtures cover shared and exclusive loans established before a branch,
conditional creation, conditional ending, and the same loan created on both
paths. A join that gives one static loan identity different owners or modes is
rejected conservatively; general multi-source returned views remain a later
origin-set experiment. Control-flow last-use inference is not smuggled into
this result: valid post-join owner transfer still requires an explicit loan
end in this slice.

## Results

**Verdict: VALIDATED for the bounded ownership kernel and differential harness.**

Fresh verification after the final implementation produced:

- unit/property corpus: pass;
- race detector: pass;
- ownership package statement coverage: 89.7%;
- 28/28 declared fixtures: pass;
- 12/12 control-flow fixtures: pass;
- 64/64 generated one-branch products: dataflow/path-oracle agreement;
- 7,381/7,381 generated short programs: checker/oracle agreement;
- injected checker defect: detected after 159 generated programs with a
  minimal four-operation counterexample.

The strongest finding is methodological: stable semantic identities and
independently executable last-use normalization rules exposed two real defects
in the workbench itself. Branch-state merging also remained small and
explainable when checked against path enumeration, and the ownership lattice
reached loop fixed points without solver machinery. The follow-on
[CFG edge-specific last-use spike](../002-cfg-edge-last-use/README.md) now
derives control-flow last use independently, and the [public origin/ability
spike](../003-public-origins-generic-abilities/README.md) now validates the next
separate-compilation boundary. Independent certification and native lowering
remain before production claims.

## Known limitations

- Branches and lexical scopes may nest and zero-or-more loops reach a finite
  fixed point, but the independent path oracle only unrolls loops through
  three iterations and does not yet combine scopes with flow analysis.
- Branch-state analysis handles explicit loan endpoints, but does not infer
  last use over the control-flow graph or carry loans through loops.
- The two linear normalizers are separately implemented and corpus-compared,
  but both could still encode the same mistaken specification. A small formal
  model remains independent evidence.
- The oracle models a dynamic ownership monitor, not physical allocation or
  memory access.
- Automatic release is structural and infallible; double-failure attachment is
  not implemented.
- The generated alphabet uses one owner, one loan identity, and one move
  target. Larger place/loan combinations need covering arrays or generation.
- Wall-time observations are not performance gates.
- No result here selects source keywords, call markers, a compiler host
  language, storage strategy, or garbage collector.
