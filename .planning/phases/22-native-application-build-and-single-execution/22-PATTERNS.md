# Phase 22: Native Application Build and Single Execution - Pattern Map

**Mapped:** 2026-09-27  
**Files analyzed:** 9 inferred implementation files  
**Analogs found:** 7 / 9

Scope is inferred from `22-CONTEXT.md` and `22-RESEARCH.md`: public build/app-run CLI; checked program validation and a retained build seam in session/native; shared C lowering and a U64 process entry; an identity example and local binding manifest; and focused CLI/native/session/emitter tests. Example and manifest paths below are proposed additions, not names locked by the user. Phase 22 research recommendations remain assumptions until the plan selects them.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `cmd/lang/main.go` | controller / CLI | request-response + process streams | `cmd/lang/main.go` command dispatch and handlers | exact |
| `internal/compiler/session/session.go` | service | transform + request-response | `RunNative` and `RunNativeCommandFile` | role-match; conformance flow differs |
| `internal/compiler/native/native.go` | service | process I/O + build | `Runner.Run` and `CompileConformanceUnit` | role-match; no retained executable API yet |
| `internal/compiler/cgen/cgen_program.go` | transform / emitter | transform | `emitProgram` and `emitProgramU64Support` | role-match; app output differs |
| `examples/phase22/identity.lang` | example / fixture | request-response | `testdata/phase19/literal_tracer.lang` (source fixture convention) | partial |
| `examples/phase22/identity.bindings.json` | config / manifest | file-I/O | `testdata/phase16/validation-corpus-run-record.manifest.json` (manifest convention) | partial |
| `cmd/lang/main_test.go` or focused CLI test | test | request-response | `cmd/lang/main_test.go` | role-match |
| `internal/compiler/native/native_app_test.go` | test | process I/O | `internal/compiler/native/native_test.go` | role-match |
| `internal/compiler/cgen/cgen_program_test.go` | test | transform | `internal/compiler/cgen/cgen_program_test.go` | exact |

The seven source/test analog paths named here were checked with `git ls-files`; each is tracked. No mirror/runtime paths are used. `22-RESEARCH.md` mentions malformed/oversized inputs, single launch, relocation, stream separation, evidence status, and independent expected pairs; tests should witness those claims without turning the ordinary app command into replay.

## Pattern Assignments

### `cmd/lang/main.go` (CLI controller, request-response + process streams)

**Analog:** `cmd/lang/main.go`

The existing CLI is a direct argument dispatcher. Preserve the convention of a small route branch calling a focused handler. The application path needs raw stdout/stderr, so it should not funnel child streams through the compiler's `emit` JSON response helper.

**Imports and dispatch** (lines 3-19, 28-100):

```go
func main() { os.Exit(run(os.Args[1:])) }

func run(args []string) int {
    args, jsonMode, ok := extractJSON(args)
    if !ok {
        return emit(usageResult(), true, false)
    }
    // ...existing subcommand dispatch...
    if len(args) == 3 && args[0] == "run" && strings.HasPrefix(args[1], "--engine=") {
        engine := strings.TrimPrefix(args[1], "--engine=")
        switch engine {
        case "interpreter":
            return runInterpreter(args[2], jsonMode)
        case "native":
            return runNative(args[2], jsonMode)
        }
    }
    return emit(usageResult(), jsonMode, false)
}
```

**Handler and error shape** (lines 117-139):

```go
func runNative(path string, jsonMode bool) int {
    result, err := session.RunNativeCommandFile(context.Background(), path, native.DefaultRunner())
    if err != nil {
        return emit(problemResult("run", protocol.StatusOperational, "tool.read_failed", "unable to read input"), jsonMode, false)
    }
    return emit(result, jsonMode, false)
}
```

Use this thin-handler pattern, but let `app run` copy child streams to `os.Stdout`/`os.Stderr` and return a documented child/tool status. Build should likewise delegate parsing/validation/build work to session/native rather than embed it in dispatch.

### `internal/compiler/session/session.go` (service, transform + request-response)

**Analog:** `internal/compiler/session/session.go`

**Checked-core gate and shared native emission** (lines 1037-1068):

```go
func runNative(ctx context.Context, source []byte, runner NativeRunner, fixture string) (NativeResult, []diagnostic.Diagnostic, error) {
    checked := Check(source)
    if len(checked.Diagnostics) > 0 {
        return NativeResult{}, checked.Diagnostics, nil
    }
    validated := corevalidate.Validate(checked.Program)
    if !validated.Valid {
        return NativeResult{}, nil, fmt.Errorf("core validation failed: %s", validated.Problems[0].Code)
    }
    checked.Program = validated.Program()
    entry, err := callgraph.EntryFunction(checked.Program)
    // ...
    cSource, err := Phase16ControlNativeC(checked.Program, fixture)
```

**Conformance-only replay seam** (lines 1053-1093):

```go
inputs, ok := interpreterInputs(checked.Program)
if !ok {
    return NativeResult{}, nil, os.ErrInvalid
}
interpreted := make([]interp.Execution, 0, len(inputs))
for _, input := range inputs {
    execution, err := interp.Run(checked.Program, entry.Name, input)
    // append interpreter result
}
o0, err := runNativeInputs(ctx, runner, cSource, "-O0", inputs, interpreted)
// error check
o3, err := runNativeInputs(ctx, runner, cSource, "-O3", inputs, interpreted)
```

Reuse the check → independent core validation → entry resolution → shared checked C emission sequence. Keep `interpreterInputs`, `interp.Run`, O0/O3, schema projection, and comparisons on an explicit verification/conformance route. The ordinary application build/run path must accept caller input and avoid this replay seam.

### `internal/compiler/native/native.go` (native service, build + process I/O)

**Analog:** `internal/compiler/native/native.go`

**Imports and process policy** (lines 3-25, 27-36):

```go
import (
    "bytes"
    "context"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "syscall"
    "time"
)

const MaxStreamBytes = 64 * 1024
const defaultSubprocessTimeout = 30 * time.Second

type ToolError struct {
    Code string
    Err  error
}
```

**Current temporary compile lifecycle** (lines 457-474, 538-569):

```go
directory, err := os.MkdirTemp("", "lang-native-")
if err != nil {
    return Result{}, &ToolError{Code: "native.temp_failed", Err: err}
}
defer os.RemoveAll(directory)

sourcePath := filepath.Join(directory, "program.c")
binaryPath := filepath.Join(directory, "program")
if err := os.WriteFile(sourcePath, []byte(cSource), 0o600); err != nil {
    return Result{}, &ToolError{Code: "native.temp_failed", Err: err}
}
// Compile with context.WithTimeout and an argv-based exec.CommandContext.
```

This is a useful compiler invocation/error pattern, but the retained build API must publish the final binary outside this temporary directory and return without launching it. Preserve bounded, timed compiler diagnostics and argument vectors; link to a temporary destination then move atomically to the caller's output path only after successful compilation.

**Existing conformance process handling (not the application stream contract)** (lines 581-642):

```go
runCtx, runCancel := context.WithTimeout(parent, r.Timeout)
runCommand := r.commandContext(runCtx, binaryPath, input)
runStdout := boundedWriter{limit: execution.MaxDocumentBytes}
var runStderr boundedWriter
runCommand.Stdout = &runStdout
runCommand.Stderr = &runStderr
runErr := runCommand.Run()
// timeout, truncation, exit/signal adjudication...
if len(runStderr.bytes()) != 0 {
    return Result{}, &ToolError{Code: "native.run_stderr", Err: fmt.Errorf("native process wrote stderr: %s", bounded(string(runStderr.bytes()), 2048))}
}
decoded, decodeErr := decodeExecution(runStdout.bytes(), expect)
```

Use `commandContext`/`exec.CommandContext` and `adjudicateExit` patterns for safe argv, timeout, and signal distinction. App run must use only one child launch, stream ordinary stdout/stderr transparently, and must not apply bounded execution-document decoding to app output. Existing cache identity at lines 156-162 and 202-230 shows content-bound inputs, but also explicitly declines cacheability when toolchain dependency closure is unknown; do not claim complete build identity by hashing only Lang source.

### `internal/compiler/cgen/cgen_program.go` (emitter, transform)

**Analog:** `internal/compiler/cgen/cgen_program.go`

**Shared checked whole-program emitter** (lines 527-558, 564-579):

```go
// emitProgram ... owns includes, typedefs, support blocks, prototypes,
// definitions, and exactly one main invoking the resolved entry function.
func emitProgram(program core.Program, executionJSON bool) (string, error) {
    _ = executionJSON
    order, err := callgraph.Order(program)
    if err != nil {
        return "", err
    }
    entry, err := callgraph.EntryFunction(program)
    if err != nil {
        return "", err
    }
    // Build the shared translation unit from the resolved checked program.
}
```

**U64 ABI/input conversion** (lines 947-955):

```go
static int lang_parse_u64_decimal(const char *text, uint64_t *out) {
  uint64_t value = UINT64_C(0);
  const unsigned char *cursor = (const unsigned char *)text;
  if (cursor == NULL || *cursor == 0u) return 0;
  for (; *cursor != 0u; ++cursor) {
    unsigned int digit;
    if (*cursor < (unsigned char)'0' || *cursor > (unsigned char)'9') return 0;
    digit = (unsigned int)(*cursor - (unsigned char)'0');
    if (value > (UINT64_MAX - digit) / UINT64_C(10)) return 0;
    value = value * UINT64_C(10) + digit;
  }
  *out = value;
  return 1;
}
```

The current entry requires exactly one argument at line 863, invokes the checked entry at 889, and writes `lang.execution/2` JSON at 902-918. Reuse the U64 decimal range guard, but separate application entry/output generation from evidence JSON so normal output is not compiler protocol. Keep both paths sharing checked body lowering as D-22-07 requires; preserve emitter refusal fences for unsupported foreign/pointer families.

### `cmd/lang/main_test.go` (CLI tests, request-response)

**Analog:** `cmd/lang/main_test.go`, lines 8-30, uses narrow dispatch assertions, positive marker, negative sibling, and empty-directory controls. Add CLI coverage for malformed invocation and build/app-run dispatch; for the exactly-once contract, the stronger witness belongs at native process integration level.

### `internal/compiler/native/native_test.go` (process integration tests, process I/O)

**Analog:** `internal/compiler/native/native_test.go`, lines 27-45 and following phase-specific cases, constructs an out-of-corpus Lang source and records shipped-binary arguments plus expected exit/diagnostic. Use temp directories and isolated marker files for retained-artifact relocation and launch counting. Keep expected outputs independently specified (`7 → 7`, `42 → 42`); verify ordinary stdout and successful stderr separately from evidence status.

### Example and manifest files

**Source analog:** `testdata/phase19/literal_tracer.lang` (tracked; inspect the small valid scalar fixture for source syntax when planning).  
**Manifest analog:** `testdata/phase16/validation-corpus-run-record.manifest.json` (tracked; machine-readable manifest convention only). These are partial analogs: no existing public application binding manifest or retained-app sample exists. Manifest parsing, path rules, symbols, ABI facts, and runtime dependencies are Phase 22 additions, not an established code pattern.

## Shared Patterns

### Command execution and failures

**Source:** `internal/compiler/native/native.go` lines 538-569, 607-642, 717-736. Use argv slices, `context.WithTimeout`, typed `ToolError` codes, `errors.As` for `*exec.ExitError`, and `syscall.WaitStatus` for signals. Preserve process nonzero exit, signal, timeout, and launch error as distinct cases. The existing code's strict stderr/JSON policy is specifically conformance evidence policy and does not transfer to ordinary app execution.

### Validation and common lowering

**Source:** `internal/compiler/session/session.go` lines 1037-1068. Keep parsing/checking and independent core validation before emission; call the same checked-body emitter from app build and verification. `runNative`'s fixture-derived input, interpreter, and optimization-tier execution are evidence-only behavior.

### Bounded scalar parsing

**Source:** `internal/compiler/cgen/cgen_program.go` lines 947-955. Decimal U64 parsing rejects empty, non-digit, and overflow input without host-sized conversion. Research recommends an argv token bound (20 decimal digits plus U64 maximum); current C parser has overflow rejection but no explicit token-length cap, so the app interface still needs a stated size limit.

### Build identity limits

**Source:** `internal/compiler/native/native.go` lines 156-162, 202-230. Identity inputs include toolchain facts, target, flags, emitted source, and discovered headers/C. Current dependency closure is host-specific and returns not-cacheable on non-Darwin; that is source-inspected behavior, not evidence that a portable retained-build identity is already solved.

## No Analog Found

| File / capability | Role | Data Flow | Reason |
|---|---|---|---|
| `examples/phase22/identity.bindings.json` | config | file-I/O | No tracked user-facing C binding/build manifest currently exists; only machine-oriented evidence manifests are analogous. |
| Retained native application build API and evidence sidecar status | service/config | build + file-I/O | Existing native runner deletes the artifact and exposes conformance documents; there is no retained standalone artifact or disabled/incomplete/capacity-exhausted app evidence contract. |

## Metadata

**Analog search scope:** `cmd/lang/`, `internal/compiler/{session,native,cgen}/`, `testdata/phase16/`, `testdata/phase19/`, and tracked phase test files.  
**Files scanned in depth:** 7 strong analogs, plus phase context/research and project instructions.  
**Git tracking:** All analog source paths named above have non-empty `git ls-files -- <path>` results.  
**Evidence distinction:** Findings are source inspection only. No tests, builds, or executables were run for this mapping. `22-RESEARCH.md` records historical inspection at `d9bde05` and proposed evidence; this mapping does not upgrade those receipts to newly executed checks or claim Phase 22 implementation.
