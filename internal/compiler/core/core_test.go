package core_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/evidence"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/pathoracle"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
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
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase17", "return_type_tracer.schway"))
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
	if classify.ID == "" || main.ID == "" || classify.Match == nil || classify.Linear != nil || main.Linear == nil {
		t.Fatalf("canonical tracer functions missing or malformed: %+v", checked.Program.Functions)
	}
	if classify.Parameter.Type != "Resource" || classify.ReturnType != "Result" || main.Parameter.Type != "Resource" || main.ReturnType != "Result" {
		t.Fatalf("tracer declarations collapsed: classify=%+v main=%+v", classify, main)
	}
	if len(classify.Match.Arms) != 1 || classify.Match.Arms[0].Pattern != "Raw" || classify.Match.Arms[0].Value != "Classified" {
		t.Fatalf("classify directional alternatives malformed: %+v", classify.Match.Arms)
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

func TestPhase19OpConstShape(t *testing.T) {
	field, ok := reflect.TypeOf(core.LinearOperation{}).FieldByName("ConstU64")
	if !ok {
		t.Fatal("LinearOperation has no ConstU64 semantic payload")
	}
	if field.Type.Kind() != reflect.String || field.Tag.Get("json") != "const_u64,omitempty" {
		t.Fatalf("ConstU64 field has type/tag %s/%q, want string/const_u64,omitempty", field.Type, field.Tag.Get("json"))
	}
	operation := core.LinearOperation{ID: "f:op:0", PointID: "f:point:linear:0", Kind: core.OperationKind("const"), TargetID: "f:place:1", TypeID: "f:type:0"}
	reflect.ValueOf(&operation).Elem().FieldByIndex(field.Index).SetString("0")
	encoded, err := json.Marshal(operation)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["const_u64"] != "0" {
		t.Fatalf("OpConst zero payload serialized as %v", decoded["const_u64"])
	}
	legacy, err := json.Marshal(core.LinearOperation{ID: "f:op:1", PointID: "f:point:linear:1", Kind: core.OpReturn, SourceID: "f:place:0", TypeID: "f:type:0"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(legacy), "const_u64") {
		t.Fatalf("unused constant field changed legacy operation JSON: %s", legacy)
	}
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
	{"testdata/phase1/comments.schway", "1a90f92b261dbc825c3e480bd53465c10706e36a357f2491b0e0c88894166818", "evidence:f4570dd58a26d7c7582901a3"},
	{"testdata/phase1/toggle.schway", "5fc207e1a572c3a9e6ef04aa85010e1b842782e92635ae3f09949a3fb1353658", "evidence:142a7ab526b8f0cbd2cf4f26"},
	{"testdata/phase2/implicit_copy.schway", "5e2010f1331206d4610ca6a1b10ce686092a6884a8d4fc014cac3b44e4c36b2a", "evidence:6f46d2e225598990faccbd0b"},
	{"testdata/phase2/owned_transfer.schway", "e17fe549a50beb20c99e0172856d5df5c1736e97dce76d79c2d693c17d22fd20", "evidence:fb5bfdc8790a8f9c175ea872"},
	{"testdata/phase3/borrowed_view.schway", "696bf2e1c2ec73411ed8b091d753bef6e6504026cc43dc3b248ae62febba438f", "evidence:dddd4a0b3097658e6398e1cf"},
	{"testdata/phase3/branch_one_arm_shared_accept.schway", "c7ebef76cd22b510aec1a8509bcc89c78883f13dac4f1b94d0e217238a094118", "evidence:da42f3641c66baf35e645c66"},
	{"testdata/phase3/branch_view.schway", "3ec55ec5717e6760e5d3d006c2a5020f953a21fbf41e0adf406132e202365d9c", "evidence:214eeb3db43ca09b07561834"},
	{"testdata/phase3/public_view.schway", "b2b6ba18f5fc89ccbf983e15767ed71d27a9c9cfae907edf2e6cf09dac15452f", "evidence:9e8559ebc60ee839634e5294"},
	{"testdata/phase3/public_view_impossible.schway", "331a5b2d81849cfbfc31ba9430ff988d060874386bc93ad574f9429a516c493b", "evidence:1c0473336df43d779c6df70e"},
	{"testdata/phase3/public_view_mixed_access.schway", "b2665aa3d0cc9ed49d542a923878b4936e378c870c229638d0549b4feefa9e52", "evidence:0c125c8dc25ffe83c6a50de4"},
	{"testdata/phase3/public_view_multi_arm_access_conflict.schway", "e35c719e6b1ef7c246cb9635c1a8f656dbde3c6df3ed19d8211ea70be3758a6f", "evidence:6b624856c3bc19fa6f197172"},
	{"testdata/phase3/public_view_multi_arm_omitted.schway", "5c323c009c0d06132167fbe1c72be37cf3e6710a465159c6980544915a5a8ebd", "evidence:dee7242791bb8aa48524f84b"},
	{"testdata/phase3/public_view_omitted.schway", "f09350cfab25d9e527c6feee9f8c2fc89b22e8eaf530d016cf8421151c0c2cfc", "evidence:a038be9d5fa9c1d3dba6ba01"},
	{"testdata/phase3/public_view_understated.schway", "b45496eb4292454be7a919c1f63a7768060f383a1dc61f3f0a77f7aba65d8574", "evidence:40f4a643baf24f1a75d0422c"},
	{"testdata/phase3/sequential_shared_then_exclusive_accept.schway", "56d20794f8a43822f4483d2da39e870b4cb036c262f5e9289d32f5859bed20bd", "evidence:11dff9f506141504f46b90e0"},
	{"testdata/phase3/shared_shared_accept.schway", "ecc14ac0c851b90546fb41502b8915ac325ba322d3a3a7d4887b86961ed231b4", "evidence:b12a60b311e5cbad4e0274bf"},
	// Phase 4 accepting fixtures (D-05-39): widened from Phase 1-3 so this
	// pin also catches a Phase 5 emitter change that silently perturbs
	// Phase 4. Only the accepting testdata/phase4 fixtures are pinned here,
	// matching the Phase 3 precedent above (reject fixtures produce
	// diagnostics and are out of scope for this byte-identity pin); the set
	// mirrors native_test.go's phase4CorpusMatrix() "clean(...)" entries.
	{"testdata/phase4/acquire_three_fail_second.schway", "fcfd88bc97a15bdd0c774148e8efb23b2210d3f6e40ad101d0b3cfcf33591dee", "evidence:db391e8cbf6a999c40fbdaf4"},
	{"testdata/phase4/acquire_three_fail_third.schway", "1a23c824283bd12341da16197690f611884c8ef3b465458e134bfacd8dc7eb06", "evidence:1fb10eea4a10a005462dfaf6"},
	{"testdata/phase4/acquire_three_success.schway", "8bbce39409d78a99c55c02053deff4ee016dfbf9d03733ad6cd8a5e0362ec88a", "evidence:d68787f08a68a17f486d1783"},
	{"testdata/phase4/defect_terminal.schway", "4f348119f72c9d8aa1b3dc1cb91942176d74727297142ca940d44c08737bab0f", "evidence:12e9d68073ab86eb9eb463c9"},
	{"testdata/phase4/discard_because.schway", "6b1b048e4e0b2d3cc2791886d7b54f31de63788b4e2c5a8b185df51e5f652d89", "evidence:ac3003a93f4568b404aa319a"},
	{"testdata/phase4/foreign_acquire_one.schway", "718bed114e0754d3bfdb36a08d10649672915644d2eaca5caf266072f7e4dbf1", "evidence:dbfea02f20d97aff459e458b"},
	{"testdata/phase4/nonlocal_exit_probe.schway", "cde5fabf98be29972f21c4331bacde11fc135f1e3981c5ed495e23d554af0eb8", "evidence:0c768d94aeee2d03625e613b"},
	// Phase 5 accepting fixtures (D-06-31/D-06-32): widened from Phase 1-4 so
	// this pin also catches a Phase 6 schway.command/schway.verify-lane bump that
	// silently perturbs a Phase 5 program. Only the accepting testdata/phase5
	// fixtures are pinned here, matching the Phase 3/4 precedent above.
	// coordinated_lie.schway (and its coordinated_lie.core.json) is the
	// declared expected-escape pair (escape:coordinated-source-to-core-false-claim)
	// and is deliberately excluded from this table.
	{"testdata/phase5/allocator_mismatch.schway", "636097c74c3f161f532284bbaf2d6567e53413a7a9652f76c29da223dd0222c2", "evidence:1cdac22aba40cacc0f5e0001"},
	{"testdata/phase5/dead_store_unused_acquire.schway", "22a734be2932d02ecc47ced7c778f05e3d56296d4bfb9f3a7f4c1e543738fd1c", "evidence:f454a99f26276fb1545ce790"},
	{"testdata/phase5/defect_dies_by_signal.schway", "488f2f4dcd596626b6b82ddd9a2857c0068a53a37485c7e90104957b2607761a", "evidence:c9eea4002b7177106948da43"},
	{"testdata/phase5/false_restrict_hoist.schway", "2e2deae3e230984bf1430bb2d4c347d172444e22ad68f71c4d1197e34766f791", "evidence:ca3e057326af5276f18f5393"},
	{"testdata/phase5/inline_across_foreign.schway", "02fd41768398d0c650462790f15162f02c0eb5a0fc07bb97d15078d2a11cb53b", "evidence:2a9644fce2f925ca50841428"},
	{"testdata/phase5/reorder_two_events.schway", "5bc35a467aadffee60cb1ad7978ed17bf54f5f372c5f68d2ea9ebf81bf0ae56d", "evidence:1d7d61d8028a9472ee203b54"},
	{"testdata/phase5/restrict_borrow.schway", "15398f69e1d647b768f361cb0bedfc6064fee3f5fd5a00e03e88d573b8d96710", "evidence:65b0b4195299987d10f0ac80"},
	{"testdata/phase5/retained_pointer.schway", "7bac4e9375cbb0cfe1cae1eed15ba6278589220d21095018b78ac9081fa69297", "evidence:4d70b9513ffa2ed821dc26ad"},
	{"testdata/phase5/tail_collapse_release_ladder.schway", "f280d9999f29812956e1bec639aaed801342c60ae8706446852e25180c7f2594", "evidence:23a9ce0dc643fb05498e1963"},
	{"testdata/phase5/typed_failure_truncated_stdout.schway", "b146e3cd1f157d393fff9eecae8d5520db826ba6cbf635f4f1ef36ffd238886c", "evidence:598f270123ab0a123a99d351"},
}

// Phase 4/5's foreign layout core value names the C type the frozen private
// header declares. The public rename changed this one field from the old
// Lang spelling to the Schway spelling; keep the original source/core pins
// above and admit only these witnessed current hashes.
const (
	phase16OldForeignTypeName = "lang" + "_foreign_resource_block"
	phase16NewForeignTypeName = "schway_foreign_resource_block"
)

var phase16SchwayForeignTypeCoreSHA256 = map[string]string{
	"testdata/phase4/acquire_three_fail_second.schway":      "a4b13c10d60b35ff60251cb95b92a0064b783983d68e766969430876893da8fe",
	"testdata/phase4/acquire_three_fail_third.schway":       "34313217e6de96c72506cb4c286fa96568493b19bf96a4a5efafb926cb23e0b4",
	"testdata/phase4/acquire_three_success.schway":          "9c7d6282fb5364ac79538c3da4fbe6a966b2280a0c2cade09de393eb707b8ab2",
	"testdata/phase4/discard_because.schway":                "fb206a1c820a26d8bc5628e9d2b902afb4df0170744cd63f41869d2d25a49599",
	"testdata/phase4/foreign_acquire_one.schway":            "011e1d90e47628c6473cde1fdcde7e3948ce6170b7bb54f2e98f54f5464e5a84",
	"testdata/phase4/nonlocal_exit_probe.schway":            "448e34fe028ec35218b3207d57091d285badf659a77c73fb73e372df20ce871e",
	"testdata/phase5/allocator_mismatch.schway":             "d983488512ef2b9b399fb525bc2d8343d18f89799aa91143beaf6fe79bc9b533",
	"testdata/phase5/dead_store_unused_acquire.schway":      "31440397a59feb2a706434de33842c7aefa240a3d04499a31e9bed4a9a18d200",
	"testdata/phase5/inline_across_foreign.schway":          "72fef86428dade289e837fd51dc9294bc9294c99d11f4c7f7b5f2ddf64943229",
	"testdata/phase5/reorder_two_events.schway":             "2be6e33c8685cb2f86c87e81427ef5da94d2fb528de0dae33ac10cc60e21532d",
	"testdata/phase5/retained_pointer.schway":               "7fc95194fce966b97d59f21bcbf15668cffe33f9c307e21fe6d2b5e8225920d6",
	"testdata/phase5/tail_collapse_release_ladder.schway":   "9be0c6179ffe348772a54802281a4988d343dc000e728455582a28b04fe45f33",
	"testdata/phase5/typed_failure_truncated_stdout.schway": "0d608061478bf8773433b0011a3522b0eb2e8b7b3c2eaa1730d8ebe6c6495319",
}

// These are the current Schway-identity manifest IDs for every pinned fixture
// that remains admitted through evidence.Build. The corresponding original
// IDs stay in pinnedFixtures; identity normalization must reproduce those
// pins except where a separate behavior-migration ledger records an additional
// witnessed change to terminal output before abort.
var phase16SchwayIdentityManifestIDs = map[string]string{
	"testdata/phase1/comments.schway":                                "evidence:09e9b18d9d45df83f0fec99c",
	"testdata/phase1/toggle.schway":                                  "evidence:9714adb24624f30bad81c68f",
	"testdata/phase2/implicit_copy.schway":                           "evidence:43eb81e7793a1f0321196bc8",
	"testdata/phase2/owned_transfer.schway":                          "evidence:2f50b1e83911ed4dd24edac5",
	"testdata/phase3/borrowed_view.schway":                           "evidence:4d79eadfe1c5ed40d2b83833",
	"testdata/phase3/branch_one_arm_shared_accept.schway":            "evidence:9a943143d07d26e35bdb23b3",
	"testdata/phase3/branch_view.schway":                             "evidence:7a406e8822507dcd0aff95fe",
	"testdata/phase3/public_view.schway":                             "evidence:3e82c1d1a217ac6f185cfd32",
	"testdata/phase3/public_view_impossible.schway":                  "evidence:c73e2c6e4cfb1d27e4868168",
	"testdata/phase3/public_view_mixed_access.schway":                "evidence:95ac555949a6fcd682b75fa1",
	"testdata/phase3/public_view_multi_arm_access_conflict.schway":   "evidence:6796f981de49c6beb4f83de7",
	"testdata/phase3/public_view_multi_arm_omitted.schway":           "evidence:84b9d61990cf24fe3b43acbd",
	"testdata/phase3/public_view_omitted.schway":                     "evidence:4990a98f096ec4609b0f2d86",
	"testdata/phase3/public_view_understated.schway":                 "evidence:737d5a21c90859b7f71a22c1",
	"testdata/phase3/sequential_shared_then_exclusive_accept.schway": "evidence:2371764fac483a4eee472595",
	"testdata/phase3/shared_shared_accept.schway":                    "evidence:ee4aa09030f857f5f93dc9c5",
	"testdata/phase4/defect_terminal.schway":                         "evidence:7ba48922b4cc04a392125649",
	"testdata/phase5/defect_dies_by_signal.schway":                   "evidence:f63c628eabe74b0f9d361ef1",
}

type phase16ManifestBehaviorMigration struct {
	OriginalManifestID           string
	IdentityNormalizedManifestID string
	CurrentSchwayManifestID      string
	ChangedBehavior              string
	Witness                      string
}

var phase16ManifestBehaviorMigrations = map[string]phase16ManifestBehaviorMigration{
	"testdata/phase4/defect_terminal.schway": {
		OriginalManifestID:           "evidence:12e9d68073ab86eb9eb463c9",
		IdentityNormalizedManifestID: "evidence:7649e234461dab67e77861c7",
		CurrentSchwayManifestID:      "evidence:7ba48922b4cc04a392125649",
		ChangedBehavior:              "flush the terminal JSON record before abort",
		Witness:                      "probe:TestProgramMatchDefectEventPrecedesAbort",
	},
	"testdata/phase5/defect_dies_by_signal.schway": {
		OriginalManifestID:           "evidence:c9eea4002b7177106948da43",
		IdentityNormalizedManifestID: "evidence:b563764272b40d6770f41273",
		CurrentSchwayManifestID:      "evidence:f63c628eabe74b0f9d361ef1",
		ChangedBehavior:              "flush the terminal JSON record before abort",
		Witness:                      "probe:TestProgramMatchDefectEventPrecedesAbort",
	},
}

// TestPreviousPhaseCoreBytesUnchanged pins every Phase 1-5 fixture's
// serialized core JSON to its exact byte value from before Phase 6
// (D-06-31/D-06-32, widened from the Phase 5 pin which stopped at Phase 4).
// It must be green before the coordinated schway.command and schway.verify-lane
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
			wantCurrent, migrated := phase16SchwayForeignTypeCoreSHA256[fixture.Path]
			if !migrated {
				if got != fixture.CoreSHA256 {
					t.Fatalf("core bytes moved for %s: got sha256 %s, want %s", fixture.Path, got, fixture.CoreSHA256)
				}
				return
			}
			if got != wantCurrent {
				t.Fatalf("core bytes moved outside the recorded foreign-type identity migration for %s: got sha256 %s, want old %s or current %s", fixture.Path, got, fixture.CoreSHA256, wantCurrent)
			}
			normalized, replacements, err := phase16NormalizeForeignTypeName(coreBytes)
			if err != nil {
				t.Fatalf("normalize foreign type identity: %v", err)
			}
			if replacements != 1 || phase16SHA256(normalized) != fixture.CoreSHA256 {
				t.Fatalf("only the recorded foreign_type_name field may differ for %s: replacements=%d normalized_sha256=%s want=%s", fixture.Path, replacements, phase16SHA256(normalized), fixture.CoreSHA256)
			}
		})
	}
	for path := range phase16SchwayForeignTypeCoreSHA256 {
		if !containsPinnedFixture(path) {
			t.Fatalf("current core identity migration is not tied to a historical pin: %s", path)
		}
	}
}

func containsPinnedFixture(path string) bool {
	for _, fixture := range pinnedFixtures {
		if fixture.Path == path {
			return true
		}
	}
	return false
}

func phase16NormalizeForeignTypeName(coreBytes []byte) ([]byte, int, error) {
	var program core.Program
	if err := json.Unmarshal(coreBytes, &program); err != nil {
		return nil, 0, err
	}
	replacements := 0
	for index := range program.Functions {
		contract := program.Functions[index].ForeignContract
		if contract == nil || contract.Layout == nil || contract.Layout.ForeignTypeName != phase16NewForeignTypeName {
			continue
		}
		contract.Layout.ForeignTypeName = phase16OldForeignTypeName
		replacements++
	}
	normalized, err := json.Marshal(program)
	return normalized, replacements, err
}

func phase16SHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func normalizeSchwayIdentityForHistoricalComparison(cSource string) string {
	return strings.NewReplacer(
		"generated by Schway;", "generated by Codename Lang;",
		"SCHWAY", "LANG",
		"schway", "lang",
		"Schway", "Lang",
	).Replace(cSource)
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
// D-06-31/D-06-32. It must be green before the coordinated schway.command and
// schway.verify-lane bump lands.
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
			currentID, migrated := phase16SchwayIdentityManifestIDs[fixture.Path]
			if migrated {
				if product.Manifest.ID != currentID {
					t.Fatalf("current Schway manifest ID moved for %s: got %s, want %s", fixture.Path, product.Manifest.ID, currentID)
				}
				historicalID, normalizeErr := phase16HistoricalManifestID(product)
				wantHistoricalID := fixture.ManifestID
				if migration, behaviorChanged := phase16ManifestBehaviorMigrations[fixture.Path]; behaviorChanged {
					if migration.OriginalManifestID != fixture.ManifestID || migration.CurrentSchwayManifestID != currentID || migration.ChangedBehavior != "flush the terminal JSON record before abort" || migration.Witness != "probe:TestProgramMatchDefectEventPrecedesAbort" {
						t.Fatalf("terminal-output manifest migration is not tied to the original/current pins for %s", fixture.Path)
					}
					witness, witnessErr := os.ReadFile(testsupport.ProjectPath("internal", "compiler", "cgen", "cgen_program_test.go"))
					generator, generatorErr := os.ReadFile(testsupport.ProjectPath("internal", "compiler", "cgen", "cgen_program.go"))
					if witnessErr != nil || generatorErr != nil || !strings.Contains(string(witness), "func TestProgramMatchDefectEventPrecedesAbort") || !strings.Contains(string(generator), "fflush(stdout)") {
						t.Fatalf("terminal-output behavior witness is not live for %s", fixture.Path)
					}
					wantHistoricalID = migration.IdentityNormalizedManifestID
				}
				if normalizeErr != nil || historicalID != wantHistoricalID {
					t.Fatalf("identity-normalized manifest ID for %s: got %s err=%v want %s", fixture.Path, historicalID, normalizeErr, wantHistoricalID)
				}
				return
			}
			if product.Manifest.ID != fixture.ManifestID {
				t.Fatalf("manifest ID moved for %s: got %s, want %s", fixture.Path, product.Manifest.ID, fixture.ManifestID)
			}
		})
	}
	for path := range phase16SchwayIdentityManifestIDs {
		if !containsPinnedFixture(path) {
			t.Fatalf("current manifest identity migration is not tied to a historical pin: %s", path)
		}
	}
	for path := range phase16ManifestBehaviorMigrations {
		if !containsPinnedFixture(path) {
			t.Fatalf("manifest behavior migration is not tied to a historical pin: %s", path)
		}
		if _, currentIdentity := phase16SchwayIdentityManifestIDs[path]; !currentIdentity {
			t.Fatalf("manifest behavior migration has no current Schway identity pin: %s", path)
		}
	}
}

func phase16HistoricalManifestID(product evidence.Product) (string, error) {
	normalizedCore, _, err := phase16NormalizeForeignTypeName(product.CoreBytes)
	if err != nil {
		return "", err
	}
	var historicalProgram core.Program
	if err := json.Unmarshal(normalizedCore, &historicalProgram); err != nil {
		return "", err
	}
	historicalC := normalizeSchwayIdentityForHistoricalComparison(string(product.CSource))
	manifest := product.Manifest
	manifest.CoreDigest = "sha256:" + phase16SHA256(normalizedCore)
	manifest.CDigest = "sha256:" + phase16SHA256([]byte(historicalC))
	if manifest.ForeignDigest != "" {
		foreignManifest, err := cgen.EmitForeignManifest(historicalProgram)
		if err != nil {
			return "", err
		}
		manifest.ForeignDigest = "sha256:" + phase16SHA256([]byte(normalizeSchwayIdentityForHistoricalComparison(foreignManifest)))
	}
	encoded, err := evidence.CanonicalBytes(manifest)
	if err != nil {
		return "", err
	}
	historical, err := evidence.DecodeStrict(encoded)
	if err != nil {
		return "", err
	}
	return historical.ID, nil
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

// previousPhaseGoldenCDigests pins the four committed Phase 1/2/4/5 C
// goldens at their current public Schway identity. Earlier emitter and
// pre-Schway digests stay in the separate ledgers below, where identity-only
// normalization must reproduce the prior bytes exactly.
var previousPhaseGoldenCDigests = map[string]string{
	"testdata/phase1/generated.golden.c":               "303a48995217b96495505d2bf90d3ed26d0ce723f42a794748dc64b1ac9412df",
	"testdata/phase2/owned_transfer.golden.c":          "332d7f6336c78feec2d8912a8b88c768351df201162864a3a89c6227dd847740",
	"testdata/phase4/foreign_layout_mismatch.golden.c": "0df9b654b20dee616ec77f09d23fee33e7a76ff89249cdbc648b9eb4ab261be8",
	"testdata/phase5/restrict_borrow.golden.c":         "d7d24696ad378fc9432a95095e52f21fe685120b7b4fecd2c271f89a93bd5418",
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
	{Path: "testdata/phase1/generated.golden.c", State: phase16GoldenPostCut, OldSHA256: "f3e4fa6b641112fc8d213d04a38fce83dcfe0cd37ffbd79bc833ee787f11dc74", NewSHA256: "303a48995217b96495505d2bf90d3ed26d0ce723f42a794748dc64b1ac9412df", MovedResponsibility: "legacy N=1 emitter to emitProgram", StructuralReason: "public match dispatch now emits the schema-2 program document", SemanticWitness: "TestN1ConvergenceDifferential/phase1/toggle.schway", N1Fixture: "testdata/phase1/toggle.schway", ReviewDisposition: "post-cut public/direct byte identity"},
	{Path: "testdata/phase2/owned_transfer.golden.c", State: phase16GoldenPostCut, OldSHA256: "91177543f89174fba680c70e79147d5ffc69714adefc8404f8de8dbdcdac65b8", NewSHA256: "332d7f6336c78feec2d8912a8b88c768351df201162864a3a89c6227dd847740", MovedResponsibility: "legacy N=1 emitter to emitProgram", StructuralReason: "public linear dispatch now emits the schema-2 program document", SemanticWitness: "TestN1ConvergenceDifferential/phase2/owned_transfer.schway", N1Fixture: "testdata/phase2/owned_transfer.schway", ReviewDisposition: "post-cut public/direct byte identity"},
	{Path: "testdata/phase4/foreign_layout_mismatch.golden.c", State: phase16GoldenPostCut, OldSHA256: "3be6ebc36032ac9cc29bb916c1cdb8a8a996c0028ddf6982546f4c3dd5ffd031", NewSHA256: "0df9b654b20dee616ec77f09d23fee33e7a76ff89249cdbc648b9eb4ab261be8", MovedResponsibility: "foreign lowering remains outside emitProgram", StructuralReason: "foreign lowering is explicit cut-M004 debt and not an admitted program shape", SemanticWitness: "TestProgramBranchValidationOrder/foreign_shape_precedes_preflight", N1Fixture: "testdata/phase4/foreign_layout_mismatch.schway", ReviewDisposition: "frozen cut-family baseline; refusal was not regenerated"},
	{Path: "testdata/phase5/restrict_borrow.golden.c", State: phase16GoldenPostCut, OldSHA256: "05a16af7e57c3a1a1e2b9af1eb4bed689d89fa53ff91e51328d51dd6f64e38f0", NewSHA256: "d7d24696ad378fc9432a95095e52f21fe685120b7b4fecd2c271f89a93bd5418", MovedResponsibility: "by-pointer lowering remains outside emitProgram", StructuralReason: "cut-M004 excludes every by-pointer family from program admission", SemanticWitness: "TestProgramBorrowedByPointerDisposition", N1Fixture: "testdata/phase5/restrict_borrow.schway", ReviewDisposition: "frozen cut-family baseline; refusal was not regenerated"},
}

type phase16GoldenIdentityMigration struct {
	Path            string
	OldSHA256       string
	NewSHA256       string
	ChangedInput    string
	SemanticWitness string
}

// This second ledger records only the current public identity transformation
// applied after the emitter responsibility cut. Its witness reverses those
// identity strings and requires the exact pre-Schway post-cut bytes.
var phase16SchwayGoldenIdentityLedger = []phase16GoldenIdentityMigration{
	{Path: "testdata/phase1/generated.golden.c", OldSHA256: "1fd8aff8ee28de7ec39e559a7ca9ce50e480ecfffede617c36b2282c60cc122a", NewSHA256: "303a48995217b96495505d2bf90d3ed26d0ce723f42a794748dc64b1ac9412df", ChangedInput: "current C ABI namespace and generated-by brand", SemanticWitness: "TestExistingEmittersAreByteIdentical"},
	{Path: "testdata/phase2/owned_transfer.golden.c", OldSHA256: "f324f24db3ca0dfa8006b5c7fbec4263a6daf2dcbe220d49ea19167920799686", NewSHA256: "332d7f6336c78feec2d8912a8b88c768351df201162864a3a89c6227dd847740", ChangedInput: "current C ABI namespace and generated-by brand", SemanticWitness: "TestExistingEmittersAreByteIdentical"},
	{Path: "testdata/phase4/foreign_layout_mismatch.golden.c", OldSHA256: "3be6ebc36032ac9cc29bb916c1cdb8a8a996c0028ddf6982546f4c3dd5ffd031", NewSHA256: "0df9b654b20dee616ec77f09d23fee33e7a76ff89249cdbc648b9eb4ab261be8", ChangedInput: "current C ABI namespace and generated-by brand", SemanticWitness: "TestPhase4CorpusThreeEngineAgreement"},
	{Path: "testdata/phase5/restrict_borrow.golden.c", OldSHA256: "05a16af7e57c3a1a1e2b9af1eb4bed689d89fa53ff91e51328d51dd6f64e38f0", NewSHA256: "d7d24696ad378fc9432a95095e52f21fe685120b7b4fecd2c271f89a93bd5418", ChangedInput: "current C ABI namespace and generated-by brand", SemanticWitness: "TestPhase5ByPointerLoweringGolden"},
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
	if problems := phase16GoldenIdentityMigrationProblems(previousPhaseGoldenCDigests, phase16SchwayGoldenIdentityLedger, phase16CurrentGoldenDigests(t)); len(problems) != 0 {
		t.Fatalf("Schway golden-C identity migration invalid:\n%s", strings.Join(problems, "\n"))
	}
}

func phase16GoldenIdentityMigrationProblems(digests map[string]string, ledger []phase16GoldenIdentityMigration, current map[string]string) []string {
	var problems []string
	if len(ledger) != 4 {
		problems = append(problems, fmt.Sprintf("identity migration count=%d, want 4", len(ledger)))
	}
	seen := make(map[string]bool, len(ledger))
	for _, entry := range ledger {
		if seen[entry.Path] {
			problems = append(problems, "duplicate identity migration path "+entry.Path)
		}
		seen[entry.Path] = true
		if !isPhase16SHA256(entry.OldSHA256) || !isPhase16SHA256(entry.NewSHA256) {
			problems = append(problems, entry.Path+": malformed identity migration digest")
		}
		if entry.ChangedInput == "" || !strings.Contains(entry.SemanticWitness, "Test") {
			problems = append(problems, entry.Path+": missing changed input or semantic witness")
		}
		if digests[entry.Path] != entry.NewSHA256 || current[entry.Path] != entry.NewSHA256 {
			problems = append(problems, entry.Path+": current identity digest, map, and file disagree")
		}
		data, err := os.ReadFile(testsupport.ProjectPath(splitPath(entry.Path)...))
		if err != nil {
			problems = append(problems, entry.Path+": cannot read current file")
			continue
		}
		normalized := normalizeSchwayIdentityForHistoricalComparison(string(data))
		if phase16SHA256([]byte(normalized)) != entry.OldSHA256 {
			problems = append(problems, entry.Path+": identity normalization does not reproduce old bytes")
		}
	}
	for path := range digests {
		if !seen[path] {
			problems = append(problems, "digest map path absent from identity ledger "+path)
		}
	}
	return problems
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
// testdata/phase1 through testdata/phase5 and compares each against the
// current public-identity digest table. It fails on any drift and if the
// corpus gains or loses a golden.c file; TestPhase16GoldenChangeLedger also
// proves that reversing only the public identity strings recovers the prior
// pinned output bytes.
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
	const declaredCount = 13 // OpCopy, OpMove, OpBorrowShared, OpBorrowExclusive, OpReturn, OpForeignCall, OpFail, OpRelease, OpDefect, OpCall, OpConst, OpConstructPayload, OpDestructurePayload
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
	"testdata/phase1/toggle.schway",
	"testdata/phase1/comments.schway",
	"testdata/phase2/implicit_copy.schway",
	"testdata/phase2/owned_transfer.schway",
	"testdata/phase3/borrowed_view.schway",
	"testdata/phase3/branch_view.schway",
	"testdata/phase3/branch_one_arm_shared_accept.schway",
	"testdata/phase3/sequential_shared_then_exclusive_accept.schway",
	"testdata/phase3/shared_shared_accept.schway",
	"testdata/phase4/foreign_acquire_one.schway",
	"testdata/phase4/acquire_three_success.schway",
	"testdata/phase4/defect_terminal.schway",
	"testdata/phase07/call_basic.schway",
	"testdata/phase07/call_from_both_match_arms.schway",
	"testdata/phase12/payload_tracer.schway",
	"testdata/phase12/payload_drop_obligation.schway",
	"testdata/phase12/payload_borrow_interaction.schway",
	"testdata/phase19/literal_tracer.schway",
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
	phase19ConstObserved := false
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
					if path == "testdata/phase19/literal_tracer.schway" && operation.Kind == core.OpConst {
						if operation.SourceID != "" || operation.ConstU64 != "42" || operation.TargetID == "" {
							return fmt.Errorf("%s: OpConst root facts are not canonical: %+v", path, operation)
						}
						phase19ConstObserved = true
					}
				}
				// pathoracle site: must not error while walking.
				if function.Linear.ID != "" {
					endpoints, _, err := pathoracle.RecomputeEndpoints(function, nil)
					if err != nil {
						return fmt.Errorf("%s/%s: pathoracle error: %w", path, function.Name, err)
					}
					if path == "testdata/phase19/literal_tracer.schway" && len(endpoints) != 0 {
						return fmt.Errorf("%s/%s: constant root unexpectedly produced loan endpoints: %+v", path, function.Name, endpoints)
					}
				}
			}
			// originvalidate site: must not crash while walking.
			origins := originvalidate.RecomputeOriginPerReturn(function, calleeContracts)
			if path == "testdata/phase19/literal_tracer.schway" && function.Linear != nil && len(origins) != 1 {
				return fmt.Errorf("%s/%s: originvalidate returned %d origins for the constant return, want one", path, function.Name, len(origins))
			}
			if path == "testdata/phase19/literal_tracer.schway" && len(origins) == 1 && origins[0].Derived {
				return fmt.Errorf("%s/%s: originvalidate derived constant return from parameter: %+v", path, function.Name, origins[0])
			}

			// interp site.
			switch {
			case function.Match != nil:
				for _, arm := range function.Match.Arms {
					// Phase 10 (D-10-21/D-10-39) made core.OpCall a real,
					// executed operation, including from inside a match
					// arm's own block (e.g.
					// testdata/phase07/call_from_both_match_arms.schway) --
					// any interp error, including one from a call, fails
					// this control.
					if _, err := interp.Run(program, function.Name, arm.Pattern); err != nil {
						return fmt.Errorf("%s/%s/%s: interp error: %w", path, function.Name, arm.Pattern, err)
					}
				}
			case function.Linear != nil:
				input, ok := linearProbeInput(function)
				if ok {
					run, err := interp.Run(program, function.Name, input)
					if err != nil {
						return fmt.Errorf("%s/%s: interp error: %w", path, function.Name, err)
					}
					if path == "testdata/phase19/literal_tracer.schway" && run.Outcome.Value != "42" {
						return fmt.Errorf("%s/%s: interpreter returned %q, want 42", path, function.Name, run.Outcome.Value)
					}
				}
			}
		}

		// cgen site: Emit requires exactly one function.
		if len(program.Functions) == 1 {
			generated, err := cgen.Emit(program)
			if err != nil {
				if strings.Contains(err.Error(), "multi-function foreign-call bodies are not supported") || strings.Contains(err.Error(), "by-pointer bodies are not supported") {
					continue
				}
				return fmt.Errorf("%s: cgen error: %w", path, err)
			}
			if path == "testdata/phase19/literal_tracer.schway" && !strings.Contains(generated, "UINT64_C(42)") {
				return fmt.Errorf("%s: cgen output did not lower the encountered OpConst", path)
			}
		}
	}
	if containsFixture(fixtures, "testdata/phase19/literal_tracer.schway") && !phase19ConstObserved {
		return fmt.Errorf("testdata/phase19/literal_tracer.schway: no canonical OpConst was observed")
	}
	for _, kind := range requiredKinds {
		if !encountered[kind] {
			return fmt.Errorf("operation kind %q is never exercised by any corpus fixture in this control", kind)
		}
	}
	return nil
}

func containsFixture(fixtures []string, want string) bool {
	for _, fixture := range fixtures {
		if fixture == want {
			return true
		}
	}
	return false
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
// call_basic.schway's functions -- the caller (main) and the callee
// (identity) -- declare a Byte parameter, so linearProbeInput must return
// ok == true for both, driving interp.Run for each rather than skipping
// it.
func TestLinearProbeInputExercisesCallBasicFixture(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath(splitPath("testdata/phase07/call_basic.schway")...))
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
		t.Fatalf("call_basic.schway: expected exactly 2 functions, got %d", len(program.Functions))
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
		t.Fatalf("expected the interp site exercised for both of call_basic.schway's functions, got %d", exercised)
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
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", "foreign_acquire_one.schway"))
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
			"testdata/phase1/toggle.schway",
			"testdata/phase2/implicit_copy.schway",
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
	// be SILENTLY SKIPPED for call_basic.schway's functions -- and this test
	// asserts the skip is DETECTED (both functions counted as skipped),
	// never silently tolerated, and that the REAL linearProbeInput
	// exercises exactly those same functions instead of skipping them.
	t.Run("linear_probe_input_arm_removed_is_detected", func(t *testing.T) {
		probeInputArmRemoved := func(core.Function) (string, bool) { return "", false }
		source, err := os.ReadFile(testsupport.ProjectPath(splitPath("testdata/phase07/call_basic.schway")...))
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
			t.Fatalf("call_basic.schway: expected exactly 2 functions, got %d", len(program.Functions))
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
