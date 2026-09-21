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

type fileFrozenEvidenceLedger struct {
	Schema  string
	Records []struct{ Fixture, FixtureSHA256, Artifact, ArtifactSHA256 string }
}

// generatedFrozenEvidenceLedger intentionally preserves the manifest's
// generator identity alongside the canonical Program bytes digest.  Generated
// controls are not file fixtures, so both facts are required to make the
// historical artifact binding fail closed.
type generatedFrozenEvidenceLedger struct {
	Schema    string
	Generator string
	Records   []struct{ ID, ProgramSHA256, Artifact, ArtifactSHA256 string }
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
	var fileFrozen fileFrozenEvidenceLedger
	read("testdata/phase16/file-frozen-evidence.json", &fileFrozen)
	var generatedFrozen generatedFrozenEvidenceLedger
	read("testdata/phase16/generated-frozen-evidence.json", &generatedFrozen)
	if err := validateLegacyEmitterEvidence(artifacts, evidence, func(path string) ([]byte, error) {
		return os.ReadFile(testsupport.ProjectPath(path))
	}); err != nil {
		t.Fatal(err)
	}
	if err := validateFileFrozenEvidence(fileFrozen, func(path string) ([]byte, error) { return os.ReadFile(testsupport.ProjectPath(path)) }); err != nil {
		t.Fatal(err)
	}
	if err := validateGeneratedFrozenEvidence(generatedFrozen, func(path string) ([]byte, error) { return os.ReadFile(testsupport.ProjectPath(path)) }); err != nil {
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

func validateGeneratedFrozenEvidence(ledger generatedFrozenEvidenceLedger, readFile func(string) ([]byte, error)) error {
	if ledger.Schema != "phase16.generated-frozen-evidence/1" || ledger.Generator != "EnumeratePhase5Closure/v1" || len(ledger.Records) == 0 {
		return fmt.Errorf("invalid generated frozen evidence manifest")
	}
	programs := make(map[string]string)
	for _, program := range session.EnumeratePhase5Closure() {
		canonical, err := json.Marshal(program)
		if err != nil {
			return fmt.Errorf("canonical generated program %q: %w", program.Module, err)
		}
		sum := sha256.Sum256(canonical)
		programs[program.Module] = hex.EncodeToString(sum[:])
	}
	seen := map[string]bool{}
	for _, record := range ledger.Records {
		if record.ID == "" || record.Artifact == "" || seen[record.ID] {
			return fmt.Errorf("invalid generated frozen record")
		}
		seen[record.ID] = true
		if got, ok := programs[record.ID]; !ok || got != record.ProgramSHA256 {
			return fmt.Errorf("generated program digest mismatch for %q", record.ID)
		}
		artifact, err := readFile(record.Artifact)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(artifact)
		if hex.EncodeToString(sum[:]) != record.ArtifactSHA256 {
			return fmt.Errorf("generated artifact digest mismatch for %q", record.ID)
		}
	}
	if len(seen) != len(programs) {
		return fmt.Errorf("generated frozen evidence does not cover every enumerated program")
	}
	return nil
}

func validateFileFrozenEvidence(ledger fileFrozenEvidenceLedger, readFile func(string) ([]byte, error)) error {
	if ledger.Schema != "phase16.file-frozen-evidence/1" || len(ledger.Records) == 0 {
		return fmt.Errorf("invalid file frozen evidence manifest")
	}
	seen := map[string]bool{}
	for _, r := range ledger.Records {
		if r.Fixture == "" || r.Artifact == "" || seen[r.Fixture] {
			return fmt.Errorf("invalid file frozen record")
		}
		seen[r.Fixture] = true
		source, err := readFile(r.Fixture)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(source)
		if hex.EncodeToString(sum[:]) != r.FixtureSHA256 {
			return fmt.Errorf("file fixture digest mismatch")
		}
		artifact, err := readFile(r.Artifact)
		if err != nil {
			return err
		}
		sum = sha256.Sum256(artifact)
		if hex.EncodeToString(sum[:]) != r.ArtifactSHA256 {
			return fmt.Errorf("file artifact digest mismatch")
		}
	}
	return nil
}

func TestFileFrozenEvidenceRejectsFaults(t *testing.T) {
	data, err := os.ReadFile(testsupport.ProjectPath("testdata/phase16/file-frozen-evidence.json"))
	if err != nil {
		t.Fatal(err)
	}
	var ledger fileFrozenEvidenceLedger
	if err := json.Unmarshal(data, &ledger); err != nil {
		t.Fatal(err)
	}
	read := func(path string) ([]byte, error) { return os.ReadFile(testsupport.ProjectPath(path)) }
	for _, mutate := range []func(*fileFrozenEvidenceLedger){
		func(l *fileFrozenEvidenceLedger) { l.Schema = "" },
		func(l *fileFrozenEvidenceLedger) { l.Records[0].FixtureSHA256 = strings.Repeat("0", sha256.Size*2) },
		func(l *fileFrozenEvidenceLedger) { l.Records[0].ArtifactSHA256 = strings.Repeat("0", sha256.Size*2) },
		func(l *fileFrozenEvidenceLedger) { l.Records = append(l.Records, l.Records[0]) },
	} {
		cloneBytes, _ := json.Marshal(ledger)
		var clone fileFrozenEvidenceLedger
		_ = json.Unmarshal(cloneBytes, &clone)
		mutate(&clone)
		if err := validateFileFrozenEvidence(clone, read); err == nil {
			t.Fatal("mutated file frozen evidence was accepted")
		}
	}
}

func TestGeneratedFrozenEvidenceRejectsFaults(t *testing.T) {
	data, err := os.ReadFile(testsupport.ProjectPath("testdata/phase16/generated-frozen-evidence.json"))
	if err != nil {
		t.Fatal(err)
	}
	var ledger generatedFrozenEvidenceLedger
	if err := json.Unmarshal(data, &ledger); err != nil {
		t.Fatal(err)
	}
	read := func(path string) ([]byte, error) { return os.ReadFile(testsupport.ProjectPath(path)) }
	for _, mutate := range []func(*generatedFrozenEvidenceLedger){
		func(l *generatedFrozenEvidenceLedger) { l.Schema = "" },
		func(l *generatedFrozenEvidenceLedger) { l.Generator = "wrong-generator/v1" },
		func(l *generatedFrozenEvidenceLedger) {
			l.Records[0].ProgramSHA256 = strings.Repeat("0", sha256.Size*2)
		},
		func(l *generatedFrozenEvidenceLedger) {
			l.Records[0].ArtifactSHA256 = strings.Repeat("0", sha256.Size*2)
		},
		func(l *generatedFrozenEvidenceLedger) { l.Records = append(l.Records, l.Records[0]) },
	} {
		cloneBytes, _ := json.Marshal(ledger)
		var clone generatedFrozenEvidenceLedger
		_ = json.Unmarshal(cloneBytes, &clone)
		mutate(&clone)
		if err := validateGeneratedFrozenEvidence(clone, read); err == nil {
			t.Fatal("mutated generated frozen evidence was accepted")
		}
	}
}

func TestLegacyEmitterEvidenceRejectsFaults(t *testing.T) {
	var artifacts legacyArtifactLedger
	var evidence legacyEvidenceLedger
	readLedger := func(path string, target any) {
		data, err := os.ReadFile(testsupport.ProjectPath(path))
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, target); err != nil {
			t.Fatal(err)
		}
	}
	readLedger("testdata/phase16/legacy-emitter-artifacts.json", &artifacts)
	readLedger("testdata/phase16/legacy-emitter-evidence.json", &evidence)
	readFile := func(path string) ([]byte, error) { return os.ReadFile(testsupport.ProjectPath(path)) }
	validate := func(artifacts legacyArtifactLedger, evidence legacyEvidenceLedger) error {
		return validateLegacyEmitterEvidence(artifacts, evidence, readFile)
	}

	for _, fault := range []struct {
		name   string
		mutate func(*legacyArtifactLedger, *legacyEvidenceLedger)
	}{
		{"duplicate entry", func(_ *legacyArtifactLedger, evidence *legacyEvidenceLedger) {
			evidence.Records = append(evidence.Records, evidence.Records[0])
		}},
		{"omitted artifact", func(artifacts *legacyArtifactLedger, _ *legacyEvidenceLedger) {
			artifacts.Artifacts = artifacts.Artifacts[1:]
		}},
		{"changed digest", func(artifacts *legacyArtifactLedger, _ *legacyEvidenceLedger) {
			artifacts.Artifacts[0].SHA256 = strings.Repeat("0", sha256.Size*2)
		}},
		{"artifact for different fixture", func(_ *legacyArtifactLedger, evidence *legacyEvidenceLedger) {
			evidence.Records[0].Artifact = evidence.Records[1].Artifact
		}},
		{"fixture identity", func(_ *legacyArtifactLedger, evidence *legacyEvidenceLedger) {
			evidence.Records[0].Fixture = evidence.Records[1].Fixture
		}},
		{"refusal code", func(_ *legacyArtifactLedger, evidence *legacyEvidenceLedger) {
			evidence.Records[0].Refusal = pointerM004Refusal
		}},
	} {
		t.Run(fault.name, func(t *testing.T) {
			mutatedArtifacts, mutatedEvidence := cloneLegacyLedgers(t, artifacts, evidence)
			fault.mutate(&mutatedArtifacts, &mutatedEvidence)
			if err := validate(mutatedArtifacts, mutatedEvidence); err == nil {
				t.Fatal("mutated provenance registry was accepted")
			}
		})
	}

	t.Run("artifact bytes", func(t *testing.T) {
		if err := validateLegacyEmitterEvidence(artifacts, evidence, func(path string) ([]byte, error) {
			data, err := readFile(path)
			if path == "testdata/phase16/historical/restrict_borrow.c" {
				data = append(data, byte('!'))
			}
			return data, err
		}); err == nil {
			t.Fatal("altered frozen artifact bytes were accepted")
		}
	})
}

func cloneLegacyLedgers(t *testing.T, artifacts legacyArtifactLedger, evidence legacyEvidenceLedger) (legacyArtifactLedger, legacyEvidenceLedger) {
	t.Helper()
	data, err := json.Marshal(struct {
		Artifacts legacyArtifactLedger `json:"artifacts"`
		Evidence  legacyEvidenceLedger `json:"evidence"`
	}{artifacts, evidence})
	if err != nil {
		t.Fatal(err)
	}
	var clone struct {
		Artifacts legacyArtifactLedger `json:"artifacts"`
		Evidence  legacyEvidenceLedger `json:"evidence"`
	}
	if err := json.Unmarshal(data, &clone); err != nil {
		t.Fatal(err)
	}
	return clone.Artifacts, clone.Evidence
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
