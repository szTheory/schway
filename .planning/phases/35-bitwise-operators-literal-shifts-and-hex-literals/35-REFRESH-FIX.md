# Phase 35 evidence adapter repair — 2026-10-15

The first current task-map run reached TestPhase35EvidenceGrounding and failed
its thirteen synthetic coverage entries: the shared summary validator now
requires a non-empty description. This was a test adapter mismatch; the
historical evidence document and compiler behavior were unchanged.

The adapter now supplies `description: Named Phase 35 execution evidence`.
It continues to call the real validator with the declared kind, exact tracked
file/test reference, pass status and explicit human_judgment field. No validator
rule, negative control or historical receipt was weakened.

Fresh focused reproduction after repair passed TestPhase35EvidenceGrounding
and TestSummaryCoverageGroundedness, including the five false-evidence
controls and the shared 12 valid / 59 invalid coverage controls. Raw JSON log:
`/tmp/schway35-evidence-fix.log`. The full 22-task selector union then passed
66 distinct names / 72 package-qualified roots across 13 packages, with zero
missing/fail/skip; exact commands, input hashes and log hashes are in
35-REFRESH-RUN.json. Its source_head is the starting committed HEAD; input
hashes include the repaired working-tree test file before its scoped commit.

The failed pre-repair raw run is retained at
`/tmp/schway35-validation-10.log`; it earns no passing credit. The successful
run uses separate `validation-fixed` log paths. No frozen fixture, production
source, dependency, completed UAT or historical Phase 35 receipt was edited.
