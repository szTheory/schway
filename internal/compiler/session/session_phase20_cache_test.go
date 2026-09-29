package session_test

import (
	"context"
	"runtime"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cache"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
)

func TestPhase20EnumeratedClosureCacheControls(t *testing.T) {
	programs := session.EnumeratePhase5Closure()
	if len(programs) != 112 {
		t.Fatalf("enumerated closure has %d programs, want exactly 112", len(programs))
	}

	var selectedIndex int
	var source string
	for index, program := range programs {
		candidate, err := session.Phase16ControlNativeC(program, "")
		if err == nil {
			selectedIndex, source = index, candidate
			break
		}
	}
	if source == "" {
		t.Skip("enumerated closure currently has no public native-emitter-accepted program; no native artifact is available to cache; see probe:TestPhase20EnumeratedClosureCacheControls")
	}
	program := programs[selectedIndex]
	inputs, ok := phase5InputsForProgram(program)
	if !ok || len(inputs) == 0 {
		t.Fatal("selected closure program has no deterministic input")
	}
	interpreted, err := interp.Run(program, program.Functions[0].Name, inputs[0])
	if err != nil {
		t.Fatal(err)
	}
	if interpreted.Outcome.Kind != "returned" {
		t.Skip("selected native closure program is not a returned-value cache control; see probe:TestPhase20EnumeratedClosureCacheControls")
	}

	store := &cache.Store{Root: t.TempDir()}
	runner := native.DefaultRunner()
	runner.BuildCache = store
	first, err := runner.Run(context.Background(), source, "-O0", []string{inputs[0]})
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "darwin" {
		if first.CacheStatus != cache.StatusNotCacheable {
			t.Fatalf("host-specific dependency discovery must refuse caching on %s, got status=%q", runtime.GOOS, first.CacheStatus)
		}
		return
	}
	if first.CacheStatus != cache.StatusArtifactRecomputed {
		t.Fatalf("first closure build status=%q", first.CacheStatus)
	}
	second, err := runner.Run(context.Background(), source, "-O0", []string{inputs[0]})
	if err != nil {
		t.Fatal(err)
	}
	if second.CacheStatus != cache.StatusArtifactReused {
		t.Fatalf("unchanged closure build status=%q", second.CacheStatus)
	}

	changed, err := runner.Run(context.Background(), source+"\n/* seeded source mutation */\n", "-O0", []string{inputs[0]})
	if err != nil {
		t.Fatal(err)
	}
	if changed.CacheStatus != cache.StatusArtifactRecomputed {
		t.Fatalf("changed emitted source status=%q", changed.CacheStatus)
	}

	runner.Expect = native.ExpectTypedFailure
	_, err = runner.Run(context.Background(), source, "-O0", []string{inputs[0]})
	if err == nil {
		t.Fatal("cached executable was accepted under a deliberately wrong current execution expectation")
	}
	if !strings.Contains(err.Error(), "native.invalid_execution") {
		t.Fatalf("wrong current execution failed for an unexpected reason: %v", err)
	}
}
