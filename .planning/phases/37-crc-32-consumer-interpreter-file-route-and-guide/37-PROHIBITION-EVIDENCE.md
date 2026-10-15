# Phase 37 prohibition evidence

Evidence date: 2026-10-15
Full checkout HEAD: `c210004ef940a6e2556fc63d5c1d53e683e9118e`
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
| bad | outcome: run=1, pass=0, fail=1, skip=0; causal: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=746ec8711a3f33c64a924d78b6341ca2a4e601a337a2907d26039b8a31edd745; causal=4e270f9b267af09a1e76b88b53bc195a900c186744d5a8a78b6419c8bc3435e1 | tests=2, pass=1, fail=1, skipped=0 | 1 | `2abe168f69774f8fe5097c6e061a94f3b0f535ad69ed021e84e4ddef270c3c44` |
| clean | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=f61c9e948f5f95a558e9d2929caff215adf77d05bacb7a4e09af54320b2874fa; causal=4e270f9b267af09a1e76b88b53bc195a900c186744d5a8a78b6419c8bc3435e1 | tests=2, pass=2, fail=0, skipped=0 | 0 | `4ece5ea753627cc980726f7e8527e2efbb39a9573f26ca4562d6ab0f3e5547b8` |
| default | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=af7871a298ebb183067ce90bcb2fabfb62909509ef9edaa4776731bcb9bf387d; causal=4e270f9b267af09a1e76b88b53bc195a900c186744d5a8a78b6419c8bc3435e1 | tests=2, pass=2, fail=0, skipped=0 | 0 | `35a1dbdeec3289f13a10b609150a20aa1f71dc4cb841db247a8fcae39af9eaf2` |

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
| bad | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=0, fail=1, skip=0; output SHA-256: outcome=e2abc15aadbdf8459be89ecbe4f508252ec75ad20a76651f260a7d8c672a4c7f; causal=6a856c1c1cd355122a3fe523b06699883f615f4506e2ecb453f5567033cf6521 | tests=2, pass=1, fail=1, skipped=0 | 1 | `ba96a50c4e55b2ca92364e2facea89b658e7e4d12573e2363478a9342c186671` |
| clean | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=f61c9e948f5f95a558e9d2929caff215adf77d05bacb7a4e09af54320b2874fa; causal=4e270f9b267af09a1e76b88b53bc195a900c186744d5a8a78b6419c8bc3435e1 | tests=2, pass=2, fail=0, skipped=0 | 0 | `cce3db30cc417ebee0d9ebf0293b11a3a88b171e8e600314c724cb967b58cdcb` |
| default | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=b79828435f73a24f4f356011e2df816ae09607697d58a3ec48706a87c6f12142; causal=5845a532a5b59e989931d972941a6c6e29587dfacfe3981604f4b128a6aa0fbe | tests=2, pass=2, fail=0, skipped=0 | 0 | `d29eb3d861d5bbd321b7e6e30dbb8e3e075d58c1598c0f70c8c2da5289cac946` |

Exact named root: `./internal/compiler/executionpeer -run '^TestPhase36CallCausalAttributionControls$' -count=1 -v -timeout=20s`. Expected violation assertion(s): `causal peer should accept the structurally plausible false-success trace; numeric replay is not its claim`. The bad subject changes only its hardcoded temporary Go overlay and reaches only these assertions; tracked source bytes are compared before/after and temp overlays are removed. Clean/default controls both pass without skips.
