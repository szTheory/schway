package session

// This file declares Phase 6's residual escape register (D-06-13, D-06-29):
// residuals named and made visible in the gate's own output rather than
// hidden, following Phase5ExpectedEscapes' precedent verbatim. Declaring an
// escape here is not closing it -- an escape is, by construction, something
// no control in this repository detects. scripts/verify-phase6.sh greps for
// every one of these strings under "expected escapes must appear ... and
// must never be claimed as a solved, detected control" (the same discipline
// verify-phase5.sh's own comment states), and
// TestPhase6EscapesAreNeverPresentedAsControls asserts the intersection of
// this list with Phase6RequiredControls() is empty.

// EscapeCacheUndeclaredEnvironment names D-06-13 hole (1): undeclared
// environment -- locale, ulimit, filesystem case-sensitivity -- that the
// artifact cache's declared-input list (cache.ArtifactSpec) does not, and
// by construction cannot exhaustively, capture.
const EscapeCacheUndeclaredEnvironment = "escape:cache-undeclared-environment"

// EscapeCacheClangVersionStringStable names D-06-13 hole (2): a Clang
// change that does not alter its reported version string -- the ccache
// __TIME__-class footgun -- which the cache key's probed Clang identity
// cannot distinguish from an unchanged toolchain.
const EscapeCacheClangVersionStringStable = "escape:cache-clang-version-string-stable"

// EscapeCacheNondeterministicCodegen names D-06-13 hole (3): any future
// nondeterministic codegen silently breaking the cache's own "same
// declared inputs implies same artifact" premise, which is an argument
// about declared-input identity, never a proof of codegen determinism
// itself.
const EscapeCacheNondeterministicCodegen = "escape:cache-nondeterministic-codegen"

// EscapeCacheDirectoryHandEdited names D-06-13 hole (4): a hand-edited or
// partially deleted cache directory is indistinguishable from a genuinely
// cold one -- there is no integrity check beyond content-hash lookup,
// consistent with the standing project stance that SHA-256 is content
// identity, never proof.
const EscapeCacheDirectoryHandEdited = "escape:cache-directory-hand-edited"

// A fifth D-06-13 cache hole existed and is now CLOSED, not declared here:
// internal/compiler/cgen/*.go -- Phase 11's own code generator's source --
// was not a declared cache input, so an edited cgen could serve a stale
// binary for an unchanged .schway fixture (confirmed reproducing end-to-end
// by Phase 11 plan 11-01's Q-02 spike, BRANCH A). Phase 11 plan 11-07
// closed it by appending cgen_source as cache.DeclaredInputNames()'s eighth
// entry (see internal/compiler/cache/probe.go's own hole-comment item 5).
// It is deliberately absent from Phase6ExpectedEscapes below: an escape
// constant here would mis-declare a closed hole as still-open residual
// risk, which TestPhase6EscapesAreNeverPresentedAsControls's own
// intersection check exists to prevent in the other direction (closed
// controls never named as escapes; a fixed escape must not remain named
// either).

// EscapeRepairHeldoutCorpusResidualOverfitting names D-06-29's residual:
// the held-out defect-corpus discipline (testdata/phase6's heldout_*/
// derivation_* split) blunts, but cannot eliminate, the automated-program-
// repair overfitting concern given how small this language is.
const EscapeRepairHeldoutCorpusResidualOverfitting = "escape:repair-heldout-corpus-residual-overfitting"

// Phase6ExpectedEscapes is Phase 6's own declared, gate-visible residual
// set (D-06-13's four cache holes plus D-06-29's residual overfitting
// debt) -- none claimed solved, none ever permitted to appear as a
// detected lane control.
func Phase6ExpectedEscapes() []string {
	return []string{
		EscapeCacheUndeclaredEnvironment,
		EscapeCacheClangVersionStringStable,
		EscapeCacheNondeterministicCodegen,
		EscapeCacheDirectoryHandEdited,
		EscapeRepairHeldoutCorpusResidualOverfitting,
	}
}
