# Phase 29 Wave-0 Gate Verdict

Written 2026-10-11 by plan 29-01. Gate plans: plans 29-00 (tracer) and 29-01 (expansion). There is no standalone spike (D-10); the pair is the gate.

## Flags

| Flag | Pass | Evidence |
|------|------|----------|
| (a) ByteView through a call: adler32 is published Callable | yes | TestPhase29GateCallable (session) builds the originvalidate interface and finds adler32 Callable with one ByteView parameter; corevalidate admits it. TestPhase29GateInterpreterVectors returns every pinned literal. |
| (b) Parser accepts byte_len and byte_at in scalar expressions | yes | The library fixture parses and checks clean (phase29Library in session_phase29_gate_test.go); parser.go scalarExprAtom carries the two atoms. |
| (c) Release-while-view-live | yes, PROCEED-compatible | Releases are compiler-minted (no source spelling); no owner or loan parameter is admissible on a loop function (pinned type.scalar_signature, TestPhase29NeighborBaseline); the semantic release-while-view-live and move-owner laws are owned by plan 29-09. |
| (e) Byte ops are event-free on success | yes | Interpreter half: TestPhase29GateInterpreterVectors asserts no event is attributed to either byte op. Native half: schema2ExecutionDocumentSize in cgen_program.go treats OpBytesLen as event-free and reserves only the terminal defect event for OpBytesAt. |
| (f) Fixture bytes | yes, route recorded | D-15's entry-parameter fixture shape is the Stage-B route (plan 29-12); the owner-shape fixture of the research flag is not needed at the gate. |

Native contract at the gate: TestPhase29GateNativeProxyVectors returns all seven pinned literals (empty 1, 0x00 65537, a 6422626, Wikipedia 300286872, 4096 x 0x00 268435457, 5552 x 0xFF 4052720524, [0x80,0xFF] 33620352) and exits 65 with stderr exactly "schway: byte index out of bounds" plus a newline and empty stdout on an out-of-bounds read, at -O0, -O3 and -O3 -flto. The bounds test is an ordinary if before the load in both shells; no assert, no restrict, no elision. The ByteView ability derivations agree (TestPhase29ByteViewAbilityAgreement, with an armed mismatch control).

## Stage-A cost

This is a proxy, gate-only reading; never the cap record (D-14). Toolchain: Go 1.24.0 darwin arm64; Apple clang version 21.0.0 (clang-2100.1.1.101); target darwin arm64 (Apple silicon). Native flags: -std=c17 -Wall -Wextra -Werror -pedantic plus the lane flags. Each cell is median over 6 timed runs (first of 7 discarded) and max, workload x 0xFF; every run asserted the oracle checksum (2170679522, 2967061426, 2006388466). Native times are whole-process wall time and are dominated by process start, so they bound cost from above only.

| Lane | 4,096 median / max | 16,384 median / max | 65,536 median / max |
|------|--------------------|---------------------|---------------------|
| interpreter, plain | 5.43 ms / 6.02 ms | 20.11 ms / 20.52 ms | 76.87 ms / 78.50 ms |
| interpreter, race | 20.16 ms / 21.86 ms | 59.78 ms / 61.54 ms | 218.13 ms / 225.78 ms |
| native proxy -O0 | 2.02 ms / 2.57 ms | 1.98 ms / 2.09 ms | 2.69 ms / 3.89 ms |
| native proxy -O3 | 1.94 ms / 4.04 ms | 1.70 ms / 1.96 ms | 2.23 ms / 2.96 ms |
| native proxy -O3 -flto | 2.08 ms / 2.21 ms | 1.86 ms / 2.01 ms | 2.10 ms / 3.58 ms |

## Linear scaling

Ratio of the 65,536 max to the 4,096 max (input grows 16x): interpreter plain 13.0x; interpreter race 10.3x; native -O0 1.5x; native -O3 0.7x; native -O3 -flto 1.6x. The interpreter is near linear (sub-linear only because of fixed startup); native lanes are flat at process-start noise, so no superlinear growth is visible on any lane.

## Budget comparison

D-13 budget at 65,536: native 3 s, interpreter 12 s. STOP threshold at 4x budget: native 12 s, interpreter 48 s. Slowest lane max at 65,536: interpreter race at 225.78 ms (about 1.9% of the 12 s interpreter budget); slowest native lane max 3.89 ms (about 0.13% of the 3 s native budget). No lane approaches budget, let alone 4x budget.

The 4x multiplier is a planner choice (plan 29-01) that operationalizes D-14's "far above the D-13 budget"; it is not a user decision and Stage B (plan 29-12) alone keeps or lowers the cap.

## Decision

Flag (a) passed, flag (c) is as recorded above, and no lane exceeds 4x budget.

verdict: PROCEED
