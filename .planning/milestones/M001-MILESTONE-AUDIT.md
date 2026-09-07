---
milestone: M001
milestone_name: Source-to-Native Semantic Spine
audited: 2026-09-07T00:00:00Z
auditor: gsd-audit-milestone (orchestrator inline + gsd-integration-checker)
status: tech_debt
scores:
  requirements: 28/28
  phases: 6/6
  integration: 6/6
  flows: 13/13
gaps:
  requirements: []
  integration: []
  flows: []
nyquist:
  compliant_phases: [01, 02, 04]
  partial_phases: []
  not_validated_phases: [03, 05, 06]
  missing_phases: []
  overall: partial
tech_debt:
  - phase: 02-owned-values-and-abilities
    items:
      - "D-02-01 (warning): repo-wide unbounded-spawn guard is a two-needle substring scan and is defeatable"
      - "D-02-02 (warning): evidence.canonical_unstable never reaches a user; evidence.ErrorCode has no production call site"
      - "D-02-03 (warning): transitive loan sets made check and corevalidate O(N^2) in body size"
      - "D-02-04 (warning): native.timeout has no in-tree falsifier"
      - "D-02-05 (info): collision suffix __LANG_ is a reserved identifier under C17 7.1.3"
      - "D-02-06 (info): MaxCLIStreamBytes is 8 MiB against a measured 1,756 B high-water mark"
      - "D-02-07 (info): backend causality control matches one exact generated line"
      - "D-02-08 (info): protocol.Human returns unconverged output where protocol.JSON errors"
      - "D-02-09 (info): Box/Pair programs pass check then fail spanless exit 3 on every engine"
  - phase: 03-borrowed-views-and-cfg-lifetimes
    items:
      - "D-03-02 (warning, OPEN PAST MILESTONE): an exported borrow-derived return with no declared origin exports indistinguishable from a fully-owned return, in the INTERPROCEDURAL half of the hazard. Single-function half closed in Phase 3. Owned by M002's OpCall charter (D-05-32/D-05-33)."
  - phase: 04-fallible-resources-and-c-boundary
    items:
      - "D-04-30 (info, M002): storable/matchable Result value and payload-carrying alternatives deferred"
      - "D-04-31 (warning, inherited): four accepted residual limitations — the coordinated three-way lie, two nonlocal-exit blind spots, the single-host single-record-shape fence, and the permanence of quarantine"
  - phase: 05-native-equivalence-and-adversarial-evidence
    items:
      - "D-05-41 (info): the runtime gate's own control:interpreter-o0-o3-lto lane drives the differential over one adversarial fixture (inline_across_foreign.lang), not the full six-fixture subset D-05-18a defines"
      - "D-05-42 (warning, M002): OpCall, interprocedural loan liveness in both admission layers, call-graph construction, cycle refusal, bounded interpreter call stack, and the cross-function rebuild of Phase 3's differentials are OUT of M001"
  - phase: 06-agent-feedback-and-performance-ratification
    items:
      - "D-06-13 (info, declared not closed): the artifact cache's soundness argument has four named holes — undeclared environment, a Clang version-string-stable change, future nondeterministic codegen, a hand-edited cache directory. Asserted never-detected by session.Phase6ExpectedEscapes()."
      - "D-06-20 (info): peak RSS stays unavailable for M001; getrusage not implemented (TestNoGetrusageAnywhere enforces this)"
      - "D-06-29 (info, declared not closed): held-out defect-corpus split blunts but cannot eliminate automated-program-repair overfitting risk"
      - "D-06-30 (info, non-gating): the agent-legibility exercise (repair_rounds vs p95:2) has not been run as part of phase close"
      - "D-06-33 (info, carried): four unresolved unclassified-category edge probes from the phase coverage report, never resolved across the whole phase"
      - "WR-01 (accepted, 06-REVIEW): D-06-19's cold-sample distribution is declared but no production path collects it; Samples.Summary() refuses sets shorter than 20. Affects SC4's literal wording — the warm half is real and gated, the cold half is absent."
      - "WR-02 (accepted, 06-REVIEW): query cursor pagination uses strict-inequality tie-breaking that would drop ties if any resolver produced them. Not reachable today."
  - phase: cross-cutting (planning metadata)
    items:
      - "Nyquist: 03, 05, 06 have VALIDATION.md that validate-phase never reconciled (status planned/draft/draft). Coverage TODO per GSD #2117, not a compliance failure."
      - "01-VALIDATION.md carries the non-canonical status value 'complete' rather than 'validated'."
      - "All six 02-*-SUMMARY.md files predate the requirements-completed frontmatter field, so OWN-01/OWN-02 have no machine-readable SUMMARY linkage (both are ✓ SATISFIED in 02-VERIFICATION.md:349-350)."
      - "06-15-SUMMARY.md lacks requirements-completed frontmatter."
      - "01-VERIFICATION.md states requirement coverage in range notation ('FND-01..03, SYN-01..04, SEM-01..02') rather than per-ID rows, which defeats per-ID traceability tooling."
      - "PROJECT.md's six 'Active' outcome checkboxes are all still unchecked despite the phases having delivered them."
---

# Milestone Audit: M001 — Source-to-Native Semantic Spine

**Verdict: all requirements satisfied, integration sound, no blockers. Accumulated
accepted tech debt warrants a review decision before archiving.**

M001's own Definition of Done is stricter than a green build:

> M001 is complete only when every requirement is implemented, independently
> verified, and represented by reproducible evidence from a clean checkout. A
> green build without interpreter/native equivalence, preserved negative controls,
> or bounded structured diagnostics is not sufficient.

All three clauses hold. Evidence below.

## 1. Requirements Coverage (3-source cross-reference)

**28/28 satisfied.** Every REQ-ID in `REQUIREMENTS.md`'s traceability table
(lines 131–158) reads `Complete`, and every one is verified by a phase whose
`VERIFICATION.md` reads `status: passed`.

| Phase | REQ-IDs | VERIFICATION | Score |
|-------|---------|--------------|-------|
| 1. Canonical Pure Spine | FND-01, FND-02, FND-03, SYN-01, SYN-02, SYN-03, SYN-04, SEM-01, SEM-02, INT-01, NAT-01, DX-01 | passed | goal-backward, 10/10 truths |
| 2. Owned Values and Abilities | OWN-01, OWN-02 | passed | 14/14 must-haves |
| 3. Borrowed Views and CFG Lifetimes | OWN-03, OWN-04 | passed | 4/4 must-haves |
| 4. Fallible Resources and C Boundary | SEM-03, RES-01, FFI-01 | passed | 8/8 must-haves |
| 5. Native Equivalence and Adversarial Evidence | INT-02, NAT-02, NAT-03, QLT-01 | passed | 4/4 must-haves |
| 6. Agent Feedback and Performance Ratification | FND-04, DX-02, DX-03, DX-04, QLT-02 | passed | 4/4 success criteria |

### Two false positives resolved by manual reading

A mechanical per-ID scan raises two flags that are **not** real gaps. Both are
recorded as planning-metadata debt above so tooling stops re-raising them.

**Six requirements appear orphaned but are not.** A literal grep for FND-02,
FND-03, SYN-02, SYN-03, SYN-04, and SEM-02 across all `*-VERIFICATION.md` files
returns nothing. `01-VERIFICATION.md:38` covers them in range notation —
"FND-01..03, SYN-01..04, SEM-01..02, INT-01, NAT-01, and DX-01 are covered" —
which per-ID matching cannot see. Confirmed by reading the file. Not orphaned.

**OWN-01/OWN-02 appear `partial` but are not.** No Phase 2 SUMMARY carries the
`requirements-completed` frontmatter field, which postdates Phase 2 — all six
`02-*-SUMMARY.md` files lack it. The status matrix would grade this
`partial (verify manually)`; manual verification resolves it.
`02-VERIFICATION.md:349-350` marks both **✓ SATISFIED** with itemized evidence
(exactly-once affine enforcement, four named ownership rejections at exit 2 with
spans including transitive reborrow, interpreter/O0/O3 parity, an exact-one-site
emitted-C mutation forcing exit 4 at both optimization levels; and for OWN-02 an
exhaustive 32-leaf/1024-pair oracle plus a spy-combiner non-tautology test).

OWN-02 carries one accepted override (`02-OVERRIDES.md` OV-02-01: `Buffer` grants
`share`), judged satisfied because the derivation mechanism is untouched and the
axiom change moved exactly the two golden fields it should have.

**Orphan detection: none.** Every REQ-ID in the traceability table is verified by
at least one phase.

## 2. Phase Verification

**6/6 phases verified, 62/62 plans executed.** No phase is missing a
`VERIFICATION.md`.

Three phases reached `passed` only after a re-verification round that overturned
an earlier verdict — worth recording, because it is evidence the verification
process has teeth rather than a sign of instability:

- **Phase 2** re-verified at `da75e95`. The prior `passed` verdict at `f1206bb`
  was wrong: an independent deep code review found two source-reachable critical
  defects under truths the earlier run marked VERIFIED, plus a third latent defect
  its differential could not see *because the "independent" oracle encoded the same
  wrong law*. All three were reproduced at `f1206bb` through the shipped CLI and
  mutation-killed at HEAD.
- **Phase 3** re-verified from `gaps_found` 3/4 → `passed` 4/4 after closing a
  multi-arm origin leak, reproduced live against the shipped binary with a
  hand-written out-of-corpus program.
- **Phase 4** re-verified from `gaps_found` 7/8 → `passed` 8/8 after plan 04-13
  closed the `core.ForeignContract` C-injection class.

## 3. Cross-Phase Integration

**Sound. No gaps.** Checked by `gsd-integration-checker` against source under
`internal/` and `cmd/`, not against the planning docs.

Pipeline composition verified at every boundary:

| Phase | Contribution | Wired into |
|-------|--------------|-----------|
| 1 | `syntax.Parse` → `check.Check` → `interp.Run` / `cgen.EmitNative` | 154+ refs in `session.go` |
| 2 | `ability.Derive` / `DeriveSealed` at the checker; execution facts | `check.go` |
| 3 | `core.Block`, `core.Edge`, `core.LoanEndpoint` | 317+ refs across check, corevalidate, interp, cgen, originvalidate |
| 4 | `OpForeignCall`, `OpFail`, `ForeignContract`, `native.ForeignSources` | all six required dispatch sites; 27+ refs in `session.go` |
| 5 | `selectsByPointerLowering`, `emitLinearBorrowedByPointer` | `cgen.go` |
| 6 | `protocol.Schema1`, `protocol.LaneSchema1` | `main.go`, `session_phase6.go`, 16+ lane sites |

**Phase 6's coordinated schema bump landed whole.** `protocol.New()` returns
`lang.command/1`; every `Lane` literal in session code uses `lang.verify-lane/1`;
no production code still references the `/0` constants; and the backward-compat
constants are preserved so pre-Phase-6 documents stay reproducible
(`protocol.go:19-39`, `:449-450`). Phases 1–5 are unaffected.

**13/13 E2E flows reachable through the shipped binary** — `format`, `check`,
`run --engine=interpreter`, `run --engine=native`, `evidence`, `verify` for the
phase-4/5/6 corpora, `interface export`, `interface core`, `interface check`,
`explain`, `query`. `cmd/lang/main.go` routes 14 commands through
`internal/compiler/session`.

**No test-only deliverables.** Every phase's output reaches the shipped binary;
no orphaned exports at any phase boundary.

## 4. Reproducible Evidence from a Clean Checkout

The DoD's evidence clause was re-established live during this audit rather than
taken from the record:

- Working tree clean; `go build ./...` succeeds.
- `sh scripts/verify-phase6.sh` — the same gate CI's `phase-gate` job runs on both
  ubuntu-latest and macos-latest — **exited 0**. It rebuilds the binary and
  re-verifies all six corpora, proving every prior phase's non-regression with
  *this* phase's binary. Every Go package `ok`. All six corpora returned
  `"status":"pass"` with every declared escape present. `lang stats` sampled 20
  warm runs at p50 166ms / p95 177ms, CoV 0.046.
- Phase 6 UAT: 68/68 deliverables pass, 0 issues, 0 gaps, all deterministically
  covered by automated tests.
- `06-SECURITY.md`: `status: verified`, `threats_open: 0`.

Interpreter/native equivalence, preserved negative controls, and bounded
structured diagnostics are all directly exercised by that gate.

## 5. Nyquist Coverage — the one soft spot

| Phase | VALIDATION.md status | nyquist_compliant | Classification | Action |
|-------|---------------------|-------------------|----------------|--------|
| 01 | `complete` | true | Compliant (non-canonical status value) | — |
| 02 | `validated` | true | Compliant | — |
| 03 | `planned` | false | **Not-validated** | `/gsd-validate-phase 3` |
| 04 | `validated` | true | Compliant | — |
| 05 | `draft` | false | **Not-validated** | `/gsd-validate-phase 5` |
| 06 | `draft` | false | **Not-validated** | `/gsd-validate-phase 6` |

Per GSD #2117, `status: planned` / `draft` means validate-phase never reconciled
the file, so `nyquist_compliant: false` is **not authoritative** — these are
coverage TODOs, not compliance failures. Only `validated` + `false` would be a
genuine PARTIAL, and no phase is in that state. Re-running validate-phase on 3,
5, and 6 would yield the real verdict.

This is discovery only. Nothing was auto-run.

## 6. Tech Debt Ledger

**~25 open items across 6 phases.** Every one is documented, dispositioned, and
accepted in a `*-DEBT.md`; none is an undiscovered defect. Severity skews
info/warning — no blockers.

### Closed during the milestone

Later phases retired earlier debt rather than letting it accumulate silently:

- **D-03-01** (two liveness derivations decide different things) — closed by 05-02.
- **D-04-26** (`discoverLoanLastUses` retirement) — closed by 05-02.
- **D-04-33** (`core.ForeignContract.Alias` outside the field audit) — closed by 05-01.
- **Phase 5's "baseline machines for ratified budgets"** blocker — closed by 06-09
  (declared-machine budget manifest, executable audit, observation-only mode,
  `recomputed_work`-only blocking rule).

### The one item that outlives the milestone by design

**D-03-02** — an exported borrow-derived return with no declared origin exports
indistinguishable from a fully-owned return, in the *interprocedural* half of the
hazard. The single-function half was closed in Phase 3.

This is an **owned, scheduled deferral, not an unowned carry.** ROADMAP.md's M002
Charter names `OpCall` as M002's lead item and states the consequence plainly:
**M001 ships without Lang-to-Lang calls**, so interprocedural `-O3` equivalence is
outside M001's proof scope *by construction*. The stated reason for deferring is
that pulling `OpCall` into Phase 5 would have landed a new `OperationKind` at six
dispatch sites simultaneously with alias-fact emission, the first sanitizer lanes,
the first reducer, and the QLT-01 registry — the exact fingerprint of the failure
that cost Phase 2 a remediation round, Phase 3 a mid-phase gate plus three
gap-closure plans, and Phase 4 thirteen plans and five review rounds.

### Phase 2's nine items and why they stayed open

Phase 2 ran six review waves; waves 5 and 6 each found blockers the previous
wave's clean verdict missed, and wave 6's blocker was *introduced by wave 5's own
fix*. The phase stopped editing because each wave was creating defects at a rate
comparable to those it removed. All nine remaining items are control strength,
cost, or observability — **none changes what a correct program does**.

That phase also produced three standing process rules now carried forward: mutation-kill
every oracle before trusting a differential; ask what inputs a green property test
actually reaches; drive behavior through the shipped binary on hand-written
programs, not only the gate's own corpus.

## Summary

| Dimension | Score | Status |
|-----------|-------|--------|
| Requirements | 28/28 | ✓ satisfied |
| Phases verified | 6/6 | ✓ passed |
| Plans executed | 62/62 | ✓ complete |
| Cross-phase integration | 6/6 boundaries | ✓ sound |
| E2E flows | 13/13 | ✓ reachable |
| Clean-checkout evidence | phase-6 gate green | ✓ reproducible |
| Security | threats_open: 0 | ✓ verified |
| Nyquist coverage | 3/6 reconciled | ⚠ coverage TODO |
| Tech debt | ~25 accepted items | ⚠ review |

**Status: `tech_debt`** — no blockers, no unsatisfied requirements, no integration
gaps. The accumulated deferred items are all documented and accepted, and the
Nyquist coverage TODO on phases 3/5/6 is worth a decision before archiving.

---
*Audited: 2026-09-07*
*Auditor: gsd-audit-milestone (orchestrator inline + gsd-integration-checker)*
