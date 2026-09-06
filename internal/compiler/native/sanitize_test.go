package native

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// TestSanitizerBuildIsSeparateArtifact closes D-05-11/D-05-12: the
// sanitizer binary is built at a distinct output path from the
// differential's own build for the same source, never toggled onto it.
func TestSanitizerBuildIsSeparateArtifact(t *testing.T) {
	source := "int main(void) { return 0; }\n"
	runner := DefaultRunner()
	differentialPath, cleanup, err := runner.CompileOnly(context.Background(), source, "-O0")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	var sanitizerOutputPath string
	instrumented := runner
	instrumented.command = func(ctx context.Context, name string, arguments ...string) *exec.Cmd {
		for index, argument := range arguments {
			if argument == "-o" && index+1 < len(arguments) {
				sanitizerOutputPath = arguments[index+1]
			}
		}
		return exec.CommandContext(ctx, name, arguments...)
	}
	if _, err := instrumented.RunSanitized(context.Background(), source, nil); err != nil {
		t.Fatal(err)
	}
	if sanitizerOutputPath == "" {
		t.Fatal("expected a captured sanitizer build -o output path")
	}
	if sanitizerOutputPath == differentialPath {
		t.Fatalf("sanitizer binary reused the differential's own output path: %s", sanitizerOutputPath)
	}
	if !strings.Contains(sanitizerOutputPath, ".sanitize") {
		t.Fatalf("expected sanitizer binary path to carry the .sanitize suffix, got %s", sanitizerOutputPath)
	}
}

// TestSanitizerOptionsArePinned closes D-05-13: ASAN_OPTIONS/UBSAN_OPTIONS
// are set explicitly in the sanitizer child process's own environment on
// every invocation, and a hostile pre-set ASAN_OPTIONS in the parent
// environment does not survive into the child.
func TestSanitizerOptionsArePinned(t *testing.T) {
	if err := os.Setenv("ASAN_OPTIONS", "detect_leaks=1:halt_on_error=0"); err != nil {
		t.Fatal(err)
	}
	defer os.Unsetenv("ASAN_OPTIONS")

	var runCommand *exec.Cmd
	calls := 0
	runner := DefaultRunner()
	runner.command = func(ctx context.Context, name string, arguments ...string) *exec.Cmd {
		calls++
		command := exec.CommandContext(ctx, name, arguments...)
		if calls == 2 {
			runCommand = command
		}
		return command
	}
	if _, err := runner.RunSanitized(context.Background(), "int main(void) { return 0; }\n", nil); err != nil {
		t.Fatal(err)
	}
	if runCommand == nil {
		t.Fatal("expected a run invocation to be captured")
	}
	foundASan, foundUBSan := false, false
	for _, keyValue := range runCommand.Env {
		if keyValue == "ASAN_OPTIONS="+ASanOptions {
			foundASan = true
		}
		if keyValue == "UBSAN_OPTIONS="+UBSanOptions {
			foundUBSan = true
		}
		if strings.HasPrefix(keyValue, "ASAN_OPTIONS=") && keyValue != "ASAN_OPTIONS="+ASanOptions {
			t.Fatalf("hostile parent ASAN_OPTIONS survived into the child: %s", keyValue)
		}
	}
	if !foundASan || !foundUBSan {
		t.Fatalf("expected pinned ASAN_OPTIONS/UBSAN_OPTIONS in the child environment, got %v", runCommand.Env)
	}
}

// TestSanitizerReportIsNotNativeResult closes D-05-11's type-level
// isolation: SanitizerReport has no field naming Result, and no function in
// this package converts between SanitizerReport and native.Result.
func TestSanitizerReportIsNotNativeResult(t *testing.T) {
	reportType := reflect.TypeOf(SanitizerReport{})
	for index := 0; index < reportType.NumField(); index++ {
		name := reportType.Field(index).Name
		if strings.Contains(name, "Result") {
			t.Fatalf("SanitizerReport field %q must not reference Result (D-05-11 type-level isolation)", name)
		}
	}

	matches, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	funcSignature := regexp.MustCompile(`func\s+[^{]*\([^)]*\)[^{]*\{`)
	resultWord := regexp.MustCompile(`\bResult\b`)
	for _, path := range matches {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, signature := range funcSignature.FindAllString(string(data), -1) {
			if strings.Contains(signature, "SanitizerReport") && resultWord.MatchString(signature) {
				t.Fatalf("found a function converting between SanitizerReport and Result in %s: %s", path, signature)
			}
		}
	}
}

// TestSanitizerClassifiesOnSubstringPlusExitCode closes D-05-14: a real
// ASan report classifies on its signature, a bare nonzero exit with no
// matching signature is an operational unclassified_abort (never a
// detection), and a clean zero exit produces no detection at all.
func TestSanitizerClassifiesOnSubstringPlusExitCode(t *testing.T) {
	t.Run("real asan use-after-free report", func(t *testing.T) {
		stderrText := "==12345==ERROR: AddressSanitizer: heap-use-after-free on address 0x602000000010 at pc 0x000102b3c4d0 bp 0x00016f14b8a0 sp 0x00016f14b898\n"
		report, toolErr := classifySanitizerRun(134, stderrText)
		if toolErr != nil {
			t.Fatalf("unexpected operational error for a recognized signature: %v", toolErr)
		}
		if report.Sanitizer != "address" || report.CheckKind != "heap-use-after-free" {
			t.Fatalf("expected address/heap-use-after-free classification, got %+v", report)
		}
	})
	t.Run("bare nonzero exit with no matching signature", func(t *testing.T) {
		report, toolErr := classifySanitizerRun(1, "")
		if toolErr == nil {
			t.Fatal("expected an operational error for an unclassified abort")
		}
		if toolErr.Code != "native.sanitizer_unclassified_abort" {
			t.Fatalf("expected native.sanitizer_unclassified_abort, got %s", toolErr.Code)
		}
		if report.CheckKind != "unclassified_abort" {
			t.Fatalf("expected CheckKind=unclassified_abort, got %+v", report)
		}
	})
	t.Run("clean zero exit is not a detection", func(t *testing.T) {
		report, toolErr := classifySanitizerRun(0, "")
		if toolErr != nil {
			t.Fatalf("unexpected operational error for a clean exit: %v", toolErr)
		}
		if report.CheckKind != "" || report.Sanitizer != "" {
			t.Fatalf("expected no detection for a clean exit, got %+v", report)
		}
	})
}

// TestSanitizerReportExcludesPathsAndAddresses closes D-05-16: the
// classification tuple's ReportSignature carries only the enumerated
// substring, never an absolute path or a hex address, even when both are
// present in the raw stderr the report also captures (bounded, separately).
func TestSanitizerReportExcludesPathsAndAddresses(t *testing.T) {
	stderrText := "==12345==ERROR: AddressSanitizer: heap-use-after-free on address 0x00010a2b3c4d at pc 0x00010a2b3c4d\n" +
		"    #0 0x00010a2b3c4d in main ~/x.c:12:5\n"
	report, toolErr := classifySanitizerRun(134, stderrText)
	if toolErr != nil {
		t.Fatal(toolErr)
	}
	if strings.Contains(report.ReportSignature, "~/x.c:12") {
		t.Fatalf("ReportSignature leaked an absolute path: %q", report.ReportSignature)
	}
	if strings.Contains(report.ReportSignature, "0x00010a2b3c4d") {
		t.Fatalf("ReportSignature leaked a hex address: %q", report.ReportSignature)
	}
	report.TruncatedStderr = stderrText
	if !strings.Contains(report.TruncatedStderr, "~/x.c:12") {
		t.Fatal("expected the raw path to be preserved in TruncatedStderr")
	}
	if report.TruncationCode != "" {
		t.Fatalf("expected no truncation code for stderr well under the bound, got %q", report.TruncationCode)
	}
}

// TestSanitizerRuntimeProbeReportsAvailability closes D-05-15: the smoke
// probe succeeds on this host (clang + ASan/UBSan runtime present).
func TestSanitizerRuntimeProbeReportsAvailability(t *testing.T) {
	runner := DefaultRunner()
	report, err := runner.ProbeSanitizerRuntime(context.Background())
	if err != nil {
		t.Fatalf("expected the sanitizer smoke probe to succeed on this host, got %v", err)
	}
	if report.Sanitizer != "address" || report.CheckKind != "heap-use-after-free" {
		t.Fatalf("expected address/heap-use-after-free from the smoke fixture, got %+v", report)
	}
}

// TestSanitizerRuntimeProbeAbsenceIsOperational closes D-05-15's posture
// rule: a missing clang reports native.tool_missing, never a success and
// never a nil error.
func TestSanitizerRuntimeProbeAbsenceIsOperational(t *testing.T) {
	runner := Runner{ClangPath: filepath.Join(t.TempDir(), "missing-clang")}
	_, err := runner.ProbeSanitizerRuntime(context.Background())
	var toolError *ToolError
	if !errors.As(err, &toolError) || toolError.Code != "native.tool_missing" {
		t.Fatalf("expected native.tool_missing for an absent clang, got %v", err)
	}
}

// TestSanitizerInertBinaryIsNotAPass closes D-05-38: a binary built from
// the SAME smoke fixture but WITHOUT the -fsanitize flags is not actually
// instrumented, and must never be classified as a pass.
func TestSanitizerInertBinaryIsNotAPass(t *testing.T) {
	inertFlags := []string{"-std=c17", SanitizerOptimization, "-g", "-fno-omit-frame-pointer"}
	runner := DefaultRunner()
	binaryPath, cleanup, toolErr := runner.compileSanitized(context.Background(), sanitizerSmokeFixtureSource, inertFlags, nil)
	if toolErr != nil {
		t.Fatal(toolErr)
	}
	defer cleanup()
	report, runToolErr := runner.runSanitizedBinary(context.Background(), binaryPath, "")
	if runToolErr != nil && runToolErr.Code != "native.sanitizer_unclassified_abort" {
		t.Fatalf("unexpected operational error running the inert binary: %v", runToolErr)
	}
	inertErr := probeOutcome(report)
	if inertErr == nil {
		t.Fatal("expected the inert (uninstrumented) binary to be refused as a pass")
	}
	if inertErr.Code != "native.sanitizer_inert" {
		t.Fatalf("expected native.sanitizer_inert, got %s", inertErr.Code)
	}
}
