# Roadmap: Schway

## Milestones

- ✅ **M001 — Source-to-Native Semantic Spine** — Phases 1–6 (shipped 2026-09-07) — [archive](milestones/M001-ROADMAP.md)
- ✅ **M002 — Interprocedural Semantic Spine** — Phases 07–13 (shipped 2026-09-14) — [archive](milestones/M002-ROADMAP.md)
- ✅ **M003 — Computation and Honest Instruments** — Phases 14–20 (shipped 2026-09-26) — [archive](milestones/M003-ROADMAP.md)
- ✅ **M004 — Native Emission Ownership and Resource Discharge** — Phases 21–25 (completed 2026-10-02) — [archive](milestones/M004-ROADMAP.md)
- ✅ **M005 — Practical Computation** — Phases 26–27.3 (shipped 2026-10-10) — [archive](milestones/M005-ROADMAP.md)
- 🚧 **M006 — Bounded Byte Utility and Local Modules** — Phases 28–32 (in progress)

## Overview

M006 makes Schway useful for a small byte-oriented utility built from a
reusable local library. Phase 28 fixes the exit classes, the `app verify`
capacity report, and entry resolution, so every later refusal reaches users
through one table and a library with unused functions can run. Phase 29 adds a
read-only, bounds-checked byte view and runs an ordinary Adler-32 library over
it. Phase 30 moves that library into its own local module behind an explicit
export list. Phase 31 delivers the bounded file `checksum` application and a
second application that imports the same library, with the exact result
matrix and independent evidence. Phase 32 closes the milestone with the
clean-checkout guide, retirement of the provisional `examples/checksum.schway`
frontier, and hosted macOS/Linux receipts at the final revision.

The current requirements are in [REQUIREMENTS.md](REQUIREMENTS.md). The
[research synthesis](research/SUMMARY.md) records ten resolved decisions; treat
them as settled unless a phase discussion records evidence to reopen one. The
existing Go 1.24 standard-library host, interpreter oracle, single C17/Clang
native route, and application runner remain the implementation path; no new
dependency or backend is required. The living
[product roadmap](PRODUCT-ROADMAP.md) and
[language-maturity record](LANGUAGE-MATURITY.md) retain the ranked successors
and current refusal frontier.

## Phases

<details>
<summary>✅ M005 Practical Computation (Phases 26–27.3) — SHIPPED 2026-10-10</summary>

- [x] Phase 26: Checked Scalar Sum (5/5 plans) — completed 2026-10-03
- [x] Phase 27: Exact FizzBuzz and Cross-Host Evidence (10/10 plans) — completed 2026-10-10
- [x] Phase 27.1: Build-route origin validation and receipt-gate docs (INSERTED) — completed 2026-10-10
- [x] Phase 27.2: Straight-line scalar let admission R1 (INSERTED) — completed 2026-10-10
- [x] Phase 27.3: App verify capacity abort and guide clean-tree outputs W1, W4 (INSERTED, 4/4 plans) — completed 2026-10-10

Requirements: [milestones/M005-REQUIREMENTS.md](milestones/M005-REQUIREMENTS.md) · Audit: [milestones/M005-MILESTONE-AUDIT.md](milestones/M005-MILESTONE-AUDIT.md) · Phase artifacts: `milestones/M005-phases/`

</details>

### 🚧 M006 Bounded Byte Utility and Local Modules (Phases 28–32)

Phase IDs continue from M005's Phase 27.3. Each phase names its runnable gain.
Phase discussions choose surface syntax and implementation detail while
preserving the acceptance behavior below.

- [x] **Phase 28: Exit Classes, Verify Report, and Entry Pruning** - Source refusals share one exit class across commands and engines, programs with unused functions run, and `app verify` keeps the named capacity reason at every input. (completed 2026-10-10)
- [ ] **Phase 29: Byte Views and the Adler-32 Library** - Ordinary Schway source reads a bounds-checked byte view in a scalar loop and computes RFC 1950 Adler-32 with pinned answers on both engines.
- [ ] **Phase 30: Local Modules** - A program imports the Adler-32 library from another local file through an explicit import and export list, with closed resolution and named module-graph refusals.
- [ ] **Phase 31: Bounded File Checksum and Second Consumer** - A caller runs `checksum FILE` with an exact result and error matrix, and a second application imports the same library without copying it.
- [ ] **Phase 32: Clean-Checkout Guide and Final Receipts** - A developer can follow the documented commands from a clean checkout, and hosted macOS/Linux receipts bind the final M006 revision.

## Phase Details

### Phase 28: Exit Classes, Verify Report, and Entry Pruning

**Goal**: A developer gets the same exit class and refusal code for the same source from every command and engine, can run a program that contains functions `main` never calls, and sees `app verify` name the capacity limit instead of losing it.

**Depends on**: Completed M005 (Phase 27.3)

**Requirements**: DX-17, DX-18, EVD-13

**Success Criteria** (what must be TRUE):

1. For one corevalidate-refusal fixture, `schway check`, `schway build`, and `schway run` each exit 2 with the same refusing code; for the `foreign_origin_omitted` fixture, `schway run --engine=interpreter` and `schway run --engine=native` both exit 2 with `core.foreign_origin_omitted` (today native exits 3 `native.tool_failure`). Tool, IO, and timeout failures exit 3, and engine disagreement exits 4.
2. Every exit code is read from one named table; a static test finds no duplicated exit-code literals in session, cmd, or native, and the existing 65 (checked failure) and 69 (event capacity) contracts keep their values.
3. A program with a function unreachable from `main` passes `schway check` and also builds and runs with its expected output, instead of failing with `callgraph.entry_ambiguous` at exit 3. A genuinely ambiguous entry exits 2 as a source refusal. Generated C for every program whose functions are all reachable stays byte-identical, pinned by a corpus golden committed before the change.
4. `schway app verify examples/sum_to_n.schway` still passes n = 14. At n = 15, 172, 173, and 1,000, each case is `unsupported` with a diagnostic starting `native.event_capacity_exhausted`. A set of 16 exhausted cases reports every case by name, stays within the report size limit, and never counts as success.

**Runnable gain**: A library file with unused functions now builds and runs, and `app verify` on `sum_to_n` at n = 1,000 names the capacity reason instead of reporting an unnamed oversize failure.

**Research flag**: Standard. Enumerate and pin every test, guide script, and hosted gate that asserts an exit code before changing any (PITFALLS 13); never renumber 65 or 69. N3 (duplicated capacity literal) is folded into the named table.

**Plans**: 8/8 plans complete (28-08 closed UAT gap G-28-3)

Plans:
**Wave 1**
- [x] 28-01-PLAN.md — Safety net: corpus entry/exit-class manifest and exit-pin inventory before any change (wave 1)

**Wave 2** *(blocked on Wave 1 completion)*
- [x] 28-02-PLAN.md — One named exit-code table, app-run pins, static no-literal scan, repair mirror pin (wave 2)
- [x] 28-03-PLAN.md — App verify outcome digest, schema /2, ID bound and size budget (wave 2)

**Wave 3** *(blocked on Wave 2 completion)*
- [x] 28-04-PLAN.md — `main` precedence, pruning through one resolveEntry, ambiguous entry exits 2, peer unreachability (wave 3)

**Wave 4** *(blocked on Wave 3 completion)*
- [x] 28-05-PLAN.md — Deliberate pin flips over frozen goldens, class-manifest changes, census (wave 4)

**Wave 5** *(blocked on Wave 4 completion)*
- [x] 28-06-PLAN.md — Admission before any tool; same exit class on check, build and both engines (wave 5)

**Wave 6** *(blocked on Wave 5 completion)*
- [x] 28-07-PLAN.md — App verify source refusals exit 2; README exit table; living-docs transition (wave 6)

**Gap closure** *(UAT G-28-3)*
- [x] 28-08-PLAN.md — Relative-path `run --engine=native` reaches Check (drop build-root path arithmetic); relative rows across all four surfaces

### Phase 29: Byte Views and the Adler-32 Library

**Goal**: Ordinary Schway source can read a read-only byte view in a scalar loop, with one defined meaning on every engine and a dynamic bounds check that is never removed, and computes RFC 1950 Adler-32 over it.

**Depends on**: Phase 28 (new refusals surface through the single exit-class table)

**Requirements**: BYTE-01, BYTE-02, BYTE-03, APP-10

**Success Criteria** (what must be TRUE):

1. An `adler32(data: ByteView) -> U64` function written in ordinary Schway, using only checked `+`, `<`, `==`, and `% 65521`, returns the pinned literals in both the interpreter and native runs: empty → `1`, one `0x00` byte → `65537`, `a` → `6422626`, `Wikipedia` → `300286872`, 4,096 × `0x00` → `268435457`, 5,552 × `0xFF` → `4052720524`.
2. Byte length and byte-at have one defined meaning across check, corevalidate, originvalidate, pathoracle, the interpreter, and C17 emission; each peer derives the facts independently, and a fixture containing `0x80` and `0xFF` reads `128` and `255` (unsigned widening to U64).
3. Reading an index at or beyond the length exits 65 with stderr exactly `schway: byte index out of bounds` and empty stdout, in the interpreter and in native code at `-O0`, `-O3`, and `-O3` with link-time optimization; no engine or optimization level removes the check.
4. A scalar loop function that takes a read-only byte view and carries only U64 and Bool state is admitted. Each unsound neighbor is refused with its existing stable code: a borrow inside the loop, copying, storing, or returning the view, releasing or moving an owner inside the loop, and owner or loan parameters on a loop function. The FLOW-02 back-edge refusal boundary is unchanged.

**Runnable gain**: `schway run --engine=interpreter` and `schway run --engine=native` on the Adler-32 library fixture print the same pinned checksum for each input vector, and the out-of-bounds control fails identically on both.

**Research flag**: Spike first. Before planning, settle: (a) a ByteView argument through `OpCall` in check's caller path and originvalidate's call-site derivation; (b) parser lookahead for byte length and indexing in scalar bodies; (c) release-while-loan-live refused by check, corevalidate, and pathoracle; (d) interpreter and native cost of 65,536 iterations, which confirms or lowers the 65,536-byte cap on recorded evidence; (e) no evented stores in the loop; (f) how the fixture's literal bytes reach the byte view before the Phase 31 file adapter exists (research proposes synthetic input to the conformance runner). Surface syntax for byte reads is decided at discussion.

**Plans**: 3/13 plans executed

Plans:
**Wave 1**
- [x] 29-00-PLAN.md — Wave-0 gate tracer: neighbor pins, ByteView {drop}, byte_len/byte_at through check, both peers and the interpreter with harness bytes, interpreter Stage-A vectors (wave 1)

**Wave 2** *(blocked on Wave 1 completion)*
- [x] 29-01-PLAN.md — Wave-0 gate expansion: cgen both shells, gate-only native proxy at -O0/-O3/-O3+LTO, Stage-A cost reading, written verdict 29-00-GATE-VERDICT.md (wave 2)

**Wave 3** *(blocked on Wave 2 completion)*
- [x] 29-02-PLAN.md — Byte-read surface: arity codes, reserved builtin names, let-first ByteView bodies, formatter goldens (wave 3)
- [ ] 29-03-PLAN.md — Independent peer laws with seeded faults, loop-function neighbor refusals plus per-peer seeded neighbor faults (D-22), no per-byte events (wave 3)
- [ ] 29-04-PLAN.md — OOB evidence: schema-2 reason, execution peer arm, interpreter boundary, bounds-check ordering in both C shells (wave 3)
- [ ] 29-05-PLAN.md — Test-only fixture bytes adapter and independent Go generator, proven byte-identical (wave 3)

**Wave 4** *(blocked on Wave 3 completion)*
- [ ] 29-06-PLAN.md — Fixture owner caller shape in check, corevalidate and originvalidate; adler32_fixture checks clean (wave 4)

**Wave 5** *(blocked on Wave 4 completion)*
- [ ] 29-07-PLAN.md — Native runnable gain: pathoracle replay, session admission, fixture-main emitter, build + app run CLI (wave 5)
- [ ] 29-08-PLAN.md — Interpreter runnable gain: RunWithFixtureBytes, D-19 OOB, D-20 abort-like byte-index defect (wave 5)

**Wave 6** *(blocked on Wave 5 completion)*
- [ ] 29-09-PLAN.md — Release-while-view-live laws (D-22 code family) in check, corevalidate and pathoracle; owner release/move in a loop block pinned per peer; loop source rows as parse refusals (wave 6)
- [ ] 29-10-PLAN.md — Pinned literals on both engines at -O0/-O3/-O3+LTO, mutants killed, OOB control lanes, per-lane sanitizer evidence and README (wave 6)

**Wave 7** *(blocked on Wave 6 completion)*
- [ ] 29-11-PLAN.md — Closure: dated ROADMAP notes (incl. D-22 SC4 amendment), register reconciliation, maturity and product roadmap, transition check (wave 7)

**Wave 8** *(blocked on Wave 7 completion)*
- [ ] 29-12-PLAN.md — Spike 008 Stage-B cap record; strict budget only via a dispatch input, hosted macOS and Linux runs; keep or lower decision (wave 8)

### Phase 30: Local Modules

**Goal**: A Schway program can import another local source file as a module through an explicit import and export list, with resolution closed under the entry directory and every module-graph error refused by name.

**Depends on**: Phase 29 (the Adler-32 library is the module worth importing)

**Requirements**: MOD-03, MOD-04, MOD-05

**Success Criteria** (what must be TRUE):

1. An entry program that imports `adler32` from a separate module file, whose `module` header matches its path under the entry directory, prints the same pinned Phase 29 answers on both engines. Imports through `..`, an absolute path, or a symlink that leaves the entry directory are refused; no search path exists.
2. Using a name that is not in the module's `export` list is refused with a stable code, and corevalidate re-derives visibility independently. Cross-module calls are admitted from the signature table alone: a guard test shows that changing a callee body without changing its signature does not change the caller's admission.
3. Self-import, length-2, and length-3 cycles, a missing module, a duplicate module, a header/path mismatch, and a name collision are each refused with their own named diagnostic attributed to the importing source. A diagnostic raised inside an imported file names that file.
4. A diamond import loads and emits the shared module once. Every program without imports produces byte-identical generated C, pinned by a corpus golden committed before the change, and C names for module functions are injective.

**Runnable gain**: The Adler-32 program from Phase 29 runs with the library in its own module file, imported rather than copied, and gives the same pinned answers.

**Research flag**: Deeper research plus a spike. Settle `os.Root` symlink and macOS case-insensitivity behavior, the `import` and `export` surface and formatter, the reach of `Diagnostic.Unit` (protocol, debug map, human output, repair patches), injective C name mangling, and module count/size/depth caps with exact-N and N+1 controls. Keep a plan-count checkpoint: this phase has the widest blast radius in M006. Modules are whole-program composition; do not describe them as separate compilation or as closing B1/DX-06.

**Plans**: TBD

### Phase 31: Bounded File Checksum and Second Consumer

**Goal**: A caller can run `checksum` on one file through the public application route and get an exact Adler-32 result or a named error, and a second application imports the same library without copying it, with every answer checked against an independent oracle.

**Depends on**: Phase 30 (module-qualified IDs and the importable library), Phase 29

**Requirements**: APP-11, APP-12, MOD-06, EVD-14

**Success Criteria** (what must be TRUE):

1. `schway app run` on the built `checksum` application with one file path prints bare decimal plus a newline: an empty file → `1`, a one-byte `0x00` file → `65537`, `Wikipedia` → `300286872`, exactly 65,536 × `0xFF` bytes → `2006388466`. The 65,536-byte case passes on every engine through the public route.
2. A file of 65,537 bytes or more is refused with its own named oversize outcome. Missing, unreadable, and non-regular files (a directory, a FIFO, a device) share one named outcome. Every error leaves stdout empty and writes fixed stderr text that never contains the path. The owned buffer is released exactly once on every success and error path, and the size check uses a bounded limit+1 read rather than the reported file size.
3. A second application imports the same `adler32` module and prints a different pinned result for the same file. The repository holds one copy of the library loop, a library mutation flips both applications' pinned outputs, and both build identities and evidence records hash every module file in sorted order.
4. Every pinned answer comes from Python `zlib.adler32` with its generating command recorded. Each reached control is observed to fail: a wrong modulus, wrong initial state, swapped halves, an off-by-one loop bound, a `-fsigned-char` mutant on `0x80`/`0xFF` fixtures, a deleted bounds check observed by ASan, each refused loop neighbor, and the library mutation.

**Runnable gain**: `checksum FILE` and the second consumer run as ordinary applications from the shared library module, with exact success and error results.

**Research flag**: Moderate. Choose the exit code for each file-error outcome and record it in the Phase 28 exit-class table; extend public CLI input from bounded U64 to one path through `app run`. The file adapter is a new copy-local generalization under `examples/checksum/`; the Phase 23 adapter, its receipts, and its observer tests stay untouched. The unreadable-file control uses a named skip when running as root.

**Plans**: TBD

### Phase 32: Clean-Checkout Guide and Final Receipts

**Goal**: A developer can reproduce every M006 program from a clean checkout using documented commands, and the milestone's native claims are bound to hosted macOS and Linux receipts at the final revision.

**Depends on**: Phase 31

**Requirements**: DX-19, EVD-15

**Success Criteria** (what must be TRUE):

1. From a clean checkout, the documented commands build and run `checksum`, the second module consumer, and the `adler32` library. The guide states the expected results, the 65,536-byte limit, and each error outcome, and a test executes those commands and compares their output with the guide.
2. The provisional `examples/checksum.schway` refused frontier is retired with a dated amendment, and the frontier test no longer pins it.
3. Hosted macOS and Linux receipts are taken at the final M006 revision and cover the M006 programs and the carried-forward `sum_to_n` and FizzBuzz programs. The receipt gate rejects a receipt whose revision differs from the checked-out source, so a later code change reopens them. This is the last plan of the milestone.

**Runnable gain**: A fresh clone reproduces `checksum`, the second consumer, and the library results by following the guide.

**Research flag**: Standard. Run a public-route audit before the receipt plan so the receipts bind the final source. Hosted runs need explicit user approval for push and dispatch.

**Plans**: TBD

## Progress

**Execution order:** 28 → 29 → 30 → 31 → 32

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 28. Exit Classes, Verify Report, and Entry Pruning | 8/8 | Complete    | 2026-10-11 |
| 29. Byte Views and the Adler-32 Library | 3/13 | In Progress | - |
| 30. Local Modules | 0/TBD | Not started | - |
| 31. Bounded File Checksum and Second Consumer | 0/TBD | Not started | - |
| 32. Clean-Checkout Guide and Final Receipts | 0/TBD | Not started | - |

**Coverage:** 16/16 M006 requirements have exactly one phase owner. MOD-06
(second consumer) is owned by Phase 31 rather than Phase 30, because its
library-mutation control flips the `checksum` application's pinned output and
its consumers are path-input applications; both depend on the Phase 31 file
adapter and public path input.
