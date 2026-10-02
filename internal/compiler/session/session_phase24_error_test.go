package session_test

import (
	"os"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/pathoracle"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase24ActivationPeerDerivesRepeatedAcquisitions(t *testing.T) {
	program := checkedPhase24ErrorProgram(t)
	main, probe, helper := phase24Named(&program, "main"), phase24Named(&program, "probe"), phase24Named(&program, "acquire")
	var acquisitionID string
	for _, operation := range helper.Linear.Operations {
		if operation.Foreign != nil && operation.Foreign.Mode == "acquire" {
			acquisitionID = operation.ID
		}
	}
	if acquisitionID == "" {
		t.Fatal("helper acquisition missing")
	}
	var callIDs []string
	for _, function := range []*core.Function{main, probe} {
		for _, operation := range function.Linear.Operations {
			if operation.Kind == core.OpCall && operation.CalleeID == helper.ID {
				callIDs = append(callIDs, operation.ID)
			}
		}
	}
	if len(callIDs) != 3 || callIDs[0] == callIDs[1] || callIDs[0] == callIDs[2] || callIDs[1] == callIDs[2] {
		t.Fatalf("helper call activations=%v; want three distinct static call sites sharing %s", callIDs, acquisitionID)
	}
	assertErrorPeers(t, program, true)
}

func TestPhase24CleanupPeerRejectsOwnershipAndOrderMutations(t *testing.T) {
	baseline := checkedPhase24ErrorProgram(t)
	mutations := map[string]func(*core.Program){
		"all releases deleted": func(program *core.Program) {
			for index := range program.Functions {
				function := &program.Functions[index]
				if function.Linear == nil {
					continue
				}
				removed := map[string]bool{}
				ops := function.Linear.Operations[:0]
				for _, operation := range function.Linear.Operations {
					if operation.Kind == core.OpRelease {
						removed[operation.ID] = true
						continue
					}
					ops = append(ops, operation)
				}
				function.Linear.Operations = ops
				for blockIndex := range function.Linear.Blocks {
					ids := function.Linear.Blocks[blockIndex].OperationIDs[:0]
					for _, id := range function.Linear.Blocks[blockIndex].OperationIDs {
						if !removed[id] {
							ids = append(ids, id)
						}
					}
					function.Linear.Blocks[blockIndex].OperationIDs = ids
				}
			}
		},
		"activation identity collision": func(program *core.Program) {
			main := phase24Named(program, "main")
			calls := phase24Calls(main, phase24Named(program, "acquire").ID)
			main.Linear.Operations[phase24OperationIndex(main, calls[1].ID)].ID = calls[0].ID
		},
		"wrong resource pairing": func(program *core.Program) {
			main := phase24Named(program, "main")
			for index := range main.Linear.Operations {
				if main.Linear.Operations[index].Kind == core.OpRelease {
					main.Linear.Operations[index].ReleasesOperationID = "forged-acquisition"
					return
				}
			}
		},
		"reversed cleanup order": func(program *core.Program) {
			main := phase24Named(program, "main")
			call := phase24Calls(main, phase24Named(program, "probe").ID)[0]
			block := phase24EdgeBlock(main, call.ErrEdgeID)
			block.OperationIDs[0], block.OperationIDs[1] = block.OperationIDs[1], block.OperationIDs[0]
		},
		"failed acquisition creates phantom owner": func(program *core.Program) {
			main := phase24Named(program, "main")
			acquire := phase24Calls(main, phase24Named(program, "acquire").ID)[1]
			block := phase24EdgeBlock(main, acquire.ErrEdgeID)
			for index := range main.Linear.Operations {
				operation := &main.Linear.Operations[index]
				if operation.Kind == core.OpRelease && operation.SourceID == acquire.TargetID {
					block.OperationIDs[0] = operation.ID
					return
				}
			}
		},
		"callee drains caller frame": func(program *core.Program) {
			probe := phase24Named(program, "probe")
			main := phase24Named(program, "main")
			callerOwner := phase24Calls(main, phase24Named(program, "acquire").ID)[0].TargetID
			for index := range probe.Linear.Operations {
				if probe.Linear.Operations[index].Kind == core.OpRelease {
					probe.Linear.Operations[index].SourceID = callerOwner
					return
				}
			}
		},
		"caller replaces propagated error": func(program *core.Program) {
			main := phase24Named(program, "main")
			callProbe := phase24Calls(main, phase24Named(program, "probe").ID)[0]
			callA := phase24Calls(main, phase24Named(program, "acquire").ID)[0]
			block := phase24EdgeBlock(main, callProbe.ErrEdgeID)
			terminalID := block.OperationIDs[len(block.OperationIDs)-1]
			main.Linear.Operations[phase24OperationIndex(main, terminalID)].SourceID = callA.ErrTargetID
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			program := cloneTransferProgram(baseline)
			mutate(&program)
			assertErrorPeers(t, program, false)
		})
	}
}

func checkedPhase24ErrorProgram(t *testing.T) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("examples", "phase24", "error.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("Phase 24 error source rejected: %+v", checked.Diagnostics)
	}
	return checked.Program
}

func assertErrorPeers(t *testing.T, program core.Program, wantValid bool) {
	t.Helper()
	coreValid := corevalidate.Validate(program).Valid
	originValid := len(originvalidate.ValidatePublished(program)) == 0
	pathValid := pathoracle.ValidateLocalOwnerPaths(program) == nil
	if coreValid != wantValid || originValid != wantValid || pathValid != wantValid {
		t.Fatalf("independent error peers core=%v origin=%v path=%v want=%v", coreValid, originValid, pathValid, wantValid)
	}
}

func phase24Named(program *core.Program, name string) *core.Function {
	for index := range program.Functions {
		if program.Functions[index].Name == name {
			return &program.Functions[index]
		}
	}
	panic("Phase 24 function missing: " + name)
}

func phase24Calls(function *core.Function, calleeID string) []core.LinearOperation {
	var calls []core.LinearOperation
	for _, operation := range function.Linear.Operations {
		if operation.Kind == core.OpCall && operation.CalleeID == calleeID {
			calls = append(calls, operation)
		}
	}
	return calls
}

func phase24OperationIndex(function *core.Function, operationID string) int {
	for index, operation := range function.Linear.Operations {
		if operation.ID == operationID {
			return index
		}
	}
	return -1
}

func phase24EdgeBlock(function *core.Function, edgeID string) *core.Block {
	for _, edge := range function.Linear.Edges {
		if edge.ID != edgeID {
			continue
		}
		for index := range function.Linear.Blocks {
			if function.Linear.Blocks[index].ID == edge.ToBlockID {
				return &function.Linear.Blocks[index]
			}
		}
	}
	panic("Phase 24 edge block missing: " + edgeID)
}
