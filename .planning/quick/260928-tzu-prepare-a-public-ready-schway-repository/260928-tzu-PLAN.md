---
quick_task: 260928-tzu
status: executing
depends_on: []
files_modified:
  - go.mod
  - .gitignore
  - .github/workflows/ci.yml
  - README.md
  - AGENTS.md
  - cmd/lang/
  - cmd/lang-repair/
  - native/lang_*
  - internal/compiler/
  - scripts/
  - examples/
  - testdata/
  - .claude/skills/
  - .planning/PROJECT.md
  - .planning/STATE.md
  - .planning/phases/22-native-application-build-and-single-execution/
  - .planning/phases/23-live-local-allocation-and-discharge/
  - .planning/quick/260928-tzu-prepare-a-public-ready-schway-repository/
must_haves:
  truths:
    - The current source tree presents Schway consistently through its module path, CLI names, `.schway` files, protocol namespaces, C ABI symbols, CI, and active developer documentation.
    - Every commit reachable from the selected public branch and milestone tags remains represented in the rewritten history with the same order, parents, and merge topology.
    - The publishable tree and history contain no detected personal home paths, personal author names, personal email addresses, credentials, or secrets.
    - The public `main` branch points to the latest Phase 23 tip and preserves milestone tags.
    - The first push starts the existing hosted Ubuntu evidence aggregate.
  artifacts:
    - `.planning/quick/260928-tzu-prepare-a-public-ready-schway-repository/260928-tzu-VERIFICATION.md` records redacted privacy and topology checks.
    - The private audit directory beside this plan contains `commit-map.tsv`, `topology-report.json`, and `scan-summary.json`; none is copied into the publishable clone.
    - `github.com/szTheory/schway` has the sanitized `main` branch and preserved milestone tags.
    - A successful hosted Ubuntu `evidence-aggregate` receipt is available for the public revision.
  key_links:
    - `go.mod` module path and Go imports resolve under `github.com/szTheory/schway`.
    - The renamed `cmd/schway` executable and `cmd/schway-repair` command are the names used by scripts, fixtures, examples, tests, and documentation.
    - `.schway` source files and extension-sensitive tools agree on the canonical source suffix.
    - `schway.*` protocol identifiers and `schway_`/`SCHWAY_` native ABI names agree across emitters, adapters, schemas, examples, and fixtures.
    - `.github/workflows/ci.yml` continues to run `scripts/verify-phase23.sh` on `ubuntu-latest`.
---

# Prepare the public Schway repository without personal data

<!-- schway-current:start -->
Current publication identity (2026-09-28): Schway uses the Go module `github.com/szTheory/schway`, the `schway` and `schway-repair` commands, `.schway` source files, `schway.*` and `schway:*` protocol identifiers, and `schway_` and `SCHWAY_` native ABI symbols. Preserve older spellings only where they document historical implementation evidence.
<!-- schway-current:end -->

## Goal

Create a publication-ready history for `github.com/szTheory/schway` that uses
the settled technical identifiers, removes personal data, and retains the
publishable commit sequence and merge topology. Publish only after a final
redacted scan is clean.

## Tasks

### 1. Coordinate the current-tree rename in an isolated writable clone

<files>
`go.mod`, `.gitignore`, `.github/workflows/ci.yml`, `README.md`, `AGENTS.md`,
`cmd/lang/`, `cmd/lang-repair/`, `native/lang_*`, `internal/compiler/`,
`scripts/`, `examples/`, `testdata/`, `.claude/skills/`, `.planning/PROJECT.md`,
`.planning/STATE.md`, Phase 22 `22-VERIFICATION.md` and `22-VALIDATION.md`,
Phase 23 `23-VERIFICATION.md`, `23-VALIDATION.md`, and `23-SECURITY.md`, and
this quick task's `260928-tzu-CONTEXT.md` and `260928-tzu-PLAN.md`.
</files>

<action>
Create a separate writable clone and carry forward only tracked project files
plus the already-authorized planning changes; omit local build binaries,
`.planning/milestone.lock`, and `.planning/research/.cache/`. In that clone,
change the Go module/import path to `github.com/szTheory/schway`, rename
`cmd/lang` and `cmd/lang-repair` to `cmd/schway` and `cmd/schway-repair`, and
update corresponding executable references. Rename tracked `.lang` source
files to `.schway` and update extension-aware code, fixtures, scripts, examples,
and active docs. Rename project-owned `lang.`/`lang:` protocol namespaces and
`lang_`/`LANG_` C ABI symbols to the `schway`/`SCHWAY` forms, including the
project-owned native filenames. Update the CI workflow description and temp
directory prefix. Update active identity docs to record the completed rename
and privacy gate, while preserving archived planning records as historical
evidence. Do not rewrite old project identifiers across historical commit
snapshots; the new migration commit records the current-tree transition.
Stage the complete migration before the tracked-path gate below. Its active
text scope is `go.mod`, `.gitignore`, `.github/workflows/ci.yml`, `README.md`,
`AGENTS.md`, `cmd/`, `native/`, `internal/compiler/`, `scripts/`, `examples/`,
`testdata/`, and the current `.claude/skills/` guidance. Exclude
`testdata/phase1/` through `testdata/phase5/`,
`testdata/phase16/historical/`, the immutable
`testdata/phase16/validation-corpus-run-record.jsonl` receipt, and
`.claude/skills/*/sources/` from the text gate because the Phase 1-5 source
bytes are frozen historical fixtures, the later paths are archived generated
material or a historical run record, and the skill paths are independent spike
sources; keep every path in the privacy scan.
In each listed active planning document, mark its current publication-facing
identity or command guidance with one `<!-- schway-current:start -->` and
`<!-- schway-current:end -->` pair and update that region to the settled
identifiers. Preserve the locked decisions in CONTEXT.md and the plan's
legacy-name search expressions; preserve Phase 22/23 historical receipts and
archived planning records outside those current guidance regions.
</action>

<verify>
  <automated>
From the isolated clone root, run this read-only gate after staging. It prints
only family counts, never matching lines or values. Every `old_*` count must
be zero; every required positive check and the exact source count must pass.

    set -eu
    scope=(go.mod .gitignore .github/workflows/ci.yml README.md AGENTS.md cmd native internal/compiler scripts examples testdata .claude/skills)
    exclusions=(':(exclude)testdata/phase1/**' ':(exclude)testdata/phase2/**' ':(exclude)testdata/phase3/**' ':(exclude)testdata/phase4/**' ':(exclude)testdata/phase5/**' ':(exclude)testdata/phase16/historical/**' ':(exclude)testdata/phase16/validation-corpus-run-record.jsonl' ':(exclude)internal/compiler/native/foreign_legacy.go' ':(exclude)internal/compiler/check/frozen_foreign_expectation_test.go' ':(exclude).claude/skills/*/sources/**' ':(exclude)testdata/phase07/call_argument_used_once.schway' ':(exclude)testdata/phase07/call_argument_used_twice.schway' ':(exclude)testdata/phase07/call_fallible_foreign_reach.schway' ':(exclude)testdata/phase07/call_two_fallible_callees_disagree.schway' ':(exclude)testdata/phase07/call_type_mismatch.schway' ':(exclude)testdata/phase07/call_uncallable_callee.schway' ':(exclude)testdata/phase07/clean_but_unpublishable.schway' ':(exclude)testdata/phase07/cycle_indirect.schway' ':(exclude)testdata/phase07/cycle_mutual.schway' ':(exclude)testdata/phase07/cycle_self.schway' ':(exclude)testdata/phase07/cycle_through_match_arm.schway' ':(exclude)testdata/phase07/duplicate_function_name.schway' ':(exclude)testdata/phase07/relay_escort_witness.schway' ':(exclude)testdata/phase08/match_arm_call.schway' ':(exclude)testdata/phase08/negative_control_fails.schway' ':(exclude)testdata/phase08/negative_control_infallible.schway' ':(exclude)testdata/phase08/relay_depth2_accept.schway' ':(exclude)testdata/phase08/relay_depth2_refuse.schway' ':(exclude)testdata/phase08/twin_a_accept.schway' ':(exclude)testdata/phase08/twin_a_refuse.schway' ':(exclude)testdata/phase08/twin_b_accept.schway' ':(exclude)testdata/phase08/twin_b_refuse.schway' ':(exclude)testdata/phase11/multi_function_gate_corpus.schway' ':(exclude)testdata/phase13/heldout_call_argument_ambiguous.schway' ':(exclude)testdata/phase13/heldout_call_argument_mismatch.schway' ':(exclude)testdata/phase13/heldout_fallible_call_unconsumed.schway' ':(exclude)testdata/phase13/heldout_shared_callee_twin_alpha.schway' ':(exclude)testdata/phase13/heldout_shared_callee_twin_mirror.schway')
    count_text() { git grep --cached -I -l -E "$1" -- "${scope[@]}" "${exclusions[@]}" | wc -l | tr -d '[:space:]'; }
    old_module=$(count_text 'github[.]com/codename-lang/lang')
    old_cli=$(count_text 'cmd/lang(-repair)?(/|[^[:alnum:]_-])|/lang(-repair)?([^[:alnum:]_-]|$)|(^|[^[:alnum:]_])lang-repair([^[:alnum:]_-]|$)|(^|[^[:alnum:]_])lang[[:space:]]+(app|build|check|run|verify)([^[:alnum:]_]|$)')
    old_extension=$(count_text '[.]lang([^[:alnum:]_]|$)')
    # Already-versioned lang.* schema names are frozen wire identifiers.
    # Reject old lang: annotation labels; current namespaces stay schway.*.
    # The Phase 13 injector and cleanup adapter retain markers for byte-frozen inputs.
    protocol_exclusions=("${exclusions[@]}" ':(exclude)internal/compiler/session/session_phase13_injectors.go' ':(exclude)internal/compiler/session/session.go')
    old_protocol=$(git grep --cached -I -l -E '(^|[^[:alnum:]_])lang:' -- "${scope[@]}" "${protocol_exclusions[@]}" | wc -l | tr -d '[:space:]')
    old_abi=$(git grep --cached -I -l -E '(^|[^[:alnum:]_])(lang_|LANG_)' -- "${scope[@]}" "${exclusions[@]}" ':(exclude)internal/compiler/check/check_blame_test.go' ':(exclude)internal/compiler/syntax/syntax_test.go' | wc -l | tr -d '[:space:]')
    old_paths=$(git ls-files | rg '(^cmd/lang(-repair)?/|^native/lang_|[.]lang$)' | wc -l | tr -d '[:space:]')
    printf 'old_module=%s old_cli=%s old_extension=%s old_protocol=%s old_abi=%s old_paths=%s\n' "$old_module" "$old_cli" "$old_extension" "$old_protocol" "$old_abi" "$old_paths"
    if [ "$old_module" -ne 0 ] || [ "$old_cli" -ne 0 ] || [ "$old_extension" -ne 0 ] || [ "$old_protocol" -ne 0 ] || [ "$old_abi" -ne 0 ] || [ "$old_paths" -ne 0 ]; then
      printf 'rename_gate=failed\n'
      exit 1
    fi
    test "$(git ls-files '*.schway' | wc -l | tr -d '[:space:]')" -eq 145
    test "$(git ls-files '*.lang' | wc -l | tr -d '[:space:]')" -eq 0
    test "$(git ls-files 'cmd/schway/main.go' 'cmd/schway-repair/main.go' | wc -l | tr -d '[:space:]')" -eq 2
    test "$(sed -n '1p' go.mod)" = 'module github.com/szTheory/schway'
    git grep --cached -I -q -E '(^|[^[:alnum:]_])schway[.:]' -- cmd internal/compiler examples testdata native
    git grep --cached -I -q -E '(^|[^[:alnum:]_])schway_' -- native internal/compiler examples
    git grep --cached -I -q -E '(^|[^[:alnum:]_])SCHWAY_' -- native internal/compiler examples
    git grep --cached -I -q 'cmd/schway' -- .github/workflows/ci.yml scripts README.md
    git grep --cached -I -q '[.]schway' -- cmd internal/compiler scripts examples testdata README.md
    planning_docs=(.planning/PROJECT.md .planning/STATE.md .planning/phases/22-native-application-build-and-single-execution/22-VERIFICATION.md .planning/phases/22-native-application-build-and-single-execution/22-VALIDATION.md .planning/phases/23-live-local-allocation-and-discharge/23-VERIFICATION.md .planning/phases/23-live-local-allocation-and-discharge/23-VALIDATION.md .planning/phases/23-live-local-allocation-and-discharge/23-SECURITY.md .planning/quick/260928-tzu-prepare-a-public-ready-schway-repository/260928-tzu-CONTEXT.md .planning/quick/260928-tzu-prepare-a-public-ready-schway-repository/260928-tzu-PLAN.md)
    doc_old=0
    doc_count=0
    for doc in "${planning_docs[@]}"; do
      current=$(git show ":$doc" | sed -n '/^<!-- schway-current:start -->$/,/^<!-- schway-current:end -->$/p')
      test "$(printf '%s\n' "$current" | rg -c '^<!-- schway-current:start -->$')" -eq 1
      test "$(printf '%s\n' "$current" | rg -c '^<!-- schway-current:end -->$')" -eq 1
      printf '%s\n' "$current" | rg -qi 'schway'
      hits=$(printf '%s\n' "$current" | rg -c 'github[.]com/codename-lang/lang|cmd/lang(-repair)?|/lang(-repair)?([^[:alnum:]_-]|$)|(^|[^[:alnum:]_])lang-repair([^[:alnum:]_-]|$)|(^|[^[:alnum:]_])lang[[:space:]]+(app|build|check|run|verify)([^[:alnum:]_]|$)|[.]lang([^[:alnum:]_]|$)|(^|[^[:alnum:]_])lang[.:]|(^|[^[:alnum:]_])(lang_|LANG_)' || true)
      doc_old=$((doc_old + hits))
      doc_count=$((doc_count + 1))
    done
    printf 'active_planning_docs=%s planning_old_lines=%s\n' "$doc_count" "$doc_old"
    test "$doc_old" -eq 0

Do not run the project test suite locally. The existing GitHub workflow owns
the hosted Ubuntu evidence receipt after publication.
  </automated>
</verify>

<done>The latest source tree and active developer materials use the chosen Schway identifiers, and every tracked source fixture has the `.schway` extension.</done>

### 2. Scrub history while preserving its graph

<files>
All blobs, commit messages, commit/tag identities, branches, and tags in the
isolated publishable clone; private original-checkout
`.planning/quick/260928-tzu-prepare-a-public-ready-schway-repository/audit/commit-map.tsv`,
`audit/topology-report.json`, and `audit/scan-summary.json`.
</files>

<action>
Make public `main` point to the current Phase 23 tip, which already descends
from local `main`. Preserve all commits reachable from that tip and the three
milestone tags, while leaving private stash and Codex checkpoint refs out of
the public ref set. Rewrite every publishable commit and tagger display name to
`szTheory`; retain the existing account-matching GitHub noreply metadata.
Replace macOS and Linux home prefixes and macOS per-user temporary-directory
prefixes with a generic marker in all blobs and commit/tag messages. Remove the one unclassified
co-author identity footer from commit messages. Configure history filtering so
empty commits and merge topology are never pruned. Keep the original checkout
untouched and retain a complete old-to-new commit map for verification.
Write private audit artifacts outside the publishable clone, beside this plan:
`audit/commit-map.tsv` with one `old_commit<TAB>new_commit` row per original
reachable commit; `audit/topology-report.json` with old/new baseline commit
counts, per-commit ordered old parents mapped to ordered new parents, root and
merge counts, all three milestone tag names and peeled old/new commit targets,
the original Phase 23 tip and its mapped successor, and separately listed
additive migration/documentation commits; and `audit/scan-summary.json` with
the exact public ref set, scanned commit/tag metadata and messages, all
reachable blob IDs and paths, current tracked tree, pattern-family candidate
counts, redaction counts, and each false-positive category with a
non-sensitive reason and count. Never store candidate values, matching lines,
or personal paths in these reports.
</action>

<verify>
  <automated>Read `audit/commit-map.tsv`, `audit/topology-report.json`, and `audit/scan-summary.json` in the original checkout. Fail unless the map has exactly one distinct target for every original commit reachable from the selected Phase 23 tip or a milestone tag, and no duplicate targets; old and rewritten baseline commit counts agree; every ordered new parent list equals the mapped old parent list; root and merge counts agree; every milestone tag peels to its mapped target; the mapped Phase 23 tip is the parent of the separately listed migration commit; and no stash/tool checkpoint ref is in the public ref set. Independently rescan every selected public ref's commit/tag identity and message, every reachable blob, and the final tracked tree for user-specific home and temporary-directory prefixes, personal names and contacts, phone-shaped strings, credential patterns, and generic secrets. Record only redacted counts and non-sensitive false-positive classifications. Fail on any confirmed or unclassified candidate.</automated>
</verify>

<done>The rewritten public history has the same commit sequence and parent topology, uses only the public handle for commit/tag display names, and has no detected personal data or credentials.</done>

### 3. Gate publication on the final audit and hosted CI receipt

<files>
`.planning/quick/260928-tzu-prepare-a-public-ready-schway-repository/260928-tzu-VERIFICATION.md`, `.planning/PROJECT.md`, `.planning/STATE.md`, and the GitHub repository refs.
</files>

<action>
Record only aggregate, value-redacted audit results and the verified history
map, including the three private audit artifact paths and their pass/fail
totals. The publication gate passes only if Task 1's six source old-name counts
and active-planning-doc old-name count are
zero, the required new-name checks pass, exactly 145 tracked `.schway` sources
exist, every topology equality in Task 2 passes, the public ref set is only
sanitized `main` plus the three milestone tags, and confirmed and
unclassified privacy/secret candidates both equal zero across the stated scan
scope. Any failed or missing check blocks publication. If any personal-data
candidate remains unclassified, stop before pushing.
Once the current tree and every public ref pass the audit, push sanitized
`main` and milestone tags to the already-created empty repository. Check the
repository's default branch and the resulting GitHub Actions run; retain the
existing Phase 23 `evidence-aggregate` result on `ubuntu-latest` as the hosted
receipt. Do not add GitHub secrets; the current workflow requires none.
</action>

<verify>
  <automated>Re-run the exact Task 1 read-only gate and Task 2 map/topology/privacy checks against the final publishable refs immediately before `git push`; require the stated publication gate to pass in full. After the push, use `git ls-remote` to check public `main` and all three milestone tag targets against the audited hashes, and inspect the GitHub Actions run for the `evidence-aggregate` job on `ubuntu-latest`. Mark the task failed if any ref differs, the job does not pass, or the final redacted scan finds a confirmed or unclassified candidate.</automated>
</verify>

<done>The public repository contains only the sanitized history, and the Phase 23 hosted Ubuntu evidence receipt is recorded without uploading PII.</done>

## Verification

- Preserve 1,538 original commits reachable from the current branch and local
  `main`, plus three milestone tags; the rewritten migration/documentation
  commits are additive.
- Preserve the current branch's merge graph, including its 18 merge commits.
- Confirm the original checkout remains unchanged apart from GSD task artifacts
  already created for planning.
- Do not print or write raw privacy candidates to logs or reports.

<threat_model>
| ID | STRIDE category | Threat | Severity | Mitigation |
|---|---|---|---|---|
| T1 | Information disclosure | A personal path, identity, contact address, or credential reaches the public GitHub repository through an old blob, message, tag, or current file. | high | Scrub the full publishable history and current tree; scan every public ref and blob with values redacted; stop before push if any candidate is unclassified. |
| T2 | Tampering | History filtering silently drops, reorders, or changes parent relationships or milestone tag targets. | high | Keep pruning disabled and compare every old commit's ordered parents through the old-to-new map; verify tag target mapping before pushing. |
| T3 | Elevation of privilege | A history-rewrite command or GitHub operation reaches the public remote before the privacy gate passes. | high | Work in an isolated copy; keep the existing repository empty until the final scan is clean; push only the approved `main` and milestone tags. |
</threat_model>

## Dated CI reconciliation amendment (2026-09-29)

The public repository is live and history preservation/privacy gates have
passed. CI run `36596749598` confirms the first additive repair fixed Phase 23
contract validation and Linux libc++ setup, but the overall checks and evidence
aggregates remain red. Apply the following compatibility boundary to the
remaining work:

1. Keep Schway as the current public identity: module path, commands, `.schway`
   source extension, active documentation, and current capability contracts.
2. Preserve already-versioned wire identifiers (`/0`, `/1`, and other issued
   versions) and their exact bytes. In particular, retain historical `lang.*`
   schema strings until a separately versioned schema migration is designed. A
   brand rename alone does not
   authorize changing a frozen wire contract.
3. Keep historical Phase 16/generated-C controls byte-frozen and linked to
   their matching legacy adapter. For current Schway ABI changes, use the
   existing change-ledger pattern: old/new hashes, moved responsibility,
   structural reason, executable semantic witness, fixture, and disposition.
4. Preserve Phase 1-5 fixture contents byte-for-byte under the selected
   `.schway` filenames; isolate legacy source-level foreign symbols in
   `internal/compiler/native/foreign_legacy.go` and
   `testdata/phase16/historical/foreign_phase1_5_legacy_adapter.c`; pin its
   original symbol in `internal/compiler/check/frozen_foreign_expectation_test.go`.
   These three paths are the only current shim/expectation exceptions to the
   old-ABI rename scan. Reconcile
   remaining source/diagnostic IDs and rewritten commit/line references with
   explicit provenance. Do not bulk-regenerate expected hashes or edit
   archived receipts to claim fresh verification.
5. Refresh current corpus counts and move current-identity guidance so it does
   not shift archived line-pinned findings. Run `git diff --check`, then use
   hosted Ubuntu/macOS CI as the verification path. Before each additive push,
   rerun the full reachable-ref privacy scan and require zero confirmed and
   zero unclassified candidates.

Complete the work in these batches: (A) versioned schema compatibility; (B)
frozen Phase 1-5 fixtures plus current versus historical C ABI and
source/diagnostic fingerprints; (C) rewritten
commit references, groundedness line pins, and corpus docs; (D) full hosted CI
green on both operating systems and both evidence aggregates. A batch is
accepted only when its hosted checks pass or its remaining red checks are
clearly isolated to the next batch.

## Dated hosted-CI outcome amendment (2026-09-29)

Commit `3b7919be` published the fixture/adapter batch after the full reachable
history scan passed. GitHub Actions run `36610374870` then completed red on
both operating systems. Both hosts passed `go vet ./...`, `go build ./...`,
and `scripts/verify-phase23.sh`; both failed `go test ./...` and
`scripts/verify-phase6.sh`, so race jobs were skipped. The logs contain 125
failed test/subtest names on Ubuntu and 113 on macOS. The failure inventory is
not a reason to refresh expected values in bulk.

The rename gate passes with 145 tracked `.schway` sources. Phase 1-5 `.schway`
fixture contents now byte-match their original pre-rename sources, and the
legacy resource, arena, nonlocal, and retained foreign calls route through the
test-only alias adapter. This fixes the Phase 1-3 core-byte subtests; Phase 4-5
core hashes and Phase 1-5 manifest IDs still fail. The Phase 4 checker test
also had a stale Schway symbol expectation for the frozen legacy fixture; pin
the fixture name in the named test helper, which is an explicit rename-gate
exception.

Continue in these bounded slices:

1. **Historical source identity** — inspect Phase 4-5 serialized core fields
   and Phase 1-5 evidence-manifest IDs. Preserve source bytes and issued
   schema IDs. Add only field-level migration records supported by the
   hosted diff and an independent semantic witness.
2. **Current artifact identity** — reconcile current C outputs, C golden
   ledgers, Phase 1 evidence, canonical program digests, and diagnostic IDs.
   Keep the old digest and new digest with a reason, changed responsibility,
   and executable witness; preserve frozen golden artifacts.
3. **Executable controls** — fix remaining runtime failures independently of
   identity pins, including the Phase 5 defect terminal-record control; verify
   the four subcommands, interpreter/native refusal boundary, and sanitizer
   positives through hosted CI.
4. **Rewritten history and docs** — map each stale commit reference through
   the private original-to-public commit map, verify its cited change, refresh
   current corpus counts, and move identity guidance to document tails so
   archived line pins retain their cited text.
5. **Acceptance** — run the existing GitHub Actions workflow on Ubuntu and
   macOS; require both full test/race jobs and both evidence aggregates green.
   Before every additive push, scan all reachable refs and require zero
   confirmed and zero unclassified PII/secret findings. Do not run the project
   suite locally.

## Dated compiler-identity correction amendment (2026-09-29)

Run `36613290969` confirmed the frozen Phase 4 symbol expectation is corrected
(failed test/subtest names decreased from 125 to 124 on Ubuntu and 113 to 111
on macOS), while the larger identity and provenance groups remain red. Static
comparison found the global branding replacement had also changed `clang` to
`cschway` in current compiler-evidence JSON tags, cache-input names, machine
fact names, and CI receipt labels. That accidental compiler-tool identity
change is separate from the intended Schway project identity. Restore those
names to `clang` in current code and fixture manifests. Leave the archived
Phase 16 run receipt unchanged. Hosted CI must confirm this correction before
revising the remaining evidence IDs.

Run `36616432386` confirms the compiler-identity correction is applied, but
the full test failure counts remain 124 names on Ubuntu and 111 on macOS.
Both hosts still pass `go vet`, `go build`, and Phase 23; both fail full tests
and Phase 6 evidence, so race jobs are skipped. Do not expect this cleanup to
clear unrelated failures. Next inspect Phase 4-5 core serialization fields
and Phase 1-5 evidence IDs separately, then reconcile the current C output,
diagnostic identities, executable controls, and rewritten history/doc refs.

## Dated Phase 7 diagnostic-input amendment (2026-09-29)

Hosted run `36617635572` on `2e39e2f9` passes vet, build, Phase 23, and Linux
libc++ setup, but full checks and Phase 6 evidence remain red; race jobs are
skipped. It identifies a Phase 7 diagnostic ID drift: the `.schway` path rename
left 13 fixtures with comment-only `.lang`→`.schway` text changes, which moved
source spans and also changed pinned payload inputs. Restore those files to the
exact pre-rename bytes while keeping their `.schway` paths, and exclude only
those 13 named paths from the public-identity text count. Their historical
comment references are explicit exceptions; all remain in the all-ref privacy
scan. Its old-ABI text count also
excludes only `internal/compiler/check/check_blame_test.go` and
`internal/compiler/syntax/syntax_test.go` for these fixture-specific
`lang_res_open` expectations; every other identity-family count still
checks both files, and the full privacy scan includes them. The same run shows the Phase 4 test helpers expecting `schway_res_open` from a byte-frozen fixture that declares
`lang_res_open`; update those named expectations to the fixture's actual legacy
symbol. Do not update diagnostic, payload, or golden IDs unless the next hosted
run proves the underlying semantic witness changed.

## Dated Phase 8/11/13 input-preservation amendment (2026-09-29)

Hosted run `36621640310` confirms the Phase 7 diagnostic fixture and Phase 4
frozen-symbol expectation repairs: those failures no longer appear. It exposes
the next `phase08/negative_control_fails.schway` diagnostic baseline mismatch,
sealed Phase 13 held-out fixture digest mismatches, and Phase 11 gate-corpus
fixture/canonical-program digest drift. Restore the nine byte-pinned Phase 8
fixtures, the Phase 11 `multi_function_gate_corpus.schway`, and the five sealed
Phase 13 `heldout_*.schway` fixtures to exact pre-rename bytes under their
`.schway` paths. The rename gate excludes only these listed files' historical
contents; every file stays in the all-ref privacy scan. Keep active Phase 6 and
Phase 13 derivation fixtures on current Schway markers. This does not resolve
the separate Phase 4/5 serialized core and evidence identities, current C
goldens, native controls, or rewritten planning references.

## Dated hosted-CI outcome amendment (2026-09-29)

Commit `2e282d21` restores the nine Phase 8 fixtures, Phase 11 gate corpus,
and five sealed Phase 13 held-out fixtures to exact pre-rename bytes under
`.schway` paths. Its full reachable-ref privacy scan passed across 1,548
commits, 3 tags, and 4,898 blobs with zero confirmed or unclassified privacy
candidates. Hosted run `36624111624` confirms the Phase 7 diagnostic ordering
and the Phase 8/11/13 source pins advance. Both hosts pass vet, build, and Phase
23; full tests and Phase 6 evidence remain red, so race tests are skipped.

Continue in these bounded slices:

1. **Frozen-input compatibility** — update only Phase 11/13 test mutation
   helpers to recognize the actual labels and C identifiers present in their
   byte-frozen inputs. Derive C result names from the frozen Phase 4 source
   symbol instead of assuming the source symbol itself was renamed. Keep all
   historical fixture bytes unchanged. The protocol text gate excludes only
   `internal/compiler/session/session_phase13_injectors.go` for its three
   legacy marker strings; all module, command, suffix, ABI, and path checks
   still scan that file.
2. **Current compiler/evidence identities** — Phase 1-3 core bytes remain
   unchanged, while Phase 4/5 core serialization changes at the current
   `ForeignTypeName` and current C/evidence fingerprints change with Schway ABI
   names. Add field-level old/new records with semantic witnesses; update only
   the related core, C, manifest, and canonical-program pins. Do not regenerate
   frozen artifacts wholesale.
3. **Executable controls** — reconcile Phase 4/5 native runtime, LTO, linker,
   and cleanup-injector controls independently from the digest migration.
4. **History and planning references** — map stale commit IDs through the
   private commit map, refresh groundedness line pins and live corpus counts,
   and preserve archived receipts as historical evidence.
5. **Acceptance** — require full Go tests, race tests, Phase 6 evidence, and
   Phase 23 evidence to pass on the hosted Ubuntu/macOS matrix. Before each
   additive push, rescan all reachable refs and require zero confirmed and
   zero unclassified candidates.

## Dated hosted-CI outcome amendment (2026-09-29, run 36628391253)

Commit `28cb4358` fixes the Phase 11 and Phase 13 frozen-input mutation seams,
derives the C header assertion from the preserved Phase 4 source symbol, and
passes those targeted compatibility controls on both Ubuntu and macOS. The
full workflow remains red: both hosts pass vet, build, and Phase 23; full Go
tests and Phase 6 evidence fail, so race jobs are skipped. CI reports 46
failing top-level tests on Ubuntu and 39 on macOS. The remaining work groups
are current core/C/evidence identity migration, frozen cleanup and event
assertions, Phase 4/5 native and LTO controls, and stale current-document and
history references. The reachable-ref privacy scan for `28cb4358` passed
with zero confirmed and zero unclassified candidates.

Proceed in these reviewable slices:

1. **Frozen C compatibility** — teach the cleanup mutation adapter to match
   the historical marker in byte-frozen Phase 4/5 generated C. Keep that
   one marker exception scoped to `internal/compiler/session/session.go` in
   the protocol identity gate. Make the archived event and by-pointer checks
   assert their historical identifiers; compare current generated C to its
   historical peer after reversing only the public Schway identity strings.
2. **Current identity migration** — record the `ForeignTypeName` core field
   change and current Schway C/evidence identities in explicit old/new
   ledgers, preserving old pins and requiring a field-normalization or
   executable semantic witness before updating a current expectation.
3. **Executable controls** — investigate Phase 4/5 runtime, cleanup, linker,
   unwind, and LTO failures independently of digest changes.
4. **Current references** — refresh active maturity counts and groundedness
   pins, map commit IDs through the private commit map, and preserve archived
   receipts as historical facts.

## Dated hosted-CI outcome amendment (2026-09-29, run 36631949966)

Commit `9edd722d` confirms `TestStreamingEmitterWritesAtPointOfOccurrence`
and `TestPhase5ByPointerLoweringGolden` now pass with the historical labels
and identity-only normalization. The cleanup probe now reaches native linking
and fails because its frozen C calls `_LANG_lang_res_open` while the test
runner supplies only the current Schway implementation; use the existing
legacy resource adapter at that test boundary. `TestProgramMatchDefectEventPrecedesAbort`
shows zero captured bytes: libc leaves the terminal JSON buffered on the pipe,
then `abort()` discards it. Flush the completed terminal record before
aborting. `TestLegacyEmitterEvidence` remains red on generated program
digests; reconcile it in the broader artifact-identity slice.

Run `36631949966` still fails both full Go test jobs and both evidence
aggregates. The cleanup adapter and defect-output fixes are pending their own
hosted receipt; the other remaining identity, native/LTO, and current-reference
groups remain separate work.

## Dated hosted-CI outcome amendment (2026-09-29, run 36633475605)

Commit `6da5ab3b` fixes the abort-output control; `TestProgramMatchDefectEventPrecedesAbort`
no longer appears among the hosted failures. The overall workflow remains red
on both hosts. Run `36633475605` exposed the exact `_LANG_` adapter mismatch in
cleanup/LTO probes, Phase 4/5 foreign-type identity drift in core and frozen
program digests, current Phase 1/2 manifest IDs and C digests, one missing
emitter-consumer registry row, and the renamed Phase 21 receipt fixture path.
Phase 6 also reports Linux undefined-symbol allowlist and LTO divergence
controls; current groundedness and commit-reference checks remain open.

The next additive repair batch uses the exact historical resource adapter at
the cleanup and LTO test boundaries, compares frozen core/program digests only
after reversing the single recorded foreign-type identity field, updates the
two current evidence goldens through the identity ledger, and repairs the
receipt and emitter-consumer inventory. Historical C artifacts and their
artifact hashes remain unchanged. Continue to use hosted CI as the acceptance
gate; local project tests remain excluded.

## Dated hosted-CI outcome amendment (2026-09-29, run 36643785834)

Run `36643785834` confirms the public Schway tree builds and passes vet and the
Phase 23 gate on both hosts, but the full Go suite and Phase 6 evidence remain
red. Its failures group into frozen Phase 4/5 foreign-source links and one
historical program-span digest, current manifest IDs, the Phase 16 emitter
consumer inventory, legacy command operands in derived corpus grading,
Linux-only composition-control behavior, and stale planning indexes. The
groundedness report measured six new R2 command-path findings and eight R2b
findings; seven prior Phase 23 R2b pins no longer match the source documents.

The next additive repair batch fetches full Git history in both checkout jobs;
maps every frozen Phase 4/5 foreign contract to its historical adapter;
normalizes only the evidenced Phase 5 span shifts and current identity IDs;
resolves archived command operands against current package names; and records
the missing emitter refusal witness. It adds the six R2 reconciliation rows,
updates the exact pinned frontier and R2b ownership set, and regenerates both
derived planning views from their registers. The NAT-07 composition-only
control is explicitly scoped to Darwin, where the current hosted receipt
demonstrates divergence. No archived fixture bytes, run records, or commit
history are rewritten. Hosted CI remains the project-test lane; do not run the
project test suite locally.

## Dated hosted-CI outcome amendment (2026-09-29, run 36652395001)

Commit `83ea17bc` is on `main`, and GitHub CI run `36652395001` completed red
on Ubuntu and macOS. Build, vet, and Phase 23 passed on both hosts; both full
test jobs and both Phase 6 aggregates failed. The shared historical adapter
defines aliases for resource, nonlocal, arena, and retained calls while each
fixture link includes only the implementation for its own declared symbol,
so the adapter introduces undefined references. Split the aliases by
operation and wire the matching frozen source sets; Phase 4 combined fixtures
need both resource and nonlocal pairs.

The remaining deterministic evidence failures are a second generated Phase 5
span identity (`enum_foreign_discard1_alt1`), two current manifest IDs, a
Phase 16 emitter call at `cgen_test.go:86` (the previous registry addition
covered a different call at line 675), and the validation-corpus pair digest.
Groundedness now measures R1/R2/R3 at zero, but six obsolete R2 frontier pins
must be removed and two current Phase 23 R2b findings at lines 45 and 47 need
their P23 owners. The suppression-witness failure still needs its exact
diagnostic before repair: the new Linux skip in
`native_lto_test.go:357` lacks a resolvable `probe:` or `env:` citation. Add
the existing named control as its witness. The undefined-symbol controls also
hit the default five-second native timeout on hosted runners; inspect that
path and use a bounded test-specific timeout if the failure is environmental.

Next: complete those bounded repairs, refresh the active planning views from
their registers, run static checks and the full redacted privacy audit, then
push one additive commit. Hosted CI remains the acceptance gate; no local
project tests.

## Dated repair progress amendment (2026-09-29)

The next additive repair batch is prepared against that hosted receipt. The
historical foreign aliases are now operation-specific, including the
allocator and retained-pointer fixtures; combined Phase 4 fixtures link the
resource and nonlocal implementations they declare. Both Phase 5 enum
variants receive the same narrowly scoped generated-span normalization, the
two current manifest IDs match the hosted diagnostics, and the emitter
consumer register now matches all 82 source call sites. Six stale R2 pins
were removed, the Phase 17 R2b pin restored, Phase 23 R2b findings at lines
45 and 47 gained P23 owners, the Darwin-only skip cites its named probe, and
`UndefinedSymbols` uses the existing bounded 30-second subprocess timeout
instead of five seconds. The validation-corpus failure now reports its
manifest digest, computed digest, and pair count to make the next hosted
diagnostic actionable; no frozen record or manifest was rewritten.

Refresh both derived planning views from their registers, finish static and
redacted privacy gates, then push additively to `main`. The next GitHub run
must decide whether the validation-corpus pair digest needs a narrowly
evidenced current expectation update or whether a source pairing is missing.
Do not run project tests locally; no hosted receipt exists yet for this batch.

## Dated hosted-CI outcome amendment (2026-09-29, run 36656246047)

The new adapter batch reached its next boundary: both `go test ./...` jobs and
both Phase 6 aggregates failed. The second generated Phase 5 enum's cgen copy
still needed the measured span normalization; the historical manifest check
also hashes execution outputs, whose Schway schema identity must be reversed
along with C and foreign-manifest identity. Two M004 provenance sentinels still
named call sites before the source-line shifts (`cgen_test.go:Emit:189` and
`session_phase5_corpus_test.go:EmitNative:501`). Ubuntu exposed that the frozen
nonlocal C fixture writes its terminal record and then aborts without flushing
stdout; keep the artifact bytes fixed and make the test-only adapter select
unbuffered stdout before `main`. The corpus consumer now requests 35 pairs,
while its checked-in receipt covers 34, so the new digest diagnostic worked.

The next batch applies these identity/provenance corrections and adds an
opt-in GitHub Actions receipt job: export the consumer's exact pairs, run the
existing sequential producer on Ubuntu, privacy-scan the receipt before its
brief artifact upload, audit it again privately, and only then refresh the
checked-in manifest/record.
Do not infer a successful receipt from the pair digest or rewrite old results.
Keep project tests on GitHub-hosted runners.

## Dated receipt-job amendment (2026-09-29, run 36659581859)

The opt-in Ubuntu receipt job failed in its producer wrapper before any
artifact upload. The pair exporter uses lowercase JSON keys (`package` and
`pattern`); the workflow initially expected Go's untagged field names. No
receipt artifact was created or uploaded. Correct the consumer keys, let the
normal push CI finish, then dispatch the receipt job separately so the
workflow's same-branch concurrency cancellation cannot interrupt CI.

## Dated hosted-CI outcome amendment (2026-09-29, run 36659880866)

The operation-specific legacy links, Linux output capture, M004 provenance
sentinels, rename gate, and public emitter inventory no longer appear among
the failures. Both full test jobs and both Phase 6 aggregates still fail in
three areas. The generated-program digest reaches `phase5.enum_foreign_try1_alt2`
after the earlier `alt1` fixes, so apply the same span reversal to all four
bounded `phase5.enum_foreign_*` variants. The two terminal-defect manifests
normalize to `evidence:7649e234461dab67e77861c7` and
`evidence:b563764272b40d6770f41273`; source inspection identifies the
additional C change as flushing the terminal record before abort. Keep the
original pins `evidence:12e9d68073ab86eb9eb463c9` and
`evidence:c9eea4002b7177106948da43`, and bind those two paths to a separate
behavior-migration ledger with the existing
`TestProgramMatchDefectEventPrecedesAbort` witness and current Schway IDs. The
validation corpus remains at 35 requested pairs versus 34 recorded.

The first manual receipt attempt failed before upload because the workflow
read uppercase pair keys; no artifact was uploaded. The workflow now reads the
export's lowercase keys. Finish the all-variant span and manifest-ledger
edits, let push-triggered CI complete, then dispatch the receipt workflow on
its own, audit the artifact privately, and update the record only from its
actual test results.

## Dated hosted-CI outcome amendment (2026-09-29, run 36660820987)

Both Ubuntu and macOS vet steps failed before running the test suites. The
all-variant span normalizer calls `strings.HasPrefix` in
`session_phase16_frozen_evidence_test.go`, but that file did not import
`strings`. Add the missing import and run local static analysis only; the
hosted test results remain the acceptance evidence for this repair chain.

## Dated hosted-CI outcome amendment (2026-09-29, run 36661109534)

The missing `strings` imports are fixed. Both hosts now pass vet, build, Phase
23, the full test suite except for the corpus-grade receipt check, and both
Phase 6 aggregates except for `TestValidationRowGradesAreEarnedOverArchivedCorpus`.
That remaining check still compares 35 live pairs with a 34-pair historical
record. A separate opt-in receipt run is required after fixing the live pair
resolver; do not reuse the old record.

## Dated receipt audit amendment (2026-09-29, run 36661507880)

The opt-in Ubuntu receipt completed and passed its pre-upload privacy gate. A
second private scan found no PII or secret candidates, and the artifact records
all 35 pair completions with matching manifest digests. One archived
`assert-go-tests.sh` operand still names `./cmd/lang-repair/...`, so its live
subprocess fails after the current-tree rename. Keep this receipt private and
out of the checked-in evidence record. Map archived command operands to the
current package identity at the shared grading and pair-export boundaries,
then regenerate and re-audit the receipt from a new hosted run.

## Dated receipt privacy-scan correction (2026-09-30, run 36663107274)

The first fresh receipt produced after the package-path fix passed the hosted
privacy step, but the independent scan found two phone-shaped matches in
`Output` values. Both matches were fully contained in Go-generated temporary
source paths from `t.TempDir()`; the artifact had no personal home paths,
contacts, secret candidates, or unclassified findings. The temporary GitHub
artifact was deleted after download. The scanner had treated any line with a
`Time` field as timestamp context, even when the candidate was in another
field. It now parses JSON fields, accepts only strict timestamp values and
phone-shaped substrings wholly inside a generated Go test temp-source path,
and requires all other candidates to be classified. The corrected scanner
passed against the private artifact; generate and audit a new hosted receipt
before updating the checked-in record.

## Dated hosted receipt amendment (2026-09-30, run 36670954303)

The corrected opt-in Ubuntu receipt ran against source revision `dc730fff` and
passed its field-aware pre-upload privacy scan. The private download contains
34 exported pairs, 34 exact pair-completion witnesses, one matching batch
witness, matching manifest digests, zero test-fail/build-fail events, and 82
skips. The second privacy scan found no personal home/contact/secret
candidates; five phone-shaped matches were contained in generated Go test
`/tmp/Test.../001/...` paths, with zero unclassified matches. Gitleaks found
zero findings. The record and manifest are copied from that private artifact;
run normal hosted CI again after committing them to confirm the corpus-grade
check now passes.

## Dated receipt-host correction (2026-09-30, run 36675117930)

The refreshed Ubuntu record is not sufficient to grade the full corpus: the
NAT-07 `TestCompositionOnlyLTODivergence` control explicitly skips outside
Darwin, while its archived validation row requires an executed pass. The prior
checked-in receipt contains a Darwin pass for this control. The same CI run
also showed that evidence grading, as well as pair export, must normalize
archived command package operands before resolving live test names. Apply that
mapping in `deriveCeiling` and its focused regression test. Move the opt-in
receipt lane to `macos-latest`, with `TMPDIR=/tmp` so `testing.T.TempDir`
outputs do not record a per-user macOS temporary path. Keep the ordinary
Ubuntu/macOS CI matrix unchanged and produce a new macOS receipt before
replacing the checked-in one.

## Dated Darwin receipt amendment (2026-09-30, run 36677759368)

The opt-in `macos-latest` receipt completed at source revision `fe271a0f` with
`TMPDIR=/tmp`. It records 34 exported pairs, 34 exact pair completions, one
matching batch witness, matching manifest digests, zero test-fail/build-fail
events, and a pass (not skip) for
`TestCompositionOnlyLTODivergence`. The hosted privacy scan and independent
private scan found zero personal home/contact/secret candidates, zero phone
candidates, and zero gitleaks findings. Replace the interim Ubuntu record with
this Mac artifact so the Darwin-only validation row remains supported.

## Dated Phase 6 aggregate diagnostic amendment (2026-09-30, run 36680850688)

The full Ubuntu/macOS checks, including both race suites, passed. Both Phase 6
aggregate jobs returned exit 3. Redacted job logs contain no ordinary Go test
failure; source inspection shows `schway verify testdata/phase6` returns an
operational status while `verify-phase6.sh` redirects its JSON to a temporary
file that the exit trap deletes. Add a failure-only summary of phase status,
non-pass lane IDs, and redacted diagnostic messages so the next hosted run can
identify the failing gate. Preserve the failing exit code and keep acceptance
open until Phase 6 passes.

## Dated Phase 6 diagnostic coverage amendment (2026-09-30, run 36684969719)

The full Ubuntu/macOS check jobs passed, including both race suites, but both
evidence aggregates again exited 3 in `scripts/verify-phase6.sh`. The
failure-only Phase 6 JSON summary did not appear in either hosted log. The
available redacted logs contain only successful package summaries followed by
the final exit code, so they do not identify whether an assertion/test/vet/
build command or an earlier Phase 1–5 corpus stopped the script. Extend
failure-only diagnostics to name each prerequisite command and to report the
redacted lane/diagnostic summary for every Phase 1–6 corpus. Keep the acceptance
gate open and use hosted CI for the next project-suite run.

## Dated script-contract correction amendment (2026-09-30, run 36687496041)

The diagnostic wrapper generalized each verifier invocation, hiding the
literal `json verify testdata/phaseN` text required by
`TestPhase6VerifierScriptContract`. That one test failed in both regular
check jobs and both evidence aggregate jobs; Phase 23 gates passed. Keep each
Phase 1–6 invocation explicit and pass the command as arguments to the failure
reporting helper. The correction passes shell syntax, diff, and source-text
checks. Do not run local project suites; hosted CI remains the acceptance lane.

## Dated Phase16 refusal-boundary amendment (2026-09-30, run 36688618461)

The full Ubuntu/macOS checks passed, including both race suites, while both
Phase 6 evidence aggregates returned exit 3. The redacted diagnostic identified
the Phase 4 `release-omitted` lane: Phase16 intentionally refuses current native
emission for the multi-function foreign-call body in
`acquire_three_success.schway`. Phase16 records that refusal boundary and binds
the supported historical output through frozen evidence; the Phase 4/5 Go
tests cover that receipt and the named refusal. Preserve the refusal as an
operational result, do not claim the live mutation controls passed, and let the
aggregate skip only the superseded Phase 4/5 dynamic-control grep when the
structured refusal is exactly recognized. The current patch has not yet been
run on hosted CI. The worktree gitleaks scan passes with only three known
synthetic `generic-api-key` test-value matches, all classified. The reachable
history scan at base `2140f700` also passes; repeat it after committing so the
new tree and commit message are included. Project suites remain hosted-CI-only.

## Dated emitter-inventory correction amendment (2026-09-30, run 36696297255)

The Phase 6 refusal-boundary handling reached the full hosted suites, but both
Ubuntu and macOS `go test ./...` failed the same structural check:
`TestPhase16PublicEmitterConsumerInventory` found the existing
`session.go:EmitNative` consumer at line 2784 while its source-derived registry
still recorded line 2780. The added refusal branch shifted this later source
location by four lines; update the single registry row, rerun static/privacy
checks, then start another hosted run. The Phase 6 aggregates also failed at
their prerequisite test step, so run 36696297255 does not yet establish the
refusal-aware aggregate result. Do not run project suites locally.

## Dated follow-up inventory amendment (2026-09-30, run 36697168016)

The refreshed Phase 4 source-call row passed, but both hosts again failed
`TestPhase16PublicEmitterConsumerInventory`. Its only unique failure was the
existing `Phase16ControlNativeC` call in `session_phase5.go`: the added
`strings` import moves it from line 25 to line 26, while the registry still
records line 25. Update that registry row and rerun static/privacy checks. The
aggregate jobs stop at the same prerequisite Go test failure, so this run still
does not verify the refusal-aware aggregate result. Do not run project suites
locally.
