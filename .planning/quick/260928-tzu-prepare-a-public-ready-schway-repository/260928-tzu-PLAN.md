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
    exclusions=(':(exclude)testdata/phase1/**' ':(exclude)testdata/phase2/**' ':(exclude)testdata/phase3/**' ':(exclude)testdata/phase4/**' ':(exclude)testdata/phase5/**' ':(exclude)testdata/phase16/historical/**' ':(exclude)testdata/phase16/validation-corpus-run-record.jsonl' ':(exclude)internal/compiler/native/foreign_legacy.go' ':(exclude)internal/compiler/check/frozen_foreign_expectation_test.go' ':(exclude).claude/skills/*/sources/**')
    count_text() { git grep --cached -I -l -E "$1" -- "${scope[@]}" "${exclusions[@]}" | wc -l | tr -d '[:space:]'; }
    old_module=$(count_text 'github[.]com/codename-lang/lang')
    old_cli=$(count_text 'cmd/lang(-repair)?(/|[^[:alnum:]_-])|/lang(-repair)?([^[:alnum:]_-]|$)|(^|[^[:alnum:]_])lang-repair([^[:alnum:]_-]|$)|(^|[^[:alnum:]_])lang[[:space:]]+(app|build|check|run|verify)([^[:alnum:]_]|$)')
    old_extension=$(count_text '[.]lang([^[:alnum:]_]|$)')
    # Already-versioned lang.* schema names are frozen wire identifiers.
    # Reject old lang: annotation labels; current namespaces stay schway.*.
    old_protocol=$(count_text '(^|[^[:alnum:]_])lang:')
    old_abi=$(count_text '(^|[^[:alnum:]_])(lang_|LANG_)')
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
