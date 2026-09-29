package session_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

type phase16FileFrozenEvidenceManifest struct {
	Schema  string
	Records []struct {
		Fixture, FixtureSHA256, ProgramSHA256, Artifact, ArtifactSHA256, Refusal string
	}
}

type phase16FileFrozenEvidenceRecord struct {
	Fixture, FixtureSHA256, ProgramSHA256, Artifact, ArtifactSHA256, Refusal string
}

func validatePhase16FileFrozenEvidence(record phase16FileFrozenEvidenceRecord, fixture string, source, canonical, artifact []byte, refusal error) error {
	if record.Fixture != fixture {
		return fmt.Errorf("fixture identity changed: record %q, requested %q", record.Fixture, fixture)
	}
	fixtureSum := sha256.Sum256(source)
	historicalCanonical, err := phase16HistoricalProgramCanonical(canonical)
	if err != nil {
		return fmt.Errorf("normalize historical program identity: %w", err)
	}
	programSum := sha256.Sum256(historicalCanonical)
	artifactSum := sha256.Sum256(artifact)
	if record.FixtureSHA256 != hex.EncodeToString(fixtureSum[:]) {
		return fmt.Errorf("fixture digest changed")
	}
	if record.ProgramSHA256 != hex.EncodeToString(programSum[:]) {
		return fmt.Errorf("canonical program digest changed")
	}
	if record.ArtifactSHA256 != hex.EncodeToString(artifactSum[:]) {
		return fmt.Errorf("frozen artifact digest changed")
	}
	if record.Refusal != "" {
		if refusal == nil || refusal.Error() != record.Refusal {
			return fmt.Errorf("public emitter refusal changed: want %q, got %v", record.Refusal, refusal)
		}
	}
	return nil
}

// phase16HistoricalProgramCanonical reverses only the public C type rename
// in current Phase 4/5 programs before checking their pre-Schway evidence
// digest. Any other program-byte change remains visible to the digest.
func phase16HistoricalProgramCanonical(canonical []byte) ([]byte, error) {
	var program core.Program
	if err := json.Unmarshal(canonical, &program); err != nil {
		return nil, err
	}
	for index := range program.Functions {
		contract := program.Functions[index].ForeignContract
		if contract != nil && contract.Layout != nil && contract.Layout.ForeignTypeName == "schway_foreign_resource_block" {
			contract.Layout.ForeignTypeName = "lang" + "_foreign_resource_block"
		}
		if contract != nil && contract.Symbol == "schway_res_open" {
			contract.Symbol = "lang" + "_res_open"
		}
	}
	return json.Marshal(program)
}

// phase16FileFrozenEvidenceC validates the fixture and its current public
// refusal, then returns the digest-bound historical C artifact. It must never
// be treated as output produced by the current emitter.
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
	_, refusal := cgen.EmitNative(supplied)
	for _, record := range manifest.Records {
		if record.Fixture != fixture {
			continue
		}
		artifact, err := os.ReadFile(testsupport.ProjectPath(record.Artifact))
		if err != nil {
			t.Fatal(err)
		}
		if err := validatePhase16FileFrozenEvidence(phase16FileFrozenEvidenceRecord{
			Fixture: record.Fixture, FixtureSHA256: record.FixtureSHA256, ProgramSHA256: record.ProgramSHA256,
			Artifact: record.Artifact, ArtifactSHA256: record.ArtifactSHA256, Refusal: record.Refusal,
		}, fixture, source, canonical, artifact, refusal); err != nil {
			t.Fatalf("%s: %v", fixture, err)
		}
		return string(artifact), nil
	}
	t.Fatalf("%s: no file frozen evidence record", fixture)
	return "", fmt.Errorf("%s: no file frozen evidence record", fixture)
}

func TestPhase16Phase11FrozenEvidenceBindsCanonicalProgram(t *testing.T) {
	for _, fixture := range []string{
		"testdata/phase11/multi_function_gate_corpus.schway",
		"testdata/phase16/historical/phase11_gate_n_two.fixture",
	} {
		checked := checkedPhase16Fixture(t, fixture)
		artifact, err := phase16FileFrozenEvidenceC(t, checked.Program, fixture)
		if err != nil || artifact == "" {
			t.Fatalf("%s: empty frozen artifact: %v", fixture, err)
		}
	}
}

func TestPhase16Phase11FrozenEvidenceRejectsProvenanceFaults(t *testing.T) {
	fixture := "testdata/phase11/multi_function_gate_corpus.schway"
	checked := checkedPhase16Fixture(t, fixture)
	source, err := os.ReadFile(testsupport.ProjectPath(fixture))
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := json.Marshal(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	artifact, err := os.ReadFile(testsupport.ProjectPath("testdata/phase16/historical/phase11_gate_corpus.c"))
	if err != nil {
		t.Fatal(err)
	}
	_, refusal := cgen.EmitNative(checked.Program)
	base := phase16FileFrozenEvidenceRecord{
		Fixture: fixture, FixtureSHA256: "", ProgramSHA256: "", Artifact: "testdata/phase16/historical/phase11_gate_corpus.c", ArtifactSHA256: "", Refusal: "",
	}
	fixtureSum, programSum, artifactSum := sha256.Sum256(source), sha256.Sum256(canonical), sha256.Sum256(artifact)
	base.FixtureSHA256, base.ProgramSHA256, base.ArtifactSHA256 = hex.EncodeToString(fixtureSum[:]), hex.EncodeToString(programSum[:]), hex.EncodeToString(artifactSum[:])
	// Pin the exact witness from the checked-in record; a generic emitter error is insufficient.
	manifestBytes, err := os.ReadFile(testsupport.ProjectPath("testdata/phase16/file-frozen-evidence.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest phase16FileFrozenEvidenceManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatal(err)
	}
	for _, record := range manifest.Records {
		if record.Fixture == fixture {
			base.Refusal = record.Refusal
		}
	}
	if base.Refusal == "" {
		t.Fatal("missing refusal witness")
	}
	for name, mutate := range map[string]func(*phase16FileFrozenEvidenceRecord, *[]byte, *[]byte, *[]byte, *error){
		"refusal": func(r *phase16FileFrozenEvidenceRecord, _ *[]byte, _ *[]byte, _ *[]byte, e *error) {
			*e = fmt.Errorf("different emitter refusal")
		},
		"fixture identity": func(r *phase16FileFrozenEvidenceRecord, _ *[]byte, _ *[]byte, _ *[]byte, _ *error) {
			r.Fixture = "testdata/phase11/other.schway"
		},
		"canonical program": func(r *phase16FileFrozenEvidenceRecord, _ *[]byte, p *[]byte, _ *[]byte, _ *error) {
			*p = append(*p, 0)
		},
		"artifact digest": func(r *phase16FileFrozenEvidenceRecord, _ *[]byte, _ *[]byte, a *[]byte, _ *error) {
			*a = append(*a, 0)
		},
	} {
		t.Run(name, func(t *testing.T) {
			r, s, p, a, e := base, append([]byte(nil), source...), append([]byte(nil), canonical...), append([]byte(nil), artifact...), refusal
			mutate(&r, &s, &p, &a, &e)
			if err := validatePhase16FileFrozenEvidence(r, fixture, s, p, a, e); err == nil {
				t.Fatal("accepted provenance substitution")
			}
		})
	}
}

func TestPhase16FileFrozenEvidenceBindsCanonicalProgram(t *testing.T) {
	checked := checkedPhase16Fixture(t, "testdata/phase4/nonlocal_exit_probe.schway")
	if got, err := phase16FileFrozenEvidenceC(t, checked.Program, "testdata/phase4/nonlocal_exit_probe.schway"); err != nil || got == "" {
		t.Fatal("empty frozen artifact")
	}
}

func TestPhase16FileFrozenEvidenceRejectsProgramSubstitution(t *testing.T) {
	checked := checkedPhase16Fixture(t, "testdata/phase4/nonlocal_exit_probe.schway")
	substitute := checkedPhase16Fixture(t, "testdata/phase4/acquire_three_success.schway")
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
