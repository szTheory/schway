---
phase: 05-native-equivalence-and-adversarial-evidence
recorded: 2026-09-06
code_head: 2383565
status: accepted
disposition: carried-from-mid-phase-gate
items: 1
blocking: 0
---

# Phase 05: Mid-Phase Gate Debt Register

Recorded by 05-09's mandatory mid-phase gate (D-05-40), following the Phase 3
and Phase 4 precedent of writing discovered debt down as a dated register
entry rather than leaving it in prose. Waves 1-3 are otherwise clean:
`sh scripts/verify-phase5.sh` is green, every Phase 5 control fires with
nonzero `RecomputedWork`, the six/one NAT-03 arithmetic holds, `discoverLoanLastUses`
is confirmed retired, and `git diff --stat` over every frozen prior-phase
corpus file, frozen foreign TU, and prior gate script reports no changes.

## Items

| ID | Source | Threat/Req | Severity | Landing phase | Item |
|---|---|---|---|---|---|
| D-05-41 | 05-09 `VerifyPhase5ControlsAndWork` (`lane:interpreter-o0-o3-lto`) | D-05-19 | info | Phase 5+ | The runtime gate's own `control:interpreter-o0-o3-lto` lane drives the interpreter/-O0/-O3/-O3-LTO differential over exactly one adversarial fixture (`inline_across_foreign.lang`), not the full six-fixture adversarial subset D-05-18a defines |

## Detail

### D-05-41 — the gate's own LTO lane samples one adversarial fixture, not all six

**Finding.** `VerifyPhase5ControlsAndWork`'s `phase5RunInterpreterO0O3LTOLane` (`internal/compiler/session/session_phase5.go`) proves `control:interpreter-o0-o3-lto` by running `testdata/phase5/inline_across_foreign.lang` through the interpreter and three native tiers (`-O0`, `-O3`, `-O3 -flto`) and comparing all four via `Phase5CompareEngines`. It does not also drive `dead_store_unused_acquire.lang`, `reorder_two_events.lang`, `tail_collapse_release_ladder.lang`, `typed_failure_truncated_stdout.lang`, or `defect_dies_by_signal.lang` through the same four-tier comparison inside this one gate function.

**Why this is not a correctness gap.** The full six-fixture adversarial subset, plus the enumerated closure, is already proven to agree across the interpreter, `-O0`, and `-O3` by `TestPhase5CorpusThreeEngineAgreement` (`session_phase5_corpus_test.go`, part of the ordinary `go test ./...`/`go test -race ./...` this gate also runs). Plan 05-06's own `TestLTOTierIsNotInert` independently proves the LTO tier is non-inert (distinct codegen between `-flto` and non-`-flto` builds) using its own fixture. `VerifyPhase5ControlsAndWork`'s own lane exists to prove `control:interpreter-o0-o3-lto` fires with real, non-zero work inside the gate's own control-and-work accounting — not to re-run the whole corpus at every tier, which would make `sh scripts/verify-phase5.sh`'s own runtime scale with the corpus size for a control the exhaustive `go test` suite already covers.

**What is real debt.** The one-fixture sample does not exercise `-flto`'s cross-TU inlining behavior against the other five adversarial shapes (dead-store elimination, event reordering, tail-collapse, truncated-stdout, and signal-terminated defect) inside the gate's own control-and-work JSON. A future toolchain regression that broke LTO agreement specifically for one of those five shapes (and not `inline_across_foreign.lang`) would still be caught by `TestPhase5CorpusThreeEngineAgreement` at `-O0`/`-O3`, but NOT by this gate's own `-O3-LTO` tier, since that test does not itself add an LTO tier.

**Phase 5+ fix.** Either (a) widen `phase5RunInterpreterO0O3LTOLane` to iterate `session.Phase5AdversarialFixtureFiles` at the `-O3-LTO` tier (cost: five more native builds and links per gate run), or (b) add an LTO tier to `TestPhase5CorpusThreeEngineAgreement`'s own adversarial-subset loop (moving the cost into the ordinary `go test` suite instead of the gate script's own runtime). Either closes the gap; neither is required before plans 05-10 through 05-14 begin, since both consume the settled interpreter/-O0/-O3 differential and the alias-fact/attribute-justification work this gate proves, not the LTO tier specifically.
