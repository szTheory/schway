package native

// SetApplicationFlagsForTest swaps the application compile flags (D-21) and
// returns a restore function. The seam is test-only: production
// applicationFlags and recorded receipt flags stay -O0. Callers must not run in
// parallel because this swaps package-level state.
func SetApplicationFlagsForTest(flags []string) (restore func()) {
	previous := applicationFlags
	applicationFlags = append([]string(nil), flags...)
	return func() { applicationFlags = previous }
}

// SetApplicationCSourceMutationForTest installs a rewrite of the emitted C text
// (D-07 signed-char mutant) and returns a restore function.
func SetApplicationCSourceMutationForTest(mutate func(string) string) (restore func()) {
	previous := applicationCSourceMutationForTest
	applicationCSourceMutationForTest = mutate
	return func() { applicationCSourceMutationForTest = previous }
}
