# Phase 35 Byte validator compatibility repair

Date: 2026-10-14. Scope: 35-04 admission repair only. This receipt does not complete 35-06 or change milestone state.

## Cause and repair

Commit 26e4b509 applied the new scalar bridge's unused-Byte law to every linear Byte function in corevalidate.linearStructural and originvalidate.PublishProblemsFor. The frozen pre-change emitter corpus failed before C emission; legacy Byte call targets lost callable admission. New independent LegacyByteCompatibility tests reproduced the defect in Byte copy, ordinary-call/callee, fallible foreign acquisition, nonlocal foreign resource and zero-call identity fixtures.

Each peer now selects the scalar bridge from existing IR authority independently: scalar instruction families, or the scalar builder's reserved Bool fact at `:type:2` / Unit fact at `:type:3`. The source witness is check/scalar.go: checkScalarFunction creates the Byte/U64/Bool/Unit table before lowering every scalar Byte body. check/check.go: checkLinear builds signature facts and ensureLiteralTypeFact adds legacy literal U64 at `:type:u64`. The five historical fixtures in each new compatibility test explicitly check that their Byte functions lack the reserved scalar facts. The complete frozen corpus comparison independently confirms unchanged historical admission, CoreIDs, semantic IDs, three-emitter outputs and refusal precedence, subject only to its previously authored exact malformed-numeric amendment.

Applicability does not depend on correct return type, scalar local eligibility, whether the formal is used, reachability, foreign metadata or call opcode. Existing independent unused-formal and scalar-signature/local laws remain authoritative. New constant-only controls have no scalar instructions and still reject used formal, wrong return, non-scalar local, flat unreachable Byte use, foreign metadata disguise and call opcode disguise. The peers retain separate laws and do not import a new shared admission helper.

Source-blind applicability cannot distinguish a coordinated erasure of all scalar instructions and reserved type-table evidence from legacy IR; this remains within the existing coordinated-source/core-lie boundary. No source authority or IR fields were added. Original source-use checks remain unchanged.

## Fresh executed evidence

All commands used GOCACHE=/tmp/ai-lang-verification-gocache, `-count=1 -v -timeout=600s`; selectors are independently fully anchored. Final executions occurred after the last source/test edit. No selected tests skipped.

| Package | Selector | Final result | Log |
| --- | --- | --- | --- |
| ./internal/compiler/corevalidate | `^TestPhase35ShiftImmediateAdmission$\|^TestPhase35UnusedByteIRAdmission$\|^TestPhase35ScalarTransferMutants$\|^TestPhase35AnalysisTriggers$\|^TestPhase35CorePeerBaseline$\|^TestPhase35LegacyByteCompatibility$` | PASS, 0.042s | /tmp/phase35-byte-compat-final-corevalidate.log |
| ./internal/compiler/originvalidate | `^TestPhase35ShiftImmediateAdmission$\|^TestPhase35UnusedByteIRAdmission$\|^TestPhase35ScalarTransferMutants$\|^TestPhase35AnalysisTriggers$\|^TestPhase35OriginPeerBaseline$\|^TestPhase35LegacyByteCompatibility$` | PASS, 0.026s | /tmp/phase35-byte-compat-final-originvalidate.log |
| ./internal/compiler/session | `^TestPhase35PreChangeEmitterCorpus$\|^TestPhase35PreChangeCorpusControls$\|^TestPhase35PreChangeNumericDeltaControls$` | PASS, 2.590s | /tmp/phase35-byte-compat-final-session.log |
| ./internal/compiler/cgen | `^TestPhase35NativeShiftTracer$\|^TestPhase35EmitterAdmissionGates$\|^TestPhase35UnsignedShiftEmissionAndEventProjection$\|^TestPhase35ShiftSiteControls$` | PASS, 0.652s | /tmp/phase35-byte-compat-final-cgen.log |

The initial corpus reproduction failed (1.303s). Agent-authored legacy tests failed before repair in both peers. Restoring only the old Byte applicability conditions, with new tests and existing eligibility checks retained, again failed LegacyByteCompatibility in core (0.006s) and origin (0.007s). Reapplying produced the final passing results above. Revert logs: /tmp/phase35-byte-compat-revert-corevalidate.log and /tmp/phase35-byte-compat-revert-originvalidate.log.

Guardrail: target, held-out neighboring contracts, non-deleting behavior review, and revert/reconfirm passed. Stryker is not configured for this Go repository; its mutation signal is explicitly skipped. The manual old-applicability mutant was killed by both agent-authored legacy tests. `git diff --check` passed.

## Prevention and limits

The earlier focused 35-04 scalar checks covered malformed new bridges but did not characterize legacy Byte admission. The independent TestPhase35LegacyByteCompatibility regressions now protect that boundary. Existing UnusedByteIRAdmission plus constant-only controls prevent a validity-based applicability carve-out from silently accepting forged bridges. The frozen corpus and its named controls protect exact historical identities/emitter outcomes rather than merely accepting programs.

Changes are limited to corevalidate.go, corevalidate_phase35_test.go, originvalidate.go and originvalidate_phase35_test.go, plus this receipt and the resolved debug record. Frozen sources/fixtures, emitter code, existing C task commits, main STATE/ROADMAP/requirements and the user's config edits are preserved. No full suite, broad native matrix, hosted or Phase36 work was executed. Final commit belongs to the manager.
