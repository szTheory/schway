# Phase 32: Clean-Checkout Guide and Final Receipts - Context

**Gathered:** 2026-10-12
**Status:** Ready for planning

<domain>
## Phase Boundary

This is the last phase of M006. It delivers three things:

1. **The guide (DX-19, SC1).** A clean-checkout guide that a test executes. It covers `checksum`, the second consumer `adler_low` and the `adler32` library. It states the expected results, the 65,536-byte limit and the error outcomes.
2. **Frontier retirement (DX-19, SC2).** The provisional refused frontier `examples/checksum.schway` is retired, with a dated amendment.
3. **Final receipts (EVD-15, SC3).** Hosted macOS and Linux receipts are taken at the final M006 code revision. They cover the M006 programs and the carried-forward `sum_to_n` and FizzBuzz programs. A gate rejects a receipt whose code revision differs from the checked-out source.

The phase adds no new language or CLI capability.

</domain>

<decisions>
## Implementation Decisions

Every decision below came from advisor research. There was one `gsd-advisor-researcher` per gray area at the `minimal_decisive` tier, and each recommendation was **auto-followed** under the user's standing fan-out boilerplate, which was re-pasted this session. A cross-area coherence pass made two adjustments, marked **[synthesis]**.

### Phase ordering (cross-cutting)
- **D-00 [synthesis]:** Every code change lands *before* the receipt freeze. The plan order is:
  1. the formatter fix (D-10);
  2. frontier retirement (D-05..D-07) and the guide doc-test (D-01..D-04), which can run in parallel;
  3. the public-route audit and gate changes (D-08), whose final code commit is the receipt target;
  4. a human checkpoint plan with the single push and dispatch line;
  5. a `.planning`-only receipt-ingest plan, which is the last plan of the milestone.

  If the hosted run fails or any code changes after step 3, the ingest plan stops and routes back to step 3.

### Guide and doc-test (DX-19, SC1)
- **D-01:** The guide stays in `examples/checksum/README.md`, the existing home, following the `examples/phase27/README.md` precedent. It is the *executable spec*.
  - Its runnable blocks are marked fences (for example ` ```sh doc-test=clean-checkout `), holding literal lines starting with `go build -o ./schway ./cmd/schway`, then `./schway build … --manifest … --output …` and `./schway app run <artifact> -- PATH` for `checksum`, `adler_low` and the `adler32_cli` library demo.
  - A results table has stable row keys equal to the case names in `checksum.expected.json`.
  - The test holds **no commands or numbers of its own**: it reads them from the README and cross-checks them against `checksum.expected.json`, including `"limit": 65536`.

  Rejected:
  - a test-side mirror table, which can pass while the guide is wrong;
  - `testscript` or txtar, which is a dependency and inverts the doc-test;
  - a new top-level guide, which duplicates claims.
- **D-02:** Clean checkout is simulated by copying `git ls-files -co --exclude-standard` into `t.TempDir()`, falling back to a walk that skips `.git` and `.schway`.
  - The test then runs the guide lines with `sh -c` in that copy, using a fresh `./schway` built from the copy, and substitutes only the `PATH` token.
  - Each subprocess gets a context timeout, and the test does not use `t.Parallel`.
  - The test lives in `cmd/schway/main_phase32_guide_test.go`.

  Rejected: `git archive HEAD` and `git worktree`. Both test the last commit rather than the edited guide, and the worktree route also mutates repo state.
- **D-03:** The guide's error rows cover only deterministic cases:
  - missing, under a fresh tempdir;
  - 65,537 bytes of 0xFF, giving `FileError.FileTooLarge`;
  - a directory, giving `FileError.FileUnavailable`;
  - plus one prose row: "other unreadable inputs give `FileError.FileUnavailable`, exit 65".

  Every error row asserts exit 65, empty stdout, the fixed stderr, and that the temp path segment is absent. FIFO, mode-000, `/dev/null`, EIO and AllocFailed stay with the existing Phase 31 matrix tests. Mode-000 is unreliable as root on CI.
- **D-04:** Anti-vacuity and armed controls for the doc-test:
  - Assert at least N marked blocks, that all three programs appear, and that the number of rows executed equals the number of table rows.
  - The negative controls run the checker on an in-memory mutated guide, once with one expected value changed and once with one command renamed. Each must fail.
  - On hosted CI, a missing clang is a failure, not a skip. The exact env mechanism is the planner's choice; see Claude's Discretion.
  - Test names: `TestPhase32GuideCleanCheckout`, `TestPhase32GuideParserNegativeControls` and `TestPhase32GuideMatchesExpectedJSON`.

  The guide gets an "Earlier programs" pointer to `examples/phase27/README.md` for `sum_to_n` and FizzBuzz, which keep their own doc pins. The guide is not expanded to cover them.

### Frontier retirement (DX-19, SC2)
- **D-05:** **Delete** `examples/checksum.schway`. Its tombstone SHA-256 is `ae95e550df67114c3464dda9cc24779f8a2efd69bd2b99dd0ccb2acc235cc154`, and its last pinned refusal was `syntax.expected_rbrace` at 521-527.
  - Replace `TestPhase20ChecksumFrontier` with `TestPhase32ChecksumFrontierRetired` in `internal/compiler/session/session_phase32_retire_test.go`. It asserts:
    - the old path is `ErrNotExist`;
    - `examples/checksum/checksum.schway` exists;
    - no `.schway` file under `examples/` or `testdata/` has the tombstone digest, which catches a renamed copy;
    - an armed control: the old bytes in a temp dir are flagged;
    - a doc-hygiene scan of the *live* docs only (PROJECT.md, LANGUAGE-MATURITY.md, PRODUCT-ROADMAP.md, `examples/checksum/README.md`) finds no line naming the old path without "retired" or "amendment".

  Rejected:
  - moving the file to `testdata/` or a history directory, because it stays a live corpus member that refuses and looks like a supported example;
  - keeping it with a header, because that re-pins anyway and leaves a confusing duplicate.

  — **Reversibility:** reversible — git history keeps the bytes, and the tombstone digest names them.
- **D-06:** The Phase 28 and Phase 30 goldens stay **byte-frozen** (D-15 rule):
  - `testdata/phase28/entry-exit-class-manifest.json`
  - `testdata/phase30/prechange-corpus.json`

  `session_phase28_manifest_test.go` gains a dated `phase28RetiredEntries` list, and `session_phase30_prechange_test.go` gains `phase30RetiredPaths`, mirroring the existing `phase28NewlyAddedEntries` and `phase30NewCommittedPaths` pattern. A retired path may be absent, and it is a **failure if a retired path is observed again**, so the list cannot go stale. Extend the existing mutated-golden negative control to cover that case.
- **D-07:** The dated amendment (2026-10-12, Phase 32, DX-19) goes in `.planning/PROJECT.md`, replacing the "remains a refused integration target" paragraph at about line 353. It records the deletion, the tombstone digest, the last refusal, the supported app and guide paths, and the frozen-golden allow-lists.
  - These live rows point at the amendment instead of re-describing it: `LANGUAGE-MATURITY.md:297` (and the `:412` note if live), `PRODUCT-ROADMAP.md:298`, and `examples/checksum/README.md:140`.
  - `REQUIREMENTS.md` and `ROADMAP.md` get status-only updates at phase close.
  - Dated historical records stay untouched: phase dirs, `milestones/`, `research/`, and the `LANGUAGE-MATURITY.md:180` M005 kickoff entry.

### Final receipts and the revision gate (EVD-15, SC3)
- **D-08:** Extend the existing M005 gate rather than building a new one. The existing pieces are `scripts/verify-phase27.sh`, which writes the identity log, and `scripts/check-phase27-receipts.sh`, which checks one macOS and one Linux log. CI already runs them in `current evidence aggregate`, uploads the `receipt-<os>.log` artifact, and uses `fetch-depth: 0`.
  - **Revision = code digest, not commit SHA.** `codedigest` is the sha256 of `git ls-tree -r HEAD -- cmd internal examples scripts go.mod .github/workflows/ci.yml`. **[synthesis]** It excludes `*.md`, so guide and README prose edits and the committed receipt cannot invalidate it, and `.planning` is outside its paths entirely.
  - The same digest is computed in three places: the verify script (a new `codedigest=` field), the gate script, and a Go helper.
  - This survives shallow clones, rebases and squash merges, and committing the receipt does not move it.

  Rejected:
  - `headSha` plus `git diff`, because history can go missing;
  - per-subtree tree hashes, which are fiddly;
  - BuildID alone, which differs per host and toolchain. It may still be an extra per-host field.

  — **Reversibility:** costly — the digest's path set becomes the receipt contract, so changing it re-takes the receipts.
- **D-09:** The receipt is `.planning/phases/32-…/32-RECEIPT.json`, committed. It records:
  - the run ID, the hosted `headSha`, the `codedigest` and `"state": "current"`;
  - per host (ubuntu-latest, macos-latest): the sha256 of the downloaded log, the runner OS and arch, and the go and clang versions;
  - per-program results for `checksum`, `adler_low`, the `adler32_cli` library demo, `sum_to_n` and FizzBuzz (which closes N2).

  At ingest, use `gh run view <id> --json headSha,conclusion,jobs` and `gh run download`, and the test recomputes the log hashes.

  The gate rejects with named reasons:
  - `digest_mismatch`;
  - a log `codedigest` different from the receipt's;
  - a dirty tree;
  - a missing host;
  - a missing program.

  Armed controls: mutate one code file, mutate the digest, drop a host, and flip the tree to modified.

  The gate checks consistency, not authenticity; authenticity comes from the hosted run ID. The script header already says this, so keep stating it.
- **D-10a, reopen semantics:** following the Phase 25 "Historical receipts" precedent, the receipt test **never fails on digest drift alone**.
  - It fails only if the code digest differs *and* a living doc (`LANGUAGE-MATURITY.md`, `PRODUCT-ROADMAP.md` or STATE) still claims current M006 hosted evidence.
  - So the first M007 code change must flip those claims to "reopened <date>" and keep the receipt as historical. CI stays green and no claim goes stale unnoticed.
- **D-10b, public-route audit:** this is the roadmap research flag.
  - One Go test enumerates the public routes: the `cmd/schway` subcommands and flags, `examples/**/*.schway` entry programs, and the `adler32_cli`.
  - It asserts each route has either a hosted-receipt program row or an explicit "not covered because" row, and fails on a new uncovered route.
  - The receipt test asserts that its program set equals the audit's enumerated set.

### Folded todo
- **D-10:** Fold **"Formatter merges a trailing result identifier into the preceding var assignment"** as the *first* Phase 32 plan, so it lands before the receipt freeze. This is a correctness defect in the lossless-format contract, not a feature.
  - The fix covers `internal/compiler/syntax/format.go` plus a round-trip regression test.
  - **Do not reformat the committed adler32 sources.** About 48 Go files pin their digests. The todo's own "format both files canonically" suggestion is dropped.
  - The guide must not claim the examples are format-canonical. The round-trip test runs on copies.
  - Run the lossless and generated-program lanes.

  The researcher suggested `/gsd-quick` before Phase 32; folding it as the first plan gives the same ordering under one roadmap owner.

### Claude's Discretion
- The fence-marker syntax, the table column layout, and the minimum-block count N.
- The env mechanism that makes a missing clang a failure on hosted CI, such as `SCHWAY_REQUIRE_NATIVE=1` or reuse of an existing CI env signal.
- Whether the guide doc-test's JSON cross-check lives in `cmd/schway` or beside `TestPhase31ExpectedMatrixJSON` in `internal/compiler/testsupport/checksum_matrix_test.go`.
- The exact file and test names for the route audit and receipt gate (`session_phase32_*` or `native/phase32_*`). Register any new tests the session meta-tests (inventory, groundedness, reconciliation) require.
- Whether the verify and gate script changes generalize `verify-phase27.sh` in place or add `verify-phase32.sh` and `check-phase32-receipts.sh`. Phase 27's receipt behaviour must stay valid either way.

### Post-research amendments (2026-10-12, auto-followed)
These amendments come from `32-RESEARCH.md` findings and open questions. Each recommendation was auto-followed under the standing fan-out preference.
- **D-08a (amends D-08):** the codedigest path set also includes `native` and `testdata`. `native/*.c` is runtime code compiled by the native path, and `testdata/` holds the goldens and fixtures the gates read. The final set is `cmd internal examples native scripts testdata go.mod .github/workflows/ci.yml`, excluding `*.md`.
  - Recipe: `git -c core.quotePath=false ls-tree -r HEAD -- <paths> | grep -v '\.md$' | shasum -a 256`.
  - Go runs the same git command and filters with `HasSuffix(".md")`.

  — **Reversibility:** costly — this is the receipt contract.
- **D-08b (refines D-08 / Discretion):** add new scripts (`scripts/codedigest.sh`, `scripts/verify-phase32.sh`, `scripts/check-phase32-receipts.sh`) plus an additive CI step and artifact upload. The phase27 scripts and their pins (`TestPhase27HostReceiptGate`, `TestPhase27CIHostMatrix`, `TestPhase27GuideContract`) stay untouched.
- **D-08c (ordering):** the receipt test lives inside the digest, so it is committed before the freeze.
  - It must pass when no receipt exists and no living doc claims bound evidence, without `t.Skip`.
  - Ingest is a single `.planning`-only commit holding the receipt, the downloaded logs and the living-doc markers. The same test enforces from then on.
  - Dispatch exactly once, on the branch ref: `ci.yml` has `cancel-in-progress: true`, and a `pull_request` run checks out a merge commit.
  - The checkpoint plan names the freeze commit explicitly.
- **D-10c (refines D-10):** the formatter fix changes the *formatted output* (`format_sha256`) of 5 committed sources:
  - `examples/checksum/adler32.schway`
  - `examples/checksum/adler_low.schway`
  - `examples/phase29/adler32_fixture.schway`
  - `examples/phase29/byte_index_oob.schway`
  - `testdata/phase29/adler32_library.schway`

  The sources themselves stay byte-identical. The frozen `testdata/phase30/prechange-corpus.json` is not regenerated. A dated `phase30FormatChanges` allow-list holds hashes recomputed from the real fix, with a stale-entry control.
  - The guard must exclude `match` context, because without it 9 phase12/18 outputs regress.
  - The regression test needs a var-body table, because the generated lanes never emit that shape.
  - `}` / `else {` line placement is out of scope.
- **D-07a (extends D-07):** frontier retirement also updates:
  - `PRODUCT-ROADMAP.md:272`;
  - the `LANGUAGE-MATURITY.md:304` counts (163 → 162 programs, 5,389 → 5,352 lines; `TestLanguageMaturityCountsAreCurrent`, `TestSelfDescribingDocsGuardIsNotInert`).

  `session_phase28_manifest_test.go` has no negative-control test yet, so extract a `phase28Compare` helper first, then add the retired-path control. In the hygiene scan, "live" means lines not under a dated heading, with an armed control showing a historical line is exempt.
- **D-01a (extends D-01):** guide outputs go under `.schway/<name>/` with `mkdir -p`, never the repo root, because the root outputs are not git-ignored and would dirty the receipt tree. This copies `phase273GuideOutputTargets`.
  - Spawn with `sh -c 'cd "$0" && eval "$1"' <copy> <line>`, which respects the spawn lint.
  - Inputs come from `testsupport.ChecksumCaseMatrix()` by case name.
  - `adler_low` values are cross-checked against `ChecksumOracleRows` (`checksum % 65536`).
  - The copy stays alive for native helper resolution, with no `-trimpath`.
- **D-04a (resolves Discretion):** native policy is a pure function. It fails when clang is missing and either `GITHUB_ACTIONS=true` or `SCHWAY_REQUIRE_NATIVE=1`, and skips otherwise. Tested as `TestPhase32NativePolicy`. `ci.yml` also sets `SCHWAY_REQUIRE_NATIVE=1`.
- **D-10d (route audit):** enumerate routes by Go AST over `run()`, `runApplication` and `runApplicationRun` in a `package main` test, with `usageResult()` text as a secondary cross-check.
  - Usage text omits `stats` and `build --manifest`. Record that as audit rows plus a pending todo; do not edit the usage text in this phase.
  - The 9 uncovered entry programs get real "not covered because" reasons.
  - The test must not call inventory-restricted `cgen` emitters or `callgraph.EntryFunction`.
- **D-11 (planning-doc conventions):**
  - PLAN docs carry `verification_role: proposal`.
  - `32-VALIDATION.md` goes `planned` → `validated` at closure.
  - Every `go test -run` span in VALIDATION and VERIFICATION docs must resolve to existing tests, so pin commands only after the tests land, or use plain test-name spans.
  - Every SUMMARY carries `coverage:`. The push-checkpoint plan is the only `human_judgment: true` entry.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements and roadmap
- `.planning/ROADMAP.md` § Phase 32: the goal, SC1–SC3, and the research flag (public-route audit; push and dispatch need user approval).
- `.planning/REQUIREMENTS.md`: DX-19 and EVD-15.
- `.planning/PROJECT.md` § Verification Operating Preference: zero human UAT, and the from-Phase-32 SUMMARY `coverage:` rule.
- `.planning/PROJECT.md` around line 353: the frontier paragraph to replace with the D-07 amendment.

### Prior phase decisions
- `.planning/phases/31-bounded-file-checksum-and-second-consumer/31-CONTEXT.md`: D-01 (layout), D-02 (frontier deferred to Phase 32), D-03 (`adler_low`), and the deferred todos.
- `.planning/phases/31-bounded-file-checksum-and-second-consumer/31-VERIFICATION.md`: hosted run 38299800002 at c7256924.
- `.planning/research/SUMMARY.md`: M006 Decisions 1–10 (settled).

### Guide and doc-test
- `examples/checksum/README.md`: the guide home and current non-claims (line 140 is the frontier note).
- `examples/checksum/checksum.expected.json`: case names and the limit.
- `examples/phase27/README.md`: the M005 guide precedent.
- `cmd/schway/main_phase31_test.go`: `phase31RouteInput`, `phase30CopyExampleDir` and the public-route probes.
- `internal/compiler/session/session_phase273_docs_test.go`: `phase273CleanCheckoutBlock` and `phase273GuideViolations`, the block-extraction style. That test declined to execute the guide; Phase 32 closes that.
- `internal/compiler/testsupport/checksum_matrix_test.go`: `TestPhase31ExpectedMatrixJSON`.

### Frontier retirement
- `examples/checksum.schway`: to delete.
- `internal/compiler/session/session_phase20_test.go`: `TestPhase20ChecksumFrontier`, which is replaced.
- `internal/compiler/session/session_phase28_manifest_test.go` and `testdata/phase28/entry-exit-class-manifest.json` (frozen).
- `internal/compiler/session/session_phase30_prechange_test.go` and `testdata/phase30/prechange-corpus.json` (frozen).
- `.planning/LANGUAGE-MATURITY.md:297`, `:412`; `.planning/PRODUCT-ROADMAP.md:298`.

### Receipts
- `scripts/verify-phase27.sh` and `scripts/check-phase27-receipts.sh`: the existing identity log and two-host gate.
- `.github/workflows/ci.yml`: the `current evidence aggregate` job (about lines 101-125) and the receipt artifact upload.
- `internal/compiler/native/phase25_utility_test.go` around lines 190-305: the `TestPhase25LivingRoadmap` "no hosted claim without a bound receipt" precedent.

### Folded todo
- `.planning/todos/pending/2026-10-12-formatter-merges-trailing-result-into-var-assignment.md`
- `internal/compiler/syntax/format.go`

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `testsupport.ProjectPath`, `testsupport.RunCLIErr` and the `exec.LookPath("clang")` skip pattern.
- The Phase 31 public-route helpers in `cmd/schway/main_phase31_test.go`.
- The Phase 27 receipt scripts and the CI artifact upload. Extend them; do not duplicate them.
- `phaseArtifactGlob` and `buildTestIndex` in the session meta-tests, used for receipt and coverage checks.

### Established Patterns
- Dated allow-lists against frozen goldens (`phase28NewlyAddedEntries`, `phase30NewCommittedPaths`), never a re-pin.
- Armed negative controls for every gate.
- Historical dated records are never rewritten; only live rows are updated.
- The stdlib-only rule: "another copy and paste is better than another dep".

### Integration Points
- CI `current evidence aggregate` on ubuntu-latest and macos-latest: add the new programs and the `codedigest`.
- The session meta-tests: register new `session_phase32_*` tests.
- Every Phase 32 SUMMARY must carry a `coverage:` block (`TestSummaryCoverageGroundedness`).

</code_context>

<specifics>
## Specific Ideas

- The README *is* the test input, so users copy and paste exactly what CI ran.
- Receipts stay valid across `.planning` and docs commits, and are reopened only by code changes. Reopening is a documentation-truth gate, not a CI-breaking tripwire.
- The only human step in the whole phase is one paste-ready line: `! git push -q origin phase26-validation && gh workflow run ci.yml --ref phase26-validation`.

</specifics>

<deferred>
## Deferred Ideas

- Giving `run` an INPUT argument (or refusing a parameterised main before engine dispatch). This is a design choice and a new capability for a later milestone.
- Reformatting the committed adler32 sources to canonical form. It would re-pin about 48 digest sites; not worth it.
- Authenticated receipts (signed attestations, full SLSA provenance). The gate stays consistency-only; authenticity comes from the hosted run ID.

### Reviewed Todos (not folded)
- **run --engine on a parameterised main reports an operational failure**
  (`.planning/todos/pending/2026-10-12-run-parameterised-main-reports-operational-failure.md`).
  - Kept pending, because the fix re-pins the deliberate Phase 28 DX-17 exit-class rows and the real fix is a design change.
  - The guide routes through `build` + `app run -- PATH`. It names `schway run` on a parameterised main as unsupported in one known-limits line, linking the todo.

</deferred>

---

*Phase: 32-clean-checkout-guide-and-final-receipts*
*Context gathered: 2026-10-12*
