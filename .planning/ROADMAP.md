# Roadmap: Schway

## Milestones

- ✅ **M001 — Source-to-Native Semantic Spine** — Phases 1–6 (shipped 2026-09-07) — [archive](milestones/M001-ROADMAP.md)
- ✅ **M002 — Interprocedural Semantic Spine** — Phases 07–13 (shipped 2026-09-14) — [archive](milestones/M002-ROADMAP.md)
- ✅ **M003 — Computation and Honest Instruments** — Phases 14–20 (shipped 2026-09-26) — [archive](milestones/M003-ROADMAP.md)
- ✅ **M004 — Native Emission Ownership and Resource Discharge** — Phases 21–25 (completed 2026-10-02) — [archive](milestones/M004-ROADMAP.md)
- ✅ **M005 — Practical Computation** — Phases 26–27.3 (shipped 2026-10-10) — [archive](milestones/M005-ROADMAP.md)
- ✅ **M006 — Bounded Byte Utility and Local Modules** — Phases 28–32.1 (shipped 2026-10-12) — [archive](milestones/M006-ROADMAP.md)
- 🚧 **M007 — Bitwise Checksum and Composable Scalar Helpers** — Phases 33–38 (in progress)

## Overview

M007 makes Schway able to express a real bitwise checksum from small, reusable
scalar helpers, and makes the language cheaper for an agent to author. Phase 33
pays the carried M006 debt in one capped phase: it bounds the native capture
read before allocation, closes the 32.1 review warnings, and replaces the
phase-specific freeze scripts with one milestone-agnostic gate. Phase 34 sets a
diagnostic floor for agents (line:col with a snippet, one diagnostic per syntax
recovery point, `surface.unsupported_*` hints, and a `schway surface --json`
manifest) without changing semantics. Phase 35 adds U64 `^ & |`, literal-amount
`<< >>` and hex literals, and re-expresses `adler32` with `<< 16` as a free
differential witness. Phase 36 lets a scalar function call another scalar
function. Phase 37 delivers the `crc32` module, the `crc32` program, a second
consumer, the interpreter file-bytes CLI route, and the clean-checkout guide.
The pre-receipt audit inserts Phase 37.1 to close three carried terminal-outcome
blockers. Phase 38 binds hosted two-host receipts once, after that closure and
a fresh `/gsd-audit-milestone`.

The current requirements are in [REQUIREMENTS.md](REQUIREMENTS.md). The
[M007 research synthesis](research/M007/SUMMARY.md) records the consumer choice
(CRC-32/ISO-HDLC, not POSIX `cksum`), the compiler cost estimates and the M006
lessons; treat them as settled unless a phase discussion records evidence to
reopen one. The Go 1.24 standard-library host, interpreter oracle, single
C17/Clang native route and application runner remain the implementation path;
no new dependency or backend is required. The living
[product roadmap](PRODUCT-ROADMAP.md) and
[language-maturity record](LANGUAGE-MATURITY.md) hold the ranked successors and
the current refusal frontier.

## Phases

<details>
<summary>✅ M006 Bounded Byte Utility and Local Modules (Phases 28–32.1) — SHIPPED 2026-10-12</summary>

- [x] Phase 28: Exit Classes, Verify Report, and Entry Pruning (8/8 plans) — completed 2026-10-11
- [x] Phase 29: Byte Views and the Adler-32 Library (13/13 plans) — completed 2026-10-11
- [x] Phase 30: Local Modules (10/10 plans) — completed 2026-10-12
- [x] Phase 31: Bounded File Checksum and Second Consumer (10/10 plans) — completed 2026-10-12
- [x] Phase 32: Clean-Checkout Guide and Final Receipts (11/11 plans) — completed 2026-10-12
- [x] Phase 32.1: Address tech debt: CLI exit classes and module-aware routes (INSERTED, 11/11 plans) — completed 2026-10-12

Requirements: [milestones/M006-REQUIREMENTS.md](milestones/M006-REQUIREMENTS.md) · Audit: [milestones/M006-MILESTONE-AUDIT.md](milestones/M006-MILESTONE-AUDIT.md) · Phase artifacts: `milestones/M006-phases/`

</details>

<details>
<summary>✅ M005 Practical Computation (Phases 26–27.3) — SHIPPED 2026-10-10</summary>

- [x] Phase 26: Checked Scalar Sum (5/5 plans) — completed 2026-10-03
- [x] Phase 27: Exact FizzBuzz and Cross-Host Evidence (10/10 plans) — completed 2026-10-10
- [x] Phase 27.1: Build-route origin validation and receipt-gate docs (INSERTED) — completed 2026-10-10
- [x] Phase 27.2: Straight-line scalar let admission R1 (INSERTED) — completed 2026-10-10
- [x] Phase 27.3: App verify capacity abort and guide clean-tree outputs W1, W4 (INSERTED, 4/4 plans) — completed 2026-10-10

Requirements: [milestones/M005-REQUIREMENTS.md](milestones/M005-REQUIREMENTS.md) · Audit: [milestones/M005-MILESTONE-AUDIT.md](milestones/M005-MILESTONE-AUDIT.md) · Phase artifacts: `milestones/M005-phases/`

</details>

### 🚧 M007 Bitwise Checksum and Composable Scalar Helpers (Phases 33–38)

Phase IDs continue from M006's Phase 32.1. Each phase names its runnable gain.
Phase discussions choose surface syntax and implementation detail while
preserving the acceptance behavior below.

**Plan caps (M006 lesson):** Phase 33 is capped at 5 plans. Every other phase
targets about 8 plans; table-driven per-peer seeded faults go in one plan, not
one plan per peer. A phase that needs more plans than its cap records a
plan-count checkpoint and splits into an inserted decimal phase rather than
growing.

**Evidence rule (M006 lesson):** No phase before 38 binds hosted receipts.
Phases 33–37 close on local automated evidence and the existing recurring CI
jobs. Phase 38 binds the hosted ubuntu and macOS receipts once, after
`/gsd-audit-milestone` and after any audit gap-closure phase it inserts.

**Ordering decision (diagnostics before bitwise):** Phase 34 runs before
Phase 35 so that every later phase is authored against line:col diagnostics,
recovery-point dedup and the manifest meta-test. To avoid churn, DX-22's hint
set in Phase 34 covers only constructs that stay absent after M007 (`&&`, `||`,
`!`, arity > 1, if without else, Bool literal tails, subtraction). Bitwise
operators and hex literals are not given `surface.unsupported_*` hints in
Phase 34, because Phase 35 admits them. The one listed DX-22 construct that
depends on shifts existing, a dynamic shift amount, gets its named refusal in
Phase 35 under BIT-02, registered in the Phase 34 manifest as a
`surface.unsupported_*` code with a one-line hint. The manifest meta-test from
Phase 34 then forces Phases 35 and 36 to register every newly admitted
construct, so the manifest cannot drift.

- [x] **Phase 33: Carried Debt, Bounded Capture Read, and Generic Freeze Gate** - The native capture file is read through a limit+1 bound before allocation, the 32.1 and Phase 29 review findings are closed or dispositioned, Phase 28 gains VALIDATION/SECURITY, and one milestone-agnostic freeze/receipt gate replaces `check-phase32-*`. (completed 2026-10-13)
- [x] **Phase 34: Agent Diagnostics Floor and Surface Manifest** - Diagnostics carry line:col and a caret snippet, a syntax error yields one diagnostic, absent constructs get `surface.unsupported_*` hints, and `schway surface --json` prints a meta-tested manifest of the admitted surface. (completed 2026-10-13)
- [x] **Phase 35: Bitwise Operators, Literal Shifts, and Hex Literals** - U64 `^ & |`, literal-amount `<< >>` and hex literals run with identical answers on both engines at every optimization level, and `adler32` uses `b << 16` with byte-identical pinned answers. (completed 2026-10-14)
- [x] **Phase 36: Scalar Helper Calls** - A scalar function calls another scalar function (one U64/Bool argument, including imported module functions), with independent peer derivation and callee defects surfacing on both engines. (completed 2026-10-15)
- [x] **Phase 37: CRC-32 Consumer, Interpreter File Route, and Guide** - `crc32 FILE` prints exact CRC-32/ISO-HDLC from an ordinary-source module, a second consumer imports it, `schway run --engine=interpreter` runs file-bytes programs from the CLI, and an executable guide is test-run. (completed 2026-10-15)
- [ ] **Phase 37.1: Close gap: legacy terminal outcomes (INSERTED)** - Close CR-03/04/06 with independently admissible terminal evidence and an explicit compatibility decision before changing frozen no-call C.
- [ ] **Phase 38: Milestone Receipts** - Hosted ubuntu and macOS receipts for every M007 and earlier pinned program are bound once to the final M007 codedigest, after the milestone audit.

## Phase Details

### Phase 33: Carried Debt, Bounded Capture Read, and Generic Freeze Gate

**Goal**: The carried M006 debt is paid before any feature work: a child-written capture file can no longer force an unbounded host allocation, the open review warnings are closed or recorded with reasons, Phase 28 has its missing validation and security records, and later milestones reuse one freeze/receipt gate instead of copying phase-specific scripts.

**Depends on**: Completed M006 (Phase 32.1)

**Requirements**: QLT-13, QLT-14, QLT-15, EVD-16

**Success Criteria** (what must be TRUE):

1. An oversize evidence capture file makes the native app report the named capacity outcome, and it does so without reading the whole file: the capture is read through a reader bounded at limit+1 bytes before any allocation. A focused Go test pins the outcome at exactly the limit and at limit+1, and a seeded mutant that restores the unbounded `os.ReadFile` read is caught by that test.
2. Running the freeze gate from the repository root and from a subdirectory gives the same verdict, pinned by a subdirectory test (32.1 WR-01). `isAbortLikeDefect` accepts only the abort-like defect reasons it names, with the live owner checked, and a mutant that widens it is caught by a focused test (32.1 WR-02).
3. Each of 32.1 IN-01..03, the coverage-`kind` pin, and Phase 29 WR-02/WR-03 is either fixed with a covering test or recorded in a written disposition with its reason, and the planning-doc meta-tests (inventory, groundedness, validation lifecycle) pass with those records in place. `milestones/M006-phases/` Phase 28 has a VALIDATION.md and a SECURITY.md built from existing evidence, accepted by the validation-lifecycle meta-test.
4. One milestone-agnostic freeze/receipt gate takes the milestone's program list and codedigest as input and replaces the `check-phase32-*` scripts. Its armed controls reproduce every existing refusal (revision mismatch, digest mismatch, missing program, malformed receipt), it accepts the archived M006 receipts when given the M006 inputs, and a static test finds no remaining `check-phase32-*` caller.

**Runnable gain**: The M006 receipts re-validate through the generic gate given only the M006 program list and codedigest, and an oversize capture is refused by name without a full-file allocation.

**Research flag**: Standard. Capped at 5 plans and binds no hosted receipt. WR-02 touches byte-op parity: pin the current interpreter/native abort-like parity rows before narrowing the predicate. The generic gate must keep the M006 `check-phase32-*` refusal behavior byte-for-byte before the scripts are deleted.

**Plans**: 5/5 plans complete (cap 5)

Plans:
**Wave 1**
- [x] 33-01-PLAN.md — docs-only M006 receipt reopen (first commit), bounded capture read, IN-02 fixture and coverage-kind pin (wave 1)

**Wave 2** *(blocked on Wave 1 completion)*
- [x] 33-02-PLAN.md — WR-02: parity pins on both routes, shared defect-reason constants, narrowed abort-like predicate and killed mutant (wave 2)
- [x] 33-03-PLAN.md — generic gate step 1: codedigest flags, M006 manifest, check-freeze.sh / check-receipts.sh, differential and archived-M006 tests, WR-01, IN-03 (wave 2)

**Wave 3** *(blocked on Wave 2 completion)*
- [x] 33-04-PLAN.md — generic gate step 2: test renames, goldens, old scripts deleted, producer reads the manifest, no-caller test (wave 3)

**Wave 4** *(blocked on Wave 3 completion)*
- [x] 33-05-PLAN.md — living-claim check by milestone label, Phase 28 VALIDATION/SECURITY, Phase 29 verify, disposition records (wave 4)

### Phase 34: Agent Diagnostics Floor and Surface Manifest

**Goal**: An agent drafting Schway gets one precise, locatable diagnostic per real cause and can ask the toolchain what surface is admitted, instead of guessing from cascading, offset-only or misleading errors.

**Depends on**: Phase 33 (generic gate and debt cleared; no semantic dependency)

**Requirements**: DX-20, DX-21, DX-22, DX-23

**Success Criteria** (what must be TRUE):

1. A text diagnostic shows `file:line:col` and a one-line source snippet with a caret under the span, and each JSON diagnostic span adds 1-based `line` and `col` next to the unchanged byte offsets. Golden tests pin both forms, including multi-byte UTF-8 and CRLF lines and a span on the last line without a newline, and a corpus test shows no diagnostic code changed.
2. A pinned corpus of agent-style drafts (including the CRC-32 and word-counter scratch drafts from research) yields one diagnostic per real syntax cause, with no follow-on `expected }` or `expected declaration` cascade. A test pins the before/after diagnostic counts per draft.
3. Each construct that stays absent after M007 (`&&`, `||`, `!`, more than one parameter, if without else, a Bool literal tail, subtraction) is refused with its own `surface.unsupported_*` code and a one-line hint. A test per construct asserts the code and hint and asserts that the previous misleading parse error (for example `expected )` for arity 2, `syntax.expected_equal` for a Bool tail) no longer appears.
4. `schway surface --json` prints a versioned manifest of types, operators, statements, builtins, call rules and limits, and two runs give byte-identical output. A meta-test fails, with an armed control for each direction, when an admitted construct or a `surface.unsupported_*` code is missing from the manifest, or when the manifest names a construct that is not admitted.

**Runnable gain**: `schway check` on a broken agent draft prints one `file:line:col` diagnostic with a caret and a hint, and `schway surface --json` tells the agent what it may write.

**Research flag**: Standard. No semantics change: generated C for every existing program stays byte-identical, pinned by the corpus golden. Bitwise operators and hex literals receive no `surface.unsupported_*` hint here (Phase 35 admits them); the dynamic-shift-amount code lands with BIT-02 in Phase 35 (see the ordering decision above). Diagnostic JSON gains fields only; existing consumers and repair patches must keep parsing it.

**Plans**: 8 (8/8 complete)
**Wave 1**
- [x] 34-01-PLAN.md

**Wave 2** *(blocked on Wave 1 completion)*
- [x] 34-02-PLAN.md

**Wave 3** *(blocked on Wave 2 completion)*
- [x] 34-03-PLAN.md
- [x] 34-04-PLAN.md

**Wave 4** *(blocked on Wave 3 completion)*
- [x] 34-05-PLAN.md

**Wave 5** *(blocked on Wave 4 completion)*
- [x] 34-06-PLAN.md

**Wave 6** *(blocked on Wave 5 completion)*
- [x] 34-07-PLAN.md

**Wave 7** *(blocked on Wave 6 completion)*
- [x] 34-08-PLAN.md

### Phase 35: Bitwise Operators, Literal Shifts, and Hex Literals

**Goal**: Ordinary Schway scalar code can use U64 `^`, `&`, `|`, literal-amount `<<` and `>>`, and hex literals, with one defined meaning on every engine and optimization level, no C17 shift undefined behavior, and no new defect path.

**Depends on**: Phase 34 (manifest meta-test and `surface.unsupported_*` family)

**Requirements**: BIT-01, BIT-02, BIT-03, BIT-04

**Success Criteria** (what must be TRUE):

1. U64 `^`, `&` and `|` give the literal expected answers in the interpreter and in native code at `-O0`, `-O3` and `-O3` with LTO. Check, corevalidate and originvalidate independently derive scalar transfers, and table-driven wrong-result transfer mutants per op per value-deriving peer are caught. Independently authored engine-answer vectors catch engine wrong-result controls. The execution peer independently validates each op's event-free causal admission and catches fabricated step/defect events; causal admission is not numeric validation.
2. `<<` and `>>` with literal amounts 0, 31, 32 and 63 give pinned answers on both engines at every optimization level, and `<<` drops high bits as the manifest and guide state. An amount of 64 or more, or a non-literal amount, is refused at check, core and origin with a named `surface.unsupported_*` code and hint registered in the manifest; a mutant removing each refusal is caught, and a static scan of generated C for the corpus finds no shift whose amount is not a literal below 64.
3. Hex literals (`0x…`, case-insensitive digits) are admitted wherever decimal U64 literals are and evaluate identically on both engines. Overflowing and malformed hex is refused with named codes, and a formatter golden shows the source spelling preserved.
4. `examples/checksum/adler32.schway` uses `b << 16` in place of 16 doublings, and the existing pinned answers plus the `checksum` and `adler_low` outputs stay byte-identical on both engines, proven by the existing Phase 29/31 answer tests passing unchanged.
5. Every program without bitwise ops, shifts or hex literals keeps byte-identical generated C, pinned by a corpus golden committed before the change, and the Phase 34 manifest meta-test passes with every new operator and refusal code registered.

**Runnable gain**: A Schway program computes `x ^ (y << 3) & 0xFF` and prints the same answer from `schway run --engine=interpreter` and `--engine=native`, and `checksum` keeps its outputs with a one-line shift.

**Dated discussion/research amendments (2026-10-14):** Criterion 1 originally required “one wrong-result mutant per op per peer” including executionpeer. Its current `Validate` contract checks observed causal structure, with no entry scalar value, so Phase 35 D-18 assigns value-transfer and engine controls to actual value knowers and per-op event-free/forged-event controls to that causal peer; no result replay, event or defect is added. The runnable expression also needs bounded transparent grouping. Public conformance `run` requires one parameter and already binds a Byte fixture; a narrowly admitted unused-Byte scalar entry supplies this witness without a new CLI input/default U64 rule. Both public routes expose the same outcome value (`264` for x=0x100/y=0x21); bare decimal application output is pinned separately through existing application runners. Byte operand use stays refused, and existing U64 input-required outcomes remain unchanged. See 35-CONTEXT.md D-07/D-18 and 35-RESEARCH.md for source evidence and the original mismatch.

**Research flag**: Standard. Literal-only shift amounts keep the ops event-free (precedent: the static-zero refusal in `scalar.go`) and avoid C17 §6.5.7 UB. Non-defecting ops join the event-free lists only. Hex is lexer-only.

**Plans**: 8 (8/8 complete)
- [x] 35-01-PLAN.md
- [x] 35-02-PLAN.md
- [x] 35-03-PLAN.md
- [x] 35-04-PLAN.md
- [x] 35-05-PLAN.md
- [x] 35-06-PLAN.md
- [x] 35-07-PLAN.md
- [x] 35-08-PLAN.md

### Phase 36: Scalar Helper Calls

**Goal**: A scalar function can be factored into small helpers: it calls another scalar function with one U64 or Bool argument and uses the result, including calls into imported modules, with the same defect behavior on both engines and every peer re-deriving the call independently.

**Depends on**: Phase 35 (helpers will be written with bitwise ops; manifest records the call rule)

**Requirements**: CALL-01, CALL-02

**Success Criteria** (what must be TRUE):

1. A scalar function that calls another scalar function with one U64 or Bool argument and uses the result in an expression or assignment prints the pinned answer on both engines; the same holds for a call to a function imported from a local module. A recursive or mutually recursive scalar call is still refused by the call graph with its existing code.
2. An overflow, remainder-by-zero or byte-bounds defect inside a callee surfaces with the callee's reason and exit 65 on both engines, pinned by one test row per defect kind.
3. Check, corevalidate, originvalidate and the execution peer each derive scalar calls independently, and a table-driven test shows mutants that drop the argument, drop the result, or drop a callee defect are each caught.
4. Every program without scalar-to-scalar calls keeps byte-identical generated C, pinned by a corpus golden committed before the change, and the manifest meta-test passes with the call rule (arity 1, no recursion) recorded.

**Runnable gain**: A scalar `crc_step`-style helper is called from a loop function and from an importing module, and both engines print the same result.

**Research flag**: Moderate. Arity stays exactly 1 (`core.Function.Parameter`), so the core IR JSON schema does not change. Cost lies in the callee defect path through cgen evidence reservation, interpreter defect events and the execution peer.

**Plans**: 8 (8/8 complete)
**Wave 1**
- [x] 36-01-PLAN.md

**Wave 2** *(blocked on Wave 1 completion)*
- [x] 36-02-PLAN.md

**Wave 3** *(blocked on Wave 2 completion)*
- [x] 36-03-PLAN.md

**Wave 4** *(blocked on Wave 3 completion)*
- [x] 36-04-PLAN.md
- [x] 36-05-PLAN.md

**Wave 5** *(blocked on Wave 4 completion)*
- [x] 36-06-PLAN.md

**Wave 6** *(blocked on Wave 5 completion)*
- [x] 36-07-PLAN.md

**Wave 7** *(blocked on Wave 6 completion)*
- [x] 36-08-PLAN.md

### Phase 37: CRC-32 Consumer, Interpreter File Route, and Guide

**Goal**: A developer can run `crc32` on one file and get the exact CRC-32/ISO-HDLC from an ordinary-source module written with bitwise ops and helper calls, a second application imports that module unchanged, both engines run file-bytes programs from the CLI with identical results, and a clean checkout reproduces it all from the guide.

**Depends on**: Phase 36 (helper calls), Phase 35 (bitwise ops and shifts)

**Requirements**: APP-13, MOD-07, APP-14, DX-24

**Success Criteria** (what must be TRUE):

1. `crc32 FILE`, built from an ordinary-source `crc32` module using a bit-at-a-time loop with no table, prints bare decimal: empty → `0`, one `0x00` byte → `3523407757`, `123456789` → `3421780262`, and 65,536 × `0xFF` → its pinned answer. A 65,537-byte file gets the existing oversize outcome, and a missing or unreadable file gets the shared outcome. Every answer comes from Python `zlib.crc32` with its generating command recorded, and armed controls (wrong polynomial, non-reflected variant, missing final xor, off-by-one loop bound, library mutation) are each observed to fail.
2. A second consumer imports `crc32` from the module without copying it and prints its pinned answer on both engines; a single-copy scan finds one copy of the CRC loop, and the library mutation flips both applications' outputs.
3. `schway run --engine=interpreter` runs `crc32` and `checksum` from the CLI and gives the same stdout, stderr and exit code as native for every pinned case, proven by one shared cross-engine case matrix test through the public CLI route rather than the Go-level `RunWithFileBytes` seam.
4. `examples/crc32/README.md` states the expected results, the 65,536-byte limit, each error outcome, and that this is CRC-32/ISO-HDLC and not POSIX `cksum` (CRC-32/CKSUM). A clean-checkout guide test, built like `TestPhase32GuideCleanCheckout`, runs its commands in a temp copy of the tracked files and matches exit code, stdout and stderr.

**Runnable gain**: `crc32 FILE` and the second consumer run on both engines from the CLI, and a fresh clone reproduces them by following the guide.

**Research flag**: Moderate. Reuses the M006 file cap, ByteView, module and file adapter unchanged. The interpreter file-bytes CLI route is new public surface: record its exit classes in the Phase 28 exit table. The unreadable-file control uses a named skip when running as root. This is the last code-moving phase before the milestone audit.

**Plans**: 9 (9/9 complete)
- [x] 37-09-PLAN.md

**Wave 1**
- [x] 37-01-PLAN.md

**Wave 2** *(blocked on Wave 1 completion)*
- [x] 37-02-PLAN.md

**Wave 3** *(blocked on Wave 2 completion)*
- [x] 37-03-PLAN.md
- [x] 37-04-PLAN.md

**Wave 4** *(blocked on Wave 3 completion)*
- [x] 37-05-PLAN.md
- [x] 37-06-PLAN.md
- [x] 37-07-PLAN.md

**Wave 5** *(blocked on Wave 4 completion)*
- [x] 37-08-PLAN.md

### Phase 37.1: Close gap: legacy terminal outcomes (INSERTED)

**Goal:** A legacy child defect, a propagated modeled typed failure, and an admitted no-call explicit defect each produce the correct terminal outcome and independently admissible evidence, closing audit blockers CR-03, CR-04 and CR-06 before the M007 freeze.

**Requirements**: None

**Audit closure**: CR-03, CR-04, CR-06; carried CALL-01/CALL-02 boundaries and EVD-17 freeze prerequisites. Canonical requirement ownership stays in Phases 36 and 38; this closure adds no requirement or language surface. The machine requirements field is deliberately empty so phase completion cannot mark Phase 38 receipts complete.

**Depends on:** Phase 37

**Success Criteria** (what must be TRUE):

1. Real interpreter/peer runs close every active frame on a whole-execution child defect in legacy and dynamic call-identity modes while preserving ordinary one-frame returns (CR-03). Reached forged-terminal and post-terminal controls remain refused.
2. The peer accepts the real checked-source modeled-failure path `foreign.failed → OpFail`, independently binding operation/source, checked failure type, represented error links and root outcome kind (CR-04, dated correction below). Child failures retain one-frame/error-edge behavior; wrong-source/type/outcome-kind and relabeled-clean-return controls fail. No acceptance follows from an event's spelling or an unsubstantiated `OpReturn` probe.
3. Checker-admitted no-call `OpDefect` programs produce structured native defect evidence with the exact reason bytes at O0, O3 and O3-LTO (CR-06), accepted by the independent peer. Application exit 65, empty stdout and ordinary-reason stderr stay fixed; reached control/NUL reason corruption is corrected with byte-counted output under the dated compatibility decision. The repaired dynamic CR-05 path stays green. The audit diagnostic becomes a permanent meaningful regression.
4. Before changing frozen no-call C, record a dated, source-grounded compatibility decision and limit any baseline amendment to affected rows, preserving unaffected bytes and original history. Refresh affected verification evidence after source changes, preserve completed UAT and historical summaries, record closure requirement dispositions after its gates, and rerun the milestone audit before Phase 38 planning or freeze.

**Runnable gain**: The three named legacy programs yield correct terminal outcomes and independently accepted evidence on their applicable routes, removing the carried semantic blockers to final receipts.

**Discussion amendment — 2026-10-16:** [37.1-CONTEXT.md](phases/37.1-close-gap-legacy-terminal-outcomes/37.1-CONTEXT.md) D-04..07 preserve the original `OpReturn` criterion and correct CR-04 to the witnessed checked-source `foreign.failed → OpFail` path. Fresh model-only Phase 24 probes reached that path; a modified `OpReturn` candidate failed core validation. The historical audit's synthetic return probe is not admitted-source evidence, and no universal raw-core impossibility is claimed. D-10 records three newly reached native application reason-byte mismatches; D-12..13 authorize exact additive path/route C amendments before emitter edits while preserving original goldens and unaffected bytes. Current causal evidence does not prove numerical values, absent failure payloads or actual CFG branch choices. Original audit, summaries and completed UAT remain historical.

**Admission correction — 2026-10-16:** Context D-21 records a further reached reason-field boundary: malformed UTF-8 inside quotes is admitted, then changed by existing JSON serialization. Enforce the existing UTF-8 validity rule for defect reasons at source and independent raw-core admission; retain valid UTF-8/control/NUL reasons and lossless source recovery. This scoped correction replaces invalid-byte normalization with an existing-vocabulary refusal, adds no wire format or language construct, and qualifies “exact reason bytes” to admitted valid UTF-8 text.

**Scope and evidence**: [M007 audit](M007-MILESTONE-AUDIT.md) and its durable integration probes own the diagnosis. The peer probes were synthetic core/events, so source-admitted interpreter/peer regressions are required here. The native probe reached the real failure but is not a permanent red/green regression. Passing frozen C bytes alone does not prove native defect parity. Phase 36 WR-02 remains separate nonblocking comment debt with its existing next-semantic-edit owner. This is one bounded safety closure; no hosted binding occurs here.

**Transition review — 2026-10-16**: The audit's [three useful next capabilities](M007-MILESTONE-AUDIT.md#three-useful-next-capabilities) remain ranked: audited same-source two-host CRC/checksum receipts, proposed M008 `wc FILE` / `longest_line FILE`, then bounded fixed-schema `key=value` checking. That table names each program, blocker, smallest slice, guarantee changes, evidence/debt, owner and reprioritizing observation. Phase 37.1 now owns terminal closure; Phase 38 retains final receipts. The audit's [dormant seed dispositions](M007-MILESTONE-AUDIT.md#dormant-seed-dispositions) remain applicable to current source/consumers: z1g (query/helper navigation), trz (capture/CRC failure diagnosis), z1s (formatter/checker/source review) all retain dormant status; existing evidence explains these failures and no measured query gap or contested source-form choice triggers an experiment. The local Dashbit appraisal is historical. No implementation or capability claim changed in this insertion.

**Plans:** 0 plans

Plans:
- [ ] TBD (run $gsd-plan-phase 37.1 to break down)

### Phase 38: Milestone Receipts

**Goal**: Every native claim made in M007 and earlier milestones is bound to hosted ubuntu and macOS receipts at the final M007 source, taken once after the milestone audit so no later fix silently invalidates them.

**Depends on**: Phase 37.1 and a fresh `/gsd-audit-milestone` run after closure

**Requirements**: EVD-17

**Success Criteria** (what must be TRUE):

1. Hosted ubuntu and macOS receipts cover every M007 program (`crc32`, the second `crc32` consumer, the bitwise and helper-call fixtures) and every earlier pinned program (`checksum`, `adler_low`, the Phase 29 fixture and OOB control, `sum_to_n`, FizzBuzz), with every pinned case per host.
2. The Phase 33 generic freeze/receipt gate accepts the receipts given the M007 program list and the final M007 codedigest, and rejects them, through its armed controls, when the codedigest or revision differs from the checked-out source.
3. The receipts are bound in the last plan of the milestone, after `/gsd-audit-milestone`; no per-phase hosted receipt binding exists in Phases 33–37, and the ingest plan touches only `.planning/`.

**Runnable gain**: The milestone's native claims are reproducible on both hosts at the shipped revision, and any later code commit reopens them.

**Research flag**: Standard. Hosted runs need explicit user approval for push and dispatch (one human checkpoint line naming the freeze commit); everything else is automated. Hosted runs take about 15–22 minutes and have failed transiently, so the plan allows one re-dispatch at the same freeze commit before treating a failure as real.

**Plans**: Not planned (expected 3: freeze, push-and-dispatch checkpoint, `.planning`-only ingest; planning follows the M007 audit)

## Progress

**Execution order:** 33 → 34 → 35 → 36 → 37 → initial audit (`gaps_found`) → 37.1 → `/gsd-audit-milestone` → 38

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 33. Carried Debt, Bounded Capture Read, and Generic Freeze Gate | 5/5 | Complete    | 2026-10-13 |
| 34. Agent Diagnostics Floor and Surface Manifest | 8/8 | Complete    | 2026-10-13 |
| 35. Bitwise Operators, Literal Shifts, and Hex Literals | 8/8 | Complete    | 2026-10-14 |
| 36. Scalar Helper Calls | 8/8 | Complete    | 2026-10-15 |
| 37. CRC-32 Consumer, Interpreter File Route, and Guide | 9/9 | Complete    | 2026-10-15 |
| 37.1. Close gap: legacy terminal outcomes (INSERTED) | 0 planned | Not started | - |
| 38. Milestone Receipts | 0 planned | Not started | - |

**Coverage:** 19/19 M007 requirements have exactly one phase owner. DX-22 is
owned by Phase 34; its dynamic-shift-amount item is delivered by Phase 35's
BIT-02 refusal, which registers as a `surface.unsupported_*` code in the
Phase 34 manifest (see the ordering decision). EVD-17 is its own phase rather
than a final plan of Phase 37, so `/gsd-audit-milestone` can run against a
complete Phase 37 and any inserted gap-closure phase lands before the receipts.
