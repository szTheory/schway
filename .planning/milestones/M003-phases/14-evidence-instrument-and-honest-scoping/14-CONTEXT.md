# Phase 14: Evidence Instrument and Honest Scoping - Context

**Gathered:** 2026-09-17
**Status:** Ready for planning

<domain>
## Phase Boundary

Every shipped claim is graded at EXERCISED or above, or names itself unreachable
with an unblocking trigger — so that **no instrument in this project can report
green for something that is merely wired.**

The subject of this phase is the **measuring instruments**, not the language.
Five deliverables: a groundedness lint over `.planning/**`, a closed grade
vocabulary, an unreachable-claims register with executed triggers, one (not two)
mutation axis-movement law, and machine-checked self-describing documents. Plus
two API-contract defects in the agent loop (DX-08/DX-09), which the milestone's
arbitration document explicitly routed here "as plans inside the evidence phase,
not as a DX phase".

**No language-surface change.** There is no arithmetic, no iteration, no `if`,
no strings, no arrays; calls are arity-1; every function's return type must equal
its parameter type. Requirements: EVD-01..EVD-08, DX-08, DX-09, PRC-01.

**The one unifying principle behind every decision below** — and the thing that
distinguishes this phase's instruments from the four that failed:

> **A grader that reads an artifact instead of executing a claim will convert an
> honest partial into a false green.**

That defect has now fired four times, each time with a different artifact:
(1) `gsd-integration-checker` marked DX-06 `✓ WIRED` and called it satisfied;
(2) `09-VALIDATION.md` graded a row green from a SUMMARY's prose while its
`go test -run` pattern matched zero tests, in a phase certified
`nyquist_compliant: true`; (3) `08-VALIDATION.md` graded rows from the *intent*
of a command rather than its resolution; (4) `TestNoNAT03RowRemainsPending`
watched a marker comment in the **production file only**, so the surviving stale
marker — which lives in the *test* file and in an *error string* — was invisible
to it.

**Every instrument in this phase therefore executes its claim.** That is the
scope anchor. An instrument that merely parses a string is out of scope even
when it would satisfy the requirement's literal text.

**Standing discipline this phase inherits (not a Phase 14 deliverable):** the
M003 Governing Gate — constructibility precondition. Phase 14 opens with its
refused frontier fixtures checked in with pinned diagnostics, and closes by
asserting those pinned diagnostics moved. For an instrument phase the
"fixture + pinned diagnostic" is realized as D-14-08's exact pinned violation
literal and D-14-22's frozen collision control.

</domain>

<decisions>
## Implementation Decisions

### Grade authority — the closed evidence vocabulary (EVD-02)

- **D-14-01:** Adopt **declared grade, mechanically capped**. The author writes a
  `Grade` cell; a Go test derives a **ceiling** from the row's own evidence cell
  and fails when `declared > derived`. Rejected: vocabulary-string validation
  only (assert the cell is a member of the closed set and nothing more) — that is
  a ~60-line clone of `debtRegisterSeverities` that would ship `gsd-integration-checker`
  writing `WIRED` again with a spellchecker attached. **No seeded fault can make a
  string-membership check go red, so it cannot carry a non-inertness proof**, and
  a control that cannot be proven non-inert is exactly what this phase exists to
  retire. Vocabulary-string validation is acceptable only as a strictly temporary
  shape migration *inside* the first commit of the capped implementation, never as
  the terminal law.
  — **Reversibility:** costly — the derivation engine and the retired `Status`
  column touch all 13 archived `*-VALIDATION.md` files plus every new one.

- **D-14-02 (the argument the plan MUST make, because it is the first objection a reviewer will raise):** `TestDebtRegistersAreWellFormed`'s doc comment draws
  a deliberate line — *"Register honesty — whether a deferral is truthfully
  described — stays a human reading and is deliberately not claimed here."*
  D-14-01 does **not** cross that line. A grade is an honesty property in
  general, but EVD-02 only needs it to be a **shape property at the bar**: the
  two grades that *satisfy* a requirement (`EXERCISED`, `MUTATION-KILLED`) are
  fully mechanically decidable from a resolving Go test name, while the three
  that refuse (`DEFINED`, `WIRED`, `REACHABLE`) are refusals, and **nobody games
  a control by under-claiming**. The cap mechanizes only the pass side and leaves
  the refusal side declarative. State this explicitly in the plan.

- **D-14-03 (the derivation ladder, `deriveGrade(row)`):**
  - Evidence cell parses to one or more exact Go identifiers → resolve each
    against `go test -list .` semantics for the named package, **reusing the exact
    algorithm in `scripts/assert-go-tests.sh:8-45`**. Unresolved ⇒ ceiling `WIRED`.
  - Resolved **and** the run record shows `"Action":"pass"` for that exact `Test`
    with no `"Action":"skip"` ⇒ ceiling `EXERCISED`.
  - The row's `Non-inertness` cell additionally names a **distinct** resolving,
    passing test ⇒ ceiling `MUTATION-KILLED`.
  - Evidence cell names an existing `testdata/` file or a CLI invocation rather
    than a test name ⇒ ceiling `REACHABLE`.
  - **Compile-time evidence** (`go build ./...` / `BUILD_OK`, as `13-VALIDATION.md`
    uses for `blameFieldWitness`) ⇒ ceiling `WIRED`, **not** EXERCISED. This arm
    is required by real corpus data and already matches that row's own prose.
  - A **call-site-count arm** so D-13-02b's classification is mechanical rather
    than prose: zero production call sites ⇒ ceiling `WIRED`.
  - `DEFINED` is the floor and is always permitted.

- **D-14-04 (verified on this tree, Go 1.24.0 — the defect is live today):**
  `go test ./internal/compiler/session/... -run TestNoSuchTestNameXYZ` prints
  `ok ... [no tests to run]` and **exits 0**. `-list` resolution is the only thing
  that closes it. Note `-list` enumerates **top-level tests only** — subtest-scoped
  claims (`13-VALIDATION.md` has several, e.g. `subtests alpha, mirror`) must name
  the **parent** test in the evidence cell and record subtest names in prose only.

- **D-14-05 (physical location — in the document, no sidecar):** the Per-Task
  Verification Map gains `| Grade |` (closed vocabulary) and `| Non-inertness |`
  (exact twin test name, or `—`), and **retires the freeform `| Status |` column**
  (`✅ green`). Frontmatter gains `evidence_vocabulary: v1` and `graded_rows: N`,
  mirroring `items: N`, whose count-vs-table-length cross-check is the single
  highest-value assertion in `checkDebtRegister`. A JSON/YAML sidecar is rejected
  on three counts: it creates a second source of truth with a sync problem (the
  documented rot mode of split RTM / `sphinx-needs`-style stores), it needs a
  parser the zero-dependency constraint disallows, and **the grade must be
  readable in the same diff hunk as the claim it grades**.
  — **Reversibility:** costly — retiring `Status` is a schema change across 13
  archived files; see D-14-19 for the ordering obligation it creates.

- **D-14-06 (back corpus — derive over all 13, enforce over none):** the
  ROADMAP's gate conflates two things; split them.
  `TestValidationRowGradesAreEarned` walks
  `phaseArtifactGlob("*", "*-VALIDATION.md")` (which already globs
  `.planning/milestones/*-phases/`) and **derivation must be total**: a historical
  row the engine cannot classify is a **hard test failure**, because that is
  precisely the ROADMAP's own signal that "the vocabulary is wrong while it is
  still a data-shape change". The **bar** (`>= EXERCISED`) is enforced only for
  phases >= 14, gated by an explicit `validationGradeBarExemptions` map keyed by
  **filename** with a dated reason string — the exact shape and review discipline
  of the existing `debtRegisterLandingPhaseExemptions` (`session_test.go:2581-2590`).
  Exemptions are **file-scoped and dated, never row-scoped**; row-scoped
  exemptions are what would collapse the whole vocabulary into hand-authoring with
  extra ceremony.

- **D-14-07 (predicted first finding — record it, do not pre-empt it):** at least
  one M001/M002 row is expected to derive **below** its shipped `✅ green`. That is
  the instrument's first real finding, **not a blocker and not a regression to
  suppress**. It is recorded as a debt row with an owning phase (PRC-01) and
  reported honestly. Precedent: D-13-33's identical instruction — "if it does,
  that is a real hole in shipped M001 evidence being surfaced, not a regression to
  suppress; escalate rather than weaken the predicate."

- **D-14-08a (honest residual, stated in the test's own doc comment — do NOT claim otherwise):** `func TestX(t *testing.T) {}` resolves, passes, never skips,
  and derives `EXERCISED`. The cap cannot see this. **`EXERCISED` certifies
  resolution and execution, not assertion strength.** Assertion strength is what
  `MUTATION-KILLED` and peer re-derivation buy. A second, partially-open gaming
  route: naming a trivially-passing twin to reach `MUTATION-KILLED` — closed only
  to the extent that the twin must be distinct from the primary and must exist.
  Both are accepted and stated, not papered over.

### The groundedness lint (EVD-01)

- **D-14-09:** Host it as a **Go test**, `internal/compiler/session/verification_groundedness_test.go`,
  package `session`, reusing `phaseArtifactGlob` and `testsupport.ProjectPath`.
  Rejected: growing `scripts/assert-go-tests.sh` into a document lint — that
  ships **two coexisting laws over one subject** (the project's own named delayed
  fuse), shell cannot reliably do `\|`-unescaping and code-span extraction, and
  the script's `--self-test` sentinel proves *selection*, not *document scanning*,
  so the existing non-inertness proof would not actually cover the new behavior.
  **Keep both artifacts with non-overlapping jobs and say so in a comment in each
  file**: the script is the *selection-time* guard for callers actually running
  tests; the lint is the *document-time* guard over `.planning/**`. Do not delete
  the script's `--self-test`.

- **D-14-10 (resolution mechanism — static, not `go test -list`):** build
  `map[importPath]map[string]bool` by walking the module with
  `go/parser.ParseFile(..., parser.SkipObjectResolution)` over every `*_test.go`,
  honoring build constraints via `go/build.Context.MatchFile`, collecting
  top-level `func (Test|Fuzz|Benchmark|Example)[A-Za-z0-9_]*` with valid
  signatures; map dirs to import paths from `go.mod`; expand `./...` over the dir
  tree. Stdlib only, zero deps, offline, hermetic, no compile.
  **Measured: 7ms** (1121 test functions, 25 packages) versus **3.35s warm** for
  whole-module `go test ./... -list .`. Ship the cheap resolver **plus one**
  accuracy control, `TestStaticTestIndexMatchesGoTestList`, which runs the 3.35s
  `-list` and asserts set equality — 1.7% of the 192.67s suite budget, and it is
  the instrument that keeps the instrument honest. Note the research doc's "146
  shell-outs" framing overstates even the `-list` path; whole-module `-list` is a
  single invocation. If wall-clock pressure arrives, content-address the control
  under QLT-12 rather than deleting it.

- **D-14-11 (file scope — by illocutionary role, not filename):**
  - **Tier A — ASSERTING, enforced.** Any `.planning/**/*.md` whose basename
    matches `(VALIDATION|VERIFICATION|SUMMARY|UAT|MIDPHASE-GATE)`, **plus** any
    other `.planning/**/*.md` containing a table row holding both a command
    code-span and a verdict token. The verdict-token clause makes filename
    renaming useless as an evasion.
  - **Tier B — PROPOSING, exempt.** `*-RESEARCH.md`, `*-PLAN.md`, `*-CONTEXT.md`,
    `*-DISCUSSION-LOG.md`, `.planning/research/**`. Exemption must be a
    **positive frontmatter declaration** (`verification_role: proposal`), never a
    filename accident; a Tier-B file carrying a verdict token beside a command
    span is **promoted to Tier A**. This matters concretely: the RESEARCH/PLAN
    corpus contributes **~22 additional dead patterns** (e.g. `04-RESEARCH.md:560-570`
    proposes ten tests never written under those names) which are **proposals, not
    evidence** — flagging them is theater and would immediately create pressure for
    a suppression channel.
  - **Caveat created by D-14-05:** retiring the `Status` column removes the
    `✅ green` token Tier-A promotion leans on. **Key Tier-A detection on the
    canonical 10-column VALIDATION header and on the `Grade` column, not on the
    retired `Status` token.** The two changes land in the same phase and the plan
    must sequence them so neither is briefly undetectable.

- **D-14-12 (the archive crux — detection IN, editing OUT):** archived
  `*-VALIDATION.md` under `.planning/milestones/M00*-phases/**` are **in scope for
  detection and append-only for repair**. A dead row is **never rewritten in
  place** — that would be exactly the falsification this phase exists to prevent —
  and excluding archives entirely would make the lint find nothing and be theater.
  The correction lives **outside** the archive, keyed by
  `(file, line, verbatim original command)`, with a closed verdict vocabulary
  where **every verdict carries an obligation the lint itself checks**:

  | verdict | obligation the lint enforces |
  |---|---|
  | `renamed` | names a replacement command which must itself resolve to >= 1 test — you cannot launder a dead pattern into another dead pattern |
  | `superseded` | names the superseding phase + commit and a live covering command, which must resolve |
  | `obsolete-by-design` | names the deleting phase + commit and the deleted symbol; the lint asserts the symbol is **absent** from the tree. For `08-VALIDATION.md` OWN-06 this is a falsifiable positive claim that `computeLoanLastUses` does not exist under `internal/compiler/check/` — which is the "adjudication, not a rename" the research called for |
  | `under-scoped` | names the missing clause and the requirement/phase that will close it; permitted only for D-14-16 findings and must carry a landing phase like a debt row |

  **There is no inline suppression syntax at all** — no `nolint`-shaped comment,
  no per-file ignore. **This is not a baseline file:** a baseline is satisfied by
  silence; every entry here is satisfied only by a claim that can itself fail.
  Count-keyed suppression ceilings are refused outright — a count-neutral swap
  lands unnoticed, which is the documented ESLint-bulk-suppressions failure.

- **D-14-13 (register topology — TWO authored laws, TWO generated views):**
  one authored law **per subject**:
  1. `*-VALIDATION.md` rows — `Grade` / `Non-inertness`, `Status` retired (D-14-05).
  2. `*-DEBT.md` rows — `Grade` / `Witness` / closed `Landing phase` (D-14-17).

  Both enforced by tests living **in the same file as `checkDebtRegister`**,
  sharing `phaseArtifactGlob` and `testsupport.ProjectPath`.
  **`.planning/UNREACHABLE-CLAIMS.md` and `.planning/EVIDENCE-RECONCILIATION.md`
  are both GENERATED, byte-compared views** — regenerated in memory and compared,
  so a hand-edit is a red. A dead archived command therefore becomes a **debt row
  with a witness**, not a third authored register. This resolves the tension the
  two researchers could not each see: the ledger's *content* is right and its
  *authorship* moves into the existing law.
  — **Reversibility:** costly — reversing means splitting one well-formedness test
  back into three and re-homing every migrated row.

- **D-14-14 (anti-decay guards on the generated views, all fail-closed):**
  (a) frontmatter `entries: N` must match table row count, exactly as
  `TestDebtRegistersAreWellFormed` does for `items:`;
  (b) an entry whose verbatim original command no longer appears anywhere in
  Tier A is a **stale entry → FAIL**, **never auto-pruned** — auto-pruning is how
  bulk-suppression stores get zeroed without anything being fixed;
  (c) the byte-compare must additionally assert a **non-zero row count** whenever
  any register holds a qualifying row, or an empty generator and an empty
  checked-in file compare equal **vacuously**;
  (d) a CI diff rule: any commit touching
  `.planning/milestones/**/*VALIDATION*.md` or `*VERIFICATION*.md` must also touch
  the reconciliation view. "Edit the archive until it's green" then requires
  writing down, by name and date, that you did it.

- **D-14-15 (the detection rule — precise, and adversarially shaped):**
  Parse Tier-A files line-by-line. A line whose first non-space char is `|` is a
  table row. Split on `|` **not preceded by `\`**; in each cell replace
  `\|`→`|` then `\\`→`\` (GFM cell semantics). Within each cell, every
  **backtick-delimited code span** whose trimmed content matches
  `^(go test|go run|go vet|go build|grep|rg|awk|sed|git |\./?scripts/)\b` is a
  verification command. Extraction is **per-code-span, not per-cell** — row
  `11-05 T1` legitimately holds two.

  - **R1 — runnability.** After deleting every occurrence of the substring
    `/...` (the Go package wildcard), FAIL on any of: `…` U+2026, `⋯` U+22EF,
    `⋮` U+22EE, `‥` U+2025, `᠁` U+1801; the ASCII sequence `...`; a placeholder
    `<[A-Za-z][^>]*>`; or the tokens `TODO`, `TBD`, `XXX`.
  - **R2 — groundedness.** Tokenize with a stdlib POSIX-quote splitter (~60
    lines), skipping leading `VAR=value` assignments and a leading `env`. Package
    operands = tokens starting `./` or the module path. Pattern = value of
    `-run`/`-run=` (also `-list`, `-fuzz`, `-bench`). Build the **union** of
    top-level names across named packages (union matches `go test` reality), take
    the pattern's first `/`-separated segment, compile as a Go `regexp`, FAIL if
    it matches nothing.
  - **R3 — grep groundedness.** For a `grep`/`rg` command with no `|`, `>`,
    `&&`, `;` or `$(`, and a literal existing file operand: **exec it directly**
    (no shell) and require >= 1 match. Cost ~10 execs, microseconds. This is the
    **archival-breakage class** — `RETROSPECTIVE` M001 Key Lesson 6's third
    recurrence — and excluding it would mean the lint watches only one of the two
    known ways a command silently stops verifying.
  - **NO SILENT SKIPS — the single most important anti-inertness rule.** Every
    verification command is classified into exactly one of
    `{ok, R1, R2, R2b, R3, pipeline-shape-R1-only}`. A command containing
    `go test` that does not parse into (packages, pattern) is a **violation**
    (`unparseable verification command`), **not a pass**. Parser-based lints die
    by silently skipping what they cannot parse.

  **Measured false-positive rate on today's corpus: R1 zero, R3 zero.** All 11
  `…` in `11-VALIDATION.md` are inside command spans and all are genuine
  (including two **path** elisions — `…/PHASE-11-DEBT.md`, `…/11-MIDPHASE-GATE.md`
  — an unrunnable class distinct from pattern elision). `09-VALIDATION.md`'s
  single `…` is prose at line 225; `02-VERIFICATION.md`'s six are digest
  truncations in evidence cells, not command spans. A scan for ASCII `...` not
  preceded by `/` or `.` across every VALIDATION/VERIFICATION file returns **0
  hits**, so banning it costs nothing today and closes the obvious evasion.
  `PROJECT.md`, `ROADMAP.md`, `REQUIREMENTS.md`, `LANGUAGE-MATURITY.md` and
  `research/M003/**` contribute **zero** matches despite being full of `…`.

- **D-14-16 (per-branch groundedness — R2b, and the phase-sizing decision made explicitly rather than silently):** split R2's first segment on top-level `|`
  and require **every alternation branch** to match >= 1 name. This turns
  `-run 'PeerLiveness|LoanChainIndex'` from a pass into a finding and is where the
  defect mass actually is: `09-VALIDATION.md` alone yields further findings at
  lines 84, 87, 88, 93, 94, 95, 97, 100 and `11-VALIDATION.md` at 67, 68, 71 —
  roughly **4-5× the dead-pattern class**.
  **Decision: ship R2b in Phase 14 as DETECTED AND PINNED, but enforce-to-zero
  only R1 + R2 + R3 this phase.** Under-scoped closure lands with QLT-10
  reconciliation in Phase 20. Re-adjudicating what a row's command was *meant* to
  prove is judgment work, not mechanical, and folding it in here would push the
  phase past its 6-8 plan sizing. Deciding this silently is how the phase blows
  its budget — so it is decided here, in writing.

- **D-14-17 (the `\|` ruling — verify first, then record):** `\|` inside a table
  cell is markdown's escape for a literal pipe. Copy-pasted **verbatim from the
  raw markdown**, `go test ./internal/compiler/corevalidate -run 'PeerLiveness\|LoanChainIndex' -count=1`
  prints `[no tests to run]` and **exits 0** — a **fourth dead command in the
  `nyquist_compliant: true` phase** — while `-list 'PeerLiveness|LoanChainIndex'`
  resolves to `TestPeerLivenessFileImportsStayIndependent`.
  **Ruling: unescape** — the rendered cell is the command, and it is copy-paste
  runnable from the rendered table. **Record this decision explicitly in the
  phase's evidence**, because the alternative reading produces **32 findings** out
  of 171 `go test -run` occurrences and the difference is otherwise invisible.

- **D-14-18 (landing order — land GREEN with an exact pinned frontier set):**
  not fix-first (the lint would never observe a real defect, contradicting
  `REQUIREMENTS.md` §Out of Scope: "the lint must do the finding"), and not
  red-on-main (its own hazard; trains people to ignore CI). Use the project's own
  Governing Gate idiom: the lint lands **green**, with
  `TestVerificationGroundednessFrontierIsPinned` asserting the current violation
  set equals an **exact committed literal** — **not a count, not a ceiling**
  (a count-keyed ceiling lets a count-neutral swap land). Each reconciliation
  shrinks the literal; the pin must be **empty** by the end of Phase 14 / QLT-10.
  **The pinned diagnostic must move — same shape as every other frontier fixture
  in this milestone.**

- **D-14-19 (re-measure the pin; do NOT assume "three"):** under R2 with union
  semantics, the Tier-A dead set reproduced today is the three named in the
  ROADMAP (`08-VALIDATION.md:51`, `08-VALIDATION.md:63`, `09-VALIDATION.md:85`)
  **plus** `12-VALIDATION.md:24`'s unfilled template row `<TestName>` over
  `./internal/compiler/<package>/...` (R1), **plus** three R1-elided patterns and
  two R1-elided paths in `11-VALIDATION.md`, **plus** whatever D-14-17's ruling
  admits. The plan measures the real set before pinning it.

- **D-14-20 (corpus floors, so the lint cannot go inert by finding nothing):**
  assert `len(tierAFiles) >= 28` and `len(verificationCommands) >= 140` (measured
  today: 28 VALIDATION/VERIFICATION files; 146 distinct pairs). A lint whose glob
  silently stops matching is the same failure class as a `-run` pattern that
  silently stops matching.

### Unreachable claims and witnesses (EVD-03, EVD-04, EVD-05, PRC-01)

- **D-14-21:** Adopt **witness-first**. Extend the existing `*-DEBT.md` law with
  `Grade` + `Witness` columns; **`.planning/UNREACHABLE-CLAIMS.md` is a GENERATED
  view** (per D-14-13), byte-compared in CI. Rejected: a new authored top-level
  register — that ships a second coexisting law without retiring the first, puts
  each claim in two places with nothing forcing agreement, and makes the grade
  vocabulary a third source of truth, all inside the phase whose purpose is
  retiring duplicated laws.
  — **Reversibility:** costly — 12 `*-DEBT.md` files gain two columns and two
  legacy exemptions extend.

- **D-14-22 (the decisive argument, and the one the plan MUST lead with):** a
  trigger written as a **grep over `check.go:255`** is the four-times-fired defect
  **in a fifth costume**. When `sameType` is refactored or renamed, the grep
  silently stops matching and the guard goes green **for the wrong reason** —
  precisely instance 4's escape, where the guard watched the production file while
  the stale marker lived in the test file and an error string. The industrial
  precedent that already solved this is **LLVM `lit`'s `XFAIL`**: the
  expected-to-fail test is *actually executed*, and an unexpected pass is reported
  as **XPASS and fails the suite** — which is verbatim EVD-04's "a guard fails when
  a cited gate has already closed". The same inverted-assertion property is what
  makes `nolintlint`'s `allow-unused: false` and PHPStan's
  `reportUnmatchedIgnoredErrors` work, and its absence is why Chromium/WebKit
  `TestExpectations` files rot into thousands of stale lines.

  **The framing for D-13-02b specifically, from DO-178C:** B1's blame branch is
  **deactivated code, not dead code** — present, provably unexecutable in the
  current configuration, justified by analysis that must be **re-examined when the
  configuration changes**. That is what an executed probe forces and what a
  phase-number citation cannot.

  **The rule in one line: a witness is an assertion that currently holds and is
  asserted to hold, so that its ceasing to hold is red.**

- **D-14-23 (closed witness grammar — four kinds, all executed):** a `Witness`
  cell is exactly one of these tokens (comma-separated only when several must
  *all* still hold). Free text FAILS — this is the direct fix to
  `"OPEN and UNOWNED — reopens only when …"` passing because it is non-empty.

  | Kind | Syntax | Fires when |
  |---|---|---|
  | `probe:` | `probe:TestB1BlameIsStructurallyUnreachable` | a test compiling a real `.lang` fixture that *would* exercise the claim, asserting it is refused by name — the refusal stops happening ⇒ red (XPASS) |
  | `callsite:` | `callsite:internal/compiler/check.resolveBlame=0` | a stdlib `go/ast` + `go/parser` scan of exact **call** sites — the count changes. **If the symbol does not exist, that is a FAIL, not a zero.** Absence is never a pass; this is the anti-grep rule |
  | `escape:` | `escape:callback-invocation-unsubjected` | the shipped escape-ID mechanism; the ID must appear in the closed escape registry and carry its own `probe:` |
  | `env:` | `env:clang` | environmental / host-capability, never closes, explicitly NOT debt — but must be declared, and `env:` is a **closed set** so a new one cannot be invented to launder a structural skip |

  **`phase:P17` is permitted only as an annotation inside the `Landing phase`
  cell, never as a `Witness`** — "Phase 17 lands" is decided by a human editing
  `ROADMAP.md`, which is an artifact read, which is the defect. The phase ID
  carries ownership (PRC-01); the probe carries the mechanics.

- **D-14-24 (close the `Landing phase` vocabulary — PRC-01's mechanization):**
  exactly three forms: `P<NN>` | `CLOSED(<commit-sha>)` | `UNOWNED(<witness-id>)`,
  enforced the way `debtRegisterSeverities` already is. Under the closed form,
  `UNOWNED` must name a witness ID that **resolves to an executed probe**, so
  non-empty is no longer sufficient. Assert the `UNOWNED` count at milestone close
  is <= the count at milestone open (PRC-02's <= 5 cap asserted mechanically).
  Also fix the naming smell: assert one convention (`PHASE-<NN>-DEBT.md`) for any
  register created from M003 on, keeping the existing two-form glob only for the
  frozen M001 files, listed the way `debtRegisterLandingPhaseExemptions` already
  lists them.

- **D-14-25 (EVD-04's guard scope — all suppression surfaces, including the two that instance 4 missed):** `TestNoSuppressionOutlivesItsWitness` parses **every**
  package with `go/parser` in `ParseComments` mode, **including `_test.go` files**,
  and enumerates:
  - every `*ast.CallExpr` whose selector is `Skip`/`Skipf`/`SkipNow` on a `testing.TB`;
  - every `//go:build` constraint outside a declared allowlist;
  - every `*ast.Comment` **and every `*ast.BasicLit` of kind `STRING`** matching
    `PENDING-\d\d-\d\d`, `\bD-\d\d-\d\d\b`, or `escape:[a-z-]+`;
  - every `NAT03Mutation` literal with `Subjected: false`.

  **Including `BasicLit` strings and parsing `_test.go` is the specific fix for
  instance 4**: it catches `internal/compiler/session/session_phase5_alias.go:523`'s
  error string `"(row not yet subjected — see PENDING-05-08)"` and the prose in
  `session_phase5_test.go:207` / `session_phase5_alias_test.go:231`. Each
  enumerated site must cite exactly one resolvable witness ID; a citation-free
  `t.Skip` is a fail; a cited ID whose witness no longer holds is a fail.

- **D-14-26 (retire the old law in the same phase — non-negotiable):** delete
  `TestNoNAT03RowRemainsPending` (`session_phase5_test.go:229-246`), the stale
  marker at `session_phase5_alias.go:523`, the `t.Skip` at
  `session_phase5_alias_test.go:241`, and the stale prose at `:185/:207/:230/:280`
  **in the same phase that lands the new guard**. Shipping the new law while the
  old one survives is the delayed fuse this project has already written down.

- **D-14-27 (EVD-05 — one axis-movement law):** delete
  `Phase5AssertMutationMovesAnAxis` (`session_phase5.go:412-428`, a thin wrapper
  delegating to `AssertMutationMovesAnAxis` at `session_phase5_alias.go:497-523`),
  remove the single `Subjected: true` fixture-path exclusion, and keep
  `retained_pointer` as a **declared escape**, not a silent skip: it stays
  `Subjected: false` with `EscapeID: "escape:callback-invocation-unsubjected"`,
  which must now resolve in the closed escape registry and carry its own probe.
  **EVD-05's "zero per-row exclusions" is thereby satisfied honestly** — the row is
  not excluded from the law, it is admitted *with a witnessed declared escape*.

- **D-14-28 (non-inertness for the suppression guard — three seeded faults, not one):** `TestSuppressionWitnessGuardIsNotInert`, following
  `TestInjectorMarkerCountGuardIsNotInert`'s temp-copy pattern, seeds one fault
  per mechanizable kind and asserts red for each: (a) inject `t.Skip("no reason")`
  into a copied package ⇒ uncited suppression flagged; (b) flip a `callsite:`
  count in a copied `*-DEBT.md` ⇒ mismatch; (c) neutralize a probe's refusal by
  editing a copied fixture so it admits ⇒ XPASS. **A guard proven red on only one
  of the three is inert for the other two.**

- **D-14-29 (source-of-truth closure — no third store):** the debt row is the
  **authored** truth, the probe is the **executed** truth, the file is
  **derived**. Two closure assertions: (i) every row with `Grade < EXERCISED` must
  carry a `Witness` or be `DEFINED`-and-withdrawn; (ii) every view row must appear
  in exactly one `*-DEBT.md`. **The interesting disagreement is the productive
  one:** if a row says `WIRED` but its probe fires (the claim became reachable),
  the suite goes red and a human must regrade. **Grading is authored; ungrading is
  forced by execution.**

- **D-14-30 (the day-one register contents, validated against real claims):**

  | Claim | Grade | Witness | Landing phase |
  |---|---|---|---|
  | **D-13-02b** (B1 blame unreachable) | `WIRED` | `probe:TestB1BlameIsStructurallyUnreachable, callsite:internal/compiler/check.resolveBlame=0` | `P17` |
  | **D-13-10a** (`use_matching_argument` no-op repair) | `DEFINED` (withdrawn) | `probe:TestB1BlameIsStructurallyUnreachable` — **the same probe; shared root cause, so Phase 17 turns both rows red in one commit** | `P17` |
  | **D-12-43** (unconstructible decisive control) | `WIRED` | `probe:TestD1243ControlIsUnconstructible` | `P17`, or `UNOWNED(...)` if the constructibility trigger differs |
  | **D-13-34** (M001 held-out pairs are alpha-renames) | `WIRED` | `probe:TestPhase6HeldoutPairsAreAlphaRenamesOnly` — asserts the weakness **still holds exactly as described**, so fixing the fixtures turns it red and forces closure; replaces the free-text `t.Skipf` at `session_phase6_injectors_test.go:186` | `UNOWNED(probe:TestPhase6HeldoutPairsAreAlphaRenamesOnly)` — honest, and counts against PRC-02's cap of 5 |
  | **`-flto` multi-function inertness** (EVD-07) | `WIRED` | `probe:` on the multi-function emission refusal; the `clang unavailable` skips at `native_lto_test.go:326,345` are `env:clang`, exempt from witness but must be declared | owning phase per PRC-01 |
  | **D-11-02** (`session_phase11_differential_test.go:365`) | `WIRED` | currently a free-text `t.Skip` — must gain a `D-11-02` citation + probe under D-14-25 | per PRC-01 |
  | **`retained_pointer`** | n/a — not an unreachable claim | `escape:callback-invocation-unsubjected` + its own probe (D-14-27) | n/a |

- **D-14-31 (mechanize "don't defer the same debt twice"):** add
  `first-recorded: <milestone>` to each row's detail section and assert that a row
  carried across two milestone boundaries requires a `gate="blocking-human"`
  ratification record. This mechanizes the project's own written rule — "D-03-02
  was carried past M001; carrying it past M002 would be a pattern, not a decision"
  — and the cluster-A base rate of **0 for 2**.

- **D-14-32 (accepted residual friction, stated in the plan rather than discovered):** one written probe per claim is real cost, and a probe can
  **over-fit to current diagnostic text**, producing a red on a benign message
  rewording. Mitigation: assert on the **structural** refusal (admission rejected
  at the `sameType` precondition) rather than on the message string. Note also
  that a probe executing every commit is what makes register growth *visible*
  rather than free — the inverse of Chromium expectations files, where a stale
  line costs nothing. That runtime cost is the anti-graveyard feature, not a bug.

### Diagnostic distinctness and self-explaining declines (DX-08, DX-09)

- **D-14-33:** Adopt the **`skipped_region` cause**. Attach the byte extent
  discarded by declaration-level error recovery as an identity-bearing
  `diagnostic.Cause` on `syntax.expected_declaration`, and ship the fix **and**
  the `diagnostic_distinctness` gate **in one plan**. Rejected: pinning the spiral
  as a failing control with a recorded trigger and deferring the fix — argued
  fairly below, and it would be the right answer if the fix had cost a schema bump
  or churned the `check_ordering_stability_test.go` pins. **It does neither.**
  — **Reversibility:** costly — it changes published diagnostic IDs for
  `syntax.expected_declaration`; in-repo churn is measured at zero, but external
  consumers pinning IDs get no version signal.

- **D-14-34 (a CORRECTION to the research's framing that the plan must carry — do not treat span relocation as the fix):** "the first diagnostic is one token
  late" is **not a defect**. `if` is not in `keywords` (`lexer.go:10-22`), so it
  lexes as `TokenIdentifier`; `funcDecl()` (`parser.go:299`) sees no `match`,
  calls `linearBody()`, and `if` is consumed as the terminal result identifier via
  `p.identifier("syntax.expected_linear_result")` (`parser.go:472`). "After the
  result, expected `}`, found `value`" is therefore **correct**. Re-pointing that
  span at `if` would require the parser to know `if` is special — which **is** the
  deferred `surface.not_in_language`. And it **does not separate the three
  programs**: all three share `if` at the same offset.

  **The real mechanism:** `parseProgram`'s `default` arm (`parser.go:169-171`)
  reports one token, then `p.recoverUntil(...)` skips everything remaining to EOF
  **and reports nothing about it**. The one datum that actually differs between
  the three programs is discarded before identity is computed. The fix is Roslyn's
  `SkippedTokensTrivia` — the parser skips tokens until it can continue and the
  skipped extent is attached as a node with a real extent — brought into the
  **diagnostic** instead of the AST.

- **D-14-35 (the exact change — ~14 lines, one call site):**
  - `internal/compiler/syntax/parser.go`, replace the `default:` arm at
    `:169-171` to capture `unexpected := p.peek()`, call a new
    `recoverRegion(...)` returning the discarded `diagnostic.Span`, and pass it as
    `diagnostic.Cause{Kind: "skipped_region", Span: &skipped}`.
  - Add `recoverRegion` next to `recoverUntil` (`parser.go:~700`) — `recoverUntil`
    that reports the byte extent of what it discarded.
  - Widen `p.problem` to variadic causes. **This is the only signature change and
    all 40+ existing call sites compile unchanged.**
  - `Cause.Kind` is bare snake_case in this codebase (`"missing_alternative"`,
    `"cycle_member"`, `"callee"`), so `"skipped_region"` fits the existing
    vocabulary. **No `diagnostic.go` change at all.**

- **D-14-36 (MEASURED, not asserted — the fix separates all three):**

  | | skipped region | `causes[0].span` | diagnostic ID | result ID |
  |---|---|---|---|---|
  | `if v { v } else { v }` | 26 bytes | `{75,101}` | `c3c949eac26c45035f0dcbd5` | `result:39d3025a54df81141bdd160d` |
  | `if v { v }` | 11 bytes | `{75,86}` | `380d0246231d20d53ed3b306` | `result:a9f985a316e0ad7d15482127` |
  | `if v` | 1 byte | `{75,76}` | `98d64518731b3694192ba613` | `result:b0781bb3e8ba07e5c891c739` |

  Three distinct diagnostic IDs, three distinct diagnostic-ID sets, three distinct
  result IDs. **`primary_span` stays `{75,76}` — one token wide — in all three**,
  honoring the rendering guidance that a narrow diagnostic under the error point
  beats one spanning a long width. The discriminator lives in `causes`, where
  nothing renders it. The crude variant (widening `primary_span` over the skipped
  region) also separates all three and also passes the suite, but regresses the
  rendering story and leaves `c`'s ID unchanged only by accident — **take the
  cause variant.**

- **D-14-37 (no schema bump — a principled ruling, not a convenience):** the
  identity basis is `{Schema, Code, Span, Causes}`. **`Causes` is ALREADY in the
  basis.** This change alters identity *content*, not the *basis*. The `/0`→`/1`
  bump happened "ONLY because the identity basis changed" (D-13-09a); the
  D-04-23 / D-13-18 additive-`omitempty` precedent is adjacent but need not even be
  invoked, because no struct gains a field. **`lang.diagnostic/0` stands.**

- **D-14-38 (churn budget — measured zero in-repo, enumerated up front anyway):**
  full `go test ./...` passes **unmodified** with the patch applied.
  `syntax.expected_declaration` appears in **no pinned artifact** — not in
  `internal/compiler/check/check_ordering_stability_test.go`, not in any
  `cmd/lang-repair/testdata/*_capture.json`; its only in-repo mention is a
  set-membership check at `check_blame_test.go:750`, which tests the code string,
  not an ID. Zero is a **measurement, not a guarantee**, so D-13-09a's discipline
  still applies as an **explicit, called-out plan task**:

  > **Task — enumerate published-ID churn.** Run `lang --json check` over every
  > `testdata/**/*.lang` before and after the parser change; diff the `id` fields
  > of all diagnostics and all results. Assert the diff set is **exactly** the new
  > `testdata/distinctness/` members plus any fixture whose declaration recovery
  > skipped >= 1 token. Publish the table in the phase's evidence. **Any fixture in
  > the diff that is not in that predicted set FAILS the task.**

  **The budget, stated precisely: broken-program diagnostic IDs may move; no valid
  program's evidence moves.** Churn is confined to programs that reach
  `parseProgram`'s `default` arm at all — i.e. are already syntactically broken —
  and had >= 1 token discarded.

- **D-14-39 (gate vs fix — both, one plan, ambiguity resolved by splitting the role):** ROADMAP criterion 5 requires the trio to mint distinct IDs, so a fix is
  mandatory and a pinned failure does not satisfy it. But "negative control" does
  mean the thing that must FAIL. Both are satisfied at once:
  - **Positive regression pin (must PASS):** the live spiral trio in
    `testdata/distinctness/`, asserted to yield 3 unique diagnostic-ID sets and 3
    unique result IDs. This is criterion 5.
  - **Negative / non-inertness control (must FAIL):** a **frozen pre-fix
    capture**, `testdata/distinctness/collision_control.json`, recording today's
    byte-identical three-way output. `TestDiagnosticDistinctnessGuardIsNotInert`
    runs the metric over it and asserts **1/3, not 1.0**.

  This is **not** a `--bless`-shaped regenerate button: the capture is a
  historical artifact that must **never** be regenerated, and the test asserts a
  *property* (collision detected) over it, not a golden match. The gate cannot be
  landed first and left red — project culture forbids a known-red gate in the
  tree, and the fix is 14 lines.

- **D-14-40 (corpus distinctness predicate for PARSE-FAILING programs):**
  D-13-33's structural summary and D-13-26's topology triple both **degenerate**
  here — these programs never reach the checker, so there is no call graph.
  **Predicate: two corpus members are structurally distinct iff their non-trivia
  TOKEN-KIND SEQUENCES differ**, computed by `internal/compiler/syntax.Lex` alone.
  Why this one:
  - **Total and cheap** — defined on every byte string, including ones that never
    parse; no AST required.
  - **Renaming-immune by construction** — every identifier lexes to
    `TokenIdentifier`, so a corpus cannot be padded with `foo`/`bar` renames. The
    most obvious gaming vector is closed by the predicate's own shape, not by a
    side rule.
  - **Independent of the instrument under test** — it never touches
    `parseProgram` or `check`. A corpus certified by the instrument under test
    measures its own assumptions.
  - Right granularity for the trio:
    `[ident ident { ident } ident { ident }]` / `[ident ident { ident }]` /
    `[ident ident]` — pairwise different, so the trio is admissible and the metric
    is not begging the question.

  Corpus: `testdata/distinctness/`, ~10-12 parse-failing fixtures — the spiral
  trio plus `for i in 0..10 {...}`, an arithmetic expression, an unclosed brace, a
  stray `;`, an `else` with no `if`, an empty file, and **a program whose skipped
  region is one token** (the hard case that nearly collides with `c`).

- **D-14-41 (the gate — `internal/compiler/check/diagnostic_distinctness_test.go`, ~60 lines, three assertions):**
  1. `TestDistinctnessCorpusMembersAreStructurallyDistinct` — pairwise
     kind-sequence inequality. **Fails the corpus, not the compiler**, if someone
     adds a near-duplicate.
  2. `TestDiagnosticDistinctnessIsOne` — unique diagnostic-ID sets == corpus size
     **and** unique result IDs == corpus size.
  3. `TestDiagnosticDistinctnessGuardIsNotInert` — the frozen collision control.

  **Anti-Goodhart:** a distinctness of 1.0 is corpus-gameable, and the denominator
  is chosen by the people who want it green. Three things constrain it and all
  three go in the plan: (i) the kind-sequence predicate makes trivial padding
  impossible; (ii) **corpus membership is ADDITIVE-ONLY** — removing a member to
  make the gate green is a reviewable deletion in the diff, and requires a
  recorded decision, the same shape as re-pinning an ID; (iii) the frozen
  collision control fails immediately if the metric is stubbed. **What does NOT
  stop it:** nothing forces the corpus to grow as the language grows — recorded as
  debt with an explicit trigger (D-14-43).

- **D-14-42 (DX-09's exact shape — additive, closed, prose-free):** the shipped
  field is `DiagnosisCode string \`json:"diagnosis_code,omitempty"\`` on `Outcome`
  (`cmd/lang-repair/repair.go:114`). It is empty on the decline path for exactly
  one reason: `selectRepair` (`:200-209`) returns `""` when the loop finds no
  driver-eligible repair. **The information was in hand and discarded.**
  **Do not change `diagnosis_code`'s type** — changing a shipped field's type is a
  protocol break for no gain. Add three `omitempty` siblings on the
  non-identity-bearing `Outcome` envelope (which carries no schema string at all,
  so D-04-23 / D-13-18 apply directly):
  - `DiagnosisCodes []string` — **all** diagnostic codes in document order; the
    honest answer for the multi-diagnostic parse-failure path.
  - `DeclineReason string` — a **closed coded vocabulary**, new constants beside
    the existing `repair.*` block (`repair.go:95-98`): `repair.none_offered`
    (invalid, >= 1 diagnostic, zero `repairs[]` anywhere — **this is the spiral's
    value**), `repair.none_eligible` (repairs present, none driver-eligible),
    `repair.no_diagnostics` (invalid with zero diagnostics — a fail-closed
    anomaly that must still read as `unrepairable`, never a pass).
  - `BestApplicability string` — max applicability observed across all repairs,
    from the closed vocabulary already re-declared locally; empty when
    `none_offered`.

  `diagnosis_code` on the decline path is **the first diagnostic's code in
  document order** — deterministic and already specified, rather than an invented
  heuristic. Worked example on the spiral:
  `{"status":"unrepairable","diagnosis_code":"syntax.expected_rbrace","diagnosis_codes":["syntax.expected_rbrace","syntax.expected_declaration"],"decline_reason":"repair.none_offered","subprocess_count":1}`.
  `selectRepair` returns `(jsonRepair, decision, bool)`; ~25 lines, all inside
  `cmd/lang-repair/repair.go`.

  **Three boundary checks, all clear and all of which the plan must state:**
  - **Import boundary** (`import_boundary_test.go`, D-06-28) untouched — every new
    value is read off the already-decoded `jsonDiagnostic.Code` or is a new local
    `const`. `driverEligible` stays re-declared.
  - **Prose-scramble anti-theater** (`antitheater_test.go:635-641`) clear, and this
    is the load-bearing one: `jsonDiagnostic` deliberately omits every prose field
    so "a field that is never decoded cannot be scraped"
    (`TestRepairDriverDecodesNoProseFields`). **No new prose field is decoded**;
    every emitted value is a code from a closed vocabulary, so lorem-ipsum
    scrambling changes nothing in the output.
  - **`TestUnrepairableDefectFailsTheGate`** untouched; `NormalizeApplicability`
    still never defaults toward driver-eligible.

  **Honest declines stay honest and stay `unrepairable`**, now carrying their
  reason: D-13-09b forward-direction loan liveness and D-13-12's call-graph cycle
  both become `none_offered` — so "a cycle has no local, mechanical edit" is
  **stated in the protocol** rather than left as silence — and D-13-10's
  zero-or->=2-match uniqueness gate becomes `none_offered` or `none_eligible`
  depending on whether it suppresses emission.

  New guard + non-inertness proof: `TestUnrepairableAlwaysCarriesDiagnosis` over
  every `cmd/lang-repair/testdata/*_capture.json` and every fixture class, plus
  `TestUnrepairableDiagnosisGuardIsNotInert` using the existing stand-in-binary
  pattern to emit an `unrepairable` with an empty `diagnosis_code` and assert the
  guard FAILS.

- **D-14-43 (the case for deferral, argued fairly, and why it still loses — plus the debt it creates):** `if` is not in the language and will not be for
  milestones; the phase's other four criteria are markdown instruments; this work
  is genuinely foreign to its host phase; and the research's own "what would
  change my mind" concedes the spiral is a hand-simulated proxy, not a measured
  agent trajectory. A pinned failing control plus a recorded trigger would be
  honest, cheap and defensible **if the fix had cost anything**. It costs 14 lines
  in one call site with a measured zero-churn full-suite pass.

  **The decisive counter is a security/evidence lens, and it should lead the
  plan:** today three distinct sources map to **one `result:` ID**, so an evidence
  cache keyed on the result ID will serve `a.lang`'s evidence for `c.lang`. That is
  **evidence confusion in a product whose product is evidence** — the same defect
  class as Bazel/Nix action-key under-specification. Not remotely exploitable in a
  112-program language, but it generalizes: any future coarse-identity site is the
  same hazard, which is why D-14-38's enumeration task matters beyond this one fix.

  **Two debts this creates, recorded rather than rediscovered:** (a) putting more
  content on `Cause.Span` **deepens the recorded Phase 13 debt** that "a coordinate
  shift must never move a diagnostic's ID" is only half-enforced — a whitespace
  edit inside a broken program's tail now moves an ID, **for broken programs only**,
  an acceptable bounded widening of an already-recorded hole; (b) **when `if`
  enters the language, the spiral trio becomes a VALID-program fixture** and must
  be replaced in the distinctness corpus by the then-current not-in-language
  surface. Both get `*-DEBT.md` rows with witnesses per D-14-23.

- **D-14-44 (sequencing vs spike S-009 — independent, do not block):** S-009 asks
  whether the **prose** or the **protocol** carried a frontier fixture. Both DX
  items are API-contract defects measurable today, and both strictly **increase**
  the protocol-only signal S-009 measures — a `skipped_region` span and a
  `decline_reason` code are exactly the prose-free content that survives
  lorem-ipsum scrambling. If S-009 reports recovery already > 80%, that finding
  should suppress appetite for a **third**, wider DX item this milestone; it
  should not touch these two.

### Claude's Discretion

- Exact column ordering within the `*-VALIDATION.md` and `*-DEBT.md` tables, and
  whether `Non-inertness` and `Witness` are one column or two per file type.
- Whether the run record (`go test -json`) is produced by a new
  `scripts/evidence-run-record.sh` or by an existing CI step, provided the gate
  **fails closed** when a row declares `EXERCISED`+ and no run record is present
  (that fail-closed behavior itself needs `TestValidationGradeCapIsNotInert`,
  seeding a row declaring `MUTATION-KILLED` with a nonexistent twin — the direct
  analogue of `assert-go-tests.sh --self-test`'s sentinel).
- Whether the two generated views live at `.planning/` root or beside the
  registers, provided the paths `.planning/UNREACHABLE-CLAIMS.md` and
  `.planning/EVIDENCE-RECONCILIATION.md` are honored as EVD-03 names them.
- Whether the groundedness lint's error messages compute nearest-live-test-names
  by edit distance (recommended as the DX budget: a contributor fixing a violation
  should not have to re-derive the rename by hand).
- Whether EVD-06's `LANGUAGE-MATURITY.md` self-check executes the file's own
  embedded re-verify greps or re-derives the counts independently — peer-re-derivation
  discipline leans toward the latter. (It says 32 guards; there are 26.)
- Whether EVD-08's `observed` wall-clock row lands as a new metric in
  `internal/compiler/session/qlt02_budget_manifest.json` or a sibling manifest,
  provided the 192.7s baseline is recorded and the ~60s claim is corrected
  wherever it appears.
- Plan-count and wave shaping within the 6-8 plan sizing, and whether the two DX
  items are one plan or two.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project-level durable context (read FIRST — these override roadmap vocabulary)

- `.planning/LANGUAGE-MATURITY.md` — the language is far less expressive than the
  roadmap implies: no arithmetic, no iteration, no `if`, no strings/arrays,
  `Byte`/`Buffer` only, arity-1 calls, return type == parameter type. Assurance
  stack ~70-75% built; **language surface ~5-10%**. Do **not** read
  `wiki/example-tour.md` as a description of the language. **Note: this file's own
  counts are wrong and EVD-06 exists to fix that — it claims 32 guards, there are
  26, and its corpus figures are stale in both directions.**
- `.planning/STANDING-VERDICTS.md` — dependency verdicts (**zero external
  production dependencies is a HARD constraint**), design anti-features, process
  anti-patterns ("two coexisting laws is a defect with a delayed fuse"; "don't
  defer the same debt item twice"), the six dispatch sites, why `-flto` is
  load-bearing.
- `.planning/research/M003/ADVERSARIAL-SYNTHESIS.md` — **AUTHORITATIVE.** Its
  contradiction resolutions override the six source documents wherever they
  conflict. **§C5 (lines 170-204)** is binding on DX-08/DX-09: it defers the
  `not_in_language` manifest and `surface.not_in_language`, routes only
  `diagnostic_distinctness` and non-empty `diagnosis` into this phase as ~1 plan
  each, and states the red line — **"Do not weaken determinism; enrich the
  content."**
- `.planning/ROADMAP.md` §"Phase 14: Evidence Instrument and Honest Scoping"
  (lines 150-215) — goal, five success criteria, the **naming note** (REQUIREMENTS'
  DX-08/DX-09 are the synthesis's DX-10/DX-12 renumbered — **read the requirement
  text, not the number**), riskiest assumption, and sizing note.
- `.planning/ROADMAP.md` §"The Governing Gate — Constructibility Precondition"
  (lines 43-71) — standing discipline for every phase 14-20.
- `.planning/REQUIREMENTS.md` §"Evidence Instruments", §"Agent Loop", §"Process"
  — EVD-01..08, DX-08, DX-09, PRC-01, PRC-02 verbatim; §"Out of Scope" ("Nyquist
  reconciliation before EVD-01's lint exists — the lint must do the finding").
- `.planning/research/M003/EVIDENCE-AND-DEBT.md` — the measured ground truth
  behind every EVD requirement: §"Is the Nyquist closure actually cheap?"
  (146 pairs, 3 dead patterns, the per-phase reconciliation cost table, the
  Phase 11 archival breakage), §"The 'graded from wiring' process defect"
  (the four instances and F1-F5), §"Debt posture" (the five clusters),
  §"Weaker-than-believed claims" (W1-W4).
- `.planning/RETROSPECTIVE.md` — M001 Key Lesson 6 ("archival is a code change",
  now fired 3×), M001/M002 Lesson 5 (Nyquist compliant exactly where it was a gate
  condition), M002 Lesson 3 ("flagged for human review is not a work item").

### The existing instrument corpus (the patterns every deliverable extends)

- `internal/compiler/session/session_test.go:2505-2700` — **the canonical
  precedent.** `phaseArtifactGlob` (globs `.planning/phases/` AND
  `.planning/milestones/*-phases/`), `testsupport.ProjectPath`,
  `debtRegisterSeverities` (a shipped closed vocabulary),
  `TestDebtRegistersAreWellFormed`, `checkDebtRegister`, and
  `debtRegisterLandingPhaseExemptions` (`:2581-2590`) — the exact shape D-14-06's
  exemption map copies. **Read its doc comment**: the honesty/shape line at
  `:2596-2601` is load-bearing for D-14-02.
- `scripts/assert-go-tests.sh` — the shipped `-list` resolution algorithm
  (`:8-45`) D-14-03 reuses, and the `--self-test` sentinel
  (`TestCodenameLangSelectionGuardMustNotExist`) that D-14-08's non-inertness proof
  mirrors. **Keep it; do not fold the lint into it.**
- `scripts/verify-phase{1..7}.sh` — per-phase verification scripts exist for 1-7
  only; relevant to Tier-A scope reasoning.
- `internal/compiler/session/session_phase5.go:412-428` and
  `session_phase5_alias.go:497-523` — **the two coexisting axis-movement
  implementations EVD-05 collapses to one.**
- `internal/compiler/session/session_phase5_test.go:207,229-246` and
  `session_phase5_alias_test.go:231,241` — `TestNoNAT03RowRemainsPending` and the
  **still-live** `PENDING-05-08` markers in a test file and an error string
  (`session_phase5_alias.go:523`). D-14-26 deletes all of them.
- `internal/compiler/session/qlt02_budget_manifest.json` +
  `session_phase6_budget.go` + `session_phase6_budget_test.go` — where EVD-08's
  `observed` wall-clock row lands; `TestBudgetManifestLoads` and
  `TestQLT02BudgetManifestFileUnchangedDuringAudit` are the existing exercises.
- `internal/compiler/session/session_phase6_injectors.go` /
  `session_phase6_injectors_test.go:186` — `Injector` / `markerGuard` /
  `matchInjectSkippingGuard` and the free-text `t.Skipf` D-14-30 replaces with a
  probe.
- Existing non-inertness controls to mirror:
  `TestRepairDriverSourceHeldoutScanIsNotInert`,
  `TestInjectorMarkerCountGuardIsNotInert`, `TestCorpusTopologyGuardIsNotInert`.

### The documents under test (EVD-01's Tier A, and what it will find)

- `.planning/milestones/M002-phases/08-interprocedural-loan-liveness-in-check/08-VALIDATION.md`
  — dead patterns at `:51` (`TestComputeLoanLastUsesAndDerivePlaceLoansAgree`,
  `obsolete-by-design`: `computeLoanLastUses` was deleted in Phase 09) and `:63`
  (`TestAuditQLT02BudgetManifest`); Task ID column is literally `TBD` on all 13 rows.
- `.planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md`
  — `:85` dead (`LoanChainIndex`) **in a `status: validated` /
  `nyquist_compliant: true` file**; `:84` under-scoped; and the `\|` case of
  D-14-17. Further R2b findings at `:87,88,93,94,95,97,100`.
- `.planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md`
  — 27 rows all `⬜ pending`; >= 6 elided command cells including two **path**
  elisions; row `11-02 T3`'s `grep .planning/ROADMAP.md` returns 0 since archival.
  R2b findings at `:67,68,71`.
- `.planning/milestones/M002-phases/12-result-payloads/12-VALIDATION.md` — `:24`'s
  unfilled template row `<TestName>`; all 15 rows already green (frontmatter flip).
- `.planning/milestones/M002-phases/13-.../13-VALIDATION.md` — all 25 rows green,
  already `nyquist_compliant: true`; subtest-scoped evidence cells relevant to
  D-14-04; `blameFieldWitness`'s compile-time evidence relevant to D-14-03.
- `.planning/milestones/M001-phases/04-.../04-RESEARCH.md:560-570` — ten proposed
  test names never written; the concrete case for D-14-11's Tier B exemption.
- `.planning/PROJECT.md` — the DX-06 claim and NAT-07 `-flto` bullet EVD-07
  corrects.

### Diagnostics, identity, and the repair driver (DX-08, DX-09)

- `internal/compiler/diagnostic/diagnostic.go` — `Span` (byte offsets only),
  `Cause` (`Kind` is bare snake_case), the sha256 identity construction in
  `Error` / `ErrorWithRepairs`, the identity basis `{Schema, Code, Span, Causes}`
  (+`RepairKinds` under `/1`), and the comment explaining why `Repair.Span` is
  deliberately non-identity-bearing. **D-14-35 changes nothing in this file.**
- `internal/compiler/syntax/parser.go` — `parseProgram`'s `default:` arm
  (`:169-171`), `recoverUntil` (~`:700`), `funcDecl` (`:299`), the
  `expect(TokenRBrace, ...)` at `:319`, `p.identifier("syntax.expected_linear_result")`
  (`:472`), and `p.problem`. **The entire DX-08 fix lives here.**
- `internal/compiler/syntax/lexer.go:10-22` — `keywords`; `if` is absent, which is
  the whole mechanism (D-14-34).
- `internal/compiler/check/check_ordering_stability_test.go` — the pinned
  diagnostic IDs. **`syntax.expected_declaration` does not appear here** (D-14-38).
- `internal/compiler/check/check_blame_test.go:750` — the only in-repo mention of
  `syntax.expected_declaration`; a code-string set-membership check, not an ID.
- `cmd/lang-repair/repair.go` — `Outcome` (`:114`), the `repair.*` const block
  (`:95-98`), `selectRepair` (`:200-209`, the empty-`diagnosis` bug),
  `Repair()` (`:262-290`, the D-06-30 single-pass structure).
- `cmd/lang-repair/import_boundary_test.go` — the structural lint forbidding
  `internal/` imports (D-06-28). **Any DX-09 change must respect it.**
- `cmd/lang-repair/antitheater_test.go:635-641` — the prose-scramble test, and
  `TestRepairDriverDecodesNoProseFields` (T-06-BOUNDARY-03). **Whatever goes in
  `diagnosis` must survive lorem-ipsum scrambling — codes, never English.**
- `cmd/lang-repair/repair_test.go` — `TestUnrepairableDefectFailsTheGate`.
- `cmd/lang-repair/testdata/*_capture.json` — the capture corpus; unaffected by
  D-14-35 per D-14-38's measurement, to be re-confirmed by the enumeration task.

### Prior-phase decisions this phase depends on or closes

- **D-13-02b** — B1 blame is structurally unreachable; `sameType` enforced at
  `check.go:255`, `:3148`, `:3399`. The register's day-one flagship row.
- **D-13-10a** — `use_matching_argument`'s no-op repair; same root cause, same
  Phase 17 trigger, **same probe**.
- **D-12-43** — an unconstructible decisive control, ratified as a terminal
  finding at a blocking-human checkpoint rather than absorbed. **The precedent for
  how this project holds a negative result.**
- **D-13-34** — the M001 held-out corpus is alpha-renames; disposition is
  **record the trigger, do not fix now** (a distinct pair drawn from a 112-program
  space is a cosmetic pass).
- **D-13-09a** — attaching a repair forced `/0`→`/1` and churned six pinned rows;
  **the precedent for D-14-38's up-front enumeration task.**
- **D-13-16 / D-13-18 / D-04-23** — why `Span` must not grow fields, and why
  additive `omitempty` on a non-identity-bearing struct does not bump a schema.
- **D-13-33** — retro-strengthening M001's corpus; "if it turns green tests red,
  that is a real hole being surfaced, not a regression to suppress" — **the
  precedent for D-14-07.**
- **D-10-60** — every debt item names an owning phase; prose did not mechanize it.
  PRC-01 and D-14-24 do.
- **D-06-24 / D-06-28 / D-06-30** — `Repair.Span` non-identity-bearing; the driver
  is a protocol consumer, not a library consumer; single-pass repair.

### External precedent cited in the decisions above

- **LLVM `lit` `XFAIL`/`XPASS`** — an expected-to-fail test is *executed*, and an
  unexpected pass **fails the suite**. The formal model for D-14-22/D-14-23.
  <https://llvm.org/docs/CommandGuide/lit.html>
- **DO-178C's dead-code vs deactivated-code distinction** — the frame for
  D-13-02b: present, provably unexecutable in the current configuration, justified
  by analysis that must be re-examined when the configuration changes.
- **`nolintlint` `allow-unused: false`; PHPStan `reportUnmatchedIgnoredErrors`** —
  the same inverted-assertion property.
  <https://github.com/ashanbrown/nolintlint>, <https://phpstan.org/user-guide/baseline>
- **Chromium / WebKit `TestExpectations`** — the counter-example: stale lines cost
  nothing, so the file rots into thousands of them. D-14-32's runtime cost is the
  deliberate inverse.
- **ESLint bulk suppressions / count-keyed ratchets** — why D-14-12 and D-14-18
  refuse a baseline and a count ceiling. <https://eslint.org/docs/latest/use/suppressions>
- **Roslyn `SkippedTokensTrivia`** and Roslyn issue #30749 on skipped-token
  recovery diagnostics — the model for D-14-35.
  <https://learn.microsoft.com/en-us/dotnet/csharp/roslyn-sdk/work-with-syntax>,
  <https://github.com/dotnet/roslyn/issues/30749>
- **SARIF 2.1.0 `fingerprints` / `partialFingerprints`** — identity enriched with a
  second cheap deterministic discriminator computed from the same content.
  <https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html>
- **Bazel / Nix action-key under-specification** — the defect class D-14-43's
  security argument invokes.
- **`sphinx-needs` / RTM practice** — requirements-with-status-in-docs, and the
  documented degeneration of split stores into paperwork; the basis for D-14-05's
  no-sidecar ruling. <https://safety.useblocks.com/tools/sphinx-needs/index.html>
- **PIT / Stryker "killed" semantics** — a seeded fault causing a red; what
  D-14-03's `MUTATION-KILLED` boundary formalizes.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- **`phaseArtifactGlob` + `testsupport.ProjectPath`
  (`internal/compiler/session/session_test.go:2505-2520`)** — already globs both
  live `.planning/phases/` and archived `.planning/milestones/*-phases/`. **Every
  new instrument in this phase uses it; none needs new path plumbing.**
- **`checkDebtRegister` (`:2628`)** — a working frontmatter + markdown-table
  parser with a declared-count-vs-table-length cross-check. D-14-21's `Grade` /
  `Witness` / closed-`Landing phase` additions extend this function rather than
  writing a sibling.
- **`debtRegisterSeverities` (`:2594`)** — a shipped closed vocabulary enforced by
  a Go test. The literal model for the grade, witness-kind, verdict, and
  `Landing phase` vocabularies.
- **`debtRegisterLandingPhaseExemptions` (`:2581-2590`)** — the exact shape and
  review discipline D-14-06's file-scoped, dated `validationGradeBarExemptions`
  copies.
- **`scripts/assert-go-tests.sh:8-45`** — the exact-name `-list` resolution
  algorithm D-14-03 reuses, plus a `--self-test` sentinel that is already a shipped
  non-inertness proof.
- **`AssertMutationMovesAnAxis` (`session_phase5_alias.go:497-523`)** — the
  surviving half of EVD-05's collapse; `Phase5AssertMutationMovesAnAxis`
  (`session_phase5.go:412-428`) is the thin wrapper that dies.
- **`Injector` / `markerGuard` / `matchInjectSkippingGuard`** — the proven
  fail-closed mutation harness with an existing not-inert proof; the temp-copy
  pattern D-14-28's three seeded faults follow.
- **`diagnostic.Cause` with its `Span` already inside the hashed identity struct**
  — the mechanism that makes D-14-35 work with **no** `diagnostic.go` change and
  **no** schema bump.
- **`Outcome`'s `omitempty` envelope with no schema string
  (`cmd/lang-repair/repair.go:114`)** — the additive surface DX-09 extends under
  the D-04-23 / D-13-18 precedent.

### Established Patterns

- **Closed vocabularies enforced by a Go test** — `debtRegisterSeverities` is the
  template for five new vocabularies in this phase (grade, witness kind,
  reconciliation verdict, `Landing phase`, `decline_reason`).
- **Every control ships a non-inertness proof** — this phase adds at minimum
  `TestVerificationGroundednessIsNotInert` (four seeded faults),
  `TestValidationGradeCapIsNotInert`, `TestSuppressionWitnessGuardIsNotInert`
  (three seeded faults), `TestDiagnosticDistinctnessGuardIsNotInert`,
  `TestUnrepairableDiagnosisGuardIsNotInert`.
- **Fail-closed defaults** — `NormalizeApplicability` never defaults toward
  driver-eligible; `markerGuard` refuses rather than returning unmutated source;
  `unrepairable` is never conflated with a pass. D-14-23's "a missing symbol is a
  FAIL, not a zero" and D-14-15's "no silent skips" inherit this posture directly.
- **Anti-theater** — no `--bless`-shaped regenerate button; property assertions
  over goldens. This is why D-14-39's collision control is a **frozen historical
  artifact** asserted against a property, and why D-14-13's generated views are
  **regenerate-and-compare** rather than checked-in truth.
- **Peer re-derivation / nothing derived once and trusted** — D-14-10's static
  index is cross-checked against real `go test -list`; D-14-40's corpus predicate
  uses the lexer, never the parser under test.
- **Structural import boundary** — `cmd/lang-repair` may not import `internal/`;
  D-14-42 respects it by reading only already-decoded JSON codes and new local
  constants.
- **"Two coexisting laws is a defect with a delayed fuse"** — the governing
  constraint behind D-14-09 (keep the script and the lint non-overlapping and say
  so in both files), D-14-13 (two authored laws, not three), D-14-21 (extend, do
  not mint), D-14-26 (delete the old guard in the same phase), and D-14-27 (one
  axis-movement law).

### Integration Points

- `internal/compiler/session/verification_groundedness_test.go` — **NEW.** EVD-01:
  the lint, the frontier pin, the not-inert proof, the static-index accuracy
  control, the reconciliation-view well-formedness check, and the corpus floors.
- `internal/compiler/session/evidence_grade_test.go` — **NEW.** EVD-02: the
  derivation ladder, the cap, the exemption map, the not-inert proof.
- `internal/compiler/session/session_test.go` — EXTENDED. `checkDebtRegister`
  grows `Grade` / `Witness` / closed `Landing phase`; the suppression enumerator
  and the generated-view byte-compares live alongside it.
- `internal/compiler/session/session_phase5.go` + `session_phase5_alias.go` +
  their tests — EVD-05's collapse and D-14-26's deletions.
- `internal/compiler/session/qlt02_budget_manifest.json` — EVD-08's `observed`
  wall-clock row (192.7s baseline; the ~60s claim is wrong by 3.2×).
- `internal/compiler/syntax/parser.go` — **the only production compiler file this
  phase touches.** DX-08's `default:` arm, `recoverRegion`, variadic `problem`.
- `cmd/lang-repair/repair.go` — DX-09's three `omitempty` fields, three decline
  constants, and `selectRepair`'s decision return.
- `internal/compiler/check/diagnostic_distinctness_test.go` — **NEW.** The gate,
  the corpus predicate, the frozen collision control.
- `cmd/lang-repair/repair_diagnosis_test.go` — **NEW.** DX-09's guard and its
  not-inert proof.
- `testdata/distinctness/` — **NEW** corpus + `collision_control.json` (frozen,
  **never regenerated**).
- `.planning/**/*-VALIDATION.md` (13 archived + new) — `Grade` / `Non-inertness`
  columns, `Status` retired, frontmatter additions.
- `.planning/**/*-DEBT.md` (12 files) — `Grade` / `Witness` columns, closed
  `Landing phase`.
- `.planning/UNREACHABLE-CLAIMS.md`, `.planning/EVIDENCE-RECONCILIATION.md` —
  **NEW, GENERATED**, byte-compared.
- `.planning/PROJECT.md`, `.planning/LANGUAGE-MATURITY.md` — EVD-06 / EVD-07
  corrections.

</code_context>

<specifics>
## Specific Ideas

- The user asked for maximum-breadth adversarial research across all relevant
  stakeholder-role lenses before each decision, then a single decisive
  one-shot recommendation per decision point, coherent with each other and with
  project intent. Four `gsd-advisor-researcher` agents ran in parallel (grade
  authority, lint scope/landing order, unreachable-claims home, distinctness fix
  depth) at `minimal_decisive` calibration. **The user selected the researcher's
  recommended option in every area**, and then chose the coherence-preserving
  branch on both follow-ups. Full tables and rationale: `14-DISCUSSION-LOG.md`.

- **All four converged independently on one principle**, which is why this phase
  reads as one design rather than four: *every instrument must execute its claim
  rather than read an artifact.* It shows up as the grade **cap** rather than a
  string check, the `-list` **resolution** rather than an intent reading, the
  executed **probe** rather than a grep over source, and the frozen **collision
  control** rather than a regenerable golden.

- **Framing to lead the plan with, because it is the first objection a reviewer
  will raise on each piece:**
  - On EVD-02: "this crosses `TestDebtRegistersAreWellFormed`'s deliberate
    honesty/shape line." It does not — see D-14-02; the cap mechanizes only the
    pass side, and nobody games a control by under-claiming.
  - On EVD-03/04: "a phase number is a perfectly good trigger." It is not — see
    D-14-22; that is an artifact read, and `lit`'s XPASS is the working precedent.
  - On DX-08: "the diagnostic is one token late, just move the span." It is not
    late, and moving it does not separate the three programs — see D-14-34.
  - On DX-08's worth: "`if` isn't in the language, why fix this now." Because
    three distinct sources currently mint **one `result:` ID**, which is evidence
    confusion in a product whose product is evidence — see D-14-43.

- **Two numbers the plan should quote rather than re-derive**, both measured on
  this tree: the static test index resolves in **7ms** vs **3.35s** for
  whole-module `go test -list` (D-14-10), and the full suite passes **unmodified**
  with the DX-08 patch applied (D-14-38).

- ⚠ **Unpackaged spikes detected.** `.planning/spikes/MANIFEST.md` exists with no
  findings skill. Run `/gsd-spike --wrap-up` to make those findings available to
  the researcher and planner. Separately, S-009 and S-010 run **alongside** this
  phase; S-009 informs the DX work but does **not** block it (D-14-44), and S-010
  is a hard entry gate for Phase 18, not Phase 14.

</specifics>

<deferred>
## Deferred Ideas

- **`surface.not_in_language` and the `not_in_language` capability manifest** —
  deferred by ADVERSARIAL-SYNTHESIS §C5 and by `REQUIREMENTS.md` §Deferred. Not
  available to this phase; DX-08 must be satisfied without it. Landing: later,
  and its value *decreases* as M003 ships surface.
- **Under-scoped verification rows (R2b) driven to zero** — detected and pinned in
  Phase 14 (D-14-16), closed with **QLT-10 reconciliation in Phase 20**. Roughly
  4-5× the dead-pattern class; re-adjudicating what each row was meant to prove is
  judgment work, not mechanical.
- **D-13-34's held-out corpus fix** — record the trigger only (D-14-30). Attaches
  to the phase that adds arithmetic or control flow, where a genuinely distinct
  move/borrow program is constructible and D-13-26's topology triple stops
  degenerating. Fixing it now buys a green checkbox from a 112-element space.
- **Distinctness corpus growth as the language grows** — nothing currently forces
  it. When `if` enters the language the spiral trio becomes a *valid* program and
  must be replaced by the then-current not-in-language surface (D-14-43b).
- **`Cause.Span` identity-bearing while `Repair.Span` is not** — the
  "a coordinate shift must never move a diagnostic's ID" principle stays
  half-enforced, and D-14-33 bounded-widens the hole for broken programs only.
  Carried as debt with a witness; becomes forcing if modules require producer-side
  function identity on `Cause`.
- **`check.go:3123`'s `declare_foreign_symbol` shell repair** — still the only
  repair advertising a `Kind` it cannot apply. Carried from Phase 13. A future
  phase must pick completing it or deleting it; deleting churns an identity-bearing
  `RepairKinds` under `/1`.
- **Cluster A — the six single-function emitters** (D-11-02, D-12-36, D-11-27) —
  **Phase 16's** business (NAT-08/NAT-09), not Phase 14's. Base rate on attaching
  it to a feature phase is **0 for 2**; D-14-31's two-milestone-carry guard is the
  mechanization that makes a third deferral visible.
- **Cluster D — the D-09-51 negative-control verdict-flip review** — a human
  review, not code. Belongs in someone's plan as a named `gate="blocking-human"`
  checkpoint; not a Phase 14 deliverable.
- **QLT-12 — content-addressing the enumerated-closure proof** (58.33s, 30% of
  suite wall-clock). Phase 20. Relevant here only as the escape hatch if
  D-14-10's 3.35s accuracy control ever comes under wall-clock pressure.
- **A measured agent-trajectory study** (20 trials with a real model) to replace
  the hand-simulated spiral proxy — named by the research as "the single most
  valuable follow-up experiment". Not scheduled; S-009 is the cheap partial
  substitute.
- **Making Nyquist a hard gate rather than a skill** (research F4; recorded as
  M001 Lesson 5 and M002 Lesson 5, same mechanism, two milestones) — the one
  remedy this project has written down twice and never implemented. **Not in
  EVD-01..08**; QLT-10's "no VALIDATION file remains `draft`" is the M003-scoped
  slice of it.
- **Demoting the integration checker by contract** (research F5) — dropping its
  requirement-satisfaction column entirely once satisfaction is derived from
  graded evidence. Follows naturally from EVD-02 but is a GSD-process change, not
  a repo change; note it, do not act on it here.

</deferred>

---

*Phase: 14-evidence-instrument-and-honest-scoping*
*Context gathered: 2026-09-17*
