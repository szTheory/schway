# Phase 14: Evidence Instrument and Honest Scoping - Research

**Researched:** 2026-09-17
**Domain:** Instrument-over-own-artifacts (Go test-suite guards over `.planning/**` and the compiler's own diagnostic/repair protocol). No external library research applies — this phase installs zero packages.
**Confidence:** HIGH — every load-bearing claim below was executed against this tree on 2026-09-17 (Go 1.24.0 darwin/arm64), not read from prose.

## Summary

Phase 14 has no library-research surface: it is a set of Go tests that read
`.planning/**` markdown, the compiler's own diagnostic identity, and the
`lang-repair` protocol, and it touches exactly one production file
(`internal/compiler/syntax/parser.go`). `14-CONTEXT.md`'s 44 `D-14-*` decisions
already constitute a de facto plan — this research's job was to re-run the
load-bearing commands and re-locate the load-bearing lines myself rather than
trust the context document's citations, per this phase's own thesis that an
instrument must execute its claim. **Every command and line citation checked
below reproduced.** One decision premise did **not** hold on the current tree
and must be corrected before planning (see Assumptions/Discrepancies): PROJECT.md's
DX-06 and `-flto`/NAT-07 corrections that EVD-07 asks for are **already shipped**
(commit `d21db90`, 2026-09-17, milestone kickoff) — only the debt-row half of
EVD-07 (a `D-11-25` row with an owning phase) remains outstanding.

**Primary recommendation:** Plan exactly the five deliverables `14-CONTEXT.md`
already scopes (groundedness lint, grade vocabulary + cap, witness/unreachable
registry, diagnostic distinctness + non-empty diagnosis, self-describing-doc
corrections), reusing `phaseArtifactGlob` / `testsupport.ProjectPath` /
`debtRegisterSeverities` / `checkDebtRegister` exactly as they exist today —
do not invent new path-resolution or table-parsing plumbing. Re-measure
D-14-19's pinned frontier set and D-14-20's corpus floors at plan time rather
than copying the numbers below verbatim, because two of them already drifted
slightly from `14-CONTEXT.md`'s stated figures (see Discrepancies).

## Architectural Responsibility Map

This phase has no browser/API/DB tiers — it is single-process Go tooling
plus a static-corpus compiler frontend. The relevant "tiers" are this
project's own layering:

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Groundedness lint over `.planning/**` (EVD-01) | `internal/compiler/session` (new `_test.go`) | `scripts/assert-go-tests.sh` (selection-time sibling, unchanged) | `phaseArtifactGlob`/`testsupport.ProjectPath` already live here; two non-overlapping laws per D-14-09 |
| Grade vocabulary + derivation cap (EVD-02) | `internal/compiler/session` (new `evidence_grade_test.go`) | `.planning/**/*-VALIDATION.md` (data) | Test owns the ceiling logic; markdown owns the declared grade |
| Unreachable-claims / witness registry (EVD-03/04/05) | `internal/compiler/session` (extends `checkDebtRegister`) | `.planning/**/*-DEBT.md` (data), `.planning/UNREACHABLE-CLAIMS.md` (generated view) | Extends the existing debt-register law rather than minting a third store |
| Self-describing doc self-checks (EVD-06/07/08) | `internal/compiler/session` (new tests over `LANGUAGE-MATURITY.md`, `qlt02_budget_manifest.json`) | `.planning/PROJECT.md`, `.planning/LANGUAGE-MATURITY.md` (data, PROJECT.md already corrected) | Tests read the doc's own embedded commands and assert the doc's stated numbers match |
| Diagnostic distinctness (DX-08) | `internal/compiler/syntax` (parser fix), `internal/compiler/check` (new gate test) | `internal/compiler/diagnostic` (identity basis, unmodified) | Fix lives in the parser; the gate is a Go test over parser+diagnostic output, no new schema |
| Non-empty diagnosis on decline (DX-09) | `cmd/lang-repair` (`repair.go`) | none — `import_boundary_test.go` forbids `internal/` imports here | Protocol-boundary process; must not decode prose, only codes |
| PRC-01 well-formedness gate | `internal/compiler/session` (extends `checkDebtRegister`) | none | Same law as the debt register itself |

## Standard Stack

Not applicable in the conventional sense — **zero external dependencies are
added or considered** (a project-wide hard constraint per `STANDING-VERDICTS.md`).
The "stack" for this phase is exclusively:

| Tool | Version (verified) | Purpose |
|------|---------------------|---------|
| Go stdlib `testing` | go1.24.0 (`go version` run on this tree) | Every new guard is a `_test.go` file; no third-party test framework anywhere in `internal/compiler` |
| `go/parser`, `go/build`, `go/ast` (stdlib) | go1.24.0 | D-14-10's static test-index resolver; D-14-25's suppression-surface scan; D-14-40's lexer-only distinctness predicate |
| `go test -list` / `go test -json` | go1.24.0 | Accuracy cross-check for the static resolver; run-record evidence for the grade cap |

No `npm view` / `pip index` / `cargo search` verification applies — there is
no package ecosystem here. **Package Legitimacy Audit is therefore N/A for
this phase** (see below).

### Installation

None. All new files are `_test.go` (or one production file,
`internal/compiler/syntax/parser.go`) inside the existing module
`github.com/codename-lang/lang`.

## Package Legitimacy Audit

**N/A — this phase installs zero external packages.** `go.mod` is untouched;
`.planning/STANDING-VERDICTS.md`'s "zero external production dependencies"
verdict is a hard constraint this phase does not touch. No `package-legitimacy
check` run was performed because there are no package names to check.

## Architecture Patterns

### System Architecture Diagram

```
.planning/**/*.md  (VALIDATION/VERIFICATION/DEBT/SUMMARY/UAT/MIDPHASE-GATE)
        │
        │  read by go/parser-free line scanner (D-14-15)
        ▼
verification_groundedness_test.go ──┐
  - Tier A/B classification (D-14-11)  │  static test index
  - R1 elision scan                    │  (D-14-10: go/parser over *_test.go,
  - R2/R2b pattern→test resolution ────┤   NOT go test -list)
  - R3 grep groundedness (exec'd)      │
  - frontier pin (D-14-18)             ▼
        │                       go test -list .  (accuracy control only,
        ▼                        D-14-10's TestStaticTestIndexMatchesGoTestList)
  violation set ──(byte-compare)──► pinned literal in test source

*-VALIDATION.md rows ──► evidence_grade_test.go ──► deriveGrade(row) ──► ceiling
  (Grade / Non-inertness cols)   (D-14-03 ladder:        │
                                  resolve → pass/no-skip   capped by declared grade
                                  → distinct twin →        (author writes Grade;
                                  MUTATION-KILLED)          test asserts declared<=ceiling)

*-DEBT.md rows ──► checkDebtRegister (EXTENDED) ──► Grade/Witness/Landing-phase
  (Witness col: probe:/callsite:/escape:/env:)   closed vocab   │
                                                                  ▼
                                          .planning/UNREACHABLE-CLAIMS.md  (GENERATED,
                                          .planning/EVIDENCE-RECONCILIATION.md byte-compared)

testdata/*.lang (spiral trio + corpus) ──► syntax.Parse ──► diagnostic.Cause{skipped_region}
                                              │
                                              ▼
                                     diagnostic identity {Schema,Code,Span,Causes}
                                              │
                                              ▼
                              diagnostic_distinctness_test.go (3 unique ID sets)

lang-repair --json ──► selectRepair() ──► Outcome{diagnosis_code, diagnosis_codes[],
  (on decline)                              decline_reason, best_applicability}
```

### Recommended Project Structure (new/extended files only)

```
internal/compiler/session/
├── verification_groundedness_test.go   # NEW — EVD-01 lint + frontier pin + not-inert proof
├── evidence_grade_test.go              # NEW — EVD-02 derivation ladder + cap + exemptions
├── session_test.go                     # EXTENDED — checkDebtRegister grows Grade/Witness/closed Landing-phase
├── session_phase5.go                   # EDITED — Phase5AssertMutationMovesAnAxis DELETED
├── session_phase5_alias.go             # EDITED — stale PENDING-05-08 marker/comment removed
├── session_phase5_alias_test.go        # EDITED — fixture-path skip on allocator_mismatch.lang removed
├── session_phase5_test.go              # EDITED — TestNoNAT03RowRemainsPending replaced by suppression-witness guard
├── qlt02_budget_manifest.json          # EDITED — new `observed` wall-clock row
internal/compiler/syntax/
├── parser.go                            # EDITED — ~14 lines: default: arm captures skipped region, recoverRegion, variadic p.problem
internal/compiler/check/
├── diagnostic_distinctness_test.go     # NEW — 3 assertions, ~60 lines
cmd/lang-repair/
├── repair.go                            # EDITED — Outcome gains 3 omitempty fields, selectRepair returns decision
├── repair_diagnosis_test.go            # NEW — TestUnrepairableAlwaysCarriesDiagnosis + not-inert proof
testdata/distinctness/                   # NEW — spiral trio + ~10-12 fixture corpus + collision_control.json (frozen)
.planning/
├── UNREACHABLE-CLAIMS.md               # NEW, GENERATED, byte-compared
├── EVIDENCE-RECONCILIATION.md          # NEW, GENERATED, byte-compared
├── PROJECT.md                           # ALREADY CORRECTED for DX-06/-flto (verify, do not re-edit blindly)
├── LANGUAGE-MATURITY.md                # EDITED — count corrected from 32 to the re-measured figure
├── **/*-VALIDATION.md (13 archived)    # EDITED — Grade/Non-inertness columns added, Status retired
├── **/*-DEBT.md (12 files)             # EDITED — Grade/Witness columns, closed Landing-phase vocabulary
```

### Pattern 1: Extend, never mint a third law
**What:** Every new EVD/PRC guard extends `checkDebtRegister` /
`phaseArtifactGlob` / `debtRegisterSeverities`'s established shape rather than
authoring a parallel table-parser.
**When to use:** Any time a new column (Grade, Witness, closed Landing-phase)
is added to `*-DEBT.md` or `*-VALIDATION.md`.
**Example (verified, `internal/compiler/session/session_test.go:2613-2660`):**
```go
// TestDebtRegistersAreWellFormed is the mechanical half of the debt-register
// checkpoint ... Register *honesty* ... stays a human reading and is
// deliberately not claimed here.
func TestDebtRegistersAreWellFormed(t *testing.T) {
	registers, err := phaseArtifactGlob("*", "*-DEBT.md")
	...
	for _, path := range registers {
		t.Run(filepath.Base(path), func(t *testing.T) {
			checkDebtRegister(t, path)
		})
	}
}

var debtRegisterSeverities = map[string]bool{"blocker": true, "warning": true, "info": true}
```
This is the literal template D-14-01/D-14-23/D-14-24's new vocabularies (grade,
witness kind, `Landing phase` form) must copy.

### Pattern 2: Non-inertness proof ships with every new guard
**What:** Every control seeds at least one fault in a temp copy and asserts
red.
**When to use:** All five new EVD/DX guards (D-14-28's three-seeded-fault
suppression guard is the richest instance).
**Example (existing precedent to mirror, confirmed in tree):**
`TestInjectorMarkerCountGuardIsNotInert` and
`TestRepairDriverSourceHeldoutScanIsNotInert` (both exist in
`internal/compiler/session`) use a temp-copy-then-mutate pattern; new guards
(`TestVerificationGroundednessIsNotInert`, `TestValidationGradeCapIsNotInert`,
`TestSuppressionWitnessGuardIsNotInert`, `TestDiagnosticDistinctnessGuardIsNotInert`,
`TestUnrepairableDiagnosisGuardIsNotInert`) all follow this shape.

### Anti-Patterns to Avoid
- **Grep-over-source as a guard.** D-14-22's central lesson: a grep for
  `check.go:255` breaks silently on refactor/rename. Every unreachability
  claim must be an *executed* probe (`probe:TestXxx`) or a `go/ast` call-site
  count (`callsite:`), never a source-text search.
- **Regenerable golden files ("bless" buttons).** D-14-39's collision control
  and the two generated `.planning/*.md` views are explicitly
  regenerate-and-byte-compare, never checked-in-and-hand-edited.
- **Count-keyed suppression ceilings.** Explicitly refused (D-14-12, D-14-18)
  — a count-neutral swap must not pass a ceiling silently; always pin an
  **exact literal**, never a bound.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Exact-test-name resolution against `go test -run` semantics | A new regex-matching heuristic | `scripts/assert-go-tests.sh:8-45`'s `-list`-then-exact-match algorithm (verified present, unchanged) | Already correct and already has a shipped `--self-test` sentinel |
| Markdown table parsing with escaped pipes | A new ad hoc splitter | `checkDebtRegister`'s existing frontmatter+table parser (`session_test.go:2628+`), extended with the `\|`-unescape rule (D-14-15) | One parser, one set of edge cases already fought (the `\|` ruling in D-14-17 is a real, reproduced defect — see Verification below) |
| POSIX shell-token splitting for command-cell tokenization | A hand-rolled space-splitter (breaks on quoted args) | stdlib-based ~60-line POSIX-quote splitter (D-14-16's R2 tokenizer) | Command cells contain `-run 'A\|B'`-shaped quoting that a naive `strings.Fields` mis-splits |
| Static Go test enumeration | Shelling out to `go test ./... -list .` per file (146 invocations) | `go/parser.ParseFile(..., parser.SkipObjectResolution)` walk (D-14-10) | Measured 7ms vs 3.35s warm for the shell-out path; keep exactly one `-list`-based accuracy control, not the resolution mechanism itself |

**Key insight:** every "don't hand-roll" item above already has a shipped
implementation in this repo to extend. This phase's entire risk is in the
*derivation logic* (grade ceiling ladder, witness kinds), not the parsing
substrate, which is proven.

## Common Pitfalls

### Pitfall 1: Trusting `\|` as an escaped pipe without re-deriving which reading the corpus actually needs
**What goes wrong:** Treating `PeerLiveness\|LoanChainIndex` (raw markdown) as
the literal regex to run yields `[no tests to run]` and exits 0 — a **fourth**
dead command hiding in a `nyquist_compliant: true` phase.
**Why it happens:** `\|` is GFM's escape for a literal pipe inside a table
cell; the *rendered* cell (unescaped) is what a human copy-pastes and runs.
**How to avoid:** Unescape `\|`→`|` before compiling the alternation, exactly
as D-14-17 rules. **Verified on this tree:**
```
$ go test ./internal/compiler/corevalidate -run 'PeerLiveness\|LoanChainIndex' -count=1
ok  	github.com/codename-lang/lang/internal/compiler/corevalidate	0.171s [no tests to run]
$ go test ./internal/compiler/corevalidate -run 'PeerLiveness|LoanChainIndex' -count=1 -v
=== RUN   TestPeerLivenessFileImportsStayIndependent
--- PASS: TestPeerLivenessFileImportsStayIndependent (0.00s)
```
**Warning signs:** `[no tests to run]` with exit 0 on a pattern that "looks"
like it should match — always re-run under `-list` before trusting a pass or
fail.

### Pitfall 2: A skip/suppression guard that only watches the production file
**What goes wrong:** `TestNoNAT03RowRemainsPending` (verified,
`internal/compiler/session/session_phase5_test.go:229-246`) greps only
`session_phase5_alias.go` for the literal line `// PENDING-05-08`. The
identical string survives as an **error message** at
`session_phase5_alias.go:523` (`"row not yet subjected — see PENDING-05-08"`)
and as **prose** in `session_phase5_test.go:207` / comments in
`session_phase5_alias_test.go`, both invisible to the guard.
**Why it happens:** The guard's own scope (`os.ReadFile` one path, compare
one exact trimmed line) was written narrowly to catch the specific marker
comment it was designed against, not the general suppression-citation class.
**How to avoid:** D-14-25's replacement guard parses **every** package
(including `_test.go`) with `go/parser` in `ParseComments` mode and enumerates
`t.Skip*` calls, `//go:build` constraints, `*ast.Comment` **and**
`*ast.BasicLit` of kind `STRING` matching the citation patterns — not a single
file path.
**Warning signs:** A suppression guard whose implementation is `os.ReadFile`
+ `strings.Split` + exact-line-match on one named file.

### Pitfall 3: A generated reconciliation view that vacuously agrees when both sides are empty
**What goes wrong:** If the register byte-compare only asserts
`generated == checkedIn`, an empty generator against an empty checked-in file
compares equal — a false green with zero content.
**Why it happens:** Byte-compare alone proves consistency, not completeness.
**How to avoid:** D-14-14(c) requires an additional non-zero-row-count
assertion whenever any register holds a qualifying row.
**Warning signs:** A `.planning/UNREACHABLE-CLAIMS.md` or
`.planning/EVIDENCE-RECONCILIATION.md` that stays empty across several phases
without an explicit assertion that it *should* be non-empty given known
`*-DEBT.md` content (today: at minimum D-13-02b, D-13-10a, D-12-43, D-13-34,
the `-flto` inertness, D-11-02 — six rows already known to qualify).

### Pitfall 4: Re-deriving the "3 dead patterns" or "32 guards" numbers from prose instead of re-running the source command
**What goes wrong:** `14-CONTEXT.md` and `ROADMAP.md` both state fixed counts
("three such patterns," "32 guards," "28 Tier-A files," "1121 test
functions"). Some of these have already drifted (see Discrepancies below).
**Why it happens:** These are living measurements over a tree that keeps
changing; citing them without re-running invites exactly the kind of stale
self-description this phase exists to close.
**How to avoid:** Every count that appears in a plan or in a test's pinned
literal must be the output of a command run at plan/authoring time, not
copied from this document or from `14-CONTEXT.md`.
**Warning signs:** A pinned literal or corpus-floor assertion whose value
does not match a freshly re-run version of the command that produced it.

## Code Examples

### The exact defect EVD-01 must catch (verified reproduction)
```bash
# Source: internal/compiler/session/session_phase5_test.go:229-246 pattern,
# applied to the ROADMAP's cited 08-VALIDATION.md rows.
$ go test ./internal/compiler/check/... -run TestComputeLoanLastUsesAndDerivePlaceLoansAgree -v
testing: warning: no tests to run
PASS
ok  	github.com/codename-lang/lang/internal/compiler/check	0.166s [no tests to run]

$ go test ./internal/compiler/session/... -run TestAuditQLT02BudgetManifest -v
testing: warning: no tests to run
PASS
ok  	github.com/codename-lang/lang/internal/compiler/session	(cached) [no tests to run]
```
Both exit 0. `TestComputeLoanLastUsesAndDerivePlaceLoansAgree` was deleted
in Phase 09 when `computeLoanLastUses` was removed (confirmed:
`grep -rn computeLoanLastUses internal/` returns nothing production-side);
`TestAuditQLT02BudgetManifest` never existed under that exact name (the real
test is `TestBudgetManifestLoads` / `TestQLT02BudgetManifestFileUnchangedDuringAudit`,
per `14-CONTEXT.md`'s own citation, itself verified present in
`internal/compiler/session/session_phase6_budget_test.go`).

### The static-index resolver's accuracy control target (verified counts, today)
```bash
$ go list ./... | wc -l
25
$ go test ./... -list '.*' 2>/dev/null | grep -c '^Test\|^Fuzz\|^Example\|^Benchmark'
1050
```
`14-CONTEXT.md`'s D-14-10 cites "1121 test functions, 25 packages" — package
count matches exactly; test-function count has drifted by ~7% since that
measurement (expected, given ongoing work). **Re-run this at plan time and
pin whatever the real number is** rather than 1121.

### DX-08's parser mechanism (verified line-for-line)
```go
// internal/compiler/syntax/parser.go:169-171 — parseProgram's default arm,
// the site D-14-35 changes:
default:
    p.problem("syntax.expected_declaration", p.peek(), "expected `data` or `fn` declaration")
    p.recoverUntil(TokenData, TokenFn, TokenForeign, TokenEOF)

// internal/compiler/syntax/lexer.go:10-22 — `if` is absent, confirmed:
var keywords = map[string]Kind{
    "module": TokenModule, "export": TokenExport, "type": TokenType,
    "data": TokenData, "fn": TokenFn, "match": TokenMatch,
    "foreign": TokenForeign, "try": TokenTry, "discard": TokenDiscard,
    "because": TokenBecause, "defect": TokenDefect,
}
// internal/compiler/syntax/parser.go:472 — linearBody() therefore consumes
// `if` as the terminal result identifier via:
result := p.identifier("syntax.expected_linear_result")
```
`recoverUntil` is defined at `internal/compiler/syntax/parser.go:699` (not
`~700` as `14-CONTEXT.md` approximates — off by one line, immaterial).
`p.problem` is a fixed 3-arg function at `parser.go:659` today — it needs
widening to variadic causes for `recoverRegion`'s output to attach, exactly
as D-14-35 specifies.

### DX-09's exact target (verified line-for-line)
```go
// cmd/lang-repair/repair.go:112-117
type Outcome struct {
    Status          string `json:"status"`
    DiagnosisCode   string `json:"diagnosis_code,omitempty"`
    RepairKind      string `json:"repair_kind,omitempty"`
    SubprocessCount int    `json:"subprocess_count"`
}

// cmd/lang-repair/repair.go:200-209 — selectRepair, the empty-diagnosis bug:
func selectRepair(result checkResult) (jsonRepair, string, bool) {
    for _, d := range result.Diagnostics {
        for _, r := range d.Repairs {
            if driverEligible(r) {
                return r, d.Code, true
            }
        }
    }
    return jsonRepair{}, "", false   // <-- the discarded information
}
```

### The identity basis DX-08 does NOT change (verified)
```go
// internal/compiler/diagnostic/diagnostic.go:16-21ish — Span, Cause exist;
// identity struct literals at :114 and :144-145:
struct{ Schema, Code string; Span Span; Causes []Cause }{Schema: Schema, Code: code, Span: span, Causes: causes}
struct{ ...; RepairKinds []string }{Schema: Schema1, Code: code, Span: span, Causes: causes, RepairKinds: repairKinds}
```
`Causes` is already inside the hashed identity struct on both `/0` and `/1` —
confirms D-14-37's "no schema bump" ruling requires no change to this file at all.

## State of the Art

| Old Approach (this repo, pre-Phase-14) | Current/target Approach | When Changed | Impact |
|--------------------------------------|--------------------------|---------------|--------|
| `*-VALIDATION.md` `Status` column (freeform `✅ green`/`⬜ pending`) | `Grade` (closed vocabulary) + `Non-inertness` columns | Phase 14 (D-14-05) | Removes the exact token (`✅ green`) that Tier-A detection currently keys on — sequencing hazard, see Pitfall discussion in `14-CONTEXT.md` D-14-11 |
| `Landing phase` cell as free prose (`"OPEN and UNOWNED — reopens only when…"`) | Closed vocabulary `P<NN>` \| `CLOSED(<sha>)` \| `UNOWNED(<witness-id>)` | Phase 14 (D-14-24) | Verified today: `PHASE-13-DEBT.md`'s D-13-02b/D-13-10a/D-13-34 rows all use exactly this free-prose form and will need migrating |
| `t.Skip`/marker-comment suppression cited by phase number or prose | `probe:`/`callsite:`/`escape:`/`env:` witness tokens, each independently executed | Phase 14 (D-14-22/23) | LLVM `lit` XFAIL/XPASS is the working precedent already cited in `14-CONTEXT.md` |
| `diagnosis_code` empty on decline | `diagnosis_codes[]`, `decline_reason`, `best_applicability` additive fields | Phase 14 (D-14-42) | No protocol break — `Outcome` carries no schema string, so D-04-23/D-13-18's additive-`omitempty` precedent applies directly |

**Deprecated/outdated:**
- `Phase5AssertMutationMovesAnAxis` (`session_phase5.go:412-428`) — a thin
  wrapper with zero added behavior over `AssertMutationMovesAnAxis`
  (`session_phase5_alias.go:497-523`); deleted this phase (EVD-05).

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `14-CONTEXT.md`'s "28 Tier-A files / 146 verification-command pairs" corpus-floor figures (D-14-20) — I could not independently reproduce "28" with a simple glob (my `*VALIDATION.md`+`*VERIFICATION.md` glob found 26; broadening to also match `*MIDPHASE-GATE.md`/`*UAT.md` per D-14-11's Tier-A basename rule adds 4 more, i.e. 30, still not exactly 28). The discrepancy is most likely explained by D-14-11's Tier-A rule also requiring the file to actually *contain* a qualifying table row (a verdict token beside a command span), which a bare glob does not test for. | Common Pitfalls / D-14-20 | Low — the exact number is re-measured mechanically by the lint itself at authoring time (D-14-19 already instructs "re-measure the pin; do NOT assume 'three'"); this just extends that instruction to the floor constants too |
| A2 | The `~192.7s` full-suite wall-clock baseline (EVD-08) — I ran a partial re-measurement (cold `internal/compiler/session` alone: 129.6s; `internal/compiler/native`: 15.9s; `internal/compiler/check`: 6.2s) that is consistent in order of magnitude with 192.7s but I did not do a single clean full-suite cold timing run in this research pass (a background `go test ./...` completed but with warm caches for most packages, showing 13.6s wall due to caching). | Code Examples / EVD-08 | Low-Medium — the planner's Wave 0 task should do one clean `go clean -testcache && time go test ./...` run and record whatever real number comes out, rather than trusting either 192.7s or my partial sum |
| A3 | D-14-10's claimed "1121 test functions" — I measured 1050 via `go test ./... -list '.*'` today. Package count (25) matched exactly. | Code Examples | Low — expected drift; the plan must re-run this exact command before authoring `TestStaticTestIndexMatchesGoTestList`'s expectations |

**If this table is empty:** N/A — three items above, all low-risk measurement
drift, not premise failures. No `[ASSUMED]` architectural claim in this
document rests on unverified training knowledge; every substantive citation
above (file:line, command output) was executed this session.

## Open Questions

1. **Does EVD-07's "PROJECT.md's DX-06 claim and NAT-07 `-flto` bullet state
   what the evidence supports" already fully close, given PROJECT.md was
   corrected in commit `d21db90` (2026-09-17, `docs: start milestone M003`)?**
   - What we know: `PROJECT.md`'s `## Current State` section, verified read
     in full, already states the corrected DX-06 characterization ("the
     entire blame subsystem ... is test-only dead code") and the corrected
     `-flto` characterization ("structurally inert on every cgen-emitted
     multi-function program ... NAT-07's ... control is hand-written C
     explicitly not emitted by cgen"). This reads as already satisfying
     EVD-07's *text-correction* half.
   - What's unclear: whether `14-CONTEXT.md`'s D-14-30 register-population
     table (D-11-25's `-flto` row, "owning phase per PRC-01") intends this as
     new work or as confirmation that a debt row is still missing. I confirmed
     the debt-row half is genuinely missing: `grep -rn D-11-25
     .planning/milestones/M002-phases/*/PHASE-*-DEBT.md` returns nothing.
   - Recommendation: the plan should treat EVD-07 as **one remaining task**
     (add the `D-11-25` debt row with `Grade`/`Witness`/owning-phase per
     D-14-30's table), not two, and should explicitly note in its own
     Verification Map that the PROJECT.md text-correction sub-claim was
     found pre-satisfied rather than re-doing prose edits that are already
     correct.

2. **Exact re-measured value for D-14-19's pinned frontier set.**
   - What we know: the three ROADMAP-named dead patterns
     (`08-VALIDATION.md:51`, `08-VALIDATION.md:63`, `09-VALIDATION.md:85`)
     all reproduced exactly as `[no tests to run]`/exit 0 in this session.
     `12-VALIDATION.md:24`'s `<TestName>` template row and 11-VALIDATION.md's
     multiple `…`-elided command/path cells (confirmed at lines 51, 56, 58,
     60, 63, 68, 69, 71) also reproduced.
   - What's unclear: the *exact* final count once D-14-17's `\|`-unescape
     ruling and R2b's alternation-branch rule are both applied — D-14-19
     explicitly defers this arithmetic to plan time ("the plan measures the
     real set before pinning it").
   - Recommendation: author `TestVerificationGroundednessFrontierIsPinned`'s
     exact literal only after running the finished R1+R2+R3 scanner once,
     never by hand-counting from this document or `14-CONTEXT.md`.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` (`go test`), go1.24.0 — confirmed no third-party test framework in `internal/compiler` or `cmd/` |
| Config file | none — plain `_test.go` files colocated with their packages (confirmed, matches `12-VALIDATION.md`'s own "Framework/Config file" row) |
| Quick run command | `go test ./internal/compiler/<package>/... -run <TestName> -count=1` |
| Full suite command | `go test ./...` (confirmed green, exit 0, on this tree before any Phase 14 change) |

### Phase Requirement → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| EVD-01 | Groundedness lint finds 0 dead `-run` patterns / `…` elisions, with a seeded-dead-pattern non-inertness control | unit | `go test ./internal/compiler/session/... -run TestVerificationGroundedness -v` | ❌ Wave 0 (new: `verification_groundedness_test.go`) |
| EVD-02 | Declared grade never exceeds derived ceiling; row below EXERCISED cannot satisfy | unit | `go test ./internal/compiler/session/... -run TestValidationRowGradesAreEarned -v` | ❌ Wave 0 (new: `evidence_grade_test.go`) |
| EVD-03 | `.planning/UNREACHABLE-CLAIMS.md` byte-matches its regenerated view | unit | `go test ./internal/compiler/session/... -run TestUnreachableClaimsViewIsCurrent -v` | ❌ Wave 0 |
| EVD-04 | A cited witness that no longer holds fails the suite (XPASS-shaped) | unit | `go test ./internal/compiler/session/... -run TestNoSuppressionOutlivesItsWitness -v` | ❌ Wave 0 |
| EVD-05 | Exactly one axis-movement implementation; zero per-row exclusions | unit | `go test ./internal/compiler/session/... -run TestNAT03MutationTableHasSevenRows -v` (extended) plus a grep-free structural check that `Phase5AssertMutationMovesAnAxis` no longer exists | ✅ existing tests extend; wrapper deletion is Wave 0 |
| EVD-06 | `LANGUAGE-MATURITY.md`'s guard count matches a live re-run of its own documented grep | unit | `go test ./internal/compiler/session/... -run TestLanguageMaturityGuardCountIsCurrent -v` | ❌ Wave 0 — confirmed today's real count is **26**, not 32 |
| EVD-07 | PROJECT.md DX-06/`-flto` text correct (pre-verified true); `D-11-25` debt row exists with owning phase | doc + unit | `grep -c "test-only dead code" .planning/PROJECT.md` (already passes); `go test ./internal/compiler/session/... -run TestDebtRegistersAreWellFormed -v` after adding the row | ✅ text already correct; ❌ debt row is new |
| EVD-08 | `qlt02_budget_manifest.json` carries an `observed` wall-clock row with the real baseline | unit | `go test ./internal/compiler/session/... -run TestBudgetManifestLoads -v` (extended) | ✅ existing test extends; new JSON row is Wave 0 data |
| DX-08 | Spiral trio mints 3 distinct diagnostic-ID sets and 3 distinct result IDs; frozen collision control proves the guard is not inert | unit | `go test ./internal/compiler/check/... -run TestDiagnosticDistinctnessIsOne -v` and `-run TestDiagnosticDistinctnessGuardIsNotInert` | ❌ Wave 0 (new: `diagnostic_distinctness_test.go`, `testdata/distinctness/`) |
| DX-09 | `lang-repair --json` never returns `unrepairable` with empty diagnosis | unit | `go test ./cmd/lang-repair/... -run TestUnrepairableAlwaysCarriesDiagnosis -v` | ❌ Wave 0 (new: `repair_diagnosis_test.go`) |
| PRC-01 | A debt row recorded without an owning phase fails well-formedness | unit | `go test ./internal/compiler/session/... -run TestDebtRegistersAreWellFormed -v` (extended for closed `Landing phase` vocabulary) | ✅ existing test extends |

### Sampling Rate
- **Per task commit:** the specific `go test ./<changed-package>/... -run <NewOrChangedTest> -count=1` command from the row above.
- **Per wave merge:** `go test ./internal/compiler/... ./cmd/... -count=1` (the full suite; confirmed green today).
- **Phase gate:** Full suite green (`go test ./...`, exit 0) before `/gsd-verify-work`, plus `TestVerificationGroundednessFrontierIsPinned` matching its committed literal exactly (not a superset, not a subset).

### Wave 0 Gaps
- [ ] `internal/compiler/session/verification_groundedness_test.go` — covers EVD-01 (does not exist yet)
- [ ] `internal/compiler/session/evidence_grade_test.go` — covers EVD-02 (does not exist yet)
- [ ] `internal/compiler/check/diagnostic_distinctness_test.go` — covers DX-08 (does not exist yet)
- [ ] `cmd/lang-repair/repair_diagnosis_test.go` — covers DX-09 (does not exist yet)
- [ ] `testdata/distinctness/` corpus + `collision_control.json` — does not exist yet, and `collision_control.json` MUST be captured **before** the parser fix lands (it is a frozen pre-fix artifact, D-14-39)
- [ ] Framework install: none — stdlib `testing` is already the sole framework project-wide

## Security Domain

`security_enforcement: true` in `.planning/config.json` (confirmed). This
phase touches no network/auth/session surface — it is a documentation lint
and a compiler-frontend diagnostic change — so most ASVS categories are N/A,
but two are load-bearing given the phase's own thesis (an instrument that can
be gamed is a security property of the evidence pipeline itself):

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | no auth surface in this phase |
| V3 Session Management | no | n/a |
| V4 Access Control | no | n/a |
| V5 Input Validation | yes (narrow) | `.planning/**` markdown is parsed by a line-scanner (D-14-15) that must fail closed on unparseable command cells (`unparseable verification command` is a violation, not a silent pass) — this is the closest analogue to input validation in this phase |
| V6 Cryptography | no | n/a — diagnostic identity uses sha256 already, unmodified by this phase |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| A control that can be satisfied by under-claiming or by editing the artifact it grades ("grading from wiring") | Repudiation / Tampering-of-evidence | Execute-the-claim design: derivation ceiling (D-14-01), executed witness probes (D-14-22/23), frozen non-regenerable collision control (D-14-39) — all already specified in `14-CONTEXT.md` and confirmed absent from the current tree (no existing grade/witness mechanism exists to audit) |
| Silent parser-skip on an unrecognized verification-command shape | Tampering (via omission) | D-14-15's explicit rule: an unparseable `go test` command cell is a **violation**, not a pass — confirmed no such classification exists in the tree today (there is no groundedness lint at all yet) |
| Command injection via `grep`/`rg` execution of untrusted `.planning/**` content (R3) | Tampering / Elevation of privilege | D-14-15 restricts R3 execution to commands with no `|`, `>`, `&&`, `;`, `$(` and a literal existing file operand, executed with **no shell** (`exec.Command` argv form, never `sh -c`) — this constraint must be enforced in the implementation, not just documented |

## Sources

### Primary (HIGH confidence — executed this session against the tree)
- `git log --oneline -- .planning/PROJECT.md` and full read of `.planning/PROJECT.md:1-100` — confirms PROJECT.md's DX-06/`-flto` corrections already landed in commit `d21db90`.
- `go test ./internal/compiler/check/... -run TestComputeLoanLastUsesAndDerivePlaceLoansAgree -v`, `go test ./internal/compiler/session/... -run TestAuditQLT02BudgetManifest -v`, `go test ./internal/compiler/corevalidate -run 'PeerLiveness\|LoanChainIndex' -count=1` vs unescaped form, `go test ./internal/compiler/corevalidate -list 'PeerLiveness|LoanChainIndex'` — all four dead/live patterns reproduced exactly as `14-CONTEXT.md` claims.
- `awk '/^func /{f=$0;l=NR} /Functions\) != 1/{print FILENAME": "f}' $(find internal cmd -name '*.go' -not -name '*_test.go')` — reproduces LANGUAGE-MATURITY.md's own documented re-verify command; returns **26**, confirming EVD-06's "32 stated, 26 real" claim exactly.
- Direct `Read`/`grep` of `internal/compiler/session/session_phase5.go:412-428`, `session_phase5_alias.go:435-523`, `session_phase5_test.go:190-246`, `session_phase5_alias_test.go:170-265` — confirms D-14-26/D-14-27's exact deletion targets and the `TestEveryMutationMovesItsClaimedAxis` fixture-path skip.
- Direct `Read` of `internal/compiler/syntax/parser.go:160-175,295-305,465-478,699`, `internal/compiler/syntax/lexer.go:8-24`, `internal/compiler/diagnostic/diagnostic.go` (Cause/Span/identity struct lines) — confirms D-14-34/35/37's exact mechanism.
- Direct `Read` of `cmd/lang-repair/repair.go:90-215` — confirms D-14-42's `Outcome`/`selectRepair` target lines.
- `grep -c "syntax.expected_declaration" internal/compiler/check/check_ordering_stability_test.go` → 0; `sed -n '745,755p' check_blame_test.go` → confirms the only in-repo mention is a code-string set-membership check — matches D-14-38 exactly.
- `grep -rn D-11-25 .planning/milestones/M002-phases/*/PHASE-*-DEBT.md` → no results — confirms the missing debt row EVD-07 still needs.
- `go test ./... ` full run (background) → exit 0, all 25 packages green, confirming the pre-Phase-14 baseline the plan should assume.
- Direct `Read` of `internal/compiler/session/session_test.go:2505-2700` — confirms `phaseArtifactGlob`, `debtRegisterSeverities`, `debtRegisterLandingPhaseExemptions`, `checkDebtRegister`, and today's free-prose `Landing phase` cells (e.g. D-13-02b's row, quoted verbatim above).
- Direct `Read` of `scripts/assert-go-tests.sh:1-45` — confirms the `-list`-then-exact-match algorithm D-14-03/D-14-09 reuse.

### Secondary (MEDIUM confidence)
- `.planning/research/M003/EVIDENCE-AND-DEBT.md`, `ADVERSARIAL-SYNTHESIS.md` — read for cross-reference only; every specific number/claim drawn from them was independently re-executed above rather than cited from prose alone.

### Tertiary (LOW confidence)
- None — this phase's research surface has no external web/library dimension, so no WebSearch-only claim appears in this document.

## Metadata

**Confidence breakdown:**
- Standard stack: N/A (no external deps) — HIGH confidence in that null finding
- Architecture: HIGH — every integration point (`phaseArtifactGlob`, `checkDebtRegister`, parser call sites, `Outcome` struct) verified by direct file read this session
- Pitfalls: HIGH — all four pitfalls are reproduced defects, not hypothesized ones

**Research date:** 2026-09-17
**Valid until:** Re-verify all pinned numbers (dead-pattern count, guard count, test-function count, wall-clock baseline) at `/gsd-plan-phase 14` authoring time — this is a fast-moving tree (commits landing same-day) and the phase's own thesis forbids trusting a stale count. Treat this document's numbers as **7 days** fresh at most.
