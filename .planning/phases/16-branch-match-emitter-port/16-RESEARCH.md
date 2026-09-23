# Phase 16: Branch/Match Emitter Port - Research

**Researched:** 2026-09-19
**Domain:** Go-hosted C17 native emitter convergence, branch/match lowering, evidence provenance, and debt retirement
**Confidence:** HIGH

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

### One emission law and atomic cutover

- **D-16-01:** Use a narrow capability port: move `emitMatch`, `emitBranch`
  (including its arm lowering), and ordinary `emitLinear` behavior into
  `emitProgram`; then make `emitProgram` the only production route for every
  admitted program. Do not rewrite all six legacy emitters and do not retain an
  N=1 production fork. — **Reversibility: one-way** — the intended deletion is
  an architectural authority cut; undoing it would recreate two independently
  maintained lowering laws and force Phase 17 to pay its two-type model twice.
- **D-16-02:** Establish three pre-deletion convergence gates, stop at a
  `checkpoint:decision`, then flip dispatch and delete `emitMatch`,
  `emitBranch`, and `emitLinear` in the same commit. Preserve
  `TestN1ConvergenceDifferential` as the characterization test; flip its
  expectations instead of deleting it.
- **D-16-03:** Preserve Phase 15's validation order: graph/entry validation,
  supported-shape validation, invocation/output preflight, then serialization.
  Adding match support must not change the diagnostic precedence of remaining
  unsupported shapes.
- **D-16-04:** Port match payload lowering through the existing checker-derived
  layout facts. Preserve tagged-struct representation and the `_Noreturn`
  defect exemption; do not independently re-derive payload layout or mint
  generic optimizer attributes.
- **D-16-05:** `live_resources` must be derived by the surviving emitter. A
  hardcoded empty JSON literal is not admissible after the broader port.

### Narrow `restrict` probe

- **D-16-06:** Run the optional probe before planning. It is a test-only,
  hand-written C17 microprogram that matches only the prospective emitted
  shape: one translation unit; a `static T f(T *restrict p)`; one pointer;
  read/copy of `*p` only; caller local access only before or after the call.
  Run it on macOS and Linux at `-O0`, `-O3`, and `-O3 -flto`, with the applicable
  sanitizer lane.
- **D-16-07:** Admit `emitLinearBorrowedByPointer` only if the normative and
  measured probe proves the exact shape and the emitter structurally refuses
  extensions: no pointee mutation, second pointer, pointer escape/forwarding,
  callback, foreign call, volatile/atomic access, or separate compilation.
  Otherwise cut it to M004 with the full discharge-pair design. This decision
  never generalizes `restrict` from the present read-only case to future richer
  aliasing.
- **D-16-08:** A successful probe removes only the former
  zero-optimizer-visible-attribute rationale. It does not establish that LTO
  is behaviorally non-inert; that evidence remains a distinct obligation.

### Compatibility and golden-C provenance

- **D-16-09:** Before deletion, require the legacy and `emitProgram` paths to
  emit byte-identical C for exactly `testdata/phase1/toggle.lang`,
  `testdata/phase2/owned_transfer.lang`, and
  `testdata/phase3/borrowed_view.lang`, for both applicable `Emit` and
  `EmitNative` modes. Keep the established four-tier semantic comparator as
  independent defense in depth. Do not require identity of explicitly cut
  foreign or unadmitted by-pointer shapes.
- **D-16-10:** Create a checked-in, machine-linked four-entry golden-change
  ledger adjacent to `previousPhaseGoldenCDigests`. Each entry names the path,
  old and new SHA-256, responsibility moved from legacy to `emitProgram`, exact
  structural reason, semantic witness, N=1 fixture, and review disposition.
  CI must fail for missing, duplicate, stale, or digest-map-mismatched entries.
  A digest proves integrity, not the rationale for a generated-C change.

### Formal M004 cut

- **D-16-11:** Formally amend `REQUIREMENTS.md` under D-10-60's amendment
  clause for `emitLinearForeign`, `emitLinearBorrowedByPointer`, and
  `emitLinearBorrowedByPointerPlain` unless D-16-07 admits the first
  by-pointer family. Refile each excluded family as an M004 debt row with an
  owner, real prerequisite, reopening condition, and honest `-flto` consequence.
  Never silently relabel a third deferral.

### the agent's Discretion

Exact helper names, test factoring, ledger encoding, and diagnostic prose are
at the planner's discretion, provided they preserve the required scope,
validation order, identity gates, structural refusal boundaries, and evidence
linkage above.

### Deferred Ideas (OUT OF SCOPE)

None — the foreign resource ledger, richer by-pointer/discharge semantics, and
any remaining cut emitter family are explicitly routed to M004 rather than
being silently carried forward.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|---|---|---|
| NAT-08 | One emission law lowers every admissible program; the three superseded single-function emitters are deleted in the same dispatch-flip commit. | Dispatch seam, Phase 15 ordering, N=1 characterization, byte gates, and atomic cutover protocol below. |
| NAT-09 | Three non-M003-consumer emitter families are formally cut through an amendment naming their landing milestone and `-flto` consequence. | D-10-60 amendment rule, existing debt-register grammar, and M004 cut plan below. |
</phase_requirements>

## Project Constraints (from AGENTS.md)

- Use Go 1.24 standard library for Stage 0 and readable C17 emitted to installed Clang; preserve macOS and Linux portability. [VERIFIED: AGENTS.md:20-31]
- Prefer standard-library-only, shallow audited dependencies. This phase must add no package: its work is existing Go/C17 source, fixtures, and documents. [VERIFIED: AGENTS.md:24-25; go.mod:1-3]
- Preserve deterministic fixtures, negative controls, properties, mutation, differential execution, and sanitizer evidence as distinct cost lanes. [VERIFIED: AGENTS.md:26-27]
- Keep unsafe operations, FFI, build authority, and optimizer/FFI guarantees explicit; do not widen implicit authority. [VERIFIED: AGENTS.md:23,29-31]
- Planning artifacts are authorized under `$gsd-plan-phase`; implementation remains a later `$gsd-execute-phase` action. [VERIFIED: AGENTS.md:47-53]

## Summary

The surviving integration point is `emitProgram`. Its current implementation already validates call graph and entry, rejects unsupported body classes, preflights the Phase 15 invocation table and output size, and only then starts serialization. The port must add match-only, branch-plus-linear arm, and ordinary single-function linear capabilities *inside that law*, preserving the ordering before deleting the three legacy production emitters. [VERIFIED: internal/compiler/cgen/cgen_program.go:412-491]

The smallest safe decomposition is evidence first, capability port second, then one one-way commit. Keep both implementations while three gates prove exact output on the three admitted fixtures in both public modes, semantic agreement is maintained by the established interpreter/O0/O3/O3-LTO comparator, and the golden-change ledger is mechanically complete. The only deletion commit flips both `Emit` and `EmitNative` to `emitProgram` and removes `emitMatch`, `emitBranch`, and `emitLinear`; it must not touch foreign lowering or unadmitted pointer lowering. [VERIFIED: internal/compiler/cgen/cgen.go:112-169; internal/compiler/cgen/cgen_n1_convergence_test.go:43-116]

**Primary recommendation:** Plan an 8–10-plan tracer-first phase: establish exact convergence/provenance/debt gates; port ordinary linear then branch/match into `emitProgram`; stop at the required decision checkpoint; then atomically flip dispatch and delete only the three admitted legacy laws.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|---|---|---|---|
| Public dispatch convergence | API / Backend (Go `cgen`) | — | `Emit` and `EmitNative` own core validation and select the current N=1 versus whole-program route. [VERIFIED: internal/compiler/cgen/cgen.go:112-169] |
| Graph/shape/preflight ordering | API / Backend (Go `cgen`) | — | `emitProgram` owns callgraph/entry validation, supported-shape admission, bounded `/2` preflight, and serialization. [VERIFIED: internal/compiler/cgen/cgen_program.go:412-491] |
| Match and payload C lowering | API / Backend (Go `cgen`) | Native C17 | C is generated from checker-derived core facts; tagged-struct payload layout remains checker-owned. [VERIFIED: internal/compiler/cgen/cgen.go:1780-2180] |
| Runtime event and resource document | Native C17 | API / Backend | emitted C produces `lang.execution/2`; Go must derive rather than hardcode the resource tail. [VERIFIED: internal/compiler/cgen/cgen_program.go:585-615] |
| Semantic defense in depth | Test / evidence | Interpreter + native runner | Four tiers compare independently after generation; byte identity alone cannot prove runtime semantics. [VERIFIED: internal/compiler/session/session_phase11_differential_test.go:74-126] |
| Debt-cut enforcement | Planning records + test guard | API / Backend test | requirement amendment and M004 rows are human-readable authority, with existing debt parser enforcing row shape. [VERIFIED: internal/compiler/session/session_test.go:2555-2710] |

## Standard Stack

### Core

| Library / tool | Version | Purpose | Why standard |
|---|---:|---|---|
| Go standard library | `go 1.24` | compiler host, tests, SHA-256 ledger verification | Existing module declares exactly `go 1.24`; no new dependency is justified. [VERIFIED: go.mod:1-3] |
| C17 + installed Clang | host-installed | generated program and bounded probe | Project constraint and existing native test lanes use this boundary. [VERIFIED: AGENTS.md:23-25; internal/compiler/session/session_phase11_differential_test.go:104-116] |

### Supporting

| Tool | Purpose | When to use |
|---|---|---|
| `go test ./internal/compiler/cgen ./internal/compiler/core ./internal/compiler/session -count=1` | focused characterization, cgen, core pin, session/debt checks | Each task commit and before cutover. |
| `scripts/assert-go-tests.sh` | rejects a `-run` target that discovers no test | Use for every plan-specified focused command. [VERIFIED: .planning/milestones/M001-phases/06-agent-feedback-and-performance-ratification/06-01-PLAN.md:100-102] |
| Clang | `-O0`, `-O3`, `-O3 -flto`, and sanitizer probe lanes | Restrict microprogram and end-to-end native comparator. |

**Installation:** None. Do not add external packages.

## Architecture Patterns

### System Architecture Diagram

```text
checked core.Program
       |
       v
Emit / EmitNative
       |
       v
emitProgram
  graph+entry -> supported shape -> invocation/output preflight -> C serialization
       |                                                       |
       |                                                       v
       |                                              one C17 translation unit
       |                                                       |
       v                                                       v
named refusal                                           native O0/O3/O3-LTO
                                                               |
interpreter ------------------------------------------------> four-tier comparator
```

### Recommended Project Structure

```text
internal/compiler/cgen/
  cgen.go                  # public dispatch; delete three legacy emitters only at cutover
  cgen_program.go          # one surviving emitter and new lowering helpers
  cgen_n1_convergence_test.go
  cgen_program_test.go
  cgen_test.go
internal/compiler/core/
  core_test.go              # digest map and its adjacent provenance ledger/check
internal/compiler/session/
  session_phase11_differential_test.go  # four-tier defense-in-depth
.planning/
  REQUIREMENTS.md           # explicit M003 amendment
  phases/16-branch-match-emitter-port/PHASE-16-DEBT.md
```

### Pattern 1: Preserve the Phase 15 refusal sequence

Use one ordered `emitProgram` admission pipeline. The required order is graph/entry validation, supported-shape checks, invocation/output preflight, and serialization. Preserve this order when allowing `Match`; unadmitted foreign and remaining pointer families must still fail at their established boundary rather than reaching preflight or partially emitting C. [VERIFIED: internal/compiler/cgen/cgen_program.go:412-491]

### Pattern 2: Port behavior, not the legacy wrapper

Move ordinary linear operation emission into a reusable `emitProgram` function-body helper, then add branch/match arm lowering there. The existing `emitProgramFunction` has one per-function local allocator, one uniform C signature, operation loop, and `OpReturn` event path; `emitBranchOperations` is the direct behavioral source for arm operation and payload lowering. Avoid copying the old inline-`main`/`lang.execution/1` tail: whole-program `main` owns the document. [VERIFIED: internal/compiler/cgen/cgen_program.go:652-790; internal/compiler/cgen/cgen.go:2002-2180]

### Pattern 3: Characterize before deleting

`TestN1ConvergenceDifferential` is an internal test expressly designed to compare legacy public dispatch and direct `emitProgram` over five shapes. Flip it from present divergence to scoped identity for the three admitted fixtures; retain explicit refusal expectations for foreign and any unadmitted pointer family. Its table must remain exactly five rows unless a deliberately reviewed scope change updates the characterization. [VERIFIED: internal/compiler/cgen/cgen_n1_convergence_test.go:13-116]

### Pattern 4: Machine-linked provenance, not a re-baseline

Place a typed/fixed four-entry ledger next to `previousPhaseGoldenCDigests`, and test both directions: every digest-map entry must have exactly one ledger entry and every ledger entry must name a digest-map path with matching old/new checksum fields. The existing map has four exact paths and `TestPreviousPhaseGoldenCUnchanged` already detects missing, unpinned, and mismatched files. Quote from the source of truth: `"testdata/phase1/generated.golden.c"`, `"testdata/phase2/owned_transfer.golden.c"`, `"testdata/phase4/foreign_layout_mismatch.golden.c"`, and `"testdata/phase5/restrict_borrow.golden.c"`. [VERIFIED: internal/compiler/core/core_test.go:148-207]

### Pattern 5: Separate integrity from semantic evidence

A SHA-256 comparison establishes byte identity/integrity. It does not establish that a changed program is semantically correct; retain the four-tier execution comparator. Reproducible Builds likewise defines reproducibility through defined inputs/instructions/environment and bit-for-bit artifacts, with hashes used for bitwise verification. [CITED: https://reproducible-builds.org/docs/definition/]

### Anti-Patterns to Avoid

- **A retained `len(program.Functions) != 1` fork:** violates NAT-08 even if all current fixtures pass. [VERIFIED: internal/compiler/cgen/cgen.go:112-169]
- **Changing unsupported-shape error precedence:** moving preflight ahead of shape validation breaks D-15-18 ordering. [VERIFIED: internal/compiler/cgen/cgen_program.go:453-491]
- **Reusing the legacy `/1` output tail in `emitProgramFunction`:** it would write per-callee documents instead of one `/2` document from `main`. [VERIFIED: internal/compiler/cgen/cgen_program.go:652-659]
- **Hardcoding `live_resources` empty:** incompatible with D-16-05; derive the tail through a surviving-emitter resource abstraction.
- **Using a successful `restrict` microprobe as LTO proof:** C permits translators to ignore restrict alias implications, while LLVM documents LTO capability rather than a project-specific non-inertness witness. [CITED: https://www.open-std.org/jtc1/sc22/wg14/www/docs/n2347.pdf; CITED: https://llvm.org/docs/LinkTimeOptimization.html]
- **A third deferral disguised as an owner change:** D-10-60 requires an explicit requirements amendment or the item becomes never-cut. [VERIFIED: .planning/milestones/M002-phases/10-trusted-interprocedural-oracle/PHASE-10-DEBT.md:384-396]

## Don't Hand-Roll

| Problem | Do not build | Use instead | Why |
|---|---|---|---|
| Payload field layout | a second C-layout derivation | checker-derived `core.DataType` alternative details and existing branch lowering | A duplicate layout law could drift from semantic admission. [VERIFIED: internal/compiler/cgen/cgen.go:1780-2180] |
| Event equality | ad-hoc output-string assertions | `session.Phase5CompareProgramEngines` four-tier comparator | It compares more than emitted bytes and already routes the native tiers. [VERIFIED: internal/compiler/session/session_phase11_differential_test.go:128-145] |
| Debt-record parser | a Phase-16-only Markdown parser | extend/use `checkDebtRegister` well-formedness law | Existing parser centralizes artifact discovery, required columns, owner vocabulary, grade, and witness checks. [VERIFIED: internal/compiler/session/session_test.go:2555-2710] |
| Restrict generalization | generic alias attributes or a discharge protocol | a bounded exact-shape structural refusal plus M004 debt for everything else | The current emitted program is one TU and project evidence records LTO inertness; the probe does not justify future alias claims. [VERIFIED: internal/compiler/session/session_phase11_differential_test.go:74-91] |

## Restrict Probe Result and Boundary

The requested hand-written one-TU C17 probe was run on the available macOS host: source shape was `static T f(T *restrict p)`, one pointer, `T copy = *p`, caller-local reads before/after call, no mutation/callback/foreign call/escape/atomic/volatile/separate compilation. `clang` completed and executed successfully under `-std=c17 -O0`, `-O3`, `-O3 -flto`, and `-O3 -fsanitize=address,undefined`; all four returned `PASS`. [VERIFIED: local probe run 2026-09-19]

Linux was unavailable in this environment, so the cross-host portion remains an execution checkpoint, not a verified portability result. The planner must schedule the same source unchanged in Linux CI before admission. The normative source says restrict creates an association over each execution of the function block and that a translator may ignore aliasing implications; its undefined-behavior appendix identifies incompatible access to a modified object through a restrict-qualified pointer and another non-based pointer. This supports only the locked exact-shape probe, never an inference that every Lang borrow permits `restrict`. [CITED: https://www.open-std.org/jtc1/sc22/wg14/www/docs/n2347.pdf]

**Planning consequence:** Place a decision checkpoint after Linux proof and a source-structural refusal test. If the probe is accepted, port only `emitLinearBorrowedByPointer`, with tests rejecting each prohibited extension; do not port `emitLinearBorrowedByPointerPlain` or foreign lowering. If either proof fails, record all three in the M004 amendment/debt cut.

## Runtime State Inventory

This is a port/delete refactor, so all five categories were checked.

| Category | Items Found | Action Required |
|---|---|---|
| Stored data | None — source compiler has no emitter-routing datastore identified. | Code/doc change only. |
| Live service config | None — no deployed service configuration participates in `cgen` dispatch. | None. |
| OS-registered state | None — emitter functions are compiled from source; no OS registration embeds their names. | None. |
| Secrets/env vars | None — no phase path reads a secret or changes environment key names. | None. |
| Build artifacts | Temporary native binaries are regenerated by the runner; checked-in golden C and digest map are the durable artifacts. | Update goldens/digests only through the ledger gate. [VERIFIED: internal/compiler/core/core_test.go:162-207] |

## Common Pitfalls

### Pitfall 1: Phase 15 regression by an apparently local port

**What goes wrong:** a branch-support change serializes before shape/preflight checks, changes `/2` event ordering, or emits one output document per nested function.

**How to avoid:** test refusal precedence explicitly with malformed graph, unsupported foreign/pointer shapes, and preflight-overflow controls; retain `emitProgram` as sole writer of `main`/document tail. [VERIFIED: internal/compiler/cgen/cgen_program.go:412-615]

### Pitfall 2: Payload arm behavior copied without its defect contract

**What goes wrong:** payload structs/tags or the defect terminal are reimplemented differently and old C either changes silently or loses its `_Noreturn` exemption.

**How to avoid:** migrate the existing `emitBranchOperations` logic as a helper, preserve tagged struct construction and the `lang_defect` path, then pin exact C plus semantics. [VERIFIED: internal/compiler/cgen/cgen.go:2002-2180]

### Pitfall 3: Provenance ledger that can go stale

**What goes wrong:** ledger prose exists but no test binds it to the digest map; new or removed map paths escape review.

**How to avoid:** enforce map↔ledger bijection, 64-hex validation, old-digest equality against the pre-port value, and current/new-digest equality against disk. Exercise missing, duplicate, stale, and mismatched fixtures as negative controls.

### Pitfall 4: M004 rows that parse but do not honestly retire debt

**What goes wrong:** an old row is silently repointed or the `-flto` impact is omitted.

**How to avoid:** amend NAT-09 in the same plan series; create/extend a Phase-16 debt register with `items:` count, owner, grade, witness, prerequisite, reopening condition, and explicit one-TU/LTO consequence. Existing format validation alone does not track deferral hops. [VERIFIED: .planning/milestones/M002-phases/10-trusted-interprocedural-oracle/PHASE-10-DEBT.md:390-396]

## State of the Art

| Old approach | Current required approach | Impact |
|---|---|---|
| Function-count dispatch picks legacy N=1 emitters and only uses `emitProgram` for N≠1. | `emitProgram` admits every in-scope program; no function-count production fork remains. | Phase 17 modifies one type model once. [VERIFIED: internal/compiler/cgen/cgen.go:112-169] |
| N=1 differential records four refusals and one byte mismatch. | It remains and flips to identity only for the three admitted fixtures; cut families remain explicit refusals. | Characterization survives the authority cut. [VERIFIED: internal/compiler/cgen/cgen_n1_convergence_test.go:43-116] |
| Digest map proves prior goldens stayed unchanged. | Map plus a four-entry ledger proves each approved port-related byte change has reviewable responsibility and witness. | No silent rebaseline. [VERIFIED: internal/compiler/core/core_test.go:148-207] |

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|---|---|---|
| A1 | The current macOS probe shape will compile and behave identically in Linux CI. | Restrict Probe Result | By-pointer admission cannot proceed; choose M004 cut at checkpoint. |
| A2 | The port can derive live resources with a small reusable abstraction without importing foreign-resource machinery. | Architecture Patterns | May require more factoring; do not weaken D-16-05. |

## Open Questions

1. **Does Linux validate the exact one-pointer read/copy-only restrict probe across the required lanes?**
   - What we know: macOS lanes passed.
   - What's unclear: Linux Clang/linker/sanitizer result.
   - Recommendation: run unchanged probe in Linux CI before the D-16-07 admission checkpoint.
2. **Can `emitLinearBorrowedByPointer` be structurally fenced without importing unbuilt discharge-pair machinery?**
   - What we know: the port is allowed only for the exact source shape.
   - What's unclear: the smallest inspectable predicate/test seam.
   - Recommendation: prototype a private predicate and exhaustive rejection matrix while legacy code remains present; reject → cut if the fence is not auditable.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|---|---|---|---|---|
| Go | implementation/tests | ✓ | `go1.24` module target | — |
| Clang | C17 probe/native lanes | ✓ | host-installed | — |
| macOS | local restrict lanes | ✓ | current host | — |
| Linux | required cross-host restrict proof | ✗ | — | CI checkpoint; no local substitute |

**Missing dependencies with no fallback:** Linux execution evidence for D-16-06.

## Validation Architecture

### Test Framework

| Property | Value |
|---|---|
| Framework | Go standard `testing` |
| Config file | `go.mod` |
| Quick run command | `env GOCACHE=/tmp/ai-lang-phase16-cache sh scripts/assert-go-tests.sh ./internal/compiler/cgen/... 'TestN1ConvergenceDifferential|Test.*Program'` |
| Full suite command | `env GOCACHE=/tmp/ai-lang-phase16-cache go test ./...` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|---|---|---|---|---|
| NAT-08 | N=1 legacy/program identity in both modes for three scoped fixtures | unit + golden | focused cgen target via `assert-go-tests.sh` | ✅ characterization; ❌ mode/identity expansion |
| NAT-08 | no legacy function-count dispatch or three deleted function declarations | structural | focused cgen/core test + grep as secondary check | ❌ Wave 0 |
| NAT-08 | Phase 15 order remains graph/entry → shape → preflight → serialization | unit/mutation | cgen package focused target | ✅ ordering seams; ❌ port-specific cases |
| NAT-08 | semantic agreement at interpreter/O0/O3/O3-LTO | integration | `go test ./internal/compiler/session/... -run TestPhase11InterproceduralDifferential -count=1` | ✅ |
| NAT-08 | golden ledger bijects with digest map and current C | unit | `go test ./internal/compiler/core/... -run 'TestPreviousPhaseGoldenCUnchanged|Test.*Golden.*Ledger' -count=1` | ✅ map; ❌ ledger |
| NAT-09 | amendment + M004 debt records are complete and well-formed | document/unit | `go test ./internal/compiler/session/... -run TestDebtRegistersAreWellFormed -count=1` | ✅ parser; ❌ Phase-16 rows |

### Sampling Rate

- **Per task commit:** corresponding `scripts/assert-go-tests.sh` focused command.
- **Per wave merge:** cgen/core/session focused suites.
- **Phase gate:** full `go test ./...` green; run macOS and Linux exact restrict lanes before deciding by-pointer admission.

### Wave 0 Gaps

- [ ] Add explicit both-mode N=1 byte-identity rows and retained scoped-refusal rows.
- [ ] Add golden-change ledger parser/bijection/current-digest negative controls.
- [ ] Add structural test proving public dispatch contains no function-count route and only `emitProgram` production route.
- [ ] Add Phase-16 debt register and amendment assertions, including M004 owner/prerequisite/reopen/LTO text.
- [ ] Add exact-shape restrict probe test/harness and extension-refusal matrix if D-16-07 is considered for admission.

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---|---|---|
| V2 Authentication | no | No user authentication surface. |
| V3 Session Management | no | No session surface. |
| V4 Access Control | yes | Keep emitter authority explicit: one checked admission route; do not retain bypass routes. |
| V5 Input Validation | yes | `corevalidate`, graph/entry/shape validation before generated-C serialization. [VERIFIED: internal/compiler/cgen/cgen.go:112-169; internal/compiler/cgen/cgen_program.go:412-491] |
| V6 Cryptography | limited | Use Go `crypto/sha256` map-based integrity check already present; do not treat it as semantic proof. [VERIFIED: internal/compiler/core/core_test.go:148-207] |

### Known Threat Patterns

| Pattern | STRIDE | Standard Mitigation |
|---|---|---|
| Untrusted/invalid core input reaches codegen | Tampering | retain validation ordering and named refusals before serialization. |
| C alias promise overclaims actual Lang proof | Tampering | exact restrict source-shape fence, dual-host lanes, no generalization, M004 cut on failure. |
| Golden digest masks a semantic regression | Tampering | four-tier semantic comparator and ledger witnesses in addition to digest. |
| Stale debt amendment conceals deferral | Repudiation | D-10-60 amendment plus well-formed debt row and focused document test. |

## Sources

### Primary (HIGH confidence)

- `internal/compiler/cgen/cgen.go`, `cgen_program.go`, `cgen_n1_convergence_test.go`, `cgen_program_test.go`, `cgen_test.go` — actual dispatch, lowerings, and tests.
- `internal/compiler/core/core_test.go` — four pinned digest map and existing bidirectional file set check.
- `internal/compiler/session/session_phase11_differential_test.go` and `session_test.go` — four-tier comparator lane and debt-register law.
- [C17 restrict formal definition](https://www.open-std.org/jtc1/sc22/wg14/www/docs/n2347.pdf) — bounded alias semantics.
- [LLVM LTO design](https://llvm.org/docs/LinkTimeOptimization.html) — LTO capability boundary.
- [Reproducible Builds definition](https://reproducible-builds.org/docs/definition/) — hash/integrity versus defined provenance inputs.

## Metadata

**Confidence breakdown:**

- Standard stack: HIGH — module and host constraints directly inspected; no package recommendation.
- Architecture: HIGH — dispatch and surviving emitter directly inspected.
- Pitfalls: HIGH — derived from locked decisions, existing characterization, and Phase 15 ordering code.

**Research date:** 2026-09-19
**Valid until:** implementation start; re-run the host probe immediately before the decision checkpoint if compiler/toolchain changes.
