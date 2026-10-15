---
phase: 37-crc-32-consumer-interpreter-file-route-and-guide
plan: "06"
subsystem: testing
tags: [crc32, public-cli, interpreter, native, module-identity, mutation-controls]

requires:
  - phase: "37-03"
    provides: "CRC-32 file consumers, shared library, and public interpreter/native file route"
  - phase: "37-05"
    provides: "Independent literal CRC-32 and low-16 output values"
provides:
  - "Four admitted CRC algorithm mutants with reached numeric disagreements on both public engines"
  - "A fifth shared-library result-bit mutation that changes both public consumers and both engines"
  - "Scoped AST, merged-import, module-digest, InputID, and native receipt identity evidence for the two consumers"
affects: ["Phase 37 acceptance", "native build identity verification"]

actuals:
  tokens: 7928 # 31,712 diff characters / 4 across the three changed artifacts
  tasks: 2
  commits: 2 # measured from plan_head_before
commits: 2
plan_head_before: 4b61046133a8f3393c92439c402b5aadf1fa144c
plan_head_after: 492ab7e3241e321e8dd78bbb2f589470b939718e

tech-stack:
  added: []
  patterns:
    - "Use exact-anchor private source copies for finite semantic mutants; require public admission and successful numeric output before counting a disagreement."
    - "Use InputID for reproducible source/closure identity and retain executable-bound BuildID semantics with the existing non-cacheable receipt posture."
key-files:
  created:
    - cmd/schway/main_phase37_crc32_controls_test.go
    - internal/compiler/session/session_phase37_reuse_test.go
  modified:
    - .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-06-PLAN.md
key-decisions:
  - "Keep all five semantic mutations in private temporary copies and compare clean and mutated public outputs to independent literal values before checking parity."
  - "Pin no-edit source/closure stability with InputID and module rows; BuildID remains bound to exact executable bytes and may vary when those bytes differ."
  - "Leave APP-13, MOD-07, and APP-14 pending parent phase acceptance; these tests record plan evidence only."
patterns-established:
  - "A semantic mutant is reached only after public check success and successful numeric output on interpreter and native routes."
  - "Module identity claims combine scoped syntax census, actual merged imports/calls, immutable entry bytes, receipt module digests, and InputID changes."

requirements-completed: [] # Plan evidence recorded; parent phase acceptance owns requirement completion.
coverage:
  - id: D1
    description: "Four algorithm mutants and the shared-library result-bit mutation reach independent numeric disagreements on both public engines, with clean outputs pinned first."
    verification:
      - kind: integration
        ref: "cmd/schway/main_phase37_crc32_controls_test.go#TestPhase37CRC32AlgorithmMutants"
        status: pass
      - kind: integration
        ref: "cmd/schway/main_phase37_crc32_controls_test.go#TestPhase37CRC32LibraryDualFlip"
        status: pass
    human_judgment: false
  - id: D2
    description: "The two scoped entries import one shared CRC implementation; neither entry contains CRC loops or copied CRC functions, and the historical crc_step helper is accounted for separately."
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase37_reuse_test.go#TestPhase37CRC32SingleCopyAndIdentity"
        status: pass
    human_judgment: false
  - id: D3
    description: "A library-only edit changes both importers' library digests, InputIDs, BuildIDs, and executable digests while preserving entry digests and entry/adapter bytes; no-edit InputIDs and module rows remain stable."
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase37_reuse_test.go#TestPhase37CRC32SingleCopyAndIdentity"
        status: pass
    human_judgment: false

duration: "14m 36s"
completed: 2026-10-15
status: complete
---

# Phase 37 Plan 06: Reached CRC Mutants and Shared-Library Identity Summary

**Five CRC-32 semantic mutants now reach numeric disagreements through the public engines, while both consumers bind to the shared module digest.**

## Performance

- **Duration:** 14 minutes 36 seconds
- **Started:** 2026-10-15T16:34:28Z
- **Completed:** 2026-10-15T16:49:04Z
- **Tasks:** 2
- **Plan task commits:** 2
- **Task-owned files created:** 2 test files; the active plan also received a dated acceptance correction.

## Accomplishments

- Reached four independent CRC algorithm mutations through public `check`, interpreter, and native routes. Every clean result matched its fixed literal, every changed result was a canonical decimal value that disagreed with that literal, and interpreter/native outputs agreed with each other.
- Reached the fifth mutation by flipping one result bit in the shared library. Both `crc32` and `crc_low` changed by one for the same input on both engines.
- Verified both entry closures import and call exactly one `lib.crc32.crc32` function. The byte and bit loops live in the shared library, neither entry contains a copied CRC function or loop, and `examples/helper_calls/crc_step.schway` remains a separately scoped historical helper.
- Verified a library-only edit changes each importer’s `lib.crc32` digest, `InputID`, `BuildID`, and executable digest. Both entry module digests, entry source bytes, and the adapter C/header/manifest bytes stayed unchanged.
- Confirmed no-edit module rows and `InputID` values stay stable. On this host, repeated native links produced different executable digests and corresponding `BuildID` values; the receipts remained non-cacheable with `dependency_closure: incomplete`. This is an observed host result, not a claim that native binaries are generally nondeterministic.
- Kept all mutations in six-file temporary closures and checked the tracked origin hashes before and after each public control and the session identity test. No package or production source was changed.

## Reached Mutants

| Control | Discriminator and clean literal | Mutated output: interpreter / native |
| --- | --- | --- |
| Wrong reflected polynomial | `123456789` → `3421780262` | `3485321504` / `3485321504` |
| Non-reflected recurrence | `80` → `1069182125` | `544806248613` / `544806248613` |
| Missing final XOR | empty file → `0` | `4294967295` / `4294967295` |
| Shortened bit-round bound | `00` → `3523407757` | `2751848219` / `2751848219` |
| Shared-library result-bit flip | Full `123456789`: `3421780262`; low16: `14630` | Full `3421780263`; low16 `14631`, on both engines |

For each algorithm control, both the clean and mutated source passed the public check route. Native runs built and exited successfully before the numeric mismatch was counted. The tests reject tool failures, timeouts, diagnostics, malformed output, and nonzero runtime exits before evaluating a semantic disagreement. No selector skipped or had zero roots.

## Verification

- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^TestPhase37CRC32AlgorithmMutants$' -count=1 -v -timeout=600s` — passed; all four named mutants reached both engines.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^TestPhase37CRC32LibraryDualFlip$' -count=1 -v -timeout=600s` — passed; both consumers reached both engines.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase37CRC32SingleCopyAndIdentity$' -count=1 -v -timeout=600s` — passed; scoped census, merged import/call evidence, module digests, `InputID`, and receipt checks all reached.

## Tracked Origin SHA-256 Pins

| File | SHA-256 |
| --- | --- |
| `examples/crc32/crc32.schway` | `8a8c7570409a6bf96258a35b5964dc899a2347871b564fb59e081beb011388c9` |
| `examples/crc32/crc_low.schway` | `797f9cd2e733c3590a593518fe671a1fda553eb6ce22ea50d70f1a26d606dccc` |
| `examples/crc32/lib/crc32.schway` | `304d47059ac5f5266776bb250794a3c64ae76ce9fe7b55a2a490facf1e35fb4f` |
| `examples/crc32/adapter.c` | `b758836a1d6dcff2fc6b4f903c0046ba3269ec83a8f3329c03553a90a5fbb1ae` |
| `examples/crc32/adapter.h` | `2c29a00ed8939668179fddf1c41cdd5db23e5a9d8574b5ec3e841dae729b612d` |
| `examples/crc32/file_bytes.bindings.json` | `c2f960ff2a92a5d9a60356711812ea60595a2f81cf6c0ff2c4b259ce6380103c` |

## Task Commits

1. **Task 1: Reach the four CRC algorithm mutants** — `15c7812c` (`test(37-06): add reached CRC32 algorithm mutation controls`). The task-1 and task-2 public tests share the same declared test file, so the fifth public dual-flip test was included in this file commit and then independently run with its exact selector before task 2 closed.
2. **Task 2: Prove shared-library identity reuse** — `492ab7e3` (`test(37-06): prove CRC library identity reuse`). This includes the session identity test and the dated plan correction below.

The plan metadata commit records this summary and canonical state/roadmap updates.

## Dated Plan Amendment

The original no-edit assertion expected `BuildID`/`Identity.ID()` to remain identical across a native rebuild. The existing `BuildIdentityInputs.InputID()` intentionally excludes `ExecutableDigest`, while `BuildID` binds exact executable bytes. On the current Apple toolchain, no-edit outputs had stable module rows and `InputID` but varying executable digests and therefore varying `BuildID`. The plan was amended before task 2 closeout to pin source/closure stability with `InputID`, require receipt self-consistency, require equal `BuildID` when executable digests match, and record variance under the existing non-cacheable, incomplete-closure posture. Production linker flags and UUID behavior were left unchanged.

## Files Created/Modified

- `cmd/schway/main_phase37_crc32_controls_test.go` — finite public algorithm and shared-library mutation controls.
- `internal/compiler/session/session_phase37_reuse_test.go` — scoped single-copy, merged import/call, module digest, `InputID`, and native receipt evidence.
- `.planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-06-PLAN.md` — dated clarification of no-edit identity semantics.

## Decisions Made

- Keep mutation inputs, source changes, and adapter files in private temporary copies; pin clean output before comparing engines.
- Use `InputID` for no-edit source/closure stability and preserve `BuildID` as an identity over the exact built executable and its inputs.
- Keep APP-13, MOD-07, and APP-14 pending parent phase acceptance; this summary does not mark broad requirements complete.

## Deviations from Plan

The active plan received the dated evidence correction above before task 2 closeout. No production linker change was made, and the plan’s module/output invalidation requirement remains fully checked. Both plan tasks are complete.

## Issues Encountered

The initial no-edit `BuildID` assertion failed because `ExecutableDigest` changed while source and module closure inputs were unchanged. The evidence correction used the existing `InputID` contract and retained exact artifact binding and receipt posture. The final focused selector passed.

## User Setup Required

None.

## Next Phase Readiness

Plan 37-06’s evidence is complete. APP-13, MOD-07, and APP-14 remain pending parent acceptance, and the parent executor owns wave/phase acceptance and living-roadmap follow-through.

---
*Phase: 37-crc-32-consumer-interpreter-file-route-and-guide*
*Completed: 2026-10-15*

## Self-Check: PASSED

- Summary file exists at the required phase path.
- Task commits `15c7812c` and `492ab7e3` are ancestors of `HEAD`.
- Stub scan found no TODO, FIXME, placeholder, coming-soon text, or skipped tests in the two created test files.
