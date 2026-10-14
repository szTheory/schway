# Phase 35 Plan Check — Revision Ledger

**Initial independent review:** 2026-10-14; eight plans, 22 tasks, seven waves; 2 blockers, 2 warnings, 1 advisory. No execution permitted before correction and recheck.

| Severity | Dimension | Required property | Initial evidence | Disposition |
|---|---|---|---|---|
| blocker | verify_command_format_sanity | Every automated command parses as a complete shell command | Syntax-only /bin/sh probe: all 22 commands missing selector quote/flags; exit 2 | pending planner revision |
| blocker | research_resolution | Every research question is resolved or explicitly dispositioned | Unmarked original Open Questions despite parent addendum/plan route decisions; tool availability untested | parent resolved four questions with explicit status; tiny local LTO/UBSan probes compile/run exit 0, feature evidence still pending |
| warning | failing_direction | Every fails_when binds in its task verify block | Canonical probe: 22 valid bindings plus 22 orphan duplicate statements | pending planner revision |
| warning | scope_sanity | Plan file scope fits its coherent task/token budget | 35-07 has 12 files versus warning threshold 10; estimate 48k/100k | pending redistribution or explicit justified disposition |
| info | scope_sanity | Token estimates state confidence | All below budget but low confidence/sample count 0 | accepted advisory; not measured actuals |

Decision coverage probe passed 23/23, all four requirements were covered, dependencies were acyclic, and baseline ordering/tracer structure passed initial review. Canonical command-path probe returned not_applicable for these Go forms; that does not establish shell runnability. This ledger preserves the initial findings; final corrected probes and independent checker verdict are appended after revision.

## Targeted revision 1 — 2026-10-14

| Initial issue | Required property satisfied by | Fresh planner evidence |
|---|---|---|
| verify_command_format_sanity blocker | All 22 automated commands contain anchored selectors, closing quotes and -count=1 -v -timeout=600s; inherited whole-plan commands are complete | Actual extracted command text passed /bin/sh -n -c: 22 checked, 0 invalid; four supplemental command syntax checks passed |
| research_resolution blocker | Parent's current research marks all four questions resolved/dispositioned, selecting three emitters and retaining actual feature evidence obligations | Current resolved heading/statuses inspected; parent local LTO/UBSan probes compile/run exit0 are availability only; no Schway feature claim |
| failing_direction warning | Removed 22 duplicate tails; each task has exactly one fails_when immediately after its command inside verify | Canonical verify-failure-directions: exit0,22 bindings,0 blockers,0 warnings |
| 35-07 scope_sanity warning | Example plus exhaustiveDispatchFixtures registration move together to interpreter task35-05-02; public/full-dispatch check follows both engine owners | 35-07 now10 owned files;35-05 now8; same-wave35-05/06 overlap0; all tasks at most5 files; eight plans/22 tasks/seven waves preserved |
| estimate confidence advisory | Updated raw/calibrated ownership estimates35-05=37k and35-07=45k with unchanged low confidence/sample_count0 | Every plan remains below100k; no execution cost claim |

Canonical verify-command-paths: exit0,22 commands,0 blockers,0 warnings, status not_applicable for Go forms. Shell validity is established only by the separate actual-text /bin/sh syntax checks, not that SDK status. All eight frontmatter.validate --schema plan and verify.plan-structure calls returned valid:true/exit0 with no errors or warnings. The exact VALIDATION map matches all22 canonical commands, owned-file lists, fails_when and acceptance clauses. Decision coverage remains23/23; four requirements,13 explicit edge acceptances and two flagged descriptor-less prohibitions remain present. No production tests, pending selectors, configuration edits or commits were performed. Parent's fresh independent checker verdict remains pending.

Detailed planner check output: /tmp/schway35-revision-results.json; reproducible syntax/SDK checker script: /tmp/schway35-revision-check.py. The planner's source/ownership check is not an independent review or executed feature receipt.

## Targeted revision 2 — 2026-10-14

Canonical issue /tmp/schway35-checker-issues-r1.yaml binds the scope_sanity property that every plan stays below the10-file warning threshold. Revision1's35-07 count10 met that threshold; its earlier disposition remains recorded above. The smallest coherent revision moves only literal fixture testdata/phase35/value-vectors.json creation/ownership to35-05 task2, whose interpreter vector owner already creates the exact example and exhaustive registration.35-07 task2 reads the committed fixture after35-05/06 and independently authors engine comparators; native/interpreter expectations never come from producer output. No new plan/task/wave or cross-wave5 test dependency is introduced.

| Remaining issue | Required property satisfied by | Fresh planner evidence |
|---|---|---|
|35-07 scope_sanity warning|All plans own fewer than10 files; fixture has sole owner35-05-02|Distinct file counts35-01..08=5/9/5/4/9/4/9/9; same-wave overlap0; eight plans/22 tasks/seven waves retained|
|Ownership consistency|Frontmatter, task files, artifact lists, read-first, must-haves and exact VALIDATION ownership describe the same create/read paths|All22 VALIDATION command/file/failure/acceptance comparisons pass; task05-02 at5 files and07-02 at3;39k/43k estimates use fresh factor1/sample_count0/low calibration|

Fresh actual-text shell syntax:22 task commands/0 invalid plus4 supplemental checks. All22 automated command texts are unchanged. Canonical verify-failure-directions:exit0/22 bindings/0 blocker/0 warning; verify-command-paths:exit0/22 commands/0 blocker/0 warning, not_applicable for Go forms. Decision coverage:exit0/passed/23 of23. All8 frontmatter and structure checks:valid:true/exit0/no errors or warnings. All4 requirements,13 explicit edge acceptances,2 descriptor-less flagged-unverified prohibitions and old compatibility/lane contracts remain intact. Only05-02's failure conditions gain fixture-specific missing/mismatch detection; its previous conditions and the other21 clauses remain. No feature tests, pending selectors, configuration/source/lifecycle edits or commits ran. Evidence: /tmp/schway35-revision2-results.json. Parent fresh independent checker verdict remains pending.

## Targeted revision 3 — 2026-10-14

The fresh independent checker after revision 2 confirmed the scope correction and identified an execution-metadata blocker: task execution metadata must honor disabled TDD. Project workflow.tdd_mode=false and the delegated phase brief's TDD_MODE=false/no RED-first history contract bind this revision. Eighteen task elements incorrectly carried tdd="true" despite the existing disabled-TDD execution prose.

| Remaining issue | Required property satisfied by | Fresh planner evidence |
|---|---|---|
| Disabled-TDD execution metadata blocker | Changed only the 18 affected task attributes to tdd="false": 35-02/03/04/05 tasks 1–3; 35-06 tasks 1–2; 35-07 tasks 1–3; 35-08 task 1 | Exact before/after comparison: seven plan files changed; 18 false attributes; zero true attributes; all other plan bytes identical |
| Preservation of scope and acceptance | Retained every behavior block, test, automated command, fails_when, acceptance clause, truth, artifact, ownership path, dependency, estimate, decision, edge and prohibition | All 22 automated commands unchanged; exact VALIDATION map still matches all 22 tasks; eight plans/22 tasks/seven waves and file counts 5/9/5/4/9/4/9/9 unchanged; same-wave overlap zero |

Fresh actual-text shell syntax: 22 task commands/zero invalid plus four supplemental checks. All eight frontmatter and structure checks returned valid:true/exit0/no errors or warnings. Canonical verify-failure-directions returned exit0/22 bindings/zero blockers/zero warnings; verify-command-paths returned exit0/22 commands/zero blockers/zero warnings, not_applicable for these Go forms. Decision coverage returned exit0/passed/23 of 23. All four requirements, 13 explicit edge acceptances, two descriptor-less flagged-unverified prohibitions, phase-unique threat rows and full source coverage remain unchanged by the byte-preservation check. Configuration and VALIDATION are unchanged. No source/test/configuration/lifecycle edits, feature tests or commits ran. Evidence: /tmp/schway35-revision3-results.json and /tmp/schway35-revision3-exact.json. This metadata correction and planner validation do not replace the parent's fresh independent checker verdict, which remains pending.

## Final independent verdict — revision 3 (2026-10-14)

Typed gsd-plan-checker returned **VERIFICATION PASSED**, zero blockers and zero warnings.
After the full revision-2 review found a new execution-contract mismatch, the delegated
guidance-and-retry corrected exactly 18 task attributes to tdd="false"; all other
plan bytes and 22 automated commands remained unchanged. This is implementation followed
by focused verification, with no RED-first history asserted. The fresh delta checker
confirmed the correction and retained the full review's cleared semantics and ownership.

Fresh metadata: eight valid schemas and structures; 22 shell-valid commands and exact
validation mappings; canonical command/failure probes zero blockers/warnings (Go command
path forms remain not_applicable, not execution evidence); 23/23 decisions; four
requirements and 13 explicit edges; seven waves, disjoint plans 05/06; file counts
5/9/5/4/9/4/9/9. Two descriptor-less prohibitions remain explicitly flagged until
source inspection and execution supply their evidence. No feature behavior was executed
by planning, and the tiny Clang probes establish local tool availability only.

Runtime limitation: the visible collaboration schema exposes terminal/mailbox completion
but no outputFile required by the optional mtime stall watcher. Native completion was
used and disclosed; no watcher result or configuration override was fabricated.
