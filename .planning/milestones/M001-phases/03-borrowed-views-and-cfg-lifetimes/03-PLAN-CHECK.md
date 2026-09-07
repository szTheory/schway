# Phase 3 Plan Check — Pre-Execution Verification

**Verified:** 2026-09-04 (pre-execution, goal-backward)
**Plans checked:** 03-01 through 03-07, 03-VALIDATION.md, against 03-CONTEXT.md, 03-RESEARCH.md, 03-PATTERNS.md, ROADMAP.md §Phase 3, 02-DEBT.md

## Verdict: **PASS_WITH_CONCERNS**

The seven plans are individually well-formed, requirement-covered, and Nyquist-compliant at the
plan level (each ~3 tasks, exact-name gated, tracer-first). No blocker was found in requirement
coverage, task completeness, dependency acyclicity, or `assert-go-tests.sh` usage. The
concern is structural (phase scope) and in two specific technical sub-decisions the planner
itself flagged as its weakest points. None of these individually block execution of Wave 1, but
the scope finding should be decided by the user before committing to all seven waves.

---

## Adjudication of the Three Named Concerns

### 1. Scope — verdict: **should be split; the planner's own named cut is correct**

This is the largest phase yet by every measure the planner itself lists: a first CFG
representation, two new surface-syntax forms (arm-bodies, `borrow mut`), a relaxed return-type
invariant, three new packages (`pathoracle`, `originvalidate`, `debugmap`), two new CLI
subcommands, and a rewrite of loan liveness in two independent admission layers. Per-plan scope
is fine (2-3 tasks each, bounded files); the problem is the **phase-level wave count and coupling
shape**, not any individual plan.

Concretely:

- The plan set is a **strictly sequential 7-wave chain** (`03-02→03-01`, `03-03→03-02`,
  `03-04→03-03`, `03-05→03-04`, `03-06→03-05`, `03-07→03-06`). Phase 2's own retrospective
  (`02-DEBT.md` "Why this was closed rather than fixed") is explicit that **each additional wave
  found defects introduced by the previous wave's own fix**, and that wave 6's blocker was
  *caused by* wave 5's repair. A 7-wave sequential chain of comparably novel work (new CFG,
  new liveness algorithm in two places, a new admission package, a new syntax form) compounds
  that exact risk further than Phase 2's 7 plans did, with no intermediate ship/verify gate
  between wave 1 and wave 7.
- The planner's own research (Q1) already names this as "the phase-split risk to flag
  explicitly," and 03-RESEARCH.md's own Architectural Responsibility Map and Summary treat
  OWN-03 (branching + liveness + conflict + oracle) and OWN-04 (public origins + separate
  compilation) as two independently motivated, independently testable feature sets — which is
  exactly the 03-01..05 / 03-06..07 cut the planner names in the prompt.
- **Real parallelism is being left on the table by the chosen dependency shape, not just by
  phase boundary.** 03-06 (public borrow origins) does not read or write anything from the CFG
  work at all — no `Block`, `Edge`, `LoanEndpoint`, `pathoracle`, or liveness fixpoint code
  appears anywhere in 03-06's action text, files, or must_haves. It depends only on
  `check.go`'s existing `sameType`/`checkLinear` seam and `corevalidate.go`'s existing replay
  dispatch — both of which exist *before* Phase 3 starts. Yet 03-06 is pinned
  `depends_on: ["03-05"]`, artificially serializing after the entire OWN-03 track (CFG lowering,
  exclusive loans, liveness rewrite ×2, path oracle) completes. This is not required by any
  file overlap: 03-06's `files_modified` (syntax, ast, core, check, a new `originvalidate`
  package, evidence, cmd/lang) barely intersects 03-01..05's touched files, and where it does
  (`check.go`, `core.go`) the changes are additive in different areas (return-type relaxation
  and a new `PublicOrigin` fact vs. CFG blocks/liveness).

**Recommendation:** Split the ROADMAP phase into two, following the planner's own named cut:
Phase 3a = 03-01..05 (OWN-03, the CFG/liveness/conflict/oracle track) and Phase 3b = 03-06..07
adapted (OWN-04, public origins/separate compilation, plus the closing gate — which necessarily
needs both tracks' controls and so stays last). This gives an intermediate verify/ship gate after
the higher-risk, more novel CFG work lands, before OWN-04's independent work is added on top —
directly addressing the "layered gates pay for themselves" lesson CONTEXT.md itself states.
If the user prefers to keep one phase (acceptable — ROADMAP already frames Phase 3 as one unit
with four criteria spanning both requirements), then **at minimum change 03-06's `depends_on` to
`["03-01"]`** and run 03-02..05 (OWN-03) and 03-06 (OWN-04) as parallel wave-2 tracks, joining
only at 03-07 which already needs both. This is a WARNING, not a blocker, because the plans as
written will still execute correctly in strict sequence — they simply spend more wall-clock/
review-wave risk than necessary.

### 2. The `ability.Derive` extension for match-arm-body borrowing — verdict: **holds, with a caveat**

**What's being added:** `internal/compiler/ability/ability.go`'s `deriveAt` switch currently
handles exactly four constructors (`Byte`, `Buffer`, `Box`, `Pair`) and fails closed on any other
(`default: unknown type constructor`). Plan 03-01 Task 2 adds a fifth case: a declared,
**field-less** nominal `DataDecl` (the closed-variant enum type already used by `match`, e.g. a
`flag` type with `On`/`Off` alternatives — confirmed by `ast.go:17-26`, `DataDecl{Name,
Alternatives []Alternative}` carries no payload fields) is treated as a sealed structural leaf
granting all five abilities, exactly parallel to how `Byte` is already treated (`ability.go:108-115`,
all-true `abilitySet`).

**Assessment against the sealed design:** `ability.go`'s own doc comments state the seal is
about *external forgeability* — "there is no production API for supplying or combining
arbitrary ability masks" and `combineStructural` is "deliberately package-private." Adding a
new case to the internal `deriveAt` switch, done by the executor inside the package (not via an
external caller reaching in), does not violate that seal: no caller outside `ability` gains the
ability to manufacture or combine masks; the new case is a hardcoded, package-internal rule of
exactly the same shape as `Byte`'s. It also does not create an authority-escalation path: a
field-less nominal type carries no data through which a false ability claim could leak (there is
nothing to `copy`/`share` incorrectly — the value is a bare discriminant), so granting all five
is the same reasoning `Byte` already uses. It does not break the exhaustive oracle's mechanism
either: `TestPairAllAbilityMasks`-style enumeration and `TestArbitraryMasksRemainTestPrivate`
(which parses `ability.go` and fails on any exported `*mask*`/`*combine*` symbol) both continue
to hold regardless of how many constructor cases exist in `deriveAt`.

**The caveat:** the plan says "re-run the exhaustive mask suites" for this addition but does not
require a **new** exhaustive test specifically covering the new nominal-leaf case (the existing
`TestBoxAllAbilityMasks`/`TestPairAllAbilityMasks` enumerate structural combinators, not leaf
constructors — `Byte`'s leaf case has no dedicated exhaustive test either, because it has no
inputs to vary). Given D-10's standing rule ("ask what inputs a green property test actually
reaches"), the plan should add one explicit test asserting the nominal-leaf case grants exactly
the five-of-five mask and is unaffected by alternative count/naming — this is a small, cheap gap,
not a design flaw. **A smaller alternative exists** and is worth naming: rather than extending
`ability.Derive`'s constructor switch, the arm body could borrow a synthetic wrapper type (e.g.
treat the scrutinee inside an arm as `Box[UnderlyingDataConstructor]`-shaped) — but this would be
a bigger footprint (inventing a synthetic wrapper concept) than the field-less-nominal-leaf case
the plan chose, so the plan's choice is in fact the smaller of the two options, not the larger.
**Not a blocker.**

### 3. OWN-04's relaxation of `checkLinear`'s return-type invariant — verdict: **fenced correctly in principle, but the plan under-specifies the discriminant and misses one call site**

**Confirmed current state:** `internal/compiler/check/check.go` enforces
`sameType(function.ReturnType, function.Parameter.Type)` **unconditionally** at two sites —
line 75 (S1/match-body admission in `Program`) and line 132-133 (`checkLinear`, returning
immediately with `type.return_mismatch` before any body analysis runs). Both are hard,
unconditional gates today.

**What the plan does:** 03-06 Task 1's action text is explicit that the relaxation is scoped:
"keeping the existing return-type identity rule for ordinary functions and relaxing it only for
the borrowed-view case, where the underlying type must still match and the origin annotation is
mandatory — a borrowed return with no declared origin is a causal span-bearing rejection, not an
inference." This is the right shape: identity is not dropped, it is redefined (underlying type
of the borrowed view, not the wrapper) for exactly one new return-type spelling, and the
must-declare-origin requirement prevents the relaxation from becoming silent inference. No
Phase 2 truth is implicated by this change: none of the 14 Phase 2 verification truths
(`02-VERIFICATION.md`, referenced in `03-CONTEXT.md`'s canonical refs) concerns return-type
identity; Phase 2's abilities/transfer/copy truths are orthogonal to this gate.

**The gap:** the plan's action text does not name **the discriminant test** — how `checkLinear`
(and the S1 match-body check at line 75, which the plan's text never mentions at all) will
decide "is this return type the borrowed-view case" versus "ordinary." Presumably a new
`TypeRef` spelling (`borrow(...)  View[T]`-shaped, per RESEARCH Q5) makes this a syntactic
question resolvable before `sameType` runs, but the plan leaves that decision to the executor
without stating it, and **never says whether a match-body (`S1`) function may also return a
borrowed view**, only ever discussing linear bodies. If a match-body function cannot return a
borrowed view, that's a fine and defensible scope limit — but it should be stated explicitly in
the plan (as a prohibition or a named scope note) rather than left as an implicit omission,
because a match-body function attempting a borrowed return today would hit the *unrelaxed* line-75
gate and get a generic `type.return_mismatch` with no origin-specific diagnosis, which is a
worse failure mode than the plan's stated "causal span-bearing rejection" standard for the
linear-body case. **Recommend**: add one sentence to 03-06 Task 1's action naming the exact
discriminant (e.g. "a `ReturnType` whose constructor is the reserved `View` wrapper") and
stating explicitly whether match-body (S1) functions are in scope for a borrowed return this
phase. This is a **WARNING**, not a blocker — the must_haves truth and fixture (`public_view.lang`)
make the linear-body case testable and falsifiable regardless of this omission, and the gap is a
specification-clarity issue for the executor, not a hole in what gets verified.

---

## Standard Plan-Quality Gates

### Requirement coverage — PASS
Every plan's frontmatter `requirements` field cites OWN-03 and/or OWN-04. ROADMAP's four success
criteria all have owning plans with a falsifiable `must_haves.truths` entry:
- C1 (checker/oracle agree) → 03-05, `TestOracleAgreesWithProduction` + mutation-kill register.
- C2 (edge-specific last use; omission detected) → 03-03, the accept/reject fixture pair whose
  verdicts both flip under the seeded uniform-join fault.
- C3 (public origin names all origins/modes) → 03-06 Task 1.
- C4 (separate compilation rejects stale/omitted/impossible) → 03-06 Task 2, three named codes.

No requirement or ROADMAP criterion is silently dropped.

### Task completeness — PASS
All 21 tasks across the seven plans carry `<files>`, `<action>`, `<verify>` with `<fails_when>`,
and `<done>`. No vague "implement X" actions were found; every action names exact files, exact
seams (`check.go:249-318`, `corevalidate.go:261-308`, etc.), and exact test names.

### `<automated>` commands and `assert-go-tests.sh` — PASS
Every `<automated>` block across all 21 tasks uses
`sh scripts/assert-go-tests.sh <package> <ExactTestName...>` with exact, spelled-out test names
(no regex, no substring). `assert-go-tests.sh` itself (read in 03-PATTERNS.md's analysis) already
fails closed on a missing/renamed target via `-list .` exact-match discovery, so this satisfies
the fail-closed requirement without any plan needing to reimplement it.

### D-12a four-site coverage for new `OperationKind`s — PASS
The only two new `OperationKind` additions this phase are the arm-body lowering (03-01, which
does not add a new kind but does add new consumers of blocks/edges to interp+cgen) and
`OpBorrowExclusive` (03-02). Plan 03-01 Task 2's action explicitly updates `interp.runLinear`'s
dispatch and `cgen` lowering "together with check and corevalidate," citing D-12a by name. Plan
03-02 Task 1's action explicitly updates all four sites for `OpBorrowExclusive` and the
`<fails_when>` for that task includes "interp or cgen was not updated alongside check and
corevalidate." No `OperationKind` addition anywhere in the plan set skips `interp`/`cgen`.

### Mutation-Kill / Generator-Reachability / Shipped-Binary registers — CONCRETE, not aspirational
The Mutation-Kill Register in 03-VALIDATION.md names the exact production hunk to revert per row
("the 03-03 backward-fixpoint liveness pass," "the 03-04 validator derivation," "the digest
comparison," "the propagation counters") and the expected failure mode, and the owning plans
(03-05 Task 2, in particular) specify the *procedure* concretely: "in a detached throwaway
worktree, revert the 03-03 production liveness hunk, run the differential, capture the failing
output, and paste it into the summary." This is method, not aspiration — it cannot name a git SHA
before the commit exists, but it does name the exact hunk by owning-plan reference and requires
verbatim captured output, which is the right standard given D-09's text. The
Generator-Reachability Register's "Must reach / Known not to reach" columns are pre-execution
templates to be filled from in-code comments (03-05 Task 3 requires the comment; the register
itself is explicitly marked "completed from those comments during execution" in
03-VALIDATION.md's own text) — this is the correct sequencing, not a shortcut. The
Shipped-Binary Register names exact CLI commands per row and is duplicated in each owning
plan's `<verification>` checklist. All three registers pass the "concrete, not aspirational"
bar Phase 2's own failure mode requires.

### Phase 1 goldens and schema versioning (D-13) — PASS
Q8's "no `lang.core/2`, additive `omitempty` fields" recommendation is followed uniformly: every
plan's `prohibitions` list forbids a non-`omitempty` field and forbids the schema bump. Plan
03-06 Task 3 closes the loop with a **key-by-key parsed-JSON assertion covering every field
introduced across all six plans** (not just its own), which is the correct aggregate check since
`evidence.manifestID`'s `CoreDigest` hashes the full serialized `core.Program` — if any field
anywhere lacked `omitempty` and leaked into a Phase 1/2 program's bytes, this single
plan-03-06-Task-3 assertion would catch it regardless of which earlier plan introduced the field.
Because no plan proposes a schema bump to `lang.core/2` (Q8's recommended path is followed
throughout), `evidence.manifestID`'s existing two identity structs (`/0` and `/1`) do not need a
third branch — the "per-version identity structs" concern is resolved by the plans' own choice
not to version-bump, not by an unaddressed gap.

### `_LANG_` sequencing (D-06 / D-02-05) — PASS
Plan 03-01 Task 1 is the literal first task of Wave 1, is required to be its own standalone
commit with an empty `testdata/` diff, and precedes Task 2 (the first task that produces any new
C artifact). This is exactly the ordering D-06 requires.

### Debug-lineage scope fence and D-04 non-ship clause — PASS
Plan 03-07 Task 1 requires the package doc comment to name all four rejected steps (2's native
half, 3's panic/segfault half, 5, 6) and requires: "If any of the caps, the schema, the timeout,
or the honest-absence reporting cannot be met, do not ship the experiment: say so in the summary
and remove it, per D-04's own terms." This is D-04's non-ship clause verbatim, present as an
executable instruction, not just a citation.

### Carried debt placement (D-02-01..09) — PASS
03-VALIDATION.md's Debt Closure Register maps all nine items to specific owning plans
(D-02-01/04/06/08 → 03-07; D-02-02 → 03-06; D-02-03 → 03-03 checker half + 03-04 validator half;
D-02-05/07 → 03-01; D-02-09 → 03-02), matching D-08's "fold into whatever plan already touches
its file; do not create a debt-cleanup plan" instruction. No standalone debt-cleanup plan exists.

### Wave/dependency structure — WARNING (see Concern 1 above)
Acyclic and internally consistent (each `depends_on` references an existing prior plan number,
wave numbers match `max(deps)+1`), so no dependency-correctness blocker exists. But the chain is
more sequential than the work requires: 03-06 (OWN-04) has no technical dependency on 03-02..05
(OWN-03's exclusive loans, liveness rewrite, or path oracle) and could run as a parallel wave-2
track from 03-01. See phase-split recommendation above.

---

## Issues Found

```yaml
issues:
  - dimension: scope_sanity
    severity: warning
    description: >
      Phase 3 is a strictly sequential 7-wave chain doing CFG introduction, two liveness
      rewrites, a new syntax form, three new packages, and two CLI subcommands, with no
      intermediate verify/ship gate — the exact compounding-wave-defect shape Phase 2's own
      retrospective names as its root cause (waves 5-6). The planner's own research and the
      prompt's own naming both identify a clean 03-01..05 (OWN-03) / 03-06..07 (OWN-04) cut.
    plan: "all"
    fix_hint: >
      Split into two ROADMAP phases at the named cut, or at minimum change 03-06's depends_on
      from ["03-05"] to ["03-01"] and run OWN-03 (03-02..05) and OWN-04 (03-06) as parallel
      wave-2 tracks joining only at 03-07.

  - dimension: key_links_planned
    severity: warning
    description: >
      03-06 Task 1 relaxes checkLinear's unconditional sameType(ReturnType, ParameterType) gate
      for the borrowed-view case but does not name the exact discriminant used to distinguish a
      borrowed return from an ordinary one, and never mentions whether the S1/match-body sameType
      gate at check.go:75 is also relaxed or intentionally left out of scope for a borrowed
      return on a match-bodied function.
    plan: "06"
    task: 1
    fix_hint: >
      State the exact ReturnType discriminant (e.g. a reserved View/borrow wrapper constructor)
      and explicitly state whether match-body (S1) functions are in scope for a borrowed return
      this phase; if out of scope, add it as a named prohibition rather than a silent omission.

  - dimension: verification_derivation
    severity: info
    description: >
      Plan 03-01 Task 2 extends ability.Derive with a new field-less-nominal-leaf case but only
      requires re-running the existing exhaustive mask suites, not a new exhaustive test
      specifically enumerating the new leaf case's mask output (parallel to how Byte's leaf case,
      having no inputs to vary, needs none — but the new nominal case does vary by alternative
      count/naming and currently has no dedicated exhaustive assertion).
    plan: "01"
    task: 2
    fix_hint: >
      Add one small exhaustive/property test asserting the nominal-leaf case grants exactly the
      five-of-five mask regardless of alternative count or naming, matching the D-10 standard
      applied elsewhere in this same plan set.
```

## Summary

No blocker was found. The plan set correctly implements D-06, D-08, D-09 (methodologically),
D-10, D-12/D-12a, D-13, and D-04's non-ship clause as executable instructions rather than
citations, and all four ROADMAP success criteria have falsifiable, owned truths. The two
technical sub-decisions the planner flagged as its own weakest points (the `ability.Derive`
extension and the `checkLinear` return-type relaxation) both hold up against the sealed-ability
design and Phase 2's invariants, with one small gap each (noted above as WARNING/INFO, not
BLOCKER). The one finding worth the user's attention before executing all seven waves is
structural: this phase is large enough, and its own dependency chain serializes enough
independent work, that a two-phase split at the planner's own named cut is the safer path given
Phase 2's documented compounding-wave-defect history.

---

## Resolution — 2026-09-04 (post-check, before execution)

All three issues above are **closed in the plans**. This section is authoritative over the
`issues` block; do not re-raise them.

| Issue | Severity | Resolution |
|---|---|---|
| `scope_sanity` — 7-wave sequential chain, no intermediate gate | warning | **Not split.** 03-07 carries the gate for both tracks, so a clean split needs replanning, not relabeling. Answered instead by the checker's own minimum remedy plus one addition: 03-06 re-pinned `depends_on: ["03-05"] → ["03-01"]` (parallel wave-2 track), and a **mandatory mid-phase gate after 03-05** — independent review and goal-backward verification scoped to OWN-03, criteria 1 and 2, before wave 6 starts. Recorded in ROADMAP §Phase 3 and in 03-05's `<verification>`. |
| `key_links_planned` — 03-06 Task 1 under-specifies the return discriminant and omits the S1 case | warning | 03-06 Task 1's action now names the discriminant explicitly (a `TypeRef` carrying a `borrow(path)` / `borrow mut(path)` origin annotation, resolved before `sameType` runs) and fences the S1 case: the `check.go:75` gate stays unconditional, a match-bodied borrowed return is out of scope this phase and is rejected with a named span-bearing cause rather than a bare `type.return_mismatch`. Added as a `must_haves` prohibition, a behavior line, a new required test (`TestMatchBodyBorrowedReturnRejectedWithCause`), and two `fails_when` clauses. |
| `verification_derivation` — no exhaustive test for the new nominal-leaf ability mask | info | 03-01 Task 2 now requires `TestNominalLeafAbilityMaskIsExhaustive`, enumerating declared field-less nominal types across alternative counts and namings and asserting the five-of-five mask, with a `fails_when` clause that rejects a single-shape assertion. |

Wave structure after the re-pin: 1 → `03-01`; 2 → `03-02` and `03-06` in parallel;
3 → `03-03`; 4 → `03-04`; 5 → `03-05` (**mid-phase gate**); 6 → `03-07`.
