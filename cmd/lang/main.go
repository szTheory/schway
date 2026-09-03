package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

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
		if engine != "interpreter" {
			fmt.Fprintln(os.Stderr, "tool.unsupported_engine: expected interpreter")
			return exitUsage
		}
		return runInterpreter(args[2])
	}
	fmt.Fprintln(os.Stderr, "usage: lang check FILE | lang run --engine=interpreter FILE")
	return exitUsage
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
