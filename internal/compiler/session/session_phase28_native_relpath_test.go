package session_test

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/protocol"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// TestPhase28NativeCommandRelativePath pins G-28-3: the native command lane
// must reach Check and admitProgram for a RELATIVE source path exactly as it
// does for an absolute one. The test cwd is internal/compiler/session, so the
// repo root is ../../..
func TestPhase28NativeCommandRelativePath(t *testing.T) {
	rel := func(t *testing.T, name string) string {
		t.Helper()
		path := filepath.Join("..", "..", "..", "testdata", "phase07", name)
		if filepath.IsAbs(path) {
			t.Fatalf("path %q must stay relative", path)
		}
		return path
	}

	t.Run("invalid_relative", func(t *testing.T) {
		result, err := session.RunNativeCommandFile(context.Background(), rel(t, "duplicate_function_name.schway"), native.DefaultRunner())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Status != protocol.StatusInvalid {
			t.Fatalf("status = %v diagnostics = %+v, want invalid", result.Status, result.Diagnostics)
		}
		if len(result.Diagnostics) == 0 || result.Diagnostics[0].Code != "core.duplicate_function_id" {
			t.Fatalf("diagnostics = %+v, want first core.duplicate_function_id", result.Diagnostics)
		}
	})

	t.Run("valid_relative", func(t *testing.T) {
		if _, err := exec.LookPath("clang"); err != nil {
			t.Skipf("env:clang toolchain unavailable on this host: %v", err)
		}
		result, err := session.RunNativeCommandFile(context.Background(), rel(t, "call_basic.schway"), native.DefaultRunner())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Status != protocol.StatusPass || len(result.Diagnostics) != 0 {
			t.Fatalf("status = %v diagnostics = %+v, want pass with none", result.Status, result.Diagnostics)
		}
	})

	t.Run("relative_matches_absolute", func(t *testing.T) {
		relative, err := session.RunNativeCommandFile(context.Background(), rel(t, "duplicate_function_name.schway"), native.DefaultRunner())
		if err != nil {
			t.Fatalf("relative: %v", err)
		}
		absolute, err := session.RunNativeCommandFile(context.Background(), testsupport.ProjectPath("testdata", "phase07", "duplicate_function_name.schway"), native.DefaultRunner())
		if err != nil {
			t.Fatalf("absolute: %v", err)
		}
		if relative.Status != absolute.Status {
			t.Fatalf("status relative=%v absolute=%v", relative.Status, absolute.Status)
		}
		if len(relative.Diagnostics) == 0 || len(absolute.Diagnostics) == 0 || relative.Diagnostics[0].Code != absolute.Diagnostics[0].Code {
			t.Fatalf("diagnostics relative=%+v absolute=%+v", relative.Diagnostics, absolute.Diagnostics)
		}
	})
}
