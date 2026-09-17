# M003 Adversarial Synthesis

**Written:** 2026-09-17 · **Tree:** `908bac3`, working tree clean
**Method:** every resolution below is grounded in a file:line I opened or a command I ran in this
session, with the output quoted. Where a document's claim survived, I say so; where it did not,
I say which document was wrong and why. The six research documents were treated as hypotheses.

**Headline.** The single most consequential thing I found is not in any of the six documents:
**`match` is a whole-function-body form whose scrutinee must be the function's own parameter, and
a linear body has no branch form at all.** Branching on a computed value therefore does not exist,
is not reachable by desugaring, and is not cheap. That fact simultaneously refutes
ROADMAP-STRATEGY's "`match` already *is* `if`", understates CONTROL-FLOW-ARITHMETIC's `if` cost by
roughly a phase, and breaks the arithmetic→comparison→`if`→D-12-43 chain that three documents rely
on. It is the load-bearing correction in this report.

---

## Contradictions resolved

### C1 — Is the emitter work a delete or a port? **EMISSION-AND-EVENT-IDENTITY wins, decisively.**

Both halves of TYPE-WIDENING's premise are true, and its conclusion still does not follow.

Verified:

- `cgen_program.go:239` — `fmt.Fprintf(&out, "static %s %s(%s);\n", typeNames[index], functionNames[index], typeNames[index])`. One name, return and parameter position.
- `cgen_program.go:363` — `fmt.Fprintf(out, "static %s %s(%s %s) {\n", typeName, functionName, typeName, locals[parameter.ID])`. Same.
- `cgen.go:2570-2579` — `linearInput` switches on `function.Parameter.Type` **only** and returns one `typeName`. Every legacy emitter derives its C type the same way (`emitLinear` at `cgen.go:318` is the literal first line of its body).
- `cgen_program.go:143-161` — `emitProgram` returns a named error for `function.Match != nil`, for `len(function.Linear.Blocks) > 0`, and for `function.ForeignContract != nil`.
- `cgen_program.go:299` — `"live_resources":[]` is a string literal, not a derivation.
- `restrict` appears in `cgen_program.go` exactly once, at line 88, in a comment saying it never mints one.
- `cgen.go:112-137` / `142-171` — `Emit`/`EmitNative` fork on `len(program.Functions) != 1`; N=1 routes to the six legacy emitters, N>1 to `emitProgram`.
- `cgen_n1_convergence_test.go` — the measured table: `emitProgram` **refuses** 4 of 5 shapes and differs in bytes on the 5th.

Measured emitter body sizes (from `grep -n '^func ' cgen.go`): `emitMatch` 235-316 (82), `emitLinear`
317-492 (176), `emitLinearBorrowedByPointer` 615-748 (134), `…Plain` 809-930 (122),
`emitLinearForeign` 953-1147 (195), `emitBranch` 1757-1916 (160) = **869 lines**, plus
`emitBranchOperations` 1979-2236 (**258 lines**) which only the branch family reaches.

**The settling point.** Widening the return type does not force the deletion; it forces the
widening to be *paid twice*. A single-function program with `return ≠ parameter` routes to
`emitLinear`/`emitBranch` by `cgen.go:118`/`148`, never to `emitProgram`. So the two-type model has
to land in `emitProgram` **and** in the four surviving linear/branch emitters. TYPE-WIDENING's
sentence — "`cgen_program.go`'s prototype emission physically cannot survive two distinct type
names via one `linearInput` call" — is true of `emitProgram` and says nothing about the legacy
family's fate. That is a *cost argument for doing the port*, which is EMISSION's argument, not a
mechanism that performs the deletion.

**Defensible plan/effort estimate.** Port branch/match into `emitProgramFunction` + retire
`emitMatch`/`emitBranch`/`emitLinear`: ~600-800 lines touched, dominated by `emitBranchOperations`'s
payload arms and by converging the preamble without perturbing the two unaffected frozen goldens
(`previousPhaseGoldenCDigests`, `core/core_test.go:155-160`, pins four). **8-10 plans, one full
phase, not a task.** The other three (`emitLinearForeign`, both by-pointer variants) should be cut
explicitly — and D-10-60's own text authorizes that: I read it at `PHASE-10-DEBT.md:57` and it says
verbatim that a twice-deferred item *"must either be cut from the milestone explicitly via a
REQUIREMENTS.md amendment, or becomes automatically never-cut. It cannot silently acquire a third
landing phase."* The audit's "land it or retire D-10-60" is a false binary. EMISSION read the rule
correctly and TYPE-WIDENING did not read it at all.

### C2 — What goes first? **EMISSION's order wins, but its dependency argument is overstated. Winner: emitters+identity first, then widening.**

Is the dependency real? Partially, and I measured which part.

```
$ go run ./cmd/lang run --engine=interpreter testdata/phase11/multi_function_diamond_call.lang
… fn:leaf:op:0:event:returned …
… fn:left:op:1:event:returned …
… fn:leaf:op:0:event:returned …      <-- collision, byte-identical ID
… fn:right:op:1:event:returned …
… fn:main:op:2:event:returned …

$ go run ./cmd/lang run --engine=native testdata/phase11/multi_function_diamond_call.lang
result:… run operational_failure
diagnostic:… native.invalid_execution [0:0]: native execution did not complete successfully
```

`native.go:586-589` refuses on `errors.New("duplicate execution event id")`.
`session_phase11_differential_test.go:167-179` shows `AllComparableFixtures` is **five** fixtures,
all straight-line chains; the diamond is quarantined into a `DiamondSharedLeaf` subtest at line 300
that *asserts the gap persists* and fails if it is ever silently fixed.

**So:** no fixture that re-invokes a callee is four-tier compared today. That half of EMISSION's
claim is VERIFIED. But the emitter port's authorizing corpus does not *need* a re-invoked callee —
it needs multi-function programs *with branch bodies*, of which the phase11 corpus has zero for a
different reason (`emitProgram` refuses `Match`). A chain `main → callee-with-a-match` needs no
identity fix. **The dependency caps corpus richness; it does not gate the port.**

The order still holds, for two reasons EMISSION states and one it states better than TYPE-WIDENING:

1. Identity is ~300-400 lines and touches `interp.go` + `emitProgram`'s event derivation. Port the
   branch arms first and you write their event emission twice — STANDING-VERDICTS' own
   "two coexisting laws is a defect with a delayed fuse."
2. Widening after the port pays the two-type model once instead of twice (C1).
3. Widening *before* the port means carrying the two-type model through ~870 lines of legacy
   emitter while simultaneously converging them — two changes in the same lines, the exact
   fingerprint PROJECT.md's own Key Decisions table blames for M001 Phases 2-4.

TYPE-WIDENING's order (i) is rejected on C1's evidence. CONTROL-FLOW-ARITHMETIC's order (ii) puts
widening before emitters and is rejected for the same reason — and its own adversarial pass A2
already conceded this, then did not reorder P14/P15 to match.

### C3 — Does surface `if` exist, and is it cheap? **Neither document is right. Both are refuted on the code.**

What I read:

- `syntax/parser.go:310-319` — the function body is parsed as **either** `p.matchExpr()` **or**
  `p.linearBody()`. A `match` is legal only as the immediate body of a function.
- `ast/ast.go:100-109` — `type Body struct { MatchExpr; Linear *LinearBody }`, with
  `HasClosedVariant()` enforcing exactly one.
- `ast/ast.go:147-159` — `LinearBody` is `Bindings []Binding` plus a `Result string`. There is no
  branch form inside a linear body, at any position.
- `check.go:258` — `if function.Body.Scrutinee != function.Parameter.Name { … "match scrutinee is
  not the function parameter" }`.
- `check.go:2163-2177` — `checkBranch` constructs the whole function: `parameterID :=
  functionID + ":place:0"`, `Places: []core.Place{{ID: parameterID, …}}`, then entry/arm/join
  blocks. The arms start from a place set containing **only the parameter**.
- `ability.go:136-162` — the four type constructors are `Byte`, `Buffer`, `Box`, `Pair`. There is no
  `Bool`. The keyword set (`grep -oE '"[a-z_]+"' syntax/token.go`) is
  `because borrow data defect discard export fn foreign let match module mut take try type` — no
  `if`, no operators.

**Therefore:**

- **ROADMAP-STRATEGY is REFUTED.** "`match` over a two-alternative `data` type already is `if`" is
  false. You cannot match on anything but the function's parameter, and you cannot match after
  computing anything. There is no construct in the language that branches on a value a program
  produced. Rejecting `if` "permanently" on the grounds that its job is already done rejects a
  capability the language does not have.
- **CONTROL-FLOW-ARITHMETIC is REFUTED on cost.** "`if` desugars into the arm-terminating two-block
  topology `check`/`cgen`/`interp` already build for `match`" is false, because the desugaring
  target does not accept a non-parameter scrutinee and cannot be preceded by bindings. `if` is not
  "a parser desugaring plus a `Bool` data declaration." It is: a new AST shape (a branch in a linear
  body's terminal position, or a generalized scrutinee), a `checkBranch` re-architecture so the
  entry block can carry a straight-line prefix, and an ownership question that does not exist today
  (a loan born in the prefix and live into one arm but not the other).
- **Does adding comparison force a `Bool` into existence?** Yes. A comparison must return something
  matchable, and `match`'s scrutinee resolution is `types[function.Parameter.Type.Constructor]`
  (`check.go:250`) — a `data` type lookup. `Byte` 0/1 is not matchable. So comparison forces a
  builtin two-alternative `data` type. **But a `Bool` you cannot branch on is inert** — which is
  precisely the DX-06 / D-13-10a / D-12-43 failure shape, three times already recorded.

**Ruling.** Do not add `if`. Do not add comparison operators either — they are inert without the
branch. What M003 should consider is the thing both documents talked past: **generalize `match` to
be a terminal form of a linear body over any in-scope place of a `data` type.** One construct, not
two laws, and it is the actual missing capability. Its cost is a phase, not a desugar.

### C4 — Arity-N in M003 or M005? **ROADMAP-STRATEGY wins. M005.**

Verified: `ast.FuncDecl` carries a singular `Parameter ast.Parameter` (`ast/ast.go:69-76`);
`check.go:3444-3451` refuses any non-unary call with `check.call_arity_unsupported` and a
`supported_arity: "1"` cause; `pathoracle_compose.go:62-85` lists `ParameterContract.Mode` with
cardinality **1** and the literal words `NAMED EXCLUSION (D-07-01) … arity fixed at 1 … The whole
dimension collapses -- this is STATED, never a silently pruned dimension`, concluding *"Roughly a
dozen structurally distinct boundary-crossing cases."*

TYPE-WIDENING's own numbers settle it against itself. It prices arity-N as "Step 2, two phases,"
then documents that `OpCall` must gain `SourceIDs []string` — a `core` schema change at six dispatch
sites, which STANDING-VERDICTS prices at a **minimum of 8-10 independent edits** each, and which
M002 spent seven phases and 61 plans doing exactly once. It further documents that arity-N reopens
the collapsed `Mode` dimension to 3^N (≈12 → ≈108 cases), breaks `deriveReturnOrigin`'s hardcoded
`Paths: []string{function.Parameter.Name}` (`originvalidate.go:334`) and three
parameter-terminated backward walks, turns `interproceduralSummary`'s scalars into vectors, makes
`loanLivenessBound` arity-dependent, and makes `restrict` soundness depend on a multi-argument
disjointness rule that does not exist. That is a milestone, and TYPE-WIDENING's own risk lens says
so before its recommendation overrides it.

Arity-N is out of M003. The schema capacity is already taken
(`core.FunctionSignature.Parameters []ParameterContract`), so the deferral stays cheap to reverse.

### C5 — Is DX work in M003? **Partly. DX-AGENT-LOOP's own concession is right; its residual splits cleanly, and I verified both halves.**

**Unconditional (a defect in a shipped versioned API — fix regardless of M003's scope).** Reproduced
from scratch, three structurally distinct programs:

```
$ lang check a.lang   # if value { value } else { value }
result:60dde6ee5f1a5620b3e39dee check invalid
diagnostic:5bfc4de388fd6dd1092987f8 syntax.expected_rbrace [68:73]: expected `}`
diagnostic:95cbbd7b69f1d44c281a4039 syntax.expected_declaration [74:75]: expected `data` or `fn` declaration
$ lang check b.lang   # if value { value }            -> IDENTICAL result ID, IDENTICAL diagnostic IDs
$ lang check c.lang   # if value                      -> IDENTICAL result ID, IDENTICAL diagnostic IDs
```

Byte offsets: `if` occupies 65-67; the first diagnostic points at 68-73 (`value`). **Confirmed one
token late.** `lang-repair --json` on the same file returns
`{"status":"unrepairable","subprocess_count":1}` — empty `diagnosis`.

One correction to DX-AGENT-LOOP: the outputs are **not** byte-identical. `recomputed_work` differs
(51 / 43 / 37). Everything that constitutes the protocol's *answer* — status, `id`, both diagnostic
IDs, codes, spans, messages — is identical. The claim is right where it matters and overstated in
its wording.

**Contingent (shrinks as features land).** The `not_in_language` manifest and
`surface.not_in_language` are worth less the more M003 ships. And M003 as I recommend it ships
neither `if` nor operators, so the manifest's entries barely move.

**Recommendation.** Take exactly two things into M003, both as plans inside the evidence phase, not
as a DX phase: (1) **DX-10, `diagnostic_distinctness`** as a hard, deterministic gate metric with
the three-program spiral as its pinned negative control; (2) **DX-12, `unrepairable` carries a
non-empty `diagnosis`**. Both are API-contract defects, both are measurable today, both cost ~1 plan
each. Defer DX-08/09 (manifest, `surface.not_in_language`) and all of DX-11/13. Note that DX-10 is
*partially satisfiable without new diagnostics*: three different programs with different parse
outcomes should not mint one result ID, and enriching the identity content (a token-specific code,
a token-specific span) is the fix DX-AGENT-LOOP itself identifies as determinism-preserving.

---

## Claim spot-checks

| # | Claim | Verdict | Evidence I produced | Charter impact |
|---|---|---|---|---|
| 1 | DX-06 does **not** close when the invariant lifts | **VERIFIED, and stronger than claimed** | `grep -rn 'resolveBlame(' internal/ cmd/` → definition at `check.go:769` + four *test* call sites (`check_blame_test.go:333,372,475,491`). `resolveCycleBlame(` and `classifyDeclaredCause(` likewise have **zero** non-test callers. `classifyDeclaredCause` returns `Violated: false` on both branches (`check.go:735`, `:737`); `Violated: true` appears only at `check_blame_test.go:472,473,488,489`. **The entire blame subsystem is test-only dead code.** | PROJECT.md:38-46, the M002 audit's Cluster C, and PHASE-13-DEBT all contain the same wrong claim. **Correct them in M003's first plan.** Type widening is necessary-not-sufficient; DX-06 needs `resolveBlame` wired *and* a `Violated`-producing classifier branch. Ratify D-13-02b as permanent with the reopening condition corrected to name a declared contract field the declaring function's own admission cannot verify — i.e. separate compilation. |
| 2 | `-flto` is structurally inert on every cgen-emitted multi-function program, and unowned | **VERIFIED** | `emitProgram` emits one TU; refuses `ForeignContract` (`cgen_program.go:153-155`); the only `restrict` token in the file is the comment at `:88`. `session_phase11_differential_test.go:78-89` says so itself (D-11-25). `TestLTOTierIsNotInert` (`native_lto_test.go:154-178`) uses `testdata/phase5/inline_across_foreign.lang`, whose cross-TU boundary is a **foreign** C unit; NAT-07's control (`:181-200`) is hand-written C, `NOT emitted by cgen (D-11-24)`. `grep 'D-11-25' PHASE-11-DEBT.md` → **no match**; it is absent from the audit's ten unowned rows (`M002-MILESTONE-AUDIT.md:309-320`). | Real unowned gap behind a shipped PROJECT.md bullet. File it as a warning-severity debt row with a reopening condition, and qualify the NAT-07 Validated bullet to name the fixture its non-inertness proof rests on. Cheap; do it in M003's evidence phase. |
| 3 | Three VALIDATION `-run` patterns resolve to zero tests, one in a `nyquist_compliant: true` phase | **VERIFIED** | `go test ./internal/compiler/check/... -list 'TestComputeLoanLastUsesAndDerivePlaceLoansAgree'` → no names. Same for `./internal/compiler/session/... -list 'TestAuditQLT02BudgetManifest'` and `./internal/compiler/corevalidate -list 'LoanChainIndex'`. `go test ./internal/compiler/corevalidate -run 'LoanChainIndex' -count=1` → `ok … [no tests to run]`, `EXIT=0`. `09-VALIDATION.md` frontmatter: `status: validated`, `nyquist_compliant: true`; line 85's row carries that dead command and the status cell `✅ green (09-01-SUMMARY.md D1/D2 …)`. The companion row at line 84 (`'PeerLiveness\|LoanChainIndex'`) resolves to exactly one test, `TestPeerLivenessFileImportsStayIndependent`, which verifies only the import-independence clause. | The groundedness lint (EVD-01) is the single highest-ROI item in all six documents. It is ~40 lines, offline, deterministic, and catches a defect class that compounds with surface width. **Land it before any Nyquist reconciliation**, so the lint does the finding. |
| 4 | `TestEveryMutationMovesItsClaimedAxis` skips two rows despite a "never a hardcoded subset" doc comment | **VERIFIED (read-only; no scratch edit, tree clean)** | `session_phase5_alias_test.go:226-232` claims it *"enumerates every NAT-03 row (never a hardcoded subset)"*; `:240-242` skips on `phase5/allocator_mismatch.lang` or `phase5/retained_pointer.lang` citing `PENDING-05-08`. `AssertMutationMovesAnAxis` (`session_phase5_alias.go:510-524`) handles 5 control IDs and its `default:` arm returns `"… (row not yet subjected — see PENDING-05-08)"`. The second law exists: `Phase5AssertMutationMovesAnAxis` (`session_phase5.go:424-429`) special-cases `ControlSanitizeAllocatorMismatch` and delegates the rest. | **One nuance EVIDENCE got right and readers may miss:** of the two skipped rows, `allocator_mismatch` is `Subjected: true` (`:485`) — a genuine hole — while `retained_pointer` is `Subjected: false` with `EscapeID: "escape:callback-invocation-unsubjected"` (`:491-492`) — an honest declared escape. So it is **one** subjected row excluded by fixture-path string match, plus a coexisting second law. EVD-05 (delete the wrapper, zero per-row exclusions) belongs in M003. |
| 5 | Suite is 188 s, one test 52.5 s = 28 % of wall; VALIDATION docs claim ~60 s | **VERIFIED (numbers slightly higher on my host)** | `/usr/bin/time -p go test ./...` → `real 192.67`, `user 98.77`, `sys 68.29`, exit 0. `session` 190.9 s, `testsupport` 54.9 s, `native` 39.0 s. `TestPhase5CorpusThreeEngineAgreement` **88.80 s**; subtest `enumerated-closure` **58.33 s** = **30 %** of total wall. Stale prose: `08-VALIDATION.md:27` "~60 seconds (full suite)", `:39` "Max feedback latency: 60 seconds"; `11-VALIDATION.md:28,41` the same; `12-VALIDATION.md:26` the same; `10-VALIDATION.md:28` "~120 seconds". | 3.2× error in an input to the project's own feedback-budget reasoning, in a project whose first constraint is feedback latency. Add a `suite_wall_clock_p95` row to `qlt02_budget_manifest.json` (`observed`, not `hard` — it is machine-dependent, per the manifest's own discipline) and content-address the enumerated closure. |
| 6 | `pathoracle` refuses every CFG cycle by name — impossible, not expensive | **VERIFIED** | `pathoracle.go:240-246`: `func (e *backEdgeError) Code() string { return "pathoracle.cfg_back_edge" }`; `EnumeratePaths`'s `walk` carries a `visiting map[string]bool` and returns `&backEdgeError{…}` on re-entry. `MaxPaths = 4096` (`:136`). Independently: `check.detectCFGCycle` (`check.go:2571`) → `cfgBackEdgeDiagnostic` `check.cfg_back_edge` (`:2623`), called at `:2733`. Third: `corevalidate.checkCallGraphAcyclic`. | **What M004 must replace, precisely:** (a) `EnumeratePaths`'s acyclic DFS — with bounded unrolling to a declared depth or a loop-summary oracle, because under a back edge the set of *executions* is infinite while the set of acyclic path shapes stays finite, so the oracle's claim silently weakens with no test going red; (b) `loanLivenessBound = 4 * blocks * (loans+1)` (`check.go:2653-2658`), whose own comment (`:2648-2652`) names iteration as the re-opening trigger; (c) `materializeLoanEndpoints`'s `point`/`edge` dichotomy (`check.go:2794`), meaningless for a loan live on a back edge; (d) `analyzeArmBody`'s soundness argument (`check.go:2872-2876`, *"mutually exclusive control flow — the two arms' places never collide"*), which a back edge invalidates outright. Note the cited line range in CONTROL-FLOW-ARITHMETIC (2887-2891) is ~15 lines off; the text is real. |
| 7 | cgen's branch emitter writes a match arm's returned value as the compile-time literal `arm.Pattern`, and this is D-12-43 | **VERIFIED** | `cgen.go:1901`: `emitBranchOperations(…, typeName, arm.Pattern)`; `cgen.go:2021`: `fmt.Fprintf(out, "      if (!lang_write_json_string(%s)) return 74;\n", strconv.Quote(returnLiteral))`. Straight-line contrast at `cgen.go:396-400`: `lang_write_buffer_hex(&%s)` / `lang_write_byte(%s)` on `locals[source.ID]` — the real runtime value. D-12-43's own detail section (`PHASE-12-DEBT.md:221-258`) names this exact mechanism: *"`cgen`'s `returnLiteral` … is a literal string baked into the generated C at emission time, from `arm.Pattern`."* | **But the onward inference is REFUTED.** "Arithmetic in an arm makes D-12-43 constructible" does not follow, because (C3) there is no way to write arithmetic in an arm that feeds a branch on a computed value, and more basically because D-12-43's reopening condition is *exposing **payload** bytes on a match arm's return path* — which needs an arm that destructures a payload and returns the bound place, not an arm that adds. What actually makes D-12-43 constructible is a generalized-scrutinee `match` whose arm returns a destructured payload value. That is the P18 recommendation below, and it is a stronger reason for it than arithmetic is. |
| 8 | Guard inventory is 26, not the 32 LANGUAGE-MATURITY.md states | **VERIFIED** | Re-ran the file's own command: `awk '/^func /{f=$0;l=NR} /Functions\) != 1/{print FILENAME": "f}' $(find internal cmd -name '*.go' -not -name '*_test.go')` → **26 matches, 2 packages**: `session` **24** (of which `verifyBorrowedCorpus` alone holds **8**, and `session.go` holds 15 total), `cgen` **2** (`Emit` `:118`, `EmitNative` `:148`), `reduce` **0**. `reduce` now uses the positive form (`reduce.go:485` `len(p.Functions) == 1`, `:851`). Corpus: **115** `.lang` files, **4,189** lines (file records 89 / 3,096). | LANGUAGE-MATURITY.md is stale in the **optimistic** direction on `reduce` and in the pessimistic direction on the guard count. All three researchers who reported 26 were right; their package breakdowns differ only because the `awk` attributes a guard to the last preceding `^func`, which mis-assigns a few within `session.go`. Make the file machine-checked (ROADMAP-STRATEGY's recommendation 1) — ~1 plan, highest leverage of its five. |

---

## Unverifiable claims passed through with a warning

Stated plainly rather than laundered:

1. **EMISSION's effort figures (~600-800 lines, 4-6 plans for the branch/match port; "a phase of its
   own" each for by-pointer and foreign).** I verified the *line counts* (869 emitter body + 258
   helper) but not the plan count. EMISSION itself rates this MEDIUM-HIGH and extrapolates from
   Phase 09. I raise it to **8-10 plans** on the strength of `emitBranchOperations`'s 258 lines of
   payload-arm handling plus four pinned golden-C digests, but that is my judgment, not a
   measurement.
2. **EMISSION's static-path-table C mechanism for `invocation` identity, and its 4096 node
   ceiling.** Unmeasured, self-rated MEDIUM, and the ceiling value is explicitly a guess. The
   `2^depth` unfolding bound is real arithmetic; whether it bites on
   `testdata/phase07/deep_diamond_acyclic.lang` (13 functions) is untested. **Measure it in the
   phase's discuss gate.**
3. **EVIDENCE's combinatorial projections (~4 min / ~12 min / ~75 min).** Self-rated LOW-MEDIUM and
   correctly flagged: they assume arithmetic widens the *chain alphabet*, which it should not. I did
   not re-derive them. Treat as a directional warning about the frozen enumeration, nothing more.
4. **TYPE-WIDENING's inference that DX-06 requires separate compilation.** The two greps behind it
   are HIGH and I confirmed both. The *inference* ("every published contract field is
   producer-derived, therefore only separate compilation creates a B1-reachable field") is an
   argument, self-rated MEDIUM, and I did not exhaustively audit `core.go:289-393` myself. It is
   plausible and I am acting on it, but it is not a grep.
5. **EVIDENCE's "146 distinct (package, pattern) pairs."** I re-ran the three reported failures and
   all three reproduce. I did **not** re-run the full 146-pair scan, so I cannot confirm the
   denominator or that three is the complete count. The direction of error is favourable: a
   differently-shaped command (e.g. a `scripts/verify-*.sh` wrapper) would only *raise* the count.
6. **DX-AGENT-LOOP's prior-art framing (Elm, Roc, Becker).** Self-rated MEDIUM, from secondary
   sources; the document says it could not fetch the Elm article body. I did not check any external
   source in this pass. Every external citation in all six documents is unverified by me.
7. **ROADMAP-STRATEGY's M004-M006 charters and the "end of M005" date for `checksum.lang`.**
   Self-rated MEDIUM. I have no basis to confirm or refute, and given C3 the ingredient list is
   wrong anyway — branching on a computed value is missing from it entirely.

---

## The strongest critique of the consensus charter

The consensus is: *retire emitters + own event identity, then lift the return-type invariant, then
add literals and arithmetic/comparison operators, proven under optimization, with an
evidence gate at the end; loops and arity-N deferred.* Here is the case against it.

**1. It is seven-to-nine phases wearing a five-phase label.** Costed honestly against what I read:
event identity is a schema `/2` bump plus `interp` frame threading plus `emitProgram` emission plus
a peer re-derivation plus comparator field routing (1 phase). The branch/match port is 869 emitter
lines, 258 of helper, four frozen golden-C digests, and a differential whose expectations must flip
rather than be deleted (1 phase, 8-10 plans — C1). The widening touches three `sameType` heads, the
match-arm alternative set, `resolveCallBinding`'s single-constructor comparison, N type facts at
three minting sites, the `Drops`/`Fresh` split **three times independently** in `check`,
`corevalidate` and `originvalidate`, `reduce.go:769`'s silent TypeID fallback, and two C type names
(1.5 phases). Literals + `OpConst` is 8-10 edits at six sites plus a new interpreter value domain
that must preserve D-12-18's byte-identical scalar projection or every serialized-execution golden
in the corpus moves (1 phase). `OpBinary` plus an overflow law plus C promotion discipline plus a
non-importing structural C scan is another (1.5 phases). Evidence and Nyquist is one. That is 7
before anything goes wrong, on a project whose two data points are 62 plans/5 days and 61 plans/6
days **for one `OperationKind`**. ROADMAP-STRATEGY's "40-45 plans, explicitly smaller than M002" is
not defensible; its own arithmetic (20-25 coordinated edits *plus* a subtraction pass *plus* an
admission change) does not include the emitter port at its measured size.

**2. Its headline deliverable is inert, and inert-by-construction is this project's signature
failure.** Comparison operators produce a `Bool`. Nothing in the language can branch on a `Bool`
(C3). So M003 as chartered ships a value form with no consumer — the fourth instance of exactly the
pattern the milestone exists to retire: DX-06 (a resolver with no call sites), D-13-10a (a repair
class with no trigger), D-12-43 (a control with no observable channel). A reviewer would be entitled
to say the project has now built an unreachable feature *in the milestone whose purpose was to make
unreachable features reachable.*

**3. It rationalizes sunk cost in three places, two of them defensibly and one not.**
*Defensible:* the assurance architecture — three non-importing peers, the four-tier differential —
is genuinely transferable and the CompCert/CakeML precedent is real. *Defensible:* D-10-60 as a
rule; retiring a forcing function the first time it fires is how you reacquire the pattern.
*Not defensible:* D-10-60 applied to **this item**. "Delete six emitters" names a mechanism, not an
outcome, and bundles three capability families of wildly different value; two of the three
(`emitLinearForeign`, the by-pointer pair) have no M003 consumer and their port opens two
designed-not-built subsystems (D-11-11/D-11-12 discharge pairs, D-10-34's multi-frame pad). Forcing
those into M003 on the strength of a process rule is textbook sunk cost, and the rule's own text
(`PHASE-10-DEBT.md:57`) authorizes cutting them. *Also not defensible:* re-proving an unchanged,
deterministic 112-program closure for 58 seconds on every push, twice with `-race`, on a language
that cannot add two numbers.

**4. Cheaper milestones that deliver more — two real candidates.** See the next section; one of them
I partially adopt.

**5. The AI-authoring premise remains untested, and the charter spends against it.** ROADMAP-STRATEGY
concedes this fully. If the premise is wrong, the wasted parts are narrow but real: further
`lang-repair` repair classes, the `not_in_language` manifest, DX-11/13, and the framing of
diagnostics as a *product* API rather than an internal one. The evidence architecture, the peers, and
the comparator survive regardless. The S-009 fresh-agent probe is the cheapest experiment available
and nobody has run it. **It should run before the roadmap is fixed, not inside M003.**

**6. The single most likely way M003 ends in another `tech_debt` closeout.** Not the emitter port
and not arithmetic. It is this: a phase attempts branching on a computed value (under the name
"`if` is cheap"), discovers mid-phase that `checkBranch` builds arms from a place set containing
only the parameter and that a loan born before the branch has no endpoint classification, descopes
the branch, and ships `Bool` and comparison operators with no consumer. The milestone then closes
with a new debt row reading *"comparison operators are unconstructible against the current control
flow — reopens when a branch on a computed value exists"* — a fifth unreachable-claim row, and the
accrual rate turns positive for the second consecutive milestone, which is ROADMAP-STRATEGY's own
stated point of no return for the registers as a control.

---

## Probability the falsifiable trigger fires, and the early warning sign

ROADMAP-STRATEGY's trigger: *"If M003's arithmetic phases require a **redesign** — not an extension —
of `loanLivenessFixpoint`, the five-axis comparator, or `corevalidate`'s forward propagation, then
the assurance stack was calibrated against a toy."*

**As literally written — arithmetic only: ~15 %.** Arithmetic operations are ownership-inert: they
read copyable operands and produce fresh places. `loanLivenessFixpoint` (`check.go:2711-2790`) is
keyed on *loans*, not on operation count; `loanLivenessBound = 4 × blocks × (loans+1)` has no term
that arithmetic moves. The comparator compares outcome values and event sequences — a richer value
in `Outcome.Value` is a content change, not a structural one. `corevalidate` needs a new
`OperationKind` case, which is extension by definition. The residual 15 % is the interpreter value
domain: `interp.value` is `{tag, payload string}` (`interp.go:355-370`) with D-12-18's mutation-seam
guarantee that the scalar projection is byte-identical to a plain string. If a numeric domain cannot
preserve that, every serialized-execution golden in the corpus moves — painful, but still extension.

**As the trigger *should* have been written — any branch on a computed value: ~60 %.** This is the
correction. The moment a `let` binding precedes a branch, two things that are true today stop being
true: (a) `analyzeArmBody`'s soundness argument (`check.go:2872-2876`) rests on sibling arms' places
never colliding *because each arm starts from the parameter alone*; with a shared straight-line
prefix, siblings share live places and the argument no longer holds as stated; (b)
`materializeLoanEndpoints`'s `point`/`edge` dichotomy (`check.go:2794`) has no classification for a
loan born in the prefix and last-used in one arm but not the other — today that shape is
unconstructible, because there is no prefix. **That is a redesign of loan-endpoint materialization,
and it is Austral Rule 3's territory arriving a full milestone before anyone budgeted for it.**

**Early warning sign, stated so it is checkable:** the first `.lang` fixture in which a `borrow`
binding is created *before* a branch and is still live inside exactly one arm. Write that fixture
**during the spike, before the phase is planned**. If `loanLivenessFixpoint` accepts it and
`materializeLoanEndpoints` classifies its endpoint without a new `Kind`, the trigger does not fire
and branching is an extension. If either needs a third endpoint kind or a per-arm state merge, the
trigger has fired and branching moves to M004 alongside loops — where it belongs anyway, because
they are the same loan-across-control-flow problem.

**Which phase would show it first:** whichever phase implements the generalized scrutinee. Under my
recommendation that is P18, and it is hard-gated on the spike precisely so the finding arrives at a
planning gate rather than mid-phase.

---

## Cheaper alternative milestones considered

**Alt-1 — Pure debt + evidence + DX, zero new semantics.** Event identity, the emitter port, the
evidence instrument (EVD-01..08), Nyquist reconciliation, D-13-34, the DX API-contract fixes,
LANGUAGE-MATURITY machine-checked. Four phases, ~30 plans.
*Rejected as a whole milestone, adopted as its first half.* It is genuinely cheaper and it retires
eight of the ten unowned items. But it adds zero language surface for the third milestone running,
and LANGUAGE-MATURITY's central finding — arithmetic, iteration, collections are "what actually make
the language writable, and none of them are on any roadmap yet" — would be true for a third
consecutive milestone. More decisively, it fails ROADMAP-STRATEGY's strongest argument: nothing in
this milestone would *load-test* the assurance stack, so the "built right, not yet loadable"
hypothesis stays untested for another cycle.

**Alt-2 — Go straight at arity-N, skip arithmetic entirely.** Rejected on C4's evidence: it is the
`OpCall` schema-change shape at six dispatch sites, it reopens `pathoracle`'s collapsed dimension
from ≈12 to ≈108 cases, it makes `restrict` soundness depend on a disjointness rule that does not
exist, and M002 is the ratified precedent that this shape costs a milestone. It also delivers no
writable program: `fn f(a: Byte, b: Byte) -> Byte` has no body without operators.

**Alt-3 — Arithmetic only, no branching, no comparison.** Four phases of language work
(literals/`OpConst`, `OpBinary`, overflow-under-optimization, evidence). Genuinely low risk —
arithmetic is ownership-inert, the trigger sits at ~15 %, and it is the first computed value.
*Rejected*, narrowly, because it cannot close D-12-43 (Claim 7), cannot make comparison useful, and
produces `let y = x + 1` — a value change with no decision. It is the correct **fallback** if the
branching spike comes back negative, and I name it as such below.

**Alt-4 (adopted in part) — Branching over arithmetic.** Generalize `match` to a terminal form of a
linear body over any in-scope place of a `data` type. This is the capability the language actually
lacks (C3), it makes M002's own `Result`-payload work usable across a call for the first time, it is
the construct that makes D-12-43 constructible (an arm returning a destructured payload place), and
it is the prerequisite for comparison. It is also the *risky* item — and this project's one ratified
sequencing lesson (deferring `OpCall` out of M001) says the risky structural item gets room, not a
corner of a crowded milestone.

---

## What I would actually do

**M003 — "Make the built things reachable."** Seven phases, ~48-57 plans. Sized honestly as an
M002-scale milestone, because it is one.

| # | Phase | Plans | Gate |
|---|---|---|---|
| **P14** | **Evidence instrument and honest scoping** — EVD-01 groundedness lint (+ non-inertness control), EVD-02 closed evidence vocabulary (`DEFINED\|WIRED\|REACHABLE\|EXERCISED\|MUTATION-KILLED`), EVD-04 no-skip-outlives-its-trigger, EVD-05 delete `Phase5AssertMutationMovesAnAxis` and remove the one `Subjected: true` exclusion, `suite_wall_clock_p95` observed row (baseline 192.7 s), content-address the 58 s enumerated closure, machine-check LANGUAGE-MATURITY, **correct PROJECT.md's DX-06 claim and NAT-07's LTO bullet**, file the `-flto` inertness debt row, DX-10 `diagnostic_distinctness` + DX-12 non-empty `diagnosis`. | 6-8 | The lint reports zero unresolvable `-run` patterns and zero elided command cells; `diagnostic_distinctness` = 1.0 with the three-program spiral pinned as its negative control; the axis-movement law has one implementation and zero per-row exclusions. |
| **P15** | **Event identity** — `lang.execution/2`, required `invocation` field, uniqueness moved to `(invocation, id)` with `/0` and `/1` bytes frozen, interp frame threading, `emitProgram` static path table with a fail-closed ceiling, non-importing peer re-derivation of the admissible invocation set, comparator field routing. | 6-8 | `multi_function_diamond_call.lang` passes `Phase5CompareEngines` across all four tiers; `DiamondSharedLeaf`'s inverted assertion is **flipped, not deleted**, per its own instruction. Closes D-11-51, D-12-21. |
| **P16** | **Branch/match emitter port** — port arm lowering, defect support and payload arms into `emitProgramFunction`; delete `emitMatch`, `emitBranch`, `emitLinear` in the **same commit** that flips dispatch (Phase 09's pattern: three pre-deletion gates, a blocking `checkpoint:decision` plan, same-commit cutover); **formally cut** `emitLinearForeign` and both by-pointer emitters via D-10-60's own amendment clause, refiled as two M004 rows with the LTO consequence stated. | 8-10 | `TestN1ConvergenceDifferential` flips to byte-identity on three of five shapes; `grep -c 'func emitMatch\|func emitBranch(\|func emitLinear('` → 0; each moved golden digest justified in writing. Closes D-11-02, D-12-36, D-11-27; honours D-10-60. |
| **P17** | **Return type ≠ parameter type** — split the three `sameType` heads; mint `functionID:type:N`; resolve arm values against the *return* type; `resolveCallBinding` multi-fact lookup; `Drops`/`Fresh` split derived **three times independently**; fix `reduce.go:769`'s TypeID fallback; two C type names in the one surviving emitter family. | 8-10 | `check.call_argument_type_mismatch` and `check.call_return_type_unrepresentable` each fire from a `.lang` fixture, not only a seam; `use_matching_argument` reaches `repaired` on a sealed held-out fixture. Closes DX-07 / D-13-10a. **Explicitly does not close DX-06.** |
| **P18** | **Branch on a computed value** — generalize `match` to a terminal form of a linear body over any in-scope place of a `data` type; `checkBranch` entry block carries a straight-line prefix. **HARD-GATED on spike S-010** (below). | 8-10 | A fixture calls a `Result`-returning callee and matches on the result, across all five axes; an arm returns a **destructured payload place** and a seeded wrong-slot write is observed on `axis:terminal-outcome` — **D-12-43 constructed, not ratified.** |
| **P19** | **Numeric literals and `OpConst`** — first numeric token, one fixed-width unsigned type, interp value-domain widening preserving D-12-18's byte-identical scalar projection, six dispatch sites. | 6-8 | `TestAllOperationKindsHandledAtEverySite` green; a literal fixture agrees across interpreter / `-O0` / `-O3` / `-O3 -flto`; no serialized-execution golden moves except by reviewed intent. |
| **P20** | **Nyquist, D-13-34, frontier fixture** — reconcile 07/08/11/12/13 **against the new surface**, with the P14 lint doing the finding; 12 and 13 are frontmatter flips done inline; re-trigger D-13-34 now that a structurally distinct move/borrow program is constructible; check in the refused frontier fixture. | 4-5 | Zero `status: draft` VALIDATION files; the frontier fixture's pinned diagnostic has moved; unowned debt count ≤ 5. |

**Mandatory pre-roadmap spikes (both cheap, both answer a gate, neither is production code):**

- **S-010 — does a loan crossing a branch point force a redesign?** Hand-write the fixture: a
  `borrow` binding created before a branch, live inside exactly one arm. Drive it through
  `loanLivenessFixpoint` and `materializeLoanEndpoints`. **If either needs a third `LoanEndpoint`
  kind or a per-arm ownership-state merge, P18 is cut and replaced by `OpBinary` + arithmetic
  under optimization (Alt-3), and branching moves to M004 with loops.** Answer this before the
  roadmap is fixed. This is the S-006 pattern applied to the one thing that can sink M003.
- **S-009 — fresh-agent authoring probe** (ROADMAP-STRATEGY's, adopted verbatim). Cheap, and it is
  the only test of the project's central untested premise.

**Explicitly OUT of M003, each with the reason:**

- **`if` / `else`** — a second control-flow law over a construct that does not yet do the first
  law's job. P18 generalizes the one law instead.
- **Comparison operators and `Bool`** — inert until P18 lands and proves out; shipping them earlier
  repeats DX-06. They are M004's first item.
- **`OpBinary` / arithmetic operators** — deferred to M004 as P19's natural successor, **unless**
  S-010 comes back negative, in which case they replace P18 and comparison stays out regardless.
- **Loops / iteration** — `pathoracle`'s enumerator, `loanLivenessBound`, `materializeLoanEndpoints`
  and `analyzeArmBody`'s soundness argument all fail (Claim 6). M004, with a pathoracle-successor
  spike as a hard entry gate.
- **Arity-N** — C4. M005.
- **DX-06 closure** — ratify D-13-02b as permanent with a corrected reopening condition. Do not
  restate the "closes automatically" claim a third time.
- **`emitLinearForeign`, `emitLinearBorrowedByPointer`, `…Plain`** — formally cut under D-10-60's
  amendment clause, refiled with M004 landings and the LTO-inertness consequence named.
- **DX-08/09 capability manifest, DX-11 rendered diagnostics, DX-13 provenance** — contingent,
  decreasing in value, and DX-11 violates "don't polish a surface a milestone is about to change."
- **Signed integers, multiple widths, floats, a numeric tower** — named and deferred.
- **Separate compilation, modules, a fourth peer, a second comparator** — unchanged.
- **Any Nyquist reconciliation before P14's lint exists** — the lint must do the finding.

---

## The one gate that prevents another `tech_debt` closeout

Not a WIP limit, not a debt-register schema, not Nyquist-as-a-gate — those are all worth doing and
none of them would have caught DX-06, D-13-10a, or D-12-43, because all three were *built correctly
and shipped unreachable*.

> **The constructibility precondition.** No M003 requirement may be admitted at the roadmap gate
> unless a `.lang` fixture that exercises it end to end is **written first** and checked in as a
> refused frontier fixture with its refusing diagnostic pinned by a test. The phase's own gate is
> that the pinned diagnostic **moved**. A requirement whose fixture cannot be written is not
> deferred and not descoped — it is **refused at the gate**, before a plan exists.

Why this one and not the others: it is the only gate that fires *before* the work, and unreachability
is discovered after the work every single time this project has hit it. It converts
"is this claim constructible?" from a post-hoc audit finding into an entry condition. It is the
project's own idiom — an executable escape, a control that fails in both directions — pointed at the
roadmap rather than at the code. And it is mechanically checkable: every requirement row cites a
fixture path and a pinned diagnostic code, and a test asserts the fixture exists and currently
produces that code.

Applied to the consensus charter, this gate would have refused comparison operators on day one,
because no fixture can be written that *uses* a `Bool`. That is exactly the failure this synthesis
exists to prevent.

---

## Confidence + what would change my mind

| Area | Confidence | Basis |
|---|---|---|
| C1 (port, not delete) | **HIGH** | Read both dispatch forks, `linearInput`, both `emitProgram` type-name sites, the three refusals, and the convergence test's measured table |
| C3 (branching on a computed value does not exist) | **HIGH** | `parser.go:310-319`, `ast.go:100-159`, `check.go:258`, `check.go:2163-2177`. Four independent confirmations; this is structural, not interpretive |
| Claims 1-8 | **HIGH** on all eight | Each is a command I ran or a line I opened, quoted above. Claim 5's absolute numbers are one host, one run |
| C2 (order) | **MEDIUM-HIGH** | The corpus cap is measured; the "blocking" framing is my downgrade, and reasonable people could rank the two-coexisting-laws argument differently |
| C4 (arity-N to M005) | **HIGH** | TYPE-WIDENING's own inventory argues against its own recommendation |
| Phase count 7 and plan range 48-57 | **MEDIUM** | Extrapolated from two milestones and the measured emitter surface. Two points is a line, not a curve — the same caveat ROADMAP-STRATEGY states about its own 5 |
| Trigger probability 15 % / 60 % | **MEDIUM** | The 15 % rests on arithmetic being ownership-inert, which is structural. The 60 % rests on my reading of `analyzeArmBody` and `materializeLoanEndpoints`; S-010 is designed to settle it empirically rather than leave it as a judgment |
| "Branching beats arithmetic" as M003's language content | **MEDIUM** | This is the one genuinely contestable call in the document. It trades a low-risk, low-value item for a high-risk, high-value one, and it is defensible only *because* S-010 gates it |

**What would change my mind:**

1. **S-010 comes back clean** — a loan crossing a branch point classifies without a new
   `LoanEndpoint` kind. Then branching is an extension, P18 shrinks, and arithmetic (`OpBinary`,
   overflow under optimization) comes back into M003 as P21-P22, making it an 8-phase milestone that
   should then be split rather than crammed.
2. **S-010 comes back dirty** — then P18 is cut, Alt-3 replaces it verbatim, and branching moves to
   M004 bundled with loops. M003 becomes a 6-phase milestone whose language delta is "a program can
   name and compute a number." Less exciting, honest, and still the first computed value.
3. **Someone shows the by-pointer port is cheap** — i.e. that a callee-side `restrict` on a `static`
   function inside one TU needs no caller-side discharge pair. That moves
   `emitLinearBorrowedByPointer` from the cut half to the landed half, de-fangs the `-flto`
   inertness gap, and is the highest-value thing to spike before P16 is planned. EMISSION names this
   too and it is right to.
4. **A measured `go test -race ./...` figure.** I timed `go test ./...` at 192.7 s and did not time
   the race lane. If the real per-commit cost is 8-10 minutes, the enumeration tiering moves from
   "worth doing" to "must do in P14."
5. **S-009 comes back negative.** If a fresh agent cannot author a small Lang function from the
   structured protocol alone, DX-10/DX-12 survive (they are API-contract defects) but the framing of
   diagnostics as a product API weakens, and the roadmap should tilt toward corpus and examples.
6. **A contract field exists that is developer-written, published in `FunctionSignature`, and not
   fully verifiable at the declaring function's own admission.** That would make DX-06's B1 branch
   reachable without separate compilation and would change the Claim-1 disposition from "ratify
   D-13-02b permanent" to "wire `resolveBlame`." I did not audit `core.go:289-393` exhaustively
   myself and am relying on TYPE-WIDENING's MEDIUM-confidence inference here.

---
*Written 2026-09-17 for M003 planning. Not committed. Working tree verified clean at completion.*
