package session

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

type phase16InternalFrozenEvidenceManifest struct {
	Schema  string
	Records []struct {
		Fixture, FixtureSHA256, ProgramSHA256, Artifact, ArtifactSHA256 string
	}
}

// phase16InternalFrozenEvidenceC is deliberately test-only. It preserves
// historical injector evidence without reopening the production M004 cut.
func phase16InternalFrozenEvidenceC(program core.Program, fixture string) (string, error) {
	manifestBytes, err := os.ReadFile(testsupport.ProjectPath("testdata/phase16/file-frozen-evidence.json"))
	if err != nil {
		return "", err
	}
	var manifest phase16InternalFrozenEvidenceManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil || manifest.Schema != "phase16.file-frozen-evidence/1" {
		return "", fmt.Errorf("invalid file frozen evidence manifest: %w", err)
	}
	source, err := os.ReadFile(testsupport.ProjectPath(fixture))
	if err != nil {
		return "", err
	}
	checked := Check(source)
	if len(checked.Diagnostics) != 0 {
		return "", fmt.Errorf("%s: diagnostics: %+v", fixture, checked.Diagnostics)
	}
	canonical, err := json.Marshal(checked.Program)
	if err != nil {
		return "", err
	}
	suppliedCanonical, err := json.Marshal(program)
	if err != nil {
		return "", err
	}
	if !bytes.Equal(canonical, suppliedCanonical) {
		return "", fmt.Errorf("%s: supplied program is not the canonical checked fixture program", fixture)
	}
	if _, err := Phase16ControlNativeC(program, fixture); err == nil {
		return "", fmt.Errorf("%s: expected current public M004 refusal", fixture)
	}
	fixtureSum := sha256.Sum256(source)
	programSum := sha256.Sum256(canonical)
	for _, record := range manifest.Records {
		if record.Fixture != fixture {
			continue
		}
		if record.FixtureSHA256 != hex.EncodeToString(fixtureSum[:]) || record.ProgramSHA256 != hex.EncodeToString(programSum[:]) {
			return "", fmt.Errorf("%s: frozen evidence provenance changed", fixture)
		}
		artifact, err := os.ReadFile(testsupport.ProjectPath(record.Artifact))
		if err != nil {
			return "", err
		}
		artifactSum := sha256.Sum256(artifact)
		if record.ArtifactSHA256 != hex.EncodeToString(artifactSum[:]) {
			return "", fmt.Errorf("%s: frozen artifact digest changed", fixture)
		}
		return string(artifact), nil
	}
	return "", fmt.Errorf("%s: no file frozen evidence record", fixture)
}
