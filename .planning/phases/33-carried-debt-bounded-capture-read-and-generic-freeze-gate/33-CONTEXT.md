# Phase 33: Carried Debt, Bounded Capture Read, and Generic Freeze Gate - Context

**Gathered:** 2026-10-13
**Status:** Ready for planning
**Mode:** Fan-out advisor research (4 parallel multi-lens researchers + adversarial pass), recommendations auto-followed per the user's standing instruction (memory: fan-out-then-auto-follow). The user can still override any decision below before planning.

<domain>
## Phase Boundary

Pay the carried M006 debt before any M007 feature work (QLT-13, QLT-14, QLT-15, EVD-16):
- the native app reads the child-written evidence capture through a limit+1 bounded reader;
- 32.1 WR-01/WR-02/IN-01..03, the coverage-`kind` pin, and Phase 29 WR-02/WR-03 are fixed with a test or dispositioned in writing;
- Phase 28 gains VALIDATION.md and SECURITY.md;
- one milestone-agnostic freeze/receipt gate replaces `scripts/check-phase32-{freeze,receipts}.sh`.

Capped at 5 plans. It binds no hosted receipt. No new language features and no new dependencies.

</domain>

<decisions>
## Implementation Decisions

### A. M006 receipt reopen (D-10a), which comes first
- **D-01:** The first commit of the phase is a docs-only commit, Plan 1 Task 1, made before any digest-path commit. It flips both machine-checked markers, `.planning/LANGUAGE-MATURITY.md:239` and `.planning/PRODUCT-ROADMAP.md:10`, to `M006 final hosted receipts: reopened 2026-10-xx (historical codedigest b9253457705a224a0580d74896e8e5bd3821446a1d467626b7d611ca0d1c4b5c, run 38335175306)`. Use the actual date and match the format `phase32ReopenedMarker` accepts. Rewrite the prose claims of current evidence in PROJECT.md (~171, ~609-610), MILESTONES.md (~19-20) and STATE.md in the Phase 25 "historical receipts" form. Without this commit, `TestPhase32ReceiptLivingClaims` (`stale_bound_claim`) turns CI red on the first code commit. — **Reversibility:** reversible
- **D-02:** Archived receipt artifacts are immutable. Do not modify `32-RECEIPT.json`, `32-RECEIPT.run-38315799626.json` or `receipts/*.log`. The reopen is recorded only in the living docs.
- **D-03:** No hosted re-take in Phase 33. The docs use this wording: "Phase 33 binds no hosted receipt. M006 receipts are historical; current hosted evidence resumes in Phase 38."

### B. Generic freeze/receipt gate (EVD-16, plus 32.1 WR-01 and IN-03)
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

### C. Narrowing `isAbortLikeDefect` (32.1 WR-02)
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

### D. Bounded capture read (QLT-13)
- **D-15:** Extract `readApplicationCapture(path string, limit int) ([]byte, error)` in `internal/compiler/native/native_app.go`, modelled on `readApplicationReceipt` (~649). It does `os.Open`, then `io.ReadAll(io.LimitReader(f, int64(limit)+1))`, and `len > limit` gives `EvidenceStatusCapacityExhausted`, which ends as `native.evidence_capacity_exhausted`. There is **no Stat-by-path pre-check**: the child writes the file, so checking then reading by path is a TOCTOU race.
- **D-16:** `TestPhase33CaptureReadBounded` runs without `t.Parallel` and injects a small limit through `Runner.evidenceLimit`. It has three cases:
  - **exactly-limit:** a valid capture padded with JSON whitespace to exactly `limit` bytes must not be capacity exhausted;
  - **limit+1:** must be capacity exhausted;
  - **mutant catcher:** a sparse 1 GiB file (`Truncate(1<<30)`) must give the capacity outcome with a `runtime.ReadMemStats` `TotalAlloc` delta under `16*limit`. The behavior catches the `os.ReadFile` mutant, so a grep cannot be beaten by an alias.

  Run the mutant once by hand and record the failing output in the SUMMARY.

### E. Small debt, fix or disposition (QLT-14 remainder, QLT-15)
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

### Suggested plan shape (planner may adjust within cap 5)
1. Reopen the docs (D-01..03), then the bounded capture read (D-15..16).
2. WR-02 pin, narrow and mutant, plus the shared defect-reason constants (D-12..14).
3. Generic gate step 1: codedigest flags, manifest, new scripts, differential and archived-M006 tests, WR-01 and IN-03 (D-04..08, D-10 step 1).
4. Gate step 2: goldens, delete the old scripts, repoint the producer, the static no-caller test, test renames and living-claim generalization (D-09..11).
5. Debt closure: IN-02, the kind pin, Phase 29 verify, the Phase 28 backfill, the disposition docs (D-17..22).

### Claude's Discretion
- The exact reopened-marker date string and the prose wording, within D-01's form.
- The exact names of the defect-reason constants, as long as the emitted C stays byte-identical.
- The manifest's line syntax (key=value plus verbatim `# case` rows) and the temp-dir normalization token in the goldens.
- Whether the gate's `.md`-only Go test lives in the freeze-gate test file or a codedigest test.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase scope
- `.planning/ROADMAP.md` §Phase 33: goal, SC1-4, research flag, 5-plan cap
- `.planning/REQUIREMENTS.md`: QLT-13, QLT-14, QLT-15, EVD-16
- `.planning/research/M007/SUMMARY.md`: M007 synthesis (LimitReader, ≤5 plans, no receipt binding)

### Review findings being closed
- `.planning/milestones/M006-phases/32.1-address-tech-debt-cli-exit-classes-and-module-aware-routes/32.1-REVIEW.md`: WR-01, WR-02, IN-01..03
- `.planning/milestones/M006-phases/32.1-address-tech-debt-cli-exit-classes-and-module-aware-routes/32.1-REVIEW-DISPOSITION.md`: currently all `open`
- `.planning/milestones/M006-phases/32.1-address-tech-debt-cli-exit-classes-and-module-aware-routes/32.1-VERIFICATION.md` (~line 44): origin of the coverage-`kind` pin
- `.planning/milestones/M006-phases/32.1-address-tech-debt-cli-exit-classes-and-module-aware-routes/32.1-CONTEXT.md`: D-09, D-25 (predicate widening rationale)
- `.planning/milestones/M006-phases/29-byte-views-and-the-adler-32-library/29-REVIEW.md` and `29-REVIEW-DISPOSITION.md`: WR-02/WR-03, fixed in 32.1
- `.planning/todos/completed/2026-10-12-phase32-review-hardening-after-m006.md`: earlier 32-REVIEW hardening record

### Receipts and freeze semantics
- `.planning/milestones/M006-phases/32-clean-checkout-guide-and-final-receipts/32-CONTEXT.md`: D-08 (revision = codedigest), D-10a (reopen semantics)
- `.planning/milestones/M006-phases/32-clean-checkout-guide-and-final-receipts/32-RECEIPT.json`: M006 digest and log sha256 (immutable)
- `.planning/milestones/M006-phases/32-clean-checkout-guide-and-final-receipts/receipts/receipt-{macos,ubuntu}-latest.log`: archived logs the generic gate must accept
- `.planning/LANGUAGE-MATURITY.md:239`, `.planning/PRODUCT-ROADMAP.md:10`: machine-checked bound markers
- `.planning/MILESTONES.md`, `.planning/PROJECT.md`, `.planning/STATE.md`: prose claims and the remaining-debt lines
- `scripts/check-phase32-freeze.sh`, `scripts/check-phase32-receipts.sh`, `scripts/verify-phase32.sh`, `scripts/codedigest.sh`

### Phase 28 backfill format
- `.planning/milestones/M006-phases/29-byte-views-and-the-adler-32-library/29-VALIDATION.md`, `29-SECURITY.md`: format to copy
- `.planning/milestones/M006-phases/28-exit-classes-verify-report-and-entry-pruning/`: 28-REVIEW, 28-REVIEW-DISPOSITION, 28-VERIFICATION, 28-0x-SUMMARY (evidence source)

### Project patterns
- `AGENTS.md`: project instructions
- `Skill("spike-findings-ai-lang")`: ownership kernel and interp/native parity semantics

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `readApplicationReceipt` (`internal/compiler/native/native_app.go:649`): the bounded-open pattern for `readApplicationCapture`. The same LimitReader idiom appears at `bindings.go:79`, `session.go:1951` and in `modload`.
- `Runner.evidenceLimit` / `evidenceCapacityLimit()`: the injection point for a small test limit.
- `decoderRequireNonEmpty` seam plus `TestDecodeInterfaceRequiredModeMutationKilled` (`internal/compiler/core/core_internal_test.go`): the repo's seeded-mutant pattern.
- `TestPhase32PublicRouteAudit` (`cmd/schway/main_phase32_routes_test.go`): already covers IN-01 drift.
- `summaryCoverageProblems` (`internal/compiler/session/summary_coverage_test.go`): where the `kind` enum goes.
- `phase32BoundMarker` / `phase32ReopenedMarker` regexes plus `TestPhase32ReceiptLivingClaims` (`session_phase32_receipt_test.go` ~289, ~442).

### Established Patterns
- Gates print one `status=fail reason=<name>` line and exit non-zero. They check consistency, not authenticity; authenticity comes from the hosted run record.
- Operation IDs have the format `%s:op:%d` and are unique within a function (`corevalidate.go:1217`).
- Defect reason strings are currently copied by hand across interp (~1705/1731/1754), `executionpeer.go` (~392/409/426) and `cgen_program.go` (~233-243, ~2722-2757).
- Do not run full project suites locally (STATE). Use focused `go test -run` only.

### Integration Points
- `isAbortLikeDefect` callers: `interp.go` ~319 (`RunWithFixtureBytes`) and ~371 (`RunWithFileBytes`).
- Capture read site: `native_app.go` ~364.
- `ci.yml` does not call the `check-phase32-*` scripts; only the Go tests do.

</code_context>

<specifics>
## Specific Ideas

- The user's dependency stance: "another copy-paste is better than another dep". The manifest reader is copied into both scripts and the kind enum is copied into Go. No new modules.
- Every "fixed" disposition must name a test that exists. A disposition never stands in for a cheap fix (the anti-pattern of laundering debt).
- Retroactive docs, such as the Phase 28 backfill, must say they are retroactive and must not overclaim.

</specifics>

<deferred>
## Deferred Ideas

- `check-phase27-receipts.sh` and other older phase-specific gates are out of scope; the static test covers only phase32. A later cleanup could migrate them to the generic gate.
- Moving the test seams for Phase 29/30 out of production packages (29 WR-04, listed in the M006 audit) is not in this phase.
- `evidence`/`interface`/`query` module-awareness stays a named refusal and is not in scope.

</deferred>

---

*Phase: 33-carried-debt-bounded-capture-read-and-generic-freeze-gate*
*Context gathered: 2026-10-13*
