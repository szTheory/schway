# Phase 14: Evidence Instrument and Honest Scoping - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-09-17
**Phase:** 14-evidence-instrument-and-honest-scoping
**Areas discussed:** Grade authority (EVD-02), Lint scope + landing order (EVD-01), Unreachable-claims home (EVD-03/04), Distinctness fix depth (DX-08/09), plus two cross-area follow-ups (register topology, enforcement scope)

**Method:** The user asked for maximum-breadth adversarial research across all
relevant stakeholder-role lenses, with pros/cons/tradeoffs, antipatterns,
best practices, footguns, external ecosystem precedent and an adversarial pass,
synthesized into one decisive recommendation per decision point. Four
`gsd-advisor-researcher` agents ran in parallel at `minimal_decisive`
calibration. The distinctness researcher additionally reproduced the spiral,
prototyped two candidate fixes, ran the full suite against each, and reverted —
so its findings below are measurements rather than arguments.

---

## Area 1 — Grade authority (EVD-02)

**Gray area:** Is a row's grade hand-authored and merely vocabulary-checked, or
mechanically derived from what the evidence cell resolves to? Where does the
grade live? Does the M001/M002 back corpus get retro-graded or parse-checked?
How is the REACHABLE/EXERCISED/MUTATION-KILLED boundary decided mechanically?

**Role lenses fanned out:** compiler/toolchain engineer, build & release
engineer, SRE/DevOps, QA/test-architect, security & supply-chain auditor
(evidence provenance, attestation), technical program manager/auditor, DX
engineer, data modeller (markdown-as-database vs sidecar), adversarial reviewer.

**External precedent weighed:** SLSA/in-toto/SPDX tiered attestation, DO-178C /
IEC 61508 assurance levels and RTM degeneration into paperwork, OpenSSF Best
Practices Badge self-assertion-vs-verified split, Rust `#[should_panic]`/`ignore`
rot, Go `testing.T.Skip` invisibility, PIT/Stryker "killed" semantics, coverage
gates and Goodhart, Bazel/Buck test-target metadata, ISO 26262 tool
qualification, docs-as-tests systems, `sphinx-needs` (requirements-with-status-
in-docs) and its documented split-store rot.

| Option | Description | Selected |
|--------|-------------|----------|
| A. Declared grade, mechanically capped | Author writes a `Grade` cell; a Go test derives a ceiling from the row's own evidence cell and fails when `declared > derived`. Mechanizes only the pass side (EXERCISED, MUTATION-KILLED); refusals stay declarative. New `Grade` + `Non-inertness` columns in `*-VALIDATION.md`; freeform `Status` retired. Derivation total over all 13 archived files; the bar enforced only for phases >= 14 via a dated, file-scoped exemption map. ~250-350 LOC, 0 deps. | ✓ |
| B. Hand-authored grade, vocabulary-string validation only | Parser asserts the cell is a member of the closed set, nothing more. ~60 LOC, exact clone of `debtRegisterSeverities`, zero back-corpus cost. | |

**User's choice:** A — Declared, mechanically capped.

**Notes:**
- B was rejected on the grounds that it *is* the shape of the defect the phase
  exists to prevent: `gsd-integration-checker` writing `WIRED` again with a
  spellchecker attached. **No seeded fault can make a string-membership check go
  red**, so it cannot carry a non-inertness proof — and this project requires one
  for every control. B survives only as a temporary shape migration inside A's
  first commit.
- The load-bearing counter-argument to A, engaged head-on:
  `TestDebtRegistersAreWellFormed`'s doc comment deliberately declines to claim
  register *honesty*. Resolution: a grade is an honesty property in general, but
  EVD-02 only needs it to be a **shape property at the bar** — the two grades
  that satisfy a requirement are mechanically decidable; the three that refuse
  are refusals, and nobody games a control by under-claiming.
- **Verified live on this tree (Go 1.24.0):**
  `go test ./internal/compiler/session/... -run TestNoSuchTestNameXYZ` prints
  `[no tests to run]` and **exits 0**. Defect instance 2's mechanism is live today.
- **Constraint found in the corpus:** `-list` enumerates top-level tests only, so
  subtest-scoped claims (several in `13-VALIDATION.md`) must name the parent test.
- **Corpus-driven arm required:** compile-time evidence (`go build` / `BUILD_OK`,
  as used for `blameFieldWitness`) derives `WIRED`, not `EXERCISED` — which
  already matches that row's own prose.
- **Sidecar rejected** on three counts: second source of truth with a sync
  problem (the documented `sphinx-needs`-style rot), needs a parser the zero-dep
  constraint disallows, and the grade must be readable in the same diff hunk as
  the claim it grades.
- **Honest residuals accepted and recorded rather than papered over:** an
  assertion-free test resolves, passes and derives `EXERCISED` — so `EXERCISED`
  certifies resolution and execution, **not** assertion strength; and naming a
  trivially-passing twin to reach `MUTATION-KILLED` is only partially closed.
- **Predicted first finding:** at least one M001/M002 row will derive *below* its
  shipped `✅ green`. That is the instrument working, recorded as debt with an
  owning phase — the same instruction D-13-33 already carries.

---

## Area 2 — Lint scope + landing order (EVD-01)

**Gray area:** Does the lint cover archived milestone docs (it must — all three
dead patterns live there, but rewriting archives to green a lint falsifies the
record)? How is a "command cell" defined so the `…` check does not fire on prose
ellipsis, which appears throughout `PROJECT.md`/`ROADMAP.md` today? How are
patterns extracted and resolved? Land red or fix first? Go test or shell script?

**Role lenses fanned out:** build/release engineer, CI/CD and DevOps (red-main
policy, hermetic builds), compiler/toolchain engineer, static-analysis/linter
author (precision-recall, suppression abuse), technical writer / docs
infrastructure, archivist / records management, security & supply-chain auditor,
DX engineer, adversarial reviewer.

**External precedent weighed:** docs-as-tests rot (Python doctest, rustdoc tested
blocks, mdBook, Go Example functions, `sphinx-needs`, Cucumber step-definition-
not-found), link checkers and allowlist decay, `pytest --strict-markers` /
empty-selection handling, JUnit Surefire `failIfNoSpecifiedTests`, Bazel
`--test_filter` target-not-found semantics, DO-178C/ISO-26262 RTM degeneration,
`nolint` suppression abuse, ratcheting lints and baseline files (ESLint
suppressions, Psalm/PHPStan baselines, Sorbet, TypeScript strict migration).

| Option | Description | Selected |
|--------|-------------|----------|
| A. Go test + static AST resolver, role-scoped corpus, append-only archives, land green with an exact pinned frontier set | New `internal/compiler/session/verification_groundedness_test.go` reusing `phaseArtifactGlob`/`testsupport.ProjectPath`. Static `go/parser` index over `*_test.go` (measured 7ms) plus one 3.35s accuracy control. Archives in scope for detection, append-only for repair via a reconciliation view whose every verdict carries an executable obligation. No suppression syntax at all. Lands green with an exact committed violation literal that must be empty by QLT-10. | ✓ |
| B. Extend `scripts/assert-go-tests.sh` into a shell lint over `.planning/**` | Grow the existing shipped control (which already has a `--self-test`) into a document lint, shelling `go test -list` per pair. One artifact rather than two. | |

**User's choice:** A — Go test + static AST resolver, append-only archives.

**Notes:**
- B rejected: two coexisting laws over one subject (the project's own named
  delayed fuse); shell cannot reliably do `\|`-unescaping and code-span
  extraction; and the script's `--self-test` sentinel proves *selection*, not
  *document scanning*, so the existing non-inertness proof would not cover the
  new behavior. **Both artifacts are kept with non-overlapping jobs, stated in a
  comment in each file.**
- **The archive crux, resolved:** detection IN, editing OUT. A dead row is never
  rewritten in place (that is the falsification the phase exists to prevent);
  excluding archives entirely would make the lint find nothing and be theater.
  Corrections live outside the archive, keyed by
  `(file, line, verbatim original command)`, under a closed verdict vocabulary
  where every verdict carries a checkable obligation — notably
  `obsolete-by-design`, which for `08-VALIDATION.md` OWN-06 becomes a falsifiable
  **absence** assertion that `computeLoanLastUses` no longer exists. That is the
  "adjudication, not a rename" the research called for.
- **Explicitly not a baseline:** a baseline is satisfied by silence; every entry
  here is satisfied only by a claim that can itself fail. Count-keyed ceilings
  refused outright — a count-neutral swap lands unnoticed (the documented ESLint
  bulk-suppressions failure).
- **Measured cost:** static index 7ms (1121 test funcs, 25 packages) vs 3.35s
  warm for whole-module `go test -list`. The research doc's "146 shell-outs"
  framing overstates even the `-list` path; whole-module `-list` is one
  invocation. The 3.35s accuracy control is 1.7% of the 192.67s suite budget and
  is kept as the instrument that keeps the instrument honest.
- **Measured false-positive rate: R1 zero, R3 zero.** All 11 `…` in
  `11-VALIDATION.md` are inside command spans and all genuine — including two
  **path** elisions not previously catalogued. `09-VALIDATION.md`'s single `…` is
  prose; `02-VERIFICATION.md`'s six are digest truncations. `PROJECT.md`,
  `ROADMAP.md`, `REQUIREMENTS.md`, `LANGUAGE-MATURITY.md` and `research/M003/**`
  contribute **zero** matches, because extraction is per-code-span inside table
  rows. A scan for ASCII `...` not preceded by `/` or `.` returns 0 hits, so
  banning it costs nothing and closes the obvious evasion.
- **A fourth dead command found during research.** Copy-pasted verbatim from raw
  markdown, `-run 'PeerLiveness\|LoanChainIndex'` resolves to zero tests and
  exits 0 — in the `nyquist_compliant: true` phase — while the rendered form
  resolves to one test. Ruling: **unescape** (the rendered cell is the command),
  and record the decision, because the other reading produces 32 findings out of
  171 occurrences and the difference is otherwise invisible.
- **`grep` groundedness included** (~10 execs, microseconds) because the archival-
  breakage class — `RETROSPECTIVE` M001 Key Lesson 6's third recurrence, live
  today in `11-02 T3` — is the *second* known way a command silently stops
  verifying, and excluding it means the lint watches only one of two.
- **"No silent skips" identified as the single most important anti-inertness
  rule:** a command containing `go test` that does not parse into
  (packages, pattern) is a violation, not a pass. Parser-based lints die by
  silently skipping what they cannot parse.
- **Landing order:** not fix-first (the lint would never observe a real defect,
  contradicting `REQUIREMENTS.md` §Out of Scope), not red-on-main. Land green
  with an **exact committed literal** — not a count — which must be empty by
  QLT-10. Same idiom as every other frontier fixture in this milestone.
- **Re-measure the pin; do not assume "three."** Today's Tier-A set is the three
  named in the ROADMAP plus `12-VALIDATION.md:24`'s unfilled template row, plus
  three elided patterns and two elided paths in `11-VALIDATION.md`, plus whatever
  the `\|` ruling admits.
- **Corpus floors** (`>= 28` Tier-A files, `>= 140` commands) added so the lint
  cannot go inert by finding nothing — the same failure class as a `-run` pattern
  that silently stops matching.

---

## Area 3 — Unreachable-claims home (EVD-03/04)

**Gray area:** New `.planning/UNREACHABLE-CLAIMS.md`, or extend the existing
`*-DEBT.md` register that already has a well-formedness test and a Landing-phase
column? What closed grammar makes an unblocking trigger mechanically checkable?
What is EVD-04's guard scope and enumeration mechanism? Which artifact is the
source of truth relative to the grade vocabulary?

**Role lenses fanned out:** compiler/toolchain engineer, QA/test-architect (skip
management, quarantine), SRE (known-issues registers and why quarantine lists
become graveyards), TPM/release manager (waiver processes), safety-critical /
regulated-systems engineer (DO-178C, ISO 26262, IEC 61508 deviation and waiver
machinery), security & vulnerability management (risk-acceptance records with
expiry), records/archival, data modeller, DX engineer, adversarial reviewer.

**External precedent weighed:** flaky-test quarantine decay (Google, Chromium and
WebKit `TestExpectations`, Android CTS waivers), Rust `#[ignore]` and
`//@ known-bug`, **LLVM `lit` `XFAIL`/`REQUIRES:` and XPASS-is-a-failure**,
`nolintlint`, PHPStan `reportUnmatchedIgnoredErrors`, security risk-acceptance
with mandatory expiry, TODO-with-deadline linters, ADR supersession lifecycles,
feature-flag staleness detection.

| Option | Description | Selected |
|--------|-------------|----------|
| A. Witness-first — extend `*-DEBT.md` with `Grade` + `Witness`; `.planning/UNREACHABLE-CLAIMS.md` is a GENERATED, byte-compared view; every trigger is an executed probe with lit XPASS semantics | One authored law; closes the `Landing phase` vocabulary to `P<NN>` / `CLOSED(sha)` / `UNOWNED(witness)`; the D-13-02b / D-13-10a shared root cause becomes one probe that turns both rows red at Phase 17; enumerator scans `_test.go` files and string literals. | ✓ |
| B. New authored top-level register with its own well-formedness test | Literal reading of EVD-03; no generator; milestone granularity matches a milestone-scoped concern. | |

**User's choice:** A — Witness-first: extend DEBT.md, generate the view.

**Notes:**
- **The decisive argument:** a trigger written as a grep over `check.go:255` is
  the four-times-fired defect in a fifth costume. Refactor or rename `sameType`
  and the grep silently stops matching, going green for the wrong reason —
  precisely instance 4's escape, where the guard watched the production file
  while the stale marker lived in the test file and an error string.
- **The working precedent is LLVM `lit`'s `XFAIL`:** the expected-to-fail test is
  *actually executed*, and an unexpected pass is reported as **XPASS and fails
  the suite** — verbatim EVD-04's "a guard fails when a cited gate has already
  closed." The same inverted-assertion property is what makes `nolintlint`'s
  `allow-unused: false` and PHPStan's `reportUnmatchedIgnoredErrors` work, and
  its absence is why Chromium/WebKit expectations files rot into thousands of
  stale lines.
- **DO-178C supplied the right frame for D-13-02b:** B1's blame branch is
  **deactivated code**, not dead code — present, provably unexecutable in the
  current configuration, justified by analysis that must be re-examined when the
  configuration changes. A probe forces that re-examination; a phase-number
  citation cannot.
- B rejected: ships a second coexisting law without retiring the first, puts each
  claim in two places with nothing forcing agreement, and makes the grade
  vocabulary a third source of truth — inside the phase whose purpose is retiring
  duplicated laws.
- **The generated view does not violate "a regenerable expectation is not
  evidence"**: the probe is the evidence; the view is a report, and
  regenerate-and-compare only prevents hand-edit drift. One honest hole noted: an
  empty generator and an empty checked-in file compare equal vacuously, so a
  non-zero row count is asserted whenever any register holds a qualifying row.
- **`phase:P17` is permitted only in `Landing phase`, never as a witness** —
  "Phase 17 lands" is decided by a human editing `ROADMAP.md`, which is an
  artifact read, which is the defect. Ownership and mechanics are separated.
- **Anti-grep rule:** for `callsite:`, a missing symbol is a **FAIL, not a zero**.
  Absence is never a pass.
- **The `Landing phase` non-empty hole is closed:**
  `"OPEN and UNOWNED — reopens only when …"` currently passes *because it is
  non-empty*. Under the closed vocabulary, `UNOWNED` must name a witness that
  resolves to an executed probe.
- **EVD-04's enumerator includes the two surfaces instance 4 missed:** it parses
  `_test.go` files and scans every string `BasicLit`, which is what catches
  `session_phase5_alias.go:523`'s error string and the prose in the two test
  files. The stale `PENDING-05-08` marker is **still live today**.
- **The old law is deleted in the same phase** (`TestNoNAT03RowRemainsPending`,
  the marker, the `t.Skip`, the stale prose) — shipping the new law while the old
  survives is the delayed fuse.
- **Non-inertness needs three seeded faults, not one** — one per mechanizable
  witness kind. A guard proven red on only one is inert for the other two.
- **The productive disagreement case:** if a row says `WIRED` but its probe fires,
  the suite goes red and a human must regrade. Grading is authored; **ungrading
  is forced by execution**.
- **Register hygiene adopted:** `first-recorded: <milestone>` plus a guard that a
  row carried across two milestone boundaries requires a `gate="blocking-human"`
  ratification — mechanizing "D-03-02 was carried past M001; carrying it past
  M002 would be a pattern" and the cluster-A base rate of 0 for 2.
- **Accepted residual friction:** one written probe per claim is real cost, and a
  probe can over-fit to current diagnostic text. Mitigation is to assert on the
  structural refusal rather than the message string. The runtime cost is the
  **anti-graveyard feature** — the inverse of expectations files, where a stale
  line costs nothing.

---

## Area 4 — Distinctness fix depth (DX-08/09)

**Gray area:** How far into the parser/diagnostic layer does the `if`-spiral fix
go, given that `surface.not_in_language` is deferred by C5 and published IDs are
pinned? What is the ID-churn budget and does a schema bump apply? Gate first or
fix first? What certifies a corpus of *parse-failing* programs as structurally
distinct? What exactly goes in `diagnosis`?

**Role lenses fanned out:** compiler front-end / parser-error-recovery
specialist, language-server/IDE tooling, API versioning & schema-evolution
architect, caching/determinism/content-addressing architect (the recorded
counter-party, whose side was taken seriously), AI-agent-tooling engineer, DX /
error-message researcher, QA/test-architect, security (cache poisoning / evidence
confusion), SRE/release, adversarial reviewer.

**External precedent weighed:** parser error-recovery literature (panic-mode,
phrase-level, Burke-Fisher) and how rustc/clang/**Roslyn**/Swift/Elm recover from
an unknown leading keyword; **Roslyn `SkippedTokensTrivia`** and Roslyn issue
#30749; clang diagnostic-stability policy; **SARIF `fingerprints` /
`partialFingerprints`**; content-addressed build systems and coarse action-key
hazards (Bazel/Nix); versioned error-code registries (Rust `E0XXX`, TS `TSxxxx`);
APR convergence-signal literature; Goodhart on ratio metrics; RFC 7807/9457
Problem Details as the analogue for the `diagnosis` shape question.

**Method note:** this researcher reproduced the spiral against the working tree,
prototyped **two** candidate fixes, ran the full `go test ./...` suite against
each, and reverted (`git status` clean). The figures below are measurements.

| Option | Description | Selected |
|--------|-------------|----------|
| A. `skipped_region` cause; fix + gate in one plan, with a frozen pre-fix capture as the must-fail non-inertness control | Attach the byte extent discarded by declaration-level recovery as an identity-bearing `Cause` on `syntax.expected_declaration`. ~14 lines, one call site. Verified to separate all three programs; verified zero in-repo test churn; no new code, no new field, no schema bump, no contact with the deferred `surface.not_in_language`. DX-09 adds three `omitempty` fields plus a closed `decline_reason` vocabulary. | ✓ |
| B. Pin the spiral as a failing control + a recorded trigger; defer the fix | Zero churn, zero blast radius in a phase whose other four criteria are markdown instruments; honest that `if` is not in the language. | |

**User's choice:** A — `skipped_region` cause; fix + gate in one plan.

**Notes:**
- **A correction to the research's own framing, carried into CONTEXT.md:** "the
  first diagnostic is one token late" is **not a defect**. `if` is absent from
  `keywords`, so it lexes as an identifier and is consumed as a valid linear
  result; "expected `}`, found `value`" is correct. Re-pointing that span at `if`
  would require the parser to know `if` is special — which *is* the deferred
  `surface.not_in_language` — and it **does not separate the three programs**,
  since all three share `if` at the same offset. **Span relocation is not the fix.**
- **The real mechanism:** `parseProgram`'s `default` arm reports one token, then
  `recoverUntil` discards everything to EOF **and reports nothing about it** —
  throwing away the one datum that differs between the three programs. The fix is
  Roslyn's `SkippedTokensTrivia` brought into the diagnostic instead of the AST.
- **Measured separation** (26 / 11 / 1 skipped bytes ⇒ three distinct
  `causes[0].span` values ⇒ three distinct diagnostic IDs and three distinct
  result IDs), with `primary_span` staying one token wide in all three so the
  rendering story does not regress.
- **The crude variant was also prototyped** (widen `primary_span` over the skipped
  region). It also separates all three and also passes the suite, but regresses
  rendering and leaves the third program's ID unchanged only by accident.
- **No schema bump, on principle rather than convenience:** the identity basis is
  `{Schema, Code, Span, Causes}` and `Causes` is **already** in it. This changes
  identity *content*, not the *basis*; the `/0`→`/1` bump happened only because
  the basis changed.
- **Churn measured at zero in-repo** — full suite green unmodified;
  `syntax.expected_declaration` appears in no pinned artifact. But zero is a
  measurement, not a guarantee, so D-13-09a's discipline still applies as an
  explicit up-front enumeration task. **Budget stated precisely: broken-program
  diagnostic IDs may move; no valid program's evidence moves.**
- **The "negative control" ambiguity resolved by splitting the role:** the live
  trio is a **positive regression pin (must PASS)** satisfying ROADMAP criterion
  5; a **frozen pre-fix capture** is the **non-inertness control (must FAIL)**,
  asserting the metric reports 1/3 over it. The capture is a historical artifact
  that must never be regenerated, and the test asserts a property over it rather
  than a golden match — so it is not a `--bless` button.
- **Corpus predicate for parse-failing programs:** non-trivia **token-kind
  sequence** inequality, computed by the lexer alone. Chosen because it is total
  and cheap, **renaming-immune by construction** (every identifier lexes to the
  same kind, so the corpus cannot be padded with renames), and **independent of
  the instrument under test** — a corpus certified by the parser it measures would
  be measuring its own assumptions. D-13-33's structural summary and D-13-26's
  topology triple both degenerate here, since these programs never reach the
  checker.
- **Anti-Goodhart on distinctness = 1.0:** the predicate blocks trivial padding;
  corpus membership is **additive-only**, so removing a member to go green is a
  reviewable deletion requiring a recorded decision; and the frozen collision
  control fails immediately if the metric is stubbed. **What does not stop it:**
  nothing forces the corpus to grow as the language grows — recorded as debt with
  an explicit trigger.
- **DX-09:** do not change `diagnosis_code`'s type (a protocol break for no
  gain); add three `omitempty` siblings on the non-identity-bearing `Outcome`
  envelope, with `decline_reason` drawn from a **closed coded vocabulary**. The
  bug is exactly that `selectRepair` returned `""` and discarded information it
  already had. Honest declines (forward-direction loan liveness, call-graph
  cycles, the uniqueness gate) stay `unrepairable` and now carry their reason — so
  "a cycle has no local, mechanical edit" is stated in the protocol rather than
  left as silence.
- **The prose-scramble boundary is the load-bearing DX-09 check and it is clear:**
  every emitted value is a code from a closed vocabulary, no new prose field is
  decoded, so lorem-ipsum scrambling changes nothing in the output and
  `TestRepairDriverDecodesNoProseFields` passes unmodified. The import boundary is
  likewise untouched.
- **B was argued fairly and would have won on different facts.** `if` is not in
  the language and will not be for milestones; the work is foreign to its host
  phase; the spiral is a hand-simulated proxy, not a measured agent trajectory.
  The researcher stated plainly that it would recommend B **if the fix had cost a
  schema bump or churned the pinned rows**. It cost neither.
- **The decisive counter is a security/evidence argument:** three distinct sources
  currently mint **one `result:` ID**, so an evidence cache keyed on the result ID
  serves one program's evidence for another. That is evidence confusion in a
  product whose product is evidence — the same class as Bazel/Nix action-key
  under-specification, and it generalizes to any future coarse-identity site.
- **Two debts recorded rather than rediscovered:** more content on `Cause.Span`
  deepens the already-recorded half-enforcement of coordinate-shift invariance
  (for broken programs only); and when `if` enters the language the spiral trio
  becomes a *valid* program and must be replaced in the corpus.
- **Sequencing vs spike S-009:** independent, do not block. Both DX items strictly
  *increase* the protocol-only signal S-009 measures. If S-009 reports recovery
  already > 80%, that should suppress appetite for a third, wider DX item this
  milestone — it should not touch these two.

---

## Follow-up 1 — Register topology (cross-area)

**Why asked:** Areas 2 and 3 collided. Area 2 proposed a new
`.planning/EVIDENCE-RECONCILIATION.md` ledger; area 3 argued hard against minting
second registers. Separately, area 1 retires the `Status` column that area 2's
Tier-A role detection leans on. Neither researcher could see the other's output.

| Option | Description | Selected |
|--------|-------------|----------|
| A. Two authored laws + two generated views | One authored law per subject — `*-VALIDATION.md` rows and `*-DEBT.md` rows — both enforced by tests in the same file as `checkDebtRegister`, sharing `phaseArtifactGlob`. Both `UNREACHABLE-CLAIMS.md` and `EVIDENCE-RECONCILIATION.md` become generated, byte-compared views, so a dead archived command becomes a debt row with a witness rather than a third authored register. Tier-A detection keys on the canonical 10-column header and the `Grade` column rather than the retired `Status` token. | ✓ |
| B. Three authored laws, kept separate | Keep the reconciliation ledger authored in its own right, on the grounds that a dead command and an unreachable claim are genuinely different subjects. Simpler to write; three well-formedness tests to keep in sync. | |

**User's choice:** A — Two authored laws + two generated views.

**Notes:** Resolves the tension in the direction area 3 argued for while keeping
area 2's ledger *content* intact — the ledger's authorship moves into the
existing law, its executable obligations survive unchanged. Also closes the
ordering hazard created by retiring `Status`: both changes land in the same phase
and must be sequenced so Tier-A detection is never briefly blind.

---

## Follow-up 2 — Enforcement scope in Phase 14 vs QLT-10

**Why asked:** The lint researcher flagged per-branch alternation checking (R2b)
as roughly 4-5× the dead-pattern class and warned explicitly that "deciding this
silently is how the phase blows its size."

| Option | Description | Selected |
|--------|-------------|----------|
| A. Detect all, enforce R1+R2+R3 now | Ship per-branch checking in Phase 14 so the lint *does* the finding (~8 more rows in `09-VALIDATION.md`, 3 in `11-VALIDATION.md`), pin those under-scoped findings in the frontier literal, but enforce-to-zero only the dead-pattern, elision and grep classes this phase. Under-scoped closure lands with QLT-10 in Phase 20. Keeps Phase 14 at its 6-8 plan sizing. | ✓ |
| B. Enforce everything in Phase 14 | Drive the under-scoped class to zero now. ~4-5× the workload, and re-adjudicating what a row's command was meant to prove is judgment work, not mechanical. | |
| C. Defer under-scoped detection entirely | Ship only R1+R2+R3; no per-branch checking this milestone. Smallest phase, but leaves the largest defect class undetected and unrecorded. | |

**User's choice:** A — Detect all, enforce R1+R2+R3 now.

**Notes:** C was the sizing-optimal choice and was rejected because leaving the
largest defect class *undetected and unrecorded* is the failure mode the phase
exists to retire — detection is cheap, adjudication is not, so they are split
across the milestone rather than traded away.

---

## Claude's Discretion

- Exact column ordering in the `*-VALIDATION.md` / `*-DEBT.md` tables, and whether
  `Non-inertness` and `Witness` are one column or two per file type.
- How the `go test -json` run record is produced, provided the gate fails closed
  when a row declares `EXERCISED`+ with no run record present (and that
  fail-closed behavior itself carries a non-inertness proof).
- Where the two generated views physically live, provided the paths EVD-03 names
  are honored.
- Whether lint error messages compute nearest-live-test-names by edit distance
  (recommended as the DX budget).
- Whether EVD-06's `LANGUAGE-MATURITY.md` self-check executes the file's own
  embedded re-verify greps or re-derives the counts independently
  (peer-re-derivation discipline leans toward the latter).
- Whether EVD-08's `observed` wall-clock row lands in the existing budget manifest
  or a sibling, provided the 192.7s baseline is recorded and the ~60s claim is
  corrected wherever it appears.
- Plan count and wave shaping within the 6-8 sizing, and whether the two DX items
  are one plan or two.

## Deferred Ideas

- `surface.not_in_language` and the `not_in_language` capability manifest —
  deferred by C5; DX-08 must be satisfied without them.
- Under-scoped verification rows (R2b) driven to zero — QLT-10, Phase 20.
- D-13-34's held-out corpus fix — record the trigger only; attaches to the phase
  that adds arithmetic or control flow.
- Distinctness corpus growth as the language grows — when `if` lands, the spiral
  trio becomes a valid program and must be replaced.
- `Cause.Span` identity-bearing while `Repair.Span` is not — carried as debt with
  a witness; this phase bounded-widens the hole for broken programs only.
- `check.go:3123`'s `declare_foreign_symbol` shell repair — carried from Phase 13;
  a future phase must complete it or delete it.
- Cluster A, the six single-function emitters — Phase 16's business (NAT-08/09).
- Cluster D, the D-09-51 verdict-flip review — a blocking-human checkpoint in
  someone's plan, not a Phase 14 deliverable.
- QLT-12, content-addressing the enumerated-closure proof — Phase 20; relevant
  here only as the escape hatch if the 3.35s accuracy control comes under
  wall-clock pressure.
- A measured agent-trajectory study (20 trials with a real model) to replace the
  hand-simulated spiral proxy — named as the single most valuable follow-up
  experiment; not scheduled. S-009 is the cheap partial substitute.
- Making Nyquist a hard gate rather than a skill (research F4) — written down
  twice across two milestones and never implemented; not in EVD-01..08. QLT-10's
  "no VALIDATION file remains `draft`" is the M003-scoped slice.
- Demoting the integration checker by contract (research F5) — follows naturally
  from EVD-02 but is a GSD-process change, not a repo change.
