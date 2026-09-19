package cgen_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// restrictReadonlyProbe is deliberately hand-written C17 rather than emitted
// C. It is D-16-06's only normative shape: a single TU, one restricted
// pointer, and a read/copy-only callee. It does not prove a general Lang
// borrow-to-restrict rule or LTO non-inertness.
const restrictReadonlyProbe = `#include <stdint.h>

typedef uint8_t T;

static T f(T *restrict p) {
  T copy = *p;
  return copy;
}

int main(void) {
  T value = 37;
  T before = value;
  T copied = f(&value);
  T after = value;
  return before == 37 && copied == 37 && after == 37 ? 0 : 1;
}
`

type restrictProbeResult struct {
	Host       string
	Toolchain  string
	Lane       string
	Result     string
	SourceSHA  string
	Diagnostic string
}

func restrictProbeSource(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase16", "restrict_readonly_probe.c"))
	if err != nil {
		t.Fatal(err)
	}
	if string(source) != restrictReadonlyProbe {
		t.Fatalf("restrict probe differs from the D-16-06 normative source:\n--- got ---\n%s\n--- want ---\n%s", source, restrictReadonlyProbe)
	}
	return source
}

func restrictProbeDigest(source []byte) string {
	sum := sha256.Sum256(source)
	return hex.EncodeToString(sum[:])
}

func TestRestrictReadonlyProbeSourceShape(t *testing.T) {
	source := restrictProbeSource(t)
	forbidden := []string{"*p =", "T *q", "return p", "callback", "extern", "volatile", "_Atomic"}
	for _, fragment := range forbidden {
		if strings.Contains(string(source), fragment) {
			t.Fatalf("D-16-06 probe contains prohibited fragment %q", fragment)
		}
	}
	if count := strings.Count(string(source), "*restrict "); count != 1 {
		t.Fatalf("restricted pointer count=%d, want 1", count)
	}
	if !strings.Contains(string(source), "static T f(T *restrict p)") {
		t.Fatal("D-16-06 probe must declare static T f(T *restrict p)")
	}
	t.Logf("restrict_probe host=%s lane=source-shape result=PASS source_sha256=%s", runtime.GOOS, restrictProbeDigest(source))
}

func TestRestrictReadonlyProbeDarwinLanes(t *testing.T) {
	runRestrictProbeHostLanes(t, "darwin")
}

func TestRestrictReadonlyProbeLinuxLanes(t *testing.T) {
	runRestrictProbeHostLanes(t, "linux")
}

func runRestrictProbeHostLanes(t *testing.T, requiredHost string) {
	t.Helper()
	source := restrictProbeSource(t)
	digest := restrictProbeDigest(source)
	if runtime.GOOS != requiredHost {
		t.Logf("restrict_probe host=%s required_host=%s lane=all result=UNAVAILABLE source_sha256=%s", runtime.GOOS, requiredHost, digest)
		t.Skipf("UNAVAILABLE: restrict probe requires host=%s (source_sha256=%s)", requiredHost, digest)
	}

	clang, err := exec.LookPath("clang")
	if err != nil {
		t.Fatalf("restrict_probe host=%s lane=all result=FAIL: installed clang is required: %v", requiredHost, err)
	}
	lanes := []struct {
		name  string
		flags []string
	}{
		{name: "O0", flags: []string{"-O0"}},
		{name: "O3", flags: []string{"-O3"}},
		{name: "O3-LTO", flags: []string{"-O3", "-flto"}},
		{name: "ASan-UBSan", flags: []string{"-O1", "-fsanitize=address,undefined", "-fno-omit-frame-pointer"}},
	}
	for _, lane := range lanes {
		lane := lane
		t.Run(lane.name, func(t *testing.T) {
			result := runRestrictProbeLane(t, requiredHost, clang, lane.name, lane.flags, source, digest)
			if result.Result != "PASS" {
				t.Fatalf("restrict_probe host=%s toolchain=%s lane=%s result=%s source_sha256=%s diagnostic=%s", result.Host, result.Toolchain, result.Lane, result.Result, result.SourceSHA, result.Diagnostic)
			}
			t.Logf("restrict_probe host=%s toolchain=%s lane=%s result=PASS source_sha256=%s", result.Host, result.Toolchain, result.Lane, result.SourceSHA)
		})
	}
}

func runRestrictProbeLane(t *testing.T, host, clang, lane string, flags []string, source []byte, digest string) restrictProbeResult {
	t.Helper()
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "restrict_readonly_probe.c")
	binaryPath := filepath.Join(dir, "restrict_readonly_probe")
	if err := os.WriteFile(sourcePath, source, 0o600); err != nil {
		t.Fatal(err)
	}
	compiled, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if got := restrictProbeDigest(compiled); got != digest {
		t.Fatalf("lane %s compiled different source bytes: got=%s want=%s", lane, got, digest)
	}
	args := append([]string{"-std=c17", "-Wall", "-Wextra", "-Werror"}, flags...)
	args = append(args, sourcePath, "-o", binaryPath)
	output, err := exec.Command(clang, args...).CombinedOutput()
	if err != nil {
		return restrictProbeResult{Host: host, Toolchain: clang, Lane: lane, Result: "FAIL", SourceSHA: digest, Diagnostic: fmt.Sprintf("compile: %v: %s", err, strings.TrimSpace(string(output)))}
	}
	output, err = exec.Command(binaryPath).CombinedOutput()
	if err != nil {
		return restrictProbeResult{Host: host, Toolchain: clang, Lane: lane, Result: "FAIL", SourceSHA: digest, Diagnostic: fmt.Sprintf("execute: %v: %s", err, strings.TrimSpace(string(output)))}
	}
	return restrictProbeResult{Host: host, Toolchain: clang, Lane: lane, Result: "PASS", SourceSHA: digest}
}

// restrictAdmissionShape is deliberately test-local. D-16-07 requires an
// auditable candidate boundary before a later decision may consider routing;
// this type neither parses Lang nor changes cgen production selection.
type restrictAdmissionShape struct {
	pointerCount          int
	readCopyOnly          bool
	pointeeMutation       bool
	pointerEscape         bool
	pointerForwarding     bool
	callback              bool
	foreignCall           bool
	volatileAccess        bool
	atomicAccess          bool
	oneTranslationUnit    bool
	callerLocalAccessOnly bool
}

func exactRestrictAdmissionShape() restrictAdmissionShape {
	return restrictAdmissionShape{
		pointerCount:          1,
		readCopyOnly:          true,
		oneTranslationUnit:    true,
		callerLocalAccessOnly: true,
	}
}

// restrictAdmissionReason is the smallest auditable D-16-07 candidate fence.
// It accepts only the hand-written probe's read/copy-only, caller-local,
// one-TU, one-pointer shape; it is test evidence, not production routing.
func restrictAdmissionReason(shape restrictAdmissionShape) string {
	if shape.pointerCount != 1 {
		return "second-pointer"
	}
	if !shape.readCopyOnly {
		return "not-read-copy-only"
	}
	if shape.pointeeMutation {
		return "pointee-mutation"
	}
	if shape.pointerEscape {
		return "pointer-escape"
	}
	if shape.pointerForwarding {
		return "pointer-forwarding"
	}
	if shape.callback {
		return "callback"
	}
	if shape.foreignCall {
		return "foreign-call"
	}
	if shape.volatileAccess {
		return "volatile-access"
	}
	if shape.atomicAccess {
		return "atomic-access"
	}
	if !shape.oneTranslationUnit {
		return "separate-compilation"
	}
	if !shape.callerLocalAccessOnly {
		return "non-local-caller-access"
	}
	return ""
}

func TestRestrictAdmissionExactShape(t *testing.T) {
	if reason := restrictAdmissionReason(exactRestrictAdmissionShape()); reason != "" {
		t.Fatalf("exact D-16-06 shape refused: %s", reason)
	}
}

func TestRestrictAdmissionRejectsExtensions(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*restrictAdmissionShape)
		want   string
	}{
		{"pointee mutation", func(shape *restrictAdmissionShape) { shape.pointeeMutation = true }, "pointee-mutation"},
		{"second pointer", func(shape *restrictAdmissionShape) { shape.pointerCount = 2 }, "second-pointer"},
		{"pointer escape", func(shape *restrictAdmissionShape) { shape.pointerEscape = true }, "pointer-escape"},
		{"pointer forwarding", func(shape *restrictAdmissionShape) { shape.pointerForwarding = true }, "pointer-forwarding"},
		{"callback", func(shape *restrictAdmissionShape) { shape.callback = true }, "callback"},
		{"foreign call", func(shape *restrictAdmissionShape) { shape.foreignCall = true }, "foreign-call"},
		{"volatile access", func(shape *restrictAdmissionShape) { shape.volatileAccess = true }, "volatile-access"},
		{"atomic access", func(shape *restrictAdmissionShape) { shape.atomicAccess = true }, "atomic-access"},
		{"separate compilation", func(shape *restrictAdmissionShape) { shape.oneTranslationUnit = false }, "separate-compilation"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			shape := exactRestrictAdmissionShape()
			test.mutate(&shape)
			if got := restrictAdmissionReason(shape); got != test.want {
				t.Fatalf("extension %q admission reason=%q, want %q", test.name, got, test.want)
			}
		})
	}
}

func TestRestrictAdmissionFenceIsNotInert(t *testing.T) {
	shape := exactRestrictAdmissionShape()
	shape.pointeeMutation = true
	if got := restrictAdmissionBypassReason(shape); got != "" {
		t.Fatalf("bypass control unexpectedly rejected mutation: %q", got)
	}
	if got := restrictAdmissionReason(shape); got == "" {
		t.Fatal("normal fence admitted a bypassed mutation")
	}
}

// restrictAdmissionBypassReason is the local negative control: it models the
// earlier one-pointer-only rule so the suite proves its D-16-07 checks are
// live rather than merely present as unused prose.
func restrictAdmissionBypassReason(shape restrictAdmissionShape) string {
	if shape.pointerCount != 1 {
		return "second-pointer"
	}
	return ""
}
