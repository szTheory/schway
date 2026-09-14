---
phase: "11"
slug: "multi-function-native-emission-and-interprocedural-equivalen"
status: verified
# threats_open = count of OPEN threats at or above workflow.security_block_on severity (the blocking gate)
threats_open: 0
asvs_level: 1
created: "2026-09-12"
---

# Phase 11 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

Register origin: **authored at plan time** — all nine `11-0N-PLAN.md` files
carry a parseable `<threat_model>` block, so this is verification of a
pre-declared register, not a retroactive STRIDE reconstruction. ASVS level 1
with `block_on: high`; the L1 short-circuit applies (`threats_open: 0` and
`register_authored_at_plan_time: true`).

Every `mitigate` row below was confirmed by locating its named artifact in the
tree and RUNNING it, not by reading its mitigation prose. The test-name column
in "Verification Evidence" is the audit — if a named test is deleted or
renamed, this register no longer verifies.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| `.lang` source → compiler | untrusted program text crosses into `check`/`corevalidate` | attacker-authored source |
| `.lang` source → generated C identifiers | attacker-influenced function/place names reach emitted C, where collision with a `LANG_`-prefixed reserved identifier is a namespace-confusion surface | identifier text |
| `core.Program` → `cgen` | emission must never bypass `corevalidate.Validate` on a multi-function fast path | typed core artifact |
| resolved entry → oracle vs. binary | if interpreter and binary resolve different entry functions, every equivalence claim compares two different programs | entry-function identity |
| emitted artifact → optimizer | any attribute in emitted C is an unverifiable promise handed to Clang; this phase's claim is that the set is empty | C attributes / qualifiers |
| hand-written C → Clang toolchain | the NAT-07 control deliberately compiles C containing a FALSE `restrict` promise | deliberate false claim |
| toolchain identity → evidence claim | the LTO divergence is a property of a specific LLVM version; an unpinned version silently invalidates the evidence | `clang --version` string |
| build cache → artifact identity | a cache hit substitutes a prior binary for a fresh compile; an undeclared input makes the substitution unsound | compiled artifact |
| `cache` package → rest of compiler | `cache` is a dependency-free leaf; importing `core`/`originvalidate` would let a cached verdict leak into a trust-crossing derivation | package dependency edge |
| reducer output → bug report | a reduced program is presented as reproducing the original defect; slippage reports a different bug as the original | defect claim |
| `reduce` → `callgraph` | an import would widen a package whose consumer set is a documented, load-bearing independence boundary | package dependency edge |
| committed register → audit verdict | the QLT-03 register is trusted data; an audit that cannot fail turns it into decoration | register JSON |
| search-time verdict → re-verification verdict | reusing a search-time artifact makes re-verification a restatement rather than an independent check | reduction signature |

---

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-11-01 | Tampering | `cache.Consult` / `phase6ArtifactSpec` | high | mitigate | Eighth declared cache input (`cgen_source`) closes the stale-artifact hole Q-02 probed; seven-name list pinned so widening is a visible diff | closed |
| T-11-02 | Spoofing | spike verdict provenance | medium | mitigate | Every verdict is a committed single-branch test; a changed verdict turns a test red rather than silently rewriting a document | closed |
| T-11-03 | Repudiation | NAT-07 evidence record | high | mitigate | D-11-22 toolchain pin: verbatim `clang --version` captured in `11-NAT07-EVIDENCE.md` alongside the matrix | closed |
| T-11-04 | Tampering | injected false `restrict` | medium | mitigate | Injection confined to hand-written test C through `native.Runner`; never enters `cgen`'s emission path or any shipped artifact (D-11-24) | closed |
| T-11-05 | Information disclosure | NAT-07 control output | low | accept | Control emits no secrets; output is a deterministic integer the matrix compares | closed |
| T-11-06 | Tampering | `cNames` across N functions | high | mitigate | D-11-08 two-tier allocation preserves honest-reservation and prefix-confinement | closed |
| T-11-07 | Spoofing | entry-function resolution | high | mitigate | `callgraph.EntryFunction` is the single resolver for both engines; zero-or-many roots is a named fail-closed refusal (`core.entry_ambiguous`), never a guess | closed |
| T-11-08 | Elevation of privilege | `emitProgram` bypassing validation | high | mitigate | `emitProgram` reached only from `Emit`/`EmitNative`, strictly after the unconditional `corevalidate.Validate` call | closed |
| T-11-09 | Repudiation | mid-phase gate verdict | high | mitigate | D-11-14 conjunction + D-11-18 mutation kill: gate must fail under the justified profile and at N=0, so a green verdict cannot come from an empty corpus | closed |
| T-11-10 | Tampering | the `restrict` write site | high | mitigate | D-11-16 token-local suppression: one `Fprintf` call site, one qualifier helper | closed |
| T-11-11 | Information disclosure | generated provenance comments | low | accept | Comment discloses only decision identifiers already public in the planning record | closed |
| T-11-12 | Spoofing | widened verification lanes | high | mitigate | Every widened lane resolves entry through the single `callgraph.EntryFunction`; comparator's five axes asserted intact, not narrowed for N>1 | closed |
| T-11-13 | Repudiation | lane deferral on stale declared inputs | high | mitigate | D-11-42: no new Phase 11 lane added to `SelectLanesForFixture`'s change-state; decision recorded in the ledger | closed |
| T-11-14 | Tampering | undecided single-function guards | medium | mitigate | `11-GUARD-LEDGER.md` forces a WIDENED/KEPT disposition on all 31 re-verified baseline sites | closed |
| T-11-15 | Repudiation | `AuditQLT03Register` | high | mitigate | D-11-50 mutation kill: an `export_test.go` seam makes the audit fail without editing the committed register | closed |
| T-11-16 | Tampering | QLT-03 register data file | medium | mitigate | Audit is read-only; committed JSON asserted unchanged after a test run, and the file is `//go:embed`-bound to the binary | closed |
| T-11-17 | Information disclosure | disclosed trust gaps in the register | low | accept | The register deliberately publishes known gaps (the `peerDeriveOriginFacts` `OpCall` case) — the intended SPARK justified-unproved-check pattern | closed |
| T-11-18 | Tampering | `cache.Consult` serving a stale `cgen` binary | high | mitigate | D-11-41 eighth input hashes `internal/compiler/cgen/*.go`, so an edited emitter moves the key; proven by a key-moves test | closed |
| T-11-19 | Elevation of privilege | `cache` importing a trust-crossing package | high | mitigate | QLT-06b dual scans (direct `go/parser` + transitive `go list -deps`), each with its own negative control | closed |
| T-11-20 | Tampering | closure-keyed cache admitting hits on less evidence | high | mitigate | D-11-38: `ClosureDigest` never wired into `cache.Input`; whole-program `FixtureSource` hash strictly dominates | closed |
| T-11-21 | Denial of service | unbounded `go list -deps` output | low | mitigate | Transitive scan reuses the repo's bounded writer (`1 << 20` bytes) and 2-minute timeout convention | closed |
| T-11-22 | Tampering | reduction slippage | high | mitigate | No move renumbers; byte-identical operation IDs asserted, which is what makes 11-09's strict-field-equality re-verification non-vacuous | closed |
| T-11-23 | Denial of service | unbounded reduction search | medium | mitigate | Derived budget `AttemptsPerFunction*len(Functions) + callSiteCount`, fail-closed with a proven floor | closed |
| T-11-24 | Elevation of privilege | `reduce` importing `callgraph` | medium | mitigate | Boundary enforced mechanically with a negative control, not left to the package doc | closed |
| T-11-25 | Spoofing | reduction slippage via a nil foreign-call sequence | high | mitigate | D-11-33: static reverse-postorder walk compared against a dynamic derivation from the `-O0` event stream, with a seeded-disagreement test proving the comparison is load-bearing | closed |
| T-11-26 | Repudiation | re-verification inheriting the search's relaxation | high | mitigate | D-11-32: strict field equality, no `CausalRole` fallback, from a cold start | closed |
| T-11-27 | Repudiation | a vacuous QLT-05 gate | high | mitigate | D-11-34: engineered fixture with a provably removable function, plus the complementary test that a zero-move reduction FAILS the assertion | closed |
| T-11-SC | Tampering | Go module dependencies | low | accept | Phase adds zero external packages — `go.mod` carries no `require` block at all and is unchanged since `032b837 feat(01-01)`. No package-legitimacy checkpoint applies | closed |

*Status: open · closed · open — below high threshold (non-blocking)*
*Severity: critical > high > medium > low — only open threats at or above workflow.security_block_on count toward threats_open*
*Disposition: mitigate (implementation required) · accept (documented risk) · transfer (third-party)*

---

## Verification Evidence

Each `mitigate` row's named control was located and executed on 2026-09-12.
All green.

| Threat(s) | Control located | Result |
|---|---|---|
| T-11-06 | `cgen_names_test.go:341` `TestMultiFunctionNameAllocation` | PASS |
| T-11-07, T-11-12 | `callgraph_entry_test.go:254` `TestSessionRunSitesDoNotIndexFunctionsZero` | PASS |
| T-11-08 | `cgen.go:30,36` / `:60,66` — `emitProgram` call sites both strictly follow `corevalidate.Validate` | Confirmed by reading the dispatch |
| T-11-09, T-11-10 | `session_phase11_gate_test.go:199` `TestPhase11SuppressionIsDiffLocal`, `TestPhase11ZeroAttributeGate` | PASS |
| T-11-03 | `11-NAT07-EVIDENCE.md:20` pins `Apple clang version 21.0.0 (clang-2100.1.1.101)`; `native_lto_test.go:343` `TestCompositionOnlyLTODivergence` | PASS, pin matches host |
| T-11-14 | `11-GUARD-LEDGER.md` — 31 WIDENED/KEPT dispositions | Present |
| T-11-15, T-11-16 | `qlt03_shape_register.go:55` `//go:embed`, `qlt03_shape_register_test.go:162` unchanged-after-run assertion, `TestQLT03*` suite | PASS |
| T-11-01, T-11-18 | `cache/probe.go:28-52` — eighth `cgen_source` input documented as CLOSING the `TestQ02StaleCgenServesReusedArtifact` hole | PASS |
| T-11-19, T-11-21 | `cache/probe_test.go:266-317` — dual import scans with negative controls; `cacheMaxGoListDepsOutputBytes = 1 << 20` | PASS |
| T-11-20 | `cache/probe_test.go:472` `TestNoClosureDigestInCache` | PASS |
| T-11-22 | `reduce_multifunction_test.go:465` `TestOperationIDsUnchangedByWholeProgramMoves` | PASS |
| T-11-24 | `reduce_multifunction_test.go:374` `TestReduceDoesNotImportCallgraph` | PASS |
| T-11-25, T-11-26, T-11-27 | `session_qlt05_reverify_test.go:175,346,367` — `TestQLT05ReverificationRejectsCausalRoleFallback`, `TestQLT05GateIsNonVacuous`, `TestQLT05EmptyReductionFailsAntiVacuity` | PASS |
| T-11-SC | `go.mod` — no `require` block; unchanged since `032b837` | Confirmed |

---

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|
| AR-11-01 | T-11-05 | The NAT-07 control emits no secret material; its output is a deterministic integer the divergence matrix compares | plan-time disposition (11-02-PLAN.md) | 2026-09-12 |
| AR-11-02 | T-11-11 | Generated provenance comments disclose only decision identifiers already public in the planning record | plan-time disposition (11-04-PLAN.md) | 2026-09-12 |
| AR-11-03 | T-11-17 | The QLT-03 register deliberately publishes its known trust gaps — the intended SPARK justified-unproved-check pattern, not an oversight | plan-time disposition (11-06-PLAN.md) | 2026-09-12 |
| AR-11-04 | T-11-SC | Phase adds zero external packages; `go.mod` has no `require` block and is unchanged since Phase 1, so no package-legitimacy checkpoint applies | plan-time disposition (all 9 plans) | 2026-09-12 |

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-09-12 | 28 | 28 | 0 | /gsd-secure-phase 11 (orchestrator, ASVS L1 short-circuit) |

### Note — WR-01 closed during this session

`11-REVIEW.md` WR-01 (`reduce.Reduce` trusting an unvalidated
`Seed.EntryFunctionID`) was adjacent to T-11-22/T-11-24's reduction-integrity
boundary but was not itself a register row. It was closed in this same session
by `reduce.Seed.Validate` (`internal/compiler/reduce/seed_validate.go`), which
refuses an empty or unmatched entry ID with `core.seed_entry_invalid` rather
than letting `dropOrphanFunction` silently delete the program's real entry
point. That strengthens the reducer-output→bug-report trust boundary above: a
seed that cannot say which function IS the program is now refused instead of
reduced. See `11-UAT.md`.

---

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-09-12
