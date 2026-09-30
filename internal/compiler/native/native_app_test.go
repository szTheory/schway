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

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase23GeneratedReleaseFollowsBorrowedUse(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("examples", "phase23", "file_byte.schway"))
	if err != nil {
		t.Fatal(err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("parse diagnostics: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("check diagnostics: %+v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("core validation problems: %+v", validated.Problems)
	}
	cSource, err := cgen.EmitApplication(validated.Program())
	if err != nil {
		t.Fatal(err)
	}
	useAt := strings.Index(cSource, "schway_file_byte_use(")
	useErrorDiagnosticAt := strings.Index(cSource, `fputs("schway_file_byte_use: UnsupportedByte\n", stderr);`)
	var releasePositions []int
	for offset := 0; offset < len(cSource); {
		next := strings.Index(cSource[offset:], "schway_file_byte_release(")
		if next < 0 {
			break
		}
		releasePositions = append(releasePositions, offset+next)
		offset += next + len("schway_file_byte_release(")
	}
	if useAt < 0 || len(releasePositions) != 2 || releasePositions[1] <= useAt || useErrorDiagnosticAt <= releasePositions[1] {
		t.Fatalf("generated app does not call borrowed use then the successful-path destructor: use=%d release=%v", useAt, releasePositions)
	}
	useArgStart := useAt + len("schway_file_byte_use(")
	useArgEnd := strings.Index(cSource[useArgStart:], ");")
	releaseArgStart := releasePositions[1] + len("schway_file_byte_release(")
	releaseArgEnd := strings.Index(cSource[releaseArgStart:], ");")
	if useArgEnd < 0 || releaseArgEnd < 0 || strings.TrimSpace(cSource[useArgStart:useArgStart+useArgEnd]) != strings.TrimSpace(cSource[releaseArgStart:releaseArgStart+releaseArgEnd]) || !strings.Contains(cSource[releasePositions[1]:], `status != 0) { fputs("schway_file_byte_use: UnsupportedByte\n", stderr); exit(65); }`) {
		t.Fatal("generated cleanup does not retain the acquired owner through use and release it before a use failure exits")
	}
}

func TestPhase23OperationABI(t *testing.T) {
	headerBytes, err := os.ReadFile(testsupport.ProjectPath("examples", "phase23", "adapter.h"))
	if err != nil {
		t.Fatal(err)
	}
	header := string(headerBytes)
	manifest := BindingManifest{
		Schema: BindingSchema, Headers: []string{"adapter.h"}, IncludeDirs: []string{"."},
		Symbols: []BindingSymbol{
			{Name: "schway_file_byte_acquire", Header: "adapter.h", FunctionType: "schway_file_byte_acquire_fn"},
			{Name: "schway_file_byte_use", Header: "adapter.h", FunctionType: "schway_file_byte_use_fn"},
			{Name: "schway_file_byte_release", Header: "adapter.h", FunctionType: "schway_file_byte_release_fn"},
		},
	}
	runner := DefaultRunner()
	clang, err := exec.LookPath("clang")
	if err != nil {
		t.Skip("Clang is required for operation ABI checks; see probe:TestPhase23PublicFileByte")
	}

	compile := func(t *testing.T, content string) error {
		t.Helper()
		directory, err := filepath.EvalSymlinks(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		resolved := ResolvedBindings{
			Manifest: manifest,
			files:    map[string][]byte{"adapter.h": []byte(content)},
		}
		_, _, _, _, compileErr := runner.compileBindings(context.Background(), clang, directory, resolved)
		return compileErr
	}
	if err := compile(t, header); err != nil {
		t.Fatalf("baseline operation header failed to compile: %v", err)
	}

	for _, test := range []struct {
		name string
		from string
		to   string
	}{
		{"acquire", "schway_file_byte_acquire_result schway_file_byte_acquire(const char *path);", "schway_file_byte_acquire_result schway_file_byte_acquire(uint64_t path);"},
		{"use", "schway_file_byte_use_result schway_file_byte_use(schway_file_byte_owner owner);", "schway_file_byte_use_result schway_file_byte_use(const schway_file_byte_owner *owner);"},
		{"release", "void schway_file_byte_release(schway_file_byte_owner owner);", "void schway_file_byte_release(const schway_file_byte_owner *owner);"},
	} {
		t.Run(test.name, func(t *testing.T) {
			mutated := strings.Replace(header, test.from, test.to, 1)
			if mutated == header {
				t.Fatal("prototype mutation did not match the fixture")
			}
			if err := compile(t, mutated); err == nil {
				t.Fatal("mismatched operation prototype passed its independent Clang probe")
			}
		})
	}

	t.Run("target layout", func(t *testing.T) {
		mutated := strings.Replace(header, "  unsigned char *data;\n  uint64_t length;", "  uint64_t length;\n  unsigned char *data;", 1)
		if mutated == header {
			t.Fatal("layout mutation did not match the fixture")
		}
		if err := compile(t, mutated); err == nil {
			t.Fatal("target record layout mutation passed the C17 probes")
		}
	})
}

func TestPhase23AcquireFailuresInitializeAndFreePartialAllocations(t *testing.T) {
	clang, err := exec.LookPath("clang")
	if err != nil {
		t.Fatal(err)
	}
	root := testsupport.ProjectPath("examples", "phase23")
	directory := t.TempDir()
	harnessPath := filepath.Join(directory, "acquire_harness.c")
	binaryPath := filepath.Join(directory, "acquire_harness")
	const harness = `#define _POSIX_C_SOURCE 200809L
#include "adapter.h"

#include <errno.h>
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <sys/types.h>

enum {
  SC_OPEN_FAILED = 1,
  SC_NON_REGULAR = 2,
  SC_FSTAT_FAILED = 3,
  SC_ALLOC_FAILED = 4,
  SC_EMPTY = 5,
  SC_TOO_LONG = 6,
  SC_READ_FAILED = 7,
  SC_READ_PROBE_FAILED = 8,
  SC_CLOSE_FAILED = 9,
  SC_RETRY_EINTR = 10,
  SC_SUCCESS = 11,
  SC_EMPTY_CLOSE_FAILED = 12
};

static int scenario;
static int open_calls;
static int fstat_calls;
static int allocation_calls;
static int free_calls;
static int read_calls;
static int close_calls;
static const char *expected_path;
static int path_unchanged;

int phase23_test_open(const char *path, int flags) {
  open_calls++;
  path_unchanged = strcmp(path, expected_path) == 0;
  if (flags != (O_RDONLY | O_NONBLOCK)) return -1;
  if (scenario == SC_OPEN_FAILED) {
    errno = ENOENT;
    return -1;
  }
  return 77;
}

int phase23_test_fstat(int descriptor, struct stat *details) {
  fstat_calls++;
  if (descriptor != 77) return -1;
  if (scenario == SC_FSTAT_FAILED) {
    errno = EIO;
    return -1;
  }
  memset(details, 0, sizeof(*details));
  details->st_mode = scenario == SC_NON_REGULAR ? S_IFIFO : S_IFREG;
  return 0;
}

void *phase23_test_malloc(size_t size) {
  allocation_calls++;
  if (size != 1u || scenario == SC_ALLOC_FAILED) return NULL;
  return malloc(size);
}

void phase23_test_free(void *pointer) {
  if (pointer != NULL) free_calls++;
  free(pointer);
}

ssize_t phase23_test_read(int descriptor, void *buffer, size_t count) {
  int call = read_calls++;
  if (descriptor != 77 || count != 1u) return -1;
  if (scenario == SC_EMPTY || scenario == SC_EMPTY_CLOSE_FAILED) return 0;
  if (scenario == SC_READ_FAILED && call == 0) {
    errno = EIO;
    return -1;
  }
  if (scenario == SC_READ_PROBE_FAILED && call == 1) {
    errno = EIO;
    return -1;
  }
  if (scenario == SC_RETRY_EINTR && (call == 0 || call == 2)) {
    errno = EINTR;
    return -1;
  }
  if (call == 0 || (scenario == SC_RETRY_EINTR && call == 1)) {
    *(unsigned char *)buffer = 0x41u;
    return 1;
  }
  if (scenario == SC_TOO_LONG && call == 1) {
    *(unsigned char *)buffer = 0x42u;
    return 1;
  }
  return 0;
}

int phase23_test_close(int descriptor) {
  close_calls++;
  if (descriptor != 77) return -1;
  if (scenario == SC_CLOSE_FAILED || scenario == SC_EMPTY_CLOSE_FAILED) {
    errno = EIO;
    return -1;
  }
  return 0;
}

struct test_case {
  const char *name;
  const char *path;
  int scenario;
  int status;
  int opens;
  int fstats;
  int allocations;
  int frees;
  int reads;
  int closes;
};

static int run_case(const struct test_case *test) {
  schway_file_byte_acquire_result result;
  scenario = test->scenario;
  open_calls = fstat_calls = allocation_calls = free_calls = 0;
  read_calls = close_calls = 0;
  expected_path = test->path;
  path_unchanged = 0;
  result = schway_file_byte_acquire(test->path);
  if (result.status != test->status || open_calls != test->opens ||
      fstat_calls != test->fstats || allocation_calls != test->allocations ||
      free_calls != test->frees || read_calls != test->reads ||
      close_calls != test->closes || (test->opens != 0 && !path_unchanged)) {
    fprintf(stderr,
            "%s: status=%d owner=%p length=%llu open=%d fstat=%d alloc=%d free=%d read=%d close=%d path_same=%d\n",
            test->name, result.status, (void *)result.owner.data,
            (unsigned long long)result.owner.length, open_calls, fstat_calls,
            allocation_calls, free_calls, read_calls, close_calls, path_unchanged);
    return 1;
  }
  if (test->status == 0) {
    if (result.owner.data == NULL || result.owner.length != 1u || result.owner.data[0] != 0x41u) {
      fprintf(stderr, "%s: success owner not initialized with byte 0x41\n", test->name);
      return 1;
    }
    schway_file_byte_release(result.owner);
    if (free_calls != test->frees + 1) {
      fprintf(stderr, "%s: release did not free the successful allocation exactly once\n", test->name);
      return 1;
    }
    return 0;
  }
  if (result.owner.data != NULL || result.owner.length != 0u) {
    fprintf(stderr, "%s: failure published an owner\n", test->name);
    return 1;
  }
  return 0;
}

static int run_unchanged_path_case(const char *path) {
  char expected[4097];
  size_t path_length = strlen(path);
  schway_file_byte_acquire_result result;
  if (path_length > sizeof expected - 1u) return 2;
  memcpy(expected, path, path_length + 1u);
  scenario = SC_OPEN_FAILED;
  open_calls = fstat_calls = allocation_calls = free_calls = 0;
  read_calls = close_calls = 0;
  expected_path = expected;
  path_unchanged = 0;
  result = schway_file_byte_acquire(path);
  if (result.status != 3 || result.owner.data != NULL || result.owner.length != 0u ||
      open_calls != 1 || fstat_calls != 0 || allocation_calls != 0 ||
      free_calls != 0 || read_calls != 0 || close_calls != 0 || !path_unchanged) {
    fprintf(stderr, "path length=%llu was changed or did not reach the open hook\n",
            (unsigned long long)path_length);
    return 1;
  }
  return 0;
}

int main(int argc, char **argv) {
  static const struct test_case cases[] = {
    {"open", "open-failed", SC_OPEN_FAILED, 3, 1, 0, 0, 0, 0, 0},
    {"non-regular", "non-regular", SC_NON_REGULAR, 4, 1, 1, 0, 0, 0, 1},
    {"fstat", "stat-failed", SC_FSTAT_FAILED, 4, 1, 1, 0, 0, 0, 1},
    {"allocation", "allocation-failed", SC_ALLOC_FAILED, 6, 1, 1, 1, 0, 0, 1},
    {"empty", "empty", SC_EMPTY, 1, 1, 1, 1, 1, 1, 1},
    {"two-byte", "two-byte", SC_TOO_LONG, 2, 1, 1, 1, 1, 2, 1},
    {"read", "read-failed", SC_READ_FAILED, 5, 1, 1, 1, 1, 1, 1},
    {"read-probe", "read-probe-failed", SC_READ_PROBE_FAILED, 5, 1, 1, 1, 1, 2, 1},
    {"close", "close-failed", SC_CLOSE_FAILED, 7, 1, 1, 1, 1, 2, 1},
    {"empty-close", "empty-close-failed", SC_EMPTY_CLOSE_FAILED, 7, 1, 1, 1, 1, 1, 1},
    {"eintr-retry", "eintr-retry", SC_RETRY_EINTR, 0, 1, 1, 1, 0, 4, 1},
    {"success", "success", SC_SUCCESS, 0, 1, 1, 1, 0, 2, 1}
  };
  size_t index;
  if (argc == 2) return run_unchanged_path_case(argv[1]);
  if (argc != 1) return 2;
  for (index = 0u; index < sizeof cases / sizeof cases[0]; index++) {
    if (run_case(&cases[index]) != 0) return 1;
  }
  return 0;
}

#include <adapter.c>
`
	if err := os.WriteFile(harnessPath, []byte(harness), 0o600); err != nil {
		t.Fatal(err)
	}
	compileCtx, cancelCompile := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelCompile()
	compile := exec.CommandContext(compileCtx, clang,
		"-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic",
		"-DSCHWAY_FILE_BYTE_OPEN=phase23_test_open",
		"-DSCHWAY_FILE_BYTE_FSTAT=phase23_test_fstat",
		"-DSCHWAY_FILE_BYTE_MALLOC=phase23_test_malloc",
		"-DSCHWAY_FILE_BYTE_FREE=phase23_test_free",
		"-DSCHWAY_FILE_BYTE_READ=phase23_test_read",
		"-DSCHWAY_FILE_BYTE_CLOSE=phase23_test_close",
		"-I", root, harnessPath, "-o", binaryPath)
	var compileStdout, compileStderr boundedWriter
	compile.Stdout, compile.Stderr = &compileStdout, &compileStderr
	if err := compile.Run(); err != nil || compileStdout.overflowed() || compileStderr.overflowed() {
		t.Fatalf("compile acquisition fault harness: %v stdout=%q stderr=%q", err, compileStdout.bytes(), compileStderr.bytes())
	}
	runCtx, cancelRun := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelRun()
	run := exec.CommandContext(runCtx, binaryPath)
	var runStdout, runStderr boundedWriter
	run.Stdout, run.Stderr = &runStdout, &runStderr
	if err := run.Run(); err != nil || runStdout.overflowed() || runStderr.overflowed() {
		t.Fatalf("acquisition fault harness: %v stdout=%q stderr=%q", err, runStdout.bytes(), runStderr.bytes())
	}
	// Permission bits may not produce an open failure for privileged test users;
	// the table above injects OpenFailed directly at this adapter boundary.
	for _, path := range []string{"/", strings.Repeat("p", 4096)} {
		pathRunCtx, cancelPathRun := context.WithTimeout(context.Background(), 10*time.Second)
		pathRun := exec.CommandContext(pathRunCtx, binaryPath, path)
		var pathStdout, pathStderr boundedWriter
		pathRun.Stdout, pathRun.Stderr = &pathStdout, &pathStderr
		pathErr := pathRun.Run()
		cancelPathRun()
		if pathErr != nil || pathStdout.overflowed() || pathStderr.overflowed() || len(pathStdout.bytes()) != 0 || string(pathStderr.bytes()) != "schway_file_byte_acquire: OpenFailed\n" {
			t.Fatalf("unchanged path length=%d: err=%v stdout=%q stderr=%q", len(path), pathErr, pathStdout.bytes(), pathStderr.bytes())
		}
	}
}

func TestPhase22EvidenceDisabledCompleteAndStreamIsolation(t *testing.T) {
	t.Run("disabled", func(t *testing.T) {
		root := t.TempDir()
		marker := filepath.Join(root, "launches.txt")
		artifact := writePhase22EvidenceScript(t, fmt.Sprintf("printf 'launch\\n' >> %s\nprintf 'app stdout\\n'\nprintf 'app stderr\\n' >&2\n", shellQuote(marker)))
		reportPath := filepath.Join(root, "disabled.json")
		var stdout, stderr bytes.Buffer
		outcome, report, err := DefaultRunner().RunApplicationWithEvidence(context.Background(), artifact, "7", reportPath, EvidenceDisabled, &stdout, &stderr)
		if err != nil || outcome.Kind != RunExited || outcome.ExitCode != 0 {
			t.Fatalf("outcome=%+v err=%v", outcome, err)
		}
		if report.Schema != ApplicationEvidenceSchema || report.CaptureStatus != EvidenceStatusDisabled || report.Verified || report.Execution != nil {
			t.Fatalf("disabled report=%+v", report)
		}
		if stdout.String() != "app stdout\n" || stderr.String() != "app stderr\n" {
			t.Fatalf("streams stdout=%q stderr=%q", stdout.String(), stderr.String())
		}
		assertPhase22LaunchCount(t, marker, 1)
		assertPhase22ReportFile(t, reportPath, report)
	})

	t.Run("complete capture keeps app bytes unchanged and launches once", func(t *testing.T) {
		root := t.TempDir()
		marker := filepath.Join(root, "launches.txt")
		capture := phase22CompleteCapture(t)
		artifact := writePhase22EvidenceScript(t, fmt.Sprintf("printf 'launch\\n' >> %s\nprintf 'app stdout {ordinary}\\n'\nprintf 'app stderr\\n' >&2\nprintf %%s %s > \"$SCHWAY_APP_EVIDENCE_PATH\"\n", shellQuote(marker), shellQuote(capture)))
		reportPath := filepath.Join(root, "complete.json")
		var stdout, stderr bytes.Buffer
		outcome, report, err := DefaultRunner().RunApplicationWithEvidence(context.Background(), artifact, "7", reportPath, EvidenceEvents, &stdout, &stderr)
		if err != nil || outcome.Kind != RunExited || outcome.ExitCode != 0 {
			t.Fatalf("outcome=%+v err=%v", outcome, err)
		}
		if report.CaptureStatus != EvidenceStatusComplete || report.Verified || report.Execution == nil || report.BuildID == "" || report.InputID == "" || report.InputDigest != digestBytes([]byte("7")) {
			t.Fatalf("complete report=%+v", report)
		}
		if stdout.String() != "app stdout {ordinary}\n" || stderr.String() != "app stderr\n" {
			t.Fatalf("capture polluted app streams: stdout=%q stderr=%q", stdout.String(), stderr.String())
		}
		assertPhase22LaunchCount(t, marker, 1)
		assertPhase22ReportFile(t, reportPath, report)
	})
}

func TestPhase22OrdinaryRunIgnoresAmbientEvidencePath(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "must-survive.txt")
	if err := os.WriteFile(target, []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SCHWAY_APP_EVIDENCE_PATH", target)
	artifact := writePhase22EvidenceScript(t, "if [ -n \"${SCHWAY_APP_EVIDENCE_PATH:-}\" ]; then printf 'overwritten' > \"$SCHWAY_APP_EVIDENCE_PATH\"; fi\nprintf 'app\\n'\n")
	var stdout, stderr bytes.Buffer
	outcome, err := DefaultRunner().RunApplication(context.Background(), artifact, "7", &stdout, &stderr)
	if err != nil || outcome.Kind != RunExited || outcome.ExitCode != 0 {
		t.Fatalf("outcome=%+v err=%v", outcome, err)
	}
	data, err := os.ReadFile(target)
	if err != nil || string(data) != "original" {
		t.Fatalf("ambient evidence target=%q err=%v; want unchanged original", data, err)
	}
}

func TestPhase22EvidenceMissingPartialAndCapacityControls(t *testing.T) {
	t.Run("missing capture", func(t *testing.T) {
		phase22AssertIncompleteEvidence(t, "printf 'app\\n'\n", nil)
	})
	t.Run("partial capture", func(t *testing.T) {
		phase22AssertIncompleteEvidence(t, "printf 'app\\n'\nprintf '{\\\"schema\\\":' > \"$SCHWAY_APP_EVIDENCE_PATH\"\n", nil)
	})
	t.Run("capacity exhausted", func(t *testing.T) {
		capture := phase22CompleteCapture(t)
		body := fmt.Sprintf("printf 'app\\n'\nprintf %%s %s > \"$SCHWAY_APP_EVIDENCE_PATH\"\n", shellQuote(capture))
		phase22AssertIncompleteEvidence(t, body, func(runner *Runner) { runner.evidenceLimit = 8 })
	})
}

func TestPhase22EvidenceReportWriteFailurePreservesAppOutcome(t *testing.T) {
	root := t.TempDir()
	marker := filepath.Join(root, "launches.txt")
	artifact := writePhase22EvidenceScript(t, fmt.Sprintf("printf 'launch\\n' >> %s\nprintf 'app stdout\\n'\nprintf 'app stderr\\n' >&2\n", shellQuote(marker)))
	var stdout, stderr bytes.Buffer
	outcome, report, err := DefaultRunner().RunApplicationWithEvidence(context.Background(), artifact, "7", filepath.Join(root, "missing", "report.json"), EvidenceDisabled, &stdout, &stderr)
	var toolError *ToolError
	if !errors.As(err, &toolError) || toolError.Code != "native.evidence_report_write_failed" {
		t.Fatalf("report error=%v, want native.evidence_report_write_failed", err)
	}
	if outcome.Kind != RunExited || outcome.ExitCode != 0 || report.ProcessOutcome != outcome || report.CaptureStatus != EvidenceStatusDisabled {
		t.Fatalf("outcome=%+v report=%+v", outcome, report)
	}
	if stdout.String() != "app stdout\n" || stderr.String() != "app stderr\n" {
		t.Fatalf("streams stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
	assertPhase22LaunchCount(t, marker, 1)
}

func TestPhase22CompleteEvidenceReportWriteFailureIsNotVerified(t *testing.T) {
	root := t.TempDir()
	marker := filepath.Join(root, "launches.txt")
	capture := phase22CompleteCapture(t)
	body := fmt.Sprintf("printf 'launch\\n' >> %s\nprintf 'app stdout\\n'\nprintf 'app stderr\\n' >&2\nprintf %%s %s > \"$SCHWAY_APP_EVIDENCE_PATH\"\n", shellQuote(marker), shellQuote(capture))
	artifact := writePhase22EvidenceScript(t, body)
	var stdout, stderr bytes.Buffer
	outcome, report, err := DefaultRunner().RunApplicationWithEvidence(context.Background(), artifact, "7", filepath.Join(root, "missing", "report.json"), EvidenceEvents, &stdout, &stderr)
	var toolError *ToolError
	if !errors.As(err, &toolError) || toolError.Code != "native.evidence_report_write_failed" {
		t.Fatalf("report error=%v, want native.evidence_report_write_failed", err)
	}
	if outcome.Kind != RunExited || outcome.ExitCode != 0 || report.ProcessOutcome != outcome || report.CaptureStatus != EvidenceStatusComplete || report.Execution == nil || report.Verified {
		t.Fatalf("outcome=%+v report=%+v; complete capture must remain explicitly unverified when report publication fails", outcome, report)
	}
	if stdout.String() != "app stdout\n" || stderr.String() != "app stderr\n" {
		t.Fatalf("streams stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
	assertPhase22LaunchCount(t, marker, 1)
}

func phase22AssertIncompleteEvidence(t *testing.T, body string, configure func(*Runner)) {
	t.Helper()
	root := t.TempDir()
	marker := filepath.Join(root, "launches.txt")
	artifact := writePhase22EvidenceScript(t, fmt.Sprintf("printf 'launch\\n' >> %s\n%s", shellQuote(marker), body))
	reportPath := filepath.Join(root, "report.json")
	runner := DefaultRunner()
	if configure != nil {
		configure(&runner)
	}
	var stdout, stderr bytes.Buffer
	outcome, report, err := runner.RunApplicationWithEvidence(context.Background(), artifact, "7", reportPath, EvidenceEvents, &stdout, &stderr)
	var toolError *ToolError
	if !errors.As(err, &toolError) {
		t.Fatalf("RunApplicationWithEvidence error=%v, want a tool error", err)
	}
	wantStatus, wantCode := EvidenceStatusIncomplete, "native.evidence_incomplete"
	if runner.evidenceLimit > 0 {
		wantStatus, wantCode = EvidenceStatusCapacityExhausted, "native.evidence_capacity_exhausted"
	}
	if toolError.Code != wantCode || report.CaptureStatus != wantStatus || report.Verified || report.Execution != nil {
		t.Fatalf("error=%v report=%+v, want %s/%s", err, report, wantCode, wantStatus)
	}
	if outcome.Kind != RunExited || outcome.ExitCode != 0 || report.ProcessOutcome != outcome {
		t.Fatalf("child outcome=%+v report=%+v", outcome, report)
	}
	if stdout.String() != "app\n" || stderr.Len() != 0 {
		t.Fatalf("streams stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
	assertPhase22LaunchCount(t, marker, 1)
	assertPhase22ReportFile(t, reportPath, report)
}

func phase22CompleteCapture(t *testing.T) string {
	t.Helper()
	document := execution.Execution{
		Schema:  execution.Schema2,
		Outcome: execution.Outcome{Kind: execution.OutcomeReturned, Value: "7"},
		Events: []execution.Event{{
			Schema: execution.Schema2, ID: "return:event:returned", Kind: "function.returned",
			FunctionID: "identity", SourcePlace: "place:input", TypeID: "U64", Invocation: "inv:entry:identity",
		}},
		LiveResources: []string{},
	}
	encodedExecution, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	encodedCapture, err := json.Marshal(applicationCapture{Schema: applicationCaptureSchema, Status: string(EvidenceStatusComplete), Execution: encodedExecution})
	if err != nil {
		t.Fatal(err)
	}
	return string(encodedCapture)
}

func writePhase22EvidenceScript(t *testing.T, body string) string {
	t.Helper()
	artifact := writePhase22Script(t, body)
	receiptData, err := os.ReadFile(applicationReceiptPath(artifact))
	if err != nil {
		t.Fatal(err)
	}
	var receipt BuildReceipt
	if err := json.Unmarshal(receiptData, &receipt); err != nil {
		t.Fatal(err)
	}
	identity := BuildIdentityInputs{
		SourceDigest: receipt.SourceDigest, EmittedCDigest: receipt.EmittedCDigest,
		CompilerDigest: digestBytes([]byte("test compiler")), CompilerVersion: receipt.CompilerVersion,
		Target: receipt.Target, HostABI: runtime.GOOS + "/" + runtime.GOARCH,
		Flags: receipt.Flags, RuntimeDependencies: receipt.RuntimeDependencies,
		ExecutableDigest: receipt.ExecutableDigest,
	}
	receipt.Identity = &identity
	receipt.BuildID, receipt.InputID = identity.ID(), identity.InputID()
	encoded, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(applicationReceiptPath(artifact), encoded, 0o600); err != nil {
		t.Fatal(err)
	}
	return artifact
}

func assertPhase22LaunchCount(t *testing.T, path string, expected int) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("launch marker: %v", err)
	}
	if count := len(strings.Fields(string(data))); count != expected {
		t.Fatalf("launch count=%d marker=%q, want %d", count, data, expected)
	}
}

func assertPhase22ReportFile(t *testing.T, path string, expected ApplicationEvidenceReport) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	var actual ApplicationEvidenceReport
	if err := json.Unmarshal(data, &actual); err != nil {
		t.Fatalf("decode report %q: %v", data, err)
	}
	if actual.Schema != expected.Schema || actual.BuildID != expected.BuildID || actual.InputID != expected.InputID || actual.InputDigest != expected.InputDigest || actual.CaptureStatus != expected.CaptureStatus || actual.Verified {
		t.Fatalf("published report=%+v, returned=%+v", actual, expected)
	}
}

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

func TestPhase22ReceiptPublicationFailureRestoresPreviousPair(t *testing.T) {
	root := t.TempDir()
	stage := filepath.Join(root, "stage")
	if err := os.Mkdir(stage, 0o700); err != nil {
		t.Fatal(err)
	}
	artifact := filepath.Join(root, "program")
	receipt := applicationReceiptPath(artifact)
	stagedArtifact := filepath.Join(stage, "new-program")
	stagedReceipt := filepath.Join(stage, "new-receipt")
	for path, data := range map[string]string{
		artifact: "old artifact", receipt: "old receipt",
		stagedArtifact: "new artifact", stagedReceipt: "new receipt",
	} {
		if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	call := 0
	rename := func(oldPath, newPath string) error {
		call++
		if call == 4 { // old pair backups, artifact publish, then receipt publish
			return errors.New("injected receipt rename failure")
		}
		return os.Rename(oldPath, newPath)
	}
	err := publishApplicationPair(stagedArtifact, stagedReceipt, artifact, stage, rename)
	var toolError *ToolError
	if !errors.As(err, &toolError) || toolError.Code != "native.receipt_publish_failed" {
		t.Fatalf("publication error=%v, want native.receipt_publish_failed", err)
	}
	for path, want := range map[string]string{artifact: "old artifact", receipt: "old receipt"} {
		got, readErr := os.ReadFile(path)
		if readErr != nil || string(got) != want {
			t.Fatalf("restored %s=%q err=%v; want %q", path, got, readErr, want)
		}
	}
}

func TestPhase22BuildPreservesBackupWhenRollbackRenameFails(t *testing.T) {
	root := t.TempDir()
	artifact := filepath.Join(root, "program")
	priorReceipt := applicationReceiptPath(artifact)
	if err := os.WriteFile(artifact, []byte("prior executable bytes"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(priorReceipt, []byte("prior receipt bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	call := 0
	runner := DefaultRunner()
	runner.publishRename = func(oldPath, newPath string) error {
		call++
		switch call {
		case 4:
			return errors.New("injected receipt publication failure")
		case 5:
			return errors.New("injected prior artifact restore failure")
		default:
			return os.Rename(oldPath, newPath)
		}
	}
	_, err := runner.BuildApplication(context.Background(), []byte("source"), "int main(void) { return 0; }\n", artifact)
	var recoveryError *applicationPublicationRecoveryError
	if !errors.As(err, &recoveryError) {
		t.Fatalf("BuildApplication error=%v, want recovery error", err)
	}
	if !strings.Contains(err.Error(), recoveryError.directory) {
		t.Fatalf("error %q does not identify recovery directory %q", err, recoveryError.directory)
	}
	info, statErr := os.Stat(recoveryError.directory)
	if statErr != nil || info.Mode().Perm() != 0o700 {
		t.Fatalf("recovery directory info=%v err=%v; want private 0700 directory", info, statErr)
	}
	recoverableArtifact := filepath.Join(recoveryError.directory, "previous-program")
	data, readErr := os.ReadFile(recoverableArtifact)
	if readErr != nil || string(data) != "prior executable bytes" {
		t.Fatalf("recoverable prior artifact=%q err=%v; want prior executable bytes", data, readErr)
	}
	if _, err := os.Stat(artifact); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("public artifact stat error=%v; want absent after injected rollback failure", err)
	}
	receiptData, readErr := os.ReadFile(priorReceipt)
	if readErr != nil || string(receiptData) != "prior receipt bytes" {
		t.Fatalf("prior public receipt=%q err=%v; want prior receipt bytes", receiptData, readErr)
	}
}

func TestPhase24PositiveTransferNativeApplication(t *testing.T) {
	sourcePath := testsupport.ProjectPath("examples", "phase24", "transfer.schway")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("transfer fixture parse diagnostics: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("transfer fixture checker diagnostics: %+v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("transfer fixture core validation problems: %+v", validated.Problems)
	}
	cSource, err := cgen.EmitApplication(validated.Program())
	if err != nil {
		t.Fatalf("emit transferred-owner application: %v", err)
	}
	manifest := testsupport.ProjectPath("examples", "phase23", "file_byte.bindings.json")
	artifact := filepath.Join(t.TempDir(), "phase24-transfer")
	buildCtx, cancelBuild := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancelBuild()
	if _, err := BuildApplication(buildCtx, source, cSource, artifact, manifest); err != nil {
		t.Fatalf("build transferred-owner application: %v", err)
	}
	for _, test := range []struct {
		name     string
		byteVal  byte
		wantText string
	}{
		{name: "0x41", byteVal: 0x41, wantText: "65\n"},
		{name: "0x42", byteVal: 0x42, wantText: "66\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := filepath.Join(t.TempDir(), "caller-selected.bin")
			if err := os.WriteFile(input, []byte{test.byteVal}, 0o600); err != nil {
				t.Fatal(err)
			}
			var stdout, stderr bytes.Buffer
			runCtx, cancelRun := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancelRun()
			outcome, err := DefaultRunner().RunApplication(runCtx, artifact, input, &stdout, &stderr)
			if err != nil || outcome.Kind != RunExited || outcome.ExitCode != 0 || stdout.String() != test.wantText || stderr.Len() != 0 {
				t.Fatalf("application outcome=%+v err=%v stdout=%q stderr=%q; want %q", outcome, err, stdout.String(), stderr.String(), test.wantText)
			}
		})
	}
}
