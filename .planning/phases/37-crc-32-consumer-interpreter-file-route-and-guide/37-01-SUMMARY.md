---
phase: 37-crc-32-consumer-interpreter-file-route-and-guide
plan: "01"
subsystem: compiler
tags: [crc32, native, modules, evidence]

requires:
  - phase: 36-scalar-helper-calls
    provides: one-argument imported calls and admitted bitwise operators for the CRC recurrence
  - phase: 31-bounded-file-checksum-and-second-consumer
    provides: retained native file-byte owner, view, release, and local C binding contract
provides:
  - Table-free CRC-32/ISO-HDLC module with one private eight-round byte helper
  - Full and low16 retained-native consumers sharing the same module and byte-identical local binding
  - Exact additive Phase37 source ledger composed independently by both historical inventory owners
  - Pre-change baseline and canonical fail-first prohibition receipts for the two named existing boundaries
affects: [phase-37, interpreter-file-route, executable-guide]

actuals:
  tokens: 23038
  tasks: 3
  commits: 5
  plan_head_before: b1b00b84a5d46d0075424a67461df1dee9d4166f
  plan_head_after: 4384e383e39c776448e63480b642251982883dd1

tech-stack:
  added: []
  patterns:
    - Table-free reflected CRC recurrence with a documented low-32-bit state invariant
    - Exact additive source ledgers remain separate from historical frozen corpora
    - Flat prohibition descriptors are produced through projectProhibitions and checked by the canonical producer

key-files:
  created:
    - examples/crc32/lib/crc32.schway
    - examples/crc32/crc32.schway
    - examples/crc32/crc_low.schway
    - examples/crc32/adapter.c
    - examples/crc32/adapter.h
    - examples/crc32/file_bytes.bindings.json
    - testdata/phase37/new-fixture-ledger.json
    - .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-BASELINE.md
    - .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-01-red-evidence.json
    - .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-prohibition-check.test.cjs
    - .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-PROHIBITION-EVIDENCE.md
    - .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/prohibition-clean.json
    - .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/prohibition-outcome-violation.json
    - .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/prohibition-causal-violation.json
  modified:
    - internal/compiler/session/session_phase37_crc32_test.go
    - internal/compiler/session/session_phase34_prechange_test.go
    - internal/compiler/session/session_phase36_prechange_test.go
    - .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-01-PLAN.md

key-decisions:
  - "Use the reflected ISO-HDLC polynomial 0xEDB88320, all-ones seed/final XOR, and one private crc_byte call per byte."
  - "Keep the existing file-bytes ABI and copy its local adapter triple byte-for-byte for the second consumer."
  - "Treat causal acceptance as structural evidence; only the native literal-vector test establishes this CRC numeric result."

patterns-established:
  - "One shared ordinary-source CRC module serves both entries; neither entry carries a byte or bit loop."
  - "Each historical inventory validates its own ledger schema and finite paths before composing the Phase37 source set."

requirements-completed: [APP-13, MOD-07, APP-14]

coverage:
  - id: D1
    description: "A real nine-byte file produces the literal CRC-32/ISO-HDLC answer through the retained native application route."
    requirement: APP-13
    verification:
      - kind: integration
        ref: "internal/compiler/session/session_phase37_crc32_test.go#TestPhase37CRC32NativeTracer"
        status: pass
    human_judgment: false
  - id: D2
    description: "The second entry imports the same CRC module, prints 14630, and uses an exact local copy of the checksum binding triple."
    requirement: MOD-07
    verification:
      - kind: integration
        ref: "internal/compiler/session/session_phase37_crc32_test.go#TestPhase37CRC32Consumers"
        status: pass
    human_judgment: false
  - id: D3
    description: "The Phase37 additive ledger closes against current tracked sources in both historical inventory owners, with omission, hash, path, ordering, and exemption controls."
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase34_prechange_test.go#TestPhase37Phase34AdditionLedgerControls"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase36_prechange_test.go#TestPhase37Phase36AdditionLedgerControls"
        status: pass
    human_judgment: false
  - id: D4
    description: "The two recorded prohibitions have canonical bad, clean, and default proofs bound to named existing assertions; causal evidence is not treated as numeric replay."
    verification:
      - kind: other
        ref: ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-prohibition-check.test.cjs#--prove"
        status: pass
    human_judgment: false

metrics:
  duration: 21min
  completed: 2026-10-15
status: complete
---

# Phase 37 Plan 01: CRC-32 Native Tracer Summary

**A shared, table-free CRC-32/ISO-HDLC module now powers full and low16 native file consumers, with exact inventory and fail-first proof wiring.**

## Performance

- **Duration:** 21 minutes, from 2026-10-15T14:01:58Z through 2026-10-15T14:22:29Z
- **Tasks:** 3
- **Files changed:** 18 plan-owned files

## Accomplishments

- Implemented reflected CRC-32/ISO-HDLC with polynomial `0xEDB88320`, all-ones seed/final XOR, a documented low-32-bit state invariant, and a private `crc_byte(U64)` helper called once per byte for eight rounds.
- Built `examples/crc32/crc32.schway` and `crc_low.schway` over the same imported module. On the real file `123456789`, the native full entry exited 0 with stdout `3421780262\n`; the low16 entry exited 0 with stdout `14630\n`; both had empty stderr.
- Copied `adapter.c`, `adapter.h`, and `file_bytes.bindings.json` byte-for-byte from the established checksum consumer. The existing empty-path result remains pinned at `(65, "", "")`.
- Added the exact three-source Phase37 ledger and independent validation/composition in the Phase34 and Phase36 historical owners. Their frozen corpus bytes and the original Phase36 ledger remain unchanged.
- Captured the pre-change route, no-call C, inventory, and toolchain baseline before creating a Schway source or editing an inventory owner.
- Proved both Phase37 prohibitions with the installed canonical producer. Each bad fixture reaches only its named existing Go assertion; clean and default controls pass. The original unresolved projection remains unchanged.

## Task Commits

1. **Task 1 RED evidence:** `2590dab5` — `test(37-01): add failing CRC native tracer`
2. **Task 1 implementation and inventory:** `5945f78e` — `feat(37-01): complete native CRC-32 source inventory`
3. **Task 2 consumers:** `c210004e` — `feat(37-01): bind both CRC consumers to local file adapter`
4. **Task 3 prohibition proof:** `4f2eb354` — `test(37-01): wire canonical prohibition evidence`
5. **Task 3 final receipt refresh:** `4384e383` — `docs(37-01): refresh canonical prohibition receipt`

The task-start ledger is `b1b00b84a5d46d0075424a67461df1dee9d4166f`; the measured five plan commits end at `4384e383e39c776448e63480b642251982883dd1`. The final planning metadata commit is recorded in the executor completion result.

## Verification

Pre-change baseline commands, all with `GOCACHE=/tmp/ai-lang-verification-gocache`, reached and passed every named root before source creation:

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^(TestPhase23PublicFileByte|TestPhase22AppRunKeepsOpaqueTokenStreamsAndChildStatus|TestPhase22ConformanceRunStillEmitsExecutionDocument)$' -count=1 -v -timeout=600s
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase36NoCallGeneratedCByteBaseline$' -count=1 -v -timeout=600s
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase34PreChangeCProgramGolden|TestPhase34PreChangeCGoldenIsNotInert|TestPhase36NoCallGeneratedCByteBaseline|TestPhase36NoCallGeneratedCByteBaselineControls)$' -count=1 -v -timeout=600s
```

After the implementation and again at final verification, these exact focused selectors passed:

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^(TestPhase23PublicFileByte|TestPhase22AppRunKeepsOpaqueTokenStreamsAndChildStatus|TestPhase22ConformanceRunStillEmitsExecutionDocument)$' -count=1 -v -timeout=600s
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase37CRC32NativeTracer|TestPhase37CRC32Consumers)$' -count=1 -v -timeout=600s
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase34PreChangeCProgramGolden|TestPhase34PreChangeCGoldenIsNotInert|TestPhase36NoCallGeneratedCByteBaseline|TestPhase36NoCallGeneratedCByteBaselineControls|TestPhase37Phase34AdditionLedgerControls|TestPhase37Phase36AdditionLedgerControls)$' -count=1 -v -timeout=600s
node --test .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-prohibition-check.test.cjs
node .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-prohibition-check.test.cjs --prove
```

`--prove` regenerated [the proof receipt](./37-PROHIBITION-EVIDENCE.md) and verified exact flat plan fields against `projectProhibitions`. The canonical producer was separately run for the outcome and causal requests; both returned `status=green`, `flagged=false`, and `failFirstProof=violation-fixture`. The TDD RED was accepted by `gsd_run check tdd-red-evidence .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-01-red-evidence.json --raw`; its named native output assertion observed `9\n` where `3421780262\n` was expected. The complete native answer is checked against a literal independent vector, while the causal peer proof makes no numeric replay claim.

Historical bindings were also checked with:

```sh
cmp -s examples/checksum/adapter.c examples/crc32/adapter.c
cmp -s examples/checksum/adapter.h examples/crc32/adapter.h
cmp -s examples/checksum/file_bytes.bindings.json examples/crc32/file_bytes.bindings.json
```

The Phase34 corpus digest remains `903010268526ce71dab023e562714a5a8e18329a4d91adca9678e5d64d50042f`; the Phase36 corpus digest remains `b9db7faa10bae93e741f2b24e602a9eda3eb393ed38efb0df6364a9ef9873001`; the original Phase36 ledger digest remains `99ee245b707175e1b27846ce825ae6371e7eb7da0e810b2d8edf486c3517eaf4`.

## Files Created/Modified

- `examples/crc32/lib/crc32.schway` — shared CRC algorithm and private helper.
- `examples/crc32/crc32.schway`, `examples/crc32/crc_low.schway` — full and low16 file consumers.
- `examples/crc32/adapter.c`, `adapter.h`, `file_bytes.bindings.json` — exact copies of the existing checksum binding.
- `internal/compiler/session/session_phase37_crc32_test.go` — full native tracer and dual-consumer tests.
- `internal/compiler/session/session_phase34_prechange_test.go`, `session_phase36_prechange_test.go` — independently validated Phase37 inventory composition and negative controls.
- `testdata/phase37/new-fixture-ledger.json` — sorted, hash-bound three-source ledger.
- `37-BASELINE.md`, `37-01-red-evidence.json` — before-mutation baseline and accepted RED event.
- `37-prohibition-check.test.cjs`, the three closed fixtures, and `37-PROHIBITION-EVIDENCE.md` — phase-local canonical proof adapter and receipt.
- `37-01-PLAN.md` — flat canonical descriptors added after successful bad/clean/default producer proofs; initial unresolved history remains in `37-PROHIBITION-PROJECTION.json`.

## Decisions Made

- Keep the CRC implementation table-free and in ordinary Schway source; both entries import the one private-helper implementation.
- Keep the existing file-bytes binding unchanged and copy-local for the second build closure.
- Keep numeric CRC evidence distinct from the causal peer's structural acceptance evidence.

## Deviations from Plan

None — the plan executed as written.

## Issues Encountered

- The first tracer scaffold had no admitted scalar operation, so C emission refused before the intended RED. Replacing the scaffold result with `byte_len(bytes)` made the native build executable; the accepted RED then reached the exact numeric mismatch (`9\n` versus `3421780262\n`).
- The first Task 3 commit attempt caught an extra trailing blank line in the generated evidence file. The writer now trims trailing whitespace; the adapter, Node test, and `--prove` were rerun before the successful commit.

## User Setup Required

None — no new dependency or external service is required.

## Next Phase Readiness

Plan 37-01 is complete and the phase remains in progress with seven plans left. The ordinary-source CRC library and both native consumers are ready for the later interpreter file route and executable guide. This plan does not claim interpreter CLI parity, the full CRC boundary/fault matrix, sanitizer evidence, or hosted receipts; numeric evidence covers the `123456789` vector and low16 projection only.

## Self-Check: PASSED

All 18 created/modified plan files and this summary exist. The five recorded plan commits (`2590dab5`, `5945f78e`, `c210004e`, `4f2eb354`, `4384e383`) are ancestors of the current HEAD, and whitespace checks pass for the proof receipt and summary.

---
*Phase: 37-crc-32-consumer-interpreter-file-route-and-guide*
*Completed: 2026-10-15*
