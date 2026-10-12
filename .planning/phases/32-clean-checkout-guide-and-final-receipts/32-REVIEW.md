---
phase: 32-clean-checkout-guide-and-final-receipts
reviewed: 2026-10-12T00:00:00Z
depth: standard
files_reviewed: 20
files_reviewed_list:
  - .github/workflows/ci.yml
  - cmd/schway/main_phase32_guide_test.go
  - cmd/schway/main_phase32_routes_test.go
  - examples/checksum/README.md
  - internal/compiler/session/evidence_grade_test.go
  - internal/compiler/session/session_phase28_manifest_test.go
  - internal/compiler/session/session_phase30_prechange_test.go
  - internal/compiler/session/session_phase32_ci_test.go
  - internal/compiler/session/session_phase32_freeze_test.go
  - internal/compiler/session/session_phase32_gate_test.go
  - internal/compiler/session/session_phase32_receipt_test.go
  - internal/compiler/session/session_phase32_retire_test.go
  - internal/compiler/session/verification_groundedness_test.go
  - internal/compiler/syntax/format.go
  - internal/compiler/syntax/syntax_phase32_format_test.go
  - scripts/check-phase32-freeze.sh
  - scripts/check-phase32-receipts.sh
  - scripts/codedigest.sh
  - scripts/verify-phase32.sh
findings:
  critical: 0
  warning: 2
  info: 5
  total: 7
status: issues_found
---

# Phase 32: Code Review Report

**Reviewed:** 2026-10-12
**Depth:** standard
**Files Reviewed:** 19 (the guide, routes, ci, retire, manifest and prechange tests, and the README, were read for the focus areas below; the deep read went to the gate scripts, formatter, receipt/gate/freeze tests and the corpusRetiredTests allow-list)
**Status:** issues_found

## Summary

No Critical findings. None of the findings below affects the soundness of the already-ingested hosted receipt (run 38315799626, codedigest c1b4e85d...): all are fail-closed inconsistencies, hardening gaps, or coverage gaps. Any code fix touches digest paths (`scripts/`, `internal/`) and would reopen the receipts under D-10a, so the recommendation for all of them is to record and defer (post-M006 follow-up) rather than fix now.

Focus-area conclusions:

- sh gate fail-open analysis: `set -eu` is on in all four scripts. `grep -c ... || true` is used correctly (grep -c prints "0" and `|| true` adds no output, so no "0\n0" trap). `reject` inside `check_log` runs in a `$(...)` subshell but the assignment status propagates under `set -e`, and the per-case loop uses a heredoc (not a pipe), so its `reject` exits the main shell. Hosted CI runs `shell: bash` (implicit `-eo pipefail`), so `verify-phase32.sh | tee` cannot mask a producer failure. Log contents are only ever interpolated via `printf '%s'` or `grep -F`; no path or command is built from log contents, so no traversal or injection from untrusted logs.
- Negative controls: the gate tests assert both a non-zero exit and the specific `reason=` string, and `mutate` refuses a vacuous anchor. The verify script treats `--- SKIP:` and `no tests to run` as failures, and requires `--- PASS: <name> ` per control. I found no vacuous-pass control.
- Formatter guard: correct for the cases it targets. It cannot double-fire with the preceding `ident =` guard (the first calls `newline()`, which clears `lineOpen`). Function parameter lists are not "inside function body" (contexts are pushed only at `{`), so multi-line params are unaffected. The `match` exclusion is by top-of-stack context only. See IN-03 for the coverage gap.
- corpusRetiredTests: cannot mask genuine drift. `receiptTimeCorpusPairs` requires (1) the name to be absent from the live index, (2) the live set to NOT already match the receipt, (3) each entry to alter the reconstructed set, and then the reconstructed digest must equal the frozen manifest digest inside `checkedInCorpusRunRecord`. An allow-list entry can only explain drift it reproduces byte-for-byte. Minor looseness noted in IN-04.

## Warnings

### WR-01: Freeze gate pathspec includes `.md`, contradicting the digest definition it claims to follow

**File:** `scripts/check-phase32-freeze.sh:12,41,44`
**Issue:** The header says "commits after the freeze that touch only .planning or prose (outside the digest paths) are accepted", and the revision is defined as the codedigest, which strips `*.md` (`scripts/codedigest.sh:25`). But `phase32_freeze_paths` is the raw path list, and both the `git log` and `git diff --quiet` checks are over it without excluding `.md`. A post-freeze edit to, for example, `examples/checksum/README.md` (a source file in this phase's own scope) or any `internal/**/*.md` leaves the codedigest unchanged yet makes the freeze gate fail with `digest_path_changed_after_freeze`. The two bindings (receipt gate: codedigest; freeze gate: path diff) therefore disagree on what "code changed" means. This fails closed, so it cannot admit a bad candidate; the cost is a spurious rejection of a legitimate prose-only follow-up, and a contract that is not what the comment states.
**Receipt impact:** None. The recorded receipt was accepted under the stricter rule.
**Fix:** Either exclude markdown from the gate (`:(exclude,glob)**/*.md` pathspecs on both git calls) so it matches codedigest, or amend the header comment and D-08 to say the freeze gate is strictly stricter than the digest. Defer past the freeze.

### WR-02: `codedigest.sh` is cwd-dependent and cannot detect an empty listing

**File:** `scripts/codedigest.sh:23-25`
**Issue:** `git ls-tree -r HEAD -- cmd internal ...` interprets pathspecs relative to the current directory. Run from any subdirectory of the repo it exits 0 with an empty listing, and the script then prints sha256 of a single newline: a plausible, well-formed, wrong digest. There is no cd to the repo root and no empty-listing check. Additionally the final pipeline (`printf | grep -v | hasher | cut`) has no pipefail, so a failure of `grep -v` or the hasher is masked by `cut`'s status (e.g. a hasher that exits non-zero still yields output from `cut`). The production callers are safe (`verify-phase32.sh` and `check-phase32-receipts.sh` both `cd` to the root first, and the Go twin uses `git -C root`), and `TestPhase32CodeDigest` runs it from the root only, so nothing exercises the failure.
**Receipt impact:** None. Both producer and gate ran from the root; the recorded digest is the real one. A fixed script would change the digest paths (scripts/) and reopen the receipt.
**Fix:**
```sh
cd "$(git rev-parse --show-toplevel)" || fail "no_repo_root"
[ -n "$listing" ] || fail "empty_listing"
```
Defer past M006; add a subdirectory-invocation row to `TestPhase32CodeDigest` at the same time.

## Info

### IN-01: Env override of the expected digest in the production gate

**File:** `scripts/check-phase32-receipts.sh:31`
**Issue:** `PHASE32_EXPECTED_CODEDIGEST` replaces the recomputed digest. The header documents it, and the tests rely on it, but a stale value exported in a shell or CI env silently turns "bound to this checkout" into "bound to whatever the env says". It is not reachable from the hosted workflow as written (the gate is not invoked there).
**Fix:** Print the digest source (`source=env|computed`) in the pass line so an override is visible in the recorded output.

### IN-02: Living-claim marker only matches at column 0

**File:** `internal/compiler/session/session_phase32_receipt_test.go:221-223`
**Issue:** `strings.HasPrefix(line, "M006 final hosted receipts:")` means an indented, bulleted, or blockquoted claim (`- M006 final hosted receipts: bound to ...`) is never parsed, so a stale or malformed claim in that form escapes `stale_bound_claim` / `malformed_marker`. Related: duplicate `Hosts` entries with the same name are collapsed by `byName` (last wins) without a problem being raised, and result rows are not checked for `codedigest`/`host` or the `controls` count in the Go checker (the sh gate does check these; the Go checker only covers log SHA, identity digest/tree, result status, program coverage).
**Fix:** Use `strings.Contains` or trim leading `-*> ` before the prefix test; add a `duplicate_host_%s` problem.

### IN-03: Formatter guard has no negative-context tests

**File:** `internal/compiler/syntax/format.go:40-49`, `internal/compiler/syntax/syntax_phase32_format_test.go:36-41`
**Issue:** The guard excludes `match` by top-of-stack context only and fires for any other context under a function or arm. The tests cover three top-level-function terminal shapes plus five committed sources, but there is no row for (a) a bare result after a nested `if`/`while` block, (b) an identifier-valued match arm followed by another arm, (c) a bare result inside an `arm { ... }` body nested in a `match`. I traced (b) and (c) by reading and expect correct output, but nothing pins them, so a future change to context handling would not fail.
**Fix:** Add three round-trip rows through `assertRoundTrip`.

### IN-04: Retirement citation check is satisfied by any `| superseded |` line mentioning the name

**File:** `internal/compiler/session/evidence_grade_test.go:~397-415` (`retiredCitationProblem`)
**Issue:** The check passes when any single line of `EVIDENCE-RECONCILIATION.md` contains both the retired name and `| superseded |`, regardless of which archived row it reconciles. A row citing the retired test could be covered by an unrelated line. It does not weaken the digest binding (see Summary), only the "every citing row is reconciled" claim.
**Fix:** Key the reconciliation lookup by validation doc and task id, not by test name alone.

### IN-05: Process-wide index cache

**File:** `internal/compiler/session/verification_groundedness_test.go:70-100`
**Issue:** `buildTestIndex` now parses once per process via `sync.Once` and returns clones. A test that creates or deletes `_test.go` files under the real project root during a run would see a stale index. None does today (tagged and temp-root variants bypass the cache), and `clone()` copies every field of `testIndex`, so this is a latent trap rather than a bug.
**Fix:** Note in the doc comment that the project root must be read-only for the duration of the test process.

---

_Reviewed: 2026-10-12_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
