# Phase 33 scalar explicit-defect repair — 2026-10-15

Independent incremental review found CR-05: emitProgramScalarFunction used
emitProgramApplicationScalarFailure for every explicit OpDefect, including
EmitNative. The final checker-admitted U64 reproducer reaches a defect at
entry and inside a nested scalar helper. Original generated C references
schway_write_capacity_capture, which only the application shell defines;
all six native O0/O3/O3-LTO cases fail compilation. Application controls pass.
This is a real source defect, not merely a missing terminal at runtime.

The final repair branches on shell and dynamic invocation support: application
and legacy no-call emission retain their handler; native scalar callers use
emitProgramDefectTerminal plus exit0, exactly as
adjacent checked-add/remainder/bounds native defects already do. The native
runner expects a structured defect document; exit0 is a harness transport
choice, not successful language evaluation. No numeric replay, checker law,
callee signature, optimizer attribute, wire field, dependency or backend changed.

TestPhase33ScalarExplicitDefectShells independently pins the literal reason,
empty defect value, final function.defected event and causal peer validation
on both interpreter and native traces. It runs entry and nested-helper cases
through O0/O3/O3-LTO and independently checks exact application empty stdout,
exit65 and stderr. The new EmitNative call at cgen_phase33_scalar_defect_test.go:66
is registered with its covering probe under the existing schema2 inventory.

## Reached evidence

- /tmp/schway33-r3-explicit-defect-final-red.log: identical final test plus
  original source639b4965 through /tmp/schway33-r3-original-overlay.json;
  six native lane failures name the undefined application-only helper,
  both application controls pass.
- /tmp/schway33-r3-explicit-defect-final-green.log: repaired source, same test,
  all entry/helper native and application cases pass.
- /tmp/schway33-r3-fix-safety.log: five named cgen roots pass, including both
  caller-shape regressions, checked overflow and existing match-defect output.
- /tmp/schway33-r3-inventory-final.log: exact inventory and inventory mutation
  roots pass; no missing/fail/skip root is credited.

Initial exploratory drafts are not the RED proof: Byte explicit-defect source
was refused by the existing checker, unused U64 produced an unrelated C warning,
and an early runner omitted ExpectDefect. The final source uses its U64 parameter
and explicitly sets the defect expectation. No Byte/ByteView explicit-defect
admission or execution claim follows. The reviewer must amend its initial
caller-shape/exit-only description to these observed limits.

CR-03/04 remain separate deferred BLOCKER propagation findings for the mandatory
M007 audit before Phase38 freeze. WR-02 in Phase36 remains deferred comment debt.
Final broad regression and independent goal verification are separate gates.

## Compatibility correction and residual legacy debt

The first broad shell branch, committed at4b9d2ff8, changed eight frozen-C
baseline/control roots for the no-call sum_to_n corpus. That failed union is
preserved in /tmp/schway33-r3-regression-13.log and is not green evidence.
The final branch repairs native explicit defects only when dynamic invocation
support is active; application and legacy no-call emission retain their bytes.
/tmp/schway33-r3-fix-final-safety.log passes five cgen roots, and
/tmp/schway33-r3-nocall-final.log passes all eight baseline/control roots.
No frozen golden, historical receipt or compatibility test was rewritten.

The remaining no-call defect behavior is real audit debt, not silently waived:
/tmp/schway33-r3-nocall-debt.log uses a temporary test overlay removing the only
helper call; all three native lanes reach exit65 and the application stderr
rather than a structured terminal. /tmp/schway33-r3-nocall-original-debt.log
repeats the identical probe against original639b4965 and reaches the same
failure. The application control passes in both. These are deliberately
failing diagnostic probes, not credited green tests or native parity.
The emitter maintainer and M007 auditor must explicitly resolve this debt
against the no-call byte-baseline obligation before Phase38 freeze, through
scoped gap work and a dated compatibility amendment when required. Independent
review owns its finding ID and severity. CR-03/04 retain separate peer owners.
