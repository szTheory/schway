package origins

import "strings"

// OracleCheck intentionally evaluates implementation facts using a separately
// written transition loop. It does not consume the public function summary.
func OracleCheck(types []TypeSpec, implementation Function, call CallCase) CheckResult {
	if implementation.Body == nil {
		return invalid(call, "oracle.missing_body", "oracle requires implementation facts")
	}
	if implementation.Body.RetainedFreshOrigin != "" && call.Callback != nil && containsString(call.Callback.ResultOrigins, implementation.Body.RetainedFreshOrigin) {
		return invalidOrigins(call, implementation.ID, "lifetime.scoped_callback_escape", []string{implementation.Body.RetainedFreshOrigin}, "fresh callback origin escaped its invocation", "return_owned_callback_result")
	}
	actualReturn, ok := returnForTag(implementation.Body.Returns, call.ReturnTag)
	if !ok {
		return invalid(call, "call.unknown_return_tag", "implementation has no selected return alternative")
	}
	lookup := map[string]TypeSpec{}
	for _, spec := range types {
		lookup[spec.Name] = spec
	}
	alive := map[string]bool{}
	for _, binding := range call.Bindings {
		alive[binding.ID] = true
	}
	var owners []string
	for _, path := range actualReturn.Origins {
		root := strings.SplitN(path, ".", 2)[0]
		if binding, exists := call.Arguments[root]; exists {
			owners = append(owners, binding)
		} else {
			return invalidOrigins(call, implementation.ID, "interface.unknown_origin", []string{path}, "actual origin is not rooted in a parameter", "fix_body_or_summary")
		}
	}
	owners = canonicalStrings(owners)
	active := true
	for position := 0; position < len(call.Steps); position++ {
		step := call.Steps[position]
		if step.Kind == "drop_result" {
			if !oracleHas(lookup, Substitute(actualReturn.Type, call.TypeArgs), Drop, call.TypeArgs) {
				return invalidAbility(call, implementation.ID, Drop, "oracle result cannot be dropped")
			}
			active = false
			continue
		}
		if step.Kind == "drop_binding" || step.Kind == "mutate_binding" || step.Kind == "read_binding" {
			conflicts := step.Kind != "read_binding" || actualReturn.Access == "exclusive"
			if conflicts && active && containsString(owners, step.Binding) && oracleFutureUse(call.Steps[position+1:]) {
				return invalidOrigins(call, implementation.ID, "ownership.borrow_conflict", []string{step.Binding}, "implementation origin is still borrowed", "end_result_before_owner_access")
			}
			if step.Kind == "drop_binding" {
				alive[step.Binding] = false
			}
			continue
		}
		if step.Kind == "use_result" {
			for _, owner := range owners {
				if !alive[owner] {
					return invalidOrigins(call, implementation.ID, "lifetime.origin_dead", []string{owner}, "actual borrowed origin has ended", "return_owned_or_extend_owner")
				}
			}
			continue
		}
		ability, relevant := map[string]Ability{"copy_result": Copy, "send_result": Send, "escape_result": Escape}[step.Kind]
		if relevant {
			if step.Kind == "send_result" && actualReturn.Kind == "borrowed" && !oracleHas(lookup, Substitute(actualReturn.Type, call.TypeArgs), Escape, call.TypeArgs) {
				return invalidAbility(call, implementation.ID, Escape, "borrowed oracle result cannot escape through send")
			}
			if !oracleHas(lookup, Substitute(actualReturn.Type, call.TypeArgs), ability, call.TypeArgs) {
				return invalidAbility(call, implementation.ID, ability, "oracle result lacks required ability")
			}
			continue
		}
		return invalid(call, "call.unknown_step", "oracle encountered an unknown step")
	}
	if call.Callback != nil {
		for _, callback := range implementation.Callbacks {
			if callback.Parameter != call.Callback.Parameter {
				continue
			}
			if containsString(call.Callback.ResultOrigins, callback.FreshOrigin) && containsAbility(callback.ResultRequires, Escape) {
				return invalidOrigins(call, implementation.ID, "lifetime.scoped_callback_escape", []string{callback.FreshOrigin}, "callback result retains a fresh invocation origin", "return_owned_callback_result")
			}
			for _, ability := range callback.ResultRequires {
				if !oracleHas(lookup, Substitute(call.Callback.ResultType, call.TypeArgs), ability, call.TypeArgs) {
					return invalidAbility(call, implementation.ID, ability, "callback result lacks required ability")
				}
			}
		}
	}
	return CheckResult{Case: call.ID, Valid: true}
}

func oracleHas(specs map[string]TypeSpec, expr TypeExpr, target Ability, substitutions map[string]TypeExpr) bool {
	if replacement, exists := substitutions[expr.Name]; exists && len(expr.Args) == 0 {
		return oracleHas(specs, replacement, target, substitutions)
	}
	spec, exists := specs[expr.Name]
	if !exists || len(expr.Args) != spec.Parameters {
		return false
	}
	for _, base := range spec.Base {
		if base == target {
			return true
		}
	}
	indexes, exists := spec.Conditional[target]
	if !exists {
		return false
	}
	for _, index := range indexes {
		if index < 0 || index >= len(expr.Args) || !oracleHas(specs, expr.Args[index], target, substitutions) {
			return false
		}
	}
	return true
}

func oracleFutureUse(steps []Step) bool {
	for _, step := range steps {
		if step.Kind == "drop_result" {
			return false
		}
		if step.Kind == "use_result" || step.Kind == "copy_result" || step.Kind == "send_result" || step.Kind == "escape_result" {
			return true
		}
	}
	return false
}
