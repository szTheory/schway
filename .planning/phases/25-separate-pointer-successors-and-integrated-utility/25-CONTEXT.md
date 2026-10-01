# Phase 25: Separate Pointer Successors and Integrated Utility - Context

**Gathered:** 2026-10-01
**Status:** Ready for planning
**Provenance:** The user adopted all four recommendations after a three-part
specialist fan-out across product/DX, ownership and compiler semantics,
security/FFI, validation, diagnostics, and CI. Choices below are accepted
implementation direction; source feasibility observations are not execution
evidence.

<domain>
## Phase Boundary

Deliver a documented, reproducible native file-byte utility that uses separate
bounded shared and exclusive read/copy helpers. Each helper is a Schway source
consumer lowered by the sole production emitter to an actual C pointer
parameter, and each has its own positive result, independent borrow checks,
conflict/escape controls, and macOS/Linux evidence. The utility composes the
existing file-owner transfer and borrowed-use path with both helpers and
returns their copied value. Preserve 0x41 → 65, 0x42 → 66, and the admitted
0x43 typed use failure and cleanup path.

The pointer helpers operate on the file-derived U64 value after the foreign
read. They do not point into or expose the FileByteOwner allocation. The C
adapter remains the trusted owner of its file descriptor; the existing Schway
resource contract remains responsible for the allocated buffer's ownership,
transfer, and discharge.

This phase owns NAT-11/12/13, EVD-10, and DX-14/15. Unsupported mutation,
forwarding, retention, callbacks, nonlocal exits, and wider pointer shapes
remain structurally refused before C serialization.
</domain>

<decisions>
## Implementation Decisions

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

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase boundary, requirements, and living direction

- .planning/ROADMAP.md § Phase 25, § NAT-09 Successor Ownership, and §
  Delivery Rules and Future Direction — phase goal, exact admitted families,
  refusals, runnable gain, and evidence boundaries.
- .planning/REQUIREMENTS.md § Bounded pointer successors, § Decisive native
  evidence, § Developer usability, and § Scope and acceptance rules — NAT-11,
  NAT-12, NAT-13, EVD-10, DX-14, DX-15 and acceptance constraints.
- .planning/PROJECT.md § Core Value, § Verification Operating Preference,
  and § Constraints — user value, evidence policy, portability, trust, and
  dependency posture.
- .planning/PRODUCT-ROADMAP.md § Checker and guarantee activation map and §
  Current three recommendations — affected consumers and ranked follow-on
  capabilities.
- .planning/LANGUAGE-MATURITY.md § What exists, § What prevents ordinary
  programs, and § Next useful thresholds — current source witness and refusal
  boundaries, and current evidence provenance.
- .planning/STANDING-VERDICTS.md — standard-library/dependency preference,
  independent derivation rules, and known process pitfalls.

### M004 and completed predecessor decisions

- .planning/research/M004/SUMMARY.md § Phase 25 and § Critical Pitfalls —
  accepted phase intent and separate pointer-family proof obligations.
- .planning/research/M004/ARCHITECTURE.md § Conservative pointer admission
  and § Runtime evidence that earns the claim — pointer shape and evidence
  boundaries.
- .planning/research/M004/PITFALLS.md § Shared and exclusive pointer proofs
  are conflated and § Evidence cadence and stopping rules — adversarial controls
  and CI ownership.
- .planning/phases/22-native-application-build-and-single-execution/22-CONTEXT.md
  — retained app execution, input, C build authority, and evidence/replay split.
- .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md and
  23-VERIFICATION.md — file-byte adapter, allocation ownership, typed use
  failure, observer boundary, and recorded receipt status.
- .planning/phases/24-ownership-transfer-through-calls-and-errors/24-CONTEXT.md,
  24-VERIFICATION.md, and 24-03-SUMMARY.md — transfer, activation identity,
  reverse cleanup, and revision-scoped historical receipts.

### Existing source and evidence seams

- examples/phase24/transfer.schway and examples/phase24/README.md — current
  transfer/use source path and clean-checkout application commands.
- examples/phase23/file_byte.schway, adapter.c, adapter.h,
  file_byte.bindings.json, and scripts/verify-phase23.sh — explicit trusted
  C operations, build manifest, and existing resource evidence.
- internal/compiler/syntax/parser.go — current function parameter, borrow,
  and linear-body syntax.
- internal/compiler/check/check.go — bounded source admission, including
  checkLocalFileByteTransferCaller and the fixed current caller shape.
- internal/compiler/core/core.go — checked operation and ownership facts.
- internal/compiler/corevalidate/corevalidate.go,
  internal/compiler/originvalidate/originvalidate.go, and
  internal/compiler/pathoracle/pathoracle.go — independently derived facts
  that must remain independent.
- internal/compiler/interp/interp.go — deterministic foreign outcome model;
  model success does not establish host IO or physical cleanup.
- internal/compiler/cgen/cgen.go and
  internal/compiler/cgen/cgen_program.go — legacy family classifiers and
  sole production serializer, which still refuses by-pointer bodies.
- internal/compiler/cgen/cgen_program_test.go and
  internal/compiler/cgen/cgen_test.go — current refusal and legacy metadata
  controls; legacy restrict metadata is not a Phase 25 promise.
- testdata/phase3/shared_shared_accept.schway,
  testdata/phase3/sequential_shared_then_exclusive_accept.schway,
  testdata/phase3/shared_exclusive_reject.schway, and
  testdata/phase5/restrict_borrow.schway — established shared/exclusive loan
  semantics and prior generated-pointer shape.
- internal/compiler/diagnostic/diagnostic.go and
  internal/compiler/session/session_borrow_conflict_test.go — structured
  diagnostics, stable schema behavior, and borrow diagnostic examples.
- internal/compiler/native/phase24_observer_test.go,
  scripts/verify-phase24.sh, and .github/workflows/ci.yml — physical owner
  observer limits, hosted evidence entry point, and existing host lanes.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- Phase 24's public application, helper-returned owner, typed-error entry, and
  Phase 23 binding manifest form the retained file-byte path.
- Existing shared/exclusive borrow fixtures establish overlapping conflict
  and last-use loan-end behavior. The historical C golden shows how a checked
  sole parameter became a C pointer; it is not current production admission.
- Existing structured diagnostics carry codes, source spans, causes, and
  repairs; no new wire schema is required for these boundary explanations.
- Existing native per-family and full macOS/Linux lanes can host the new
  evidence with one owner per expensive lane.

### Established Patterns

- check, corevalidate, originvalidate, pathoracle, interp, and cgen have
  separate responsibilities. Validators independently re-derive affected
  borrow facts rather than importing checker approval.
- emitProgram is the sole production serializer and currently refuses both
  by-pointer body families before C serialization.
- C binding declarations are trusted build inputs. Pointer lowering does not
  itself authorize restrict, noalias, capture, or alignment claims.
- The Phase 24 observer follows the allocated owner's actual pointer through
  use and release. It does not prove pointer-parameter behavior for a copied
  scalar value.

### Integration Points

- Extend the source-consumer and fixed owner-transfer caller shapes together,
  then independently update validation, interpreter, serializer, manifests,
  and source-attributed refusals as the same bounded witness requires.
- Extend the Phase 24 application/result path with two copy-helper calls while
  preserving single-launch app semantics and the existing typed error.
- Add native evidence for both distinct pointer families and connect its
  source/build/compiler/target/flags, expected results, and applicable lanes to
  the clean-checkout evidence index.

</code_context>

<specifics>
## Specific Ideas

- The intended utility path is acquire → transfer → use → shared read/copy →
  exclusive read/copy → return. Each helper result is consumed by the next.
- Both helpers read/copy the derived U64; the Phase 23 malloc-backed buffer
  remains governed by the existing FileByteOwner lifecycle.
- Preserve supplied one-byte inputs and expected results: 0x41 → 65,
  0x42 → 66; 0x43 reaches the existing UseError.UnsupportedByte after
  acquisition and exercises cleanup before either copy helper.
- Use reached wrong-result controls for each helper, plus family-specific
  conflict and escape controls. Keep C pointer-parameter inspection and the
  attribute manifest family-specific. Do not mistake sanitizer or engine
  agreement for independent borrow validation.
- The research pass inspected source and historical Phase 22–24 records; it
  ran no project tests, builds, or CI. Record future executed checks and hosted
  receipts at their actual revisions.

</specifics>

<deferred>
## Deferred Ideas

- Direct pointer access to fields or bytes inside the owned buffer, owner
  projection, and general buffer views remain outside this bounded scalar
  witness.
- Mutation, forwarding, retention, callbacks, nonlocal exits, and wider pointer
  forms remain refused.
- Arithmetic, loops, FizzBuzz, general strings/arrays, JSON, and a reusable
  library/module system belong to later consumer-led work.
- One bounded JSON configuration consumer remains future work after a named
  program justifies it.

</deferred>

---

*Phase: 25-separate-pointer-successors-and-integrated-utility*  
*Context gathered: 2026-10-01*
