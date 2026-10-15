# Phase 37 prohibition evidence

Evidence date: 2026-10-15
Full checkout HEAD: `3868c8f85d346e3387f1d61425169b3ac45b48ec`
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
| bad | outcome: run=1, pass=0, fail=1, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=ed976b4713ee178347da523767739af4d5a064706931067a8e32a3095ea23f1d; causal=b024dc0ad8137fecba43eb61c87a0d7fcc1a97c2968bcb726a83411f392cc2ba; interpreterOutcome=807cf474cd44683f94ae851ec412dd4ca72c13e04fccae71a1ca34454218378c | tests=3, pass=2, fail=1, skipped=0 | 1 | `e163dca57ed15c70a37d6d85f7520e2f69bb06ba37397f1b33a93238a87bcc88` |
| clean | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=7a05c53c85df411d97b23fbb07efdbc4d13c0e328948a9b075aa033c9f9a9fa1; causal=4e270f9b267af09a1e76b88b53bc195a900c186744d5a8a78b6419c8bc3435e1; interpreterOutcome=e37184e9c5c0300a2945374ffef3176eb671918780bfc1f4c6feddef7a333f98 | tests=3, pass=3, fail=0, skipped=0 | 0 | `15c3344ee34607d8b8db339798574f081a8e2cb67621340412b420eed47c18a9` |
| default | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=66bc76b84392d4f49aa3c761969fbc7f6f7396e1039887b2cdde14f7abf078ca; causal=0a0cfeaaa7eb93a612090418d80c49890ad0e30aca368a1b37050d09a91341a5; interpreterOutcome=24a78f05ab40ee3d68011c4e990d2b3f5222053436d9c127168e5926750a50df | tests=3, pass=3, fail=0, skipped=0 | 0 | `a49817433f2211c58c6337302e7847886df7ae104f876a25a76c0b905f8a28a2` |

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
| bad | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=0, fail=1, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=e8864ca12d70a053ff25b22ebe67da1177529e20613a8d6af5b941b021da9c88; causal=6a856c1c1cd355122a3fe523b06699883f615f4506e2ecb453f5567033cf6521; interpreterOutcome=e37184e9c5c0300a2945374ffef3176eb671918780bfc1f4c6feddef7a333f98 | tests=3, pass=2, fail=1, skipped=0 | 1 | `3207d695ec5e01c9cdc0d8036673b59250053321767168cc0539dcd8a9a60009` |
| clean | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=fa8bfc9e01addb359fbb8e4ceff433e42a2fa641145965b9b848a9096b5330be; causal=b024dc0ad8137fecba43eb61c87a0d7fcc1a97c2968bcb726a83411f392cc2ba; interpreterOutcome=807cf474cd44683f94ae851ec412dd4ca72c13e04fccae71a1ca34454218378c | tests=3, pass=3, fail=0, skipped=0 | 0 | `07484ec09d21e4f3f9980d801dcccfdc594f46268cd012b094a849101f6137b6` |
| default | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=e561d56a80ff6e1fd9706a40ee128796b7722200efee27ea8bd8e212cb35b42b; causal=0a0cfeaaa7eb93a612090418d80c49890ad0e30aca368a1b37050d09a91341a5; interpreterOutcome=24a78f05ab40ee3d68011c4e990d2b3f5222053436d9c127168e5926750a50df | tests=3, pass=3, fail=0, skipped=0 | 0 | `f990cdc86e72bfc3b773ecbc6e4dba32d86d1d6eee131bfe9d3147ed7653a6b8` |

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
| bad | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=0, fail=1, skip=0; output SHA-256: outcome=30a8f6bfb89c396cc49cc222439bb53d6ad78f6315dbd3e02e92bbc8f2ff13fa; causal=b024dc0ad8137fecba43eb61c87a0d7fcc1a97c2968bcb726a83411f392cc2ba; interpreterOutcome=fbefeb5874e519552facd139a888abeaf124fb0b267e5efe9c57861d0a421a40 | tests=3, pass=2, fail=1, skipped=0 | 1 | `1df9bb5703082f6e4f590d8987c7ea097fc055db067ced6214dfa30aef84f25f` |
| clean | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=be2fd4395299f075093e8cf8e03fcd5a5f244cb2e9ad36db88839ea30facac3b; causal=b024dc0ad8137fecba43eb61c87a0d7fcc1a97c2968bcb726a83411f392cc2ba; interpreterOutcome=807cf474cd44683f94ae851ec412dd4ca72c13e04fccae71a1ca34454218378c | tests=3, pass=3, fail=0, skipped=0 | 0 | `0768453663bb633c71a953596bc8cd603785119178d2e1aac6aa5333447af487` |
| default | outcome: run=1, pass=1, fail=0, skip=0; causal: run=1, pass=1, fail=0, skip=0; interpreterOutcome: run=1, pass=1, fail=0, skip=0; output SHA-256: outcome=7deafc557fab7c58b674004068ba0fe1c7aa074745339b4e16076f711a2d6a55; causal=e09798f2777f9125f441c408b20c486f9919a0435266127d4dd4da34c4c03f7f; interpreterOutcome=e37184e9c5c0300a2945374ffef3176eb671918780bfc1f4c6feddef7a333f98 | tests=3, pass=3, fail=0, skipped=0 | 0 | `3d8f6bef7ea1bc852f7bb254a01aeac73a17e9a74a5aad7a4cb437ad61b15109` |

Exact named root: `./internal/compiler/interp -run '^TestPhase37CaptureDisabledOutcomes$' -count=1 -v -timeout=20s`. Expected violation assertion(s): `want late defect preserved after evidence exhaustion`. The bad subject changes only its hardcoded temporary Go overlay and reaches only these assertions; tracked source bytes are compared before/after and temp overlays are removed. Clean/default controls both pass without skips.
