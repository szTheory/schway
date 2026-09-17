# Ownership Kernel Semantics

The executable core: owned places, shared/exclusive loans, transfer, lexical
release, branch joins, and loop fixed points — plus the diagnostic identities
everything downstream binds to.

## Requirements

From the `ownership-kernel` idea (MANIFEST.md):

- Preserve affine ownership, shared/exclusive loans, explicit transfer,
  deterministic lexical release, and stable machine-readable diagnostics.
- Do not infer final source syntax or production performance from kernel
  notation.

## How to Build It

**Model places and loans as separate state machines with stable identities.**
Every program, source operation, semantic event, loan, place, and diagnostic
carries a machine-readable ID. Two defects in spike 001 were found only because
identities were stable enough to compare — one of them was a loan identifier
reused after its loan ended, which made evidence ambiguous. **Loan identities
are unique for the whole program; reuse is rejected even after a loan ends.**

**Use a finite lattice for flow states, then reach a fixed point.** The
branch/loop analyzer in `sources/001-*/ownership/flow.go` carries these place
states across a join:

- definitely live
- moved
- released
- `ownership.maybe_uninitialized` — produced when a conditional move's source is
  used after the join

and these loan states: active, ended, absent. A finite height plus a monotone
transfer is what makes zero-or-more loops converge without a solver. Read-only
loops converge; a repeated move is rejected.

**Join conservatively on disagreement.** A join that gives one static loan
identity different owners or different access modes is rejected. Do not try to
unify them — general multi-source views are the public-origin problem
(`public-interface-origins-abilities.md`), not a join rule.

**Make lexical release structural.** Scopes carry stable identities with
balanced begin/end checks. On scope exit, resources release in **reverse
acquisition order**, and this holds on all four exits: success, expected error,
cancellation, and contained panic. Release evidence records the **owning lexical
scope**, not a nesting depth — depth is not an identity.

**Attach semantic repair candidates to common failures.** Diagnostics carry
code, causes, and structured repair candidates, not just a code.

## What to Avoid

- **One shared transition function for checker and oracle.** Rejected in spike
  001's research table: agreement becomes tautological. Write the checker and
  the oracle separately even though both read one fixture format.
- **Comparing only validity and error codes.** Spike 001 iteration 4 tightened
  comparison to require **identical ordered semantic events**, and iteration 7
  showed why: an `end_loan` before its borrow incorrectly suppressed the
  checker's inferred endpoint, and result-only comparison hid it because the
  program already failed on the earlier operation. Compare the trace, not the
  verdict.
- **Counting the rejecting instruction twice.** Invalid-operation counting in
  the first implementation double-counted it.
- **Letting timestamps or map iteration order into semantic output.** Wall time
  is a measurement field, never a semantic result.
- **Assuming a short counterexample length.** The first fault-injection test
  expected three operations; last-use inference ended the unused loan
  immediately, so keeping the borrow live required a fourth operation
  (`declare p; borrow_shared p as l; move p to q; read_loan l`). Let the search
  tell you the minimum.
- **Parser and surface syntax first.** Explicitly rejected — it freezes
  decoration before semantic behavior. Nothing in spike 001 selects source
  keywords, call markers, a host language, storage strategy, or a GC.

## Constraints

- The path oracle unrolls loops through **three iterations** only; the analyzer
  itself reaches a fixpoint and is unbounded.
- Scopes and flow analysis are **not yet combined** in spike 001.
- The two linear normalizers are independently written and corpus-compared, but
  both could still encode the same mistaken specification. A formal model
  remains separate evidence.
- The oracle models a dynamic ownership monitor — not allocation, not memory
  access.
- Automatic release is structural and infallible; double-failure attachment is
  not implemented.
- The generated alphabet uses one owner, one loan identity, one move target.
  Larger place/loan combinations need covering arrays.

## Measured result (spike 001, VALIDATED)

28/28 fixtures · 12/12 control-flow fixtures · 64/64 generated one-branch
products agree with the path oracle · 7,381/7,381 generated short programs agree
· 89.7% statement coverage · injected checker defect found after 159 generated
programs with a minimal four-operation counterexample.

## Origin

Synthesized from spikes: 001
Source files available in: `sources/001-ownership-kernel-workbench/`
