---
spike: 002
idea: ownership-kernel
name: cfg-edge-last-use
type: standard
validates: "Given loans used on only some branch or loop paths, backward CFG liveness places path-specific loan ends that agree with bounded per-path normalization without enumerating paths in the analyzer"
verdict: VALIDATED
related: [001]
tags: [ownership, cfg, liveness, property, differential, diagnostics]
---

# Spike 002: CFG edge-specific last use

## What this validates

Given a loan created before branches or loops, when its final use differs by
path, a polynomial backward dataflow analysis should place synthetic loan ends
at either a program point or a control-flow edge. Bounded path expansion then
acts as an independent executable oracle: every analyzed path should accept or
reject exactly as its independently normalized linear form does.

The experiment is specifically intended to answer whether the calm rule
"local access ends at its proven last use" remains implementable and
explainable once control flow is real. It does not attempt public returned
origins, higher-ranked access, async suspension, field-sensitive places, or
native lowering.

## Research

Rust's compiler guide describes non-lexical lifetimes as regions derived from
the control-flow graph and separates move analysis, region inference, loans in
scope, and the final error walk. The NLL RFC describes the in-scope-loan set as
a fixed-point dataflow computation. Polonius makes issue, kill, invalidation,
origin liveness, and CFG propagation explicit relations, while also retaining
faster conservative analysis grades. Clang's dataflow guide gives the general
engineering constraint: a finite-height lattice plus monotone transfer reaches
a fixpoint, and a worklist avoids reprocessing unaffected blocks.

- [Rust compiler borrow checking](https://rustc-dev-guide.rust-lang.org/borrow_check.html)
- [Rust non-lexical lifetime RFC](https://rust-lang.github.io/rfcs/2094-nll.html)
- [Polonius loan analysis](https://rust-lang.github.io/polonius/rules/loans.html)
- [Clang dataflow analysis](https://clang.llvm.org/docs/DataFlowAnalysisIntro.html)

| Approach | Strength | Weakness | Disposition |
|---|---|---|---|
| lexical loan extent | tiny, predictable checker | rejects safe branch-local release and adds ceremony | control only |
| exhaustive paths | exact for a bounded graph and easy to inspect | exponential; loops require a bound | independent oracle |
| backward loan liveness | finite set lattice, edge-sensitive, worklist-friendly | needs careful endpoint materialization and diagnostics | chosen analyzer |
| full origin/loan relations | extends toward returned and multi-source views | introduces origin propagation before local last use is settled | later spike |
| general symbolic path conditions | potentially more path precision | solver cost and diagnostic instability exceed this question | rejected for v0 |

## Gates

1. Deterministic branch, nested-branch, and loop fixtures agree with the path
   oracle.
2. Generated small CFG products produce no disagreement under the normal
   analyzer.
3. Metamorphic changes such as successor reordering and alpha-renaming preserve
   validity and endpoint shape.
4. Deliberately omitting edge endpoints produces a stable, small
   counterexample.
5. A large synthetic CFG reaches a finite worklist fixpoint without path
   enumeration; measurements are reported but not promoted to product budgets.
6. All evidence has stable block, edge, operation, and loan identities.

## How to run

From this directory:

```sh
go test ./...
go test -race ./...
go test -cover ./...
go run ./cmd/cfg-last-use -pretty=false
go run ./cmd/cfg-last-use -pretty=false -inject-edge-bug
```

The injected-fault command is expected to exit unsuccessfully after reporting
the first disagreement.

## What to expect

The ordinary run emits a compact JSON report with:

- 9 fixture passes and zero fixture failures;
- 100 generated diamond programs with no disagreement;
- one 1,501-block/2,000-edge scale analysis completed without path expansion;
- stable endpoint identities such as `edge:entry:else:view` and
  `point:then:0:view`.

The injected-fault run intentionally exits unsuccessfully. It omits all
edge-specific endpoints and finds a disagreement after seven generated
programs: one branch uses a shared view, the other does not, and a mutation
after the join is legal only if the unused edge ends the loan.

## Investigation trail

### Iteration 1 — isolate one semantic question

The prior flow analyzer proved explicit branch-crossing loans and loop place
states but required an explicit loan end before post-join owner access. This
spike isolates automatic control-flow last use. It rejects source-level
`end_loan` operations so manual/inferred mixing cannot accidentally determine
the result.

### Iteration 2 — backward set liveness and edge materialization

Implemented a finite set of live loan identities at each block entry and exit.
A reverse worklist computes a fixpoint. A final use can produce either a point
endpoint after an operation or an edge endpoint where only one successor keeps
the loan live. Loop-carried uses remain live through the back edge and end on
the loop-exit edge.

### Iteration 3 — independent path oracle and masking defense

Bounded path expansion independently converts each CFG path into a linear
program and uses the prior spike's separate linear normalizer as the oracle.
Analyzed paths materialize the CFG endpoints instead. A deliberately late
terminal guard prevents the imported normalizer from silently repairing a
missing CFG endpoint; an owner conflict therefore exposes the injected bug.

### Iteration 4 — machine-feedback budget corrected

The first successful CLI serialized all path traces plus liveness for a
1,501-block graph, producing roughly 24,000 tokens of output. The default
report now returns compact fixture, generation, and scale summaries while
retaining full path evidence only for a mismatch. During this pass the endpoint
schema also stopped omitting operation index zero, which had made a valid
coordinate lossy in JSON.

### Iteration 5 — property and complexity probes

Added exhaustive generation of 100 small diamond products, 500 seeded
randomized/metamorphic checks under block and successor reordering, loan
alpha-renaming, malformed-graph diagnostics, nested branches, zero-or-more
loops, two independent loans, stable JSON, and a 1,501-block scale case. The
large acyclic case required exactly one transfer evaluation per block under
the reverse worklist order.

## Results

**Verdict: VALIDATED for local CFG last-use inference and edge-specific loan
endpoints.**

Fresh verification after implementation established:

- 9/9 declared CFG fixtures agree with the bounded path oracle;
- 100/100 exhaustive diamond products agree;
- 500/500 seeded randomized/metamorphic trials preserve agreement;
- nested branches and loops through three visits per block agree on every
  terminating oracle path;
- the 1,501-block/2,000-edge scale graph converges in 1,501 transfer
  evaluations without enumerating paths;
- omission of edge endpoints is detected after seven generated programs;
- Go race detection and vet pass;
- the `cfg` package has 93.6% statement coverage.

The semantic conclusion is narrow but consequential: a loan's inferred end is
not always attached to a source expression. At divergent control flow and loop
exit it can be an edge fact. The typed IR and machine diagnostics therefore
need stable edge identities even if ordinary source contains no explicit
loan-ending syntax.

The engineering conclusion is also useful: exhaustive paths belong in the
small oracle, not the production-shaped analyzer. A monotone finite-set
worklist answers this slice in polynomial space and can report its actual
transfer work.

## Known limitations

- Path evidence is exhaustive only within a bound of three visits per block;
  the analyzer itself is unbounded and reaches a set fixpoint.
- The spike models one static loan definition per CFG. It does not model fresh
  dynamic loan identities created on each loop iteration.
- Explicit source `end_loan` is rejected in this slice. A later surface policy
  may offer scoped escape hatches, but ordinary source should not require one.
- Loan use is first-order and local. Returned origin sets, callbacks,
  higher-ranked access, generic ability propagation, field-sensitive places,
  suspension, and FFI retention remain separate experiments.
- The independent path oracle reuses the prior spike's linear checker and
  normalizer. It is a different algorithm but not a formal proof or independent
  compiler.
- The scale observation is a regression baseline, not a ratified compiler
  latency budget or proof of asymptotic behavior for richer origin relations.

The follow-on [public origin and generic ability
spike](../003-public-origins-generic-abilities/README.md) now covers first-order
returned origins, one higher-ranked callback pattern, and generic abilities.
Field-sensitive disjointness, suspension, and FFI retention remain open.
