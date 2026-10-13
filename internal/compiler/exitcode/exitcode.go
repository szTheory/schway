// Package exitcode is the single named exit-code table for the schway
// binaries (D-07, DX-17). It imports nothing so every layer may read it.
// No 126 or 127 constant exists because nothing in schway emits them.
package exitcode

const (
	// Success is a clean exit (all domains).
	Success = 0
	// Unrepaired is the schway-repair "no repair applied" exit (repair domain).
	Unrepaired = 1
	// Invalid means the compiler judged the source invalid (compiler domain).
	Invalid = 2
	// Operational means a tool or environment failure (compiler, repair).
	Operational = 3
	// Mismatch means a comparison or verification mismatch (compiler domain).
	Mismatch = 4
	// Usage means a command-line usage error (all domains); it also covers
	// `run` refusing an entry whose parameter it cannot bind
	// (tool.run_input_required).
	Usage = 64
	// AppInputRefused means app run refused its input argument (app-run
	// domain). Published contract baked into generated C and
	// scripts/verify-phase27.sh; never renumber. The checksum program's
	// checked failures (FileTooLarge, FileUnavailable, AllocFailed) exit with
	// this code.
	AppInputRefused = 65
	// EventCapacity is the generated C exit when its static event array is
	// full (app-run domain). Published contract baked into generated C and
	// scripts/verify-phase27.sh; never renumber.
	EventCapacity = 69
	// AppTimeout means the app run timed out (app-run domain).
	AppTimeout = 124
	// AppLaunchError means the app could not be launched or run (app-run).
	AppLaunchError = 125
	// SignalBase is the POSIX shell base for signal exits (app-run domain).
	SignalBase = 128
)

// SignalExit returns the POSIX shell 128+n exit status for signal n.
func SignalExit(n int) int { return SignalBase + n }
