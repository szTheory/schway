# Phase 22: Native Application Build and Single Execution — Context

**Captured:** 2026-09-27
**Status:** Ready for planning
**Provenance:** Handoff distilled from the user-approved M004 milestone review
and the user's request to preserve context before clearing the conversation.
This records existing decisions; it does not claim a separate discussion,
completed implementation plan, or executed Phase 22 work.

<domain>
## Phase Boundary

A developer can build a retained native executable and run it once on bounded
caller input, with ordinary application streams separate from compiler evidence.
Phase 22 owns APP-02 through APP-06, FFI-02, and EVD-11. ROADMAP and REQUIREMENTS
remain authoritative for acceptance. Start with the scalar witness `7 → 7` and
`42 → 42`, including malformed/oversized input and an observed single launch.
</domain>

<decisions>
## Accepted Decisions

- **D-22-01:** Keep Go 1.24, standard-library-first Stage 0, readable C17,
  installed Clang, and the independent interpreter. No new VM/backend/runtime.
- **D-22-02:** Separate ordinary application execution from conformance replay.
  Build performs no application effects; application run launches once.
  Do not interpret or run both optimization tiers against ordinary user effects.
- **D-22-03:** Produce a retained executable that works outside its temporary
  build directory, with declared runtime dependencies and build identity.
- **D-22-04:** Accept real bounded caller input. Define ordinary stdout, stderr,
  exit/signal outcomes and a separate evidence channel. Disabled, incomplete or
  exhausted evidence must not be mistaken for verified success.
- **D-22-05:** Local C sources/headers, symbols and ABI inputs are explicit
  trusted build inputs. Resolve paths reproducibly across checkout relocation;
  relevant changes invalidate artifacts/evidence. Do not generalize hardcoded
  fixture-symbol lookup or runtime.Caller paths into the application interface.
- **D-22-06:** Differential verification uses controlled isolated/replayable
  inputs, declared foreign outcomes, and independent expected answers. Modeled
  foreign success does not prove actual host IO or physical cleanup.
- **D-22-07:** Deliver a runnable gain in this phase, with the source/input/output
  frontier established before implementation. Each additional plan must serve a
  named witness or safety obligation. Preserve existing conformance contracts
  where needed; both application and evidence routes share checked body lowering.

### Delegated implementation choices

The user authorized following the reviewed recommendations automatically.
The planner should choose the smallest coherent CLI syntax, scalar/input
encoding, size limits, evidence transport, output encoding, artifact layout,
local C build-description format, and process failure policy that satisfy these
requirements. Settle them explicitly in phase planning using inspected code and
targeted research. Do not present an unchosen representation as a user decision.
Ask only for genuinely missing user intent, not to repeat milestone approval.

The input design must accommodate Phase 23's caller-selected file path without
implementing its resource semantics early. Fixed scalar output here does not
require a general String runtime or arithmetic.
</decisions>

<specifics>
## Product Intent and Progress

The user wants a usable language, maintained near/mid/long direction, and
proactive suggestions about checker obligations. M004 establishes real native
applications and resource lifetime. The following milestone targets scalar
arithmetic/iteration, minimal output, sum_to_n and FizzBuzz; byte libraries,
bounded JSON and an independent HTTP branch follow concrete consumers.

The latest completed implementation phase is **21**, archived under M004.
The just-completed work is **M004 kickoff, research, 24 requirements and the
four-phase roadmap**. Phase 22 has zero plans and zero execution summaries.
Next command is `$gsd-plan-phase 22`; after successful planning and plan checks,
the next implementation command is `$gsd-execute-phase 22`.
Do not rerun Phase 21, reopen M004, or mistake 0/0 plans for a completed phase.
</specifics>

<canonical_refs>
## Required Reading After a Context Clear

Read these repository-relative paths before researching/planning:

- `.planning/STATE.md` — current position and M004 handoff; later history is dated.
- `.planning/ROADMAP.md` — Phase 22 boundary, five criteria and dependencies.
- `.planning/REQUIREMENTS.md` — exact seven owned IDs and milestone-wide constraints.
- `.planning/PROJECT.md` — project intent and accepted current milestone.
- `.planning/PRODUCT-ROADMAP.md` — horizons, checker activation map and next suggestions.
- `.planning/LANGUAGE-MATURITY.md` — implemented capabilities and current refusals.
- `.planning/research/M004/SUMMARY.md` — accepted second fan-out synthesis.
- `.planning/research/M004/ARCHITECTURE.md` — actual execution/ownership boundaries.
- `.planning/research/M004/FEATURES.md` — user-visible witnesses and alternatives.
- `.planning/research/M004/STACK.md` — retained stack and explicit integration scope.
- `.planning/research/M004/PITFALLS.md` — adversarial findings and evidence costs.
- `.planning/research/M004/KICKOFF-VERIFICATION.md` — what kickoff checks established.

Use the M004 research directory; root research and M003 research are historical.
Do not repeat the ecosystem fan-out. Phase-specific research should resolve the
concrete implementation choices above. Phase 21's archived contract is prework;
its release/transfer correction belongs to the following resource phase.
</canonical_refs>

<code_context>
## Existing Code Insights

- `cmd/lang/main.go`: current public run command exposes engine/source selection.
- `internal/compiler/session/session.go`: `interpreterInputs` supplies fixture
  values; `runNative` interprets and runs O0/O3. This is the main separation seam.
- `internal/compiler/native/native.go`: temporary executable lifecycle and strict
  execution-JSON stdout/stderr handling constrain ordinary application behavior.
- `internal/compiler/cgen/cgen_program.go`: shared checked emitter and generated
  entry shell; foreign and by-pointer family refusals remain during Phase 22.
- `internal/compiler/native/foreign_resource.go` and `foreign_nonlocal.go`:
  repository fixture linkage is evidence infrastructure, not the public build API.

These observations were inspected during milestone research at `d9bde05`.
Recheck touched code during planning; don't treat inspection as runtime evidence.
</code_context>

<deferred>
## Deferred Scope

Real owned allocation/local discharge is Phase 23; call/error transfer is 24;
shared/exclusive read-copy pointer admission and integrated acceptance are 25.
No arithmetic/loops, owning aggregates, fallible implicit destruction, unwind,
callbacks, retained pointers or broad pointer mutation in Phase 22. No change to
Phase 21's completed UAT or historical verification scope.
</deferred>
