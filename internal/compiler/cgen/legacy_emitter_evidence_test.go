package cgen_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

type legacyArtifactLedger struct {
	Schema    string                `json:"schema"`
	CutCommit string                `json:"cut_commit"`
	Artifacts []legacyArtifactEntry `json:"artifacts"`
}

type legacyArtifactEntry struct{ Fixture, Path, SHA256 string }

type legacyEvidenceLedger struct {
	Schema    string                `json:"schema"`
	CutCommit string                `json:"cut_commit"`
	Records   []legacyEvidenceEntry `json:"records"`
}

type legacyEvidenceEntry struct {
	Fixture       string `json:"fixture"`
	FixtureSHA256 string `json:"fixture_sha256"`
	Family        string `json:"family"`
	Artifact      string `json:"artifact"`
	Refusal       string `json:"refusal"`
	Witness       string `json:"witness"`
}

const (
	legacyArtifactSchema = "phase16.legacy-emitter-artifacts/1"
	legacyEvidenceSchema = "phase16.legacy-emitter-evidence/1"
	foreignM004Refusal   = "multi-function foreign-call bodies are not supported by native emission this phase"
	pointerM004Refusal   = "by-pointer bodies are not supported by whole-program native emission this phase"
)

func TestLegacyEmitterEvidence(t *testing.T) {
	var artifacts legacyArtifactLedger
	var evidence legacyEvidenceLedger
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
	read("testdata/phase16/legacy-emitter-evidence.json", &evidence)
	if err := validateLegacyEmitterEvidence(artifacts, evidence, func(path string) ([]byte, error) {
		return os.ReadFile(testsupport.ProjectPath(path))
	}); err != nil {
		t.Fatal(err)
	}
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

func validateLegacyEmitterEvidence(artifacts legacyArtifactLedger, evidence legacyEvidenceLedger, readFile func(string) ([]byte, error)) error {
	if artifacts.Schema != legacyArtifactSchema || evidence.Schema != legacyEvidenceSchema {
		return fmt.Errorf("unexpected legacy evidence schemas: artifacts=%q evidence=%q", artifacts.Schema, evidence.Schema)
	}
	if artifacts.CutCommit == "" || artifacts.CutCommit != evidence.CutCommit {
		return fmt.Errorf("artifact and evidence cut commits must be present and identical")
	}
	artifactByFixture := make(map[string]legacyArtifactEntry, len(artifacts.Artifacts))
	artifactPaths := make(map[string]struct{}, len(artifacts.Artifacts))
	for _, artifact := range artifacts.Artifacts {
		if artifact.Fixture == "" || artifact.Path == "" || len(artifact.SHA256) != sha256.Size*2 {
			return fmt.Errorf("artifact entry has absent or stale fields: %+v", artifact)
		}
		if _, duplicate := artifactByFixture[artifact.Fixture]; duplicate {
			return fmt.Errorf("duplicate artifact fixture %q", artifact.Fixture)
		}
		if _, duplicate := artifactPaths[artifact.Path]; duplicate {
			return fmt.Errorf("duplicate artifact path %q", artifact.Path)
		}
		data, err := readFile(artifact.Path)
		if err != nil {
			return fmt.Errorf("read artifact %q: %w", artifact.Path, err)
		}
		sum := sha256.Sum256(data)
		if got := hex.EncodeToString(sum[:]); got != artifact.SHA256 {
			return fmt.Errorf("artifact %q digest mismatch: got %s want %s", artifact.Path, got, artifact.SHA256)
		}
		artifactByFixture[artifact.Fixture] = artifact
		artifactPaths[artifact.Path] = struct{}{}
	}
	if len(artifactByFixture) == 0 {
		return fmt.Errorf("legacy artifact registry is empty")
	}

	evidenceFixtures := make(map[string]struct{}, len(evidence.Records))
	for _, record := range evidence.Records {
		if record.Fixture == "" || record.FixtureSHA256 == "" || record.Artifact == "" || record.Witness == "" {
			return fmt.Errorf("evidence entry has absent provenance fields: %+v", record)
		}
		if _, duplicate := evidenceFixtures[record.Fixture]; duplicate {
			return fmt.Errorf("duplicate evidence fixture %q", record.Fixture)
		}
		artifact, ok := artifactByFixture[record.Fixture]
		if !ok || artifact.Path != record.Artifact {
			return fmt.Errorf("evidence fixture %q is not bound to its own registered artifact", record.Fixture)
		}
		wantRefusal, ok := map[string]string{"foreign-m004": foreignM004Refusal, "by-pointer-m004": pointerM004Refusal}[record.Family]
		if !ok || record.Refusal != wantRefusal {
			return fmt.Errorf("evidence fixture %q has invalid M004 refusal identity", record.Fixture)
		}
		source, err := readFile(record.Fixture)
		if err != nil {
			return fmt.Errorf("read fixture %q: %w", record.Fixture, err)
		}
		sum := sha256.Sum256(source)
		if got := hex.EncodeToString(sum[:]); got != record.FixtureSHA256 {
			return fmt.Errorf("fixture %q digest mismatch: got %s want %s", record.Fixture, got, record.FixtureSHA256)
		}
		evidenceFixtures[record.Fixture] = struct{}{}
	}
	if len(evidenceFixtures) != len(artifactByFixture) {
		return fmt.Errorf("artifact/evidence registry is not bijective: artifacts=%d evidence=%d", len(artifactByFixture), len(evidenceFixtures))
	}
	return nil
}
