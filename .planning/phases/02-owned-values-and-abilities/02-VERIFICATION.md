---
phase: 02-owned-values-and-abilities
verified: 2026-09-03T23:26:43Z
status: gaps_found
score: 13/14 must-haves verified
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: passed
  previous_score: 14/14
  gaps_closed: []
  gaps_remaining:
    - "Restore byte-identical Phase 1 generated C and evidence while retaining global collision freedom."
  regressions:
    - "Plan 02-07's Phase 1 byte-identical C/evidence golden contract was broken by unconditional ordinal C identifier renaming."
gaps:
  - truth: "Native execution remains strict and bounded while Phase 1 schemas and goldens remain byte-identical."
    status: failed
    reason: "Commit 91a6206 rewrote every generated identifier, changing both the Phase 1 C golden and its evidence digest despite the plan's explicit no-golden-rewrite prohibition."
    artifacts:
      - path: "internal/compiler/cgen/cgen.go"
        issue: "Uses unconditional ordinal names even when the prior source-derived name is globally collision-free."
      - path: "testdata/phase1/generated.golden.c"
        issue: "SHA-256 changed from f3e4fa6b... to 5143458d... relative to the completed pre-fix Phase 2 state."
      - path: "testdata/phase1/evidence.golden.json"
        issue: "C digest and evidence identity changed as a consequence of the C golden rewrite."
    missing:
      - "Use one global C-name allocator that preserves prior source-derived names when unique and deterministically suffixes only actual collisions."
      - "Restore the Phase 1 C and evidence goldens byte-for-byte, or add a developer-accepted verification override documenting the intentional compatibility break."
---

# Phase 2: Owned Values and Abilities Verification Report

**Phase Goal:** The end-to-end compiler distinguishes cheap implicit copy from explicit transfer and derives independent value abilities coherently.
**Verified:** 2026-09-03T23:26:43Z
**Status:** gaps_found
**Re-verification:** Yes — after commits `91a6206` and `a8c14da`

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|---|---|---|
| 1 | A valid owned-buffer program transfers exactly once and has equivalent interpreter/native events. | ✓ VERIFIED | Fresh `TestOwnedTransferInterpreterNative` and full gate passed. Runtime C appends one transfer and one return event; O0/O3 equal the interpreter. |
| 2 | Use after move and move during an active loan reject with stable causes and smallest repairs. | ✓ VERIFIED | Fresh causal diagnostic tests passed; strengthened identity-aware oracle covers shadowing and multiple owners/loans. |
| 3 | Copy, drop, share, send, and escape derive independently through aggregate/generic shapes. | ✓ VERIFIED | Fresh Box/Pair source and exhaustive 32-leaf/1,024-Pair selectors passed; validator recomputation remains independent. |
| 4 | A canonical Byte program copies implicitly and leaves its source initialized. | ✓ VERIFIED | Fresh `TestImplicitByteCopy` passed through source, checker, and interpreter. |
| 5 | Native execution remains strict and bounded while Phase 1 schemas and goldens remain byte-identical. | ✗ FAILED | Runtime/schema compatibility passes, but `91a6206` rewrote `testdata/phase1/generated.golden.c` and cascaded a new `testdata/phase1/evidence.golden.json`; both differ from their completed pre-fix Phase 2 bytes. |
| 6 | Private leaf/Pair masks agree with an independent oracle without production arbitrary masks. | ✓ VERIFIED | Exhaustive ability selectors and private-combiner spy passed; no production arbitrary-mask constructor is exposed. |
| 7 | Nested owned type syntax is lossless, canonical, recoverable, and bounded. | ✓ VERIFIED | Fresh round-trip, recovery, token-budget, and type-limit tests passed; all declared input-shape caps are enforced. |
| 8 | The checker agrees with an independent model on intermediate states and counted work. | ✓ VERIFIED | Fresh 36-symbol exhaustive oracle, multi-owner/shadowing regressions, and 10/100/1,000/10,000 work-series tests passed. |
| 9 | A source-blind validator rejects forged abilities, ambiguous IDs, and illegal ownership transitions before engines. | ✓ VERIFIED | Fresh mutation, target-freshness, reorder, scale, and unvalidated-core selectors passed; every semantic consumer validates. |
| 10 | The coordinated source-to-core lie remains a named expected escape. | ✓ VERIFIED | `escape:coordinated-source-core-lie` appeared only in `expected_escapes`, never as a detected control. |
| 11 | Compile/run stdout and stderr are separately bounded and native stdout is strictly decoded. | ✓ VERIFIED | Four max-plus-one streams remain independent; duplicate keys, unsupported schemas/kinds, invalid ordering, malformed/trailing/oversized documents fail closed. |
| 12 | Evidence binds validated owned source/core/C/execution facts without overclaiming. | ✓ VERIFIED | Fresh binding, mutation-matrix, tool-probe stdout/stderr/timeout bounds, and current-golden tests passed; `/1` remains `content-identity-only`. |
| 13 | One bounded gate preserves Phase 1 and observes every Phase 2 control with nonzero work. | ✓ VERIFIED | Fresh `scripts/verify-phase2.sh` exited 0: Phase 1 work 18, Phase 2 work 51, five passing Phase 2 lanes, all eight controls. |
| 14 | Targeted test commands fail closed when a requested target is absent. | ✓ VERIFIED | A deliberately absent selector was rejected as `target not discovered`; the full gate's harness self-test passed. |

**Score:** 13/14 truths verified (0 present, behavior-unverified)

### Required Artifacts

| Artifact | Expected | Status | Details |
|---|---|---|---|
| `testdata/phase2/*.lang` | Positive and causal-negative corpus | ✓ VERIFIED | Six substantive fixtures enter production paths. |
| `internal/compiler/ability/ability.go` | Sealed five-ability derivation | ✓ VERIFIED | Wired from checker; exhaustive oracle and validator cover it. |
| `internal/compiler/check/check.go` | Copy/move/borrow/return lowering | ✓ VERIFIED | Identity-aware last-use, stable causes, work accounting, and mixed-body rejection are active. |
| `internal/compiler/corevalidate/corevalidate.go` | Independent fail-closed admission | ✓ VERIFIED | Enforces schema, IDs, fresh ordinal targets, abilities, and ownership transitions. |
| `internal/compiler/interp/interp.go` | Independent execution | ✓ VERIFIED | Validated core drives ordered `/1` facts. |
| `internal/compiler/cgen/cgen.go` | Runtime-causal C17 | ⚠️ PARTIAL | Linear causality and global collision avoidance pass, but unconditional ordinal naming violates the Phase 1 byte-stability contract. |
| `internal/compiler/native/native.go` | Strict bounded native boundary | ✓ VERIFIED | Four streams and duplicate-key/schema-aware one-document decode are wired. |
| `internal/compiler/session/session.go` | Differential and controls | ✓ VERIFIED | O0/O3 comparison, synchronized mutation, bounded reads, and work are wired. |
| `internal/compiler/evidence/evidence.go` | Honest validated binding | ✓ VERIFIED | Bounded strict decode and core admission precede binding. |
| `scripts/assert-go-tests.sh` | Exact selector | ✓ VERIFIED | Missing targets fail; self-test runs in the gate. |
| `scripts/verify-phase2.sh` | Phase 1+2 gate | ✓ VERIFIED | Builds once and exercises tests/race/vet, both corpora, and observations. |

All 21 plan-declared artifacts passed automated existence/substance checks. None is missing, stubbed, or orphaned.

### Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| source | syntax → checker | production session frontend | ✓ WIRED | Bounded source is parsed/checked; mixed schemas fail before core emission. |
| checker | ability/core | sealed derivation and explicit operations | ✓ WIRED | Type/place/loan/operation facts materialize with counted work. |
| semantic consumers | corevalidate | validation before consumption | ✓ WIRED | Interpreter, C generation, native session, and evidence validate first. |
| generated C | native execution | runtime event writes and terminal-place serialization | ✓ WIRED | Native and mutation tests prove executed state flows into facts. |
| native | session comparison | strict decode and canonical equality | ✓ WIRED | Complete interpreter/O0/O3 records compare. |
| mutation seam | native runner | exact transfer-site XOR at O0/O3 | ✓ WIRED | Drift is `native.engine_mismatch`/exit 4; both modes are synchronized and recorded. |
| evidence | validated products | schema/digest/content binding | ✓ WIRED | Source/core/C/executions/tool facts feed deterministic `/1`. |
| gate | shipped CLI | prebuilt binary verifies both corpora | ✓ WIRED | Mechanical plan check misses obsolete `go run` text, but build-once wiring is stronger and passed. |

### Data-Flow Trace (Level 4)

| Artifact | Data | Source | Real | Status |
|---|---|---|---|---|
| checked core | abilities/places/operations | bounded canonical source | Yes | ✓ FLOWING |
| interpreter | outcome/events | validated-core replay | Yes | ✓ FLOWING |
| native | outcome/events | terminal C place and operation sites | Yes; exact-site mutation detected | ✓ FLOWING |
| evidence | digests/executions | validated compiler products | Yes, within content-only boundary | ✓ FLOWING |
| controls | lane IDs/work | actual rejection/execution paths | Yes | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|---|---|---|---|
| Full gate | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/verify-phase2.sh` | Exit 0; tests/race/vet/corpora passed | ✓ PASS |
| Backend causality | exact `TestOwnedBackendMutationIsMismatch` | Real Clang O0/O3 mutation produced mismatch exit 4 | ✓ PASS |
| Original cross-category collision | native CLI on `Thing` / `thing` / `thing_LANG_THING` source | Pass; both executions returned correctly | ✓ PASS |
| Owned session path | six exact selectors | Transfer/parity/abilities/mixed schema/name corpus/concurrency passed | ✓ PASS |
| Ownership model | three exact checker selectors | Exhaustive, multi-owner, and scale passed | ✓ PASS |
| Core admission | four exact validator selectors | Mutation, scale, reorder, and fail-closed execution passed | ✓ PASS |
| Decoder/bounds | exact native plus three CLI selectors | Malformed facts and over-limit inputs rejected | ✓ PASS |
| Evidence | three exact selectors | Binding/mutations/Phase 1 compatibility passed | ✓ PASS |

### Probe Execution

| Probe | Command | Result | Status |
|---|---|---|---|
| `scripts/verify-phase2.sh` | `sh scripts/verify-phase2.sh` | Exit 0; Phase 1 work 18, Phase 2 work 51 | PASS |

Twenty-sample fresh p95 observations: format 75.5 µs, check 118 µs, interpreter 421.5 µs, native 1.27 s, full verify 1.39 s. Peak RSS was honestly unavailable. These are observations, not ratified SLOs.

### Requirements Coverage

| Requirement | Plans | Status | Evidence |
|---|---|---|---|
| OWN-01 | 01, 03, 04, 05, 06, 07 | ✓ SATISFIED | Explicit transfer, identity-aware loans, independent validation, interpreter/native parity, and real backend causality passed. |
| OWN-02 | 01, 02, 04, 06 | ✓ SATISFIED | Real Box/Pair source, exhaustive private oracle, sealed authority, and validator recomputation passed. |

No Phase 2 requirement is orphaned.

### Test Quality Audit

| Area | Req | Active | Skipped | Circular | Assertion | Verdict |
|---|---|---:|---:|---|---|---|
| session/native causality | OWN-01 | active | 0 | No | behavioral/value | PASS |
| ownership oracle | OWN-01 | active | 0 | No | exhaustive behavioral | PASS |
| ability oracle | OWN-02 | active | 0 | No | exhaustive value | PASS |
| validator/evidence mutations | OWN-01/02 | active | 0 | No | value/rejection | PASS |
| syntax/input bounds | OWN-01/02 | active | 0 | No | boundary/property | PASS |

**Disabled requirement tests:** 0. **Circular patterns:** 0. **Insufficient assertions:** 0 for the phase contract.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|---|---:|---|---|---|
| `internal/compiler/cgen/cgen.go` | 51-73 | Unconditional ordinal naming fixes collisions by changing all existing C output. | 🛑 Blocker | Violates Plan 02-07's byte-identical Phase 1 golden truth and explicit no-golden-rewrite prohibition. |

The prior cross-category C-name collision is closed, and evidence tool identity probes are now bounded on stdout, stderr, and time. No unreferenced debt markers, disabled requirement tests, merged native streams, Go-side linear replay, or unbounded compiler entry point was found.

### Decision Coverage

SKIPPED — no Phase 02 `CONTEXT.md` or decisions block exists.

### Human Verification Required

N/A — compiler-foundation phase with no user-facing elements. All acceptance behavior is exercised programmatically.

### Gaps Summary

One compatibility gap remains. The collision repair works behaviorally, but its unconditional ordinal renaming violates the phase's explicit Phase 1 byte-stability contract. Preserve existing names when globally unique and suffix only collisions, restoring both Phase 1 goldens; alternatively, the developer must explicitly accept this deviation with an override. All ownership, ability, causality, boundedness, evidence, and control behavior otherwise passes.

---

_Verified: 2026-09-03T23:26:43Z_

_Verifier: the agent (gsd-verifier)_
