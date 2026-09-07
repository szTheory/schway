package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const (
	exitSuccess     = 0
	exitUnrepaired  = 1
	exitOperational = 3
	exitUsage       = 64
)

func main() { os.Exit(run(os.Args[1:])) }

// run follows cmd/lang/main.go's own style: flat arg matching, no `flag`
// package, no third-party library (D-06-28's protocol-only driver is its
// own tier even at the CLI-parsing level).
func run(args []string) int {
	langBinary, source, jsonMode, ok := parseArgs(args)
	if !ok {
		fmt.Fprintln(os.Stderr, usage())
		return exitUsage
	}
	outcome, err := Repair(context.Background(), langBinary, source)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return exitOperational
	}
	if jsonMode {
		encoded, encErr := json.Marshal(outcome)
		if encErr != nil {
			fmt.Fprintln(os.Stderr, encErr.Error())
			return exitOperational
		}
		fmt.Fprintln(os.Stdout, string(encoded))
	} else {
		fmt.Fprintf(os.Stdout, "status=%s diagnosis=%s repair=%s subprocess_count=%d\n",
			outcome.Status, outcome.DiagnosisCode, outcome.RepairKind, outcome.SubprocessCount)
	}
	if outcome.Status == OutcomeRepaired || outcome.Status == OutcomeAlreadyClean {
		return exitSuccess
	}
	return exitUnrepaired
}

func usage() string {
	return "usage: lang-repair --lang=PATH_TO_LANG_BINARY --source=FILE [--json]"
}

// parseArgs recognizes exactly `--lang=PATH`, `--source=FILE`, and the
// optional `--json` flag -- both `--lang` and `--source` are required.
func parseArgs(args []string) (langBinary, source string, jsonMode, ok bool) {
	for _, arg := range args {
		switch {
		case arg == "--json":
			jsonMode = true
		case strings.HasPrefix(arg, "--lang="):
			langBinary = strings.TrimPrefix(arg, "--lang=")
		case strings.HasPrefix(arg, "--source="):
			source = strings.TrimPrefix(arg, "--source=")
		default:
			return "", "", false, false
		}
	}
	if langBinary == "" || source == "" {
		return "", "", false, false
	}
	return langBinary, source, jsonMode, true
}
