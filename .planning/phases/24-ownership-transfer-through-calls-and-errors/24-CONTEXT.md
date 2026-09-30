# Phase 24: Ownership Transfer Through Calls and Errors - Context

**Gathered:** 2026-09-30  
**Status:** Ready for planning

<domain>
## Phase Boundary

Deliver a runnable native file-byte program where a helper acquires a live
`FileByteOwner`, transfers it through a Schway call and return, and the
receiving owner uses and releases it exactly once. Also exercise repeated
activations of the same acquisition site followed by a real typed error after
multiple successful acquisitions. On admitted normal and typed-error paths,
non-transferred resources are destroyed in reverse successful-acquisition
completion order. Preserve the Phase 23 bounded adapter and application input
boundary.

This phase owns RES-05, RES-06, OWN-10, OWN-11, OWN-12, and EVD-09. It extends
EVD-11's modeled foreign outcomes without changing the distinction between
model evidence and actual host IO. Defects, process termination, unwind,
cancellation, general owning aggregates, and the shared/exclusive pointer
families remain outside this phase.
</domain>

<decisions>
## Implementation Decisions

These decisions carry forward the M004 roadmap and Phase 21–23 artifacts. They
are not new direct user answers; the prior Phase 22 context records approval to
follow the reviewed M004 recommendations automatically.

### Owner transitions and identity

- **D-24-01:** A move or owning return transfers the live resource identity and
  release obligation to its new owner without running the destructor. A borrow
  leaves ownership and the obligation with the current owner. The final owner
  invokes the declared consuming destructor exactly once.
- **D-24-02:** Acquisitions from separate dynamic activations of one static
  site are distinct semantic resources. Their identities remain stable across
  transfer and must not use raw host addresses as portable IDs.
- **D-24-03:** Copying an owner, using a moved-from owner, and returning an
  owning result from process entry without an external receiver are refused
  before execution with source-attributed diagnostics.

### Typed errors and cleanup

- **D-24-04:** A failed acquisition creates no owner. On admitted normal and
  typed-error paths, release every still-owned resource in reverse successful
  acquisition completion order before returning or propagating the error.
  Preserve the declared infallible consuming destructor contract.
- **D-24-05:** Reuse Phase 23's `0x43` `UseError.UnsupportedByte` outcome as the
  later typed failure after multiple successful acquisitions. The expected
  failure and the reverse destruction sequence must be observable through the
  ordinary application entry. This is a derived fixture choice from the
  existing Phase 23 contract, not a new user answer.
- **D-24-06:** Cleanup guarantees cover admitted normal and typed-error paths.
  Defects and process termination remain outside the guarantee; do not add
  unwind, cancellation, callbacks, or nonlocal transfer behavior.

### Runnable evidence

- **D-24-07:** Keep the positive witness on caller-selected one-byte files:
  `0x41` and `0x42` must still produce independently expected values `65` and
  `66`. A helper acquires and returns the owner; its caller uses and releases
  the returned owner.
- **D-24-08:** Repeated calls to the same helper must create separately
  identified activations. An independent native observer proves allocation,
  use after transfer, and physical destruction before exit; reached omitted,
  premature, duplicate, wrong-resource, and identity-collision controls must
  fail even when compiler events remain plausible.
- **D-24-09:** Extend EVD-11's deterministic foreign-outcome model for the new
  admitted operation shapes. Model results do not claim physical IO or cleanup;
  physical cleanup requires the independent native observer and required host
  evidence.

### the agent's Discretion

- Choose the smallest source syntax, internal activation/resource identity
  representation, and evidence schema that preserve the decisions above and
  existing checked-core contracts.
- Choose how to split the positive transfer and repeated-activation error
  entries while preserving the existing one-path application boundary.
- Extend the affected checker, independent validators, interpreter, native
  emitter, observer, and host CI evidence only as required by the witness.
  Keep `emitProgram` the sole production serializer and fail unsupported shapes
  before C serialization.
- Keep source inspection, newly executed checks, and historical CI receipts
  distinct; do not run project suites locally because STATE explicitly forbids
  that for this checkout.
</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase boundary and requirements

- `.planning/ROADMAP.md` § M004 Goal and § Phase 24 — boundary, requirements,
  success criteria, runnable witness, and admission obligations.
- `.planning/REQUIREMENTS.md` § Live local resource ownership, § Ownership
  across Schway calls, § Decisive native evidence, and § Constraints — RES-05/06,
  OWN-10/11/12, EVD-09, and the EVD-11 extension boundary.
- `.planning/PROJECT.md` § Constraints and § Verification Operating Preference
  — runtime, trust, portability, and evidence expectations.
- `.planning/PRODUCT-ROADMAP.md` § Current three recommendations — capability
  order and next-phase dependencies.
- `.planning/LANGUAGE-MATURITY.md` § What exists, § What prevents ordinary
  programs, and § Next useful thresholds — current admitted behavior and
  refusal boundaries.

### Accepted M004 design and predecessor decisions

- `.planning/research/M004/SUMMARY.md` — accepted milestone synthesis.
- `.planning/research/M004/ARCHITECTURE.md` § Successor resource contract and
  § Runtime evidence that earns the claim — ownership transitions, identity,
  cleanup, and independent observation.
- `.planning/research/M004/FEATURES.md` — bounded runnable witness and
  application behavior.
- `.planning/research/M004/PITFALLS.md` — reached mutation controls, host
  evidence, and CI cost/ownership rules.
- `.planning/milestones/M004-phases/21-native-emission-ownership-and-resource-discharge-m004/21-CONTEXT.md`
  and `21-RESOURCE-DISCHARGE-CONTRACT.json` — archived contract and its
  corrected release-versus-transfer wording.
- `.planning/phases/22-native-application-build-and-single-execution/22-CONTEXT.md`
  — preserved app input/run boundary and approved M004 implementation discretion.
- `.planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md` and
  `23-VERIFICATION.md` — live local-owner contract, `0x43` typed failure,
  independent observer, and Phase 23 evidence status.

### Current implementation and evidence seams

- `internal/compiler/check/check.go` — source admission and current local
  resource-lifecycle lowering.
- `internal/compiler/core/core.go` — operation, place, invocation, and
  ownership facts consumed by independent peers.
- `internal/compiler/corevalidate/corevalidate.go` — acquisition-derived
  obligations and local path validation.
- `internal/compiler/originvalidate/originvalidate.go` — independently
  re-derived source-origin facts.
- `internal/compiler/pathoracle/pathoracle.go` — bounded path and local-owner
  obligation validation.
- `internal/compiler/interp/interp.go` — dynamic call frames, modeled foreign
  outcomes, and current per-frame resource tracking.
- `internal/compiler/cgen/cgen_program.go` — sole production serializer and
  current multi-function foreign/local-owner refusal boundary.
- `internal/compiler/native/phase23_observer_test.go` and
  `internal/compiler/native/native_app_test.go` — independent physical observer
  and current public app lifecycle witnesses to extend.
- `examples/phase23/file_byte.schway`, `examples/phase23/adapter.c`,
  `examples/phase23/README.md`, and `scripts/verify-phase23.sh` — existing
  source, adapter, public contract, and recurring focused evidence.
- `.github/workflows/ci.yml` — existing macOS and Linux lanes; avoid duplicate
  expensive suite ownership.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- `examples/phase23/file_byte.schway` and its adapter already express the
  bounded acquire/use/release foreign contract, including the typed
  `UnsupportedByte` use outcome.
- The native Phase 23 observer and app tests establish actual allocation,
  borrowed use, physical release, and reached cleanup mutation controls.
- The interpreter already has a dynamic call stack, caller return targets, and
  deterministic modeled foreign outcomes.

### Established Patterns

- `check`, `corevalidate`, `originvalidate`, `pathoracle`, `interp`, and `cgen`
  have separate responsibilities; validators must re-derive ownership facts
  instead of trusting checker-produced success or release lists.
- Current cleanup reasoning is local to an activation. `corevalidate` and
  `pathoracle` seed Phase 23 obligations from acquisitions, while interpreter
  ledgers are tracked per frame using static operation IDs.
- `emitProgram` is the single native emission authority. It currently refuses
  multi-function foreign-call bodies/contracts and only accepts the exact
  Phase 23 local-owner shape; unsupported candidates must remain refused
  before serialization.
- Native event reports are not physical proof. The independent observer must
  continue to establish actual destruction and outstanding-resource absence.

### Integration Points

- Carry ownership and dynamic resource identity across `OpCall`, `OpMove`,
  `OpReturn`, and typed `OpFail` paths in the checked core and all affected
  independent validators.
- Update the interpreter's frame/ledger boundary and native call/cleanup
  lowering without allowing static acquisition IDs to alias across activations.
- Extend the public Phase 23 app witness and its independent observer through
  the existing session/native binding path; retain ordinary one-run semantics.
- Extend existing macOS/Linux CI lanes with the distinct evidence needed for
  transfer, repeated activation, and typed-error destruction order.

</code_context>

<specifics>
## Specific Ideas

- Use the existing file-byte behavior: `0x41` → `65`, `0x42` → `66`, and
  `0x43` → `UseError.UnsupportedByte` after successful acquisition.
- A helper-acquired owner crosses its return boundary without destruction;
  the receiving function performs the use and consuming release.
- A separate error witness reaches the typed failure only after multiple
  successful acquisitions from repeated activations of the same acquisition
  site, then proves reverse completion-order destruction before error exit.
- These are inherited and derived fixture constraints, not new direct user
  answers. The prior M004 authorization leaves source syntax and internal
  proof representation to research and planning.
</specifics>

<deferred>
## Deferred Ideas

- Separate shared/exclusive read-copy pointer families and the integrated
  utility remain Phase 25.
- U64 arithmetic, loops, `sum_to_n`, and FizzBuzz remain next-milestone work.
- General owning aggregates, fallible implicit destructors, unwind,
  cancellation, callbacks, escaped pointers, and broader foreign-call shapes
  remain refused unless a later phase explicitly owns them.

</deferred>

---

*Phase: 24-ownership-transfer-through-calls-and-errors*  
*Context gathered: 2026-09-30*
