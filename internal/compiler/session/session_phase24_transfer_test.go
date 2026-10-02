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

func TestPhase24TransferPeerValidatesAcquisitionDerivedDischarge(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("examples", "phase24", "transfer.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture rejected: %+v", checked.Diagnostics)
	}
	assertTransferPeers(t, checked.Program, true)
}

func TestPhase24TransferPeerRejectsMutatedOwnership(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("examples", "phase24", "transfer.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture rejected: %+v", checked.Diagnostics)
	}
	mutations := map[string]func(*core.Program){
		"copied owner": func(p *core.Program) {
			c := caller(p)
			owner := c.Linear.Places[1]
			owner.ID = "synthetic-owner-copy"
			owner.Name = "owner_copy"
			c.Linear.Places = append(c.Linear.Places, owner)
			copyOp := core.LinearOperation{ID: "synthetic-owner-copy-op", PointID: "synthetic-owner-copy-point", Kind: core.OpCopy, SourceID: c.Linear.Operations[0].TargetID, TargetID: owner.ID, TypeID: owner.TypeID}
			c.Linear.Operations = append(c.Linear.Operations[:1], append([]core.LinearOperation{copyOp}, c.Linear.Operations[1:]...)...)
		},
		"stale owner": func(p *core.Program) { caller(p).Linear.Operations[1].SourceID = "stale-owner" },
		"all candidate releases deleted": func(p *core.Program) {
			ops := caller(p).Linear.Operations[:0]
			for _, op := range caller(p).Linear.Operations {
				if op.Kind != core.OpRelease {
					ops = append(ops, op)
				}
			}
			caller(p).Linear.Operations = ops
		},
		"duplicate release": func(p *core.Program) {
			c := caller(p)
			for _, op := range c.Linear.Operations {
				if op.Kind == core.OpRelease {
					dup := op
					dup.ID += ":duplicate"
					dup.PointID += ":duplicate"
					c.Linear.Operations = append(c.Linear.Operations, dup)
					return
				}
			}
		},
		"wrong acquisition identity": func(p *core.Program) {
			for i := range caller(p).Linear.Operations {
				if caller(p).Linear.Operations[i].Kind == core.OpRelease {
					caller(p).Linear.Operations[i].ReleasesOperationID = "invented-acquire"
				}
			}
		},
		"entry owner escape": func(p *core.Program) {
			c := caller(p)
			c.ReturnType = "FileByteOwner"
			for i := range c.Linear.Operations {
				if c.Linear.Operations[i].Kind == core.OpReturn {
					c.Linear.Operations[i].SourceID = c.Linear.Operations[0].TargetID
				}
			}
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			p := cloneTransferProgram(checked.Program)
			mutate(&p)
			assertTransferPeers(t, p, false)
		})
	}
}

func assertTransferPeers(t *testing.T, program core.Program, wantValid bool) {
	t.Helper()
	coreValid := corevalidate.Validate(program).Valid
	originValid := len(originvalidate.ValidatePublished(program)) == 0
	pathValid := pathoracle.ValidateLocalOwnerPaths(program) == nil
	if coreValid != wantValid || originValid != wantValid || pathValid != wantValid {
		t.Fatalf("peer disagreement: core=%v origin=%v path=%v want=%v", coreValid, originValid, pathValid, wantValid)
	}
}

func caller(program *core.Program) *core.Function {
	for i := range program.Functions {
		if program.Functions[i].Name == "main" {
			return &program.Functions[i]
		}
	}
	panic("fixture main missing")
}

func cloneTransferProgram(in core.Program) core.Program {
	out := in
	out.Functions = append([]core.Function(nil), in.Functions...)
	for i := range out.Functions {
		if in.Functions[i].Linear == nil {
			continue
		}
		linear := *in.Functions[i].Linear
		linear.Types = append([]core.TypeFact(nil), linear.Types...)
		linear.Places = append([]core.Place(nil), linear.Places...)
		linear.Operations = append([]core.LinearOperation(nil), linear.Operations...)
		linear.Blocks = append([]core.Block(nil), linear.Blocks...)
		linear.Edges = append([]core.Edge(nil), linear.Edges...)
		for j := range linear.Operations {
			if linear.Operations[j].Foreign != nil {
				contract := *linear.Operations[j].Foreign
				linear.Operations[j].Foreign = &contract
			}
		}
		out.Functions[i].Linear = &linear
	}
	return out
}
