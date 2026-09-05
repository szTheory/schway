---
phase: 04-fallible-resources-and-c-boundary
verified: 2026-09-05T23:10:00Z
status: gaps_found
score: 7/8 must-haves verified
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: gaps_found
  previous_score: 6/8
  gaps_closed:
    - "checkReleaseOrder's independent rederivation must not silently collapse two disagreeing histories converging on an INTERIOR (non-terminal) block into one arbitrarily-kept edge (RES-01) — closed by plan 04-11."
    - "corevalidate must validate core.ForeignContract.Symbol as a syntactically safe C identifier before cgen splices it into generated C source (FFI-01) — closed by plan 04-12."
  gaps_remaining: []
  regressions: []
gaps:
  - truth: "Every core.ForeignContract string field cgen ever splices into generated C text — not just Symbol — must be validated or escaped before EmitForeignHeader/EmitForeignConformance emit it, so a corrupted OR ordinary source-reachable foreign policy value cannot inject arbitrary top-level C source at the one audited C boundary the phase goal names (FFI-01, D-04-12's own 'no layer can invent a fact the contract does not carry')."
    status: failed
    reason: >
      Newly surfaced by this round's fresh code review (04-REVIEW.md CR-01, distinct from
      the three prior-round CR-01/CR-02 findings, both now closed) and independently
      reproduced here end-to-end, not accepted on the review's word alone. The 04-12 fix
      (`corevalidate.validCIdentifier` / `cgen.validForeignSymbol`) closed the injection gap
      for `ForeignContract.Symbol` only. `cgen.go`'s `EmitForeignHeader` (cgen.go:1332-1335)
      splices three OTHER flat string fields raw into C comments with no validation beyond
      presence: `fmt.Fprintf(&out, "/* allocator: %s */\n", contract.Allocator)`,
      `"/* unwind: %s */\n"`, `"/* nonlocal_exit: %s */\n"`. `corevalidate.go:327` checks only
      `Unwind != "" && NonlocalExit != ""` (non-emptiness); no shape check exists anywhere for
      `Allocator`, and none for comment-safety on any of the three. Unlike `Symbol` (always
      resolved from an identifier token), these three are populated from
      `ast.ForeignPolicy.Value`, which the parser (`parser.go:227-240`) accepts as either a
      bare identifier or a double-quoted string literal, and the lexer's string rule
      (`lexer.go:58-82`) forbids only a literal newline and an unescaped closing quote —
      every other byte, including `*` and `/`, passes through with zero escaping. I
      independently reproduced this twice in this session: (1) a direct `EmitForeignHeader`
      call on a `core.Program` with `ForeignContract.Allocator` set to
      `"*/ int injected(void){return 1;} /*"` produces
      `/* allocator: */ int injected(void){return 1;} /* */` — a live top-level C function
      definition outside any comment; (2) the SAME result reached from ordinary, honest Lang
      source text (`allocator: "*/ int injected(void){return 1;} /*"` inside a `foreign C { }`
      block) through `session.Check` with ZERO diagnostics, then `cgen.EmitForeignHeader` on
      the resulting `checked.Program` — no gate anywhere in the honest pipeline rejects it.
      This is strictly more severe than the now-fixed `Symbol` gap: it needs no corrupted
      `core.Program`, only a `foreign` block any ordinary Lang author can write. This header
      is not descriptive-only output — `session.go` routes `EmitForeignHeader`'s and
      `EmitForeignConformance`'s output directly into `native.Runner.CompileConformanceUnit`,
      which invokes a real C compiler on the generated source; `cgen.ScanForBannedAttributes`
      only searches for a fixed optimizer-attribute token list and does not detect an escaped
      comment or injected function definition. Directly hits FFI-01's own field list
      ("allocator identity ... unwind obligations") and the phase goal's "one audited C
      boundary" framing — corevalidate never actually audits these three fields' shape before
      cgen trusts them. 04-12's own SUMMARY.md explicitly recorded this exact finding as "a
      named, unfixed finding" and "explicitly out of scope" for that plan, but no later Phase
      4 plan closed it, and no later ROADMAP phase (5: native equivalence/optimization/ASan; 6:
      agent feedback protocols) addresses foreign-contract field sanitization — so it is not a
      deferred item, it is an open gap in this phase's own scope.
    artifacts:
      - path: internal/compiler/cgen/cgen.go
        issue: "EmitForeignHeader (cgen.go:1332-1335) splices contract.Allocator, contract.Unwind, and contract.NonlocalExit raw into C comments via fmt.Fprintf(\"%s\"), with no validation, unlike contract.Symbol which is now guarded"
      - path: internal/compiler/corevalidate/corevalidate.go
        issue: "corevalidate.go:327 checks only Unwind/NonlocalExit non-emptiness; no shape/comment-safety check exists for Allocator, Unwind, or NonlocalExit anywhere in the file"
      - path: internal/compiler/syntax/parser.go
        issue: "parser.go:227-240 admits a double-quoted string literal verbatim (minus surrounding quotes) as any foreign-policy value, including allocator/unwind/nonlocal_exit, with no character restriction"
      - path: internal/compiler/syntax/lexer.go
        issue: "lexer.go:58-82's string-literal rule forbids only a literal newline and an unescaped closing quote; every other byte including '*' and '/' passes through unescaped"
    missing:
      - "Add a C-identifier-shape (or minimally comment-terminator-forbidding) check for Allocator, Unwind, and NonlocalExit in corevalidate.linear's ForeignContract validation block, alongside the existing Symbol check, refusing with a new code before cgen ever sees the program"
      - "Add the same check at the source admission boundary in check.go's collectForeignSymbols, so an ordinary Lang author seeing check.foreign_policy_value_unsafe catches this well before corevalidate or cgen"
      - "As defense-in-depth, make cgen.EmitForeignHeader itself refuse (or escape) any of these fields it is about to splice into a comment if it contains '*/', mirroring validForeignSymbol's role as an independent peer guard for the entry points that skip corevalidate.Validate"
      - "Add falsifiers in corevalidate_test.go and cgen_test.go for each of Allocator, Unwind, and NonlocalExit containing a comment-terminator payload, asserting refusal (or at minimum that EmitForeignHeader's output never contains the injected fragment), analogous to the now-passing Symbol falsifiers"
deferred: []
human_verification: []
---

# Phase 4: Fallible Resources and C Boundary Verification Report

**Phase Goal:** A noncopyable resource crosses one audited C boundary while partial
initialization, failure propagation, and cleanup remain defined.
**Verified:** 2026-09-05
**Status:** gaps_found
**Re-verification:** Yes — fourth round, after gap-closure plans 04-11 (RES-01 interior-merge
rederivation) and 04-12 (FFI-01 Symbol C-identifier audit)

## Goal Achievement

### Observable Truths

Re-verified against the third-round VERIFICATION.md's 8 must-haves. Both of that round's
failed truths were independently re-confirmed fixed by direct code read plus running the
specific named regression tests (not the full suite). This round's fresh code review
(04-REVIEW.md, committed 15b71c1) surfaced one new critical finding, which this verification
independently reproduced twice — via a direct `core.Program` mutation and via ordinary honest
Lang source through `session.Check` with zero diagnostics — before accepting it as a truth
failure, per this agent's standing rule not to take a review's word alone.

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | SC1: A fallible acquisition releases only initialized resources, exactly once, in reverse order on success and typed failure | ✓ VERIFIED | Regression spot-check: unchanged since round 3; `check.go` still materializes `OpRelease` reverse of completed-acquisition order. |
| 1a | Independent rederivation proves release order at TERMINAL merge points | ✓ VERIFIED | Regression spot-check: unchanged; per-terminal-block loop over `edgesByTo[block.ID]` confirmed present at corevalidate.go:~1284-1319. |
| 1b | `rederive`'s backward walk stays defined (no hang) against a cyclic "ok"-edge chain | ✓ VERIFIED | Regression spot-check: unchanged; visited-set guard and `TestCyclicOkEdgeChainRefusedNotHung`/`TestAcyclicChainsStillValidateUnderCycleGuard` confirmed present. |
| 1c | `checkReleaseOrder`'s rederivation must not silently collapse two disagreeing histories converging on an INTERIOR block (round-3 gap) | ✓ VERIFIED | Fixed by plan 04-11. Confirmed by direct code read: `okEdgeInto` is now `map[string][]core.Edge` (corevalidate.go:1222), and `rederive`'s interior hop (corevalidate.go:1281-1318) independently rederives every candidate edge and requires `sameReleaseHistory` agreement, refusing with `core.release_order_merge_mismatch` on disagreement. `TestInteriorMergeDivergentHistoriesRefused` (both edge-ordering subtests) and `TestInteriorMergeAgreeingHistoriesAccepted` run in this session and PASS. |
| 2 | SC2: Generated C declarations and adapters make target layout, allocator identity, alias/capture, callback retention, and unwind policy inspectable | ✓ VERIFIED | Regression spot-check: unchanged; `core.ForeignContract` carries every named obligation, `EmitForeignHeader`/`EmitForeignConformance` unchanged in field coverage. |
| 2a | `core.ForeignContract.Symbol` is validated as a syntactically safe C identifier before cgen splices it into generated C (round-3 gap) | ✓ VERIFIED | Fixed by plan 04-12. Confirmed by direct code read: `corevalidate.go:316` calls `validCIdentifier(function.ForeignContract.Symbol)`, refusing with `foreign.symbol_not_identifier`; `cgen.go` independently defines `validForeignSymbol` (cgen.go:789-804) applied in `emitLinearForeign` and `singleForeignFunction`, gating all three exported `EmitForeign*` entry points. `TestForeignSymbolNotIdentifierRefused`, `TestForeignSymbolInjectionNeverReachesGeneratedC`, and `TestForeignEmittersRefuseNonIdentifierSymbolIndependently` run in this session and PASS, including the comment-terminator (`*/`) and semicolon-brace subtests. |
| 2b | Every OTHER `core.ForeignContract` string field spliced into generated C — `Allocator`, `Unwind`, `NonlocalExit` — must be validated the same way `Symbol` now is, closing the SAME injection class the phase goal's "audited C boundary" and FFI-01 require for `Symbol` specifically | ✗ FAILED | New finding this round, surfaced by 04-REVIEW.md's new CR-01 and independently reproduced twice in this session (see Gaps Summary). `cgen.go:1333-1335` splices `Allocator`/`Unwind`/`NonlocalExit` raw into C comments with no shape validation anywhere in the pipeline; this is reachable from ordinary Lang source (a quoted foreign-policy value), not only from a corrupted `core.Program`. |
| 3 | SC3 first half: Panic cannot cross the ordinary non-unwinding C boundary | ✓ VERIFIED | Regression spot-check: unchanged, not touched by 04-11/04-12. |
| 4 | SC3 second half: A foreign nonlocal exit cannot silently bypass Lang cleanup | ✓ VERIFIED | Regression spot-check: unchanged, not touched by 04-11/04-12. |
| 5 | SC4: Interpreter and native executions agree on primary failure and cleanup events | ✓ VERIFIED | `go build ./...` and `go test ./...` both re-run in this session, exit 0, all 20 packages ok (native 9.078s, session 32.471s, rest cached/fast). |

**Score:** 7/8 truths verified. One new gap surfaced by this round's fresh code review,
independently confirmed by two separate reproductions (direct `core.Program` mutation and
end-to-end honest source-to-C pipeline) rather than accepted on the review's word alone.

### Deferred Items

None. The new finding (truth 2b) bears directly on this phase's own goal wording ("one
audited C boundary") and FFI-01's field list, which explicitly names "allocator identity...
and unwind obligations" — the exact fields at issue. No later ROADMAP phase addresses
foreign-contract field sanitization: Phase 5 concerns native optimization equivalence and
ASan/UBSan evidence; Phase 6 concerns agent feedback/query protocols. Neither is a plausible
home for this gap, so it is not deferred.

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/compiler/corevalidate/corevalidate.go` | Independent, source-blind rederivation and full ForeignContract field audit | ⚠️ PARTIAL | Interior-merge rederivation and `Symbol` identifier audit both now correct and tested; `Allocator`/`Unwind`/`NonlocalExit` shape audit still absent (only non-emptiness for Unwind/NonlocalExit; nothing for Allocator) |
| `internal/compiler/core/core.go` | Authoritative `core.ForeignContract` | ✓ VERIFIED | Field set unchanged and complete |
| `internal/compiler/cgen/cgen.go` | All emitted C text (declarations AND comments) sanitized | ✗ GAP | `Symbol` guarded via `validForeignSymbol`; `Allocator`/`Unwind`/`NonlocalExit` still spliced raw into comments at `EmitForeignHeader` (cgen.go:1332-1335) |
| `internal/compiler/corevalidate/corevalidate_test.go` / `internal/compiler/cgen/cgen_test.go` | Falsifiers for every adversarial shape claimed defended | ⚠️ PARTIAL | Interior-merge and Symbol falsifiers present and passing; no falsifier exists for Allocator/Unwind/NonlocalExit comment-injection |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| `corevalidate.checkReleaseOrder.rederive` (interior blocks) | `okEdgeInto` (now plural) | per-candidate independent rederivation + agreement | ✓ WIRED | Confirmed corevalidate.go:1222,1281-1318; tests pass |
| `core.ForeignContract.Symbol` | `corevalidate.linear` | `validCIdentifier` | ✓ WIRED | Confirmed corevalidate.go:316; tests pass |
| `core.ForeignContract.Symbol` | `cgen.emitLinearForeign`/`singleForeignFunction` | `validForeignSymbol` (independent peer guard) | ✓ WIRED | Confirmed cgen.go:326,789-804,1261; tests pass |
| `core.ForeignContract.Allocator`/`Unwind`/`NonlocalExit` | `corevalidate.linear` | shape/comment-safety validation | ✗ NOT WIRED | Only non-empty check for Unwind/NonlocalExit (corevalidate.go:327); none for Allocator; none for comment-safety on any |
| `core.ForeignContract.Allocator`/`Unwind`/`NonlocalExit` | `cgen.EmitForeignHeader` | direct, unsanitized `fmt.Fprintf` splice into C comment | ✓ WIRED (but unsafely — this is the gap) | Confirmed cgen.go:1332-1335; reproduced with a live injected C function definition escaping the comment |
| Ordinary Lang source (`allocator: "..."` string-literal policy value) | `core.ForeignContract.Allocator` | `syntax.parser` → `check.collectForeignSymbols` | ✓ WIRED (source-reachable, no gate) | Reproduced end-to-end via `session.Check` with zero diagnostics on a crafted `.lang` fixture |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|---------------------|--------|
| `cgen.EmitForeignHeader` | `contract.Allocator`/`Unwind`/`NonlocalExit` | `core.ForeignContract` (from `check.Program`, itself from parsed source) | Yes — real, unvalidated, source-controllable string flows directly into generated C text | ⚠️ HOLLOW (validated only by omission — the value flows real but the audit step does not exist) |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Interior-merge divergent-history refusal | `go test ./internal/compiler/corevalidate/ -run TestInteriorMergeDivergentHistoriesRefused` (named test) | PASS, both subtests | ✓ PASS |
| Interior-merge agreeing-history acceptance | `go test ./internal/compiler/corevalidate/ -run TestInteriorMergeAgreeingHistoriesAccepted` (named test) | PASS | ✓ PASS |
| Non-identifier Symbol refused by corevalidate | `go test ./internal/compiler/corevalidate/ -run TestForeignSymbolNotIdentifierRefused` (named test) | PASS, all 9 subtests | ✓ PASS |
| Symbol injection never reaches generated C (cgen independent guard) | `go test ./internal/compiler/cgen/ -run 'TestForeignSymbolInjectionNeverReachesGeneratedC\|TestForeignEmittersRefuseNonIdentifierSymbolIndependently'` (named tests) | PASS | ✓ PASS |
| Allocator comment-terminator payload injects live C via `EmitForeignHeader` on a direct `core.Program` mutation | ad hoc test written and run in this session, then removed | Injected `int injected(void){return 1;}` fragment present, live, outside any comment | ✗ FAIL (confirms the gap) |
| Same injection reachable from ordinary honest Lang source with zero `session.Check` diagnostics | ad hoc test written and run in this session, then removed | `session.Check` returned zero diagnostics; `cgen.EmitForeignHeader` on the resulting `checked.Program` produced the same live injected fragment | ✗ FAIL (confirms the gap is source-reachable, not merely a corrupted-`core.Program` concern) |
| Full build and test suite | `go build ./...`; `go test ./...` | Exit 0; exit 0, all 20 packages ok | ✓ PASS |

### Requirements Coverage

| Requirement | Source Plan(s) | Description | Status | Evidence |
|-------------|----------------|--------------|--------|----------|
| SEM-03 | 04-01, 04-02, 04-04, 04-06, 04-07, 04-08 | `Result` propagation and ignored-result rules produce explicit control flow | ✓ SATISFIED | Truths 1/3/4 verified above; not implicated by the new gap. Note: REQUIREMENTS.md itself already marks SEM-03 unchecked / "Gaps Found" in its status table, predating this round — not a new discrepancy introduced by this verification. |
| RES-01 | 04-02, 04-05, 04-07, 04-08, 04-10, 04-11 | Partially initialized noncopyable resources release exactly once in reverse order | ✓ SATISFIED | The round-3 interior-merge gap (truth 1c) is now closed and independently confirmed with passing named tests; not implicated by the new gap (which concerns FFI-01's fields, not release ordering). |
| FFI-01 | 04-01, 04-03, 04-05, 04-06, 04-07, 04-09, 04-12 | Foreign contracts carry target layout, initialized state, allocator identity, capture/retention, aliasing, unwind obligations, all inspectable and non-inventable | ⚠️ NOT YET SATISFIED | The contract's field completeness (SC2) and the `Symbol` injection gap are both fixed. But the SAME injection class remains open for `Allocator`/`Unwind`/`NonlocalExit` — three of the exact obligation fields FFI-01 itself enumerates ("allocator identity... and unwind obligations") — and is source-reachable, not merely a corrupted-artifact concern. **Note:** REQUIREMENTS.md's status table currently marks FFI-01 `[x] Complete` (set by plan 04-12's completion commit, before this round's review ran); this verification's finding contradicts that marking and it should be reverted to reflect the open gap, consistent with the project's own precedent of reverting premature "Complete" marks (see commit `fdd6bb0`). |

No orphaned requirements: SEM-03, RES-01, FFI-01 are the only IDs REQUIREMENTS.md maps to
Phase 4, and all three appear in at least one plan's `requirements:` frontmatter (04-01
through 04-12 collectively cover all three).

### Anti-Patterns Found

None newly introduced by plans 04-11/04-12's own changes (plural `okEdgeInto`, recursive
`rederive` closure, `validCIdentifier`/`validForeignSymbol`) — scoped, minimal, matches
existing file idiom. No TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER markers found in the
touched region. The gap identified here (`Allocator`/`Unwind`/`NonlocalExit` unsanitized
splice) is pre-existing code from earlier in the phase (04-01/04-03), not new debt
introduced by 04-11/04-12 — both of those plans' own SUMMARY.md files correctly named
it as a known, unfixed, out-of-scope-for-that-plan finding rather than silently omitting it.

### Human Verification Required

None. The new finding is independently confirmed by direct code inspection plus two
separate, reproducible end-to-end test executions (both removed after confirming and
leaving the working tree clean) — no ambiguity requiring human judgment.

### Gaps Summary

Round 4's plans 04-11 and 04-12 correctly and cleanly closed both gaps the prior round's
review surfaced:

1. **Interior-merge collapse (RES-01):** `okEdgeInto` is now `map[string][]core.Edge`, and
   `rederive`'s interior-block hop independently rederives every candidate edge into a merge
   point and requires agreement, refusing with `core.release_order_merge_mismatch` on
   disagreement. Confirmed by direct code read and by running the two named falsifier tests,
   both PASS.
2. **Unsanitized Symbol splice (FFI-01, partial):** `corevalidate.validCIdentifier` and
   cgen's independent `validForeignSymbol` peer guard now refuse a non-identifier-shaped
   `Symbol` before it ever reaches generated C, at both the shared-validator path and the
   three exported `EmitForeign*` entry points that skip `corevalidate.Validate` entirely.
   Confirmed by direct code read and by running three named falsifier tests, all PASS.

However, this round's fresh code review (04-REVIEW.md) surfaced one NEW critical finding
that this verification independently reproduced rather than took on faith:

3. **Sibling field injection (FFI-01, unclosed):** `EmitForeignHeader` still splices
   `Allocator`, `Unwind`, and `NonlocalExit` raw into C comments, with no validation beyond
   presence anywhere in the pipeline. Unlike the now-fixed `Symbol` gap, this is reachable
   directly from ordinary, honest Lang source (a string-literal foreign-policy value) with
   zero diagnostics from `session.Check` — no corrupted `core.Program` is required. I
   reproduced this twice in this session: once via a direct `core.Program` field mutation,
   and once via an actual `.lang` source file carrying a comment-terminator payload in its
   `allocator` policy value, confirming a live, escaped C function definition lands in the
   generated header both ways. This is a genuine, unclosed gap squarely inside this phase's
   own scope — FFI-01 explicitly names "allocator identity" and "unwind obligations" among
   the fields the audited boundary must make trustworthy, and the phase goal's own "one
   audited C boundary" framing is the same framing the now-closed `Symbol` gap was measured
   against. 04-12's own SUMMARY.md already flagged this exact finding as unfixed and
   out-of-scope for that plan; this verification's judgment is that it cannot remain
   out-of-scope for the PHASE as a whole, since no later phase in ROADMAP.md addresses
   foreign-contract field sanitization.

Neither this new gap nor either of the two now-closed gaps is a regression introduced by
04-11/04-12's own changes — all three are pre-existing defects in code central to RES-01 and
FFI-01, surfaced progressively by successive review rounds.

---

_Verified: 2026-09-05T23:10:00Z_
_Verifier: Claude (gsd-verifier)_
