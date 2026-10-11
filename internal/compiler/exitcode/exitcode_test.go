package exitcode_test

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/exitcode"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/protocol"
)

func TestExitCodeTable(t *testing.T) {
	table := []struct {
		name string
		got  int
		want int
	}{
		{"Success", exitcode.Success, 0},
		{"Unrepaired", exitcode.Unrepaired, 1},
		{"Invalid", exitcode.Invalid, 2},
		{"Operational", exitcode.Operational, 3},
		{"Mismatch", exitcode.Mismatch, 4},
		{"Usage", exitcode.Usage, 64},
		{"AppInputRefused", exitcode.AppInputRefused, 65},
		{"EventCapacity", exitcode.EventCapacity, 69},
		{"AppTimeout", exitcode.AppTimeout, 124},
		{"AppLaunchError", exitcode.AppLaunchError, 125},
		{"SignalBase", exitcode.SignalBase, 128},
	}
	seen := map[int]string{}
	for _, row := range table {
		if row.got != row.want {
			t.Errorf("%s = %d, want %d", row.name, row.got, row.want)
		}
		if prior, dup := seen[row.got]; dup {
			t.Errorf("%s and %s share value %d", row.name, prior, row.got)
		}
		seen[row.got] = row.name
	}
	if len(table) != 11 {
		t.Errorf("table rows = %d, want 11", len(table))
	}
	for n, want := range map[int]int{0: 128, 1: 129, 9: 137} {
		if got := exitcode.SignalExit(n); got != want {
			t.Errorf("SignalExit(%d) = %d, want %d", n, got, want)
		}
	}
	if native.EventCapacityExitStatus != exitcode.EventCapacity {
		t.Errorf("native.EventCapacityExitStatus = %d, want %d", native.EventCapacityExitStatus, exitcode.EventCapacity)
	}
	if exitcode.AppInputRefused != 65 || exitcode.EventCapacity != 69 {
		t.Error("published 65/69 contract values moved")
	}
	for status, want := range map[string]int{
		protocol.StatusPass:        exitcode.Success,
		protocol.StatusInvalid:     exitcode.Invalid,
		protocol.StatusOperational: exitcode.Operational,
		protocol.StatusMismatch:    exitcode.Mismatch,
		protocol.StatusUsage:       exitcode.Usage,
		"":                         exitcode.Usage,
		"bogus":                    exitcode.Usage,
	} {
		if got := protocol.ExitCode(status); got != want {
			t.Errorf("protocol.ExitCode(%q) = %d, want %d", status, got, want)
		}
	}
}
