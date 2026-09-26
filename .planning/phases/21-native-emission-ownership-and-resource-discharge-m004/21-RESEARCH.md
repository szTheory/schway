# Phase 21: Native Emission Ownership and Resource Discharge (M004) — Research

**Researched:** 2026-09-25  
**Domain:** Foreign ABI/resource ownership contracts, native compiler evidence, CI validation  
**Confidence:** HIGH for repository seams and locked boundaries; MEDIUM for external ABI/LTO precedent

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

### Machine-checkable resource-discharge contract
- **D-21-01:** Phase 21 must define machine-checkable acceptance cases and refusal rules for the resource-discharge and foreign-ownership contract. Reserve executable native witnesses and any production emitter admission for a later phase that satisfies the family-specific prerequisites. Contract-structure checks establish completeness and consistency; they do not prove runtime cleanup behavior.
- Keep the single-emission-law direction from Phase 16. Do not introduce a second/coexisting ownership or discharge law merely to support an emitter family.

### Foreign-boundary exit and cleanup classification
- **D-21-02:** The checked contract must explicitly classify modeled foreign exits, including normal return, error return, supported unwind, nonlocal transfer, defects/cancellation, and process termination. Require an explicit, checked discharge rule for each exit that is admitted; refuse any path whose cleanup behavior is opaque or unproved.
- Classification does not itself introduce new Lang exception, cancellation, or unwind semantics. State which classes are supported, refused, or outside cleanup guarantees, and require family-specific evidence before any emitter is reopened.

### D-14-45 one-TU / `-flto` evidence scope
- **D-21-03:** Preserve the current one-TU/no-`restrict`/foreign-refusal boundary, add one scoped compiler comparison of an emitted multi-function fixture to substantiate D-14-45, and retain a cheap structural guard as recurring CI evidence. Broader recurring host/toolchain matrices require demonstrated regression value.
- Reuse the existing semantic comparator and its independent controls; do not create a second comparator. Distinguish behavioral equivalence under `-O3` and `-O3 -flto` from evidence about optimization activity or performance. Do not infer that one-TU LTO performs no optimizations, and do not imply that this work proves future foreign or by-pointer emitter behavior.
- Keep the measured claim bounded to the tested fixture, compiler/toolchain, and host lanes. Any widened claim needs corresponding measured evidence.

### the agent's Discretion
The researcher and planner may choose the contract encoding, structural assertions, fixture details, and CI lane placement if they preserve the decisions above, the existing refusal boundary, and honest evidence grades. They may not use the design artifact as authorization to route a cut emitter into production.

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope.
</user_constraints>

## Summary

Phase 21 should produce an explicit, machine-checkable contract artifact for resource discharge and foreign exits, plus structural checks that prove the contract is complete and fail closed. Those checks must be described as contract consistency evidence, not as proof of native cleanup behavior. No cut emitter family becomes admitted in this phase.

The repository already has one whole-program C emission law, refuses foreign and by-pointer shapes in `emitProgram`, derives `live_resources` for admitted shapes, and has a resource-ledger precedent only in legacy code. The existing test named `TestLTOInertnessOnMultiFunctionEmission` verifies that a multi-function fixture emits schema-2 output, but does not invoke Clang or compare execution under `-O3` and `-O3 -flto`. Add that one bounded emitted-fixture comparison using the existing comparator and independent controls. Let the regular structural tests recur in the existing macOS/Linux `go test ./...` checks; do not add an extra host/toolchain matrix without measured need.

**Primary recommendation:** define the contract as data with a strict validator and mutation-killed completeness/refusal checks; keep native cleanup witnesses behind later family-specific admission work; separately run and preserve one scoped emitted-fixture semantic comparison with exact toolchain/host provenance.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Exit classification and discharge obligations | Planning contract / compiler semantic policy | cgen admission | Policy must be explicit before an emitter can claim support. |
| Contract structure and refusal consistency | Go tests in compiler/session or cgen test boundary | CI aggregate | Existing test suite owns deterministic structural regression checks. |
| Emitted fixture LTO comparison | Session/native integration evidence | Clang toolchain | Must exercise actual emitted C through the existing semantic comparator. |
| Foreign/by-pointer family admission | Later family-specific implementation phase | cgen | Phase 21 does not reopen the families or create a parallel ownership law. |

## Standard Stack

| Concern | Existing stack/pattern | Use in Phase 21 |
|---------|-----------------------|-----------------|
| Implementation and tests | Go 1.24, standard library, `go test` | No new dependencies. |
| Native compiler | Installed Clang, C17 emission | Record exact compiler and host for the scoped comparison. |
| Program behavior comparator | Existing schema-2 semantic comparator and independent controls | Reuse; do not fork or duplicate it. |
| CI | `.github/workflows/ci.yml`, full tests on `ubuntu-latest` and `macos-latest` | Put only the cheap structural guard in recurring CI; existing full test lanes are already recurring. |

## Architecture Patterns

### Contract data plus independent structural guard

Represent each foreign-exit class explicitly, with a disposition and discharge-evidence requirement. The validator should reject missing/unknown classes, unsupported classifications, admitted paths with absent cleanup proof, and any attempt to infer emitter admission from classification alone. Ensure the test exercises a mutated contract copy with a missing case or evidence field so it demonstrates that the guard is not inert. Keep the accepted production emission law singular.

### Separate contract evidence from execution evidence

Give each evidence claim a ceiling: contract completeness/refusal is structural; the LTO fixture comparison is execution evidence for one fixture/compiler/host configuration; neither establishes runtime cleanup for cut foreign/by-pointer emitters. The Rustonomicon documents ABI-specific unwind permissions and failure behavior; that is precedent for explicit ABI boundaries, not a source of Codename Lang semantics ([Rustonomicon FFI and unwinding](https://doc.rust-lang.org/nomicon/ffi.html)).

### Bounded compiler comparison

Use the emitted multi-function fixture, compile/execute under `-O3` and `-O3 -flto`, compare semantic observations using the existing comparator, and retain an independent control that proves the comparison path can detect divergence. Capture Clang version, host, fixture identity/hash, flags, and outcome. Clang describes LTO as whole-program/cross-module optimization capability, but that documentation does not predict optimizer activity in this one-translation-unit fixture ([Clang ThinLTO](https://clang.llvm.org/docs/ThinLTO.html), [Clang `-flto` options](https://clang.llvm.org/docs/CommandGuide/clang.html)).

### Test layering and CI economics

LLVM distinguishes focused regression tests from whole-program compile/run tests and recommends small reproducible cases; its test guide describes regression tests as routine checks and whole-program tests as compile-and-execute evidence ([LLVM Testing Infrastructure Guide](https://www.llvm.org/docs/TestingGuide.html)). Apply that separation here: cheap contract structure checks recur in current CI, while the specifically scoped compiler comparison is run once and its scope is recorded. Current CI already runs `go test ./...` on macOS and Linux; adding a second job that repeats the same comparator would need evidence of added regression value.

### Residual legacy-emitter convergence

Both public `cgen.Emit` and `cgen.EmitNative` validate the program and route to `emitProgram` (`internal/compiler/cgen/cgen.go`). The three retained foreign/by-pointer implementation functions have no production call sites in the current tree; the whole-program path explicitly refuses the corresponding foreign shapes (`internal/compiler/cgen/cgen_program.go`). This creates a bounded convergence option for D-11-02/D-12-36: confirm the absence of production/test API consumers, remove only unreachable private lowering code and helpers, and keep the current refusal behavior plus public manifest/header/conformance APIs. Do not port these families into `emitProgram`. If any live consumer is found, preserve its behavior and report the exact consumer rather than widening admission.

## Don't Hand-Roll

- A second program semantic comparator or a new emitted-program execution protocol.
- A new Lang unwind, cancellation, exception, or cleanup semantic to make an emitter pass.
- A production resource ledger for currently refused families as part of a contract-design phase.
- An optimizer-activity or benchmark claim based only on equal program output.
- A larger recurring OS/compiler matrix before the scoped comparison demonstrates why it is needed.

## Common Pitfalls

1. Treating a documented foreign exit class as supported without an explicit discharge obligation and family-specific evidence.
2. Treating structural schema completeness as evidence that generated native code cleans resources at runtime.
3. Reading the current multi-function LTO witness name as proof that Clang compiled the emitted fixture. Source inspection shows it currently only checks clean checking, multiple functions, successful emission, and schema-2 output (`internal/compiler/session/witness_registry_test.go`, `TestLTOInertnessOnMultiFunctionEmission`).
4. Concluding that `-flto` performs no optimizations from behavioral equality. Equality answers semantic behavior for this fixture; it does not measure optimization activity or performance.
5. Allowing this design artifact to reopen `emitLinearForeign`, `emitLinearBorrowedByPointer`, or `emitLinearBorrowedByPointerPlain`. `emitProgram` currently refuses foreign-call bodies and foreign contracts in multi-function programs (`internal/compiler/cgen/cgen_program.go`).
6. Deleting shared foreign manifest/header/conformance helpers merely because executable legacy bodies are unreachable; preserve their separately owned public functionality.
7. Creating a duplicate CI lane: current `.github/workflows/ci.yml` runs the full Go suite on both initial host priorities, so a recurring structural test is already covered there.

## Code Examples / Repository Anchors

- `internal/compiler/cgen/cgen_program.go`: `deriveProgramLiveResources`, `emitProgram`, and its foreign/by-pointer refusal checks.
- `internal/compiler/cgen/cgen_program_test.go`: live-resource derivation, refusal-order, and whole-program emission tests.
- `internal/compiler/session/witness_registry_test.go`: `TestLTOInertnessOnMultiFunctionEmission`, currently a structural emission witness rather than an executed compiler comparison.
- `internal/compiler/native/native_lto_test.go`: existing compiler flag and LTO test patterns; distinguish its foreign-TU/composition-control evidence from the emitted multi-function fixture.
- `.github/workflows/ci.yml`: macOS/Linux full Go suite and current evidence aggregate.

## Validation Architecture

| Layer | Purpose | Candidate command | Recurring? |
|-------|---------|-------------------|------------|
| Focused contract validator | Reject malformed, incomplete, or unsupported exit/discharge cases; include a mutation control | `go test ./internal/compiler/session -run 'TestPhase21.*(Contract|Discharge)' -count=1` | Yes, through existing full-suite CI on macOS/Linux |
| Scoped emitted fixture comparison | Execute one real multi-function emitted fixture at `-O3` and `-O3 -flto`; reuse comparator and controls; record provenance | A named focused test or phase-specific command finalized by the planner after confirming the existing harness | One scoped evidence run, not a new recurring matrix |
| Refusal regression | Keep cut foreign/by-pointer families refused through the existing whole-program admission boundary | `go test ./internal/compiler/cgen -run 'TestProgramBorrowedByPointerDisposition|TestUnsupportedProgramShapePrecedesSchema2Preflight' -count=1` | Yes, via full-suite CI |
| Full phase preflight | Ensure all compiler/session tests remain green | `go test ./... -count=1` | Existing CI runs full suite; no duplicate job required |

The named focused test pattern is a planning target and must be resolved to the actual Phase 21 test names when implementation creates them. No tests were run during planning.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | Repository tests and compiler code | Yes | `go1.24.0 darwin/arm64` | None required |
| Clang | Native fixture compilation | Yes | Apple clang 21.0.1 | CI explicitly requires `clang --version`; Linux lane uses installed runner Clang |
| macOS host | Current local scoped evidence lane | Yes | Darwin arm64 | Linux evidence remains a distinct claim and requires a Linux run if asserted |
| Linux host | Project portability evidence | Not this local host | — | Existing CI Linux runner |

## Project Constraints (from AGENTS.md)

- Start repository edits through a GSD command and keep planning artifacts synchronized.
- Use `$gsd-plan-phase` for this planned phase and preserve existing project patterns.

## Research Questions — Resolved

- **Contract representation and independent check:** Use a versioned JSON artifact at `21-RESOURCE-DISCHARGE-CONTRACT.json`; the session test independently enumerates the required exit IDs, allowed disposition states, and D-16 family IDs, then mutation-tests absent cases/evidence. The artifact remains design-only and no production code consumes it.
- **Existing one-shot compiler harness:** Reuse `phase16CompareDirectProgramFourTiers` from `internal/compiler/session/session_phase11_differential_test.go`. Add a build-tagged Phase 21 test to invoke the existing emitted-C path, runner, and comparator, then record its exact receipt; do not add a second comparator or recurring matrix.
- **Host and toolchain scope:** This research session is macOS arm64 with Apple Clang 21.0.1. It makes no Linux compiler claim. The planned one-shot run records only its actual host/toolchain; existing Linux CI continues to run the cheap untagged structural tests.
