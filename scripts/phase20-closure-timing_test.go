package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestPhase20ClosureTimingHarness(t *testing.T) {
	t.Run("three paired runs use one private cache per cold warm pair", func(t *testing.T) {
		var calls []sample
		fake := func(_ time.Time, _ string, label, cacheRoot, name string, args ...string) sample {
			got := sample{Label: label, Command: append([]string{name}, args...), Cache: cacheRoot, ExitCode: 0}
			calls = append(calls, got)
			return got
		}
		got, err := runFullSuitePairsWith(time.Now().Add(time.Minute), "/go-cache", func() bool { return false }, fake)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 6 || len(calls) != 6 {
			t.Fatalf("runs=%d calls=%d, want six", len(got), len(calls))
		}
		wantLabels := []string{"full-1-cold", "full-1-warm", "full-2-cold", "full-2-warm", "full-3-cold", "full-3-warm"}
		var labels []string
		for _, call := range got {
			labels = append(labels, call.Label)
			if !reflect.DeepEqual(call.Command, []string{"go", "test", "./...", "-count=1"}) {
				t.Errorf("%s command=%v", call.Label, call.Command)
			}
		}
		if !reflect.DeepEqual(labels, wantLabels) {
			t.Fatalf("labels=%v, want %v", labels, wantLabels)
		}
		for i := 0; i < 6; i += 2 {
			if got[i].Cache == "" || got[i].Cache != got[i+1].Cache {
				t.Errorf("pair %d cold/warm cache roots differ or are empty: %q %q", i/2+1, got[i].Cache, got[i+1].Cache)
			}
			if i >= 2 && got[i].Cache == got[i-2].Cache {
				t.Errorf("pair %d reused prior pair cache root %q", i/2+1, got[i].Cache)
			}
		}
	})

	t.Run("failed or timed out cold sample prevents warm run", func(t *testing.T) {
		calls := 0
		fake := func(_ time.Time, _ string, label, cacheRoot, _ string, _ ...string) sample {
			calls++
			return sample{Label: label, Cache: cacheRoot, ExitCode: 1, TimedOut: true}
		}
		got, err := runFullSuitePairsWith(time.Now().Add(time.Minute), "/go-cache", nil, fake)
		if err == nil || !strings.Contains(err.Error(), "full-1-cold") {
			t.Fatalf("error=%v, want cold failure", err)
		}
		if calls != 1 || len(got) != 1 {
			t.Fatalf("calls=%d samples=%d; must stop before warm run", calls, len(got))
		}
	})

	t.Run("input drift stops after the measured cold sample", func(t *testing.T) {
		calls, checks := 0, 0
		fake := func(_ time.Time, _ string, label, cacheRoot, _ string, _ ...string) sample {
			calls++
			return sample{Label: label, Cache: cacheRoot, ExitCode: 0}
		}
		got, err := runFullSuitePairsWith(time.Now().Add(time.Minute), "/go-cache", func() bool { checks++; return true }, fake)
		if err == nil || !strings.Contains(err.Error(), "source inputs changed") {
			t.Fatalf("error=%v, want source drift", err)
		}
		if calls != 1 || checks != 1 || len(got) != 1 {
			t.Fatalf("calls=%d checks=%d samples=%d", calls, checks, len(got))
		}
	})

	t.Run("timing report preserves raw samples and blocker", func(t *testing.T) {
		root := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, phaseDir), 0o700); err != nil {
			t.Fatal(err)
		}
		old, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Chdir(root); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chdir(old) })
		samples := []sample{{Label: "full-1-cold", Command: []string{"go", "test", "./...", "-count=1"}, RawDurationNS: 123, Duration: 123 * time.Nanosecond, ExitCode: 1, Output: "red"}}
		writeReport(samples, machineFacts{OS: "darwin", Arch: "arm64"}, "test-machine", "head", "sha", time.Now(), "seeded red sample")
		data, err := os.ReadFile(filepath.Join(root, phaseDir, "20-CLOSURE-TIMING.md"))
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{"123 (0.000)", "seeded red sample", "go test ./... -count=1", "Status:** BLOCKED"} {
			if !strings.Contains(string(data), want) {
				t.Errorf("timing report missing %q", want)
			}
		}
	})

	if !failSample(sample{ExitCode: 0, TimedOut: true}) {
		t.Fatal("timed-out sample was accepted")
	}
}
