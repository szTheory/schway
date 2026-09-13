package session

import (
	"bytes"
	"os"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// phase13Fixture reads a fixture from testdata/phase13 by name, mirroring
// phase6Fixture's shape exactly (session_phase6_injectors_test.go).
func phase13Fixture(t *testing.T, name string) []byte {
	t.Helper()
	path := testsupport.ProjectPath("testdata", "phase13", name)
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", name, err)
	}
	return source
}

// The three inline base programs below are Task 1's own self-contained
// fixtures -- Task 1 lands BEFORE Task 2 authors the real testdata/phase13
// corpus, so TestPhase13InjectorsProduceExactlyOneDefect and
// TestPhase13InjectorGuardsAreNotInert (Task 1's own required tests) must
// not depend on files Task 2 has not written yet. Each mirrors the SHAPE
// of its corresponding testdata/phase13 corpus member Task 2 later adds,
// so both find and exercise the identical defect mechanism.

var phase13InjectorLoanBase = []byte(`module phase13.injector_base_loan

export {
  fn relay
}

fn sink(buffer: Buffer) -> Buffer {
  let held = take buffer
  held
}

fn relay(buffer: Buffer) -> Buffer {
  let borrowed = borrow buffer
  let routed = sink(borrowed)
  let delivered = take buffer // lang:interprocedural-loan-target
  routed
}`)

var phase13InjectorConsumeBase = []byte(`module phase13.injector_base_consume

export {
  fn main
}

foreign C {

  fn lang_res_open(request: Byte) -> Byte {
    unwind: forbidden
    nonlocal_exit: forbidden
    allocator: "libc_malloc"
    fails: AcquireError
  }
}

data AcquireError =
  | OpenFailed

fn main(request: Byte) -> Byte {
  let handle = try lang_res_open(request) // lang:fallible-consume-target
  handle
}`)

var phase13InjectorArgumentBase = []byte(`module phase13.injector_base_argument

export {
  fn main
}

fn identity(value: Byte) -> Byte { // lang:call-argument-target
  value
}

fn main(value: Byte) -> Byte {
  let result = identity(value)
  result
}`)

// TestPhase13InjectorsProduceExactlyOneDefect is Task 1's central
// assertion: each of the three phase-13 injectors, applied to its own
// clean base, produces a mutated source that checks with EXACTLY ONE
// diagnostic, of exactly the code that injector's defect class names.
func TestPhase13InjectorsProduceExactlyOneDefect(t *testing.T) {
	cases := []struct {
		name     string
		injector Injector
		base     []byte
		wantCode string
	}{
		{"interprocedural_loan", InterproceduralLoanInjector{}, phase13InjectorLoanBase, "check.interprocedural_loan_liveness"},
		{"fallible_consume", FallibleConsumeInjector{}, phase13InjectorConsumeBase, "syntax.fallible_call_not_consumed"},
		{"call_argument_type", CallArgumentTypeInjector{}, phase13InjectorArgumentBase, "check.call_argument_type_mismatch"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			baseResult := Check(tc.base)
			if len(baseResult.Diagnostics) != 0 {
				t.Fatalf("%s: base did not check clean: %+v", tc.name, baseResult.Diagnostics)
			}
			mutated, err := tc.injector.Inject(tc.base)
			if err != nil {
				t.Fatalf("%s: Inject failed: %v", tc.name, err)
			}
			result := Check(mutated)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("%s: got %d diagnostics, want 1: %+v", tc.name, len(result.Diagnostics), result.Diagnostics)
			}
			if result.Diagnostics[0].Code != tc.wantCode {
				t.Fatalf("%s: got code %q, want %q", tc.name, result.Diagnostics[0].Code, tc.wantCode)
			}
		})
	}
}

// TestPhase13InjectorGuardsAreNotInert is Task 1's D-13-30(a) assertion:
// each injector's guard-disabled twin, given marker-free source, returns
// it completely unchanged (which therefore still checks clean) -- the
// silent pass its real, guarded Inject refuses instead.
func TestPhase13InjectorGuardsAreNotInert(t *testing.T) {
	cases := []struct {
		name      string
		marker    string
		base      []byte
		guardSkip func([]byte) []byte
		injector  Injector
	}{
		{"interprocedural_loan", loanTargetMarker, phase13InjectorLoanBase, interproceduralLoanInjectSkippingGuard, InterproceduralLoanInjector{}},
		{"fallible_consume", fallibleConsumeTargetMarker, phase13InjectorConsumeBase, fallibleConsumeInjectSkippingGuard, FallibleConsumeInjector{}},
		{"call_argument_type", callArgumentTargetMarker, phase13InjectorArgumentBase, callArgumentTypeInjectSkippingGuard, CallArgumentTypeInjector{}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			markerFree := bytes.ReplaceAll(tc.base, []byte(tc.marker), []byte(""))
			unguarded := tc.guardSkip(markerFree)
			if !bytes.Equal(unguarded, markerFree) {
				t.Fatalf("%s: guard-disabled twin mutated marker-free source; want unchanged", tc.name)
			}
			uncheckedResult := Check(unguarded)
			if len(uncheckedResult.Diagnostics) != 0 {
				t.Fatalf("%s: guard-disabled path's unmutated source did not check clean: %+v", tc.name, uncheckedResult.Diagnostics)
			}
			_, err := tc.injector.Inject(markerFree)
			typed := injectorError(err)
			if typed == nil || typed.Code != InjectorTargetMissingCode {
				t.Fatalf("%s: guarded Inject did not refuse on marker-free input: %v", tc.name, err)
			}
		})
	}
}
