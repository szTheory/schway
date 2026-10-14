
# Phase 35-02 execution compatibility amendment — 2026-10-14

During Task 2's focused inherited selector, newly valid XOR/shifts remove the two
historical parser diagnostics for the reconstructed CRC draft. This follows the locked
D-03/D-04 feature, not a change to Phase 34 recovery semantics. Task 1 is committed at
7958eebf. Depth Task 2 remains uncommitted; the executor returned a terminal checkpoint.

The user explicitly delegated coherent recommendations and auto-follow. The reported
blocking-human label was an inferred scope decision, not external authority or an
irreversible architectural decision. This amendment resolves that routine scope issue
under the active execute workflow; it does not invent a human approval.

Required property: keep Phase 34 source, before/after fixtures and delta ledger unchanged;
add a literal, narrowly attributed Phase 35 recovery after-state row for crc32-reconstructed.
Its historical before projection is the exact two code/span/cause rows from draft-after;
its current projection has zero parser diagnostics and the crc32 sibling. No numeric
CRC result or whole-program CRC admission is implied: this is parser evidence only.

Task 2 scope gains three necessary compatibility files:
- internal/compiler/syntax/parser_phase34_recovery_test.go
- internal/compiler/session/session_phase34_compat_test.go
- testdata/phase35/recovery-after.json

The syntax corpus compares all rows normally, applying only a validated exact Phase 35
literal delta. The session compatibility ledger continues to prove the original Phase 34
before/after delta as historical evidence, and independently checks the Phase 35 current
projection. Reject absent/extra/duplicate delta rows, mismatched literal before projection,
changed current result or sibling, and mutation of frozen files. Do not auto-capture expected
values from current output, broadly skip CRC, or weaken unrelated recovery controls.

Fresh required focused selectors include TestPhase34RecoveryCorpus,
TestPhase34RecoveryControls, TestPhase34CompatibilityDeltaLedger and a new
TestPhase35RecoveryDeltaControls for reached/restored missing/wrong-row controls.
Record this Rule 3 verification-boundary deviation and exact fresh evidence in the summary.
Keep the existing source/numeric/index baseline immutable. No source feature scope,
dependency, public CLI, wire contract, or Phase 36 work is added.
