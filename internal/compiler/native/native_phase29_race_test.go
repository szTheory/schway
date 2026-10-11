package native_test

import "runtime/debug"

// phase29RaceEnabled labels the interpreter lane of the spike 008 harness and
// disables its default-mode ceiling: the race detector slows the interpreter by
// a large, host-dependent factor, so timing is only logged under -race. It is
// read from the binary's build settings rather than a //go:build race pair so no
// build-constraint term outside the closed allowlist is needed.
var phase29RaceEnabled = func() bool {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return false
	}
	for _, setting := range info.Settings {
		if setting.Key == "-race" {
			return setting.Value == "true"
		}
	}
	return false
}()
