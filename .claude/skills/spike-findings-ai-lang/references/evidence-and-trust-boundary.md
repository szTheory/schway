# Evidence and the Trust Boundary

Binding ownership evidence to a canonical typed-core subject, and being precise
about what an independent checker does and does not prove.

## Requirements

From the `ownership-kernel` idea (MANIFEST.md):

- Bind reusable ownership evidence to a canonical versioned typed-core subject.
  Use compact independent recomputation at package, cache, CI, and release
  boundaries; materialize full replay traces only on demand.
- Do not claim that a downstream certificate proves source-to-core correctness.
  A coordinated false typed-core statement remains inside the frontend trust
  boundary and must be attacked by different evidence.

## How to Build It

**Bind the statement, not just the proof.** The first invariant is subject
identity. A certificate carries a SHA-256 digest of the canonical typed-core
artifact. Without it, evidence can be replayed against a different module or a
changed artifact.

**Canonicalize only what is semantically unordered.** The digest sorts
declaration and set-like collections but **preserves event order, because event
order is semantic**, and includes the schema. Consequence: a declaration reorder
keeps the digest stable; a trace reorder changes it. 500 seeded
declaration/set reorderings preserved the digest; a semantic event reorder did
not.

**Default to compact recomputation, not replay.** Two encodings from the same
artifact:

| Encoding | Carries | Size in fixture / at scale |
|---|---|---|
| `recompute` | claimed summaries, abilities, final states, artifact digest | **242 bytes** |
| `replay` | all of the above plus full state after every event | ~1.0 MB for the scale trace (vs. ~675 KB typed-core artifact) |

**Both require the checker to implement the same small transition relation** —
blindly trusting a snapshot is not validation. So for this kernel, replay adds
**forensic detail rather than soundness**, at several orders of magnitude of
evidence size. Generate per-event snapshots only for a mismatch, a forensic
export, or a policy tier that explicitly requires replay.

**Share inert wire types, never inference.** Producer and verifier are separate
Go packages sharing only wire types. Their ability derivation and ownership
transition functions are written independently. (SpecTec's lesson: independent
implementations should share a small declarative contract, not hand-duplicated
prose, inference, and executable rules.)

**Run the validator at trust crossings, not after every edit.** Ordinary `check`
validates source directly and caches the result. The independent validator
belongs at package import, shared cache acceptance, CI/release evidence, and
possibly an explicit high-assurance `verify` lane.

**Call the artifact a validation manifest.** Not a proof.

## What to Avoid

- **Claiming source-to-core correctness.** The single deliberate escape in the
  mutation matrix is `coordinated-frontend-summary-lie`: change a return origin
  in the typed-core fact, the published summary, the digest, *and* the
  certificate together and the result is self-consistent, so the checker accepts
  it. Catching that needs an independently trusted source-to-core relationship,
  translation validation at a later lowering boundary, or a formal proof.
  **Including an expected escape in the matrix is the point** — it keeps the
  trust boundary represented honestly rather than letting the harness inflate its
  own assurance claim.
- **Rerunning the full compiler as "independent" evidence.** It shares the
  compiler's bugs and does not reduce the trusted implementation.
- **Building a second independent frontend for v1.** It duplicates the largest
  and fastest-changing part of the system. Use targeted translation and native
  differential checks instead (`native-lowering-ffi-contract.md`).
- **Formally verifying the checker now.** Premature proof and migration cost
  while the kernel is still changing. Preserve as a later assurance tier.
- **Treating SHA-256 as authorization.** It provides content binding — not signer
  identity, freshness, revocation, or authorization. Those belong to a signed
  evidence-envelope policy if a remote trust boundary later requires one.

## Constraints

- The typed-core artifact is a **hand-authored experiment format**, not the
  future production IR.
- The transition relation covers the validated first-order slice only — not CFG
  inference, variance, partial moves, async suspension, recursive origins, or
  native pointer provenance.
- **Ability rules are trusted typed-core declarations.** A coordinated false
  primitive rule is outside this checker's power.
- Independent code still shares the Go compiler/runtime and this project's
  semantic specification. **Implementation diversity is not formal soundness.**
- The verifier is 363 lines — much smaller than the combined experimental
  semantics — but that is not proof a production checker stays small as the
  language grows.
- Work grows linearly: a 10,001-event artifact produced ~20,004 checker
  operations (two simple passes over event identities and transitions).
- Wall-clock figures are single-host observations without isolated CPU or
  allocation distributions. They establish no release budget.

## Measured result (spike 004, PARTIAL)

Both encodings validate after JSON round-trip. **12 of 13 mutations detected**
with stable codes — stale and forged digests, omitted origins, borrowed-as-owned
and exclusive-as-shared summaries, a forged `copy` ability, false event and final
states, use after move, a missing loan end, duplicate event identity. Three
event-sequence failures reduce automatically to two-event counterexamples.
Repeated invalid runs emit identical diagnostics. The one escape is
`coordinated-frontend-summary-lie`, by design.

**Verdict: PARTIAL — keep the small validator, narrow the claim.**

## Origin

Synthesized from spikes: 004
Source files available in: `sources/004-independent-certificate-checker/`
