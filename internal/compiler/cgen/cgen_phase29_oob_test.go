package cgen

import (
	"regexp"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

const phase29BoundsSource = `module phase29.bounds.order

export {
  fn main
}

fn main(n: U64) -> U64 {
  n
}

fn probe(bytes: ByteView) -> U64 {
  var n = byte_len(bytes)
  var first = byte_at(bytes, 0)
  var i = 0
  var total = first
  while i < n {
    var b = byte_at(bytes, i)
    total = total + b
    i = i + 1
  }
  total
}
`

func phase29CheckedProgram(t *testing.T, source string) core.Program {
	t.Helper()
	parsed := syntax.Parse([]byte(source))
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("parse: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("check: %+v", checked.Diagnostics)
	}
	return checked.Program
}

// TestPhase29BoundsCheckPrecedesLoad pins, for every OpBytesAt in both shells,
// that an ordinary (never assert) bounds test textually precedes the load.
func TestPhase29BoundsCheckPrecedesLoad(t *testing.T) {
	program := phase29CheckedProgram(t, phase29BoundsSource)
	atCount := 0
	for _, function := range program.Functions {
		if function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind == core.OpBytesAt {
				atCount++
			}
		}
	}
	if atCount < 2 {
		t.Fatalf("source lowered %d byte_at operations, want at least 2", atCount)
	}
	boundsRE := regexp.MustCompile(`if \((\w+) >= (\w+)\.length\) \{`)
	for _, shell := range []struct {
		name     string
		shell    programEntryShell
		json     bool
		contains []string
	}{
		{"application", programApplicationShell, false, []string{"fputs(\"schway: byte index out of bounds\\n\", stderr);", "exit(65);"}},
		{"conformance", programExecutionShell, true, []string{"\"byte index out of bounds\"", "exit(0);"}},
	} {
		t.Run(shell.name, func(t *testing.T) {
			text, err := emitProgramWithShell(program, shell.shell, shell.json)
			if err != nil {
				t.Fatalf("emit %s shell: %v", shell.name, err)
			}
			if strings.Contains(text, "assert") {
				t.Fatalf("%s shell text contains assert", shell.name)
			}
			if !strings.Contains(text, "const unsigned char *data;") {
				t.Fatalf("%s shell lacks the const byte view typedef", shell.name)
			}
			for _, want := range shell.contains {
				if !strings.Contains(text, want) {
					t.Fatalf("%s shell lacks %q", shell.name, want)
				}
			}
			loads := regexp.MustCompile(`\(uint64_t\)(\w+)\.data\[(\w+)\]`).FindAllStringSubmatchIndex(text, -1)
			if len(loads) != atCount {
				t.Fatalf("%s shell has %d byte loads, want %d", shell.name, len(loads), atCount)
			}
			for _, load := range loads {
				view, index := text[load[2]:load[3]], text[load[4]:load[5]]
				prefix := text[:load[0]]
				bounds := boundsRE.FindAllStringSubmatchIndex(prefix, -1)
				found := false
				if len(bounds) > 0 {
					last := bounds[len(bounds)-1]
					between := prefix[last[1]:]
					// The nearest preceding bounds test must guard this exact
					// index and view, with only its own terminal block between.
					found = prefix[last[2]:last[3]] == index && prefix[last[4]:last[5]] == view && strings.Contains(between, shell.contains[1]+"\n  }\n") && !strings.Contains(between, ".data[")
				}
				if !found {
					t.Fatalf("%s shell: load of %s[%s] is not preceded by its bounds test", shell.name, view, index)
				}
			}
		})
	}
}
