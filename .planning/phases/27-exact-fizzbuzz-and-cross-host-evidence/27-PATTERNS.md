# Phase 27: Exact FizzBuzz and Cross-Host Evidence - Pattern Map

**Mapped:** 2026-10-03  
**Files analyzed:** 8 planned file groups  
**Analogs found:** 8 / 8

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `examples/fizzbuzz.schway` | component (ordinary source witness) | request-response / streaming output | `examples/sum_to_n.schway` | role-match |
| `internal/compiler/{syntax,core,check,corevalidate,originvalidate,pathoracle,interp,cgen}/*` | utility (language pipeline consumers) | transform | corresponding existing operator handling, e.g. `internal/compiler/core/core.go` and `internal/compiler/interp/interp.go` | role-match |
| `internal/compiler/session/session.go`, `internal/compiler/native/native_app.go`, `cmd/schway/main.go` | controller / app boundary | request-response / streaming output | `internal/compiler/session/session_phase26_test.go` app path and `cmd/schway/main.go` | role-match |
| `internal/compiler/session/*test.go`, `internal/compiler/native/*test.go`, `internal/compiler/cgen/*test.go` | test | transform / request-response | `internal/compiler/session/session_phase26_test.go`, `session_phase26_overflow_test.go` | exact |
| `scripts/verify-phase27.sh` (name recommended by RESEARCH) | utility / evidence driver | batch | `scripts/verify-phase25.sh` | role-match |
| `examples/phase27/README.md` | documentation | request-response / batch | `examples/phase24/README.md` | role-match |
| `README.md` | documentation / index | request-response | existing Start here list | exact |
| `.github/workflows/ci.yml` | config | batch | Phase 25 verification job | exact |

The RESEARCH structure is a recommendation, not a fixed exhaustive file list. Tests, script, example guide, and workflow integration are inferred from the locked evidence/clean-checkout requirements. No generated/runtime mirror paths are used. Every named analog below was confirmed by `git ls-files`.

## Pattern Assignments

### `examples/fizzbuzz.schway` (ordinary source witness, request-response / streaming output)

**Analog:** `examples/sum_to_n.schway`

**Source module and export pattern** (lines 1-7):
```schway
module examples.sum_to_n

export {
  fn main
}

fn main(n: U64) -> U64 {
```

**Scalar loop and pre-output input gate** (lines 8-19):
```schway
  var i = 0
  var total = 0
  if n < 1001 {
    while i < n {
      i = i + 1
      let snapshot = i
      total = total + snapshot
    }
  } else {
    defect "input exceeds 1000"
  }
  total
```
Use this for mutable scalar initialization, checked arithmetic, branches, and the bounded-input failure shape. FizzBuzz needs a terminal loop arrangement that includes `n` without overflowing at the public maximum; its explicit Unit result and writes are new.

### Language operations across frontend, core, checkers, peers, interpreter, and C emitter (transform)

**Primary analogs:** `internal/compiler/core/core.go`, `internal/compiler/check/scalar.go`, `internal/compiler/interp/interp.go`, `internal/compiler/cgen/cgen.go`

The phase names the precise consumer set in `27-CONTEXT.md`: `syntax/{lexer,parser,format}.go`, `core/core.go`, `check/{scalar,check}.go`, `corevalidate/corevalidate.go`, `originvalidate/originvalidate.go`, `pathoracle/pathoracle.go`, `interp/interp.go`, and `cgen/{cgen_program,cgen}.go`. These are all tracked. Match the nearest existing `OpLessU64` / `OpAddU64` handling in each file; keep producer admission and peer derivation separate. The relevant interpreter location is `internal/compiler/interp/interp.go:1242-1265` (constant/addition/less-than execution); C output accounting is shown below. Research identified `internal/compiler/core/core.go:701-704` as the existing operation inventory seam.

**C application output accounting** (`internal/compiler/cgen/cgen.go`, lines 731-738):
```go
out.WriteString("static size_t schway_event_count = 0u;\nstatic size_t schway_output_count = 0u;\n\n")
if applicationEvidence {
    out.WriteString("static FILE *schway_output_stream = NULL;\nstatic size_t schway_output_limit = SCHWAY_OUTPUT_LIMIT;\nstatic int schway_event_overflow = 0;\n\n")
    out.WriteString("static int schway_write_bytes(const char *data, size_t length) {\n  FILE *stream = schway_output_stream != NULL ? schway_output_stream : stdout;\n  if (schway_output_count > schway_output_limit || length > schway_output_limit - schway_output_count) return 0;\n  if (length != 0u && fwrite(data, 1u, length, stream) != length) return 0;\n  schway_output_count += length;\n  return 1;\n}\n\n")
}
out.WriteString("static int schway_write_literal(const char *value) {\n  return schway_write_bytes(value, strlen(value));\n}\n\n")
```
The existing helper checks each chunk before writing and charges successful chunks. Keep application bytes separate from serialized evidence output; `SetExecutionOutputLimitForTest` in `internal/compiler/cgen/export_test.go:98-105` controls execution-document capacity, not app stdout.

### Application entry and emitted app (controller, request-response / streaming output)

**Analog:** `internal/compiler/session/session.go`

**Build, independent validation, and entry admission** (`internal/compiler/session/session.go`, lines 1345-1367):
```go
func BuildApplication(ctx context.Context, source []byte, outputPath string, runner native.Runner, manifestPath ...string) (native.BuildReceipt, []diagnostic.Diagnostic, error) {
    checked := Check(source)
    if len(checked.Diagnostics) > 0 { return native.BuildReceipt{}, checked.Diagnostics, nil }
    validated := corevalidate.Validate(checked.Program)
    // ... reject failed independent validation ...
    program := validated.Program()
    entry, err := callgraph.EntryFunction(program)
    // ...
    if entry.Match != nil || entry.Linear == nil || (entry.Parameter.Type != "U64" && !localOwnerEntry) || entry.ReturnType != "U64" {
        return native.BuildReceipt{}, nil, fmt.Errorf("application entry %q: only a linear U64-to-U64 or checked PathToken-to-U64 entry is supported", entry.ID)
    }
```
Extend the declared return-mode gate for the narrow Unit entry while preserving U64 auto-print and the PathToken local-owner exception. Follow the existing generated app and CLI process outcome conventions rather than adding a second runner. The phase context points to `internal/compiler/cgen/cgen_program.go:1009-1010`, `internal/compiler/native/native_app.go`, and `cmd/schway/main.go:176-236` as downstream app boundaries.

### Independent output evidence and reached controls (test, request-response)

**Analogs:** `internal/compiler/session/session_phase26_test.go`, `internal/compiler/session/session_phase26_overflow_test.go`

**Per-engine goldens, native outcome checks, input boundary, and reached source mutants** (`session_phase26_test.go`, lines 163-212):
```go
for _, test := range []struct{ input, want string }{{"0", "0"}, {"10", "55"}, {"1000", "500500"}} {
    interpreted, err := interp.Run(checked.Program, "main", test.input)
    if err != nil || interpreted.Outcome.Kind != "returned" || interpreted.Outcome.Value != test.want { ... }
    var stdout, stderr bytes.Buffer
    outcome, err := native.RunApplication(context.Background(), artifact, test.input, &stdout, &stderr)
    if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 0 || stdout.String() != test.want+"\n" || stderr.Len() != 0 { ... }
}
// 1001 must produce exit 65, empty stdout, and the exact app diagnostic.
// Source mutations are required to reach execution and diverge from literal 55.
```
Extend this style so interpreter and native each match independent byte-for-byte FizzBuzz answers before comparing engines; add bounded output and dynamic-zero process controls. `session_phase26_overflow_test.go:21-76` is the close error-path model: interpreter defect, native app exit 65, empty output, stable stderr.

### `scripts/verify-phase27.sh` (utility, batch)

**Analog:** `scripts/verify-phase25.sh`

**Host/build identity and declared-input hash pattern** (lines 24-63):
```sh
host_os=$(uname -s)
host_arch=$(uname -m)
go_os=$(go env GOOS)
go_arch=$(go env GOARCH)
go_version=$(go version)
clang_version=$(clang --version | sed -n '1p')
clang_target=$(clang -dumpmachine)
revision=$(git rev-parse HEAD)
source_inputs="..."
for input in $source_inputs; do
    test -f "$input" || exit 1
    input_hash=$(git hash-object "$input")
done
```
Copy the host identification, source/build input provenance, and fail-closed missing-host/lane behavior. Adapt receipts to Phase 27's declared applications and exact inputs; Phase 25 receipt data is historical and does not evidence Phase 27.

### `examples/phase27/README.md` (documentation, request-response / batch)

**Analog:** `examples/phase24/README.md`

Copy the clean-checkout command section pattern at lines 10-41, followed by an evidence index at lines 43-65. Keep build/run command truth in this one guide; document exact outputs, bounds, failure status/stderr and receipt revision/host provenance. The root README should link here.

### `.github/workflows/ci.yml` (config, batch) and `README.md` (documentation/index)

**Analogs:** `.github/workflows/ci.yml` lines 100-105; `README.md` lines 18-35.

CI currently runs `scripts/verify-phase25.sh` in a dual-host matrix, so add the Phase 27 evidence driver alongside the established aggregate lane. Root README's `Start here` list is the project link pattern; add one link to the Phase 27 guide.

## Shared Patterns

### Independent semantic consumers
**Sources:** `internal/compiler/check/scalar.go`, `internal/compiler/corevalidate/corevalidate.go`, `internal/compiler/originvalidate/originvalidate.go`, `internal/compiler/pathoracle/pathoracle.go`  
**Apply to:** equality, remainder, write operation admission and Unit entry semantics. Each consumer derives its own supported facts; do not accept a checker-only flag as semantic authority.

### App output contract
**Sources:** `internal/compiler/cgen/cgen.go:731-738`, `internal/compiler/session/session_phase26_overflow_test.go:67-76`  
**Apply to:** explicit text/U64/LF chunks and auto-print compatibility. Count app bytes incrementally against one 65,536-byte ceiling. A failure after prior writes may retain the prefix; tests distinguish child status/stderr from tool errors.

### Reproducible host evidence
**Source:** `scripts/verify-phase25.sh:24-63`  
**Apply to:** Phase 27 macOS/Linux native receipts. Record revision, host/target/compiler and declared input hashes; unavailable hosts are unavailable, never passing rows.

## No Analog Found

| File/Concern | Role | Data Flow | Reason |
|---|---|---|---|
| `fizzbuzz.schway` explicit Unit entry and literal-only writes | source witness | streaming output | No current source program combines Unit/no-result entry with explicit source output; adapt the `sum_to_n` structure. |
| Phase 27 independent exact-output evidence | test | request-response | Phase 26 covers returned U64s and failures, not explicit streamed output or FizzBuzz; use it as the closest scaffold. |

## Metadata

**Analog search scope:** `examples/`, `internal/compiler/{syntax,core,check,corevalidate,originvalidate,pathoracle,interp,cgen,session,native}/`, `cmd/schway/`, `scripts/`, `.github/workflows/`  
**Files scanned:** 19 named source analogs/integration points  
**Pattern extraction date:** 2026-10-03
