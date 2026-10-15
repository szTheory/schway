# Phase 37 remediation evidence

This plan-local receipt records the CRC helper-placement repair and the missing interpreter-retention and file-reader mutation controls. It does not close security findings, accept Phase 37, mark requirements complete, or make a phase-wide validation claim.

## Revision and hash bindings

The production repair was committed at `8d0d7c6acb2ecefee5c1fc5643a70d5289405976` (`fix(37-09): call crc byte helper once per input byte`), after RED commit `fa27e67e` (`test(37-09): pin crc byte-helper placement and activations`). The focused remediation receipts below ran with that code revision. Task 2 test files were uncommitted during those runs; their exact working-tree hashes are included here.

| File | SHA-256 |
| --- | --- |
| `examples/crc32/lib/crc32.schway` | `ce98ba7b055625c0e5a627f7d769cfbba948695755aeadb471578153fdea3707` |
| `examples/crc32/crc32.schway` | `8a8c7570409a6bf96258a35b5964dc899a2347871b564fb59e081beb011388c9` |
| `examples/crc32/crc_low.schway` | `797f9cd2e733c3590a593518fe671a1fda553eb6ce22ea50d70f1a26d606dccc` |
| `internal/compiler/session/session_phase37_reuse_test.go` | `80b688ab11c2ac6911e745cc7ca82375ea4712d5ab3c503821f56ca5a12922e7` |
| `cmd/schway/main_phase37_crc32_controls_test.go` | `9f14a8b3e330fd6a4546b9e750ed9acc2402792f7493bfde2b68dadb67532517` |
| `internal/compiler/interp/interp_phase37_test.go` | `ec85968a189fb4a4e1259ef76bd3007393165227120318d99cd14ce48ad6bce8` |
| `internal/compiler/interp/filebytes_reader_test.go` | `0f132482303a87d4b548836a8c220149b9ea3a9710f14bf45f63e66471c1a9c0` |
| `internal/compiler/interp/interp.go` (unchanged) | `aba3e9245fb5189527faf0b7e95757bfc8c8c5985f063701b0f697dbdfd2afbd` |
| `internal/compiler/interp/filebytes_reader.go` (unchanged) | `f84db45d89f51130810d0652b655f88f127a168c1cce5a7c0aad4473e51cfffd` |
| `testdata/phase37/new-fixture-ledger.json` | `2fd2bbf340d95212966aade306a155ab4cb526c077cfb10f9b6405e2dcaded7d` |
| `.planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-PROHIBITION-EVIDENCE.md` | `d69a477cb4a15025dde65630e717f8918d5ccb86d0147a84168f1007730205d8` |

The existing `37-PROHIBITION-EVIDENCE.md` was refreshed by the requested `--prove` command. It now binds the three projected-descriptor proof to the repair revision and its current test output digests. Historical corpus files, Phase 36 ledger rows, and prior acceptance receipts were not rewritten.

## CRC helper placement and activation

`crc32` now has one byte loop and makes one private `crc_byte` call per byte. `crc_byte` owns the eight-round bit loop. The reuse AST check pins both loop bounds and the single call site. `TestPhase37CRC32ByteHelperCalls` independently runs a three-byte interpreter input and counts three actual `function.called` events for the private helper. This is a structural activation check, not CRC numeric proof and not a maximum-size activation run. The 65,536-byte source bound follows from one helper call per byte and the separately tested reader cap.

The exact task-1 verification command passed all ten selected roots, including the native tracer, both consumers, source identity/copy check, helper activation, golden controls, and both additive-ledger controls:

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase37CRC32NativeTracer|TestPhase37CRC32Consumers|TestPhase37CRC32SingleCopyAndIdentity|TestPhase37CRC32ByteHelperCalls|TestPhase34PreChangeCProgramGolden|TestPhase34PreChangeCGoldenIsNotInert|TestPhase36NoCallGeneratedCByteBaseline|TestPhase36NoCallGeneratedCByteBaselineControls|TestPhase37Phase34AdditionLedgerControls|TestPhase37Phase36AdditionLedgerControls)$' -count=1 -v -timeout=600s
```

The raw output was captured at `/tmp/schway37-task1-final.log`; final line: `ok github.com/szTheory/schway/internal/compiler/session 5.592s`.

## Reached negative controls

The controls use temporary Go overlays with exact single-occurrence anchors. Each child is deadline-bounded (`go test -timeout=75s`, parent context 90s), runs only its target root, verifies the child's computed project root equals `/Users/jon/projects/schway`, and must print both the exact `=== RUN` marker and its named semantic assertion. The parent rejects a passing mutant, a missing root marker, a compile/setup failure, timeout, or any failure that omits the specific semantic assertion. When unarmed, both target roots run clean baseline assertions: the capture target verifies zero growth on successful and late-defect executions; the reader target verifies cap-plus-one refusal, nonregular refusal, and close-error handling. Armed children reach their named mutation assertions. No package or project suite was run.

The repeated-call fixture returns `8` and writes `8\n` after three scalar calls. Every capture mutant also runs the existing checked late-defect program and requires outcome `defect`, prior output `5\n`, the existing one-live-owner policy, and zero semantic releases. The clean `TestPhase37CaptureDisabledOutcomes` roots separately passed successful zero/nonzero inputs, all three typed failures, the prior-output late defect, and the capacity-exhaustion late-defect case.

Verbatim named assertion lines from the final focused run:

```text
RETENTION_ASSERTION[recorder]: evidence creations=1
RETENTION_ASSERTION[event]: event constructions=1 retained events=1
RETENTION_ASSERTION[invocation]: retained invocation="inv:entry:s1:phase321.file_defect:fn:main" entryID=""
RETENTION_ASSERTION[segment]: segment allocations=4
RETENTION_ASSERTION[ordinal]: ordinal map allocations=2
READER_ASSERTION[wrong-read-bound]: cap+1 must return nil/FileTooLarge after reading 65537 bytes; got len=65536 alternative="" closes=1
READER_ASSERTION[omitted-regular-check]: nonregular descriptor must return nil/FileUnavailable without reads and close once; got len=6 alternative="" reads=2 closes=1
READER_ASSERTION[ignored-close-error]: successful read with close error must return nil/FileUnavailable; got bytes="ok" alternative="" closes=1
```

The reader mutants operated on `filebytes_reader.go` overlays: the cap mutant slices to `FileBytesLimit`; the type-check mutant replaces refusal with a no-op mode query; the close mutant explicitly ignores `closeErr`. An initial overlay attempt for the latter two did not compile because their now-unused locals remained; those attempts were rejected as evidence and corrected. The final child runs above compile, reach the reader, and fail at the named wrong-result/count assertion. The ordinary reader fault/bound controls also passed, including cap-1/cap/cap+1, forged size metadata, replacement-after-open, FIFO/device refusal, and close-error precedence.

The task-2 focused command and canonical prohibition proof both passed:

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/interp -run '^(TestPhase37CaptureRetentionControls|TestPhase37CaptureRetentionMutationTarget|TestPhase37CaptureDisabledOutcomes|TestPhase37FileReaderFaults|TestPhase37FileReaderBoundControls|TestPhase37FileReaderMutationControls|TestPhase37FileReaderMutationTarget)$' -count=1 -v -timeout=600s
node .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-prohibition-check.test.cjs --prove
```

The final focused output is `/tmp/schway37-task2-hygiene-final.log` and ends `ok github.com/szTheory/schway/internal/compiler/interp 2.018s`. The final proof output is `/tmp/schway37-C06-hygiene-final.log`: `Phase37 canonical prohibition proof passed for all three projected descriptors.`

## Affected validation receipts

These are the exact affected C03/C04/C06/C07/C08/C12/C13/C14/C15/C16/C17/C18/C22/C23/C24/C26 commands from `37-VALIDATION.md`. C07, C08, and C12 were rerun after the final interpreter test edits; C22-C24 were rerun after those edits and the C06 evidence refresh. The C22 clean-checkout guide manifest contained 2,696 tracked files with SHA-256 `97c1ef6c4490ce21f5848c09fc67e85bd078eb377b0cfb42efcf9ebf3a774182`.

```sh
# C03
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase37CRC32NativeTracer$' -count=1 -v -timeout=600s
# C04
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase37CRC32NativeTracer|TestPhase37CRC32Consumers)$' -count=1 -v -timeout=600s
# C06
node .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-prohibition-check.test.cjs --prove
# C07
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/interp ./internal/compiler/session -run '^(TestPhase37InterpreterFileApplication|TestPhase37CaptureRetentionControls)$' -count=1 -v -timeout=600s
# C08
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/interp -run '^(TestPhase37CaptureDisabledOutcomes|TestPhase37FileReaderBasic|TestPhase36InterpreterEvidenceBoundaries|TestPhase321AbortLikeDefectFileRoute)$' -count=1 -v -timeout=600s
# C12
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/interp -run '^(TestPhase37FileReaderFaults|TestPhase37FileReaderBoundControls|TestPhase37CaptureRetentionControls|TestPhase37CaptureDisabledOutcomes)$' -count=1 -v -timeout=600s
# C13
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/native -run '^(TestPhase37CRC32PhysicalCleanup|TestPhase37CRC32BoundControls|TestPhase37CRC32Sanitizer)$' -count=1 -v -timeout=600s
# C14
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/testsupport -run '^(TestPhase37CRC32OracleRows|TestPhase37CRC32OracleDriftControls)$' -count=1 -v -timeout=600s
# C15
GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^(TestPhase37PublicFileMatrix|TestPhase37PublicMatrixDriftControl)$' -count=1 -v -timeout=600s
# C16
GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^TestPhase37CRC32AlgorithmMutants$' -count=1 -v -timeout=600s
# C17
GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^TestPhase37CRC32LibraryDualFlip$' -count=1 -v -timeout=600s
# C18
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase37CRC32SingleCopyAndIdentity$' -count=1 -v -timeout=600s
# C22
GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^TestPhase37GuideCleanCheckout$' -count=1 -v -timeout=600s
# C23
GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^(TestPhase37GuideMatchesExpectedJSON|TestPhase37GuideParserNegativeControls|TestPhase32GuideCleanCheckout|TestPhase32GuideMatchesExpectedJSON|TestPhase32GuideParserNegativeControls)$' -count=1 -v -timeout=600s
# C24
GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^(TestPhase37GuideCleanCheckout|TestPhase37GuideMatchesExpectedJSON|TestPhase37GuideParserNegativeControls)$' -count=1 -v -timeout=600s
# C26
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase34PreChangeCProgramGolden|TestPhase34PreChangeCGoldenIsNotInert|TestPhase36NoCallGeneratedCByteBaseline|TestPhase36NoCallGeneratedCByteBaselineControls|TestPhase37Phase34AdditionLedgerControls|TestPhase37Phase36AdditionLedgerControls)$' -count=1 -v -timeout=600s
```

The C08 command above is the exact validation command from `37-VALIDATION.md`; its C06 proof command is listed separately. Every listed receipt exited zero. The logs contain no skipped affected roots. C13’s host-capability, sanitizer success, and cap-plus-one cleanup subtests all ran.

| Receipt | Reached roots and observed result | Captured output |
| --- | --- | --- |
| C03/C04 | Native tracer; both `crc32.schway` and `crc_low.schway` consumers. | `/tmp/schway37-C03.log`, `/tmp/schway37-C04.log` |
| C06 | Canonical bad, clean, and default projections passed the prohibition producer. | `/tmp/schway37-C06-hygiene-final.log` |
| C07/C08/C12 | Interpreter file app; reader basic/fault/bounds; capture retention/outcomes; Phase 36 evidence boundaries; Phase 32 abort-like defect route. Five capture mutants all reached named assertions in C07/C12. | `/tmp/schway37-C07-hygiene-final.log`, `/tmp/schway37-C08-hygiene-final.log`, `/tmp/schway37-C12-hygiene-final.log` |
| C13/C14 | Native cleanup, both physical-bound mutants, sanitizer host/success/cap-plus-one cases; oracle rows and four drift controls. | `/tmp/schway37-C13.log`, `/tmp/schway37-C14.log` |
| C15 | `TestPhase37PublicFileMatrix` reached all 29 named input/control groups: empty, zero, check, highbit-80, highbit-80ff, all-bytes-00-ff, cap-minus-one, cap, cap-plus-one, Wikipedia, missing, directory, FIFO, device, unreadable, input-symlink, opaque-json, opaque-depth, opaque-separator, opaque-leading-dash, opaque-space, opaque-unicode, opaque-shell, empty-token, path-4096, path-4097, nul-direct-dispatch, manifest-controls, isolated-bounded-parallel-pair. The four consumers were exercised through interpreter/native routes where applicable; matrix drift control passed. | `/tmp/schway37-C15.log` |
| C16/C17/C18 | Four numeric algorithm mutants reached both engines; dual-flip reached both local library copies; identity/copy test passed while observing stable InputID and the documented non-cacheable native artifact/BuildID variance. | `/tmp/schway37-C16.log`, `/tmp/schway37-C17.log`, `/tmp/schway37-C18.log` |
| C22/C23/C24 | Phase 37 clean checkout guide; Phase 32 clean checkout and parser controls; Phase 37 JSON (`rows=13`) and parser controls. C22/C24 clean checkouts each ran the guide build and cases for the two CRC apps. | `/tmp/schway37-C22-hygiene-final.log`, `/tmp/schway37-C23-hygiene-final.log`, `/tmp/schway37-C24-hygiene-final.log` |
| C26 | Both historical goldens/baseline receipts and both current ledger owners reached all named closure/mutation controls. | `/tmp/schway37-C26.log` |

Final raw command-result lines:

```text
C03: PASS; ok github.com/szTheory/schway/internal/compiler/session 0.501s
C04: PASS; ok github.com/szTheory/schway/internal/compiler/session 1.463s
C06: Phase37 canonical prohibition proof passed for all three projected descriptors.
C07: PASS; ok github.com/szTheory/schway/internal/compiler/interp 1.905s; ok github.com/szTheory/schway/internal/compiler/session 2.241s
C08: PASS; ok github.com/szTheory/schway/internal/compiler/interp 0.103s
C12: PASS; ok github.com/szTheory/schway/internal/compiler/interp 1.795s
C13: PASS; ok github.com/szTheory/schway/internal/compiler/native 2.642s
C14: PASS; ok github.com/szTheory/schway/internal/compiler/testsupport 0.007s
C15: PASS; ok github.com/szTheory/schway/cmd/schway 38.369s
C16: PASS; ok github.com/szTheory/schway/cmd/schway 2.676s
C17: PASS; ok github.com/szTheory/schway/cmd/schway 1.453s
C18: PASS; ok github.com/szTheory/schway/internal/compiler/session 1.654s
C22: PASS root=TestPhase37GuideCleanCheckout; ok github.com/szTheory/schway/cmd/schway 23.069s
C23: PASS; ok github.com/szTheory/schway/cmd/schway 3.533s
C24: PASS root=TestPhase37GuideCleanCheckout; ok github.com/szTheory/schway/cmd/schway 19.988s
C26: PASS; ok github.com/szTheory/schway/internal/compiler/session 3.597s
```

The four C16 mutants reached these interpreter and native results: wrong reflected polynomial `3421780262 → 3485321504`; non-reflected recurrence `1069182125 → 544806248613`; missing final XOR `0 → 4294967295`; shortened bit-round bound `3523407757 → 2751848219`. C17 reached the paired CRC library flips for both `crc32` and `crc_low`.

## Receipt limits

The activation count is demonstrated on three bytes, not measured across a maximum-size input. C15/C22/C24 are exact focused phase receipts, not permission to infer general native target portability. C06 proves the existing projected prohibition descriptors, not CRC numeric truth. These receipts do not close the 31-item security register or substitute for the parent’s independent audit and phase acceptance.
