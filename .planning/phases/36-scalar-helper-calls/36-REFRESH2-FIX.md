# Phase 36 caller-shape repair — 2026-10-15

Independent review CR-01 found checker-admitted structured callers that failed
native emission unless another eligible caller enabled the global dynamic flag.
The exact original-source Go overlay now reproduces both cases at their named
emission assertions. The final test itself is unchanged between that overlay
and the repaired run; compile/setup failures do not count as reproduction.

The native classifier now includes the existing unused-Byte-to-U64 bridge and
ByteView-to-Unit body. The Byte formal must remain unused in all six operation
reference fields; helper callee signatures remain exactly U64/Bool pairs. No
checker/peer admission, wire field, optimizer attribute or dependency changed.

TestPhase36UnusedByteStructuredHelper has the program's only helper call in
its structured Byte entry. Independent literal 11 passes interpreter and
native O0/O3/O3-LTO; each execution passes causal attribution separately.
TestPhase36ByteViewUnitHelperEmission retains an uncalled Unit body in an
application translation unit, compiles/links it, and runs the unrelated U64
entry to literal 7 plus newline. This proves translation-unit integration,
not runtime execution of that Unit body or a new ByteView entry/value-call route.

The first local test draft used an invalid synthetic Byte input (3/03). The
existing conformance Byte shell requires its canned input 7; correcting the
test input changed no language or expected-answer contract. A transient test
insertion syntax error is likewise not credited as RED evidence. Only the
final original-source overlay and named native-emission failures earn RED.

Current mapped/regression gates and independent re-review/goal verification
follow this repair. WR-02 remains deferred documentation debt; Phase33 CR-03/04
remain separate BLOCKER propagation debt through the M007 audit before freeze.

## Local reproduction logs

- /tmp/schway36-r2-caller-shapes-original-red.log, sha256 656dcc55f0e5393148222ff96346db66f6628e0acd12ce6b11433640641083dd
- /tmp/schway36-r2-caller-shapes-green.log, sha256 0c955c68870d5536659fce4bb5bb7852eff00d5ffef72f750449e6f7dfd7cf64
