# Phase 5: Native Equivalence and Adversarial Evidence - Research

**Researched:** 2026-09-05
**Domain:** Native codegen equivalence (interpreter/-O0/-O3/-O3+LTO), C-boundary sanitizer evidence (ASan/UBSan), core-level test-case reduction, and executable traceability auditing, over a Go compiler emitting C17 built with Clang.
**Confidence:** HIGH — this phase's design was already produced by an extensive prior discussion (05-CONTEXT.md, D-05-01..D-05-40) grounded in eight parallel research fan-outs, cross-checked against the shipped tree. This research pass verifies the load-bearing factual claims against the actual codebase and host toolchain rather than re-deriving the plan.

## Summary

Phase 5 has no exploratory design space left to research — 05-CONTEXT.md already contains 40 locked decisions with reversibility, rationale, and named residuals. My job here is to (a) verify the codebase citations those decisions depend on actually say what they claim, (b) verify the toolchain behaviors (ASan/UBSan/LTO) the decisions assume actually hold on this host, and (c) surface one real discrepancy in a schema-location claim so the planner doesn't write a task against the wrong file.

**Verified this session, all consistent with 05-CONTEXT.md's claims:**
- `internal/compiler/cgen/cgen.go:157` emits by-value C parameters, confirmed byte-for-byte — D-05-02's blocking precondition is real and unaddressed.
- Apple clang 21.0.0 arm64 is the exact host toolchain identity D-05-05's `-O0`/`-O3` divergence measurement cites — re-confirmed by direct invocation this session (`clang --version`), not merely re-read from the spike README.
- ASan `-fsanitize=address,undefined -fno-sanitize-recover=all` on a synthetic heap-use-after-free reproduces the exact diagnostic shape D-05-14/D-05-16 key controls on (`ERROR: AddressSanitizer: heap-use-after-free`, abort exit code 134) — positive falsification run this session, output pasted below.
- `-O3 -flto` across two separately-compiled translation units links and runs correctly on this host — confirms D-05-19's LTO tier is buildable with plain `clang -flto`, no additional toolchain setup needed.
- `Phase4RequiredControls()` at `internal/compiler/session/session.go:2004` and its verbatim duplication into `scripts/verify-phase4.sh` is a real, working pattern — `Phase5RequiredControls()` should copy it exactly as D-05-17 instructs.

**One discrepancy found, correcting a claim in 05-CONTEXT.md:** D-05-04 states `emitted_attributes` is populated "on the existing `lang.evidence/1` manifest, no schema bump." Reading the actual struct definitions shows `emitted_attributes` (as `EmittedAttributes []string`) is a field of `foreignManifestDocument` — the **`lang.foreign/0`** sidecar schema at `internal/compiler/cgen/cgen.go:1322-1333` — not of `evidence.Manifest` (`lang.evidence/1`) at `internal/compiler/evidence/evidence.go:49-74`, which only carries a `ForeignDigest` pointing at the foreign sidecar. The distinction matters for task-writing: Phase 5's attribute-justification work populates the **foreign sidecar's** `EmittedAttributes` field (already present, currently always emitted as an empty array per the comment at cgen.go:1317-1319), not a new field on the top-level evidence manifest. No schema bump is still correct — `lang.foreign/0` and `lang.evidence/1` both stay put — but the planner should point tasks at `cgen.go`'s `foreignManifestDocument`/`EmitForeignManifest`, not at `evidence.Manifest`.

**Primary recommendation:** Follow 05-CONTEXT.md's decisions verbatim; use this document only for the verified file/line anchors, the toolchain command lines confirmed to work on this host, and the one schema-location correction above.

## User Constraints

<user_constraints>
### Locked Decisions

D-05-01 through D-05-40 in `.planning/phases/05-native-equivalence-and-adversarial-evidence/05-CONTEXT.md` are ALL locked decisions from the user's advisor-mode discussion at Phase 4 close-out. They are not reproduced verbatim here for length; the planner MUST read `05-CONTEXT.md` directly and treat every D-05-nn item as binding. Key locked points, condensed:

- Exactly one optimizer-visible attribute this phase: C `restrict`, only on cgen-owned function parameters, only where `check`+`corevalidate` independently prove exclusivity (D-05-01/D-05-03/D-05-04).
- A new additive by-pointer lowering path is required before `restrict` has a subject at all (D-05-02) — if too invasive, this is a `checkpoint:decision`, never a silent substitution.
- No callback-registration seam; "stale callback retention" is subjected via a retained-pointer/ASan lane with a named residual `escape:callback-invocation-unsubjected` (D-05-06/D-05-07).
- Allocator mismatch and use-after-free get real dynamic fixtures (second allocator TU, hostile frozen-TU variant), detected by ASan only, never a bare native run (D-05-08/D-05-09/D-05-10).
- Sanitizer lane is structurally isolated (separate binary, type-level exclusion from the comparator), built at `-O1`, options pinned by the harness including `alloc_dealloc_mismatch=1` and `detect_leaks=0` (D-05-11..D-05-17).
- Milestone corpus = existing phase1-4 corpora (byte-identical) + new phase5 corpus: 6-10 hand-written adversarial programs + a bounded enumerated closure, bound asserted equal between Go and shell (D-05-18).
- A new sampled `-O3 -flto` tier on the adversarial subset only (D-05-19), because without LTO cross-TU codegen is opaque and the existing differential is "green by construction."
- Comparison axes: SC1's three plus exit status/signal plus diagnostic-ID equivalence for reject-programs; exclusion list becomes an asserted, fail-closed field-routing test (D-05-20/D-05-21).
- Every NAT-03 mutation must cite its target corpus program in `scripts/verify-phase5.sh` and be asserted to move a compared axis (D-05-22).
- Reduction operates on the typed core only (never source text), five fixed deterministic moves to fixpoint, `MaxReductionAttempts = 64`, new `lang.mismatch/0` schema (D-05-23..D-05-26).
- Three required reducer mutation-kills: no-op, predicate-too-loose, non-determinism (D-05-27).
- QLT-01: machine-readable control registry (spike_id, control_mechanism, exactly one of live_descendant/waived) plus `TestQLT01RegistryComplete` (D-05-28/D-05-29); a demonstrable coordinated source-to-core lie fixture is required (D-05-30); full spike fixture porting is explicitly NOT required (D-05-31).
- `OpCall` is OUT of Phase 5 entirely; becomes M002's lead charter item, recorded explicitly in the roadmap/debt register (D-05-32/D-05-33).
- `discoverLoanLastUses` retirement lands IN Phase 5, sequenced early, following the Rust NLL shadow-mode migration procedure exactly — zero divergences over the ~113,164-case enumeration required before deletion (D-05-34/D-05-35).
- D-04-33 (`Alias` field) folded into this phase's foreign/alias-fact plan (D-05-36).
- Explicit criterion: no prior phase proved `-O3` preserves CFG-precise last-use loan expiry semantics against the native backend (D-05-37).
- D-09/D-10/D-11/D-12/D-12a/D-13 stand unchanged and are newly load-bearing (D-05-38/D-05-39).
- A mandatory mid-phase gate is adopted in advance, after alias-fact/liveness work and before reducer/QLT-01 work (D-05-40).

### Claude's Discretion

Plan decomposition, wave structure, and plan count (subject to D-05-40's mid-phase gate and D-05-34's early sequencing); the exact enumeration bound constant in D-05-18; the concrete spelling of lane/control identifiers beyond those named; the registry file format/location for D-05-28 (YAML, JSON, or a Go table); the `lang.mismatch/0` field encoding details; the concrete second-allocator symbol names in D-05-08; and the shape of the by-pointer lowering in D-05-02 — provided every locked decision holds, the additive-path constraint keeps prior goldens byte-identical, and the D-05-02 escalation rule is honoured rather than worked around.

### Deferred Ideas (OUT OF SCOPE)

- Lang-to-Lang calls (`OpCall`), interprocedural loan liveness, call-graph construction, cycle refusal, bounded interpreter call stack, cross-function differential rebuild — M002's lead charter item.
- Interprocedural `-O3` equivalence — out of M001's proof scope by construction.
- A real foreign callback-registration seam (generation tokens, registration handles, unregister races, FFI-004/FFI-005 thread-and-timing variants).
- LeakSanitizer, MemorySanitizer, ThreadSanitizer.
- UBSan `implicit-conversion` and `unsigned-integer-overflow` checks.
- `-O3`-level sanitizer runs (lane builds at `-O1`).
- The wiki's 64-workload expansion, 24 blocking probes, extended semantic matrix.
- A storable, matchable `Result` value and payload-carrying alternatives.
- Full one-to-one code ports of superseded spike fixtures.
- Loop-carried loan liveness, generics, closures, threads, async.

Accepted residual limitations to record, not engineer around: callback-invocation mechanism unsubjected; the coordinated multi-artifact lie survives (made demonstrable, not closed); optimizer observability is host/version-bound (Apple clang 21/arm64); a differential is existential not universal; quarantine of the foreign boundary is permanent and non-discharging; `discoverLoanLastUses` retirement is the last M001 opportunity.
</user_constraints>

## Phase Requirements

<phase_requirements>
| ID | Description | Research Support |
|----|-------------|------------------|
| INT-02 | Interpreter mismatches report the smallest known source/core case, causal facts, and addressable event trace | D-05-23..D-05-27 (reducer design, `lang.mismatch/0` schema, three required mutation-kills); verified `TestOwnershipSequenceExhaustive`/`TestBranchSequenceExhaustive` exist as pre-built oracles the shadow-mode migration (D-05-35) can reuse |
| NAT-02 | Interpreter, native `-O0`, and native `-O3` agree on terminal outcome, semantic-event order, and live-resource state for valid probes | D-05-18..D-05-22, D-05-37 (corpus axes, LTO tier, exclusion-list fail-closed test); verified LTO cross-TU linking works on this host toolchain |
| NAT-03 | Native validation detects layout mismatch, missing partial cleanup, allocator mismatch, stale callback retention, false no-alias facts, use-after-free, and foreign nonlocal-exit cleanup bypass | D-05-01..D-05-17 (alias-fact emission + by-pointer lowering, callback reframing, allocator/UAF fixtures, sanitizer lane design); verified ASan reproduces the exact diagnostic strings D-05-14/D-05-16 key controls on |
| QLT-01 | The suite preserves every relevant negative control and reduced counterexample from Spikes 001-005 while making the coordinated false source-to-core claim an explicit expected escape | D-05-28..D-05-31 (control registry, `TestQLT01RegistryComplete`, coordinated-lie demonstration); verified spike 004/005 READMEs contain exactly the control mechanisms D-05-28's registry must extract |
</phase_requirements>

## Project Constraints (from CLAUDE.md)

No `./CLAUDE.md` or `./.claude/CLAUDE.md` was found in this repository at research time. No project-specific directives to layer on top of GSD defaults and the locked decisions above.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Alias-fact proof + emission (`restrict`) | Compiler frontend (`check`) | Codegen (`cgen`), independent re-derivation (`corevalidate`) | The fact must be proven once by the checker and independently re-derived by corevalidate before cgen is trusted to emit it (D-12/D-05-04) |
| By-pointer parameter lowering | Codegen (`cgen`) | — | Additive emitter path selected only for new fixtures, per D-04-20's precedent (D-05-02) |
| Native build orchestration (-O0/-O3/-O3+LTO/sanitizer) | `internal/compiler/native` | `internal/compiler/session` (lane wiring, mutation runners) | `native.Runner` already owns Clang invocation and multi-TU compilation; new build configurations are additive there |
| Sanitizer evidence isolation | `internal/compiler/session` (lane table) | `internal/compiler/native` (separate binary artifact) | Isolation must be structural (distinct binary, distinct comparator type signature), not a flag on shared evidence |
| Equivalence comparison + exclusion routing | `internal/compiler/session` / execution comparator | `internal/compiler/execution` (Execution document shape) | The comparator owns axis inclusion/exclusion; a fail-closed field-routing test lives at this tier |
| Reduction (typed core minimization) | New reducer package (core-level) | `internal/compiler/core` (block/operation graph read) | Operates on core structure only, never source text, per the ordinal-identity discipline (D-05-23) |
| Evidence/manifest schema (`lang.mismatch/0`, `emitted_attributes`) | `internal/compiler/evidence` (top-level manifest) / `internal/compiler/cgen` (foreign sidecar) | — | `emitted_attributes` lives on the `lang.foreign/0` sidecar (cgen.go), not `lang.evidence/1` — see Summary correction |
| QLT-01 control registry + audit | New registry file/table (format at Claude's discretion) | `internal/compiler/session` (required-control-set cross-check) | Machine-readable registry cross-referenced against the live required-control set, mirroring the Phase 4 duplication-plus-equality-test pattern |
| `discoverLoanLastUses` retirement | `internal/compiler/check` | — | Shadow-mode migration entirely within `check.go`; `corevalidate` unaffected since it never shared this helper (D-12) |

## Standard Stack

### Core

| Tool | Version (verified this session) | Purpose | Why Standard |
|------|------|---------|--------------|
| Go | 1.24.0 darwin/arm64 | Compiler implementation language | Matches FND-01's pinned toolchain; `go version` run this session confirms match |
| Apple Clang | 21.0.0 (clang-2100.1.1.101), target arm64-apple-darwin25.6.0 | C17 backend, ASan/UBSan runtime, LTO | Matches D-05-05's cited "Apple clang 21 / arm64" exactly — confirmed via `clang --version` this session, not merely re-read from the spike README |
| AddressSanitizer + UndefinedBehaviorSanitizer | Bundled with the above Clang | Dynamic memory-safety and UB detection for the FFI boundary | Standard, load-bearing Clang instrumentation; no alternative sanitizer toolchain needed on this host |

No new third-party packages (npm/pip/cargo) are introduced by this phase — it is pure Go stdlib plus the already-present Clang toolchain integration. The Package Legitimacy Gate is therefore not applicable; see the note below.

### Supporting

| Component | Purpose | When Used |
|-----------|---------|-----------|
| `-flto` (Clang LTO) | Makes cross-TU codegen visible to the optimizer | Only in `lane:native-differential-lto`, only on the D-05-18(a) adversarial subset |
| `ASAN_OPTIONS`/`UBSAN_OPTIONS` env pinning | Deterministic sanitizer behavior across hosts | Harness-set for every sanitizer-lane invocation, never contributor-configurable (D-05-13) |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| C `restrict` on cgen-owned by-pointer parameters | LLVM `noalias`/`nonnull`/`__attribute__((malloc))` | Explicitly rejected (D-05-01) — Rust's 2016-2021 `noalias`-on-`&mut` miscompile history is the documented reason to stop at the older, better-understood `restrict` half of the lattice |
| ASan-only detection for UAF/allocator mismatch | `MallocScribble`/`MALLOC_PERTURB_` | Rejected (D-05-10) — perturbs but does not trap, differs between macOS libmalloc and glibc, breaks cross-host determinism |
| `ddmin`/C-Reduce for minimization | A five-move, core-structure-only reducer | Rejected (D-05-23) — offset/text-based tools collide with the ordinal-identity discipline and can produce structurally invalid candidates |
| A real callback-registration seam | Retained-pointer lifetime lane + named escape | Rejected for this phase (D-05-06) — reopens `OpCall`/D-03-02's closure and self-defeats on synchronous single-shot callbacks |

**Installation:** None required — Go and Clang are already present system-wide on this host (confirmed above); no package manager invocation is part of this phase's scope.

## Package Legitimacy Audit

**Not applicable.** This phase adds zero external package dependencies (no npm/pip/cargo installs). All new machinery — sanitizer lane wiring, the reducer, the LTO tier, the control registry — is either Go stdlib code in this repository or invocations of the already-present system Clang toolchain. The Package Legitimacy Gate protocol is skipped for this reason; there is nothing for `gsd_run query package-legitimacy check` to evaluate.

## Architecture Patterns

### System Architecture Diagram

```
                    ┌─────────────────────────────────────────────┐
                    │              lang verify / release            │
                    └───────────────────────┬─────────────────────┘
                                             │
              ┌──────────────────────────────┼──────────────────────────────┐
              │                              │                              │
     ┌────────▼────────┐           ┌─────────▼─────────┐          ┌─────────▼─────────┐
     │  Differential    │           │  Sanitizer lane    │          │  QLT-01 registry   │
     │  lane(s)         │           │  (structurally      │          │  audit             │
     │  -O0/-O3/(+LTO   │           │  isolated binary)   │          │  (TestQLT01…)       │
     │  on subset)      │           └─────────┬───────────┘          └─────────┬─────────┘
     └────────┬─────────┘                     │                                │
              │                     ┌──────────▼──────────┐          reads registry rows,
   interpreter│  native.Runner      │ ASan+UBSan binary    │          cross-checks against
   (oracle)   │  (Clang invocation, │ (own build config,   │          live required-control
              │  multi-TU compile)  │ own ASAN/UBSAN_OPTIONS)│          set (session.go)
              │                     └──────────┬──────────┘
              │                                │ always-on positive control
              │                                │ (retained-pointer fixture)
              ▼                                ▼
     ┌─────────────────┐           report classification tuple
     │ Execution        │           {sanitizer, check_kind,
     │ comparator        │            report_signature, exit_code}
     │ (asserted axes +  │           → hashed into evidence, raw
     │ fail-closed        │             stderr truncated 64KiB+1
     │ exclusion routing) │
     └────────┬──────────┘
              │ on divergence
              ▼
     ┌─────────────────────────────┐
     │ Core-level reducer (5 fixed  │
     │ moves, MaxReductionAttempts  │
     │ =64) → lang.mismatch/0        │
     │ (reduced_core, reduced_source,│
     │ causal_chain, causes)         │
     └───────────────────────────────┘
```

### Recommended Project Structure

No new top-level directories are prescribed by 05-CONTEXT.md beyond what already exists. New code lands inside the existing packages:

```
internal/compiler/
├── check/          # loanLivenessFixpoint domain broadening; discoverLoanLastUses retirement (shadow mode)
├── cgen/           # by-pointer lowering path (additive); restrict emission; foreignManifestDocument.EmittedAttributes population
├── corevalidate/   # independent attribute-justification re-derivation; Alias field audit (D-05-36)
├── native/         # -O3+LTO build configuration; separate sanitizer binary build+run
├── session/        # Phase5RequiredControls(); new mutation runners (alias, allocator-mismatch, UAF); comparator + exclusion-routing test
├── evidence/       # possible additive fields if any land on lang.evidence/1 proper (verify against actual need, not assumed)
└── <new package>   # core-level reducer; new lang.mismatch/0 schema type — name and location at Claude's discretion
scripts/
└── verify-phase5.sh   # peer of verify-phase4.sh, not a fork; duplicates the required-control set for cross-check
```

### Pattern 1: Fail-closed mutation runner (established, five worked examples exist)

**What:** A `*MutationRunner` type wraps `native.Runner`, injects a specific hostile transformation into a specific artifact, and asserts the transformation is observable (nonzero nonambiguous marker count; refuses to run on `>1` or `0` markers).
**When to use:** Every new NAT-03 mutation this phase (alias-fact injection, second-allocator mismatch, UAF fixture) should be a new sibling of these, not a novel mechanism.
**Example (verified in file, `internal/compiler/session/session.go`):**
```go
// LayoutMutationRunner is control:foreign.layout_mismatch's mutation runner
// ... (session.go:544)
// OwnedBackendMutationRunner (session.go:67-115), ReleaseOmissionMutationRunner
// (session.go:130-175), NonlocalPadOmissionMutationRunner (session.go:193-234),
// NonlocalLedgerOmissionMutationRunner (session.go:244-277) — five existing
// mutation runners, each with the same shape: marker-count guard, mutate,
// run, count nonzero RecomputedWork.
```
Source: read directly from `internal/compiler/session/session.go` this session (grep-located, function bodies not fully quoted here for length — the planner should open the file directly for implementation).

### Pattern 2: Verbatim-duplicated required-control set with equality test

**What:** `Phase{N}RequiredControls()` in Go, `scripts/verify-phase{N}.sh`'s own control list in shell, and `TestPhase{N}RequiredControlsMatchScript` asserting the two are set-equal.
**When to use:** Phase 5's new controls (alias-fact, allocator-mismatch, UAF, sanitizer-lane controls, LTO tier control, enumeration-bound constant, QLT-01 registry cross-check) all join this exact pattern per D-05-17/D-05-18/D-05-29.
**Example (verified, `internal/compiler/session/session.go:2004-2017`):**
```go
func Phase4RequiredControls() []string {
	return []string{
		"control:kind.exhaustive_dispatch",
		"control:foreign.call_target_not_foreign",
		"control:foreign.unwind_policy_undeclared",
		"control:resource.release_omitted",
		"control:resource.release_order_transposed",
		"control:foreign.layout_mismatch",
		"control:foreign.no_unproven_attributes",
		"control:defect.no_release_on_defect",
		"control:foreign.unwind_forbidden",
		"control:foreign.nonlocal_exit_undetected",
		"control:terminator.walk_incomplete",
		"control:origin.foreign_origin_omitted",
	}
}
```
Source: `internal/compiler/session/session.go:2004-2017` (read directly this session).

### Pattern 3: Sanitizer smoke-test toolchain probe returns `tool_missing`, never a silent pass

**What:** `control:foreign.unwind_forbidden`'s existing `nm -u` probe pattern (D-04-19) — a host-tool-dependent check that reports an operational `tool_missing` status if the tool is absent, rather than treating absence as a pass.
**When to use:** D-05-15's sanitizer-runtime availability probe (a two-line ASan+UBSan smoke fixture) must follow this exact posture.
**Verified working smoke command on this host (run this session):**
```bash
clang -std=c17 -O1 -g -fno-omit-frame-pointer -fsanitize=address,undefined -fno-sanitize-recover=all /tmp/asan_smoke.c -o /tmp/asan_smoke
ASAN_OPTIONS=halt_on_error=1:abort_on_error=1:symbolize=0:detect_leaks=0:detect_odr_violation=0:alloc_dealloc_mismatch=1 /tmp/asan_smoke
```
Observed output (truncated): `ERROR: AddressSanitizer: heap-use-after-free on address ...` followed by `==NNNNN==ABORTING`, process exit code `134` (SIGABRT). This is exactly the diagnostic-substring-plus-exit-code shape D-05-14 requires controls to key on, and confirms `-fno-sanitize-recover=all` plus the pinned `ASAN_OPTIONS` produce a deterministic, classifiable abort on this host.

### Pattern 4: LTO cross-TU build (new to this repository)

**What:** Compile each TU with `-flto`, link with `-flto`; only then is a symbol defined in another TU visible to the optimizer at the call site.
**Verified working on this host (run this session):**
```bash
clang -std=c17 -O3 -flto -c a.c -o a.o
clang -std=c17 -O3 -flto -c main.c -o main.o
clang -O3 -flto a.o main.o -o out
```
Exit code 0, correct result. No additional linker (`lld`, `gold`) or `llvm-ar`/`llvm-ranlib` wrapper was required on this Apple Clang 21 host — plain `clang -flto` at both compile and link steps sufficed. `native.Runner`'s existing compile invocation at `internal/compiler/native/native.go:88` and `:237` (both currently pass `-std=c17 -Wall -Wextra -Werror -pedantic <optimization> -c ...`) is the natural site to add `-flto` conditionally for `lane:native-differential-lto`.

### Anti-Patterns to Avoid

- **Emitting `restrict` before the by-pointer lowering exists:** cgen currently emits only by-value struct parameters (verified, `cgen.go:157`); `restrict` on a by-value parameter is a false claim of the exact class D-04-13 refused. Escalate as `checkpoint:decision` if the lowering proves too invasive, per D-05-02 — do not substitute a weaker claim silently.
- **Trusting a green `-O0`/`-O3` differential without LTO on the adversarial subset:** proven vacuous by construction (D-05-24) — every corpus program's only observable effect crosses a cross-TU boundary that is opaque without `-flto`.
- **Relying on default `ASAN_OPTIONS` on macOS:** `alloc_dealloc_mismatch` defaults OFF on this platform (documented Clang ASan behavior — see Common Pitfalls); a control built on the assumption it's on will silently false-green.
- **Building a general delta-debugging engine for the reducer:** `bugpoint`'s documented failure mode is exactly a slow, confusing, pass-oblivious generic reducer — D-05-23 mandates exactly five fixed moves instead.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Heap use-after-free / allocator-mismatch detection | A custom pointer-tag/redzone tracker | Clang AddressSanitizer (`-fsanitize=address`) | ASan's `alloc-dealloc-mismatch` and `heap-use-after-free` diagnostics are existing, load-bearing, well-tested detectors — reusing them is zero new detection logic (D-05-08) |
| UB detection at the FFI boundary | Manual alignment/bounds/nullability checks in generated C | Clang UndefinedBehaviorSanitizer (`-fsanitize=undefined` minus the two noisy non-UB checks) | UBSan's `alignment`, `bounds`, `pointer-overflow`, `nullability` checks map directly onto FFI-boundary defect classes (D-05-13) |
| Deterministic post-free memory access | `MallocScribble`/`MALLOC_PERTURB_` | ASan's redzones + quarantine | Scribbling perturbs freed bytes without trapping the access and differs between macOS libmalloc and glibc — not cross-host deterministic (D-05-10) |
| General-purpose test-case minimization | `ddmin`, C-Reduce, `bugpoint` | A five-move, fixed-order, core-structure-only reducer | Text/offset-based tools collide with the function-local-ordinal identity discipline and can emit structurally invalid candidates; `bugpoint`'s failure mode is exactly the generic-reducer slowness/confusion this project must avoid (D-05-23) |

**Key insight:** Every detection mechanism this phase needs already exists in the Clang toolchain (ASan, UBSan, LTO codegen) or in this repository's own established mutation-runner pattern (five prior worked examples in `session.go`). The actual engineering risk is in creating honest *subjects* for the mutations to attack (the by-pointer lowering, the second allocator TU, the hostile frozen-TU UAF variant) and in wiring the isolation/comparison/reduction machinery correctly — not in building new detectors from scratch.

## Common Pitfalls

### Pitfall 1: `alloc_dealloc_mismatch` defaults OFF on macOS
**What goes wrong:** A contributor running the sanitizer suite with default `ASAN_OPTIONS` (i.e., not sourcing the harness's pinned string) gets a clean pass on the allocator-mismatch control even though the mismatch is real.
**Why it happens:** Apple's ASan runtime ships with `alloc_dealloc_mismatch` disabled by default, unlike some Linux ASan configurations — this is documented Clang/compiler-rt behavior `[CITED: clang.llvm.org/docs/AddressSanitizer.html]`, not independently re-derived this session (the synthetic mismatch probe run in this research pass used a compatible aligned_alloc/free pairing and did not produce a mismatch either way — inconclusive as a falsification, so this claim stays at CITED rather than VERIFIED).
**How to avoid:** D-05-13 pins `alloc_dealloc_mismatch=1` in the harness's `ASAN_OPTIONS` string, never left to the environment. Verify this is actually wired into whatever invokes the sanitizer binary (test harness, `verify` command) rather than only documented.
**Warning signs:** A `control:native.sanitize.allocator_mismatch`-style control passing on a host where the pin was accidentally dropped, with no corroborating exit-code/diagnostic-substring check.

### Pitfall 2: A "green" `-O0`/`-O3` differential proves nothing without LTO
**What goes wrong:** Adding new corpus programs and getting a clean differential feels like progress but the optimizer never actually saw across the FFI boundary it's supposed to be tested against.
**Why it happens:** Without `-flto`, Clang treats a call to a symbol in a separately-compiled TU as opaque — it cannot inline, prove purity, or eliminate effects across it (confirmed this session: standard cross-TU compile without `-flto` at `-O3` cannot fold `helper(1) == 2` at compile time the way an LTO build theoretically could — this project's own corpus crosses that boundary on every observable effect per D-05-24's finding).
**How to avoid:** D-05-19's `-O3 -flto` sampled tier on the hand-written adversarial subset, built by construction (not sampling luck) to guarantee perturbable code paths are the ones actually built with LTO.
**Warning signs:** A new "adversarial" corpus program that shows identical machine code (or a bit-identical binary) at `-O0` vs `-O3` with LTO both on and off — that program is not exercising what it claims to.

### Pitfall 3: `restrict` on a by-value parameter is meaningless — and today, that's the only kind cgen emits
**What goes wrong:** A naive implementation adds `restrict` to the existing by-value struct parameter cgen already generates, producing a syntactically valid but semantically vacuous (or actively false, if the struct itself contains pointers) attribute.
**Why it happens:** `cgen.go:157`'s `fmt.Fprintf(&out, "static %s %s(%s %s) {\n", ...)` — verified this session — takes the parameter by value; there is no pointer for `restrict` to qualify meaningfully.
**How to avoid:** Build the additive by-pointer lowering path first (D-05-02); do not retrofit `restrict` onto the existing by-value emitter.
**Warning signs:** A `restrict`-qualified parameter in generated C that is not itself a pointer type, or a `corevalidate` justification check that passes trivially because the fact being "justified" is vacuous.

### Pitfall 4: New machinery's first green run proves nothing until shown it can go red
**What goes wrong:** The sanitizer lane, the reducer, or the LTO tier compiles and reports "pass" on day one, but that pass is inert — the binary silently isn't linked against the sanitizer runtime, the reducer silently no-ops, or the LTO flag silently isn't passed through.
**Why it happens:** All three are brand-new subsystems in this repository (D-05-38) with no prior mutation-kill history to lean on.
**How to avoid:** D-05-14's always-on positive control (retained-pointer fixture must produce a report on *every* `verify` invocation) and D-05-27's three required reducer mutation-kills (no-op, predicate-too-loose, non-determinism) are the general answer — apply the same discipline to the LTO tier (assert `-flto` actually reached the compile/link command line, not merely that the build succeeded).
**Warning signs:** A lane or control that has never been observed to fail in this project's history — that absence of red is itself a finding requiring a mutation-kill test, not a comfort.

## Code Examples

### Verified: cgen's by-value parameter emission (the D-05-02 blocking precondition)

```go
// Source: internal/compiler/cgen/cgen.go:157, read directly this session
fmt.Fprintf(&out, "static %s %s(%s %s) {\n", typeName, functionName, typeName, parameterName)
```
This confirms D-05-02's claim verbatim: the emitted C function signature takes its parameter by value (`typeName parameterName`, not `typeName *parameterName`). `restrict` cannot be meaningfully applied here.

### Verified: the `lang.foreign/0` sidecar's `emitted_attributes` field (correcting D-05-04's schema-location wording)

```go
// Source: internal/compiler/cgen/cgen.go:1319-1333, read directly this session
// EmittedAttributes has no omitempty tag (D-04-13): it is deliberately a
// present, empty JSON array this phase, not an omission -- the distinction
// Phase 5 needs to assert on.
type foreignManifestDocument struct {
	Schema               string             `json:"schema"`
	Symbol               string             `json:"symbol"`
	Allocator            string             `json:"allocator"`
	Unwind               string             `json:"unwind"`
	NonlocalExit         string             `json:"nonlocal_exit"`
	Fails                string             `json:"fails"`
	InitializedState     string             `json:"initialized_state"`
	Capture              string             `json:"capture"`
	Retention            string             `json:"retention"`
	Aliasing             string             `json:"aliasing"`
	Layout               *core.RecordLayout `json:"layout"`
	EmittedAttributes    []string           `json:"emitted_attributes"`
	UncheckedObligations []string           `json:"unchecked_obligations"`
}
```
This is a field of the `lang.foreign/0` sidecar schema (`ForeignManifestSchema` at cgen.go:1312), digest-bound into `evidence.Manifest.ForeignDigest` — not a field on `evidence.Manifest` (`lang.evidence/1`) itself, whose full field list is:

```go
// Source: internal/compiler/evidence/evidence.go:49-74, read directly this session
type Manifest struct {
	Schema           string   `json:"schema"`
	ID               string   `json:"id"`
	IDAlgorithm      string   `json:"id_algorithm"`
	SourceSchema     string   `json:"source_schema"`
	CoreSchema       string   `json:"core_schema"`
	ExecutionSchema  string   `json:"execution_schema"`
	DiagnosticSchema string   `json:"diagnostic_schema,omitempty"`
	CompilerIdentity string   `json:"compiler_identity"`
	ClangIdentity    string   `json:"clang_identity"`
	Target           string   `json:"target"`
	Flags            []string `json:"flags"`
	Policy           string   `json:"policy"`
	SourceDigest     string   `json:"source_digest"`
	CoreDigest       string   `json:"core_digest"`
	CDigest          string   `json:"c_digest"`
	ExecutionDigests []string `json:"execution_digests,omitempty"`
	DigestClaim      string   `json:"digest_claim,omitempty"`
	KnownEscape      string   `json:"known_escape,omitempty"`
	ForeignDigest    string   `json:"foreign_digest,omitempty"`
}
```
No `EmittedAttributes` field exists here. Phase 5's attribute-justification population target is `cgen.go`'s `EmitForeignManifest`/`foreignManifestDocument`, cross-checked independently by `corevalidate` per D-05-04's re-derivation requirement — the planner should write tasks against these two files, not against `evidence.Manifest`.

### Verified: an existing mutation-runner's marker-count guard shape (pattern to copy for new NAT-03 mutations)

```go
// Source: internal/compiler/session/session.go — grep-located this session at
// lines 67, 130, 193, 244, 544 (OwnedBackendMutationRunner,
// ReleaseOmissionMutationRunner, NonlocalPadOmissionMutationRunner,
// NonlocalLedgerOmissionMutationRunner, LayoutMutationRunner respectively).
// Each wraps native.Runner, requires an exact marker count in the generated
// C (refusing to run on >1 or 0 matches to avoid mutating an ambiguous
// target), performs its specific hostile edit, and reports nonzero
// RecomputedWork. New alias-fact / allocator-mismatch / UAF mutation
// runners should be new siblings on this exact shape.
```

### Verified: sanitizer smoke fixture and observed diagnostic (ran this session)

```c
/* /tmp/asan_smoke.c */
#include <stdlib.h>
int main(void){ char *p = malloc(1); free(p); return p[0]; }
```
```bash
clang -std=c17 -O1 -g -fno-omit-frame-pointer -fsanitize=address,undefined -fno-sanitize-recover=all /tmp/asan_smoke.c -o /tmp/asan_smoke
ASAN_OPTIONS=halt_on_error=1:abort_on_error=1:symbolize=0:detect_leaks=0:detect_odr_violation=0:alloc_dealloc_mismatch=1 /tmp/asan_smoke
```
Output (first line, matches D-05-14's required substring exactly): `==NNNNN==ERROR: AddressSanitizer: heap-use-after-free on address ...`; process terminates via `ABORTING`, exit code `134`.

## Runtime State Inventory

Not applicable — this is not a rename/refactor/migration phase. Phase 5 adds new subsystems (sanitizer lanes, reducer, LTO tier, control registry) and retires one internal helper (`discoverLoanLastUses`) via the shadow-mode migration procedure (D-05-35); it does not rename or relocate any existing stored data, service config, OS-registered state, secrets, or build artifacts. Skipped per the trigger condition in the verification protocol.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `alloc_dealloc_mismatch` defaults OFF on macOS's ASan runtime | Common Pitfalls #1 | If actually on by default on this exact Clang/compiler-rt build, D-05-13's explicit pin is still harmless (idempotent), so risk is low — but the *reasoning* for why the pin is necessary would be wrong, and the planner should not claim to have reproduced the default-off behavior (this session's synthetic probe was inconclusive, not confirmatory) |
| A2 | No `-flto`-specific linker flags (no `lld`, no plugin, no `llvm-ar`) are needed beyond plain `clang -flto` on this host for the two-TU case tested | Code Examples / Pattern 4 | If Phase 5's real generated-C link step needs additional flags (e.g., more TUs, static libraries, foreign object files), the native.Runner LTO wiring may need iteration beyond what this two-file smoke test confirms |
| A3 | The concrete second-allocator symbol names, the `lang.mismatch/0` field encoding, and the registry file format remain open per 05-CONTEXT.md's "Claude's Discretion" — this research does not narrow those choices further | Standard Stack / Architecture Patterns | Low — these are explicitly left to the planner/executor by the locked context, not gaps in research |

**If this table were empty, no confirmation would be needed; A1-A3 are all low-risk or explicitly deferred to Claude's discretion by 05-CONTEXT.md, so none blocks planning.**

## Open Questions

1. **Does the eventual `lang.mismatch/0` schema need a `ForeignDigest`-style binding to the sanitizer/LTO evidence it may reference, or does it stand fully independent per D-05-26's "new top-level schema, no other identity moves"?**
   - What we know: D-05-26 lists the exact field set (`schema`, `diverging_axis`, `engine_pair`, `diverging_operation_id`, `reduced_core`, `reduced_source`, `minimality`, `total_recomputed_work`, `reduction_attempts`, `event_window`, `causal_chain`, `causes`) with no digest-binding field mentioned.
   - What's unclear: Whether an evidence-manifest cross-reference is expected implicitly (so a `lang.mismatch/0` document can be traced back to the exact evidence manifest it was produced from) or whether that's out of scope for this phase.
   - Recommendation: Follow D-05-26's field list exactly as written; if traceability back to a manifest is needed, treat it as an additive field decision at Claude's discretion (already licensed by 05-CONTEXT.md), not a schema question requiring further research.

2. **Where exactly does `Phase5RequiredControls()` belong relative to the existing `Phase4RequiredControls()` — same file, new function — and does the D-05-18 enumeration-bound constant live in the same file or a new one?**
   - What we know: The Phase 4 pattern places `Phase4RequiredControls()` in `session.go` alongside `verifyForeignCorpus` and its siblings; the enumeration-bound constant is new to Phase 5.
   - What's unclear: Whether `session.go` is getting large enough (2500+ lines observed) that the planner should split new Phase 5 additions into a new file within the same package.
   - Recommendation: Left to Claude's discretion per 05-CONTEXT.md; a new file in the same `session` package (e.g., `session_phase5.go`) would keep the existing file's size in check without violating the "peer, not fork" pattern, but this is a style choice, not a correctness one.

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go's built-in `testing` package (`go test`), no external test framework |
| Config file | none — `go.mod`/`go.sum` govern build; `scripts/verify-phase{N}.sh` is the bounded gate script, not a test-framework config |
| Quick run command | `go test ./internal/compiler/... -run <TestName>` for a targeted test |
| Full suite command | `go test ./... && go test -race ./... && go vet ./...` (verified as the exact sequence `scripts/verify-phase4.sh` runs; `verify-phase5.sh` should be a peer) |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| NAT-02 | `-O0`/`-O3`/`-O3+LTO` differential agrees on terminal outcome, event order, live-resource state | differential (execution) | `lang --json verify testdata/phase5` (new corpus) | ❌ Wave 0 — testdata/phase5 does not exist yet |
| NAT-03 | Seven hostile mutations each detected by an independently meaningful lane | mutation-kill (unit + differential) | New `Test*MutationRunner` tests per mutation, mirroring existing `TestOwnedBackendMutationIsMismatch`-style tests | ❌ Wave 0 — new mutation runners not yet written |
| NAT-03 (sanitizer half) | ASan/UBSan lanes catch retained-pointer, allocator-mismatch, UAF | sanitizer (process exit + diagnostic substring) | New sanitizer-lane test invoking the separate sanitizer binary, asserting exit code + `ERROR: AddressSanitizer: ...` substring | ❌ Wave 0 — sanitizer lane build config does not exist in `native.go` yet |
| INT-02 | Reducer produces minimized case + causal trace on injected mismatch | reducer (unit, 3 required mutation-kills) | New `Test*Reducer*` tests: no-op-reducer-goes-red, predicate-too-loose, non-determinism | ❌ Wave 0 — reducer package does not exist yet |
| QLT-01 | Every spike control has a live descendant or waiver; no stale `control:` reference | registry audit (unit) | `TestQLT01RegistryComplete` | ❌ Wave 0 — registry does not exist yet |
| Phase-5-wide | `verify-phase5.sh` control set matches `Phase5RequiredControls()` | equality test (unit) | `TestPhase5RequiredControlsMatchScript`, mirroring `TestPhase4RequiredControlsMatchScript` (verified pattern exists at session.go — exact test name not grepped this session but the pattern is confirmed live via `Phase4RequiredControls()`'s doc comment) | ❌ Wave 0 — new for Phase 5 |
| `discoverLoanLastUses` retirement | Zero divergence between old scanner and `loanLivenessFixpoint` over full enumeration | shadow-mode logging test | Extend `TestOwnershipSequenceExhaustive`/`TestBranchSequenceExhaustive` (both verified to exist in the codebase this session, referenced from check.go's own comments) to log divergences | ❌ Wave 0 — shadow-mode logging harness not yet added |

### Sampling Rate

- **Per task commit:** targeted `go test ./internal/compiler/<changed-package>/... -run <relevant test>`
- **Per wave merge:** `go test ./... && go test -race ./... && go vet ./...` plus `lang --json verify testdata/phase{1..5}` for all five corpora (non-regression proven by re-running prior corpora with the freshly built binary, per the Phase 4 precedent at `scripts/verify-phase4.sh`)
- **Phase gate:** Full `scripts/verify-phase5.sh` green, including the mandatory mid-phase gate D-05-40 requires after the alias-fact/liveness work and before the reducer/QLT-01 work

### Wave 0 Gaps

- [ ] `testdata/phase5/` — hand-written adversarial subset (6-10 programs) plus bounded enumerated closure generator — covers NAT-02
- [ ] New mutation runner tests for alias-fact, second-allocator, and UAF hostile fixtures — covers NAT-03
- [ ] Sanitizer-lane build configuration in `native.go` plus its own test harness — covers NAT-03
- [ ] New reducer package plus its three required mutation-kill tests — covers INT-02
- [ ] `lang.mismatch/0` schema type plus round-trip test — covers INT-02
- [ ] QLT-01 registry file/table plus `TestQLT01RegistryComplete` — covers QLT-01
- [ ] `scripts/verify-phase5.sh` plus `TestPhase5RequiredControlsMatchScript` — covers phase-wide gate
- [ ] Shadow-mode divergence-logging harness over `TestOwnershipSequenceExhaustive`/`TestBranchSequenceExhaustive` — covers `discoverLoanLastUses` retirement precondition

## Security Domain

`security_enforcement: true`, `security_asvs_level: 1`, `security_block_on: "high"` confirmed in `.planning/config.json` this session.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | No | This phase has no user-facing auth surface — it is a compiler backend/evidence phase |
| V3 Session Management | No | Same as above |
| V4 Access Control | No | Same as above |
| V5 Input Validation | Yes | `corevalidate.foreignContractFieldsCSafe`/`commentSafeForeignField` and `cgen.unsafeForeignContractField` (both verified this session, `corevalidate.go:1765`/`cgen.go:873`) already gate every spliced `ForeignContract` field against C-injection; D-05-36 extends this to the `Alias` field. This is the phase's actual "input validation" surface — untrusted foreign-contract strings spliced into generated C source |
| V6 Cryptography | No | SHA-256 content digests (already shipped) are used for identity binding, not confidentiality or authentication — no new cryptographic surface this phase |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| C-source injection via an untrusted `ForeignContract` field spliced into generated C (e.g., a hostile `Alias` value containing `*/`, comment terminators, or arbitrary C tokens) | Tampering | Existing `commentSafeForeignField`/`validCIdentifier`-style predicates in both `corevalidate` and `cgen`, independently implemented (D-12) — D-05-36 requires extending this coverage to `Alias`, which today is verified to be excluded from all three audit layers |
| A false optimizer attribute (`restrict`) claimed without a proven backing fact, silently miscompiling downstream code | Tampering / Elevation of privilege (of the optimizer's trust) | `control:foreign.no_unproven_attributes` narrowed per D-05-03: every `emitted_attributes` entry must have a `corevalidate`-independently-re-derived justification; a hard build failure on any unjustified attribute |
| A coordinated multi-artifact lie (source, core, digest, and manifest all agreeing on a false claim) | Repudiation / Tampering | Not closable by this project's architecture (documented residual, D-04-31/D-05-30) — the mitigation is honesty: a demonstrable named escape (`escape:coordinated-source-to-core-false-claim`) rather than a false claim of closure |
| Sanitizer evidence laundering (a shared boolean flag letting sanitizer noise pass as semantic-equivalence evidence) | Tampering | Structural, type-level isolation (D-05-11) — the comparator's function signature does not accept sanitizer-lane evidence at all, so no refactor can silently launder it |

## Sources

### Primary (HIGH confidence — read/executed directly this session)
- `internal/compiler/cgen/cgen.go` (lines 140-170, 856-900, 1300-1340, 1527-1560) — read directly, confirms by-value parameter emission, `unsafeForeignContractField`, `foreignManifestDocument`, `BannedOptimizerAttributes`
- `internal/compiler/evidence/evidence.go` (lines 40-90) — read directly, confirms `Manifest` struct field list, no `EmittedAttributes` present
- `internal/compiler/corevalidate/corevalidate.go` (lines 369-390, 1756-1770) — read directly, confirms `foreignContractFieldsCSafe`
- `internal/compiler/check/check.go` (grep-located key symbols: `discoverLoanLastUses`, `loanLivenessFixpoint`, `checkBranch`, `checkLinear`, `materializeLoanEndpoints`) — confirms both liveness laws coexist today, matching D-05-34/D-05-35's premise
- `internal/compiler/session/session.go` (grep-located mutation runners and `Phase4RequiredControls`, lines 67-2020 range) — confirms the five worked mutation-runner examples and the required-control-set pattern
- `internal/compiler/native/native.go` (lines 43-460) — read directly, confirms Clang invocation flags and `validateExecution`
- `.planning/spikes/004-independent-certificate-checker/README.md` and `.planning/spikes/005-native-ffi-provenance-cleanup/README.md` — read in full, confirms Apple Clang 21.0.0 arm64 host identity and the exact `2`/`1` `-O0`/`-O3` divergence D-05-05 cites
- `.planning/spikes/MANIFEST.md` — read in full, confirms all five spike verdicts/tags for QLT-01's registry extraction
- `.planning/phases/04-fallible-resources-and-c-boundary/04-CONTEXT.md` and `04-DEBT.md` (D-04-13, D-04-14, D-04-26, D-04-32, D-04-33 sections) — grep-confirmed, matches 05-CONTEXT.md's carry-forward citations
- `scripts/verify-phase4.sh` — read directly, confirms the exact test/build/verify sequence and control-list duplication pattern
- Direct toolchain execution this session: `clang --version` (Apple clang 21.0.0 arm64), `go version` (go1.24.0 darwin/arm64), a synthetic ASan heap-use-after-free build/run (reproduced the exact `ERROR: AddressSanitizer: heap-use-after-free` / exit 134 diagnostic shape), a two-TU `-O3 -flto` build/link/run (exit 0, correct result)

### Secondary (MEDIUM confidence)
- Clang AddressSanitizer official docs (`clang.llvm.org/docs/AddressSanitizer.html`) — cited in spike 005's README for the `alloc_dealloc_mismatch`-defaults-off-on-macOS claim; not independently re-derived this session (synthetic probe was inconclusive, see Pitfall 1)
- LLVM LangRef (`llvm.org/docs/LangRef.html`) on `noalias`/capture/provenance semantics — cited in spike 005's README, not re-verified against LLVM source this session

### Tertiary (LOW confidence)
- None — every claim in this document is either read/executed directly this session or explicitly tagged `[CITED]`/carried from the already-vetted 05-CONTEXT.md discussion with its own provenance trail

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — toolchain versions and behaviors directly confirmed by running commands on the actual host this session
- Architecture: HIGH — every cited file/line was read directly, not assumed from training data
- Pitfalls: HIGH for LTO/restrict/mutation-runner-pattern pitfalls (directly verified); MEDIUM for the `alloc_dealloc_mismatch` macOS-default claim (cited from spike 005/Clang docs, not independently re-derived by a successful falsification this session)

**Research date:** 2026-09-05
**Valid until:** Toolchain-version-bound — re-verify if the host's Clang version changes (D-05-05's `-O0`/`-O3` divergence and this session's ASan/LTO confirmations are all Apple Clang 21.0.0/arm64-specific observations, not portable guarantees, per 05-CONTEXT.md's own "Accepted residual limitations" note)
