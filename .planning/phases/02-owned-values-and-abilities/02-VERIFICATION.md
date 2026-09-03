---
phase: 02-owned-values-and-abilities
verified: 2026-09-03T21:53:07Z
status: gaps_found
score: 13/14 must-haves verified
behavior_unverified: 0
overrides_applied: 0
gaps:
  - truth: "A valid owned-buffer program transfers exactly once and has equivalent interpreter/native events."
    status: failed
    reason: "The native execution document is precomputed by Go code generation and printed as a constant, so it is not causally derived from the C operations that Clang executes. A verifier mutation changed the runtime returned buffer and the binary still emitted the unchanged successful outcome and events."
    artifacts:
      - path: "internal/compiler/cgen/cgen.go"
        issue: "emitLinear calls linearExecution in Go, JSON-encodes that result, and emits puts(<constant JSON>); the C return place is only cast to void."
      - path: "internal/compiler/session/session_test.go"
        issue: "The O0/O3 test compares decoded documents and the mutation matrix injects documents through a fake runner; neither injects backend operation drift and observes it through runtime-derived output."
    missing:
      - "Make generated C serialize the terminal value and semantic events from runtime state produced by the emitted operations."
      - "Add a causal backend negative control that mutates an emitted C operation/value and proves the O0/O3 differential fails."
---

# Phase 2: Owned Values and Abilities Verification Report

**Phase Goal:** The end-to-end compiler distinguishes cheap implicit copy from explicit transfer and derives independent value abilities coherently.
**Verified:** 2026-09-03T21:53:07Z
**Status:** gaps_found
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|---|---|---|
| 1 | A valid owned-buffer program transfers exactly once and has equivalent interpreter/native events. | ✗ FAILED | Source, checker, core, validator, interpreter, and C artifacts exist, and the ordinary O0/O3 run matches. However, `cgen.emitLinear` computes `linearExecution` in Go and emits its JSON as a constant `puts` string (`cgen.go:123-127,174,190-217`). A fresh mutation of the compiled C changed `lang_value_delivered.bytes[0]` to `9`; the binary still reported outcome `01020304` and the original events. Native equivalence is therefore not causal evidence of C operation preservation. |
| 2 | Use after move and move during an active loan are rejected with stable cause chains and smallest repair choices. | ✓ VERIFIED | Real source fixtures pass through `session.Check`; `check.go:211-263` distinguishes moved-place and future-used-loan states. `TestUseAfterMoveDiagnostic`, `TestMoveWhileBorrowedDiagnostic`, and `TestTransferRequiresTake` exercise exact `/1` cause and sorted repair records and prove invalid source produces no execution. |
| 3 | Copy, drop, share, send, and escape abilities derive independently through representative aggregate and generic shapes. | ✓ VERIFIED | `ability_shapes.lang` contains real `Box<Byte>`, `Box<Buffer>`, and `Pair<Byte, Buffer>` functions. `check.go:112-134` derives and serializes typed-core facts. `ability.go:43-56,75-97,101-166` implements sealed field-wise derivation; `TestSourceBoxPairAbilityFacts` verifies all five ordered facts and stable negative witnesses. |
| 4 | A canonical Byte program copies implicitly and leaves its source initialized. | ✓ VERIFIED | `implicit_copy.lang` lowers its bare Byte binding to `OpCopy`; the terminal return still reads the parameter. `TestImplicitByteCopy` checks both core operations and interpreter events. |
| 5 | Untouched Phase 1 programs retain core, diagnostic, execution, and evidence `/0` compatibility. | ✓ VERIFIED | Feature-specific schema tests and frozen Phase 1 diagnostic/evidence goldens pass. The fresh phase gate also verified `testdata/phase1` with five passing lanes and 18 work units. |
| 6 | All 32 private leaf masks and all 1,024 Pair combinations agree with an independent oracle without exposing arbitrary production masks. | ✓ VERIFIED | `ability_test.go:43-91` exhausts the private combiner with direct per-bit boolean expectations; the request-local spy proves sealed Box/Pair derivation invokes it. Production exports expose no mask/combiner authority. |
| 7 | Nested owned type syntax is lossless, canonical, recoverable, and bounded. | ✓ VERIFIED | Parser caps are 64 depth and 4,096 nodes (`parser.go:27-30,204-231`). Round-trip, recovery, exact boundary, and fuzz-seed tests pass, preserving bytes/comments and the diagnostic cap. |
| 8 | The straight-line checker agrees with an independent model on intermediate states and counted linear work. | ✓ VERIFIED | `TestOwnershipSequenceExhaustive` compares operations, loan final uses, snapshots, work, and diagnostic code through all short sequences; `TestOwnershipWorkSeries` passes at 10/100/1,000/10,000 operations. |
| 9 | A source-blind validator independently rejects forged abilities, ambiguous IDs, and illegal move/loan transitions before engines. | ✓ VERIFIED | `corevalidate` imports only inert `core`, independently derives abilities/replays transitions, deep-copies inputs, and gates session/interpreter/C generation. Its mutation matrix and exact `16n+13` scale checks pass. |
| 10 | The coordinated source-to-core lie remains a named expected escape. | ✓ VERIFIED | `corevalidate.KnownEscape` is `escape:coordinated-source-core-lie`; the dedicated test proves an internally consistent false claim is accepted, and Phase 2 protocol output lists it under `expected_escapes`, never controls. |
| 11 | Compile/run stdout and stderr are separately bounded and native stdout is strictly decoded. | ✓ VERIFIED | `native.go` uses four independent max-plus-one writers, never `CombinedOutput`, parses stdout only, rejects successful stderr, and uses strict one-document JSON decoding. Flood and malformed/unknown/trailing/truncated/oversized tests pass. |
| 12 | Versioned evidence binds validated owned source/core/C/execution facts without claiming signing, freshness, or translation proof. | ✓ VERIFIED | `evidence.Build` validates/core-owns the program before hashing; `/1` binds concrete schemas, ordered execution digests, tool/target/flags/policy, `content-identity-only`, and the named escape. Golden and mutation tests pass. |
| 13 | One bounded gate preserves Phase 1 and observes every required Phase 2 control with nonzero work. | ✓ VERIFIED | Fresh `sh scripts/verify-phase2.sh` exited 0; Phase 2 emitted four passing lanes, all seven exact controls, 49 work units, and the separate expected escape. Five 20-sample warm distributions were reported. |
| 14 | Every targeted test command fails closed if a requested test/fuzz target is absent. | ✓ VERIFIED | `assert-go-tests.sh` enumerates tests, exact-matches every requested target, anchors the run regex, and self-tests a guaranteed absent sentinel before known positives. The phase gate exercised this self-test successfully. |

**Score:** 13/14 truths verified (0 present, behavior-unverified)

### Required Artifacts

| Artifact | Expected | Status | Details |
|---|---|---|---|
| `testdata/phase2/*.lang` | Canonical positive and negative source corpus | ✓ VERIFIED | All six fixtures exist, are substantive, and enter real parser/checker/session paths. |
| `internal/compiler/ability/ability.go` | Sealed five-ability structural derivation | ✓ VERIFIED | Substantive, wired from checker, and independently recomputed by core validation. |
| `internal/compiler/check/check.go` | Explicit copy/move/borrow/return lowering and causal rejection | ✓ VERIFIED | Substantive and wired from session; typed facts flow to serialized core. |
| `internal/compiler/corevalidate/corevalidate.go` | Independent fail-closed core admission | ✓ VERIFIED | Substantive and wired before interpreter, C generation, native orchestration, and evidence. |
| `internal/compiler/interp/interp.go` | Independent replay into inert execution facts | ✓ VERIFIED | Validated core drives private runtime state and ordered `/1` events. |
| `internal/compiler/cgen/cgen.go` | Readable C17 whose runtime semantics produce native facts | ⚠️ HOLLOW | It emits real assignments, but the reported outcome/events come from a separate Go replay and a constant JSON string, not the C runtime state. |
| `internal/compiler/native/native.go` | Bounded process boundary and strict stdout decoder | ✓ VERIFIED | Four streams are separately bounded; stdout-only exact decoding is wired into session. |
| `internal/compiler/evidence/evidence.go` | Honest validated content binding | ✓ VERIFIED | Substantive, wired, deterministic, and mutation-tested. |
| `scripts/assert-go-tests.sh` | Exact fail-closed test selection | ✓ VERIFIED | Self-test rejects the nonexistent sentinel and executes exact known targets. |
| `scripts/verify-phase2.sh` | Nonduplicating offline Phase 1+2 gate | ✓ VERIFIED | Builds the shipped CLI once, verifies both corpora, and reports bounded observations. |

### Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| source fixtures | syntax → AST → checker | real session frontend | ✓ WIRED | `session.Check` parses source then calls `check.Program`; source facts are asserted in session tests. |
| checker | ability/core | sealed derivation and explicit operations | ✓ WIRED | Calls `ability.Derive` and materializes ordered `TypeFact` plus `LinearOperation` records. |
| session/interpreter/C/evidence | corevalidate | validation before consumption | ✓ WIRED | Each entry validates and replaces input with a content-owned program. |
| interpreter | execution | private replay emits inert facts | ✓ WIRED | Operations drive runtime map transitions and ordered events. |
| C generator | execution | runtime operations should drive reported facts | ✗ NOT CAUSAL | Operations generate C assignments, but `linearExecution` separately computes the document at generation time and `puts` prints it unchanged. |
| native | execution/session | strict decode then canonical comparison | ✓ WIRED | Decoded stdout is compared across interpreter/O0/O3 on complete canonical records. |
| evidence | corevalidate/core/C/execution | independently admitted content binding | ✓ WIRED | Digests and schemas flow into `/1` evidence and mutation checks. |
| phase gate | shipped CLI | prebuilt binary verifies both corpora | ✓ WIRED | The generic probe missed the literal `go run` pattern, but lines 12-14 build and invoke `cmd/lang` for both corpora; this is the intended nonduplicating implementation. |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|---|---|---|---|---|
| checked core | type/operation facts | canonical source via parser + checker | Yes | ✓ FLOWING |
| interpreter execution | outcome/events | validated core replay over runtime input | Yes | ✓ FLOWING |
| native execution JSON | outcome/events | Go `cgen.linearExecution`, embedded as literal C string | No runtime derivation | ⚠️ STATIC |
| evidence manifest | core/C/execution digests | independently validated, content-owned compiler products | Yes, within stated content-binding boundary | ✓ FLOWING |
| protocol controls | lanes/control IDs/work | actual checker/validator/native/evidence runs | Yes | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|---|---|---|---|
| Full Phase 1+2 gate | `sh scripts/verify-phase2.sh` | Exit 0; full tests/race/vet passed; 7 Phase 2 controls; 49 work units; 20 warm samples for five commands | ✓ PASS |
| C runtime causally determines owned outcome | Compile/run a temporary copy of the owned C golden after injecting `lang_value_delivered.bytes[0] = 9u` | Binary still emitted returned value `01020304` and unchanged transfer/return events | ✗ FAIL |

### Probe Execution

| Probe | Command | Result | Status |
|---|---|---|---|
| `scripts/verify-phase2.sh` | `sh scripts/verify-phase2.sh` | Exit 0 with both corpus JSON results `status: pass` | PASS |

### Requirements Coverage

| Requirement | Source Plans | Description | Status | Evidence |
|---|---|---|---|---|
| OWN-01 | 01, 03, 04, 05, 06 | Noncopyable values transfer exactly once; post-move and live-loan moves reject causally. | ✗ BLOCKED | Source/checker/core/interpreter behavior and causal diagnostics are strong, but the phase's required native preservation is not causally observed: backend state can drift while the precomputed JSON remains green. |
| OWN-02 | 01, 02, 04, 06 | Five abilities derive independently through generic and aggregate types. | ✓ SATISFIED | Real Box/Pair source, ordered core facts, exhaustive independent private-mask oracle, sealed authority, and independent validator recomputation all pass. |

No Phase 2 requirements are orphaned: both `OWN-01` and `OWN-02` appear in plan frontmatter and in the roadmap/requirements traceability tables.

### Prohibition Audit

| Prohibition family | Status | Evidence |
|---|---|---|
| No production arbitrary masks or source-authored positive roots | ✓ VERIFIED | Masks exist only in same-package `_test.go`; production exports and wire structures expose ordered facts only; source grant syntax is rejected. |
| No CFG/exclusive-borrow/resource/allocation/FFI/concurrency scope expansion | ✓ VERIFIED | Production Phase 2 paths remain sealed straight-line ownership with inline C values; generated C contains no allocator/cleanup/alias claims. |
| Checker, validator, and interpreter do not share authorization/transition helpers | ✓ VERIFIED | Each contains a separate implementation and `corevalidate` imports no producer/engine package. |
| No merged native output or stderr semantic parsing | ✓ VERIFIED | `native.go` has independent writers and strict stdout-only decode; `CombinedOutput` appears only in unrelated test harness code. |
| Evidence makes no source-translation/signing/freshness authority claim | ✓ VERIFIED | Manifest labels `content-identity-only` and names the coordinated source/core escape. |
| Phase gate does not nest Phase 1, fuzz actively, use network, or duplicate shared suites | ✓ VERIFIED | Static script contract plus fresh execution confirm one test, race, and vet run followed by two shipped corpus commands. |

### Test Quality Audit

| Test File | Linked Req | Active | Skipped | Circular | Assertion Level | Verdict |
|---|---|---:|---:|---|---|---|
| `internal/compiler/session/session_test.go` | OWN-01/02 | active | 0 | Native equivalence path is circular with respect to C runtime state | Behavioral assertions, but backend causality absent | 🛑 BLOCKER |
| `internal/compiler/ability/ability_test.go` | OWN-02 | active | 0 | No | Exhaustive value + non-tautology spy | PASS |
| `internal/compiler/check/check_test.go` | OWN-01 | active | 0 | No | Independent intermediate-state behavioral oracle | PASS |
| `internal/compiler/corevalidate/corevalidate_test.go` | OWN-01/02 | active | 0 | No | One-boundary mutations + exact scale values | PASS |
| `internal/compiler/native/native_test.go` | OWN-01 | active | 0 | No | Boundary/error value assertions | PASS |
| `internal/compiler/evidence/evidence_test.go` | OWN-01/02 | active | 0 | No for stated content-binding claim | Golden + mutation value assertions | PASS |

**Disabled tests on requirements:** 0  
**Circular patterns detected:** 1 — the native result is generated by Go from the same core before C execution, then compared to interpreter output; backend operations do not determine the emitted facts.  
**Insufficient assertions:** 1 — fake-runner mutations prove comparator behavior, not backend preservation.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|---|---:|---|---|---|
| `internal/compiler/cgen/cgen.go` | 123-127, 174 | Precomputed semantic output embedded as a constant | 🛑 Blocker | Allows actual C runtime value/operation drift to remain invisible to the native differential. |

No unreferenced `TBD`, `FIXME`, or `XXX` markers were found in phase-modified production files. No disabled requirement-linked tests were found.

### Decision Coverage

SKIPPED — Phase 02 has no `CONTEXT.md` or trackable `<decisions>` block.

### Human Verification Required

None. This is an infrastructure/compiler phase, and the remaining gap is deterministically observable with an automated C mutation control.

### Gaps Summary

The frontend ownership and ability work is substantial: source syntax is lossless and bounded; copy versus explicit transfer is represented in typed core; three ownership errors are causal and repairable; five abilities are independently derived and exhaustively checked; invalid core is independently rejected; streams and evidence are strict and bounded; and the complete phase gate is green.

The phase goal is nevertheless not achieved end to end. Native O0/O3 reports are produced by a second Go replay in `cgen.linearExecution` and embedded into C as constant JSON. Since compiled C state does not feed the reported outcome/events, the existing differential and `control:interpreter-o0-o3-owned` can stay green when backend value semantics are corrupted. The fix is to serialize runtime-derived state/events in generated C and add a causal backend mutation that must make the differential fail.

This gap is not deferred to Phase 5: Phase 2's own roadmap success criterion requires equivalent native events, and Plans 02-05/02-06 explicitly claim backend-drift detection now. Phase 5 broadens the corpus and hostile native evidence; it does not erase the current phase contract.

---

_Verified: 2026-09-03T21:53:07Z_  
_Verifier: the agent (gsd-verifier)_
