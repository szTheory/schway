# M003 Research: Type System Widening

**Researched:** 2026-09-17
**Scope:** one decision point — how to widen `sameType(ReturnType, Parameter.Type)`, and how far in M003.
**Every codebase claim below was grepped and read today.** Line numbers are against the
working tree at `908bac3`.

---

## Verdict

**Widen in two ordered steps, and stop promising that widening closes DX-06.** Step 1
(one phase): allow a function to carry **more than one type fact** — distinct return type
at arity 1. This is the true minimal unblock: it makes `check.call_argument_type_mismatch`
and `check.call_return_type_unrepresentable` reachable from source for the first time
(`check.go:1769` currently documents the latter as "UNREACHABLE from any legal source
program"), it genuinely closes **DX-07** (a second in-scope type makes
`use_matching_argument`'s uniqueness gate non-vacuous), it forces the
`D-11-02 → D-12-36` six-emitter deletion as a side effect rather than a third deferral
(because `cgen`'s prototype emitter derives return and parameter C type names from the
*same* `linearInput` call — `cgen_program.go:239,363`), and it repairs a latent silent bug
in the reducer (`reduce.go:769` recovers the parameter's TypeID by falling back to the
first operation's TypeID, justified verbatim by "every function has exactly one parameter
and one type"). Step 2 (two phases): **arity N with user-declared parameter modes**
(`owned` / `borrow` / `borrow mut`, i.e. Mojo's `var`/`read`/`mut`) — this is where the
project's actual asset, the ownership+aliasing assurance stack, gets its first genuinely
new hazard classes and its first non-engineered `restrict` soundness control.

**The correction that matters most:** `PROJECT.md:38-46`, the M002 audit, and
`PHASE-13-DEBT.md` all assert DX-06 and DX-07 "close automatically the moment the
invariant lifts." **DX-07 does. DX-06 does not.** Two independent facts, both verified
today: (a) `resolveBlame` has zero production call sites — `grep -rn 'resolveBlame('
--include='*.go' internal/ cmd/ | grep -v _test` returns only the definition at
`check.go:769`; and (b) nothing in production ever sets `blameFact.Violated = true` —
`classifyDeclaredCause` hardcodes `Violated: false` on *both* of its branches
(`check.go:735`, `check.go:737`), and `Violated: true` occurs only in
`check_blame_test.go:472,473,488,489`. B1 requires a *declared* contract field that the
declaring function's own admission cannot fully verify. Under whole-program compilation
with callee-before-caller admission, no such field exists — and every field
`FunctionSignature` publishes today is producer-derived from the body
(`Mode` is the literal string `"owned"` at `corevalidate.go:2264` and
`originvalidate.go:934`; `Drops`/`Fresh` are derived at `corevalidate.go:2340-2361`).
Type widening creates the *first* user-declarable contract fact the body can contradict —
necessary, but not sufficient: wiring `resolveBlame` and adding a `Violated`-producing
classifier branch are separate, named work. M003 should either scope DX-06 to separate
compilation (M004) or ratify D-13-02b as permanent. It should not restate it as a
side effect of Step 1.

---

## What the invariant buys today (grounded inventory)

### The invariant itself

`sameType(left, right)` is a string-key comparison over `coreType` (`check.go:5047-5049`),
enforced at exactly three admission heads, all independent of any call:

| Site | Function | Diagnostic |
|---|---|---|
| `check.go:255` | bare-value match path in `Check` | `type.return_mismatch` — "S1 match result must have the parameter type" |
| `check.go:3148` | `checkLinear` | `type.return_mismatch` — "linear result must have the parameter type" |
| `check.go:3399` | `checkFallibleLinear` | `type.return_mismatch` — "linear result must have the parameter type" |

`ast.FuncDecl` carries a singular `Parameter ast.Parameter` (`ast/ast.go:69-76`), and the
parser reads exactly one `name: Type` pair between the parens (`syntax/parser.go:302-307`).

### What it actually buys: **one type fact per function**

`sameType` is the *syntactic* enforcement; the load-bearing derived property is that every
function has exactly one `core.TypeFact`, minted as `functionID + ":type:0"`. Seven
non-test sites name that literal:

| File:line | Use |
|---|---|
| `check.go:2169` | `checkBranch` mints the sole type fact |
| `check.go:3173` | `checkLinear` mints the sole type fact |
| `check.go:3406` | `checkFallibleLinear` mints the sole type fact |
| `corevalidate.go:2328` | reads abilities for the published `ParameterContract.Drops` / `ReturnContract.Fresh` |
| `originvalidate.go:917` | same, in the peer's own `BuildInterface` derivation |
| `originvalidate.go:639` | doc-comment invariant for `/0`→`/1` byte stability |
| `core.go:302` | `FunctionSignature.Abilities` doc contract |

Everything below is a **derivation that depends on that single-type-fact property**, not
on `sameType` directly. These are the ~guards that break when it lifts:

1. **`resolveCallBinding`'s return-type derivation** (`check.go:3544-3570`). The comment is
   explicit: *"This function has exactly one type fact (typeFact) … so 'resolved against the
   caller's own type facts' means: the callee's declared return type must equal this
   function's single type fact's own constructor."* Becomes a lookup over N facts.
   **BREAKS (widens).**
2. **`checkCallReturnTypeUnrepresentable`** (`check.go:1767-1775`). Documented as
   *"Currently UNREACHABLE from any legal source program (sameType forces every function's
   return type to equal its parameter type) — mutation-killed through
   `callReturnTypeDerivationSeam`, never through a `.lang` fixture."* Becomes
   source-reachable. **KEPT, newly reachable — needs a real fixture.**
3. **`check.call_argument_type_mismatch`** (`check.go:1755-1765`, emitted at
   `check.go:3542`). Same story. **KEPT, newly reachable.**
4. **The match-arm value check** (`check.go:305-307`). `contains(dataType.Alternatives,
   arm.Value)` where `dataType` is resolved from `function.Parameter.Type`
   (`check.go:251`) — yet the error text already reads *"match result is not an alternative
   of the **return type**"*. The code conflates scrutinee type and return type because
   `sameType` guarantees they are equal. **BREAKS — must split into two lookups.** (This
   is also the *constructive* proof that Step 1 has a writable body form: a `match` over a
   `data Color` scrutinee whose arms yield `data Shape` alternatives.)
5. **`reduce.go:769`** — recovers the parameter's TypeID by *falling back to the first
   operation's TypeID*, justified verbatim: *"every operation in this project's grammar
   shares its function's sole type fact, since every function has exactly one parameter and
   one type."* **BREAKS SILENTLY — this is a latent wrong-answer bug, not a refusal.**
   Highest-priority re-derivation item.
6. **`corevalidate.go:2318-2361`** — abilities read from the *parameter's* single type fact
   are used to compute **both** `ParameterContract.Drops` **and** `ReturnContract.Fresh`.
   Once the return type differs, `Fresh` (does the caller inherit a new drop obligation?)
   is being derived from the wrong type. **BREAKS — a real soundness-adjacent conflation.**
   `originvalidate.go:918-940` has the same shape and must break independently.
7. **`cgen`'s prototype and definition emitters.** `linearInput(function)`
   (`cgen.go:2570-2579`) switches on `function.Parameter.Type` and returns a single
   `typeName`; `cgen_program.go:239` emits `static %s %s(%s);` with that one name in both
   return and parameter position, and `cgen_program.go:363` does the same for the
   definition. **BREAKS — the emitted C prototype is simply wrong.**
8. **`pathoracle`'s declared enumeration case space** (`pathoracle_compose.go:62-85`).
   `ParameterContract.Mode` is listed with cardinality **1** as a *NAMED EXCLUSION*
   (D-07-01), collapsing the whole dimension; the stated product is "roughly a dozen
   structurally distinct boundary-crossing cases." **KEPT for Step 1; explodes in Step 2**
   (arity N × 3 modes → 3^N; at N=2 the space grows from ~12 to ~108 cases — still
   exhaustible, which is the good news).
9. **`corevalidate.peerOriginContained`** (`corevalidate.go:2540-2570`). Already written
   defensively: *"widening arity past 1 (D-07-07) does not silently change this function's
   semantics."* **WIDENED FOR FREE — an existing anticipation that pays off.**
10. **`originvalidate` origin derivation** — `deriveReturnOrigin` hardcodes
    `Paths: []string{function.Parameter.Name}` (`originvalidate.go:334`), and three
    backward walks terminate on `function.Parameter.ID`
    (`originvalidate.go:246,441,1091`). **KEPT for Step 1; BREAKS in Step 2** — with arity
    N the origin must name *which* parameter, and `ReturnContract.Paths` is already a slice.
11. **`interp`** — `partitionFrameForCall` seeds exactly one binding
    (`interp.go:537: seeded := map[string]value{callee.Parameter.ID: argument}`) from a
    single `operation.SourceID` (`interp.go:531`); three entry paths do the same
    (`interp.go:205,219,286`); the match scrutinee's data type is looked up by
    `top.function.Parameter.Type` (`interp.go:722-729`). **KEPT for Step 1; BREAKS in
    Step 2** (`OpCall` must gain `SourceIDs []string`).
12. **`check`'s interprocedural summary** (`check.go:874-899`) carries scalar `usesParam`
    and scalar `parameterMode` read from `signature.Parameters[0]` (`check.go:1059-1063`).
    **KEPT for Step 1; becomes a per-parameter vector in Step 2.**
13. **`cgen`'s `restrict` derivation** (`cgen.go:920-937`) binds the attribute to
    `function.Parameter.ID`, justified by `Operations[0].LoanID`.
    **KEPT for Step 1; is the central Step 2 hazard.**
14. **`loanLivenessBound`** (`check.go:2631-2652`) explicitly names "arity 1" among the
    size assumptions and flags itself as *"a PLANNER ASSUMPTION … that must be re-opened if
    a later milestone makes block or loan count program-controlled."* **KEPT for Step 1;
    must be re-derived in Step 2** (loans per block become O(arity)).

### Correction to LANGUAGE-MATURITY.md's "32-site single-function guard inventory"

Re-ran that file's own re-verify command today:

```
awk '/^func /{f=$0;l=NR} /Functions\) != 1/{print FILENAME": "f}' \
  $(find internal cmd -name '*.go' -not -name '*_test.go')
```

**26 guards, not 32; 2 packages, not 3.** `session` 24 (of which `verifyBorrowedCorpus`
alone holds 8), `cgen` 2 (`Emit` at `cgen.go:118`, `EmitNative` at `cgen.go:148`), `reduce`
**0** — `reduce` now uses the positive form `len(p.Functions) == 1` at `reduce.go:485` and
`reduce.go:851`, so Phase 11 already widened it. Corpus is also stale: **115 `.lang`
programs, 4,189 lines** (the file records 89 / 3,096).

Note these are multi-*function* guards, orthogonal to multi-*type*. They matter here only
because `cgen.go:118/148` are the exact fork that keeps the six legacy single-function
emitters alive (D-11-02 / D-12-36), and Step 1 forces that fork's resolution.

---

## Options considered

| Option | Scope | Unblocks | Cost | Risk |
|---|---|---|---|---|
| **0. Do nothing** | — | nothing | 0 | D-13-02b/D-13-10a cross a second milestone boundary; `reduce.go:769` stays a silent-wrong-answer landmine |
| **1. Distinct return type, arity 1** ("N type facts per function") | `check` (3 heads + match-arm split + call-return derivation), `corevalidate`/`originvalidate` abilities split, `cgen` two type names, `reduce` TypeID fix | **DX-07 fully**; `check.call_return_type_unrepresentable` + `check.call_argument_type_mismatch` source-reachable; forces D-12-36 emitter deletion | ~1 phase | Low. Zero new ownership hazards. Peers already anticipate it in two places |
| **2. Arity 2, fixed, all `owned`** | Option 1 + `ast.Parameters []Parameter`, `core.Function.Parameters`, `OpCall.SourceIDs` | Option 1 + first multi-loan call site | ~1.5 phases | Medium. `OpCall` shape change = D-04-22's 8-10-edit standard at six sites, for a bound (2) you will immediately want to lift |
| **3. Arity N, all `owned`** | Option 2 without the fixed bound | Option 2 + drop-order obligations on N owned args | ~2 phases | Medium. No aliasing hazard (owned args cannot alias post-move) but `f(x, x)` becomes a use-after-move question |
| **4. Arity N + declared parameter modes** (`owned`/`borrow`/`borrow mut`) | Option 3 + `ParameterContract.Mode` becomes *declared* rather than the literal `"owned"`; exclusivity rule at call sites; per-parameter `restrict` | Everything above + the first non-engineered `restrict` aliasing control + the first user-declarable contract a body can contradict | ~2-3 phases | High but high-value. This is where the assurance stack earns its keep |
| **5. Option 4 + local `let` type inference** | + bidirectional inference inside bodies | ergonomics | +0.5 phase | **Reject for M003** — see "Type inference" |
| **6. Separate compilation (consume `lang.interface/1` as admission authority)** | new trust boundary | **the only thing that actually makes DX-06/B1 reachable** | ≥1 whole milestone | High. `STANDING-VERDICTS.md` explicitly warns: "don't bundle modules/separate compilation into a milestone already extending five-plus-consumer machinery" |

---

## Recommended widening + why

**M003 = Option 1, then Option 4. Not Option 6.**

Sequenced:

1. **Phase A — "More than one type per function."** Split the three `sameType` heads into
   independent scrutinee-type and return-type resolution; mint type facts as
   `functionID:type:N`; fix `check.go:305` to resolve the arm-value alternative set against
   the *return* type's `DataType`; make `resolveCallBinding` look up the callee's declared
   return type in a multi-fact map instead of comparing against a single constructor; split
   `Drops` (parameter's abilities) from `Fresh` (return type's abilities) in *both*
   `corevalidate` and `originvalidate`, independently; give `cgen` two C type names; fix
   `reduce.go:769` to carry the parameter's real TypeID rather than infer it.
   **Delete the six single-function emitters in this same phase** — `Emit`/`EmitNative`
   must stop forking on `len(Functions) != 1`, which discharges D-11-02/D-12-36 and
   honours D-10-60 instead of triggering a third deferral.
2. **Phase B — Arity N, all parameters `owned`.** `ast.FuncDecl.Parameters []Parameter`;
   `core.Function.Parameters []Parameter`; `core.LinearOperation` gains `SourceIDs []string`
   for `OpCall`; declare and test a left-to-right argument evaluation order and a
   reverse-of-declaration drop order (Rust's rule — see Prior Art); re-derive
   `interproceduralSummary` per parameter; re-derive `loanLivenessBound` with arity as a
   term.
3. **Phase C — Declared parameter modes and the exclusivity rule.** `borrow` /
   `borrow mut` on parameters, mapping to the *already-reserved* `ParameterContract.Mode`
   values `"shared"` / `"exclusive"` (`core.go:341-362`, validated by `DecodeInterface` at
   `core.go:555-562`). Call-site disjointness rule (Hylo's formulation). Per-parameter
   `restrict`. This is the phase that makes `emitLinearBorrowedByPointerPlain`'s dormant
   `aliasProbeParameterName` machinery (`cgen.go:783-800`) into a *real* differential
   control rather than an engineered one.

**Why Step 1 first even though it adds no ownership hazard.** Two reasons, neither
ergonomic. First, it is the only change that makes two already-shipped, mutation-killed-
but-fixture-unreachable diagnostics (`check.go:1769`, `check.go:1760`) reachable from
source — converting seam-only evidence into corpus evidence, which is exactly the kind of
"wiring is not reachability" gap the M002 audit caught. Second, it is the smallest change
that mechanically forces the D-12-36 emitter deletion, because `cgen_program.go`'s
prototype emission physically cannot survive two distinct type names via one
`linearInput` call.

**Why not stop at Step 1.** The borrow-checker and verification lenses are right that
Step 1 buys ~zero new assurance evidence. `pathoracle`'s own declared case space
(`pathoracle_compose.go:62-85`) says the `ParameterContract.Mode` dimension has cardinality
1 *as a named exclusion* — it is a hole in the exhaustive enumeration that the project
documented rather than hid. Only Option 4 fills it.

---

## Blast radius by dispatch site

Project standing rules: *no fact crosses a trust boundary without being independently
re-derived*; *no callee body is ever read to admit a caller*.

| Site | Step 1 (distinct return type) | Step 2-3 (arity N + modes) | RE-DERIVE or widen? |
|---|---|---|---|
| **`check`** | 3 `sameType` heads; match-arm alternative set (`:305`); N type facts (`:2169/:3173/:3406`); `resolveCallBinding` return-type lookup (`:3560-3570`); arity gate (`:3444-3451`) stays | `Parameters []`; `OpCall.SourceIDs`; per-parameter `interproceduralSummary` (`:874-899`, `:1059-1063`); `loanLivenessBound` arity term (`:2631-2652`); exclusivity rule at call sites; `loanFinalUseEvidence`'s sole-argument resolution (`:4746-4760`) | **Widen** (it is the producer) |
| **`corevalidate`** | `Drops`/`Fresh` split must be derived *from its own* second type fact — must NOT reuse check's split (`:2318-2361`) | `parameterContractMode()` (`:2260-2265`) stops being the constant `"owned"` and must derive the declared mode *independently*; `peerCalleeFrameDrained` drains N params; `peerOriginContained` (`:2540-2570`) already loop-shaped | **RE-DERIVE** — both the split and the mode |
| **`originvalidate`** | Same `Drops`/`Fresh` split, third independent time (`:918-940`) | `deriveReturnOrigin` (`:334`) must name *which* parameter; three parameter-terminated backward walks (`:246,:441,:1091`) become N-terminated; `Mode: "owned"` literal (`:934`) becomes derived | **RE-DERIVE** |
| **`pathoracle`** | no change | the collapsed `ParameterContract.Mode` dimension (`pathoracle_compose.go:62-85`) reopens: 3^arity; `MaxCompositionDepth = 3` unchanged but the per-depth case count multiplies | **RE-DERIVE the case-space argument** — the "roughly a dozen cases" claim must be recomputed and re-published, not silently widened |
| **`interp`** | no change (values are untyped at runtime; `interp.go:722` scrutinee lookup still uses the parameter type, correctly) | 3 entry seeds (`:205,:219,:286`), `partitionFrameForCall` (`:531-537`) multi-arg; per-argument copy-vs-move decision; **declared argument evaluation order becomes observable** | **Widen** (it is the oracle) |
| **`cgen`** | two type names from `linearInput`; `cgen_program.go:239,363`; **six legacy emitters deleted**; `cgen.go:118,148` forks removed | N-ary C signatures; per-parameter `restrict` (`cgen.go:920-937`); `byPointerQualifier` gains a disjointness precondition; `aliasProbeParameterName` control (`cgen.go:783-800`) becomes real | **Widen**, but the `restrict` *justification* must be **RE-DERIVED** from a disjointness fact, not from `Operations[0].LoanID` |
| **`reduce`** (non-dispatch, but breaks) | `reduce.go:769` TypeID fallback is a **silent wrong answer** today's invariant hides | `ProjectSource`/`Reduce` multi-parameter projection | **RE-DERIVE** |
| **`session`** (non-dispatch) | corpus/lane guards; 24 `len(Functions) != 1` sites are orthogonal but co-located | Phase-5/6/7 verification lanes need arity-bearing fixtures | Widen |

**Non-importing peers:** `corevalidate`, `originvalidate`, `pathoracle`. Nothing in Step 1
or Step 2 lets a peer import `check`'s split — the three `Drops`/`Fresh` derivations must be
written three times, and the three transitive-dependency guards must stay clean. The
`callSignatureTable` is built by `check` calling `originvalidate.BuildInterface(program)`
(`check.go:1931-1943`), so `check` already consumes a peer's derivation; that seam does not
change shape, only content.

---

## New hazard classes (ownership × multi-arg)

These appear in Step 2-3 only. None of them exist today.

1. **Same place in two argument positions** — `f(x, x)`. Legal if both positions are
   `borrow` (shared); illegal if either is `borrow mut` or `owned`. Mojo states it as:
   *"if a function receives a mutable reference to a value (such as a `mut` argument), it
   can't receive any other references to the same value."* Hylo generalises it to
   *"multiple `inout` formal parameters are allowed as long as at the call site the actual
   parameters are all disjoint trees."* **Recommend adopting Hylo's phrasing** — it
   degenerates correctly at Lang's current field-less executable shapes (`Byte`, `Buffer`)
   to "distinct places," and it survives the day fields arrive.
2. **Move-and-loan at one operation** — `f(take x, borrow x)`. `OpCall` is currently one
   operation with one `SourceID`, consumed atomically (`interp.go:531`,
   `check.go:3455-3460`). With `SourceIDs []string`, one operation creates a *set* of
   ownership effects that must be ordered against each other. `loanLivenessFixpoint` has no
   notion of two loans born at one program point.
3. **Partial-move on argument-evaluation failure.** If arg 1 is moved and arg 2 is refused,
   is arg 1 gone? Today unanswerable. Rust's answer is left-to-right evaluation, with
   arguments *"dropped in reverse order of declaration"* — and the partially-initialised
   state during unwind is exactly what its drop-flag machinery exists for.
   **Lang must declare an order and prove interpreter/`-O0`/`-O3`/`-flto` agreement on it.**
   This is a new axis for the five-axis comparator, not a fixture.
4. **Drop-order obligations on multiple owned arguments.** `ParameterContract.Drops` becomes
   a vector; `peerCalleeFrameDrained` (`corevalidate.go:2324`) must drain N parameters; and
   release *order* becomes externally observable through the existing `checkReleaseOrder`
   machinery. A two-owned-`Buffer` callee is the first program where "drained" is not a
   boolean about one place.
5. **`restrict` unsoundness.** This is the sharp one. Today exactly one parameter can carry
   `restrict`, and its justification is structural (`selectsByPointerLowering`:
   one exclusive borrow of the sole parameter, unbroken to the terminator). With two
   `borrow mut` parameters lowered by pointer, Clang gets two `restrict` pointers, and
   **the correctness of the emitted C now depends on the checker's disjointness rule being
   right.** This is the first time an optimizer attribute's soundness rests on a
   multi-argument analysis — and it is precisely the attack `aliasProbeParameterName`
   (`cgen.go:783-800`) was built to simulate. It stops being an engineered adversary and
   becomes a real one. **Highest-value new differential control in M003.**
6. **Two-phase borrows — avoidable, and should be avoided.** Rust needed `&mut2`
   (RFC 2025) because `vec.push(vec.len())` desugars to a mutable autoref whose borrow
   starts before the argument is evaluated. Lang's call arguments are plain identifiers
   resolved in the caller's `places` map (`check.go:3453-3457`), not nested expressions.
   **Recommend a standing rule: call arguments remain names-only through M003.** That
   single restriction buys immunity from the entire two-phase-borrow design space.

---

## Role-lens disagreements

- **Type theorist vs compiler engineer.** The theorist says go straight to Option 4: affine
  typing with N arguments is standard metatheory — the substructural rule is just
  "the multiset of capabilities consumed at a call site must be pairwise compatible" — and
  staging it in two steps means proving soundness twice. The compiler engineer says
  `OpCall` gaining `SourceIDs` is a `core` schema change at six dispatch sites, which
  `STANDING-VERDICTS.md` prices at a *minimum* of 8-10 independent edits before it is real
  by D-04-22's standard, and that M002 spent seven phases on exactly one such addition.
  **Resolution: the engineer wins on sequencing, the theorist wins on destination.** Step 1
  deliberately does *not* touch `OpCall`'s shape.
- **Borrow-checker specialist vs DX lens.** The borrow-checker specialist calls Step 1
  worthless: it adds zero aliasing hazards, so it produces no ownership evidence, and the
  interesting question (call-site disjointness) is entirely in Step 3. The DX lens
  disagrees: Step 1 alone closes DX-07, and `lang-repair` is a first-class product surface
  (`PROJECT.md` key-decisions row: *"Treat structured diagnostics and evidence as a
  versioned product API"*), so a permanently-unrepairable defect class is a product hole,
  not a cosmetic one. **They do not reconcile.** Both are right about different currencies.
- **Verification engineer vs project/risk lens.** The verification engineer wants Step 3 in
  M003 because it supplies the first non-engineered `restrict` control and repopulates
  `pathoracle`'s collapsed enumeration dimension. The risk lens observes that Steps 1-3 plus
  the two M002 debt clusters plus Nyquist closure is already a milestone larger than M002,
  and that LANGUAGE-MATURITY.md's warning — *"analysis or corpora that were linear on one
  function becoming superlinear across a call graph"* — applies directly: loans per call
  site go from 1 to arity, and `loanLivenessBound` is `4 × blocks × (loans+1)`.
  **Resolution: Step 3 needs an S-006-style cost-scaling spike as a hard planning gate**,
  exactly as Phase 08 did.
- **Backend engineer, alone in agreement with everyone.** Step 1 is genuinely cheap at the
  backend (two `linearInput` results instead of one). Step 3 is where the C ABI questions
  are, and none of them are hard *except* `restrict`.
- **Everyone vs `PROJECT.md`.** All five lenses agree DX-06 does not close here. The
  document should be amended before M003's requirements are written.

---

## Prior art and lessons

**Mojo — the closest analogue, and the one to copy.** Mojo's five argument conventions map
almost one-to-one onto fields Lang has *already reserved*: default/`read` → `Mode:"shared"`,
`mut` → `Mode:"exclusive"`, `var` (with the `^` transfer sigil) → `Mode:"owned"`, plus `out`
and `ref` which Lang does not need. Critically, Mojo enforces **argument exclusivity at the
call site**: a `mut` argument cannot coexist with any other reference to the same value —
with a documented carve-out for register-passable trivial types, which are always copied.
Lang already has that carve-out in the form of the copy ability
(`argumentIsCopyable`, `interp.go:532`). **Lesson: adopt Mojo's convention vocabulary
verbatim rather than minting new words, because `core.ParameterContract.Mode`'s closed set
(`"owned"`/`"shared"`/`"exclusive"`, `core.go:349-354`, enforced at `core.go:555-562`) is
already exactly that vocabulary.** Widening is a *population* problem, not a schema problem.
[[docs]](https://mojolang.org/docs/manual/values/ownership) [[blog]](https://www.modular.com/blog/deep-dive-into-ownership-in-mojo)

**Hylo/Val — the best formulation of the multi-argument rule.** *"Multiple `inout` formal
parameters are allowed as long as at the call site the actual parameters are all disjoint
trees, i.e. they do not share any part of their value."* Hylo credits Swift's law of
exclusivity and `_modify`/`_read` accessors as direct influences. **Lesson: state the rule
over *values and their parts*, not over variables — it degenerates correctly today and
scales when fields arrive.**
[[Hylo spec]](https://hylo-lang.org/docs/reference/specification/) [[val-for-swift-users]](https://github.com/hylo-lang/Documentation/blob/main/val-for-swift-users.md)

**Swift — the retrofit cost, and the exception you will want.** SE-0176 ("Enforce Exclusive
Access to Memory") landed exclusivity in Swift 4 and required *dynamic* enforcement in
release builds by Swift 5, because static enforcement could not cover every case. It
explicitly prohibits passing the same variable as two `inout` arguments, but carves out
overlapping accesses to *different stored properties* of the same variable. **Lesson:
Swift needed a runtime check because it added exclusivity after the language existed. Lang
can be fully static because it is adding it before. Do not spend that advantage.**
[[Swift blog]](https://www.swift.org/blog/swift-5-exclusivity/) [[SE-0176]](https://github.com/apple/swift-evolution/blob/master/proposals/0176-enforce-exclusive-access-to-memory.md)

**Rust — two named regrets, both avoidable here.** (1) *Two-phase borrows.* RFC 2025 had to
introduce a compiler-internal `&mut2` form with no user syntax purely to make
`vec.push(vec.len())` compile, by delaying the start of a mutable borrow past argument
evaluation. This exists only because method-call autoref creates the borrow before the
arguments. **Lang avoids it entirely by keeping call arguments as plain names.**
(2) *Polonius.* NLL's location-insensitive analysis is faster but rejects sound programs;
Polonius Alpha landed on nightly only in August 2026, and the pipeline still runs
location-insensitive first and falls back to location-sensitive only on error — an explicit
precision/cost tradeoff. **Lesson: Lang's `loanLivenessFixpoint` is already
location-sensitive on a per-function CFG. Widening to arity N multiplies loans per point;
measure before generalising** — this is LANGUAGE-MATURITY's named precedent, and it applies.
Rust's drop rule is also the one to copy for hazard 3: *"function arguments are likewise
dropped in reverse order of declaration."*
[[RFC 2025]](https://rust-lang.github.io/rfcs/2025-nested-method-calls.html) [[RFC 2094 NLL]](https://rust-lang.github.io/rfcs/2094-nll.html) [[Polonius Alpha]](https://blog.rust-lang.org/2026/08/04/enabling-polonius-alpha-on-nightly/) [[Reference: Destructors]](https://doc.rust-lang.org/reference/destructors.html)

**Austral — the argument for keeping it small.** Austral's linearity checker is
*"less than a thousand lines of heavily-commented OCaml"* and its borrow checker under 600
lines, with lexical lifetimes and an explicit borrow construct, deliberately simpler than
Rust's. References cannot escape their call site because their region type cannot be
written. **Lesson: Lang's existing `borrow(path)` return-origin annotation is already
Austral-shaped. Multi-parameter borrow does not require regions — it requires
per-parameter origin paths, which `ReturnContract.Paths` is already a slice for
(`core.go:375-377`).**
[[Austral spec]](https://austral-lang.org/spec/spec.html) [[how the checker works]](https://borretti.me/article/how-australs-linear-type-checker-works)

**Koka/Perceus — the cautionary note on inferring conventions.** Perceus lists borrow
inference as an optimisation, but Koka *"currently has no automatic borrow inference and
generally only uses borrowing for built-in primitives."* Lorenzen's thesis on optimising
reference counting with borrowing is a whole master's thesis on getting it right.
**Lesson: do not infer ownership conventions. Declare them.**
[[Perceus]](https://www.microsoft.com/en-us/research/uploads/prod/2020/11/perceus-tr-v1.pdf) [[Lorenzen]](https://antonlorenzen.de/master_thesis_perceus_borrowing.pdf)

**Swift's type checker vs Rust's — the inference verdict.** Swift's *"expression was too
complex"* error is a direct consequence of constraint-based expression inference behaving
super-linearly or exponentially; associated-type inference is described in SE-0108 as
*"the only place in Swift where we have a global type inference problem."* Rust chose
bidirectional local inference with **mandatory** annotations at function boundaries
specifically to keep complexity in check. **Lesson for Lang, stated as a rule: annotations
at the signature are non-negotiable.**
[[Why Swift's type checker is slow]](https://danielchasehooper.com/posts/why-swift-is-slow/) [[Type inference in Rust and C++]](https://herecomesthemoon.net/2025/01/type-inference-in-rust-and-cpp/) [[SE-0108]](https://github.com/apple/swift-evolution/blob/master/proposals/0108-remove-assoctype-inference.md)

### Type inference: should M003 introduce any?

**No. Stay fully explicit, and write it down as a standing verdict.** Three reasons, in
increasing order of force:

1. *Empirical.* Rust's local-only, signature-mandatory inference is not regretted; Swift's
   more ambitious inference is, loudly and expensively.
2. *Structural.* Lang's evidence model is content-bound and digest-keyed; inference makes
   a signature a *derived* artifact, so any change to inference changes every downstream
   digest. `ClosureDigest`/`CoreDigest` invalidation becomes inference-sensitive.
3. **Decisive, and specific to this project: inference destroys blame.** DX-06's entire
   premise is that a *declared* contract can disagree with a body, and that the disagreement
   has an owner. If modes and types are inferred, the declaration is a projection of the
   body and can never contradict it — B1 becomes not merely unreachable but *incoherent*.
   Inference at the signature is an anti-feature for a language whose product is
   attribution.

Local `let` type elision inside a body is the only defensible concession, and it should
still be deferred: it adds a diagnostic surface and a repair class with no requirement
behind it, and `let x = take y` already has a unique RHS type today.

---

## Adversarial pass

**The strongest case against widening in M003:**

> *Widen nothing. Do arithmetic and control flow first.* LANGUAGE-MATURITY.md's central
> finding is that items 2-4 — arithmetic, iteration, collections — *"are what actually make
> the language writable, and none of them are on any roadmap yet."* A multi-parameter
> function with no `+` operator cannot compute anything; `fn add(a: Byte, b: Byte) -> Byte`
> has no writable body. Widening the *signature* while the *body* stays a single expression
> is expanding the contract surface of a language that still cannot express FizzBuzz —
> precisely the "roadmap vocabulary outruns language surface" trap that file exists to
> prevent. Meanwhile the invariant is load-bearing: it is what lets three non-importing
> peers agree cheaply, what keeps `pathoracle`'s case space at a dozen entries, and what
> keeps `loanLivenessBound` provably small. And DX-06/DX-07 are cosmetic — an unreachable
> blame branch and one missing repair class are not why anyone would use this language.

**Rebuttal, point by point.**

*On "no writable body" — partially conceded, and materially narrowed.* It is false for
Step 1. `check.go:305`'s alternative-set check resolves arm values against the *parameter's*
`DataType` while its own error message already says *"return type"* — so
`fn f(c: Color) -> Shape { match c { Red => Circle, ... } }` is a fully writable Step 1
program requiring no new operators, and `let y = g(x)` where `g: Byte -> Buffer` is another.
It is **true** for Step 3: two `borrow mut Buffer` parameters have no interesting body
without operators. **Concession: Phase C's fixtures will be ownership fixtures, not
programs** — which is what every fixture in this repo already is, but it should be said out
loud in the roadmap rather than discovered in verification.

*On "the invariant is load-bearing for peer independence" — rejected.* The peers re-derive
from `core`, not from `sameType`. Two peer sites already anticipate the widening in their
own comments: `corevalidate.go:2553-2558` writes path containment as a loop explicitly *"so
widening arity past 1 (D-07-07) does not silently change this function's semantics,"* and
`core.go:293-297` took `Parameters []ParameterContract` as a slice on purpose — *"schema
capacity is taken now, with one producer, rather than later across five consumers."* The
project already decided this, in code, in Phase 07.

*On "keeps `pathoracle`'s case space small" — conceded as a cost, rejected as an argument.*
The dimension is collapsed with a **NAMED EXCLUSION** (`pathoracle_compose.go:65-77`),
which is a disclosed hole in an exhaustiveness claim, not a feature. An exhaustive
enumeration over a space one axis of which has been pinned to cardinality 1 by a language
limitation is weaker evidence than it reads as. Filling it is the *point*.

*On "DX-06/DX-07 are cosmetic" — conceded for DX-06, and more strongly than the objection
makes it.* My own finding is that type widening does not close DX-06 at all: `resolveBlame`
has zero call sites and nothing sets `Violated: true`. The honest M003 disposition is to
*stop promising it*. **Rejected for DX-07:** `cmd/lang-repair` repairing defects through the
JSON protocol alone is a ratified key decision, and D-13-10a's own reopening condition —
*"the single-type-per-function invariant relaxes"* — is satisfied exactly by Step 1.

*On sequencing versus the older debt — conceded, and it helps.* D-10-60 forbids a third
deferral of the six-emitter deletion, and D-12-36 already re-deferred it once with a stated
reversal. Step 1 does not merely coexist with that deletion: it **forces** it, because
`cgen_program.go:239/363` cannot emit a correct prototype from one `linearInput` type name
once return and parameter types differ. Doing the widening is the cheapest available way to
honour D-10-60.

**Net:** widen, but ship Step 1 and Step 2 in M003 with Step 3 hard-gated on a cost-scaling
spike, and amend the DX-06 claim rather than carrying it a third time.

---

## Proposed M003 requirements this implies (draft REQ text)

**TYP-01 — A function's return type may differ from its parameter type.**
A developer can declare `fn f(x: A) -> B` with `A ≠ B` and have it check, interpret, and run
natively. *Testable:* a `.lang` fixture with a `match` whose scrutinee is one `data` type and
whose arm values are alternatives of another checks clean, interprets, and produces identical
observable outcomes across interpreter / `-O0` / `-O3` / `-O3 -flto`; the three
`type.return_mismatch` sites no longer fire on it.

**TYP-02 — A caller that passes the wrong type gets a diagnostic with a working repair.**
`check.call_argument_type_mismatch` and `check.call_return_type_unrepresentable` are each
triggered by at least one `.lang` fixture (not only by a fault-injection seam), and
`use_matching_argument` emits a `MachineApplicable` repair that reaches `repaired` — not
`unrepairable` — on a sealed held-out fixture where a second in-scope place of the callee's
declared parameter type exists. *Closes DX-07 / D-13-10a.*

**TYP-03 — Abilities published for the return are derived from the return's own type.**
`ReturnContract.Fresh` and `ParameterContract.Drops` are derived from distinct type facts,
independently in `check`, `corevalidate`, and `originvalidate`, with no shared helper.
*Testable:* a function whose parameter type carries `drop` and whose return type does not
(and the converse) publishes the correct `Drops`/`Fresh` pair; a mutation that swaps the two
derivations is killed in all three packages independently.

**TYP-04 — The reducer produces a valid program for multi-type functions.**
`reduce.Reduce` never emits a core program whose parameter place carries a TypeID it did not
actually have. *Testable:* a mutation-kill control on `reduce.go`'s parameter-TypeID
recovery; today's fallback silently produces a wrong answer that no test would catch.

**TYP-05 — The six single-function emitters are gone.**
`cgen.Emit` and `cgen.EmitNative` contain no `len(program.Functions) != 1` fork; every
program, single- or multi-function, takes one emission path. *Closes D-11-02 / D-12-36 and
honours D-10-60.* *Testable:* the guard count for `Functions) != 1` in `internal/compiler/cgen`
is zero; every Phase 1-5 generated-C golden is either unchanged or its change is reviewed
and re-frozen.

**TYP-06 — A function may take more than one parameter.**
A developer can declare and call `fn f(a: A, b: B) -> C`. *Testable:* `OpCall` carries N
argument sources at all six dispatch sites with both exhaustive-dispatch controls green;
a 2-parameter program checks, is re-derived by all three non-importing peers, interprets, and
agrees across the five-axis comparator.

**TYP-07 — Argument evaluation and drop order are declared, not emergent.**
The language declares left-to-right argument evaluation and reverse-of-declaration drop for
owned arguments, and the interpreter and every native tier agree. *Testable:* a fixture with
two owned droppable arguments produces an identical release-event sequence on all four
execution axes; an engineered order-reversal mutation is caught.

**TYP-08 — Two arguments at one call site may not conflict.**
Passing the same place in two argument positions is refused unless both positions are shared
borrows, with a named code and a span naming both positions. *Testable:* the refusal fires for
`(owned, owned)`, `(owned, shared)`, `(owned, exclusive)`, `(exclusive, shared)`, and
`(exclusive, exclusive)` on the same place, and does not fire for `(shared, shared)` or for
copyable types; each cell has a fixture.

**TYP-09 — `restrict` on two pointer parameters is justified by a checked disjointness fact.**
When two parameters lower by pointer and both carry `restrict`, the emitted attribute record
names the disjointness fact that justifies it. *Testable:* the existing alias-probe control
(`cgen.go:783-800`) is re-pointed at a genuine two-parameter aliasing subject and kills a
seeded disjointness-rule fault at `-O3 -flto`; the `restrict` claim is proven non-inert.

**TYP-10 (disposition, not build) — DX-06's B1 branch gets an honest home.**
Either (a) `resolveBlame` is wired with at least one production `Violated: true` producer and
a `.lang` fixture that reaches `blameFunction`, or (b) D-13-02b is ratified as permanent with
its reopening condition **corrected** from "`sameType` stops being an admission precondition"
to "a declared contract field exists that the declaring function's own admission cannot fully
verify — i.e. separate compilation." *Testable either way:* `grep -c 'resolveBlame('` over
non-test sources is either ≥2 (wired) or the debt row's reopening condition text has changed.

---

## Confidence + what would change my mind

**Overall: HIGH** on the codebase inventory, **HIGH** on prior art, **MEDIUM** on the
DX-06 reinterpretation.

| Area | Confidence | Basis |
|---|---|---|
| Guard inventory and blast radius | HIGH | Every row grepped and read today; the 32→26 and 89→115 corrections are reproducible with the commands in this document |
| `resolveBlame` is dead code | HIGH | `grep -rn 'resolveBlame(' --include='*.go' internal/ cmd/ \| grep -v _test` → one line, the definition |
| Nothing sets `Violated: true` in production | HIGH | `grep -rn Violated internal/ --include='*.go'` → only `check.go:735/737` (both `false`) and four test lines |
| *Therefore* DX-06 needs separate compilation, not type widening | **MEDIUM** | This is my inference from "every published contract field is producer-derived, and admission is total and callee-before-caller." It is an argument, not a grep |
| Step 1 is one phase; Steps 2-3 are two-plus | MEDIUM | Calibrated against M002 (7 phases, 61 plans for one `OperationKind`) and D-04-22's 8-10-edit standard. `OpCall.SourceIDs` is a schema change of the same family |
| Multi-`restrict` is the highest-value new control | MEDIUM-HIGH | Follows from `cgen.go:920-937` + the dormant alias-probe machinery; not empirically demonstrated |

**What would change my mind on the DX-06 finding — the load-bearing one:** exhibit a
contract field that is (i) written by the developer in source, (ii) published in
`FunctionSignature`, and (iii) **not** fully verifiable at the declaring function's own
admission under whole-program compilation. If such a field exists or can be designed —
for instance a declared cost, effect, or panic-freedom claim checked only at composition —
then B1 is reachable without separate compilation and TYP-10(a) becomes the right branch.
I looked for one in `core.go:289-393` and did not find it; every field is either derived
(`Drops`, `Fresh`, `Callable`, `Abilities`, `ForeignReach`, `ClosureDigest`) or locally
checked (`Return.Paths` at `check.go:3158`, `origin.unknown_path`).

**What would change my mind on sequencing:** a cost-scaling measurement showing
`loanLivenessFixpoint` going superlinear at arity 2-3 on the composition-depth-3 corpus.
That would push Step 3 out of M003 entirely and make Steps 1-2 the whole widening.
LANGUAGE-MATURITY names this exact failure shape (Go 1.18 generics, Polonius) and Phase 08
already established the precedent of hard-gating on spike S-006. **Run that spike before
Phase C is planned.**

---

## Sources

Codebase (read directly, working tree `908bac3`):
`internal/compiler/check/check.go` (:255, :305, :616-860, :874-899, :1059, :1755-1775,
:2044, :2169, :2631-2652, :3148, :3173, :3399, :3406, :3425-3460, :3495-3570, :3616,
:4746-4760, :5047), `internal/compiler/core/core.go` (:127-150, :232-244, :289-393,
:424-449, :540-575, :614-625, :897-912), `internal/compiler/corevalidate/corevalidate.go`
(:2249, :2260-2265, :2310-2361, :2540-2570), `internal/compiler/originvalidate/originvalidate.go`
(:246, :334, :441, :639, :917-940, :1091), `internal/compiler/cgen/cgen.go` (:15-60, :118,
:148, :265, :575-620, :730-800, :915-960, :2570-2579), `internal/compiler/cgen/cgen_program.go`
(:128-190, :239, :363), `internal/compiler/interp/interp.go` (:205, :219, :286, :505-560,
:722-729, :767), `internal/compiler/pathoracle/pathoracle_compose.go` (:40-110),
`internal/compiler/reduce/reduce.go` (:485, :755-785, :851),
`internal/compiler/ast/ast.go` (:69-99), `internal/compiler/syntax/parser.go` (:300-352).

Planning: `.planning/PROJECT.md`, `.planning/LANGUAGE-MATURITY.md`,
`.planning/STANDING-VERDICTS.md`, `.planning/milestones/M002-MILESTONE-AUDIT.md`,
`.planning/milestones/M002-phases/13-agent-loop-for-interprocedural-defects/PHASE-13-DEBT.md`.

External (all HIGH confidence — official docs, language specs, RFCs, or peer-reviewed papers,
except where noted):

- [Mojo — Ownership (official manual)](https://mojolang.org/docs/manual/values/ownership)
- [Modular — Deep dive into ownership in Mojo](https://www.modular.com/blog/deep-dive-into-ownership-in-mojo) (vendor blog, MEDIUM)
- [Hylo Language Specification](https://hylo-lang.org/docs/reference/specification/)
- [Hylo — Val for Swift users](https://github.com/hylo-lang/Documentation/blob/main/val-for-swift-users.md)
- [Swift.org — Swift 5 Exclusivity Enforcement](https://www.swift.org/blog/swift-5-exclusivity/)
- [SE-0176 — Enforce Exclusive Access to Memory](https://github.com/apple/swift-evolution/blob/master/proposals/0176-enforce-exclusive-access-to-memory.md)
- [SE-0108 — Remove associated type inference](https://github.com/apple/swift-evolution/blob/master/proposals/0108-remove-assoctype-inference.md)
- [Rust RFC 2025 — Nested method calls (two-phase borrows)](https://rust-lang.github.io/rfcs/2025-nested-method-calls.html)
- [Rust RFC 2094 — Non-Lexical Lifetimes](https://rust-lang.github.io/rfcs/2094-nll.html)
- [Rust Blog — Enabling Polonius Alpha on nightly (2026-08-04)](https://blog.rust-lang.org/2026/08/04/enabling-polonius-alpha-on-nightly/)
- [The Rust Reference — Destructors (drop order)](https://doc.rust-lang.org/reference/destructors.html)
- [The Austral Language Specification](https://austral-lang.org/spec/spec.html)
- [Borretti — How Austral's Linear Type Checker Works](https://borretti.me/article/how-australs-linear-type-checker-works)
- [Perceus: Garbage Free Reference Counting with Reuse (MSR TR)](https://www.microsoft.com/en-us/research/uploads/prod/2020/11/perceus-tr-v1.pdf)
- [Lorenzen — Optimizing Reference Counting with Borrowing (MSc thesis)](https://antonlorenzen.de/master_thesis_perceus_borrowing.pdf)
- [Hooper — Why Swift's Type Checker Is So Slow](https://danielchasehooper.com/posts/why-swift-is-slow/) (blog, MEDIUM)
- [Type Inference in Rust and C++](https://herecomesthemoon.net/2025/01/type-inference-in-rust-and-cpp/) (blog, MEDIUM)
