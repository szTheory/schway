package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

// phase30FakeLang writes a stand-in `lang` binary that prints a canned check
// document on the first invocation and a passing one afterwards, so the
// driver's decision is observed through its real subprocess path.
func phase30FakeLang(t *testing.T, dir, firstDoc string) string {
	t.Helper()
	first := filepath.Join(dir, "first.json")
	if err := os.WriteFile(first, []byte(firstDoc), 0o644); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(dir, "ran")
	script := "#!/bin/sh\nif [ -f '" + marker + "' ]; then\n  printf '%s' '{\"status\":\"pass\",\"diagnostics\":[]}'\nelse\n  : > '" + marker + "'\n  cat '" + first + "'\nfi\n"
	bin := filepath.Join(dir, "fakelang")
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return bin
}

func phase30Run(t *testing.T, doc string) (Outcome, []byte, []byte) {
	t.Helper()
	dir := t.TempDir()
	source := filepath.Join(dir, "main.schway")
	original := []byte("0123456789")
	if err := os.WriteFile(source, original, 0o644); err != nil {
		t.Fatal(err)
	}
	bin := phase30FakeLang(t, dir, doc)
	outcome, err := Repair(context.Background(), bin, source)
	if err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	return outcome, original, after
}

func phase30Doc(diagUnit, repairUnit string) string {
	unit := func(u string) string {
		if u == "" {
			return ""
		}
		return `,"unit":"` + u + `"`
	}
	body := `"repairs":[{"kind":"fix","span":{"start":2,"end":4},"replacement":"XX","applicability":"MachineApplicable"` + unit(repairUnit) + `}]`
	return `{"status":"invalid","diagnostics":[{"code":"x.y"` + unit(diagUnit) + `,` + body + `}]}`
}

// TestPhase30RepairDriverDeclinesCrossUnit pins D-23 / T-30-07: a repair
// attributed to another unit is never spliced into the entry file.
func TestPhase30RepairDriverDeclinesCrossUnit(t *testing.T) {
	t.Run("unit-bearing diagnostic is declined", func(t *testing.T) {
		outcome, before, after := phase30Run(t, phase30Doc("adler32", ""))
		if outcome.Status != OutcomeUnrepairable || outcome.DeclineReason != DeclineNoneEligible {
			t.Fatalf("outcome = %+v, want unrepairable/%s", outcome, DeclineNoneEligible)
		}
		if outcome.SubprocessCount != 1 {
			t.Fatalf("subprocess count %d, want 1", outcome.SubprocessCount)
		}
		if !bytes.Equal(before, after) {
			t.Fatalf("entry file changed: %q -> %q", before, after)
		}
	})
	t.Run("empty unit applies the repair", func(t *testing.T) {
		outcome, before, after := phase30Run(t, phase30Doc("", ""))
		if outcome.Status != OutcomeRepaired {
			t.Fatalf("outcome = %+v, want repaired", outcome)
		}
		if bytes.Equal(before, after) || string(after) != "01XX456789" {
			t.Fatalf("entry file = %q", after)
		}
	})
	t.Run("repair-level unit inside an entry diagnostic is declined", func(t *testing.T) {
		outcome, before, after := phase30Run(t, phase30Doc("", "lib"))
		if outcome.Status != OutcomeUnrepairable || outcome.DeclineReason != DeclineNoneEligible {
			t.Fatalf("outcome = %+v", outcome)
		}
		if !bytes.Equal(before, after) {
			t.Fatalf("entry file changed: %q -> %q", before, after)
		}
	})
	t.Run("control: ignoring unit patches the entry file", func(t *testing.T) {
		ignoreUnitForTest = true
		defer func() { ignoreUnitForTest = false }()
		outcome, before, after := phase30Run(t, phase30Doc("adler32", ""))
		if outcome.Status != OutcomeRepaired || bytes.Equal(before, after) {
			t.Fatalf("control did not observe a patch: outcome=%+v after=%q", outcome, after)
		}
	})
}
