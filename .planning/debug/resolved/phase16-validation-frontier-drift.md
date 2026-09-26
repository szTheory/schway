---
status: resolved
trigger: "Archived validation evidence no longer matches the current corpus or groundedness scanner output after the full Go gate."
created: 2026-09-23T00:00:00-04:00
updated: 2026-09-26T19:47:39Z
---

## Current Focus

hypothesis: Confirmed — committed Phase 17 test/index changes altered two consolidated validation corpus pairs after the Phase 16 record was sealed, while committed Phase 16/17 artifacts and the renamed Phase 14 witness produced six findings absent from the Phase 14-era pinned frontier and five R2b keys absent from its ownership registry.
test: Completed by isolated reproduction, live pair export/digest, exact pair diff against record_pair_complete entries, clean workspace checks for all implicated inputs, and line-level commit provenance.
expecting: Confirmed: live digest 07e30b... differs from archived 85a6d7... at exactly the session and cgen consolidated pairs; scanner reports the same six deterministic extras on repeat.
next_action: Return root cause only; preserve all policy and phase decisions and make no fix.

reasoning_checkpoint:
  hypothesis: "Committed corpus/test-index evolution causes deterministic snapshot rejection because the archived pair digest, pinnedFrontier, and r2bLandingPhases were not refreshed atomically with those changes."
  confirming_evidence:
    - "The live producer seam exports 31 pairs with SHA-256 07e30b..., while the manifest binds 31 pairs at revision 8733e5b with SHA-256 85a6d7...."
    - "The exact pair diff changes only session (removed TestB1BlameIsStructurallyUnreachable from the resolved alternation) and cgen (added five TestPhase17ProgramTwoType/TypePair tests)."
    - "The groundedness tests deterministically report six unpinned findings and five ownerless R2b findings, all from committed files with no workspace diff."
  falsification_test: "The hypothesis would be false if the live exported pair bytes matched the manifest or if reverting only post-snapshot corpus/index additions left the same extra findings."
  fix_rationale: "No fix is applied in diagnose-only mode; any remediation must regenerate/re-pin from the live derived corpus and assign the new R2b rows under the existing P20 policy rather than weaken the guards."
  blind_spots: "A full historical checkout replay was not run; exact byte deltas and commit provenance make it unnecessary for identifying the current causal set."
  candidate_causes:
    - "data: committed validation/verification artifacts and test names changed the derived corpus and scanner output"
    - "config/code snapshot: manifest PairDigest, pinnedFrontier, and r2bLandingPhases remained at older derived state"
    - "environment: Go cache or local workspace mutation"
  and_gate: "yes — each failure requires both a committed derived-input change and a stale exact snapshot/ownership registry; environment/workspace is ruled out."

## Symptoms

expected: Archived validation evidence matches the current requested corpus, and every retained verification command is assigned by the groundedness scanner.
actual: The archived corpus-pair digest differs from the current requested corpus, and new R2b/unparseable findings appear in Phase 14, Phase 16, and Phase 17 artifacts, including Phase16-VALIDATION commands at lines 48-49.
errors: "TestValidationRowGradesAreEarnedOverArchivedCorpus: checked-in validation corpus pair digest does not match the current requested corpus; TestVerificationGroundednessFrontierIsPinned and TestVerificationGroundednessThreeClassesAreEmpty: new R2b/unparseable findings"
reproduction: "GOCACHE=/tmp/ai-lang-go-cache go test ./... (diagnosis will isolate the three named tests)"
started: Observed during Phase 16 UAT test 21 after the full Go integration gate.

## Eliminated


## Evidence

- timestamp: 2026-09-23T00:00:00-04:00
  checked: Required project context and Phase 16 UAT
  found: UAT gap G-16-21-B groups one archived corpus digest failure with groundedness frontier failures; STATE records a recent reconciliation that reduced the pinned frontier from 124 to 58 entries while retaining per-branch R2b under Phase 20/QLT-10 ownership.
  implication: The investigation must separate a corpus identity mismatch from newly scanned command ownership findings and preserve the recorded policy that R2b can remain pinned.
- timestamp: 2026-09-23T22:00:00-04:00
  checked: Debug knowledge base and the archived Phase 16 validation run record
  found: No exact prior resolution exists. The checked-in Phase 16 run record already captured the same deterministic failures on 2026-09-20, including unparseable 16-RESEARCH.md:144, R2b 16-VALIDATION.md:48-49, and an R2b count of 12 with the two Phase 16 rows lacking landing phases.
  implication: The problem predates the present full-suite invocation and is not a transient Go cache failure; the run record itself was archived from a red corpus-wide execution.
- timestamp: 2026-09-23T22:08:00-04:00
  checked: Isolated rerun of the three named session tests with count=1
  found: All three fail deterministically. The exact extra frontier is Phase 14 VALIDATION:74 (R2b), Phase 16 RESEARCH:144 (unparseable), Phase 16 VALIDATION:48-49 (R2b), and Phase 17 VERIFICATION:86-87 (R2b). Five R2b rows have no landing phase; the live scan reports 550 enforced-tier documents, 628 commands, and 15 total R2b entries.
  implication: There are two distinct deterministic snapshot failures: the exact frontier literal omits six current findings, and the landing registry omits five current R2b findings. The unparseable Phase 16 research prose affects the frontier pin but not the R2b ownership count.
- timestamp: 2026-09-23T22:16:00-04:00
  checked: Digest and groundedness implementations, source lines, workspace diff, and initial history
  found: The corpus consumer derives a sorted/consolidated package-pattern list from every current primary validation row and refuses unless its SHA-256 equals the manifest. The manifest is pinned to revision 8733e5b, digest 85a6d729..., 31 pairs, produced 2026-09-21T03:17:15Z. The current implicated source/test/snapshot files have no workspace diff. Phase 16 final validation changed after the producer revision (commit 1218c57), and Phase 17 verification changed on 2026-09-22 (59d6664/c7f3692).
  implication: Uncommitted workspace state is eliminated as the cause. At least part of the digest/frontier mismatch is committed corpus drift after the archived evidence revision; exact pair and scanner deltas remain to be isolated.
- timestamp: 2026-09-23T22:24:00-04:00
  checked: Live validation pair export and exact diff against archived record_pair_complete entries
  found: The live producer seam still derives 31 consolidated pairs but hashes to 07e30b53013291e76df6cd4d01d06ddbf4184be0fa33c230000aa9c17d4814cb rather than manifest 85a6d72957ccf508d16259f370126e2d44c938526d17a0ac9f7f0e83f7e5976d. Exactly two pair values changed: session lost TestB1BlameIsStructurallyUnreachable after commit 163fc8b ratified the M006 blame boundary, and cgen gained five Phase 17 two-type tests introduced from b49e8ae onward.
  implication: The digest mismatch is proven committed test-index/corpus drift after the Phase 16 artifact was sealed, not an unexplained hash or workspace problem.
- timestamp: 2026-09-23T22:24:00-04:00
  checked: Blame/history for all six groundedness findings
  found: Phase 16 RESEARCH:144 was committed 7576784 on 2026-09-19; Phase 16 VALIDATION:48-49 were committed 1218c57 on 2026-09-21; Phase 17 VERIFICATION:86-87 were committed 59d6664 on 2026-09-22. Phase 14 VALIDATION:74 dates to 0dcb460, but became an R2b finding when TestB1BlameIsStructurallyUnreachable was removed/renamed by 163fc8b on 2026-09-22. pinnedFrontier/r2bLandingPhases were last materially derived in Phase 14 and contain only the original 10 owned R2b rows.
  implication: The frontier failure is stale exact derived state across both document growth and test-index evolution. Existing policy remains valid: R1/R2/R3 stay empty, while all R2b findings must remain pinned and owned by P20.

## Resolution

root_cause: "The failures require two conditions: (1) committed derived inputs changed after the snapshots were sealed — Phase 17 removed/renamed TestB1BlameIsStructurallyUnreachable, added five cgen two-type tests, and Phase 16/17 added scanner-visible commands — and (2) the Phase 16 corpus manifest/run record plus Phase 14 pinnedFrontier/r2bLandingPhases were not regenerated/reconciled atomically. This changes the live pair digest from 85a6d7... to 07e30b... and leaves six scanner findings unpinned, including five R2b findings without their required P20 ownership. Workspace changes and Go cache behavior are not causal."
fix: "Not applied (find_root_cause_only). Preserve existing policy; remediation direction is to regenerate the corpus artifact from the live producer and reconcile/pin the six findings, assigning surviving R2b rows under the existing P20/QLT-10 decision."
verification: "Three isolated tests fail deterministically; live producer export and exact pair diff prove the digest delta; git diff is empty for every implicated input/snapshot; blame identifies committed provenance for all six scanner findings."
files_changed: []

## Post-diagnosis closure (2026-09-26)

Plans 16-17/18 refreshed the consumer-derived corpus record and reconciled the exact scanner frontier; Phase 20 Plans 07/10 then reconciled the final live frontier and regenerated the complete 33-pair corpus record. Their groundedness, reconciliation, and archived-grade checks passed. Phase 16 UAT gap G-16-21-B is recorded resolved, and the refreshed Phase 16 verification passes. The original diagnosis-only scope is preserved above; this section records the later closure evidence.
