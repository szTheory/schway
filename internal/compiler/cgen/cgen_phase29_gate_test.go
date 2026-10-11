package cgen_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// The Phase 29 native proxy is a disposable, gate-only driver (D-14). It
// compiles the real application-shell C for a ByteView function and drives it
// from a hand-written C main. It is never the D-12 cap record.

const phase29OOBSource = `module phase29.gate.oob

export {
  fn main
}

fn past(bytes: ByteView) -> U64 {
  var n = byte_len(bytes)
  var x = byte_at(bytes, n)
  x
}

fn main(n: U64) -> U64 {
  var kept = n
  kept
}
`

var phase29ByteViewFunction = regexp.MustCompile(`static uint64_t (\w+)\(SCHWAY_BYTE_VIEW \w+, unsigned int invocation_index\) \{`)

// phase29Levels are the three distinct optimization lanes.
var phase29Levels = []struct {
	name  string
	flags []string
}{
	{"O0", []string{"-O0"}},
	{"O3", []string{"-O3"}},
	{"O3-flto", []string{"-O3", "-flto"}},
}

func phase29ApplicationC(t *testing.T, source []byte) (string, []string) {
	t.Helper()
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("check diagnostics: %+v", checked.Diagnostics)
	}
	// EmitApplication accepts the two-function unpruned program: the entry is
	// the U64-to-U64 main, and the ByteView function is emitted as a callee.
	c, err := cgen.EmitApplication(checked.Program)
	if err != nil {
		t.Fatalf("EmitApplication: %v", err)
	}
	var names []string
	for _, match := range phase29ByteViewFunction.FindAllStringSubmatch(c, -1) {
		names = append(names, match[1])
	}
	if len(names) == 0 {
		t.Fatalf("no ByteView function in the emitted C:\n%s", c)
	}
	return c, names
}

const phase29DriverBody = `
#undef main
static void phase29_fill(unsigned char *buffer, size_t count, unsigned char value) {
  for (size_t i = 0; i < count; i++) buffer[i] = value;
}
int main(int argc, char **argv) {
  static unsigned char big[65536];
  const char *which = argc > 1 ? argv[1] : "";
  size_t count = 0;
  const unsigned char *bytes = big;
  static const unsigned char wiki[] = {'W','i','k','i','p','e','d','i','a'};
  static const unsigned char one_a[] = {'a'};
  static const unsigned char pair[] = {0x80, 0xFF};
  static const unsigned char zero[] = {0x00};
  if (strcmp(which, "empty") == 0) { count = 0; }
  else if (strcmp(which, "zero") == 0) { bytes = zero; count = 1; }
  else if (strcmp(which, "a") == 0) { bytes = one_a; count = 1; }
  else if (strcmp(which, "wikipedia") == 0) { bytes = wiki; count = 9; }
  else if (strcmp(which, "z4096") == 0) { phase29_fill(big, 4096, 0x00); count = 4096; }
  else if (strcmp(which, "f5552") == 0) { phase29_fill(big, 5552, 0xFF); count = 5552; }
  else if (strcmp(which, "hi") == 0) { bytes = pair; count = 2; }
  else if (strcmp(which, "f4096") == 0) { phase29_fill(big, 4096, 0xFF); count = 4096; }
  else if (strcmp(which, "f16384") == 0) { phase29_fill(big, 16384, 0xFF); count = 16384; }
  else if (strcmp(which, "f65536") == 0) { phase29_fill(big, 65536, 0xFF); count = 65536; }
  else return 2;
  SCHWAY_BYTE_VIEW view = {bytes, (uint64_t)count};
  uint64_t result = PHASE29_TARGET(view, 0u);
  printf("%llu\n", (unsigned long long)result);
  return 0;
}
`

func phase29BuildProxy(t *testing.T, c, target string, flags []string, dir, name string) string {
	t.Helper()
	text := "/* proxy, gate-only (Phase 29 D-14) */\n" + c + "\n#define PHASE29_TARGET " + target + "\n" + phase29DriverBody
	cPath := filepath.Join(dir, name+".c")
	if err := os.WriteFile(cPath, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
	exe := filepath.Join(dir, name)
	args := []string{"-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic"}
	args = append(args, flags...)
	args = append(args, "-Dmain=schway_phase29_proxy_shell_main", "-o", exe, cPath)
	command := exec.Command("clang", args...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("clang %v: %v\n%s", flags, err, stderr.String())
	}
	return exe
}

func phase29ClangOrSkip(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("clang is absent; the native proxy needs the installed C compiler")
	}
}

func TestPhase29GateNativeProxyVectors(t *testing.T) {
	phase29ClangOrSkip(t)
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase29", "adler32_library.schway"))
	if err != nil {
		t.Fatal(err)
	}
	libraryC, libraryNames := phase29ApplicationC(t, source)
	oobC, oobNames := phase29ApplicationC(t, []byte(phase29OOBSource))

	// The bounds test is an ordinary if before the load, never assert.
	for label, c := range map[string]string{"library": libraryC, "oob": oobC} {
		guard := strings.Index(c, "if (")
		load := strings.Index(c, ".data[")
		check := strings.Index(c, ".length) {")
		if guard < 0 || load < 0 || check < 0 || check > load {
			t.Fatalf("%s: bounds test does not precede the byte load (check=%d load=%d)", label, check, load)
		}
		if !strings.Contains(c, "(uint64_t)schway_scalar_") {
			t.Fatalf("%s: byte load is not widened through an unsigned (uint64_t) cast", label)
		}
		if strings.Contains(c, "assert(") || strings.Contains(c, "restrict") {
			t.Fatalf("%s: emitted C contains assert or restrict", label)
		}
	}

	// Oracle literals: Python 3.14.4 zlib.adler32 and Go 1.24.0 hash/adler32.
	vectors := []struct{ selector, want string }{
		{"empty", "1"},
		{"zero", "65537"},
		{"a", "6422626"},
		{"wikipedia", "300286872"},
		{"z4096", "268435457"},
		{"f5552", "4052720524"},
		{"hi", "33620352"},
	}
	dir := t.TempDir()
	for _, level := range phase29Levels {
		library := phase29BuildProxy(t, libraryC, libraryNames[0], level.flags, dir, "adler-"+level.name)
		oob := phase29BuildProxy(t, oobC, oobNames[0], level.flags, dir, "oob-"+level.name)
		t.Run(level.name, func(t *testing.T) {
			for _, vector := range vectors {
				out, err := exec.Command(library, vector.selector).Output()
				if err != nil {
					t.Fatalf("%s: %v", vector.selector, err)
				}
				if got := strings.TrimSpace(string(out)); got != vector.want {
					t.Fatalf("%s: got %s, want %s", vector.selector, got, vector.want)
				}
			}
			for _, selector := range []string{"empty", "a", "hi"} {
				command := exec.Command(oob, selector)
				var stdout, stderr bytes.Buffer
				command.Stdout, command.Stderr = &stdout, &stderr
				err := command.Run()
				exit, ok := err.(*exec.ExitError)
				if !ok || exit.ExitCode() != 65 {
					t.Fatalf("oob %s: err = %v, want exit 65", selector, err)
				}
				if stderr.String() != "schway: byte index out of bounds\n" || stdout.Len() != 0 {
					t.Fatalf("oob %s: stderr %q stdout %q", selector, stderr.String(), stdout.String())
				}
			}
		})
	}
}

// TestPhase29GateNativeProxyCost is the native Stage-A reading (D-14). It is
// env-gated, proxy-only and never the D-12 cap record.
func TestPhase29GateNativeProxyCost(t *testing.T) {
	if os.Getenv("SCHWAY_SPIKE008") != "1" {
		t.Skip("SCHWAY_SPIKE008 is not 1; the native Stage-A cost reading is opt-in")
	}
	phase29ClangOrSkip(t)
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase29", "adler32_library.schway"))
	if err != nil {
		t.Fatal(err)
	}
	libraryC, names := phase29ApplicationC(t, source)
	workloads := []struct{ selector, want string }{
		{"f4096", "2170679522"},
		{"f16384", "2967061426"},
		{"f65536", "2006388466"},
	}
	dir := t.TempDir()
	for _, level := range phase29Levels {
		started := time.Now()
		exe := phase29BuildProxy(t, libraryC, names[0], level.flags, dir, "cost-"+level.name)
		t.Logf("proxy gate-only lane=%s compile=%s", level.name, time.Since(started))
		for _, workload := range workloads {
			var runs []time.Duration
			for run := 0; run < 7; run++ {
				begin := time.Now()
				out, err := exec.Command(exe, workload.selector).Output()
				elapsed := time.Since(begin)
				if err != nil {
					t.Fatalf("%s %s: %v", level.name, workload.selector, err)
				}
				if got := strings.TrimSpace(string(out)); got != workload.want {
					t.Fatalf("%s %s run %d: got %s, want %s", level.name, workload.selector, run, got, workload.want)
				}
				if run > 0 {
					runs = append(runs, elapsed)
				}
			}
			sort.Slice(runs, func(i, j int) bool { return runs[i] < runs[j] })
			t.Logf("proxy gate-only lane=%s workload=%s median=%s max=%s", level.name, workload.selector, runs[len(runs)/2], runs[len(runs)-1])
		}
	}
}
