package origins

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

const (
	ValueOrigins    = "value_origins"
	ExplicitRegions = "explicit_regions"
	CallbackOnly    = "callback_only"
)

var lexicalToken = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*|'[A-Za-z_][A-Za-z0-9_]*|->|[()\[\]{},:|&<>]`)

func RenderAll(functions []Function, approach string) ([]string, EncodingMetric) {
	metric := EncodingMetric{Approach: approach}
	var rendered []string
	for _, function := range functions {
		line, adapted, binders, origins := RenderFunction(function, approach)
		rendered = append(rendered, line)
		metric.Characters += len(line)
		metric.LexicalTokens += len(lexicalToken.FindAllString(line, -1))
		metric.ExplicitBinders += binders
		metric.OriginReferences += origins
		if adapted {
			metric.AdaptedFunctions++
		} else {
			metric.DirectFunctions++
		}
	}
	return rendered, metric
}

func RenderFunction(function Function, approach string) (text string, adapted bool, binders int, originReferences int) {
	switch approach {
	case ExplicitRegions:
		return renderRegions(function)
	case CallbackOnly:
		return renderCallbacks(function)
	default:
		return renderValueOrigins(function)
	}
}

func renderValueOrigins(function Function) (string, bool, int, int) {
	var parameters []string
	for _, parameter := range function.Params {
		parameters = append(parameters, renderParameter(parameter))
	}
	binders := 0
	origins := 0
	for _, callback := range function.Callbacks {
		binders++
		origins += len(callback.BorrowedFrom)
		parameters = append(parameters, fmt.Sprintf("%s: for %s fn(view: borrow(%s) View[Byte]) -> R where R: %s", callback.Parameter, callback.FreshOrigin, callback.FreshOrigin, joinAbilities(callback.ResultRequires)))
	}
	result := renderReturns(function.Returns, func(returned ReturnCase) string {
		origins += len(returned.Origins)
		if returned.Kind == "borrowed" {
			mode := "borrow"
			if returned.Access == "exclusive" {
				mode = "borrow mut"
			}
			return fmt.Sprintf("%s(%s) %s", mode, strings.Join(returned.Origins, " | "), renderType(returned.Type))
		}
		return renderType(returned.Type)
	})
	return fmt.Sprintf("fn %s%s(%s) -> %s", function.ID, renderTypeParams(function.TypeParams), strings.Join(parameters, ", "), result), false, binders, origins
}

func renderRegions(function Function) (string, bool, int, int) {
	regions := map[string]string{}
	for _, parameter := range function.Params {
		if parameter.Mode == "borrow" || parameter.Mode == "borrow_mut" {
			regions[parameter.Name] = "'" + parameter.Name
		}
	}
	var regionNames []string
	for _, region := range regions {
		regionNames = append(regionNames, region)
	}
	sort.Strings(regionNames)
	var parameters []string
	for _, parameter := range function.Params {
		prefix := parameter.Mode + " "
		if parameter.Mode == "borrow" || parameter.Mode == "borrow_mut" {
			if parameter.Mode == "borrow_mut" {
				prefix = "&'" + parameter.Name + " mut "
			} else {
				prefix = "&'" + parameter.Name + " "
			}
		} else if parameter.Mode == "" || parameter.Mode == "owned" {
			prefix = ""
		}
		parameters = append(parameters, fmt.Sprintf("%s: %s%s", parameter.Name, prefix, renderType(parameter.Type)))
	}
	binders := len(regionNames)
	originRefs := len(regionNames)
	for _, callback := range function.Callbacks {
		binders++
		originRefs += len(callback.BorrowedFrom)
		parameters = append(parameters, fmt.Sprintf("%s: for['%s] fn(&'%s View[Byte]) -> R where R: 'static", callback.Parameter, callback.FreshOrigin, callback.FreshOrigin))
	}
	needsResultRegion := false
	result := renderReturns(function.Returns, func(returned ReturnCase) string {
		if returned.Kind != "borrowed" {
			return renderType(returned.Type)
		}
		if len(returned.Origins) == 1 {
			return renderRegionType(returned.Type, "'"+rootOf(returned.Origins[0]))
		}
		needsResultRegion = true
		return renderRegionType(returned.Type, "'result")
	})
	var constraints []string
	for _, returned := range function.Returns {
		if len(returned.Origins) <= 1 {
			continue
		}
		for _, origin := range returned.Origins {
			constraints = append(constraints, fmt.Sprintf("'%s: 'result", rootOf(origin)))
		}
	}
	constraints = canonicalStrings(constraints)
	allBinders := append([]string{}, regionNames...)
	if needsResultRegion {
		allBinders = append(allBinders, "'result")
		binders++
	}
	generic := renderTypeParams(function.TypeParams)
	if len(allBinders) > 0 {
		generic = "[" + strings.Join(append(allBinders, function.TypeParams...), ", ") + "]"
	}
	line := fmt.Sprintf("fn %s%s(%s) -> %s", function.ID, generic, strings.Join(parameters, ", "), result)
	if len(constraints) > 0 {
		line += " where " + strings.Join(constraints, ", ")
	}
	return line, false, binders, originRefs + len(constraints)
}

func renderRegionType(expr TypeExpr, region string) string {
	switch expr.Name {
	case "View", "BorrowIter":
		var args []string
		for _, arg := range expr.Args {
			args = append(args, renderType(arg))
		}
		return expr.Name + "[" + strings.Join(append([]string{region}, args...), ", ") + "]"
	case "AnyView":
		return "AnyView[" + region + "]"
	case "Option":
		if len(expr.Args) == 1 {
			return "Option[" + renderRegionType(expr.Args[0], region) + "]"
		}
	}
	return "Borrowed[" + region + ", " + renderType(expr) + "]"
}

func renderCallbacks(function Function) (string, bool, int, int) {
	hasBorrowed := false
	for _, returned := range function.Returns {
		hasBorrowed = hasBorrowed || returned.Kind == "borrowed"
	}
	if !hasBorrowed {
		text, _, binders, origins := renderValueOrigins(function)
		return text, false, binders, origins
	}
	var parameters []string
	for _, parameter := range function.Params {
		parameters = append(parameters, renderParameter(parameter))
	}
	origins := 0
	for _, returned := range function.Returns {
		origins += len(returned.Origins)
	}
	parameters = append(parameters, "use: for access fn(view: borrow(access) Result) -> R where R: escape")
	return fmt.Sprintf("fn with_%s%s(%s) -> R", function.ID, renderTypeParams(append(function.TypeParams, "R")), strings.Join(parameters, ", ")), true, 1, origins
}

func renderParameter(parameter Parameter) string {
	prefix := ""
	if parameter.Mode != "" && parameter.Mode != "owned" {
		prefix = strings.ReplaceAll(parameter.Mode, "_", " ") + " "
	}
	return fmt.Sprintf("%s: %s%s", parameter.Name, prefix, renderType(parameter.Type))
}

func renderType(expr TypeExpr) string {
	if len(expr.Args) == 0 {
		return expr.Name
	}
	var args []string
	for _, arg := range expr.Args {
		args = append(args, renderType(arg))
	}
	return expr.Name + "[" + strings.Join(args, ", ") + "]"
}

func renderTypeParams(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return "[" + strings.Join(values, ", ") + "]"
}

func renderReturns(returns []ReturnCase, render func(ReturnCase) string) string {
	if len(returns) == 1 && returns[0].Tag == "" {
		return render(returns[0])
	}
	var result []string
	for _, returned := range returns {
		result = append(result, returned.Tag+"("+render(returned)+")")
	}
	return strings.Join(result, " | ")
}

func joinAbilities(abilities []Ability) string {
	var values []string
	for _, ability := range canonicalAbilities(abilities) {
		values = append(values, string(ability))
	}
	return strings.Join(values, " + ")
}
