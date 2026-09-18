# Requirements: M003 Computation and Honest Instruments

**Defined:** 2026-09-17
**Core Value:** Give an AI agent and a human reviewer the shortest reliable path
from intent to sound, reproducible evidence without wasting iteration time or
hiding runtime costs.

**Derived from:** `.planning/research/M003/` — six parallel research dimensions
plus an adversarial arbitration pass that resolved five inter-document
contradictions against the tree and spot-checked eight high-consequence claims
(all eight verified; three stronger than originally reported).

**Milestone thesis:** M001 proved one meaning survives lowering. M002 proved it
survives a function boundary. M003 proves it survives *computation* — the first
value Lang creates rather than moves — on instruments that cannot report green
for work that is merely wired.

## The Governing Gate

**Constructibility precondition.** No requirement below is admitted into a phase
unless its `.lang` fixture is checked in *first*, as a refused frontier fixture
with its refusing diagnostic pinned by a test. The phase gate is that the pinned
diagnostic moved.

This is the only gate that fires *before* the work. It exists because four
separate instances of one failure mode shipped in M001/M002 — DX-06's blame
subsystem, `-flto` inertness on the multi-function corpus, three VALIDATION
`-run` patterns matching zero tests, and a mutation law silently skipping rows.
Applied on day one it would have refused comparison operators, which cannot be
branched on and would have become the fifth instance.

## M003 Requirements

### Evidence Instruments

- [x] **EVD-01**: A CI lint fails when any `.planning/**` verification document
      cites a `go test -run` pattern that resolves to zero tests, or a command
      cell elided with `…`. Three such patterns exist today and exit 0.

- [ ] **EVD-02**: Every requirement and success criterion carries a grade from
      the closed vocabulary `DEFINED | WIRED | REACHABLE | EXERCISED |
      MUTATION-KILLED`, and is satisfiable only at EXERCISED or above.

- [x] **EVD-03**: A claim that is built but structurally unreachable is recorded
      in `.planning/UNREACHABLE-CLAIMS.md` with its unblocking trigger, rather
      than graded as satisfied or silently downgraded.

- [x] **EVD-04**: No test skip or row exclusion outlives the trigger it cites; a
      guard fails when a cited gate has already closed.

- [x] **EVD-05**: The mutation axis-movement law has exactly one implementation
      and zero per-row exclusions.

- [x] **EVD-06**: `LANGUAGE-MATURITY.md`'s own counts are machine-checked by a
      test that independently re-derives them from the tree (never by running
      the document's own embedded re-verify greps), so the file cannot go
      stale unnoticed. (Closed 2026-09-18 by 14-05: corrected 32 guards to the
      measured 22; see `.planning/phases/14-evidence-instrument-and-honest-scoping/14-05-SUMMARY.md`.)

- [x] **EVD-07**: PROJECT.md's DX-06 claim and NAT-07 `-flto` bullet state what
      the evidence supports, and the `-flto` multi-function inertness has a debt
      row with an owning phase.

- [x] **EVD-08**: Suite wall-clock is an observed row in the budget manifest
      with a recorded baseline. (Closed 2026-09-18 by 14-05: a real clean
      cold `go test ./...` run measured 191.89s on `machine:4797d76b7863`,
      recorded as the `suite_wall_clock_ns` row; documents previously
      claimed ~60s in five archived VALIDATION.md files, still unfixed and
      enumerated in the 14-05 SUMMARY.)

### Observability — Event Identity

- [ ] **OBS-01**: Two activations of the same callee through a shared-leaf
      diamond produce distinct event identities.

- [ ] **OBS-02**: A call emits an observable event, so the causal edge between
      caller and callee is visible rather than inferred.

- [ ] **OBS-03**: Event identity is re-derived independently by a non-importing
      peer over its own traversal.

- [ ] **OBS-04**: The `/0` and `/1` execution-document bytes remain frozen; the
      new required field lands under a `/2` schema bump.

### Native Emission

- [ ] **NAT-08**: One emission law lowers every admissible program; the three
      superseded single-function emitters are deleted in the same commit that
      flips dispatch.

- [ ] **NAT-09**: The three emitter families with no M003 consumer are formally
      cut by a recorded amendment naming their landing milestone and the
      `-flto` consequence — not deferred a third time.

- [ ] **NAT-10**: A re-invoking multi-function fixture is compared across
      interpreter, `-O0`, `-O3`, and `-O3 -flto`.

### Types

- [ ] **TYP-01**: A function's declared return type may differ from its
      parameter type, and a program relying on that checks, runs, and lowers.

- [ ] **TYP-02**: A call whose argument type does not match the callee's
      declared parameter type is refused by name, from a `.lang` fixture.

- [ ] **TYP-03**: A return contract that cannot be represented is refused by
      name, from a `.lang` fixture.

- [ ] **TYP-04**: Drop obligations and freshness are derived from the return
      type's own abilities, independently at each admission layer — not
      inherited from the parameter's.

- [ ] **TYP-05**: `lang-repair` reaches `repaired` on `use_matching_argument`
      against a sealed held-out fixture. (Closes DX-07 / D-13-10a.)

### Control — Branch on a Computed Value

Gated on spike **S-010**. If S-010 shows a loan crossing a branch point forces a
redesign rather than an extension, this category is cut and replaced by
arithmetic operators, and branching moves to M004.

- [ ] **CTL-01**: A branch discriminates a value the function computed, not only
      the function's own parameter.

- [ ] **CTL-02**: A program that calls a `Result`-returning callee and matches
      on the result agrees across all five axes.

- [ ] **CTL-03**: An arm returns a destructured payload place, and a seeded
      wrong-slot write is observed on a comparator axis — constructing D-12-43
      rather than ratifying it as unconstructible.

### Values

- [ ] **VAL-01**: A numeric literal can be written, checked, interpreted, and
      lowered.

- [ ] **VAL-02**: Every operation kind is handled at all six dispatch sites,
      proven by the exhaustive-dispatch control.

- [ ] **VAL-03**: A literal-bearing program agrees across interpreter, `-O0`,
      `-O3`, and `-O3 -flto`.

### Quality and Frontier

- [ ] **QLT-10**: Phases 07, 08, and 11 reconcile against the post-M003 surface,
      with EVD-01's lint doing the finding; no VALIDATION file remains `draft`.

- [ ] **QLT-11**: A refused frontier fixture for `examples/checksum.lang` is
      checked in with its refusing diagnostic pinned, and the milestone moved
      that diagnostic.

- [ ] **QLT-12**: The enumerated-closure proof is content-addressed rather than
      re-run every commit. (Currently 58.33s, 30% of suite wall-clock, re-proving
      a frozen 112-program closure.)

### Agent Loop

- [x] **DX-08**: Structurally different defective programs produce
      distinguishable diagnostics. Today three such programs yield identical
      diagnostic IDs and an identical result ID, so an editing agent gets no
      convergence signal.

- [x] **DX-09**: An `unrepairable` verdict explains itself; the `diagnosis` field
      is never empty.

### Process

- [x] **PRC-01**: Every debt item names an owning phase when recorded, and a
      well-formedness gate fails otherwise — mechanizing D-10-60, which prose
      did not.

- [ ] **PRC-02**: M003 does not close with more than 5 open, unowned debt items.

## Deferred — Named Landing, Not Dropped

| Requirement | Landing | Reason |
|---|---|---|
| Comparison operators, `Bool` | M004 | Inert until CTL lands; shipping earlier repeats DX-06 |
| `OpBinary` / arithmetic operators | M004 (or replaces CTL if S-010 fails) | Natural successor to VAL |
| Loops, iteration, back edges | M004 | `pathoracle` refuses every CFG cycle by name; needs a successor spike as a hard entry gate |
| Arity-N, aggregates | M005 | Takes the `pathoracle` case space ~12 → ~108 |
| DX-06 / B1 blame | M006 | Needs a user-declared contract field the declaring function cannot verify — that is separate compilation |
| D-13-34 held-out pair repair | Grammar-widening phase | Admissible space is 112 programs; "structurally distinct" is cosmetic at this maturity |
| DX capability manifest, rendered diagnostics, repair provenance | Later | Contingent and decreasing in value; polishing a surface this milestone changes |

## Out of Scope

| Feature | Reason |
|---|---|
| `if` / `else` as surface syntax | A second control-flow law over a construct not yet doing the first law's job; CTL generalizes the one law instead |
| Signed integers, multiple widths, floats, a numeric tower | One fixed-width unsigned type makes signed-overflow UB unconstructible rather than merely avoided |
| `-fwrapv` | Since GCC 8 it disables `-fsanitize=signed-integer-overflow`, blinding the existing UBSan lane |
| Content-addressed digest for event identity | Categorically wrong here — the two colliding activations are content-identical, so it collides by design |
| Type inference | Inferred signatures make declared-contract blame incoherent; an anti-feature for a language whose product is attribution |
| Separate compilation, modules, a fourth peer, a second comparator | Unchanged from M002 |
| Nyquist reconciliation before EVD-01's lint exists | The lint must do the finding |

## Traceability

Mapped during roadmap creation, 2026-09-17. Phase numbering continues from M002
(which ended at Phase 13); M003 runs Phases 14-20.

| Requirement | Phase | Status |
|-------------|-------|--------|
| EVD-01 | Phase 14 | Complete |
| EVD-02 | Phase 14 | Pending |
| EVD-03 | Phase 14 | Complete |
| EVD-04 | Phase 14 | Complete |
| EVD-05 | Phase 14 | Complete |
| EVD-06 | Phase 14 | Complete |
| EVD-07 | Phase 14 | Complete |
| EVD-08 | Phase 14 | Complete |
| OBS-01 | Phase 15 | Pending |
| OBS-02 | Phase 15 | Pending |
| OBS-03 | Phase 15 | Pending |
| OBS-04 | Phase 15 | Pending |
| NAT-08 | Phase 16 | Pending |
| NAT-09 | Phase 16 | Pending |
| NAT-10 | Phase 15 | Pending |
| TYP-01 | Phase 17 | Pending |
| TYP-02 | Phase 17 | Pending |
| TYP-03 | Phase 17 | Pending |
| TYP-04 | Phase 17 | Pending |
| TYP-05 | Phase 17 | Pending |
| CTL-01 | Phase 18 | Pending |
| CTL-02 | Phase 18 | Pending |
| CTL-03 | Phase 18 | Pending |
| VAL-01 | Phase 19 | Pending |
| VAL-02 | Phase 19 | Pending |
| VAL-03 | Phase 19 | Pending |
| QLT-10 | Phase 20 | Pending |
| QLT-11 | Phase 20 | Pending |
| QLT-12 | Phase 20 | Pending |
| DX-08 | Phase 14 | Complete |
| DX-09 | Phase 14 | Complete |
| PRC-01 | Phase 14 | Complete |
| PRC-02 | Phase 20 | Pending |

**Coverage:**

- M003 requirements: 33 total
- Mapped to phases: 33 ✓
- Unmapped: 0

**Per-phase counts:** P14 = 11, P15 = 5, P16 = 2, P17 = 5, P18 = 3, P19 = 3,
P20 = 4. No orphans, no duplicates.

**Mapping deviations from the natural category grouping** (justified in full in
`ROADMAP.md` § Requirement Coverage):

- **NAT-10 → Phase 15**, not Phase 16. NAT-10 *is* Phase 15's ratified gate
  (`multi_function_diamond_call.lang` across all four tiers). Filing it with the
  emitter port would let Phase 16 claim credit for an already-green row — the
  exact "green because it is wired" pattern this milestone retires.

- **PRC-01 → Phase 14, PRC-02 → Phase 20.** PRC-01 is an instrument and must
  exist before debt accrues; PRC-02 is a close condition adjudicable only at the
  end.

- **DX-08 / DX-09 → Phase 14.** Both are shipped-API defects measurable today
  and instances of the milestone's central theme. Note these two IDs are
  ADVERSARIAL-SYNTHESIS's DX-10 and DX-12 renumbered; that document's own
  DX-08/DX-09 (capability manifest, `surface.not_in_language`) are deferred.

**Contingency.** If spike S-010 comes back dirty, Phase 18 is cut: CTL-01,
CTL-02 and CTL-03 move to the Deferred table with an M004 landing, and new
`ARI-NN` rows for the replacement arithmetic phase are amended in during the
same pass. Coverage is re-validated at 33 rows; it does not silently drop to 30.

---
*Requirements defined: 2026-09-17*
*Last updated: 2026-09-17 after M003 roadmap creation (Phases 14-20)*
