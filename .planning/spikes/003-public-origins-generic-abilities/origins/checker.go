package origins

import (
	"fmt"
	"strings"
)

func ConsumerCheck(public Interface, call CallCase) CheckResult {
	functions, err := FunctionMap(public.Functions)
	if err != nil {
		return invalid(call, "interface.invalid", err.Error())
	}
	function, ok := functions[call.Function]
	if !ok {
		return invalid(call, "interface.unknown_function", "function is absent from the public interface")
	}
	registry, err := NewRegistry(public.Types)
	if err != nil {
		return invalid(call, "interface.invalid", err.Error())
	}
	bindings := map[string]Binding{}
	alive := map[string]bool{}
	for _, binding := range call.Bindings {
		if _, exists := bindings[binding.ID]; exists {
			return invalid(call, "call.duplicate_binding", "binding identity is duplicated")
		}
		bindings[binding.ID] = binding
		alive[binding.ID] = true
	}
	for _, parameter := range function.Params {
		bindingID, exists := call.Arguments[parameter.Name]
		if !exists {
			return invalid(call, "call.missing_argument", "required named argument is missing")
		}
		binding, exists := bindings[bindingID]
		if !exists {
			return invalidBinding(call, "call.unknown_binding", bindingID, "argument names an unknown binding")
		}
		if !equalType(binding.Type, Substitute(parameter.Type, call.TypeArgs)) {
			return invalidBinding(call, "call.argument_type_mismatch", bindingID, "argument type does not match the public parameter type")
		}
	}
	if diagnostic := checkCallback(registry, function, call); diagnostic != nil {
		return CheckResult{Case: call.ID, Valid: false, Diagnostic: diagnostic}
	}
	returned, ok := returnForTag(function.Returns, call.ReturnTag)
	if !ok {
		return invalid(call, "call.unknown_return_tag", "selected return alternative is absent from the public interface")
	}
	resultType := Substitute(returned.Type, call.TypeArgs)
	originBindings, diagnostic := bindOrigins(function.ID, call, returned.Origins)
	if diagnostic != nil {
		return CheckResult{Case: call.ID, Valid: false, Diagnostic: diagnostic}
	}
	resultActive := true
	for index, step := range call.Steps {
		switch step.Kind {
		case "use_result":
			for _, bindingID := range originBindings {
				if !alive[bindingID] {
					return invalidOrigins(call, function.ID, "lifetime.origin_dead", []string{bindingID}, "borrowed result is used after an origin ended", "return_owned_or_extend_owner")
				}
			}
		case "drop_binding", "mutate_binding", "read_binding":
			if _, exists := bindings[step.Binding]; !exists {
				return invalidBinding(call, "call.unknown_binding", step.Binding, "step names an unknown binding")
			}
			conflicts := step.Kind != "read_binding" || returned.Access == "exclusive"
			if conflicts && resultActive && containsString(originBindings, step.Binding) && resultUsedAfter(call.Steps, index) {
				return invalidOrigins(call, function.ID, "ownership.borrow_conflict", []string{step.Binding}, "origin cannot end or mutate while its borrowed result remains live", "end_result_before_owner_access")
			}
			if step.Kind == "drop_binding" {
				alive[step.Binding] = false
			}
		case "copy_result", "drop_result", "send_result", "escape_result":
			ability := map[string]Ability{
				"copy_result": Copy, "drop_result": Drop, "send_result": Send, "escape_result": Escape,
			}[step.Kind]
			if step.Kind == "send_result" && returned.Kind == "borrowed" {
				hasEscape, abilityErr := registry.Has(resultType, Escape, call.TypeArgs)
				if abilityErr != nil {
					return invalid(call, "interface.invalid_type", abilityErr.Error())
				}
				if !hasEscape {
					return invalidAbility(call, function.ID, Escape, "sending to an unstructured destination would escape a borrowed result")
				}
			}
			has, abilityErr := registry.Has(resultType, ability, call.TypeArgs)
			if abilityErr != nil {
				return invalid(call, "interface.invalid_type", abilityErr.Error())
			}
			if !has {
				return invalidAbility(call, function.ID, ability, fmt.Sprintf("result type does not permit %s", ability))
			}
			if step.Kind == "drop_result" {
				resultActive = false
			}
		default:
			return invalid(call, "call.unknown_step", "call contains an unknown step")
		}
	}
	return CheckResult{Case: call.ID, Valid: true}
}

func checkCallback(registry Registry, function Function, call CallCase) *Diagnostic {
	if call.Callback == nil {
		if len(function.Callbacks) != 0 {
			return &Diagnostic{Code: "call.missing_callback", Function: function.ID, Case: call.ID, Message: "function requires a callback contract"}
		}
		return nil
	}
	var contract *CallbackContract
	for index := range function.Callbacks {
		if function.Callbacks[index].Parameter == call.Callback.Parameter {
			contract = &function.Callbacks[index]
			break
		}
	}
	if contract == nil {
		return &Diagnostic{Code: "call.unknown_callback", Function: function.ID, Case: call.ID, Message: "callback argument has no public contract"}
	}
	if contract.ResultType.Name != "" && !equalType(Substitute(contract.ResultType, call.TypeArgs), Substitute(call.Callback.ResultType, call.TypeArgs)) {
		return &Diagnostic{Code: "call.callback_result_type_mismatch", Function: function.ID, Case: call.ID, Message: "callback result type does not match the public callback contract"}
	}
	if containsString(call.Callback.ResultOrigins, contract.FreshOrigin) && containsAbility(contract.ResultRequires, Escape) {
		return &Diagnostic{Code: "lifetime.scoped_callback_escape", Function: function.ID, Case: call.ID, Origins: []string{contract.FreshOrigin}, Message: "callback result would retain a fresh invocation-scoped origin", RepairKind: "return_owned_callback_result"}
	}
	for _, ability := range contract.ResultRequires {
		has, err := registry.Has(Substitute(call.Callback.ResultType, call.TypeArgs), ability, call.TypeArgs)
		if err != nil {
			return &Diagnostic{Code: "interface.invalid_type", Function: function.ID, Case: call.ID, Message: err.Error()}
		}
		if !has {
			return &Diagnostic{Code: "ability." + string(ability) + "_required", Function: function.ID, Case: call.ID, Ability: ability, Message: "callback result lacks a required ability", RepairKind: "return_owned_callback_result"}
		}
	}
	return nil
}

func bindOrigins(function string, call CallCase, paths []string) ([]string, *Diagnostic) {
	var result []string
	for _, path := range paths {
		root := strings.SplitN(path, ".", 2)[0]
		binding, ok := call.Arguments[root]
		if !ok {
			return nil, &Diagnostic{Code: "interface.unknown_origin", Function: function, Case: call.ID, Origins: []string{path}, Message: "origin path does not begin at a named parameter"}
		}
		result = append(result, binding)
	}
	return canonicalStrings(result), nil
}

func returnForTag(returns []ReturnCase, tag string) (ReturnCase, bool) {
	if tag == "" && len(returns) == 1 {
		return returns[0], true
	}
	for _, returned := range returns {
		if returned.Tag == tag {
			return returned, true
		}
	}
	return ReturnCase{}, false
}

func resultUsedAfter(steps []Step, index int) bool {
	for _, step := range steps[index+1:] {
		switch step.Kind {
		case "drop_result":
			return false
		case "use_result", "copy_result", "send_result", "escape_result":
			return true
		}
	}
	return false
}

func originSubset(actual, declared []string) bool {
	for _, value := range actual {
		if !containsString(declared, value) {
			return false
		}
	}
	return true
}

func missingOrigins(actual, declared []string) []string {
	var result []string
	for _, value := range actual {
		if !containsString(declared, value) {
			result = append(result, value)
		}
	}
	return canonicalStrings(result)
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func equalType(left, right TypeExpr) bool {
	if left.Name != right.Name || len(left.Args) != len(right.Args) {
		return false
	}
	for index := range left.Args {
		if !equalType(left.Args[index], right.Args[index]) {
			return false
		}
	}
	return true
}

func invalid(call CallCase, code, message string) CheckResult {
	return CheckResult{Case: call.ID, Valid: false, Diagnostic: &Diagnostic{Code: code, Function: call.Function, Case: call.ID, Message: message}}
}

func invalidBinding(call CallCase, code, binding, message string) CheckResult {
	result := invalid(call, code, message)
	result.Diagnostic.Binding = binding
	return result
}

func invalidOrigins(call CallCase, function, code string, values []string, message, repair string) CheckResult {
	return CheckResult{Case: call.ID, Valid: false, Diagnostic: &Diagnostic{Code: code, Function: function, Case: call.ID, Origins: canonicalStrings(values), Message: message, RepairKind: repair}}
}

func invalidAbility(call CallCase, function string, ability Ability, message string) CheckResult {
	return CheckResult{Case: call.ID, Valid: false, Diagnostic: &Diagnostic{Code: "ability." + string(ability) + "_required", Function: function, Case: call.ID, Ability: ability, Message: message, RepairKind: "change_operation_or_add_ability"}}
}
