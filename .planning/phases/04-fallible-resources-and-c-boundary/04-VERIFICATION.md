---
phase: 04-fallible-resources-and-c-boundary
verified: 2026-09-05T21:30:00Z
status: gaps_found
score: 6/8 must-haves verified
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: gaps_found
  previous_score: 6/7
  gaps_closed:
    - "corevalidate's independent rederivation walks (checkReleaseOrder's rederive backward walk) are source-blind and must stay defined against a corrupted/adversarial core.Program — the visited-set cycle guard closing the prior CR-01 (unbounded hang on a cyclic ok-edge chain)."
  gaps_remaining: []
  regressions: []
gaps:
  - truth: "checkReleaseOrder's independent rederivation must not silently collapse multiple disagreeing histories converging on an INTERIOR (non-terminal) block into a single arbitrarily-kept edge — the same 'every path must agree' principle already enforced at terminal blocks must also hold one hop earlier, inside rederive's own backward walk (RES-01, D-04-07's own wording: 'independently rederives... and compares')."
    status: failed
    reason: >
      Newly surfaced by this run's fresh code review (04-REVIEW.md new CR-01, distinct from
      the now-closed cycle-hang CR-01) and independently confirmed here by direct code read,
      not accepted on the review's word alone. `checkReleaseOrder` builds `okEdgeInto` as
      `map[string]core.Edge` (corevalidate.go:1206) — a SINGULAR map keyed by `ToBlockID`,
      populated by `okEdgeInto[edge.ToBlockID] = edge` (corevalidate.go:~1212), which silently
      overwrites any earlier "ok"-pattern edge into the same target block with whichever one
      is encountered last while iterating `linear.Edges`. `rederive`'s own backward hop through
      every INTERIOR (non-terminal) block on its way to the entry reads this same singular map
      (`edge, ok := okEdgeInto[currentBlockID]`, corevalidate.go:~1260) — only ever following
      the one kept edge, never independently walking and comparing every incoming "ok" edge
      into an interior block the way the outer per-terminal-block loop now correctly does
      (`incoming := edgesByTo[block.ID]`, a slice, iterated in full at corevalidate.go:~1303-1319,
      closing the ORIGINAL CR-01 from the prior review round). A hand-corrupted core.Program
      declaring two "ok"-pattern edges from two blocks A and A' — representing genuinely
      different completed-acquisition histories — into the SAME interior block T causes the
      discarded edge's entire history to never be rederived or compared at all; if the terminal
      block's actual release list happens to match the KEPT edge's rederived history, the
      corrupted program validates successfully despite the discarded path's disagreeing
      acquisition set never being checked. This is the identical map-collapse defect class the
      prior CR-01 fix's own explanatory comment (corevalidate.go, directly above the now-fixed
      terminal-block loop) states corevalidate exists to prevent ("What this rederivation cannot
      tolerate is SKIPPING the check for a merge point... checking against every incoming edge,
      not skipping the block, is what closes the gap") — but that comment's own fix was applied
      only at the outer terminal-block loop, never inside rederive's own interior-block hop,
      leaving the identical class of gap one level removed. Confirmed untested: no test in
      corevalidate_test.go constructs two "ok"-pattern edges converging on a non-terminal block
      (`TestMergeTerminalBlockDivergentReleaseSetsRefused` and
      `TestMergeTerminalBlockAgreeingChainsAccepted` both add the second edge directly into the
      TERMINAL `success` block, never into an interior `step:N` block). check.go's own honest
      lowering never produces this shape today (confirmed: each step's single successor and
      discard's ok/err edges never both use pattern "ok" into the same target), but corevalidate's
      documented purpose — independent of what check.go could produce — is to stay defined
      against an adversarially hand-constructed core.Program, and this shape defeats it silently
      rather than refusing it.
    artifacts:
      - path: internal/compiler/corevalidate/corevalidate.go
        issue: "okEdgeInto (corevalidate.go:~1206-1212) is a map[string]core.Edge (singular, last-writer-wins), not a map[string][]core.Edge; rederive's interior-block hop (~1260) follows only the single kept edge instead of independently rederiving and comparing every incoming \"ok\" edge into an interior block, unlike the fixed terminal-block loop which now iterates all of edgesByTo[block.ID]"
    missing:
      - "Change okEdgeInto to map[string][]core.Edge, collecting every \"ok\"-pattern edge into a given block instead of overwriting"
      - "Inside rederive, when more than one \"ok\" edge targets the current interior block, independently rederive each candidate edge and require they all produce an identical accumulated release set before continuing the walk; refuse (a new or existing mismatch code) if they disagree"
      - "Add a falsifier constructing two \"ok\"-pattern edges into a shared INTERIOR (non-terminal) block from two blocks with different completed-acquisition histories, asserting Validate refuses it — plus a companion accepting-path test where both incoming edges genuinely agree"
  - truth: "corevalidate must validate core.ForeignContract.Symbol as a syntactically safe C identifier before cgen splices it into generated C source, so a corrupted core.Program cannot inject arbitrary C at the one audited C boundary (FFI-01, D-04-12; phase goal's own \"audited\" C boundary)."
    status: failed
    reason: >
      Newly surfaced by this run's fresh code review (04-REVIEW.md new CR-02) and independently
      confirmed here by direct code read. `corevalidate.go:300` checks only
      `function.ForeignContract.Symbol != ""` — non-emptiness, nothing else. A repo-wide grep
      of corevalidate.go and cgen.go for `validCIdentifier` or `symbol_not_identifier` (or any
      equivalent identifier-shape check) returns zero matches: no code anywhere in the validator
      requires Symbol to match a C-identifier pattern. `cgen.go` then splices `contract.Symbol`
      directly, unescaped, into generated C source in at least three places: `foreignExternName`
      (cgen.go:778, `return "_LANG_" + symbol`), its use building the extern declaration
      (cgen.go:352, `symbolC := foreignExternName(function.ForeignContract.Symbol)` then
      `fmt.Fprintf(&out, "extern %s %s(unsigned char argument);\n\n", resultType, symbolC)`),
      and its use as a call-expression callee (cgen.go:1278 and the call-site Fprintf). Unlike
      every other emitted identifier (type names, function names, parameter names, place names),
      which are routed through `cName`/`cLocal` — both of which strip every character outside
      `[A-Za-z0-9_]` per their documented invariant — `Symbol` reaches generated C completely
      raw. corevalidate.Validate is the one documented, source-blind gate standing between an
      arbitrary caller-supplied core.Program and C source generation (cgen.Emit/EmitNative are
      exported and accept any core.Program); a Symbol containing e.g. a semicolon, parenthesis,
      or full function body would pass Validate today and then be spliced verbatim into the
      extern declaration and call site, injecting arbitrary top-level C source into a file this
      compiler is about to compile and (via session.RunNative) execute. check.go's own parser
      always produces identifier-shaped Symbol values in the honest pipeline, but corevalidate's
      stated purpose is independence from what check.go could produce. No test in cgen_test.go
      or corevalidate_test.go constructs a non-identifier-shaped Symbol, so this gap is untested.
      This is a direct hit against the phase goal's own framing — "one AUDITED C boundary" —
      and against FFI-01/D-04-12's "one authoritative core.ForeignContract" being the sole
      source of truth cgen may trust without inventing or re-checking facts; an unsanitized
      identifier field is exactly a fact cgen trusts without corevalidate having actually
      audited its shape.
    artifacts:
      - path: internal/compiler/corevalidate/corevalidate.go
        issue: "Only a non-empty check exists for ForeignContract.Symbol (line 300); no C-identifier-shape check anywhere in the file"
      - path: internal/compiler/cgen/cgen.go
        issue: "contract.Symbol (via foreignExternName, line 778) is spliced unsanitized into extern declarations (line 352) and call-expression callees (line 1278), unlike every other identifier which is routed through cName/cLocal sanitization"
    missing:
      - "Add a C-identifier-shape check (e.g. ^[A-Za-z_][A-Za-z0-9_]*$) to corevalidate's existing ForeignContract validation block, immediately after the existing core.foreign_contract_missing non-empty check, refusing with a new code (e.g. foreign.symbol_not_identifier) before cgen ever sees the program"
      - "Add a falsifier in corevalidate_test.go mutating a valid foreign-call program's ForeignContract.Symbol to contain a semicolon/parenthesis/newline and asserting refusal with the new code"
      - "Add a regression test in cgen_test.go proving cgen.Emit is never reached with such a Symbol (or independently refuses it if invoked directly, bypassing corevalidate)"
deferred: []
human_verification: []
---

# Phase 4: Fallible Resources and C Boundary Verification Report

**Phase Goal:** A noncopyable resource crosses one audited C boundary while partial
initialization, failure propagation, and cleanup remain defined.
**Verified:** 2026-09-05
**Status:** gaps_found
**Re-verification:** Yes — after third round of gap closure (plan 04-10, commits f875062, 7f3509d, 97a73a8)

## Goal Achievement

### Observable Truths

Re-verified against the previous VERIFICATION.md's 7 must-haves plus one new must-have
derived from this run's fresh code review (04-REVIEW.md, committed a05ae2a), which surfaced
two NEW critical-severity findings distinct from the now-closed prior gap. Each review
finding was independently confirmed here by direct code read (file/line inspection), not
accepted on the review's word alone, before being folded into the truth table.

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | SC1: A fallible two-step (in practice three-step) acquisition releases only initialized resources, exactly once, in reverse order on success and typed failure | ✓ VERIFIED | Regression spot-check: unchanged since prior round; not touched by 04-10. `check.go` still materializes `OpRelease` reverse of completed-acquisition order. |
| 1a | The *independent* rederivation (`corevalidate.checkReleaseOrder`) proves release order without relying on `check`'s own bookkeeping, at TERMINAL merge points (gap from round 2 closed by plan 04-08) | ✓ VERIFIED | Regression spot-check: the per-terminal-block loop iterating every entry of `edgesByTo[block.ID]` (a slice) and calling `rederive` for each, requiring `len(expected) == len(actual)` for every one, is unchanged and confirmed present at corevalidate.go:~1284-1319. |
| 1b | `rederive`'s backward walk stays defined (does not hang) against a corrupted/adversarial `core.Program` with a cyclic "ok"-edge chain (gap from round 3 / prior-round's CR-01, closed by plan 04-10) | ✓ VERIFIED | Direct code read confirms `rederive` (corevalidate.go:~1244) now allocates `visited := make(map[string]bool, len(linear.Blocks))` fresh per call, and its loop opens with `if visited[currentBlockID] { v.check(false, "core.release_order_cyclic", currentBlockID); return nil, false }` before marking `visited[currentBlockID] = true` — a hard refusal, not a silent truncation, matching the doc comment's own reasoning ("a truncated expected... could ACCEPT a corrupted program, which is worse than the hang this guard replaces"). Confirmed present and correctly asserting: `TestCyclicOkEdgeChainRefusedNotHung` (corevalidate_test.go:922, asserts the cyclic case returns `core.release_order_cyclic` and does not hang) and `TestAcyclicChainsStillValidateUnderCycleGuard` (corevalidate_test.go:975, regression-guards every existing acyclic fixture still validates under the new guard). This closes the prior round's gap cleanly. |
| 1c | `checkReleaseOrder`'s independent rederivation must not silently collapse two disagreeing histories converging on an INTERIOR (non-terminal) block into one arbitrarily-kept edge — the "every path must agree" discipline enforced at terminal blocks (1a) must also hold inside `rederive`'s own interior-block hop | ✗ FAILED | New finding this run, surfaced by 04-REVIEW.md's new CR-01 (distinct from the closed cycle-hang CR-01) and independently confirmed by direct code read: `okEdgeInto` (corevalidate.go:~1206) is `map[string]core.Edge` — singular, last-writer-wins (`okEdgeInto[edge.ToBlockID] = edge`) — and `rederive`'s interior-block hop (`edge, ok := okEdgeInto[currentBlockID]`, ~1260) follows only the one kept edge. Two "ok"-pattern edges from different blocks into the same interior block silently discard one path's history rather than independently rederiving and comparing both, unlike the fixed terminal-block loop. Confirmed untested: no fixture in `corevalidate_test.go` constructs two "ok" edges into a non-terminal block. |
| 2 | SC2: Generated C declarations and adapters make target layout, allocator identity, alias/capture, callback retention, and unwind policy inspectable | ✓ VERIFIED | Regression spot-check: unchanged since prior round; `core.ForeignContract` still carries every named obligation, `EmitForeignHeader`/`EmitForeignConformance` unchanged. |
| 2a | Every emitted identifier reaching generated C is sanitized/validated so a corrupted `core.Program` cannot inject arbitrary C at the one audited boundary the phase goal names — the same "no layer can invent a fact the contract does not carry" discipline (FFI-01, D-04-12) must extend to the shape of the Symbol field itself, not just its presence | ✗ FAILED | New finding this run, surfaced by 04-REVIEW.md's new CR-02 and independently confirmed by direct code read: `corevalidate.go:300` checks only `Symbol != ""`. A repo-wide grep of corevalidate.go and cgen.go confirms zero occurrences of any identifier-shape check (`validCIdentifier`, `symbol_not_identifier`, or equivalent). `cgen.go:352`/`:778`/`:1278` splice `contract.Symbol` unsanitized into `extern` declarations and call-expression callees, unlike every other identifier (`cName`/`cLocal` sanitize all others). No test constructs a non-identifier-shaped Symbol. |
| 3 | SC3 first half: Panic cannot cross the ordinary non-unwinding C boundary | ✓ VERIFIED | Regression spot-check: unchanged, not touched by 04-10. |
| 4 | SC3 second half: A foreign nonlocal exit cannot silently bypass Lang cleanup | ✓ VERIFIED | Regression spot-check: unchanged, not touched by 04-10. |
| 5 | SC4: Interpreter and native executions agree on primary failure and cleanup events | ✓ VERIFIED | `go build ./...`, `go vet ./...`, and `env GOCACHE=/tmp/ai-lang-phase4-cache go test ./...` all exit 0 per orchestrator pre-run; `TestCyclicOkEdgeChainRefusedNotHung`/`TestAcyclicChainsStillValidateUnderCycleGuard` presence and assertions confirmed directly in this session. |

**Score:** 6/8 truths verified. Two new gaps surfaced by this run's fresh code review, both
independently confirmed by direct code inspection (not accepted on the review's word alone):
one soundness gap in the independent rederivation's handling of interior merge points (a
narrower instance of the same defect class the prior round's cycle-guard fix addressed one
level out), and one code-injection gap in Symbol validation at the C boundary the phase goal
explicitly calls "audited."

### Deferred Items

None. Both new findings bear directly on this phase's own must-have wording (RES-01/D-04-07's
"independently rederives... and compares," and the phase goal's "one audited C boundary" /
FFI-01's "no layer can invent a fact the contract does not carry") and are not addressed by
any later phase in ROADMAP.md.

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/compiler/corevalidate/corevalidate.go` | Independent, source-blind release-order rederivation, defined against corrupted input at every merge shape | ⚠️ PARTIAL | Cycle-hang guard present and correct (round-3 gap closed); interior-merge collapse (`okEdgeInto` singular map) is a newly-confirmed gap |
| `internal/compiler/core/core.go` | Authoritative `core.ForeignContract` | ✓ VERIFIED | Field set unchanged and complete per prior rounds |
| `internal/compiler/cgen/cgen.go` | All emitted C identifiers sanitized | ✗ GAP | `Symbol` is the one identifier field that bypasses `cName`/`cLocal` sanitization entirely |
| `internal/compiler/corevalidate/corevalidate_test.go` | Falsifiers for every adversarial shape corevalidate claims to defend against | ⚠️ PARTIAL | Cycle falsifiers present (`TestCyclicOkEdgeChainRefusedNotHung`, `TestAcyclicChainsStillValidateUnderCycleGuard`); interior-merge-divergence and non-identifier-Symbol falsifiers absent |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| `corevalidate.checkReleaseOrder` (terminal blocks) | `edgesByTo[block.ID]` | per-incoming-edge independent rederivation | ✓ WIRED | Confirmed at corevalidate.go:~1303-1319 |
| `corevalidate.checkReleaseOrder.rederive` (interior blocks) | `okEdgeInto` | single-edge lookup, no independent-agreement check | ⚠️ NOT WIRED for merge-agreement | Confirmed singular map, last-writer-wins |
| `core.ForeignContract.Symbol` | `corevalidate.linear` | identifier-shape validation | ✗ NOT WIRED | Only non-empty check exists |
| `core.ForeignContract.Symbol` | `cgen.foreignExternName` | direct, unsanitized splice into extern decl and call site | ✓ WIRED (but unsafely — this is the gap) | Confirmed cgen.go:352, :778, :1278 |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Cyclic ok-edge chain refused, not hung | `TestCyclicOkEdgeChainRefusedNotHung` (named test, not full suite) | Present, asserts `core.release_order_cyclic` and bounded wall-clock | ✓ PASS (confirmed present, not independently re-run this session; orchestrator's full `go test ./...` exit 0 covers it) |
| Interior-merge divergent-history falsifier | grep for such a fixture in corevalidate_test.go | Zero matches | ✗ FAIL (absent) |
| Non-identifier Symbol falsifier | grep for such a fixture in corevalidate_test.go / cgen_test.go | Zero matches | ✗ FAIL (absent) |

### Requirements Coverage

| Requirement | Source Plan(s) | Description | Status | Evidence |
|-------------|----------------|--------------|--------|----------|
| SEM-03 | 04-01, 04-02, 04-04, 04-06, 04-07, 04-08 | `Result` propagation and ignored-result rules produce explicit control flow | ✓ SATISFIED | Marked Complete in REQUIREMENTS.md; SC1/SC3 truths verified above; not implicated by either new gap |
| RES-01 | 04-02, 04-05, 04-07, 04-08, 04-10 | Partially initialized noncopyable resources release exactly the completed set, in reverse order | ⚠️ PARTIALLY SATISFIED | SC1 mechanism itself (`check.go`'s forward accumulation, truth 1) is verified; but the *independent rederivation* half of this requirement's own wording ("independently rederives... and compares," D-04-07) has a confirmed unclosed soundness gap at interior merge points (truth 1c) |
| FFI-01 | 04-01, 04-03, 04-05, 04-06, 04-07, 04-09 | Foreign contracts carry target layout, initialized state, allocator identity, capture/retention, aliasing, unwind obligations, all inspectable and non-inventable | ⚠️ PARTIALLY SATISFIED | The contract's field completeness (SC2) is verified; but "no layer can invent a fact the contract does not carry" is undermined by Symbol reaching cgen unsanitized (truth 2a) — corevalidate never actually audits the field's shape before cgen trusts it |

No orphaned requirements: SEM-03, RES-01, FFI-01 are the only IDs REQUIREMENTS.md maps to
Phase 4, and all three appear in at least one plan's `requirements:` frontmatter (04-01
through 04-10 collectively cover all three).

### Anti-Patterns Found

None newly introduced by plan 04-10's changes (visited-set guard, two falsifiers) — scoped,
minimal, and matches the idiom of `blockReach`/`loanChainIndex.carriedLoans` already in the
file. No TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER markers found in the touched region.

### Human Verification Required

None. Both new gaps are independently confirmed by direct code inspection (file/line) with
no ambiguity requiring human judgment — the singular `okEdgeInto` map and the absent
identifier-shape check are objectively present in the code as described.

### Gaps Summary

Round 3's plan 04-10 correctly and cleanly closed the round-2 gap (unbounded hang on a
cyclic "ok"-edge chain in `rederive`'s backward walk) — confirmed by direct code read of the
visited-set guard and its two falsifiers, both present and correctly asserting refusal
without hanging.

However, this round's fresh code review (04-REVIEW.md) surfaced two NEW critical findings
that this verification independently confirms rather than takes on faith:

1. **Interior-merge collapse (new CR-01):** The same class of defect the round-2 fix closed
   for *terminal* blocks (silently accepting a corrupted program by skipping/collapsing a
   merge point instead of checking every incoming path) still exists one hop earlier, inside
   `rederive`'s own backward walk through *interior* (non-terminal) blocks. `okEdgeInto` is a
   singular `map[string]core.Edge` that silently keeps only the last "ok"-pattern edge into a
   given interior block; two edges representing genuinely different completed-acquisition
   histories converging on the same interior block cause one entire history to be silently
   dropped from consideration, rather than being independently rederived and required to
   agree — the exact discipline correctly enforced at the terminal-block level.

2. **Unsanitized Symbol splice (new CR-02):** `core.ForeignContract.Symbol` is checked only
   for non-emptiness by `corevalidate`, then spliced unsanitized into generated C as an
   `extern` declaration and call-expression callee — unlike every other identifier `cgen`
   emits, which is routed through `cName`/`cLocal` sanitization. A corrupted `core.Program`
   with a `Symbol` containing C syntax metacharacters would pass validation and inject
   arbitrary C source into the file this compiler is about to compile and execute — a direct
   violation of the phase goal's own framing of "one AUDITED C boundary."

Both findings bear directly on this phase's own must-have wording and are not deferred to a
later phase. Neither is a regression introduced by 04-10 — both are pre-existing gaps in code
central to RES-01 and FFI-01, now surfaced by a fresh, thorough review pass.

---

_Verified: 2026-09-05T21:30:00Z_
_Verifier: Claude (gsd-verifier)_
