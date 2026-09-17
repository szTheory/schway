# Spike Conventions

Patterns and stack choices established for the ownership-kernel idea. New
spikes follow these unless the question requires otherwise.

## Stack

- Use Go 1.24 and the standard library for the first semantic workbenches. This
  is an experiment-host choice, not the final compiler-host decision.
- Add no dependency until it answers a specific experiment question better
  than a small local implementation.

## Structure

- Keep versioned JSON fixtures separate from implementations.
- Give every program, source operation, semantic event, loan, place, and
  diagnostic a stable machine-readable identity.
- Keep the dynamic oracle and static checker transition implementations
  separate so agreement is not purely tautological.

## Patterns

- Breadth-first enumeration makes the first discovered sequence minimal by
  operation count within the declared alphabet.
- Include a deliberate fault-injection path proving that the harness detects
  semantic drift.
- Record pivots and harness defects as findings rather than deleting their
  history.
- Separate semantic events from measurement fields such as elapsed wall time.
- Preserve CFG block and edge identities in semantic evidence; a derived fact
  may belong to an edge even when no source expression spells it.
- Keep exhaustive path expansion in bounded test oracles. Production-shaped
  analyzers use finite monotone dataflow and report transfer work directly.
- Default CLI output is a compact decision summary. Retain full traces for a
  mismatch or an explicit detail request rather than charging every healthy
  iteration for them.
- Cross-package semantic experiments serialize and deserialize the public
  interface before consumer checking. Implementation facts and exported
  summaries must not share mutable backing storage.
- Public ownership summaries name parameter/field origins and return
  alternatives. Access mode and type abilities remain separate facts.
- Representation comparisons render from one normalized contract and report
  shape/count evidence without promoting token count into a usability verdict.
- Independent validators share inert versioned wire types with producers, not
  inference, transition, or derivation implementations.
- Content digests canonicalize only semantically unordered declarations and
  sets. They preserve control-flow or event order and include the schema.
- Every trust-boundary experiment includes at least one explicitly expected
  escape so the harness cannot inflate its assurance claim by construction.
- Native spikes compare a separately encoded semantic oracle against both
  unoptimized and optimized artifacts, then add sanitizers as an orthogonal
  lane rather than treating one execution mode as sufficient.
- External tool invocations record exact toolchain, target, flags, content
  digest, compile time, and artifact size; temporary outputs stay outside the
  source tree and are removed after the run.
- Cost comparisons price at least one deliberately charitable variant of the
  mechanism a spike expects to reject. A headline that only beats the weakest
  possible opponent is not a finding.
- Growth is fitted against program operation count as well as against the
  structural size axis, and classified on the operation-count fit. A corpus
  whose composition drifts with size moves the structural fit on its own.
- Cost arms are compared only after every arm has been shown to agree with an
  independent oracle on the same corpus.
- Every corpus template used in a cost sweep is a discriminator: some admission
  decision must flip on the fact being derived, so a defect cannot hide inside
  a semantically inert body.
- A cache is priced on both sides -- what it saves and what one upstream edit
  forces it to throw away.
- Agreement is compared over ordered semantic traces, not over verdicts and
  error codes. A result-only comparison hides a support defect whenever the
  program already fails on an earlier operation.
- Where an oracle is more precise than the mechanism it checks, the projection
  scope is written down rather than assumed: which facts project exactly, and
  which are deliberately excluded because a body-blind analyzer cannot see them.
- An armed fault-injection run exits unsuccessfully on purpose. The unsuccessful
  exit is the passing result, and each seam is independently mutation-killed.

## Tools and libraries

- `go test`, its race detector, and built-in coverage are sufficient for the
  first workbench.
- Use task-local Go cache directories in restricted environments.
- The Go standard library's seeded `testing/quick` is sufficient for early
  randomized/metamorphic properties without adding a fuzzing dependency.
