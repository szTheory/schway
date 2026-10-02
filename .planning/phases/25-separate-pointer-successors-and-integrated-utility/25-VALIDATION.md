---
phase: "25"
slug: separate-pointer-successors-and-integrated-utility
status: complete
nyquist_compliant: true
wave_0_complete: true
created: "2026-10-01"
---

# Phase 25 — Validation Strategy

> Local and hosted execution validation is complete. Run 36971855722 closes EVD-10 with native Linux/x86_64 and macOS/arm64 receipts at merge revision `a90c27c5b432ef6fc59fbafaa68b50a1374ae138`.

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing standard library; Go version constrained by the project to 1.24 |
| **Config file** | go.mod |
| **Quick run command** | go test -count=1 -run '^TestPhase25' ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/pathoracle ./internal/compiler/interp ./internal/compiler/session ./internal/compiler/cgen ./internal/compiler/native |
| **Full suite command** | go test ./... |
| **Estimated runtime** | Not measured during planning; report cold and warm distributions in the Phase 25 evidence script. |

## Sampling Rate

- After each task, run the focused command in this map.
- After each plan wave, run the quick Phase 25 command above.
- Before phase verification, run `go test -count=1 ./...`, relevant race/vet/build checks, and `scripts/verify-phase25.sh` on both configured hosts through existing CI ownership.
- Keep focused feedback below 60 seconds where practical. Record cold and warm distributions for the integrated build/run/observe/repair loop.
- Evidence is a matrix by foreign/shared/exclusive family × macOS/Linux × applicable baseline, optimized, and sanitizer lane. A missing row remains incomplete; replay or cross-compilation does not substitute for a native host run.

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Grade | Non-inertness | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|-------|---------------|--------|
| T-25-01 | 01 | 1 | NAT-11, NAT-13 | T-25-01 | Shared source family is checked as read/copy-only and exact-shape | checker/source positive and conflict control | `go test -count=1 -run '^TestPhase25Shared(PointerSuccessor|Source)' ./internal/compiler/check` | Created in task | EXERCISED | — | ✅ local pass |
| T-25-02 | 01 | 1 | NAT-11, NAT-13 | T-25-02, T-25-03 | Actual shared pointer C, call-site representation, and manifest agree without unsupported attributes | generated C/manifest contract | `go test -count=1 -run '^TestPhase25Shared(PointerABI|PointerManifest|NativeShape)' ./internal/compiler/cgen ./internal/compiler/native` | Created in task | EXERCISED | — | ✅ local pass |
| T-25-03 | 02 | 2 | NAT-11, NAT-12, NAT-13 | T-25-04, T-25-06 | Core and origin peers independently reject wrong family, conflict, and escape | independent peer mutations | `go test -count=1 -run '^TestPhase25(PointerFamily|PointerOrigin)' ./internal/compiler/corevalidate ./internal/compiler/originvalidate` | Created in task | EXERCISED | — | ✅ local pass |
| T-25-04 | 02 | 2 | NAT-11, NAT-12 | T-25-05, T-25-06 | CFG endpoints and straight-line loan replay separately cover compatible, sequential, overlapping, and borrowed-result escape paths; empty/forged endpoint claims cannot steer straight-line replay | independent path-liveness controls | `go test -count=1 -run '^(TestPhase25PointerPathLiveOverlap|TestPhase25PointerPathBorrowedResultEscape|TestPhase25IndependentPeerMutations)$' ./internal/compiler/pathoracle ./internal/compiler/session` | Created in task | EXERCISED | — | ✅ local pass at 2026-10-02 follow-up |
| T-25-05 | 03 | 3 | NAT-12, NAT-13 | T-25-07 | Exclusive helper and bounded owner-transfer composition check with independent U64-copy origin semantics and 0x43 ordering | source/checker and peer integration | `go test -count=1 -run '^TestPhase25(ExclusiveSource|TransferCallerComposition)' ./internal/compiler/check && go test -count=1 -run '^TestPhase25(U64CopyOrigin|PointerFamily|PointerOrigin)' ./internal/compiler/corevalidate ./internal/compiler/originvalidate` | Created in task | EXERCISED | — | ✅ local pass |
| T-25-06 | 03 | 3 | NAT-12, NAT-13, DX-15 | T-25-08, T-25-09 | Each family has distinct conflict/escape evidence; all three peer verdicts, command admission, source/cause spans, schema, function identity, and absent unsafe repairs are exercised | source and independent peer negative controls | `go test -count=1 -run '^(TestPhase25IndependentPeerMutations|TestPhase25CheckCommandPeerGate|TestPhase25CheckCommandPeerObservation|TestPhase25EscapeDiagnosticSourceAttribution|TestPhase25EscapeDiagnosticFunctionIdentity|TestPhase25FamilyConflict)$' ./internal/compiler/session` | Created in task | EXERCISED | — | ✅ local pass at 2026-10-02 follow-up |
| T-25-07 | 04 | 4 | NAT-12, NAT-13 | T-25-10 | Exclusive C/manifest agree; local wrong-result is reached; unsupported shapes fail before serialization | emitter/native negative controls | `go test -count=1 -run '^TestPhase25Exclusive(PointerABI|PointerManifest|WrongResult|Refusal|PointerRefusal)$' ./internal/compiler/cgen ./internal/compiler/native` | Created in task | EXERCISED | — | ✅ local pass |
| T-25-08 | 04 | 4 | NAT-12, NAT-13, DX-15 | T-25-11, T-25-12 | Model answer and error ordering are stable; exact owner-result flow and resource discharge are independently checked; source-expressible ownership boundaries and serializer-only pointer mutations have controls | model, source refusal, peer, serializer, and diagnostic controls | `go test -count=1 -run '^(TestPeerCalleeFrameDrained|TestPhase25(OwnerTransfer|InterpreterComposition|ErrorBeforeHelpers|StructuredDiagnostic|UnsupportedPointerShape|OwnershipDiagnosticBoundary))' ./internal/compiler/corevalidate ./internal/compiler/interp ./internal/compiler/diagnostic ./internal/compiler/session && go test -count=1 -run '^TestPhase25Exclusive(PointerABI|PointerManifest|WrongResult|Refusal)' ./internal/compiler/cgen ./internal/compiler/native` | Created in task | EXERCISED | — | ✅ local pass |
| T-25-09 | 05 | 5 | NAT-11, NAT-12, NAT-13 | T-25-13 | Exact shared U64-copy ABI and Phase 25 caller emit natively; path and origin peers independently accept only the same exact chain; utility returns 65/66 and typed 0x43 failure with native reached wrong-result controls | cgen/application/independent peer/native integration | `go test -count=1 -run '^TestPhase25SharedPointerCopyABI' ./internal/compiler/cgen && go test -count=1 -run '^TestPhase25UtilityOwnerTransfer' ./internal/compiler/pathoracle ./internal/compiler/originvalidate && go test -count=1 -run '^TestPhase25(Utility|NativeFamilyWrongResult)' ./internal/compiler/native` | Created in task | EXERCISED | — | ✅ local pass |
| T-25-10 | 06 | 6 | EVD-10, DX-14, NAT-11, NAT-12, NAT-13 | T-25-14, T-25-15 | Every family × host × applicable lane is indexed and fail-closed; README reproduces utility results; the Evidence index has section-scoped local file-link integrity with a missing-target negative control; existing family fixtures use bounded subprocesses; living claims separate source, local/hosted checks, and history | evidence script, README file-link integrity, bounded native controls, and living-document contract | `go test -count=1 -run '^(TestSourceNeverSpawnsUnboundedProcesses|TestPhase25(EvidenceScript|EvidenceIndex|LivingRoadmap))$' ./internal/compiler/native && sh scripts/verify-phase25.sh` | Created in task | EXERCISED | — | ✅ local pass; ✅ hosted run 36971855722, 18/18 rows |

## Hosted evidence closeout — 2026-10-02

**Newly executed hosted evidence:** GitHub Actions run `36971855722` tested
branch head `421b5b94eb867e940a207dfd64d7971dc8198172` at PR merge revision
`a90c27c5b432ef6fc59fbafaa68b50a1374ae138`. On both native hosts, the `checks`
job passed vet, build, the full Go suite, and race tests; the `current evidence
aggregate` job passed, including `scripts/verify-phase25.sh`. The Linux job ran
Go `linux/amd64` on `x86_64-pc-linux-gnu` with Ubuntu Clang 18.1.3. The macOS
job ran Go `darwin/arm64` on `arm64-apple-darwin25.6.0` with Apple Clang 21.

Each matrix row below is a native execution at that same merge revision. All
18 rows passed; every row expected and returned 65 and 66, with the inherited
typed `0x43` use error observed before either helper call.

| Host / target | Family | Lane | Result |
|---|---|---|---|
| Linux/x86_64, Ubuntu Clang 18.1.3 | foreign | baseline `-O0` | pass; 65/66; typed 0x43 before helpers |
| Linux/x86_64, Ubuntu Clang 18.1.3 | foreign | optimized `-O2` | pass; 65/66; typed 0x43 before helpers |
| Linux/x86_64, Ubuntu Clang 18.1.3 | foreign | ASan+UBSan | pass; 65/66; typed 0x43 before helpers |
| Linux/x86_64, Ubuntu Clang 18.1.3 | shared | baseline `-O0` | pass; 65/66; typed 0x43 before helpers |
| Linux/x86_64, Ubuntu Clang 18.1.3 | shared | optimized `-O2` | pass; 65/66; typed 0x43 before helpers |
| Linux/x86_64, Ubuntu Clang 18.1.3 | shared | ASan+UBSan | pass; 65/66; typed 0x43 before helpers |
| Linux/x86_64, Ubuntu Clang 18.1.3 | exclusive | baseline `-O0` | pass; 65/66; typed 0x43 before helpers |
| Linux/x86_64, Ubuntu Clang 18.1.3 | exclusive | optimized `-O2` | pass; 65/66; typed 0x43 before helpers |
| Linux/x86_64, Ubuntu Clang 18.1.3 | exclusive | ASan+UBSan | pass; 65/66; typed 0x43 before helpers |
| macOS/arm64, Apple Clang 21 | foreign | baseline `-O0` | pass; 65/66; typed 0x43 before helpers |
| macOS/arm64, Apple Clang 21 | foreign | optimized `-O2` | pass; 65/66; typed 0x43 before helpers |
| macOS/arm64, Apple Clang 21 | foreign | ASan+UBSan | pass; 65/66; typed 0x43 before helpers |
| macOS/arm64, Apple Clang 21 | shared | baseline `-O0` | pass; 65/66; typed 0x43 before helpers |
| macOS/arm64, Apple Clang 21 | shared | optimized `-O2` | pass; 65/66; typed 0x43 before helpers |
| macOS/arm64, Apple Clang 21 | shared | ASan+UBSan | pass; 65/66; typed 0x43 before helpers |
| macOS/arm64, Apple Clang 21 | exclusive | baseline `-O0` | pass; 65/66; typed 0x43 before helpers |
| macOS/arm64, Apple Clang 21 | exclusive | optimized `-O2` | pass; 65/66; typed 0x43 before helpers |
| macOS/arm64, Apple Clang 21 | exclusive | ASan+UBSan | pass; 65/66; typed 0x43 before helpers |

Cold feedback distributions were Linux 9.350/9.540/9.980s and macOS
12.470/12.670/12.930s; warm distributions were Linux 0.390/0.490/0.550s and
macOS 0.820/1.400/2.130s (min/median/max). The script's absent-other-host rows
in either individual job are local to that invocation; the paired native job
passed those rows on its own host. No row is inferred from replay or
cross-compilation. EVD-10 is closed by the two successful native aggregate
jobs together.

**Source inspection, local checks, and history:** source inspection and the
security follow-up below describe code facts, not hosted execution. Local
macOS receipts and earlier Phase 25 records remain scoped to their own source
revisions. The 2026-10-02 security follow-up below predates this hosted matrix
and remains a record of its specific local peer/diagnostic checks.

## Security audit follow-up — 2026-10-02

Source inspection at code revision `b19446f` confirms that the command gate
runs core validation, origin validation, then the independent path oracle;
the path oracle derives straight-line loan ancestry, last use, access-family
overlap, and terminal borrowed-result escape from operation/place/type facts,
without consulting `LoanEndpoints`. `RecomputeEndpoints` remains limited to
CFG endpoint synthesis. Origin refusals carry structured function and return
operation identity; the command projects that identity through checked core
facts and lossless source tokens to the terminal result and creating borrow.

Fresh local checks passed with `GOCACHE=/private/tmp/schway-gocache`:

- `TestPhase25PointerPathLiveOverlap`,
  `TestPhase25PointerPathBorrowedResultEscape`,
  `TestPhase25IndependentPeerMutations`, and
  `TestPhase25CheckCommandPeerGate` and `TestPhase25CheckCommandPeerObservation`.
- `TestPhase25EscapeDiagnosticSourceAttribution` and
  `TestPhase25EscapeDiagnosticFunctionIdentity`, plus the retained conflict,
  schema, unsupported-shape, and ownership-boundary controls.
- The final combined verification reran those controls and
  `TestLanguageMaturityCountsAreCurrent`.

The direct overlap and borrowed-result escape controls reject empty and
forged endpoint declarations. Core validation reports the overlap refusal and
marks the escaped function non-callable in its independently derived
signature; origin validation reports `core.origin_omitted`; pathoracle reports
its own overlap/escape refusal. The actual `CheckCommandFile` cases refuse
both negative source witnesses. Escape diagnostics select exact nonempty
primary and borrow-origin cause spans, retain `lang.diagnostic/0` and
`core.origin_omitted`, and carry no repair. Existing Phase 25 native receipts
remain historical at their recorded revisions. No hosted Phase 25 matrix was
run by this follow-up; that statement describes only that follow-up. The later
hosted evidence closeout above closes EVD-10.

## Wave 0 Requirements

No separate Wave 0 setup was needed. Each task created its focused tests and source fixtures alongside implementation. Local focused tests, the repository-wide suite, and the native evidence script are recorded during execution. Hosted dual-host run 36971855722 completed the required matrix: all foreign/shared/exclusive families passed baseline, optimized, and ASan+UBSan lanes on Linux/x86_64 and macOS/arm64 (18/18 rows); EVD-10 is closed.

## Manual-Only Verifications

All objective behavioral criteria have automated checks. macOS and Linux receipts must originate on their native host lanes; a local run on one host cannot stand in for the other. The automated `TestPhase25EvidenceIndex` check owns direct navigation by resolving every local file link in the heading-bounded Evidence index and rejecting a missing-target negative control. Human review remains responsible for the subjective question of whether a first-time reader can understand the clean-checkout README sequence; link resolution does not establish comprehension.

## Validation Sign-Off

- [x] All tasks have automated verification or a Wave 0 dependency.
- [x] Sampling continuity: no three consecutive tasks lack automated verification.
- [x] Wave 0 covers all missing references.
- [x] No watch-mode flags are used.
- [x] Focused feedback latency is below 60 seconds where practical and cold/warm distributions are recorded.
- [x] Every required host/family/lane result is present with no silent skips (18/18 in hosted run 36971855722).
- [x] nyquist_compliant: true set in frontmatter after execution validation.

**Approval:** local and hosted validation complete; EVD-10 is closed by run 36971855722.
