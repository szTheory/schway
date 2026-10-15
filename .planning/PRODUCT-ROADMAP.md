# Schway — Living Product Roadmap

Updated 2026-10-03. This document carries current direction; ROADMAP.md owns
committed milestone phases, REQUIREMENTS.md owns acceptance, and
LANGUAGE-MATURITY.md records demonstrated capability. Historical milestone
forecasts and evidence remain in their archives.

## Intent and present position

M006 final hosted receipts: reopened 2026-10-13 (historical codedigest b9253457705a224a0580d74896e8e5bd3821446a1d467626b7d611ca0d1c4b5c)

The hosted receipts cover the M006 programs (checksum, adler_low, the adler32 library demo, the Phase 29 Adler-32 fixture and out-of-bounds control) plus sum_to_n and FizzBuzz (N2 closed). Receipt rows are default-build only; the optimization matrix is bound by the same run's go test lanes. Receipts were re-taken after Phase 32.1 at the 32.1 freeze; run 38315799626 is kept as historical; any later code change reopens them.

Amended 2026-10-13: Hosted run 38335175306 (headSha 9f8f878f) receipts were bound to codedigest b9253457705a224a0580d74896e8e5bd3821446a1d467626b7d611ca0d1c4b5c and are now historical; Phase 33 code changes reopened them. Phase 33 binds no hosted receipt. M006 receipts are historical; current hosted evidence resumes in Phase 38.

Build a usable general-purpose language for AI-authored, human-audited software:
fast structured feedback, readable canonical source, explicit costs, low-level
ownership, deterministic cleanup, and reproducible evidence. A native compiler
already exists (Go 1.24 → readable C17 → Clang), alongside an interpreter. A new
VM or LLVM backend is not a prerequisite for useful programs.

M001 established source-to-native behavior; M002 added executable Schway calls;
M003 added independent returns, computed matches/payloads, and U64 constants.
Phase 21 completed contract/retirement prework. Phase 22 delivered its
application/build/evidence implementation and passed objective README contract
UAT; its refreshed 2026-09-30 verifier report passes 5/5 truths. No subjective
readability claim or human UAT remains. The separate `schway app build` /
`schway app run` route retains a native app, accepts a bounded U64 input, starts
it once, and preserves ordinary streams and process outcomes. Evidence capture
and explicit differential replay remain separate. The older `schway run` path
is still the synthetic-input O0/O3 conformance harness.

Phase 23 admits a bounded file-byte allocation that remains live through
Schway-directed use and generated local cleanup. Its independent physical
observer and close-failure controls have historical focused receipts on macOS
and a local Linux ARM64 container. Hosted run
[36707529870](https://github.com/szTheory/schway/actions/runs/36707529870)
passed the Ubuntu/macOS check suites and both evidence aggregates, including
`scripts/verify-phase23.sh`, at source SHA
`f991298b29779838a2b1a5c3cd5ac90aafcb84fc`. The refreshed 2026-09-30 verifier
passes 5/5 truths and Phase 23 is complete. No local project suites or UAT were
rerun for that refresh; the hosted receipt supplies current behavioral
evidence, and the phase context requires no human UAT. Phase 24 now admits the
bounded file-byte owner transfer through Schway calls and returns, repeated
helper activations, and exactly-once cleanup on normal and typed-error paths.
Its independent physical observer proves use and destruction, while model
replay remains explicit about not proving host IO or physical cleanup. The
2026-10-01 verifier passes 5/5 truths; hosted run
[36856048690](https://github.com/szTheory/schway/actions/runs/36856048690)
passed the focused aggregate and full/race suites on Ubuntu Linux/x86_64 and
macOS Darwin/arm64 at source SHA
`ed94ef7972b25deb85ab90fafbf3403dc31629f5`. No local project tests or scripts
were run for closeout. Phase 25 completed the bounded shared/exclusive
read-copy pointer families and integrated utility, with focused local checks
and hosted dual-host evidence.
Hosted run 36971855722 passes all 18 family/host/lane rows and closes EVD-10;
see the dated provenance amendment below. App build receipts remain
incomplete/non-cacheable when runtime closure is unknown. M004 is complete;
M005 has delivered checked scalar computation in Phase 26. Phase 27 is in
progress: exact FizzBuzz, checked U64 equality/remainder, fixed literal writes,
and the shared application-output ceiling are implemented. Focused local
semantic, output-boundary, evidence, and public-CLI checks pass. The remaining
Phase 27 work is the fail-closed macOS/Linux receipt route, actual same-revision
hosted receipts, and the clean-checkout guide.

### Phase 26 completion amendment — 2026-10-03

**Source inspection:** `examples/sum_to_n.schway` now uses local mutable U64
state, `<`, checked `+`, `if/else`, and pre-tested `while`; its immutable scalar
copy emits distinct dynamic occurrences. Checker, core validator, and origin
validator derive scalar admission independently with deterministic 65,536-work
budgets. Ownership, resources, loans, loan-derived provenance, and unsupported
cyclic events remain refused. The path oracle remains acyclic; it is not used
as a scalar-loop proof. No new backend or dependency was added.

**Newly executed hosted evidence:** focused run
[37118144404](https://github.com/szTheory/schway/actions/runs/37118144404) and full
run [37118315516](https://github.com/szTheory/schway/actions/runs/37118315516)
passed on Ubuntu and macOS at source revision
`46bee44ad87891d8a8547b2489f029d0fe237899`. Public inputs 0/10/1,000 produce
exactly `0\n`/`55\n`/`500500\n`; 1,001 fails without stdout. Direct overflow
fails with exit 65 and exact `schway: U64 addition overflow\n` stderr. Each
engine is checked against literal expected answers, including reached
wrong-result and skipped-iteration controls. Repeated-copy ordinals, forged
sequences, independent analysis exhaustion/refusals, and evidence capacity
controls pass. Full/race suites and current evidence aggregates pass too.
Independent verification passes 5/5 roadmap truths; review is clean and all
12 declared security mitigations are closed at ASVS L1.

**Limits and provenance:** no local tests or native runs were performed for
Phase 26. Prior M004 receipts remain historical. The 35-pair validation corpus
receipt is from hosted run 37098702559; current semantic proof comes from the
later implementation-bound runs above. Cold/warm latency distributions were
not measured. Capture remains separate from independent verification, and the
event peer checks structural attribution rather than proving overflow operands.
Equality, remainder, fixed text, shared output limits, and the complete M005
cross-program evidence/documentation contract remain Phase 27-owned.

### Phase 27 progress amendment — 2026-10-03

**Source inspection:** `examples/fizzbuzz.schway` and
`examples/sum_to_n.schway` are ordinary-source witnesses. The checker and both
independent validators admit U64 equality/remainder, reject a constant zero
divisor, and retain the dynamic zero failure path. Fixed literal writes are
bounded by the shared 65,536-byte application-output limit; application output
and evidence capacity remain separate. Ownership, resource, loan, and
loan-derived loop-carried state remain refused. The native host receipt driver,
two-host gate, CI wiring, and clean-checkout guide are still the next plans.

**Newly executed local evidence:** focused Go selectors for exact answers,
reached wrong-remainder/wrong-arithmetic/skipped-iteration controls, overflow
and zero cases, repeated event identity/order, evidence capacity, output-limit
separation, legacy document compatibility, and the public app CLI passed on
macOS/arm64 in Phase 27 Plans 27-07/27-08 at implementation commit `21fb123`.
The checks exercised interpreter and native builds. They were focused checks;
no full project suite, hosted Phase 27 run, or Linux native execution is claimed.

**Remaining boundary:** the local macOS results do not close EVD-12. Both
actual hosted hosts must pass the new Phase 27 driver at one source/fixture
revision, and DX-16 still needs its linked clean-checkout guide. Phase 25/26
receipts remain historical and cannot substitute for those runs.

### Phase 27.1 transition amendment — 2026-10-10

**What 27.1 changed:** application admission routes (`schway build`, `schway app verify`, and the API-only interpreter application route now apply checker, corevalidate, then the origin peer through one shared admission helper) and the receipt-gate guide (pinned re-check command and revision hint). No new language capability was added.

**Seed transition check:**

| Seed | Disposition | Reason |
|------|-------------|--------|
| SEED-261008-trz | retain dormant | 27.1 produced no named failure that meets its trigger_when for consumer-triggered failure observability. |
| SEED-261008-z1g | retain dormant | No reviewer query gap for a bounded symbol/caller query surfaced during 27.1. |
| SEED-261008-z1s | retain dormant | No contested contract/source form arose in 27.1. |

**Current three recommendations re-checked:** recommendation 1 (finish Phase 27) is complete once 27.1's hosted run lands; recommendations 2 and 3 are unchanged. Next action: the D-27.1-11 sequence (`/gsd-secure-phase 27`, `/gsd-validate-phase 27`, milestone audit re-run), then milestone completion. Nine audit/review items are recorded as deferred in STATE.md (D-27.1-10); none is implemented.

### Phase 27.2 transition amendment — 2026-10-10

**What 27.2 changed:** audit finding R1 is restored. Branch-free scalar bodies (let, var, mixed, pure copy, pure constant, Unit write-only) are admitted end to end again through a checker return-block split and a structural cgen classifier clause; validators are unchanged. No new capability was added. R1b (a scalar-first `let` followed by a call `let`, refused with syntax.expected_linear_result by a671308 parser routing) is deferred and pinned by a characterization test.

**Seed transition check:**

| Seed | Disposition | Reason |
|------|-------------|--------|
| SEED-261008-trz | retain dormant | 27.2 produced no trigger for consumer-triggered failure observability. |
| SEED-261008-z1g | retain dormant | 27.2 produced no bounded symbol or caller query gap. |
| SEED-261008-z1s | retain dormant | 27.2 produced no contested contract or source form. |

**Current three recommendations re-checked:** unchanged. Next action after the hosted run is the milestone audit re-run.

### Phase 27.3 transition amendment — 2026-10-10

**What 27.3 changed:** milestone-audit warnings W1 and W4 are closed without new capability. `schway app verify` on a loop program now names event-capacity exhaustion per case (native.event_capacity_exhausted, reported unsupported) instead of aborting the whole set, through a producer-side named exit in the loop-bearing conformance C and a runner mapping; loop-free conformance C and all application C stay byte-identical. The Phase 27 guide writes its outputs under ignored `.schway/phase27`, so the clean-checkout block no longer dirties the tree. Validators are unchanged. Loop-aware event capacity (full-domain app verify of `sum_to_n`) remains deferred backlog per D-27.3-03.

**Seed transition check:**

| Seed | Disposition | Reason |
|------|-------------|--------|
| SEED-261008-trz | retain dormant | W1 did not meet its trigger: the failure was diagnosed from source and existing capture during discuss-phase research, so current capture, observers, debug map and source replay evidence were sufficient. |
| SEED-261008-z1g | retain dormant | No reviewer query gap for a bounded symbol or caller query surfaced during 27.3. |
| SEED-261008-z1s | retain dormant | No contested contract or source form arose in 27.3. |

**Runnable gain check (AGENTS.md: a second consecutive enabling phase without one triggers scope review):** 27.1, 27.2 and 27.3 are consecutive tech-debt closures, so the check is applied explicitly. Result: 27.3 has a runnable gain. `schway app verify` on `sum_to_n` now completes with per-case results (n <= 14 pass, larger n reported unsupported with native.event_capacity_exhausted) where it previously aborted the whole set, and the Phase 27 guide's clean-checkout block now runs without dirtying the tree. No scope review is triggered.

**Plan count justification (four plans):** the safety obligations are milestone-audit warnings W1 (fail-closed naming of capacity exhaustion) and W4 (clean-tree receipts). The W4 guide work stays separate in plan 03, and plan 04 holds only evidence and the gated hosted run.

**Current three recommendations re-checked:** unchanged. Next action after the hosted run is the milestone audit re-run.

### M006 kickoff amendment — 2026-10-10

**What changed:** M005 shipped (tag `vM005`), so recommendation 1 below
(Phase 27 FizzBuzz and cross-host evidence) is complete. Recommendation 2 is
now milestone M006 "Bounded Byte Utility and Local Modules", Phases 28–32. See
[ROADMAP.md](ROADMAP.md), [REQUIREMENTS.md](REQUIREMENTS.md) and
[research/SUMMARY.md](research/SUMMARY.md).

The consumer is a `checksum FILE` program that prints RFC 1950 Adler-32 for
files up to 65,536 bytes. It is built from an `adler32` library in a local
module, which a second application imports. Adler-32 was chosen because it
needs only the admitted checked addition, comparison and remainder. CRC-32
waits for a bitwise-operator milestone.

The loop reads a read-only byte view, and no loan crosses a back edge, so the
FLOW-02 refusal boundary is unchanged. Modules are whole-program composition.
They do not establish B1/DX-06 or separate compilation.

M005 carry-over: W2, W3, N3 and N1 go to Phase 28, and N2 is the last plan of
Phase 32. N4 and R1b stay tracked but unscheduled.

**Current three recommendations at kickoff:**
1. M006 Phases 28–32, as above.
2. A bounded JSON configuration consumer (unchanged from item 3 below). It
   becomes reachable once byte views and modules exist, but still waits for a
   named tool.
3. Bitwise U64 operators (xor, shift, and) with independent evidence. The
   user-visible program is CRC-32/`cksum`-compatible output. It is promoted
   only if a named consumer needs interop that Adler-32 cannot provide.

**Seed transition check:**

| Seed | Disposition | Reason |
|------|-------------|--------|
| SEED-261008-trz | retain dormant | No named program failure that current capture cannot diagnose. |
| SEED-261008-z1g | retain dormant | Module imports may later create a caller-query need; no reviewer gap observed yet. Re-check at Phase 30. |
| SEED-261008-z1s | retain dormant | Import and byte-read surface syntax is a Phase 29/30 discussion item; promote only if a reviewer contests a concrete form. |

### Phase 28 transition amendment — 2026-10-10

**What 28 changed:** `core.entry_ambiguous` is now a source refusal that exits 2 on check, build, both run engines and app verify. A program with functions `main` never calls is pruned after validation and runs: `examples/phase28/unused_function.schway` builds, and `app run` with input 7 prints 7. In app verify, source refusals (checker, core, origin, path oracle, entry) now exit 2 with the refusing code and still write the fail report, while capacity-exhausted cases stay `unsupported` with exit 3 and a digest (schema `lang.app-verification/2`). Core and origin refusals exit 2 on every command (W2 and W3 closed). One exit-code table lives in `internal/compiler/exitcode` and is published in the README "Exit codes" section, pinned by TestExitCodeDocsTable. Witnesses: TestExitClassesUniformAcrossSurfaces, TestUnreachableFunctionProgramRuns and TestAppVerifySourceRefusalExitsInvalid (newly executed locally, focused). Hosted macOS and Linux CI has not run for Phase 28 yet.

**Seed transition check:**

| Seed | Disposition | Reason |
|------|-------------|--------|
| SEED-261008-trz | retain dormant | Every Phase 28 failure was diagnosed from source, existing diagnostics and the exit-class tests; no failure was left unexplained by current capture and replay evidence. |
| SEED-261008-z1g | retain dormant | Finding the callers of an entry did not need a symbol or caller query beyond ordinary search and the callgraph tests; no measured reviewer gap. Re-check at Phase 30. |
| SEED-261008-z1s | retain dormant | The entry-selection rule (`main` first, then a unique root) was settled in phase discussion, not contested by a consumer or reviewer. |

**Runnable gain check (AGENTS.md):** Phase 28 has a runnable gain. A library-style file with an unused function now builds and runs, and `app verify` of `sum_to_n` at n=1000 reports a named `native.event_capacity_exhausted` unsupported case with a digest instead of an unnamed failure. No scope review is triggered.

**Current three recommendations re-checked:** unchanged. Recommendation 2 (bounded byte utility and local modules) continues with Phase 29. Reprioritize if a named consumer needs bitwise operators or a smaller byte API before byte views.

### Phase 29 transition amendment — 2026-10-11

**What 29 changed:** ordinary Schway source reads a read-only `ByteView` in a scalar loop (`byte_len`, `byte_at`, parameter named `bytes`) with a never-elided bounds check (exit 65, stderr `schway: byte index out of bounds`), and `adler32` is pinned on both engines. Release before the last use of a view is refused by `check.release_while_view_live`, `core.release_while_view_live` and `pathoracle.release_while_view_live` (D-22). The native runnable gain is `schway build examples/phase29/adler32_fixture.schway --manifest examples/phase29/fixture_bytes.bindings.json --output ART` plus `schway app run ART -- N`; `schway run --engine=*` is unchanged. Witnesses (newly executed locally, focused): TestPhase29ReleaseBeforeViewUseRefused, TestPhase29OwnerLoopBlockRefused, TestPhase29OOBApplicationLanes, TestPhase29FixtureAppCLI. Hosted CI run 38186068052 at 906d648 passes on macOS and Linux, and the strict spike 008 step there keeps the 65,536-byte cap (plan 29-12, spike 008 VALIDATED; worst hosted maxima 8.0 ms native and 376.5 ms interpreter against 3 s and 12 s budgets). Full check: [29-TRANSITION-CHECK.md](phases/29-byte-views-and-the-adler-32-library/29-TRANSITION-CHECK.md).

**Seed transition check:**

| Seed | Disposition | Reason |
|------|-------------|--------|
| SEED-261008-trz | retain dormant | Phase 29 failures were diagnosed from source, refusal codes and focused tests. |
| SEED-261008-z1g | retain dormant | No measured reviewer gap; re-check at Phase 30 when imports create cross-file callers. |
| SEED-261008-z1s | retain dormant | Byte-read surface syntax is now settled by D-01 and D-18 and no reviewer contests a concrete form; this supersedes the earlier "Phase 29/30 discussion item" wording. |

**Current three recommendations re-checked:** recommendation 2 (bounded byte utility and local modules) is partly delivered: byte views and the Adler-32 library exist, while local modules (Phase 30) and the `checksum FILE` consumer with the measured cap (Phase 31) remain. Recommendations 1 and 3 are unchanged. Reprioritize if a consumer needs bitwise operators first or needs files above the 65,536-byte cap.

### Phase 30 transition amendment — 2026-10-12

**What 30 changed:** local modules exist as whole-program composition. An entry imports named functions from files under its own directory; the loader merges them into one program that is checked, interpreted and emitted together, with the module carried in identity, `unit=` diagnostics and native C names. The new refusal frontier is the `module.*` code family. The runnable gain is the Phase 29 Adler-32 program with the library imported from `examples/phase30/adler32.schway`: `schway build examples/phase30/adler32_cli.schway --manifest examples/phase29/fixture_bytes.bindings.json --output ART` plus `schway app run ART -- N` prints the pinned checksums, and the interpreter gives the same ten answers through the session API. Witnesses (newly executed locally, focused): TestPhase30AdlerModuleAppCLI, TestPhase30AdlerModuleInterpreter, TestPhase30LibraryIsSingleCopy. *Amended 2026-10-12 (Phase 31 D-01): the library and CLI moved to examples/checksum/; commands above are historical.*

**Non-claims:** this is not separate compilation, it does not establish B1/DX-06, and modules are not hashed in build identity or evidence until Phase 31 (MOD-06). Hosted CI for Phase 30 has not run; it runs at the phase gate and its run IDs go in 30-VERIFICATION.md. M006 receipts belong to Phase 32 (EVD-15).

**Seed transition check:**

| Seed | Disposition | Reason |
|------|-------------|--------|
| SEED-261008-trz | retain dormant | Every Phase 30 failure was diagnosed from refusal codes, `unit=` diagnostics and focused tests. |
| SEED-261008-z1g | retain dormant | Cross-file callers were found by ordinary search plus TestPhase30SameBareNameTwoModules and the debug-map unit; no measured reviewer gap. Re-check when a second consumer exists. |
| SEED-261008-z1s | retain dormant | No reviewer contested the import surface syntax. |

**Current three recommendations re-checked:** recommendation 2 (bounded byte utility and local modules) now has byte views, the Adler-32 library and local modules; the `checksum FILE` consumer with the measured cap and the second application importing the library (Phase 31, including MOD-06 module hashing) remain. Recommendations 1 and 3 are unchanged. Full check: [30-TRANSITION-CHECK.md](phases/30-local-modules/30-TRANSITION-CHECK.md).

### Phase 26 discussion — 2026-10-02 (historical planning snapshot)

The user adopted the specialist fan-out recommendations for Phase 26: use
function-local mutable `U64`/`Bool` scalars for loop state, a pre-tested
`while` form and ordinary `if/else` blocks, and only the U64 `<` comparison
needed by `sum_to_n`. U64 addition uses ordinary infix source spelling with
checked semantics. Overflow is a program-level non-success at the application
boundary, with bounded stderr and no stdout; it is not a catchable source-level
typed error in this phase. Equality and remainder remain Phase 27 work.
Resource, loan, and loan-derived values remain refused across loop back edges.
These are accepted planning decisions, not implemented behavior. Current
source still refuses cyclic CFGs, and no Phase 26 tests or native checks were
run during this discussion. The existing Go standard library, interpreter,
and C17/Clang route remain sufficient; no dependency or backend was identified.

The user explicitly requested another deep fan-out and automatic adoption of
recommendations. The 2026-09-27 decision integrates product/DX, compiler and
resource architecture, FFI/security, verification, portability, and delivery
reviews. Research and dissent: [M004 research](research/M004/SUMMARY.md).


### Phase 31 transition amendment — 2026-10-12

**What 31 changed:** a file can now be checksummed. The runnable gain: `schway build examples/checksum/checksum.schway --manifest examples/checksum/file_bytes.bindings.json --output ART` plus `schway app run ART -- PATH` prints the Adler-32 of a file of at most 65,536 bytes (Wikipedia gives 300286872), and the same build and run with `examples/checksum/adler_low.schway` prints the low 16 bits (920) from the one shared library. The new refusal frontier is the `FileError` outcomes (`FileTooLarge`, `FileUnavailable`, `AllocFailed`, all exit 65, path-free stderr) and the file-bytes shape codes such as `check.file_bytes_owner_shape_unsupported`. Build identity, receipt and evidence now carry sorted `modules` rows when a program imports (MOD-06). Witnesses (newly executed locally, focused): TestPhase31ChecksumPublicRoute, TestPhase31AdlerLowPublicRoute, TestPhase31CaseMatrix, TestPhase31ReleaseExactlyOnce, TestPhase31BoundedReadMutants, TestPhase31ModuleDigests, TestPhase31SingleCopyAndDualFlip, TestPhase31Controls.

**What 32.1 changed:** exit classes and the module boundary were tightened with no new language feature. `schway run` on an entry that cannot be bound (parameterised `main`) is `tool.run_input_required`, a usage refusal with exit 64; a `ByteView` `main` is refused at `check`; the module-aware route set is named once (`check.ModuleAwareRoutes()`: check, run, build, app verify, debug-map) and the usage text lists every route. `explain` and `query` surface `module.imports_not_loaded` as a passing lookup (exit 0). Witnesses (focused, local): TestPhase321RunInputRequired, TestPhase321ByteViewEntryRefused, TestPhase321ModuleAwareRoutes, TestPhase32PublicRouteAudit. Hosted CI has not run for 32.1.

**Non-claims:** modules stay whole-program composition and B1/DX-06 stays open. The provisional `examples/checksum.schway` frontier (retired 2026-10-12, Phase 32 amendment), the M006 receipts (EVD-15) and the clean-checkout guide belong to Phase 32. Hosted CI for Phase 31 has not run; it runs at the phase gate and its run IDs go in 31-VERIFICATION.md.

**Three most useful next capabilities** (full entries in [31-TRANSITION-CHECK.md](phases/31-bounded-file-checksum-and-second-consumer/31-TRANSITION-CHECK.md)):

1. Phase 32: retire the provisional checksum frontier and record M006 hosted receipts and the guide. User-visible program: the `checksum FILE` clean-checkout guide. Blocker: hosted macOS and Linux receipts. Smallest slice: the receipt gate plus guide. Priority changes if hosted CI cannot be reached or a Phase 31 hosted run fails.
2. Bounded JSON configuration consumer once a tool is named. User-visible program: validate a size-limited config against a fixed schema. Blocker: no aggregates or parser stack. Smallest slice: one schema with byte and depth limits. Priority changes if a named consumer appears or aggregates land first.
3. Bitwise U64 operators with independent evidence (CRC-32). User-visible program: CRC-32 output over the same file bytes, in its own module. Blocker: no xor, shift or and operator. Smallest slice: one operator set with a pinned CRC-32 vector set. Priority changes if a consumer needs interop Adler-32 cannot give or files above the 65,536 limit are needed first.

### M007 kickoff amendment — 2026-10-13

Recommendations 1 (FizzBuzz/cross-host evidence, M005) and 2 (bounded byte
utility and local modules, M006) below have shipped and are historical. A
five-lens fan-out (product, compiler/PL, security, devops/evidence, DX) chose
M007 "Bitwise Checksum and Composable Scalar Helpers": `crc32 FILE`
(CRC-32/ISO-HDLC, bare decimal) from an ordinary `crc32` module, admitting U64
`^ & |`, literal-amount shifts (< 64, no C17 shift UB), hex literals and
arity-1 scalar→scalar calls, with a debt-first Phase 33, an agent diagnostics
floor plus `schway surface --json`, an interpreter file-bytes CLI route, and
receipts bound once after audit (Phases 33–38; see ROADMAP.md and
research/M007/SUMMARY.md). The JSON path is now staged: M008 arity-N,
scalar-only records and exported types via a `wc` + `longest_line` pair, then a
bounded key=value config parser, then recommendation 3's JSON validator. HTTP
remains a separate, later branch.

## Horizons

| Horizon | User-visible result | Dependencies and scope | Exit observation |
|---|---|---|---|
| **Now: M005 practical computation** | `sum_to_n`, then FizzBuzz from ordinary source | Defined U64 arithmetic/remainder and overflow, comparison/Bool, scalar-loop fixed points, fixed text/byte literals, bounded writes and decimal formatting | Public command produces exact expected output; arithmetic boundaries and changed-assumption checker controls pass |
| **Mid term: reusable libraries** | Small byte/file utilities and checksum; a reusable bounded JSON parser/serializer | Arity-N and small aggregates as consumers require, explicit byte views/lengths/indexing, fallible APIs, local modules, explicit resource transfer | Second consumer imports a library without copying it; malformed/truncated/oversized input has specified behavior |
| **Mid term: network branch** | A bounded HTTP client or server for a selected use case | Explicit sockets/timeouts/body/framing/error ownership; audited C/OS adapter or dependency; cleanup on every admitted outcome | One documented real endpoint flow plus adversarial protocol cases; TLS policy specified when needed |
| **Long term: production ecosystem** | Larger independently built libraries and real services/tools | Separate compilation/contracts, package/version reproducibility, richer types/generics when justified, effects/concurrency only with resource semantics | Independent consumers and measured limitations justify each expansion |

These are dependency horizons, not dated delivery promises. JSON and HTTP are
independent branches: an HTTP client can fetch bytes before JSON, and a JSON
tool can consume stdin before networking. Prefer a bounded explicit parser
stack over making language recursion a prerequisite. A local module system
need not introduce separately verified binary contracts.

The former M003-era forecast (M004 loops, M005 aggregates, M006 modules) is
superseded prospectively. M005 now targets `sum_to_n` and FizzBuzz, with phase
numbering continuing at Phase 26. The next milestone will be scoped from a
named byte-oriented consumer. `examples/checksum.schway` was retired on 2026-10-12
(see the PROJECT.md Phase 32 amendment); no unsupported syntax is treated as settled.

## Decisions and alternatives

| Decision | Recommendation and benefit | Cost / rejected alternative | Reopen when |
|---|---|---|---|
| Native backend | Keep Go/standard library, C17/Clang, interpreter oracle | C ABI and defined-semantics lowering require care; a VM/LLVM rewrite adds another backend before user value | A measured consumer/latency/target limit cannot be resolved within this path |
| M004 versus arithmetic first | Honor resource commitments, establish real application IO and lifetime first; deliver an early scalar application witness | Delays arithmetic by this bounded milestone; computation-first retains synthetic execution and leaves NAT-09 unresolved | Resource scope grows beyond the specified witness; reduce shapes before adding phases |
| First owned resource | Audited C adapter returns a live bounded allocation; Schway uses/transfers/releases it | Adapter still owns file open/close; this does not demonstrate Schway-owned file handles | A consumer needs direct file/socket ownership with specified close-error semantics |
| Pointer families | Separate bounded shared/exclusive read-copy support, no added `restrict`/`noalias` | More work than ordinary foreign alone; no generic mutation/escape/retention | A concrete consumer requires a broader access contract and independent evidence |
| Cleanup | Successful acquisition creates an obligation; release consumes it, transfer moves it, borrow preserves it | Requires real dynamic identities and per-operation contracts; release-op-seeded tracking can miss deleted cleanup | Aggregates, loops, cancellation or fallible destructors change the model |
| Iteration | Start with scalar loop-carried state; retain resource/loan-across-back-edge refusal initially if needed | Needs CFG fixed-point analysis, event occurrence identity and bounded oracle policy | FizzBuzz/sum witness exposes a necessary broader construct |
| Text and libraries | Fixed text + decimal output first; byte-oriented reusable APIs next | Full String/Unicode/collection abstractions remain separate work | A selected consumer needs encoding/ownership behavior beyond byte literals |
| Planning automation | AGENTS procedure + two living documents | Requires agents to perform the review; no scheduled/background service | Repeated missed triggers demonstrate a need for a small executable check |

## Checker and guarantee activation map

The six existing operation consumers are `check`, `corevalidate`,
`originvalidate`, `pathoracle`, `interp`, and `cgen`. Update affected consumers
within the feature slice. Preserve independent derivation at trust crossings;
do not make all peers import the producer's conclusion. A registry entry or
engine agreement alone cannot prove correctness.

| Capability trigger | Analysis/contract that changes | Decisive evidence and refusal boundary | Owner |
|---|---|---|---|
| Real application IO | Entry inputs, ordinary IO/exit versus evidence; side effects | Phase 22's automated tests establish bounded U64 input, one native launch, ordinary streams, separate evidence capture, and explicit independent replay; `TestPhase22READMEContract` pins public forms and limits; general file IO remains unimplemented | Phase 22 closed; extend only for a concrete consumer |
| Live owned allocation | Acquisition-based obligation conservation; per-operation ABI/destructor; noncopyability | Omitted physical destructor with unchanged events must fail; discard must not erase obligation | M004 acquisition/discharge phase |
| Transfer/calls/errors | Owner and resource identity across frames; reverse completion order; failed acquisition | Phase 24's ordinary success and typed-error apps, distinct activations, acquisition-derived peers, and physical observer passed on both hosted hosts; general cleanup shapes remain refused | Phase 24 complete; extend only for a selected broader cleanup consumer |
| Shared/exclusive pointer access | Borrow endpoints, independent straight-line overlap/escape refusal, source-attributed escape diagnostics, actual C ABI and emitted attributes | `TestPhase25PointerPathLiveOverlap`, `TestPhase25PointerPathBorrowedResultEscape`, `TestPhase25IndependentPeerMutations`, and command/diagnostic attribution checks pass locally; hosted run 36971855722 passes 18/18 rows. No unsupported alias/alignment/capture promises | Phase 25 complete; preserve the existing refusal frontier and use hosted evidence aggregate for regression |
| Arithmetic and Bool | Checked U64 addition, `<`, `==`, and `%` run in the scalar path; constant zero remainder is refused and dynamic zero fails consistently | Phase 26 literal answers plus Phase 27 independent peers, zero-divisor cases, and reached arithmetic/remainder mutations; native/interpreter behavior is compared with fixed answers | Phase 27 retains the refusal/evidence contract; U64-02 closes with phase verification |
| Scalar CFG back edges | Independent finite scalar fixed points and dynamic occurrence identity | Phase 26 admits U64/Bool state, pins 0/1/2 events and forged controls, and fails closed at analysis/evidence bounds; resource/loan carries stay refused; path oracle stays acyclic | Phase 26 complete; preserve refusal controls in Phase 27 |
| Byte views/indexing/aggregates | Bounds, initialization, subobject layout/ownership and error paths | Empty/one/max/truncated cases; out-of-bounds controls; only admitted moves/copies | Byte-library capability |
| JSON/HTTP input | Untrusted bytes, limits, framing/encoding, allocation/error cleanup | Malformed input, depth/size budget, split/truncated messages; adversarial cases from selected protocol | Respective library consumer |
| Separate compilation | User-declared facts not verifiable inside producer; contract compatibility | Real B1 boundary witness before DX-06 closure; modules alone do not establish it | Separate-compilation capability (D-13-02b) |
| FFI expansion/async | Unwind, partial acquisition, cleanup failure, suspension/cancellation | Define consumed/retained/unknown resource states and error precedence first | Future selected consumer; refused in M004 |

Security and low-level control remain explicit: local C inputs are build
authority, foreign declarations are trusted assertions, external inputs have
limits, and ordinary pointers do not authorize unchecked Schway aliasing. This
roadmap does not promise to prove arbitrary foreign C correct.

## Current three recommendations

1. **Finish Phase 27's exact FizzBuzz and cross-host evidence.** User-visible
   programs: ordinary `fizzbuzz.schway` and `sum_to_n.schway` produce exact
   bounded output. Current blocker: actual macOS and Linux native receipts at
   the same final revision, plus the linked clean-checkout guide. Smallest
   complete slice: implement the 27-09 host driver, receipt gate, and existing
   CI matrix wiring; then complete 27-10's guide and collect the actual hosted
   artifacts. Checker/guarantee changes: require both distinct hosts, exact
   source/fixture hashes, declared inputs, and successful jobs; unavailable or
   stale evidence stays open. Preserve the implemented static-zero refusal,
   dynamic checked failure, shared output ceiling, and resource/loan loop
   refusals. Evidence/debt: Phase 26 U64-01/FLOW-01/FLOW-02/APP-07 and the
   Phase 27 focused local answer, mutation, output, and CLI checks are current;
   Phase 27 EVD-12 and DX-16 remain open. Owner/next action: execute Plans
   27-09 and 27-10; historical Phase 25/26 receipts do not count. Reprioritize
   by reacquiring both hosts if code, compiler, fixture, or CI inputs change
   after the claimed SHA, or leave EVD-12 open if either host is unavailable.
2. **Build a bounded byte utility and reusable local module boundary.**
   User-visible program: a `checksum` command reads a size-limited file and
   computes one specified checksum with exact success/error results. Current
   blocker: byte views, lengths/indexing, required checksum operations, and
   local modules are absent; scalar loops now exist, but resource/loan carries
   remain refused. Smallest complete slice: name one consumer and algorithm,
   specify empty/maximum/truncated behavior, then admit only its byte operations
   and an import used by a second consumer. Checker/guarantee changes: bounds,
   initialization, necessary scalar operators, borrow lifetime around iteration,
   cleanup on errors, and module visibility. Evidence/debt: Phase 24 RES-06
   cleanup, Phase 25 pointer controls, Phase 26 exact-loop and refusal tests,
   plus new byte-boundary and wrong-result controls; the provisional checksum
   fixture is still a refusal witness. Owner/next action: post-M005 consumer
   selection. Reprioritize if a real file tool needs a smaller byte API first.
3. **Enable one bounded JSON configuration consumer when a real need is named.**
   User-visible program: validate a size-limited config against one fixed schema,
   with exact success and typed-error results. Current blocker: byte indexing,
   aggregates, a bounded parser stack, and reusable modules remain unavailable;
   scalar loops alone do not supply them. Smallest complete slice: choose a
   consuming tool, define schema and byte/depth limits, and implement only its
   parser forms with failure cleanup. Checker/guarantee changes: bounds and
   initialization, parser-state ownership, borrow lifetime during iteration,
   and owned-buffer cleanup on every error. Evidence/debt: Phase 24 RES-06,
   Phase 26 scalar/refusal controls, and consumer-specific malformed, truncated,
   oversized, depth, and schema cases with independent answers. Assign new
   requirement/evidence IDs at kickoff; make no general JSON completeness claim.
   Owner/next action: post-M005 planning after a named Schway tool needs JSON.
   Reprioritize if another format or a smaller bounded representation serves
   the actual consumer.

For each candidate, the smallest complete slice, checker changes, evidence,
owner/next action, and reprioritize observation determine whether its ranking
should change.

M005 has delivered `sum_to_n` and the Phase 27 source path for exact FizzBuzz.
Checked addition, `<`, Bool branches, scalar fixed points, equality/remainder,
fixed text, and shared output limits are implemented. Phase 27 still closes the
cross-program evidence/documentation contract with its native host driver,
same-revision hosted receipts, and clean-checkout guide. Broaden loop-carried
state only for a named runnable witness with independent safety evidence.

## Maintaining pace and truth

**2026-10-08 source-review amendment.** Relevant dormant research hypotheses are indexed in
[`seeds/`](seeds/) and the [Dashbit AI-era appraisal](sources/DASHBIT-AI-ERA.md).
At the existing transition checks, revisit only seeds matching a named
consumer and current source witness; record their disposition in the active
planning artifact. The appraisal does not change the ranked recommendations
above.

- At kickoff, before phase discussion/planning, and after phase completion,
  inspect source witnesses and refusal boundaries, then refresh this document
  and LANGUAGE-MATURITY only where facts changed. Surface three ranked next
  capabilities without waiting for a prompt (procedure in AGENTS.md).
- Every feature phase names a source/input/output witness and expected result
  before plans expand. A refused frontier is a starting point; acceptance
  requires the intended behavior, not merely a moved first diagnostic.
- Allow at most one consecutive enabling phase without a runnable gain. A
  second triggers scope review. When a phase exceeds its initial plan estimate,
  identify the specific witness/safety obligation before adding more plans.
- Record plans and effort per runnable gain, remaining blockers, and measured
  cold/warm feedback distributions. Avoid percentage estimates of language
  completeness. M003 delivered 85 plans against research forecasts of 40–57;
  that variance justifies tighter slices, not removing correctness evidence.
- Use focused evidence during edits and full required lanes at integration
  boundaries. Assign each expensive CI lane one owner. The known duplicate
  full/race/vet work is a measured-maintenance candidate, not kickoff scope.
- Keep source inspection, executed checks, and historical receipts distinct.
  macOS-only evidence cannot close a Linux claim; missing evidence keeps its
  family incomplete. Do not rerun completed Phase 21 UAT to rewrite history.
- At milestone close, review this arc and promote only the next useful scope.
  The charter may change; record why, successor ownership, and affected claims.

## Provenance

**2026-10-02 M005 kickoff amendment.** M004 Phase 25 is complete with hosted
dual-host evidence recorded in the closeout section above. The user accepted
M005 Practical Computation: caller-run `sum_to_n`, then FizzBuzz, with explicit
U64 arithmetic/overflow, scalar-loop fixed points, bounded text output, and
independent interpreter/native evidence. Source inspection at kickoff confirms
arithmetic is absent and CFG cycles remain refused; the checksum example
remains a refusal witness. This is a planning transition, not newly executed
compiler evidence. Keep the dependency tree shallow and add a dependency only
when a concrete consumer need earns it.

**2026-10-01 Phase 24 closeout amendment.** Phase 24 completes bounded
PathToken/file-byte owner transfer through a helper and return, distinct
repeated activations, and reverse cleanup on success and a real typed error.
The verifier passes 5/5 roadmap truths. Hosted run 36856048690 passed the full
and focused suites on Linux/x86_64 and Darwin/arm64 at source SHA
`ed94ef7972b25deb85ab90fafbf3403dc31629f5`; focused Phase 24 receipts were
25s and 22s. No local project tests or scripts were run. The source still
refuses by-pointer function bodies; Phase 25 owns separate shared/exclusive
pointer witnesses and the integrated utility.

**2026-10-01 Phase 25 Plan 06 amendment.** Source inspection confirms the
production emitter admits only the bounded shared and exclusive U64 read-copy
helpers used by `examples/phase24/transfer.schway`; their manifests make no
unsupported alias, capture, alignment, or ownership promise. Mutation,
forwarding, retention, callbacks, nonlocal exits, and wider pointer forms
remain refused before C serialization. Newly executed local Plan06 checks—the
focused native contract tests and `sh scripts/verify-phase25.sh`—passed on this
macOS/arm64 host with Go 1.24.0 and Apple Clang 21.0.0 at source revision
`f00cdf8` with a modified working tree. Baseline `-O0`, optimized `-O2`, and
ASan+UBSan native utility runs returned 65/66 and preserved the typed 0x43
failure before either helper. Reached wrong-result controls, separate
conflict/escape controls, and pointer manifests also passed. Across the three
lanes, cold Go-cache min/median/max was 2.44/2.72/2.73s and warm-cache
min/median/max was 0.27/0.43/0.53s. The full local test suite, race suite,
`go vet ./...`, and `go build ./...` also passed on macOS/arm64. The Phase 25
script marks every Linux family/lane row incomplete; these local checks are
not hosted receipts. Hosted macOS and Linux Phase 25 family-by-lane receipts
remain pending, so no hosted pass is claimed and EVD-10 remains open.
Historical receipts are the Phase 24 run 36856048690 and earlier Phase 23
runs, scoped to their archived source revisions and behaviors.

**2026-10-02 Phase 25 security-audit follow-up.** Source inspection at
`b19446f` confirms that the command gate invokes core, origin, and path
validation in precedence order. The independent path oracle derives
straight-line borrow ancestry, last use, conflict families, and terminal
escape without reading endpoint claims; the origin peer supplies structured
function/return identity for source projection. Newly executed local tests
`TestPhase25PointerPathLiveOverlap`,
`TestPhase25PointerPathBorrowedResultEscape`,
`TestPhase25IndependentPeerMutations`,
`TestPhase25CheckCommandPeerGate`, `TestPhase25CheckCommandPeerObservation`,
`TestPhase25EscapeDiagnosticSourceAttribution`, and
`TestPhase25EscapeDiagnosticFunctionIdentity` pass, along with the retained
conflict/schema/boundary controls. Historical Phase 25 native results remain
scoped to source revision `f00cdf8`; this follow-up ran no native or hosted
matrix. That statement describes the follow-up only and is superseded for
current status by the hosted evidence amendment below.

**2026-10-02 Phase 25 hosted-evidence closeout.** GitHub Actions run
`36971855722` tested branch head
`421b5b94eb867e940a207dfd64d7971dc8198172` at PR merge revision
`a90c27c5b432ef6fc59fbafaa68b50a1374ae138`. Linux ran Go `linux/amd64` on
`x86_64-pc-linux-gnu` with Ubuntu Clang 18.1.3; macOS ran Go `darwin/arm64` on
`arm64-apple-darwin25.6.0` with Apple Clang 21. On each host, foreign, shared,
and exclusive families each passed baseline `-O0`, optimized `-O2`, and
ASan+UBSan: 18/18 distinct native rows. Every row returned expected/actual
65/66 and the typed `0x43` use error before either helper. The `checks` and
`current evidence aggregate` jobs passed on both hosts; checks included vet,
build, full Go tests, and race tests. Cold min/median/max feedback was
9.350/9.540/9.980s on Linux and 12.470/12.670/12.930s on macOS; warm was
0.390/0.490/0.550s and 0.820/1.400/2.130s respectively. A single-host script's
absent-other-host rows describe that invocation only; the paired native job
passed the other host's rows. This is newly executed hosted evidence, distinct
from source inspection, local macOS checks, and historical receipts. EVD-10 is
closed. The admitted pointer behavior remains bounded to read/copy helpers in
the exact utility chain, with independent borrow overlap and escape refusal;
mutation, forwarding, retention, callbacks, nonlocal exits, wider shapes, and
unsupported attributes remain refused.

**2026-09-30 Phase 23 closeout amendment.** Phase 22's verifier passes 5/5
roadmap truths and its completed objective README UAT is preserved. Phase 23's
refreshed verifier also passes 5/5, closes the hosted Ubuntu receipt gap, and
records Phase 23 complete. Hosted run 36707529870 passed the Ubuntu/macOS
check suites and both evidence aggregates, including the Phase 23 script, at
source SHA `f991298b29779838a2b1a5c3cd5ac90aafcb84fc`. No local project suites
or human UAT were rerun/needed for Phase 23 closeout. The earlier same-date
pre-closeout snapshot is historical; 2026-09-27 and 2026-09-28 observations
below remain true at their recorded revisions.

Repository research baseline: `d9bde05`; research date 2026-09-27. Current
observations are refreshed after Phase 22 from source inspection, its named
tests, and the full Go suite. The refreshed verifier at 2026-09-27T23:05:37Z
reports 17/17 truths and `passed` after the objective README contract UAT. No
subjective readability claim is made. These are macOS
observations; Linux remains a separate required host lane. Roadmap rows are
proposed future acceptance, not evidence of implementation. Detailed source
paths, official ecosystem references, and adversarial findings are in
[research/M004](research/M004/SUMMARY.md).

## Phase 33 transition amendment — 2026-10-13

Phase 33 implements bounded capture ingestion, owner/reason-scoped interpreter defects, generic freeze/receipt gates, and grounded carried-debt dispositions. Focused current checks and reached mutants are recorded in its five summaries; the review follow-up refuses malformed control lists. Phase 33 binds no hosted receipt. M006 receipts are historical; current hosted evidence resumes in Phase 38.

The current three capabilities are: agent diagnostics plus `schway surface --json` (Phase 34), bitwise/literal-shift/hex computation for CRC-32 (Phase 35), and scalar helper calls used by the CRC-32 module (Phase 36; integrated consumer Phase 37). Their named programs, blockers, smallest slices, checker obligations, evidence owners and reprioritization observations are in [33-TRANSITION-CHECK.md](phases/33-carried-debt-bounded-capture-read-and-generic-freeze-gate/33-TRANSITION-CHECK.md). Earlier recommendation sections remain historical; this dated amendment owns the current order.

All three conditional seeds remain dormant. The capture-allocation concern in the Dashbit source appraisal is resolved by `readApplicationCapture`, `TestPhase33CaptureReadBounded`, and the reached unbounded-read mutant. Current source search, capture, observers and replay supplied the diagnosis; no broader query, observability or source-form experiment is promoted.

## Phase 34 transition amendment — 2026-10-13

This post-execution review updates the ordered next capabilities from the current source witnesses and focused checks. The runnable diagnostic and source-free discovery slice is implemented; its report locations still derive only from captured source and do not grant repair authority. The next two capabilities remain Phase 35 and Phase 36/37 work, with the absent operations and helper-call boundary stated below. Reprioritize only when an admitted boundary, independent cause, or named consumer changes the evidence.

| Program | Status | Blocker | Smallest complete slice | Checker/guarantee | Evidence/debt | Owner/next action | Reprioritization observation |
|---|---|---|---|---|---|---|---|
| Broken CRC/word-counter check and schway surface --json | implemented | None remaining for this bounded diagnostic and discovery slice | Keep the runnable broken-program check and source-free public surface discovery usable from one checkout; preserve refusal behavior when source is unavailable | Captured source supplies diagnostic locations; byte identity remains unchanged and repair authority remains the exact byte span and replacement, with cross-unit edits declined | DX-20; `internal/compiler/session/session_phase34_routes_test.go#TestPhase34SourceRoutePresentation`; `internal/compiler/session/session_phase34_surface_test.go#TestPhase34SurfaceManifestWitnesses`; `internal/compiler/session/session_phase34_prechange_test.go#TestPhase34PreChangeCProgramGolden`; `internal/compiler/session/session.go#BuildApplicationCommandFile`; `cmd/schway/main.go#runSurface` | Phase 34 | Reconsider if an admitted source route loses its snapshot or an independent cause changes the diagnostic contract |
| CRC-32 operator expression x ^ (y << 3) & 0xFF and Adler b << 16 | planned | bitwise operators and shifts are absent | Add literal shifts and independent vectors for the named bitwise forms | Prove literal operands, range bounds, and safe shift width before emission; retain refusal for dynamic shifts | BIT-01; EVD-17; CRC-32 independent vectors | Phase 35 | Reprioritize if a named consumer needs dynamic shift or different bounds first |
| Imported crc_step scalar helper inside a loop | planned | scalar helper calls are absent | Add an arity-one imported helper and one file consumer with an independent answer | Require arity-1, acyclic call structure, and result-defect behavior across checker and execution peers | CALL-01; APP-13; MOD-07; independent imported-helper result and defect cases | Phase 36; file integration in Phase 37 | Reprioritize if a second consumer or arity-N call is required |

| Seed | Disposition | Current source witness | Named consumer | Trigger/reason | Next action |
|---|---|---|---|---|---|
| z1g | retain dormant | `internal/compiler/session/session_phase6_query.go#QueryCommandFile`; `internal/compiler/session/session_phase6_explain.go#ExplainCommandFile` | crc_step declaration and all direct call occurrences | measured named query gap | Let this seed remain dormant until a named reviewer task shows a measured gap in current query plus ordinary search; then scope one source-unit query |
| trz | retain dormant | `internal/compiler/native/native_app.go#readApplicationCapture`; `internal/compiler/native/native_phase33_capture_test.go#TestPhase33CaptureReadBounded`; `internal/compiler/session/session_app_verify.go#VerifyApplicationCasesFile` | one named checksum/file utility failure | failure unexplained by current capture, observers, and replay | The old allocation concern is historical and was resolved in Phase 33 by the bounded read and reached unbounded-read mutant; remain dormant until a named failure remains unexplained |
| z1s | retain dormant | `internal/compiler/syntax/format.go#Format` | CRC/word-counter canonical source review | named canonical-source or contract dispute | Let this seed remain dormant until a named consumer or reviewer identifies a concrete canonical-source or contract dispute |

| Kind | Status | Evidence/claim |
|---|---|---|
| Source inspection | inspected | Read `internal/compiler/session/session.go`, `cmd/schway-repair/repair.go`, `.planning/PRODUCT-ROADMAP.md`, `.planning/LANGUAGE-MATURITY.md`, the three named seeds, and the Dashbit source appraisal; source inspection does not establish execution behavior |
| Focused execution | passed | Newly executed focused selectors: `TestPhase34SourceRoutePresentation`, `TestPhase34DiagnosticRouteInventory`, `TestPhase34RepairJSONCompatibility`, `TestPhase34CompatibilityDeltaLedger`, `TestPhase34PreChangeCProgramGolden`, `TestPhase34PreChangeCGoldenIsNotInert`, `TestPhase34BoolNameBaseline`, `TestPhase34IdentityBaseline`, `TestPhase34SurfaceManifestWitnesses`, `TestPhase16PublicEmitterConsumerInventory`, `TestPhase321RunInputRequiredCLI`, `TestPhase272AcceptedCorpusByteIdentity`, `TestPhase273ConformanceCorpusByteIdentity`, `TestPhase28PrunedEmitMatchesGolden`, and `TestEmittedCExitLiteralsMatchTable`; Phase 34 document controls are checked against the committed amendment |
| Historical M006 receipts | historical | M006 receipt records remain historical evidence scoped to their original revision and hosts |
| Phase 38 hosted binding | pending | Phase 38; not yet bound |
| Human readability | unmeasured | No human review; readability remains unmeasured |
| Whole-loop latency | unmeasured | No cold/warm whole-loop distributions; end-to-end latency remains unmeasured |

## Phase 34 completion review — 2026-10-13

The source-review follow-up corrected manifest route applicability and aggregate outer-body binding scope. The newly executed TestPhase34SurfaceLimitConsumers, TestPhase34InterfaceLimitWitnesses and TestPhase34LinearBindingLimitScope provide finite consumer, executable interface-threshold and direct AST-guard evidence; see the [review-fix record](phases/34-agent-diagnostics-floor-and-surface-manifest/34-REVIEW-FIX.md). Parser/checker admission and enforcement are unchanged. The next three runnable capabilities and retained dormant-seed dispositions are now recorded in [34-TRANSITION-CHECK.md](phases/34-agent-diagnostics-floor-and-surface-manifest/34-TRANSITION-CHECK.md): Phase 35 bitwise/CRC expressions, Phase 36 imported scalar helpers, and Phase 37 file CRC, second-consumer reuse and CLI interpreter parity. Final hosted binding remains Phase 38; human readability and whole-loop latency remain unmeasured.

## Phase 35 discussion review — 2026-10-13

Source inspection corrects the kickoff's hex estimate: the lexer and scalar checker already admit both hex prefixes, but legacy linear conversion lacks uppercase-prefix support. Phase 35 completes and independently proves that contract while preserving existing decimal/binary route policies. The specified bitwise expression also requires bounded scalar grouping, absent at this baseline. The causal execution peer validates observed events rather than pure numeric answers; the [Phase 35 context](phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-CONTEXT.md) records the source-grounded evidence clarification and independent result controls before planning. These are discussion findings, not executed feature evidence.

The current three capability slices remain bitwise main/Adler (Phase 35), imported scalar helper (Phase 36), and file CRC with second importer and public CLI interpreter parity (Phase 37). Their blockers, guarantees, evidence owners, reprioritization observations and retained dormant seed dispositions are scoped in that context. No seed trigger was established. Final hosted binding remains Phase 38; human readability and whole-loop latency remain unmeasured.

## Phase 35 execution transition amendment — 2026-10-14

Source inspection at `83c01058` confirms that `examples/checksum/adler32.schway`
replaces its sixteen doublings with `b << 16` and retains the final addition;
the existing `b <= 65520` bound makes the shift exact: the largest shifted
term is 4,293,918,720, and adding the maximum `a` of 65,520 yields
4,293,984,240. Fresh local selectors
after the Adler change passed both imported consumers on interpreter and native
runners, including seven independent answer rows per consumer and a reached
copied `b << 15` wrong-result control (`150143896` versus restored `300286872`).
The new operator guide is checked against the current manifest and the actual
public and application witnesses. These local receipts are not the phase-wide
verification, security review, code review, hosted binding, human readability
review, or a whole-loop cold/warm latency distribution; those remain pending or
unmeasured.

| Program | Status | Blocker | Smallest complete slice | Checker/guarantee | Evidence/debt | Owner/next action | Reprioritization observation |
|---|---|---|---|---|---|---|---|
| `examples/bitwise.schway` public computation and imported Adler checksum | Implemented and locally witnessed; Phase 35 plan/phase closeout remains separate | Phase-level independent verification, security review, and code review remain pending; hosted binding belongs to Phase 38 | Keep the witnessed grouped bitwise expression and literal shifts alongside the one-line Adler consumer used by `checksum` and `adler_low` | Checked U64 bitwise operators; literal shift range 0..63; low-64-bit left shift and logical right shift; bounded expression nesting; unchanged Adler answer and cleanup | BIT-01..04, D-07/D-18/D-20, EVD-17; `TestPhase35ExistingAdlerAnswers`, `TestPhase35AdlerWrongShiftControl`, public 264/alternate 8, and application `264\n` witnesses; see the Phase 35 plan evidence artifact when closed | Phase 35 executor completes scoped plan artifacts; parent verifier/review owns phase acceptance; Phase 38 owns hosted evidence | Reprioritize if a named consumer requires dynamic shifts or a different bound |
| Imported `crc_step` helper used by a checksum loop | Planned for Phase 36 | Scalar helper calls are absent from scalar expressions; imported call/result behavior still needs independent checker and execution evidence | One arity-one acyclic imported U64/Bool helper and a direct call in the loop consumer | Validate arity, import identity, acyclic calls, argument/result types, and independently mutated call/result defects | CALL-01/02; Phase 36 requirement and helper call/defect controls; Phase 37 consumer evidence | Phase 36 owner plans and implements the helper; Phase 37 integrates it into the file utility | Reprioritize if a named consumer requires arity-N calls or a second call shape |
| `crc32 FILE`, a second importer, and interpreter CLI parity | Planned for Phase 37 | Imported helper support and file input through the interpreter CLI are not yet present | One ISO-HDLC CRC-32 module, bounded file consumer, second importer, and a guide with literal vectors | Preserve module boundaries and file limits; independently check CRC vectors, failure outcomes, mutation behavior, and CLI parity | APP-13/14, MOD-07, DX-24; independent CRC vectors, wrong-result mutation, and file/second-consumer controls | Phase 37 owns the runnable integration; Phase 38 binds hosted evidence | Reprioritize if an actual file consumer requires a different bounded input or failure contract |

| Seed | Disposition | Current source witness | Named consumer | Trigger/reason | Next action |
|---|---|---|---|---|---|
| z1g | retain dormant | `internal/compiler/session/session_phase6_query.go#QueryCommandFile` | `crc_step` declaration and direct call occurrences | No measured query/search gap is present | Keep dormant until a named reviewer task measures a gap in current query plus ordinary search; then scope one source-unit query |
| trz | retain dormant | `internal/compiler/native/native_app.go#readApplicationCapture`; `internal/compiler/native/native_phase33_capture_test.go#TestPhase33CaptureReadBounded`; `internal/compiler/session/session_app_verify.go#VerifyApplicationCasesFile` | One named checksum/file-utility failure | No unexplained failure is present; the earlier allocation concern was resolved in Phase 33 | Revisit only after a named failure remains unexplained by bounded capture, observers, and replay |
| z1s | retain dormant | `internal/compiler/syntax/format.go#Format` | Canonical bitwise/CRC source review | No named reviewer dispute is present | Revisit only when a consumer identifies a concrete canonical-source or contract dispute |

| Evidence kind | Status | Receipt |
|---|---|---|
| Source inspection | inspected | Adler bound and rewrite, current guide, manifest constraints, seed witnesses, and Dashbit appraisal were inspected; these observations do not establish execution behavior |
| Fresh local execution | passed | On Go 1.24.0, Apple Clang 21.0.1, macOS/arm64, checkout revision `26380801`: `TestPhase35ExistingAdlerAnswers` and `TestPhase35AdlerWrongShiftControl` passed after the source rewrite; `TestPhase35OperatorDocumentation` and `TestPhase35PublicDocumentValues` passed with their copied prose controls; `TestPhase35PublicConformanceWitness` passed on interpreter/native for 264 and alternate 8; `TestPhase35FocusedValueAnswers` passed with application output `264\n` on both engines. Each selector was run with `-count=1 -v -timeout=600s`; native paths ran without skips. |
| Historical Phase 35-07 execution | historical | At its recorded revision `64dfec6f`, the committed 33-row fixture ran at O0/O3/O3-LTO; all 16 shift rows ran UBSan. This is historical and is not presented as a fresh receipt for this amendment. |
| Hosted evidence | pending | Phase 38; no current hosted receipt is bound |
| Human readability | unmeasured | No human review was run |
| Whole-loop feedback latency | unmeasured | No cold/warm end-to-end distributions were collected |

## Phase 35 closeout transition review — 2026-10-14

The implementation and independent repair work now support the bitwise/Adler slice. Phase acceptance is owned by [35-VERIFICATION.md](phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-VERIFICATION.md), rather than the earlier implementation-only pending notes. The [transition check](phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-TRANSITION-CHECK.md) records current source witnesses, the 94-name regression refresh, the target-freshness repair, 26 admission mutation controls and all evidence boundaries.

The next three scoped capabilities are: Phase 36 imported `crc_step` calls in a loop; Phase 37 bounded file CRC-32 with a second importer and interpreter CLI parity; Phase 38 reproducible two-host checksum/CRC receipts. The linked transition table names each program, blocker, smallest slice, checker obligations, evidence/debt, owner and priority-changing observation. All three conditional seeds remain dormant on their named source witnesses. Historical M006 receipts remain historical; final hosted binding, human readability and whole-loop cold/warm distributions remain pending or unmeasured.

Phase 35 final acceptance update (2026-10-14): canonical verification now passes 38/38 unique truths; all 33 vectors and 16 UBSan shift rows were freshly rerun locally, together with 15 arithmetic and 26 admission mutation controls. The earlier implementation-only pending notes remain historical. Security/review are closed; final two-host receipts remain Phase 38-owned.

## Phase 36 discussion transition — 2026-10-14

The [Phase 36 context](phases/36-scalar-helper-calls/36-CONTEXT.md) records three
specialist source-research passes and delegated decisions. Current code has
direct calls and local imports, but scalar-expression calls, Bool
formals/results and repeated call activation handling remain absent. Existing
invocation ordinals are always zero and native evidence reservation counts
static operations, so enabling a call in the scalar builder alone would not
complete the loop witness. Keep bounded evidence capture separate from ordinary
application execution. The source-grounded CALL acceptance clarification in
REQUIREMENTS preserves the old byte-bounds callee regression and causal/numeric
proof distinction without widening U64/Bool helper arguments to ByteView.

The current next three capabilities are Phase 36 imported `crc_step` calls in a
loop, Phase 37 bounded file CRC with second importer/interpreter CLI/guide, and
Phase 38 final two-host receipts after audit. Their concrete programs, blockers,
smallest slices, checker obligations, evidence/debt, owners and priority-changing
observations are in the context's transition table. Seeds z1g/trz/z1s remain
dormant on their named source witnesses; no measured navigation gap,
unexplained failure or real source-form dispute was established. This is source
inspection and primary-document research, with no newly executed compiler
behavior evidence. Earlier receipts retain their recorded scope; human
readability and full cold/warm feedback-loop distributions remain unmeasured.

## Phase 36 completion — 2026-10-15

The local and imported reflected-bit helpers are implemented and exercised through the public command, interpreter, and native application routes. The dated [acceptance evidence receipt](.planning/phases/36-scalar-helper-calls/36-ACCEPTANCE-EVIDENCE.md) binds claim rows to fixed owner captures at source revisions recorded in their manifests. CALL-01/CALL-02 remain unclassified because no SPEC resolves their edge classification; descriptorless safety prohibitions remain unverified. Phase 36 still awaits its independent verifier, security, and code-review gates. No hosted evidence, human readability result, or complete-loop cold/warm distribution is claimed here.

Source inspection: The checked helper contract is in `internal/compiler/check/scalar.go`; independent execution and causal boundaries are in `internal/compiler/interp/interp.go` and `internal/compiler/executionpeer/executionpeer.go`; the runnable examples are `examples/helper_calls/crc_step.schway` and `examples/helper_calls/importer.schway`.

New focused checks: The current amended final owner capture at full HEAD `979357388171339005fd6dbd65aa760810ae9384` freshly records the named claims in the linked acceptance receipt, including `internal/compiler/session/session_phase36_calls_test.go#TestPhase36HelperAnswersAndDefects`, `internal/compiler/corevalidate/corevalidate_phase36_test.go#TestPhase36RawCallShapeAndFreshness`, `internal/compiler/executionpeer/executionpeer_phase36_test.go#TestPhase36CallCausalAttributionControls`, and the corrected emitter census. It passed 45/45 roots across 12 owner groups; actual logs are linked from the receipt. The prior-phase regression evidence remains a separate 80-root/13-command capture at source/test revision `b51043c6566f793893c8c4e143493e7ec3af5bcd` ([36-REGRESSION.md](phases/36-scalar-helper-calls/36-REGRESSION.md)); source files are unchanged from that revision through the current capture, so it was not rerun. The final 52-file code review has zero open dispositions at `97935738` ([review report](phases/36-scalar-helper-calls/36-REVIEW.md), [disposition](phases/36-scalar-helper-calls/36-REVIEW-DISPOSITION.md)). Earlier captures and the prior security record remain bound to their recorded revisions.

Historical receipts: Historical only; not fresh Phase 36 evidence. [Phase 35 verification](.planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-VERIFICATION.md) passed on 2026-10-14 for its own source revision and host.

| Phase | Program | Blocker | Smallest slice | Guarantee/evidence | Owner/next action | Priority changes if |
|---|---|---|---|---|---|---|
| Phase 36 | Repeated reflected-bit `crc_step` through a local helper, an importer, and the public application route | Implementation evidence is present; independent phase verification, security, code review, and unresolved requirement classification remain | Close the already implemented arity-one U64/Bool call slice across source lowering, raw peers, interpreter/native results, and repeated causal capture | Checker lowering, independent raw call contracts, literal helper answers and overflow/remainder defects, causal-only attribution, the distinct old/new ByteView bounds obligations, and frozen no-call C are tied to named rows in the acceptance receipt | Phase 36 audit owners complete goal-backward verification and security/review; keep requirements pending until their gates pass | A named consumer requires arity-N or another authority shape, or an independent owner finds a reached call/result defect |
| Phase 37 | Bounded file CRC-32 with a second importer, interpreter file CLI, and executable guide | No complete file-integrated CRC consumer, second importer, or interpreter file route exists yet | Compose the current helper with bounded file input, one independent second importer, CLI parity, and a guide using independent vectors | Check bounded read/refusal behavior, module identity, interpreter/native answer parity, file failures, and reached wrong-result controls; APP-13/14 and MOD-07 remain the evidence owners | Phase 37 owner plans the smallest complete runnable consumer after Phase 36 audit closes | A named file consumer requires a different input bound, call signature, or failure contract |
| Phase 38 | Final two-host checksum/CRC acceptance bound to one source revision | No current M007 hosted two-host receipt is bound | After Phase 36/37 audits, run the exact final acceptance on Ubuntu and macOS and bind both logs to the same committed source revision | Require independent host results, exact source/fixture identity, no skipped roots, and preserved application outcomes; hosted evidence remains pending | Phase 38 owner prepares and records the final hosted binding after the audit owners hand off | Host drift, a source revision mismatch, or a platform-specific result changes the tested guarantee |

| Seed | Disposition | Source witness | Trigger observation | Next action |
|---|---|---|---|---|
| z1g | retain dormant | `internal/compiler/session/session_phase6_query.go#QueryCommandFile`; current helper declaration and callsites remain available to ordinary search | No named reviewer task has measured an accuracy or material-cost gap in query plus ordinary search | Keep dormant until a named reviewer measures a concrete lookup gap; then scope one source-unit query |
| trz | retain dormant | `internal/compiler/native/native_app.go#readApplicationCapture`; `internal/compiler/session/session_app_verify.go#VerifyApplicationCasesFile` | Current bounded capture and replay explain the reviewed Phase 36 outcomes; the older allocation concern was addressed in Phase 33, and no new unexplained failure is named | Review one actual unexplained consumer failure before proposing additional observation |
| z1s | retain dormant | `internal/compiler/syntax/format.go#Format` | No consumer or reviewer has contested a concrete canonical-source or contract choice | Require a named dispute, paired source forms, and independent obligations before an experiment |

## Phase 36 post-verification transition — 2026-10-15

This amendment supersedes the pending-gate statements in the earlier Phase 36
implementation snapshot. All eight plans and CALL-01/CALL-02 are complete.
The independent goal verifier passed 32/32 must-haves; the two previously
descriptorless prohibitions now have real producer-backed RED/clean/default
proof. The 52-file independent review is clean with no open finding, its later
proof-bridge supplement is reviewed separately, and all 24 L1 threats are
closed. The classifier's unclassified SPEC-less rows remain planning history,
not a claim of untested runtime contracts.

The 45-root feature and 80-occurrence regression captures retain their exact
source revisions in the acceptance/regression receipts. Later work adds only
phase-local proof wiring and planning records. No hosted run, new backend,
compiler/runtime dependency, subjective readability result or whole-loop
performance measurement is claimed. Current source witnesses remain the
checked scalar transfer, the independent peers, interpreter/native execution,
and the local/imported examples.

The three next capabilities are full file CRC-32 plus a second importer
(Phase 37, APP-13/MOD-07), public interpreter file CLI plus an executable guide
(Phase 37, APP-14/DX-24), and final audited two-host receipts (Phase 38, EVD-17).
Their programs, blockers, smallest complete slices, checker obligations,
evidence/debt, owners and priority-changing observations are recorded in
[36-TRANSITION-CHECK.md](phases/36-scalar-helper-calls/36-TRANSITION-CHECK.md).
The same transition rechecked the Dashbit appraisal and relevant z1g/trz/z1s
triggers against named consumers and current source witnesses: retain all three
dormant, with no measured navigation gap, unexplained failure or source-form
dispute. Phase 37 discussion is next; it has not started.

## Phase 37 discussion transition — 2026-10-15

The [Phase 37 context](phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-CONTEXT.md)
captures the user-delegated specialist/adversarial synthesis. Source inspection
supports composing the existing ByteView/scalar-call surface without another
language feature: one nested `lib.crc32` module, a private eight-round scalar
byte helper and two importing file entries. The public interpreter route needs
a fixed production host reader and an explicitly evidence-disabled application
boundary; the existing file seam auto-captures scalar calls. The new path suffix
must dispatch before global option extraction so a flag-shaped path stays data.

New execution during discussion was limited to an independent Python/zlib oracle
calculation (maximum 65,536 x FF gives 3735780942), not Schway behavior. Prior
Phase 36 receipts remain bound to their recorded source. The context's transition
table names the three current programs/capabilities, blockers, complete slices,
guarantees, evidence/debt owners and priority-changing observations: complete CRC
reuse; public file-engine parity plus guide; audited two-host binding in Phase 38.
Relevant z1g/trz/z1s triggers were checked against current sources and remain
dormant. No runtime dependency, hosted receipt, human readability result or
whole-loop latency result is claimed by this discussion.

Research follow-up on 2026-10-15 exercised one temporary ordinary-source CRC
program: public check and native build/run produced `3421780262` for `123456789`.
The same bounded probe and an existing CLI test show that an empty PathToken
returns 65 with empty streams before the adapter runs. CONTEXT D-10 is amended
to preserve this behavior. This is a small constructibility probe, not a
committed Phase 37 test, maximum-file result or phase acceptance receipt.

## Phase 37 focused execution update — 2026-10-15

**Current status:** the bounded CRC consumers, both public file routes, and
tracked-source guide now have source-bound local evidence. Phase 37 remains
pending the independent parent verifier, security, review, and requirement
acceptance gates; this note does not mark the phase complete.

**Source inspection:** `examples/crc32/crc32.schway` and
`crc_low.schway` import the same table-free CRC-32/ISO-HDLC implementation.
The new file route retains the fixed interpreter adapter and existing
binding-specific native reader. The guide records exact exits, stream tuples,
the 65,536-byte interpreter limit, the formatter limitation, and the scope of
capture and timeout evidence.

**Newly executed local checks:** at code revision
`b900dfe31abb600a82b31ab6e8cc2a8b9b1e74ca`, the C01-C26 owner map was run on
Darwin/arm64 with Go 1.24.0; all 63 Go root invocations passed, the three
canonical proof subjects passed their bad/clean/default contract, and the
Phase 32 node-test adapter passed 3/3. The tracked-only guide copied 2,691
tracked files and ran 78/78 documented CRC tuples across both applications,
both engines, one-shot native, and retained native runs. The maximum-size
interpreter case completed in 2.46 seconds in one measured run; this is not a
distribution or latency guarantee. Per-command logs and file digests are in
[37-ACCEPTANCE-EVIDENCE.md](phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-ACCEPTANCE-EVIDENCE.md).

**Historical receipts and limits:** earlier baseline receipts remain at their
recorded source revisions; final C01/C02/C25 runs are regression checks, not
baseline replacements. Phase 35/36 and prior M006 hosted evidence stay
historical. No Phase 37 hosted result, local Linux execution, human readability
assessment, arbitrary timeout parity, or complete cold/warm feedback-loop
distribution is claimed. The no-edit native artifact digest and BuildID may
vary while module rows and InputID remain stable under the existing
incomplete/non-cacheable receipt contract.

The current next capabilities and their seven-column blocker, slice,
guarantee, evidence, owner, and reprioritization facts are in the
[Phase 37 transition check](phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-TRANSITION-CHECK.md).
It keeps z1g, trz, and z1s dormant because their named trigger observations
remain absent, and leaves M008 `wc`/`longest_line`, general arity expansion,
and Phase 38 hosted binding with their existing owners.

## Phase 37 accepted transition — 2026-10-15

This dated amendment supersedes the pending-gate statements above. Nine plans
and 21 tasks are complete; canonical verification passes 14/14 truths with
zero unverified behavior and no human-verification item. APP-13, MOD-07,
APP-14 and DX-24 are complete. Independent security re-audit closed 31/31
dispositions, validation is compliant and review has zero open findings.
The repaired private CRC helper owns eight rounds and is called once per
byte; structural activation evidence is separate from independent CRC values.
Five capture-growth and three reader mutations reached their named assertions.

New independent local checks on the repaired implementation include the
public expected-first matrix, actual native cleanup/ASan/UBSan, and 78 guide
tuples from a tracked-only copy. Source/data-flow inspection and exact commands
are in [37-VERIFICATION.md](phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-VERIFICATION.md);
the earlier implementation captures remain historical at their recorded
revisions. No dependency, backend, language construct or wire field was added.

The current three recommendations are audited same-source two-host receipts,
the existing M008 `wc`/`longest_line` landing, then a bounded key=value config
consumer before JSON. Their programs, blockers, smallest slices, checker
obligations, evidence/debt, owners and priority-changing observations are in
[37-CLOSEOUT.md](phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-CLOSEOUT.md).
The same artifact records the relevant seed review: z1g/trz/z1s stay dormant
without a measured navigation gap, unexplained failure or source-form dispute.
Only EVD-17 is current milestone scope; later programs need normal ownership.

Phase 33–36 implementation remains complete, but their whole-phase reports
are stale after successor changes; focused regressions do not refresh those
fingerprints. Reconcile the canonical verifier-only route (earliest Phase 33),
then run the milestone audit and any gap closure before Phase 38. This run
stops before Phase 38 planning, push or hosted dispatch. Local evidence is
Darwin/arm64; final two-host receipts, subjective readability and whole-loop
cold/warm latency distributions remain pending or unmeasured.


## Phase 33 verification refresh after Phase 37 — 2026-10-15

This dated amendment updates the earlier freshness routing. Phase 33 now
passes independent 60/60 verification with a fresh 59-file fingerprint. Source
inspection and reached controls close incomplete coverage metadata and receipt
record/inventory false accepts; the shared gate follows the producer's real
whitespace and free-text stderr grammar. Compiler/runtime semantics and the
existing language frontier are unchanged. Frozen receipts/goldens remain
historical at their original source.

The final parent run passed 49 exact roots across six packages and seven
metadata/ledger roots, with no missing/fail/skip; the verifier executed its own
critical paths and checked current log/input bindings. No hosted result,
dependency or latency distribution is claimed. Phases 34-37 implementation
remains complete but their reports are stale, including Phase 37 after the
metadata corrections. Resume with `$gsd-execute-phase 34`, preserve completed
plans/UAT, then audit M007 and close gaps before Phase 38.

The current three recommendations remain same-source audited two-host CRC
receipts, the proposed M008 wc/longest_line pair, then a bounded fixed-schema
key=value consumer. Their programs, blockers, complete slices, checker laws,
evidence/debt, owners and reprioritization conditions, plus relevant dormant
z1g/trz/z1s dispositions, are in [33-CLOSEOUT.md](phases/33-carried-debt-bounded-capture-read-and-generic-freeze-gate/33-CLOSEOUT.md).
Only EVD-17 is active milestone scope; no later consumer was promoted.


## Phase 34 verification refresh after Phase 37 — 2026-10-15

This dated amendment supersedes the earlier Phase 34 freshness pointer.
Phases 33 and 34 now pass current independent verification (60/60 and 38/38).
The Phase 34 report binds 84 current source/evidence files; all 22 validation
rows passed 42 distinct names / 45 package roots and the regression union
passed 107 distinct names / 110 roots across 11 packages, zero missing/fail/skip.
The verifier checked current input/log bindings and ran seven independent
focused command groups. Review has no new findings and the L1 register remains
19 closed / zero open. No implementation, dependency, language admission or
hosted receipt changed. Historical successor capabilities retain their owners.

The exact next gate is `$gsd-execute-phase 35`, followed by stale 36/37 reports,
the M007 audit and gap closure before Phase 38. Preserve completed plans/UAT.
The current three recommendations remain audited same-source two-host CRC
receipts, proposed M008 wc/longest_line, then bounded fixed-schema key=value
configuration parsing. Their programs, blockers, complete slices, checker laws,
evidence/debt, owners, reprioritization conditions and dormant z1g/trz/z1s
source/consumer dispositions are in [34-REFRESH-CLOSEOUT.md](phases/34-agent-diagnostics-floor-and-surface-manifest/34-REFRESH-CLOSEOUT.md).
Only EVD-17 remains active; human readability and whole-loop cold/warm latency
remain unmeasured. No future consumer or dormant experiment was promoted.
