package origins

import "fmt"

func GenerateCases(functions []Function) []CallCase {
	var result []CallCase
	for _, function := range functions {
		if len(function.Callbacks) != 0 || function.Body == nil {
			continue
		}
		for _, returned := range function.Body.Returns {
			if returned.Kind != "borrowed" {
				continue
			}
			base := generatedBase(function, returned.Tag)
			result = append(result, withSteps(base, "use", []Step{{Kind: "use_result"}}))
			for _, origin := range returned.Origins {
				binding := base.Arguments[rootOf(origin)]
				result = append(result,
					withSteps(base, "mutate-"+binding, []Step{{Kind: "mutate_binding", Binding: binding}, {Kind: "use_result"}}),
					withSteps(base, "drop-"+binding, []Step{{Kind: "drop_binding", Binding: binding}, {Kind: "use_result"}}),
					withSteps(base, "end-then-mutate-"+binding, []Step{{Kind: "drop_result"}, {Kind: "mutate_binding", Binding: binding}}),
				)
			}
			result = append(result,
				withSteps(base, "copy", []Step{{Kind: "copy_result"}}),
				withSteps(base, "escape", []Step{{Kind: "escape_result"}}),
				withSteps(base, "send", []Step{{Kind: "send_result"}}),
			)
		}
	}
	return result
}

func FindInjectedMismatch(file FixtureFile) SearchResult {
	implementations, _ := FunctionMap(file.Functions)
	public, diagnostics := ExportInterface(file, ProducerOptions{})
	if len(diagnostics) != 0 {
		return SearchResult{}
	}
	for functionIndex := range public.Functions {
		changed := false
		for returnIndex := range public.Functions[functionIndex].Returns {
			origins := public.Functions[functionIndex].Returns[returnIndex].Origins
			if len(origins) > 1 {
				public.Functions[functionIndex].Returns[returnIndex].Origins = origins[:len(origins)-1]
				changed = true
				break
			}
		}
		if changed {
			break
		}
	}
	cases := GenerateCases(file.Functions)
	for index, call := range cases {
		comparison := Compare(public, implementations, call)
		if !comparison.Agreement {
			copy := call
			return SearchResult{Programs: index + 1, Mismatch: &comparison, Input: &copy}
		}
	}
	return SearchResult{Programs: len(cases)}
}

func LargeUnionFunction(origins int) Function {
	function := Function{ID: fmt.Sprintf("large_choose_%d", origins)}
	returned := ReturnCase{Kind: "borrowed", Type: TypeExpr{Name: "View", Args: []TypeExpr{{Name: "Byte"}}}}
	for index := 0; index < origins; index++ {
		name := fmt.Sprintf("source_%04d", index)
		function.Params = append(function.Params, Parameter{Name: name, Mode: "borrow", Type: TypeExpr{Name: "Buffer"}})
		returned.Origins = append(returned.Origins, name)
	}
	function.Returns = []ReturnCase{returned}
	function.Body = &BodyFacts{Returns: []ReturnCase{returned}}
	return function
}

func generatedBase(function Function, tag string) CallCase {
	call := CallCase{
		ID:        function.ID + ":" + tag,
		Function:  function.ID,
		Arguments: map[string]string{},
		TypeArgs:  map[string]TypeExpr{},
		ReturnTag: tag,
	}
	for _, parameter := range function.TypeParams {
		call.TypeArgs[parameter] = TypeExpr{Name: "Byte"}
	}
	for _, parameter := range function.Params {
		binding := "arg_" + parameter.Name
		call.Arguments[parameter.Name] = binding
		call.Bindings = append(call.Bindings, Binding{ID: binding, Type: Substitute(parameter.Type, call.TypeArgs)})
	}
	return call
}

func withSteps(base CallCase, suffix string, steps []Step) CallCase {
	base.ID += ":" + suffix
	base.Steps = append([]Step(nil), steps...)
	return base
}

func rootOf(origin string) string {
	for index, character := range origin {
		if character == '.' {
			return origin[:index]
		}
	}
	return origin
}
