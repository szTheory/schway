package session

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"github.com/codename-lang/lang/internal/compiler/cache"
	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/execution"
	"github.com/codename-lang/lang/internal/compiler/interp"
	"github.com/codename-lang/lang/internal/compiler/measure"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/protocol"
	"github.com/codename-lang/lang/internal/compiler/syntax"
)

// phase6NativeDifferentialLane is the one lane this plan wires end-to-end
// through the cache and the risk-lane selector: session.go's own
// "lane:native-differential" identifier (LiveLaneIDs's liveLanesPureMatch),
// reused here rather than a new invented ID so this reporting lines up with
// the lane the pure_match corpus already recognizes.
const phase6NativeDifferentialLane = "lane:native-differential"

// phase6BuildFlags is the declared build_flags input this lane's compiled
// artifact is keyed on (D-06-07). A fixed, single optimization tier ("-O3")
// keeps this plan's own scope to ONE real cacheable artifact per fixture,
// distinct from VerifyCorpus's own -O0-and-O3 differential.
const phase6BuildFlags = "-O3"

// Phase6CacheRootOverrideForTest and Phase6ChangeStatePathOverrideForTest
// are mutable seams ONLY for tests, mirroring
// Phase5ClangPathOverrideForTest's own precedent: VerifyPhase6ChangedRisk's
// signature is fixed by the plan (ctx, corpus, runner) and carries no
// cache-root or state-path parameter, so a test roots both under its own
// t.TempDir() through these package-level variables instead of widening the
// production signature. Empty means "use the real os.UserCacheDir()-rooted
// locations" (cache.Open, DefaultChangeStatePath) -- production callers
// must never set these.
var (
	Phase6CacheRootOverrideForTest       = ""
	Phase6ChangeStatePathOverrideForTest = ""
	// Phase6FixtureKindOverrideForTest forces classifyPhase6FixtureKind's
	// result, ONLY for TestEveryLiveLaneIsAccountedFor's widened-to-every-
	// live-lane scenario (D-06-11's unclassified-kind path), which the real
	// corpus-marker classification below can never produce on its own
	// in-tree fixtures. Empty means "classify normally". Production callers
	// must never set this.
	Phase6FixtureKindOverrideForTest = ""
)

func phase6Store() (*cache.Store, error) {
	if Phase6CacheRootOverrideForTest != "" {
		return &cache.Store{Root: Phase6CacheRootOverrideForTest}, nil
	}
	return cache.Open()
}

func phase6ChangeStatePath() (string, error) {
	if Phase6ChangeStatePathOverrideForTest != "" {
		return Phase6ChangeStatePathOverrideForTest, nil
	}
	return DefaultChangeStatePath()
}

// classifyPhase6FixtureKind mirrors VerifyCorpus's own file-marker
// classification (session.go) so this sibling file's fixture-kind naming
// stays aligned with the corpus shapes the shipped verify dispatch already
// recognizes -- duplicated rather than extracted, per this plan's own
// "session.go untouched" constraint.
func classifyPhase6FixtureKind(corpus string) string {
	if Phase6FixtureKindOverrideForTest != "" {
		return Phase6FixtureKindOverrideForTest
	}
	if _, err := os.Stat(filepath.Join(corpus, "foreign_acquire_one.lang")); err == nil {
		return "foreign"
	}
	if _, err := os.Stat(filepath.Join(corpus, "borrowed_view.lang")); err == nil {
		return "borrowed"
	}
	if _, err := os.Stat(filepath.Join(corpus, "owned_transfer.lang")); err == nil {
		return "owned"
	}
	return "pure_match"
}

// phase6SelfSourcePath returns this file's own absolute path via
// runtime.Caller(0), used as the mutation_runner_source declared input's
// placeholder (D-06-07): this verify path has no real per-mutant runner of
// its own the way NAT03Mutations does, but InputsFor refuses an empty
// declared value outright (D-06-11), so a stable, always-resolvable file
// stands in -- explicit, not silently blank -- and is independent of the
// caller's own working directory, unlike a relative path would be.
func phase6SelfSourcePath() string {
	_, thisFile, _, _ := runtime.Caller(0)
	return thisFile
}

// phase6ArtifactSpec builds spec's seven declared inputs for one fixture's
// compiled-binary artifact.
func phase6ArtifactSpec(source []byte, clangPath string) cache.ArtifactSpec {
	return cache.ArtifactSpec{
		Kind:                     cache.KindCompiledBinary,
		FixtureSource:            source,
		BuildFlags:               phase6BuildFlags,
		ClangPath:                clangPath,
		RuntimeIdentity:          "none",
		ForeignTranslationUnit:   []byte("none"),
		MutationRunnerSourcePath: phase6SelfSourcePath(),
		GoToolchain:              runtime.Version(),
	}
}

// phase6CompileBinary compiles cSource at optimization into a standalone
// binary and returns its bytes. It is a minimal, local duplicate of
// native.Runner.Run's own compile step (not a call into native.go, which is
// out of this plan's file scope, and which never exposes a compiled
// binary's own path or bytes independent of also running it) -- compiling
// here is the ONLY way to obtain artifact bytes worth caching before
// deciding whether to run them.
//
// It runs as two genuinely separate clang invocations -- compile-to-object,
// then link-object-to-executable -- rather than one combined
// compile-and-link command, so D-06-21's native_compile and link stages are
// real, independently timed pipeline boundaries via recorder rather than
// two labels stamped on one measured span.
func phase6CompileBinary(ctx context.Context, clangPath, cSource, optimization string, timeout time.Duration, recorder *StageRecorder) ([]byte, error) {
	if clangPath == "" {
		clangPath = "clang"
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	directory, err := os.MkdirTemp("", "lang-phase6-compile-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(directory)

	sourcePath := filepath.Join(directory, "program.c")
	objectPath := filepath.Join(directory, "program.o")
	binaryPath := filepath.Join(directory, "program")
	if err := os.WriteFile(sourcePath, []byte(cSource), 0o600); err != nil {
		return nil, err
	}

	if err := recorder.Start("native_compile"); err != nil {
		return nil, err
	}
	compileCtx, compileCancel := context.WithTimeout(ctx, timeout)
	compileArguments := []string{"-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic", optimization, "-c", sourcePath, "-o", objectPath}
	compileCommand := exec.CommandContext(compileCtx, clangPath, compileArguments...)
	var compileStderr bytes.Buffer
	compileCommand.Stderr = &compileStderr
	compileErr := compileCommand.Run()
	compileCancel()
	if compileErr != nil {
		return nil, fmt.Errorf("phase6.compile_failed: %w (stderr: %s)", compileErr, compileStderr.String())
	}
	if err := recorder.Stop("native_compile"); err != nil {
		return nil, err
	}

	if err := recorder.Start("link"); err != nil {
		return nil, err
	}
	linkCtx, linkCancel := context.WithTimeout(ctx, timeout)
	linkArguments := []string{optimization, objectPath, "-o", binaryPath}
	linkCommand := exec.CommandContext(linkCtx, clangPath, linkArguments...)
	var linkStderr bytes.Buffer
	linkCommand.Stderr = &linkStderr
	linkErr := linkCommand.Run()
	linkCancel()
	if linkErr != nil {
		return nil, fmt.Errorf("phase6.link_failed: %w (stderr: %s)", linkErr, linkStderr.String())
	}
	if err := recorder.Stop("link"); err != nil {
		return nil, err
	}

	return os.ReadFile(binaryPath)
}

// phase6WriteExecutable materializes binary bytes -- whether freshly
// compiled or reused from the cache -- as a runnable temp file, and returns
// a cleanup func. This is the "hand that path to the existing native
// execution path" step the plan describes: on a cache hit the bytes never
// touch clang at all, only this write.
func phase6WriteExecutable(binary []byte) (path string, cleanup func(), err error) {
	directory, err := os.MkdirTemp("", "lang-phase6-run-")
	if err != nil {
		return "", nil, err
	}
	binaryPath := filepath.Join(directory, "program")
	if err := os.WriteFile(binaryPath, binary, 0o700); err != nil {
		os.RemoveAll(directory)
		return "", nil, err
	}
	return binaryPath, func() { os.RemoveAll(directory) }, nil
}

// phase6RunCompiledBinary runs binaryPath directly (no recompilation) with
// input as its sole argument and decodes its stdout as an
// execution.Execution document -- a minimal, local run loop distinct from
// native.Runner.Run's own (which always compiles internally and has no seam
// for "run this binary I already have").
func phase6RunCompiledBinary(ctx context.Context, binaryPath, input string, timeout time.Duration) (execution.Execution, error) {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	command := exec.CommandContext(runCtx, binaryPath, input)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		return execution.Execution{}, fmt.Errorf("phase6.native_run_failed: %w (stderr: %s)", err, stderr.String())
	}
	var value execution.Execution
	if err := json.Unmarshal(stdout.Bytes(), &value); err != nil {
		return execution.Execution{}, fmt.Errorf("phase6.native_decode_failed: %w", err)
	}
	return value, nil
}

// verifyPhase6NativeDifferentialLane runs phase6NativeDifferentialLane
// end-to-end for one fixture: the checker and the interpreter-vs-native
// comparator ALWAYS re-run fresh against whatever binary is on disk this
// invocation, entirely OUTSIDE the if/else that decides whether that binary
// came from the cache or a fresh compile (D-06-06, T-06-VERIFY-01). Neither
// branch of that if/else may ever skip the comparator call below it; a
// go/ast structural test (TestNativeDifferentialAssertionOutsideCacheBranch,
// in the sibling _test.go file) proves this by construction, not merely by
// review.
func verifyPhase6NativeDifferentialLane(ctx context.Context, source []byte, runner native.Runner, store *cache.Store) (protocol.Lane, cache.Outcome, error) {
	laneStarted := time.Now()
	recorder := &StageRecorder{}

	// D-06-21's five fixed, sequential stage boundaries, instrumented on
	// this ONE lane end-to-end: parse and check are timed as two genuinely
	// separate steps (not one combined call into session.Check, which
	// would collapse both into a single measured span), duplicating
	// Check's own parse-then-check-then-validate shape locally -- the same
	// "duplicate rather than modify session.go" discipline
	// classifyPhase6FixtureKind and phase6CompileBinary already follow in
	// this sibling file.
	if err := recorder.Start("parse"); err != nil {
		return protocol.Lane{}, cache.Outcome{}, err
	}
	parsed := syntax.Parse(source)
	if err := recorder.Stop("parse"); err != nil {
		return protocol.Lane{}, cache.Outcome{}, err
	}
	if len(parsed.Diagnostics) != 0 {
		return protocol.Lane{}, cache.Outcome{}, fmt.Errorf("phase6.fixture_invalid: parser rejected the native-differential fixture")
	}

	// The checker (and independent core validator) run next, unconditionally
	// -- before the cache is even consulted, so neither can ever be skipped
	// by a cache outcome.
	if err := recorder.Start("check"); err != nil {
		return protocol.Lane{}, cache.Outcome{}, err
	}
	checked := check.Program(parsed.Program)
	// Phase 11 (11-GUARD-LEDGER.md): KEPT. verifyPhase6NativeDifferentialLane
	// is Phase 6's own cache-backed differential, driven only over
	// testdata/phase1's single-function toggle.lang (Phase6RequiredControls'
	// control:interpreter-o0-o3). NAT-06's own four-tier interprocedural
	// differential is a separate, new lane
	// (session_phase11_differential_test.go) that does not call this
	// function; widening it would also require extending
	// SelectLanesForFixture's change-state to a shape D-11-42 explicitly
	// declines to add without the same scrutiny D-11-41 demands.
	if len(checked.Diagnostics) != 0 || len(checked.Program.Functions) != 1 {
		return protocol.Lane{}, cache.Outcome{}, fmt.Errorf("phase6.fixture_invalid: checker rejected the native-differential fixture")
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		return protocol.Lane{}, cache.Outcome{}, fmt.Errorf("phase6.fixture_invalid: core validation rejected the native-differential fixture")
	}
	if err := recorder.Stop("check"); err != nil {
		return protocol.Lane{}, cache.Outcome{}, err
	}
	program := validated.Program()
	inputs, ok := interpreterInputs(program)
	if !ok || len(inputs) == 0 {
		return protocol.Lane{}, cache.Outcome{}, fmt.Errorf("phase6.fixture_invalid: no interpreter inputs for the native-differential fixture")
	}

	if err := recorder.Start("lower"); err != nil {
		return protocol.Lane{}, cache.Outcome{}, err
	}
	cSource, err := cgen.EmitNative(program)
	if err != nil {
		return protocol.Lane{}, cache.Outcome{}, fmt.Errorf("phase6.cgen_failed: %w", err)
	}
	if err := recorder.Stop("lower"); err != nil {
		return protocol.Lane{}, cache.Outcome{}, err
	}

	clangPath := runner.ClangPath
	if clangPath == "" {
		clangPath = "clang"
	}
	spec := phase6ArtifactSpec(source, clangPath)
	outcome, err := cache.Consult(ctx, store, spec)
	if err != nil {
		return protocol.Lane{}, cache.Outcome{}, fmt.Errorf("phase6.cache_consult_failed: %w", err)
	}

	var binary []byte
	if outcome.Status == cache.StatusArtifactReused {
		binary = outcome.Artifact
	} else {
		// native_compile and link are recorded INSIDE phase6CompileBinary,
		// as two separate clang invocations, so a cache-artifact-reused run
		// simply never starts either stage rather than reporting a
		// zero-elapsed entry for work that did not happen this invocation.
		compiled, compileErr := phase6CompileBinary(ctx, clangPath, cSource, phase6BuildFlags, runner.Timeout, recorder)
		if compileErr != nil {
			return protocol.Lane{}, outcome, fmt.Errorf("phase6.compile_failed: %w", compileErr)
		}
		binary = compiled
		if outcome.Status == cache.StatusArtifactRecomputed {
			if putErr := store.Put(outcome.Key, binary); putErr != nil {
				return protocol.Lane{}, outcome, fmt.Errorf("phase6.cache_put_failed: %w", putErr)
			}
		}
	}

	// The comparator ALWAYS re-runs fresh here, unconditionally, entirely
	// after (never nested inside) the if/else above -- a cache hit changed
	// only whether clang ran, never whether this comparison runs.
	binaryPath, cleanup, err := phase6WriteExecutable(binary)
	if err != nil {
		return protocol.Lane{}, outcome, fmt.Errorf("phase6.write_executable_failed: %w", err)
	}
	defer cleanup()

	work := 0
	for _, input := range inputs {
		interpreted, interpErr := interp.Run(program, program.Functions[0].Name, input)
		if interpErr != nil {
			return protocol.Lane{}, outcome, fmt.Errorf("phase6.interpret_failed: %w", interpErr)
		}
		work++
		actual, runErr := phase6RunCompiledBinary(ctx, binaryPath, input, runner.Timeout)
		if runErr != nil {
			return protocol.Lane{}, outcome, fmt.Errorf("phase6.native_run_failed: %w", runErr)
		}
		work++
		if !execution.Equal(interpreted, actual) {
			return protocol.Lane{
				Schema: protocol.LaneSchema1, ID: phase6NativeDifferentialLane, Status: protocol.StatusMismatch,
				RecomputedWork: work, ElapsedNS: time.Since(laneStarted).Nanoseconds(), PeakRSSStatus: protocol.PeakRSSUnavailable,
				CacheStatus: string(outcome.Status), StageBreakdown: recorder.Breakdown(),
			}, outcome, nil
		}
	}

	lane := protocol.Lane{
		Schema: protocol.LaneSchema1, ID: phase6NativeDifferentialLane, Status: protocol.StatusPass,
		Controls: []string{"control:interpreter-o0-o3"}, RecomputedWork: work,
		ElapsedNS: time.Since(laneStarted).Nanoseconds(), PeakRSSStatus: protocol.PeakRSSUnavailable,
		StageBreakdown: recorder.Breakdown(),
		CacheStatus:    string(outcome.Status),
	}
	return lane, outcome, nil
}

// addDeferredLane hardcodes protocol.StatusDeferred and takes NO status
// parameter at all (T-06-VERIFY-02): a lane rendered through this closure is
// structurally incapable of reporting anything but deferred, whether the
// selector's own reason for it was "deferred: no declared dependency",
// "selected: ... matched" (selected by the risk table but not yet wired
// through this function's own executable-lane set), or "widened:
// undeclared-input risk". TestDeferredLaneNeverRendersPass's go/ast half
// scans for exactly this: protocol.StatusPass must never appear inside this
// closure's body.
func phase6AddDeferredLane(result *protocol.Result, id, reason, coldOrWarm string) {
	result.Lanes = append(result.Lanes, protocol.Lane{
		Schema: protocol.LaneSchema1, ID: id, Status: protocol.StatusDeferred,
		SelectionReason: reason, ColdOrWarm: coldOrWarm,
	})
}

// VerifyPhase6ChangedRisk wires the change-risk selector (06-05) and the
// artifact cache (06-04) into a real verify path (D-06-06/D-06-10/D-06-12):
// it selects lanes by changed risk, runs the one lane this plan wires
// end-to-end (phase6NativeDifferentialLane) over whatever artifact -- cached
// or freshly compiled -- resolves this invocation, and renders every other
// lane in scope for this fixture's classified kind as deferred (never
// silently absent, never rendered as pass) via phase6AddDeferredLane.
func VerifyPhase6ChangedRisk(ctx context.Context, corpus string, runner native.Runner) (protocol.Result, error) {
	started := time.Now()
	result := protocol.New("verify", protocol.StatusPass)

	kind := classifyPhase6FixtureKind(corpus)
	fixtureID := filepath.Clean(corpus)

	clangPath := runner.ClangPath
	if clangPath == "" {
		clangPath = "clang"
	}

	fixturePath := filepath.Join(corpus, "toggle.lang")
	source, readErr := os.ReadFile(fixturePath)

	current := FixtureInputs{}
	var computeErr error
	if readErr != nil {
		computeErr = readErr
	} else {
		inputs, inputsErr := cache.InputsFor(ctx, phase6ArtifactSpec(source, clangPath))
		if inputsErr != nil {
			computeErr = inputsErr
		} else {
			for _, in := range inputs {
				current[in.Name] = in.Digest
			}
		}
	}

	statePath, err := phase6ChangeStatePath()
	if err != nil {
		return protocol.Result{}, err
	}

	selection, nextState, cold, err := SelectLanesForFixture(kind, fixtureID, current, computeErr, statePath)
	if err != nil {
		return protocol.Result{}, err
	}

	coldOrWarm := protocol.LaneWarm
	if cold {
		coldOrWarm = protocol.LaneCold
	}

	store, err := phase6Store()
	if err != nil {
		return protocol.Result{}, err
	}

	// Declared-machine ratification (D-06-17/D-06-18): probe this host once,
	// and resolve whether it is a declared machine in the checked-in budget
	// manifest. Loading the manifest here is the one-time membership check
	// RatificationMode's own signature requires to answer that question --
	// distinct from the manifest-CONSULTING (BudgetFor + EvaluateBudget)
	// that stays strictly inside the Ratified==true branch below. On an
	// undeclared machine no such consulting ever runs; every lane's
	// GateVerdict renders not_ratified instead (D-06-18).
	machineID := ""
	var ratificationRows []QLT02BudgetRow
	ratified := false
	if facts, probeErr := measure.ProbeMachine(ctx); probeErr == nil {
		machineID = measure.MachineID(facts)
		if rows, loadErr := LoadQLT02BudgetManifest(); loadErr == nil {
			mode := RatificationMode(rows, machineID)
			if mode.Ratified {
				ratificationRows = rows
				ratified = true
			}
		}
	}

	selectedSet := make(map[string]bool, len(selection.LaneIDs))
	for _, id := range selection.LaneIDs {
		selectedSet[id] = true
	}

	laneIDs := make([]string, 0, len(selection.Reasons))
	for id := range selection.Reasons {
		laneIDs = append(laneIDs, id)
	}
	sort.Strings(laneIDs)

	var firstErr error
	for _, laneID := range laneIDs {
		reason := selection.Reasons[laneID]
		if laneID == phase6NativeDifferentialLane && selectedSet[laneID] && computeErr == nil {
			lane, _, runErr := verifyPhase6NativeDifferentialLane(ctx, source, runner, store)
			if runErr != nil {
				if firstErr == nil {
					firstErr = runErr
				}
				phase6AddDeferredLane(&result, laneID, reason+"; operational failure running the lane", coldOrWarm)
				continue
			}
			lane.SelectionReason = reason
			lane.ColdOrWarm = coldOrWarm
			lane.MachineID = machineID

			// Blocking rule (D-06-14, D-06-18, D-06-22): only recomputed_work
			// on a declared, ratified machine may ever gate. An undeclared
			// machine never consults the manifest -- not_ratified,
			// unconditionally.
			lane.GateVerdict = protocol.LaneGateNotRatified
			if ratified {
				if budgetRow, found := BudgetFor(ratificationRows, machineID, "recomputed_work"); found {
					summary := deterministicSummary(int64(lane.RecomputedWork))
					verdict, _ := EvaluateBudget(budgetRow, summary, lane.StageBreakdown)
					lane.GateVerdict = verdict
				}
			}

			if verr := protocol.ValidateLaneVocabularies(lane); verr != nil {
				if firstErr == nil {
					firstErr = verr
				}
				phase6AddDeferredLane(&result, laneID, reason+"; lane vocabulary invalid", coldOrWarm)
				continue
			}
			result.Lanes = append(result.Lanes, lane)
			result.Metrics.RecomputedWork += lane.RecomputedWork
			if lane.CacheStatus == string(cache.StatusArtifactReused) {
				result.Metrics.CacheInputsReusedCount++
			}
			if lane.GateVerdict == protocol.LaneGateBlocking && result.Status == protocol.StatusPass {
				// A blocking recomputed_work regression is the ONLY gate
				// this project ever fails on for cost -- never a wall-clock
				// or output-bytes observation on its own (D-06-22).
				result.Status = protocol.StatusInvalid
			}
			if lane.Status != protocol.StatusPass && result.Status == protocol.StatusPass {
				result.Status = lane.Status
			}
			continue
		}
		// Every other lane in scope for this invocation -- whether the
		// selector deferred it (nothing changed), selected it (matched but
		// not yet wired through this function's own executable-lane set),
		// or widened it (undeclared-input risk) -- is rendered deferred:
		// not run this invocation, never silently absent, never pass.
		phase6AddDeferredLane(&result, laneID, reason, coldOrWarm)
	}

	if err := nextState.Save(statePath); err != nil {
		return protocol.Result{}, err
	}

	if firstErr != nil && result.Status == protocol.StatusPass {
		result.Status = protocol.StatusOperational
	}

	return completeCommand(result, started, result.Metrics.RecomputedWork), nil
}
