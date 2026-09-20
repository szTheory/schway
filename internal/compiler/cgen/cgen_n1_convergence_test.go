package cgen

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/syntax"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// TestN1ConvergenceDifferential is the post-cut D-16-09 authority gate. It
// keeps the five pre-cut rows, but now proves every admitted public call is
// byte-identical to the direct sole implementation. The foreign row remains
// an explicit cut-M004 refusal on both paths; it is not admission evidence.
func TestN1ConvergenceDifferential(t *testing.T) {
	table := []struct {
		fixture  string
		admitted bool
	}{
		{fixture: "testdata/phase1/toggle.lang", admitted: true},
		{fixture: "testdata/phase2/owned_transfer.lang", admitted: true},
		{fixture: "testdata/phase3/borrowed_view.lang", admitted: true},
		{fixture: "testdata/phase4/foreign_acquire_one.lang", admitted: false},
		{fixture: "testdata/phase4/defect_terminal.lang", admitted: true},
	}
	if len(table) != 5 {
		t.Fatalf("expected exactly 5 fixtures in the N=1 convergence table, got %d", len(table))
	}

	for _, row := range table {
		row := row
		t.Run(row.fixture, func(t *testing.T) {
			source, err := os.ReadFile(testsupport.ProjectPath(row.fixture))
			if err != nil {
				t.Fatal(err)
			}
			parsed := syntax.Parse(source)
			if len(parsed.Diagnostics) != 0 {
				t.Fatalf("parse: %+v", parsed.Diagnostics)
			}
			checked := check.Program(parsed.Program)
			if len(checked.Diagnostics) != 0 {
				t.Fatalf("check: %+v", checked.Diagnostics)
			}

			for _, mode := range []struct {
				name   string
				public func() (string, error)
				direct func() (string, error)
			}{
				{name: "Emit", public: func() (string, error) { return Emit(checked.Program) }, direct: func() (string, error) { return emitProgram(checked.Program, false) }},
				{name: "EmitNative", public: func() (string, error) { return EmitNative(checked.Program) }, direct: func() (string, error) { return emitProgram(checked.Program, true) }},
			} {
				mode := mode
				t.Run(mode.name, func(t *testing.T) {
					public, publicErr := mode.public()
					direct, directErr := mode.direct()
					if row.admitted {
						if publicErr != nil || directErr != nil {
							t.Fatalf("admitted row: public err=%v direct err=%v", publicErr, directErr)
						}
						if public != direct {
							publicDigest := sha256.Sum256([]byte(public))
							directDigest := sha256.Sum256([]byte(direct))
							t.Fatalf("post-cut public/direct bytes differ: public_sha256=%s direct_sha256=%s", hex.EncodeToString(publicDigest[:]), hex.EncodeToString(directDigest[:]))
						}
						return
					}
					if publicErr == nil || directErr == nil {
						t.Fatalf("cut-M004 refusal lost: public err=%v direct err=%v", publicErr, directErr)
					}
				})
			}
		})
	}
}
