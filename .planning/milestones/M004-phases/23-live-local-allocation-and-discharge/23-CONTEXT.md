# Phase 23: Live Local Allocation and Discharge — Context

**Gathered:** 2026-09-27  
**Status:** Ready for planning

<domain>
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

</domain>

<decisions>
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

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase boundary and acceptance

- `.planning/ROADMAP.md` § Phase 23 — runnable witness, admission obligations,
  phase ordering, and Phase 24/25 boundaries.
- `.planning/REQUIREMENTS.md` § FFI-03, RES-04, RES-07–09, EVD-09–11, DX-14–15
  — exact owned requirements and their negative controls.
- `.planning/PROJECT.md` § Verification Operating Preference — objective
  shift-left evidence and when checks belong in recurring CI.
- `.planning/PRODUCT-ROADMAP.md` § Current three recommendations — current
  capability order and the next source/checker obligations.
- `.planning/LANGUAGE-MATURITY.md` — implemented behavior, current refusals,
  named test anchors, and portability limits.

### M004 architecture and evidence

- `.planning/research/M004/SUMMARY.md` — accepted milestone synthesis and
  phase-level evidence boundaries.
- `.planning/research/M004/ARCHITECTURE.md` — per-operation foreign contracts,
  acquisition-derived obligations, ownership transitions, and the independent
  native observer.
- `.planning/research/M004/FEATURES.md` — bounded path input, actual file-byte
  use, independent expected result, and public application behavior.
- `.planning/research/M004/STACK.md` — retained Go/C17/Clang path, bounded path
  input, adapter ownership of file descriptors, and byte-level scope.
- `.planning/research/M004/PITFALLS.md` — real-storage proof, reached mutations,
  host evidence, and CI cost/ownership rules.
- `.planning/research/M004/KICKOFF-VERIFICATION.md` — what kickoff checks
  established and what they did not execute.
- `.planning/milestones/M004-phases/21-native-emission-ownership-and-resource-discharge-m004/21-CONTEXT.md`
  — archived ownership and foreign-exit decisions.
- `.planning/milestones/M004-phases/21-native-emission-ownership-and-resource-discharge-m004/21-RESOURCE-DISCHARGE-CONTRACT.json`
  — contract-only predecessor that the new version explicitly corrects.
- `.planning/phases/22-native-application-build-and-single-execution/22-CONTEXT.md`
  — application boundary and preserved U64 path.
- `.planning/phases/22-native-application-build-and-single-execution/22-VERIFICATION.md`
  — Phase 22 verified scope and host limitations.

### Current implementation and recurring checks

- `internal/compiler/check/check.go` — current foreign-call/resource lifecycle
  admission and per-operation limitations.
- `internal/compiler/corevalidate/corevalidate.go`,
  `internal/compiler/originvalidate/originvalidate.go`, and
  `internal/compiler/pathoracle/pathoracle.go` — independent facts that must be
  re-derived when acquisition creates ownership obligations.
- `internal/compiler/interp/interp.go` — modeled foreign outcomes and resource
  behavior; model results are not physical host IO or cleanup evidence.
- `internal/compiler/cgen/cgen_program.go` — single native emission authority
  and existing foreign/by-pointer refusals.
- `internal/compiler/native/native_app.go` and
  `internal/compiler/native/bindings.go` — retained app execution and explicit
  local build inputs to extend without fixture-name authority.
- `internal/compiler/native/foreign_resource.go` and
  `native/lang_foreign_resource.c` — historical fixture tracer; it frees
  internally and is not a live-Lang-owner witness.
- `cmd/lang/main.go` and `examples/phase22/README.md` — public app command and
  documented contract to extend while preserving the U64 route.
- `.github/workflows/ci.yml` — existing host lanes; avoid duplicating expensive
  checks without a separate evidence question.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- `native.BuildApplication`, `native.RunApplication`, and the closed binding
  manifest in `internal/compiler/native/` provide the retained build, one-run,
  and explicit local-C seams.
- The Go test suite already has deadline-bounded subprocess helpers, negative
  controls, and generated-C integration tests suitable for new reached cleanup
  mutations.
- The current `foreign_resource` fixture and its ledger tests show the frozen
  historical path. They free inside C and must not be mistaken for the new live
  ownership behavior.

### Established Patterns

- `check`, `corevalidate`, `originvalidate`, `pathoracle`, `interp`, and `cgen`
  independently derive affected semantic facts; do not make peers import a
  producer's conclusion.
- `emitProgram` refuses foreign contracts and by-pointer bodies before C
  serialization. Keep it the sole emission authority and retain refusals for
  unsupported ownership shapes.
- Current native application entry/run code accepts bounded U64 input. Add the
  narrow path input while keeping that route stable.
- The interpreter's foreign-call behavior is a model, while physical malloc,
  use, and free require a separate native observer.

### Integration Points

- `cmd/lang/main.go` → `session` → `native` is the public build/run path for
  path input and explicit adapter bindings.
- `check.go` and core facts feed all independent validators, `interp`, and the
  shared C emitter; acquisition, use, and release each need their own binding
  and failure/ownership facts.
- Existing macOS and Linux CI host lanes should run the stable acceptance path
  and the exact-shape native evidence required for this foreign family.

</code_context>

<specifics>
## Specific Ideas

- The public witness uses two separate one-byte files: `0x41` prints `65`, and
  `0x42` prints `66`.
- An empty file and a file longer than one byte fail during acquisition; a
  one-byte `0x43` file reaches the separate use operation and triggers the
  post-acquisition typed-error cleanup path.
- The user authorized automatic follow-through and wants objective verification
  to be the default. These decisions were selected in `--auto` mode from the
  roadmap and project preference; they are not recorded as new direct user
  answers.

</specifics>

<deferred>
## Deferred Ideas

- Ownership transfer through calls and typed errors belongs to Phase 24.
- Shared/exclusive read-copy pointer families and integrated utility closure
  belong to Phase 25.
- General strings, arrays, loops, arithmetic, owning aggregates, fallible
  destructors, unwind, cancellation, and broader FFI remain outside M004 scope.

</deferred>

---

*Phase: 23-live-local-allocation-and-discharge*  
*Context gathered: 2026-09-27*
