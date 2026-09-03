# Spike Manifest

## Ideas

### ownership-kernel

Turn the leading ownership semantics into a small executable oracle before
building a parser, compiler backend, runtime, or user-facing syntax.

**Requirements:**

- Preserve affine ownership, shared/exclusive loans, explicit transfer,
  deterministic lexical release, and stable machine-readable diagnostics.
- Keep the spike dependency-free and cheap to run locally or in CI.
- Treat disagreement and minimized counterexamples as useful results.
- Do not infer final source syntax or production performance from kernel
  notation.
- Infer local loan endings across control flow without requiring source-level
  end markers on ordinary paths.
- Keep the production-shaped analysis polynomial while retaining bounded path
  expansion only as an independent oracle.
- Treat explicit source loan endings as outside the CFG-last-use slice rather
  than silently mixing manual and inferred policies.
- Export borrowed-result origins as verified value/field paths and tagged
  alternatives, with explicit shared/exclusive access in interface metadata.
- Keep `copy`, `drop`, `share`, `send`, and `escape` independent and derive
  their generic propagation structurally.
- Use higher-ranked fresh origins only at genuinely scoped callback/lending
  boundaries; do not force all borrowed results through callback-only APIs.
- Bind reusable ownership evidence to a canonical versioned typed-core subject.
  Use compact independent recomputation at package, cache, CI, and release
  boundaries; materialize full replay traces only on demand.
- Do not claim that a downstream certificate proves source-to-core correctness.
  A coordinated false typed-core statement remains inside the frontend trust
  boundary and must be attacked by different evidence.
- Treat target layout, initialization, allocator pairing, callback retention,
  address/provenance capture, cleanup, and unwind as explicit typed-core/FFI
  obligations. Derive optimizer attributes only from proven facts.
- Require unoptimized, optimized, semantic-oracle, and sanitizer evidence as
  distinct native lanes. Do not allow panic, unwind, or foreign nonlocal exit
  across an unaudited C ABI boundary.

## Spikes

| # | Idea | Name | Type | Validates | Verdict | Tags |
|---|---|---|---|---|---|---|
| 001 | ownership-kernel | ownership-kernel-workbench | standard | Given small linear, scoped, and branching ownership programs, independently checked and executed, valid cases agree and invalid cases produce stable minimal evidence | VALIDATED | ownership, interpreter, checker, differential |
| 002 | ownership-kernel | cfg-edge-last-use | standard | Given loans used on only some branch or loop paths, backward CFG liveness places path-specific loan ends that agree with bounded per-path normalization without enumerating paths in the analyzer | VALIDATED | ownership, cfg, liveness, property, differential |
| 003 | ownership-kernel | public-origins-generic-abilities | comparison | Given separately compiled APIs that return borrowed or generic values, a body-blind consumer checks verified origin, access, and ability summaries and agrees with an independent implementation oracle | VALIDATED | ownership, origins, generics, abilities, separate-compilation, higher-ranked |
| 004 | ownership-kernel | independent-certificate-checker | comparison | Given canonical typed-core ownership facts and an untrusted certificate, a source- and body-blind checker catches distinct corruption classes without recreating frontend inference or taxing the ordinary edit loop | PARTIAL | ownership, certificates, trusted-computing-base, validation, mutation, provenance |
| 005 | ownership-kernel | native-ffi-provenance-cleanup | standard | Given candidate ownership/resource invariants expressed through a minimal native C path, O0/O3 execution and hostile ABI, alias, provenance, allocator, cleanup, sanitizer, and nonlocal-exit cases expose contract violations | PARTIAL | native, ffi, abi, provenance, cleanup, optimization, sanitizers |
