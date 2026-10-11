package session

import (
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/pathoracle"
)

// admitProgram is the single admission sequence for a checker-clean program
// (Phase 28, D-09): corevalidate.Validate, then
// originvalidate.ValidatePublished on the peer-normalized program, then
// pathoracle.ValidateLocalOwnerPaths -- the exact order CheckCommandFile has
// always used. It returns the peer-normalized program and nil diagnostics on
// admission, or exactly one diagnostic (the first problem of the refusing
// layer) on refusal; a refusal is never an error.
//
// The layers stay independent derivations (07-10 refusing-union rule): none
// reads another's verdict; only the diagnostic formatting is shared
// (peerRefusalDiagnostic falls back to core.peer_refusal_unnamed rather than
// indexing Problems[0] unguarded).
//
// Order (D-18): admission runs on the FULL checked program, before
// resolveEntry prunes, so dead code is validated before it is dropped.
//
// Lane rule (D-27.1-02, restated for D-19): called only from the command-file
// layer -- CheckCommandFile, BuildApplication, RunInterpreterApplication,
// RunInterpreterCommandFile and RunNativeCommandFile; never from
// RunInterpreter, RunNative or runNativeChecked, because differential and
// verify-corpus lanes deliberately run undeclared-borrow fixtures
// (TestExclusiveBorrowInterpreterNative).
func admitProgram(checked CheckResult) (core.Program, []diagnostic.Diagnostic) {
	program := checked.Program
	if !checkCommandPeerSeam {
		observeCheckCommandPeer("corevalidate")
		validated := corevalidate.Validate(program)
		if !validated.Valid {
			return program, []diagnostic.Diagnostic{peerRefusalDiagnostic(validated, program.ModuleID)}
		}
		program = validated.Program()
	}
	observeCheckCommandPeer("originvalidate")
	if problems := originvalidate.ValidatePublished(program); len(problems) > 0 {
		// D-04-27/WR-01: the foreign declaration surface is a second place an
		// alias fact can be silently absent (D-04-28), so every published
		// origin is recomputed from the typed core, not trusted from the
		// checker.
		return program, []diagnostic.Diagnostic{originPeerRefusalDiagnostic(problems[0], program, checked.Tree)}
	}
	observeCheckCommandPeer("pathoracle")
	if err := pathoracle.ValidateLocalOwnerPaths(program); err != nil {
		return program, []diagnostic.Diagnostic{diagnostic.Error("core.pathoracle_refused", diagnostic.Span{}, err.Error())}
	}
	return program, nil
}
