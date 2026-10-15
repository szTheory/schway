# Phase 35 mechanical prohibition wiring — 2026-10-15

Independent verification at c5a22142 established 38/38 behavior truths but
returned human_needed because two 35-08 acceptance entries retained unresolved
status and null tiers. This is missing mechanical acceptance wiring, not a
subjective or external user-authority decision. The existing Go evidence
checker already rejects both false causal/numeric claims and false lane claims.

The dated plan amendment now declares both items resolved/test, with flat
check_kind/check_target/check_violation_fixture/check_clean_fixture fields
accepted by GSD's existing projection. The standard-library Node adapter runs
only TestPhase35EvidenceGrounding and checks its exact RUN/PASS, exit and no
fail/skip. It contains no copy of the evidence logic. The Go test accepts the
GSD_PROHIB_SUBJECT convention for the test-only subject; relative paths resolve
from the project root despite Go's package working directory. Production code,
normal test subject and historical receipt are unchanged.

The two known-bad text fixtures copy the historical receipt with exactly one
false clause each (plus an explicit known-bad header). They are test subjects,
not new passing receipts. Causal-as-numeric and absent-lane claims must fail
at the actual Go assertion. The unchanged 35-EVIDENCE.md is the clean control.

Canonical parseMustHavesBlock/descriptorFromProjection generated each request;
check prohibition-enforcement executed the producer. Both results are green,
flagged=false, located=true, failFirst=true, passed=true, with machine proof
violation-fixture. Each entails bad-subject RED, clean-subject GREEN and normal
GREEN. Full requests/results and source hashes: 35-PROHIBITION-ENFORCEMENT.json.
An initial relative-path attempt remained unverified because both subjects
failed to open from Go's package cwd; it earned no passing credit. Resolving
relative subjects from the project root repaired it; the clean causation
control prevented that setup failure from being counted as semantic proof.

No new dependency, package installation, evaluator, runtime boundary or human
approval was introduced. The prior null-tier report remains in commit history.
Independent review, final regression and verifier re-verification follow this
wiring; no phase completion is claimed by this record.

## Report command formatting correction

The first post-wiring regression passed the runtime/mutation checks but
failed two groundedness roots because eight verifier commands were placed in
Markdown table cells with unescaped regex alternation. The exact commands
were moved to fenced shell blocks, retaining the reported checks and verdict.
The final verifier still owns the new verdict and fingerprint. Fresh
TestVerificationGroundednessFrontierIsPinned,
TestVerificationGroundednessThreeClassesAreEmpty and
TestPhase35EvidenceGrounding passed after the formatting repair; raw log is
/tmp/schway35-report-format-fix.log. No groundedness rule or frontier was
relaxed, and the failed earlier run earns no passing credit. The final full
focused regression is recorded separately before acceptance.
