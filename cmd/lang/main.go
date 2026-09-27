package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/evidence"
	"github.com/codename-lang/lang/internal/compiler/measure"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/protocol"
	"github.com/codename-lang/lang/internal/compiler/session"
)

const (
	exitSuccess       = 0
	exitInvalidSource = 2
	exitOperational   = 3
	exitUsage         = 64
)

func main() { os.Exit(run(os.Args[1:])) }

func run(args []string) int {
	// Route the application command before global flag extraction so the token
	// after `--` stays opaque, even when it is spelled like a lang flag.
	if len(args) > 0 && args[0] == "app" {
		return runApplication(args)
	}
	if len(args) > 1 && args[0] == "--json" && args[1] == "app" {
		return runApplication(args[1:])
	}
	args, jsonMode, ok := extractJSON(args)
	if !ok {
		return emit(usageResult(), true, false)
	}
	args, depth, ok := extractDepth(args)
	if !ok {
		return emit(usageResult(), jsonMode, false)
	}
	args, kind, cursor, ok := extractKindAndCursor(args)
	if !ok {
		return emit(usageResult(), jsonMode, false)
	}
	args, expand, ok := extractExpand(args)
	if !ok {
		return emit(usageResult(), jsonMode, false)
	}
	if len(args) == 4 && args[0] == "build" && args[2] == "--output" {
		return runBuild(args[1], args[3], jsonMode)
	}
	if len(args) == 6 && args[0] == "build" {
		if args[2] == "--manifest" && args[4] == "--output" {
			return runBuild(args[1], args[5], jsonMode, args[3])
		}
		if args[2] == "--output" && args[4] == "--manifest" {
			return runBuild(args[1], args[3], jsonMode, args[5])
		}
	}
	if len(args) == 3 && args[0] == "explain" {
		return runExplain(args[1], args[2], depth, jsonMode)
	}
	if len(args) == 3 && args[0] == "query" {
		return runQuery(args[1], args[2], kind, cursor, depth, jsonMode)
	}
	if len(args) == 2 && args[0] == "format" {
		return runFormat(args[1], false, jsonMode)
	}
	if len(args) == 3 && args[0] == "format" && args[1] == "--check" {
		return runFormat(args[2], true, jsonMode)
	}
	if len(args) == 2 && args[0] == "check" {
		return runCheck(args[1], jsonMode)
	}
	if len(args) == 2 && args[0] == "evidence" {
		return runEvidence(args[1], jsonMode)
	}
	if len(args) == 4 && args[0] == "evidence" && args[1] == "--validate" {
		return runEvidenceValidation(args[2], args[3], expand, jsonMode)
	}
	if len(args) == 2 && args[0] == "verify" {
		return runVerify(args[1], jsonMode)
	}
	if len(args) == 1 && args[0] == "stats" {
		return runStats()
	}
	if len(args) == 4 && args[0] == "interface" && args[1] == "export" {
		return runInterfaceExport(args[2], args[3], jsonMode)
	}
	if len(args) == 4 && args[0] == "interface" && args[1] == "core" {
		return runInterfaceCore(args[2], args[3], jsonMode)
	}
	if len(args) == 4 && args[0] == "interface" && args[1] == "check" {
		return runInterfaceCheck(args[2], args[3], jsonMode)
	}
	if len(args) == 2 && args[0] == "debug-map" {
		return runDebugMap(args[1], "", jsonMode)
	}
	if len(args) == 3 && args[0] == "debug-map" {
		return runDebugMap(args[1], args[2], jsonMode)
	}
	if len(args) == 3 && args[0] == "run" && strings.HasPrefix(args[1], "--engine=") {
		engine := strings.TrimPrefix(args[1], "--engine=")
		switch engine {
		case "interpreter":
			return runInterpreter(args[2], jsonMode)
		case "native":
			return runNative(args[2], jsonMode)
		default:
			return emit(problemResult("run", protocol.StatusUsage, "tool.unsupported_engine", "expected interpreter or native"), jsonMode, false)
		}
	}
	return emit(usageResult(), jsonMode, false)
}

func runFormat(path string, checkOnly, jsonMode bool) int {
	result, err := session.FormatCommandFile(path, checkOnly)
	if err != nil {
		return emit(problemResult("format", protocol.StatusOperational, "tool.read_failed", "unable to read input"), jsonMode, false)
	}
	if !checkOnly && !jsonMode && result.Status == protocol.StatusPass {
		if _, err := os.Stdout.Write([]byte(result.Formatted)); err != nil {
			return emit(problemResult("format", protocol.StatusOperational, "tool.write_failed", "unable to write output"), false, false)
		}
		return exitSuccess
	}
	return emit(result, jsonMode, false)
}

func runNative(path string, jsonMode bool) int {
	result, err := session.RunNativeCommandFile(context.Background(), path, native.DefaultRunner())
	if err != nil {
		return emit(problemResult("run", protocol.StatusOperational, "tool.read_failed", "unable to read input"), jsonMode, false)
	}
	return emit(result, jsonMode, false)
}

func runBuild(sourcePath, outputPath string, jsonMode bool, manifestPath ...string) int {
	receipt, diagnostics, err := session.BuildApplicationFile(context.Background(), sourcePath, outputPath, native.DefaultRunner(), manifestPath...)
	if len(diagnostics) > 0 {
		result := protocol.New("build", protocol.StatusInvalid)
		result.Diagnostics = diagnostics
		return emit(result, jsonMode, false)
	}
	if err != nil {
		code := "native.build_failed"
		var toolError *native.ToolError
		if errors.As(err, &toolError) && toolError.Code != "" {
			code = toolError.Code
		}
		message := "unable to build retained application: " + err.Error()
		return emit(problemResult("build", protocol.StatusOperational, code, message), jsonMode, false)
	}
	if jsonMode {
		if err := json.NewEncoder(os.Stdout).Encode(receipt); err != nil {
			fmt.Fprintln(os.Stderr, "build: unable to write build receipt summary")
			return exitOperational
		}
		return exitSuccess
	}
	if _, err := fmt.Fprintf(os.Stdout, "built %s\n", outputPath); err != nil {
		return exitOperational
	}
	return exitSuccess
}

func runApplication(args []string) int {
	if len(args) < 5 || args[1] != "run" {
		fmt.Fprintln(os.Stderr, "usage: lang app run ARTIFACT [--report REPORT] [--evidence=events] -- INPUT")
		return exitUsage
	}
	separator := -1
	for index := 3; index < len(args); index++ {
		if args[index] == "--" {
			separator = index
			break
		}
	}
	if separator < 0 || separator+2 != len(args) {
		fmt.Fprintln(os.Stderr, "usage: lang app run ARTIFACT [--report REPORT] [--evidence=events] -- INPUT")
		return exitUsage
	}
	reportPath := ""
	evidenceMode := native.EvidenceDisabled
	seenReport, seenEvidence := false, false
	for index := 3; index < separator; index++ {
		switch {
		case args[index] == "--report":
			if seenReport || index+1 >= separator || args[index+1] == "" {
				fmt.Fprintln(os.Stderr, "lang app run: --report requires one path and may appear only once")
				return exitUsage
			}
			seenReport = true
			reportPath = args[index+1]
			index++
		case args[index] == "--evidence=events":
			if seenEvidence {
				fmt.Fprintln(os.Stderr, "lang app run: --evidence may appear only once")
				return exitUsage
			}
			seenEvidence = true
			evidenceMode = native.EvidenceEvents
		case strings.HasPrefix(args[index], "--evidence="):
			fmt.Fprintln(os.Stderr, "lang app run: --evidence supports only events")
			return exitUsage
		default:
			fmt.Fprintln(os.Stderr, "lang app run: unknown option")
			return exitUsage
		}
	}
	if seenEvidence && !seenReport {
		fmt.Fprintln(os.Stderr, "lang app run: --evidence=events requires --report")
		return exitUsage
	}
	var outcome native.RunOutcome
	var err error
	if seenReport {
		outcome, _, err = native.RunApplicationWithEvidence(context.Background(), args[2], args[separator+1], reportPath, evidenceMode, os.Stdout, os.Stderr)
	} else {
		outcome, err = native.RunApplication(context.Background(), args[2], args[separator+1], os.Stdout, os.Stderr)
	}
	if err != nil {
		var toolError *native.ToolError
		if errors.As(err, &toolError) && toolError.Code == "native.input_too_long" {
			fmt.Fprintf(os.Stderr, "lang app run: %v\n", err)
			return 65
		}
		fmt.Fprintf(os.Stderr, "lang app run: %v\n", err)
		return 125
	}
	switch outcome.Kind {
	case native.RunExited:
		return outcome.ExitCode
	case native.RunSignaled:
		return 128 + outcome.SignalNumber
	case native.RunTimedOut:
		fmt.Fprintf(os.Stderr, "lang app run: %s\n", outcome.Diagnostic)
		return 124
	case native.RunLaunchError:
		fmt.Fprintf(os.Stderr, "lang app run: %s\n", outcome.Diagnostic)
		return 125
	default:
		fmt.Fprintln(os.Stderr, "lang app run: invalid process outcome")
		return 125
	}
}

func runCheck(path string, jsonMode bool) int {
	result, err := session.CheckCommandFile(path)
	if err != nil {
		return emit(problemResult("check", protocol.StatusOperational, "tool.read_failed", "unable to read input"), jsonMode, false)
	}
	return emit(result, jsonMode, false)
}

func runInterpreter(path string, jsonMode bool) int {
	result, err := session.RunInterpreterCommandFile(path)
	if err != nil {
		return emit(problemResult("run", protocol.StatusOperational, "tool.read_failed", "unable to read input"), jsonMode, false)
	}
	return emit(result, jsonMode, false)
}

func runEvidence(path string, jsonMode bool) int {
	product, result, err := session.EvidenceCommandFile(context.Background(), path)
	if err != nil {
		// D-02-02: route the specific ValidationError code (e.g.
		// evidence.canonical_unstable) through to the CLI instead of
		// collapsing every build failure into the generic
		// evidence.operation_failed code that only tests could previously see.
		return emit(problemResult("evidence", protocol.StatusOperational, evidence.ErrorCode(err), "unable to construct evidence"), jsonMode, false)
	}
	if !jsonMode && result.Status == protocol.StatusPass {
		if _, err := os.Stdout.Write(product.ManifestBytes); err != nil {
			return emit(problemResult("evidence", protocol.StatusOperational, "tool.write_failed", "unable to write output"), false, false)
		}
		return exitSuccess
	}
	return emit(result, jsonMode, false)
}

func runEvidenceValidation(manifestPath, sourcePath string, expand, jsonMode bool) int {
	result := session.ValidateEvidenceExpanded(context.Background(), manifestPath, sourcePath, expand)
	return emit(result, jsonMode, false)
}

func runVerify(corpus string, jsonMode bool) int {
	if isPhase7Corpus(corpus) {
		result, err := session.VerifyPhase7ControlsAndWork(context.Background())
		if err != nil {
			return emit(problemResult("verify", protocol.StatusOperational, "tool.phase7_gate_failed", "unable to run the phase 7 control-and-work gate"), jsonMode, false)
		}
		return emit(result, jsonMode, false)
	}
	if isPhase6Corpus(corpus) {
		result, err := session.VerifyPhase6ControlsAndWork(context.Background())
		if err != nil {
			return emit(problemResult("verify", protocol.StatusOperational, "tool.phase6_gate_failed", "unable to run the phase 6 control-and-work gate"), jsonMode, false)
		}
		return emit(result, jsonMode, false)
	}
	if isPhase5Corpus(corpus) {
		result, err := session.VerifyPhase5ControlsAndWork(context.Background())
		if err != nil {
			return emit(problemResult("verify", protocol.StatusOperational, "tool.phase5_gate_failed", "unable to run the phase 5 control-and-work gate"), jsonMode, false)
		}
		return emit(result, jsonMode, false)
	}
	result := session.VerifyCorpusFile(context.Background(), corpus, native.DefaultRunner())
	return emit(result, jsonMode, false)
}

// runStats is plan 06-15's small Go statistics seam for
// scripts/verify-phase6.sh's own D-06-19 20-sample observation loop: it
// reads newline-separated integer samples from stdin and prints their
// measure.Samples.Summary() (p50, p95, CoV, count) as a single line of
// JSON. This keeps the shell side of the sampling loop thin -- it drives
// the binary WarmSampleCount times and collects raw elapsed_ns values,
// then pipes them through THIS command -- rather than reimplementing
// sorted-index percentile selection in sed, so the statistical logic
// itself stays unit-tested Go (measure.Samples.Summary(), already covered
// by internal/compiler/measure's own test suite) instead of untested
// shell arithmetic.
func runStats() int {
	scanner := bufio.NewScanner(os.Stdin)
	var samples measure.Samples
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		value, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "stats: invalid sample %q: %v\n", line, err)
			return exitUsage
		}
		samples = append(samples, value)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "stats: %v\n", err)
		return exitOperational
	}
	summary, err := samples.Summary()
	if err != nil {
		fmt.Fprintf(os.Stderr, "stats: %v\n", err)
		return exitOperational
	}
	fmt.Printf("{\"p50\":%d,\"p95\":%d,\"cov\":%g,\"count\":%d}\n", summary.P50, summary.P95, summary.CoV, summary.Count)
	return exitSuccess
}

// isPhase7Corpus recognizes the testdata/phase07 corpus by its own
// characteristic marker fixture (call_basic.lang, the Phase 07 tracer
// fixture), the same dispatch-by-marker-file discipline isPhase5Corpus/
// isPhase6Corpus already established. Checked BEFORE isPhase6Corpus so a
// Phase 07 corpus is never accidentally swallowed by an earlier check; a
// directory with no Phase 07 marker falls through unchanged to the
// existing dispatch chain (isPhase6Corpus, isPhase5Corpus, then the
// default VerifyCorpusFile path). D-07-41: this dispatch lives at the CLI
// layer, mirroring isPhase5Corpus/isPhase6Corpus's own precedent of
// keeping session.go's generic VerifyCorpus untouched by each new phase.
func isPhase7Corpus(corpus string) bool {
	_, err := os.Stat(filepath.Join(corpus, "call_basic.lang"))
	return err == nil
}

// isPhase6Corpus recognizes the testdata/phase6 corpus by its own
// characteristic marker fixture (heldout_match_defect.lang, per
// testdata/phase6/README), the same dispatch-by-marker-file discipline
// isPhase5Corpus already established. Checked BEFORE isPhase5Corpus so a
// Phase 6 corpus is never accidentally swallowed by an earlier check; a
// directory with no Phase 6 marker falls through unchanged to the existing
// dispatch chain (isPhase5Corpus, then the default VerifyCorpusFile path).
func isPhase6Corpus(corpus string) bool {
	_, err := os.Stat(filepath.Join(corpus, "heldout_match_defect.lang"))
	return err == nil
}

// isPhase5Corpus recognizes the testdata/phase5 corpus by its own
// characteristic marker fixture (restrict_borrow.lang), the same
// dispatch-by-marker-file discipline session.VerifyCorpus itself uses to
// pick between the foreign/borrowed/owned corpora. D-05-17 requires the
// Phase 5 gate to be reachable through `lang verify testdata/phase5` (the
// mandatory verify/release cost lane scripts/verify-phase5.sh drives), but
// session.go stays untouched throughout Phase 5 (every new Phase 5
// surface lives in session_phase5*.go siblings) -- so this dispatch lives
// at the CLI layer instead of inside session.VerifyCorpus itself.
func isPhase5Corpus(corpus string) bool {
	_, err := os.Stat(filepath.Join(corpus, "restrict_borrow.lang"))
	return err == nil
}

// runInterfaceExport is OWN-04's producer-side CLI seam: it checks SRC,
// independently recomputes every declared public origin from the typed core
// alone, and — only on success — writes a body-stripped, digest-bound
// interface summary to OUT.
func runInterfaceExport(source, out string, jsonMode bool) int {
	result, err := session.InterfaceExportCommandFile(source, out)
	if err != nil {
		return emit(problemResult("interface", protocol.StatusOperational, "tool.operation_failed", "unable to export interface"), jsonMode, false)
	}
	return emit(result, jsonMode, false)
}

// runInterfaceCore writes the checked-and-validated core artifact for SRC to
// OUT — a standalone way to obtain the exact bytes an interface summary's
// digest is bound to (a real pipeline already retains this from `lang
// check`).
func runInterfaceCore(source, out string, jsonMode bool) int {
	result, err := session.InterfaceCoreCommandFile(source, out)
	if err != nil {
		return emit(problemResult("interface", protocol.StatusOperational, "tool.operation_failed", "unable to write core artifact"), jsonMode, false)
	}
	return emit(result, jsonMode, false)
}

// runInterfaceCheck is OWN-04's consumer-side CLI seam: a genuinely separate
// process invocation that decides origin/access questions from SUMMARY alone.
// CORE's bytes are hashed and compared against SUMMARY's recorded digest but
// are never decoded into a core.Program — this is the real second invocation
// success criterion 4 requires, not an in-process simulation.
func runInterfaceCheck(summary, core string, jsonMode bool) int {
	result, err := session.InterfaceCheckCommandFile(summary, core)
	if err != nil {
		return emit(problemResult("interface", protocol.StatusOperational, "tool.read_failed", "unable to read interface inputs"), jsonMode, false)
	}
	return emit(result, jsonMode, false)
}

// runDebugMap is the bounded debug-lineage experiment's CLI seam
// (D-01..D-04, Task 03-07-01): with no query it dumps the full joined
// source/core/operation map for SRC; with a query it resolves a single
// operation ID, proving the honest not_captured report through the shipped
// binary for an ID the map never produced.
func runDebugMap(source, query string, jsonMode bool) int {
	result, err := session.DebugMapCommandFile(source, query)
	if err != nil {
		return emit(problemResult("debug-map", protocol.StatusOperational, "tool.read_failed", "unable to read input"), jsonMode, false)
	}
	return emit(result, jsonMode, false)
}

// runExplain is D-06-02's/D-06-04's CLI seam for the net-new `lang.explain/0`
// schema: it re-derives the requested diagnostic's bounded cause DAG from
// SRC on this cold invocation alone (D-06-02 forbids any persisted store or
// daemon a bare ID could resolve against, matching the `debug-map SRC
// [QUERY]` operand convention).
func runExplain(source, id string, depth int, jsonMode bool) int {
	result, err := session.ExplainCommandFile(source, id, depth)
	if err != nil {
		return emit(problemResult("explain", protocol.StatusOperational, "tool.read_failed", "unable to read input"), jsonMode, false)
	}
	return emit(result, jsonMode, false)
}

// runQuery is D-06-01's CLI seam for the net-new `lang.query/0` schema: one
// joined addressing surface over the five stable ID vocabularies already
// shipped in the tree, re-derived from SRC on this cold invocation alone
// (matching explain/debug-map's no-persisted-store discipline).
func runQuery(source, address, kind, cursor string, depth int, jsonMode bool) int {
	options := session.QueryOptions{Kind: kind, Cursor: cursor, Depth: depth}
	result, err := session.QueryCommandFile(source, address, options)
	if err != nil {
		return emit(problemResult("query", protocol.StatusOperational, "tool.read_failed", "unable to read input"), jsonMode, false)
	}
	return emit(result, jsonMode, false)
}

// extractKindAndCursor strips `--kind=K` and `--cursor=C` flags from args,
// mirroring extractDepth/extractJSON's own strip-and-report shape. A
// repeated flag is a usage error (returns ok=false), exactly like
// extractJSON's repeated-flag case.
func extractKindAndCursor(args []string) ([]string, string, string, bool) {
	filtered := make([]string, 0, len(args))
	kind := ""
	cursor := ""
	seenKind := false
	seenCursor := false
	for _, argument := range args {
		switch {
		case strings.HasPrefix(argument, "--kind="):
			if seenKind {
				return nil, "", "", false
			}
			seenKind = true
			kind = strings.TrimPrefix(argument, "--kind=")
		case strings.HasPrefix(argument, "--cursor="):
			if seenCursor {
				return nil, "", "", false
			}
			seenCursor = true
			cursor = strings.TrimPrefix(argument, "--cursor=")
		default:
			filtered = append(filtered, argument)
		}
	}
	return filtered, kind, cursor, true
}

// extractDepth strips one `--depth=N` flag from args, mirroring
// extractJSON's own strip-and-report shape. A missing flag resolves to
// protocol.ExplainDefaultDepth; a malformed or repeated flag is a usage
// error (returns ok=false), exactly like extractJSON's repeated-flag case.
func extractDepth(args []string) ([]string, int, bool) {
	filtered := make([]string, 0, len(args))
	depth := protocol.ExplainDefaultDepth
	seen := false
	for _, argument := range args {
		if strings.HasPrefix(argument, "--depth=") {
			if seen {
				return nil, 0, false
			}
			seen = true
			parsed, err := strconv.Atoi(strings.TrimPrefix(argument, "--depth="))
			if err != nil || parsed <= 0 {
				return nil, 0, false
			}
			depth = parsed
			continue
		}
		filtered = append(filtered, argument)
	}
	return filtered, depth, true
}

// extractExpand strips the boolean "--expand" flag (DX-03's evidence-trace
// expansion request), mirroring extractJSON's own boolean-flag-stripping
// shape exactly.
func extractExpand(args []string) ([]string, bool, bool) {
	filtered := make([]string, 0, len(args))
	expand := false
	for _, argument := range args {
		if argument == "--expand" {
			if expand {
				return nil, false, false
			}
			expand = true
			continue
		}
		filtered = append(filtered, argument)
	}
	return filtered, expand, true
}

func extractJSON(args []string) ([]string, bool, bool) {
	filtered := make([]string, 0, len(args))
	jsonMode := false
	for _, argument := range args {
		if argument == "--json" {
			if jsonMode {
				return nil, true, false
			}
			jsonMode = true
			continue
		}
		filtered = append(filtered, argument)
	}
	return filtered, jsonMode, true
}

func emit(result protocol.Result, jsonMode, forceStderr bool) int {
	result = result.Finalize()
	var encoded []byte
	var err error
	if jsonMode {
		encoded, err = protocol.JSON(result)
	} else {
		var humanOutput string
		humanOutput, err = protocol.Human(result)
		encoded = []byte(humanOutput)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "tool.encode_failed: unable to encode command result")
		return exitOperational
	}
	destination := io.Writer(os.Stdout)
	if !jsonMode && (result.Status != protocol.StatusPass || forceStderr) {
		destination = os.Stderr
	}
	if _, err := destination.Write(encoded); err != nil {
		return exitOperational
	}
	return protocol.ExitCode(result.Status)
}

func problemResult(command, status, code, message string) protocol.Result {
	result := protocol.New(command, status)
	result.Diagnostics = []diagnostic.Diagnostic{diagnostic.Error(code, diagnostic.Span{}, message)}
	return result
}

func usageResult() protocol.Result {
	return problemResult("usage", protocol.StatusUsage, "tool.usage", "usage: lang [--json] build SOURCE --output ARTIFACT | app run ARTIFACT [--report REPORT] [--evidence=events] -- INPUT | app verify SOURCE --cases CASES --report REPORT | format [--check] FILE | check FILE | run --engine=interpreter|native FILE | evidence FILE | evidence --validate MANIFEST FILE [--expand] | verify CORPUS | interface export SRC OUT | interface core SRC OUT | interface check SUMMARY CORE | debug-map SRC [QUERY] | explain SRC ID [--depth=N] | query SRC ID_OR_PATTERN [--kind=symbol|type|ownership|dependency|test] [--depth=N] [--cursor=C]")
}
