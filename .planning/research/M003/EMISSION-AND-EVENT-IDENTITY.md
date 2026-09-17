# M003 Research: Emitter Retirement and Event Identity

**Researched:** 2026-09-17
**Scope:** Cluster A (D-11-02 → D-12-36, the six single-function emitters) and
Cluster B (D-11-51 → D-12-21, event identity), plus D-10-60's disposition.
**Method:** every codebase claim below was read at the cited path and line, and
the diamond fixture was actually executed. External claims are cited in
`## Sources`.
**Overall confidence:** HIGH on the grounded findings, MEDIUM-HIGH on effort,
MEDIUM on the C-side cost of the recommended identity scheme.

---

## Verdict (one-shot recommendation, one paragraph per cluster)

**Cluster A — split it, land the half M003 needs anyway, and formally cut the
other half.** The framing "delete six emitters" is wrong in a way that has cost
two deferrals: `emitProgram` is not a superset of the legacy family, it is a
strict *subset* — it refuses `core.Match` bodies, block-based (foreign) bodies
and `ForeignContract` outright (`cgen_program.go:143-155`), emits no `restrict`
anywhere, and hardcodes `"live_resources":[]` (`cgen_program.go:299`). So this
is a **port-then-delete**, not a delete. But the three capabilities split
cleanly by whether M003's *other* scope needs them. Branch/match bodies are
needed: `match` is the only control flow the language has
(`LANGUAGE-MATURITY.md:57-62`), so every multi-function fixture with a branch,
a `defect`, or a typed failure is currently unlowerable (D-11-52), and both
"widen return type ≠ parameter type" and "real control flow" produce exactly
those fixtures. By-pointer lowering and foreign bodies are *not* needed by any
named M003 goal. Recommendation: in M003 port branch/match into `emitProgram`
and delete **three** of the six (`emitMatch`, `emitBranch`, `emitLinear`);
explicitly cut `emitLinearForeign`, `emitLinearBorrowedByPointer`, and
`emitLinearBorrowedByPointerPlain` from the D-11-02 item via a REQUIREMENTS.md
amendment and re-file them as two differently-scoped items with named M004
landings — which is the disposition D-10-60 itself already authorizes, and the
D-10-30 "withdrawn premise + refiled row" precedent already blessed. Reuse
Phase 09's build-then-delete gate shape verbatim (see below); expect ~900 lines
of emitter body plus ~400 of shared helper surface in scope, not the full
~1,500 D-12-36 quotes for all six.

**Cluster B — own it, and own it *first*, because it gates Cluster A's own
evidence.** An "event" here is an `execution.Event`
(`internal/compiler/execution/execution.go:41-51`) whose `ID` is derived from
the emitting operation's **static** ID — `operation.ID + ":event:returned"` in
`interp.terminalOutcome` (`interp.go:623-641`) and the byte-identical string in
`cgen.emitProgramFunction` (`cgen_program.go:~400`). That is a per-*declaration*
identity. A shared-leaf diamond re-invokes one callee, so the identity repeats;
I ran it and confirmed two identical
`s1:…:fn:leaf:op:0:event:returned` lines in one document. The correct primitive
is **occurrence identity, not content identity** — and this is the single most
important design finding in this report: a content-addressed digest over the
event's semantic payload is *guaranteed* to collide here, because the diamond's
two `leaf` activations are content-identical by construction. Merkle/content
addressing identifies values; distributed tracing identifies occurrences, and
that is the right analogy. Recommendation: a **context-path invocation
identity** (a call-site chain, W3C-Trace-Context-shaped), carried as a new
required `invocation` field under a `lang.execution/2` schema bump, with
uniqueness moved from `ID` to the pair `(invocation, id)`. Do it before the
emitter port, because until it lands **no fixture that re-invokes a callee can
be four-tier compared at all**, which caps the corpus that would authorize the
deletion.

---

## Cluster A: the six emitters, grounded

All six live in `internal/compiler/cgen/cgen.go` (2,658 lines). All six are
unexported; their only production callers are `Emit` (`cgen.go:112-137`) and
`EmitNative` (`cgen.go:142-171`), which dispatch identically except for one
`executionJSON` flag.

| # | Identifier | Lines | Body LOC | What it emits | Why it still exists |
|---|---|---|---|---|---|
| 1 | `emitMatch` | `cgen.go:235-316` | 81 | An `enum`-typed `switch` function + `main`, and — uniquely — a **`lang.execution/0`** document with `Input`/`Output` event fields (`cgen.go:302-310`, `execution.Schema0`) | `emitProgram` refuses any `function.Match != nil` (`cgen_program.go:143-145`). Also the sole producer of schema `/0`; retiring it retires a whole schema on the producer side |
| 2 | `emitLinear` | `cgen.go:317-492` | 175 | One inline `main` containing the whole straight-line body plus the `lang.execution/1` JSON tail | The only shape `emitProgram` *can* already express — and even here the two outputs differ (see the N=1 table) |
| 3 | `emitLinearBorrowedByPointer` | `cgen.go:615-748` | 133 | By-pointer parameter lowering **with the `restrict` qualifier** (`byPointerQualifier`, `cgen.go:588-614`; `emittedAttributeForByPointerParameter`, `cgen.go:931-952`) | `emitProgram` emits zero `restrict` — the only occurrence of the token in `cgen_program.go` is a comment at line 88 saying it never mints one |
| 4 | `emitLinearBorrowedByPointerPlain` | `cgen.go:809-930` | 121 | The same by-pointer lowering **without** `restrict` (shared-loan-only case, `selectsByPointerLoweringSharedOnly`, `cgen.go:749-808`) | Same as #3; it is the negative control that makes #3's `restrict` claim non-vacuous |
| 5 | `emitLinearForeign` | `cgen.go:953-1147` | 194 | Block-based bodies: foreign calls, the resource ledger (`cgen.go:1222-1315`), leak events, `emitNonlocalPad` (`cgen.go:1148-1192`), and a real non-empty `live_resources` array | `emitProgram` refuses `len(function.Linear.Blocks) > 0` **and** `function.ForeignContract != nil` (`cgen_program.go:149-155`), and hardcodes `"live_resources":[]` (`cgen_program.go:299`) |
| 6 | `emitBranch` | `cgen.go:1757-1916` | 159 | Match-armed functions with linear bodies, `defect` terminators (`emitDefectSupport`, `cgen.go:1739`), and payload construct/destructure arms via `emitBranchOperations` (`cgen.go:1979-2236`, 258 lines) | Same `core.Match` refusal as #1 |

Six bodies = **863 lines**. Add the helper surface only these six reach —
`emitBranchOperations` (258), `resourceLedger` + `emitForeignReleases*` +
`foreignFailureLiteral` (~250), `emitLinearOutputSupport` /
`emitLinearForeignOutputSupport` (~65), the two `selectsByPointerLowering*`
predicates + `byPointerQualifier` + `emittedAttributeForByPointerParameter`
(~130) — and D-12-36's "roughly 1,500 lines of `cgen.go` emitter logic" is
accurate for the *whole* family.

### Is the multi-function path a genuine superset? No — it is measurably a subset.

`TestN1ConvergenceDifferential` (`cgen/cgen_n1_convergence_test.go`, committed
`1bac938`) is already the gate, already RED, and already pins the exact state:

| Fixture | legacy `Emit` | `emitProgram` | identical |
|---|---|---|---|
| `testdata/phase1/toggle.lang` (match-only) | ok | **refuses** | — |
| `testdata/phase2/owned_transfer.lang` (plain linear) | ok | ok | **no** |
| `testdata/phase3/borrowed_view.lang` (branch+linear) | ok | **refuses** | — |
| `testdata/phase4/foreign_acquire_one.lang` (foreign blocks) | ok | **refuses** | — |
| `testdata/phase4/defect_terminal.lang` (defect terminator) | ok | **refuses** | — |

The one shape both handle diverges two ways: `emitProgram` writes a `static`
function plus a separate `main` where `emitLinear` writes one inline `main`, and
its preamble carries an extra `#include <stdlib.h>` plus
`emitCallBoundaryAttributeComment` (`cgen_program.go:44-53`).

### What concretely breaks on deletion today

1. **Every match-bodied program.** `toggle.lang`, `borrowed_view.lang`,
   `defect_terminal.lang` and the whole Phase 12 payload corpus route through
   `emitMatch`/`emitBranch`/`emitBranchOperations`.
2. **Every foreign program.** The resource ledger, leak events, the process-root
   nonlocal landing pad, and non-empty `live_resources` exist only in
   `emitLinearForeign`. `emitProgram` would emit a document asserting no live
   resources for a program that has them — a *silent wrong answer*, not a
   refusal.
3. **The `restrict` evidence family.** `emitLinearBorrowedByPointer` is the
   only producer of the token. Deleting it without porting deletes the D-05-01
   `restrict` claim, and with it the meaning of `BannedOptimizerAttributes` /
   `JustifiableAttributes` (`cgen.go:2509`, `:2550`) and
   `ScanForUnjustifiedAttributes`.
4. **A three-way cross-layer agreement.** `selectsByPointerLowering` is
   exported as `SelectsByPointerLowering` and independently re-derived twice:
   by `check` (`check_exclusive_test.go:233-249`,
   `TestAliasFactAgreesWithByPointerSelection`) and by `session`
   (`session_phase11_gate.go:50-90`, `phase11WouldCarryRestrict`). This
   predicate is a *contract*, not an emitter detail; it must survive any
   deletion of its consumer.
5. **Four pinned golden-C digests.** `previousPhaseGoldenCDigests`
   (`core/core_test.go:155-160`) pins `phase1/generated.golden.c` (emitMatch),
   `phase2/owned_transfer.golden.c` (emitLinear),
   `phase4/foreign_layout_mismatch.golden.c`, `phase5/restrict_borrow.golden.c`
   (by-pointer). `TestPreviousPhaseGoldenCUnchanged` fails on any drift *and*
   on a corpus gaining or losing a file.
6. **Schema `lang.execution/0` on the producer side.** Only `emitMatch` emits
   it. `native.validateExecution` still has a live `Schema0` branch
   (`native.go:553`, `:610-613`). Deleting `emitMatch` makes that branch
   decoder-only; that should be a deliberate, recorded narrowing.

### Is this a delete, a merge, or a rewrite?

**A capability port followed by a delete**, in three separable families:

| Family | Emitters | Work | Needed by M003's other scope? |
|---|---|---|---|
| Branch/match bodies | `emitMatch`, `emitBranch` (+ `emitBranchOperations`) | Port arm lowering + defect support into `emitProgramFunction`; converge on `lang.execution/1`; re-pin `phase1/generated.golden.c` | **Yes.** `match` is the only control flow; D-11-52 says a diverging callee is unlowerable today purely because of this |
| Plain linear | `emitLinear` | Converge preamble + `static`-fn-plus-`main` shape at N=1; re-pin `phase2/owned_transfer.golden.c` | **Yes, as a consequence** — once branch is ported, keeping a second straight-line writer is pure duplication |
| By-pointer + foreign | `emitLinearBorrowedByPointer`, `…Plain`, `emitLinearForeign` | Interprocedural by-pointer lowering requires caller-side address passing, which opens D-11-11/D-11-12 (the `EmittedAttribute` discharge-pair design, designed-not-built). Foreign bodies require a real multi-function resource ledger and multi-frame nonlocal pad (D-10-34) | **No.** No named M003 goal needs either |

**Effort estimate (real):**

- Branch/match port + `emitLinear` retirement + golden re-pin + flipping
  `TestN1ConvergenceDifferential` to byte-identity on three of five shapes:
  **~600-800 lines touched, 4-6 plans**, dominated by `emitBranchOperations`'s
  payload arms and by making the preamble converge without perturbing the two
  non-affected goldens.
- By-pointer port (with discharge pairs): **a phase of its own**, because it is
  not a port — D-11-11's caller-side `discharged_by` mechanism has never been
  built and NAT-05's attribute set is empty by construction today.
- Foreign body port: **a phase of its own**, blocked behind D-10-34's
  multi-frame nonlocal pad and a multi-function resource ledger.

### The reusable "build-then-delete boundary gate" — Phase 09's shape, found

Phase 09 deleted `computeLoanLastUses` (the second, older liveness decision
point) and the pattern is documented and reusable verbatim:

1. **Three pre-deletion gates land in a *prior* plan (09-07):** an
   ordering-stability baseline, a bounded enumeration answering the one
   behavioural question the deletion could change (D-09-49 Q1), and a
   shadow-path reachability/refusal corpus. Artifacts:
   `check/check_ordering_stability_test.go`,
   `check/check_shadow_subsumption_test.go` (`TestShadowPathSubsumptionCorpus`).
2. **A mandatory mid-phase `checkpoint:decision` plan (09-08) whose only job is
   to authorize or refuse the deletion**, rated `reversibility="one-way"`, with
   a fixed evidence agenda: per-shape zero-divergence results at every swept
   size, both seeded-fault directions with companion assertions, the cost curve
   **re-measured at the gate's own commit**, and an explicit
   "authorized / not authorized, and if not, exactly what must land first."
   (`09-08-PLAN.md:113-175`.)
3. **Deletion and cutover in the SAME commit (09-09, D-09-10):** "never an
   intermediate state where neither path decides." The old function was replaced
   by a renamed evidence-only successor (`loanFinalUseEvidence`) so the
   evidence fields stayed populated while the *decision* moved.
4. **The differential test is never deleted, only its expectations flip** —
   `TestN1ConvergenceDifferential`'s own doc comment already commits to this.

A second reusable artifact worth copying: Phase 11's **`11-GUARD-LEDGER.md`** —
every `len(Functions) != 1` site enumerated with a `WIDENED`/`KEPT` disposition
and a one-line reason, on the principle that "a guard nobody decided about is an
undecided hole." Re-run today the inventory is **26 non-test matches** (was 32
in `LANGUAGE-MATURITY.md:97-103`), eight of them inside
`session.verifyBorrowedCorpus` alone, and `reduce.Reduce`/`ProjectSource` are
**already widened** — `LANGUAGE-MATURITY.md:109-111` is stale on that point and
should be corrected.

---

## Cluster A: delete / merge / rewrite — options table and recommendation

| Option | What it is | Cost | Risk | Verdict |
|---|---|---|---|---|
| **A0. Third deferral, silent** | Rewrite the landing-phase cell again | 0 | Violates D-10-60 as written | **Reject** |
| **A1. Retire D-10-60, keep all six** | Declare the rule wrong; carry the emitters indefinitely | 0 now | Pays the two-families tax on *every* M003 change; `emitProgram` hardcodes one type per function (`cgen_program.go:~230`, `:~330`), so the return≠parameter widening must be implemented **twice** | **Reject** — see the D-10-60 section |
| **A2. Full port of all six, then delete** | D-12-36's literal reading | ~1,500 lines, 3 phases incl. discharge pairs and multi-frame pads | Consumes most of M003; delivers zero user-visible surface; two of the three families have no M003 consumer | **Reject** |
| **A3. Split: port branch/match + retire 3; formally cut 3** | Amend REQUIREMENTS.md, re-file the cut half as two new rows with M004 landings | ~600-800 lines, 4-6 plans | The by-pointer/foreign families keep living alongside `emitProgram` — but behind a *narrow, predicate-defined* dispatch seam (`selectsByPointerLowering*` / `len(Blocks)>0`), which is Branch-by-Abstraction's correct intermediate state, not "two families racing" | **RECOMMEND** |
| **A4. Rewrite both families into a table-driven lowering DSL** | Cranelift-ISLE-style single declarative lowering | Very large | Right long-term shape; wildly premature at 5-10% language surface | **Reject now, name as the M005+ destination** |

**Recommendation: A3.** Plus one explicit constraint that prevents the
two-families tax from compounding: *no surviving legacy emitter may be extended
to support return type ≠ parameter type.* A program that needs both by-pointer
lowering and a differing return type is **refused with a named code** until the
M004 port lands. That converts an open-ended maintenance liability into a
declared, testable language restriction — exactly the shape the project already
used for `check.duplicate_payload_type` (D-12-44).

---

## On D-10-60: keep the forcing function, or retire the rule?

**Read the rule literally first.** D-10-60
(`PHASE-10-DEBT.md` items table, row D-10-60) does **not** say "never defer
twice." It says an item that would be deferred a second time "must either be cut
from the milestone explicitly via a REQUIREMENTS.md amendment, or becomes
automatically never-cut. It cannot **silently** acquire a third landing phase."

So the framing in `M002-MILESTONE-AUDIT.md` §4 — "M003 should either land the
deletion or retire D-10-60" — presents a false binary. **The rule's own text
authorizes a third disposition: an explicit, formal scope cut.** And D-10-30
already ratified the adjacent move: a withdrawn premise plus a freshly filed,
differently-scoped row "is structurally different from a `Landing phase` cell
rewritten to point past the phase it already named, which is what D-10-60
forbids."

**Is the rule producing sunk-cost commitment?** Partly, and the partly is
diagnostic. The item D-10-60 is forcing is *badly specified*: "delete six
emitters" names a mechanism, not an outcome, and bundles three capability
families with wildly different value. A forcing function attached to a
badly-scoped item will reliably produce either a bad "pay it" or a
rule-retirement — which is exactly the two-horned choice the audit reached. The
defect is in the item, not the rule.

**Legitimate grounds to retire the RULE (stated honestly, then weighed):**

- It is enforced by review, not by mechanism — the row says so itself, and names
  `TestDebtRegistersAreWellFormed` (`session/session_test.go:2613`) as checking
  format only. A rule nobody can mechanically violate-detect is weak.
- It creates an incentive to file *new* rows rather than defer old ones, which
  launders deferral into growth of the register. (77 items, 26 closed, is
  consistent with that pressure.)
- It has no external stakeholder. It is a self-imposed discipline on a
  pre-alpha language.

None of these is a reason to retire it. All three are reasons to *strengthen*
it.

### Recommendation: **KEEP D-10-60, discharge Cluster A under its own amendment clause, and mechanize it.**

Named recommendation, three parts:

1. **Do not retire D-10-60.** Retiring a forcing function on the first occasion
   it actually fires is how you reacquire the pattern it was written to prevent.
   Its record is good: it is the only reason this decision is being made
   deliberately rather than drifting into M004.
2. **Discharge Cluster A via the amendment clause, not via deferral.** Land
   branch/match + `emitLinear`. Withdraw the "six emitters" item as
   mis-specified (D-10-30 precedent), and file two new rows —
   *multi-function by-pointer lowering with attribute discharge pairs* (depends
   on D-11-11/D-11-12) and *multi-function foreign bodies and resource ledger*
   (depends on D-10-34) — each with an M004 landing and its own success
   condition. Record the withdrawal verbatim; never erase the superseded text
   (D-09-45 discipline).
3. **Mechanize the hop count.** Extend the debt-register format with a required
   `superseded_landing_phases:` list per row and teach
   `TestDebtRegistersAreWellFormed` to fail when that list has length ≥ 1 unless
   the row also carries an `amendment:` field naming a REQUIREMENTS.md section.
   That turns D-10-60 from a review convention into a CI gate — the same
   "a debt note is read once, a test runs on every CI invocation" reasoning the
   project already ratified for Phase 11's two human-judgment items.

---

## Cluster B: what an event is today, and why diamonds collide (grounded)

**An event is `execution.Event`** (`execution/execution.go:41-51`): a flat
struct of `Schema, ID, Kind, FunctionID, Input, Output, SourcePlace,
TargetPlace, TypeID`. There is no notion of a frame, an activation, a parent, or
a time. `Execution` (`:53-57`) carries `Schema`, `Outcome`, an **ordered**
`Events []Event`, and `LiveResources []string`.

**Event IDs are pure functions of static operation IDs:**

| Producer | Site | Derivation |
|---|---|---|
| interp, terminators | `interp.go:623-641` (`terminalOutcome`) | `operation.ID + ":event:returned"` / `":event:failed"` / `":event:defected"` |
| interp, transitions | `interp.go:846-851` (`ownedEvent`) | `operation.ID + ":event"` |
| interp, depth refusal | `interp.go:836` | `operation.ID + ":event:call_depth_exceeded"` |
| cgen multi-function | `cgen_program.go` `emitProgramFunction` | `operation.ID+":event"` and `operation.ID+":event:returned"` — byte-identical strings, independently written |
| cgen single-function match | `cgen.go:302-310` | `arm.ID + ":event:returned"`, schema `/0` |

**Why diamonds collide.** These are *per-declaration* identities. A callee
invoked from two distinct static call sites executes its own `OpReturn` twice,
producing the same string twice. Verified by execution, not by reading:

```
$ go run ./cmd/lang run --engine=interpreter testdata/phase11/multi_function_diamond_call.lang
s1:phase11.multi_function_diamond_call:fn:leaf:op:0:event:returned  function.returned …
s1:phase11.multi_function_diamond_call:fn:left:op:1:event:returned  function.returned …
s1:phase11.multi_function_diamond_call:fn:leaf:op:0:event:returned  function.returned …   <-- collision
s1:phase11.multi_function_diamond_call:fn:right:op:1:event:returned function.returned …
s1:phase11.multi_function_diamond_call:fn:main:op:2:event:returned  function.returned …
```

`interp.Run` performs no duplicate check and returns this happily.
`native.validateExecution` (`native.go:586-589`) refuses it:
`errors.New("duplicate execution event id")`. So three of four tiers produce no
comparable document at all, and `session.Phase5CompareEngines`'s
`AxisEventOrder` (`session_phase5_compare.go:103-115`) — a positional,
field-by-field comparison of the whole `Event` struct — never runs.

**Three consequences the debt rows do not state:**

1. **The collision is not confined to `function.returned`.** `ownedEvent`
   (`interp.go:846`) uses the same `operation.ID+":event"` derivation, so *every*
   transition event of a re-invoked callee collides too. Today `leaf` is a
   one-operation function so only the return event shows; a re-invoked callee
   with copies/moves/borrows collides on all of them.
2. **`OpCall` emits no event at all.** Confirmed in the run above (five events,
   all `function.returned`) and in the code — `core.go:559-564` documents
   `OpCall` as never a terminator, and neither engine records anything at the
   call site. So the caller→callee causal edge is **unobservable in the
   document**. Any identity scheme that needs a parent link must either derive
   it or introduce the event.
3. **This is a soundness hole in the *evidence*, and the audit is right to call
   it a warning.** Events are one of the comparator's five axes. Two
   indistinguishable events in a document mean the four-tier claim's event axis
   is not well-defined for the exact class of program M002 was built to prove
   things about.

---

## Cluster B: identity scheme options (table) and the recommended design, specified

| # | Scheme | Deterministic? | Same across interp/-O0/-O3/-flto? | Peer re-derivable without importing an engine? | Non-self-referential? | Verdict |
|---|---|---|---|---|---|---|
| 0 | Status quo: static operation ID | yes | yes | yes | yes | **Broken** — not unique per occurrence |
| 1 | Global monotonic counter over emission order | yes | yes, *if* both engines emit in identical order (they do today) | **No** — requires simulating execution; a peer cannot derive it from the program | yes | Reject as the identity. Fragile: inserting any earlier event renumbers every later one, so goldens, reducer signatures and cause-DAG references all churn on unrelated edits. Order is already an axis; do not overload it as identity |
| 2 | Content digest over the event's semantic payload | yes | yes | yes | only if the ID is excluded from the preimage — the exact ClosureDigest trap | **Categorically wrong.** The diamond's two `leaf` activations are *content-identical* by construction, so a content hash collides **by design**. Content addressing identifies values; this problem needs occurrence identity |
| 3 | Path/context-sensitive: call-site chain | yes | yes | **yes** — the call graph is a guaranteed DAG (`callgraph.Order` refuses cycles, TRU-04/D-09-22), so the set of admissible paths is statically enumerable | yes — inputs are frontend operation IDs only | **Strong** |
| 4 | Hybrid: call-site chain + per-context occurrence ordinal | yes | yes | yes (path statically; ordinal bounded by loop semantics) | yes | **RECOMMEND** — #3 plus forward-compatibility with the loops M003 wants |
| 5 | Digest *of* the path (fixed-width hash of #4's preimage) | yes | yes | yes | yes, with a domain separator | Fallback only. Loses readability, adds a collision argument for no benefit at bounded depth. Keep as the escape hatch if paths get long |

### Recommended design — specified

**Concept: `Invocation`.** A canonical, textual, human-auditable identifier for
one *activation* of one function in one run. Directly analogous to a W3C Trace
Context span: the parent link is in the identifier, and the identifier is
deterministic rather than random because this language's execution is
deterministic.

**Grammar (normative):**

```
invocation ::= "inv:" root ( "/" step )*
root       ::= "entry:" <entry function ID>
step       ::= <call-site OpCall operation ID> "#" <ordinal>
ordinal    ::= 0-based decimal, the count of PRIOR activations of that same
               call-site operation within the SAME parent invocation
```

The program-entry frame's invocation is `"inv:entry:" + entry.ID`, where `entry`
is `callgraph.EntryFunction`'s resolved function (`callgraph.go:279`). On
executing `OpCall` operation `O` in parent invocation `P`, the child invocation
is `P + "/" + O.ID + "#" + itoa(k)`. Today, with no loops and no recursion, `k`
is always `0`; the field is in the grammar **now** so that iteration does not
force a second schema change later.

For the diamond, `leaf`'s two activations become
`inv:entry:…fn:main/…fn:main:op:0#0/…fn:left:op:0#0` and
`inv:entry:…fn:main/…fn:main:op:1#0/…fn:right:op:0#0`. Distinct, and each one
*states its own causal chain*.

**Wire change.** Add to `execution.Event`:

```go
Invocation string `json:"invocation"`   // required, non-empty, schema >= /2
```

Bump to `lang.execution/2`. Freeze the `/1` decoder byte-for-byte the way `/0`
was frozen at the `/0`→`/1` bump (`MILESTONES.md:49` records that bump landing
across 12 literal sites with `/0` bytes provably frozen) — reuse that exact
playbook, including the schema-peek dispatch.

**Uniqueness invariant moves, `ID` does not change.** `validateExecution`
(`native.go:574-590`) keeps its duplicate check but keys it on the pair
`(Invocation, ID)` for `/2` documents, and on `ID` alone for `/0` and `/1`.
Crucially, **`Event.ID`'s existing derivation is untouched** — so every
diagnostic, reducer signature, cause-DAG reference and golden that names a
static event ID keeps working. This is the smallest change that fixes the
invariant, and it follows Phase 09's own trick of splitting a decision out of an
existing artifact rather than mutating it.

**Comparator.** Add `"Execution.Events.Invocation"` to the routed field list
(`session_phase5_compare.go:179-187`). The fail-closed field-routing test forces
this to be a deliberate classification rather than a silent omission.

**Interpreter implementation** (`interp.go`): `frame` gains
`invocation string` and `callOrdinals map[string]int`. `interp.Run` sets the
base frame's invocation. `runFrameStack`'s `core.OpCall` arm
(`interp.go:~810-824`) computes `k = top.callOrdinals[operation.ID]++` and sets
the pushed frame's invocation. `terminalOutcome` and `ownedEvent` take the
frame's invocation as a parameter. ~60 lines.

**Native implementation** (`cgen_program.go`) — the part with real design
choice. Three sub-options:

- **(a) Runtime string concatenation.** Thread `const char *lang_inv` as an
  extra leading parameter; at each call site `snprintf` the child path into a
  stack buffer. Correct, but adds an unbounded-looking buffer to every frame and
  puts `snprintf` on the hot path.
- **(b) Static path table (RECOMMENDED).** The emitter statically unfolds the
  call DAG into a *call tree*: one node per (parent node, call-site operation)
  pair. Emit
  `static const char *const lang_inv_paths[N] = { … };` holding each node's
  canonical path string, and thread `unsigned lang_inv` (the node index) as an
  extra leading parameter. Each call site passes a **compile-time constant**
  index, so `-O3` can constant-propagate, clone, and inline freely without
  changing the recorded strings. `lang_record_event` gains an
  `lang_inv_paths[lang_inv]` argument. Fail closed: if `N` exceeds a declared
  ceiling (propose 4096 — a diamond chain of depth `d` unfolds to `2^d` nodes),
  refuse emission with a named error, in the same idiom as `pathoracle`'s
  `MaxPaths` and `interp`'s `maxCallDepth` refusal.
- **(c) Fixed-width rolling hash** of the path, computed identically in Go and
  C. Cheapest in C, unreadable in the document, and reintroduces a collision
  argument for nothing. Fallback only.

Choose **(b)**. It keeps the emitted C readable (a named table of literal path
strings is *more* auditable than what exists today), keeps the value a
compile-time constant so the optimizer is unconstrained, and costs no runtime
string work.

**Scope note, worth deciding explicitly:** consider also emitting a
`function.called` event at each `OpCall` carrying the *child's* invocation. It
is not required for uniqueness, but it makes the call edge observable, gives the
comparator a genuine call-order signal, and is what would let a peer reconstruct
the call tree from the document alone. Flag it as an in-or-out decision at the
phase's discuss gate rather than smuggling it in.

---

## Re-derivability by a non-importing peer (how the peer derives the same identity independently)

This is the property that makes scheme #4 clearly better than #1, and it rests
on a fact the project already proved: **the call graph is a DAG.**
`callgraph.Order` (`callgraph.go:460`) runs a three-color DFS and refuses every
cycle before any downstream derivation (TRU-04, D-09-22; `PROJECT.md` Out of
Scope: "Recursive programs").

A peer holding only the checked `core.Program` — importing neither `interp` nor
`cgen`, satisfying the same `go list -deps` guard the other three peers pass —
can:

1. Resolve the entry with its own copy of the in-degree-zero + closure-size rule
   (`callgraph.EntryFunction`; `callgraph` is already permitted to
   `originvalidate` and forbidden to `corevalidate`, D-10-C03 — so the peer's
   import list must be decided explicitly, not inherited).
2. **Unfold the DAG into the call tree** by walking `OpCall` operations in
   declaration order, generating exactly the string set the emitter's
   `lang_inv_paths` table contains, *without running anything*. This is a pure
   syntactic derivation from operation IDs.
3. Check every event in a decoded document against that set:
   - `event.Invocation ∈ AdmissiblePaths(program)` — membership;
   - every `(invocation, id)` pair occurs at most once — uniqueness;
   - the `event.FunctionID` implied by the last step of `event.Invocation`
     equals the event's own `FunctionID` — **cross-field consistency**, which is
     what catches an engine that computes the path correctly but attributes the
     event to the wrong frame;
   - the sequence of invocations is a valid pre-order traversal of the call tree
     (a parent's path is a prefix of any path appearing between its first and
     last event) — **structural ordering**, independent of the comparator.

Checks 3c and 3d are genuinely new evidence, not a re-run of the interpreter.
A monotonic counter (#1) supports none of them; a content digest (#2) supports
none of them and is wrong besides.

**Not self-referential — and here is the lesson learned.** The project already
hit this exact trap once. `07-RESEARCH.md:366` records the superseded
recommendation `ClosureDigest = hash(own FunctionSignature bytes)` with the
correction: "**This recommendation is self-referential and must not be
implemented:** the 'own `FunctionSignature` bytes' include the `ClosureDigest`
field itself." The cross-AI review flagged it HIGH; the fix (D-07-37) was a
canonical preimage — the signature with `ClosureDigest` **zeroed**, a domain
separator, and callee pairs **sorted by ID**
(`originvalidate.go:555-600`, `core.go:320-337`), with
`originvalidate_internal_test.go:36` asserting non-self-reference directly and
`closureDigestSortOverride` as a mutation-kill seam for the sort.

The invocation scheme is structurally immune: its inputs are the entry function
ID and a sequence of `OpCall` operation IDs, all assigned by the frontend before
any event exists. No event field, no digest, and no prior invocation value
derived from event content ever enters the derivation. **State this property in
the design doc explicitly and pin it with a test in the
`originvalidate_internal_test.go:36` shape** — mutate an event's `ID` and assert
its `Invocation` does not move — rather than leaving it as an argument.

---

## Sequencing against the rest of M003

Two interactions dominate, and both were verified in code.

**1. `emitProgram` bakes in one type per function.** `emitProgram` computes a
single `typeNames[index]` from `linearInput(function)` and uses it for the
prototype's return type *and* parameter type
(`static %s %s(%s);`), for every local declaration in `emitProgramFunction`, and
for the call target's local in `emitCall`. So the return≠parameter widening
lands in exactly the same code as the emitter port. **If the six emitters
survive the widening, the widening is implemented twice** — in `emitProgram` and
in the four surviving linear/branch emitters, each of which derives its type the
same single-valued way. That is the concrete, priceable cost of A1
("retire the rule, keep the code"), and it is the strongest argument for A3.

**2. Event identity gates Cluster A's own evidence.** Until identity is fixed,
any fixture that re-invokes a callee produces no comparable native document
(`native.go:588` refuses it), so it cannot enter
`TestPhase11InterproceduralDifferential`'s `AllComparableFixtures` sweep. The
corpus available to authorize an emitter deletion is therefore *restricted to
call graphs that never reuse a callee* — which excludes essentially every
realistic program shape, and certainly every shape the port is supposed to make
possible.

**Recommended order:**

| # | Work | Why here |
|---|---|---|
| 1 | **Event identity** (Cluster B): `lang.execution/2`, `invocation`, `(invocation,id)` uniqueness, interp + `emitProgram`, peer re-derivation checks, comparator field routing | Small (~300-400 lines), closes D-11-51 **and** D-12-21, and unblocks the diamond/shared-leaf corpus that everything after it needs for evidence |
| 2 | **Branch/match port + retire 3 emitters** (Cluster A), under Phase 09's three-gate + blocking-checkpoint + same-commit pattern | Needs step 1's corpus. Must precede type widening so the widening is paid once |
| 3 | **Type widening** (return ≠ parameter), which auto-closes DX-06/DX-07 (D-13-02b, D-13-10a) | Lands in one lowering family instead of five |
| 4 | **Control flow + arithmetic** | Needs 2 and 3; also the first user of the `#ordinal` slot reserved in step 1's grammar |
| 5 | Nyquist debt (07, 08, 11, 12, 13) and D-13-34 | Cheap, independent; do not Nyquist-validate 11/12 *before* step 2 touches `cgen` — STANDING-VERDICTS' own "don't Nyquist-validate a surface a milestone is about to change" applies directly |

**Does the emitter port get harder or easier after type widening?** Harder, and
measurably so: widening first means the port must carry the two-type model
through ~900 lines of legacy emitter while simultaneously converging them —
two changes at once in the same lines, which is the fingerprint the project
already identified as costing M001 Phases 2-4 extra remediation rounds.

**Does event identity get harder after the port?** Yes, mildly: the ported
branch/match arms would emit events under the *old* scheme and then be rewritten
— which is literally the "two coexisting laws is a defect with a delayed fuse"
anti-pattern in STANDING-VERDICTS. Identity first means the ported arms are born
with the correct scheme.

---

## The unnamed third risk

**`-flto` is structurally inert on every cgen-emitted multi-function program,
and nothing carries that as debt.**

The chain, all verified:

- STANDING-VERDICTS declares `-flto` **load-bearing** for the equivalence claim:
  "`restrict` and TBAA promises are per-translation-unit; without LTO, Clang
  cannot see across Lang call boundaries, so an interprocedural `-O3` claim
  would be inert by construction."
- `emitProgram` emits the **entire program as one translation unit**, and
  refuses any `ForeignContract` (`cgen_program.go:153-155`). So there is no
  second TU for LTO to cross.
- `emitProgram` emits **zero `restrict`** and, by D-11-09, zero call-boundary
  attributes — the only mention of the token in the file is the comment at
  `cgen_program.go:88` saying it never mints one. The single producer of
  `restrict`, `emitLinearBorrowedByPointer`, runs **only** at N=1.
- The four-tier runner's own doc comment already says it out loud
  (`session_phase11_differential_test.go:78-89`, D-11-25): "`-flto` is inert by
  construction for Lang-to-Lang code in this corpus … it is run here because
  running it costs nothing … NOT because this differential exercises real
  cross-TU LTO inlining."
- The two things that *do* prove the tier live are both non-multi-function:
  `TestLTOTierIsNotInert` (`native/native_lto_test.go:155-178`) uses
  `testdata/phase5/inline_across_foreign.lang`, a **single-function** program
  whose cross-TU boundary comes from the frozen *foreign* C unit; and NAT-07's
  composition-only divergence control is **hand-written C, explicitly not
  emitted by cgen** (`native_lto_test.go:181-200`, D-11-24).

**Why this is unnamed rather than known.** D-11-25 is a *design-decision ID in a
test comment*. It is not a row in `PHASE-11-DEBT.md`'s items table, it is not in
the milestone audit's ten unowned items, and it has no owner, no landing phase
and no reopening condition. Meanwhile `PROJECT.md`'s Validated list reads:
"Interpreter, `-O0`, `-O3`, and `-O3 -flto` produce equivalent semantic outcomes
and events for the interprocedural corpus, with the tier proven non-inert by an
engineered composition-only divergence." Both halves are true; the conjunction
reads as if the non-inertness proof applies to the interprocedural corpus, and
it does not — the two clauses are about two different programs, one of which
`cgen` did not write.

**Second-order consequence, and why it matters for Cluster A.** The cheapest
route to a *live* multi-function LTO tier is precisely the by-pointer +
call-boundary-attribute family that A3 proposes to cut (D-11-11/D-11-12's
discharge pairs), or a translation-unit split of `emitProgram`. So cutting that
family is not free — it defers the only mechanism that would make the LTO tier
non-inert for Lang-to-Lang code. That should be stated in the amendment, not
discovered later.

**Recommended disposition:** file it as a real warning-severity debt row in
M003's first register, with a reopening condition ("`emitProgram` splits into
more than one translation unit, **or** a multi-function program emits any
optimizer-visible attribute") and, in M003, add a one-line qualification to
`PROJECT.md`'s NAT-07 Validated bullet naming the fixture the non-inertness
proof actually used. Honest scoping, in the D-12-26 / D-10-37 style the project
already uses.

Two smaller unnamed items, for completeness:

- **`emitProgram` hardcodes `"live_resources":[]`** (`cgen_program.go:299`).
  Correct today (no foreign in multi-function) but it is a *literal*, not a
  derivation — the moment foreign bodies are ported it becomes a silent wrong
  answer on a comparator axis (`AxisResourceLedger`). Convert it to a derivation
  that is provably empty today, per D-04-12's "no hand-written value beside a
  generated one."
- **`LANGUAGE-MATURITY.md` is stale** on the guard inventory (26 non-test
  matches, not 32) and on `reduce` (`Reduce`/`ProjectSource` were widened at
  plan 11-08). A calibration document that is wrong in the optimistic direction
  is worse than none.

---

## Role-lens disagreements

Where the lenses genuinely pull apart, and how I resolved it:

- **Compiler-backend architect vs. engineering manager.** The architect wants
  A2 (one lowering path, now) — parallel paths rot, and LLVM's SelectionDAG /
  FastISel / GlobalISel history is the canonical warning. The manager sees three
  phases of zero user-visible output. *Resolution:* A3 splits along the value
  line rather than the purity line. The architect gets one path for the shapes
  M003 actually touches; the manager gets the expensive third of the work
  formally deferred with a named condition rather than silently carried.
- **C/Clang specialist vs. observability engineer.** The C specialist objects to
  threading an extra parameter through every emitted function: it perturbs the
  ABI of the generated TU and the frozen goldens. The observability engineer
  notes that context propagation is *exactly* how W3C Trace Context works and
  that there is no cheaper way to get occurrence identity. *Resolution:* the
  static path table (option b) — the threaded value is a compile-time constant
  `unsigned`, so the optimizer is unconstrained and the ABI concern is confined
  to `static` functions inside one TU.
- **Content-addressing specialist vs. observability engineer.** This one is not
  a tie. The Merkle instinct — "hash the semantic payload" — is *wrong here* and
  would ship a guaranteed collision, because the two `leaf` activations are
  content-identical. Content addressing answers "is this the same value?";
  event identity answers "is this the same occurrence?" Unison's hashing is the
  right model for `ClosureDigest` and the wrong model for events.
- **Verification engineer vs. refactoring specialist.** The verifier wants
  exhaustive enumeration over diamond shapes before the deletion; the
  refactoring specialist wants characterization tests and a fast strangler
  cutover. *Resolution:* Phase 09's own shape already reconciles these — a
  bounded enumeration answering *one* named behavioural question, plus the
  differential whose expectations flip rather than being deleted.
- **Everyone vs. the language-maturity lens.** The maturity document's warning
  stands: none of this makes the language writable. That is the adversarial
  pass's whole case, and it is why A3 cuts what it cuts.

---

## Prior art and lessons (with sources)

- **Duplicated lowering paths rot — LLVM is the standing proof.** SelectionDAG
  and FastISel "are radically different and share very little code"; GlobalISel
  was explicitly designed so the fast and optimized selectors share one Core
  Pipeline, and its stated plan was to replace FastISel and then SelectionDAG.
  LLVM has carried *three* selectors for years. Lesson: the cost of a parallel
  path is not the duplicated code, it is that every subsequent semantic change
  must be paid N times — which is precisely what the return≠parameter widening
  is about to charge this project.
- **Cranelift chose one lowering path and then made it declarative.** The
  MachInst backend lowers to VCode in a single pass, and ISLE replaced
  hand-written pattern-matching with a DSL "with a relatively low defect rate
  over the migration." Lesson: the migration succeeded because it was
  incremental *and* converged on one path — not because it kept both.
- **Strangler fig / branch by abstraction / parallel run.** Fowler's pattern is
  transform → coexist → **eliminate**; the eliminate step is the one that
  actually shrinks the system, and parallel-run exists for the capabilities
  where output correctness must be proven before cutover. Phase 09 already ran
  this playbook correctly (shadow run → blocking gate → same-commit deletion).
  Cluster A has been stuck in "coexist" for two milestones, which is the known
  failure mode.
- **Michael Feathers, *Working Effectively with Legacy Code*** — characterization
  tests pin current behaviour before it changes. `TestN1ConvergenceDifferential`
  is already a characterization test; it records the *measured* divergence
  rather than a hoped-for equivalence, which is why it is red-by-measurement and
  therefore trustworthy.
- **W3C Trace Context / OpenTelemetry.** A span identity is
  (trace-id, parent-span-id, span-id): the parent link travels in the
  identifier, and identity is per-*occurrence*, not per-code-site. Spec span-ids
  are random because distributed systems cannot coordinate; this system is
  deterministic and single-process, so the equivalent identity can be *derived*
  rather than drawn — which is strictly better, since it makes the identity
  re-derivable by a peer.
- **Lamport, "Time, Clocks, and the Ordering of Events in a Distributed
  System."** The relevant borrow is that causal structure must be carried
  explicitly; it cannot be recovered from a flat sequence. Today's document is a
  flat sequence with no call edge at all (no `OpCall` event), which is exactly
  why identity had to break before anyone noticed.
- **Unison, Git, Nix, Bazel — content addressing, and where it stops.** Unison
  identifies definitions by a hash of structure plus dependency hashes; Git
  identifies objects by content hash; Nix and Bazel key build artifacts on
  input/action digests. All four identify *values or derivations*. None of them
  identifies *occurrences* — Git does not distinguish two identical blobs,
  because it must not. That distinction is the whole reason option #2 fails
  here.
- **The project's own ClosureDigest near-miss** is the best available lesson and
  it is internal: a digest preimage that includes its own output field. The fix
  pattern (canonical preimage, field zeroed, domain separator, sorted
  components, plus an internal test that asserts non-self-reference and a
  mutation seam that proves the sort is load-bearing) should be applied as a
  checklist to any new identity, including this one.

---

## Adversarial pass

**The strongest case that both clusters should be deferred out of M003, and
D-10-60 simply retired:**

The language is 5-10% of a language you could write a program in. You cannot add
two numbers. You cannot loop. Hello-world requires a `foreign C` declaration.
Against that, M003 is being asked to spend its first several phases on (a)
deleting backend code that emits correct output today, and (b) fixing an
identity scheme for events in programs nobody can write. Neither delivers one
token of language surface. The assurance stack is already unusually far ahead of
the language — the project's own maturity document says so — and the honest read
is that widening the axis that is at 5% beats polishing the axis that is at 65%.

On D-10-60 specifically: it is a self-imposed process rule with no external
stakeholder, unenforced by CI (its own row admits this), and it is now
generating a forced ~1,500-line backend project on an item whose only benefit is
tidiness. Worse, the emitters are about to be rewritten anyway when control flow
and arithmetic land — so deleting them *now* is deleting code that a later
change would delete more cheaply, for free. And the goldens pin bytes, so the
port also buys a re-pinning churn that touches M001 evidence. A rule that
converts "we chose not to do low-value work twice" into "we must now do it" is
producing sunk-cost commitment and should be retired with thanks.

**Rebuttal, and one real concession:**

The case is correct about *half* the work and wrong about the other half, and
wrong about the rule.

- Wrong about branch/match, and this is the load-bearing rebuttal. `emitProgram`
  cannot emit a `match` body at all, and `match` is the **only** control flow
  the language has. So "widen the type system" and "add real control flow" both
  produce, on their very first fixture, a multi-function program with a branch —
  which is currently unlowerable (`cgen_program.go:143-145`, D-11-52). The
  branch/match port is not cleanup competing with language surface; it is a
  **precondition** for M003's language surface. Mislabelling it "emitter
  deletion" is what made it look optional for two milestones.
- Wrong about "they'd be rewritten anyway, so deleting now costs more." The
  opposite: `emitProgram` hardcodes one type per function, so keeping the legacy
  family means implementing the return≠parameter widening in both families.
  Deferring converts a one-time port into a recurring tax whose first
  installment is due inside M003.
- Wrong about event identity being about programs nobody can write. Reusing a
  function from two call sites is not exotic; it is the first thing any real
  program does. Today that program cannot be four-tier compared at all. The gap
  is not cosmetic and it is not hypothetical — it is a hole in the one axis the
  five-axis comparator was extended for in M002.
- Wrong about retiring D-10-60, for a reason independent of this item: the rule
  already contains the escape hatch being asked for. Retiring it would trade an
  explicit amendment for an unexamined drift, and would do so on the first
  occasion the rule produced friction — which is the definition of removing a
  forcing function because it worked.

**Concession, and it is a real one:** the adversarial case is right about
`emitLinearForeign`, `emitLinearBorrowedByPointer` and
`emitLinearBorrowedByPointerPlain`. Those three have **no M003 consumer**. Their
port opens two designed-not-built subsystems (D-11-11/D-11-12 discharge pairs,
D-10-34's multi-frame nonlocal pad). Forcing them into M003 on the strength of a
process rule would be exactly the sunk-cost behaviour the adversary describes.
That concession *is* the recommendation: cut them explicitly, under D-10-60's
own amendment clause, with a named M004 landing and the LTO-inertness
consequence stated in the amendment.

---

## Proposed M003 requirements this implies (draft REQ text, testable)

- **OBS-01 (event identity).** Every event in a `lang.execution/2` document
  carries an `invocation` identifying one activation of one function, derived
  solely from the entry function ID and the chain of `OpCall` operation IDs
  reaching it. The pair `(invocation, id)` is unique within a document.
  *Testable:* `multi_function_diamond_call.lang` passes
  `session.Phase5CompareEngines` across interpreter / `-O0` / `-O3` /
  `-O3 -flto`; `native.validateExecution` refuses a document with a repeated
  `(invocation, id)` pair and accepts one with a repeated `id` alone; the
  existing `DiamondSharedLeaf` subtest's inverted assertion is flipped rather
  than deleted, per its own instruction.
- **OBS-02 (identity is not self-referential).** No event field participates in
  the derivation of any event's `invocation`. *Testable:* an internal test in
  the `originvalidate_internal_test.go:36` shape — mutate every non-`invocation`
  field of an event and assert `invocation` is unchanged.
- **OBS-03 (peer re-derivation of identity).** A validator that imports neither
  `interp` nor `cgen` enumerates the admissible invocation set from the checked
  `core.Program` alone and verifies membership, uniqueness, last-step/`FunctionID`
  consistency, and pre-order structure for every event in a decoded document.
  *Testable:* `go list -deps` guard in the existing four-guard style; a seeded
  fault (an event attributed to the wrong frame) is caught by the peer and by no
  other check.
- **OBS-04 (bounded identity space, fail closed).** If the unfolded call tree
  exceeds the declared ceiling, emission is refused with a named error.
  *Testable:* a synthetic diamond chain past the ceiling produces the named
  refusal, never a timeout and never a truncated table.
- **NAT-08 (one lowering path for branch and match).** `emitProgram` emits
  `core.Match` bodies, `defect` terminators and payload construct/destructure
  arms; `emitMatch`, `emitBranch` and `emitLinear` are deleted in the same
  commit that flips dispatch. *Testable:*
  `TestN1ConvergenceDifferential`'s expectations flip to byte-identity on
  `toggle.lang`, `owned_transfer.lang` and `borrowed_view.lang`;
  `grep -c 'func emitMatch\|func emitBranch(\|func emitLinear('` returns 0;
  `previousPhaseGoldenCDigests` is re-pinned with each moved digest justified in
  writing.
- **NAT-09 (no legacy emitter is widened).** No surviving single-function
  emitter is extended to support return type ≠ parameter type; a program
  requiring both is refused with a named diagnostic code. *Testable:* a fixture
  combining by-pointer lowering with a differing return type produces the named
  code, and a source scan shows the two-type model appears in exactly one
  emitter family.
- **NAT-10 (LTO inertness is disclosed).** The `-O3 -flto` tier's non-inertness
  claim names the fixture and translation-unit boundary it rests on, and the
  absence of a cgen-emitted multi-function cross-TU boundary is carried as an
  open debt row with a reopening condition. *Testable:* a test asserts the
  disclosure text exists beside the claim, in the D-04-12 derived-not-restated
  style.
- **PRC-01 (D-10-60 is mechanized).** Each debt row declares
  `superseded_landing_phases`; `TestDebtRegistersAreWellFormed` fails when that
  list is non-empty unless the row names a REQUIREMENTS.md amendment.
  *Testable:* the test fails on a synthetic twice-deferred row with no
  amendment, and passes with one.

---

## Confidence + what would change my mind

| Claim | Confidence | Basis |
|---|---|---|
| The six emitters, their line ranges, and what each uniquely emits | **HIGH** | Read `cgen.go` and `cgen_program.go` directly; line ranges computed from the file |
| `emitProgram` is a subset, not a superset | **HIGH** | Three explicit refusals at `cgen_program.go:143-155`; `TestN1ConvergenceDifferential`'s measured table |
| The diamond collision and its mechanism | **HIGH** | Executed the fixture; read both derivation sites |
| Content addressing cannot solve event identity | **HIGH** | Follows from the two activations being content-identical; not a judgment call |
| `-flto` is inert for cgen-emitted multi-function programs | **HIGH** | D-11-25's own comment, plus `emitProgram` emitting one TU, no `restrict`, no foreign |
| That inertness is *unowned* (no debt row, not in the audit) | **MEDIUM-HIGH** | Checked the Phase 11 items table and the audit's ten unowned rows; D-11-25 appears in neither |
| Effort: branch/match port ≈ 600-800 lines, 4-6 plans | **MEDIUM-HIGH** | Line counts are measured; plan count extrapolates from Phase 09's comparable deletion (3 plans for a much smaller surface) |
| Static path table is the right C-side mechanism | **MEDIUM** | Unmeasured. The `2^d` unfolding bound is real; the ceiling makes it safe, but the ceiling value is a guess |
| Keeping D-10-60 and discharging via amendment | **HIGH** | The rule's own text authorizes it; D-10-30 set the precedent |

**What would change my mind:**

- **If M003 does *not* take type widening or control flow.** Then the
  branch/match port loses its consumer and A3 collapses into A2-lite with no
  M003 value — cut all six explicitly and defer the whole cluster. The
  recommendation is contingent on M003's other scope; say so at the roadmap
  gate.
- **If the unfolded call tree blows up on a realistic corpus** (measure it on
  `deep_diamond_acyclic.lang` first — 13 functions). If node counts are
  uncomfortable at realistic sizes, switch to runtime concatenation (option a)
  or the hashed path (option c/#5), keeping the same grammar.
- **If threading `unsigned lang_inv` measurably changes `-O3` behaviour** in a
  way that produces a tier disagreement. I expect constant propagation to erase
  it; if not, the parameter must become a global or the scheme reverts to a
  per-function static table keyed differently.
- **If a `function.called` event turns out to be required** for the peer's
  pre-order check to be decidable on a document alone. I believe prefix
  structure over `invocation` is sufficient; if a counterexample exists, the
  call event moves from optional to required and the schema bump should carry it
  in one go rather than forcing a `/3`.
- **If someone shows the by-pointer port is cheap** — i.e. that D-11-11's
  discharge pairs are not actually required because a callee-side `restrict` on
  a `static` function in a single TU needs no caller-side discharge. That would
  move `emitLinearBorrowedByPointer` from the cut half to the landed half and
  simultaneously de-fang the LTO-inertness risk. This is the single most
  valuable thing to spike before the roadmap is fixed.

---

## Sources

**Internal (all read at the cited path):**
`internal/compiler/cgen/cgen.go` (112, 142, 235, 317, 588, 615, 749, 809, 931,
953, 1148, 1739, 1757, 1979, 2509, 2550) ·
`internal/compiler/cgen/cgen_program.go` (44, 88, 120-310, 330-436) ·
`internal/compiler/cgen/cgen_n1_convergence_test.go` ·
`internal/compiler/interp/interp.go` (619-641, 800-851) ·
`internal/compiler/execution/execution.go` (9-10, 41-57) ·
`internal/compiler/native/native.go` (550-615) ·
`internal/compiler/native/native_lto_test.go` (140-200) ·
`internal/compiler/callgraph/callgraph.go` (279, 460) ·
`internal/compiler/core/core.go` (300-337) · `internal/compiler/core/core_test.go` (148-210) ·
`internal/compiler/originvalidate/originvalidate.go` (555-600) ·
`internal/compiler/originvalidate/originvalidate_internal_test.go` (36) ·
`internal/compiler/session/session_phase5_compare.go` (103-187) ·
`internal/compiler/session/session_phase11_gate.go` (30-90) ·
`internal/compiler/session/session_phase11_differential_test.go` (73-330) ·
`internal/compiler/session/session_phase5.go` (225-240) ·
`internal/compiler/check/check_exclusive_test.go` (233-249) ·
`testdata/phase11/multi_function_diamond_call.lang` ·
`testdata/phase5/inline_across_foreign.lang` ·
`.planning/milestones/M002-MILESTONE-AUDIT.md` §4 ·
`…/10-trusted-interprocedural-oracle/PHASE-10-DEBT.md` (D-10-30, D-10-60) ·
`…/11-…/PHASE-11-DEBT.md` (D-11-02, D-11-51, D-11-52, D-11-27) ·
`…/11-…/11-GUARD-LEDGER.md` ·
`…/12-result-payloads/PHASE-12-DEBT.md` (D-12-21, D-12-36, D-12-43, D-12-44) ·
`…/09-…/09-08-PLAN.md`, `09-09-SUMMARY.md`, `09-VERIFICATION.md` ·
`…/07-…/07-RESEARCH.md:366`, `07-REVIEWS.md` ·
`.planning/PROJECT.md`, `.planning/LANGUAGE-MATURITY.md`,
`.planning/STANDING-VERDICTS.md`, `.planning/MILESTONES.md`.

**External:**
- [LLVM — Global Instruction Selection](https://llvm.org/docs/GlobalISel/)
- [Cranelift's Instruction Selector DSL, ISLE — Chris Fallin](https://cfallin.org/blog/2023/01/20/cranelift-isle/)
- [A New Backend for Cranelift, Part 1: Instruction Selection — Mozilla Hacks](https://hacks.mozilla.org/2020/10/a-new-backend-for-cranelift-part-1-instruction-selection/)
- [bytecodealliance/rfcs — Cranelift ISel: ISLE](https://github.com/bytecodealliance/rfcs/blob/main/accepted/cranelift-isel-isle-peepmatic.md)
- [Martin Fowler — Strangler Fig Application](https://martinfowler.com/bliki/StranglerFigApplication.html)
- [Martin Fowler — Branch By Abstraction](https://martinfowler.com/bliki/BranchByAbstraction.html)
- [Martin Fowler — Parallel Change (expand/contract)](https://martinfowler.com/bliki/ParallelChange.html)
- [Strangler Fig Pattern — Azure Architecture Center](https://learn.microsoft.com/en-us/azure/architecture/patterns/strangler-fig)
- [W3C Trace Context (Recommendation)](https://www.w3.org/TR/trace-context/)
- [OpenTelemetry — Traces (spans, span id, parent span id)](https://opentelemetry.io/docs/concepts/signals/traces/)
- [Traceparent Header: Format and Fields Explained — Last9](https://last9.io/blog/traceparent-explained/)
- [Lamport — Time, Clocks, and the Ordering of Events in a Distributed System (1978)](https://lamport.azurewebsites.net/pubs/time-clocks.pdf)
- [Unison — The big idea (content-addressed code)](https://www.unison-lang.org/docs/the-big-idea/)
- [Git Internals — Git Objects](https://git-scm.com/book/en/v2/Git-Internals-Git-Objects)
- [Bazel — Remote Caching (action keys)](https://bazel.build/remote/caching)
- [Nix — Derivations](https://nix.dev/manual/nix/stable/language/derivations)
- [SHAttered — SHA-1 collision](https://shattered.io/)
- Michael Feathers, *Working Effectively with Legacy Code* (characterization
  tests; safe deletion) — book, no canonical URL.
