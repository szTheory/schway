---
phase: "02"
slug: "owned-values-and-abilities"
audited: "2026-09-03T23:51:20Z"
reaudited: "2026-09-04T02:20:00Z"
audited_head: "da75e95133fa9f90ebb74bef9cb382cf9e5baf3e"
code_head: "da75e95133fa9f90ebb74bef9cb382cf9e5baf3e"
asvs_level: 2
block_on: high
threats_total: 8
mitigated: 8
partial: 0
open: 0
threats_open: 0
status: SECURED
---

# Phase 2 — Security Audit

> Retroactive verification of the Phase 2 threat model (`02-VALIDATION.md` → Threat Model, repeated per-plan in `02-0*-PLAN.md` `<threat_model>` blocks) against implemented code at `HEAD` `da75e95`. Every mitigation below was located in source at a cited `file:line` and its falsifier was executed. Documentation and intent were not accepted as evidence.
>
> **Re-audit, 2026-09-04.** The first pass (at `f1206bb`) returned SECURED with T-02-03 PARTIAL and three non-blocking findings. This pass re-verifies the whole register against the fix wave `d9b370f..da75e95`. SEC-02-A and SEC-02-C are now **CLOSED** by executed load-bearing probes; SEC-02-B remains open as a falsifier-coverage gap; SEC-02-D is carried forward unfixed by developer decision; one new non-blocking finding (SEC-02-E) is raised. T-02-03 moves PARTIAL → **MITIGATED**. Newly in scope this pass: the `share`-ability borrow gate, transitive loan liveness on both admission layers, the ninth negative control, the `evidence.canonical_unstable` fail-closed guard, and the test-support spawn bounds.

## Audit Environment

| Property | Value |
|----------|-------|
| Cache | `GOCACHE=/tmp/ai-lang-resecurity` (created fresh for this re-audit) |
| Host | darwin/arm64, Go 1.24, Apple clang |
| Gate | `env GOCACHE=/tmp/ai-lang-resecurity sh scripts/verify-phase2.sh` → **exit 0** |
| Suite | `go vet ./...` clean; `go test ./...` all packages `ok` |

Commands actually run for this audit:

```sh
env GOCACHE=/tmp/ai-lang-resecurity sh scripts/verify-phase2.sh            # exit 0
# (the script itself runs go test ./..., go test -race ./..., go vet ./...,
#  builds ./cmd/lang, and emits both verify documents)
```

Four adversarial probes were run **out of tree**, against a `git archive HEAD`
copy in a scratch directory. No file in the repository was modified by this
audit. Each probe result is cited at the finding it settles.

Gate output confirmed all **nine** required negative-control IDs fired
(`control:ownership.use_after_move`, `control:ownership.move_while_borrowed`,
`control:ownership.transfer_requires_take`,
**`control:ownership.move_while_reborrowed`**, `control:ability.forged_copy`,
`control:core.duplicate_operation_id`, `control:interpreter-o0-o3-owned`,
`control:evidence.core_mismatch`, `control:backend.runtime_causality`),
with Phase 1 at **18** recomputed work units and Phase 2 at **52**, and
`"expected_escapes":["escape:coordinated-source-core-lie"]` reported in a
field structurally separate from every lane's `controls` array.

The ninth control is required fail-closed on three independent surfaces:
`session/session.go:600` declares the fixture/diagnostic pair, `:608-620` fails
`verify.fixture_missing` / `verify.fixture_input_limit` / `verify.control_missing`
if the fixture is absent, oversized, or does not produce exactly one
`ownership.move_while_borrowed` diagnostic, `:690-705` lists it in
`requiredControls` and rejects with `verify.control_missing` plus a
`verify.zero_work` check on every lane, and `testsupport/cli_test.go:256`
re-asserts all nine through the shipped CLI (`da75e95`). Lane work is derived
from the control table (`:602-605,621`), not from a hardcoded index, so adding a
control cannot silently leave the work count stale.

## Threat Verdicts

| ID | Threat | Severity | Disposition | Verdict | Mitigation (file:line) | Falsifier executed |
|----|--------|----------|-------------|---------|------------------------|--------------------|
| T-02-01 | Adversarial nested types / long bodies → unbounded work | high | mitigate | **MITIGATED** | `syntax/parser.go:12-19,37-38,252-253`; `syntax/lexer.go:39`; `corevalidate/corevalidate.go:14,317,339-359`; `session/session.go:394-401` | `TestTypeApplicationLimits`, `TestOwnershipRecovery`, `TestOwnershipRoundTrip`, `TestOwnershipWorkSeries`, `TestCoreValidationWorkSeries` — all PASS |
| T-02-02 | Forged positive ability reaches execution | high | mitigate | **MITIGATED** | `corevalidate/corevalidate.go:167` (independent re-derivation + `reflect.DeepEqual` on abilities *and* negative witnesses → `core.ability_mismatch`), `:281` `core.ability.share_denied` gate on `OpBorrowShared`, `:376-416` sealed per-ability recursive descent; production side `ability/ability.go:44-54,107-135` sealed structural combiner; new source-layer gate `check/check.go:282-292` `ownership.borrow_requires_share` | `TestBoxAllAbilityMasks`, `TestPairAllAbilityMasks`, `TestShareIsUniversallyGrantedAfterBufferShare`, `TestTypeRefDerivationUsesStructuralCombiner`, `TestNegativeWitnessPath`, `TestArbitraryMasksRemainTestPrivate`, `TestArbitraryMaskCannotEnterCoreValidation`, `TestSourceCannotGrantAbilityRoots`, `TestOwnershipMutationMatrix` — all PASS |
| T-02-03 | Ambiguous body representation / duplicate IDs (incl. generated-C identifier collapse) | high | mitigate | **MITIGATED** (was PARTIAL; SEC-02-A closed by `dd9c0a8`) | `core/core.go:102` closed union; `corevalidate/corevalidate.go:88-93,133-137,198,418-462`; `cgen/cgen.go:47-75` the two-property closure argument stated in code, `:79-81,87-103` honest `matchFixedNames`/`linearFixedNames`, `:340-368` single global allocator, `:369-405` `cName`/`cLocal` prefix-confinement invariants | `TestClosedBodyUnion`, `TestLinearIdentityStability`, `TestFeatureSpecificCoreExecutionSchemas`, `TestOwnedClaimReorderRejected`, `TestLinearCSerializesRuntimeState`, `TestNativeIdentifiersRemainCollisionFree`, `TestNativeToggleO0O3`, **`TestGeneratedIdentifierNamespacesStayConfined`** (3/3 subtests), **`TestReservedSetsCoverTheirOwnNamespace`**, **`TestIdentifierPrefixInvariance`** — all PASS; plus two auditor probes proving the new tests load-bearing |
| T-02-04 | Implicit noncopyable transfer | high | mitigate | **MITIGATED** | explicit `take` requirement enforced in `check/check.go` linear path; lowering preserves it at `cgen/cgen.go:171-196` | `TestTransferRequiresTake`, `TestImplicitByteCopy`, `TestOwnedTransferInterpreter` — PASS; `control:ownership.transfer_requires_take` fires in gate |
| T-02-05 | Corrupted move/loan core executes | high | mitigate | **MITIGATED** | `corevalidate/corevalidate.go:228-310` independent replay (`initialized`/`produced`/`ownerBlockedUntil` state machine), `:231-249` **transitive** `loansForPlace` loan-liveness closure, move-while-borrowed and use-after-move gates, separate from checker and interpreter; production side `check/check.go:359-398` `discoverLoanLastUses` with inherited loan association | `TestUseAfterMoveDiagnostic`, `TestMoveWhileBorrowedDiagnostic`, `TestOwnershipSequenceExhaustive`, `TestOwnershipOracleTracksLoansPerOwner`, **`TestTransitiveLoanBlocksMove`**, `TestUnvalidatedCoreCannotExecute`, `TestOwnershipMutationMatrix`, `FuzzOwnershipLinear` — all PASS; `control:ownership.move_while_reborrowed` fires in the gate |
| T-02-06 | Compiler/native child hangs, floods, or malforms output | high | mitigate | **MITIGATED** (SEC-02-C closed; one falsifier-coverage gap remains: SEC-02-B) | `native/native.go:68-69,77-79,97,106-108` deadlines; `:72-74,99-101` four independent bounded writers; `:137-155` `MaxStreamBytes+1` detection byte; `:168-192` strict one-document decoder; `evidence/evidence.go:121-159` bounded probes | `TestNativeStreamsIndependentlyBounded` (4/4 subtests), **`TestSourceNeverSpawnsUnboundedProcesses`** (repo-wide, empty allowlist), `TestExecutionDecoderRejectsMalformedOutput` (14/14 subtests), `TestNativeToolFailureIsOperational`, `TestDefaultFactsBoundsBothToolProbes` (6/6 subtests) — all PASS; plus an auditor-authored out-of-tree deadline probe re-run at `da75e95` (below) |
| T-02-07 | Backend re-infers source intent → optimizer-dependent drift | high | mitigate | **MITIGATED** | `cgen/cgen.go:16-29,33-46` validate-then-lower; `:171-213` C emitted strictly per admitted `LinearOperation`; runtime-derived events via `lang_record_event`/`lang_write_events` (`:218-276`); `session/session.go:74-82` exact-one real C mutation | `TestOwnedBackendMutationIsMismatch`, `TestOwnedEventReorderIsMismatch`, `TestOwnedExecutionFieldMutationMatrix`, `TestOwnedTransferInterpreterNative` — all PASS; `control:backend.runtime_causality` fires with exit 4 |
| T-02-08 | Digest mistaken for translation proof (incl. a manifest bound to a corrupted canonical projection) | high | mitigate | **MITIGATED** | `evidence/evidence.go:34` `DigestClaim = "content-identity-only"` (unchanged), `:220-221` claim + `KnownEscape` written into `lang.evidence/1`, `:332` bound in validation, `:379-382` bound into the ID preimage, `:169-172` explicit non-proof comment; **`:176-189` fail-closed canonical-projection guard** (`evidence.canonical_unstable` when the canonical bytes do not reparse cleanly or are not a formatter fixed point), placed *before* `check.Program` so no partial product is returned; `corevalidate/corevalidate.go:16-19` named escape; `session/session.go:582` escape reported outside controls | `TestCoordinatedSourceCoreEscapeIsNamed`, **`TestCanonicalRoundTripFailsClosed`** (2/2 subtests), **`TestCanonicalRoundTripAdmitsValidPrograms`**, `TestOwnedEvidenceBindings`, `TestOwnedEvidenceMutationMatrix`, `TestOwnershipProjectionIdentityParity`, `TestPhase1EvidenceGoldenUnchanged`, `TestVerifyPhase2ControlsAndWork`, `TestVerifyPhase2CLI` — all PASS |

**threats_open: 0** — no threat's declared mitigation is absent from implemented
code, and no threat is PARTIAL. Four non-blocking findings remain open
(SEC-02-B falsifier coverage, SEC-02-D spec purity, SEC-02-E test-support
ceiling, SEC-02-F cosmetic); none is an absent mitigation, so none counts toward
`threats_open` under `block_on: high`.

---

## T-02-03 — Deep Verification (handoff emphasis)

### Single global `cNames` allocator

*(Line references re-resolved at `da75e95`.)*

`internal/compiler/cgen/cgen.go:340-368` defines one allocator per emitted
translation unit. Both emitters enforce exactly one function per program
(`cgen.go:22-24`, `cgen.go:39-41`), so "per translation unit" is "global".

`emitMatch` (`cgen.go:114-124`) routes **every** variable identifier through it:

| Category | Line | Routed through allocator |
|----------|------|--------------------------|
| enum typedef name | `cgen.go:115` | yes |
| enum constants | `cgen.go:119` | yes (per alternative) |
| function name | `cgen.go:122` | yes |
| function parameter | `cgen.go:123` | yes |
| `_name` helper | `cgen.go:124` | yes |

`emitLinear` (`cgen.go:199-206`) reserves all fixed macro/struct/global/helper
names and routes every place local through the allocator at `cgen.go:206`.

`allocate` (`:355-368`) never returns a string already in `used` and records
every string it returns, so it is injective by construction. Two distinct core
identities therefore cannot receive the same C identifier.

### `cName()` lossy rune → `_` mapping

`cgen.go:375-387` collapses every non-`[A-Za-z0-9_]` rune to a single `_`;
`cLocal` (`:393-405`) does the same. Distinct source names *can* collapse to the
same preferred string — this is real. It cannot silently merge, because the
collapsed string is presented to `allocate`, which suffixes the second and later
claimants with `__LANG_<CATEGORY>_<ordinal>`.

Falsifier located and run: `session_test.go:374-388`
`TestNativeIdentifiersRemainCollisionFree` — PASS (1.45s). Its first case
(`let α`, `let β`, `let __1`) is exactly the lossy-merge case: `α` and `β` both
collapse to `lang_value__`. The test compiles the result with real clang under
`-Wall -Wextra -Werror -pedantic` and runs it at both `-O0` and `-O3`; a
collapsed pair would be a C redefinition error and fail the test. The suite also
covers same-name shadowing, `a`/`A`/`A_1` variant collapse, and a cross-category
`thing` vs `thing_LANG_THING` case.

### Reserved-name completeness — **finding SEC-02-A, now CLOSED**

The first pass found the reserved sets omitted 14 emitted fixed identifiers, and
that the real closure argument rested on an *unstated, untested* prefix
invariant. `dd9c0a8` fixes both halves and adds the missing enforcement.

**The invariant is now stated in code.** `cgen.go:47-75` names the two
cooperating properties — PREFIX CONFINEMENT (`cName` always yields
`^LANG_[A-Z0-9_]*$`, `cLocal` always yields `^lang_value_[A-Za-z0-9_]*$`, and
`allocate` only ever *appends*, so allocated names never leave those two
namespaces) and HONEST RESERVATION (the fixed lists are supersets of what the
emitters write). The comment says explicitly why over-reservation is inert and
under-reservation is the dangerous direction. The invariants are repeated at
`cName` (`:369-374`) and `cLocal` (`:389-392`).

**The reserved sets are now honest.** `matchFixedNames` (`:79-81`) adds `value`;
`linearFixedNames` (`:87-103`) adds `bytes`, `length`, `kind`, `id`,
`function_id`, `source_place`, `target_place`, `type_id`, `data`, `value`,
`hex`, `byte`, `escape`, `encoded`, `event`, `index`. Every name the first pass
listed as missing is present.

**The new test derives the fixed set instead of restating it.**
`cgen/cgen_names_test.go` emits two structurally identical programs whose source
names are pairwise disjoint, through *both* `Emit` and `EmitNative`, strips
comments/string literals/`#include` headers/numeric suffixes, and takes the
identifiers present in **both** outputs as exactly what the emitter contributes
itself. It then asserts honest reservation over that derived set (modulo an
explicit `cExternalNames` list of C keywords and libc names) and prefix
confinement over the source-derived complement. `leftOnly` sanity assertions
fail the test if the differential stops separating the two sets at all.

**Verified the test cannot be satisfied by widening the reserved list.** Two
out-of-tree probes against a `git archive HEAD` copy:

| Probe | Result |
|-------|--------|
| Emitter writes a new fixed local `lang_value_scratch` inside `lang_write_bytes` | `TestGeneratedIdentifierNamespacesStayConfined/{linear_Buffer, linear_Byte}` FAIL: *"emitter writes fixed identifiers that are neither reserved nor known external C names: [lang_value_scratch]"* |
| ...then silence it by adding `lang_value_scratch` to `linearFixedNames` | `TestReservedSetsCoverTheirOwnNamespace` FAIL: *"fixed identifier "lang_value_scratch" sits in the cLocal namespace; rename it or the closure argument becomes reservation-only"* |

So in the one namespace where reservation could paper over a real aliasing risk
(`lang_value_*`, where `cLocal` can actually produce a colliding preferred name),
widening the list is **forbidden outright** — the only accepted fix is a rename.
In the `LANG_*` and lowercase namespaces, adding a name to the list is a genuine
mitigation, because reservation is what makes `allocate` refuse to re-issue it.
Widening the list also cannot weaken the prefix-confinement half: that check
classifies by the two-fixture differential, not by the reserved set.

Additionally `TestIdentifierPrefixInvariance` directly falsifies property 1 over
26 adversarial inputs (empty, `α`, `日本語`, `a b`, `a-b`, `x__LANG_PLACE_1`,
every reserved lowercase name), and asserts the collision suffix never leaves
the namespace and never returns a reserved name.

**Verdict: MITIGATED.** SEC-02-A is closed.

*Residual, non-blocking:* the one remaining way to silence the honest-reservation
check without a rename is to add an identifier to `cExternalNames`
(`cgen_names_test.go:32-42`). That list is documented as requiring a deliberate
edit and is confined to C keywords and libc names; it is a review-visible
channel, not a silent one. Also, the file comment at `cgen.go:50` cites a
`names_internal_test.go` that does not exist — the enforcement actually lives
entirely in `cgen_names_test.go`. Cosmetic.

### `__LANG_` collision suffix — **observation SEC-02-D, carried forward OPEN**

`cgen.go:363` still builds the collision suffix as
`preferred + "__LANG_" + strings.ToUpper(category) + "_" + ordinal`. The leading
double underscore is reserved to the implementation by C17 §7.1.3. It compiles
clean under `-Wall -Wextra -Werror -pedantic` at both `-O0` and `-O3` today, and
it is not a security defect — it cannot cause aliasing, only nonconformance.

This was **deliberately not fixed** in the fix wave. The recommendation carried
forward is: change `__LANG_` to `_LANG_` in a **separate, standalone commit**,
before more C artifacts are frozen as goldens, so that the golden churn is
attributable to exactly that change and to nothing else. The window closes as
soon as another phase freezes generated-C goldens.

### Closed variants and ID uniqueness

`corevalidate/corevalidate.go` verified line by line:

- closed match/linear union — `core/core.go:102` `HasClosedBody()` is exclusive-or; enforced at `corevalidate.go:102` and schema-gated at `:105-111`
- duplicate type IDs — `:82`, `:436-448`
- duplicate type *names* — `:85`
- duplicate data alternatives — `:88-93`
- duplicate function IDs — `:99`
- duplicate arm IDs / edge IDs / patterns — `:133,136`
- exhaustive final claim — `:145`
- duplicate place IDs — `:174`, `:450-462`
- **duplicate operation IDs — `:198` `core.duplicate_operation_id`** (fires as a required gate control)
- duplicate point IDs — `:198`
- duplicate loan IDs — `:216`
- positional ID derivation, so IDs cannot be reordered — `:160,177,201,306`
- empty IDs rejected — `:421-425`
- validated program is a deep JSON clone (`:473-483`), so callers cannot mutate the validator's copy or race it

Falsifier: `TestOwnedClaimReorderRejected` (PASS) and the
`control:core.duplicate_operation_id` gate lane (fires).

---

## T-02-06 — Deep Verification (handoff emphasis)

### Complete enumeration of process spawns

`grep -rn "exec\.Command\|CommandContext\|CombinedOutput\|\.Output()\|\.Run()" --include='*.go' .`

| # | Site | Deadline | Bounded stdout | Bounded stderr | CombinedOutput |
|---|------|----------|----------------|----------------|----------------|
| 1 | clang **compile** — `native/native.go:71-75` | `ctx` from `WithTimeout(parent, r.Timeout)` `:68` (default 5s `:47`) | `boundedWriter` `:72-73` | **independent** `boundedWriter` `:72,74` | no |
| 2 | program **run** (per input) — `native/native.go:98-102` | **own** `runCtx` `:97`, cancelled per iteration `:103` | `boundedWriter` `:99-100` | **independent** `boundedWriter` `:99,101` | no |
| 3 | evidence probe `clang --version` — `evidence/evidence.go:103` → `runToolProbe:138-159` | `WithTimeout(parent, 5s)` `:139` | `boundedProbeWriter` `:142-143` | **independent** `boundedProbeWriter` `:142,144` | no |
| 4 | evidence probe `clang -dumpmachine` — `evidence/evidence.go:107` → same helper | same | same | same | no |

No other production spawn exists. `testsupport/testsupport.go` and
`session_test.go` are test harness only, but as of `2f2e2e2` and `8d87a1a` they
are held to the same discipline (below). `.planning/spikes/005-.../lab.go` is
archived spike code outside the shipped module path and is the *only* excluded
directory.

### 64 KiB plus one detection byte

`native/native.go:137-155`: the writer buffers up to `MaxStreamBytes + 1`
(`:144`) while counting **all** bytes written in `total` (`:143`), and
`overflowed()` is `total > MaxStreamBytes` (`:155`). The `+1` is a genuine
detection byte, not a truncation-and-accept. Identical construction for probes
at `evidence/evidence.go:126-136,149-153` and for test support at
`testsupport/testsupport.go:41-63`.

All four native stream caps have distinct error codes and a distinct falsifier
subtest: `TestNativeStreamsIndependentlyBounded/{compile_stdout, compile_stderr,
run_stdout, run_stderr}` — 4/4 PASS. Both probes have four:
`TestDefaultFactsBoundsBothToolProbes/{version_stdout, version_stderr,
target_stdout, target_stderr}` — PASS.

### Deadline falsifier — **finding SEC-02-B, still OPEN**

Re-checked at `da75e95`: `TestNativeHelperProcess`
(`native/native_test.go:172-200`) still implements flood modes only —
`compile-stdout-flood`, `compile-stderr-flood`, `run-stdout-flood`,
`run-stderr-flood`, and a default valid-document mode. There is **no hang mode
and no in-tree assertion of `native.timeout`**. `grep -n "native.timeout"
internal/compiler/native/native_test.go` returns nothing.

The mitigation itself is present and effective. I re-ran the out-of-tree deadline
probe against a `git archive da75e95` copy (a new test file in a scratch
directory; no repository file was modified), driving a `sleep 30` child through
the unexported `command` seam at each stage with `Runner.Timeout = 300ms`:

```
=== RUN   TestProbeTimeoutFires/compile
    stage=compile deadline fired in 301.593333ms with native.timeout
=== RUN   TestProbeTimeoutFires/run
    stage=run deadline fired in 305.687459ms with native.timeout
--- PASS: TestProbeTimeoutFires (0.61s)
```

Both stages honour their deadline and `exec.CommandContext` reaps the child.
**The gap is validation coverage, not an absent mitigation**, so it does not
count toward `threats_open`. It is also cheap to close: the probe above is ~40
lines and needs only a `hang` mode added to `TestNativeHelperProcess`.

### Repo-wide unbounded-spawn guard — **finding SEC-02-C, now CLOSED**

`8d87a1a` replaced the `native.go`-only grep with
`TestSourceNeverSpawnsUnboundedProcesses` (`native/native_test.go:66-115`).
Verified properties:

- **Scope is the whole module.** It walks from `runtime.Caller(0)` up three
  directories to the repo root, skipping only `.git` and `.planning/spikes/`,
  and reads every `*.go` file.
- **Allowlist is empty.** `unboundedSpawnAllowlist` (`:53`) is `map[string]string{}`
  with a comment requiring written justification per entry.
- **Cannot pass vacuously.** `scanned == 0` is a hard `t.Fatal` (`:112-114`).
- **Cannot match its own source.** Both needles are assembled at runtime
  (`"Combined" + "Output"`, `"exec." + "Command("`), so the scanner file itself
  does not trip the check — and, correspondingly, cannot be defeated by moving
  the offending call into the scanner's own file.
- **Two patterns, not one.** It rejects the merged-output helper *and* the
  context-free spawn constructor.

Two out-of-tree probes confirm it is load-bearing on exactly the files the first
pass said were unguarded:

| Probe (against `git archive HEAD` copy) | Result |
|---|---|
| add `exec.Command("echo").CombinedOutput()` to `internal/compiler/evidence/evidence.go` | FAIL — names `evidence.go` for **both** patterns |
| add the same to `internal/compiler/testsupport/testsupport.go` | FAIL — names `testsupport.go` for **both** patterns |

**Verdict: SEC-02-C is closed.**

*Residual, non-blocking:* the guard is a textual scan of two needles. It would
not catch `os.StartProcess`, `syscall.ForkExec`, an aliased
`var c = exec.Command`, or an `exec.CommandContext(context.Background(), ...)`
with no deadline attached. It is a regression tripwire for the two known
mistakes, not a proof of boundedness; the actual boundedness is proven by the
per-stream falsifiers above. Acceptable at ASVS L2.

### Test-support spawn bounds — **finding SEC-02-E (new, non-blocking)**

`2f2e2e2` brought `testsupport` up to the production discipline:
`BuildCLIErr` (`testsupport.go:88-104`) and `RunCLIErr` (`:132-150`) each take a
`context.WithTimeout` (`BuildCLITimeout` 5m, `RunCLITimeout` 2m), spawn via
`exec.CommandContext`, attach **two independent** `boundedWriter`s, and return
typed `CLIError` codes (`testsupport.build_timeout`, `run_timeout`,
`{build,run}_{stdout,stderr}_truncated`, `build_failed`, `run_failed`). The
writer uses the same `limit + 1` detection-byte construction as production. This
is a genuine fail-closed ceiling: a runaway or hung child fails the test with a
stable code rather than exhausting or hanging the test binary.

**On the 8 MiB figure specifically.** The stated rationale is that "CLI tests
legitimately capture whole `--json verify` corpora." I measured it. I patched
the scratch copy to log both stream totals at every test-support spawn and ran
`go test ./...`:

```
spawns=50  max_stdout=1756 bytes  max_stderr=344 bytes
```

The largest verify document in the suite is **1,756 bytes**. `MaxCLIStreamBytes
= 8 << 20` is therefore about **4,800×** the observed high-water mark, and the
rationale as written does not justify it — the corpus documents are ~1.7 KB, not
megabytes.

**Judgement: acceptable, but loosely calibrated — do not block.** `testsupport`
is test-only code that never runs in the shipped `lang` binary, so it is outside
the Phase 2 production trust boundary; the ceiling is hard, typed, and paired
with a deadline; and being too generous costs only a slower failure, never a
missed one. **Recommendation (non-blocking):** tighten `MaxCLIStreamBytes` to
`1 << 20` (matching `syntax.MaxSourceBytes`), which still leaves ~600× headroom
over the measured maximum, and replace the prose rationale with the measured
high-water mark so the number is defensible rather than asserted.

### Strictly one document on stdout

`native/native.go:168-192`: size re-check (`:169`), duplicate-JSON-key rejection
(`:172`, implemented `:194-241`), `DisallowUnknownFields` (`:177`), explicit
trailing-document rejection requiring `io.EOF` (`:181-187`), and a full
structural contract check (`:243-282`) covering schema, outcome kind/value,
non-empty events, empty live resources, per-event schema/ID/function-ID
agreement, duplicate event IDs, return-must-be-last, and per-schema field
shape. `run_stderr` is treated as fatal even when empty-of-error
(`native.go:118-120`), so stderr can never contaminate the fact stream.
Falsifier: `TestExecutionDecoderRejectsMalformedOutput` — **14/14 subtests PASS**
(missing fields, unknown field, duplicate document, trailing value, malformed,
truncated, oversized, duplicate key, unknown schema, unknown outcome, empty
outcome, unknown event, missing transition fields, return before transition).

### Bounded input reads

Every untrusted read is bounded with the same `limit + 1` detection pattern:

- `session/session.go:394-401` `readBoundedFile` = `io.ReadAll(io.LimitReader(file, limit+1))`
- source files — `session.go:123,154,227,308,345,374` at `syntax.MaxSourceBytes` (1 MiB)
- verification corpus — `session.go:448,471,500,603,618`, each followed by an explicit `len(source) > MaxSourceBytes` → `verify.fixture_input_limit` check (`:452,476,505,607,622`); corpus members are a fixed name list, not a directory glob
- evidence manifest — `session.go:370` at `evidence.MaxManifestBytes` (1 MiB), re-checked in `evidence.go:272-274` → `evidence.input_limit`
- parser self-bounds regardless of caller — `syntax/parser.go:42-44`, plus `MaxTokens` (`lexer.go:39`), `maxDeclarations`, `maxFunctions`, `maxAlternatives`, `maxLinearBindings` (`parser.go:10-19`) and type depth/node caps (`parser.go:37-38,252-253`)
- no `os.Stdin` read and no unbounded `io.ReadAll` exists anywhere in `cmd/` or `internal/`

### Generated-program output ceiling

The emitted C bounds its own document: `LANG_OUTPUT_LIMIT 65536u`
(`cgen.go:219`), every write funnelled through `lang_write_bytes` with an
underflow-safe remaining check (`cgen.go:226`), a fixed `LANG_EVENT_CAPACITY`
array with a pre-increment guard (`cgen.go:250`), fixed-size hex/byte buffers
(`cgen.go:245,269,273`), and data-argument-only formatting — no generated
format strings. Every write failure returns a stable nonzero exit (`74`) before
output can exceed the ceiling (`cgen.go:193,198-209`).

---

## T-02-02 / T-02-05 — Deep Verification (re-audit: the two admission layers)

The fix wave changed both admission layers in the same commits (`2549311`,
`3d9493a`). The specific risk this raises is that the layers stop being
independent — that one becomes a transcription of the other and the
differential stops being able to see a shared blind spot. Verified they did not.

### The layers remain structurally independent

`corevalidate` imports **only** `encoding/json`, `fmt`, `reflect`, and
`internal/compiler/core` (`corevalidate.go:6-11`). It does not import
`internal/compiler/ability`, `internal/compiler/check`, or
`internal/compiler/interp`, so it cannot be reusing production authorization
code. The package doc states the intent (`corevalidate.go:1-3`) and the import
list enforces it.

| Property | Production (`ability` / `check`) | Independent validator (`corevalidate`) |
|---|---|---|
| ability derivation | one pass computing a **whole `abilitySet`** per shape through a structural AND combiner (`ability.go:44-54`), returning granted set + negative witnesses | **per-requested-ability recursive descent**, one ability at a time, returning `(ok, path, known)` (`corevalidate.go:376-416`) |
| how they are reconciled | — | `reflect.DeepEqual` on **both** `fact.Abilities` and `fact.NegativeWitnesses`, failing `core.ability_mismatch` (`corevalidate.go:167`) |
| loan liveness | forward pass over **source AST bindings**, keyed by binding index, inheriting `loansForBinding[]int` and separately handling the result expression (`check.go:359-398`) | pass over **source-blind core operations**, keyed by place ID, carrying `loansForPlace[]string` and excluding `OpReturn` targets (`corevalidate.go:231-249`) |
| test oracle (third implementation) | `check_test.go` materializes the derivation relation as an explicit edge set and closes each loan by order-independent fixed-point iteration, then reduces to a maximum ordinal — a genuinely different algorithm from the single streaming forward pass | — |

The two implementations now agree on the same *law* (loan association is
transitive), which is unavoidable and correct: two admission layers that
disagreed on the law would be the defect. What matters is that they compute it
by different means over different data, and they do.

### Granting `Buffer` the `share` ability creates no escalation path

`ability.go:116-127` grants `Buffer` `{drop, share, send, escape}` and keeps
`copy` denied with the witness `{"Buffer"}` intact. `corevalidate.go:386-396`
re-derives the same row independently. Verified the blast radius by enumerating
every non-test use of `core.AbilityShare` in the module:

```
core/core.go:53                 constant definition
ability/ability.go:10,32        ordering + set accessor
check/check.go:282,285          the new borrow gate
corevalidate/corevalidate.go:281  core.ability.share_denied on OpBorrowShared
corevalidate/corevalidate.go:329  ordering
```

`share` gates **exactly one operation**: `OpBorrowShared`. It does not unlock
`copy`, `drop`, `send`, or `escape`; those are separate rows with independent
witnesses. A shared loan observes without duplicating ownership, and every other
ownership law still applies to a borrowed `Buffer` — `TestOwnershipMutationMatrix`
and `TestOwnershipSequenceExhaustive` PASS, and `control:ability.forged_copy`
still fires in the gate, so `Buffer` remains noncopyable. **No authority
escalation.**

The new source-layer gate (`check.go:282-292`) is defence in depth: it emits
`ownership.borrow_requires_share` via `ErrorWithRepairs` with the same cause
shape as the copy gate, converting what was previously a spanless operational
failure (exit 3, raised only by the source-blind validator) into a proper
invalid-program diagnostic (exit 2). Because every source-reachable constructor
currently grants `share`, the gate is unreachable from source today; that claim
is itself made executable by `TestShareIsUniversallyGrantedAfterBufferShare`
(PASS), which enumerates every shape to depth 4 and fails the moment a
constructor withholds `share` — so the gate cannot quietly become dead code that
nobody notices is now reachable.

### Transitive loan liveness

The pre-fix law expired a loan one hop early, so `let view = borrow code; let
review = borrow view; let delivered = take code; let observed = review` was
**wrongly accepted and ran to completion** — a real corrupted-move admission
(T-02-05), not a hypothetical. Both layers now carry loans transitively, and the
case is wired into the gate as the ninth control
(`testdata/phase2/reborrow_while_moved.lang`, expected diagnostic
`ownership.move_while_borrowed`, control ID
`control:ownership.move_while_reborrowed`). Falsifiers: `TestTransitiveLoanBlocksMove`
(core layer), the exhaustive `TestOwnershipSequenceExhaustive` differential
against the rewritten oracle, `FuzzOwnershipLinear`, and the gate lane itself —
all PASS. Dual controls in the same commit assert transitivity does not
*over*-block: a transitively derived loan with no use after the move still ends.

---

## Declared Residual Trust Boundary

These are **accepted, declared limitations**, not undetected defects. They are
recorded here so that no downstream consumer mistakes Phase 2 evidence for a
stronger claim than it makes.

### 1. `escape:coordinated-source-core-lie` — EXPECTED ESCAPE, never a control

`corevalidate.Validate` is deliberately **source-blind**
(`corevalidate/corevalidate.go:16-19,42-44`). It proves internal consistency of a
typed-core statement. It does **not** prove that a coordinated frontend
translated user source into that core truthfully. A producer that emits a false
source claim together with internally consistent core facts remains outside what
this validator can detect.

This limitation is:

- **named as a constant**, not prose — `KnownEscape = "escape:coordinated-source-core-lie"` (`corevalidate.go:19`)
- **asserted to be undetected** — `TestCoordinatedSourceCoreEscapeIsNamed` (`corevalidate_test.go:182-200`) constructs an internally consistent coordinated false claim and **fails if the validator reports it as detected** (PASS). This is the correct polarity: it forbids overclaiming detection.
- **reported in a separate protocol field** — `session/session.go:582` sets `result.ExpectedEscapes`; `protocol/protocol.go:61,84,90,178-180` renders it as `expected_escapes` / `expected_escape`, structurally disjoint from every lane's `controls` array
- **asserted never to be counted as a detected control** — `session_test.go:739-744` fails if `ExpectedEscapes` is missing *or* if the escape ID appears in any lane's controls (PASS); `testsupport/cli_test.go:253-254` re-asserts this through the real CLI (PASS)
- **observed correctly at runtime** — the fresh gate output above carries `"expected_escapes":["escape:coordinated-source-core-lie"]` while all five Phase 2 lanes carry only the eight genuine `control:` IDs

**Verified: the escape is recorded as a declared trust-boundary limitation and never as a detected control.**

### 2. Evidence claims content identity only, never translation proof

`lang.evidence/1` manifests carry `digest_claim: "content-identity-only"`
(`evidence/evidence.go:33,202`). The manifest proves that a specific canonical
source, core, C source, and ordered execution set are **mutually bound by
content digest**. It does **not** prove semantic translation correctness.

- the claim string is unchanged at `content-identity-only` (`evidence.go:34`) and is bound into the manifest ID preimage (`evidence.go:379-382`), so it cannot be stripped or altered without changing the evidence ID
- `Validate` enforces `evidence.digest_claim_mismatch` and `evidence.escape_mismatch` for schema-1 manifests (`evidence.go:332-333`)
- digest comparisons use `subtle.ConstantTimeCompare`
- the non-proof boundary is stated in code at the construction site (`evidence.go:169-172`), not only in documentation
- **new (`d9b370f`): the claim is now honest about the bytes it binds.** `SourceDigest` binds the *canonical projection*, not the bytes the user wrote. `evidence.go:176-189` therefore asserts the round trip instead of assuming it: if the canonical bytes do not reparse without diagnostics, or are not a fixed point of the formatter, `build` returns `evidence.canonical_unstable` and **no** product — `Product{}`, `nil` diagnostics. The guard sits *before* `check.Program`, so a corrupted projection can neither be laundered into a source diagnostic naming offsets in text nobody authored, nor reach manifest construction. Without it a COMPLETE MANIFEST could be issued bound to corrupted bytes, which is precisely the overclaim this section forbids.
- falsified by `TestCanonicalRoundTripFailsClosed` (2/2: `canonical_does_not_reparse`, `canonical_not_fixed_point`), which asserts all four properties — error present, error code exact, diagnostics empty, and `ManifestBytes`/`CanonicalSource` both nil — driven through an explicit `format`/`parse` seam so the cases stay valid once any underlying formatter defect is repaired; and by `TestCanonicalRoundTripAdmitsValidPrograms`, which asserts the admitted canonical source really is the formatter's fixed point (no over-blocking). Also `TestOwnedEvidenceBindings`, `TestOwnedEvidenceMutationMatrix`, `TestOwnershipProjectionIdentityParity` (all PASS)

**Verified: evidence claims content identity only (`lang.evidence/1`, `content-identity-only`) and makes no translation-proof claim.**

### 3. Operational observations excluded from semantic identity

Timing, output byte counts, and peak RSS are reported as operational facts and
never enter any digest or identity computation. Peak RSS is reported honestly as
`unavailable` rather than fabricated (observed in every lane of the fresh gate).

---

## Findings Ledger

| ID | Threat | Class | Blocking | Status | Summary |
|----|--------|-------|----------|--------|---------|
| SEC-02-A | T-02-03 | Hardening (defense-in-depth) | no | **CLOSED** (`dd9c0a8`) | Reserved sets are now honest supersets, the prefix-confinement invariant is documented on `newCNames`/`cName`/`cLocal`, and `cgen_names_test.go` derives the fixed-identifier set from generated output. Proven load-bearing by two auditor probes; proven **not** satisfiable by merely widening the reserved list, because `TestReservedSetsCoverTheirOwnNamespace` forbids any reserved entry in the `lang_value_` namespace. |
| SEC-02-B | T-02-06 | Falsifier coverage | no | **OPEN** | `native.timeout` still has no in-tree negative control; `TestNativeHelperProcess` implements flood modes only. Auditor re-proved both compile and run deadlines fire out of tree at `da75e95` (302 ms / 306 ms). Fix: add a `hang` mode to `TestNativeHelperProcess` and assert `native.timeout` at both stages. |
| SEC-02-C | T-02-06 | Falsifier coverage | no | **CLOSED** (`8d87a1a`) | `TestSourceNeverSpawnsUnboundedProcesses` walks the whole module with an empty allowlist, runtime-assembled needles, and a `scanned == 0` hard fail. Auditor probes confirm it catches an injected `CombinedOutput`/`exec.Command(` in **both** `evidence/evidence.go` and `testsupport/testsupport.go`. Residual: it is a two-needle textual tripwire, not a proof of boundedness. |
| SEC-02-D | — | Observation (spec purity) | no | **OPEN — deliberately deferred** | Allocator collision suffix `__LANG_` (`cgen.go:363`) leads with a double underscore, reserved to the implementation by C17 §7.1.3. Compiles clean under `-pedantic -Werror` at `-O0`/`-O3`. **Recommendation carried forward: change `__LANG_` → `_LANG_` in a separate standalone commit, before more C artifacts are frozen as goldens**, so the golden churn is attributable to exactly that change. |
| SEC-02-E | T-02-06 | Calibration (new this pass) | no | **OPEN** | `testsupport.MaxCLIStreamBytes = 8 MiB` is a genuine fail-closed, deadlined, per-stream ceiling on test-only code — but measurement shows the largest capture across all 50 test-support spawns is **1,756 bytes** stdout / 344 bytes stderr, so the ceiling is ~4,800× the observed maximum and the stated "whole verify-corpus documents" rationale does not support the figure. Recommendation: tighten to `1 << 20` and cite the measured high-water mark. |
| SEC-02-F | T-02-03 | Observation (cosmetic) | no | **OPEN** | `cgen.go:50` cites a `names_internal_test.go` that does not exist; the enforcement lives entirely in `cgen_names_test.go`. Also, the honest-reservation check can be silenced by adding an identifier to `cExternalNames` (`cgen_names_test.go:32-42`) — a documented, review-visible channel confined to C keywords and libc names. |

No finding is an absent mitigation, so `threats_open` is 0 and the phase is not
blocked under `block_on: high`.

## Unregistered Attack Surface

None. All seven plan summaries declare no new trust boundary
(`02-05-SUMMARY.md:162`, `02-06-SUMMARY.md:152`, `02-07-SUMMARY.md:147`), and
this pass re-checked the fix wave `d9b370f..da75e95` for surface added *after*
the plans were written. The wave added no new external interface: it added one
fixture and one control ID, one source-layer diagnostic
(`ownership.borrow_requires_share`), one evidence failure code
(`evidence.canonical_unstable`), deadlines and bounds on existing test-support
spawns, and three test files. The one behavioural widening — `Buffer` gaining
`share` — was traced to its complete set of five non-test uses and gates exactly
one operation. No unmapped attack surface was found.

## Verdict

**SECURED.** All eight declared mitigations are present in implemented code at
`HEAD` `da75e95`, each located at a cited `file:line` and each exercised by an
executed falsifier; the gate runs clean at exit 0 with Phase 1 at 18 recomputed
work units, Phase 2 at 52, and all **nine** required negative controls firing.

All eight threats are now **MITIGATED** — T-02-03 moves from PARTIAL to
MITIGATED because the reserved sets are honest, the prefix-confinement invariant
is stated in code, and the new test derives the fixed-identifier set from
generated output rather than restating a hand-maintained list. Two of the four
prior findings are closed (SEC-02-A, SEC-02-C), each confirmed by an executed
out-of-tree probe rather than by reading the fix commit. SEC-02-B remains an
open falsifier-coverage gap on `native.timeout` (the mitigation itself was
re-proved effective out of tree). SEC-02-D is carried forward unfixed by
developer decision, with the standing recommendation to change `__LANG_` to
`_LANG_` in a separate commit before more C goldens are frozen. Two new
non-blocking observations, SEC-02-E and SEC-02-F, are recorded above.

The honest-limits posture holds and is now stronger: the coordinated source/core
lie is still recorded as an expected escape and never as a detected control;
`digest_claim` is still `content-identity-only`; and evidence now additionally
refuses to issue a manifest at all when the canonical projection it would bind
is unstable.


## Close-Out Addendum (2026-09-04, `3399ddc`)

This audit was pinned to `da75e95`. `3399ddc` landed after it, fixing a formatter
regression (`syntax` package only). It touches no trust boundary, no process spawn,
no bound, and no generated C identifier, so every threat verdict above is unchanged.
Re-confirmed at `3399ddc`: `sh scripts/verify-phase2.sh` exit 0, Phase 1 work 18,
Phase 2 work 52, nine controls, `expected_escapes` disjoint from every lane's
`controls`.

The open non-blocking findings — SEC-02-B (`native.timeout` has no in-tree
falsifier), SEC-02-D (`__LANG_` reserved under C17 §7.1.3), SEC-02-E (8 MiB CLI
ceiling against a measured 1,756 B maximum), and SEC-02-F (stale comment reference) —
are carried into Phase 03 as `02-DEBT.md` D-02-04, D-02-05, D-02-06.

SEC-02-D carries a deadline rather than a priority: rename `__LANG_` to `_LANG_`
**before any additional C artifact is frozen**. It moves no golden today; every
future frozen artifact widens the freeze.
