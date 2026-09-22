package core_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/evidence"
	"github.com/codename-lang/lang/internal/compiler/interp"
	"github.com/codename-lang/lang/internal/compiler/originvalidate"
	"github.com/codename-lang/lang/internal/compiler/pathoracle"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/syntax"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// pinnedFacts is a fixed, machine-independent evidence.Facts value used only
// to pin core bytes and manifest IDs against a stable baseline. Real clang/Go
// identity strings would make this pin brittle across hosts; what D-04-23
// requires is that OUR changes never move these bytes, not that the pin
// reproduce a real toolchain's identity.
var pinnedFacts = evidence.Facts{
	CompilerIdentity: "test-pin/1", ClangIdentity: "test-pin-clang/1", Target: "test-pin-target",
	Flags: []string{"-pin"}, Policy: "test-pin-policy",
}

// TestPhase17TwoTypeCoreFacts pins the directional core identity that a call
// uses: the caller source matches the callee parameter type while the target
// uses the caller's separately minted return-side fact.
func TestPhase17TwoTypeCoreFacts(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase17", "return_type_tracer.lang"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("check canonical tracer: %+v", checked.Diagnostics)
	}
	var classify, main core.Function
	for _, function := range checked.Program.Functions {
		switch function.Name {
		case "classify":
			classify = function
		case "main":
			main = function
		}
	}
	if classify.ID == "" || main.ID == "" || classify.Linear == nil || main.Linear == nil {
		t.Fatalf("canonical tracer functions missing or malformed: %+v", checked.Program.Functions)
	}
	if classify.Parameter.Type != "Resource" || classify.ReturnType != "Result" || main.Parameter.Type != "Resource" || main.ReturnType != "Result" {
		t.Fatalf("tracer declarations collapsed: classify=%+v main=%+v", classify, main)
	}
	if len(classify.Linear.Types) < 2 || classify.Linear.Types[0].ID != classify.ID+":type:0" || classify.Linear.Types[1].ID != classify.ID+":type:1" {
		t.Fatalf("classify directional facts malformed: %+v", classify.Linear.Types)
	}
	if classify.Linear.Types[0].Shape.Constructor != "Resource" || classify.Linear.Types[1].Shape.Constructor != "Result" {
		t.Fatalf("classify facts do not preserve Resource -> Result: %+v", classify.Linear.Types[:2])
	}
	for _, operation := range main.Linear.Operations {
		if operation.Kind != core.OpCall {
			continue
		}
		if operation.CalleeID != classify.ID || operation.TypeID != main.ID+":type:1" {
			t.Fatalf("call does not target main's return-side fact: %+v", operation)
		}
		var source, target core.Place
		for _, place := range main.Linear.Places {
			if place.ID == operation.SourceID {
				source = place
			}
			if place.ID == operation.TargetID {
				target = place
			}
		}
		if source.TypeID != main.ID+":type:0" || target.TypeID != main.ID+":type:1" {
			t.Fatalf("call source/target facts = %q/%q, want main parameter/return facts", source.TypeID, target.TypeID)
		}
		if source.TypeID == target.TypeID || source.TypeID == operation.TypeID {
			t.Fatalf("call source and target facts collapsed: %+v", operation)
		}
		facts := map[string]core.TypeFact{}
		for _, fact := range main.Linear.Types {
			facts[fact.ID] = fact
		}
		if facts[source.TypeID].Shape.Constructor != classify.Parameter.Type || facts[target.TypeID].Shape.Constructor != classify.ReturnType {
			t.Fatalf("call source/target shapes = %q/%q, want classify parameter/return %q/%q", facts[source.TypeID].Shape.Constructor, facts[target.TypeID].Shape.Constructor, classify.Parameter.Type, classify.ReturnType)
		}
		return
	}
	t.Fatal("main has no call operation")
}

// pinnedFixture is one fixture this pin asserts is byte-identical to its
// value at the Phase 4 phase-start commit. CoreSHA256 and ManifestID were
// captured by running evidence.Build with pinnedFacts against the exact
// pre-Phase-4 source of each file (D-04-23): this test fails loudly the
// moment either value moves for a program no Phase 4 change should touch.
type pinnedFixture struct {
	Path       string
	CoreSHA256 string
	ManifestID string
}

var pinnedFixtures = []pinnedFixture{
	{"testdata/phase1/comments.lang", "1a90f92b261dbc825c3e480bd53465c10706e36a357f2491b0e0c88894166818", "evidence:f4570dd58a26d7c7582901a3"},
	{"testdata/phase1/toggle.lang", "5fc207e1a572c3a9e6ef04aa85010e1b842782e92635ae3f09949a3fb1353658", "evidence:142a7ab526b8f0cbd2cf4f26"},
	{"testdata/phase2/implicit_copy.lang", "5e2010f1331206d4610ca6a1b10ce686092a6884a8d4fc014cac3b44e4c36b2a", "evidence:6f46d2e225598990faccbd0b"},
	{"testdata/phase2/owned_transfer.lang", "e17fe549a50beb20c99e0172856d5df5c1736e97dce76d79c2d693c17d22fd20", "evidence:fb5bfdc8790a8f9c175ea872"},
	{"testdata/phase3/borrowed_view.lang", "696bf2e1c2ec73411ed8b091d753bef6e6504026cc43dc3b248ae62febba438f", "evidence:dddd4a0b3097658e6398e1cf"},
	{"testdata/phase3/branch_one_arm_shared_accept.lang", "c7ebef76cd22b510aec1a8509bcc89c78883f13dac4f1b94d0e217238a094118", "evidence:da42f3641c66baf35e645c66"},
	{"testdata/phase3/branch_view.lang", "3ec55ec5717e6760e5d3d006c2a5020f953a21fbf41e0adf406132e202365d9c", "evidence:214eeb3db43ca09b07561834"},
	{"testdata/phase3/public_view.lang", "b2b6ba18f5fc89ccbf983e15767ed71d27a9c9cfae907edf2e6cf09dac15452f", "evidence:9e8559ebc60ee839634e5294"},
	{"testdata/phase3/public_view_impossible.lang", "331a5b2d81849cfbfc31ba9430ff988d060874386bc93ad574f9429a516c493b", "evidence:1c0473336df43d779c6df70e"},
	{"testdata/phase3/public_view_mixed_access.lang", "b2665aa3d0cc9ed49d542a923878b4936e378c870c229638d0549b4feefa9e52", "evidence:0c125c8dc25ffe83c6a50de4"},
	{"testdata/phase3/public_view_multi_arm_access_conflict.lang", "e35c719e6b1ef7c246cb9635c1a8f656dbde3c6df3ed19d8211ea70be3758a6f", "evidence:6b624856c3bc19fa6f197172"},
	{"testdata/phase3/public_view_multi_arm_omitted.lang", "5c323c009c0d06132167fbe1c72be37cf3e6710a465159c6980544915a5a8ebd", "evidence:dee7242791bb8aa48524f84b"},
	{"testdata/phase3/public_view_omitted.lang", "f09350cfab25d9e527c6feee9f8c2fc89b22e8eaf530d016cf8421151c0c2cfc", "evidence:a038be9d5fa9c1d3dba6ba01"},
	{"testdata/phase3/public_view_understated.lang", "b45496eb4292454be7a919c1f63a7768060f383a1dc61f3f0a77f7aba65d8574", "evidence:40f4a643baf24f1a75d0422c"},
	{"testdata/phase3/sequential_shared_then_exclusive_accept.lang", "56d20794f8a43822f4483d2da39e870b4cb036c262f5e9289d32f5859bed20bd", "evidence:11dff9f506141504f46b90e0"},
	{"testdata/phase3/shared_shared_accept.lang", "ecc14ac0c851b90546fb41502b8915ac325ba322d3a3a7d4887b86961ed231b4", "evidence:b12a60b311e5cbad4e0274bf"},
	// Phase 4 accepting fixtures (D-05-39): widened from Phase 1-3 so this
	// pin also catches a Phase 5 emitter change that silently perturbs
	// Phase 4. Only the accepting testdata/phase4 fixtures are pinned here,
	// matching the Phase 3 precedent above (reject fixtures produce
	// diagnostics and are out of scope for this byte-identity pin); the set
	// mirrors native_test.go's phase4CorpusMatrix() "clean(...)" entries.
	{"testdata/phase4/acquire_three_fail_second.lang", "fcfd88bc97a15bdd0c774148e8efb23b2210d3f6e40ad101d0b3cfcf33591dee", "evidence:db391e8cbf6a999c40fbdaf4"},
	{"testdata/phase4/acquire_three_fail_third.lang", "1a23c824283bd12341da16197690f611884c8ef3b465458e134bfacd8dc7eb06", "evidence:1fb10eea4a10a005462dfaf6"},
	{"testdata/phase4/acquire_three_success.lang", "8bbce39409d78a99c55c02053deff4ee016dfbf9d03733ad6cd8a5e0362ec88a", "evidence:d68787f08a68a17f486d1783"},
	{"testdata/phase4/defect_terminal.lang", "4f348119f72c9d8aa1b3dc1cb91942176d74727297142ca940d44c08737bab0f", "evidence:12e9d68073ab86eb9eb463c9"},
	{"testdata/phase4/discard_because.lang", "6b1b048e4e0b2d3cc2791886d7b54f31de63788b4e2c5a8b185df51e5f652d89", "evidence:ac3003a93f4568b404aa319a"},
	{"testdata/phase4/foreign_acquire_one.lang", "718bed114e0754d3bfdb36a08d10649672915644d2eaca5caf266072f7e4dbf1", "evidence:dbfea02f20d97aff459e458b"},
	{"testdata/phase4/nonlocal_exit_probe.lang", "cde5fabf98be29972f21c4331bacde11fc135f1e3981c5ed495e23d554af0eb8", "evidence:0c768d94aeee2d03625e613b"},
	// Phase 5 accepting fixtures (D-06-31/D-06-32): widened from Phase 1-4 so
	// this pin also catches a Phase 6 lang.command/lang.verify-lane bump that
	// silently perturbs a Phase 5 program. Only the accepting testdata/phase5
	// fixtures are pinned here, matching the Phase 3/4 precedent above.
	// coordinated_lie.lang (and its coordinated_lie.core.json) is the
	// declared expected-escape pair (escape:coordinated-source-to-core-false-claim)
	// and is deliberately excluded from this table.
	{"testdata/phase5/allocator_mismatch.lang", "636097c74c3f161f532284bbaf2d6567e53413a7a9652f76c29da223dd0222c2", "evidence:1cdac22aba40cacc0f5e0001"},
	{"testdata/phase5/dead_store_unused_acquire.lang", "22a734be2932d02ecc47ced7c778f05e3d56296d4bfb9f3a7f4c1e543738fd1c", "evidence:f454a99f26276fb1545ce790"},
	{"testdata/phase5/defect_dies_by_signal.lang", "488f2f4dcd596626b6b82ddd9a2857c0068a53a37485c7e90104957b2607761a", "evidence:c9eea4002b7177106948da43"},
	{"testdata/phase5/false_restrict_hoist.lang", "2e2deae3e230984bf1430bb2d4c347d172444e22ad68f71c4d1197e34766f791", "evidence:ca3e057326af5276f18f5393"},
	{"testdata/phase5/inline_across_foreign.lang", "02fd41768398d0c650462790f15162f02c0eb5a0fc07bb97d15078d2a11cb53b", "evidence:2a9644fce2f925ca50841428"},
	{"testdata/phase5/reorder_two_events.lang", "5bc35a467aadffee60cb1ad7978ed17bf54f5f372c5f68d2ea9ebf81bf0ae56d", "evidence:1d7d61d8028a9472ee203b54"},
	{"testdata/phase5/restrict_borrow.lang", "15398f69e1d647b768f361cb0bedfc6064fee3f5fd5a00e03e88d573b8d96710", "evidence:65b0b4195299987d10f0ac80"},
	{"testdata/phase5/retained_pointer.lang", "7bac4e9375cbb0cfe1cae1eed15ba6278589220d21095018b78ac9081fa69297", "evidence:4d70b9513ffa2ed821dc26ad"},
	{"testdata/phase5/tail_collapse_release_ladder.lang", "f280d9999f29812956e1bec639aaed801342c60ae8706446852e25180c7f2594", "evidence:23a9ce0dc643fb05498e1963"},
	{"testdata/phase5/typed_failure_truncated_stdout.lang", "b146e3cd1f157d393fff9eecae8d5520db826ba6cbf635f4f1ef36ffd238886c", "evidence:598f270123ab0a123a99d351"},
}

// TestPreviousPhaseCoreBytesUnchanged pins every Phase 1-5 fixture's
// serialized core JSON to its exact byte value from before Phase 6
// (D-06-31/D-06-32, widened from the Phase 5 pin which stopped at Phase 4).
// It must be green before the coordinated lang.command and lang.verify-lane
// bump lands.
func TestPreviousPhaseCoreBytesUnchanged(t *testing.T) {
	for _, fixture := range pinnedFixtures {
		t.Run(fixture.Path, func(t *testing.T) {
			source, err := os.ReadFile(testsupport.ProjectPath(splitPath(fixture.Path)...))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			coreBytes := phase16CoreBytes(t, source)
			sum := sha256.Sum256(coreBytes)
			got := hex.EncodeToString(sum[:])
			if got != fixture.CoreSHA256 {
				t.Fatalf("core bytes moved for %s: got sha256 %s, want %s", fixture.Path, got, fixture.CoreSHA256)
			}
		})
	}
}

// phase16CoreBytes deliberately stops at the checked core boundary.  The
// historical pin is a source-to-core invariant, not an assertion that every
// pre-cut foreign/by-pointer fixture remains admitted to current native C
// lowering.  Keeping that boundary explicit prevents a public M004 refusal
// from laundering into an unrelated core-byte regression.
func phase16CoreBytes(t *testing.T, source []byte) []byte {
	t.Helper()
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("parse: %v", parsed.Diagnostics)
	}
	canonical := syntax.Format(parsed.Tree)
	checked := check.Program(syntax.Parse(canonical).Program)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("check: %v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("core validation: %v", validated.Problems)
	}
	encoded, err := json.Marshal(validated.Program())
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

// TestPreviousPhaseManifestIDsUnchanged is TestPreviousPhaseCoreBytesUnchanged's
// evidence-manifest-identity sibling (D-04-23), widened to Phase 5 by
// D-06-31/D-06-32. It must be green before the coordinated lang.command and
// lang.verify-lane bump lands.
func TestPreviousPhaseManifestIDsUnchanged(t *testing.T) {
	for _, fixture := range pinnedFixtures {
		t.Run(fixture.Path, func(t *testing.T) {
			source, err := os.ReadFile(testsupport.ProjectPath(splitPath(fixture.Path)...))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			product, diagnostics, err := evidence.Build(source, pinnedFacts)
			if err != nil {
				if phase16M004Refusal(t, source, err) {
					return
				}
				t.Fatalf("build: %v", err)
			}
			if len(diagnostics) > 0 {
				t.Fatalf("unexpected diagnostics: %v", diagnostics)
			}
			if product.Manifest.ID != fixture.ManifestID {
				t.Fatalf("manifest ID moved for %s: got %s, want %s", fixture.Path, product.Manifest.ID, fixture.ManifestID)
			}
		})
	}
}

func phase16M004Refusal(t *testing.T, source []byte, buildErr error) bool {
	t.Helper()
	message := buildErr.Error()
	if !strings.Contains(message, "multi-function foreign-call bodies are not supported") && !strings.Contains(message, "by-pointer bodies are not supported") {
		return false
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("cut fixture unexpectedly failed check: %v", checked.Diagnostics)
	}
	if _, err := cgen.Emit(checked.Program); err == nil || err.Error() != message {
		t.Fatalf("cut fixture refusal drift: build=%q emit=%v", message, err)
	}
	return true
}

// previousPhaseGoldenCDigests pins the SHA-256 digest of every committed
// *.golden.c file under testdata/phase1 through testdata/phase4 to its exact
// value before any Phase 5 emitter change (D-05-39). This is Task 1's
// tripwire: the by-pointer lowering emitter Task 2 adds must be additive, and
// this test is the only mechanical proof that it did not perturb a prior
// phase's committed generated-C golden. Digests were computed from the tree
// as it stood immediately before this plan's Task 2 change.
var previousPhaseGoldenCDigests = map[string]string{
	"testdata/phase1/generated.golden.c":               "1fd8aff8ee28de7ec39e559a7ca9ce50e480ecfffede617c36b2282c60cc122a",
	"testdata/phase2/owned_transfer.golden.c":          "f324f24db3ca0dfa8006b5c7fbec4263a6daf2dcbe220d49ea19167920799686",
	"testdata/phase4/foreign_layout_mismatch.golden.c": "3be6ebc36032ac9cc29bb916c1cdb8a8a996c0028ddf6982546f4c3dd5ffd031",
	"testdata/phase5/restrict_borrow.golden.c":         "05a16af7e57c3a1a1e2b9af1eb4bed689d89fa53ff91e51328d51dd6f64e38f0",
}

// phase16GoldenCutState makes the four pinned generated-C files an explicit
// two-state record.  PreCut is the only state checked in before Plan 16-09:
// its old digests describe the real files and it deliberately has no guessed
// future digest.  Plan 16-09 must switch every row to PostCut in the same
// authority-cut commit that changes the files and digest map.
type phase16GoldenCutState string

const (
	phase16GoldenPreCut  phase16GoldenCutState = "pre-cut"
	phase16GoldenPostCut phase16GoldenCutState = "post-cut"
)

type phase16GoldenChange struct {
	Path                string
	State               phase16GoldenCutState
	OldSHA256           string
	NewSHA256           string
	MovedResponsibility string
	StructuralReason    string
	SemanticWitness     string
	N1Fixture           string
	ReviewDisposition   string
}

// phase16GoldenChangeLedger is intentionally fixed at the same four paths as
// previousPhaseGoldenCDigests. A hash establishes file integrity only; the
// executable semantic witness records the independent evidence for the
// emitter responsibility change.
var phase16GoldenChangeLedger = []phase16GoldenChange{
	{Path: "testdata/phase1/generated.golden.c", State: phase16GoldenPostCut, OldSHA256: "f3e4fa6b641112fc8d213d04a38fce83dcfe0cd37ffbd79bc833ee787f11dc74", NewSHA256: "1fd8aff8ee28de7ec39e559a7ca9ce50e480ecfffede617c36b2282c60cc122a", MovedResponsibility: "legacy N=1 emitter to emitProgram", StructuralReason: "public match dispatch now emits the schema-2 program document", SemanticWitness: "TestN1ConvergenceDifferential/phase1/toggle.lang", N1Fixture: "testdata/phase1/toggle.lang", ReviewDisposition: "post-cut public/direct byte identity"},
	{Path: "testdata/phase2/owned_transfer.golden.c", State: phase16GoldenPostCut, OldSHA256: "91177543f89174fba680c70e79147d5ffc69714adefc8404f8de8dbdcdac65b8", NewSHA256: "f324f24db3ca0dfa8006b5c7fbec4263a6daf2dcbe220d49ea19167920799686", MovedResponsibility: "legacy N=1 emitter to emitProgram", StructuralReason: "public linear dispatch now emits the schema-2 program document", SemanticWitness: "TestN1ConvergenceDifferential/phase2/owned_transfer.lang", N1Fixture: "testdata/phase2/owned_transfer.lang", ReviewDisposition: "post-cut public/direct byte identity"},
	{Path: "testdata/phase4/foreign_layout_mismatch.golden.c", State: phase16GoldenPostCut, OldSHA256: "3be6ebc36032ac9cc29bb916c1cdb8a8a996c0028ddf6982546f4c3dd5ffd031", NewSHA256: "3be6ebc36032ac9cc29bb916c1cdb8a8a996c0028ddf6982546f4c3dd5ffd031", MovedResponsibility: "foreign lowering remains outside emitProgram", StructuralReason: "foreign lowering is explicit cut-M004 debt and not an admitted program shape", SemanticWitness: "TestProgramBranchValidationOrder/foreign_shape_precedes_preflight", N1Fixture: "testdata/phase4/foreign_layout_mismatch.lang", ReviewDisposition: "frozen cut-family baseline; refusal was not regenerated"},
	{Path: "testdata/phase5/restrict_borrow.golden.c", State: phase16GoldenPostCut, OldSHA256: "05a16af7e57c3a1a1e2b9af1eb4bed689d89fa53ff91e51328d51dd6f64e38f0", NewSHA256: "05a16af7e57c3a1a1e2b9af1eb4bed689d89fa53ff91e51328d51dd6f64e38f0", MovedResponsibility: "by-pointer lowering remains outside emitProgram", StructuralReason: "cut-M004 excludes every by-pointer family from program admission", SemanticWitness: "TestProgramBorrowedByPointerDisposition", N1Fixture: "testdata/phase5/restrict_borrow.lang", ReviewDisposition: "frozen cut-family baseline; refusal was not regenerated"},
}

func phase16GoldenLedgerProblems(digests map[string]string, ledger []phase16GoldenChange, current map[string]string) []string {
	var problems []string
	if len(ledger) != 4 {
		problems = append(problems, fmt.Sprintf("ledger count=%d, want 4", len(ledger)))
	}
	seen := make(map[string]bool, len(ledger))
	state := phase16GoldenCutState("")
	for _, entry := range ledger {
		if seen[entry.Path] {
			problems = append(problems, "duplicate ledger path "+entry.Path)
		}
		seen[entry.Path] = true
		if _, ok := digests[entry.Path]; !ok {
			problems = append(problems, "ledger path absent from digest map "+entry.Path)
		}
		if state == "" {
			state = entry.State
		} else if state != entry.State {
			problems = append(problems, "mixed cut states")
		}
		if entry.State != phase16GoldenPreCut && entry.State != phase16GoldenPostCut {
			problems = append(problems, "invalid state for "+entry.Path)
		}
		for field, value := range map[string]string{
			"old digest": entry.OldSHA256, "moved responsibility": entry.MovedResponsibility,
			"structural reason": entry.StructuralReason, "semantic witness": entry.SemanticWitness,
			"N=1 fixture": entry.N1Fixture, "review disposition": entry.ReviewDisposition,
		} {
			if value == "" {
				problems = append(problems, entry.Path+": missing "+field)
			}
		}
		if !isPhase16SHA256(entry.OldSHA256) {
			problems = append(problems, entry.Path+": malformed old digest")
		}
		if !strings.Contains(entry.SemanticWitness, "Test") {
			problems = append(problems, entry.Path+": semantic witness must name an executable test")
		}
		actual, known := current[entry.Path]
		if !known {
			problems = append(problems, entry.Path+": current file digest missing")
		}
		switch entry.State {
		case phase16GoldenPreCut:
			if entry.NewSHA256 != "" {
				problems = append(problems, entry.Path+": pre-cut row prematurely records new digest")
			}
			if digests[entry.Path] != entry.OldSHA256 || actual != entry.OldSHA256 {
				problems = append(problems, entry.Path+": pre-cut map/current digest must equal old digest")
			}
		case phase16GoldenPostCut:
			if !isPhase16SHA256(entry.NewSHA256) {
				problems = append(problems, entry.Path+": malformed or missing post-cut new digest")
			}
			if digests[entry.Path] != entry.NewSHA256 || actual != entry.NewSHA256 {
				problems = append(problems, entry.Path+": post-cut map/current/new digests disagree")
			}
		}
	}
	for path := range digests {
		if !seen[path] {
			problems = append(problems, "digest map path absent from ledger "+path)
		}
	}
	return problems
}

func isPhase16SHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func phase16CurrentGoldenDigests(t *testing.T) map[string]string {
	t.Helper()
	current := make(map[string]string, len(previousPhaseGoldenCDigests))
	for path := range previousPhaseGoldenCDigests {
		data, err := os.ReadFile(testsupport.ProjectPath(splitPath(path)...))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		sum := sha256.Sum256(data)
		current[path] = hex.EncodeToString(sum[:])
	}
	return current
}

func TestPhase16GoldenChangeLedger(t *testing.T) {
	if problems := phase16GoldenLedgerProblems(previousPhaseGoldenCDigests, phase16GoldenChangeLedger, phase16CurrentGoldenDigests(t)); len(problems) != 0 {
		t.Fatalf("Phase 16 golden-C ledger invalid:\n%s", strings.Join(problems, "\n"))
	}
}

func TestPhase16GoldenChangeLedgerRejectsFaults(t *testing.T) {
	baselineCurrent := phase16CurrentGoldenDigests(t)
	cloneLedger := func() []phase16GoldenChange { return append([]phase16GoldenChange(nil), phase16GoldenChangeLedger...) }
	cloneMap := func(source map[string]string) map[string]string {
		copy := make(map[string]string, len(source))
		for key, value := range source {
			copy[key] = value
		}
		return copy
	}
	tests := []struct {
		name, want string
		mutate     func([]phase16GoldenChange, map[string]string, map[string]string)
	}{
		{"malformed_old_digest", "malformed old digest", func(l []phase16GoldenChange, _, _ map[string]string) { l[0].OldSHA256 = "bad" }},
		{"missing_post_cut_digest", "malformed or missing post-cut new digest", func(l []phase16GoldenChange, _, _ map[string]string) { l[0].NewSHA256 = "" }},
		{"missing_witness", "missing semantic witness", func(l []phase16GoldenChange, _, _ map[string]string) { l[0].SemanticWitness = "" }},
		{"duplicate_path", "duplicate ledger path", func(l []phase16GoldenChange, _, _ map[string]string) { l[1].Path = l[0].Path }},
		{"stale_current_file", "post-cut map/current/new digests disagree", func(_ []phase16GoldenChange, _ map[string]string, current map[string]string) {
			current["testdata/phase1/generated.golden.c"] = strings.Repeat("0", 64)
		}},
		{"map_mismatch", "post-cut map/current/new digests disagree", func(_ []phase16GoldenChange, digests, _ map[string]string) {
			digests["testdata/phase1/generated.golden.c"] = strings.Repeat("0", 64)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ledger, digests, current := cloneLedger(), cloneMap(previousPhaseGoldenCDigests), cloneMap(baselineCurrent)
			test.mutate(ledger, digests, current)
			if problems := phase16GoldenLedgerProblems(digests, ledger, current); !strings.Contains(strings.Join(problems, "\n"), test.want) {
				t.Fatalf("fault %s passed or lacked actionable problem %q: %v", test.name, test.want, problems)
			}
		})
	}
	t.Run("post_cut_state_requires_coherent_new_baseline", func(t *testing.T) {
		ledger, digests, current := cloneLedger(), cloneMap(previousPhaseGoldenCDigests), cloneMap(baselineCurrent)
		if problems := phase16GoldenLedgerProblems(digests, ledger, current); len(problems) != 0 {
			t.Fatalf("coherent post-cut state rejected: %v", problems)
		}
		// Changing only the map simulates the forbidden partial cutover in
		// which the golden and ledger advanced but the pinned map did not.
		digests[ledger[0].Path] = ledger[0].OldSHA256
		if problems := phase16GoldenLedgerProblems(digests, ledger, current); !strings.Contains(strings.Join(problems, "\n"), "post-cut map/current/new digests disagree") {
			t.Fatalf("partial post-cut transition passed: %v", problems)
		}
	})
}

// TestPreviousPhaseGoldenCUnchanged hashes every committed *.golden.c under
// testdata/phase1 through testdata/phase5 and compares each against
// previousPhaseGoldenCDigests. It fails on any drift, naming the exact file
// path and both digests, and also fails if the corpus gains or loses a
// golden.c file relative to the pinned table -- so a Phase 6 change that
// accidentally perturbs a prior golden (or silently deletes one) is caught
// here rather than in review (D-06-31/D-06-32). It must be green before the
// coordinated lang.command and lang.verify-lane bump lands.
func TestPreviousPhaseGoldenCUnchanged(t *testing.T) {
	var found []string
	for _, phaseDir := range []string{"phase1", "phase2", "phase3", "phase4", "phase5"} {
		matches, err := filepath.Glob(testsupport.ProjectPath("testdata", phaseDir, "*.golden.c"))
		if err != nil {
			t.Fatalf("glob testdata/%s: %v", phaseDir, err)
		}
		for _, match := range matches {
			relative := "testdata/" + phaseDir + "/" + filepath.Base(match)
			found = append(found, relative)
		}
	}
	sort.Strings(found)

	seen := make(map[string]bool, len(found))
	for _, relative := range found {
		seen[relative] = true
		t.Run(relative, func(t *testing.T) {
			want, known := previousPhaseGoldenCDigests[relative]
			if !known {
				t.Fatalf("unpinned golden.c file %s found on disk; add its digest to previousPhaseGoldenCDigests", relative)
			}
			data, err := os.ReadFile(testsupport.ProjectPath(splitPath(relative)...))
			if err != nil {
				t.Fatalf("read %s: %v", relative, err)
			}
			sum := sha256.Sum256(data)
			got := hex.EncodeToString(sum[:])
			if got != want {
				t.Fatalf("golden.c bytes moved for %s: got sha256 %s, want %s", relative, got, want)
			}
		})
	}
	for relative := range previousPhaseGoldenCDigests {
		if !seen[relative] {
			t.Fatalf("pinned golden.c file %s no longer exists on disk", relative)
		}
	}
}

func splitPath(path string) []string {
	var parts []string
	start := 0
	for index := 0; index < len(path); index++ {
		if path[index] == '/' {
			parts = append(parts, path[start:index])
			start = index + 1
		}
	}
	parts = append(parts, path[start:])
	return parts
}

// TestAllOperationKindsRegistered asserts core.AllOperationKinds()'s length
// equals the number of declared OperationKind constants, and that no kind is
// listed twice, so a constant added without registering it is caught
// (D-04-22).
func TestAllOperationKindsRegistered(t *testing.T) {
	const declaredCount = 12 // OpCopy, OpMove, OpBorrowShared, OpBorrowExclusive, OpReturn, OpForeignCall, OpFail, OpRelease, OpDefect, OpCall, OpConstructPayload, OpDestructurePayload
	all := core.AllOperationKinds()
	if len(all) != declaredCount {
		t.Fatalf("AllOperationKinds() has %d entries, want %d", len(all), declaredCount)
	}
	seen := make(map[core.OperationKind]bool, len(all))
	for _, kind := range all {
		if seen[kind] {
			t.Fatalf("duplicate operation kind %q in registry", kind)
		}
		seen[kind] = true
	}
}

// TestTerminatorKindsIsSubsetOfAll proves core.TerminatorKinds() is a subset
// of core.AllOperationKinds().
func TestTerminatorKindsIsSubsetOfAll(t *testing.T) {
	all := make(map[core.OperationKind]bool)
	for _, kind := range core.AllOperationKinds() {
		all[kind] = true
	}
	for _, kind := range core.TerminatorKinds() {
		if !all[kind] {
			t.Fatalf("terminator kind %q is not in AllOperationKinds()", kind)
		}
	}
}

func linearProbeInput(function core.Function) (string, bool) {
	switch function.Parameter.Type {
	case "Byte":
		return "7", true
	case "Buffer":
		return "01020304", true
	default:
		return "", false
	}
}

// TestAllOperationKindsHandledAtEverySite is the control:kind.exhaustive_dispatch
// table (D-04-22): for every existing OperationKind, drive real programs
// containing that kind through check (session.Check), corevalidate,
// interp, cgen, pathoracle, and originvalidate, and assert none of them
// reject or crash. Every declared kind must be exercised by at least one
// fixture, so a kind that no corpus program ever produces cannot silently
// pass this control by omission.
//
// pathoracle and originvalidate do not switch exhaustively on
// core.OperationKind today (they walk only terminators); "handled" for these
// two sites means the walk completes without error on a program containing
// the kind, matching their current terminator-membership-test shape. Widening
// them to a stronger, kind-aware notion of "handled" is D-04-29, out of this
// plan's scope.
// exhaustiveDispatchFixtures is the in-process control's own literal,
// hand-maintained, phase-scoped fixture list (A-05). Extracted to a package
// var (rather than inlined in TestAllOperationKindsHandledAtEverySite) so
// Task 3's mutation-kill test (T-07-22/D-07-41) can drive
// runExhaustiveDispatchControl against a DIFFERENT, OpCall-free subset
// without duplicating this list.
var exhaustiveDispatchFixtures = []string{
	"testdata/phase1/toggle.lang",
	"testdata/phase1/comments.lang",
	"testdata/phase2/implicit_copy.lang",
	"testdata/phase2/owned_transfer.lang",
	"testdata/phase3/borrowed_view.lang",
	"testdata/phase3/branch_view.lang",
	"testdata/phase3/branch_one_arm_shared_accept.lang",
	"testdata/phase3/sequential_shared_then_exclusive_accept.lang",
	"testdata/phase3/shared_shared_accept.lang",
	"testdata/phase4/foreign_acquire_one.lang",
	"testdata/phase4/acquire_three_success.lang",
	"testdata/phase4/defect_terminal.lang",
	"testdata/phase07/call_basic.lang",
	"testdata/phase07/call_from_both_match_arms.lang",
	"testdata/phase12/payload_tracer.lang",
	"testdata/phase12/payload_drop_obligation.lang",
	"testdata/phase12/payload_borrow_interaction.lang",
}

// runExhaustiveDispatchControl is control:kind.exhaustive_dispatch.phase07_in_process's
// (and, historically, control:kind.exhaustive_dispatch's) own driving logic,
// extracted out of TestAllOperationKindsHandledAtEverySite into a function
// of its own two genuinely load-bearing parameters -- fixtures and
// requiredKinds -- so Task 3's mutation-kill test
// (TestPhase7DispatchControlsMutationKilled, T-07-22/D-07-41) can call the
// REAL control logic with a DIFFERENT requiredKinds list and prove the
// required-kinds loop at the bottom is what makes the control load-bearing,
// rather than duplicating this logic in a second, divergence-prone copy.
// This is this control's own unexported seam (D-07-42): a plain function
// parameter, never a package-level mutable var, so there is nothing for
// -race or -shuffle=on to trip over.
//
// Returns an error (never t.Fatalf) so the caller decides whether a given
// call is expected to fail (Test 1's "clean" beat) or expected to pass
// (Test 1's "mutated" beat).
func runExhaustiveDispatchControl(fixtures []string, requiredKinds []core.OperationKind) error {
	encountered := make(map[core.OperationKind]bool)
	for _, path := range fixtures {
		source, err := os.ReadFile(testsupport.ProjectPath(splitPath(path)...))
		if err != nil {
			return fmt.Errorf("%s: read: %w", path, err)
		}
		checked := session.Check(source)
		if len(checked.Diagnostics) > 0 {
			return fmt.Errorf("%s: unexpected diagnostics: %v", path, checked.Diagnostics)
		}
		program := checked.Program

		// corevalidate site.
		validated := corevalidate.Validate(program)
		if !validated.Valid {
			return fmt.Errorf("%s: corevalidate rejected: %v", path, validated.Problems)
		}
		program = validated.Program()
		calleeContracts := originvalidate.BuildCalleeOriginFacts(program)

		for _, function := range program.Functions {
			if function.Linear != nil {
				for _, operation := range function.Linear.Operations {
					encountered[operation.Kind] = true
				}
				// pathoracle site: must not error while walking.
				if function.Linear.ID != "" {
					if _, _, err := pathoracle.RecomputeEndpoints(function, nil); err != nil {
						return fmt.Errorf("%s/%s: pathoracle error: %w", path, function.Name, err)
					}
				}
			}
			// originvalidate site: must not crash while walking.
			_ = originvalidate.RecomputeOriginPerReturn(function, calleeContracts)

			// interp site.
			switch {
			case function.Match != nil:
				for _, arm := range function.Match.Arms {
					// Phase 10 (D-10-21/D-10-39) made core.OpCall a real,
					// executed operation, including from inside a match
					// arm's own block (e.g.
					// testdata/phase07/call_from_both_match_arms.lang) --
					// any interp error, including one from a call, fails
					// this control.
					if _, err := interp.Run(program, function.Name, arm.Pattern); err != nil {
						return fmt.Errorf("%s/%s/%s: interp error: %w", path, function.Name, arm.Pattern, err)
					}
				}
			case function.Linear != nil:
				input, ok := linearProbeInput(function)
				if ok {
					if _, err := interp.Run(program, function.Name, input); err != nil {
						return fmt.Errorf("%s/%s: interp error: %w", path, function.Name, err)
					}
				}
			}
		}

		// cgen site: Emit requires exactly one function.
		if len(program.Functions) == 1 {
			if _, err := cgen.Emit(program); err != nil {
				if strings.Contains(err.Error(), "multi-function foreign-call bodies are not supported") || strings.Contains(err.Error(), "by-pointer bodies are not supported") {
					continue
				}
				return fmt.Errorf("%s: cgen error: %w", path, err)
			}
		}
	}
	for _, kind := range requiredKinds {
		if !encountered[kind] {
			return fmt.Errorf("operation kind %q is never exercised by any corpus fixture in this control", kind)
		}
	}
	return nil
}

// TestAllOperationKindsHandledAtEverySite is the control:kind.exhaustive_dispatch
// table (D-04-22): for every existing OperationKind, drive real programs
// containing that kind through check (session.Check), corevalidate,
// interp, cgen, pathoracle, and originvalidate, and assert none of them
// reject or crash. Every declared kind must be exercised by at least one
// fixture, so a kind that no corpus program ever produces cannot silently
// pass this control by omission.
//
// pathoracle and originvalidate do not switch exhaustively on
// core.OperationKind today (they walk only terminators); "handled" for these
// two sites means the walk completes without error on a program containing
// the kind, matching their current terminator-membership-test shape. Widening
// them to a stronger, kind-aware notion of "handled" is D-04-29, out of this
// plan's scope.
func TestAllOperationKindsHandledAtEverySite(t *testing.T) {
	if err := runExhaustiveDispatchControl(exhaustiveDispatchFixtures, core.AllOperationKinds()); err != nil {
		t.Fatal(err)
	}
}

// TestLinearProbeInputExercisesCallBasicFixture is T-07-22's own assertion
// that the interp site is genuinely EXERCISED for the phase07 two-function
// linear fixture, never silently skipped: a skipped site (linearProbeInput
// returning ok == false) is indistinguishable from a covered one in
// TestAllOperationKindsHandledAtEverySite's bookkeeping, which is exactly
// the way that control could go green while proving nothing. Both of
// call_basic.lang's functions -- the caller (main) and the callee
// (identity) -- declare a Byte parameter, so linearProbeInput must return
// ok == true for both, driving interp.Run for each rather than skipping
// it.
func TestLinearProbeInputExercisesCallBasicFixture(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath(splitPath("testdata/phase07/call_basic.lang")...))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) > 0 {
		t.Fatalf("unexpected diagnostics: %v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("corevalidate rejected: %v", validated.Problems)
	}
	program := validated.Program()
	if len(program.Functions) != 2 {
		t.Fatalf("call_basic.lang: expected exactly 2 functions, got %d", len(program.Functions))
	}
	exercised := 0
	for _, function := range program.Functions {
		if function.Linear == nil {
			t.Fatalf("%s: expected a straight-line (function.Linear != nil) function", function.Name)
		}
		input, ok := linearProbeInput(function)
		if !ok {
			t.Fatalf("%s: linearProbeInput returned ok == false -- the interp site would be SILENTLY SKIPPED for this function, not exercised", function.Name)
		}
		if _, err := interp.Run(program, function.Name, input); err != nil {
			// Phase 10 made core.OpCall a real, executed operation: main's
			// call to identity now actually runs, so any interp error here
			// is a genuine failure.
			t.Fatalf("%s: interp error: %v", function.Name, err)
		}
		exercised++
	}
	if exercised != 2 {
		t.Fatalf("expected the interp site exercised for both of call_basic.lang's functions, got %d", exercised)
	}
}

// TestCancelledOutcomeIsUnconstructible is D-04-08's fail-closed control:
// the terminal-outcome axis reserves "cancelled" as unconstructible, and
// this asserts it over the REACHABLE CONSTRUCTORS in every engine's own
// source -- not over observed test output, which would be exactly the
// vacuous evidence shape this project has already paid for three times
// (03-08/03-09/03-10, D-10). No engine (interp, cgen, native) may ever
// contain the literal outcome-kind string "cancelled" as a value it
// constructs.
func TestCancelledOutcomeIsUnconstructible(t *testing.T) {
	for _, relative := range [][]string{
		{"internal", "compiler", "interp", "interp.go"},
		{"internal", "compiler", "cgen", "cgen.go"},
		{"internal", "compiler", "native", "native.go"},
	} {
		path := testsupport.ProjectPath(relative...)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for _, line := range strings.Split(string(data), "\n") {
			code := line
			if index := strings.Index(code, "//"); index >= 0 {
				code = code[:index]
			}
			if strings.Contains(code, `"cancelled"`) {
				t.Fatalf("%s constructs the reserved, unconstructible cancelled outcome kind: %q", path, line)
			}
		}
	}
}

// TestNoErrorValueConstructorExists is D-04-09's structural control: no
// operation kind, expression form, or core field constructs an error value.
// The only producer of typed_failure is an OpFail terminator, and the only
// producer of an OpFail is an err edge. This asserts the second half by
// hand-corrupting a real checked core.Program's sole err edge's Pattern and
// confirming corevalidate's independent re-derivation refuses it -- an
// OpFail reached from any predecessor other than a real err edge must never
// be admitted, even by a producer other than check.go itself.
func TestNoErrorValueConstructorExists(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", "foreign_acquire_one.lang"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) > 0 {
		t.Fatalf("unexpected diagnostics: %v", checked.Diagnostics)
	}
	program := checked.Program
	found := false
	for _, function := range program.Functions {
		if function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind == core.OpFail {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("fixture does not exercise OpFail")
	}

	corrupted := cloneCoreProgram(t, program)
	mutated := false
	for fi := range corrupted.Functions {
		linear := corrupted.Functions[fi].Linear
		if linear == nil {
			continue
		}
		for ei := range linear.Edges {
			if linear.Edges[ei].Pattern == "err" {
				linear.Edges[ei].Pattern = "ok"
				mutated = true
			}
		}
	}
	if !mutated {
		t.Fatal("fixture has no err edge to corrupt")
	}
	validated := corevalidate.Validate(corrupted)
	if validated.Valid {
		t.Fatal("corevalidate admitted an OpFail reached from a non-err edge")
	}
}

func cloneCoreProgram(t *testing.T, program core.Program) core.Program {
	t.Helper()
	data, err := json.Marshal(program)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var clone core.Program
	if err := json.Unmarshal(data, &clone); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return clone
}

// pinnedInterfaceV0JSON is a hand-written lang.interface/0 document, pinned
// at the exact bytes a pre-Stage-0 producer would have emitted (D-07-08).
// Following protocol_test.go:169-192's frozen-literal discipline: this
// string must never be regenerated to make a later test pass — a failure
// here means already-published /0 document bytes would have been
// perturbed.
const pinnedInterfaceV0JSON = `{"schema":"lang.interface/0","module_id":"m1","core_digest":"sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcd","functions":[{"id":"f1","name":"identity","parameter":{"id":"p1","name":"buffer","type":"Buffer"},"return_type":"Buffer","public_origin":{"paths":["buffer"],"access":"shared"},"abilities":["share"]}]}`

// TestFrozenInterfaceV0BytesUnchanged is 07-01 Task 2's Test 1 (D-07-08): a
// pinned /0 JSON literal decodes into core.InterfaceV0 field-for-field
// against a hand-written expected value, then re-encodes to those identical
// pinned bytes. core.InterfaceSchema itself must still be the exported
// "lang.interface/0" constant after the /1 bump.
func TestFrozenInterfaceV0BytesUnchanged(t *testing.T) {
	if core.InterfaceSchema != "lang.interface/0" {
		t.Fatalf("core.InterfaceSchema = %q, want frozen %q", core.InterfaceSchema, "lang.interface/0")
	}
	var v0 core.InterfaceV0
	if err := json.Unmarshal([]byte(pinnedInterfaceV0JSON), &v0); err != nil {
		t.Fatalf("decode pinned /0 literal: %v", err)
	}
	want := core.InterfaceV0{
		Schema: "lang.interface/0", ModuleID: "m1",
		CoreDigest: "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcd",
		Functions: []core.FunctionSignatureV0{{
			ID: "f1", Name: "identity",
			Parameter:    core.Parameter{ID: "p1", Name: "buffer", Type: "Buffer"},
			ReturnType:   "Buffer",
			PublicOrigin: &core.PublicOrigin{Paths: []string{"buffer"}, Access: "shared"},
			Abilities:    []core.Ability{core.AbilityShare},
		}},
	}
	if v0.Schema != want.Schema || v0.ModuleID != want.ModuleID || v0.CoreDigest != want.CoreDigest {
		t.Fatalf("decoded /0 top-level mismatch: got %+v, want %+v", v0, want)
	}
	if len(v0.Functions) != 1 || v0.Functions[0].ID != want.Functions[0].ID ||
		v0.Functions[0].PublicOrigin == nil || v0.Functions[0].PublicOrigin.Access != "shared" {
		t.Fatalf("decoded /0 function mismatch: got %+v, want %+v", v0.Functions, want.Functions)
	}
	reencoded, err := json.Marshal(v0)
	if err != nil {
		t.Fatalf("re-encode: %v", err)
	}
	if string(reencoded) != pinnedInterfaceV0JSON {
		t.Fatalf("re-encoded /0 bytes moved:\n got  %s\n want %s", reencoded, pinnedInterfaceV0JSON)
	}
}

// TestDecodeInterfaceV0NeverAdmissible is 07-01 Task 2's Test 2 (D-07-36/
// T-07-02): DecodeInterface on a /0 document returns the pinned InterfaceV0
// shape with Admissible always false; the V0 type itself has no Callable,
// Parameters, Return, or ClosureDigest field, so admission is structurally
// unreachable rather than merely refused.
func TestDecodeInterfaceV0NeverAdmissible(t *testing.T) {
	decoded, err := core.DecodeInterface([]byte(pinnedInterfaceV0JSON))
	if err != nil {
		t.Fatalf("expected a /0 document to decode cleanly, got %v", err)
	}
	if decoded.V0 == nil {
		t.Fatal("expected V0 to be populated for a lang.interface/0 document")
	}
	if decoded.V1 != nil {
		t.Fatal("expected V1 to stay nil for a lang.interface/0 document")
	}
	if decoded.Admissible {
		t.Fatal("expected a /0 document to never be Admissible")
	}
}

// validHexDigest is a syntactically valid sha256:+64-lowercase-hex digest
// shape, used only to satisfy DecodeInterface's shape check in hand-built
// fixtures below — it is not a real content digest of anything.
const validHexDigest = "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcd" + "ef"

// validInterfaceV1Document returns a fresh, independent map[string]any
// representation of a minimal but fully valid lang.interface/1 document, so
// each subtest below can safely mutate its own copy without cross-test
// interference.
func validInterfaceV1Document(t *testing.T) map[string]any {
	t.Helper()
	return map[string]any{
		"schema":      "lang.interface/1",
		"module_id":   "m1",
		"core_digest": validHexDigest,
		"functions": []any{
			map[string]any{
				"id":   "f1",
				"name": "identity",
				"parameters": []any{
					map[string]any{"id": "p1", "name": "buffer", "type": "Buffer", "mode": "owned", "drops": false},
				},
				"return":         map[string]any{"type": "Buffer", "mode": "owned", "paths": []any{}, "fresh": false},
				"abilities":      []any{},
				"callable":       false,
				"foreign":        map[string]any{"allocator": "", "unwind": "", "nonlocal_exit": ""},
				"closure_digest": validHexDigest,
			},
		},
	}
}

// deepCopyJSON round-trips value through JSON so a subtest can mutate its
// own independent copy of a shared nested-map fixture.
func deepCopyJSON(t *testing.T, value map[string]any) map[string]any {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal for deep copy: %v", err)
	}
	var clone map[string]any
	if err := json.Unmarshal(data, &clone); err != nil {
		t.Fatalf("unmarshal for deep copy: %v", err)
	}
	return clone
}

func firstFunction(t *testing.T, document map[string]any) map[string]any {
	t.Helper()
	functions, ok := document["functions"].([]any)
	if !ok || len(functions) == 0 {
		t.Fatal("expected at least one function in the fixture document")
	}
	function, ok := functions[0].(map[string]any)
	if !ok {
		t.Fatal("expected functions[0] to be an object")
	}
	return function
}

func decodeErrorCode(err error) string {
	var typed *core.DecodeError
	if err == nil {
		return ""
	}
	if de, ok := err.(*core.DecodeError); ok {
		typed = de
	}
	if typed == nil {
		return ""
	}
	return typed.Code
}

// TestDecodeInterfaceV1BaselineAccepted proves the shared fixture itself is
// valid before every mutation subtest below relies on it being refused only
// because of the ONE thing each subtest breaks.
func TestDecodeInterfaceV1BaselineAccepted(t *testing.T) {
	document := validInterfaceV1Document(t)
	data, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("marshal baseline: %v", err)
	}
	decoded, err := core.DecodeInterface(data)
	if err != nil {
		t.Fatalf("expected the baseline /1 document to decode cleanly, got %v", err)
	}
	if decoded.V1 == nil || !decoded.Admissible {
		t.Fatalf("expected the baseline /1 document to be Admissible: %+v", decoded)
	}
}

// TestDecodeInterfaceV1RequiredFieldsRefused is 07-01 Task 2's Test 3
// (D-07-36): DecodeInterface refuses a /1 document with a missing or empty
// required field. Each subtest breaks exactly one field of the shared valid
// baseline.
func TestDecodeInterfaceV1RequiredFieldsRefused(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(function map[string]any)
	}{
		{"missing Mode", func(function map[string]any) {
			parameters := function["parameters"].([]any)
			parameter := parameters[0].(map[string]any)
			delete(parameter, "mode")
		}},
		{"empty Mode", func(function map[string]any) {
			parameters := function["parameters"].([]any)
			parameter := parameters[0].(map[string]any)
			parameter["mode"] = ""
		}},
		{"missing Return", func(function map[string]any) {
			delete(function, "return")
		}},
		{"missing Foreign", func(function map[string]any) {
			delete(function, "foreign")
		}},
		{"empty ClosureDigest", func(function map[string]any) {
			function["closure_digest"] = ""
		}},
		{"empty Parameters where the function declares a parameter", func(function map[string]any) {
			function["parameters"] = []any{}
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			document := deepCopyJSON(t, validInterfaceV1Document(t))
			testCase.mutate(firstFunction(t, document))
			data, err := json.Marshal(document)
			if err != nil {
				t.Fatalf("marshal mutated document: %v", err)
			}
			decoded, err := core.DecodeInterface(data)
			if err == nil {
				t.Fatalf("expected refusal for %q, got a clean decode: %+v", testCase.name, decoded)
			}
			if code := decodeErrorCode(err); code != "core.interface_missing_field" {
				t.Fatalf("%s: expected code core.interface_missing_field, got %q (%v)", testCase.name, code, err)
			}
		})
	}
}

// TestDecodeInterfaceV1ValueDomainRefused is 07-01 Task 2's Test 4
// (D-07-36): DecodeInterface refuses a Mode outside the closed set, a
// malformed digest, a duplicate function ID, and any unknown schema value.
func TestDecodeInterfaceV1ValueDomainRefused(t *testing.T) {
	t.Run("Mode outside closed set", func(t *testing.T) {
		document := deepCopyJSON(t, validInterfaceV1Document(t))
		function := firstFunction(t, document)
		parameters := function["parameters"].([]any)
		parameters[0].(map[string]any)["mode"] = "unspecified"
		data, _ := json.Marshal(document)
		_, err := core.DecodeInterface(data)
		if code := decodeErrorCode(err); code != "core.interface_invalid_mode" {
			t.Fatalf("expected core.interface_invalid_mode, got %q (%v)", code, err)
		}
	})
	t.Run("malformed ClosureDigest", func(t *testing.T) {
		document := deepCopyJSON(t, validInterfaceV1Document(t))
		firstFunction(t, document)["closure_digest"] = "not-a-digest"
		data, _ := json.Marshal(document)
		_, err := core.DecodeInterface(data)
		if code := decodeErrorCode(err); code != "core.interface_invalid_digest" {
			t.Fatalf("expected core.interface_invalid_digest, got %q (%v)", code, err)
		}
	})
	t.Run("malformed CoreDigest", func(t *testing.T) {
		document := deepCopyJSON(t, validInterfaceV1Document(t))
		document["core_digest"] = "sha256:tooshort"
		data, _ := json.Marshal(document)
		_, err := core.DecodeInterface(data)
		if code := decodeErrorCode(err); code != "core.interface_invalid_digest" {
			t.Fatalf("expected core.interface_invalid_digest, got %q (%v)", code, err)
		}
	})
	t.Run("duplicate function ID", func(t *testing.T) {
		document := deepCopyJSON(t, validInterfaceV1Document(t))
		functions := document["functions"].([]any)
		duplicate := deepCopyJSON(t, functions[0].(map[string]any))
		document["functions"] = append(functions, duplicate)
		data, _ := json.Marshal(document)
		_, err := core.DecodeInterface(data)
		if code := decodeErrorCode(err); code != "core.interface_duplicate_function_id" {
			t.Fatalf("expected core.interface_duplicate_function_id, got %q (%v)", code, err)
		}
	})
	t.Run("unknown schema", func(t *testing.T) {
		document := deepCopyJSON(t, validInterfaceV1Document(t))
		document["schema"] = "lang.interface/2"
		data, _ := json.Marshal(document)
		_, err := core.DecodeInterface(data)
		if code := decodeErrorCode(err); code != "core.interface_unknown_schema" {
			t.Fatalf("expected core.interface_unknown_schema, got %q (%v)", code, err)
		}
	})
}

// TestPhase7DispatchControlsMutationKilled is Task 3's D-07-41 mutation-kill
// suite (QLT-08): each of Phase 07's exhaustive-dispatch controls is
// observed to FAIL under a seeded mutation, in the plan that introduces it,
// proving the control is load-bearing rather than merely present. See
// 07-04-PLAN.md's Task 3 <behavior> for the six numbered cases this test
// covers; the session-package half of Test 2 (the phase07 CLI-observable
// lane's own required-kinds seam) lives in session_phase7_mutation_test.go
// alongside its own package's fault-injection seam (D-07-42), and the
// interp/cgen recognized-not-executed kills (Tests 3-4) live in
// internal/compiler/interp/interp_test.go and
// internal/compiler/cgen/cgen_test.go respectively -- each control's kill
// lives next to the control it kills, mirroring D-07-42's "same-package
// tests" discipline rather than collecting every kill into one file no
// package actually owns.
func TestPhase7DispatchControlsMutationKilled(t *testing.T) {
	// Test 1 (in-process control, T-07-22): removing core.OpCall from the
	// required-kinds list runExhaustiveDispatchControl's own bottom loop
	// checks proves that loop -- not the per-fixture site calls above it --
	// is what makes the control load-bearing. opCallFreeFixtures is a real,
	// legitimate Phase 1/2 sub-corpus that produces no core.OpCall at all.
	t.Run("in_process_required_kinds_seam", func(t *testing.T) {
		opCallFreeFixtures := []string{
			"testdata/phase1/toggle.lang",
			"testdata/phase2/implicit_copy.lang",
		}
		// Clean: core.OpCall IS required, but this corpus never produces
		// it -- the control MUST fail.
		if err := runExhaustiveDispatchControl(opCallFreeFixtures, []core.OperationKind{core.OpCall}); err == nil {
			t.Fatal("expected the control to fail: core.OpCall is required but no fixture in this OpCall-free corpus produces it")
		}
		// Mutated: remove core.OpCall from the required-kinds list -- the
		// SAME OpCall-free corpus now PASSES, because the loop that would
		// have caught the omission no longer checks for it. This is the
		// observable effect that proves the loop, not the per-fixture
		// site calls, is what T-07-22 depends on.
		if err := runExhaustiveDispatchControl(opCallFreeFixtures, []core.OperationKind{}); err != nil {
			t.Fatalf("expected the control to pass once core.OpCall is excluded from required kinds, got: %v", err)
		}
	})

	// Test 5: with linearProbeInput's Byte/Buffer recognition unavailable
	// (stood in here by a probe that always returns ok == false, exactly
	// what an unrecognized parameter type produces), the interp site would
	// be SILENTLY SKIPPED for call_basic.lang's functions -- and this test
	// asserts the skip is DETECTED (both functions counted as skipped),
	// never silently tolerated, and that the REAL linearProbeInput
	// exercises exactly those same functions instead of skipping them.
	t.Run("linear_probe_input_arm_removed_is_detected", func(t *testing.T) {
		probeInputArmRemoved := func(core.Function) (string, bool) { return "", false }
		source, err := os.ReadFile(testsupport.ProjectPath(splitPath("testdata/phase07/call_basic.lang")...))
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		checked := session.Check(source)
		if len(checked.Diagnostics) > 0 {
			t.Fatalf("unexpected diagnostics: %v", checked.Diagnostics)
		}
		validated := corevalidate.Validate(checked.Program)
		if !validated.Valid {
			t.Fatalf("corevalidate rejected: %v", validated.Problems)
		}
		program := validated.Program()
		if len(program.Functions) != 2 {
			t.Fatalf("call_basic.lang: expected exactly 2 functions, got %d", len(program.Functions))
		}

		skipped := 0
		for _, function := range program.Functions {
			if _, ok := probeInputArmRemoved(function); !ok {
				skipped++
			}
		}
		if skipped != len(program.Functions) {
			t.Fatalf("expected the arm-removed stand-in to skip all %d functions, skipped %d", len(program.Functions), skipped)
		}

		exercised := 0
		for _, function := range program.Functions {
			if _, ok := linearProbeInput(function); ok {
				exercised++
			}
		}
		if exercised != len(program.Functions) {
			t.Fatalf("expected the real linearProbeInput to exercise all %d functions the stand-in skipped, exercised %d", len(program.Functions), exercised)
		}
	})
}
