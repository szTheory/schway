package session_test

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const phase37ValidationPath = ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-VALIDATION.md"

// This finite registry schema is separate from the archived direct-command
// grade law. Resolving a selector here supplies no execution or grade evidence.
var phase37ValidationTaskCommands = map[string]string{
	"37-01-01": "C01 C02 C25 C03 C26",
	"37-01-02": "C04 C26",
	"37-01-03": "C05 C06",
	"37-02-01": "C07",
	"37-02-02": "C08 C06",
	"37-03-01": "C09",
	"37-03-02": "C10 C11",
	"37-04-01": "C12",
	"37-04-02": "C13",
	"37-05-01": "C14",
	"37-05-02": "C15",
	"37-06-01": "C16",
	"37-06-02": "C17 C18",
	"37-07-01": "C19",
	"37-07-02": "C20",
	"37-07-03": "C21",
	"37-08-01": "C22",
	"37-08-02": "C23",
	"37-08-03": "C06 C24",
	"37-09-01": "C27 C03 C04 C15 C16 C17 C18 C26",
	"37-09-02": "C28 C06 C07 C08 C12 C13 C14 C22 C23 C24",
}

var phase37ValidationCommandPackages = map[string]string{
	"C01": "./cmd/schway", "C02": "./internal/compiler/session",
	"C03": "./internal/compiler/session", "C04": "./internal/compiler/session",
	"C07": "./internal/compiler/interp ./internal/compiler/session",
	"C08": "./internal/compiler/interp", "C09": "./cmd/schway",
	"C10": "./internal/compiler/session", "C11": "./cmd/schway",
	"C12": "./internal/compiler/interp", "C13": "./internal/compiler/native",
	"C14": "./internal/compiler/testsupport", "C15": "./cmd/schway",
	"C16": "./cmd/schway", "C17": "./cmd/schway",
	"C18": "./internal/compiler/session", "C19": "./cmd/schway",
	"C20": "./internal/compiler/session", "C21": "./internal/compiler/session",
	"C22": "./cmd/schway", "C23": "./cmd/schway", "C24": "./cmd/schway",
	"C25": "./internal/compiler/session", "C26": "./internal/compiler/session",
	"C27": "./internal/compiler/session", "C28": "./internal/compiler/interp",
}

var (
	phase37CommandIDPattern  = regexp.MustCompile(`\bC[0-9]+\b`)
	phase37DefinitionPattern = regexp.MustCompile("^- \\*\\*(C[0-9]+)\\*\\*: `([^`]+)`$")
	phase37ExactRootPattern  = regexp.MustCompile(`^Test[A-Za-z0-9_]+$`)
)

func phase37ValidationRegistryError(markdown string, index *testIndex) error {
	const primary = "## Per-Task Verification Map"
	const remediation = "## Remediation task map —2026-10-15"
	tasks, definitions, referenced := map[string]bool{}, map[string]string{}, map[string]bool{}
	taskReferences := map[string][]string{}
	headers := map[string]int{}
	section, width, commandColumn := "", 0, -1
	for _, line := range strings.Split(markdown, "\n") {
		if strings.HasPrefix(line, "##") {
			section, width, commandColumn = line, 0, -1
		}
		if strings.HasPrefix(line, "- **C") {
			match := phase37DefinitionPattern.FindStringSubmatch(line)
			if match == nil {
				return fmt.Errorf("malformed command definition: %s", line)
			}
			id := match[1]
			if _, duplicate := definitions[id]; duplicate {
				return fmt.Errorf("duplicate command definition %s", id)
			}
			definitions[id] = match[2]
		}
		if (section != primary && section != remediation) || !strings.HasPrefix(line, "|") {
			continue
		}
		cells := rawRowCells(line)
		if len(cells) == 0 {
			return fmt.Errorf("empty task row in %s", section)
		}
		if cells[0] == "Task ID" || cells[0] == "Task" {
			headers[section]++
			column := "Command IDs"
			if section == remediation {
				column = "Exact owners"
			}
			width, commandColumn = len(cells), slices.Index(cells, column)
			if commandColumn < 0 || headers[section] != 1 {
				return fmt.Errorf("task table schema in %s requires one %s header", section, column)
			}
			continue
		}
		if strings.Trim(cells[0], "-: ") == "" {
			continue
		}
		if width == 0 || len(cells) != width || commandColumn < 0 {
			return fmt.Errorf("malformed task row in %s: %s", section, line)
		}
		id := cells[0]
		_, known := phase37ValidationTaskCommands[id]
		if !known {
			return fmt.Errorf("unknown task %s", id)
		}
		if tasks[id] {
			return fmt.Errorf("duplicate task %s", id)
		}
		if strings.HasPrefix(id, "37-09-") != (section == remediation) {
			return fmt.Errorf("task %s is in the wrong task table", id)
		}
		tasks[id] = true
		gotRefs := phase37CommandIDPattern.FindAllString(cells[commandColumn], -1)
		seenRefs := map[string]bool{}
		for _, ref := range gotRefs {
			if seenRefs[ref] {
				return fmt.Errorf("task %s has duplicate command reference %s", id, ref)
			}
			seenRefs[ref], referenced[ref] = true, true
		}
		taskReferences[id] = gotRefs
	}
	// Iterate the independent pins, so deleting the document's last rows cannot
	// silently reduce the checker to the subset the document still declares.
	for _, id := range phase37SortedKeys(phase37ValidationTaskCommands) {
		if !tasks[id] {
			return fmt.Errorf("missing task %s", id)
		}
	}
	if len(tasks) != 21 || headers[primary] != 1 || headers[remediation] != 1 {
		return fmt.Errorf("task registry must contain both tables and exactly 21 tasks")
	}
	for _, id := range phase37SortedKeys(definitions) {
		_, goCommand := phase37ValidationCommandPackages[id]
		if !goCommand && id != "C05" && id != "C06" {
			return fmt.Errorf("unknown command definition %s", id)
		}
	}
	for n := 1; n <= 28; n++ {
		id := fmt.Sprintf("C%02d", n)
		if _, exists := definitions[id]; !exists {
			return fmt.Errorf("missing command definition %s", id)
		}
	}
	for _, ref := range phase37SortedKeys(referenced) {
		if _, exists := definitions[ref]; !exists {
			return fmt.Errorf("unknown command reference %s", ref)
		}
	}
	for _, id := range phase37SortedKeys(definitions) {
		if !referenced[id] {
			return fmt.Errorf("unreferenced command definition %s", id)
		}
	}
	for _, id := range phase37SortedKeys(taskReferences) {
		gotRefs, wantRefs := taskReferences[id], strings.Fields(phase37ValidationTaskCommands[id])
		slices.Sort(gotRefs)
		slices.Sort(wantRefs)
		if !slices.Equal(gotRefs, wantRefs) {
			return fmt.Errorf("task %s command references = %v, want %v", id, gotRefs, wantRefs)
		}
	}
	for _, id := range phase37SortedKeys(definitions) {
		if err := phase37ValidationCommandError(id, definitions[id], index); err != nil {
			return err
		}
	}
	return nil
}

func phase37SortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

func phase37ValidationCommandError(id, command string, index *testIndex) error {
	const proof = ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-prohibition-check.test.cjs"
	if id == "C05" || id == "C06" {
		want := "node --test " + proof
		if id == "C06" {
			want = "node " + proof + " --prove"
		}
		if command != want {
			return fmt.Errorf("%s must retain its exact Node exception", id)
		}
		return nil
	}
	command = strings.TrimPrefix(command, "GOCACHE=/tmp/ai-lang-verification-gocache ")
	parsed, ok := parseGoTestCommand(command)
	if !ok || parsed.Pattern == "" || len(parsed.Tags) != 0 {
		return fmt.Errorf("%s is not a focused Go selector command", id)
	}
	if strings.Join(parsed.Packages, " ") != phase37ValidationCommandPackages[id] {
		return fmt.Errorf("%s declared packages = %v, want %s", id, parsed.Packages, phase37ValidationCommandPackages[id])
	}
	want := "go test " + strings.Join(parsed.Packages, " ") + " -run '" + parsed.Pattern + "' -count=1 -v -timeout=600s"
	if command != want {
		return fmt.Errorf("%s requires the bounded exact Go command form", id)
	}
	pattern := parsed.Pattern
	if !strings.HasPrefix(pattern, "^") || !strings.HasSuffix(pattern, "$") {
		return fmt.Errorf("%s selector must be anchored exact test names", id)
	}
	branches := strings.TrimSuffix(strings.TrimPrefix(pattern, "^"), "$")
	if strings.HasPrefix(branches, "(") && strings.HasSuffix(branches, ")") {
		branches = branches[1 : len(branches)-1]
	}
	names := strings.Split(branches, "|")
	seenNames, resolvedNames := map[string]bool{}, map[string]bool{}
	for _, name := range names {
		if !phase37ExactRootPattern.MatchString(name) || seenNames[name] {
			return fmt.Errorf("%s has a non-exact or duplicate selector %s", id, name)
		}
		seenNames[name] = true
		owners := 0
		for _, pkg := range parsed.Packages {
			live, exists := index.resolvePackageNames(pkg)
			if !exists {
				return fmt.Errorf("%s package %s has no live tests", id, pkg)
			}
			if live[name] {
				owners++
			}
		}
		if owners != 1 {
			return fmt.Errorf("%s selector %s has %d declared live owners, want 1", id, name, owners)
		}
	}
	// Use the existing resolver for every declared package, including C07's
	// two-package union. Neither a missing package nor an unresolved branch may
	// disappear merely because another branch selects a real test.
	pairs := resolvedPkgPatterns(index, command)
	if len(pairs) != len(parsed.Packages) {
		return fmt.Errorf("%s does not select roots in every declared package", id)
	}
	for _, pair := range pairs {
		for _, name := range patternNames(pair.Pattern) {
			resolvedNames[name] = true
		}
	}
	if !slices.Equal(phase37SortedKeys(seenNames), phase37SortedKeys(resolvedNames)) {
		return fmt.Errorf("%s resolved roots differ from its exact selector names", id)
	}
	return nil
}

func phase37ValidationText(t testing.TB) string {
	t.Helper()
	data, err := os.ReadFile(testsupport.ProjectPath(phase37ValidationPath))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestPhase37ValidationCommandRegistry(t *testing.T) {
	if err := phase37ValidationRegistryError(phase37ValidationText(t), buildTestIndex(t)); err != nil {
		t.Fatal(err)
	}
	t.Log("all 21 task mappings and C01-C28 definitions/references resolve; static resolution supplies no RUN/PASS evidence")
}

func TestPhase37ValidationCommandRegistryControls(t *testing.T) {
	clean, index := phase37ValidationText(t), buildTestIndex(t)
	if err := phase37ValidationRegistryError(clean, index); err != nil {
		t.Fatalf("unarmed registry: %v", err)
	}
	removeLine := func(prefix string) string {
		lines, removed := strings.Split(clean, "\n"), 0
		out := make([]string, 0, len(lines))
		for _, line := range lines {
			if strings.HasPrefix(line, prefix) {
				removed++
				continue
			}
			out = append(out, line)
		}
		if removed != 1 {
			t.Fatalf("control prefix %q reached %d lines, want 1", prefix, removed)
		}
		return strings.Join(out, "\n")
	}
	replaceOnce := func(old, replacement string) string {
		if strings.Count(clean, old) != 1 {
			t.Fatalf("control anchor %q is not unique", old)
		}
		return strings.Replace(clean, old, replacement, 1)
	}
	controls := []struct{ name, text, want string }{
		{"deleted task", removeLine("| 37-05-02 |"), "missing task 37-05-02"},
		{"extra task", replaceOnce("| 37-05-02 |", "| 37-99-01 |"), "unknown task 37-99-01"},
		{"duplicate task", replaceOnce("| 37-05-02 |", "| 37-05-01 |"), "duplicate task 37-05-01"},
		{"deleted definition", removeLine("- **C14**:"), "missing command definition C14"},
		{"duplicate definition", clean + "\n- **C14**: `go test duplicate`\n", "duplicate command definition C14"},
		{"unknown definition", clean + "\n- **C99**: `go test unknown`\n", "unknown command definition C99"},
		{"unknown reference", replaceOnce("| C14 |", "| C99 |"), "unknown command reference C99"},
		{"duplicate reference", replaceOnce("| C14 |", "| C14, C14 |"), "duplicate command reference C14"},
		{"task mapping drift", replaceOnce("| C14 |", "| C13 |"), "task 37-05-01 command references"},
		{"unreferenced command", replaceOnce("| C05, C06 |", "| C06 |"), "unreferenced command definition C05"},
		{"nonexistent selected root", replaceOnce("'^TestPhase37CRC32SingleCopyAndIdentity$'", "'^TestPhase37NeverDeclared$'"), "C18 selector TestPhase37NeverDeclared has 0 declared live owners"},
		{"wrong package", replaceOnce("go test ./internal/compiler/session -run '^TestPhase37CRC32SingleCopyAndIdentity$'", "go test ./internal/compiler/interp -run '^TestPhase37CRC32SingleCopyAndIdentity$'"), "C18 declared packages"},
		{"partial selector resolution", replaceOnce("'^(TestPhase37CRC32OracleRows|TestPhase37CRC32OracleDriftControls)$'", "'^(TestPhase37CRC32OracleRows|TestPhase37NeverDeclared)$'"), "C14 selector TestPhase37NeverDeclared has 0 declared live owners"},
		{"C07 omitted package", replaceOnce("go test ./internal/compiler/interp ./internal/compiler/session -run", "go test ./internal/compiler/interp -run"), "C07 declared packages"},
		{"C05 Node exception drift", replaceOnce("node --test .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-prohibition-check.test.cjs", "node --test another.test.cjs"), "C05 must retain its exact Node exception"},
		{"C06 Node exception drift", replaceOnce("node .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-prohibition-check.test.cjs --prove", "node another.test.cjs --prove"), "C06 must retain its exact Node exception"},
		{"omitted C27 task", removeLine("|37-09-01|"), "missing task 37-09-01"},
		{"omitted C28 task", removeLine("|37-09-02|"), "missing task 37-09-02"},
		{"omitted C27 definition", removeLine("- **C27**:"), "missing command definition C27"},
		{"omitted C28 definition", removeLine("- **C28**:"), "missing command definition C28"},
		{"omitted C27 reference", replaceOnce("|C27; affected", "|affected"), "unreferenced command definition C27"},
		{"omitted C28 reference", replaceOnce("|C28; affected", "|affected"), "unreferenced command definition C28"},
		{"archived header confusion", replaceOnce("| Command IDs |", "| Automated Command |"), "requires one Command IDs header"},
	}
	for _, control := range controls {
		t.Run(control.name, func(t *testing.T) {
			if control.text == clean {
				t.Fatal("control did not change the registry")
			}
			err := phase37ValidationRegistryError(control.text, index)
			if err == nil || !strings.Contains(err.Error(), control.want) {
				t.Fatalf("reached control error = %v, want %q", err, control.want)
			}
			t.Logf("reached expected registry refusal: %s", err)
		})
	}
	if err := phase37ValidationRegistryError(clean, index); err != nil {
		t.Fatalf("restored registry: %v", err)
	}
}
