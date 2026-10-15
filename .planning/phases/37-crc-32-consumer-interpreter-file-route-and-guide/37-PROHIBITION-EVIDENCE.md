# Phase 37 prohibition evidence

Evidence date: 2026-10-15
Full checkout HEAD: `8d0d7c6acb2ecefee5c1fc5643a70d5289405976`
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
| bad | outcome: run=1, pass=0, fail=1, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=a362cbb25a1f9b5092bca3057fab8cb90cc94bab8078ac9248876221019c9519; causal=88aa5c0c4189ed0ace81a8c07a7924b47aaef2cd8bd995b71bbf262ee9eaf9e8; interpreterOutcome=f478b30d6f4bc16c901f68da60e247de74b549b954eb3c711640bd9dc2f7384a | tests=3, pass=2, fail=1, skipped=0 | 1 | `a1463e3424fa97c9c63bb0b39b9b9386cb86a083467c82f503631e65234c78fa` |
| clean | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=6d3dc2f736b079ab9b5f76f0933151bcba0664ea9b9b39c586af07315866c9af; causal=c7025c9e88d88faa9cc0932a912496ac22485494c2896b33abe95c3d742a13c1; interpreterOutcome=a99918a46980a578aa1e00be8ac59fab21a9d9946070e91df187e0af9f0e0991 | tests=3, pass=3, fail=0, skipped=0 | 0 | `069a3c0148cd41f99f7aad88f21e55c180b363b8db028589cf35f0f27d4ffb3b` |
| default | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=9b75f372061864757987fa8afdd2c965d4330a3892149e7cb6cf869b0b023166; causal=4a2ddf56b50b7818eb2edc8ba1c4dceed97465dd04a540c0bef69ce2f2ce6b30; interpreterOutcome=de13bc7feec17fc31bd76199c35492d49e462595c236ecdf81a613ff7d67b171 | tests=3, pass=3, fail=0, skipped=0 | 0 | `7128b522a3838eab2473033104ede893606e5a2d9b5108236b35cd0f200b01f1` |

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
| bad | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=0, fail=1, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=f500ac8fe45e0034b50ba9a3ed53b6933fc2336a73fd3171023d2234f6d42d6c; causal=f28f149bfa0950561774fed20a9a5a70222efd9fe738d0603fb92a43eecef62c; interpreterOutcome=b89e541113360ff6e7454cad14e9bfcee1ba1e634f7d5ace4d8476951771f280 | tests=3, pass=2, fail=1, skipped=0 | 1 | `81c0b0ae9cacef04a9ae94fad4edb76784f5c8871af665b4bd7b87c28ea9c6bb` |
| clean | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=0acb36bf9f8dcfb8973956b65bffee2ffeeb8747ca44b052c8f3ef3ffa9bf142; causal=693838bad255ea5657ad86ae6c7edf2bd2780c973465e79790c47502cd0e0ebd; interpreterOutcome=b9780fcfa9447569d9c9f349ff2ee975b336ffbe2a571d72fa1f19c843b68958 | tests=3, pass=3, fail=0, skipped=0 | 0 | `db2cbef6266f48c1c228dbebd56946b546c58c73bdc19c15968d26c341d302b8` |
| default | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=61c3275546551af4e5a43be58da7c02acd5c73985eb06ce9900ba7a679d08f33; causal=9e00f5def695330154bd6aa4a213127ac7d97c5c9f12aa18110944e2088bdf29; interpreterOutcome=f3f5165487031d8bf41b9c6c048edfd18aeaf1384e8b4e5080ac41702793cf5a | tests=3, pass=3, fail=0, skipped=0 | 0 | `576480b87b4afb22278124faf00c7f5226d047f50a27191f0eba043becd25e53` |

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
| bad | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=0, fail=1, skip=0; output SHA-256: outcome=caf259a94e325292fc980ce132d73573cd4918b938e6dce2d0b53c40518a9489; causal=a6c41bc1c89a50e5b7e69b483c2d1a26fae655a4da11a738c5f64045287f7146; interpreterOutcome=7c015912ecff19683692de3ef631fd6994350064e29f756e681f947621a7ce46 | tests=3, pass=2, fail=1, skipped=0 | 1 | `ffc1c008fd987a5288a5e941cfcac2d78c71b7333747c5f9abcf9fba73da2c1c` |
| clean | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=0e4b8939b75db5dc1616c1392f7e240d2f259d36ac78485b046d2f8afdeb9e14; causal=693838bad255ea5657ad86ae6c7edf2bd2780c973465e79790c47502cd0e0ebd; interpreterOutcome=de13bc7feec17fc31bd76199c35492d49e462595c236ecdf81a613ff7d67b171 | tests=3, pass=3, fail=0, skipped=0 | 0 | `eedcf8fd701b42007ae588c4b604408f1f9b55204e6cbb1f95f8b19933cb3558` |
| default | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=cc493407065dacc1decabf140c36c8a210a3acf5e5f8c13fc95567d600fd0651; causal=9d62727cb17ad7235cb4f4dde1b56d1f8523949119ea195e5022c9c5ea59389c; interpreterOutcome=ef667990b95b21e278986ec83632fe706595cbb751d21842676075e6d05f9b11 | tests=3, pass=3, fail=0, skipped=0 | 0 | `4b0c27d0bc27be8eceb0a4a49cab906f42923fca867ca4b3ae40305956873587` |

Exact named root: `./internal/compiler/interp -run '^TestPhase37CaptureDisabledOutcomes$' -count=1 -v -timeout=20s`. Expected violation assertion(s): `want late defect preserved after evidence exhaustion`. The bad subject changes only its hardcoded temporary Go overlay and reaches only these assertions; tracked source bytes are compared before/after and temp overlays are removed. Clean/default controls both pass without skips.
