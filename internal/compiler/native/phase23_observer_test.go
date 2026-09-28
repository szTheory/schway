package native

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/execution"
	"github.com/codename-lang/lang/internal/compiler/syntax"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

const phase23UseErrorDiagnostic = "fputs(\"lang_file_byte_use: UnsupportedByte\\n\", stderr);"

type phase23ObserverFixture struct {
	root     string
	manifest string
	source   []byte
	cSource  string
}

func TestPhase23ObserverPublicLifecycleAndUseFailure(t *testing.T) {
	cli := testsupport.BuildCLI(t)
	fixture := newPhase23ObserverFixture(t)

	for _, test := range []struct {
		name        string
		value       byte
		wantCode    int
		wantStdout  string
		wantStderr  string
		wantReceipt string
	}{
		{
			name: "0x41", value: 0x41, wantCode: 0, wantStdout: "65\n",
			wantReceipt: "malloc\tp1\nuse\tp1\nfree\tp1\nfinal\toutstanding=0\nverdict\tpass\n",
		},
		{
			name: "0x42", value: 0x42, wantCode: 0, wantStdout: "66\n",
			wantReceipt: "malloc\tp1\nuse\tp1\nfree\tp1\nfinal\toutstanding=0\nverdict\tpass\n",
		},
		{
			name: "0x43 typed use failure", value: 0x43, wantCode: 65,
			wantStderr:  "lang_file_byte_use: UnsupportedByte\n",
			wantReceipt: "malloc\tp1\nuse\tp1\nfree\tp1\nreport_boundary\tpost-release\nfinal\toutstanding=0\nverdict\tpass\n",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			artifact := fixture.build(t, fixture.cSource)
			input := filepath.Join(t.TempDir(), "input.bin")
			if err := os.WriteFile(input, []byte{test.value}, 0o600); err != nil {
				t.Fatal(err)
			}
			receiptPath := filepath.Join(t.TempDir(), "native-lifecycle.tsv")
			code, stdout, stderr := runPhase23ObserverApp(t, cli, artifact, input, receiptPath)
			if code != test.wantCode || stdout != test.wantStdout || stderr != test.wantStderr {
				t.Fatalf("app run code=%d stdout=%q stderr=%q; want %d/%q/%q", code, stdout, stderr, test.wantCode, test.wantStdout, test.wantStderr)
			}
			receipt, err := os.ReadFile(receiptPath)
			if err != nil {
				t.Fatalf("read independent native observer receipt: %v", err)
			}
			if string(receipt) != test.wantReceipt {
				t.Fatalf("native receipt=%q, want ordered pointer lifecycle %q", receipt, test.wantReceipt)
			}
		})
	}
}

func TestPhase23ObserverMutationControlsPreservePlausibleEvents(t *testing.T) {
	fixture := newPhase23ObserverFixture(t)
	for _, test := range []struct {
		name        string
		mutate      func(*testing.T, string) string
		wantCode    int
		wantStdout  string
		wantReceipt string
	}{
		{
			name:        "positive",
			wantCode:    0,
			wantStdout:  "65\n",
			wantReceipt: "malloc\tp1\nuse\tp1\nfree\tp1\nfinal\toutstanding=0\nverdict\tpass\n",
		},
		{
			name:        "omitted release",
			mutate:      func(t *testing.T, cSource string) string { return phase23MutateDestructor(t, cSource, "omit") },
			wantCode:    86,
			wantReceipt: "malloc\tp1\nuse\tp1\nattempt\tomitted-release\nfinal\toutstanding=1\nverdict\treject\n",
		},
		{
			name:        "premature release",
			mutate:      func(t *testing.T, cSource string) string { return phase23MutateDestructor(t, cSource, "premature") },
			wantCode:    86,
			wantReceipt: "malloc\tp1\nattempt\tpremature-release\nuse\tp1\nfinal\toutstanding=1\nverdict\treject\n",
		},
		{
			name:        "duplicate release",
			mutate:      func(t *testing.T, cSource string) string { return phase23MutateDestructor(t, cSource, "duplicate") },
			wantCode:    86,
			wantReceipt: "malloc\tp1\nuse\tp1\nfree\tp1\nattempt\tduplicate-release\nfinal\toutstanding=0\nverdict\treject\n",
		},
		{
			name: "wrong resource",
			mutate: func(t *testing.T, cSource string) string {
				return phase23MutateDestructor(t, cSource, "wrong-resource")
			},
			wantCode:    86,
			wantReceipt: "malloc\tp1\nuse\tp1\nfree\tp1\nattempt\twrong-resource\nfinal\toutstanding=0\nverdict\treject\n",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			cSource := fixture.cSource
			if test.mutate != nil {
				cSource = test.mutate(t, cSource)
			}
			artifact := fixture.build(t, cSource)
			input := filepath.Join(t.TempDir(), "input.bin")
			if err := os.WriteFile(input, []byte{0x41}, 0o600); err != nil {
				t.Fatal(err)
			}
			receiptPath := filepath.Join(t.TempDir(), "native-lifecycle.tsv")
			capturePath := filepath.Join(t.TempDir(), "compiler-events.json")
			code, stdout, stderr := runPhase23ObserverArtifact(t, artifact, input, receiptPath, capturePath)
			if code != test.wantCode || stdout != test.wantStdout || stderr != "" {
				t.Fatalf("mutated artifact code=%d stdout=%q stderr=%q; want %d/%q/empty stderr", code, stdout, stderr, test.wantCode, test.wantStdout)
			}
			receipt, err := os.ReadFile(receiptPath)
			if err != nil {
				t.Fatalf("read independent native observer receipt: %v", err)
			}
			if string(receipt) != test.wantReceipt {
				t.Fatalf("native receipt=%q, want reached mutation receipt %q", receipt, test.wantReceipt)
			}
			captureBytes, err := os.ReadFile(capturePath)
			if err != nil {
				t.Fatalf("read compiler event capture: %v", err)
			}
			capture, err := decodeApplicationCapture(captureBytes, execution.MaxApplicationEvidenceBytes)
			if err != nil || capture.Status != EvidenceStatusComplete || capture.Execution.Schema != execution.Schema2 ||
				capture.Execution.Outcome.Kind != execution.OutcomeReturned || capture.Execution.Outcome.Value != "65" || len(capture.Execution.LiveResources) != 0 {
				t.Fatalf("compiler capture=%+v decode error=%v raw=%s; want plausible returned semantic record independent of physical verdict", capture, err, captureBytes)
			}
			for _, event := range capture.Execution.Events {
				if strings.Contains(event.Kind, "release") || strings.Contains(event.Kind, "free") {
					t.Fatalf("semantic capture makes an unsupported physical destructor claim: %+v", event)
				}
			}
		})
	}
}

func newPhase23ObserverFixture(t *testing.T) *phase23ObserverFixture {
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
		"void *phase23_observer_malloc(size_t);\n" +
		"void phase23_observer_free(void *);\n" +
		"unsigned char phase23_observer_load_byte(const unsigned char *);\n\n" +
		"#define LANG_FILE_BYTE_MALLOC phase23_observer_malloc\n" +
		"#define LANG_FILE_BYTE_FREE phase23_observer_free\n" +
		"#define LANG_FILE_BYTE_LOAD_BYTE(pointer) phase23_observer_load_byte(pointer)\n"
	adapterText := strings.Replace(string(adapter), include, observerHooks, 1)
	if adapterText == string(adapter) {
		t.Fatal("adapter observer hooks were not inserted at the adapter include boundary")
	}
	if err := os.WriteFile(filepath.Join(root, "adapter.c"), []byte(adapterText), 0o600); err != nil {
		t.Fatal(err)
	}
	observer, err := os.ReadFile(filepath.Join(project, "internal", "compiler", "native", "testdata", "phase23_observer.c"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "phase23_observer.c"), observer, 0o600); err != nil {
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
	manifest.Sources = append(manifest.Sources, "phase23_observer.c")
	manifestBytes, err = json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, manifestBytes, 0o600); err != nil {
		t.Fatal(err)
	}

	source, err := os.ReadFile(filepath.Join(project, "examples", "phase23", "file_byte.lang"))
	if err != nil {
		t.Fatal(err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("Phase 23 source parse diagnostics: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("Phase 23 source check diagnostics: %+v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("Phase 23 core validation problems: %+v", validated.Problems)
	}
	cSource, err := cgen.EmitApplication(validated.Program())
	if err != nil {
		t.Fatalf("emit Phase 23 native app: %v", err)
	}
	if strings.Count(cSource, phase23UseErrorDiagnostic) == 1 {
		cSource = strings.Replace(cSource, phase23UseErrorDiagnostic,
			"phase23_observer_report_boundary(); "+phase23UseErrorDiagnostic, 1)
		cSource = "extern void phase23_observer_report_boundary(void);\n" + cSource
	}
	return &phase23ObserverFixture{root: root, manifest: manifestPath, source: source, cSource: cSource}
}

func (fixture *phase23ObserverFixture) build(t *testing.T, cSource string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	artifact := filepath.Join(t.TempDir(), "file-byte")
	if _, err := BuildApplication(ctx, fixture.source, cSource, artifact, fixture.manifest); err != nil {
		t.Fatalf("build observed Phase 23 application: %v", err)
	}
	return artifact
}

func runPhase23ObserverApp(t *testing.T, cli, artifact, input, observerPath string) (int, string, string) {
	t.Helper()
	arguments := []string{"app", "run", artifact}
	arguments = append(arguments, "--", input)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, cli, arguments...)
	command.Env = append(os.Environ(), "LANG_PHASE23_OBSERVER_PATH="+observerPath)
	var stdout, stderr strings.Builder
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	if ctx.Err() != nil {
		t.Fatalf("lang app run timed out: %v", ctx.Err())
	}
	if err == nil {
		return 0, stdout.String(), stderr.String()
	}
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) {
		t.Fatalf("lang app run could not launch: %v", err)
	}
	return exitError.ExitCode(), stdout.String(), stderr.String()
}

func runPhase23ObserverArtifact(t *testing.T, artifact, input, observerPath, capturePath string) (int, string, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, artifact, input)
	command.Env = append(os.Environ(), "LANG_PHASE23_OBSERVER_PATH="+observerPath, "LANG_APP_EVIDENCE_PATH="+capturePath)
	var stdout, stderr strings.Builder
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	if ctx.Err() != nil {
		t.Fatalf("mutated native artifact timed out: %v", ctx.Err())
	}
	if err == nil {
		return 0, stdout.String(), stderr.String()
	}
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) {
		t.Fatalf("mutated native artifact could not launch: %v", err)
	}
	return exitError.ExitCode(), stdout.String(), stderr.String()
}

func phase23MutateDestructor(t *testing.T, cSource, mutation string) string {
	t.Helper()
	useCall := "lang_file_byte_use("
	useAt := strings.Index(cSource, useCall)
	if useAt < 0 {
		t.Fatal("generated C has no Phase 23 borrowed use call")
	}
	useLineStart := strings.LastIndex(cSource[:useAt], "\n") + 1
	useLineEnd := strings.Index(cSource[useAt:], "\n") + useAt
	if useLineEnd < useAt {
		t.Fatal("generated borrowed-use call has no terminating line")
	}
	useLine := cSource[useLineStart:useLineEnd]
	closeParen := strings.Index(cSource[useAt+len(useCall):], ");")
	if closeParen < 0 {
		t.Fatal("generated borrowed-use call has no closing delimiter")
	}
	owner := strings.TrimSpace(cSource[useAt+len(useCall) : useAt+len(useCall)+closeParen])
	releaseCall := "lang_file_byte_release(" + owner + ");"
	releaseAt := strings.Index(cSource[useLineEnd:], releaseCall)
	if releaseAt < 0 {
		t.Fatal("generated C has no matching release after borrowed use")
	}
	releaseAt += useLineEnd
	releaseLineStart := strings.LastIndex(cSource[:releaseAt], "\n") + 1
	releaseLineEnd := strings.Index(cSource[releaseAt:], "\n") + releaseAt
	if releaseLineEnd < releaseAt {
		t.Fatal("generated release call has no terminating line")
	}
	releaseLine := cSource[releaseLineStart:releaseLineEnd]
	switch mutation {
	case "omit":
		return cSource[:releaseLineStart] + cSource[releaseLineEnd+1:]
	case "premature":
		block := useLine + "\n" + releaseLine
		if !strings.Contains(cSource, block) {
			t.Fatal("generated use and release are not adjacent for premature-release control")
		}
		return strings.Replace(cSource, block, releaseLine+"\n"+useLine, 1)
	case "duplicate":
		return cSource[:releaseLineStart] + releaseLine + "\n" + releaseLine + cSource[releaseLineEnd:]
	case "wrong-resource":
		wrong := "  lang_file_byte_release((lang_file_byte_owner){(unsigned char *)\"phase23-decoy\", UINT64_C(1)});"
		return cSource[:releaseLineStart] + releaseLine + "\n" + wrong + cSource[releaseLineEnd:]
	default:
		t.Fatalf("unknown Phase 23 release mutation %q", mutation)
		return cSource
	}
}
