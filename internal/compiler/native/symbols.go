package native

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// UndefinedSymbolStatus is the tri-state result of the D-04-19 undefined-
// symbol allowlist control: pass, fail, or operational (the listing tool is
// absent). Per D-04-19, "a tool that is absent is not evidence of absence"
// -- the control must never report pass when it cannot answer the question.
type UndefinedSymbolStatus string

const (
	SymbolsPass        UndefinedSymbolStatus = "pass"
	SymbolsFail        UndefinedSymbolStatus = "fail"
	SymbolsOperational UndefinedSymbolStatus = "operational"
)

// AllowedUndefinedSymbol names one undefined symbol this project's linked
// binaries are permitted to pull in, together with the written rationale
// D-04-19 requires beside it.
type AllowedUndefinedSymbol struct {
	Symbol    string
	Rationale string
}

// AllowedUndefinedSymbols is the checked-in POSITIVE allowlist (D-04-19):
// any undefined symbol a linked binary pulls in that is NOT named here fails
// the control. This is deliberately a positive list, never a denylist --
// adding an entry is a reviewable act with a written rationale, not a
// silent append, and a newly linked undefined symbol fails until a human
// adds one.
var AllowedUndefinedSymbols = []AllowedUndefinedSymbol{
	{Symbol: "malloc", Rationale: "schway_foreign_resource.c's and schway_foreign_nonlocal.c's own acquisition paths allocate via libc malloc (D-04-10)."},
	{Symbol: "free", Rationale: "schway_foreign_resource.c's and schway_foreign_nonlocal.c's own release paths free via libc free (D-04-10)."},
	{Symbol: "longjmp", Rationale: "schway_foreign_nonlocal.c's witness performs a genuine foreign nonlocal exit back into the process-root landing pad (D-04-17)."},
	{Symbol: "abort", Rationale: "the generated defect path and the nonlocal-exit landing pad both terminate via libc abort -- never a caught or contained signal (D-04-15/D-04-18)."},
	{Symbol: "fwrite", Rationale: "the generated streaming and buffered event writers flush JSON output through libc fwrite."},
	{Symbol: "fflush", Rationale: "the generated terminal-defect path flushes its final JSON record before abort so captured stdout is not lost."},
	{Symbol: "snprintf", Rationale: "the generated scalar byte writer formats its decimal encoding through libc snprintf."},
	{Symbol: "__snprintf_chk", Rationale: "on the recorded Apple Clang/libSystem host, an ordinary snprintf call under the platform's default _FORTIFY_SOURCE hardening resolves to this fortified variant even at -O0 -- the same generated call site as snprintf above, not a second one."},
	{Symbol: "strcmp", Rationale: "generated main() dispatches on argv[1] via libc strcmp."},
	{Symbol: "strlen", Rationale: "the generated literal writer measures a C string via libc strlen."},
	{Symbol: "puts", Rationale: "the Phase 1 match emitter's non-JSON and JSON output paths both write their single line via libc puts."},
	{Symbol: "__stdoutp", Rationale: "on the recorded Apple libSystem host, fwrite's stdout argument resolves through this exported FILE* symbol rather than a plain 'stdout' undefined reference -- a platform C-library detail, not project code."},
	{Symbol: "setjmp", Rationale: "the process-root landing pad installs its own return point via libc setjmp (D-04-17); its paired longjmp is named separately below since only the frozen foreign TU ever calls it."},
	{Symbol: "setvbuf", Rationale: "the test-only legacy nonlocal adapter makes byte-frozen terminal output observable across libc abort-buffering behavior before the archived program enters main."},
}

var platformAllowedUndefinedSymbols = map[string][]AllowedUndefinedSymbol{
	"linux": {
		{Symbol: "stdout", Rationale: "on GNU ELF hosts, the generated event writer's stdout reference remains an undefined libc data symbol instead of resolving through Apple's __stdoutp alias."},
		{Symbol: "ITM_deregisterTMCloneTable", Rationale: "GNU ELF startup objects may leave this transactional-memory registration hook undefined when linking a program that does not use transactional memory."},
		{Symbol: "ITM_registerTMCloneTable", Rationale: "GNU ELF startup objects may leave this transactional-memory registration hook undefined when linking a program that does not use transactional memory."},
		{Symbol: "_cxa_finalize", Rationale: "GNU ELF startup code references the C++ finalization hook even for the project's C-only native program; the linker versions it by the host glibc ABI."},
		{Symbol: "_gmon_start__", Rationale: "GNU ELF startup code may retain its optional profiling hook as an undefined symbol on an otherwise unprofiled C program."},
		{Symbol: "_libc_start_main", Rationale: "GNU ELF startup objects reference glibc's process entry wrapper; the linker versions it by the host glibc ABI."},
	},
}

func allowedSymbolSet(goos string) map[string]struct{} {
	set := make(map[string]struct{}, len(AllowedUndefinedSymbols)+len(platformAllowedUndefinedSymbols[goos]))
	for _, allowed := range AllowedUndefinedSymbols {
		set[allowed.Symbol] = struct{}{}
	}
	for _, allowed := range platformAllowedUndefinedSymbols[goos] {
		set[allowed.Symbol] = struct{}{}
	}
	return set
}

// normalizeSymbol drops an optional ELF symbol version and strips exactly
// one leading underscore -- the Mach-O/ELF
// object-format difference this control must not be confused by: Apple's
// linker and nm prefix every C symbol with a leading underscore (Mach-O),
// while ELF (Linux) hosts do not. Stripping at most one leading underscore
// (never every one) keeps a genuinely double-underscore-prefixed reserved
// libc identifier's own leading underscore intact.
func normalizeSymbol(name string) string {
	if version := strings.IndexByte(name, '@'); version >= 0 {
		name = name[:version]
	}
	return strings.TrimPrefix(name, "_")
}

// UndefinedSymbols runs `nm -u` against binaryPath and returns every
// undefined symbol it lists, normalized (D-04-19). It uses the SAME
// bounded, timed os/exec pattern as every other native.go tool invocation:
// its own command, its own boundedWriter pair, its own timeout context,
// never shared with a compile or run invocation.
func UndefinedSymbols(parent context.Context, nmPath, binaryPath string, timeout time.Duration) ([]string, *ToolError) {
	if nmPath == "" {
		nmPath = "nm"
	}
	if timeout <= 0 {
		timeout = defaultSubprocessTimeout
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	command := exec.CommandContext(ctx, nmPath, "-u", binaryPath)
	var stdout, stderr boundedWriter
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return nil, &ToolError{Code: "native.timeout", Err: ctx.Err()}
	}
	if stdout.overflowed() {
		return nil, &ToolError{Code: "native.nm_stdout_truncated", Err: fmt.Errorf("process stream exceeded %d bytes", MaxStreamBytes)}
	}
	if stderr.overflowed() {
		return nil, &ToolError{Code: "native.nm_stderr_truncated", Err: fmt.Errorf("process stream exceeded %d bytes", MaxStreamBytes)}
	}
	if err != nil {
		code := "native.nm_failed"
		if errors.Is(err, exec.ErrNotFound) || errors.Is(err, os.ErrNotExist) {
			code = "native.tool_missing"
		}
		return nil, &ToolError{Code: code, Err: withStderr(err, stderr.bytes())}
	}
	var symbols []string
	for _, line := range strings.Split(string(stdout.bytes()), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		name := fields[len(fields)-1]
		symbols = append(symbols, normalizeSymbol(name))
	}
	return symbols, nil
}

// CheckUndefinedSymbolAllowlist runs UndefinedSymbols against binaryPath and
// compares the normalized result to AllowedUndefinedSymbols (D-04-19). It
// NEVER inspects unwind-section presence -- rejected as a signal because a
// compact-unwind section can be present on a program that never unwinds on
// this platform (04-RESEARCH.md), so a section assertion would pass or fail
// for the wrong reason. On tool absence it reports SymbolsOperational, never
// SymbolsPass: a tool that is absent is not evidence of absence.
func CheckUndefinedSymbolAllowlist(parent context.Context, nmPath, binaryPath string, timeout time.Duration) (UndefinedSymbolStatus, []string, *ToolError) {
	symbols, toolErr := UndefinedSymbols(parent, nmPath, binaryPath, timeout)
	if toolErr != nil {
		return SymbolsOperational, nil, toolErr
	}
	allowed := allowedSymbolSet(runtime.GOOS)
	var rejected []string
	for _, symbol := range symbols {
		if _, ok := allowed[symbol]; !ok {
			rejected = append(rejected, symbol)
		}
	}
	if len(rejected) != 0 {
		return SymbolsFail, rejected, nil
	}
	return SymbolsPass, nil, nil
}

// CompileOnly compiles cSource (plus r.ForeignSources) into a linked binary
// and returns its path WITHOUT executing it, for a caller that needs to
// inspect the linked artifact itself (D-04-19's undefined-symbol allowlist
// control) rather than its runtime behavior. The caller MUST call the
// returned cleanup function when done; on any error the temp directory is
// already removed and cleanup is a no-op. Deliberately duplicates Run's
// compile-only steps rather than sharing code with it, so a change to the
// execute-and-decode path can never accidentally change what this control
// inspects.
func (r Runner) CompileOnly(parent context.Context, cSource, optimization string) (string, func(), error) {
	if r.ClangPath == "" {
		r.ClangPath = "clang"
	}
	if r.Timeout <= 0 {
		r.Timeout = 5 * time.Second
	}
	directory, err := os.MkdirTemp("", "lang-native-symbols-")
	if err != nil {
		return "", func() {}, &ToolError{Code: "native.temp_failed", Err: err}
	}
	cleanup := func() { os.RemoveAll(directory) }

	sourcePath := directory + "/program.c"
	binaryPath := directory + "/program"
	if err := os.WriteFile(sourcePath, []byte(cSource), 0o600); err != nil {
		cleanup()
		return "", func() {}, &ToolError{Code: "native.temp_failed", Err: err}
	}

	objectPaths := make([]string, 0, len(r.ForeignSources))
	for index, foreignSource := range r.ForeignSources {
		objectPath := fmt.Sprintf("%s/foreign_%d.o", directory, index)
		foreignCtx, foreignCancel := context.WithTimeout(parent, r.Timeout)
		foreignCommand := r.commandContext(foreignCtx, r.ClangPath, "-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic", optimization, "-c", foreignSource, "-o", objectPath)
		var foreignStdout, foreignStderr boundedWriter
		foreignCommand.Stdout = &foreignStdout
		foreignCommand.Stderr = &foreignStderr
		foreignErr := foreignCommand.Run()
		deadlineExceeded := errors.Is(foreignCtx.Err(), context.DeadlineExceeded)
		foreignCancel()
		if deadlineExceeded {
			cleanup()
			return "", func() {}, &ToolError{Code: "native.timeout", Err: foreignCtx.Err()}
		}
		if foreignErr != nil {
			cleanup()
			code := "native.compile_failed"
			if errors.Is(foreignErr, exec.ErrNotFound) || errors.Is(foreignErr, os.ErrNotExist) {
				code = "native.tool_missing"
			}
			return "", func() {}, &ToolError{Code: code, Err: withStderr(foreignErr, foreignStderr.bytes())}
		}
		objectPaths = append(objectPaths, objectPath)
	}

	ctx, cancel := context.WithTimeout(parent, r.Timeout)
	defer cancel()
	arguments := append([]string{"-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic", optimization, sourcePath}, objectPaths...)
	arguments = append(arguments, "-o", binaryPath)
	command := r.commandContext(ctx, r.ClangPath, arguments...)
	var stdout, stderr boundedWriter
	command.Stdout = &stdout
	command.Stderr = &stderr
	err = command.Run()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		cleanup()
		return "", func() {}, &ToolError{Code: "native.timeout", Err: ctx.Err()}
	}
	if err != nil {
		cleanup()
		code := "native.compile_failed"
		if errors.Is(err, exec.ErrNotFound) || errors.Is(err, os.ErrNotExist) {
			code = "native.tool_missing"
		}
		return "", func() {}, &ToolError{Code: code, Err: withStderr(err, stderr.bytes())}
	}
	return binaryPath, cleanup, nil
}
