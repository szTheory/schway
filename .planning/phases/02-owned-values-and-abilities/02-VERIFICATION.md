---
phase: 02-owned-values-and-abilities
verified: 2026-09-04T01:05:40Z
status: passed
score: 14/14 must-haves verified
behavior_unverified: 0
overrides_applied: 1
re_verification:
  previous_status: passed
  previous_score: 14/14
  previous_commit: f1206bb
  verified_commit: da75e95
  note: >-
    The previous `passed` verdict was WRONG at the commit it was issued against.
    An independent deep code review subsequently found two source-reachable
    critical defects underneath truths that the prior run had marked VERIFIED,
    and a third latent defect the prior run's differential could not see because
    the "independent" oracle encoded the same wrong law. This re-verification
    reproduces all three defects at f1206bb through the shipped CLI and
    mutation-kills each fix at HEAD.
  gaps_closed:
    - "Truth 7: `lang format` silently emitted corrupted, non-reparsable source at exit 0 for any linear function with a generic return type plus a binding (890bcd0; generator reach proven by 2d98a78)."
    - "Truth 12: evidence.Build discarded canonical-reparse diagnostics and bound a COMPLETE MANIFEST to corrupted bytes (d9b370f)."
    - "Truths 2/8: loan liveness was not transitive; one hop of reborrow admitted an illegal move, and the oracle carried the identical one-hop law so the differential was blind (3d9493a)."
    - "Truth 2: `check` admitted `borrow` with no share gate, so a valid-looking program passed check (exit 0) then failed both engines as a spanless operational failure (exit 3) (2549311)."
    - "Truth 13: control:ownership.move_while_reborrowed asserted at both the session and shipped-CLI layers; nine controls, not eight (3d9493a, da75e95)."
    - "Truth 5: generated-C ordinary-identifier namespace invariant made explicit and test-derived rather than hand-listed (dd9c0a8)."
  gaps_remaining: []
  regressions: []
  prior_run_root_cause: >-
    The prior run treated a passing test NAME as evidence of a behavior. It cited
    TestOwnershipRoundTrip, TestFormatIdempotent, TestOwnershipSequenceExhaustive
    and the evidence suite without ever asking what INPUTS those tests reach, and
    without ever driving the behavior through the shipped `lang` binary. It also
    never mutation-tested any differential, so it could not distinguish "the
    oracle agrees with production" from "the oracle and production share a bug".
deferred: []
human_verification:
  - test: >-
      Ratify the semantic decision that `Buffer` GRANTS the `share` ability.
      Read internal/compiler/ability/ability.go (Buffer row), the mirrored row in
      internal/compiler/corevalidate/deriveAbility, and commit 2549311.
    expected: >-
      The developer confirms Buffer is a shareable-but-noncopyable resource. This
      SUPERSEDES two plan-declared must_haves that now read false as written:
      02-01-PLAN task text "Buffer grants drop/send/escape only", and 02-02-PLAN
      truth "Box<Buffer> and Pair<Byte, Buffer> have drop/send/escape but not
      copy/share". Both plans should be amended or an override recorded.
    why_human: >-
      Which abilities a primitive type grants is a language-design axiom, not a
      derivable fact. A verifier can prove the derivation is coherent under the
      axiom (it is, see Truth 3) but cannot ratify the axiom itself. Note also
      that both admission layers were edited in the same commit by the same
      author, so the production-vs-validator differential cannot cross-check this
      row - by construction it only checks agreement, not correctness.
  - test: >-
      Accept or reject the consequence that `ownership.borrow_requires_share` is
      unreachable from source. Every source-reachable constructor now grants
      share (Byte, Buffer grant it; Box, Pair only conjoin), so no negative
      source fixture can exist for the share ability.
    expected: >-
      The developer accepts defence-in-depth-plus-enumeration as sufficient, or
      requires a non-shareable source constructor so the gate becomes reachable
      and gains a source-level negative control like implicit_noncopy.lang has
      for copy.
    why_human: >-
      CR-01's own text called out "no counterpart [negative control] for share,
      so the gate cannot regress-detect this" as part of the defect. The
      implemented fix closes the exit-3 incoherence but ALSO removes the
      possibility of that control, which is a different resolution from the one
      the review prescribed. My verifier judgment is that this is ACCEPTABLE
      (see Findings), but it is a deliberate weakening relative to review intent
      and deserves an explicit developer sign-off.
  - test: >-
      Confirm that aggregate/generic values being check-only is an intended
      Phase 2 boundary. `lang check testdata/phase2/ability_shapes.lang` passes
      (exit 0) but `lang run --engine=interpreter|native` and `lang evidence` on
      the same file fail with a spanless `tool.run_failed` / `native.tool_failure`
      / `evidence.operation_failed` at exit 3.
    expected: >-
      Developer confirms Box/Pair are type-level-only in Phase 2 (which the plan
      and roadmap wording supports: "reaches typed core", "derive"), and decides
      whether the CLI should say so with a real diagnostic instead of a spanless
      operational failure.
    why_human: >-
      This is the SAME error-taxonomy smell that CR-01 classified as a BLOCKER -
      a program that passes `check` and then dies as a spanless exit 3. It is
      pre-existing (reproduced identically at f1206bb) and out of the declared
      truth set, so it is not a gap; but whether that UX is acceptable is a
      product decision, not a verification one.
---

# Phase 2: Owned Values and Abilities Verification Report

**Phase Goal:** The end-to-end compiler distinguishes cheap implicit copy from explicit transfer and derives independent value abilities coherently.
**Verified:** 2026-09-04T01:05:40Z (commit `da75e95`)
**Status:** passed (developer decisions recorded 2026-09-04)
**Re-verification:** Yes — this REPLACES my own `passed` 14/14 report issued at `f1206bb`, which was wrong.

## Why the previous verdict was wrong

I issued `passed 14/14` at `f1206bb`. An independent deep code review then found
two source-reachable critical defects under truths I had marked VERIFIED. I
reproduced both of them, plus a third, at `f1206bb` in this run using a
worktree build of the pre-fix CLI:

**1. Truth 7 — the formatter emitted corrupted source at exit 0.** Pre-fix binary,
on a linear function with a generic return type plus a binding:

```
$ lang-f1206bb format probe.lang   # exit 0
fn wrap(value: Box<Byte>) -> Box<Byte> {
  let kept =
    valuekept
  }
```

That output does not reparse (`syntax.expected_linear_result` at 155:156). My
prior evidence cited `TestOwnershipRoundTrip`, `TestFormatIdempotent` and
`TestCSTRoundTrip` as passing. They did pass — the property generator only ever
emitted Phase 1 match programs, so no test input ever reached the linear+generic
surface. **I cited a green test without asking what inputs it reaches.**

**2. Truth 12 — evidence bound a complete manifest to corrupted bytes.** I
mutation-tested this rather than trusting the fix: removing only the two guard
blocks from `evidence.build` at HEAD makes
`TestCanonicalRoundTripFailsClosed/canonical_does_not_reparse` and
`/canonical_not_fixed_point` fail with a fully populated manifest
(`ID:evidence:b08bf4b4…`, all digests present) emitted for an unstable canonical
projection. That is exactly the overclaim I had marked VERIFIED.

**3. Truths 2/8 — loan liveness was not transitive, and the oracle was not
independent.** Pre-fix, `let view = borrow code; let review = borrow view;
let delivered = take code; let observed = review` was accepted. I verified this
is now genuinely caught by reverting only the production hunk of 3d9493a in a
scratch worktree at HEAD:

- `go test ./internal/compiler/check -run TestOwnershipSequenceExhaustive` FAILS at
  `length=2 case=330` (`LoanFinalUses OperationIndex:1` vs oracle's `2`).
- the rebuilt CLI accepts `testdata/phase2/reborrow_while_moved.lang` at exit 0,
  where HEAD rejects it at exit 2.

Before 3d9493a the oracle carried the same one-hop law, so `TestOwnershipSequenceExhaustive`
was green on a wrong law. **I never mutation-tested the differential, so I could not
tell "oracle agrees with production" from "oracle and production share a bug".**

## Goal Achievement

### Observable Truths

All fourteen truths were re-derived from the ROADMAP success criteria plus the
seven plans' `must_haves`. Evidence below is from commands executed in this run
at `da75e95` with `GOCACHE=/tmp/ai-lang-reverify`.

| # | Truth | Status | Evidence (executed, not cited) |
|---|---|---|---|
| 1 | A valid owned-buffer program transfers exactly once and has equivalent interpreter/native events. | ✓ VERIFIED | Shipped CLI: `lang --json run --engine=native testdata/phase2/owned_transfer.lang` → `status:pass`, exit 0. Gate lane `lane:owned-native-differential` pass, control `interpreter-o0-o3-owned`, work 3. `testdata/phase2/owned_transfer.golden.c` byte-identical to `91a6206^` (`sha256:8462ff0a…`). |
| 2 | Use after move and move during an active loan reject with stable causes and smallest repairs. | ✓ VERIFIED (upgraded from a false pass) | Four CLI rejections, each exit 2 with a real primary span: `use_after_move` [142,148], `move_while_borrowed` [144,150], `implicit_noncopy`→`ownership.transfer_requires_take` [109,115], `reborrow_while_moved`→`ownership.move_while_borrowed` [238,242] with causes `borrow_created_here` [183,187], `borrow_used_later` [260,266], loan/owner/type details and repair `move_after_last_borrow_use`. Mutation-killed: reverting the production liveness hunk makes the CLI accept the reborrow program at exit 0. |
| 3 | Copy, drop, share, send, and escape derive independently through aggregate/generic shapes. | ✓ VERIFIED (under a changed axiom — see Findings) | `internal/compiler/ability` green in the gate: `TestBoxAllAbilityMasks`, `TestPairAllAbilityMasks`, `TestTypeRefDerivationUsesStructuralCombiner` (spy combiner proves the production Box/Pair path calls the request-local combiner with masks `[30]` then `[30,31]`, distinguishing conjunction from first-child/union/implication), `TestNegativeWitnessPath` (smallest stable path `Box.value → Pair.right → Box.value → Buffer` for the still-denied copy), `TestArbitraryMasksRemainTestPrivate`. `corevalidate.deriveAbility` re-derives per-ability by independent recursive descent. |
| 4 | A canonical Byte program copies implicitly and leaves its source initialized. | ✓ VERIFIED | CLI `run --engine=native testdata/phase2/implicit_copy.lang` → pass, exit 0. `cgen` and `session` packages green in the gate. |
| 5 | Native execution remains strict and bounded while Phase 1 schemas and goldens remain byte-identical. | ✓ VERIFIED | `git diff 91a6206^ --exit-code -- testdata/phase1/generated.golden.c testdata/phase1/evidence.golden.json testdata/phase2/owned_transfer.golden.c` → **exit 0, empty**. Digests: `f3e4fa6b…`, `31d3b0cf…`, `8462ff0a…`. Whole-tree `git diff 91a6206^ --stat -- testdata/` shows exactly two entries: the legitimately-moved phase2 evidence golden and the new `reborrow_while_moved.lang` fixture. Phase 1 verify: `status:pass`, `recomputed_work: 18`, 5 lanes. |
| 6 | Private leaf/Pair masks agree with an independent oracle without production arbitrary masks. | ✓ VERIFIED | `ability.go` exports only `Derive` and `Has`; `abilitySet`, `combineStructural` and `deriver` are unexported (read directly). `TestArbitraryMasksRemainTestPrivate` parses the production file and passes. |
| 7 | Nested owned type syntax is lossless, canonical, recoverable, and bounded. | ✓ VERIFIED (was a false pass) | **Shipped CLI, end to end:** formatted a linear function with a generic return type plus a binding (`Box<Byte>` and `Pair<Byte, Buffer>`, tab-delimited ugly input) → exit 0; output reparses (`lang check` → `status:pass`, exit 0); re-formatting is byte-identical (fixed point); `lang format --check` on the output → exit 0. **Mutation-killed:** reverting only 890bcd0's `format.go` hunk makes `TestGeneratedLinearRoundTrips` fail (`seed=1174080`, "canonical form does not reparse") and `FuzzParseFormat/seed#6,#7` fail — i.e. 2d98a78's generator genuinely reaches the linear+generic surface that was previously untested. |
| 8 | The checker agrees with an independent model on intermediate states and counted work. | ✓ VERIFIED (was a false pass) | `internal/compiler/check` green (6.3s plain, 20.5s race). Independence proven by mutation, not by name: the reverted production law is caught at `length=2 case=330`. The oracle now materializes the derivation relation as an explicit edge set and closes each loan by order-independent fixed-point iteration, where production streams forward once with inherited associations — genuinely different methods, confirmed by reading both. |
| 9 | A source-blind validator rejects forged abilities, ambiguous IDs, and illegal ownership transitions before engines. | ✓ VERIFIED | `internal/compiler/corevalidate` green. Gate lane `lane:owned-core-controls` pass, controls `ability.forged_copy` + `core.duplicate_operation_id`, work 41. The validator's loan set is now transitive independently of production (per-place transitive loan sets over source-blind core operations). |
| 10 | The coordinated source-to-core lie remains a named expected escape. | ✓ VERIFIED | Fresh Phase 2 verify JSON: `"expected_escapes":["escape:coordinated-source-core-lie"]`; grepped all five lanes' `controls` arrays — it appears in none. `TestVerifyPhase2ControlsAndWork` asserts both halves. |
| 11 | Compile/run stdout and stderr are separately bounded and native stdout is strictly decoded. | ✓ VERIFIED | `internal/compiler/native` green (1.9s / 7.9s race). Corrupted/incomplete native output surfaces as `native.tool_failure` exit 3 (observed directly on the aggregate probe), never as a silent pass. |
| 12 | Evidence binds validated owned source/core/C/execution facts without overclaiming. | ✓ VERIFIED (was a false pass) | **Mutation-killed:** deleting only the two `evidence.canonical_unstable` guards makes `TestCanonicalRoundTripFailsClosed` fail in both subcases with a complete manifest bound to an unstable projection. At HEAD, `lang evidence testdata/phase2/owned_transfer.lang` → `status:pass`, `evidence:bfeedd3d…`, exit 0. Gate lane `lane:owned-evidence-bindings` pass, control `evidence.core_mismatch`. |
| 13 | One bounded gate preserves Phase 1 and observes every Phase 2 control with nonzero work. | ✓ VERIFIED | `sh scripts/verify-phase2.sh` → **exit 0**. Phase 1 work **18**; Phase 2 work **52** across 5 lanes, all pass, all nonzero. **Nine** controls observed (see Probe Execution), including the new `control:ownership.move_while_reborrowed`, asserted at BOTH the session layer (`TestVerifyPhase2ControlsAndWork`) and the shipped-CLI layer (`TestVerifyPhase2CLI`, fixed in da75e95). |
| 14 | Targeted test commands fail closed when a requested target is absent. | ✓ VERIFIED | `scripts/verify-phase2.sh:8` runs `assert-go-tests.sh --self-test ./internal/compiler/session TestTogglePipeline TestOwnedBackendMutationIsMismatch TestVerifyPhase2ControlsAndWork` before anything else; the gate exited 0. |

**Score:** 14/14 truths verified (0 present-but-behavior-unverified). All fourteen
now rest on either a CLI-observed behavior or a killed mutant, not on a test name.

### Findings requiring a human decision

#### F1 — `Buffer` now grants `share`: coherent, but it is an axiom change that falsifies plan text

`Buffer`'s primitive ability row changed from `{drop,send,escape}` to
`{drop,share,send,escape}` (2549311), applied in `ability.go` and mirrored in
`corevalidate.deriveAbility`. I verified the consequences rather than accepting them:

- **The incoherence it fixes is real and was source-reachable.** At `f1206bb`,
  `fn observe(buffer: Buffer) -> Buffer { let view = borrow buffer; let kept = take buffer; kept }`
  gave `lang check` → `status:pass` exit 0, then `lang run --engine=interpreter`
  → `tool.run_failed` spanless exit 3. At `da75e95` the same program runs, exit 0.
- **The golden move is exactly and only what was predicted.** `testdata/phase2/evidence.golden.json`
  moved in `core_digest` (`f06bd365…` → `e17fe549…`) and `id`
  (`fe86f7e0…` → `cea9b525…`) and in NOTHING else — `source_digest`, `c_digest`,
  `execution_digests`, schema and `digest_claim` are unchanged, consistent with C
  not encoding abilities. `owned_transfer.golden.c` and both Phase 1 goldens are
  byte-identical to `91a6206^`.
- **OWN-02 verdict: still satisfied, coherently.** The requirement is that
  abilities *derive independently, including through generic and aggregate types*.
  The derivation MECHANISM is untouched: sealed primitive rows, field-wise
  conjunction through Box/Pair via an unexported request-local combiner, an
  exhaustive 32-leaf / 1024-pair private oracle computed from the mask algebra
  rather than the production table, a spy-combiner non-tautology test, and a
  second independent per-ability recursive descent in the source-blind validator.
  One axiom row moved; the derivation propagated it correctly (Box<Buffer> and
  Pair<Byte,Buffer> now grant share by conjunction, and copy remains denied with
  its smallest witness path intact). Nothing about independence degraded.
- **But two plan-declared must_haves now read false as written** (02-01-PLAN:
  "Buffer grants drop/send/escape only"; 02-02-PLAN truth: "Box<Buffer> and
  Pair<Byte, Buffer> have drop/send/escape but not copy/share"). No `overrides:`
  entry exists. This is the human item — a verifier can certify the derivation is
  coherent under the axiom, but cannot ratify the axiom.
- **Independence caveat worth stating plainly:** both admission layers were edited
  in the same commit by the same author. The production-vs-validator differential
  checks *agreement*, not *correctness*, so it structurally cannot cross-check
  this row. That is the same shape as the WR-02 blind spot, even though here it is
  an intentional decision rather than a bug.

#### F2 — the `ownership.borrow_requires_share` gate is unreachable from source: ACCEPTABLE, with sign-off

My verdict: **acceptable, not a gap.** Reasons:

- The gate's *law* is covered non-vacuously: `TestOwnershipSequenceExhaustive`
  runs the full exhaustive alphabet a second time against a synthetic
  non-shareable `TypeFact` that deliberately grants copy (so the copy gate cannot
  mask a missing share gate), and requires production and oracle to agree on
  which borrow is refused first, with asserted diagnostic schema, span, cause
  ordering and zero emitted operations.
- The unreachability claim is itself executable and self-invalidating:
  `TestShareIsUniversallyGrantedAfterBufferShare` enumerates all four
  constructors to depth 4 (>100 shapes) and fails the moment any constructor
  withholds share — at which point the gate becomes reachable and the test tells
  you to add a fixture.
- An unreachable gate is strictly safer than the alternative it replaced (a
  spanless exit 3 from the source-blind validator).

The reason it still needs sign-off: CR-01's own analysis listed "no negative
source control for share, so the gate cannot regress-detect this" as *part of the
defect*, and the implemented fix eliminates the possibility of such a control
rather than adding one. That is a different resolution from the prescribed one and
the developer should knowingly accept it.

#### F3 — aggregate/generic programs pass `check` then die as a spanless exit 3 (pre-existing, in-scope-but-ugly)

Discovered in this run, not in the review. `testdata/phase2/ability_shapes.lang`
— a shipped Phase 2 fixture — gives `lang check` exit 0, but
`lang run --engine=interpreter` → `tool.run_failed` exit 3,
`--engine=native` → `native.tool_failure` exit 3, and `lang evidence` →
`evidence.operation_failed` exit 3. I confirmed the typed core is *valid*
(`corevalidate.Validate` → `Valid:true, Checks:85`), so this is an engine-side
gap, not an admission bug. I reproduced identical behavior with the `f1206bb`
binary, so it is not a regression.

**Not a gap against the truth set:** ROADMAP SC 3 says abilities "derive"; the
02-02-PLAN truth says the fixture "reaches typed core". Execution of aggregates is
not claimed anywhere in Phase 2. Recorded as a human decision only because it is
the same exit-3 taxonomy smell CR-01 called a BLOCKER.

Also noted (INFO): `evidence.canonical_unstable` is correctly raised internally
but is flattened to a generic spanless `evidence.operation_failed` at the CLI, so
the specific fail-closed reason never reaches the user.

### Required Artifacts

| Artifact | Expected | Status | Details |
|---|---|---|---|
| `internal/compiler/syntax/format.go` | Lossless canonical projection | ✓ VERIFIED | Brace context now classified by declaration keyword; CLI fixed-point confirmed; mutation-killed by the property suite. |
| `internal/compiler/syntax/syntax_test.go` | Property/fuzz coverage of the real surface | ✓ VERIFIED | Generator emits linear + generic shapes (`TestGeneratedLinearRoundTrips`, `FuzzParseFormat` seeds #6/#7); proven by mutation, not by name. |
| `internal/compiler/ability/ability.go` | Sealed five-ability derivation | ✓ VERIFIED | Only `Derive`/`Has` exported; Buffer row changed per F1 with copy witness intact. |
| `internal/compiler/check/check.go` | Copy/move/borrow/return lowering | ✓ VERIFIED | Transitive `discoverLoanLastUses` (inherited association); share gate on `borrow` with the copy gate's cause shape. |
| `internal/compiler/corevalidate/corevalidate.go` | Independent fail-closed admission | ✓ VERIFIED | Per-place transitive loan sets over source-blind core ops; independent per-ability descent. |
| `internal/compiler/evidence/evidence.go` | Honest validated binding | ✓ VERIFIED | Canonical round trip asserted (reparse + fixed point) behind a `format`/`parse` seam; mutation-killed. |
| `internal/compiler/cgen/cgen.go` | Runtime-causal C17, closed namespace | ✓ VERIFIED | Prefix-confinement + honest-reservation documented; reserved sets completed with 15 previously omitted emitted identifiers. |
| `internal/compiler/cgen/cgen_names_test.go` | Namespace invariant, derived not restated | ✓ VERIFIED | Fixed-identifier set derived from two disjoint-name programs' common output; confinement + regex invariants asserted directly. |
| `internal/compiler/native/native.go` | Strict bounded native boundary | ✓ VERIFIED | Package green plain and race. |
| `internal/compiler/session/session.go` | Differential and controls | ✓ VERIFIED | Control list and work derived from the control table, not hardcoded indices. |
| `internal/compiler/testsupport/testsupport.go` | Bounded, deadlined test-support spawns | ✓ VERIFIED | Spawn bounds + deadlines (2f2e2e2), guard widened repo-wide (8d87a1a). |
| `testdata/phase2/reborrow_while_moved.lang` | Transitive-liveness negative fixture | ✓ VERIFIED | 15 lines; rejected exit 2 through the shipped CLI; wired as the ninth control. |
| `scripts/verify-phase2.sh` | Phase 1+2 bounded gate | ✓ VERIFIED | Self-test, `go test ./...`, `go test -race ./...`, `go vet ./...`, both corpora, 20-sample observations. Exit 0. |
| `scripts/assert-go-tests.sh` | Exact selector, fail closed | ✓ VERIFIED | `--self-test` is line 8 of the gate and passed. |
| frozen goldens (3) | Byte-identical to `91a6206^` | ✓ VERIFIED | `git diff --exit-code` empty. |

### Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| source | syntax → checker | production frontend | ✓ WIRED | CLI-confirmed on linear+generic and on all four negatives. |
| checker loan tracking | corevalidate loan tracking | two independent transitive laws | ✓ WIRED | Both reject the reborrow program; production inherits forward, validator carries sets. |
| checker | ability derivation | sealed request-local combiner | ✓ WIRED | Spy-combiner test proves the production path calls it with the right masks. |
| formatter | evidence trust boundary | canonical reparse + fixed point | ✓ WIRED | Guard is load-bearing (mutation-killed). |
| emitted C | native execution | exactly-one-site XOR at O0 and O3 | ✓ WIRED | `TestOwnedBackendMutationIsMismatch` PASS: `StatusMismatch`, `ExitCode == 4`, exactly one `native.engine_mismatch`, `Optimizations() == ["-O0","-O3"]`. |
| session control table | shipped CLI verify JSON | nine required control IDs | ✓ WIRED | Both layers assert the same nine (da75e95 closed the eight/nine drift). |

### Data-Flow Trace (Level 4)

| Artifact | Data | Source | Real | Status |
|---|---|---|---|---|
| checked core | abilities/places/operations/loans | bounded canonical source | Yes | ✓ FLOWING |
| interpreter | outcome/events | validated-core replay | Yes | ✓ FLOWING |
| native | outcome/events | terminal C place and operation sites | Yes; exact-site mutation → exit 4 | ✓ FLOWING |
| evidence manifest | digests | validated products behind a canonical-stability guard | Yes | ✓ FLOWING |
| verify lanes | control IDs and work | control table, not literals | Yes; all five lanes nonzero | ✓ FLOWING |

### Behavioral Spot-Checks (all executed this run)

| Behavior | Command | Result | Status |
|---|---|---|---|
| Full gate | `env GOCACHE=/tmp/ai-lang-reverify sh scripts/verify-phase2.sh` | exit 0; Phase 1 work 18, Phase 2 work 52, 9 controls | ✓ PASS |
| Plain + race suites | inside the gate | 24 `ok` package lines across both runs, 0 FAIL | ✓ PASS |
| `go vet ./...` | run independently | exit 0, 0 bytes of output | ✓ PASS |
| Format linear+generic through the shipped CLI | `lang format probe.lang` | exit 0, correct output | ✓ PASS |
| Canonical output reparses | `lang check out1.lang` | `status:pass`, exit 0 | ✓ PASS |
| Canonical output is a fixed point | `lang format out1.lang \| cmp` | identical; `format --check` exit 0 | ✓ PASS |
| Reborrow-then-move rejected | `lang --json check testdata/phase2/reborrow_while_moved.lang` | exit 2, `ownership.move_while_borrowed`, primary span [238,242], two cause spans, one repair | ✓ PASS |
| Exact-one C mutation at O0 AND O3 | `go test ./internal/compiler/session -run TestOwnedBackendMutationIsMismatch` | PASS; `native.engine_mismatch`, exit 4, `["-O0","-O3"]` | ✓ PASS |
| Frozen goldens | `git diff 91a6206^ --exit-code -- <3 goldens>` | exit 0, empty | ✓ PASS |
| Phase 2 evidence golden moved only where predicted | `git diff 91a6206^ -- testdata/phase2/evidence.golden.json` | only `core_digest` + `id` changed | ✓ PASS |
| Owned positives run native | `lang run --engine=native owned_transfer.lang`, `implicit_copy.lang` | both `status:pass`, exit 0 | ✓ PASS |
| Four negatives reject with spans | `lang --json check` ×4 | all exit 2 with real primary spans | ✓ PASS |
| **Mutant kill:** formatter | revert 890bcd0 `format.go` at HEAD | `TestGeneratedLinearRoundTrips` + `FuzzParseFormat/seed#6,#7` FAIL | ✓ PASS |
| **Mutant kill:** evidence guard | delete the two `canonical_unstable` blocks | `TestCanonicalRoundTripFailsClosed` FAILs, both subcases, with a complete manifest | ✓ PASS |
| **Mutant kill:** loan liveness | revert 3d9493a `check.go` hunk | oracle differential FAILs at `length=2 case=330`; rebuilt CLI accepts the illegal program at exit 0 | ✓ PASS |
| **Regression repro at f1206bb** | pre-fix CLI on the same inputs | corrupted format output at exit 0; `borrow buffer` check-pass-then-exit-3 | ✓ PASS |

### Probe Execution

| Probe | Command | Result | Status |
|---|---|---|---|
| `scripts/verify-phase2.sh` | `env GOCACHE=/tmp/ai-lang-reverify sh scripts/verify-phase2.sh` | exit 0 | PASS |

Phase 1 verify: `status: pass`, `recomputed_work: 18`, 5 lanes, evidence `lang.evidence/0`.

Phase 2 verify: `status: pass`, `recomputed_work: 52`, 5 lanes, all pass, all nonzero:

| Lane | Work | Controls |
|---|---:|---|
| `lane:owned-negative-controls` | 4 | `ownership.use_after_move`, `ownership.move_while_borrowed`, `ownership.transfer_requires_take`, **`ownership.move_while_reborrowed`** |
| `lane:owned-core-controls` | 41 | `ability.forged_copy`, `core.duplicate_operation_id` |
| `lane:owned-native-differential` | 3 | `interpreter-o0-o3-owned` |
| `lane:owned-backend-causality` | 2 | `backend.runtime_causality` |
| `lane:owned-evidence-bindings` | 2 | `evidence.core_mismatch` |

**Nine** controls total. `expected_escapes` is exactly
`["escape:coordinated-source-core-lie"]` and it appears in no lane's `controls`.

Fresh 20-sample observations: format p95 58.3 µs, check p95 92.8 µs, interpreter
p95 260 µs, native p95 745 ms, full verify p95 1.42 s. Peak RSS honestly
`unavailable`. Observations, not ratified SLOs.

### Requirements Coverage

| Requirement | Status | Evidence |
|---|---|---|
| OWN-01 | ✓ SATISFIED | Explicit `take`, exactly-once affine enforcement, four named ownership rejections at exit 2 with spans (now including transitive reborrow), independent source-blind admission, interpreter/O0/O3 parity, and an exact-one-site emitted-C mutation forcing exit 4 at both optimization levels. |
| OWN-02 | ✓ SATISFIED (see F1) | Real Box/Pair source shapes reach typed core with per-ability facts; exhaustive 32-leaf / 1024-pair oracle computed from the mask algebra; spy-combiner non-tautology test; second independent per-ability descent in the validator; sealed derivation with no exported arbitrary-mask constructor. The `Buffer`-grants-`share` axiom change propagated coherently and moved exactly the two golden fields it should have. |

No orphaned Phase 2 requirements.

### Anti-Patterns Found

None. Scanned all 19 Go/shell files changed between `91a6206^` and `da75e95` for
`TBD|FIXME|XXX` and `TODO|HACK|PLACEHOLDER|not yet implemented|coming soon` — zero
hits. `t.Skip` across `internal/` and `cmd/`: zero.

### Gaps Summary

No gaps. All fourteen truths hold at `da75e95`, and unlike my previous report each
one now rests on a behavior observed through the shipped `lang` binary or on a
mutant I killed myself. Three items are routed to the developer for decision
(F1 the `Buffer`-grants-`share` axiom and the two plan must_haves it falsifies;
F2 the deliberately unreachable share gate; F3 the pre-existing check-passes-then-
exit-3 behavior for aggregate values), which is why the status is `human_needed`
rather than `passed`.

---

_Verified: 2026-09-04T01:05:40Z at `da75e95`_

_Verifier: the agent (gsd-verifier) — this report supersedes my own incorrect `passed 14/14` report at `f1206bb`_


---

## Decision Resolution (2026-09-04, `3399ddc`)

_Added after the developer ruled on the three items this report raised. The body
above is preserved verbatim as written at `da75e95`._

### 1. OWN-02 under `Buffer` gaining `share` — ACCEPTED as an explicit override

The verifier correctly flagged that this falsifies two plan must_haves
(`02-01-PLAN`, `02-02-PLAN`) with no override recorded. That gap is now closed:
**`02-OVERRIDES.md` OV-02-01** records the decision, the plans it contradicts, the
self-contradiction in the source plans that forced it, both coherent resolutions
considered, and the three consequences accepted — including the verifier's own
caveat that the Buffer row cannot be cross-checked by the differential because both
admission layers changed in one commit.

OWN-02 is judged satisfied: the derivation *mechanism* is untouched and still
independently recomputed; one axiom row moved and propagated correctly.

### 2. Unreachable `borrow_requires_share` gate — ACCEPTED

Kept as defence in depth against `core.ability.share_denied`. Its law is covered by
an exhaustive differential over a synthetic non-shareable fact that grants copy, and
`TestShareIsUniversallyGrantedAfterBufferShare` makes the unreachability claim
self-invalidating. Recorded in OV-02-01 consequence 1.

### 3. F3 (`Box`/`Pair` check but die spanless exit 3) — ACCEPTED as debt

Out of the declared truth set, pre-existing, not introduced by the fix wave.
Recorded as **D-02-09** with the verifier's observation preserved: it is the same
taxonomy smell that made CR-01 a blocker, and a program that type-checks should not
die without a span. Phase 03 remedy is a causal diagnostic, not a wider C backend.

### Verification of the intervening fix

This report was written at `da75e95`. One commit landed after it — `3399ddc`, fixing
a formatter regression that the concurrent deep review caught and that this report
did not (truth #7 again). Re-run at `3399ddc` by the orchestrating agent:
`go test ./...`, `go test -race ./...`, `go vet ./...`, `sh scripts/verify-phase2.sh`
all exit 0; Phase 1 work 18; Phase 2 work 52; nine controls; the three frozen
goldens byte-identical to `91a6206^`; and the regression's own repro now formats
correctly, reparses, and is a fixed point.

Truth #7 is therefore verified at `3399ddc` rather than at `da75e95`. The remaining
open items are recorded in `02-DEBT.md`; none is a Phase 02 truth.

### Standing note on this report's method

This verifier's self-critique — that it read test *names* rather than test *inputs*,
never mutation-tested a differential, and never drove the shipped binary outside the
gate's own corpus — is adopted as a standing rule for later phases and recorded in
`02-DEBT.md` under "Process debt".
