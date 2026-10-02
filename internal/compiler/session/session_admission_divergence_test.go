package session_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/protocol"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// This file mechanizes the divergence 11-MIDPHASE-GATE.md flagged "for human
// review": `schway check` reports status:invalid with core.origin_omitted for
// testdata/phase11/multi_function_gate_corpus.schway, while session.Check +
// corevalidate.Validate -- the admission path every Phase 5/6+ production
// run site, every cgen/gate test, and every verify-corpus lane actually uses
// -- accepts the identical program cleanly.
//
// The gate asked a human two questions: is this divergence Phase 11's fault,
// and does it need a tracked debt item? Both are decidable mechanically, and
// a one-time human answer would not stay true. So instead of answering them
// once in prose, this test pins the WHOLE corpus-wide divergence set and
// re-answers them on every run.
//
// What the sweep establishes, and what a reader should take from the table
// below: the divergence is neither new nor specific to Phase 11. Ten
// fixtures spanning phases 3, 4, 5, 7, 10 and 11 exhibit it, the earliest by
// several phases. It is the designed consequence of CheckCommandFile
// applying a THIRD admission layer that the internal path does not:
// originvalidate.ValidatePublished (session.go, D-04-27/WR-01), which gates
// PUBLICATION on a declared origin. Most entries below are negative fixtures
// that exist precisely to be rejected that way; the rest are positive
// fixtures for other lanes whose body shape necessarily derives a
// borrow-based return origin, which is what the publication gate is defined
// to flag.
//
// The test fails in BOTH directions. A new divergence -- the two admission
// surfaces drifting further apart -- fails. A listed divergence
// disappearing, or changing its diagnostic code, also fails, so that
// reconciling the two paths is a deliberate, reviewed edit to this table
// rather than a silent change nobody notices. That is the recurring value
// that justifies carrying this in CI (`go test ./...`, both hosts) instead
// of as a debt note: the note would be read once, this is checked forever.

// admissionDivergence is one fixture the CLI `check` command surface refuses
// but the internal admission path accepts, together with the diagnostic code
// the CLI answers with.
type admissionDivergence struct {
	fixture string
	code    string
	// why records whether the rejection is the publication gate doing its
	// intended job on a fixture built to be rejected (intended), or the
	// incidental consequence of a positive fixture whose shape happens to
	// derive an undeclared origin (incidental). Documentation only -- the
	// assertion below is on fixture and code.
	why string
}

// knownAdmissionDivergences is the committed, exhaustive table. Sorted by
// fixture path; keep it that way.
var knownAdmissionDivergences = []admissionDivergence{
	{"testdata/phase07/clean_but_unpublishable.schway", "core.origin_omitted", "intended: the fixture's own name states it -- it checks clean but is unpublishable"},
	{"testdata/phase10/compose_per_path_borrow_callee_accept.schway", "core.origin_omitted", "incidental: a positive compose-lane fixture whose callee derives a per-path borrow origin"},
	{"testdata/phase11/multi_function_gate_corpus.schway", "core.origin_omitted", "incidental: the fixture 11-MIDPHASE-GATE.md flagged; `touch` is modeled on phase5/restrict_borrow.schway below"},
	{"testdata/phase25/exclusive_escape_reject.schway", "core.origin_omitted", "intended: the exclusive successor escape fixture is admitted by the internal checker but refused by the CLI publication gate"},
	{"testdata/phase25/shared_copy_accept.schway", "core.origin_omitted", "incidental: a positive shared-copy fixture whose public borrowed return lacks a declared origin"},
	{"testdata/phase3/public_view_mixed_access.schway", "core.origin_access_mismatch", "intended: a declared-vs-derived access conflict fixture"},
	{"testdata/phase3/public_view_multi_arm_access_conflict.schway", "core.origin_omitted", "intended: a multi-arm access conflict fixture"},
	{"testdata/phase3/public_view_multi_arm_omitted.schway", "core.origin_omitted", "intended: an omitted-origin fixture"},
	{"testdata/phase3/public_view_omitted.schway", "core.origin_omitted", "intended: an omitted-origin fixture"},
	{"testdata/phase4/foreign_origin_omitted.schway", "core.foreign_origin_omitted", "intended: the foreign-declaration half of the same publication gate"},
	{"testdata/phase5/false_restrict_hoist.schway", "core.origin_omitted", "incidental: a positive restrict-hoist fixture that derives a borrow origin"},
	{"testdata/phase5/restrict_borrow.schway", "core.origin_omitted", "incidental: the already-shipped Phase 5 fixture, proving the divergence predates Phase 11 by six phases"},
}

// TestCLICheckAdmissionDivergenceIsExactlyKnown sweeps every committed .schway
// fixture through both admission surfaces and asserts the divergence set
// equals knownAdmissionDivergences exactly.
func TestCLICheckAdmissionDivergenceIsExactlyKnown(t *testing.T) {
	observed, bothAccepted := sweepAdmissionSurfaces(t)

	want := map[string]string{}
	for _, entry := range knownAdmissionDivergences {
		if previous, duplicated := want[entry.fixture]; duplicated {
			t.Fatalf("knownAdmissionDivergences lists %s twice (%q and %q)", entry.fixture, previous, entry.code)
		}
		want[entry.fixture] = entry.code
	}

	for fixture, code := range observed {
		expected, listed := want[fixture]
		if !listed {
			t.Errorf("UNDECLARED admission divergence: %s is accepted by session.Check+corevalidate.Validate but refused by the CLI `check` surface with %s.\n"+
				"The two admission surfaces have drifted further apart. Either fix the drift, or add this fixture to knownAdmissionDivergences with a why.", fixture, code)
			continue
		}
		if code != expected {
			t.Errorf("admission divergence for %s changed code: got %s, knownAdmissionDivergences says %s.\n"+
				"Update the table deliberately if the new code is correct.", fixture, code, expected)
		}
	}

	for fixture, code := range want {
		if _, stillDiverges := observed[fixture]; !stillDiverges {
			t.Errorf("RESOLVED admission divergence: %s no longer diverges (the table expects %s).\n"+
				"If the two admission surfaces were reconciled, remove this entry from knownAdmissionDivergences in the same change.", fixture, code)
		}
	}

	// Anti-vacuity: a sweep that reached no fixture, or a CLI surface that
	// refuses everything, would make every assertion above trivially true.
	if len(observed) == 0 {
		t.Fatal("anti-vacuity: the sweep observed zero divergences, which means it is not exercising the CLI admission surface at all")
	}
	if bothAccepted == 0 {
		t.Fatal("anti-vacuity: no fixture was accepted by BOTH surfaces, so the divergence set is indistinguishable from 'the CLI refuses everything'")
	}
}

// sweepAdmissionSurfaces walks every committed .schway fixture and returns the
// fixtures accepted by session.Check+corevalidate.Validate but refused by
// session.CheckCommandFile (the CLI `check` command's own surface), mapped to
// the diagnostic code the CLI answered with, plus a count of fixtures both
// surfaces accepted (the anti-vacuity witness).
func sweepAdmissionSurfaces(t *testing.T) (map[string]string, int) {
	t.Helper()
	root := testsupport.ProjectPath("testdata")
	projectRoot := testsupport.ProjectPath(".")

	observed := map[string]string{}
	bothAccepted := 0
	walked := 0

	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() || !strings.HasSuffix(path, ".schway") {
			return nil
		}
		walked++

		source, readErr := os.ReadFile(path)
		if readErr != nil || len(source) > syntax.MaxSourceBytes {
			return nil
		}

		// Surface A: the internal admission path -- what every production
		// run site and verify-corpus lane in this repository uses.
		checked := session.Check(source)
		if len(checked.Diagnostics) > 0 {
			return nil
		}
		if validated := corevalidate.Validate(checked.Program); !validated.Valid {
			return nil
		}

		// Surface B: the CLI `check` command's own refusing union, which
		// adds originvalidate.ValidatePublished on top.
		result, cmdErr := session.CheckCommandFile(path)
		if cmdErr != nil {
			return nil
		}

		relative, relErr := filepath.Rel(projectRoot, path)
		if relErr != nil {
			relative = path
		}
		relative = filepath.ToSlash(relative)

		if result.Status == protocol.StatusPass {
			bothAccepted++
			return nil
		}
		code := ""
		if len(result.Diagnostics) > 0 {
			code = result.Diagnostics[0].Code
		}
		observed[relative] = code
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	if walked == 0 {
		t.Fatalf("anti-vacuity: found no .schway fixtures under %s", root)
	}
	return observed, bothAccepted
}

// TestKnownAdmissionDivergencesIsSortedAndComplete keeps the committed table
// readable and every entry justified, so a future edit cannot quietly append
// an unexplained fixture to silence the sweep above.
func TestKnownAdmissionDivergencesIsSortedAndComplete(t *testing.T) {
	fixtures := make([]string, len(knownAdmissionDivergences))
	for i, entry := range knownAdmissionDivergences {
		fixtures[i] = entry.fixture
		if entry.code == "" {
			t.Errorf("knownAdmissionDivergences[%d] (%s) has no expected diagnostic code", i, entry.fixture)
		}
		if strings.TrimSpace(entry.why) == "" {
			t.Errorf("knownAdmissionDivergences[%d] (%s) has no why: every entry must say whether the rejection is intended or incidental", i, entry.fixture)
		}
		if path := testsupport.ProjectPath(filepath.FromSlash(entry.fixture)); !fileExists(path) {
			t.Errorf("knownAdmissionDivergences[%d] names %s, which does not exist", i, entry.fixture)
		}
	}
	if !sort.StringsAreSorted(fixtures) {
		t.Errorf("knownAdmissionDivergences must stay sorted by fixture path, got %v", fixtures)
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
