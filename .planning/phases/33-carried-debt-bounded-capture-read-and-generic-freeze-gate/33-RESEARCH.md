---
verification_role: proposal
---

# Phase 33: Carried Debt, Bounded Capture Read, and Generic Freeze Gate - Research

**Researched:** 2026-10-13
**Domain:** Internal tooling and test infrastructure (POSIX sh gate scripts, Go stdlib tests, one bounded file read, one interpreter predicate, planning-doc meta-tests). No external libraries.
**Confidence:** HIGH (everything below was read in this session from the repo; the only external fact, the `VALID_KINDS` enum, was read from the local gsd-core install)

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

#### A. M006 receipt reopen (D-10a), which comes first
- **D-01:** The first commit of the phase is a docs-only commit, Plan 1 Task 1, made before any digest-path commit. It flips both machine-checked markers, `.planning/LANGUAGE-MATURITY.md:239` and `.planning/PRODUCT-ROADMAP.md:10`, to `M006 final hosted receipts: reopened 2026-10-xx (historical codedigest b9253457705a224a0580d74896e8e5bd3821446a1d467626b7d611ca0d1c4b5c, run 38335175306)`. Use the actual date and match the format `phase32ReopenedMarker` accepts. Rewrite the prose claims of current evidence in PROJECT.md (~171, ~609-610), MILESTONES.md (~19-20) and STATE.md in the Phase 25 "historical receipts" form. Without this commit, `TestPhase32ReceiptLivingClaims` (`stale_bound_claim`) turns CI red on the first code commit. — **Reversibility:** reversible
- **D-02:** Archived receipt artifacts are immutable. Do not modify `32-RECEIPT.json`, `32-RECEIPT.run-38315799626.json` or `receipts/*.log`. The reopen is recorded only in the living docs.
- **D-03:** No hosted re-take in Phase 33. The docs use this wording: "Phase 33 binds no hosted receipt. M006 receipts are historical; current hosted evidence resumes in Phase 38."

#### B. Generic freeze/receipt gate (EVD-16, plus 32.1 WR-01 and IN-03)
- **D-04:** The gate is two POSIX sh scripts, `scripts/check-freeze.sh` and `scripts/check-receipts.sh`. Do not build a Go tool: it would give up byte parity and put a new binary in the digest. The roughly 10-line manifest reader is copy-pasted into both scripts, with no shared sh library. — **Reversibility:** costly — Phase 38 and later milestones call these names
- **D-05:** The per-milestone manifest is `testdata/receipts/<MILESTONE>.manifest`, starting with `M006.manifest`. It sits inside the digest paths, so a change to the case table moves the digest. It holds only facts known before the freeze:
  - `label=phase32`
  - `programs=…`
  - `controls=guide,route-audit,codedigest,receipt-rules` (this replaces the hard-coded `4`)
  - the `# case …` rows copied verbatim

  The **freeze commit and the expected codedigest never go in the manifest**, because committing them would trip the gate's own freeze check. They are CLI args or env:
  - `check-freeze.sh MANIFEST FREEZE CANDIDATE [SAME_AS…]`
  - `check-receipts.sh MANIFEST LOG_A LOG_B`, with `EXPECTED_CODEDIGEST` generalizing `PHASE32_EXPECTED_CODEDIGEST`. When it is set the digest is never recomputed. Without it, the gate recomputes from the checkout, and Phase 38 uses that default.

  Placement conflict, resolved: one researcher proposed keeping the M006 manifest under `.planning/.../receipts/`. The uniform `testdata/receipts/` location won. M006 provenance is pinned instead in two ways. First, a test asserts that the manifest's sha256 matches, and that every manifest case row appears in both archived logs. Second, the executor checks once, in the SUMMARY, that the rows equal `git show 9f8f878f:scripts/verify-phase32.sh`'s table. CI does not depend on deep git history.
- **D-06:** "Byte-for-byte" means that, for the M006 manifest, the new scripts match the old ones on all of these:
  - the exit code;
  - stdout (empty on failure);
  - the single stderr line `${label} receipt gate: status=fail reason=…` or `${label} freeze gate: …`, including the `_$log` suffix as the caller typed the path;
  - the pass line's field order, including `digest_source=env|computed`.

  The label is data. The manifest is read before the arg-count check, so `reason=usage` keeps the `phase32` prefix. Only `manifest_missing` and `no_repo_root` are new reasons, and they appear only where the old scripts had no equivalent.
- **D-07:** WR-01: both gates resolve `root=$(git rev-parse --show-toplevel 2>/dev/null) || reject no_repo_root; cd "$root"`. Before the `cd`, MANIFEST and the log args are made absolute for reading, but the original strings are kept for the reason bytes. A subdirectory test runs from `cmd/` with a digest path changed and must give `digest_path_changed_after_freeze`.
- **D-08:** IN-03: `scripts/codedigest.sh` becomes the single source of the digest path set and the `.md` filter. git rejects `:(exclude)` magic in `ls-tree`, so a shared pathspec cannot work. Add these flags:
  - `--rev REV` (default HEAD)
  - `--paths`
  - `--changed A B`, which prints `git log --format= --name-only A..B` together with `git diff --name-only A B` over the paths, with `\.md$` and blank lines removed and grep's exit 1 guarded

  The freeze gate rejects on non-empty `--changed` output. It keeps both the log and the diff, so a change that is reverted inside the range is still caught. The producer's tree-state uses `--paths`. Replace the hard-coded `64` with a length derived from the hasher's empty-input output. A Go test shows that an `.md`-only change moves neither the digest nor the freeze verdict. IN-03 is closed only if that test ships.
- **D-09:** `scripts/verify-phase32.sh` stays per-milestone and keeps its name. Its builds, fixtures and controls are bespoke. It switches to reading its case table and program list from `testdata/receipts/M006.manifest`, and its `# phase32-cases-begin/end` block and `phase32_programs=` line are deleted, which leaves one table. Phase 38 writes the M007 producer against `M007.manifest`.
- **D-10:** Migration happens in two steps.
  - **Step 1 (side by side):** add the manifest, the two scripts, and a Go differential test that runs every existing refusal and mutation case (freeze and receipt) through the old and the new script and asserts byte-equal exit, stdout and stderr. Add a test that the manifest equals the producer's table. Add `TestReceiptGateArchivedM006`: the archived pair plus the digest from `32-RECEIPT.json` must pass, and armed controls must reject it on each of: one byte of a log mutated, digest mismatch, a host dropped, a program missing, a malformed row, a dirty tree. Add the M006 freeze row: `0aac0a7b` to `9f8f878f` must accept, and `0aac0a7b` against a post-freeze code commit must reject.
  - **Step 2 (cut-over):** freeze the differential outputs into literal goldens, with the temp dir normalized. Delete `scripts/check-phase32-{freeze,receipts}.sh` and repoint the producer. Add `TestNoCheckPhase32Caller`, which walks tracked files outside `.planning/milestones/`, matches a needle built by concatenation, and also fails if either file still exists.
- **D-11:** Test files:
  - `session_phase32_freeze_test.go` becomes `session_freeze_gate_test.go` (`TestFreezeGate`).
  - `session_phase32_gate_test.go` becomes `session_receipt_gate_test.go` (`TestReceiptGate`, `TestReceiptGateArchivedM006`).
  - `session_phase32_receipt_test.go` and `session_phase32_ci_test.go` stay, because they pin M006 artifacts and the archived logs name their control tests.
  - `TestPhase32ProducerCaseTable` is retargeted to the manifest.

  Generalize the bound/reopened living-claim check by milestone label so Phase 38 can write `M007 final hosted receipts: bound to …`. Port every existing table case, including `digest-drift-alone-passes` and `reopened-marker-passes-when-digest-drifted`, before any deletion. Add one real-checkout row: archive receipt, reopened markers and a drifted digest must pass. Extend the living-claims doc set to PROJECT.md and MILESTONES.md. `ci.yml` has no `check-phase32-*` caller and needs no change.

#### C. Narrowing `isAbortLikeDefect` (32.1 WR-02)
- **D-12:** The acceptance rule is owner identity plus a closed per-kind reason. Signature: `isAbortLikeDefect(program, result, ownerOpID)`. The callers at interp.go ~319 and ~371 already find the single acquire op. It accepts only when all of these hold:
  - `result.LiveResources` is exactly `[ownerOpID]`, compared by ID, not by count;
  - the operation is looked up in the function named by `final.FunctionID`, not by a global scan;
  - `final.Output` equals the kind's reason. For `OpBytesAt`, `OpAddChecked` and `OpRemainderU64` that is a constant in `internal/compiler/execution`, for example `DefectReasonByteIndexOutOfBounds`, `DefectReasonU64AddOverflow`, `DefectReasonU64RemainderByZero`. For `OpDefect` it is `operation.Reason`.

  Use a `switch` on kind, not a set of strings. The interpreter, `executionpeer` and `cgen` all use these constants, which removes the three-way copy. The emitted C text must stay byte-identical, and the existing parity tests catch any difference.
- **D-13:** A defect with an unrelated non-owner live resource is **refused** with a hard error, even though native exits 65. This is the interpreter's leak-detection role: Miri-style guarding beside codegen. Route admission plus the ownership kernel should make that state unreachable. A test pins the refusal using a hand-built Execution with `[owner, "x:op:9"]` and an add-overflow defect. — **Reversibility:** reversible
- **D-14:** Order is pin first, then narrow, then mutant.
  1. Before changing the predicate, add explicit parity rows to `native_phase321_defect_parity_test.go`: exit 65, stderr, and the interpreter's `Output`, for all four kinds on both the fixture and file routes. `TestPhase321AbortLikeDefectPredicate` and `…FileRoute` must be green.
  2. Narrow the predicate. Accepted rows carry the owner ID and the matching reason. Add refusal rows for wrong reason, extra live resource, missing owner, and op in the wrong function, including a two-function falsifier with the same op ID.
  3. Add `TestPhase33AbortLikeWideningMutantKilled`. It uses unexported seam vars (`abortLikeAcceptLive`, `abortLikeReasonMatches`), swaps each one in a same-package test and restores it with `defer`, following the D-07-42 pattern (`decoderRequireNonEmpty`).

#### D. Bounded capture read (QLT-13)
- **D-15:** Extract `readApplicationCapture(path string, limit int) ([]byte, error)` in `internal/compiler/native/native_app.go`, modelled on `readApplicationReceipt` (~649). It does `os.Open`, then `io.ReadAll(io.LimitReader(f, int64(limit)+1))`, and `len > limit` gives `EvidenceStatusCapacityExhausted`, which ends as `native.evidence_capacity_exhausted`. There is **no Stat-by-path pre-check**: the child writes the file, so checking then reading by path is a TOCTOU race.
- **D-16:** `TestPhase33CaptureReadBounded` runs without `t.Parallel` and injects a small limit through `Runner.evidenceLimit`. It has three cases:
  - **exactly-limit:** a valid capture padded with JSON whitespace to exactly `limit` bytes must not be capacity exhausted;
  - **limit+1:** must be capacity exhausted;
  - **mutant catcher:** a sparse 1 GiB file (`Truncate(1<<30)`) must give the capacity outcome with a `runtime.ReadMemStats` `TotalAlloc` delta under `16*limit`. The behavior catches the `os.ReadFile` mutant, so a grep cannot be beaten by an alias.

  Run the mutant once by hand and record the failing output in the SUMMARY.

#### E. Small debt, fix or disposition (QLT-14 remainder, QLT-15)
- **D-17:** IN-01 is **covered, with no code change**. `TestPhase32PublicRouteAudit` (`cmd/schway/main_phase32_routes_test.go`) already fails on `uncovered_route` and `module_aware_set_drift`; the `frobnicate` and `set-drift` subtests seed it. The disposition names that test.
- **D-18:** IN-02 is **fixed**. Use an inline fixture `fn main(zq_param_marker: U64) -> U64` and assert that both `zq_param_marker` and `dir` are absent. A control assertion checks that the source does contain the marker, so the absence check cannot pass vacuously.
- **D-19:** The coverage-`kind` pin is **fixed**. Add `case "kind":` to `summaryCoverageProblems` in `internal/compiler/session/summary_coverage_test.go`, checking the closed enum `unit, integration, e2e, automated_ui, manual_procedural, other`. Copy the enum as a Go literal and add a comment citing `gsd-core/bin/lib/coverage.cjs:66`, with no dependency. Add a seeded `kind: manual` rejection case.
- **D-20:** Phase 29 WR-02/WR-03 are **verified and recorded**. They are already fixed in 32.1-03 and 32.1-04, and their witnesses `TestPhase321ByteViewEntryRefused` and `TestPhase321DefectParityWithLiveOwner` exist. Run both with one focused `go test -run` and record the result in the SUMMARY `coverage:` block. Do not conflate them with 32.1 WR-02.
- **D-21:** Phase 28 VALIDATION.md and SECURITY.md are **hand-authored from existing evidence**. Do not run `/gsd-validate-phase` or `/gsd-secure-phase`, because that would write new tests into frozen history and break the plan cap.
  - **VALIDATION.md:** copy the `29-VALIDATION.md` format, with `status: validated` (never `draft`) and a "Reconstructed after execution (State B)" comment. Every `go test -run` row must name a real test, because `TestVerificationGroundedness`, `TestValidationRowGradesAreEarnedOverArchivedCorpus` and `TestPhase20ValidationLifecycle` check this. Set `nyquist_compliant: false` if any requirement has no automated row. Mark gaps `partial` with a reason.
  - **SECURITY.md:** copy the `29-SECURITY.md` format. Build the threat register from 28-REVIEW, 28-REVIEW-DISPOSITION and 28-VERIFICATION, and state "retroactive, no new tests".
- **D-22:** Dispositions are recorded in three places:
  - a new `.planning/phases/33-…/33-CARRIED-DEBT-DISPOSITION.md`, one row per item: finding, fixed or covered, the witnessing test, and the plan;
  - minimal edits to the archived `32.1-REVIEW-DISPOSITION.md`, moving each item from `open` to `fixed` or `covered` with a pointer to 33, plus a one-line note that the receipt re-take precondition is superseded by D-01/D-03. No Go test parses REVIEW-DISPOSITION files;
  - updated "Remaining debt" lines in `.planning/MILESTONES.md` (~23-26) and `.planning/PROJECT.md` (~174).

### Claude's Discretion
- The exact reopened-marker date string and the prose wording, within D-01's form.
- The exact names of the defect-reason constants, as long as the emitted C stays byte-identical.
- The manifest's line syntax (key=value plus verbatim `# case` rows) and the temp-dir normalization token in the goldens.
- Whether the gate's `.md`-only Go test lives in the freeze-gate test file or a codedigest test.

### Deferred Ideas (OUT OF SCOPE)
- `check-phase27-receipts.sh` and other older phase-specific gates are out of scope; the static test covers only phase32. A later cleanup could migrate them to the generic gate.
- Moving the test seams for Phase 29/30 out of production packages (29 WR-04, listed in the M006 audit) is not in this phase.
- `evidence`/`interface`/`query` module-awareness stays a named refusal and is not in scope.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| QLT-13 | Native app reads the evidence capture through a limit+1 bounded reader; seeded unbounded-read mutant is caught | "Bounded capture read" below: exact current code (`native_app.go:364`), the model (`readApplicationReceipt`, `native_app.go:649-655`), the mutant-catcher shape that works end to end, memory-budget numbers |
| QLT-14 | 32.1 WR-01, WR-02, IN-01..03 and the coverage-`kind` pin closed with a test or a written disposition | WR-01 reproduced live; IN-03 hard-coded `64` and duplicated `.md` filter located; WR-02 predicate and all four defect-reason copies located; IN-01 witness confirmed; IN-02 fixture verified to exit 64; kind-pin patch already exists as commit `31d71e94` |
| QLT-15 | Phase 29 WR-02/WR-03 triaged; Phase 28 VALIDATION.md and SECURITY.md | 29 dispositions already `fixed` with named witnesses; Phase 28 evidence inventory with verified test names and packages; meta-test rules for the new docs |
| EVD-16 | One milestone-agnostic freeze and receipt gate replaces `check-phase32-*` | Complete refusal-reason catalog with triggers and test coverage gaps, manifest design, archived-receipt acceptance verified by running the old gate, ordering hazards |
</phase_requirements>

## Summary

This is an internal-tooling phase, so the research is a codebase audit. Every behavior the plan must preserve was read or executed in this session. Four findings change the plan or contradict wording in CONTEXT.md and must be resolved before the plans are written; they are listed first under "Findings That Amend CONTEXT.md".

The riskiest work is EVD-16. The old receipt gate has about 38 distinct refusal reasons, and only about 20 are exercised by an existing Go test. "Reproduce every existing refusal" therefore needs new differential rows for the untested reasons (catalog below). The old scripts hard-code the label `phase32` in two places that a find-and-replace will miss: `cut -c1-20` and `cut -c1-19` length checks on the first and last row. The second riskiest work is the WR-02 narrowing: the defect reason strings are copied in five places, not three (a fourth is `session.go:535-539`, a fifth is the C stderr text in `cgen_program.go`), and `session.go` is line-pinned by the Phase 16 emitter inventory.

The 5-plan shape in CONTEXT is feasible, but plans 3 and 4 are the heavy ones. The only hard ordering rule is that Plan 1 Task 1 (docs-only reopen) lands before any commit that touches `cmd internal examples native scripts testdata go.mod .github/workflows/ci.yml`.

**Primary recommendation:** Follow the CONTEXT plan shape. Resolve the four amendments below in the plans. Write the differential test table first (Plan 3), from the reason catalog in this document, so the new scripts are developed against an already-armed test.

## Findings That Amend CONTEXT.md

These are facts that conflict with, or sharpen, a locked decision. The decision's intent is unchanged; the implementation detail needs adjusting. The planner should record each as a plan-level note, not reopen the decision.

1. **The D-01 marker template does not match the regex.** D-01 shows `reopened 2026-10-xx (historical codedigest <64hex>, run 38335175306)`, then says to match what `phase32ReopenedMarker` accepts. The live regex is `^M006 final hosted receipts: reopened ([0-9]{4}-[0-9]{2}-[0-9]{2}) \(historical codedigest ([0-9a-f]{64})\)\.?$` [VERIFIED: internal/compiler/session/session_phase32_receipt_test.go:207]. It accepts no `, run N` suffix, and the Plan 1 Task 1 commit is docs-only, so the regex cannot be widened in the same commit (any `internal/` edit moves the digest). A line with the suffix would be classified `malformed_marker` and turn the test red. **Use exactly** `M006 final hosted receipts: reopened 2026-10-13 (historical codedigest b9253457705a224a0580d74896e8e5bd3821446a1d467626b7d611ca0d1c4b5c)` and put `run 38335175306` in the next prose sentence. The marker must stay on its own line, with at most the decoration characters ` \t>*-` before it [VERIFIED: session_phase32_receipt_test.go:253-259 trims `" \t>*-"`].

2. **`TestNoCheckPhase32Caller` scope: "outside `.planning/milestones/`" will fail on legitimate prose.** `git grep check-phase32` outside `.planning/milestones` currently matches these non-caller files [VERIFIED: git grep run this session]: `.planning/PROJECT.md:103`, `.planning/REQUIREMENTS.md:23`, `.planning/ROADMAP.md:99,121,125`, `.planning/phases/33-*/33-CONTEXT.md` (lines 14, 58, 65, 141, 173), `.planning/research/M007/SUMMARY.md:60`, `.planning/todos/completed/2026-10-12-phase32-review-hardening-after-m006.md:7,9`, plus this RESEARCH.md and every Phase 33 PLAN. The requirement text itself (EVD-16) names the old scripts. **Recommendation:** scan tracked files that are not under `.planning/` (a "caller" is a script, workflow, Go test, Makefile or README, never a planning document), and keep the existence check for the two deleted files. State this refinement in the plan. If the planner wants to keep `.planning/` in scope, the needle must be the invocation form (`sh scripts/` + `check-phase32-`), and still has to exempt the Phase 33 directory.

3. **`session.go` is line-pinned.** `testdata/phase16/public-emitter-consumers.json` pins `file:Func:line` for call sites of `cgen.Emit`/`EmitNative`, and `internal/compiler/session/session.go` is one of its files [VERIFIED: registry file list read this session]. The defect-reason copy at `session.go:535-539` is *above* the pinned `EmitNative` call (the 28-REVIEW cites it at line 3042). Replacing those three literals with `execution.DefectReason…` constants is safe only if **no line is added or removed in `session.go`**. Phase 28 hit exactly this failure (28-REVIEW CR-01, `TestPhase16PublicEmitterConsumerInventory`). Do the replacement in place, one line for one line, and run that test after the edit.

4. **New `33-VALIDATION.md` must be seeded `status: planned`, not `draft`.** `TestPhase20ValidationLifecycle` fails if any live or archived `*-VALIDATION.md` has `status: draft` [VERIFIED: verification_groundedness_test.go:1517-1519, 1536-1548]. The 32.1 precedent: `# status lifecycle: planned (plan-phase; TestPhase20ValidationLifecycle forbids committed drafts) → validated` [VERIFIED: 32.1-VALIDATION.md frontmatter]. Also keep the Phase 33 and Phase 28 VALIDATION files out of the *graded* path (next section).

## Architectural Responsibility Map

The generic tier table (browser, SSR, API) does not apply. The relevant owners in this repo:

| Capability | Primary owner | Secondary | Rationale |
|------------|---------------|-----------|-----------|
| Capture-size bound | `internal/compiler/native` (host side of the child process) | `execution.MaxApplicationEvidenceBytes` (the constant) | The parent reads a file the untrusted child wrote; the bound must be enforced by the reader, before allocation |
| Abort-like acceptance | `internal/compiler/interp` (leak detection) | `execution` (shared reason constants), `cgen` (emits the same text), `executionpeer` (re-checks reason), `session` (projection) | Interpreter is the model; native is the oracle for parity; one constant set removes drift |
| Digest path set and `.md` filter | `scripts/codedigest.sh` | Go twin `phase32CodeDigest` in a `_test.go` | One shell source, one Go twin proven equal by `TestPhase32CodeDigest` |
| Freeze and receipt verdicts | `scripts/check-freeze.sh`, `scripts/check-receipts.sh` | `testdata/receipts/<M>.manifest` (data) | Verdict logic is milestone-agnostic; all milestone facts live in the manifest and CLI args |
| Receipt production | `scripts/verify-phase32.sh` (per milestone) | manifest (case table, program list) | Builds and fixtures are bespoke; table is shared data |
| Living-claim policing | `session_phase32_receipt_test.go` | `.planning` docs | Docs carry the claim, a Go test polices it |

## Standard Stack

No external packages. Per the user's dependency stance, this phase uses POSIX `sh` (macOS `/bin/sh` is bash 3.2 in POSIX mode; Ubuntu CI `sh` is dash) and the Go standard library only. Go is `go1.24.0 darwin/arm64` locally [VERIFIED: `go version`]; Apple clang 21.0.0 is present [VERIFIED: `clang --version`]; `shellcheck` is installed locally at `/opt/homebrew/bin/shellcheck` but is **not** run in CI [VERIFIED: ci.yml grep for shellcheck returned nothing].

### Core
| Tool | Version | Purpose | Why Standard |
|------|---------|---------|--------------|
| POSIX sh | n/a | gate scripts | existing `check-phase32-*`, `codedigest.sh`, `verify-phase32.sh` are sh; the Go-tool alternative is locked out by D-04 |
| Go `testing`, `os/exec` via `testsupport.RunCLIErr` | go1.24 | differential and gate tests | every spawn in this repo goes through `testsupport.RunCLIErr` (bounded output, timeout); a repo lint (`scanUnboundedSpawns`) enforces it [CITED: testsupport.go:168-175 comment] |
| `io.LimitReader` + `io.ReadAll` | stdlib | bounded capture read | already the idiom at `readApplicationReceipt` |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| two sh scripts with a copied manifest reader | Go gate binary | rejected by D-04 (byte parity, new binary in the digest) |
| `os.Stat` then read | LimitReader only | rejected by D-15 (TOCTOU: the child writes the file) |

**Installation:** none. `go.mod` and `go.sum` must stay unchanged; assert with `git diff --stat -- go.mod go.sum` in the final verification.

## Package Legitimacy Audit

No external packages are installed or recommended by this phase. `gsd-tools package-legitimacy check` was not run because there is nothing to check.

**Packages removed due to [SLOP] verdict:** none
**Packages flagged as suspicious [SUS]:** none

## Architecture Patterns

### Data flow: the generic gate

```
                 testdata/receipts/M006.manifest
                 (label, programs, controls, "# case" rows)
                              |
   CLI: MANIFEST FREEZE CAND [SAME..]          CLI: MANIFEST LOG_A LOG_B   (+ env EXPECTED_CODEDIGEST)
              |                                          |
      check-freeze.sh                            check-receipts.sh
   read manifest -> label                      read manifest -> label,programs,controls,table
   argc check (prefix = label)                 argc check
   cd git toplevel (no_repo_root)              cd git toplevel; digest = env | sh scripts/codedigest.sh
   sha shape / SAME_AS / unknown_commit        per-log check_log (host, digest, tree, programs,
   not_descendant                                 case rows, control rows)  + cross-log digest check
   sh scripts/codedigest.sh --changed F C      duplicate_host
        (git log names + git diff names,
         .md removed)  non-empty => reject
              |                                          |
   "label freeze gate: status=pass ..."        "label receipt gate: status=pass ..."
                              |
                  scripts/codedigest.sh  (single source of: path set, .md filter, hash length)
                    default      -> sha256 of `git ls-tree -r HEAD -- paths`, .md lines removed
                    --rev REV    -> same at REV
                    --paths      -> the path set
                    --changed A B-> names changed in A..B over the path set, .md removed
```

### Recommended file layout

```
scripts/codedigest.sh                    # + --rev, --paths, --changed; derived hash length
scripts/check-freeze.sh                  # new, generic
scripts/check-receipts.sh                # new, generic
scripts/verify-phase32.sh                # reads table/programs from the manifest
scripts/check-phase32-freeze.sh          # deleted in plan 4 (not before goldens exist)
scripts/check-phase32-receipts.sh        # deleted in plan 4
testdata/receipts/M006.manifest          # label, programs, controls, 39 "# case" rows
internal/compiler/session/session_freeze_gate_test.go      # TestFreezeGate (renamed)
internal/compiler/session/session_receipt_gate_test.go     # TestReceiptGate, TestReceiptGateArchivedM006 (renamed)
internal/compiler/interp/interp_phase33_defect_test.go     # narrowing + mutant (same package)
internal/compiler/execution/defect_reasons.go              # shared constants (new file, avoids editing execution.go)
internal/compiler/native/native_phase33_capture_test.go    # TestPhase33CaptureReadBounded (new file: native_app_test.go is not registry-pinned, but a new file keeps diffs small)
```

A manifest file name must **not** end in `.schway`: several corpus walkers walk all of `testdata/` and select by that extension [VERIFIED: session_admission_divergence_test.go:134-146, session_phase5_corpus_test.go, native_test.go:1670-1685]. `M006.manifest` is invisible to them.

### Manifest syntax (recommended, discretionary)

Keep it greppable by the same `sed`/`grep` idioms the old scripts used, so the reader is about 10 lines and the case rows stay byte-identical to the producer's old block:

```
label=phase32
programs=checksum,adler_low,adler32_cli,adler32_fixture,byte_index_oob,sum_to_n,fizzbuzz
controls=guide,route-audit,codedigest,receipt-rules
# case checksum empty 0 1 -
...
```

Reader (copy-pasted into both scripts; `field` style from the old scripts, no `eval`):

```sh
# Source: pattern from scripts/check-phase32-receipts.sh (sed -n 's/^key=//p' | sed -n '1p')
mf_get() { sed -n "s/^$1=//p" "$manifest" | sed -n '1p'; }
label=$(mf_get label)
want_programs=$(mf_get programs)
controls=$(mf_get controls)
case_table=$(grep '^# case ' "$manifest" || true)
```

Every comparison that old code made against the hard-coded `phase32 ` prefix must use `"$label "` instead. Never put `$label` in a `printf` format position; use `printf '%s freeze gate: ...' "$label"`.

### Pattern: the unexported seam mutant (D-07-42), the model for WR-02

`decoderRequireNonEmpty` is a package-level `var` holding a closure, defaulting to the production behaviour; the same-package test swaps it and restores with `defer` [VERIFIED: internal/compiler/core/core.go:509 `var decoderRequireNonEmpty = func(value string) bool { return value != "" }`; core_internal_test.go:35-37]. For WR-02 the two seams are `abortLikeAcceptLive` (is the live set exactly `[ownerOpID]`) and `abortLikeReasonMatches` (does `Output` equal the kind's reason). The test must show that swapping each seam to "always true" makes a previously refused row pass (that is what "killed" means here), then restores.

### Anti-patterns to avoid
- **Hard-coding `phase32` anywhere in the new scripts**, including the first/last-row checks (see Pitfall 1).
- **A shared sh library** for the manifest reader (D-04: copy-paste).
- **Recomputing the digest when `EXPECTED_CODEDIGEST` is set.** The pass line's `digest_source=env` depends on it.
- **Deleting the old scripts before the goldens are frozen** (D-10 ordering).
- **Adding or removing a line in `session.go`** (Finding 3).

## Don't Hand-Roll

| Problem | Don't build | Use instead | Why |
|---------|-------------|-------------|-----|
| Bounded file read | `Stat` then `ReadFile` | `os.Open` + `io.ReadAll(io.LimitReader(f, limit+1))` | `readApplicationReceipt` is the in-repo model; Stat-then-read is a TOCTOU race against the child |
| Subprocess in a test | raw `exec.Command` | `testsupport.RunCLIErr` | bounded output and timeout; repo lint |
| Spawning git in fixtures | a new helper | `phase32Git`, `phase32ShDigest`, `phase32CodeDigest` in `session_phase32_receipt_test.go:49-77` | already used by 4 test files; keep the names or move them with the rename |
| Finding archived artifacts | hard-coded milestone path | `phaseArtifactGlob("32-*", "32-RECEIPT.json")` (`session_test.go:2579-2589`) | searches both live and `milestones/*-phases/` |
| Seam restore | manual cleanup | `defer func(){ seam = original }()` | D-07-42 pattern |
| Coverage-kind enum | importing gsd-core | copy the six-value literal | user stance; the patch already exists as commit `31d71e94` |

**Key insight:** every piece of this phase has an in-repo precedent. The risk is parity drift, not novelty.

## Exact Current Behavior (the parity baseline)

All values below were read from the scripts this session [VERIFIED: scripts/check-phase32-freeze.sh, check-phase32-receipts.sh, codedigest.sh, verify-phase32.sh].

### Freeze gate: reasons in evaluation order

stderr line format: `phase32 freeze gate: status=fail reason=%s`; exit 1; stdout empty. Pass line: `phase32 freeze gate: status=pass freeze=%s candidate=%s`.

| Order | Reason | Trigger | Existing test |
|-------|--------|---------|---------------|
| 1 | `usage` | fewer than 2 args | `bad-sha` subtest (`expectReject usage, f.freeze`) |
| 2 | `bad_sha` | any arg not 40 lowercase hex chars (checked over **all** args including SAME_AS) | `bad-sha` |
| 3 | `sha_mismatch` | any SAME_AS differs from CANDIDATE | `headsha-mismatch`, `origin-mismatch` |
| 4 | `unknown_commit` | freeze or candidate not a commit | `unknown-commit` |
| 5 | `not_descendant` | freeze not an ancestor of candidate | `not-descendant` |
| 6 | `git_log_failed` | `git log` errors | none (not reachable without a corrupt repo) |
| 7 | `digest_path_changed_after_freeze` | `git log F..C` over paths (minus `*.md`) non-empty, or `git diff --quiet` fails | `post-freeze-digest-change`, `md-only-change-after-freeze-accepted` (armed half) |

New arg shape: `check-freeze.sh MANIFEST FREEZE CANDIDATE [SAME_AS…]`, so the arity floor becomes 3 after the manifest. With **zero** args there is no manifest, so no label: define the fallback prefix explicitly (recommend an empty label, giving `freeze gate: status=fail reason=usage`). This case has no old equivalent that can be compared and is not an M006 parity obligation.

When `git_log_failed` becomes `codedigest.sh --changed`, redirect that script's own stderr (`2>/dev/null`) so the gate still prints exactly one stderr line.

### Receipt gate: reasons in evaluation order

stderr: `phase32 receipt gate: status=fail reason=%s`; exit 1. Pass: `phase32 receipt gate: status=pass hosts=%s,%s codedigest=%s programs=%s cases=%s digest_source=%s`. Suffix `_$log` is the path as typed.

Gate-level: `usage_two_logs_required` (argc != 2; new: != 3 after the manifest), `expected_codedigest_unknown` (empty digest), `producer_missing` (**no new equivalent**: becomes `manifest_missing`), `case_table_empty`, `programs_line_missing`, then the cross-log check `log_digest_differs` (runs first, only when both files exist), then per-log `check_log` for LOG_A, then LOG_B, then `duplicate_host_<host>`.

Per-log `check_log`, in order. **"Tested?" is whether a Go test in `session_phase32_gate_test.go` already seeds it.** The untested ones are the differential-test work (a mutation recipe is given so the planner can write the rows directly):

| # | Reason (suffix `_$log` unless noted) | Trigger / mutation recipe | Tested? |
|---|---|---|---|
| 1 | `host_missing_` | file absent | yes |
| 2 | `log_empty_` | zero-byte file | yes |
| 3 | `failed_or_unavailable_row_` | `grep -E 'status=(fail\|incomplete\|unavailable\|unobserved)\|unavailable'` matches anywhere | yes (`failed-case-status`) |
| 4 | `identity_row_count_<n>_` | `^phase32 host identity ` count != 1 | yes |
| 5 | `result_row_count_<n>_` | `^phase32 host result ` count != 1 | **no**: duplicate the result row (but then also keep it last) |
| 6 | `identity_not_first_` | first `^phase32 ` row is not the identity row | **no**: move the toolchain row above identity |
| 7 | `result_not_last_` | last `^phase32 ` row is not the result row | yes |
| 8 | `unknown_phase32_row_` | a `^phase32 ` row outside {host identity, toolchain, case, control, host result} | yes |
| 9 | `toolchain_row_count_` | toolchain rows != 1 | **no**: duplicate or drop |
| 10 | `toolchain_row_malformed_` | toolchain row not `go="…" clang="…"` | **no**: `go=""` |
| 11 | `unknown_host_` | identity `host=` not macOS/Linux | **no** |
| 12 | `arch_missing_` | identity lacks `arch=` | **no**: delete ` arch=x86_64` |
| 13 | `target_missing_` | lacks `clang_target=` | **no** |
| 14 | `go_host_missing_` | lacks `go=` | **no** (note `field` is anchored: `go=` on the toolchain row does not satisfy it) |
| 15 | `headsha_missing_` | lacks `headsha=` | **no** |
| 16 | `digest_mismatch_` | identity codedigest != expected | yes |
| 17 | `tree_not_clean_` | identity `tree=` != clean | yes |
| 18 | `result_host_mismatch_` | result host != identity host | **no** |
| 19 | `result_digest_mismatch_` | result codedigest != expected | **no** |
| 20 | `result_tree_not_clean_` | result tree != clean | **no** |
| 21 | `result_not_pass_` | result status != pass **and** not in the pattern of #3 | **no**: use `status=skipped` (`fail` would be caught at #3 first) |
| 22 | `result_programs_mismatch_` | result `programs=` != manifest programs | yes |
| 23 | `program_without_table_entry_<program>` (no `_$log`) | a manifest program has no `# case <program> ` row | **no**: needs a variant **manifest** (tests may write one to a temp dir) |
| 24 | `program_missing_<program>_$log` | no `phase32 case program=<p> ` row in the log | yes |
| 25 | `result_control_count_` | result `controls=` != 4 (new: != number of names in `controls=`) | **no** |
| 26 | `result_case_count_` | result `cases=` != table count | **no** |
| 27 | `case_row_count_` | case rows != table count | **no**: add a duplicate case row |
| 28 | `case_missing_or_altered_<program>_<key>_$log` | exact expected row string absent or not exactly once | yes |
| 29 | `control_row_count_` | control rows != 4 | yes |
| 30 | `control_missing_or_altered_<name>_$log` | per-name regex `^phase32 control name=<n> package=[^ ]+ tests=[^ ]+( [^ ]+)* status=pass$` not exactly once | yes |
| gate | `log_digest_differs` | both identity digests present and different | yes |
| gate | `duplicate_host_<host>` | both logs same host | yes |
| gate | `case_table_empty`, `programs_line_missing`, `expected_codedigest_unknown`, `usage_two_logs_required` | manifest/argc/env variants | **no** |

Notes for the differential test:
- To feed old and new the same inputs, write the logs once and call the two scripts on the **same paths**, so `_$log` is byte-equal.
- For manifest-variant rows (#23, `case_table_empty`, `programs_line_missing`), the old script reads `$root/scripts/verify-phase32.sh` (root derived from `$0`). Run the old script from a temp tree containing `scripts/check-phase32-receipts.sh`, a variant `scripts/verify-phase32.sh`, and `scripts/codedigest.sh`, with `PHASE32_EXPECTED_CODEDIGEST` set so no git repo is needed. Run the new script with a variant manifest.
- The field regexes need exact count control: `grep -Fxc -- "$want" "$log"` must equal `1`.
- The `cut -c1-20` check is `[ "$(grep '^phase32 ' "$log" | sed -n '1p' | cut -c1-20)" = "phase32 host identit" ]` and `cut -c1-19` against `"phase32 host result"`. They are tied to the 7-character label (Pitfall 1).

### `codedigest.sh`

Listing: `git -c core.quotePath=false ls-tree -r HEAD -- cmd internal examples native scripts testdata go.mod .github/workflows/ci.yml`, then `grep -v '\.md$'`, then the hasher; output must be exactly 64 lowercase hex chars; failures print `codedigest: status=fail reason=<name>` with names `unavailable_tool_sha256`, `no_repo_root`, `git_ls_tree_failed`, `empty_listing`, `bad_digest` [VERIFIED: scripts/codedigest.sh, whole file]. Because `ls-tree` lines contain the file mode, the digest also changes with a mode flip: `check-phase32-receipts.sh` is `-rw-r--r--` and the freeze script `-rwxr-xr-x` [VERIFIED: ls -l]. Always invoke the gates as `sh script`, never by path.

The hash-length derivation (IN-03): `expected_len=$(printf '' | $hasher | cut -d' ' -f1)` then `${#expected_len}`. sha256 of empty input is `e3b0c442…b855`, so the derived length is 64 on both `shasum -a 256` and `sha256sum`. [ASSUMED: this is the standard sha256 empty-input value; the planner's test should assert the derived length equals the length of the real digest.]

### Producer `verify-phase32.sh`

The case table (39 rows) and `phase32_programs=` are at lines 24-70 and 21. It reads its own file twice to build `case_table` and prints `phase32 control ...` rows for exactly the four names `guide route-audit codedigest receipt-rules` and `controls=$control_count` in the result row [VERIFIED: verify-phase32.sh]. Its `tree_state` duplicates the digest path list as `digest_paths=` (line 70) and must switch to `sh scripts/codedigest.sh --paths` (D-08). The four control tests it runs are `TestPhase32GuideCleanCheckout`, `TestPhase32PublicRouteAudit`, `TestPhase32CodeDigest` and `TestPhase32ReceiptLivingClaims` [VERIFIED: verify-phase32.sh `run_control` lines]. **Renaming `TestPhase32CodeDigest` or `TestPhase32ReceiptLivingClaims` would break archived log provenance** (D-11 keeps those files for that reason). Do not rename those test functions.

`git diff 9f8f878f HEAD -- scripts/verify-phase32.sh scripts/check-phase32-*.sh scripts/codedigest.sh` is empty [VERIFIED: git diff --stat run this session], so the HEAD table equals the table at the M006 receipt commit. The D-05 "equals `git show 9f8f878f:…`" check is therefore trivially satisfiable now and can be made a one-line executor check.

## Archived M006 Receipts: What the Generic Gate Must Accept

[VERIFIED by running the old gate this session]

```
PHASE32_EXPECTED_CODEDIGEST=b9253457705a224a0580d74896e8e5bd3821446a1d467626b7d611ca0d1c4b5c \
sh scripts/check-phase32-receipts.sh \
  .planning/milestones/M006-phases/32-clean-checkout-guide-and-final-receipts/receipts/receipt-macos-latest.log \
  .planning/milestones/M006-phases/32-clean-checkout-guide-and-final-receipts/receipts/receipt-ubuntu-latest.log
→ phase32 receipt gate: status=pass hosts=macOS,Linux codedigest=b9253457…1c4b5c programs=checksum,adler_low,adler32_cli,adler32_fixture,byte_index_oob,sum_to_n,fizzbuzz cases=39 digest_source=env   (exit 0)
```

So the manifest needs exactly: label `phase32`; programs `checksum,adler_low,adler32_cli,adler32_fixture,byte_index_oob,sum_to_n,fizzbuzz`; the four control names; the 39 verbatim `# case` rows. Nothing else. The digest comes from env (`32-RECEIPT.json` field `codedigest`, same value) [VERIFIED: 32-RECEIPT.json lines 1-6]. Also present: `32-RECEIPT.run-38315799626.json` and `receipts/historical-38315799626/` (do not touch; D-02).

Freeze rows, executed this session with the old script [VERIFIED]:

| Call | Result |
|------|--------|
| `0aac0a7b…` → `9f8f878f…` (same-as `9f8f878f…`) | pass |
| `0aac0a7b…` → `31d71e94…` (a real post-freeze code commit) | `digest_path_changed_after_freeze` |
| `0aac0a7b…` → `7ab2f48e…` (the revert commit; **tree equal** to the frozen digest) | `digest_path_changed_after_freeze`, from the `git log` half only |
| `0aac0a7b…` → HEAD | `digest_path_changed_after_freeze` |
| same HEAD call but from `cmd/` | **pass (false accept)** |

The last two rows are the WR-01 bug reproduced live. The `7ab2f48e` row is the best "keeps both the log and the diff" witness: at that commit the net `git diff` over the digest paths is empty (32.1 added a change then reverted it, and `sh scripts/codedigest.sh` at HEAD still prints `b9253457…`), so only the log half rejects. Use full 40-char SHAs: `31d71e9428eeb3aa89cd376b2a6fd5047bb4e69d`, `7ab2f48e28f2401f0545e9c414c496795651090b`, `0aac0a7b4801292b3f54f7ef57ed5fb83374c08e`, `9f8f878fc5328f953eac6ed04a716478ae986fab`.

History dependence: CI checks out with `fetch-depth: 0` in all four jobs [VERIFIED: ci.yml lines 63-65, 103-105, 171-173, 204-206], and all four commits are reachable from tag `vM006` [VERIFIED: `git merge-base --is-ancestor 0aac0a7b vM006`] but **not** from `main` (main is far behind). If the work is ever squash-merged, the pinned SHAs vanish unless tags are fetched. Make the real-repo rows start with `git cat-file -e <sha>^{commit}` and fail with a message naming the missing SHA and the remedy (`git fetch --tags`), rather than skipping silently. Keep a temp-repo fixture version of every row so the logic is covered with no history dependence.

### "64-path count"

The roadmap mentions a "64-path count". There is no 64-path count. The `64` in `codedigest.sh` is the sha256 hex length (`[ "${#digest}" -eq 64 ]`), and `phase32Hex64` in the Go test is the same width. The path set has 8 entries. D-08's "derive from the hasher's empty-input output" is the correct reading.

## Living-Claim Check (D-10a) Mechanics

[VERIFIED: session_phase32_receipt_test.go:204-405, 436-470]

- Docs read by `TestPhase32ReceiptLivingClaims/real-checkout`: `.planning/LANGUAGE-MATURITY.md`, `.planning/PRODUCT-ROADMAP.md`, `.planning/STATE.md` (line 442). Only the first two carry a marker today (line 239 and line 10).
- Parsing: for each line, `strings.TrimLeft(line, " \t>*-")`, then `HasPrefix(line, "M006 final hosted receipts:")`. Prose that merely mentions the phrase mid-line is ignored; a line *starting* with it that matches neither grammar is `malformed_marker`.
- Bound marker: `^M006 final hosted receipts: bound to codedigest ([0-9a-f]{64}) \(run ([0-9]+)\)\.?$`. A bound digest that differs from either `receipt.Codedigest` or `ComputedDigest` is `stale_bound_claim`. The reopened marker never fails on drift.
- `ComputedDigest` is `phase32CodeDigest(t, root)`: the Go twin, HEAD listing. So the first digest-moving commit (e.g. a `scripts/` edit) turns the bound marker stale. This is why Plan 1 Task 1 comes first.
- The receipt is found by `phaseArtifactGlob("32-*", "32-RECEIPT.json")`, which searches `.planning/phases/` and `.planning/milestones/*-phases/`. Moving or renaming the archive directory breaks this.
- Schema is hard-coded `schway.m006-receipt/1`, hosts `ubuntu-latest`/`macos-latest`, and `phase32ReceiptPrograms`. Generalizing "by milestone label" (D-11) means parameterizing at least: the marker prefix (`<MILESTONE> final hosted receipts:`), the schema string, and the program list. Keep the M006 values as defaults/rows so the 28 existing table rows port unchanged. [ASSUMED: M007 will name its schema `schway.m007-receipt/1`; this is the natural continuation but is Phase 38's decision.]
- Safe to rewrite prose: `TestPhase25LivingRoadmap` requires both living roadmap docs to keep the words `Historical receipts`, `hosted`, `pointer`, etc. and to not contain the sentence `Phase 25 passed on Linux and macOS` [VERIFIED: native/phase25_utility_test.go:293-306]. `Historical receipts` occurs at `PRODUCT-ROADMAP.md:493` and `LANGUAGE-MATURITY.md:94,131`, outside the lines being rewritten, so the D-01 rewrite does not endanger it. Re-run `TestPhase25LivingRoadmap` after the doc edits.
- Extending the doc set to `PROJECT.md` and `MILESTONES.md` is safe: neither has a line starting with the marker phrase (their receipt claims are mid-sentence prose at `PROJECT.md:171,609-610`, `MILESTONES.md:19-20`). Keep it that way when rewriting them: do not start a line with `M006 final hosted receipts:` unless it is the real marker.

## Defect Parity and the WR-02 Narrowing

### Current predicate

`isAbortLikeDefect(program core.Program, result Execution) bool` at `interp.go:439-461`: requires `Outcome.Kind == execution.OutcomeDefect`, a non-empty `Events`, final event `Kind == "function.defected"` with ID suffix `:event:defected`, then **scans every function** for an operation with that ID and accepts kinds `OpBytesAt, OpAddChecked, OpRemainderU64, OpDefect`. It ignores `LiveResources` and `Output` [VERIFIED: interp.go:432-461]. Callers: `interp.go:319` (`RunWithFixtureBytes`) and `interp.go:371` (`RunWithFileBytes`), both as `if len(result.LiveResources) != 0 && !isAbortLikeDefect(program, result)`.

The callers *already count* the acquire ops (`acquires` counter over `operation.Foreign.Mode == "acquire"`) but do not keep the ID. The change is to record `ownerOpID = operation.ID` in those loops (exactly one by the `acquires != 1` guard). Live resource IDs are operation IDs: `top.live[operation.ID] = true; top.liveOrder = append(top.liveOrder, operation.ID)` in both `fixtureForeignCall` and `fileForeignCall`, and the live list comes from `frameLiveResourceList` [VERIFIED: interp.go:430-440 region, 715-723].

### Reason strings: five copies, not three

[VERIFIED by grep this session]

| Reason | interp | executionpeer | cgen (event reservation) | cgen (C text) | session |
|---|---|---|---|---|---|
| `U64 addition overflow` | `interp.go:1705` | `executionpeer.go:392` | `cgen_program.go:233` | `cgen_program.go:2725` (`overflow.Reason`) and the stderr `fputs` at ~2738 region | `session.go:535` |
| `U64 remainder by zero` | `interp.go:1731` | `executionpeer.go:409` | `cgen_program.go:238` | `cgen_program.go:2741` and the `fputs` | `session.go:537` |
| `byte index out of bounds` | `interp.go:1754` | `executionpeer.go:426` | `cgen_program.go:243` | `cgen_program.go:2757` and the `fputs` at ~2754 | `session.go:539` |

The native app shell prints these as `schway: <reason>` followed by `exit(65)` (inline literal in the emitted C, e.g. `fputs("schway: U64 addition overflow\n", stderr)`). When replacing with a constant the generated bytes must not change: build the string with concatenation or `fmt.Fprintf`, and confirm with `TestPhase272AcceptedCorpusByteIdentity`, `TestPhase273ConformanceCorpusByteIdentity` and `TestPhase28PrunedEmitMatchesGolden`. The user-defect text uses a different prefix (`schway app: boom` in the parity test), so `OpDefect` keeps `operation.Reason`.

Put the constants in a **new file** in `internal/compiler/execution` (e.g. `defect_reasons.go`) rather than editing `execution.go`, to keep the diff reviewable. `interp`, `executionpeer`, `cgen` and `session` already import `execution`.

### Existing tests that must stay green (and need signature updates)

- `TestPhase321AbortLikeDefectPredicate` and `TestPhase321AbortLikeDefectFileRoute` in `interp/interp_phase321_defect_test.go`. The predicate test builds a program with **no function IDs** (`core.Program{Functions: []core.Function{{Linear: ...}}}`) and defect events with no `FunctionID`, and calls the 2-arg form directly [VERIFIED: interp_phase321_defect_test.go:15-70]. After the narrowing, every accepted row needs a function ID, a matching event `FunctionID`, the right `Output` and a `LiveResources` of exactly `[ownerOpID]`. The existing refused rows (call, write, nonlocal, call-depth, unknown op, wrong event kind, returned, typed_failure, empty events) must be kept and re-pass with the 3-arg form.
- `TestPhase321DefectParityWithLiveOwner` in `native/native_phase321_defect_parity_test.go`: exits 65 natively with stderr equal to `variant.prefix + interpreter final Output + "\n"`, for four variants on the **fixture route only**. [VERIFIED: file read in full.] D-14 step 1 asks for the **file route** too, which does not exist natively yet. Build it from the existing Phase 31 helpers `phase31ControlCopy`, `phase31ControlMutateLibrary` (replace `fn adler32`), `phase31ControlBuild`, `phase31ControlNative`, `phase31ControlInterp` (`native/native_phase31_controls_test.go:26-125`), which already build the checksum example in a temp dir, run natively with a file path, and run `interp.RunWithFileBytes` on it. Call-depth and nonlocal defects stay out of the accepted set.
- `session/session_phase29_interp_fixture_test.go` (`requireAbortLike`, lines ~110-185) also exercises the predicate through `session`.

### D-13 conflict check

29 WR-03's fix accepted a live owner for all four defect kinds; D-13 additionally *refuses* a defect that has an unrelated second live resource. Both can hold. WR-03 is about the owner being live; D-13 is about nothing else being live. The new refusal must still return a hard error with the existing message prefix `model execution reached a terminal outcome with live local resources` so `TestPhase29*` expectations on that text keep passing. [ASSUMED: no test pins that exact text for the new case; the planner should grep for it before changing the message.]

### Mutant design

`TestPhase33AbortLikeWideningMutantKilled` needs the seams as package-level `var` closures used by `isAbortLikeDefect` (not constants). "Killed" means: with the seam widened to always true, a row that production refuses (extra live resource; wrong reason) is accepted, proving the row bites. Restore with `defer`. Test file in package `interp` (same-package), like `interp_phase321_defect_test.go`.

## Bounded Capture Read (QLT-13)

### Current code

```go
// Source: internal/compiler/native/native_app.go:362-380 (read this session)
if mode == EvidenceEvents {
    report.CaptureStatus = EvidenceStatusIncomplete
    data, readErr := os.ReadFile(capturePath)
    limit := r.evidenceCapacityLimit()
    switch {
    case readErr == nil && len(data) > limit:
        report.CaptureStatus = EvidenceStatusCapacityExhausted
    case readErr == nil:
        capture, captureErr := decodeApplicationCapture(data, limit)
        ...
```

`evidenceCapacityLimit()` returns `r.evidenceLimit` when `> 0`, else `execution.MaxApplicationEvidenceBytes` which is `64 * 1024` [VERIFIED: native_app.go:108-115; execution.go:13-15]. The outcome string is `EvidenceStatusCapacityExhausted EvidenceCaptureStatus = "capacity_exhausted"` (`native_app.go:82`), ending as `ToolError{Code: "native.evidence_capacity_exhausted"}`. `decodeApplicationCapture` already refuses `len(data) == 0 || len(data) > limit` (belt and braces).

### Minimal change

Replace only the `os.ReadFile` with a helper; keep the `switch` and its status mapping exactly:

```go
// Source: modelled on internal/compiler/native/native_app.go:649-659 (readApplicationReceipt)
func readApplicationCapture(path string, limit int) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(io.LimitReader(file, int64(limit)+1))
}
```

The caller keeps `len(data) > limit` -> capacity exhausted. A missing file or a read error stays `EvidenceStatusIncomplete` (existing `phase22AssertIncompleteEvidence` rows "missing capture" and "partial capture" must stay green). `io` and `os` are already imported in this file.

### The mutant catcher must go through `RunApplicationWithEvidence`

If a test only calls `readApplicationCapture` directly, a mutant that restores `os.ReadFile` *at the call site* (bypassing the helper) is not caught. The catcher has to drive `Runner.RunApplicationWithEvidence` so the allocation happens on the real path. Pattern already in the package: `writePhase22EvidenceScript(t, body)` builds a shell-script "artifact" whose body writes to `"$SCHWAY_APP_EVIDENCE_PATH"`, and `phase22CompleteCapture(t)`, `shellQuote`, `assertPhase22LaunchCount` exist in `native_app_test.go` (package `native`, internal) [VERIFIED: native_app_test.go:458-520, 560-640].

Create the sparse file from the child script, not Go. Verified on this macOS host: `dd if=/dev/null of=F bs=1 seek=1073741824` produces a 1073741824-byte file occupying 0 blocks, and `/usr/bin/truncate` exists [VERIFIED: ran both]. GNU `dd` behaves the same for a regular file. Use the `dd` form (POSIX).

### Budget numbers (the one real risk in D-16)

`runtime.ReadMemStats` `TotalAlloc` is a process-wide cumulative counter, so the delta also contains the exec plumbing and JSON marshalling for the run. With a tiny injected limit such as `evidenceLimit = 8` (as the existing capacity tests do), `16*limit = 128` bytes, which any run exceeds, so the test would fail on the **correct** code. Pick the limit for the mutant-catcher case so that `16*limit` is comfortably above the run's own overhead and far below 1 GiB: use `limit = execution.MaxApplicationEvidenceBytes` (64 KiB, i.e. do not inject), giving a 1 MiB budget against a 1 GiB mutant allocation. [ASSUMED: the non-capture overhead of one `RunApplicationWithEvidence` call is well under 1 MiB; the planner should measure it once on the correct code and record the number in the SUMMARY.] Keep the exactly-limit and limit+1 cases at a small injected limit (they assert status only, not allocation). Run without `t.Parallel`.

A `Truncate`d 1 GiB file read by the mutant touches ~1 GiB of zero pages. Expect roughly 1 GiB RSS for under a second on the one-time hand run. If that is a worry on a small machine, 256 MiB is still four orders of magnitude over the budget; the 1 GiB figure is a CONTEXT choice, not a correctness requirement.

Exactly-limit: pad a valid capture with trailing spaces to exactly `limit` bytes. `decodeApplicationCapture` does `decoder.Decode(&trailing)` and requires `io.EOF`; trailing whitespace is consumed as EOF (verified by reading the function at `native_app.go:560-590`; the existing `TestPhase27OutputEvidenceExactLimit` already proves a non-padded capture of exactly `len(capture)` bytes completes). A cheap alternative that proves the same boundary without padding: set `runner.evidenceLimit = len(capture)` for exactly-limit and `len(capture)-1` for limit+1, both using the real capture (the existing `TestPhase27OutputEvidenceExactLimit` already does the first). Prefer reusing this; it avoids a shell-side padding command.

## IN-01, IN-02, IN-03, kind pin

- **IN-01 (covered):** `uncovered_route` is emitted at `main_phase32_routes_test.go:387`, `module_aware_set_drift` at `:497`; seeded subtests `frobnicate` (`:725-728`) and `set-drift` (`:788-790`) [VERIFIED]. Disposition names `TestPhase32PublicRouteAudit`. Run it once and record.
- **IN-02 (fix):** The assertion at `cmd/schway/main_phase321_run_input_test.go:31` is `strings.Contains(stdout+stderr, dir) || strings.Contains(stdout+stderr, "n:")`. Replace the on-disk `examples/sum_to_n.schway` read with this inline fixture and write it to `dir`:

  ```
  module phase321.run_input

  export {
    fn main
  }

  fn main(zq_param_marker: U64) -> U64 {
    zq_param_marker
  }
  ```
  [VERIFIED: I built the CLI and ran `schway --json run --engine=interpreter|native` on exactly this file this session: both print `status":"usage_error"`, a single diagnostic `tool.run_input_required`, and exit 64.] Add the control `strings.Contains(string(fixture), "zq_param_marker")`. The two engines' JSON must still be equal (existing last assertion).
- **IN-03 (fix):** D-08. Closed only if the `.md`-only test ships. Note two tests already exist and must be *ported, not dropped*: `md-edit-ignored` (`TestPhase32CodeDigest`) and `md-only-change-after-freeze-accepted` (freeze). Extend the second to also assert `sh scripts/codedigest.sh --changed F C` prints nothing for the `.md`-only commit and prints the file for the `.go` commit.
- **Coverage-kind pin (fix):** the exact patch already exists as commit `31d71e94` (`git show 31d71e94 -- internal/compiler/session/summary_coverage_test.go`): adds a `kinds` slice, `case "kind":` in the entry switch, an enum `switch` over `"unit", "integration", "e2e", "automated_ui", "manual_procedural", "other"`, and a `"bad-kind"` seeded control (`kind: unit` -> `kind: manual`). It was reverted by `7ab2f48e` only to preserve the M006 digest. Re-apply it (cherry-pick or retype) and add the comment citing the enum. The enum in the local gsd-core install is `VALID_KINDS = ['unit','integration','e2e','automated_ui','manual_procedural','other']` at `~/.claude/gsd-core/bin/lib/coverage.cjs:65-67` (CONTEXT cites line 66; the `const` is at 65) [VERIFIED: file read]. All existing SUMMARYs from phase 32 on use only `unit` (77) and `manual_procedural` (1) [VERIFIED: grep this session], so the pin turns no existing file red. New Phase 33 SUMMARYs are scanned too (`summaryCoverageFirstPhase = 32`), so executors must write `kind:` values from the enum.
- **Phase 29 WR-02/WR-03:** already `fixed` in `29-REVIEW-DISPOSITION.md`, with witnesses named (`TestPhase321ByteViewEntryRefused`, `TestPhase321DefectParityWithLiveOwner`) [VERIFIED]. `TestPhase321ByteViewEntryRefused` lives in `internal/compiler/check/check_phase321_byteview_entry_test.go`; the parity test in `internal/compiler/native/`. One command: `go test ./internal/compiler/check ./internal/compiler/native -run 'TestPhase321ByteViewEntryRefused|TestPhase321DefectParityWithLiveOwner' -count=1`.

## Phase 28 VALIDATION.md and SECURITY.md

### Which meta-tests apply, and how to stay out of the graded path

| Meta-test | What it checks | Consequence for a new Phase 28 VALIDATION.md |
|---|---|---|
| `TestPhase20ValidationLifecycle` | no `*-VALIDATION.md` (live or archived) has `status: draft` | use `status: validated` |
| `TestVerificationGroundedness` | `go test` commands in Tier-A documents resolve to real tests via the static index | every `-run` name must exist in the named package |
| `TestValidationRowGradesAreEarnedOverArchivedCorpus` | grades a table **only if** it is under a heading starting `## Per-Task Verification Map` **and** has an `Automated Command` column (`validationSection`, `primaryValidationTable`), and `validationTableEligibleForGrading` is true for `validated` status [VERIFIED: evidence_grade_test.go:1037-1056, 1075-1100, 1382-1393] | a graded table also needs `Grade` and `Non-inertness` columns and a run record. **Avoid it:** title the section `## Per-Requirement Verification Map` (the 29 form) or `## Requirement → Test Map` (the 32.1 form). Neither triggers grading. |
| `TestStaticTestIndexMatchesGoTestList` | static index equals `go test ./... -list .` | no new test files needed for Phase 28, so unaffected |

Do not call the section `Per-Task Verification Map`.

### Verified evidence inventory for the Phase 28 rows

All names exist [VERIFIED: grep of `^func Test…(` this session]. Requirements are DX-17, DX-18, EVD-13 (28-VERIFICATION Requirements Coverage).

| Req | Behavior | Package | Tests |
|---|---|---|---|
| DX-17 | uniform exit class across check/build/run on both engines, absolute and relative paths | `./internal/compiler/native` | `TestExitClassesUniformAcrossSurfaces` (`native_phase28_test.go`) |
| DX-17 | native tool failure is operational (exit 3) | `./internal/compiler/native`, `./internal/compiler/session` | `TestNativeToolFailureIsOperational` (exists in both packages) |
| DX-17 | exit codes from one table | `./internal/compiler/exitcode`, `./internal/compiler/session`, `./internal/compiler/cgen` | `TestNoDuplicatedExitLiterals`, `TestExitCodeDocsTable`, `TestEmittedCExitLiteralsMatchTable` |
| DX-17 | G-28-3 relative-path native run | `./internal/compiler/session` | `TestPhase28NativeCommandRelativePath` |
| DX-18 | unreachable function runs; entry manifest | `./internal/compiler/native`, `./internal/compiler/session` | `TestUnreachableFunctionProgramRuns`, `TestPhase28EntryExitClassManifest` |
| DX-18 | prune properties and golden; byte-identical C for reachable programs | `./internal/compiler/session` | `TestPhase28PruneCorpusProperties`, `TestPhase28PrunedEmitMatchesGolden`, `TestPhase272AcceptedCorpusByteIdentity`, `TestPhase273ConformanceCorpusByteIdentity` |
| EVD-13 | `app verify` names the capacity limit and stays in budget | `./internal/compiler/session` | `TestAppVerifyExhaustedCases`, `TestAppVerifySixteenExhausted`, `TestAppVerifyReportBudget`, `TestAppVerifyReportOversize`, `TestAppVerifyOutcomeDigest` |
| EVD-13 | refusal survives report-write failure; stale report removed (28 WR-02/WR-03) | `./internal/compiler/session` | `TestAppVerifyRefusalSurvivesReportWriteFailure`, `TestAppVerifyRemovesStaleReportOnRefusal` |
| (CR-01) | emitter inventory pin | `./internal/compiler/session` | `TestPhase16PublicEmitterConsumerInventory` |

Frontmatter and sections to copy from `29-VALIDATION.md`: `phase`, `slug`, `status`, the "Reconstructed after execution from the … PLAN/SUMMARY set (State B)" comment, `nyquist_compliant`, `wave_0_complete`, `created`, then Test Infrastructure, Sampling Rate, the map, Wave 0, Manual-Only, Sign-Off, Validation Audit. Honest-gap notes the evidence supports: 28-VERIFICATION says "Hosted CI: not run for this revision; all evidence is local focused tests plus the shipped binary", and the `ambiguous entry` binary observation had its numeric exit code hidden by a pipe (pinned by tests instead). State both. `nyquist_compliant: true` is defensible only if each of DX-17, DX-18, EVD-13 has an automated row (it does, per the table); say "reconstructed, retroactive" in the body.

SECURITY.md: copy the `29-SECURITY.md` skeleton (`status: verified`, `threats_open: 0`, `asvs_level: 1`, Trust Boundaries, Threat Register, Accepted Risks Log, Notes, Audit Trail, Sign-Off). Sources are `28-REVIEW.md` (CR-01, WR-01), `28-REVIEW-DISPOSITION.md` (CR-01 fixed `eface8a`; WR-01 deferred as process; WR-02/03/04, IN-01/02 fixed via `28-REVIEW-FIX.md`) and `28-VERIFICATION.md`. Natural Phase 28 trust boundaries from that evidence: source text -> exit-class mapping (a source refusal must not be masked as a tool failure: WR-02 `errors.Join`), report path -> filesystem (stale/symlinked report removal, WR-03), relative path -> project-root resolution (G-28-3), entry pruning -> emitted C byte identity. No test parses `*-SECURITY.md` [VERIFIED: grep for `SECURITY.md` in `*_test.go` returned nothing], so the constraint is honesty, not format. Mark it "retroactive, no new tests".

## Disposition Documents (D-22)

- Create `33-CARRIED-DEBT-DISPOSITION.md` (not a `*-REVIEW-DISPOSITION.md`, so no gsd gate rewrites it).
- Archived `32.1-REVIEW-DISPOSITION.md` is generated by a gate that "keeps every row it can" and accepts only `open|fixed|skipped|deferred` [CITED: footer text of the file]. D-22 wants `covered`. Use `fixed` for WR-01, WR-02, IN-02, IN-03 and the kind pin, and `fixed` with a Source cell reading "Covered by TestPhase32PublicRouteAudit (no code change needed)" for IN-01, so the file stays inside the vocabulary its footer documents. No Go test parses these files [VERIFIED per D-22; grep confirmed no test reads REVIEW-DISPOSITION]. Update the `open: 5` count in its frontmatter to match.
- `32.1-VERIFICATION.md`'s `covered_digest`/`covered_files` are consumed only by the gsd-core tooling (`bin/lib/verification.cjs`), never by a Go test [VERIFIED: grep `covered_digest` over `*.go` found nothing]. Editing code listed in `covered_files` (e.g. `interp.go`, `codedigest.sh`) will make a gsd staleness check report 32.1 stale-by-successor. That is expected and already the state of phases 28-32 ("5 phases stale-by-successor", MILESTONES.md). Not a CI concern.

## Common Pitfalls

### Pitfall 1: label-length checks hard-coded to `phase32`
**What goes wrong:** `cut -c1-20` compared to `"phase32 host identit"` and `cut -c1-19` compared to `"phase32 host result"` silently stop matching any other label, or still match only M006.
**Why:** copy-edit of the old script with `phase32` replaced only in `printf` and `grep` patterns.
**Avoid:** replace with an anchored prefix test on the label: `first=$(grep "^${label} " "$log" | sed -n '1p'); case $first in "${label} host identity "*) ;; *) reject "identity_not_first_$log";; esac`. Same for last/result. Add a differential row (#6 in the catalog) and a **second-label** row: run the new gate with a variant manifest `label=m999` and logs rewritten to `m999 …` rows, expecting pass. This proves the label really is data.
**Warning sign:** the M006 differential passes but nothing exercises a different label.

### Pitfall 2: the old receipt gate resolved its root from `$0`, the new one from git
**What goes wrong:** with `EXPECTED_CODEDIGEST` set the old gate needed no git at all; the new one rejects `no_repo_root` outside a repository.
**Avoid:** accept it (D-07 is locked) but note it, and make the differential run both inside the project checkout. The new gate reads `scripts/codedigest.sh` via `sh "$root/scripts/codedigest.sh"` from the toplevel, so tests that copy the script into a temp git repo (the existing `phase32GateTempRepo` pattern) keep working.

### Pitfall 3: `RunCLIErr` has no cwd parameter
**What goes wrong:** the subdirectory test cannot just set `Dir`.
**Avoid:** use the existing pattern `testsupport.RunCLIErr(ctx, "sh", nil, "-c", `cd "$0/cmd" && sh ../scripts/check-freeze.sh "$@"`, repo, args…)` (as `subdirectory-equals-root` does for `codedigest.sh`). Use the temp-repo fixture for the CI-safe row and a real-repo row starting at `0aac0a7b…` as the extra witness.

### Pitfall 4: relative MANIFEST or log args after `cd "$root"`
**What goes wrong:** `cd` happens, then `[ -f "$1" ]` is evaluated against the new cwd.
**Avoid:** make a read-path copy first (`case $log in /*) abs=$log;; *) abs=$PWD/$log;; esac`) and keep `$log` untouched for the reason suffix (D-07). Test from a subdirectory with relative log paths.

### Pitfall 5: stderr pollution from `codedigest.sh --changed`
**What goes wrong:** a failing sub-invocation prints its own `codedigest: status=fail …` line, so the gate emits two stderr lines (breaks D-06).
**Avoid:** `changed=$(sh "$root/scripts/codedigest.sh" --changed "$freeze" "$candidate" 2>/dev/null) || reject git_log_failed`. Also add `-c core.quotePath=false` to the `git log --name-only` and `git diff --name-only` calls, as the listing already does, or a non-ASCII `*.md` name arrives quoted and ends in `"` instead of `.md` (conservative false reject).

### Pitfall 6: `session.go` line shifts
See Finding 3. The inventory test names the exact stale pin; run `go test ./internal/compiler/session -run 'TestPhase16PublicEmitterConsumerInventory' -count=1` after any `session.go` edit.

### Pitfall 7: parity test false-green because both scripts are told the same wrong thing
**What goes wrong:** the differential only proves the new script equals the old one; if a catalog row is missing, both agree on silence.
**Avoid:** every catalog reason in the table above gets a row that asserts the *expected reason string*, not only old==new. Then freeze the literal goldens (D-10 step 2) from the **expected** strings, with the temp dir normalized. A golden that was generated by the new script and never read by a human is the weak form.

### Pitfall 8: renaming the test functions the archived logs name
The archived logs record `tests=TestPhase32GuideCleanCheckout`, `TestPhase32PublicRouteAudit`, `TestPhase32CodeDigest`, `TestPhase32ReceiptLivingClaims`. The generic gate's control check only matches the names in the `control` rows and the producer runs them by regex, so renaming them breaks the producer, not the gate. Keep all four names.

### Pitfall 9: grading trap on new VALIDATION docs
See "Which meta-tests apply". Also keep `33-VALIDATION.md` at `status: planned` until the end.

### Pitfall 10: scope drift of the static no-caller test
See Finding 2.

### Pitfall 11: `TotalAlloc` noise in the memory assertion
See "Budget numbers".

## Code Examples

### Abort-like predicate shape (D-12)

```go
// Source: shape derived from internal/compiler/interp/interp.go:432-461 (current) and D-12.
// Seams follow internal/compiler/core/core.go:509 (decoderRequireNonEmpty).
var abortLikeAcceptLive = func(live []string, ownerOpID string) bool {
	return len(live) == 1 && live[0] == ownerOpID
}
var abortLikeReasonMatches = func(got, want string) bool { return got == want }

func isAbortLikeDefect(program core.Program, result Execution, ownerOpID string) bool {
	if result.Outcome.Kind != execution.OutcomeDefect || len(result.Events) == 0 {
		return false
	}
	if !abortLikeAcceptLive(result.LiveResources, ownerOpID) {
		return false
	}
	final := result.Events[len(result.Events)-1]
	if final.Kind != "function.defected" || !strings.HasSuffix(final.ID, ":event:defected") {
		return false
	}
	operationID := strings.TrimSuffix(final.ID, ":event:defected")
	function, ok := findFunctionByID(program, final.FunctionID) // interp.go:1502
	if !ok || function.Linear == nil {
		return false
	}
	for _, operation := range function.Linear.Operations {
		if operation.ID != operationID {
			continue
		}
		switch operation.Kind {
		case core.OpBytesAt:
			return abortLikeReasonMatches(final.Output, execution.DefectReasonByteIndexOutOfBounds)
		case core.OpAddChecked:
			return abortLikeReasonMatches(final.Output, execution.DefectReasonU64AddOverflow)
		case core.OpRemainderU64:
			return abortLikeReasonMatches(final.Output, execution.DefectReasonU64RemainderByZero)
		case core.OpDefect:
			return abortLikeReasonMatches(final.Output, operation.Reason)
		}
		return false
	}
	return false
}
```

The constant values must be exactly the current literals: `"U64 addition overflow"`, `"U64 remainder by zero"`, `"byte index out of bounds"` (quoted from interp.go:1705, 1731, 1754). Operation IDs are unique within a function (`%s:op:%d`, `corevalidate.go:1217` per CONTEXT), so a two-function falsifier with the same op ID is a valid refusal row: the op exists in function A but the event names function B.

### Freeze script skeleton

```sh
#!/bin/sh
# Generic freeze binding gate. Usage: check-freeze.sh MANIFEST FREEZE CANDIDATE [SAME_AS ...]
set -eu

manifest_arg=${1:-}
label=
[ -n "$manifest_arg" ] && [ -f "$manifest_arg" ] && label=$(sed -n 's/^label=//p' "$manifest_arg" | sed -n '1p')
reject() {
	printf '%s freeze gate: status=fail reason=%s\n' "$label" "$1" >&2   # label is data, never a format
	exit 1
}
```

Note the empty-label output has a leading space (`" freeze gate: …"`) if written this way; decide the zero-arg/empty-label rendering deliberately (suggest omitting the leading space when the label is empty).

### Differential test shape (Plan 3)

```go
// Source: pattern from internal/compiler/session/session_phase32_gate_test.go:117-135 and session_phase32_freeze_test.go:13-23
type gateRun struct{ exit int; stdout, stderr string }
func runBoth(t *testing.T, oldArgs, newArgs []string, env []string) (gateRun, gateRun) { /* RunCLIErr "sh" twice, same file paths */ }
// for each catalog row: old, new := runBoth(...); require old == new; require strings.Contains(new.stderr, "reason="+row.reason)
```

## State of the Art

| Old approach | Current approach | Why |
|---|---|---|
| per-phase gate scripts reading the producer script as the table | milestone-agnostic gate reading `testdata/receipts/<M>.manifest` | Phase 38 and later milestones call the same two names |
| `isAbortLikeDefect` accepting any live set | owner identity + closed per-kind reason | WR-02 |
| `os.ReadFile` on a child-written capture | limit+1 reader | QLT-13 |

**Deprecated after this phase:** `scripts/check-phase32-freeze.sh`, `scripts/check-phase32-receipts.sh`, the `# phase32-cases-begin/end` block and `phase32_programs=` line in the producer.

## Plan-Sizing and Ordering

The 5-plan shape in CONTEXT is feasible. Dependencies:

| Plan | Depends on | Digest-moving? | Sizing |
|---|---|---|---|
| 1. Reopen docs (T1), bounded read | T1 first, always | T1 no; read yes | small |
| 2. WR-02 pin, narrow, mutant, shared constants | 1 | yes (`interp`, `execution`, `cgen`, `executionpeer`, `session.go` in place) | medium; touches the most packages. Parity pin rows first |
| 3. Gate step 1 | 1 | yes (`scripts/`, `testdata/`) | **large**: codedigest flags, manifest, two scripts, differential catalog (~35 rows), archived-M006 test with 6 armed controls, freeze rows, subdirectory and `.md` tests |
| 4. Gate step 2 | 3 | yes | **large**: goldens, delete scripts, repoint producer, static test, 2 test-file renames, 28-row living-claim generalization, doc-set extension |
| 5. Debt closure | 1 (docs), 2-4 for the final run | IN-02 and kind pin touch `cmd/` and `internal/` | medium, mostly docs |

Plans 2 and 3 are independent of each other (disjoint files) and can run in the same wave after Plan 1; Plan 4 needs Plan 3; Plan 5's test edits are independent, but its disposition rows need 2-4's results, so run it last. To lighten Plan 4, move the D-11 living-claim generalization and the PROJECT/MILESTONES doc-set extension into Plan 5 (they do not depend on deleting the scripts). Within the cap of 5 this keeps each plan at a coherent size.

Digest-path ordering hazards:
- Plan 1 Task 1 must be a separate, earlier commit than anything under the digest paths. Verify with `git diff --name-only HEAD~1 HEAD | grep -v '^\.planning/'` printing nothing.
- `.planning/phases/33-*/` PLAN/SUMMARY files are `.md` under `.planning/`, never in the digest.
- Any worktree executor (`use_worktrees: true`) must branch from a commit that already contains Plan 1 Task 1.
- Every plan's last step runs the session meta-tests (memory note: run them locally before hosted dispatch): `TestPhase16PublicEmitterConsumerInventory`, `TestVerificationGroundedness`, `TestPhase20ValidationLifecycle`, `TestValidationRowGradesAreEarnedOverArchivedCorpus`, `TestSummaryCoverageGroundedness`, `TestPhase32ReceiptLivingClaims`, `TestPhase32CodeDigest`.

## Runtime State Inventory

This is not a rename migration, but it retires named artifacts; checked explicitly:

| Category | Items found | Action |
|---|---|---|
| Stored data | None. The archived receipts and logs are files under `.planning/milestones/` and are immutable (D-02) | none |
| Live service config | `.github/workflows/ci.yml` runs `scripts/verify-phase32.sh` (line 133-137) and uploads `phase32-receipt-*.log`; it has no `check-phase32-*` caller [VERIFIED: grep]. `TestPhase32CIHostMatrix` pins the run and upload steps | keep the `verify-phase32.sh` step and artifact names untouched |
| OS-registered state | None | none |
| Secrets/env vars | `PHASE32_EXPECTED_CODEDIGEST` is read only by the old receipt gate and set by `runPhase32Gate` and one subtest in `session_phase32_gate_test.go` [VERIFIED: grep] | replaced by `EXPECTED_CODEDIGEST` in the new gate and tests; no CI secret involved |
| Build artifacts | none | none |

## Environment Availability

| Dependency | Required by | Available | Version | Fallback |
|---|---|---|---|---|
| go | all tests | yes | 1.24.0 darwin/arm64 | none needed |
| git | gate scripts, fixtures | yes | repo has full history; 2213 commits; not shallow | n/a |
| clang | native parity, capture integration | yes | Apple clang 21.0.0 | tests already `t.Skip` via `phase29RequireClang` when absent; hosted CI sets `SCHWAY_REQUIRE_NATIVE=1` |
| sh | gate scripts | yes (bash 3.2 in POSIX mode locally; dash on Ubuntu CI) | n/a | bashisms would pass locally and fail on Ubuntu: avoid `[[`, `local`, `${var//}`, arrays, `echo -e` |
| shellcheck | local lint only | yes, not in CI | n/a | `sh -n script` |
| `dd`, `head`, `tr` | sparse-file child script | yes | n/a | `/usr/bin/truncate` also present |
| Python / zlib oracle | not needed | n/a | n/a | n/a |

**Missing dependencies with no fallback:** none.

## Validation Architecture

Nyquist validation is enabled (`workflow.nyquist_validation: true` in `.planning/config.json`). Per STATE.md ("Do not run project suites locally"), every command below is a focused `go test -run` or a static check. Test names for new tests are the recommended names from CONTEXT; the planner may refine.

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go `testing`, go1.24.0 |
| Config file | none (go.mod only) |
| Quick run command | per-requirement commands below |
| Full suite command | hosted CI only (never local) |

### Phase Requirements → Test Map

| Req | Success criterion | Behavior | Type | Automated command | File exists? |
|-----|-------------------|----------|------|-------------------|--------------|
| QLT-13 | SC1 | exactly-limit completes; limit+1 is `native.evidence_capacity_exhausted`; sparse 1 GiB file allocates < 16*limit | integration | `go test ./internal/compiler/native -run 'TestPhase33CaptureReadBounded' -count=1` | Wave 0 (new) |
| QLT-13 | SC1 | existing capture controls unchanged (missing, partial, capacity, exact limit) | integration | `go test ./internal/compiler/native -run 'TestPhase22EvidenceMissingPartialAndCapacityControls|TestPhase27OutputEvidenceExactLimit|TestPhase27EvidenceCapacity|TestPhase26FailureChannels' -count=1` | exists |
| QLT-13 | SC1 | seeded `os.ReadFile` mutant caught | manual-once | hand-run: restore `os.ReadFile(capturePath)` at `native_app.go:364`, run the command above, paste the failing output in the SUMMARY, `git checkout` the file | n/a (D-16) |
| QLT-14 (WR-01) | SC2 | gate verdict identical from root and from `cmd/` | integration | `go test ./internal/compiler/session -run 'TestFreezeGate/subdirectory' -count=1` | Wave 0 (new subtest) |
| QLT-14 (WR-02) | SC2 | pinned parity rows (4 kinds x fixture + file routes) | integration | `go test ./internal/compiler/native -run 'TestPhase321DefectParityWithLiveOwner|TestPhase33DefectParityFileRoute' -count=1` | partly exists; file route Wave 0 |
| QLT-14 (WR-02) | SC2 | predicate accepts only named reason + live owner; refuses wrong reason, extra live, missing owner, wrong function, same-op-ID falsifier | unit | `go test ./internal/compiler/interp -run 'TestPhase321AbortLikeDefectPredicate|TestPhase321AbortLikeDefectFileRoute|TestPhase33AbortLike' -count=1` | exists / Wave 0 |
| QLT-14 (WR-02) | SC2 | widening mutant killed | unit | `go test ./internal/compiler/interp -run 'TestPhase33AbortLikeWideningMutantKilled' -count=1` | Wave 0 |
| QLT-14 (WR-02) | SC2 | emitted C unchanged by the constants refactor | integration | `go test ./internal/compiler/cgen ./internal/compiler/session -run 'TestPhase272AcceptedCorpusByteIdentity|TestPhase273ConformanceCorpusByteIdentity|TestPhase28PrunedEmitMatchesGolden|TestEmittedCExitLiteralsMatchTable' -count=1` | exists |
| QLT-14 (WR-02) | SC2 | executionpeer still checks the reason | unit | `go test ./internal/compiler/executionpeer -run 'Phase29|Phase321' -count=1` | exists |
| QLT-14 (IN-01) | SC3 | route audit covers drift | unit | `go test ./cmd/schway -run 'TestPhase32PublicRouteAudit' -count=1` | exists |
| QLT-14 (IN-02) | SC3 | marker and `dir` absent, control present | unit | `go test ./cmd/schway -run 'TestPhase321RunInputRequiredCLI' -count=1` | exists, edited |
| QLT-14 (IN-03) | SC3 | `.md`-only change moves neither digest nor freeze verdict; `--changed` empty | integration | `go test ./internal/compiler/session -run 'TestPhase32CodeDigest|TestFreezeGate/md' -count=1` | exists / Wave 0 |
| QLT-14 (kind) | SC3 | unknown kind rejected | unit | `go test ./internal/compiler/session -run 'TestSummaryCoverageGroundedness' -count=1` | exists, edited |
| QLT-15 | SC3 | Phase 29 WR-02/WR-03 witnesses green | unit/integration | `go test ./internal/compiler/check ./internal/compiler/native -run 'TestPhase321ByteViewEntryRefused|TestPhase321DefectParityWithLiveOwner' -count=1` | exists |
| QLT-15 | SC3 | Phase 28 VALIDATION not draft; commands ground to real tests | meta | `go test ./internal/compiler/session -run 'TestPhase20ValidationLifecycle|TestVerificationGroundedness|TestValidationRowGradesAreEarnedOverArchivedCorpus|TestStaticTestIndexMatchesGoTestList' -count=1` | exists |
| EVD-16 | SC4 | old==new for every catalog row; new rows reject with the expected reason | integration | `go test ./internal/compiler/session -run 'TestReceiptGateDifferential|TestFreezeGateDifferential' -count=1` | Wave 0 (step 1 only; replaced by goldens in step 2) |
| EVD-16 | SC4 | archived M006 pair + `32-RECEIPT.json` digest passes; six armed controls reject | integration | `go test ./internal/compiler/session -run 'TestReceiptGateArchivedM006' -count=1` | Wave 0 |
| EVD-16 | SC4 | M006 freeze rows (`0aac0a7b`->`9f8f878f` accept; post-freeze reject) | integration | `go test ./internal/compiler/session -run 'TestFreezeGate/m006' -count=1` | Wave 0 |
| EVD-16 | SC4 | manifest sha256 pinned; rows appear in both archived logs; manifest equals producer table | unit | `go test ./internal/compiler/session -run 'TestPhase32ProducerCaseTable' -count=1` | exists, retargeted |
| EVD-16 | SC4 | second label proves label is data | integration | `go test ./internal/compiler/session -run 'TestReceiptGate/second-label' -count=1` | Wave 0 |
| EVD-16 | SC4 | no `check-phase32-*` caller; both files absent | static | `go test ./internal/compiler/session -run 'TestNoCheckPhase32Caller' -count=1` | Wave 0 (step 2) |
| EVD-16 | SC4 | living-claim rows (28 ported + real-checkout drifted-digest row) | unit | `go test ./internal/compiler/session -run 'TestPhase32ReceiptLivingClaims' -count=1` | exists, extended |
| EVD-16 | SC4 | digest-path agreement `--paths` vs Go twin | unit | `go test ./internal/compiler/session -run 'TestPhase32CodeDigest|TestFreezeGate/paths-match-digest' -count=1` | exists / retargeted |
| (guard) | all | emitter inventory still pinned | meta | `go test ./internal/compiler/session -run 'TestPhase16PublicEmitterConsumerInventory' -count=1` | exists |
| (guard) | all | CI wiring untouched | unit | `go test ./internal/compiler/session -run 'TestPhase32CIHostMatrix|TestPhase32HostedCIEnv' -count=1` | exists |
| (guard) | all | living docs still pass the Phase 25 word check | unit | `go test ./internal/compiler/native -run 'TestPhase25LivingRoadmap' -count=1` | exists |

Static checks (no Go test): `git diff --stat -- go.mod go.sum` is empty; `sh -n scripts/check-freeze.sh scripts/check-receipts.sh scripts/codedigest.sh scripts/verify-phase32.sh`; optional `shellcheck -s sh` on the same four; `git grep -n 'check-phase32' -- . ':!.planning'` returns nothing after Plan 4.

### Sampling Rate
- **Per task commit:** the focused command(s) for the touched package, plus `TestPhase16PublicEmitterConsumerInventory` whenever `session.go` or an emitter consumer changed.
- **Per wave merge:** all rows for the wave's packages, plus the session meta-test set above.
- **Phase gate:** every row above green, run locally as focused commands; full suites are for hosted CI only. The phase binds no hosted receipt (D-03).

### Wave 0 Gaps
- [ ] `internal/compiler/native/native_phase33_capture_test.go` — QLT-13
- [ ] `internal/compiler/interp/interp_phase33_defect_test.go` — WR-02 narrowing and mutant
- [ ] file-route rows in `native_phase321_defect_parity_test.go` (or a new `native_phase33_defect_parity_test.go`) — WR-02 pin
- [ ] `testdata/receipts/M006.manifest`, `scripts/check-freeze.sh`, `scripts/check-receipts.sh` — EVD-16
- [ ] differential tests and `TestReceiptGateArchivedM006` — EVD-16
- [ ] `TestNoCheckPhase32Caller` — EVD-16
- No framework install needed.

## Security Domain

`security_enforcement` is enabled (ASVS level 1, `security_block_on: high`).

### Applicable ASVS Categories

| ASVS category | Applies | Standard control |
|---|---|---|
| V2 Authentication | no | n/a |
| V3 Session Management | no | n/a |
| V4 Access Control | no | n/a |
| V5 Input Validation | yes | bounded reader for the untrusted capture file; strict manifest key grammar; fixed-position `printf '%s'`; no `eval` |
| V6 Cryptography | no (sha256 via `shasum`/`sha256sum`, existing; not hand-rolled) | existing hasher selection in `codedigest.sh` |
| V12 Files and Resources | yes | open-then-limited-read (no stat-by-path) |

### Known Threat Patterns

| Pattern | STRIDE | Standard mitigation |
|---|---|---|
| Child writes a huge capture file to force an unbounded host allocation (CWE-770, CWE-400) | Denial of service | `LimitReader(file, limit+1)` before any allocation; behavioral allocation test |
| TOCTOU between size check and read (CWE-367) | Tampering | no `Stat`; read from the opened handle |
| Command or format injection through manifest fields (`label`, `programs`, case rows) | Tampering / EoP | manifest is repo-controlled and digest-bound, but still: `printf '%s'` formats, quoted expansions, no `eval`, `grep -F` where a literal is intended |
| Gate evaluated from the wrong directory gives a false accept (CWE-693) | Spoofing / Repudiation | resolve toplevel, `cd`, reject `no_repo_root` (WR-01, reproduced live) |
| Gate accepts a receipt for different code | Spoofing | digest binding retained; `EXPECTED_CODEDIGEST` never recomputed when set, `digest_source` printed. The gates check consistency, not authenticity; the header comment of the old script says so and the new ones should keep it |
| Interpreter hides a leak by accepting a non-owner live resource | Repudiation | D-13 refusal + seam-killed test |
| Retroactive docs overclaim | Repudiation | Phase 28 docs state "retroactive, reconstructed, no new tests" and the `Hosted CI: not run` fact |

## Assumptions Log

| # | Claim | Section | Risk if wrong |
|---|-------|---------|---------------|
| A1 | sha256 of empty input is the standard `e3b0c442…b855` and both `shasum -a 256` and `sha256sum` print 64 hex chars for it | Exact Current Behavior (codedigest) | the derived-length check would reject every digest; planner's test asserts derived length equals the real digest length, so it is caught immediately |
| A2 | M007 will use schema string `schway.m007-receipt/1` | Living-Claim Check | Phase 38 would parameterize differently; low, Phase 38 owns it |
| A3 | Non-capture allocation of one `RunApplicationWithEvidence` call is well under 1 MiB | Bounded Capture Read | a spurious failure on correct code; mitigated by measuring once and recording the number |
| A4 | No test pins the exact "live local resources" error text for the new D-13 refusal | Defect Parity | an unexpected red test; mitigate by `grep -rn 'live local resources' --include='*_test.go'` before editing |

## Open Questions (RESOLVED)

1. **Zero-argument label fallback for `check-freeze.sh` / `check-receipts.sh`.**
   - Known: with no manifest arg there is no label; the old scripts always printed `phase32 …`.
   - Unclear: whether an empty-label line (`freeze gate: status=fail reason=usage`) or a fixed word (`gate`) is preferred.
   - Recommendation: empty label, no leading space; it has no M006 parity obligation. Pin it with one golden.
   - RESOLVED: empty label, no leading space (`freeze gate: status=fail reason=usage`, and `manifest_missing` renders the same way); pinned by golden in plan 33-03 Task 1.

2. **Where `33-CARRIED-DEBT-DISPOSITION.md` marks IN-01 `covered`.**
   - The archived 32.1 file's documented vocabulary lacks `covered`. Recommendation above uses `fixed` plus a Source cell. If the user prefers the literal word `covered`, it is harmless because no test parses the file, but the generating gate may reject it on a re-run.
   - RESOLVED: IN-01 is recorded as `fixed` with Source `Covered by TestPhase32PublicRouteAudit (no code change needed)`, in the archived file's `open|fixed|skipped|deferred` vocabulary; plan 33-05 Task 3.

3. **Real-repo freeze rows and branch history.** The pinned SHAs are reachable from tag `vM006` and from `phase26-validation`, not from `main`. Recommendation: fail loudly (not skip) with the `git fetch --tags` remedy, and keep fixture-repo twins of every row.
   - RESOLVED: real-repo freeze rows fail loudly (t.Fatalf naming the SHA and the `git fetch --tags` remedy), never skip, and every row has a fixture-repo twin; plan 33-03 Task 1.

## Project Constraints (from CLAUDE.md and AGENTS.md)

- `CLAUDE.md` defers to `AGENTS.md` and routes spike findings for ownership semantics, interprocedural loan liveness, certificate evidence, native FFI/ABI and differential-harness patterns to `Skill("spike-findings-ai-lang")`. This phase touches interpreter/native parity (WR-02), so the executor should load that skill before editing `interp.go`/`cgen`.
- `AGENTS.md`: Go 1.24 stdlib hosts Stage 0; "Dependencies: prefer standard-library-only and shallow audited boundaries". No new modules.
- Any repo edit must start through a GSD command (`$gsd-execute-phase` for planned work).
- At phase transitions: read `.planning/PRODUCT-ROADMAP.md` and `.planning/LANGUAGE-MATURITY.md`, compare claims with source witnesses, keep "source inspection / newly executed / historical receipt" distinct, and update living docs inside the workflow. Phase 33's D-01 reopen is exactly that obligation.
- Prefer a runnable gain per feature phase; a second consecutive enabling phase triggers scope review. Phase 33 is explicitly a debt phase (capped at 5 plans) and the roadmap already records that.
- Memory notes that apply: run the planning-doc meta-tests locally before hosted dispatch (inventory, groundedness, validation lifecycle); end each GSD step with the exact next command; automate verify-work checkpoints.
- STATE.md: do not run project suites locally.

## Sources

### Primary (HIGH confidence): files read this session
- `scripts/check-phase32-freeze.sh`, `scripts/check-phase32-receipts.sh`, `scripts/codedigest.sh`, `scripts/verify-phase32.sh` (whole files)
- `internal/compiler/session/session_phase32_{freeze,gate,receipt,ci,retire}_test.go`
- `internal/compiler/interp/interp.go` (lines 255-470, 715-730, 1502-1509, 1537-1548, 1690-1775), `interp_phase321_defect_test.go`
- `internal/compiler/native/native_app.go` (80-115, 340-400, 560-590, 630-700), `native_app_test.go` (455-700), `native_phase321_defect_parity_test.go`, `native_phase31_controls_test.go` (20-150)
- `internal/compiler/{executionpeer/executionpeer.go:375-440, cgen/cgen_program.go:215-250 and 2700-2765, session/session.go:515-550, core/core.go:495-512, core/core_internal_test.go}`
- `internal/compiler/session/{summary_coverage_test.go, verification_groundedness_test.go:1040-1100 and 1517-1560, evidence_grade_test.go:1037-1130 and 1380-1460 and 2282-2340, session_test.go:2579-2589}`; `testdata/phase16/public-emitter-consumers.json`
- `.planning/milestones/M006-phases/{28,29,32,32.1}-*/` documents named in the sections above; `32-RECEIPT.json`; archived logs (via running the old gate)
- `.planning/{PROJECT,MILESTONES,PRODUCT-ROADMAP,LANGUAGE-MATURITY,REQUIREMENTS,ROADMAP,STATE}.md` (the cited lines); `.planning/config.json`; `.github/workflows/ci.yml`
- `~/.claude/gsd-core/bin/lib/coverage.cjs:60-70` (`VALID_KINDS`)
- Executed this session: old freeze/receipt gates on the M006 pairs and on the four pinned SHAs, from root and from `cmd/`; `git grep check-phase32`; `go build ./cmd/schway` plus `schway --json run` on the IN-02 fixture; `dd` sparse-file creation; `git show 31d71e94`

### Secondary / Tertiary
- none. No web search was used; this phase needs no external documentation.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH, no external packages, every tool probed.
- Architecture: HIGH, all four target behaviors were read and the old gates were executed.
- Pitfalls: HIGH for items 1-10 (each tied to a read line or executed command); MEDIUM for the `TotalAlloc` noise estimate (A3).

**Research date:** 2026-10-13
**Valid until:** 2026-11-12 (stable; invalidated early by any commit touching the cited scripts, `interp.go`, `native_app.go`, or `session.go` line positions)
