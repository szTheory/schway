package session_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

type phase16FileFrozenEvidenceManifest struct {
	Schema  string
	Records []struct {
		Fixture, FixtureSHA256, ProgramSHA256, Artifact, ArtifactSHA256 string
	}
}

func phase16FileFrozenEvidenceC(t *testing.T, supplied core.Program, fixture string) (string, error) {
	t.Helper()
	manifestBytes, err := os.ReadFile(testsupport.ProjectPath("testdata/phase16/file-frozen-evidence.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest phase16FileFrozenEvidenceManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil || manifest.Schema != "phase16.file-frozen-evidence/1" {
		t.Fatalf("invalid file frozen evidence manifest: %v", err)
	}
	source, err := os.ReadFile(testsupport.ProjectPath(fixture))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("%s: diagnostics: %+v", fixture, checked.Diagnostics)
	}
	canonical, err := json.Marshal(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	suppliedCanonical, err := json.Marshal(supplied)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(canonical, suppliedCanonical) {
		t.Fatalf("%s: supplied program is not the canonical checked fixture program", fixture)
	}
	if _, err := cgen.EmitNative(supplied); err == nil {
		t.Fatalf("%s: expected current public M004 refusal", fixture)
	}
	fixtureSum := sha256.Sum256(source)
	programSum := sha256.Sum256(canonical)
	for _, record := range manifest.Records {
		if record.Fixture != fixture {
			continue
		}
		if record.FixtureSHA256 != hex.EncodeToString(fixtureSum[:]) || record.ProgramSHA256 != hex.EncodeToString(programSum[:]) {
			t.Fatalf("%s: frozen evidence provenance changed", fixture)
		}
		artifact, err := os.ReadFile(testsupport.ProjectPath(record.Artifact))
		if err != nil {
			t.Fatal(err)
		}
		artifactSum := sha256.Sum256(artifact)
		if record.ArtifactSHA256 != hex.EncodeToString(artifactSum[:]) {
			t.Fatalf("%s: frozen artifact digest changed", fixture)
		}
		return string(artifact), nil
	}
	t.Fatalf("%s: no file frozen evidence record", fixture)
	return "", fmt.Errorf("%s: no file frozen evidence record", fixture)
}

func TestPhase16FileFrozenEvidenceBindsCanonicalProgram(t *testing.T) {
	checked := checkedPhase16Fixture(t, "testdata/phase4/nonlocal_exit_probe.lang")
	if got, err := phase16FileFrozenEvidenceC(t, checked.Program, "testdata/phase4/nonlocal_exit_probe.lang"); err != nil || got == "" {
		t.Fatal("empty frozen artifact")
	}
}

func TestPhase16FileFrozenEvidenceRejectsProgramSubstitution(t *testing.T) {
	checked := checkedPhase16Fixture(t, "testdata/phase4/nonlocal_exit_probe.lang")
	substitute := checkedPhase16Fixture(t, "testdata/phase4/acquire_three_success.lang")
	canonical, err := json.Marshal(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	substituted, err := json.Marshal(substitute.Program)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(canonical, substituted) {
		t.Fatal("substituted checked program unexpectedly matched canonical fixture program")
	}
}
