# Phase 29: Byte Views and the Adler-32 Library - Context

**Gathered:** 2026-10-11
**Status:** Ready for planning

<domain>
## Phase Boundary

Ordinary Schway source can take a read-only `ByteView` parameter, read its
length and the byte at an index (widened to U64 0..255) in a scalar loop that
carries only U64 and Bool state, and compute RFC 1950 Adler-32 over it. Every
byte read is dynamically bounds-checked on every engine and optimization level;
an out-of-range read exits 65 with stderr exactly
`schway: byte index out of bounds` and empty stdout. Interpreter and native runs
return the six pinned APP-10 literals. Each unsound neighbor shape is refused
with its existing stable code; the FLOW-02 back-edge boundary is unchanged.

Not in this phase: file input, the public `checksum FILE` route, the
65,536-byte cap case on the public route (Phase 31); modules (Phase 30).

</domain>

<decisions>
## Implementation Decisions

All decisions below were produced by advisor research (one researcher per gray
area, `minimal_decisive` tier) and **auto-followed** per the user's standing
"fan out, then auto-follow" preference. The ten M006 research decisions in
`.planning/research/SUMMARY.md` remain settled and are not repeated here
(no `U8`; `OpBytesLen`/`OpBytesAt`; no loan across a back edge; never-elided
dynamic checks; no per-byte events; 65,536 cap target; `B*65536` from 16
checked self-additions; pinned literals from Python `zlib.adler32` / Go
`hash/adler32`, test-only).

### Byte-read surface syntax
- **D-01:** Byte reads are spelled as name-recognized builtins:
  `byte_len(data)` → `OpBytesLen`, `byte_at(data, i)` → `OpBytesAt`. No `[`/`]`
  tokens, no `data.len` / method forms, no iterator form. — **Reversibility:**
  costly — surface is pinned by fixtures, formatter goldens and repair coverage;
  core ops are unchanged, so a later `[]` can be added as sugar lowering to the
  same ops without breaking these spellings.
- **D-02:** `byte_at` is the first two-argument call form. It is a special-cased
  scalar-expression atom (identifier text `byte_at` + `(` name `,` scalarExpr
  `)`), not general multi-argument call grammar. User functions stay
  one-parameter. `TokenComma` already exists; no lexer change.
- **D-03:** Parser changes: `scalarExprAtom` gains the two builtin branches;
  `scalarLetStart` admits `let x = byte_len(` / `let x = byte_at(` instead of
  rejecting every `ident(`; `scalarBodyStart` admits a `ByteView` parameter
  (today it is U64-gated).
- **D-04:** `byte_len` and `byte_at` are reserved names: check refuses a user
  definition of either, and wrong arity gets a named refusal following the
  `write_*` arity precedent. Formatter renders `byte_at(data, i)` with one space
  after the comma.

### Fixture byte delivery (pre-Phase 31)
- **D-05:** Input bytes reach the view through a **test-only foreign fixture
  owner** (working name `FixtureBytes`) declared in a `foreign C` block with the
  existing acquire / borrow / consume modes: acquire takes a U64 case id, a
  borrow-mode function yields the read-only `ByteView`, consume releases. Copy
  the Phase 23 pattern; **do not edit `examples/phase23`**.
- **D-06:** Fixture `main` is straight-line: acquire with a literal case id
  baked into the source, borrow shared, call `adler32(view)`, release, return
  the result. `schway run --engine=…` passes no input today and stays that way;
  no new CLI flag, no ByteView entry parameter, no byte-string literal syntax.
  One small fixture source per vector (generated from one template is fine):
  empty, one `0x00`, `a`, `Wikipedia`, 4,096×`0x00`, 5,552×`0xFF`, plus the OOB
  control (`byte_at(data, byte_len(data))`).
- **D-07:** Bytes are generated independently per engine: C in the native
  adapter under `examples/phase29/`, Go in the interpreter binding. A table test
  asserts both generators produce identical bytes. Expected values are pinned
  literals with the producing command and version recorded beside each.
  Wrong-modulus, off-by-one and signed-char mutants run against these fixtures.
- **D-08:** Refused-neighbor fixtures (SC4) reuse the same caller shape: borrow
  inside the loop, copy/store/return the view, release or move the owner inside
  the loop, owner or loan parameter on a loop function — each refused with its
  existing stable code, plus one seeded fault per peer.
- **D-09:** The fixture owner is test-only and must not be mistaken for, or
  reused as, the Phase 31 file-bytes owner.

### Spike sequencing and cap evidence
- **D-10:** No standalone `/gsd-spike`. The phase researcher settles research
  flags (a) ByteView through `OpCall`/originvalidate call-site derivation,
  (b) parser lookahead, (c) release-while-loan-live in check/corevalidate/
  pathoracle, (e) no evented stores in the loop, (f) delivery per D-05 — each as
  a pinned yes/no with file:line references in RESEARCH.md.
- **D-11:** Plan 29-00 is a Wave-0 gate with a written verdict that every later
  plan depends on: minimum real ByteView plumbing + the Adler-32 loop, then
  measure flag (d). If (a), (c) or (d) fails, stop and re-plan before plumbing
  spreads across all six consumers.
- **D-12:** Cap measurement protocol: workload 65,536×`0xFF` (expect
  `2006388466`) through the D-05 fixture path, not a hand-built proxy. Lanes:
  interpreter (plain and `go test -race`), native `-O0`, `-O3`, `-O3`+LTO with
  compile and run timed separately. 7 runs per lane, first discarded; record
  median and max; also 4,096 and 16,384 bytes to confirm linear scaling. Record
  toolchain, target and flags. Findings go to
  `.planning/spikes/008-byteview-cap/` (README table + MANIFEST row +
  VALIDATED/PARTIAL verdict); no wall time in golden files. Local macOS is
  primary; one hosted macOS and one hosted Linux run (existing CI job, `-race`
  included) are required to *keep* the cap; local evidence alone suffices only
  to *lower* it.
- **D-13:** Budget and lowering rule: slowest lane's max ≤ 10% of the tightest
  deadline — 3 s native (vs 30 s `defaultSubprocessTimeout`), 12 s interpreter
  (vs 2 min `RunCLITimeout`). Keep 65,536 unless a lane's max exceeds budget on
  any runner; then lower to the largest power of two whose linearly-scaled max
  fits. Never lower on a within-budget median, aesthetics, or `app verify`
  capacity. A lowering is recorded here as a reopening of research Decision 5,
  linked to the spike findings, and changes Phase 31's cap/cap+1 literals.
  — **Reversibility:** costly — a lowered cap propagates into Phase 31's
  pinned oversize matrix and the guide.

### Post-research decisions (2026-10-11, resolving 29-RESEARCH.md Open Questions 1-8)
Auto-followed per the standing preference; each takes the researcher's
recommendation unless noted.
- **D-14 (amends D-11/D-12):** Two-stage cap evidence. Stage A in plan 29-00
  uses harness-delivered bytes (a Go-injected `[]byte` seam for the
  interpreter, a disposable gate-only native driver labeled "proxy, gate-only")
  to prove flag (a), the classifier/place-gate changes, and an
  order-of-magnitude cost reading. Stage A triggers a re-plan only if the
  slowest lane is far above the D-13 budget, and it never sets the cap. Stage B
  is the last plan: the authoritative D-12 record through the real D-05 fixture
  owner, with hosted runs. Only Stage B may keep or lower the cap.
  *Planning note:* the gate is split across plans 29-00 (tracer) and 29-01
  (cgen proxy, Stage-A cost, `29-00-GATE-VERDICT.md`) for per-plan budget; the
  written verdict still precedes every other plan.
- **D-15 (amends D-06; keeps D-05):** The fixture owner is kept and budgeted as
  its own multi-plan workstream. The case id comes from the entry parameter:
  `fn main(case: U64) -> U64` with `try` acquire on `case`. One built artifact
  serves every vector (`schway app run ART -- <case>`), and the acquire keeps the
  Phase 23 shape, with the entry parameter as its source. Per-vector baked
  literals are dropped. The reason: D-06's rationale assumed `schway run` was
  the route, and it cannot run U64 entries at all (D-16). There are still no
  ByteView entry parameters, no byte-string literals and no `PathToken` reuse
  (D-09 separation holds).
- **D-16 (amends the ROADMAP runnable-gain wording):** `schway run --engine=*`
  stays unchanged (D-06). The runnable gain is met by
  `schway build FIXTURE --manifest M --output ART` plus `schway app run ART -- N`
  for native, and by the session interpreter API (the same route Phase 23 uses)
  for the interpreter. Record this as a dated note on the ROADMAP Phase 29
  runnable-gain line.
- **D-17:** `ByteView` has the ability set {`drop`} only; `copy`, `share`,
  `send` and `escape` are denied. Both independent ability derivations
  (`ability.Derive`, `corevalidate.deriveAbility`) must agree.
- **D-18 (amends SC1 text and the D-01/D-04 examples):** `data` is a lexer
  keyword, so the library parameter is named `bytes`:
  `adler32(bytes: ByteView) -> U64`, `byte_len(bytes)`, `byte_at(bytes, i)`.
  Add a dated note to ROADMAP SC1.
- **D-19 (SC3 for the interpreter):** The interpreter has no process exit, so
  "exits 65" means a defect outcome with reason `byte index out of bounds`, empty
  output, and the existing exit-class projection to 65 wherever the interpreter
  result is surfaced as a process exit. Native asserts the real exit 65 plus the
  exact stderr line.
- **D-20 (defect with a live owner):** A byte-index defect is abort-like. The
  process exits 65 without running the fixture owner's release; the OS reclaims
  it. The interpreter accepts a defect terminal with live resources for this
  defect class only. Every other terminal still rejects live resources
  (`interp.go:253`). Document this beside the OOB control.
- **D-21:** The `-O3` and `-O3 -flto` application-shell lanes for SC3 are a
  test-only seam, not a `native.Runner` knob. Receipts and recorded application
  flags stay `-O0`.

- **D-22 (amends D-08 / BYTE-03 / SC4 "existing stable code", dated 2026-10-11):**
  Research flag (c) found that no layer refuses a release while a view is
  still live. Phase 29 therefore adds a new code family:
  `check.release_while_view_live`, `core.release_while_view_live` and
  `pathoracle.release_while_view_live`. The family covers only release before
  the last use of the view, and each code is proven by a seeded fault per peer.
  Neighbors in the loop body keep their existing codes:
  - An `OpRelease` or `OpMove` of the owner placed in a loop block is pinned by
    seeded typed-core rows. Each row asserts the first existing code it
    produces when characterized at PLAN_BASE.
  - A loop-body neighbor that no existing code refuses falls under this
    amendment, and the plan must say so explicitly.
  Source-level rows that the parser refuses (`linearBody` admits no `while`) are
  documented as parse refusals, not ownership proofs. Every other D-08 neighbor
  (copy, store, or return the view; borrow inside the loop; an owner or loan
  parameter on a loop function) gets the source refusal plus one seeded fault
  per peer that has a matching law.
- **D-23:** The Stage-A STOP multiplier is 4× the D-13 budget, a planner choice
  confirmed under the standing auto-follow preference. The Go generator stays
  in `internal/compiler/testsupport`, which is test-only and consistent with D-07
  and D-09. `FixtureBytes` takes the abilities {drop, share, send, escape} on the
  FileByteOwner model.

- **D-24 (contingency for 29-09):** At PLAN_BASE, row R2 (an owner release
  after the view's last use but still on the loop cycle) or row M1 may be
  accepted. That outcome is a safety finding, but the phase does not stop.
  - **R2 accepted:** the D-22 code family is extended to cover "release of
    the owner on a loop cycle". The 29-09 SUMMARY records the extension, and
    29-11 adds it to the dated SC4 note.
  - **M1 accepted:** the executor stops and re-plans. Moving an owner has had
    an existing code since Phase 9.

### Claude's Discretion
- Exact `ast.Expr` kind names, fixture owner type name, adapter file layout,
  case-id numbering, and whether fixture sources are generated or hand-written.
- Repair-tool coverage depth for the new arity/reserved-name refusals.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase scope and requirements
- `.planning/ROADMAP.md` §Phase 29 — goal, 4 success criteria, research flags (a)–(f)
- `.planning/REQUIREMENTS.md` — BYTE-01, BYTE-02, BYTE-03, APP-10 (and APP-11/12 for what is deferred)

### Settled milestone decisions
- `.planning/research/SUMMARY.md` — ten resolved decisions (esp. 2 byte repr, 3 loops/loans, 4 bounds/events, 5 cap), pinned literals, anti-patterns, six-consumer / four-classifier list
- `.planning/research/ARCHITECTURE.md` — pipeline and component placement
- `.planning/research/PITFALLS.md` — drift across consumers, common-mode answer error
- `.planning/research/STACK.md`, `.planning/research/FEATURES.md` — supporting detail

### Spike conventions and prior findings
- `.planning/spikes/CONVENTIONS.md` — spike harness and recording rules (applies to D-12)
- `.planning/spikes/MANIFEST.md` — add the 008 row
- `.claude/skills/spike-findings-ai-lang/SKILL.md` — ownership kernel, loan liveness, FFI contract findings
- `.planning/spikes/002-cfg-edge-last-use`, `.planning/spikes/007-loan-across-branch` — loan/back-edge evidence

### Precedent code and fixtures
- `examples/phase23/` — foreign acquire/borrow/consume pattern to copy (frozen, do not edit)
- `examples/sum_to_n.schway` — scalar `var`/`while` loop shape
- `examples/checksum.schway` — frozen provisional frontier (rejected iterator form; retired in Phase 32)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/compiler/syntax/parser.go` `isWriteCallStart` — name-recognized builtin precedent for D-01/D-02.
- `internal/compiler/syntax/parser.go` `scalarLetStart` (~L339) and `scalarBodyStart` (~L358) — the gates D-03 widens.
- `internal/compiler/syntax/lexer.go` — `TokenComma` already exists; no `[`/`]`.
- Phase 23 foreign resource modes (`mode: "acquire" | "borrow" | "consume"`, `release:`) in `internal/compiler/native/foreign_resource.go` and bindings — template for the fixture owner.
- `internal/compiler/exitcode` — single exit-class table from Phase 28; 65 is the checked-failure code for OOB.

### Established Patterns
- Six consumers (check, corevalidate, originvalidate, pathoracle, interp, cgen) re-derive facts independently; four copied scalar classifiers (`check`, `corevalidate.hasScalarCFG`, `originvalidate.originHasScalarCFG`, `cgen.hasScalarOperations`) must all learn the new ops, plus exhaustive-dispatch fixtures.
- Corpus goldens committed before behavior changes; byte-identical generated C for untouched programs.
- Functions take exactly one parameter; extra call arguments are arity refusals.

### Integration Points
- `cmd/schway/main.go` `run --engine=` (~L112) — unchanged; fixtures bake inputs in source.
- `internal/compiler/native/native_app.go` `defaultSubprocessTimeout` (30 s) and the `RunCLITimeout` test deadline — budgets for D-13.

</code_context>

<specifics>
## Specific Ideas

- Pinned APP-10 literals: empty → `1`, `0x00` → `65537`, `a` → `6422626`, `Wikipedia` → `300286872`, 4,096×`0x00` → `268435457`, 5,552×`0xFF` → `4052720524`; cap workload 65,536×`0xFF` → `2006388466`.
- OOB stderr exactly `schway: byte index out of bounds`, stdout empty, exit 65, at interpreter and native `-O0`/`-O3`/`-O3`+LTO.
- A `0x80`/`0xFF` fixture must read `128`/`255` (unsigned widening; signed-char mutant must fail).

</specifics>

<deferred>
## Deferred Ideas

- `[]` indexing sugar over `OpBytesAt` — only with a future arrays milestone.
- Byte-string literal syntax (and a repeat form) — not needed; revisit with a literals/arrays milestone.
- Public input for `schway run` (flag or ByteView entry) — Phase 31 public route owns input.
- Phase 31: `checksum FILE`, file-bytes owner and `examples/checksum` adapter, bounded limit+1 read, oversize/missing/non-regular outcomes, cap case on the public route, independent interpreter file read.

</deferred>

---

*Phase: 29-byte-views-and-the-adler-32-library*
*Context gathered: 2026-10-11*
