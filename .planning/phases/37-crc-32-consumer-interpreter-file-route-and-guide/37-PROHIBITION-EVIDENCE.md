# Phase 37 prohibition evidence

Evidence date: 2026-10-15
Full checkout HEAD: `4f2eb354ce002878699d627b123dbec6ce9f0cf9`
Host/toolchain: darwin arm64; Node `v26.10.0`; Go cache `/tmp/ai-lang-verification-gocache`.
Adapter SHA-256: `e85d67e7487f4aa03022c1d0efa5e0becdc93bfc1bf45bbdf1af3d86c239086c`.
Plan SHA-256 after descriptor wiring: `ee401dac81b262544bbc3c94954cdb7362be57e1b58e40adb5450f15085f987f`.

This local proof runs the supported installed `projectProhibitions` projection, reconstructs each descriptor with `descriptorFromProjection`, and calls the canonical `check prohibition-enforcement` producer. The initial descriptorless/unresolved projection remains preserved in `37-PROHIBITION-PROJECTION.json`; this evidence proves only the two named existing Go boundaries. Structural causal acceptance is not CRC numeric replay.

## Native capture outcome prohibition

Statement: `MUST NOT turn optional capture exhaustion into application success or hide a later checked defect.`
Source: `internal/compiler/native/native_app.go` SHA-256 `0124d35186677f20a88704968de42c32a5d7f0fb6ea850533bd5b891629bac64`; adapter/test SHA-256 `e85d67e7487f4aa03022c1d0efa5e0becdc93bfc1bf45bbdf1af3d86c239086c`.
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
| bad | outcome: run=1, pass=0, fail=1, skip=0; causal: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=037079a643e3935803db9cb989709e2d3e9a084f2a104a74f2190780b2dd7251; causal=4e270f9b267af09a1e76b88b53bc195a900c186744d5a8a78b6419c8bc3435e1 | tests=2, pass=1, fail=1, skipped=0 | 1 | `ef63718dcb335230ca6bf1a6cd1136d66f14611b5f22552484fa5ba07e752fde` |
| clean | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=b535101c240f05560dab0b8d9ba166b2becb958cfe6aa5d7ab8b445e3d91b3d8; causal=5845a532a5b59e989931d972941a6c6e29587dfacfe3981604f4b128a6aa0fbe | tests=2, pass=2, fail=0, skipped=0 | 0 | `85334fc0a48e62e3fa991ac9c97634010504a6995d8ce948fd67c05bdf461042` |
| default | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=7c019be3c7fb9e2ed530815978507c46457bf281a0e24d806a0387a57733cb9d; causal=5845a532a5b59e989931d972941a6c6e29587dfacfe3981604f4b128a6aa0fbe | tests=2, pass=2, fail=0, skipped=0 | 0 | `f4c6475b1d0189e9be2c1118103e0972069eb49a9419730d6f8bce843dca84f9` |

Exact named root: `./internal/compiler/native -run '^TestPhase36CaptureOutcomeSeparation$' -count=1 -v -timeout=20s`. Expected violation assertion(s): `capture exhaustion changed successful work`; `capture exhaustion hid the later defect`. The bad subject changes only its hardcoded temporary Go overlay and reaches only these assertions; tracked source bytes are compared before/after and temp overlays are removed. Clean/default controls both pass without skips.

## Causal peer numeric-limit prohibition

Statement: `MUST NOT present causal structural acceptance as independent CRC numeric verification.`
Source: `internal/compiler/executionpeer/executionpeer.go` SHA-256 `867185ab16461dd079451ead43ae47c68c3339ac6db7b790430f54954a02393c`; adapter/test SHA-256 `e85d67e7487f4aa03022c1d0efa5e0becdc93bfc1bf45bbdf1af3d86c239086c`.
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
| bad | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=0, fail=1, skip=0; output SHA-256: outcome=af7871a298ebb183067ce90bcb2fabfb62909509ef9edaa4776731bcb9bf387d; causal=c886531596f18515293285368388cf20e65f232e41e34e782f4ec471ae65a739 | tests=2, pass=1, fail=1, skipped=0 | 1 | `3be9839699905be70c105212a8723f62cd4e447fe6fe3809d5f304f861f3c948` |
| clean | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=7a05c53c85df411d97b23fbb07efdbc4d13c0e328948a9b075aa033c9f9a9fa1; causal=5845a532a5b59e989931d972941a6c6e29587dfacfe3981604f4b128a6aa0fbe | tests=2, pass=2, fail=0, skipped=0 | 0 | `3f83252fd1b6f076bfe723b7aa9993725c3402e8116c9c46b0c87bc0c63506c2` |
| default | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=b9c49ec2863e185addf76ebf367b6e259eb155dd4f36f3ead9b692c26a686715; causal=5845a532a5b59e989931d972941a6c6e29587dfacfe3981604f4b128a6aa0fbe | tests=2, pass=2, fail=0, skipped=0 | 0 | `380d9f65b9a5d2a066bec37ccd03318e72c81249fb22df8b6251f05c9d1ab6a9` |

Exact named root: `./internal/compiler/executionpeer -run '^TestPhase36CallCausalAttributionControls$' -count=1 -v -timeout=20s`. Expected violation assertion(s): `causal peer should accept the structurally plausible false-success trace; numeric replay is not its claim`. The bad subject changes only its hardcoded temporary Go overlay and reaches only these assertions; tracked source bytes are compared before/after and temp overlays are removed. Clean/default controls both pass without skips.
