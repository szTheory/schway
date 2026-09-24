---
spike: 007
idea: ownership-kernel
name: loan-across-branch
type: standard
validates: "Given a borrow created before a branch and live in exactly one arm, the CFG liveness fixpoint accepts it and materializes existing endpoint kinds without per-arm ownership-state merging"
verdict: VALIDATED
related: [002-cfg-edge-last-use, 006-interprocedural-liveness-cost-scaling]
tags: [ownership, cfg, branching, liveness, endpoints, phase-gate]
---

# Spike 007: Loan Across a Branch

## What This Validates

Given an explicit CFG whose entry block creates a shared borrow, when exactly
one successor consumes that loan, then `loanLivenessFixpoint` accepts the CFG
and `materializeLoanEndpoints` classifies the diverging arm with the existing
`edge` endpoint kind and the consuming use with the existing `point` kind.
No third endpoint kind or per-arm ownership-state merge is needed for this
shape.

## Research

This is a pure algorithmic question against the production Go implementation;
there are no external APIs or libraries to research. The roadmap already
defines the falsifiable criterion and its dirty-result contingency in
`.planning/ROADMAP.md` (S-010).

## How to Run

```sh
GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/check -run '^TestEdgeSpecificLiveOut$' -count=1
```

## What to Expect

The test should pass with exactly one `edge` endpoint on the successor that
does not use the loan and one `point` endpoint at the use in the other
successor. Changing the produced edge kind to an invented `arm-merge` kind
must make the focused test fail.

## Investigation Trail

1. Read S-010's phase-gate criterion and the implementation of
   `loanLivenessFixpoint` and `materializeLoanEndpoints`.
2. Found that `TestEdgeSpecificLiveOut` already constructs the required
   topology: `fn:block:entry` contains `OpBorrowShared` and branches to
   `fn:block:used` and `fn:block:unused`; only `used` consumes the borrowed
   place.
3. Ran the focused test successfully. It asserts the loan is live entering
   `used`, not live entering `unused`, and checks the exact endpoint kinds,
   edge, and consuming block.
4. Injected a production-code mutation using Go's `-overlay`, changing the
   materialized edge endpoint kind from `edge` to `arm-merge`. The focused
   test failed because no expected edge endpoint was produced. The mutation
   was killed without changing the working tree.

## Results

**VALIDATED (clean for the S-010 criterion).** The production CFG liveness
algorithm accepts the pre-branch borrow and distinguishes the diverging arm
with the existing edge endpoint representation. The focused regression test
also kills a seeded third-kind mutation. The Phase 18 planning gate is
answered for this algorithmic shape.

Scope boundary: this does not prove that a source-level computed branch is
admitted or correctly lowered. Phase 18 must still use a fixture-first plan
and rerun the loan-crossing fixture through the production source-to-core
path, as its roadmap acceptance criteria require.
