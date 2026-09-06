package session

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/interp"
	"github.com/codename-lang/lang/internal/compiler/native"
)

// byPointerParamMarker is duplicated, verbatim, from cgen.go's own
// borrowByPointerMarker constant (D-05-01/D-05-02) — matching this file's
// own established precedent (session.go's releaseMarker/padInstallMarker/
// padEndMarker/ledgerPopulateMarker, all duplicated verbatim from cgen.go
// rather than exported and shared). It is the mutation's single target:
// AliasFactMutationRunner refuses to run when it does not appear on
// exactly one line.
const byPointerParamMarker = "/* lang:by-pointer-param */"

// aliasProbeParameterName is duplicated, verbatim, from cgen.go's own
// constant of the same name (D-05-05): the dormant second raw pointer
// parameter emitLinearBorrowedByPointerPlain exposes on a function
// selectsByPointerLoweringSharedOnly selects. Its PRESENCE on the marked
// line is how this runner decides whether a genuine aliasing-hoist
// demonstration is possible for the given fixture (testdata/phase5/
// false_restrict_hoist.lang) or whether the marked function has no such
// probe (e.g. testdata/phase5/restrict_borrow.lang's EXCLUSIVE, already-
// justified by-pointer path), in which case only the (already-satisfied,
// idempotent) restrict insertion applies and no divergence is possible —
// exactly the load-bearing negative case D-05-05 requires.
const aliasProbeParameterName = "lang_alias_probe"

// ControlAliasFalseNoAlias is D-05-05's control identifier: the mutation
// injects `restrict` onto a by-pointer parameter the checker did NOT prove
// exclusive (deriveAliasFacts, check package, returns zero facts for the
// fixture this control attacks), asserting the observable
// `interpreter == -O0 != -O3` signature a genuinely false alias claim
// produces under Clang's optimizer.
const ControlAliasFalseNoAlias = "control:alias.false_no_alias"

// AliasFactMutationRunner is control:alias.false_no_alias's fail-closed
// verification seam, on the exact marker-count-guard shape every existing
// NAT-03 mutation runner in this file establishes (OwnedBackendMutationRunner,
// ReleaseOmissionMutationRunner, NonlocalPadOmissionMutationRunner,
// NonlocalLedgerOmissionMutationRunner, LayoutMutationRunner): it scans the
// generated C for byPointerParamMarker, refuses to run on an ambiguous (>1)
// or absent (0) marker, and only then mutates. Unlike its siblings, its
// mutation has two parts: (1) inject `restrict` onto the marked parameter
// (a no-op, per injectRestrictIntoSignature's idempotence, when the
// parameter already carries it — e.g. a legitimately-justified by-pointer
// function like restrict_borrow.lang), and (2), ONLY when the marked
// function also exposes cgen's dormant aliasProbeParameterName, rewrite its
// body into the write/reread-through-a-second-pointer demonstration
// VerifyAliasFalseNoAlias asserts against (D-05-05). A marked function with
// no alias-probe parameter is mutated ONLY by (1), which is a genuine no-op
// when restrict is already present — the fail-closed marker guard still
// runs, but no divergence is possible, exactly the negative case
// TestAliasMutationWithoutDivergenceFailsTheLane requires.
type AliasFactMutationRunner struct {
	runner        native.Runner
	fixturePath   string
	mu            sync.Mutex
	optimizations []string
	clangVersion  string
}

// NewAliasFactMutationRunner wraps runner, attacking the .lang source at
// fixturePath. Unlike this file's other mutation runners (which attack
// whatever cSource their caller supplies), this one also needs to know
// which fixture it is attacking, since VerifyAliasFalseNoAlias must derive
// an independent interpreter oracle for the SAME source the mutated native
// runs are driven against.
func NewAliasFactMutationRunner(runner native.Runner, fixturePath string) *AliasFactMutationRunner {
	return &AliasFactMutationRunner{runner: runner, fixturePath: fixturePath}
}

// Optimizations returns every optimization level this runner has been
// asked to build, in call order — the same "recorded, not assumed" nonzero-
// work accessor every sibling mutation runner exposes.
func (r *AliasFactMutationRunner) Optimizations() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.optimizations...)
}

// RecomputedWork reports the number of optimization levels this runner has
// built — nonzero once VerifyAliasFalseNoAlias has run its -O0/-O3 pair,
// asserted rather than assumed by TestAliasFactMutationIsDetected.
func (r *AliasFactMutationRunner) RecomputedWork() int {
	return len(r.Optimizations())
}

// ClangVersion returns the `clang --version` identity string recorded on
// this runner's first build, so a toolchain upgrade that kills the
// -O0/-O3 divergence this control depends on is diagnosable against a
// recorded baseline rather than a memory (D-05-05's own repudiation
// mitigation, T-05-25). Empty until this runner's Run has executed at
// least once.
func (r *AliasFactMutationRunner) ClangVersion() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.clangVersion
}

// byPointerSignature is what Run's mutation extracts from the marked
// signature line: the primary pointer parameter's C type and name, and —
// when present — the dormant alias-probe parameter's type and name.
type byPointerSignature struct {
	paramType string
	paramName string
	aliasType string
	aliasName string
	hasAlias  bool
}

// injectRestrictIntoSignature inserts `restrict` immediately after the
// marked line's FIRST pointer declaration's `*`, matching D-05-05's own
// engineering finding (verified empirically on this host): marking only
// the PRIMARY parameter restrict — never a plain local pointer copy, which
// C17 6.7.3.1 treats as "based on" the restrict pointer and is therefore
// safe regardless — is sufficient to make the false claim observable.
// Idempotent: a line that already carries `restrict` (a legitimately-
// justified by-pointer function, e.g. restrict_borrow.lang) is returned
// unchanged, since there is nothing false to inject onto an
// already-true claim.
func injectRestrictIntoSignature(line string) (string, error) {
	if strings.Contains(line, "restrict") {
		return line, nil
	}
	index := strings.Index(line, " *")
	if index == -1 {
		return "", fmt.Errorf("marked signature line has no pointer parameter to attack: %q", line)
	}
	insertAt := index + len(" *")
	return line[:insertAt] + "restrict " + line[insertAt:], nil
}

// splitPointerDecl parses one parameter declaration of the marked line's
// parameter list, e.g. "unsigned char *restrict LANG_VALUE" or
// "unsigned char *lang_alias_probe", into its type and name.
func splitPointerDecl(decl string) (typeName, paramName string, err error) {
	starIndex := strings.Index(decl, "*")
	if starIndex == -1 {
		return "", "", fmt.Errorf("declaration %q is not a pointer", decl)
	}
	typeName = strings.TrimSpace(decl[:starIndex])
	rest := strings.TrimSpace(decl[starIndex+1:])
	rest = strings.TrimPrefix(rest, "restrict ")
	return typeName, strings.TrimSpace(rest), nil
}

// parseByPointerSignature parses the (possibly restrict-injected) marked
// line's parameter list.
func parseByPointerSignature(line string) (byPointerSignature, error) {
	openParen := strings.Index(line, "(")
	closeParen := strings.Index(line, ")")
	if openParen == -1 || closeParen == -1 || closeParen < openParen {
		return byPointerSignature{}, fmt.Errorf("marked line has no parameter list: %q", line)
	}
	parts := strings.Split(line[openParen+1:closeParen], ",")
	paramType, paramName, err := splitPointerDecl(strings.TrimSpace(parts[0]))
	if err != nil {
		return byPointerSignature{}, err
	}
	signature := byPointerSignature{paramType: paramType, paramName: paramName}
	if len(parts) > 1 {
		aliasType, aliasName, err := splitPointerDecl(strings.TrimSpace(parts[1]))
		if err != nil {
			return byPointerSignature{}, err
		}
		if aliasName == aliasProbeParameterName {
			signature.hasAlias = true
			signature.aliasType = aliasType
			signature.aliasName = aliasName
		}
	}
	return signature, nil
}

// Run implements the fail-closed marker-count guard every sibling mutation
// runner in this file establishes, then mutates as documented on
// AliasFactMutationRunner itself.
func (r *AliasFactMutationRunner) Run(ctx context.Context, cSource, optimization string, inputs []string) (native.Result, error) {
	lines := strings.Split(cSource, "\n")
	matched := -1
	for index, line := range lines {
		if strings.Contains(line, byPointerParamMarker) {
			if matched != -1 {
				return native.Result{}, &native.ToolError{Code: "native.alias_control_invalid", Err: fmt.Errorf("alias mutation marker count is >1, want 1")}
			}
			matched = index
		}
	}
	if matched == -1 {
		return native.Result{}, &native.ToolError{Code: "native.alias_control_invalid", Err: fmt.Errorf("alias mutation marker count is 0, want 1")}
	}

	mutatedLine, err := injectRestrictIntoSignature(lines[matched])
	if err != nil {
		return native.Result{}, &native.ToolError{Code: "native.alias_control_invalid", Err: err}
	}
	lines[matched] = mutatedLine

	signature, err := parseByPointerSignature(mutatedLine)
	if err != nil {
		return native.Result{}, &native.ToolError{Code: "native.alias_control_invalid", Err: err}
	}

	if signature.hasAlias {
		endIndex := -1
		for index := matched + 1; index < len(lines); index++ {
			if strings.TrimSpace(lines[index]) == "}" {
				endIndex = index
				break
			}
		}
		if endIndex == -1 {
			return native.Result{}, &native.ToolError{Code: "native.alias_control_invalid", Err: fmt.Errorf("marked function body has no closing brace")}
		}
		returnIndex := -1
		for index := endIndex - 1; index > matched; index-- {
			if strings.HasPrefix(strings.TrimSpace(lines[index]), "return ") {
				returnIndex = index
				break
			}
		}
		if returnIndex == -1 {
			return native.Result{}, &native.ToolError{Code: "native.alias_control_invalid", Err: fmt.Errorf("marked function body has no terminal return")}
		}
		// Everything between the marker and the original `return` — the
		// emitter's own lang_record_event calls, carrying real, valid
		// place/type IDs — is PRESERVED, not discarded: native.Runner's own
		// execution-document validator requires at least one well-formed
		// transition event followed by a well-formed return event, and a
		// synthetic replacement carrying NULL fields fails that validator.
		// Only the terminal `return` is replaced with the write/reread
		// demonstration itself.
		demonstration := []string{
			fmt.Sprintf("  %s lang_alias_control_first = *%s;", signature.paramType, signature.paramName),
			fmt.Sprintf("  *%s = (%s)2;", signature.aliasName, signature.paramType),
			fmt.Sprintf("  %s lang_alias_control_second = *%s;", signature.paramType, signature.paramName),
			"  (void)lang_alias_control_first;",
			"  return lang_alias_control_second;",
		}
		rewritten := make([]string, 0, matched+1+(returnIndex-matched-1)+len(demonstration)+(len(lines)-endIndex))
		rewritten = append(rewritten, lines[:matched+1]...)
		rewritten = append(rewritten, lines[matched+1:returnIndex]...)
		rewritten = append(rewritten, demonstration...)
		rewritten = append(rewritten, lines[endIndex:]...)
		lines = rewritten
	}

	mutated := strings.Join(lines, "\n")

	r.mu.Lock()
	r.optimizations = append(r.optimizations, optimization)
	if r.clangVersion == "" {
		r.clangVersion = detectClangVersion(r.runner.ClangPath)
	}
	r.mu.Unlock()

	return r.runner.Run(ctx, mutated, optimization, inputs)
}

// detectClangVersion runs `clang --version` (or the runner's own
// ClangPath), returning its first line, or an operational placeholder if
// the tool cannot be located — this recording is diagnostic evidence
// (T-05-25), never a gate, so a missing tool never blocks the control.
// maxClangVersionBytes bounds the `clang --version` probe's captured
// output, matching this codebase's own "no unbounded process I/O" rule
// (TestSourceNeverSpawnsUnboundedProcesses) — a version string is at most
// a few lines, so this bound is never reached by a well-behaved toolchain.
const maxClangVersionBytes = 4096

// boundedVersionWriter caps captured bytes at maxClangVersionBytes, then
// silently discards the rest — mirroring native.go's own boundedWriter
// discipline (an independently bounded writer per D-04-19's "host-tool-
// absence returns tool_missing/operational, never a silent pass" posture,
// duplicated here rather than exported since this probe is diagnostic
// evidence only, never a gate).
type boundedVersionWriter struct {
	buffer bytes.Buffer
}

func (w *boundedVersionWriter) Write(data []byte) (int, error) {
	if w.buffer.Len() < maxClangVersionBytes {
		remaining := maxClangVersionBytes - w.buffer.Len()
		if remaining > len(data) {
			remaining = len(data)
		}
		w.buffer.Write(data[:remaining])
	}
	return len(data), nil
}

func detectClangVersion(clangPath string) string {
	if clangPath == "" {
		clangPath = "clang"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, clangPath, "--version")
	var output boundedVersionWriter
	command.Stdout = &output
	if err := command.Run(); err != nil {
		return "clang version unavailable: " + err.Error()
	}
	first, _, _ := strings.Cut(output.buffer.String(), "\n")
	return strings.TrimSpace(first)
}

// VerifyAliasFalseNoAlias builds runner's fixture through runner at both
// -O0 and -O3, runs the UNMUTATED interpreter as an independent oracle, and
// asserts D-05-05's exact signature: interpreter == -O0 (the mutation is
// inert at -O0, matching genuinely-correct semantics) AND -O0 != -O3 (the
// false restrict IS observably exploited at -O3). Any other combination is
// a lane failure: all three equal means the mutation produced no
// divergence (an unexercised control, T-05-23's spoofing threat --
// returned as an error naming ControlAliasFalseNoAlias, never a pass), and
// interpreter != -O0 means the mutation broke the program even without
// optimizer aggressiveness, which is not an optimizer-observability
// control at all.
func VerifyAliasFalseNoAlias(ctx context.Context, runner *AliasFactMutationRunner) error {
	source, err := os.ReadFile(runner.fixturePath)
	if err != nil {
		return err
	}
	checked := Check(source)
	if len(checked.Diagnostics) != 0 {
		return fmt.Errorf("fixture %s failed to check: %+v", runner.fixturePath, checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		return fmt.Errorf("fixture %s rejected by corevalidate: %+v", runner.fixturePath, validated.Problems)
	}
	program := validated.Program()
	if len(program.Functions) != 1 {
		return fmt.Errorf("fixture %s must declare exactly one function", runner.fixturePath)
	}

	cSource, err := cgen.EmitNative(program)
	if err != nil {
		return err
	}

	nativeInputs, ok := interpreterInputs(program)
	if !ok || len(nativeInputs) == 0 {
		return fmt.Errorf("fixture %s has no derivable interpreter input", runner.fixturePath)
	}
	nativeInput := nativeInputs[0]

	// D-05-05's engineered demonstration ignores the fixture's own real
	// parameter value (it hard-codes the write/reread constants below), so
	// a fixture whose marked function carries the dormant alias-probe
	// parameter is compared against the FIXED value that demonstration
	// computes when genuinely (non-hoisted) correct — never the fixture's
	// own real identity value. A fixture with no alias-probe (the
	// no-divergence negative case) is unaffected by the mutation at all, so
	// its own real canonical input is the correct oracle.
	interpreterInput := nativeInput
	if strings.Contains(cSource, aliasProbeParameterName) {
		interpreterInput = "2"
	}
	interpreted, err := interp.Run(program, program.Functions[0].Name, interpreterInput)
	if err != nil {
		return err
	}

	o0, err := runner.Run(ctx, cSource, "-O0", []string{nativeInput})
	if err != nil {
		return err
	}
	o3, err := runner.Run(ctx, cSource, "-O3", []string{nativeInput})
	if err != nil {
		return err
	}
	if len(o0.Pairs) != 1 || len(o3.Pairs) != 1 {
		return fmt.Errorf("expected exactly one execution per optimization level, got -O0=%d -O3=%d", len(o0.Pairs), len(o3.Pairs))
	}

	interpreterValue := interpreted.Outcome.Value
	o0Value := o0.Pairs[0].Execution.Outcome.Value
	o3Value := o3.Pairs[0].Execution.Outcome.Value

	if interpreterValue == o0Value && o0Value == o3Value {
		return fmt.Errorf("%s produced no divergence (interpreter=%s, -O0=%s, -O3=%s) -- the mutation is unexercised", ControlAliasFalseNoAlias, interpreterValue, o0Value, o3Value)
	}
	if interpreterValue != o0Value {
		return fmt.Errorf("%s mutation broke -O0 (expected agreement with the interpreter): interpreter=%s, -O0=%s", ControlAliasFalseNoAlias, interpreterValue, o0Value)
	}
	return nil
}
