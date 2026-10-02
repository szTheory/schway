package native_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	nativecompiler "github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

type phase25UtilityFixture struct {
	root      string
	source    []byte
	program   core.Program
	cSource   string
	shared    core.LinearOperation
	exclusive core.LinearOperation
}

const phase25ProcessOutputLimit = 1 << 20

type phase25BoundedOutput struct {
	buffer bytes.Buffer
	total  int
}

func (w *phase25BoundedOutput) Write(data []byte) (int, error) {
	w.total += len(data)
	remaining := phase25ProcessOutputLimit - w.buffer.Len()
	if remaining > 0 {
		if remaining > len(data) {
			remaining = len(data)
		}
		_, _ = w.buffer.Write(data[:remaining])
	}
	return len(data), nil
}

func (w *phase25BoundedOutput) String() string {
	if w.total > phase25ProcessOutputLimit {
		return w.buffer.String() + "\n[process output truncated]"
	}
	return w.buffer.String()
}

func TestPhase25UtilityNative(t *testing.T) {
	fixture := phase25LoadUtility(t)
	lane := os.Getenv("SCHWAY_PHASE25_LANE")
	if lane == "" {
		lane = "baseline"
	}
	flags, err := phase25LaneFlags(lane)
	if err != nil {
		t.Fatal(err)
	}
	instrumented := phase25InstrumentUtility(t, fixture.cSource, fixture.shared.ID, fixture.exclusive.ID)
	binary := phase25CompileUtility(t, fixture.root, fixture.source, instrumented, flags, lane)

	for _, test := range []struct {
		name       string
		value      byte
		wantExit   int
		wantStdout string
		wantStderr string
	}{
		{name: "0x41", value: 0x41, wantStdout: "65\n", wantStderr: "phase25:shared\nphase25:exclusive\n"},
		{name: "0x42", value: 0x42, wantStdout: "66\n", wantStderr: "phase25:shared\nphase25:exclusive\n"},
		{name: "0x43 typed failure precedes helpers", value: 0x43, wantExit: 65, wantStderr: "UseError.UnsupportedByte\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			code, stdout, stderr := phase25RunUtility(t, binary, lane, test.value)
			if code != test.wantExit || stdout != test.wantStdout || stderr != test.wantStderr {
				t.Fatalf("native utility lane=%s exit=%d stdout=%q stderr=%q; want exit=%d stdout=%q stderr=%q", lane, code, stdout, stderr, test.wantExit, test.wantStdout, test.wantStderr)
			}
			if test.value != 0x43 {
				t.Logf("phase25 native receipt family=foreign lane=%s host=%s/%s input=0x%02x expected=%q actual=%q status=pass", lane, runtime.GOOS, runtime.GOARCH, test.value, strings.TrimSpace(test.wantStdout), strings.TrimSpace(stdout))
				t.Logf("phase25 native receipt family=shared lane=%s host=%s/%s input=0x%02x expected=%q actual=%q status=pass", lane, runtime.GOOS, runtime.GOARCH, test.value, strings.TrimSpace(test.wantStdout), strings.TrimSpace(stdout))
				t.Logf("phase25 native receipt family=exclusive lane=%s host=%s/%s input=0x%02x expected=%q actual=%q status=pass", lane, runtime.GOOS, runtime.GOARCH, test.value, strings.TrimSpace(test.wantStdout), strings.TrimSpace(stdout))
			} else {
				t.Logf("phase25 native receipt family=utility-error lane=%s host=%s/%s input=0x43 expected=UseError.UnsupportedByte actual=%s helper_calls=0 status=pass", lane, runtime.GOOS, runtime.GOARCH, strings.TrimSpace(stderr))
			}
		})
	}
}

func TestPhase25NativeFamilyWrongResult(t *testing.T) {
	fixture := phase25LoadUtility(t)
	for _, test := range []struct {
		name      string
		operation core.LinearOperation
		wrong     string
		want      string
	}{
		{name: "shared helper", operation: fixture.shared, wrong: "99", want: "99\n"},
		{name: "exclusive helper", operation: fixture.exclusive, wrong: "98", want: "98\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			mutated := phase25MutateUtilityOperation(t, fixture.cSource, test.operation.ID, test.wrong)
			flags, err := phase25LaneFlags("baseline")
			if err != nil {
				t.Fatal(err)
			}
			binary := phase25CompileUtility(t, fixture.root, fixture.source, mutated, flags, "baseline")
			code, stdout, stderr := phase25RunUtility(t, binary, "baseline", 0x41)
			if code != 0 || stdout != test.want || stderr != "" {
				t.Fatalf("reached wrong-result control exit=%d stdout=%q stderr=%q; want 0/%q/empty", code, stdout, stderr, test.want)
			}
			t.Logf("phase25 reached wrong-result control family=%s expected=65 actual=%s status=pass", test.name, strings.TrimSpace(stdout))
		})
	}
}

func TestPhase25UtilityFamilyControls(t *testing.T) {
	for _, test := range []struct {
		name, source, wantCode string
	}{
		{"shared conflict", "module phase25.utility_shared_conflict\nexport { fn relay }\nfn relay(value: Buffer) -> Buffer {\n let shared = borrow value\n let exclusive = borrow mut value\n let observed = take shared\n exclusive\n}\n", "ownership.borrow_conflict"},
		{"exclusive conflict", "module phase25.utility_exclusive_conflict\nexport { fn relay }\nfn relay(value: Buffer) -> Buffer {\n let exclusive = borrow mut value\n let shared = borrow value\n let observed = take exclusive\n shared\n}\n", "ownership.borrow_conflict"},
		{"shared escape", "module phase25.utility_shared_escape\nexport { fn relay }\nfn relay(value: Buffer) -> Buffer {\n let shared = borrow value\n shared\n}\n", "core.origin_omitted"},
		{"exclusive escape", string(phase25ReadFixture(t, "exclusive_escape_reject.schway")), "core.origin_omitted"},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "control.schway")
			if err := os.WriteFile(path, []byte(test.source), 0o600); err != nil {
				t.Fatal(err)
			}
			checked, err := session.CheckCommandFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if checked.Status != "invalid" || len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != test.wantCode {
				t.Fatalf("control status=%s diagnostics=%+v; want invalid with one %s diagnostic", checked.Status, checked.Diagnostics, test.wantCode)
			}
			if strings.Contains(test.name, "conflict") && checked.Diagnostics[0].Primary.Start == checked.Diagnostics[0].Primary.End {
				t.Fatalf("control lost source attribution: %+v", checked.Diagnostics[0])
			}
		})
	}
}

func TestPhase25UtilityPointerManifests(t *testing.T) {
	for _, test := range []struct{ fixture, access, cType string }{
		{"shared_copy_accept.schway", "shared", "const uint64_t *"},
		{"exclusive_copy_accept.schway", "exclusive", "uint64_t *"},
	} {
		t.Run(test.access, func(t *testing.T) {
			checked := session.Check(phase25ReadFixture(t, test.fixture))
			if len(checked.Diagnostics) != 0 {
				t.Fatalf("%s family fixture check diagnostics: %+v", test.access, checked.Diagnostics)
			}
			phase25ManifestTypes(t, checked.Program, test.access, test.cType)
		})
	}
}

func TestPhase25EvidenceScript(t *testing.T) {
	script := string(phase25ReadRepoFile(t, "scripts", "verify-phase25.sh"))
	for _, want := range []string{"foreign", "shared", "exclusive", "baseline", "optimized", "sanitizer", "-O0", "-O2", "-fsanitize=address,undefined", "-fno-sanitize-recover=all", "Darwin", "Linux", "status=incomplete", "status=pass", "cold", "warm"} {
		if !strings.Contains(script, want) {
			t.Errorf("Phase 25 evidence script omits %q", want)
		}
	}
	if strings.Contains(script, "go test ./...") || strings.Contains(script, "go test -race") {
		t.Fatal("Phase 25 evidence script duplicates the existing full-suite CI owner")
	}
}

func TestPhase25EvidenceIndex(t *testing.T) {
	readmePath := filepath.Join(testsupport.ProjectPath(), "examples", "phase24", "README.md")
	readme := string(phase25ReadRepoFile(t, "examples", "phase24", "README.md"))
	for _, want := range []string{"transfer.schway", "file_byte.bindings.json", "0x41", "65", "0x42", "66", "0x43", "UseError.UnsupportedByte", "scripts/verify-phase25.sh", "shared", "exclusive", "Linux", "macOS", "ASan+UBSan"} {
		if !strings.Contains(readme, want) {
			t.Errorf("Phase 24/25 README evidence index omits %q", want)
		}
	}
	const observerLink = "[native observer](../../internal/compiler/native/phase24_observer_test.go)"
	const validationLink = "[hosted receipt for run 36856048690](../../.planning/phases/24-ownership-transfer-through-calls-and-errors/24-VALIDATION.md)"
	for _, want := range []string{observerLink, validationLink} {
		if !strings.Contains(readme, want) {
			t.Errorf("Phase 24 transfer and cleanup evidence omits required link %q", want)
		}
	}
	for _, destination := range []string{
		"../../internal/compiler/native/phase24_observer_test.go",
		"../../.planning/phases/24-ownership-transfer-through-calls-and-errors/24-VALIDATION.md",
	} {
		resolved := filepath.Clean(filepath.Join(filepath.Dir(readmePath), filepath.FromSlash(destination)))
		info, err := os.Stat(resolved)
		if err != nil {
			t.Errorf("Phase 24 evidence link %q does not resolve: %v", destination, err)
			continue
		}
		if !info.Mode().IsRegular() {
			t.Errorf("Phase 24 evidence link %q resolves to a non-file target", destination)
		}
	}
	validationPath := filepath.Clean(filepath.Join(filepath.Dir(readmePath), "../../.planning/phases/24-ownership-transfer-through-calls-and-errors/24-VALIDATION.md"))
	validation := string(phase25ReadRepoFile(t, ".planning", "phases", "24-ownership-transfer-through-calls-and-errors", "24-VALIDATION.md"))
	if !strings.Contains(validation, "36856048690") {
		t.Errorf("Phase 24 validation record %q omits hosted run 36856048690", validationPath)
	}
	if err := phase25CheckEvidenceIndexLinks(readme, filepath.Dir(readmePath)); err != nil {
		t.Fatalf("Phase 24/25 README evidence index links: %v", err)
	}
	broken := strings.Replace(readme, "## Evidence index", "## Evidence index\n\n- [missing control](missing-evidence-control.md)", 1)
	if err := phase25CheckEvidenceIndexLinks(broken, filepath.Dir(readmePath)); err == nil || !strings.Contains(err.Error(), "missing-evidence-control.md") {
		t.Fatalf("broken-link negative control error = %v; want diagnostic naming missing-evidence-control.md", err)
	}
}

var phase25EvidenceLinkPattern = regexp.MustCompile(`\[[^\]]*\]\(([^)\s]+)(?:\s+"[^"]*")?\)`)
var phase25MarkdownHeadingPattern = regexp.MustCompile(`^#{1,6}\s+`)

func phase25CheckEvidenceIndexLinks(readme, readmeDir string) error {
	const heading = "## Evidence index"
	var section strings.Builder
	inSection := false
	foundHeading := false
	for _, line := range strings.Split(readme, "\n") {
		trimmed := strings.TrimSpace(line)
		if !foundHeading {
			if trimmed == heading {
				foundHeading = true
				inSection = true
			}
			continue
		}
		if phase25MarkdownHeadingPattern.MatchString(trimmed) {
			break
		}
		if inSection {
			section.WriteString(line)
			section.WriteByte('\n')
		}
	}
	if !foundHeading {
		return fmt.Errorf("missing %q heading", heading)
	}

	links := phase25EvidenceLinkPattern.FindAllStringSubmatch(section.String(), -1)
	localCount := 0
	for _, match := range links {
		destination := strings.Trim(match[1], "<>")
		parsed, err := url.Parse(destination)
		if err != nil {
			return fmt.Errorf("invalid link destination %q: %w", destination, err)
		}
		if parsed.Scheme != "" || strings.HasPrefix(destination, "//") || destination == "" || strings.HasPrefix(destination, "#") {
			continue
		}
		localCount++
		filePath := strings.SplitN(destination, "#", 2)[0]
		if filePath == "" {
			continue
		}
		resolved := filepath.Clean(filepath.Join(readmeDir, filepath.FromSlash(filePath)))
		info, err := os.Stat(resolved)
		if err != nil {
			return fmt.Errorf("local link %q does not resolve to a file: %w", destination, err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("local link %q resolves to a non-file target", destination)
		}
	}
	if localCount == 0 {
		return fmt.Errorf("%q contains no local file links", heading)
	}
	return nil
}

func TestPhase25LivingRoadmap(t *testing.T) {
	product := string(phase25ReadRepoFile(t, ".planning", "PRODUCT-ROADMAP.md"))
	maturity := string(phase25ReadRepoFile(t, ".planning", "LANGUAGE-MATURITY.md"))
	for file, text := range map[string]string{"PRODUCT-ROADMAP.md": product, "LANGUAGE-MATURITY.md": maturity} {
		for _, want := range []string{"Phase 25", "Source inspection", "Newly executed", "Historical receipts", "hosted", "pointer", "Phase 26", "FizzBuzz", "JSON", "smallest complete slice", "checker", "evidence", "owner", "reprioritize"} {
			if !strings.Contains(text, want) {
				t.Errorf("%s does not distinguish/include %q", file, want)
			}
		}
	}
	if strings.Contains(product, "Phase 25 passed on Linux and macOS") || strings.Contains(maturity, "Phase 25 passed on Linux and macOS") {
		t.Fatal("living documents claim hosted Phase 25 evidence without a bound receipt")
	}
}

func phase25LoadUtility(t *testing.T) phase25UtilityFixture {
	t.Helper()
	root := testsupport.ProjectPath()
	sourcePath := filepath.Join(root, "examples", "phase24", "transfer.schway")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("integrated utility parse diagnostics: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("integrated utility source diagnostics: %+v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("integrated utility independent validation problems: %+v", validated.Problems)
	}
	program := validated.Program()
	cSource, err := cgen.EmitApplication(program)
	if err != nil {
		t.Fatalf("emit integrated utility: %v", err)
	}
	var shared, exclusive *core.Function
	for index := range program.Functions {
		function := &program.Functions[index]
		switch function.Name {
		case "shared_copy":
			shared = function
		case "exclusive_copy":
			exclusive = function
		}
	}
	if shared == nil || exclusive == nil {
		t.Fatalf("integrated utility is missing separate shared_copy and exclusive_copy functions; found %q", phase25FunctionNames(program))
	}
	sharedOperation, ok := phase25FirstOperation(shared, core.OpBorrowShared)
	if !ok {
		t.Fatal("shared_copy has no checked shared-borrow operation")
	}
	exclusiveOperation, ok := phase25FirstOperation(exclusive, core.OpCopy)
	if !ok {
		t.Fatal("exclusive_copy has no checked copied-value operation")
	}
	for _, want := range []string{"const uint64_t *", "uint64_t *", "SCHWAY_SHARED_COPY(&", "SCHWAY_EXCLUSIVE_COPY(&"} {
		if !strings.Contains(cSource, want) {
			t.Errorf("integrated emitted C omits checked family representation %q", want)
		}
	}
	for _, line := range strings.Split(cSource, "\n") {
		if !strings.Contains(line, "uint64_t *") {
			continue
		}
		for _, forbidden := range []string{"restrict", "noalias", "capture", "align("} {
			if strings.Contains(line, forbidden) {
				t.Fatalf("integrated pointer declaration contains unsupported promise %q: %s", forbidden, line)
			}
		}
	}
	return phase25UtilityFixture{root: root, source: source, program: program, cSource: cSource, shared: sharedOperation, exclusive: exclusiveOperation}
}

func phase25FunctionNames(program core.Program) string {
	names := make([]string, 0, len(program.Functions))
	for _, function := range program.Functions {
		names = append(names, function.Name)
	}
	return strings.Join(names, ", ")
}

func phase25FirstOperation(function *core.Function, kind core.OperationKind) (core.LinearOperation, bool) {
	for _, operation := range function.Linear.Operations {
		if operation.Kind == kind {
			return operation, true
		}
	}
	return core.LinearOperation{}, false
}

func phase25InstrumentUtility(t *testing.T, cSource, sharedID, exclusiveID string) string {
	t.Helper()
	markers := []struct{ id, marker string }{{sharedID, "phase25:shared"}, {exclusiveID, "phase25:exclusive"}}
	for _, item := range markers {
		lineIndex := strings.Index(cSource, item.id)
		if lineIndex < 0 {
			t.Fatalf("generated application omits checked pointer operation %q", item.id)
		}
		lineStart := strings.LastIndex(cSource[:lineIndex], "\n") + 1
		insertion := "  (void)fprintf(stderr, \"" + item.marker + "\\n\");\n"
		cSource = cSource[:lineStart] + insertion + cSource[lineStart:]
	}
	return cSource
}

func phase25MutateUtilityOperation(t *testing.T, cSource, operationID, value string) string {
	t.Helper()
	lines := strings.Split(cSource, "\n")
	for index, line := range lines {
		comment := strings.Index(line, "/*")
		if comment < 0 || !strings.Contains(line[comment:], operationID) {
			continue
		}
		assignment := strings.Index(line[:comment], " = ")
		if assignment < 0 {
			t.Fatalf("operation %q has no native scalar assignment: %q", operationID, line)
		}
		original := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(line[assignment+3:comment]), ";"))
		if original == "" {
			t.Fatalf("operation %q has an empty native scalar source: %q", operationID, line)
		}
		lines[index] = line[:assignment+3] + "(UINT64_C(" + value + ") + (" + original + " - " + original + ")); /* " + operationID + " reached wrong-result control */"
		return strings.Join(lines, "\n")
	}
	t.Fatalf("generated application has no operation marker %q", operationID)
	return cSource
}

func phase25LaneFlags(lane string) ([]string, error) {
	switch lane {
	case "baseline":
		return []string{"-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic", "-O0"}, nil
	case "optimized":
		return []string{"-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic", "-O2"}, nil
	case "sanitizer":
		return append([]string(nil), nativecompiler.SanitizerCompileFlags...), nil
	default:
		return nil, fmt.Errorf("unknown Phase 25 lane %q", lane)
	}
}

func phase25CompileUtility(t *testing.T, root string, source []byte, cSource string, flags []string, lane string) string {
	t.Helper()
	clang, err := exec.LookPath("clang")
	if err != nil {
		t.Fatalf("Phase 25 host evidence requires clang: %v", err)
	}
	directory := t.TempDir()
	cPath := filepath.Join(directory, "program.c")
	binary := filepath.Join(directory, "program")
	manifestPath := filepath.Join(root, "examples", "phase23", "file_byte.bindings.json")
	if lane == "baseline" {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		if _, err := nativecompiler.BuildApplication(ctx, source, cSource, binary, manifestPath); err != nil {
			t.Fatalf("build Phase 25 baseline utility with explicit bindings: %v", err)
		}
		return binary
	}
	if err := os.WriteFile(cPath, []byte(cSource), 0o600); err != nil {
		t.Fatal(err)
	}
	bindings, err := nativecompiler.ResolveBindings(manifestPath)
	if err != nil {
		t.Fatalf("resolve explicit Phase 23 C bindings: %v", err)
	}
	localDirectory := filepath.Join(directory, "local")
	if err := os.Mkdir(localDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	bindingRoot := filepath.Dir(manifestPath)
	var inputs []string
	for _, name := range append(append([]string(nil), bindings.Manifest.Headers...), bindings.Manifest.Sources...) {
		data, err := os.ReadFile(filepath.Join(bindingRoot, name))
		if err != nil {
			t.Fatalf("read declared Phase 23 binding input %q: %v", name, err)
		}
		destination := filepath.Join(localDirectory, name)
		if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(destination, data, 0o600); err != nil {
			t.Fatal(err)
		}
		if strings.HasSuffix(name, ".c") {
			inputs = append(inputs, destination)
		}
	}
	arguments := append([]string(nil), flags...)
	arguments = append(arguments, "-I", directory, "-I", localDirectory, cPath)
	arguments = append(arguments, inputs...)
	if lane == "sanitizer" {
		arguments = append(arguments, "-lc++")
	}
	arguments = append(arguments, "-o", binary)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, clang, arguments...)
	var stdout, stderr phase25BoundedOutput
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("compile Phase 25 %s utility with %q: %v\nstdout:\n%s\nstderr:\n%s", lane, arguments, err, stdout.String(), stderr.String())
	}
	return binary
}

func phase25RunUtility(t *testing.T, binary, lane string, input byte) (int, string, string) {
	t.Helper()
	inputPath := filepath.Join(t.TempDir(), "input.bin")
	if err := os.WriteFile(inputPath, []byte{input}, 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary, inputPath)
	command.Env = phase25LaneEnviron(lane)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	if err == nil {
		return 0, stdout.String(), stderr.String()
	}
	if exit, ok := err.(*exec.ExitError); ok {
		return exit.ExitCode(), stdout.String(), stderr.String()
	}
	t.Fatalf("run Phase 25 native utility: %v", err)
	return -1, stdout.String(), stderr.String()
}

func phase25LaneEnviron(lane string) []string {
	var environment []string
	for _, entry := range os.Environ() {
		if strings.HasPrefix(entry, "ASAN_OPTIONS=") || strings.HasPrefix(entry, "UBSAN_OPTIONS=") {
			continue
		}
		environment = append(environment, entry)
	}
	if lane == "sanitizer" {
		environment = append(environment, "ASAN_OPTIONS="+nativecompiler.ASanOptions, "UBSAN_OPTIONS="+nativecompiler.UBSanOptions)
	}
	return environment
}

func phase25ReadFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase25", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func phase25ReadRepoFile(t *testing.T, parts ...string) []byte {
	t.Helper()
	data, err := os.ReadFile(testsupport.ProjectPath(parts...))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func phase25ManifestTypes(t *testing.T, program core.Program, access, cType string) {
	t.Helper()
	manifest, err := cgen.EmitForeignManifest(program)
	if err != nil {
		t.Fatalf("emit %s family pointer manifest: %v", access, err)
	}
	var document struct {
		Attributes []json.RawMessage            `json:"emitted_attributes"`
		Pointers   []map[string]json.RawMessage `json:"emitted_pointer_parameters"`
	}
	if err := json.Unmarshal([]byte(manifest), &document); err != nil {
		t.Fatal(err)
	}
	if len(document.Attributes) != 0 || len(document.Pointers) != 1 || string(document.Pointers[0]["access"]) != `"`+access+`"` || string(document.Pointers[0]["c_type"]) != `"`+cType+`"` {
		t.Fatalf("%s family manifest disagrees with pointer ABI: %s", access, manifest)
	}
}
