package interp_test

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/session"
)

const phase25IntegratedSource = `module phase25.integrated_utility

export {
  fn acquire
  fn shared_copy
  fn exclusive_copy
  fn main
}

foreign C {
  fn schway_file_byte_acquire(path: PathToken) -> FileByteOwner {
    unwind: forbidden
    nonlocal_exit: forbidden
    mode: "acquire"
    abi: schway_file_byte_acquire_fn
    allocator: "libc_malloc"
    release: "schway_file_byte_release"
    fails: AcquireError
  }
  fn schway_file_byte_use(owner: FileByteOwner) -> U64 {
    unwind: forbidden
    nonlocal_exit: forbidden
    mode: "borrow"
    abi: schway_file_byte_use_fn
    fails: UseError
  }
  fn schway_file_byte_release(owner: FileByteOwner) -> Unit {
    unwind: forbidden
    nonlocal_exit: forbidden
    mode: "consume"
    abi: schway_file_byte_release_fn
    allocator: "libc_malloc"
  }
}

data AcquireError = | AcquireFailed
data UseError = | UnsupportedByte

fn acquire(path: PathToken) -> FileByteOwner {
  let owner = try schway_file_byte_acquire(path)
  owner
}

fn shared_copy(value: U64) -> U64 {
  let borrowed = borrow value
  let copied = borrowed
  copied
}

fn exclusive_copy(value: U64) -> U64 {
  let borrowed = borrow mut value
  let copied = borrowed
  copied
}

fn main(path: PathToken) -> U64 {
  let owner = acquire(path)
  let value = try schway_file_byte_use(owner)
  let shared = shared_copy(value)
  let exclusive = exclusive_copy(shared)
  exclusive
}
`

func phase25CheckedProgram(t *testing.T) core.Program {
	t.Helper()
	checked := session.Check([]byte(phase25IntegratedSource))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("integrated utility source failed to check: %+v causes=%+v", checked.Diagnostics, checked.Diagnostics[0].Causes)
	}
	if validation := corevalidate.Validate(checked.Program); !validation.Valid {
		t.Fatalf("integrated utility failed core validation: %+v", validation.Problems)
	}
	return checked.Program
}

func phase25ModelOutcomes(t *testing.T, program core.Program, value string, failure bool) map[string]interp.ForeignOutcome {
	t.Helper()
	var acquireID, useID string
	for _, function := range program.Functions {
		if function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Foreign == nil {
				continue
			}
			switch operation.Foreign.Mode {
			case "acquire":
				acquireID = operation.ID
			case "borrow":
				useID = operation.ID
			}
		}
	}
	if acquireID == "" || useID == "" {
		t.Fatal("integrated utility acquire/use operations are missing")
	}
	useOutcome := interp.ForeignOutcome{Kind: "success", Type: "U64", Value: value}
	if failure {
		useOutcome = interp.ForeignOutcome{Kind: "failure", Type: "UseError", Value: "UnsupportedByte"}
	}
	return map[string]interp.ForeignOutcome{
		acquireID: {Kind: "success", Type: "FileByteOwner", Value: value},
		useID:     useOutcome,
	}
}

func TestPhase25InterpreterComposition(t *testing.T) {
	program := phase25CheckedProgram(t)
	for _, test := range []struct{ input, want string }{{"65", "65"}, {"66", "66"}} {
		result, err := interp.RunWithForeignOutcomes(program, "main", "opaque-path-token", phase25ModelOutcomes(t, program, test.input, false))
		if err != nil {
			t.Fatalf("model byte %s: %v", test.input, err)
		}
		if result.Execution.Outcome.Kind != execution.OutcomeReturned || result.Execution.Outcome.Value != test.want {
			t.Fatalf("model byte %s produced %+v, want returned %s", test.input, result.Execution.Outcome, test.want)
		}
		var calls []string
		for _, event := range result.Execution.Events {
			if event.Kind == "function.called" {
				calls = append(calls, event.CalleeFunctionID)
			}
		}
		if len(calls) != 3 || functionName(program, calls[1]) != "shared_copy" || functionName(program, calls[2]) != "exclusive_copy" {
			t.Fatalf("model call order=%v, want acquire transfer then shared and exclusive copies", calls)
		}
		if result.ActualHostIO || result.PhysicalCleanup || result.EvidenceScope != interp.EvidenceScopeModelOnly {
			t.Fatalf("model execution overstated native evidence: %+v", result)
		}
	}
}

func TestPhase25OwnerTransferExactHelperChain(t *testing.T) {
	program := phase25CheckedProgram(t)
	if result := corevalidate.Validate(program); !result.Valid {
		t.Fatalf("exact shared-to-exclusive copied result chain was refused: %+v", result.Problems)
	}

	for _, mutate := range []struct {
		name string
		edit func(*core.Function)
	}{
		{"reordered helpers", func(main *core.Function) {
			calls := helperCallIndexes(main)
			main.Linear.Operations[calls[0]].CalleeID, main.Linear.Operations[calls[1]].CalleeID = main.Linear.Operations[calls[1]].CalleeID, main.Linear.Operations[calls[0]].CalleeID
		}},
		{"arbitrary returned intermediate", func(main *core.Function) {
			for i := range main.Linear.Operations {
				if main.Linear.Operations[i].Kind == core.OpReturn {
					main.Linear.Operations[i].SourceID = main.Linear.Operations[helperCallIndexes(main)[0]].TargetID
					return
				}
			}
		}},
	} {
		t.Run(mutate.name, func(t *testing.T) {
			candidate := phase25CheckedProgram(t)
			main := functionByName(&candidate, "main")
			mutate.edit(main)
			if result := corevalidate.Validate(candidate); result.Valid {
				t.Fatalf("validator admitted tampered helper flow into return: %+v", result.PeerSignatures())
			}
		})
	}
}

func helperCallIndexes(function *core.Function) []int {
	var indexes []int
	for index, operation := range function.Linear.Operations {
		if operation.Kind == core.OpCall && operation.CalleeID != "" {
			indexes = append(indexes, index)
		}
	}
	// The first call transfers the owner; the remaining calls exercise the
	// shared/exclusive scalar helper chain.
	if len(indexes) >= 3 {
		return indexes[1:]
	}
	return indexes
}

func functionByName(program *core.Program, name string) *core.Function {
	for index := range program.Functions {
		if program.Functions[index].Name == name {
			return &program.Functions[index]
		}
	}
	return nil
}

func TestPhase25ErrorBeforeHelpers(t *testing.T) {
	program := phase25CheckedProgram(t)
	result, err := interp.RunWithForeignOutcomes(program, "main", "opaque-path-token", phase25ModelOutcomes(t, program, "67", true))
	if err != nil {
		t.Fatalf("modeled 0x43 use failure: %v", err)
	}
	if result.Execution.Outcome.Kind != execution.OutcomeTypedFailure || result.Execution.Outcome.Value != "UnsupportedByte" {
		t.Fatalf("error outcome=%+v, want inherited UnsupportedByte typed failure", result.Execution.Outcome)
	}
	for _, event := range result.Execution.Events {
		if event.Kind == "function.called" && (functionName(program, event.CalleeFunctionID) == "shared_copy" || functionName(program, event.CalleeFunctionID) == "exclusive_copy") {
			t.Fatalf("infallible pointer helper ran before typed use failure: %+v", event)
		}
	}
	if len(result.Execution.LiveResources) != 0 || result.ActualHostIO || result.PhysicalCleanup {
		t.Fatalf("error model left resources live or overstated evidence: %+v", result)
	}
}

func functionName(program core.Program, id string) string {
	for _, function := range program.Functions {
		if function.ID == id {
			return function.Name
		}
	}
	return ""
}
