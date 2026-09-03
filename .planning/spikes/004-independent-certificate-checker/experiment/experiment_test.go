package experiment

import (
	"encoding/json"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"example.com/ai-lang/spike004/format"
	"example.com/ai-lang/spike004/producer"
	"example.com/ai-lang/spike004/verifier"
)

func fixture(t *testing.T) format.Artifact {
	t.Helper()
	artifact, err := LoadFixture("../fixtures/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	return artifact
}

func TestBothCertificateModesValidate(t *testing.T) {
	artifact := fixture(t)
	for _, mode := range []string{"recompute", "replay"} {
		certificate, err := producer.Build(artifact, mode)
		if err != nil {
			t.Fatal(err)
		}
		result := verifier.Verify(artifact, certificate)
		if !result.Valid {
			t.Fatalf("%s rejected: %+v", mode, result.Diagnostics)
		}
	}
}

func TestMutationMatrixHasOneExplicitTrustBoundaryEscape(t *testing.T) {
	report, err := Run(fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	if report.DetectedMutations != 12 || report.EscapedMutations != 1 {
		data, _ := json.MarshalIndent(report.Mutations, "", "  ")
		t.Fatalf("unexpected mutation score: %d detected, %d escaped\n%s", report.DetectedMutations, report.EscapedMutations, data)
	}
	for _, mutation := range report.Mutations {
		if mutation.Expected != "ESCAPES_BY_DESIGN" && mutation.Code != mutation.Expected {
			t.Errorf("%s: code=%q want=%q", mutation.ID, mutation.Code, mutation.Expected)
		}
	}
}

func TestCanonicalDigestIgnoresDeclarationAndSetOrder(t *testing.T) {
	base := fixture(t)
	want := format.ArtifactDigest(base)
	property := func(seed uint16) bool {
		candidate := format.CloneArtifact(base)
		random := rand.New(rand.NewSource(int64(seed)))
		random.Shuffle(len(candidate.Types), func(i, j int) { candidate.Types[i], candidate.Types[j] = candidate.Types[j], candidate.Types[i] })
		random.Shuffle(len(candidate.Functions), func(i, j int) {
			candidate.Functions[i], candidate.Functions[j] = candidate.Functions[j], candidate.Functions[i]
		})
		random.Shuffle(len(candidate.Traces), func(i, j int) { candidate.Traces[i], candidate.Traces[j] = candidate.Traces[j], candidate.Traces[i] })
		for index := range candidate.Functions {
			random.Shuffle(len(candidate.Functions[index].Published.Origins), func(i, j int) {
				candidate.Functions[index].Published.Origins[i], candidate.Functions[index].Published.Origins[j] = candidate.Functions[index].Published.Origins[j], candidate.Functions[index].Published.Origins[i]
			})
			random.Shuffle(len(candidate.Functions[index].Published.Abilities), func(i, j int) {
				candidate.Functions[index].Published.Abilities[i], candidate.Functions[index].Published.Abilities[j] = candidate.Functions[index].Published.Abilities[j], candidate.Functions[index].Published.Abilities[i]
			})
		}
		return format.ArtifactDigest(candidate) == want
	}
	if err := quick.Check(property, &quick.Config{MaxCount: 500, Rand: rand.New(rand.NewSource(0xC371F1CA))}); err != nil {
		t.Fatal(err)
	}
}

func TestDigestChangesForSemanticEventOrder(t *testing.T) {
	artifact := fixture(t)
	want := format.ArtifactDigest(artifact)
	artifact.Traces[0].Events[0], artifact.Traces[0].Events[1] = artifact.Traces[0].Events[1], artifact.Traces[0].Events[0]
	if format.ArtifactDigest(artifact) == want {
		t.Fatal("semantic event reorder did not change digest")
	}
}

func TestCertificateRoundTripIsStable(t *testing.T) {
	artifact := fixture(t)
	certificate, err := producer.Build(artifact, "replay")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(certificate)
	if err != nil {
		t.Fatal(err)
	}
	var decoded format.Certificate
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(certificate, decoded) {
		t.Fatal("certificate changed across JSON round trip")
	}
	if !verifier.Verify(artifact, decoded).Valid {
		t.Fatal("round-tripped certificate was rejected")
	}
}

func TestStableDiagnostics(t *testing.T) {
	artifact := fixture(t)
	certificate, _ := producer.Build(artifact, "recompute")
	artifact = mutateUseAfterMove(artifact)
	certificate.ArtifactDigest = format.ArtifactDigest(artifact)
	first := verifier.Verify(artifact, certificate)
	second := verifier.Verify(artifact, certificate)
	if !reflect.DeepEqual(first.Diagnostics, second.Diagnostics) {
		t.Fatalf("diagnostics changed: %+v != %+v", first.Diagnostics, second.Diagnostics)
	}
}

func TestScaleCheckCountIsLinear(t *testing.T) {
	for _, events := range []int{101, 1001, 10001} {
		artifact := ScaleArtifact(events)
		certificate, err := producer.Build(artifact, "recompute")
		if err != nil {
			t.Fatal(err)
		}
		result := verifier.Verify(artifact, certificate)
		if !result.Valid {
			t.Fatalf("events=%d diagnostics=%+v", events, result.Diagnostics)
		}
		if result.Checks > 2*events+8 {
			t.Fatalf("events=%d checks=%d is not linear with small constant", events, result.Checks)
		}
	}
}

func TestReplayEvidenceIsLargerThanRecomputeEvidence(t *testing.T) {
	artifact := ScaleArtifact(1001)
	recompute, _ := producer.Build(artifact, "recompute")
	replay, _ := producer.Build(artifact, "replay")
	compact, _ := json.Marshal(recompute)
	verbose, _ := json.Marshal(replay)
	if len(verbose) <= len(compact)*10 {
		t.Fatalf("replay=%d compact=%d; expected material evidence amplification", len(verbose), len(compact))
	}
}
