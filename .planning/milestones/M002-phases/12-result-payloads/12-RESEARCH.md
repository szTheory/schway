# Phase 12: `Result` Payloads - Research

**Researched:** 2026-09-12
**Domain:** Compiler IR extension (sum-type payloads), parser/grammar, interpreter value-model widening, C17 ABI/layout codegen, cross-engine differential verification
**Confidence:** HIGH

## Summary

Phase 12 has already been researched exhaustively by an eight-way advisor
fan-out at `/gsd-discuss-phase`, every load-bearing claim independently
re-verified against the shipped tree by the discussion orchestrator, and the
criterion-3 pre-flight gate has already run and returned **BRANCH A
(accepted)** (`internal/compiler/corevalidate/corevalidate_result_payload_probe_test.go`,
commit `3f9ceee`). This RESEARCH.md's job is therefore not to re-litigate any
of that — `12-CONTEXT.md`'s 42 decisions (D-12-01..D-12-42) are locked — but to
(1) independently re-verify the tree-facts a planner will build tasks on top
of, this session, so the planner does not have to re-open every cited file
itself, (2) organize those facts into the architecture/pitfalls/code-example
shape a planner consumes, and (3) supply the Validation Architecture and
Security Domain sections the phase's config toggles require, which
`12-CONTEXT.md` does not itself produce.

**Every claim below carrying a `core.go`/`ast.go`/`parser.go`/`interp.go`/
`corevalidate.go`/`cgen.go`/`cgen_program.go`/`session_phase5_compare.go`
line citation was re-opened and read with `Read`/`sed -n` in this research
session** (not merely grepped), independently of the discussion orchestrator's
own pass — so the two verifications are now redundant confirmations of the
same lines, not one unverified claim repeated twice.

**Primary recommendation:** Build the two new `core.OperationKind`s
(construct/destructure) first, in a plan that also updates
`TestAllOperationKindsHandledAtEverySite`, `AllOperationKinds()`, and all six
literal dispatch sites in one green commit; only then touch the parser binder,
the `interp` value-struct widening, and the `cgen` layout/emitter work, each in
its own plan; land the inherited D-11-02 six-emitter deletion (gate-then-N=1-
differential-then-delete) strictly *before* writing any `Result`-specific
`cgen` `case` arms, so there is exactly one emitter family to write them into.

## User Constraints (from CONTEXT.md)

<user_constraints>

### Locked Decisions

All 42 decisions in `12-CONTEXT.md` (`D-12-01` through `D-12-42`) are locked —
copied here by reference rather than verbatim (the source file is 1037 lines);
the planner **must** read `.planning/phases/12-result-payloads/12-CONTEXT.md`
directly before planning, not rely on this summary. Headline locks, by area:

1. **Criterion 3 pre-flight probe (D-12-01..D-12-04c):** committed Go test,
   already run, **BRANCH A accepted** — Phase 12 proceeds. Do not re-run or
   re-litigate.
2. **Core IR representation (D-12-05..D-12-10):** TWO NEW `core.OperationKind`s
   (construct, destructure) — an explicit deviation from D-04-30's own
   recorded omitempty-field design, which is rejected. `PayloadType` is a bare
   string (matches `ReturnType`/`Parameter.Type`/`MatchArm.Pattern`
   convention). No `lang.core/2` schema bump. A single constructor function is
   the only permitted way to build a payload-carrying `core.DataType`, plus a
   `corevalidate` invariant arm. One minting authority for the destructured
   place (`PayloadTargetID`, following `OkEdgeID`/`ErrTargetID` naming).
3. **Source syntax (D-12-11..D-12-16):** `| Ok(Buffer)` declares,
   `Ok(v) => {...}` binds — constructor-application syntax. General `data`
   feature; `Result` is a **user fixture**, not a built-in. Binding a payload
   **moves** it via a real `OpMove`, with loud diagnostics (Rust RFC 2005
   match-ergonomics risk adopted as a constraint). Three new named fail-closed
   refusals: arity mismatch, binder-on-nullary, missing-binder-for-payload.
   Bare-tag-plus-`take` syntax is rejected (needs flow-sensitive narrowing the
   checker doesn't have).
4. **Interpreter value model (D-12-17..D-12-21):** widen to
   `value{tag, payload string}` / `map[string]value`, not an interface/sum
   type, not string-encoding, not parallel maps. Evidence-invisible by
   construction: `tag == ""` on every non-payload site, serialized as the bare
   string. D-11-51 (duplicate event-ID) flagged as likely-tripped-first by this
   phase's fixtures — NOT claimed or fixed here.
5. **C17 layout (D-12-22..D-12-26):** FLAT per-alternative-field struct, never
   a real C `union` (UB/TBAA/first-non-checker-derived-ABI-claim risk). Niche
   optimization is **designed and recorded as debt, not built** — it is
   currently **uninstantiable** (`Byte` is plausibly fully inhabited, no niche
   exists). Layout is one shared derived fact all three engines read (a
   narrow, deliberate exception to peer-independence discipline). "One
   meaning" = observable-behavior agreement, not byte-identical layout
   (`interp` has no layout at all).
6. **Resource payloads (D-12-27..D-12-30):** REFUSED this phase, named
   fail-closed diagnostic, three-part landing condition (D-10-C01 closed AND
   D-10-C02 order-independence proven AND D-10-C04 reviewed) recorded as debt.
   "Dropped exactly once" means a static by-function-end accounting property,
   not a countable runtime event (no `OpDrop` exists).
7. **Inherited D-11-02 deletion (D-12-31..D-12-36):** TAKE IT — gate (N=1
   convergence differential green) → lift D-11-52's refusal → delete the six
   emitters, re-pin digests in an isolated diff → write payload `case`s once.
   Represent as explicitly-labeled inherited-debt-closure plans with
   non-requirement-shaped verification, kept separate from RES-02/RES-03's
   actual requirement-shaped criteria. Fallback (re-defer) triggers only if the
   N=1 differential cannot be made green cheaply.
8. **Criterion 2 anti-vacuity (D-12-37..D-12-42):** TWO controls — a frozen
   `testdata/phase12/*.golden.c` struct-shape mutation (necessary, not
   sufficient) AND a reverted production-hunk `cgen` mutation that swaps which
   alternative's slot a match arm reads for a correct tag, producing a genuine
   interp-vs-native value divergence, with a companion unmutated run proving
   the harness reports agreement absent the bug. **No harness extension is
   needed** — `session_phase5_compare.go:89-96`'s existing `Outcome.Value`
   comparison on `axis:terminal-outcome` already includes "the ok payload and
   the err-edge ADT alternative". QLT-09's Phase-5 Nyquist portion is NOT
   closed this phase.

### Claude's Discretion

- Plan decomposition and wave structure, subject to D-12-31's ordering
  (gate-then-delete before payload `case`s) and D-12-34's labeling requirement
  (inherited-debt-closure plans kept separate from RES-* criteria).
- Which file each new derivation lives in, subject to the existing
  package-independence import allowlists.
- Exact diagnostic code strings for D-12-15's three new pattern refusals and
  D-12-27's resource-payload refusal, within the established `check.*` /
  `core.*` namespacing (e.g. `core.callee_not_callable`,
  `syntax.expected_pattern`, `check.interprocedural_loan_liveness` are the
  established naming shape).
- Whether `Fault` (the example payload ADT in `| Err(Fault)`) or another name
  is used in fixtures.
- Whether the corpus characterization replay (D-12-19) lives as one test or a
  table-driven family.

### Deferred Ideas (OUT OF SCOPE)

- Resource-carrying payloads — three-part landing condition (D-10-C01 +
  D-10-C02 + D-10-C04), none satisfied yet.
- Niche optimization — reopens only when a payload type with a provably
  invalid bit pattern is added; currently uninstantiable.
- D-10-C01's missing `case core.OpCall` in `peerDeriveOriginFacts`; D-10-C02's
  order-independence proof; D-10-C04's human review — all OPEN and UNOWNED,
  not repaired this phase.
- D-11-51's per-invocation event-ID collision — flagged, not fixed; needs its
  own reviewed plan touching `interp.go` AND `cgen_program.go`.
- D-11-07's `lang.foreign/0` widening — OPEN and unowned.
- QLT-09's Phase-5 Nyquist portion — end-of-Phase-12 or Phase 13, against a
  pre-registered threshold, not this phase.
- A real generic `Result<T, E>` — GEN-01, M003.
- Bare-tag-plus-`take` syntax — needs a flow-sensitive narrowing pass first.

</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| RES-02 | `Result` values with payload-carrying alternatives are storable and matchable; moving out of a matched payload obeys the affine drop obligation (D-04-30). | Architecture Patterns §"Two new OperationKinds" + §"Move-on-bind"; Common Pitfalls §"Vacuous exhaustive dispatch"; Code Examples §1-3; scoped to `Byte`/`Buffer`/nullary-ADT payloads only (D-12-27) — resource payloads are named debt, not this requirement's scope. |
| RES-03 | `Result` layout — tagged union, with niche optimization where a checked ability fact permits it — has one meaning in the core IR, the interpreter, and emitted C17. | Architecture Patterns §"Flat struct layout, one shared fact"; Common Pitfalls §"Niche is vacuous, not absent"; Validation Architecture §"Anti-vacuity control"; "one meaning" is defined precisely in D-12-26 as observable-behavior agreement, not byte layout. |

</phase_requirements>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Payload grammar (`\| Ok(Buffer)`, `Ok(v) => {...}`) | Frontend (syntax/parser) | Check (semantic validation) | Parser owns token-level grammar; `check` owns arity/binder-shape refusals (D-12-15) since they need type/alternative information the parser doesn't carry. |
| Payload construct/destructure semantics | Core IR (`core` package) | check, corevalidate, interp, cgen, pathoracle, originvalidate | Two new `OperationKind`s are the IR's single source of truth for payload meaning; the six dispatch sites are consumers, not co-owners (D-12-05). |
| Affine drop obligation for payloads | check (worklist) + corevalidate (independent re-derivation) | — | Mirrors the project's existing two-independent-knower pattern for every other ownership fact; "dropped exactly once" is a static accounting property both peers already derive independently (D-12-29). |
| Payload origin/ownership across `OpCall` | corevalidate (`peerDeriveOriginFacts`) | check, originvalidate, pathoracle | Already proven Phase 08-11 machinery per the accepted probe (BRANCH A) — this phase adds no new interprocedural rule; it only exercises the existing four-peer admission gate on payload-shaped fixtures. |
| Interpreter value representation | interp (`frame.values map[string]value`) | — | Sole owner; no byte layout obligation (D-12-26) — `interp` is the semantic oracle, not a layout consumer. |
| C17 struct layout for payload alternatives | cgen (via `core.RecordLayout`/`LayoutField`) | check (derives the shared layout fact) | Layout is a *definition*, computed once and read by all consumers (D-12-25) — a deliberate narrow exception to the four-peer independence discipline, which governs *judgements* not *definitions*. |
| Cross-engine equivalence proof | session (`session_phase5_compare.go`, the five-axis comparator) | — | Sole owner of "one meaning" verification; existing `axis:terminal-outcome` already carries payload values (D-12-39) — no new axis needed (D-12-40/D-11-42). |

## Standard Stack

Not applicable in the conventional sense: this phase adds **zero third-party
dependencies**. It is pure extension of an existing, fully in-house Go 1.24
compiler/interpreter/C17-codegen pipeline (`internal/compiler/*`). The
"stack" here is internal architectural convention, documented in Architecture
Patterns below rather than as an installable-package table.

**Toolchain versions in this environment** (`[VERIFIED: local shell]`, checked
this session):
- Go: `go version go1.24.0 darwin/arm64`
- Clang: `Apple clang version 21.0.0 (clang-2100.1.1.101)` — the same compiler
  version named throughout `12-CONTEXT.md`'s `-O3 -flto` reasoning (D-12-22,
  D-12-38).

### Alternatives Considered

No third-party alternatives were in scope — see `12-DISCUSSION-LOG.md` areas
2 ("Follow D-04-30 as written" vs. two new kinds vs. hybrid), 4 (struct vs.
interface vs. string-encoding vs. parallel maps for `interp`'s value model),
and 5 (flat struct vs. real C `union`) for the in-house design alternatives
that were considered and rejected, each with its rejection reason preserved
in `12-CONTEXT.md`'s decision text.

**Installation:** none — no new packages.

## Package Legitimacy Audit

**Not applicable.** This phase installs zero external packages (Go modules or
otherwise). No `go get`, no npm/pip/cargo dependency. Skipping the
Package Legitimacy Gate protocol is correct here because its trigger
condition ("whenever this phase installs external packages") does not fire.

## Architecture Patterns

### System Architecture Diagram

```
 Source (.lang)
     │
     ▼
 syntax.parser  ──"Ok(v) => {...}"──▶  ast.MatchArm (NEW: binder field)
     │  (D-12-11, D-12-13: genuine new parser work —
     │   today's grammar reads ONE identifier then `=>`,
     │   parser.go:530-533, no binder slot exists)
     ▼
 check (semantic pass)
     │  ├─ resolves alternative payload type (new field on ast.Alternative)
     │  ├─ emits OpConstruct[Payload] / OpDestructure[Payload]  (NEW OperationKinds)
     │  ├─ emits a real OpMove at arm entry for bound payloads (D-12-14)
     │  ├─ refuses: arity mismatch / binder-on-nullary / missing-binder (D-12-15)
     │  ├─ refuses: resource-carrying payload type (D-12-27, fail-closed)
     │  └─ derives the SHARED payload RecordLayout fact (D-12-25) via
     │      standardForeignLayout()'s existing pattern (check.go:3118)
     ▼
 core.Program (JSON IR)
     │  DataType{Alternatives []string, + new payload-detail list, name-set
     │  invariant enforced at construction — D-12-09}
     │  LinearOperation{Kind: OpConstructPayload | OpDestructurePayload,
     │                  PayloadTargetID string, ...}
     │
     ├──────────────┬───────────────┬───────────────┬──────────────┐
     ▼              ▼               ▼               ▼              ▼
 corevalidate   originvalidate  pathoracle      interp         cgen
 (independent   (independent    (independent    (semantic      (native C17)
  re-derivation, re-derivation) re-derivation)   oracle)
  4th peer)          │               │               │              │
     │               │               │               │              │
     │   all four admission peers consult the SAME shared        reads the
     │   payload RecordLayout fact (D-12-25) — a deliberate      SAME shared
     │   narrow exception to "independence by derivation         layout fact;
     │   method + import boundary" (D-09-02), justified          emits FLAT
     │   because layout is a definition, not a judgement.        per-alternative
     │                                                             struct (D-12-22),
     ▼                                                             never a real
 Five-axis differential comparator (session_phase5_compare.go)     C union
     │  axis:terminal-outcome already compares Outcome.Value
     │  "INCLUDING the ok payload and the err-edge ADT
     │  alternative" (line 89-96 comment, D-12-39) — no new
     │  axis, no harness extension needed (D-12-40).
     ▼
 "One meaning" verified as OBSERVABLE-BEHAVIOR AGREEMENT across
 interp / -O0 / -O3(-flto), never byte-identical layout (D-12-26 —
 interp has no sizes/alignments/offsets at all).
```

A reader can trace the primary use case — a `Result`-shaped `data Outcome =
| Ok(Buffer) | Err(Fault)` value constructed, matched, and its payload moved
out — from parse (top) through IR (middle) to the five-axis proof (bottom) by
following the arrows.

### Recommended Project Structure

No new top-level directories. All work lands inside existing packages:

```
internal/compiler/
├── ast/ast.go            # ast.Alternative gets a payload-type field (additive);
│                          # ast.MatchArm gets a binder field (D-12-13's genuine
│                          # new parser work — NOT an extension of existing Value)
├── syntax/parser.go       # matchExpr() at :523, alternative-declaration parse
│                          # site both grow constructor-application syntax
├── core/core.go           # DataType grows payload-detail list; TWO new
│                          # OperationKind consts; AllOperationKinds() :704
│                          # literal slice gets both; LinearOperation grows
│                          # PayloadTargetID (D-12-10, mirrors OkEdgeID/ErrTargetID)
├── core/core_test.go      # TestAllOperationKindsHandledAtEverySite :267 is the
│                          # anti-vacuity control this whole design serves —
│                          # every new kind MUST get real fixture coverage here
├── check/check.go         # checkBranch :1597 (arm-level place binding, where
│                          # OpMove-on-bind is emitted); new constructor function
│                          # enforcing DataType's name-set invariant (D-12-09);
│                          # standardForeignLayout :3118's pattern reused for
│                          # the shared payload RecordLayout (D-12-25)
├── corevalidate/
│   └── corevalidate.go    # NOT modified for the probe's D-10-C01 gap this
│                          # phase (that stays open, D-12-28) — but DOES gain
│                          # the new payload construction/destructure re-
│                          # derivation arms at whichever dispatch switch
│                          # already handles OpCopy/OpMove (:2407 area)
├── interp/interp.go       # ~17 sites widen values map[string]string to
│                          # map[string]value{tag, payload string} (D-12-17)
├── cgen/cgen.go           # six emitters (ledger below) — DELETED per D-12-31
│                          # BEFORE payload case arms are written into
│                          # emitProgram's single surviving family
├── cgen/cgen_program.go   # emitProgram :128, D-11-52 refusal at :152 (lifted
│                          # by D-12-31 step ii), emitProgramFunction :339
├── session/
│   └── session_phase5_compare.go  # NO CHANGE NEEDED (D-12-39) — existing
│                          # axis:terminal-outcome already compares payload
│                          # values; used as-is for the anti-vacuity control
├── session/session.go     # LayoutMutationRunner :565-587 — pattern D-12-37's
│                          # frozen-fixture control clones for payload layout
└── ability/ability.go     # extension point for D-12-24's recorded (not
                            # built) niche/inhabitance fact — currently
                            # uninstantiable, do not build this phase
testdata/
├── phase12/               # NEW: fixtures for payload construct/match/move,
│                          # the frozen golden.c mutation control (D-12-37),
│                          # the N=1 convergence differential fixture (D-12-31)
├── phase07/
│   └── call_from_both_match_arms.lang  # probe's match-arm-call reduction
│                          # shape — header already names it Phase 12's guard
└── phase08/
    └── relay_depth2_accept.lang  # probe's forwarding reduction shape —
                            # header needs correction per D-12-04c (stale
                            # move_while_borrowed claim, superseded by D-09-03)
```

### Pattern 1: Two new `OperationKind`s, following the `OkEdgeID`/`ErrTargetID` precedent

**What:** Payload construction and destructuring each get a genuinely new
`core.OperationKind`, rather than riding `OpCopy`/`OpMove` with a
conditionally-populated field.

**When to use:** Any time a new IR event has a *different meaning* from every
existing kind, not merely a refinement of one kind's existing meaning. The
precedent that validates this design is already in the tree:

```go
// Source: internal/compiler/core/core.go:704 (AllOperationKinds — the single
// table every dispatch site is tested against) and :722-730 (verified this
// session, quoted verbatim)
func AllOperationKinds() []OperationKind {
	return []OperationKind{OpCopy, OpMove, OpBorrowShared, OpBorrowExclusive, OpReturn, OpForeignCall, OpFail, OpRelease, OpDefect, OpCall}
}

// OkEdgeID, ErrEdgeID, and ErrTargetID are Phase 4 additive omitempty
// facts populated only on an OpForeignCall operation (D-04-04): the ok
// edge continues at TargetID (an ordinary place, exactly like OpCopy's
// target), while the err edge's own synthesized failure-ADT place is
// ErrTargetID. Every pre-Phase-4 operation, and every operation kind
// other than OpForeignCall, leaves all three empty.
OkEdgeID    string `json:"ok_edge_id,omitempty"`
ErrEdgeID   string `json:"err_edge_id,omitempty"`
ErrTargetID string `json:"err_target_id,omitempty"`
```

Note precisely what this precedent supports: `OkEdgeID`/`ErrEdgeID`/
`ErrTargetID` refine `OpForeignCall`'s **one existing meaning** ("a call with
two outcome edges"). D-12-05's argument (independently re-verified this
session by reading the field's own doc comment above) is that this precedent
does *not* extend to bolting a **second, conditionally-active** meaning onto
`OpCopy`/`OpMove` — that would make `TestAllOperationKindsHandledAtEverySite`
pass vacuously for the new semantics.

**Example — the anti-vacuity control this design must keep green:**
```go
// Source: internal/compiler/core/core_test.go:267-274 (verified this session)
// TestAllOperationKindsHandledAtEverySite is the control:kind.exhaustive_dispatch
// table (D-04-22): for every existing OperationKind, drive real programs
// containing that kind through check (session.Check), corevalidate,
// interp, cgen, pathoracle, and originvalidate, and assert none of them
// reject or crash. Every declared kind must be exercised by at least one
// fixture, so a kind that no corpus program ever produces cannot silently
// pass this control by omission.
```
Adding `OpConstructPayload`/`OpDestructurePayload` to `AllOperationKinds()`
forces this control to demand real fixture coverage at all six sites — which
is the entire point (D-12-05's "guarantee degrades to a comment" argument).

### Pattern 2: Additive-omitempty, byte-freeze discipline

**What:** Every new struct field this phase adds (`PayloadTargetID`, the
payload-detail list on `DataType`, the `PayloadType` string) must be
`omitempty` and populated *only* on operations/types that use the new
feature, so no existing fixture's serialized bytes move.

**When to use:** Any additive fact on an existing `core` struct. This is the
established pattern for `PublicOrigin`, `ForeignContract`, `CalleeID`,
`Blocks`/`Edges`/`LoanEndpoints` — all `omitempty`, all documented inline with
an explicit "every pre-Phase-N operation... leaves this empty" byte-freeze
claim (verified in `core.go`, lines 722-762, this session).

**Example:**
```go
// Source: internal/compiler/core/core.go:753-762 (verified this session,
// the CalleeID precedent — same additive-omitempty shape PayloadTargetID follows)
// CalleeID is Phase 07's additive omitempty fact (D-07-29): populated
// only on an OpCall operation, it holds the resolved callee's function
// ID -- never its name. Every pre-Phase-07 operation, and every
// operation kind other than OpCall, leaves this empty, so no
// pre-Phase-07 core artifact moves a byte.
CalleeID string `json:"callee_id,omitempty"`
```

### Pattern 3: Move-on-bind via a real `OpMove`

**What:** `Ok(v) => { ... }` emits a genuine `OpMove` from the scrutinee's
payload slot into a fresh place at arm entry — binding *is* consuming, not
merely naming.

**When to use:** Any pattern-match binder in this affine-typed IR. The
existing precedent for "arm entry does real ownership work" is
`checkBranch`'s own arm-level place binding:

```go
// Source: internal/compiler/check/check.go:1597 (verified this session,
// checkBranch's doc comment — the site where the new OpMove-on-bind lands)
func checkBranch(module, functionID, matchID string, function ast.FuncDecl,
	dataType core.DataType, sealed map[string]bool,
	calleeContracts map[string]calleeContract,
	foreignSymbols map[string]foreignSymbolInfo) (core.Function, []diagnostic.Diagnostic, int, map[string]diagnostic.Span) {
```

**Named risk to carry into diagnostics (D-12-14):** Rust RFC 2005's
match-ergonomics surprise — a move/non-move distinction invisible in the
pattern — is adopted as a binding constraint, not merely noted: diagnostics on
a payload binder move must be loud enough for both an AI author and a human
reviewer to see the consumption.

### Pattern 4: Flat per-alternative-field struct, one shared layout fact

**What:** C17 emission uses a struct with a tag field plus one field per
alternative (the tag says which is live), never a real C `union`. Layout is
computed once and read identically by `check` (deriving it),
`corevalidate`/`originvalidate`/`pathoracle` (consuming it for their re-
derivations), and `cgen` (emitting it) — a single shared fact, not three
competing derivations.

**Existing precedent to extend, not reinvent:**
```go
// Source: internal/compiler/check/check.go:3118-3125 (verified this session)
func standardForeignLayout() *core.RecordLayout {
	return &core.RecordLayout{
		Size: 1, Alignment: 1, ForeignTypeName: "lang_foreign_resource_block",
		Fields: []core.LayoutField{
			{Name: "payload", Size: 1, Alignment: 1, Offset: 0, CType: "unsigned char"},
		},
	}
}
```
```go
// Source: internal/compiler/core/core.go:111-127 (verified this session,
// core.RecordLayout / core.LayoutField — the exact types the payload layout
// fact will populate)
type RecordLayout struct {
	Size      int           `json:"size"`
	Alignment int           `json:"alignment"`
	Fields    []LayoutField `json:"fields"`
	ForeignTypeName string  `json:"foreign_type_name"`
}
type LayoutField struct {
	Name      string `json:"name"`
	Size      int    `json:"size"`
	Alignment int    `json:"alignment"`
	Offset    int    `json:"offset"`
	CType     string `json:"c_type"`
}
```

**Why not a real C `union` (D-12-22, re-verified this session as sound C17
reasoning):** reading an inactive union member is UB outside the
common-initial-sequence exception (C17 §6.5.2.3), and this project's own
`-O3 -flto` + TBAA + sanitizer proof regime is exactly the environment where
that UB gets exploited — a flat struct keeps every byte checker-derivable
with zero new machinery, at a measured cost of one tag byte over the two
currently-available payload types (`Byte`, `Buffer`).

### Pattern 5: Existing five-axis comparator already carries payload values — do not extend it

**What:** The differential harness comparing `interp`/`-O0`/`-O3(-flto)`
already includes extracted payload values in its terminal-outcome comparison.
No new axis, no new comparator field, is needed for criterion 2's anti-
vacuity control.

**Verified this session:**
```go
// Source: internal/compiler/session/session_phase5_compare.go:89-96
func comparePhase5Pair(fixture, pair string, left, right execution.Execution) error {
	// axis:terminal-outcome -- the closed value|typed_failure|defect set,
	// INCLUDING the ok payload and the err-edge ADT alternative. Schema is
	// compared here too: a schema mismatch makes the two documents
	// incomparable, which is itself a terminal-outcome-level disagreement.
	if left.Schema != right.Schema || left.Outcome.Kind != right.Outcome.Kind || left.Outcome.Value != right.Outcome.Value {
		return &Phase5EngineDisagreement{Fixture: fixture, EnginePair: pair, Axis: AxisTerminalOutcome, Detail: fmt.Sprintf("schema=%s/%s outcome=%+v vs %+v", left.Schema, right.Schema, left.Outcome, right.Outcome)}
	}
```
The five axis constants, confirmed at their cited lines:
```go
// Source: internal/compiler/session/session_phase5_compare.go:20-24
const (
	AxisTerminalOutcome  = "axis:terminal-outcome"
	AxisEventOrder       = "axis:event-order"
	AxisResourceLedger   = "axis:resource-ledger"
	AxisExitStatusSignal = "axis:exit-status-signal"
	AxisDiagnosticID     = "axis:diagnostic-id"
)
```
**Standing constraint (D-11-42, still binding):** `SelectLanesForFixture`'s
lane set must not be extended without the same scrutiny D-11-41/Q-02 applied
to the native-differential lane. Do not add a sixth axis for this phase.

### Anti-Patterns to Avoid

- **Riding `OpCopy`/`OpMove` with an omitempty payload field (D-04-30's
  literal recorded design):** rejected this phase — makes the exhaustive-
  dispatch control pass vacuously for the new affine semantics. Do not
  resurrect this shape even though it is the ROADMAP's own canonical ref;
  D-12-05 records the deviation explicitly.
- **A real C `union` for the payload:** UB outside common-initial-sequence,
  and the project's first non-checker-derived ABI claim. Use the flat struct.
- **String-encoding tag+payload into `interp`'s existing `map[string]string`:**
  in-band signalling — a payload containing the string `"err"` (already a
  literal foreign-err-edge value, `interp.go:667`) would corrupt the value
  space. Widen to a struct instead (D-12-20).
- **Parallel side maps (`values` + `tags` + `payloads`) in `interp`:**
  parallel-map desynchronization — a place present in one map and absent from
  another is representable-but-invalid. Use one struct-valued map (D-12-20).
- **Treating the peer-gate register (`peerDivergenceExpected`,
  `session_peer_gate_test.go:43`) as a complete divergence inventory:** the
  probe found it structurally blind to a finding `originvalidate.ValidatePublished`
  surfaces that it never calls (D-12-04a finding 2). Do not use it as the sole
  oracle for "are there any known divergences."
- **Closing D-10-C01's `case core.OpCall` gap as a side effect of this
  phase's work:** explicitly excluded by the probe's BRANCH A/B gate clause
  (b) — filling that existing empty switch arm is legitimate future
  maintenance, but doing it *as part of* Phase 12 without a dedicated,
  reviewed decision would silently expand this phase's interprocedural claim
  beyond what the probe validated.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Payload struct layout (size/alignment/offset) | A second, `cgen`-local layout deriver | `core.RecordLayout`/`core.LayoutField` via `standardForeignLayout()`'s existing pattern (`check.go:3118`) | Already checker-derived, already has a `_Static_assert`/`offsetof` conformance-generation path; a second deriver reintroduces the exact "three competing definitions" ambiguity D-12-25 explicitly rejects. |
| Payload value comparison across engines | A new differential-harness field/axis for "payload value" | The existing `axis:terminal-outcome` `Outcome.Value` comparison (`session_phase5_compare.go:89-96`) | Already includes "the ok payload and the err-edge ADT alternative" per its own comment (D-12-39) — adding a channel that already exists duplicates code and risks the two channels silently disagreeing. |
| Frozen-fixture struct-shape mutation control | A new mutation-runner mechanism | `LayoutMutationRunner` / `control:foreign.layout_mismatch` (`session.go:565-587`) | Already a complete, proven pattern for "commit a fixture whose layout is wrong on purpose, kill it with `_Static_assert`" — D-12-37 clones it rather than inventing a parallel one. |
| Payload-carrying alternative name/detail synchronization | Hand-maintained parallel lists reconciled by convention/review | A single constructor function enforcing name-set equality at construction, plus one `corevalidate` corpus-wide invariant arm (D-12-09) | Protobuf `oneof` and FlatBuffers union-type-vector pairing are both kept in sync by generated accessor code, never by convention — the same reasoning applies here; a name present in one list and absent from the other is representable-but-invalid otherwise. |
| Cross-engine ownership/origin re-derivation for payloads flowing through calls | A new, payload-specific interprocedural analysis | The already-proven Phase 08-11 machinery (`peerDeriveOriginFacts`, `pathoracle.BuildCalleeLookup`/`RecomputeEndpoints`, `originvalidate.ValidatePublished`) | The accepted probe (BRANCH A) is the affirmative finding that payload origin/ownership reduces to this existing machinery — building something new here would contradict the very gate that authorized this phase to proceed. |

**Key insight:** every "don't hand-roll" item above is not a general software-
engineering platitude but a **specific, already-proven mechanism in this exact
codebase** that a planner might be tempted to duplicate under time pressure.
The discussion phase already did the work of identifying and rejecting the
naive alternative for each; the risk this research flags is a plan silently
reintroducing one of them without recognizing it as the same rejected shape.

## Common Pitfalls

### Pitfall 1: Vacuous exhaustive-dispatch coverage

**What goes wrong:** A new `OperationKind` is added to `AllOperationKinds()`
but no fixture in the corpus actually exercises it, so
`TestAllOperationKindsHandledAtEverySite` (`core_test.go:267`) reports "every
kind handled" while the new kind's semantics were never truly driven through
all six sites.

**Why it happens:** The control's own doc comment admits the gap for
`pathoracle`/`originvalidate`: "handled" for these two means only "the walk
completes without error", not a kind-aware switch — so a wrong-but-silent
default path can still report success.

**How to avoid:** Author fixtures that specifically stress
`OpConstructPayload`/`OpDestructurePayload` interacting with existing loan/
origin machinery (per D-12-02's "darkest corner, not average" discipline),
not merely a happy-path single construct-then-match.

**Warning signs:** A payload feature plan whose fixture set contains only one
trivial `data Outcome = | Ok(Buffer) | Err(Fault)` program with no interaction
with borrows, calls, or the affine drop obligation.

### Pitfall 2: Treating "niche optimization" as skippable rather than a stated finding

**What goes wrong:** RES-03's niche clause is read as "optional, so omit it
silently", producing a plan with no mention of niche optimization at all.

**Why it happens:** The clause looks satisfiable-or-not, when the actual
tree-fact is that it is **currently uninstantiable** — `Byte` is plausibly
fully inhabited (every bit pattern is a valid `Byte`), so there is no niche to
exploit with the only two payload types that exist.

**How to avoid:** Per D-12-41 (itself citing D-11-36), a control that cannot
be built must be **escalated as a defect in the criterion**, stated explicitly
in the phase's verification artifacts as an absence-of-applicable-input
finding — not silently treated as "n/a" or, worse, quietly skipped so the
criterion appears satisfied by omission.

**Warning signs:** A `-VERIFICATION.md` or `-VALIDATION.md` for this phase
that does not mention niche optimization anywhere.

### Pitfall 3: Closing D-10-C01 "for free" while fixing something else

**What goes wrong:** A plan working on payload origin/ownership notices
`peerDeriveOriginFacts` has no `case core.OpCall:` and adds one, believing it
is "just completing the switch."

**Why it happens:** The switch genuinely does look incomplete, and the fix is
small (one case arm).

**How to avoid:** The probe's BRANCH A/B gate clause (b) treats "filling an
existing empty switch arm" as in-scope maintenance ONLY when it does not
close a load-bearing gap other decisions depend on remaining open —
specifically, D-12-27's refusal of resource-carrying payloads depends on
D-10-C01 **staying open** as one of its three named landing conditions (along
with D-10-C02 and D-10-C04). Closing D-10-C01 mid-phase without also
resolving D-10-C02 and D-10-C04 would create an inconsistent state where the
landing condition is two-thirds met but the refusal is still in force with no
plan tracking the partial completion.

**Warning signs:** A diff touching `corevalidate.go:2407`'s switch statement
in a Phase 12 plan not explicitly labeled as closing all of D-10-C01/C02/C04
together.

### Pitfall 4: Conflating the D-11-02 emitter deletion with a RES-* criterion

**What goes wrong:** The six-emitter deletion (serving NAT-04..NAT-07) gets
folded into a plan's success criteria phrased as "RES-02 done", making the
phase's actual requirement-shaped criteria (RES-02/RES-03) unverifiable
independently of an inherited-debt-closure task.

**Why it happens:** The deletion is a hard sequencing prerequisite for writing
payload `case`s at all (D-12-32's `emitMatch` collision), so it's tempting to
bundle them.

**How to avoid:** D-12-34 requires the deletion be represented as
explicitly-labeled inherited-debt-closure plans with their own non-
requirement-shaped verification (differential green; digest identity),
sequenced *before* the RES-* plans, never inside them.

**Warning signs:** A plan's `requirements:` frontmatter listing RES-02/RES-03
against tasks whose actual diff is emitter deletion with no payload `case`
arms.

### Pitfall 5: Moving frozen golden-C digests without attribution

**What goes wrong:** The four pinned golden-C digests
(`testdata/phase1/generated.golden.c`, `testdata/phase2/owned_transfer.golden.c`,
`testdata/phase4/foreign_layout_mismatch.golden.c`,
`testdata/phase5/restrict_borrow.golden.c`, per `core_test.go:156-159`) move
in the same diff that introduces payload-related `interp`/`cgen` changes,
making it impossible to tell which change moved which digest.

**Why it happens:** A broad refactor of `interp.go`'s value model or `cgen`'s
emitter family can touch shared code paths these four fixtures also exercise.

**How to avoid:** D-12-18 requires the widening to be evidence-invisible BY
CONSTRUCTION (every scalar-path `tag: ""`, serialized identically to today) —
verify these four digests are byte-identical before and after each plan's
diff, as a named check, not an incidental side effect of running the suite.

**Warning signs:** `git diff` on any of the four `testdata/phase*/...golden.c`
files in a commit whose stated purpose is unrelated to them.

## Code Examples

### Existing value-map site interp.go must widen (representative, not exhaustive — ~17 total per D-12-17)

```go
// Source: internal/compiler/interp/interp.go:357-364 (verified this session)
// newFlatFrame builds a frame for a flat (non-block) linear body: the
// function's own Operations list, walked once, in order.
func newFlatFrame(function core.Function, values map[string]string) frame {
	ops := make(map[string]core.LinearOperation, len(function.Linear.Operations))
	ids := make([]string, len(function.Linear.Operations))
	for i, operation := range function.Linear.Operations {
		ops[operation.ID] = operation
		ids[i] = operation.ID
	}
	return frame{
		function: function, values: values, live: map[string]bool{}, operations: ops, ids: ids,
		placeTypes: placeTypeIndex(function), types: typeFactIndex(function),
	}
}
```
Every call site constructing `map[string]string{...}` (three literal frame
seeds, plus `newFlatFrame`/`newBlockFrame`/`newArmFrame`'s own `values
map[string]string` parameters) is a site the Go compiler will force to update
when `frame.values` becomes `map[string]value` — this is the mechanism D-12-17
relies on for "no silent misses."

### `MatchArm`'s current shape — confirms genuine new binder work is required (D-12-13)

```go
// Source: internal/compiler/core/core.go:861-872 (verified this session)
type MatchArm struct {
	ID      string `json:"id"`
	EdgeID  string `json:"edge_id"`
	Pattern string `json:"pattern"`
	Value   string `json:"value"`
	BlockID string `json:"block_id,omitempty"`
}
```
```go
// Source: internal/compiler/ast/ast.go:111-124 (verified this session)
type MatchArm struct {
	Pattern string
	Value   string
	// Body is the Phase 3 extension: an arm's value position may hold a full
	// linear body instead of a bare alternative name. Exactly one of Value
	// and Body is populated — see HasClosedVariant, the arm-level analog of
	// Body.HasClosedVariant (03-PATTERNS inconsistency I-7).
	Body *LinearBody
	Span diagnostic.Span
}
func (a MatchArm) HasClosedVariant() bool { return (a.Value != "") != (a.Body != nil) }
```
`Value` here is confirmed, by its own doc comment, to be the arm's bare-name
result expression — mutually exclusive with `Body` — not a binder slot of any
kind. A binder field is additive new work on this struct.

### Parser's current one-identifier-then-arrow shape (D-12-13)

```go
// Source: internal/compiler/syntax/parser.go:523-556 (verified this session)
const maxArmsPerMatch = 64

func (p *parser) matchExpr() ast.MatchExpr {
	start := p.expect(TokenMatch, "syntax.expected_match")
	scrutinee := p.identifier("syntax.expected_scrutinee")
	p.expect(TokenLBrace, "syntax.expected_lbrace")
	expression := ast.MatchExpr{Scrutinee: scrutinee.Text, Span: spanFrom(start, scrutinee)}
	for !p.atAny(TokenRBrace, TokenData, TokenFn, TokenEOF) {
		startPosition := p.position
		pattern := p.identifier("syntax.expected_pattern")
		if !p.accept(TokenFatArrow) {
			p.problem("syntax.expected_fat_arrow", p.peek(), "expected `=>`")
			// ... recovery ...
			continue
		}
		if len(expression.Arms) >= maxArmsPerMatch {
			p.problem("syntax.arm_limit", pattern, "match exceeds the declared arm limit")
			// ... recovery ...
			continue
		}
		// ... bare-name-or-body-arm dispatch ...
	}
	p.expect(TokenRBrace, "syntax.expected_rbrace")
	// ...
}
```
`Ok(v) => {...}` requires inserting an optional `( identifier )` production
immediately after `pattern := p.identifier(...)` and before the `=>` check,
with new named refusals for malformed forms (unmatched paren, non-identifier
binder, etc.) — this is the "genuine new parser work" D-12-13 records.

### Probe test's structure — the model for any new committed-test gate this phase needs

```go
// Source: internal/compiler/corevalidate/corevalidate_result_payload_probe_test.go:1-11
// (verified this session)
package corevalidate_test

import (
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/originvalidate"
	"github.com/codename-lang/lang/internal/compiler/pathoracle"
)
```
This file demonstrates the project's committed-single-branch-test pattern
(D-12-01's Q-01 shape): four peers driven **directly**, never through
`session`, so no peer is consulted through another peer. Any new phase-12
committed-test gate (e.g. an N=1 convergence differential per D-12-31, or the
corpus characterization replay per D-12-19) should follow this same direct-
driver shape rather than routing through the CLI or `session`.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|---------------|--------|
| D-04-30's recorded design: ride `OpCopy`/`OpMove` via an additive `alternative_details`/`PayloadSourceID`/`PayloadTargetID` omitempty field pair | Two new `core.OperationKind`s (D-12-05) | This phase's discussion (2026-09-12), superseding the ROADMAP's own canonical ref | The exhaustive-dispatch control (`core_test.go:267`) now genuinely forces new-semantics coverage at all six sites, rather than passing vacuously while an omitempty field silently changes an existing kind's meaning. Cost: ~16-20 edits vs. D-04-22's 8-10-edit single-kind floor. |
| `relay_depth2_accept.lang`'s own fixture header, which claims `corevalidate` refuses it via `core.move_while_borrowed` | `corevalidate.Valid == true` on the unmodified fixture (measured by the probe, this session's re-verification agrees) | Phase 09's D-09-03 closed this deliberately; the fixture header was never updated | A planner reading only the fixture header (not `12-CONTEXT.md` D-12-04c) would misunderstand which peer accepts/refuses this shape. Corrected per D-09-45 (record, don't silently fix) — a Phase 12 plan touching this fixture should update its header. |
| Assuming the differential harness needs extending to compare payload values | `session_phase5_compare.go:89-96` already compares `Outcome.Value` "INCLUDING the ok payload and the err-edge ADT alternative" | Pre-existing (verified this session, not newly added) | D-12-38's decisive anti-vacuity control needs zero harness-extension work; a plan budgeting time for a new comparator channel is over-scoping. |

**Deprecated/outdated:**
- D-04-30's own recorded `Result` payload design (the ROADMAP's canonical ref)
  is explicitly superseded by D-12-05; a planner should not implement it as
  written even though the ROADMAP still cites it as the canonical reference.
- `11-CONTEXT.md`'s `cgen.go` emitter line numbers (152/234/456/649/793/1597/
  1383-1395) are stale after Phase 11's own edits; use D-12-33's re-verified
  ledger (re-confirmed again this session via direct `grep -n` against
  `cgen.go`/`cgen_program.go` — see the Ledger table below).

**`cgen`/`cgen_program.go` emitter line ledger, re-verified this session
(command: `grep -n '^func emit...'`), matching D-12-33 exactly:**

| Emitter / writer | Verified this session |
|---|---|
| `emitMatch` | `cgen.go:152` |
| `emitLinear` | `cgen.go:234` |
| `emitLinearBorrowedByPointer` | `cgen.go:532` |
| `emitLinearBorrowedByPointerPlain` | `cgen.go:726` |
| `emitLinearForeign` | `cgen.go:870` |
| `emitLinearOutputSupport` | `cgen.go:1481` |
| `emitLinearForeignOutputSupport` | `cgen.go:1516` |
| `emitBranch` | `cgen.go:1674` |
| `emitBranchOperations` | `cgen.go:1762` |
| `emitProgram` | `cgen_program.go:128` |
| D-11-52 refusal ("multi-function branch bodies are not supported...") | `cgen_program.go:152` |
| `emitProgramFunction` | `cgen_program.go:339` |

## Assumptions Log

No new `[ASSUMED]` claims were introduced in this research session — every
factual claim above tagged with a file:line citation was independently
re-opened and read (not merely grepped) this session, and matches
`12-CONTEXT.md`'s own already-orchestrator-verified claims exactly, so all
carry `[VERIFIED: <path>:<lines>, re-verified this session]` status. The
table below lists the residual `[ASSUMED]` items that `12-CONTEXT.md` itself
already carries forward as open/unowned — restated here so the planner does
not need to cross-reference back to CONTEXT.md to find them.

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `Byte` is "plausibly fully inhabited" (every bit pattern of an `unsigned char` is a valid `Byte`), making niche optimization currently uninstantiable (D-12-23). | Common Pitfalls §2; Architecture Patterns §4 | If `Byte`'s declared representation later gains a validity constraint (e.g. a range restriction), the niche debt's reopening condition (D-12-24) fires and a real niche mechanism must be built — low risk to this phase since it only affects debt-register wording, not a shipped mechanism. |
| A2 | The N=1 convergence differential (D-12-31 step i) "can be made green cheaply" — this is asserted as the default expectation, with re-deferral as the sole named fallback if it turns out not to hold. | Architecture Patterns §"Recommended Project Structure"; deferred-ideas cross-reference | If the differential turns out expensive, D-12-36's fallback (re-defer with a stated reversal) must fire — this is a scheduling risk, not a correctness risk, and is explicitly pre-planned for. |

**If this table is empty:** N/A — see above; both entries are inherited,
already-flagged uncertainties from `12-CONTEXT.md`, not new assumptions this
research session introduced.

## Open Questions

1. **Exact diagnostic code strings for the three new pattern refusals and the
   resource-payload refusal (D-12-15, D-12-27).**
   - What we know: the established namespacing shape (`check.*` for
     checker-derived semantic refusals, `core.*` for core-IR-derivable facts,
     `syntax.*` for grammar-level refusals) is confirmed by existing examples
     (`core.callee_not_callable`, `check.interprocedural_loan_liveness`,
     `syntax.expected_pattern`, `syntax.arm_limit` — all re-verified this
     session at their cited lines).
   - What's unclear: the exact strings (e.g. `check.payload_arity_mismatch`
     vs. `check.match_arity_mismatch`) are explicitly left to Claude's
     Discretion in `12-CONTEXT.md`.
   - Recommendation: the planner should choose and record these strings in
     the plan itself (they become part of the published diagnostic-ID
     contract once shipped, per this project's stale-reference/D-09-45
     discipline for anything published).

2. **Whether the corpus characterization replay (D-12-19) is one test or a
   table-driven family.**
   - What we know: it must replay the entire pre-Phase-12 fixture corpus and
     assert byte-identical execution documents.
   - What's unclear: table-driven vs. single test is explicitly Claude's
     Discretion.
   - Recommendation: given the existing precedent of `phaseArtifactGlob`-based
     mechanical scans (e.g. `TestDebtRegistersAreWellFormed`,
     `session_test.go:2613`, re-verified this session), a single test that
     globs the corpus and iterates is more consistent with this codebase's
     style than a hand-enumerated table.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | All package builds/tests | ✓ | go1.24.0 darwin/arm64 (verified this session) | — |
| Clang (C17 backend) | `cgen` native emission + `-O3 -flto` differential lane | ✓ | Apple clang 21.0.0 (verified this session) | — |

**Missing dependencies with no fallback:** none.
**Missing dependencies with fallback:** none — both dependencies this phase
needs are already present and match the versions `12-CONTEXT.md`'s reasoning
assumes (notably the `-flto` / Apple clang 21 combination D-12-38's decisive
control depends on).

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go's standard `testing` package (`go test`), project-wide — no third-party test framework anywhere in `internal/compiler` |
| Config file | none — no `pytest.ini`/`jest.config.*` equivalent; Go tests are plain `_test.go` files colocated with the packages they test |
| Quick run command | `go test ./internal/compiler/<package>/... -run <TestName>` (package-scoped, targeted) |
| Full suite command | `go test ./...` (project convention per `PHASE-11-*` and `Phase 12` STATE.md notes — "full suite, `go vet`, and `gofmt` all green" is the standing bar the probe test itself was held to) |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| RES-02 | Payload construct/match/move affine drop obligation holds (unmoved alternative still dropped exactly once, statically) | unit + differential | `go test ./internal/compiler/core/... ./internal/compiler/check/... -run TestAllOperationKindsHandledAtEverySite` plus a new corpus characterization replay | ❌ Wave 0 — new `OpConstructPayload`/`OpDestructurePayload` fixtures and the replay test do not exist yet |
| RES-02 | Payload origin/ownership across `OpCall` reduces to existing Phase 08-11 machinery (no new interprocedural rule) | unit (already committed) | `go test ./internal/compiler/corevalidate/... -run TestC03ResultPayloadOriginAcrossOpCall` | ✅ — `corevalidate_result_payload_probe_test.go`, already green (BRANCH A) |
| RES-03 | Layout — one shared derived fact, correct at all six dispatch sites | unit + `_Static_assert`/`offsetof` conformance | `go test ./internal/compiler/check/... ./internal/compiler/cgen/...` plus a new frozen-fixture mutation test (D-12-37, `control:foreign.layout_mismatch`-style) | ❌ Wave 0 — `testdata/phase12/*.golden.c` and its mutation test do not exist yet |
| RES-03 | "One meaning" = observable-behavior agreement across interp/-O0/-O3(-flto), verified via a reverted production-hunk slot-swap mutation | mutation / differential | The existing five-axis comparator (`session_phase5_compare.go`), driven via a new `cgen`-mutation test analogous to `session.go:565-587`'s `LayoutMutationRunner`, plus a companion unmutated cross-fixture run | ❌ Wave 0 — the mutation harness for this specific `cgen` slot-swap bug does not exist yet |
| RES-03 | Niche optimization stated as an explicit absence-of-applicable-input finding, not silently skipped | debt-register format check | `go test ./internal/compiler/session/... -run TestDebtRegistersAreWellFormed` | ✅ — mechanism exists; the new debt-register *entry* (D-12-24) is Wave 0 content, not new test code |
| — (inherited D-11-02 closure, non-requirement-shaped) | N=1 convergence differential green before six-emitter deletion | differential | A new single-function-program `emitProgram`-vs-legacy-emitter comparison test | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** targeted `go test ./internal/compiler/<touched-package>/...` for the package(s) the task's diff touches, plus `go vet ./...` and `gofmt -l .` (matching the probe's own stated bar: "full suite, `go vet`, and `gofmt` all green").
- **Per wave merge:** `go test ./...` (the full suite) — this project's standing convention, not a phase-specific addition.
- **Phase gate:** Full suite green, plus the four pinned golden-C digests (`core_test.go:156-159`) verified byte-identical (Pitfall 5), before `/gsd-verify-work`.

### Wave 0 Gaps
- [ ] `internal/compiler/core/core_test.go` (or a new `core_payload_test.go`) — fixtures exercising `OpConstructPayload`/`OpDestructurePayload` at all six `TestAllOperationKindsHandledAtEverySite` dispatch sites, per RES-02.
- [ ] A corpus-characterization-replay test (D-12-19) asserting byte-identical execution documents across the entire pre-Phase-12 fixture corpus.
- [ ] `testdata/phase12/` — the frozen struct-shape mutation fixture(s) for D-12-37's `control:foreign.layout_mismatch`-style control, plus its companion `_Static_assert`-killing test.
- [ ] A reverted-production-hunk mutation test for `cgen`'s match-arm/constructor codegen (D-12-38) — the decisive criterion-2 control, plus its companion unmutated cross-fixture run proving the comparator reports agreement absent the bug.
- [ ] The N=1 convergence differential test gating the D-11-02 six-emitter deletion (D-12-31 step i).
- [ ] Framework install: none — `go test` is already the project's only test framework; no new tooling needed.

## Security Domain

`security_enforcement` is `true` (ASVS L1, `block_on: high`) per
`.planning/config.json`. This is a compiler/interpreter/native-codegen phase
with no network, authentication, session, or external-user-input surface in
the conventional web/app sense — the "input" is Lang source text compiled by
the same toolchain that authors it, and the "attacker" model this project has
consistently applied (per Phase 4/10/11's own `*-SECURITY.md` registers) is a
malformed or adversarial *program*, not a network actor.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | No authentication surface in a local compiler CLI. |
| V3 Session Management | no | No session concept exists. |
| V4 Access Control | no | No multi-principal access model. |
| V5 Input Validation | yes | Parser/grammar refusals (`syntax.expected_pattern`, `syntax.arm_limit`, and the three new named refusals D-12-15 adds) plus `check`'s fail-closed admission gates (arity mismatch, binder-on-nullary, missing-binder, resource-payload refusal D-12-27) — every malformed payload-pattern input is refused with a coded diagnostic, never silently coerced or defaulted. |
| V6 Cryptography | no | No cryptographic operation is introduced by this phase. |
| V11 Business Logic / Denial of Service | yes | The existing declared-cap discipline (`maxAlternatives` 4096, `maxArmsPerMatch` 64, `pathoracle.MaxPaths`, `interp.MaxCallDepth`) must be honored — a payload feature must not introduce an unbounded dimension (e.g. unbounded payload nesting/size) without its own declared, fail-closed cap. |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Undefined behavior via reading an inactive C `union` member | Tampering / Information Disclosure | D-12-22's flat-struct-not-union decision already closes this by construction — do not reopen it for a "smaller footprint" optimization. |
| Silent affine-obligation violation (payload dropped twice, or not at all) | Tampering (memory-safety-adjacent, in the resource-lifecycle sense this project treats as a first-class hazard) | Two independent knowers — `check`'s worklist and `corevalidate`'s set-propagation — both derive the static by-function-end accounting property (D-12-29); a plan must not implement only one side. |
| Vacuous verification (a control that reports "pass" without exercising the real hazard) | Tampering (of the evidence, not the program) | D-12-37/D-12-38's two-control design (frozen-fixture + reverted-production-hunk mutation) — this is this project's standing anti-vacuity discipline (D-11-36/D-12-41: an unconstructible control is a defect in the criterion, never a silent skip). |
| Resource-carrying payload escaping affine tracking through an interprocedural gap | Elevation of Privilege (resource-lifecycle sense: a resource escaping its owner's lifetime) | D-12-27's fail-closed refusal — resource-carrying payloads are refused entirely this phase, with a three-part, explicitly-gated landing condition (D-10-C01/C02/C04) before the refusal is ever lifted. |

## Sources

### Primary (HIGH confidence)
- `.planning/phases/12-result-payloads/12-CONTEXT.md` — the phase's own
  locked decision record (D-12-01..D-12-42), itself independently re-verified
  against the tree by the discussion orchestrator on 2026-09-12.
- `.planning/phases/12-result-payloads/12-DISCUSSION-LOG.md` — the alternatives
  considered and rejected at each of the eight decision points.
- `.planning/STATE.md` — the probe verdict (BRANCH A accepted), Phase 10/11
  carry-forward inventory, and the pending-todos entry naming this phase's
  gate.
- Direct `Read`/`grep -n`/`sed -n` inspection of the shipped tree, this
  session: `internal/compiler/core/core.go`, `internal/compiler/core/core_test.go`,
  `internal/compiler/ast/ast.go`, `internal/compiler/syntax/parser.go`,
  `internal/compiler/interp/interp.go`, `internal/compiler/check/check.go`,
  `internal/compiler/corevalidate/corevalidate.go`,
  `internal/compiler/corevalidate/corevalidate_result_payload_probe_test.go`,
  `internal/compiler/cgen/cgen.go`, `internal/compiler/cgen/cgen_program.go`,
  `internal/compiler/session/session_phase5_compare.go`,
  `internal/compiler/session/session.go`, `internal/compiler/session/session_test.go`,
  `internal/compiler/native/native.go`, `internal/compiler/ability/ability.go`.
- `.planning/REQUIREMENTS.md` (RES-02/RES-03 exact wording) and
  `.planning/config.json` (`workflow.nyquist_validation`,
  `workflow.security_enforcement`, `security_asvs_level`, `security_block_on`).

### Secondary (MEDIUM confidence)
- `.planning/research/ARCHITECTURE.md` §7/§6/Stage 9 and
  `.planning/research/SUMMARY.md`/`OPTIONS.md` — cited by CONTEXT.md as the
  canonical design references this phase deviates from or extends; not
  independently re-read line-by-line this session (CONTEXT.md's own citations
  of these are treated as sufficient given the orchestrator's stated
  independent re-verification pass at discussion time).

### Tertiary (LOW confidence)
- None — no WebSearch/training-data-only claims were needed for this
  research; the phase's own prior-art citations (Rust MIR, Swift SIL, C17
  §6.5.2.3, Protobuf/FlatBuffers, `-Z borrowck=migrate`) are already
  attributed and reasoned about in `12-CONTEXT.md`'s decision text, not
  independently re-verified against external sources in this session (no
  claim in this RESEARCH.md rests on them beyond restating CONTEXT.md's own
  already-locked reasoning).

## Metadata

**Confidence breakdown:**
- Standard stack: N/A (no external dependencies) — HIGH confidence there are none, verified this session (`go.mod`/package structure unchanged by this phase's scope).
- Architecture: HIGH — every cited pattern was re-opened and read this session, independently matching `12-CONTEXT.md`'s own already-orchestrator-verified claims.
- Pitfalls: HIGH — each pitfall traces to a specific, cited decision (D-12-05, D-12-23, D-12-28, D-12-34, D-12-18) with its own stated rationale in `12-CONTEXT.md`, not inferred.
- Validation architecture: MEDIUM — the test framework and sampling rate are HIGH confidence (directly observed `go test` convention and existing precedent tests), but the exact Wave 0 test file names/locations are a planning-time decision (Claude's Discretion per CONTEXT.md), so they are recommended, not prescribed.
- Security domain: MEDIUM — ASVS category applicability is a reasoned mapping (this is not a networked/authenticated system), following the same reasoning shape prior phases' `*-SECURITY.md` registers used, not a novel assessment framework.

**Research date:** 2026-09-12
**Valid until:** Until the next tree-changing commit in the packages cited above (this is a fast-moving pre-implementation research artifact for a phase whose planning starts immediately after this file is written — treat as valid for the remainder of Phase 12's planning and execution, not a durable 30-day reference).
