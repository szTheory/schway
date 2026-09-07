---
phase: 06-agent-feedback-and-performance-ratification
recorded: 2026-09-07
status: accepted
disposition: phase-close
items: 5
blocking: 0
---

# Phase 06: Phase-Close Debt Register

Recorded by plan 06-15, the phase's final-gate/ratification-close plan, following
the Phase 3/4/5 precedent of writing discovered and declared debt down as a dated
register entry rather than leaving it in prose scattered across fourteen plan
summaries. `sh scripts/verify-phase6.sh` is green, `go test ./...`, `go test -race
./...`, and `go vet ./...` are clean at HEAD, and every Phase 6 required control
fires with nonzero `RecomputedWork`.

This register exists because this project's own house style is to record what a
phase did NOT close, not to quietly drop it. Every item below was surfaced
deliberately during Phase 6's fourteen prior plans or during this plan's own close;
none is a new discovery this plan made by accident.

## Items

| ID | Source | Threat/Req | Severity | Landing phase | Item |
|---|---|---|---|---|---|
| D-06-13 | 06-04 (D-06-13, CONTEXT.md) | DX-03 | info | Phase 6 (declared, not closed) | The artifact cache's soundness argument has four named holes: undeclared environment, a Clang version-string-stable change, future nondeterministic codegen, and a hand-edited/partially-deleted cache directory |
| D-06-20 | 06-10 (D-06-20, CONTEXT.md) | QLT-02 | info | Deferred, revisit with a second (Linux) machine | Peak RSS stays `"unavailable"` for M001; `getrusage` is not implemented |
| D-06-29 | 06-11/06-15 (D-06-29, CONTEXT.md) | DX-04 | info | Phase 6 (declared, not closed) | The held-out defect-corpus split blunts, but cannot eliminate, automated-program-repair overfitting risk given how small this language is |
| D-06-33 | 06-02/06-04/06-05/06-06/06-07/06-08/06-09/06-10/06-11/06-12/06-13/06-14/06-15 | DX-02, DX-03, DX-04, QLT-02 | info | Carried to a future verification pass | Four unresolved `unclassified`-category edge probes from the phase coverage report, never resolved across the whole phase |
| D-06-30 | 06-11 (D-06-30, CONTEXT.md) | DX-04 | info | Non-gating, per-milestone/on-demand cadence | The recorded agent-legibility exercise measuring `repair_rounds` against the `p95: 2` intent is evidence, not a gate, and has not itself been run as part of this plan's close |

## Detail

### D-06-13 — the artifact cache's soundness argument has four named holes

**Finding.** The Phase 6 artifact cache (`internal/compiler/cache`, D-06-06 through
D-06-12) makes an *artifact*-cache argument, not a test-selection argument:
identical declared inputs imply an identical intermediate artifact. That argument
has four holes, named at design time (D-06-13, `06-CONTEXT.md`) and now declared,
gate-visible, and mechanically asserted never-detected by this plan's escape
register (`session.Phase6ExpectedEscapes()`, `session_phase6_escapes.go`):

1. **`escape:cache-undeclared-environment`** — undeclared environment (locale,
   `ulimit`, filesystem case-sensitivity) is not on `cache.ArtifactSpec`'s declared-
   input list and cannot silently cause a cache hit to diverge from what a cold
   build would have produced.
2. **`escape:cache-clang-version-string-stable`** — a Clang change that does not
   alter its reported version string (the ccache `__TIME__`-class footgun) is
   indistinguishable from an unchanged toolchain at the cache-key layer.
3. **`escape:cache-nondeterministic-codegen`** — any future nondeterministic
   codegen would silently break the cache's own "same declared inputs implies same
   artifact" premise; the cache has no independent way to detect this.
4. **`escape:cache-directory-hand-edited`** — a hand-edited or partially deleted
   cache directory is indistinguishable from a genuinely cold one; there is no
   integrity check beyond content-hash lookup, consistent with the standing
   project stance that SHA-256 is content identity, never proof.

**Why this is not closed.** Declaring an escape and asserting it is never
mis-presented as a solved control (`TestPhase6EscapesAreNeverPresentedAsControls`)
is real evidence — strictly better than a silent gap — but it does not eliminate
any of the four holes. Each remains a genuine way a cache-artifact-reused run could
diverge from what a fresh build would have produced, undetected by any control in
this repository.

**Not a target for Phase 6+ closure.** D-06-13's own text calls this list
"anything not on this list is an escape by construction" — closing it would mean
either an exhaustive environment fingerprint (against PROJECT.md's anti-host-
fingerprint stance) or abandoning local caching altogether. Recorded as accepted,
permanent residual debt, not a scheduled fix.

### D-06-20 — Peak RSS stays unavailable for M001

**Decision (D-06-20, `06-CONTEXT.md`).** `getrusage` is not implemented for M001.
`ru_maxrss` is bytes on macOS and kilobytes on Linux, and a Go process's RSS is
dominated by runtime/GC allocation unrelated to compiler work; implementing it on
the only available host (a single Apple-silicon laptop) would trade an honest gap
for exactly the host-encoding trap PROJECT.md warns against.

**Where the gap is visible.** `protocol.PeakRSSUnavailable` is the sole value
every `peak_rss_status` producer site emits, across all six phases' verify gates
(carried forward from Phase 5, extended unchanged through Phase 6 —
`TestNoGetrusageAnywhere`/`TestPeakRSSStaysUnavailable`, added by 06-10, make the
gap mechanically enforced rather than merely documented).

**Revisit condition.** A second, Linux, machine, to validate the bytes-vs-
kilobytes unit conversion and to separate Go-runtime RSS from compiler work before
this becomes a real measurement rather than a host-specific guess.

### D-06-29 — held-out corpus discipline blunts, does not eliminate, repair overfitting risk

**Finding (D-06-29, `06-CONTEXT.md`).** DX-04's defect-injection fixtures
(`testdata/phase6/heldout_*.lang`) used by the CI gate are held disjoint from the
`derivation_*.lang` fixtures anyone may consult when hand-deriving the repair
driver's kind-to-edit mapping (`testdata/phase6/README`, enforced by 06-13's own
test that no `derivation_*` reference ever names a `heldout_*` path).

**Why this is not closed.** This split blunts the automated-program-repair
overfitting failure mode named in the automated-program-repair literature (Qi et
al. 2015; Smith et al. 2015) and cited by D-06-23 — it does not eliminate it, given
how small this language is. A driver with three defect classes and six fixtures
total has a materially smaller reachable-input space than a production language's
corpus would; a driver that happens to generalize correctly across six held-out
fixtures is not proof it generalizes across the space of programs this language
can express. Declared here as the honest residual: `escape:repair-heldout-corpus-
residual-overfitting` (`session.Phase6ExpectedEscapes()`).

### D-06-33 — four unresolved `unclassified`-category edge probes, carried across the whole phase

**Finding.** The phase's own coverage report flagged one edge probe per gray area
as category `unclassified — review manually`, and none of the four was ever
resolved across all fourteen prior plans plus this plan's own close:

1. **DX-02 (`explain`/`query` addressing).** Surfaced 06-02, carried through
   06-03. Reviewer question: is there an addressing or bounding edge case for
   `explain`/`query` that the depth/node-budget/empty/ordering criteria already
   covered do not?
2. **DX-03 (change-risk selection and cache).** Surfaced 06-04, carried through
   06-05/06-06/06-07. Reviewer question: is there a cache or lane-selection edge
   case beyond adjacency/empty/ordering that the shipped selector does not cover?
3. **DX-04 (repair exercise).** Surfaced 06-11, carried through 06-12/06-13/06-14.
   Reviewer question: is there a repair-shape edge case beyond incomplete-
   `MachineApplicable` and empty-applicability that eligibility should refuse?
4. **QLT-02 (budgets and ratification).** Surfaced 06-08, carried through
   06-09/06-10. Reviewer question: is there a ratification edge case beyond
   empty/short sample sets and undeterminable CoV that the shipped ratification
   path should refuse?

**Why this is not closed.** Each plan that carried one of these forward
deliberately did not attempt to resolve it — the plan's own stated scope did not
cover it, and resolving a coverage-report `unclassified` item requires a human
reviewer's judgment call about what edge case the automated coverage report could
not itself name. This plan (06-15), the phase's designated final review pass, also
does not resolve them: no new edge case surfaced during this plan's own execution
that either confirms or refutes any of the four reviewer questions above. They are
recorded here, honestly open, rather than silently closed by omission.

**Disposition.** Carried to a future verification pass (a milestone audit, or a
future phase revisiting one of these four areas) as four standing reviewer
questions, not as blocking work.

### D-06-30 — the recorded agent-legibility exercise is non-gating and has not been run as part of this plan's close

**Decision (D-06-23/D-06-30, `06-CONTEXT.md`).** Phase 6's answer to "is the
repair protocol agent-legible" is two-part: a deterministic CI gate (proven by
`go test ./...`, `TestProseScrambleLeavesRepairBehaviourIdentical`,
`TestVocabularyRemovalDrivesTheDriverRed`, `TestVocabularyRemovalGuardIsNotInert`,
and the per-class injector marker-mutation-kill tests) plus a SEPARATE, non-gating,
recorded exercise that measures whether an actual consumer who did not author the
diagnostics can repair real defects using only `--json` protocol output.

**What this plan's gate proves, and what it does not.** `sh scripts/verify-
phase6.sh` and `go test ./...` prove the repair MECHANISM depends on the
structured channel (D-06-27's three anti-theater guards) and that the driver's
import boundary is enforced (D-06-28). They do NOT prove an LLM, or any consumer
who did not author the diagnostics, finds the diagnostics legible in practice —
that is D-06-30's separate, explicitly non-gating exercise: repair rounds against
the `agent.repair_rounds.p95: 2` intent from `wiki/compute-efficiency-
constitution.md`, plus token and tool-call cost, and whether protocol-only access
sufficed.

**Status at phase close.** This exercise has NOT been run as part of this plan.
It is evidence, not a gate, explicitly scheduled at a per-milestone or on-demand
cadence (D-06-30) — never as part of an individual phase's own close. Recording
this plainly here so a green `sh scripts/verify-phase6.sh` is never later mistaken
for proof that the diagnostics are agent-legible; it proves only that the
mechanism is structurally sound.

## Named residual limitations (accepted, Phase 6 close)

Recorded as accepted residuals rather than open work, following 05-DEBT.md's own
precedent — demonstrating an escape or measuring a boundary is real evidence, but
it does not close the underlying hazard:

1. **The artifact cache's soundness argument has four permanent holes**
   (D-06-13, above) — declared, gate-visible, and asserted never-detected.
2. **Peak RSS is permanently unavailable for M001** (D-06-20, above) — a
   deliberate, mechanically-enforced gap, not neglect.
3. **The repair exercise's held-out corpus discipline blunts, not eliminates,
   overfitting risk** (D-06-29, above) — six fixtures across three classes is a
   materially smaller reachable space than production scale.
4. **Wall-clock cold/warm distributions are ratified as bounded observations,
   never as hard p95 gates** (D-06-15) — this project has exactly one laptop-class
   host with no CI fleet; there is no cross-machine averaging and no prior
   distribution from which to set a credible statistical threshold.
5. **The recorded agent-legibility exercise is evidence, not proof** (D-06-30,
   above) — a green CI gate proves the mechanism is sound, never that the
   diagnostics are legible to an LLM or any other consumer who did not author them.

## Carry-forward from prior phases (unchanged, restated for completeness)

- **D-03-02 remains open past the milestone** (carried through 04-DEBT.md and
  05-DEBT.md, restated in 05-DEBT.md's D-05-42 detail): an exported borrow-derived
  return with no declared origin exports indistinguishable from a fully-owned
  return, in the INTERPROCEDURAL half of the hazard (the single-function half was
  closed in Phase 3). Phase 6 is agent feedback and performance ratification, not
  language surface — this remains M002's concern (`OpCall`'s lead charter item,
  D-05-32/D-05-33), untouched by this phase.
- **D-05-41** (the Phase 5 gate's own LTO lane samples one adversarial fixture, not
  the full six-fixture adversarial subset) is unrelated in cause and scope to
  anything in Phase 6 and is not touched by this phase's work.

## Closure — carried debt closed by Phase 6

- **The Phase 5 blocker "Baseline machines for ratified feedback budgets remain to
  be chosen before Phase 6" (STATE.md) is CLOSED by 06-09.** D-06-16/D-06-17/
  D-06-18/D-06-22 implemented a declared-machine budget manifest
  (`session/qlt02_budget_manifest.json`), an executable audit
  (`AuditQLT02BudgetManifest`), observation-only mode for undeclared machines, and
  the exact `recomputed_work`-only blocking rule wired into `protocol.Lane`.
