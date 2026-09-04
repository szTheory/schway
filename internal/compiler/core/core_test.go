package core_test

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/evidence"
	"github.com/codename-lang/lang/internal/compiler/interp"
	"github.com/codename-lang/lang/internal/compiler/originvalidate"
	"github.com/codename-lang/lang/internal/compiler/pathoracle"
	"github.com/codename-lang/lang/internal/compiler/session"
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
	{"testdata/phase1/comments.lang", "1a90f92b261dbc825c3e480bd53465c10706e36a357f2491b0e0c88894166818", "evidence:a6a8335266e9b260158eca76"},
	{"testdata/phase1/toggle.lang", "5fc207e1a572c3a9e6ef04aa85010e1b842782e92635ae3f09949a3fb1353658", "evidence:82c85f837be952f23add7166"},
	{"testdata/phase2/implicit_copy.lang", "5e2010f1331206d4610ca6a1b10ce686092a6884a8d4fc014cac3b44e4c36b2a", "evidence:39b489f42ec5bf8b6ae62634"},
	{"testdata/phase2/owned_transfer.lang", "e17fe549a50beb20c99e0172856d5df5c1736e97dce76d79c2d693c17d22fd20", "evidence:4e6aa4b836352b204c2f8358"},
	{"testdata/phase3/borrowed_view.lang", "696bf2e1c2ec73411ed8b091d753bef6e6504026cc43dc3b248ae62febba438f", "evidence:ca43a139f8aba5de4bbdd86a"},
	{"testdata/phase3/branch_one_arm_shared_accept.lang", "c7ebef76cd22b510aec1a8509bcc89c78883f13dac4f1b94d0e217238a094118", "evidence:eed8b471bb58d9bdb12e531d"},
	{"testdata/phase3/branch_view.lang", "3ec55ec5717e6760e5d3d006c2a5020f953a21fbf41e0adf406132e202365d9c", "evidence:317410b3ff52becf5c8077b4"},
	{"testdata/phase3/public_view.lang", "b2b6ba18f5fc89ccbf983e15767ed71d27a9c9cfae907edf2e6cf09dac15452f", "evidence:31176129c6f33db690edb1ae"},
	{"testdata/phase3/public_view_impossible.lang", "331a5b2d81849cfbfc31ba9430ff988d060874386bc93ad574f9429a516c493b", "evidence:e6b6371c320b9c69b256561e"},
	{"testdata/phase3/public_view_mixed_access.lang", "b2665aa3d0cc9ed49d542a923878b4936e378c870c229638d0549b4feefa9e52", "evidence:df44ab20b04a2451111b7e4c"},
	{"testdata/phase3/public_view_multi_arm_access_conflict.lang", "e35c719e6b1ef7c246cb9635c1a8f656dbde3c6df3ed19d8211ea70be3758a6f", "evidence:01dc1f4dc6e12a522a94c5f7"},
	{"testdata/phase3/public_view_multi_arm_omitted.lang", "5c323c009c0d06132167fbe1c72be37cf3e6710a465159c6980544915a5a8ebd", "evidence:26ca6744e03c4bc94beb3352"},
	{"testdata/phase3/public_view_omitted.lang", "f09350cfab25d9e527c6feee9f8c2fc89b22e8eaf530d016cf8421151c0c2cfc", "evidence:78f1024b3afd153abc465359"},
	{"testdata/phase3/public_view_understated.lang", "b45496eb4292454be7a919c1f63a7768060f383a1dc61f3f0a77f7aba65d8574", "evidence:9b24e9af419804d5e7cb22ec"},
	{"testdata/phase3/sequential_shared_then_exclusive_accept.lang", "56d20794f8a43822f4483d2da39e870b4cb036c262f5e9289d32f5859bed20bd", "evidence:a7bf54fca28c4ddc52303c8f"},
	{"testdata/phase3/shared_shared_accept.lang", "ecc14ac0c851b90546fb41502b8915ac325ba322d3a3a7d4887b86961ed231b4", "evidence:00cb19b4caa2945e6ce59859"},
}

// TestPreviousPhaseCoreBytesUnchanged pins every Phase 1/2/3 fixture's
// serialized core JSON to its exact byte value at the Phase 4 phase-start
// commit (D-04-23). It must be green before any new operation kind lands.
func TestPreviousPhaseCoreBytesUnchanged(t *testing.T) {
	for _, fixture := range pinnedFixtures {
		t.Run(fixture.Path, func(t *testing.T) {
			source, err := os.ReadFile(testsupport.ProjectPath(splitPath(fixture.Path)...))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			product, diagnostics, err := evidence.Build(source, pinnedFacts)
			if err != nil {
				t.Fatalf("build: %v", err)
			}
			if len(diagnostics) > 0 {
				t.Fatalf("unexpected diagnostics: %v", diagnostics)
			}
			sum := sha256.Sum256(product.CoreBytes)
			got := hex.EncodeToString(sum[:])
			if got != fixture.CoreSHA256 {
				t.Fatalf("core bytes moved for %s: got sha256 %s, want %s", fixture.Path, got, fixture.CoreSHA256)
			}
		})
	}
}

// TestPreviousPhaseManifestIDsUnchanged is TestPreviousPhaseCoreBytesUnchanged's
// evidence-manifest-identity sibling (D-04-23).
func TestPreviousPhaseManifestIDsUnchanged(t *testing.T) {
	for _, fixture := range pinnedFixtures {
		t.Run(fixture.Path, func(t *testing.T) {
			source, err := os.ReadFile(testsupport.ProjectPath(splitPath(fixture.Path)...))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			product, diagnostics, err := evidence.Build(source, pinnedFacts)
			if err != nil {
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
	const declaredCount = 5 // OpCopy, OpMove, OpBorrowShared, OpBorrowExclusive, OpReturn
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
func TestAllOperationKindsHandledAtEverySite(t *testing.T) {
	fixtures := []string{
		"testdata/phase1/toggle.lang",
		"testdata/phase1/comments.lang",
		"testdata/phase2/implicit_copy.lang",
		"testdata/phase2/owned_transfer.lang",
		"testdata/phase3/borrowed_view.lang",
		"testdata/phase3/branch_view.lang",
		"testdata/phase3/branch_one_arm_shared_accept.lang",
		"testdata/phase3/sequential_shared_then_exclusive_accept.lang",
		"testdata/phase3/shared_shared_accept.lang",
	}
	encountered := make(map[core.OperationKind]bool)
	for _, path := range fixtures {
		source, err := os.ReadFile(testsupport.ProjectPath(splitPath(path)...))
		if err != nil {
			t.Fatalf("%s: read: %v", path, err)
		}
		checked := session.Check(source)
		if len(checked.Diagnostics) > 0 {
			t.Fatalf("%s: unexpected diagnostics: %v", path, checked.Diagnostics)
		}
		program := checked.Program

		// corevalidate site.
		validated := corevalidate.Validate(program)
		if !validated.Valid {
			t.Fatalf("%s: corevalidate rejected: %v", path, validated.Problems)
		}
		program = validated.Program()

		for _, function := range program.Functions {
			if function.Linear != nil {
				for _, operation := range function.Linear.Operations {
					encountered[operation.Kind] = true
				}
				// pathoracle site: must not error while walking.
				if function.Linear.ID != "" {
					if _, _, err := pathoracle.RecomputeEndpoints(function); err != nil {
						t.Fatalf("%s/%s: pathoracle error: %v", path, function.Name, err)
					}
				}
			}
			// originvalidate site: must not crash while walking.
			_ = originvalidate.RecomputeOriginPerReturn(function)

			// interp site.
			switch {
			case function.Match != nil:
				for _, arm := range function.Match.Arms {
					if _, err := interp.Run(program, function.Name, arm.Pattern); err != nil {
						t.Fatalf("%s/%s/%s: interp error: %v", path, function.Name, arm.Pattern, err)
					}
				}
			case function.Linear != nil:
				input, ok := linearProbeInput(function)
				if ok {
					if _, err := interp.Run(program, function.Name, input); err != nil {
						t.Fatalf("%s/%s: interp error: %v", path, function.Name, err)
					}
				}
			}
		}

		// cgen site: Emit requires exactly one function.
		if len(program.Functions) == 1 {
			if _, err := cgen.Emit(program); err != nil {
				t.Fatalf("%s: cgen error: %v", path, err)
			}
		}
	}
	for _, kind := range core.AllOperationKinds() {
		if !encountered[kind] {
			t.Fatalf("operation kind %q is never exercised by any corpus fixture in this control", kind)
		}
	}
}
