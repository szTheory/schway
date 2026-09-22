package session

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/interp"
	"github.com/codename-lang/lang/internal/compiler/originvalidate"
	"github.com/codename-lang/lang/internal/compiler/syntax"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// TestPhase17ThreePeerAgreement compares the checker fact, core peer, and
// origin peer without giving any peer a return fact derived by another. The
// local oracle reads only the checked program's declared parameter/return
// type names and the source declaration that makes Result a nominal value;
// consequently coordinated peer faults cannot turn agreement into success.
func TestPhase17ThreePeerAgreement(t *testing.T) {
	source := phase17ReturnTracerSource(t)
	baseline := phase17PeerRun(t, source)
	if problems := phase17AgreementProblems(baseline, phase17DeclaredReturnOracle(source, baseline.checked)); len(problems) != 0 {
		t.Fatalf("unmodified peers disagree: %s", strings.Join(problems, "; "))
	}
	parameterBytes := phase17ParameterBytes(t, baseline.checked)

	for _, tc := range []struct {
		name string
		seed func() []func()
	}{
		{"checker", func() []func() { return []func(){check.SetPhase17ReturnLookupFaultForTest(true)} }},
		{"core", func() []func() { return []func(){corevalidate.SetPhase17ReturnLookupFaultForTest(true)} }},
		{"origin", func() []func() { return []func(){originvalidate.SetPhase17ReturnLookupFaultForTest(true)} }},
		{"coordinated", func() []func() {
			return []func(){
				check.SetPhase17ReturnLookupFaultForTest(true),
				corevalidate.SetPhase17ReturnLookupFaultForTest(true),
				originvalidate.SetPhase17ReturnLookupFaultForTest(true),
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			restores := tc.seed()
			for _, restore := range restores {
				t.Cleanup(restore)
			}
			faulted := phase17PeerRun(t, source)
			if got := phase17ParameterBytes(t, faulted.checked); !bytes.Equal(got, parameterBytes) {
				t.Fatalf("parameter facts changed under %s return-only seed", tc.name)
			}
			if problems := phase17AgreementProblems(faulted, phase17DeclaredReturnOracle(source, faulted.checked)); len(problems) == 0 {
				t.Fatalf("%s return-only seed passed the agreement gate", tc.name)
			}
			for _, restore := range restores {
				restore()
			}
		})
	}
}

type phase17Peers struct {
	checked core.Program
	core    corevalidate.Result
	origin  core.Interface
}

func phase17ReturnTracerSource(t testing.TB) []byte {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase17", "return_type_tracer.lang"))
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func phase17PeerRun(t testing.TB, source []byte) phase17Peers {
	t.Helper()
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("parse tracer: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("check tracer: %+v", checked.Diagnostics)
	}
	corePeer := corevalidate.Validate(checked.Program)
	originPeer, err := originvalidate.BuildInterface(checked.Program)
	if err != nil {
		t.Fatalf("originvalidate tracer: %v", err)
	}
	return phase17Peers{checked: checked.Program, core: corePeer, origin: originPeer}
}

func phase17ParameterBytes(t testing.TB, program core.Program) []byte {
	t.Helper()
	parameters := make([]core.TypeFact, 0, len(program.Functions))
	for _, function := range program.Functions {
		if function.Linear == nil || len(function.Linear.Types) < 2 {
			t.Fatalf("%s lacks directional type facts", function.Name)
		}
		parameters = append(parameters, function.Linear.Types[0])
	}
	encoded, err := json.Marshal(parameters)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func phase17DeclaredReturnOracle(source []byte, program core.Program) map[string]bool {
	// This deliberately does not call check/corevalidate/originvalidate return
	// derivation code. The canonical source declares Result as a nominal data
	// value, so every declared Result return is fresh independently of a peer.
	declaresResult := bytes.Contains(source, []byte("data Result"))
	want := make(map[string]bool, len(program.Functions))
	for _, function := range program.Functions {
		want[function.ID] = declaresResult && function.ReturnType == "Result"
	}
	return want
}

func phase17AgreementProblems(peers phase17Peers, oracle map[string]bool) []string {
	if !peers.core.Valid {
		return []string{"corevalidate rejected the checker-produced contract"}
	}
	coreSignatures := peers.core.PeerSignatures()
	originSignatures := make(map[string]core.FunctionSignature, len(peers.origin.Functions))
	for _, signature := range peers.origin.Functions {
		originSignatures[signature.ID] = signature
	}
	var problems []string
	for _, function := range peers.checked.Functions {
		if function.Linear == nil || len(function.Linear.Types) < 2 {
			return append(problems, function.Name+": missing directional type facts")
		}
		checkerReturnFresh := phase17FactHasAbility(function.Linear.Types[1], core.AbilityDrop)
		coreSignature, coreOK := coreSignatures[function.ID]
		originSignature, originOK := originSignatures[function.ID]
		if !coreOK || !originOK || len(coreSignature.Parameters) != 1 || len(originSignature.Parameters) != 1 {
			problems = append(problems, function.Name+": missing peer contract")
			continue
		}
		if function.Linear.Types[0].Shape.Constructor != coreSignature.Parameters[0].Type || function.Linear.Types[0].Shape.Constructor != originSignature.Parameters[0].Type {
			problems = append(problems, function.Name+": parameter type disagreement")
		}
		if function.ReturnType != coreSignature.Return.Type || function.ReturnType != originSignature.Return.Type {
			problems = append(problems, function.Name+": return type disagreement")
		}
		if checkerReturnFresh != oracle[function.ID] || coreSignature.Return.Fresh != oracle[function.ID] || originSignature.Return.Fresh != oracle[function.ID] {
			problems = append(problems, function.Name+": return freshness disagrees with declared-contract oracle")
		}
	}
	return problems
}

func phase17FactHasAbility(fact core.TypeFact, wanted core.Ability) bool {
	for _, ability := range fact.Abilities {
		if ability == wanted {
			return true
		}
	}
	return false
}

func TestPhase17InterpreterTracer(t *testing.T) {
	checked := Check(phase17ReturnTracerSource(t))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("check tracer: %+v", checked.Diagnostics)
	}
	first, err := interp.Run(checked.Program, "main", "Raw")
	if err != nil {
		t.Fatalf("interpret tracer: %v", err)
	}
	second, err := interp.Run(checked.Program, "main", "Raw")
	if err != nil {
		t.Fatalf("interpret tracer repeat: %v", err)
	}
	if first.Outcome.Kind != "returned" || first.Outcome.Value != "Raw" {
		t.Fatalf("tracer outcome = %+v, want returned Raw", first.Outcome)
	}
	firstBytes, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	secondBytes, err := json.Marshal(second)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstBytes, secondBytes) {
		t.Fatalf("tracer execution is not deterministic:\nfirst=%s\nsecond=%s", firstBytes, secondBytes)
	}
}

func phase17RepairFixture(t testing.TB, name string) []byte {
	t.Helper()
	path := testsupport.ProjectPath("testdata", "phase17", name)
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestPhase17RepairCorpusReachable(t *testing.T) {
	for _, name := range []string{"derivation_call_argument_mismatch.lang", "heldout_call_argument_mismatch.lang"} {
		t.Run(name, func(t *testing.T) {
			result := Check(phase17RepairFixture(t, name))
			if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "check.call_argument_type_mismatch" {
				t.Fatalf("%s diagnostics = %+v, want exactly check.call_argument_type_mismatch", name, result.Diagnostics)
			}
		})
	}
}

func TestPhase17RepairCorpusStructurallyDistinct(t *testing.T) {
	derivation := phase17RepairFixture(t, "derivation_call_argument_mismatch.lang")
	heldout := phase17RepairFixture(t, "heldout_call_argument_mismatch.lang")
	derivationTopology := phase17RepairTopology(derivation)
	heldoutTopology := phase17RepairTopology(heldout)
	if derivationTopology == heldoutTopology || heldoutTopology.hops < 2 {
		t.Fatalf("repair corpus topology is not distinct: derivation=%+v heldout=%+v", derivationTopology, heldoutTopology)
	}
	renamed := string(heldout)
	for _, replacement := range []struct{ from, to string }{{"dispatch", "deliver"}, {"relay", "forward"}, {"main", "entry"}, {"resource", "input"}, {"value", "output"}, {"result", "answer"}} {
		renamed = regexp.MustCompile(`\b`+replacement.from+`\b`).ReplaceAllString(renamed, replacement.to)
	}
	if bytes.Equal([]byte(renamed), heldout) {
		t.Fatal("alpha-rename control did not change source")
	}
	if topology := phase17RepairTopology([]byte(renamed)); topology != heldoutTopology {
		t.Fatalf("alpha rename changed topology: got=%+v want=%+v", topology, heldoutTopology)
	}
}

type phase17Topology struct{ functions, calls, hops int }

func phase17RepairTopology(source []byte) phase17Topology {
	text := string(source)
	functions := len(regexp.MustCompile(`(?m)^fn `).FindAllStringIndex(text, -1))
	calls := len(regexp.MustCompile(`(?m)^\s+let \w+ = \w+\(`).FindAllStringIndex(text, -1))
	hops := 0
	if functions == 5 && calls == 4 {
		hops = 2
	}
	return phase17Topology{functions: functions, calls: calls, hops: hops}
}

func TestPhase17RepairCorpusUniqueCandidate(t *testing.T) {
	heldout := string(phase17RepairFixture(t, "heldout_call_argument_mismatch.lang"))
	if strings.Count(heldout, "classify(value)") != 1 || strings.Count(heldout, "fn dispatch(resource: Resource)") != 1 {
		t.Fatalf("held-out fixture no longer has one mismatched call and one Resource parameter candidate")
	}
	if !strings.Contains(heldout, "let value = produce(resource)") {
		t.Fatal("held-out fixture lacks the real Result argument whose replacement is Resource")
	}
}

func TestPhase17RepairCorpusPhase13ControlsFrozen(t *testing.T) {
	manifest := phase17RepairFixture(t, "../phase13/HELDOUT.sha256")
	for _, line := range strings.Split(strings.TrimSpace(string(manifest)), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			t.Fatalf("malformed phase13 manifest line %q", line)
		}
		data, err := os.ReadFile(testsupport.ProjectPath(fields[1]))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != fields[0] {
			t.Fatalf("phase13 historical control changed: %s", fields[1])
		}
	}
}

func TestPhase17HeldoutSeal(t *testing.T) {
	manifestPath := testsupport.ProjectPath("testdata", "phase17", "HELDOUT.sha256")
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	listed := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || listed[fields[1]] {
			t.Fatalf("invalid or duplicate held-out manifest entry %q", line)
		}
		listed[fields[1]] = true
		data, err := os.ReadFile(testsupport.ProjectPath(fields[1]))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(data)
		if got := hex.EncodeToString(sum[:]); got != fields[0] {
			t.Fatalf("digest mismatch for %s: got %s want %s", fields[1], got, fields[0])
		}
	}
	entries, err := os.ReadDir(testsupport.ProjectPath("testdata", "phase17"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "heldout_") && strings.HasSuffix(entry.Name(), ".lang") && !listed["testdata/phase17/"+entry.Name()] {
			t.Fatalf("unlisted held-out fixture %s", entry.Name())
		}
	}
}

func TestPhase17HeldoutSealGuardIsNotInert(t *testing.T) {
	data := phase17RepairFixture(t, "heldout_call_argument_mismatch.lang")
	tampered := append([]byte(nil), data...)
	tampered[len(tampered)/2] ^= 1
	if sha256.Sum256(tampered) == sha256.Sum256(data) {
		t.Fatal("one-byte held-out mutation did not alter its digest")
	}
	if _, err := os.Stat(filepath.Join(t.TempDir(), "unlisted_heldout.lang")); !os.IsNotExist(err) {
		t.Fatal("temporary unlisted-file control is not isolated")
	}
}

// TestPhase17SourceFrontierMoved retains the pre-widening fixture commit as
// provenance and proves its refusals now reach the source-level call boundary.
func TestPhase17SourceFrontierMoved(t *testing.T) {
	for _, tc := range []struct {
		fixture string
		codes   []string
	}{
		{"return_type_tracer.lang", nil},
		{"call_argument_type_mismatch.lang", []string{"check.call_argument_type_mismatch"}},
		{"call_return_type_unrepresentable.lang", []string{"check.call_return_type_unrepresentable"}},
	} {
		t.Run(tc.fixture, func(t *testing.T) {
			checked, err := CheckFile(testsupport.ProjectPath("testdata", "phase17", tc.fixture))
			if err != nil {
				t.Fatalf("CheckFile: %v", err)
			}
			if len(checked.Diagnostics) != len(tc.codes) {
				t.Fatalf("diagnostic count = %d, want %d: %+v", len(checked.Diagnostics), len(tc.codes), checked.Diagnostics)
			}
			for i, code := range tc.codes {
				if checked.Diagnostics[i].Code != code {
					t.Fatalf("diagnostic[%d] = %q, want %q: %+v", i, checked.Diagnostics[i].Code, code, checked.Diagnostics)
				}
			}
		})
	}
}

func TestPhase17SourceFrontierFixturesAreParserValid(t *testing.T) {
	for _, fixture := range []string{
		"return_type_tracer.lang",
		"call_argument_type_mismatch.lang",
		"call_return_type_unrepresentable.lang",
	} {
		t.Run(fixture, func(t *testing.T) {
			result, err := FormatFile(testsupport.ProjectPath("testdata", "phase17", fixture))
			if err != nil || len(result.Diagnostics) != 0 {
				t.Fatalf("fixture must parse and format: result=%+v err=%v", result, err)
			}
		})
	}
}
