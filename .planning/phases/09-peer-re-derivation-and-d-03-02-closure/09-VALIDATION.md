---
phase: "09"
slug: "peer-re-derivation-and-d-03-02-closure"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-09-10"
---

# Phase 09 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
>
> Every checkbox in this document is deliberately **unchecked**. A box is ticked
> only when execution has supplied the evidence named beside it.
> `nyquist_compliant` and `wave_0_complete` stay `false` until then.

---

## Scope Limitation Recorded Up Front

This phase's differential gates run over **acyclic** call graphs for every
*liveness* fact, because `callgraph`'s three-color DFS refuses cycles with
`core.call_graph_cycle` before any liveness derivation executes. TRU-04's
"recursion" shape is therefore satisfied as a **cycle-peer differential** —
both peers must independently agree on refusal-or-not and on the witness —
never as a liveness differential over an admitted recursive program, which
cannot exist in this language (D-09-22).

No test, comment, fixture, or report in this phase may claim that interprocedural
loan *liveness* was exercised over a recursive call graph.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go 1.24 standard `testing` (`go test`) |
| **Config file** | none — no test-runner config beyond `go.mod` |
| **Quick run command** | `go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/session` |
| **Full suite command** | `go test ./... && go vet ./...` |
| **Estimated runtime** | ~10 s quick; full suite dominated by `check` (~6.4 s) + `corevalidate` (~1.7 s) |

---

## Sampling Rate

- **After every task commit:** `go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/session`
- **After every plan wave:** `go test ./... && go vet ./...`
- **Before the `computeLoanLastUses` deletion lands:** re-run the TRU-04
  differential specifically — this is D-09-10's build-then-delete gate and is
  the one sampling point that is not merely a regression check.
- **Before `/gsd-verify-work`:** full suite green.
- **Max feedback latency:** ~10 s for the quick lane.

---

## Per-Task Verification Map

Task IDs are provisional until the planner assigns them; the requirement,
behavior, and command columns are the binding content.

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 09-0?-01 | ? | 0 | TRU-04 | — | `generateCallGraphCorpus` reachable from `session`'s test package without granting any production import of `check` (D-09-46) | unit | `go test ./internal/compiler/session -run TestNoUndeclaredCheckPeerDivergenceAcrossCorpus` | ❌ Wave 0 — visibility fix required first | ⬜ pending |
| 09-0?-02 | ? | 1 | OWN-07 | — | Peer derives its own liveness bits by forward set-propagation, memoized in the existing postorder loop, never importing `check` or `callgraph` | unit | `go test ./internal/compiler/corevalidate -run 'PeerLiveness\|LoanChainIndex' -v` | ❌ Wave 0 — new test | ⬜ pending |
| 09-0?-03 | ? | 1 | OWN-07 | V5 | `buildLoanChainIndex` consults the callee bit before `parent[TargetID] = SourceID` across an `OpCall`; corrupted-core input still terminates (iterative walk preserved) | unit | `go test ./internal/compiler/corevalidate -run 'LoanChainIndex' -v` | ❌ Wave 0 — new test | ⬜ pending |
| 09-0?-04 | ? | 1 | OWN-07, TRU-04 | — | Both `peerDivergenceExpected` entries retired: `testdata/phase08/twin_a_accept.lang` and `testdata/phase08/relay_depth2_accept.lang` no longer diverge, and the exact-set gate proves it in both directions | integration | `go test ./internal/compiler/session -run TestNoUndeclaredCheckPeerDivergenceAcrossCorpus -v` | ✅ exists — entries must be removed | ⬜ pending |
| 09-0?-05 | ? | 1 | OWN-07 | — | Import independence still holds for the new peer code (`check`, `ast`, `originvalidate`, `callgraph` all still forbidden) | unit | `go test ./internal/compiler/corevalidate -run 'ImportsStayIndependent\|ImportIndependence' -v` | ✅ exists (4 guards) — coverage extends to new code | ⬜ pending |
| 09-0?-06 | ? | 1 | TRU-04 | — | Seeded endpoint-level fault in `corevalidate` diverges the peers **and** `check` (unseamed) is provably unaffected — the companion assertion, without which the gate is vacuous (D-09-25) | mutation/fault injection | `go test ./internal/compiler/corevalidate ./internal/compiler/check -run 'Seam.*StillRefuses\|EndpointFault' -v` | ⚠️ one direction exists (`check_test.go:3322`), mirror is new | ⬜ pending |
| 09-0?-07 | ? | 1 | TRU-04 | — | Cycle-peer differential: self, mutual, and indirect (3+ hop) cycles — both peers agree on refusal and witness | unit | `go test ./internal/compiler/corevalidate -run 'CyclePeer' -v` | ✅ exists — extend to indirect cycles | ⬜ pending |
| 09-0?-08 | ? | 2 | OWN-08 | — | Peer independently re-derives all four `PublishProblemsFor` classes; `TestPeerDoesNotRederiveNarrowedClasses` semantically flips from asserting false agreement to asserting real re-derivation | unit | `go test ./internal/compiler/corevalidate -run 'PeerCallable\|NarrowedClasses' -v` | ✅ exists — semantic flip required | ⬜ pending |
| 09-0?-09 | ? | 2 | OWN-08 | — | D-03-02's interprocedural half refused by **both** layers: an exported borrow-derived return with no declared origin | integration | `go run ./cmd/lang --json check testdata/phase07/relay_escort_witness.lang` | ✅ fixture exists | ⬜ pending |
| 09-0?-10 | ? | 2 | OWN-05 | — | Call-site convention override non-expressible: no grammar production admits it, and no `core.LinearOperation` field can carry it | unit (negative space) | `go test ./internal/compiler/... -run 'ConventionOverrideNotExpressible' -v` | ❌ Wave 0 — new test | ⬜ pending |
| 09-0?-11 | ? | 2 | OWN-05 | V5 | The existing closed-set `ParameterContract.Mode` decode check is the core-level fail-closed control; a hostile non-`owned`/`shared`/`exclusive` value is refused | unit | `go test ./internal/compiler/corevalidate -run 'Mode.*Invalid\|DecodeMode' -v` | ✅ exists — assert, do not add a seam (D-09-35) | ⬜ pending |
| 09-0?-12 | ? | 3 | TRU-04 | — | A **third** gate-eligible metric name for the peer's closure cost curve, with both naming chokepoints widened and proven to agree | unit | `go test ./internal/compiler/session ./internal/compiler/measure -run 'GateEligibleMetricSetsAgreeAcrossChokepoints\|Demote' -v` | ✅ exists — needs 3-way widening (D-09-28) | ⬜ pending |
| 09-0?-13 | ? | 3 | TRU-04 | — | Peer's counted-work growth exponent is linear-or-declared-bounded against **operation** count, with a ratio-stability tripwire | unit | `go test ./internal/compiler/corevalidate -run 'ClosureCostScaling\|GrowthExponent' -v` | ❌ Wave 0 — new test | ⬜ pending |
| 09-0?-14 | ? | 3 | TRU-04 | — | `qlt02_budget_manifest.json` row for the peer metric is well-formed, machine-ratified, and no duplicate `(machine_id, metric)` pair exists | unit | `go test ./internal/compiler/session -run 'QLT02Budget' -v` | ✅ exists | ⬜ pending |
| 09-0?-15 | ? | 4 | OWN-09 | — | Diagnostic **ordering stability**: cross-function diagnostic selection is unchanged pre/post restructure. This is D-09-13's risk in its concrete form and the gate for D-09-49 Q1 | unit | `go test ./internal/compiler/check -run 'OrderingStability\|DiagnosticSelectionOrder' -v` | ❌ Wave 0 — new test, MUST precede deletion | ⬜ pending |
| 09-0?-16 | ? | 4 | OWN-09 | — | Widened pre-deletion shadow run: synthetic programs reach every `ownership.*` code **through the AST-shadow path specifically**, proving the deleted path's refusals are genuinely subsumed rather than merely unreached (D-09-11) | generated property | `go test ./internal/compiler/check -run 'ShadowPathSubsumption' -v` | ❌ Wave 0 — new test, MUST precede deletion | ⬜ pending |
| 09-0?-17 | ? | 4 | OWN-09 | — | `computeLoanLastUses` deleted; `grep` finds no definition or caller; exactly one loan-liveness decision point remains | source assertion | `go test ./internal/compiler/check -run 'SingleLoanLivenessLaw' -v` | ✅ name exists from Phase 08 — semantics extend | ⬜ pending |
| 09-0?-18 | ? | 4 | OWN-09 | — | The three timing-independent codes are untouched: `ownership.use_after_move`, `ownership.borrow_requires_share`, `ownership.transfer_requires_take` still fire from per-binding facts (D-09-47) | unit | `go test ./internal/compiler/check -run 'UseAfterMove\|BorrowRequiresShare\|TransferRequiresTake' -v` | ✅ exists | ⬜ pending |
| 09-0?-19 | ? | 4 | OWN-09 | — | Exhaustive enumeration still green post-restructure | generated property | `go test ./internal/compiler/check -run 'TestOwnershipSequenceExhaustive\|TestBranchSequenceExhaustive' -v` | ✅ exists | ⬜ pending |
| 09-0?-20 | ? | 5 | QLT-07 | — | M001 Phase 3 loan-liveness Nyquist subset closed in this document's section below; `03-VALIDATION.md` receives a pointer line only, frontmatter and checkboxes untouched | doc + source assertion | `go test ./internal/compiler/session -run 'TestDebtRegistersAreWellFormed' -v` | ✅ exists (register format) | ⬜ pending |
| 09-0?-21 | ? | 5 | all | — | `PHASE-09-DEBT.md` well-formed, written at planning time, carrying the items D-09-44 enumerates | unit | `go test ./internal/compiler/session -run TestDebtRegistersAreWellFormed -v` | ✅ exists | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Required Negative Controls

A control that has never been seen to fail is not a control (QLT-08).

| Control | Falsifier | What must fail red first | Status |
|---------|-----------|--------------------------|--------|
| Peer liveness derivation | Seeded endpoint-level fault in `corevalidate` only | Peers diverge, **and** `check` unaffected — both halves asserted | ⬜ pending |
| Peer independence | Attempt to import `check` / `callgraph` from `corevalidate` production source | Import guard test fails | ⬜ pending |
| Peer's four-class `Callable` re-derivation | Mutate an honestly-declared origin into understated / access-mismatched / foreign-omitted | Peer now **refuses** where Phase 07's narrowed peer falsely agreed | ⬜ pending |
| Peer cost bound | A deliberately unmemoized or reordered peer derivation | Growth exponent exceeds its declared bound, or the ratio tripwire fires | ⬜ pending |
| Third chokepoint widening | Widen only one of the two naming chokepoints | `TestGateEligibleMetricSetsAgreeAcrossChokepoints` fails | ⬜ pending |
| Diagnostic ordering stability | Reorder the post-assembly pass relative to lowering | Ordering-stability test fails, naming the moved diagnostic | ⬜ pending |
| Shadow-path subsumption | Delete the interprocedural clause while keeping the deletion | A synthetic program reaching an `ownership.*` code through the shadow path is silently admitted | ⬜ pending |

---

## Wave 0 Requirements

- [ ] Resolve `generateCallGraphCorpus`'s package-visibility gap
      (`internal/compiler/check/costcorpus_test.go:50`, unexported in a `_test.go`)
      so `session`'s test package can consume it — **without** granting any
      production import of `check` (D-09-46). **Blocks TRU-04's vehicle.**
- [ ] A `corevalidate`-side seeded-fault seam plus its companion assertion —
      the direction not already covered by
      `TestVerifyCallableRefusalSeamCheckDisabledCorevalidateStillRefuses`
      (`check_test.go:3322`) (D-09-25).
- [ ] A diagnostic-ordering-stability test in package `check` (D-09-13,
      D-09-49 Q1). **Must be green before the deletion lands.**
- [ ] Synthetic programs reaching every `ownership.*` code specifically through
      the AST-shadow path (D-09-11). **Must be green before the deletion lands** —
      this is what converts "the deleted path's refusals were subsumed" from an
      assertion into a proof.
- [ ] A peer-side cost-scaling harness for the closure curve (its own metric
      name, not `check`'s) (D-09-28).

---

## Mutation-Kill Register

| Control | Mutation | Expected failure | Lands in | Status |
|---------|----------|------------------|----------|--------|
| Peer liveness bit | Force the bit true unconditionally | `twin_a_accept` / `relay_depth2_accept` refuse again; exact-set gate fails | the plan that adds the bit | ⬜ pending |
| Peer `OpCall` consultation | Restore unconditional `parent[TargetID] = SourceID` | Both retired divergence entries reappear as undeclared divergences | same plan | ⬜ pending |
| Companion assertion | Remove the "unseamed peer unaffected" half | Gate passes while proving nothing — caught by review, and recorded as the reason the half exists | same plan | ⬜ pending |
| Peer four-class re-derivation | Revert one class to unconditional agreement | The class's mutation case falsely agrees again | the plan that closes D-07-33 | ⬜ pending |
| Third chokepoint | One-sided widening | Chokepoint-agreement test fails | the plan that adds the metric | ⬜ pending |

---

## Generator Reachability Register

| Generator | Reaches | Provably does NOT reach | Lands in | Status |
|-----------|---------|-------------------------|----------|--------|
| `generateCallGraphCorpus` (5 shapes: chain, diamond, dense, parser-shaped, forward) | Acyclic call graphs of increasing size and sharing; relay depth ≥ 2 | Any cyclic shape (refused before liveness runs); loops and back edges (none exist in the language); arity > 1 | Wave 0 + TRU-04 plan | ⬜ pending |
| Cycle-peer synthetic builder | Self, mutual, and indirect (3+ hop) cycles | Admitted recursive programs (cannot exist) | TRU-04 plan | ⬜ pending |
| AST-shadow-path reachability corpus | Every `ownership.*` code reached through the pre-deletion shadow path | Codes with no shadow-path route (to be enumerated, not assumed) | OWN-09 plan | ⬜ pending |

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| The two peers are genuinely independent *derivations*, not one wearing two names | OWN-07 | Independence is an architectural property; the seeded fault plus companion assertion is the strongest automatable proxy, but a reviewer must still read the two derivations side by side and confirm the forward/backward asymmetry is real | Read `corevalidate`'s new forward propagation against `check.go:579`/`:686`'s backward derivation; confirm no shared helper was extracted and the four import guards still hold |
| The QLT-07 subset closure claims no more than it proves | QLT-07 | Whether a subset closure is honest is a judgement about scope wording, not a testable predicate | Confirm the closure section names only 03-03/03-04/03-05 rows, explicitly excludes OWN-04 and loop-carried liveness, and that `03-VALIDATION.md`'s frontmatter is byte-identical to before |

---

## M001 Phase 3 Debt Closure (loan-liveness subset)

**Status: pending.** This section is the closure vehicle D-09-42 specifies; it
is written but **not** signed off until execution ratifies it.

Per D-09-40a's inventory (run 2026-09-10, before any `09-PLAN.md` was drafted),
all 33 tests named across M001 Phase 3's loan-liveness-scoped rows exist and
pass. Verification command:

```
go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/pathoracle
```

All three packages returned `ok`. The debt is therefore a
**ratification/documentation** debt, not a test-authorship debt.

**Rows claimed closed by this section** (from
`.planning/milestones/M001-phases/03-borrowed-views-and-cfg-lifetimes/03-VALIDATION.md`):

- **Per-Task Verification Map:** 03-03-01, 03-03-02, 03-03-03, 03-04-01,
  03-04-02, 03-04-03, 03-05-01, 03-05-02, 03-05-03
- **Mutation-Kill Register:** path-oracle endpoint differential; path-oracle
  uniform-join falsifier; validator endpoint recomputation; edge-specificity
  fixture pair; counted-work honesty
- **Generator Reachability Register:** `generatedLinearProgram` (extended);
  exhaustive ownership sequence enumeration; path-oracle metamorphic trials

**Explicitly NOT closed, and never to be represented as closed:**

- Every **OWN-04** row (03-06-\*, 03-07-\*) — outside QLT-07's text.
- The **loop-carried loan liveness / loop-exit edge / per-iteration loan
  identity** limitation `03-VALIDATION.md` records up front. The language still
  has no loops, so this is **structurally un-closable today**.
- `03-VALIDATION.md`'s own `nyquist_compliant: false` flag, which **stays
  false** and must not be edited (D-09-41). Because that document's own scope
  limitation can never honestly close, its boolean can never legitimately flip.

**The only permitted edit to the M001 document** is appending one
non-mutating pointer line noting that its loan-liveness subset is superseded by
this section as of its date. Frontmatter and checkboxes are byte-identical
before and after.

---

## Threat Model

| Ref | Threat | Why it matters here | Falsifier |
|-----|--------|---------------------|-----------|
| T-09-01 | The two peers quietly become one derivation with two call sites | This is the phase's named riskiest assumption and the literal site of the D-02-03/D-03-01 hazard the project already paid for once | The seeded endpoint-level fault **plus** its companion assertion; the four import guards |
| T-09-02 | A corrupted or adversarial `core.Program` causes unbounded recursion or a crash in the new peer code | `corevalidate`'s whole role is validating a core it did not produce; `carriedLoans`' existing iterative walk with cycle detection (`corevalidate.go:1160-1181`) is exactly this defense | New peer walks must be iterative from the start; a synthetic corrupted-core case must terminate with a refusal, never a panic |
| T-09-03 | The deletion silently loses coverage the deleted path had | The corpus is ~58 programs averaging ~28 lines; "no fixture regressed" is a weak claim | The widened AST-shadow-path subsumption corpus, green **before** the deletion |
| T-09-04 | A newly-named cost metric is silently demoted to `observed`, producing a gate that can never block | Both naming chokepoints are closed sets keyed by literal metric name; Phase 08 caught this exact trap once (D-08-32) | `TestGateEligibleMetricSetsAgreeAcrossChokepoints`, re-derived for three-way agreement |
| T-09-05 | Peer agreement is asserted where the two mechanisms are structurally incapable of it | Promoting to a shared `core.*` code would assert exactly this; D-09-31 declined | The codes stay divergent; sameness is documented as a relationship, not merged into one identifier |

---

## Security Domain

`workflow.security_enforcement` active, ASVS L1, block on `high`.

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | No | Compiler-internal admission-layer change; no auth surface |
| V3 Session Management | No | Not applicable |
| V4 Access Control | No | Not applicable |
| V5 Input Validation | **Yes (narrow)** | `corevalidate` is a source-blind, adversarial-input-tolerant validator. Its iterative-with-cycle-detection walks are a defense against a corrupted `core.Program` causing unbounded recursion. **Any new peer code must preserve the "never crash on a corrupted core artifact" property** — T-09-02 |
| V6 Cryptography | No | `diagnostic.Error`'s SHA-256 is identity/dedup, not a security boundary |
| V7 Error Handling & Logging | **Yes (narrow)** | Diagnostics are the product surface. A refusal must name its cause without leaking callee body facts — SEM-05 body-blindness holds because only declared `Mode` fields are disclosed |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or a named Wave 0 dependency
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references, including `generateCallGraphCorpus`'s visibility gap
- [ ] No watch-mode flags
- [ ] Feedback latency < 15 s for the quick lane
- [ ] The build-then-delete ordering gate was honored: the peer landed and the differential was green **before** `computeLoanLastUses` was deleted
- [ ] The M001 Phase 3 subset closure section above is ratified, and `03-VALIDATION.md`'s frontmatter is byte-identical to its pre-phase state
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending. This document is the contract to be satisfied, not a
record of satisfaction.
