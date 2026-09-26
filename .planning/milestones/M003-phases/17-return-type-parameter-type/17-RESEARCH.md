# Phase 17: Return Type ≠ Parameter Type - Research

**Researched:** 2026-09-22
**Domain:** Go-hosted compiler semantic widening: independent parameter/return type facts, call contracts, C17 lowering, and protocol-only repair
**Confidence:** HIGH

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-17-01:** Use one checked-in, multi-function nominal-data Resource -> Result program as the canonical TYP-01 tracer. classify receives Resource and returns Result; main calls it and returns Result. This must check, interpret, lower, and agree across interpreter, -O0, -O3, and -O3 -flto. It exercises distinct AST/core type facts, call target resolution, generated-C prototypes/definitions/call sites, and entry output without duplicating the entire proof in the reverse direction.
- **D-17-02:** Do not represent this nominal Resource fixture as proof of tracked-resource ownership. A source-reachable Buffer <-> Result proof is not constructible on the current surface; claiming it would smuggle future payload/control-flow capability into this phase. A reverse-direction nominal regression is discretionary only if it catches a demonstrated directional defect.
- **D-17-03:** Create two minimal, parser-valid .lang frontier fixtures, each with one intended call-contract failure. The first isolates a caller/callee parameter mismatch and must move to check.call_argument_type_mismatch; the second has a matching argument but omits the callee return type from the caller's available facts and must move to check.call_return_type_unrepresentable. Pin the pre-widening refusal and assert that it moves; seam-level tests remain supplemental only.
- **D-17-04:** Make both diagnostics explain the contract distinction at the source call site using stable structured causes: callee, actual argument type, and declared parameter type for TYP-02; callee, declared return type, and the caller's available type facts for TYP-03. Do not let an incidental use-after-move or an earlier declaration error become the asserted proof.
- **D-17-05:** Derive parameter Drops and return Fresh locally and independently in check, corevalidate, and originvalidate, each from that layer's own inputs and type lookup. Sharing primitive ability rules is acceptable; sharing a return-contract helper, derived return facts, or a producer artifact is not. — **Reversibility:** costly — collapsing the derivations would invalidate TYP-04's independent-peer claim and require new evidence rather than a local refactor.
- **D-17-06:** Require return-only seeded mutations at each layer. A mutation must alter only the return-side ability/type lookup, leave parameter facts intact, and produce independently observable failure or divergence in that peer; a coordinated mutation must make the agreement gate fail. Extend import-boundary guards to cover any new production derivation files.
- **D-17-07:** Preserve Phase 13's sealed held-out call-argument fixtures as historical no-op/unrepairable controls. Create a new Phase 17 derivation/held-out pair before repair emission: its held-out program has a real, uniquely in-scope parameter-typed alternative enabled by the widened return contract, structurally differs from derivation (at least the existing main -> relay -> dispatch depth pattern), and is sealed before the repair is implemented. — **Reversibility:** costly — changing a sealed fixture afterward would invalidate the anti-overfitting claim and require a new corpus and provenance record.
- **D-17-08:** The held-out proof must restore exact original source bytes, finish repaired, use exactly two real lang --json check subprocesses plus an independent clean re-check, and retain protocol-only checks for diagnosis, repair kind, span, replacement, and prose-scramble invariance. Add red controls for stripped repair fields/kind/span/replacement and a production-source scan forbidding held-out path references.
- **D-17-09:** Ratify D-13-02b as permanent until M006. Replace the obsolete sameType-based witness and P17 landing with an executable witness that proves no production resolveBlame path exists and all currently admitted signature fields are producer-verifiable. Its sole reopening condition is a user-declared contract field that the declaring function's own admission cannot verify (separate compilation); do not restate automatic DX-06 closure.
- **D-17-10:** S-009 has no recorded result. Treat it as a required planning input that calibrates appetite for later repair-surface expansion, not as a reason to weaken TYP-05 or as a blocker of the semantic type-widening work.

### the agent's Discretion

Exact nominal type names, fixture module names, helper names, test factoring, and diagnostic prose are at the planner's discretion, provided the source reachability, independent derivation, sealed-corpus, and structured-protocol properties above remain true.

### Deferred Ideas (OUT OF SCOPE)

- A genuine Buffer <-> Result tracked-resource source proof — requires a future constructible surface; do not simulate it with a seam and call it Phase 17 evidence.
- Reverse-direction nominal two-type regression — add only if a real directional bug warrants its cost.
- DX-06/B1 blame wiring — M006, when separate compilation introduces a user-declared contract field the declaring function cannot verify.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|---|---|---|
| TYP-01 | Return type may differ from parameter type; a relying program checks, runs, lowers. | Widen facts and sole emitter; prove four-tier source tracer. |
| TYP-02 | Bad call argument type is named from .lang source. | Preserve argument gate and prove source diagnostic/cause ordering. |
| TYP-03 | Unrepresentable return contract is named from .lang source. | Resolve target from return fact, fail closed if absent. |
| TYP-04 | Drops/Fresh derive from return abilities independently. | Split lookup in three peers and mutation-kill it. |
| TYP-05 | Repair reaches repaired on sealed held-out use_matching_argument. | Seal distinct corpus before repair emission; real subprocess proof. |
</phase_requirements>

## Summary

Phase 17 is a narrow but cross-cutting semantic widening. The IR stores a function parameter and return type separately, but admission, fact minting, call-target resolution, peers, and generated C still collapse them into the parameter-side fact. Sequence work fixture-first: pin old refusals, introduce distinct facts and independent derivations, then make the source tracer executable through the one surviving emitter. [VERIFIED: internal/compiler/check/check.go:255-257,3146-3150,3397-3407; internal/compiler/core/core.go:128-135]

The correctness boundary is directional: call source retains caller argument type; call target resolves from callee return type. The checker currently has one caller-local fact and refuses a callee return constructor not equal to it. Core validation already independently compares source-to-parameter and target-to-return. [VERIFIED: internal/compiler/check/check.go:3545-3576; internal/compiler/corevalidate/corevalidate.go:2838-2855]

**Primary recommendation:** Fixture-first diagnostics; two facts plus independent Drops/Fresh; one emitProgram parameter/return C-type port; sealed protocol-only repair proof and permanent D-13-02b witness.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|---|---|---|---|
| Source signature/call admission | Checker | Core artifact | Checker translates source contracts into core operations. |
| Return ownership facts | Three admission peers | Interface summary | Each peer derives parameter and return facts before agreeing. |
| Call target validation | corevalidate | originvalidate | Core peer checks target-to-return; origin peer publishes contract. |
| Native execution | Generated C artifact | Native runner | emitProgram emits prototypes, definitions, calls, entry output. |
| Automated repair | External repair driver | Source bytes | Driver uses structured JSON and declared span. |

## Project Constraints (from AGENTS.md)

- Go 1.24 standard library hosts Stage 0; readable C17 emits to installed Clang. [VERIFIED: AGENTS.md:25-26]
- Prefer standard-library-only, shallow audited boundaries; add no external production dependency. [VERIFIED: AGENTS.md:27]
- Preserve deterministic fixtures, negative controls, mutation, differential execution, and sanitizer cost lanes. [VERIFIED: AGENTS.md:29]
- Preserve macOS/Linux portability; do not encode Apple arm64 host facts. [VERIFIED: AGENTS.md:30]
- Keep unsafe operations, FFI, secrets, and build authority explicit. [VERIFIED: AGENTS.md:31]
- This is a GSD planning artifact, not authorization for implementation changes. [VERIFIED: AGENTS.md:49-56]

## Standard Stack

### Core

| Tool | Version | Purpose | Why Standard |
|---|---:|---|---|
| Go standard library | go 1.24 | Checker, peers, reducer, tests, repair driver | Module declares go 1.24 and no external module. [VERIFIED: go.mod:1-3 — “module github.com/codename-lang/lang” / “go 1.24”] |
| Clang | Apple Clang 21.0.0 | Generated C17 native tiers | Installed native path. [VERIFIED: environment probe 2026-09-22] |

### Supporting

| Tool | Purpose | When |
|---|---|---|
| go test | Focused and full regression | Per task, merge, and gate. [VERIFIED: environment probe 2026-09-22] |
| emitProgram | Sole whole-program C law | Extend once; never revive cut emitters. [VERIFIED: internal/compiler/cgen/cgen_program.go:472-475,639-652] |

**Installation:** None. No new external package or tool is required. [VERIFIED: go.mod:1-3]

## Architecture Patterns

### System Architecture Diagram

    .lang frontier/tracer
             |
             v
    syntax + check
      parameter fact -> argument gate -> call source
      return fact ----> return gate ---> call target
             |
             v
    core.Program -> corevalidate + originvalidate -> agreement/mutation gates
             |
             +-> emitProgram -> C -> Clang tiers
             +-> interpreter comparison

    sealed source -> JSON check -> repair splice -> recheck -> independent clean check

### Recommended Project Structure

    internal/compiler/check/            checker facts and source diagnostics
    internal/compiler/corevalidate/     core re-derivation and contract peer
    internal/compiler/originvalidate/   independently built interface
    internal/compiler/cgen/             sole native emitter
    internal/compiler/reduce/           source projection preserving distinct types
    internal/compiler/session/          fixtures, gates, seals, differential evidence
    testdata/phase17/                   frontier/tracer/derivation/held-out sources
    cmd/lang-repair/                    protocol-only driver tests

### Pattern 1: Two facts, directional call contract

**What:** Mint parameter and return type facts per function. Resolve call source from the former and target from callee return type matched against a caller-visible fact.

**When:** Every linear, fallible-linear, and match/branch path assuming one functionID:type:0 fact.

**Evidence:** Core peer already compares source constructor to callee parameter and target constructor to callee return. [VERIFIED: internal/compiler/corevalidate/corevalidate.go:2838-2855]

### Pattern 2: Independent fact derivation

**What:** Each of check, corevalidate, and originvalidate looks up parameter abilities for Drops and return abilities for Fresh from its own inputs. Do not pass a precomputed return fact across packages.

**Evidence:** Both peers currently use parameter-side hasDropAbility for both fields. [VERIFIED: internal/compiler/corevalidate/corevalidate.go:2327-2360 — “Drops: hasDropAbility && !peerParameterEscapesOwned(function)” / “returnContract.Fresh = returnContract.Mode == "owned" && hasDropAbility”; internal/compiler/originvalidate/originvalidate.go:913-950 — “Drops: hasDropAbility && !parameterEscapesOwned(function)” / “returnContract.Fresh = returnContract.Mode == "owned" && hasDropAbility”]

### Pattern 3: C type pair in sole emitter

**What:** Carry parameter and return C types through prototypes, definitions, calls, entry input, and entry output.

**Evidence:** Current typeNames is used in all these positions. [VERIFIED: internal/compiler/cgen/cgen_program.go:554-571,639-652,725-740,1136-1160 — “typeNames := make([]string, len(functions))” / “static %s %s(%s, unsigned int);” / “%s lang_entry_output = %s(lang_entry_input, 0u);” / “static %s %s(%s %s, unsigned int invocation_index) {”]

### Pattern 4: Seal before repair emission

**What:** Add derivation/held-out fixtures, prove structural distinction, write SHA-256 manifest, then emit repair.

**Evidence:** Existing seal hashes every listed fixture; repair helper uses temporary copy, exact repaired status, two driver subprocesses, independent clean check, and original byte stability. [VERIFIED: internal/compiler/session/session_phase13_injectors_test.go:451-500; cmd/lang-repair/repair_test.go:570-617]

### Anti-Patterns to Avoid

- One fact for both directions. [VERIFIED: internal/compiler/check/check.go:3545-3571]
- Shared peer helper/checker-produced return summary. [VERIFIED: 17-CONTEXT.md:48-56]
- Seam-only proof; TYP-02/TYP-03 require source fixtures and movement. [VERIFIED: 17-CONTEXT.md:29-40]
- Old no-op repair claimed as success. [VERIFIED: 17-CONTEXT.md:57-63]
- Reopening DX-06/B1 before M006. [VERIFIED: 17-CONTEXT.md:72-77]

## Don't Hand-Roll

| Problem | Don’t Build | Use Instead | Why |
|---|---|---|---|
| C compiler behavior | Mock C checker | Installed Clang/four-tier comparator | Only real generated C validates ABI. [VERIFIED: 17-CONTEXT.md:9-16] |
| Repair parsing | Diagnostic prose matching | Structured JSON protocol/driver | Driver consumes structured fields only. [VERIFIED: cmd/lang-repair/repair.go:2-7,409-445] |
| Held-out provenance | Fixture comments | Existing SHA-256 seal/guards | Makes alteration loud. [VERIFIED: internal/compiler/session/session_phase13_injectors_test.go:451-500] |
| Peer independence | Fourth peer/shared helper | Existing three peers plus seeds/import guards | Requirement is independent derivation. [VERIFIED: 17-CONTEXT.md:48-56] |

## Common Pitfalls

### Pitfall 1: Remove only sameType gates

Single type fact remains and target cannot represent return type. Current checker declares functionID:type:0 and resolves return contract against it. [VERIFIED: internal/compiler/check/check.go:3397-3407,3545-3576]

**Avoid:** Inventory all producers/consumers first; land refused frontier fixtures before production widening.

### Pitfall 2: Fresh from parameter ability

Both peers currently use parameter-side hasDropAbility for Drops and Fresh. [VERIFIED: internal/compiler/corevalidate/corevalidate.go:2327-2360; internal/compiler/originvalidate/originvalidate.go:913-950]

**Avoid:** Separate lookups in all layers; test return-only faults and coordinated disagreement.

### Pitfall 3: Partial C signature widening

One typeNames element supplies prototype, definition, and entry output positions. [VERIFIED: internal/compiler/cgen/cgen_program.go:554-571,639-652,725-740,1136-1160]

**Avoid:** Type pair throughout one emitter; structural C assertions plus tier execution.

### Pitfall 4: No-op repair

Former scope made the unique candidate the already-passed argument. [VERIFIED: 17-CONTEXT.md:57-69]

**Avoid:** Seal a structurally distinct fixture with real alternative before repair emission; assert exact protocol result and unchanged historical controls.

## Code Examples

### Contract-flow skeleton

    declare(parameterType, returnType)
      -> mint parameter fact from parameterType
      -> mint return fact from returnType

    call:
      source type equals callee parameter type
      target type equals callee return type
      otherwise emit named source diagnostic

Existing core peer establishes the two directional comparisons. [VERIFIED: internal/compiler/corevalidate/corevalidate.go:2838-2855]

### C lowering skeleton

    prototype: return-C-type function(parameter-C-type, invocation-index)
    definition: return-C-type function(parameter-C-type local, invocation-index)
    call: return-C-type target = callee(parameter-C-type argument, child-index)
    entry: input uses parameter C type; output uses return C type

Current emission is single-typed, so every position named above must be widened. [VERIFIED: internal/compiler/cgen/cgen_program.go:639-652,725-740,1136-1160]

## State of the Art

| Old Approach | Current Approach | When | Impact |
|---|---|---|---|
| One fact for parameter and return | Separate facts/directional calls | Phase 17 | Enables source-reachable return contracts. [ASSUMED] |
| Phase 13 unrepairable control | New sealed widened-contract proof | Phase 17 | Repair only with real alternative. [VERIFIED: 17-CONTEXT.md:57-69] |
| Three emitters | Sole emitProgram | Phase 16 | One lowering port. [VERIFIED: 17-CONTEXT.md:143-145] |

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|---|---|---|
| A1 | Two facts can retain current core schema without a version bump. | State of the Art | Add schema/version work before implementation. |

## Resolved Planning Status

1. **S-009 result remains unrecorded; its planning disposition is accepted and resolved.** S-009 is required planning input that calibrates later repair-surface appetite. It does not block semantic widening, weaken TYP-05, or authorize an additional repair class. This records the disposition, not a claim that S-009 is complete. [VERIFIED: 17-CONTEXT.md:94-98; 17-09-PLAN.md:23-26,89]

2. **Exact fact identifiers and minting locations are resolved by fixture-first inventory and producer-consumer work.** Plan 17-01 first pins parser-valid source fixtures and then introduces deterministic `functionID:type:N` facts, preserving the parameter fact while adding the return fact; Plan 17-02 preserves the split through explicit reducer fact selection. The implementation plans therefore own final identifier assignment and minting placement, with source movement and directional projection controls before lowering. [VERIFIED: 17-01-PLAN.md:64-66,87-88; 17-02-PLAN.md:62-65,105-106]

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|---|---|---:|---|---|
| Go | Compile/test | ✓ | go1.24.0 | — [VERIFIED: environment probe 2026-09-22] |
| Clang | Native tiers | ✓ | Apple Clang 21.0.0 | — [VERIFIED: environment probe 2026-09-22] |
| Git | Fixture/seal checks | ✓ | 2.41.0 | — [VERIFIED: environment probe 2026-09-22] |

**Missing dependencies with no fallback:** None.

## Validation Architecture

### Test Framework

| Property | Value |
|---|---|
| Framework | Go standard testing [VERIFIED: go.mod:1-3] |
| Config file | none — Go module root uses go.mod [VERIFIED: go.mod:1-3] |
| Quick run | go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/cgen ./internal/compiler/session ./cmd/lang-repair -count=1 |
| Full suite | go test ./... -count=1 |

### Phase Requirements → Test Map

| Req | Behavior | Test Type | Command | File Exists? |
|---|---|---|---|---|
| TYP-01 | Nominal tracer across interpreter/native tiers | integration/differential/C structure | focused session/cgen then full suite | ❌ Wave 0 |
| TYP-02 | Source frontier argument mismatch/cause | CLI/source | focused check/session | ❌ Wave 0 |
| TYP-03 | Source frontier return-unrepresentable/cause | CLI/source | focused check/session | ❌ Wave 0 |
| TYP-04 | Independent Drops/Fresh plus mutations | unit/peer/import | focused peer/session | ❌ Wave 0 |
| TYP-05 | Sealed injection exact repaired | end-to-end subprocess | focused repair/session | ❌ Wave 0 |

### Sampling Rate

- **Per task:** owning package focused test; quick command for shared seam.
- **Per merge:** go test ./... -count=1.
- **Phase gate:** full suite, four-tier tracer, diagnostic movement, mutation kills, sealed repair evidence.

### Wave 0 Gaps

- [ ] Refused-frontier tests with old diagnostic pins and movement assertions.
- [ ] Canonical tracer, four-tier comparison, generated-C structure assertions.
- [ ] Per-layer return-only/coordination mutations and import-boundary extension.
- [ ] Derivation/held-out fixtures, structural distinction, SHA-256 seal, protocol-field red controls.
- [ ] Executable D-13-02b witness: no production blame route; producer-verifiable fields.

## Security Domain

### Applicable ASVS Categories

| ASVS | Applies | Control |
|---|---|---|
| V2 Authentication | no | No authentication surface. |
| V3 Session Management | no | No session surface. |
| V4 Access Control | yes | Bounded local-source repair with explicit input/span. [VERIFIED: cmd/lang-repair/repair.go:409-445] |
| V5 Input Validation | yes | Fail-closed contracts, bounded JSON, parser-valid fixtures, span checks. [VERIFIED: internal/compiler/check/check.go:3545-3576; cmd/lang-repair/repair.go:409-445] |
| V6 Cryptography | yes | SHA-256 fixture integrity seal only; no hand-rolled crypto. [VERIFIED: internal/compiler/session/session_phase13_injectors_test.go:451-500] |

### Known Threat Patterns

| Pattern | STRIDE | Mitigation |
|---|---|---|
| Target type inherits argument type | Tampering | Separate facts and directional mutation kills. [VERIFIED: internal/compiler/check/check.go:3545-3576] |
| Held-out altered after evidence | Tampering | SHA-256 seal and mutation control. [VERIFIED: internal/compiler/session/session_phase13_injectors_test.go:451-500] |
| Prose controls repair | Tampering | Structured protocol only. [VERIFIED: cmd/lang-repair/repair.go:2-7,442-445] |
| Native ABI mismatch | Denial of Service | One type-pair port and Clang tier tests. [VERIFIED: internal/compiler/cgen/cgen_program.go:639-652,1136-1160] |

## Sources

### Primary (HIGH confidence)

- Phase 17 context/requirements — locked decisions and TYP-01 through TYP-05.
- check.go — sameType gates, fact use, source diagnostics.
- corevalidate.go/originvalidate.go — peer behavior.
- cgen_program.go — emitter type flow.
- repair.go/Phase 13 tests — protocol, subprocess, sealing precedent.

### Secondary (MEDIUM confidence)

- [Go specification — function signatures](https://go.dev/ref/spec) — parameter and result are distinct signature parts. [CITED: https://go.dev/ref/spec]
- [C17 draft](https://kcir.pwr.edu.pl/~mucha/PProg/c17_updated_proposed_fdis.pdf) — compatible function types require compatible return types. [CITED: https://kcir.pwr.edu.pl/~mucha/PProg/c17_updated_proposed_fdis.pdf]

## Metadata

**Confidence breakdown:**

- Standard stack: HIGH — module and tools checked locally.
- Architecture: HIGH — material seams opened in current tree.
- Pitfalls: HIGH — current explicit assumptions or locked evidence; A1 is assumed.

**Research date:** 2026-09-22
**Valid until:** 2026-10-22
