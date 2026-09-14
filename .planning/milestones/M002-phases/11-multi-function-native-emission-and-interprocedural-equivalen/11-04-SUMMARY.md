---
phase: 11-multi-function-native-emission-and-interprocedural-equivalen
plan: 04
subsystem: cgen
tags: [cgen, session, native, mid-phase-gate, attributes, mutation-kill]

# Dependency graph
requires:
  - phase: 11-multi-function-native-emission-and-interprocedural-equivalen
    provides: "11-03's callgraph.EntryFunction, cgen.emitProgram/emitCall whole-program assembler, and the testdata/phase11 corpus this plan's gate corpus extends"
provides:
  - "cgen_program.go's emitCallBoundaryAttributeComment: a generated, deterministic, D-11-09-citing comment naming the call-boundary attribute set as explicitly empty (NAT-05)"
  - "cgen.AttributeSuppressionProfile (AttributesJustified/AttributesSuppressed) and byPointerQualifier: a permanent, token-local suppression control for the one restrict qualifier this package emits, defaulting to unchanged pre-Phase-11 behavior"
  - "session.VerifyPhase11ZeroAttributeGate: Phase 11's mandatory mid-phase gate, a four-part conjunction with two independent knowers, proven non-vacuous at both sides of the N threshold and mutation-killed"
  - "testdata/phase11/multi_function_gate_corpus.lang: the engineered non-vacuous gate corpus (3 functions, 1 call edge, 1 would-have-carried-restrict function)"
  - "11-MIDPHASE-GATE.md: the ratified adjudication admitting waves 4-6"
affects: [11-05, 11-06, 11-07, 11-08, 11-09]

# Actuals (#2632)
actuals:
  tokens: 42000
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Named-derivation-not-literal comments (D-04-12 applied to D-11-09): a package-level derivation function feeds both the emitted comment and any future gate that asks the same question, with a test-only override seam proving the two never drift."
    - "Permanent, token-local attribute suppression profile (D-11-16/D-11-17): one Fprintf call site, one qualifier helper, a package-level profile whose zero value preserves all pre-existing behavior -- used as a mid-phase gate's own bisection tool, not a temporary flag."
    - "Two-knower gate conjunction: one knower reads emitted OUTPUT BYTES (cgen.ScanForBannedAttributes), the other independently re-derives the same structural fact from the PROGRAM in a different package, sharing zero helpers (D-12)."

key-files:
  created:
    - internal/compiler/session/session_phase11_gate.go
    - internal/compiler/session/session_phase11_gate_test.go
    - testdata/phase11/multi_function_gate_corpus.lang
    - .planning/phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-MIDPHASE-GATE.md
  modified:
    - internal/compiler/cgen/cgen.go
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/cgen/cgen_program_test.go
    - internal/compiler/cgen/export_test.go
    - internal/compiler/session/session_phase6_pin_test.go

key-decisions:
  - "The mid-phase gate's Knower A (cgen.ScanForBannedAttributes) scans the whole-program TU plus one counterfactual single-function artifact per would-carry function, synthesized under the gate's own profile parameter -- not the lang.foreign/0 EmitForeignManifest sidecar, which is an unrelated, unmodified Phase 4/5 feature that would unconditionally report a populated restrict entry for any by-pointer-eligible function regardless of any new profile, making it unusable as this gate's own zero-attribute evidence."
  - "AttributeSuppressionProfile's zero value is AttributesJustified (pre-Phase-11 behavior unchanged) rather than AttributesSuppressed, so no existing Phase 4/5/6+ caller's output changes unless it explicitly opts in via SetAttributeSuppressionProfileForTest -- the mid-phase gate's own default invocation always passes AttributesSuppressed explicitly."
  - "Fixed a real -Werror -Wunused-function bug in emitProgram (Rule 1): a declared-but-never-called function of a parameter type (Buffer/Byte) different from the resolved entry's own type left that type's writer helper (lang_write_buffer_hex/lang_write_byte) unreferenced. The gate corpus's own touch function is what first exercises this combination."

requirements-completed: [NAT-04, NAT-05]

coverage:
  - id: D1
    description: "emitProgram writes a generated, deterministic comment naming the whole-program call-boundary attribute set as explicitly EMPTY and citing D-11-09, derived from a named function rather than a hand-written literal"
    requirement: "NAT-05"
    verification:
      - kind: unit
        ref: "internal/compiler/cgen/cgen_program_test.go#TestEmittedAttributeSetIsExplicitlyEmpty"
        status: pass
      - kind: unit
        ref: "internal/compiler/cgen/cgen_program_test.go#TestEmittedAttributeSetCommentIsDerivedNotLiteral"
        status: pass
      - kind: unit
        ref: "internal/compiler/cgen/cgen_program_test.go#TestEmittedAttributeSetOrderIsStable"
        status: pass
      - kind: unit
        ref: "internal/compiler/cgen/cgen_program_test.go#TestCallSitesEmitNoArithmeticConversion"
        status: pass
    human_judgment: false
  - id: D2
    description: "The mid-phase gate is a four-part conjunction with two independent knowers, proven non-vacuous at both N=0 (fails) and N=2 (counts adjacently without merging), and mutation-killed under the justified profile"
    requirement: "NAT-05"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase11_gate_test.go#TestPhase11ZeroAttributeGate"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase11_gate_test.go#TestPhase11GateIsNonVacuous"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase11_gate_test.go#TestPhase11GateFailsAtNZero"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase11_gate_test.go#TestPhase11GateCountsAdjacentWouldCarryFunctions"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase11_gate_test.go#TestPhase11GateMutationKill"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase11_gate_test.go#TestPhase11SuppressionIsDiffLocal"
        status: pass
    human_judgment: false
  - id: D3
    description: "The mid-phase gate is adjudicated in writing with all seven D-11-19 items recorded as concrete values; option A (gate PASSED) is ratified and waves 4-6 are admitted, with NAT-05 written up as a requirement weakened by evidence"
    requirement: "NAT-04"
    verification: []
    human_judgment: true
    rationale: "This is a checkpoint:decision task with a one-way reversibility rating. The executor recorded a data-driven ratification from the concrete, reproducible test-suite values in 11-MIDPHASE-GATE.md (auto-mode, no gate=\"blocking-human\" attribute), but the document itself asks a human reviewer to confirm the ratification and the flagged CLI-check divergence before treating them as fully settled."

duration: 95min
completed: 2026-09-12
status: complete
---

# Phase 11 Plan 4: Mid-Phase Gate -- Explicit Empty Attribute Set and a Non-Vacuous, Mutation-Killed Conjunction Summary

**NAT-05 ships as an explicit, generated, deterministic empty call-boundary attribute set citing D-11-09; the mandatory mid-phase gate is a four-part conjunction with two independent knowers, proven non-vacuous at both sides of the N threshold and mutation-killed, and is ratified PASS in writing, admitting waves 4-6.**

## Performance

- **Duration:** ~95 min
- **Started:** 2026-09-12
- **Completed:** 2026-09-12
- **Tasks:** 3 completed
- **Files modified:** 5 modified, 4 created

## Accomplishments

- **`emitCallBoundaryAttributeComment`** (`cgen_program.go`) writes a generated, byte-stable comment into every `emitProgram` artifact naming the call-boundary alias/capture attribute set as explicitly EMPTY and citing `D-11-09` by identifier -- proven to track its own derivation (not a hand-written literal) via a test-only injection seam, and byte-identical across re-emission and declaration-order permutation.
- **`cgen.AttributeSuppressionProfile`** (`AttributesJustified`/`AttributesSuppressed`) and `byPointerQualifier` make the single existing `restrict` write site (`emitLinearBorrowedByPointer`) token-local and permanently toggleable, with the zero value preserving every pre-existing Phase 4/5/6+ caller's behavior exactly. This is the mid-phase gate's own bisection tool (D-11-18), not a temporary flag.
- **`session.VerifyPhase11ZeroAttributeGate`** is the mid-phase gate itself: Knower A (`cgen.ScanForBannedAttributes` over the whole-program TU plus a per-would-carry-function counterfactual artifact) and Knower B (`phase11WouldCarryRestrict`, session's own independent re-derivation of `selectsByPointerLowering`'s structural condition, sharing zero helpers with `cgen`), plus structural floors and a real interpreter/`-O0`/`-O3` agreement run via `RunNative`.
- **`testdata/phase11/multi_function_gate_corpus.lang`** is the committed, engineered non-vacuous corpus: 3 functions, 1 call edge, and 1 would-have-carried-restrict function (`touch`, modeled on `testdata/phase5/restrict_borrow.lang`'s own exclusive-borrow-then-reborrow-to-terminator shape) that is declared but never called.
- **Both sides of the N threshold are asserted**: `TestPhase11GateFailsAtNZero` (N=0, the gate FAILS) and `TestPhase11GateCountsAdjacentWouldCarryFunctions` (N=2, adjacency does not merge or double-count). `TestPhase11GateMutationKill` and `TestPhase11SuppressionIsDiffLocal` prove the scan is not vacuous and the suppression is diff-local.
- **`11-MIDPHASE-GATE.md`** records all seven D-11-19 adjudication items as concrete values and ratifies option A: gate PASSED, waves 4-6 admitted, zero-attribute state terminal, NAT-05 written up as a requirement weakened by evidence.
- **A real `-Werror -Wunused-function` bug** in `emitProgram` (Rule 1) was found and fixed: `lang_write_buffer_hex`/`lang_write_byte` were referenced only when the resolved ENTRY function's own type needed them, leaving the writer for a declared-but-never-called function of a *different* parameter type genuinely unreferenced -- exactly the gate corpus's own `touch` function.

## Task Commits

1. **Task 1: NAT-05's explicit empty attribute set, generated not hand-written** -- `2d86416` (feat)
2. **Task 2: The mid-phase gate lane -- four-part conjunction, two knowers, mutation-killed** -- `c5a7f8c` (feat)
3. **Task 3: Mid-phase gate adjudication** -- `ab8d1af` (docs)

## Files Created/Modified

- `internal/compiler/cgen/cgen_program.go` -- `emitCallBoundaryAttributeSet`/`emitCallBoundaryAttributeComment`, the unused-function fix for `lang_write_buffer_hex`/`lang_write_byte`
- `internal/compiler/cgen/cgen_program_test.go` -- 4 Task 1 tests
- `internal/compiler/cgen/cgen.go` -- `AttributeSuppressionProfile`, `SetAttributeSuppressionProfileForTest`, `byPointerQualifier`, the parametrized `restrict` Fprintf site
- `internal/compiler/cgen/export_test.go` -- `SetCallBoundaryAttributeSetForTest`
- `internal/compiler/session/session_phase11_gate.go` -- `VerifyPhase11ZeroAttributeGate`, `Phase11GateReport`, `phase11WouldCarryRestrict`
- `internal/compiler/session/session_phase11_gate_test.go` -- 6 Task 2 tests
- `internal/compiler/session/session_phase6_pin_test.go` -- pinned lane-schema-literal-site count updated for the new file (Rule 3)
- `testdata/phase11/multi_function_gate_corpus.lang` -- the gate corpus
- `.planning/phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-MIDPHASE-GATE.md` -- the written adjudication

## Decisions Made

See `key-decisions` above. The two load-bearing ones: (1) Knower A scans a synthesized counterfactual single-function artifact per would-carry function rather than the pre-existing `EmitForeignManifest` sidecar, since that sidecar is an unrelated, unmodified Phase 4/5 feature that would always report a populated `restrict` entry regardless of any new profile; and (2) `AttributeSuppressionProfile`'s zero value preserves existing behavior, so the suppression control is additive and never a silent regression to Phase 4/5/6+.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] `emitProgram`'s writer-helper unused-function bug on a declared-but-never-called function of a different parameter type**
- **Found during:** Task 2, first `session.RunNative` run of the gate corpus
- **Issue:** `main` only takes `(void)`-references of `lang_write_buffer_hex`/`lang_write_byte` implicitly via the entry function's own output-writing call; a Buffer-typed function declared but never called, in a program whose entry is Byte-typed, left `lang_write_buffer_hex` genuinely unreferenced, failing `-Werror -Wunused-function`.
- **Fix:** Added explicit `(void)lang_write_buffer_hex;`/`(void)lang_write_byte;` references in `main`, mirroring the existing per-function address-taking pattern (11-03's own fix for declared-but-never-called user functions).
- **Files modified:** `internal/compiler/cgen/cgen_program.go`
- **Verification:** `TestPhase11ZeroAttributeGate` and siblings compile/link/run the gate corpus cleanly at both `-O0` and `-O3`.
- **Committed in:** `c5a7f8c`

**2. [Rule 3 - Blocking issue] Pinned lane-schema-literal-site count test needed updating**
- **Found during:** Task 2, full-suite verification after adding `session_phase11_gate.go`
- **Issue:** `TestLaneSchemaLiteralSiteCountIsPinned` pins the exact per-file count of every `protocol.LaneSchema1` literal reference in the `session` package; the new file's one reference moved the total from 17 to 18 without a coordinated update.
- **Fix:** Added `"session_phase11_gate.go": 1` to the pinned map and bumped the expected total to 18.
- **Files modified:** `internal/compiler/session/session_phase6_pin_test.go`
- **Verification:** `TestLaneSchemaLiteralSiteCountIsPinned` passes; full `go test ./...` green.
- **Committed in:** `c5a7f8c`

### Architectural Note (Rule 4-adjacent -- flagged, not silently resolved)

**3. `go run ./cmd/lang --json check` diverges from `session.Check` on the gate corpus, for a pre-existing reason this plan's corpus is the first to expose**

The plan's own Task 2 `<verify>` block requires `go run ./cmd/lang --json check testdata/phase11/multi_function_gate_corpus.lang` to report a clean, valid result. It does not: the CLI `check` command's `publishedOriginProblemFile` additionally consults `originvalidate.ValidatePublished`, which unconditionally flags ANY function (exported or not) whose body derives a borrow-based return origin without a declared `PublicOrigin` -- exactly the structural shape `selectsByPointerLowering` (and this plan's own `phase11WouldCarryRestrict` re-derivation) REQUIRES `PublicOrigin == nil` for. The two invariants are mutually exclusive for any would-carry-restrict function: satisfying one gate's own structural condition necessarily fails the other's. This is not new: running the identical CLI command against `testdata/phase5/restrict_borrow.lang` (the already-shipped fixture `touch`'s body is modeled on verbatim) reproduces the identical `core.origin_omitted` diagnostic -- no prior test or production path in this codebase exercises a `restrict_borrow.lang`-shaped function through the CLI `check` command; `session.Check` (which never consults `originvalidate`) is what every existing Phase 5/6+ path, and this plan's own gate, actually use. All of this plan's own verification runs through `session.Check`/`corevalidate.Validate`, where the corpus checks and validates cleanly. **Flagged for human review** in `11-MIDPHASE-GATE.md`, matching 11-03-SUMMARY.md's own precedent of flagging a structural divergence rather than silently working around it; it does not affect the gate's own verdict.

---

**Total deviations:** 2 auto-fixed (1 Rule 1, 1 Rule 3), 1 architectural note flagged for human review.
**Impact on plan:** None on correctness or scope. The unused-function fix is required for the gate corpus to compile at all; the pinned-count update is mechanical bookkeeping; the CLI-check divergence is pre-existing and does not affect the gate's own verdict, which never depends on the CLI `check` command.

## Issues Encountered

None beyond the two auto-fixes and one architectural note above -- all discovered during normal end-to-end verification, not left as open questions.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `session.VerifyPhase11ZeroAttributeGate` and its corpus are now the mid-phase gate's own committed artifact; any future re-run should reuse `testdata/phase11/multi_function_gate_corpus.lang` rather than re-engineering a new non-vacuous corpus.
- `cgen.AttributeSuppressionProfile` is now a real, permanent, production export -- future work on Phase 11's own attribute story (or any later phase revisiting call-boundary attributes) should extend this control, not introduce a second one.
- Waves 4-6 (NAT-06, QLT-03, QLT-06, QLT-05) are admitted per the ratified adjudication; `11-VERIFICATION.md` (whenever written) must record NAT-05 as a requirement weakened by evidence, not a clean pass.
- Human review recommended for the flagged CLI-check divergence (Deviation 3) before treating `selectsByPointerLowering`'s `PublicOrigin == nil` gate and `originvalidate.ValidatePublished`'s own unconditional scope as fully reconciled.
- No blockers to proceeding with wave 4.

---
*Phase: 11-multi-function-native-emission-and-interprocedural-equivalen*
*Completed: 2026-09-12*

## Self-Check: PASSED

All created files verified present on disk (`internal/compiler/session/session_phase11_gate.go`,
`session_phase11_gate_test.go`, `testdata/phase11/multi_function_gate_corpus.lang`,
`11-MIDPHASE-GATE.md`). All three task commit hashes (`2d86416`, `c5a7f8c`, `ab8d1af`)
verified present in `git log`. `go build ./...` clean. `go vet ./...` clean. `go test ./...`
green (re-run after all three commits). `go test ./internal/compiler/cgen/... -run
'TestEmittedAttributeSet|TestCallSitesEmitNoArithmeticConversion' -v` and `go test
./internal/compiler/session/... -run 'TestPhase11' -v` both print all required PASS lines
(4 and 6 respectively). `11-MIDPHASE-GATE.md`'s own automated verify command
(`grep -v '^#' ... | grep -c -E '...'`) returns 9, exceeding the required floor of 5.
