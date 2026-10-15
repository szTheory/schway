# Phase 37 prohibition evidence

Evidence date: 2026-10-15
Full checkout HEAD: `b900dfe31abb600a82b31ab6e8cc2a8b9b1e74ca`
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
| bad | outcome: run=1, pass=0, fail=1, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=29c04c33b194f09a2c4cb111f602c6f324096d8f04db5389c23ba52077b51cd6; causal=0a0cfeaaa7eb93a612090418d80c49890ad0e30aca368a1b37050d09a91341a5; interpreterOutcome=24a78f05ab40ee3d68011c4e990d2b3f5222053436d9c127168e5926750a50df | tests=3, pass=2, fail=1, skipped=0 | 1 | `9b211cde444d4143753280cdf62d1f487c422bc1f81d766ac8889036007e4cbd` |
| clean | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=dd2403938430ca4b4b3006a001517d99507662c2b2221b6cef1a6714e9628305; causal=e09798f2777f9125f441c408b20c486f9919a0435266127d4dd4da34c4c03f7f; interpreterOutcome=24a78f05ab40ee3d68011c4e990d2b3f5222053436d9c127168e5926750a50df | tests=3, pass=3, fail=0, skipped=0 | 0 | `7327e3c3cd6ab5c0b4c4c94ddaabcc009992536bf19360cf0e16fb67ff77f35f` |
| default | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=b7033901199c134cf6aef04d9b2ec79d2baae72feed7e75c1f3ae8c5afbdb5db; causal=e09798f2777f9125f441c408b20c486f9919a0435266127d4dd4da34c4c03f7f; interpreterOutcome=24a78f05ab40ee3d68011c4e990d2b3f5222053436d9c127168e5926750a50df | tests=3, pass=3, fail=0, skipped=0 | 0 | `8b238d7d0e3717938df92a17b294f1020b21d7465983606d0898240e3e48a6a5` |

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
| bad | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=0, fail=1, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=66bc76b84392d4f49aa3c761969fbc7f6f7396e1039887b2cdde14f7abf078ca; causal=4684898a5fd1a524f37c926b5624826ea051c59cfeaa8501e22a02d10be7cdc8; interpreterOutcome=807cf474cd44683f94ae851ec412dd4ca72c13e04fccae71a1ca34454218378c | tests=3, pass=2, fail=1, skipped=0 | 1 | `f91233685fb21edda43be1a168074a388bbfca7fc34fd03b4d4af05fac7ca9ca` |
| clean | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=e9b9177675d9c26907c96558a0aa580fe59c760bcfd9839a18d61ca739f87ccf; causal=57329fac795102aa9a8fc9eee870f2825ee296e384a9cf3bc05efca5f161dd84; interpreterOutcome=24a78f05ab40ee3d68011c4e990d2b3f5222053436d9c127168e5926750a50df | tests=3, pass=3, fail=0, skipped=0 | 0 | `368a1880f5b2bcd15351ee3c7793218ac5bf5aa1f5aff5d761b72bb652a3b4cd` |
| default | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=3895cd0b5f7e6dd53a506820e0aa0a7f9f7f4d440bc0840e9e37f08311d021ab; causal=57329fac795102aa9a8fc9eee870f2825ee296e384a9cf3bc05efca5f161dd84; interpreterOutcome=e37184e9c5c0300a2945374ffef3176eb671918780bfc1f4c6feddef7a333f98 | tests=3, pass=3, fail=0, skipped=0 | 0 | `a1e4f3394571049ca19c2b0e384ff31b09f7186a4133ad067f3bedc7bdc1e21d` |

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
| bad | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=0, fail=1, skip=0; output SHA-256: outcome=2ba6fff8541cc8de54ab9a25ba11aebc86a39e4904226606e1524ea2f32af830; causal=e09798f2777f9125f441c408b20c486f9919a0435266127d4dd4da34c4c03f7f; interpreterOutcome=a595c42f5bf4f9da7a492be3f0287899d6a328d39bc8b9d50869fe359d52c2b9 | tests=3, pass=2, fail=1, skipped=0 | 1 | `760f5ce2120186f1f86d1b1931a538a0c79a55dd3e9b4e993c9ed41f8d28aeab` |
| clean | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=3895cd0b5f7e6dd53a506820e0aa0a7f9f7f4d440bc0840e9e37f08311d021ab; causal=e09798f2777f9125f441c408b20c486f9919a0435266127d4dd4da34c4c03f7f; interpreterOutcome=807cf474cd44683f94ae851ec412dd4ca72c13e04fccae71a1ca34454218378c | tests=3, pass=3, fail=0, skipped=0 | 0 | `17928039030c7b120ccfd6ed9bef280411cb774d3b60b5eae6dca9df4e7a0ca6` |
| default | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=3895cd0b5f7e6dd53a506820e0aa0a7f9f7f4d440bc0840e9e37f08311d021ab; causal=e09798f2777f9125f441c408b20c486f9919a0435266127d4dd4da34c4c03f7f; interpreterOutcome=24a78f05ab40ee3d68011c4e990d2b3f5222053436d9c127168e5926750a50df | tests=3, pass=3, fail=0, skipped=0 | 0 | `9b019c95d6bda1cad6c01e42c49b709669fc2875dacfae9d1d6bcfba52188b52` |

Exact named root: `./internal/compiler/interp -run '^TestPhase37CaptureDisabledOutcomes$' -count=1 -v -timeout=20s`. Expected violation assertion(s): `want late defect preserved after evidence exhaustion`. The bad subject changes only its hardcoded temporary Go overlay and reaches only these assertions; tracked source bytes are compared before/after and temp overlays are removed. Clean/default controls both pass without skips.
