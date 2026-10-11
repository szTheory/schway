---
verification_role: proposal
---

# Phase 29: Byte Views and the Adler-32 Library - Research

**Researched:** 2026-10-11
**Domain:** Go 1.24 stdlib-only compiler (syntax / check / corevalidate / originvalidate / pathoracle / interp / cgen / native / session); C17 via Clang; no new dependencies
**Confidence:** HIGH on every repo fact below (each read from source or probed with a built binary this session); MEDIUM on cost estimates and on three design recommendations flagged in Open Questions; interpreter/native loop cost deliberately NOT measured (plan 29-00 owns flag (d)).

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

All decisions below were produced by advisor research (one researcher per gray
area, `minimal_decisive` tier) and **auto-followed** per the user's standing
"fan out, then auto-follow" preference. The ten M006 research decisions in
`.planning/research/SUMMARY.md` remain settled and are not repeated here
(no `U8`; `OpBytesLen`/`OpBytesAt`; no loan across a back edge; never-elided
dynamic checks; no per-byte events; 65,536 cap target; `B*65536` from 16
checked self-additions; pinned literals from Python `zlib.adler32` / Go
`hash/adler32`, test-only).

#### Byte-read surface syntax
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

#### Fixture byte delivery (pre-Phase 31)
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

#### Spike sequencing and cap evidence
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

### Claude's Discretion
- Exact `ast.Expr` kind names, fixture owner type name, adapter file layout,
  case-id numbering, and whether fixture sources are generated or hand-written.
- Repair-tool coverage depth for the new arity/reserved-name refusals.

### Deferred Ideas (OUT OF SCOPE)
- `[]` indexing sugar over `OpBytesAt` — only with a future arrays milestone.
- Byte-string literal syntax (and a repeat form) — not needed; revisit with a literals/arrays milestone.
- Public input for `schway run` (flag or ByteView entry) — Phase 31 public route owns input.
- Phase 31: `checksum FILE`, file-bytes owner and `examples/checksum` adapter, bounded limit+1 read, oversize/missing/non-regular outcomes, cap case on the public route, independent interpreter file read.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| BYTE-01 | Schway source can read the length of a read-only byte view and read the byte at an index, as an unsigned U64 value from 0 to 255; checking, independent validation, interpretation, C17 emission give one defined meaning. | Flag (b) parser change list; "Everything that must learn OpBytesLen/OpBytesAt" tables; ByteView representation section; C emission design (unsigned char, `(uint64_t)` widen). |
| BYTE-02 | Every byte read is bounds-checked at run time on every engine and optimization level; OOB is exit 65, stderr exactly `schway: byte index out of bounds`, empty stdout. | "OOB failure path today" section (interp event + app-shell `fputs`/`exit(65)` + conformance-shell defect terminal); the missing -O3/LTO application lane; interpreter "exit 65" is a projection, not a process. |
| BYTE-03 | A scalar loop function takes a read-only byte view and iterates carrying only U64 and Bool state; every unsound neighbor refused with its existing stable code; FLOW-02 unchanged. | Neighbor-refusal table with codes probed on today's binary; flag (c) verdict; scalar fixed-point place gates that must be taught the single ByteView parameter place. |
| APP-10 | `adler32` in ordinary Schway computes RFC 1950 Adler-32 over a byte view using checked `+`, `<`, `==`, `% 65521`; interpreter and native equal the six pinned literals. | Pinned literals re-verified with Python 3.14.4 `zlib` and Go 1.24.0 `hash/adler32` this session; grammar traps (no parentheses, precedence, `data` keyword); flag (e) event-free loop; fixture delivery verdict (f). |
</phase_requirements>

## Summary

Phase 29 adds one new type shape (`ByteView`), two new core operations (`OpBytesLen`, `OpBytesAt`), a two-argument builtin atom in the scalar expression grammar, and a pure-Schway `adler32` library function. The scalar-loop machinery that makes this cheap already exists and works: a straight-line `main` can already call a scalar CFG function containing `%` and run it natively for 100,000 iterations with no event-capacity problem when the loop uses only `var`/assignment [VERIFIED: built binary this session, `app run -- 100000` printed `11449`, run time 0.00 s]. The checker, both peer validators and both engines each re-derive scalar facts independently (four deliberately copied classifiers), so the new operations must be added to every copy; nothing is shared, by design.

**The finding that most changes planning is flag (f), not (a)/(b)/(c).** The Phase 23/24/25 "foreign owner" machinery is not a reusable pattern parameterized by a symbol table: it is a family of closed, exact-shape matchers hard-coded to the names `PathToken`, `FileByteOwner`, `schway_file_byte_*`, `AcquireError`, `UseError`, `libc_malloc`, a three-operation body, and `length == 1` C text, repeated in check, corevalidate, originvalidate, pathoracle, cgen, interp, and session. D-05 ("copy the Phase 23 pattern") therefore means writing a *new bespoke owner shape across seven packages*; Phase 23 itself took seven plans. D-05 is feasible, but it is a Phase-23-sized sub-effort inside a phase that also carries two new ops across the same packages. The design is not blocked, but plan 29-00 as literally specified (minimum plumbing, then measure D-12 "through the D-05 fixture path") cannot be small. See flag (f) and Open Question 1.

Two real blockers for flag (a) and (c) were found and are bounded: (a) `originvalidate` will mark `adler32` not Callable (so `check` refuses every call to it with `core.callee_not_callable`) until its scalar fixed point learns the ByteView parameter place; (c) there is no release-while-loan-live law in any layer today (loan liveness is consulted only for moves and borrow conflicts; releases are compiler-inserted and pinned by exact operation order), so the three-peer refusal must be authored as new independent laws and proven only by seeded faults. Neither threatens the design.

**Primary recommendation:** Keep the locked design. Make plan 29-00 prove (a), the new classifier/place-gate changes, and an order-of-magnitude cost reading using harness-delivered bytes (Go-injected `[]byte` for the interpreter; a gate-only driver for native), then take the D-12 record through the real D-05 fixture owner at the end of the phase. Budget the fixture owner as its own multi-plan workstream.

## Research Flag Verdicts (D-10 / D-11)

| Flag | Verdict | Today's tree | Needs (concrete) |
|------|---------|--------------|------------------|
| (a) ByteView through `OpCall` (check caller path + originvalidate call-site derivation) | **NO as-is; fixable, not design-threatening.** | `originvalidate` makes `adler32` un-Callable; `check` then refuses the call. | Teach `originScalarFixedPointWithFailure` the one ByteView parameter place and the two ops; no `Mode` plumbing needed for admission. Prove in 29-00. |
| (b) Parser lookahead for `byte_len`/`byte_at` | **NO today; small change.** | `byte_len(` is refused at the `(`; `data` is a keyword. | `scalarExprAtom` branches (parser.go:528); `scalarLetStart`/`scalarBodyStart` only matter for `let`-first bodies. Rename the parameter (not `data`). |
| (c) Release-while-loan-live refused by check, corevalidate, pathoracle | **NO as a semantic law in any layer; YES only as accidental fail-closed exact-shape refusal.** | `OpRelease` is never compared to loan liveness; only `OpMove` is. | New independent law + named code per layer, proven by seeded faults. Not a source-level hazard (release is compiler-inserted). |
| (e) No evented stores in the loop | **YES.** | `var`/assignment lower to `OpScalarStore` (no event in interp, cgen, or the evidence-size bound); `let` lowers to `OpCopy` (one event per execution). | Library uses `var` + assignment only; add a structural test and a `let`-in-loop negative control. New ops must record no event. |
| (f) Fixture bytes via D-05 foreign fixture owner, copying Phase 23 | **FEASIBLE but NOT a copy; large.** | Phase 23/24 shapes are exact-shape, name-pinned matchers in 7 packages; interpreter side is a one-byte *modeled* outcome that never holds real bytes. | A new bespoke owner shape (check, corevalidate, originvalidate, pathoracle, cgen, interp, session/native admission) plus an adapter, binding manifest, and Go-side generator. See flag (f). |
| (d) 65,536-iteration cost | Not measured here (plan 29-00). | Harness design below; no interpreter step limit exists (grep of `interp.go`: only `MaxCallDepth = 128` and an output limit). | Stage A in 29-00, Stage B at phase end (Open Question 1). |

### Flag (a) in detail: ByteView argument through `OpCall`

**Check, caller path.** Straight-line callers resolve `adler32(view)` through `resolveCallBinding`, which gates the argument by exact constructor-string equality against the AST-derived callee contract and derives the call target's type from the callee's declared return type [VERIFIED: internal/compiler/check/check.go:4189-4290; quote: `argumentTypeFact.Shape.Constructor != contract.ParameterType`]. A non-copyable argument is moved by the call [VERIFIED: check.go ~4330, `!hasTypeAbility(argumentTypeFact, core.AbilityCopy)` sets `argument.initialized = false`]. That machinery would accept a `ByteView`-typed place **if** the caller's type facts contain `ByteView`. The fixture `main` is not an ordinary `checkLinear` body, though: a body containing `try` goes through `checkFallibleLinear` (check.go:3693 region), whose generic branches admit only "a single `try` immediately returned" or "try/discard calls whose result is the function's own parameter" and otherwise refuse with `check.foreign_call_shape_unsupported`. The Phase 23/24 entries are built by hand-assembled core (for example `checkLocalFileByteEntry`, check.go:3961, builds the four-operation list literally). So the fixture caller needs a new dispatch branch and a hand-built core body of `[acquire, borrow->view, OpCall adler32, release, return]`. Recommendation: construct the `OpCall` from the same callee-contract table (not a hand-forged operation) so the argument-type and return-type gates still run.

**Call admission is where it fails today.** After every function is checked, `check.Program` builds the signature table with `buildCallSignatureTable` (check.go:2010) and runs `verifyCallableRefusal` (check.go:2067). `Callable` is computed as `callable := len(PublishProblemsFor(function, calleeContracts)) == 0` [VERIFIED: originvalidate.go:2131]. `PublishProblemsFor` (originvalidate.go:1320) takes the remainder branch first:

```go
if hasRemainderOperation(function) {
    valid, zeroOperationID := originScalarFixedPointWithFailure(function)
    ...
    if !valid {
        return []Problem{{Code: "core.scalar_analysis_invalid", Detail: function.ID}}
    }
}
```
[VERIFIED: originvalidate.go:1325-1334, quoted from the Read/sed output this session]. `adler32` contains `% 65521`, so this branch runs, and `originScalarFixedPointWithFailure` begins with the place-type gate

```go
for _, place := range linear.Places {
    if types[place.TypeID] != "U64" && types[place.TypeID] != "Bool" {
        return false, ""
    }
```
[VERIFIED: originvalidate.go:314-317]. The ByteView parameter place fails it, so `adler32` is un-Callable and `check` emits `core.callee_not_callable` for every call to it. `corevalidate.peerCallable` (corevalidate.go:4029) does **not** have this gate (it consults only foreign-origin omission and borrow-derived return), so today the two derivations would silently diverge in the opposite direction; teaching originvalidate the view place makes them agree. `originHasScalarCFG` (originvalidate.go:293-300) and `originTransferScalarOperation` (originvalidate.go:417) must also learn the two ops, otherwise `RecomputeOriginPerReturn` returns nil for the callee.

**No `Mode` plumbing is required for admission.** `BuildInterface` hard-codes every parameter as owned (`Mode:  "owned"`, originvalidate.go:2099-2100) and `check` only reads `Parameters[0].Mode` to build a diagnostic detail string (check.go:1142). Admission reads the AST callee contract plus the `Callable` bit. `deriveFunctionUsesParam` (check.go:1232) treats any non-call, non-return operation that reads a parameter-derived place as a use (`default: usesParam = true`), so it is conservatively correct for `OpBytesLen`/`OpBytesAt` with no change.

**Subtle trap: `:type:0` is U64 in scalar functions.** `checkScalarFunction` appends the type facts `U64`, `Bool`, `Unit` in that order (scalar.go:40-49) and adds the parameter place with `b.addPlace(function.Parameter.Name, "U64", false)` (scalar.go:50). `interfaceParameterAbilities` reads `function.ID+":type:0"` (originvalidate.go:2279-2281). For a ByteView parameter the parameter type fact must be `:type:0`, or the published `Drops`/`Abilities` will describe a U64. Existing U64-parameter functions must keep their current ordering so no core hash moves.

### Flag (b) in detail: parser

**Probed on today's binary** (scratchpad fixtures built with `go build ./cmd/schway`): a body containing `byte_len(view)` is refused at the `(`: `syntax.expected_lbrace [116:117]`, then `syntax.expected_scalar_operand`. A ByteView parameter on a plain scalar body is refused `type.scalar_signature` (scalar path) or `type.unknown ... unknown type constructor "ByteView"` (linear path). There are zero hits for `ByteView` in any `.go`, `.schway`, or `.json` file in the repository [VERIFIED: grep of the repo, only planning documents match].

Facts the planner needs, with the exact code:

- `scalarExprAtom` returns a bare `name` for any identifier and never looks ahead (parser.go:528-541). **This is the only parser change the adler32 body needs**: a body that begins with `var`/`if`/`while` is already routed to the scalar parser regardless of parameter type.
- `scalarBodyStart` widens only for `let`-first bodies, and only for U64 parameters:

```go
func (p *parser) scalarBodyStart(parameterType, returnType ast.TypeRef) bool {
	if p.atAny(TokenVar, TokenIf, TokenWhile) {
		return true
	}
	if parameterType.Constructor == "U64" && (returnType.Constructor == "U64" || returnType.Constructor == "Unit") && p.scalarLetStart() {
```
[VERIFIED: parser.go:358-364]. `scalarLetStart` ends `return next != TokenLParen && next != TokenDot` (parser.go:355), which is what rejects `let x = byte_len(`. So D-03's `scalarLetStart`/`scalarBodyStart` edits are needed only if a library body may *begin* with `let`. If it begins with `var`, they are dead weight; if they are skipped, a `let`-first ByteView body falls to the generic `linearBody` path and produces confusing call-resolution diagnostics, so either widen them or document and test that library bodies start with `var`.
- **`TokenComma` works with no lexer change**, but not because the lexer produces it: the lexer returns `TokenUnknown` for `,` and `normalizeOwnershipTokens` rewrites it in place (parser.go ~140-152), with `filterNormalizedTokenDiagnostics` dropping the matching `syntax.unexpected_byte` problem. Call-argument lists already accept commas (`callArguments`, parser.go ~745-760); arity is refused later by check (`check.call_arity_unsupported`, check.go:4189 region).
- **`data` is a keyword.** The lexer keyword map is `"module" "export" "type" "data" "fn" "match" "var" "if" "else" "while" "foreign" "try" "discard" "because" "defect"` [VERIFIED: lexer.go:10-26], and `let`, `take`, `borrow`, `mut` are normalized to keyword tokens (parser.go ~80-100). Probed: `fn adler32(data: ByteView)` fails with `syntax.expected_parameter_name [47:51]`. **The literal signature in the ROADMAP success criterion and the `byte_at(data, i)` examples in D-01/D-04 cannot parse.** Use a non-keyword parameter name (for example `bytes` or `view`) in the library, fixtures, formatter goldens and docs; amend the criterion text with a dated note.
- **The scalar expression grammar has no parentheses and `%` binds tighter than `+`.** Precedence is `<`/`==` = 1, `+` = 2, `%` = 3 (`scalarBinaryOperator`, parser.go ~500). `a = a + x % 65521` parses as `a + (x % 65521)`, so each reduction must be its own statement. A statement may not end in an expression: the result must be a bare identifier (`scalarLinearBody`), so compute `var r = b + a` and return `r`.
- **Formatter:** `Format` is a token projection (`syntax/format.go:11`). By reading it, `var n = byte_len(view)` and `a = a + byte_at(view, i)` render unchanged (comma writes `", "` at format.go:372; `(`/`)` have no spacing logic outside `let`/`try`/write-call contexts). A `let` binding whose value is a call sets `callBinding` and breaks the line at the closing paren, so `let x = byte_at(view, i) + 1` would split after the paren. The formatter probably needs no code change but **needs goldens** [ASSUMED: derived by reading format.go; not run because the parser cannot yet parse the input].

### Flag (c) in detail: release while a loan is live

- Loan liveness is consulted for `OpMove` only. In corevalidate: `case core.OpMove: blockedUntil, hasLoan := ownerBlockedUntil[operation.SourceID]; if !v.check(!hasLoan || blockedUntil < index, "core.move_while_borrowed", ...)` [VERIFIED: corevalidate.go:2653-2656 and 3000-3003]. `OpRelease` there only checks that `ReleasesOperationID != ""` and allocator pairing (`core.release_target_unknown`, corevalidate.go:2749-2765). In check the same decision lives in `ownership.move_while_borrowed` (check.go:3116 ff., 1634-1650), again moves only.
- Releases are not user-spellable. They are minted by the checker from the acquire's declared `release:` policy (for example `phase24ReleaseOp`, check.go:3953; `checkLocalFileByteEntry` builds `Kind: core.OpRelease` literally at check.go:~4014). So a "release inside the loop" cannot be written in Schway source; the only way to exercise the refusal is a seeded fault in typed core (D-08's "one seeded fault per peer").
- What "refuses" a misplaced release today is accidental: the local-owner replays pin the exact mode sequence (`len(foreignModes) != 3 || foreignModes[0] != "acquire" || foreignModes[1] != "borrow" || foreignModes[2] != "consume"`, corevalidate.go:~1967), and pathoracle/originvalidate/cgen each match positional operations (`ops[0].Kind == core.OpCall && ops[1].Kind == core.OpForeignCall && ops[2].Kind == core.OpRelease && ops[3].Kind == core.OpReturn`, pathoracle.go:~419). A new shape written the same way is fail-closed against a reordered release for the same reason, but with shape-mismatch codes, not a loan-liveness law.
- **Recommendation:** author one named law per peer, derived independently, for "an `OpRelease` of owner place P while any view derived from P is read later" (suggested codes in the `core.`/`pathoracle.` namespaces, distinct per layer, as with `core.callee_not_callable` vs `check.*`), and cover each with a seeded fault that moves the release before the `OpCall`. This is the SC4 "release or move an owner inside the loop" boundary at the caller.
- Not a design threat: the design never lets an owner or loan exist inside the loop function (no owner/loan parameter is admissible there; see the neighbor table), so FLOW-02's back-edge boundary is untouched.

### Flag (e) in detail: no evented stores

Event behavior per construct, verified at the engines:

| Construct | Core op | interp | cgen (scalar function) | Evidence-size bound (`schema2ExecutionDocumentSize`) |
|-----------|---------|--------|------------------------|-----|
| `var x = e`, `x = e` | `OpScalarStore` | no event (interp.go:1422-1424) | plain assignment | `continue` (cgen_program.go:239) |
| `let x = e` | `OpCopy` | `value.copied` event (interp.go:1439-1442) | `schway_record_event("value.copied", ...)` per execution (cgen_program.go:2685) | event reserved |
| literal in an expression | `OpConst` | no event | `UINT64_C(..)` | `continue` |
| `+` | `OpAddChecked` | event only on overflow | overflow branch only | terminal-defect reserve |

Event capacity is *static*: `invocationEventCapacity` sums `len(function.Linear.Operations)` over the bounded unfolding (cgen_program.go:~177-195), and its own comment says "a scalar body with a back edge may record more events at runtime, in which case the execution/conformance shell exits with the named event-capacity status". Therefore a `let` inside the byte loop would exhaust capacity (exit 69 / `native.event_capacity_exhausted`) long before 65,536 iterations. The library must use `var` and plain assignment only, and the new ops must be added as event-free in `schema2ExecutionDocumentSize` (OpBytesLen: `continue`; OpBytesAt: reserve only the terminal OOB defect event, like `OpAddChecked`). Observed baseline: a `%`-containing, let-free 100,000-iteration scalar loop called from a straight-line `main` ran natively in 0.00 s and printed `11449` (probe `rem.schway` in the session scratchpad).

### Flag (f) in detail: fixture bytes through a foreign fixture owner

How the Phase 23 pattern works end to end today (all verified by reading the listed files):

1. **Source.** `foreign C { fn schway_file_byte_acquire(path: PathToken) -> FileByteOwner { unwind: forbidden  nonlocal_exit: forbidden  mode: "acquire"  abi: schway_file_byte_acquire_fn  allocator: "libc_malloc"  release: "schway_file_byte_release"  fails: AcquireError } ... }` plus `data AcquireError = | AcquireFailed` and `fn main(path: PathToken) -> U64 { let owner = try schway_file_byte_acquire(path)  let value = try schway_file_byte_use(owner)  value }` [VERIFIED: examples/phase23/file_byte.schway]. Parsing is generic (`foreignBlock`, `ForeignPolicy` key/value lines; `ast.ForeignSymbol`, ast.go).
2. **Check.** `collectForeignSymbols` (check.go:3610) keeps `mode`, `abi`, `release`, `allocator`, `fails`, `alias` policies (check.go:3640-3660). The entry is dispatched by *parameter type*: `parameterType.Constructor == "PathToken"` then `function.ReturnType.Constructor == "FileByteOwner"` (check.go:3699-3704) into `checkLocalFileByteAcquireHelper` / `checkPhase24ErrorFunction` / `checkLocalFileByteEntry`. Each hard-codes the symbol names and signature (`foreignSignature(acquire, "PathToken", "FileByteOwner")`, `schway_file_byte_acquire_fn`, `acquire.Allocator == "libc_malloc"`, check.go:3812-3814, 3986) and assembles exactly four operations `[OpForeignCall acquire, OpForeignCall use, OpRelease, OpReturn]` with exactly three types and three places (check.go:~4000-4020). The release is *materialized from the acquire's own `release:` policy*; it is not source syntax.
3. **Core.** Each foreign operation carries a per-operation `core.ForeignOperationContract{Symbol, ABIType, Mode, ParameterType, ResultType, Fails, Allocator, Release, Unwind, NonlocalExit}` (core.go, `ForeignOperationContract`). This type and the per-operation contract are generic.
4. **Peers (all name-pinned).** corevalidate `validateLocalForeignOperation` (corevalidate.go:1405) and the lifecycle replay (corevalidate.go:~1930-1975); originvalidate `localOwnerOriginProblem` and acquire/borrow/release matchers (originvalidate.go:778-850, 935, 1569); pathoracle `ValidateLocalOwnerPaths` (pathoracle.go:59-140) and `validateTransferredOwner` (pathoracle.go:296-400, 481). Example of the pinning, verbatim: `places[operation.SourceID] != "PathToken" || places[operation.TargetID] != "FileByteOwner" || ... contract.Symbol != "schway_file_byte_acquire" || contract.ABIType != "schway_file_byte_acquire_fn" ... contract.Allocator != "libc_malloc"` (pathoracle.go:118). `ability.Derive` and `corevalidate.deriveAbility` each have a hard-coded `case "FileByteOwner":` (ability.go:157, corevalidate.go ~4687).
5. **cgen.** `localFileByteTransferHelper`, `isLocalFileByteFunction`, `validateLocalFileByteFunction` (cgen_program.go:1563), `emitProgramLocalFileByteFunction` (1773), `emitProgramTransferredOwnerHelper/Caller`. The emitted C is literal text naming `schway_file_byte_acquire_result`, `schway_file_byte_owner`, and `if (%s.data == NULL || %s.length != UINT64_C(1)) { schway_file_byte_release(%s); exit(65); }` (cgen_program.go:~1663, 1807). The generated TU includes the adapter as `#include "local/adapter.h"` only when `localFileByteProgram` (cgen_program.go:890).
6. **Session/native admission.** `BuildApplication` routes `localOwnerEntry := entry.Parameter.Type == "PathToken" && hasLocalForeignOperations(program)` (session.go:1401), runs `pathoracle.ValidateLocalOwnerPaths` and `validateLocalOperationBindings` (symbol-generic: expected set from `operation.Foreign.Symbol`/`ABIType` must equal the manifest's symbols exactly, session.go:1446-1470). A `U64` entry with foreign operations bypasses that block entirely.
7. **Native adapter and manifest.** `examples/phase23/file_byte.bindings.json` is schema `schway.local-c/1` (`sources`, `headers`, `include_dirs`, `symbols[{name, header, function_type}]`, `runtime_dependencies`); `native.ResolveBindings` hashes and stages it and `compileBindings` compiles each source separately then links (native/bindings.go:21-60, 224-300, 519). This layer is **generic** and reusable as-is for an `examples/phase29/` adapter. Compile flags are `-std=c17 -Wall -Wextra -Werror -pedantic -O0` for applications (`applicationFlags`, native_app.go:115).
8. **Interpreter.** There are no bytes. `interp.RunWithForeignOutcomes` runs a *model*: "The opaque input is treated only as a modeled PathToken; this function never opens it" (interp.go:196-199); it demands a `PathToken`-to-`U64` entry (interp.go:213) and an acquire outcome that is "one canonical byte" (`strconv.ParseUint(outcome.Value, 10, 8)`, interp.go:311), and rejects any terminal outcome that leaves live resources (interp.go:253). `value` has no byte payload (interp.go:824-833). `RunInterpreterApplication` accepts only `U64` entries (session.go:1153-1180).

**Consequence.** For `FixtureBytes` the following are all new: the `check` dispatch branch and hand-built core for `[acquire(case id), borrow->ByteView, OpCall adler32, release, return]`; the ability entries for `FixtureBytes` and `ByteView` in two independent tables; the corevalidate lifecycle replay variant; the originvalidate and pathoracle matchers and their seeded-fault coverage; the cgen validator/emitter and the C `ByteView` argument ABI at the call; the interpreter `[]byte` binding with an independent Go generator (D-07) and a defect-with-live-owner semantics (the OOB control runs while the fixture owner is live); the session admission branch; the adapter C, header, manifest, and a table test that C and Go generators agree byte-for-byte. Precedent size: Phase 23 had seven plans (`.planning/milestones/M004-phases/23-live-local-allocation-and-discharge/23-01..07-PLAN.md`). Estimate: comparable.

**Why D-05 is still the right call (verified constraint):** the application route cannot carry bytes in its argument. `MaxApplicationArgumentBytes = 4096` (native_app.go:25) and `app verify` cases are canonical U64 only (session_app_verify.go:181), so a hex ByteView entry parameter cannot reach the 65,536-byte cap on any host; adapter-generated bytes can. (On Linux a single argv string is also capped far below 131,072 hex characters [ASSUMED: kernel `MAX_ARG_STRLEN`, not probed].)

**Cheaper shapes the planner/user may want to price at the 29-00 gate (not recommended over D-05 unless the gate fails):**

1. *Case id from the entry parameter instead of a source literal.* `fn main(case: U64) -> U64` with `try fixture_acquire(case)`: the acquire's `SourceID` is the entry parameter exactly as in Phase 23, no `OpConst` appears in the exact-shape body, and **one built artifact serves every vector** (`app run ./art -- 4096` ...). This also makes the D-12 compile time measured once. It deviates from D-06's "literal case id baked into the source" (the entry is still U64, not a ByteView, and `schway run --engine` is untouched); it is within Claude's Discretion on "case-id numbering" only if the user accepts the deviation. Flag for confirmation.
2. *Reuse `PathToken` as the acquire parameter* with the case id as the path string: reuses entry dispatch (`PathToken` entry), the `app run -- <string>` route and `RunWithForeignOutcomes`' entry check, and matches the eventual Phase 31 `main(path: PathToken)` shape. Conflicts with D-05 (acquire takes U64) and with D-09's separation, so it needs an explicit user decision.

### Runnable-gain mismatch the planner must resolve

The ROADMAP "Runnable gain" says `schway run --engine=interpreter` and `--engine=native` print the checksum. **No U64-entry program can run through `schway run --engine=*` today**: probed `run --engine=interpreter` and `--engine=native` on `examples/sum_to_n.schway` and on a scratch U64 program, both exit 3 (`tool.run_failed`, `native.tool_failure`), because `interpreterInputs` only synthesizes inputs for `Byte`/`Buffer`/data-type entries (session.go:1190-1223, `default: return nil, false`). D-06 forbids changing `schway run`. The honest public surfaces under D-06 are native: `schway build FIXTURE --manifest M --output ART` then `schway app run ART -- N` (Phase 23 precedent); interpreter: the `session` API. Record this in the plan and amend the ROADMAP sentence with a dated note, or ask the user to relax D-06.

## Architectural Responsibility Map

This is a compiler, so tiers are pipeline stages rather than web tiers.

| Capability | Primary tier | Secondary tier | Rationale |
|------------|-------------|----------------|-----------|
| `byte_len`/`byte_at` surface syntax, reserved names, arity | syntax (`parser.go`, `format.go`) | check (reserved-name and arity refusals) | Parsing owns the atom; the one-parameter-function rule and call-arity refusal already live in check. |
| `ByteView` type, abilities | check + `ability` | corevalidate `deriveAbility` | Two independent ability derivations exist; both need the entry. |
| Byte-op lowering and scalar fixed point | check (`scalar.go`) | corevalidate, originvalidate (each re-derive) | Four deliberately copied scalar classifiers (D-27-07 copy-over-dep). |
| Call admission of a ByteView-parameter callee | originvalidate (`BuildInterface`/`PublishProblemsFor`) | check (`verifyCallableRefusal`), corevalidate (`peerCallable`) | `Callable` is produced by originvalidate and consumed by check; corevalidate re-derives. |
| Dynamic bounds check + OOB defect | cgen (C text, both shells) | interp (event + defect outcome), executionpeer (evidence attribution) | Never elided: an ordinary `if` with `fputs`/`exit(65)`; evidence peers must classify the event. |
| Fixture byte delivery | adapter C + Go interpreter binding | check/peers/cgen/session (owner shape) | Bytes are generated independently per engine (D-07). |
| Release-before-view-use refusal | corevalidate + pathoracle + check (post-assembly) | originvalidate | New independent laws; releases are compiler-inserted. |
| Cap evidence | spike 008 harness (test + README) | CI (`go test`, `go test -race`) | Wall time never enters goldens. |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go standard library | go1.24.0 (darwin/arm64 local) [VERIFIED: `go version`] | Compiler host and test host | Project constraint: Go 1.24 stdlib Stage 0, no dependencies (AGENTS.md). |
| Clang (C17) | Apple clang 21.0.0 local [VERIFIED: `clang --version`]; hosted Ubuntu/macOS via CI | Native emission target | Flags `-std=c17 -Wall -Wextra -Werror -pedantic` (native_app.go:115; native.go Run). |
| Python `zlib` (test-only oracle) | Python 3.14.4 [VERIFIED: run this session with `python3 -I`] | Independent pinned literals | Decision 1; produces the six literals plus cap values. |
| Go `hash/adler32` (test-only oracle) | go1.24.0 [VERIFIED: ran a scratch program this session] | Second independent oracle | Decision 1. |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `internal/compiler/exitcode` | in-tree (Phase 28) | single exit table (`AppInputRefused = 65`) | New C `exit(65)` text must stay within the scanned literal set `{0, 65, 69}`. |
| `internal/compiler/testsupport` | in-tree | `RunCLITimeout = 2 * time.Minute`, `BuildCLITimeout = 5 * time.Minute` [VERIFIED: testsupport.go:36-37] | CLI-spawning tests and the interpreter budget in D-13. |
| `native.Runner` | in-tree | `Run(ctx, cSource, optimization, inputs)`, `LTO bool`, returns `CompileTime`/`RunTime` separately | Conformance-shell lanes at `-O0`/`-O3`/`-O3`+`-flto`. |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Foreign fixture owner (D-05, locked) | Entry-parameter case id or `PathToken` reuse | Much cheaper; deviates from D-06/D-05 text; needs user confirmation (see flag (f)). |
| Hex `ByteView` entry parameter (the earlier ARCHITECTURE.md proposal) | n/a | Cannot carry 65,536 bytes: argument limit is 4,096 bytes. |

**Installation:** none. No packages are added.

**Version verification:** versions above were read from the local toolchain this session; no registry lookup applies (stdlib only).

## Package Legitimacy Audit

No external packages are installed or recommended. All code is in-tree Go standard library and C17 against libc. Test-only oracles (Python `zlib`, Go `hash/adler32`) are standard-library modules, not installed packages.

**Packages removed due to [SLOP] verdict:** none.
**Packages flagged as suspicious [SUS]:** none.

## Architecture Patterns

### System Architecture Diagram

```
.schway source (fixture main + adler32 library)
   |
   v
syntax.Parse ──► ast.Program  (new: Expr kinds for byte_len / byte_at; ByteView is just an identifier TypeRef)
   |
   v
check.Program
   |-- scalar path: checkScalarFunction (U64 gate -> admit exactly one ByteView param place)
   |-- fixture-main path: NEW closed shape  acquire -> borrow(view) -> OpCall adler32 -> release -> return
   |-- post-assembly: buildCallSignatureTable -> verifyCallableRefusal (needs originvalidate Callable == true)
   v
core.Program  (new kinds: bytes_len, bytes_at; new TypeRef shape "ByteView"; no new LinearOperation field)
   |
   +--► corevalidate.Validate      (own scalar fixed point, own per-op law, own release-vs-view law)
   +--► originvalidate.ValidatePublished (own scalar fixed point; BuildInterface/Callable)
   +--► pathoracle.ValidateLocalOwnerPaths (fixture owner replay; reads do not extend loans)
   |
   +--► interp.Run*   ──► bytes []byte in value; OOB -> function.defected, Output "byte index out of bounds", Outcome defect
   |                     bytes for fixture owner from an independent Go generator (case id -> []byte)
   |
   +--► cgen.EmitApplication / EmitNative
           SCHWAY_BYTE_VIEW {const unsigned char *data; uint64_t length;}
           OpBytesAt:  if (index >= view.length) { app shell: fputs("schway: byte index out of bounds\n", stderr); exit(65);
                                                    conformance shell: defect terminal event }   then (uint64_t)view.data[index]
           adapter C generates bytes per case id (independent of the Go generator)
   |
   v
native.BuildApplication (-O0 only today)  /  Runner.Run (-O0/-O3/-O3+LTO, conformance shell)
   |
   v
executionpeer.Validate (must classify OpBytesAt defect)  +  session.ProjectExecutionSchema2 (defect reason map)
```

### Recommended Project Structure
```
examples/phase29/
├── adler32 library + fixtures (one template, generated or hand-written; names not `data`)
├── fixture_bytes_adapter.c / .h        # independent C generator, keyed by case id
├── fixture_bytes.bindings.json         # schway.local-c/1 manifest
└── README.md
internal/compiler/<pkg>/*_phase29_test.go     # new tests per package (naming follows phase26/27/28 files)
.planning/spikes/008-byteview-cap/README.md   # NO .schway files here (see Pitfall 6)
```

### Pattern 1: A new scalar operation, written once per consumer
**What:** Each consumer re-derives the same two facts (`bytes_len`: source ByteView place -> U64 target; `bytes_at`: source ByteView place + `RightID` U64 index -> U64 target) from core only. **When to use:** every consumer in "Everything that must learn ..." below. **Example (the exact shape of the model to copy, from the existing remainder op):**
```go
// Source: internal/compiler/check/scalar.go:217-225 (existing OpRemainderU64 arm; copy this shape)
case core.OpRemainderU64:
    right := state[op.RightID]
    if !source.initialized || !right.initialized {
        return work, "check.scalar_uninitialized", ""
    }
```

### Anti-Patterns to Avoid
- **A shared helper between peers.** The classifiers are copied on purpose (cgen_program.go:2508-2514 comment: "the sets are deliberately copied, not shared (no shared package, D-27-07 copy-over-dep), so drift stays visible").
- **`let` inside the byte loop** (an evented copy per iteration).
- **A new `LinearOperation` field.** Reuse `SourceID`/`RightID`/`TargetID`; new omitempty fields would be safe for hashes but are unnecessary, and a non-omitempty field would move every accepted-corpus core hash.
- **Eliding the check anywhere**, including "loop guard implies in range" in cgen.
- **Editing `examples/phase23` or the Phase 23-25 matchers.**

## Everything That Must Learn `OpBytesLen` / `OpBytesAt`

The orchestrator listed six consumers and four classifiers. The tree has **more sites**; the extra ones are the ones likely to be forgotten. Line numbers are from this session's reads/greps.

### Registry and exhaustive-dispatch (core)
| Site | Change |
|------|--------|
| `core/core.go:644-717` (constants), `core.go:858` `AllOperationKinds()` | Add the two kinds to the const block and the literal slice. Current quote: `return []OperationKind{OpCopy, OpMove, ... OpWriteU64, OpWriteNewline, OpWriteText}` (22 entries). |
| `core/core_test.go:799-800` | `const declaredCount = 22` must become 24. |
| `core/core_test.go` `exhaustiveDispatchFixtures` (≈line 889) and `runExhaustiveDispatchControl` | The control requires every kind to be produced by a fixture through `session.Check`, `corevalidate`, `interp`, `cgen`, `pathoracle`, `originvalidate` ("operation kind %q is never exercised by any corpus fixture"). Add a ByteView fixture (a scalar-only file with `adler32` and a trivial `main`, no foreign block, avoids needing the owner shape). `linearProbeInput` returns `ok == false` for non-Byte/Buffer so interp is skipped for it; pathoracle `RecomputeEndpoints` is skipped for functions containing `OpBranch`. |
| `core.TerminatorKinds()` | No change (neither op terminates a block); `TestPhase26CoreInventory`-style inventory test should assert that. |

### Syntax
| Site | Change |
|------|--------|
| `syntax/parser.go:528` `scalarExprAtom` | Branches for identifier text `byte_len` (+ `(` name `)`) and `byte_at` (+ `(` name `,` scalarExpr `)`); new `ast.Expr` kinds (discretion). Decide the diagnostic codes for wrong arity (follow `syntax.write_u64_arity`, parser.go:468). |
| `syntax/parser.go:342-366` | `scalarLetStart` / `scalarBodyStart` only if `let`-first bodies must work (see flag (b)). |
| `syntax/format.go` | Probably none; add formatter goldens (`byte_at(bytes, i)`, one space after the comma). |
| `ast/ast.go` (`Expr`, Kind comment at line ~120: `// name, number, text, add, less`) | Doc comment update; reuse `Name`/`Left`/`Right`, no new fields required. |

### Six consumers plus the extra sites
| Consumer / site | Function and approximate line | What to add |
|-----------------|-------------------------------|--------------|
| **check** | `check/scalar.go:34` `checkScalarFunction` (U64 gate at 35-37; type facts at 40-49; parameter place at 50) | Admit exactly one `ByteView` parameter (return stays U64/Unit); type order so `:type:0` = ByteView for that shape only. |
| | `scalar.go:125` `checkScalarFixedPointDetailed` (cases at ~190-262; `default` at 261 -> `check.cfg_back_edge`) | Add `bytes_len`/`bytes_at` arms; a `view bool` fact (the fact struct is compared with `==`, so no slices). |
| | `scalar.go:280` `scalarAuthorityCategory` | Add `"ByteView"` beside the existing `"Loan", "Borrowed", "View"` case. |
| | `scalar.go:370` `lower`, `:569` `expression` | Lower the two Expr kinds; refuse a `ByteView` name used anywhere else (`type.scalar_operand`, `type.scalar_binding`, `type.scalar_result` already fire for non-U64/Bool). |
| | `check.go` reserved-name and arity refusals (D-04) | New diagnostics for user `fn byte_len`/`fn byte_at`; wrong arity. |
| | `check.go:5869` `executableShape` | Leave ByteView **out** (a straight-line ByteView-parameter function stays `check.unexecutable_shape`, which is what keeps the view confined to scalar library functions). |
| | `ability/ability.go:~130-165` | New `case "ByteView"` (see Open Question 4 for the ability set). |
| | fixture owner path | New dispatch + hand-built core (flag (f)). |
| **corevalidate** | `corevalidate.go:4298-4307` `hasScalarCFG` | Add the two kinds. |
| | `:4344-4352` `validateScalarFixedPoint` | Place gate `constructor == "U64" \|\| constructor == "Bool"` must admit the single parameter ByteView place; add arms at ~4429-4440. |
| | `:4468` straight-line facts | Add arms (default arm marks only U64/Bool targets). |
| | `:1237` structural per-op law, `:1386` `core.scalar_field_kind_exclusive` | `RightID` is currently legal only on add/less/equal/remainder: extend to `OpBytesAt`; add per-op structural law (types ByteView/U64/U64, no stray fields, target is not the parameter). |
| | `:2597`, `:2928` replay switches | Add arms (these fail with `core.unknown_operation` otherwise). |
| | `:4687` `deriveAbility` | `ByteView` (and the owner type) entries. |
| | `peerCallable` (4029) | No scalar gate today; confirm it still returns true for a ByteView callee. |
| **originvalidate** | `originvalidate.go:293-300` `originHasScalarCFG` | Add the two kinds. |
| | `:307-319` `originScalarFixedPointWithFailure` | Place gate admits the single parameter ByteView place (this is the flag (a) blocker). |
| | `:417-470` `originTransferScalarOperation` | Add arms. |
| | `:1403` `phase27ArithmeticProblem`, `:1388` `hasRemainderOperation` | Add a structural problem check for the two ops (this peer must refuse a forged operator shape on its own). |
| | `:2047` `BuildInterface` (line 2099 `Mode: "owned"`), `:2279` `interfaceParameterAbilities` | `:type:0` convention; keep Mode "owned" unless a decision is recorded. |
| **pathoracle** | `pathoracle.go` | Scalar loops are excluded by design (cyclic); confirm `RecomputeEndpoints` ignores the new kinds on acyclic bodies and add its own seeded-fault coverage. The substantive pathoracle work is the fixture-owner replay (`ValidateLocalOwnerPaths`, 59-140; `validateTransferredOwner`, 296-400). |
| **interp** | `interp.go:824-833` `value` | Add a bytes payload (value is copied across frames; a slice header is fine because the view is read-only). |
| | `interp.go:1377-1420` op switch | `OpBytesLen`, `OpBytesAt`; OOB returns `Execution{... Outcome: Outcome{Kind: execution.OutcomeDefect, Value: ""} ...}` with an event whose `Output` is `byte index out of bounds` (copy the `OpRemainderU64` construction, interp.go:1409-1418). |
| | `interp.go:199` `RunWithForeignOutcomes`, `:1153` session `RunInterpreterApplication` | Owner binding with real bytes (flag (f)). |
| **cgen** | `cgen_program.go:228-239` `schema2ExecutionDocumentSize` | `OpBytesLen` -> `continue`; `OpBytesAt` -> reserve a terminal `function.defected` event. A missing arm is a hard error in the default path. |
| | `cgen_program.go:2515-2530` `hasScalarOperations` (+ `isStraightLineScalarBody`) | Add the two kinds to the operation list; keep the U64-only places/params rule of `isStraightLineScalarBody` as is. |
| | `cgen_program.go:2650-2760` `emitProgramScalarFunction` | `case core.OpBytesLen`, `case core.OpBytesAt` in BOTH shells; non-U64 parameter type; local `schway_scalar_N` declaration loop currently types every non-parameter place `uint64_t`/`int`. |
| | `cgen.go:1515` `linearInput` | `ByteView` parameter type name (`SCHWAY_BYTE_VIEW`); add the typedef where `SCHWAY_BUFFER` is emitted (cgen_program.go:~905) gated on use. |
| | `cgen.go:~190-215` `linearFixedNames` | Reserve the new typedef/member names; `TestGeneratedIdentifierNamespacesStayConfined` (cgen_names_test.go:147) enforces the `SCHWAY_*`/`schway_value_*` namespaces. |
| | `cgen_program.go:1000+` emission of the `OpCall` with a view argument | The call passes the by-value `SCHWAY_BYTE_VIEW` (D-04's "`const uint8_t*` + length" ABI is satisfied by the struct's two members; an alternative of two C parameters would break the one-parameter-per-function C signature `static T f(P, unsigned int)` at cgen_program.go:~910). |
| **executionpeer** (extra) | `executionpeer.go:367-368`, `:497` | Add a checked-bytes-at defect arm beside `checkedOverflow`/`checkedRemainderZero` (fixed reason, source-place attribution, final event, empty defect outcome) and add `OpBytesAt` to the `function.defected` classifier at line 497. Without this the peer rejects the OOB execution as an unclassified event. |
| **session** (extra) | `session.go:~527-533` `ProjectExecutionSchema2` `defectReasons` map | Add `case core.OpBytesAt: defectReasons[...] = "byte index out of bounds"`. |
| **session** (extra) | `session.go` admission (`BuildApplication` ~1385-1430) | Fixture-owner entry branch and manifest validation. |

### Tests and registers that move
| Register/test | File | What moves |
|---------------|------|-----------|
| `TestAllOperationKindsRegistered`, `TestAllOperationKindsHandledAtEverySite`, `TestPhase26CoreInventory` | `core/core_test.go` | count 22 -> 24, fixture list |
| `TestPhase28EntryExitClassManifest` | `session/session_phase28_manifest_test.go:444` | Every new `.schway` under `examples/` and `testdata/` must be added to `phase28NewlyAddedEntries` (and `phase28MinCorpus` floors). |
| `TestPhase272AcceptedCorpusByteIdentity` | `session/session_phase272_identity_test.go:119` | New accepted sources go in `phase272NewlyAcceptedPaths`; existing core/C hashes must not move. |
| `TestPhase273ConformanceCorpusByteIdentity` | `session/session_phase273_identity_test.go:137` | New accepted sources go in `phase273NewlyAcceptedPaths`. |
| `TestLanguageMaturityCountsAreCurrent` | `session/self_describing_docs_test.go:350` | `.planning/LANGUAGE-MATURITY.md`'s `**N .schway programs, N lines total**` sentence counts every `.schway` file in the whole tree (it skips only `.git`); also the `len(Functions) != 1` guard table, so do not add such guards. |
| `TestPhase16PublicEmitterConsumerInventory` | `session/session_phase16_emitter_inventory_test.go:100` | Keyed by emitter call site `file:Func:line`; any `session.go` edit that shifts an `EmitNative`/`Emit` call needs a registry update. |
| `TestEmittedCExitLiteralsMatchTable` | `cgen/cgen_exit_literals_test.go` | Passes unchanged if OOB uses `exit(65)`: the scanned set must stay exactly `{Success, AppInputRefused, EventCapacity}`. |
| `TestGeneratedIdentifierNamespacesStayConfined` | `cgen/cgen_names_test.go:147` | New C names stay inside `SCHWAY_*`. |
| `TestPhase20ValidationLifecycle`, groundedness tests | `session/verification_groundedness_test.go` | A `29-VALIDATION.md` left `status: draft` fails; `go test` command rows in enforced docs must resolve to real tests. |
| `TestExitCodeDocsTable` | `session/session_phase28_exitdocs_test.go` | README "Exit codes" table; unaffected unless a new exit code is introduced (none is). |

### Repair tooling
`cmd/schway-repair` consumes `diagnostic.Repair` records generically and its tests assert a *closed decline-reason vocabulary*, not a per-diagnostic-code registry (`repair_diagnosis_test.go`, `antitheater_test.go`). New refusal codes need no registry edit; machine-applicable repairs (for example "rename user function colliding with a reserved builtin") are optional, per Claude's Discretion.

## ByteView Representation: today vs. to add

| Aspect | Today | To add |
|--------|-------|--------|
| Syntax | `ByteView` is just an identifier `TypeRef`; zero repository hits otherwise. | None (type name only). |
| core type | None. | `core.TypeRef{Constructor: "ByteView"}`; a `TypeFact` with abilities. |
| Abilities | `ability.Derive` default: `unknown type constructor "ByteView"` (probed: `type.unknown`). | Entries in `ability.go` and `corevalidate.deriveAbility`. Recommendation [ASSUMED]: grant `drop` only; deny `copy`, `share`, `send`, `escape` with witness path `["ByteView"]`. `ARCHITECTURE.md` is inconsistent (line 89: "droppable view"; line 148: "grants no abilities"): pin one in 29-00. |
| Where it may appear | n/a | Only as the single parameter of a scalar CFG function (check `checkScalarFunction`), and as the borrow result in the fixture-main shape. `executableShape` stays false so straight-line ByteView functions remain `check.unexecutable_shape`. |
| Place facts | n/a | A parameter place never assigned, copied, moved, borrowed, stored, or returned; the only legal readers are the two ops. |
| cgen ABI | n/a | `typedef struct SCHWAY_BYTE_VIEW { const unsigned char *data; uint64_t length; } SCHWAY_BYTE_VIEW;` passed by value (satisfies "`const uint8_t*` + length"; `unsigned char` is the C type for unsigned widening; the signed-char mutant changes this type). Index compared as `uint64_t` before the load; `data` may be NULL only when `length == 0`, where the check fires first. No `restrict`/`noalias`: Phase 25 pointer shapes carry none and `ScanForUnjustifiedAttributes` allows only a justified `restrict`. |
| interp value | `value` has `u64`, `isU64`, `boolean`, `isBool`, `payload string`, owner ids (interp.go:824-833). | `bytes []byte` (+ tag). Comparable-struct uses of `value` must be checked. |

## OOB failure path today (to reuse for the byte bounds check)

| Layer | Existing checked failure (U64 overflow / remainder by zero) | What OOB does |
|-------|-----------------------------------------------------------------|---------------|
| interp | Appends an event and returns the defect outcome: `Event{... Kind: "function.defected", ... Output: "U64 addition overflow"}` then `Execution{... Outcome: Outcome{Kind: execution.OutcomeDefect, Value: ""} ...}` (interp.go:1377-1390; remainder 1406-1418). There is **no process exit and no stderr** in the interpreter. | Same construction with `Output: "byte index out of bounds"`. SC3's "exits 65 with stderr exactly ... in the interpreter" is therefore a *projection*; the plan must define the interpreter assertion honestly: `Outcome.Kind == defect`, final event `function.defected` with that reason, empty application output. Precedent: `TestPhase26OverflowProcessOutcome` asserts exactly this for the interpreter and the exit-65 stderr only for native (session_phase26_overflow_test.go:19-80). |
| cgen, application shell | `fputs("schway: U64 addition overflow\n", stderr); exit(65);` (cgen_program.go:2690; remainder 2706). | `fputs("schway: byte index out of bounds\n", stderr); exit(65);` emitted as an ordinary `if (index >= view.length)` *before* the load. A real branch with a side effect cannot be removed by `-O3` or LTO under C17 semantics [ASSUMED: language rule; to be evidenced by the lanes below]. Never `assert` (vanishes under `NDEBUG`, Decision 4). |
| cgen, conformance shell | `emitProgramDefectTerminal(...)` records the `function.defected` event then `exit(0)` (cgen_program.go:2687-2696, 2315). | Same arm for the new op; evidence validated by `executionpeer`. |
| Exit table | `AppInputRefused = 65` is the named constant for the generated-C refusal status; `exit(65)` already appears in generated C and the scan pins the distinct set `{0, 65, 69}` (cgen_exit_literals_test.go). | No new code. The returned-status meaning map in that test says `exitcode.AppInputRefused: "application refusal"`; consider a dated clarification. |

**Missing lane (important).** `native.BuildApplication` compiles with the package variable `applicationFlags = []string{"-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic", "-O0"}` only (native_app.go:115, used at 206, 221, 239). The existing `-O0`/`-O3`/`-O3`+`-flto` lanes (`Runner.Run(ctx, cSource, "-O3", inputs)`, `Runner.LTO`) compile the **conformance** shell and decode a JSON execution document from stdout, so they cannot observe `stderr == "schway: byte index out of bounds\n"` and `exit 65`. SC3 therefore requires a new test-only application-shell compile at `-O3` and `-O3 -flto` (compile `cgen.EmitApplication` output, link the staged adapter object, run, assert exit/stdout/stderr). Recommend a follow-the-repo-pattern test seam (an unexported override that is nil in production, like `scalarAnalysisBudgetForTest`) rather than changing production flags or receipts. Application runs also need ASan/LSan care: the OOB exit leaves the fixture owner live by design (see Pitfall 8).

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Expected checksum values | Deriving answers from the Schway implementation | Pinned literals from Python `zlib.adler32` and Go `hash/adler32`, recorded with command and version | Common-mode error is the top pitfall (Pitfall 2 in research); both oracles agreed this session. |
| Local C build authority | A new build/linker path | `schway.local-c/1` manifest + `native.ResolveBindings`/`compileBindings` | Already symbol-generic; hashes and stages inputs; `validateLocalOperationBindings` checks coverage. |
| Per-case timing statistics | Ad hoc timing loops | `native.Runner.Run` result `CompileTime`/`RunTime` and `internal/compiler/measure` statistics | Compile and run are already timed separately. |
| Exit statuses | Numeric literals in new code | `internal/compiler/exitcode` constants / the scanned literal set | An AST scan fails on duplicated literals (exitcode/scan_test.go). |
| Execution-evidence classification | New evidence format | `executionpeer` + `ProjectExecutionSchema2` defect reasons | Keeps the interpreter and native evidence peer-validated. |

**Key insight:** every new analysis here must be re-derived independently per consumer, so "don't hand-roll" applies to *oracles, build plumbing, and the exit table*, not to the ops.

## Common Pitfalls

### Pitfall 1: The success-criterion signature does not parse
**What goes wrong:** `fn adler32(data: ByteView)` is a syntax error (`data` is a keyword). **Why:** lexer keyword map (lexer.go:10-26). **Avoid:** use `bytes`/`view`; amend criterion text and D-01/D-04 examples with a dated note. **Warning sign:** `syntax.expected_parameter_name` at the parameter.

### Pitfall 2: No parentheses, `%` binds tighter than `+`
**What goes wrong:** `a = (a + x) % 65521` does not parse; `a = a + x % 65521` silently means a different function (`a + (x % 65521)`), which equals Adler-32 only by accident for tiny inputs. **Avoid:** two statements per reduction. Wrong-modulus and swapped-half mutants must fail the 5,552 x `0xFF` vector, not only the small ones.

### Pitfall 3: `let` in the loop is an evented copy
**What goes wrong:** event capacity exhausted (exit 69). **Avoid:** `var` + assignment only; structural test asserting zero `OpCopy` in the loop blocks; a `let`-in-loop negative control that is seen to fail.

### Pitfall 4: `adler32` is un-Callable until originvalidate learns the view place
**What goes wrong:** `core.callee_not_callable` at the first call. **Avoid:** make this the first assertion of plan 29-00 (check -> signature table -> Callable == true).

### Pitfall 5: `:type:0` convention
**What goes wrong:** published `Drops`/abilities describe U64 instead of ByteView. **Avoid:** parameter type fact is `:type:0` for ByteView functions; keep U64 functions' ordering unchanged (hashes).

### Pitfall 6: Corpus walkers count every `.schway` in the tree
**What goes wrong:** a committed `.schway` under `.planning/spikes/008-...` or any fixture directory fails `TestLanguageMaturityCountsAreCurrent`, `TestPhase28EntryExitClassManifest` (examples/ and testdata/), `TestPhase272...`, `TestPhase273...`. **Avoid:** generate spike sources at run time in a temp dir; register every committed fixture in the three `...NewlyAdded/Accepted` maps; update the LANGUAGE-MATURITY sentence.

### Pitfall 7: Forgotten consumers
**What goes wrong:** `executionpeer` rejects the OOB event; `ProjectExecutionSchema2` emits an empty reason; `schema2ExecutionDocumentSize` errors on the new kind; corevalidate's kind-exclusive field law rejects `RightID` on `OpBytesAt`. **Avoid:** treat "Everything that must learn" as the inventory and add a test per site (seeded fault for each).

### Pitfall 8: Defect while the fixture owner is live
**What goes wrong:** the OOB control necessarily runs with the owner acquired. The interpreter wrapper rejects terminal outcomes with live resources (interp.go:253), native `exit(65)` leaks the allocation (LeakSanitizer reports under ASan). **Avoid:** decide the semantics once (defect is abort-like; live owner is recorded, not an error) and document it; use non-leaking adapter storage or `detect_leaks=0` for the OOB sanitizer lane. (ARCHITECTURE.md scheduled "defect-with-live-owner documentation" for Phase 31; it arrives here.)

### Pitfall 9: Interpreter "exit 65"
**What goes wrong:** asserting a process exit code for the interpreter. **Avoid:** assert the defect outcome + event reason + empty output (see OOB table); native asserts exit/stdout/stderr.

### Pitfall 10: `schway run --engine=*` cannot run any of this
See "Runnable-gain mismatch". Tests must use `app run` and the session API.

### Pitfall 11: Signedness and C warnings under `-Werror -pedantic`
Load `unsigned char`, widen with `(uint64_t)`, compare `uint64_t` to `uint64_t`; the adapter and the emitted C compile at `-Wall -Wextra -Werror -pedantic` at every optimization level. The signed-char mutant should be a mutation of the emitted C text (the `TestPhase26WrongResultMutationKilled` pattern, session_phase26_overflow_test.go:91) and must turn the `0x80`/`0xFF` fixture red.

### Pitfall 12: Validation docs and the meta-tests
`29-VALIDATION.md` must not stay `status: draft` (TestPhase20ValidationLifecycle); `go test ... -run` commands in enforced planning documents must name tests that exist (groundedness); this RESEARCH.md is exempt only because of its `verification_role: proposal` frontmatter and must not put a verdict token beside a command span.

## Code Examples

### Pinned vectors (re-verified this session)
```text
# python3 -I (3.14.4): zlib.adler32(...)        # go1.24.0: hash/adler32.Checksum(...)  -> identical
empty -> 1                    0x00 -> 65537             a -> 6422626            Wikipedia -> 300286872
4096 x 0x00 -> 268435457      5552 x 0xFF -> 4052720524
65536 x 0xFF -> 2006388466    65537 x 0xFF -> 2273841137 (python only; cap+1, Phase 31)
D-12 scaling workloads (python): 4096 x 0xFF -> 2170679522    16384 x 0xFF -> 2967061426
```
[VERIFIED: `python3 -I -c 'import zlib...'` and a scratch Go program, run 2026-10-11 in the session scratchpad]. The 16-doubling construction (`B*65536`) was cross-checked in Python against the same values.

### Library shape (grammar-checked by reading the parser; will not parse until the atoms land)
```text
fn adler32(bytes: ByteView) -> U64 {
  var a = 1
  var b = 0
  var i = 0
  var n = byte_len(bytes)
  while i < n {
    a = a + byte_at(bytes, i)
    a = a % 65521
    b = b + a
    b = b % 65521
    i = i + 1
  }
  var r = b
  r = r + r      # repeat 16 times: b * 65536 by self-addition, no multiply or shift operator
  ...
  r = r + a
  r
}
```
(Statement shapes match `scalarStatements`/`scalarBinaryExpr`, parser.go:392-455, 500-524; the final line must be a bare identifier result. `r = r + r` sixteen times then `r = r + a` is the intended Decision-1 construction.)

### Emitted C for the check (design, application shell)
```c
case core.OpBytesAt:
  if (schway_scalar_idx >= view.length) {
    fputs("schway: byte index out of bounds\n", stderr);
    exit(65);
  }
  schway_scalar_dst = (uint64_t)view.data[schway_scalar_idx];
```
(Same structure as the existing remainder arm, cgen_program.go:2702-2710; the check precedes the load; `exit(65)` is the literal the exit-literal scan already allows.)

### Neighbor refusals: today's codes, probed on the built binary
| SC4 neighbor | Source spelling | Code today (probe) | After ByteView lands |
|--------------|-----------------|--------------------|-----------------------|
| borrow inside the loop | `let b = borrow n` in a `while` body | `syntax.expected_scalar_operand` (+ cascade) | unchanged |
| move/take an owner inside the loop | `let t = take n` | `syntax.expected_scalar_operand` (+ cascade) | unchanged |
| owner parameter on a loop function | `fn main(b: Buffer) -> U64 { ... while ... }` | `type.scalar_signature` | unchanged (ByteView admitted only as parameter type) |
| returning a non-scalar from a loop function | `-> Buffer` | `type.scalar_signature` | unchanged; `-> ByteView` same code |
| copy the view | `let c = view` | n/a | `type.scalar_binding` ("let accepts only U64 or Bool scalars", scalar.go:424) |
| store the view | `var s = view` / `s = view` | n/a | `type.scalar_binding` (scalar.go:453) / `type.scalar_assignment` |
| return the view | result `view` | n/a | `type.scalar_result` (scalar.go:79) |
| release an owner in the loop | not spellable | n/a | seeded fault only (flag (c)) |
| loan parameter | no parameter-mode syntax | n/a | forged-core only: `check.cfg_back_edge` category `loan` (existing `TestPhase26BackEdgeCauseCategories`) |

The cascade diagnostics (the parser keeps going after the first error) mean tests should pin the **first** diagnostic code, as `phase28Classify` does (`result.Diagnostics[0].Code`).

### Spike 008 harness design (flag (d); design only, nothing measured here)
- **Location/convention.** `.planning/spikes/008-byteview-cap/README.md` (frontmatter like spike 007: `spike`, `idea: ownership-kernel`, `name`, `type: standard`, `validates`, `verdict`, `related`, `tags`), one MANIFEST row (spikes/MANIFEST.md table), CONVENTIONS: stdlib only; external tool runs record exact toolchain, target, flags, content digest, compile time, artifact size; wall time kept out of semantic output; include an armed fault (for example a mutant that adds a per-iteration `let` and must trip capacity) so the harness cannot pass vacuously. Spike 005/006 kept Go code beside the README with their own `go.mod`; spike 007 reused a production test. Prefer an **env-gated Go test** (for example `SCHWAY_SPIKE008=1`) in `internal/compiler/session` or `native` so it is skipped in normal `go test ./...`, and keep fixture sources generated into `t.TempDir()` (Pitfall 6).
- **Lanes (D-12).** interpreter plain and under `go test -race`; native `-O0`, `-O3`, `-O3`+`-flto` with compile and run timed separately. Existing primitives: `native.Runner.Run` returns `CompileTime` and `RunTime`; `Runner{LTO: true}`; the missing application-shell `-O3`/LTO lane is the same test seam as SC3.
- **Workloads.** 65,536 x `0xFF` (expect `2006388466`), 16,384 x `0xFF` (`2967061426`), 4,096 x `0xFF` (`2170679522`); assert the checksum on every run so a fast wrong answer cannot be recorded.
- **Statistics.** 7 runs per lane, drop the first, record median and max.
- **Budget.** native max <= 3 s (10% of `defaultSubprocessTimeout = 30 * time.Second`, native.go:30, which bounds compile and each run separately), interpreter max <= 12 s (10% of `RunCLITimeout = 2 * time.Minute`, testsupport.go:37). There is no interpreter step limit to hit (only `MaxCallDepth = 128` and an output limit).
- **Hosted evidence.** CI already runs `go test ./...` and `go test -race -timeout=20m ./...` on `ubuntu-latest` and `macos-latest` (.github/workflows/ci.yml, `checks` job), which satisfies "one hosted macOS and one hosted Linux run, `-race` included" without a new job; record run ids in the spike README.
- **Baseline datum (not the D-12 record):** a 100,000-iteration let-free scalar loop ran natively in 0.00 s here; treat that only as evidence that the event-free loop shape scales, not as the cap measurement.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| ByteView entry parameter fed hex through the conformance runner (ARCHITECTURE.md flag (f) proposal) | Test-only foreign fixture owner (D-05) | 29-CONTEXT discussion 2026-10-11 | Needed because `MaxApplicationArgumentBytes = 4096`; cost is the bespoke owner shape. |
| Per-byte events | No events; `var`/assignment only | M006 Decision 4 | Loop cost independent of capacity. |
| `ByteView` mode `shared` through interface plumbing | Admission by AST contract + `Callable` bit; Mode stays `"owned"` | this research | No `Mode` plumbing needed in this phase. |

**Deprecated/outdated:** `examples/checksum.schway` (iterator frontier) is retired in Phase 32; do not extend it.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Formatter renders `byte_len(x)` / `byte_at(x, i)` unchanged (derived by reading `format.go`, not run) | Flag (b) | Formatter goldens need a code change after all. |
| A2 | ByteView ability set = `drop` only; `copy`, `share`, `send`, `escape` denied | ByteView representation | If `share` is needed (for example a borrowed view passed on), refusals and `Drops` change; ARCHITECTURE.md is internally inconsistent. |
| A3 | A C17 `if (index >= length) { fputs; exit(65); }` is not removable by `-O3`/LTO | OOB failure path | A lane would show the check removed; unlikely (observable side effects), to be proven by the lanes. |
| A4 | Linux single-argv-string cap (`MAX_ARG_STRLEN`) makes a hex entry impractical at 65,536 bytes | Flag (f) | None for the decision: the 4,096-byte `MaxApplicationArgumentBytes` already excludes it. |
| A5 | D-05 fixture-owner effort is comparable to Phase 23 (seven plans) | Flag (f) | If cheaper, 29-00 can include more; if dearer, split the phase. |
| A6 | The three suggested release-vs-view law codes/namespaces | Flag (c) | Naming only. |
| A7 | Interpreter cost at 65,536 iterations is within the 12 s budget | Harness design | Unknown until 29-00 measures it (map-keyed `values`, per-op map lookups). |

## Open Questions

1. **D-11 vs D-12: when is the cap measured, and through what?**
   - What we know: D-11 makes 29-00 "minimum real ByteView plumbing + the Adler-32 loop, then measure flag (d)"; D-12 requires the measurement "through the D-05 fixture path, not a hand-built proxy"; the D-05 owner shape is Phase-23-sized (flag (f)).
   - What's unclear: whether 29-00 must include the whole fixture owner.
   - Recommendation: two stages. Stage A (29-00): harness-delivered bytes (a Go-injected `[]byte` seam for the interpreter; a gate-only driver appended to the emitted TU for native, disposable and labeled "proxy, gate-only"), enough to prove (a), the classifier/place-gate changes, and an order-of-magnitude reading that triggers a re-plan only if the slowest lane is far above budget. Stage B (last plan): the authoritative D-12 record through the real fixture owner, with hosted runs. Needs user acknowledgment because D-12's wording forbids proxies for the *record*.
2. **Scope of D-05.** Confirm the user accepts a Phase-23-sized fixture-owner workstream, and decide between the literal-case-id source form (D-06) and the cheaper entry-parameter case id (flag (f), cheaper shape 1).
3. **ROADMAP "runnable gain" cannot be met by `schway run --engine=*` under D-06.** Which public commands count (native `build --manifest` + `app run`; interpreter via session API)? Amend the roadmap line with a dated note, or relax D-06.
4. **ByteView abilities** (A2): decide `drop`-only vs "no abilities" and record it before 29-00 code lands (two independent derivations must agree).
5. **Reserved-parameter-name amendment.** Approve renaming `data` in the SC1 text and D-01/D-04 examples (Pitfall 1).
6. **Interpreter "exit 65".** Approve the interpreter-side assertion form (defect outcome + reason + empty output) as the meaning of SC3 for the interpreter.
7. **Defect-with-live-owner semantics** (Pitfall 8): record the rule (abort-like; owner left live; documented) before the OOB control is written.
8. **-O3/LTO application lane:** test-only seam versus a `native.Runner` knob (a knob would change receipts' recorded flags).

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | everything | yes | go1.24.0 darwin/arm64 | none needed |
| Clang | native lanes, adapter | yes (local macOS) | Apple clang 21.0.0 | none; Linux evidence comes from hosted CI |
| Python 3 (`-I`) | pinned-literal oracle | yes | 3.14.4 | Go `hash/adler32` agrees independently |
| Linux host | one-host-required hosted run (D-12) | no (local is macOS) | n/a | hosted `ubuntu-latest` via existing CI `checks` job |
| ASan/LSan | optional OOB sanitizer lane | present on Apple clang [ASSUMED] | n/a | skip with a named skip, as existing sanitizer lanes do |

**Missing dependencies with no fallback:** none.
**Missing dependencies with fallback:** Linux host (hosted CI).

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go `testing`, go1.24.0 |
| Config file | none (go.mod only) |
| Quick run command | `go test ./internal/compiler/syntax ./internal/compiler/core -run 'Phase29\|TestAllOperationKinds\|TestPhase26CoreInventory' -count=1` |
| Full suite command | hosted CI only (`go test ./...` and `go test -race -timeout=20m ./...` on ubuntu-latest and macos-latest); never run the full suite locally (project policy recorded in 28-CONTEXT.md and the session-meta-tests memory note) |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| BYTE-01 | parse `byte_len`/`byte_at`, arity refusals, reserved names, formatter goldens | unit | `go test ./internal/compiler/syntax -run 'Phase29' -count=1` | Wave 0 |
| BYTE-01 | each consumer derives the facts independently; `0x80`/`0xFF` read 128/255 | unit/peer | `go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/pathoracle -run 'Phase29' -count=1` | Wave 0 |
| BYTE-01 | op registry and exhaustive dispatch | unit | `go test ./internal/compiler/core -run 'TestAllOperationKindsRegistered\|TestAllOperationKindsHandledAtEverySite\|TestPhase26CoreInventory' -count=1` | exists (edit) |
| BYTE-01 | interp and cgen semantics, `0x80`/`0xFF` unsigned widening, signed-char mutant killed | unit | `go test ./internal/compiler/interp ./internal/compiler/cgen -run 'Phase29' -count=1` | Wave 0 |
| BYTE-02 | OOB interpreter defect + native `-O0`/`-O3`/`-O3`+LTO exit 65, exact stderr, empty stdout | integration (clang) | `go test ./internal/compiler/native -run 'Phase29OOB' -count=1` | Wave 0 |
| BYTE-02 | evidence peer classifies the OOB defect; schema-2 projection carries the reason | unit | `go test ./internal/compiler/executionpeer ./internal/compiler/session -run 'Phase29' -count=1` | Wave 0 |
| BYTE-02 | emitted C exit literals and identifier namespace stay confined | unit | `go test ./internal/compiler/cgen -run 'TestEmittedCExitLiteralsMatchTable\|TestGeneratedIdentifierNamespacesStayConfined' -count=1` | exists |
| BYTE-03 | admitted loop shape; each neighbor refused with its code; FLOW-02 unchanged | unit | `go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate -run '^(TestBackEdgeRejected\|TestPhase26(FixedPoint\|AnalysisExhaustion\|PeerIndependence\|BackEdgeAuthority\|BackEdgeCauseCategories\|CFGBackEdgeCategories\|ScalarCopyCycleBoundary)\|TestPhase29)' -count=1` | exists + Wave 0 |
| BYTE-03 | release-while-view-live refused independently by check, corevalidate, pathoracle (seeded faults) | unit | `go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/pathoracle -run 'Phase29Release' -count=1` | Wave 0 |
| BYTE-03 | no per-byte events: zero `OpCopy` in the loop, event count identical for 4,096 and 5,552 bytes | unit | `go test ./internal/compiler/check ./internal/compiler/interp -run 'Phase29NoLoopEvents' -count=1` | Wave 0 |
| APP-10 | six pinned vectors on interpreter and native; wrong-modulus, off-by-one (`i <= n`), signed-char, swapped-half mutants killed; C and Go byte generators identical | integration | `go test ./internal/compiler/session ./internal/compiler/native -run 'Phase29Adler' -count=1` | Wave 0 |
| APP-10 | fixture owner end to end (acquire/borrow/call/release), adapter manifest coverage | integration | `go test ./internal/compiler/native -run 'Phase29Fixture' -count=1` | Wave 0 |
| meta | corpus registers, LANGUAGE-MATURITY counts, emitter inventory, groundedness, validation lifecycle | unit | `go test ./internal/compiler/session -run 'Inventory\|Groundedness\|Reconciliation\|ValidationLifecycle\|TestPhase28EntryExitClassManifest\|TestPhase272AcceptedCorpusByteIdentity\|TestPhase273ConformanceCorpusByteIdentity\|TestLanguageMaturityCountsAreCurrent\|TestExitCodeDocsTable' -count=1` | exists |
| D-12 | cap measurement (env-gated spike harness) | spike | `SCHWAY_SPIKE008=1 go test ./internal/compiler/native -run 'Spike008' -count=1 -v` and the same with `-race` | Wave 0 (29-00 Stage A, final record at phase end) |

### Sampling Rate
- **Per task commit:** the focused `-run` for the package touched, plus the `session` meta-test set above whenever a `.schway` file, a `session.go` call site, or a planning document with a `go test` row was touched.
- **Per wave merge:** all `Phase29` rows for the wave plus `go test ./internal/compiler/session -run 'TestPhase26\|TestPhase27\|TestPhase28' -count=1`.
- **Phase gate:** hosted full and race CI green on both OSes before `/gsd-verify-work`; spike 008 hosted run ids recorded.

### Wave 0 Gaps
- [ ] Characterization pins of today's neighbor-refusal codes (table above) committed before any parser/check change.
- [ ] Baseline green of `TestPhase272AcceptedCorpusByteIdentity`, `TestPhase273ConformanceCorpusByteIdentity`, `TestPhase28EntryExitClassManifest` at HEAD before edits (byte-identity of the existing corpus is the regression boundary).
- [ ] `TestPhase29*` skeletons per package with each seeded fault proven red (no zero-match `-run` patterns; confirm every new pattern lists at least one test).
- [ ] Test-only application-shell `-O3`/`-O3 -flto` compile-and-run helper (SC3 and D-12).
- [ ] Go-side `[]byte` binding seam for the interpreter and gate-only native driver (Stage A), then the real fixture owner.
- [ ] `29-VALIDATION.md` finalized (not `draft`) before any hosted dispatch.

## Security Domain

### Applicable ASVS Categories (Level 1; security_enforcement enabled)

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | n/a (no auth surface) |
| V3 Session Management | no | n/a |
| V4 Access Control | no | n/a (no file or path access in this phase; Phase 31 owns it) |
| V5 Input Validation | yes | Dynamic bounds check on every byte read; `uint64_t` index compared against `uint64_t` length before the load; the fixture case id is a canonical U64 (`app run` input is bounded to 4,096 bytes, native_app.go:25); adapter rejects unknown case ids with a typed failure instead of reading out of range. |
| V6 Cryptography | no (Adler-32 is a checksum, not an integrity or security primitive; do not describe it as one) | n/a |
| V12 Files / V14 Config | minimal | Adapter is test-only, built from a manifest with hashed inputs (`native.ResolveBindings`); no ambient file or network access. |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Out-of-bounds read in emitted C (CWE-125) | Information disclosure / Tampering | Index check precedes every load, never elided, never `assert`; `unsigned char` loads; NULL data only with length 0; mutants and ASan lane over the OOB control. |
| Signed-char widening producing wrong checksum | Tampering | `unsigned char` storage plus `(uint64_t)` widen; signed-char mutant must fail the `0x80`/`0xFF` fixture. |
| Adapter-generated buffer overrun or leak | Denial of service / Info disclosure | Allocation sized exactly from the case table, `max(1, n)` for empty, release exactly once on non-defect paths; OOB defect path documented as abort-like. |
| Evidence forgery (false OOB attribution) | Repudiation | `executionpeer` checks fixed reason, source-place attribution, terminal position; seeded fault per peer. |
| Timeout used as an oracle | Denial of service | Budgets from D-13 against the existing deadlines; no wall time in goldens. |

## Project Constraints (from CLAUDE.md / AGENTS.md)

- Project instructions live in `AGENTS.md`; `CLAUDE.md` adds skill routing only (`Skill("spike-findings-ai-lang")` for ownership kernel, loan liveness, FFI contract, differential harness). Findings applied: affine ownership and deterministic lexical release are preserved (releases stay compiler-inserted); loan liveness is a CFG fact (no loan crosses a loop back edge); optimizer attributes derive from proven facts (no `restrict`/`noalias` on the view); independently written implementations plus armed fault injection (one seeded fault per peer); distinct native lanes at `-O0`/`-O3`/LTO plus sanitizer; memoized interprocedural summaries (the existing `Callable` table is the mechanism).
- Go 1.24 stdlib only; dependencies must earn more than copy-local; text source is formatter-owned; readable C17 to installed Clang; stable machine-readable diagnostics.
- GSD workflow enforcement: file-changing work goes through a GSD command.
- Living-document rule: at milestone kickoff, before discussion/planning and after completion read `.planning/PRODUCT-ROADMAP.md` and `.planning/LANGUAGE-MATURITY.md`; scan `.planning/seeds/` and `.planning/sources/DASHBIT-AI-ERA.md`; record a scoped disposition. (This research did not execute that transition check; the planner should.)
- "Prefer a runnable gain in each feature phase"; "keep independent validation and negative controls tied to the capability being introduced"; no new backend; no invented completeness percentages.
- Focused tests only locally; full suites run on hosted CI (recorded in 28-CONTEXT.md; memory note `session-meta-tests-in-focused-runs`).

## Suggested Plan Decomposition (advisory)

1. **29-00 Wave-0 gate:** characterization pins; ByteView type + abilities; the two ops through check -> corevalidate -> originvalidate -> interp -> cgen with harness-delivered bytes; assert (a) Callable == true and pinned vectors on both engines; Stage-A cost reading; written verdict (stop and re-plan on failure).
2. **29-01 Surface:** parser atoms, reserved names/arity refusals, formatter goldens, repair decision.
3. **29-02 Peers:** independent scalar fixed points, structural laws, kind-exclusive fields, seeded faults per peer, neighbor-refusal fixtures, FLOW-02 unchanged.
4. **29-03 Engines and evidence:** interp op + OOB; cgen both shells; `executionpeer`, `ProjectExecutionSchema2`, evidence-size bound; application-shell `-O3`/LTO lane; OOB matrix.
5. **29-04..: Fixture owner (D-05)** as its own workstream (likely 2-3 plans): check shape, three peers + release-vs-view laws, cgen/interp binding, adapter + manifest + generator-equality test, session admission.
6. **29-0N APP-10 vectors and mutants:** six literals, mutants, byte-identity table test, cap-scaling workloads.
7. **29-0N+1 Closure:** spike 008 final record (D-12/D-13) with hosted runs, register updates (the three corpus maps, LANGUAGE-MATURITY, emitter inventory), `29-VALIDATION.md`, dated amendments (parameter name, runnable-gain wording).

## Sources

### Primary (HIGH confidence; read or executed this session)
- Repository source: `internal/compiler/{syntax/parser.go, syntax/lexer.go, syntax/format.go, ast/ast.go, core/core.go, check/check.go, check/scalar.go, ability/ability.go, corevalidate/corevalidate.go, originvalidate/originvalidate.go, pathoracle/pathoracle.go, interp/interp.go, cgen/cgen.go, cgen/cgen_program.go, executionpeer/executionpeer.go, native/{native.go, native_app.go, bindings.go, foreign_resource.go}, session/{session.go, session_admit.go}, exitcode/, testsupport/testsupport.go}`, `cmd/schway/main.go`, `examples/{phase23,phase24,sum_to_n.schway}`, `.github/workflows/ci.yml`.
- Tests read for guard behavior: `core/core_test.go`, `cgen/cgen_exit_literals_test.go`, `cgen/cgen_names_test.go`, `session/{session_phase26_overflow_test.go, session_phase28_manifest_test.go, session_phase272_identity_test.go, session_phase273_identity_test.go, session_phase16_emitter_inventory_test.go, self_describing_docs_test.go, verification_groundedness_test.go, session_phase28_exitdocs_test.go}`.
- Planning: `29-CONTEXT.md`, `REQUIREMENTS.md`, `STATE.md`, `ROADMAP.md` (Phase 29), `research/SUMMARY.md`, `research/ARCHITECTURE.md`, `spikes/CONVENTIONS.md`, `spikes/MANIFEST.md`, `spikes/007-loan-across-branch/README.md`, `28-RESEARCH.md` (format and Validation Architecture precedent), `.claude/skills/spike-findings-ai-lang/SKILL.md`.
- Executed probes (scratchpad, built `./cmd/schway`): parse/check of ByteView and builtin spellings; neighbor-refusal shapes; straight-line call of a scalar `%` loop (check, build, `app run`); `run --engine` on U64 entries; Python 3.14.4 `zlib.adler32` and Go 1.24.0 `hash/adler32` for all pinned and scaling vectors.

### Secondary (MEDIUM confidence)
- `research/PITFALLS.md` and `research/STACK.md` (not re-read line by line; claims here rest on the code reads above rather than on them).

### Tertiary (LOW confidence)
- Kernel `MAX_ARG_STRLEN`, ASan availability, and C17 optimizer behavior for the bounds check: training knowledge, tagged [ASSUMED]; the lanes will evidence the last one.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH (no new dependencies; toolchains probed).
- Architecture / flag verdicts: HIGH for (a), (b), (c), (e) (code read and probes); MEDIUM-HIGH for (f) (the bespoke-shape finding is certain; the cost estimate is a judgment).
- Pitfalls: HIGH (each reproduced or read in source), except the formatter claim (A1).
- Cap evidence: not measured (by instruction); harness designed.

**Research date:** 2026-10-11
**Valid until:** 2026-11-10 (stable codebase; re-check line numbers if Phase 28 follow-ups or session.go edits land first)
