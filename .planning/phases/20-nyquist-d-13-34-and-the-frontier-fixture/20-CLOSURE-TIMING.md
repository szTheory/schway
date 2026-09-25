# Phase 20 closure timing evidence

- **Status:** PASS
- **Started:** 2026-09-25T08:04:08Z
- **Completed:** 2026-09-25T08:45:43Z
- **Elapsed:** 1507.346 s
- **Machine:** `machine:4797d76b7863` (`darwin`, `arm64`, 18 cores)
- **Go:** `go1.24.0`
- **Clang:** `Apple clang version 21.0.0 (clang-2100.1.1.101)`
- **Input HEAD at start:** `d8421776b25e2aee022ce1a40f360149d9f3a958`
- **Source fingerprint:** `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`
- **Go build cache:** `/private/tmp/phase20-gocache` (stable across samples)
- **GOMAXPROCS:** `4` for preflight and all timed commands
- **CPU model provenance:** `Apple M5 Pro`, verified by a direct host `sysctl` query. The sandbox denied that query, so a temporary shim supplied the verified value during collection; the emitted machine ID exactly matches the Phase 14 baseline.
- **Protocol cap:** 30 minutes; monotonic `time.Since` measurements

Three paired samples expose spread; they do not establish a high-confidence percentile. Full-suite commands use `go test ./... -count=1`; each pair has an empty cold closure cache and reuses that exact populated cache for warm. The Go build cache was prewarmed successfully in **3.355 s** (exit 0) and remained stable.

## Full-suite preflight

- **Command:** `go test ./... -count=1`
- **Revision:** `d842177`
- **Environment:** Go 1.24.0, Apple Clang 21.0.0, darwin/arm64, `GOMAXPROCS=4`
- **Exit:** 0
- **Wall time:** 179.035 s

The measurement used the same verified machine identity as Phase 14. The CPU model shim returned the directly observed host value only for the sandbox-blocked `sysctl -n machdep.cpu.brand_string` probe; all compiler, test, cache, and timing commands ran normally.

| Sample | Command | Cache | Raw monotonic nanoseconds (seconds) | Exit | Timeout |
|---|---|---|---:|---:|---|
| full-1-cold | `go test ./... -count=1` | `~/T/phase20-full-pair-1-3844346473` | 152363262458 (152.363) | 0 | false |
| full-1-warm | `go test ./... -count=1` | `~/T/phase20-full-pair-1-3844346473` | 120373627542 (120.374) | 0 | false |
| full-2-cold | `go test ./... -count=1` | `~/T/phase20-full-pair-2-1509477539` | 129425412666 (129.425) | 0 | false |
| full-2-warm | `go test ./... -count=1` | `~/T/phase20-full-pair-2-1509477539` | 139388643708 (139.389) | 0 | false |
| full-3-cold | `go test ./... -count=1` | `~/T/phase20-full-pair-3-1742789899` | 466304738667 (466.305) | 0 | false |
| full-3-warm | `go test ./... -count=1` | `~/T/phase20-full-pair-3-1742789899` | 381853587750 (381.854) | 0 | false |
| closure-cold | `go test ./internal/compiler/session -run ^TestPhase5CorpusThreeEngineAgreement/enumerated-closure$ -count=1 -v` | `~/T/phase20-closure-pair-1204560473` | 65623183917 (65.623) | 0 | false |
| closure-warm | `go test ./internal/compiler/session -run ^TestPhase5CorpusThreeEngineAgreement/enumerated-closure$ -count=1 -v` | `~/T/phase20-closure-pair-1204560473` | 48048590833 (48.049) | 0 | false |

- **Cold ordered seconds:** 129.425, 152.363, 466.305; min/median/max = 129.425 / 152.363 / 466.305
- **Warm ordered seconds:** 120.374, 139.389, 381.854; min/median/max = 120.374 / 139.389 / 381.854
- **Per-pair cold minus warm seconds:** 31.990 -9.963 84.451

## Enumerated closure cold/warm

- **closure-cold:** 65.623 s; exit=0; reused=0 recomputed=112 not-cacheable=0 unavailable=0. Command: `go test ./internal/compiler/session -run ^TestPhase5CorpusThreeEngineAgreement/enumerated-closure$ -count=1 -v`.
- **closure-warm:** 48.049 s; exit=0; reused=112 recomputed=0 not-cacheable=0 unavailable=0. Command: `go test ./internal/compiler/session -run ^TestPhase5CorpusThreeEngineAgreement/enumerated-closure$ -count=1 -v`.

## Phase 14 comparison

- Roadmap reference: **192.7 s**.
- QLT-02 measured reference: **191.89 s**.
- Warm median: **139.389 s**, below both references and the paired cold median.

## Raw process output

### full-1-cold

```text
ok  	github.com/codename-lang/lang/cmd/lang	0.172s
ok  	github.com/codename-lang/lang/cmd/lang-repair	8.995s
ok  	github.com/codename-lang/lang/internal/compiler/ability	0.421s
?   	github.com/codename-lang/lang/internal/compiler/ast	[no test files]
ok  	github.com/codename-lang/lang/internal/compiler/cache	3.442s
ok  	github.com/codename-lang/lang/internal/compiler/callgraph	0.447s
ok  	github.com/codename-lang/lang/internal/compiler/cgen	10.110s
ok  	github.com/codename-lang/lang/internal/compiler/check	6.121s
ok  	github.com/codename-lang/lang/internal/compiler/core	0.653s
ok  	github.com/codename-lang/lang/internal/compiler/corevalidate	1.932s
ok  	github.com/codename-lang/lang/internal/compiler/debugmap	0.207s
ok  	github.com/codename-lang/lang/internal/compiler/diagnostic	0.325s
ok  	github.com/codename-lang/lang/internal/compiler/evidence	1.401s
ok  	github.com/codename-lang/lang/internal/compiler/execution	0.256s
ok  	github.com/codename-lang/lang/internal/compiler/executionpeer	1.481s
ok  	github.com/codename-lang/lang/internal/compiler/interp	0.675s
?   	github.com/codename-lang/lang/internal/compiler/interp/interptestdirect	[no test files]
ok  	github.com/codename-lang/lang/internal/compiler/measure	8.502s
ok  	github.com/codename-lang/lang/internal/compiler/native	23.330s
ok  	github.com/codename-lang/lang/internal/compiler/originvalidate	0.414s
ok  	github.com/codename-lang/lang/internal/compiler/pathoracle	0.346s
ok  	github.com/codename-lang/lang/internal/compiler/protocol	0.352s
ok  	github.com/codename-lang/lang/internal/compiler/reduce	0.702s
ok  	github.com/codename-lang/lang/internal/compiler/session	141.343s
ok  	github.com/codename-lang/lang/internal/compiler/syntax	0.373s
ok  	github.com/codename-lang/lang/internal/compiler/testsupport	46.799s
ok  	github.com/codename-lang/lang/scripts	0.611s
```

### full-1-warm

```text
ok  	github.com/codename-lang/lang/cmd/lang	0.163s
ok  	github.com/codename-lang/lang/cmd/lang-repair	8.711s
ok  	github.com/codename-lang/lang/internal/compiler/ability	0.404s
?   	github.com/codename-lang/lang/internal/compiler/ast	[no test files]
ok  	github.com/codename-lang/lang/internal/compiler/cache	3.588s
ok  	github.com/codename-lang/lang/internal/compiler/callgraph	0.443s
ok  	github.com/codename-lang/lang/internal/compiler/cgen	9.639s
ok  	github.com/codename-lang/lang/internal/compiler/check	6.089s
ok  	github.com/codename-lang/lang/internal/compiler/core	0.640s
ok  	github.com/codename-lang/lang/internal/compiler/corevalidate	1.895s
ok  	github.com/codename-lang/lang/internal/compiler/debugmap	0.238s
ok  	github.com/codename-lang/lang/internal/compiler/diagnostic	0.173s
ok  	github.com/codename-lang/lang/internal/compiler/evidence	1.595s
ok  	github.com/codename-lang/lang/internal/compiler/execution	0.439s
ok  	github.com/codename-lang/lang/internal/compiler/executionpeer	1.452s
ok  	github.com/codename-lang/lang/internal/compiler/interp	0.805s
?   	github.com/codename-lang/lang/internal/compiler/interp/interptestdirect	[no test files]
ok  	github.com/codename-lang/lang/internal/compiler/measure	7.930s
ok  	github.com/codename-lang/lang/internal/compiler/native	14.254s
ok  	github.com/codename-lang/lang/internal/compiler/originvalidate	0.483s
ok  	github.com/codename-lang/lang/internal/compiler/pathoracle	0.266s
ok  	github.com/codename-lang/lang/internal/compiler/protocol	0.292s
ok  	github.com/codename-lang/lang/internal/compiler/reduce	0.689s
ok  	github.com/codename-lang/lang/internal/compiler/session	109.492s
ok  	github.com/codename-lang/lang/internal/compiler/syntax	0.329s
ok  	github.com/codename-lang/lang/internal/compiler/testsupport	26.949s
ok  	github.com/codename-lang/lang/scripts	0.330s
```

### full-2-cold

```text
ok  	github.com/codename-lang/lang/cmd/lang	0.158s
ok  	github.com/codename-lang/lang/cmd/lang-repair	8.629s
ok  	github.com/codename-lang/lang/internal/compiler/ability	0.402s
?   	github.com/codename-lang/lang/internal/compiler/ast	[no test files]
ok  	github.com/codename-lang/lang/internal/compiler/cache	3.471s
ok  	github.com/codename-lang/lang/internal/compiler/callgraph	0.442s
ok  	github.com/codename-lang/lang/internal/compiler/cgen	9.507s
ok  	github.com/codename-lang/lang/internal/compiler/check	6.093s
ok  	github.com/codename-lang/lang/internal/compiler/core	0.644s
ok  	github.com/codename-lang/lang/internal/compiler/corevalidate	1.904s
ok  	github.com/codename-lang/lang/internal/compiler/debugmap	0.206s
ok  	github.com/codename-lang/lang/internal/compiler/diagnostic	0.230s
ok  	github.com/codename-lang/lang/internal/compiler/evidence	1.522s
ok  	github.com/codename-lang/lang/internal/compiler/execution	0.396s
ok  	github.com/codename-lang/lang/internal/compiler/executionpeer	1.480s
ok  	github.com/codename-lang/lang/internal/compiler/interp	0.789s
?   	github.com/codename-lang/lang/internal/compiler/interp/interptestdirect	[no test files]
ok  	github.com/codename-lang/lang/internal/compiler/measure	7.813s
ok  	github.com/codename-lang/lang/internal/compiler/native	14.248s
ok  	github.com/codename-lang/lang/internal/compiler/originvalidate	0.433s
ok  	github.com/codename-lang/lang/internal/compiler/pathoracle	0.263s
ok  	github.com/codename-lang/lang/internal/compiler/protocol	0.281s
ok  	github.com/codename-lang/lang/internal/compiler/reduce	0.719s
ok  	github.com/codename-lang/lang/internal/compiler/session	118.683s
ok  	github.com/codename-lang/lang/internal/compiler/syntax	0.276s
ok  	github.com/codename-lang/lang/internal/compiler/testsupport	26.800s
ok  	github.com/codename-lang/lang/scripts	0.338s
```

### full-2-warm

```text
ok  	github.com/codename-lang/lang/cmd/lang	0.185s
ok  	github.com/codename-lang/lang/cmd/lang-repair	12.789s
ok  	github.com/codename-lang/lang/internal/compiler/ability	0.487s
?   	github.com/codename-lang/lang/internal/compiler/ast	[no test files]
ok  	github.com/codename-lang/lang/internal/compiler/cache	4.287s
ok  	github.com/codename-lang/lang/internal/compiler/callgraph	0.553s
ok  	github.com/codename-lang/lang/internal/compiler/cgen	13.533s
ok  	github.com/codename-lang/lang/internal/compiler/check	10.178s
ok  	github.com/codename-lang/lang/internal/compiler/core	0.975s
ok  	github.com/codename-lang/lang/internal/compiler/corevalidate	3.473s
ok  	github.com/codename-lang/lang/internal/compiler/debugmap	0.539s
ok  	github.com/codename-lang/lang/internal/compiler/diagnostic	0.275s
ok  	github.com/codename-lang/lang/internal/compiler/evidence	2.415s
ok  	github.com/codename-lang/lang/internal/compiler/execution	0.286s
ok  	github.com/codename-lang/lang/internal/compiler/executionpeer	2.309s
ok  	github.com/codename-lang/lang/internal/compiler/interp	0.954s
?   	github.com/codename-lang/lang/internal/compiler/interp/interptestdirect	[no test files]
ok  	github.com/codename-lang/lang/internal/compiler/measure	8.016s
ok  	github.com/codename-lang/lang/internal/compiler/native	19.940s
ok  	github.com/codename-lang/lang/internal/compiler/originvalidate	0.373s
ok  	github.com/codename-lang/lang/internal/compiler/pathoracle	0.282s
ok  	github.com/codename-lang/lang/internal/compiler/protocol	0.414s
ok  	github.com/codename-lang/lang/internal/compiler/reduce	0.866s
ok  	github.com/codename-lang/lang/internal/compiler/session	124.036s
ok  	github.com/codename-lang/lang/internal/compiler/syntax	0.431s
ok  	github.com/codename-lang/lang/internal/compiler/testsupport	36.298s
ok  	github.com/codename-lang/lang/scripts	0.528s
```

### full-3-cold

```text
ok  	github.com/codename-lang/lang/cmd/lang	0.161s
ok  	github.com/codename-lang/lang/cmd/lang-repair	10.583s
ok  	github.com/codename-lang/lang/internal/compiler/ability	0.412s
?   	github.com/codename-lang/lang/internal/compiler/ast	[no test files]
ok  	github.com/codename-lang/lang/internal/compiler/cache	3.484s
ok  	github.com/codename-lang/lang/internal/compiler/callgraph	0.449s
ok  	github.com/codename-lang/lang/internal/compiler/cgen	12.645s
ok  	github.com/codename-lang/lang/internal/compiler/check	7.663s
ok  	github.com/codename-lang/lang/internal/compiler/core	0.527s
ok  	github.com/codename-lang/lang/internal/compiler/corevalidate	2.762s
ok  	github.com/codename-lang/lang/internal/compiler/debugmap	0.273s
ok  	github.com/codename-lang/lang/internal/compiler/diagnostic	0.502s
ok  	github.com/codename-lang/lang/internal/compiler/evidence	2.306s
ok  	github.com/codename-lang/lang/internal/compiler/execution	0.518s
ok  	github.com/codename-lang/lang/internal/compiler/executionpeer	2.866s
ok  	github.com/codename-lang/lang/internal/compiler/interp	1.121s
?   	github.com/codename-lang/lang/internal/compiler/interp/interptestdirect	[no test files]
ok  	github.com/codename-lang/lang/internal/compiler/measure	8.000s
ok  	github.com/codename-lang/lang/internal/compiler/native	18.804s
ok  	github.com/codename-lang/lang/internal/compiler/originvalidate	0.543s
ok  	github.com/codename-lang/lang/internal/compiler/pathoracle	0.380s
ok  	github.com/codename-lang/lang/internal/compiler/protocol	0.413s
ok  	github.com/codename-lang/lang/internal/compiler/reduce	0.993s
ok  	github.com/codename-lang/lang/internal/compiler/session	452.125s
ok  	github.com/codename-lang/lang/internal/compiler/syntax	0.438s
ok  	github.com/codename-lang/lang/internal/compiler/testsupport	39.189s
ok  	github.com/codename-lang/lang/scripts	0.434s
```

### full-3-warm

```text
ok  	github.com/codename-lang/lang/cmd/lang	0.672s
ok  	github.com/codename-lang/lang/cmd/lang-repair	40.576s
ok  	github.com/codename-lang/lang/internal/compiler/ability	1.887s
?   	github.com/codename-lang/lang/internal/compiler/ast	[no test files]
ok  	github.com/codename-lang/lang/internal/compiler/cache	15.263s
ok  	github.com/codename-lang/lang/internal/compiler/callgraph	2.270s
ok  	github.com/codename-lang/lang/internal/compiler/cgen	42.754s
ok  	github.com/codename-lang/lang/internal/compiler/check	40.340s
ok  	github.com/codename-lang/lang/internal/compiler/core	3.838s
ok  	github.com/codename-lang/lang/internal/compiler/corevalidate	12.139s
ok  	github.com/codename-lang/lang/internal/compiler/debugmap	1.140s
ok  	github.com/codename-lang/lang/internal/compiler/diagnostic	1.433s
ok  	github.com/codename-lang/lang/internal/compiler/evidence	5.661s
ok  	github.com/codename-lang/lang/internal/compiler/execution	1.458s
ok  	github.com/codename-lang/lang/internal/compiler/executionpeer	8.164s
ok  	github.com/codename-lang/lang/internal/compiler/interp	4.025s
?   	github.com/codename-lang/lang/internal/compiler/interp/interptestdirect	[no test files]
ok  	github.com/codename-lang/lang/internal/compiler/measure	10.386s
ok  	github.com/codename-lang/lang/internal/compiler/native	63.811s
ok  	github.com/codename-lang/lang/internal/compiler/originvalidate	1.106s
ok  	github.com/codename-lang/lang/internal/compiler/pathoracle	1.161s
ok  	github.com/codename-lang/lang/internal/compiler/protocol	1.258s
ok  	github.com/codename-lang/lang/internal/compiler/reduce	4.095s
ok  	github.com/codename-lang/lang/internal/compiler/session	327.929s
ok  	github.com/codename-lang/lang/internal/compiler/syntax	1.537s
ok  	github.com/codename-lang/lang/internal/compiler/testsupport	143.316s
ok  	github.com/codename-lang/lang/scripts	1.047s
```

### closure-cold

```text
=== RUN   TestPhase5CorpusThreeEngineAgreement
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[0]:phase5.enum_chain_byte_l0_e0_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[1]:phase5.enum_chain_byte_l0_e0_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[2]:phase5.enum_chain_byte_l1_e0_tfalse
    session_phase5_corpus_test.go:608: enumerated[2]:phase5.enum_chain_byte_l1_e0_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l1_e0_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[3]:phase5.enum_chain_byte_l1_e0_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[4]:phase5.enum_chain_byte_l1_e1_tfalse
    session_phase5_corpus_test.go:608: enumerated[4]:phase5.enum_chain_byte_l1_e1_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l1_e1_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[5]:phase5.enum_chain_byte_l1_e2_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[6]:phase5.enum_chain_byte_l2_e0_tfalse
    session_phase5_corpus_test.go:608: enumerated[6]:phase5.enum_chain_byte_l2_e0_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l2_e0_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[7]:phase5.enum_chain_byte_l2_e0_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[8]:phase5.enum_chain_byte_l2_e1_tfalse
    session_phase5_corpus_test.go:608: enumerated[8]:phase5.enum_chain_byte_l2_e1_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l2_e1_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[9]:phase5.enum_chain_byte_l2_e2_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[10]:phase5.enum_chain_byte_l2_e3_tfalse
    session_phase5_corpus_test.go:608: enumerated[10]:phase5.enum_chain_byte_l2_e3_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l2_e3_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[11]:phase5.enum_chain_byte_l2_e3_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[12]:phase5.enum_chain_byte_l2_e4_tfalse
    session_phase5_corpus_test.go:608: enumerated[12]:phase5.enum_chain_byte_l2_e4_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l2_e4_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[13]:phase5.enum_chain_byte_l2_e5_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[14]:phase5.enum_chain_byte_l2_e6_tfalse
    session_phase5_corpus_test.go:608: enumerated[14]:phase5.enum_chain_byte_l2_e6_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l2_e6_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[15]:phase5.enum_chain_byte_l2_e6_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[16]:phase5.enum_chain_byte_l2_e7_tfalse
    session_phase5_corpus_test.go:608: enumerated[16]:phase5.enum_chain_byte_l2_e7_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l2_e7_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[17]:phase5.enum_chain_byte_l2_e8_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[18]:phase5.enum_chain_byte_l3_e0_tfalse
    session_phase5_corpus_test.go:608: enumerated[18]:phase5.enum_chain_byte_l3_e0_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e0_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[19]:phase5.enum_chain_byte_l3_e0_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[20]:phase5.enum_chain_byte_l3_e1_tfalse
    session_phase5_corpus_test.go:608: enumerated[20]:phase5.enum_chain_byte_l3_e1_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e1_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[21]:phase5.enum_chain_byte_l3_e2_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[22]:phase5.enum_chain_byte_l3_e3_tfalse
    session_phase5_corpus_test.go:608: enumerated[22]:phase5.enum_chain_byte_l3_e3_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e3_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[23]:phase5.enum_chain_byte_l3_e3_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[24]:phase5.enum_chain_byte_l3_e4_tfalse
    session_phase5_corpus_test.go:608: enumerated[24]:phase5.enum_chain_byte_l3_e4_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e4_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[25]:phase5.enum_chain_byte_l3_e5_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[26]:phase5.enum_chain_byte_l3_e6_tfalse
    session_phase5_corpus_test.go:608: enumerated[26]:phase5.enum_chain_byte_l3_e6_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e6_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[27]:phase5.enum_chain_byte_l3_e6_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[28]:phase5.enum_chain_byte_l3_e7_tfalse
    session_phase5_corpus_test.go:608: enumerated[28]:phase5.enum_chain_byte_l3_e7_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e7_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[29]:phase5.enum_chain_byte_l3_e8_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[30]:phase5.enum_chain_byte_l3_e9_tfalse
    session_phase5_corpus_test.go:608: enumerated[30]:phase5.enum_chain_byte_l3_e9_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e9_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[31]:phase5.enum_chain_byte_l3_e9_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[32]:phase5.enum_chain_byte_l3_e10_tfalse
    session_phase5_corpus_test.go:608: enumerated[32]:phase5.enum_chain_byte_l3_e10_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e10_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[33]:phase5.enum_chain_byte_l3_e11_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[34]:phase5.enum_chain_byte_l3_e12_tfalse
    session_phase5_corpus_test.go:608: enumerated[34]:phase5.enum_chain_byte_l3_e12_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e12_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[35]:phase5.enum_chain_byte_l3_e12_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[36]:phase5.enum_chain_byte_l3_e13_tfalse
    session_phase5_corpus_test.go:608: enumerated[36]:phase5.enum_chain_byte_l3_e13_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e13_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[37]:phase5.enum_chain_byte_l3_e14_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[38]:phase5.enum_chain_byte_l3_e15_tfalse
    session_phase5_corpus_test.go:608: enumerated[38]:phase5.enum_chain_byte_l3_e15_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e15_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[39]:phase5.enum_chain_byte_l3_e15_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[40]:phase5.enum_chain_byte_l3_e16_tfalse
    session_phase5_corpus_test.go:608: enumerated[40]:phase5.enum_chain_byte_l3_e16_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e16_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[41]:phase5.enum_chain_byte_l3_e17_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[42]:phase5.enum_chain_byte_l3_e18_tfalse
    session_phase5_corpus_test.go:608: enumerated[42]:phase5.enum_chain_byte_l3_e18_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e18_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[43]:phase5.enum_chain_byte_l3_e18_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[44]:phase5.enum_chain_byte_l3_e19_tfalse
    session_phase5_corpus_test.go:608: enumerated[44]:phase5.enum_chain_byte_l3_e19_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e19_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[45]:phase5.enum_chain_byte_l3_e20_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[46]:phase5.enum_chain_byte_l3_e21_tfalse
    session_phase5_corpus_test.go:608: enumerated[46]:phase5.enum_chain_byte_l3_e21_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e21_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[47]:phase5.enum_chain_byte_l3_e21_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[48]:phase5.enum_chain_byte_l3_e22_tfalse
    session_phase5_corpus_test.go:608: enumerated[48]:phase5.enum_chain_byte_l3_e22_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e22_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[49]:phase5.enum_chain_byte_l3_e23_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[50]:phase5.enum_chain_byte_l3_e24_tfalse
    session_phase5_corpus_test.go:608: enumerated[50]:phase5.enum_chain_byte_l3_e24_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e24_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[51]:phase5.enum_chain_byte_l3_e24_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[52]:phase5.enum_chain_byte_l3_e25_tfalse
    session_phase5_corpus_test.go:608: enumerated[52]:phase5.enum_chain_byte_l3_e25_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e25_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[53]:phase5.enum_chain_byte_l3_e26_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[54]:phase5.enum_chain_buffer_l0_e0_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[55]:phase5.enum_chain_buffer_l0_e0_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[56]:phase5.enum_chain_buffer_l1_e0_tfalse
    session_phase5_corpus_test.go:608: enumerated[56]:phase5.enum_chain_buffer_l1_e0_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l1_e0_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[57]:phase5.enum_chain_buffer_l1_e0_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[58]:phase5.enum_chain_buffer_l1_e1_tfalse
    session_phase5_corpus_test.go:608: enumerated[58]:phase5.enum_chain_buffer_l1_e1_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l1_e1_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[59]:phase5.enum_chain_buffer_l1_e2_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[60]:phase5.enum_chain_buffer_l2_e0_tfalse
    session_phase5_corpus_test.go:608: enumerated[60]:phase5.enum_chain_buffer_l2_e0_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l2_e0_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[61]:phase5.enum_chain_buffer_l2_e0_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[62]:phase5.enum_chain_buffer_l2_e1_tfalse
    session_phase5_corpus_test.go:608: enumerated[62]:phase5.enum_chain_buffer_l2_e1_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l2_e1_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[63]:phase5.enum_chain_buffer_l2_e2_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[64]:phase5.enum_chain_buffer_l2_e3_tfalse
    session_phase5_corpus_test.go:608: enumerated[64]:phase5.enum_chain_buffer_l2_e3_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l2_e3_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[65]:phase5.enum_chain_buffer_l2_e3_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[66]:phase5.enum_chain_buffer_l2_e4_tfalse
    session_phase5_corpus_test.go:608: enumerated[66]:phase5.enum_chain_buffer_l2_e4_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l2_e4_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[67]:phase5.enum_chain_buffer_l2_e5_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[68]:phase5.enum_chain_buffer_l2_e6_tfalse
    session_phase5_corpus_test.go:608: enumerated[68]:phase5.enum_chain_buffer_l2_e6_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l2_e6_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[69]:phase5.enum_chain_buffer_l2_e6_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[70]:phase5.enum_chain_buffer_l2_e7_tfalse
    session_phase5_corpus_test.go:608: enumerated[70]:phase5.enum_chain_buffer_l2_e7_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l2_e7_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[71]:phase5.enum_chain_buffer_l2_e8_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[72]:phase5.enum_chain_buffer_l3_e0_tfalse
    session_phase5_corpus_test.go:608: enumerated[72]:phase5.enum_chain_buffer_l3_e0_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e0_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[73]:phase5.enum_chain_buffer_l3_e0_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[74]:phase5.enum_chain_buffer_l3_e1_tfalse
    session_phase5_corpus_test.go:608: enumerated[74]:phase5.enum_chain_buffer_l3_e1_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e1_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[75]:phase5.enum_chain_buffer_l3_e2_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[76]:phase5.enum_chain_buffer_l3_e3_tfalse
    session_phase5_corpus_test.go:608: enumerated[76]:phase5.enum_chain_buffer_l3_e3_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e3_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[77]:phase5.enum_chain_buffer_l3_e3_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[78]:phase5.enum_chain_buffer_l3_e4_tfalse
    session_phase5_corpus_test.go:608: enumerated[78]:phase5.enum_chain_buffer_l3_e4_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e4_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[79]:phase5.enum_chain_buffer_l3_e5_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[80]:phase5.enum_chain_buffer_l3_e6_tfalse
    session_phase5_corpus_test.go:608: enumerated[80]:phase5.enum_chain_buffer_l3_e6_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e6_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[81]:phase5.enum_chain_buffer_l3_e6_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[82]:phase5.enum_chain_buffer_l3_e7_tfalse
    session_phase5_corpus_test.go:608: enumerated[82]:phase5.enum_chain_buffer_l3_e7_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e7_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[83]:phase5.enum_chain_buffer_l3_e8_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[84]:phase5.enum_chain_buffer_l3_e9_tfalse
    session_phase5_corpus_test.go:608: enumerated[84]:phase5.enum_chain_buffer_l3_e9_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e9_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[85]:phase5.enum_chain_buffer_l3_e9_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[86]:phase5.enum_chain_buffer_l3_e10_tfalse
    session_phase5_corpus_test.go:608: enumerated[86]:phase5.enum_chain_buffer_l3_e10_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e10_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[87]:phase5.enum_chain_buffer_l3_e11_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[88]:phase5.enum_chain_buffer_l3_e12_tfalse
    session_phase5_corpus_test.go:608: enumerated[88]:phase5.enum_chain_buffer_l3_e12_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e12_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[89]:phase5.enum_chain_buffer_l3_e12_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[90]:phase5.enum_chain_buffer_l3_e13_tfalse
    session_phase5_corpus_test.go:608: enumerated[90]:phase5.enum_chain_buffer_l3_e13_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e13_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[91]:phase5.enum_chain_buffer_l3_e14_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[92]:phase5.enum_chain_buffer_l3_e15_tfalse
    session_phase5_corpus_test.go:608: enumerated[92]:phase5.enum_chain_buffer_l3_e15_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e15_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[93]:phase5.enum_chain_buffer_l3_e15_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[94]:phase5.enum_chain_buffer_l3_e16_tfalse
    session_phase5_corpus_test.go:608: enumerated[94]:phase5.enum_chain_buffer_l3_e16_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e16_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[95]:phase5.enum_chain_buffer_l3_e17_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[96]:phase5.enum_chain_buffer_l3_e18_tfalse
    session_phase5_corpus_test.go:608: enumerated[96]:phase5.enum_chain_buffer_l3_e18_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e18_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[97]:phase5.enum_chain_buffer_l3_e18_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[98]:phase5.enum_chain_buffer_l3_e19_tfalse
    session_phase5_corpus_test.go:608: enumerated[98]:phase5.enum_chain_buffer_l3_e19_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e19_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[99]:phase5.enum_chain_buffer_l3_e20_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[100]:phase5.enum_chain_buffer_l3_e21_tfalse
    session_phase5_corpus_test.go:608: enumerated[100]:phase5.enum_chain_buffer_l3_e21_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e21_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[101]:phase5.enum_chain_buffer_l3_e21_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[102]:phase5.enum_chain_buffer_l3_e22_tfalse
    session_phase5_corpus_test.go:608: enumerated[102]:phase5.enum_chain_buffer_l3_e22_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e22_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[103]:phase5.enum_chain_buffer_l3_e23_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[104]:phase5.enum_chain_buffer_l3_e24_tfalse
    session_phase5_corpus_test.go:608: enumerated[104]:phase5.enum_chain_buffer_l3_e24_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e24_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[105]:phase5.enum_chain_buffer_l3_e24_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[106]:phase5.enum_chain_buffer_l3_e25_tfalse
    session_phase5_corpus_test.go:608: enumerated[106]:phase5.enum_chain_buffer_l3_e25_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e25_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[107]:phase5.enum_chain_buffer_l3_e26_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[108]:phase5.enum_foreign_try1_alt1
    session_phase5_corpus_test.go:608: enumerated[108]:phase5.enum_foreign_try1_alt1: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_foreign_try1_alt1:fn:main": multi-function foreign-call bodies are not supported by native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[109]:phase5.enum_foreign_discard1_alt1
    session_phase5_corpus_test.go:608: enumerated[109]:phase5.enum_foreign_discard1_alt1: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_foreign_discard1_alt1:fn:main": multi-function foreign-call bodies are not supported by native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[110]:phase5.enum_foreign_try1_alt2
    session_phase5_corpus_test.go:608: enumerated[110]:phase5.enum_foreign_try1_alt2: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_foreign_try1_alt2:fn:main": multi-function foreign-call bodies are not supported by native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[111]:phase5.enum_foreign_discard1_alt2
    session_phase5_corpus_test.go:608: enumerated[111]:phase5.enum_foreign_discard1_alt2: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_foreign_discard1_alt2:fn:main": multi-function foreign-call bodies are not supported by native emission this phase
=== NAME  TestPhase5CorpusThreeEngineAgreement/enumerated-closure
    session_phase5_corpus_test.go:611: phase20-closure-cache programs=112 reused=0 recomputed=112 not_cacheable=0 unavailable=0
--- PASS: TestPhase5CorpusThreeEngineAgreement (65.10s)
    --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure (65.09s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[0]:phase5.enum_chain_byte_l0_e0_tfalse (1.05s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[1]:phase5.enum_chain_byte_l0_e0_ttrue (1.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[2]:phase5.enum_chain_byte_l1_e0_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[3]:phase5.enum_chain_byte_l1_e0_ttrue (1.06s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[4]:phase5.enum_chain_byte_l1_e1_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[5]:phase5.enum_chain_byte_l1_e2_tfalse (1.08s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[6]:phase5.enum_chain_byte_l2_e0_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[7]:phase5.enum_chain_byte_l2_e0_ttrue (1.05s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[8]:phase5.enum_chain_byte_l2_e1_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[9]:phase5.enum_chain_byte_l2_e2_tfalse (1.04s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[10]:phase5.enum_chain_byte_l2_e3_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[11]:phase5.enum_chain_byte_l2_e3_ttrue (1.04s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[12]:phase5.enum_chain_byte_l2_e4_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[13]:phase5.enum_chain_byte_l2_e5_tfalse (1.01s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[14]:phase5.enum_chain_byte_l2_e6_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[15]:phase5.enum_chain_byte_l2_e6_ttrue (1.05s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[16]:phase5.enum_chain_byte_l2_e7_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[17]:phase5.enum_chain_byte_l2_e8_tfalse (1.04s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[18]:phase5.enum_chain_byte_l3_e0_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[19]:phase5.enum_chain_byte_l3_e0_ttrue (1.08s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[20]:phase5.enum_chain_byte_l3_e1_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[21]:phase5.enum_chain_byte_l3_e2_tfalse (1.13s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[22]:phase5.enum_chain_byte_l3_e3_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[23]:phase5.enum_chain_byte_l3_e3_ttrue (1.18s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[24]:phase5.enum_chain_byte_l3_e4_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[25]:phase5.enum_chain_byte_l3_e5_tfalse (1.19s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[26]:phase5.enum_chain_byte_l3_e6_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[27]:phase5.enum_chain_byte_l3_e6_ttrue (1.13s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[28]:phase5.enum_chain_byte_l3_e7_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[29]:phase5.enum_chain_byte_l3_e8_tfalse (1.09s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[30]:phase5.enum_chain_byte_l3_e9_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[31]:phase5.enum_chain_byte_l3_e9_ttrue (1.06s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[32]:phase5.enum_chain_byte_l3_e10_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[33]:phase5.enum_chain_byte_l3_e11_tfalse (1.03s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[34]:phase5.enum_chain_byte_l3_e12_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[35]:phase5.enum_chain_byte_l3_e12_ttrue (1.04s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[36]:phase5.enum_chain_byte_l3_e13_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[37]:phase5.enum_chain_byte_l3_e14_tfalse (1.07s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[38]:phase5.enum_chain_byte_l3_e15_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[39]:phase5.enum_chain_byte_l3_e15_ttrue (1.46s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[40]:phase5.enum_chain_byte_l3_e16_tfalse (0.01s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[41]:phase5.enum_chain_byte_l3_e17_tfalse (1.96s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[42]:phase5.enum_chain_byte_l3_e18_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[43]:phase5.enum_chain_byte_l3_e18_ttrue (1.50s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[44]:phase5.enum_chain_byte_l3_e19_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[45]:phase5.enum_chain_byte_l3_e20_tfalse (1.41s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[46]:phase5.enum_chain_byte_l3_e21_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[47]:phase5.enum_chain_byte_l3_e21_ttrue (1.35s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[48]:phase5.enum_chain_byte_l3_e22_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[49]:phase5.enum_chain_byte_l3_e23_tfalse (1.34s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[50]:phase5.enum_chain_byte_l3_e24_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[51]:phase5.enum_chain_byte_l3_e24_ttrue (1.43s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[52]:phase5.enum_chain_byte_l3_e25_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[53]:phase5.enum_chain_byte_l3_e26_tfalse (1.37s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[54]:phase5.enum_chain_buffer_l0_e0_tfalse (1.13s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[55]:phase5.enum_chain_buffer_l0_e0_ttrue (1.02s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[56]:phase5.enum_chain_buffer_l1_e0_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[57]:phase5.enum_chain_buffer_l1_e0_ttrue (1.01s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[58]:phase5.enum_chain_buffer_l1_e1_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[59]:phase5.enum_chain_buffer_l1_e2_tfalse (0.94s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[60]:phase5.enum_chain_buffer_l2_e0_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[61]:phase5.enum_chain_buffer_l2_e0_ttrue (1.02s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[62]:phase5.enum_chain_buffer_l2_e1_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[63]:phase5.enum_chain_buffer_l2_e2_tfalse (1.09s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[64]:phase5.enum_chain_buffer_l2_e3_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[65]:phase5.enum_chain_buffer_l2_e3_ttrue (1.09s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[66]:phase5.enum_chain_buffer_l2_e4_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[67]:phase5.enum_chain_buffer_l2_e5_tfalse (1.13s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[68]:phase5.enum_chain_buffer_l2_e6_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[69]:phase5.enum_chain_buffer_l2_e6_ttrue (1.30s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[70]:phase5.enum_chain_buffer_l2_e7_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[71]:phase5.enum_chain_buffer_l2_e8_tfalse (1.31s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[72]:phase5.enum_chain_buffer_l3_e0_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[73]:phase5.enum_chain_buffer_l3_e0_ttrue (1.18s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[74]:phase5.enum_chain_buffer_l3_e1_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[75]:phase5.enum_chain_buffer_l3_e2_tfalse (1.12s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[76]:phase5.enum_chain_buffer_l3_e3_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[77]:phase5.enum_chain_buffer_l3_e3_ttrue (1.05s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[78]:phase5.enum_chain_buffer_l3_e4_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[79]:phase5.enum_chain_buffer_l3_e5_tfalse (1.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[80]:phase5.enum_chain_buffer_l3_e6_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[81]:phase5.enum_chain_buffer_l3_e6_ttrue (0.96s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[82]:phase5.enum_chain_buffer_l3_e7_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[83]:phase5.enum_chain_buffer_l3_e8_tfalse (0.98s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[84]:phase5.enum_chain_buffer_l3_e9_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[85]:phase5.enum_chain_buffer_l3_e9_ttrue (1.05s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[86]:phase5.enum_chain_buffer_l3_e10_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[87]:phase5.enum_chain_buffer_l3_e11_tfalse (1.05s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[88]:phase5.enum_chain_buffer_l3_e12_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[89]:phase5.enum_chain_buffer_l3_e12_ttrue (1.03s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[90]:phase5.enum_chain_buffer_l3_e13_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[91]:phase5.enum_chain_buffer_l3_e14_tfalse (1.19s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[92]:phase5.enum_chain_buffer_l3_e15_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[93]:phase5.enum_chain_buffer_l3_e15_ttrue (1.41s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[94]:phase5.enum_chain_buffer_l3_e16_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[95]:phase5.enum_chain_buffer_l3_e17_tfalse (1.66s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[96]:phase5.enum_chain_buffer_l3_e18_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[97]:phase5.enum_chain_buffer_l3_e18_ttrue (1.42s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[98]:phase5.enum_chain_buffer_l3_e19_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[99]:phase5.enum_chain_buffer_l3_e20_tfalse (1.36s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[100]:phase5.enum_chain_buffer_l3_e21_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[101]:phase5.enum_chain_buffer_l3_e21_ttrue (1.20s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[102]:phase5.enum_chain_buffer_l3_e22_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[103]:phase5.enum_chain_buffer_l3_e23_tfalse (1.10s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[104]:phase5.enum_chain_buffer_l3_e24_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[105]:phase5.enum_chain_buffer_l3_e24_ttrue (1.04s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[106]:phase5.enum_chain_buffer_l3_e25_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[107]:phase5.enum_chain_buffer_l3_e26_tfalse (0.94s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[108]:phase5.enum_foreign_try1_alt1 (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[109]:phase5.enum_foreign_discard1_alt1 (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[110]:phase5.enum_foreign_try1_alt2 (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[111]:phase5.enum_foreign_discard1_alt2 (0.00s)
PASS
ok  	github.com/codename-lang/lang/internal/compiler/session	65.362s
```

### closure-warm

```text
=== RUN   TestPhase5CorpusThreeEngineAgreement
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[0]:phase5.enum_chain_byte_l0_e0_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[1]:phase5.enum_chain_byte_l0_e0_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[2]:phase5.enum_chain_byte_l1_e0_tfalse
    session_phase5_corpus_test.go:608: enumerated[2]:phase5.enum_chain_byte_l1_e0_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l1_e0_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[3]:phase5.enum_chain_byte_l1_e0_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[4]:phase5.enum_chain_byte_l1_e1_tfalse
    session_phase5_corpus_test.go:608: enumerated[4]:phase5.enum_chain_byte_l1_e1_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l1_e1_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[5]:phase5.enum_chain_byte_l1_e2_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[6]:phase5.enum_chain_byte_l2_e0_tfalse
    session_phase5_corpus_test.go:608: enumerated[6]:phase5.enum_chain_byte_l2_e0_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l2_e0_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[7]:phase5.enum_chain_byte_l2_e0_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[8]:phase5.enum_chain_byte_l2_e1_tfalse
    session_phase5_corpus_test.go:608: enumerated[8]:phase5.enum_chain_byte_l2_e1_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l2_e1_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[9]:phase5.enum_chain_byte_l2_e2_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[10]:phase5.enum_chain_byte_l2_e3_tfalse
    session_phase5_corpus_test.go:608: enumerated[10]:phase5.enum_chain_byte_l2_e3_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l2_e3_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[11]:phase5.enum_chain_byte_l2_e3_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[12]:phase5.enum_chain_byte_l2_e4_tfalse
    session_phase5_corpus_test.go:608: enumerated[12]:phase5.enum_chain_byte_l2_e4_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l2_e4_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[13]:phase5.enum_chain_byte_l2_e5_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[14]:phase5.enum_chain_byte_l2_e6_tfalse
    session_phase5_corpus_test.go:608: enumerated[14]:phase5.enum_chain_byte_l2_e6_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l2_e6_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[15]:phase5.enum_chain_byte_l2_e6_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[16]:phase5.enum_chain_byte_l2_e7_tfalse
    session_phase5_corpus_test.go:608: enumerated[16]:phase5.enum_chain_byte_l2_e7_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l2_e7_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[17]:phase5.enum_chain_byte_l2_e8_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[18]:phase5.enum_chain_byte_l3_e0_tfalse
    session_phase5_corpus_test.go:608: enumerated[18]:phase5.enum_chain_byte_l3_e0_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e0_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[19]:phase5.enum_chain_byte_l3_e0_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[20]:phase5.enum_chain_byte_l3_e1_tfalse
    session_phase5_corpus_test.go:608: enumerated[20]:phase5.enum_chain_byte_l3_e1_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e1_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[21]:phase5.enum_chain_byte_l3_e2_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[22]:phase5.enum_chain_byte_l3_e3_tfalse
    session_phase5_corpus_test.go:608: enumerated[22]:phase5.enum_chain_byte_l3_e3_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e3_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[23]:phase5.enum_chain_byte_l3_e3_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[24]:phase5.enum_chain_byte_l3_e4_tfalse
    session_phase5_corpus_test.go:608: enumerated[24]:phase5.enum_chain_byte_l3_e4_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e4_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[25]:phase5.enum_chain_byte_l3_e5_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[26]:phase5.enum_chain_byte_l3_e6_tfalse
    session_phase5_corpus_test.go:608: enumerated[26]:phase5.enum_chain_byte_l3_e6_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e6_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[27]:phase5.enum_chain_byte_l3_e6_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[28]:phase5.enum_chain_byte_l3_e7_tfalse
    session_phase5_corpus_test.go:608: enumerated[28]:phase5.enum_chain_byte_l3_e7_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e7_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[29]:phase5.enum_chain_byte_l3_e8_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[30]:phase5.enum_chain_byte_l3_e9_tfalse
    session_phase5_corpus_test.go:608: enumerated[30]:phase5.enum_chain_byte_l3_e9_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e9_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[31]:phase5.enum_chain_byte_l3_e9_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[32]:phase5.enum_chain_byte_l3_e10_tfalse
    session_phase5_corpus_test.go:608: enumerated[32]:phase5.enum_chain_byte_l3_e10_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e10_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[33]:phase5.enum_chain_byte_l3_e11_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[34]:phase5.enum_chain_byte_l3_e12_tfalse
    session_phase5_corpus_test.go:608: enumerated[34]:phase5.enum_chain_byte_l3_e12_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e12_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[35]:phase5.enum_chain_byte_l3_e12_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[36]:phase5.enum_chain_byte_l3_e13_tfalse
    session_phase5_corpus_test.go:608: enumerated[36]:phase5.enum_chain_byte_l3_e13_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e13_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[37]:phase5.enum_chain_byte_l3_e14_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[38]:phase5.enum_chain_byte_l3_e15_tfalse
    session_phase5_corpus_test.go:608: enumerated[38]:phase5.enum_chain_byte_l3_e15_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e15_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[39]:phase5.enum_chain_byte_l3_e15_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[40]:phase5.enum_chain_byte_l3_e16_tfalse
    session_phase5_corpus_test.go:608: enumerated[40]:phase5.enum_chain_byte_l3_e16_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e16_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[41]:phase5.enum_chain_byte_l3_e17_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[42]:phase5.enum_chain_byte_l3_e18_tfalse
    session_phase5_corpus_test.go:608: enumerated[42]:phase5.enum_chain_byte_l3_e18_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e18_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[43]:phase5.enum_chain_byte_l3_e18_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[44]:phase5.enum_chain_byte_l3_e19_tfalse
    session_phase5_corpus_test.go:608: enumerated[44]:phase5.enum_chain_byte_l3_e19_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e19_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[45]:phase5.enum_chain_byte_l3_e20_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[46]:phase5.enum_chain_byte_l3_e21_tfalse
    session_phase5_corpus_test.go:608: enumerated[46]:phase5.enum_chain_byte_l3_e21_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e21_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[47]:phase5.enum_chain_byte_l3_e21_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[48]:phase5.enum_chain_byte_l3_e22_tfalse
    session_phase5_corpus_test.go:608: enumerated[48]:phase5.enum_chain_byte_l3_e22_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e22_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[49]:phase5.enum_chain_byte_l3_e23_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[50]:phase5.enum_chain_byte_l3_e24_tfalse
    session_phase5_corpus_test.go:608: enumerated[50]:phase5.enum_chain_byte_l3_e24_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e24_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[51]:phase5.enum_chain_byte_l3_e24_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[52]:phase5.enum_chain_byte_l3_e25_tfalse
    session_phase5_corpus_test.go:608: enumerated[52]:phase5.enum_chain_byte_l3_e25_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_byte_l3_e25_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[53]:phase5.enum_chain_byte_l3_e26_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[54]:phase5.enum_chain_buffer_l0_e0_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[55]:phase5.enum_chain_buffer_l0_e0_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[56]:phase5.enum_chain_buffer_l1_e0_tfalse
    session_phase5_corpus_test.go:608: enumerated[56]:phase5.enum_chain_buffer_l1_e0_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l1_e0_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[57]:phase5.enum_chain_buffer_l1_e0_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[58]:phase5.enum_chain_buffer_l1_e1_tfalse
    session_phase5_corpus_test.go:608: enumerated[58]:phase5.enum_chain_buffer_l1_e1_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l1_e1_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[59]:phase5.enum_chain_buffer_l1_e2_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[60]:phase5.enum_chain_buffer_l2_e0_tfalse
    session_phase5_corpus_test.go:608: enumerated[60]:phase5.enum_chain_buffer_l2_e0_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l2_e0_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[61]:phase5.enum_chain_buffer_l2_e0_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[62]:phase5.enum_chain_buffer_l2_e1_tfalse
    session_phase5_corpus_test.go:608: enumerated[62]:phase5.enum_chain_buffer_l2_e1_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l2_e1_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[63]:phase5.enum_chain_buffer_l2_e2_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[64]:phase5.enum_chain_buffer_l2_e3_tfalse
    session_phase5_corpus_test.go:608: enumerated[64]:phase5.enum_chain_buffer_l2_e3_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l2_e3_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[65]:phase5.enum_chain_buffer_l2_e3_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[66]:phase5.enum_chain_buffer_l2_e4_tfalse
    session_phase5_corpus_test.go:608: enumerated[66]:phase5.enum_chain_buffer_l2_e4_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l2_e4_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[67]:phase5.enum_chain_buffer_l2_e5_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[68]:phase5.enum_chain_buffer_l2_e6_tfalse
    session_phase5_corpus_test.go:608: enumerated[68]:phase5.enum_chain_buffer_l2_e6_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l2_e6_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[69]:phase5.enum_chain_buffer_l2_e6_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[70]:phase5.enum_chain_buffer_l2_e7_tfalse
    session_phase5_corpus_test.go:608: enumerated[70]:phase5.enum_chain_buffer_l2_e7_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l2_e7_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[71]:phase5.enum_chain_buffer_l2_e8_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[72]:phase5.enum_chain_buffer_l3_e0_tfalse
    session_phase5_corpus_test.go:608: enumerated[72]:phase5.enum_chain_buffer_l3_e0_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e0_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[73]:phase5.enum_chain_buffer_l3_e0_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[74]:phase5.enum_chain_buffer_l3_e1_tfalse
    session_phase5_corpus_test.go:608: enumerated[74]:phase5.enum_chain_buffer_l3_e1_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e1_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[75]:phase5.enum_chain_buffer_l3_e2_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[76]:phase5.enum_chain_buffer_l3_e3_tfalse
    session_phase5_corpus_test.go:608: enumerated[76]:phase5.enum_chain_buffer_l3_e3_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e3_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[77]:phase5.enum_chain_buffer_l3_e3_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[78]:phase5.enum_chain_buffer_l3_e4_tfalse
    session_phase5_corpus_test.go:608: enumerated[78]:phase5.enum_chain_buffer_l3_e4_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e4_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[79]:phase5.enum_chain_buffer_l3_e5_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[80]:phase5.enum_chain_buffer_l3_e6_tfalse
    session_phase5_corpus_test.go:608: enumerated[80]:phase5.enum_chain_buffer_l3_e6_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e6_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[81]:phase5.enum_chain_buffer_l3_e6_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[82]:phase5.enum_chain_buffer_l3_e7_tfalse
    session_phase5_corpus_test.go:608: enumerated[82]:phase5.enum_chain_buffer_l3_e7_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e7_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[83]:phase5.enum_chain_buffer_l3_e8_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[84]:phase5.enum_chain_buffer_l3_e9_tfalse
    session_phase5_corpus_test.go:608: enumerated[84]:phase5.enum_chain_buffer_l3_e9_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e9_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[85]:phase5.enum_chain_buffer_l3_e9_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[86]:phase5.enum_chain_buffer_l3_e10_tfalse
    session_phase5_corpus_test.go:608: enumerated[86]:phase5.enum_chain_buffer_l3_e10_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e10_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[87]:phase5.enum_chain_buffer_l3_e11_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[88]:phase5.enum_chain_buffer_l3_e12_tfalse
    session_phase5_corpus_test.go:608: enumerated[88]:phase5.enum_chain_buffer_l3_e12_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e12_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[89]:phase5.enum_chain_buffer_l3_e12_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[90]:phase5.enum_chain_buffer_l3_e13_tfalse
    session_phase5_corpus_test.go:608: enumerated[90]:phase5.enum_chain_buffer_l3_e13_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e13_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[91]:phase5.enum_chain_buffer_l3_e14_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[92]:phase5.enum_chain_buffer_l3_e15_tfalse
    session_phase5_corpus_test.go:608: enumerated[92]:phase5.enum_chain_buffer_l3_e15_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e15_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[93]:phase5.enum_chain_buffer_l3_e15_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[94]:phase5.enum_chain_buffer_l3_e16_tfalse
    session_phase5_corpus_test.go:608: enumerated[94]:phase5.enum_chain_buffer_l3_e16_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e16_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[95]:phase5.enum_chain_buffer_l3_e17_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[96]:phase5.enum_chain_buffer_l3_e18_tfalse
    session_phase5_corpus_test.go:608: enumerated[96]:phase5.enum_chain_buffer_l3_e18_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e18_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[97]:phase5.enum_chain_buffer_l3_e18_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[98]:phase5.enum_chain_buffer_l3_e19_tfalse
    session_phase5_corpus_test.go:608: enumerated[98]:phase5.enum_chain_buffer_l3_e19_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e19_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[99]:phase5.enum_chain_buffer_l3_e20_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[100]:phase5.enum_chain_buffer_l3_e21_tfalse
    session_phase5_corpus_test.go:608: enumerated[100]:phase5.enum_chain_buffer_l3_e21_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e21_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[101]:phase5.enum_chain_buffer_l3_e21_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[102]:phase5.enum_chain_buffer_l3_e22_tfalse
    session_phase5_corpus_test.go:608: enumerated[102]:phase5.enum_chain_buffer_l3_e22_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e22_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[103]:phase5.enum_chain_buffer_l3_e23_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[104]:phase5.enum_chain_buffer_l3_e24_tfalse
    session_phase5_corpus_test.go:608: enumerated[104]:phase5.enum_chain_buffer_l3_e24_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e24_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[105]:phase5.enum_chain_buffer_l3_e24_ttrue
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[106]:phase5.enum_chain_buffer_l3_e25_tfalse
    session_phase5_corpus_test.go:608: enumerated[106]:phase5.enum_chain_buffer_l3_e25_tfalse: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_chain_buffer_l3_e25_tfalse:fn:f": by-pointer bodies are not supported by whole-program native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[107]:phase5.enum_chain_buffer_l3_e26_tfalse
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[108]:phase5.enum_foreign_try1_alt1
    session_phase5_corpus_test.go:608: enumerated[108]:phase5.enum_foreign_try1_alt1: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_foreign_try1_alt1:fn:main": multi-function foreign-call bodies are not supported by native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[109]:phase5.enum_foreign_discard1_alt1
    session_phase5_corpus_test.go:608: enumerated[109]:phase5.enum_foreign_discard1_alt1: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_foreign_discard1_alt1:fn:main": multi-function foreign-call bodies are not supported by native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[110]:phase5.enum_foreign_try1_alt2
    session_phase5_corpus_test.go:608: enumerated[110]:phase5.enum_foreign_try1_alt2: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_foreign_try1_alt2:fn:main": multi-function foreign-call bodies are not supported by native emission this phase
=== RUN   TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[111]:phase5.enum_foreign_discard1_alt2
    session_phase5_corpus_test.go:608: enumerated[111]:phase5.enum_foreign_discard1_alt2: public native emitter refusal has no frozen evidence; native differential is inapplicable: function "s1:phase5.enum_foreign_discard1_alt2:fn:main": multi-function foreign-call bodies are not supported by native emission this phase
=== NAME  TestPhase5CorpusThreeEngineAgreement/enumerated-closure
    session_phase5_corpus_test.go:611: phase20-closure-cache programs=112 reused=112 recomputed=0 not_cacheable=0 unavailable=0
--- PASS: TestPhase5CorpusThreeEngineAgreement (47.51s)
    --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure (47.51s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[0]:phase5.enum_chain_byte_l0_e0_tfalse (0.79s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[1]:phase5.enum_chain_byte_l0_e0_ttrue (0.75s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[2]:phase5.enum_chain_byte_l1_e0_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[3]:phase5.enum_chain_byte_l1_e0_ttrue (0.73s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[4]:phase5.enum_chain_byte_l1_e1_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[5]:phase5.enum_chain_byte_l1_e2_tfalse (0.68s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[6]:phase5.enum_chain_byte_l2_e0_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[7]:phase5.enum_chain_byte_l2_e0_ttrue (0.70s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[8]:phase5.enum_chain_byte_l2_e1_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[9]:phase5.enum_chain_byte_l2_e2_tfalse (0.81s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[10]:phase5.enum_chain_byte_l2_e3_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[11]:phase5.enum_chain_byte_l2_e3_ttrue (0.78s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[12]:phase5.enum_chain_byte_l2_e4_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[13]:phase5.enum_chain_byte_l2_e5_tfalse (0.84s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[14]:phase5.enum_chain_byte_l2_e6_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[15]:phase5.enum_chain_byte_l2_e6_ttrue (0.88s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[16]:phase5.enum_chain_byte_l2_e7_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[17]:phase5.enum_chain_byte_l2_e8_tfalse (1.14s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[18]:phase5.enum_chain_byte_l3_e0_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[19]:phase5.enum_chain_byte_l3_e0_ttrue (1.16s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[20]:phase5.enum_chain_byte_l3_e1_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[21]:phase5.enum_chain_byte_l3_e2_tfalse (1.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[22]:phase5.enum_chain_byte_l3_e3_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[23]:phase5.enum_chain_byte_l3_e3_ttrue (0.94s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[24]:phase5.enum_chain_byte_l3_e4_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[25]:phase5.enum_chain_byte_l3_e5_tfalse (0.91s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[26]:phase5.enum_chain_byte_l3_e6_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[27]:phase5.enum_chain_byte_l3_e6_ttrue (0.88s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[28]:phase5.enum_chain_byte_l3_e7_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[29]:phase5.enum_chain_byte_l3_e8_tfalse (0.84s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[30]:phase5.enum_chain_byte_l3_e9_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[31]:phase5.enum_chain_byte_l3_e9_ttrue (0.89s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[32]:phase5.enum_chain_byte_l3_e10_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[33]:phase5.enum_chain_byte_l3_e11_tfalse (0.79s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[34]:phase5.enum_chain_byte_l3_e12_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[35]:phase5.enum_chain_byte_l3_e12_ttrue (0.78s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[36]:phase5.enum_chain_byte_l3_e13_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[37]:phase5.enum_chain_byte_l3_e14_tfalse (0.81s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[38]:phase5.enum_chain_byte_l3_e15_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[39]:phase5.enum_chain_byte_l3_e15_ttrue (0.87s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[40]:phase5.enum_chain_byte_l3_e16_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[41]:phase5.enum_chain_byte_l3_e17_tfalse (0.77s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[42]:phase5.enum_chain_byte_l3_e18_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[43]:phase5.enum_chain_byte_l3_e18_ttrue (0.79s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[44]:phase5.enum_chain_byte_l3_e19_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[45]:phase5.enum_chain_byte_l3_e20_tfalse (0.83s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[46]:phase5.enum_chain_byte_l3_e21_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[47]:phase5.enum_chain_byte_l3_e21_ttrue (0.84s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[48]:phase5.enum_chain_byte_l3_e22_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[49]:phase5.enum_chain_byte_l3_e23_tfalse (0.86s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[50]:phase5.enum_chain_byte_l3_e24_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[51]:phase5.enum_chain_byte_l3_e24_ttrue (0.84s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[52]:phase5.enum_chain_byte_l3_e25_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[53]:phase5.enum_chain_byte_l3_e26_tfalse (0.86s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[54]:phase5.enum_chain_buffer_l0_e0_tfalse (0.87s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[55]:phase5.enum_chain_buffer_l0_e0_ttrue (0.84s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[56]:phase5.enum_chain_buffer_l1_e0_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[57]:phase5.enum_chain_buffer_l1_e0_ttrue (0.95s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[58]:phase5.enum_chain_buffer_l1_e1_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[59]:phase5.enum_chain_buffer_l1_e2_tfalse (0.92s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[60]:phase5.enum_chain_buffer_l2_e0_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[61]:phase5.enum_chain_buffer_l2_e0_ttrue (1.01s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[62]:phase5.enum_chain_buffer_l2_e1_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[63]:phase5.enum_chain_buffer_l2_e2_tfalse (0.98s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[64]:phase5.enum_chain_buffer_l2_e3_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[65]:phase5.enum_chain_buffer_l2_e3_ttrue (0.84s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[66]:phase5.enum_chain_buffer_l2_e4_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[67]:phase5.enum_chain_buffer_l2_e5_tfalse (0.85s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[68]:phase5.enum_chain_buffer_l2_e6_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[69]:phase5.enum_chain_buffer_l2_e6_ttrue (0.86s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[70]:phase5.enum_chain_buffer_l2_e7_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[71]:phase5.enum_chain_buffer_l2_e8_tfalse (0.79s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[72]:phase5.enum_chain_buffer_l3_e0_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[73]:phase5.enum_chain_buffer_l3_e0_ttrue (0.80s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[74]:phase5.enum_chain_buffer_l3_e1_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[75]:phase5.enum_chain_buffer_l3_e2_tfalse (0.78s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[76]:phase5.enum_chain_buffer_l3_e3_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[77]:phase5.enum_chain_buffer_l3_e3_ttrue (0.78s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[78]:phase5.enum_chain_buffer_l3_e4_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[79]:phase5.enum_chain_buffer_l3_e5_tfalse (0.76s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[80]:phase5.enum_chain_buffer_l3_e6_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[81]:phase5.enum_chain_buffer_l3_e6_ttrue (0.78s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[82]:phase5.enum_chain_buffer_l3_e7_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[83]:phase5.enum_chain_buffer_l3_e8_tfalse (0.82s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[84]:phase5.enum_chain_buffer_l3_e9_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[85]:phase5.enum_chain_buffer_l3_e9_ttrue (0.86s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[86]:phase5.enum_chain_buffer_l3_e10_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[87]:phase5.enum_chain_buffer_l3_e11_tfalse (0.89s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[88]:phase5.enum_chain_buffer_l3_e12_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[89]:phase5.enum_chain_buffer_l3_e12_ttrue (0.82s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[90]:phase5.enum_chain_buffer_l3_e13_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[91]:phase5.enum_chain_buffer_l3_e14_tfalse (0.84s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[92]:phase5.enum_chain_buffer_l3_e15_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[93]:phase5.enum_chain_buffer_l3_e15_ttrue (0.85s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[94]:phase5.enum_chain_buffer_l3_e16_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[95]:phase5.enum_chain_buffer_l3_e17_tfalse (0.80s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[96]:phase5.enum_chain_buffer_l3_e18_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[97]:phase5.enum_chain_buffer_l3_e18_ttrue (0.77s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[98]:phase5.enum_chain_buffer_l3_e19_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[99]:phase5.enum_chain_buffer_l3_e20_tfalse (0.85s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[100]:phase5.enum_chain_buffer_l3_e21_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[101]:phase5.enum_chain_buffer_l3_e21_ttrue (0.85s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[102]:phase5.enum_chain_buffer_l3_e22_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[103]:phase5.enum_chain_buffer_l3_e23_tfalse (0.86s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[104]:phase5.enum_chain_buffer_l3_e24_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[105]:phase5.enum_chain_buffer_l3_e24_ttrue (0.85s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[106]:phase5.enum_chain_buffer_l3_e25_tfalse (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[107]:phase5.enum_chain_buffer_l3_e26_tfalse (0.83s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[108]:phase5.enum_foreign_try1_alt1 (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[109]:phase5.enum_foreign_discard1_alt1 (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[110]:phase5.enum_foreign_try1_alt2 (0.00s)
        --- PASS: TestPhase5CorpusThreeEngineAgreement/enumerated-closure/enumerated[111]:phase5.enum_foreign_discard1_alt2 (0.00s)
PASS
ok  	github.com/codename-lang/lang/internal/compiler/session	47.735s
```
