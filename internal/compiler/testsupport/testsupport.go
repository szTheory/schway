package testsupport

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func ProjectPath(parts ...string) string {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
	return filepath.Join(append([]string{root}, parts...)...)
}

func BuildCLI(t testing.TB) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "lang")
	command := exec.Command("go", "build", "-o", binary, "./cmd/lang")
	command.Dir = ProjectPath()
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, output)
	}
	return binary
}

type CLIResult struct {
	Stdout []byte
	Stderr []byte
	Exit   int
}

func RunCLI(t testing.TB, binary string, environment []string, arguments ...string) CLIResult {
	t.Helper()
	command := exec.Command(binary, arguments...)
	if environment == nil {
		command.Env = os.Environ()
	} else {
		command.Env = environment
	}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	exit := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exit = exitError.ExitCode()
		} else {
			t.Fatalf("run CLI: %v", err)
		}
	}
	return CLIResult{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), Exit: exit}
}
