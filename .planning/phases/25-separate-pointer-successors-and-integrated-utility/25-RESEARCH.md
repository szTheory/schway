# Phase 25: Separate Pointer Successors and Integrated Utility - Research

**Researched:** 2026-10-01  
**Domain:** Schway checker, borrow semantics, conservative C pointer lowering, native evidence  
**Confidence:** HIGH for inspected implementation boundaries and locked scope; MEDIUM for the amount of independent peer expansion required

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

### Integrated utility

- **D-25-01:** Use one visible application flow: acquire and transfer the
  FileByteOwner, read its byte as U64, pass that value through a shared
  read/copy helper and then an exclusive read/copy helper, and return the last
  result. Each helper result feeds the next step so both calls affect the
  independent expected result. Use the existing success bytes and 0x43
  failure. The typed use failure happens before either infallible pointer
  helper; do not invent a pointer-specific error.
- **D-25-02:** Keep helper parameter and body syntax within the existing
  ordinary Schway function and borrow forms. Shared and exclusive are separate
  named helpers with separate checked source witnesses. Do not add a new
  foreign contract, raw pointer syntax, owner projection, or access-mode
  parameter to express this slice. Actual pointer-parameter C must result from
  checked source lowering and be reflected accurately in manifests.
- **D-25-03:** Extend the fixed Phase 24 resource-transfer caller shape only as
  far as needed to admit this composition. Update each affected independent
  validator, interpreter path, the C emitter, and the evidence/application
  boundary. The current caller checker assumes two bindings and a fixed
  operation sequence; lifting only the emitter refusal is insufficient.

### Borrow behavior and diagnostics

- **D-25-04:** Shared means compatible readers; exclusive means no conflicting
  access while that loan is live. The exclusive helper remains read/copy-only
  and grants no mutation capability. Keep the two families' positive,
  incompatible/conflicting-access, and escape witnesses distinct. A shared
  loan may end before the exclusive loan begins; an overlapping shared and
  exclusive loan must be rejected.
- **D-25-05:** Preserve the existing structured diagnostic wire contract.
  Distinguish a borrow conflict from a safe shape the implementation does not
  support. Attribute the primary span to the offending use or argument and
  include borrow-origin and conflict/escape causes when available. Explain the
  boundary without suggesting a safety-weakening workaround; do not emit a
  repair unless that edit is demonstrably safe.

### Evidence, documentation, and dependencies

- **D-25-06:** Provide one clean-checkout quick start for the integrated
  utility, with explicit build inputs/C bindings and a directly navigable
  evidence index. The index may share setup instructions, but every family,
  host, and applicable optimizer/sanitizer lane remains separately visible.
  Missing host or lane evidence stays incomplete. Reuse existing CI ownership
  when it answers the same evidence question rather than duplicating expensive
  full-suite work.
- **D-25-07:** Each family must have a positive native result whose expected
  value changes under a reached helper-local wrong-result control, plus its own
  independent conflict and escape controls. Assert emitted C contains the
  actual pointer parameter and that its attribute manifest agrees. Do not infer
  aliasing, capture, alignment, or ownership guarantees from a pointer shape or
  from Phase 24's physical resource observer. Model replay remains distinct
  from host IO and native pointer behavior.
- **D-25-08:** Prefer the existing Go standard library and copy-local code over
  new runtime/build dependencies. Add a dependency only if planning identifies
  a concrete need that justifies its transitive surface and audit cost.

### the agent's Discretion

- Choose helper and example names, the smallest precise source grammar
  extensions needed by the already-approved composition, diagnostic codes and
  wording, and the clean-checkout command layout.
- Choose separate family-focused fixtures and how their outputs combine in
  the application, provided both results contribute to 65/66 and the error
  path remains the inherited 0x43 typed use failure.
- Choose focused verification placement and the split across existing CI
  lanes. Record source inspection, newly executed checks, and historical
  receipts separately. Current Phase 24 receipts are background only; they
  are not Phase 25 evidence.

### Deferred Ideas (OUT OF SCOPE)

- Direct pointer access to fields or bytes inside the owned buffer, owner
  projection, and general buffer views remain outside this bounded scalar
  witness.
- Mutation, forwarding, retention, callbacks, nonlocal exits, and wider pointer
  forms remain refused.
- Arithmetic, loops, FizzBuzz, general strings/arrays, JSON, and a reusable
  library/module system belong to later consumer-led work.
- One bounded JSON configuration consumer remains future work after a named
  program justifies it.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| NAT-11 | The sole production emitter admits a bounded shared plain by-pointer read/copy helper with actual pointer-parameter C, a source-level consumer, independent shared-borrow checking, and its own positive and incompatible-access/escape negative witnesses. | Separate the shared classifier/admission from historical exclusive `restrict` metadata; extend independent borrow derivations and prove pointer C plus family-specific controls. |
| NAT-12 | The sole production emitter admits a bounded exclusive borrowed-by-pointer read/copy helper with actual pointer-parameter C, a source-level consumer, independent exclusive-borrow checking, and its own positive and conflicting-access/escape negative witnesses. | Extend the exclusive family conservatively; historical `restrict` classifier output is not a Phase 25 attribute grant. |
| NAT-13 | Bounded pointer lowering and its manifests agree on emitted attributes and add no `restrict`, `noalias`, capture, or alignment promise unsupported by checked facts; unsupported mutation, forwarding, retention, callbacks, nonlocal exits, and wider pointer shapes remain structurally refused before C serialization. | Make the emitted C and manifest contract describe only pointer representation; preserve fail-closed structural gates and test the forbidden attribute set. |
| EVD-10 | Every admitted foreign/shared/exclusive family has its own reproducible macOS and Linux native receipt identifying source/build inputs, compiler, target, flags, expected result, and applicable optimizer/sanitizer lanes; missing host or lane evidence keeps that claim incomplete. | Extend the existing per-host evidence pattern with separately identified family, exact source/build identity, expected result, and lane state. Phase 24 receipts are historical context only. |
| DX-14 | From a clean checkout, a developer can follow documented commands to build and run the bounded file-byte utility against two supplied files, observe the expected differing result and an admitted failure, and locate its explicit C bindings and resource-lifetime evidence. | Preserve the Phase 24 utility/build boundary; document integrated inputs, bindings, exact 65/66 outputs, 0x43 failure, and navigable evidence index. |
| DX-15 | Unsupported ownership/resource/pointer uses produce stable structured diagnostics with source attribution and an actionable boundary explanation; examples cover moved-from use, discarded ownership, pointer escape, and unsupported cleanup/exit behavior. | Reuse structured diagnostic schema and cause spans; distinguish access conflicts from unsupported forms and avoid speculative fix-its. |
</phase_requirements>

## Summary

Phase 25 is a bounded extension across several trust boundaries, not a new pointer syntax or a generic FFI design. The already-locked utility reads a real file byte into a `U64`, passes that copied scalar through distinct shared and exclusive read/copy helpers, and returns the last result. Both helper results must influence the observed 65/66 result. `0x43` remains the existing typed use failure before either infallible helper, so it continues to exercise owner cleanup without inventing a pointer-helper error. The copied scalar helpers do not establish access to the `FileByteOwner` allocation itself. [CITED: `.planning/phases/25-separate-pointer-successors-and-integrated-utility/25-CONTEXT.md`; `.planning/REQUIREMENTS.md`]

The source currently does not admit this complete composition. `checkLocalFileByteTransferCaller` recognizes one owner-producing call followed by one fallible `schway_file_byte_use` binding and return; its guard includes the exact discrete check `len(body.Bindings) != 2`, and its refusal text is `PathToken caller may only receive one owner helper, borrow it once, and return U64`. [VERIFIED: `internal/compiler/check/check.go:4044-4045`] `emitProgram` refuses any function selected by either legacy by-pointer classifier with the error text `function %q: by-pointer bodies are not supported by whole-program native emission this phase`. [VERIFIED: `internal/compiler/cgen/cgen_program.go:641-643`] Those classifiers and `EmitForeignManifest` describe legacy pointer shapes separately from production-body admission, and the exclusive metadata helper currently returns `EmittedAttribute{Attr: "restrict", CoreNode: function.ID, Parameter: function.Parameter.ID, JustifiedBy: loanID}`. [VERIFIED: `internal/compiler/cgen/cgen.go:245-278,345-397,1080-1099`] A plan must extend the exact accepted caller shape across checking, independent validation, interpreter, emitter, ABI manifest, and application evidence, while ensuring the new conservative pointer family facts do not inherit that old attribute.

**Primary recommendation:** admit one precise ordinary-function helper shape for each access family, lower each checked `U64` borrow to a real pointer-parameter C signature with no unsupported optimizer attributes, independently derive the family facts at each existing validator boundary, and bind distinct positive/wrong-result/conflict/escape evidence to macOS and Linux receipts. Keep helper spelling within existing function/borrow syntax; exact checker expansion belongs to the plan and should be evidenced with source fixtures before implementation tasks multiply.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Shared/exclusive loan admission and conflict/escape refusal | API / Backend | — | Source checking and independent core/origin/path validators own semantic access facts. |
| Pointer-parameter C representation and attribute manifest | API / Backend | — | The sole production serializer and manifest authority must agree on the facts actually checked. |
| File-byte acquisition, use, transfer, and discharge | API / Backend | Database / Storage (native allocation) | Schway directs use and destructor behavior for its live buffer; the C adapter owns its file descriptor. |
| Retained utility build/run and public evidence index | Browser / Client (CLI client) | API / Backend | The command-line app route is the user-facing entry; session/native code binds build inputs and executes the artifact once. |
| Cross-host evidence receipts | API / Backend | — | Focused native test/CI lanes own claims by family, host, and applicable flags. |

## Project Constraints (from AGENTS.md)

- Start work through the configured GSD workflow before file changes; the parent workflow owns the docs commit. [VERIFIED: repository `AGENTS.md` supplied in task]
- At phase planning and completion, compare `PRODUCT-ROADMAP.md` and `LANGUAGE-MATURITY.md` with source witnesses, refusal boundaries, and named evidence; preserve historical receipts and distinguish source inspection from executed checks. [VERIFIED: repository `AGENTS.md` supplied in task]
- Surface the three next useful capabilities with each one's user-visible program, blocker, smallest complete slice, checker/guarantee changes, evidence/debt references, owner or next action, and a priority-changing observation. Update the living documents only when those facts change and within the active GSD workflow. [VERIFIED: repository `AGENTS.md` supplied in task]
- Prefer a runnable gain in every feature phase. Extra plan count must name a user witness or safety obligation. [VERIFIED: repository `AGENTS.md` supplied in task]
- Treat a second consecutive enabling phase without a runnable gain as a scope-review trigger; keep independent validation and negative controls tied to the introduced capability. [VERIFIED: repository `AGENTS.md` supplied in task]
- Do not invent completeness percentages, require a new backend, or add a general planning framework without a demonstrated consumer need. [VERIFIED: repository `AGENTS.md` supplied in task]
- Preserve standard-library-first and shallow audited dependency boundaries; portability claims must not encode the current Apple arm64 host. [VERIFIED: repository `AGENTS.md` supplied in task]
- Maintain explicit boundaries for untrusted input, unsafe operations, FFI, secrets, and build authority. [VERIFIED: repository `AGENTS.md` supplied in task]

## Standard Stack

No external package is needed or recommended for this phase. The current project stack is Go 1.24 standard-library Stage 0, generated readable C17, installed Clang, and an independent interpreter; Phase 25 should extend those existing paths and the existing C adapter. [CITED: `.planning/PROJECT.md`; `.planning/research/M004/STACK.md`; `.planning/phases/22-native-application-build-and-single-execution/22-CONTEXT.md`]

| Component | Version / policy | Purpose | Why standard here |
|---|---|---|---|
| Go standard library | Go 1.24 (project constraint) | Compiler, checks, interpreter, test harness, evidence encoding | Existing Stage 0 and no new runtime/build dependency. |
| C17 and installed Clang | C17 (project constraint); host compiler recorded per receipt | Sole production native lowering and native behavior | Existing production backend and platform-specific evidence route. |
| Existing phase24 file-byte adapter | In-repo, audited build input | Allocate/use/release the bounded owner | Carries existing trusted FFI and physical cleanup boundary. |

**Installation:** none. There is no external dependency to legitimacy-audit or registry-verify.

## Architecture Patterns

### Recommended implementation boundary

Treat the program as one vertical slice across the source producer and its independent consumers. Existing code has separate implementations in `check`, `corevalidate`, `originvalidate`, `pathoracle`, `interp`, and `cgen`; project architecture calls for independent derivation rather than peers importing checker approval. [CITED: `.planning/PRODUCT-ROADMAP.md`; `.planning/REQUIREMENTS.md`]

| Boundary | Current source witness | Planning consequence |
|---|---|---|
| Source syntax / checker | `examples/phase24/transfer.schway`; checker hard-codes the two-binding owner caller. [VERIFIED: `internal/compiler/check/check.go:4039-4054`] | Add only the helper call/use chain required for the selected utility. Preserve existing `try` failure and owner release semantics. |
| Independent validators | `corevalidate`, `originvalidate`, `pathoracle` derive separate ownership/loan facts. [CITED: `.planning/phases/25-separate-pointer-successors-and-integrated-utility/25-CONTEXT.md`] | Each peer needs its own family-specific derivation and mutations for incompatible/conflicting access and escape. |
| Interpreter | Foreign outcomes are deterministic model outcomes; not proof of actual IO or pointer lowering. [CITED: `.planning/REQUIREMENTS.md`; `.planning/research/M004/ARCHITECTURE.md`] | Model the helper's scalar result, but label model replay separately from host IO and native pointer behavior. |
| C emitter and ABI metadata | The sole serializer refuses both legacy by-pointer bodies; manifest classifier can still describe a legacy exclusive `restrict` fact. [VERIFIED: `internal/compiler/cgen/cgen_program.go:641-645`; `internal/compiler/cgen/cgen.go:381-397,1080-1099`] | Implement narrow family admission at the production emitter and derive actual pointer declaration plus manifest from checked facts. Avoid translating old `restrict` metadata into the new ABI. |
| App/native evidence | Phase 24 already retains app build/run, explicit bindings, and independent resource lifecycle observer. [CITED: `examples/phase24/README.md`; `.planning/phases/24-ownership-transfer-through-calls-and-errors/24-VERIFICATION.md`] | Extend the utility and receipt index, but use pointer-family-specific evidence; the allocation observer does not prove scalar pointer-parameter behavior. |

### Borrow-family shape and ABI

The legacy exclusive classifier only selects a straight-line body whose first operation exclusively borrows the sole parameter, after which operations form an unbroken chain to the return. The shared classifier has the same structure with a shared first borrow. [VERIFIED: `internal/compiler/cgen/cgen.go:245-278,345-379`]

> The current classifier source says: `if first.Kind != core.OpBorrowExclusive || first.SourceID != function.Parameter.ID || first.TargetID == "" { return false }` and, for shared, `if first.Kind != core.OpBorrowShared || first.SourceID != function.Parameter.ID || first.TargetID == "" { return false }`. [VERIFIED: `internal/compiler/cgen/cgen.go:253-255,353-355`]

These classifier shapes are useful as implementation precedent, not proof that the requested source helper currently compiles or that its existing ABI/manifest is safe for Phase 25. Current production emission stops before serialization when either classifier matches. [VERIFIED: `internal/compiler/cgen/cgen_program.go:641-643`]

The existing family classifiers are type-agnostic after core construction: their visible guards distinguish shared versus exclusive loan and parameter origin, with no `U64`-specific exclusion. Thus a by-value source `U64` helper can be family-classified in principle. The classifiers do not themselves emit the production pointer-parameter body: `emitProgram` rejects both families, and the current manifest route can attach historical exclusive `restrict`. A new production lowering must emit the actual pointer parameter and matching call argument without carrying that legacy attribute. The call-site representation/address-taking is the critical unresolved part to settle and pin in a generated-C fixture. [VERIFIED: `internal/compiler/cgen/cgen.go:245-278,345-379`; `internal/compiler/cgen/cgen_program.go:641-643`] A pointer parameter does not logically require `restrict`, but this compatibility observation is not a proof of Schway borrow safety or of the exact generated declaration. [ASSUMED]

Conservative pointer lowering is technically coherent without `restrict`: C's pointer parameter representation alone does not require adding a restrict-qualified declaration. However, C declaration and call sites must agree, and the manifest must represent the admitted ABI without claiming no-alias/capture/alignment. This is an implementation feasibility conclusion from the bounded lowering boundary, not a proof of Schway borrow safety; the latter comes from independent checker peers and family evidence. [ASSUMED]

### Pattern: preserve error-before-helper ordering

The integrated flow must be acquire → transfer → fallible owner use → shared read/copy → exclusive read/copy → return. `0x41` and `0x42` retain expected values 65 and 66; `0x43` fails in the existing use call before the infallible helpers. [CITED: Phase 25 D-25-01 and `25-CONTEXT.md`]

Illustrative flow only (not asserted as accepted source grammar):

```text
owner = acquire(path)
value = try use(owner)          // 0x43 fails here and cleans up
shared_value = shared_copy(value)
result = exclusive_copy(shared_value)
return result
```

The source fixtures should pin the ordinary accepted form and each refused frontier before implementation planning proceeds. Existing source facts include ordinary `fn` syntax and `let first = borrow buffer` / `let second = borrow mut buffer`, but do not establish that either helper is currently accepted or emitted in the proposed composition. [VERIFIED: `testdata/phase3/sequential_shared_then_exclusive_accept.schway`; `testdata/phase5/restrict_borrow.schway`]

### Diagnostics

Keep the existing structured diagnostic wire contract. The diagnostic record has schema, stable ID, code, severity, primary span, message, causes, and repairs. [VERIFIED: `internal/compiler/diagnostic/diagnostic.go:97-117`]

Use the source span at the offending helper argument/use, and report origin/conflicting loan or escape cause when the analysis has it. A true overlapping exclusive/shared access is a borrow conflict; a structurally safe but unsupported form is an unsupported-shape refusal. Do not attach a repair that changes access mode or removes the use unless that edit is proven to preserve behavior. [CITED: Phase 25 D-25-04/D-25-05]

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---|---|---|---|
| Borrow overlap, last-use, and conflict analysis | A Phase 25-only lexical scan or a second liveness law | Existing loan-liveness and independently implemented peer validators | The family distinction concerns live overlap; existing fixtures already cover compatible shared borrows, sequential shared→exclusive access, and overlapping shared/exclusive rejection. [CITED: `testdata/phase3/`] |
| Ownership cleanup accounting | Pointer-address identity or release-operation-seeded obligation tracking | Phase 24 acquisition-derived identities and cleanup path model | Dynamic activations and cleanup guarantees are already independently checked; host addresses stay private to the observer. [CITED: Phase 24 verification and M004 architecture] |
| ABI declaration consistency | Separate handwritten C signature and unrelated manifest constants | One checked ABI fact source consumed independently by emitter and manifest derivation | Prevent prototype/call mismatch and manifest drift. [CITED: `internal/compiler/cgen/cgen.go`; Phase 25 D-25-07] |
| Cross-host evidence | A second duplicated full-suite CI job or a cross-compile-only receipt | Existing macOS/Linux lanes, with one owner per expensive lane and exact family/host receipts | EVD-10 requires native runs and records missing lanes as incomplete. [CITED: `.planning/REQUIREMENTS.md`; `.planning/research/M004/PITFALLS.md`] |

**Key insight:** a pointer-shaped C parameter is only an ABI representation. It does not itself prove lifetime, alias compatibility, non-escape, capture, or actual pointer behavior. Keep each claim attached to the checker/peer facts and exact native evidence that establish it. [CITED: Phase 25 D-25-07; `.planning/research/M004/ARCHITECTURE.md`]

## Common Pitfalls

### 1. Treating legacy classification as production admission

**What goes wrong:** tests or manifest metadata recognize an exclusive/shared body, and a plan assumes the production emitter already accepts it.  
**Why it happens:** historical classifiers and metadata APIs exist for refused shapes.  
**How to avoid:** update the whole-program serializer's explicit refusal only alongside exact-shape checking, peer validation, interpreter behavior, C output, and manifest changes.  
**Warning signs:** a positive check or `EmitForeignManifest` result with `emitProgram` still refusing the program. [VERIFIED: `internal/compiler/cgen/cgen_program.go:641-645`; `internal/compiler/cgen/cgen.go:1080-1099`]

### 2. Letting exclusive classification imply `restrict`

**What goes wrong:** old D-05 metadata is copied into emitted C or new manifest entries, making an unsupported optimizer promise.  
**Why it happens:** `emittedAttributeForByPointerParameter` returns `restrict` for the exclusive classifier, and that remains part of the legacy manifest path.  
**How to avoid:** make Phase 25's C and manifest agree on an empty unsupported-attribute set; separately test actual pointer representation and absence of `restrict`, `noalias`, capture, and alignment promises. Preserve old fixture/golden obligations as historical controls where needed.  
**Warning signs:** deriving current Phase 25 policy from the classifier name, `restrict_borrow` golden, or a pointer parameter alone. [VERIFIED: `internal/compiler/cgen/cgen.go:381-397,1281-1323`; [CITED: Phase 25 D-25-07]]

### 3. Calling by-value lowering a pointer-family success

**What goes wrong:** both helpers return correct scalars but generated signatures still pass `uint64_t` by value.  
**Why it happens:** the result-level fixture does not inspect generated C.  
**How to avoid:** assert family-specific pointer parameter declarations and matching call-site types, then mutate each helper's result in a reached control so each family independently changes the expected application answer.  
**Warning signs:** only interpreter/native equality or final 65/66 output is checked. [CITED: NAT-11/12; Phase 25 D-25-07]

### 4. Conflating unsupported shape with an actual borrow conflict

**What goes wrong:** unsupported safe bodies are diagnosed as conflicting loans, or genuine overlapping access is described as an implementation limitation.  
**Why it happens:** both are refusals but carry different semantic facts and actionable explanations.  
**How to avoid:** preserve distinct diagnostics and cause trails; pin primary spans at offending uses/arguments and keep separate family witnesses for incompatible/conflicting access and escape. Never fabricate a fix-it.  
**Warning signs:** one generic “by-pointer unsupported” diagnostic handles both semantic conflict and supported-subset limits. [CITED: Phase 25 D-25-04/D-25-05]

### 5. Assuming the transfer caller checker is open-ended

**What goes wrong:** checker or emitter gets a local edit, but caller composition or an independent validator still assumes the original fixed operation sequence.  
**Why it happens:** Phase 24's shape is deliberately hand-bounded.  
**How to avoid:** inventory and extend each affected consumer with the intended utility chain, plus family-specific refusal mutations. Plan one end-to-end slice, then separate evidence ownership if necessary.  
**Warning signs:** source parses, but check/core validation/pathoracle/interpreter/emitter diverge or use different operation order. [VERIFIED: `internal/compiler/check/check.go:4039-4064`; [CITED: Phase 25 D-25-03]]

### 6. Reusing Phase 24 receipts as Phase 25 proof

**What goes wrong:** historical host success is presented as proof of pointer-family behavior.  
**Why it happens:** the same app, hosts, and adapter are reused.  
**How to avoid:** retain Phase 24's hosted receipt only as owner-transfer/cleanup history; create distinct family × host × applicable lane receipts bound to current source/build identity, compiler, target, flags, and expected results.  
**Warning signs:** evidence index shows one aggregate pass with no per-family outcomes or pointer-specific controls. [CITED: Phase 25 D-25-06 and discretion; `.planning/phases/24-ownership-transfer-through-calls-and-errors/24-03-SUMMARY.md`]

## Code Examples

### Existing shared-to-exclusive loan boundary (fixture precedent)

The following source shape is already present as a borrow-semantics fixture; it is not a Phase 25 production utility or native witness:

```schway
fn relay(buffer: Buffer) -> Buffer {
  let first = borrow buffer
  let reviewed = borrow first
  let second = borrow mut buffer
  let inspected = borrow second
  let delivered = take buffer
  delivered
}
```

Source: `testdata/phase3/sequential_shared_then_exclusive_accept.schway`. Its accompanying fixture comment states that the shared loan's last use precedes the exclusive borrow; the separate `shared_exclusive_reject.schway` keeps a shared loan live across the exclusive borrow and expects rejection. [VERIFIED: `testdata/phase3/sequential_shared_then_exclusive_accept.schway`; `testdata/phase3/shared_exclusive_reject.schway`]

### Existing exclusive pointer precedent

`testdata/phase5/restrict_borrow.schway` uses `fn touch(buffer: Buffer) -> Buffer` and begins with `let first = borrow mut buffer`, then reborrows to return. It is a legacy lowering/manifest precedent; current production `emitProgram` refuses the classified body. Do not use this historical fixture as proof that Phase 25's `U64` helpers are currently accepted or that `restrict` remains permitted. [VERIFIED: `testdata/phase5/restrict_borrow.schway`; `internal/compiler/cgen/cgen_program.go:641-643`]

## State of the Art

| Prior approach | Phase 25 target | Change / impact |
|---|---|---|
| Legacy single-function by-pointer classification can describe a shape and exclusive manifest metadata may contain `restrict`; whole-program serializer refuses the body. | Narrow production admission for the two checked `U64` read/copy families; no added unsupported optimizer promise; explicit per-family C and manifest evidence. | This is a new runtime admission, not a revival of the retired/legacy emitter. Existing historical fixtures retain value as refusal and metadata regression controls. [VERIFIED: source citations above; CITED: Phase 25 context] |
| Phase 24 application demonstrates ownership transfer/use/cleanup and a typed error. | Integrated application passes the derived `U64` through both distinct helpers; error still occurs before the helpers. | Pointer behavior and owned-allocation lifecycle remain distinct evidence claims. [CITED: Phase 24 verification; Phase 25 D-25-01/D-25-07] |

**Deprecated/outdated:** describing the legacy exclusive `restrict` manifest entry as a required Phase 25 ABI fact. D-25-07 explicitly rules out unsupported alias/capture/alignment guarantees, and the new pointer admission must reconcile actual C and manifest conservatively. [CITED: Phase 25 context]

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | A by-value source `U64` parameter can be lowered at the helper boundary to a real C pointer parameter consistently at both definition and call site without an access-mode parameter or raw-pointer syntax. | Architecture Patterns | If source-to-C call representation cannot be reconciled with ordinary borrow lowering, the smallest approved grammar extension may differ; NAT-11/12 cannot be satisfied by a by-value fallback. |
| A2 | The helper's read/copy result can remain `U64` and flow through existing app result transport without new output/runtime features. | Summary | A hidden output limitation could require revisiting the pinned minimal composition and app path. |

## Planning Resolutions

These decisions were made after source inspection and official Clang documentation review. They settle research questions for planning; they are not implementation or host-execution evidence.

1. **Source syntax — resolved:** Keep the existing ordinary function and body-level borrow forms. The shared witness has the shape fn copy_shared(value: U64) -> U64 { let view = borrow value; view }; the exclusive witness is the same shape with borrow mut value. No parser extension is planned unless implementation demonstrates that this already-locked spelling cannot be represented. The fixture must prove the checker accepts the exact source before widening predicates. [VERIFIED: internal/compiler/syntax/parser.go; existing borrow witnesses in testdata/phase3]
2. **Checked ABI fact — resolved:** Use one narrow checked family fact to select the C pointer parameter, matching caller address argument, and manifest representation. Keep emitProgram as the sole production serializer and assert C plus manifest for each family. The current emitter refusal and historical classifier/manifest do not establish the new behavior. [CITED: Phase 25 D-25-02/D-25-03; internal/compiler/cgen/cgen.go; internal/compiler/cgen/cgen_program.go]
3. **Applicable native lanes — resolved:** Run the foreign, shared, and exclusive positive witnesses separately on both macOS and Linux under baseline -O0, optimized -O2, and the existing sanitizer lane -O1 -g -fno-omit-frame-pointer -fsanitize=address,undefined -fno-sanitize-recover=all. Do not add LTO as a Phase 25 lane: each generated helper/caller pair is one translation unit, and the phase makes no inter-unit alias or LTO-specific optimizer claim. The repository pins the sanitizer flags in internal/compiler/native/sanitize.go; Clang documents ASan and UBSan support on macOS/Linux, pointer/alignment checks in UBSan, and -O0/-O2 behavior. Missing host/lane evidence remains incomplete. [VERIFIED: internal/compiler/native/sanitize.go; CITED: https://clang.llvm.org/docs/AddressSanitizer.html, https://clang.llvm.org/docs/UndefinedBehaviorSanitizer.html, https://clang.llvm.org/docs/CommandGuide/clang.html]

## Environment Availability

This phase has no external package dependency. Clang is available in this research environment as Apple clang 21.0.1; no compilation was performed. The required macOS/Linux native host receipts remain future evidence prerequisites. Do not treat the historical Phase 24 receipt as current Phase 25 evidence. [VERIFIED: command availability/version probe in this session; CITED: Phase 25 D-25-06 and discretion; `.planning/phases/24-ownership-transfer-through-calls-and-errors/24-03-SUMMARY.md`]

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Clang | Native C build and family evidence | ✓ | Apple clang 21.0.1 (`clang-2100.1.1.101`) | None for native claim; host receipts still required |
| macOS native host lane | EVD-10 macOS receipt | Historical Phase 24 receipt only; current Phase 25 availability not established | Historical Darwin/arm64 receipt at Phase 24 revision | None for macOS claim |
| Linux native host lane | EVD-10 Linux receipt | Historical Phase 24 receipt only; current Phase 25 availability not established | Historical Linux/x86_64 receipt at Phase 24 revision | None for Linux claim |

**Missing dependencies with no fallback:** current Phase 25 native host receipts cannot be established from this read-only research pass.

## Validation Architecture

Configuration has the exact setting `"nyquist_validation": true`; include focused validation planning. [VERIFIED: `.planning/config.json:24-27`]

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go `testing` standard library; Go version from project constraint is 1.24 |
| Config file | `go.mod` |
| Quick run command | Set by planner from focused Phase 25 tests; this research did not run it |
| Full suite command | `go test ./...` is a project suite shape; this research did not run it |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|-------------|----------|-----------|-------------------|-------------|
| NAT-11 | Shared C pointer parameter, expected result, incompatible access and escape refusal | checker + emitter/native integration | Planner to bind to focused tests | ❌ Phase 25 witness required |
| NAT-12 | Exclusive C pointer parameter, expected result, conflict and escape refusal | checker + emitter/native integration | Planner to bind to focused tests | ❌ Phase 25 witness required |
| NAT-13 | C and manifest agree; prohibited promises absent; unsupported shapes refuse before serialization | emitter/manifest/source refusal | Planner to bind to focused tests | ❌ Phase 25 controls required |
| EVD-10 | Each family has current native receipts for macOS and Linux and applicable lanes | host integration | Existing evidence script/lane extended by plan | ❌ Phase 25 receipts required |
| DX-14 | Clean-checkout documented utility returns 65/66 and admitted 0x43 failure | end-to-end/documentation | Public build/run commands selected by plan | ❌ integrated utility docs/evidence required |
| DX-15 | Structured source-attributed diagnostics for moved owner, discarded ownership, escape, unsupported cleanup/exit | checker/CLI integration | Existing diagnostic tests plus Phase 25 fixture commands | Partial legacy tests; Phase 25 public examples required |

### Wave 0 Gaps

- [ ] Shared and exclusive `U64` helper source witnesses and refused neighboring shapes.
- [ ] Independent peer mutations for incompatible/conflicting access and escape.
- [ ] Production C pointer parameter/call-site and manifest consistency controls.
- [ ] Family-specific native expected-answer and reached wrong-result controls on both hosts.
- [ ] Clean-checkout documentation/evidence index and stable structured diagnostic examples.

## Security Domain

`security_enforcement` is enabled in project config. OWASP describes ASVS as a basis for testing web application technical security controls; this phase is a local CLI/compiler and native C adapter slice, so the table below records applicability by analogous control concern and does not claim ASVS certification or web-application coverage. Keep untrusted file path/input bounds, local C build inputs, and foreign declarations within explicit trusted boundaries. The pointer helpers operate on a copied scalar and must not expand authority to owner memory, foreign callbacks, retention, or nonlocal exits. [CITED: `.planning/config.json`; Phase 25 context; `.planning/REQUIREMENTS.md`; https://owasp.org/projects/asvs]

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| Authentication / session controls | no | No authentication or session surface in this local utility. |
| Access-control concern | yes | Keep authority within checked source operations and explicit local C bindings; do not infer it from pointer representation. |
| Input-validation concern | yes | Preserve bounded path/file input and typed empty/oversize/use failure contracts. |
| Cryptography concern | no | No cryptographic operation is introduced. |
| Web-only controls (e.g. browser session/CORS) | no | The selected product is a local CLI/native utility, not a web application. |

### Known Threat Patterns for Go/C FFI

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Pointer escape/retention or callback extends lifetime beyond checked loan | Information disclosure / Tampering | Structurally refuse before serialization; add family-specific escape controls. |
| Shared/exclusive conflict hidden by a correct-looking output | Tampering | Independent peer derivation and reached conflicting/incompatible-access controls. |
| Manifest claims stronger alias, capture, or alignment facts than emitted/checked ABI | Tampering | Derive/check manifest against emitted C; forbid unsupported promises. |
| Trusted adapter exceeds documented file bounds or returns invalid resource state | Tampering / Denial of service | Keep adapter audited and bounded; preserve typed error and cleanup behavior. |

## Sources

### Primary (HIGH confidence)

- `.planning/phases/25-separate-pointer-successors-and-integrated-utility/25-CONTEXT.md` — locked scope, helper behavior, utility inputs/results, evidence and refusal boundaries.
- `internal/compiler/check/check.go:4039-4064` — exact current Phase 24 caller admission.
- `internal/compiler/cgen/cgen_program.go:641-645` — production by-pointer refusal and foreign contract boundary.
- `internal/compiler/cgen/cgen.go:245-278,345-397,1080-1099,1281-1323` — legacy exclusive/shared classifiers, restrict metadata, manifest selector, and banned/justifiable attribute policy.
- `internal/compiler/diagnostic/diagnostic.go:97-117` — structured diagnostic fields and stable ID construction.
- `testdata/phase3/sequential_shared_then_exclusive_accept.schway`, `shared_exclusive_reject.schway`, and `testdata/phase5/restrict_borrow.schway` — existing borrow and historical pointer-shape fixtures.
- `.planning/REQUIREMENTS.md`, `.planning/ROADMAP.md`, `.planning/PROJECT.md`, `.planning/STANDING-VERDICTS.md` — requirement ownership, admission obligations, trust/dependency rules.

### Secondary (MEDIUM confidence)

- `.planning/phases/24-ownership-transfer-through-calls-and-errors/24-VERIFICATION.md` and `24-03-SUMMARY.md` — historical Phase 24 host verification and evidence boundary.
- `.planning/research/M004/SUMMARY.md`, `ARCHITECTURE.md`, and `PITFALLS.md` — accepted predecessor synthesis and adversarial evidence guidance.
- `.planning/PRODUCT-ROADMAP.md` and `.planning/LANGUAGE-MATURITY.md` — living current witness/refusal state and ranked successor direction.
- [OWASP ASVS project page](https://owasp.org/projects/asvs) — current standard scope is web application technical security verification; used to avoid overstating ASVS applicability to this local CLI/native phase.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — project and phase constraints explicitly select the existing Go/C17/Clang path; no new package is needed.
- Architecture: HIGH for current refusals and source boundaries; MEDIUM for exact helper grammar and pointer call-site lowering, which remain implementation questions.
- Pitfalls: HIGH — grounded in explicit current refusal paths, legacy `restrict` metadata, locked acceptance controls, and historical dual-host evidence.

**Research date:** 2026-10-01  
**Valid until:** 2026-10-31 for stable source architecture; recheck current source when planning because the tree is actively changing.
