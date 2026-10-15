package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const (
	phase37MatrixChildDeadline = 30 * time.Second
	phase37MatrixOutputLimit   = 32 * 1024
)

type phase37CLITuple struct {
	Exit           int
	Stdout, Stderr string
}

type phase37BoundedBuffer struct {
	bytes.Buffer
	total int
}

func (b *phase37BoundedBuffer) Write(p []byte) (int, error) {
	b.total += len(p)
	remaining := phase37MatrixOutputLimit - b.Buffer.Len()
	if remaining > len(p) {
		remaining = len(p)
	}
	if remaining > 0 {
		_, _ = b.Buffer.Write(p[:remaining])
	}
	return len(p), nil
}

func (b *phase37BoundedBuffer) overflowed() bool {
	return b.total > phase37MatrixOutputLimit
}

// TestPhase37PublicFileMatrix runs every representable input row through the
// actual built schway CLI. Each engine is checked against its independent
// consumer tuple before interpreter/native parity is compared.
func TestPhase37PublicFileMatrix(t *testing.T) {
	cli := testsupport.BuildCLI(t)
	recipes, err := testsupport.CRC32OracleRows()
	if err != nil {
		t.Fatal(err)
	}
	recipeByID := make(map[string]testsupport.CRC32OracleRow, len(recipes))
	for _, recipe := range recipes {
		recipeByID[recipe.ID] = recipe
	}
	consumers, err := testsupport.CRC32Consumers()
	if err != nil {
		t.Fatal(err)
	}
	cases, err := testsupport.CRC32CaseMatrix()
	if err != nil {
		t.Fatal(err)
	}
	manifestControls, err := testsupport.CRC32ManifestControls()
	if err != nil {
		t.Fatal(err)
	}

	for _, row := range cases {
		row := row
		t.Run(row.ID, func(t *testing.T) {
			if row.Kind == "direct-argument" {
				phase37RunDirectNULControls(t, consumers, row)
				return
			}
			work := t.TempDir()
			path := testsupport.CRC32MaterializeCase(t, work, row, recipeByID)
			for _, consumer := range consumers {
				consumer := consumer
				t.Run(consumer.ID, func(t *testing.T) {
					gotByEngine := make(map[string]phase37CLITuple, 2)
					want, ok := row.Expected[consumer.ID]
					if !ok {
						t.Fatalf("fixture has no expected tuple for consumer %s", consumer.ID)
					}
					for _, engine := range []string{"interpreter", "native"} {
						engine := engine
						t.Run(engine, func(t *testing.T) {
							tmp := t.TempDir()
							args := phase37PublicFileArgs(engine, consumer, path)
							started := time.Now()
							got, err := phase37RunCLI(context.Background(), cli, args, work, tmp)
							elapsed := time.Since(started)
							if row.ID == "cap" && consumer.ID == "crc32" && engine == "interpreter" {
								t.Logf("maximum CRC-32 interpreter observation: %s for %d bytes", elapsed, testsupport.CRC32Limit)
							}
							if err != nil {
								t.Fatal(err)
							}
							if err := phase37CompareTuple(got, want); err != nil {
								t.Fatal(err)
							}
							if strings.Contains(got.Stdout+got.Stderr, "phase37-private") {
								t.Fatalf("private input path leaked into output: (%q,%q)", got.Stdout, got.Stderr)
							}
							phase37RequireTempClean(t, tmp)
							gotByEngine[engine] = got
						})
					}
					if gotByEngine["interpreter"] != gotByEngine["native"] {
						t.Fatalf("expected-first engine parity mismatch: interpreter=%+v native=%+v", gotByEngine["interpreter"], gotByEngine["native"])
					}
				})
			}
		})
	}

	t.Run("manifest-controls", func(t *testing.T) {
		phase37RunManifestControls(t, cli, consumers, manifestControls, recipeByID)
	})
	t.Run("isolated-bounded-parallel-pair", func(t *testing.T) {
		phase37RunParallelPair(t, cli, consumers, recipeByID)
	})
}

// TestPhase37PublicMatrixDriftControl arms the same comparator used by the
// matrix and proves a changed expected tuple cannot pass against a good result.
func TestPhase37PublicMatrixDriftControl(t *testing.T) {
	cases, err := testsupport.CRC32CaseMatrix()
	if err != nil {
		t.Fatal(err)
	}
	var check testsupport.CRC32Tuple
	for _, row := range cases {
		if row.ID == "check" {
			check = row.Expected["crc32"]
			break
		}
	}
	if check.Exit != 0 || check.Stdout != "3421780262\n" || check.Stderr != "" {
		t.Fatalf("fixed check tuple = %+v", check)
	}
	actual := phase37CLITuple{Exit: check.Exit, Stdout: check.Stdout, Stderr: check.Stderr}
	mutatedExpected := check
	mutatedExpected.Stdout = "3421780263\n"
	if err := phase37CompareTuple(actual, mutatedExpected); err == nil {
		t.Fatal("expected tuple drift was accepted by the public matrix comparator")
	}
}

func phase37PublicFileArgs(engine string, consumer testsupport.CRC32Consumer, path string) []string {
	args := []string{"run", "--engine=" + engine, testsupport.ProjectPath(filepath.FromSlash(consumer.Source))}
	if engine == "native" {
		args = append(args, "--manifest", testsupport.ProjectPath(filepath.FromSlash(consumer.Manifest)))
	}
	return append(args, "--", path)
}

func phase37RunCLI(parent context.Context, cli string, args []string, cwd, tmpDir string) (phase37CLITuple, error) {
	ctx, cancel := context.WithTimeout(parent, phase37MatrixChildDeadline)
	defer cancel()
	cmd := exec.CommandContext(ctx, cli, args...)
	cmd.Dir = cwd
	cmd.Env = phase37SetEnv(os.Environ(), "TMPDIR", tmpDir)
	stdout := &phase37BoundedBuffer{}
	stderr := &phase37BoundedBuffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		return phase37CLITuple{}, fmt.Errorf("public CLI exceeded %s: %w", phase37MatrixChildDeadline, ctx.Err())
	}
	if stdout.overflowed() || stderr.overflowed() {
		return phase37CLITuple{}, fmt.Errorf("public CLI output exceeded %d bytes", phase37MatrixOutputLimit)
	}
	code := 0
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			return phase37CLITuple{}, fmt.Errorf("run public CLI: %w", err)
		}
		code = exitErr.ExitCode()
	}
	return phase37CLITuple{Exit: code, Stdout: stdout.String(), Stderr: stderr.String()}, nil
}

func phase37SetEnv(env []string, name, value string) []string {
	prefix := name + "="
	out := make([]string, 0, len(env)+1)
	for _, item := range env {
		if !strings.HasPrefix(item, prefix) {
			out = append(out, item)
		}
	}
	return append(out, prefix+value)
}

func phase37CompareTuple(got phase37CLITuple, want testsupport.CRC32Tuple) error {
	if got.Exit != want.Exit || got.Stdout != want.Stdout || got.Stderr != want.Stderr {
		return fmt.Errorf("(exit,stdout,stderr) = (%d,%q,%q), want (%d,%q,%q)", got.Exit, got.Stdout, got.Stderr, want.Exit, want.Stdout, want.Stderr)
	}
	return nil
}

func phase37RequireTempClean(t *testing.T, path string) {
	t.Helper()
	entries, err := os.ReadDir(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("private temporary artifacts remain: %v", entries)
	}
}

func phase37RunDirectNULControls(t *testing.T, consumers []testsupport.CRC32Consumer, row testsupport.CRC32Case) {
	t.Helper()
	input := testsupport.CRC32MaterializeCase(t, t.TempDir(), row, nil)
	fakeBin := t.TempDir()
	sentinel := filepath.Join(t.TempDir(), "clang-was-invoked")
	clang := filepath.Join(fakeBin, "clang")
	script := "#!/bin/sh\n: > \"$PHASE37_CLANG_SENTINEL\"\nexit 99\n"
	if err := os.WriteFile(clang, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", fakeBin)
	t.Setenv("PHASE37_CLANG_SENTINEL", sentinel)
	for _, consumer := range consumers {
		for _, engine := range []string{"interpreter", "native"} {
			name := consumer.ID + "/" + engine + "/direct-public-dispatch"
			t.Run(name, func(t *testing.T) {
				args := phase37PublicFileArgs(engine, consumer, input)
				code, stdout, stderr := captureLangRun(t, args)
				got := phase37CLITuple{Exit: code, Stdout: stdout, Stderr: stderr}
				if err := phase37CompareTuple(got, row.Expected[consumer.ID]); err != nil {
					t.Fatal(err)
				}
				if _, err := os.Stat(sentinel); !os.IsNotExist(err) {
					t.Fatalf("NUL refusal launched Clang; sentinel stat error=%v", err)
				}
			})
		}
	}
}

func phase37RunManifestControls(t *testing.T, cli string, consumers []testsupport.CRC32Consumer, controls []testsupport.CRC32ManifestControl, recipes map[string]testsupport.CRC32OracleRow) {
	t.Helper()
	want := testsupport.CRC32Tuple{Exit: 3, Stderr: "schway run: native build failed\n", Outcome: "ManifestFailure"}
	for _, control := range controls {
		control := control
		t.Run(control.ID, func(t *testing.T) {
			for _, consumer := range consumers {
				consumer := consumer
				t.Run(consumer.ID, func(t *testing.T) {
					work := t.TempDir()
					inputRow := testsupport.CRC32Case{ID: "manifest-input", Kind: "file", Path: "manifest-input.bin", Recipe: "check"}
					input := testsupport.CRC32MaterializeCase(t, work, inputRow, recipes)
					manifest := filepath.Join(work, "bad.bindings.json")
					if control.Kind == "missing" {
						manifest = filepath.Join(work, "missing.bindings.json")
					} else if control.Kind == "malformed" {
						if err := os.WriteFile(manifest, []byte("{"), 0o600); err != nil {
							t.Fatal(err)
						}
					} else {
						t.Fatalf("unknown manifest control kind %q", control.Kind)
					}
					tmp := t.TempDir()
					args := []string{"run", "--engine=native", testsupport.ProjectPath(filepath.FromSlash(consumer.Source)), "--manifest", manifest, "--", input}
					got, err := phase37RunCLI(context.Background(), cli, args, work, tmp)
					if err != nil {
						t.Fatal(err)
					}
					if err := phase37CompareTuple(got, want); err != nil {
						t.Fatal(err)
					}
					phase37RequireTempClean(t, tmp)
				})
			}
		})
	}
}

func phase37RunParallelPair(t *testing.T, cli string, consumers []testsupport.CRC32Consumer, recipes map[string]testsupport.CRC32OracleRow) {
	t.Helper()
	consumerByID := make(map[string]testsupport.CRC32Consumer, len(consumers))
	for _, consumer := range consumers {
		consumerByID[consumer.ID] = consumer
	}
	type job struct {
		consumer testsupport.CRC32Consumer
		caseRow  testsupport.CRC32Case
		engine   string
		work     string
		tmp      string
		path     string
		want     testsupport.CRC32Tuple
	}
	makeJob := func(consumerID, recipeID, caseID, pathName string, answer uint64) job {
		work := t.TempDir()
		row := testsupport.CRC32Case{ID: caseID, Kind: "file", Path: pathName, Recipe: recipeID}
		path := testsupport.CRC32MaterializeCase(t, work, row, recipes)
		tmp := t.TempDir()
		return job{
			consumer: consumerByID[consumerID], caseRow: row, engine: "native", work: work, tmp: tmp, path: path,
			want: testsupport.CRC32Tuple{Exit: 0, Stdout: fmt.Sprintf("%d\n", answer), Outcome: "returned"},
		}
	}
	jobs := []job{
		makeJob("crc32", "check", "parallel-crc", "crc input.bin", 3421780262),
		makeJob("adler_low", "zero", "parallel-adler", "--json", 1),
	}
	type result struct {
		index int
		got   phase37CLITuple
		err   error
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	results := make(chan result, len(jobs))
	for index, item := range jobs {
		index, item := index, item
		go func() {
			args := phase37PublicFileArgs(item.engine, item.consumer, item.path)
			got, err := phase37RunCLI(ctx, cli, args, item.work, item.tmp)
			results <- result{index: index, got: got, err: err}
		}()
	}
	gotByIndex := make([]phase37CLITuple, len(jobs))
	for range jobs {
		result := <-results
		if result.err != nil {
			t.Fatalf("parallel child %d: %v", result.index, result.err)
		}
		gotByIndex[result.index] = result.got
	}
	for index, item := range jobs {
		if err := phase37CompareTuple(gotByIndex[index], item.want); err != nil {
			t.Fatalf("parallel %s outcome: %v", item.caseRow.ID, err)
		}
		phase37RequireTempClean(t, item.tmp)
	}
	if gotByIndex[0] == gotByIndex[1] {
		t.Fatalf("parallel applications cross-contaminated results: %+v", gotByIndex)
	}
}
