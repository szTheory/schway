# Phase 37 prohibition evidence

Evidence date: 2026-10-15
Full checkout HEAD: `d6078b39e30c0d8ce5d4f011018a19e5785291e0`
Host/toolchain: darwin arm64; Node `v26.10.0`; Go cache `/tmp/ai-lang-verification-gocache`.
Adapter SHA-256: `287d5dd803c040fe8d42ad9161fa2dd635c6d820207c9f961502adfc7e751e88`.
Plan SHA-256 after descriptor wiring: `37-01-PLAN.md=ee401dac81b262544bbc3c94954cdb7362be57e1b58e40adb5450f15085f987f`; `37-02-PLAN.md=989bf14eebd815f2357953356aecd18e47d16799cbf5f4976aa276a6fee2bef4`.

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
| bad | outcome: run=1, pass=0, fail=1, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=847eccb59438e8eade069685302781935c843590367a4e7c9b694f81924056f8; causal=e09798f2777f9125f441c408b20c486f9919a0435266127d4dd4da34c4c03f7f; interpreterOutcome=24a78f05ab40ee3d68011c4e990d2b3f5222053436d9c127168e5926750a50df | tests=3, pass=2, fail=1, skipped=0 | 1 | `c807d1d5b6708095bf89e190930dbb596d0b489474a796e67ba877b067bbdb68` |
| clean | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=4c09bd19e9da808fc17f3c1c1c7f10fe38f6bdf9b3ecb949a6ba5ab13caeb38f; causal=4e270f9b267af09a1e76b88b53bc195a900c186744d5a8a78b6419c8bc3435e1; interpreterOutcome=e37184e9c5c0300a2945374ffef3176eb671918780bfc1f4c6feddef7a333f98 | tests=3, pass=3, fail=0, skipped=0 | 0 | `bafe7fe6f9773d7e684d8715263d95d98af87d830b5aad74d0ee3eb90efcbaf8` |
| default | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=e47455d511e429f1ef318b784c177b2d29cfc03fbf89fdcea13f1194210ae8aa; causal=0a0cfeaaa7eb93a612090418d80c49890ad0e30aca368a1b37050d09a91341a5; interpreterOutcome=e37184e9c5c0300a2945374ffef3176eb671918780bfc1f4c6feddef7a333f98 | tests=3, pass=3, fail=0, skipped=0 | 0 | `bf371fc4cce7b90600f01874d93388ff8272ff20a6d6204914098a44ace5309b` |

Exact named root: `./internal/compiler/native -run '^TestPhase36CaptureOutcomeSeparation$' -count=1 -v -timeout=20s`. Expected violation assertion(s): `capture exhaustion changed successful work`; `capture exhaustion hid the later defect`. The bad subject changes only its hardcoded temporary Go overlay and reaches only these assertions; tracked source bytes are compared before/after and temp overlays are removed. Clean/default controls both pass without skips.

## Causal peer numeric-limit prohibition

Statement: `MUST NOT present causal structural acceptance as independent CRC numeric verification.`
Source: `internal/compiler/executionpeer/executionpeer.go` SHA-256 `867185ab16461dd079451ead43ae47c68c3339ac6db7b790430f54954a02393c`; adapter/test SHA-256 `287d5dd803c040fe8d42ad9161fa2dd635c6d820207c9f961502adfc7e751e88`.
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
| bad | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=0, fail=1, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=98161dd75a0af15f2299c856513fc4683285af262b2489ab42251e1afae9b35e; causal=5702e31d154a10125d291a28005850914dfe0a6525b1bfd4164989071fa0ffda; interpreterOutcome=807cf474cd44683f94ae851ec412dd4ca72c13e04fccae71a1ca34454218378c | tests=3, pass=2, fail=1, skipped=0 | 1 | `84a04ab4f5faf55bce3d0cdd4385ee1c014b00cf079f6cdd9139a5db69e774e1` |
| clean | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=219b9f68174b7c7aad00ae63c5402dba08d7eee4b108bc4e336f49a700f742bf; causal=4e270f9b267af09a1e76b88b53bc195a900c186744d5a8a78b6419c8bc3435e1; interpreterOutcome=e37184e9c5c0300a2945374ffef3176eb671918780bfc1f4c6feddef7a333f98 | tests=3, pass=3, fail=0, skipped=0 | 0 | `c24ad464c1baa6a49f43b7269e299e96a03b8a8e910ebbb93efad5cb5ce9d958` |
| default | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=7deafc557fab7c58b674004068ba0fe1c7aa074745339b4e16076f711a2d6a55; causal=4e270f9b267af09a1e76b88b53bc195a900c186744d5a8a78b6419c8bc3435e1; interpreterOutcome=e37184e9c5c0300a2945374ffef3176eb671918780bfc1f4c6feddef7a333f98 | tests=3, pass=3, fail=0, skipped=0 | 0 | `f7e1166d249491d117c9be1a8d761aae7ca78c6904fd6657f85b5382a76b1333` |

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
| bad | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=0, fail=1, skip=0; output SHA-256: outcome=7deafc557fab7c58b674004068ba0fe1c7aa074745339b4e16076f711a2d6a55; causal=4e270f9b267af09a1e76b88b53bc195a900c186744d5a8a78b6419c8bc3435e1; interpreterOutcome=fbefeb5874e519552facd139a888abeaf124fb0b267e5efe9c57861d0a421a40 | tests=3, pass=2, fail=1, skipped=0 | 1 | `6e89cac882861a81c35c74da999b13d08e62b30f8a9554060ee7bb91ed2320e2` |
| clean | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=82b736902fc1f06fd2483e1664a988b85fc920064e8a492f05a2e599eba54078; causal=4e270f9b267af09a1e76b88b53bc195a900c186744d5a8a78b6419c8bc3435e1; interpreterOutcome=b15fd6132b5ac0a0626b1c10f13c3b8b287eaf4a5da549a186e82a72a6a3aebd | tests=3, pass=3, fail=0, skipped=0 | 0 | `b0da111d724f0bde62a7f56b0efa9f9dd04311b878501e68552cc92cea4bb8ea` |
| default | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=31ad7be1daa29907595ccedf4d7b877e45837fa7bfb90c48de30ea157516d50e; causal=b024dc0ad8137fecba43eb61c87a0d7fcc1a97c2968bcb726a83411f392cc2ba; interpreterOutcome=e37184e9c5c0300a2945374ffef3176eb671918780bfc1f4c6feddef7a333f98 | tests=3, pass=3, fail=0, skipped=0 | 0 | `7ac65d6bb9009759b9bae9ec9783133504ed05df2a8d623a48d11106f73f80e1` |

Exact named root: `./internal/compiler/interp -run '^TestPhase37CaptureDisabledOutcomes$' -count=1 -v -timeout=20s`. Expected violation assertion(s): `want late defect preserved after evidence exhaustion`. The bad subject changes only its hardcoded temporary Go overlay and reaches only these assertions; tracked source bytes are compared before/after and temp overlays are removed. Clean/default controls both pass without skips.
