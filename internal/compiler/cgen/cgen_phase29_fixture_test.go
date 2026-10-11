package cgen_test

import (
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// TestPhase29FixtureMainEmission pins the fixture main in the application
// shell: acquire precedes the view, the view precedes the call, the call
// precedes the release (T-29-22), the callee's out-of-bounds path exits 65
// without any release call (D-20), no restrict is minted, and the conformance
// shell refuses the fixture owner with a typed error.
func TestPhase29FixtureMainEmission(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("examples", "phase29", "adler32_fixture.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("check diagnostics: %+v", checked.Diagnostics)
	}
	c, err := cgen.EmitApplication(checked.Program)
	if err != nil {
		t.Fatalf("EmitApplication: %v", err)
	}
	if !strings.Contains(c, "#include \"local/adapter.h\"") {
		t.Fatal("fixture-owner program does not include the adapter header")
	}
	if strings.Contains(c, "restrict") {
		t.Fatal("emitted C mints restrict")
	}
	mainStart := strings.Index(c, "schway_fixture_bytes_acquire_result schway_fixture_acquire_result")
	if mainStart < 0 {
		t.Fatalf("no fixture acquire in the emitted C:\n%s", c)
	}
	body := c[mainStart:]
	body = body[:strings.Index(body, "\n}\n")]
	ordered := []string{
		"schway_fixture_bytes_acquire(",
		"if (schway_fixture_acquire_result.status != 0) exit(65);",
		"schway_fixture_bytes_view(",
		"SCHWAY_BYTE_VIEW ",
		"/* call: ",
		"schway_fixture_bytes_release(",
		"return ",
	}
	previous := -1
	for _, fragment := range ordered {
		index := strings.Index(body, fragment)
		if index < 0 || index <= previous {
			t.Fatalf("fragment %q out of order (index %d after %d) in:\n%s", fragment, index, previous, body)
		}
		previous = index
	}
	for _, fragment := range []string{"schway_fixture_bytes_acquire(", "schway_fixture_bytes_view(", "schway_fixture_bytes_release("} {
		if strings.Count(body, fragment) != 1 {
			t.Fatalf("fixture main calls %q %d times, want once", fragment, strings.Count(body, fragment))
		}
	}

	// D-20: the ByteView callee leaves through exit(65) and never releases.
	matches := phase29ByteViewFunction.FindStringSubmatchIndex(c)
	if matches == nil {
		t.Fatalf("no ByteView callee in the emitted C:\n%s", c)
	}
	callee := c[matches[0]:]
	callee = callee[:strings.Index(callee, "\n}\n")]
	if !strings.Contains(callee, "exit(65);") {
		t.Fatalf("callee has no exit(65) out-of-bounds path:\n%s", callee)
	}
	if strings.Contains(callee, "schway_fixture_bytes_release") {
		t.Fatalf("callee releases the owner on its out-of-bounds path:\n%s", callee)
	}

	// The conformance shell never lowers the fixture owner.
	if _, err := cgen.EmitProgramNativeForTest(checked.Program); err == nil || !strings.Contains(err.Error(), "cgen.fixture_owner_shell_unsupported") {
		t.Fatalf("conformance shell error = %v, want cgen.fixture_owner_shell_unsupported", err)
	}
}
