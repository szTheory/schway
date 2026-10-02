# Phase 23: Live Local Allocation and Discharge - Research

**Researched:** 2026-09-27
**Domain:** Checked foreign operations, local ownership, C17 native application IO
**Confidence:** MEDIUM

<user_constraints>
## User Constraints (from CONTEXT.md)

## Phase Boundary

Deliver a public native application that reads one caller-selected file byte
through an explicitly linked C adapter. The adapter returns a real malloc-backed
allocation that remains live in Lang; a separate contracted use operation
reads the byte, and Lang's generated cleanup calls its paired infallible
destructor. The runnable witness uses files containing `0x41` and `0x42` and
reports decimal `65` and `66`. Failed acquisition creates no Lang owner, and a
real typed use failure after successful acquisition proves error-path cleanup.

This phase owns local acquisition, use, release, and their independent evidence.
It does not admit ownership transfer across Lang calls, general strings or byte
arrays, loops, owning aggregates, pointer-helper families, unwind, cancellation,
or cleanup guarantees after defects and process termination. The adapter owns
and closes its file descriptor; Lang owns only the returned allocation.

## Implementation Decisions

### Application input and output

- **D-23-01:** Pass exactly one caller-selected path token of at most 4096 bytes
  through the public app-run route, into the source entry, and then to the
  explicit acquire operation. Keep the value opaque outside this use; add no
  general string, path-manipulation, or array API. Preserve Phase 22's existing
  U64 application route. The exact source spelling and CLI flag are technical
  choices, but they must preserve this narrow explicit-input boundary.
- **D-23-02:** Accept exactly one raw file byte. Empty input and files longer
  than one byte return distinct typed acquisition errors. Return the actual
  byte value as decimal U64 through the existing result/output path (`0x41` →
  `65`, `0x42` → `66`); do not add a general text-output runtime for this
  witness.

### Resource and failure behavior

- **D-23-03:** Give acquire, borrowed use, and infallible consuming release
  distinct checked per-operation contracts. A successful acquire creates one
  noncopyable local owner. A failed or partial acquire creates no Lang owner;
  the adapter frees any partial allocation before returning failure. Use reads
  the real buffer and returns its value for `0x41`/`0x42`; `0x43` is the
  documented unsupported-byte case and returns a typed use error after
  acquisition. On that error, generated cleanup frees the still-owned
  allocation exactly once before the application reports failure. Acquisition
  failures and use failures use the existing ordinary application failure
  channel with a nonzero outcome and bounded diagnostic. No transfer, unwind,
  cancellation, or fallible destructor is admitted here.
- **D-23-04:** Publish a successor to Phase 21's contract-only artifact. State
  explicitly that release consumes the obligation, borrow preserves it, and
  transfer preserves it under a new owner; this phase executes local release
  but does not implement transfer. Do not amend the archived contract or treat
  its structural checks as runtime cleanup proof.

### Verification and CI

- **D-23-05:** Replace subjective phase UAT with objective automated checks.
  Cover source admission, independent peer validation, interpreter/model
  behavior, generated-C/build behavior, public app execution, two different
  files, failed acquisition, post-acquisition use failure, and reached negative
  controls. A native observer independent of compiler events must establish
  real allocation, use-after-acquire, and destruction before successful exit.
  Omitted, premature, duplicate, and wrong-resource destructor controls must
  fail even when compiler events remain plausible.
- **D-23-06:** Run exact-shape native evidence on macOS and Linux. Put stable
  recurring checks in the existing CI host lanes when their regression value
  justifies their runtime and maintenance cost; assign each expensive full or
  sanitizer lane one owner per host. Report any unavailable host or lane as
  incomplete, never as a pass. Do not create duplicate full-suite CI work with
  no distinct evidence question.

### The agent's Discretion

- Choose the smallest source-level opaque path representation and CLI spelling
  that satisfy D-23-01 without adding generic text semantics.
- Choose bounded foreign ABI records and the independent native observer
  design; retain explicit path, initialization, length, and allocator pairing
  checks for macOS and Linux.
- Choose where focused tests live and how expensive evidence lanes are split
  across existing CI jobs. Keep source inspections, newly executed checks, and
  historical receipts distinct.
- Keep the `0x43` unsupported-byte case a real, documented use-operation error
  reached through the public application entry; do not replace it with a direct
  cleanup-block call, a synthetic event, or a process crash.

## Deferred Ideas

- Ownership transfer through calls and typed errors belongs to Phase 24.
- Shared/exclusive read-copy pointer families and integrated utility closure
  belong to Phase 25.
- General strings, arrays, loops, arithmetic, owning aggregates, fallible
  destructors, unwind, cancellation, and broader FFI remain outside M004 scope.

</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| FFI-03 | Each admitted foreign operation resolves its own checked signature, operand modes, acquisition/failure behavior, and release pairing; distinct acquire/use/release operations can consume admitted local values without inheriting the function's first foreign symbol contract. [VERIFIED: .planning/REQUIREMENTS.md:27] | Lower foreign facts onto each core operation and make checker, core peer, origin peer, path oracle, interpreter, ABI conformance, and C emission consume that operation's own contract. |
| RES-04 | A source-constructible opaque noncopyable resource can receive a real bounded malloc-backed buffer from an explicitly linked C adapter, remain live after the adapter returns, and supply a byte determined by a caller-selected file through Lang-directed use. [VERIFIED: .planning/REQUIREMENTS.md:31] | Add one entry-only path token and one opaque owner shape; link a narrow adapter; demonstrate use while live and generated release. |
| RES-07 | A failed foreign acquisition creates no Lang-owned resource; the bounded adapter contract specifies maximum size, empty input, initialization/length, failure representation, and cleanup of its own partial acquisition before exposing a result. [VERIFIED: .planning/REQUIREMENTS.md:34] | Model success and each acquisition error separately; specify pointer/length initialization; free adapter-owned partial storage on all failures. |
| RES-08 | Discarding an owning acquisition cannot erase its obligation: the admitted form either performs immediate consuming cleanup or is rejected before C serialization. [VERIFIED: .planning/REQUIREMENTS.md:35] | Prefer a pre-serialization structured refusal for discarded owning acquisition in this narrow phase. |
| RES-09 | Independent validation derives obligations from successful acquisitions and follows each admitted path; missing, duplicate, wrong-resource, or fabricated cleanup is rejected without relying on the presence of an existing release operation to discover the obligation. [VERIFIED: .planning/REQUIREMENTS.md:36] | Re-seed peer obligation analysis from acquire operations and mutate away every release, plus duplicate, premature, wrong-owner, and fabricated cases. |
</phase_requirements>

## Summary

The Phase 22 route passes one bounded opaque argv token to one retained native executable, but its public entry remains U64-to-U64 and the emitter refuses foreign-call/by-pointer bodies. The frozen foreign-resource shim frees inside C before returning, so it does not prove a live Lang owner. [VERIFIED: cmd/lang/main.go:183-237; internal/compiler/session/session.go:1140-1166; internal/compiler/cgen/cgen_program.go:583-611; native/lang_foreign_resource.c:49-60]

Extend the current parser → checker → core → independent peers → interpreter/C emitter → retained-app path for exactly one opaque path token, one noncopyable owner, one borrowed use, and one infallible release. Seed every peer's obligation analysis from successful acquisitions, and use an independent native observer for actual allocation/use/free. Preserve the U64 app route and frozen fixture. [VERIFIED: .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md:9-21,30-59; internal/compiler/corevalidate/corevalidate.go:3216-3259]

**Primary recommendation:** Add the narrow source-to-native resource spine and refuse all neighboring shapes before C serialization. Do not treat interpreter events as physical cleanup evidence. [VERIFIED: internal/compiler/cgen/cgen_program.go:583-611; internal/compiler/interp/interp.go:958-985]

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Opaque path-token input | Browser / Client (CLI caller) | API / Backend (native app entry) | The CLI transports one token as argv; generated main admits it and passes it to the source entry without general string operations. [VERIFIED: cmd/lang/main.go:183-237; internal/compiler/native/native_app.go:292-299,379-416] |
| File open/read and descriptor close | API / Backend (explicit C adapter) | — | The context assigns fd ownership and close to the adapter; the adapter alone performs bounded host IO. [VERIFIED: .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md:17-21] |
| Buffer allocation lifetime | API / Backend (Lang checked core and generated C) | — | Successful acquisition creates the Lang obligation; generated C calls the paired destructor, while libc provides process-local storage. [VERIFIED: .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md:9-21,44-54; internal/compiler/core/core.go:835-868] |
| Ownership and borrow guarantees | API / Backend (compiler peers) | — | check, corevalidate, originvalidate, and pathoracle independently derive operand modes, owner identity, and path discharge; no peer should trust check's release list. [VERIFIED: .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md:69-76; internal/compiler/corevalidate/corevalidate.go:3216-3259; internal/compiler/originvalidate/originvalidate.go:321-338; internal/compiler/pathoracle/pathoracle.go:230-268] |
| Native result / error reporting | API / Backend (generated app shell) | Browser / Client (CLI caller) | Generated main writes the existing decimal U64 result or bounded failure diagnostic and returns the ordinary process outcome. [VERIFIED: internal/compiler/cgen/cgen_program.go:1001-1006; cmd/lang/main.go:231-245] |
| Physical lifetime evidence | API / Backend (independent native observer) | — | A separate native observation channel must test actual pointer lifecycle independently of compiler event records. [VERIFIED: .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md:63-76; .planning/REQUIREMENTS.md:52-53] |

## Project Constraints (from AGENTS.md)

- Begin file changes through a GSD workflow. The Phase 23 GSD phase operation was run before this artifact was written. [VERIFIED: AGENTS.md:77-87]
- Re-read PRODUCT-ROADMAP and LANGUAGE-MATURITY before planning, compare them with source/refusal/evidence witnesses, and surface the three next capabilities. The 2026-09-27 documents still align with inspected source: Phase 23 file-byte ownership, Phase 24 call/error transfer, Phase 25 bounded pointer families. No living-document update is indicated because these source boundaries have not changed. [VERIFIED: AGENTS.md:98-122; .planning/PRODUCT-ROADMAP.md:150-192; .planning/LANGUAGE-MATURITY.md:57-81]
- Keep one runnable gain in a feature phase, scope-review a second consecutive enabling phase, and tie independent validation and negative controls to the new capability. Do not add completeness percentages, a backend, or a generalized planning framework without a concrete consumer. [VERIFIED: AGENTS.md:113-118]
- Growing plan counts need a named user witness or safety obligation; preserve archived decisions unless a dated amendment is explicit; distinguish source inspection, new checks, and historical receipts. The managed developer profile is not a phase artifact and must not be edited as a shortcut. [VERIFIED: AGENTS.md:90-111,113-118]
- Keep active milestone requirements and phase ownership in REQUIREMENTS.md and ROADMAP.md; this procedure runs at workflow transitions, not as a background scheduler. [VERIFIED: AGENTS.md:120-122]
- Preserve the project stack and safety posture. AGENTS.md says “Go 1.24 standard library hosts Stage 0” and “readable C17 emitted to installed Clang is the development native path”; it also requires defined safe behavior, explicit ownership/cleanup/FFI, shallow audited dependencies, formatter-owned source, separate evidence cost lanes, macOS/Linux portability, and explicit untrusted-input/build authority. [VERIFIED: AGENTS.md:23-44]

## Standard Stack

### Core

| Component | Version | Purpose | Why Standard |
|-----------|---------|---------|--------------|
| Go standard library | 1.24; go.mod says “go 1.24” | Stage 0 compiler, process runner, tests, manifest hashing | Existing bootstrap stack; no new Go dependency is needed. [VERIFIED: go.mod:3; AGENTS.md:32-35] |
| C compiler | Installed Clang; local probe reported Apple clang 21.0.0 on arm64 macOS | Compile and link generated C plus declared local adapter sources | The runner captures compiler, target, flags, and host ABI in build identity; CI already requires Clang on both host lanes. [VERIFIED: local tool probe 2026-09-27; internal/compiler/native/native_app.go:202-224; .github/workflows/ci.yml:35-78] |
| C language mode | C17; runner flag is "-std=c17" | Generated application and adapter ABI | Keep the current strict application compiler mode and warnings-as-errors. [VERIFIED: internal/compiler/native/native_app.go:114] |
| POSIX file and allocation calls | POSIX.1-2024 | Narrow adapter operations: open, bounded read, close, malloc, free | This is the shallow native boundary required; POSIX read may return short counts, and freeing a non-live/non-malloc pointer is undefined behavior. [CITED: pubs.opengroup.org/onlinepubs/9799919799/functions/read.html; pubs.opengroup.org/onlinepubs/9799919799/functions/free.html] |
| Native interpreter | Repository implementation | Deterministic model, not actual filesystem or physical cleanup proof | Keep it as semantic model for fixed expected outcomes; actual IO/free is a separate evidence question. [VERIFIED: internal/compiler/interp/interp.go:962-985; internal/compiler/session/session_app_verify.go:241-267] |

### Supporting

| Component | Version | Purpose | When to Use |
|-----------|---------|---------|-------------|
| Explicit local-C binding manifest | Existing schema "lang.local-c/1" | Declare adapter source/header/symbol/type and build inputs | Use the existing manifest and resolver; extend operation-to-manifest ABI conformance rather than adding fixture-symbol lookup. [VERIFIED: internal/compiler/native/bindings.go:21-45,222-270] |
| Go test / native integration harness | go test from Go 1.24 | Fast semantic mutations and public native app evidence | Use existing package tests, subprocess helpers, and native app tests; keep host evidence in a focused command. [VERIFIED: go.mod:3; cmd/lang/main_test.go; internal/compiler/native/native_app_test.go; .github/workflows/ci.yml:54-79] |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| One bounded opaque argv path token | General Lang strings, arrays, and path functions | Generic text adds encoding, indexing, and ownership semantics not needed by this witness and explicitly deferred. [VERIFIED: .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md:30-40,91-92] |
| Opaque pointer-bearing owner representation with checked lifetime | Cast pointer to U64 and treat it as an integer handle | The frozen fixture casts a pointer through uintptr_t into uint64_t, baking an integer-width assumption into the wire value. Keep a pointer-typed ABI field and target C ABI rules. [VERIFIED: native/lang_foreign_resource.c:20-31; AGENTS.md:41-44] |
| Narrow adapter plus matching consuming C destructor | Implement file IO or allocator logic in generated shell code | Keeps fd ownership in C and resource ownership in Lang, as locked by context; keeps generated shell focused on argv, result, and error channel. [VERIFIED: .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md:9-21] |
| Acquirer-seeded independent resource analysis | Infer ownership from present release operations | The current peer's tracked set is seeded from releases, so deleting all releases can hide the acquisition obligation. [VERIFIED: internal/compiler/corevalidate/corevalidate.go:3216-3259] |
| Separate native observer | Treat interpreter state or compiler release events as physical proof | Those channels can confirm modeled intent but cannot establish that a real allocated pointer was used and freed. [VERIFIED: internal/compiler/interp/interp.go:958-985; .planning/REQUIREMENTS.md:52] |

**Installation:** None. No third-party package or runtime is recommended. The local build uses the repository's Go/C/Clang path and standard OS/libc calls. [VERIFIED: AGENTS.md:32-35; go.mod:3]

## Architecture Patterns

### System Architecture Diagram

~~~mermaid
flowchart TD
    A[lang app run: one bounded argv path token] --> B[generated native main validates token]
    B --> C[Lang entry: opaque path token]
    C --> D[acquire operation / explicit C adapter]
    D -->|open/read failure, empty, or oversized| E[typed acquire error; no Lang owner]
    D -->|one initialized byte allocated| F[one noncopyable local owner]
    F --> G[borrowed use reads the actual buffer]
    G -->|supported byte| H[infallible consuming release]
    H --> I[existing decimal U64 result writer]
    G -->|typed use error| H
    H --> J[bounded diagnostic and nonzero outcome]
    E --> J
    D -. adapter closes fd before returning .-> K[OS file descriptor]
    F -. actual allocation .-> L[libc malloc storage]
    H -. paired destructor calls free .-> L
    M[independent native observer] -. pointer lifecycle, not compiler events .-> L
~~~

The command already passes one opaque argument after the separator to one retained executable. The runner declares “MaxApplicationArgumentBytes = 4096”; preserve this route and the existing U64 application shape while adding the path-token entry form. The new source and ABI representations remain technical decisions, not current compiler capability. [VERIFIED: cmd/lang/main.go:183-237; internal/compiler/native/native_app.go:20-25,292-299,379-382; .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md:30-40]

### Recommended Project Structure

Keep production changes in the current pipeline:

- Source declaration/body: internal/compiler/ast/ast.go and internal/compiler/syntax/parser.go. [VERIFIED: internal/compiler/ast/ast.go:24-45,171-195; internal/compiler/syntax/parser.go:215-255,390-435]
- Admission, lowering, and diagnostics: internal/compiler/check/check.go. [VERIFIED: internal/compiler/check/check.go:3641-3684,3966-3975]
- Per-operation core facts: internal/compiler/core/core.go. [VERIFIED: internal/compiler/core/core.go:152-201,627-649,835-868]
- Independent derivation: internal/compiler/corevalidate/corevalidate.go, internal/compiler/originvalidate/originvalidate.go, and internal/compiler/pathoracle/pathoracle.go. [VERIFIED: internal/compiler/corevalidate/corevalidate.go:202-217,3216-3259; internal/compiler/originvalidate/originvalidate.go:321-338,555-571; internal/compiler/pathoracle/pathoracle.go:230-268]
- Model and sole C serialization authority: internal/compiler/interp/interp.go and internal/compiler/cgen/cgen_program.go; public emitter entry is cgen.EmitApplication. [VERIFIED: internal/compiler/cgen/cgen.go:132-138; internal/compiler/interp/interp.go:166,958-985]
- Public build/run and binding: internal/compiler/session/session.go, internal/compiler/native/native_app.go, internal/compiler/native/bindings.go, and cmd/lang/main.go. [VERIFIED: internal/compiler/session/session.go:1142-1166; internal/compiler/native/native_app.go:120-145,292-299; internal/compiler/native/bindings.go:30-45; cmd/lang/main.go:183-237]
- Add the utility, adapter/header, observer, and README alongside the Phase 22 example. New Phase 23 example/test paths are proposals, not existing files. [ASSUMED]

### Component Responsibilities

| File / subsystem | Current responsibility and exact seam | Planning consequence |
|------------------|----------------------------------------|----------------------|
| internal/compiler/check/check.go | checkFallibleLinear dispatches only a single immediately returned fallible call or a sequence of try/discard calls; resolveForeignStep only accepts the function's parameter; checkResourceLifecycle attaches the first symbol contract to the function. [VERIFIED: internal/compiler/check/check.go:3641-3684,3966-3975,4433-4461,4592-4602] | Replace the sample-specific sequence with the narrow acquire → bind owner → borrowed use → consuming release shape. Resolve each call's source place and operation-specific C contract. Refuse discarded owner values before serialization. |
| internal/compiler/core/core.go | ForeignContract is function-scoped. Current operation values include “OpForeignCall OperationKind = "foreign_call"” and “OpRelease OperationKind = "release"”; LinearOperation carries ReleasesOperationID and Allocator. [VERIFIED: internal/compiler/core/core.go:152-201,627-649,835-868] | Add per-operation declaration facts or an operation-keyed contract structure. Keep acquire/use contracts distinct; let the consuming release carry its destructor pairing and infallible fact. Update JSON, cloning, validation, and constructors together. |
| internal/compiler/corevalidate/corevalidate.go | Validate is source-blind; foreign checks read function.ForeignContract, and checkReleaseOrder discovers tracked acquisitions from release operations already present. [VERIFIED: internal/compiler/corevalidate/corevalidate.go:202-217,1230-1314,3216-3259] | Independently rederive ABI/operand mode and acquire-success obligations. Prove each terminal path discharges every local acquisition once. Add a control that deletes every release; an empty release list must not imply there was no owner. |
| internal/compiler/originvalidate/originvalidate.go | ValidatePublished recomputes origins; OpForeignCall borrow/retain origin currently consults function.ForeignContract.Alias. [VERIFIED: internal/compiler/originvalidate/originvalidate.go:321-338,555-571] | Derive borrowed-use source and result origin from the operation's declared mode/signature, not the function's first foreign symbol. |
| internal/compiler/pathoracle/pathoracle.go | EnumeratePaths performs bounded path expansion and refuses cycles; it declares “const MaxPaths = 4096”. [VERIFIED: internal/compiler/pathoracle/pathoracle.go:63-67,230-268] | Extend independent path replay with owner state at each operation and terminal. Keep this phase loop-free and local; refuse cycles and unsupported ownership shapes. |
| internal/compiler/interp/interp.go | Run validates core before execution; current OpForeignCall takes a modeled success path and OpRelease emits a semantic release event. [VERIFIED: internal/compiler/interp/interp.go:166,958-985] | Model acquisition/use outcomes by operation identity: failed acquire mints no owner; use error after success retains the owner until local release. Label results as model evidence only. |
| internal/compiler/session/session_app_verify.go | Source replay requires empty foreign_outcomes and refuses foreign contracts/local C; verifier-model mode is separate. [VERIFIED: internal/compiler/session/session_app_verify.go:235-267,360-402] | Keep differential replay independent of host IO. Add only a closed deterministic ownership fixture if needed; app verify must not execute adapter C or claim actual free. |
| internal/compiler/cgen/cgen_program.go | emitProgram is the whole-program emitter. It currently gates app entry to linear U64-to-U64 and refuses foreign-call and by-pointer bodies. [VERIFIED: internal/compiler/cgen/cgen_program.go:575-611] | Preserve it as sole serializer. Widen only the exact entry/resource/use/release shape; refuse malformed or unsupported shapes before serialization. Emit cleanup before reporting a use error and keep existing U64 result writing. |
| internal/compiler/native/native_app.go | BuildApplication compiles generated C and resolved manifest sources; RunApplication verifies artifact receipt/digest and invokes one child with the token as argv. Build identity includes compiler/target/flags/host ABI. [VERIFIED: internal/compiler/native/native_app.go:120-145,202-224,292-299,379-416] | Reuse the retained app runner. Bind adapter source/header through the closed manifest and include those inputs in receipts. Preserve direct argv; do not interpolate path text into shell commands. |
| internal/compiler/native/bindings.go | BindingManifest names sources, headers, include dirs, symbols, and function typedefs; its resolver confines resolved inputs to manifest root and snapshots file bytes. [VERIFIED: internal/compiler/native/bindings.go:30-45,222-270] | Extend type conformance so each operation's checked C signature matches its declared header typedef; manifest membership alone is build authority, not semantic proof. |
| internal/compiler/native/foreign_resource.go and native/lang_foreign_resource.c | The historical helper returns the repository fixture path; its source contains the exact expression “return filepath.Join(root, "native", "lang_foreign_resource.c")”, and that fixture frees before returning. [VERIFIED: internal/compiler/native/foreign_resource.go:18-22; native/lang_foreign_resource.c:49-60] | Preserve this frozen tracer and do not use repo-location lookup for the new utility. Bind new C files through the explicit user-declared local manifest. |
| cmd/lang/main.go and examples/phase22/README.md | The CLI usage string is “usage: lang app run ARTIFACT [--report REPORT] [--evidence=events] -- INPUT”; it accepts one token after -- and calls RunApplication once. [VERIFIED: cmd/lang/main.go:183-237] | Document the file-byte utility and command while preserving the one-run routing and U64 app route. |

### Operation and ABI Pattern

Use three distinct semantic operations and attach each checked signature to the operation that consumes it:

| Operation | Operand | Success/error fact | Ownership transition |
|-----------|---------|--------------------|----------------------|
| Acquire | Opaque path token | Success returns a fully initialized pointer-bearing owner record with length one; empty, oversized, and IO/allocation failures are typed acquisition errors. | Only successful acquisition mints a fresh local owner identity. Failure returns no owner. |
| Borrowed use | Owner borrowed for this operation | Reads the real allocation, returns the supported byte, or returns typed unsupported-byte use error. | Borrow ends at operation completion; the owner's obligation remains live on both result branches. |
| Consuming release | Owner | Declared infallible destructor; exact allocator/destructor association and ABI are checked. | Consumes the owner once on each terminal path; generated C calls paired free exactly once. |

The dated Phase 23 resolutions in `## Open Questions` freeze source spelling, type names, error variants, and the C17 record layout. The earlier recommendation to preserve a pointer rather than serialize an address through U64 remains binding for the selected approach. [RESOLVED: 2026-09-27]

For the adapter, bound reads to the desired byte plus a second-byte probe. POSIX read permits short results, so loop/retry interrupted reads and distinguish EOF after zero bytes, EOF after one byte, and a second byte. Allocate one byte before reading; initialize only after receiving the first byte; on empty, oversized, or read failure, free that partial allocation and close the descriptor before returning typed failure. On success, close the descriptor and return the still-live initialized allocation with length one. This makes partial-allocation cleanup observable and keeps fd ownership in C. [CITED: pubs.opengroup.org/onlinepubs/9799919799/functions/read.html; VERIFIED: .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md:17-21,44-54]

The observer should be a separately compiled native test translation unit. One portable strategy is a test-only allocator/destructor shim that delegates to real malloc/free, records returned and passed pointers, and writes a private lifecycle receipt from an atexit finalizer before process exit. Negative controls mutate the generated public executable call path but leave compiler events unchanged. The observer must flag missing, early, duplicate, or unrelated-pointer release before dereferencing or calling libc free a second time. Avoid GNU-only linker wrapping unless it is demonstrated on both hosts. [ASSUMED]

Do not use Buffer as the owner. The current source condition is “value.Constructor == "Byte" || value.Constructor == "Buffer" || (value.Constructor == "U64" && len(value.Arguments) == 0)”; ability derivation describes Buffer as noncopyable inline resource-like data. Introduce a dedicated opaque owner whose native representation is pointer-bearing. [VERIFIED: internal/compiler/check/check.go:5371-5382; internal/compiler/ability/ability.go:144-156] The current operation values include “OpRelease OperationKind = "release"”; generated release must invoke the declared destructor rather than only record a semantic event. [VERIFIED: internal/compiler/core/core.go:644-649,852-868; internal/compiler/interp/interp.go:958-960]

## Dependency Order for Planning

1. **Freeze the narrow witness and ABI choices.** Choose the entry-only source type/spelling, pointer-bearing result layout, typed acquire/use errors, regular-file policy, and observer receipt format. Keep the exact locked examples and U64 output. [VERIFIED: .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md:30-59; unchosen details remain assumptions below.]
2. **Add source forms and diagnostics.** Extend AST/parser only enough to pass the path value into acquire, bind its successful owner, borrow it for use, and consume it at release. Add refusals for moved-from use, discard, duplicate release, transfer/escape, and unsupported exits. [VERIFIED: internal/compiler/ast/ast.go:24-45,171-195; internal/compiler/syntax/parser.go:215-255,390-435]
3. **Add per-operation typed-core facts and checker lowering.** Record operation identity, checked foreign declaration/signature, operand mode, failure shape, allocator/destructor association, and acquisition identity. Update serialization/cloning and structural invariants before accepting source. [VERIFIED: internal/compiler/core/core.go:152-201,835-868; internal/compiler/check/check.go:4388-4398,4513-4518,4592-4602]
4. **Implement independent peers before emission.** corevalidate seeds obligations from successful acquires; originvalidate independently checks borrowed provenance; pathoracle verifies every admitted path discharges the exact local owner. Test deletion of every release, duplicate/wrong/fabricated release, use after release, and missing cleanup on a typed use error. [VERIFIED: internal/compiler/corevalidate/corevalidate.go:3216-3259; internal/compiler/originvalidate/originvalidate.go:321-338; internal/compiler/pathoracle/pathoracle.go:230-268]
5. **Add deterministic interpreter model outcomes.** Model acquisition success/failure and use success/error without host filesystem access. Keep source replay refusal for local C. [VERIFIED: internal/compiler/interp/interp.go:962-985; internal/compiler/session/session_app_verify.go:241-267]
6. **Widen the sole C emitter for the exact shape.** Emit the pointer ABI, borrowed C use, consuming destructor, and cleanup on every normal/use-error branch. Clean up before the ordinary diagnostic/outcome. Preserve all refusals for neighboring owner/FFI/pointer shapes. [VERIFIED: internal/compiler/cgen/cgen.go:132-138; internal/compiler/cgen/cgen_program.go:583-611,1001-1006]
7. **Wire explicit bindings, public app command, and runnable example.** Put adapter sources/headers in a relocatable manifest; compile conformance from each checked operation signature; keep the command's direct argv and one-run contract. [VERIFIED: internal/compiler/session/session.go:1142-1166; internal/compiler/native/native_app.go:202-224,292-299; internal/compiler/native/bindings.go:30-45,222-270; cmd/lang/main.go:183-237]
8. **Prove model and physical behavior independently; then add CI.** Run the two public files and all acquisition/use errors; reach the four destructor controls. Add one focused evidence command to the existing macOS/Linux aggregate, not a duplicated full/race suite. [VERIFIED: .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md:63-76; .github/workflows/ci.yml:35-79]

### Existing Fixture and Test Patterns

| Existing seam | Reuse |
|---------------|-------|
| cmd/lang/main_test.go and internal/compiler/native/native_app_test.go | Follow the Phase 22 public CLI, retained build/run, stream, and process-outcome integration-test pattern for the new example. [VERIFIED: cmd/lang/main_test.go; internal/compiler/native/native_app_test.go] |
| internal/compiler/corevalidate/corevalidate_test.go | Extend the existing release mutation matrix, which already moves, drops, duplicates, and invents release records; add the crucial all-releases-removed acquisition-seeded control and wrong-resource case. [VERIFIED: internal/compiler/corevalidate/corevalidate_test.go:1054-1171] |
| internal/compiler/session/session_phase21_contract_test.go | Add a successor contract artifact and contract test; leave the Phase 21 archived contract and its historical test untouched. [VERIFIED: .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md:55-59; internal/compiler/session/session_phase21_contract_test.go] |
| internal/compiler/interp/interptestdirect/interptestdirect.go | Keep as semantic/test-only block runner; do not use it for the real public 0x43 failure or native cleanup evidence. [VERIFIED: internal/compiler/interp/interptestdirect/interptestdirect.go:27-43; .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md:88-90] |

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| File-backed byte acquisition | General Lang filesystem, string, or byte-array runtime | One explicit C adapter with a bounded read and local binding | The adapter owns the fd and returns only the exact allocation contract in scope. [VERIFIED: .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md:17-21,30-40] |
| Memory lifetime | Generic allocator, raw-address integer handle, or global tracing heap | libc malloc/free paired in the checked operation contract | The project requires explicit resource cost and excludes a mandatory tracing heap; pointer-width facts stay target-derived. [VERIFIED: AGENTS.md:28-35,41-44; native/lang_foreign_resource.c:20-31] |
| C ABI proof | Hand-maintained generated record without compile conformance | Shared C typedef plus existing conformance-unit/build path | Record size/alignment, pointer fields, and prototypes must agree on each host ABI. [VERIFIED: internal/compiler/cgen/cgen.go:1174-1203; internal/compiler/native/bindings.go:41-45; internal/compiler/native/native_app.go:202-224] |
| Cleanup proof | Count compiler release events or inspect generated text only | Separate native observer over the real pointer and public process | Events/text prove intent, not actual discharge. [VERIFIED: .planning/REQUIREMENTS.md:52; internal/compiler/interp/interp.go:958-960] |
| App execution | New launcher/backend or shell-command construction | native.RunApplication and existing direct argv runner | The route verifies the retained artifact, passes one token directly, and launches one child. [VERIFIED: internal/compiler/native/native_app.go:292-299,379-416] |

**Key insight:** The proof obligation begins at successful acquire. A peer that discovers owners only by scanning present releases cannot catch total release deletion. Identity and discharge must flow from acquisition through every admitted path. [VERIFIED: .planning/REQUIREMENTS.md:36; internal/compiler/corevalidate/corevalidate.go:3216-3259]

## Common Pitfalls

### Short reads mistaken for one-byte files

**What goes wrong:** A one-byte read is treated as proof of EOF, or a short read is mistaken for failure.  
**Why it happens:** POSIX read can return fewer bytes without EOF.  
**How to avoid:** Loop until EOF or a second byte is observed; cap accepted content at one byte and retry interrupted reads.  
**Warning signs:** Split reads receive inconsistent empty/size classifications. [CITED: pubs.opengroup.org/onlinepubs/9799919799/functions/read.html]

### Release list is used as the source of obligations

**What goes wrong:** Deleting every release also deletes the peer's set of resources to check.  
**Why it happens:** corevalidate currently initializes tracked resources by iterating OpRelease records. The interpreter uses the same seed pattern.  
**How to avoid:** Independently discover each successful acquisition, then require one matching release on every non-transferred terminal path. Keep a reached all-releases-removed mutation. [VERIFIED: internal/compiler/corevalidate/corevalidate.go:3242-3259; internal/compiler/interp/interp.go:545-550]

### Function-level first-symbol contract is reused

**What goes wrong:** Acquire, use, and release inherit one C symbol's signature or failure policy.  
**Why it happens:** The checker chooses “first := infos[0].symbol” and assigns “ForeignContract: contract” once on the function.  
**How to avoid:** Put checked declaration facts on each operation and make each independent consumer inspect that operation. [VERIFIED: internal/compiler/check/check.go:4592-4602; internal/compiler/core/core.go:152-201]

### Integer handles hide target layout

**What goes wrong:** A host address is cast through U64 and treated as a language value.  
**Why it happens:** The frozen fixture writes “*out_handle = (uint64_t)(uintptr_t)block;”; this assumes an integer width for a host pointer.  
**How to avoid:** Keep the pointer as a C pointer field in a checked opaque owner record; reject arbitrary pointer conversion and escape. [VERIFIED: native/lang_foreign_resource.c:20-31; AGENTS.md:41-44]

### The old C fixture is mistaken for live Lang ownership

**What goes wrong:** A successful malloc and release event are mistaken for a live owner.  
**Why it happens:** The legacy open operation calls release before returning.  
**How to avoid:** Preserve the fixture as historical; the new adapter returns live storage and generated cleanup calls its paired destructor. [VERIFIED: native/lang_foreign_resource.c:49-60]

### A synthetic use error stands in for the public operation

**What goes wrong:** A direct interpreter block or synthetic failure event claims to prove the documented 0x43 use error.  
**Why it happens:** Interpreter foreign behavior is modeled, and Phase 22 source replay refuses local C.  
**How to avoid:** Keep model outcomes separate; run 0x43 through public app build/run and independently observe real buffer use and cleanup. [VERIFIED: internal/compiler/interp/interp.go:962-985; internal/compiler/session/session_app_verify.go:241-267]

### Negative controls are not reached or invoke undefined behavior

**What goes wrong:** Tests only inspect generated C, or duplicate-free/use-after-free calls libc with an invalid pointer and crashes nondeterministically.  
**Why it happens:** The mutation may be refused before the intended bad operation or may invoke undefined behavior.  
**How to avoid:** Assert the observer saw each attempt; flag pointer mismatch, second release, or use-after-release before dereferencing/freeing; leave compiler events plausible and unchanged. [ASSUMED; requirement: .planning/REQUIREMENTS.md:52]

### Cleanup runs after the use diagnostic

**What goes wrong:** The error edge writes a diagnostic or terminates before generated cleanup calls release.  
**Why it happens:** Error reporting is emitted directly instead of routing through local discharge.  
**How to avoid:** Route typed use failure through cleanup, consume the owner, then write bounded diagnostic and return nonzero. Defect/process termination remain outside guarantee. [VERIFIED: .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md:44-59]

## Code Examples

### Reuse the public retained-app route

Current native runner signature:

~~~go
func RunApplication(ctx context.Context, artifactPath, input string, stdout, stderr io.Writer) (RunOutcome, error)
~~~

The runner invokes one child with one opaque argv token; extend the accepted entry representation without changing this transport. [VERIFIED: internal/compiler/native/native_app.go:292-299]

### Current pipeline signatures to extend

~~~go
func BuildApplication(ctx context.Context, source []byte, outputPath string, runner native.Runner, manifestPath ...string) (native.BuildReceipt, []diagnostic.Diagnostic, error)
func EmitApplication(program core.Program) (string, error)
func Validate(input core.Program) Result
func ValidatePublished(program core.Program) []Problem
func EnumeratePaths(functionID, entryBlockID string, idx blockIndex, cap int) ([][]string, error)
func Run(program core.Program, functionName, input string) (Execution, error)
func RunApplication(ctx context.Context, artifactPath, input string, stdout, stderr io.Writer) (RunOutcome, error)
~~~

These existing interfaces establish the trust boundary: session checks source and validates typed core before calling EmitApplication; interpreter, core validator, origin validator, path oracle, and native runner remain separate consumers. Add operation-specific facts through these seams rather than creating a parallel compiler route. [VERIFIED: internal/compiler/session/session.go:1142-1166; internal/compiler/cgen/cgen.go:132-138; internal/compiler/corevalidate/corevalidate.go:202-217; internal/compiler/originvalidate/originvalidate.go:563-571; internal/compiler/pathoracle/pathoracle.go:237-268; internal/compiler/interp/interp.go:166; internal/compiler/native/native_app.go:292-299]

### Preserve the one-byte U64 result writer

The existing writer helper is emitProgramU64ApplicationWriter; its emitted C begins “static int lang_write_u64_plain(uint64_t value)”. Keep the returned byte on this decimal U64 path. [VERIFIED: internal/compiler/cgen/cgen_program.go:1001-1006]

### Resource transition replay

~~~text
acquire(path) succeeds  -> owner O becomes live
acquire(path) fails     -> no owner exists
borrow-use(O) succeeds  -> O remains live
borrow-use(O) fails     -> O remains live on the error edge
release(O)              -> O is consumed; paired destructor runs once
terminal path           -> no locally owned O remains
~~~

This is a proposed single-frame rule, not an existing implementation. [ASSUMED; scope: .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md:44-59]

## State of the Art

- Keep semantic lifetime, interpreter model results, compiler event records, and physical native allocation/free as distinct claims. Each requires its own evidence; the roadmap directs updates to all six compiler consumers without sharing a producer's conclusion. [VERIFIED: .planning/PRODUCT-ROADMAP.md:83-102]
- Keep the native ABI pointer-typed and compile it against target C layout. POSIX says invalid/double free is undefined; observer controls should detect the lifecycle violation before invoking that undefined behavior. [CITED: pubs.opengroup.org/onlinepubs/9799919799/functions/free.html]
- Keep sanitizer evidence distinct from observer evidence. A sanitizer checks memory safety for the exercised build; the observer records expected allocation identity and discharge. Neither claim subsumes the other. [VERIFIED: AGENTS.md:38-40; .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md:63-76]

**Deprecated/outdated:**

- The legacy foreign-resource shim is not a live Lang owner because it frees before returning. Keep its old tests and contract intact and add a successor artifact rather than rewriting Phase 21 evidence. [VERIFIED: native/lang_foreign_resource.c:49-60; .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md:55-59]

## Living Roadmap Handoff

AGENTS.md asks for the three next useful capabilities at planning transitions. The living documents and inspected refusal boundaries still agree on this order:

| Capability | User-visible program / blocker | Smallest complete slice and checker change | Evidence/debt, owner, reprioritizing observation |
|------------|---------------------------------|-------------------------------------------|-------------------------------------------------|
| Phase 23: live file byte | Read files with 0x41/0x42 and report their values; current emitter refuses FFI and existing C fixture frees before return. | One opaque path token, one malloc-backed local owner, per-operation contracts, borrowed use, generated release; update check/corevalidate/originvalidate/pathoracle/interp/cgen. | FFI-03, RES-04/07/08/09, local EVD-09 evidence. Owner: Phase 23. Reduce/reorder only if a narrow ABI or physical observer needs broader scope. |
| Phase 24: call and typed-error transfer | Delegate use through a helper and release once after normal/error results; current phase excludes transfer. | Dynamic activation identity, cross-frame ownership/typed-error cleanup; extend independent peers and generated cleanup. | RES-05/06, OWN-10/11/12, full EVD-09. Owner: Phase 24. Reprioritize only if Phase 23's minimum consumer requires transfer. |
| Phase 25: bounded pointer families | Complete a reusable byte utility; current emitter refuses shared/exclusive helper shapes. | One bounded helper per access family, borrow/escape peers, conservative lowering without unsupported alias claims. | NAT-11/12/13, EVD-10, DX-14/15. Owner: Phase 25. Move earlier only if Phase 23/24 witnesses require a family. |

[VERIFIED: AGENTS.md:98-122; .planning/PRODUCT-ROADMAP.md:150-192; .planning/LANGUAGE-MATURITY.md:57-81]

## Assumptions Log

> Historical research assumptions. The dated resolutions below select the Phase 23 implementation approach; these rows preserve what was uncertain when research was first written.

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Use one new entry-only nominal path token and one dedicated opaque owner type; exact source names/syntax remain open. | Operation and ABI Pattern | A type could accidentally enable general strings or conflict with source conventions. |
| A2 | A pointer-bearing record with fixed-width bounded length/status is the smallest portable foreign ABI; exact fields/error variants remain open. | Operation and ABI Pattern | Header/generated C layout mismatch or unsupported target layout breaks the ABI. |
| A3 | Accept only regular files; special objects such as FIFOs/devices should return typed acquire failure to avoid unbounded blocking. | Open Questions | This narrows caller-selected path behavior beyond a locked decision. |
| A4 | A separate observer translation unit with allocator/free shims and atexit receipt can give cross-host evidence without linker-specific wrapping. | Operation and ABI Pattern | The observer may not be sufficiently independent or may behave differently across Apple/Linux linkers. |
| A5 | Allocate one byte before reading so empty/oversize/read failures exercise adapter partial-allocation cleanup. | Operation and ABI Pattern | A different allocation order may leave the partial-cleanup guarantee untested. |
| A6 | A test-only decoy allocation can make wrong-resource destruction reachable without broadening the user program beyond one path and owner. | Validation Architecture | The control might not establish the intended resource-identity property. |
| A7 | Focused Phase 23 evidence has enough regression value for the existing aggregate on both hosts. | Validation Architecture | If cost is too high, required cross-host receipts may be missing or duplicate work may result. |

## Open Questions

1. **Are non-regular filesystem objects valid inputs?**
   - What we know: The token and content are bounded; the context does not classify the opened object. [VERIFIED: .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md:30-40]
   - What's unclear: Whether a FIFO, device, or directory should be rejected or may block/fail under POSIX semantics.
   - Recommendation: Treat the witness as a regular file and reject non-regular objects before a potentially blocking read; record this as a phase-local assumption if selected. [ASSUMED]
   - **RESOLVED (2026-09-27):** Accept regular files only. Open with `O_RDONLY | O_NONBLOCK`, then use `fstat` and reject a non-regular descriptor before any read; this avoids blocking while opening a FIFO. Symlinks may resolve to a regular file. The adapter closes every opened descriptor. Non-regular and `fstat` failures are typed acquisition errors. This selects the regular-file boundary already used by `native.readBindingFile` and the Phase 23 task contract; it does not broaden the public path API.
2. **What source names and C record layout should be frozen?**
   - What we know: The path is opaque; the owner is noncopyable; acquire/use/release have distinct contracts. [VERIFIED: .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md:30-59]
   - What's unclear: Nominal names, record fields, and machine-readable failure variants.
   - Recommendation: Choose the smallest closed syntax and a pointer-preserving record, then pin size/alignment/offset/prototype with compiler-built conformance on both hosts. [ASSUMED]
   - **RESOLVED (2026-09-27):** Use compiler-known entry-only `PathToken` and noncopyable `FileByteOwner`, parsed by the existing identifier `TypeRef` path in `syntax.parser` and the existing `try` binding form (`testdata/phase4/foreign_acquire_one.lang`); no new AST or parser production is required. The source witness is `fn main(path: PathToken) -> U64`, with `let owner = try lang_file_byte_acquire(path)`, then `let value = try lang_file_byte_use(owner)`, then `value`. The generated entry accepts exactly one path token of 1..4096 bytes and rejects an embedded NUL before calling the adapter; it passes the accepted token unchanged. The acquire declaration carries `mode: acquire`, `release: "lang_file_byte_release"`, `allocator: "libc_malloc"`, `fails: AcquireError`, and the existing required unwind/nonlocal-exit policies; the use declaration carries `mode: borrow`, `fails: UseError`, and the same required exit policies. Generated local cleanup creates an `OpRelease` with its own checked consuming contract and explicit destructor symbol, without adding a general source destructor-call form. Source `AcquireError` alternatives are `EmptyFile`, `FileTooLong`, `OpenFailed`, `NotRegular`, `ReadFailed`, `AllocFailed`, `CloseFailed`, `InvalidPath`; source `UseError` alternatives are `UnsupportedByte` and `InvalidOwner`. The header defines `lang_file_byte_owner { unsigned char *data; uint64_t length; }`, `lang_file_byte_acquire_result { int32_t status; lang_file_byte_owner owner; }`, and `lang_file_byte_use_result { int32_t status; uint64_t value; }`. Prototypes are `lang_file_byte_acquire_result lang_file_byte_acquire(const char *path)`, `lang_file_byte_use_result lang_file_byte_use(lang_file_byte_owner owner)`, and `void lang_file_byte_release(lang_file_byte_owner owner)`. `status == 0` means success; acquire statuses 1..8 match the listed alternatives in order, and use statuses 1..2 match its alternatives. Every failure has null owner pointer, zero length/value, and a fully initialized record; success has nonnull pointer and length one. The entry maps `PathToken` directly from its single bounded argv token; only these three foreign operations may use `FileByteOwner`. On the initial 64-bit macOS/Linux targets, C17 conformance must establish owner size/alignment 16/8 and offsets data=0,length=8; acquire result 24/8 and offsets status=0,owner=8; use result 16/8 and offsets status=0,value=8. Derive and check these facts with `sizeof`, `_Alignof`, and `offsetof` against the actual installed target/header, then refuse an unproved target rather than assuming Apple arm64 layout. Existing `BindingManifest` typed symbol probes (`native/bindings.go`) compile all three declared prototypes separately.
3. **How should the observer expose evidence?**
   - What we know: It must independently observe pointer allocation/use/free, run through public app entry, and prove all four controls were reached. [VERIFIED: .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md:63-76]
   - What's unclear: Whether a private file receipt, pipe, or exit status best avoids changing stdout/stderr.
   - Recommendation: Use a separate observer with a private receipt containing ordered pointer identities and final outstanding count; demonstrate it on both hosts first. [ASSUMED]
   - **RESOLVED (2026-09-27):** Use a separately compiled test-only C17 observer in the explicit temporary manifest root. The adapter/header expose narrow C preprocessor hook names for allocation, use observation, and destruction that default to real libc operations in the shipped manifest; the test harness copies the same adapter/header to a temporary manifest root and binds those hooks to separately compiled observer functions. Valid observer calls delegate to real malloc/free and read the same returned buffer. The observer assigns a process-local ordinal to each actual returned pointer and writes a private receipt selected by a test-only `LANG_PHASE23_OBSERVER_RECEIPT` environment path. Record acquire, borrowed use, release, and `atexit` final outstanding count in order; keep stdout/stderr untouched. To prove cleanup precedes error reporting, the test-only generated-C copy places an exact-marker observer report-boundary call immediately before the existing diagnostic, after the emitted cleanup; the observer records its order without replacing the real diagnostic or changing production output. Each negative control first records that its intended branch was reached and rejects invalid early, duplicate, or unrelated-pointer operations before undefined libc behavior. Build/run harness follows `native_app_test.go` subprocess/timeout patterns and `session.go` marker mutation; `23-PATTERNS.md` correctly records that no direct physical-lifetime observer analog exists.
4. **Which adapter errors can occur after allocation?**
   - What we know: A partial allocation must be freed before failure and a failed acquire creates no Lang owner. [VERIFIED: .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md:44-54]
   - What's unclear: Whether close failure is an acquire error and whether it occurs before owner publication.
   - Recommendation: Specify read/size/close/allocation failures and pointer/length state in the contract; ensure any allocated pointer is freed on every failure branch. [ASSUMED]
   - **RESOLVED (2026-09-27):** After successful open and regular-file `fstat`, allocate one byte before reading. Empty, second-byte, read, and close failures after allocation all free that partial byte and return the matching initialized typed failure with no Lang owner. `malloc` failure returns `AllocFailed`; open/path and non-regular failures precede allocation. Retry interrupted reads, probe EOF/second byte, and treat close failure as `CloseFailed` before publishing a successful owner. The `0x43` case is a separate `UnsupportedByte` use failure after ownership has been published; generated cleanup releases the owner before the ordinary nonzero diagnostic.

### Plan 23-04 implementation status — 2026-09-28

**Owner:** Plan 23-04 owns the adapter's acquisition error boundary and its
native/public boundary tests. The regular-file policy above is now implemented:
open uses `O_RDONLY | O_NONBLOCK`, `fstat` admits regular files before any
allocation or read, and all opened descriptors receive one adapter close call.
Empty, two-byte, read, and close failures after allocation return their distinct
status with a null pointer and zero length after freeing the partial byte.
Failures carry a bounded constant diagnostic naming the typed status; caller
path bytes are never copied into diagnostics.

**Newly executed evidence:** on the local macOS arm64 host,
`GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/native -run '^TestPhase23AcquireFailuresInitializeAndFreePartialAllocations$' -count=1`
passed with injected open, `fstat`, allocation, read, and close outcomes,
including both read positions and `EINTR` retry. This is a local focused test,
not a Linux receipt; the Ubuntu CI lane remains the owner for Linux evidence.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|-------------|-----------|---------|----------|
| Go toolchain | Compiler and test suite | Yes | go1.24.0 darwin/arm64 | Existing CI installs the version from go.mod. [VERIFIED: local tool probe 2026-09-27; go.mod:3] |
| Clang | Generated C and adapter compile/link | Yes | Apple clang 21.0.0, arm64-apple-darwin25.6.0 | Existing macOS/Linux CI requires clang. The Linux result is a CI receipt, not a local result. [VERIFIED: local tool probe 2026-09-27; .github/workflows/ci.yml:35-78] |
| Linux native host | Cross-host ABI and cleanup evidence | Not available in this local macOS session | — | ubuntu-latest CI lane; until its receipt exists, Linux evidence remains incomplete. [VERIFIED: local tool probe 2026-09-27; .github/workflows/ci.yml:59-78; .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md:71-76] |

**Missing dependencies with no fallback:** None identified; macOS/Linux execution is required evidence and CI supplies a Linux lane.  
**Missing dependencies with fallback:** Local Linux execution is unavailable; use the Ubuntu CI lane and report it separately from local macOS observations.

## Validation Architecture

Nyquist validation is enabled; config contains “"nyquist_validation": true”. Security enforcement is enabled and ASVS level is 1; config contains “"security_enforcement": true” and “"security_asvs_level": 1”. [VERIFIED: .planning/config.json:22-27,49-50]

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go testing package from the module's Go 1.24 toolchain; native integration compiles through installed Clang. [VERIFIED: go.mod:3; internal/compiler/native/native_app.go:114-127] |
| Config file | go.mod; existing CI does not use a separate test-runner configuration. [VERIFIED: go.mod:1-3; .github/workflows/ci.yml:44-57] |
| Quick run command | Proposed focused command after adding tests: go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/pathoracle -run '^TestPhase23' -count=1. [ASSUMED: future tests] |
| Full suite command | Existing CI commands: go vet ./..., go build ./..., go test ./..., and go test -race ./.... [VERIFIED: .github/workflows/ci.yml:52-57] |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| FFI-03 | Every acquire/use/release has its own checked signature, operand mode, failure behavior, and destructor pairing; wrong operation contract is refused. | Unit + C ABI integration | go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/cgen ./internal/compiler/native -run '^TestPhase23' -count=1 | No Phase 23 tests yet; add beside existing foreign/check/corevalidate/cgen tests. |
| RES-04 | Public app obtains a real pointer-bearing one-byte buffer, uses it while live, and releases through generated code. | Native application integration | go test ./internal/compiler/session ./internal/compiler/native ./cmd/lang -run '^TestPhase23PublicFileByte' -count=1 | No; add public CLI test and example utility. |
| RES-07 | Empty, overlong, unreadable, malformed-status/pointer/length, allocation-failure, and partial-acquire cases produce distinct bounded typed errors with no Lang owner and no adapter leak. | Unit + native adapter integration | go test ./internal/compiler/native ./internal/compiler/session -run '^TestPhase23Acquire' -count=1 | No; include observer counts on error cases. |
| RES-08 | Discarding the owner is rejected before C serialization or is immediately and correctly consumed. | Source admission + emitter boundary | go test ./internal/compiler/check ./internal/compiler/cgen -run '^TestPhase23Discard' -count=1 | No; recommended behavior is refusal before cgen.EmitApplication. |
| RES-09 | Acquisition-seeded peers reject all-release-deleted, duplicate, premature, wrong-resource, and fabricated release on every admitted path. | Independent peer mutation suite | go test ./internal/compiler/corevalidate ./internal/compiler/pathoracle -run '^TestPhase23ResourceMutation' -count=1 | No; assert controls are reached and include use failure after acquire. |
| EVD-09 local slice | Observer proves malloc, later use, matching free before exit; all four destructor controls fail with plausible compiler events. | Native public process + independent observer | Proposed: sh scripts/verify-phase23.sh | No Phase 23 test/receipt found in the repository inventory; add focused script and evidence-aggregate step. Requirements tracker assigns full EVD-09 closure to Phase 24; Phase 23 reports the local-operation slice only. [VERIFIED: repository file inventory 2026-09-27; .planning/REQUIREMENTS.md:52,164; .planning/ROADMAP.md:88-111] |
| EVD-10 host receipts (D-23-06 subset) | Exact native shape runs on macOS and Linux with build identity, compiler, target, flags, and expected result recorded. | Host-specific native receipts | Proposed: sh scripts/verify-phase23.sh | No Phase 23 receipt; CI's matrix line is “os: [ubuntu-latest, macos-latest]”. [VERIFIED: .github/workflows/ci.yml:35-78; .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md:71-76] |
| EVD-11 boundary | Model outcomes compare with independent answers but never claim filesystem IO or physical release. | Model replay | go test ./internal/compiler/interp ./internal/compiler/session -run '^TestPhase23Model' -count=1 | Existing replay/model framework; no Phase 23 fixture. [VERIFIED: internal/compiler/session/session_app_verify.go:241-267,360-402] |
| DX-14 / DX-15 | Clean-checkout command produces two different byte results and admitted error; diagnostics remain structured, bounded, attributed, and explain refused ownership. | CLI + documentation contract | go test ./cmd/lang ./internal/compiler/session -run '^TestPhase23(Readme|Public)' -count=1 | Existing README contract/CLI patterns; add the Phase 23 contract test. [VERIFIED: cmd/lang/main_test.go; .planning/REQUIREMENTS.md:58-59] |

These are proposed test names/commands, not tests run in this research pass. The research request is planning-focused; no test suite was executed. [VERIFIED: current research-session activity]

### Sampling Rate

- **Per checker/core task:** Run focused tests for the changed producer and independent peer; keep mutation cases deterministic and bounded.
- **Per C-emission/native task:** Run one positive public-app case and the reached control for the changed operation; compile generated C under the current strict C17 flags.
- **Per wave integration:** Run the Phase 23 focused source/model/native command on the current host.
- **Phase gate:** Run existing full vet/build/test/race commands and obtain native receipts on both macOS and Linux before marking host evidence complete.
- **CI ownership:** Add one focused Phase 23 command to existing evidence-aggregate jobs on both hosts. Do not repeat full/race/vet work there. Add a sanitizer variant only if it answers a distinct lifetime-safety question and one job owns it per host. [VERIFIED: .github/workflows/ci.yml:35-79; .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md:71-76]

### Wave 0 Gaps

- [ ] Source fixtures for path-token success, empty/oversized acquisition, missing/unreadable input, successful use, and public 0x43 use error.
- [ ] Operation-specific ABI/error contract fixtures, plus refusals for discard, moved-from use, duplicate release, escape, and unsupported exits.
- [ ] corevalidate/pathoracle mutations seeded from acquisition, especially deletion of every release.
- [ ] Independent native observer plus reached omitted/premature/duplicate/wrong-resource public-process controls; include positive malloc → use → matching free → exit receipt.
- [ ] Phase 23 example with explicit binding manifest and developer README.
- [ ] Focused verification command in existing macOS/Linux evidence-aggregate jobs.

The repository already has Go test infrastructure, public application tests, mutation helpers, local-C manifest tests, Clang-required CI, and both host lanes. Phase-specific acceptance evidence is not yet present. [VERIFIED: cmd/lang/main_test.go; internal/compiler/native/bindings_test.go; internal/compiler/corevalidate/corevalidate_test.go:1054-1171; .github/workflows/ci.yml:35-79]

## Security Domain

Security enforcement is enabled at ASVS level 1. The Phase 23 config lines are “\"security_enforcement\": true” and “\"security_asvs_level\": 1”. ASVS 5.0.0 is written for web applications/services, so apply only the relevant input, file, and memory controls to this native CLI; do not claim web compliance. [VERIFIED: .planning/config.json:49-50; CITED: github.com/OWASP/ASVS/blob/master/5.0/en/0x02-Preface.md]

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V1 Encoding and Sanitization | Yes, adapted | Keep the token opaque argv data; bound it, never splice it into C or shell text, and preserve pointer/init/length/release facts. [CITED: github.com/OWASP/ASVS/blob/master/5.0/docs_en/OWASP_Application_Security_Verification_Standard_5.0.0_en.flat.json; VERIFIED: internal/compiler/native/native_app.go:379-416; internal/compiler/native/bindings.go:222-270] |
| V2 Validation and Business Logic | Yes | Enforce one token, one accepted file byte, and distinct empty/oversize/acquisition/use error cases at the trusted app/adapter boundary. [CITED: github.com/OWASP/ASVS/blob/master/5.0/docs_en/OWASP_Application_Security_Verification_Standard_5.0.0_en.flat.json; VERIFIED: .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md:30-54] |
| V5 File Handling | Yes, narrowly adapted | Document permitted input shape and maximum content; reject content outside the one-byte contract. This is a local read, not an upload/store/download workflow. [CITED: github.com/OWASP/ASVS/blob/master/5.0/docs_en/OWASP_Application_Security_Verification_Standard_5.0.0_en.flat.json; VERIFIED: .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md:30-40] |
| V3 Web Frontend Security | No | No browser, HTTP frontend, origin, or response-header boundary is introduced. [CITED: github.com/OWASP/ASVS/blob/master/5.0/docs_en/OWASP_Application_Security_Verification_Standard_5.0.0_en.flat.json; VERIFIED: AGENTS.md:7-15] |
| V4 API and Web Service | No | The feature is a local retained CLI app with argv input and no network API. [CITED: github.com/OWASP/ASVS/blob/master/5.0/docs_en/OWASP_Application_Security_Verification_Standard_5.0.0_en.flat.json; VERIFIED: cmd/lang/main.go:183-237] |
| V6 Authentication | No | No authentication or account/session boundary is added; file access remains under invoking OS-user authority. [CITED: github.com/OWASP/ASVS/blob/master/5.0/docs_en/OWASP_Application_Security_Verification_Standard_5.0.0_en.flat.json; VERIFIED: .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md:9-21] |

### Known Threat Patterns for the Go/C17 Native-App Stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Path token used as shell text or generated C source | Tampering / Elevation of privilege | Keep argv transport direct; admit bounded opaque bytes; compile only explicit manifest inputs confined to its root. [VERIFIED: internal/compiler/native/native_app.go:379-416; internal/compiler/native/bindings.go:222-270] |
| Unbounded or special-file read | Denial of service | Read only the byte plus second-byte probe; decide and verify regular-file handling before potentially blocking reads. [CITED: pubs.opengroup.org/onlinepubs/9799919799/functions/read.html; ASSUMED: regular-file restriction] |
| Use after free, double free, or wrong allocator pairing | Tampering / Elevation of privilege | Noncopyable owner, borrow preserves obligation, exact consuming destructor/allocator pair, and independent pointer-identity observer; never rely on process reclamation. [CITED: pubs.opengroup.org/onlinepubs/9799919799/functions/free.html; VERIFIED: .planning/REQUIREMENTS.md:36,52] |
| Mismatched C struct/prototype ABI or uninitialized pointer/length | Tampering | Shared header typedef, generated declaration conformance, initialized status/pointer/length, and per-host compiler/target receipts. [VERIFIED: internal/compiler/native/bindings.go:30-45; internal/compiler/cgen/cgen.go:1174-1203] |
| Adapter misbehavior despite a checked declaration | Tampering / Repudiation | Treat declared C as a trust boundary; the compiler checks contracts but does not prove arbitrary C correct. Limit evidence claims to the audited adapter. [VERIFIED: internal/compiler/native/bindings.go:30-45; .planning/PRODUCT-ROADMAP.md:129-136] |

## Sources

### Repository and host sources (HIGH confidence)

- Repository source inspected this session: AGENTS.md, phase CONTEXT/REQUIREMENTS/STATE, PRODUCT-ROADMAP, LANGUAGE-MATURITY, M004 research docs, compiler/check/core/corevalidate/originvalidate/pathoracle/interp/cgen/session/native, CLI, CI, and the legacy C resource fixture. In-repo claims carry exact path/line citations above.
### Official external references (MEDIUM confidence)

- POSIX.1-2024 read: https://pubs.opengroup.org/onlinepubs/9799919799/functions/read.html — short-read behavior for adapter design.
- POSIX.1-2024 free: https://pubs.opengroup.org/onlinepubs/9799919799/functions/free.html — allocator pairing and invalid/double-free behavior.
- OWASP ASVS 5.0.0 official requirements index: https://github.com/OWASP/ASVS/blob/master/5.0/docs_en/OWASP_Application_Security_Verification_Standard_5.0.0_en.flat.json — category names and relevant input/file controls.
- OWASP ASVS 5.0.0 preface: https://github.com/OWASP/ASVS/blob/master/5.0/en/0x02-Preface.md — standard scope.

### Local Environment

- Tool probes on 2026-09-27 reported Go 1.24.0 and Apple Clang 21.0.0 targeting arm64-apple-darwin25.6.0. Linux was not available locally; repository CI has an Ubuntu lane. This is local probe evidence, not a Linux result.
- For local Go commands, STATE.md specifies “GOCACHE=/tmp/ai-lang-verification-gocache”; CI uses the normal Go setup-go cache. [VERIFIED: .planning/STATE.md:80; .github/workflows/ci.yml:44-48]

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — repository constraints, go.mod, Clang invocation, and CI host jobs were inspected.
- Architecture: MEDIUM — current producer/peer/emitter boundaries are source-verified; the dated Open Questions resolutions fix source form and ABI layout for execution, while implementation and cross-host conformance remain unverified.
- Pitfalls: MEDIUM — release seeding and legacy fixture behavior are source-verified; cross-host observer integration is proposed, not executed.

**Research date:** 2026-09-27
**Valid until:** 2026-10-27 for stable compiler architecture; recheck CI/Clang and POSIX references if toolchain or host lanes change.

Research is complete from source inspection, targeted official-document lookup, and host probes. No implementation tests or subjective UAT were run during this research pass. [VERIFIED: research-session activity]
