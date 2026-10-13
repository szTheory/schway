package check

import "strings"

// ModuleAwareRoutes names, once, the commands that load imports (D-12). The
// module.imports_not_loaded refusal text, the CLI usage text and the Phase 32
// route audit all derive from it. It returns a fresh slice so callers cannot
// mutate the set. The session package's phase30ModuleRouteFunctions and
// phase30BoundedWant name the same routes by function; they are cross-referenced
// by comment only because check must not import session.
func ModuleAwareRoutes() []string {
	return []string{"check", "run", "build", "app verify", "debug-map"}
}

// importsNotLoadedMessage builds the module.imports_not_loaded refusal text from
// ModuleAwareRoutes.
func importsNotLoadedMessage() string {
	return "imports are resolved only by the module-aware commands (" +
		strings.Join(ModuleAwareRoutes(), ", ") +
		"); this command reads a single file by design"
}
