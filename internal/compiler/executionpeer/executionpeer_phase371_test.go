package executionpeer_test

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/executionpeer"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const phase371AdlerDefectFunction = `fn adler32(bytes: ByteView) -> U64 {
  var n = byte_len(bytes)
  if n < 1 {
    defect "legacy child defect"
  } else {
    n = n + 0
  }
  n
}
`

func TestPhase371LegacyDefectClosesAllActiveFrames(t *testing.T) {
	program, entry, leaf, call, sourceDigest := phase371LegacyDefectProgram(t)
	t.Logf("phase371 source witness sha256=%x", sourceDigest)
	if leaf.Parameter.Type != "ByteView" || leaf.ReturnType != "U64" {
		t.Fatalf("callee signature = %s -> %s, want the legacy non-scalar ByteView -> U64 route", leaf.Parameter.Type, leaf.ReturnType)
	}

	defectRun, err := interp.RunWithFileBytes(program, "main", "phase371-empty", func(string) ([]byte, string) {
		return []byte{}, ""
	})
	if err != nil {
		t.Fatalf("interpreter legacy child defect: %v", err)
	}
	if defectRun.Outcome.Kind != execution.OutcomeDefect || defectRun.Outcome.Value != "" {
		t.Fatalf("defect outcome = %+v, want an empty-valued defect", defectRun.Outcome)
	}
	if len(defectRun.Events) == 0 {
		t.Fatal("interpreter legacy child defect produced no events")
	}
	last := defectRun.Events[len(defectRun.Events)-1]
	child, err := execution.FormatInvocation(entry.ID, []execution.InvocationSegment{{OpCallID: call.ID, Ordinal: 0}})
	if err != nil {
		t.Fatalf("format legacy child invocation: %v", err)
	}
	if last.Kind != "function.defected" || last.FunctionID != leaf.ID || last.Invocation != child || last.Output != "legacy child defect" {
		t.Fatalf("final interpreter event = %+v, want leaf %q at %q with the exact defect reason", last, leaf.ID, child)
	}
	peerDocument, err := session.ProjectExecutionSchema2(program, defectRun)
	if err != nil {
		t.Fatalf("project legacy defect trace: %v", err)
	}
	if err := executionpeer.Validate(program, peerDocument); err != nil {
		t.Fatalf("peer rejected actual legacy child defect trace: %v", err)
	}

	returnedRun, err := interp.RunWithFileBytes(program, "main", "phase371-nonempty", func(string) ([]byte, string) {
		return []byte("x"), ""
	})
	if err != nil {
		t.Fatalf("interpreter nested return: %v", err)
	}
	if returnedRun.Outcome.Kind != execution.OutcomeReturned || returnedRun.Outcome.Value != "1" {
		t.Fatalf("nested return outcome = %+v, want returned 1", returnedRun.Outcome)
	}
	returnedPeerDocument, err := session.ProjectExecutionSchema2(program, returnedRun)
	if err != nil {
		t.Fatalf("project nested return trace: %v", err)
	}
	if err := executionpeer.Validate(program, returnedPeerDocument); err != nil {
		t.Fatalf("peer rejected ordinary nested return trace: %v", err)
	}
}

func phase371LegacyDefectProgram(t *testing.T) (core.Program, core.Function, core.Function, core.LinearOperation, [32]byte) {
	t.Helper()
	root := t.TempDir()
	entryPath := testsupport.ProjectPath("examples", "checksum", "checksum.schway")
	entrySource, err := os.ReadFile(entryPath)
	if err != nil {
		t.Fatalf("read actual checksum entry source: %v", err)
	}
	moduleSource, err := os.ReadFile(testsupport.ProjectPath("examples", "checksum", "adler32.schway"))
	if err != nil {
		t.Fatalf("read actual Adler module source: %v", err)
	}
	const functionStart = "fn adler32(bytes: ByteView) -> U64 {"
	if strings.Count(string(moduleSource), functionStart) != 1 || !strings.HasSuffix(string(moduleSource), "}\n") {
		t.Fatal("actual adler32.schway no longer has the expected single final function")
	}
	start := strings.Index(string(moduleSource), functionStart)
	moduleSource = append(append([]byte(nil), moduleSource[:start]...), []byte(phase371AdlerDefectFunction)...)
	for _, row := range []struct {
		name   string
		source []byte
	}{{"checksum.schway", entrySource}, {"adler32.schway", moduleSource}} {
		if err := os.WriteFile(filepath.Join(root, row.name), row.source, 0o600); err != nil {
			t.Fatalf("write temporary checked-source witness %s: %v", row.name, err)
		}
	}
	checked, err := session.CheckFile(filepath.Join(root, "checksum.schway"))
	if err != nil || len(checked.Diagnostics) != 0 {
		t.Fatalf("check temporary checksum source: err=%v diagnostics=%+v", err, checked.Diagnostics)
	}
	var entry, leaf core.Function
	for _, function := range checked.Program.Functions {
		switch function.Name {
		case "main":
			entry = function
		case "adler32":
			leaf = function
		}
	}
	if entry.ID == "" || entry.Linear == nil || leaf.ID == "" || leaf.Linear == nil {
		t.Fatalf("checked entry/helper missing: main=%+v adler32=%+v", entry, leaf)
	}
	var call core.LinearOperation
	for _, operation := range entry.Linear.Operations {
		if operation.Kind == core.OpCall && operation.CalleeID == leaf.ID {
			if call.ID != "" {
				t.Fatalf("entry has duplicate calls to the child: %q and %q", call.ID, operation.ID)
			}
			call = operation
		}
	}
	if call.ID == "" {
		t.Fatal("checked entry has no direct call to the non-scalar child")
	}
	boundSource := append(append([]byte(nil), entrySource...), moduleSource...)
	return checked.Program, entry, leaf, call, sha256.Sum256(boundSource)
}
