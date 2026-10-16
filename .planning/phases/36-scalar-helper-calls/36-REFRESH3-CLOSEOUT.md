# Phase 36 native-defect compatibility refresh closeout — 2026-10-16

Phase 36 independently verifies 32/32 must-haves, 24 artifact declarations,
19 links and 26/26 context decisions. The canonical report covers 102 files:
`v3:sha256:c90fbf587b476354621ea29e16323bb1d73741c6e287b39b0e28c656ddaf3dc0`.
CALL-01/02 and all eight completed plans remain complete. No plan or UAT replay.

## Evidence and corrections

Fresh owner acceptance passes 86/86 roots in 13 packages (14 commands,
34 plan-selector groups), 74.718 seconds. The bounded prior/current regression
union passes 252 root executions in 16 packages, 137.091 seconds: 240 distinct
names and 251 package/name selectors, with one checker test executing in both
internal and external test packages. There are zero missing, failed or skipped
roots. Input/post-input and raw-log hashes were checked independently by the
verifier. Both mechanical prohibitions have expected bad RED and clean/default
GREEN results plus canonical green/unflagged enforcement output.

The independent verifier additionally ran TestPhase33ScalarExplicitDefectShells
successfully. That witness distinguishes structured conformance defect output
(exit 0 with a defect document) from ordinary application stderr/exit 65.
It does not establish legacy no-call native defect-document parity.

Current review adds no finding and retains WR-02 as deferred comment debt.
Security retains all 24 authored Level 1 mitigations closed; UI is N/A.
The shared verification sequence, exact selectors, hashes and local evidence
are recorded in [36-REFRESH3-GATES.md](36-REFRESH3-GATES.md),
[36-REFRESH3-RUN.json](36-REFRESH3-RUN.json),
[36-REFRESH3-REGRESSION.json](36-REFRESH3-REGRESSION.json) and
[36-REFRESH3-PROHIBITIONS.json](36-REFRESH3-PROHIBITIONS.json).

Parent review caught missing canonical requirements/roadmap and direct
caller/bounds authority files in the verifier's initial coverage, trailing
whitespace and imprecise shell/outcome prose. The report owner corrected them,
regenerated its CLI fingerprint and confirmed canonical passed status.
The exact TestVerificationGroundedness and
TestVerificationGroundednessThreeClassesAreEmpty roots passed together in
1.490 seconds; git diff --check passed after correction. No guard was weakened.

## Completion and transition

The authoritative phase.complete 36 call ran once and returned 8/8 plans,
no warnings, no milestone conflict and no indeterminate freshness.
[36-REFRESH3-COMPLETION.json](36-REFRESH3-COMPLETION.json) preserves that result.
ROADMAP and REQUIREMENTS already matched, so neither was rewritten.

Fresh init.progress reports phases 33/34/35/36 passed, Phase 37 stale and
Phase 38 unplanned. Next command is `$gsd-execute-phase 37`, for gates and
verifier only, then re-query. Mandatory M007 audit/gap closure must explicitly
disposition Phase 33 CR-03/04/06 before Phase 38 planning/freeze/hosted binding.
They remain deferred BLOCKER obligations, not fixes or accepted risks.
The canonical state.json next-unimplemented Phase 38 projection is retained;
STATE.md explains the earlier real freshness gate.

Post-completion transition starts after phase.complete, without calling it
again. PROJECT's capability/requirement claims and the shared living documents
remain substantively accurate; they need no status-only rewrite. The three
ranked next capabilities and relevant dormant-seed dispositions are recorded
in [36-REFRESH3-TRANSITION.md](36-REFRESH3-TRANSITION.md). No future consumer or
seed was promoted. Global learnings is disabled; no recent LEARNINGS item,
matching pending phase todo, stale phase handoff or other active workstream
requires action.

Strict state validation reports only expected S004 because unplanned Phase 38
has no directory. No phase directory or plan was fabricated to suppress it.

## Limits and preservation

This refresh changes planning/evidence/review artifacts only. Compiler source,
tests, frozen fixtures, historical receipts and completed UAT are preserved.
The user's light/standard xhigh config edits and persistent preferences remain
uncommitted; the ephemeral chain flag is false. No full local project/package
suite, current Linux/hosted result, human-readability result, new dependency,
backend or cold/warm latency distribution is claimed. The run stops at Phase 36
closeout; no next phase, push or hosted dispatch was started.
