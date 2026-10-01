package corevalidate

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
)

func phase25FrameDrainProbe(mode string, withContract bool) *core.Function {
	foreign := (*core.ForeignOperationContract)(nil)
	if withContract {
		foreign = &core.ForeignOperationContract{Mode: mode}
	}
	return &core.Function{
		ID: "phase25:main", Name: "main", Parameter: core.Parameter{ID: "phase25:source", Type: "PathToken"}, ReturnType: "U64",
		Linear: &core.LinearBody{
			Places: []core.Place{
				{ID: "phase25:source", TypeID: "path"},
				{ID: "phase25:foreign-result", TypeID: "u64"},
				{ID: "phase25:shared-result", TypeID: "u64"},
				{ID: "phase25:exclusive-result", TypeID: "u64"},
				{ID: "phase25:abandoned", TypeID: "u64"},
			},
			Operations: []core.LinearOperation{
				{ID: "phase25:foreign", Kind: core.OpForeignCall, SourceID: "phase25:source", TargetID: "phase25:foreign-result", Foreign: foreign},
				{ID: "phase25:shared", Kind: core.OpCall, SourceID: "phase25:foreign-result", TargetID: "phase25:shared-result"},
				{ID: "phase25:exclusive", Kind: core.OpCall, SourceID: "phase25:shared-result", TargetID: "phase25:exclusive-result"},
				{ID: "phase25:return", Kind: core.OpReturn, SourceID: "phase25:exclusive-result"},
			},
		},
	}
}

func TestPeerCalleeFrameDrained(t *testing.T) {
	tests := []struct {
		name         string
		mode         string
		withContract bool
		want         bool
	}{
		{name: "borrow result may feed scalar helper calls", mode: "borrow", withContract: true, want: true},
		{name: "consume is not a frame acquisition", mode: "consume", withContract: true, want: true},
		{name: "unknown mode remains conservatively tracked", mode: "unknown", withContract: true, want: false},
		{name: "legacy missing contract remains conservatively tracked", withContract: false, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			function := phase25FrameDrainProbe(test.mode, test.withContract)
			if got := peerCalleeFrameDrained(function); got != test.want {
				t.Fatalf("peerCalleeFrameDrained(mode=%q, contract=%v) = %v, want %v", test.mode, test.withContract, got, test.want)
			}
		})
	}

	abandoned := phase25FrameDrainProbe("acquire", true)
	abandoned.Linear.Operations[0].TargetID = "phase25:abandoned"
	abandoned.Linear.Operations[1].SourceID = "phase25:foreign-result"
	if peerCalleeFrameDrained(abandoned) {
		t.Fatal("unreleased acquisition was admitted after the non-acquiring mode exception")
	}
}
