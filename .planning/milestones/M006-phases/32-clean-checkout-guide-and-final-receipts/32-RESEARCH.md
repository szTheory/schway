---
verification_role: proposal
---

# Phase 32: Clean-Checkout Guide and Final Receipts - Research

**Researched:** 2026-10-12
**Domain:** Go 1.24 stdlib-only compiler repo. This phase touches the lossless formatter, frozen-golden allow-lists, a README-driven executed doc-test, POSIX shell receipt scripts, a GitHub Actions workflow and receipt-ingest tests. No new dependencies.
**Confidence:** HIGH on every repo fact: each was opened, run or probed this session. Probes ran in a scratch copy of the tree outside the repo, and nothing was committed. MEDIUM on the design recommendations, which are flagged where they deviate from CONTEXT.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

Every decision came from advisor research at the `minimal_decisive` tier and was auto-followed. Two cross-area synthesis adjustments are marked **[synthesis]**.

- **D-00 [synthesis]:** Every code change lands *before* the receipt freeze. Plan order:
  1. the formatter fix (D-10);
  2. frontier retirement (D-05..D-07) and the guide doc-test (D-01..D-04), which can run in parallel;
  3. the public-route audit and gate changes (D-08), whose final code commit is the receipt target;
  4. a human checkpoint plan with the single push and dispatch line;
  5. a `.planning`-only receipt-ingest plan, which is the last plan of the milestone.

  If the hosted run fails or any code changes after step 3, the ingest plan stops and routes back to step 3.
- **D-01:** The guide stays in `examples/checksum/README.md`, following the `examples/phase27/README.md` precedent. It is the *executable spec*.
  - Runnable blocks are marked fences (for example ` ```sh doc-test=clean-checkout `), holding literal lines starting with `go build -o ./schway ./cmd/schway`, then `./schway build … --manifest … --output …` and `./schway app run <artifact> -- PATH` for `checksum`, `adler_low` and the `adler32_cli` library demo.
  - A results table has stable row keys equal to the case names in `checksum.expected.json`.
  - The test holds **no commands or numbers of its own**: it reads them from the README and cross-checks them against `checksum.expected.json`, including `"limit": 65536`.
  - Rejected: a test-side mirror table; `testscript`/txtar (a dependency, and it inverts the doc-test); a new top-level guide.
- **D-02:** Clean checkout is simulated by copying `git ls-files -co --exclude-standard` into `t.TempDir()`, falling back to a walk that skips `.git` and `.schway`. The test runs the guide lines with `sh -c` in that copy, using a fresh `./schway` built from the copy, and substitutes only the `PATH` token. Each subprocess gets a context timeout; no `t.Parallel`. The test lives in `cmd/schway/main_phase32_guide_test.go`. Rejected: `git archive HEAD` and `git worktree`.
- **D-03:** The guide's error rows cover only deterministic cases: missing (fresh tempdir), 65,537 bytes of 0xFF giving `FileError.FileTooLarge`, a directory giving `FileError.FileUnavailable`, plus one prose row "other unreadable inputs give `FileError.FileUnavailable`, exit 65". Every error row asserts exit 65, empty stdout, the fixed stderr, and that the temp path segment is absent. FIFO, mode-000, `/dev/null`, EIO and AllocFailed stay with the Phase 31 matrix tests.
- **D-04:** Anti-vacuity and armed controls: assert at least N marked blocks, that all three programs appear, and that rows executed equals table rows. Negative controls run the checker on an in-memory mutated guide, once with one expected value changed and once with one command renamed; each must fail. On hosted CI a missing clang is a failure, not a skip (mechanism is Claude's discretion). Test names: `TestPhase32GuideCleanCheckout`, `TestPhase32GuideParserNegativeControls`, `TestPhase32GuideMatchesExpectedJSON`. The guide gets an "Earlier programs" pointer to `examples/phase27/README.md` for `sum_to_n` and FizzBuzz; it is not expanded to cover them.
- **D-05:** **Delete** `examples/checksum.schway` (tombstone SHA-256 `ae95e550df67114c3464dda9cc24779f8a2efd69bd2b99dd0ccb2acc235cc154`; last pinned refusal `syntax.expected_rbrace` at 521-527). Replace `TestPhase20ChecksumFrontier` with `TestPhase32ChecksumFrontierRetired` in `internal/compiler/session/session_phase32_retire_test.go`. It asserts: the old path is `ErrNotExist`; `examples/checksum/checksum.schway` exists; no `.schway` under `examples/` or `testdata/` has the tombstone digest; an armed control (old bytes in a temp dir are flagged); a doc-hygiene scan of the *live* docs only (PROJECT.md, LANGUAGE-MATURITY.md, PRODUCT-ROADMAP.md, `examples/checksum/README.md`) finds no line naming the old path without "retired" or "amendment". Reversible.
- **D-06:** The Phase 28 and Phase 30 goldens stay **byte-frozen** (`testdata/phase28/entry-exit-class-manifest.json`, `testdata/phase30/prechange-corpus.json`). `session_phase28_manifest_test.go` gains a dated `phase28RetiredEntries` list and `session_phase30_prechange_test.go` gains `phase30RetiredPaths`, mirroring `phase28NewlyAddedEntries` and `phase30NewCommittedPaths`. A retired path may be absent, and it is a **failure if a retired path is observed again**. Extend the existing mutated-golden negative control to cover that case.
- **D-07:** The dated amendment (2026-10-12, Phase 32, DX-19) goes in `.planning/PROJECT.md`, replacing the "remains a refused integration target" paragraph at about line 353. It records the deletion, tombstone digest, last refusal, supported app and guide paths, and the frozen-golden allow-lists. Live rows point at the amendment: `LANGUAGE-MATURITY.md:297` (and `:412` if live), `PRODUCT-ROADMAP.md:298`, `examples/checksum/README.md:140`. `REQUIREMENTS.md` and `ROADMAP.md` get status-only updates at phase close. Dated historical records stay untouched (phase dirs, `milestones/`, `research/`, the `LANGUAGE-MATURITY.md:180` M005 kickoff entry).
- **D-08:** Extend the existing M005 gate rather than building a new one (`scripts/verify-phase27.sh`, `scripts/check-phase27-receipts.sh`, CI `current evidence aggregate`, `receipt-<os>.log` artifact upload, `fetch-depth: 0`). **Revision = code digest, not commit SHA.** `codedigest` is the sha256 of `git ls-tree -r HEAD -- cmd internal examples scripts go.mod .github/workflows/ci.yml` **[synthesis]** excluding `*.md`; `.planning` is outside its paths. The same digest is computed in three places: the verify script (a new `codedigest=` field), the gate script, and a Go helper. Rejected: `headSha` plus `git diff`; per-subtree tree hashes; BuildID alone (may be an extra per-host field). Reversibility: costly, because the digest's path set becomes the receipt contract.
- **D-09:** The receipt is `.planning/phases/32-…/32-RECEIPT.json`, committed. It records the run ID, hosted `headSha`, `codedigest`, `"state": "current"`; per host (ubuntu-latest, macos-latest) the sha256 of the downloaded log, runner OS and arch, go and clang versions; per-program results for `checksum`, `adler_low`, the `adler32_cli` library demo, `sum_to_n` and FizzBuzz (closes N2). At ingest use `gh run view <id> --json headSha,conclusion,jobs` and `gh run download`; the test recomputes the log hashes. The gate rejects with named reasons: `digest_mismatch`; a log `codedigest` different from the receipt's; a dirty tree; a missing host; a missing program. Armed controls: mutate one code file, mutate the digest, drop a host, flip the tree to modified. The gate checks consistency, not authenticity; authenticity comes from the hosted run ID.
- **D-10a, reopen semantics:** following the Phase 25 "Historical receipts" precedent, the receipt test **never fails on digest drift alone**. It fails only if the code digest differs *and* a living doc (`LANGUAGE-MATURITY.md`, `PRODUCT-ROADMAP.md` or STATE) still claims current M006 hosted evidence. The first M007 code change must flip those claims to "reopened <date>" and keep the receipt as historical.
- **D-10b, public-route audit:** one Go test enumerates the public routes (`cmd/schway` subcommands and flags, `examples/**/*.schway` entry programs, the `adler32_cli`). It asserts each route has either a hosted-receipt program row or an explicit "not covered because" row, and fails on a new uncovered route. The receipt test asserts its program set equals the audit's enumerated set.
- **D-10 (folded todo):** Fold "Formatter merges a trailing result identifier into the preceding var assignment" as the *first* Phase 32 plan. The fix covers `internal/compiler/syntax/format.go` plus a round-trip regression test. **Do not reformat the committed adler32 sources** (about 48 Go files pin their digests). The guide must not claim the examples are format-canonical; the round-trip test runs on copies. Run the lossless and generated-program lanes.

### Claude's Discretion

- The fence-marker syntax, the table column layout, and the minimum-block count N.
- The env mechanism that makes a missing clang a failure on hosted CI, such as `SCHWAY_REQUIRE_NATIVE=1` or reuse of an existing CI env signal.
- Whether the guide doc-test's JSON cross-check lives in `cmd/schway` or beside `TestPhase31ExpectedMatrixJSON` in `internal/compiler/testsupport/checksum_matrix_test.go`.
- The exact file and test names for the route audit and receipt gate (`session_phase32_*` or `native/phase32_*`). Register any new tests the session meta-tests (inventory, groundedness, reconciliation) require.
- Whether the verify and gate script changes generalize `verify-phase27.sh` in place or add `verify-phase32.sh` and `check-phase32-receipts.sh`. Phase 27's receipt behaviour must stay valid either way.

### Deferred Ideas (OUT OF SCOPE)

- Giving `run` an INPUT argument (or refusing a parameterised main before engine dispatch).
- Reformatting the committed adler32 sources to canonical form.
- Authenticated receipts (signed attestations, full SLSA provenance). The gate stays consistency-only.
- Reviewed, not folded: the todo "run --engine on a parameterised main reports an operational failure". The guide routes through `build` + `app run -- PATH` and names `schway run` on a parameterised main as unsupported in one known-limits line.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| DX-19 | From a clean checkout a developer follows documented commands to build and run `checksum`, the second consumer and the `adler32` library. The guide states expected results, the 65,536-byte limit and the error outcomes. The provisional `examples/checksum.schway` frontier is retired with a dated amendment. | Findings 1, 2 and 3 below: the formatter fix, the exact retirement blast radius (measured by running the full suite on a deletion experiment), and the guide doc-test design and its pitfalls. |
| EVD-15 | Hosted macOS and Linux receipts are taken at the final M006 revision and cover the M006 and carried-forward M005 programs (closes N2). A later code change reopens the receipts. | Findings 4, 5 and 6: the codedigest algorithm, script and gate design, the route-audit enumeration, and the reopen semantics. |
</phase_requirements>

## Project Constraints (from CLAUDE.md / AGENTS.md)

- CLAUDE.md only routes spike findings (`Skill("spike-findings-ai-lang")`) and points to AGENTS.md. Those spikes cover ownership kernel, loan liveness, origins, certificates and the FFI ABI. None of them bears on Phase 32, so the skill was not loaded.
- AGENTS.md "Constraints": Go 1.24 stdlib only, Clang for the native path, macOS and Linux are the host priorities, and deterministic evidence with negative controls. It also says to prefer copy-local code over a dependency.
- AGENTS.md "GSD Workflow Enforcement": repo edits go through a GSD command. Research writes only this file.
- AGENTS.md "Living language roadmap": at phase transitions read `.planning/PRODUCT-ROADMAP.md` and `.planning/LANGUAGE-MATURITY.md` and update them within the workflow. Preserve archived decisions except for explicit dated amendments.
- PROJECT.md § Verification Operating Preference: zero human UAT; every plan names concrete verification commands; from Phase 32 on, every SUMMARY carries a `coverage:` block (see Finding 7).
- Memory: session meta-tests catch planning-doc and call-site drift, so run them locally before hosted dispatch. Every GSD step ends with the exact next command.

## Summary

All seven questions were checked against code, and four checks changed the plan.

1. **The formatter fix is small and safe, but it moves three frozen Phase 30 golden hashes.** I reproduced the bug and prototyped a fix in a scratch copy. The fix passes the syntax, check, core and cmd suites. Run over all 163 committed `.schway` files, it changes the output of exactly the 5 files that end a `var` body with a bare result. Three of those 5 sit in the frozen `testdata/phase30/prechange-corpus.json`, which pins `format_sha256` per file: `examples/phase29/adler32_fixture.schway`, `examples/phase29/byte_index_oob.schway` and `testdata/phase29/adler32_library.schway`. The golden must stay byte-frozen (D-06), so plan 32-01 needs a dated `phase30FormatChanges` allow-list. This is not in CONTEXT. The same bug also hits `adler_low.schway` (the todo names only the two adler32 files).
2. **Frontier retirement trips more than the two named goldens.** I ran `go test ./...` on a copy with the file deleted. The failures that matter are `TestPhase20ChecksumFrontier`, `TestPhase28EntryExitClassManifest`, `TestPhase30PreChangeGolden` and `…IsNotInert`, plus `TestLanguageMaturityCountsAreCurrent`. That last one is new information: `LANGUAGE-MATURITY.md:304` states "163 programs, 5,389 lines" and must become 162 and 5,352. The other failures in that run came from my copy's fake git history (reconciliation tests), not from the change. No Go file outside `session_phase20_test.go` references the old path.
3. **The guide doc-test is feasible with the existing helpers, but the current README writes artifacts into the repo root.** `./checksum`, `./adler_low` and `./adler32-module` are not git-ignored, so following the guide dirties the tree. The Phase 27 guide was already forced to route outputs under the ignored `.schway/` (D-27.3-15). The Phase 32 guide must do the same. `RunCLIErr` has no working-directory parameter, so run the guide lines through `sh -c 'cd "$0" && …'` with the temp copy as `$0`.
4. **The receipt design needs two corrections to D-08's path set.** `native/*.c` is runtime code resolved by `runtime.Caller` and sits outside the digest paths. `testdata/` holds the frozen goldens and is also outside. Recommend adding both; this deviates from D-08 and must be confirmed before the freeze (Open Question 1). Also, `verify-phase27.sh` and the phase27 gate are pinned by about 25 mutation-row tests plus CI-text tests, so add `verify-phase32.sh` and `check-phase32-receipts.sh` rather than generalizing in place.
5. **The CLI route list cannot come from the usage string.** `usageResult()` omits the `stats` subcommand and the `build --manifest` flag, both live in `run()`. Enumerate by AST over `run()`, and treat the usage text as a secondary cross-check.

**Primary recommendation:** Execute the D-00 order, with four additions: (a) a `phase30FormatChanges` allow-list in plan 32-01; (b) move the guide's artifacts under `.schway/` and update the LANGUAGE-MATURITY corpus counts in the retirement plan; (c) add `native` and `testdata` to the codedigest path set, subject to confirmation; (d) add new `phase32` scripts instead of editing the phase27 ones. Land every Go test, including the receipt and route-audit tests, before the freeze, so the ingest plan changes only `.planning`.

## Architectural Responsibility Map

This is a single-binary CLI repo with no browser, server or CDN tier. The tiers below are the repo's own layers.

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Lossless formatting | `internal/compiler/syntax` (`format.go`) | `session.Format`, `cmd/schway format` | The formatter is a token and CST projection; the fix belongs in the projection loop, not in the CLI. |
| Guide execution | `cmd/schway` tests (public binary route) | `testsupport` (`RunCLIErr`, `ChecksumCaseMatrix`) | Only the built `./schway` proves what a user types. The README owns commands and numbers. |
| Frontier retirement | `session` tests (corpus walkers and goldens) | `.planning` living docs | The walkers live in `session_phase28_*` and `session_phase30_*`; the claims live in docs. |
| Revision binding (`codedigest`) | `scripts/*.sh` (producer and gate) | Go helper in `session` tests | The producer runs on hosted runners, where only sh and git are guaranteed. Go re-derives the value independently. |
| Receipt ingest and reopen semantics | `session` tests over `.planning` | `gh` CLI, a human-run step | Authenticity comes from the hosted run ID, so the gate is consistency-only. |
| Route audit | `cmd/schway` tests (AST of `main.go`, `usageResult`) | `examples/**` walk | Subcommands exist only as the `if` chain in `run()`. |
| Hosted execution | `.github/workflows/ci.yml` | `gh workflow run` (human) | The workflow already has the two-host matrix and artifact upload. |

## Standard Stack

### Core

| Tool | Version | Purpose | Why Standard |
|------|---------|---------|--------------|
| Go stdlib `testing`, `os/exec` via `testsupport.RunCLIErr`, `go/parser`, `crypto/sha256`, `encoding/json` | go 1.24 (`go.mod`: `go 1.24`; local `go1.24.0`; hosted runners `go1.24.13`) | All new tests | The repo's only test stack. `scanUnboundedSpawns` lints every spawn in the repo, so spawns must go through the existing bounded helpers. |
| POSIX `sh` + `git` + `grep`/`sed`/`tr`/`cut` | system | `verify-phase32.sh`, `check-phase32-receipts.sh` | Same stack as the Phase 27 scripts. |
| `shasum -a 256` (macOS and Ubuntu) or `sha256sum` (Ubuntu) | system | codedigest in sh | There is no portable single tool. Probe `command -v` and fall back. |
| `gh` 2.101 | local | Ingest only (human-adjacent plan) | `gh run view --json headSha,conclusion,jobs` was run this session against run 38299800002 and returned exactly those fields. |

### Supporting

| Asset | Purpose | When to Use |
|-------|---------|-------------|
| `testsupport.ProjectPath`, `RunCLIErr`, `BuildCLIErr`, `ChecksumCaseMatrix`, `ChecksumOracleRows`, `Phase29FixtureExpected`, `ChecksumLimit` | Paths, bounded spawns, materialising inputs by case name, independent oracle values | Guide doc-test and cross-checks. |
| `phaseArtifactGlob`, `buildTestIndex` (session meta helpers) | Find SUMMARY and receipt files, resolve test names | Receipt and coverage tests inside `session_test`. |
| `phase273GuideOutputTargets` / `phase273TargetAllowed` (session_test) | Output-target checker (only `.schway/…` or `schway` allowed) | **Copy** the logic into `cmd/schway`'s test file. A `cmd/schway` test cannot import `session_test`, and copying matches the repo rule. |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| New `phase32` scripts | Generalize `verify-phase27.sh` in place | `TestPhase27HostReceiptGate` (about 25 mutation rows keyed to the phase27 row grammar), `TestPhase27CIHostMatrix` and `TestPhase27GuideContract` pin the phase27 scripts and CI text. In-place edits risk all three and the pinned M005 receipts. |
| `sh -c 'cd "$0" && …'` wrapper | New `testsupport.RunCLIDirErr` with `Dir` | The wrapper needs no change to the shared helper, which is audited by the spawn lint. A new helper is also acceptable and cleaner if the planner wants it. |

**Installation:** none. Stdlib only.

**Version verification:** `go version` reports `go1.24.0 darwin/arm64`. `gh --version` reports `2.101.0`. The hosted runners' versions come from the downloaded Phase 27 logs: `go1.24.13 linux/amd64` with `Ubuntu clang version 18.1.3 (1ubuntu1)`, and `go1.24.13 darwin/arm64` with `Apple clang version 21.0.0 (clang-2100.1.1.101)` [VERIFIED: `gh run download 38299800002 -R szTheory/schway`].

## Package Legitimacy Audit

No external packages are installed by this phase. The legitimacy check has no subjects.

**Packages removed due to [SLOP] verdict:** none
**Packages flagged as suspicious [SUS]:** none

## Architecture Patterns

### System Architecture Diagram

```
 README (examples/checksum/README.md)           checksum.expected.json   testsupport.ChecksumCaseMatrix
   marked fences + results table  ----------------+------------------------+  (input bytes per case name)
            |                                      |
            v                                      v
  [doc-test parser] --cross-check--> limit 65536, error exit 65, case names, stdout/stderr
            |
            v  copy `git ls-files -co --exclude-standard` -> t.TempDir()   (skip missing, skip .git/.schway)
  [sh -c 'cd "$0" && <block>'] go build -o ./schway ./cmd/schway
            -> ./schway build <entry> --manifest <m> --output .schway/checksum/<app>
            -> per table row: materialise input -> ./schway app run <artifact> -- <path>
            -> compare exit / stdout / stderr to README row; assert tempdir segment absent from stderr

 push (human, once) -> workflow_dispatch on phase26-validation
   ubuntu-latest + macos-latest: checks (go test ./..., race) + evidence-aggregate
        evidence-aggregate: verify-phase27.sh (unchanged) + NEW verify-phase32.sh
              rows: identity(codedigest, headsha, tree) / toolchain / case per program / control / result
              -> receipt-<os>.log artifact (phase32-receipt-<os>-<sha>)
 gh run view/download (ingest plan) -> logs + 32-RECEIPT.json (.planning only)
 check-phase32-receipts.sh LOG_A LOG_B: recompute codedigest from checkout, compare, check hosts/programs/tree
 Go receipt test: recompute log sha256; if a living doc claims current M006 evidence, require digest match
```

### Recommended Project Structure

```
internal/compiler/syntax/format.go                                  # 32-01: one added guard in Format()
internal/compiler/syntax/syntax_phase32_format_test.go              # 32-01: var-body regression, copies of 5 sources
internal/compiler/session/session_phase30_prechange_test.go         # 32-01: phase30FormatChanges; 32-02: phase30RetiredPaths
internal/compiler/session/session_phase28_manifest_test.go          # 32-02: phase28RetiredEntries + extracted compare + control
internal/compiler/session/session_phase32_retire_test.go            # 32-02: TestPhase32ChecksumFrontierRetired
internal/compiler/session/session_phase20_test.go                   # 32-02: delete TestPhase20ChecksumFrontier, keep phase20SHA256 if used elsewhere
cmd/schway/main_phase32_guide_test.go                               # 32-03: guide doc-test (D-02)
cmd/schway/main_phase32_routes_test.go                              # 32-04: route audit (package main sees run() and usageResult())
scripts/codedigest.sh, scripts/verify-phase32.sh, scripts/check-phase32-receipts.sh   # 32-04
internal/compiler/session/session_phase32_receipt_test.go           # 32-04: digest, gate mutation table, living-doc reopen rule
.planning/phases/32-.../32-RECEIPT.json + receipts/receipt-{macos,ubuntu}-latest.log    # 32-06 only
```

### Pattern 1: Dated allow-list against a frozen golden

Both frozen goldens already use "only these paths may differ" maps with dated comments. Phase 32 needs a third and fourth kind of entry on the same pattern.

[VERIFIED: internal/compiler/session/session_phase30_prechange_test.go:61-73, 237-275]. Current shape, quoted:

```go
var phase30NewCommittedPaths = map[string]bool{
	"examples/checksum/adler32.schway":     true,
	...
}
...
			got, ok := byPath[want.Path]
			if !ok {
				... problems = append(problems, want.Path+": recorded at golden revision "+golden.Revision+" but no longer present")
```

Required additions:
- `phase30RetiredPaths = map[string]bool{"examples/checksum.schway": true}` with a dated comment. In `phase30Compare`, a golden path that is absent is accepted only when it is in the map. A path in the map that *is* observed is a problem ("retired path observed again").
- `phase30FormatChanges = map[string]string{path: newFormatSHA256}` for the three files whose formatter output legitimately changes in 32-01. In the field loop, the `format_sha256` comparison uses the allow-listed hash instead of the golden. Fail if the allow-listed value equals the golden (stale), as `phase28ClassChanges` does ("stale phase28ClassChanges entry equals the golden"). Compute the three hashes **after** the final fix is written. The prototype hashes below are not authoritative.
- `TestPhase30PreChangeGoldenIsNotInert` already mutates one hash. Add two sub-controls: observed contains a retired path, and a format-change entry equal to the golden.

For Phase 28, CONTEXT says "extend the existing mutated-golden negative control". **That control does not exist for Phase 28.** `session_phase28_manifest_test.go` has only `TestPhase28EntryExitClassManifest`, with the compare logic inlined in the test body at lines 580-608. Plan 32-02 must first extract a pure `phase28Compare(golden, observed) []string`, mirroring `phase30Compare`, and then add the control. The current "no longer observed" loop is:

```go
	for path := range goldenByPath {
		if _, ok := observedByPath[path]; !ok {
			t.Errorf("%s: in the golden at %s but no longer observed", path, golden.Revision)
		}
	}
```

### Pattern 2: Executed guide, README is the only source of commands and numbers

[VERIFIED: examples/checksum/README.md:31-36, 65-71, 113-117] The README currently holds unmarked ` ```sh ` fences and a free-text results table (`| empty file | 1 | 1 | python… |`). Required changes:
- Mark the runnable fences. Recommended info-string: ` ```sh doc-test=clean-checkout ` for the one-time setup block (go build, `mkdir -p .schway/checksum`, the three `./schway build` lines), and ` ```sh doc-test=run app=checksum ` for the templated line `./schway app run .schway/checksum/checksum -- PATH`. Same for `adler_low` and `adler32_cli`.
- Replace the first column of the results table with the stable key from `checksum.expected.json` (`empty`, `zero`, `wikipedia`, `ff65536`, `ff65537`, `missing`, `directory`). Keep one column per app (`checksum`, `adler_low`). Keep the Python command column for provenance.
- Inputs: materialise by key from `testsupport.ChecksumCaseMatrix()` (it carries `Name`, `Kind`, `Content`). The expected JSON has no input bytes, so this is the single source and keeps the doc-test free of byte literals.
- `adler_low` is not in `checksum.expected.json`. Cross-check each `adler_low` cell two ways: `checksum % 65536` (the README states this rule) and the bytes-keyed `testsupport.ChecksumOracleRows()` (`App == "adler_low"`). Do this via input bytes, because oracle names differ (`Wikipedia`, `65536x0xFF`).
- The `adler32_cli` demo is a case-id program, not a file reader. Its rows are `3 -> 300286872` and `10 -> exit 65, empty stdout, empty stderr`. I ran both against a build this session. `testsupport.Phase29FixtureExpected(caseID)` supplies values for ids 0 to 9 if the guide lists more.
- Output targets must pass the Phase 27 rule: every `--output` (plus its `.schway-build.json` sidecar) and every `app run` artifact must be under `.schway/`, and a `mkdir -p` must precede the first build. The doc-test should assert this on the parsed block (copy `phase273GuideOutputTargets`).

### Anti-Patterns to Avoid

- **A doc-test with its own mirror of numbers** (rejected by D-01). The parser must fail loudly if the table has no rows.
- **Passing the whole tempdir path into error assertions.** Assert that the *segment* (for example `ChecksumSecretSegment`) is absent from stderr, as the Phase 31 matrix does.
- **Running `go build` in the real repo** (D-02 rejects worktree and archive), and **using `-trimpath`**. Native helper sources resolve through `runtime.Caller` at build time (see Pitfall 6), so the copy must outlive the binary's run.
- **Re-using `PHASE27_*` env names or phase27 row prefixes in Phase 32 logs.** The phase27 gate treats any `phase27 ` row of an unknown kind as forgery.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Bounded, timeout-guarded subprocesses | A raw `exec.Command` | `testsupport.RunCLIErr` / `BuildCLIErr` | `scanUnboundedSpawns` in `native_test.go` fails any spawn lacking `CommandContext`, a timeout context and bounded writers. |
| Input materialisation per case | Inline byte literals in the new test | `testsupport.ChecksumCaseMatrix()` | Already independently verified against Python zlib and Go `hash/adler32`. |
| Fixture-case expected values | New constants | `testsupport.Phase29FixtureExpected`, `ChecksumOracleRows` | One source of truth. |
| Output-path guard | A new checker | Copy `phase273GuideOutputTargets` and `phase273TargetAllowed` | Already has bypass controls (`..`, `>>`, `--output=`). |
| Hash of a tree revision | Parsing `.git` objects | `git ls-tree -r HEAD -- <paths>` in both sh and Go | Same bytes in both implementations; a Go reimplementation of tree walking would drift. |
| Test-name resolution for SUMMARY coverage | A new scanner | `buildTestIndex` (static `go/parser` index) | `TestSummaryCoverageGroundedness` already uses it. |

**Key insight:** every Phase 32 mechanism already has a precedent in-repo (Phase 27 scripts, Phase 28/30 allow-lists, Phase 273 guide checker, Phase 25 living-doc test). The work is wiring and extending, and the risk is in the existing pins that these precedents sit behind.

## Findings by research question

### 1. Formatter bug [VERIFIED: ran in scratch copy]

**Root cause.** `Format` (`format.go`) breaks the line before an identifier only when that identifier is followed by `=` (lines 34-39, the loop quoted in the todo's area). Quote:

```go
if token.Kind == TokenIdentifier && next == TokenEqual && index > 0 && f.lineOpen && f.insideFunctionBody() {
	previous := tokens[index-1]
	if previous.Span.End <= token.Span.Start && bytes.Contains(tree.Source[previous.Span.End:token.Span.Start], []byte("\n")) {
		f.newline()
	}
}
```

For `var r` bodies, `linearBinding` is false (only `let` sets it, `format.go:158`). So after `r = r + a`, nothing ends the line, and the next bare result identifier `r` is appended with no separator: `r = r + ar`. Source whitespace tokens are dropped at the top of `Format`. The same defect occurs after a number (`r = r % 7` then `r` gives `7r`) and after a call's closing paren (`var x = byte_at(bytes, n)` then `x` gives `byte_at(bytes, n) x`). The first reproduction on `examples/checksum/adler32.schway` printed `r = r + ar`, matching the todo.

**Minimal fix (prototype, passes).** Extend the loop with a second guard just before `f.token(token, next)`:

```go
if token.Kind == TokenIdentifier && index > 0 && f.lineOpen && f.insideFunctionBody() && f.context() != "match" {
	previous := tokens[index-1]
	if (previous.Kind == TokenIdentifier || previous.Kind == TokenNumber || previous.Kind == TokenRParen) &&
		previous.Span.End <= token.Span.Start && bytes.Contains(tree.Source[previous.Span.End:token.Span.Start], []byte("\n")) {
		f.newline()
	}
}
```

The `f.context() != "match"` clause is **required**. My first attempt without it broke match arms (`Ok(v)` newline `Err(e) => Err` was split across lines in `testdata/phase12/payload_tracer.schway`). With the clause, a comparison of old versus new output over all 163 committed `.schway` files shows exactly 5 files change:

| File | In frozen phase30 golden? |
|------|---------------------------|
| `examples/checksum/adler32.schway` | no (allow-listed new path) |
| `examples/checksum/adler_low.schway` | no (allow-listed new path); not named in the todo |
| `examples/phase29/adler32_fixture.schway` | **yes** |
| `examples/phase29/byte_index_oob.schway` | **yes** |
| `testdata/phase29/adler32_library.schway` | **yes** |

Prototype `format_sha256` values (non-authoritative; recompute after the real fix): `b2824fdd0f08960277fbc1a1d1670d373bb169f38c93995d0b60427e96c51ca3`, `736a53528232a598ba5d5d51710e7bf824c80c8997cee56db4be357a82e9b6e4`, `ec1f3a40c576cc19d47563b5e5e76ca2213de1ad35a578845da9169c94237a36`. The prechange test reported exactly those three plus the frontier path in the full-suite experiment.

Formatted copies of `adler32.schway` and `adler_low.schway` built and ran correctly (`Wikipedia` gives 300286872 for `checksum` and 920 for `adler_low`), and `format --check` passes on the result. The formatted output is not byte-identical to the committed files: the formatter inserts a blank line after the leading comment. This is consistent with D-10, which says not to claim canonical form.

The cosmetic `}` newline `else {` split (noted in the 29-02 SUMMARY) is not semantic and is left alone. Changing it would move more golden hashes.

**Lanes to run.** `go test ./internal/compiler/syntax -count=1` (includes `TestFormatIdempotent`, `TestGeneratedRoundTrips`, `TestGeneratedLinearRoundTrips`, `TestGeneratedBranchBodiesRoundTrip`, `TestPhase29FormatByteReads`, `TestFormatLetBinaryExpressions`, `TestPhase30FormatImport`). Also `go test ./internal/compiler/session -run 'TestPhase27FormatterRemainder|TestPhase27LiteralEscapeFormat|TestPhase30PreChangeGolden' -count=1` and `go test ./cmd/schway -count=1`.

**Coverage gap in the existing lanes.** The generators emit `let`-style linear bodies, so no generated program ends a `var` body with a bare result. That is why the bug survived 1000-case properties. Add either a `generatedVarProgram(caseID)` generator with the same property body as `TestGeneratedLinearRoundTrips` (reparse, fixed point, `semanticTokens` and `withoutSpans` equality, same `session.Check` diagnostic codes), or at minimum a table test with terminal tokens identifier, number and call-paren. The five committed sources should be read from the repo and formatted **as in-memory copies**, never rewritten.

### 2. Frontier retirement [VERIFIED: grep + deletion experiment]

**Every Go reference to the old path.** Only `internal/compiler/session/session_phase20_test.go:15` (`testsupport.ProjectPath("examples", "checksum.schway")`) references the top-level file. Every other hit of `checksum.schway` in Go is `examples/checksum/checksum.schway` or a bare `checksum.schway` inside a copied dir. No shell script or workflow references it. Non-Go references: `testdata/phase28/entry-exit-class-manifest.json`, `testdata/phase30/prechange-corpus.json` (frozen), and live docs `LANGUAGE-MATURITY.md:297,412`, `PRODUCT-ROADMAP.md:272,298`, `PROJECT.md:353`, `examples/checksum/README.md:140`. **`PRODUCT-ROADMAP.md:272` is missed by CONTEXT D-07** (it is the Phase 31 non-claims line naming the file). `REQUIREMENTS.md:51` and `ROADMAP.md:23,262` also name it; both are status-only updates at close.

**Experiment.** I copied the tree to scratch, deleted the file, applied the formatter prototype and ran `go test ./...`. Failures:

| Test | Cause | Fix owner |
|------|-------|-----------|
| `TestPhase20ChecksumFrontier` | opens the deleted file | replaced by `TestPhase32ChecksumFrontierRetired` |
| `TestPhase28EntryExitClassManifest` | `examples/checksum.schway: in the golden at c2a6f719… but no longer observed` | `phase28RetiredEntries` |
| `TestPhase30PreChangeGolden`, `TestPhase30PreChangeGoldenIsNotInert` | frontier path absent, plus the three `format_sha256` changes | `phase30RetiredPaths` + `phase30FormatChanges` |
| `TestLanguageMaturityCountsAreCurrent`, `TestSelfDescribingDocsGuardIsNotInert` | `LANGUAGE-MATURITY.md:304` says "163 programs, 5,389 lines"; re-derived values are 162 and 5,352 | update the counts in the retirement plan |
| `TestReconciliationVerdictsCarryTheirObligations` | **artifact of my scratch copy**: its fresh `git init` lacks historical commits | ignore |

The inventory, 272/273 identity-corpus, exit-manifest (other than the frontier), groundedness and summary tests passed unchanged. `TestPhase272AcceptedCorpusByteIdentity` and `TestPhase273ConformanceCorpusByteIdentity` do not trip because the frontier was never in the accepted corpus (it refuses). The corpus count 5,352 assumes the retirement plan adds **no** new `.schway` file. Any added example changes these counts again.

**Walker shape.** `phase28ObserveCorpus` and `phase30Observe` walk `examples/` and `testdata/` with `filepath.Walk` for `.schway`. Minimums (`phase28MinCorpus = 155`, `phase30PreChangeMinCorpus = 160`) are comfortably below the post-delete count of 162.

**Hygiene-scan trap.** D-05 requires "no live line naming the old path without 'retired' or 'amendment'", yet D-07 leaves `LANGUAGE-MATURITY.md:180` (inside the dated "### M005 kickoff — 2026-10-02" entry) untouched, and that line names the path with "remains a provisional refused loop witness". A naive scan fails on it. Define "live" structurally: skip any line under a heading that carries an ISO date (`^#{2,4} .*\d{4}-\d{2}-\d{2}`), and add a control that proves the skip cannot hide a live row (move a naming line out of a dated section and expect a finding). Confirm which of `:297`/`:412` fall under dated headings before relying on this.

### 3. Guide doc-test mechanics [VERIFIED]

- **README structure.** Quoted above under Pattern 2. Sections: Layout, `## checksum` (one fence), `## adler_low` (one fence), `## Results`, `## Errors`, `## Single copy and identity`, `## Phase 30 fixture CLI` (one fence), `## Non-claims`. The line-140 note says "The provisional `examples/checksum.schway` frontier is unchanged and owned by Phase 32… Hosted CI has not run for Phase 31". It must become a pointer to the PROJECT.md amendment.
- **`checksum.expected.json`** [VERIFIED: examples/checksum/checksum.expected.json:1-9]: keys `schema` (`"schway.checksum-expected/1"`), `limit` (`65536`), `error_exit` (`{"code": 65, "name": "AppInputRefused"}`), `outcomes` (`FileTooLarge`, `FileUnavailable`, `AllocFailed`), and `cases` with `name`, `kind`, `exit`, `stdout`, `stderr`, `outcome`, `engines`, `witness`. Case names: `empty`, `zero`, `wikipedia`, `ff65536`, `ff65537`, `missing`, `directory`, `fifo`, `devnull`, `mode000`, `alloc-failed`, `read-eio`. The doc-test uses the first seven (D-03). `TestPhase31ExpectedMatrixJSON` pins the JSON to `ChecksumCaseMatrix`. If the doc cross-check sits beside it in `testsupport`, the new test can reuse the same JSON struct. In `cmd/schway` it would copy the struct. Recommendation: put the **JSON-vs-README** check (`TestPhase32GuideMatchesExpectedJSON`) in `cmd/schway` next to the executor, and keep `testsupport` untouched.
- **`phase273CleanCheckoutBlock` style.** Extraction by `strings.Cut` on a heading and a fence; D-01 changes the fence to a marked info-string. Parse by scanning lines for a fence open matching `^```sh doc-test=(\S+)(?: (.*))?$` through the next ` ``` ` line. Join `\` continuations before use (the README lines for `build` wrap with a trailing `\`).
- **How `cmd/schway` tests build the binary.** Existing tests call `run(args)` in-process (`captureLangRun`) and never exec the binary. `testsupport.BuildCLI(t)` and `BuildCLIErr` build `./cmd/schway` with `Dir = ProjectPath()` (source root, 5-minute timeout). The doc-test instead runs the guide's own `go build -o ./schway ./cmd/schway` inside the copy. That build took about 1.2 s with a warm cache in a different directory (cache keys are content-based), so only a cold hosted cache is a concern, and `go vet`/`go build ./...` precede `go test` in the `checks` job.
- **Spawn constraints.** `RunCLIErr` has a hard 2-minute timeout (`RunCLITimeout`), bounded 1 MiB streams and no `Dir`. Use `RunCLIErr(ctx, "sh", env, "-c", `cd "$0" && eval "$1"`, copyDir, line)`. For `git ls-files -co --exclude-standard` use `git -C <root>`. Tracked-but-deleted paths appear under `-c`, so skip `ENOENT` when copying. A copy of the whole listing is about 20 MB of `.planning`; filter it out if desired, since the guide needs only `cmd/`, `internal/`, `native/`, `examples/` and `go.mod`. There is no `go.sum` (stdlib only), so no network is needed.
- **CI clang and env.** [VERIFIED: .github/workflows/ci.yml] Both `checks` and `evidence-aggregate` have a `Require clang` step running `clang --version`, and Ubuntu installs `libc++-dev libc++abi-dev`. There is no `CI`, `GITHUB_ACTIONS` or `SCHWAY_REQUIRE_*` reference in any Go or shell file (grep returned nothing outside `.planning`). Existing native tests skip with `t.Skip("env:clang toolchain unavailable on this host")`. Recommended mechanism: a small pure function `phase32NativePolicy(clangErr error, getenv func(string) string) (fail bool)` that fails when `SCHWAY_REQUIRE_NATIVE=1` **or** `GITHUB_ACTIONS=true`, and skips otherwise. The pure function is testable without removing clang. `GITHUB_ACTIONS=true` is set on hosted runners [ASSUMED: standard GitHub default env, not re-fetched this session]; add `env: SCHWAY_REQUIRE_NATIVE: "1"` to the two `go test` steps only if explicitness is preferred. Editing `ci.yml` changes the codedigest, so decide this before the freeze.
- **`git` in CI tests.** Present: `TestPhase27HostReceiptGate` runs `git hash-object`, `TestPhase27GuideContract` runs `git rev-parse HEAD`, and `TestPhase273GuideCleanCheckoutOutputsIgnored` runs `git check-ignore`, all through `RunCLIErr`, and `fetch-depth: 0` is set. Pattern for source archives: fall back, never fail.
- **Timeouts.** Default `go test` package timeout 10 minutes; the race lane passes `-timeout=20m`. The doc-test is three `clang` builds plus about 12 runs; the subprocess `go build` is not instrumented with `-race`. Budget: well under a minute warm.
- **Dirty-tree hazard.** `.gitignore` ignores only `/.schway/`, `/schway` and `/.gsd/`. Guide outputs at the repo root (`./checksum`, `./adler_low`, `./adler32-module` and their `*.schway-build.json` sidecars) would leave untracked files that `git status --porcelain` reports. That makes a receipt run `tree=modified`. All guide outputs go under `.schway/`.

### 4. Receipts [VERIFIED: read both scripts, ran `gh`, downloaded the artifacts]

**What exists.** `verify-phase27.sh` builds `schway`, builds only `fizzbuzz` and `sum_to_n`, runs 10 cases and 3 `go test` controls, and prints rows `phase27 host identity …`, `phase27 toolchain …`, `phase27 case …`, `phase27 control …`, `phase27 host result …`. Revision is `git rev-parse HEAD` (line 33), and `tree=` is `git status --porcelain` at *start only* (line 42). `check-phase27-receipts.sh` takes exactly two logs, rejects on named reasons (reject() lines), recomputes `git hash-object` for sources and fixtures, and requires exact per-case row text. Its header says: "This gate checks consistency, not authenticity".

**What the hosted logs cover today.** I downloaded the artifacts of run 38299800002 (success, `workflow_dispatch`, `headBranch` `phase26-validation`, `headSha` `c72569244f67d68bb636a184fac285ca17035429`). The two logs contain only the Phase 27 rows: FizzBuzz and `sum_to_n`. **None of the M006 programs are in any hosted receipt log yet**, which is EVD-15's gap. `gh run view <id> --json headSha,conclusion,jobs` returns job names such as `current evidence aggregate (ubuntu-latest)` with per-job `conclusion`.

**Why not generalize in place.** The Phase 27 gate is pinned by `TestPhase27HostReceiptGate` (a literal known-good log builder plus about 25 mutation rows keyed to exact row text), `TestPhase27CIHostMatrix` (workflow text) and `TestPhase27GuideContract` (re-runs the gate on committed M005 receipts at pinned revision `6224f8b2…` with `PHASE27_EXPECTED_REVISION`). Add new scripts, and leave all three tests untouched.

**Recommended Phase 32 receipt log grammar** (prefix `phase32 `, values contain no spaces, as the phase27 `field()` parser requires):

```
phase32 host identity host=<macOS|Linux> arch=<..> go=<os/arch> clang_target=<..> headsha=<git rev-parse HEAD> codedigest=<sha256> tree=<clean|modified>
phase32 toolchain go="<go version>" clang="<first line of clang --version>"
phase32 case program=<checksum|adler_low|adler32_cli|sum_to_n|fizzbuzz> input=<key> expected_exit=<n> actual_exit=<n> stdout=exact stderr=exact status=pass
phase32 control name=<n> package=<pkg> tests=<names> status=pass
phase32 host result host=<..> codedigest=<sha256> tree=<..> cases=<n> controls=<n> programs=checksum,adler_low,adler32_cli,sum_to_n,fizzbuzz status=pass
```

The producer runs the programs itself, as the phase27 `run_case` does (copy that function; a copy beats a new dependency). `sum_to_n` and FizzBuzz reuse `examples/phase27/fizzbuzz-15.out` and `-1000.out` with `cmp`. The M006 programs reuse committed manifests (`file_bytes.bindings.json`, `fixture_bytes.bindings.json`). Record `tree=` at both start and end (an improvement over phase27, which checks only the start) and fail if the tree moved. Controls: run the new Go tests with `go test -v -count=1 -run …` and require `--- PASS:` lines and no `--- SKIP:`, exactly as `run_control` does. The skip check matters because a clang-less run would skip native tests.

**CI wiring.** Add one step after the phase27 step in `evidence-aggregate`, in the same shape (`shell: bash`, `sh scripts/verify-phase32.sh 2>&1 | tee "$RUNNER_TEMP/phase32/receipt-${{ matrix.os }}.log"`), and an upload step named `phase32-receipt-${{ matrix.os }}-${{ github.sha }}` with `if: ${{ always() }}`, `if-no-files-found: error`, `retention-days: 14`. `shell: bash` runs with `-eo pipefail` on GitHub-hosted runners [ASSUMED: GitHub's documented default for `shell: bash`; the existing phase27 step relies on it]. The existing pins are inclusion-only (`strings.Contains`), so extra steps do not break `TestPhase27CIHostMatrix` or `TestCIWorkflowRunsCurrentAggregateGate`. Keep exactly one `sh scripts/verify-phase23.sh` and do not add `continue-on-error`.

**codedigest, one algorithm in three places.**

```sh
# scripts/codedigest.sh  (sourced or executed by verify, gate and a Go test)
git -c core.quotePath=false ls-tree -r HEAD -- cmd internal native examples scripts testdata go.mod .github/workflows/ci.yml \
  | grep -v '\.md$' | { shasum -a 256 2>/dev/null || sha256sum; } | cut -d' ' -f1
```

Go side: run the **same git command** via `RunCLIErr`, drop lines with `strings.HasSuffix(line, ".md")`, join with `"\n"` plus a trailing newline, then `sha256`. Git emits entries in tree order, which is deterministic, and both implementations consume identical bytes, so there is no sorting step to diverge. I checked that no tracked path under these directories is quoted or non-ASCII (zero matches), and with the D-08 path set the digest is over 474 entries (805 with `native` and `testdata`). Armed controls: edit one `.go` byte in a temp clone (digest moves), edit only a `.md` file (digest does not move), edit only `.planning/` (does not move). Pass the sh digest and the Go digest through an equality test.

**Pitfall: `ls-tree` reads HEAD, not the worktree.** Hence the separate `tree=clean` requirement. Also hosted `pull_request` events check out a *merge* commit, so `HEAD` differs from the branch tip. Dispatch on the branch ref only (as Phase 31 did).

**Gate reasons** (for `check-phase32-receipts.sh LOG_A LOG_B`, expected digest defaulting to the checkout's computed value, overridable with `PHASE32_EXPECTED_CODEDIGEST` as the phase27 gate does): `digest_mismatch_<log>`, `log_digest_differs` (the two logs disagree), `tree_not_clean_<log>`, `duplicate_host_<h>`/`host_missing`, `program_missing_<name>_<log>`, `case_missing_or_altered_…`, `control_*`, `unknown_phase32_row_<log>`, `identity_row_count_…`, `result_not_last_…`. Reuse the phase27 mutation-table test structure for the armed controls D-09 lists: mutate one code file, mutate the digest, drop a host, flip `tree=modified`. Add: drop one program's rows; swap one log's `programs=` list.

**Receipt JSON (`32-RECEIPT.json`).** Suggested keys: `schema` (`"schway.m006-receipt/1"`), `state` (`"current"` or `"reopened"`), `runId`, `headSha`, `codedigest`, `digestPaths`, `hosts` (each: `name`, `runnerOs`, `arch`, `go`, `clang`, `logPath`, `logSha256`, `jobConclusions`), `programs` (each: `name`, per-host `status`). Commit the two logs next to it (about 20 lines each) so the Go test can recompute `logSha256`. The Phase 27 precedent committed `host-macos.receipt` and `host-linux.receipt` the same way.

**Phase 25 reopen precedent.** [VERIFIED: internal/compiler/native/phase25_utility_test.go:293-306] `TestPhase25LivingRoadmap` is a plain text check: both living docs must contain a set of words including "Historical receipts", and must not contain the exact sentence "Phase 25 passed on Linux and macOS". It has no digest or run binding. D-10a is a stronger rule built on the same idea. Implement it as: choose one fixed living-claim marker (for example the line `M006 final hosted receipts: bound to codedigest <hex>` in `LANGUAGE-MATURITY.md`, `PRODUCT-ROADMAP.md` and `STATE.md`, or `... reopened <date>`). Test: no marker means pass; marker present means `32-RECEIPT.json` exists, its logs hash correctly, every marker digest equals the receipt digest, and the receipt digest equals the freshly computed one; a stale digest with a "bound" marker fails, while "reopened" passes. Include an armed control on in-memory docs.

**The ordering trap this creates (important).** The receipt test lives in `internal/` and therefore *is inside the digest*. It must be committed before the freeze, and it will run in the hosted run when no receipt exists yet. So it must **pass when the receipt is absent and no living doc claims bound evidence**. Do not use `t.Skip` for absence. The ingest plan then adds the receipt, logs and marker lines in one `.planning`-only commit, and the same test starts enforcing. This also keeps the ingest commit from moving the digest.

### 5. Public-route audit [VERIFIED: read `cmd/schway/main.go`, ran `examples` listing]

- **Subcommands.** They exist only as the `if` chain in `run()` (`main.go:33-126`) and in `runApplication` (`app run`, `app verify`). The set is: `app` (run, verify), `build`, `explain`, `query`, `format`, `check`, `evidence`, `verify`, `stats`, `interface` (export, core, check), `debug-map`, `run --engine=…`. The usage string in `usageResult()` (`main.go:679`) **omits `stats` and the `build … --manifest` form**. Both exist in `run()`. This is itself an audit finding; record it as a "not covered because" or a follow-up todo, and do not edit the usage text in this phase (no new CLI capability; usage text is likely pinned).
- **Enumerate without drift.** In a `package main` test, parse `main.go` with `go/parser` and collect string literals compared against `args[N]` in `run`, `runApplication` and `runApplicationRun` (binary `==` expressions whose other side is `args[<int>]`). Collect `"--…"` literals from the `extract*` helpers (`--json`, `--depth=`, `--kind=`, `--cursor=`, `--expand`) and from the dispatch (`--manifest`, `--output`, `--check`, `--validate`, `--engine=`, `--report`, `--evidence=`, `--cases`). Cross-check against `usageResult()` tokens only for the "usage mentions something dispatch lacks" direction. Fail if the AST set contains a route with no row in the audit table.
- **Entry programs.** Walk `examples/` for `*.schway` and treat a file as an entry program when `syntax.Parse` finds a function named `main`. Avoid `callgraph.EntryFunction`, because `TestEntryFunctionCalledOnlyViaResolveEntry` restricts its callers. Currently 15 files have `fn main` (including `examples/checksum.schway` until it is deleted; `adler32.schway` has none). Receipt rows cover `checksum/checksum`, `checksum/adler_low`, `checksum/adler32_cli`, `sum_to_n`, `fizzbuzz`. Every other entry (`phase22/identity`, `phase23/file_byte`, `phase24/error`, `phase24/transfer`, `phase26/checked_add_overflow`, `phase272/straight_line_copy`, `phase28/unused_function`, `phase29/adler32_fixture`, `phase29/byte_index_oob`) needs a "not covered because" row with a real reason, such as "earlier milestone, own hosted gate", "negative or refusal fixture", or "fixture for a golden". Don't invent reasons: for `phase23`/`24`, `verify-phase25.sh` references `examples/phase24/transfer.schway` and `examples/phase23/*`; the rest must be justified from the phase READMEs.
- **Spawn and emitter inventories.** The audit must not call `cgen` emitters (the Phase 16 public-emitter inventory tracks every call, including in tests) and must not call `callgraph.EntryFunction`.
- **Receipt coupling.** The receipt test asserts that the programs in `32-RECEIPT.json` equal the audit's "covered by receipt" set, so adding an example without a row fails the audit and a row without a receipt program fails the receipt test.

### 6. Session meta-tests [VERIFIED: read the tests]

What a new Phase 32 test or document must satisfy:

| Meta-test | Rule | Phase 32 consequence |
|-----------|------|----------------------|
| `TestPhase20ValidationLifecycle` | no `*-VALIDATION.md` (live or archived) may have `status: draft` | `32-VALIDATION.md` must be committed as `validated`, not `draft` (the 31-VALIDATION header says the same). |
| `TestVerificationGroundedness*` | every `go test … -run PATTERN` span in an **enforced-tier** document under `.planning/**` must resolve; every alternation branch must match an existing test (classes R2/R2b) | `*-VALIDATION.md`, `*-VERIFICATION.md` and any doc with a command-plus-verdict table row are always enforced. `*-RESEARCH.md`, `*-PLAN.md`, `*-CONTEXT.md` are exempt **only** with `verification_role: proposal` in frontmatter (this file has it; the PLAN files need it too). A VALIDATION command naming a test that does not exist yet fails the lint, so either land the test in the same commit or use only patterns that already resolve. Phase 31 solved this by putting per-test names in plain code spans (not `go test` commands) and keeping the one runnable `-run 'Phase31'` pattern resolvable. |
| `TestSummaryCoverageGroundedness` (`summary_coverage_test.go`) | every SUMMARY from phase 32 on needs a `coverage:` list; each entry has `human_judgment: false` with at least one `ref: "<file>#<TestName>"` whose `status: pass` and whose test exists in the static index, or `human_judgment: true` with a `rationale`; `coverage: []` is accepted | Each Phase 32 SUMMARY must name real test functions. The push-and-dispatch checkpoint plan's SUMMARY is the only `human_judgment: true` entry, with the rationale "push and dispatch need explicit user approval". |
| `TestPhase16PublicEmitterConsumerInventory` | any call to a public `cgen` emitter, in non-test or test files, must be registered | Avoid emitter calls. None are needed. |
| `TestEntryFunctionCalledOnlyViaResolveEntry` | `callgraph.EntryFunction` called from exactly the registered sites | Avoid. |
| `TestSourceNeverSpawnsUnboundedProcesses` (native) | AST lint on `os/exec` use | Spawn only through `testsupport`. If a test must import `os/exec` (for `exec.LookPath("clang")`, as the existing tests do), that is fine, because `LookPath` is not a spawn. |
| `TestStaticTestIndexMatchesGoTestList` | the static test index equals `go test ./... -list .` | Use plain top-level `func TestX(t *testing.T)`; no build-tag tricks. |
| `TestLanguageMaturityCountsAreCurrent` | corpus counts in `LANGUAGE-MATURITY.md:304` equal the re-derived values | Update to 162 programs and 5,352 lines in the retirement plan. Re-derive again at the end if any `.schway` file changes. |
| `TestEvidenceReconciliation*` | archived VALIDATION or VERIFICATION edits must be coupled to `.planning/EVIDENCE-RECONCILIATION.md` | Do not edit archived `milestones/**` validation docs. |

The meta set as written in `31-VALIDATION.md:26`, reusable verbatim:
`go test ./internal/compiler/session -run 'Inventory|Groundedness|Reconciliation|ValidationLifecycle|TestPhase28EntryExitClassManifest|TestPhase272AcceptedCorpusByteIdentity|TestPhase273ConformanceCorpusByteIdentity|TestPhase30PreChangeGolden|TestLanguageMaturityCountsAreCurrent|TestExitCodeDocsTable|TestNoSuppressionOutlivesItsWitness|TestPhase16PublicEmitterConsumerInventory' -count=1`. Add `TestSummaryCoverageGroundedness` (the `Groundedness` alternation already matches it). Note `TestReconciliationVerdictsCarryTheirObligations` needs the real git history, so it cannot run in a fresh-`git init` copy.

## Common Pitfalls

### Pitfall 1: The formatter fix silently moves a frozen golden
**What goes wrong:** `TestPhase30PreChangeGolden` fails on three `format_sha256` fields after a correct fix.
**Why it happens:** the golden pins `syntax.Format` output of every committed `.schway`, including three files that end a `var` body with a bare result.
**How to avoid:** add the dated `phase30FormatChanges` map in the same plan as the fix; never regenerate the golden (the file header says "The golden is never regenerated after a production change in this phase").
**Warning signs:** the failure text names `format_sha256 changed` for `examples/phase29/*` or `testdata/phase29/adler32_library.schway`.

### Pitfall 2: An over-broad formatter guard breaks match arms
**What goes wrong:** `Ok(v)` and `Err(e) => Err` end up on separate lines, changing 9 `testdata/phase12`/`phase18` outputs.
**Why it happens:** `)` followed by an identifier on a new source line is legal in a `match` body.
**How to avoid:** keep the `f.context() != "match"` clause; assert in the regression test that the full-corpus format output is unchanged for every file outside the five known ones.

### Pitfall 3: Guide outputs dirty the tree
**What goes wrong:** the README commands write `./checksum`, `./adler_low` and `./adler32-module` plus sidecars into the repo root. They are untracked and not ignored, so a receipt run after a guide run reports `tree=modified`, and the gate rejects it.
**How to avoid:** route all outputs to `.schway/<name>/…` with a `mkdir -p` first, and have the doc-test assert it with the copied target checker.

### Pitfall 4: The receipt test or script placed after the freeze
**What goes wrong:** committing a Go test or script in the ingest plan moves the codedigest, so the receipt would immediately be stale.
**How to avoid:** every file under `cmd internal native examples scripts testdata go.mod ci.yml` is final at the freeze. The ingest plan commits only `.planning/**` (JSON, logs, docs). Add a check in the ingest plan: recompute the digest and compare with the hosted log before committing anything.

### Pitfall 5: Hosted dispatch cancels itself or runs a merge commit
**What goes wrong:** `ci.yml` has `concurrency: ci-${{ github.ref }}` with `cancel-in-progress: true`, so a second dispatch on the branch cancels the first. A `pull_request` run checks out a merge commit with a different tree.
**How to avoid:** dispatch once, on the branch ref, and record `headSha` from `gh run view`. Verify that `headSha` equals the pushed tip and that the log `codedigest` equals the checkout's computed digest before ingesting.

### Pitfall 6: Native helper sources resolve by build-time source path
**What goes wrong:** `native/schway_foreign_*.c` are located through `runtime.Caller` (`foreign_retained.go:15-17`), baked into the binary at build time. A binary built in the temp copy refers to the copy. Deleting the copy early, or building with `-trimpath`, breaks it.
**How to avoid:** keep the copy alive for the whole test; do not pass `-trimpath`.

### Pitfall 7: Groundedness lint on new planning docs
**What goes wrong:** `32-VALIDATION.md` lists `go test … -run 'TestPhase32…'` before the tests exist, or `32-NN-PLAN.md` lacks `verification_role: proposal`, and `TestVerificationGroundednessFrontierIsPinned` reports an unexpected violation.
**How to avoid:** see the meta-test table. Run the meta set before each commit that touches `.planning`.

### Pitfall 8: A path-substituting `sh -c` and shell metacharacters
**What goes wrong:** the temp path (macOS `/var/folders/…`) is substituted into a shell line; spaces or quotes in it could split the command.
**How to avoid:** substitute only the exact token `PATH`, as a single-quoted value (`'`+path+`'`, escaping `'`), and never substitute into the build lines.

## Code Examples

### Guide fence parser (shape)

```go
// Source: modelled on phase273CleanCheckoutBlock (internal/compiler/session/session_phase273_docs_test.go:30-46)
var fenceOpen = regexp.MustCompile("^```sh doc-test=([a-z0-9-]+)((?: [a-z]+=[a-z0-9_]+)*)$")

type guideBlock struct {
	Kind, App string
	Lines     []string // `\` continuations joined
}
```

### Running a guide block in the copy without a new helper

```go
// RunCLIErr has no Dir; make the copy the shell's cwd. $0 and $1 are passed as arguments, never interpolated.
res, err := testsupport.RunCLIErr(ctx, "sh", os.Environ(), "-c", `cd "$0" && eval "$1"`, copyDir, line)
```

### Native-required policy, testable without removing clang

```go
func phase32NativePolicy(clangErr error, getenv func(string) string) (mustFail bool) {
	return clangErr != nil && (getenv("SCHWAY_REQUIRE_NATIVE") == "1" || getenv("GITHUB_ACTIONS") == "true")
}
```

### Retired-path rule inside the existing compare function (shape)

```go
// phase30Compare, golden loop, when the golden path is not observed:
if !ok {
	if phase30RetiredPaths[want.Path] { continue }          // dated allow-list, D-06
	problems = append(problems, want.Path+": recorded at golden revision "+golden.Revision+" but no longer present")
	continue
}
// ...and in the observed loop:
if phase30RetiredPaths[entry.Path] {
	problems = append(problems, entry.Path+": retired path observed again")
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Receipt bound to commit SHA (`PHASE27_EXPECTED_REVISION`) | Bound to `codedigest` over code paths | Phase 32 (D-08) | Docs-only commits and the receipt commit itself no longer invalidate receipts. |
| Guide test is static (`TestPhase273GuideCleanCheckoutOutputsIgnored` declines to execute) | Guide is executed from a temp copy | Phase 32 (D-01/D-02) | The README becomes the test input, with no second source of numbers. |
| Frozen frontier refusal pinned by `TestPhase20ChecksumFrontier` | Deleted, with dated retire allow-lists on frozen goldens | Phase 32 (D-05/D-06) | The corpus no longer carries a file that looks supported but refuses. |

**Deprecated/outdated:** the README line-140 note "The provisional `examples/checksum.schway` frontier is unchanged and owned by Phase 32" and the "Hosted CI has not run for Phase 31" sentence (Phase 31 hosted run 38299800002 passed).

## Runtime State Inventory

This phase deletes one file and renames nothing, but it retires a path that other state may reference. All five categories, answered:

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | Two frozen golden JSON files in git hold the old path (`testdata/phase28/…`, `testdata/phase30/…`). No databases. | Code edit: allow-lists only; goldens stay byte-frozen. |
| Live service config | None. The only external state is GitHub Actions run history (run 38299800002 predates the deletion). | None. |
| OS-registered state | None. `/schway`, `/.schway/` are git-ignored build outputs, not registrations. | None. |
| Secrets/env vars | None for this retirement. Phase 32 adds at most one optional env var (`SCHWAY_REQUIRE_NATIVE`). | Optional `env:` in `ci.yml`. |
| Build artifacts | Stale `./schway`, `./checksum`, `*.schway-build.json` files in a developer's repo root from the old README flow. The old guide wrote them to the root. | None in repo. Mention in the guide that outputs now live under ignored `.schway/`. |

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | all tests | yes | go1.24.0 (local), go1.24.13 (hosted) | none needed |
| clang | guide doc-test, receipts | yes | Apple clang 21.0.0 (local); Ubuntu clang 18.1.3, Apple clang 21.0.0 (hosted logs) | local skip; hosted must fail (D-04) |
| git | doc-test copy, digest, gate | yes | system | file-walk fallback for archive checkouts (copy); digest tests skip without `.git` |
| `shasum`/`sha256sum` | `codedigest.sh` | yes (`/usr/bin/shasum`; `sha256sum` also present) | system | probe in the script, and reject with a named reason if neither is found |
| `gh` | ingest only | yes | 2.101.0 | none; the ingest plan is human-adjacent |
| Network / push rights | hosted run | n/a | n/a | the push and dispatch line is the one human step |

**Missing dependencies with no fallback:** none.

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing`, go 1.24 |
| Config file | none (`go.mod` only) |
| Quick run command | `go test ./internal/compiler/syntax ./cmd/schway -run 'Phase32' -count=1` (after Wave 0 files exist; use `-count=1` always) |
| Full suite command | hosted only: `go test ./...` and `go test -race -timeout=20m ./...`, on ubuntu-latest and macos-latest |

### Phase Requirements to Test Map

Test names are proposed; the planner fixes the final names. D-04 fixes the three guide names. Commands below are for the planner and are not required to resolve until the tests exist.

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| D-10 (todo) | A `var` body ending in a bare identifier, number or call-paren round-trips: reparses, is a fixed point, same tokens and same checker codes; the 5 committed sources format correctly **as copies**; unchanged output for all other corpus files | unit | `go test ./internal/compiler/syntax -run 'TestPhase32FormatVarBody' -count=1` | Wave 0 (new) |
| D-10 | Lossless and generated lanes still pass | unit | `go test ./internal/compiler/syntax -count=1` | exists |
| D-10 / D-06 | Phase 30 golden holds with the dated format-change allow-list; stale entry control | integration | `go test ./internal/compiler/session -run 'TestPhase30PreChangeGolden' -count=1` | exists, edited |
| DX-19 / SC1 | The README's marked blocks run from a temp copy; every row matches exit, stdout and stderr; error rows hide the path | e2e | `go test ./cmd/schway -run 'TestPhase32GuideCleanCheckout' -count=1` | Wave 0 (new) |
| DX-19 / D-04 | Mutated expected value and renamed command each make the checker fail; marker-count, all-three-programs, row-count anti-vacuity | unit | `go test ./cmd/schway -run 'TestPhase32GuideParserNegativeControls' -count=1` | Wave 0 (new) |
| DX-19 / D-01 | Guide limit, case names, stdout and stderr agree with `checksum.expected.json` | unit | `go test ./cmd/schway -run 'TestPhase32GuideMatchesExpectedJSON' -count=1` | Wave 0 (new) |
| DX-19 / D-04 | Hosted + missing clang fails, local + missing clang skips | unit | `go test ./cmd/schway -run 'TestPhase32NativePolicy' -count=1` | Wave 0 (new) |
| DX-19 / SC2 | Old path is `ErrNotExist`; replacement exists; no renamed copy by tombstone digest; armed control; live-doc hygiene | unit | `go test ./internal/compiler/session -run 'TestPhase32ChecksumFrontierRetired' -count=1` | Wave 0 (new) |
| DX-19 / D-06 | Retired path may be absent in both goldens and fails if observed again | integration | `go test ./internal/compiler/session -run 'TestPhase28EntryExitClassManifest|TestPhase30PreChangeGolden|TestPhase30PreChangeGoldenIsNotInert' -count=1` | exists, edited |
| DX-19 / D-07 | LANGUAGE-MATURITY counts re-derived (162 and 5,352) | integration | `go test ./internal/compiler/session -run 'TestLanguageMaturityCountsAreCurrent' -count=1` | exists |
| EVD-15 / D-08 | sh and Go codedigest agree; moves on code edit; does not move on `.md` or `.planning` edit | integration | `go test ./internal/compiler/session -run 'TestPhase32CodeDigest' -count=1` | Wave 0 (new) |
| EVD-15 / D-09 | Gate accepts a good pair; rejects each named mutation (digest, tree, host, program, case, control, unknown row) | integration | `go test ./internal/compiler/session -run 'TestPhase32ReceiptGate' -count=1` | Wave 0 (new) |
| EVD-15 / D-10a | Receipt absent plus no living claim passes; claim without receipt fails; bound claim with a stale digest fails; "reopened" passes | unit | `go test ./internal/compiler/session -run 'TestPhase32ReceiptLivingClaims' -count=1` | Wave 0 (new) |
| EVD-15 / D-10b | Every `run()` route, flag and `examples` entry has a covered or "not covered because" row; new route fails; receipt program set equals audit set | unit | `go test ./cmd/schway -run 'TestPhase32PublicRouteAudit' -count=1` | Wave 0 (new) |
| EVD-15 / SC3 | Hosted macOS and Linux logs bound to the final digest | hosted | after the human step: `sh scripts/check-phase32-receipts.sh <macos.log> <ubuntu.log>` and the receipt test | at ingest |
| Process rule | Every Phase 32 SUMMARY has a resolvable `coverage:` block | meta | `go test ./internal/compiler/session -run 'TestSummaryCoverageGroundedness' -count=1` | exists |

### Sampling Rate

- **Per task commit:** the focused package command for each touched package, with `-count=1`.
- **Per wave merge:** `go test ./cmd/schway ./internal/compiler/syntax -count=1`, plus the session meta set quoted in Finding 6, whenever a `.schway` file, a register or a counted planning doc changes.
- **Phase gate:** hosted `go test ./...` and race runs green on both OSes with the run ID recorded, then the receipt ingest.

### Wave 0 Gaps

- [ ] `internal/compiler/syntax/syntax_phase32_format_test.go` for D-10.
- [ ] `cmd/schway/main_phase32_guide_test.go` for DX-19 and D-04.
- [ ] `cmd/schway/main_phase32_routes_test.go` for D-10b.
- [ ] `internal/compiler/session/session_phase32_retire_test.go` for SC2.
- [ ] `internal/compiler/session/session_phase32_receipt_test.go` for EVD-15 and D-10a.
- [ ] `scripts/codedigest.sh`, `scripts/verify-phase32.sh`, `scripts/check-phase32-receipts.sh`.
- No framework install. The test infrastructure exists.

## Security Domain

`security_enforcement` is enabled in config (ASVS level 1). Most categories do not apply (no auth, sessions, accounts or crypto beyond hashing).

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | none |
| V3 Session Management | no | none |
| V4 Access Control | no | the push and dispatch step is gated by the user's own `gh` auth, with explicit approval |
| V5 Input Validation | yes | Guide-driven shell lines: substitute only the `PATH` token, single-quoted; pass copy dir and line as `sh -c` positional arguments; validate output targets (`.schway/` only). Gate scripts parse logs with fixed-prefix greps and reject unknown rows. |
| V6 Cryptography | partly | SHA-256 via `crypto/sha256` or `shasum`/`sha256sum` for integrity only; no custom crypto. The gate is consistency, not authenticity (stated in the script header; keep stating it). |
| V12 Files | yes | The doc-test writes only under `t.TempDir()`; error rows assert the temp path segment never appears in stderr. |

### Known Threat Patterns

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| README edited to run arbitrary shell in CI | Tampering | The README is a reviewed repo file; the test runs only marked fences in a temp copy with bounded timeouts. |
| Forged receipt log | Spoofing | Out of scope by decision (deferred authenticated receipts); consistency checks plus the hosted run ID and artifact. |
| Stale receipt presented as current | Repudiation / Tampering | `digest_mismatch` plus the living-claim reopen rule. |
| Path echo in app errors | Information disclosure | Existing Phase 31 behavior; the doc-test pins "fixed stderr, no path". |

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `GITHUB_ACTIONS=true` is set on GitHub-hosted runners | Finding 3 (native policy) | The hosted "missing clang fails" rule would never fire; mitigated by also honoring `SCHWAY_REQUIRE_NATIVE=1` set in `ci.yml` |
| A2 | `shell: bash` in an Actions step runs with `-eo pipefail`, so a failing `sh … \| tee` fails the step | Finding 4 (CI wiring) | A failed verify run could look green; the existing phase27 step relies on the same behavior |
| A3 | A cold hosted `go build` of the copy stays well inside the 2-minute `RunCLITimeout` | Finding 3 | A flaky timeout; mitigate by adding a `BuildCLIErr`-style helper with the 5-minute limit |
| A4 | The prototype `format_sha256` values and the final fix's values are equal | Finding 1 | None if recomputed; the planner must compute them after the real change |
| A5 | Dated-heading structure puts `LANGUAGE-MATURITY.md:180` (and perhaps `:412`) outside "live" rows | Finding 2 hygiene scan | The hygiene scan fails on a historical line or misses a live one; confirm headings before coding |

Everything else above was read from source or run this session.

## Open Questions

1. **Add `native` and `testdata` to the codedigest paths? (deviation from D-08; needs a decision before the freeze)**
   - What we know: the D-08 set is `cmd internal examples scripts go.mod .github/workflows/ci.yml`. `native/*.c` is committed C runtime compiled by the native path and resolved from the repo root (`foreign_*.go`), and `testdata/` holds the frozen goldens and `.schway` fixtures that gates read. The digest's path set is the receipt contract and costly to change later (D-08 reversibility).
   - What's unclear: whether the user meant to exclude them.
   - Recommendation: add both (805 entries instead of 474). If the user declines, record "`native/` and `testdata/` are outside the digest" in the receipt and the amendment, so the omission is a stated non-claim.
2. **Where does the corrected usage text go?** `usageResult()` omits `stats` and `build --manifest`. Fixing it is a CLI text change. Recommendation: log it as an audit row or a pending todo, and leave the text alone in this phase.
3. **Do the 5 formatter-affected sources get a combined test?** The todo suggested "format both files canonically"; D-10 drops that. Recommendation: copies only, as above.
4. **Should phase 32's controls inside `verify-phase32.sh` include the doc-test?** Recommended yes (`TestPhase32GuideCleanCheckout`, `TestPhase32PublicRouteAudit`), with the `--- SKIP:` check so a clang-less host cannot pass.
5. **Single-commit freeze.** Does the plan want the final code commit to be the CI-wiring commit, or a separate "freeze" marker commit? Either works because the digest ignores `.planning`. Recommend naming the target commit explicitly in the checkpoint plan.

## Suggested Plan Decomposition (for the planner; follows D-00)

1. **32-01 formatter fix:** the `format.go` guard, the `syntax_phase32_format_test.go` regression, the `phase30FormatChanges` allow-list with recomputed hashes, a lossless-lane run, and a todo status update. Touches `session_phase30_prechange_test.go`, so it must finish before 32-02 edits that file.
2. **32-02 frontier retirement:** delete the file; replace `TestPhase20ChecksumFrontier`; extract `phase28Compare` plus `phase28RetiredEntries` plus control; `phase30RetiredPaths` plus control; PROJECT.md amendment; live-row pointers (`LANGUAGE-MATURITY.md:297,412`, `PRODUCT-ROADMAP.md:272,298`, README:140); LANGUAGE-MATURITY counts to 162 and 5,352.
3. **32-03 guide and doc-test** (parallel with 32-02 except for the README:140 line, which belongs to one plan only): README restructure with `.schway/` outputs, marked fences, keyed tables, "Earlier programs" pointer, known-limits line, then the three D-04 tests and `TestPhase32NativePolicy`.
4. **32-04 route audit, digest, scripts, CI, receipt tests:** all code that feeds the digest, ending in the single freeze commit.
5. **32-05 human checkpoint:** run the meta set locally, show the exact line `! git push -q origin phase26-validation && gh workflow run ci.yml --ref phase26-validation`, then wait for approval.
6. **32-06 ingest (`.planning` only):** `gh run view <id> --json headSha,conclusion,jobs`, `gh run download`, verify digest equality before writing anything, commit logs, `32-RECEIPT.json`, living-doc markers, REQUIREMENTS and ROADMAP status.

Each SUMMARY needs a `coverage:` block naming the tests above. The 32-05 SUMMARY is the lone `human_judgment: true` entry.

## Sources

### Primary (HIGH confidence, opened or run this session)
- `internal/compiler/syntax/format.go` (full read; bug reproduced and fix prototyped in a scratch copy outside the repo)
- `internal/compiler/session/session_phase20_test.go`, `session_phase28_manifest_test.go`, `session_phase30_prechange_test.go`, `session_phase273_docs_test.go`, `session_phase273_identity_test.go`, `session_phase27_host_test.go`, `session_phase27_docs_test.go`, `summary_coverage_test.go`, `verification_groundedness_test.go` (relevant ranges), `self_describing_docs_test.go`
- `internal/compiler/native/phase25_utility_test.go:170-306`, `native_test.go` (spawn lint)
- `internal/compiler/testsupport/testsupport.go`, `checksum_oracle.go`, `checksum_matrix_test.go`
- `cmd/schway/main.go` (dispatch), `main_phase30_test.go`, `main_phase31_test.go`, `main_phase31_adler_low_test.go`
- `scripts/verify-phase27.sh`, `scripts/check-phase27-receipts.sh`, `.github/workflows/ci.yml`, `.gitignore`
- `examples/checksum/README.md`, `checksum.expected.json`, `examples/phase27/README.md`
- `gh run view 38299800002 --json …` and `gh run download 38299800002 -R szTheory/schway` (hosted logs for both OSes)
- A full `go test ./...` run on a scratch copy with the file deleted and the formatter prototype applied

### Secondary (MEDIUM confidence)
- `.planning/phases/31-*/31-VALIDATION.md` and `31-VERIFICATION.md` for the session meta-set command and the planning-doc conventions

### Tertiary (LOW confidence)
- GitHub Actions default environment variables and `shell: bash` flags: training knowledge only, tagged `[ASSUMED]` (A1, A2)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH. The stack is the repo's own existing stack, with nothing new.
- Architecture: MEDIUM. The patterns are precedent-backed, but the script and log grammar is a recommendation and the digest path set needs a user decision.
- Pitfalls: HIGH. Each was reproduced or read in source, except cold-build timing (A3).

**Research date:** 2026-10-12
**Valid until:** 2026-10-26, or until any code change lands (the line-number references move).
