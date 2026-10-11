package session_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/session"
)

const phase29FixturePath = "../../../examples/phase29/adler32_fixture.schway"

func phase29FixtureProgram(t *testing.T) core.Program {
	t.Helper()
	source, err := os.ReadFile(phase29FixturePath)
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture diagnostics: %+v", checked.Diagnostics)
	}
	return checked.Program
}

// TestPhase29FixtureChecks proves the D-05/D-15 fixture program checks clean
// through check, corevalidate and originvalidate, that main is exactly the
// five-operation acquire/view/call/release/return shape, and that the frozen
// Phase 23 owner shape stays separate from the fixture owner (D-09).
func TestPhase29FixtureChecks(t *testing.T) {
	program := phase29FixtureProgram(t)
	var main *core.Function
	for index := range program.Functions {
		if program.Functions[index].Name == "main" {
			main = &program.Functions[index]
		}
	}
	if main == nil || main.Linear == nil {
		t.Fatal("fixture main is missing a linear body")
	}
	wantKinds := []core.OperationKind{core.OpForeignCall, core.OpForeignCall, core.OpCall, core.OpRelease, core.OpReturn}
	if len(main.Linear.Operations) != len(wantKinds) {
		t.Fatalf("main has %d operations, want %d: %+v", len(main.Linear.Operations), len(wantKinds), main.Linear.Operations)
	}
	for index, want := range wantKinds {
		if got := main.Linear.Operations[index].Kind; got != want {
			t.Fatalf("operation %d kind = %q, want %q", index, got, want)
		}
	}
	acquire, view, call, release := main.Linear.Operations[0], main.Linear.Operations[1], main.Linear.Operations[2], main.Linear.Operations[3]
	if acquire.SourceID != main.Parameter.ID || acquire.Foreign == nil || acquire.Foreign.Symbol != "schway_fixture_bytes_acquire" {
		t.Fatalf("acquire operation = %+v", acquire)
	}
	if view.Foreign == nil || view.Foreign.Mode != "borrow" || view.Foreign.ResultType != "ByteView" {
		t.Fatalf("view operation = %+v", view)
	}
	if call.CalleeID == "" || call.SourceID != view.TargetID {
		t.Fatalf("call operation = %+v, want the ByteView place as argument", call)
	}
	if release.ReleasesOperationID != acquire.ID || release.Foreign == nil || release.Foreign.Symbol != "schway_fixture_bytes_release" {
		t.Fatalf("release operation = %+v", release)
	}
	validated := corevalidate.Validate(program)
	if !validated.Valid {
		t.Fatalf("corevalidate rejected the fixture: %+v", validated.Problems)
	}
	if problems := originvalidate.ValidatePublished(validated.Program()); len(problems) != 0 {
		t.Fatalf("originvalidate rejected the fixture: %+v", problems)
	}

	t.Run("phase 23 file byte keeps its own shape and core hash", func(t *testing.T) {
		source, err := os.ReadFile("../../../examples/phase23/file_byte.schway")
		if err != nil {
			t.Fatal(err)
		}
		checked := session.Check(source)
		if len(checked.Diagnostics) != 0 {
			t.Fatalf("file_byte diagnostics: %+v", checked.Diagnostics)
		}
		document, err := json.Marshal(checked.Program)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(document)
		raw, err := os.ReadFile(phase272GoldenPath())
		if err != nil {
			t.Fatal(err)
		}
		var golden phase272CorpusGolden
		if err := json.Unmarshal(raw, &golden); err != nil {
			t.Fatal(err)
		}
		want := ""
		for _, entry := range golden.Entries {
			if entry.Path == "examples/phase23/file_byte.schway" {
				want = entry.CoreSHA256
			}
		}
		if want == "" {
			t.Fatal("file_byte is absent from the phase272 golden")
		}
		if got := hex.EncodeToString(sum[:]); got != want {
			t.Fatalf("file_byte core hash = %s, want golden %s", got, want)
		}
	})

	t.Run("phase 23 shape naming fixture symbols is refused", func(t *testing.T) {
		source, err := os.ReadFile("../../../examples/phase23/file_byte.schway")
		if err != nil {
			t.Fatal(err)
		}
		mixed := strings.NewReplacer(
			"schway_file_byte_acquire", "schway_fixture_bytes_acquire",
			"schway_file_byte_use", "schway_fixture_bytes_view",
			"schway_file_byte_release", "schway_fixture_bytes_release",
			"module phase23.file_byte", "module phase29.mixed_owner",
		).Replace(string(source))
		if mixed == string(source) || !strings.Contains(mixed, "path: PathToken") {
			t.Fatal("mixed source was not constructed")
		}
		checked := session.Check([]byte(mixed))
		if len(checked.Diagnostics) == 0 {
			t.Fatal("PathToken program naming the fixture symbols was accepted")
		}
		if code := checked.Diagnostics[0].Code; !strings.HasPrefix(code, "check.fixture_owner_") {
			t.Fatalf("first code = %q, want a check.fixture_owner_* refusal", code)
		}
	})
}
