package session_test

import (
	"os"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

func TestPhase19LiteralFrontier(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase19", "literal_tracer.lang"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(source), "let count = 42") {
		t.Fatal("literal_tracer.lang lost its direct numeric let")
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) == 0 {
		t.Fatal("literal_tracer.lang unexpectedly passed production checking")
	}
	first := checked.Diagnostics[0]
	if first.Code != "type.unknown" || strings.HasPrefix(first.Code, "syntax.") {
		t.Fatalf("literal_tracer.lang did not move to checker refusal: got id=%q code=%q span=%+v", first.ID, first.Code, first.Primary)
	}
}

func TestPhase19NumericRefusalFrontiers(t *testing.T) {
	for _, tc := range []struct {
		fixture string
		literal string
		code    string
		start   int
	}{
		{fixture: "literal_overflow.lang", literal: "18446744073709551616", code: "type.unknown", start: 99},
		{fixture: "literal_malformed.lang", literal: "0x_FF", code: "syntax.malformed_numeric_literal", start: 100},
	} {
		t.Run(tc.fixture, func(t *testing.T) {
			source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase19", tc.fixture))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(source), "let count = "+tc.literal) {
				t.Fatalf("%s lost literal witness %q", tc.fixture, tc.literal)
			}
			checked := session.Check(source)
			if len(checked.Diagnostics) == 0 {
				t.Fatalf("%s unexpectedly passed production checking", tc.fixture)
			}
			first := checked.Diagnostics[0]
			if first.Code != tc.code || (tc.code == "syntax.malformed_numeric_literal" && (first.Primary.Start != tc.start || first.Primary.End != tc.start+len(tc.literal))) {
				t.Fatalf("%s refusal moved: got id=%q code=%q span=%+v", tc.fixture, first.ID, first.Code, first.Primary)
			}
		})
	}
}

func TestPhase19ScalarGate(t *testing.T) {
	// Reuse the established D-12-18 mutation control so this gate inherits its
	// pinned corpus comparison instead of creating a second comparison law.
	TestPayloadCorpusCharacterizationReplayMutationKilled(t)
	TestPhase19LiteralFrontier(t)
}
