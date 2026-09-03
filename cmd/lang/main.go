package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/codename-lang/lang/internal/compiler/native"
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
	if len(args) == 2 && args[0] == "check" {
		return runCheck(args[1])
	}
	if len(args) == 3 && args[0] == "run" && strings.HasPrefix(args[1], "--engine=") {
		engine := strings.TrimPrefix(args[1], "--engine=")
		switch engine {
		case "interpreter":
			return runInterpreter(args[2])
		case "native":
			return runNative(args[2])
		default:
			fmt.Fprintln(os.Stderr, "tool.unsupported_engine: expected interpreter or native")
			return exitUsage
		}
	}
	fmt.Fprintln(os.Stderr, "usage: lang check FILE | lang run --engine=interpreter|native FILE")
	return exitUsage
}

func runNative(path string) int {
	result, diagnostics, err := session.RunNativeFile(context.Background(), path, native.DefaultRunner())
	if len(diagnostics) > 0 {
		for _, problem := range diagnostics {
			fmt.Fprintln(os.Stderr, problem.String())
		}
		return exitInvalidSource
	}
	if err != nil {
		var mismatch *session.EngineMismatch
		if errors.As(err, &mismatch) {
			fmt.Fprintf(os.Stderr, "native.engine_mismatch: %v\n", err)
			return 4
		}
		fmt.Fprintf(os.Stderr, "native.tool_failure: %v\n", err)
		return exitOperational
	}
	encoded, err := json.Marshal(result.Interpreter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tool.encode_failed: %v\n", err)
		return exitOperational
	}
	fmt.Fprintln(os.Stdout, string(encoded))
	return exitSuccess
}

func runCheck(path string) int {
	result, err := session.CheckFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tool.read_failed: %v\n", err)
		return exitOperational
	}
	if len(result.Diagnostics) > 0 {
		for _, problem := range result.Diagnostics {
			fmt.Fprintln(os.Stderr, problem.String())
		}
		return exitInvalidSource
	}
	fmt.Fprintf(os.Stdout, "checked %s\n", result.Program.ModuleID)
	return exitSuccess
}

func runInterpreter(path string) int {
	executions, diagnostics, err := session.RunInterpreterFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tool.run_failed: %v\n", err)
		return exitOperational
	}
	if len(diagnostics) > 0 {
		for _, problem := range diagnostics {
			fmt.Fprintln(os.Stderr, problem.String())
		}
		return exitInvalidSource
	}
	encoded, err := json.Marshal(executions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tool.encode_failed: %v\n", err)
		return exitOperational
	}
	fmt.Fprintln(os.Stdout, string(encoded))
	return exitSuccess
}
