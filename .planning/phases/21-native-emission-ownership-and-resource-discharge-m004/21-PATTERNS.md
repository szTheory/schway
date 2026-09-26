# Phase 21: Native Emission Ownership and Resource Discharge — Pattern Map

**Mapped:** 2026-09-26  
**Files analyzed:** 4 existing test files identified by refreshed verification  
**Analogs found:** 4 / 4

The current task is gap closure after M003 archival. These are modifications to existing guards, not authorization to change compiler behavior, reopen emitter families, rerun UAT, or rewrite the original Phase 21 plans. M004 remains provisional and completed UAT remains preserved.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/compiler/cgen/cgen_program_test.go` | test | request-response / refusal assertion | `TestUnsupportedProgramShapePrecedesSchema2Preflight` in the same file | exact |
| `internal/compiler/session/session_test.go` | test | batch document lookup and contract validation | `phaseArtifactGlob` plus `TestPhase16EmitterCutsAreAmendedAndOwned` in the same file | exact |
| `internal/compiler/session/evidence_grade_test.go` | test | batch corpus transform and digest validation | `corpusRunRecord` / `requestedCorpusPairs` in the same file | exact |
| `internal/compiler/session/verification_groundedness_test.go` | test | batch document scan and reconciliation | `measuredViolations` / `reconciledFindings` in the same file | exact |

## Pattern Assignments

### `internal/compiler/cgen/cgen_program_test.go` (test, refusal assertion)

**Analog:** `TestUnsupportedProgramShapePrecedesSchema2Preflight`, same file, lines 625-644.

The failing test at lines 646-681 establishes a multi-function by-pointer refusal and currently hard-codes the pre-archive Phase 16 summary path. Keep its semantic assertion intact; resolve the human-reviewed decision from the archived M003 artifact while preserving the checks that refusal happens before serialization.

**Imports pattern** (lines 3-21):

```go
import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)
```

**Core refusal pattern** (`TestUnsupportedProgramShapePrecedesSchema2Preflight`, lines 625-644):

```go
_, err = cgen.EmitProgramForTest(checked.Program)
if err == nil {
	t.Fatal("expected by-pointer program shape to be refused")
}
if !strings.Contains(err.Error(), "by-pointer bodies are not supported by whole-program native emission this phase") {
	t.Fatalf("expected structural refusal before schema-2 preflight, got: %v", err)
}
if cgen.InvocationSerializationReachedForTest() {
	t.Fatal("unsupported program shape reached C serialization")
}
```

**Path pattern:** project fixtures and source artifacts are addressed through `testsupport.ProjectPath(...)`; source-backed decisions must use their current archived location under `.planning/milestones/M003-phases/16-branch-match-emitter-port/` (tracked summary: `16-05-SUMMARY.md`).

---

### `internal/compiler/session/session_test.go` (test, batch artifact lookup)

**Analog:** `phaseArtifactGlob` (lines 2552-2568) and `TestPhase16EmitterCutsAreAmendedAndOwned` (lines 4572-4610).

The failing test at lines 4576-4618 reads removed `.planning/REQUIREMENTS.md`. Current M003 requirements are tracked at `.planning/milestones/M003-REQUIREMENTS.md`. Preserve the single shared debt-register parser and existing mutation controls; update source resolution to account for the archive without creating another Markdown grammar. The tracked Phase 16 debt register is found in the M003 phase archive.

**Imports pattern** (lines 3-30):

```go
import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)
```

**Live/archive lookup pattern** (`phaseArtifactGlob`, lines 2556-2568):

```go
live, err := filepath.Glob(testsupport.ProjectPath(append([]string{".planning", "phases"}, parts...)...))
if err != nil {
	return nil, err
}
archived, err := filepath.Glob(testsupport.ProjectPath(append([]string{".planning", "milestones", "*-phases"}, parts...)...))
if err != nil {
	return nil, err
}
return append(live, archived...), nil
```

**Shared parser and negative-control pattern** (`TestPhase16EmitterCutsAreAmendedAndOwned`, lines 4582-4610):

```go
registers, err := phaseArtifactGlob("16-branch-match-emitter-port", "PHASE-16-DEBT.md")
if err != nil {
	t.Fatal(err)
}
// ...read the selected register...
if problems := phase16EmitterCutProblems(string(requirements), string(roadmap), string(debt)); len(problems) != 0 {
	t.Fatalf("Phase 16 NAT-09 amendment/debt mismatch:\n%s", strings.Join(problems, "\n"))
}
mutated := strings.Replace(string(debt), "emitLinearBorrowedByPointerPlain", "missingPointerFamily", 1)
if problems := phase16EmitterCutProblems(string(requirements), string(roadmap), mutated); len(problems) == 0 {
	t.Fatal("seeded missing pointer family passed the NAT-09 amendment/debt control")
}
```

---

### `internal/compiler/session/evidence_grade_test.go` (test, batch corpus/digest validation)

**Analog:** `corpusRunRecord` and `requestedCorpusPairs`, same file, lines 1519-1570; `loadCheckedInCorpusRunRecord` is at lines 373-385.

The failing test is at lines 1944 onward. Its requested pair set is derived from the live archived validation documents and indexed test names; `checkedInCorpusRunRecord` rejects stale pair digests, stale record digests, incomplete evidence, or excessive elapsed time. Refresh the checked-in record/manifest from the existing producer seam and keep the consumer as the sole derivation authority.

**Imports pattern** (lines 26-39):

```go
import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)
```

**Core scan and refusal pattern** (`TestValidationRowGradesAreEarnedOverArchivedCorpus`, lines 1944-1987):

```go
index := buildTestIndex(t)
byDoc := allPrimaryValidationRows(t)
if len(byDoc) == 0 {
	t.Fatal("no *-VALIDATION.md primary table discovered -- the glob or table detector has gone inert")
}
record := corpusRunRecord(t, index, byDoc)
// ...sort documents and validate each Grade, Non-inertness, and evidence row...
if problem := validationRowProblem(index, record, taskID, gradeCell, nonInertCell, evidence); problem != "" {
	t.Fatalf("%s: %s", name, problem)
}
```

**Deterministic requested-pair pattern** (`requestedCorpusPairs`, lines 1540-1570): sort document keys, resolve each row's evidence against the shared static test index, reject self-citation, deduplicate by package and pattern, and return the consolidated pairs. The record digest is checked by `checkedInCorpusRunRecord` (lines 324-371); preserve this one producer/consumer contract rather than hard-coding a new digest in the test.

---

### `internal/compiler/session/verification_groundedness_test.go` (test, batch scan/reconciliation)

**Analog:** `measuredViolations` and `reconciledFindings`, same file, lines 1309-1334 and 2145-2156.

`TestVerificationGroundednessThreeClassesAreEmpty` starts at line 2200. It compares fresh document-scan results with explicit reconciliation entries and a closed landing-phase map. The current archive move changes paths represented by the existing reconciliation/frontier data; update that data against actual archived findings and preserve set equality, class floors, and stale-entry checks.

**Imports pattern** (lines 21-38):

```go
import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)
```

**Measured finding pattern** (`measuredViolations`, lines 1309-1326):

```go
index := buildTestIndex(t)
docs := enforcedTierDocuments(t)
var measured []violationRecord
for _, doc := range docs {
	violations, err := classifyDocument(index, doc)
	if err != nil {
		t.Fatalf("classifyDocument(%s): %v", doc, err)
	}
	measured = append(measured, violations...)
}
sort.Slice(measured, func(i, j int) bool {
	if measured[i].File != measured[j].File {
		return measured[i].File < measured[j].File
	}
	return measured[i].Line < measured[j].Line
})
return measured
```

**Reconciliation pattern** (`reconciledFindings`, lines 2145-2156): obtain entries through `deriveEvidenceReconciliation`, then key the set by the full `(File, Line, Command, Classification)` record. The adjacent three-class test (lines 2200-2249) excludes only those full matches and fails on unreconciled R1/R2/R3 findings, missing R2b landing owners, stale landing entries, and corpus floors.

## Shared Patterns

### Project-relative file access

**Source:** `internal/compiler/testsupport/testsupport.go`, `ProjectPath`, lines 79-83.  
**Apply to:** all four test files.

```go
func ProjectPath(parts ...string) string {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
	return filepath.Join(append([]string{root}, parts...)...)
}
```

### Archive-aware phase lookup

**Source:** `internal/compiler/session/session_test.go`, `phaseArtifactGlob`, lines 2556-2568.  
**Apply to:** tests resolving phase artifacts moved under `.planning/milestones/<milestone>-phases/`; do not assume all artifacts remain in `.planning/phases/`.

### Evidence must fail closed

**Sources:** `checkedInCorpusRunRecord` in `evidence_grade_test.go`, lines 324-371; `TestVerificationGroundednessThreeClassesAreEmpty`, lines 2200-2249; cgen refusal checks, lines 625-681.  
**Apply to:** all changes. Preserve stale-input failures, complete-record checks, mutation controls, and the cgen serialization-reached assertion. Structural guard repair does not prove runtime cleanup and does not admit the cut emitters.

## No Analog Found

None. Each affected test has an in-file helper or adjacent refusal test matching its current role and flow.

## Metadata

**Analog search scope:** `internal/compiler/cgen/`, `internal/compiler/session/`, `internal/compiler/testsupport/`, archived M003 Phase 16 artifacts, validation corpus record paths, and Phase 21 verification findings.  
**Files scanned:** 4 target tests plus in-file helper regions and the project path helper.  
**Tracked-source gate:** all named code and archived artifact analogs were confirmed with `git ls-files`; no install/runtime mirrors are referenced.  
**Pattern extraction date:** 2026-09-26
