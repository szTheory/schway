package native

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestPhase22BuildRetainsRelocatableArtifactWithoutLaunching(t *testing.T) {
	root := t.TempDir()
	marker := filepath.Join(root, "launches.txt")
	artifact := filepath.Join(root, "build", "identity")
	cSource := "#include <stdio.h>\n" +
		"int main(int argc, char **argv) {\n" +
		"  if (argc != 2) return 64;\n" +
		"  FILE *marker = fopen(" + cString(marker) + ", \"a\");\n" +
		"  if (marker == NULL) return 90;\n" +
		"  fputs(\"launch\\n\", marker);\n" +
		"  fclose(marker);\n" +
		"  printf(\"%s\\n\", argv[1]);\n" +
		"  return 0;\n}\n"

	receipt, err := DefaultRunner().BuildApplication(context.Background(), []byte("identity source"), cSource, artifact)
	if err != nil {
		t.Fatalf("BuildApplication: %v", err)
	}
	if _, err := os.Stat(marker); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("build started the application: marker stat error=%v", err)
	}
	if receipt.Cacheable || receipt.DependencyClosure != "incomplete" {
		t.Fatalf("receipt overstates dependency closure: %+v", receipt)
	}
	if len(receipt.RuntimeDependencies) != 1 || receipt.RuntimeDependencies[0] != "platform-c-runtime" {
		t.Fatalf("runtime dependencies=%v, want platform-c-runtime", receipt.RuntimeDependencies)
	}
	for _, path := range []string{artifact, applicationReceiptPath(artifact)} {
		info, err := os.Stat(path)
		if err != nil || (path == artifact && info.Mode()&0o111 == 0) {
			t.Fatalf("retained output %s missing or not executable: info=%v err=%v", path, info, err)
		}
	}
	entries, err := os.ReadDir(filepath.Dir(artifact))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("staging directory survived build cleanup: entries=%v", entries)
	}

	// Relocate both adjacent outputs after the private build staging directory
	// has been removed; neither receipt identity nor generated code binds the
	// artifact to its original checkout/build path.
	relocatedDirectory := filepath.Join(root, "relocated")
	if err := os.Mkdir(relocatedDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	relocated := filepath.Join(relocatedDirectory, "identity")
	if err := os.Rename(artifact, relocated); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(applicationReceiptPath(artifact), applicationReceiptPath(relocated)); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	outcome, err := DefaultRunner().RunApplication(context.Background(), relocated, "7", &stdout, &stderr)
	if err != nil || outcome.Kind != RunExited || outcome.ExitCode != 0 {
		t.Fatalf("RunApplication outcome=%+v err=%v", outcome, err)
	}
	if stdout.String() != "7\n" || stderr.Len() != 0 {
		t.Fatalf("streams stdout=%q stderr=%q, want 7 newline and empty stderr", stdout.String(), stderr.String())
	}
	launches, err := os.ReadFile(marker)
	if err != nil || string(launches) != "launch\n" {
		t.Fatalf("application launches=%q err=%v, want one launch", launches, err)
	}
}

func TestPhase22RunApplicationPreservesStreamsAndProcessOutcomes(t *testing.T) {
	t.Run("raw streams and child exit", func(t *testing.T) {
		artifact := writePhase22Script(t, "printf 'ordinary {not json}\\n'\nprintf 'child diagnostic\\n' >&2\nexit 23\n")
		var stdout, stderr bytes.Buffer
		outcome, err := DefaultRunner().RunApplication(context.Background(), artifact, "opaque", &stdout, &stderr)
		if err != nil || outcome.Kind != RunExited || outcome.ExitCode != 23 {
			t.Fatalf("outcome=%+v err=%v", outcome, err)
		}
		if stdout.String() != "ordinary {not json}\n" || stderr.String() != "child diagnostic\n" {
			t.Fatalf("streams stdout=%q stderr=%q", stdout.String(), stderr.String())
		}
	})

	t.Run("signal", func(t *testing.T) {
		artifact := writePhase22Script(t, "kill -TERM $$\n")
		outcome, err := DefaultRunner().RunApplication(context.Background(), artifact, "signal", &bytes.Buffer{}, &bytes.Buffer{})
		if err != nil || outcome.Kind != RunSignaled || outcome.SignalNumber != int(syscall.SIGTERM) || outcome.Signal == "" {
			t.Fatalf("signal outcome=%+v err=%v", outcome, err)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		artifact := writePhase22Script(t, "exec /bin/sleep 5\n")
		runner := Runner{Timeout: 40 * time.Millisecond}
		outcome, err := runner.RunApplication(context.Background(), artifact, "slow", &bytes.Buffer{}, &bytes.Buffer{})
		if err != nil || outcome.Kind != RunTimedOut || outcome.Diagnostic == "" {
			t.Fatalf("timeout outcome=%+v err=%v", outcome, err)
		}
	})

	t.Run("launch error", func(t *testing.T) {
		artifact := writePhase22Script(t, "exit 0\n")
		runner := Runner{command: func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
			return exec.CommandContext(ctx, filepath.Join(t.TempDir(), "missing-child"))
		}}
		outcome, err := runner.RunApplication(context.Background(), artifact, "input", &bytes.Buffer{}, &bytes.Buffer{})
		if err != nil || outcome.Kind != RunLaunchError || outcome.Diagnostic == "" {
			t.Fatalf("launch outcome=%+v err=%v", outcome, err)
		}
	})

	t.Run("missing executable fails receipt check", func(t *testing.T) {
		artifact := writePhase22Script(t, "exit 0\n")
		if err := os.Remove(artifact); err != nil {
			t.Fatal(err)
		}
		_, err := DefaultRunner().RunApplication(context.Background(), artifact, "input", &bytes.Buffer{}, &bytes.Buffer{})
		var toolError *ToolError
		if !errors.As(err, &toolError) || toolError.Code != "native.artifact_unreadable" {
			t.Fatalf("missing artifact error=%v, want native.artifact_unreadable", err)
		}
	})

	t.Run("changed executable fails digest check", func(t *testing.T) {
		artifact := writePhase22Script(t, "exit 0\n")
		if err := os.WriteFile(artifact, []byte("changed\n"), 0o700); err != nil {
			t.Fatal(err)
		}
		_, err := DefaultRunner().RunApplication(context.Background(), artifact, "input", &bytes.Buffer{}, &bytes.Buffer{})
		var toolError *ToolError
		if !errors.As(err, &toolError) || toolError.Code != "native.artifact_digest_mismatch" {
			t.Fatalf("digest mismatch error=%v", err)
		}
	})

	t.Run("transport bound refuses before launch", func(t *testing.T) {
		marker := filepath.Join(t.TempDir(), "launches.txt")
		artifact := writePhase22Script(t, fmt.Sprintf("printf x >> %s\n", shellQuote(marker)))
		_, err := DefaultRunner().RunApplication(context.Background(), artifact, strings.Repeat("x", MaxApplicationArgumentBytes+1), &bytes.Buffer{}, &bytes.Buffer{})
		var toolError *ToolError
		if !errors.As(err, &toolError) || toolError.Code != "native.input_too_long" {
			t.Fatalf("oversize error=%v", err)
		}
		if _, err := os.Stat(marker); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("oversize input launched application: %v", err)
		}
	})
}

func TestPhase22ConcurrentRequestsLaunchIndependently(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "launches.txt")
	body := fmt.Sprintf("printf '%%s\\n' \"$1\" >> %s\nif [ \"$1\" = slow ]; then exec /bin/sleep 5; fi\nprintf 'fast\\n'\n", shellQuote(marker))
	artifact := writePhase22Script(t, body)
	ctx, cancel := context.WithCancel(context.Background())
	type result struct {
		outcome RunOutcome
		err     error
	}
	slowDone := make(chan result, 1)
	var slowOut, slowErr bytes.Buffer
	go func() {
		outcome, err := DefaultRunner().RunApplication(ctx, artifact, "slow", &slowOut, &slowErr)
		slowDone <- result{outcome: outcome, err: err}
	}()

	deadline := time.Now().Add(2 * time.Second)
	for {
		data, err := os.ReadFile(marker)
		if err == nil && strings.Contains(string(data), "slow\n") {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("slow request did not launch: marker=%q err=%v", data, err)
		}
		time.Sleep(5 * time.Millisecond)
	}

	var fastOut, fastErr bytes.Buffer
	fast, err := DefaultRunner().RunApplication(context.Background(), artifact, "fast", &fastOut, &fastErr)
	if err != nil || fast.Kind != RunExited || fast.ExitCode != 0 || fastOut.String() != "fast\n" || fastErr.Len() != 0 {
		t.Fatalf("fast request outcome=%+v stdout=%q stderr=%q err=%v", fast, fastOut.String(), fastErr.String(), err)
	}
	cancel()
	slow := <-slowDone
	if slow.err != nil || slow.outcome.Kind != RunSignaled {
		t.Fatalf("canceled request outcome=%+v err=%v", slow.outcome, slow.err)
	}
	launches, err := os.ReadFile(marker)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Fields(string(launches))
	if len(lines) != 2 || strings.Join(lines, ",") != "slow,fast" {
		t.Fatalf("launches=%q, want exactly one independent launch per request", launches)
	}
}

func writePhase22Script(t *testing.T, body string) string {
	t.Helper()
	directory := t.TempDir()
	artifact := filepath.Join(directory, "app")
	script := "#!/bin/sh\nset -eu\n" + body
	if err := os.WriteFile(artifact, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	fileBytes, err := os.ReadFile(artifact)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(fileBytes)
	receipt := BuildReceipt{
		Schema:              ApplicationBuildSchema,
		ExecutableDigest:    hex.EncodeToString(digest[:]),
		SourceDigest:        digestBytes([]byte("test source")),
		EmittedCDigest:      digestBytes([]byte("test C")),
		Compiler:            "/test/clang",
		CompilerVersion:     "test clang",
		Target:              runtime.GOOS + "/" + runtime.GOARCH,
		Flags:               append([]string(nil), applicationFlags...),
		RuntimeDependencies: []string{"platform-c-runtime"},
		DependencyClosure:   "incomplete",
		Cacheable:           false,
	}
	encoded, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(applicationReceiptPath(artifact), encoded, 0o600); err != nil {
		t.Fatal(err)
	}
	return artifact
}

func cString(value string) string {
	quoted, _ := json.Marshal(value)
	return string(quoted)
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
