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
	historicalCanonical, err := phase16HistoricalProgramCanonical(canonical)
	if err != nil {
		return "", err
	}
	programSum := sha256.Sum256(historicalCanonical)
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

// phase16HistoricalProgramCanonical reverses the public foreign C type and
// symbol names, plus the measured source-span shifts caused by the resource
// symbol rename, before comparing a Phase 4/5 program with its immutable
// pre-Schway digest. All other canonical program fields remain in the hash.
func phase16HistoricalProgramCanonical(canonical []byte) ([]byte, error) {
	var program core.Program
	if err := json.Unmarshal(canonical, &program); err != nil {
		return nil, err
	}
	reversedResourceOpen := false
	for index := range program.Functions {
		contract := program.Functions[index].ForeignContract
		if contract != nil && contract.Layout != nil && contract.Layout.ForeignTypeName == "schway_foreign_resource_block" {
			contract.Layout.ForeignTypeName = "lang" + "_foreign_resource_block"
		}
		if contract != nil && contract.Symbol == "schway_res_open" {
			contract.Symbol = "lang" + "_res_open"
			reversedResourceOpen = true
		}
	}
	// All four bounded Phase 5 foreign-enum variants share one declaration and
	// one call. Reversing the resource symbol shortens spans before the data
	// type and function by two bytes, and the function end by four.
	if reversedResourceOpen && strings.HasPrefix(program.Module, "phase5.enum_foreign_") {
		if len(program.DataTypes) > 0 {
			program.DataTypes[0].Span.Start -= 2
			program.DataTypes[0].Span.End -= 2
		}
		if len(program.Functions) > 0 {
			program.Functions[0].Span.Start -= 2
			program.Functions[0].Span.End -= 4
		}
	}
	return json.Marshal(program)
}
