---
phase: 14-evidence-instrument-and-honest-scoping
plan: 01
subsystem: evidence-instruments
tags: [go, testing, go-parser, groundedness-lint, static-analysis, evd-01]

requires:
  - phase: 13-agent-loop-for-interprocedural-defects
    provides: the archived M001/M002 evidence corpus (*-VALIDATION.md, *-VERIFICATION.md) this lint scans
provides:
  - internal/compiler/session/verification_groundedness_test.go — the groundedness lint (EVD-01), a static Go test index, a Tier-A document scanner, R1/R2/unparseable classification, a pinned exact violation frontier, and a three-fault non-inertness proof
affects: ["14-06 (R2b alternation-branch enforcement and archive reconciliation verdicts)", "14-10 (Nyquist reconciliation empties the frontier pin)", "any later plan in this milestone that touches *-VALIDATION.md or *-VERIFICATION.md content"]

actuals:
  tokens: 9074
  tasks: 3
  commits: 4

tech-stack:
  added: []
  patterns:
    - "Go-test-hosted document lint over .planning/** using go/parser + go/build.Context.MatchFile for a static, zero-shell-out test index (D-14-10)"
    - "phaseArtifactGlob-only discovery, never filepath.Glob directly"
    - "Exact SET EQUALITY frontier pin (never a count/ceiling) with a committed literal, re-measured at authoring time rather than copied from prose"
    - "Temp-copy-and-seed-one-fault non-inertness proof, one subtest per mechanizable classification"

key-files:
  created:
    - internal/compiler/session/verification_groundedness_test.go
  modified: []

key-decisions:
  - "Hosted new file in package session_test (not the plan's literally stated 'package session') because phaseArtifactGlob, the required discovery entry point, is unexported and defined in session_test.go's session_test package; a file declared package session cannot see it. Documented as a Rule 3 auto-fix (blocking compile-boundary issue), not an architectural change."
  - "Fixed a real backtick-awareness bug in the table-row splitter found via testing against the live corpus: the naive split-on-unescaped-pipe algorithm from D-14-15's literal description truncated commands whose alternation pattern uses a raw, unescaped '|' inside backticks (valid, rendering GFM — M001's 01-VALIDATION.md and 04-VERIFICATION.md use this form; 09-VALIDATION.md's Mode.*Invalid row uses the backslash-escaped form instead). The splitter now tracks backtick nesting and never splits inside a code span, tolerating both conventions."
  - "Fixed a second real bug found the same way: classifyCommand used strings.Contains(command, \"go test\") to detect go-test cells, which false-positived on \"session.go testdata/...\" (the literal substring \"go test\" occurs inside \".go test[data]\"), misrouting a git diff command into the go-test parser as a spurious 'unparseable' violation. Switched to strings.HasPrefix since the command has already matched the anchored verificationCommandPattern."
  - "A bare 'go test' cell with no package operand (doc-prose mentions of the test framework, e.g. a VALIDATION.md Framework row) classifies unparseable rather than ok, per D-14-15's no-silent-skip rule — 6 of the 26 pinned entries are this shape."
  - "A go test cell with package operands but no -run/-list/-fuzz/-bench flag classifies ok (parses successfully, pattern is empty, no groundedness check applies) rather than unparseable — this is the common 'go test ./...' full-suite invocation, which is legitimate and not a groundedness question."
  - "Task 3's non-inertness proof base fixture is 09-VERIFICATION.md, confirmed to produce zero violations against the live tree, so a seeded fault's single new violation is unambiguously attributable to the seed."
  - "Task 3's git-status-unchanged assertion compares before/after .planning porcelain status rather than asserting an empty tree, because STATE.md/state.json are legitimately touched by concurrent GSD tracking machinery during phase execution — asserting empty would be a false positive in this repo's normal working state."

patterns-established:
  - "Any future .planning/**-scanning Go test in this repo should reuse phaseArtifactGlob and the backtick-aware table-row splitter established here rather than re-deriving pipe-splitting from scratch."

requirements-completed: [EVD-01]

coverage:
  - id: D1
    description: "Groundedness lint discovers Tier-A evidence documents, extracts verification commands per code span, and classifies each into exactly one of {ok, R1, R2, unparseable} with no silent skips"
    requirement: "EVD-01"
    verification:
      - kind: unit
        ref: "internal/compiler/session/verification_groundedness_test.go#TestVerificationGroundednessClassifier"
        status: pass
      - kind: integration
        ref: "internal/compiler/session/verification_groundedness_test.go#TestVerificationGroundedness"
        status: pass
    human_judgment: false
  - id: D2
    description: "The current violation set is pinned as an exact committed literal (set equality, not count/ceiling), proven bidirectionally mutable-to-red by manual verification during this task"
    requirement: "EVD-01"
    verification:
      - kind: unit
        ref: "internal/compiler/session/verification_groundedness_test.go#TestVerificationGroundednessFrontierIsPinned"
        status: pass
    human_judgment: false
  - id: D3
    description: "The lint is proven non-inert on all three of its mechanizable violation classes (R1, R2, unparseable) via seeded faults on a temp copy"
    requirement: "EVD-01"
    verification:
      - kind: unit
        ref: "internal/compiler/session/verification_groundedness_test.go#TestVerificationGroundednessIsNotInert"
        status: pass
    human_judgment: false

duration: 15min (first commit to last commit; excludes read-only context-gathering)
completed: 2026-09-17
status: complete
---

# Phase 14 Plan 01: Groundedness Lint Summary

**A Go test (`internal/compiler/session/verification_groundedness_test.go`) that statically resolves every `go test -run` pattern cited across 27 archived and live evidence documents against a `go/parser`-built test index, classifies unrunnable and dead commands, and pins the measured 26-violation frontier as an exact, bidirectionally-provable literal.**

## Performance

- **Duration:** 15 min (commit-to-commit span; RED at 20:34:59, final GREEN at 20:49:33)
- **Tasks:** 3
- **Files created:** 1 (924 lines)

## Accomplishments

- Static test index: walks the module from `testsupport.ProjectPath()`, parses every `*_test.go` with `go/parser.ParseFile(..., parser.SkipObjectResolution)`, honors build constraints via `go/build.Context.MatchFile`, and resolves `Test`/`Fuzz`/`Benchmark`/`Example` identifiers by exact match only.
- Tier-A document scanner: discovers evidence documents exclusively through `phaseArtifactGlob("*", "*-VALIDATION.md")` / `("*", "*-VERIFICATION.md")`, parses table rows with a backtick-aware, escape-tolerant splitter, and extracts every backtick-delimited command span per cell (a row may legitimately carry two commands).
- R1 (runnability) and R2 (groundedness) classification, plus a hard `unparseable` classification for any `go test` cell that fails to parse into package operands — never a silent pass.
- `TestVerificationGroundednessFrontierIsPinned` pins the measured 26-record violation set as an exact literal, asserted for SET EQUALITY (both directions) against a fresh run — verified manually during Task 2 to fail correctly on both a deletion and a fabricated addition.
- `TestVerificationGroundednessIsNotInert` proves the lint is not inert on all three mechanizable classifications via three independently seeded faults on a temp copy of a real, currently-clean document (`09-VERIFICATION.md`), leaving the real `.planning/**` tree untouched (verified via before/after `git status --porcelain`).
- Found and fixed two real, previously-invisible bugs in the classifier while testing against the live corpus (see Deviations).

## Task Commits

1. **Task 1: End-to-end groundedness lint over one Tier-A document class** — `93613cc` (test, RED) → `f71aba8` (feat, GREEN)
2. **Task 2: Pin the measured violation frontier as an exact committed literal** — `3ab5396` (feat)
3. **Task 3: Prove the lint is not inert with a seeded dead pattern** — `9d10d20` (test)

_Note: Task 1 is `tdd="true"`: `93613cc` is the intentional RED (classifier stubbed to always return `ok`, confirmed failing on assertions not compile errors), `f71aba8` is GREEN (real classification wired in, plus the backtick-splitter and HasPrefix fixes discovered along the way)._

## Files Created/Modified

- `internal/compiler/session/verification_groundedness_test.go` — the groundedness lint: static test index, Tier-A scanner, R1/R2/unparseable classifier, frontier pin, non-inertness proof.

## Decisions Made

See `key-decisions` in frontmatter. Summary: (1) hosted in `package session_test` rather than the plan's literal `package session` text, because `phaseArtifactGlob` is unexported in `session_test`'s external test package and a same-directory `package session` file cannot see it — a Rule 3 blocking-issue auto-fix, not an architectural change; (2) fixed a backtick-awareness bug in table-row splitting; (3) fixed a `Contains`-vs-`HasPrefix` substring bug in go-test-cell detection; (4) a bare `go test` prose mention classifies `unparseable` (no package operand); (5) a package-only invocation with no `-run`/`-list`/`-fuzz`/`-bench` classifies `ok` (nothing to check groundedness of).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] File hosted in `package session_test`, not the plan's literal `package session`**
- **Found during:** Task 1, before writing any test code
- **Issue:** The plan's action text says "Create ... in package `session`", but the required discovery entry point `phaseArtifactGlob` (and `checkDebtRegister`, the Pattern-1 analog cited in `14-PATTERNS.md`) is an unexported symbol defined in `session_test.go`, which declares `package session_test` (Go's external test package), not `package session`. Every other file in this package that calls `phaseArtifactGlob` (`session_peer_gate_test.go`, `session_payload_replay_test.go`) is also `package session_test`. A file declared `package session` cannot see an unexported symbol from a different package, even one compiled into the same test binary.
- **Fix:** Declared `verification_groundedness_test.go` as `package session_test`.
- **Files modified:** `internal/compiler/session/verification_groundedness_test.go`
- **Verification:** `grep -c 'filepath.Glob' internal/compiler/session/verification_groundedness_test.go` → `0`; `go build ./...` and `go vet ./internal/compiler/session/...` both clean; `phaseArtifactGlob` call sites present.
- **Committed in:** `93613cc`

**2. [Rule 1 - Bug] Table-row splitter was not backtick-aware, truncating commands with an unescaped `|` inside code spans**
- **Found during:** Task 1, first real-corpus run (RED→GREEN transition)
- **Issue:** D-14-15's literal description ("split on `|` not preceded by `\`, then unescape") assumes every cell that needs a literal pipe inside a code span escapes it with `\|`. The real corpus is inconsistent: `01-VALIDATION.md:36`, `04-VERIFICATION.md:86/88/89`, and others use a raw, unescaped `|` inside backticks (valid, code-span-aware GFM rendering — GitHub does not require escaping `|` inside inline code), while `09-VALIDATION.md:93` uses the backslash-escaped form. The naive splitter split mid-command on the unescaped form, producing truncated, garbled command text and false `unparseable` classifications for at least 8 real rows.
- **Fix:** `splitTableRow` now tracks whether it is inside a backtick span and never splits on `|` while inside one, regardless of escaping — tolerating both conventions actually present in the corpus.
- **Files modified:** `internal/compiler/session/verification_groundedness_test.go`
- **Verification:** Re-ran `TestVerificationGroundedness -v`; the 8 previously-truncated entries disappeared from the violation log and the three required dead patterns (08:51, 08:63, 09:85) remained correctly reported.
- **Committed in:** `f71aba8`

**3. [Rule 1 - Bug] `strings.Contains(command, "go test")` false-positived inside "session.go testdata/..."**
- **Found during:** Task 1/2 boundary, while transcribing the measured frontier for the Task 2 pin
- **Issue:** `05-VERIFICATION.md:51`'s command is a `git diff --stat ... testdata/phase1..4 ...` invocation. `strings.Contains` matched the literal substring `"go test"` inside `".go test[data]"` (i.e., `session.go testdata`), routing a `git diff` command into the go-test parser, which correctly failed to parse it into `(packages, pattern)` and reported a spurious `unparseable` violation.
- **Fix:** Changed the discriminator from `strings.Contains` to `strings.HasPrefix`, since `command` has already matched `verificationCommandPattern`'s anchored-at-start regex — `HasPrefix` correctly identifies which prefix matched without the substring pitfall.
- **Files modified:** `internal/compiler/session/verification_groundedness_test.go`
- **Verification:** Re-ran `TestVerificationGroundedness -v`; the false `05-VERIFICATION.md:51` entry disappeared; violation count dropped from 27 to the final pinned 26.
- **Committed in:** `3ab5396`

---

**Total deviations:** 3 auto-fixed (1 blocking package-boundary fix, 2 bugs found via testing against the real corpus).
**Impact on plan:** All three were necessary for the lint to correctly classify the real `.planning/**` corpus rather than silently misreporting it — exactly the failure mode (an instrument that reports wrong findings because it was never run against real data) this phase exists to retire. No scope creep; no architectural change.

## Issues Encountered

None beyond the deviations above.

## Measured Violation Frontier (verbatim, for plan 14-10's reconciliation)

26 records, `(file, line, classification, command)`, sorted by `(file, line)`:

| File | Line | Class | Command |
|---|---|---|---|
| `.planning/milestones/M001-phases/05-native-equivalence-and-adversarial-evidence/05-VALIDATION.md` | 22 | unparseable | `go test` |
| `.planning/milestones/M001-phases/06-agent-feedback-and-performance-ratification/06-VALIDATION.md` | 23 | unparseable | `go test` |
| `.planning/milestones/M002-phases/07-calls-signatures-and-call-graph-refusal/07-VERIFICATION.md` | 87 | R1 | `grep -nE 'TBD\|FIXME\|XXX'` |
| `.planning/milestones/M002-phases/08-interprocedural-loan-liveness-in-check/08-VALIDATION.md` | 24 | unparseable | `go test` |
| `.planning/milestones/M002-phases/08-interprocedural-loan-liveness-in-check/08-VALIDATION.md` | 51 | R2 | `go test ./internal/compiler/check/... -run TestComputeLoanLastUsesAndDerivePlaceLoansAgree` |
| `.planning/milestones/M002-phases/08-interprocedural-loan-liveness-in-check/08-VALIDATION.md` | 63 | R2 | `go test ./internal/compiler/session/... -run TestAuditQLT02BudgetManifest` |
| `.planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md` | 44 | unparseable | `go test` |
| `.planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md` | 85 | R2 | `go test ./internal/compiler/corevalidate -run 'LoanChainIndex' -v` |
| `.planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md` | 93 | R2 | `go test ./internal/compiler/corevalidate -run 'Mode.*Invalid\|DecodeMode' -v` |
| `.planning/milestones/M002-phases/10-trusted-interprocedural-oracle/10-VALIDATION.md` | 25 | unparseable | `go test` |
| `.planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md` | 24 | unparseable | `go test` |
| `.planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md` | 26 | R1 | `go test ./internal/compiler/<touched-package>/...` |
| `.planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md` | 51 | R1 | `grep -c -E 'D-11-(02\|07\|11\|12\|13\|27\|36\|40\|42)' …/PHASE-11-DEBT.md` |
| `.planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md` | 56 | R1 | `go test ./internal/compiler/callgraph/... -run 'TestEntryFunction…' -v -count=1` |
| `.planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md` | 58 | R1 | `go test ./internal/compiler/cgen/... -run 'TestEmittedAttributeSet…' -v -count=1` |
| `.planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md` | 60 | R1 | `grep -c -E 'function count\|call-edge count\|N =…' …/11-MIDPHASE-GATE.md` |
| `.planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md` | 61 | R1 | `awk … \| wc -l` |
| `.planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md` | 63 | R1 | `grep -c 'D-11-25' …/session_phase11_differential_test.go` |
| `.planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md` | 64 | R1 | `go test ./internal/compiler/session/... -run 'TestQLT03GeneratorOpKindClosure…' -v -count=1` |
| `.planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md` | 68 | R1 | `go test ./internal/compiler/cache/... -run 'TestDeclaredInputNames\|TestCache…\|TestNoClosureDigestInCache' -v -count=1` |
| `.planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md` | 69 | R1 | `grep -c -E 'QLT-06a\|QLT-06b\|strictly dominates…' …/11-QLT06-ABSTENTION.md` |
| `.planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md` | 71 | R1 | `go test ./internal/compiler/reduce/... -run 'TestDropCallSite\|TestDropOrphanFunction\|…' -v -count=1` |
| `.planning/milestones/M002-phases/12-result-payloads/12-VALIDATION.md` | 22 | unparseable | `go test` |
| `.planning/milestones/M002-phases/12-result-payloads/12-VALIDATION.md` | 24 | R1 | `go test ./internal/compiler/<package>/... -run <TestName> -count=1` |
| `.planning/milestones/M002-phases/13-agent-loop-for-interprocedural-defects/13-VALIDATION.md` | 23 | unparseable | `go test` |
| `.planning/phases/14-evidence-instrument-and-honest-scoping/14-VALIDATION.md` | 24 | R1 | `go test ./<changed-package>/...` |

**Note:** this frontier differs from the ROADMAP's "three" floor and from `14-CONTEXT.md` D-14-19's prediction, per that decision's own instruction to measure rather than assume. It also differs from an interim 27-entry measurement taken mid-Task-1 (before the `Contains`→`HasPrefix` fix removed one false positive at `05-VERIFICATION.md:51`) — the 26-entry set above is the one actually pinned in the committed literal.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- The groundedness lint is live, green, and pinned. Plan 14-06 (R2b alternation-branch enforcement, archive reconciliation verdicts `renamed`/`superseded`/`obsolete-by-design`/`under-scoped`) will extend this same file rather than create a new one.
- Plan 14-10 (Nyquist reconciliation) must reconcile every one of the 26 pinned entries above and drive the R1/R2/unparseable portion of the frontier to empty, per D-14-18.
- No blockers for the next plan in this phase.

---
*Phase: 14-evidence-instrument-and-honest-scoping*
*Completed: 2026-09-17*

## Self-Check: PASSED

- FOUND: internal/compiler/session/verification_groundedness_test.go
- FOUND: commit 93613cc (test, RED)
- FOUND: commit f71aba8 (feat, GREEN, Task 1)
- FOUND: commit 3ab5396 (feat, Task 2)
- FOUND: commit 9d10d20 (test, Task 3)
- `go test ./internal/compiler/session/... -count=1` exits 0
- `go build ./...` exits 0, `go vet ./internal/compiler/session/...` silent
- `git diff --stat -- go.mod go.sum scripts/assert-go-tests.sh` prints nothing
