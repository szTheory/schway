---
phase: 02-owned-values-and-abilities
recorded: 2026-09-04
code_head: 3399ddc
status: accepted
disposition: carried-to-phase-03
items: 9
blocking: 0
---

# Phase 02: Accepted Debt Register

Phase 02 closed with these items open, deliberately and with evidence. None is a
correctness defect in shipped behaviour: every one is control strength, cost, or
observability. Each was found by an independent gate, reproduced, and judged not
worth another fix wave — see "Why this was closed rather than fixed" below.

**Decision:** the developer chose to close Phase 02 recording these as dated debt
rather than run a seventh review wave (2026-09-04).

## Items

| ID | Source | Threat/Req | Severity | Item |
|---|---|---|---|---|
| D-02-01 | 02-REVIEW WR-01 | T-02-06 | warning | The repo-wide unbounded-spawn guard is a two-needle substring scan and is defeatable |
| D-02-02 | 02-REVIEW WR-02 | T-02-08 | warning | `evidence.canonical_unstable` never reaches a user; `evidence.ErrorCode` has no production call site |
| D-02-03 | 02-REVIEW WR-04 | T-02-01 | warning | Transitive loan sets made `check` Θ(N²) in body size |
| D-02-04 | 02-SECURITY SEC-02-B | T-02-06 | warning | `native.timeout` has no in-tree falsifier |
| D-02-05 | 02-SECURITY SEC-02-D | T-02-03 | info | Collision suffix `__LANG_` is a reserved identifier under C17 §7.1.3 |
| D-02-06 | 02-SECURITY SEC-02-E | T-02-06 | info | `MaxCLIStreamBytes` is 8 MiB against a measured 1,756 B high-water mark |
| D-02-07 | 02-REVIEW IN-02 | T-02-07 | info | The backend causality control matches one exact generated line |
| D-02-08 | 02-REVIEW IN-03 | — | info | `protocol.Human` returns unconverged output where `protocol.JSON` errors |
| D-02-09 | 02-VERIFICATION F3 | T-02-03 | info | `Box`/`Pair` programs pass `check` then fail spanless exit 3 on every engine |

## Detail

### D-02-01 — the unbounded-spawn guard is defeatable

`TestSourceNeverSpawnsUnboundedProcesses` (`internal/compiler/native/native_test.go`)
scans module sources for two literal spellings. The independent review defeated it
four ways; the most serious is `exec.CommandContext(context.Background(), ...)` with
plain `bytes.Buffer` streams, which passes the scan while having neither a deadline
nor a bound — precisely the hazard the guard is named for. `.Output()`,
`StdoutPipe()` + `io.ReadAll`, and an aliased constructor also evade it.

**Why not blocking:** every production spawn was hand-verified bounded at `3399ddc`
(clang compile, program run, and both evidence tool probes: deadline, independent
stdout/stderr writers at max-plus-one, no `CombinedOutput`). This is control
strength, not a live vulnerability.

**Phase 03 fix:** replace the substring scan with an AST/`go/analysis` pass that
resolves the spawn constructor and requires a non-background context plus bounded
writers at each call site.

### D-02-02 — the evidence canonical guard is observationally test-only

`evidence.Build` correctly fails closed with `evidence.canonical_unstable` when the
canonical reparse yields diagnostics or is not a formatter fixed point, and the
guard was mutation-proven (removing it produces a complete manifest bound to
corrupted bytes). But `evidence.ErrorCode` has zero production call sites, so the
CLI reports `evidence.operation_failed` and the distinct code is visible only to
tests.

**Why not blocking:** the fail-closed behaviour — the security property — is real
and covered. Only the diagnostic specificity is missing.

**Phase 03 fix:** route `evidence.ErrorCode` through the CLI error taxonomy so the
code surfaces, and add a CLI-level assertion.

### D-02-03 — checker cost is quadratic in body size

Making loan liveness transitive (`3d9493a`) propagates loan sets per binding, so
`check` is Θ(N²) in operation count. A 358 KB source burns ~19s (was ~7s) before
`check.work_limit` fires, while `recomputed_work` still reports linearly — so the
counted-work signal understates real cost.

**Why not blocking:** bounded by `MaxTokens` and the work limit; it fails closed,
it is slow. No unbounded path.

**Phase 03 fix:** this belongs with OWN-03 anyway. CFG-based liveness with a
union-find or interval representation removes the quadratic factor; make
`recomputed_work` count the propagation so the metric stops understating.

### D-02-04 — `native.timeout` has no in-tree falsifier

`TestNativeHelperProcess` implements flood modes only, no hang mode, so the compile
and run deadlines have no committed negative control. The mitigation was proven
effective out of tree twice (compile deadline fired at ~301 ms, run at ~305 ms,
both returning `native.timeout`).

**Why not blocking:** mitigation present and demonstrably effective; only the
falsifier is missing. ~40 lines to close.

**Phase 03 fix:** add a hang mode to the helper process and assert both deadlines.

### D-02-05 — `__LANG_` is a reserved identifier

The collision suffix at `internal/compiler/cgen/cgen.go` contains a double
underscore, which C17 §7.1.3 reserves to the implementation *anywhere* in an
identifier, not only at file scope.

**Why not blocking:** it appears only on an actual collision, clang accepts it
under `-std=c17 -Wall -Wextra -Werror -pedantic`, no golden contains it, and it
cannot collide with a libc or compiler internal.

**Phase 03 fix:** rename to `_LANG_` in a standalone commit, gated on the existing
collision-freedom suite. **Do this before any additional C artifact is frozen** —
the cost only grows.

### D-02-06 — the CLI stream ceiling is loose

`MaxCLIStreamBytes` is 8 MiB. Instrumenting all 50 test-support spawns measured a
maximum stdout of 1,756 B and stderr of 344 B, so the ceiling is ~4,800× the
high-water mark, and the "whole verify-corpus documents" rationale it was chosen
for does not hold — those documents are 1.7 KB.

**Why not blocking:** it is a genuine fail-closed ceiling on test-only code outside
the production trust boundary.

**Phase 03 fix:** tighten to `1 << 20` and cite the measured maximum in a comment.

### D-02-07 — the causality control is coupled to one literal line

`OwnedBackendMutationRunner` matches an exact generated line including a
source-derived identifier. It is correctly fail-closed (`strings.Count != 1` →
`native.backend_control_invalid`), but renaming the binding in
`testdata/phase2/owned_transfer.lang`, reindenting the emitter, or editing the
comment converts the phase's strongest control into an opaque operational failure.

**Phase 03 fix:** emit a stable generated marker (e.g. `/* lang:mutation-site:value */`)
and match on that, decoupling the seam from names and formatting.

### D-02-08 — `protocol.Human` diverges from `protocol.JSON`

`protocol.Human` returns unconverged output in a case where `protocol.JSON`
errors. Cosmetic today; a divergence between the two renderings is a latent
honesty gap.

### D-02-09 — `Box`/`Pair` check but cannot execute

`testdata/phase2/ability_shapes.lang` passes `lang check` at exit 0, and its core
is valid (`corevalidate` → `Valid:true, Checks:85`), but every engine and
`evidence` fail it as a spanless exit 3. Only `Byte` and `Buffer` lower to C or
run; `Box`/`Pair` exist for ability derivation only.

**Why not blocking:** outside the Phase 02 declared truth set, which requires only
that these shapes reach typed core — which they do. Pre-existing, not introduced by
the fix wave.

**Phase 03 fix:** this is the same taxonomy smell that made CR-01 a blocker — a
program that type-checks should not die without a span. Either lower `Box`/`Pair`
or reject them at check time with a causal "not executable in this phase"
diagnostic. Prefer the diagnostic; do not widen the C backend for it.

## Why this was closed rather than fixed

Phase 02 ran six review waves. Waves 5 and 6 each found blockers that the previous
wave's `clean` / `passed` verdict had missed, and wave 6's blocker was *introduced
by wave 5's own fix* — the CR-02 formatter repair regressed on any declaration
header carrying a trailing comment.

That pattern is the reason to stop editing. Each additional wave was finding
defects created by the previous wave's fixes, at a rate comparable to the defects
it removed. The nine items above are all control strength, cost, or observability;
none changes what a correct program does. The phase's actual guarantees — nine
fail-closed controls, causal native evidence proven by an O0/O3 source mutation,
independently re-derived ownership and ability oracles, byte-frozen Phase 1
goldens — are covered and were each mutation-tested at `3399ddc`.

## Process debt (carry into every later phase)

The three gate failures shared one shape: **a green test whose reachable input
space did not contain the hard case.** Coverage was not the problem; production
code was fully executed in every case.

- The syntax generator emitted only Phase 1 match programs, so the round-trip
  property could never reach the linear surface (missed CR-02).
- The generator then emitted comments only *before* `module` and *before* a
  binding, never trailing a header — the one placement the new brace classifier
  was sensitive to (missed the CR-02 regression, past 1000 generated cases).
- The ownership oracle encoded the *same law* as production, so the differential
  agreed on the wrong answer (missed WR-02).

Standing rules adopted for Phase 03:

1. A differential test is not evidence until reverting the production hunk makes it
   fail. Mutation-kill every oracle.
2. Ask what inputs a green property test actually reaches before trusting it.
3. Drive behaviour through the shipped binary on hand-written programs, not only
   through the gate's own corpus — the gate only ever sees what ships with it.

---

_Recorded: 2026-09-04 at `3399ddc`_
