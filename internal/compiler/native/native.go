package native

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/codename-lang/lang/internal/compiler/cache"
	"github.com/codename-lang/lang/internal/compiler/execution"
)

const MaxStreamBytes = 64 * 1024

const defaultSubprocessTimeout = 30 * time.Second

type ToolError struct {
	Code string
	Err  error
}

func (e *ToolError) Error() string { return e.Code + ": " + e.Err.Error() }
func (e *ToolError) Unwrap() error { return e.Err }

type Pair struct {
	Input     string
	Execution execution.Execution
}

type Result struct {
	Optimization string
	Pairs        []Pair
	CompileTime  time.Duration
	RunTime      time.Duration
	OutputBytes  int
	CacheStatus  cache.CacheStatus
}

type Runner struct {
	ClangPath string
	Timeout   time.Duration
	// Expect names the terminal outcome this runner's caller expects every
	// execution document to carry (D-04-08). The zero value behaves exactly
	// as ExpectValue, so every pre-Phase-4 caller that never sets this field
	// keeps the original "returned"-only contract unchanged.
	Expect TerminalOutcome
	// ForeignSources names additional C files (the frozen foreign
	// translation unit, D-04-10) to compile as their own separate, bounded,
	// timed invocations and link into the final binary. Empty for every
	// pre-Phase-4 caller, which keeps the original single-file compile-and-
	// link invocation byte-for-byte unchanged.
	ForeignSources []string
	// BuildCache enables reuse of a content-bound native executable artifact.
	// It never stores executions or comparator outcomes.
	BuildCache *cache.Store
	// LTO gates D-05-19's `-flto` tier: when true, `-flto` is appended to
	// BOTH the per-TU foreign compile argument list and the final link
	// argument list -- never just one, since LTO is inert if either is
	// missing. Gated on this field only: with LTO false, every argument
	// vector this Runner constructs is byte-for-byte identical to the
	// pre-Phase-5 shape, so the existing -O0/-O3 differential is untouched.
	LTO bool
	// recorder, when non-nil (via EnableCommandRecording), captures every
	// constructed command line and the final compiled binary's bytes. This
	// exists so "did -flto reach clang" and "did LTO change codegen" are
	// assertions a test makes directly, never inferences from a successful
	// build (D-05-19/D-05-38). Nil for every production caller.
	recorder *commandRecorder
	command  func(context.Context, string, ...string) *exec.Cmd
	// publishRename is an internal test seam for retained application publication.
	publishRename func(string, string) error
	// evidenceLimit is an internal test seam for capacity-exhaustion controls.
	evidenceLimit int
}

// commandRecorder is a pointer-shared recording sink: Runner is used by
// value throughout this package, so every constructed command line and the
// resulting binary are recorded through this indirection to survive value
// copies of the Runner that holds it.
type commandRecorder struct {
	mu     sync.Mutex
	lines  [][]string
	binary []byte
}

// EnableCommandRecording returns a copy of r with command-line and binary
// recording turned on. Test-only: used exclusively to assert that `-flto`
// reaches both the compile and link command lines, and that the LTO and
// non-LTO tiers produce different codegen (D-05-19/D-05-38).
func (r Runner) EnableCommandRecording() Runner {
	r.recorder = &commandRecorder{}
	return r
}

// LastCommandLines returns every command line recorded by this Runner's
// most recent Run call, in construction order. Returns nil when command
// recording was never enabled via EnableCommandRecording.
func (r Runner) LastCommandLines() [][]string {
	if r.recorder == nil {
		return nil
	}
	r.recorder.mu.Lock()
	defer r.recorder.mu.Unlock()
	return append([][]string(nil), r.recorder.lines...)
}

// LastBinary returns the compiled binary's bytes from this Runner's most
// recent Run call. Returns nil when command recording was never enabled via
// EnableCommandRecording.
func (r Runner) LastBinary() []byte {
	if r.recorder == nil {
		return nil
	}
	r.recorder.mu.Lock()
	defer r.recorder.mu.Unlock()
	return append([]byte(nil), r.recorder.binary...)
}

func (r Runner) recordCommandLine(name string, arguments []string) {
	if r.recorder == nil {
		return
	}
	line := append([]string{name}, arguments...)
	r.recorder.mu.Lock()
	r.recorder.lines = append(r.recorder.lines, line)
	r.recorder.mu.Unlock()
}

func (r Runner) recordBinary(path string) {
	if r.recorder == nil {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	r.recorder.mu.Lock()
	r.recorder.binary = data
	r.recorder.mu.Unlock()
}

func DefaultRunner() Runner { return Runner{ClangPath: "clang", Timeout: defaultSubprocessTimeout} }

var clangTokenPattern = regexp.MustCompile(`"(?:[^"\\]|\\.)*"|\S+`)

// cacheInputs discovers byte-level compiler dependencies before permitting a
// lookup. Anything this deliberately narrow Darwin toolchain probe cannot
// resolve is not cacheable; callers still compile and execute normally.
func (r Runner) cacheInputs(ctx context.Context, source, optimization string, foreign []string) ([]cache.Input, bool) {
	if runtime.GOOS != "darwin" {
		return nil, false // linker and SDK dependency discovery is host-specific
	}
	clang, err := exec.LookPath(r.ClangPath)
	if err != nil {
		return nil, false
	}
	clang, err = filepath.Abs(clang)
	if err != nil {
		return nil, false
	}
	clangBytes, err := os.ReadFile(clang)
	if err != nil {
		return nil, false
	}
	version, ok := r.probe(ctx, clang, "--version")
	if !ok {
		return nil, false
	}
	target, ok := r.probe(ctx, clang, "-dumpmachine")
	if !ok || strings.TrimSpace(target) == "" {
		return nil, false
	}

	temp, err := os.MkdirTemp("", "lang-native-deps-")
	if err != nil {
		return nil, false
	}
	defer os.RemoveAll(temp)
	main := filepath.Join(temp, "program.c")
	if err := os.WriteFile(main, []byte(source), 0o600); err != nil {
		return nil, false
	}
	files := []string{main}
	for _, path := range foreign {
		files = append(files, path)
	}

	flags := []string{"-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic", optimization}
	if r.LTO {
		flags = append(flags, "-flto")
	}
	inputs := []cache.Input{
		{Name: "compiler", Digest: digestBytes(append([]byte(clang+"\x00"), clangBytes...))},
		{Name: "compiler-version", Digest: digestBytes([]byte(version))},
		{Name: "target-triple", Digest: digestBytes([]byte(strings.TrimSpace(target)))},
		{Name: "host-target", Digest: digestBytes([]byte(runtime.GOOS + "/" + runtime.GOARCH))},
		{Name: "optimization-and-flags", Digest: digestBytes([]byte(strings.Join(flags, "\x00")))},
		{Name: "cache-schema", Digest: digestBytes([]byte(cache.Schema))},
		{Name: "emitted-c", Digest: digestBytes([]byte(source))},
	}
	for i, file := range files {
		depArgs := append([]string{"-M", "-MT", "lang-deps"}, flags...)
		depArgs = append(depArgs, file)
		depOutput, ok := r.probe(ctx, clang, depArgs...)
		if !ok {
			return nil, false
		}
		deps, ok := parseMakeDependencies(depOutput)
		if !ok {
			return nil, false
		}
		for j, dependency := range deps {
			if filepath.Clean(dependency) == filepath.Clean(file) {
				continue
			}
			data, readErr := os.ReadFile(dependency)
			if readErr != nil {
				return nil, false
			}
			inputs = append(inputs, cache.Input{Name: fmt.Sprintf("translation-unit-%d-header-%d:%s", i, j, dependency), Digest: digestBytes(data)})
		}
		if i > 0 {
			data, readErr := os.ReadFile(foreign[i-1])
			if readErr != nil {
				return nil, false
			}
			inputs = append(inputs, cache.Input{Name: fmt.Sprintf("foreign-source-%d", i-1), Digest: digestBytes(data)})
		}
	}

	// Ask the installed driver for its real link invocation. Hash every
	// resolved executable/library and each SDK search root identity; unknown
	// -l/-framework inputs force a fresh build.
	linkArgs := append([]string{"-###"}, flags...)
	linkArgs = append(linkArgs, main, "-o", filepath.Join(temp, "program"))
	linkOutput, ok := r.probe(ctx, clang, linkArgs...)
	if !ok {
		return nil, false
	}
	normalizedDriver, ok := normalizeDriverCommands(linkOutput, main)
	if !ok {
		return nil, false
	}
	inputs = append(inputs, cache.Input{Name: "normalized-clang-driver-commands", Digest: digestBytes([]byte(normalizedDriver))})
	words := clangTokenPattern.FindAllString(linkOutput, -1)
	var sysroot string
	search := []string{}
	var paths []string
	var libraries []string
	for i := 0; i < len(words); i++ {
		word := unquoteClangToken(words[i])
		switch {
		case word == "-framework" || strings.HasPrefix(word, "-framework"):
			return nil, false
		case word == "-lto_library":
			if i+1 >= len(words) {
				return nil, false
			}
			i++
			libraryPath := unquoteClangToken(words[i])
			if !filepath.IsAbs(libraryPath) {
				return nil, false
			}
			paths = append(paths, libraryPath)
		case word == "-syslibroot" || word == "-isysroot":
			if i+1 >= len(words) {
				return nil, false
			}
			i++
			sysroot = unquoteClangToken(words[i])
			search = append(search, sysroot)
		case strings.HasPrefix(word, "-L") && len(word) > 2:
			search = append(search, strings.TrimPrefix(word, "-L"))
		case strings.HasPrefix(word, "-l") && len(word) > 2:
			libraries = append(libraries, strings.TrimPrefix(word, "-l"))
		case filepath.IsAbs(word):
			if info, statErr := os.Stat(word); statErr == nil && info.Mode().IsRegular() && word != main && word != filepath.Join(temp, "program") {
				paths = append(paths, word)
			}
		}
	}
	if sysroot == "" || len(search) == 0 {
		return nil, false
	}
	sdkSettings, err := os.ReadFile(filepath.Join(sysroot, "SDKSettings.json"))
	if err != nil {
		return nil, false
	}
	inputs = append(inputs, cache.Input{Name: "sdk-settings:" + filepath.Join(sysroot, "SDKSettings.json"), Digest: digestBytes(sdkSettings)})
	for _, library := range libraries {
		if library != "System" {
			return nil, false
		}
		// Search explicit roots in their emitted order, then the SDK's
		// system-library root. Hash each possible name's presence as well
		// as bytes: a newly-created higher-priority candidate must turn a
		// prior key into a miss even if the previously resolved library is
		// unchanged.
		roots := append(append([]string(nil), search...), filepath.Join(sysroot, "usr", "lib"))
		resolved := false
		for rootIndex, root := range roots {
			for _, suffix := range []string{".dylib", ".tbd", ".a"} {
				candidate := filepath.Join(root, "libSystem"+suffix)
				data, readErr := os.ReadFile(candidate)
				present := readErr == nil
				if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
					return nil, false
				}
				marker := "absent"
				if present {
					marker = "present:" + digestBytes(data)
					paths = append(paths, candidate)
					if !resolved {
						resolved = true
					}
				}
				inputs = append(inputs, cache.Input{
					Name:   fmt.Sprintf("link-resolution-candidate-%d:%s", rootIndex, candidate),
					Digest: digestBytes([]byte(marker)),
				})
			}
		}
		if !resolved {
			return nil, false
		}
	}
	for i, root := range search {
		inputs = append(inputs, cache.Input{Name: fmt.Sprintf("ordered-link-search-%d", i), Digest: digestBytes([]byte(root))})
	}
	seen := map[string]bool{}
	for i, path := range paths {
		if seen[path] {
			continue
		}
		seen[path] = true
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil, false
		}
		inputs = append(inputs, cache.Input{Name: fmt.Sprintf("link-input-%d:%s", i, path), Digest: digestBytes(data)})
	}
	return inputs, true
}

func normalizeDriverCommands(output, sourcePath string) (string, bool) {
	var compile, link []string
	for _, line := range strings.Split(output, "\n") {
		raw := clangTokenPattern.FindAllString(line, -1)
		if len(raw) == 0 {
			continue
		}
		tokens := make([]string, len(raw))
		for i, token := range raw {
			tokens[i] = unquoteClangToken(token)
		}
		isCompile, isLink := false, false
		for _, token := range tokens {
			if token == "-cc1" {
				isCompile = true
			}
			if filepath.Base(token) == "ld" {
				isLink = true
			}
		}
		if !isCompile && !isLink {
			continue
		}
		filtered := make([]string, 0, len(tokens))
		for i := 0; i < len(tokens); i++ {
			token := tokens[i]
			if token == "-o" || token == "-dumpdir" {
				i++
				if i >= len(tokens) {
					return "", false
				}
				continue
			}
			if token == sourcePath || strings.HasSuffix(token, ".o") {
				continue
			}
			filtered = append(filtered, token)
		}
		if isCompile {
			compile = filtered
		}
		if isLink {
			link = filtered
		}
	}
	if len(compile) == 0 || len(link) == 0 {
		return "", false
	}
	return "compile\x00" + strings.Join(compile, "\x00") + "\x00link\x00" + strings.Join(link, "\x00"), true
}

func (r Runner) probe(parent context.Context, name string, args ...string) (string, bool) {
	timeout := r.Timeout
	if timeout <= 0 {
		timeout = defaultSubprocessTimeout
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	command := r.commandContext(ctx, name, args...)
	var stdout, stderr boundedWriter
	command.Stdout, command.Stderr = &stdout, &stderr
	if command.Run() != nil || ctx.Err() != nil || stdout.overflowed() || stderr.overflowed() {
		return "", false
	}
	return string(stdout.bytes()) + string(stderr.bytes()), true
}

func parseMakeDependencies(output string) ([]string, bool) {
	output = strings.ReplaceAll(output, "\\\n", " ")
	colon := strings.IndexByte(output, ':')
	if colon < 0 {
		return nil, false
	}
	text := strings.TrimSpace(output[colon+1:])
	if text == "" {
		return nil, false
	}
	// Clang's escaped path syntax is intentionally rejected unless paths are
	// plain whitespace-free tokens. A partial dependency parse is unsafe.
	for _, r := range text {
		if r == '\\' {
			return nil, false
		}
	}
	deps := strings.Fields(text)
	if len(deps) == 0 {
		return nil, false
	}
	return deps, true
}

func unquoteClangToken(token string) string {
	if strings.HasPrefix(token, "\"") {
		if value, err := strconv.Unquote(token); err == nil {
			return value
		}
	}
	return token
}

func digestBytes(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }

func (r Runner) Run(parent context.Context, cSource, optimization string, inputs []string) (Result, error) {
	if r.ClangPath == "" {
		r.ClangPath = "clang"
	}
	if r.Timeout <= 0 {
		r.Timeout = defaultSubprocessTimeout
	}
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

	cacheStatus := cache.StatusNotCacheable
	var cacheKey cache.Key
	if r.BuildCache != nil {
		if declared, complete := r.cacheInputs(parent, cSource, optimization, r.ForeignSources); complete {
			cacheKey, err = cache.ComputeKey(declared)
			if err == nil {
				cacheStatus = cache.StatusArtifactRecomputed
				if artifact, found, getErr := r.BuildCache.Get(cacheKey); getErr != nil {
					cacheStatus = cache.StatusUnavailable
				} else if found {
					if writeErr := os.WriteFile(binaryPath, artifact, 0o700); writeErr == nil {
						cacheStatus = cache.StatusArtifactReused
					} else {
						cacheStatus = cache.StatusUnavailable
					}
				}
			}
		}
	}

	// Per D-04-10, the frozen foreign translation unit is compiled as its
	// own separate, bounded, timed invocation -- never merged into one
	// clang invocation with program.c's own compile -- and only the
	// resulting object is linked into the final binary below.
	objectPaths := make([]string, 0, len(r.ForeignSources))
	var compileTime time.Duration
	if cacheStatus != cache.StatusArtifactReused {
		for index, foreignSource := range r.ForeignSources {
			objectPath := filepath.Join(directory, fmt.Sprintf("foreign_%d.o", index))
			foreignCtx, foreignCancel := context.WithTimeout(parent, r.Timeout)
			foreignArguments := []string{"-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic", optimization}
			if r.LTO {
				foreignArguments = append(foreignArguments, "-flto")
			}
			foreignArguments = append(foreignArguments, "-c", foreignSource, "-o", objectPath)
			foreignCommand := r.commandContext(foreignCtx, r.ClangPath, foreignArguments...)
			r.recordCommandLine(r.ClangPath, foreignArguments)
			var foreignStdout, foreignStderr boundedWriter
			foreignCommand.Stdout = &foreignStdout
			foreignCommand.Stderr = &foreignStderr
			foreignErr := foreignCommand.Run()
			deadlineExceeded := errors.Is(foreignCtx.Err(), context.DeadlineExceeded)
			foreignCancel()
			if deadlineExceeded {
				return Result{}, &ToolError{Code: "native.timeout", Err: foreignCtx.Err()}
			}
			if foreignStdout.overflowed() {
				return Result{}, streamError("native.compile_stdout_truncated")
			}
			if foreignStderr.overflowed() {
				return Result{}, streamError("native.compile_stderr_truncated")
			}
			if foreignErr != nil {
				code := "native.compile_failed"
				if errors.Is(foreignErr, exec.ErrNotFound) || errors.Is(foreignErr, os.ErrNotExist) {
					code = "native.tool_missing"
				}
				return Result{}, &ToolError{Code: code, Err: withStderr(foreignErr, foreignStderr.bytes())}
			}
			objectPaths = append(objectPaths, objectPath)
		}

		ctx, cancel := context.WithTimeout(parent, r.Timeout)
		defer cancel()
		started := time.Now()
		arguments := []string{"-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic", optimization}
		if r.LTO {
			arguments = append(arguments, "-flto")
		}
		arguments = append(arguments, sourcePath)
		arguments = append(arguments, objectPaths...)
		arguments = append(arguments, "-o", binaryPath)
		command := r.commandContext(ctx, r.ClangPath, arguments...)
		r.recordCommandLine(r.ClangPath, arguments)
		var compileStdout, compileStderr boundedWriter
		command.Stdout = &compileStdout
		command.Stderr = &compileStderr
		err = command.Run()
		compileTime = time.Since(started)
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return Result{}, &ToolError{Code: "native.timeout", Err: ctx.Err()}
		}
		if compileStdout.overflowed() {
			return Result{}, streamError("native.compile_stdout_truncated")
		}
		if compileStderr.overflowed() {
			return Result{}, streamError("native.compile_stderr_truncated")
		}
		if err != nil {
			code := "native.compile_failed"
			if errors.Is(err, exec.ErrNotFound) || errors.Is(err, os.ErrNotExist) {
				code = "native.tool_missing"
			}
			return Result{}, &ToolError{Code: code, Err: withStderr(err, compileStderr.bytes())}
		}
		if cacheStatus == cache.StatusArtifactRecomputed {
			if artifact, readErr := os.ReadFile(binaryPath); readErr == nil {
				if putErr := r.BuildCache.Put(cacheKey, artifact); putErr != nil {
					cacheStatus = cache.StatusUnavailable
				}
			}
		}
	}
	r.recordBinary(binaryPath)

	result := Result{Optimization: optimization, CompileTime: compileTime, CacheStatus: cacheStatus, Pairs: make([]Pair, 0, len(inputs))}
	for _, input := range inputs {
		runStarted := time.Now()
		runCtx, runCancel := context.WithTimeout(parent, r.Timeout)
		runCommand := r.commandContext(runCtx, binaryPath, input)
		runStdout := boundedWriter{limit: execution.MaxDocumentBytes}
		var runStderr boundedWriter
		runCommand.Stdout = &runStdout
		runCommand.Stderr = &runStderr
		runErr := runCommand.Run()
		runCancel()
		result.RunTime += time.Since(runStarted)
		result.OutputBytes += runStdout.total + runStderr.total
		if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
			return Result{}, &ToolError{Code: "native.timeout", Err: runCtx.Err()}
		}
		if runStdout.overflowed() {
			return Result{}, executionStreamError("native.run_stdout_truncated")
		}
		if runStderr.overflowed() {
			return Result{}, streamError("native.run_stderr_truncated")
		}
		expect := r.Expect
		if expect == "" {
			expect = ExpectValue
		}
		if runErr != nil {
			// D-04-24: adjudicate a nonzero exit through
			// ProcessState.Sys().(syscall.WaitStatus) -- never a hardcoded
			// exit code, which is a shell's encoding and this codebase
			// correctly uses no shell. A signalled SIGABRT termination while
			// the caller expects a defect terminal outcome is D-04-15's
			// expected abort-only shape, not a failure: fall through and
			// decode whatever the aborting process wrote before it died.
			adjudication, known := adjudicateExit(runErr)
			expectedAbort := known && adjudication.Signaled && adjudication.Signal == syscall.SIGABRT && expect == ExpectDefect
			if !expectedAbort {
				if known && adjudication.Signaled {
					return Result{}, &ToolError{Code: "native.run_signaled", Err: fmt.Errorf("process terminated by signal %v: %w", adjudication.Signal, withStderr(runErr, runStderr.bytes()))}
				}
				return Result{}, &ToolError{Code: "native.run_failed", Err: withStderr(runErr, runStderr.bytes())}
			}
			// An expected SIGABRT termination is not required to have
			// written nothing to stderr the way an ordinary clean exit is --
			// the stderr-must-be-empty gate below exists to catch a
			// SUCCESSFUL-looking run that quietly also wrote diagnostics,
			// which does not apply to a process that just died.
			decoded, decodeErr := decodeExecution(runStdout.bytes(), expect)
			if decodeErr != nil {
				return Result{}, decodeErr
			}
			result.Pairs = append(result.Pairs, Pair{Input: input, Execution: decoded})
			continue
		}
		if len(runStderr.bytes()) != 0 {
			return Result{}, &ToolError{Code: "native.run_stderr", Err: fmt.Errorf("native process wrote stderr: %s", bounded(string(runStderr.bytes()), 2048))}
		}
		decoded, decodeErr := decodeExecution(runStdout.bytes(), expect)
		if decodeErr != nil {
			return Result{}, decodeErr
		}
		result.Pairs = append(result.Pairs, Pair{Input: input, Execution: decoded})
	}
	return result, nil
}

// CompileConformanceUnit compiles source (a generated conformance
// translation unit, D-04-11) as its own separate, bounded, timed
// invocation -- its own command, its own boundedWriter pair, its own timeout
// context, never merged with a program compile or the frozen foreign TU's
// own compile. It compiles with "-c" (compile only, never link), since the
// conformance unit defines no symbol and must never reach the linker. A
// compile failure (including a _Static_assert refusal under -Werror)
// reports the distinct "native.conformance_failed" code, so it is never
// mistaken for an ordinary program compile failure ("native.compile_failed").
func (r Runner) CompileConformanceUnit(parent context.Context, source string) error {
	if r.ClangPath == "" {
		r.ClangPath = "clang"
	}
	if r.Timeout <= 0 {
		r.Timeout = defaultSubprocessTimeout
	}
	directory, err := os.MkdirTemp("", "lang-conformance-")
	if err != nil {
		return &ToolError{Code: "native.temp_failed", Err: err}
	}
	defer os.RemoveAll(directory)

	sourcePath := filepath.Join(directory, "lang_foreign_conformance.c")
	objectPath := filepath.Join(directory, "lang_foreign_conformance.o")
	if err := os.WriteFile(sourcePath, []byte(source), 0o600); err != nil {
		return &ToolError{Code: "native.temp_failed", Err: err}
	}

	ctx, cancel := context.WithTimeout(parent, r.Timeout)
	defer cancel()
	command := r.commandContext(ctx, r.ClangPath, "-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic", "-c", sourcePath, "-o", objectPath)
	var stdout, stderr boundedWriter
	command.Stdout = &stdout
	command.Stderr = &stderr
	runErr := command.Run()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return &ToolError{Code: "native.timeout", Err: ctx.Err()}
	}
	if stdout.overflowed() {
		return streamError("native.compile_stdout_truncated")
	}
	if stderr.overflowed() {
		return streamError("native.compile_stderr_truncated")
	}
	if runErr != nil {
		code := "native.conformance_failed"
		if errors.Is(runErr, exec.ErrNotFound) || errors.Is(runErr, os.ErrNotExist) {
			code = "native.tool_missing"
		}
		return &ToolError{Code: code, Err: withStderr(runErr, stderr.bytes())}
	}
	return nil
}

// exitAdjudication is D-04-24's signal-aware verdict: a clean exit, a
// nonzero exit with no signal, and a signalled termination are three
// distinct, mutually exclusive outcomes, decided ONLY through
// ProcessState.Sys().(syscall.WaitStatus) -- never a hardcoded numeric exit
// code, which is a shell's encoding and this codebase correctly uses no
// shell.
type exitAdjudication struct {
	Signaled bool
	Signal   syscall.Signal
}

// adjudicateExit inspects a *exec.ExitError's real wait status. known is
// false only when err is not an *exec.ExitError, or the platform's
// ProcessState.Sys() is not a syscall.WaitStatus (never true on this
// project's supported hosts) -- callers must treat that as an ordinary
// run failure, never as an adjudicated signal.
func adjudicateExit(err error) (exitAdjudication, bool) {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return exitAdjudication{}, false
	}
	status, ok := exitErr.ProcessState.Sys().(syscall.WaitStatus)
	if !ok {
		return exitAdjudication{}, false
	}
	if status.Signaled() {
		return exitAdjudication{Signaled: true, Signal: status.Signal()}, true
	}
	return exitAdjudication{}, true
}

func (r Runner) commandContext(ctx context.Context, name string, arguments ...string) *exec.Cmd {
	if r.command != nil {
		return r.command(ctx, name, arguments...)
	}
	return exec.CommandContext(ctx, name, arguments...)
}

type boundedWriter struct {
	buffer bytes.Buffer
	total  int
	limit  int
}

func (w *boundedWriter) byteLimit() int {
	if w.limit > 0 {
		return w.limit
	}
	return MaxStreamBytes
}

func (w *boundedWriter) Write(data []byte) (int, error) {
	w.total += len(data)
	remaining := w.byteLimit() + 1 - w.buffer.Len()
	if remaining > 0 {
		if remaining > len(data) {
			remaining = len(data)
		}
		_, _ = w.buffer.Write(data[:remaining])
	}
	return len(data), nil
}

func (w *boundedWriter) bytes() []byte    { return w.buffer.Bytes() }
func (w *boundedWriter) overflowed() bool { return w.total > w.byteLimit() }

func streamError(code string) error {
	return &ToolError{Code: code, Err: fmt.Errorf("process stream exceeded %d bytes", MaxStreamBytes)}
}

func executionStreamError(code string) error {
	return &ToolError{Code: code, Err: fmt.Errorf("execution document exceeded %d bytes", execution.MaxDocumentBytes)}
}

func withStderr(err error, stderr []byte) error {
	if len(stderr) == 0 {
		return err
	}
	return fmt.Errorf("%w: %s", err, bounded(string(stderr), 2048))
}

func decodeExecution(stdout []byte, expect TerminalOutcome) (execution.Execution, error) {
	if len(stdout) > execution.MaxDocumentBytes {
		return execution.Execution{}, executionStreamError("native.run_stdout_truncated")
	}
	if err := rejectDuplicateJSONKeys(stdout); err != nil {
		if isUnterminatedJSON(err) {
			return execution.Execution{}, terminalRecordAbsentError(stdout, err)
		}
		return execution.Execution{}, &ToolError{Code: "native.invalid_execution", Err: err}
	}
	var value execution.Execution
	decoder := json.NewDecoder(bytes.NewReader(stdout))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		if isUnterminatedJSON(err) {
			// D-04-20: a run whose stdout stops before a terminal record was
			// ever written (the process died mid-stream) is a HARD FAILURE,
			// never a tolerated truncation -- distinct from
			// native.run_stdout_truncated, which fires only when the
			// existing output cap was actually exceeded above. Tolerating
			// this as an ordinary truncation would make a missing release
			// event on an aborting path indistinguishable from a capped one.
			return execution.Execution{}, terminalRecordAbsentError(stdout, err)
		}
		return execution.Execution{}, &ToolError{Code: "native.invalid_execution", Err: err}
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			err = errors.New("trailing execution document")
		}
		return execution.Execution{}, &ToolError{Code: "native.trailing_execution", Err: err}
	}
	if err := validateExecution(value, expect); err != nil {
		return execution.Execution{}, &ToolError{Code: "native.invalid_execution", Err: err}
	}
	return value, nil
}

// isUnterminatedJSON reports whether err indicates the input ended before a
// value finished decoding -- the shape a process that died mid-write leaves
// behind -- as opposed to a complete-but-malformed document.
func isUnterminatedJSON(err error) bool {
	return errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF)
}

// terminalRecordAbsentError is D-04-20's distinct hard-failure code for a
// stdout stream that ended before its terminal record was written. Embeds
// the captured bytes (bounded) so a caller can confirm any events that DID
// stream through before the process died are not silently lost.
func terminalRecordAbsentError(stdout []byte, err error) error {
	return &ToolError{Code: "native.terminal_record_absent", Err: fmt.Errorf("stdout ended before a terminal record was written (%d bytes captured: %s): %w", len(stdout), bounded(string(stdout), 512), err)}
}

// TerminalOutcome names the closed axis validateExecution accepts (D-04-08:
// value | typed_failure | defect, with "defect" reserved for a later plan --
// see Pitfall 3, 04-RESEARCH.md). It exists so a caller states in advance
// which terminal shape it expects, rather than validateExecution silently
// discovering the shape from the document -- an execution document is
// validated AGAINST an expectation, never used to infer one, matching the
// project's general "never trust the producer" posture.
type TerminalOutcome string

const (
	ExpectValue        TerminalOutcome = "value"
	ExpectTypedFailure TerminalOutcome = "typed_failure"
	// ExpectDefect is Phase 4 plan 04's addition (D-04-15): a defect-
	// expecting call's process is expected to terminate via SIGABRT, and its
	// terminal record carries outcome kind "defect", no value, and every
	// still-live acquisition at the moment of the defect.
	ExpectDefect TerminalOutcome = "defect"
)

func rejectDuplicateJSONKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var scan func() error
	scan = func() error {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		delimiter, ok := token.(json.Delim)
		if !ok {
			return nil
		}
		switch delimiter {
		case '{':
			seen := make(map[string]struct{})
			for decoder.More() {
				keyToken, keyErr := decoder.Token()
				if keyErr != nil {
					return keyErr
				}
				key, ok := keyToken.(string)
				if !ok {
					return errors.New("object key is not a string")
				}
				if _, duplicate := seen[key]; duplicate {
					return fmt.Errorf("duplicate JSON key %q", key)
				}
				seen[key] = struct{}{}
				if err := scan(); err != nil {
					return err
				}
			}
			_, err = decoder.Token()
			return err
		case '[':
			for decoder.More() {
				if err := scan(); err != nil {
					return err
				}
			}
			_, err = decoder.Token()
			return err
		default:
			return errors.New("unexpected JSON delimiter")
		}
	}
	return scan()
}

// validateExecution asserts the execution document against the caller's
// declared expectation (D-04-08's closed value|typed_failure|defect axis;
// "cancelled" is reserved and unconstructible -- no expectation ever admits
// it, and a document claiming it is rejected by the default case below).
// Per Pitfall 3 (04-RESEARCH.md), the pre-Phase-4 "value" contract is kept
// byte-for-byte: every one of its five original rejection grounds still
// rejects (TestValidateExecutionStillRejectsOldGrounds), this function
// simply no longer treats every OTHER expectation as automatically invalid.
func validateExecution(value execution.Execution, expect TerminalOutcome) error {
	if value.Schema != execution.Schema0 && value.Schema != execution.Schema1 && value.Schema != execution.Schema2 {
		return errors.New("unsupported execution schema")
	}
	switch expect {
	case ExpectValue, "":
		if value.Outcome.Kind != "returned" || value.Outcome.Value == "" || value.Events == nil || value.LiveResources == nil || len(value.Events) == 0 || len(value.LiveResources) != 0 {
			return errors.New("execution document violates required contract")
		}
	case ExpectTypedFailure:
		if value.Outcome.Kind != "typed_failure" || value.Outcome.Value == "" || value.Events == nil || value.LiveResources == nil || len(value.Events) == 0 {
			return errors.New("execution document violates required contract")
		}
	case ExpectDefect:
		if value.Outcome.Kind != "defect" || value.Outcome.Value != "" || value.Events == nil || value.LiveResources == nil || len(value.Events) == 0 {
			return errors.New("execution document violates required contract")
		}
	default:
		return fmt.Errorf("unsupported expected terminal outcome %q", expect)
	}
	seenIDs := make(map[string]struct{}, len(value.Events))
	seenInvocationIDs := make(map[string]struct{}, len(value.Events))
	for index, event := range value.Events {
		// D-11-05/Phase 11: a multi-function execution's own events span
		// more than one function -- a callee's own function.returned event
		// is recorded under the CALLEE's function ID, never the caller's
		// (interp.terminalOutcome's identical rule, D-10-32) -- so this no
		// longer requires every event to share ONE FunctionID with the
		// first event. Every pre-Phase-11 single-function execution still
		// has exactly one function ID across every event by construction
		// (there is nothing else to attribute an event to), so this widening
		// changes no existing document's admissibility.
		if event.Schema != value.Schema || event.ID == "" || event.FunctionID == "" {
			return errors.New("execution event identity or schema mismatch")
		}
		if _, duplicate := seenIDs[event.ID]; duplicate {
			if value.Schema != execution.Schema2 {
				return errors.New("duplicate execution event id")
			}
		}
		seenIDs[event.ID] = struct{}{}
		if value.Schema == execution.Schema2 {
			if _, err := execution.ParseInvocation(event.Invocation); err != nil {
				return errors.New("execution event invocation is invalid")
			}
			if event.Kind != "function.called" && event.CalleeFunctionID != "" {
				return errors.New("callee function ID is reserved for function called events")
			}
			key := event.Invocation + "\x00" + event.ID
			if _, duplicate := seenInvocationIDs[key]; duplicate {
				return errors.New("duplicate execution invocation and event id")
			}
			seenInvocationIDs[key] = struct{}{}
		}
		isLast := index == len(value.Events)-1
		switch event.Kind {
		case "function.returned":
			// D-11-05/Phase 11: a multi-function execution pops more than
			// one frame, and interp.terminalOutcome (D-10-32) records a
			// "function.returned" event for EVERY popped frame, not only
			// the outermost one -- an intermediate callee's own return is
			// therefore no longer required to be the document's last
			// event. The outermost (base) frame's own terminal event is
			// still always the actual last event written (cgen's `main`
			// and interp's runFrameStack both write/return only after
			// every callee frame has already popped), so this widening
			// changes no existing single-function document's
			// admissibility -- there, the sole function.returned event was
			// already last by construction.
			if event.TargetPlace != "" {
				return errors.New("return event must have no target")
			}
			if value.Schema == execution.Schema0 {
				if event.Input == "" || event.Output == "" || event.SourcePlace != "" || event.TypeID != "" {
					return errors.New("match return event fields are invalid")
				}
			} else if event.SourcePlace == "" || event.TypeID == "" || event.Input != "" || event.Output != "" {
				return errors.New("linear return event fields are invalid")
			}
		case "value.copied", "value.transferred", "value.borrowed", "value.borrowed_exclusive", "foreign.called", "value.payload_constructed", "value.payload_destructured":
			// value.payload_constructed/value.payload_destructured are
			// Phase 12's own linear transition events (D-12-05): shaped
			// identically to value.copied/value.transferred (a non-terminal
			// transition naming its own source/target places and type).
			if (value.Schema != execution.Schema1 && value.Schema != execution.Schema2) || isLast || event.SourcePlace == "" || event.TargetPlace == "" || event.TypeID == "" || event.Input != "" || event.Output != "" {
				return errors.New("linear transition event fields are invalid")
			}
		case "function.called":
			if value.Schema != execution.Schema2 || isLast || event.CalleeFunctionID == "" || event.SourcePlace == "" || event.TargetPlace == "" || event.TypeID == "" || event.Input != "" || event.Output != "" {
				return errors.New("function called event fields are invalid")
			}
		case "function.failed":
			if !isLast || (value.Schema != execution.Schema1 && value.Schema != execution.Schema2) || event.SourcePlace == "" || event.TypeID == "" || event.Input != "" || event.Output != "" {
				return errors.New("typed-failure event fields are invalid")
			}
		case "function.defected":
			// D-04-15's terminal event: like function.failed it must be
			// last, but it carries its required non-empty reason string in
			// Output rather than leaving it empty.
			if !isLast || (value.Schema != execution.Schema1 && value.Schema != execution.Schema2) || event.SourcePlace == "" || event.TypeID == "" || event.Input != "" || event.Output == "" {
				return errors.New("defect event fields are invalid")
			}
		case "foreign.nonlocal_exit":
			// D-04-17: the process-root landing pad's own first event on the
			// nonlocal-exit path. Never last (a function.defected terminator
			// always follows) and carries no place/type facts of its own --
			// it is a process-level event, not a per-value transition.
			if (value.Schema != execution.Schema1 && value.Schema != execution.Schema2) || isLast || event.SourcePlace != "" || event.TargetPlace != "" || event.TypeID != "" || event.Input != "" || event.Output != "" {
				return errors.New("nonlocal exit event fields are invalid")
			}
		case "resource.leaked":
			// D-04-17: one per still-live acquisition the landing pad found
			// undischarged. Never last (a function.defected terminator always
			// follows) and carries the acquisition's own place in
			// SourcePlace, mirroring resource.released's shape but naming a
			// leak rather than a discharge.
			if (value.Schema != execution.Schema1 && value.Schema != execution.Schema2) || isLast || event.SourcePlace == "" || event.TargetPlace != "" || event.Input != "" || event.Output != "" {
				return errors.New("resource leaked event fields are invalid")
			}
		case "resource.released":
			// D-04-07's non-terminal release transition: like the other
			// linear transitions it is never last (a terminator always
			// follows), but unlike them it produces no new place, so
			// TargetPlace must stay empty rather than required.
			if (value.Schema != execution.Schema1 && value.Schema != execution.Schema2) || isLast || event.SourcePlace == "" || event.TargetPlace != "" || event.TypeID == "" || event.Input != "" || event.Output != "" {
				return errors.New("resource release event fields are invalid")
			}
		default:
			return errors.New("unknown execution event kind")
		}
	}
	return nil
}

func bounded(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "…"
}
