# Phase 37 prohibition evidence

Evidence date: 2026-10-16
Full checkout HEAD: `9258a7360222ff17a3267e1ccb873a0dfae79346`
Host/toolchain: darwin arm64; Node `v26.10.0`; Go cache `/tmp/ai-lang-verification-gocache`.
Adapter SHA-256: `287d5dd803c040fe8d42ad9161fa2dd635c6d820207c9f961502adfc7e751e88`.
Plan SHA-256 after descriptor wiring: `37-01-PLAN.md=ee401dac81b262544bbc3c94954cdb7362be57e1b58e40adb5450f15085f987f`; `37-02-PLAN.md=b6d305297e9d6534df63a171b01e3e9e7ddb821810721455b9b530444fc0d488`.

This local proof runs the supported installed `projectProhibitions` projection, reconstructs each descriptor with `descriptorFromProjection`, and calls the canonical `check prohibition-enforcement` producer. The initial descriptorless/unresolved projection remains preserved in `37-PROHIBITION-PROJECTION.json`; this evidence proves the two existing Go boundaries and the interpreter's capture-disabled FileBytes outcome boundary. Structural causal acceptance is not CRC numeric replay.

## Native capture outcome prohibition

Statement: `MUST NOT turn optional capture exhaustion into application success or hide a later checked defect.`
Source: `internal/compiler/native/native_app.go` SHA-256 `0124d35186677f20a88704968de42c32a5d7f0fb6ea850533bd5b891629bac64`; adapter/test SHA-256 `287d5dd803c040fe8d42ad9161fa2dd635c6d820207c9f961502adfc7e751e88`.
Violation fixture: `.planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/prohibition-outcome-violation.json` SHA-256 `7f38214b22d1748b5290a999c6a6e3ec9b4776596f79b006b2fbfd591c111654`; clean fixture: `.planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/prohibition-clean.json` SHA-256 `6f1aa9862183245c6523b25829cee150ef4926953da22c5fbfec5ff83ee0e27a`.
Canonical request SHA-256: `cab0ebf78aee5eac8817f9926767944bfed11931c8e61c552ab1173efa476a69`; producer stdout SHA-256: `6fdefa9dfee70ee1df10d0edb016fc7754732f205590901910cb5a4ca0bbba2a`; producer verdict: `green`, `flagged=false`, `kind=node-test`, `failFirstProof=violation-fixture`.

Canonical request:

```json
{
  "prohibition": {
    "statement": "MUST NOT turn optional capture exhaustion into application success or hide a later checked defect.",
    "status": "resolved",
    "verification": "test",
    "check_kind": "node-test",
    "check_target": ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-prohibition-check.test.cjs",
    "check_violation_fixture": ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/prohibition-outcome-violation.json",
    "check_clean_fixture": ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/prohibition-clean.json"
  },
  "check": {
    "kind": "node-test",
    "target": ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-prohibition-check.test.cjs",
    "violationFixture": ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/prohibition-outcome-violation.json",
    "cleanFixture": ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/prohibition-clean.json"
  },
  "mode": "auto"
}
```

| Subject | Go root counts and output digests | Direct node-test TAP counts | Exit | TAP SHA-256 |
| --- | --- | --- | ---: | --- |
| bad | outcome: run=1, pass=0, fail=1, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=db82a35ae426c283b1dfb747c9c736ff96bbe5a59fe92901cdf81684ac13f6fa; causal=4e270f9b267af09a1e76b88b53bc195a900c186744d5a8a78b6419c8bc3435e1; interpreterOutcome=e37184e9c5c0300a2945374ffef3176eb671918780bfc1f4c6feddef7a333f98 | tests=3, pass=2, fail=1, skipped=0 | 1 | `244ae96405173f695b7d934074f1d4d594c053a8a8fa4439bb02622fa895d571` |
| clean | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=219b9f68174b7c7aad00ae63c5402dba08d7eee4b108bc4e336f49a700f742bf; causal=4e270f9b267af09a1e76b88b53bc195a900c186744d5a8a78b6419c8bc3435e1; interpreterOutcome=e37184e9c5c0300a2945374ffef3176eb671918780bfc1f4c6feddef7a333f98 | tests=3, pass=3, fail=0, skipped=0 | 0 | `97c70a81322fe08c1bf8a65d7eb4be2866fcce0b7302f33ca1dbc8481fb4e0f9` |
| default | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=a05a4362ed640d1439a2bff13cb6fcc75f29eb8f2d6e9c2401fc9aeadf56180b; causal=4e270f9b267af09a1e76b88b53bc195a900c186744d5a8a78b6419c8bc3435e1; interpreterOutcome=e37184e9c5c0300a2945374ffef3176eb671918780bfc1f4c6feddef7a333f98 | tests=3, pass=3, fail=0, skipped=0 | 0 | `30d4414be79a2d2a1e4dac69c299093b647c0f8670336c140dd028a59f6843e3` |

Exact named root: `./internal/compiler/native -run '^TestPhase36CaptureOutcomeSeparation$' -count=1 -v -timeout=20s`. Expected violation assertion(s): `capture exhaustion changed successful work`; `capture exhaustion hid the later defect`. The bad subject changes only its hardcoded temporary Go overlay and reaches only these assertions; tracked source bytes are compared before/after and temp overlays are removed. Clean/default controls both pass without skips.

## Causal peer numeric-limit prohibition

Statement: `MUST NOT present causal structural acceptance as independent CRC numeric verification.`
Source: `internal/compiler/executionpeer/executionpeer.go` SHA-256 `a884b40df2c5ef2dcb51082d2b6b6400b56bb7d07bf9c72e9a2f7e45a5124ba4`; adapter/test SHA-256 `287d5dd803c040fe8d42ad9161fa2dd635c6d820207c9f961502adfc7e751e88`.
Violation fixture: `.planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/prohibition-causal-violation.json` SHA-256 `ca7150830d5d654a71a3dfcce53790efa33d4fe9486ea4e0fc8490b3a360b2d8`; clean fixture: `.planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/prohibition-clean.json` SHA-256 `6f1aa9862183245c6523b25829cee150ef4926953da22c5fbfec5ff83ee0e27a`.
Canonical request SHA-256: `c3a08cb3989970e0b95f68cdbc4b57282195d74d9d8aac23240f51584cd29e80`; producer stdout SHA-256: `6fdefa9dfee70ee1df10d0edb016fc7754732f205590901910cb5a4ca0bbba2a`; producer verdict: `green`, `flagged=false`, `kind=node-test`, `failFirstProof=violation-fixture`.

Canonical request:

```json
{
  "prohibition": {
    "statement": "MUST NOT present causal structural acceptance as independent CRC numeric verification.",
    "status": "resolved",
    "verification": "test",
    "check_kind": "node-test",
    "check_target": ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-prohibition-check.test.cjs",
    "check_violation_fixture": ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/prohibition-causal-violation.json",
    "check_clean_fixture": ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/prohibition-clean.json"
  },
  "check": {
    "kind": "node-test",
    "target": ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-prohibition-check.test.cjs",
    "violationFixture": ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/prohibition-causal-violation.json",
    "cleanFixture": ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/prohibition-clean.json"
  },
  "mode": "auto"
}
```

| Subject | Go root counts and output digests | Direct node-test TAP counts | Exit | TAP SHA-256 |
| --- | --- | --- | ---: | --- |
| bad | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=0, fail=1, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=b535101c240f05560dab0b8d9ba166b2becb958cfe6aa5d7ab8b445e3d91b3d8; causal=fc03e81cc84d6392efa4c05154332510b5a06c8aaf3c6782cabbf6a2a12eced6; interpreterOutcome=e37184e9c5c0300a2945374ffef3176eb671918780bfc1f4c6feddef7a333f98 | tests=3, pass=2, fail=1, skipped=0 | 1 | `da2f8548af2ded9509beb125881789007614dc0adfbd55e5c5fd8bfaf67bc82b` |
| clean | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=7c019be3c7fb9e2ed530815978507c46457bf281a0e24d806a0387a57733cb9d; causal=4e270f9b267af09a1e76b88b53bc195a900c186744d5a8a78b6419c8bc3435e1; interpreterOutcome=e37184e9c5c0300a2945374ffef3176eb671918780bfc1f4c6feddef7a333f98 | tests=3, pass=3, fail=0, skipped=0 | 0 | `947e00ec68f428124cf1169ab6415118880a5182b28e7be6f25480e540455b2d` |
| default | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=7c019be3c7fb9e2ed530815978507c46457bf281a0e24d806a0387a57733cb9d; causal=4e270f9b267af09a1e76b88b53bc195a900c186744d5a8a78b6419c8bc3435e1; interpreterOutcome=e37184e9c5c0300a2945374ffef3176eb671918780bfc1f4c6feddef7a333f98 | tests=3, pass=3, fail=0, skipped=0 | 0 | `882dee2a0aa058aa13dde31c58a38d0fff4c4be55d188fb295ec1b37a21e6a4c` |

Exact named root: `./internal/compiler/executionpeer -run '^TestPhase36CallCausalAttributionControls$' -count=1 -v -timeout=20s`. Expected violation assertion(s): `causal peer should accept the structurally plausible false-success trace; numeric replay is not its claim`. The bad subject changes only its hardcoded temporary Go overlay and reaches only these assertions; tracked source bytes are compared before/after and temp overlays are removed. Clean/default controls both pass without skips.

## Interpreter FileBytes capture outcome prohibition

Statement: `MUST NOT let capture exhaustion replace the interpreter's actual FileBytes application outcome or prior output.`
Source: `internal/compiler/interp/interp.go` SHA-256 `aba3e9245fb5189527faf0b7e95757bfc8c8c5985f063701b0f697dbdfd2afbd`; adapter/test SHA-256 `287d5dd803c040fe8d42ad9161fa2dd635c6d820207c9f961502adfc7e751e88`.
Violation fixture: `.planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/prohibition-interpreter-violation.json` SHA-256 `de5d195ea4e163baff0dcebedd0654222ad8a24001fad01a3731ae490340119b`; clean fixture: `.planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/prohibition-clean.json` SHA-256 `6f1aa9862183245c6523b25829cee150ef4926953da22c5fbfec5ff83ee0e27a`.
Canonical request SHA-256: `bebb1693768604a59003c8f479e85495a17ccbf0a5c4323e45802fa3a9e1daff`; producer stdout SHA-256: `6fdefa9dfee70ee1df10d0edb016fc7754732f205590901910cb5a4ca0bbba2a`; producer verdict: `green`, `flagged=false`, `kind=node-test`, `failFirstProof=violation-fixture`.

Canonical request:

```json
{
  "prohibition": {
    "statement": "MUST NOT let capture exhaustion replace the interpreter's actual FileBytes application outcome or prior output.",
    "status": "resolved",
    "verification": "test",
    "check_kind": "node-test",
    "check_target": ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-prohibition-check.test.cjs",
    "check_violation_fixture": ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/prohibition-interpreter-violation.json",
    "check_clean_fixture": ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/prohibition-clean.json"
  },
  "check": {
    "kind": "node-test",
    "target": ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-prohibition-check.test.cjs",
    "violationFixture": ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/prohibition-interpreter-violation.json",
    "cleanFixture": ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/prohibition-clean.json"
  },
  "mode": "auto"
}
```

| Subject | Go root counts and output digests | Direct node-test TAP counts | Exit | TAP SHA-256 |
| --- | --- | --- | ---: | --- |
| bad | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=0, fail=1, skip=0; output SHA-256: outcome=f475fd3f706010603ef9732b04d9651ac38ac9e3a8768d0b4df0a6a28b9c4a60; causal=4e270f9b267af09a1e76b88b53bc195a900c186744d5a8a78b6419c8bc3435e1; interpreterOutcome=1c53b89c1075f4283de624504113af5b799f567f1d83c1795d2f7b77f67aff5b | tests=3, pass=2, fail=1, skipped=0 | 1 | `3bccffb227abc1d8b89f4fda78bc0a7d1f913e7eabd08cc5ba279257e19bd4e0` |
| clean | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=fa8bfc9e01addb359fbb8e4ceff433e42a2fa641145965b9b848a9096b5330be; causal=4e270f9b267af09a1e76b88b53bc195a900c186744d5a8a78b6419c8bc3435e1; interpreterOutcome=e37184e9c5c0300a2945374ffef3176eb671918780bfc1f4c6feddef7a333f98 | tests=3, pass=3, fail=0, skipped=0 | 0 | `c86b33c3c2f5c74ae77cb90bb1392d121b0e279bb464c3f08494beb50298fa8d` |
| default | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=b79828435f73a24f4f356011e2df816ae09607697d58a3ec48706a87c6f12142; causal=0a0cfeaaa7eb93a612090418d80c49890ad0e30aca368a1b37050d09a91341a5; interpreterOutcome=e37184e9c5c0300a2945374ffef3176eb671918780bfc1f4c6feddef7a333f98 | tests=3, pass=3, fail=0, skipped=0 | 0 | `269384ca94ef132a1f8b446621dafb5116e8bc019d3d55aac62f763d49d992b5` |

Exact named root: `./internal/compiler/interp -run '^TestPhase37CaptureDisabledOutcomes$' -count=1 -v -timeout=20s`. Expected violation assertion(s): `want late defect preserved after evidence exhaustion`. The bad subject changes only its hardcoded temporary Go overlay and reaches only these assertions; tracked source bytes are compared before/after and temp overlays are removed. Clean/default controls both pass without skips.
