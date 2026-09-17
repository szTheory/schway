# M003 Research: Evidence Integrity and Debt Posture

**Researched:** 2026-09-17 · **Scope:** one decision point — the evidence/assurance
posture M003 should adopt, and the disposition of M002's carried debt.
**Method:** every codebase claim below was produced by running a command against
the tree at `908bac3`, not read out of a planning document. External claims are
cited. The adversarial pass is real, and it concedes.

---

## Verdict

The "cheap to close" Nyquist claim is **half true and dangerously framed**: Phases 12
and 13 are a literal frontmatter flip, Phase 07 is an hour of re-running grounded
commands, and Phases 08 and 11 are not cheap at all — reconciling them surfaces two
dead `go test -run` patterns in Phase 08 that exit 0 forever, six command cells in
Phase 11 that are elided with `…` and cannot be run as written, and one doc-grep row
broken by milestone archival. Worse, the same scan found a **third** dead pattern in
**Phase 09** — a phase the audit certifies `validated` / `nyquist_compliant: true` —
on a row marked `✅ green` whose stated evidence is a SUMMARY's prose. That is the
same defect as the integration checker's "graded from wiring," in a second costume:
*grading a claim from an artifact instead of from execution.* So the durable fix is
one mechanism, not three — a CI-resident groundedness lint plus a closed evidence
vocabulary (`DEFINED | WIRED | REACHABLE | EXERCISED | MUTATION-KILLED`) in which
DX-06's honest partial is expressible without a human having to invent it. On posture:
**do not build a dedicated debt-closure phase for the ten items** — seven of them close
for free or have a natural trigger — but **do front-load one small phase** holding the
residue with no trigger (the six-emitter deletion, the unreviewed D-09-51 flip, Nyquist
07/08/11) plus the mechanization, because every deferral in this project's history
happened when debt sat *behind* feature work in the same phase. On cost: the measured
full suite is **188 s** (not the 60 s every VALIDATION.md still claims), **42 % of the
`session` package's time is one frozen 112-program enumeration**, and there is
**zero** `testing.Short()` and **zero** build-tag lane separation in the tree — so
widening the grammar to include `if` and four arithmetic operators takes that one
subtest from 52 s to an estimated ~25 minutes per commit. Tier it, and buy coverage
growth with metamorphic relations instead of enumeration depth. Finally: yes, there is
a weaker-than-believed claim, and it is a good one — `TestEveryMutationMovesItsClaimedAxis`,
the control whose whole stated value is that it "enumerates every NAT-03 row, never a
hardcoded subset," carries a hardcoded exclusion whose closing condition fired in M001,
and the production helper behind it still hard-refuses the row with a `PENDING-05-08`
error. I proved this by removing the skip: the test goes red.

---

## Is the Nyquist closure actually cheap? (verified)

### What I did

I extracted every `go test <packages> -run <pattern>` command from every
`*VALIDATION*.md` and `*VERIFICATION*.md` under `.planning/` — **146 distinct
(package, pattern) pairs** — and ran `go test <pkg> -list <pattern>` on each,
counting resolved test functions. A pattern that resolves to zero tests is a dead
verification command: `go test -run` exits **0** with `[no tests to run]`, so it
passes forever and can never fail. This is exactly the class plan 13-07 found by hand
(two ungrounded `-run` patterns), generalized to the whole tree.

### Result: three dead patterns, one of them in a "compliant" phase

| Pattern | Package | Source file | Resolves to |
|---|---|---|---|
| `TestComputeLoanLastUsesAndDerivePlaceLoansAgree` | `check/...` | `08-VALIDATION.md` (OWN-06 row) | **0 tests** |
| `TestAuditQLT02BudgetManifest` | `session/...` | `08-VALIDATION.md` (EFF-02 row) | **0 tests** |
| `LoanChainIndex` | `corevalidate` | **`09-VALIDATION.md:85`** (OWN-07 row, marked `✅ green`) | **0 tests** |

Reproduced directly:

```
$ go test ./internal/compiler/check/... -run TestComputeLoanLastUsesAndDerivePlaceLoansAgree -count=1
ok  github.com/codename-lang/lang/internal/compiler/check  0.209s [no tests to run]
EXIT=0
$ go test ./internal/compiler/corevalidate -run 'LoanChainIndex' -count=1
ok  github.com/codename-lang/lang/internal/compiler/corevalidate  0.204s [no tests to run]
EXIT=0
```

**The Phase 09 finding is the important one.** `09-VALIDATION.md` is `status: validated`,
`nyquist_compliant: true`, and the audit names it one of only two compliant phases. Two
of its rows are affected:

- Line 85, OWN-07: `go test ./internal/compiler/corevalidate -run 'LoanChainIndex' -v`
  — matches nothing. Status cell: `✅ green (09-01-SUMMARY.md D1/D2 — corrupted-core
  termination preserved)`. The row was graded from the SUMMARY's prose, not from the
  command.
- Line 84, OWN-07: `-run 'PeerLiveness\|LoanChainIndex'` resolves to exactly **one**
  test, `TestPeerLivenessFileImportsStayIndependent`. The row's stated behavior is
  "peer derives its own liveness bits by forward set-propagation, memoized in the
  existing postorder loop, **never importing `check` or `callgraph`**." The command
  verifies only the final clause. The derivation claim — the substance of OWN-07 — is
  not touched by this command.

**Honest scoping: these are provenance defects, not coverage holes.** The substance
exists under other names. `corevalidate` ships `TestPeerLoanCarryDerivesForwardFromOperations`,
`TestPeerLoanCarryIsMemoizedCalleeBeforeCaller`, `TestPeerLoanCarryTerminatesOnCorruptedCoreArtifact`,
`TestPeerLoanCarryPropagatesAcrossTwoCallHops`, `TestPeerLoanCarryOrderingIsLoadBearingAtDepthTwo`,
plus two mutation kills (`TestPeerLoanCarryForcedTrueReintroducesRetiredDivergence`,
`TestPeerLoanCarryConsultDisabledReintroducesRetiredDivergence`). The tests are named
after `peerLoanCarry`; the map is named after `buildLoanChainIndex`. Likewise Phase 08's
EFF-02 row: the manifest row it claims *does* exist
(`internal/compiler/session/qlt02_budget_manifest.json`, metric
`recomputed_work_growth_exponent`, hard gate, bound 1200 milliexponent, ratified
2026-09-10, commit `b28a92f9`), and `TestBudgetManifestLoads` /
`TestQLT02BudgetManifestFileUnchangedDuringAudit` exercise it. And Phase 08's OWN-06 row
is *obsolete by design* — `computeLoanLastUses` was deleted in Phase 09 (only comments
mentioning it survive, at `core.go:939` and `check_shadow_subsumption_test.go:22`), so
the differential it names cannot exist. That still needs adjudication, not a rename.

So: OWN-06, OWN-07 and EFF-02 are all genuinely supported. What is *not* supported is
the certification that said so.

### Per-phase reconciliation cost, measured

| Phase | Map state | Patterns grounded? | Real reconciliation cost | Cheap? |
|---|---|---|---|---|
| **07** | 9 rows, **all `❌ W0`**; every Wave-0 checkbox and every Sign-Off box unchecked; `wave_0_complete: false` | **Yes, 9/9** (`scripts/verify-phase7.sh` also exists) | Re-run 9 commands, tick 11 Wave-0 items, tick sign-off. Note `-run MutationMatrix` in `originvalidate` resolves to exactly **1** test (`TestStage0SummaryMutationMatrix`) for a row claiming "five seeded faults at BOTH replay sites" — worth a precision check | **Moderate** (~1 h) |
| **08** | 13 rows, all `⬜ pending`, **Task ID column is literally `TBD` on every row** | **No — 2/13 dead** | Adjudicate one obsolete row (deleted mechanism), re-point one row, assign 13 task IDs that were never assigned | **No** |
| **09** | certified `validated` / compliant | **No — 1 dead, 1 under-scoped** | Re-point 2 rows; **and re-examine the certification itself**, since it passed with a dead command | **Already "closed", and wrong** |
| **11** | 27 rows, all `⬜ pending`; **≥6 command cells contain `…`** (`TestEntryFunction…`, `TestEmittedAttributeSet…`, `TestQLT03GeneratorOpKindClosure…`, `TestDropCallSite\|TestDropOrphanFunction\|…`, `TestCache…`, `N =…`) — not runnable as written | Grounded once de-elided (28/28 checked) | De-elide 6 cells, run 27 rows, **plus fix one row broken by archival** (below) | **No** |
| **12** | **All 15 rows already `✅ green`** | Yes | `status: draft` → `validated`, `nyquist_compliant: false` → `true` | **Yes — literal frontmatter flip** |
| **13** | **All 25 rows already `✅ green`**, already `nyquist_compliant: true`; 13-07 re-ran every row for real and corrected two names | Yes | `status: draft` → `validated` | **Yes — literal frontmatter flip** |

**The archival breakage in Phase 11.** Row `11-02 T3` verifies by
`grep -c -E 'flto|escalation, not a pass' .planning/ROADMAP.md`. Today that returns
**0**; the text moved to `.planning/milestones/M002-ROADMAP.md` (7 hits) when the
milestone was archived. `.planning/ROADMAP.md` is now a 59-line index stub. This is
RETROSPECTIVE M001 Key Lesson 6 — "archival is a code change" — recurring for the third
time, and it is the second reason a verification command can silently stop verifying.

### Verdict on the claim

**PARTIALLY REFUTED.** `/gsd-validate-phase 12 13` is genuinely cheap. `07` is cheap-ish.
`08` and `11` are not, and `09` — the phase nobody was going to revisit because it is
already certified — is the one that most needed the scan. The base rate is now measured
twice: 13-07 found 2 dead patterns hand-checking ~25 rows; my scan found 3 dead patterns
across 146 pairs. Reconciliation *does* surface real defects. But note precisely what
kind: **defects in the measuring instrument, not holes in coverage.** That distinction
is what makes the right response mechanization rather than paperwork.

---

## D-13-34: the held-out corpus hole, and its real blast radius

### The alpha-rename claim: CONFIRMED by reading the fixtures

`testdata/phase6/` holds 7 `.lang` files. Stripping comments:

| Class | `heldout_*` | `derivation_*` | Identical? |
|---|---|---|---|
| **move** | `fn relay(buffer: Buffer) -> Buffer { let delivered = take buffer; let echoed = take delivered; echoed }` | `fn relay(item: Buffer) -> Buffer { let moved_once = take item; let moved_twice = take moved_once; moved_twice }` | **Yes — pure alpha-rename** |
| **borrow** | `fn relay(buffer: Buffer) -> borrow(buffer) Buffer { let first = borrow buffer; let second = borrow buffer; let reviewed = borrow first; second }` | `fn relay(item: Buffer) -> borrow(item) Buffer { let alpha = borrow item; let beta = borrow item; let reviewed = borrow alpha; beta }` | **Yes — pure alpha-rename** |
| **match** | `data Signal = Red \| Green \| Blue`, 3 self-mapping arms | `data Mode = Idle \| Active`, 2 self-mapping arms | No — genuinely distinct |

This matches D-13-34's recorded structural summaries exactly
(`{bindingCount:2 matchArmCount:0 borrowCount:0 takeCount:2 maxDepth:1}` for both
move halves; `{3,0,3,0,1}` for both borrow halves). The claim is not restated — it is
reproduced from source.

### Blast radius: narrower than it looks, and not where the audit points

**What does *not* rest on it.** DX-04's requirement text is: *"An agent can introduce,
locate, and repair representative match, move, borrow, cleanup, and stale-evidence
defects **without scraping prose**"* (`M001-REQUIREMENTS.md:90`). That sentence makes
no generalization claim. Its load-bearing evidence is three controls that are wholly
independent of corpus identity:

1. **Protocol-only boundary** — `TestRepairDriverImportsStayOutsideInternal` plus the
   AST scan `TestRepairDriverSourceNeverReferencesHeldoutFixtures`, widened in 13-06 to
   every non-test file in `internal/compiler/check`, with its own non-inertness proof
   `TestRepairDriverSourceHeldoutScanIsNotInert` (a planted `heldout_` reference in a
   scratch copy of `check.go` must be reported).
2. **Prose-scramble anti-theater** — `cmd/lang-repair/antitheater_test.go:635-641` runs
   the match/move/borrow classes against a stand-in binary with scrambled prose. This is
   the direct evidence for "without scraping prose," and an alpha-renamed corpus cannot
   weaken it.
3. **`TestUnrepairableDefectFailsTheGate`** — `unrepairable` is never laundered into a pass.

So **DX-04 as written is intact.** The audit's phrase "it does weaken a claim M001
already made" is true only of the *unwritten* stronger reading.

**What does rest on it.** D-06-29's anti-overfitting split, and through it D-06-23's
inference that the JSON protocol is *agent-legible* rather than author-legible. The
`testdata/phase6/README` states the rule plainly: `derivation_*` are "the ONLY fixtures
anyone may consult when hand-deriving the repair driver's kind→edit mapping in 06-13."
For **move and borrow** the derivation fixture and the scored fixture are the same
program modulo identifiers, so the mapping was hand-derived against a program the CI
gate then scores (`cmd/lang-repair/repair_test.go:514-520` runs all three classes
through `testSourceClassRepair`). **The generalization inference is void for 2 of the 3
source-defect classes.** It survives for `match`. `cleanup` and `stale_evidence` never
had a pair by design (README: the cleanup class reuses
`ReleaseOmissionMutationRunner` against `testdata/phase4/acquire_three_success.lang`;
`stale_evidence_subject.lang` is neither prefix, because there is no kind→edit mapping
to overfit).

This is precisely the overfitting failure mode the APR literature names: patches that
pass the tests that guided them without being a general solution, and evaluation bias
when the analyst of the patches is the author of the repair system
([Smith et al., *Is the Cure Worse Than the Disease? Overfitting in Automated Program
Repair*](https://clairelegoues.com/assets/papers/SmithOverfitting2015.pdf);
[Ye et al., *Automated patch assessment for program repair at
scale*](https://link.springer.com/article/10.1007/s10664-020-09920-w)). The project
identified the hazard correctly in M001 and then built a split that, for two classes,
did not implement it.

**M002 is not affected, and that is verified, not assumed.** `testdata/phase13`'s pairs
are genuinely distinct in topology, not just names:
`derivation_interprocedural_loan_defect.lang` has 2 functions and 1 call edge, while
`heldout_shared_callee_twin_{alpha,mirror}.lang` have 6 functions and a diamond;
`derivation_fallible_call_unconsumed.lang` has 1 function, `heldout_fallible_call_unconsumed.lang`
has 3 in a chain. They are sealed by `testdata/phase13/HELDOUT.sha256` (5 rows) and
policed by `TestCorpusTopologyDisjoint` plus the non-inertness control
`TestCorpusTopologyGuardIsNotInert`. DX-07's held-out evidence is real.

### Recommended disposition — do **not** fix it now

Option C (author replacement M001 fixtures) is the obvious move and it is the wrong
one at this maturity. The entire admissible program space at `testdata/phase6`'s
grammar slice is small enough to enumerate: `session.EnumeratePhase5Closure()` produces
**112 accepted programs from 164 candidates** over `{borrow, borrow mut, take}` at depth
≤ 3. A "structurally distinct" fixture pair drawn from a 112-element space is a
cosmetic pass of the `(bindingCount, matchArmCount, borrowCount, takeCount, maxDepth)`
predicate, not evidence of generalization — you would be buying a green checkbox.
**Attach D-13-34 to the phase that adds arithmetic or control flow**, where a genuinely
distinct move/borrow program is actually constructible, and upgrade the predicate to be
topology-aware at the same time (D-13-26's triple, which "degenerates completely" on
today's zero-interprocedural phase-6 corpus). Record the trigger; that is the whole
work item for M003.

---

## The "graded from wiring" process defect and its durable fix

### It is one defect, not three, and it has now fired four times

The M002 audit records the integration checker marking DX-06 `✓ WIRED` and calling the
requirement satisfied. The generalization is: **a grader that reads an artifact instead
of executing a claim will convert an honest partial into a false green.** Every instance
I found in this investigation is the same shape with a different artifact:

| # | Grader | Artifact read instead of execution | Consequence |
|---|---|---|---|
| 1 | `gsd-integration-checker` | the call graph / wiring | DX-06, DX-07 reported satisfied (rejected by the developer) |
| 2 | `09-VALIDATION.md` row status | `09-01-SUMMARY.md` prose (`✅ green (…D1/D2)`) | a dead `-run` pattern certified green in a `nyquist_compliant: true` phase |
| 3 | `08-VALIDATION.md` rows | the intent of a command, not its resolution | two commands that exit 0 forever |
| 4 | `TestNoNAT03RowRemainsPending` | the marker comment `// PENDING-05-08` **in the production file only** | the surviving stale marker lives in the *test* file and in an *error string*, so the guard cannot see it (see hunt results) |

Instance 4 is the sharpest, because it is a *test* written specifically to prevent
staleness that watches the wrong artifact. Its own doc comment even says it "checks for
the exact marker-comment line, not any prose mention of the string PENDING-05-08
elsewhere in the file (e.g. a stale error message)" — the exact escape it then fell
through.

### The durable fix: make the grader execute, and close the vocabulary

Five changes, ordered by ROI. The first is small enough to land in a single plan and
would have caught three of the four instances.

**F1 — Verification-command groundedness lint (highest ROI in this document).**
A Go test that walks `.planning/**/*VALIDATION*.md` and `*VERIFICATION*.md`, extracts
every `go test <pkgs> -run <pattern>`, and fails if `go test -list <pattern>` returns
zero tests for any pair, or if any command cell contains an elision character (`…`).
I ran exactly this today: **146 pairs, 3 failures**. It is fast (`-list` compiles but
runs nothing), deterministic, offline, and requires no new dependency. Give it a
non-inertness control in the project's own style — a planted unresolvable pattern in a
scratch copy of a VALIDATION.md must be reported — mirroring
`TestRepairDriverSourceHeldoutScanIsNotInert`.

**F2 — A closed evidence vocabulary, so the honest partial is expressible.**
Replace free-text satisfaction cells with exactly five grades, machine-checked the way
`debtRegisterSeverities` already closes the severity vocabulary
(`session_test.go` `checkDebtRegister`):

| Grade | Means | DX-06 today |
|---|---|---|
| `DEFINED` | the rule exists in a spec/plan | ✓ |
| `WIRED` | code exists and is reachable from a dispatch site | ✓ |
| `REACHABLE` | an input exists that reaches it in production | ✗ — `resolveBlame` has 1 definition, 0 call sites |
| `EXERCISED` | a resolving test drives it end to end | ✗ |
| `MUTATION-KILLED` | a seeded fault makes the control go red | ✗ (compile-time guard only) |

**A requirement may be marked satisfied only at `EXERCISED` or above.** This is the
direct fix to the M002 observation: the checker's failure was that its grammar had no
way to say "wired but not reachable," so it said "satisfied." Give it the word and it
cannot make the error. As a bonus, the vocabulary retro-classifies D-12-43
(`WIRED`, unconstructible control) and D-13-10a (`DEFINED`, withdrawn) without prose.

**F3 — Evidence cells cite a resolvable test name or a declared escape ID. Never free
text.** F1 enforces resolvability; this enforces that there is something to resolve.
The project already has escape IDs (`escape:coordinated-source-to-core-false-claim`,
`escape:callback-invocation-unsubjected`) — reuse the mechanism.

**F4 — Make Nyquist a gate, not a skill.** RETROSPECTIVE records this as M001 Lesson 5
and M002 Lesson 5, same mechanism, two milestones: *compliant exactly for the phases
where it was a gate condition.* Phase 13 is the proof — its substance was fully there
(13-07 re-ran every row and corrected two names) and the marker still said `draft`
because `validate-phase` never ran. Refuse phase-complete while
`status: draft`. This is the one remedy the project has written down twice and never
implemented.

**F5 — Demote the integration checker by contract.** Consume its structural findings
(dispatch sites, import guards, orphaned exports, E2E flows — all of which held under
spot-check) and *do not consume* its requirement-satisfaction column at all. Under F2,
satisfaction is derived from graded evidence, so the column has no consumer and can be
dropped rather than argued with each milestone.

---

## Debt posture: dedicated phase vs attached-to-phase

### The ten items are not ten problems

| Cluster | Items | Natural trigger | Marginal cost if attached |
|---|---|---|---|
| **A — six single-function emitters** | D-11-02, D-12-36, D-11-27 | **None.** Multi-function emission has been the only live path since Phase 11 | Pure debt; no feature phase wants it |
| **B — event identity on shared-leaf diamonds** | D-11-51 → D-12-21 | Control flow (a loop over a diamond multiplies colliding identities) | Low — but must be a *precondition*, not a trailing task |
| **C — single-type-per-function** | D-13-02b, D-13-10a, twin-pair addendum | Return type ≠ parameter type | **Zero.** They close automatically |
| **D — unreviewed verdict flip** | D-10-C04, D-11-27 | None. It is a human review, not code | A checkpoint in someone's plan |
| **E — trigger-only findings** | D-12-43, D-13-34 | Grammar widening / payload-value-aware channel | Zero — record the trigger |

### Options

| | Dedicated debt-closure phase | Attached to the phase that touches it | **Hybrid: front-loaded residue phase + attachment** |
|---|---|---|---|
| **Cluster C cost** | Wasted — a full phase budget spent on items that close for free | Zero | Zero |
| **Cluster B quality** | Poor — you need the control-flow design in hand to answer event identity | Good | Good, as a wave-0 precondition |
| **Cluster A survival** | Good — it finally gets a budget | **Bad. This is exactly how it was deferred twice**: attached to Phase 11, lost; re-attached to Phase 12, lost | Good |
| **D-10-60 compliance** | Satisfied | A third deferral is likely, violating a rule this project wrote for itself | Satisfied |
| **Evidence mechanization (F1–F5)** | Fits naturally | Has no owner and will not happen | Fits naturally |
| **Historical base rate** | Untested here | **0 for 2** on Cluster A | — |
| **Risk** | Register theater: a phase whose deliverable is paperwork | Feature pressure wins every time | Scope creep into a mini-milestone |

### Recommendation: **hybrid, front-loaded**

Create one small M003 Phase 0 — *Evidence Mechanization and Untriggered Debt* — holding
exactly:

1. **F1 groundedness lint + F2 evidence vocabulary + F4 Nyquist gate** (the mechanization;
   this is the majority of the value).
2. **Cluster A** — delete the six single-function emitters, or retire D-10-60 by an
   explicit `REQUIREMENTS.md` amendment naming the reversal. No third option.
3. **Cluster D** — the D-09-51 negative-control verdict-flip review, as a named
   `gate="blocking-human"` checkpoint. RETROSPECTIVE M002 Lesson 3: *"Flagged for human
   review is not a work item."*
4. **Nyquist 07, 08, 11** — reconciled *after* F1 exists, so the lint does the finding.
   Do 12 and 13 inline as frontmatter flips; they need no phase.

Everything else is attached, with the attachment recorded as a **wave-0 precondition of
its owning phase**, not a trailing task: Cluster B → the control-flow phase; Cluster C
and D-12-43 and D-13-34 → the type-widening / grammar phase.

**Why front-loaded, not trailing.** A debt phase at the *end* of M003 will lose to M004
by the same mechanism that lost D-11-02 twice. A debt phase at the *front* is paid before
there is a feature budget to compete with it, and F1 then improves every subsequent
phase's evidence rather than auditing it afterwards.

### Register hygiene — three cheap mechanical fixes

`TestDebtRegistersAreWellFormed` (`internal/compiler/session/session_test.go:2613`)
enforces format only: an `items:` count matching the table, one `### D-XX-NN` detail per
row, a closed severity vocabulary, and a **non-empty** `Landing phase` cell. The string
`"OPEN and UNOWNED — reopens only when …"` is non-empty, so it passes. Three fixes:

- **Close the `Landing phase` vocabulary** the way severity is closed: a phase ID,
  `CLOSED (<commit>)`, or `UNOWNED (<reopening-trigger>)`. Assert the `UNOWNED` count at
  milestone close is ≤ the count at milestone open.
- **Enforce D-10-60 mechanically.** D-10-60's own text states the gap: *"Verified at
  planning time that the suite will NOT catch this … does not track deferral hop count.
  This is therefore a declared rule enforced by review."* A rule enforced by review is
  the thing that failed. Record prior landing phases per row; a third distinct one fails
  the test.
- **Generate a project-level `.planning/DEBT-INDEX.md`** from the 12 archived registers.
  77 items across 12 files inside archived milestone directories is where a register
  stops being read. Provenance stays in the registers; attention goes to the index.

---

## Evidence strategy when the language surface widens (cost-controlled)

### Measured cost today, not estimated

```
go test ./...            188.2 s wall  (user 103 s, sys 67 s), exit 0
  session                186.3 s       ← 99 % of the critical path
  testsupport             56.0 s
  native                  44.0 s
  measure                 17.0 s
1,050 test functions · 25 packages · 115 .lang fixtures · 4,189 lines of Lang
```

Inside `session`:

| Test | Time | Share of session |
|---|---|---|
| `TestPhase5CorpusThreeEngineAgreement` | **79.2 s** | 42 % |
| └ subtest `enumerated-closure` | **52.5 s** | 28 % |
| `TestPhase11InterproceduralDifferential` | 9.5 s | 5 % |
| everything else | ~97 s | 53 % |

Every `*-VALIDATION.md` still declares "~60 seconds (full suite)" (08, 11, 12) or
"~120 s" (10). **The measurement is stale by 1.6–3×**, and nothing detects that, because
the estimate lives in prose.

There is **zero** `testing.Short()` and **zero** `//go:build` tag in any `_test.go` file
in `internal/` or `cmd/`. CI (`.github/workflows/ci.yml`) runs `go test ./...` **and**
`go test -race ./...` on every push. `risk_lanes.json` (40 rows) is a **product** feature
— changed-risk lane selection that `lang verify` offers to users — and is **not applied
to the project's own suite.** The project built a lane selector for its users and does
not use it on itself.

### The combinatorial cliff, computed from the real generator

`enumeratePhase5Closure()` (`session_phase5_corpus.go:213`) iterates
`{Byte, Buffer} × length 0..3 × 3^length chain kinds × {touchParam, ¬touchParam}` plus 4
foreign shapes: **164 candidates → 112 accepted, 52 rejected**, at **~0.47 s/program**
(each pays a full `-O0` / `-O3` / `-O3 -flto` clang round trip).

| Grammar | Chain alphabet | Candidates | Est. subtest time |
|---|---|---|---|
| Today | 3 (`borrow`, `borrow mut`, `take`) | 164 | **52 s** (measured) |
| \+ 4 arithmetic operators | 7 | ~800 | ~4 min |
| \+ `if/else` (×2 body shapes) | 7 | ~1,600 | ~12 min |
| \+ depth 4 (loops need it) | 7 | ~9,600 | **~75 min** |

Per commit. Twice (`-race`). This is the "machinery built in the absence of load" risk
LANGUAGE-MATURITY.md names, arriving as a wall rather than a slope.

### The strategy, in priority order

**S1 — Split the enumeration by layer, not by size.** The bounded closure's value is
exhaustiveness over a grammar slice; its cost is entirely the native tier. Run the
**admission layer** (parse → `check` → `corevalidate` → `originvalidate` → `pathoracle`)
**exhaustively, per commit** — that is milliseconds per program and tolerates a much
larger alphabet. Promote a **declared, risk-weighted sample** to the four-tier native
differential. This preserves the strong claim (admission rules are exhaustively checked
over the slice) and samples only the claim with the lowest marginal return per second.
The external warrant is blunt: nearly half of fuzzer-found LLVM miscompilations propagate
to a binary, but the affected code is tiny and almost none cause a test failure or
execution divergence ([Marcozzi et al., *Compiler Fuzzing: How Much Does It
Matter?*, OOPSLA 2019](https://dl.acm.org/doi/10.1145/3360581)).

**S2 — Make the native tier content-addressed, not schedule-addressed.** The enumeration
is a pure deterministic function of the grammar and the generator
(`TestPhase5EnumerationIsDeterministic` already asserts this). Hash its inputs; run the
native tier only when the digest moves, or nightly. The project already has exactly this
discipline in `internal/compiler/cache` and in the `ClosureDigest` chain — apply it to
its own evidence. Today it re-proves an unchanged 112-program closure on every commit.

**S3 — Buy coverage with metamorphic relations, not enumeration depth.** Metamorphic
testing addresses the oracle problem by asserting relations across *multiple* inputs
rather than needing a correct output for one ([Chen et al., *Metamorphic Testing: A
Review of Challenges and Opportunities*, ACM CSUR
2018](https://dl.acm.org/doi/10.1145/3143561)); EMI is the compiler-specific instance,
generating variants guaranteed to behave identically and finding a large number of
additional bugs in GCC and LLVM ([Le, Afshari, Su, *Compiler Validation via Equivalence
Modulo Inputs*, PLDI 2014](https://fm.csl.sri.com/SSFT14/PLDI14-Orion.pdf)). For Lang
the relations are already implied by the semantics and cost O(1) each:
alpha-rename invariance; insertion of an unreachable `match` arm; insertion of an unused
`let`; reordering independent function declarations. Each is an *unbounded* program
source with a *free* oracle, and each scales *with* the grammar instead of against it.
The project already ships one without naming it —
`TestExplainFunctionIdentitySurvivesReorder`. Generalize the idea; do not grow depth.
As a bonus, alpha-rename invariance is the exact relation whose *absence* is D-13-34.

**S4 — Turn `risk_lanes.json` on the project's own suite.** Per-commit lane = admission
layer + changed package + **all** mutation-kill controls (they are cheap:
`TestEveryMutationMovesItsClaimedAxis` is 1.44 s). Nightly = native four tiers +
enumeration + `-race` + sanitizers. State the risk honestly rather than hiding it: at
industrial scale, test selection cuts median feedback by up to 96 % but can miss up to
55 % of failures ([Contrasting test selection, prioritization, and batch testing at
scale, EMSE 2024](https://link.springer.com/article/10.1007/s10664-024-10589-8)) — which
is precisely why the *controls* stay in the fast lane and only the *corpus* is sampled.

**S5 — Give the suite a wall-clock budget row.** `qlt02_budget_manifest.json` already
holds a hard `recomputed_work_growth_exponent` bound (1200 milliexponent) with a
`machine_id` and a ratifying commit. Add a `suite_wall_clock_p95` row the same way. A
phase that exceeds it must declare its tiering or its cut, in the manifest, at the moment
it exceeds it. This converts the stale "~60 seconds" prose into a gate.

**S6 — Grow mutation-kills per *rule*, not per *fixture*.** Mutation score is a
contested proxy for real-fault detection, and the coupling between common mutation
operators and real faults is partial at best ([Just et al., *Are mutants a valid
substitute for real faults in software testing?*, FSE
2014](https://homes.cs.washington.edu/~mernst/pubs/mutation-effectiveness-fse2014.pdf);
[Papadakis et al., *Are Mutation Scores Correlated with Real Fault
Detection?*, ICSE 2018](https://dl.acm.org/doi/pdf/10.1145/3180155.3180183)). This
project's hand-authored one-control-per-claim discipline (QLT-08) is *stronger* than a
generic mutation tool for exactly that reason, and STANDING-VERDICTS already rejects
`go-mutesting` and `gremlins`. Keep it — but bind the obligation to each new **semantic
rule** (each operator, each control-flow construct), not to each new fixture, or the
control set grows with the corpus.

---

## Weaker-than-believed claims I found

### W1 (primary) — the enumerating mutation-kill control carries a hardcoded exclusion whose closing gate fired in M001

`TestEveryMutationMovesItsClaimedAxis`
(`internal/compiler/session/session_phase5_alias_test.go:233`) is the mechanized form of
M001's Key Lesson 1. Its own doc comment states its value: it *"enumerates every NAT-03
row (**never a hardcoded subset**)."* It then contains, at line 240:

```go
if strings.Contains(row.CorpusProgram, "phase5/allocator_mismatch.lang") ||
   strings.Contains(row.CorpusProgram, "phase5/retained_pointer.lang") {
    t.Skip("PENDING-05-08: closed at plan 05-09's gate, which depends on this plan and 05-08")
}
```

Plan 05-09 shipped in **M001, 2026-09-07**. The skip reason cites a gate that closed ten
days ago; the skip is still green in CI today.

**I proved the skip is not merely cosmetic.** Removing the `allocator_mismatch` half in a
scratch edit and re-running (file restored afterwards; `git status` clean):

```
--- FAIL: TestEveryMutationMovesItsClaimedAxis/control:native.sanitize.allocator_mismatch
    row {ControlID:control:native.sanitize.allocator_mismatch … Subjected:true}
    did not move its claimed axis: AssertMutationMovesAnAxis: unsupported control
    "control:native.sanitize.allocator_mismatch" (row not yet subjected — see PENDING-05-08)
```

So the production helper `session.AssertMutationMovesAnAxis`
(`session_phase5_alias.go:510`) **still hard-refuses the row in its `default:` arm at
line 523**, with a `PENDING-05-08` error string, for a row declared `Subjected: true`.

**Root cause: two coexisting laws.** Plan 05-09 closed the gate not by teaching
`AssertMutationMovesAnAxis` the row, but by adding a *second* law —
`Phase5AssertMutationMovesAnAxis` (`session_phase5.go:424`), a strict superset wrapper
that special-cases `ControlSanitizeAllocatorMismatch` and delegates everything else —
plus a *bespoke single-row test*, `TestNAT03SanitizerRowMovesItsClaimedAxis`. That test
passes (0.35 s), so the **claim** is supported. What is not supported is the belief that
the enumerating control enumerates. This is RETROSPECTIVE M001 Key Lesson 2 — *"two
coexisting laws is a defect with a delayed fuse"* — shipped, unretired, and invisible
because the fuse is a green skip.

**And the guard written to catch this cannot see it.** `TestNoNAT03RowRemainsPending`
(`session_phase5_test.go:235`) exists specifically to refuse a stale `PENDING-05-08`, but
scans only `session_phase5_alias.go` for the exact line `// PENDING-05-08`. The two
surviving markers are (a) a `t.Skip` string in `session_phase5_alias_test.go:241` and
(b) an error string at `session_phase5_alias.go:523` — which the guard's own comment
explicitly excludes: *"not any prose mention of the string PENDING-05-08 elsewhere in the
file (e.g. a stale error message)."*

**Honest bound on the damage.** `control:native.sanitize.retained_pointer` is declared
`Subjected: false` with `EscapeID: "escape:callback-invocation-unsubjected"` — a named,
honest escape, not a hidden hole; the NAT-03 arithmetic is genuinely 6 subjected + 1
escape. And `allocator_mismatch`'s axis movement *is* asserted, by the bespoke test. So
no shipped requirement is unsupported. The weaker-than-believed thing is the **control
architecture**: the general law is blind to a row it is believed to cover, the
anti-staleness guard watches the wrong artifact, and a `Subjected: true` row is excluded
from the enumeration by fixture-path string match.

### W2 — Nyquist compliance itself

Phase 09 is `status: validated`, `nyquist_compliant: true`, and cited in the milestone
audit as one of two compliant phases. It contains a verification command that resolves
to zero tests and is marked `✅ green` on the strength of a SUMMARY's prose. The
certification therefore does not mean what the audit's `nyquist: compliant_phases:
["09","10"]` field implies a reader should take from it. (Substance intact — see §2.)

### W3 — the stale runtime estimate as an unguarded input

Five VALIDATION.md files declare a 60–120 s full-suite runtime and "max feedback latency"
of 60 s. Measured: 188 s. Those numbers are inputs to the sampling-rate contract ("after
every plan wave: `go test ./...`"), so a 3× error is a 3× error in the phase's own
feedback-budget reasoning — in a project whose first constraint is feedback latency. No
mechanism detects it, because it lives in prose rather than in
`qlt02_budget_manifest.json`.

### W4 — the register's ownership claim

`TestDebtRegistersAreWellFormed` enforces a *non-empty* `Landing phase`. Thirteen cells
read `OPEN and UNOWNED — …` and pass. The register's strongest apparent property —
"every deferral has a named landing phase, and a test enforces it" — is enforced only
against blankness, not against unownedness.

---

## Role-lens disagreements

| Lens vs lens | The disagreement | Resolution I recommend |
|---|---|---|
| **QA test architect** vs **SRE** | Architect: exhaustive enumeration is the claim; sampling weakens it. SRE: 52 s/commit to re-prove a frozen 112-program closure is waste. | Both are right about different layers. **S1 + S2**: exhaustive at the admission layer per commit, content-addressed at the native layer. The claim is unchanged because the corpus is a pure function of a hashable input. |
| **Formal-methods engineer** vs **compiler-testing specialist** | FM: independent re-derivation is the asset — add a fourth peer. Specialist: Alive2, the state of the art, [explicitly does not support interprocedural optimizations](https://github.com/AliveToolkit/alive2); the return is in generators and oracles, not more peers. | Already adjudicated in STANDING-VERDICTS ("Reject a second comparator"; "Name it, don't schedule it"). **Keep it adjudicated.** Spend M003's evidence budget on metamorphic relations (S3), not a fourth deriver. |
| **Data scientist** vs **auditor** on the word *Nyquist* | The sampling metaphor does not hold: Nyquist requires a bandlimited signal and a sampling rate ≥ 2× its highest frequency. There is no frequency here. What the artifact actually checks is *gap length* ("no 3 consecutive tasks without automated verify") and *command groundedness*. | The auditor's reading is the useful one. **Rename it in M003** to what it measures — verification-command groundedness and sampling-gap length. The metaphor is actively harmful: it implies coverage is being measured, which is how three dead commands passed certification. |
| **Engineering manager** vs **auditor** on registers | EM: 77 items in 12 archived files is past the point anyone reads it; cut to ≤10 live rows. Auditor: every row is a real recorded finding; deleting history destroys provenance. | **Both.** Generate a live `.planning/DEBT-INDEX.md`; leave the archived registers untouched as provenance. Attention and history are different artifacts. |
| **Auditor** vs **everyone** on DX-06 | Auditor: "built but unreachable" must never read as satisfied. Others: the code is correct and exhaustively unit-tested, so calling it unsatisfied undersells it. | The vocabulary (F2) dissolves this. DX-06 is `WIRED`, not `EXERCISED`. Nobody has to argue. |

---

## Prior art and lessons

- **Csmith / EMI / YARPGen.** Generation-based (Csmith, YARPGen) and mutation-based (EMI)
  compiler testing are complementary; EMI's key property is that it needs no reference
  implementation and no ground truth, because the variant is guaranteed to behave
  identically ([Le et al., PLDI
  2014](https://fm.csl.sri.com/SSFT14/PLDI14-Orion.pdf)). **Lesson for this project:**
  the interpreter oracle is expensive; metamorphic/EMI relations give you a free one.
- **Compiler fuzzing's real-world impact.** Across 309 Debian packages and >10 M LOC,
  nearly half of fuzzer-found Clang/LLVM miscompilations propagate into a binary, but the
  affected code is typically tiny, and the resulting syntactic changes either have no
  semantic impact or need very specific runtime circumstances to diverge
  ([OOPSLA 2019](https://dl.acm.org/doi/10.1145/3360581)). **Lesson:** the marginal value
  of the Nth optimizer-tier run falls fast. This is the strongest argument for tiering.
- **Alive2.** Bounded translation validation with a zero-false-alarm design; found 47 new
  LLVM bugs. Its stated limits are loops (unrolled once), **no interprocedural analysis**,
  and SMT timeouts ([Lopes et al., PLDI
  2021](https://users.cs.utah.edu/~regehr/alive2-pldi21.pdf)). **Lesson:** the project's
  standing verdict to pursue empirical differential testing rather than formal
  interprocedural proof is correct and should not be revisited in M003.
- **CompCert's TCB.** Even the most verified production compiler leaves preprocessing,
  lexing, parsing, elaboration, the assembly-expansion pass, the assembler and the linker
  unverified and merely trusted, plus Coq extraction and the OCaml toolchain
  ([Monniaux & Boulmé, *The Trusted Computing Base of the CompCert Verified
  Compiler*, ESOP 2022](https://arxiv.org/abs/2201.10280)). **Lesson:** trust boundaries
  always leak somewhere; the project's honest `escape:coordinated-source-to-core-false-claim`
  is the right genre. Adding assurance at an already-strong layer does not move the
  weakest link.
- **Mutation testing.** Coupling between common mutation operators and real faults is
  partial, mutation score is a contested predictor of real-fault detection, and the
  equivalent-mutant problem inflates it ([Just et al., FSE
  2014](https://homes.cs.washington.edu/~mernst/pubs/mutation-effectiveness-fse2014.pdf);
  [Papadakis et al., ICSE
  2018](https://dl.acm.org/doi/pdf/10.1145/3180155.3180183)). **Lesson:** the project's
  hand-authored, claim-bound controls are better than a score. Keep QLT-08; bind it to
  rules, not fixtures.
- **MC/DC and structural coverage.** MC/DC does not guarantee fault *propagation* to an
  oracle, masking lets corrupted internal state be absorbed downstream, and its
  effectiveness depends on program structure; a single random input has been observed to
  reach 100 % coverage on some code ([Rapita, *MC/DC
  coverage*](https://www.rapitasystems.com/mcdc-coverage); [Modified Condition/Decision
  Coverage in GCC, arXiv 2501.02133](https://arxiv.org/pdf/2501.02133)). **Lesson:**
  directly analogous to "wired ≠ reachable ≠ exercised." Structural presence is not
  evidence of exercise. This is the DX-06 lesson stated by a standards body.
- **APR overfitting.** Patches that pass the guiding tests need not generalize; held-out
  tests help but can themselves be overfit; and evaluation bias is systemic because
  analysts are usually the repair system's authors
  ([Smith et al. 2015](https://clairelegoues.com/assets/papers/SmithOverfitting2015.pdf);
  [Ye et al., EMSE 2020](https://link.springer.com/article/10.1007/s10664-020-09920-w)).
  **Lesson:** D-06-29's split was the right instinct; D-13-34 is what happens when the
  split is nominal for two of three classes.
- **Metamorphic testing.** The canonical answer to the test-oracle problem; effective on
  compilers specifically ([Chen et al., ACM CSUR
  2018](https://dl.acm.org/doi/10.1145/3143561)). **Lesson:** the single highest-leverage
  *new* technique available to this project, and it costs no new dependency.
- **Test selection at scale.** Up to 96 % median feedback-time reduction, at the cost of
  missing up to 55 % of failures; flakiness erodes trust in the signal
  ([EMSE 2024](https://link.springer.com/article/10.1007/s10664-024-10589-8)). **Lesson:**
  sample the corpus, never the controls.

---

## Adversarial pass: is this project over-invested in assurance?

### The strongest case for yes

LANGUAGE-MATURITY.md's own numbers are the prosecution's best exhibit: **~60–70 % of a
verification platform, ~5–10 % of a language.** The assurance stack — three
non-importing peers, a four-tier native differential, an HDD reducer, an ASan/UBSan
lane, a content-bound evidence layer, a repair driver, a cause DAG — has only ever been
exercised against 115 fixtures totalling 4,189 lines, average ~36 lines, none with a
loop, none with arithmetic, most existing to exercise exactly one admission rule. The
machinery is validated against a workload that is not the workload.

The concrete absurdity: **188 seconds of CI per commit on a language that cannot add two
numbers**, of which 52 seconds proves four-way optimizer agreement over 112
machine-generated programs whose entire alphabet is `{borrow, borrow mut, take}` — a
corpus that has not changed since M001 and structurally *cannot* change until the
grammar does. That is pure carrying cost, re-paid on every push, twice (`-race`).

The external evidence cuts the same way. The OOPSLA'19 study is the sharpest blade: for a
*mature compiler with millions of real users*, fuzzer-found miscompilations largely fail
to produce user-visible divergence. The expected user-visible value of a fourth optimizer
tier on a language with **zero** users is, to a first approximation, zero. And CompCert's
TCB paper says the second quiet part: even total verification leaves the parser, the
elaborator, the assembler and the linker trusted. This project's own
`escape:coordinated-source-to-core-false-claim` documents that source→core remains
unproven. Piling more assurance onto the already-strongest layer does not move the
weakest link — it just makes the strong layer's report longer.

Debt-register work is the purest instance. 77 items across 12 registers, thousands of
lines of adjudication prose, for a language with no `if`. When thirteen cells read
`OPEN and UNOWNED` and satisfy a well-formedness test by being non-empty strings, and
when a phase's entire remaining Nyquist obligation is flipping `status: draft` to
`status: validated`, you are close to the line where a register becomes theatre.

And the opportunity cost is the real argument. M002 spent 61 plans over 6 days to add
exactly one `OperationKind`. LANGUAGE-MATURITY.md's dependency table lists arithmetic,
iteration, and collections as "not started, not scheduled" and warns to "expect several
more milestones the size of M002." Every hour spent on evidence paperwork is an hour not
spent shortening that queue — on a project whose stated core value is *the shortest
reliable path from intent to sound evidence.*

### Where I concede

- **The frozen native enumeration is over-invested as currently scheduled.** 52 s/commit
  re-proving an unchanged, deterministic 112-program closure is indefensible. Tier it to
  content-addressed + nightly. The claim is *not* weakened, because the input is hashable.
- **A dedicated debt-closure phase spending a full phase budget on ten items would be
  over-investment.** Cluster C closes for free. D-12-43 and D-13-34 are triggers, not
  tasks. Seven of ten items should cost approximately nothing.
- **D-13-34 should not be fixed now.** In a 112-program admissible space, "structurally
  distinct" is a cosmetic predicate pass. Buying it today buys a checkbox.
- **Nyquist compliance, as compliance, is genuinely low-value.** Nobody is harmed by
  `status: draft`. The frontmatter flip for 12 and 13 is worth five minutes precisely
  because it costs five minutes, not because it matters.

### Where I do not concede

- **The three dead `-run` patterns are the counterexample to the whole thesis, and they
  were found *by* the evidence-debt work the adversary wants to skip.** They are not
  defects in the product; they are defects in the *measuring instrument*. When the
  instrument is wrong, every future claim inherits the error, and the error compounds
  with surface width rather than shrinking with it. One test file (F1) fixes the class
  permanently, and I demonstrated today that it catches three real instances in 146 pairs.
  That is the highest ROI item in this document by a wide margin, and it is *cheaper than
  the paperwork the adversary is objecting to.*
- **The over-investment framing mismeasures the asset.** LANGUAGE-MATURITY.md's own
  strategic read is right: adding arithmetic to a sound ownership core is comparatively
  easy; retrofitting ownership soundness onto an expressive language is what takes years
  or never lands (Rust took years; C++ never did). The peer architecture will still be
  the asset at M010. The cost objection is about the *runtime of one frozen corpus*,
  which is a tiering problem — not about the *architecture*, which is the investment.
- **The named risk argues for spending, but on a different thing.** LANGUAGE-MATURITY.md
  warns that this is "machinery built in the absence of load," and that Go 1.18's generics
  front-end regression and Rust's Polonius wall are the precedents. That argues for
  spending on evidence **infrastructure** — lane selection, budget ceilings, groundedness
  lints, metamorphic relations — precisely because those determine whether the machinery
  survives load. It argues *against* spending on evidence **ceremony**.

### Net honest verdict

**The project is over-invested in evidence *ceremony* and correctly invested in evidence
*architecture*, and it has been spending on the wrong one of the two.** M003 should
spend roughly one phase-equivalent on evidence, front-loaded, with ~80 % of it on
mechanization (F1, F2, F4, S1, S2, S5) and ~20 % on the untriggered debt residue
(Cluster A, Cluster D). Reconciliation paperwork for its own sake — Nyquist 12 and 13 —
is five minutes, done inline, and should never again be described as a milestone
liability. Everything else waits for its trigger.

---

## Proposed M003 requirements this implies

Draft REQ text, each with a stated failing direction.

- **EVD-01 (groundedness).** Every verification command recorded in any phase's
  `VALIDATION.md` or `VERIFICATION.md` resolves to at least one test.
  *Verified by* `TestVerificationCommandsAreGrounded`: extract every
  `go test <pkgs> -run <pattern>` from `.planning/**`, fail if `go test -list <pattern>`
  returns zero tests for any pair, and fail on any command cell containing an elision
  character. Non-inertness control: a planted unresolvable pattern in a scratch copy must
  be reported.
  *Fails when:* a map row cites a renamed, deleted, or never-created test — today, 3 of
  146 pairs.

- **EVD-02 (closed evidence vocabulary).** Requirement satisfaction is graded on
  `DEFINED | WIRED | REACHABLE | EXERCISED | MUTATION-KILLED`, and no requirement may be
  marked satisfied below `EXERCISED`.
  *Verified by* a traceability test enforcing the closed vocabulary, in the shape of
  `checkDebtRegister`'s existing severity check, plus a pinned assertion that DX-06 reads
  `WIRED` until a B1-shaped interprocedural diagnostic is constructible.
  *Fails when:* a grader writes a satisfaction verdict from wiring.

- **EVD-03 (evidence provenance).** Every evidence cell cites either a test name that
  `go test -list` resolves or a declared `escape:` ID. Free text is refused.
  *Fails when:* a row's evidence is a SUMMARY citation (the Phase 09 case).

- **EVD-04 (no skip outlives its trigger).** Every `t.Skip` in the tree names a debt ID or
  an escape ID, and no skip reason may cite a plan, phase, or gate that has already
  closed. AST scan with a non-inertness control.
  *Fails when:* a `PENDING-NN-NN` skip survives the gate it names — today,
  `session_phase5_alias_test.go:241`.

- **EVD-05 (one axis-movement law).** `AssertMutationMovesAnAxis` is the sole
  axis-movement law; `Phase5AssertMutationMovesAnAxis` is deleted, and every NAT-03 row
  with `Subjected: true` is asserted by the enumerating control with **zero** per-row
  exclusions.
  *Fails when:* a second coexisting law exists, or the enumerating control excludes a
  subjected row (M001 Key Lesson 2 and 3, both currently live).

- **EVD-06 (suite wall-clock budget).** Total suite wall-clock has a declared p95 ceiling
  in `qlt02_budget_manifest.json` with a `machine_id` and a ratifying commit, measured by
  the existing `measure` protocol. A phase exceeding it must amend the manifest naming
  its tiering or its cut.
  *Fails when:* measured p95 exceeds the bound with no amendment. Baseline to ratify:
  188 s measured 2026-09-17.

- **EVD-07 (layered enumeration).** The bounded enumerated closure runs exhaustively at
  the admission layer on every commit; the native four-tier layer runs only when the
  generator's input digest changes, or nightly, over a sample drawn by a declared rule.
  *Fails when:* the native tier runs on an unchanged digest (waste), or the admission tier
  is sampled (weakened claim).

- **EVD-08 (metamorphic relations).** At least three relations — alpha-rename invariance,
  unreachable-`match`-arm insertion invariance, independent-function reorder invariance —
  hold across all five comparator axes over the admission corpus, each with a
  mutation-killed non-inertness control.
  *Fails when:* a relation ships without a control that has been observed red.

- **DBT-01 (register ownership vocabulary).** Every debt row's `Landing phase` cell draws
  from a closed vocabulary: a phase ID, `CLOSED (<commit>)`, or
  `UNOWNED (<reopening-trigger>)`. The `UNOWNED` count at milestone close is ≤ the count
  at milestone open.
  *Fails when:* an item is deferred without a trigger, or the unowned population grows —
  today, 13 cells pass only because they are non-empty.

- **DBT-02 (mechanical D-10-60).** Each debt row records its prior landing phases; a third
  distinct landing phase fails `TestDebtRegistersAreWellFormed`.
  *Fails when:* an item acquires a third landing phase — the D-11-02 → D-12-36 pattern
  that D-10-60 was written to stop and did not.

- **DBT-03 (Cluster A disposition).** The six single-function emitters (D-11-02 / D-12-36)
  are deleted, **or** D-10-60 is retired by an explicit `REQUIREMENTS.md` amendment naming
  the reversal.
  *Fails when:* neither happens and the item crosses a second milestone boundary.

- **DBT-04 (Nyquist becomes a gate).** A phase cannot be marked complete while its
  `VALIDATION.md` `status` is `draft`.
  *Fails when:* a phase closes unreconciled — the mechanism that produced 3/6 in M001 and
  2/7 in M002.

---

## Confidence + what would change my mind

| Claim | Confidence | Basis | What would change my mind |
|---|---|---|---|
| Three dead `-run` patterns exist, including one in Phase 09 | **HIGH** | Executed `go test -list` over 146 extracted pairs; reproduced exit-0 `[no tests to run]` | A flaw in my extraction regex — I mitigated by skipping elided cells and de-duplicating, but a differently-shaped command (e.g. a `sh scripts/…` wrapper) would be missed, which would only *increase* the count |
| "Cheap to close" is true for 12/13, false for 08/11 | **HIGH** | Read all seven maps; 12 and 13 are fully `✅ green`, 08 has `TBD` task IDs and 2 dead patterns, 11 has ≥6 un-runnable elided cells | Discovering that `/gsd-validate-phase` auto-repairs elided commands and re-derives task IDs, which would shrink 11's cost substantially |
| D-13-34's move/borrow pairs are alpha-renames; match is distinct | **HIGH** | Read all six fixtures directly | Nothing — this is textual |
| DX-04's written claim survives D-13-34 | **MEDIUM-HIGH** | The requirement text says "without scraping prose," and the prose-scramble + protocol-boundary controls are corpus-independent | Finding that an M001 phase document asserts held-out generalization as a *criterion* rather than a control; I read `M001-REQUIREMENTS.md` and the phase-6 README, not all of 06-*-PLAN.md |
| W1 — the axis-movement control is weaker than believed | **HIGH** | Empirically: removed the skip, the test went red with the production helper's own `PENDING-05-08` refusal; file restored, tree clean | Nothing about the mechanism. The *severity* would drop further if `TestNAT03SanitizerRowMovesItsClaimedAxis` were shown to subsume the enumerating control entirely — it does not, since it hardcodes one row |
| Suite is 188 s, 42 % of `session` is one frozen enumeration | **HIGH** | `/usr/bin/time` on `go test ./...`; per-test timings from `-v` | Machine variance; this is one Apple-arm64 host, single run, no p50/p95. The ratio is more robust than the absolute |
| The combinatorial projections (~25 min, ~75 min) | **LOW-MEDIUM** | Extrapolated from the real generator's loop structure at 0.47 s/program | A different M003 grammar shape. If arithmetic lands as a separate operation kind rather than widening the chain alphabet, growth is additive, not exponential — this is the single biggest uncertainty in §6 |
| Hybrid front-loaded phase beats a dedicated debt phase | **MEDIUM** | Cluster analysis plus the 0-for-2 base rate on Cluster A attachment | Evidence that M003's first phase cannot absorb the six-emitter deletion without blocking on type-system work I have not examined; I did not read `cgen`'s six emitters |
| Metamorphic relations are the best coverage-per-second buy | **MEDIUM-HIGH** | External literature + the fact that one already ships unnamed | Discovering that Lang's semantics make a relation like unreachable-arm insertion non-trivially *unsound* (e.g. exhaustiveness interactions), which would raise the cost |
| "Over-invested in ceremony, correctly invested in architecture" | **MEDIUM** | This is a judgment, argued both directions above | A user. The entire cost-benefit inverts the moment someone other than the author depends on the compiler |

**Two things I did not verify and a reader should not assume I did:** I did not read the
six `cgen` single-function emitters, so my Cluster A cost estimate is structural, not
measured; and I did not time `go test -race ./...`, so the true per-commit CI cost is
higher than 188 s by an unmeasured factor (typically 2–5×).

---

## Sources

- Le, Afshari, Su. *Compiler Validation via Equivalence Modulo Inputs.* PLDI 2014 — https://fm.csl.sri.com/SSFT14/PLDI14-Orion.pdf
- Marcozzi, Tang, Donaldson, Cadar. *Compiler Fuzzing: How Much Does It Matter?* OOPSLA 2019 — https://dl.acm.org/doi/10.1145/3360581 · PDF: https://srg.doc.ic.ac.uk/files/papers/compilerbugs-oopsla-19.pdf
- Lopes, Lee, Hur, Liu, Regehr. *Alive2: Bounded Translation Validation for LLVM.* PLDI 2021 — https://users.cs.utah.edu/~regehr/alive2-pldi21.pdf · limits: https://github.com/AliveToolkit/alive2
- Monniaux, Boulmé. *The Trusted Computing Base of the CompCert Verified Compiler.* ESOP 2022 — https://arxiv.org/abs/2201.10280
- Leroy. *Formal Verification of a Realistic Compiler.* CACM — https://cacm.acm.org/research/formal-verification-of-a-realistic-compiler/
- Just, Jalali, Inozemtseva, Ernst, Holmes, Fraser. *Are Mutants a Valid Substitute for Real Faults in Software Testing?* FSE 2014 — https://homes.cs.washington.edu/~mernst/pubs/mutation-effectiveness-fse2014.pdf
- Papadakis et al. *Are Mutation Scores Correlated with Real Fault Detection?* ICSE 2018 — https://dl.acm.org/doi/pdf/10.1145/3180155.3180183
- Chen, Kuo, Liu, Poon, Towey, Tse, Zhou. *Metamorphic Testing: A Review of Challenges and Opportunities.* ACM Computing Surveys 51(1), 2018 — https://dl.acm.org/doi/10.1145/3143561
- Smith, Barr, Le Goues, Brun. *Is the Cure Worse Than the Disease? Overfitting in Automated Program Repair.* FSE 2015 — https://clairelegoues.com/assets/papers/SmithOverfitting2015.pdf
- Ye, Martinez, Monperrus. *Automated patch assessment for program repair at scale.* EMSE 2020 — https://link.springer.com/article/10.1007/s10664-020-09920-w
- *Contrasting test selection, prioritization, and batch testing at scale.* EMSE 2024 — https://link.springer.com/article/10.1007/s10664-024-10589-8
- Rapita Systems. *MC/DC Coverage.* — https://www.rapitasystems.com/mcdc-coverage
- *Modified Condition/Decision Coverage in the GNU Compiler Collection.* arXiv 2501.02133 — https://arxiv.org/pdf/2501.02133
- *Skeletal Program Enumeration for Rigorous Compiler Testing.* PLDI 2017 — https://arxiv.org/pdf/1610.03148

### Project artifacts read (all claims above trace to one of these)

`.planning/PROJECT.md` · `.planning/LANGUAGE-MATURITY.md` · `.planning/STANDING-VERDICTS.md` ·
`.planning/RETROSPECTIVE.md` · `.planning/ROADMAP.md` ·
`.planning/milestones/M002-MILESTONE-AUDIT.md` · `.planning/milestones/M001-REQUIREMENTS.md` ·
all seven `.planning/milestones/M002-phases/*/PHASE-*-DEBT.md` and the M001 `*-DEBT.md` set ·
all thirteen `*-VALIDATION.md` files · `.planning/milestones/M002-phases/09-*/09-VERIFICATION.md` ·
`testdata/phase6/{README,*.lang}` · `testdata/phase13/{README,HELDOUT.sha256,*.lang}` ·
`internal/compiler/session/session_phase5_alias.go` · `…/session_phase5_alias_test.go` ·
`…/session_phase5.go` · `…/session_phase5_test.go` · `…/session_phase5_corpus.go` ·
`…/session_phase5_corpus_test.go` · `…/session_phase6_injectors_test.go` ·
`…/session_test.go` (`TestDebtRegistersAreWellFormed`) · `…/qlt02_budget_manifest.json` ·
`…/risk_lanes.json` · `cmd/lang-repair/{repair_test.go,antitheater_test.go}` ·
`.github/workflows/ci.yml`
