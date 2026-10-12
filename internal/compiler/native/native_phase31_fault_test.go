package native_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const phase31FaultInclude = "#include \"adapter.h\"\n"

const phase31FaultHooks = "#include <stddef.h>\n#include <sys/stat.h>\n#include <sys/types.h>\n\n" +
	"int phase31_observer_open(const char *, int);\n" +
	"int phase31_observer_fstat(int, struct stat *);\n" +
	"ssize_t phase31_observer_read(int, void *, size_t);\n" +
	"void *phase31_observer_malloc(size_t);\n" +
	"void *phase31_observer_realloc(void *, size_t);\n" +
	"void phase31_observer_free(void *);\n" +
	"int phase31_observer_close(int);\n\n" +
	"#define SCHWAY_FILE_BYTES_OPEN phase31_observer_open\n" +
	"#define SCHWAY_FILE_BYTES_FSTAT phase31_observer_fstat\n" +
	"#define SCHWAY_FILE_BYTES_READ phase31_observer_read\n" +
	"#define SCHWAY_FILE_BYTES_MALLOC phase31_observer_malloc\n" +
	"#define SCHWAY_FILE_BYTES_REALLOC phase31_observer_realloc\n" +
	"#define SCHWAY_FILE_BYTES_FREE phase31_observer_free\n" +
	"#define SCHWAY_FILE_BYTES_CLOSE phase31_observer_close\n\n"

type phase31FaultReceipt struct {
	Allocations, Frees, Doubles, BytesRead int
}

// phase31FaultReplaceOnce is the exact-count-1 replace contract: an anchor that
// is absent or ambiguous fails the test instead of silently doing nothing.
func phase31FaultReplaceOnce(t *testing.T, text, old, replacement string) string {
	t.Helper()
	if n := strings.Count(text, old); n != 1 {
		t.Fatalf("mutant anchor %q occurs %d times in the adapter, want exactly 1", old, n)
	}
	return strings.Replace(text, old, replacement, 1)
}

// phase31FaultBuild builds a TempDir copy of examples/checksum. With hooks the
// counting observer is linked in; the copy is the only thing ever mutated.
func phase31FaultBuild(t *testing.T, hooks bool, mutate func(adapter string) string) string {
	t.Helper()
	phase29RequireClang(t)
	project := testsupport.ProjectPath("examples", "checksum")
	root := t.TempDir()
	entries, err := os.ReadDir(project)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		data, err := os.ReadFile(filepath.Join(project, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, entry.Name()), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	adapterPath := filepath.Join(root, "adapter.c")
	adapterBytes, err := os.ReadFile(adapterPath)
	if err != nil {
		t.Fatal(err)
	}
	adapter := string(adapterBytes)
	manifestPath := filepath.Join(root, "file_bytes.bindings.json")
	if hooks {
		adapter = phase31FaultReplaceOnce(t, adapter, phase31FaultInclude, phase31FaultHooks+phase31FaultInclude)
		observer, err := os.ReadFile(testsupport.ProjectPath("internal", "compiler", "native", "testdata", "phase31_observer.c"))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "phase31_observer.c"), observer, 0o600); err != nil {
			t.Fatal(err)
		}
		manifestBytes, err := os.ReadFile(manifestPath)
		if err != nil {
			t.Fatal(err)
		}
		var manifest native.BindingManifest
		if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
			t.Fatal(err)
		}
		manifest.Sources = append(manifest.Sources, "phase31_observer.c")
		manifestBytes, err = json.Marshal(manifest)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(manifestPath, manifestBytes, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if mutate != nil {
		adapter = mutate(adapter)
	}
	if err := os.WriteFile(adapterPath, []byte(adapter), 0o600); err != nil {
		t.Fatal(err)
	}
	artifact := filepath.Join(t.TempDir(), "checksum")
	_, diagnostics, err := session.BuildApplicationFile(context.Background(), filepath.Join(root, "checksum.schway"), artifact, native.DefaultRunner(), manifestPath)
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("build hooked checksum: diagnostics=%+v err=%v", diagnostics, err)
	}
	return artifact
}

func phase31FaultRunCtx(t *testing.T, ctx context.Context, artifact, path, fault string) (native.RunOutcome, string, string, phase31FaultReceipt, error) {
	t.Helper()
	receiptPath := filepath.Join(t.TempDir(), "receipt")
	t.Setenv("SCHWAY_PHASE31_FAULT", fault)
	t.Setenv("SCHWAY_PHASE31_RECEIPT", receiptPath)
	var stdout, stderr bytes.Buffer
	outcome, err := native.RunApplication(ctx, artifact, path, &stdout, &stderr)
	var receipt phase31FaultReceipt
	if data, readErr := os.ReadFile(receiptPath); readErr == nil {
		if _, scanErr := fmt.Sscanf(string(data), "allocations=%d frees=%d doubles=%d bytes_read=%d", &receipt.Allocations, &receipt.Frees, &receipt.Doubles, &receipt.BytesRead); scanErr != nil {
			t.Fatalf("receipt %q: %v", data, scanErr)
		}
	}
	return outcome, stdout.String(), stderr.String(), receipt, err
}

func phase31FaultRun(t *testing.T, artifact, path, fault string) (native.RunOutcome, string, string, phase31FaultReceipt) {
	t.Helper()
	outcome, stdout, stderr, receipt, err := phase31FaultRunCtx(t, context.Background(), artifact, path, fault)
	if err != nil {
		t.Fatalf("run %s fault %s: %v", path, fault, err)
	}
	return outcome, stdout, stderr, receipt
}

func phase31FaultWrite(t *testing.T, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "input.bin")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func phase31FaultOracle(t *testing.T, name string) (data []byte, want string) {
	t.Helper()
	for _, row := range testsupport.ChecksumOracleRows() {
		if row.App == "checksum" && row.Name == name {
			return row.Input, fmt.Sprintf("%d\n", row.Want)
		}
	}
	t.Fatalf("oracle row %q missing", name)
	return nil, ""
}

func TestPhase31ReleaseExactlyOnce(t *testing.T) {
	artifact := phase31FaultBuild(t, true, nil)
	wiki, wikiWant := phase31FaultOracle(t, "Wikipedia")
	capData, capWant := phase31FaultOracle(t, "65536x0xFF")
	over, _ := phase31FaultOracle(t, "65537x0xFF")
	type row struct {
		name   string
		data   []byte
		fault  string
		stdout string // success rows
		stderr string // error rows
		check  func(t *testing.T, r phase31FaultReceipt)
	}
	const (
		tooLarge    = "FileError.FileTooLarge\n"
		allocFailed = "FileError.AllocFailed\n"
		unavailable = "FileError.FileUnavailable\n"
	)
	rows := []row{
		{name: "success", data: wiki, fault: "none", stdout: wikiWant, check: func(t *testing.T, r phase31FaultReceipt) {
			if r.Frees < 1 {
				t.Errorf("frees=%d, want >=1", r.Frees)
			}
		}},
		{name: "oversize", data: over, fault: "none", stderr: tooLarge, check: func(t *testing.T, r phase31FaultReceipt) {
			if r.Frees != 1 {
				t.Errorf("frees=%d, want 1", r.Frees)
			}
		}},
		{name: "malloc-null", data: wiki, fault: "malloc-null", stderr: allocFailed, check: func(t *testing.T, r phase31FaultReceipt) {
			if r.Allocations != 0 {
				t.Errorf("allocations=%d, want 0", r.Allocations)
			}
		}},
		{name: "realloc-null", data: wiki, fault: "realloc-null", stderr: allocFailed, check: func(t *testing.T, r phase31FaultReceipt) {
			if r.Frees != 1 {
				t.Errorf("frees=%d, want 1", r.Frees)
			}
		}},
		{name: "fstat-fail", data: wiki, fault: "fstat-fail", stderr: unavailable},
		{name: "read-eio", data: wiki, fault: "read-eio", stderr: unavailable, check: func(t *testing.T, r phase31FaultReceipt) {
			if r.Frees != 1 {
				t.Errorf("frees=%d, want 1", r.Frees)
			}
		}},
		{name: "close-fail", data: wiki, fault: "close-fail", stderr: unavailable, check: func(t *testing.T, r phase31FaultReceipt) {
			if r.Frees != 1 {
				t.Errorf("frees=%d, want 1", r.Frees)
			}
		}},
		{name: "read-eintr-once", data: wiki, fault: "read-eintr-once", stdout: wikiWant},
		{name: "read-one-byte", data: wiki, fault: "read-one-byte", stdout: wikiWant},
		{name: "cap-65536", data: capData, fault: "none", stdout: capWant},
		{name: "one-mebibyte", data: bytes.Repeat([]byte{0xFF}, 1<<20), fault: "none", stderr: tooLarge, check: func(t *testing.T, r phase31FaultReceipt) {
			if r.BytesRead > 65537 {
				t.Errorf("bytes_read=%d, want <= 65537", r.BytesRead)
			}
		}},
	}
	for _, tc := range rows {
		t.Run(tc.name, func(t *testing.T) {
			outcome, stdout, stderr, receipt := phase31FaultRun(t, artifact, phase31FaultWrite(t, tc.data), tc.fault)
			if tc.stderr == "" {
				if outcome.Kind != native.RunExited || outcome.ExitCode != 0 || stdout != tc.stdout || stderr != "" {
					t.Fatalf("outcome=%+v stdout=%q stderr=%q, want exit 0 %q", outcome, stdout, stderr, tc.stdout)
				}
			} else if outcome.Kind != native.RunExited || outcome.ExitCode != 65 || stdout != "" || stderr != tc.stderr {
				t.Fatalf("outcome=%+v stdout=%q stderr=%q, want exit 65 empty stdout %q", outcome, stdout, stderr, tc.stderr)
			}
			if receipt.Allocations != receipt.Frees || receipt.Doubles != 0 {
				t.Fatalf("receipt=%+v, want allocations==frees and no double free", receipt)
			}
			if tc.check != nil {
				tc.check(t, receipt)
			}
		})
	}
}
