package session_test

import (
	"context"
	"errors"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// phase273PreFixCleanCheckoutBlock is the verbatim Phase 27 guide block before
// D-27.3-15 routed its outputs under the ignored .schway/phase27/ directory. It
// is the negative control: the checker must keep reporting it.
const phase273PreFixCleanCheckoutBlock = `go build -o ./schway ./cmd/schway
./schway build examples/fizzbuzz.schway --output ./phase27-fizzbuzz
./schway build examples/sum_to_n.schway --output ./phase27-sum
./schway app run ./phase27-fizzbuzz -- 15 > ./fizzbuzz-15.actual
cmp ./fizzbuzz-15.actual examples/phase27/fizzbuzz-15.out
./schway app run ./phase27-fizzbuzz -- 1000 > ./fizzbuzz-1000.actual
cmp ./fizzbuzz-1000.actual examples/phase27/fizzbuzz-1000.out
./schway app run ./phase27-sum -- 10
`

const phase273MissingMkdir = "missing mkdir -p .schway/phase27 before first ./schway build"

// phase273CleanCheckoutBlock returns the body of the first sh fence after the
// "## Clean-checkout run" heading.
func phase273CleanCheckoutBlock(guide string) (string, error) {
	_, after, found := strings.Cut(guide, "\n## Clean-checkout run\n")
	if !found {
		return "", errors.New("guide has no '## Clean-checkout run' heading")
	}
	_, body, found := strings.Cut(after, "\n```sh\n")
	if !found {
		return "", errors.New("no ```sh fence after '## Clean-checkout run'")
	}
	block, _, found := strings.Cut(body, "```")
	if !found {
		return "", errors.New("unterminated sh fence after '## Clean-checkout run'")
	}
	return block, nil
}

// phase273GuideOutputTargets collects every output path the block writes:
// -o and --output targets (plus the .schway-build.json sidecar of --output),
// > redirect targets, and the artifact operand of "app run".
func phase273GuideOutputTargets(block string) []string {
	var targets []string
	for _, line := range strings.Split(block, "\n") {
		fields := strings.Fields(line)
		for i, field := range fields {
			next := ""
			if i+1 < len(fields) {
				next = fields[i+1]
			}
			switch {
			case field == "-o" && next != "":
				targets = append(targets, next)
			case field == "--output" && next != "":
				targets = append(targets, next, next+".schway-build.json")
			case field == ">" && next != "":
				targets = append(targets, next)
			case field == "run" && i > 0 && fields[i-1] == "app" && next != "":
				targets = append(targets, next)
			}
		}
	}
	return targets
}

func phase273TargetAllowed(target string) bool {
	return strings.HasPrefix(target, "./.schway/") || strings.HasPrefix(target, ".schway/") || target == "./schway"
}

// phase273GuideViolations lists every output target outside the ignored tree,
// plus the missing-mkdir message when the mkdir line is absent or follows the
// first build.
func phase273GuideViolations(block string) []string {
	var violations []string
	seen := map[string]bool{}
	for _, target := range phase273GuideOutputTargets(block) {
		if !phase273TargetAllowed(target) && !seen[target] {
			seen[target] = true
			violations = append(violations, target)
		}
	}
	mkdirLine, buildLine := -1, -1
	for i, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "mkdir -p .schway/phase27" && mkdirLine < 0 {
			mkdirLine = i
		}
		if strings.HasPrefix(trimmed, "./schway build ") && buildLine < 0 {
			buildLine = i
		}
	}
	if mkdirLine < 0 || (buildLine >= 0 && mkdirLine > buildLine) {
		violations = append(violations, phase273MissingMkdir)
	}
	return violations
}

// TestPhase273GuideCleanCheckoutOutputsIgnored is static, clang-free and always
// on. It never runs a guide command; git is only asked whether a path is
// ignored. The executed temp-repo variant of the guide run is declined
// (D-27.3-16 discretion) and stays deferred.
func TestPhase273GuideCleanCheckoutOutputsIgnored(t *testing.T) {
	t.Run("guide", func(t *testing.T) {
		raw, err := os.ReadFile(testsupport.ProjectPath("examples", "phase27", "README.md"))
		if err != nil {
			t.Fatal(err)
		}
		block, err := phase273CleanCheckoutBlock(string(raw))
		if err != nil {
			t.Fatal(err)
		}
		if violations := phase273GuideViolations(block); len(violations) != 0 {
			t.Fatalf("clean-checkout block writes outside the ignored .schway/ tree: %s", strings.Join(violations, ", "))
		}
		targets := phase273GuideOutputTargets(block)
		var builds, sidecars, redirects int
		for _, line := range strings.Split(block, "\n") {
			builds += strings.Count(line, "--output ")
			redirects += strings.Count(line, " > ")
		}
		for _, target := range targets {
			if strings.HasSuffix(target, ".schway-build.json") {
				sidecars++
			}
		}
		if builds < 2 || sidecars < 2 || redirects < 2 {
			t.Fatalf("expected at least 2 builds, 2 sidecars and 2 redirects, got %d/%d/%d", builds, sidecars, redirects)
		}
		distinct := map[string]bool{}
		for _, target := range targets {
			distinct[strings.TrimPrefix(target, "./")] = true
		}
		var names []string
		for name := range distinct {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			result, err := testsupport.RunCLIErr(context.Background(), "git", os.Environ(),
				"-C", testsupport.ProjectPath(), "check-ignore", "-q", "--", name)
			if err != nil {
				t.Fatalf("git check-ignore %s: %v", name, err)
			}
			switch result.Exit {
			case 0:
			case 1:
				t.Errorf("guide output %s is not ignored by git", name)
			default:
				t.Fatalf("git check-ignore %s exit %d: %s", name, result.Exit, result.Stderr)
			}
		}
	})
	t.Run("pre-fix-negative-control", func(t *testing.T) {
		violations := strings.Join(phase273GuideViolations(phase273PreFixCleanCheckoutBlock), "\n")
		for _, want := range []string{"./phase27-fizzbuzz", "./phase27-sum", "./fizzbuzz-15.actual", "./fizzbuzz-1000.actual", phase273MissingMkdir} {
			if !strings.Contains(violations, want) {
				t.Errorf("negative control did not report %q; violations:\n%s", want, violations)
			}
		}
	})
}
