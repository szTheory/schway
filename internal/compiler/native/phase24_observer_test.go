package native

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const (
	phase24AcquireOperationEnv = "SCHWAY_PHASE24_ACQUIRE_OPERATION_ID"
	phase24UseError            = "UseError.UnsupportedByte\n"
)

type phase24ObserverFixture struct {
	manifest  string
	source    []byte
	cSource   string
	program   core.Program
	acquireID string
	useID     string
	probeFn   string
	mainFn    string
}

func TestPhase24ObserverPublicLifecycleAndTypedError(t *testing.T) {
	fixture := newPhase24ObserverFixture(t)
	artifact := fixture.build(t, fixture.cSource)
	for _, test := range []struct {
		name       string
		value      byte
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{name: "0x41", value: 0x41, wantCode: 0, wantStdout: "65\n"},
		{name: "0x42", value: 0x42, wantCode: 0, wantStdout: "66\n"},
		{name: "0x43 typed error", value: 0x43, wantCode: 65, wantStderr: phase24UseError},
	} {
		t.Run(test.name, func(t *testing.T) {
			phase24AssertModelResult(t, fixture, test.value)
			code, stdout, stderr, receipt := fixture.run(t, artifact, test.value, false)
			if code != test.wantCode || stdout != test.wantStdout || stderr != test.wantStderr {
				t.Fatalf("native app code=%d stdout=%q stderr=%q; want %d/%q/%q", code, stdout, stderr, test.wantCode, test.wantStdout, test.wantStderr)
			}
			wantReceipt := fixture.expectedReceipt(false, test.value == 0x43)
			if string(receipt) != wantReceipt {
				t.Fatalf("independent physical receipt=%q; want actual pointer lifecycle %q", receipt, wantReceipt)
			}
		})
	}
}

func TestPhase24ObserverReachedPhysicalDestructorControls(t *testing.T) {
	fixture := newPhase24ObserverFixture(t)
	for _, test := range []struct {
		name      string
		mutate    func(*testing.T, string) string
		collision bool
		marker    string
	}{
		{
			name: "omitted release",
			mutate: func(t *testing.T, source string) string {
				return phase24MutateUseErrorRelease(t, source, "omit")
			},
			marker: "attempt\tomitted-release\n",
		},
		{
			name: "premature release",
			mutate: func(t *testing.T, source string) string {
				return phase24MutateUseErrorRelease(t, source, "premature")
			},
			marker: "attempt\tpremature-release\n",
		},
		{
			name: "duplicate release",
			mutate: func(t *testing.T, source string) string {
				return phase24MutateUseErrorRelease(t, source, "duplicate")
			},
			marker: "attempt\tduplicate-release\n",
		},
		{
			name: "wrong resource",
			mutate: func(t *testing.T, source string) string {
				return phase24MutateUseErrorRelease(t, source, "wrong-resource")
			},
			marker: "attempt\twrong-resource\n",
		},
		{
			name:      "semantic identity collision",
			collision: true,
			marker:    "attempt\tidentity-collision\n",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			cSource := fixture.cSource
			if test.mutate != nil {
				cSource = test.mutate(t, cSource)
			}
			artifact := fixture.build(t, cSource)
			phase24AssertModelResult(t, fixture, 0x43)
			code, stdout, stderr, receipt := fixture.run(t, artifact, 0x43, test.collision)
			if code != 86 || stdout != "" || stderr != phase24UseError {
				t.Fatalf("mutated app code=%d stdout=%q stderr=%q; want reached observer rejection 86 and the ordinary typed error", code, stdout, stderr)
			}
			if !strings.Contains(string(receipt), test.marker) || !strings.HasSuffix(string(receipt), "verdict\treject\n") {
				t.Fatalf("physical observer did not reach and reject %s: %q", test.name, receipt)
			}
			if !strings.Contains(string(receipt), "report_boundary\t") {
				t.Fatalf("typed error reporting boundary was not reached for %s: %q", test.name, receipt)
			}
		})
	}
}

func newPhase24ObserverFixture(t *testing.T) *phase24ObserverFixture {
	t.Helper()
	project := testsupport.ProjectPath()
	root := t.TempDir()
	for _, name := range []string{"adapter.h", "file_byte.bindings.json"} {
		data, err := os.ReadFile(filepath.Join(project, "examples", "phase23", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, name), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	adapter, err := os.ReadFile(filepath.Join(project, "examples", "phase23", "adapter.c"))
	if err != nil {
		t.Fatal(err)
	}
	const include = "#include \"adapter.h\"\n"
	const observerHooks = "#include \"adapter.h\"\n\n" +
		"void *phase24_observer_malloc(size_t);\n" +
		"void phase24_observer_free(void *);\n" +
		"unsigned char phase24_observer_load_byte(const unsigned char *);\n\n" +
		"#define SCHWAY_FILE_BYTE_MALLOC phase24_observer_malloc\n" +
		"#define SCHWAY_FILE_BYTE_FREE phase24_observer_free\n" +
		"#define SCHWAY_FILE_BYTE_LOAD_BYTE(pointer) phase24_observer_load_byte(pointer)\n"
	adapterText := strings.Replace(string(adapter), include, observerHooks, 1)
	if adapterText == string(adapter) {
		t.Fatal("Phase 24 observer hooks were not inserted at the adapter include boundary")
	}
	if err := os.WriteFile(filepath.Join(root, "adapter.c"), []byte(adapterText), 0o600); err != nil {
		t.Fatal(err)
	}
	observer, err := os.ReadFile(filepath.Join(project, "internal", "compiler", "native", "testdata", "phase24_observer.c"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "phase24_observer.c"), observer, 0o600); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(root, "file_byte.bindings.json")
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest BindingManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.Sources = append(manifest.Sources, "phase24_observer.c")
	manifestBytes, err = json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, manifestBytes, 0o600); err != nil {
		t.Fatal(err)
	}

	source, err := os.ReadFile(filepath.Join(project, "examples", "phase24", "error.schway"))
	if err != nil {
		t.Fatal(err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("Phase 24 source parse diagnostics: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("Phase 24 source check diagnostics: %+v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("Phase 24 core validation problems: %+v", validated.Problems)
	}
	program := validated.Program()
	var acquireID, useID, acquireFn, probeFn, mainFn string
	for _, function := range program.Functions {
		switch function.Name {
		case "acquire":
			acquireFn = function.ID
		case "probe":
			probeFn = function.ID
		case "main":
			mainFn = function.ID
		}
		if function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Foreign == nil {
				continue
			}
			if function.Name == "acquire" && operation.Foreign.Mode == "acquire" {
				acquireID = operation.ID
			}
			if function.Name == "probe" && operation.Foreign.Mode == "borrow" {
				useID = operation.ID
			}
		}
	}
	if acquireID == "" || useID == "" || acquireFn == "" || probeFn == "" || mainFn == "" || strings.ContainsAny(acquireID, "\t\r\n") {
		t.Fatalf("Phase 24 static/dynamic identity facts are incomplete or unsafe for a receipt: acquire=%q use=%q functions=%q/%q/%q", acquireID, useID, acquireFn, probeFn, mainFn)
	}
	cSource, err := cgen.EmitApplication(program)
	if err != nil {
		t.Fatalf("emit Phase 24 typed-error application: %v", err)
	}
	cSource = phase24InstrumentTransferAndErrorBoundaries(t, cSource)
	return &phase24ObserverFixture{
		manifest: manifestPath, source: source, cSource: cSource,
		program: program, acquireID: acquireID, useID: useID,
		probeFn: probeFn, mainFn: mainFn,
	}
}

func phase24InstrumentTransferAndErrorBoundaries(t *testing.T, cSource string) string {
	t.Helper()
	const helperReturn = "if (!schway_phase24_acquire(path, &schway_owner_c, error)) return 1;\n"
	const observedHelperReturn = helperReturn + "  phase24_observer_transfer_return();\n"
	if strings.Count(cSource, helperReturn) != 1 {
		t.Fatal("generated Phase 24 probe has no unique successful helper-return boundary")
	}
	const cleanupAndReport = `schway_file_byte_release(schway_owner_b); schway_file_byte_release(schway_owner_a); fprintf(stderr, "%s\n", error);`
	const observedCleanupAndReport = `schway_file_byte_release(schway_owner_b); schway_file_byte_release(schway_owner_a); phase24_observer_error_report_boundary(); fprintf(stderr, "%s\n", error);`
	if strings.Count(cSource, cleanupAndReport) != 1 {
		t.Fatal("generated Phase 24 error branch does not have one caller cleanup/report boundary")
	}
	cSource = strings.Replace(cSource, helperReturn, observedHelperReturn, 1)
	cSource = strings.Replace(cSource, cleanupAndReport, observedCleanupAndReport, 1)
	return "extern void phase24_observer_transfer_return(void);\n" +
		"extern void phase24_observer_error_report_boundary(void);\n" + cSource
}

func (fixture *phase24ObserverFixture) build(t *testing.T, cSource string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	artifact := filepath.Join(t.TempDir(), "phase24-error")
	if _, err := BuildApplication(ctx, fixture.source, cSource, artifact, fixture.manifest); err != nil {
		t.Fatalf("build observed Phase 24 application: %v", err)
	}
	return artifact
}

func (fixture *phase24ObserverFixture) run(t *testing.T, artifact string, inputValue byte, collision bool) (int, string, string, []byte) {
	t.Helper()
	input := filepath.Join(t.TempDir(), "caller-selected-byte.bin")
	if err := os.WriteFile(input, []byte{inputValue}, 0o600); err != nil {
		t.Fatal(err)
	}
	receiptPath := filepath.Join(t.TempDir(), "phase24-physical-lifecycle.tsv")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, artifact, input)
	command.Env = make([]string, 0, len(os.Environ())+3)
	for _, entry := range os.Environ() {
		if strings.HasPrefix(entry, "SCHWAY_PHASE24_OBSERVER_PATH=") ||
			strings.HasPrefix(entry, phase24AcquireOperationEnv+"=") ||
			strings.HasPrefix(entry, "SCHWAY_PHASE24_OBSERVER_COLLIDE=") {
			continue
		}
		command.Env = append(command.Env, entry)
	}
	collisionSetting := "0"
	if collision {
		collisionSetting = "2"
	}
	command.Env = append(command.Env,
		"SCHWAY_PHASE24_OBSERVER_PATH="+receiptPath,
		phase24AcquireOperationEnv+"="+fixture.acquireID,
		"SCHWAY_PHASE24_OBSERVER_COLLIDE="+collisionSetting,
	)
	var stdout, stderr strings.Builder
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	if ctx.Err() != nil {
		t.Fatalf("observed Phase 24 app timed out: %v", ctx.Err())
	}
	code := 0
	if err != nil {
		var exitError *exec.ExitError
		if !errors.As(err, &exitError) {
			t.Fatalf("observed Phase 24 app could not launch: %v", err)
		}
		code = exitError.ExitCode()
	}
	receipt, err := os.ReadFile(receiptPath)
	if err != nil {
		t.Fatalf("read independent Phase 24 physical observer receipt: %v", err)
	}
	return code, stdout.String(), stderr.String(), receipt
}

func (fixture *phase24ObserverFixture) expectedReceipt(collision, errorPath bool) string {
	activation := []int{1, 2, 3}
	if collision {
		activation[1] = 1
	}
	var receipt strings.Builder
	for index, id := range activation {
		if collision && index == 1 {
			receipt.WriteString("attempt\tidentity-collision\n")
		}
		fmt.Fprintf(&receipt, "malloc\t%s@activation%d\n", fixture.acquireID, id)
	}
	fmt.Fprintf(&receipt, "transfer_return\t%s@activation3\n", fixture.acquireID)
	fmt.Fprintf(&receipt, "use\t%s@activation3\n", fixture.acquireID)
	for _, id := range []int{3, 2, 1} {
		fmt.Fprintf(&receipt, "free\t%s@activation%d\n", fixture.acquireID, id)
	}
	if errorPath {
		receipt.WriteString("report_boundary\tpost-release\n")
	}
	receipt.WriteString("final\toutstanding=0\n")
	if collision {
		receipt.WriteString("verdict\treject\n")
	} else {
		receipt.WriteString("verdict\tpass\n")
	}
	return receipt.String()
}

func phase24MutateUseErrorRelease(t *testing.T, cSource, mutation string) string {
	t.Helper()
	const release = "schway_file_byte_release(schway_owner_c);"
	const use = "schway_file_byte_use_result use = schway_file_byte_use(schway_owner_c);"
	const errorBranch = `if (use.status != SCHWAY_FILE_BYTE_USE_OK) { schway_file_byte_release(schway_owner_c); *error = "UseError.UnsupportedByte"; return 2; }`
	if strings.Count(cSource, errorBranch) != 1 {
		t.Fatal("generated Phase 24 C has no unique typed-use-error cleanup branch")
	}
	switch mutation {
	case "omit":
		return strings.Replace(cSource, release+` *error = "UseError.UnsupportedByte"`, `*error = "UseError.UnsupportedByte"`, 1)
	case "premature":
		if strings.Count(cSource, use) != 1 {
			t.Fatal("generated Phase 24 C has no unique transferred-owner use")
		}
		return strings.Replace(cSource, use, release+"\n  "+use, 1)
	case "duplicate":
		return strings.Replace(cSource, errorBranch,
			strings.Replace(errorBranch, release, release+" "+release, 1), 1)
	case "wrong-resource":
		const probeUse = "  schway_file_byte_use_result use = schway_file_byte_use(schway_owner_c);\n"
		if strings.Count(cSource, probeUse) != 1 {
			t.Fatal("generated Phase 24 C has no unique probe use boundary for wrong-resource mutation")
		}
		withDecoy := "  unsigned char phase24_decoy_byte = 0u;\n" +
			"  schway_file_byte_owner phase24_decoy_owner = {&phase24_decoy_byte, UINT64_C(1)};\n" + probeUse
		mutated := strings.Replace(cSource, probeUse, withDecoy, 1)
		wrongBranch := strings.Replace(errorBranch, release, "schway_file_byte_release(phase24_decoy_owner);", 1)
		return strings.Replace(mutated, errorBranch, wrongBranch, 1)
	default:
		t.Fatalf("unknown Phase 24 physical mutation %q", mutation)
		return cSource
	}
}

func phase24AssertModelResult(t *testing.T, fixture *phase24ObserverFixture, input byte) {
	t.Helper()
	byteValue := int(input)
	useOutcome := interp.ForeignOutcome{Kind: "success", Type: "U64", Value: fmt.Sprint(byteValue)}
	if input == 0x43 {
		useOutcome = interp.ForeignOutcome{Kind: "failure", Type: "UseError", Value: "UnsupportedByte"}
	}
	result, err := interp.RunWithForeignOutcomes(fixture.program, "main", "opaque-path-token", map[string]interp.ForeignOutcome{
		fixture.acquireID: {Kind: "success", Type: "FileByteOwner", Value: fmt.Sprint(byteValue)},
		fixture.useID:     useOutcome,
	})
	if err != nil {
		t.Fatalf("run independent Phase 24 model: %v", err)
	}
	if result.ActualHostIO || result.PhysicalCleanup || result.EvidenceScope != interp.EvidenceScopeModelOnly || len(result.Execution.LiveResources) != 0 {
		t.Fatalf("model evidence scope/live owners=%+v; want model-only, no host IO, no physical cleanup, and no semantic obligations", result)
	}
	if input == 0x43 {
		if result.Execution.Outcome.Kind != "typed_failure" || result.Execution.Outcome.Value != "UnsupportedByte" {
			t.Fatalf("model typed outcome=%+v; want UseError.UnsupportedByte", result.Execution.Outcome)
		}
		var acquisitions, releases []string
		for _, event := range result.Execution.Events {
			if event.ID == fixture.acquireID+":event" {
				acquisitions = append(acquisitions, event.Invocation)
			}
			if event.Kind == "resource.released" {
				releases = append(releases, event.FunctionID)
			}
		}
		if len(acquisitions) != 3 || acquisitions[0] == "" || acquisitions[0] == acquisitions[1] || acquisitions[1] == acquisitions[2] || acquisitions[0] == acquisitions[2] {
			t.Fatalf("model repeated acquisition activations=%v; want three distinct activations", acquisitions)
		}
		if len(releases) != 3 || releases[0] != fixture.probeFn || releases[1] != fixture.mainFn || releases[2] != fixture.mainFn {
			t.Fatalf("model cleanup event functions=%v; want helper C then caller B,A", releases)
		}
		return
	}
	want := "65"
	if input == 0x42 {
		want = "66"
	}
	if result.Execution.Outcome.Kind != "returned" || result.Execution.Outcome.Value != want {
		t.Fatalf("model success outcome=%+v; want fixed independent value %s", result.Execution.Outcome, want)
	}
}
