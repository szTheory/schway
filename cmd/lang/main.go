package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/codename-lang/lang/internal/compiler/diagnostic"
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
	args, jsonMode, ok := extractJSON(args)
	if !ok {
		return emit(usageResult(), true, false)
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
		return runEvidenceValidation(args[2], args[3], jsonMode)
	}
	if len(args) == 2 && args[0] == "verify" {
		return runVerify(args[1], jsonMode)
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
		return emit(problemResult("evidence", protocol.StatusOperational, "evidence.operation_failed", "unable to construct evidence"), jsonMode, false)
	}
	if !jsonMode && result.Status == protocol.StatusPass {
		if _, err := os.Stdout.Write(product.ManifestBytes); err != nil {
			return emit(problemResult("evidence", protocol.StatusOperational, "tool.write_failed", "unable to write output"), false, false)
		}
		return exitSuccess
	}
	return emit(result, jsonMode, false)
}

func runEvidenceValidation(manifestPath, sourcePath string, jsonMode bool) int {
	result := session.ValidateEvidenceCommandFile(context.Background(), manifestPath, sourcePath)
	return emit(result, jsonMode, false)
}

func runVerify(corpus string, jsonMode bool) int {
	result := session.VerifyCorpusFile(context.Background(), corpus, native.DefaultRunner())
	return emit(result, jsonMode, false)
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
		encoded = []byte(protocol.Human(result))
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
	return problemResult("usage", protocol.StatusUsage, "tool.usage", "usage: lang [--json] format [--check] FILE | check FILE | run --engine=interpreter|native FILE | evidence FILE | evidence --validate MANIFEST FILE | verify CORPUS | interface export SRC OUT | interface core SRC OUT | interface check SUMMARY CORE")
}
