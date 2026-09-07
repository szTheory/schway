# Phase 4: Fallible Resources and C Boundary - Research

**Researched:** 2026-09-04
**Domain:** Compiler internals — typed-core extension (foreign call + fallible control-flow edge + resource release), C17 code generation, and native-execution semantics parity for a single audited FFI boundary.
**Confidence:** HIGH (decisions are locked in CONTEXT.md and cross-checked against in-repo source; residual LOW areas are flagged explicitly)

## Summary

Phase 4 is almost entirely pre-decided: `04-CONTEXT.md` records 29 locked decisions (D-04-01 through D-04-29) produced by four parallel research fan-outs at discuss-phase time, each already cross-checked against the wiki corpus, spike 005, and Phase 3's failure history. This RESEARCH.md's job is narrower than usual — it is not choosing a stack or exploring alternatives, it is (a) grounding those decisions against the actual current state of the repository so the planner does not re-derive stale assumptions, (b) mapping the concrete "six dispatch sites" and "two independent admission layers" the decisions repeatedly reference to real file/line locations, and (c) carrying forward the Phase 3 validation discipline (independent second derivation, a third decision-procedure oracle, mutation-kill negative controls) onto Phase 4's new claims.

The central engineering fact, confirmed by reading `internal/compiler/core/core.go`: `core.DataType.Alternatives` is `[]string` (line 21) and `core.OperationKind` today has exactly five members — `OpCopy`, `OpMove`, `OpBorrowShared`, `OpBorrowExclusive`, `OpReturn` (lines 128-134). Phase 4 adds new `OperationKind` values (`OpForeignCall`, `OpFail`, `OpRelease`, `OpDefect` — exact names are Claude's Discretion per CONTEXT.md) without touching `Alternatives`. Every one of the four dispatch/analysis packages that switches exhaustively on `operation.Kind` today — `interp` (two switches, confirmed at `interp.go:83-105` and `interp.go:124-146`, both ending in an error-returning `default:`), `corevalidate` (two switches, confirmed at `corevalidate.go:738-797` and `corevalidate.go:870-933`), `cgen` (two switches, confirmed at `cgen.go:237-286` and `cgen.go:470-520`) — must grow a case for each new kind or its `default:` will (correctly, fail-closed) reject the program. `check.go` does not dispatch on `core.OperationKind` at all (it dispatches on AST-level binding-kind strings like `"borrow_mut"`, confirmed at `check.go:820,854`); its "site" in D-04-22's six-site count is the point where it must *emit* the new operation kinds and its own execution-admission gate, not a `switch operation.Kind`. `originvalidate` and `pathoracle` each walk only `core.OpReturn` today (`originvalidate.go:103`, `pathoracle.go:247`) and must be widened to walk `{OpReturn, OpFail, OpDefect}` — this is D-04-29, independently confirmed as the single highest-risk item in the phase.

`native.go`'s `validateExecution` (confirmed at `native.go:243-249`) hard-codes exactly one legal terminal shape: `value.Outcome.Kind != "returned" || ... || len(value.LiveResources) != 0` is an error — meaning today ANY nonzero exit, ANY `LiveResources`, and ANY outcome other than `"returned"` is rejected outright. This function must grow an explicit expected-terminal-outcome axis before a typed-failure, defect-abort, or nonlocal-exit program can be run at all, exactly as D-04-24/native.go's code_context entry states.

The host this research ran on matches spike 005's recorded host exactly: Apple clang 21.0.0, arm64-apple-darwin, `nm`/`otool` present (llvm-nm-compatible), Go 1.24.0 darwin/arm64 — so D-04-19's `nm -u` allowlist control and D-04-10/11's frozen-fixture/conformance-TU approach are both executable on this machine without a fallback.

**Primary recommendation:** Follow CONTEXT.md's decisions verbatim; this phase's planning risk is not "what to build" but "landing all six dispatch sites and both independent admission layers (check vs. corevalidate, originvalidate vs. pathoracle) without missing one," and "proving the three-acquisition reverse-order release claim with a fixture where release *order* and release *set* are provably distinguishable" (per the `<specifics>` section of CONTEXT.md — this is the single sharpest finding in the whole research set).

## User Constraints (from CONTEXT.md)

<user_constraints>
### Locked Decisions

**Call surface**
- D-04-01: Phase 4 adds **`OpForeignCall` only**. No Lang-to-Lang calls. A `foreign C { }` declaration block is the only callable surface; acquisition steps are direct calls to declared foreign C symbols inside a linear body. Reversibility: costly.
- D-04-02: `OpForeignCall.Callee` resolving to a Lang `Function.ID` is a hard refusal (`core.call_target_not_foreign`) in **both** `check` and `corevalidate`, independently derived.
- D-04-03: The lift condition for Lang-to-Lang calls is recorded as **callable ⊆ publishable** — a function is callable only if `originvalidate.ValidatePublished` would publish it.

**Failure representation**
- D-04-04: Typed failure is a **two-successor control-flow edge in core**, not a data value. No `Result` type, no generics, no payload-carrying alternatives. `core.DataType.Alternatives` is **not touched this phase**. Reversibility: one-way.
- D-04-05: The error payload is an ordinary **existing nullary ADT** riding the `err` edge (e.g. `data AcquireError = | FirstRefused | SecondRefused`). The ok payload is an ordinary place on the ok successor block.
- D-04-06: `try` and `discard … because "<rationale>"` are the **only two admissible consumers** of a fallible operation. `acquire` is not an expression. `let x = acquire …` does not parse. No must-use lint — this is an IR well-formedness invariant.
- D-04-07: The reverse-order release sequence is **materialized as explicit `OpRelease` operations into each failure block by `check`**. `corevalidate` independently *rederives* the expected order from the block/edge graph and compares — never reads what `check` wrote.
- D-04-08: Terminal outcome is a **closed named set carried as its own axis**: `value | typed_failure | defect`, with `cancelled` **reserved but unconstructible** and a fail-closed control asserting no engine emits it.
- D-04-09: There is **no error-value constructor anywhere in the IR** — no `OpMakeErr`, no `Err(...)` expression. Only producer of `typed_failure` is an `OpFail` terminator; only producer of `OpFail` is an `err` edge carrying a place of the function's declared failure type.

**The C boundary**
- D-04-10: The boundary is a **hand-written, byte-frozen, separately-compiled foreign translation unit that wraps real libc `malloc`/`free`** and returns a record by value. The compiler forms its contract from Lang source and **never parses the foreign TU's header**. Reversibility: costly.
- D-04-11: A **generated conformance TU** (`lang_foreign_conformance.c`) is the single, explicit, auditable place where Lang's declaration and the foreign TU's own private header are permitted to meet. Defines no symbols, compiled but not linked, carries `_Static_assert(sizeof/_Alignof/offsetof)` pairs.
- D-04-12: Obligations are inspectable in **three layers derived from one authoritative `core.ForeignContract`**: (a) generated `_LANG_`-namespaced header with obligation comments + self-layout `_Static_assert`s, (b) the conformance TU, (c) a `lang.foreign/0` sidecar manifest digest-bound into `evidence.Manifest` via a new `foreign_digest` **omitempty** field.
- D-04-13: **cgen emits zero optimizer-visible attributes this phase** — no `restrict`, `noalias`, `nothrow`, `__attribute__((malloc))`, `nonnull`, `returns_nonnull`. Fail-closed control `control:foreign.no_unproven_attributes`. `emitted_attributes: []` is a field, not an omission. Reversibility: reversible.
- D-04-14: `_Noreturn` on the generated `lang_defect` function is **exempt** from D-04-13 — it's a property of code cgen itself emits (every path ends in `abort()`), not a claim about a foreign callee.

**Panic and foreign nonlocal exit**
- D-04-15: Phase 4 introduces a **real but deliberately terminal `defect` operation**: abort-only at the process root, no catch, no containment, no unwinding, and **no cleanup**. Reversibility: reversible.
- D-04-16: SC3's second half is **detection, not prevention**. Prevention *is* delivered statically at the declaration layer: **a foreign declaration that does not state its `unwind`/`nonlocal_exit` policy is refused admission**, independently, by both `check` and `corevalidate`. No default value.
- D-04-17: Detection is one **process-root `setjmp` landing pad** (one per process) plus a **cleanup ledger in `static` storage**. On a foreign nonlocal exit the pad emits `foreign.nonlocal_exit` plus one `resource.leaked` per still-live acquisition, then terminates as a defect.
- D-04-18: The pad **must not run releases**. Ledger lives in `static` storage. Enforced as `control:defect.no_release_on_defect` (zero `resource.released` events after a defect terminal record).
- D-04-19: The non-unwind control is **`nm -u` on the linked binary** — underscore-normalized, **allowlist** (not denylist). Must return operational/`tool_missing` on a host without `nm`, never pass. Do **not** build on unwind-section absence (Darwin/arm64 mandates compact unwind for many function shapes).
- D-04-20: Events must be **streamed at the point they occur**, not accumulated and written once at the end. Terminal record must be the last thing written; its **absence must be a failure, never a tolerated truncation**. This is an **additive emitter path**; Phase 1/2/3 emitters untouched.

**Method — carried forward from Phase 3, still binding**
- D-04-21: D-09/D-10/D-11 stand unchanged. Mutation-kill every oracle. Interrogate what inputs a green property test actually reaches. Drive the shipped `./cmd/lang` binary on hand-written **out-of-corpus** programs.
- D-04-22: D-12/D-12a stand and **widen**. `check` and `corevalidate` remain independent derivations with no shared helpers. A new `OperationKind` must be handled at **six** sites — `check`, `corevalidate`, `interp`, `cgen`, `pathoracle`, `originvalidate`. Required control: `control:kind.exhaustive_dispatch`, table-driven over a new `core.AllOperationKinds()` registry, mutation-killed by deleting one `case`.
- D-04-23: D-13 stands. No `lang.core/2`. Every new core field is additive and `omitempty`. Assert byte-identity of Phase 1/2/3 core programs **and** evidence-manifest IDs **by test**, before anything else lands.
- D-04-24: D-15 stands and is newly load-bearing: a test that runs a program which `abort()`s or `longjmp`s must handle nonzero exit, signals, and truncated output. SIGABRT assertion goes through `ProcessState.Sys().(syscall.WaitStatus)` — **never** a hardcoded `134`.

**Carried Phase 3 debt — disposition**
- D-04-25: D-03-01 gets the cheap half: instrument `discoverLoanLastUses` to count its own transitive-scan work, matching `blockLoanLiveness`'s `result.Work++`-per-operation convention.
- D-04-26: Retiring `discoverLoanLastUses` carries to Phase 5. Re-record as dated open debt in `04-DEBT.md`.
- D-04-27: WR-01 is closed this phase: wire `originvalidate.ValidatePublished` into `lang check`/`lang run`, not only `interface export`.
- D-04-28: `originvalidate` must recognise a value returned from an `OpForeignCall` whose declaration says it borrows/retains an argument as **borrow-derived**, with new `core.foreign_origin_omitted` refusal.
- D-04-29: `originvalidate` and `pathoracle` must walk **every** terminator — `{OpReturn, OpFail, OpDefect}` — not just `OpReturn`. **Single highest-risk item in the phase.** Required control asserting the walked terminator set equals the full set, mutation-killed by deleting `OpFail` from either walker.

### Claude's Discretion

Plan decomposition, wave structure, plan count, the concrete spelling of the `foreign` block and of `try` / `discard … because`, `OperationKind` naming (reconcile `OpForeignCall`, `OpAcquire`/`OpResourceAcquire`, `OpRelease`/`OpResourceRelease`, `OpFail`, `OpDiscard`, `OpDefect` into one minimal set at planning time), the exact `lang.foreign/0` field layout, the fixture type shape, and the diagnostic code names — provided every decision above holds and the D-04-22 six-site dispatch obligation is discharged.

### Deferred Ideas (OUT OF SCOPE)

- Lang-to-Lang calls (`OpCall`) — Phase 5, gated on **callable ⊆ publishable** (D-04-03).
- A storable, matchable `Result` value and payload-carrying alternatives — M002 or Phase 6. The additive move is a sibling `alternative_details []Alternative omitempty` keyed by name, never a shape change to `Alternatives []string`.
- Retiring `discoverLoanLastUses` — Phase 5 (D-04-26).
- Allocator-mismatch and stale-callback-retention detection — Phase 5's NAT-03. Phase 4 only carries allocator identity in the contract and independently refuses a release whose declared allocator differs from its acquisition's.
- ASan/UBSan lanes — Phase 5, isolated from semantic-equivalence evidence.
- Callback registration, generation tokens, dynamic-library handles, `errno` translation, symbol versioning, `blocking`/`reentrant` enforcement, header ingestion — declared and refused-if-non-default this phase; never exercised.
- Panic containment, isolation boundaries, `catch_unwind` analogues, cancellation shields/budgets/checks — Phase 6 and beyond.
- Loop-carried loan liveness — still deferred; no loop/recursion construct this phase.
- Full DWARF/CodeView emission, crash capture, proof/SMT tiers — unchanged from Phase 3's D-03.

**Accepted residual limitations (record, do not engineer around):**
- The coordinated three-way lie (frozen fixture + `foreign` declaration + conformance TU expectations edited together) passes Phase 4 on a wrong boundary — QLT-01's documented escape class, one layer down.
- The pad cannot see a `longjmp` to a foreign-owned `jmp_buf` established below it (reachable only through foreign-invoked callbacks); foreign `exit()`/`_Exit()`/`raise()` is not observable at all.
- One small by-value record on one host (Apple arm64/Clang). Unions, bit fields, vectors, variadics, packed/aligned records, aggregate-passing thresholds, ELF/x86-64/GCC/non-Apple-Clang remain open — spike 005's Known Limitations fence carries forward.
- Quarantine is permanent and non-discharging. A green gate never proves the foreign implementation obeys its declared contract.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| SEM-03 | `Result` propagation and ignored-result rules produce explicit typed control flow; panic and cancellation cannot be erased as ordinary errors. | D-04-04..D-04-09 (edge-based failure, no error constructor, `try`/`discard` as sole consumers); D-04-15 (real terminal `defect`); confirmed structurally enforced because `OpFail` has no route from a panic path — see Architecture Patterns, Pattern 2. |
| RES-01 | Partially initialized noncopyable resources release exactly once in reverse completed-acquisition order on return and typed failure. | D-04-07 (materialized `OpRelease` by `check`, independently rederived by `corevalidate`); three-acquisition fixture requirement in `<specifics>`; see Common Pitfalls #1 and Validation Architecture. |
| FFI-01 | Foreign contracts carry target layout, initialized state, allocator identity, capture/retention, aliasing, and unwind obligations. | D-04-10..D-04-14 (`core.ForeignContract`, conformance TU, `_Static_assert` layout proof, zero unproven optimizer attributes); D-04-16 (mandatory `unwind`/`nonlocal_exit` declaration, no default). |
</phase_requirements>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Foreign call declaration + admission (`foreign C {}`, `OpForeignCall`) | Frontend/checker (`check`) | Independent validator (`corevalidate`) | Contract must be asserted at parse/check time and independently re-derived — no shared code path per D-12/D-04-22. |
| Typed-failure control flow (`try`, `discard`, `err`/`ok` edges, `OpFail`) | Core IR (`core` package) | Checker (`check`) emits, `corevalidate` re-derives | This is a new IR shape, not a library; it must be representable in `core.Program` before any consumer can exist. |
| Reverse-order release materialization (`OpRelease`) | Checker (`check`) materializes | `corevalidate` independently rederives from block/edge graph | D-04-07 explicitly requires two non-shared derivations of the same release order. |
| Two execution engines' agreement on outcome/events | Interpreter (`interp`) is the semantic oracle | Native (`cgen` + `native.go` harness) must match it | Established precedent since Phase 1 (`INT-01`); Phase 4 extends the oracle to `typed_failure`/`defect` outcomes. |
| C boundary layout/ABI proof | Generated C (`cgen`) + hand-frozen foreign TU + generated conformance TU | `evidence.Manifest`'s `foreign_digest` | D-04-10..12: quarantine is an information-flow property, enforced by what the compiler is permitted to parse, not by file location. |
| Panic/defect termination and nonlocal-exit detection | Native runtime harness (`native.go` process supervision) + generated `setjmp` pad in `cgen` output | `evidence`/session control (`control:defect.no_release_on_defect`, `control:foreign.unwind_forbidden` via `nm -u`) | Detection must be observable in the linked binary (`nm -u`) and in the process's own exit/signal state — neither is a checker-time property. |
| Path/origin re-derivation over new terminators | `pathoracle` (paths), `originvalidate` (origin) | — | D-04-29: both currently walk only `OpReturn`; must walk `{OpReturn, OpFail, OpDefect}` — this is a shared *risk*, not a shared *tier*, since the two packages remain independent derivations. |

## Standard Stack

This phase adds **no new external dependencies**. It extends the existing Go 1.24 stdlib compiler (`internal/compiler/...`) and the existing hand-written/generated C17 + Apple Clang 21 native path. No npm/pip/cargo packages are installed; the "Package Legitimacy Audit" and "Installation" sections below are therefore N/A and are included only to document that fact explicitly.

### Core
| Component | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go stdlib (`os/exec`, `syscall`, `encoding/json`, `crypto/sha256`) | Go 1.24.0 (confirmed via `go version`) | Compiler host language; process supervision for native execution; JSON core/evidence serialization; content-identity digests | Already the project's entire toolchain (FND-01); no reason to introduce a dependency for a phase whose new surface is IR shapes and C generation |
| Apple Clang | 21.0.0 (confirmed via `clang --version` on this host, matches spike 005's recorded host exactly) | Compiles generated C17 and the frozen foreign TU/conformance TU at `-O0`/`-O3` | Existing `NAT-01` toolchain; `DefaultFlags` in `evidence.go` already pins `-std=c17 -Wall -Wextra -Werror -pedantic -O0 -O3` (verbatim, `evidence.go:38`) |
| `nm` (llvm-nm, GNU-compatible) | present on this host at `/usr/bin/nm` | D-04-19's non-unwind allowlist control | Confirmed present; D-04-19 requires the control degrade to `operational/tool_missing` (never silently pass) when absent — plan must include that fallback path even though this host has it |
| `otool` | present at `/usr/bin/otool` | Spike 005 precedent used `otool -l | grep -c __unwind_info` and explicitly rejected it as the primary control (Darwin/arm64 mandates compact unwind for many function shapes) — available as a secondary/diagnostic tool only, never the pass/fail gate | D-04-19 |

### Supporting
| Component | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `setjmp.h`/`longjmp` (C standard library, C17 §7.13) | C17 | Process-root landing pad (D-04-17) | Only at the single process-root pad; never per-call, per D-04-17's cost constraint |
| `_Static_assert` (C17 §6.7.10) | C17 | Conformance TU's `sizeof`/`_Alignof`/`offsetof` pairing (D-04-11) | Compile-time layout proof between Lang's declared contract and the foreign TU's private header |
| `abort()` (C standard library) | C17 | Terminal defect / `lang_defect` (D-04-14/D-04-15) | Every path of `_Noreturn lang_defect` ends here; this is the one function cgen may mark `_Noreturn` on despite D-04-13's blanket ban |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Hand-frozen foreign TU wrapping libc `malloc`/`free` (D-04-10) | Pure libc directly, no wrapper | Rejected in CONTEXT.md — `malloc` has no record layout, so SC2's "target layout inspectable" has no subject, and 6 of Phase 5's 7 hostile mutations become non-injectable |
| `nm -u` allowlist for unwind-forbidden proof (D-04-19) | `otool -l \| grep -c __unwind_info` (section-absence check) | Rejected — Darwin/arm64 mandates compact unwind for many function shapes regardless of `-fno-exceptions`, so a section-count assertion passes/fails for the wrong reason; measured `0` on this exact host under `-fno-exceptions -fno-asynchronous-unwind-tables -fno-unwind-tables` is a coincidence, not a load-bearing signal |
| Edge-based typed failure (D-04-04) | A storable `Result<T,E>` ADT value now | Rejected — no generics infrastructure exists yet; bolting a storable Result onto an edge-based core later is a one-way re-lowering, explicitly priced as debt for M002/Phase 6 |

**Installation:** N/A — no new packages. All tooling above is already present in the Go toolchain or the Apple Clang / Xcode command-line tools already used by Phase 1-3.

**Version verification:** Confirmed live on this host during this research session:
```
$ clang --version
Apple clang version 21.0.0 (clang-2100.1.1.101)
Target: arm64-apple-darwin25.6.0
$ go version
go version go1.24.0 darwin/arm64
$ nm --version  (llvm-nm, GNU nm compatible)
$ otool  (present, /usr/bin/otool)
```
This exactly matches spike 005's recorded host ("Apple Clang 21.0.0 arm64"), so spike 005's iteration 4 (`restrict` divergence) and D-04-19's unwind-section observation (`otool -l | grep -c __unwind_info` returned `0`) are reproducible on the actual planning/execution machine, not just asserted from the spike's README.

## Package Legitimacy Audit

**N/A — this phase installs no external packages.** All new code is Go stdlib plus hand-written/generated C17. No `npm view`/`pip index versions`/`cargo search` verification is applicable. If a future plan proposes any third-party dependency (e.g., a C static-analysis tool, a fuzzing harness), the Package Legitimacy Gate protocol must be run at that time — none is anticipated for Phase 4's locked decisions.

## Architecture Patterns

### System Architecture Diagram

```text
 Lang source (.lang)
   │  parses a `foreign C { ... }` block + a linear body with
   │  acquire-step calls, `try`, `discard ... because "..."`
   ▼
 ┌─────────────────────────────────────────────────────────────┐
 │ check.go (frontend admission)                                │
 │  - parses foreign decl -> core.ForeignContract                │
 │  - refuses OpForeignCall.Callee == a Function.ID (D-04-02)     │
 │  - refuses missing unwind/nonlocal_exit policy (D-04-16)       │
 │  - emits OpForeignCall / OpFail / OpRelease / OpDefect ops      │
 │  - materializes reverse-order OpRelease into each fail block    │
 │    (D-04-07)                                                    │
 └───────────────┬───────────────────────────────────────────────┘
                 │ core.Program (JSON, lang.core/0 — byte-identical
                 │ for pre-Phase-4 programs; new fields additive
                 │ omitempty, D-04-23)
                 ▼
   ┌─────────────────────────┬───────────────────────────┬─────────────────────┐
   │ corevalidate            │ originvalidate            │ pathoracle           │
   │ (independent re-derive) │ (walks {OpReturn,OpFail,   │ (walks {OpReturn,    │
   │ - re-derives release    │  OpDefect} per D-04-29)    │  OpFail,OpDefect}    │
   │   order from block/edge │ - foreign_origin_omitted   │  per D-04-29)        │
   │   graph, compares       │   for retained/borrowed    │ - enumerates paths   │
   │   (D-04-07)             │   foreign returns (D-04-28)│   through new        │
   │ - refuses               │ - ValidatePublished now     │   terminators        │
   │   call_target_not_      │   wired into `check`/`run`  │                      │
   │   foreign (D-04-02)     │   (WR-01/D-04-27)           │                      │
   └─────────────┬───────────┴─────────────┬─────────────┴──────────┬──────────┘
                 │ all independently agree │                        │
                 ▼                          ▼                        ▼
   ┌────────────────────────────┐   ┌──────────────────────────────────────────┐
   │ interp (semantic oracle)   │   │ cgen (C17 generation)                     │
   │ - two dispatch switches    │   │ - two dispatch switches gain new cases     │
   │   (runLinear/runBranchArm) │   │ - emits `_LANG_`-namespaced foreign decl,  │
   │   gain OpForeignCall/      │   │   generated header w/ _Static_assert self- │
   │   OpFail/OpRelease/OpDefect│   │   layout, conformance TU, streaming event  │
   │   cases (D-04-22)          │   │   emitter (additive path, D-04-20), zero   │
   │ - emits ordered semantic   │   │   optimizer attributes (D-04-13/14)        │
   │   events incl. resource.*  │   │ - process-root setjmp pad + static ledger  │
   │   and defect events        │   │   (D-04-17/18)                             │
   └──────────────┬─────────────┘   └───────────────────┬────────────────────────┘
                  │                                       │ compiles separately with
                  │                                       │ frozen foreign TU (D-04-10)
                  │                                       ▼
                  │                          ┌─────────────────────────────┐
                  │                          │ native.go harness             │
                  │                          │ - validateExecution widened   │
                  │                          │   for value|typed_failure|    │
                  │                          │   defect outcome axis          │
                  │                          │   (currently hard-rejects      │
                  │                          │   anything but "returned")     │
                  │                          │ - runs -O0 and -O3              │
                  │                          │ - nm -u allowlist control       │
                  │                          │   (D-04-19)                     │
                  │                          └────────────────┬────────────────┘
                  └──────────────────────────┬─────────────────┘
                                              ▼
                          differential comparison: interp vs O0 vs O3
                          on terminal outcome + ordered semantic events
                          + live-resource state  (INT-01/NAT-02 parity,
                          extended to typed_failure/defect this phase)
                                              ▼
                          evidence.Manifest (+ new foreign_digest omitempty
                          field binding the lang.foreign/0 sidecar manifest)
```

### Recommended Project Structure

No new top-level directories. Additive changes land inside the existing package layout:

```
internal/compiler/
├── core/           # + OperationKind values, core.ForeignContract, AllOperationKinds()
├── check/          # + foreign decl admission, OpFail/OpRelease/OpDefect emission
├── corevalidate/   # + independent re-derivation of release order, foreign-call refusal
├── interp/         # + two new dispatch cases (runLinear, runBranchArm)
├── cgen/           # + two new dispatch cases, header/conformance-TU/manifest emission,
│                   #   streaming event emitter, setjmp pad + static ledger
├── originvalidate/ # + walk OpFail/OpDefect, foreign_origin_omitted refusal, wire into
│                   #   `lang check`/`lang run` (WR-01)
├── pathoracle/     # + walk OpFail/OpDefect
├── native/         # + widened validateExecution terminal-outcome axis, multi-TU compile,
│                   #   nonzero-exit/signal handling
├── evidence/       # + foreign_digest omitempty field
└── session/        # + new required controls in the gate's control table

testdata/phase4/    # new fixtures: three-acquisition resource, foreign decl variants,
                    # nonlocal-exit probe, layout-mutation golden
native/ (generated) # frozen foreign TU (byte-frozen, hand-written, analogous to
                    # testdata/phase2/owned_transfer.golden.c), generated conformance TU
```

### Pattern 1: Additive, omitempty core evolution (established Phase 1-3 precedent)

**What:** Every new core field lands as `omitempty` on the existing schema (`lang.core/0`); the schema version does not bump. `core.Block`/`core.Edge`/`core.LoanEndpoint`/`core.MatchArm.BlockID` all landed this way without moving a Phase 1/2 golden byte.

**When to use:** Every Phase 4 core addition — `core.ForeignContract`, the `err`/`ok` edge shape, `OpRelease`'s target resource reference — must follow this exact precedent.

**Example (verified in-repo, `internal/compiler/core/core.go:151-161`):**
```go
type LinearBody struct {
	ID         string            `json:"id"`
	Types      []TypeFact        `json:"types"`
	Places     []Place           `json:"places"`
	Operations []LinearOperation `json:"operations"`
	// Blocks, Edges, and LoanEndpoints are additive omitempty Phase 3 facts:
	// a plain straight-line linear body (Phase 2 and earlier) never
	// populates them, so its serialized bytes are unchanged (D-13).
	Blocks        []Block        `json:"blocks,omitempty"`
	Edges         []Edge         `json:"edges,omitempty"`
	LoanEndpoints []LoanEndpoint `json:"loan_endpoints,omitempty"`
}
```
Phase 4's new fields (foreign contract reference, failure-edge target, release-target place) should read exactly like this: a comment explaining why the field is absent for every pre-Phase-4 program, and a `,omitempty` tag.

### Pattern 2: Fail-closed exhaustive dispatch (established Phase 1 precedent, now the D-04-22 gate)

**What:** Every package that switches on `core.OperationKind` ends with a `default:` that returns an error rather than silently ignoring an unknown kind.

**Example (verified in-repo, `internal/compiler/interp/interp.go:83-105`):**
```go
switch operation.Kind {
case core.OpCopy:
	values[operation.TargetID] = value
	events = append(events, ownedEvent(function, operation, "value.copied"))
case core.OpMove:
	delete(values, operation.SourceID)
	values[operation.TargetID] = value
	events = append(events, ownedEvent(function, operation, "value.transferred"))
case core.OpBorrowShared:
	values[operation.TargetID] = value
	events = append(events, ownedEvent(function, operation, "value.borrowed"))
case core.OpBorrowExclusive:
	values[operation.TargetID] = value
	events = append(events, ownedEvent(function, operation, "value.borrowed_exclusive"))
case core.OpReturn:
	events = append(events, Event{ /* ... */ })
	return Execution{ /* ... */ }, nil
default:
	return Execution{}, fmt.Errorf("operation %q has unknown kind %q", operation.ID, operation.Kind)
}
```
This exact shape repeats at `interp.go:124-146` (the sibling `runLinear` switch), `corevalidate.go:738-797` and `corevalidate.go:870-933`, and `cgen.go:237-286` and `cgen.go:470-520` — six switch statements total across three files (`interp` ×2, `corevalidate` ×2, `cgen` ×2), all currently missing cases for the four new `OperationKind` values. **This is the concrete meaning of D-04-22's "six sites"** for the switch-based half; `check`'s emission side and `pathoracle`/`originvalidate`'s terminator-walk side are the other two of the six sites named in the decision, but they are not `switch operation.Kind` dispatchers — they are, respectively, an emission point and two terminator-set membership tests (`operation.Kind == core.OpReturn`, confirmed at `originvalidate.go:103` and `pathoracle.go:247`) that must become `{OpReturn, OpFail, OpDefect}` set-membership tests.

### Pattern 3: Independent re-derivation, never shared helpers (D-12/D-12a, unchanged since Phase 2/3)

**What:** `check` and `corevalidate` compute the same fact (e.g., loan conflicts, now: release order) through materially different mechanisms and compare, so a bug in one derivation's *algorithm* — not just its *output* — is still caught.

**Concrete Phase 3 precedent (verified in-repo comment, `corevalidate.go` region around line 738, referenced by 03-DEBT.md D-03-01):** `corevalidate.recomputeLoanEndpoints` independently re-derives loan endpoints via a reachability closure + reduction, never `check.go`'s iterative worklist fixpoint.

**Phase 4 application:** `check` *materializes* explicit `OpRelease` operations into each failure block in the order it computes (D-04-07). `corevalidate` must *rederive* the expected reverse-acquisition-order release sequence purely from the block/edge graph structure — walking backward from each failure edge to find every `OpForeignCall` (acquisition) that dominates it and is not yet released, in reverse discovery order — and compare against what `check` emitted, **never reading `check`'s intermediate release-order data structure**. This is what makes the "transpose two `OpRelease` emissions" mutation (see `<specifics>` in CONTEXT.md) a real negative control rather than a tautology: if `corevalidate`'s rederivation used the same ordering data as `check`, the mutation would corrupt both derivations identically and the differential would stay green.

### Pattern 4: Quarantine as an information-flow property, not an authorship property (D-04-10's reframe)

**What:** A generated C artifact "sees" only what the compiler is permitted to parse. The compiler forming its contract from Lang source alone and never parsing the foreign TU's header is what makes the foreign TU quarantined — even though it's hand-written and lives in-repo. A shim whose private header the generated C `#include`s would destroy quarantine even if the shim were nominally "foreign," because the *information* (real struct layout) would flow into the compiler's assumptions.

**Applied consequence for planning:** the conformance TU (D-04-11) is the *only* artifact allowed to `#include` both the generated `_LANG_` header and the foreign TU's private header. `cgen`'s main emission path must never reference the foreign TU's header at all — this is a structural constraint on which files a plan's tasks are allowed to make `cgen` depend on.

### Anti-Patterns to Avoid

- **Deriving optimizer attributes (`restrict`, `nothrow`, `noalias`) from a *declared* fact instead of a *proven* one.** CONTEXT.md's D-04-13 explicitly overrules two of the four research fan-outs that proposed emitting `nothrow` from an admitted `unwind: forbidden` declaration. Spike 005 iteration 4 demonstrated that a false `restrict` returns `2` at `-O0` and `1` at `-O3` on this exact host — an unproven attribute produces *wrong code*, not slow code. Do not let a plan task "helpfully" add an attribute because the declaration technically supports it; Phase 4 has no alias/capture analysis to derive it from.
- **Running releases from inside the `setjmp` pad.** C17 §7.13.2.1 (confirmed via SEI CERT MSC22-C, matching the standard's own wording) leaves non-`volatile` automatic objects changed since `setjmp` indeterminate after `longjmp`. Releasing from indeterminate handles converts a leak into a use-after-free — this is D-04-18, and the reason the cleanup ledger and any release bookkeeping must live in `static` storage, confirmed by spike 005's actual `native/nonlocal_exit.c` (`static jmp_buf destination; static int acquired = 0; static int released = 0;`).
- **Building the non-unwind control on unwind-*section absence*.** D-04-19 explicitly rejects `otool -l | grep -c __unwind_info` as the pass/fail signal, because Darwin/arm64 mandates compact unwind for many function shapes regardless of `-fno-exceptions` flags — the section can be present on a program that never actually unwinds. Use the `nm -u` allowlist instead.
- **Letting the bounded-writer output caps (`MaxStreamBytes` = 64 KiB, confirmed `native.go:18`; `LANG_OUTPUT_LIMIT` = 65536, confirmed `cgen.go:318`) silently truncate the terminal defect/nonlocal-exit record.** D-04-20 requires the terminal record's *absence* be a hard failure, never tolerated as an ordinary truncation — because with today's buffered `emitEventSupport` (confirmed: it "writes the LANG_EVENT macros ... " and buffers, per `cgen.go:310` doc comment), an aborting process emits zero events, which would make SC4 unfalsifiable on exactly the paths SC3 is about.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Cross-TU layout agreement | A custom layout-comparison tool that inspects compiled object files | C17 `_Static_assert(sizeof(...) == N, ...)` / `_Alignof` / `offsetof` pairs in a generated conformance TU (D-04-11) | The C compiler itself is the layout oracle; a bespoke binary inspector would need to reimplement ABI knowledge Clang already encodes, and would not fail at compile time the way `-Werror` + `_Static_assert` does |
| Detecting a foreign-triggered nonlocal exit | A custom stack-scanning or signal-trampoline mechanism | One process-root `setjmp`/`longjmp` landing pad (D-04-17) | The wiki's cost constraint explicitly rejects per-call/per-borrow landing pads; a single process-root pad is the cheapest mechanism that can observe *any* nonlocal exit reaching the root, and spike 005 already validated the pattern end-to-end |
| Proving "no unwind crosses this boundary" | Parsing DWARF/CFI unwind tables to reason about what *could* unwind | `nm -u` allowlist over the linked binary's undefined symbols (D-04-19) | Spike 005's own investigation (iteration + D-04-19's rationale) shows unwind-table presence is a poor proxy on Darwin/arm64; the allowlist approach directly answers "did we link anything we didn't audit" without needing DWARF expertise |
| Reasoning about whether a C implementation obeys its declared contract | A verifier that inspects the foreign TU's compiled machine code for compliance | Nothing — this is explicitly an accepted, permanent, non-discharging residual (wiki: "a generated binding can create a candidate contract, never proof that the C implementation obeys it") | Attempting to "solve" this would be scope creep into a full C verifier; the project has already adjudicated this as out of reach and documents it honestly in `unchecked_obligations` |

**Key insight:** Every "Don't Hand-Roll" item above is really the same insight restated: Phase 4's job is to make obligations *inspectable*, not to *prove* them are met by the foreign side. The compiler's own toolchain (Clang's `_Static_assert`, the linker's undefined-symbol table, the process's own exit/signal state) is always preferred over a bespoke analysis, because the toolchain's failure mode is a hard compile/link/run error rather than a silently-wrong custom heuristic.

## Common Pitfalls

### Pitfall 1: The two-acquisition fixture makes release *order* indistinguishable from release *set*

**What goes wrong:** A three-acquisition fixture is required (not two), because with only two acquisitions the release *set* and the release *order* coincide — a differential over a two-stage fixture would ask "were both things released" and stay green even if `cgen` derives its cleanup order from its own internal `goto`-ladder structure rather than from the materialized `OpRelease` list `check` emitted.

**Why it happens:** It is easy to satisfy SC1 ("releases only initialized resources, exactly once, in reverse order") with a fixture that cannot actually falsify "reverse order" as distinct from "all of them."

**How to avoid:** Ship a **three**-acquisition fixture and a mutation that **transposes** two `OpRelease` emissions specifically (not omits one — that's a different mutation, covered by pitfall 2) and confirm the differential turns red. This is CONTEXT.md's own "sharpest finding in the whole research set" (`<specifics>` section) and directly instantiates the standing rule D-10 (mutation-kill every oracle).

**Warning signs:** A plan that only specifies a two-stage acquisition fixture for SC1; a mutation-kill test that only tests *omission* of a release, never *transposition* of two releases.

### Pitfall 2: Layout mutation and omitted-release mutation must attack *different* artifacts

**What goes wrong:** If both hostile mutations are applied to the same artifact (e.g., both mutate the generated C), an author who keeps the emitter and the fixture in sync passes both controls without either control actually proving independence.

**Why it happens:** It's the path of least resistance to write one mutation-injection harness and reuse it for both controls.

**How to avoid:** Layout mutation must attack the **frozen foreign fixture** (the hand-written, byte-frozen TU — D-04-10). Omitted-release mutation must attack the **production emitter's own output** (what `cgen` itself generates). CONTEXT.md states this explicitly as "the structural answer to 'an in-repo shim is self-confirming'" — passing one control requires the emitter to be right; passing the other requires the fixture to be independent of the emitter.

**Warning signs:** A single `MutationRunner`-style helper (cf. `session.go`'s existing `OwnedBackendMutationRunner`, confirmed present at `session.go:67-115`) being reused verbatim for both the layout control and the release-omission control without adapting which file it mutates.

### Pitfall 3: `validateExecution`'s current hard-reject makes every new terminal outcome unrunnable until explicitly widened

**What goes wrong:** `internal/compiler/native/native.go`'s `validateExecution` (confirmed verbatim, `native.go:247`):
```go
if value.Outcome.Kind != "returned" || value.Outcome.Value == "" || value.Events == nil || value.LiveResources == nil || len(value.Events) == 0 || len(value.LiveResources) != 0 {
	return errors.New("execution document violates required contract")
}
```
means literally any nonzero exit, any populated `LiveResources`, or any `Outcome.Kind` other than the exact string `"returned"` is rejected as a validation error *before* a Phase 4 typed-failure, defect-abort, or nonlocal-exit program can even be compared. If a plan doesn't widen this function first, every other Phase 4 native-execution task is silently blocked.

**Why it happens:** This function's shape was correct and sufficient through Phase 3 (nothing could produce another outcome); it becomes a landmine specifically because Phase 4 is the phase that introduces a second and third terminal outcome.

**How to avoid:** Add an explicit expected-terminal-outcome axis to the validation call, and — per the `code_context` note in CONTEXT.md — add a regression test asserting Phase 1-3 documents are *still* rejected on the *old* grounds (i.e., the widening must not accidentally make a genuinely malformed Phase 1-3 execution document pass).

**Warning signs:** A plan task that adds `OpDefect`/`OpFail` cgen/interp support but never touches `native.go`'s `validateExecution`; a native-execution test for a Phase 4 fixture that fails with "execution document violates required contract" and gets treated as an unrelated bug rather than the expected missing-widening symptom.

### Pitfall 4: `originvalidate`/`pathoracle` walking only `OpReturn` makes the failure path invisible to origin/path analysis while O0/O3 differentials still pass

**What goes wrong:** Both `originvalidate.RecomputeOriginPerReturn` (confirmed, walks "backward from EVERY `core.OpReturn`", `originvalidate.go:84-103`) and `pathoracle`'s path enumeration (confirmed, checks `operation.Kind == core.OpReturn`, `pathoracle.go:247`) currently only recognize `OpReturn` as a terminator. If either is not widened to `{OpReturn, OpFail, OpDefect}`, a program whose *only* exit on some path is `OpFail` or `OpDefect` is invisible to that walker — and because the O0/O3 differential only compares paths `pathoracle` enumerated, the differential can pass while an entire origin/path fact is silently never checked.

**Why it happens:** This is explicitly named in CONTEXT.md as reproducing "exactly the defect class 03-08/03-09/03-10 closed three times (an analysis that walked *a* return rather than *every* return)" — the same shape of bug recurring at a new terminator set.

**How to avoid:** Per D-04-29, add a required control asserting the walked terminator set equals the full set (`core.AllOperationKinds()` filtered to terminators, or an explicit `{OpReturn, OpFail, OpDefect}` constant), mutation-killed by deleting `OpFail` from either walker.

**Warning signs:** A test suite that is green for both `originvalidate` and `pathoracle` on Phase 4 fixtures, but where every fixture's *only* exercised path ends in `OpReturn` (i.e., no fixture actually exits via `OpFail` or `OpDefect` on a path that reaches an origin- or path-sensitive check).

### Pitfall 5: Buffered event emission silently erases the evidence an aborting/`longjmp`ing process was supposed to produce

**What goes wrong:** `cgen.go`'s existing `emitEventSupport` buffers events and writes once at the end (confirmed, doc comment at `cgen.go:310`, `LANG_OUTPUT_LIMIT` constant at `cgen.go:318`). A process that calls `abort()` or is redirected by `longjmp` never reaches its own "write buffer" step, so it emits **zero** events under the old mechanism — silently making SC4 (interpreter/native agreement on primary failure and cleanup events) unfalsifiable on exactly the paths SC3 (panic/nonlocal-exit) is about.

**Why it happens:** The buffered-write pattern was correct and sufficient for Phase 1-3, where every path ends in an ordinary `OpReturn`.

**How to avoid:** D-04-20 requires a new, **additive** streaming emitter path (leaving `emitEventSupport` and the Phase 1/2/3 emitters untouched, so those goldens stay byte-identical per D-13) that writes each event as it occurs, ends with the terminal record as the last write, and treats the terminal record's *absence* as a hard failure — never a tolerated truncation, even under the existing `MaxStreamBytes`/`LANG_OUTPUT_LIMIT` caps.

**Warning signs:** A plan task that adds `OpDefect`/nonlocal-exit support to `cgen` without adding a distinct streaming write path; a test that can't tell the difference between "the process aborted before writing any events" (a real defect) and "the process's output was truncated by the byte cap" (a tooling limitation) because both currently look identical (empty/short output).

## Code Examples

### Existing exhaustive-dispatch shape to extend (verified, `internal/compiler/interp/interp.go:83-105` and its sibling at `:124-146`)

```go
switch operation.Kind {
case core.OpCopy:
	values[operation.TargetID] = value
	events = append(events, ownedEvent(function, operation, "value.copied"))
case core.OpMove:
	delete(values, operation.SourceID)
	values[operation.TargetID] = value
	events = append(events, ownedEvent(function, operation, "value.transferred"))
case core.OpBorrowShared:
	values[operation.TargetID] = value
	events = append(events, ownedEvent(function, operation, "value.borrowed"))
case core.OpBorrowExclusive:
	values[operation.TargetID] = value
	events = append(events, ownedEvent(function, operation, "value.borrowed_exclusive"))
case core.OpReturn:
	events = append(events, Event{
		Schema: execution.Schema1, ID: operation.ID + ":event:returned", Kind: "function.returned",
		FunctionID: function.ID, SourcePlace: operation.SourceID, TypeID: operation.TypeID,
	})
	return Execution{Schema: execution.Schema1, Outcome: Outcome{Kind: "returned", Value: value}, Events: events, LiveResources: []string{}}, nil
default:
	return Execution{}, fmt.Errorf("operation %q has unknown kind %q", operation.ID, operation.Kind)
}
```
Phase 4 adds `case core.OpForeignCall:`, `case core.OpFail:`, `case core.OpRelease:`, `case core.OpDefect:` (names per Claude's Discretion) before the `default:` in this switch, in `interp.go`'s sibling switch, and in `corevalidate.go`'s two switches (`:738-797`, `:870-933`) and `cgen.go`'s two switches (`:237-286`, `:470-520`).

### Existing terminator-walk shape to widen (verified, `internal/compiler/originvalidate/originvalidate.go:94-103`)

```go
// walks backward from EVERY core.OpReturn in [a function's operations]
for _, operation := range function.Linear.Operations {
	if operation.Kind == core.OpReturn {
		// ...
	}
}
```
and (verified, `internal/compiler/pathoracle/pathoracle.go` around line 247):
```go
if operation.Kind == core.OpReturn {
	// ...
}
```
Both must become membership tests against `{OpReturn, OpFail, OpDefect}` per D-04-29.

### Existing bounded-writer/output-cap constants to respect when adding the streaming emitter (verified)

```go
// native.go:18
const MaxStreamBytes = 64 * 1024
```
```go
// cgen.go:318
fmt.Fprintf(out, "#define LANG_OUTPUT_LIMIT 65536u\n#define LANG_EVENT_CAPACITY %du\n\n", capacity)
```

### Existing nonlocal-exit ledger pattern to generalize (verified, spike 005, `.planning/spikes/005-native-ffi-provenance-cleanup/native/nonlocal_exit.c`, in full)

```c
#include <setjmp.h>
#include <stdio.h>

static jmp_buf destination;
static int acquired = 0;
static int released = 0;

static void foreign_callback(void) { longjmp(destination, 1); }

static void call_with_resource(void) {
  acquired++;
  foreign_callback();
  released++;
}

int main(void) {
  if (setjmp(destination) == 0) call_with_resource();
  printf("{\"id\":\"nonlocal-exit\",\"valid\":%s,\"a\":%d,\"b\":%d,\"trace\":\"cleanup-skipped\"}\n",
         acquired == released ? "true" : "false", acquired, released);
  return acquired == released ? 0 : 6;
}
```
D-04-17's process-root pad and D-04-18's "ledger lives in `static` storage" directly generalize this: `static jmp_buf`/`static int acquired`/`static int released` here becomes a `static` cleanup ledger tracking every still-live acquisition, emitting `foreign.nonlocal_exit` plus one `resource.leaked` per entry, on the real generated multi-acquisition fixture.

### Existing evidence.Manifest additive-field precedent for `foreign_digest` (verified, `internal/compiler/evidence/evidence.go:49-66`)

```go
type Manifest struct {
	Schema           string   `json:"schema"`
	ID               string   `json:"id"`
	IDAlgorithm      string   `json:"id_algorithm"`
	SourceSchema     string   `json:"source_schema"`
	CoreSchema       string   `json:"core_schema"`
	ExecutionSchema  string   `json:"execution_schema"`
	DiagnosticSchema string   `json:"diagnostic_schema,omitempty"`
	CompilerIdentity string   `json:"compiler_identity"`
	ClangIdentity    string   `json:"clang_identity"`
	Target           string   `json:"target"`
	Flags            []string `json:"flags"`
	// ... ExecutionDigests, DigestClaim, KnownEscape (all present fields)
}
```
`foreign_digest` (D-04-12) should land here exactly like `DiagnosticSchema` did — a new `,omitempty` field, additive, no schema bump.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `core.OperationKind` has 5 members: `OpCopy, OpMove, OpBorrowShared, OpBorrowExclusive, OpReturn` (Phase 1-3) | Gains `OpForeignCall`, `OpFail`, `OpRelease`, `OpDefect` (exact names Claude's Discretion) | Phase 4 | Six dispatch sites across `interp`(×2)/`corevalidate`(×2)/`cgen`(×2) must add cases or their fail-closed `default:` correctly rejects every Phase 4 program |
| `originvalidate`/`pathoracle` walk only `core.OpReturn` | Must walk `{OpReturn, OpFail, OpDefect}` | Phase 4 (D-04-29) | Single highest-risk item in the phase — same defect shape that cost Phase 3 three gap-closure rounds (03-08/03-09/03-10) |
| `native.go`'s `validateExecution` hard-rejects any `Outcome.Kind != "returned"`, any nonzero exit, any `LiveResources` | Must gain an explicit expected-terminal-outcome axis | Phase 4 | Blocks *every* Phase 4 native-execution task until widened; regression test must confirm Phase 1-3 documents are still rejected on the old grounds |
| `cgen.go`'s `emitEventSupport` buffers all events, writes once at the end | Additive streaming emitter path required for the resource/foreign boundary; existing buffered emitters untouched | Phase 4 (D-04-20) | Without this, an aborting/`longjmp`ing process emits zero events, making SC4 unfalsifiable on exactly the paths SC3 is about |
| `core.DataType.Alternatives []string` (nullary-only ADTs) | Unchanged this phase — deliberately | Phase 4 decision (D-04-04), not yet | The payload-carrying-alternative debt is explicitly deferred to M002/Phase 6, recorded now rather than discovered later |

**Deprecated/outdated:** None — this phase does not remove or replace any existing Phase 1-3 mechanism; every change listed above is additive per D-13/D-04-23.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Exact `OperationKind` names (`OpForeignCall`, `OpFail`, `OpRelease`, `OpDefect`) will be the names the planner ultimately chooses | Code Examples, Architecture Patterns | Low — CONTEXT.md explicitly reserves exact naming as Claude's Discretion; using these placeholder names in research is only illustrative and does not constrain planning |
| A2 | `otool -l | grep -c __unwind_info` returning `0` on this exact host under `-fno-exceptions -fno-asynchronous-unwind-tables -fno-unwind-tables` (as recorded in spike 005 and re-confirmed by matching Clang/arch identity) will reproduce identically when Phase 4's actual generated C (not the spike's hand-written C) is compiled | Standard Stack, Anti-Patterns | Low-medium — this is why D-04-19 explicitly does NOT rely on this observation as the pass/fail signal; it is documented here only as corroborating context for why the `nm -u` approach was chosen instead |
| A3 | No third-party C static-analysis, fuzzing, or additional toolchain package will be needed to satisfy Phase 4's success criteria | Package Legitimacy Audit | Low — all required primitives (`_Static_assert`, `setjmp`/`longjmp`, `nm -u`, `abort()`) are C17-standard or already-present host tools; if a plan later needs e.g. a header-diffing library, the Package Legitimacy Gate must run at that time |

**If this table is empty:** N/A — see above; all three assumptions are low-risk and none blocks planning.

## Open Questions

1. **Exact spelling/shape of `core.ForeignContract` and the `lang.foreign/0` sidecar manifest fields**
   - What we know: D-04-12 requires three derived-from-one-source layers (header comment+`_Static_assert`, conformance TU, `lang.foreign/0` sidecar digest-bound via `evidence.Manifest`'s new `foreign_digest` field) and that the JSON is authoritative (C comments generated *from* it).
   - What's unclear: the precise field names/shape of `core.ForeignContract` itself (target layout, allocator identity, alias/capture, callback retention, unwind policy) — CONTEXT.md leaves "the exact `lang.foreign/0` field layout" as Claude's Discretion.
   - Recommendation: the planner should design this shape by first enumerating exactly what D-04-12's "target layout, initialized state, allocator identity, capture/retention, and unwind obligations" (FFI-01's own wording) require as fields, then verify each field round-trips through all three layers without any layer inventing a fact the JSON doesn't carry.

2. **Whether `check`'s own execution-admission gate (the one that currently rejects match-only/unexecutable shapes, referenced in Phase 3's `03-02` decision log entry: "checkLinear refuses a straight-line linear function whose parameter shape has no native execution lowering") needs a parallel gate for foreign-call shapes**
   - What we know: Phase 3 established a precedent of rejecting execution-inadmissible shapes at check time rather than discovering them at cgen/native time.
   - What's unclear: whether a `foreign C {}` declaration whose parameter/return types have no native lowering should be caught by the same existing gate, a widened version of it, or a new Phase-4-specific gate.
   - Recommendation: the planner should check whether the existing execution-admission gate already covers this by construction (since Phase 4's fixture type shape is Claude's Discretion and can simply be chosen to be within the existing lowerable set) — likely resolvable by fixture design alone, not new gate logic.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | Entire compiler build/test | ✓ | 1.24.0 darwin/arm64 | — |
| Apple Clang | Native C17 compilation, `-O0`/`-O3` differential | ✓ | 21.0.0 (clang-2100.1.1.101), arm64-apple-darwin25.6.0 | — |
| `nm` (llvm-nm) | D-04-19 non-unwind allowlist control | ✓ | GNU-nm-compatible, `/usr/bin/nm` | Per D-04-19, the control must itself degrade to `operational`/`tool_missing` (never silently pass) on a host without `nm` — this fallback path must be implemented and tested even though this host has `nm` |
| `otool` | Secondary/diagnostic tool only (spike 005 precedent, explicitly NOT the pass/fail gate per D-04-19) | ✓ | `/usr/bin/otool` present | Not required for the primary control; absence would not block D-04-19 |

**Missing dependencies with no fallback:** None — all required tools are present on this host.

**Missing dependencies with fallback:** None currently missing; the `nm`-absent fallback path (`operational`/`tool_missing`) must still be implemented defensively per D-04-19 since CI/other execution hosts are not guaranteed to match this research host.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go's built-in `testing` package (`go test`), the project's sole test framework since Phase 1 |
| Config file | none — `go.mod` at repo root (`go 1.24`) is the only config; per-phase bounded gate scripts (`scripts/verify-phase3.sh`, to be mirrored as `scripts/verify-phase4.sh` per D-04-21's carried method and the code_context note) |
| Quick run command | `go test ./internal/compiler/... -run <Test...>` scoped to touched packages |
| Full suite command | `env GOCACHE=<private tmp> go test -count=1 ./...` plus `-race` plus `scripts/verify-phase4.sh` (the bounded phase gate, copying `verify-phase3.sh`'s shape per the code_context note: "Phase 4 copies the shape ... it does not fork the frozen Phase 3 script and does not extract a shared helper") |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| RES-01 | Three-acquisition fixture releases exactly the acquired set, in reverse order, on success and on first/second-stage typed failure | differential (interp vs O0 vs O3) + mutation-kill (transpose two `OpRelease`, omit one `OpRelease`) | `go test ./internal/compiler/session/... -run TestResourceReleaseOrder` (name illustrative — planner assigns) | ❌ Wave 0 — new fixture + new test |
| FFI-01 | Generated header/conformance TU/sidecar manifest all derive from one `core.ForeignContract` and never drift | mutation-kill (field-order mutation in consumer TU must be a compile-time `_Static_assert` failure) | `go test ./internal/compiler/cgen/... -run TestForeignLayoutConformance` | ❌ Wave 0 |
| FFI-01 | Zero optimizer-visible attributes emitted; `emitted_attributes: []` present as a field | fail-closed control scan | `go test ./internal/compiler/session/... -run TestNoUnprovenAttributes` (`control:foreign.no_unproven_attributes`) | ❌ Wave 0 |
| SEM-03 | `let x = acquire ...` does not parse; `try`/`discard ... because` are the only consumers | parser rejection test | `go test ./internal/compiler/syntax/... -run TestFallibleOperationConsumers` | ❌ Wave 0 |
| SEM-03 | No `OpFail` route exists from a panic/defect path (structural, not a runtime test) | exhaustive-dispatch table-driven test (`control:kind.exhaustive_dispatch`) | `go test ./internal/compiler/core/... -run TestAllOperationKindsHandled` | ❌ Wave 0 — new `core.AllOperationKinds()` registry |
| SEM-03/RES-01 | Terminal outcome axis is a closed `value \| typed_failure \| defect` set; `cancelled` reserved but unconstructible | fail-closed control asserting no engine emits `cancelled` | `go test ./internal/compiler/interp/... -run TestNoCancelledOutcomeEmitted` | ❌ Wave 0 |
| FFI-01 (SC3) | Panic cannot cross the ordinary non-unwinding C boundary | `nm -u` allowlist control (fails closed to `operational`/`tool_missing` without `nm`) | `go test ./internal/compiler/native/... -run TestNoUnauditedUndefinedSymbols` (`control:foreign.unwind_forbidden`, name illustrative) | ❌ Wave 0 |
| FFI-01 (SC3) | Foreign nonlocal exit does not silently bypass Lang cleanup — detection, not prevention | native process-run test asserting `foreign.nonlocal_exit` + `resource.leaked` events + defect terminal record, using the widened `validateExecution` and streamed emitter | `go test ./internal/compiler/native/... -run TestNonlocalExitDetected` | ❌ Wave 0 — needs `native.go` widened first |
| RES-01/SEM-03 (SC4) | Interpreter and native executions agree on primary failure and cleanup events, extended to `typed_failure`/`defect` outcomes | differential (interp vs O0 vs O3) over widened `validateExecution` | `go test ./internal/compiler/evidence/... -run TestInterpNativeAgreementIncludingDefect` | ❌ Wave 0 |
| D-04-29 | `originvalidate`/`pathoracle` walk every terminator, not just `OpReturn` | mutation-kill (delete `OpFail` from either walker) required control | `go test ./internal/compiler/originvalidate/... -run TestWalksAllTerminators`; `go test ./internal/compiler/pathoracle/... -run TestWalksAllTerminators` | ❌ Wave 0 |
| D-04-24 | A program that `abort()`s or `longjmp`s is handled via `ProcessState.Sys().(syscall.WaitStatus)`, never a hardcoded `134` | native process-run test asserting SIGABRT via WaitStatus | `go test ./internal/compiler/native/... -run TestAbortSignalHandling` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** the scoped `go test ./internal/compiler/<touched-package>/... -run <Test...>` for the package(s) the task touched.
- **Per wave merge:** `env GOCACHE=<private tmp> go test -count=1 ./...` plus `go vet ./...` (matching the exact invocation pattern spike 005's own `README.md` "How to run" section documents, confirmed verbatim: `env GOCACHE=/private/tmp/... go test -count=1 ./...`).
- **Phase gate:** `scripts/verify-phase4.sh` (mirrored from `scripts/verify-phase3.sh`'s shape, per the code_context note explicitly forbidding forking the frozen Phase 3 script or extracting a shared helper) green, including all new required controls, before `/gsd-verify-work`.

### Wave 0 Gaps
- [ ] `testdata/phase4/` fixtures: a three-acquisition resource-acquisition program (success, first-stage-typed-failure, second-stage-typed-failure paths), a foreign declaration missing `unwind`/`nonlocal_exit` (must be refused), a nonlocal-exit probe program (analogous to spike 005's `native/nonlocal_exit.c` but through the real generated pipeline), a layout-mutation golden (frozen foreign TU + conformance TU pair).
- [ ] `core.AllOperationKinds()` registry — does not exist yet; required by `control:kind.exhaustive_dispatch` (D-04-22).
- [ ] `scripts/verify-phase4.sh` — does not exist yet; copies `scripts/verify-phase3.sh`'s shape per the explicit no-shared-helper, no-fork instruction.
- [ ] `native.go`'s widened `validateExecution` plus its own regression test confirming Phase 1-3 documents are still rejected on the old grounds — does not exist yet, and blocks every other native-execution-dependent Wave 0 test above.
- [ ] Frozen foreign TU (`native/` generated, analogous to `testdata/phase2/owned_transfer.golden.c`) wrapping libc `malloc`/`free` and returning a record by value — does not exist yet.

## Security Domain

### Applicable ASVS Categories

This phase is compiler-internals work, not a network-facing application; most ASVS categories (authentication, session management, access control) do not apply in their usual web-app sense. The categories below are reinterpreted for a compiler whose "untrusted input" is Lang source text and whose "trust boundary" is the audited C FFI edge.

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | No | N/A — no runtime identity in this phase |
| V3 Session Management | No | N/A |
| V4 Access Control | No | N/A |
| V5 Input Validation | Yes (reinterpreted) | Parser/checker rejection of malformed `foreign C {}` declarations (missing `unwind`/`nonlocal_exit` policy is a hard refusal, no default, per D-04-16), and of fallible-operation misuse (`let x = acquire ...` does not parse, per D-04-06) |
| V6 Cryptography | No | N/A — `sha256` content-identity digests (existing `evidence.go` use, `IDAlgorithm = "sha256-v1"`, confirmed) are identity binding, not a cryptographic security boundary; unchanged this phase |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| A generated binding silently trusting a foreign declaration as proof of safety | Tampering (of the safety claim itself) | Quarantine: the compiler never parses the foreign TU's header (D-04-10); all layout claims are proven by a separate, auditable conformance TU with `_Static_assert` (D-04-11), not trusted from the declaration |
| An unproven optimizer attribute (`restrict`, `nothrow`) producing miscompiled code under `-O3` | Tampering (silent semantic corruption) | Zero optimizer-visible attributes emitted this phase, enforced by a fail-closed control scanning all emitted C and the manifest's `emitted_attributes` field (D-04-13) |
| A foreign nonlocal exit (`longjmp`, signal, foreign `exit()`) silently bypassing resource cleanup, leaking or double-freeing a resource | Tampering / Denial of Service (resource leak) | Process-root `setjmp` landing pad + `static`-storage cleanup ledger emitting `foreign.nonlocal_exit`/`resource.leaked`, then terminating as a defect rather than attempting an unsafe release from indeterminate state (D-04-17/18) |
| Panic/unwind crossing an unaudited C ABI boundary, corrupting foreign-side state that expects no unwind | Tampering / Elevation of Privilege (undefined foreign-side behavior) | `nm -u` allowlist over the linked binary's undefined symbols — any new undefined symbol fails until a human adds it with rationale (D-04-19); this is the closest analogue to a supply-chain/dependency-allowlist control in this phase |
| Coordinated edit across the frozen fixture, the Lang `foreign` declaration, and the conformance TU's expectations, passing the gate on a wrong boundary | Repudiation (an author can construct false-but-self-consistent evidence) | Explicitly accepted as a residual limitation, not engineered around — named in CONTEXT.md as "QLT-01's documented source-to-core escape class, one layer down"; the mitigation is human code review of any PR touching all three artifacts together, not an automated control |

## Sources

### Primary (HIGH confidence)
- `internal/compiler/core/core.go` (read in full this session) — `DataType.Alternatives []string` (line 21), `OperationKind` constants (lines 128-134), additive-`omitempty` precedent (lines 151-161)
- `internal/compiler/interp/interp.go` (lines 75-149 read this session) — two exhaustive-dispatch switches with fail-closed `default:`
- `internal/compiler/corevalidate/corevalidate.go` (grep-confirmed switch locations, lines 738-797, 870-933)
- `internal/compiler/cgen/cgen.go` (grep-confirmed switch locations, lines 237-286, 470-520; `emitEventSupport`/`LANG_OUTPUT_LIMIT`, lines 310-318)
- `internal/compiler/native/native.go` (lines 240-254 read this session) — `validateExecution`'s current hard-reject shape, `MaxStreamBytes` constant (line 18)
- `internal/compiler/originvalidate/originvalidate.go` (grep-confirmed, line 103) and `internal/compiler/pathoracle/pathoracle.go` (grep-confirmed, line 247) — both walk only `core.OpReturn` today
- `internal/compiler/evidence/evidence.go` (lines 1-66 read this session) — `Manifest` struct, `DefaultFlags`, `IDAlgorithm`
- `.planning/spikes/005-native-ffi-provenance-cleanup/README.md` (read in full this session) and `.planning/spikes/005-native-ffi-provenance-cleanup/native/nonlocal_exit.c` (read in full this session) — validated prior art for the `setjmp`/`static`-ledger pattern, the `restrict`/`-O3` divergence, and the recorded host identity
- `.planning/phases/03-borrowed-views-and-cfg-lifetimes/03-DEBT.md` (read in full this session) — D-03-01/D-03-02 detail and closure history, directly informing D-04-25/26/27/28/29
- `.planning/phases/04-fallible-resources-and-c-boundary/04-CONTEXT.md` (read in full this session) — all 29 locked decisions
- `.planning/REQUIREMENTS.md`, `.planning/STATE.md` (read in full this session)
- `wiki/semantic-kernel-contract.md`, `wiki/native-and-low-level-profile.md`, `wiki/memory-reclamation-policy.md` (all read in full this session)
- Live host verification this session: `clang --version` → Apple clang 21.0.0 arm64-apple-darwin25.6.0; `go version` → go1.24.0 darwin/arm64; `nm`/`otool` present at `/usr/bin/nm`, `/usr/bin/otool`

### Secondary (MEDIUM confidence)
- [SEI CERT MSC22-C — Use the setjmp(), longjmp() facility securely](https://wiki.sei.cmu.edu/confluence/display/c/MSC22-C.+Use+the+setjmp(),+longjmp()+facility+securely) — corroborates C17 §7.13's "values of objects of automatic storage duration... are unspecified if... changed between the setjmp() invocation and longjmp() call", cited to ground D-04-18's stated C17 §7.13.2.1 claim
- [GNU nm documentation](https://sourceware.org/binutils/docs/binutils/nm.html) — `-u` flag semantics (display only undefined symbols), general confirmation of the mechanism D-04-19 builds on

### Tertiary (LOW confidence)
- None — every claim in this document is either grounded in a file read this session, a decision already locked in CONTEXT.md, or a MEDIUM-confidence secondary source above.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — no new dependencies; all tooling versions confirmed live on the actual host
- Architecture: HIGH — the "six sites" and "two independent admission layers" claims are grounded in actual `grep`/`Read` results against current source, not paraphrased from CONTEXT.md's prose
- Pitfalls: HIGH — each pitfall is either a verbatim CONTEXT.md decision rationale or a direct code-read finding (e.g., `validateExecution`'s current hard-reject)
- Package legitimacy: N/A — no external packages this phase
- Validation architecture: MEDIUM — test names/commands are illustrative (planner's to finalize); the requirement-to-test mapping and the mutation-kill obligations themselves are HIGH confidence (directly from CONTEXT.md's required controls)

**Research date:** 2026-09-04
**Valid until:** 2026-09-18 (14 days) — shorter than the default 30-day stable window because this research is tightly coupled to the exact current byte-state of `internal/compiler/...`; if Phase 4 planning/execution is delayed and other work lands on `main` first, the file/line citations above must be re-verified before being treated as authoritative.
