// Command phase20-closure-timing runs the bounded Phase 20 cold/warm suite
// protocol and writes its evidence artifact. It has no non-standard imports.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	phaseDir     = ".planning/phases/20-nyquist-d-13-34-and-the-frontier-fixture"
	wallCap      = 30 * time.Minute
	maxOutput    = 16 << 20
	probeTimeout = 10 * time.Second
)

type machineFacts struct {
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	CPUModel     string `json:"cpu_model"`
	LogicalCores int    `json:"logical_cores"`
	GoVersion    string `json:"go_version"`
	ClangVersion string `json:"cschway_version"`
}

type sample struct {
	Label         string
	Command       []string
	Cache         string
	Duration      time.Duration
	RawDurationNS int64
	ExitCode      int
	TimedOut      bool
	Output        string
	Fingerprint   string
}

// boundedOutput captures stdout and stderr independently. The extra byte lets
// the caller report truncation without retaining unbounded child output.
type boundedOutput struct {
	bytes.Buffer
	total int
}

func (b *boundedOutput) Write(p []byte) (int, error) {
	b.total += len(p)
	left := maxOutput + 1 - b.Len()
	if left > 0 {
		if left > len(p) {
			left = len(p)
		}
		_, _ = b.Buffer.Write(p[:left])
	}
	return len(p), nil
}

func runBounded(parent context.Context, dir, name string, args ...string) ([]byte, []byte, error) {
	ctx, cancel := context.WithTimeout(parent, probeTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	var stdout, stderr boundedOutput
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		return stdout.Bytes(), stderr.Bytes(), ctx.Err()
	}
	if stdout.total > maxOutput || stderr.total > maxOutput {
		return stdout.Bytes(), stderr.Bytes(), fmt.Errorf("%s output exceeded %d bytes", name, maxOutput)
	}
	return stdout.Bytes(), stderr.Bytes(), err
}

func main() {
	started := time.Now()
	deadline := started.Add(wallCap)
	repo, err := os.Getwd()
	if err != nil {
		fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repo, phaseDir), 0o755); err != nil {
		fatal(err)
	}
	gocache := os.Getenv("GOCACHE")
	if gocache == "" {
		gocache = filepath.Join(os.TempDir(), "phase20-gocache")
	}
	if err := os.MkdirAll(gocache, 0o700); err != nil {
		writeFailure(err.Error())
		fatal(err)
	}

	facts, machineID := collectMachineFacts()
	head, _ := commandOutput("git", "rev-parse", "HEAD")
	fingerprint, fingerprintErr := sourceFingerprint(repo)
	if fingerprintErr != nil {
		writeFailure("source revision could not be fingerprinted: " + fingerprintErr.Error())
		fatal(fingerprintErr)
	}

	var all []sample
	prewarm := run(deadline, gocache, "go-build-cache-prewarm", "", "go", "test", "./...", "-run", "^$", "-count=1")
	all = append(all, prewarm)
	if prewarm.ExitCode != 0 || prewarm.TimedOut {
		writeReport(all, facts, machineID, strings.TrimSpace(head), fingerprint, started, "Go build cache prewarm failed or timed out.")
		os.Exit(1)
	}
	if current, checkErr := sourceFingerprint(repo); checkErr != nil || current != fingerprint {
		writeReport(all, facts, machineID, strings.TrimSpace(head), fingerprint, started, "Source inputs changed during Go cache prewarm; no timing pair was measured.")
		os.Exit(1)
	}

	full := make([]sample, 0, 6)
	full, err = runFullSuitePairs(deadline, gocache, func() bool { return changedInputs(repo, fingerprint) })
	if len(full) > 0 {
		all = append(all, full...)
	}
	if err != nil {
		writeReport(all, facts, machineID, strings.TrimSpace(head), fingerprint, started, err.Error())
		os.Exit(1)
	}

	closure := make([]sample, 0, 2)
	if len(full) == 6 && time.Now().Before(deadline) && !changedInputs(repo, fingerprint) {
		cacheRoot, mkErr := os.MkdirTemp(os.TempDir(), "phase20-closure-pair-")
		if mkErr == nil {
			args := []string{"test", "./internal/compiler/session", "-run", "^TestPhase5CorpusThreeEngineAgreement/enumerated-closure$", "-count=1", "-v"}
			cold := run(deadline, gocache, "closure-cold", cacheRoot, "go", args...)
			all, closure = append(all, cold), append(closure, cold)
			if !failSample(cold) && !changedInputs(repo, fingerprint) {
				warm := run(deadline, gocache, "closure-warm", cacheRoot, "go", args...)
				all, closure = append(all, warm), append(closure, warm)
			}
		}
	}

	blocker := evaluate(full, closure, machineID, facts)
	if changedInputs(repo, fingerprint) {
		blocker = "Source inputs changed during measurement; samples do not share one input revision."
	}
	writeReport(all, facts, machineID, strings.TrimSpace(head), fingerprint, started, blocker)
	if blocker != "" {
		fmt.Fprintln(os.Stderr, "Phase 20 timing blocked:", blocker)
		os.Exit(1)
	}
}

func run(deadline time.Time, gocache, label, cacheRoot string, name string, args ...string) sample {
	command := append([]string{name}, args...)
	result := sample{Label: label, Command: command, Cache: cacheRoot, ExitCode: -1}
	if time.Now().After(deadline) {
		result.TimedOut = true
		result.Output = "30-minute wall-clock cap reached before process start"
		return result
	}
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = setEnv(os.Environ(), "GOCACHE", gocache)
	if cacheRoot != "" {
		cmd.Env = setEnv(cmd.Env, "SCHWAY_PHASE5_CLOSURE_CACHE", cacheRoot)
	}
	var stdout, stderr boundedOutput
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	started := time.Now()
	err := cmd.Run()
	result.Duration = time.Since(started)
	result.RawDurationNS = result.Duration.Nanoseconds()
	result.Fingerprint = "measured by monotonic time.Since"
	if ctx.Err() != nil {
		result.TimedOut = true
	}
	if err == nil {
		result.ExitCode = 0
	} else {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			result.ExitCode = exitErr.ExitCode()
		}
	}
	result.Output = strings.TrimSpace(stdout.String() + "\n" + stderr.String())
	if stdout.total > maxOutput || stderr.total > maxOutput {
		result.Output += fmt.Sprintf("\n[output truncated after %d bytes per stream; observed stdout=%d stderr=%d bytes]", maxOutput, stdout.total, stderr.total)
	}
	return result
}

// runFullSuitePairs owns the exact three-pair sequence. The injected
// runner keeps ordering, cache-root pairing, early failure, and input-drift
// behavior testable without launching the repository suite.
func runFullSuitePairs(deadline time.Time, gocache string, inputsChanged func() bool) ([]sample, error) {
	return runFullSuitePairsWith(deadline, gocache, inputsChanged, run)
}

func runFullSuitePairsWith(deadline time.Time, gocache string, inputsChanged func() bool, runSample func(time.Time, string, string, string, string, ...string) sample) ([]sample, error) {
	var samples []sample
	for pair := 1; pair <= 3; pair++ {
		cacheRoot, err := os.MkdirTemp(os.TempDir(), fmt.Sprintf("phase20-full-pair-%d-", pair))
		if err != nil {
			return samples, err
		}
		defer os.RemoveAll(cacheRoot)
		cold := runSample(deadline, gocache, fmt.Sprintf("full-%d-cold", pair), cacheRoot, "go", "test", "./...", "-count=1")
		samples = append(samples, cold)
		if failSample(cold) {
			return samples, fmt.Errorf("%s full-suite run failed or timed out (exit=%d timeout=%v)", cold.Label, cold.ExitCode, cold.TimedOut)
		}
		if inputsChanged != nil && inputsChanged() {
			return samples, errors.New("source inputs changed during measurement; samples do not share one input revision")
		}
		warm := runSample(deadline, gocache, fmt.Sprintf("full-%d-warm", pair), cacheRoot, "go", "test", "./...", "-count=1")
		samples = append(samples, warm)
		if failSample(warm) {
			return samples, fmt.Errorf("%s full-suite run failed or timed out (exit=%d timeout=%v)", warm.Label, warm.ExitCode, warm.TimedOut)
		}
		if inputsChanged != nil && inputsChanged() {
			return samples, errors.New("source inputs changed during measurement; samples do not share one input revision")
		}
	}
	return samples, nil
}

func setEnv(env []string, key, value string) []string {
	out := make([]string, 0, len(env)+1)
	prefix := key + "="
	replaced := false
	for _, item := range env {
		if strings.HasPrefix(item, prefix) {
			if !replaced {
				out = append(out, prefix+value)
				replaced = true
			}
			continue
		}
		out = append(out, item)
	}
	if !replaced {
		out = append(out, prefix+value)
	}
	return out
}

func failSample(s sample) bool { return s.ExitCode != 0 || s.TimedOut }
func changedInputs(repo, before string) bool {
	now, err := sourceFingerprint(repo)
	return err != nil || now != before
}

func sourceFingerprint(repo string) (string, error) {
	var data bytes.Buffer
	for _, command := range [][]string{{"diff", "--binary", "HEAD"}, {"ls-files", "--others", "--exclude-standard", "--", "*.go", "scripts/*.go"}} {
		out, _, err := runBounded(context.Background(), repo, "git", command...)
		if err != nil {
			return "", err
		}
		data.Write(out)
	}
	paths, _, err := runBounded(context.Background(), repo, "git", "ls-files", "--others", "--exclude-standard", "-z", "--", "*.go", "scripts/*.go")
	if err != nil {
		return "", err
	}
	for _, path := range bytes.Split(paths, []byte{0}) {
		if len(path) == 0 {
			continue
		}
		content, readErr := os.ReadFile(filepath.Join(repo, string(path)))
		if readErr != nil {
			return "", readErr
		}
		data.Write(path)
		data.Write(content)
	}
	sum := sha256.Sum256(data.Bytes())
	return hex.EncodeToString(sum[:]), nil
}

func commandOutput(name string, args ...string) (string, error) {
	out, _, err := runBounded(context.Background(), "", name, args...)
	return string(out), err
}

func collectMachineFacts() (machineFacts, string) {
	clang, _ := commandOutput("clang", "--version")
	cpu, _ := commandOutput("sysctl", "-n", "machdep.cpu.brand_string")
	facts := machineFacts{OS: runtime.GOOS, Arch: runtime.GOARCH, CPUModel: strings.TrimSpace(string(bytes.TrimSpace([]byte(cpu)))), LogicalCores: runtime.NumCPU(), GoVersion: runtime.Version(), ClangVersion: strings.TrimSpace(strings.SplitN(clang, "\n", 2)[0])}
	encoded, _ := json.Marshal(facts)
	sum := sha256.Sum256(encoded)
	return facts, "machine:" + hex.EncodeToString(sum[:6])
}

func evaluate(full, closure []sample, machineID string, facts machineFacts) string {
	if len(full) != 6 {
		return fmt.Sprintf("incomplete full-suite protocol: got %d of 6 runs", len(full))
	}
	for _, s := range full {
		if failSample(s) {
			return fmt.Sprintf("%s full-suite run failed or timed out (exit=%d timeout=%v)", s.Label, s.ExitCode, s.TimedOut)
		}
	}
	if len(closure) != 2 {
		return fmt.Sprintf("enumerated-closure cold/warm pair incomplete: got %d runs", len(closure))
	}
	for _, s := range closure {
		if failSample(s) {
			return fmt.Sprintf("%s closure run failed or timed out", s.Label)
		}
	}
	warmStats := closureCacheCounts(closure[1].Output)
	if warmStats.reused == 0 {
		return "warm closure run reported no content-addressed artifact reuse"
	}
	if machineID != "machine:4797d76b7863" {
		return fmt.Sprintf("non-comparable Phase 14 machine baseline: current %s, baseline machine:4797d76b7863", machineID)
	}
	cold, warm := durationsByKind(full)
	if median(warm) >= median(cold) {
		return "warm median did not beat paired cold median"
	}
	if median(warm) >= 192700*time.Millisecond {
		return "warm median is not below the Phase 14 roadmap reference of 192.7 seconds"
	}
	if median(warm) >= 191890*time.Millisecond {
		return "warm median is not below the Phase 14 measured 191.89-second reference"
	}
	if facts.ClangVersion == "" {
		return "Clang toolchain identity unavailable"
	}
	return ""
}

type cacheCounts struct{ reused, recomputed, notCacheable, unavailable int }

func closureCacheCounts(output string) cacheCounts {
	var counts cacheCounts
	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, "phase20-closure-cache") {
			continue
		}
		for _, field := range strings.Fields(line) {
			key, value, ok := strings.Cut(field, "=")
			if !ok {
				continue
			}
			parsed, err := strconv.Atoi(value)
			if err != nil {
				continue
			}
			switch key {
			case "reused":
				counts.reused = parsed
			case "recomputed":
				counts.recomputed = parsed
			case "not_cacheable":
				counts.notCacheable = parsed
			case "unavailable":
				counts.unavailable = parsed
			}
		}
	}
	return counts
}

func durationsByKind(samples []sample) ([]time.Duration, []time.Duration) {
	var cold, warm []time.Duration
	for _, s := range samples {
		if strings.Contains(s.Label, "cold") {
			cold = append(cold, s.Duration)
		} else {
			warm = append(warm, s.Duration)
		}
	}
	sort.Slice(cold, func(i, j int) bool { return cold[i] < cold[j] })
	sort.Slice(warm, func(i, j int) bool { return warm[i] < warm[j] })
	return cold, warm
}
func median(values []time.Duration) time.Duration {
	if len(values) == 0 {
		return 0
	}
	return values[len(values)/2]
}
func seconds(values []time.Duration) []string {
	out := make([]string, len(values))
	for i, v := range values {
		out[i] = fmt.Sprintf("%.3f", v.Seconds())
	}
	return out
}
func minMax(values []time.Duration) (time.Duration, time.Duration) {
	if len(values) == 0 {
		return 0, 0
	}
	return values[0], values[len(values)-1]
}

func writeFailure(message string) {
	facts, id := collectMachineFacts()
	head, _ := commandOutput("git", "rev-parse", "HEAD")
	fp, _ := sourceFingerprint(".")
	writeReport(nil, facts, id, strings.TrimSpace(head), fp, time.Now(), message)
}
func writeReport(samples []sample, facts machineFacts, machineID, head, fingerprint string, started time.Time, blocker string) {
	completed := time.Now()
	var b strings.Builder
	b.WriteString("# Phase 20 closure timing evidence\n\n")
	fmt.Fprintf(&b, "- **Status:** %s\n- **Started:** %s\n- **Completed:** %s\n- **Elapsed:** %.3f s\n- **Machine:** `%s` (`%s`, `%s`, %d cores)\n- **Go:** `%s`\n- **Cschway:** `%s`\n- **Input HEAD at start:** `%s`\n- **Source fingerprint:** `%s`\n- **Go build cache:** `%s` (stable across samples)\n- **Protocol cap:** 30 minutes; monotonic `time.Since` measurements\n", map[bool]string{true: "PASS", false: "BLOCKED"}[blocker == ""], started.UTC().Format(time.RFC3339), completed.UTC().Format(time.RFC3339), completed.Sub(started).Seconds(), machineID, facts.OS, facts.Arch, facts.LogicalCores, facts.GoVersion, facts.ClangVersion, head, fingerprint, os.Getenv("GOCACHE"))
	b.WriteString("\nThree paired samples expose spread; they do not establish a high-confidence percentile. Full-suite commands use `go test ./... -count=1`; each pair has an empty cold closure cache and reuses that exact populated cache for warm. The Go build cache is prewarmed once and remains stable.\n\n")
	b.WriteString("| Sample | Command | Cache | Raw monotonic nanoseconds (seconds) | Exit | Timeout |\n|---|---|---|---:|---:|---|\n")
	for _, s := range samples {
		fmt.Fprintf(&b, "| %s | `%s` | `%s` | %d (%.3f) | %d | %v |\n", s.Label, strings.Join(s.Command, " "), s.Cache, s.RawDurationNS, s.Duration.Seconds(), s.ExitCode, s.TimedOut)
	}
	var full []sample
	for _, s := range samples {
		if strings.HasPrefix(s.Label, "full-") {
			full = append(full, s)
		}
	}
	var cold, warm []time.Duration
	for _, s := range full {
		if strings.HasSuffix(s.Label, "-cold") {
			cold = append(cold, s.Duration)
		} else {
			warm = append(warm, s.Duration)
		}
	}
	sort.Slice(cold, func(i, j int) bool { return cold[i] < cold[j] })
	sort.Slice(warm, func(i, j int) bool { return warm[i] < warm[j] })
	if len(cold) > 0 && len(warm) > 0 {
		cmin, cmax := minMax(cold)
		wmin, wmax := minMax(warm)
		fmt.Fprintf(&b, "\n- **Cold ordered seconds:** %s; min/median/max = %.3f / %.3f / %.3f\n- **Warm ordered seconds:** %s; min/median/max = %.3f / %.3f / %.3f\n- **Per-pair cold minus warm seconds:**", strings.Join(seconds(cold), ", "), cmin.Seconds(), median(cold).Seconds(), cmax.Seconds(), strings.Join(seconds(warm), ", "), wmin.Seconds(), median(warm).Seconds(), wmax.Seconds())
		for i := 0; i < len(full)/2; i++ {
			fmt.Fprintf(&b, " %.3f", (full[i*2].Duration - full[i*2+1].Duration).Seconds())
		}
		b.WriteByte('\n')
	}
	b.WriteString("\n## Enumerated closure cold/warm\n\n")
	for _, s := range samples {
		if strings.HasPrefix(s.Label, "closure-") {
			counts := closureCacheCounts(s.Output)
			fmt.Fprintf(&b, "- **%s:** %.3f s; exit=%d; reused=%d recomputed=%d not-cacheable=%d unavailable=%d. Command: `%s`.\n", s.Label, s.Duration.Seconds(), s.ExitCode, counts.reused, counts.recomputed, counts.notCacheable, counts.unavailable, strings.Join(s.Command, " "))
		}
	}
	b.WriteString("\n## Phase 14 comparison\n\n- Roadmap reference: **192.7 s**.\n- QLT-02 measured reference: **191.89 s**.\n")
	if blocker != "" {
		fmt.Fprintf(&b, "- **Non-passing blocker:** %s\n", blocker)
	} else {
		fmt.Fprintf(&b, "- Warm median: **%.3f s**, below both references and the paired cold median.\n", median(warm).Seconds())
	}
	b.WriteString("\n## Raw process output\n\n")
	for _, s := range samples {
		fmt.Fprintf(&b, "### %s\n\n```text\n%s\n```\n\n", s.Label, s.Output)
	}
	path := filepath.Join(phaseDir, "20-CLOSURE-TIMING.md")
	if err := os.WriteFile(path, []byte(b.String()), 0o600); err != nil {
		fmt.Fprintln(os.Stderr, "write timing report:", err)
	}
}

func fatal(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }

var _ io.Writer = (*boundedOutput)(nil)
