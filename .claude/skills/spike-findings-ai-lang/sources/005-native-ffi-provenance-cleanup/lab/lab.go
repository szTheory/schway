package lab

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Event struct {
	ID    string `json:"id"`
	Valid bool   `json:"valid"`
	A     uint64 `json:"a"`
	B     uint64 `json:"b"`
	Trace string `json:"trace"`
}

type Build struct {
	ID            string `json:"id"`
	Optimization  string `json:"optimization"`
	CompileMillis int64  `json:"compile_millis"`
	BinaryBytes   int64  `json:"binary_bytes"`
	Events        int    `json:"events"`
}

type Mutation struct {
	ID       string `json:"id"`
	Class    string `json:"class"`
	Detected bool   `json:"detected"`
	Evidence string `json:"evidence"`
}

type Report struct {
	Schema            int        `json:"schema"`
	Toolchain         string     `json:"toolchain"`
	Target            string     `json:"target"`
	SourceDigest      string     `json:"source_digest"`
	Builds            []Build    `json:"builds"`
	Mutations         []Mutation `json:"mutations"`
	DetectedMutations int        `json:"detected_mutations"`
	TotalMutations    int        `json:"total_mutations"`
	Verdict           string     `json:"verdict"`
}

var expected = []Event{
	{ID: "layout-roundtrip", Valid: true, A: 0x1122334455667788, B: 7, Trace: "tag:9"},
	{ID: "cleanup-fail-a", Valid: true, A: 1, B: 1, Trace: "acquire:a,release:a"},
	{ID: "cleanup-fail-b", Valid: true, A: 2, B: 2, Trace: "acquire:a,acquire:b,release:b,release:a"},
	{ID: "cleanup-success", Valid: true, A: 2, B: 2, Trace: "acquire:a,acquire:b,release:b,release:a"},
	{ID: "allocator-pairing", Valid: true, A: 1, B: 1, Trace: "mismatch-rejected,release"},
	{ID: "callback-retention", Valid: true, A: 1, B: 1, Trace: "live-read,stale-rejected"},
}

func Run(root string) (Report, error) {
	clang, err := exec.LookPath("clang")
	if err != nil {
		return Report{}, err
	}
	version, err := command(clang, "--version")
	if err != nil {
		return Report{}, err
	}
	target, _ := command(clang, "-dumpmachine")
	temporary, err := os.MkdirTemp("", "ai-lang-native-spike-")
	if err != nil {
		return Report{}, err
	}
	defer os.RemoveAll(temporary)
	report := Report{Schema: 1, Toolchain: firstLine(version), Target: strings.TrimSpace(target), Verdict: "PARTIAL"}
	report.SourceDigest, err = sourceDigest(root)
	if err != nil {
		return Report{}, err
	}
	for _, optimization := range []string{"-O0", "-O3"} {
		binary := filepath.Join(temporary, "baseline"+optimization)
		elapsed, err := compile(clang, root, binary, optimization, nil, "producer.c", "probe.c")
		if err != nil {
			return Report{}, err
		}
		events, _, runErr := runEvents(binary, nil)
		if runErr != nil || !equalEvents(events, expected) {
			return Report{}, fmt.Errorf("baseline %s disagreed: err=%v events=%+v", optimization, runErr, events)
		}
		info, _ := os.Stat(binary)
		report.Builds = append(report.Builds, Build{ID: "baseline", Optimization: optimization, CompileMillis: elapsed.Milliseconds(), BinaryBytes: info.Size(), Events: len(events)})
	}
	for _, item := range []struct{ id, class, define, event string }{
		{"layout-contract", "abi-layout", "INJECT_LAYOUT_BUG", "layout-roundtrip"},
		{"partial-cleanup", "cleanup", "INJECT_CLEANUP_LEAK", "cleanup-fail-b"},
		{"allocator-pairing", "allocator", "INJECT_ALLOCATOR_PAIRING_BUG", "allocator-pairing"},
		{"callback-retention", "provenance", "INJECT_RETENTION_BUG", "callback-retention"},
	} {
		binary := filepath.Join(temporary, item.id)
		_, compileErr := compile(clang, root, binary, "-O3", []string{item.define}, "producer.c", "probe.c")
		events, output, runErr := runEvents(binary, nil)
		detected := compileErr == nil && runErr != nil && invalidEvent(events, item.event)
		evidence := "invalid semantic event"
		if compileErr != nil {
			evidence = compileErr.Error()
		}
		if !detected && output != "" {
			evidence = strings.TrimSpace(output)
		}
		report.Mutations = append(report.Mutations, Mutation{ID: item.id, Class: item.class, Detected: detected, Evidence: evidence})
	}
	restrictMutation, err := restrictProbe(clang, root, temporary)
	if err != nil {
		return Report{}, err
	}
	report.Mutations = append(report.Mutations, restrictMutation)
	sanitizerMutation, err := sanitizerProbe(clang, root, temporary)
	if err != nil {
		return Report{}, err
	}
	report.Mutations = append(report.Mutations, sanitizerMutation)
	nonlocalMutation, err := nonlocalProbe(clang, root, temporary)
	if err != nil {
		return Report{}, err
	}
	report.Mutations = append(report.Mutations, nonlocalMutation)
	report.TotalMutations = len(report.Mutations)
	for _, mutation := range report.Mutations {
		if mutation.Detected {
			report.DetectedMutations++
		}
	}
	return report, nil
}

func compile(clang, root, binary, optimization string, defines []string, sources ...string) (time.Duration, error) {
	args := []string{"-std=c17", "-Wall", "-Wextra", "-Werror", optimization}
	for _, define := range defines {
		args = append(args, "-D"+define)
	}
	for _, source := range sources {
		args = append(args, filepath.Join(root, "native", source))
	}
	args = append(args, "-o", binary)
	started := time.Now()
	output, err := exec.Command(clang, args...).CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("clang %s: %w\n%s", strings.Join(args, " "), err, output)
	}
	return time.Since(started), nil
}

func runEvents(binary string, environment []string) ([]Event, string, error) {
	cmd := exec.Command(binary)
	cmd.Env = append(os.Environ(), environment...)
	output, err := cmd.CombinedOutput()
	var events []Event
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		var event Event
		if json.Unmarshal(scanner.Bytes(), &event) == nil && event.ID != "" {
			events = append(events, event)
		}
	}
	return events, string(output), err
}

func restrictProbe(clang, root, temporary string) (Mutation, error) {
	values := map[string]string{}
	for _, optimization := range []string{"-O0", "-O3"} {
		for _, unsafe := range []bool{false, true} {
			id := optimization + map[bool]string{false: "-safe", true: "-unsound"}[unsafe]
			binary := filepath.Join(temporary, "restrict"+strings.ReplaceAll(id, "-", ""))
			defines := []string{}
			if unsafe {
				defines = append(defines, "UNSOUND_NOALIAS")
			}
			if _, err := compile(clang, root, binary, optimization, defines, "restrict_alias.c"); err != nil {
				return Mutation{}, err
			}
			output, runErr := command(binary)
			values[id] = strings.TrimSpace(output)
			if !unsafe && runErr != nil {
				return Mutation{}, fmt.Errorf("alias-safe %s failed: %w", optimization, runErr)
			}
		}
	}
	detected := values["-O0-safe"] == "2" && values["-O3-safe"] == "2" && values["-O0-unsound"] == "2" && values["-O3-unsound"] == "1"
	return Mutation{ID: "unsound-noalias", Class: "optimizer-contract", Detected: detected, Evidence: fmt.Sprintf("safe O0/O3=%s/%s; restrict O0/O3=%s/%s", values["-O0-safe"], values["-O3-safe"], values["-O0-unsound"], values["-O3-unsound"])}, nil
}

func sanitizerProbe(clang, root, temporary string) (Mutation, error) {
	binary := filepath.Join(temporary, "unsafe-uaf")
	args := []string{"-std=c17", "-O1", "-g", "-fsanitize=address,undefined", "-fno-omit-frame-pointer", filepath.Join(root, "native", "unsafe_uaf.c"), "-o", binary}
	if output, err := exec.Command(clang, args...).CombinedOutput(); err != nil {
		return Mutation{}, fmt.Errorf("sanitizer compile: %w\n%s", err, output)
	}
	_, output, runErr := runEvents(binary, []string{"ASAN_OPTIONS=detect_leaks=0:abort_on_error=1"})
	detected := runErr != nil && strings.Contains(output, "AddressSanitizer: heap-use-after-free")
	return Mutation{ID: "retained-pointer-uaf", Class: "sanitizer-provenance", Detected: detected, Evidence: "AddressSanitizer heap-use-after-free"}, nil
}

func nonlocalProbe(clang, root, temporary string) (Mutation, error) {
	binary := filepath.Join(temporary, "nonlocal-exit")
	if _, err := compile(clang, root, binary, "-O3", nil, "nonlocal_exit.c"); err != nil {
		return Mutation{}, err
	}
	events, _, runErr := runEvents(binary, nil)
	detected := runErr != nil && invalidEvent(events, "nonlocal-exit") && len(events) == 1 && events[0].A == 1 && events[0].B == 0
	return Mutation{ID: "foreign-nonlocal-exit", Class: "unwind-cleanup", Detected: detected, Evidence: "longjmp skipped one release"}, nil
}

func sourceDigest(root string) (string, error) {
	files, err := filepath.Glob(filepath.Join(root, "native", "*.c"))
	if err != nil {
		return "", err
	}
	sort.Strings(files)
	hash := sha256.New()
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return "", err
		}
		hash.Write([]byte(filepath.Base(file)))
		hash.Write([]byte{0})
		hash.Write(data)
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
}

func equalEvents(left, right []Event) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func invalidEvent(events []Event, id string) bool {
	for _, event := range events {
		if event.ID == id {
			return !event.Valid
		}
	}
	return false
}

func command(name string, args ...string) (string, error) {
	output, err := exec.Command(name, args...).CombinedOutput()
	return string(output), err
}

func firstLine(value string) string {
	if index := strings.IndexByte(value, '\n'); index >= 0 {
		return value[:index]
	}
	return strings.TrimSpace(value)
}
