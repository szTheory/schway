package native_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const phase37CRCObserverHooks = "#include <stddef.h>\n#include <sys/stat.h>\n#include <sys/types.h>\n\n" +
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

type phase37CRCReceipt struct {
	Allocations int
	Frees       int
	Doubles     int
	BytesRead   int
	Opens       int
	Closes      int
	ReadCalls   int
	MaxRequest  int
}

type phase37CRCResult struct {
	Outcome native.RunOutcome
	Stdout  string
	Stderr  string
	Receipt phase37CRCReceipt
}

func phase37CRCReplaceOnce(t *testing.T, source, anchor, replacement string) string {
	t.Helper()
	if count := strings.Count(source, anchor); count != 1 {
		t.Fatalf("CRC adapter mutant anchor %q occurs %d times, want exactly one", anchor, count)
	}
	return strings.Replace(source, anchor, replacement, 1)
}

func phase37CRCSnapshot(t *testing.T) map[string][]byte {
	t.Helper()
	paths := []string{
		testsupport.ProjectPath("examples", "crc32", "adapter.c"),
		testsupport.ProjectPath("examples", "crc32", "adapter.h"),
		testsupport.ProjectPath("examples", "crc32", "crc32.schway"),
		testsupport.ProjectPath("examples", "crc32", "file_bytes.bindings.json"),
		testsupport.ProjectPath("examples", "crc32", "lib", "crc32.schway"),
		testsupport.ProjectPath("internal", "compiler", "native", "testdata", "phase31_observer.c"),
	}
	snapshot := make(map[string][]byte, len(paths))
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		snapshot[path] = data
	}
	t.Cleanup(func() {
		for path, original := range snapshot {
			current, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(current, original) {
				t.Errorf("tracked CRC source changed during private control: %s (read error: %v)", path, err)
			}
		}
	})
	return snapshot
}

func phase37CRCCopy(t *testing.T, withObserver bool, mutate func(string) string) (root, artifact, cSource string) {
	t.Helper()
	phase29RequireClang(t)
	_ = phase37CRCSnapshot(t)
	sourceRoot := testsupport.ProjectPath("examples", "crc32")
	root = t.TempDir()
	files := []string{"adapter.c", "adapter.h", "crc32.schway", "crc_low.schway", "file_bytes.bindings.json", "lib/crc32.schway"}
	for _, relative := range files {
		data, err := os.ReadFile(filepath.Join(sourceRoot, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		destination := filepath.Join(root, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(destination, data, 0o600); err != nil {
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
	if withObserver {
		adapter = phase37CRCReplaceOnce(t, adapter, "#include \"adapter.h\"\n", phase37CRCObserverHooks+"#include \"adapter.h\"\n")
		observer, err := os.ReadFile(testsupport.ProjectPath("internal", "compiler", "native", "testdata", "phase31_observer.c"))
		if err != nil {
			t.Fatal(err)
		}
		observed := string(observer)
		observed = phase37CRCReplaceOnce(t, observed, "static unsigned long read_calls;", "static unsigned long read_calls;\nstatic unsigned long opens;\nstatic unsigned long closes;\nstatic unsigned long max_request;")
		observed = phase37CRCReplaceOnce(t, observed,
			"allocations=%lu frees=%lu doubles=%lu bytes_read=%lu\\n\", allocations, frees, doubles, bytes_read)",
			"allocations=%lu frees=%lu doubles=%lu bytes_read=%lu opens=%lu closes=%lu read_calls=%lu max_request=%lu\\n\", allocations, frees, doubles, bytes_read, opens, closes, read_calls, max_request)")
		observed = phase37CRCReplaceOnce(t, observed, "int phase31_observer_open(const char *path, int flags) {\n  ensure();", "int phase31_observer_open(const char *path, int flags) {\n  ensure();\n  opens++;")
		observed = phase37CRCReplaceOnce(t, observed, "  read_calls++;\n  if (first", "  read_calls++;\n  if ((unsigned long)count > max_request) max_request = (unsigned long)count;\n  if (first")
		observed = phase37CRCReplaceOnce(t, observed, "int phase31_observer_close(int descriptor) {\n  int status = close(descriptor);", "int phase31_observer_close(int descriptor) {\n  int status;\n  closes++;\n  status = close(descriptor);")
		observerPath := filepath.Join(root, "phase31_observer.c")
		if err := os.WriteFile(observerPath, []byte(observed), 0o600); err != nil {
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

	captured := ""
	restore := native.SetApplicationCSourceMutationForTest(func(source string) string {
		captured = source
		return source
	})
	defer restore()
	artifact = filepath.Join(t.TempDir(), "crc32")
	entry := filepath.Join(root, "crc32.schway")
	_, diagnostics, err := session.BuildApplicationFile(context.Background(), entry, artifact, native.DefaultRunner(), manifestPath)
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("build copied CRC application: diagnostics=%+v err=%v", diagnostics, err)
	}
	if captured == "" {
		t.Fatal("native build did not expose the generated CRC consumer source")
	}
	return root, artifact, captured
}

func phase37CRCInput(t *testing.T, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func phase37CRCRun(t *testing.T, artifact, path, fault string) phase37CRCResult {
	t.Helper()
	receiptPath := filepath.Join(t.TempDir(), "receipt")
	t.Setenv("SCHWAY_PHASE31_FAULT", fault)
	t.Setenv("SCHWAY_PHASE31_RECEIPT", receiptPath)
	var stdout, stderr bytes.Buffer
	outcome, err := native.RunApplication(context.Background(), artifact, path, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run CRC artifact under %q: %v", fault, err)
	}
	data, err := os.ReadFile(receiptPath)
	if err != nil {
		t.Fatalf("read physical receipt: %v", err)
	}
	var receipt phase37CRCReceipt
	if _, err := fmt.Sscanf(string(data), "allocations=%d frees=%d doubles=%d bytes_read=%d opens=%d closes=%d read_calls=%d max_request=%d", &receipt.Allocations, &receipt.Frees, &receipt.Doubles, &receipt.BytesRead, &receipt.Opens, &receipt.Closes, &receipt.ReadCalls, &receipt.MaxRequest); err != nil {
		t.Fatalf("parse physical receipt %q: %v", data, err)
	}
	return phase37CRCResult{Outcome: outcome, Stdout: stdout.String(), Stderr: stderr.String(), Receipt: receipt}
}

func phase37CRCAssertSuccess(t *testing.T, run phase37CRCResult, data []byte) {
	t.Helper()
	want := fmt.Sprintf("%d\n", crc32.ChecksumIEEE(data))
	if run.Outcome.Kind != native.RunExited || run.Outcome.ExitCode != 0 || run.Stdout != want || run.Stderr != "" {
		t.Fatalf("CRC run outcome=%+v stdout=%q stderr=%q, want exit 0 and %q", run.Outcome, run.Stdout, run.Stderr, want)
	}
}

func phase37CRCAssertTyped(t *testing.T, run phase37CRCResult, alternative string) {
	t.Helper()
	want := "FileError." + alternative + "\n"
	if run.Outcome.Kind != native.RunExited || run.Outcome.ExitCode != 65 || run.Stdout != "" || run.Stderr != want {
		t.Fatalf("CRC error outcome=%+v stdout=%q stderr=%q, want exit 65 and %q", run.Outcome, run.Stdout, run.Stderr, want)
	}
}

func phase37CRCAssertPhysical(t *testing.T, receipt phase37CRCReceipt, allocations, frees, closes, bytesRead int) {
	t.Helper()
	if receipt.Allocations != allocations || receipt.Frees != frees || receipt.Doubles != 0 || receipt.Opens != 1 || receipt.Closes != closes || receipt.BytesRead != bytesRead {
		t.Fatalf("physical receipt=%+v, want alloc=%d free=%d close=%d bytes_read=%d and one open/no double free", receipt, allocations, frees, closes, bytesRead)
	}
}

func phase37CRCAssertReadCalls(t *testing.T, receipt phase37CRCReceipt, want int) {
	t.Helper()
	if receipt.ReadCalls != want {
		t.Fatalf("physical read calls=%d, want %d", receipt.ReadCalls, want)
	}
}

func TestPhase37CRC32PhysicalCleanup(t *testing.T) {
	root, artifact, _ := phase37CRCCopy(t, true, nil)
	_ = root
	small := []byte("123456789")
	for _, row := range []struct {
		name        string
		data        []byte
		fault       string
		alternative string
		wantSuccess bool
		allocations int
		frees       int
		closes      int
		bytesRead   int
		readCalls   int
	}{
		{name: "success-and-consumer-release", data: small, fault: "none", wantSuccess: true, allocations: 1, frees: 1, closes: 1, bytesRead: len(small), readCalls: 2},
		{name: "stat-error-closes-once", data: small, fault: "fstat-fail", alternative: "FileUnavailable", closes: 1},
		{name: "read-error-frees-and-closes", data: small, fault: "read-eio", alternative: "FileUnavailable", allocations: 1, frees: 1, closes: 1, readCalls: 1},
		{name: "success-close-error-frees-buffer", data: small, fault: "close-fail", alternative: "FileUnavailable", allocations: 1, frees: 1, closes: 1, bytesRead: len(small), readCalls: 2},
		{name: "oversize-precedes-close-error", data: bytes.Repeat([]byte{0xff}, 65537), fault: "close-fail", alternative: "FileTooLarge", allocations: 1, frees: 1, closes: 1, bytesRead: 65537, readCalls: 1},
		{name: "eintr-retries-then-consumer-releases", data: small, fault: "read-eintr-once", wantSuccess: true, allocations: 1, frees: 1, closes: 1, bytesRead: len(small), readCalls: 3},
	} {
		t.Run(row.name, func(t *testing.T) {
			path := phase37CRCInput(t, "input.bin", row.data)
			run := phase37CRCRun(t, artifact, path, row.fault)
			if row.wantSuccess {
				phase37CRCAssertSuccess(t, run, row.data)
			} else {
				phase37CRCAssertTyped(t, run, row.alternative)
			}
			phase37CRCAssertPhysical(t, run.Receipt, row.allocations, row.frees, row.closes, row.bytesRead)
			phase37CRCAssertReadCalls(t, run.Receipt, row.readCalls)
		})
	}
}

func TestPhase37CRC32BoundControls(t *testing.T) {
	_, faithful, _ := phase37CRCCopy(t, true, nil)
	capData := bytes.Repeat([]byte{0xff}, 65536)
	overData := bytes.Repeat([]byte{0xff}, 65537)
	capPath := phase37CRCInput(t, "cap.bin", capData)
	overPath := phase37CRCInput(t, "oversize.bin", overData)
	t.Run("faithful-cap-and-cap-plus-one-physical-reads", func(t *testing.T) {
		capRun := phase37CRCRun(t, faithful, capPath, "none")
		phase37CRCAssertSuccess(t, capRun, capData)
		phase37CRCAssertPhysical(t, capRun.Receipt, 1, 1, 1, len(capData))
		phase37CRCAssertReadCalls(t, capRun.Receipt, 2)
		if capRun.Receipt.MaxRequest != 65537 {
			t.Fatalf("cap max request=%d, want 65537", capRun.Receipt.MaxRequest)
		}
		overRun := phase37CRCRun(t, faithful, overPath, "none")
		phase37CRCAssertTyped(t, overRun, "FileTooLarge")
		phase37CRCAssertPhysical(t, overRun.Receipt, 1, 1, 1, 65537)
		phase37CRCAssertReadCalls(t, overRun.Receipt, 1)
		if overRun.Receipt.MaxRequest != 65537 {
			t.Fatalf("cap+1 max request=%d, want 65537", overRun.Receipt.MaxRequest)
		}
	})
	t.Run("missing-free-mutant-is-killed-by-physical-receipt", func(t *testing.T) {
		_, mutant, _ := phase37CRCCopy(t, true, func(adapter string) string {
			return phase37CRCReplaceOnce(t, adapter, "if (data != NULL) SCHWAY_FILE_BYTES_FREE(data);", "if (data != NULL) (void)data;")
		})
		run := phase37CRCRun(t, mutant, overPath, "none")
		phase37CRCAssertTyped(t, run, "FileTooLarge")
		if run.Receipt.Allocations != 1 || run.Receipt.Frees != 0 || run.Receipt.Closes != 1 || run.Receipt.Doubles != 0 || run.Receipt.ReadCalls != 1 {
			t.Fatalf("missing-free mutant receipt=%+v, want the physical 1-allocation/0-free mismatch", run.Receipt)
		}
		t.Logf("killed missing-free mutant after successful build and FileTooLarge execution: %+v", run.Receipt)
	})
	t.Run("one-byte-short-read-mutant-is-killed-at-cap-plus-one", func(t *testing.T) {
		_, mutant, _ := phase37CRCCopy(t, true, func(adapter string) string {
			return phase37CRCReplaceOnce(t, adapter,
				"(size_t)SCHWAY_FILE_BYTES_LIMIT + 1u - received);",
				"(size_t)SCHWAY_FILE_BYTES_LIMIT - received);")
		})
		run := phase37CRCRun(t, mutant, overPath, "none")
		if run.Outcome.Kind != native.RunExited || run.Outcome.ExitCode != 0 || run.Receipt.BytesRead != 65536 || run.Receipt.Closes != 1 || run.Receipt.ReadCalls != 2 || run.Receipt.MaxRequest != 65536 {
			t.Fatalf("short-read mutant outcome=%+v stdout=%q stderr=%q receipt=%+v, want successful truncation after exactly 65536 physical bytes", run.Outcome, run.Stdout, run.Stderr, run.Receipt)
		}
		if run.Stderr == "FileError.FileTooLarge\n" {
			t.Fatalf("short-read mutant survived cap+1 error assertion: %+v", run)
		}
		t.Logf("killed one-byte-short-read mutant after successful build by observing %d/65537 bytes", run.Receipt.BytesRead)
	})
}

func phase37CRCHeaderInclude(source, header string) string {
	quoted := `#include "` + filepath.ToSlash(header) + `"`
	source = strings.ReplaceAll(source, `#include "local/adapter.h"`, quoted)
	return strings.ReplaceAll(source, `#include "adapter.h"`, quoted)
}

func TestPhase37CRC32Sanitizer(t *testing.T) {
	available := false
	t.Run("host-capability", func(t *testing.T) {
		if _, err := native.DefaultRunner().ProbeSanitizerRuntime(context.Background()); err != nil {
			t.Skipf("env:focused ASan+UBSan unavailable on this host: %v", err)
		}
		available = true
	})
	if !available {
		t.Run("sanitized-crc-consumer-and-adapter-errors", func(t *testing.T) {
			t.Skip("env:ASan+UBSan host-capability subtest did not pass")
		})
		return
	}

	root, _, generated := phase37CRCCopy(t, false, nil)
	adapter := filepath.Join(root, "adapter.c")
	header := filepath.Join(root, "adapter.h")
	sanitizedSource := phase37CRCHeaderInclude(generated, header)
	small := phase37CRCInput(t, "small.bin", []byte("123456789"))
	capPath := phase37CRCInput(t, "cap.bin", bytes.Repeat([]byte{0xff}, 65536))
	t.Run("generated-crc-consumer-small-and-cap-success", func(t *testing.T) {
		runner := native.DefaultRunner()
		runner.ForeignSources = []string{adapter}
		report, err := runner.RunSanitized(context.Background(), sanitizedSource, []string{small, capPath})
		if err != nil || report.CheckKind != "" || report.ExitCode != 0 {
			t.Fatalf("sanitized generated CRC consumer report=%+v err=%v, want clean success", report, err)
		}
	})

	// The generated application intentionally exits 65 for a typed FileError,
	// which RunSanitized classifies as an abort. This small C caller makes that
	// actual adapter status branch return normally so ASan+UBSan can inspect its
	// cap+1 cleanup without changing the generated consumer's error contract.
	harness := fmt.Sprintf(`#include %q
#include <string.h>
int main(int argc, char **argv) {
  schway_file_bytes_acquire_result result;
  int expected;
  if (argc != 2) return 20;
  expected = strstr(argv[1], "oversize") != NULL ? SCHWAY_FILE_BYTES_TOO_LARGE : SCHWAY_FILE_BYTES_OK;
  result = schway_file_bytes_acquire(argv[1]);
  if (result.status != expected) return 21;
  if (result.status == SCHWAY_FILE_BYTES_OK) schway_file_bytes_release(result.owner);
  return 0;
}
`, filepath.ToSlash(header))
	overPath := phase37CRCInput(t, "oversize.bin", bytes.Repeat([]byte{0xff}, 65537))
	t.Run("actual-adapter-cap-plus-one-error-cleanup", func(t *testing.T) {
		runner := native.DefaultRunner()
		runner.ForeignSources = []string{adapter}
		report, err := runner.RunSanitized(context.Background(), harness, []string{small, overPath})
		if err != nil || report.CheckKind != "" || report.ExitCode != 0 {
			t.Fatalf("sanitized adapter boundary report=%+v err=%v, want clean success including expected FileTooLarge status", report, err)
		}
	})
}
