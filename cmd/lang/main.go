package main

import (
	"fmt"
	"os"

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
	if len(args) != 2 || args[0] != "check" {
		fmt.Fprintln(os.Stderr, "usage: lang check FILE")
		return exitUsage
	}
	result, err := session.CheckFile(args[1])
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
