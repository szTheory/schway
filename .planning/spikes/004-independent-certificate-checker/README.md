---
spike: 004
idea: ownership-kernel
name: independent-certificate-checker
type: comparison
validates: "Given canonical typed-core ownership facts and an untrusted certificate, when a source- and body-blind checker validates summaries, abilities, and flow witnesses, then it catches distinct corruption classes without recreating frontend inference or taxing the ordinary edit loop"
verdict: PARTIAL
related: [001, 002, 003]
tags: [ownership, certificates, trusted-computing-base, validation, mutation, provenance]
---

# Spike 004: Independent certificate checker

## What this validates

This spike asks whether an independently implemented checker can shrink the
trusted ownership boundary. The checker receives only:

- a canonical, versioned typed-core artifact containing declared summaries,
  ability rules, and lowered ownership events; and
- an untrusted certificate bound to that artifact by a SHA-256 content digest.

It does not parse source, infer types, inspect function bodies, compute CFG
last use, or import the certificate producer. It independently checks public
return summaries, generic abilities, place/loan transitions, final states,
stable identities, and optional per-event witnesses.

The result is deliberately partial. This checker detects corruption or
dishonesty *after* typed-core construction. It cannot detect a coordinated
frontend error that emits a false typed-core statement and a matching
certificate. Claiming otherwise would confuse artifact integrity with compiler
correctness.

## Research

Proof-carrying code established the producer/consumer pattern: untrusted code
can carry a proof that a much smaller consumer checks against its own safety
policy. Lean similarly separates rich elaboration and tactics from kernel
checking, while its current validation guidance is explicit about what an
external checker still trusts. WebAssembly shows that a compact, deterministic
validation pass can run directly over a typed instruction stream. Translation
validation checks each concrete compiler result rather than attempting to
prove an evolving optimizer wholesale; Alive2 demonstrates both the practical
bug-finding value and the honest limits of a bounded validator.

SpecTec contributes a second lesson: manually maintaining multiple equivalent
semantic descriptions creates drift. Independent implementations should share
a small declarative contract, not duplicate prose, inference, and executable
rules by hand.

- [Necula and Lee, proof-carrying code](https://people.eecs.berkeley.edu/~necula/Papers/pcc_fmcs97.pdf)
- [Lean proof validation and trust boundary](https://lean-lang.org/doc/reference/latest/ValidatingProofs/)
- [WebAssembly validation algorithm](https://webassembly.github.io/spec/core/appendix/algorithm.html)
- [Pnueli et al., validation of optimizing compilers](https://cs.nyu.edu~/in_memoriam/pnueli/ZPL01.pdf)
- [Alive2 bounded translation validation](https://web.ist.utl.pt/nuno.lopes/pubs.php?id=alive2-pldi21)
- [CompCert architecture and verified validators](https://compcert.org/doc/)
- [SpecTec adoption and single-source semantics](https://webassembly.org/news/2025-03-27-spectec/)

| Approach | Strength | Cost or failure mode | Disposition |
|---|---|---|---|
| full replay certificate | localizes the first dishonest event and preserves a complete forensic trace | repeats derived state at every event; large artifact and extra comparison work | retain as on-demand failure evidence |
| compact claim plus recomputation | tiny certificate; checker recomputes a linear deterministic transition relation | checker must implement the small validation semantics | selected for package/cache/CI/release boundaries |
| rerun the full compiler | no additional certificate format | shares compiler bugs and does not reduce the trusted implementation | insufficient as independent assurance |
| independent frontend/compiler | can catch coordinated source-to-core defects | duplicates the largest and fastest-changing portion of the system | reject for v1; use targeted translation/native differential checks instead |
| formally verified checker | strongest story once the semantic contract stabilizes | premature proof and migration cost while the kernel is changing | preserve as a later assurance tier |

## Experiment shape

The producer and verifier are separate Go packages and share only inert wire
types. Their ability derivation and ownership transition functions are written
independently. The fixture contains three public functions, two generic ability
requests, two ownership traces, and seven events.

Two certificate encodings are generated from the same artifact:

1. `recompute` carries claimed summaries, abilities, final states, and the
   artifact digest. The verifier replays all semantic transitions itself.
2. `replay` additionally carries the full state after every event. The verifier
   still recomputes the state and compares each snapshot.

Canonical hashing sorts declaration and set-like order, but preserves event
order because event order is semantic. A declaration reorder therefore keeps
the digest stable; a trace reorder changes it.

## Gates

1. Both encodings accept the valid fixture after serialization.
2. Seeded false certificates and typed-core mutations cover binding, summary,
  ability, flow-witness, transition, and identity defects.
   The three event-sequence failures reduce automatically to two-event
   counterexamples.
3. The matrix includes a coordinated false typed-core statement that is
   expected to escape, proving the trust boundary is represented honestly.
4. Five hundred seeded declaration/set reorderings preserve the digest, while
   a semantic event reorder changes it.
5. Repeated invalid runs emit identical machine-readable diagnostics.
6. Work grows linearly over 101, 1,001, and 10,001 events.
7. The experiment compares artifact size, certificate size, check count, and
   observed p50/p95 validation time over 21 runs.
8. Certificates survive JSON round-trip without semantic change.

## How to run

From this directory:

```sh
env GOCACHE=/private/tmp/ai-lang-spike004-go-cache go test -count=1 ./...
env GOCACHE=/private/tmp/ai-lang-spike004-go-cache go test -race -count=1 ./...
env GOCACHE=/private/tmp/ai-lang-spike004-go-cache go test -coverpkg=./... ./...
env GOCACHE=/private/tmp/ai-lang-spike004-go-cache go vet ./...
env GOCACHE=/private/tmp/ai-lang-spike004-go-cache go run ./cmd/certificate-lab -pretty=false
```

The CLI emits one compact JSON report. Measurements are observations on the
current host, not product performance budgets.

## What to expect

- Both certificate modes validate.
- Twelve of thirteen mutations are detected with stable codes.
- The sole escape is `coordinated-frontend-summary-lie`, in which the typed-core
  fact, published summary, digest, and certificate all agree on the same false
  origin.
- A 10,001-event artifact produces about 20,004 checker operations.
- The compact certificate is 242 bytes in the current fixture.
- Full replay evidence is roughly 1.0 MB for the scale trace, compared with a
  roughly 675 KB typed-core artifact, and takes measurably longer to validate.

## Investigation trail

### Iteration 1 — bind the statement, not just the proof

The first invariant is subject identity. A certificate without a canonical
artifact digest can be replayed against another module or against a changed
typed-core artifact. The digest canonicalizes only semantically unordered
collections. This catches stale and substituted evidence but does not certify
that the frontend emitted a truthful statement.

### Iteration 2 — compare replay with recomputation

The replay encoding stores every derived place and loan state. The compact
encoding stores only final claims. Both require the checker to implement the
same small transition relation, because blindly trusting a snapshot is not
validation. For this kernel, replay therefore adds forensic detail rather than
soundness and amplifies evidence size by several orders of magnitude.

### Iteration 3 — mutate every represented boundary

The checker rejected stale and forged digests, omitted origins, borrowed-as-
owned and exclusive-as-shared summaries, a forged `copy` ability, false event
and final states, use after move, a missing loan end, and duplicate event
identity. These are distinct from merely running the same producer twice.

### Iteration 4 — include a mutation the checker must miss

Changing a return origin in the typed-core fact, public summary, digest, and
certificate together is self-consistent, so the checker accepts it. Detecting
that defect requires an independently trusted source-to-core relationship,
translation validation at a later lowering boundary, or a formal proof. Adding
source parsing and inference here would recreate the compiler and fail the
spike's size gate.

### Iteration 5 — test scale and stable representation

The verifier performs two simple passes over event identities and transitions.
The 10,001-event case required 20,004 counted checks. Full snapshots grew to
roughly 1,025 KB, while compact evidence remained 242 bytes. This supports
recomputation as the default and mismatch-only trace materialization.

## Results

**Verdict: PARTIAL. Keep the small validator, but narrow the claim and call the
artifact a validation manifest rather than implying a universal proof.**

The checker is useful when a consumer needs to establish that:

- a package, cache entry, release artifact, or agent handoff matches an exact
  typed-core subject;
- published ownership facts have not diverged from committed typed-core facts;
- simple generic abilities and lowered ownership traces obey the kernel; and
- diagnostics and evidence can be reproduced without rerunning the frontend.

It should not run as an expensive second frontend after every edit. Ordinary
`check` should validate source directly and cache the result. The independent
validator belongs at trust crossings—package import, shared cache acceptance,
CI/release evidence, and possibly an explicit high-assurance `verify` lane.

Per-event snapshots should be generated only for a mismatch, forensic export,
or a policy tier that explicitly requires replay. Compact recomputation is both
smaller and faster for the normal path.

The experiment does not justify claiming source-to-core correctness. The next
high-leverage boundary is native lowering and C ABI/provenance/cleanup, where
translation validation and interpreter/native differential evidence can catch
a class this certificate intentionally cannot.

## Known limitations

- The typed-core artifact is a hand-authored experiment format, not the future
  production IR.
- The transition relation covers the validated first-order slice, not CFG
  inference, variance, partial moves, async suspension, recursive origins, or
  native pointer provenance.
- SHA-256 provides content binding, not signer identity, freshness, revocation,
  or authorization. Those belong to a signed evidence-envelope policy if a
  remote trust boundary later requires them.
- Ability rules are trusted typed-core declarations. A coordinated false
  primitive rule is outside this checker's power.
- The independent code still shares the Go compiler/runtime and this project's
  semantic specification. Implementation diversity is not formal soundness.
- The 363-line verifier is much smaller than the combined experimental
  semantics built so far, but that comparison is not proof that a production
  checker remains small as the language grows.
- Wall-clock measurements are single-host observations without isolated CPU or
  allocation distributions. They establish no release budget.
