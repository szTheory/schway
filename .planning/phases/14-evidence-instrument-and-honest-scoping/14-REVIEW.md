---
phase: 14-evidence-instrument-and-honest-scoping
reviewed: 2026-09-18T00:00:00Z
depth: standard
files_reviewed: 26
files_reviewed_list:
  - cmd/lang-repair/repair.go
  - cmd/lang-repair/repair_diagnosis_test.go
  - internal/compiler/cgen/cgen_program_test.go
  - internal/compiler/check/diagnostic_distinctness_test.go
  - internal/compiler/native/native_lto_test.go
  - internal/compiler/native/native_test.go
  - internal/compiler/native/symbols_test.go
  - internal/compiler/session/evidence_grade_test.go
  - internal/compiler/session/qlt02_budget_manifest.json
  - internal/compiler/session/self_describing_docs_test.go
  - internal/compiler/session/session_payload_replay_test.go
  - internal/compiler/session/session_phase11_differential_test.go
  - internal/compiler/session/session_phase5.go
  - internal/compiler/session/session_phase5_alias.go
  - internal/compiler/session/session_phase5_alias_test.go
  - internal/compiler/session/session_phase5_corpus_test.go
  - internal/compiler/session/session_phase5_test.go
  - internal/compiler/session/session_phase6_budget.go
  - internal/compiler/session/session_phase6_evidence_test.go
  - internal/compiler/session/session_phase6_injectors_test.go
  - internal/compiler/session/session_test.go
  - internal/compiler/session/verification_groundedness_test.go
  - internal/compiler/session/witness_registry_test.go
  - internal/compiler/syntax/parser.go
  - internal/compiler/syntax/parser_skipped_region_test.go
  - scripts/assert-reconciliation-touched.sh
  - scripts/evidence-run-record.sh
findings:
  critical: 0
  warning: 3
  info: 1
  total: 4
status: issues_found
---

# Phase 14: Evidence Instrument and Honest Scoping — Code Review Report

**Reviewed:** 2026-09-18
**Depth:** standard (with targeted deep-dives into the four items the brief flagged for verification)
**Files Reviewed:** 26
**Status:** issues_found (Warning-tier only; no Critical findings)

## Summary

This phase is unusually inspectable because almost all of it *is* the test
suite, and the plan's own D-14-30/D-14-41/D-14-28 discipline requires each
guard to carry a non-inertness proof. I read the two touched production
files (`cmd/lang-repair/repair.go`, `internal/compiler/syntax/parser.go`)
and the two small production edits inside `session_phase5*.go` /
`session_phase6_budget.go` in full, ran the diagnostic-distinctness suite
live and captured real `causes[0].span` values to check the D-14-36
measurement table against the actual implementation, and traced the
self-reported "worth investigating" item (two frontier tests deriving only
`WIRED` under the corpus-wide cap) to its root cause in the run-record
timeout budget.

**What is genuinely clean:** `cmd/lang-repair/repair.go`'s new
`classifyDecline`/`diagnosisCodeList` machinery is careful, bounded, and
free of the prose-leak class of bug the import-boundary tests exist to
catch. `parser.go`'s `recoverRegion` correctly separates the reported token
from the discarded region (verified live: `primary_span` is the `if`
token, `causes[0].span` starts one byte later at the next token, and this
already-existing separation is exactly what makes the three spiral
programs mint distinct diagnostic IDs — confirmed by running
`TestDiagnosticDistinctnessIsOne`/`GuardIsNotInert`, both pass on real
data, not a golden mock). The two new shell scripts pass every argument as
a discrete argv element (no `eval`, no unquoted interpolation into a shell
string), so there is no injection surface in either. The D-14-25
suppression-citation grammar's narrowing to `PENDING-NN-NN` for the
module-wide comment/string-literal surface (self-reported item 1) is
real and documented at the call site (`witness_registry_test.go:208-239`)
with the false-positive count that justified it — sound as described. The
`DEFINED`-grade exemption (item 3) is likewise narrow and matches
`session_test.go:2942,2960-2961` exactly as disclosed.

**What is not clean**, detailed below: the anchoring fix for the
`-run`-pattern self-recursion bug (self-reported item 6) was applied to
two data rows by hand rather than to the code path that constructs
consolidated patterns, so the same failure class can recur silently for
any of the corpus's other ~140+ evidence cells; the `evidenceRunRecordTimeout`
budget's own doc comment is now measurably false and the underlying
timing-margin bug that produced the "unexplained WIRED" grades
(self-reported item 4) is still live in the tree, merely dormant; and the
`//go:build` suppression surface D-14-25 says it enumerates is collected
but never checked against anything, so it is structurally unable to ever
fail.

## Warnings

### WR-01: Run-record timeout budget is measurably too tight, and its own doc comment is stale — the root cause of the two "unexplained WIRED" frontier grades

**File:** `internal/compiler/session/evidence_grade_test.go:230-243`
**Issue:** `evidenceRunRecordTimeout` is `300 * time.Second`, and the comment
directly above it says: *"corpusRunRecord's consolidatePkgPatterns call...
measured at ~136s wall-clock; 300s leaves ample margin."* That claim is
now contradicted by the phase's own later work: commit `0dcb460`'s message
states the corpus-wide session-package run "now runs in ~269-347s" after
the anchoring fix — i.e. up to **347s against a 300s ceiling** — and
`PHASE-14-DEBT.md` row `D-14-121` (added in the same commit) records an
actual reproduction: *"The whole corpus-wide test completed at 300.11s,
suspiciously close to `evidenceRunRecordTimeout`'s 300s ceiling... several
packages' citations falling back to an unresolved WIRED ceiling rather than
their true EXERCISED grade."* This is exactly the mechanism behind
self-reported item 4 (two frontier tests deriving `WIRED` instead of
`EXERCISED` under the corpus-wide cap despite passing standalone): when
`generateRunRecord`'s `context.WithTimeout` fires mid-batch, `go test`'s
`-json` stream is truncated, so whichever tests hadn't yet reported
`"Action":"pass"` at that instant silently get no entry in `parseRunRecord`'s
maps — `record.passedNoSkip(name)` then returns `false` for tests that in
fact passed, and `deriveCeiling`/`validationRowProblem` fall back to
`WIRED`. Because `validationRowProblem` checks `declared > ceiling` for
**every** row regardless of the `validationGradeBarExemptions` file-scoped
bar exemption, this is not a dormant risk confined to the two rows that
happened to hit it once — on a slower CI machine or under load, *any*
currently-`EXERCISED`-declared row in the corpus can spuriously fail
`TestValidationRowGradesAreEarnedOverArchivedCorpus` with "declared grade
EXERCISED exceeds the ceiling WIRED" for a reason that has nothing to do
with the claim being graded. D-14-121 already records this as debt, but it
is `UNOWNED(none-yet-scheduled)` and the fix (raising the timeout or
splitting the batch) was never applied — the doc comment claiming "ample
margin" makes this look safer than the phase's own evidence shows it to
be.
**Fix:** Update the comment to state the true measured range (269-347s,
not ~136s) and either raise `evidenceRunRecordTimeout` with real margin
above the observed worst case (e.g. 480-500s, still comfortably under Go's
600s per-package ceiling), or split the session-package batch so no single
`evidence-run-record.sh` invocation is close to the ceiling. Until then,
this gate is flaky in a direction that produces false CI failures, not
false passes — but a "guard that fails for reasons unrelated to the claim"
is exactly the failure mode this phase exists to retire elsewhere.

### WR-02: The `-run` self-recursion fix was applied to two data rows by hand, not to the code that builds consolidated patterns — nothing prevents recurrence

**File:** `internal/compiler/session/evidence_grade_test.go:802-814` (`pkgPatternsFor`), `:966-980` (`consolidatePkgPatterns`)
**Issue:** Per commit `0dcb460`'s own message, an unanchored
`TestValidationRowGradesAreEarned` pattern from a `14-VALIDATION.md`
evidence cell, once merged by `consolidatePkgPatterns` into the
`internal/compiler/session` package's single alternation, matched
`TestValidationRowGradesAreEarnedOverArchivedCorpus` as a substring and
caused that (already expensive, ~130s+) test to recursively re-invoke
itself inside its own generated run-record subprocess, pushing the whole
package past Go's 600s per-package timeout. The fix committed was to add a
trailing `$` to the two offending cells' patterns in the markdown table
(`14-VALIDATION.md:87-88`), **not** to `pkgPatternsFor`, which still
returns `parsed.Pattern` verbatim (line 811) with no anchor, and not to
`consolidatePkgPatterns`, which still joins raw, unanchored patterns with
`")|("` (lines 977-979) and has no anchoring or self-reference check of
its own. The corpus has 140+ `go test -run` evidence cells (D-14-20's own
floor); the fix only touched the two that were observed to collide. Any
future or currently-uninspected cell whose pattern happens to be an
unanchored substring of another top-level test name in the same package —
particularly one of this file's own several expensive
`TestValidationRowGradesAreEarnedOverArchivedCorpus`/`TestVerificationGroundedness*`
names — reproduces the identical multi-minute recursive blowup, silently,
the next time a row cites it. Note the `assert-go-tests.sh`-shaped branch
of `pkgPatternsFor` (line 837) *does* anchor (`^(...)$`) — only the plain
`go test -run` branch does not.
**Fix:** Either anchor each branch's operand programmatically before
joining in `consolidatePkgPatterns` (wrap each `byPackage[pkg]` entry in
`^(...)$` at consolidation time, independent of what the source markdown
cell wrote), or add a mechanized check (mirroring `TestValidationGradeCapIsNotInert`'s
seeded-fault style) that fails when any evidence-cell `-run` pattern is
unanchored and shares a package with another cell whose pattern it would
match as a substring.

### WR-03: The `//go:build` suppression surface is enumerated but never checked — a structurally inert piece of the D-14-25 guard

**File:** `internal/compiler/session/witness_registry_test.go:311-398` (`scanSuppressionSurfaces`), `:400-429` (`suppressionProblems`)
**Issue:** D-14-25 (context, line 419-421) states the guard enumerates
"every `//go:build` constraint outside a declared allowlist." The scanner
does collect these sites (`kind: "build-constraint"`, lines 353-357), but
`suppressionProblems` only ever raises a problem for (a) a citation-shaped
substring that fails to resolve, or (b) a `kind == "skip"` site with no
citation at all (line 424: `if site.kind == "skip" && !found`). There is
no equivalent check for `kind == "build-constraint"` — no declared
allowlist exists anywhere in this file, and a `//go:build` line containing
no citation-shaped substring (the overwhelmingly common case — e.g.
`//go:build darwin`) produces `found=false, problems=nil` and is silently
dropped, exactly like a Tier-B-exempt file would be, except here there is
no positive exemption declaration at all — the check for this surface
simply does not exist. Today the tree has **zero** `//go:build` lines
anywhere (`grep -rl '^//go:build'` over the module returns nothing), so
this part of the guard has never been exercised against real content
either way. It is also absent from `TestSuppressionWitnessGuardIsNotInert`'s
three seeded faults (D-14-28: uncited `t.Skip`, mismatched `callsite:`
count, neutralized probe) — the guard's own non-inertness proof does not
cover this surface, consistent with it having no assertion to prove
non-inert. This is the class of finding the review brief calls out by
name: a check that cannot fail regardless of what the tree contains.
**Fix:** Either implement the stated allowlist check (fail when a
`//go:build` constraint appears outside a small, explicitly declared set —
e.g. GOOS/GOARCH constraints only — and require anything else to carry a
citation), or narrow the doc comment and D-14-25's own text to state that
build-constraint enumeration is currently informational/unused, and add
the missing seeded-fault case once real enforcement exists.

## Info

### IN-01: `callsite:` witness counter matches by selector name only, not by receiver/package — no observed false positive today, but the design repeats grep's own risk class

**File:** `internal/compiler/session/session_test.go:2841-2857` (`debtRegisterCallSiteWitnessMatches`)
**Issue:** For a `*ast.SelectorExpr` call (`x.Symbol(...)`), the scanner
counts a match purely on `fn.Sel.Name == symbol`, without checking the
receiver's type or that it resolves to the same package-level identifier
the `declared` check found. This is scoped to a single package directory
per call, which bounds the blast radius, and for the one witness this
phase actually ships (`callsite:internal/compiler/check.resolveBlame=0`)
there is no collision — `resolveBlame` is a free function
(`internal/compiler/check/check.go:769`), so its own call sites are all
matched via the `*ast.Ident` branch, and I confirmed by `grep` that no
other declaration of that name exists in the package. But D-14-22's own
framing is that a witness must not silently stop meaning what it claims
when the tree changes around it (that is the whole argument against the
grep-based predecessor it replaces). A future `callsite:` witness naming a
common method name (e.g. `Resolve`, `Check`) in a package that later grows
an unrelated type with a same-named method would inflate the count without
any test catching it, reintroducing a narrower version of exactly the
defect class this instrument was built to retire.
**Fix:** When the witnessed symbol is a package-level function (the
`declared` check already establishes this), restrict the call-site count
to the `*ast.Ident` branch only (an unqualified same-package call), or
additionally verify the `SelectorExpr`'s `X` resolves to the same package
import path before counting it.

---

_Reviewed: 2026-09-18_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
