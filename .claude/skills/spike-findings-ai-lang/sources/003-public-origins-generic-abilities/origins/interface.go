package origins

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type ProducerOptions struct {
	OmitLastOrigin bool
}

func ExportInterface(file FixtureFile, options ProducerOptions) (Interface, []Diagnostic) {
	output := Interface{Schema: 1, Types: append([]TypeSpec(nil), file.Types...)}
	var diagnostics []Diagnostic
	for _, function := range file.Functions {
		candidate := publicFunction(function)
		if options.OmitLastOrigin {
			for index := range candidate.Returns {
				if len(candidate.Returns[index].Origins) > 1 {
					candidate.Returns[index].Origins = candidate.Returns[index].Origins[:len(candidate.Returns[index].Origins)-1]
					break
				}
			}
		}
		if diagnostic := VerifyProducer(candidate, function.Body); diagnostic != nil {
			diagnostics = append(diagnostics, *diagnostic)
			continue
		}
		output.Functions = append(output.Functions, candidate)
	}
	return output, diagnostics
}

func RoundTripInterface(input Interface) (Interface, error) {
	data, err := json.Marshal(input)
	if err != nil {
		return Interface{}, err
	}
	if string(data) == "" || containsJSONKey(data, "body") {
		return Interface{}, fmt.Errorf("public interface leaked implementation body")
	}
	var output Interface
	if err := json.Unmarshal(data, &output); err != nil {
		return Interface{}, err
	}
	return output, nil
}

func VerifyProducer(function Function, body *BodyFacts) *Diagnostic {
	if body == nil {
		return &Diagnostic{Code: "producer.missing_body", Function: function.ID, Message: "producer verification requires implementation facts"}
	}
	if body.RetainedFreshOrigin != "" {
		for _, callback := range function.Callbacks {
			if callback.FreshOrigin == body.RetainedFreshOrigin {
				return &Diagnostic{Code: "producer.scoped_callback_escape", Function: function.ID, Origins: []string{body.RetainedFreshOrigin}, Message: "implementation retains an origin promised to be fresh and callback-scoped", RepairKind: "return_owned_or_remove_retention"}
			}
		}
	}
	if len(function.Returns) != len(body.Returns) {
		return &Diagnostic{Code: "producer.return_shape_mismatch", Function: function.ID, Message: "public return alternatives do not match implementation alternatives"}
	}
	for _, actual := range body.Returns {
		declared, ok := returnForTag(function.Returns, actual.Tag)
		if !ok {
			return &Diagnostic{Code: "producer.return_shape_mismatch", Function: function.ID, Message: "implementation return tag is absent from public summary"}
		}
		if actual.Kind == "borrowed" && declared.Kind != "borrowed" {
			return &Diagnostic{Code: "producer.ownership_mismatch", Function: function.ID, Origins: canonicalStrings(actual.Origins), Message: "borrowed implementation result was declared owned", RepairKind: "declare_borrowed_origin"}
		}
		if actual.Kind == "borrowed" && actual.Access == "exclusive" && declared.Access != "exclusive" {
			return &Diagnostic{Code: "producer.access_mismatch", Function: function.ID, Origins: canonicalStrings(actual.Origins), Message: "exclusive borrowed result was declared with weaker shared access", RepairKind: "declare_exclusive_access"}
		}
		if !originSubset(actual.Origins, declared.Origins) {
			return &Diagnostic{Code: "producer.summary_omits_origin", Function: function.ID, Origins: missingOrigins(actual.Origins, declared.Origins), Message: "public summary omits a possible implementation origin", RepairKind: "add_origin_or_return_tagged_variant"}
		}
	}
	return nil
}

func publicFunction(function Function) Function {
	result := Function{ID: function.ID, TypeParams: append([]string(nil), function.TypeParams...)}
	for _, parameter := range function.Params {
		parameter.Type = cloneType(parameter.Type)
		result.Params = append(result.Params, parameter)
	}
	for _, returned := range function.Returns {
		returned.Type = cloneType(returned.Type)
		returned.Origins = canonicalStrings(returned.Origins)
		result.Returns = append(result.Returns, returned)
	}
	for _, callback := range function.Callbacks {
		callback.BorrowedFrom = append([]string(nil), callback.BorrowedFrom...)
		callback.ResultType = cloneType(callback.ResultType)
		callback.ResultRequires = append([]Ability(nil), callback.ResultRequires...)
		result.Callbacks = append(result.Callbacks, callback)
	}
	return result
}

func cloneType(input TypeExpr) TypeExpr {
	result := TypeExpr{Name: input.Name}
	for _, arg := range input.Args {
		result.Args = append(result.Args, cloneType(arg))
	}
	return result
}

func containsJSONKey(data []byte, key string) bool {
	var object map[string]any
	if json.Unmarshal(data, &object) != nil {
		return false
	}
	var visit func(any) bool
	visit = func(value any) bool {
		switch typed := value.(type) {
		case map[string]any:
			if _, ok := typed[key]; ok {
				return true
			}
			for _, child := range typed {
				if visit(child) {
					return true
				}
			}
		case []any:
			for _, child := range typed {
				if visit(child) {
					return true
				}
			}
		}
		return false
	}
	return visit(object)
}

func InterfacesEqual(left, right Interface) bool {
	return reflect.DeepEqual(left, right)
}
