package evidence

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/execution"
	"github.com/codename-lang/lang/internal/compiler/interp"
	"github.com/codename-lang/lang/internal/compiler/syntax"
)

const (
	Schema0           = "lang.evidence/0"
	Schema1           = "lang.evidence/1"
	Schema            = Schema0
	IDAlgorithm       = "sha256-v1"
	SourceSchema      = "lang.source/s1"
	DigestClaim       = "content-identity-only"
	MaxManifestBytes  = 1 << 20
	MaxToolProbeBytes = 64 * 1024
)

var DefaultFlags = []string{"-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic", "-O0", "-O3"}

type Facts struct {
	CompilerIdentity string
	ClangIdentity    string
	Target           string
	Flags            []string
	Policy           string
}

type Manifest struct {
	Schema           string   `json:"schema"`
	ID               string   `json:"id"`
	IDAlgorithm      string   `json:"id_algorithm"`
	SourceSchema     string   `json:"source_schema"`
	CoreSchema       string   `json:"core_schema"`
	ExecutionSchema  string   `json:"execution_schema"`
	DiagnosticSchema string   `json:"diagnostic_schema,omitempty"`
	CompilerIdentity string   `json:"compiler_identity"`
	ClangIdentity    string   `json:"clang_identity"`
	Target           string   `json:"target"`
	Flags            []string `json:"flags"`
	Policy           string   `json:"policy"`
	SourceDigest     string   `json:"source_digest"`
	CoreDigest       string   `json:"core_digest"`
	CDigest          string   `json:"c_digest"`
	ExecutionDigests []string `json:"execution_digests,omitempty"`
	DigestClaim      string   `json:"digest_claim,omitempty"`
	KnownEscape      string   `json:"known_escape,omitempty"`
}

type Product struct {
	Manifest        Manifest
	CanonicalSource []byte
	CoreBytes       []byte
	CSource         []byte
	ManifestBytes   []byte
	Executions      []execution.Execution
}

type ValidationError struct {
	Code string
}

type ToolProbeError struct {
	Code string
	Err  error
}

func (e *ToolProbeError) Error() string { return e.Code + ": " + e.Err.Error() }
func (e *ToolProbeError) Unwrap() error { return e.Err }

type commandFactory func(context.Context, string, ...string) *exec.Cmd

func (e *ValidationError) Error() string { return e.Code }

func DefaultFacts(ctx context.Context, clangPath string) (Facts, error) {
	return defaultFacts(ctx, clangPath, exec.CommandContext)
}

func defaultFacts(ctx context.Context, clangPath string, command commandFactory) (Facts, error) {
	if clangPath == "" {
		clangPath = "clang"
	}
	versionOutput, err := runToolProbe(ctx, command, clangPath, "--version")
	if err != nil {
		return Facts{}, err
	}
	targetOutput, err := runToolProbe(ctx, command, clangPath, "-dumpmachine")
	if err != nil {
		return Facts{}, err
	}
	version := strings.TrimSpace(strings.SplitN(string(versionOutput), "\n", 2)[0])
	return Facts{
		CompilerIdentity: "codename-lang-stage0/" + runtime.Version(),
		ClangIdentity:    version,
		Target:           strings.TrimSpace(string(targetOutput)),
		Flags:            append([]string(nil), DefaultFlags...),
		Policy:           "phase1-pure-c17-v1",
	}, nil
}

type boundedProbeWriter struct {
	buffer bytes.Buffer
	total  int
}

func (w *boundedProbeWriter) Write(data []byte) (int, error) {
	w.total += len(data)
	remaining := MaxToolProbeBytes + 1 - w.buffer.Len()
	if remaining > len(data) {
		remaining = len(data)
	}
	if remaining > 0 {
		_, _ = w.buffer.Write(data[:remaining])
	}
	return len(data), nil
}

func runToolProbe(parent context.Context, command commandFactory, name string, arguments ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	cmd := command(ctx, name, arguments...)
	var stdout, stderr boundedProbeWriter
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return nil, &ToolProbeError{Code: "evidence.tool_timeout", Err: ctx.Err()}
	}
	if stdout.total > MaxToolProbeBytes {
		return nil, &ToolProbeError{Code: "evidence.tool_stdout_truncated", Err: fmt.Errorf("tool stdout exceeded %d bytes", MaxToolProbeBytes)}
	}
	if stderr.total > MaxToolProbeBytes {
		return nil, &ToolProbeError{Code: "evidence.tool_stderr_truncated", Err: fmt.Errorf("tool stderr exceeded %d bytes", MaxToolProbeBytes)}
	}
	if err != nil {
		return nil, &ToolProbeError{Code: "evidence.tool_failed", Err: err}
	}
	return append([]byte(nil), stdout.buffer.Bytes()...), nil
}

func Build(source []byte, facts Facts) (Product, []diagnostic.Diagnostic, error) {
	// This manifest proves that these compiler products are mutually bound. It
	// does not independently prove that a coordinated frontend translated user
	// intent into the correct core; later certificate checks retain that named
	// trust boundary rather than overstating this artifact as a proof.
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) > 0 {
		return Product{}, parsed.Diagnostics, nil
	}
	canonicalSource := syntax.Format(parsed.Tree)
	canonicalParsed := syntax.Parse(canonicalSource)
	checked := check.Program(canonicalParsed.Program)
	if len(checked.Diagnostics) > 0 {
		return Product{}, checked.Diagnostics, nil
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		return Product{}, nil, fmt.Errorf("core validation failed: %s", validated.Problems[0].Code)
	}
	checked.Program = validated.Program()
	coreBytes, err := json.Marshal(checked.Program)
	if err != nil {
		return Product{}, nil, err
	}
	cSource, err := cgen.Emit(checked.Program)
	if err != nil {
		return Product{}, nil, err
	}
	manifest := Manifest{
		Schema: Schema0, IDAlgorithm: IDAlgorithm,
		SourceSchema: SourceSchema, CoreSchema: core.Schema, ExecutionSchema: interp.Schema,
		CompilerIdentity: facts.CompilerIdentity, ClangIdentity: facts.ClangIdentity,
		Target: facts.Target, Flags: append([]string(nil), facts.Flags...), Policy: facts.Policy,
		SourceDigest: digest(canonicalSource), CoreDigest: digest(coreBytes), CDigest: digest([]byte(cSource)),
	}
	executions := []execution.Execution{}
	if checked.Program.Schema == core.Schema1 {
		manifest.Schema = Schema1
		manifest.CoreSchema = core.Schema1
		manifest.ExecutionSchema = execution.Schema1
		manifest.DiagnosticSchema = diagnostic.Schema1
		manifest.DigestClaim = DigestClaim
		manifest.KnownEscape = corevalidate.KnownEscape
		if manifest.Policy == "phase1-pure-c17-v1" {
			manifest.Policy = "phase2-owned-c17-v1"
		}
		for _, function := range checked.Program.Functions {
			input, ok := evidenceInput(function)
			if !ok {
				continue
			}
			value, runErr := interp.Run(checked.Program, function.Name, input)
			if runErr != nil {
				return Product{}, nil, runErr
			}
			owned, cloneErr := cloneExecution(value)
			if cloneErr != nil {
				return Product{}, nil, cloneErr
			}
			executions = append(executions, owned)
			encoded, encodeErr := execution.CanonicalBytes(owned)
			if encodeErr != nil {
				return Product{}, nil, encodeErr
			}
			manifest.ExecutionDigests = append(manifest.ExecutionDigests, digest(encoded))
		}
	}
	manifest.ID = manifestID(manifest)
	manifestBytes, err := CanonicalBytes(manifest)
	if err != nil {
		return Product{}, nil, err
	}
	return Product{Manifest: manifest, CanonicalSource: canonicalSource, CoreBytes: coreBytes, CSource: []byte(cSource), ManifestBytes: manifestBytes, Executions: executions}, nil, nil
}

func evidenceInput(function core.Function) (string, bool) {
	if function.Linear == nil || function.Match != nil {
		return "", false
	}
	switch function.Parameter.Type {
	case "Byte":
		return "7", true
	case "Buffer":
		return "01020304", true
	default:
		return "", false
	}
}

func cloneExecution(value execution.Execution) (execution.Execution, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return execution.Execution{}, err
	}
	var cloned execution.Execution
	if err := json.Unmarshal(encoded, &cloned); err != nil {
		return execution.Execution{}, err
	}
	return cloned, nil
}

func CanonicalBytes(manifest Manifest) ([]byte, error) {
	manifest.ID = manifestID(manifest)
	encoded, err := json.Marshal(manifest)
	if err != nil {
		return nil, err
	}
	return append(encoded, '\n'), nil
}

func DecodeStrict(data []byte) (Manifest, error) {
	if len(data) > MaxManifestBytes {
		return Manifest{}, &ValidationError{Code: "evidence.input_limit"}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var manifest Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, &ValidationError{Code: "evidence.invalid_json"}
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return Manifest{}, &ValidationError{Code: "evidence.trailing_json"}
	}
	return manifest, nil
}

func Validate(manifest Manifest, source []byte, facts Facts) error {
	expected, diagnostics, err := Build(source, facts)
	if err != nil {
		return err
	}
	if len(diagnostics) > 0 {
		return &ValidationError{Code: "evidence.source_invalid"}
	}
	checks := []struct {
		code string
		got  string
		want string
	}{
		{"evidence.schema_mismatch", manifest.Schema, expected.Manifest.Schema},
		{"evidence.id_algorithm_mismatch", manifest.IDAlgorithm, expected.Manifest.IDAlgorithm},
		{"evidence.source_schema_mismatch", manifest.SourceSchema, expected.Manifest.SourceSchema},
		{"evidence.core_schema_mismatch", manifest.CoreSchema, expected.Manifest.CoreSchema},
		{"evidence.execution_schema_mismatch", manifest.ExecutionSchema, expected.Manifest.ExecutionSchema},
		{"evidence.compiler_mismatch", manifest.CompilerIdentity, expected.Manifest.CompilerIdentity},
		{"evidence.clang_mismatch", manifest.ClangIdentity, expected.Manifest.ClangIdentity},
		{"evidence.target_mismatch", manifest.Target, expected.Manifest.Target},
		{"evidence.policy_mismatch", manifest.Policy, expected.Manifest.Policy},
	}
	if expected.Manifest.Schema == Schema1 {
		checks = append(checks,
			struct{ code, got, want string }{"evidence.diagnostic_schema_mismatch", manifest.DiagnosticSchema, expected.Manifest.DiagnosticSchema},
			struct{ code, got, want string }{"evidence.digest_claim_mismatch", manifest.DigestClaim, expected.Manifest.DigestClaim},
			struct{ code, got, want string }{"evidence.escape_mismatch", manifest.KnownEscape, expected.Manifest.KnownEscape},
		)
	}
	for _, check := range checks {
		if check.got != check.want {
			return &ValidationError{Code: check.code}
		}
	}
	if !equalStrings(manifest.Flags, expected.Manifest.Flags) {
		return &ValidationError{Code: "evidence.flags_mismatch"}
	}
	if expected.Manifest.Schema == Schema1 && !equalStrings(manifest.ExecutionDigests, expected.Manifest.ExecutionDigests) {
		return &ValidationError{Code: "evidence.execution_mismatch"}
	}
	for _, check := range []struct {
		code string
		got  string
		want string
	}{
		{"evidence.source_mismatch", manifest.SourceDigest, expected.Manifest.SourceDigest},
		{"evidence.core_mismatch", manifest.CoreDigest, expected.Manifest.CoreDigest},
		{"evidence.c_mismatch", manifest.CDigest, expected.Manifest.CDigest},
	} {
		if subtle.ConstantTimeCompare([]byte(check.got), []byte(check.want)) != 1 {
			return &ValidationError{Code: check.code}
		}
	}
	if subtle.ConstantTimeCompare([]byte(manifest.ID), []byte(expected.Manifest.ID)) != 1 {
		return &ValidationError{Code: "evidence.id_mismatch"}
	}
	return nil
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func manifestID(manifest Manifest) string {
	if manifest.Schema == Schema1 {
		identity := struct {
			Schema, IDAlgorithm, SourceSchema, CoreSchema, ExecutionSchema, DiagnosticSchema string
			CompilerIdentity, ClangIdentity, Target                                          string
			Flags                                                                            []string
			Policy, SourceDigest, CoreDigest, CDigest                                        string
			ExecutionDigests                                                                 []string
			DigestClaim, KnownEscape                                                         string
		}{manifest.Schema, manifest.IDAlgorithm, manifest.SourceSchema, manifest.CoreSchema, manifest.ExecutionSchema, manifest.DiagnosticSchema,
			manifest.CompilerIdentity, manifest.ClangIdentity, manifest.Target, manifest.Flags, manifest.Policy, manifest.SourceDigest, manifest.CoreDigest, manifest.CDigest,
			manifest.ExecutionDigests, manifest.DigestClaim, manifest.KnownEscape}
		encoded, _ := json.Marshal(identity)
		sum := sha256.Sum256(encoded)
		return "evidence:" + hex.EncodeToString(sum[:12])
	}
	identity := struct {
		Schema           string
		IDAlgorithm      string
		SourceSchema     string
		CoreSchema       string
		ExecutionSchema  string
		CompilerIdentity string
		ClangIdentity    string
		Target           string
		Flags            []string
		Policy           string
		SourceDigest     string
		CoreDigest       string
		CDigest          string
	}{
		manifest.Schema, manifest.IDAlgorithm, manifest.SourceSchema, manifest.CoreSchema,
		manifest.ExecutionSchema, manifest.CompilerIdentity, manifest.ClangIdentity,
		manifest.Target, manifest.Flags, manifest.Policy, manifest.SourceDigest,
		manifest.CoreDigest, manifest.CDigest,
	}
	encoded, _ := json.Marshal(identity)
	sum := sha256.Sum256(encoded)
	return "evidence:" + hex.EncodeToString(sum[:12])
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func ErrorCode(err error) string {
	var validation *ValidationError
	if errors.As(err, &validation) {
		return validation.Code
	}
	return "evidence.operation_failed"
}

func Summary(product Product) string {
	return fmt.Sprintf("%s %s", product.Manifest.ID, digest(product.ManifestBytes))
}

func ContentDigest(data []byte) string { return digest(data) }
