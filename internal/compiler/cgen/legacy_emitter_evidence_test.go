package cgen_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

type legacyArtifactLedger struct {
	Artifacts []struct{ Fixture, Path, SHA256 string } `json:"artifacts"`
}
type legacyEvidenceLedger struct {
	Records []struct{ Fixture, Artifact, Refusal string } `json:"records"`
}

func TestLegacyEmitterEvidence(t *testing.T) {
	var artifacts legacyArtifactLedger
	read := func(path string, target any) {
		data, err := os.ReadFile(testsupport.ProjectPath(path))
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, target); err != nil {
			t.Fatal(err)
		}
	}
	read("testdata/phase16/legacy-emitter-artifacts.json", &artifacts)
	read("testdata/phase16/legacy-emitter-evidence.json", new(legacyEvidenceLedger))
	for _, artifact := range artifacts.Artifacts {
		data, err := os.ReadFile(testsupport.ProjectPath(artifact.Path))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != artifact.SHA256 {
			t.Fatalf("%s digest mismatch", artifact.Path)
		}
	}
	var evidence legacyEvidenceLedger
	read("testdata/phase16/legacy-emitter-evidence.json", &evidence)
	for _, record := range evidence.Records {
		source, err := os.ReadFile(testsupport.ProjectPath(record.Fixture))
		if err != nil {
			t.Fatal(err)
		}
		checked := session.Check(source)
		if len(checked.Diagnostics) != 0 {
			t.Fatalf("%s: %v", record.Fixture, checked.Diagnostics)
		}
		if _, err := cgen.EmitNative(checked.Program); err == nil || !strings.Contains(err.Error(), record.Refusal) {
			t.Fatalf("%s did not preserve refusal %q: %v", record.Fixture, record.Refusal, err)
		}
	}
}

func TestLegacyEmitterEvidenceRejectsFaults(t *testing.T) {
	data, err := os.ReadFile(testsupport.ProjectPath("testdata/phase16/historical/restrict_borrow.c"))
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(append(data, byte('!')))
	if hex.EncodeToString(sum[:]) == "05a16af7e57c3a1a1e2b9af1eb4bed689d89fa53ff91e51328d51dd6f64e38f0" {
		t.Fatal("artifact digest mutation was inert")
	}
}
