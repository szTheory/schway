# Differential Verification Harness

The method every spike used, and the one that found more real defects than any
single semantic result. This is cross-cutting: it describes how to build the
*evidence* for the subsystems in the other reference files.

## Requirements

From the `ownership-kernel` idea (MANIFEST.md):

- Keep the spike dependency-free and cheap to run locally or in CI.
- Treat disagreement and minimized counterexamples as useful results.
- Do not infer final source syntax or production performance from kernel
  notation.

## How to Build It

**Two independently written implementations, one fixture format.** A checker and
an oracle that share a transition function agree tautologically. Every spike
keeps derivation implementations separate and shares only inert versioned wire
types and fixtures. Spike 006 goes furthest: its oracle inlines every call into
one monolithic CFG and runs *naive round-robin* liveness — no worklist, no
predecessors, no summaries, no canonicalization — so it shares nothing with the
mechanism it prices.

**Give everything a stable machine-readable identity** — programs, source
operations, semantic events, loans, places, blocks, edges, contexts, functions,
bindings, origins, abilities, diagnostics. Two of spike 001's real defects were
findable only because identities were stable enough to compare.

**Compare ordered traces, not verdicts.** Result-only comparison hid a real
support defect in spike 001 (an `end_loan` before its borrow suppressed the
inferred endpoint) because the program already failed earlier. Spike 004
similarly requires the checker to recompute transitions rather than trust a
snapshot.

**Every harness carries a deliberate fault-injection path, armed from the CLI.**
`-inject-checker-bug`, `-inject-edge-bug`, `-inject-origin-bug`,
`-inject-summary-drop`, `-inject-stale-cache`. They exit unsuccessfully after
reporting the first disagreement — **the unsuccessful exit is the passing
result.** Each is independently mutation-killed.

**Breadth-first enumeration makes the first counterexample minimal** by operation
count within the declared alphabet. Spike 001 found a four-operation
counterexample after 159 generated programs; spike 002 found its missing edge
endpoint after seven.

**Make each corpus template a discriminator.** Some admission decision must flip
on the fact being derived, or a defect hides inside a semantically inert body.
Spike 006 designed its two summary bits backwards from the caller patterns that
refuse on them.

**State the comparison scope rather than assuming it.** Spike 006 unions its
oracle's per-context answers and then declares exactly where projection is
exact — conflicts always, local liveness only for loans born in the same
function, parameter loans deliberately excluded.

**Include at least one explicitly expected escape** in every trust-boundary
experiment. Spike 004's `coordinated-frontend-summary-lie` and spike 006's
`TestInertSummaryDefectEscapesByDesign` both fail if the corpus ever silently
starts covering them. Without an expected escape, a harness inflates its own
assurance claim by construction.

**Budget machine feedback.** Default CLI output is a compact JSON decision
summary — gate counts, measurements, elapsed time. Full traces materialize only
for a mismatch or an explicit detail flag. Spike 002's first working CLI
serialized ~24,000 tokens of path traces and liveness for one run.

**Separate semantic events from measurements.** Deterministic work counters
classify mechanisms; wall-clock nanoseconds live in a separate field and never
do. No timestamps or unstable map iteration reach semantic output.

**Record pivots and harness defects as findings** rather than deleting their
history. Three of the five most useful results in this spike series are
*harness* defects: the shallow-copy interface leak (003), the reused loan
identity (001), and the wrong placement of the interprocedural fact (006).

## What to Avoid

- **Sharing mutable backing storage across a boundary you claim is isolated.**
  Spike 003's exporter stripped bodies but shallow-copied return slices, so fault
  injection mutated the oracle. Deep-copy at the boundary; add a regression test
  that proves injection cannot reach the other side.
- **Letting an imported normalizer silently repair the fault you injected.**
  Spike 002 needed a deliberately late terminal guard for exactly this.
- **Quoting a cost number before the arm agrees with the oracle.**
- **Pricing only the weakest opponent.** Add a deliberately charitable variant of
  the mechanism you expect to reject.
- **Fitting growth against the structural size axis alone.** Fit against program
  operation count too, and classify on the operation-count fit — a corpus whose
  composition drifts with size moves the structural fit on its own.
- **Adding a dependency before it answers a specific experiment question better
  than a small local implementation.** An external parser, solver, IR, or test
  framework adds more assumptions than evidence at this stage.
- **Promoting an observation into a budget.** Every timing, size, and token count
  across all six spikes is explicitly a single-host observation.

## Constraints

- Go 1.24 standard library only. `go test`, `-race`, `-cover`, `go vet`, and
  seeded `testing/quick` were sufficient for all six spikes.
- Use task-local Go cache directories in restricted environments —
  `env GOCACHE=/private/tmp/ai-schway-spikeNNN-go-cache go test -count=1 ./...`.
  Spike 001 iteration 1 failed before compilation because the sandbox denied
  Go's default user cache.
- Independent implementations still share the Go compiler/runtime and this
  project's semantic specification. **Implementation diversity is not formal
  soundness**, and two separately written passes can still encode the same
  mistaken specification.

## Coverage achieved

| Spike | Package coverage |
|---|---|
| 001 | 89.7% |
| 002 | 93.6% |
| 006 | 92.1% |

## Origin

Synthesized from spikes: 001, 002, 003, 004, 005, 006
Source files available in: all of `sources/`
