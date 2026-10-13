package session

import (
	"fmt"
	"os"
	"time"

	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/protocol"
)

// ErrRunInputRequired is returned when `run` resolved a valid entry whose
// parameter it has no way to bind (Phase 32.1, D-01/D-03). It wraps
// os.ErrInvalid so any caller that tested the earlier sentinel keeps working.
var ErrRunInputRequired = fmt.Errorf("run supplies no input for the entry parameter: %w", os.ErrInvalid)

// runInputRequiredCode is the single diagnostic code for the refusal.
const runInputRequiredCode = "tool.run_input_required"

// runInputRequiredMessage is a static constant (D-01 security): it never
// carries the source path, source text or parameter name.
const runInputRequiredMessage = "run supplies no input for this entry's parameter; use build, then app run ARTIFACT -- INPUT"

// runInputRequiredResult builds the usage-class (exit 64) refusal result.
func runInputRequiredResult(started time.Time) protocol.Result {
	result := protocol.New("run", protocol.StatusUsage)
	// append (not a typed literal) keeps this file out of the Unit consumer
	// inventory: it builds no Diagnostic value of its own shape.
	result.Diagnostics = append(result.Diagnostics, diagnostic.Error(runInputRequiredCode, diagnostic.Span{}, runInputRequiredMessage))
	return completeCommand(result, started, 1)
}
