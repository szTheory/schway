---
phase: "22"
slug: "native-application-build-and-single-execution"
status: verified
threats_open: 0
asvs_level: 1
created: "2026-09-27"
---

# Phase 22 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| Source and manifest to native builder | Program source and local binding declarations are untrusted input; paths and compiler authority are constrained before invoking Clang. | Lang source, JSON manifest, C source/header paths, compiler arguments |
| Builder to installed toolchain | Clang, the host SDK, and declared local C are trusted host components; arbitrary C behavior is outside the language semantic guarantee. | Fixed argv, staged input bytes, output artifact |
| CLI to retained application | The selected artifact is treated as an executable process; its streams and process outcome are surfaced without shell evaluation. | One bounded U64 argument, stdout/stderr, exit/signal/timeout/start result |
| Application to evidence report and replay verifier | Captured evidence is untrusted until its schema, identity, and outcome are checked; replay models only declared foreign outcomes. | Private capture path, artifact digest/build ID, report JSON, replay cases |

---

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-22-01 | Tampering / Elevation | Manifest path and include resolution | high | mitigate | `ResolveBindings` confines canonical paths and symlink targets to the manifest root and rejects undeclared inputs. `TestPhase22BindingsRejectInvalidInputs` covers absolute/parent paths, symlink escapes, and undeclared includes; relocation is covered by `TestPhase22RelocatedBindingsCLI`. | closed |
| T-22-02 | Tampering / Elevation | Compiler and application process dispatch | high | mitigate | Compiler and app launches use context-bound argument vectors. `TestPhase22BindingsBuildNeverLaunchesAndRunLaunchesOnce`, `TestPhase22RunApplicationPreservesStreamsAndProcessOutcomes`, and the ordinary-run routing test check launch count and explicit replay dispatch. | closed |
| T-22-03 | Tampering / Repudiation | Build and execution evidence identity | high | mitigate | The identity covers the declared source/header/toolchain inputs and binary; mutation and swapped-receipt controls are in `TestPhase22BuildIdentityDeclaredInputMutations` and `TestPhase22BuildIdentityReceiptAndBinaryMismatch`. | closed |
| T-22-04 | Denial of service | Generated U64 input and CLI token | medium | mitigate | The generated entry bounds transport length and validates decimal U64 input before the Lang body. Phase 22 emitter and CLI controls reject malformed, overlong, extra, and overflowing input. | closed |
| T-22-05 | Tampering / Repudiation | Evidence capture and sidecar report | high | mitigate | Capture/report states fail closed; receipts and reports bind to the artifact identity. Evidence state and partial/capacity tests include `TestPhase22CompleteEvidenceReportWriteFailureIsNotVerified`, which proves a complete capture is still not verified if report publication fails. | closed |
| T-22-06 | Spoofing | Trusted C implementation and modeled foreign outcomes | medium | transfer | `examples/phase22/BINDINGS.md` assigns arbitrary local C behavior to trusted audited input and excludes it from the semantic claim. Replay rejects local C execution and marks modeled host IO/cleanup as unsupported. This boundary remains outside Phase 22's guarantee. | closed |
| T-22-07 | Repudiation | Child process outcome classification | medium | mitigate | `TestPhase22RunApplicationPreservesStreamsAndProcessOutcomes` distinguishes nonzero exit, signal, timeout, and launch failure from compiler protocol errors. | closed |

*Status: open · closed · open — below high threshold (non-blocking)*
*Severity: critical > high > medium > low — only open threats at or above workflow.security_block_on count toward threats_open*
*Disposition: mitigate (implementation required) · accept (documented risk) · transfer (third-party)*

---

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|

No accepted risks. T-22-06 is a documented transfer boundary, not an accepted claim that arbitrary C is safe.

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-09-27 | 7 | 7 | 0 | Codex, ASVS Level 1 artifact and test review |

---

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-09-27
